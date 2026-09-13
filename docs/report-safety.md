# Safe report boundary

`internal/redact.NewReport` converts application results into a closed, immutable
DTO. Renderers must accept only a `Report` with `Valid() == true`, never raw
application/model values. The zero report is invalid; conversion failures return
only `redact: unsafe report`, without input details.

## Comparison report boundary

`NewComparisonReport` is a separate opaque safety boundary; comparison CLI/output is not activated yet.
It retains no raw path, config, provenance, message, or runtime token. Status `unavailable` is a redaction state, not a core outcome.
App-produced side digests identify canonical paths; redaction validates syntax, not origin. Unsalted values permit low-entropy dictionary guessing and cross-report correlation.
Redaction provides neither anonymity, credential detection, semantic equivalence, integrity, runtime enforcement, nor compliance.
Collections are defensive copies: sources sort by side, reasons by side/code, differences lexically by capability/scope/matcher/side/outcome/reason, findings by code.
Unsafe versions are empty with one withholding marker in permitted incomplete states; complete/equivalent rejects them. Missing requested versions are not withholding.
Comparison privacy tests cover closed enums, raw caps before deduplication, ignored-channel canaries, 100 permutations, copies, and bounded accessor-projection fuzzing.

## Comparison JSON (not CLI activated)

`internal/render/json.MarshalComparison` consumes only an opaque valid
`redact.ComparisonReport`. It returns complete `auditor-comparison-report/v1alpha1`
bytes with exactly one trailing newline, or nil bytes and the fixed error
`json: invalid comparison report` for an invalid/zero report or marshal failure.
No partial buffer is returned. The alpha schema is intentionally unstable.

Every field is present, in this order: `schema`, `tool`, `result`, `exit_code`,
`target`, `completeness`, `comparison_status`, `comparison_sources`, `reasons`,
`differences`, `findings`, `limitations`. Nested field order is:

- `tool`: `name`, `version`; name is `universal-agent-policy-auditor`.
- `target`: `name`, `requested_version`, `support_status`.
- Each comparison source: `side`, `digest`; each reason: `side`, `code`.
- Each difference: `capability`, `scope`, `matcher`, `only_in`, `outcome`, `reason`.
- Each finding: `code`.

Empty strings remain `""` (including semantic differences' `only_in`); empty
collections are `[]`, never null or omitted. Validated accessor order is retained
using private ordered structs and standard `encoding/json` escaping, not maps.
Six whole-byte goldens cover exits 0/2/4, all report statuses, empty fields,
100 repeated renders per case, and returned-buffer independence. Existing audit
JSON is unchanged. Comparison text and CLI activation remain later slices.

Serialization adds no semantic equivalence, runtime enforcement, security,
anonymity, integrity, credential-detection, or compliance guarantee. Side-digest
origin is not verified; dictionary guessing and cross-report correlation risks
remain as described above.

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

Explicit CLI formats map redaction/render errors or empty output to only
`operational_failure` on stderr / exit `4`, never the rejected result or raw error. `internal/render/json.Marshal` now accepts
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
No-format audit/version results remain byte-compatible; help intentionally advertises new flags. Audit-only
`--format text` and `--format json` select safe reports after one application run.
The exact ordered grammar and examples are in the [README](../README.md#source-built-audit-output).
`--no-color` is a no-op only after `--format text`; JSON plus no-color is invalid.
Help/version do not accept formats; version JSON remains deferred. Exit 1 is reserved.
Explicit-format empty/flag-prefixed values are syntax errors; no-format values retain app validation.
Syntactic errors produce only `invalid_request` on stderr / exit 3 before application
I/O. Valid explicit requests render safe categories 0/2/3/4 entirely to stdout,
with empty stderr. A complete renderer buffer is written once; short/error writes
return 4 without appending diagnostics. Sentinel write failures also return 4.
JSON adds no comparison, enforcement, security,
compliance, anonymity, or other deferred guarantees. Renderer golden tests pin
required fields, safe digests, empty arrays, exits and 100 byte-identical repeats:
`go test ./internal/render/json -count=25`.

## Plain human layout

`internal/render/text.Render` accepts only valid safe reports and returns complete
bytes, or nil bytes with the fixed error `text: invalid report`. It preserves
redacted collection order. Lines appear in this order: Tool, Result, Findings,
Target, Requested version, Support status, Completeness, Source digests,
Permissions, Limitations. Empty sections and empty requested versions are omitted;
an empty target is `Target:`. Empty tool versions add no space after the tool name.
Findings contain codes only and precede the target so uncertainty stays visible.

Exact unsupported example (ending with one newline):

```text
Tool: universal-agent-policy-auditor
Result: unsupported_or_incomplete (exit 2)
Findings:
- redact-version-withheld
- unsupported-version
Target: opencode
Support status: unsupported
Completeness: incomplete
Limitations:
- modeled-static-configuration
```

Nonempty source and permission sections use this fixed layout (the digest below
is illustrative, not proof of source or evidence origin):

```text
Source digests:
- sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
Permissions:
- capability=bash; effect=ask; scope=opencode; matcher=*; provenance=[sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855]
```

Permission fields always use that order and punctuation; provenance contains only
digests, separated by `, ` if plural. Output has exactly one final newline and no
trailing spaces. There is no ANSI/color, terminal/TTY detection, width wrapping,
locale, environment, hostname, clock, path lookup, or pager dependency. Rendering
adds no guarantees beyond the safe-report boundary described above. Tests pin
success, unsupported/incomplete, invalid exit 3, operational exit 4, omitted
sections, canary absence and 100 independent byte-identical renders:
`go test ./internal/render/text -count=25`.
