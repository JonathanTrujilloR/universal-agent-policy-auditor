package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JonathanTrujilloR/universal-agent-policy-auditor/internal/app"
	"github.com/JonathanTrujilloR/universal-agent-policy-auditor/internal/redact"
	jsonreport "github.com/JonathanTrujilloR/universal-agent-policy-auditor/internal/render/json"
	"github.com/JonathanTrujilloR/universal-agent-policy-auditor/internal/render/text"
)

func comparisonArgs(root, reference, target string) []string {
	return []string{"compare", "opencode", "--root", root, "--reference-config", reference, "--target-config", target, "--opencode-version", "1.18.27"}
}

func TestComparisonCLI(t *testing.T) {
	root := t.TempDir()
	reference := filepath.Join(root, "private-canary-reference")
	target := filepath.Join(root, "private-canary-target")
	content := []byte(`{"permission":{"read":"allow","edit":"deny","bash":"ask"}}`)
	for _, path := range []string{reference, target} {
		if err := os.WriteFile(path, content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	before, err := os.Stat(reference)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "auditor")
	if !testing.Short() {
		if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
			t.Fatalf("build: %s: %v", output, err)
		}
	}
	for _, tc := range []struct {
		name, target, version string
		code                  int
	}{
		{"equivalent", target, "1.18.27", 0}, {"same path", reference, "1.18.27", 0},
		{"unsupported", target, "private-canary", 2}, {"missing", filepath.Join(root, "missing"), "1.18.27", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, format := range []string{"", "text", "json"} {
				args := comparisonArgs(root, reference, tc.target)
				args[9] = tc.version
				result := app.Run(app.Request{Mode: app.ModeCompare, Target: app.TargetOpenCode, Root: root, ReferenceConfig: reference, TargetConfig: tc.target, TargetVersion: tc.version, Build: buildInfo()})
				want := string(result.Category) + "\n"
				wantErr := ""
				if format == "" && tc.code == 4 {
					want, wantErr = "", want
				}
				if format != "" {
					report, err := redact.NewComparisonReport(result)
					if err != nil {
						t.Fatal(err)
					}
					render := text.RenderComparison
					if format == "json" {
						render = jsonreport.MarshalComparison
					}
					data, err := render(report)
					if err != nil {
						t.Fatal(err)
					}
					want = string(data)
					args = append(args, "--format", format)
				}
				for i := 0; i < 100; i++ {
					code, out, diagnostic := runForTest(args...)
					if code != tc.code || out != want || diagnostic != wantErr || strings.Contains(out, "private-canary") || strings.Contains(out, root) {
						t.Fatalf("got (%d,%q,%q)", code, out, diagnostic)
					}
				}
				if format == "text" {
					code, out, diagnostic := runForTest(append(args, "--no-color")...)
					if code != tc.code || out != want || diagnostic != "" {
						t.Fatal("no-color changed output")
					}
				}
				if !testing.Short() {
					cmd := exec.Command(binary, args...)
					cmd.Dir = t.TempDir()
					var out, diagnostic bytes.Buffer
					cmd.Stdout, cmd.Stderr = &out, &diagnostic
					_ = cmd.Run()
					if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != tc.code || out.String() != want || diagnostic.String() != wantErr {
						t.Fatal("binary differs")
					}
				}
			}
		})
	}
	after, err := os.Stat(reference)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{reference, target} {
		data, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(data, content) {
			t.Fatal("input mutated")
		}
	}
	if before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("metadata mutated")
	}
}

func TestComparisonGrammarBeforeApplication(t *testing.T) {
	base := comparisonArgs("/root", "/root/reference", "/root/target")
	cases := [][]string{nil, {"opencode"}, append(append([]string{}, base[1:]...), "--no-color"), append(append([]string{}, base[1:]...), "--format", "json", "--no-color")}
	for i := 0; i < len(base)-1; i++ {
		args := append([]string{}, base[1:]...)
		args[i] = ""
		cases = append(cases, args)
		args = append([]string{}, base[1:]...)
		args[i] = "--unknown"
		cases = append(cases, args)
	}
	for _, suffix := range [][]string{{"--format"}, {"--format", "yaml"}, {"--format", "text", "--format", "text"}, {"--format", "text", "--no-color", "--no-color"}, {"extra"}} {
		cases = append(cases, append(append([]string{}, base[1:]...), suffix...))
	}
	cases = append(cases, []string{"claudecode"}, []string{"opencode", "--config", "/root/target", "--root", "/root"})
	for _, args := range cases {
		var out, diagnostic bytes.Buffer
		deps := dependencies{run: func(app.Request) app.Result { t.Fatal("application called for invalid syntax"); return app.Result{} }}
		if code := runWith(append([]string{"compare"}, args...), &out, &diagnostic, deps); code != 3 || out.Len() != 0 || diagnostic.String() != "invalid_request\n" {
			t.Fatalf("args=%q code=%d", args, code)
		}
	}
}

func TestComparisonFailuresAndSingleWrite(t *testing.T) {
	for _, format := range []string{"text", "json"} {
		for _, failure := range []string{"redaction", "render", "empty", "none"} {
			args := append(comparisonArgs("/root", "/root/reference", "/root/target"), "--format", format)
			deps := dependencies{run: func(app.Request) app.Result { return app.Result{Category: app.CompleteWithFindings} }, comparison: renderComparison}
			if failure != "redaction" {
				deps.comparison = func(app.Result, string) ([]byte, error) {
					if failure == "empty" {
						return nil, nil
					}
					if failure == "render" {
						return []byte("private-canary"), errors.New("private-canary")
					}
					return []byte("safe report\n"), nil
				}
			}
			var out, diagnostic bytes.Buffer
			code := runWith(args, &out, &diagnostic, deps)
			if failure != "none" && (code != 4 || out.Len() != 0 || diagnostic.String() != "operational_failure\n") {
				t.Fatal("unsafe failure routing")
			}
			if failure == "none" {
				if out.String() != "safe report\n" || diagnostic.Len() != 0 {
					t.Fatal("safe output routing")
				}
				for _, category := range []app.Category{app.CompleteNoFindings, app.UnsupportedOrIncomplete, app.InvalidRequest, app.OperationalFailure} {
					deps.run = func(app.Request) app.Result { return app.Result{Category: category} }
					writer := &shortWriter{}
					diagnostic.Reset()
					if runWith(args, writer, &diagnostic, deps) != 4 || writer.calls != 1 || diagnostic.Len() != 0 {
						t.Fatal("short write retried or diagnosed")
					}
					if runWith(args, errWriter{}, &diagnostic, deps) != 4 {
						t.Fatal("write error ignored")
					}
				}
			}
		}
	}
}

func TestComparisonNonEquivalentAndInvalidPaths(t *testing.T) {
	root := t.TempDir()
	reference, target := filepath.Join(root, "reference"), filepath.Join(root, "target")
	for _, tc := range []struct {
		body string
		code int
	}{
		{`{"permission":{"read":"deny","edit":"deny","bash":"ask"}}`, 2},
		{`{"permission":{"read":"private-canary"}}`, 2},
	} {
		if err := os.WriteFile(reference, []byte(`{"permission":{"read":"allow","edit":"deny","bash":"ask"}}`), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(tc.body), 0600); err != nil {
			t.Fatal(err)
		}
		for _, format := range []string{"", "text", "json"} {
			args := comparisonArgs(root, reference, target)
			if format != "" {
				args = append(args, "--format", format)
			}
			code, out, diagnostic := runForTest(args...)
			if code != tc.code || out == "" || diagnostic != "" || strings.Contains(out, "private-canary") {
				t.Fatalf("got (%d,%q,%q)", code, out, diagnostic)
			}
		}
	}
	for _, path := range []string{"relative", filepath.Join(t.TempDir(), "outside")} {
		code, out, diagnostic := runForTest(comparisonArgs(root, reference, path)...)
		if code != 3 || out != "" || diagnostic != "invalid_request\n" {
			t.Fatal("invalid path admitted")
		}
	}
}
