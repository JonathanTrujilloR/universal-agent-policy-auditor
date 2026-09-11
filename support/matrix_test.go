package support

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	matrix, err := Load(supportFS(map[string]string{"matrix.json": `{"schema":"support-matrix/v1","entries":[],"gate":"valeur-unicodé"}`}))
	if err != nil || len(matrix.Entries) != 0 {
		t.Fatalf("Load rejected Unicode string value: matrix=%+v err=%v", matrix, err)
	}
}

func TestLoadAcceptsStrictGenericMetadata(t *testing.T) {
	matrix, err := Load(metadataFS(nil))
	if err != nil {
		t.Fatalf("Load returned err=%v", err)
	}
	entry := matrix.Entries[0]
	if result := Validate(matrix, entry); result.Completeness != model.CompletenessComplete {
		t.Fatalf("Validate result=%+v", result)
	}
	if !matrix.Registry.Fixtures["fixture"] || !matrix.Registry.Evidence["evidence"] {
		t.Fatalf("registry did not include validated records: %+v", matrix.Registry)
	}
}

func TestLoadRejectsMetadataIntegrityFailures(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mutate func(*metadataDocs)
	}{
		{"matrix missing target", func(d *metadataDocs) { replaceOnce(&d.matrix, `"target":"target",`, ``) }},
		{"fixture surrounding whitespace id", func(d *metadataDocs) { replaceOnce(&d.fixtures, `"id":"fixture"`, `"id":" fixture"`) }},
		{"evidence whitespace claim", func(d *metadataDocs) { replaceOnce(&d.evidence, `"claim:defaults"`, `" claim:defaults"`) }},
		{"source surrounding whitespace path", func(d *metadataDocs) { replaceOnce(&d.evidence, `"docs/source.md"`, `"docs/source.md "`) }},
		{"uppercase commit", func(d *metadataDocs) { replaceOnce(&d.evidence, strings.Repeat("b", 40), strings.Repeat("B", 40)) }},
		{"short commit", func(d *metadataDocs) { replaceOnce(&d.evidence, strings.Repeat("b", 40), strings.Repeat("b", 39)) }},
		{"uppercase fixture sha", func(d *metadataDocs) { replaceOnce(&d.fixtures, fixtureHash(), strings.ToUpper(fixtureHash())) }},
		{"short source sha", func(d *metadataDocs) { replaceOnce(&d.evidence, strings.Repeat("a", 64), strings.Repeat("a", 63)) }},
		{"dot fixture path", func(d *metadataDocs) { replaceOnce(&d.fixtures, "fixtures/generic.txt", ".") }},
		{"escaping fixture path", func(d *metadataDocs) { replaceOnce(&d.fixtures, "fixtures/generic.txt", "../generic.txt") }},
		{"unreadable fixture path", func(d *metadataDocs) { replaceOnce(&d.fixtures, "fixtures/generic.txt", "fixtures/missing.txt") }},
		{"mismatched fixture sha", func(d *metadataDocs) { replaceOnce(&d.fixtures, fixtureHash(), strings.Repeat("c", 64)) }},
		{"empty claims", func(d *metadataDocs) {
			replaceOnce(&d.evidence, `"claims":["claim:defaults","claim:matcher"]`, `"claims":[]`)
		}},
		{"duplicate claims", func(d *metadataDocs) { replaceOnce(&d.evidence, `"claim:matcher"`, `"claim:defaults"`) }},
		{"empty sources", func(d *metadataDocs) {
			replaceOnce(&d.evidence, `"sources":[{"path":"docs/source.md","sha256":"`+strings.Repeat("a", 64)+`"}]`, `"sources":[]`)
		}},
		{"duplicate source paths", func(d *metadataDocs) {
			replaceOnce(&d.evidence, `}]}`, `},{"path":"docs/source.md","sha256":"`+strings.Repeat("a", 64)+`"}]}`)
		}},
		{"invalid source path", func(d *metadataDocs) { replaceOnce(&d.evidence, "docs/source.md", "docs//source.md") }},
		{"dot source path", func(d *metadataDocs) { replaceOnce(&d.evidence, "docs/source.md", ".") }},
		{"duplicate fixture IDs", func(d *metadataDocs) {
			replaceOnce(&d.fixtures, `]}`, `,{"id":"fixture","target":"target","version":"1","construct":"permission","path":"fixtures/generic.txt","sha256":"`+fixtureHash()+`"}]}`)
		}},
		{"duplicate evidence IDs", func(d *metadataDocs) {
			replaceOnce(&d.evidence, `]}`, `,{"id":"evidence","target":"target","version":"1","construct":"permission","fixture":"fixture","authority":{"repository":"https://github.com/owner/repo","tag":"v1.18.27","commit":"`+strings.Repeat("b", 40)+`"},"claims":["claim:other"],"sources":[{"path":"docs/other.md","sha256":"`+strings.Repeat("a", 64)+`"}]}]}`)
		}},
		{"duplicate matrix triples", func(d *metadataDocs) {
			replaceOnce(&d.matrix, `]`, `,{"target":"target","version":"1","construct":"permission","fixture":"fixture","evidence":"evidence","permission_bearing":true,"defaults_known":true,"matcher_known":true,"conditions_known":true}]`)
		}},
		{"orphan evidence fixture", func(d *metadataDocs) { replaceOnce(&d.evidence, `"fixture":"fixture"`, `"fixture":"missing"`) }},
		{"orphan evidence tuple", func(d *metadataDocs) {
			d.matrix = `{"schema":"support-matrix/v1","entries":[]}`
			replaceOnce(&d.evidence, `"target":"target"`, `"target":"other"`)
		}},
		{"matrix missing fixture", func(d *metadataDocs) { replaceOnce(&d.matrix, `"fixture":"fixture"`, `"fixture":"missing"`) }},
		{"matrix missing evidence", func(d *metadataDocs) { replaceOnce(&d.matrix, `"evidence":"evidence"`, `"evidence":"missing"`) }},
		{"matrix fixture tuple mismatch", func(d *metadataDocs) { replaceOnce(&d.fixtures, `"construct":"permission"`, `"construct":"other"`) }},
		{"matrix evidence tuple mismatch", func(d *metadataDocs) { replaceOnce(&d.evidence, `"version":"1"`, `"version":"2"`) }},
		{"matrix evidence fixture mismatch", func(d *metadataDocs) {
			replaceOnce(&d.evidence, `"fixture":"fixture"`, `"fixture":"fixture-two"`)
			replaceOnce(&d.fixtures, `]}`, `,{"id":"fixture-two","target":"target","version":"1","construct":"permission","path":"fixtures/generic.txt","sha256":"`+fixtureHash()+`"}]}`)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) { expectLoadError(t, tt.mutate) })
	}
	for _, repo := range []string{"http://github.com/owner/.github", "https://user@github.com/owner/.github", "https://github.com:443/owner/.github", "https://github.com/owner/.github?x=1", "https://github.com/owner/.github#frag", "https://github.com/owner/.github/extra", "https://github.com/owner/.github/", "https://github.com/owner/./repo", "https://github.com/ow ner/repo", "https://github.com/owner/re%70o", "https://github.com/a--b/repo"} {
		t.Run("repo/"+repo, func(t *testing.T) {
			expectLoadError(t, func(d *metadataDocs) { replaceOnce(&d.evidence, "https://github.com/owner/.github", repo) })
		})
	}
	for _, tag := range []string{".", "..", "bad tag", "bad\n", "/v1", "v1/", ".v1", "v1.", "v//1", "v@{1", "v1.lock", "v~1", "v^1", "v:1", "v?1", "v*1", "v[1", `v\\1`} {
		t.Run("tag/"+tag, func(t *testing.T) {
			expectLoadError(t, func(d *metadataDocs) { replaceOnce(&d.evidence, `"tag":"v1.18.27"`, fmt.Sprintf(`"tag":%q`, tag)) })
		})
	}
}

func TestLoadActivatesOnlyCheckedInOpenCodeScalarPermissionRow(t *testing.T) {
	matrix, err := Load(os.DirFS("."))
	if err != nil {
		t.Fatalf("Load returned err=%v", err)
	}
	if len(matrix.Entries) != 1 {
		t.Fatalf("matrix entries=%+v, want one production support row", matrix.Entries)
	}
	entry := matrix.Entries[0]
	if entry.Target != "opencode" || entry.Version != "1.18.27" || entry.Construct != "permission" || entry.Fixture != openCodeLegacyFixtureID || entry.Evidence != openCodeLegacyEvidenceID {
		t.Fatalf("entry tuple/refs=%+v", entry)
	}
	if !entry.PermissionBearing || !entry.DefaultsKnown || !entry.MatcherKnown || !entry.ConditionsKnown {
		t.Fatalf("entry flags=%+v, want all semantic gates true for admitted scalar shape", entry)
	}
	if !matrix.Registry.Fixtures[openCodeLegacyFixtureID] || len(matrix.Registry.Fixtures) != 1 {
		t.Fatalf("fixture registry=%+v", matrix.Registry.Fixtures)
	}
	if !matrix.Registry.Evidence[openCodeLegacyEvidenceID] || len(matrix.Registry.Evidence) != 1 {
		t.Fatalf("evidence registry=%+v", matrix.Registry.Evidence)
	}
	result := Validate(matrix, Entry{Target: "opencode", Version: "1.18.27", Construct: "permission"})
	if result.Completeness != model.CompletenessComplete || len(result.Findings) != 0 {
		t.Fatalf("Validate result=%+v", result)
	}
	for _, tt := range []struct {
		name  string
		entry Entry
		code  string
	}{
		{"nearby version unsupported", Entry{Target: "opencode", Version: "1.18.28", Construct: "permission"}, "unsupported-version"},
		{"claude unsupported", Entry{Target: "claudecode", Version: "1.18.27", Construct: "permission"}, "unsupported-version"},
		{"construct unsupported", Entry{Target: "opencode", Version: "1.18.27", Construct: "resource_map"}, "unknown-construct"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(matrix, tt.entry)
			if result.Completeness != model.CompletenessIncomplete || !hasCode(result.Findings, tt.code) {
				t.Fatalf("Validate result=%+v", result)
			}
		})
	}
}

func TestValidateRejectsFalseSemanticFlagsOnReferencedRows(t *testing.T) {
	complete := Entry{Target: "opencode", Version: "1.18.27", Construct: "permission", Fixture: openCodeLegacyFixtureID, Evidence: openCodeLegacyEvidenceID, PermissionBearing: true, DefaultsKnown: true, MatcherKnown: true, ConditionsKnown: true}
	for _, tt := range []struct {
		name   string
		mutate func(*Entry)
	}{
		{"permission-bearing false", func(entry *Entry) { entry.PermissionBearing = false }},
		{"defaults-known false", func(entry *Entry) { entry.DefaultsKnown = false }},
		{"matcher-known false", func(entry *Entry) { entry.MatcherKnown = false }},
		{"conditions-known false", func(entry *Entry) { entry.ConditionsKnown = false }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			entry := complete
			tt.mutate(&entry)
			matrix := Matrix{Entries: []Entry{entry}, Registry: Registry{Fixtures: map[string]bool{openCodeLegacyFixtureID: true}, Evidence: map[string]bool{openCodeLegacyEvidenceID: true}}}
			result := Validate(matrix, complete)
			if result.Completeness != model.CompletenessIncomplete || !hasCode(result.Findings, "incomplete-evidence") {
				t.Fatalf("Validate result=%+v", result)
			}
		})
	}
}

func TestCheckedInOpenCodeMatrixRejectsMissingReferences(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mutate func(map[string]string)
	}{
		{"missing fixture", func(files map[string]string) {
			files["matrix.json"] = strings.Replace(files["matrix.json"], `"fixture": "`+openCodeLegacyFixtureID+`"`, `"fixture": "missing"`, 1)
		}},
		{"missing evidence", func(files map[string]string) {
			files["matrix.json"] = strings.Replace(files["matrix.json"], `"evidence": "`+openCodeLegacyEvidenceID+`"`, `"evidence": "missing"`, 1)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Load(checkedInOpenCodeFS(t, tt.mutate)); err == nil {
				t.Fatal("Load accepted a production matrix row with missing references")
			}
		})
	}
}

func TestCheckedInOpenCodePinnedRegistryEvidenceAndFixture(t *testing.T) {
	fixtures, evidence := readCheckedInSupportDocs(t)
	if len(fixtures.Fixtures) != 1 {
		t.Fatalf("fixture records=%d", len(fixtures.Fixtures))
	}
	fixture := fixtures.Fixtures[0]
	if fixture.ID != openCodeLegacyFixtureID || fixture.Target != "opencode" || fixture.Version != "1.18.27" || fixture.Construct != "permission" || fixture.Path != openCodeLegacyFixturePath || fixture.SHA256 != openCodeLegacyFixtureSHA256 {
		t.Fatalf("fixture=%+v", fixture)
	}
	data, err := os.ReadFile(fixture.Path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != openCodeLegacyFixtureSHA256 {
		t.Fatalf("fixture digest=%s", got)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 1 || body["permission"] == nil {
		t.Fatalf("fixture has extra top-level semantics: %+v", body)
	}
	var permission map[string]string
	if err := json.Unmarshal(body["permission"], &permission); err != nil {
		t.Fatal(err)
	}
	if len(permission) != 3 || permission["read"] != "allow" || permission["edit"] != "deny" || permission["bash"] != "ask" {
		t.Fatalf("permission=%+v", permission)
	}

	if len(evidence.Evidence) != 1 {
		t.Fatalf("evidence records=%d", len(evidence.Evidence))
	}
	record := evidence.Evidence[0]
	if record.ID != openCodeLegacyEvidenceID || record.Target != fixture.Target || record.Version != fixture.Version || record.Construct != fixture.Construct || record.Fixture != fixture.ID {
		t.Fatalf("evidence tuple=%+v fixture=%+v", record, fixture)
	}
	if record.Authority.Repository != "https://github.com/anomalyco/opencode" || record.Authority.Tag != "v1.18.27" || record.Authority.Commit != "4b7e19e315cca414121ba1d61523fef74bb3ae8b" {
		t.Fatalf("authority=%+v", record.Authority)
	}
	assertStrings(t, record.Claims, openCodeLegacyClaims)
	if len(record.Sources) != len(openCodeLegacySources) {
		t.Fatalf("sources=%+v", record.Sources)
	}
	for i, source := range record.Sources {
		if source != openCodeLegacySources[i] {
			t.Fatalf("source[%d]=%+v want %+v", i, source, openCodeLegacySources[i])
		}
	}
}

func TestCheckedInOpenCodeRegistryRejectsFixtureDataDrift(t *testing.T) {
	if _, err := Load(checkedInOpenCodeFS(t, func(files map[string]string) {
		files[openCodeLegacyFixturePath] = strings.Replace(files[openCodeLegacyFixturePath], `"allow"`, `"deny"`, 1)
	})); err == nil {
		t.Fatal("Load accepted fixture bytes that no longer match the pinned digest")
	}
}

func TestCheckedInOpenCodeAuditCLIStillExitsUnsupported(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "auditor")
	build := exec.Command("go", "build", "-o", binary, "./cmd/auditor")
	build.Dir = repoRoot
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build err=%v output=%q", err, output)
	}
	cmd := exec.Command(binary, "audit", "opencode", "--root", repoRoot, "--config", filepath.Join(repoRoot, "support", openCodeLegacyFixturePath), "--opencode-version", "1.18.27")
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 2 {
		t.Fatalf("auditor exit err=%v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	if stdout.String() != "unsupported_or_incomplete\n" || stderr.String() != "" {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

type metadataDocs struct{ matrix, fixtures, evidence string }

const fixtureBytes = "generic fixture\n"

func metadataFS(mutate func(*metadataDocs)) fstest.MapFS {
	docs := metadataDocs{
		matrix:   `{"schema":"support-matrix/v1","entries":[{"target":"target","version":"1","construct":"permission","fixture":"fixture","evidence":"evidence","permission_bearing":true,"defaults_known":true,"matcher_known":true,"conditions_known":true}],"gate":"optional note"}`,
		fixtures: `{"schema":"conformance-fixtures/v1","fixtures":[{"id":"fixture","target":"target","version":"1","construct":"permission","path":"fixtures/generic.txt","sha256":"` + fixtureHash() + `"}]}`,
		evidence: `{"schema":"support-evidence/v1","evidence":[{"id":"evidence","target":"target","version":"1","construct":"permission","fixture":"fixture","authority":{"repository":"https://github.com/owner/.github","tag":"v1.18.27","commit":"` + strings.Repeat("b", 40) + `"},"claims":["claim:defaults","claim:matcher"],"sources":[{"path":"docs/source.md","sha256":"` + strings.Repeat("a", 64) + `"}]}]}`,
	}
	if mutate != nil {
		mutate(&docs)
	}
	return supportFS(map[string]string{
		"matrix.json":          docs.matrix,
		"fixtures.json":        docs.fixtures,
		"evidence.json":        docs.evidence,
		"fixtures/generic.txt": fixtureBytes,
	})
}

func expectLoadError(t *testing.T, mutate func(*metadataDocs)) {
	t.Helper()
	if _, err := Load(metadataFS(mutate)); err == nil {
		t.Fatal("Load accepted invalid metadata")
	}
}

func fixtureHash() string { return fmt.Sprintf("%x", sha256.Sum256([]byte(fixtureBytes))) }

func replaceOnce(s *string, old, new string) { *s = strings.Replace(*s, old, new, 1) }

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

const (
	openCodeLegacyFixtureID     = "opencode-1.18.27-legacy-permission-scalar"
	openCodeLegacyEvidenceID    = "opencode-1.18.27-legacy-permission-scalar-source"
	openCodeLegacyFixturePath   = "testdata/opencode-1.18.27-permission-legacy-scalar.json"
	openCodeLegacyFixtureSHA256 = "22ab8de00a73350aedcb72b62db5c962c910f15e12fbef80d844e725593bacf9"
)

var openCodeLegacyClaims = []string{
	"root-permission-legacy-info",
	"legacy-actions-scalar-resource-map",
	"scalar-migrates-wildcard-action-effect",
	"v2-ordered-last-wildcard-fallback-ask",
	"runtime-effective-policy-out-of-scope",
}

var openCodeLegacySources = []sourceRecord{
	{Path: "packages/core/src/v1/config/config.ts", SHA256: "b99bcbd98df6da79e59cda482363f759cea9d4b9792c6c8e83b6a8d686138d30"},
	{Path: "packages/core/src/v1/config/permission.ts", SHA256: "f7669733d939c4affd38c4a27ce14deff970c7ffa4e26071ce6a3cd39ce0d4dd"},
	{Path: "packages/core/src/v1/config/migrate.ts", SHA256: "2a23a56469575d3edfe2044fe52468ee499d1c1c2281c34ebbcb624c7b25dae6"},
	{Path: "packages/schema/src/permission.ts", SHA256: "229f2da7d245bf86deffed59b10cddf2dd04ffa05d10ef97fc5df289d0e6fb98"},
	{Path: "packages/core/src/permission.ts", SHA256: "f5d5b295452f2ddf17c475a24e77e302bb2b4b65d8a9e02a0f7bc5ef5504a0ba"},
}

func readCheckedInSupportDocs(t *testing.T) (fixturesDocument, evidenceDocument) {
	t.Helper()
	var fixtures fixturesDocument
	if err := decode(os.DirFS("."), "fixtures.json", &fixtures); err != nil {
		t.Fatal(err)
	}
	var evidence evidenceDocument
	if err := decode(os.DirFS("."), "evidence.json", &evidence); err != nil {
		t.Fatal(err)
	}
	return fixtures, evidence
}

func checkedInOpenCodeFS(t *testing.T, mutate func(map[string]string)) fstest.MapFS {
	t.Helper()
	files := map[string]string{}
	for _, path := range []string{"matrix.json", "fixtures.json", "evidence.json", openCodeLegacyFixturePath} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		files[path] = string(data)
	}
	if mutate != nil {
		mutate(files)
	}
	return supportFS(files)
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("strings=%+v want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("strings[%d]=%q want %q", i, got[i], want[i])
		}
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
