# Apply Progress: Alpha Release Readiness
## Work Unit 1: Public Project Identity, Legal, and Docs Baseline

Status: complete for assigned Work Unit 1 only. Delivery boundary: stacked-to-main PR 5 of 13, approved issue #20 (`status:approved`, `type:docs`), no commit/push/open/merge performed.

### Structured status consumed

- changeName: `alpha-release-readiness`
- artifactStore: `openspec`
- applyState: ready (parent-provided for this workspace)
- actionContext: repo-local, workspaceRoot repository root, allowed edits limited to `LICENSE`, `README.md`, this tasks file, and this apply-progress file; warnings: none.
- Review workload guard resolved by parent: `stacked-to-main`, assigned slice only, hard maximum 400 authored changed lines.

### TDD Cycle Evidence

| Phase | Evidence |
|---|---|
| RED | `test -f LICENSE && grep -Fq 'Apache License' LICENSE && test -f README.md && grep -Fq 'Apache-2.0' README.md` exited 1 because `LICENSE` and `README.md` were absent. |
| GREEN | Added canonical Apache-2.0 `LICENSE` and concise README; structural check confirmed Apache identity, zero supported targets, OpenCode evidence gate, and Claude unsupported wording. |
| TRIANGULATE | Claim-boundary checks verified required negative/support text and absence of positive forbidden claims such as Claude support, OpenCode support, package-manager install commands, production readiness, and alpha binary download claims. A broader first grep false-positive matched intentional negative wording and was replaced with explicit positive-claim checks. |
| REFACTOR | README readback confirmed scannable status table, current verification commands, architecture, roadmap, contribution path, license link, and maintainer-controlled metadata follow-up only. Full Go checks passed. |

### Completed task checkbox updates
- WU1 RED, GREEN, TRIANGULATE, REFACTOR, acceptance evidence, and rollback boundary are checked; the already-delivered Planning PR D prerequisite is synchronized in `tasks.md`.

### Files changed

- `LICENSE` — canonical Apache License 2.0 text.
- `README.md` — pre-alpha public identity, support/status table, limitations, architecture, verification, roadmap, contribution path, metadata follow-up, Apache-2.0 statement.
- `openspec/changes/alpha-release-readiness/tasks.md` — WU1 checkbox updates plus synchronization of the delivered Planning PR D prerequisite.
- `openspec/changes/alpha-release-readiness/apply-progress.md` — this evidence record.

### Verification commands

- RED structural precondition: expected failure, exit 1.
- GREEN structural readback: pass.
- TRIANGULATE claim-boundary checks: pass after replacing an over-broad false-positive grep.
- `git diff --check`: pass.
- `gofmt -l .`: pass, no output.
- `go vet ./...`: pass, no output.
- `go test ./... -count=1`: pass.
- `go build ./...`: pass, no output.

### Acceptance evidence and rollback

- Apache-2.0 appears in `LICENSE` and `README.md`; no conflicting license claim was found.
- README states pre-alpha, not yet installable/supported, zero supported target versions, OpenCode evidence-gated, Claude unsupported, static modeled semantics only, no runtime enforcement/compliance/security guarantee, and no adoption claim.
- No target config, source, runtime, GitHub metadata, release, tag, or publication surface changed.
- Rollback boundary: remove/revert `LICENSE`, `README.md`, WU1 task checkbox updates, and this progress note only.

### Remaining scope and deviations

- Assigned Work Unit 1: no unchecked implementation rows remain.
- Non-WU1 implementation rows and parent-owned lifecycle gates remain unchecked in `tasks.md`; they are intentionally deferred.
- Gatekeeper correction: restored the official Apache-2.0 header bytes, removed a private absolute workspace path, and synchronized the already-delivered Planning PR D state; repository metadata remains maintainer-controlled and unmodified.

## Work Unit 2: CLI/App Shell

Status: complete for assigned WU2 only. Boundary: stacked-to-main implementation PR 6 of 13, approved issue #27 (`status:approved`, `type:feature`); no stage/commit/push/GitHub metadata/tag/release action performed.

### Structured status and TDD evidence
- Status consumed: `alpha-release-readiness`, `openspec`, apply ready, repo-local actionContext, allowed edits limited to `cmd/auditor/**`, `internal/app/**`, and WU2 OpenSpec artifacts. Workload: WU2 only, stacked-to-main, final 396 authored additions+deletions, hard max 400, no `size:exception`.
- RED: `go test ./cmd/auditor ./internal/app` first failed on missing app categories/request/exit mapping and CLI `run`; remediation RED failed on relative path, malformed help, Claude shape, and writer-error regressions.
- GREEN/TRIANGULATE: added typed app shell and transport-only CLI, then required absolute root/config before lexical containment, validated paths before Claude unsupported, rejected help trailing args, and returned operational failure on write errors.
- REFACTOR: removed unnecessary sleep/time test code; kept target semantics and target I/O out of `cmd/auditor`; app exposes no writer/process/network/Git/GitHub/config-update port and does not read target bytes.

### Completed tasks, files, verification
- WU2 RED, GREEN, TRIANGULATE, REFACTOR, acceptance evidence, and rollback boundary are checked in `tasks.md`.
- Files: `cmd/auditor/main.go`, `cmd/auditor/main_test.go`, `internal/app/app.go`, `internal/app/app_test.go`, WU2 task/progress artifacts.
- Commands: focused RED/GREEN/TRIANGULATE/REFACTOR, repeated focused tests, race, `git diff --check`, `gofmt -l .`, `go vet ./...`, `go test ./... -count=1`, and `go build ./...` passed.
- Acceptance: deterministic dev version/help; exact audit grammar; missing version/Claude well-formed requests return 2; malformed/relative/outside-root/mutation-like/help-trailing requests return 3; unknown app category and write errors map to 4; `t.TempDir()` config bytes/mode/modtime remain unchanged.
- Remaining unchecked tasks are WU3-WU9 and parent-owned lifecycle gates. Rollback: remove `cmd/auditor/**`, `internal/app/**`, and WU2 task/progress edits only.


## Work Unit 3-A1: Strict JSON and envelope integrity

Status: complete for the reduced WU3-A1 slice only; boundary is implementation PR 7 of 17, intermediate `Refs #29`, with no stage, commit, push, PR, release, or review-lifecycle action.
Chain context remains five WU3 slices: A1 JSON/envelope, A2 metadata integrity, A3 pinned evidence/fixture metadata with the production matrix still unsupported, B adapter conformance that activates only the proven row, and C app integration that closes `#29`.

TDD evidence:
- Safety net: `go test ./support` passed before edits.
- RED: `go test ./support -run 'TestLoadRejectsStrictEnvelopeViolations|TestStrictJSONScannerBoundsDepthDirectly|TestLoadAcceptsStrictGenericEnvelopeWithIDOnlyRegistries' -count=1` failed before production edits on missing strict scanner plumbing.
- GREEN: same focused strict JSON/envelope command passed after implementing scanner, envelopes, unknown-field checks, and `Entry` JSON tags.
- TRIANGULATE: focused depth-limit, Unicode-value, generic-envelope, and zero-support cases passed.
- REFACTOR: `gofmt -w support/matrix.go support/matrix_test.go && go test ./support -count=1` passed.

Changed repo-relative paths:
- `support/matrix.go`
- `support/matrix_test.go`
- `openspec/changes/alpha-release-readiness/tasks.md`
- `openspec/changes/alpha-release-readiness/apply-progress.md`

Acceptance: `support.Load` now requires exact schemas `support-matrix/v1`, `conformance-fixtures/v1`, and `support-evidence/v1`; requires non-null arrays `entries`, `fixtures`, and `evidence`; allows empty arrays; allows matrix `gate`; rejects malformed JSON, trailing documents, unknown modeled fields, recursive duplicate keys, noncanonical object keys, Unicode-confusable keys, uppercase/case variants, and nesting deeper than 8.
Checked-in `matrix.json`, `fixtures.json`, and `evidence.json` remain unchanged with empty arrays, so support remains zero and existing unsupported CLI behavior is unchanged.
Deferred: WU3-A2 owns metadata integrity such as empty/duplicate IDs, URLs, tags, source paths, hashes, digests, evidence links, and row tuples; WU3-A3 owns pinned evidence/fixture metadata while leaving the production matrix unsupported until WU3-B proves and activates one row.
Rollback: revert `support/matrix.go`, `support/matrix_test.go`, the WU3-A1 task edits, and this progress block only.
Final verification: `go test ./support -count=25`, `go test -race ./support -count=1`, `git diff --check`, `gofmt -l support/matrix.go support/matrix_test.go`, `go vet ./...`, `go test ./... -count=1`, and `go build ./...` passed; final diff was 307 insertions and 15 deletions across the four repo-relative paths.
