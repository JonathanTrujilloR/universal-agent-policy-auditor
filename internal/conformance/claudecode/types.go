package claudecode

type Result string

const (
	ResultUnknown           Result = ""
	ResultEvidenceCandidate Result = "evidence_candidate"
	ResultFail              Result = "fail"
	ResultInconclusive      Result = "inconclusive"
)

type ReasonCode string

const (
	ReasonUnknown                       ReasonCode = ""
	ReasonEvidenceCandidate             ReasonCode = "evidence_candidate"
	ReasonFixtureNotCandidateEvidence   ReasonCode = "fixture_not_candidate_evidence"
	ReasonFakeNotCandidateEvidence      ReasonCode = "fake_not_candidate_evidence"
	ReasonPreflightNotCandidateEvidence ReasonCode = "preflight_not_candidate_evidence"
	ReasonParserMalformedEvent          ReasonCode = "parser_malformed_event"
	ReasonParserNoStructuredEvents      ReasonCode = "parser_no_structured_events"
	ReasonParserBoundsExceeded          ReasonCode = "parser_bounds_exceeded"
	ReasonUnsupportedStructuredEvent    ReasonCode = "unsupported_structured_event"
	ReasonDisplayOnlyEvent              ReasonCode = "display_only_event"
	ReasonMissingToolUse                ReasonCode = "missing_tool_use"
	ReasonDenialWithoutToolUse          ReasonCode = "denial_without_tool_use"
	ReasonMissingDenial                 ReasonCode = "missing_denial"
	ReasonMatcherCorrelationFailed      ReasonCode = "matcher_correlation_failed"
	ReasonToolUseIDCorrelationFailed    ReasonCode = "tool_use_id_correlation_failed"
	ReasonDuplicateTargetEvent          ReasonCode = "duplicate_target_event"
	ReasonContradictoryExecutionSuccess ReasonCode = "contradictory_execution_success"
	ReasonSentinelPresent               ReasonCode = "sentinel_present"
	ReasonSentinelUnknown               ReasonCode = "sentinel_unknown"
	ReasonRedactionFailed               ReasonCode = "redaction_failed"
	ReasonRuntimeIdentityMissing        ReasonCode = "runtime_identity_missing"
	ReasonRuntimeIsolationUnproven      ReasonCode = "runtime_isolation_unproven"
	ReasonExecutionAuthorizationMissing ReasonCode = "execution_authorization_missing"
	ReasonExecutionAttemptCountInvalid  ReasonCode = "execution_attempt_count_invalid"
	ReasonPreflightBinaryMissing        ReasonCode = "preflight_binary_missing"
	ReasonPreflightVersionUnavailable   ReasonCode = "preflight_version_unavailable"
	ReasonPreflightHashFailed           ReasonCode = "preflight_hash_failed"
	ReasonPreflightInvalidSettings      ReasonCode = "preflight_invalid_settings"
	ReasonPreflightInvalidFlags         ReasonCode = "preflight_invalid_flags"
	ReasonExecutionTimeout              ReasonCode = "execution_timeout"
	ReasonExecutionOutputBoundsExceeded ReasonCode = "execution_output_bounds_exceeded"
	ReasonCleanupFailed                 ReasonCode = "cleanup_failed"
	ReasonUnsupportedSupportClaim       ReasonCode = "unsupported_support_claim"
	ReasonEncodingSchemaUnproven        ReasonCode = "encoding_schema_unproven"
	ReasonEncodingStabilityUnproven     ReasonCode = "encoding_stability_unproven"
	ReasonProvenanceUnknown             ReasonCode = "provenance_unknown"
)

type Provenance string

const (
	ProvenanceUnknown             Provenance = ""
	ProvenanceFixture             Provenance = "fixture"
	ProvenanceFake                Provenance = "fake"
	ProvenancePreflight           Provenance = "preflight"
	ProvenanceAuthorizedExecution Provenance = "authorized_execution"
)

type SentinelState string

const (
	SentinelUnknown SentinelState = ""
	SentinelAbsent  SentinelState = "absent"
	SentinelPresent SentinelState = "present"
)

type RedactionStatus string

const (
	RedactionUnknown RedactionStatus = ""
	RedactionPassed  RedactionStatus = "passed"
	RedactionFailed  RedactionStatus = "failed"
)

type EvidenceFacts struct {
	TargetToolUseObserved      bool
	CorrelatedDenialObserved   bool
	MatcherMatched             bool
	ToolUseIDMatched           bool
	TargetExecutionSucceeded   bool
	DuplicateTargetEvent       bool
	UnsupportedStructuredEvent bool
	DisplayOnlyEvent           bool
}

type RuntimeProof struct {
	MaintainerAuthorized  bool
	AttemptCount          int
	BinaryVersionCaptured bool
	BinaryHashCaptured    bool
	IsolatedRuntime       bool
	CleanupSucceeded      bool
}

type ParserProof struct {
	StructuredEventsParsed bool
	NoMalformedEvents      bool
	BoundsHeld             bool
	NoUnsupportedEvents    bool
}

type EncodingProof struct {
	SchemaVersioned bool
	ByteStable      bool
	Bounded         bool
	Safe            bool
}

type ClassificationInput struct {
	Provenance     Provenance
	Facts          EvidenceFacts
	Parser         ParserProof
	Sentinel       SentinelState
	Encoding       EncodingProof
	Redaction      RedactionStatus
	Runtime        RuntimeProof
	SupportClaimed bool
}

type Classification struct {
	Result        Result
	PrimaryReason ReasonCode
}
