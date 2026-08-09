# Design: Universal Agent Policy Auditor

## Technical Approach

Build a local Go single binary. This architecture is accepted because the alpha is CPU/file-bound, offline, and needs portable releases, explicit resource bounds, and a mature test toolchain—not runtime plugins or services. Versioned adapters preserve target semantics; the shared model is an internal comparison representation, never an authored policy DSL.

```text
CLI -> Application -> Adapter discovery plan -> Source executor -> Adapter resolver
                                                              -> Comparator -> Renderers
```

## Architecture Decisions

| Decision | Choice and rationale | Rejected alternative |
|---|---|---|
| Deployment | One Go binary with embedded schemas/support metadata; no target executable dependency. | Plugins/processes enlarge trust and version boundaries. |
| Adapter boundary | OpenCode and Claude Code adapters declare discovery semantics, parse raw native data, and resolve native scope, merge, precedence, defaults, matchers, and runtime limits. | Shared permission engine would erase differences. |
| Comparison | Immutable `ResolvedPermission` facts retain capability, resource, effect, scope, matcher, conditions, limitations, trace, provenance, and typed target extensions. Only evidence-backed dimensions are compared; every loss is recorded. | Text comparison or lowest-common-denominator DSL. |
| Support gate | Embedded matrix rows identify target/config version, construct, semantic claims, evidence, and fixture IDs. Build tests require passing fixtures for every supported claim. Unknown versions or permission-bearing constructs block success. | Optimistic parsing risks false assurance. |
| CLI boundary | `internal/app` accepts typed requests and returns reports; `cmd/auditor` maps arguments, streams, and exit categories. Syntax and numeric exits remain UX decisions. | Domain behavior coupled to a CLI framework. |

## Interfaces and Data Flow

`Adapter.DiscoveryPlan(context) DiscoveryPlan` returns declarative candidate locations, scope/selection order, permitted environment keys, and bounded reference policy; it performs no I/O. `internal/source.Executor.Execute(plan, ReadFS, EnvSnapshot) SelectionReport` alone performs canonicalized filesystem/environment reads, deduplication, root enforcement, cycle detection, and selected/skipped/ambiguous reporting. The adapter then receives immutable `SelectedSource` bytes and identities for raw parsing and semantic resolution. It cannot write or invoke processes.

Each permission contains an immutable `ResolutionTrace`: ordered `TraceStep` values (`rule`, `default`, `merge`, `precedence`, `runtime-dependency`) with before/after effect, contributing rule IDs, decision rationale, and redacted `EvidenceRef`; plus unresolved runtime dependencies and final status. Constructors defensively copy slices and expose read-only accessors. Thus defaults, overridden rules, merge decisions, trust/approval/hooks/plugins/sandbox dependencies, and source locations remain inspectable.

`Comparator` emits `equivalent`, `broader`, `narrower`, `target-only`, `unsupported`, `ambiguous`, `lossy`, or `not-comparable` with dimension evidence. Reports carry completeness (`complete`, `incomplete`, `operational-failure`), stable findings, versions, and redacted provenance. Redaction precedes rendering. Versioned JSON uses sorted collections, stable input-derived IDs, and no implicit clock/host fields; human output uses the same report and prioritizes blockers. Debug diagnostics remain explicit, structured, redacted, and local—never telemetry or raw dumps.

## Package / File Structure

| Path | Responsibility |
|---|---|
| `cmd/auditor/`, `internal/app/` | Composition, use cases, transport/exit mapping |
| `internal/source/` | Discovery execution, identity, platform, read-only ports |
| `internal/adapter/{opencode,claudecode}/` | Plans, raw parsing, native resolution |
| `internal/model/`, `internal/compare/` | Reports, traces, findings, comparison |
| `internal/redact/`, `internal/render/` | Safe deterministic output |
| `support/`, `testdata/conformance/` | Matrix, evidence manifests, fixtures, goldens |

## Testing Strategy

Bootstrap `go.mod`, formatting, vetting, and `go test ./...`. Use table-driven parser tests and semantic fixtures for every matrix row, including precedence/default traces, malformed and unknown permission input. With `t.TempDir()`, injected platform/environment data, and read-only fakes, test Linux/macOS/Windows discovery, ambiguity, traversal, symlinks, cycles, case behavior, rejected write opens, and unchanged content/metadata. Secret fixtures cover values, paths, traces, and errors. Schema, byte-stability, and inspected `-update` goldens cover outputs. Fuzz/property tests enforce depth, file, byte, rule, reference, diagnostic, and cancellation limits.

## Trust-Boundary / Threat Matrix

| Boundary | Control | Safe failure |
|---|---|---|
| Malicious config content | Data-only parsers; no evaluation or dynamic loading | Parse finding; incomplete |
| Paths/references/traversal | Canonical roots, bounded depth/count, cycle/symlink checks | Reject source/reference; incomplete |
| Secret-bearing diagnostics | Minimize evidence; redact before report/render | Suppress value; blocking internal failure if safe redaction is impossible |
| Resource exhaustion | Size/count/depth/time limits and cancellation | Limit finding; incomplete, no panic |
| Environment-derived inputs | Allowlisted captured keys; report redacted influence | Unknown/disallowed dependency; unresolved/incomplete |
| Mutation/process execution | Capability-limited `ReadFS`; no write/process interface | Construction/test failure; no target action |

## Migration / Evolution

No migration. Version schemas, adapters, matrices, traces, and extensions independently. New versions/constructs require evidence and fixtures. Baselines, authored policy, compilation, mutation, runtime enforcement, and networking require separate designs and threat reviews.

## Open Questions

- Exact supported versions, evidence authorities, discovery roots, and limits remain release gates; none default to supported.
