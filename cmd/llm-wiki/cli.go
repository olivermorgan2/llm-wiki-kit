package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/platform"
	"github.com/olivermorgan2/llm-wiki-kit/internal/validate"
)

// version is the llm-wiki build version. It is a skeleton placeholder until the
// release pipeline stamps it.
const version = "0.1.0-dev"

const usage = `llm-wiki — deterministic engine for portable knowledge bundles

Usage:
  llm-wiki <command> [--json] [args]

Commands:
  version     Print the engine version.
  validate    Validate repository contents (no rules implemented yet).
  selfcheck   Verify this platform's shipped binary against its bundled
              checksum (ADR-002). Use --root <dir> to point at the bundle;
              defaults to the directory containing the running executable.

Flags:
  --json      Emit the versioned JSON contract envelope instead of human text.
              Equivalent to setting LLM_WIKI_JSON=1.
`

// run is the testable CLI entry point. It parses args, dispatches to a command,
// and returns the process exit code. Output is written to stdout/stderr rather
// than the real streams so it can be captured in tests.
func run(args []string, stdout, stderr io.Writer) int {
	jsonMode := jsonEnabled()
	var positional []string
	for _, a := range args {
		switch a {
		case "--json", "-json":
			jsonMode = true
		case "-h", "--help", "help":
			fmt.Fprint(stdout, usage)
			return int(contract.ExitSuccess)
		default:
			positional = append(positional, a)
		}
	}

	if len(positional) == 0 {
		if jsonMode {
			env := contract.New("", contract.StatusInvalidInvocation)
			return emit(stdout, env)
		}
		fmt.Fprint(stderr, usage)
		return int(contract.ExitInvalidInvocation)
	}

	command, cmdArgs := positional[0], positional[1:]
	switch command {
	case "version":
		return runVersion(stdout, jsonMode)
	case "validate":
		return runValidate(cmdArgs, stdout, jsonMode)
	case "selfcheck":
		return runSelfcheck(cmdArgs, stdout, stderr, jsonMode)
	default:
		return runUnknown(command, stdout, stderr, jsonMode)
	}
}

// jsonEnabled reports whether the LLM_WIKI_JSON env toggle requests JSON output.
func jsonEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LLM_WIKI_JSON"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func runVersion(stdout io.Writer, jsonMode bool) int {
	env := contract.New("version", contract.StatusSuccess)
	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stdout, "llm-wiki %s (contract %s)\n", version, contract.ContractVersion)
	return int(contract.ExitSuccess)
}

func runValidate(paths []string, stdout io.Writer, jsonMode bool) int {
	findings := validate.New(nil).Run(paths)
	env := contract.New("validate", contract.StatusSuccess)
	env.Findings = findings

	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stdout, "validate: OK — %d finding(s) (no validation rules implemented yet)\n", len(findings))
	return int(contract.ExitCodeForStatus(env.Status))
}

// runSelfcheck detects the running platform and verifies its shipped binary
// against the bundled SHA256SUMS manifest (ADR-002). It fails closed: any
// verification error maps to the contract's system-or-filesystem-failure bucket
// (exit 5). Consistent with the other commands, the outcome is carried by the
// envelope status rather than a validation finding — findings are OKF/profile
// results (ADR-004), not integrity failures.
func runSelfcheck(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	root, err := selfcheckRoot(args)
	if err == nil {
		var p platform.Platform
		p, err = platform.VerifyBundle(os.DirFS(root))
		if err == nil {
			return selfcheckOK(stdout, jsonMode, p.ArtifactPath())
		}
	}
	return selfcheckFail(stdout, stderr, jsonMode, err)
}

// selfcheckRoot resolves the bundle root: an explicit --root/--root=<dir> flag,
// otherwise the directory containing the running executable (where the plugin
// ships bin/ alongside the engine).
func selfcheckRoot(args []string) (string, error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--root" || a == "-root":
			if i+1 >= len(args) {
				return "", fmt.Errorf("--root requires a directory argument")
			}
			return args[i+1], nil
		case strings.HasPrefix(a, "--root="):
			return strings.TrimPrefix(a, "--root="), nil
		case strings.HasPrefix(a, "-root="):
			return strings.TrimPrefix(a, "-root="), nil
		}
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate executable: %w", err)
	}
	return filepath.Dir(exe), nil
}

func selfcheckOK(stdout io.Writer, jsonMode bool, artifact string) int {
	env := contract.New("selfcheck", contract.StatusSuccess)
	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stdout, "selfcheck: OK — verified %s\n", artifact)
	return int(contract.ExitCodeForStatus(env.Status))
}

func selfcheckFail(stdout, stderr io.Writer, jsonMode bool, cause error) int {
	env := contract.New("selfcheck", contract.StatusSystemFailure)
	if jsonMode {
		return emit(stdout, env)
	}
	fmt.Fprintf(stderr, "llm-wiki: selfcheck failed: %v\n", cause)
	return int(contract.ExitCodeForStatus(env.Status))
}

func runUnknown(command string, stdout, stderr io.Writer, jsonMode bool) int {
	if jsonMode {
		env := contract.New(command, contract.StatusInvalidInvocation)
		return emit(stdout, env)
	}
	fmt.Fprintf(stderr, "llm-wiki: unknown command %q\n\n%s", command, usage)
	return int(contract.ExitInvalidInvocation)
}

// emit writes the envelope as JSON and returns the exit code its status maps to.
func emit(stdout io.Writer, env *contract.Envelope) int {
	if err := env.WriteJSON(stdout); err != nil {
		return int(contract.ExitSystemFailure)
	}
	return int(contract.ExitCodeForStatus(env.Status))
}
