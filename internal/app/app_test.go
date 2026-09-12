package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/compare"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/source"
	"github.com/jkelevra/universal-agent-policy-auditor/support"
	"reflect"
)

type comparisonFS struct {
	source.OSReadFS
	reads []string
}

func (f *comparisonFS) ReadFile(path string) ([]byte, error) {
	f.reads = append(f.reads, path)
	return f.OSReadFS.ReadFile(path)
}

func TestCompareIntegration(t *testing.T) {
	good := []byte(`{"permission":{"read":"allow","edit":"deny","bash":"ask"}}`)
	for _, variant := range []string{"equivalent", "same", "effect", "malformed-reference", "incomplete-target", "read-failure"} {
		t.Run(variant, func(t *testing.T) {
			root, a := writeConfig(t, good)
			b := filepath.Join(root, "target.json")
			body := good
			if variant == "effect" {
				body = []byte(`{"permission":{"read":"deny","edit":"deny","bash":"ask"}}`)
			}
			if variant == "incomplete-target" {
				body = []byte(`{"permission":{"read":"allow"}}`)
			}
			must(t, os.WriteFile(b, body, 0600))
			if variant == "malformed-reference" {
				must(t, os.WriteFile(a, []byte(`{`), 0600))
			}
			if variant == "same" {
				b = a
			}
			beforeA, beforeB := snapshot(t, a), snapshot(t, b)
			req := Request{Mode: ModeCompare, Target: TargetOpenCode, Root: root, ReferenceConfig: a, TargetConfig: b, TargetVersion: "1.18.27"}
			if variant == "read-failure" {
				req.ReferenceConfig = a + ".missing"
			}
			files, calls := &comparisonFS{}, 0
			result := runWithOptions(req, options{files: files, comparator: func(ref, target compare.Operand) compare.Result {
				calls++
				for _, operand := range []compare.Operand{ref, target} {
					if operand.Canonicalization != "opencode-1.18.27-legacy-permission-scalar" || operand.Coverage != "opencode-1.18.27/read-edit-bash-scalar" {
						t.Fatalf("operand=%+v", operand)
					}
				}
				if (variant == "malformed-reference" || variant == "read-failure") && (ref.Completeness != model.CompletenessIncomplete || target.Completeness != model.CompletenessComplete) {
					t.Fatalf("operands=%+v %+v", ref, target)
				}
				if variant == "incomplete-target" && (ref.Completeness != model.CompletenessComplete || target.Completeness != model.CompletenessIncomplete) {
					t.Fatalf("operands=%+v %+v", ref, target)
				}
				return compare.Compare(ref, target)
			}})
			want, outcome := UnsupportedOrIncomplete, compare.NotComparable
			if variant == "equivalent" || variant == "same" {
				want, outcome = CompleteNoFindings, compare.Equivalent
			}
			if variant == "read-failure" {
				want = OperationalFailure
			}
			if calls != 1 || result.Category != want || result.Comparison == nil || result.Comparison.Outcome != outcome || result.Comparison.Reasons == nil || result.Comparison.Differences == nil || result.SupportStatus != SupportSupported || (result.Completeness == model.CompletenessComplete) != (want == CompleteNoFindings) {
				t.Fatalf("calls=%d result=%+v", calls, result)
			}
			expectedReads := []string{a, b}
			expectedSources := []ComparisonSource{{compare.Reference, "sha256:" + sha256Hex(a)}, {compare.Target, "sha256:" + sha256Hex(b)}}
			if variant == "read-failure" {
				expectedReads, expectedSources = expectedReads[1:], expectedSources[1:]
			}
			if !reflect.DeepEqual(files.reads, expectedReads) || !reflect.DeepEqual(result.ComparisonSources, expectedSources) {
				t.Fatalf("reads=%v sources=%v", files.reads, result.ComparisonSources)
			}
			again := Run(req)
			if again.Category != result.Category || again.Comparison == nil || again.Comparison.Outcome != result.Comparison.Outcome || !reflect.DeepEqual(again.ComparisonSources, result.ComparisonSources) {
				t.Fatal("unstable production comparison")
			}
			result.ComparisonSources[0].Digest = "changed"
			if reflect.DeepEqual(again.ComparisonSources, result.ComparisonSources) {
				t.Fatal("aliased sources")
			}
			assertUnchanged(t, a, beforeA.data, beforeA)
			assertUnchanged(t, b, beforeB.data, beforeB)
		})
	}
}

func TestCompareAdmission(t *testing.T) {
	req := Request{Mode: ModeCompare, Target: TargetOpenCode, Root: "/root", ReferenceConfig: "/root/a", TargetConfig: "/root/b", TargetVersion: "1.18.27"}
	for _, tc := range []struct {
		name   string
		change func(*Request)
		want   Category
		loads  int
	}{
		{"unknown", func(r *Request) { r.Target = "unknown" }, InvalidRequest, 0},
		{"root", func(r *Request) { r.Root = "" }, InvalidRequest, 0},
		{"relative-root", func(r *Request) { r.Root = "root" }, InvalidRequest, 0},
		{"reference", func(r *Request) { r.ReferenceConfig = "" }, InvalidRequest, 0},
		{"target", func(r *Request) { r.TargetConfig = "" }, InvalidRequest, 0},
		{"relative-reference", func(r *Request) { r.ReferenceConfig = "a" }, InvalidRequest, 0},
		{"relative-target", func(r *Request) { r.TargetConfig = "b" }, InvalidRequest, 0},
		{"outside-reference", func(r *Request) { r.ReferenceConfig = "/root/../a" }, InvalidRequest, 0},
		{"outside-target", func(r *Request) { r.TargetConfig = "/else/b" }, InvalidRequest, 0},
		{"explicit", func(r *Request) { r.ExplicitConfig = "/root/a" }, InvalidRequest, 0},
		{"missing-version", func(r *Request) { r.TargetVersion = "" }, UnsupportedOrIncomplete, 0},
		{"nearby-version", func(r *Request) { r.TargetVersion = "1.18.28" }, UnsupportedOrIncomplete, 1},
		{"claude", func(r *Request) { r.Target = TargetClaudeCode }, UnsupportedOrIncomplete, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := req
			tc.change(&r)
			files, loads := &fakeFS{}, 0
			got := runWithOptions(r, options{files: files, loadSupport: func() (support.Matrix, error) { loads++; return support.LoadCheckedIn() }, comparator: func(compare.Operand, compare.Operand) compare.Result {
				t.Fatal("unexpected comparison")
				return compare.Result{}
			}})
			if got.Category != tc.want || files.calls != 0 || loads != tc.loads || got.Comparison != nil {
				t.Fatalf("result=%+v reads=%d loads=%d", got, files.calls, loads)
			}
		})
	}
	for _, loaderFailure := range []bool{false, true} {
		deps := options{files: (*fakeFS)(nil), comparator: func(compare.Operand, compare.Operand) compare.Result {
			t.Fatal("unexpected comparison")
			return compare.Result{}
		}}
		if loaderFailure {
			deps.loadSupport = func() (support.Matrix, error) { return support.Matrix{}, errors.New("private path") }
		}
		if got := runWithOptions(req, deps); got.Category != OperationalFailure {
			t.Fatalf("result=%+v", got)
		}
	}
	files := &fakeFS{root: req.Root, path: req.ReferenceConfig, readErr: true}
	calls := 0
	got := runWithOptions(req, options{files: files, comparator: func(a, b compare.Operand) compare.Result { calls++; return compare.Compare(a, b) }})
	if got.Category != OperationalFailure || calls != 1 {
		t.Fatalf("result=%+v calls=%d", got, calls)
	}
}

func TestCompareOutcomesAndDefensiveSlices(t *testing.T) {
	root, path := writeConfig(t, []byte(`{"permission":{"read":"allow","edit":"deny","bash":"ask"}}`))
	for _, outcome := range []compare.Outcome{compare.TargetOnly, compare.Ambiguous, compare.NotComparable, "unknown"} {
		injected := compare.Result{Outcome: outcome, Reasons: []compare.Reason{{Side: compare.Target, Code: compare.EffectDiffers}}, Differences: []compare.Difference{{Outcome: outcome}}}
		calls := 0
		got := runWithOptions(Request{Mode: ModeCompare, Target: TargetOpenCode, Root: root, ReferenceConfig: path, TargetConfig: path, TargetVersion: "1.18.27"}, options{comparator: func(compare.Operand, compare.Operand) compare.Result { calls++; return injected }})
		if calls != 1 || ExitCode(got.Category) != 2 || got.Completeness != model.CompletenessIncomplete {
			t.Fatalf("result=%+v calls=%d", got, calls)
		}
		got.Comparison.Reasons[0].Side = compare.Reference
		got.Comparison.Differences[0].Outcome = compare.Equivalent
		if injected.Reasons[0].Side != compare.Target || injected.Differences[0].Outcome != outcome {
			t.Fatal("comparator output aliased")
		}
	}
}

func TestCompareAPI(t *testing.T) {
	result := Run(Request{Mode: ModeCompare, ReferenceConfig: "/root/a", TargetConfig: "/root/b"})
	if result.Category != InvalidRequest || result.Comparison != nil || len(result.ComparisonSources) != 0 {
		t.Fatalf("result=%+v", result)
	}
}

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

func TestAuditBranchMetadataIsClosedAndConsistent(t *testing.T) {
	valid := []byte(`{"permission":{"read":"allow","edit":"deny","bash":"ask"}}`)
	malformed := []byte(`{`)
	root, config := writeConfig(t, valid)
	missingRoot := filepath.Join(t.TempDir(), "missing")

	for _, tt := range []struct {
		name           string
		req            Request
		body           []byte
		options        options
		wantCategory   Category
		wantSupport    SupportStatus
		wantComplete   model.Completeness
		wantLimit      bool
		wantDigestPath string
	}{
		{"success", Request{Mode: ModeAudit, Target: TargetOpenCode, Root: root, ExplicitConfig: config, TargetVersion: "1.18.27"}, nil, options{}, CompleteNoFindings, SupportSupported, model.CompletenessComplete, true, config},
		{"malformed admitted input", Request{Mode: ModeAudit, Target: TargetOpenCode, TargetVersion: "1.18.27"}, malformed, options{}, UnsupportedOrIncomplete, SupportSupported, model.CompletenessIncomplete, true, ""},
		{"source error after admitted version", Request{Mode: ModeAudit, Target: TargetOpenCode, Root: missingRoot, ExplicitConfig: filepath.Join(missingRoot, "opencode.json"), TargetVersion: "1.18.27"}, nil, options{}, UnsupportedOrIncomplete, SupportSupported, model.CompletenessIncomplete, true, ""},
		{"missing version", Request{Mode: ModeAudit, Target: TargetOpenCode, Root: root, ExplicitConfig: config}, nil, options{}, UnsupportedOrIncomplete, SupportUnresolved, model.CompletenessIncomplete, true, ""},
		{"nearby version", Request{Mode: ModeAudit, Target: TargetOpenCode, Root: root, ExplicitConfig: config, TargetVersion: "1.18.28"}, nil, options{}, UnsupportedOrIncomplete, SupportUnsupported, model.CompletenessIncomplete, true, ""},
		{"claude", Request{Mode: ModeAudit, Target: TargetClaudeCode, Root: root, ExplicitConfig: config, TargetVersion: "1"}, nil, options{}, UnsupportedOrIncomplete, SupportUnsupported, model.CompletenessIncomplete, false, ""},
		{"invalid path", Request{Mode: ModeAudit, Target: TargetOpenCode, Root: ".", ExplicitConfig: "opencode.json", TargetVersion: "1.18.27"}, nil, options{}, InvalidRequest, SupportUnresolved, model.CompletenessIncomplete, true, ""},
		{"metadata load error", Request{Mode: ModeAudit, Target: TargetOpenCode, Root: root, ExplicitConfig: config, TargetVersion: "1.18.27"}, nil, options{loadSupport: func() (support.Matrix, error) { return support.Matrix{}, errors.New("boom") }}, OperationalFailure, SupportUnresolved, model.CompletenessIncomplete, true, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.req
			if tt.body != nil {
				req.Root, req.ExplicitConfig = writeConfig(t, tt.body)
			}
			got := runWithOptions(req, tt.options)
			assertAuditMetadata(t, got, req.Target, req.TargetVersion, tt.wantCategory, tt.wantSupport, tt.wantComplete)
			if tt.wantLimit != hasString(got.Limitations, "modeled-static-configuration") {
				t.Fatalf("limitations=%v want modeled-static-configuration present=%v", got.Limitations, tt.wantLimit)
			}
			if got.SourceDigests == nil {
				t.Fatalf("SourceDigests is nil")
			}
			if tt.wantDigestPath != "" {
				wantDigest := "sha256:" + sha256Hex(tt.wantDigestPath)
				if len(got.SourceDigests) != 1 || got.SourceDigests[0] != wantDigest {
					t.Fatalf("SourceDigests=%v want [%s]", got.SourceDigests, wantDigest)
				}
				if got.SourceDigests[0] == tt.wantDigestPath || got.SourceDigests[0] == string(valid) {
					t.Fatalf("source digest leaked raw source identity material: %q", got.SourceDigests[0])
				}
			}
		})
	}
	unknown := Run(Request{Mode: ModeAudit, Target: Target("private-target"), Root: root, ExplicitConfig: config, TargetVersion: "1"})
	if unknown.Category != InvalidRequest || unknown.Target != "" {
		t.Fatalf("unknown target metadata=%+v", unknown)
	}
}

func TestSourceDigestsAreDeterministicDeduplicatedAndNonAliased(t *testing.T) {
	a := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	b := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	z := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	selected := []source.SelectedSource{
		{Identity: b, Path: "/private/b", Data: []byte("secret-b")},
		{Identity: a, Path: "/private/a", Data: []byte("secret-a")},
		{Identity: b, Path: "/private/duplicate", Data: []byte("secret-duplicate")},
	}
	got := sourceDigests(selected)
	want := []string{"sha256:" + a, "sha256:" + b}
	if !sameStrings(got, want) {
		t.Fatalf("sourceDigests=%v want %v", got, want)
	}
	selected[1].Identity = z
	got[0] = "mutated"
	again := sourceDigests(selected)
	if !sameStrings(again, []string{"sha256:" + b, "sha256:" + z}) {
		t.Fatalf("sourceDigests after mutation=%v", again)
	}
	empty := sourceDigests([]source.SelectedSource{{Identity: "not-a-digest"}})
	if empty == nil || len(empty) != 0 {
		t.Fatalf("invalid sourceDigests=%v", empty)
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

func assertAuditMetadata(t *testing.T, got Result, target Target, requested string, category Category, support SupportStatus, completeness model.Completeness) {
	t.Helper()
	if got.Category != category || got.Target != target || got.RequestedTargetVersion != requested || got.SupportStatus != support || got.Completeness != completeness {
		t.Fatalf("metadata got category=%q target=%q requested=%q support=%q completeness=%q; want category=%q target=%q requested=%q support=%q completeness=%q", got.Category, got.Target, got.RequestedTargetVersion, got.SupportStatus, got.Completeness, category, target, requested, support, completeness)
	}
}

func sha256Hex(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
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
