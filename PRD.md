# Make AI Coding-Agent Permissions Comparable and Auditable

**Status:** Draft for review
**Product:** Universal Agent Policy Auditor (working name)
**Document type:** Product Requirements Document
**Initial release:** Audit-only alpha
**Last updated:** 2026-08-08

## Executive Summary

Teams use AI coding agents whose permission systems differ in syntax, scope, precedence, and expressive power. As a result, teams duplicate security rules, struggle to determine whether two clients enforce equivalent protections, and risk configuration drift or unintended access.

This product will be a local, CLI-first policy audit tool for AI coding agents. It will discover OpenCode and Claude Code configurations, parse them into a canonical capability model, calculate effective permissions, compare clients, and report drift, unsupported semantics, and lossy interpretations. The MVP is strictly read-only: it will not write target configurations or claim equivalence when enforcement semantics differ.

The product is expected to be distributed as a single Go binary, subject to technical validation. It will require no hosted service, account, cloud control plane, runtime proxy, or LLM. Later releases may introduce guarded policy compilation only after audit behavior is mature and trusted.

## Product Decision Summary

| Area | Decision |
|---|---|
| Primary outcome | Users can understand and compare effective coding-agent permissions before relying on them. |
| Initial clients | OpenCode and Claude Code. |
| MVP boundary | Read-only discovery, parsing, audit, and comparison. |
| Policy architecture | Canonical intermediate capability and permission model with target-specific evidence. |
| Safety posture | Unknown, unsupported, ambiguous, or lossy semantics are surfaced explicitly and never treated as equivalent. |
| Operation | Local-first CLI, with no network dependency for core auditing. |
| Distribution | Likely a single Go binary; final implementation choice remains subject to design validation. |
| Later work | CI integration, additional clients, and tightly gated compilation or config writing. |

## Problem Statement

AI coding agents expose permissions through different configuration formats and enforcement models. A rule that appears similar across two clients may differ in resource scope, command matching, precedence, defaults, or enforcement timing. Teams therefore face five practical problems:

1. **Duplicated policy work:** The same intent must be expressed and maintained separately for each client.
2. **Unclear equivalence:** Similar-looking rules do not prove equivalent effective protection.
3. **Configuration drift:** Client policies evolve independently and may grant more access than intended.
4. **Hidden representational gaps:** Some policy intent cannot be represented by every target.
5. **False assurance:** A tool that silently drops or approximates semantics can make an unsafe configuration appear compliant.

The immediate product problem is not automatic policy generation. It is establishing a trustworthy, inspectable answer to: **What can each configured agent effectively do, how do those permissions differ, and where can equivalence not be proven?**

### Why Existing Categories Do Not Fully Resolve This Problem

The broader agent-security ecosystem includes MCP security tools, command guards, Git-aware protections, scanners, firewalls, and runtime controls. These products validate demand for agent security, but they generally address different layers or narrower fragments. The candidate product gap is a local normalization and audit layer for coding-agent permission semantics across clients.

This PRD does **not** claim that cross-client agent security is a novel category or that no adjacent solution exists. Competitive claims must remain evidence-based and be revisited before public positioning.

## Target Users

| Persona | Context | Primary need | Current pain |
|---|---|---|---|
| Security-conscious developer | Uses multiple coding agents locally or across projects. | Verify that each agent has only intended permissions. | Must inspect client-specific files and mentally reconcile semantics. |
| Platform or developer-experience engineer | Defines team defaults and CI safeguards. | Establish one reviewable policy intent and detect drift across repositories. | Maintains duplicated rules and cannot easily prove equivalent enforcement. |
| Security engineer | Reviews agent adoption and repository controls. | Obtain evidence of effective permissions, exceptions, and unsupported controls. | Configuration syntax obscures actual access and creates false confidence. |
| Open-source maintainer | Wants safe defaults for contributors using different agents. | Publish clear expectations without mandating one client. | Cannot express or verify a consistent baseline across tools. |

## Jobs to Be Done

- When I adopt or update an AI coding agent, I want to inspect its effective permissions so that I can identify unintended access before using it.
- When my team supports multiple coding agents, I want to compare their effective capabilities so that I can detect drift and review meaningful differences.
- When a canonical policy cannot be represented by a client, I want an explicit explanation so that I can choose a compensating control or reject the configuration.
- When client configuration formats change, I want the tool to identify unsupported versions or uncertain interpretations so that stale parsing does not produce false assurance.
- When I automate audits, I want deterministic machine-readable output and meaningful exit behavior so that CI can enforce reviewed policy expectations later.

## Product Principles

1. **Truth over convenience:** Report uncertainty and semantic mismatch instead of manufacturing equivalence.
2. **Read before write:** Earn trust through audit accuracy before introducing configuration mutation.
3. **Effective permissions over raw syntax:** Show what the resolved configuration permits, while retaining evidence back to source rules.
4. **Local by default:** Keep policies and configurations on the user's machine unless the user explicitly exports results.
5. **Deterministic and inspectable:** The same supported inputs and tool version must produce the same result.
6. **Adapters contain client complexity:** Client-specific parsing and semantics must not leak into the canonical policy contract without explicit modeling.
7. **Safe failure:** Unknown input, parser uncertainty, and partial analysis must be visible and must not result in a passing audit.

## Value Proposition

The product gives developers and teams a single local view of effective AI coding-agent permissions across clients. Unlike runtime guards or client-specific policy fragments, it focuses on normalization, cross-client comparison, drift detection, and explicit reporting of semantic loss. It complements rather than replaces client enforcement, runtime security controls, repository protections, or organizational security review.

## Scope

### MVP: Audit-Only Alpha

The MVP includes:

- Local discovery of supported OpenCode and Claude Code policy/configuration sources.
- Explicit path input when automatic discovery is undesirable or ambiguous.
- Version-aware parsing of supported configuration formats.
- A canonical model for capabilities, resources, effects, scope, conditions, provenance, and interpretation confidence.
- Calculation of effective permissions according to each supported client's documented precedence and defaults.
- Human-readable audit output with evidence linked to source files and rules.
- Cross-client comparison and drift reporting.
- Explicit unsupported, ambiguous, unknown, and lossy-semantic findings.
- Deterministic output suitable for review.
- Versioned machine-readable output.
- Documented exit codes for successful audits, policy findings, unsupported input, and operational failure.
- A proposed `audit` workflow and a proposed `diff` workflow.

### Explicit MVP Non-Goals

- Writing, rewriting, formatting, or deleting OpenCode or Claude Code configuration.
- Generating target configurations from a canonical policy.
- Claiming enforcement equivalence when client semantics differ or cannot be verified.
- A web UI, desktop UI, hosted service, account system, telemetry service, or cloud control plane.
- An LLM dependency or probabilistic policy interpretation.
- A runtime proxy, MCP firewall, shell interceptor, sandbox, or endpoint security agent.
- Replacing native client enforcement or proving runtime behavior beyond supported configuration semantics.
- Supporting Codex, Cursor, or every version and extension mechanism in the first release.
- Organization-wide policy distribution or fleet management.
- Automatic remediation of findings.

### Later Roadmap

| Stage | Outcome | Candidate scope | Entry condition |
|---|---|---|---|
| Alpha | Users can inspect permissions safely. | Read-only OpenCode and Claude Code audit, comparison, machine output. | Core adapters pass fixtures and semantic conformance tests. |
| Beta | Teams can enforce reviewed expectations in automation. | Stable output schema, CI examples/integration, baseline or expected-policy checks, broader version coverage. | Alpha false-assurance risks are understood; output and exits are stable enough for automation. |
| Expansion | More users can compare additional clients. | Codex, Cursor, or other adapters selected by demand and documented semantics. | Adapter contract is proven and target semantics can be modeled responsibly. |
| Guarded compilation preview | Users can preview target-specific policy generation. | Proposed `compile --dry-run`, explicit loss report, no mutation by default. | Audit and canonical model are mature; round-trip and loss tests meet release gates. |
| Guarded writing | Users may opt into controlled updates. | Explicit opt-in, dry-run review, backups, atomic replacement, preservation checks, rollback guidance. | Separate security review and product approval; config-preservation behavior is demonstrated. |

Roadmap stages are directional, not date commitments.

## Functional Requirements

### Discovery and Parsing

| ID | Requirement | Priority |
|---|---|---|
| FR-001 | The tool shall discover supported OpenCode and Claude Code configuration sources using documented client-specific search rules. | Must |
| FR-002 | The tool shall allow users to provide explicit configuration paths and shall report which sources were selected. | Must |
| FR-003 | The tool shall detect ambiguous discovery, duplicate sources, unreadable files, and conflicting explicit inputs rather than silently choosing. | Must |
| FR-004 | Each target adapter shall parse only declared configuration versions and constructs; unknown versions or constructs shall produce visible findings. | Must |
| FR-005 | Parsing diagnostics shall include target, source path, and location or rule identity when available, without exposing unrelated sensitive values. | Must |
| FR-006 | The tool shall preserve source provenance for every parsed permission represented in audit output. | Must |

### Canonical Capability Model

| ID | Requirement | Priority |
|---|---|---|
| FR-007 | The tool shall normalize supported client rules into a canonical intermediate model without discarding target-specific provenance. | Must |
| FR-008 | The canonical model shall represent, at minimum, capability/action, resource or subject, effect, scope, conditions, source provenance, target, and interpretation status. | Must |
| FR-009 | The model shall distinguish explicit allow, explicit deny, inherited or default behavior, unknown behavior, and unsupported semantics. | Must |
| FR-010 | The model shall support target-specific extensions when semantics cannot be safely generalized, while marking their portability status. | Must |
| FR-011 | The canonical model and its serialized machine representation shall be explicitly versioned. | Must |

### Effective Permissions

| ID | Requirement | Priority |
|---|---|---|
| FR-012 | Each adapter shall resolve effective permissions using the supported client's documented precedence, inheritance, and default behavior. | Must |
| FR-013 | Effective-permission output shall distinguish resolved grants and denials from unresolved or condition-dependent behavior. | Must |
| FR-014 | Each effective permission shall retain an explanation trace showing the source rules and defaults that contributed to it. | Must |
| FR-015 | If the tool cannot resolve precedence or defaults confidently, it shall mark the result incomplete and shall not present it as a successful equivalent audit. | Must |
| FR-016 | Audit output shall identify permissions that are broader than a supplied canonical expectation or comparison baseline when such an expectation is provided. | Should |

### Audit and Cross-Client Comparison

| ID | Requirement | Priority |
|---|---|---|
| FR-017 | The tool shall produce a per-client audit summarizing effective capabilities, denied capabilities, unresolved findings, and source evidence. | Must |
| FR-018 | The tool shall compare normalized effective permissions between OpenCode and Claude Code. | Must |
| FR-019 | Comparison shall classify differences as equivalent, broader, narrower, target-only, unsupported, ambiguous, or not comparable. | Must |
| FR-020 | The tool shall detect drift between current effective permissions and a user-supplied canonical policy or approved baseline when that artifact is available. | Should |
| FR-021 | Comparison output shall explain the semantic dimension responsible for each difference, such as scope, matcher behavior, default, precedence, condition, or enforcement limitation. | Must |
| FR-022 | The tool shall never classify policies as equivalent solely because their source text or normalized labels match. | Must |

### Unsupported and Lossy Semantics

| ID | Requirement | Priority |
|---|---|---|
| FR-023 | The tool shall report every detected unsupported, unknown, ambiguous, or lossy mapping as a first-class finding. | Must |
| FR-024 | Each loss finding shall identify the source semantic, affected target, reason, likely consequence, and whether comparison remains possible. | Must |
| FR-025 | Findings shall use stable categories and severity levels suitable for filtering and automation. | Must |
| FR-026 | Any unrecognized permission-bearing construct shall prevent an unqualified success result. | Must |
| FR-027 | The tool shall distinguish parser limitations from target-model limitations so users know whether the issue is tool support or client expressiveness. | Must |

### Output and Failure Behavior

| ID | Requirement | Priority |
|---|---|---|
| FR-028 | The CLI shall provide concise human-readable output with an optional detailed evidence view. | Must |
| FR-029 | The CLI shall provide versioned machine-readable output, initially JSON unless design validation identifies a stronger default. | Must |
| FR-030 | Machine-readable output shall include tool version, adapter versions, detected client/config versions, selected sources, analysis completeness, findings, and effective permissions. | Must |
| FR-031 | Output ordering, identifiers derived from inputs, and exit behavior shall be deterministic for the same supported inputs and tool version. | Must |
| FR-032 | The tool shall use documented, distinct exit outcomes for: complete audit with no blocking findings, completed audit with findings, incomplete/unsupported analysis, invalid input, and internal or operational failure. | Must |
| FR-033 | Human-readable output shall place blocking uncertainty before lower-priority informational findings. | Must |
| FR-034 | The tool shall support redaction of sensitive configuration values in diagnostics and output; secrets shall be redacted by default. | Must |
| FR-035 | The tool shall not modify discovered or explicitly supplied target configuration files in the MVP. | Must |

### Later Compilation Requirements

These requirements are intentionally outside MVP and exist to constrain future design.

| ID | Requirement | Release |
|---|---|---|
| FR-101 | A future compiler shall default to preview-only behavior and expose proposed changes through a dry run. | Guarded compilation preview |
| FR-102 | A future compiler shall fail before generation when required semantics are unsupported, unless the user explicitly accepts a documented approximation. | Guarded compilation preview |
| FR-103 | A future writer shall require explicit opt-in, create a recoverable backup, and use an atomic update strategy. | Guarded writing |
| FR-104 | A future writer shall preserve unsupported fields, comments, ordering, and formatting where the target format and parser permit; inability to preserve them shall block writing or require explicit acknowledgement. | Guarded writing |
| FR-105 | A future writer shall emit an exact change summary and loss report before mutation. | Guarded writing |

## Non-Functional Requirements

Performance figures below are initial targets for representative local repositories and may be refined after fixtures and benchmarks exist. They are not claims about current implementation.

| ID | Requirement | Initial target or constraint |
|---|---|---|
| NFR-001 | Local-only core operation | Core discovery, parsing, audit, comparison, and report generation shall work without network access. |
| NFR-002 | Data privacy | The tool shall not transmit policy, configuration, paths, or audit content. Any future update check or telemetry must be separately proposed, disabled by default, and explicit. |
| NFR-003 | Determinism | Identical supported inputs, explicit environment inputs, and tool version shall produce byte-stable machine output except for fields explicitly documented as variable; variable fields should be opt-in or separable. |
| NFR-004 | Security | Treat all configuration content as untrusted data; parsing shall not execute commands, evaluate arbitrary code, expand unsafe includes, or follow unbounded references. |
| NFR-005 | Least privilege | Auditing shall require only read access to selected configurations and shall not require elevated privileges. |
| NFR-006 | Secret handling | Known secret-bearing fields and values shall be redacted by default from terminal and machine-readable output. |
| NFR-007 | Performance | Audit two typical local client configurations in under 500 ms at p95 on a documented reference machine after startup; larger fixture sets shall have separately published benchmarks. |
| NFR-008 | Resource usage | For typical inputs, target peak memory below 100 MB; limits and behavior for unusually large inputs shall be documented. |
| NFR-009 | Compatibility reporting | Every run shall report or serialize the tool version, adapter version, detected config/client version when available, and support status. |
| NFR-010 | Forward safety | Unknown versions or permission-bearing fields shall produce incomplete or unsupported status rather than being ignored. |
| NFR-011 | Testability | Adapters shall support fixture-based parser tests, semantic conformance tests, precedence tests, malformed-input tests, and golden output tests. |
| NFR-012 | Cross-platform operation | Initial supported operating systems shall be documented before alpha; the intended target is common developer environments on Linux, macOS, and Windows where client config behavior can be verified. |
| NFR-013 | Distribution integrity | Release artifacts shall be checksummed; signing and provenance should be added before stable release. |
| NFR-014 | Accessibility of output | Human output shall not rely on color alone, and a no-color mode shall be available. |
| NFR-015 | Diagnosability | Internal errors shall include actionable context without leaking secrets; debug output shall be explicit and separately enabled. |

## Proposed CLI Experience

The following examples communicate intended workflows. Exact command names, flags, wording, and exit code numbers are **proposed** and must be validated during CLI design.

### Journey 1: Audit Discovered Configurations

```bash
agent-policy audit
```

Expected outcome:

```text
Audit incomplete: 2 clients analyzed, 1 blocking uncertainty

OpenCode      12 effective capabilities   0 unresolved
Claude Code   11 effective capabilities   1 unsupported rule

BLOCKING
  Claude Code rule "..." uses a matcher that cannot be compared safely.
  Result: cross-client equivalence was not established.

DIFFERENCES
  shell.execute: OpenCode scope is broader for path "..."

No files were modified.
```

The actual product output must redact sensitive values and provide source evidence without overwhelming the summary.

### Journey 2: Audit Explicit Sources with JSON Output

```bash
agent-policy audit \
  --opencode-config <path> \
  --claude-config <path> \
  --format json
```

Expected outcome: a versioned document containing selected sources, compatibility status, analysis completeness, effective permissions, findings, and evidence references.

### Journey 3: Compare Clients or Baselines

```bash
agent-policy diff --from <source-or-baseline> --to <source-or-baseline>
```

Expected outcome: differences classified by semantic impact rather than textual changes. The final interface may instead use target-specific flags if that produces a clearer mental model.

### Journey 4: Preview Compilation in a Later Release

```bash
agent-policy compile --target <client> --dry-run
```

Expected outcome: a proposed target configuration and explicit loss report, with no file mutation. This command is not part of MVP and must not ship until audit maturity gates are met.

## Acceptance Criteria

### MVP Release Acceptance

- [ ] AC-001: Given supported OpenCode and Claude Code fixtures, the tool discovers or accepts explicit sources and reports exactly which files were analyzed.
- [ ] AC-002: Given valid supported configurations, the tool produces effective permissions with source evidence and adapter/config compatibility metadata.
- [ ] AC-003: Given rules affected by precedence or defaults, effective output matches independently documented semantic fixtures for each client.
- [ ] AC-004: Given semantically different rules with similar labels, the comparison does not classify them as equivalent.
- [ ] AC-005: Given a capability broader in one client, the tool identifies the broader client and the semantic dimension causing the difference.
- [ ] AC-006: Given an unsupported or unrecognized permission-bearing construct, the run is marked incomplete or unsupported and cannot return an unqualified pass.
- [ ] AC-007: Given a lossy cross-client mapping, output identifies the lost semantic, affected target, consequence, and comparison status.
- [ ] AC-008: Given malformed, unreadable, or ambiguously discovered input, the tool fails safely with an actionable diagnostic and no file mutation.
- [ ] AC-009: Repeating a run with identical supported inputs, environment, and tool version produces byte-stable machine output under the documented determinism contract.
- [ ] AC-010: Machine output validates against its published schema and includes schema version, tool version, adapter versions, selected sources, completeness, findings, and effective permissions.
- [ ] AC-011: Secret-bearing fixture values do not appear in default human output, JSON output, or error diagnostics.
- [ ] AC-012: Tests verify that MVP commands never open target configuration files for writing and leave file content and metadata unchanged.
- [ ] AC-013: Human output remains understandable with color disabled and prioritizes blocking uncertainty.
- [ ] AC-014: Representative benchmark fixtures meet or document variance from the initial performance and memory targets on the published reference environment.
- [ ] AC-015: User-facing documentation states the supported client/config versions and clearly explains that audit results describe modeled configuration semantics, not guaranteed runtime enforcement.

### Guarded Compilation Entry Criteria

Compilation work may begin only after all of the following are true:

- [ ] The alpha adapter fixture suites cover supported permission constructs, defaults, and precedence behavior.
- [ ] Known unsupported constructs are classified and produce blocking or explicit outcomes.
- [ ] The canonical model has a versioning and migration strategy.
- [ ] Machine output and exit behavior have completed at least one compatibility review cycle.
- [ ] Maintainers have defined measurable round-trip, preservation, backup, and atomicity tests.
- [ ] A separate threat model and security review approve the proposed mutation boundary.

## Success Metrics

Early success should measure audit trust and utility, not raw popularity.

| Metric | Why it matters | Initial evaluation approach |
|---|---|---|
| Semantic fixture conformance | Demonstrates that adapters model declared client behavior correctly. | 100% pass rate for supported constructs before release; unsupported constructs must fail visibly. |
| Unsafe silent-degradation defects | Directly measures false-assurance risk. | Zero known cases released where a permission-bearing construct is ignored without an incomplete/unsupported result. |
| Actionable audit rate | Shows whether users receive a usable answer rather than parser noise. | Track opt-in issue templates or structured test cohorts; establish a baseline before setting a target. |
| Finding reproducibility | Supports review and CI trust. | Identical fixture inputs produce identical machine findings and ordering. |
| Time to first audit | Measures adoption friction for a local developer tool. | In usability testing, measure install-to-first-result; define a target after initial observations. |
| Maintainer response to upstream changes | Measures adapter sustainability. | Record time from verified upstream permission-format change to compatibility classification or adapter update. |
| Cross-client drift found and confirmed | Tests whether the core comparison solves real problems. | Count user-reported findings confirmed as meaningful, without treating volume alone as success. |

Repository stars, downloads, and social engagement may indicate reach but are secondary indicators, not primary evidence of product correctness or safety.

## Competitive Context and Differentiation

| Category | Examples | Primary focus | Relationship to this product |
|---|---|---|---|
| MCP and general agent security | `stacklok/toolhive`, `snyk/agent-scan`, `ThinkWatchProject/ThinkWatch`, `provnai/McpVanguard`, `ressl/mcp-firewall` | Scanning, isolation, firewalling, runtime or MCP security controls. | Adjacent and potentially complementary; validates broader agent-security needs but operates at a different layer. |
| Coding-agent command and Git guards | `shvmmshr/no-nuke`, `karlkfi/claude-branch-guard`, `sharayras/claude-git-guard`, `IliasAlmerekov/aegis-shellguard`, `askalf/redstamp`, `bifrost-mcp/claude-shield` | Prevent destructive commands, protect branches, add snapshots, rollback, or client-specific guardrails. | Addresses important enforcement fragments; may complement policy audit and reveal capabilities the canonical model should eventually describe. |
| Proposed product | This project | Normalize supported coding-agent policy semantics, resolve effective permissions, compare clients, detect drift, and report semantic loss. | Differentiates by audit scope and cross-client semantic evidence, not by claiming a new security category. |

### Positioning Guardrails

- Do not claim “the first,” “the only,” or “universal enforcement” without current, independently verifiable evidence.
- Describe the product as an audit and translation layer, not a replacement for native enforcement or runtime controls.
- Treat the identified market gap as a hypothesis until validated with users and ongoing competitive research.
- Reassess competitors before launch because agent-security tooling changes quickly.

## Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Semantic mismatch between clients | Policies may look equivalent while permitting different behavior. | Model semantic dimensions explicitly; require evidence; classify non-comparable behavior; never infer equivalence from syntax. |
| Upstream config or default changes | An adapter may silently become stale. | Declare supported versions, detect unknown constructs, maintain versioned fixtures, monitor upstream releases, and fail incomplete on uncertainty. |
| False assurance | Users may interpret an audit pass as a runtime security guarantee. | Use precise language, expose completeness, document limitations, distinguish modeled config semantics from runtime enforcement, and block unqualified success on unknowns. |
| Config preservation during future writing | Rewrites may remove comments, extensions, formatting, or user-managed values. | Keep writing out of MVP; require preservation tests, backups, atomic updates, dry-run, and explicit acknowledgement of unavoidable loss. |
| Scope explosion | Supporting every client and security layer could prevent a trustworthy release. | Limit alpha to two clients and audit-only behavior; add targets only through a stable adapter contract and demand evidence. |
| Canonical model becomes lowest common denominator | Important target-specific semantics may be erased. | Support typed target extensions, preserve provenance, and report portability instead of forcing all semantics into generic labels. |
| Client semantics are undocumented or context-dependent | Effective permissions may not be statically resolvable. | Mark conditional or unresolved results, test observable behavior where feasible, and avoid unsupported certainty. |
| Secret leakage in reports | Audits could expose credentials or sensitive paths. | Redact by default, minimize captured values, test secret fixtures, and require explicit opt-in for sensitive debug data. |
| Path and include traversal abuse | Malicious repositories could cause unintended reads or resource exhaustion. | Bound traversal, make include behavior explicit, avoid code execution, detect cycles, and document repository trust assumptions. |
| Determinism conflicts with environment-dependent config | Environment variables or machine state may alter effective permissions. | Enumerate relevant inputs, record non-secret environment provenance, support explicit overrides, and mark unresolved dependencies. |
| Adapter maintenance burden | Frequent upstream changes may exceed maintainer capacity. | Keep adapters isolated, publish a support matrix, automate fixtures, and prefer fewer trustworthy adapters over broad nominal support. |

## Release Strategy

### 1. Audit-Only Alpha

Focus on semantic correctness, visible uncertainty, and fixture quality. Support a narrow, documented set of OpenCode and Claude Code configuration versions. Make no config changes. Solicit structured reports around incorrect effective permissions, unsupported constructs, and confusing explanations.

### 2. Automation-Ready Beta

Stabilize the machine-readable schema and exit behavior. Add CI usage examples or an integration only after deterministic output and compatibility behavior are reliable. Introduce approved-baseline checks if user validation confirms the workflow.

### 3. Adapter Expansion

Evaluate Codex and Cursor based on user demand, public/documented semantics, and maintenance cost. Adding a client requires semantic fixtures and explicit unsupported behavior; parsing syntax alone is insufficient.

### 4. Guarded Compilation Preview

Introduce preview-only compilation behind explicit experimental labeling. Every preview must include a loss report. Unsupported required semantics block generation unless the user makes a deliberate, recorded choice to accept an approximation.

### 5. Guarded Config Writing

Consider file mutation only as a separately approved release. It must require explicit opt-in, show a dry run, preserve content where possible, create backups, update atomically, and provide recovery guidance. Writing must never become an implicit side effect of `audit` or `diff`.

## Dependencies and Constraints

- Reliable public documentation or verified fixtures for OpenCode and Claude Code configuration semantics.
- A maintained support matrix for client/config versions and adapter capabilities.
- A canonical-model versioning strategy before external machine-output stability is promised.
- Security review of parsing, path traversal, redaction, and any future write path.
- Cross-platform filesystem and configuration-discovery validation.
- Legal and licensing review for any reused schemas, fixtures, or examples before distribution.

## Assumptions

These are planning assumptions, not validated user-research findings:

- Developers and teams use more than one AI coding agent or need portable policy intent across contributors.
- OpenCode and Claude Code expose enough static configuration semantics to support useful local audits for a declared subset.
- Users will accept explicit “unknown” or “not comparable” results in exchange for avoiding false equivalence.
- A CLI and machine-readable output are sufficient for initial users; a graphical interface is not required for MVP validation.
- A single Go binary is a practical distribution target, but implementation design may revise this if parser or platform constraints justify it.
- Effective-permission analysis can provide value without intercepting runtime actions.
- CI demand exists but should be validated after local audit output is stable.

## Open Product Questions

1. What should the canonical policy authoring format be, and should MVP consume it directly or initially focus on comparing native configurations?
2. Which OpenCode and Claude Code configuration versions and permission constructs define the minimum useful alpha support matrix?
3. How should conditional permissions be represented when outcomes depend on interactive approval, environment, plugins, hooks, or runtime context?
4. What constitutes a blocking finding versus a warning for local interactive use and later CI use?
5. Should drift compare against a committed canonical policy, a generated baseline snapshot, another client, or all three?
6. What source-path information can safely appear in reports by default without leaking user or repository details?
7. Which exit-code categories provide enough automation control without becoming unstable or overly granular?
8. How should adapter semantic claims be validated when client documentation is incomplete: source inspection, executable conformance tests, maintainers' confirmation, or a combination?
9. What is the minimum evidence required before publicly describing two target policies as equivalent?
10. Should the product name emphasize audit, policy portability, or agent permissions, given that “universal” could overstate initial coverage?

## Review Checklist

- [ ] The MVP remains read-only and limited to OpenCode and Claude Code.
- [ ] Every permission-bearing unknown prevents an unqualified successful audit.
- [ ] Effective permissions retain evidence back to source rules and defaults.
- [ ] Cross-client comparison reports semantic differences rather than textual differences.
- [ ] Machine output and compatibility information are versioned and deterministic.
- [ ] Later compilation and writing are visibly separated from MVP.
- [ ] Public positioning does not claim category novelty or guaranteed runtime enforcement.
- [ ] Assumptions are labeled and no user research is presented as completed fact.
- [ ] Acceptance criteria are observable and testable.
- [ ] Success metrics prioritize correctness, safety, and utility over popularity.
