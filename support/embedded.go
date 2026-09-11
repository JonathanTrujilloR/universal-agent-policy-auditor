package support

import "embed"

//go:embed matrix.json fixtures.json evidence.json testdata/opencode-1.18.27-permission-legacy-scalar.json
var checkedInFS embed.FS

// LoadCheckedIn loads the immutable support metadata compiled into the binary.
func LoadCheckedIn() (Matrix, error) { return Load(checkedInFS) }
