// Package json renders valid redacted reports as versioned JSON.
package json

import (
	"encoding/json"
	"errors"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/redact"
)

var ErrInvalidReport = errors.New("json: invalid report")

type document struct {
	Schema        string              `json:"schema"`
	Tool          tool                `json:"tool"`
	Result        redact.Category     `json:"result"`
	ExitCode      int                 `json:"exit_code"`
	Target        target              `json:"target"`
	Completeness  redact.Completeness `json:"completeness"`
	SourceDigests []redact.Digest     `json:"source_digests"`
	Permissions   []permission        `json:"permissions"`
	Findings      []finding           `json:"findings"`
	Limitations   []redact.Limitation `json:"limitations"`
}

type tool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type target struct {
	Name             redact.Target        `json:"name"`
	RequestedVersion string               `json:"requested_version"`
	SupportStatus    redact.SupportStatus `json:"support_status"`
}

type permission struct {
	Capability        redact.Capability `json:"capability"`
	Effect            redact.Effect     `json:"effect"`
	Scope             redact.Scope      `json:"scope"`
	Matcher           redact.Matcher    `json:"matcher"`
	ProvenanceDigests []redact.Digest   `json:"provenance_digests"`
}

type finding struct {
	Code redact.FindingCode `json:"code"`
}

// Marshal returns one complete JSON document with exactly one trailing newline.
// Struct order and encoding/json's default HTML escaping are intentional.
func Marshal(report redact.Report) ([]byte, error) {
	if !report.Valid() {
		return nil, ErrInvalidReport
	}
	doc := document{
		Schema: "auditor-report/v1alpha1",
		Tool:   tool{Name: "universal-agent-policy-auditor", Version: report.ToolVersion()},
		Result: report.Category(), ExitCode: report.ExitCode(),
		Target:       target{Name: report.Target(), RequestedVersion: report.RequestedVersion(), SupportStatus: report.SupportStatus()},
		Completeness: report.Completeness(), SourceDigests: report.SourceDigests(),
		Permissions: make([]permission, 0), Findings: make([]finding, 0),
		Limitations: report.Limitations(),
	}
	for _, p := range report.Permissions() {
		doc.Permissions = append(doc.Permissions, permission{
			Capability: p.Capability(), Effect: p.Effect(), Scope: p.Scope(),
			Matcher: p.Matcher(), ProvenanceDigests: p.ProvenanceDigests(),
		})
	}
	for _, f := range report.Findings() {
		doc.Findings = append(doc.Findings, finding{Code: f.Code()})
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, ErrInvalidReport
	}
	return append(data, '\n'), nil
}
