## Linked approved Work Unit issue

<!--
One approved issue governs one coherent Work Unit, even when delivery is split across bounded PR slices.
Replace N and keep exactly one link:
- Intermediate PR: `Refs #N` (MUST NOT close the Work Unit issue).
- Final PR: `Closes #N` (or an equivalent closing keyword). Exactly the final PR closes the issue.
-->
Refs #N

## PR type

<!-- Select exactly one and add the matching type:* label to this PR. -->
- [ ] Bug fix (`type:bug`)
- [ ] New feature (`type:feature`)
- [ ] Documentation only (`type:docs`)
- [ ] Code refactoring (`type:refactor`)
- [ ] Maintenance or tooling (`type:chore`)
- [ ] Breaking change (`type:breaking-change`)

## Summary

<!-- In 1-3 bullets, explain the outcome and why it matters. -->
- <!-- Summary item -->

## Changes

| File or area | Change |
|--------------|--------|
|              |        |

## Test plan

<!-- List reproducible checks and their results as review evidence. -->
- [ ] Relevant automated tests pass
- [ ] Affected behavior was verified manually, when applicable

## Chain and rollback context

<!-- Link adjacent PRs when chained. Write "Not chained" where an item does not apply. -->
- Previous PR: Not chained
- Next PR: Not chained
- Final Work Unit slice: Yes / No
- Rollback boundary: <!-- Explain how this slice can be reverted independently. -->

## Contributor checklist

- [ ] I linked one issue with the `status:approved` label that governs this coherent Work Unit.
- [ ] This PR does not combine outcomes unrelated to the Work Unit.
- [ ] I used `Refs #N` without a closing keyword for an intermediate PR, or `Closes #N` (or equivalent) for the final PR only.
- [ ] I selected exactly one PR type and added exactly one matching `type:*` label.
- [ ] My commits follow Conventional Commits and contain no `Co-Authored-By` trailers.
- [ ] I kept this PR within the 400 changed-line budget, or documented the approved exception.
- [ ] I documented chain context, review evidence, rollback boundaries, and anything intentionally out of scope.
- [ ] I added or updated tests for behavior changes, or explained why tests are not applicable.
- [ ] I updated documentation when user-facing behavior changed.
