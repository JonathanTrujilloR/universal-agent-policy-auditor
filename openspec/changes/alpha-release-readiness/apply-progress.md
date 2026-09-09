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
