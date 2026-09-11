package app

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jkelevra/universal-agent-policy-auditor/support"
)

func TestExitCategoryCodesAreStableAndFailClosed(t *testing.T) {
	for _, tt := range []struct {
		category Category
		want     int
	}{
		{CompleteNoFindings, 0},
		{CompleteWithFindings, 1},
		{UnsupportedOrIncomplete, 2},
		{InvalidRequest, 3},
		{OperationalFailure, 4},
		{Category("surprise"), 4},
	} {
		if got := ExitCode(tt.category); got != tt.want {
			t.Fatalf("ExitCode(%q)=%d want %d", tt.category, got, tt.want)
		}
	}
}

func TestAuditShellClassifiesRequestsWithoutMutatingConfig(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "opencode.json")
	content := []byte(`{"permission":{"bash":"allow"}}`)
	if err := os.WriteFile(config, content, 0o640); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(config)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name string
		req  Request
		want Category
	}{
		{"valid opencode remains unsupported until evidence exists", Request{Mode: ModeAudit, Target: TargetOpenCode, Root: root, ExplicitConfig: config, TargetVersion: "1.0.0"}, UnsupportedOrIncomplete},
		{"missing opencode version is unsupported", Request{Mode: ModeAudit, Target: TargetOpenCode, Root: root, ExplicitConfig: config}, UnsupportedOrIncomplete},
		{"claude is unsupported", Request{Mode: ModeAudit, Target: TargetClaudeCode, Root: root, ExplicitConfig: config, TargetVersion: "1"}, UnsupportedOrIncomplete},
		{"relative paths are invalid", Request{Mode: ModeAudit, Target: TargetOpenCode, Root: ".", ExplicitConfig: "opencode.json", TargetVersion: "1"}, InvalidRequest},
		{"claude missing root is invalid", Request{Mode: ModeAudit, Target: TargetClaudeCode, ExplicitConfig: config, TargetVersion: "1"}, InvalidRequest},
		{"claude outside root is invalid", Request{Mode: ModeAudit, Target: TargetClaudeCode, Root: root, ExplicitConfig: filepath.Join(t.TempDir(), "opencode.json"), TargetVersion: "1"}, InvalidRequest},
		{"missing config is invalid", Request{Mode: ModeAudit, Target: TargetOpenCode, Root: root, TargetVersion: "1"}, InvalidRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := Run(tt.req).Category; got != tt.want {
				t.Fatalf("category=%q want %q", got, tt.want)
			}
		})
	}
	after, err := os.Stat(config)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) || after.Mode() != before.Mode() || !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("config mutated: before=%v after=%v content=%q", before.Mode(), after.Mode(), got)
	}
}

func TestAuditSupportedOpenCodeScalarConfigCompletesAndPreservesFile(t *testing.T) {
	fixture := readFile(t, filepath.Join("..", "..", "support", "testdata", "opencode-1.18.27-permission-legacy-scalar.json"))
	for _, tt := range []struct {
		name    string
		body    []byte
		effects map[string]string
	}{
		{"registered fixture bytes", fixture, map[string]string{"read": "allow", "edit": "deny", "bash": "ask"}},
		{"structurally exact non-fixture scalars", []byte(`{"permission":{"read":"deny","edit":"allow","bash":"deny"}}`), map[string]string{"read": "deny", "edit": "allow", "bash": "deny"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root, config := writeConfig(t, tt.body)
			before := snapshot(t, config)
			t.Setenv("OPENCODE_CONFIG", filepath.Join(t.TempDir(), "ignored.json"))
			result := Run(Request{Mode: ModeAudit, Target: TargetOpenCode, Root: root, ExplicitConfig: config, TargetVersion: "1.18.27"})
			if result.Category != CompleteNoFindings || string(result.Completeness) != "complete" || len(result.Findings) != 0 || len(result.Permissions) != 3 || !hasString(result.Limitations, "modeled-static-configuration") {
				t.Fatalf("result=%+v", result)
			}
			for _, permission := range result.Permissions {
				if string(permission.Effect()) != tt.effects[permission.Capability()] || permission.Matcher() != "*" {
					t.Fatalf("permission=%+v", permission)
				}
			}
			assertUnchanged(t, config, tt.body, before)
		})
	}
}

func TestAuditUnsupportedAdmissionDoesNotReadTarget(t *testing.T) {
	root := filepath.Join(t.TempDir(), "root")
	config := filepath.Join(root, "opencode.json")
	for _, req := range []Request{
		{Mode: ModeAudit, Target: TargetOpenCode, Root: root, ExplicitConfig: config},
		{Mode: ModeAudit, Target: TargetOpenCode, Root: root, ExplicitConfig: config, TargetVersion: "1.18.28"},
		{Mode: ModeAudit, Target: TargetClaudeCode, Root: root, ExplicitConfig: config, TargetVersion: "1.18.27"},
	} {
		files := &fakeFS{}
		result := runWithOptions(req, options{files: files, loadSupport: support.LoadCheckedIn})
		if result.Category != UnsupportedOrIncomplete || files.calls != 0 {
			t.Fatalf("result=%+v filesystem calls=%d", result, files.calls)
		}
	}
}

func TestAuditInputFailuresAreUnsupportedOrIncomplete(t *testing.T) {
	good := []byte(`{"permission":{"read":"allow","edit":"deny","bash":"ask"}}`)
	for _, tt := range []struct {
		name string
		body []byte
		prep func(*testing.T, string, string) (string, string, []byte)
	}{
		{"malformed json", []byte(`{`), nil},
		{"duplicate key", []byte(`{"permission":{"read":"allow","read":"deny","edit":"deny","bash":"ask"}}`), nil},
		{"extra key", []byte(`{"permission":{"read":"allow","edit":"deny","bash":"ask"},"mode":{}}`), nil},
		{"missing scalar", []byte(`{"permission":{"read":"allow","edit":"deny"}}`), nil},
		{"resource object", []byte(`{"permission":{"read":{"*":"allow"},"edit":"deny","bash":"ask"}}`), nil},
		{"nonexistent root", good, func(t *testing.T, _, _ string) (string, string, []byte) {
			root := filepath.Join(t.TempDir(), "missing")
			return root, filepath.Join(root, "opencode.json"), nil
		}},
		{"nonexistent", good, func(_ *testing.T, root, config string) (string, string, []byte) {
			return root, config + ".missing", nil
		}},
		{"directory", good, func(t *testing.T, root, _ string) (string, string, []byte) {
			dir := filepath.Join(root, "dir")
			must(t, os.Mkdir(dir, 0o700))
			return root, dir, nil
		}},
		{"escaping symlink", good, func(t *testing.T, root, config string) (string, string, []byte) {
			outside := filepath.Join(t.TempDir(), "outside.json")
			must(t, os.WriteFile(outside, good, 0o600))
			must(t, os.Remove(config))
			must(t, os.Symlink(outside, config))
			return root, config, good
		}},
		{"oversize", append([]byte(`{"permission":{"read":"allow","edit":"deny","bash":"ask"},"pad":"`), append(bytes.Repeat([]byte("x"), 300<<10), []byte(`"}`)...)...), nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root, config := writeConfig(t, tt.body)
			if tt.prep != nil {
				root, config, tt.body = tt.prep(t, root, config)
			}
			before, snapOK := maybeSnapshot(t, config)
			result := Run(Request{Mode: ModeAudit, Target: TargetOpenCode, Root: root, ExplicitConfig: config, TargetVersion: "1.18.27"})
			if result.Category != UnsupportedOrIncomplete {
				t.Fatalf("result=%+v", result)
			}
			if snapOK {
				assertUnchanged(t, config, tt.body, before)
			}
		})
	}
}

func TestAuditOperationalAndReadFailuresFailClosed(t *testing.T) {
	root := filepath.Join(t.TempDir(), "root")
	config := filepath.Join(root, "opencode.json")
	request := Request{Mode: ModeAudit, Target: TargetOpenCode, Root: root, ExplicitConfig: config, TargetVersion: "1.18.27"}
	loaderErr := runWithOptions(request, options{files: &fakeFS{}, loadSupport: func() (support.Matrix, error) { return support.Matrix{}, errors.New("boom") }})
	if loaderErr.Category != OperationalFailure || loaderErr.Message != "support metadata unavailable" {
		t.Fatalf("loader result=%+v", loaderErr)
	}
	if typedNil := runWithOptions(request, options{files: (*fakeFS)(nil)}); typedNil.Category != OperationalFailure {
		t.Fatalf("typed-nil result=%+v", typedNil)
	}
	readErr := runWithOptions(request, options{files: &fakeFS{root: root, path: config, readErr: true}, loadSupport: support.LoadCheckedIn})
	if readErr.Category != UnsupportedOrIncomplete {
		t.Fatalf("read-error result=%+v", readErr)
	}
}

func writeConfig(t *testing.T, body []byte) (string, string) {
	t.Helper()
	root := t.TempDir()
	config := filepath.Join(root, "opencode.json")
	must(t, os.WriteFile(config, body, 0o640))
	return root, config
}
func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	must(t, err)
	return b
}
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type fileSnapshot struct {
	mode fs.FileMode
	mod  time.Time
	data []byte
}

func snapshot(t *testing.T, path string) fileSnapshot {
	t.Helper()
	info, err := os.Stat(path)
	must(t, err)
	return fileSnapshot{info.Mode(), info.ModTime(), readFile(t, path)}
}
func maybeSnapshot(t *testing.T, path string) (fileSnapshot, bool) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return fileSnapshot{}, false
	}
	return fileSnapshot{info.Mode(), info.ModTime(), readFile(t, path)}, true
}
func assertUnchanged(t *testing.T, path string, want []byte, before fileSnapshot) {
	t.Helper()
	after := snapshot(t, path)
	if !bytes.Equal(after.data, want) || !bytes.Equal(after.data, before.data) || after.mode != before.mode || !after.mod.Equal(before.mod) {
		t.Fatalf("config mutated before=%+v after=%+v", before, after)
	}
}
func hasString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

type fakeFS struct {
	calls      int
	root, path string
	readErr    bool
}

func (f *fakeFS) ReadFile(string) ([]byte, error) {
	f.calls++
	if f.readErr {
		return nil, errors.New("read failed")
	}
	return nil, fs.ErrNotExist
}
func (f *fakeFS) Stat(string) (fs.FileInfo, error) { f.calls++; return fakeInfo{}, nil }
func (f *fakeFS) EvalSymlinks(name string) (string, error) {
	f.calls++
	if f.root == "" || name == f.root || name == f.path {
		return filepath.Clean(name), nil
	}
	return "", fs.ErrNotExist
}

type fakeInfo struct{}

func (fakeInfo) Name() string       { return "opencode.json" }
func (fakeInfo) Size() int64        { return 1 }
func (fakeInfo) Mode() fs.FileMode  { return 0o600 }
func (fakeInfo) ModTime() time.Time { return time.Unix(1, 0) }
func (fakeInfo) IsDir() bool        { return false }
func (fakeInfo) Sys() any           { return nil }
