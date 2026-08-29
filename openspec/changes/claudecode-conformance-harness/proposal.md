# Proposal: Claude Code Conformance Harness

## Intent

Create a separate deterministic conformance harness for Claude Code permission behavior so maintainers can collect privacy-safe, version-bound evidence without changing production auditor behavior or bypassing the blocked `universal-agent-policy-auditor` change.

The harness output is an `evidence_candidate`. It is not a support claim, not an automatic `support.Matrix` update, and not a reason to mark Claude Code support complete without a maintainer decision.

## Problem

The main auditor change needs executable Claude Code evidence before maintainers can responsibly decide whether Claude Code permission semantics are supported. The current auditor is intentionally local, offline, and read-only, so running an installed Claude Code binary inside the normal audit path would violate that product boundary.

The missing capability is a controlled evidence path that can prove one narrow behavior without resetting, bypassing, or redefining the blocked main objective: when an allow rule overlaps a deny rule for a Bash matcher, Claude Code must request the target command as structured tool use, deny the same matcher, and leave a sentinel absent to prove the command did not execute.

## Goals

- Provide a deterministic harness design that is separate from production audit execution.
- Produce schema-versioned, byte-stable, redacted evidence records classified as `evidence_candidate`.
- Preserve the auditor's normal offline/read-only behavior.
- Keep the first implementation slice limited to deterministic internals and fixtures: schema, parser, redactor, and classifier.
- Define a later, maintainer-only path for workspace construction, command construction, and a single explicitly authorized paid/model execution.
- Keep each stacked PR slice within the 400-line authored review budget.

## Non-goals

- Do not implement production Claude Code support in this change.
- Do not mutate `support.Matrix` automatically.
- Do not run Claude Code, Claude, or any model process in the first implementation slice.
- Do not consume paid/model attempts without explicit maintainer authorization.
- Do not reset, bypass, or alter the blocked `universal-agent-policy-auditor` change.
- Do not mutate Git ledgers, repository settings, user settings, Git state, issues, or unrelated OpenSpec changes.
- Do not store raw transcripts, raw prompts, system prompts, absolute paths, usernames, hostnames, credentials, environment values, or repository source contents.

## User and maintainer outcome

Maintainers get a safe reviewable path from deterministic harness internals to an eventual authorized evidence run. The outcome should make it clear whether an observed Claude Code binary produced the required structured tool-use and denial behavior for the pinned scenario, while keeping the human decision about support metadata explicit.

Users of the normal auditor should see no behavior change. Audits remain offline/read-only and Claude Code remains unsupported-by-default until maintainers separately accept evidence and update support metadata.

## Scope boundaries

### In scope for the first slice

- Evidence schema for Claude Code conformance records.
- Structured event parser fixtures for tool-use, denial, malformed, and missing-event cases.
- Redaction rules and fixtures for private paths, usernames, hostnames, tokens, environment values, raw prompts, and raw transcripts.
- Classifier rules that map normalized evidence to `evidence_candidate`, failure, or inconclusive reason codes.
- Stable evidence encoding expectations suitable for golden tests.

### Deferred to later slices

- Isolated workspace and settings construction.
- Binary resolver, version capture, and SHA-256 hashing.
- Command builder and fake process runner seams.
- Maintainer-only CLI or integration gate.
- Preflight checks that do not invoke a model.
- One explicitly authorized, bounded conformance execution against an installed Claude Code binary.
- Any future proposal to add or reject a specific `support.Matrix` row based on collected evidence.

## Required evidence threshold

A future conformance execution can produce successful candidate evidence only when all of the following are true:

1. The Claude Code binary version and SHA-256 hash are captured.
2. The run uses isolated temporary HOME/config/project roots and does not read user or project Claude settings.
3. The scenario contains overlapping permission rules where the allow rule matches broadly and the deny rule matches the target Bash command.
4. Structured events contain the exact target Bash `tool_use` matcher.
5. Structured events contain the exact denial for the same matcher.
6. The sentinel remains absent, proving the denied command did not execute.
7. Output is schema-versioned, redacted, byte-stable, and contains no raw transcript or sensitive local context.
8. Cleanup succeeds, or retained evidence is written only to an explicit caller-selected evidence directory outside normal product execution.
9. The execution is single-attempt, bounded, and explicitly authorized.
10. The result is classified as `evidence_candidate`, never `supported`.

## Proposal-level acceptance

This proposal is accepted when it establishes a planning boundary that downstream specs can verify:

- The harness remains separate from production audit execution.
- The first implementation slice has no real Claude/model invocation and no maintainer gate that could execute paid inference.
- The evidence model captures version/hash, isolation, structured matcher, denial, sentinel absence, cleanup, and redaction guarantees.
- The relationship to the blocked `universal-agent-policy-auditor` change is explicit: the harness may inform a later maintainer decision but cannot bypass or complete that change by itself.
- Staged delivery uses stacked PRs to main, with each PR planned under the 400-line authored budget.

## Affected areas

- New conformance harness internals and fixtures, likely under a dedicated harness package such as `internal/conformance/claudecode` or equivalent support/conformance area.
- Test fixtures for structured Claude Code event parsing and redaction behavior.
- Future maintainer-only command or integration entry point, deferred beyond the first slice.

The normal auditor CLI, existing Claude Code adapter behavior, support matrix, runtime ledgers, Git state, issues, and the blocked `universal-agent-policy-auditor` change are intentionally unaffected by the first slice.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| Harness evidence is mistaken for a support claim. | Classify output only as `evidence_candidate` and require separate maintainer approval for support metadata. |
| Sensitive local context leaks into evidence. | Use strict redaction, do not store raw transcripts/prompts, and block persistence on redaction failure. |
| A paid/model run happens accidentally. | Keep the first slice deterministic only; later execution requires an explicit maintainer-only gate and one authorized attempt. |
| The harness changes production audit semantics. | Keep it separate from normal audit execution and do not add Claude runtime dependencies to the auditor path. |
| Evidence is flaky or hard to reproduce. | Require structured events, version/hash capture, stable JSON encoding, bounded output, and explicit reason codes for inconclusive results. |
| The blocked main change is bypassed. | Treat this harness as evidence collection only; do not reset, mutate, or mark tasks complete in the blocked change. |

## Rollback

Rollback for the first slice is to remove the new deterministic conformance schema/parser/redactor/classifier code and fixtures. Because the first slice must not alter production audit execution, support metadata, runtime ledgers, settings, or Git state, rollback should not affect existing auditor behavior.

For later slices, rollback boundaries must remain per work unit: remove workspace/command construction independently from any maintainer-only integration gate, and remove any collected evidence or support-matrix proposal only through a separate maintainer-reviewed change.

## Staged delivery

Delivery strategy is `auto-chain` with `stacked-to-main`. Each PR slice must stay within the 400-line authored budget and remain reviewable as an independent work unit.

| Slice | Scope | Runtime/model execution |
|---|---|---|
| 1 | Schema, parser, redactor, classifier, and deterministic fixtures. | None. |
| 2 | Workspace factory, settings writer, binary resolver/hasher, command builder, and fake runner seams. | None. |
| 3 | Optional maintainer-only CLI/integration gate and evidence writer. | Preflight only by default; no paid/model execution without explicit authorization. |
| 4 | Authorized evidence run and any later support-matrix proposal. | One explicit single-attempt run only, if maintainers approve. |

## Relationship to `universal-agent-policy-auditor`

The `universal-agent-policy-auditor` change remains blocked/partial for Claude Code support evidence. This harness is a separate work unit that may generate candidate evidence for maintainers to evaluate later. It must not alter the blocked change, redefine alpha support, mutate support metadata, or claim that Claude Code is supported.

The only intended dependency is informational: if a later authorized harness run produces sufficient evidence, maintainers may use that evidence in a separate decision about the main auditor support matrix.

## Success criteria

- Downstream specs can describe deterministic harness internals without requiring Claude/model execution.
- The first slice can be tested entirely with fixtures and fake inputs.
- Future execution criteria are explicit enough to prevent ambiguous support claims.
- Maintainers can distinguish `evidence_candidate` from `supported` in every proposed output path.
- Normal auditor behavior remains offline, read-only, and unchanged.
