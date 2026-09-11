// Package app owns the auditor application request/result boundary.
package app

import (
	"context"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/adapter/opencode"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/source"
	"github.com/jkelevra/universal-agent-policy-auditor/support"
)

type Mode string
type Target string
type Category string
type SupportStatus string

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

	SupportSupported   SupportStatus = "supported"
	SupportUnsupported SupportStatus = "unsupported"
	SupportUnresolved  SupportStatus = "unresolved"
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
	Category               Category
	Message                string
	Build                  BuildInfo
	Target                 Target
	RequestedTargetVersion string
	SupportStatus          SupportStatus
	SourceDigests          []string
	Completeness           model.Completeness
	Permissions            []model.Permission
	Findings               []model.Finding
	Limitations            []string
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
		return auditResult(req, build, OperationalFailure, "source dependency unavailable", SupportUnresolved)
	}
	if req.Target != TargetOpenCode && req.Target != TargetClaudeCode || req.Root == "" || req.ExplicitConfig == "" {
		return auditResult(req, build, InvalidRequest, "invalid audit request", SupportUnresolved)
	}
	if !lexicallyWithin(req.Root, req.ExplicitConfig) {
		return auditResult(req, build, InvalidRequest, "config must be within root", SupportUnresolved)
	}
	if req.Target == TargetClaudeCode {
		return auditResult(req, build, UnsupportedOrIncomplete, "Claude Code is unsupported in this alpha", SupportUnsupported)
	}
	if req.TargetVersion == "" {
		return auditResult(req, build, UnsupportedOrIncomplete, "OpenCode version evidence is missing", SupportUnresolved)
	}
	matrix, err := options.loadSupport()
	if err != nil {
		return auditResult(req, build, OperationalFailure, "support metadata unavailable", SupportUnresolved)
	}
	if validation := support.Validate(matrix, support.Entry{Target: "opencode", Version: req.TargetVersion, Construct: "permission"}); validation.Completeness != model.CompletenessComplete {
		status := SupportUnresolved
		for _, finding := range validation.Findings {
			if finding.Code() == "unsupported-version" {
				status = SupportUnsupported
			}
		}
		result := auditResult(req, build, UnsupportedOrIncomplete, "OpenCode support evidence is incomplete", status)
		result.Findings = append([]model.Finding(nil), validation.Findings...)
		return result
	}

	plan := opencode.DiscoveryPlan(req.Root, []string{req.ExplicitConfig})
	selection := source.Execute(context.Background(), plan, options.files, source.CaptureEnv(plan.EnvKeys, options.lookupEnv))
	if len(selection.Findings) != 0 || len(selection.Selected) != 1 {
		result := auditResult(req, build, UnsupportedOrIncomplete, "OpenCode source selection is incomplete", SupportSupported)
		result.SourceDigests = sourceDigests(selection.Selected)
		result.Findings = sourceFindings(selection.Findings)
		return result
	}
	report := opencode.ResolveWithOptions(selection.Selected, opencode.Options{Version: req.TargetVersion, Support: matrix})
	result := auditResult(req, build, UnsupportedOrIncomplete, "OpenCode audit is unsupported or incomplete", SupportSupported)
	result.SourceDigests = sourceDigests(selection.Selected)
	result.Permissions = append([]model.Permission(nil), report.Permissions...)
	result.Findings = append([]model.Finding(nil), report.Findings...)
	if report.Completeness == model.CompletenessComplete && len(report.Findings) == 0 {
		result.Category = CompleteNoFindings
		result.Message = ""
		result.Completeness = model.CompletenessComplete
		return result
	}
	return result
}

func auditResult(req Request, build BuildInfo, category Category, message string, status SupportStatus) Result {
	return Result{
		Category:               category,
		Message:                message,
		Build:                  build,
		Target:                 safeTarget(req.Target),
		RequestedTargetVersion: req.TargetVersion,
		SupportStatus:          status,
		SourceDigests:          []string{},
		Completeness:           model.CompletenessIncomplete,
		Permissions:            []model.Permission{},
		Findings:               []model.Finding{},
		Limitations:            auditLimitations(req),
	}
}

func safeTarget(target Target) Target {
	if target == TargetOpenCode || target == TargetClaudeCode {
		return target
	}
	return ""
}

func auditLimitations(req Request) []string {
	if req.Target != TargetOpenCode {
		return []string{}
	}
	return []string{"modeled-static-configuration"}
}

func sourceDigests(selected []source.SelectedSource) []string {
	if len(selected) == 0 {
		return []string{}
	}
	seen := make(map[string]bool, len(selected))
	out := make([]string, 0, len(selected))
	for _, item := range selected {
		if len(item.Identity) != 64 || strings.ToLower(item.Identity) != item.Identity || seen[item.Identity] {
			continue
		}
		if _, err := hex.DecodeString(item.Identity); err != nil {
			continue
		}
		seen[item.Identity] = true
		out = append(out, "sha256:"+item.Identity)
	}
	sort.Strings(out)
	return out
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
