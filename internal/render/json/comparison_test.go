package json_test

import (
	"bytes"
	"testing"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/compare"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/redact"
	render "github.com/jkelevra/universal-agent-policy-auditor/internal/render/json"
)

func TestMarshalComparisonGolden(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mutate func(*app.Result)
		want   string
	}{
		{
			name:   "equivalent",
			mutate: func(v *app.Result) {},
			want:   `{"schema":"auditor-comparison-report/v1alpha1","tool":{"name":"universal-agent-policy-auditor","version":"dev"},"result":"complete_no_findings","exit_code":0,"target":{"name":"opencode","requested_version":"1.18.27","support_status":"supported"},"completeness":"complete","comparison_status":"equivalent","comparison_sources":[{"side":"reference","digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},{"side":"target","digest":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}],"reasons":[],"differences":[],"findings":[],"limitations":["modeled-static-configuration"]}` + "\n",
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
			want: `{"schema":"auditor-comparison-report/v1alpha1","tool":{"name":"universal-agent-policy-auditor","version":"dev"},"result":"unsupported_or_incomplete","exit_code":2,"target":{"name":"opencode","requested_version":"1.18.27","support_status":"supported"},"completeness":"incomplete","comparison_status":"target-only","comparison_sources":[{"side":"reference","digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},{"side":"target","digest":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}],"reasons":[],"differences":[{"capability":"read","scope":"opencode","matcher":"*","only_in":"target","outcome":"target-only","reason":{"side":"target","code":"permission-only"}}],"findings":[],"limitations":["modeled-static-configuration"]}` + "\n",
		},
		{
			name: "ambiguous",
			mutate: func(v *app.Result) {
				v.Comparison = &compare.Result{Outcome: compare.Ambiguous, Reasons: []compare.Reason{{Side: compare.Reference, Code: compare.ConflictingDuplicate}}}
			},
			want: `{"schema":"auditor-comparison-report/v1alpha1","tool":{"name":"universal-agent-policy-auditor","version":"dev"},"result":"unsupported_or_incomplete","exit_code":2,"target":{"name":"opencode","requested_version":"1.18.27","support_status":"supported"},"completeness":"incomplete","comparison_status":"ambiguous","comparison_sources":[{"side":"reference","digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},{"side":"target","digest":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}],"reasons":[{"side":"reference","code":"conflicting-duplicate"}],"differences":[],"findings":[],"limitations":["modeled-static-configuration"]}` + "\n",
		},
		{
			name: "semantic-not-comparable",
			mutate: func(v *app.Result) {
				v.Comparison = &compare.Result{Outcome: compare.NotComparable, Differences: []compare.Difference{{
					Key: compare.Key{Capability: "bash", Scope: "opencode", Matcher: "*"}, Outcome: compare.NotComparable,
					Reason: compare.Reason{Side: compare.Target, Code: compare.EffectDiffers},
				}}}
			},
			want: `{"schema":"auditor-comparison-report/v1alpha1","tool":{"name":"universal-agent-policy-auditor","version":"dev"},"result":"unsupported_or_incomplete","exit_code":2,"target":{"name":"opencode","requested_version":"1.18.27","support_status":"supported"},"completeness":"incomplete","comparison_status":"not-comparable","comparison_sources":[{"side":"reference","digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},{"side":"target","digest":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}],"reasons":[],"differences":[{"capability":"bash","scope":"opencode","matcher":"*","only_in":"","outcome":"not-comparable","reason":{"side":"target","code":"effect-differs"}}],"findings":[],"limitations":["modeled-static-configuration"]}` + "\n",
		},
		{
			name: "unavailable-unsupported",
			mutate: func(v *app.Result) {
				v.SupportStatus = app.SupportUnsupported
				v.RequestedTargetVersion = ""
				v.Build.Version = "withheld"
				v.Comparison, v.ComparisonSources = nil, nil
			},
			want: `{"schema":"auditor-comparison-report/v1alpha1","tool":{"name":"universal-agent-policy-auditor","version":""},"result":"unsupported_or_incomplete","exit_code":2,"target":{"name":"opencode","requested_version":"","support_status":"unsupported"},"completeness":"incomplete","comparison_status":"unavailable","comparison_sources":[],"reasons":[],"differences":[],"findings":[{"code":"redact-version-withheld"}],"limitations":["modeled-static-configuration"]}` + "\n",
		},
		{
			name: "operational-not-comparable",
			mutate: func(v *app.Result) {
				v.Category = app.OperationalFailure
				v.ComparisonSources = nil
				v.Comparison = &compare.Result{Outcome: compare.NotComparable, Reasons: []compare.Reason{{Side: compare.Reference, Code: compare.OperandIncomplete}}}
			},
			want: `{"schema":"auditor-comparison-report/v1alpha1","tool":{"name":"universal-agent-policy-auditor","version":"dev"},"result":"operational_failure","exit_code":4,"target":{"name":"opencode","requested_version":"1.18.27","support_status":"supported"},"completeness":"incomplete","comparison_status":"not-comparable","comparison_sources":[],"reasons":[{"side":"reference","code":"operand-incomplete"}],"differences":[],"findings":[],"limitations":["modeled-static-configuration"]}` + "\n",
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
			if tt.name == "equivalent" {
				v.Category, v.Completeness = app.CompleteNoFindings, model.CompletenessComplete
			}
			tt.mutate(&v)
			report, err := redact.NewComparisonReport(v)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 100; i++ {
				got, err := render.MarshalComparison(report)
				if err != nil || string(got) != tt.want {
					t.Fatalf("render %d: %v\ngot %s\nwant %s", i, err, got, tt.want)
				}
				if bytes.Contains(got, []byte(":null")) || bytes.Count(got, []byte("\n")) != 1 || got[len(got)-1] != '\n' {
					t.Fatal("null collection or newline contract")
				}
				got[0] = '!'
			}
		})
	}
}

func TestMarshalComparisonInvalid(t *testing.T) {
	got, err := render.MarshalComparison(redact.ComparisonReport{})
	if got != nil || err != render.ErrInvalidComparisonReport || err.Error() != "json: invalid comparison report" {
		t.Fatalf("invalid report: %q, %v", got, err)
	}
}
