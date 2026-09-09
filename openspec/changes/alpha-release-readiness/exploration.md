# Exploration: alpha-release-readiness

## Executive finding

The safest `v0.1.0-alpha.1` is not a broad cross-client security claim. It is an installable, local-only Go CLI alpha that truthfully demonstrates the evidence-gated audit model on the smallest supported surface, publishes clear limitations, and refuses unsupported semantics instead of implying protection.

Current repository evidence supports a planning path but not a release claim yet: `origin/main` at `ab855161d4b800a76129ed229412e8c636e296af` reportedly passes `gofmt`, `go vet`, `go test ./...`, and build, but the public project has no README, LICENSE, repository metadata, CI, tags, releases, `cmd/main`, install instructions, supported target version, or adoption evidence. The checked-in support matrix and evidence registries are intentionally empty, so no OpenCode or Claude Code support claim is currently true.

## Minimal usable alpha outcome

A truthful first alpha should let a user:

1. discover the repository and understand exactly what the tool does and does not do;
2. install or download a versioned CLI artifact for a documented platform;
3. run a bounded command against explicit local fixture/config input;
4. receive deterministic human and machine-readable output that includes completeness, support status, findings, provenance, and redacted evidence;
5. understand that results describe modeled static configuration semantics only, not native runtime enforcement or organizational compliance.

The minimum useful journey can be OpenCode-only if it is framed as an alpha capability slice, not as the full PRD promise of OpenCode + Claude comparison. If comparison exists before Claude support unblocks, it should be limited to deterministic internal comparison semantics over modeled records or fixtures and must not advertise Claude support.

## First supported target assessment

OpenCode-only appears to be the safest first supported target because:

- `#11` Claude support and `#13` Claude harness are blocked and must not be dependencies for first alpha;
- the repository already contains an OpenCode adapter surface with fail-closed support gating, bounded discovery planning, parsing tests, precedence modeling tests, malformed/unknown construct tests, and unresolved default behavior;
- the support gate explicitly requires a version-pinned matrix entry with official documentation, upstream source or executable evidence, and a passing fixture before any support claim;
- `#19` deterministic fail-closed comparison core is approved and independent, so comparison infrastructure can progress without Claude runtime support.

Evidence gaps before claiming OpenCode support:

- exact supported OpenCode version or config version must be named;
- support matrix entry must cite stable evidence and a passing fixture;
- defaults, matcher behavior, precedence, source order, runtime-dependent behavior, and unknown permission-bearing constructs must have fixture coverage;
- support documentation must state platform scope and static-analysis limitations;
- release notes and README must avoid implying runtime enforcement.

## Legal/license decision options

A release should not ship without a maintainer-selected license. This exploration does not choose it.

Options to present upstream:

- permissive OSS license such as MIT or Apache-2.0: easier adoption and packaging, with different patent/license-text implications;
- copyleft license such as GPL-family: stronger sharing requirements, potentially higher adoption friction for commercial users;
- no public release until license is chosen: legally safer than publishing ambiguous reuse terms, but blocks external pilot distribution.

The selected license must be reflected in repository root `LICENSE`, package metadata, release artifacts, and README wording.

## Required release-readiness surfaces

- CLI/application: add a real `cmd/main` entrypoint, version output, command help, stable exit-code contract, explicit input paths, and no mutation of target configs.
- Support evidence: populate only evidence-backed OpenCode support entries; keep Claude entries absent or unsupported until blockers resolve.
- Comparison: use the approved deterministic fail-closed core from `#19`; classify unsupported/ambiguous/not-comparable states without unqualified success.
- Redaction: redact secrets and sensitive values by default in terminal and JSON output; keep provenance useful without leaking raw config values.
- Rendering: provide concise human output plus versioned JSON; deterministic ordering is required.
- CI: run at least `gofmt` check, `go vet ./...`, `go test ./...`, and build on documented release platforms.
- Docs: README, install instructions, usage examples, support matrix, limitations, contribution/testing notes, and release notes.
- Release: semantic prerelease tag `v0.1.0-alpha.1`, checksums, platform matrix, and truthful changelog.
- Repository metadata: description, homepage if available, topics, and license.

## Distribution options for a Go binary

Initial release can use GitHub Releases with prebuilt binaries and checksums. Platform scope should be intentionally small and verified, for example Linux amd64 first, or Linux/macOS/Windows amd64 only if CI/build evidence exists for each. Do not list platforms that are not built and smoke-tested.

Possible distribution maturity levels:

- source-only tag: lowest release machinery, but weakest install journey;
- GitHub Release binaries plus `checksums.txt`: good alpha default if build is reproducible enough;
- package managers/homebrew/docker: later, after naming, license, support scope, and release cadence stabilize.

Checksums are required for published binaries. Signing/provenance can be documented as future hardening unless the maintainer wants it in alpha.

## Work Unit boundaries and dependencies

| Work Unit | Purpose | Dependencies | Rollback boundary | Line-budget risk |
|---|---|---|---|---|
| WU1 public project identity | README skeleton, license decision application, metadata checklist, truthful positioning | maintainer license decision | docs/metadata only | Low |
| WU2 CLI shell | `cmd/main`, command routing, version/help/exit-code contract | none after current build baseline | CLI entrypoint files | Medium |
| WU3 OpenCode evidence gate | support matrix entry, fixtures, adapter evidence tests | OpenCode version/evidence decision | support files + OpenCode fixture tests | Medium |
| WU4 output/redaction/rendering | human summary, JSON schema, deterministic ordering, redaction defaults | model/output decisions; WU2 | rendering/redaction package | High, likely split |
| WU5 comparison core integration | integrate approved `#19` fail-closed comparison into CLI/output | `#19` implementation | comparison package + CLI wiring | Medium |
| WU6 CI and release build | CI workflow, cross-platform build, checksums | WU2 and platform decision | CI/release scripts | Medium |
| WU7 release docs | install path, usage examples, support matrix, limitations, release notes | WU2-WU6 facts | docs only | Low |
| WU8 external pilot pack | pilot instructions, evidence template, feedback criteria | released alpha | docs/templates only | Low |

Any unit approaching 400 authored changed lines should become a separate chained PR slice. WU4 is the most likely to exceed the budget and should probably split into JSON schema/redaction first, then human rendering. WU3 can proceed independently of Claude. WU5 can proceed while Claude is blocked if it only integrates deterministic comparison semantics and does not claim two-client support.

## Work that can proceed while Claude RDD is blocked

- public README/license/repository metadata preparation;
- CLI entrypoint and version/help/exit behavior;
- OpenCode-only support evidence and fixtures;
- deterministic fail-closed comparison core from `#19`;
- redaction and output schema;
- CI/build/checksum release machinery;
- docs that explicitly state Claude is unsupported in this alpha;
- external pilot design that asks for structured feedback without promising ecosystem acceptance.

Do not depend on `#11` or `#13` for first alpha. Do not claim Claude support, Claude equivalence, or cross-client coverage until those blockers are resolved and evidence is in the support matrix.

## External pilot design

The pilot should be separate from the release. Its goal is evidence collection, not adoption proof.

Pilot candidate profile:

- users already configuring OpenCode locally;
- willing to run a local CLI on non-sensitive or redacted configs;
- willing to report unsupported constructs, confusing findings, and false-assurance concerns.

Evidence needed before later ecosystem application:

- release artifact URL and checksum;
- documented support matrix and limitations;
- passing CI/build/test evidence;
- examples of deterministic output on sanitized fixtures;
- pilot feedback with issue links or structured summaries;
- no fabricated usage metrics, stars, installs, or enterprise adoption claims.

Pilot materials should include a privacy warning, redaction guidance, exact command to run, expected outputs, feedback template, and forbidden data to share.

## Explicit non-goals for `v0.1.0-alpha.1`

- no config writing, formatting, deletion, or remediation;
- no canonical policy authoring DSL;
- no runtime enforcement, sandbox, proxy, firewall, or compliance certification;
- no Claude support unless blockers are resolved and evidence is added before release;
- no Codex/Cursor/additional target support;
- no telemetry, hosted service, account, or cloud control plane;
- no package-manager promise unless packaging is actually implemented;
- no adoption, security effectiveness, or ecosystem acceptance claim without evidence.

## Product decisions unresolved before proposal

1. Which license should the maintainer choose?
2. Is OpenCode-only acceptable for `v0.1.0-alpha.1`, with Claude explicitly unsupported?
3. Which exact OpenCode version/config version is the first support target?
4. What initial platform matrix is acceptable for release binaries?
5. Is source-only release acceptable, or are prebuilt binaries required for alpha?
6. Is signing/provenance required now, or are checksums sufficient for alpha?
7. What is the exact CLI command shape for the first journey?
8. What minimum JSON schema fields are release-blocking?
9. What evidence threshold allows the README to say "supports OpenCode" instead of "experimental OpenCode fixture support"?
10. What pilot audience and feedback channel should be used?

## Claims forbidden before evidence

- "supports Claude Code";
- "compares OpenCode and Claude Code permissions" as a release capability if Claude remains blocked;
- "secure", "compliant", "proves enforcement", or "prevents unsafe agent behavior";
- "production ready", "stable schema", or "recommended for CI enforcement";
- platform support without build/smoke evidence;
- package-manager installation without an actual package;
- adoption/usage/community claims without verifiable evidence.

## Ready for proposal

Yes, with gates. Proceed to proposal only if it frames alpha as a truthful installable prerelease with evidence-backed OpenCode support, explicit unsupported states, release integrity basics, and a separate pilot. The proposal must preserve <=400 authored-line work-unit slices and must not depend on blocked Claude work.
