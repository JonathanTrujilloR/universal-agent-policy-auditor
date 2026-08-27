package claudecode

import (
	"os"
	"slices"
	"testing"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/source"
	"github.com/jkelevra/universal-agent-policy-auditor/support"
)

func TestDiscoveryPlanIsDeclarativeAndBounded(t *testing.T) {
	plan := DiscoveryPlan("/repo", []string{"/repo/.claude/settings.local.json"})
	if len(plan.Roots) != 1 || plan.Roots[0] != "/repo" || len(plan.Sources) != 1 || plan.Sources[0] != "/repo/.claude/settings.local.json" || plan.MaxFiles != 8 || plan.MaxBytes != 262144 || !slices.Contains(plan.EnvKeys, "CLAUDE_CONFIG_DIR") || !slices.Contains(plan.EnvKeys, "XDG_CONFIG_HOME") || !slices.Contains(plan.EnvKeys, "HOME") {
		t.Fatalf("plan=%+v", plan)
	}
}

func TestResolveFailsClosedWithoutSupportMatrixEntry(t *testing.T) {
	report := Resolve([]source.SelectedSource{{Identity: "project", Path: "/repo/.claude/settings.json", Data: []byte(`{"permissions":{"allow":["Bash(git status:*)"]}}`)}})
	if report.Completeness != model.CompletenessIncomplete || !hasFinding(report.Findings, "claudecode-unsupported-version") || len(report.Permissions) != 0 {
		t.Fatalf("report=%+v", report)
	}
}

func TestResolveWithEvidenceBackedFixtureModelsDenyPrecedenceAndRuntimeLimit(t *testing.T) {
	fixture, err := os.ReadFile("../../../testdata/conformance/claudecode/permissions-precedence-runtime.json")
	if err != nil {
		t.Fatal(err)
	}
	report := ResolveWithOptions([]source.SelectedSource{
		{Identity: "01-user", Path: "/repo/user-settings.json", Data: []byte(`{"permissions":{"allow":["Bash(git status:*)"]}}`)},
		{Identity: "02-project", Path: "/repo/.claude/settings.json", Data: fixture},
	}, supportedOptions())
	if report.Completeness != model.CompletenessIncomplete || !hasFinding(report.Findings, "claudecode-runtime-limit") {
		t.Fatalf("report=%+v", report)
	}
	bash := permissionByMatcher(t, report.Permissions, "Bash(git status:*)")
	if bash.Effect() != model.EffectDeny || bash.Scope() != model.Scope("claudecode") {
		t.Fatalf("bash=%+v", bash)
	}
	steps := bash.Trace().Steps()
	if len(steps) != 2 || steps[0].Kind != model.TraceRule || steps[1].Kind != model.TracePrecedence || steps[1].After != model.EffectDeny {
		t.Fatalf("trace=%+v", steps)
	}
	webfetch := permissionByMatcher(t, report.Permissions, "WebFetch(*)")
	webSteps := webfetch.Trace().Steps()
	if webfetch.Effect() != model.EffectConditional || webfetch.Conditions()[0] != model.Condition("interactive approval") || webSteps[len(webSteps)-1].Kind != model.TraceRuntime {
		t.Fatalf("webfetch=%+v trace=%+v", webfetch, webSteps)
	}
}

func TestResolveFailsClosedForAmbiguousSourceOrder(t *testing.T) {
	report := ResolveWithOptions([]source.SelectedSource{
		{Identity: "same", Path: "/repo/a.json", Data: []byte(`{"permissions":{"ask":["Bash(git status:*)"]}}`)},
		{Identity: "same", Path: "/repo/b.json", Data: []byte(`{"permissions":{"deny":["Bash(git status:*)"]}}`)},
	}, supportedOptions())
	if report.Completeness != model.CompletenessIncomplete || !hasFinding(report.Findings, "claudecode-ambiguous-source-order") || len(report.Permissions) != 0 {
		t.Fatalf("report=%+v", report)
	}
}

func TestResolveUsesUnresolvedDefaultForRequestedMissingCapability(t *testing.T) {
	options := supportedOptions()
	options.RequestedCapabilities = []string{"Edit"}
	report := ResolveWithOptions([]source.SelectedSource{{Identity: "project", Path: "/repo/.claude/settings.json", Data: []byte(`{"permissions":{"allow":["Bash(git status:*)"]}}`)}}, options)
	permission := permissionByMatcher(t, report.Permissions, "Edit")
	steps := permission.Trace().Steps()
	if report.Completeness != model.CompletenessIncomplete || permission.Effect() != model.EffectUnresolved || len(steps) != 1 || steps[0].Kind != model.TraceDefault || !hasFinding(report.Findings, "claudecode-unresolved-default") {
		t.Fatalf("report=%+v permission=%+v trace=%+v", report, permission, steps)
	}
}

func TestResolveFailsClosedForMalformedUnknownAndRuntimeConstructs(t *testing.T) {
	tests := []struct{ name, data, code string }{
		{"malformed", `{"permissions":`, "claudecode-malformed"},
		{"unknown effect bucket", `{"permissions":{"approve":["Bash(*)"]}}`, "claudecode-unknown-effect"},
		{"unknown permission entry", `{"permissions":{"allow":[{"tool":"Bash"}]}}`, "claudecode-unknown-permission"},
		{"runtime hook construct", `{"hooks":{"PreToolUse":[]}}`, "claudecode-runtime-construct"},
		{"top-level runtime field", `{"permissions":{"allow":["Bash(git status:*)"]},"mcpServers":{}}`, "claudecode-unknown-semantic"},
		{"rule-level runtime field", `{"permissions":{"allow":[{"match":"Bash(git status:*)","sandbox":"workspace-write"}]}}`, "claudecode-unknown-semantic"},
		{"malformed condition bool", `{"permissions":{"allow":[{"match":"WebFetch(*)","condition":true}]}}`, "claudecode-malformed-semantic"},
		{"malformed condition object", `{"permissions":{"allow":[{"match":"WebFetch(*)","condition":{}}]}}`, "claudecode-malformed-semantic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := ResolveWithOptions([]source.SelectedSource{{Identity: tt.name, Path: "/repo/.claude/settings.json", Data: []byte(tt.data)}}, supportedOptions())
			if report.Completeness != model.CompletenessIncomplete || !hasFinding(report.Findings, tt.code) {
				t.Fatalf("report=%+v", report)
			}
			if tt.code == "claudecode-malformed-semantic" && len(report.Permissions) != 0 {
				t.Fatalf("malformed condition became modeled permission: %+v", report.Permissions)
			}
		})
	}
}

func TestResolveModelsExplicitAskEffectWithoutRuntimeCondition(t *testing.T) {
	report := ResolveWithOptions([]source.SelectedSource{{Identity: "project", Path: "/repo/.claude/settings.json", Data: []byte(`{"permissions":{"ask":["Bash(git diff:*)"]}}`)}}, supportedOptions())
	permission := permissionByMatcher(t, report.Permissions, "Bash(git diff:*)")
	if report.Completeness != model.CompletenessComplete || permission.Effect() != model.EffectAsk || permission.Capability() != "Bash" || len(permission.Conditions()) != 0 {
		t.Fatalf("report=%+v permission=%+v", report, permission)
	}
}

func TestResolvePreservesDistinctMatcherRules(t *testing.T) {
	report := ResolveWithOptions([]source.SelectedSource{{Identity: "project", Path: "/repo/.claude/settings.json", Data: []byte(`{"permissions":{"allow":["Bash(git status:*)"],"deny":["Bash(rm:*)"]}}`)}}, supportedOptions())
	status := permissionByMatcher(t, report.Permissions, "Bash(git status:*)")
	remove := permissionByMatcher(t, report.Permissions, "Bash(rm:*)")
	if report.Completeness != model.CompletenessComplete || len(report.Permissions) != 2 || status.Effect() != model.EffectAllow || remove.Effect() != model.EffectDeny {
		t.Fatalf("report=%+v status=%+v remove=%+v", report, status, remove)
	}
}

func permissionByMatcher(t testing.TB, permissions []model.Permission, matcher string) model.Permission {
	t.Helper()
	for _, permission := range permissions {
		if permission.Matcher() == matcher {
			return permission
		}
	}
	t.Fatalf("missing permission matcher %q in %d permissions", matcher, len(permissions))
	return model.Permission{}
}
func hasFinding(findings []model.Finding, code string) bool {
	return slices.ContainsFunc(findings, func(finding model.Finding) bool { return finding.Code() == code })
}
func supportedOptions() Options {
	entry := support.Entry{Target: "claudecode", Version: "local-fixture", Construct: "permission", Fixture: "claudecode-runtime", Evidence: "claudecode-doc", PermissionBearing: true, DefaultsKnown: true, MatcherKnown: true, ConditionsKnown: true}
	return Options{Version: "local-fixture", Support: support.Matrix{Entries: []support.Entry{entry}, Registry: support.Registry{Fixtures: map[string]bool{"claudecode-runtime": true}, Evidence: map[string]bool{"claudecode-doc": true}}}}
}
