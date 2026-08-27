package claudecode

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/model"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/source"
	"github.com/jkelevra/universal-agent-policy-auditor/support"
)

type Report struct {
	Completeness model.Completeness
	Permissions  []model.Permission
	Findings     []model.Finding
}

type Options struct {
	Version               string
	Support               support.Matrix
	RequestedCapabilities []string
}

func DiscoveryPlan(root string, explicit []string) source.Plan {
	sources := append([]string(nil), explicit...)
	if len(sources) == 0 {
		sources = []string{filepath.Join(root, ".claude", "settings.json"), filepath.Join(root, ".claude", "settings.local.json")}
	}
	return source.Plan{Roots: []string{root}, Sources: sources, EnvKeys: []string{"CLAUDE_CONFIG_DIR", "XDG_CONFIG_HOME", "HOME"}, MaxFiles: 8, MaxBytes: 256 << 10}
}

func Resolve(sources []source.SelectedSource) Report { return ResolveWithOptions(sources, Options{}) }

func ResolveWithOptions(sources []source.SelectedSource, options Options) Report {
	if unsupported(options) {
		return Report{Completeness: model.CompletenessIncomplete, Findings: []model.Finding{model.NewFinding("claudecode-unsupported-version", "Claude Code version is not supported by evidence-backed metadata")}}
	}
	if ambiguousSourceOrder(sources) {
		return Report{Completeness: model.CompletenessIncomplete, Findings: []model.Finding{model.NewFinding("claudecode-ambiguous-source-order", "Claude Code source order is ambiguous")}}
	}
	report := Report{Completeness: model.CompletenessComplete}
	rules := map[string][]permissionRule{}
	for _, selected := range sources {
		parsed, findings := parse(selected)
		report.Findings = append(report.Findings, findings...)
		for matcher, chain := range parsed {
			rules[matcher] = append(rules[matcher], chain...)
		}
	}
	if len(report.Findings) > 0 {
		report.Completeness = model.CompletenessIncomplete
	}
	seenCapabilities := map[string]bool{}
	for _, chain := range rules {
		capability := chain[len(chain)-1].capability()
		seenCapabilities[capability] = true
		permission := chain[len(chain)-1].permission(capability, chain)
		if permission.Trace().Status() == model.CompletenessIncomplete {
			report.Completeness = model.CompletenessIncomplete
			report.Findings = append(report.Findings, model.NewFinding("claudecode-runtime-limit", "Claude Code permission depends on runtime context outside static analysis"))
		}
		report.Permissions = append(report.Permissions, permission)
	}
	for _, capability := range options.RequestedCapabilities {
		if !seenCapabilities[capability] {
			report.Completeness = model.CompletenessIncomplete
			report.Findings = append(report.Findings, model.NewFinding("claudecode-unresolved-default", "Claude Code default permission is not modeled for the requested capability"))
			report.Permissions = append(report.Permissions, unresolvedDefault(capability))
		}
	}
	sort.Slice(report.Permissions, func(i, j int) bool {
		return report.Permissions[i].Capability()+report.Permissions[i].Matcher() < report.Permissions[j].Capability()+report.Permissions[j].Matcher()
	})
	return report
}

type permissionRule struct {
	effect                            model.Effect
	matcher, condition, location, raw string
}

func parse(selected source.SelectedSource) (map[string][]permissionRule, []model.Finding) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(selected.Data, &top); err != nil {
		return nil, []model.Finding{model.NewFinding("claudecode-malformed", "Claude Code configuration is malformed")}
	}
	var findings []model.Finding
	for key := range top {
		switch key {
		case "permissions":
		case "hooks":
			findings = append(findings, model.NewFinding("claudecode-runtime-construct", "Claude Code runtime hook configuration is outside static permission analysis"))
		default:
			findings = append(findings, model.NewFinding("claudecode-unknown-semantic", "Claude Code configuration field is outside static permission analysis"))
		}
	}
	var doc struct {
		Permissions map[string][]json.RawMessage `json:"permissions"`
	}
	if err := json.Unmarshal(selected.Data, &doc); err != nil {
		return nil, []model.Finding{model.NewFinding("claudecode-malformed", "Claude Code configuration is malformed")}
	}
	parsed := map[string][]permissionRule{}
	for _, bucket := range []string{"allow", "ask", "deny"} {
		for index, raw := range doc.Permissions[bucket] {
			rule, finding := parseRule(raw, bucket)
			rule.location, rule.raw = fmt.Sprintf("%s#permissions.%s[%d]", selected.Path, bucket, index), string(raw)
			if finding != "" {
				message := "Claude Code permission rule field is outside static permission analysis"
				if finding == "claudecode-malformed-semantic" {
					message = "Claude Code permission rule field has malformed JSON type"
				}
				findings = append(findings, model.NewFinding(finding, message))
			}
			if rule.matcher == "" {
				if finding != "claudecode-malformed-semantic" {
					findings = append(findings, model.NewFinding("claudecode-unknown-permission", "Claude Code permission entry is unsupported"))
				}
				continue
			}
			parsed[rule.matcher] = append(parsed[rule.matcher], rule)
		}
	}
	for bucket := range doc.Permissions {
		if bucket != "allow" && bucket != "ask" && bucket != "deny" {
			findings = append(findings, model.NewFinding("claudecode-unknown-effect", "Claude Code permission effect is unsupported"))
		}
	}
	return parsed, findings
}

func parseRule(raw json.RawMessage, bucket string) (permissionRule, string) {
	var matcher string
	if err := json.Unmarshal(raw, &matcher); err == nil {
		return permissionRule{effect: toEffect(bucket), matcher: matcher}, ""
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return permissionRule{}, ""
	}
	finding, condition := "", ""
	for key, raw := range object {
		switch strings.ToLower(key) {
		case "match":
			if err := json.Unmarshal(raw, &matcher); err != nil {
				return permissionRule{}, "claudecode-malformed-semantic"
			}
		case "condition":
			if err := json.Unmarshal(raw, &condition); err != nil {
				return permissionRule{}, "claudecode-malformed-semantic"
			}
		default:
			finding = "claudecode-unknown-semantic"
		}
	}
	if matcher == "" {
		return permissionRule{}, finding
	}
	return permissionRule{effect: toEffect(bucket), matcher: matcher, condition: condition}, finding
}

func (r permissionRule) capability() string {
	before, _, found := strings.Cut(r.matcher, "(")
	if found && before != "" {
		return before
	}
	return r.matcher
}

func (r permissionRule) permission(capability string, chain []permissionRule) model.Permission {
	effect, conditions, unresolved := r.effect, []model.Condition(nil), []string(nil)
	if r.condition != "" {
		effect, conditions, unresolved = model.EffectConditional, []model.Condition{model.Condition(r.condition)}, []string{r.condition}
	}
	steps := []model.TraceStep{{Kind: model.TraceRule, After: chain[0].effect, Rules: []string{chain[0].location}, Rationale: "Claude Code permission rule", Evidence: evidence(chain[0])}}
	if len(chain) > 1 {
		steps = append(steps, model.TraceStep{Kind: model.TracePrecedence, Before: chain[len(chain)-2].effect, After: effect, Rules: []string{r.location}, Rationale: "later Claude Code source or higher-priority bucket overrides earlier rule", Evidence: evidence(r)})
	}
	extensions := []model.Extension{{Target: "claudecode", Key: "modeled", Value: "static-configuration"}}
	if len(unresolved) > 0 {
		steps = append(steps, model.TraceStep{Kind: model.TraceRuntime, Before: r.effect, After: effect, Rules: []string{r.location}, Rationale: "runtime context is outside static configuration analysis", Evidence: evidence(r)})
		extensions = append(extensions, model.Extension{Target: "claudecode", Key: "limitation", Value: "static-not-runtime-enforcement"})
	}
	status := model.CompletenessComplete
	if len(unresolved) > 0 {
		status = model.CompletenessIncomplete
	}
	trace := model.NewResolutionTrace(steps, unresolved, status)
	return model.NewPermissionWithProvenance(capability, effect, model.Scope("claudecode"), r.matcher, conditions, []model.Provenance{evidence(r)}, extensions, trace)
}

func unresolvedDefault(capability string) model.Permission {
	provenance := model.NewProvenance("claudecode", "requested:"+capability, "")
	trace := model.NewResolutionTrace([]model.TraceStep{{Kind: model.TraceDefault, After: model.EffectUnresolved, Rules: []string{"requested:" + capability}, Rationale: "Claude Code default is unresolved without version-pinned default semantics", Evidence: provenance}}, []string{"default permission for " + capability}, model.CompletenessIncomplete)
	return model.NewPermissionWithProvenance(capability, model.EffectUnresolved, model.Scope("claudecode"), capability, nil, []model.Provenance{provenance}, []model.Extension{{Target: "claudecode", Key: "modeled", Value: "static-configuration"}, {Target: "claudecode", Key: "limitation", Value: "unresolved-default"}}, trace)
}

func unsupported(options Options) bool {
	if options.Version == "" {
		return true
	}
	return support.Validate(options.Support, support.Entry{Target: "claudecode", Version: options.Version, Construct: "permission"}).Completeness != model.CompletenessComplete
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
	return model.NewProvenance("claudecode", rule.location, rule.raw)
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
