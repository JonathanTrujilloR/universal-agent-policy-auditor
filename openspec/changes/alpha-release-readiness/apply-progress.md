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
Chain context remains five WU3 slices: A1 JSON/envelope, A2 metadata integrity, A3 pinned evidence/fixture metadata with the production matrix still unsupported, B adapter conformance that activates only the proven row, and C app integration; Issue #29 remains open until WU3-C.

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

## Work Unit 3-A2: Support metadata integrity

Status: complete for WU3-A2 only; boundary is implementation PR 8 of 17, intermediate `Refs #29`. Issue #29 remains open until WU3-C; no stage, commit, push, PR, release, or review-lifecycle action performed.

### Structured status and workload
- Parent-selected change: `alpha-release-readiness`; artifact store: `openspec`; mode: repo-local; allowed edits limited to `support/matrix.go`, `support/matrix_test.go`, this tasks file, and this progress file.
- Workload guard resolved as `stacked-to-main`, WU3-A2 slice only, hard maximum 400 authored changed lines, no `size:exception`.

### TDD Cycle Evidence
| Phase | Evidence |
|---|---|
| SAFETY | `go test ./support -count=1` passed before production edits. |
| RED | Focused A2 metadata tests first failed on missing strict fixture/evidence metadata fields. |
| GREEN | Added tagged metadata records and fail-closed validation; focused A2 tests passed. |
| TRIANGULATE | Added table cases for scalar whitespace, repo/tag/commit/SHA/path/claim/source/reference/triple failures; `go test ./support -count=1` passed. |
| REFACTOR | Consolidated validation helpers, documented lexical fs.FS fixture confinement, and reran focused package tests. |

### Acceptance, deferrals, and rollback
- `support.Load` now validates explicit fixture, evidence, authority, and source records before registry maps are built.
- Authority repository validation is offline and limited to canonical GitHub HTTPS repository-root URLs; tags and commits are lexical only.
- Fixture paths are trimmed non-dot `fs.ValidPath` names read from the supplied `fs.FS` and SHA-256 checked; this is lexical fs-relative validation, not OS or symlink sandboxing.
- Source paths are trimmed non-dot unique `fs.ValidPath` upstream names; they are not read locally and remain pinned source metadata for A3.
- Checked registries are unchanged and empty; `matrix.json`, `fixtures.json`, and `evidence.json` keep empty arrays, so support remains zero.
- Remaining WU3 slice lines: `- [ ] WU3-A3: Pin selected exact OpenCode version evidence and fixture metadata while the production matrix remains unsupported. <!-- sdd-owner: implementation -->`; `- [ ] WU3-B: Implement adapter conformance and activate only the proven production matrix row without changing app support claims. <!-- sdd-owner: implementation -->`; `- [ ] WU3-C: Integrate WU3 support behavior into the app; Issue #29 remains open until WU3-C. <!-- sdd-owner: implementation -->`; broad WU3 rows remain unchanged/unchecked.
- RED command: `go test ./support -run 'TestLoadAcceptsStrictGenericMetadata|TestLoadRejectsMetadataIntegrityFailures' -count=1` failed before production correction; focused A2 tests passed after implementation.
- Final verification: `go test ./support -count=25`, `go test -race ./support -count=1`, `git diff --check`, `gofmt -l support/matrix.go support/matrix_test.go`, `go vet ./...`, `go test ./... -count=1`, and `go build ./...` passed.
- Final diff: 364 insertions and 34 deletions, 398 authored changed lines across the four allowed paths.
- Rollback: revert `support/matrix.go`, `support/matrix_test.go`, WU3-A2 task checkbox/current-slice text, and this progress block.

## Work Unit 3-A3: Pinned OpenCode evidence and fixture metadata
Status: complete for WU3-A3 only; boundary is intermediate PR 9 of 17, `Refs #29`, with no stage, commit, push, PR, release, review, tag, or publication action.
Structured status consumed: parent-selected `alpha-release-readiness`, `openspec`, repo-local; allowed edits were limited to WU3-A3 support registries/fixture/docs/tests and OpenSpec task/progress artifacts.
Workload guard: explicit stacked-to-main delivery path, hard <=400 authored changed lines, no `size:exception`; final authored count: 322 additions+deletions.
TDD Cycle Evidence:
| Phase | Evidence |
|---|---|
| SAFETY | `go test ./support -count=1` passed before edits. |
| RED | Added pinned registry/manifest/fixture tests first; focused commands failed on empty checked-in registries and the absent fixture, including the fixture data-drift sentinel. |
| GREEN | Added only `support/testdata/opencode-1.18.27-permission-legacy-scalar.json`, one fixture record, one evidence record, and evidence docs; focused pinned tests passed. |
| TRIANGULATE | Wrong-pinned authority/source equality, fixture digest/data drift, zero matrix entries, unsupported `Validate`, and CLI exit 2 boundaries are covered by focused tests. |
| REFACTOR | `gofmt -w support/matrix_test.go` then focused support tests passed; no production code changed. |
Changed files: `support/fixtures.json`, `support/evidence.json`, `support/evidence.md`, `support/testdata/opencode-1.18.27-permission-legacy-scalar.json`, `support/matrix_test.go`, this tasks file, and this progress file.
Fixture SHA-256: `22ab8de00a73350aedcb72b62db5c962c910f15e12fbef80d844e725593bacf9`; fixture has only scalar `read:allow`, `edit:deny`, and `bash:ask` under singular `permission`.
Authority: `anomalyco/opencode` tag `v1.18.27`, direct tag commit `4b7e19e315cca414121ba1d61523fef74bb3ae8b`, release URL available; source hashes are the five exact checked values, including corrected V2 hash `229f2da7d245bf86deffed59b10cddf2dd04ffa05d10ef97fc5df289d0e6fb98`.
Claims: `root-permission-legacy-info`, `legacy-actions-scalar-resource-map`, `scalar-migrates-wildcard-action-effect`, `v2-ordered-last-wildcard-fallback-ask`, and `runtime-effective-policy-out-of-scope`.
Zero-support statement: `support/matrix.json` and `support/matrix.go` are byte-unchanged, matrix entries remain empty, `support.Load` validates registries, `Validate` stays `unsupported-version`, and the OpenCode CLI request exits 2.
Verification passed: `go test ./support -run 'TestCheckedInOpenCodePinnedRegistryEvidenceAndFixture' -count=1`; `go test ./support -run 'TestCheckedInOpenCodeRegistryRejectsFixtureDataDrift|TestLoadConsumesConfiguredPinnedOpenCodeRegistryWithoutSupportRows|TestCheckedInOpenCodeAuditCLIStillExitsUnsupported' -count=1`.
Verification passed: `go test ./support -count=25`; `go test -race ./support -count=1`; `git diff --check`; `test -z "$(gofmt -l support/matrix_test.go)"`; `go vet ./...`; `go test ./... -count=1`; `go build ./...`.
Completed task checkbox update: WU3-A3 is checked in `tasks.md`; WU3-B, WU3-C, and broad WU3 rows remain unchecked.
Remaining exact unchecked rows: `- [ ] WU3-B: Implement adapter conformance and activate only the proven production matrix row without changing app support claims. <!-- sdd-owner: implementation -->`; `- [ ] WU3-C: Integrate WU3 support behavior into the app; Issue #29 remains open until WU3-C. <!-- sdd-owner: implementation -->`; broad WU3 rows remain deferred.
Rollback boundary: revert the WU3-A3 fixture, fixture/evidence registries, evidence docs, pinned tests, the WU3-A3 checkbox, and this progress block only.
Deviations: none from the assigned slice; runtime effective policy, agents/modes, saved approvals, multiple sources, V2 arrays, resource maps, enforcement, security, and compliance remain excluded.
## Work Unit 3-B1: Exact OpenCode parser/conformance cleanup
Status: complete for intermediate PR 10 of 18, `Refs #29`; consumed alpha worktree status override, allowed edits only, final authored count 400 lines, no stage/commit/push/PR/review.
### TDD Cycle Evidence
| Phase | Evidence |
|---|---|
| SAFETY/RED | Baseline `go test ./internal/adapter/opencode -count=1` passed; focused RED failed on exact-version bypass, source identity/path/bounds, wildcard scalar fixture, invented object fixture, strict JSON shapes, and requested-capability cleanup. |
| GREEN/TRIANGULATE/REFACTOR | Implemented exact 1.18.27 gate before `support.Validate`, one-source bound, token duplicate/trailing/key scan, exact scalar `read/edit/bash`, wildcard permissions, unresolved deduped requested capabilities, and removed invented matcher/condition/precedence/runtime helpers; focused/x25/race adapter tests passed. |
| Verification/Rollback/Remaining | Passed support checked Load/Validate/CLI-exit2 focused test, `go test ./support -count=1`, `go test ./internal/app ./cmd/auditor -count=1`, `git diff --check`, `gofmt -l`, `go vet ./...`, `go test ./... -count=1`, `go build ./...`; matrix empty, B2/C deferred; rollback adapter files plus B1 OpenSpec edits. |

## Work Unit 3-B2: Production OpenCode support row activation

Status: implementation complete for intermediate PR 11 of 18, `Refs #29`; lifecycle and delivery remain parent-owned.

### TDD evidence

| Phase | Evidence |
|---|---|
| SAFETY / RED | Baseline support and adapter tests passed; focused B2 tests then failed while the checked matrix had zero rows. |
| GREEN | Added exactly one checked `opencode` / `1.18.27` / `permission` row linked to the pinned A3 fixture and evidence; focused tests passed. |
| TRIANGULATE | Covered the exact tuple, IDs, four scoped flags, nearby version, Claude, unknown construct, false flags, missing references, production adapter conformance, invented object rejection, and unresolved requested capabilities. |
| REFACTOR | Adapter tests load checked production metadata; exact CLI sentinel remains `stdout="unsupported_or_incomplete\n"`, empty stderr, exit 2. |

Focused support command passed: `go test ./support -run 'TestLoadActivatesOnlyCheckedInOpenCodeScalarPermissionRow|TestCheckedInOpenCodeMatrixRejectsMissingReferences|TestValidateRejectsFalseSemanticFlagsOnReferencedRows|TestLoadRejectsMetadataIntegrityFailures|TestCheckedInOpenCodeAuditCLIStillExitsUnsupported' -count=1`.
Focused adapter command passed: `go test ./internal/adapter/opencode -run 'TestResolveConformsExactOpenCode11827ScalarFixture|TestResolveRejectsOldInventedObjectMatcherConditionFixture|TestResolveDeduplicatesRequestedCapabilitiesAndRejectsEmptyRequestedName|TestResolveRequiresExactVersionBeforeSupportValidation|TestResolveStrictlyRejectsUnsupportedJSONShapesAtomically' -count=1`.
Verification also passed: package tests at count 25, race tests, app/CLI tests, `git diff --check`, gofmt, vet, full suite, and build.

Boundary: the singleton row proves only the static legacy scalar subset. Explicit `read`, `edit`, and `bash` mean no defaults are inferred; `*` is limited to scalar migration; admitted rules have no conditions. Resource maps, V2 authoring, agents/modes, multiple sources, runtime behavior, enforcement, security, and compliance remain excluded.

Changed files: `support/matrix.json`, `support/matrix_test.go`, `internal/adapter/opencode/opencode_test.go`, `support/evidence.md`, and the two OpenSpec progress artifacts. WU3-B2 alone is checked; WU3-C and broad WU3 acceptance remain pending.

Rollback: remove the singleton row, B2 production-binding tests and evidence wording, the B2 task update, and this block; preserve A3 evidence and B1 parser behavior.

## Work Unit 3-C: App support wiring
Status: implementation complete for final WU3 PR 12 of 18; parent delivery owns `Closes #29`; no stage, commit, push, PR, review, tag, or publication action performed.
Structured status: alpha worktree authority, `openspec`, repo-local, allowed WU3-C files only; ambient original-worktree ambiguity ignored per parent; workload resolved as stacked-to-main, final slice, 388 authored lines including new file, no compression/exception.
TDD evidence: SAFETY `go test ./support ./internal/app ./cmd/auditor -count=1` passed; RED focused support/app tests failed on missing `LoadCheckedIn`, app completion fields, and app exact support path; GREEN added embedded support FS and app orchestration; TRIANGULATE covered unsupported/missing version, Claude, malformed/duplicate/extra/missing/resource-object config, nonexistent root/config, directory, escaping symlink, oversize/read errors, explicit config over env, metadata loader error, and no read before unsupported admission; REFACTOR kept CLI grammar and category-only output.
Changed files: `support/embedded.go`, `support/matrix.json` gate text, `support/matrix_test.go`, `internal/app/app.go`, `internal/app/app_test.go`, `README.md`, `support/evidence.md`, tasks/progress artifacts.
Focused verification passed: `go test ./support ./internal/app ./cmd/auditor -count=1`; `go test ./support -run 'TestLoadCheckedInMatchesRepositoryMetadata|TestCheckedInOpenCodeAuditCLIExitsCompleteForExactFixture' -count=1`; `go test ./internal/app -count=25`; `go test -race ./internal/app ./support -count=1`; `go test ./internal/app ./cmd/auditor ./support ./internal/adapter/opencode -count=1`.
Final verification passed: `git diff --check`; `gofmt -l .`; `go vet ./...`; `go test ./... -count=1`; `go build ./...`.
Acceptance: source-built CLI exact fixture/version from unrelated cwd exits 0 with stdout `complete_no_findings\n` and empty stderr; wrong/unsupported and incomplete inputs fail closed; category-only CLI emits no raw config/path/error details.
WU3 aggregate rows are now checked from A1/A2/A3/B1/B2/C evidence; remaining unchecked implementation rows are WU4-WU9 renderer/comparison/release/pilot slices, and parent lifecycle gates remain deferred.
Rollback: revert WU3-C app/support embedding/docs/tests and this OpenSpec update; preserve merged WU3-A/B evidence unless reverting all WU3 broad acceptance rows together.

## Work Unit 4-A1: App metadata for later redaction
Status: complete for intermediate PR 13 of 21, `Refs #36`; no stage/commit/push/PR/review/tag/release action.
Scope: app metadata and adjacent tests plus WU4-A1 OpenSpec tracking only; redaction, renderers, CLI format changes, and later slices remain deferred.
Workload: issue #36 four-slice chain; WU4-A1 only, A2 redaction core/A3 privacy/B JSON pending; CLI JSON activation remains WU5.
TDD evidence: SAFETY `go test ./internal/app ./cmd/auditor ./support -count=1` passed; RED focused app metadata tests failed to compile on missing `SupportStatus`/source digest metadata; GREEN focused app metadata tests passed.
TRIANGULATE: covered exact success source digest, malformed admitted input, source error, missing/nearby version, Claude, unknown target withholding, invalid path/request, metadata load error, valid deterministic/deduped/nonaliasing digests, and unchanged target files.
REFACTOR: added `auditResult`, audit limitations, and `sourceDigests` helper so every audit return carries closed metadata without changing CLI output.
Verification passed: `go test ./internal/app -count=1`; `go test ./internal/app ./cmd/auditor ./support -count=1`; `go test ./internal/app -count=25`; `go test -race ./internal/app -count=1`.
Verification passed: `git diff --check`; `test -z "$(gofmt -l .)"`; `go vet ./...`; `go test ./... -count=1`; `go build ./...`.
Changed files: `internal/app/app.go`, `internal/app/app_test.go`, `openspec/changes/alpha-release-readiness/tasks.md`, and this progress file.
Final authored churn: 230 additions + 31 deletions = 261 authored lines; hard cap 400.
Rollback: revert WU4-A1 app metadata/tests and this OpenSpec update only; preserve WU3 app support behavior and leave A2/A3/B pending.
Remaining unchecked implementation rows include WU4-A2, WU4-A3, WU4-B, WU5-WU9; parent lifecycle rows remain deferred.
## Work Unit 4-A2: Redaction core
PR 14/21 (`Refs #36`), A2 only; A3/B deferred. Prior recorded RED: duplicate findings/state acceptance failures; remediation rejects unrelated trace evidence and accepts the reachable OpenCode unsupported-version result with fixed finding projection and exact sole trace/provenance evidence equality; GREEN: `go test ./internal/redact -count=25`, `go test -race ./internal/redact -count=1`, `git diff --check`, `gofmt -w internal/redact/report.go internal/redact/report_test.go`, `go vet ./...`, `go test ./... -count=1`, `go build ./...` passed. Rollback: report.go/report_test.go and A2 task/progress edits. Final authored count: 400 additions+deletions (cap 400); no delivery actions.

## Work Unit 4-A3: Privacy hardening, fuzzing and documentation
Status: implementation complete for intermediate PR 15/21, `Refs #36`; settlement/delivery remain parent-owned. A1 PR #37 and A2 PR #38 are merged; B remains pending.
Safety: `go test ./... -count=1` passed before edits.
RED: `go test ./internal/redact -run 'TestPrivacy' -count=1` failed on the permissive tool-version policy.
GREEN: the same focused command passed with exact public version allowlists; empty requested version preserves A2 missing/unresolved semantics.
TRIANGULATE: discarded metadata, hostile paths/tokens/control/invalid UTF-8/confusables, unknown codes, generic errors and repeated nested-copy mutation are covered.
REFACTOR: removed syntax-based version admission; retained existing deep-copy implementation and protected it with accessor projection tests.
Verification: `go test ./internal/redact -count=25`; `go test -race ./internal/redact -count=1` passed.
Fuzz: `go test ./internal/redact -run '^$' -fuzz '^FuzzReportPrivacy$' -fuzztime=3s -parallel=1` passed with eight seeds.
Full checks: `git diff --check`; `gofmt -l .`; `go vet ./...`; `go test ./... -count=1`; `go build ./...` passed.
Scope: safe report/version policy, privacy tests, report-safety docs and these tracking edits only; no renderer or CLI changes.
Workload: 217 authored additions+deletions, below 400; no compression or delivery actions.
Rollback: revert A3 report policy, privacy test, report-safety document and A3 task/progress edits only; preserve A1/A2.

## Work Unit 4-B: Versioned JSON
Status: implementation complete for final PR 16/21; A1/#37, A2/#38, A3/#39 merged; issue #36 stays open until parent delivery.
Safety: `go test ./... -count=1` passed before edits.
RED: `go test ./internal/render/json -count=1` failed because the renderer package had no production Go files.
GREEN: the same focused command passed after adding valid-report-only `Marshal` and a fixed invalid-report error.
TRIANGULATE: exact success/unsupported/incomplete goldens, exits 0/2/3/4, reserved category rejection, nil arrays, canary absence and 100 repeats passed.
REFACTOR: private ordered structs retain standard JSON escaping; returned bytes are independent; CLI activation stays WU5.
Focused checks passed: `go test ./internal/render/json -count=25`; `go test -race ./internal/render/json -count=1`.
Formatting passed: `gofmt -w internal/render/json/report.go internal/render/json/report_test.go`; `gofmt -l .`.
Full checks passed: `go vet ./...`; `go test ./... -count=1`; `go build ./...`.
Acceptance: WU4 implementation and aggregate evidence complete; WU5+ and parent delivery gates remain pending.
Rollback: revert B renderer/tests, report-safety JSON wording and B task/progress edits only; preserve A1/A2/A3 boundaries above.
Workload: `git diff --check` passed; final numstat totals 242 additions + 7 deletions = 249; cap 400, no compression or delivery/review actions.
