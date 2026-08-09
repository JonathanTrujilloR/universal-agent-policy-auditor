# Proposal: Audit Native Agent Policies Safely

## Intent

Enable users to compare the modeled effective permissions of supported OpenCode and Claude Code configurations without modification or overstated runtime guarantees.

## Scope

### In Scope
- Read-only selection or discovery of supported native configuration sources.
- Version-aware modeling of supported scope, precedence, defaults, matchers, and permission effects with source provenance.
- Per-client audit and cross-client comparison with deterministic evidence.

### Out of Scope
- An authored canonical policy DSL, compilation, remediation, or file mutation.
- Runtime interception/enforcement, CI enforcement, baselines as policy, or guaranteed runtime behavior.
- Additional clients beyond OpenCode and Claude Code.

## Capabilities

### New Capabilities
- `native-policy-audit`: Read-only native configuration resolution, evidence-backed permission reporting, and bounded cross-client comparison.

### Modified Capabilities
None.

## Approach

Use client-specific semantic adapters and a versioned comparison model that preserves target-specific evidence. Keep modeled static configuration semantics distinct from native runtime enforcement, interactive approvals, trust, hooks, plugins, sandboxing, and other unresolved runtime context.

**Support-matrix gate:** no version or construct is supported until current official documentation plus upstream source or executable evidence establish discovery/merge order, matcher behavior, defaults, and malformed-input handling. Each matrix entry requires conformance fixtures; everything else remains unsupported.

Unknown, unsupported, ambiguous, permission-bearing unrecognized, or lossy behavior MUST produce incomplete or not-comparable status and block unqualified success. Auditing MUST remain local, non-executing, read-only, deterministic for declared inputs, and secret-redacted by default.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `openspec/specs/native-policy-audit/spec.md` | New | Behavioral contract for audit, comparison, evidence, and safe failure. |
| OpenCode/Claude Code adapters | New | Isolated native semantic interpretation; concrete paths deferred to design. |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| False equivalence from semantic drift | High | Narrow support matrix, evidence gates, blocking uncertainty. |
| Sensitive evidence leakage | Medium | Data minimization and default redaction fixtures. |
| Static results mistaken for enforcement | High | Completeness and runtime-limit labels in every output. |

## Rollback Plan

Remove the alpha capability and its artifacts; no target configuration requires restoration because the feature cannot mutate files.

## Dependencies

- Current official OpenCode and Claude Code documentation, upstream source, and independently maintained conformance fixtures.

## Success Criteria

- [ ] 100% conformance for every declared support-matrix construct and zero silent permission-bearing omissions.
- [ ] Every uncertainty or lossy mapping blocks unqualified success and identifies its evidence and semantic dimension.
- [ ] Repeated supported inputs produce byte-stable machine output; read-only and default-redaction tests pass.
- [ ] User-facing output identifies analyzed sources, completeness, versions, and that results model configuration rather than prove runtime enforcement.
