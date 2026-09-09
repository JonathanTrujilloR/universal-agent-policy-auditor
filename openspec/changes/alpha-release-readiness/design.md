# Design: Alpha Release Readiness

## Decision summary

Ship `v0.1.0-alpha.1` as a conservative, local, read-only Go CLI alpha for one evidence-backed surface: OpenCode on Linux amd64. The design adds release and application boundaries around the existing source, adapter, model, and support packages without silently expanding semantics. Apache-2.0 is the public identity. Claude Code remains unsupported while issues `#11` and `#13` are blocked. Cross-client comparison is admitted only through the approved fail-closed semantics from issue `#19`; this change does not absorb or redefine `#19`.

```text
cmd/auditor
  -> internal/app request/result boundary
  -> internal/source read-only selection
  -> internal/adapter/opencode resolution
  -> support evidence gate
  -> internal/model facts
  -> internal/compare from #19 only, when available and requested
  -> internal/redact safe report
  -> internal/render text/json
```

## Current-state constraints

The repository already has these usable boundaries:

| Area | Existing boundary | Conservative design consequence |
| --- | --- | --- |
| `internal/source` | Declarative `Plan`, allowlisted environment capture, canonical root checks, bounded `ReadFS`, selected source identities, and no write/process methods. | Reuse as the only target filesystem reader, but preserve adapter-declared source ordering before using it for precedence. |
| `internal/adapter/opencode` | OpenCode discovery plan and fail-closed resolver gated by `support.Matrix`. | OpenCode can be the first target, but no exact version is supported until evidence rows, fixtures, and conformance pass. |
| `support` | Empty matrix and registries plus validation requiring version, construct, fixture, evidence, and semantic flags. | Release code must not hardcode support; it must load checked-in support metadata and report unresolved/unsupported until populated. |
| `internal/model` | Permissions, effects, conditions, extensions, provenance digests, findings, completeness, and resolution traces. | Use as semantic facts, not as the JSON wire schema or release claim surface. |
| Claude conformance package | Maintainer-only evidence-domain work exists separately from normal auditor execution. | Do not import or use it for alpha support, docs, release claims, or normal CLI behavior. |

Missing release boundaries are intentional gaps today: `cmd/auditor`, `internal/app`, comparison package integration, redaction package, rendering package, release version metadata, deterministic output schema, stable exit categories, and release workflow evidence.

## CLI boundary

`cmd/auditor` is transport only. It parses arguments, validates obvious request shape, injects build version metadata, selects stdout/stderr streams, calls `internal/app`, renders the safe result, and maps the app category to the documented exit code. It must not parse target configuration semantics, read target files directly, mutate target files, call Git/GitHub, run OpenCode, run Claude Code, or perform comparison logic itself.

Recommended minimal alpha syntax:

```text
auditor version [--format text|json]
auditor audit opencode --root <dir> --config <path> --opencode-version <version> [--format text|json] [--no-color]
auditor compare --format text|json <#19-defined operands>
```

The release-blocking command is `audit opencode`. `compare` is present only if the approved issue `#19` implementation exists; otherwise it returns `unsupported_or_incomplete` and the release docs must not market cross-client support. `--config` is explicit local input. `--root` defines the allowed read boundary. `--opencode-version` is an evidence key, not a trust assertion; if the support matrix does not admit that exact version, the run is incomplete/unsupported.

Stable exit categories:

| Code | Category | Meaning |
| ---: | --- | --- |
| 0 | `complete_no_findings` | Supported input, complete processing, no blocking findings. |
| 1 | `complete_with_findings` | Supported input was processed completely and deterministic findings/differences exist. |
| 2 | `unsupported_or_incomplete` | Unsupported version, missing evidence, malformed/unknown/lossy/ambiguous input, unresolved defaults, unavailable #19 comparison, or incomplete source selection. |
| 3 | `invalid_request` | CLI arguments, root/config relationship, output format, or operand shape is invalid before target analysis. |
| 4 | `operational_failure` | Internal error, filesystem failure outside modeled source findings, schema/render failure, or redaction failure. |

The category names are part of the alpha contract. Numeric values may be documented as alpha-stable for automation, with later incompatible changes requiring a schema/release-note break.

## Application request/result contract

`internal/app` owns the use-case boundary and should expose a small typed request/result contract, independent of CLI flag names and JSON rendering.

| Request field | Meaning | Conservative rule |
| --- | --- | --- |
| mode | `version`, `audit`, or `compare`. | Unknown modes are `invalid_request`. |
| target | `opencode` for alpha audit. | `claudecode` is visibly unsupported while `#11` and `#13` are blocked. |
| root | Canonical local root allowed for target reads. | Missing or unresolved root is invalid; sources outside it are incomplete. |
| explicitConfig | User-selected local config path. | Required for the alpha audit path to avoid broad discovery claims. |
| targetVersion | Exact target version/config-version evidence key. | Missing or unsupported version yields `unsupported_or_incomplete`, not success. |
| requestedCapabilities | Optional explicit capabilities to resolve. | Missing defaults remain unresolved unless evidence proves defaults. |
| formatIntent | Text or JSON preference from transport. | App records intent only; renderers own bytes. |
| buildInfo | Tool version, commit, date, schema set. | Injected by `cmd/auditor`; never discovered from Git at runtime. |

The app result should contain semantic status, exit category, support status, selected-source summaries, adapter reports, comparison reports when #19 is available, findings, limitations, and provenance references. It should not contain raw config values. Rendering-specific DTOs are produced only after redaction.

Aggregation is fail-closed: any source-selection finding, support validation failure, unsupported target, unknown permission-bearing construct, lossy comparison state, redaction failure, render/schema failure, or internal invariant violation prevents `complete_no_findings`. Incomplete/unsupported semantic conditions map to code 2; redaction and internal failures map to code 4 because unsafe output must not be emitted as a normal audit.

## Source, adapter, support, and model flow

For `audit opencode`, `internal/app` builds `opencode.DiscoveryPlan(root, []string{explicitConfig})`, captures only plan-allowlisted environment keys, and calls `source.Execute` through the read-only filesystem port. The app must preserve the adapter-declared source precedence order. If the current `SelectionReport` cannot prove that order after selection, the app must mark the run incomplete rather than relying on identity sorting for precedence-sensitive resolution.

The app then loads checked-in support metadata, calls `opencode.ResolveWithOptions`, and converts adapter findings into app findings without changing their semantics. The current support registries are empty, so any alpha build before the evidence Work Unit must report OpenCode as unsupported/unresolved. The design does not choose an OpenCode version; the evidence Work Unit must add an exact version, stable documentation/source or executable evidence, registered fixtures, and passing conformance evidence before release docs may say that version is supported.

`internal/model` remains the semantic core for permissions and traces. Do not add release claims directly to `model.Permission`. Support status, exit category, selected source summaries, schema versions, and release limitations belong in app/report DTOs unless a later model change proves a reusable semantic concept.

## Comparison boundary for issue #19

Comparison is not implemented by this change. The alpha release may integrate comparison only by importing and calling the package delivered by approved issue `#19` after both sides are already normalized into existing model facts or #19-defined operands. This design does not define comparison algorithms, equivalence rules, or operand schema beyond the dependency boundary.

If #19 is absent, unavailable, or returns unsupported/ambiguous/lossy/not-comparable, the app must surface that state and use a non-success category. Release docs may say comparison is fail-closed only when #19 evidence exists; they must not claim complete OpenCode-versus-Claude comparison while Claude Code is unsupported.

## Redaction and rendering boundaries

Redaction happens before rendering and owns the privacy gate. Renderers accept only a redacted safe report, never raw adapter reports, raw source bytes, raw environment values, raw absolute paths, raw config values, credentials, tokens, usernames, hostnames, or arbitrary diagnostic strings.

Recommended package responsibilities:

| Package | Responsibility |
| --- | --- |
| `internal/redact` | Convert app results and model provenance into safe references, digests, closed reason codes, relative-or-opaque source labels, and bounded messages. Fail closed if a value cannot be proven safe. |
| `internal/render/text` | Concise deterministic human output, blocking uncertainty first, no color dependency, no raw values. |
| `internal/render/json` | Versioned deterministic JSON DTOs, schema ownership, byte-stability tests, sorted dynamic collections, no map-dependent ordering. |

The deterministic JSON schema should be owned by the JSON renderer package, not by `internal/model`. Initial schema id: `auditor-report/v1alpha1`. Required fields: schema version, tool version, target, requested target version, support status, completeness, exit category, selected source digests/safe labels, permissions or differences, findings, provenance digests, and limitations. Optional fields must be explicitly omitted or represented as empty arrays according to the schema; no implicit clock, hostname, username, absolute workspace path, or network-derived field is allowed.

## Version injection

Use build-time version metadata injected into the CLI binary. The release pipeline supplies `v0.1.0-alpha.1`, commit, build date, and target platform through linker variables or an equivalent compile-time mechanism. Development builds default to safe `dev` values and must not claim release identity. Runtime code must not call Git, GitHub, the network, OpenCode, or Claude Code to discover its own version.

`auditor version --format json` should render build metadata through the same redaction/rendering discipline and should be included in release smoke verification.

## No-mutation enforcement

The alpha has no write path. Enforcement layers:

| Layer | Control |
| --- | --- |
| CLI | No `fix`, `write`, `format`, `delete`, `remediate`, or `compile` command in alpha. Unknown mutation-like commands are invalid. |
| App | Request modes are read-only; no writer, process runner, network client, Git client, or config update port exists. |
| Source | Use the existing `ReadFS` interface only; it exposes `ReadFile`, `Stat`, and `EvalSymlinks`. |
| Adapters | Receive immutable selected bytes and support metadata; they do not perform I/O or process execution. |
| Tests | Strict TDD starts with no-mutation tests using `t.TempDir()`, content/metadata snapshots, and interface-shape checks. |
| Docs/output | Every release surface states modeled static configuration semantics and no target mutation. |

## OpenCode evidence admission

OpenCode support is admitted only when all of these evidence boundaries are true in checked-in repository state:

| Boundary | Required evidence |
| --- | --- |
| Exact version | A precise OpenCode version/config-version token used by the support matrix. |
| Stable authority | Official documentation, upstream source, or executable evidence referenced by a stable evidence ID. |
| Fixture registration | A fixture ID in the registry that covers permission-bearing constructs, defaults, matcher behavior, precedence, source order, malformed input, and runtime-dependent behavior. |
| Conformance pass | Focused tests prove the fixture and adapter behavior for the exact version. |
| Release wording | README, support matrix, release notes, and CLI output cite only the admitted version and still state static-analysis limitations. |

Until all rows are true, the CLI may run structurally but must report unsupported/unresolved status. This design intentionally does not pick an OpenCode version.

## Linux amd64 release workflow design

Linux amd64 is the only binary platform for `v0.1.0-alpha.1`. Additional platforms, package-manager distribution, signing, and provenance attestations remain deferred.

Trusted release sequence:

1. Land all release Work Units through reviewed slices, each within the `<=400` authored-line budget or explicitly split.
2. Verify repository identity and release docs consistently project Apache-2.0.
3. Run formatting check, vetting, `go test ./...`, Linux amd64 build, checksum generation, and Linux amd64 smoke checks from the release candidate commit.
4. Create the trusted source tag `v0.1.0-alpha.1` only after evidence is recorded.
5. Build the Linux amd64 binary from that tag with release version injection.
6. Generate `checksums.txt` for published artifacts.
7. Publish a GitHub prerelease with the binary, checksums, source tag/archive, Apache-2.0 identity, limitations, and no unsupported claims.

Smoke evidence should include at least `auditor version --format json` and an `audit opencode` run against the checked-in supported fixture for the admitted OpenCode version. The smoke record must not include secrets, private paths, raw local configs, or unevidenced platform claims.

## Public identity and release/pilot separation

Apache-2.0 must appear consistently in root `LICENSE`, README positioning, release notes, release artifacts, and repository metadata. No public surface may conflict with Apache-2.0 or omit license identity for the alpha.

Release artifacts are the GitHub prerelease, Linux amd64 binary, source tag/archive, `checksums.txt`, release notes, README/install/support documentation, and checked-in evidence references. Pilot artifacts are feedback templates or Discussions/issue-template material for post-release evidence collection. Pilot material must be clearly post-release, warn users not to share secrets or private local context, and must not be represented as adoption, usage, security effectiveness, or ecosystem acceptance.

## Work Unit and rollback boundaries

These are architecture work boundaries for the tasks phase to convert into implementation tasks; they are not implementation checklists.

| Work Unit | Issue boundary | Dependencies | Rollback boundary | Budget risk |
| --- | --- | --- | --- | --- |
| Public project identity | Apache-2.0 license, README positioning, metadata, claim guardrails. | License decision already confirmed. | Revert docs/metadata only. | Low. |
| CLI/app shell | `cmd/auditor`, `internal/app`, request/result, version/help/exit mapping, no-mutation tests. | Existing source/adapter/model/support packages. | Remove CLI/app shell without touching adapters/support. | Medium. |
| OpenCode evidence gate | Exact version evidence, matrix rows, fixtures, conformance evidence, support-status behavior. | Version/evidence decision; no Claude dependency. | Revert support entries and OpenCode fixtures/tests together. | Medium. |
| Output/redaction/rendering | Safe report, `auditor-report/v1alpha1`, deterministic text/JSON, redaction-before-rendering. | App result contract and model provenance. | Revert redaction/render packages and schema docs. | High; split JSON/redaction from text rendering if needed. |
| #19 comparison integration | Wire approved fail-closed comparison into app/rendering without redefining #19. | Completed approved #19 package. | Revert integration only; keep #19 core intact. | Medium. |
| CI/release build | Linux amd64 build, checksums, smoke workflow, release evidence recording. | CLI/app shell and supported fixture. | Revert workflow/release scripts; no target config state exists. | Medium. |
| Release docs | Install, usage, support matrix, limitations, release notes grounded in evidence. | Facts from prior Work Units. | Revert docs/release notes only. | Low. |
| Pilot pack | Post-release feedback template/discussion guidance and privacy warnings. | Release docs; pilot starts after release. | Revert pilot templates only. | Low. |

Each Work Unit should keep tests and docs with the behavior they verify, follow strict TDD, and remain below `<=400` authored changed lines. If a unit approaches the budget, split into chained PR slices rather than accepting a larger review. This change must not absorb issue `#19` implementation, blocked issue `#11`, or blocked issue `#13`.

## Missing interfaces and conservative behavior

| Missing interface/semantic | Why it matters | Conservative behavior |
| --- | --- | --- |
| Ordered selected sources | Adapter precedence depends on stable source order, but selected source identity sorting may erase plan order. | Preserve explicit source order in app/source before precedence-sensitive resolution; otherwise mark source order ambiguous/incomplete. |
| App-level support status | Model completeness is not the same as release support status. | Add app/report status; unsupported versions cannot become model success. |
| Finding severity/category | Current findings have code/message only. | App maps known blocking conditions to exit categories; unknown finding categories block success. |
| JSON report schema | Model types are not stable wire DTOs. | Renderer-owned schema `auditor-report/v1alpha1`; no direct serialization of internal model as public API. |
| Redacted evidence references | Model provenance stores digests but locations may include sensitive paths. | Redaction converts locations to safe labels/digests before render. |
| Comparison result model | No current compare package exists. | Wait for #19; unavailable comparison is unsupported/incomplete. |
| Release version metadata | No current binary identity boundary exists. | Inject at build time; dev builds say `dev` and cannot claim release. |
| Operational error taxonomy | Filesystem and internal failures need stable exits. | App maps only proven semantic incompleteness to code 2; unsafe/internal failures to code 4. |

## Security and privacy rationale

The release protects users by refusing false assurance. Static configs are untrusted data; the tool reads bounded explicit local inputs, never executes config content, never runs target clients, never phones home, and never mutates user files. Unknown permission-bearing constructs, unsupported versions, ambiguous source order, unresolved defaults, lossy comparison, and redaction uncertainty all block unqualified success.

Privacy is protected by minimizing capture and moving redaction before rendering. Provenance remains useful through safe labels and `sha256:` digests instead of raw values. Pilot guidance must repeat the same privacy rule because pilot participants may otherwise paste secrets, private paths, or raw configs into public feedback.

## Failure modes

| Failure mode | User-visible behavior |
| --- | --- |
| Missing Apache-2.0 identity on any release surface | Release readiness blocked. |
| OpenCode version lacks checked-in evidence | Audit reports unsupported/unresolved; no support claim. |
| Claude requested | Unsupported/incomplete while `#11` and `#13` are blocked. |
| Config unreadable, outside root, duplicate, too large, or source order ambiguous | Visible finding and non-success category. |
| Malformed or unknown permission-bearing OpenCode construct | Visible finding and non-success category. |
| Runtime-dependent or unresolved default semantics | Incomplete result with static-analysis limitation. |
| #19 comparison unavailable or not-comparable | Visible unsupported/not-comparable result; no cross-client claim. |
| Redaction cannot prove output safe | Rendering stops with operational failure; unsafe content is not emitted. |
| Linux amd64 build/checksum/smoke evidence missing | Release publication blocked. |

## Alternatives rejected

| Alternative | Rejection reason |
| --- | --- |
| Claim OpenCode support from current docs without exact version evidence | Violates evidence-gated support and risks stale semantics. |
| Ship Claude or OpenCode-versus-Claude success wording in alpha | Blocked by `#11`/`#13` and would imply unsupported cross-client maturity. |
| Serialize `internal/model` directly as public JSON | Couples internal evolution to release schema and risks leaking raw locations. |
| Redact inside renderers | Makes safety format-specific and can leak through one renderer while another is safe. |
| Add mutation/remediation commands for usability | Breaks the audit-only alpha trust boundary. |
| Build platforms beyond Linux amd64 | No independent build and smoke evidence is in scope. |
| Treat pilot materials as release adoption evidence | Fabricates social proof and confuses feedback collection with validated usage. |

## Verification strategy for later implementation

Implementation must follow strict TDD. Use focused Go tests for app aggregation, support gating, no-mutation enforcement, version injection, redaction safety, deterministic JSON, stable exit mapping, and OpenCode fixture conformance. Use `t.TempDir()` and read-only fakes around filesystem boundaries. Run narrow package tests before `go test ./...`; release workflow evidence additionally records Linux amd64 build, checksum generation, and smoke results. Documentation-only slices use structural readback and claim-boundary review.
