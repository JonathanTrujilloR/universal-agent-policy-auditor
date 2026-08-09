## Exploration: universal-agent-policy-auditor

### Current State

The project has a reviewed PRD but no implementation, tests, or Git repository. The intended product is a local, read-only CLI that audits effective permissions for OpenCode and Claude Code, preserves source evidence, compares semantics, and fails visibly on unknown or lossy interpretation.

The product problem is valid and appropriately narrow for an alpha: users cannot safely infer equivalent protection from similar-looking client configuration. The first valuable outcome is a trustworthy per-client audit plus a bounded comparison of supported permission constructs—not policy generation or runtime enforcement.

Authoritative documentation already shows that the clients have materially different models:

- OpenCode documents a `permission` map with `allow`, `ask`, and `deny`, global and tool-specific rules, input-pattern matching, last-match-wins evaluation, agent overrides, defaults, auto mode, and an `external_directory` guard. Its docs also state that legacy `tools` configuration is deprecated but supported.
- Claude Code documents hierarchical user, project, local, and managed settings, with permission rules merged across scopes rather than treated like ordinary scalar settings. Its settings reference includes `permissions.allow`, `permissions.ask`, and `permissions.deny`, managed-only restrictions, version-dependent behavior, and tolerant handling of some invalid managed entries.

Therefore, “permission” cannot yet be treated as a shared three-valued rule. The alpha must model client-specific scope, precedence/merge behavior, matcher grammar, defaults, policy mode, version, and unresolved runtime context as first-class evidence. Documentation is a starting authority, not sufficient proof for every behavior; source inspection and executable conformance fixtures are required where docs are silent or version-sensitive.

### Affected Areas

- `PRD.md` — defines the product boundary, canonical-model goals, safety posture, open questions, and proposed CLI outcomes.
- `openspec/config.yaml` — establishes hybrid persistence, read-only safety rules, uncertainty semantics, and the absence of an implementation/test runner.
- `openspec/changes/universal-agent-policy-auditor/exploration.md` — this phase artifact; it records the MVP boundary and verification questions for proposal work.
- OpenCode permission documentation and repository source — authoritative inputs for discovery locations, rule matching, defaults, auto mode, agent overrides, and legacy compatibility.
- Claude Code settings and permissions documentation/source — authoritative inputs for scope discovery, merged permission rules, managed policies, settings precedence, version behavior, and invalid-entry handling.

### Approaches

1. **Native-configuration comparison only** — discover or accept explicit OpenCode and Claude Code sources, resolve each target’s effective modeled permissions, and compare only the supported native configurations. Do not author or consume a canonical policy in alpha.
   - Pros: smallest trustworthy boundary; avoids inventing a policy language before target semantics are understood; directly validates the core audit value; keeps provenance and loss reporting tied to real source rules; reduces false assurance and review surface.
   - Cons: no direct “does this meet our intended policy?” workflow; baseline/drift checks remain limited to client-to-client comparison or a later snapshot format; users may still duplicate intent externally.
   - Effort: Medium.

2. **Canonical authored policy in alpha** — define a user-authored canonical policy format alongside native adapters, then compare each client against that policy and optionally compare clients through it.
   - Pros: answers the policy-as-intent use case earlier; creates a direct foundation for future compilation and baseline checks; makes portability gaps explicit against a stable artifact.
   - Cons: expands the product contract before semantics and evidence rules are proven; risks a lowest-common-denominator language or premature target extensions; requires policy schema/versioning, authoring UX, validation, merge semantics, and migration decisions; increases false-equivalence and scope risk.
   - Effort: High.

3. **Native comparison with a non-authored evidence/baseline snapshot** — keep policy authoring out of alpha, but permit deterministic serialized audit results to be reviewed or compared later.
   - Pros: preserves the audit-first boundary while enabling reproducible CI-oriented workflows; separates observed configuration from asserted intent; supports future baseline design using real output experience.
   - Cons: snapshot semantics can be mistaken for a canonical policy; requires careful identity, redaction, environment metadata, and compatibility rules; still does not express desired intent.
   - Effort: Medium.

### Recommendation

Recommend **Approach 1 for the first valuable slice**, with the deterministic machine-readable audit result from Approach 3 as an output contract, not an authored policy language. The alpha should prove that the tool can discover, parse, resolve, explain, and safely compare a declared subset of native rules before asking users to trust a new policy abstraction.

The proposal should explicitly defer canonical policy authoring and compilation. It may reserve extension points for target-specific semantics and record a future baseline/snapshot direction, but it should not define a policy DSL yet. This keeps the MVP aligned with the PRD’s audit-only decision and makes the central release claim verifiable: the tool reports modeled configuration semantics and uncertainty without modifying files or claiming runtime enforcement.

### Verification Required Before Proposal

The following questions must be answered from current official documentation and, where needed, upstream source or executable conformance tests. Do not infer answers from syntax or from another client:

- What exact OpenCode discovery sources and merge order apply across global, project, agent, and Markdown agent configuration, and which versions support each construct?
- What exact OpenCode matcher language is implemented for `bash`, file paths, and other tool inputs; are patterns normalized, anchored, case-sensitive, or transformed before last-match evaluation?
- How do OpenCode defaults, `--auto`, session approvals, legacy `tools`, and agent-level permissions interact in the effective decision?
- What exact Claude Code settings sources are discoverable on Linux, macOS, and Windows, including managed files, drop-ins, environment-selected config directories, repository-root behavior, and version cutovers?
- Are Claude Code `allow`, `ask`, and `deny` rules merged, ordered, or independently evaluated; what is the precise precedence when the same matcher appears across scopes?
- Which Claude Code permission matcher forms exist (`Bash`, `Read`, and other tools), and what grammar, normalization, wildcard, path, command, and argument semantics do they use?
- How do workspace trust, managed-only restrictions, permission modes/auto mode, hooks, plugins, MCP, sandbox settings, and interactive approvals affect what can be concluded statically?
- How are malformed user/project/local settings treated versus malformed managed settings, and can a parser safely reproduce the client’s fail-closed or tolerant behavior?
- Which configuration and client versions can the alpha support with independent fixtures, and what evidence threshold is required before classifying two effective permissions as equivalent?

### Product Questions and Assumptions for Proposal

**Questions requiring a product decision:**

1. Is alpha comparison limited to native configuration-to-configuration audits, with canonical authored policy explicitly deferred?
2. Which minimum constructs and client versions define “useful” support, and which are immediately reported as unsupported?
3. Should the first comparison operate on explicit paths only, or include discovery once source-selection ambiguity is represented clearly?
4. What is the user-facing distinction between a blocking finding, a warning, and informational evidence?
5. Should deterministic audit snapshots be in alpha, and if so, are they observations only rather than policy expectations?
6. What path redaction and provenance detail are safe by default?
7. What is the minimum evidence required for `equivalent`, and should any unresolved condition make the result `not comparable`?

**Working assumptions to validate:**

- Users will accept “unknown” and “not comparable” when the alternative is false assurance.
- A narrow, documented support matrix is more valuable than broad syntax coverage.
- Effective static configuration semantics are useful even though they do not prove runtime enforcement.
- Explicit source paths are necessary for deterministic fixtures and should be supported even if discovery is also provided.
- Canonical policy authoring is a separate product decision, not an implicit consequence of normalization.

### Risks

- **Semantic overreach:** Similar `allow/ask/deny` labels conceal different scope, matcher, merge, default, and runtime behaviors; treating them as equivalent would violate the product safety posture.
- **Documentation drift:** Both clients evolve quickly and document version-specific behavior; unsupported versions or constructs must produce incomplete results rather than silent fallback.
- **Discovery leakage or ambiguity:** Global, project, local, managed, and agent sources can make “the configuration” plural and environment-dependent; source selection must be explicit and auditable.
- **Static-analysis limits:** Interactive approval, trust, hooks, plugins, sandboxing, and runtime context may prevent a complete effective-permission claim; the model needs conditional/unresolved states.
- **Premature canonical language:** An authored policy DSL could become a lowest-common-denominator or create false portability expectations before adapter semantics are validated.
- **Sensitive evidence:** Source paths, commands, environment-derived values, and settings may contain secrets or private repository details; redaction must be designed before machine output is stabilized.
- **Repository readiness:** There is no Git repository, implementation, or test runner yet; proposal work should include project/bootstrap decisions without pretending existing code patterns or test conventions exist.

### Ready for Proposal

**Yes, with explicit verification gates.** Proceed to proposal for a read-only native-configuration audit/comparison alpha, but require the proposal to name the supported client/version matrix, authoritative evidence plan, unsupported semantics policy, and canonical-policy deferral. Do not proceed to design or implementation assumptions that depend on matcher precedence, source discovery, or managed-policy behavior until those items are verified against current official docs and upstream source/fixtures.
