# Pre-Proposal Gate: alpha-release-readiness

## Status

Ready. The maintainer confirmed `apache-2.0`; proposal may start.

## Confirmed inputs

- Execution mode: automatic.
- Artifact store: OpenSpec.
- Delivery strategy: ask-on-risk.
- Review budget: 400 authored changed lines per slice.
- Research lane: unselected.
- Release goal: truthful, installable `v0.1.0-alpha.1`, followed by a separate external pilot.
- First alpha target: OpenCode only. Claude Code remains explicitly unsupported while issues #11 and #13 are blocked.
- Support claim gate: no OpenCode version is supported until an exact version, stable evidence, registered fixture, and passing conformance evidence are checked in.
- Platform scope: Linux amd64 first. Additional platforms require independent build and smoke evidence.
- Distribution: GitHub Release binary plus `checksums.txt`; source tag included. Package managers are deferred.
- Signing/provenance: deferred beyond the first alpha; checksums are mandatory.
- CLI boundary: `cmd/auditor` with explicit local input, deterministic text/JSON output, stable exit categories, and no target mutation. Exact syntax is a design decision.
- JSON minimum: versioned schema, target/version/support status, completeness, permissions/differences, findings, provenance digests, and explicit limitations; no raw secrets or configuration values.
- Pilot: GitHub issue template or Discussions-based structured feedback after release; no fabricated usage/adoption claims.
- Work Unit boundaries: public project identity, CLI shell, OpenCode evidence gate, output/redaction/rendering, comparison integration (#19), CI/release build, release docs, and pilot pack.

## Confirmed license decision

- License: `apache-2.0`.
- Rationale: permissive reuse with an explicit patent grant supports ecosystem and commercial adoption.
- Required projection: repository-root `LICENSE`, README license statement, release artifacts, and repository metadata MUST agree.

## Proposal readiness

Ready. Research is unselected, product decisions are confirmed, repository evidence references are valid, and the OpenSpec store is writable in the isolated worktree.
