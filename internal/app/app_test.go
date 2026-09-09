package app

import (
	"os"
	"path/filepath"
	"testing"
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
