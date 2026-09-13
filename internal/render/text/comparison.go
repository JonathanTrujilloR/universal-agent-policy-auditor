package text

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/redact"
)

var ErrInvalidComparisonReport = errors.New("text: invalid comparison report")

// RenderComparison returns complete owned bytes from an opaque validated report.
// Blockers precede target metadata; accessor order is preserved without ambient inputs.
func RenderComparison(report redact.ComparisonReport) ([]byte, error) {
	if !report.Valid() {
		return nil, ErrInvalidComparisonReport
	}
	var out bytes.Buffer
	out.WriteString("Tool: universal-agent-policy-auditor")
	if version := report.ToolVersion(); version != "" {
		fmt.Fprintf(&out, " %s", version)
	}
	fmt.Fprintf(&out, "\nResult: %s (exit %d)\n", report.Category(), report.ExitCode())
	fmt.Fprintf(&out, "Comparison status: %s\n", report.Status())
	if reasons := report.Reasons(); len(reasons) > 0 {
		out.WriteString("Reasons:\n")
		for _, reason := range reasons {
			fmt.Fprintf(&out, "- side=%s; code=%s\n", reason.Side(), reason.Code())
		}
	}
	if differences := report.Differences(); len(differences) > 0 {
		out.WriteString("Differences:\n")
		for _, difference := range differences {
			reason := difference.Reason()
			fmt.Fprintf(&out, "- capability=%s; scope=%s; matcher=%s; only_in=%s; outcome=%s; reason_side=%s; reason_code=%s\n",
				difference.Capability(), difference.Scope(), difference.Matcher(), difference.OnlyIn(),
				difference.Outcome(), reason.Side(), reason.Code())
		}
	}
	if findings := report.Findings(); len(findings) > 0 {
		out.WriteString("Findings:\n")
		for _, finding := range findings {
			fmt.Fprintf(&out, "- %s\n", finding.Code())
		}
	}
	fmt.Fprintf(&out, "Target: %s\n", report.Target())
	if version := report.RequestedVersion(); version != "" {
		fmt.Fprintf(&out, "Requested version: %s\n", version)
	}
	fmt.Fprintf(&out, "Support status: %s\n", report.SupportStatus())
	fmt.Fprintf(&out, "Completeness: %s\n", report.Completeness())
	if sources := report.ComparisonSources(); len(sources) > 0 {
		out.WriteString("Comparison sources:\n")
		for _, source := range sources {
			fmt.Fprintf(&out, "- side=%s; digest=%s\n", source.Side(), source.Digest())
		}
	}
	if limitations := report.Limitations(); len(limitations) > 0 {
		out.WriteString("Limitations:\n")
		for _, limitation := range limitations {
			fmt.Fprintf(&out, "- %s\n", limitation)
		}
	}
	return out.Bytes(), nil
}
