# Tasks: Alpha Release Readiness

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 2,340-3,340+ authored lines total: 940 current planning lines plus 1,400-2,400 implementation lines and this correction |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | Planning PRs A-D → PR 5 public identity → PR 6 CLI/app shell → PRs 7-12 WU3 → PRs 13-16 WU4 → PRs 17-18 WU5 → PR 19 #19 comparison core → PRs 20-25 #45 comparison app/redaction DTO/privacy/JSON/text/CLI → PR 26 CI/release build → PR 27 release docs/prerelease → PR 28 post-release pilot pack |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

Current slice: implementation PR 22 of 28, #45 comparison integration B2 (privacy/fuzz/docs evidence). Planning PRs A-D, WU1-WU6, and #45-A/B1 are delivered; issues `#19`, `#36`, and `#41` are closed. Approved issue `#45` governs integration slices A, B1, B2, C, D, and E; only final E delivery closes it. Chain: `stacked-to-main`, each slice targets `main`, stays within 400 authored changed lines, and records rollback/evidence independently. No `size:exception` is authorized.

## Task Ordering and Completion Rules

- Do not start implementation apply for any Work Unit until the planning-artifact PR path is approved and delivered, and exactly one coherent Work Unit issue is approved with `status:approved` and exactly one `type:*` label. Issue `#19` governs comparison core, issue `#20` governs public identity, issue `#21` governs planning only, issue `#29` governed WU3, and issue `#36` governed WU4 output/redaction/JSON slices, and approved issue `#41` governs WU5 human renderer/CLI slices. <!-- sdd-owner: implementation -->
- Keep every Work Unit independently reviewable, rollbackable, and below the 400 authored changed-line budget; split before exceeding the budget instead of using an oversized PR unless a maintainer explicitly grants `size:exception`. <!-- sdd-owner: implementation -->
- For runtime/code Work Units, follow strict TDD in this order: RED failing focused test, GREEN minimal implementation, TRIANGULATE additional behavior/edge case, REFACTOR with focused and full verification. <!-- sdd-owner: implementation -->
- Record exact focused commands and full commands for each Work Unit; default full verification is `gofmt -l .`, `go vet ./...`, `go test ./...`, and `go build ./...` when relevant. <!-- sdd-owner: implementation -->
- Do not depend on blocked issues `#11` or `#13`; Claude Code remains visibly unsupported throughout the alpha. <!-- sdd-owner: implementation -->
- Do not create tags, GitHub releases, repository metadata changes, issue approvals, or publication events during apply; those are explicit human-controlled delivery gates. <!-- sdd-owner: implementation -->
- Do not mark a task complete until its acceptance evidence, rollback boundary, issue gate, and changed-line forecast/actual count are recorded. <!-- sdd-owner: implementation -->

## Planning-Artifact Delivery Before Implementation Apply

Planning artifacts already account for 940 authored changed lines before this correction: `openspec/config.yaml` 1 addition + 1 deletion, exploration 159 lines, preproposal 33 lines, proposal 167 lines, spec 216 lines, design 225 lines, and prior tasks 138 lines. These lines are part of review workload and MUST NOT be bundled into implementation Work Unit 1.

Planning issue/reference policy: use one coherent approved planning issue only if it truthfully covers the full alpha-release planning scope. Do not reuse issue `#20` for planning artifacts if `#20` is limited to Work Unit 1 public identity/legal/docs baseline. Intermediate planning PRs use `Refs #N`; only the final planning PR may close the planning issue. Each planning PR requires `status:approved`, exactly one `type:*` label, chain context, rollback boundary, and structural readback evidence.

- [x] Planning PR A delivered config, exploration, preproposal, and proposal in PR `#22` (361 authored changed lines), merged as `30d5732817300fdb2ee553e3618f957eb6805756` after independent structural/test verification. <!-- sdd-owner: implementation -->
- [x] Planning PR B delivered the corrected alpha-release spec in PR `#23` (281 authored changed lines), merged as `ce546dba0aa9eda28f1b0b3e8774bdd37fa31fbd` after independent traceability/test verification. <!-- sdd-owner: implementation -->
- [x] Planning PR C delivered the conservative architecture in PR `#24` (225 authored changed lines), merged as `2eccd0d59ef8e5dbd937dbba828061d306328e6f` after independent architecture/test verification. <!-- sdd-owner: implementation -->
- [x] Planning PR D delivered this corrected task plan in PR `#25` (152 authored changed lines), merged as `cd4bd0904898530645b9cd3a6edab25c1b6e6ef1`, and closed planning issue `#21`. <!-- sdd-owner: implementation -->
- [x] Planning delivery evidence records issue `#21`, intermediate `Refs #21`, the final closing reference, stacked-to-main order, per-PR counts, structural readback, and rollback boundaries across PRs `#22`-`#25`. <!-- sdd-owner: implementation -->
- [x] Implementation Work Units remained pending until all planning PRs were delivered and the maintainer approved the revised chain, including six WU3 slices: A1 JSON/envelope, A2 metadata integrity, A3 pinned evidence, B1 parser, B2 row activation, and C app. <!-- sdd-owner: implementation -->

## Work Unit 1: Public Project Identity, Legal, and Docs Baseline

Issue gate: approved issue `#20` (`status:approved`, `type:docs`). Forecast: 120-250 authored lines. File surfaces: `LICENSE`, `README.md`, release-facing docs under `docs/` or `.github/` discovered during apply, repository metadata checklist documentation only.

- [x] RED: Add/adjust documentation checks or review fixtures that fail when Apache-2.0 identity is absent from `LICENSE`, `README.md`, and release-facing documentation; focused command: `go test ./...` if checks are Go-based, otherwise structural readback of changed Markdown/license files. <!-- sdd-owner: implementation -->
- [x] GREEN: Add Apache-2.0 `LICENSE`, README positioning, alpha limitation language, and metadata checklist text without claiming OpenCode version support, Claude support, runtime enforcement, package-manager install, signing/provenance, or adoption. <!-- sdd-owner: implementation -->
- [x] TRIANGULATE: Add a claim-boundary case covering missing/conflicting license or forbidden release claims; focused command matches the RED check or structural readback. <!-- sdd-owner: implementation -->
- [x] REFACTOR: Tighten docs for scannability and reviewer verification; full evidence: `gofmt -l .`, `go vet ./...`, `go test ./...`, and Markdown/license readback as applicable. <!-- sdd-owner: implementation -->
- [x] Record acceptance evidence that every release-facing surface says Apache-2.0 consistently and that Claude/OpenCode/support limitations are visible. <!-- sdd-owner: implementation -->
- [x] Record rollback boundary: revert `LICENSE`, `README.md`, and identity docs/metadata checklist only, with no source/runtime changes. <!-- sdd-owner: implementation -->

## Work Unit 2: CLI/App Shell

Issue gate: approved issue `#27` (`status:approved`, `type:feature`). Forecast: 250-390 authored lines. File surfaces: `cmd/auditor/**`, `internal/app/**`, narrow tests beside those packages, build metadata file if needed.

- [x] RED: Add failing focused tests for `auditor version`, help/invalid request handling, stable exit categories, explicit `audit opencode --root --config --opencode-version`, and no mutation of `t.TempDir()` target files; focused command: `go test ./cmd/auditor ./internal/app`. <!-- sdd-owner: implementation -->
- [x] GREEN: Implement transport-only `cmd/auditor` and typed `internal/app` request/result shell with build metadata defaults, exit category mapping, explicit input validation, and no writer/process/network/Git/GitHub ports. <!-- sdd-owner: implementation -->
- [x] TRIANGULATE: Add unsupported `claudecode`, unsupported/missing OpenCode version, outside-root config, and mutation-like command cases; focused command: `go test ./cmd/auditor ./internal/app`. <!-- sdd-owner: implementation -->
- [x] REFACTOR: Keep CLI parsing out of semantic packages and app results independent of renderer DTOs; full command: `gofmt -l . && go vet ./... && go test ./... && go build ./...`. <!-- sdd-owner: implementation -->
- [x] Record acceptance evidence for deterministic version/help behavior, numeric/category exits `0-4`, explicit local input, and unchanged target file snapshots. <!-- sdd-owner: implementation -->
- [x] Record rollback boundary: remove `cmd/auditor` and `internal/app` shell without touching existing `internal/source`, `internal/adapter`, `internal/model`, or `support` semantics. <!-- sdd-owner: implementation -->

## Work Unit 3: OpenCode Evidence Gate

Issue gate: approved issue `#29` (`status:approved`, implementation WU3 scope) defines six bounded slices: A1 JSON/envelope, A2 metadata integrity, A3 pinned evidence, B1 parser, B2 row activation, and C app. Forecast is enforced per slice under 400 authored changed lines.

### WU3 reduced slice tracking for issue `#29`

- [x] WU3-A1 RED: Add failing focused `support.Load` and direct scanner tests for malformed/trailing/unknown/schema/missing-array/duplicate-key/case-variant/Unicode-fold/noncanonical-key/depth strict JSON failures. <!-- sdd-owner: implementation -->
- [x] WU3-A1 GREEN: Implement strict JSON document scanning, exact envelope schemas, required non-null arrays, optional matrix `gate`, unknown-field rejection, duplicate-key rejection, ASCII lower-snake object keys, depth limit, and explicit `Entry` JSON tags without adding support rows. <!-- sdd-owner: implementation -->
- [x] WU3-A1 TRIANGULATE: Prove Unicode values remain valid, strict generic envelopes with ID-only fixture/evidence records load, and checked-in zero-support registries remain empty. <!-- sdd-owner: implementation -->
- [x] WU3-A1 REFACTOR: Keep strict scanning local to `support`, preserve `Validate`, and run focused plus full verification for this slice. <!-- sdd-owner: implementation -->
- [x] WU3-A1 acceptance and rollback: record evidence that checked-in registries are empty, CLI support behavior is unchanged, A2/A3 metadata semantics are deferred, and rollback is limited to `support/matrix.go`, `support/matrix_test.go`, and WU3-A1 OpenSpec notes. <!-- sdd-owner: implementation -->
- [x] WU3-A2: Validate metadata integrity such as empty/duplicate IDs, repository URLs, tags, source paths, hashes, digests, row tuples, evidence-to-fixture links, and RED/GREEN/TRIANGULATE/REFACTOR acceptance. <!-- sdd-owner: implementation -->
- [x] WU3-A3: Pin selected exact OpenCode version evidence and fixture metadata while the production matrix remains unsupported. <!-- sdd-owner: implementation -->
- [x] WU3-B1: Clean OpenCode adapter parsing/conformance to exact 1.18.27 scalar `permission` semantics using a test-local support row; production matrix remained empty and app claims unchanged in merged PR `#33` at `c09cf130d13cdf3e15529d6e214a3e401f81ddfe`. <!-- sdd-owner: implementation -->
- [x] WU3-B2: Activate only the proven production matrix row without changing app support claims; merged in PR `#34` as `5026c1b25f59855c2ec33f1793f7c7a147a3653b`. <!-- sdd-owner: implementation -->
- [x] WU3-C: Integrate WU3 support behavior into the app; parent delivery closes issue `#29`. <!-- sdd-owner: implementation -->

The broad WU3 completion rows below are checked from aggregate A1/A2/A3/B1/B2/C evidence; later renderer, release, comparison, and pilot rows remain pending.

- [x] RED: Add failing conformance tests proving unsupported/unresolved status when exact version evidence, registry rows, fixtures, or semantic flags are missing; focused command: `go test ./support ./internal/adapter/opencode ./internal/app`. <!-- sdd-owner: implementation -->
- [x] GREEN: Register only the selected exact OpenCode evidence boundary, fixtures, and support matrix rows needed for the alpha; do not hardcode support in CLI/app. <!-- sdd-owner: implementation -->
- [x] TRIANGULATE: Cover malformed input, unknown permission-bearing constructs, defaults, matcher behavior, precedence/source order, and runtime-dependent behavior as unsupported/incomplete unless evidence proves them. <!-- sdd-owner: implementation -->
- [x] REFACTOR: Keep evidence metadata declarative and adapter behavior fail-closed; full command: `gofmt -l . && go vet ./... && go test ./...`. <!-- sdd-owner: implementation -->
- [x] Record acceptance evidence citing the exact OpenCode version, stable authority, fixture IDs, and passing conformance command output before any support claim is allowed. <!-- sdd-owner: implementation -->
- [x] Record rollback boundary: revert support entries, evidence files, fixtures, and OpenCode conformance tests together. <!-- sdd-owner: implementation -->

## Work Unit 4: Output, Redaction, and Versioned JSON

Issue gate: approved issue `#36` (`status:approved`, output/redaction/JSON scope) defines four bounded slices. Forecast is enforced per slice under 400 authored changed lines. File surfaces are slice-specific: WU4-A1 only `internal/app/**`; WU4-A2/A3 later `internal/redact/**` and privacy tests/docs; WU4-B later `internal/render/json/**`. CLI JSON activation remains WU5 and this Work Unit MUST NOT absorb the human renderer or use `size:exception`.

- [x] WU4-A1 app metadata: Add closed internal audit metadata only: `SupportStatus`, result target/requested-version fields, completeness/support status on every audit branch, stable limitations, and source identity digests from `source.SelectedSource.Identity`; keep default stdout/stderr/exits byte-compatible. <!-- sdd-owner: implementation -->
- [x] WU4-A2 redaction core: Add `internal/redact` safe-report conversion over app results, excluding raw secrets/config values/private paths/raw target data and accepting only closed metadata/provenance digests. <!-- sdd-owner: implementation -->
- [x] WU4-A3 privacy hardening/fuzz/docs: Add focused privacy edge cases, fuzz/property coverage where practical, and docs/checks proving no secret, credential, username, hostname, absolute private path, raw value, token, or unredacted local context is emitted. <!-- sdd-owner: implementation -->
- [x] WU4-B versioned JSON: Add deterministic `auditor-report/v1alpha1` JSON renderer over redacted DTOs with schema version, tool version, target/version/support status, completeness, permissions, findings, provenance digests, exit category, and limitations; no differences or CLI JSON activation here. <!-- sdd-owner: implementation -->
- [x] Record WU4-A2/A3/B acceptance evidence: merged redaction/privacy tests plus B exact sanitized success/unsupported/incomplete goldens, empty arrays, exits, invalid-report rejection, escaping policy and 100 byte-identical repeats. All four WU4 slices are merged, including B PR `#40`; issue `#36` is closed. WU5 is tracked below. <!-- sdd-owner: implementation -->
- [x] Record WU4 rollback boundaries per slice: A1 reverts only app metadata/tests; A2/A3 revert redaction/privacy files; B reverts JSON renderer/schema files without changing adapter/model semantics. <!-- sdd-owner: implementation -->

## Work Unit 5: Human Renderer

Issue gate: approved issue `#41`, split into A renderer and B CLI activation in the 22-PR chain. WU5-A forecast: 300-390 authored lines, hard cap 400. A surfaces: `internal/render/text/report.go`, `internal/render/text/report_test.go`, `docs/report-safety.md`, and these two tracking artifacts. B owns CLI selection/tests including JSON activation; no app/model/redact/JSON/CLI changes in A.

- [x] WU5-A (PR 17/22, `Refs #41`): Deterministic valid-report-only plain text, complete bytes or fixed generic error, exact goldens, exits 0/2/3/4, reserved exit 1 rejection, no color/ambient inputs, no canary, 100 independent repeats, formatting checks and bounded rollback; commands/count in apply-progress. <!-- sdd-owner: implementation -->
- [x] WU5-B (PR 18/22): Wire audit-only CLI text/JSON selection and `--no-color` no-op with strict ordered grammar, legacy default compatibility, single-write transport, fixed failure routing, real CLI tests, docs, and final issue #41 delivery. <!-- sdd-owner: implementation -->

The aggregate WU5 evidence below covers both bounded slices; neither adds comparison semantics.

- [x] RED: Focused renderer tests failed before A; focused explicit-format CLI tests failed with `invalid_request` before B. <!-- sdd-owner: implementation -->
- [x] GREEN: Human rendering consumes only safe reports; audit-only text/JSON selection preserves app semantics and no-format audit/version result bytes; help intentionally advertises new flags. <!-- sdd-owner: implementation -->
- [x] TRIANGULATE: Cover unsupported Claude/OpenCode, missing/unsafe versions, incomplete sources, semantic invalidity, syntax-before-I/O, failure seams, no-color equality, short/error writes, canary absence, real binary output, and nonmutation. <!-- sdd-owner: implementation -->
- [x] REFACTOR: Stable claim-safe output; `gofmt -l .`, `go vet ./...`, `go test ./...`, and `go build ./...` pass. <!-- sdd-owner: implementation -->
- [x] Record acceptance evidence in exact text/JSON outputs, README/report-safety readback, repeated/race tests, and apply-progress. <!-- sdd-owner: implementation -->
- [x] Record rollback boundary: A removes text renderer/docs; B removes CLI format wiring/docs while retaining redaction and JSON. <!-- sdd-owner: implementation -->

## Work Unit 6: #19 Comparison Core

Issue gate: approved issue `#19` governs only a pure canonical permission comparison core. Forecast: 330-400 authored lines. File surfaces: `internal/compare/**` plus this task and apply-progress evidence. App, adapters, support, sources, redaction, renderers, CLI, broader/narrower inference, and Claude support remain excluded.

- [x] RED: Focused tests fail before the comparison API exists. <!-- sdd-owner: implementation -->
- [x] GREEN: Compare complete same-profile, same-coverage canonical facts as `equivalent`, directional `target-only`, `ambiguous`, or `not-comparable`; `OnlyIn` disambiguates reference and target. <!-- sdd-owner: implementation -->
- [x] TRIANGULATE: Cover incompatible/incomplete operands, invalid identities/traces/effects, exact duplicates, conflicting duplicates, semantic mismatches, deterministic ordering, structured keys, provenance independence, and input/result isolation. <!-- sdd-owner: implementation -->
- [x] REFACTOR: Keep the package pure and deterministic; focused repeated/race tests plus formatting, vet, full tests, and build pass. <!-- sdd-owner: implementation -->
- [x] Record acceptance evidence without claiming broader/narrower, runtime enforcement, Claude support, or CLI availability. <!-- sdd-owner: implementation -->
- [x] Record rollback boundary: revert `internal/compare/**` and this tracking only. <!-- sdd-owner: implementation -->

## Work Unit 7: #45 Safe OpenCode Comparison Integration

Issue gate: approved issue `#45` governs six bounded slices and expands the chain to 28 PRs. The first alpha compares only two explicit OpenCode 1.18.27 configs; exit 1 remains reserved, no-format output remains category-only, and JSON uses `auditor-comparison-report/v1alpha1`. Cross-client comparison, Claude support, broader/narrower, lossy inference, and runtime enforcement remain excluded.

- [x] A (PR 20/28): Integrate two independently admitted operands in `internal/app`, delegate exactly once to #19, map only equivalent to success, and preserve audit behavior. <!-- sdd-owner: implementation -->
- [x] B1 (PR 21/28): Add a nominal comparison-mode marker and opaque zero-value-invalid DTO with closed state, source, reason, difference, finding, limitation, version, ordering, bound, and copy validation. <!-- sdd-owner: implementation -->
- [x] B2 (PR 22/28): Add exhaustive privacy canaries, fuzzing, accessor/cardinality matrices, repeated ordering evidence, and comparison digest-risk docs. <!-- sdd-owner: implementation -->
- [ ] C (PR 23/28): Add deterministic atomic JSON for `auditor-comparison-report/v1alpha1`; audit JSON remains unchanged. <!-- sdd-owner: implementation -->
- [ ] D (PR 24/28): Add deterministic atomic blockers-first human comparison text. <!-- sdd-owner: implementation -->
- [ ] E (PR 25/28): Activate strict CLI grammar, category-only default, explicit text/JSON, and single-write transport; final E delivery closes #45. <!-- sdd-owner: implementation -->
- [x] RED A: Focused app tests failed before `ModeCompare` and comparison request/result metadata existed. <!-- sdd-owner: implementation -->
- [x] GREEN A: Exact supported operands are independently selected/resolved; equivalent maps to 0 and every non-equivalent core result maps to 2. <!-- sdd-owner: implementation -->
- [x] TRIANGULATE A: Cover same/different paths, semantic/incomplete/operational failures, admission-before-I/O, exact one comparator call, digests, nonmutation, defensive copies, and reserved exit 1. <!-- sdd-owner: implementation -->
- [x] REFACTOR A: App-only changes pass repeated/race focused tests plus formatting, vet, full tests, and build. <!-- sdd-owner: implementation -->
- [ ] Record aggregate A-E acceptance and rollback evidence; reverting integration must retain the pure #19 core. <!-- sdd-owner: implementation -->

## Work Unit 8: CI, Release Build, Checksums, and Smoke

Issue gate: blocked until a dedicated CI/release-build Work Unit issue is approved. Forecast: 220-380 authored lines. File surfaces: `.github/workflows/**`, `scripts/**` or `tools/**` release helpers, `cmd/auditor` build metadata integration, release evidence docs under `docs/` if needed.

- [ ] RED: Add failing workflow/script tests or dry-run checks for formatting, vetting, tests, Linux amd64 build, version injection, checksum generation, and smoke command capture; focused command: relevant script dry-run plus `go test ./...`. <!-- sdd-owner: implementation -->
- [ ] GREEN: Add CI/release readiness workflow or scripts for `gofmt -l .`, `go vet ./...`, `go test ./...`, `GOOS=linux GOARCH=amd64 go build`, `checksums.txt`, and smoke commands without publishing. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Add smoke coverage for plain `auditor version` and explicit-format `audit opencode` against the checked-in supported fixture; version JSON remains deferred. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Keep release helpers deterministic and free of package-manager, signing, provenance, extra-platform, GitHub publication, or tag creation side effects. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence for formatting, vetting, tests, build, checksum generation, and Linux amd64 smoke from the release candidate commit. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: revert workflow/scripts/evidence docs only; no target configuration state exists. <!-- sdd-owner: implementation -->

## Work Unit 9: Release Docs and Prerelease Preparation

Issue gate: blocked until a dedicated release-docs Work Unit issue is approved and prior evidence Work Units have landed. Forecast: 180-320 authored lines. File surfaces: `README.md`, `docs/**`, `CHANGELOG.md` or release notes file discovered during apply, support matrix documentation.

- [ ] RED: Add documentation claim-boundary checks or structural review checklist that fails when release docs omit Apache-2.0, static-analysis limitations, Linux amd64 evidence, checksums, unsupported Claude, unsupported package managers, or OpenCode evidence references. <!-- sdd-owner: implementation -->
- [ ] GREEN: Write install, usage, support matrix, limitations, release notes draft, checksum instructions, and smoke evidence references grounded only in checked-in facts. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Add negative claim checks for production readiness, security/compliance guarantees, runtime enforcement, package-manager install, unsupported platforms, signing/provenance, and adoption/pilot success. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Make docs reviewer-friendly with quick path, evidence table, limitations, and rollback/publication notes; full verification includes structural readback plus `gofmt -l . && go vet ./... && go test ./...` if code-adjacent references changed. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence that release remains blocked until exact OpenCode evidence and Linux amd64 smoke evidence exist; do not invent a version. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: revert release docs/release notes/support docs only. <!-- sdd-owner: implementation -->

## Work Unit 10: Post-Release Pilot Pack

Issue gate: blocked until a dedicated pilot-pack Work Unit issue is approved; pilot execution starts only after the alpha release exists. Forecast: 80-180 authored lines. File surfaces: `.github/ISSUE_TEMPLATE/**`, `.github/DISCUSSION_TEMPLATE/**` if present, `docs/pilot/**` or equivalent docs discovered during apply.

- [ ] RED: Add structural/readback check that pilot materials position feedback as post-release only and forbid sharing secrets, credentials, private paths, raw sensitive configs, or unredacted local context. <!-- sdd-owner: implementation -->
- [ ] GREEN: Add a feedback template or discussion guide with exact artifact/checksum fields, sanitized command/output guidance, unsupported construct prompts, false-assurance prompts, and privacy warnings. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Add wording checks that pilot feedback is not adoption, usage, security effectiveness, production readiness, or ecosystem acceptance evidence. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Keep pilot pack short, structured, and separate from release readiness claims; verification is structural readback of changed docs/templates. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence that pilot starts after release publication and does not block release readiness. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: revert pilot docs/templates only. <!-- sdd-owner: implementation -->

## Human-Controlled Delivery Gates

- [ ] Approve one issue per implementation Work Unit before its apply, including labels and scope. Completed issues govern delivered WUs; approved issue `#45` governs only comparison integration A-E, while later release Work Units still require their own approvals. <!-- sdd-owner: parent -->
- [x] Approved the corrected 28-PR total chain forecast after splitting #45 comparison redaction into B1 DTO validation and B2 privacy/fuzz/docs; no `size:exception` is authorized. <!-- sdd-owner: parent -->
- [x] Approved issue `#21` as the coherent planning authority for Planning PRs A-D; issue `#20` remains limited to public identity implementation. <!-- sdd-owner: parent -->
- [ ] Approve repository metadata mutations outside source files, including GitHub description, topics, homepage, and license metadata. <!-- sdd-owner: parent -->
- [ ] Confirm the exact OpenCode version/evidence boundary before Work Unit 3 claims support. <!-- sdd-owner: parent -->
- [ ] Confirm release candidate evidence is complete before creating tag `v0.1.0-alpha.1`. <!-- sdd-owner: parent -->
- [ ] Create and publish the GitHub prerelease, upload Linux amd64 binary and `checksums.txt`, and verify published artifacts manually. <!-- sdd-owner: parent -->
- [ ] Start post-release pilot outreach only after the prerelease is published and artifact/checksum links are known. <!-- sdd-owner: parent -->
