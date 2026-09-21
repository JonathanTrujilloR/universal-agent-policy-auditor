package text_test

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/compare"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/redact"
	render "github.com/jkelevra/universal-agent-policy-auditor/internal/render/text"
)

func TestRenderComparisonGolden(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mutate func(*app.Result)
		want   string
	}{
		{
			name: "equivalent",
			mutate: func(v *app.Result) {
				v.Category, v.Completeness = app.CompleteNoFindings, model.CompletenessComplete
			},
			want: `Tool: universal-agent-policy-auditor dev
Result: complete_no_findings (exit 0)
Comparison status: equivalent
Target: opencode
Requested version: 1.18.27
Support status: supported
Completeness: complete
Comparison sources:
- side=reference; digest=sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
- side=target; digest=sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
Limitations:
- modeled-static-configuration
`,
		},
		{
			name: "target-only",
			mutate: func(v *app.Result) {
				v.Comparison = &compare.Result{Outcome: compare.TargetOnly, Differences: []compare.Difference{{
					Key:    compare.Key{Capability: "read", Scope: "opencode", Matcher: "*"},
					OnlyIn: compare.Target, Outcome: compare.TargetOnly,
					Reason: compare.Reason{Side: compare.Target, Code: compare.PermissionOnly},
				}}}
			},
			want: `Tool: universal-agent-policy-auditor dev
Result: unsupported_or_incomplete (exit 2)
Comparison status: target-only
Differences:
- capability=read; scope=opencode; matcher=*; only_in=target; outcome=target-only; reason_side=target; reason_code=permission-only
Target: opencode
Requested version: 1.18.27
Support status: supported
Completeness: incomplete
Comparison sources:
- side=reference; digest=sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
- side=target; digest=sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
Limitations:
- modeled-static-configuration
`,
		},
		{
			name: "ambiguous",
			mutate: func(v *app.Result) {
				v.Comparison = &compare.Result{Outcome: compare.Ambiguous, Reasons: []compare.Reason{{Side: compare.Reference, Code: compare.ConflictingDuplicate}}}
			},
			want: `Tool: universal-agent-policy-auditor dev
Result: unsupported_or_incomplete (exit 2)
Comparison status: ambiguous
Reasons:
- side=reference; code=conflicting-duplicate
Target: opencode
Requested version: 1.18.27
Support status: supported
Completeness: incomplete
Comparison sources:
- side=reference; digest=sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
- side=target; digest=sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
Limitations:
- modeled-static-configuration
`,
		},
		{
			name: "semantic-not-comparable",
			mutate: func(v *app.Result) {
				v.Comparison = &compare.Result{Outcome: compare.NotComparable, Differences: []compare.Difference{{
					Key: compare.Key{Capability: "bash", Scope: "opencode", Matcher: "*"}, Outcome: compare.NotComparable,
					Reason: compare.Reason{Side: compare.Target, Code: compare.EffectDiffers},
				}}}
			},
			want: `Tool: universal-agent-policy-auditor dev
Result: unsupported_or_incomplete (exit 2)
Comparison status: not-comparable
Differences:
- capability=bash; scope=opencode; matcher=*; only_in=; outcome=not-comparable; reason_side=target; reason_code=effect-differs
Target: opencode
Requested version: 1.18.27
Support status: supported
Completeness: incomplete
Comparison sources:
- side=reference; digest=sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
- side=target; digest=sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
Limitations:
- modeled-static-configuration
`,
		},
		{
			name: "unavailable-unsupported",
			mutate: func(v *app.Result) {
				v.SupportStatus = app.SupportUnsupported
				v.RequestedTargetVersion = ""
				v.Build.Version = "withheld"
				v.Comparison, v.ComparisonSources = nil, nil
			},
			want: `Tool: universal-agent-policy-auditor
Result: unsupported_or_incomplete (exit 2)
Comparison status: unavailable
Findings:
- redact-version-withheld
Target: opencode
Support status: unsupported
Completeness: incomplete
Limitations:
- modeled-static-configuration
`,
		},
		{
			name: "operational-not-comparable",
			mutate: func(v *app.Result) {
				v.Category = app.OperationalFailure
				v.ComparisonSources = nil
				v.Comparison = &compare.Result{Outcome: compare.NotComparable, Reasons: []compare.Reason{{Side: compare.Reference, Code: compare.OperandIncomplete}}}
			},
			want: `Tool: universal-agent-policy-auditor dev
Result: operational_failure (exit 4)
Comparison status: not-comparable
Reasons:
- side=reference; code=operand-incomplete
Target: opencode
Requested version: 1.18.27
Support status: supported
Completeness: incomplete
Limitations:
- modeled-static-configuration
`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			v := app.Result{
				Mode: app.ModeCompare, Category: app.UnsupportedOrIncomplete,
				Build: app.BuildInfo{Version: "dev"}, Target: app.TargetOpenCode,
				RequestedTargetVersion: "1.18.27", SupportStatus: app.SupportSupported,
				Completeness: model.CompletenessIncomplete, Limitations: []string{"modeled-static-configuration"},
				Comparison: &compare.Result{Outcome: compare.Equivalent},
				ComparisonSources: []app.ComparisonSource{
					{Side: compare.Target, Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
					{Side: compare.Reference, Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
				},
			}
			tt.mutate(&v)
			report, err := redact.NewComparisonReport(v)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 100; i++ {
				got, err := render.RenderComparison(report)
				if err != nil || string(got) != tt.want {
					t.Fatalf("render %d: %v\ngot %s\nwant %s", i, err, got, tt.want)
				}
				if !utf8.Valid(got) || !bytes.HasSuffix(got, []byte("\n")) || bytes.HasSuffix(got, []byte("\n\n")) {
					t.Fatal("UTF-8 or final newline contract")
				}
				for _, forbidden := range []string{"\x1b", "CANARY", "/private/", "/home/", "\r"} {
					if bytes.Contains(got, []byte(forbidden)) {
						t.Fatalf("forbidden output %q", forbidden)
					}
				}
				for _, line := range strings.Split(strings.TrimSuffix(string(got), "\n"), "\n") {
					if line == "" || strings.TrimRight(line, " \t") != line {
						t.Fatal("blank line or trailing whitespace")
					}
				}
				got[0] = '!'
			}
		})
	}
}

func TestRenderComparisonInvalid(t *testing.T) {
	got, err := render.RenderComparison(redact.ComparisonReport{})
	if got != nil || err != render.ErrInvalidComparisonReport || err.Error() != "text: invalid comparison report" {
		t.Fatalf("invalid report: %q, %v", got, err)
	}
}
