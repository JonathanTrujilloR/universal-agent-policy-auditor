package docs

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAlphaReleaseClaimBoundaries(t *testing.T) {
	files := map[string][]string{
		"README.md": {
			"Alpha candidate planned: `v0.1.0-alpha.1`",
			"GitHub Releases are the canonical source of actual availability.",
			"OpenCode `1.18.27` legacy scalar `permission`",
			"Claude Code | Unsupported",
		},
		"docs/installation.md": {
			"Linux amd64 only", "checksums.txt", "not signatures",
			"--opencode-version 1.18.27", "--reference-config", "Apache-2.0",
		},
		"docs/release-notes/v0.1.0-alpha.1.md": {
			"Release notes draft", "static configuration analysis", "Post-release pilot",
			"scripts/test_release.py", "support/evidence.md",
		},
	}
	for name, want := range files {
		data, err := os.ReadFile(filepath.Join("..", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		text := string(data)
		for _, phrase := range want {
			if !strings.Contains(text, phrase) {
				t.Errorf("%s missing %q", name, phrase)
			}
		}
	}
}

func TestAlphaReleaseCandidateLinksAndCommands(t *testing.T) {
	data, err := os.ReadFile("installation.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		"releases/download/v0.1.0-alpha.1/auditor_v0.1.0-alpha.1_linux_amd64",
		"releases/download/v0.1.0-alpha.1/checksums.txt",
		"releases/download/v0.1.0-alpha.1/LICENSE",
		"auditor\" audit opencode --root", "auditor\" compare opencode --root",
		"--format text --no-color", "No `sudo` is needed.",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("installation.md missing %q", want)
		}
	}
}

func TestAlphaReleaseInstallSnippetFailsClosed(t *testing.T) {
	for _, test := range []struct {
		name, failAsset string
		corrupt         bool
		wantDownloads   string
	}{
		{name: "fresh-home", wantDownloads: "auditor_v0.1.0-alpha.1_linux_amd64\nchecksums.txt\nLICENSE"},
		{name: "checksum-failure", corrupt: true, wantDownloads: "auditor_v0.1.0-alpha.1_linux_amd64\nchecksums.txt\nLICENSE"},
		{name: "download-failure", failAsset: "checksums.txt", wantDownloads: "auditor_v0.1.0-alpha.1_linux_amd64\nchecksums.txt"},
	} {
		t.Run(test.name, func(t *testing.T) {
			home, fixture, fakebin := t.TempDir(), t.TempDir(), t.TempDir()
			binary := []byte("#!/usr/bin/env bash\nprintf '%s\\n' \"$1\" >> \"$SENTINEL\"\n")
			writeInstallFixture(t, fixture, "auditor_v0.1.0-alpha.1_linux_amd64", binary)
			checksum := fmt.Sprintf("%x  auditor_v0.1.0-alpha.1_linux_amd64\n", sha256.Sum256(binary))
			if test.corrupt {
				checksum = strings.Repeat("0", 64) + "  auditor_v0.1.0-alpha.1_linux_amd64\n"
			}
			writeInstallFixture(t, fixture, "checksums.txt", []byte(checksum))
			writeInstallFixture(t, fixture, "LICENSE", []byte("fixture license\n"))
			writeFakeCurl(t, fakebin)
			sentinel, downloads := filepath.Join(home, "executed"), filepath.Join(home, "downloads")
			cmd := exec.Command("bash", "-s")
			cmd.Dir = t.TempDir()
			cmd.Stdin = strings.NewReader(documentedBashBlocks(t))
			cmd.Env = installTestEnv("HOME="+home, "PATH="+fakebin+":"+os.Getenv("PATH"), "TMPDIR="+t.TempDir(), "FIXTURE="+fixture, "SENTINEL="+sentinel, "DOWNLOAD_LOG="+downloads, "FAIL_ASSET="+test.failAsset)
			err := cmd.Run()
			installed, globErr := filepath.Glob(filepath.Join(home, ".local", "opt", "auditor-alpha.*", "auditor"))
			if globErr != nil {
				t.Fatal(globErr)
			}
			downloaded, readErr := os.ReadFile(downloads)
			if readErr != nil || strings.TrimSpace(string(downloaded)) != test.wantDownloads {
				t.Fatalf("downloads %q: %v", downloaded, readErr)
			}
			if test.name == "fresh-home" {
				if err != nil || len(installed) != 1 {
					t.Fatalf("success err=%v installed=%v", err, installed)
				}
				data, readErr := os.ReadFile(sentinel)
				if readErr != nil || string(data) != "version\n--help\naudit\ncompare\n" {
					t.Fatalf("executions %q: %v", data, readErr)
				}
			} else if err == nil || len(installed) != 0 {
				t.Fatalf("failure err=%v installed=%v", err, installed)
			} else if _, statErr := os.Stat(sentinel); !os.IsNotExist(statErr) {
				t.Fatalf("binary executed after failure: %v", statErr)
			}
		})
	}
}

func TestAlphaReleaseDocumentedJourneyUsesPackagedBinary(t *testing.T) {
	fixture, home, fakebin := t.TempDir(), t.TempDir(), t.TempDir()
	script, err := filepath.Abs("../scripts/release.sh")
	if err != nil {
		t.Fatal(err)
	}
	packageCommand := exec.Command("bash", script, "v0.1.0-alpha.1", fixture)
	packageCommand.Dir = t.TempDir()
	packageCommand.Env = append(os.Environ(), "GOTOOLCHAIN=local", "TMPDIR="+t.TempDir())
	if output, err := packageCommand.CombinedOutput(); err != nil {
		t.Fatalf("package fixture: %v: %s", err, output)
	}
	writeFakeCurl(t, fakebin)
	command := exec.Command("bash", "-s")
	command.Dir = t.TempDir()
	command.Stdin = strings.NewReader(documentedBashBlocks(t))
	command.Env = installTestEnv("HOME="+home, "PATH="+fakebin+":"+os.Getenv("PATH"), "TMPDIR="+t.TempDir(), "FIXTURE="+fixture, "DOWNLOAD_LOG="+filepath.Join(home, "downloads"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("documented journey: %v: %s", err, output)
	}
	for _, want := range []string{"auditor v0.1.0-alpha.1", "Usage:", `"result":"complete_no_findings"`, "Comparison status: equivalent"} {
		if !strings.Contains(string(output), want) {
			t.Errorf("journey output missing %q: %s", want, output)
		}
	}
}

func writeFakeCurl(t *testing.T, directory string) {
	t.Helper()
	writeInstallFixture(t, directory, "curl", []byte("#!/usr/bin/env bash\nset -eu\nasset=${!#}\nasset=${asset##*/}\ncase \"$asset\" in auditor_v0.1.0-alpha.1_linux_amd64|checksums.txt|LICENSE) ;; *) exit 64 ;; esac\nprintf '%s\\n' \"$asset\" >> \"$DOWNLOAD_LOG\"\n[ \"$asset\" != \"${FAIL_ASSET:-}\" ] || exit 22\ncp \"$FIXTURE/$asset\" \"$asset\"\n"))
	if err := os.Chmod(filepath.Join(directory, "curl"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func installTestEnv(extra ...string) []string {
	env := make([]string, 0, len(os.Environ())+len(extra))
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "HOME=") && !strings.HasPrefix(entry, "PATH=") {
			env = append(env, entry)
		}
	}
	return append(env, extra...)
}

func documentedBashBlocks(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("installation.md")
	if err != nil {
		t.Fatal(err)
	}
	var blocks []string
	for text := string(data); ; {
		start := strings.Index(text, "```bash\n")
		if start < 0 {
			break
		}
		text = text[start+7:]
		end := strings.Index(text, "```")
		if end < 0 {
			t.Fatal("unterminated bash block")
		}
		blocks = append(blocks, text[:end])
		text = text[end+3:]
	}
	if len(blocks) == 0 {
		t.Fatal("installation bash blocks missing")
	}
	return strings.Join(blocks, "\n")
}

func writeInstallFixture(t *testing.T, directory, name string, data []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, name), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestAlphaReleaseDocsRejectUnevidencedClaims(t *testing.T) {
	for _, name := range []string{"README.md", "docs/installation.md", "docs/release-notes/v0.1.0-alpha.1.md"} {
		data, err := os.ReadFile(filepath.Join("..", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, forbidden := range []string{
			"Production-ready", "compliance certified", "enforces runtime policy",
			"macOS binary", "Windows binary", "Install with a package manager",
			"Signed artifacts", "provenance attestation is included", "Claude Code is supported",
		} {
			if strings.Contains(string(data), forbidden) {
				t.Errorf("%s contains unevidenced claim %q", name, forbidden)
			}
		}
	}
}
