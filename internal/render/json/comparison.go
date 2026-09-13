package json

import (
	"encoding/json"
	"errors"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/compare"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/redact"
)

var ErrInvalidComparisonReport = errors.New("json: invalid comparison report")

type comparisonDocument struct {
	Schema            string                  `json:"schema"`
	Tool              tool                    `json:"tool"`
	Result            redact.Category         `json:"result"`
	ExitCode          int                     `json:"exit_code"`
	Target            target                  `json:"target"`
	Completeness      redact.Completeness     `json:"completeness"`
	ComparisonStatus  redact.ComparisonStatus `json:"comparison_status"`
	ComparisonSources []comparisonSource      `json:"comparison_sources"`
	Reasons           []comparisonReason      `json:"reasons"`
	Differences       []comparisonDifference  `json:"differences"`
	Findings          []finding               `json:"findings"`
	Limitations       []redact.Limitation     `json:"limitations"`
}

type comparisonSource struct {
	Side   compare.Side  `json:"side"`
	Digest redact.Digest `json:"digest"`
}

type comparisonReason struct {
	Side compare.Side       `json:"side"`
	Code compare.ReasonCode `json:"code"`
}

type comparisonDifference struct {
	Capability redact.Capability `json:"capability"`
	Scope      redact.Scope      `json:"scope"`
	Matcher    redact.Matcher    `json:"matcher"`
	OnlyIn     compare.Side      `json:"only_in"`
	Outcome    compare.Outcome   `json:"outcome"`
	Reason     comparisonReason  `json:"reason"`
}

// MarshalComparison renders only an opaque validated comparison report.
// It preserves accessor order and returns a complete buffer with one final newline.
func MarshalComparison(report redact.ComparisonReport) ([]byte, error) {
	if !report.Valid() {
		return nil, ErrInvalidComparisonReport
	}
	doc := comparisonDocument{
		Schema: "auditor-comparison-report/v1alpha1",
		Tool:   tool{Name: "universal-agent-policy-auditor", Version: report.ToolVersion()},
		Result: report.Category(), ExitCode: report.ExitCode(),
		Target:       target{Name: report.Target(), RequestedVersion: report.RequestedVersion(), SupportStatus: report.SupportStatus()},
		Completeness: report.Completeness(), ComparisonStatus: report.Status(),
		ComparisonSources: make([]comparisonSource, 0), Reasons: make([]comparisonReason, 0),
		Differences: make([]comparisonDifference, 0), Findings: make([]finding, 0),
		Limitations: report.Limitations(),
	}
	for _, source := range report.ComparisonSources() {
		doc.ComparisonSources = append(doc.ComparisonSources, comparisonSource{Side: source.Side(), Digest: source.Digest()})
	}
	for _, reason := range report.Reasons() {
		doc.Reasons = append(doc.Reasons, comparisonReason{Side: reason.Side(), Code: reason.Code()})
	}
	for _, difference := range report.Differences() {
		reason := difference.Reason()
		doc.Differences = append(doc.Differences, comparisonDifference{
			Capability: difference.Capability(), Scope: difference.Scope(), Matcher: difference.Matcher(),
			OnlyIn: difference.OnlyIn(), Outcome: difference.Outcome(),
			Reason: comparisonReason{Side: reason.Side(), Code: reason.Code()},
		})
	}
	for _, f := range report.Findings() {
		doc.Findings = append(doc.Findings, finding{Code: f.Code()})
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, ErrInvalidComparisonReport
	}
	return append(data, '\n'), nil
}
