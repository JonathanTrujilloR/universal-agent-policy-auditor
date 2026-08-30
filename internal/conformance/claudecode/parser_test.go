package claudecode

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const parserMatcher = "bash-deny-overlapping-allow"

func TestParseEventsFixturesClassifyFailClosed(t *testing.T) {
	for _, tt := range []struct {
		name, file string
		result     Result
		reason     ReasonCode
	}{
		{"malformed JSONL", "events_malformed.jsonl", ResultInconclusive, ReasonParserMalformedEvent},
		{"unsupported event", "events_unsupported.jsonl", ResultInconclusive, ReasonUnsupportedStructuredEvent},
		{"matcher mismatch", "events_matcher_mismatch.jsonl", ResultFail, ReasonMatcherCorrelationFailed},
		{"duplicate target", "events_duplicate.jsonl", ResultInconclusive, ReasonDuplicateTargetEvent},
		{"contradictory success", "events_contradictory_success.jsonl", ResultFail, ReasonContradictoryExecutionSuccess},
		{"display-only denial", "events_display_only.jsonl", ResultInconclusive, ReasonDisplayOnlyEvent},
		{"denial without tool use", "events_denial_without_tool_use.jsonl", ResultInconclusive, ReasonDenialWithoutToolUse},
		{"missing denial", "events_missing_denial.jsonl", ResultFail, ReasonMissingDenial},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(authorizedParserInput(parseFixture(t, tt.file, DefaultParserBounds())))
			if got.Result != tt.result || got.PrimaryReason != tt.reason {
				t.Fatalf("Classify(%s) = %+v, want %s/%s", tt.file, got, tt.result, tt.reason)
			}
		})
	}
}

func TestParseEventsMatchesToolUseAndDenial(t *testing.T) {
	parsed := parseFixture(t, "events_matching.jsonl", DefaultParserBounds())
	input := authorizedParserInput(parsed)
	input.Provenance = ProvenanceFixture
	got, facts := Classify(input), parsed.Input.Facts
	if got.PrimaryReason != ReasonFixtureNotCandidateEvidence || !facts.TargetToolUseObserved || !facts.CorrelatedDenialObserved || !facts.MatcherMatched || !facts.ToolUseIDMatched || parsed.Counts.ToolUses != 1 || parsed.Counts.Denials != 1 || len(parsed.Diagnostics) != 0 {
		t.Fatalf("classification/facts/counts/diagnostics = %+v/%+v/%+v/%+v, want exact fixture correlation", got, facts, parsed.Counts, parsed.Diagnostics)
	}
}

func TestParseEventsRequiresExactMatcherAndStableToolUseID(t *testing.T) {
	for _, tt := range []struct {
		name, json string
		reason     ReasonCode
	}{
		{"different matcher", tool("e1", "tu-1", parserMatcher) + deny("e2", "tu-1", "bash-deny-other"), ReasonMatcherCorrelationFailed},
		{"different tool-use id", tool("e1", "tu-1", parserMatcher) + deny("e2", "tu-2", parserMatcher), ReasonToolUseIDCorrelationFailed},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parsed, got := parseAndClassify(tt.json)
			if got.PrimaryReason != tt.reason {
				t.Fatalf("reason = %s, want %s; facts=%+v", got.PrimaryReason, tt.reason, parsed.Input.Facts)
			}
		})
	}
}

func TestParseEventsRejectsRequiredFieldsAndUnknownStatusSafely(t *testing.T) {
	for _, tt := range []struct {
		name, json, field string
		reason            ReasonCode
	}{
		{"empty tool-use id", `{"event_id":"e1","type":"tool_use","tool":"Bash","tool_use_id":"","matcher_template":"bash-deny-overlapping-allow"}` + "\n", "tool_use_id", ReasonParserMalformedEvent},
		{"missing denial matcher", `{"event_id":"e1","type":"permission_denial","tool_use_id":"tu-secret"}` + "\n", "matcher_template", ReasonParserMalformedEvent},
		{"non-string denial id", `{"event_id":"e1","type":"permission_denial","tool_use_id":7,"matcher_template":"bash-deny-overlapping-allow"}` + "\n", "tool_use_id", ReasonParserMalformedEvent},
		{"missing result status", `{"event_id":"e1","type":"tool_result","tool_use_id":"tu-secret"}` + "\n", "status", ReasonParserMalformedEvent},
		{"missing event id", `{"type":"display","message":"permission denied"}` + "\n", "event_id", ReasonParserMalformedEvent},
		{"unknown result status", matchingEvents("tu-secret") + resultEvent("e3", "tu-secret", "blocked_by_policy"), "status", ReasonUnsupportedStructuredEvent},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parsed, got := parseAndClassify(tt.json)
			if got.PrimaryReason != tt.reason || len(parsed.Diagnostics) != 1 || parsed.Diagnostics[0].Field != tt.field {
				t.Fatalf("got classification=%+v diagnostics=%+v, want %s %s", got, parsed.Diagnostics, tt.reason, tt.field)
			}
			assertNoDiagnosticLeak(t, parsed, "tu-secret", "blocked_by_policy")
		})
	}
}

func TestParseEventsDetectsDistinctDuplicateDenialsAndResults(t *testing.T) {
	for _, tt := range []struct{ name, json string }{
		{"duplicate denial", matchingEvents("tu-secret") + deny("e3", "tu-secret", parserMatcher)},
		{"contradictory duplicate result", matchingEvents("tu-secret") + resultEvent("e3", "tu-secret", "denied") + resultEvent("e4", "tu-secret", "success")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parsed, got := parseAndClassify(tt.json)
			if got.PrimaryReason != ReasonDuplicateTargetEvent || !parsed.Input.Facts.DuplicateTargetEvent {
				t.Fatalf("got classification=%+v facts=%+v, want duplicate ambiguity", got, parsed.Input.Facts)
			}
			assertNoDiagnosticLeak(t, parsed, "tu-secret")
		})
	}
}

func TestParseEventsHonorsInputLineEventAndDiagnosticBounds(t *testing.T) {
	for _, tt := range []struct {
		name, json string
		bounds     ParserBounds
	}{
		{"input", strings.Repeat("x", 9), ParserBounds{MaxInputBytes: 8, MaxLineBytes: 100, MaxEvents: 10, MaxDiagnostics: 2}},
		{"line", display("e1"), ParserBounds{MaxInputBytes: 100, MaxLineBytes: 8, MaxEvents: 10, MaxDiagnostics: 2}},
		{"event", display("e1") + display("e2"), ParserBounds{MaxInputBytes: 100, MaxLineBytes: 100, MaxEvents: 1, MaxDiagnostics: 2}},
		{"diagnostic", "{bad}\n{bad}\n{bad}\n", ParserBounds{MaxInputBytes: 100, MaxLineBytes: 100, MaxEvents: 10, MaxDiagnostics: 1}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParseEventsJSONL([]byte(tt.json), parserMatcher, tt.bounds)
			if parsed.Input.Parser.BoundsHeld || len(parsed.Diagnostics) > tt.bounds.MaxDiagnostics {
				t.Fatalf("bounds/diagnostics = %t/%+v, want bounded failure", parsed.Input.Parser.BoundsHeld, parsed.Diagnostics)
			}
		})
	}
}

func TestParseEventsDiagnosticsArePanicSafeAndNeverRaw(t *testing.T) {
	parsed := ParseEventsJSONL([]byte("{bad /home/alice SECRET_TOKEN=abc}\n"+`{"type":{"bad":"shape"},"matcher_template":"/repo/private","tool_use_id":"secret-id"}`+"\n"), parserMatcher, DefaultParserBounds())
	assertNoDiagnosticLeak(t, parsed, "/home/alice", "SECRET_TOKEN", "/repo/private", "secret-id")
	if len(parsed.Diagnostics) == 0 || parsed.Diagnostics[0].Line == 0 || parsed.Diagnostics[0].Field == "" {
		t.Fatalf("diagnostics = %+v, want bounded line/field identifiers", parsed.Diagnostics)
	}
}
func parseFixture(t *testing.T, name string, bounds ParserBounds) ParseResult {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return ParseEventsJSONL(data, parserMatcher, bounds)
}
func parseAndClassify(json string) (ParseResult, Classification) {
	parsed := ParseEventsJSONL([]byte(json), parserMatcher, DefaultParserBounds())
	return parsed, Classify(authorizedParserInput(parsed))
}
func assertNoDiagnosticLeak(t *testing.T, parsed ParseResult, values ...string) {
	t.Helper()
	diagnostics := fmt.Sprint(parsed.Diagnostics)
	for _, value := range values {
		if strings.Contains(diagnostics, value) {
			t.Fatalf("diagnostics leaked raw value: %+v", parsed.Diagnostics)
		}
	}
}
func matchingEvents(id string) string {
	return tool("e1", id, parserMatcher) + deny("e2", id, parserMatcher)
}
func tool(eventID, id, matcher string) string {
	return `{"event_id":"` + eventID + `","type":"tool_use","tool":"Bash","tool_use_id":"` + id + `","matcher_template":"` + matcher + `"}` + "\n"
}
func deny(eventID, id, matcher string) string {
	return `{"event_id":"` + eventID + `","type":"permission_denial","tool_use_id":"` + id + `","matcher_template":"` + matcher + `"}` + "\n"
}
func resultEvent(eventID, id, status string) string {
	return `{"event_id":"` + eventID + `","type":"tool_result","tool_use_id":"` + id + `","status":"` + status + `"}` + "\n"
}
func display(eventID string) string { return `{"event_id":"` + eventID + `","type":"display"}` + "\n" }
func authorizedParserInput(parsed ParseResult) ClassificationInput {
	input := candidateInput(ProvenanceAuthorizedExecution)
	input.Facts, input.Parser = parsed.Input.Facts, parsed.Input.Parser
	return input
}
