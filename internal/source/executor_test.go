package source

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestExecuteRejectsUnsafeSources(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "safe.json")
	if err := os.WriteFile(inside, []byte(`{"token":"super-secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "escape.json")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	cycle := filepath.Join(root, "cycle.json")
	if err := os.Symlink(cycle, cycle); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	for _, tt := range []struct {
		name, candidate, want string
	}{
		{"traversal", filepath.Join(root, "..", filepath.Base(outside), "outside.json"), "outside-root"},
		{"symlink escape", link, "outside-root"},
		{"symlink cycle", cycle, "unreadable"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := Execute(context.Background(), Plan{Roots: []string{root}, Sources: []string{tt.candidate}}, OSReadFS{}, EnvSnapshot{})
			if len(r.Selected) != 0 || !hasCode(r.Findings, tt.want) {
				t.Fatalf("selected=%d findings=%+v", len(r.Selected), r.Findings)
			}
		})
	}
}

func TestExecuteBoundsDeduplicatesAndRedacts(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "config.json")
	if err := os.WriteFile(file, []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	plan := Plan{Roots: []string{root}, Sources: []string{file, file}, MaxFiles: 1, MaxBytes: 4, EnvKeys: []string{"HOME"}}
	r := Execute(context.Background(), plan, OSReadFS{}, CaptureEnv(plan.EnvKeys, func(key string) (string, bool) { return "super-secret", true }))
	if len(r.Selected) != 0 || !hasCode(r.Findings, "byte-limit") || !hasCode(r.Findings, "duplicate") {
		t.Fatalf("selected=%d findings=%+v", len(r.Selected), r.Findings)
	}
	if strings.Contains(reportText(r), "super-secret") {
		t.Fatal("secret leaked in findings")
	}
}

func TestCaptureEnvExcludesNonAllowlistedKeys(t *testing.T) {
	environment := map[string]string{"HOME": "/safe/home", "API_TOKEN": "super-secret"}
	snapshot := CaptureEnv([]string{"HOME"}, func(key string) (string, bool) {
		value, ok := environment[key]
		return value, ok
	})
	if snapshot["HOME"] != "/safe/home" {
		t.Fatalf("HOME = %q", snapshot["HOME"])
	}
	if _, ok := snapshot["API_TOKEN"]; ok {
		t.Fatal("non-allowlisted environment key was captured")
	}
}

func TestExecuteEnforcesMaxFiles(t *testing.T) {
	root := t.TempDir()
	sources := []string{filepath.Join(root, "first.json"), filepath.Join(root, "second.json")}
	for _, source := range sources {
		if err := os.WriteFile(source, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	r := Execute(context.Background(), Plan{Roots: []string{root}, Sources: sources, MaxFiles: 1}, OSReadFS{}, EnvSnapshot{})
	if len(r.Selected) != 1 || !hasCode(r.Findings, "file-limit") {
		t.Fatalf("selected=%d findings=%+v", len(r.Selected), r.Findings)
	}
}

func TestExecuteTreatsConfigurationAsOpaqueData(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "malformed.json")
	content := []byte(`{"command":"$(touch should-never-exist)",`)
	if err := os.WriteFile(file, content, 0o600); err != nil {
		t.Fatal(err)
	}
	r := Execute(context.Background(), Plan{Roots: []string{root}, Sources: []string{file}}, OSReadFS{}, EnvSnapshot{})
	if len(r.Selected) != 1 || string(r.Selected[0].Data) != string(content) {
		t.Fatalf("selected=%+v findings=%+v", r.Selected, r.Findings)
	}
	if _, err := os.Stat(filepath.Join(root, "should-never-exist")); !os.IsNotExist(err) {
		t.Fatalf("configuration content was executed: %v", err)
	}
}

func TestExecuteCancellationAndReadOnlySurface(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := Execute(ctx, Plan{Roots: []string{t.TempDir()}, Sources: []string{"ignored"}}, OSReadFS{}, EnvSnapshot{})
	if !hasCode(r.Findings, "cancelled") {
		t.Fatalf("findings=%+v", r.Findings)
	}
	for i := 0; i < reflect.TypeOf((*ReadFS)(nil)).Elem().NumMethod(); i++ {
		name := reflect.TypeOf((*ReadFS)(nil)).Elem().Method(i).Name
		if strings.Contains(name, "Write") || strings.Contains(name, "Exec") || strings.Contains(name, "Network") {
			t.Fatalf("unsafe capability %q", name)
		}
	}
}

func hasCode(findings []Finding, code string) bool {
	for _, finding := range findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}

func reportText(r SelectionReport) string { return strings.Join(r.Messages(), " ") }
