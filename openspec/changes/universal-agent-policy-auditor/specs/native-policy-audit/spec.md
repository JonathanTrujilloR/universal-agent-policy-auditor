# Native Policy Audit Specification

## Purpose

The alpha SHALL provide a local, read-only audit of supported OpenCode and Claude Code configurations. It SHALL report modeled static semantics and SHALL NOT imply runtime enforcement.

## Requirements

### Requirement: Select bounded native sources

The auditor MUST accept explicit OpenCode and Claude Code sources and MAY discover them using bounded search rules. It MUST report selected, skipped, duplicate, unreadable, and ambiguous sources without silently choosing.

#### Scenario: Explicit sources are audited
- GIVEN explicit sources for both supported clients
- WHEN an audit is requested
- THEN only the selected sources are analyzed and their identities are reported

#### Scenario: Discovery is ambiguous
- GIVEN bounded discovery finds competing sources with no deterministic selection
- WHEN an audit is requested
- THEN the result is incomplete and identifies the ambiguity

### Requirement: Classify support and preserve target evidence

The auditor MUST classify each source by client/configuration version and support status. Adapters MUST parse declared constructs, preserve target semantics, and attach redacted provenance to every represented permission. Unknown versions or permission-bearing constructs MUST prevent unqualified success.

#### Scenario: Supported rule has evidence
- GIVEN a supported configuration and recognized permission rule
- WHEN it is parsed
- THEN the modeled rule includes target, version, source reference, rule identity or location, and status

#### Scenario: Unknown permission construct is encountered
- GIVEN a configuration contains an unrecognized permission-bearing construct
- WHEN it is parsed
- THEN a blocking unknown or unsupported finding is emitted and the audit is not an unqualified success

### Requirement: Resolve effective modeled permissions

The auditor MUST model scopes, merge/precedence, defaults, matchers, conditions, and effects according to verified target semantics. It MUST distinguish allow, deny, ask, inherited/default, conditional, and unresolved outcomes with an explanation trace. Static results MUST be labeled as modeled configuration semantics, not guaranteed runtime enforcement.

#### Scenario: Precedence changes an effective result
- GIVEN supported rules whose documented precedence produces a final effect
- WHEN effective permissions are resolved
- THEN the final effect and contributing rules/defaults are reported deterministically

#### Scenario: Runtime context exceeds static analysis
- GIVEN permissions depend on trust, interactive approval, hooks, plugins, sandboxing, or other runtime context
- WHEN the audit resolves permissions
- THEN the affected result is conditional or unresolved and carries a static-vs-runtime limitation

### Requirement: Compare only evidence-backed semantics

The auditor MUST compare effective permissions across OpenCode and Claude Code by capability, resource, effect, scope, matcher, condition, and enforcement limitation. It MUST classify results as equivalent, broader, narrower, target-only, unsupported, ambiguous, lossy, or not comparable; equivalent MUST require evidence across dimensions.

#### Scenario: One client is broader
- GIVEN both clients have comparable modeled permissions and one grants broader scope
- WHEN they are compared
- THEN the result identifies the broader client and the responsible semantic dimension

#### Scenario: Mapping loses meaning
- GIVEN a target-specific rule cannot be represented without losing a permission-bearing semantic
- WHEN clients are compared
- THEN the result is lossy or not comparable and names the lost semantic and consequence

### Requirement: Fail safely and report deterministically

The auditor MUST treat malformed, unreadable, unsupported, unknown, ambiguous, or lossy input as visible findings with safe failure status. It MUST NOT execute configuration content, follow unbounded references, or write files. Human output MUST prioritize blocking findings; machine output MUST be versioned, schema-valid, ordered, and redacted by default.

#### Scenario: Malformed input fails without mutation
- GIVEN a selected source is malformed or unreadable
- WHEN an audit is requested
- THEN an actionable incomplete or operational result is returned, no target file changes, and no unredacted secret values appear

#### Scenario: Repeated audit is deterministic
- GIVEN identical supported inputs, declared environment, and tool version
- WHEN the audit is run repeatedly in human and machine modes
- THEN machine output is byte-stable and human output reports the same sources, findings, and limitations
