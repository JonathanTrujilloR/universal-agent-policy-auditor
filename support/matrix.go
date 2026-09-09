// Package support validates evidence-gated native-policy support metadata.
package support

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"

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

type matrixDocument struct {
	Schema  string  `json:"schema"`
	Entries []Entry `json:"entries"`
	Gate    string  `json:"gate,omitempty"`
}
type fixturesDocument struct {
	Schema   string `json:"schema"`
	Fixtures []struct {
		ID string `json:"id"`
	} `json:"fixtures"`
}
type evidenceDocument struct {
	Schema   string `json:"schema"`
	Evidence []struct {
		ID string `json:"id"`
	} `json:"evidence"`
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

// Load makes the checked-in matrix and its registries executable inputs.
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

	matrix := Matrix{Entries: doc.Entries, Registry: Registry{Fixtures: map[string]bool{}, Evidence: map[string]bool{}}}
	for _, item := range fixtures.Fixtures {
		matrix.Registry.Fixtures[item.ID] = true
	}
	for _, item := range evidence.Evidence {
		matrix.Registry.Evidence[item.ID] = true
	}
	return matrix, nil
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
