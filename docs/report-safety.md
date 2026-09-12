# Safe report boundary

`internal/redact.NewReport` converts application results into a closed, immutable
DTO. Renderers must accept only a `Report` with `Valid() == true`, never raw
application/model values. The zero report is invalid; conversion failures return
only `redact: unsafe report`, without input details.

## Accepted and withheld fields

| Input | Report policy |
|---|---|
| Category, target, support, completeness, limitations | Closed values and validated state combinations only. |
| Tool version | Only `dev` or planned alpha `v0.1.0-alpha.1`; this does not claim a published release. |
| Requested version | Only exact OpenCode `1.18.27`; empty means missing/unresolved and stays empty without a finding. |
| Other versions | Withheld; one fixed `redact-version-withheld` finding even when both versions are withheld. Incompatible complete states fail closed. |
| Permissions | Only modeled scalar bash/edit/read, allow/deny/ask, OpenCode scope and wildcard matcher; no arbitrary matcher/config text. |
| Findings | Known codes only; messages discarded. Unknown codes fail closed. |
| Evidence and traces | Only validated digest values survive. Source/location, rules, rationale and raw evidence are not report fields. |

Source digests identify **canonical-path identity only**, not file contents.
Permission provenance digests hash **the adapter evidence value only**, not a
whole configuration, trace, or source identity. Digest syntax validation does not
verify origin. Unsalted hashes permit dictionary guessing and correlation,
especially for low-entropy evidence values; sharing reports still needs judgment.

All collection accessors return defensive copies, including nested permission
provenance. Backing fields are private; public nominal string value types do not authorize
constructing a valid report. Closed values and digests are intentionally retained,
not anonymized. This boundary makes no anonymity, content-integrity, enforcement,
credential-detection, security, or compliance guarantee.

## Verification and integration

Privacy tests exercise discarded metadata, hostile strings, generic errors,
repeated access and nested aliasing. Run `go test ./internal/redact -count=25` and
`go test ./internal/redact -run '^$' -fuzz '^FuzzReportPrivacy$' -fuzztime=3s -parallel=1`.
Fuzzing checks bounded adversarial projections, not exhaustive privacy proof.

Later integration must map a redaction error to `operational_failure` / exit `4`
and must not print the rejected result. `internal/render/json.Marshal` now accepts
only valid reports and returns complete `auditor-report/v1alpha1` JSON bytes with
one trailing newline; invalid reports return no bytes and `json: invalid report`.
The alpha schema is not stable. Field order is `schema`, `tool`, `result`,
`exit_code`, `target`, `completeness`, `source_digests`, `permissions`, `findings`,
then `limitations`. Tool fields are `name`, `version`; target fields are `name`,
`requested_version`, `support_status`; permission fields are `capability`, `effect`,
`scope`, `matcher`, `provenance_digests`; each finding has only `code`. Every field
is present. Empty strings remain `""`, and empty collections are `[]`. Redacted
array order is preserved. Standard `encoding/json` HTML escaping is accepted.

A sanitized unsupported result therefore retains the closed state without local data:

```json
{"schema":"auditor-report/v1alpha1","tool":{"name":"universal-agent-policy-auditor","version":"dev"},"result":"unsupported_or_incomplete","exit_code":2,"target":{"name":"opencode","requested_version":"","support_status":"unsupported"},"completeness":"incomplete","source_digests":[],"permissions":[],"findings":[{"code":"unsupported-version"}],"limitations":["modeled-static-configuration"]}
```

No raw application/model values or injectable output DTOs are accepted.
CLI activation and human rendering remain WU5: the source-built CLI is still
category-only and pre-release. JSON adds no comparison, enforcement, security,
compliance, anonymity, or other deferred guarantees. Renderer golden tests pin
required fields, safe digests, empty arrays, exits and 100 byte-identical repeats:
`go test ./internal/render/json -count=25`.
