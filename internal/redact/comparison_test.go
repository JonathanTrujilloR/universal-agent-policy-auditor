package redact

import (
	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/compare"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"strings"
	"testing"
)

func comparisonFixture() app.Result {
	return app.Result{Mode: app.ModeCompare, Category: app.CompleteNoFindings, Build: app.BuildInfo{Version: "dev"}, Target: app.TargetOpenCode, RequestedTargetVersion: "1.18.27", SupportStatus: app.SupportSupported, Completeness: model.CompletenessComplete, Limitations: []string{"modeled-static-configuration"}, Comparison: &compare.Result{Outcome: compare.Equivalent}, ComparisonSources: []app.ComparisonSource{{Side: compare.Reference, Digest: "sha256:" + strings.Repeat("a", 64)}, {Side: compare.Target, Digest: "sha256:" + strings.Repeat("b", 64)}}}
}
func TestComparisonReportStates(t *testing.T) {
	for _, status := range []compare.Outcome{compare.TargetOnly, compare.Ambiguous, compare.NotComparable} {
		v := comparisonFixture()
		v.Category = app.UnsupportedOrIncomplete
		v.Completeness = model.CompletenessIncomplete
		v.Comparison.Outcome = status
		switch status {
		case compare.TargetOnly:
			v.Comparison.Differences = []compare.Difference{{Key: compare.Key{Capability: "read", Scope: "opencode", Matcher: "*"}, OnlyIn: compare.Target, Outcome: status, Reason: compare.Reason{Side: compare.Target, Code: compare.PermissionOnly}}}
		case compare.Ambiguous:
			v.Comparison.Reasons = []compare.Reason{{Side: compare.Reference, Code: compare.ConflictingDuplicate}}
		case compare.NotComparable:
			v.Comparison.Differences = []compare.Difference{{Key: compare.Key{Capability: "bash", Scope: "opencode", Matcher: "*"}, Outcome: status, Reason: compare.Reason{Side: compare.Target, Code: compare.EffectDiffers}}}
		}
		r, err := NewComparisonReport(v)
		if err != nil || r.Status() != ComparisonStatus(status) {
			t.Fatalf("%s: %v", status, err)
		}
	}
	v := comparisonFixture()
	v.Category = app.OperationalFailure
	v.Completeness = model.CompletenessIncomplete
	v.ComparisonSources = nil
	v.Comparison = &compare.Result{Outcome: compare.NotComparable, Reasons: []compare.Reason{{Side: compare.Reference, Code: compare.OperandIncomplete}}}
	if _, err := NewComparisonReport(v); err != nil {
		t.Fatal(err)
	}
	for _, category := range []app.Category{app.UnsupportedOrIncomplete, app.OperationalFailure} {
		for _, support := range []app.SupportStatus{app.SupportUnresolved, app.SupportUnsupported, app.SupportSupported} {
			if category == app.UnsupportedOrIncomplete && support == app.SupportSupported || category == app.OperationalFailure && support == app.SupportUnsupported {
				continue
			}
			v.Category, v.SupportStatus, v.Comparison = category, support, nil
			for _, version := range []string{"", "1.18.28"} {
				v.RequestedTargetVersion = version
				r, err := NewComparisonReport(v)
				if err != nil || r.Status() != "unavailable" {
					t.Fatalf("pre-core: %v", err)
				}
				if version != "" && (r.RequestedVersion() != "" || len(r.Findings()) != 1 || r.Findings()[0].Code() != "redact-version-withheld") {
					t.Fatal("withholding")
				}
			}
		}
	}
}
func TestComparisonReportRejects(t *testing.T) {
	for name, mutate := range map[string]func(*app.Result){
		"audit":       func(v *app.Result) { v.Mode = app.ModeAudit },
		"version":     func(v *app.Result) { v.Mode = app.ModeVersion },
		"category1":   func(v *app.Result) { v.Category = app.CompleteWithFindings },
		"invalid":     func(v *app.Result) { v.Category = app.InvalidRequest },
		"nil-success": func(v *app.Result) { v.Comparison = nil },
		"pre-core-finding": func(v *app.Result) {
			v.Category = app.UnsupportedOrIncomplete
			v.SupportStatus = app.SupportUnresolved
			v.Completeness = model.CompletenessIncomplete
			v.Comparison = nil
			v.ComparisonSources = nil
			v.Findings = []model.Finding{model.NewFinding("source-unreadable", "secret")}
		},
		"side":          func(v *app.Result) { v.ComparisonSources[0].Side = "secret" },
		"digest":        func(v *app.Result) { v.ComparisonSources[0].Digest = "secret" },
		"conflict":      func(v *app.Result) { v.ComparisonSources[1].Side = compare.Reference },
		"cap":           func(v *app.Result) { v.ComparisonSources = append(v.ComparisonSources, v.ComparisonSources[0]) },
		"finding":       func(v *app.Result) { v.Findings = []model.Finding{model.NewFinding("secret", "secret")} },
		"limitation":    func(v *app.Result) { v.Limitations = []string{"secret"} },
		"version-state": func(v *app.Result) { v.Build.Version = "secret" },
		"state":         func(v *app.Result) { v.Completeness = model.CompletenessIncomplete },
		"outcome":       func(v *app.Result) { v.Comparison.Outcome = "secret" },
		"operational-semantic": func(v *app.Result) {
			v.Category = app.OperationalFailure
			v.Completeness = model.CompletenessIncomplete
			v.ComparisonSources = nil
			v.Comparison = &compare.Result{
				Outcome: compare.NotComparable,
				Differences: []compare.Difference{{
					Key:     compare.Key{Capability: "read", Scope: "opencode", Matcher: "*"},
					Outcome: compare.NotComparable,
					Reason:  compare.Reason{Side: compare.Target, Code: compare.EffectDiffers},
				}},
			}
		},
		"code": func(v *app.Result) {
			v.Category = app.UnsupportedOrIncomplete
			v.Completeness = model.CompletenessIncomplete
			v.Comparison = &compare.Result{Outcome: compare.NotComparable, Reasons: []compare.Reason{{Side: compare.Target, Code: "secret"}}}
		},
		"key": func(v *app.Result) {
			v.Category = app.UnsupportedOrIncomplete
			v.Completeness = model.CompletenessIncomplete
			v.Comparison = &compare.Result{Outcome: compare.TargetOnly, Differences: []compare.Difference{{Key: compare.Key{Capability: "secret", Scope: "opencode", Matcher: "*"}, OnlyIn: compare.Target, Outcome: compare.TargetOnly, Reason: compare.Reason{Side: compare.Target, Code: compare.PermissionOnly}}}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			v := comparisonFixture()
			mutate(&v)
			r, err := NewComparisonReport(v)
			if err != ErrUnsafeReport || r.Valid() || r.Status() != "" || r.Target() != "" {
				t.Fatal("unsafe result accepted or echoed")
			}
		})
	}
}

func TestComparisonReportEquivalent(t *testing.T) {
	r, err := NewComparisonReport(comparisonFixture())
	if err != nil || !r.Valid() || r.Status() != "equivalent" || r.ExitCode() != 0 {
		t.Fatalf("report: %v", err)
	}
	if r.ComparisonSources() == nil || r.Reasons() == nil || r.Differences() == nil || r.Findings() == nil || r.Limitations() == nil {
		t.Fatal("nil collection")
	}
	sources := r.ComparisonSources()
	sources[0] = ComparisonSource{}
	if r.ComparisonSources()[0].Digest() == "" {
		t.Fatal("alias")
	}
	if (ComparisonReport{}).Valid() {
		t.Fatal("zero valid")
	}
}
