package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jkelevra/universal-agent-policy-auditor/internal/app"
)

var version = "dev"

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || len(args) == 1 && isHelp(args[0]) {
		if _, err := fmt.Fprint(stdout, helpText()); err != nil {
			return app.ExitCode(app.OperationalFailure)
		}
		return 0
	}
	var result app.Result
	if isHelp(args[0]) {
		result = app.Result{Category: app.InvalidRequest, Message: "invalid help request"}
	} else {
		switch args[0] {
		case "version":
			if len(args) != 1 {
				result = app.Result{Category: app.InvalidRequest, Message: "invalid version request"}
			} else {
				result = app.Run(app.Request{Mode: app.ModeVersion, Build: buildInfo()})
			}
		case "audit":
			result = app.Run(parseAudit(args[1:]))
		default:
			result = app.Result{Category: app.InvalidRequest, Message: "invalid request"}
		}
	}
	if writeResult(result, stdout, stderr) != nil {
		return app.ExitCode(app.OperationalFailure)
	}
	return app.ExitCode(result.Category)
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
	_, err := fmt.Fprint(writer, line)
	return err
}

func isHelp(arg string) bool { return arg == "help" || arg == "--help" || arg == "-h" }

func helpText() string {
	return "auditor " + buildInfo().Version + "\n" +
		"Usage:\n" +
		"  auditor version\n" +
		"  auditor audit opencode --root <dir> --config <path> --opencode-version <version>\n" +
		"Exit categories:\n" +
		"  0 complete_no_findings\n" +
		"  1 complete_with_findings\n" +
		"  2 unsupported_or_incomplete\n" +
		"  3 invalid_request\n" +
		"  4 operational_failure\n"
}
