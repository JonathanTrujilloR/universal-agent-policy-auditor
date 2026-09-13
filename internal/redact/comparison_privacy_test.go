package redact

import (
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/compare"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
)

var comparisonOperandCodes = []compare.ReasonCode{compare.InvalidOperandMetadata, compare.IncompatibleCanonicalization, compare.IncompatibleCoverage, compare.OperandIncomplete, compare.InvalidIdentity, compare.TraceIncomplete, compare.UnresolvedPermission, compare.UnsupportedEffect, compare.ConflictingDuplicate}
var comparisonSemanticCodes = []compare.ReasonCode{compare.EffectDiffers, compare.ConditionsDiffer, compare.ExtensionsDiffer}

func privacyComparison() app.Result {
	v := comparisonFixture()
	v.Category, v.Completeness = app.UnsupportedOrIncomplete, model.CompletenessIncomplete
	v.Comparison = &compare.Result{Outcome: compare.NotComparable, Reasons: []compare.Reason{{Side: compare.Reference, Code: compare.OperandIncomplete}}}
	return v
}

func comparisonCheck(t testing.TB, v app.Result, accept bool) ComparisonReport {
	t.Helper()
	r, err := NewComparisonReport(v)
	if accept {
		if err != nil || !r.Valid() {
			t.Fatalf("expected valid report: %v", err)
		}
	} else if err != ErrUnsafeReport || err.Error() != "redact: unsafe report" || !reflect.DeepEqual(r, ComparisonReport{}) {
		t.Fatal("expected fixed error and zero report")
	}
	return r
}

// Project each public accessor, never format opaque structs.
func comparisonProjection(r ComparisonReport) []string {
	p := []string{string(r.Category()), string(r.Target()), r.ToolVersion(), r.RequestedVersion(), string(r.SupportStatus()), string(r.Completeness()), string(r.Status())}
	for _, s := range r.ComparisonSources() {
		p = append(p, string(s.Side()), string(s.Digest()))
	}
	for _, reason := range r.Reasons() {
		p = append(p, string(reason.Side()), string(reason.Code()))
	}
	for _, d := range r.Differences() {
		p = append(p, string(d.Capability()), string(d.Scope()), string(d.Matcher()), string(d.OnlyIn()), string(d.Outcome()), string(d.Reason().Side()), string(d.Reason().Code()))
	}
	for _, f := range r.Findings() {
		p = append(p, string(f.Code()))
	}
	for _, l := range r.Limitations() {
		p = append(p, string(l))
	}
	return p
}

func TestComparisonVersionPrivacy(t *testing.T) {
	for _, category := range []app.Category{app.UnsupportedOrIncomplete, app.OperationalFailure, app.CompleteNoFindings} {
		for _, field := range []string{"tool", "requested", "both"} {
			v := privacyComparison()
			v.Category, v.ComparisonSources = category, nil
			if category == app.CompleteNoFindings {
				v = comparisonFixture()
			}
			if field != "requested" {
				v.Build.Version = "https://private.invalid/token-8472"
			}
			if field != "tool" {
				v.RequestedTargetVersion = "https://private.invalid/token-8472"
			}
			r := comparisonCheck(t, v, category != app.CompleteNoFindings)
			if r.Valid() && (len(r.Findings()) != 1 || r.Findings()[0].Code() != "redact-version-withheld" || field != "requested" && r.ToolVersion() != "" || field != "tool" && r.RequestedVersion() != "") {
				t.Fatal("version withholding contract")
			}
		}
	}
	v := privacyComparison()
	v.Comparison, v.ComparisonSources, v.SupportStatus, v.RequestedTargetVersion = nil, nil, app.SupportUnresolved, ""
	if len(comparisonCheck(t, v, true).Findings()) != 0 {
		t.Fatal("missing evidence is not withheld")
	}
}

func TestComparisonEnumsPrivacy(t *testing.T) {
	comparisonCheck(t, comparisonFixture(), true)
	for _, category := range []app.Category{app.UnsupportedOrIncomplete, app.OperationalFailure} {
		for _, support := range []app.SupportStatus{app.SupportUnresolved, app.SupportUnsupported, app.SupportSupported} {
			v := privacyComparison()
			v.Category, v.SupportStatus, v.Comparison, v.ComparisonSources = category, support, nil, nil
			comparisonCheck(t, v, category == app.UnsupportedOrIncomplete && support != app.SupportSupported || category == app.OperationalFailure && support != app.SupportUnsupported)
		}
		for _, side := range []compare.Side{compare.Reference, compare.Target} {
			for _, code := range comparisonOperandCodes {
				v := privacyComparison()
				v.Category, v.ComparisonSources = category, nil
				v.Comparison.Reasons = []compare.Reason{{Side: side, Code: code}}
				if code == compare.ConflictingDuplicate {
					v.Comparison.Outcome = compare.Ambiguous
				}
				comparisonCheck(t, v, category != app.OperationalFailure || code != compare.ConflictingDuplicate)
				for _, bad := range []compare.ReasonCode{code + "-", " " + code, compare.PermissionOnly, compare.EffectDiffers} {
					v.Comparison.Reasons[0].Code = bad
					comparisonCheck(t, v, false)
				}
			}
		}
	}
	for _, outcome := range []compare.Outcome{compare.TargetOnly, compare.NotComparable} {
		for _, side := range []compare.Side{compare.Reference, compare.Target} {
			for _, capability := range []string{"bash", "edit", "read"} {
				for _, code := range append([]compare.ReasonCode{compare.PermissionOnly}, comparisonSemanticCodes...) {
					v := privacyComparison()
					v.Comparison = &compare.Result{Outcome: outcome, Differences: []compare.Difference{{Key: compare.Key{Capability: capability, Scope: "opencode", Matcher: "*"}, Outcome: outcome, Reason: compare.Reason{Side: side, Code: code}}}}
					if outcome == compare.TargetOnly {
						v.Comparison.Differences[0].OnlyIn = side
					}
					ok := outcome == compare.TargetOnly && code == compare.PermissionOnly || outcome == compare.NotComparable && side == compare.Target && code != compare.PermissionOnly
					comparisonCheck(t, v, ok)
					v.Comparison.Differences[0].Reason.Code = code + "-"
					comparisonCheck(t, v, false)
				}
			}
		}
	}
}

func semanticComparison() app.Result {
	v := privacyComparison()
	v.Comparison.Reasons = nil
	for _, capability := range []string{"bash", "edit", "read"} {
		for _, code := range comparisonSemanticCodes {
			v.Comparison.Differences = append(v.Comparison.Differences, compare.Difference{Key: compare.Key{Capability: capability, Scope: "opencode", Matcher: "*"}, Outcome: compare.NotComparable, Reason: compare.Reason{Side: compare.Target, Code: code}})
		}
	}
	return v
}

func TestComparisonCardinality(t *testing.T) {
	v := privacyComparison()
	comparisonCheck(t, v, true)
	v.ComparisonSources = append(v.ComparisonSources, v.ComparisonSources[0])
	comparisonCheck(t, v, false)
	v.ComparisonSources = v.ComparisonSources[:2]
	v.ComparisonSources[1].Side = compare.Reference
	comparisonCheck(t, v, false)
	v.ComparisonSources[1] = v.ComparisonSources[0]
	comparisonCheck(t, v, true)
	for len(v.Comparison.Reasons) < 24 {
		v.Comparison.Reasons = append(v.Comparison.Reasons, v.Comparison.Reasons[0])
	}
	if len(comparisonCheck(t, v, true).Reasons()) != 1 {
		t.Fatal("reason deduplication")
	}
	v.Comparison.Reasons = append(v.Comparison.Reasons, v.Comparison.Reasons[0])
	comparisonCheck(t, v, false)
	v = semanticComparison()
	if len(comparisonCheck(t, v, true).Differences()) != 9 {
		t.Fatal("semantic cardinality")
	}
	v.Comparison.Differences = append(v.Comparison.Differences, v.Comparison.Differences[0])
	comparisonCheck(t, v, false)
	v = privacyComparison()
	v.Build.Version = "withhold-private"
	for code := range allowedFindings {
		v.Findings = append(v.Findings, model.NewFinding(code, "ignored"))
	}
	if len(comparisonCheck(t, v, true).Findings()) != 15 {
		t.Fatal("generated marker cardinality")
	}
	v.Findings = append(v.Findings, v.Findings[0])
	comparisonCheck(t, v, true)
	v.Findings = append(v.Findings, v.Findings[0])
	comparisonCheck(t, v, false)
}

func TestComparisonOrderingCopies(t *testing.T) {
	rng := rand.New(rand.NewSource(45))
	for _, v := range []app.Result{privacyComparison(), semanticComparison()} {
		if len(v.Comparison.Reasons) > 0 {
			for _, side := range []compare.Side{compare.Reference, compare.Target} {
				for _, code := range comparisonOperandCodes[:8] {
					v.Comparison.Reasons = append(v.Comparison.Reasons, compare.Reason{Side: side, Code: code})
				}
			}
			for code := range allowedFindings {
				v.Findings = append(v.Findings, model.NewFinding(code, "ignored"))
			}
		}
		want := comparisonProjection(comparisonCheck(t, v, true))
		for run := 0; run < 100; run++ {
			rng.Shuffle(len(v.ComparisonSources), func(i, j int) {
				v.ComparisonSources[i], v.ComparisonSources[j] = v.ComparisonSources[j], v.ComparisonSources[i]
			})
			rng.Shuffle(len(v.Comparison.Reasons), func(i, j int) {
				v.Comparison.Reasons[i], v.Comparison.Reasons[j] = v.Comparison.Reasons[j], v.Comparison.Reasons[i]
			})
			rng.Shuffle(len(v.Comparison.Differences), func(i, j int) {
				v.Comparison.Differences[i], v.Comparison.Differences[j] = v.Comparison.Differences[j], v.Comparison.Differences[i]
			})
			rng.Shuffle(len(v.Findings), func(i, j int) { v.Findings[i], v.Findings[j] = v.Findings[j], v.Findings[i] })
			r := comparisonCheck(t, v, true)
			if !reflect.DeepEqual(want, comparisonProjection(r)) {
				t.Fatal("permutation changed projection")
			}
			sources, reasons, differences, findings, limitations := r.ComparisonSources(), r.Reasons(), r.Differences(), r.Findings(), r.Limitations()
			for _, values := range [][]string{reasonOrder(reasons), differenceOrder(differences), findingOrder(findings)} {
				if !slices.IsSorted(values) {
					t.Fatal("not lexical/side order")
				}
			}
			if sources[0].Side() != compare.Reference || sources[1].Side() != compare.Target {
				t.Fatal("source side order")
			}
			clear(sources)
			clear(reasons)
			clear(differences)
			clear(findings)
			clear(limitations)
			if !reflect.DeepEqual(want, comparisonProjection(r)) {
				t.Fatal("accessor alias, including nested difference reason")
			}
		}
	}
	v := privacyComparison()
	v.Comparison, v.ComparisonSources, v.SupportStatus = nil, nil, app.SupportUnresolved
	for _, v := range []app.Result{v, comparisonFixture()} {
		r := comparisonCheck(t, v, true)
		if r.ComparisonSources() == nil || r.Reasons() == nil || r.Differences() == nil || r.Findings() == nil || r.Limitations() == nil || (ComparisonReport{}).Valid() {
			t.Fatal("zero/empty contract")
		}
	}
}

func reasonOrder(values []ComparisonReason) []string {
	out := []string{}
	for _, v := range values {
		out = append(out, string(v.Side())+"/"+string(v.Code()))
	}
	return out
}
func differenceOrder(values []ComparisonDifference) []string {
	out := []string{}
	for _, v := range values {
		out = append(out, strings.Join([]string{string(v.Capability()), string(v.Scope()), string(v.Matcher()), string(v.OnlyIn()), string(v.Outcome()), string(v.Reason().Side()), string(v.Reason().Code())}, "/"))
	}
	return out
}
func findingOrder(values []Finding) []string {
	out := []string{}
	for _, v := range values {
		out = append(out, string(v.Code()))
	}
	return out
}

var comparisonCanaries = []string{"/home/private-8472/config", "https://private-8472.invalid/token", "Bearer private-8472-secret", "user:private-8472-password", "private-8472\x00\n\x1b", "private-8472\xff", "prіvate-8472", "{\"token\":\"private-8472\"}"}

func comparisonChannel(v *app.Result, channel int, value string) {
	switch channel {
	case 0:
		v.Message = value
	case 1:
		v.SourceDigests = []string{value}
	case 2:
		v.Permissions = []model.Permission{privatePermission(value)}
	case 3:
		v.Findings = []model.Finding{model.NewFinding("source-unreadable", value)}
	case 4:
		v.Build.Version = value
	case 5:
		v.RequestedTargetVersion = value
	case 6:
		v.ComparisonSources[0].Digest = value
	case 7:
		v.ComparisonSources[0].Side = compare.Side(value)
	case 8:
		v.Comparison.Outcome = compare.Outcome(value)
	case 9:
		v.Comparison.Reasons[0].Side = compare.Side(value)
	case 10:
		v.Comparison.Reasons[0].Code = compare.ReasonCode(value)
	case 11:
		v.Findings = []model.Finding{model.NewFinding(value, "ignored")}
	case 12:
		v.Limitations = []string{value}
	default:
		*v = semanticComparison()
		d := &v.Comparison.Differences[0]
		switch channel {
		case 13:
			d.Key.Capability = value
		case 14:
			d.Key.Scope = model.Scope(value)
		case 15:
			d.Key.Matcher = value
		case 16:
			d.OnlyIn = compare.Side(value)
		case 17:
			d.Outcome = compare.Outcome(value)
		case 18:
			d.Reason.Side = compare.Side(value)
		case 19:
			d.Reason.Code = compare.ReasonCode(value)
		}
	}
}

func comparisonPrivacyOracle(t testing.TB, value string, channel int) {
	t.Helper()
	v := privacyComparison()
	comparisonChannel(&v, channel, value)
	r, err := NewComparisonReport(v)
	if err != nil {
		comparisonCheck(t, v, false)
		return
	}
	closed := append(comparisonProjection(comparisonCheck(t, comparisonFixture(), true)), "", "v0.1.0-alpha.1", "unsupported_or_incomplete", "incomplete", "not-comparable", "ambiguous", "target-only", "reference", "target", "bash", "edit", "read", "*", "redact-version-withheld", string(compare.PermissionOnly))
	for _, code := range append(slices.Clone(comparisonOperandCodes), comparisonSemanticCodes...) {
		closed = append(closed, string(code))
	}
	for code := range allowedFindings {
		closed = append(closed, code)
	}
	got := comparisonProjection(r)
	if !r.Valid() || r.ExitCode() != app.ExitCode(v.Category) {
		t.Fatal("projection escaped closed state")
	}
	for _, part := range got {
		if !slices.Contains(closed, part) && !validDigest(part) {
			t.Fatal("non-closed accessor value")
		}
		if !utf8.ValidString(part) || strings.Contains(part, "private-8472") || strings.Contains(part, "prіvate-8472") {
			t.Fatal("canary or invalid UTF-8 escaped")
		}
	}
	if len(value) >= 8 && !slices.Contains(closed, value) && !validDigest(value) && strings.Contains(strings.Join(got, "|"), value) {
		t.Fatal("full canary escaped")
	}
}

func TestComparisonChannelsPrivacy(t *testing.T) {
	for channel, valid := range map[int]string{6: digest("a"), 7: "reference", 8: "not-comparable", 9: "reference", 10: string(compare.OperandIncomplete), 11: "source-unreadable", 12: "modeled-static-configuration", 13: "bash", 14: "opencode", 15: "*", 16: "", 17: "not-comparable", 18: "target", 19: string(compare.EffectDiffers)} {
		for _, nearby := range []string{" " + valid, valid + "-"} {
			v := privacyComparison()
			comparisonChannel(&v, channel, nearby)
			comparisonCheck(t, v, false)
		}
	}
	for _, canary := range comparisonCanaries {
		for _, value := range []string{canary, "prefix-" + canary, canary + "-suffix", "private-8472"} {
			for channel := 0; channel < 20; channel++ {
				v := privacyComparison()
				comparisonChannel(&v, channel, value)
				comparisonCheck(t, v, channel < 6)
				comparisonPrivacyOracle(t, value, channel)
			}
		}
	}
}

func FuzzComparisonReportPrivacy(f *testing.F) {
	for _, seed := range comparisonCanaries {
		for channel := uint8(0); channel < 20; channel++ {
			f.Add(seed, channel)
		}
	}
	f.Fuzz(func(t *testing.T, value string, channel uint8) {
		if len(value) > 512 {
			t.Skip()
		}
		comparisonPrivacyOracle(t, value, int(channel%20))
	})
}
