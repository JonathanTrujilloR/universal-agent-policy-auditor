# Design: Claude Code Conformance Harness

## Decision summary

Create a deterministic Claude Code conformance package that is separate from normal auditor execution. The harness produces schema-versioned, byte-stable, redacted evidence records for maintainer review; it never claims Claude Code is `supported` and never mutates support metadata, runtime ledgers, Git state, issues, user/project settings, or `openspec/changes/universal-agent-policy-auditor/**`.

The first implementation work remains deterministic: evidence-domain types, classifier, parser, safe constructors, stable encoder, fixtures, and tests. Workspace construction, binary resolution, command construction, process execution, maintainer gates, evidence writing, and any support-matrix decision remain later, separately reviewed slices.

## Placement and dependency boundaries

| Area | Decision |
| --- | --- |
| Deterministic package | `internal/conformance/claudecode` in the current Go module; if nested later, use `packages/coding-agent/internal/conformance/claudecode`. |
| Normal auditor path | `internal/adapter/claudecode`, `internal/source`, `support`, and normal CLI entry points MUST NOT import the conformance package. |
| Dependency direction | Future maintainer-only commands may import conformance internals; production auditor packages remain independent. |
| Support metadata | Do not change `support/matrix.go`, `support/matrix.json`, `support/evidence.json`, `support/fixtures.json`, or existing Claude Code adapter behavior. |
| Runtime dependencies | Deterministic slices use fixtures/fakes only and expose no reachable binary, workspace, command, model, or process-runner integration. |

## Deterministic module shape

| File | Responsibility |
| --- | --- |
| `internal/conformance/claudecode/types.go` | Closed result, reason-code, provenance, sentinel, event, scenario, bounds, redaction, and evidence-record types. |
| `internal/conformance/claudecode/classifier.go` | Total decision table from normalized facts, provenance, and runtime proofs to `evidence_candidate`, `fail`, or `inconclusive`. |
| `internal/conformance/claudecode/parser.go` | Bounded JSONL parser that emits normalized facts, parse failures, or unsupported-event facts, never raw transcript. |
| `internal/conformance/claudecode/safety.go` | Safe constructors for scenario IDs, matcher template IDs, event IDs, digests, and persistable tokens. |
| `internal/conformance/claudecode/encoder.go` | Deterministic schema-versioned JSON encoder for golden tests. |
| `internal/conformance/claudecode/*_test.go` | Table-driven tests kept with the behavior under review. |
| `internal/conformance/claudecode/testdata/*.jsonl` | Minimal structured event fixtures for matching, malformed, missing, unsupported, mismatch, duplicate, contradictory, display-only, and denial-without-tool-use cases. |
| `internal/conformance/claudecode/testdata/*.golden.json` | Stable evidence JSON expectations for deterministic records only. |

The package should stay cohesive and stdlib-only unless a later implementation proves a narrow dependency is necessary.

## Data flow

```text
fixture JSONL bytes or future bounded process event bytes
  -> ParseEvents(bounds)
  -> NormalizedFacts plus parse/unsupported-event status
  -> Safe constructors and structural privacy checks
  -> Classify(scenario, facts, sentinel, provenance, runtime proofs)
  -> EvidenceRecord built only from safe types
  -> EncodeStableJSON
  -> golden comparison or later explicit evidence output
```

Raw prompt, raw transcript, absolute path, hostname, username, token, environment value, repository source content, and arbitrary caller-supplied diagnostic text are never copied into normalized facts, classifications, evidence records, public errors, diagnostics, or goldens.

## Provenance and candidate-evidence gate

| Mode | Meaning | Candidate evidence allowed |
| --- | --- | --- |
| `fixture` | Checked-in JSONL/golden input used by deterministic tests. | No. |
| `fake` | In-memory fake process/runtime used by tests. | No. |
| `preflight` | Later deterministic local validation that does not run the conformance prompt. | No. |
| `authorized_execution` | Later maintainer-authorized, isolated, single-attempt Claude Code execution. | Yes, only when every threshold passes. |

The classifier may return `evidence_candidate` only when all gates are true: provenance is `authorized_execution`; attempt count is exactly 1; explicit maintainer authorization exists; binary version and SHA-256 are captured as safe tokens; HOME/config/settings/project roots are isolated and user/repository settings are not read; cleanup succeeds or retained evidence is caller-selected outside normal product execution; structured events contain the exact target `tool_use` and exact denial correlated by stable tool-use ID and canonical matcher; sentinel inspection proves absence; no successful target execution, malformed input, unsupported event, unknown event, duplicate ambiguity, privacy failure, or output-bound failure exists; encoding is schema-versioned, byte-stable, bounded, and safe.

Fixture, fake, and preflight records can only produce test-only diagnostics, `inconclusive`, or `fail`; their positive facts must still use provenance-specific not-candidate reasons.

## Core classification contracts

`Result` is closed: `evidence_candidate`, `fail`, `inconclusive`. Every classification returns exactly one primary reason and may include sorted secondary reasons that cannot soften the primary result.

| Reason code | Result | Meaning |
| --- | --- | --- |
| `evidence_candidate` | `evidence_candidate` | All authorized-execution thresholds passed. |
| `fixture_not_candidate_evidence` | `inconclusive` | Fixture provenance cannot be candidate evidence. |
| `fake_not_candidate_evidence` | `inconclusive` | Fake provenance cannot be candidate evidence. |
| `preflight_not_candidate_evidence` | `inconclusive` | Preflight provenance cannot be candidate evidence. |
| `parser_malformed_event` | `inconclusive` | A structured event cannot be decoded safely. |
| `parser_no_structured_events` | `inconclusive` | Input contains no accepted structured events. |
| `parser_bounds_exceeded` | `inconclusive` | Event count, byte, line, or field bounds were exceeded. |
| `unsupported_structured_event` | `inconclusive` | An unknown or unsupported structured event may affect execution semantics. |
| `display_only_event` | `inconclusive` | Evidence appears only in free-form display text. |
| `missing_tool_use` | `inconclusive` | Target Bash tool use was not observed. |
| `denial_without_tool_use` | `inconclusive` | A denial exists without correlated target tool use. |
| `missing_denial` | `fail` | Target tool use exists but no correlated denial exists. |
| `matcher_correlation_failed` | `fail` | Tool use and denial reference different canonical matchers. |
| `tool_use_id_correlation_failed` | `fail` | Denial/result references a different stable tool-use ID. |
| `duplicate_target_event` | `inconclusive` | Multiple distinct target events make correlation ambiguous. |
| `contradictory_execution_success` | `fail` | Events show successful target execution despite denial. |
| `sentinel_present` | `fail` | Sentinel exists, proving execution or side effects. |
| `sentinel_unknown` | `inconclusive` | Sentinel absence cannot be proven. |
| `redaction_failed` | `inconclusive` | Sensitive content may remain in persistable evidence. |
| `runtime_identity_missing` | `inconclusive` | Executable evidence lacks binary version or hash. |
| `runtime_isolation_unproven` | `inconclusive` | Isolated runtime roots cannot be proven. |
| `execution_authorization_missing` | `inconclusive` | Explicit maintainer authorization is missing. |
| `execution_attempt_count_invalid` | `inconclusive` | Execution was not exactly one authorized attempt. |
| `preflight_binary_missing` | `inconclusive` | Claude Code cannot be resolved in preflight. |
| `preflight_version_unavailable` | `inconclusive` | Version cannot be captured. |
| `preflight_hash_failed` | `inconclusive` | Binary hashing failed. |
| `preflight_invalid_settings` | `inconclusive` | Generated settings are invalid. |
| `preflight_invalid_flags` | `inconclusive` | Command shape is unsupported. |
| `execution_timeout` | `inconclusive` | Runtime bounds were exceeded. |
| `execution_output_bounds_exceeded` | `inconclusive` | Output or artifact bounds were exceeded. |
| `cleanup_failed` | `inconclusive` | Cleanup failed or cannot be proven. |
| `unsupported_support_claim` | `fail` | Any output attempts to classify Claude Code as `supported`. |

Decision ordering is fail-safe: parser/bounds/unsupported/privacy problems block candidate evidence; missing tool use is inconclusive; missing denial after target tool use is fail; matcher or stable-ID mismatch is fail; duplicate target facts are inconclusive; any successful target result is fail regardless of event order; sentinel present is fail; sentinel unknown is inconclusive; authorized execution missing identity, isolation, cleanup, authorization, or one-attempt proof is inconclusive.

## Sentinel and event correlation

Sentinel state is closed tri-state: `SentinelAbsent`, `SentinelPresent`, `SentinelUnknown`. The zero value MUST decode to `SentinelUnknown`, not absence. Evidence stores only state plus a closed source enum such as `fixture`, `fake`, or future `isolated_workspace`, never a raw path. Sentinel absence is required but never sufficient.

The parser accepts only closed fact kinds: `tool_use`, `permission_denial`, `tool_result`, `known_diagnostic`, and `unsupported_structured_event`. Free-form display text is never evidence. Unknown structured objects are unsafe by default unless whitelisted as non-execution diagnostics.

Correlation rules are exact: compare closed scenario and matcher template IDs, not display text; map future literal matchers only from exact approved literals; require byte-identical stable tool-use IDs after JSON string decoding; never use substring, glob, case-folded, fuzzy, order-only, or prose matching. Future executable runs use a relative harmless sentinel command so the matcher template contains no private absolute path.

Duplicate handling deduplicates only identical event ID, kind, matcher template ID, tool-use ID, and sequence source. Distinct duplicate attempts, denials, or successes are ambiguous or contradictory and cannot pass.

## Structural privacy model

Persistable evidence is built from safe normalized facts and closed domains, not redacted arbitrary strings.

| Category | Allowed representation |
| --- | --- |
| Scenario/matchers | Closed scenario ID and matcher template ID. |
| Event references | Safe stable IDs or bounded SHA-256 digests. |
| Binary identity | Safe version token plus `sha256:<hex>` digest. |
| Runtime locations | Closed labels like `workspace_temp`, `home_temp`, `config_temp`, `settings_temp`, or opaque bounded digests. |
| Execution facts | Enums, booleans, counts, and bounded durations. |
| Diagnostics | Closed reason codes, field names, line numbers, event counts, byte counts, and bounded digests. |

Constructors MUST reject raw prompts, raw paths, raw environment, raw source, raw transcripts, usernames, hostnames, credentials, tokens, env-like strings, and unbounded diagnostics. Use private fields plus constructors such as `NewScenarioID`, `NewMatcherTemplateID`, `NewSafeToken`, `NewDigest`, and `NewEvidenceRecord`.

Raw event bytes may exist only at parser input boundaries and local decode variables. Public errors return reason codes, fields, line numbers, and counts only. Persistable structs cannot hold raw event bytes, source excerpts, paths, environment values, prompts, or transcript text.

## Bounds and encoding

| Bound | Initial value | Failure reason |
| --- | ---: | --- |
| JSONL input bytes | 64 KiB | `parser_bounds_exceeded` |
| Event count | 128 | `parser_bounds_exceeded` |
| Single line bytes | 8 KiB | `parser_bounds_exceeded` |
| Persistable safe token bytes | 128 | `redaction_failed` or `parser_bounds_exceeded` |
| Persistable digest bytes | 71 bytes for `sha256:<hex>` | `redaction_failed` |
| Encoded evidence bytes | 16 KiB | `execution_output_bounds_exceeded` |

Schema id is `claudecode-conformance-evidence/v1`. Use struct-backed JSON, not maps, for persisted records; sort dynamic slices; use `json.Encoder` with `SetEscapeHTML(false)`, two-space indentation for checked-in goldens, and a trailing newline. Later executable slices may lower process-output bounds but must not raise bounds without tests and OpenSpec rationale.

## Deferred executable seams

Do not make these seams reachable in deterministic PRs 1-3. Add them only in later work units with tests and fakes.

| Seam | Future responsibility | Earliest code PR |
| --- | --- | ---: |
| `BinaryResolver` | Locate explicit/PATH Claude Code binary without user settings. | 4 |
| `BinaryHasher` | Compute SHA-256 for version-bound evidence. | 4 |
| `WorkspaceFactory` | Create isolated HOME, config, project, settings, and sentinel roots. | 4 |
| `SettingsWriter` | Write minimal deny-over-allow settings under isolated workspace. | 4 |
| `CommandBuilder` | Build argv/env/cwd/stdin without shell interpolation. | 4 |
| `ProcessRunner` | Execute one bounded process only behind an explicit later gate. | 5 |
| `EvidenceWriter` | Persist redacted evidence to caller-selected output outside normal auditor execution. | 5 |

## Testing architecture

Follow strict TDD with small table-driven tests and deterministic goldens. Required coverage: provenance gate, authorized candidate gate, parser valid/malformed/display-only/oversized/unsupported cases, exact matcher and ID correlation, denial without tool use, ID mismatch, absent-ID limitation, contradictory success, sentinel absent/present/unknown/default, one row per reachable primary reason, structural privacy rejections, golden update followed by clean rerun, shuffled-input determinism, panic safety, and output bounds.

No deterministic PR 1-3 test may invoke Claude Code, Claude, a model process, Git ledger mutation, issue mutation, user settings, repository settings, or a real external process.

## File-change contract

Deterministic PRs 1-3 may add only `internal/conformance/claudecode/**`. They must not edit `internal/adapter/claudecode/**`, `internal/adapter/opencode/**`, `internal/source/**`, `support/**`, `openspec/changes/universal-agent-policy-auditor/**`, `.git/**`, runtime ledgers, Git state, issues, or user/project settings. If implementation needs existing production packages, stop and replan before coding.

## Delivery accounting update

The Review Workload Guard counts authored additions plus deletions for every PR, including OpenSpec artifacts. Generated goldens and the later candidate-evidence artifact may be excluded from authored estimates only when deterministic/generated, reviewed separately, and still included in snapshot identity.

Current untracked OpenSpec planning artifacts are authored review lines and MUST land before code:

| Artifact | Exact path | Lines |
| --- | --- | ---: |
| Exploration | `openspec/changes/claudecode-conformance-harness/exploration.md` | 183 |
| Proposal | `openspec/changes/claudecode-conformance-harness/proposal.md` | 133 |
| Spec | `openspec/changes/claudecode-conformance-harness/specs/claudecode-conformance/spec.md` | 246 |
| Design | `openspec/changes/claudecode-conformance-harness/design.md` | 229 |
| Tasks | `openspec/changes/claudecode-conformance-harness/tasks.md` | 106 |
| Total planning artifacts | All OpenSpec planning artifacts above | 897 |

Future `sdd-tasks` and `sdd-apply` phases MUST account for untracked or uncommitted OpenSpec authored lines before selecting a code slice. Code PRs MUST NOT silently include planning artifacts; before code PR 1, the planning chain below must be landed or otherwise removed from the code diff.

## Planning-artifact delivery chain

Delivery strategy is `auto-chain`; chain strategy is `stacked-to-main`; hard authored budget is 400 lines per PR; no `size:exception` exists. Planning PRs use `Refs #N` for the approved issue only, never a closing keyword; if no issue is approved, omit issue references and do not mutate issues.

| PR | Dependencies | Paths | Current lines and limit | Checks | Rollback |
| --- | --- | --- | --- | --- | --- |
| Planning PR A | None | `openspec/changes/claudecode-conformance-harness/exploration.md`, `openspec/changes/claudecode-conformance-harness/proposal.md` | 183 + 133 = 316 authored, limit <=400 | Markdown structural readback; Markdown/LSP diagnostics if available; root `go test ./...` only if repository policy requires no-regression evidence for docs-only changes. | Revert those two files; no runtime, code, ledger, issue, or support behavior changes. |
| Planning PR B | Planning PR A | `openspec/changes/claudecode-conformance-harness/design.md` | 229 authored, limit <=400 | Markdown structural readback; verify final design line count <=390; Markdown/LSP diagnostics if available; root `go test ./...` only if required. | Revert `design.md` to the previous planning state; no code/runtime behavior exists. |
| Planning PR C | Planning PR B | `openspec/changes/claudecode-conformance-harness/specs/claudecode-conformance/spec.md`, `openspec/changes/claudecode-conformance-harness/tasks.md` | 246 + 106 = 352 authored, limit <=400 | Markdown structural readback; verify tasks forecast includes OpenSpec accounting; Markdown/LSP diagnostics if available; root `go test ./...` where repository gate requires it before code. | Revert spec/tasks only; code PR 1 remains blocked until C lands. |

## Implementation rollout, dependencies, and rollback

Keep tests with the behavior they verify. The `Dependencies` column is normative; row order mirrors stacked-to-main delivery order. Code PR 1 explicitly depends on Planning PR C.

| PR | Dependencies | Authored-line estimate | Package/file boundary | Included tests/checks | Rollback boundary |
| --- | --- | ---: | --- | --- | --- |
| Code PR 1 - Domain types and classifier only | Planning PR C | 180-280 | `internal/conformance/claudecode/types.go`, `internal/conformance/claudecode/classifier.go`, `internal/conformance/claudecode/classifier_test.go` | Provenance gate, result enum, reason-code domain, sentinel states, authorized-threshold table; `go test ./internal/conformance/claudecode`; `go test ./...`; runtime harness `N/A`. | Remove initial package files; no parser, runtime, auditor, support, or ledger behavior exists. |
| Code PR 2 - Structured event parser and JSONL fixtures | Code PR 1 | 240-380 | `parser.go`, `parser_test.go`, `testdata/events_*.jsonl` under `internal/conformance/claudecode/` | Parser, unsupported/malformed/bounds, exact matcher template, stable ID normalization; package and root Go tests; runtime `N/A`. | Remove parser files and event fixtures while keeping direct-fact classifier usable. |
| Code PR 3 - Safety constructors and stable encoder | Code PR 2 | 240-390 excluding deterministic generated goldens reviewed separately | `safety.go`, `encoder.go`, `safety_test.go`, `encoder_test.go`, listed `testdata/*.golden.json` | Structural privacy, stable JSON, deterministic ordering, golden readback with explicit update then clean rerun; package and root tests; runtime `N/A`. | Remove safety/encoder/goldens; prior parser/classifier remains or this PR is reverted alone. |
| Code PR 4 - Runtime identity and execution planning without execution | Code PR 3 | 260-380 | `runtime_identity.go`, `workspace_plan.go`, `command_plan.go` and tests under `internal/conformance/claudecode/` | Binary version/hash value objects, isolated workspace plan, settings/argv plan validation, no process execution; package and root tests; runtime `N/A`. | Remove executable-preparation files without touching deterministic evidence semantics. |
| Code PR 5 - Preflight and evidence writer behind maintainer gates | Code PR 4 | 260-390 | `preflight.go`, `evidence_writer.go` and tests; optional after maintainer decision only: `cmd/claudecode-conformance/main.go`, `cmd/claudecode-conformance/main_test.go` | Preflight status, explicit output writer, optional wrapper validation, no model/process scenario; package/root tests; runtime `N/A - preflight only`. | Remove preflight/writer and optional wrapper files only. |
| Code PR 6 - Authorized evidence artifact and separate support-decision proposal | Code PR 5 plus explicit maintainer authorization | 160-320 authored; generated candidate evidence reviewed separately | `support/evidence/claudecode/bash-deny-overlapping-allow/evidence.candidate.json`, `openspec/changes/claudecode-support-matrix-decision/proposal.md`, `openspec/changes/claudecode-support-matrix-decision/specs/support-matrix/spec.md`; never support metadata files. | One authorized bounded run only if approved, schema/redaction/version/hash/isolation/cleanup/attempt/correlation validation, package/root tests, maintainer support-decision review. | Revert candidate evidence separately from decision proposal/spec; normal auditor and support metadata remain unchanged. |

PRs 1-5 use `Refs #N`; only final Code PR 6 may use a closing keyword after the approved work-unit scope is actually complete. Every PR body must include chain context, dependency diagram marking the current PR, exact start/end, out-of-scope items, changed-line count, verification evidence, and rollback boundary.

## Pitfall checklist

- Fixture, fake, and preflight records are structurally incapable of `evidence_candidate`.
- `evidence_candidate` requires authorized execution, attempt count 1, binary version/hash, isolation proof, cleanup proof, exact tool-use/denial correlation, sentinel absence, no target success, no unsupported events, and privacy-safe encoding.
- Unknown/unsupported structured events, display text, model prose, duplicate ambiguity, and contradictory success never silently pass.
- Exact correlation uses stable tool-use ID plus closed matcher template ID; arbitrary observed matcher text is not persisted.
- Raw transcript and prompt bytes are transient parser inputs only and never enter persistable records or public errors.
- Persisted JSON avoids maps or sorts all dynamic inputs before encoding.
- The result vocabulary never includes `supported`.
- No production support metadata, ledger, real process, Git state, issue, setting, or blocked-main-change mutation occurs in deterministic slices.

## Verification plan

Run focused package tests first, then the root suite when code exists:

```bash
go test ./internal/conformance/claudecode
go test ./...
```

Planning PR verification is Markdown structural readback plus Markdown/LSP diagnostics when available; root Go tests are applicable only as repository no-regression evidence, not because OpenSpec files form a Go package. Do not run tests from an OpenSpec subdirectory as a separate Go package.
