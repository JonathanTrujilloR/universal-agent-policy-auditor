package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/redact"
	jsonreport "github.com/jkelevra/universal-agent-policy-auditor/internal/render/json"
	"github.com/jkelevra/universal-agent-policy-auditor/internal/render/text"
)

var version = "dev"

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

type dependencies struct {
	run        func(app.Request) app.Result
	redact     func(app.Result) (redact.Report, error)
	text, json func(redact.Report) ([]byte, error)
}

func run(args []string, stdout, stderr io.Writer) int {
	return runWith(args, stdout, stderr, dependencies{app.Run, redact.NewReport, text.Render, jsonreport.Marshal})
}

func runWith(args []string, stdout, stderr io.Writer, deps dependencies) int {
	if len(args) == 0 || len(args) == 1 && isHelp(args[0]) {
		if err := writeExact(stdout, []byte(helpText())); err != nil {
			return app.ExitCode(app.OperationalFailure)
		}
		return 0
	}
	var result app.Result
	var format string
	if isHelp(args[0]) {
		result = app.Result{Category: app.InvalidRequest, Message: "invalid help request"}
	} else {
		switch args[0] {
		case "version":
			if len(args) != 1 {
				result = app.Result{Category: app.InvalidRequest, Message: "invalid version request"}
			} else {
				result = deps.run(app.Request{Mode: app.ModeVersion, Build: buildInfo()})
			}
		case "audit":
			var req app.Request
			req, format = parseOutputAudit(args[1:])
			if req.Mode != app.ModeAudit {
				result = app.Result{Category: app.InvalidRequest}
			} else {
				result = deps.run(req)
			}
		default:
			result = app.Result{Category: app.InvalidRequest, Message: "invalid request"}
		}
	}
	if format != "" {
		report, err := deps.redact(result)
		var data []byte
		if err == nil {
			render := deps.text
			if format == "json" {
				render = deps.json
			}
			data, err = render(report)
		}
		if err != nil || len(data) == 0 {
			_ = writeExact(stderr, []byte("operational_failure\n"))
			return 4
		}
		if writeExact(stdout, data) != nil {
			return 4
		}
		return app.ExitCode(result.Category)
	}
	if writeResult(result, stdout, stderr) != nil {
		return app.ExitCode(app.OperationalFailure)
	}
	return app.ExitCode(result.Category)
}

func parseOutputAudit(args []string) (app.Request, string) {
	invalid := app.Request{Mode: "invalid", Build: buildInfo()}
	end, format := len(args), ""
	noColor := end > 0 && args[end-1] == "--no-color"
	if noColor {
		end--
	}
	if end >= 2 && args[end-2] == "--format" {
		format = args[end-1]
		end -= 2
		if format != "text" && format != "json" {
			return invalid, ""
		}
	}
	if noColor && format != "text" {
		return invalid, ""
	}
	req := parseAudit(args[:end])
	for i := 2; format != "" && i < end; i += 2 {
		if args[i] == "" || strings.HasPrefix(args[i], "--") {
			return invalid, ""
		}
	}
	if req.Mode != app.ModeAudit {
		return invalid, ""
	}
	return req, format
}

func parseAudit(args []string) app.Request {
	req := app.Request{Mode: app.ModeAudit, Build: buildInfo()}
	if len(args) == 0 {
		return app.Request{Mode: "invalid", Build: buildInfo()}
	}
	req.Target = app.Target(args[0])
	if req.Target != app.TargetOpenCode && req.Target != app.TargetClaudeCode {
		return app.Request{Mode: "invalid", Build: buildInfo()}
	}
	if (len(args) != 5 && len(args) != 7) || args[1] != "--root" || args[3] != "--config" {
		return app.Request{Mode: "invalid", Build: buildInfo()}
	}
	req.Root, req.ExplicitConfig = args[2], args[4]
	if len(args) == 7 {
		if args[5] != "--opencode-version" {
			return app.Request{Mode: "invalid", Build: buildInfo()}
		}
		req.TargetVersion = args[6]
	}
	return req
}

func buildInfo() app.BuildInfo {
	if version == "" {
		version = "dev"
	}
	return app.BuildInfo{Version: version}
}

func writeResult(result app.Result, stdout, stderr io.Writer) error {
	writer := stdout
	line := string(result.Category) + "\n"
	if result.Build.Version != "" && result.Message == "version" {
		line = fmt.Sprintf("auditor %s\n", result.Build.Version)
	}
	if result.Category == app.InvalidRequest || result.Category == app.OperationalFailure {
		writer = stderr
	}
	return writeExact(writer, []byte(line))
}

func writeExact(writer io.Writer, data []byte) error {
	n, err := writer.Write(data)
	if err == nil && n != len(data) {
		return io.ErrShortWrite
	}
	return err
}

func isHelp(arg string) bool { return arg == "help" || arg == "--help" || arg == "-h" }

func helpText() string {
	return "auditor " + buildInfo().Version + "\n" +
		"Usage:\n" +
		"  auditor version\n" +
		"  auditor audit <target> --root <abs> --config <abs> [--opencode-version <v>] [--format text|json] [--no-color]\n" +
		"Exit categories:\n" +
		"  0 complete_no_findings\n" +
		"  1 complete_with_findings\n" +
		"  2 unsupported_or_incomplete\n" +
		"  3 invalid_request\n" +
		"  4 operational_failure\n"
}
