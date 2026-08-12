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
