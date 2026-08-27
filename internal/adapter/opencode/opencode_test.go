package opencode

import (
	"os"
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

func TestResolveFailsClosedWithoutSupportMatrixEntry(t *testing.T) {
	report := Resolve([]source.SelectedSource{{Identity: "project", Path: "/repo/opencode.json", Data: []byte(`{"permission":{"bash":"allow"}}`)}})
	if report.Completeness != model.CompletenessIncomplete || !hasFinding(report.Findings, "opencode-unsupported-version") || len(report.Permissions) != 0 {
		t.Fatalf("report=%+v", report)
	}
}

func TestResolveWithEvidenceBackedFixtureModelsPrecedenceAndRuntimeLimit(t *testing.T) {
	fixture, err := os.ReadFile("../../../testdata/conformance/opencode/permissions-precedence-runtime.json")
	if err != nil {
		t.Fatal(err)
	}
	report := ResolveWithOptions([]source.SelectedSource{
		{Identity: "01-global", Path: "/repo/opencode.global.json", Data: []byte(`{"permission":{"bash":"ask","edit":"deny"}}`)},
		{Identity: "02-project", Path: "/repo/opencode.json", Data: fixture},
	}, supportedOptions())
	if report.Completeness != model.CompletenessIncomplete || !hasFinding(report.Findings, "opencode-runtime-limit") {
		t.Fatalf("report=%+v", report)
	}
	bash := permissionByCapability(report.Permissions, "bash")
	if bash.Effect() != model.EffectAllow || bash.Scope() != model.Scope("opencode") || bash.Matcher() != "bash" {
		t.Fatalf("bash=%+v", bash)
	}
	steps := bash.Trace().Steps()
	if len(steps) != 2 || steps[0].Kind != model.TraceRule || steps[1].Kind != model.TracePrecedence || steps[1].After != model.EffectAllow {
		t.Fatalf("trace=%+v", steps)
	}
	webfetch := permissionByCapability(report.Permissions, "webfetch")
	webSteps := webfetch.Trace().Steps()
	if webfetch.Effect() != model.EffectConditional || webfetch.Conditions()[0] != model.Condition("interactive approval") || webSteps[len(webSteps)-1].Kind != model.TraceRuntime {
		t.Fatalf("webfetch=%+v trace=%+v", webfetch, webSteps)
	}
}

func TestResolveFailsClosedForAmbiguousSourceOrder(t *testing.T) {
	report := ResolveWithOptions([]source.SelectedSource{
		{Identity: "same", Path: "/repo/a.json", Data: []byte(`{"permission":{"bash":"ask"}}`)},
		{Identity: "same", Path: "/repo/b.json", Data: []byte(`{"permission":{"bash":"allow"}}`)},
	}, supportedOptions())
	if report.Completeness != model.CompletenessIncomplete || !hasFinding(report.Findings, "opencode-ambiguous-source-order") || len(report.Permissions) != 0 {
		t.Fatalf("report=%+v", report)
	}
}

func TestResolveUsesUnresolvedDefaultForRequestedMissingCapability(t *testing.T) {
	options := supportedOptions()
	options.RequestedCapabilities = []string{"edit"}
	report := ResolveWithOptions([]source.SelectedSource{{Identity: "project", Path: "/repo/opencode.json", Data: []byte(`{"permission":{"bash":"allow"}}`)}}, options)
	permission := permissionByCapability(report.Permissions, "edit")
	steps := permission.Trace().Steps()
	if report.Completeness != model.CompletenessIncomplete || permission.Effect() != model.EffectUnresolved || len(steps) != 1 || steps[0].Kind != model.TraceDefault || !hasFinding(report.Findings, "opencode-unresolved-default") {
		t.Fatalf("report=%+v permission=%+v trace=%+v", report, permission, steps)
	}
}

func TestResolveFailsClosedForMalformedUnknownAndRuntimeLimits(t *testing.T) {
	tests := []struct {
		name string
		data string
		code string
	}{
		{"malformed", `{"permission":`, "opencode-malformed"},
		{"unknown effect", `{"permission":{"bash":"maybe"}}`, "opencode-unknown-effect"},
		{"unknown permission object", `{"permission":{"bash":{"mode":"allow"}}}`, "opencode-unknown-permission"},
		{"permission-bearing unknown top-level", `{"permissions":{"bash":"allow"}}`, "opencode-unknown-construct"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := ResolveWithOptions([]source.SelectedSource{{Identity: tt.name, Path: "/repo/opencode.json", Data: []byte(tt.data)}}, supportedOptions())
			if report.Completeness != model.CompletenessIncomplete || !hasFinding(report.Findings, tt.code) {
				t.Fatalf("report=%+v", report)
			}
		})
	}
}

func permissionByCapability(permissions []model.Permission, capability string) model.Permission {
	for _, permission := range permissions {
		if permission.Capability() == capability {
			return permission
		}
	}
	return model.Permission{}
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

func supportedOptions() Options {
	entry := support.Entry{Target: "opencode", Version: "local-fixture", Construct: "permission", Fixture: "opencode-runtime", Evidence: "opencode-doc", PermissionBearing: true, DefaultsKnown: true, MatcherKnown: true, ConditionsKnown: true}
	return Options{Version: "local-fixture", Support: support.Matrix{Entries: []support.Entry{entry}, Registry: support.Registry{Fixtures: map[string]bool{"opencode-runtime": true}, Evidence: map[string]bool{"opencode-doc": true}}}}
}
