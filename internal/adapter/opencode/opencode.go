// Package opencode models OpenCode native permission configuration semantics.
package opencode

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/source"
	"github.com/jkelevra/universal-agent-policy-auditor/support"
)

const (
	maxFiles = 8
	maxBytes = 256 << 10
)

// Report is the adapter-local semantic result before cross-client comparison.
type Report struct {
	Completeness model.Completeness
	Permissions  []model.Permission
	Findings     []model.Finding
}

// Options binds adapter resolution to evidence-backed support metadata.
type Options struct {
	Version               string
	Support               support.Matrix
	RequestedCapabilities []string
}

// DiscoveryPlan declares bounded OpenCode source selection; it performs no I/O.
func DiscoveryPlan(root string, explicit []string) source.Plan {
	sources := append([]string(nil), explicit...)
	if len(sources) == 0 {
		sources = []string{filepath.Join(root, "opencode.json"), filepath.Join(root, ".opencode.json")}
	}
	return source.Plan{
		Roots:    []string{root},
		Sources:  sources,
		EnvKeys:  []string{"OPENCODE_CONFIG", "XDG_CONFIG_HOME", "HOME"},
		MaxFiles: maxFiles,
		MaxBytes: maxBytes,
	}
}

// Resolve parses data-only OpenCode JSON with the checked-in zero-support gate.
func Resolve(sources []source.SelectedSource) Report {
	return ResolveWithOptions(sources, Options{})
}

// ResolveWithOptions parses data-only OpenCode JSON and applies last-rule precedence by source order.
func ResolveWithOptions(sources []source.SelectedSource, options Options) Report {
	if unsupported(options) {
		return Report{Completeness: model.CompletenessIncomplete, Findings: []model.Finding{model.NewFinding("opencode-unsupported-version", "OpenCode version is not supported by evidence-backed metadata")}}
	}
	if ambiguousSourceOrder(sources) {
		return Report{Completeness: model.CompletenessIncomplete, Findings: []model.Finding{model.NewFinding("opencode-ambiguous-source-order", "OpenCode source order is ambiguous")}}
	}
	report := Report{Completeness: model.CompletenessComplete}
	rules := map[string][]permissionRule{}
	for _, selected := range sources {
		parsed, findings := parse(selected)
		report.Findings = append(report.Findings, findings...)
		for capability, rule := range parsed {
			rules[capability] = append(rules[capability], rule)
		}
	}
	if len(report.Findings) > 0 {
		report.Completeness = model.CompletenessIncomplete
	}
	for capability, chain := range rules {
		if len(chain) == 0 {
			continue
		}
		final := chain[len(chain)-1]
		if final.invalid != "" {
			continue
		}
		permission := final.permission(capability, chain)
		if permission.Trace().Status() == model.CompletenessIncomplete {
			report.Completeness = model.CompletenessIncomplete
			report.Findings = append(report.Findings, model.NewFinding("opencode-runtime-limit", "OpenCode permission depends on runtime context outside static analysis"))
		}
		report.Permissions = append(report.Permissions, permission)
	}
	for _, capability := range options.RequestedCapabilities {
		if _, ok := rules[capability]; !ok {
			report.Completeness = model.CompletenessIncomplete
			report.Findings = append(report.Findings, model.NewFinding("opencode-unresolved-default", "OpenCode default permission is not modeled for the requested capability"))
			report.Permissions = append(report.Permissions, unresolvedDefault(capability))
		}
	}
	sort.Slice(report.Permissions, func(i, j int) bool { return report.Permissions[i].Capability() < report.Permissions[j].Capability() })
	return report
}

type permissionRule struct {
	effect    model.Effect
	matcher   string
	condition string
	location  string
	raw       string
	invalid   string
}

func parse(selected source.SelectedSource) (map[string]permissionRule, []model.Finding) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(selected.Data, &top); err != nil {
		return nil, []model.Finding{model.NewFinding("opencode-malformed", "OpenCode configuration is malformed")}
	}
	if _, ok := top["permissions"]; ok {
		return nil, []model.Finding{model.NewFinding("opencode-unknown-construct", "OpenCode permission-bearing construct is unsupported")}
	}
	var doc struct {
		Permission map[string]json.RawMessage `json:"permission"`
	}
	if err := json.Unmarshal(selected.Data, &doc); err != nil {
		return nil, []model.Finding{model.NewFinding("opencode-malformed", "OpenCode configuration is malformed")}
	}
	parsed := map[string]permissionRule{}
	var findings []model.Finding
	for capability, raw := range doc.Permission {
		rule := parseRule(raw)
		rule.matcher = defaultString(rule.matcher, capability)
		rule.location = selected.Path + "#permission." + capability
		rule.raw = string(raw)
		if rule.invalid != "" {
			findings = append(findings, model.NewFinding(rule.invalid, "OpenCode permission rule is unsupported"))
		} else {
			parsed[capability] = rule
		}
	}
	return parsed, findings
}

func parseRule(raw json.RawMessage) permissionRule {
	var effect string
	if err := json.Unmarshal(raw, &effect); err == nil {
		return permissionRule{effect: toEffect(effect), invalid: invalidEffect(effect)}
	}
	var object struct {
		Effect    string `json:"effect"`
		Match     string `json:"match"`
		Condition string `json:"condition"`
	}
	if err := json.Unmarshal(raw, &object); err != nil || object.Effect == "" {
		return permissionRule{invalid: "opencode-unknown-permission"}
	}
	return permissionRule{effect: toEffect(object.Effect), matcher: object.Match, condition: object.Condition, invalid: invalidEffect(object.Effect)}
}

func (r permissionRule) permission(capability string, chain []permissionRule) model.Permission {
	effect := r.effect
	conditions := []model.Condition(nil)
	unresolved := []string(nil)
	if r.condition != "" {
		effect = model.EffectConditional
		conditions = []model.Condition{model.Condition(r.condition)}
		unresolved = []string{r.condition}
	}
	steps := []model.TraceStep{{Kind: model.TraceRule, After: chain[0].effect, Rules: []string{chain[0].location}, Rationale: "OpenCode permission rule", Evidence: evidence(chain[0])}}
	if len(chain) > 1 {
		steps = append(steps, model.TraceStep{Kind: model.TracePrecedence, Before: chain[len(chain)-2].effect, After: effect, Rules: []string{r.location}, Rationale: "later OpenCode source overrides earlier rule", Evidence: evidence(r)})
	}
	extensions := []model.Extension{{Target: "opencode", Key: "modeled", Value: "static-configuration"}}
	if len(unresolved) > 0 {
		steps = append(steps, model.TraceStep{Kind: model.TraceRuntime, Before: r.effect, After: effect, Rules: []string{r.location}, Rationale: "runtime context is outside static configuration analysis", Evidence: evidence(r)})
		extensions = append(extensions, model.Extension{Target: "opencode", Key: "limitation", Value: "static-not-runtime-enforcement"})
	}
	trace := model.NewResolutionTrace(steps, unresolved, completeness(unresolved))
	return model.NewPermissionWithProvenance(capability, effect, model.Scope("opencode"), r.matcher, conditions, []model.Provenance{evidence(r)}, extensions, trace)
}

func unresolvedDefault(capability string) model.Permission {
	provenance := model.NewProvenance("opencode", "requested:"+capability, "")
	trace := model.NewResolutionTrace([]model.TraceStep{{Kind: model.TraceDefault, After: model.EffectUnresolved, Rules: []string{"requested:" + capability}, Rationale: "OpenCode default is unresolved without version-pinned default semantics", Evidence: provenance}}, []string{"default permission for " + capability}, model.CompletenessIncomplete)
	return model.NewPermissionWithProvenance(capability, model.EffectUnresolved, model.Scope("opencode"), capability, nil, []model.Provenance{provenance}, []model.Extension{{Target: "opencode", Key: "modeled", Value: "static-configuration"}, {Target: "opencode", Key: "limitation", Value: "unresolved-default"}}, trace)
}

func unsupported(options Options) bool {
	if options.Version == "" {
		return true
	}
	result := support.Validate(options.Support, support.Entry{Target: "opencode", Version: options.Version, Construct: "permission"})
	return result.Completeness != model.CompletenessComplete
}

func ambiguousSourceOrder(sources []source.SelectedSource) bool {
	seen := map[string]bool{}
	for _, selected := range sources {
		if selected.Identity == "" || seen[selected.Identity] {
			return true
		}
		seen[selected.Identity] = true
	}
	return false
}

func evidence(rule permissionRule) model.Provenance {
	return model.NewProvenance("opencode", rule.location, rule.raw)
}
func completeness(unresolved []string) model.Completeness {
	if len(unresolved) > 0 {
		return model.CompletenessIncomplete
	}
	return model.CompletenessComplete
}
func toEffect(value string) model.Effect {
	switch strings.ToLower(value) {
	case "allow":
		return model.EffectAllow
	case "deny":
		return model.EffectDeny
	case "ask":
		return model.EffectAsk
	default:
		return model.EffectUnresolved
	}
}
func invalidEffect(value string) string {
	if toEffect(value) == model.EffectUnresolved {
		return "opencode-unknown-effect"
	}
	return ""
}
func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
