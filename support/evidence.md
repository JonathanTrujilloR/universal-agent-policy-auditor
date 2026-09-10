# OpenCode 1.18.27 Pinned Static Evidence

This file records a static evidence boundary only. It does not add production support: `support/matrix.json` keeps an empty `entries` array, `support.Load` validates fixture/evidence registries, and `support.Validate` for OpenCode `1.18.27` remains `unsupported-version` until a later slice adds and proves a matrix row.

## Authority and fixture

- Upstream repository: https://github.com/anomalyco/opencode
- Tag: `v1.18.27`
- Direct tag commit: `4b7e19e315cca414121ba1d61523fef74bb3ae8b`
- Release URL: https://github.com/anomalyco/opencode/releases/tag/v1.18.27
- Source URLs use commit-pinned form: `https://github.com/anomalyco/opencode/blob/4b7e19e315cca414121ba1d61523fef74bb3ae8b/<path>`
- Fixture ID: `opencode-1.18.27-legacy-permission-scalar`
- Evidence ID: `opencode-1.18.27-legacy-permission-scalar-source`
- Fixture path: `support/testdata/opencode-1.18.27-permission-legacy-scalar.json`
- Fixture SHA-256: `22ab8de00a73350aedcb72b62db5c962c910f15e12fbef80d844e725593bacf9`

Fixture bytes contain only this legacy scalar shape:

```json
{
  "permission": {
    "read": "allow",
    "edit": "deny",
    "bash": "ask"
  }
}
```

## Sources

| Source path | SHA-256 |
|---|---|
| [`packages/core/src/v1/config/config.ts`](https://github.com/anomalyco/opencode/blob/4b7e19e315cca414121ba1d61523fef74bb3ae8b/packages/core/src/v1/config/config.ts) | `b99bcbd98df6da79e59cda482363f759cea9d4b9792c6c8e83b6a8d686138d30` |
| [`packages/core/src/v1/config/permission.ts`](https://github.com/anomalyco/opencode/blob/4b7e19e315cca414121ba1d61523fef74bb3ae8b/packages/core/src/v1/config/permission.ts) | `f7669733d939c4affd38c4a27ce14deff970c7ffa4e26071ce6a3cd39ce0d4dd` |
| [`packages/core/src/v1/config/migrate.ts`](https://github.com/anomalyco/opencode/blob/4b7e19e315cca414121ba1d61523fef74bb3ae8b/packages/core/src/v1/config/migrate.ts) | `2a23a56469575d3edfe2044fe52468ee499d1c1c2281c34ebbcb624c7b25dae6` |
| [`packages/schema/src/permission.ts`](https://github.com/anomalyco/opencode/blob/4b7e19e315cca414121ba1d61523fef74bb3ae8b/packages/schema/src/permission.ts) | `229f2da7d245bf86deffed59b10cddf2dd04ffa05d10ef97fc5df289d0e6fb98` |
| [`packages/core/src/permission.ts`](https://github.com/anomalyco/opencode/blob/4b7e19e315cca414121ba1d61523fef74bb3ae8b/packages/core/src/permission.ts) | `f5d5b295452f2ddf17c475a24e77e302bb2b4b65d8a9e02a0f7bc5ef5504a0ba` |

## Claim mapping

| Claim ID | Static fact | Source mapping |
|---|---|---|
| `root-permission-legacy-info` | Root config has optional singular `permission` using legacy Info. | `packages/core/src/v1/config/config.ts`; `packages/core/src/v1/config/permission.ts` |
| `legacy-actions-scalar-resource-map` | Legacy actions are `ask`, `allow`, and `deny`; a legacy rule is either a scalar Action or a resource map. | `packages/core/src/v1/config/permission.ts`; `packages/schema/src/permission.ts` |
| `scalar-migrates-wildcard-action-effect` | Scalar legacy migration emits `{action, resource:"*", effect}`. | `packages/core/src/v1/config/migrate.ts`; `packages/core/src/v1/config/permission.ts` |
| `v2-ordered-last-wildcard-fallback-ask` | V2 uses ordered rules; the evaluator uses the last matching wildcard rule and falls back to ask when no rule decides. | `packages/schema/src/permission.ts`; `packages/core/src/permission.ts` |
| `runtime-effective-policy-out-of-scope` | Effective runtime policy also depends on resolved agent permissions, missing-agent behavior, saved rules, and resource aggregation. | `packages/core/src/permission.ts`; `packages/core/src/v1/config/config.ts` |

## What this proves

- The checked fixture is schema/migration-compatible by static inspection for singular legacy `permission` scalar actions `read:allow`, `edit:deny`, and `bash:ask` at OpenCode `1.18.27`.
- The evidence is pinned to one repository, tag, direct tag commit, fixture ID, evidence ID, source list, and exact source SHA-256 values.
- The V2 source hash used here is the corrected `229f2da7d245bf86deffed59b10cddf2dd04ffa05d10ef97fc5df289d0e6fb98` value.

## Explicit exclusions

This evidence does not prove V2 array config authoring, resource maps, tool names, agents or modes, defaults beyond the observed evaluator fallback, saved approvals, multiple config sources, source precedence, effective runtime behavior, enforcement, security guarantees, or compliance guarantees. The fixture was not executed upstream and is not evidence of OpenCode support until adapter conformance and a production matrix row are added later.
