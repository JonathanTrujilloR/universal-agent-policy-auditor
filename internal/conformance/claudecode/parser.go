package claudecode

import (
	"bytes"
	"encoding/json"
)

const (
	defaultMaxJSONLInputBytes = 64 * 1024
	defaultMaxJSONLLineBytes  = 8 * 1024
	defaultMaxJSONLEvents     = 128
	defaultMaxDiagnostics     = 16
)

type ParserBounds struct{ MaxInputBytes, MaxLineBytes, MaxEvents, MaxDiagnostics int }
type ParseCounts struct{ Events, ToolUses, Denials int }
type ParseDiagnostic struct {
	Line   int
	Field  string
	Reason ReasonCode
}
type ParseResult struct {
	Input          ClassificationInput
	Counts         ParseCounts
	Diagnostics    []ParseDiagnostic
	maxDiagnostics int
}

func DefaultParserBounds() ParserBounds {
	return ParserBounds{defaultMaxJSONLInputBytes, defaultMaxJSONLLineBytes, defaultMaxJSONLEvents, defaultMaxDiagnostics}
}

func ParseEventsJSONL(data []byte, matcher string, bounds ParserBounds) (result ParseResult) {
	defer func() {
		if recover() != nil {
			result.Input.Parser.StructuredEventsParsed = true
			result.Input.Parser.NoMalformedEvents = false
			result.addDiagnostic(0, "panic", ReasonParserMalformedEvent)
		}
	}()
	bounds = normalizeParserBounds(bounds)
	result.maxDiagnostics = bounds.MaxDiagnostics
	result.Input = ClassificationInput{Provenance: ProvenanceFixture, Parser: ParserProof{NoMalformedEvents: true, BoundsHeld: true, NoUnsupportedEvents: true}, Redaction: RedactionPassed}
	if len(data) > bounds.MaxInputBytes {
		result.fail(0, "input_bytes", ReasonParserBoundsExceeded)
		return result
	}

	seen := map[string]bool{}
	target, denyMatch, denyOther, results, success := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}
	var display, duplicate, unsupported bool
	for i, line := range bytes.Split(bytes.TrimSuffix(data, []byte("\n")), []byte("\n")) {
		lineNo := i + 1
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		if len(line) > bounds.MaxLineBytes {
			result.fail(lineNo, "line_bytes", ReasonParserBoundsExceeded)
			continue
		}
		result.Counts.Events++
		if result.Counts.Events > bounds.MaxEvents {
			result.fail(lineNo, "event_count", ReasonParserBoundsExceeded)
			continue
		}
		event, ok := decodeParserEvent(line, &result, lineNo)
		if !ok {
			continue
		}
		result.Input.Parser.StructuredEventsParsed = true
		if key := event.key(); seen[key] {
			continue
		} else {
			seen[key] = true
		}
		switch event.Type {
		case "tool_use":
			if event.Tool == "Bash" && event.MatcherTemplate == matcher {
				duplicate = duplicate || len(target) > 0
				target[event.ToolUseID]++
				result.Counts.ToolUses++
			}
		case "permission_denial":
			result.Counts.Denials++
			if event.MatcherTemplate == matcher {
				denyMatch[event.ToolUseID]++
			} else {
				denyOther[event.ToolUseID]++
			}
		case "tool_result":
			results[event.ToolUseID]++
			if event.Status == "success" {
				success[event.ToolUseID]++
			}
		case "display":
			display = true
		case "diagnostic":
		default:
			unsupported = true
			result.fail(lineNo, "type", ReasonUnsupportedStructuredEvent)
		}
	}

	facts := EvidenceFacts{MatcherMatched: true, ToolUseIDMatched: true, DisplayOnlyEvent: display, UnsupportedStructuredEvent: unsupported}
	facts.TargetToolUseObserved = len(target) > 0
	for id := range target {
		duplicate = duplicate || target[id] > 1 || denyMatch[id]+denyOther[id] > 1 || results[id] > 1
		facts.CorrelatedDenialObserved = facts.CorrelatedDenialObserved || denyMatch[id] > 0 || denyOther[id] > 0
		facts.MatcherMatched = facts.MatcherMatched && denyOther[id] == 0
		facts.TargetExecutionSucceeded = facts.TargetExecutionSucceeded || success[id] > 0
	}
	if !facts.TargetToolUseObserved && (len(denyMatch) > 0 || len(denyOther) > 0) {
		facts.CorrelatedDenialObserved = true
	}
	if facts.TargetToolUseObserved && !facts.CorrelatedDenialObserved && len(denyMatch) > 0 {
		facts.CorrelatedDenialObserved, facts.ToolUseIDMatched = true, false
	}
	facts.DuplicateTargetEvent = duplicate
	if duplicate {
		result.addDiagnostic(0, "event_multiplicity", ReasonDuplicateTargetEvent)
	}
	result.Input.Facts = facts
	return result
}

type parserEvent struct{ EventID, Type, Tool, ToolUseID, MatcherTemplate, Status string }

func decodeParserEvent(line []byte, result *ParseResult, lineNo int) (parserEvent, bool) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil {
		result.fail(lineNo, "json", ReasonParserMalformedEvent)
		return parserEvent{}, false
	}
	var event parserEvent
	for _, field := range []struct {
		name   string
		target *string
	}{{"type", &event.Type}, {"event_id", &event.EventID}} {
		if !requiredString(raw, field.name, field.target) {
			result.fail(lineNo, field.name, ReasonParserMalformedEvent)
			return parserEvent{}, false
		}
	}
	required := map[string][]struct {
		name   string
		target *string
	}{
		"tool_use":          {{"tool", &event.Tool}, {"tool_use_id", &event.ToolUseID}, {"matcher_template", &event.MatcherTemplate}},
		"permission_denial": {{"tool_use_id", &event.ToolUseID}, {"matcher_template", &event.MatcherTemplate}},
		"tool_result":       {{"tool_use_id", &event.ToolUseID}, {"status", &event.Status}},
	}
	for _, field := range required[event.Type] {
		if !requiredString(raw, field.name, field.target) {
			result.fail(lineNo, field.name, ReasonParserMalformedEvent)
			return parserEvent{}, false
		}
	}
	if event.Type == "tool_result" && event.Status != "success" && event.Status != "denied" {
		result.fail(lineNo, "status", ReasonUnsupportedStructuredEvent)
		return parserEvent{}, false
	}
	return event, true
}
func requiredString(raw map[string]json.RawMessage, name string, target *string) bool {
	value, ok := raw[name]
	return ok && json.Unmarshal(value, target) == nil && *target != ""
}
func (event parserEvent) key() string {
	return event.EventID + "|" + event.Type + "|" + event.Tool + "|" + event.MatcherTemplate + "|" + event.ToolUseID + "|" + event.Status
}
func normalizeParserBounds(bounds ParserBounds) ParserBounds {
	defaults := DefaultParserBounds()
	return ParserBounds{positive(bounds.MaxInputBytes, defaults.MaxInputBytes), positive(bounds.MaxLineBytes, defaults.MaxLineBytes), positive(bounds.MaxEvents, defaults.MaxEvents), positive(bounds.MaxDiagnostics, defaults.MaxDiagnostics)}
}
func positive(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
func (result *ParseResult) fail(line int, field string, reason ReasonCode) {
	result.Input.Parser.StructuredEventsParsed = true
	switch reason {
	case ReasonParserBoundsExceeded:
		result.Input.Parser.BoundsHeld = false
	case ReasonParserMalformedEvent:
		result.Input.Parser.NoMalformedEvents = false
	case ReasonUnsupportedStructuredEvent:
		result.Input.Parser.NoUnsupportedEvents = false
	}
	result.addDiagnostic(line, field, reason)
}
func (result *ParseResult) addDiagnostic(line int, field string, reason ReasonCode) {
	limit := result.maxDiagnostics
	if limit <= 0 {
		limit = defaultMaxDiagnostics
	}
	if len(result.Diagnostics) >= limit {
		result.Input.Parser.BoundsHeld = false
		return
	}
	result.Diagnostics = append(result.Diagnostics, ParseDiagnostic{Line: line, Field: field, Reason: reason})
}
