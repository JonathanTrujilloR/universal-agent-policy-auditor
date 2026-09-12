// Package text renders closed safe reports as deterministic plain text.
package text

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/redact"
)

var ErrInvalidReport = errors.New("text: invalid report")

// Render returns a complete report or no bytes for an invalid report.
func Render(report redact.Report) ([]byte, error) {
	if !report.Valid() {
		return nil, ErrInvalidReport
	}
	var out strings.Builder
	line := func(label, value string) {
		out.WriteString(label)
		out.WriteByte(':')
		if value != "" {
			out.WriteByte(' ')
			out.WriteString(value)
		}
		out.WriteByte('\n')
	}
	tool := "universal-agent-policy-auditor"
	if version := report.ToolVersion(); version != "" {
		tool += " " + version
	}
	line("Tool", tool)
	line("Result", fmt.Sprintf("%s (exit %d)", report.Category(), report.ExitCode()))
	if findings := report.Findings(); len(findings) > 0 {
		line("Findings", "")
		for _, finding := range findings {
			fmt.Fprintf(&out, "- %s\n", finding.Code())
		}
	}
	line("Target", string(report.Target()))
	if version := report.RequestedVersion(); version != "" {
		line("Requested version", version)
	}
	line("Support status", string(report.SupportStatus()))
	line("Completeness", string(report.Completeness()))
	if sources := report.SourceDigests(); len(sources) > 0 {
		line("Source digests", "")
		for _, digest := range sources {
			fmt.Fprintf(&out, "- %s\n", digest)
		}
	}
	if permissions := report.Permissions(); len(permissions) > 0 {
		line("Permissions", "")
		for _, permission := range permissions {
			digests := permission.ProvenanceDigests()
			provenance := make([]string, len(digests))
			for i, digest := range digests {
				provenance[i] = string(digest)
			}
			fmt.Fprintf(&out, "- capability=%s; effect=%s; scope=%s; matcher=%s; provenance=[%s]\n",
				permission.Capability(), permission.Effect(), permission.Scope(), permission.Matcher(), strings.Join(provenance, ", "))
		}
	}
	if limitations := report.Limitations(); len(limitations) > 0 {
		line("Limitations", "")
		for _, limitation := range limitations {
			fmt.Fprintf(&out, "- %s\n", limitation)
		}
	}
	return []byte(out.String()), nil
}
