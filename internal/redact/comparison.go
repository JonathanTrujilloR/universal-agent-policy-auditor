package redact

import (
	"cmp"
	"slices"
	"sort"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/compare"
)

type ComparisonStatus string

const (
	ComparisonUnavailable   ComparisonStatus = "unavailable"
	ComparisonEquivalent    ComparisonStatus = "equivalent"
	ComparisonTargetOnly    ComparisonStatus = "target-only"
	ComparisonAmbiguous     ComparisonStatus = "ambiguous"
	ComparisonNotComparable ComparisonStatus = "not-comparable"
)

// ComparisonReport is a closed projection, independent of the audit report contract.
type ComparisonReport struct {
	data        Report
	status      ComparisonStatus
	sources     []ComparisonSource
	reasons     []ComparisonReason
	differences []ComparisonDifference
}
type ComparisonSource struct {
	side   compare.Side
	digest Digest
}
type ComparisonReason struct {
	side compare.Side
	code compare.ReasonCode
}
type ComparisonDifference struct {
	capability Capability
	scope      Scope
	matcher    Matcher
	onlyIn     compare.Side
	outcome    compare.Outcome
	reason     ComparisonReason
}

func (r ComparisonReport) Valid() bool                  { return r.data.valid }
func (r ComparisonReport) Category() Category           { return r.data.category }
func (r ComparisonReport) ExitCode() int                { return r.data.exitCode }
func (r ComparisonReport) ToolVersion() string          { return r.data.toolVersion }
func (r ComparisonReport) Target() Target               { return r.data.target }
func (r ComparisonReport) RequestedVersion() string     { return r.data.requestedVersion }
func (r ComparisonReport) SupportStatus() SupportStatus { return r.data.supportStatus }
func (r ComparisonReport) Completeness() Completeness   { return r.data.completeness }
func (r ComparisonReport) Status() ComparisonStatus     { return r.status }
func (r ComparisonReport) ComparisonSources() []ComparisonSource {
	return append([]ComparisonSource{}, r.sources...)
}
func (r ComparisonReport) Reasons() []ComparisonReason {
	return append([]ComparisonReason{}, r.reasons...)
}
func (r ComparisonReport) Differences() []ComparisonDifference {
	return append([]ComparisonDifference{}, r.differences...)
}
func (r ComparisonReport) Findings() []Finding          { return r.data.Findings() }
func (r ComparisonReport) Limitations() []Limitation    { return r.data.Limitations() }
func (s ComparisonSource) Side() compare.Side           { return s.side }
func (s ComparisonSource) Digest() Digest               { return s.digest }
func (r ComparisonReason) Side() compare.Side           { return r.side }
func (r ComparisonReason) Code() compare.ReasonCode     { return r.code }
func (d ComparisonDifference) Capability() Capability   { return d.capability }
func (d ComparisonDifference) Scope() Scope             { return d.scope }
func (d ComparisonDifference) Matcher() Matcher         { return d.matcher }
func (d ComparisonDifference) OnlyIn() compare.Side     { return d.onlyIn }
func (d ComparisonDifference) Outcome() compare.Outcome { return d.outcome }
func (d ComparisonDifference) Reason() ComparisonReason { return d.reason }

func NewComparisonReport(result app.Result) (ComparisonReport, error) {
	r, ok := comparisonReport(result)
	if !ok {
		return ComparisonReport{}, ErrUnsafeReport
	}
	r.data.valid = true
	return r, nil
}
func comparisonReport(v app.Result) (ComparisonReport, bool) {
	r := ComparisonReport{status: ComparisonUnavailable}
	d := &r.data
	if v.Mode != app.ModeCompare || v.Target != app.TargetOpenCode {
		return r, false
	}
	var ok bool
	if d.category, ok = categories[v.Category]; !ok {
		return r, false
	}
	if d.supportStatus, ok = supports[v.SupportStatus]; !ok {
		return r, false
	}
	if d.completeness, ok = completes[v.Completeness]; !ok {
		return r, false
	}
	d.target, d.exitCode = TargetOpenCode, app.ExitCode(v.Category)
	d.toolVersion, ok = safeToolVersion(v.Build.Version)
	withheld := !ok
	d.requestedVersion, ok = safeRequestedVersion(v.Target, v.RequestedTargetVersion)
	withheld = withheld || !ok
	if len(v.Findings) > 15 || len(v.ComparisonSources) > 2 || len(v.Limitations) != 1 {
		return r, false
	}
	if d.findings, ok = safeFindings(v.Findings); !ok {
		return r, false
	}
	if withheld {
		d.findings = append(d.findings, Finding{code: "redact-version-withheld"})
	}
	if len(d.findings) > 15 {
		return r, false
	}
	sort.Slice(d.findings, func(i, j int) bool { return d.findings[i].code < d.findings[j].code })
	if d.limitations, ok = safeLimitations(v.Limitations); !ok {
		return r, false
	}
	for _, s := range v.ComparisonSources {
		if !comparisonSide(s.Side) || !validDigest(s.Digest) {
			return r, false
		}
		for _, prior := range r.sources {
			if prior.side == s.Side && prior.digest != Digest(s.Digest) {
				return r, false
			}
		}
		r.sources = append(r.sources, ComparisonSource{s.Side, Digest(s.Digest)})
	}
	slices.SortFunc(r.sources, func(a, b ComparisonSource) int { return cmp.Compare(a.side, b.side) })
	r.sources = slices.Compact(r.sources)
	if v.Comparison == nil {
		return r, d.completeness == CompletenessIncomplete && len(r.sources) == 0 && appFindingCount(d.findings) == 0 && ((d.category == CategoryUnsupportedOrIncomplete && (d.supportStatus == SupportUnresolved || d.supportStatus == SupportUnsupported)) || (d.category == CategoryOperationalFailure && (d.supportStatus == SupportUnresolved || d.supportStatus == SupportSupported)))
	}
	c := v.Comparison
	if len(c.Reasons) > 24 || len(c.Differences) > 9 || d.supportStatus != SupportSupported || (d.category == CategoryOperationalFailure && len(r.sources) > 1) {
		return r, false
	}
	r.status = ComparisonStatus(c.Outcome)
	conflict := false
	for _, reason := range c.Reasons {
		if !comparisonSide(reason.Side) || !operandReason(reason.Code) {
			return r, false
		}
		conflict = conflict || reason.Code == compare.ConflictingDuplicate
		r.reasons = append(r.reasons, ComparisonReason{reason.Side, reason.Code})
	}
	slices.SortFunc(r.reasons, compareReasons)
	r.reasons = slices.Compact(r.reasons)
	semantic := false
	for _, diff := range c.Differences {
		k := diff.Key
		if (k.Capability != "read" && k.Capability != "edit" && k.Capability != "bash") || k.Scope != "opencode" || k.Matcher != "*" {
			return r, false
		}
		only := diff.Outcome == compare.TargetOnly && comparisonSide(diff.OnlyIn) && diff.Reason.Side == diff.OnlyIn && diff.Reason.Code == compare.PermissionOnly
		matched := diff.Outcome == compare.NotComparable && diff.OnlyIn == "" && diff.Reason.Side == compare.Target && (diff.Reason.Code == compare.EffectDiffers || diff.Reason.Code == compare.ConditionsDiffer || diff.Reason.Code == compare.ExtensionsDiffer)
		if !only && !matched {
			return r, false
		}
		for _, prior := range r.differences {
			if prior.capability == Capability(k.Capability) && (prior.onlyIn != diff.OnlyIn || prior.outcome != diff.Outcome) {
				return r, false
			}
		}
		semantic = semantic || matched
		r.differences = append(r.differences, ComparisonDifference{Capability(k.Capability), Scope(k.Scope), Matcher(k.Matcher), diff.OnlyIn, diff.Outcome, ComparisonReason{diff.Reason.Side, diff.Reason.Code}})
	}
	slices.SortFunc(r.differences, func(a, b ComparisonDifference) int {
		return cmp.Or(cmp.Compare(a.capability, b.capability), cmp.Compare(a.scope, b.scope), cmp.Compare(a.matcher, b.matcher), cmp.Compare(a.onlyIn, b.onlyIn), cmp.Compare(a.outcome, b.outcome), compareReasons(a.reason, b.reason))
	})
	r.differences = slices.Compact(r.differences)
	if c.Outcome == compare.Equivalent {
		return r, d.category == CategoryCompleteNoFindings && d.completeness == CompletenessComplete && len(r.sources) == 2 && len(r.reasons)+len(r.differences)+len(d.findings) == 0 && v.RequestedTargetVersion == "1.18.27"
	}
	if d.completeness != CompletenessIncomplete || (d.category != CategoryUnsupportedOrIncomplete && d.category != CategoryOperationalFailure) {
		return r, false
	}
	if d.category == CategoryOperationalFailure {
		return r, c.Outcome == compare.NotComparable && !conflict && len(r.reasons) > 0 && len(r.differences) == 0
	}
	switch c.Outcome {
	case compare.TargetOnly:
		return r, d.category == CategoryUnsupportedOrIncomplete && len(r.differences) > 0 && !semantic && len(r.reasons) == 0 && appFindingCount(d.findings) == 0
	case compare.Ambiguous:
		return r, d.category == CategoryUnsupportedOrIncomplete && conflict && len(r.differences) == 0 && appFindingCount(d.findings) == 0
	case compare.NotComparable:
		return r, !conflict && ((len(r.reasons) > 0 && len(r.differences) == 0) || (len(r.reasons) == 0 && semantic && appFindingCount(d.findings) == 0))
	}
	return r, false
}
func comparisonSide(s compare.Side) bool { return s == compare.Reference || s == compare.Target }
func compareReasons(a, b ComparisonReason) int {
	return cmp.Or(cmp.Compare(a.side, b.side), cmp.Compare(a.code, b.code))
}
func operandReason(c compare.ReasonCode) bool {
	switch c {
	case compare.InvalidOperandMetadata, compare.IncompatibleCanonicalization, compare.IncompatibleCoverage, compare.OperandIncomplete, compare.InvalidIdentity, compare.TraceIncomplete, compare.UnresolvedPermission, compare.UnsupportedEffect, compare.ConflictingDuplicate:
		return true
	}
	return false
}
