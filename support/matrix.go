// Package support validates evidence-gated native-policy support metadata.
package support

import (
	"encoding/json"
	"io/fs"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
)

type Entry struct {
	Target, Version, Construct, Fixture, Evidence                   string
	PermissionBearing, DefaultsKnown, MatcherKnown, ConditionsKnown bool
}
type Registry struct{ Fixtures, Evidence map[string]bool }
type Matrix struct {
	Entries  []Entry
	Registry Registry
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
	var matrix Matrix
	if err := decode(files, "matrix.json", &matrix); err != nil {
		return Matrix{}, err
	}
	var fixtures struct{ Fixtures []struct{ ID string } }
	if err := decode(files, "fixtures.json", &fixtures); err != nil {
		return Matrix{}, err
	}
	var evidence struct{ Evidence []struct{ ID string } }
	if err := decode(files, "evidence.json", &evidence); err != nil {
		return Matrix{}, err
	}
	matrix.Registry = Registry{Fixtures: map[string]bool{}, Evidence: map[string]bool{}}
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
	return json.Unmarshal(data, into)
}
