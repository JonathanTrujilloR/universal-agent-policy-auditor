# Apply Progress: Claude Code Conformance Harness

## Batch: Code PR 1 - Domain types and classifier only

Status consumed: OpenSpec authoritative; parent supplied bounded correction token `sha256:a601620b021206c66c5596b0ca455ab7d801075cfbf9f563306e424f95d13eac` for failed evidence revision `sha256:453a7cb4d43436b8826dd6efd4bbfd198af28f2a30150422b8abd190167e1b7e`; parent owns settlement. Strict TDD active (`go test ./...`). Allowed edit surfaces were the three Code PR 1 Go files plus this change's `tasks.md`/`apply-progress.md`; no other paths edited. Delivery path remains `auto-chain`, `stacked-to-main`, Code PR 1 only, no size exception.

Completed implementation tasks: Code PR 1 RED, GREEN, TRIANGULATE, REFACTOR, Verify, and PR-boundary preparation are still checked in `tasks.md`. Correction keeps those checkboxes truthful by adding explicit parser-proof and encoding-proof classifier gates before `evidence_candidate`.

Files changed: `internal/conformance/claudecode/types.go`; `internal/conformance/claudecode/classifier.go`; `internal/conformance/claudecode/classifier_test.go`; `openspec/changes/claudecode-conformance-harness/tasks.md`; `openspec/changes/claudecode-conformance-harness/apply-progress.md`. `tasks.md` was re-read and Code PR 1 checkboxes remain visibly checked.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Code PR 1 original classifier domain | `internal/conformance/claudecode/classifier_test.go` | Unit | N/A (new package files) | Earlier `go test ./internal/conformance/claudecode` failed on undefined domain/classifier symbols | Earlier package test passed after minimal `types.go`/`classifier.go` | Earlier package test passed after decision-table cases | Earlier package/root tests passed; no runtime seams |
| Gatekeeper correction: parser/encoding proof gates | `internal/conformance/claudecode/classifier_test.go` | Unit | `go test ./internal/conformance/claudecode` -> ok (cached) before correction | `go test ./internal/conformance/claudecode` -> build failed: missing `ParserProof`, `EncodingProof`, parser/encoding fields, and reason codes | `go test ./internal/conformance/claudecode` -> ok 0.002s after minimal proof types/gates | Table covers absent and false parser proof, malformed, bounds, unsupported, absent/unstable/unbounded/unsafe encoding | `gofmt -w` then focused/root tests stayed green |

## Verification evidence

- Corrective focused/root runs passed; after the parent applied final `gofmt`, an independent verifier reran `go test ./internal/conformance/claudecode` (ok 0.002s) and `go test ./...` (all listed packages passed).
- Runtime harness: N/A; no Claude/model/network/process harness was invoked.

## Design alignment and deviations

- Added closed `ParserProof` and `EncodingProof` fields to `ClassificationInput`; `evidence_candidate` now requires valid bounded structured-event proof and schema-versioned, byte-stable, bounded, safe encoding proof.
- Added closed fail-safe reason codes only for schema/stability proof absence: `encoding_schema_unproven`, `encoding_stability_unproven`; bounded and unsafe encoding reuse existing closed reasons.
- Deterministic primary reason precedence preserved: support claim/privacy/parser/fact/sentinel/encoding/provenance/runtime.
- No parser, encoder, runtime, workspace, process, support metadata, ledger, issue, or blocked-main-change behavior was implemented or edited.

## Workload and PR boundary

Code PR 1 remains the single stacked-to-main slice after Planning PR C. Final authored count after parent normalization and progress compaction: 387 additions+deletions across the five allowed paths; independent semantic/tests verification passed, while final Git path/count readback remains pending. Rollback boundary: remove the three conformance package files and this Code PR 1 progress metadata.

## Remaining tasks

Code PR 1 has no unchecked implementation rows. Remaining unchecked rows begin at Code PR 2 and remain unchanged in `openspec/changes/claudecode-conformance-harness/tasks.md`; parent-owned archive actions remain deferred to parent lifecycle.
