# Proposal: Alpha Release Readiness

## Intent

Prepare a truthful, installable `v0.1.0-alpha.1` release for the local policy auditor without overstating support, security, adoption, or cross-client comparison maturity.

The alpha should make the project understandable, legally reusable, downloadable, and runnable for one narrow supported surface: OpenCode on Linux amd64, gated by checked-in evidence. Claude Code remains explicitly unsupported while issues `#11` and `#13` are blocked. The later external pilot is a separate evidence-collection activity after the release, not a source of fabricated adoption claims.

## Business and user problem

The project is public, but it is not yet an open-source release. Without an Apache-2.0 `LICENSE`, release artifacts, support evidence, install instructions, and truthful limitations, outside users cannot responsibly reuse, evaluate, or pilot the auditor.

The core user risk is false assurance: users may believe the tool supports specific agent clients, compares OpenCode and Claude Code, proves runtime enforcement, or provides production security/compliance guarantees before the repository has evidence for those claims. The alpha must reduce that confusion by publishing a small, verifiable capability slice and refusing unsupported semantics.

## Target users and situations

- Maintainers preparing the first public prerelease and release notes.
- Early technical evaluators who already use OpenCode locally and can run a Linux amd64 CLI against non-sensitive or redacted configuration inputs.
- Future pilot participants who need clear privacy guidance, expected outputs, and feedback channels after the release exists.

The primary workflow is local evaluation: a user discovers the repository, installs or downloads a binary, runs a read-only command against explicit local input, receives deterministic human and JSON output, and understands the support status and limitations.

## Alpha outcome

A successful `v0.1.0-alpha.1` alpha provides:

- Apache-2.0 licensing projected consistently to repository-root `LICENSE`, README, release artifacts, and repository metadata.
- A documented Linux amd64 GitHub Release binary plus `checksums.txt`, with signing/provenance deferred.
- A CLI entry point under `cmd/auditor` with deterministic help/version behavior, explicit local input, stable exit categories, text and JSON output, and no target mutation.
- Evidence-backed OpenCode support only after an exact version, stable documentation/source or executable evidence, registered fixtures, and passing conformance evidence are checked in.
- Clear unsupported status for Claude Code in this alpha.
- Comparison behavior only through the approved deterministic fail-closed comparison work in issue `#19`; the alpha must not imply Claude support or complete cross-client coverage.
- Release documentation that explains modeled static configuration semantics, redaction, completeness, and unsupported states.
- A separate post-release pilot pack for structured feedback without adoption, usage, or ecosystem acceptance claims.

## Scope

### In scope

- Public project identity for a real open-source alpha: Apache-2.0 license, README positioning, metadata alignment, and release notes.
- Linux amd64 prerelease distribution through GitHub Releases with checksum publication.
- Local-only CLI surface under `cmd/auditor`; exact command syntax is deferred to design.
- Deterministic output contract at proposal level: concise human output and versioned JSON containing target/version/support status, completeness, permissions or differences, findings, provenance digests, limitations, and no raw secrets or configuration values.
- OpenCode-only support gate for the first alpha, with no version claim until evidence is registered.
- Integration dependency on approved issue `#19` for deterministic fail-closed comparison semantics.
- CI/release readiness evidence for formatting, vetting, tests, build, and documented release-platform smoke evidence.
- Post-release pilot materials that collect structured feedback safely.

### Out of scope and non-goals

- Claude Code support while `#11` and `#13` remain blocked.
- Depending on, bypassing, or absorbing blocked issues `#11` or `#13` into this release readiness change.
- Claiming support for any exact OpenCode version before version-pinned evidence and fixtures prove it.
- Runtime enforcement, sandboxing, policy remediation, config writing, config formatting, deletion, hosted service, telemetry, compliance certification, or CI enforcement claims.
- Package-manager distribution, signing, provenance attestations, and additional platforms unless separately evidenced.
- Adoption, usage, community, or ecosystem acceptance claims before verified evidence exists.
- Research lane work; the research lane remains unselected for this proposal.

## Affected areas

| Area | Expected impact |
|---|---|
| Repository identity | Add Apache-2.0 licensing, README positioning, metadata, and release-facing documentation. |
| CLI release surface | Add or formalize `cmd/auditor`, help/version behavior, local input boundaries, exit categories, and no-mutation guarantees. |
| Support evidence | Register only evidence-backed OpenCode support; keep unsupported targets visible and fail-closed. |
| Output and redaction | Define deterministic human/JSON release output with provenance, limitations, completeness, and secret-safe defaults. |
| Comparison | Depend on approved issue `#19` for fail-closed comparison behavior without claiming Claude support. |
| CI and release packaging | Produce Linux amd64 binary, source tag, checksums, and recorded verification evidence. |
| Pilot materials | Provide post-release feedback template or discussion path with privacy guidance and no adoption claims. |

## Dependencies

- Maintainer-confirmed license decision: Apache-2.0.
- Approved issue `#19` for deterministic fail-closed comparison core.
- OpenCode version-pinned evidence, official documentation/source or executable evidence, fixtures, and passing conformance results before any exact support/version claim.
- Linux amd64 build and smoke evidence before listing Linux amd64 as a supported release artifact.
- GitHub Releases availability for binary and checksum publication.

This proposal must not depend on blocked issues `#11` or `#13`.

## Truthful claims policy

The release may claim only what the repository can prove at release time.

Forbidden until evidenced:

- `supports Claude Code`.
- unqualified `compares OpenCode and Claude Code permissions` while Claude is unsupported.
- exact OpenCode version support without registered version-pinned evidence and passing fixtures.
- `secure`, `compliant`, `prevents unsafe behavior`, `production ready`, or `recommended for enforcement`.
- platform support without build and smoke evidence.
- package-manager installation without an actual package.
- adoption or pilot success without verifiable post-release evidence.

The README and release notes must make clear that the tool models static configuration semantics and does not prove native runtime enforcement.

## Technical release versus adoption pilot

The alpha release and the external pilot are separate phases.

- The alpha release creates a legally reusable, installable, evidence-gated prerelease artifact.
- The pilot begins only after the release exists and gathers structured feedback from early OpenCode users.
- Pilot materials may ask for unsupported constructs, confusing findings, false-assurance concerns, and output usability feedback.
- Pilot materials must warn users not to share secrets, private paths, raw sensitive configs, credentials, or unredacted local context.
- Pilot feedback may inform later roadmap decisions, but it must not be represented as adoption, security effectiveness, or ecosystem acceptance unless independently verified.

## Work-unit and review boundary policy

Implementation should be planned as coherent Work Units, with one approved issue per coherent Work Unit and each PR slice kept within the `<=400` authored changed-line review budget. Tests and docs belong with the behavior they verify. Any unit that approaches the budget should split into a chained PR slice instead of becoming an oversized review.

Proposal-level Work Unit candidates:

| Candidate Work Unit | Purpose | Review budget risk |
|---|---|---|
| Public project identity | Apache-2.0 license projection, README positioning, metadata, and release claim boundaries. | Low |
| CLI shell | `cmd/auditor`, help/version, explicit input, exit categories, and no-mutation boundary. | Medium |
| OpenCode evidence gate | Version/evidence matrix, fixtures, and support-status behavior for OpenCode only. | Medium |
| Output, redaction, and rendering | Versioned JSON, deterministic human output, provenance digests, limitations, and secret-safe defaults. | High; likely split if near 400 authored lines. |
| Comparison integration | Integrate approved issue `#19` fail-closed comparison semantics without Claude support claims. | Medium |
| CI and release build | Linux amd64 build, checksums, source tag, and recorded verification evidence. | Medium |
| Release docs | Install, usage, support matrix, limitations, and release notes grounded in checked-in evidence. | Low |
| Pilot pack | Post-release feedback template or discussion path with privacy guidance. | Low |

This is a forecast only; implementation tasks are deferred to the SDD tasks phase.

## Risks and mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Public repository is mistaken for an open-source release before licensing is added. | Legal ambiguity and adoption friction. | Make Apache-2.0 `LICENSE` and matching metadata release-blocking. |
| Users infer Claude or cross-client support. | False assurance and roadmap confusion. | State Claude unsupported; depend only on `#19` for fail-closed comparison; do not depend on `#11` or `#13`. |
| OpenCode support is claimed without version-pinned evidence. | Incorrect audit results and loss of trust. | Require exact version, evidence, fixtures, and passing conformance before support/version claims. |
| Output leaks sensitive configuration values. | User privacy and security harm. | Redact secrets and raw config values by default in human and JSON output. |
| Static analysis is mistaken for runtime enforcement. | Misuse in security/compliance decisions. | Label results as modeled static semantics and surface runtime limitations in every output. |
| Release packaging exceeds review capacity. | Burned-out reviewers and low-quality review. | Split by coherent Work Unit, one approved issue per Work Unit, and keep each PR slice within 400 authored lines. |
| Pilot feedback is overstated as adoption. | Misleading public claims. | Keep pilot separate and report only verified feedback evidence after release. |

## Rollback

Rollback should be possible per Work Unit:

- Remove or revert release identity/docs without touching auditor semantics.
- Remove the `cmd/auditor` alpha entrypoint and release packaging without mutating user configuration files.
- Remove unsupported or unproven support-matrix entries and fixtures as a unit.
- Revert output/redaction/rendering changes without changing target configs.
- Revert comparison wiring independently from approved `#19` core if integration proves unsuitable.
- Withdraw a GitHub prerelease artifact and publish corrected release notes if a claim or artifact is wrong.

Because the alpha must remain local, read-only, and non-mutating, rollback must not require restoring user configuration state.

## Success criteria

- Apache-2.0 is present and consistently reflected in repository identity, README, release artifacts, and metadata.
- `v0.1.0-alpha.1` has a GitHub Release binary for Linux amd64, source tag, and `checksums.txt`.
- CI/release evidence covers formatting, vetting, tests, build, and Linux amd64 release smoke verification.
- The CLI under `cmd/auditor` can run against explicit local input and returns deterministic human and JSON output without mutating target files.
- The JSON output is versioned and includes target/version/support status, completeness, permissions or differences, findings, provenance digests, and limitations while excluding raw secrets and configuration values.
- No exact OpenCode support/version claim appears unless version-pinned evidence and passing fixtures are checked in.
- Claude Code is explicitly unsupported in the alpha while `#11` and `#13` remain blocked.
- Comparison claims are limited to the approved `#19` fail-closed semantics and do not imply two-client support.
- Release docs separate alpha capability from the later external pilot.
- The pilot pack collects structured feedback after release without fabricating adoption or security effectiveness claims.
- Each implementation slice is organized around one coherent Work Unit, one approved issue, and the `<=400` authored changed-line review budget.

## Review-workload forecast

Chained PRs are likely to be appropriate if output/redaction/rendering or release packaging expands beyond the 400 authored-line budget. The proposal itself forecasts medium overall review risk and high localized risk for output/redaction/rendering. The tasks phase should preserve coherent Work Units, keep tests and docs with each unit, and recommend chained PRs for any slice that cannot remain focused and under budget.
