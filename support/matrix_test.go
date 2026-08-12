package support

import (
	"os"
	"testing"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
)

func TestValidateFailsClosedForUnprovenPermissionMetadata(t *testing.T) {
	tests := []struct {
		name   string
		matrix Matrix
		entry  Entry
		code   string
	}{
		{"unsupported version", Matrix{}, Entry{Target: "opencode", Version: "future", Construct: "permission", PermissionBearing: true}, "unsupported-version"},
		{"malformed entry", Matrix{}, Entry{Target: "opencode", Version: "x", PermissionBearing: true}, "malformed-entry"},
		{"unknown construct", Matrix{}, Entry{Target: "claude-code", Version: "x", Construct: "mystery", PermissionBearing: true, Fixture: "f", Evidence: "e"}, "unknown-construct"},
		{"uncertain semantics", Matrix{Entries: []Entry{{Target: "claude-code", Version: "x", Construct: "permission", PermissionBearing: true, Fixture: "f", Evidence: "e"}}}, Entry{Target: "claude-code", Version: "x", Construct: "permission", PermissionBearing: true, Fixture: "f", Evidence: "e"}, "incomplete-evidence"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.matrix, tt.entry)
			if result.Completeness != model.CompletenessIncomplete || !hasCode(result.Findings, tt.code) {
				t.Fatalf("result=%+v", result)
			}
		})
	}
}

func TestValidateRequiresFixtureEvidenceAndKnownSemantics(t *testing.T) {
	entry := Entry{Target: "opencode", Version: "1", Construct: "permission", PermissionBearing: true, Fixture: "default.json", Evidence: "official", DefaultsKnown: true, MatcherKnown: true, ConditionsKnown: true}
	result := Validate(Matrix{Entries: []Entry{entry}, Registry: Registry{Fixtures: map[string]bool{"default.json": true}, Evidence: map[string]bool{"official": true}}}, entry)
	if result.Completeness != model.CompletenessComplete || len(result.Findings) != 0 {
		t.Fatalf("result=%+v", result)
	}
}

func TestValidateUsesCanonicalRowAndRegistry(t *testing.T) {
	complete := Entry{Target: "opencode", Version: "1", Construct: "permission", PermissionBearing: true, Fixture: "fixture", Evidence: "evidence", DefaultsKnown: true, MatcherKnown: true, ConditionsKnown: true}
	forged := complete
	forged.Fixture, forged.Evidence, forged.DefaultsKnown, forged.MatcherKnown, forged.ConditionsKnown = "forged", "forged", true, true, true
	incomplete := complete
	incomplete.Fixture, incomplete.Evidence, incomplete.DefaultsKnown = "", "", false
	for _, tt := range []struct {
		name, code string
		matrix     Matrix
		entry      Entry
	}{
		{name: "forged candidate cannot strengthen row", matrix: Matrix{Entries: []Entry{incomplete}}, entry: forged, code: "incomplete-evidence"},
		{name: "missing fixture registry id", matrix: Matrix{Entries: []Entry{complete}, Registry: Registry{Evidence: map[string]bool{"evidence": true}}}, entry: complete, code: "missing-fixture"},
		{name: "missing evidence registry id", matrix: Matrix{Entries: []Entry{complete}, Registry: Registry{Fixtures: map[string]bool{"fixture": true}}}, entry: complete, code: "missing-evidence"},
		{name: "incomplete canonical row", matrix: Matrix{Entries: []Entry{incomplete}}, entry: incomplete, code: "incomplete-evidence"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.matrix, tt.entry)
			if result.Completeness != model.CompletenessIncomplete || !hasCode(result.Findings, tt.code) {
				t.Fatalf("result=%+v", result)
			}
		})
	}
}

func TestLoadConsumesConfiguredZeroSupportRegistry(t *testing.T) {
	matrix, err := Load(os.DirFS("."))
	if err != nil || len(matrix.Entries) != 0 || len(matrix.Registry.Fixtures) != 0 || len(matrix.Registry.Evidence) != 0 {
		t.Fatalf("matrix=%+v err=%v", matrix, err)
	}
}

func hasCode(findings []model.Finding, want string) bool {
	for _, finding := range findings {
		if finding.Code() == want {
			return true
		}
	}
	return false
}
