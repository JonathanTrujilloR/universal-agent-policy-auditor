// Package opencode models OpenCode native permission configuration semantics.
package opencode

import (
	"bytes"
	"encoding/json"
	"io"
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

var capabilities = []string{"bash", "edit", "read"}
var effects = map[string]model.Effect{"allow": model.EffectAllow, "deny": model.EffectDeny, "ask": model.EffectAsk}

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

// ResolveWithOptions parses the exact OpenCode 1.18.27 legacy scalar permission object.
func ResolveWithOptions(sources []source.SelectedSource, options Options) Report {
	if unsupported(options) {
		return failed("opencode-unsupported-version", "OpenCode version is not supported by evidence-backed metadata")
	}
	if len(sources) != 1 || strings.TrimSpace(sources[0].Identity) == "" || strings.TrimSpace(sources[0].Path) == "" {
		return failed("opencode-source-selection-incomplete", "OpenCode requires exactly one selected source with identity and path")
	}
	selected := sources[0]
	if len(selected.Data) > maxBytes {
		return failed("opencode-source-data-too-large", "OpenCode selected source exceeds the adapter byte limit")
	}
	rules, ok := parse(selected.Data, selected.Path)
	if !ok {
		return failed("opencode-permission-shape-unsupported", "OpenCode permission object shape is unsupported")
	}
	report := Report{Completeness: model.CompletenessComplete}
	for _, capability := range capabilities {
		report.Permissions = append(report.Permissions, modeledPermission(capability, rules[capability]))
	}
	seen := map[string]bool{}
	var requested []string
	for _, value := range options.RequestedCapabilities {
		capability := strings.TrimSpace(value)
		if capability == "" {
			if !seen[capability] {
				report.Completeness = model.CompletenessIncomplete
				report.Findings = append(report.Findings, model.NewFinding("opencode-invalid-requested-capability", "OpenCode requested capability is empty"))
			}
			seen[capability] = true
			continue
		}
		if !seen[capability] {
			requested = append(requested, capability)
		}
		seen[capability] = true
	}
	sort.Strings(requested)
	for _, capability := range requested {
		if _, ok := rules[capability]; ok {
			continue
		}
		report.Completeness = model.CompletenessIncomplete
		report.Findings = append(report.Findings, model.NewFinding("opencode-unresolved-default", "OpenCode default permission is not modeled for the requested capability"))
		report.Permissions = append(report.Permissions, unresolvedDefault(capability))
	}
	sort.Slice(report.Permissions, func(i, j int) bool { return report.Permissions[i].Capability() < report.Permissions[j].Capability() })
	return report
}

type permissionRule struct {
	effect        model.Effect
	location, raw string
}

func failed(code, message string) Report {
	return Report{Completeness: model.CompletenessIncomplete, Findings: []model.Finding{model.NewFinding(code, message)}}
}

func parse(data []byte, path string) (map[string]permissionRule, bool) {
	if !strictObjectKeys(data) {
		return nil, false
	}
	var doc map[string]map[string]string
	if err := json.Unmarshal(data, &doc); err != nil || len(doc) != 1 || doc["permission"] == nil || len(doc["permission"]) != 3 {
		return nil, false
	}
	permissions := doc["permission"]
	rules := map[string]permissionRule{}
	for _, capability := range []string{"read", "edit", "bash"} {
		effect, ok := effects[permissions[capability]]
		if !ok {
			return nil, false
		}
		rules[capability] = permissionRule{effect: effect, location: path + "#permission." + capability, raw: permissions[capability]}
	}
	return rules, true
}

func strictObjectKeys(data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') || !scanObject(decoder) {
		return false
	}
	token, err := decoder.Token()
	return err == io.EOF && token == nil
}

func scanObject(decoder *json.Decoder) bool {
	allowed, seen := map[string]bool{"permission": true, "read": true, "edit": true, "bash": true}, map[string]bool{}
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || seen[key] || !allowed[key] {
			return false
		}
		seen[key] = true
		token, err = decoder.Token()
		if delimiter, ok := token.(json.Delim); err != nil || ok && (delimiter != '{' || !scanObject(decoder)) {
			return false
		}
	}
	token, err := decoder.Token()
	return err == nil && token == json.Delim('}')
}

func modeledPermission(capability string, rule permissionRule) model.Permission {
	provenance := evidence(rule)
	trace := model.NewResolutionTrace([]model.TraceStep{{Kind: model.TraceRule, After: rule.effect, Rules: []string{rule.location}, Rationale: "OpenCode permission rule", Evidence: provenance}}, nil, model.CompletenessComplete)
	return model.NewPermissionWithProvenance(capability, rule.effect, model.Scope("opencode"), "*", nil, []model.Provenance{provenance}, []model.Extension{{Target: "opencode", Key: "modeled", Value: "static-configuration"}}, trace)
}

func unresolvedDefault(capability string) model.Permission {
	provenance := model.NewProvenance("opencode", "requested:"+capability, "")
	trace := model.NewResolutionTrace([]model.TraceStep{{Kind: model.TraceDefault, After: model.EffectUnresolved, Rules: []string{"requested:" + capability}, Rationale: "OpenCode default is unresolved without version-pinned default semantics", Evidence: provenance}}, []string{"default permission for " + capability}, model.CompletenessIncomplete)
	return model.NewPermissionWithProvenance(capability, model.EffectUnresolved, model.Scope("opencode"), "*", nil, []model.Provenance{provenance}, []model.Extension{{Target: "opencode", Key: "modeled", Value: "static-configuration"}, {Target: "opencode", Key: "limitation", Value: "unresolved-default"}}, trace)
}

func unsupported(options Options) bool {
	if options.Version != "1.18.27" {
		return true
	}
	result := support.Validate(options.Support, support.Entry{Target: "opencode", Version: options.Version, Construct: "permission"})
	return result.Completeness != model.CompletenessComplete
}

func evidence(rule permissionRule) model.Provenance {
	return model.NewProvenance("opencode", rule.location, rule.raw)
}
