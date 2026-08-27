---
name: work-unit-pr-policy
description: "Trigger: prepare PR, open PR, chained PR, Work Unit. Apply this repository's issue-linking policy to bounded PR slices."
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
---

## Activation Contract

Load this skill when preparing, reviewing, or opening a pull request in this repository.

## Hard Rules

- Use one approved issue for one coherent Work Unit, including all bounded PR slices that deliver it.
- Use `Refs #N` without a closing keyword on every intermediate PR. Use `Closes #N`, `Fixes #N`, `Resolves #N`, or an equivalent closing reference on exactly the final PR.
- Create separate issues for unrelated outcomes; never group them under one Work Unit issue.
- For this repository, this skill supersedes global `branch-pr` only where `branch-pr` requires a closing keyword on every PR. Keep every other applicable `branch-pr` rule active.
- Require the issue's `status:approved` label, exactly one `type:*` label, Conventional Commits, review evidence, chain and rollback context, and the 400 changed-line budget or an approved documented exception.

## Decision Gates

| PR role | Required issue reference |
| --- | --- |
| Intermediate bounded slice | `Refs #N`; do not close the issue |
| Final Work Unit slice | One closing reference for `#N` |
| Unrelated outcome | Separate approved issue |

## Execution Steps

1. Confirm the issue is approved and defines the coherent Work Unit.
2. Identify whether the PR is intermediate or final before selecting its issue reference.
3. Complete the PR template with reproducible review evidence, adjacent chain links, rollback boundary, scope exclusions, and changed-line budget status.
4. Verify exactly one matching `type:*` label and Conventional Commit formatting.
5. Confirm only the final PR closes the Work Unit issue.

## Output Contract

Return the approved issue number, PR role, issue-reference form, type label, changed-line count or exception, review evidence, and chain/rollback context.

## References

- `../../../.github/PULL_REQUEST_TEMPLATE.md`
- `../../../.github/ISSUE_TEMPLATE/feature_request.yml`
- `../../../AGENTS.md`
