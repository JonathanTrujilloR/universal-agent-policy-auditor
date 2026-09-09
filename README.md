# Universal Agent Policy Auditor

Universal Agent Policy Auditor is a **pre-alpha**, local-first Go project for auditing AI coding-agent policy semantics. It is not yet an installable or supported release. The current repository contains source foundations, evidence-gating structures, and release planning work, but no supported target version and no usable public alpha binary.

## Current status

| Area | Status for this repository today |
|---|---|
| Release maturity | Pre-alpha source; `v0.1.0-alpha.1` is planned, not available. |
| Supported targets | **0 supported target versions.** No OpenCode or Claude Code version is supported yet. |
| OpenCode | Evidence-gated only; support requires version-pinned evidence, fixtures, and passing conformance before any support claim. |
| Claude Code | Unsupported while issues [#11](https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/issues/11) and [#13](https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/issues/13) remain blocked. |
| Comparison | Planned fail-closed semantics are tracked in [#19](https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/issues/19); no complete OpenCode-versus-Claude support is claimed. |
| Runtime behavior | Static modeled configuration semantics only; no runtime enforcement, compliance, security, or unsafe-behavior prevention guarantee. |
| Distribution | No package-manager install, supported binary download, supported platform, signing, provenance, or release artifact is available yet. |
| Adoption | No adoption, usage, pilot success, or ecosystem acceptance claim is made. |
| License | Apache License 2.0. See [LICENSE](LICENSE). |

## What exists today

- Go module scaffolding and tests for source, support, model, and adapter foundations.
- OpenSpec planning for the alpha release under `openspec/changes/alpha-release-readiness/`.
- Support evidence files and registries that default to no admitted target support until checked-in evidence proves it.
- This public identity/docs baseline from approved issue [#20](https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/issues/20).

## What this project is for

The intended product is a read-only auditor that helps maintainers understand AI coding-agent permission configurations without turning uncertainty into false assurance. The auditor should eventually:

1. Read explicit local configuration inputs.
2. Normalize supported policy constructs into a canonical model.
3. Report unsupported, ambiguous, lossy, or unknown semantics visibly.
4. Compare modeled permissions only when evidence proves the comparison is safe.
5. Emit deterministic human and machine-readable results without exposing secrets.

It is **not** a sandbox, runtime guard, policy enforcement layer, compliance certification tool, hosted service, telemetry system, or configuration writer.

## Architecture at a glance

```text
cmd/auditor (planned CLI transport)
  -> internal/app (planned request/result boundary)
  -> internal/source (read-only source selection)
  -> internal/adapter/* (target-specific parsing and resolution)
  -> support (evidence matrix and fixtures)
  -> internal/model (canonical semantic facts)
  -> internal/compare (only through #19 when available)
  -> internal/redact + internal/render (planned safe output)
```

The current source is intentionally conservative: target support must pass through evidence gates, and unsupported or incomplete input must not become a successful audit.

## Verification for contributors

Run the current repository checks before proposing changes:

```bash
git diff --check
gofmt -l .
go vet ./...
go test ./... -count=1
go build ./...
```

For documentation-only changes, also perform structural readback: confirm the README and related release-facing files repeat Apache-2.0 consistently, show zero supported target versions until evidence exists, and avoid unsupported install, platform, security, compliance, enforcement, Claude, or adoption claims.

## Development roadmap

| Work area | Status |
|---|---|
| Public identity and Apache-2.0 baseline | This Work Unit. |
| CLI/app shell | Planned separately; no CLI command is documented as available today. |
| OpenCode evidence gate | Planned; exact version support remains unresolved until evidence is checked in. |
| Redaction and renderers | Planned; output must be safe before release claims. |
| Comparison integration | Planned only through approved fail-closed work in [#19](https://github.com/JonathanTrujilloR/universal-agent-policy-auditor/issues/19). |
| Release build and checksums | Planned; required before any Linux amd64 artifact support claim. |
| Post-release pilot | Future feedback collection only; not adoption evidence. |

## Contributing

Contributions should preserve the audit-only, fail-closed posture:

- Keep each change tied to one approved issue and one reviewable Work Unit.
- Add tests or structural evidence with the behavior or documentation being changed.
- Do not add target support claims without exact version evidence, fixtures, and passing conformance.
- Do not add write, format, delete, remediate, runtime enforcement, network, or telemetry behavior to the alpha without a separate approved proposal.

Maintainer-controlled repository metadata may be updated later if approved. Suggested description: `Pre-alpha local auditor for AI coding-agent policy semantics.` Suggested topics, if maintainers choose to set them: `ai-agents`, `policy-audit`, `opencode`, `claude-code`, `golang`, `security-tools`.

## License

This project is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).
