// Package source executes declarative discovery plans without mutation.
package source

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	defaultMaxFiles = 128
	defaultMaxBytes = 1 << 20
)

// ReadFS is intentionally limited to the filesystem operations needed to read sources.
type ReadFS interface {
	ReadFile(string) ([]byte, error)
	Stat(string) (fs.FileInfo, error)
	EvalSymlinks(string) (string, error)
}

// OSReadFS is the production read-only filesystem adapter.
type OSReadFS struct{}

func (OSReadFS) ReadFile(name string) ([]byte, error)     { return os.ReadFile(name) }
func (OSReadFS) Stat(name string) (fs.FileInfo, error)    { return os.Stat(name) }
func (OSReadFS) EvalSymlinks(name string) (string, error) { return filepath.EvalSymlinks(name) }

// EnvSnapshot contains only values captured from an adapter-declared allowlist.
type EnvSnapshot map[string]string

// CaptureEnv copies allowlisted values at the boundary; it never reads global state itself.
func CaptureEnv(keys []string, lookup func(string) (string, bool)) EnvSnapshot {
	values := make(EnvSnapshot, len(keys))
	for _, key := range keys {
		if value, ok := lookup(key); ok {
			values[key] = value
		}
	}
	return values
}

// Plan is declarative input for bounded source selection.
type Plan struct {
	Roots    []string
	Sources  []string
	EnvKeys  []string
	MaxFiles int
	MaxBytes int
}

type SelectedSource struct {
	Identity string
	Path     string
	Data     []byte
}

type Finding struct {
	Code    string
	Source  string
	Message string
}

type SelectionReport struct {
	Selected []SelectedSource
	Findings []Finding
}

func (r SelectionReport) Messages() []string {
	messages := make([]string, len(r.Findings))
	for i, finding := range r.Findings {
		messages[i] = finding.Message
	}
	return messages
}

// Execute selects regular files within canonical roots. Content remains opaque bytes.
func Execute(ctx context.Context, plan Plan, files ReadFS, _ EnvSnapshot) SelectionReport {
	report := SelectionReport{}
	roots := canonicalRoots(plan.Roots, files, &report)
	maxFiles, maxBytes := plan.MaxFiles, plan.MaxBytes
	if maxFiles <= 0 {
		maxFiles = defaultMaxFiles
	}
	if maxBytes <= 0 {
		maxBytes = defaultMaxBytes
	}
	seen := map[string]bool{}
	usedBytes := 0
	for _, candidate := range plan.Sources {
		if ctx.Err() != nil {
			report.add("cancelled", "", "source selection cancelled")
			break
		}
		if !withinAny(filepath.Clean(candidate), roots) {
			report.add("outside-root", candidate, "source is outside an allowed root")
			continue
		}
		canonical, err := files.EvalSymlinks(candidate)
		if err != nil {
			report.add("unreadable", candidate, "source cannot be resolved safely")
			continue
		}
		if !withinAny(canonical, roots) {
			report.add("outside-root", candidate, "resolved source is outside an allowed root")
			continue
		}
		if seen[canonical] {
			report.add("duplicate", candidate, "duplicate source skipped")
			continue
		}
		seen[canonical] = true
		info, err := files.Stat(canonical)
		if err != nil || !info.Mode().IsRegular() {
			report.add("unreadable", candidate, "source is not a readable regular file")
			continue
		}
		if len(report.Selected) >= maxFiles {
			report.add("file-limit", candidate, "source file limit reached")
			continue
		}
		data, err := files.ReadFile(canonical)
		if err != nil {
			report.add("unreadable", candidate, "source cannot be read")
			continue
		}
		if len(data) > maxBytes-usedBytes {
			report.add("byte-limit", candidate, "source byte limit reached")
			continue
		}
		usedBytes += len(data)
		digest := sha256.Sum256([]byte(canonical))
		report.Selected = append(report.Selected, SelectedSource{Identity: hex.EncodeToString(digest[:]), Path: canonical, Data: append([]byte(nil), data...)})
	}
	sort.Slice(report.Selected, func(i, j int) bool { return report.Selected[i].Identity < report.Selected[j].Identity })
	return report
}

func canonicalRoots(roots []string, files ReadFS, report *SelectionReport) []string {
	result := make([]string, 0, len(roots))
	for _, root := range roots {
		canonical, err := files.EvalSymlinks(root)
		if err != nil {
			report.add("invalid-root", root, "allowed root cannot be resolved")
			continue
		}
		result = append(result, filepath.Clean(canonical))
	}
	return result
}

func withinAny(path string, roots []string) bool {
	for _, root := range roots {
		rel, err := filepath.Rel(root, path)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return true
		}
	}
	return false
}

func (r *SelectionReport) add(code, path, message string) {
	r.Findings = append(r.Findings, Finding{Code: code, Source: filepath.Base(path), Message: message})
}
