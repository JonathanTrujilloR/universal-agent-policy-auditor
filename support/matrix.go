// Package support validates evidence-gated native-policy support metadata.
package support

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"strings"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
)

const maxJSONDepth = 8

type Entry struct {
	Target            string `json:"target"`
	Version           string `json:"version"`
	Construct         string `json:"construct"`
	Fixture           string `json:"fixture"`
	Evidence          string `json:"evidence"`
	PermissionBearing bool   `json:"permission_bearing"`
	DefaultsKnown     bool   `json:"defaults_known"`
	MatcherKnown      bool   `json:"matcher_known"`
	ConditionsKnown   bool   `json:"conditions_known"`
}
type Registry struct{ Fixtures, Evidence map[string]bool }
type Matrix struct {
	Entries  []Entry
	Registry Registry
}

type fixtureRecord struct {
	ID        string `json:"id"`
	Target    string `json:"target"`
	Version   string `json:"version"`
	Construct string `json:"construct"`
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
}
type evidenceRecord struct {
	ID        string          `json:"id"`
	Target    string          `json:"target"`
	Version   string          `json:"version"`
	Construct string          `json:"construct"`
	Fixture   string          `json:"fixture"`
	Authority authorityRecord `json:"authority"`
	Claims    []string        `json:"claims"`
	Sources   []sourceRecord  `json:"sources"`
}
type authorityRecord struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Commit     string `json:"commit"`
}
type sourceRecord struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type matrixDocument struct {
	Schema  string  `json:"schema"`
	Entries []Entry `json:"entries"`
	Gate    string  `json:"gate,omitempty"`
}
type fixturesDocument struct {
	Schema   string          `json:"schema"`
	Fixtures []fixtureRecord `json:"fixtures"`
}
type evidenceDocument struct {
	Schema   string           `json:"schema"`
	Evidence []evidenceRecord `json:"evidence"`
}
type Result struct {
	Completeness model.Completeness
	Findings     []model.Finding
}

func Validate(matrix Matrix, entry Entry) Result {
	if entry.Target == "" || entry.Version == "" || entry.Construct == "" {
		return incomplete("malformed-entry")
	}
	if entry.Construct != "permission" {
		return incomplete("unknown-construct")
	}
	row, ok := canonical(matrix, entry)
	if !ok {
		return incomplete("unsupported-version")
	}
	if !row.PermissionBearing || row.Fixture == "" || row.Evidence == "" || !row.DefaultsKnown || !row.MatcherKnown || !row.ConditionsKnown {
		return incomplete("incomplete-evidence")
	}
	if !matrix.Registry.Fixtures[row.Fixture] {
		return incomplete("missing-fixture")
	}
	if !matrix.Registry.Evidence[row.Evidence] {
		return incomplete("missing-evidence")
	}
	return Result{Completeness: model.CompletenessComplete}
}
func incomplete(code string) Result {
	return Result{Completeness: model.CompletenessIncomplete, Findings: []model.Finding{model.NewFinding(code, "support metadata is incomplete")}}
}
func canonical(matrix Matrix, entry Entry) (Entry, bool) {
	for _, candidate := range matrix.Entries {
		if candidate.Target == entry.Target && candidate.Version == entry.Version && candidate.Construct == entry.Construct {
			return candidate, true
		}
	}
	return Entry{}, false
}

// Load validates the checked-in metadata. Fixture paths are lexical fs.FS paths; this is not an OS or symlink sandbox.
func Load(files fs.FS) (Matrix, error) {
	var doc matrixDocument
	if err := decode(files, "matrix.json", &doc); err != nil {
		return Matrix{}, err
	}
	if doc.Schema != "support-matrix/v1" {
		return Matrix{}, fmt.Errorf("matrix.json: unsupported schema %q", doc.Schema)
	}
	if doc.Entries == nil {
		return Matrix{}, fmt.Errorf("matrix.json: entries array is required")
	}

	var fixtures fixturesDocument
	if err := decode(files, "fixtures.json", &fixtures); err != nil {
		return Matrix{}, err
	}
	if fixtures.Schema != "conformance-fixtures/v1" {
		return Matrix{}, fmt.Errorf("fixtures.json: unsupported schema %q", fixtures.Schema)
	}
	if fixtures.Fixtures == nil {
		return Matrix{}, fmt.Errorf("fixtures.json: fixtures array is required")
	}

	var evidence evidenceDocument
	if err := decode(files, "evidence.json", &evidence); err != nil {
		return Matrix{}, err
	}
	if evidence.Schema != "support-evidence/v1" {
		return Matrix{}, fmt.Errorf("evidence.json: unsupported schema %q", evidence.Schema)
	}
	if evidence.Evidence == nil {
		return Matrix{}, fmt.Errorf("evidence.json: evidence array is required")
	}

	fixtureRecords, err := validateFixtures(files, fixtures.Fixtures)
	if err != nil {
		return Matrix{}, err
	}
	evidenceRecords, err := validateEvidence(evidence.Evidence, fixtureRecords)
	if err != nil {
		return Matrix{}, err
	}
	if err := validateEntries(doc.Entries, fixtureRecords, evidenceRecords); err != nil {
		return Matrix{}, err
	}
	matrix := Matrix{Entries: doc.Entries, Registry: Registry{Fixtures: map[string]bool{}, Evidence: map[string]bool{}}}
	for id := range fixtureRecords {
		matrix.Registry.Fixtures[id] = true
	}
	for id := range evidenceRecords {
		matrix.Registry.Evidence[id] = true
	}
	return matrix, nil
}

func validateFixtures(files fs.FS, records []fixtureRecord) (map[string]fixtureRecord, error) {
	validated := map[string]fixtureRecord{}
	for _, record := range records {
		if err := requireScalars("fixture", record.ID, record.Target, record.Version, record.Construct, record.Path, record.SHA256); err != nil {
			return nil, err
		}
		if _, exists := validated[record.ID]; exists {
			return nil, fmt.Errorf("duplicate fixture id %q", record.ID)
		}
		if !validFSPath(record.Path) || !lowerHex(record.SHA256, 64) {
			return nil, fmt.Errorf("invalid fixture path or sha256 for %q", record.ID)
		}
		data, err := fs.ReadFile(files, record.Path)
		if err != nil {
			return nil, fmt.Errorf("fixture %q: %w", record.ID, err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != record.SHA256 {
			return nil, fmt.Errorf("fixture %q digest mismatch", record.ID)
		}
		validated[record.ID] = record
	}
	return validated, nil
}

func validateEvidence(records []evidenceRecord, fixtures map[string]fixtureRecord) (map[string]evidenceRecord, error) {
	validated := map[string]evidenceRecord{}
	for _, record := range records {
		if err := requireScalars("evidence", record.ID, record.Target, record.Version, record.Construct, record.Fixture); err != nil {
			return nil, err
		}
		fixture, ok := fixtures[record.Fixture]
		if _, exists := validated[record.ID]; exists {
			return nil, fmt.Errorf("duplicate evidence id %q", record.ID)
		}
		if !ok || !sameTuple(record.Target, record.Version, record.Construct, fixture.Target, fixture.Version, fixture.Construct) {
			return nil, fmt.Errorf("evidence %q does not match a fixture", record.ID)
		}
		if err := validateAuthority(record.Authority); err != nil {
			return nil, fmt.Errorf("evidence %q: %w", record.ID, err)
		}
		if err := validateClaims(record.Claims); err != nil {
			return nil, fmt.Errorf("evidence %q: %w", record.ID, err)
		}
		if err := validateSources(record.Sources); err != nil {
			return nil, fmt.Errorf("evidence %q: %w", record.ID, err)
		}
		validated[record.ID] = record
	}
	return validated, nil
}

func validateEntries(entries []Entry, fixtures map[string]fixtureRecord, evidence map[string]evidenceRecord) error {
	seen := map[[3]string]bool{}
	for _, entry := range entries {
		if err := requireScalars("matrix", entry.Target, entry.Version, entry.Construct, entry.Fixture, entry.Evidence); err != nil {
			return err
		}
		triple := [3]string{entry.Target, entry.Version, entry.Construct}
		fixture, fixtureOK := fixtures[entry.Fixture]
		evidenceRecord, evidenceOK := evidence[entry.Evidence]
		if seen[triple] || !fixtureOK || !evidenceOK {
			return fmt.Errorf("matrix row has duplicate tuple or missing references")
		}
		seen[triple] = true
		if !sameTuple(entry.Target, entry.Version, entry.Construct, fixture.Target, fixture.Version, fixture.Construct) || !sameTuple(entry.Target, entry.Version, entry.Construct, evidenceRecord.Target, evidenceRecord.Version, evidenceRecord.Construct) || evidenceRecord.Fixture != entry.Fixture {
			return fmt.Errorf("matrix row tuple does not match fixture and evidence")
		}
	}
	return nil
}

func validateAuthority(authority authorityRecord) error {
	if err := requireScalars("authority", authority.Repository, authority.Tag, authority.Commit); err != nil {
		return err
	}
	if !validGitHubRepository(authority.Repository) || !validGitTag(authority.Tag) || !lowerHex(authority.Commit, 40) {
		return fmt.Errorf("invalid authority")
	}
	return nil
}

func validateClaims(claims []string) error {
	if len(claims) == 0 {
		return fmt.Errorf("claims are required")
	}
	seen := map[string]bool{}
	for _, claim := range claims {
		if err := requireScalars("claim", claim); err != nil {
			return err
		}
		if seen[claim] {
			return fmt.Errorf("duplicate claim %q", claim)
		}
		seen[claim] = true
	}
	return nil
}

func validateSources(sources []sourceRecord) error {
	if len(sources) == 0 {
		return fmt.Errorf("sources are required")
	}
	seen := map[string]bool{}
	for _, source := range sources {
		if err := requireScalars("source", source.Path, source.SHA256); err != nil {
			return err
		}
		if seen[source.Path] || !validFSPath(source.Path) || !lowerHex(source.SHA256, 64) {
			return fmt.Errorf("invalid or duplicate source %q", source.Path)
		}
		seen[source.Path] = true
	}
	return nil
}

func requireScalars(kind string, values ...string) error {
	for _, value := range values {
		if value == "" || strings.TrimSpace(value) != value {
			return fmt.Errorf("%s has a missing or whitespace-padded required field", kind)
		}
	}
	return nil
}

func sameTuple(aTarget, aVersion, aConstruct, bTarget, bVersion, bConstruct string) bool {
	return aTarget == bTarget && aVersion == bVersion && aConstruct == bConstruct
}

func validFSPath(path string) bool { return path != "." && fs.ValidPath(path) }

func validGitHubRepository(raw string) bool {
	match := regexp.MustCompile(`\Ahttps://github\.com/([A-Za-z0-9](?:[A-Za-z0-9-]{0,37}[A-Za-z0-9])?)/([A-Za-z0-9._-]{1,100})\z`).FindStringSubmatch(raw)
	if match == nil || strings.Contains(match[1], "--") {
		return false
	}
	repo := match[2]
	return repo != "." && repo != ".." && !strings.HasSuffix(repo, ".")
}

func validGitTag(tag string) bool {
	if tag == "." || tag == ".." || strings.HasPrefix(tag, "/") || strings.HasSuffix(tag, "/") || strings.HasPrefix(tag, ".") || strings.HasSuffix(tag, ".") || strings.HasPrefix(tag, "-") || strings.HasSuffix(tag, "-") || strings.Contains(tag, "//") || strings.Contains(tag, "@{") || strings.Contains(tag, "..") {
		return false
	}
	for _, r := range tag {
		if r <= ' ' || r == 0x7f || strings.ContainsRune(`~^:?*[\\`, r) {
			return false
		}
	}
	for _, part := range strings.Split(tag, "/") {
		if part == "" || strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".lock") {
			return false
		}
	}
	return true
}

func lowerHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, r := range value {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') {
			continue
		}
		return false
	}
	return true
}

func decode(files fs.FS, name string, into any) error {
	data, err := fs.ReadFile(files, name)
	if err != nil {
		return err
	}
	if err := checkStrictJSON(data); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("%s: trailing JSON document", name)
	}
	return nil
}

func checkStrictJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder, 0); err != nil {
		return err
	}
	if token, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("trailing JSON document after %v", token)
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder, depth int) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}

	depth++
	if depth > maxJSONDepth {
		return fmt.Errorf("JSON nesting exceeds %d", maxJSONDepth)
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if !canonicalJSONKey(key) {
				return fmt.Errorf("noncanonical object key %q", key)
			}
			if seen[key] {
				return fmt.Errorf("duplicate object key %q", key)
			}
			seen[key] = true
			if err := scanJSONValue(decoder, depth); err != nil {
				return err
			}
		}
		return closeDelim(decoder, '}')
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder, depth); err != nil {
				return err
			}
		}
		return closeDelim(decoder, ']')
	default:
		return fmt.Errorf("unexpected closing delimiter %q", delim)
	}
}

func closeDelim(decoder *json.Decoder, want json.Delim) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token != want {
		return fmt.Errorf("expected closing delimiter %q", want)
	}
	return nil
}

func canonicalJSONKey(key string) bool {
	if len(key) == 0 || key[0] < 'a' || key[0] > 'z' {
		return false
	}
	for i := 1; i < len(key); i++ {
		if key[i] >= 'a' && key[i] <= 'z' {
			continue
		}
		if key[i] >= '0' && key[i] <= '9' {
			continue
		}
		if key[i] == '_' {
			continue
		}
		return false
	}
	return true
}
