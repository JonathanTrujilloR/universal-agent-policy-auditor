// Package opencode models OpenCode native permission configuration semantics.
package opencode

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/source"
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

// Resolve parses data-only OpenCode JSON and applies last-rule precedence by source order.
func Resolve(sources []source.SelectedSource) Report {
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
		report.Permissions = append(report.Permissions, final.permission(capability, chain))
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
	trace := model.NewResolutionTrace(steps, unresolved, completeness(unresolved))
	return model.NewPermissionWithProvenance(capability, effect, model.Scope("opencode"), r.matcher, conditions, []model.Provenance{evidence(r)}, []model.Extension{{Target: "opencode", Key: "modeled", Value: "static-configuration"}}, trace)
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
