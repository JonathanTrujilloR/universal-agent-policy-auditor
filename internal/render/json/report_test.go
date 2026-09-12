package json_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/redact"
	renderjson "github.com/jkelevra/universal-agent-policy-auditor/internal/render/json"
)

const safeDigest = "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
const canary = "/private/path/token-canary<>&\"\n"

func TestMarshalSuccessGoldenAndRepeat(t *testing.T) {
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
	want := `{"schema":"auditor-report/v1alpha1","tool":{"name":"universal-agent-policy-auditor","version":"dev"},"result":"complete_no_findings","exit_code":0,"target":{"name":"opencode","requested_version":"1.18.27","support_status":"supported"},"completeness":"complete","source_digests":["` + safeDigest + `"],"permissions":[` +
		`{"capability":"bash","effect":"ask","scope":"opencode","matcher":"*","provenance_digests":["` + safeDigest + `"]},` +
		`{"capability":"edit","effect":"ask","scope":"opencode","matcher":"*","provenance_digests":["` + safeDigest + `"]},` +
		`{"capability":"read","effect":"ask","scope":"opencode","matcher":"*","provenance_digests":["` + safeDigest + `"]}],"findings":[],"limitations":["modeled-static-configuration"]}` + "\n"
	assertGolden(t, result, want)
}

func TestMarshalUnsupportedAndIncompleteGolden(t *testing.T) {
	for _, status := range []app.SupportStatus{app.SupportUnsupported, app.SupportSupported} {
		code := "unsupported-version"
		if status == app.SupportSupported {
			code = "opencode-unresolved-default"
		}
		result := app.Result{
			Category: app.UnsupportedOrIncomplete, Build: app.BuildInfo{Version: canary},
			Target: app.TargetOpenCode, RequestedTargetVersion: canary, SupportStatus: status,
			Completeness: model.CompletenessIncomplete, Limitations: []string{"modeled-static-configuration"},
			Findings: []model.Finding{model.NewFinding(code, canary)},
		}
		codes := `{"code":"redact-version-withheld"},{"code":"unsupported-version"}`
		if status == app.SupportSupported {
			codes = `{"code":"opencode-unresolved-default"},{"code":"redact-version-withheld"}`
		}
		want := fmt.Sprintf(`{"schema":"auditor-report/v1alpha1","tool":{"name":"universal-agent-policy-auditor","version":""},"result":"unsupported_or_incomplete","exit_code":2,"target":{"name":"opencode","requested_version":"","support_status":"%s"},"completeness":"incomplete","source_digests":[],"permissions":[],"findings":[%s],"limitations":["modeled-static-configuration"]}`+"\n", status, codes)
		assertGolden(t, result, want)
	}
}

func TestMarshalEmptyCollectionsAndExits(t *testing.T) {
	for _, tc := range []struct {
		category app.Category
		exit     int
	}{{app.InvalidRequest, 3}, {app.OperationalFailure, 4}, {app.UnsupportedOrIncomplete, 2}} {
		category := tc.category
		status := app.SupportUnresolved
		if category == app.UnsupportedOrIncomplete {
			status = app.SupportUnsupported
		}
		result := app.Result{
			Category: category, Build: app.BuildInfo{Version: "dev"}, Target: app.TargetClaudeCode,
			SupportStatus: status, Completeness: model.CompletenessIncomplete,
		}
		want := fmt.Sprintf(`{"schema":"auditor-report/v1alpha1","tool":{"name":"universal-agent-policy-auditor","version":"dev"},"result":"%s","exit_code":%d,"target":{"name":"claudecode","requested_version":"","support_status":"%s"},"completeness":"incomplete","source_digests":[],"permissions":[],"findings":[],"limitations":[]}`+"\n", category, tc.exit, status)
		assertGolden(t, result, want)
	}
	report, err := redact.NewReport(app.Result{Category: app.CompleteWithFindings})
	if err == nil {
		t.Fatal("reserved category must remain unconstructable")
	}
	if data, err := renderjson.Marshal(report); data != nil || err != renderjson.ErrInvalidReport {
		t.Fatalf("rejected conversion: data=%q err=%v", data, err)
	}
}

func assertGolden(t *testing.T, result app.Result, want string) {
	t.Helper()
	report, err := redact.NewReport(result)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		data, err := renderjson.Marshal(report)
		if err != nil || string(data) != want {
			t.Fatalf("iteration %d: data=%s err=%v; want=%s", i, data, err, want)
		}
		if !utf8.Valid(data) || !json.Valid(data) || bytes.Count(data, []byte{'\n'}) != 1 || strings.Contains(string(data), "token-canary") || strings.Contains(string(data), "/private/path") {
			t.Fatalf("unsafe or invalid JSON: %q", data)
		}
		data[0] = '!' // Returned bytes cannot affect the next render.
	}
}

// Closed report strings cannot contain HTML/control characters. Pin the standard
// encoder policy separately without adding an injectable renderer DTO.
func TestStandardJSONEscapingPolicy(t *testing.T) {
	data, err := json.Marshal("<>&\"\n\u2028\u2029")
	if err != nil || string(data) != `"\u003c\u003e\u0026\"\n\u2028\u2029"` {
		t.Fatalf("data=%s err=%v", data, err)
	}
}

func TestMarshalRejectsZeroReport(t *testing.T) {
	data, err := renderjson.Marshal(redact.Report{})
	if data != nil || !errors.Is(err, renderjson.ErrInvalidReport) || err.Error() != "json: invalid report" {
		t.Fatalf("data=%q err=%v", data, err)
	}
}
