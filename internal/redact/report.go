// Package redact converts application results into safe report values.
package redact

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
)

var ErrUnsafeReport = errors.New("redact: unsafe report")

type Category string
type Target string
type SupportStatus string
type Completeness string
type Capability string
type Effect string
type Scope string
type Matcher string
type Digest string
type FindingCode string
type Limitation string

const (
	CategoryCompleteNoFindings      Category      = "complete_no_findings"
	CategoryUnsupportedOrIncomplete Category      = "unsupported_or_incomplete"
	CategoryInvalidRequest          Category      = "invalid_request"
	CategoryOperationalFailure      Category      = "operational_failure"
	TargetOpenCode                  Target        = "opencode"
	TargetClaudeCode                Target        = "claudecode"
	SupportSupported                SupportStatus = "supported"
	SupportUnsupported              SupportStatus = "unsupported"
	SupportUnresolved               SupportStatus = "unresolved"
	CompletenessComplete            Completeness  = "complete"
	CompletenessIncomplete          Completeness  = "incomplete"
)

var (
	categories      = map[app.Category]Category{app.CompleteNoFindings: CategoryCompleteNoFindings, app.UnsupportedOrIncomplete: CategoryUnsupportedOrIncomplete, app.InvalidRequest: CategoryInvalidRequest, app.OperationalFailure: CategoryOperationalFailure}
	targets         = map[app.Target]Target{"": "", app.TargetOpenCode: TargetOpenCode, app.TargetClaudeCode: TargetClaudeCode}
	supports        = map[app.SupportStatus]SupportStatus{app.SupportSupported: SupportSupported, app.SupportUnsupported: SupportUnsupported, app.SupportUnresolved: SupportUnresolved}
	completes       = map[model.Completeness]Completeness{model.CompletenessComplete: CompletenessComplete, model.CompletenessIncomplete: CompletenessIncomplete}
	allowedFindings = map[string]bool{"unsupported-version": true, "opencode-unsupported-version": true, "opencode-source-selection-incomplete": true, "opencode-source-data-too-large": true, "opencode-permission-shape-unsupported": true, "opencode-invalid-requested-capability": true, "opencode-unresolved-default": true, "source-cancelled": true, "source-invalid-root": true, "source-outside-root": true, "source-unreadable": true, "source-duplicate": true, "source-file-limit": true, "source-byte-limit": true}
)

type Report struct {
	valid                         bool
	category                      Category
	exitCode                      int
	toolVersion, requestedVersion string
	target                        Target
	supportStatus                 SupportStatus
	completeness                  Completeness
	sources                       []Digest
	permissions                   []Permission
	findings                      []Finding
	limitations                   []Limitation
}
type Permission struct {
	capability Capability
	effect     Effect
	scope      Scope
	matcher    Matcher
	provenance []Digest
}
type Finding struct{ code FindingCode }

func NewReport(result app.Result) (Report, error) {
	data := Report{exitCode: app.ExitCode(result.Category)}
	var ok bool
	if data.category, ok = categories[result.Category]; !ok {
		return Report{}, ErrUnsafeReport
	}
	if data.target, ok = targets[result.Target]; !ok {
		return Report{}, ErrUnsafeReport
	}
	if data.supportStatus, ok = supports[result.SupportStatus]; !ok {
		return Report{}, ErrUnsafeReport
	}
	if data.completeness, ok = completes[result.Completeness]; !ok {
		return Report{}, ErrUnsafeReport
	}
	withheld := false
	if data.toolVersion, ok = safeToolVersion(result.Build.Version); !ok {
		withheld = true
	}
	if data.requestedVersion, ok = safeRequestedVersion(result.Target, result.RequestedTargetVersion); !ok {
		withheld = true
	}
	if data.sources, ok = safeDigests(result.SourceDigests); !ok {
		return Report{}, ErrUnsafeReport
	}
	if data.permissions, ok = safePermissions(result.Permissions); !ok {
		return Report{}, ErrUnsafeReport
	}
	if data.findings, ok = safeFindings(result.Findings); !ok {
		return Report{}, ErrUnsafeReport
	}
	if withheld {
		data.findings = append(data.findings, Finding{code: "redact-version-withheld"})
	}
	sort.Slice(data.findings, func(i, j int) bool { return data.findings[i].code < data.findings[j].code })
	if data.limitations, ok = safeLimitations(result.Limitations); !ok || !validState(data) {
		return Report{}, ErrUnsafeReport
	}
	data.valid = true
	return data, nil
}

func (r Report) Valid() bool                     { return r.valid }
func (r Report) Category() Category              { return r.category }
func (r Report) ExitCode() int                   { return r.exitCode }
func (r Report) ToolVersion() string             { return r.toolVersion }
func (r Report) RequestedVersion() string        { return r.requestedVersion }
func (r Report) Target() Target                  { return r.target }
func (r Report) SupportStatus() SupportStatus    { return r.supportStatus }
func (r Report) Completeness() Completeness      { return r.completeness }
func (p Permission) Capability() Capability      { return p.capability }
func (p Permission) Effect() Effect              { return p.effect }
func (p Permission) Scope() Scope                { return p.scope }
func (p Permission) Matcher() Matcher            { return p.matcher }
func (f Finding) Code() FindingCode              { return f.code }
func (r Report) SourceDigests() []Digest         { return append([]Digest{}, r.sources...) }
func (r Report) Findings() []Finding             { return append([]Finding{}, r.findings...) }
func (r Report) Limitations() []Limitation       { return append([]Limitation{}, r.limitations...) }
func (r Report) Permissions() []Permission       { return copyPermissions(r.permissions) }
func (p Permission) ProvenanceDigests() []Digest { return append([]Digest{}, p.provenance...) }

func validState(d Report) bool {
	if !limitationsMatchTarget(d) {
		return false
	}
	appFindings := appFindingCount(d.findings)
	noDetails := len(d.sources) == 0 && len(d.permissions) == 0 && appFindings == 0
	switch d.category {
	case CategoryCompleteNoFindings:
		return d.target == TargetOpenCode && d.supportStatus == SupportSupported && d.completeness == CompletenessComplete && len(d.findings) == 0 && len(d.sources) > 0 && successPermissions(d.permissions)
	case CategoryUnsupportedOrIncomplete:
		if d.completeness != CompletenessIncomplete || len(d.permissions) != 0 {
			return false
		}
		if d.target == TargetOpenCode && d.supportStatus == SupportSupported {
			return appFindings > 0
		}
		if d.target == TargetOpenCode && d.supportStatus == SupportUnsupported {
			return len(d.sources) == 0 && appFindings == 1 && d.findings[len(d.findings)-1].code == "unsupported-version"
		}
		return noDetails && (d.target == TargetOpenCode && d.supportStatus == SupportUnresolved || d.target == TargetClaudeCode && d.supportStatus == SupportUnsupported)
	case CategoryInvalidRequest:
		return d.supportStatus == SupportUnresolved && d.completeness == CompletenessIncomplete && noDetails
	case CategoryOperationalFailure:
		return d.target != "" && d.supportStatus == SupportUnresolved && d.completeness == CompletenessIncomplete && noDetails
	}
	return false
}
func limitationsMatchTarget(d Report) bool {
	switch d.target {
	case TargetOpenCode:
		return len(d.limitations) == 1 && d.limitations[0] == "modeled-static-configuration"
	case TargetClaudeCode, "":
		return len(d.limitations) == 0
	}
	return false
}
func appFindingCount(values []Finding) int {
	count := 0
	for _, value := range values {
		if value.code != "redact-version-withheld" {
			count++
		}
	}
	return count
}

// Versions are public constants, not a syntax-based channel for local metadata.
func safeToolVersion(value string) (string, bool) {
	if value == "dev" || value == "v0.1.0-alpha.1" {
		return value, true
	}
	return "", false
}
func safeRequestedVersion(target app.Target, value string) (string, bool) {
	if value == "" || target == app.TargetOpenCode && value == "1.18.27" {
		return value, true
	}
	return "", false
}
func safeDigests(values []string) ([]Digest, bool) {
	seen, out := map[string]bool{}, make([]Digest, 0, len(values))
	for _, value := range values {
		if !validDigest(value) {
			return nil, false
		}
		if !seen[value] {
			seen[value] = true
			out = append(out, Digest(value))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, true
}
func validDigest(value string) bool {
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return len(value) == 71 && strings.HasPrefix(value, "sha256:") && !strings.Contains(value[7:], "sha256:") && strings.ToLower(value) == value && err == nil
}
func safePermissions(values []model.Permission) ([]Permission, bool) {
	out, byKey, seen := make([]Permission, 0, len(values)), map[string]Effect{}, map[string]bool{}
	for _, value := range values {
		permission, ok := safePermission(value)
		if !ok {
			return nil, false
		}
		key := string(permission.capability) + "\x00" + string(permission.scope) + "\x00" + string(permission.matcher)
		if prior, exists := byKey[key]; exists && prior != permission.effect {
			return nil, false
		}
		byKey[key] = permission.effect
		full := key + "\x00" + string(permission.effect) + "\x00" + fmt.Sprint(permission.provenance)
		if !seen[full] {
			seen[full] = true
			out = append(out, permission)
		}
	}
	sort.Slice(out, func(i, j int) bool { return permissionKey(out[i]) < permissionKey(out[j]) })
	return out, true
}
func safePermission(value model.Permission) (Permission, bool) {
	if !closedPermission(value) || !closedTrace(value) {
		return Permission{}, false
	}
	provenance, ok := safeDigests([]string{value.Provenance()[0].Value()})
	if !ok {
		return Permission{}, false
	}
	return Permission{capability: Capability(value.Capability()), effect: Effect(value.Effect()), scope: Scope(value.Scope()), matcher: Matcher(value.Matcher()), provenance: provenance}, true
}
func closedPermission(p model.Permission) bool {
	capability, effect := p.Capability(), p.Effect()
	return (capability == "bash" || capability == "edit" || capability == "read") && (effect == model.EffectAllow || effect == model.EffectDeny || effect == model.EffectAsk) && p.Scope() == "opencode" && p.Matcher() == "*" && len(p.Conditions()) == 0 && len(p.Extensions()) == 1 && p.Extensions()[0] == (model.Extension{Target: "opencode", Key: "modeled", Value: "static-configuration"})
}
func closedTrace(p model.Permission) bool {
	trace := p.Trace()
	steps := trace.Steps()
	return trace.Status() == model.CompletenessComplete && len(trace.Unresolved()) == 0 && len(steps) == 1 && steps[0].Kind == model.TraceRule && steps[0].After == p.Effect() && len(p.Provenance()) == 1 && steps[0].Evidence == p.Provenance()[0]
}
func permissionKey(p Permission) string {
	return string(p.capability) + "\x00" + string(p.effect) + "\x00" + string(p.scope) + "\x00" + string(p.matcher) + "\x00" + fmt.Sprint(p.provenance)
}
func copyPermissions(values []Permission) []Permission {
	out := append([]Permission{}, values...)
	for i := range out {
		out[i].provenance = append([]Digest{}, values[i].provenance...)
	}
	return out
}
func successPermissions(values []Permission) bool {
	return len(values) == 3 && values[0].capability == "bash" && values[1].capability == "edit" && values[2].capability == "read"
}
func safeFindings(values []model.Finding) ([]Finding, bool) {
	seen, out := map[string]bool{}, make([]Finding, 0, len(values))
	for _, value := range values {
		code := value.Code()
		if !allowedFindings[code] {
			return nil, false
		}
		if !seen[code] {
			seen[code] = true
			out = append(out, Finding{code: FindingCode(code)})
		}
	}
	return out, true
}
func safeLimitations(values []string) ([]Limitation, bool) {
	if len(values) == 0 {
		return nil, true
	}
	for _, value := range values {
		if value != "modeled-static-configuration" {
			return nil, false
		}
	}
	return []Limitation{"modeled-static-configuration"}, true
}
