# Claude Code Conformance Specification

## Purpose

Define the required behavior for a separate Claude Code conformance harness that produces deterministic, privacy-safe `evidence_candidate` records for maintainer review without changing normal auditor execution, support metadata, Git/runtime ledgers, issues, or the blocked `universal-agent-policy-auditor` change.

## Requirements

### Requirement: First Slice Deterministic Scope

The system MUST keep the first implementation slice limited to deterministic schema, parser, redactor, classifier, fixtures, and fakes, and MUST NOT run Claude Code, Claude, or any model process in that slice.

#### Scenario: Fixture-only first slice

- GIVEN the first slice is implemented or tested
- WHEN conformance behavior is exercised
- THEN every input MUST come from fixtures or fakes
- AND no Claude Code, Claude, model, paid inference, external process, maintainer execution gate, Git ledger mutation, issue mutation, or blocked main-change mutation MUST occur.

#### Scenario: Deferred executable integration remains out of first slice

- GIVEN requirements describe future workspace construction, binary resolution, command construction, preflight, or authorized execution
- WHEN first-slice tasks are planned or implemented
- THEN those executable capabilities MUST remain deferred and MUST NOT be reachable from first-slice code paths.

### Requirement: Stable Evidence Schema and Encoding

The system MUST emit schema-versioned conformance evidence with deterministic field names, deterministic ordering, stable JSON encoding, and bounded output suitable for golden-file comparison.

#### Scenario: Byte-stable evidence output

- GIVEN the same normalized conformance result is encoded twice
- WHEN the evidence record is serialized
- THEN both serializations MUST be byte-identical
- AND the record MUST include its schema version, result classification, reason code, scenario identity, evidence facts, redaction status, execution boundary, and artifact identity fields where applicable.

#### Scenario: Unsupported classification excluded

- GIVEN any conformance evidence record is produced
- WHEN the record is serialized or persisted
- THEN it MUST classify the result as `evidence_candidate`, `fail`, or `inconclusive`
- AND it MUST NOT classify or imply Claude Code as `supported`.

### Requirement: Structured Event Parsing and Exact Matcher Correlation

The system MUST parse structured Claude Code event fixtures into normalized tool-use and denial facts and MUST correlate denials only when the same exact Bash matcher appears in both the target `tool_use` event and the denial event.

#### Scenario: Matching tool use and denial

- GIVEN structured events contain a target Bash `tool_use` matcher and a denial event for the same exact matcher
- WHEN the parser and classifier evaluate the events
- THEN the normalized evidence MUST record both events as correlated
- AND the matcher comparison MUST preserve exact command semantics rather than using partial, fuzzy, or display-only matching.

#### Scenario: Denial for a different matcher

- GIVEN structured events contain the target Bash `tool_use` matcher
- AND a denial event references a different matcher
- WHEN the classifier evaluates the events
- THEN the result MUST NOT be successful candidate evidence
- AND the reason code MUST identify a matcher-correlation failure.

### Requirement: Denial Correlation and Sentinel Proof

The system MUST require both correlated denial evidence and sentinel absence before producing successful candidate evidence.

#### Scenario: Denied command did not execute

- GIVEN structured events contain the target Bash `tool_use` matcher
- AND structured events contain the exact denial for that matcher
- AND sentinel inspection proves the sentinel is absent
- WHEN the result is classified
- THEN the result MAY be classified as `evidence_candidate` if all other evidence thresholds pass.

#### Scenario: Sentinel present or unknown

- GIVEN matcher and denial events appear correlated
- WHEN sentinel inspection reports the sentinel present or cannot prove absence
- THEN the result MUST NOT be classified as successful candidate evidence
- AND the reason code MUST identify either command execution or inconclusive sentinel proof.

### Requirement: Malformed, Missing, and Contradictory Event Handling

The system MUST classify malformed, missing, unsupported, duplicate, ambiguous, or contradictory structured event inputs with explicit failure or inconclusive reason codes rather than silently accepting them.

#### Scenario: Missing required event

- GIVEN evidence input lacks the target Bash `tool_use` event or lacks the correlated denial event
- WHEN the classifier evaluates the evidence
- THEN the result MUST be `fail` or `inconclusive`
- AND the reason code MUST identify the missing event category.

#### Scenario: Malformed structured event

- GIVEN an event fixture cannot be parsed into the expected structured event shape
- WHEN the parser evaluates the fixture
- THEN the parser MUST return a structured parse failure
- AND persistence MUST NOT claim successful candidate evidence from that input.

#### Scenario: Contradictory event facts

- GIVEN events simultaneously claim a target matcher was denied and successfully executed
- WHEN the classifier evaluates the evidence
- THEN the result MUST NOT be `evidence_candidate`
- AND the reason code MUST identify contradictory evidence.

### Requirement: Strict Redaction and Persistence Blocking

The system MUST redact or reject sensitive local context before persistence, and MUST block evidence persistence when leakage is detected or redaction confidence is insufficient.

#### Scenario: Sensitive values are removed

- GIVEN evidence input contains absolute paths, usernames, hostnames, tokens, credentials, environment values, raw prompts, raw transcripts, system prompts, or repository source contents
- WHEN redaction is applied
- THEN persisted evidence MUST NOT contain those raw values
- AND replacement markers MUST be deterministic and non-reversible.

#### Scenario: Leakage blocks persistence

- GIVEN redaction detects unredacted sensitive content or cannot determine whether sensitive content remains
- WHEN evidence persistence is requested
- THEN persistence MUST be blocked
- AND the result MUST include a redaction failure or inconclusive redaction reason code.

### Requirement: Explicit Classification and Reason Codes

The system MUST return explicit pass, fail, and inconclusive reason codes for every classification path, including parser failures, matcher mismatch, missing denial, sentinel uncertainty, redaction failure, preflight failure, authorization absence, execution bounds exceeded, cleanup failure, and unsupported support-claim attempts.

#### Scenario: Reason code accompanies every terminal result

- GIVEN the harness classifies normalized evidence
- WHEN the classification is returned
- THEN the output MUST include exactly one primary reason code
- AND MAY include secondary diagnostic reason codes when they do not change the primary classification.

#### Scenario: Inconclusive remains distinct from failure

- GIVEN required evidence cannot be observed because input is incomplete, preflight is unavailable, sentinel status is unknown, or cleanup evidence is unavailable
- WHEN the result is classified
- THEN the result SHOULD be `inconclusive` rather than `fail` when the observed data does not prove non-conformance.

### Requirement: Preflight and Paid Execution Boundary

The system MUST separate preflight checks from paid/model execution. Preflight MAY inspect deterministic local prerequisites, but MUST NOT invoke Claude Code in a way that consumes model inference or executes the conformance scenario.

#### Scenario: Preflight does not spend an attempt

- GIVEN a maintainer requests preflight for a future executable slice
- WHEN preflight validates paths, binary availability metadata, output locations, bounds, and isolation prerequisites
- THEN it MUST NOT run the conformance prompt or consume paid/model execution
- AND it MUST report only preflight status and reason codes.

#### Scenario: Paid execution requires explicit authorization

- GIVEN no explicit maintainer authorization exists for the candidate run
- WHEN executable conformance is requested
- THEN the system MUST refuse to run Claude Code or any model process
- AND the reason code MUST identify missing execution authorization.

### Requirement: Single-Attempt Authorization and Bounded Execution

The system MUST allow at most one explicitly authorized conformance execution per authorized evidence run and MUST bound runtime, output size, retained artifacts, and process behavior.

#### Scenario: Authorized single attempt

- GIVEN a maintainer explicitly authorizes one conformance execution
- WHEN the executable harness runs in a later slice
- THEN it MUST perform no more than one attempt for that authorization
- AND it MUST enforce configured runtime and output bounds.

#### Scenario: Bounds exceeded

- GIVEN an authorized execution exceeds runtime, output, artifact, or process bounds
- WHEN the harness terminates or classifies the run
- THEN the result MUST NOT be `evidence_candidate`
- AND the reason code MUST identify the exceeded bound.

### Requirement: Binary Identity and Isolated Runtime Context

The system MUST bind executable evidence to the observed Claude Code binary version and SHA-256 hash, and MUST run future executable checks only with isolated temporary HOME, config, settings, and project roots that do not read user or repository Claude settings.

#### Scenario: Binary identity captured

- GIVEN a future authorized execution observes an installed Claude Code binary
- WHEN evidence is produced
- THEN the evidence MUST include the binary version and SHA-256 hash
- AND MUST identify the binary path only through a redacted or safe representation.

#### Scenario: User settings are isolated

- GIVEN future executable conformance constructs a runtime workspace
- WHEN Claude Code is invoked
- THEN HOME, config, settings, and project roots MUST be isolated temporary locations
- AND user settings, repository settings, environment secrets, and unrelated project files MUST NOT be read into evidence or execution context.

### Requirement: Cleanup Evidence

The system MUST record cleanup evidence for temporary workspaces and settings used by future executable slices, and MUST retain artifacts only when the caller explicitly selected an evidence directory outside normal product execution.

#### Scenario: Cleanup succeeds

- GIVEN a future executable conformance run creates temporary workspace, HOME, config, or settings paths
- WHEN the run completes or aborts
- THEN cleanup MUST be attempted
- AND evidence MUST record whether cleanup succeeded without exposing sensitive paths.

#### Scenario: Cleanup cannot be proven

- GIVEN cleanup fails or cannot be verified
- WHEN evidence is classified
- THEN the result MUST be `fail` or `inconclusive`
- AND retained evidence MUST exist only in an explicit caller-selected evidence directory.

### Requirement: Separation from Normal Auditor Execution and Blocked Main Change

The system MUST keep conformance harness behavior separate from normal auditor execution and MUST NOT mutate support metadata, normal auditor semantics, runtime ledgers, Git state, issues, unrelated OpenSpec changes, or the blocked `universal-agent-policy-auditor` change.

#### Scenario: Normal auditor remains unchanged

- GIVEN a user runs the normal auditor CLI or support matrix behavior
- WHEN this conformance harness exists
- THEN normal audit behavior MUST remain offline, read-only, and unchanged
- AND Claude Code MUST remain unsupported-by-default unless a separate maintainer-reviewed change updates support metadata.

#### Scenario: Harness cannot complete blocked main change

- GIVEN conformance evidence exists for maintainers to inspect
- WHEN the blocked `universal-agent-policy-auditor` change is evaluated
- THEN the evidence MUST NOT automatically reset, bypass, complete, or mutate that change
- AND any support decision MUST happen through a separate maintainer action.

### Requirement: Rollback-Friendly PR Slicing

The system MUST preserve stacked, rollback-friendly delivery boundaries with each authored PR slice planned under the 400-line review budget.

#### Scenario: First slice rollback

- GIVEN the first slice adds deterministic schema, parser, redactor, classifier, and fixtures only
- WHEN that slice is rolled back
- THEN removing those internals and fixtures MUST leave normal auditor behavior, support metadata, runtime ledgers, settings, Git state, issues, and unrelated SDD changes unaffected.

#### Scenario: Later executable slices remain independently removable

- GIVEN later slices add workspace construction, command construction, preflight, execution gate, or evidence writing
- WHEN one later slice is rolled back
- THEN rollback MUST be possible without removing unrelated previous deterministic internals or mutating normal auditor behavior.
