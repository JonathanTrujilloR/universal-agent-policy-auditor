# Exploration: Claude Code Conformance Harness

## Executive summary

Create a separate deterministic conformance harness work unit that can execute an installed Claude Code binary in an isolated temporary workspace and produce privacy-safe, version-bound evidence. The harness must not promote Claude Code support by itself and must not mutate the blocked `universal-agent-policy-auditor` change, runtime ledger, repository settings, user settings, Git state, or issues.

The first valuable outcome is executable evidence for one narrow permission behavior: a target Bash matcher is requested as a structured tool use, the same matcher is denied despite an overlapping allow rule, and a sentinel proves the command did not execute. That evidence can later inform a maintainer decision on whether support metadata may be added; it is not itself a support claim.

## Current architecture context

| Area | Current fact | Implication for harness |
|---|---|---|
| Product boundary | The auditor is local, read-only, and currently models static configuration semantics only. | The harness belongs beside support/conformance evidence, not inside normal audit execution. |
| Adapter model | `internal/adapter/claudecode` already has a declarative discovery plan, parser, support gate, and unsupported-by-default behavior. | Keep production adapter support gated until executable evidence and maintainer approval exist. |
| Support gate | `support.Matrix` requires target, version, construct, fixture, evidence, and semantic booleans before support is complete. | Harness output should be a candidate evidence input to the matrix, not a matrix mutation. |
| Test style | Existing code uses small package boundaries, table-driven tests, `t.TempDir()`, copied immutable data, and fail-safe findings. | Design harness seams around command execution, filesystem isolation, parsing, redaction, and classification. |
| Main SDD change | `universal-agent-policy-auditor` remains blocked/partial for Claude support evidence. | This change must produce new evidence without resetting or bypassing that blocked objective. |

## Evidence objective

The harness should produce one reproducible evidence record for installed Claude Code version `2.1.245` or whatever executable is detected at runtime. Required observations:

- executable identity: resolved binary path after redaction, executable-reported version, and binary SHA-256 hash;
- isolated workspace: temporary project root, temporary HOME/config roots, generated `.claude/settings.json`, and no user/project settings read;
- scenario settings: overlapping permission rules where an allow matches broadly and a deny matches the target command;
- model prompt/request: a bounded instruction that can only satisfy the scenario by attempting the target Bash command;
- structured event stream: exact `tool_use` for the target Bash matcher, exact structured denial for the same matcher, and no successful tool result for the sentinel;
- sentinel proof: a file or marker that would exist only if the denied command executed remains absent;
- cleanup proof: temporary workspace/settings are removed or left only under an explicit caller-selected evidence directory;
- cost bound: one paid/model execution per accepted conformance attempt, with no unbounded retry loop.

## Preflight versus paid execution

Separate execution into two phases so invalid CLI flags, config syntax, or workspace construction fail before consuming the single conformance attempt.

| Phase | Allowed actions | Must not do | Output |
|---|---|---|---|
| CLI/config preflight | Resolve binary, hash binary, run free/help/version/config validation commands, create temp workspace, validate JSON/settings shape, dry-run command construction where supported. | Launch a model prompt, invoke paid inference, mutate repository/user settings. | `preflight_pass` or typed failure with version/hash and redacted diagnostics. |
| Conformance execution | Launch exactly one bounded Claude Code process in the isolated workspace with the validated command and settings. | Retry on semantic failure, read user settings, write repository files, execute dangerous commands. | `execution_pass`, `execution_fail`, or `execution_inconclusive` evidence record. |

If Claude Code lacks a true no-model validation mode for the final command shape, preflight should still validate every locally checkable input and classify the remaining boundary as `requires_paid_execution` rather than pretending it was validated.

## Proposed harness architecture

```text
cmd/claude-conformance? or internal/conformance/claudecode
  -> Runner(use case)
      -> BinaryResolver + Hasher
      -> WorkspaceFactory(t.TempDir / temp root)
      -> SettingsWriter(validated isolated .claude/settings.json)
      -> CommandBuilder(argv/env/cwd/stdin contract)
      -> ProcessRunner(interface seam)
      -> EventParser(JSONL / structured transcript only)
      -> Redactor
      -> EvidenceEncoder(stable JSON)
      -> Classifier
```

Recommended seams:

| Component | Responsibility | Deterministic tests |
|---|---|---|
| `BinaryResolver` | Finds an explicit or PATH Claude binary and captures version/hash. | Fake executable fixture and hash assertions. |
| `WorkspaceFactory` | Creates temp HOME, config, project root, sentinel path, and cleanup plan. | `t.TempDir()`, no real home, cleanup idempotence. |
| `SettingsWriter` | Writes minimal Claude settings with deny-over-allow scenario. | Golden JSON and malformed/settings failure tests. |
| `CommandBuilder` | Builds argv/env/cwd without shell interpolation. | Table-driven argv snapshots; invalid flag classification. |
| `ProcessRunner` | Runs external command only behind explicit integration gate. | Unit fake for stdout/stderr/exit/timeouts; short-mode skip for real command. |
| `EventParser` | Accepts only structured events, never free-form model claims. | JSONL fixtures for tool_use, denial, malformed, missing sentinel. |
| `Redactor` | Removes prompts, private paths, usernames, hostnames, tokens, raw env, and raw transcript. | Secret fixtures and stable redacted output. |
| `Classifier` | Maps evidence to pass/fail/inconclusive with reason codes. | Table-driven failure classification. |
| `EvidenceEncoder` | Emits schema-versioned, sorted, byte-stable JSON. | Golden tests, rerun without update. |

The normal auditor CLI should not depend on Claude Code. Keep executable conformance as a separate command, test package, or maintainer-only harness so product auditing remains offline/read-only by default.

## Evidence schema sketch

```json
{
  "schema_version": "claudecode-conformance-evidence/v1",
  "target": "claudecode",
  "construct": "permission",
  "binary": {
    "version": "2.1.245",
    "sha256": "sha256:<hex>",
    "path_redacted": "<redacted>"
  },
  "scenario": {
    "id": "bash-deny-overlapping-allow",
    "matcher": "Bash(<target command pattern>)",
    "expected_tool_use": true,
    "expected_denial": true,
    "expected_sentinel_absent": true
  },
  "preflight": {
    "status": "pass",
    "validated": ["binary", "version", "hash", "settings-json", "argv", "workspace-isolation"]
  },
  "execution": {
    "status": "pass|fail|inconclusive|not_run",
    "events_observed": ["tool_use", "permission_denial"],
    "sentinel_executed": false,
    "attempt_count": 1
  },
  "redaction": {
    "raw_transcript_stored": false,
    "system_prompt_stored": false,
    "secrets_detected": false
  },
  "classification": {
    "result": "evidence_candidate",
    "reason_codes": []
  }
}
```

Do not store raw transcript, full prompt, absolute paths, usernames, hostnames, environment values, credentials, repository source contents, or system prompts. Store only normalized event facts required for reproducibility.

## Success criteria

A conformance run may be considered a candidate evidence success only when all are true:

1. Preflight passes without model execution.
2. The binary version and binary hash are captured.
3. The workspace and settings are created under isolated temp roots; no user/project settings are used.
4. The final execution runs once with a bounded timeout and bounded output capture.
5. Structured events show the target Bash `tool_use` matcher.
6. Structured events show denial for the same matcher even with an overlapping allow present.
7. The sentinel remains absent, proving the denied command did not execute.
8. The evidence JSON validates against the schema and is byte-stable after redaction.
9. Cleanup succeeds or any retained evidence path is explicit and outside the product repository.
10. The result is labeled `evidence_candidate`, not `supported`.

## Failure classifications

| Code | Meaning | Maintainer impact |
|---|---|---|
| `preflight_binary_missing` | Claude executable not found or not executable. | No paid attempt consumed. |
| `preflight_version_unavailable` | Version cannot be captured deterministically. | No support decision. |
| `preflight_hash_failed` | Binary hash cannot be computed. | No support decision. |
| `preflight_invalid_settings` | Generated settings rejected or malformed. | Fix harness/settings before execution. |
| `preflight_invalid_flags` | CLI command shape unsupported by installed version. | Fix command builder before execution. |
| `execution_timeout` | Process exceeded bound. | Inconclusive; do not retry automatically. |
| `execution_no_structured_events` | Only free-form text or malformed stream observed. | Inconclusive; evidence threshold not met. |
| `execution_no_tool_use` | Target Bash request not observed. | Fail/inconclusive depending on prompt determinism. |
| `execution_no_denial` | Denial event absent. | Fail; cannot prove deny behavior. |
| `execution_sentinel_executed` | Sentinel was created or command succeeded. | Critical fail; deny did not protect. |
| `redaction_failed` | Evidence may contain private data. | Block persistence/publication. |
| `cleanup_failed` | Temp material remains unexpectedly. | Operational failure; inspect before proceeding. |

## Threats and non-goals

- Do not promote production Claude Code support in the auditor.
- Do not reset, bypass, or redefine the blocked `universal-agent-policy-auditor` runtime objective.
- Do not read or write user Claude settings, project settings, global Git config, source files, issues, runtime ledgers, or existing SDD artifacts.
- Do not store raw transcript, raw prompt, system prompt, secrets, absolute paths, usernames, hostnames, or raw environment.
- Do not execute dangerous commands. Use a harmless sentinel command under the temp workspace only.
- Do not rely on free-form model statements as evidence.
- Do not run unbounded retries. One conformance execution per authorized attempt.
- Do not claim that one scenario proves full Claude Code permission support; it only proves the observed behavior for the pinned binary/scenario.

## Path to maintainer decision

This work can unblock a future maintainer decision only by producing genuinely new, reproducible evidence:

1. deterministic harness code and fixtures are reviewed separately;
2. a maintainer authorizes one execution against an installed Claude Code binary;
3. the harness produces a redacted evidence record with version/hash and pass/fail classification;
4. the maintainer decides whether that evidence is sufficient to add or reject a specific `support.Matrix` row in the main change.

The decision remains human-owned. The harness output must not mutate the support matrix or check off blocked tasks automatically.

## Review workload and likely slices

Keep each PR under the 400-line review budget by slicing around independent work units:

| Slice | Goal | Expected files | Runtime execution |
|---|---|---|---|
| 1 | Schema, parser, redactor, classifier fixtures. | `internal/conformance/claudecode` tests/fixtures or `support/conformance` fixtures. | None. |
| 2 | Workspace, settings, binary resolver/hasher, command builder with fake runner. | Harness package and unit tests. | None. |
| 3 | Optional maintainer CLI/integration gate and evidence writer. | Small command wrapper plus docs. | Preflight only by default; paid execution explicit. |
| 4 | Authorized evidence run and support-matrix proposal, if accepted. | Evidence artifact/support metadata in later change. | One explicit run, not automatic. |

Recommended next SDD phase: proposal. The proposal should keep the first implemented slice to deterministic harness internals and fixtures; real Claude/model execution should remain a later explicit maintainer action.
