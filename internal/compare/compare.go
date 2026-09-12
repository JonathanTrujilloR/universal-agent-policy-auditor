// Package compare compares normalized policy facts without inferring policy ordering.
package compare

import (
	"cmp"
	"slices"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
)

type Outcome string

const (
	Equivalent    Outcome = "equivalent"
	TargetOnly    Outcome = "target-only"
	Ambiguous     Outcome = "ambiguous"
	NotComparable Outcome = "not-comparable"
)

type Side string

const (
	Reference Side = "reference"
	Target    Side = "target"
)

type ReasonCode string

const (
	InvalidOperandMetadata       ReasonCode = "invalid-operand-metadata"
	IncompatibleCanonicalization ReasonCode = "incompatible-canonicalization"
	IncompatibleCoverage         ReasonCode = "incompatible-coverage"
	OperandIncomplete            ReasonCode = "operand-incomplete"
	InvalidIdentity              ReasonCode = "invalid-identity"
	TraceIncomplete              ReasonCode = "trace-incomplete"
	UnresolvedPermission         ReasonCode = "unresolved-permission"
	UnsupportedEffect            ReasonCode = "unsupported-effect"
	ConflictingDuplicate         ReasonCode = "conflicting-duplicate"
	EffectDiffers                ReasonCode = "effect-differs"
	ConditionsDiffer             ReasonCode = "conditions-differ"
	ExtensionsDiffer             ReasonCode = "extensions-differ"
	PermissionOnly               ReasonCode = "permission-only"
)

type Operand struct {
	Canonicalization, Coverage string
	Completeness               model.Completeness
	Permissions                []model.Permission
}
type Key struct {
	Capability string
	Scope      model.Scope
	Matcher    string
}
type Reason struct {
	Side Side
	Code ReasonCode
}
type Difference struct {
	Key     Key
	OnlyIn  Side
	Outcome Outcome
	Reason  Reason
}
type Result struct {
	Outcome     Outcome
	Reasons     []Reason
	Differences []Difference
}

// Compare returns fresh value slices. Metadata failure prevents fact comparison.
// Matched mismatches are attributed to target relative to reference; OnlyIn is empty.
func Compare(reference, target Operand) Result {
	result := Result{Outcome: Equivalent}
	reasons := map[Reason]bool{}
	add := func(side Side, code ReasonCode, invalid bool) {
		if invalid {
			reasons[Reason{side, code}] = true
		}
	}
	for i, o := range []Operand{reference, target} {
		side := []Side{Reference, Target}[i]
		add(side, InvalidOperandMetadata, o.Canonicalization == "" || o.Coverage == "")
		add(side, OperandIncomplete, o.Completeness != model.CompletenessComplete)
		add(side, IncompatibleCanonicalization, reference.Canonicalization != "" && target.Canonicalization != "" && reference.Canonicalization != target.Canonicalization)
		add(side, IncompatibleCoverage, reference.Coverage != "" && target.Coverage != "" && reference.Coverage != target.Coverage)
	}
	if len(reasons) == 0 {
		sets := [2]map[Key]model.Permission{{}, {}}
		keys := map[Key]bool{}
		for i, o := range []Operand{reference, target} {
			side := []Side{Reference, Target}[i]
			for _, p := range o.Permissions {
				key := Key{p.Capability(), p.Scope(), p.Matcher()}
				effect, trace := p.Effect(), p.Trace()
				add(side, InvalidIdentity, key.Capability == "" || key.Scope == "" || key.Matcher == "")
				add(side, TraceIncomplete, trace.Status() != model.CompletenessComplete)
				add(side, UnresolvedPermission, effect == model.EffectUnresolved || len(trace.Unresolved()) != 0)
				add(side, UnsupportedEffect, effect != model.EffectUnresolved && effect != model.EffectAllow && effect != model.EffectDeny && effect != model.EffectAsk)
				if old, ok := sets[i][key]; ok {
					add(side, ConflictingDuplicate, len(semanticDifferences(old, p)) != 0)
				} else {
					sets[i][key] = p
				}
				keys[key] = true
			}
		}
		// Invalid facts are never used to construct differences.
		if len(reasons) == 0 {
			for key := range keys {
				a, hasA := sets[0][key]
				b, hasB := sets[1][key]
				if !hasA || !hasB {
					side := Reference
					if !hasA {
						side = Target
					}
					result.Differences = append(result.Differences, Difference{key, side, TargetOnly, Reason{side, PermissionOnly}})
				} else {
					for _, code := range semanticDifferences(a, b) {
						result.Differences = append(result.Differences, Difference{key, "", NotComparable, Reason{Target, code}})
					}
				}
			}
		}
	}
	for reason := range reasons {
		result.Reasons = append(result.Reasons, reason)
		result.Outcome = NotComparable
	}
	for _, reason := range result.Reasons {
		if reason.Code == ConflictingDuplicate {
			result.Outcome = Ambiguous
		}
	}
	slices.SortFunc(result.Reasons, func(a, b Reason) int { return cmp.Or(cmp.Compare(a.Side, b.Side), cmp.Compare(a.Code, b.Code)) })
	slices.SortFunc(result.Differences, func(a, b Difference) int {
		return cmp.Or(cmp.Compare(a.Key.Capability, b.Key.Capability), cmp.Compare(a.Key.Scope, b.Key.Scope), cmp.Compare(a.Key.Matcher, b.Key.Matcher), cmp.Compare(a.OnlyIn, b.OnlyIn), cmp.Compare(a.Outcome, b.Outcome), cmp.Compare(a.Reason.Side, b.Reason.Side), cmp.Compare(a.Reason.Code, b.Reason.Code))
	})
	for _, d := range result.Differences {
		if d.Outcome == NotComparable {
			result.Outcome = NotComparable
		} else if result.Outcome == Equivalent {
			result.Outcome = TargetOnly
		}
	}
	return result
}
func semanticDifferences(a, b model.Permission) []ReasonCode {
	var codes []ReasonCode
	if a.Effect() != b.Effect() {
		codes = append(codes, EffectDiffers)
	}
	if !slices.Equal(a.Conditions(), b.Conditions()) {
		codes = append(codes, ConditionsDiffer)
	}
	if !slices.Equal(a.Extensions(), b.Extensions()) {
		codes = append(codes, ExtensionsDiffer)
	}
	return codes
}
