package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/redact"
	jsonreport "github.com/jkelevra/universal-agent-policy-auditor/internal/render/json"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/render/text"
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

func TestExactHelpAndVersion(t *testing.T) {
	want := "auditor dev\nUsage:\n  auditor version\n" +
		"  auditor audit <target> --root <abs> --config <abs> [--opencode-version <v>] [--format text|json] [--no-color]\n" +
		"Exit categories:\n  0 complete_no_findings\n  1 complete_with_findings\n  2 unsupported_or_incomplete\n  3 invalid_request\n  4 operational_failure\n"
	for _, args := range [][]string{nil, {"help"}, {"--help"}, {"-h"}, {"version"}} {
		expected := want
		if len(args) == 1 && args[0] == "version" {
			expected = "auditor dev\n"
		}
		code, out, stderr := runForTest(args...)
		if code != 0 || out != expected || stderr != "" {
			t.Fatalf("got (%d, %q, %q)", code, out, stderr)
		}
	}
}

func TestEmptyVersionCompatibility(t *testing.T) {
	root := t.TempDir()
	for _, format := range []string{"", "text", "json"} {
		args := []string{"audit", "opencode", "--root", root, "--config", filepath.Join(root, "missing"), "--opencode-version", ""}
		deps := dependencies{run: app.Run}
		wantCode, wantOut, wantErr := 2, "unsupported_or_incomplete\n", ""
		if format != "" {
			args = append(args, "--format", format)
			deps.run = func(app.Request) app.Result {
				t.Fatal("application called for empty explicit value")
				return app.Result{}
			}
			wantCode, wantOut, wantErr = 3, "", "invalid_request\n"
		}
		var out, stderr bytes.Buffer
		if code := runWith(args, &out, &stderr, deps); code != wantCode || out.String() != wantOut || stderr.String() != wantErr {
			t.Fatalf("format=%q got (%d, %q, %q)", format, code, out.String(), stderr.String())
		}
	}
}

func TestExplicitClaudeFormats(t *testing.T) {
	for _, tt := range []struct{ format, want string }{
		{"text", "Tool: universal-agent-policy-auditor dev\nResult: unsupported_or_incomplete (exit 2)\nTarget: claudecode\nSupport status: unsupported\nCompleteness: incomplete\n"},
		{"json", "{\"schema\":\"auditor-report/v1alpha1\",\"tool\":{\"name\":\"universal-agent-policy-auditor\",\"version\":\"dev\"},\"result\":\"unsupported_or_incomplete\",\"exit_code\":2,\"target\":{\"name\":\"claudecode\",\"requested_version\":\"\",\"support_status\":\"unsupported\"},\"completeness\":\"incomplete\",\"source_digests\":[],\"permissions\":[],\"findings\":[],\"limitations\":[]}\n"},
	} {
		t.Run(tt.format, func(t *testing.T) {
			root := t.TempDir()
			code, out, errText := runForTest("audit", "claudecode", "--root", root, "--config", filepath.Join(root, "missing"), "--format", tt.format)
			if code != 2 || out != tt.want || errText != "" {
				t.Fatalf("got (%d, %q, %q)", code, out, errText)
			}
		})
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

func TestExplicitSyntaxBeforeApplication(t *testing.T) {
	base := "audit opencode --root /private-canary --config /private-canary/config"
	for _, args := range []string{
		base + " --format", base + " --format yaml", base + " --no-color",
		base + " --format json --no-color", base + " --format text --format text",
		base + " --format text --no-color --no-color", base + " --format text --opencode-version 1",
		base + " --opencode-version --format text", base + " --unknown text",
		base + " --opencode-version 1 --opencode-version 1 --format text",
		"audit opencode --config /private-canary/config --root /private-canary --format text",
		"audit unknown --root /private-canary --config /private-canary/config --format text",
		"version --format json", "help --format text", "--help --format text", "-h --format json",
	} {
		t.Run(args, func(t *testing.T) {
			var out, stderr bytes.Buffer
			deps := dependencies{run: func(app.Request) app.Result {
				t.Fatal("application called for invalid syntax")
				return app.Result{}
			}}
			if code := runWith(strings.Fields(args), &out, &stderr, deps); code != 3 || out.Len() != 0 || stderr.String() != "invalid_request\n" {
				t.Fatalf("got (%d, %q, %q)", code, out.String(), stderr.String())
			}
		})
	}
}

func TestExplicitOpenCodeAndBinary(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "private-canary.json")
	content := []byte(`{"permission":{"read":"allow","edit":"deny","bash":"ask"}}`)
	if err := os.WriteFile(config, content, 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(config)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "auditor")
	if !testing.Short() {
		if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
			t.Fatalf("build: %s: %v", output, err)
		}
	}
	for _, version := range []string{"1.18.27", "", "private-canary"} {
		for _, format := range []string{"text", "json"} {
			args := []string{"audit", "opencode", "--root", root, "--config", config}
			if version != "" {
				args = append(args, "--opencode-version", version)
			}
			result := app.Run(parseAudit(args[1:]))
			wantCode := 2
			if version == "1.18.27" {
				wantCode = 0
			}
			report, err := redact.NewReport(result)
			if err != nil {
				t.Fatal(err)
			}
			render := text.Render
			if format == "json" {
				render = jsonreport.Marshal
			}
			want, err := render(report)
			if err != nil {
				t.Fatal(err)
			}
			args = append(args, "--format", format)
			code, out, stderr := runForTest(args...)
			if code != wantCode || out != string(want) || stderr != "" || strings.Contains(out, "private-canary") || strings.Contains(out, root) || strings.Contains(out, "\x1b") {
				t.Fatalf("unexpected report (%d, %q, %q)", code, out, stderr)
			}
			if format == "text" {
				c, o, e := runForTest(append(args, "--no-color")...)
				if c != code || o != out || e != stderr {
					t.Fatal("no-color changed output")
				}
			}
			if !testing.Short() {
				cmd := exec.Command(binary, args...)
				cmd.Dir = t.TempDir()
				var actual, diagnostic bytes.Buffer
				cmd.Stdout, cmd.Stderr = &actual, &diagnostic
				_ = cmd.Run()
				if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != code || actual.String() != out || diagnostic.Len() != 0 {
					t.Fatal("real CLI differs")
				}
			}
		}
	}
	after, err := os.Stat(config)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(config)
	if err != nil || !bytes.Equal(got, content) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("input mutated")
	}
}

func TestExplicitFailuresAndWrites(t *testing.T) {
	for _, format := range []string{"text", "json"} {
		for _, failure := range []string{"none", "redact", "render", "empty", "reserved"} {
			for _, category := range []app.Category{app.InvalidRequest, app.OperationalFailure} {
				args := []string{"audit", "opencode", "--root", ".", "--config", "private-canary", "--format", format}
				calls := 0
				deps := dependencies{app.Run, redact.NewReport, text.Render, jsonreport.Marshal}
				deps.run = func(req app.Request) app.Result {
					calls++
					r := app.Run(req)
					r.Category, r.Message = category, "private-canary raw message"
					if failure == "reserved" {
						r.Category = app.CompleteWithFindings
					}
					return r
				}
				if failure == "redact" {
					deps.redact = func(app.Result) (redact.Report, error) { return redact.Report{}, errors.New("private-canary") }
				}
				if failure == "render" || failure == "empty" {
					render := func(redact.Report) ([]byte, error) {
						if failure == "empty" {
							return nil, nil
						}
						return []byte("private-canary"), errors.New("raw error")
					}
					deps.text, deps.json = render, render
				}
				var out, stderr bytes.Buffer
				code := runWith(args, &out, &stderr, deps)
				if calls != 1 || strings.Contains(out.String()+stderr.String(), "private-canary") {
					t.Fatal("call count or privacy")
				}
				if failure != "none" {
					if code != 4 || out.Len() != 0 || stderr.String() != "operational_failure\n" {
						t.Fatal("failure leaked output")
					}
				} else if code != app.ExitCode(category) || out.Len() == 0 || stderr.Len() != 0 {
					t.Fatal("safe failure report routing")
				}
				for _, writer := range []io.Writer{errWriter{}, &shortWriter{}} {
					stderr.Reset()
					if runWith(args, writer, &stderr, deps) != 4 {
						t.Fatal("stdout failure exit")
					}
					if failure == "none" && stderr.Len() != 0 {
						t.Fatal("stdout failure diagnostic")
					}
					if failure != "none" && runWith(args, &out, writer, deps) != 4 {
						t.Fatal("stderr failure exit")
					}
				}
			}
		}
	}
	for _, args := range [][]string{{"version"}, {"--help"}, {"invalid"}} {
		if run(args, &shortWriter{}, &shortWriter{}) != 4 {
			t.Fatal("short write accepted")
		}
	}
	writer := &shortWriter{}
	if err := writeExact(writer, []byte("complete report")); err != io.ErrShortWrite || writer.calls != 1 {
		t.Fatal("write must be exact and single-call")
	}
}

type shortWriter struct{ calls int }

func (w *shortWriter) Write(data []byte) (int, error) {
	w.calls++
	return len(data) - 1, nil
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func runForTest(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}
