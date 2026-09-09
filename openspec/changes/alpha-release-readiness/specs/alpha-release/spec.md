# Alpha Release Specification

## Purpose

The alpha release SHALL provide a truthful, legally reusable, installable `v0.1.0-alpha.1` boundary for the local policy auditor. It SHALL publish only evidence-backed OpenCode-on-Linux-amd64 capability, keep unsupported targets explicit, and separate release readiness from any later adoption or pilot claims.

## Requirements

### Requirement: Apache-2.0 Release Identity

The release MUST project Apache-2.0 identity consistently across the repository license, public project positioning, release metadata, and release artifacts. A release MUST NOT be considered ready when any public release surface omits the license or conflicts with Apache-2.0.

#### Scenario: License identity is consistent

- GIVEN the alpha release is being prepared
- WHEN repository identity, README positioning, release notes, and release artifacts are reviewed
- THEN every release-facing surface MUST identify Apache-2.0 consistently
- AND no conflicting license claim MAY appear.

#### Scenario: Missing license blocks release

- GIVEN a release artifact or public release surface lacks Apache-2.0 identity
- WHEN release readiness is evaluated
- THEN the release MUST be blocked until the missing identity is corrected.

### Requirement: Evidence-Gated OpenCode Support Claim

The release MUST NOT claim OpenCode support for any exact version until version-pinned evidence, stable documentation/source or executable evidence, registered fixtures, and passing conformance evidence are checked in. The exact OpenCode version MAY remain unresolved before the evidence Work Unit, but any unresolved version MUST be reported as not yet supported.

#### Scenario: Version-pinned evidence permits support claim

- GIVEN an exact OpenCode version has checked-in stable evidence, registered fixtures, and passing conformance evidence
- WHEN release support claims are generated
- THEN the release MAY claim support for that exact OpenCode version only
- AND the claim MUST cite or reference the evidence boundary.

#### Scenario: Missing evidence blocks support claim

- GIVEN OpenCode behavior is described without version-pinned evidence and passing fixtures
- WHEN release readiness is evaluated
- THEN the release MUST NOT claim OpenCode support
- AND the support status MUST remain experimental, unsupported, or unresolved rather than successful.

### Requirement: Claude Code Unsupported Status

The alpha release MUST state that Claude Code is unsupported while issues `#11` and `#13` remain blocked. It MUST NOT absorb the blocked Claude scope, depend on those issues for alpha readiness, or imply cross-client support through wording, examples, release notes, or output.

#### Scenario: Claude remains unsupported

- GIVEN issues `#11` and `#13` remain blocked
- WHEN alpha release documentation or CLI output describes target support
- THEN Claude Code MUST be shown as unsupported for this alpha
- AND no Claude Code support or success claim MAY appear.

#### Scenario: Unsupported target exits visibly

- GIVEN a user requests an alpha audit path that requires Claude Code support
- WHEN the release behavior is evaluated
- THEN the result MUST use a visible unsupported or incomplete exit
- AND it MUST NOT report unqualified success.

### Requirement: Linux amd64 Artifact and Checksum Boundary

The release MUST publish a GitHub Release for `v0.1.0-alpha.1` with a Linux amd64 binary, a source tag, and `checksums.txt` before Linux amd64 is listed as a supported release artifact. Additional platforms, package-manager distribution, signing, and provenance attestations MUST remain out of scope unless independently evidenced by a later change.

#### Scenario: Linux amd64 artifact is release-ready

- GIVEN the alpha release lists Linux amd64 as the supported binary platform
- WHEN release artifacts are inspected
- THEN a Linux amd64 binary, matching source tag, and `checksums.txt` MUST be present
- AND release verification MUST include Linux amd64 build and smoke evidence.

#### Scenario: Unevidenced distribution claim is forbidden

- GIVEN no package-manager package, signing material, provenance attestation, or additional platform smoke evidence exists
- WHEN release notes or install documentation are reviewed
- THEN those surfaces MUST NOT claim package-manager installation, signing, provenance, or additional platform support.

### Requirement: Read-Only Explicit Local Input

The alpha CLI MUST operate only on explicit local input selected by the user and MUST NOT mutate target files, write configuration, format configuration, delete configuration, execute configuration content, or follow unbounded references. The CLI command syntax MAY be finalized later, but the release behavior MUST preserve a local read-only boundary.

#### Scenario: Explicit input is audited without mutation

- GIVEN a user provides an explicit local input path
- WHEN the alpha CLI evaluates that input
- THEN only the selected input MUST be analyzed
- AND target files MUST remain unchanged.

#### Scenario: Mutation behavior is rejected

- GIVEN a requested behavior would write, format, delete, remediate, or execute target configuration
- WHEN release readiness is evaluated
- THEN that behavior MUST be excluded from the alpha release
- AND no release documentation MAY imply it is available.

### Requirement: Deterministic CLI Identity and Help Behavior

The alpha CLI MUST provide deterministic `auditor version` and help behavior for identical build metadata. Development builds MUST identify as `dev` and MUST NOT claim release identity. Unknown or invalid commands MUST be visible invalid requests, and transport behavior MUST perform no target I/O before a valid audit request exists.

#### Scenario: Version and help are deterministic

- GIVEN identical build metadata for the alpha CLI
- WHEN `auditor version` and help output are requested repeatedly
- THEN the reported version identity and help text MUST remain deterministic
- AND development builds MUST identify as `dev` rather than a release version.

#### Scenario: Development build cannot claim release identity

- GIVEN an alpha CLI build without release metadata
- WHEN version identity is displayed
- THEN the CLI MUST identify the build as `dev`
- AND it MUST NOT claim `v0.1.0-alpha.1` or any other release identity.

#### Scenario: Invalid command is visible without target access

- GIVEN a user supplies an unknown command or invalid command shape
- WHEN the alpha CLI handles the request
- THEN the request MUST be classified as invalid_request
- AND no target I/O MAY occur before a valid audit request exists.

### Requirement: Deterministic Human and JSON Output

The release MUST provide deterministic concise human output and deterministic versioned JSON output for identical supported inputs, declared environment, and tool version. JSON output MUST include schema version, target/version/support status, completeness, permissions or differences, findings, provenance digests, and limitations.

#### Scenario: Repeated output is deterministic

- GIVEN identical supported inputs, declared environment, and tool version
- WHEN human and JSON output are produced repeatedly
- THEN the reported sources, support status, completeness, findings, provenance digests, and limitations MUST remain stable
- AND JSON field ordering and values MUST be deterministic.

#### Scenario: JSON includes release-blocking fields

- GIVEN JSON output is requested for an alpha audit
- WHEN the output is produced
- THEN it MUST include schema version, target/version/support status, completeness, permissions or differences, findings, provenance digests, and limitations.

### Requirement: Redaction and Privacy Safety

Human and JSON output MUST exclude raw secrets, credentials, private configuration values, and unnecessary local context by default. Provenance MUST remain useful through redacted references or digests rather than leaking sensitive input content.

#### Scenario: Sensitive values are redacted

- GIVEN selected input contains secrets, credentials, private paths, or sensitive configuration values
- WHEN human or JSON output is produced
- THEN raw sensitive values MUST NOT appear
- AND the output MUST retain only safe redacted provenance or digests needed to understand the finding.

#### Scenario: Pilot guidance protects privacy

- GIVEN post-release pilot materials request feedback
- WHEN they describe what users may share
- THEN they MUST warn users not to share secrets, credentials, private paths, raw sensitive configs, or unredacted local context.

### Requirement: Completeness and Unsupported Exits

The release MUST classify malformed, unreadable, unsupported, unknown, ambiguous, lossy, or incomplete input as visible findings with safe non-success status. Unqualified success MUST require supported inputs, complete evidence, deterministic processing, and no blocking findings.

#### Scenario: Incomplete input blocks unqualified success

- GIVEN selected input is malformed, unreadable, ambiguous, unsupported, unknown, or lossy
- WHEN the alpha CLI evaluates it
- THEN the result MUST identify the condition as a visible finding
- AND the result MUST NOT be an unqualified success.

#### Scenario: Complete supported input may succeed

- GIVEN selected input is supported by version-pinned evidence and no blocking findings are present
- WHEN the alpha CLI evaluates it deterministically
- THEN the result MAY report success
- AND the result MUST still include limitations for modeled static semantics.

### Requirement: Stable Exit Categories and Codes

The alpha CLI MUST expose stable exit categories and codes: `0` for `complete_no_findings`, `1` for `complete_with_findings`, `2` for `unsupported_or_incomplete`, `3` for `invalid_request`, and `4` for `operational_failure`. Unsupported or incomplete results MUST never map to exit code `0`. Redaction or rendering failures MUST map to exit code `4` and MUST emit no unsafe report.

#### Scenario: Complete audit without findings exits zero

- GIVEN a supported audit completes deterministically with complete evidence and no findings
- WHEN the alpha CLI exits
- THEN it MUST use category `complete_no_findings`
- AND it MUST return exit code `0`.

#### Scenario: Complete audit with findings exits one

- GIVEN a supported audit completes deterministically with one or more findings
- WHEN the alpha CLI exits
- THEN it MUST use category `complete_with_findings`
- AND it MUST return exit code `1`.

#### Scenario: Unsupported or incomplete audit exits two

- GIVEN selected input is unsupported, incomplete, ambiguous, unknown, lossy, or lacks required evidence
- WHEN the alpha CLI exits
- THEN it MUST use category `unsupported_or_incomplete`
- AND it MUST return exit code `2` rather than `0`.

#### Scenario: Invalid request exits three

- GIVEN a user supplies an invalid command, missing required audit request, or invalid request shape
- WHEN the alpha CLI exits
- THEN it MUST use category `invalid_request`
- AND it MUST return exit code `3`.

#### Scenario: Operational failure exits four without unsafe report

- GIVEN an operational failure occurs, including redaction or rendering failure
- WHEN the alpha CLI exits
- THEN it MUST use category `operational_failure`
- AND it MUST return exit code `4`
- AND it MUST emit no unsafe report.

### Requirement: Comparison Claim Boundary

The release MUST depend on approved issue `#19` for deterministic fail-closed comparison semantics and MUST NOT duplicate or redefine `#19` implementation requirements. Alpha documentation and output MUST NOT imply complete OpenCode-versus-Claude comparison while Claude Code is unsupported.

#### Scenario: Comparison delegates to approved semantics

- GIVEN a release surface describes comparison behavior
- WHEN the description is reviewed
- THEN it MUST reference the approved fail-closed comparison semantics from `#19`
- AND it MUST NOT add separate comparison implementation requirements in this release spec.

#### Scenario: Unsupported comparison is not marketed as cross-client support

- GIVEN Claude Code remains unsupported
- WHEN release notes, README content, examples, or output mention comparison
- THEN they MUST avoid unqualified OpenCode-versus-Claude support claims
- AND unsupported or not-comparable states MUST remain visible.

### Requirement: CI and Release Verification Evidence

The release MUST have recorded CI or release-readiness evidence for formatting, vetting, tests, build, checksum generation, and Linux amd64 smoke verification before publication. A missing required check MUST block release readiness.

#### Scenario: Required checks pass before release

- GIVEN the alpha release candidate is being evaluated
- WHEN release-readiness evidence is reviewed
- THEN formatting, vetting, tests, build, checksum generation, and Linux amd64 smoke verification MUST have recorded passing evidence.

#### Scenario: Missing check blocks release

- GIVEN any required CI or release verification evidence is absent or failing
- WHEN release readiness is evaluated
- THEN the alpha release MUST be blocked until passing evidence exists.

### Requirement: Truthful Release Notes and Public Claims

Release notes, README content, metadata, examples, and support matrices MUST state only release capabilities that are supported by checked-in evidence. They MUST NOT claim Claude Code support, unqualified OpenCode-versus-Claude comparison, production readiness, security or compliance guarantees, runtime enforcement, unsafe-behavior prevention, package-manager installation, unsupported platforms, signing/provenance, or adoption/pilot success without corresponding evidence.

#### Scenario: Forbidden claim blocks release

- GIVEN a public release surface contains a forbidden success, support, security, production, distribution, platform, or adoption claim without evidence
- WHEN release readiness is evaluated
- THEN the release MUST be blocked until the claim is removed or evidence is added by an appropriate change.

#### Scenario: Static-analysis limitation is visible

- GIVEN public release documentation or CLI output describes audit results
- WHEN a user reads the release surface
- THEN it MUST state that results model static configuration semantics
- AND it MUST NOT imply native runtime enforcement or compliance certification.

### Requirement: Post-Release Pilot Separation

The external pilot MUST be treated as a post-release evidence-collection activity, not as a release-readiness prerequisite or adoption proof. Pilot materials MAY collect structured feedback after the alpha exists, but they MUST NOT fabricate usage, adoption, ecosystem acceptance, security effectiveness, or production-readiness claims.

#### Scenario: Pilot starts after release

- GIVEN pilot materials are prepared
- WHEN they describe participation
- THEN they MUST position the pilot as post-release feedback collection
- AND they MUST NOT represent the pilot as evidence that the alpha has already been adopted.

#### Scenario: Pilot feedback remains evidence-bound

- GIVEN pilot feedback is received after release
- WHEN public claims are made from that feedback
- THEN claims MUST be limited to verified evidence
- AND unsupported adoption or security-effectiveness claims MUST remain forbidden.
