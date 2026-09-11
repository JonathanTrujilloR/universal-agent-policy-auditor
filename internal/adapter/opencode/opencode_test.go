package opencode

import (
	"os"
	"reflect"
	"testing"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/source"
	"github.com/jkelevra/universal-agent-policy-auditor/support"
)

func TestDiscoveryPlanIsDeclarativeAndBounded(t *testing.T) {
	plan := DiscoveryPlan("/repo", []string{"/repo/custom-opencode.json"})
	if len(plan.Roots) != 1 || plan.Roots[0] != "/repo" {
		t.Fatalf("roots=%v", plan.Roots)
	}
	if len(plan.Sources) != 1 || plan.Sources[0] != "/repo/custom-opencode.json" {
		t.Fatalf("sources=%v", plan.Sources)
	}
	if plan.MaxFiles != 8 || plan.MaxBytes != 262144 {
		t.Fatalf("limits files=%d bytes=%d", plan.MaxFiles, plan.MaxBytes)
	}
	if !contains(plan.EnvKeys, "OPENCODE_CONFIG") || !contains(plan.EnvKeys, "XDG_CONFIG_HOME") || !contains(plan.EnvKeys, "HOME") {
		t.Fatalf("env keys=%v", plan.EnvKeys)
	}
}

func TestResolveRequiresExactVersionBeforeSupportValidation(t *testing.T) {
	entry := support.Entry{Target: "opencode", Version: "synthetic", Construct: "permission", Fixture: "fixture", Evidence: "evidence", PermissionBearing: true, DefaultsKnown: true, MatcherKnown: true, ConditionsKnown: true}
	report := ResolveWithOptions(validSource(`{"permission":{"read":"allow","edit":"deny","bash":"ask"}}`), Options{Version: "synthetic", Support: support.Matrix{Entries: []support.Entry{entry}, Registry: support.Registry{Fixtures: map[string]bool{"fixture": true}, Evidence: map[string]bool{"evidence": true}}}})
	assertIncomplete(t, report, "opencode-unsupported-version")
}

func TestResolveRequiresExactlyOneSelectedSourceWithIdentityPathAndBoundedData(t *testing.T) {
	valid := validSource(`{"permission":{"read":"allow","edit":"deny","bash":"ask"}}`)
	tests := []struct {
		sources []source.SelectedSource
		code    string
	}{
		{append(valid, source.SelectedSource{Identity: "other", Path: "/repo/other.json", Data: []byte(`{}`)}), "opencode-source-selection-incomplete"},
		{[]source.SelectedSource{{Identity: " ", Path: "/repo/opencode.json", Data: []byte(`{}`)}}, "opencode-source-selection-incomplete"},
		{[]source.SelectedSource{{Identity: "project", Path: "\t", Data: []byte(`{}`)}}, "opencode-source-selection-incomplete"},
		{[]source.SelectedSource{{Identity: "project", Path: "/repo/opencode.json", Data: make([]byte, maxBytes+1)}}, "opencode-source-data-too-large"},
	}
	for _, tt := range tests {
		assertIncomplete(t, ResolveWithOptions(tt.sources, productionOptions(t)), tt.code)
	}
}

func TestResolveConformsExactOpenCode11827ScalarFixture(t *testing.T) {
	fixture, err := os.ReadFile("../../../support/testdata/opencode-1.18.27-permission-legacy-scalar.json")
	if err != nil {
		t.Fatal(err)
	}
	report := ResolveWithOptions(validSource(string(fixture)), productionOptions(t))
	if report.Completeness != model.CompletenessComplete || len(report.Findings) != 0 {
		t.Fatalf("report=%+v", report)
	}
	if got, want := permissionSummary(report.Permissions), []string{"bash:ask:*", "edit:deny:*", "read:allow:*"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("permissions=%v", got)
	}
	wantEffects := map[string]model.Effect{"bash": model.EffectAsk, "edit": model.EffectDeny, "read": model.EffectAllow}
	for _, permission := range report.Permissions {
		steps := permission.Trace().Steps()
		wantLocation := "/repo/opencode.json #permission." + permission.Capability()
		if permission.Effect() != wantEffects[permission.Capability()] || permission.Matcher() != "*" || permission.Scope() != model.Scope("opencode") || len(permission.Conditions()) != 0 || len(steps) != 1 || steps[0].Kind != model.TraceRule || steps[0].After != permission.Effect() || len(permission.Provenance()) != 1 || permission.Provenance()[0].Source() != "opencode" || permission.Provenance()[0].Location() != wantLocation {
			t.Fatalf("permission=%+v trace=%+v wantLocation=%q", permission, steps, wantLocation)
		}
	}
}

func TestResolveRejectsOldInventedObjectMatcherConditionFixture(t *testing.T) {
	fixture, err := os.ReadFile("../../../testdata/conformance/opencode/permissions-precedence-runtime.json")
	if err != nil {
		t.Fatal(err)
	}
	assertIncomplete(t, ResolveWithOptions(validSource(string(fixture)), productionOptions(t)), "opencode-permission-shape-unsupported")
}

func TestResolveStrictlyRejectsUnsupportedJSONShapesAtomically(t *testing.T) {
	for _, data := range []string{
		`{"permission":{"r\u0065ad":"allow","read":"allow","edit":"deny","bash":"ask"}}`,
		`{"permission":{"read":"allow","edit":"deny","bash":"ask"}} {}`,
		`{"permission":null}`,
		`{"permission":{"read":"allow","edit":"deny","bash":["ask"]}}`,
		`{"permission":{"read":"allow","edit":"deny","bash":"ASK"}}`,
		`{"permission":{"read":"allow","edit":"deny","båsh":"ask"}}`,
		`{"permission":{"read":"allow","edit":"deny","bash":"ask"},"tools":{}}`,
		`{"permission":{"read":"allow","edit":"deny"}}`,
	} {
		report := ResolveWithOptions(validSource(data), productionOptions(t))
		assertIncomplete(t, report, "opencode-permission-shape-unsupported")
	}
}

func TestResolveDeduplicatesRequestedCapabilitiesAndRejectsEmptyRequestedName(t *testing.T) {
	options := productionOptions(t)
	options.RequestedCapabilities = []string{"deploy", " read ", "deploy", " ", ""}
	report := ResolveWithOptions(validSource(`{"permission":{"read":"allow","edit":"deny","bash":"ask"}}`), options)
	want := []string{"bash:ask:*", "deploy:unresolved:*", "edit:deny:*", "read:allow:*"}
	if !hasFinding(report.Findings, "opencode-invalid-requested-capability") || !hasFinding(report.Findings, "opencode-unresolved-default") || !reflect.DeepEqual(permissionSummary(report.Permissions), want) {
		t.Fatalf("report=%+v", report)
	}
}

func validSource(data string) []source.SelectedSource {
	return []source.SelectedSource{{Identity: "project", Path: "/repo/opencode.json ", Data: []byte(data)}}
}
func permissionSummary(permissions []model.Permission) []string {
	out := make([]string, len(permissions))
	for i, permission := range permissions {
		out[i] = permission.Capability() + ":" + string(permission.Effect()) + ":" + permission.Matcher()
	}
	return out
}
func assertIncomplete(t *testing.T, report Report, code string) {
	t.Helper()
	if report.Completeness != model.CompletenessIncomplete || !hasFinding(report.Findings, code) || len(report.Permissions) != 0 {
		t.Fatalf("report=%+v", report)
	}
}
func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
func hasFinding(findings []model.Finding, code string) bool {
	for _, finding := range findings {
		if finding.Code() == code {
			return true
		}
	}
	return false
}

func productionOptions(t *testing.T) Options {
	t.Helper()
	matrix, err := support.Load(os.DirFS("../../../support"))
	if err != nil {
		t.Fatal(err)
	}
	return Options{Version: "1.18.27", Support: matrix}
}
