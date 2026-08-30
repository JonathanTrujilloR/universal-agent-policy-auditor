package claudecode

func Classify(input ClassificationInput) Classification {
	switch {
	case input.SupportClaimed:
		return failed(ReasonUnsupportedSupportClaim)
	case input.Redaction != RedactionPassed:
		return inconclusive(ReasonRedactionFailed)
	case !input.Parser.StructuredEventsParsed:
		return inconclusive(ReasonParserNoStructuredEvents)
	case !input.Parser.NoMalformedEvents:
		return inconclusive(ReasonParserMalformedEvent)
	case !input.Parser.BoundsHeld:
		return inconclusive(ReasonParserBoundsExceeded)
	case !input.Parser.NoUnsupportedEvents || input.Facts.UnsupportedStructuredEvent:
		return inconclusive(ReasonUnsupportedStructuredEvent)
	case input.Facts.DisplayOnlyEvent:
		return inconclusive(ReasonDisplayOnlyEvent)
	case input.Facts.DuplicateTargetEvent:
		return inconclusive(ReasonDuplicateTargetEvent)
	case input.Facts.TargetExecutionSucceeded:
		return failed(ReasonContradictoryExecutionSuccess)
	case !input.Facts.TargetToolUseObserved && input.Facts.CorrelatedDenialObserved:
		return inconclusive(ReasonDenialWithoutToolUse)
	case !input.Facts.TargetToolUseObserved:
		return inconclusive(ReasonMissingToolUse)
	case !input.Facts.CorrelatedDenialObserved:
		return failed(ReasonMissingDenial)
	case !input.Facts.MatcherMatched:
		return failed(ReasonMatcherCorrelationFailed)
	case !input.Facts.ToolUseIDMatched:
		return failed(ReasonToolUseIDCorrelationFailed)
	case input.Sentinel == SentinelPresent:
		return failed(ReasonSentinelPresent)
	case input.Sentinel != SentinelAbsent:
		return inconclusive(ReasonSentinelUnknown)
	case !input.Encoding.SchemaVersioned:
		return inconclusive(ReasonEncodingSchemaUnproven)
	case !input.Encoding.ByteStable:
		return inconclusive(ReasonEncodingStabilityUnproven)
	case !input.Encoding.Bounded:
		return inconclusive(ReasonExecutionOutputBoundsExceeded)
	case !input.Encoding.Safe:
		return inconclusive(ReasonRedactionFailed)
	}

	switch input.Provenance {
	case ProvenanceFixture:
		return inconclusive(ReasonFixtureNotCandidateEvidence)
	case ProvenanceFake:
		return inconclusive(ReasonFakeNotCandidateEvidence)
	case ProvenancePreflight:
		return inconclusive(ReasonPreflightNotCandidateEvidence)
	case ProvenanceAuthorizedExecution:
		return classifyAuthorizedExecution(input.Runtime)
	default:
		return inconclusive(ReasonProvenanceUnknown)
	}
}

func classifyAuthorizedExecution(runtime RuntimeProof) Classification {
	switch {
	case !runtime.MaintainerAuthorized:
		return inconclusive(ReasonExecutionAuthorizationMissing)
	case runtime.AttemptCount != 1:
		return inconclusive(ReasonExecutionAttemptCountInvalid)
	case !runtime.BinaryVersionCaptured || !runtime.BinaryHashCaptured:
		return inconclusive(ReasonRuntimeIdentityMissing)
	case !runtime.IsolatedRuntime:
		return inconclusive(ReasonRuntimeIsolationUnproven)
	case !runtime.CleanupSucceeded:
		return inconclusive(ReasonCleanupFailed)
	default:
		return Classification{Result: ResultEvidenceCandidate, PrimaryReason: ReasonEvidenceCandidate}
	}
}

func failed(reason ReasonCode) Classification {
	return Classification{Result: ResultFail, PrimaryReason: reason}
}
func inconclusive(reason ReasonCode) Classification {
	return Classification{Result: ResultInconclusive, PrimaryReason: reason}
}
