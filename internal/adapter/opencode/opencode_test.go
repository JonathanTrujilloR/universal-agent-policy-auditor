package opencode

import (
	"testing"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/source"
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

func TestResolveMergesSourcesWithLastRulePrecedence(t *testing.T) {
	report := Resolve([]source.SelectedSource{
		{Identity: "global", Path: "/repo/opencode.global.json", Data: []byte(`{"permission":{"bash":"ask","edit":"deny"}}`)},
		{Identity: "project", Path: "/repo/opencode.json", Data: []byte(`{"permission":{"bash":"allow","webfetch":{"effect":"deny","match":"https://example.com/*","condition":"network disabled"}}}`)},
	})
	if report.Completeness != model.CompletenessComplete || len(report.Findings) != 0 {
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
	if webfetch.Effect() != model.EffectConditional || webfetch.Conditions()[0] != model.Condition("network disabled") || webfetch.Matcher() != "https://example.com/*" {
		t.Fatalf("webfetch=%+v", webfetch)
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
			report := Resolve([]source.SelectedSource{{Identity: tt.name, Path: "/repo/opencode.json", Data: []byte(tt.data)}})
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
