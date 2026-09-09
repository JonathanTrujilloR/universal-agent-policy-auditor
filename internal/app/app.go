// Package app owns the auditor application request/result boundary.
package app

import (
	"path/filepath"
	"strings"
)

type Mode string
type Target string
type Category string

const (
	ModeVersion Mode = "version"
	ModeAudit   Mode = "audit"

	TargetOpenCode   Target = "opencode"
	TargetClaudeCode Target = "claudecode"

	CompleteNoFindings      Category = "complete_no_findings"
	CompleteWithFindings    Category = "complete_with_findings"
	UnsupportedOrIncomplete Category = "unsupported_or_incomplete"
	InvalidRequest          Category = "invalid_request"
	OperationalFailure      Category = "operational_failure"
)

type BuildInfo struct{ Version string }

type Request struct {
	Mode           Mode
	Target         Target
	Root           string
	ExplicitConfig string
	TargetVersion  string
	Build          BuildInfo
}

type Result struct {
	Category Category
	Message  string
	Build    BuildInfo
}

func Run(req Request) Result {
	switch req.Mode {
	case ModeVersion:
		return Result{Category: CompleteNoFindings, Message: "version", Build: normalize(req.Build)}
	case ModeAudit:
		return audit(req)
	default:
		return Result{Category: InvalidRequest, Message: "invalid request", Build: normalize(req.Build)}
	}
}

func audit(req Request) Result {
	build := normalize(req.Build)
	if req.Target != TargetOpenCode && req.Target != TargetClaudeCode || req.Root == "" || req.ExplicitConfig == "" {
		return Result{Category: InvalidRequest, Message: "invalid audit request", Build: build}
	}
	if !lexicallyWithin(req.Root, req.ExplicitConfig) {
		return Result{Category: InvalidRequest, Message: "config must be within root", Build: build}
	}
	if req.Target == TargetClaudeCode {
		return Result{Category: UnsupportedOrIncomplete, Message: "Claude Code is unsupported in this alpha", Build: build}
	}
	if req.TargetVersion == "" {
		return Result{Category: UnsupportedOrIncomplete, Message: "OpenCode version evidence is missing", Build: build}
	}
	return Result{Category: UnsupportedOrIncomplete, Message: "OpenCode support evidence is incomplete", Build: build}
}

func ExitCode(category Category) int {
	switch category {
	case CompleteNoFindings:
		return 0
	case CompleteWithFindings:
		return 1
	case UnsupportedOrIncomplete:
		return 2
	case InvalidRequest:
		return 3
	case OperationalFailure:
		return 4
	default:
		return 4
	}
}

func normalize(build BuildInfo) BuildInfo {
	if build.Version == "" {
		build.Version = "dev"
	}
	return build
}

func lexicallyWithin(root, config string) bool {
	root = filepath.Clean(root)
	config = filepath.Clean(config)
	if !filepath.IsAbs(root) || !filepath.IsAbs(config) {
		return false
	}
	rel, err := filepath.Rel(root, config)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
