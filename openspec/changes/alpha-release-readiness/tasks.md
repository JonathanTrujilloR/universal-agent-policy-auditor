# Tasks: Alpha Release Readiness

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 2,340-3,340+ authored lines total: 940 current planning lines plus 1,400-2,400 implementation lines and this correction |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | Planning PR A config/exploration/preproposal/proposal → Planning PR B alpha-release spec → Planning PR C design → Planning PR D corrected tasks → PR 5 public identity → PR 6 CLI/app shell → PR 7 OpenCode evidence gate → PR 8 JSON/redaction → PR 9 human renderer → PR 10 #19 integration → PR 11 CI/release build → PR 12 release docs/prerelease → PR 13 post-release pilot pack |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

Recommended next PR slice: Planning PR D, the final planning slice. Planning PRs A-C are merged and approved issue `#21` governs the planning chain. Do not start source/docs apply until Planning PR D merges. Chain implication: use `stacked-to-main`; each planning and implementation slice targets `main`, references one coherent approved issue, stays under 400 authored changed lines, and records rollback/evidence independently. No `size:exception` is authorized.

## Task Ordering and Completion Rules

- Do not start implementation apply for any Work Unit until the planning-artifact PR path is approved and delivered, and exactly one coherent Work Unit issue is approved with `status:approved` and exactly one `type:*` label. Issue `#19` governs comparison core, issue `#20` governs public identity, and issue `#21` governs planning only. <!-- sdd-owner: implementation -->
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
- [ ] Planning PR D: Deliver `openspec/changes/alpha-release-readiness/tasks.md` as the corrected task-plan slice; keep the final authored changed-line count for this PR below 400, or split tasks into an additional planning PR before apply. <!-- sdd-owner: implementation -->
- [ ] Record planning delivery evidence: approved planning issue number or explicit maintainer-approved issue path, `Refs`/closing-reference policy, stacked-to-main order, changed-line counts, structural readback, and rollback boundary for each planning PR. <!-- sdd-owner: implementation -->
- [ ] Keep implementation Work Units 1-9 pending until planning PRs are delivered and the maintainer approves the revised chain count and planning issue path. <!-- sdd-owner: implementation -->

## Work Unit 1: Public Project Identity, Legal, and Docs Baseline

Issue gate: approved issue `#20` (`status:approved`, `type:docs`). Forecast: 120-250 authored lines. File surfaces: `LICENSE`, `README.md`, release-facing docs under `docs/` or `.github/` discovered during apply, repository metadata checklist documentation only.

- [ ] RED: Add/adjust documentation checks or review fixtures that fail when Apache-2.0 identity is absent from `LICENSE`, `README.md`, and release-facing documentation; focused command: `go test ./...` if checks are Go-based, otherwise structural readback of changed Markdown/license files. <!-- sdd-owner: implementation -->
- [ ] GREEN: Add Apache-2.0 `LICENSE`, README positioning, alpha limitation language, and metadata checklist text without claiming OpenCode version support, Claude support, runtime enforcement, package-manager install, signing/provenance, or adoption. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Add a claim-boundary case covering missing/conflicting license or forbidden release claims; focused command matches the RED check or structural readback. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Tighten docs for scannability and reviewer verification; full evidence: `gofmt -l .`, `go vet ./...`, `go test ./...`, and Markdown/license readback as applicable. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence that every release-facing surface says Apache-2.0 consistently and that Claude/OpenCode/support limitations are visible. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: revert `LICENSE`, `README.md`, and identity docs/metadata checklist only, with no source/runtime changes. <!-- sdd-owner: implementation -->

## Work Unit 2: CLI/App Shell

Issue gate: blocked until a dedicated CLI/app shell Work Unit issue is approved. Forecast: 250-390 authored lines. File surfaces: `cmd/auditor/**`, `internal/app/**`, narrow tests beside those packages, build metadata file if needed.

- [ ] RED: Add failing focused tests for `auditor version`, help/invalid request handling, stable exit categories, explicit `audit opencode --root --config --opencode-version`, and no mutation of `t.TempDir()` target files; focused command: `go test ./cmd/auditor ./internal/app`. <!-- sdd-owner: implementation -->
- [ ] GREEN: Implement transport-only `cmd/auditor` and typed `internal/app` request/result shell with build metadata defaults, exit category mapping, explicit input validation, and no writer/process/network/Git/GitHub ports. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Add unsupported `claudecode`, unsupported/missing OpenCode version, outside-root config, and mutation-like command cases; focused command: `go test ./cmd/auditor ./internal/app`. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Keep CLI parsing out of semantic packages and app results independent of renderer DTOs; full command: `gofmt -l . && go vet ./... && go test ./... && go build ./...`. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence for deterministic version/help behavior, numeric/category exits `0-4`, explicit local input, and unchanged target file snapshots. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: remove `cmd/auditor` and `internal/app` shell without touching existing `internal/source`, `internal/adapter`, `internal/model`, or `support` semantics. <!-- sdd-owner: implementation -->

## Work Unit 3: OpenCode Evidence Gate

Issue gate: blocked until a dedicated OpenCode evidence Work Unit issue is approved and an exact OpenCode version/evidence source is selected. Forecast: 250-380 authored lines. File surfaces: `support/**`, `internal/adapter/opencode/**`, checked-in fixtures under existing testdata/evidence directories discovered during apply, related tests.

- [ ] RED: Add failing conformance tests proving unsupported/unresolved status when exact version evidence, registry rows, fixtures, or semantic flags are missing; focused command: `go test ./support ./internal/adapter/opencode ./internal/app`. <!-- sdd-owner: implementation -->
- [ ] GREEN: Register only the selected exact OpenCode evidence boundary, fixtures, and support matrix rows needed for the alpha; do not hardcode support in CLI/app. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Cover malformed input, unknown permission-bearing constructs, defaults, matcher behavior, precedence/source order, and runtime-dependent behavior as unsupported/incomplete unless evidence proves them. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Keep evidence metadata declarative and adapter behavior fail-closed; full command: `gofmt -l . && go vet ./... && go test ./...`. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence citing the exact OpenCode version, stable authority, fixture IDs, and passing conformance command output before any support claim is allowed. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: revert support entries, evidence files, fixtures, and OpenCode conformance tests together. <!-- sdd-owner: implementation -->

## Work Unit 4: Output, Redaction, and Versioned JSON

Issue gate: blocked until a dedicated output/redaction/JSON Work Unit issue is approved. Forecast: 350-500 authored lines; split before 400, likely JSON/redaction first. File surfaces: `internal/redact/**`, `internal/render/json/**`, `internal/app/**` DTO handoff tests, schema docs if needed.

- [ ] RED: Add failing tests for redaction-before-rendering, no raw secrets/config values/private paths, schema `auditor-report/v1alpha1`, deterministic JSON field ordering, required fields, and safe provenance digests; focused command: `go test ./internal/redact ./internal/render/json ./internal/app`. <!-- sdd-owner: implementation -->
- [ ] GREEN: Implement safe report conversion and JSON renderer that accepts only redacted DTOs and emits schema version, tool version, target/version/support status, completeness, permissions or differences, findings, provenance digests, exit category, and limitations. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Add repeated-run determinism, map-order sorting, redaction failure as `operational_failure`, and unsupported/incomplete semantic cases; focused command: `go test ./internal/redact ./internal/render/json ./internal/app`. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Separate internal model types from public JSON DTOs and remove format-specific safety duplication; full command: `gofmt -l . && go vet ./... && go test ./...`. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence with sanitized fixture output proving no secret, credential, raw value, username, hostname, or absolute private path is emitted. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: revert `internal/redact`, `internal/render/json`, schema docs, and app DTO glue without changing adapter/model semantics. <!-- sdd-owner: implementation -->

## Work Unit 5: Human Renderer if Separately Needed

Issue gate: blocked until a dedicated human-renderer Work Unit issue is approved, unless merged into Work Unit 4 while staying under budget. Forecast: 120-240 authored lines. File surfaces: `internal/render/text/**`, CLI renderer selection tests, docs snippets if paired with behavior.

- [ ] RED: Add failing tests for concise deterministic text output, uncertainty first, no color dependency with `--no-color`, limitations visibility, and no raw sensitive values; focused command: `go test ./internal/render/text ./cmd/auditor`. <!-- sdd-owner: implementation -->
- [ ] GREEN: Implement text renderer over redacted safe report only and wire CLI format selection without changing app semantics. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Cover unsupported Claude, unsupported OpenCode evidence, incomplete source selection, and findings/differences summaries. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Keep human wording stable, short, and claim-safe; full command: `gofmt -l . && go vet ./... && go test ./...`. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence with text snapshots/readback showing static-analysis limitations and unsupported states. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: remove `internal/render/text` and CLI text wiring while retaining JSON/redaction behavior. <!-- sdd-owner: implementation -->

## Work Unit 6: #19 Comparison Integration

Issue gate: issue `#19` is approved for comparison core; this Work Unit may proceed only after #19 implementation exists and must not reimplement or redefine it. Forecast: 180-330 authored lines. File surfaces: `internal/app/**`, `cmd/auditor/**`, `internal/render/**`, imports/calls to the #19 package discovered during apply, integration tests.

- [ ] RED: Add failing tests that `compare` delegates to the #19 package when available and returns `unsupported_or_incomplete` for unavailable, ambiguous, lossy, unsupported, or not-comparable states; focused command: `go test ./internal/app ./cmd/auditor`. <!-- sdd-owner: implementation -->
- [ ] GREEN: Wire app/CLI/rendering to #19-defined operands and results only, with no new comparison algorithms, equivalence rules, or Claude support claims. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Cover docs/output wording so fail-closed comparison does not imply complete OpenCode-versus-Claude support while Claude remains unsupported. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Keep #19 integration removable and separate from core #19 package implementation; full command: `gofmt -l . && go vet ./... && go test ./...`. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence linking to #19 behavior and showing unavailable comparison maps to visible non-success. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: revert comparison wiring in app/CLI/renderers only; keep #19 core intact. <!-- sdd-owner: implementation -->

## Work Unit 7: CI, Release Build, Checksums, and Smoke

Issue gate: blocked until a dedicated CI/release-build Work Unit issue is approved. Forecast: 220-380 authored lines. File surfaces: `.github/workflows/**`, `scripts/**` or `tools/**` release helpers, `cmd/auditor` build metadata integration, release evidence docs under `docs/` if needed.

- [ ] RED: Add failing workflow/script tests or dry-run checks for formatting, vetting, tests, Linux amd64 build, version injection, checksum generation, and smoke command capture; focused command: relevant script dry-run plus `go test ./...`. <!-- sdd-owner: implementation -->
- [ ] GREEN: Add CI/release readiness workflow or scripts for `gofmt -l .`, `go vet ./...`, `go test ./...`, `GOOS=linux GOARCH=amd64 go build`, `checksums.txt`, and smoke commands without publishing. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Add smoke coverage for `auditor version --format json` and `audit opencode` against the checked-in supported fixture only after Work Unit 3 evidence exists. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Keep release helpers deterministic and free of package-manager, signing, provenance, extra-platform, GitHub publication, or tag creation side effects. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence for formatting, vetting, tests, build, checksum generation, and Linux amd64 smoke from the release candidate commit. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: revert workflow/scripts/evidence docs only; no target configuration state exists. <!-- sdd-owner: implementation -->

## Work Unit 8: Release Docs and Prerelease Preparation

Issue gate: blocked until a dedicated release-docs Work Unit issue is approved and prior evidence Work Units have landed. Forecast: 180-320 authored lines. File surfaces: `README.md`, `docs/**`, `CHANGELOG.md` or release notes file discovered during apply, support matrix documentation.

- [ ] RED: Add documentation claim-boundary checks or structural review checklist that fails when release docs omit Apache-2.0, static-analysis limitations, Linux amd64 evidence, checksums, unsupported Claude, unsupported package managers, or OpenCode evidence references. <!-- sdd-owner: implementation -->
- [ ] GREEN: Write install, usage, support matrix, limitations, release notes draft, checksum instructions, and smoke evidence references grounded only in checked-in facts. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Add negative claim checks for production readiness, security/compliance guarantees, runtime enforcement, package-manager install, unsupported platforms, signing/provenance, and adoption/pilot success. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Make docs reviewer-friendly with quick path, evidence table, limitations, and rollback/publication notes; full verification includes structural readback plus `gofmt -l . && go vet ./... && go test ./...` if code-adjacent references changed. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence that release remains blocked until exact OpenCode evidence and Linux amd64 smoke evidence exist; do not invent a version. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: revert release docs/release notes/support docs only. <!-- sdd-owner: implementation -->

## Work Unit 9: Post-Release Pilot Pack

Issue gate: blocked until a dedicated pilot-pack Work Unit issue is approved; pilot execution starts only after the alpha release exists. Forecast: 80-180 authored lines. File surfaces: `.github/ISSUE_TEMPLATE/**`, `.github/DISCUSSION_TEMPLATE/**` if present, `docs/pilot/**` or equivalent docs discovered during apply.

- [ ] RED: Add structural/readback check that pilot materials position feedback as post-release only and forbid sharing secrets, credentials, private paths, raw sensitive configs, or unredacted local context. <!-- sdd-owner: implementation -->
- [ ] GREEN: Add a feedback template or discussion guide with exact artifact/checksum fields, sanitized command/output guidance, unsupported construct prompts, false-assurance prompts, and privacy warnings. <!-- sdd-owner: implementation -->
- [ ] TRIANGULATE: Add wording checks that pilot feedback is not adoption, usage, security effectiveness, production readiness, or ecosystem acceptance evidence. <!-- sdd-owner: implementation -->
- [ ] REFACTOR: Keep pilot pack short, structured, and separate from release readiness claims; verification is structural readback of changed docs/templates. <!-- sdd-owner: implementation -->
- [ ] Record acceptance evidence that pilot starts after release publication and does not block release readiness. <!-- sdd-owner: implementation -->
- [ ] Record rollback boundary: revert pilot docs/templates only. <!-- sdd-owner: implementation -->

## Human-Controlled Delivery Gates

- [ ] Approve one issue per Work Unit before apply, including labels and scope; only #19 is currently approved and only as the comparison dependency. <!-- sdd-owner: parent -->
- [x] Approved the corrected 13-PR total chain forecast: 4 planning PRs followed by 9 implementation PRs under `ask-on-risk` using `stacked-to-main`; no `size:exception` is authorized. <!-- sdd-owner: parent -->
- [x] Approved issue `#21` as the coherent planning authority for Planning PRs A-D; issue `#20` remains limited to public identity implementation. <!-- sdd-owner: parent -->
- [ ] Approve repository metadata mutations outside source files, including GitHub description, topics, homepage, and license metadata. <!-- sdd-owner: parent -->
- [ ] Confirm the exact OpenCode version/evidence boundary before Work Unit 3 claims support. <!-- sdd-owner: parent -->
- [ ] Confirm release candidate evidence is complete before creating tag `v0.1.0-alpha.1`. <!-- sdd-owner: parent -->
- [ ] Create and publish the GitHub prerelease, upload Linux amd64 binary and `checksums.txt`, and verify published artifacts manually. <!-- sdd-owner: parent -->
- [ ] Start post-release pilot outreach only after the prerelease is published and artifact/checksum links are known. <!-- sdd-owner: parent -->
