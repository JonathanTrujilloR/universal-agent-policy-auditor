package compare

import (
	"reflect"
	"testing"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
)

func permission(key string, effect model.Effect, conditions []string, extensions []model.Extension) model.Permission {
	return model.NewPermission(key, effect, "workspace", "*", conditions, extensions, model.NewResolutionTrace(nil, nil, model.CompletenessComplete))
}
func operand(ps ...model.Permission) Operand {
	return Operand{"v1", "all", model.CompletenessComplete, ps}
}
func TestCompare(t *testing.T) {
	a := permission("a", model.EffectAllow, nil, nil)
	b := permission("b", model.EffectDeny, nil, nil)
	tests := []struct {
		name        string
		left, right Operand
		outcome     Outcome
		code        ReasonCode
		differences int
	}{
		{"empty", operand(), operand(), Equivalent, "", 0},
		{"reordered duplicates", operand(a, b, a), operand(b, a), Equivalent, "", 0},
		{"reference only", operand(a), operand(), TargetOnly, PermissionOnly, 1},
		{"target only", operand(), operand(a), TargetOnly, PermissionOnly, 1},
		{"effect", operand(a), operand(permission("a", model.EffectAsk, nil, nil)), NotComparable, EffectDiffers, 1},
		{"conditions", operand(a), operand(permission("a", model.EffectAllow, []string{"x"}, nil)), NotComparable, ConditionsDiffer, 1},
		{"extensions", operand(a), operand(permission("a", model.EffectAllow, nil, []model.Extension{{Target: "x"}})), NotComparable, ExtensionsDiffer, 1},
		{"duplicate", operand(a, permission("a", model.EffectDeny, nil, nil)), operand(a), Ambiguous, ConflictingDuplicate, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compare(tt.left, tt.right)
			if got.Outcome != tt.outcome || len(got.Differences) != tt.differences {
				t.Fatalf("%+v", got)
			}
			if tt.code != "" {
				found := false
				for _, r := range got.Reasons {
					found = found || r.Code == tt.code
				}
				for _, d := range got.Differences {
					found = found || d.Reason.Code == tt.code
				}
				if !found {
					t.Fatalf("missing %s: %+v", tt.code, got)
				}
			}
			if tt.name == "reference only" && got.Differences[0].OnlyIn != Reference {
				t.Fatal(got)
			}
			if tt.name == "target only" && got.Differences[0].OnlyIn != Target {
				t.Fatal(got)
			}
		})
	}
}
func TestInvalid(t *testing.T) {
	a := permission("a", model.EffectAllow, nil, nil)
	cases := []struct {
		name   string
		change func(*Operand)
		code   ReasonCode
	}{
		{"metadata", func(o *Operand) { o.Canonicalization = "" }, InvalidOperandMetadata},
		{"coverage empty", func(o *Operand) { o.Coverage = "" }, InvalidOperandMetadata},
		{"profile", func(o *Operand) { o.Canonicalization = "v2" }, IncompatibleCanonicalization},
		{"coverage", func(o *Operand) { o.Coverage = "other" }, IncompatibleCoverage},
		{"incomplete", func(o *Operand) { o.Completeness = model.CompletenessIncomplete }, OperandIncomplete},
		{"identity", func(o *Operand) { o.Permissions[0] = permission("", model.EffectAllow, nil, nil) }, InvalidIdentity},
		{"scope", func(o *Operand) {
			o.Permissions[0] = model.NewPermission("a", model.EffectAllow, "", "*", nil, nil, a.Trace())
		}, InvalidIdentity},
		{"matcher", func(o *Operand) {
			o.Permissions[0] = model.NewPermission("a", model.EffectAllow, "workspace", "", nil, nil, a.Trace())
		}, InvalidIdentity},
		{"trace", func(o *Operand) {
			o.Permissions[0] = model.NewPermission("a", model.EffectAllow, "workspace", "*", nil, nil, model.ResolutionTrace{})
		}, TraceIncomplete},
		{"unresolved", func(o *Operand) {
			o.Permissions[0] = model.NewPermission("a", model.EffectAllow, "workspace", "*", nil, nil, model.NewResolutionTrace(nil, []string{"secret"}, model.CompletenessComplete))
		}, UnresolvedPermission},
		{"unresolved effect", func(o *Operand) { o.Permissions[0] = permission("a", model.EffectUnresolved, nil, nil) }, UnresolvedPermission},
		{"unsupported effect", func(o *Operand) { o.Permissions[0] = permission("a", model.EffectInherited, nil, nil) }, UnsupportedEffect},
	}
	for _, tt := range cases {
		for _, side := range []Side{Reference, Target} {
			t.Run(tt.name+string(side), func(t *testing.T) {
				left, right := operand(a), operand(a)
				if side == Reference {
					tt.change(&left)
				} else {
					tt.change(&right)
				}
				got := Compare(left, right)
				if got.Outcome != NotComparable || len(got.Differences) != 0 {
					t.Fatal(got)
				}
				found := false
				for _, r := range got.Reasons {
					found = found || (r.Code == tt.code && r.Side == side)
				}
				if !found {
					t.Fatal(got)
				}
			})
		}
	}
}
func TestOrderingAndIndependence(t *testing.T) {
	a, b := permission("a", model.EffectAllow, nil, nil), permission("b", model.EffectDeny, nil, nil)
	left, right := operand(b, a), operand(permission("c", model.EffectAsk, nil, nil))
	before := append([]model.Permission(nil), left.Permissions...)
	want := Compare(left, right)
	for i := 0; i < 100; i++ {
		l := operand(a, b)
		if i%2 == 0 {
			l = operand(b, a)
		}
		if got := Compare(l, right); !reflect.DeepEqual(got, want) {
			t.Fatal(got, want)
		}
	}
	if !reflect.DeepEqual(left.Permissions, before) {
		t.Fatal("input mutated")
	}
	want.Differences[0].Key.Capability = "changed"
	want.Differences[0].Reason.Code = InvalidIdentity
	if Compare(left, right).Differences[0].Key.Capability != "a" {
		t.Fatal("result aliases state")
	}
	invalid := operand(a)
	invalid.Coverage = ""
	result := Compare(invalid, right)
	result.Reasons[0].Code = EffectDiffers
	if Compare(invalid, right).Reasons[0].Code == EffectDiffers {
		t.Fatal("reason aliases state")
	}
}
func TestSemanticDetails(t *testing.T) {
	trace := model.NewResolutionTrace([]model.TraceStep{{Kind: model.TraceRule, After: model.EffectAllow}}, nil, model.CompletenessComplete)
	a := permission("a", model.EffectAllow, []string{"x", "y"}, []model.Extension{{Key: "x"}, {Key: "y"}})
	b := model.NewPermissionWithProvenance("a", model.EffectAllow, "workspace", "*", a.Conditions(), []model.Provenance{model.NewProvenance("source", "location", "raw")}, a.Extensions(), trace)
	if Compare(operand(a), operand(b)).Outcome != Equivalent {
		t.Fatal("derivation participated")
	}
	c := permission("a", model.EffectDeny, []string{"y", "x"}, []model.Extension{{Key: "y"}, {Key: "x"}})
	got := Compare(operand(a), operand(c))
	if len(got.Differences) != 3 {
		t.Fatal(got)
	}
	for i, code := range []ReasonCode{ConditionsDiffer, EffectDiffers, ExtensionsDiffer} {
		if got.Differences[i].Reason.Code != code {
			t.Fatal(got)
		}
	}
	if Compare(operand(a, c), operand(permission("z", model.EffectInherited, nil, nil))).Outcome != Ambiguous {
		t.Fatal("ambiguity must win")
	}
	// Structured identities must not collide when values contain separators.
	x := model.NewPermission("a/b", model.EffectAllow, "c", "d", nil, nil, trace)
	y := model.NewPermission("a", model.EffectAllow, "b/c", "d", nil, nil, trace)
	if len(Compare(operand(x), operand(y)).Differences) != 2 {
		t.Fatal("identity collision")
	}
}
