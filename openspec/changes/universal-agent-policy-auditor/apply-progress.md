# Apply Progress: Universal Agent Policy Auditor

## Work Unit 1: Source Foundation

- Status: complete
- Delivery: auto-chain, stacked-to-main; this uncommitted slice is bounded to bootstrap and `internal/source/`.
- Git policy: repository was already initialized locally on `master`; no remote, branch rename, commit, push, or PR was created.
- Quality decisions: use Go's built-in `gofmt`, `go vet`, and `go test`; CI and an external linter are deferred until a runnable CLI exists.

## Completed Tasks

- [x] 1.1 Bootstrap Go module and quality commands.
- [x] 1.2 Add RED-first source-executor behavior tests.
- [x] 1.3 Implement bounded read-only source execution.

## Work Unit Evidence

| Evidence | Result |
|---|---|
| Focused test | `go test ./internal/source/...` — PASS (`ok .../internal/source`) |
| Runtime harness | N/A — no CLI/runtime boundary exists in this slice; `cmd/auditor` is intentionally deferred. |
| Rollback boundary | Remove `go.mod`, `internal/source/`, and these three task marks; no target configuration was modified. |

## Safety Evidence

- `ReadFS` exposes only read, stat, and symlink-resolution operations; no write, process, or network capability is accepted.
- Tests cover opaque malformed content, traversal, symlink escape/cycle, bounded bytes, duplicate identity, cancellation, allowlisted environment capture, and secret-free findings.
- No source content is evaluated; selected bytes are copied and findings contain generic messages plus basenames only.

## Verification

- `gofmt -l internal/source` — no output.
- `go vet ./...` — PASS.
- `go test ./...` — PASS.
- `go build ./cmd/auditor` — N/A; the CLI is intentionally deferred to task 4.3.

## Remaining

Tasks 2.1–5.2 remain unchecked. The next autonomous slice is Work Unit 2: model, support matrix, evidence manifests, and fixtures.

## Correction Rerun: Source Foundation

- Gate corrections add direct tests for environment allowlist exclusion and the `MaxFiles` limit; implementation behavior required no change.
- `git diff --check` — exit 0, no output.

### Exact Changed Files

- `go.mod` — created Go module.
- `internal/source/executor.go` — created read-only bounded executor.
- `internal/source/executor_test.go` — created source tests; corrected with two direct boundary tests.
- `openspec/changes/universal-agent-policy-auditor/tasks.md` — only tasks 1.1–1.3 marked complete.
- `openspec/changes/universal-agent-policy-auditor/apply-progress.md` — created and corrected progress evidence.

### Authored-Line Accounting

- Initial slice: 320 lines (314 new-file lines plus 3 additions and 3 deletions in task checkboxes).
- Correction: 47 lines (28 test lines plus 19 progress-evidence lines).
- Cumulative source-foundation slice: 367 authored changed lines; correction stays below its 120-line limit.

## Work Unit 2: Model Evidence

- Status: complete. Delivery remains auto-chain, stacked-to-main; this uncommitted slice is only `internal/model/`, `support/`, and its three task marks.
- Native original accounting: 329/400 authored changed lines — independently within its native limit.
- Supported clients: none. The empty support matrix is an evidence gate, not an assertion of OpenCode or Claude Code support.

### Completed Tasks

- [x] 2.1 Research and record documentation/source evidence plus an empty support matrix and fail-closed fixtures.
- [x] 2.2 Add immutable permission, evidence/provenance, finding, completeness, extension, and resolution-trace types.
- [x] 2.3 Validate malformed, unknown, unsupported, missing-evidence, default, matcher, and condition metadata as incomplete.

### TDD Cycle Evidence

| Task | RED | GREEN | REFACTOR |
|---|---|---|---|
| 2.1 | `matrix_test.go` failed: package/API absent | focused tests PASS | clean |
| 2.2 | `model_test.go` failed: constructors/types undefined; typed provenance initially undefined | focused tests PASS | copied accessors |
| 2.3 | table cases failed: validator absent | focused tests PASS | table-driven cases |

### Work Unit Evidence

| Evidence | Result |
|---|---|
| Focused test | `go test ./internal/model/... ./support/...` — PASS |
| Runtime harness | N/A — library-only slice; adapters and CLI are deferred. |
| Rollback boundary | Remove `internal/model/`, `support/`, this section, and task marks; no target configuration was read or modified. |

### Sources and limits

- Authoritative docs and upstream source links are recorded in `support/evidence.md`; no licensed code was copied.
- No process, network client, filesystem-write API, adapter, or canonical policy DSL was added. Documentation fetches were research-only; production code is deterministic and local.

## Correction Rerun: Model Evidence

- Native correction accounting: 138/140 authored changed lines — independently within its native limit.
- Gate failure corrected: validation now resolves the canonical matrix row by target/version/construct and ignores candidate fixture, evidence, and semantic-strength fields.
- `support.Load` executes the checked-in `matrix.json`, `fixtures.json`, and `evidence.json` registries; all remain empty, so zero client versions are supported.
- RED: `go test ./support/... -run 'TestValidateUsesCanonicalRowAndRegistry|TestLoadConsumesConfiguredZeroSupportRegistry'` failed for undefined `Registry` and `Load`; GREEN: `go test ./support/...` PASS.
- Direct cases reject forged candidate metadata, absent fixture ID, absent evidence ID, and incomplete canonical semantics.
- Verification: `gofmt -l .` clean; `go vet ./...`, `go test ./...`, `go test -cover ./...`, `go build ./...`, and `git diff --check` PASS.
- Correction rollback: revert `support/matrix.go`, `support/matrix_test.go`, the three support registries/docs, and this section. No process, network client, write API, adapter, or CLI was introduced.

## Work Unit 3 Slice 1: OpenCode Adapter Foundation

- Status: partial; independently reviewable adapter-foundation slice only. Task 3.1 remains unchecked because support-matrix-backed conformance fixtures and full OpenCode support acceptance are not complete.
- Branch/base: `feat/opencode-adapter-wu3-pr4` from updated `origin/main` merge commit `3d69d1b6ff1963226134fd962b1dea8f51d760e4` (PR #4 merge).
- Delivery: auto-chain, stacked-to-main. Current PR boundary is `internal/adapter/opencode/` plus this progress evidence; Claude Code and comparison/output work are explicitly out of scope.
- Runtime attempt token supplied by parent: `sha256:ab76282168547f4bb4c38bbb24364be389662b080269f8459a4d24806b8ecea4`; parent owns settlement.

### Completed Scope in This Slice

- Added `internal/adapter/opencode/` with a declarative bounded discovery plan, data-only JSON parser, last-source OpenCode rule precedence, adapter-local report, immutable model permissions, redacted provenance, target extension metadata, and conditional/static-runtime trace limitations.
- Added fail-closed diagnostics for malformed OpenCode JSON, unknown effects, unknown permission objects, and plural `permissions` permission-bearing constructs that are not modeled.
- Preserved read-only boundaries: the adapter accepts `source.SelectedSource` bytes, has no write/process/network interface, and delegates filesystem selection to `internal/source` plans.

### TDD Cycle Evidence

| Task / behavior | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|
| OpenCode adapter foundation | `go test ./internal/adapter/opencode/...` failed for undefined `DiscoveryPlan` and `Resolve`. | Implemented `opencode.go`; focused tests passed. | Added plural `permissions` unknown-construct case; focused test failed with complete/no findings. | Added top-level permission-bearing unknown-construct detection; `gofmt` clean and focused tests passed. |

### Work Unit Evidence

| Evidence | Result |
|---|---|
| Focused RED | `go test ./internal/adapter/opencode/...` — FAIL as expected before implementation (`undefined: DiscoveryPlan`, `undefined: Resolve`). |
| Focused GREEN | `go test ./internal/adapter/opencode/...` — PASS after initial implementation. |
| Triangulation RED | `go test ./internal/adapter/opencode/...` — FAIL for `permission-bearing unknown top-level` returning complete/no findings. |
| Focused final | `go test ./internal/adapter/opencode/...` — PASS. |
| Adapter package | `go test ./internal/adapter/...` — PASS. |
| Formatter check | `gofmt -l .` — no output. |
| Diff whitespace | `git diff --check` — no output. |
| Full tests | `go test ./...` — PASS. |
| Vet | `go vet ./...` — PASS. |
| Build | `go build ./...` — PASS. |
| Runtime harness | N/A — adapter library-only slice; CLI fixture audit is deferred until task 4.3 wiring. |
| Rollback boundary | Remove `internal/adapter/opencode/` and this Work Unit 3 Slice 1 section; no target configuration was read or modified. |

### Deviations from Design

- Full task 3.1 is intentionally not marked complete. This slice establishes the OpenCode adapter boundary and fail-closed parser/resolver behavior, but does not yet declare a support-matrix-backed OpenCode version or complete evidence-backed conformance fixtures.
- Source ordering for precedence is modeled from the `[]source.SelectedSource` order supplied to the adapter; broader discovery/merge ordering remains part of the remaining task 3.1/3.3 work.

### Authored-Line Accounting

- Adapter code and tests before progress update: 265 authored new lines (`internal/adapter/opencode/opencode.go` 173, `internal/adapter/opencode/opencode_test.go` 92).
- Progress evidence adds this section only; task checkboxes are unchanged.
- Exact candidate accounting after progress update: **333 authored changed lines** (333 additions, 0 deletions), including untracked adapter files and OpenSpec progress.

### Remaining

Exact unchecked implementation tasks remain:

- [ ] 3.1 RED then implement `internal/adapter/opencode/` plan, parser, resolver, runtime limits, and evidence-backed fixtures.
- [ ] 3.2 RED then implement `internal/adapter/claudecode/` with the same gates; preserve target semantics, not a shared policy engine.
- [ ] 3.3 Test allow/deny/ask/default/conditional/unresolved outcomes, precedence traces, ambiguity, and static-vs-runtime limitations.
- [ ] 4.1 RED then implement `internal/compare/` dimension-aware equivalent/broader/narrower/target-only/unsupported/ambiguous/lossy/not-comparable findings.
- [ ] 4.2 Implement `internal/redact/` before `internal/render/`; add versioned JSON, stable ordering/IDs, human blockers-first output, and redaction goldens.
- [ ] 4.3 Wire `internal/app/` and `cmd/auditor` requests, explicit sources/discovery, streams, help, and exits; test read-only, no-write, byte-stable audits.
- [ ] 4.4 Document alpha limits, modeled-not-enforced semantics, matrix/evidence process, verification commands, and exclusions.
- [ ] 5.1 Run `gofmt -l .`, `go vet ./...`, `go test ./...`, and `go build ./cmd/auditor`; inspect all goldens generated through the approved `-update` path, then rerun without `-update`.
- [ ] 5.2 Run fuzz/property tests for depth, files, bytes, rules, references, diagnostics, and cancellation; verify unchanged content/metadata and no process/network execution.

### Structured Status Consumed

- `schema`: `gentle-ai.sdd-status/v1`; `artifactStore`: `openspec` with session hybrid persistence.
- `nextRecommended`: `apply`; `dependencies.apply`: `ready`; `taskProgress`: 6/15 complete.
- `actionContext.mode`: repo-local; workspace root and allowed edit root: `/home/jkelevra/Code/Proyect_OpenSource`.
- Remediation required: false. No action-context warning; all edited paths are under the allowed root.
