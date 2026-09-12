package redact

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
)

func TestPrivacyVersionPolicy(t *testing.T) {
	for _, version := range []string{"", "1.18.27", "1.18.28", "dev", "v1.18.27", "1.2.3-privatehost", "1.2.3", "1.18.27\n", "１.18.27"} {
		for _, tool := range []string{"dev", "v0.1.0-alpha.1", "", "1.2.3-privateuser", "v0.1.0-alpha.2"} {
			result := unsupported(model.NewFinding("unsupported-version", "private message"))
			result.SupportStatus = app.SupportUnsupported
			result.RequestedTargetVersion, result.Build.Version = version, tool
			r, err := NewReport(result)
			wantVersion, wantTool := "", ""
			if version == "1.18.27" {
				wantVersion = version
			}
			if tool == "dev" || tool == "v0.1.0-alpha.1" {
				wantTool = tool
			}
			withheld := version != wantVersion || tool != wantTool || tool == ""
			count := 1
			if withheld {
				count++
			}
			if err != nil || !r.Valid() || r.RequestedVersion() != wantVersion || r.ToolVersion() != wantTool || len(r.Findings()) != count {
				t.Fatalf("version policy mismatch: requested=%q tool=%q", version, tool)
			}
			if withheld && r.Findings()[0].Code() != "redact-version-withheld" {
				t.Fatal("missing fixed withholding code")
			}
		}
	}
	result := unsupported()
	result.Target, result.Limitations = app.TargetClaudeCode, nil
	result.SupportStatus, result.RequestedTargetVersion = app.SupportUnsupported, "1.18.27"
	r, err := NewReport(result)
	if err != nil || r.RequestedVersion() != "" || len(r.Findings()) != 1 {
		t.Fatal("OpenCode version must not pass through for another target")
	}
}

func privacyProjection(r Report) string {
	parts := []any{r.Valid(), r.Category(), r.ExitCode(), r.ToolVersion(), r.RequestedVersion(), r.Target(), r.SupportStatus(), r.Completeness(), r.SourceDigests(), r.Limitations()}
	for _, p := range r.Permissions() {
		parts = append(parts, p.Capability(), p.Effect(), p.Scope(), p.Matcher(), p.ProvenanceDigests())
	}
	for _, f := range r.Findings() {
		parts = append(parts, f.Code())
	}
	return fmt.Sprint(parts...)
}

func privatePermission(canary string) model.Permission {
	evidence := model.NewEvidence(canary, canary, canary)
	trace := model.NewResolutionTrace([]model.TraceStep{{Kind: model.TraceRule, After: model.EffectAllow, Rules: []string{canary}, Rationale: canary, Evidence: evidence}}, nil, model.CompletenessComplete)
	return model.NewPermissionWithProvenance("read", model.EffectAllow, "opencode", "*", nil, []model.Provenance{evidence}, []model.Extension{{Target: "opencode", Key: "modeled", Value: "static-configuration"}}, trace)
}

func checkPrivacy(t testing.TB, canary string, mode uint8) {
	t.Helper()
	result := success([]string{digest("a")}, perm("bash", model.EffectAsk), perm("edit", model.EffectDeny), privatePermission(canary))
	switch mode % 8 {
	case 1:
		result = unsupported(model.NewFinding("opencode-unresolved-default", canary))
		result.SupportStatus = app.SupportSupported
	case 2:
		result.Build.Version, result.RequestedTargetVersion = canary, canary
	case 3:
		result.Findings = []model.Finding{model.NewFinding(canary, canary)}
	case 4:
		result.SourceDigests = []string{canary}
	case 5:
		result.Limitations = []string{canary}
	case 6:
		result.Target = app.Target(canary)
	case 7:
		p := result.Permissions[2]
		result.Permissions[2] = model.NewPermissionWithProvenance(canary, model.Effect(canary), model.Scope(canary), canary, []model.Condition{model.Condition(canary)}, p.Provenance(), p.Extensions(), p.Trace())
	}
	r, err := NewReport(result)
	if err != nil {
		if mode%8 < 2 {
			t.Fatal("discarded metadata must not prevent a valid projection")
		}
		if err != ErrUnsafeReport || err.Error() != "redact: unsafe report" || r.Valid() || privacyProjection(r) != privacyProjection(Report{}) {
			t.Fatal("unsafe error projection")
		}
		return
	}
	before := privacyProjection(r)
	if !r.Valid() || strings.Contains(before, canary) {
		t.Fatal("private canary escaped")
	}
	result.SourceDigests = append(result.SourceDigests[:0], canary)
	for i := 0; i < 3; i++ {
		for j := range r.SourceDigests() {
			r.SourceDigests()[j] = Digest(canary)
		}
		for j := range r.Findings() {
			r.Findings()[j] = Finding{code: FindingCode(canary)}
		}
		for j := range r.Limitations() {
			r.Limitations()[j] = Limitation(canary)
		}
		for _, p := range r.Permissions() {
			p.ProvenanceDigests()[0] = Digest(canary)
			p.provenance[0] = Digest(canary)
		}
		if privacyProjection(r) != before {
			t.Fatal("accessor or input aliases report backing")
		}
	}
}

func TestPrivacyCanaries(t *testing.T) {
	for _, canary := range []string{"PRIVATE_secret=credential", "PRIVATE_user@host", "/home/PRIVATE_user/config", `C:\Users\PRIVATE_user\config`, "file:///PRIVATE_config", "https://PRIVATE_host/config?token=PRIVATE", "PRIVATE_\x00\x1b\n", "PRIVATE_\xff\xfe", "PRIVATE_раth", `PRIVATE_{"permission":{"read":"secret"}}`} {
		for mode := uint8(0); mode < 8; mode++ {
			checkPrivacy(t, canary, mode)
		}
	}
}

func FuzzReportPrivacy(f *testing.F) {
	for mode := uint8(0); mode < 8; mode++ {
		f.Add("seed\x00\xff/раth", mode)
	}
	f.Fuzz(func(t *testing.T, raw string, mode uint8) {
		if len(raw) > 4096 {
			t.Skip()
		}
		checkPrivacy(t, "PRIVATE_CANARY_"+raw, mode)
	})
}
