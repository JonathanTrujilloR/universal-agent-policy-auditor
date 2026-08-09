# Tasks: Universal Agent Policy Auditor

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 900–1,300 authored lines; goldens out |
| Risk level | High |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 source; PR 2 model; PR 3 adapters; PR 4 output |
| Delivery strategy | auto-forecast |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Work Units

| Unit | Goal | Focused test command | Runtime harness | Rollback boundary |
|---|---|---|---|---|
| 1 | Bootstrap and source executor | `go test ./internal/source/...` | Help command after wiring; otherwise N/A | Bootstrap and `internal/source/` |
| 2 | Model, matrix, evidence fixtures | `go test ./internal/model/... ./support/...` | N/A: library-only | Model/support/testdata |
| 3 | Adapters and resolution | `go test ./internal/adapter/...` | Fixture audit | Adapter directories |
| 4 | Compare, findings, redaction, render, CLI | `go test ./internal/compare/... ./internal/render/... ./internal/app/... ./cmd/auditor/...` | Repeated human/JSON audit | Comparison/output/CLI files |

## Phase 1: Bootstrap and Safety Boundary

- [x] 1.1 Decide whether to initialize Git; if yes, record local-only policy, create `go.mod` and layout, decide formatter/vet/tests/CI/lint, and establish `gofmt -w`, `gofmt -l`, `go vet ./...`, `go test ./...`, `go build ./cmd/auditor`.
- [x] 1.2 RED: table-driven tests for malformed/no-eval, traversal/symlink/cycle, rejected writes/processes, environment allowlist, resource/cancellation bounds, and secret leakage.
- [x] 1.3 Implement `internal/source/` read-only `ReadFS`, `EnvSnapshot`, bounded executor, canonical roots, deduplication, identity, and safe findings; pass 1.2 with `t.TempDir()`.

## Phase 2: Evidence, Model, and Fixtures

- [ ] 2.1 Research official docs/upstream source/executable behavior; create `support/` matrix, evidence manifests, and conformance fixtures; support nothing without evidence plus a fixture.
- [ ] 2.2 Implement `internal/model/` immutable permissions, provenance, findings/completeness, extensions, copied accessors, and immutable `ResolutionTrace`; test stable traces/redaction.
- [ ] 2.3 RED then implement fixtures for every matrix row, defaults, merge/precedence, matchers, conditions, malformed input, unknown permission constructs, and unsupported versions.

## Phase 3: Adapters and Resolution

- [ ] 3.1 RED then implement `internal/adapter/opencode/` plan, parser, resolver, runtime limits, and evidence-backed fixtures.
- [ ] 3.2 RED then implement `internal/adapter/claudecode/` with the same gates; preserve target semantics, not a shared policy engine.
- [ ] 3.3 Test allow/deny/ask/default/conditional/unresolved outcomes, precedence traces, ambiguity, and static-vs-runtime limitations.

## Phase 4: Comparison, Output, and CLI

- [ ] 4.1 RED then implement `internal/compare/` dimension-aware equivalent/broader/narrower/target-only/unsupported/ambiguous/lossy/not-comparable findings.
- [ ] 4.2 Implement `internal/redact/` before `internal/render/`; add versioned JSON, stable ordering/IDs, human blockers-first output, and redaction goldens.
- [ ] 4.3 Wire `internal/app/` and `cmd/auditor/` requests, explicit sources/discovery, streams, help, and exits; test read-only, no-write, byte-stable audits.
- [ ] 4.4 Document alpha limits, modeled-not-enforced semantics, matrix/evidence process, verification commands, and exclusions.

## Phase 5: Verify

- [ ] 5.1 Run `gofmt -l .`, `go vet ./...`, `go test ./...`, and `go build ./cmd/auditor`; inspect all goldens generated through the approved `-update` path, then rerun without `-update`.
- [ ] 5.2 Run fuzz/property tests for depth, files, bytes, rules, references, diagnostics, and cancellation; verify unchanged content/metadata and no process/network execution.
