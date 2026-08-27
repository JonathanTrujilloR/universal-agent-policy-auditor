# Evidence Manifest

## Authoritative documentation consulted (2026-08-09)

- OpenCode permissions: https://opencode.ai/docs/permissions/ — effects, defaults,
  wildcard matching, last-match precedence, and agent/global merge.
- Claude Code settings: https://code.claude.com/docs/en/settings — scopes, precedence,
  permission merging, and malformed managed-setting behavior.

## Upstream source evidence consulted

- https://github.com/anomalyco/opencode/blob/dev/packages/web/src/content/docs/permissions.mdx
- https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md

## Gate and assumptions

No version is supported. `matrix.json`, `fixtures.json`, and `evidence.json` are
loaded together by `support.Load`; all registries are intentionally empty. Current
documentation is not version-pinned executable fixture evidence, and this slice has
no adapters. Future entries must cite stable version evidence and a passing fixture.
