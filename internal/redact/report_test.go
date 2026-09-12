package redact

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
)

func TestNewReportAcceptsExactAppRunSuccess(t *testing.T) {
	report, err := NewReport(success([]string{digest("a")}, perms()...))
	if err != nil || !report.Valid() || report.Category() != CategoryCompleteNoFindings || report.ExitCode() != 0 || report.Target() != TargetOpenCode || report.RequestedVersion() != "1.18.27" || (Report{}).Valid() {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}

func TestNewReportAcceptsUnsupportedOpenCodeVersion(t *testing.T) {
	result := unsupported(model.NewFinding("unsupported-version", "raw private context"))
	result.SupportStatus = app.SupportUnsupported
	report, err := NewReport(result)
	if err != nil || report.SupportStatus() != SupportUnsupported || len(report.Findings()) != 1 || report.Findings()[0].Code() != "unsupported-version" {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}

func TestNewReportRejectsUnsafeFactsWithGenericError(t *testing.T) {
	canary := "private-token-canary"
	digest := "sha256:" + strings.Repeat("a", 64)
	cases := []app.Result{{}, unsupported(model.NewFinding("unknown-"+canary, "message "+canary)), success([]string{"SHA256:" + strings.Repeat("a", 64)}, perms()...), success([]string{digest}, perm("write", model.EffectAllow)), success([]string{digest}, perm("read", model.EffectAllow), perm("read", model.EffectDeny), perm("edit", model.EffectDeny), perm("bash", model.EffectAsk)), {Category: app.InvalidRequest, Build: app.BuildInfo{Version: "dev"}, Target: app.TargetOpenCode, SupportStatus: app.SupportUnresolved, Completeness: model.CompletenessIncomplete, Permissions: []model.Permission{perm("read", model.EffectAllow)}}, {Category: app.CompleteWithFindings, Build: app.BuildInfo{Version: "dev"}, Target: app.TargetOpenCode, SupportStatus: app.SupportSupported, Completeness: model.CompletenessComplete}}
	for i, result := range cases {
		_, err := NewReport(result)
		if !errors.Is(err, ErrUnsafeReport) || err.Error() != "redact: unsafe report" || strings.Contains(err.Error(), canary) {
			t.Fatalf("case %d error=%v", i, err)
		}
	}
}

func TestReportAccessorsSortDedupAndDefendCopies(t *testing.T) {
	result := success([]string{digest("b"), digest("a"), digest("a")}, append(perms(), perms()...)...)
	report, err := NewReport(result)
	if err != nil {
		t.Fatal(err)
	}
	digests, permissions := report.SourceDigests(), report.Permissions()
	digests[0], permissions[0].provenance[0] = Digest(digest("f")), Digest(digest("f"))
	if got := fmt.Sprint(report.SourceDigests()); got != "["+digest("a")+" "+digest("b")+"]" {
		t.Fatalf("digests=%s", got)
	}
	if got := strings.Join(permissionKeys(report.Permissions()), ","); got != "bash:ask,edit:deny,read:allow" || len(report.Permissions()) != 3 || report.Permissions()[0].ProvenanceDigests()[0] != Digest(result.Permissions[0].Provenance()[0].Value()) {
		t.Fatalf("permissions=%s", got)
	}
}

func TestNewReportAcceptsSupportedIncompleteOpenCodeOnlyWithFindings(t *testing.T) {
	result := unsupported(model.NewFinding("opencode-unresolved-default", "raw private context"), model.NewFinding("opencode-unresolved-default", "duplicate"))
	result.Build.Version, result.RequestedTargetVersion = "dev/private", "1.18.27/private"
	result.SupportStatus = app.SupportSupported
	result.SourceDigests = []string{digest("a")}
	report, err := NewReport(result)
	findings := report.Findings()
	if err != nil || report.ToolVersion() != "" || report.RequestedVersion() != "" || len(findings) != 2 || findings[0].Code() != "opencode-unresolved-default" || findings[1].Code() != "redact-version-withheld" {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}

func TestNewReportRejectsClosedStateViolations(t *testing.T) {
	cases := []app.Result{
		{Category: app.UnsupportedOrIncomplete, Build: app.BuildInfo{Version: "dev"}, Target: app.TargetClaudeCode, SupportStatus: app.SupportSupported, Completeness: model.CompletenessIncomplete},
		{Category: app.InvalidRequest, Build: app.BuildInfo{Version: "dev"}, Target: app.TargetOpenCode, SupportStatus: app.SupportSupported, Completeness: model.CompletenessIncomplete, Limitations: []string{"modeled-static-configuration"}},
		unsupported(),
		{Category: app.UnsupportedOrIncomplete, Build: app.BuildInfo{Version: "dev"}, Target: app.TargetClaudeCode, SupportStatus: app.SupportUnsupported, Completeness: model.CompletenessIncomplete, Limitations: []string{"modeled-static-configuration"}},
		{Category: app.OperationalFailure, Build: app.BuildInfo{Version: "dev"}, SupportStatus: app.SupportUnresolved, Completeness: model.CompletenessIncomplete},
	}
	p := perm("read", model.EffectAllow)
	unrelated := model.NewPermissionWithProvenance(p.Capability(), p.Effect(), p.Scope(), p.Matcher(), nil, []model.Provenance{model.NewProvenance("opencode", "unrelated", "allow")}, p.Extensions(), p.Trace())
	cases = append(cases, success([]string{digest("a")}, perm("bash", model.EffectAsk), perm("edit", model.EffectDeny), unrelated))
	cases[2].Permissions = []model.Permission{p}
	for i, result := range cases {
		_, err := NewReport(result)
		if !errors.Is(err, ErrUnsafeReport) {
			t.Fatalf("case %d error=%v", i, err)
		}
	}
}
func unsupported(findings ...model.Finding) app.Result {
	return app.Result{Category: app.UnsupportedOrIncomplete, Build: app.BuildInfo{Version: "dev"}, Target: app.TargetOpenCode, SupportStatus: app.SupportUnresolved, Completeness: model.CompletenessIncomplete, Findings: findings, Limitations: []string{"modeled-static-configuration"}}
}
func digest(fill string) string { return "sha256:" + strings.Repeat(fill, 64) }
func success(digests []string, permissions ...model.Permission) app.Result {
	return app.Result{Category: app.CompleteNoFindings, Build: app.BuildInfo{Version: "dev"}, Target: app.TargetOpenCode, RequestedTargetVersion: "1.18.27", SupportStatus: app.SupportSupported, SourceDigests: digests, Completeness: model.CompletenessComplete, Permissions: permissions, Limitations: []string{"modeled-static-configuration"}}
}
func perms() []model.Permission {
	return []model.Permission{perm("bash", model.EffectAsk), perm("edit", model.EffectDeny), perm("read", model.EffectAllow)}
}
func perm(capability string, effect model.Effect) model.Permission {
	provenance := model.NewProvenance("opencode", "/private/path#permission."+capability, string(effect))
	trace := model.NewResolutionTrace([]model.TraceStep{{Kind: model.TraceRule, After: effect, Evidence: provenance}}, nil, model.CompletenessComplete)
	return model.NewPermissionWithProvenance(capability, effect, model.Scope("opencode"), "*", nil, []model.Provenance{provenance}, []model.Extension{{Target: "opencode", Key: "modeled", Value: "static-configuration"}}, trace)
}
func permissionKeys(values []Permission) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = string(value.Capability()) + ":" + string(value.Effect())
	}
	return out
}
