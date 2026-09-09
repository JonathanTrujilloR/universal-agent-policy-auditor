package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionAndHelpAreDeterministicDevIdentity(t *testing.T) {
	for _, args := range [][]string{{"version"}, {"--help"}} {
		firstCode, firstOut, firstErr := runForTest(args...)
		secondCode, secondOut, secondErr := runForTest(args...)
		if firstCode != 0 || firstCode != secondCode || firstOut != secondOut || firstErr != secondErr {
			t.Fatalf("nondeterministic args=%v first=(%d,%q,%q) second=(%d,%q,%q)", args, firstCode, firstOut, firstErr, secondCode, secondOut, secondErr)
		}
		if strings.Contains(firstOut, "v0.1.0-alpha.1") || !strings.Contains(firstOut, "dev") {
			t.Fatalf("development identity not visible or overclaimed: %q", firstOut)
		}
	}
}

func TestAuditCommandGrammarAndExitCategories(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "opencode.json")
	content := []byte(`{"permission":{"bash":"allow"}}`)
	if err := os.WriteFile(config, content, 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(config)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name string
		args []string
		code int
		want string
	}{
		{"valid shell unsupported", []string{"audit", "opencode", "--root", root, "--config", config, "--opencode-version", "1.0.0"}, 2, "unsupported_or_incomplete"},
		{"missing opencode version unsupported", []string{"audit", "opencode", "--root", root, "--config", config}, 2, "unsupported_or_incomplete"},
		{"reordered flags invalid", []string{"audit", "opencode", "--config", config, "--root", root, "--opencode-version", "1"}, 3, "invalid_request"},
		{"relative paths invalid", []string{"audit", "opencode", "--root", ".", "--config", "opencode.json", "--opencode-version", "1"}, 3, "invalid_request"},
		{"delete as config invalid", []string{"audit", "opencode", "--root", root, "--config", "--delete", "--opencode-version", "1"}, 3, "invalid_request"},
		{"help trailing invalid", []string{"--help", "--unknown"}, 3, "invalid_request"},
		{"help command trailing invalid", []string{"help", "fix"}, 3, "invalid_request"},
		{"missing root invalid", []string{"audit", "opencode", "--config", config, "--opencode-version", "1"}, 3, "invalid_request"},
		{"outside root config invalid", []string{"audit", "opencode", "--root", root, "--config", filepath.Join(t.TempDir(), "opencode.json"), "--opencode-version", "1"}, 3, "invalid_request"},
		{"claude visible unsupported", []string{"audit", "claudecode", "--root", root, "--config", config, "--opencode-version", "1"}, 2, "unsupported_or_incomplete"},
		{"mutation command invalid", []string{"fix", "opencode", "--root", root}, 3, "invalid_request"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			code, out, errText := runForTest(tt.args...)
			if code != tt.code || !strings.Contains(out+errText, tt.want) {
				t.Fatalf("code=%d out=%q err=%q want code %d text %q", code, out, errText, tt.code, tt.want)
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
		t.Fatalf("config mutated")
	}
}

func TestWriteErrorsReturnOperationalFailure(t *testing.T) {
	if code := run([]string{"version"}, errWriter{}, &bytes.Buffer{}); code != 4 {
		t.Fatalf("version write error code=%d", code)
	}
	if code := run([]string{"--help"}, errWriter{}, &bytes.Buffer{}); code != 4 {
		t.Fatalf("help write error code=%d", code)
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func runForTest(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}
