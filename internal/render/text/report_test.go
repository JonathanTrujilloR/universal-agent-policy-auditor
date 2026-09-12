package text_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/redact"
	rendertext "github.com/jkelevra/universal-agent-policy-auditor/internal/render/text"
)

const safeDigest = "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
const canary = "/private/path/token-canary\x1b[31m\n"

func TestRenderCompleteGolden(t *testing.T) {
	result := app.Result{
		Category: app.CompleteNoFindings, Build: app.BuildInfo{Version: "dev"},
		Target: app.TargetOpenCode, RequestedTargetVersion: "1.18.27",
		SupportStatus: app.SupportSupported, Completeness: model.CompletenessComplete,
		SourceDigests: []string{safeDigest}, Limitations: []string{"modeled-static-configuration"},
	}
	for _, capability := range []string{"read", "edit", "bash"} {
		evidence := model.NewProvenance(canary, canary, "")
		trace := model.NewResolutionTrace([]model.TraceStep{{Kind: model.TraceRule, After: model.EffectAsk, Evidence: evidence}}, nil, model.CompletenessComplete)
		result.Permissions = append(result.Permissions, model.NewPermissionWithProvenance(
			capability, model.EffectAsk, "opencode", "*", nil, []model.Provenance{evidence},
			[]model.Extension{{Target: "opencode", Key: "modeled", Value: "static-configuration"}}, trace))
	}
	want := `Tool: universal-agent-policy-auditor dev
Result: complete_no_findings (exit 0)
Target: opencode
Requested version: 1.18.27
Support status: supported
Completeness: complete
Source digests:
- ` + safeDigest + `
Permissions:
- capability=bash; effect=ask; scope=opencode; matcher=*; provenance=[` + safeDigest + `]
- capability=edit; effect=ask; scope=opencode; matcher=*; provenance=[` + safeDigest + `]
- capability=read; effect=ask; scope=opencode; matcher=*; provenance=[` + safeDigest + `]
Limitations:
- modeled-static-configuration
`
	assertGolden(t, result, want)
}

func assertGolden(t *testing.T, result app.Result, want string) {
	t.Helper()
	report, err := redact.NewReport(result)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		data, err := rendertext.Render(report)
		if err != nil || string(data) != want {
			t.Fatalf("iteration %d: data=%q err=%v; want=%q", i, data, err, want)
		}
		output := string(data)
		if !utf8.Valid(data) || !strings.HasSuffix(output, "\n") || strings.HasSuffix(output, "\n\n") {
			t.Fatalf("invalid UTF-8 or final newline: %q", data)
		}
		for _, line := range strings.Split(output, "\n") {
			if strings.TrimRight(line, " \t\r") != line {
				t.Fatalf("trailing whitespace: %q", line)
			}
		}
		for _, forbidden := range []string{"token-canary", "/private/path", "\x1b"} {
			if strings.Contains(output, forbidden) {
				t.Fatalf("unsafe output: %q", data)
			}
		}
		data[0] = '!' // Each render owns its bytes, independent of earlier callers.
	}
}

func TestRenderUnsupportedGolden(t *testing.T) {
	for _, tc := range []struct {
		name, code, findings string
		status               app.SupportStatus
	}{
		{"unsupported", "unsupported-version", "- redact-version-withheld\n- unsupported-version\n", app.SupportUnsupported},
		{"incomplete", "opencode-unresolved-default", "- opencode-unresolved-default\n- redact-version-withheld\n", app.SupportSupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := app.Result{
				Category: app.UnsupportedOrIncomplete, Build: app.BuildInfo{Version: canary},
				Target: app.TargetOpenCode, RequestedTargetVersion: canary,
				SupportStatus: tc.status, Completeness: model.CompletenessIncomplete,
				Findings:    []model.Finding{model.NewFinding(tc.code, canary)},
				Limitations: []string{"modeled-static-configuration"},
			}
			want := "Tool: universal-agent-policy-auditor\nResult: unsupported_or_incomplete (exit 2)\nFindings:\n" + tc.findings +
				"Target: opencode\nSupport status: " + string(tc.status) + "\nCompleteness: incomplete\nLimitations:\n- modeled-static-configuration\n"
			assertGolden(t, result, want)
		})
	}
}

func TestRenderEmptySectionsAndExits(t *testing.T) {
	for _, tc := range []struct {
		name     string
		category app.Category
		exit     int
		target   app.Target
		status   app.SupportStatus
	}{
		{"invalid", app.InvalidRequest, 3, "", app.SupportUnresolved},
		{"operational", app.OperationalFailure, 4, app.TargetClaudeCode, app.SupportUnresolved},
		{"claude", app.UnsupportedOrIncomplete, 2, app.TargetClaudeCode, app.SupportUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := app.Result{Category: tc.category, Build: app.BuildInfo{Version: "v0.1.0-alpha.1"},
				Target: tc.target, SupportStatus: tc.status, Completeness: model.CompletenessIncomplete}
			target := "Target:"
			if tc.target != "" {
				target += " " + string(tc.target)
			}
			want := fmt.Sprintf("Tool: universal-agent-policy-auditor v0.1.0-alpha.1\nResult: %s (exit %d)\n%s\nSupport status: %s\nCompleteness: incomplete\n", tc.category, tc.exit, target, tc.status)
			assertGolden(t, result, want)
		})
	}
	report, err := redact.NewReport(app.Result{Category: app.CompleteWithFindings})
	if err == nil {
		t.Fatal("reserved exit 1 category must remain unconstructable")
	}
	if data, err := rendertext.Render(report); data != nil || err != rendertext.ErrInvalidReport {
		t.Fatalf("rejected conversion: data=%q err=%v", data, err)
	}
}

func TestRenderIncompleteSourceDigestsGolden(t *testing.T) {
	other := "sha256:" + strings.Repeat("f", 64)
	result := app.Result{
		Category: app.UnsupportedOrIncomplete, Build: app.BuildInfo{Version: "dev"},
		Target: app.TargetOpenCode, SupportStatus: app.SupportSupported,
		Completeness:  model.CompletenessIncomplete,
		SourceDigests: []string{other, safeDigest, other},
		Findings:      []model.Finding{model.NewFinding("source-unreadable", canary)},
		Limitations:   []string{"modeled-static-configuration"},
	}
	want := "Tool: universal-agent-policy-auditor dev\nResult: unsupported_or_incomplete (exit 2)\nFindings:\n- source-unreadable\n" +
		"Target: opencode\nSupport status: supported\nCompleteness: incomplete\nSource digests:\n- " + safeDigest + "\n- " + other +
		"\nLimitations:\n- modeled-static-configuration\n"
	assertGolden(t, result, want)
}

func TestRenderRejectsZeroReport(t *testing.T) {
	data, err := rendertext.Render(redact.Report{})
	if data != nil || err != rendertext.ErrInvalidReport || err.Error() != "text: invalid report" {
		t.Fatalf("data=%q err=%v", data, err)
	}
}
