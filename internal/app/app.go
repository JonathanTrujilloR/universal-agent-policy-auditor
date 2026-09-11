// Package app owns the auditor application request/result boundary.
package app

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/adapter/opencode"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/source"
	"github.com/jkelevra/universal-agent-policy-auditor/support"
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
	Category     Category
	Message      string
	Build        BuildInfo
	Completeness model.Completeness
	Permissions  []model.Permission
	Findings     []model.Finding
	Limitations  []string
}

type options struct {
	files       source.ReadFS
	loadSupport func() (support.Matrix, error)
	lookupEnv   func(string) (string, bool)
}

func Run(req Request) Result { return runWithOptions(req, options{}) }

func runWithOptions(req Request, options options) Result {
	deps := normalizeOptions(options)
	switch req.Mode {
	case ModeVersion:
		return Result{Category: CompleteNoFindings, Message: "version", Build: normalize(req.Build), Completeness: model.CompletenessComplete}
	case ModeAudit:
		return audit(req, deps)
	default:
		return Result{Category: InvalidRequest, Message: "invalid request", Build: normalize(req.Build)}
	}
}

func audit(req Request, options options) Result {
	build := normalize(req.Build)
	files := reflect.ValueOf(options.files)
	if files.Kind() == reflect.Pointer && files.IsNil() {
		return Result{Category: OperationalFailure, Message: "source dependency unavailable", Build: build}
	}
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
	matrix, err := options.loadSupport()
	if err != nil {
		return Result{Category: OperationalFailure, Message: "support metadata unavailable", Build: build}
	}
	if result := support.Validate(matrix, support.Entry{Target: "opencode", Version: req.TargetVersion, Construct: "permission"}); result.Completeness != model.CompletenessComplete {
		return Result{Category: UnsupportedOrIncomplete, Message: "OpenCode support evidence is incomplete", Build: build, Completeness: result.Completeness, Findings: result.Findings}
	}

	plan := opencode.DiscoveryPlan(req.Root, []string{req.ExplicitConfig})
	selection := source.Execute(context.Background(), plan, options.files, source.CaptureEnv(plan.EnvKeys, options.lookupEnv))
	if len(selection.Findings) != 0 || len(selection.Selected) != 1 {
		return Result{Category: UnsupportedOrIncomplete, Message: "OpenCode source selection is incomplete", Build: build, Completeness: model.CompletenessIncomplete, Findings: sourceFindings(selection.Findings)}
	}
	report := opencode.ResolveWithOptions(selection.Selected, opencode.Options{Version: req.TargetVersion, Support: matrix})
	result := Result{Build: build, Completeness: report.Completeness, Permissions: report.Permissions, Findings: report.Findings, Limitations: []string{"modeled-static-configuration"}}
	if report.Completeness == model.CompletenessComplete && len(report.Findings) == 0 {
		result.Category = CompleteNoFindings
		return result
	}
	result.Category = UnsupportedOrIncomplete
	result.Message = "OpenCode audit is unsupported or incomplete"
	return result
}

func sourceFindings(findings []source.Finding) []model.Finding {
	out := make([]model.Finding, 0, len(findings))
	for _, finding := range findings {
		out = append(out, model.NewFinding("source-"+finding.Code, finding.Message))
	}
	return out
}

func normalizeOptions(options options) options {
	if options.files == nil {
		options.files = source.OSReadFS{}
	}
	if options.loadSupport == nil {
		options.loadSupport = support.LoadCheckedIn
	}
	if options.lookupEnv == nil {
		options.lookupEnv = os.LookupEnv
	}
	return options
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
