package claudecode

import "testing"

func TestDomainZeroValuesAreUnknownSafe(t *testing.T) {
	var sentinel SentinelState
	if Result("") == ResultEvidenceCandidate || ReasonCode("") == ReasonEvidenceCandidate || Provenance("") == ProvenanceAuthorizedExecution || sentinel != SentinelUnknown {
		t.Fatalf("zero domain values are not unknown-safe")
	}
}

func TestClassifyRejectsNonExecutionProvenanceAsCandidateEvidence(t *testing.T) {
	for _, tt := range []struct {
		provenance Provenance
		want       ReasonCode
	}{
		{ProvenanceFixture, ReasonFixtureNotCandidateEvidence},
		{ProvenanceFake, ReasonFakeNotCandidateEvidence},
		{ProvenancePreflight, ReasonPreflightNotCandidateEvidence},
	} {
		t.Run(string(tt.provenance), func(t *testing.T) {
			assertClassifies(t, candidateInput(tt.provenance), ResultInconclusive, tt.want)
		})
	}
}

func TestClassifyAuthorizedExecutionRequiresEveryGate(t *testing.T) {
	assertClassifies(t, candidateInput(ProvenanceAuthorizedExecution), ResultEvidenceCandidate, ReasonEvidenceCandidate)
	for _, tt := range []struct {
		name   string
		mutate func(*ClassificationInput)
		want   ReasonCode
	}{
		{"missing authorization", func(in *ClassificationInput) { in.Runtime.MaintainerAuthorized = false }, ReasonExecutionAuthorizationMissing},
		{"missing runtime version", func(in *ClassificationInput) { in.Runtime.BinaryVersionCaptured = false }, ReasonRuntimeIdentityMissing},
		{"missing runtime hash", func(in *ClassificationInput) { in.Runtime.BinaryHashCaptured = false }, ReasonRuntimeIdentityMissing},
		{"isolation unproven", func(in *ClassificationInput) { in.Runtime.IsolatedRuntime = false }, ReasonRuntimeIsolationUnproven},
		{"invalid attempt count", func(in *ClassificationInput) { in.Runtime.AttemptCount = 2 }, ReasonExecutionAttemptCountInvalid},
		{"cleanup failed", func(in *ClassificationInput) { in.Runtime.CleanupSucceeded = false }, ReasonCleanupFailed},
		{"parser proof absent", func(in *ClassificationInput) { in.Parser = ParserProof{} }, ReasonParserNoStructuredEvents},
		{"malformed event not excluded", func(in *ClassificationInput) { in.Parser.NoMalformedEvents = false }, ReasonParserMalformedEvent},
		{"parser bounds not proven", func(in *ClassificationInput) { in.Parser.BoundsHeld = false }, ReasonParserBoundsExceeded},
		{"unsupported events not excluded", func(in *ClassificationInput) { in.Parser.NoUnsupportedEvents = false }, ReasonUnsupportedStructuredEvent},
		{"encoding proof absent", func(in *ClassificationInput) { in.Encoding = EncodingProof{} }, ReasonEncodingSchemaUnproven},
		{"encoding not byte stable", func(in *ClassificationInput) { in.Encoding.ByteStable = false }, ReasonEncodingStabilityUnproven},
		{"encoded output unbounded", func(in *ClassificationInput) { in.Encoding.Bounded = false }, ReasonExecutionOutputBoundsExceeded},
		{"encoded output unsafe", func(in *ClassificationInput) { in.Encoding.Safe = false }, ReasonRedactionFailed},
	} {
		t.Run(tt.name, func(t *testing.T) {
			input := candidateInput(ProvenanceAuthorizedExecution)
			tt.mutate(&input)
			assertClassifies(t, input, ResultInconclusive, tt.want)
		})
	}
}

func TestClassifyFailsClosedForCorrelationSentinelAndPrivacyProblems(t *testing.T) {
	for _, tt := range []struct {
		name       string
		mutate     func(*ClassificationInput)
		wantResult Result
		wantReason ReasonCode
	}{
		{"matcher mismatch", func(in *ClassificationInput) { in.Facts.MatcherMatched = false }, ResultFail, ReasonMatcherCorrelationFailed},
		{"tool use id mismatch", func(in *ClassificationInput) { in.Facts.ToolUseIDMatched = false }, ResultFail, ReasonToolUseIDCorrelationFailed},
		{"unsupported structured event", func(in *ClassificationInput) { in.Facts.UnsupportedStructuredEvent = true }, ResultInconclusive, ReasonUnsupportedStructuredEvent},
		{"duplicate target event", func(in *ClassificationInput) { in.Facts.DuplicateTargetEvent = true }, ResultInconclusive, ReasonDuplicateTargetEvent},
		{"missing tool use", func(in *ClassificationInput) {
			in.Facts.TargetToolUseObserved, in.Facts.CorrelatedDenialObserved = false, false
		}, ResultInconclusive, ReasonMissingToolUse},
		{"missing denial", func(in *ClassificationInput) { in.Facts.CorrelatedDenialObserved = false }, ResultFail, ReasonMissingDenial},
		{"denial without tool use", func(in *ClassificationInput) {
			in.Facts.TargetToolUseObserved, in.Facts.CorrelatedDenialObserved = false, true
		}, ResultInconclusive, ReasonDenialWithoutToolUse},
		{"contradictory success", func(in *ClassificationInput) { in.Facts.TargetExecutionSucceeded = true }, ResultFail, ReasonContradictoryExecutionSuccess},
		{"sentinel present", func(in *ClassificationInput) { in.Sentinel = SentinelPresent }, ResultFail, ReasonSentinelPresent},
		{"sentinel unknown", func(in *ClassificationInput) { in.Sentinel = SentinelUnknown }, ResultInconclusive, ReasonSentinelUnknown},
		{"redaction failure", func(in *ClassificationInput) { in.Redaction = RedactionFailed }, ResultInconclusive, ReasonRedactionFailed},
		{"unsupported support claim", func(in *ClassificationInput) { in.SupportClaimed = true }, ResultFail, ReasonUnsupportedSupportClaim},
	} {
		t.Run(tt.name, func(t *testing.T) {
			input := candidateInput(ProvenanceAuthorizedExecution)
			tt.mutate(&input)
			assertClassifies(t, input, tt.wantResult, tt.wantReason)
		})
	}
}

func assertClassifies(t *testing.T, input ClassificationInput, result Result, reason ReasonCode) {
	t.Helper()
	got := Classify(input)
	if got.Result != result || got.PrimaryReason != reason {
		t.Fatalf("Classify = %+v, want %s/%s", got, result, reason)
	}
}

func candidateInput(provenance Provenance) ClassificationInput {
	return ClassificationInput{
		Provenance: provenance,
		Facts: EvidenceFacts{
			TargetToolUseObserved:    true,
			CorrelatedDenialObserved: true,
			MatcherMatched:           true,
			ToolUseIDMatched:         true,
		},
		Parser: ParserProof{
			StructuredEventsParsed: true,
			NoMalformedEvents:      true,
			BoundsHeld:             true,
			NoUnsupportedEvents:    true,
		},
		Sentinel: SentinelAbsent,
		Encoding: EncodingProof{
			SchemaVersioned: true,
			ByteStable:      true,
			Bounded:         true,
			Safe:            true,
		},
		Redaction: RedactionPassed,
		Runtime: RuntimeProof{
			MaintainerAuthorized:  true,
			AttemptCount:          1,
			BinaryVersionCaptured: true,
			BinaryHashCaptured:    true,
			IsolatedRuntime:       true,
			CleanupSucceeded:      true,
		},
	}
}
