package support

import (
	"os"
	"testing"
	"testing/fstest"

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

func TestLoadRejectsStrictEnvelopeViolations(t *testing.T) {
	type document struct {
		name, file, schema, array string
		item                      string
	}
	docs := []document{
		{name: "matrix", file: "matrix.json", schema: "support-matrix/v1", array: "entries", item: `{"target":"opencode","unexpected":true}`},
		{name: "fixtures", file: "fixtures.json", schema: "conformance-fixtures/v1", array: "fixtures", item: `{"id":"fixture","unexpected":true}`},
		{name: "evidence", file: "evidence.json", schema: "support-evidence/v1", array: "evidence", item: `{"id":"evidence","unexpected":true}`},
	}
	for _, doc := range docs {
		valid := `{"schema":"` + doc.schema + `","` + doc.array + `":[]}`
		cases := []struct {
			name string
			body string
		}{
			{name: "malformed JSON", body: `{`},
			{name: "trailing document", body: valid + `{}`},
			{name: "unknown top-level field", body: `{"schema":"` + doc.schema + `","` + doc.array + `":[],"unexpected":true}`},
			{name: "wrong schema", body: `{"schema":"wrong","` + doc.array + `":[]}`},
			{name: "missing array", body: `{"schema":"` + doc.schema + `"}`},
			{name: "null array", body: `{"schema":"` + doc.schema + `","` + doc.array + `":null}`},
			{name: "duplicate top-level key", body: `{"schema":"` + doc.schema + `","schema":"` + doc.schema + `","` + doc.array + `":[]}`},
			{name: "duplicate nested key", body: `{"schema":"` + doc.schema + `","` + doc.array + `":[{"id":"a","id":"b"}]}`},
			{name: "case-variant key", body: `{"Schema":"` + doc.schema + `","` + doc.array + `":[]}`},
			{name: "Unicode-fold key", body: `{"\u017fchema":"` + doc.schema + `","` + doc.array + `":[]}`},
			{name: "noncanonical key", body: `{"schema-version":"` + doc.schema + `","` + doc.array + `":[]}`},
			{name: "unknown item field", body: `{"schema":"` + doc.schema + `","` + doc.array + `":[` + doc.item + `]}`},
		}
		for _, tt := range cases {
			t.Run(doc.name+"/"+tt.name, func(t *testing.T) {
				if _, err := Load(supportFS(map[string]string{doc.file: tt.body})); err == nil {
					t.Fatalf("Load accepted %s %s", doc.name, tt.name)
				}
			})
		}
	}
}

func TestStrictJSONScannerBoundsDepthDirectly(t *testing.T) {
	atLimit := []byte(`{"a":[{"b":[{"c":[{"d":[]}]}]}]}`)
	if err := checkStrictJSON(atLimit); err != nil {
		t.Fatalf("checkStrictJSON rejected depth at the limit: %v", err)
	}
	tooDeep := []byte(`{"a":[{"b":[{"c":[{"d":[{"e":[]}]}]}]}]}`)
	if err := checkStrictJSON(tooDeep); err == nil {
		t.Fatal("checkStrictJSON accepted nesting deeper than eight levels")
	}
}

func TestLoadAllowsUnicodeValuesButNotUnicodeKeys(t *testing.T) {
	matrix, err := Load(supportFS(map[string]string{
		"matrix.json":   `{"schema":"support-matrix/v1","entries":[],"gate":"valeur-unicodé"}`,
		"fixtures.json": `{"schema":"conformance-fixtures/v1","fixtures":[{"id":"fixturé"}]}`,
		"evidence.json": `{"schema":"support-evidence/v1","evidence":[{"id":"évidence"}]}`,
	}))
	if err != nil {
		t.Fatalf("Load rejected Unicode string values: %v", err)
	}
	if !matrix.Registry.Fixtures["fixturé"] || !matrix.Registry.Evidence["évidence"] {
		t.Fatalf("registry did not preserve Unicode values: %+v", matrix.Registry)
	}
}

func TestLoadAcceptsStrictGenericEnvelopeWithIDOnlyRegistries(t *testing.T) {
	matrix, err := Load(supportFS(map[string]string{
		"matrix.json":   `{"schema":"support-matrix/v1","entries":[{"target":"opencode","version":"1","construct":"permission","fixture":"fixture","evidence":"evidence","permission_bearing":true,"defaults_known":true,"matcher_known":true,"conditions_known":true}],"gate":"optional note"}`,
		"fixtures.json": `{"schema":"conformance-fixtures/v1","fixtures":[{"id":"fixture"}]}`,
		"evidence.json": `{"schema":"support-evidence/v1","evidence":[{"id":"evidence"}]}`,
	}))
	if err != nil {
		t.Fatalf("Load returned err=%v", err)
	}
	entry := matrix.Entries[0]
	if result := Validate(matrix, entry); result.Completeness != model.CompletenessComplete {
		t.Fatalf("Validate result=%+v", result)
	}
}

func TestLoadConsumesConfiguredZeroSupportRegistry(t *testing.T) {
	matrix, err := Load(os.DirFS("."))
	if err != nil || len(matrix.Entries) != 0 || len(matrix.Registry.Fixtures) != 0 || len(matrix.Registry.Evidence) != 0 {
		t.Fatalf("matrix=%+v err=%v", matrix, err)
	}
}

func supportFS(overrides map[string]string) fstest.MapFS {
	docs := map[string]string{
		"matrix.json":   `{"schema":"support-matrix/v1","entries":[]}`,
		"fixtures.json": `{"schema":"conformance-fixtures/v1","fixtures":[]}`,
		"evidence.json": `{"schema":"support-evidence/v1","evidence":[]}`,
	}
	for name, body := range overrides {
		docs[name] = body
	}
	files := fstest.MapFS{}
	for name, body := range docs {
		files[name] = &fstest.MapFile{Data: []byte(body)}
	}
	return files
}

func hasCode(findings []model.Finding, want string) bool {
	for _, finding := range findings {
		if finding.Code() == want {
			return true
		}
	}
	return false
}
