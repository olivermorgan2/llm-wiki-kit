package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

// exec runs the CLI with the given args and captures stdout, stderr, and the
// exit code.
func exec(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var out, errb bytes.Buffer
	code = run(args, &out, &errb)
	return out.String(), errb.String(), code
}

// decodeEnvelope parses stdout as a contract envelope and asserts it carries
// exactly the six ADR-003 fields.
func decodeEnvelope(t *testing.T, stdout string) contract.Envelope {
	t.Helper()
	var generic map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &generic); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout)
	}
	for _, field := range []string{"contractVersion", "operation", "status", "findings", "affectedPaths", "approval"} {
		if _, ok := generic[field]; !ok {
			t.Errorf("envelope missing field %q: %s", field, stdout)
		}
	}
	if len(generic) != 6 {
		t.Errorf("envelope must carry exactly 6 fields, got %d: %s", len(generic), stdout)
	}
	var env contract.Envelope
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return env
}

func TestVersionIsHumanReadableByDefault(t *testing.T) {
	stdout, stderr, code := exec(t, "version")

	if code != int(contract.ExitSuccess) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSuccess))
	}
	if strings.Contains(stdout, "{") {
		t.Errorf("default output must be human-readable, not JSON: %s", stdout)
	}
	if !strings.Contains(stdout, "llm-wiki") {
		t.Errorf("version output should name the tool: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("no stderr expected on success, got: %s", stderr)
	}
}

func TestVersionJSONEmitsSuccessEnvelope(t *testing.T) {
	stdout, _, code := exec(t, "version", "--json")

	env := decodeEnvelope(t, stdout)
	if env.ContractVersion != "v1" {
		t.Errorf("contractVersion = %q, want v1", env.ContractVersion)
	}
	if env.Operation != "version" {
		t.Errorf("operation = %q, want version", env.Operation)
	}
	if env.Status != contract.StatusSuccess {
		t.Errorf("status = %q, want %q", env.Status, contract.StatusSuccess)
	}
	if code != int(contract.ExitSuccess) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSuccess))
	}
}

func TestValidateJSONEmitsSuccessEnvelope(t *testing.T) {
	stdout, _, code := exec(t, "validate", "--json")

	env := decodeEnvelope(t, stdout)
	if env.Operation != "validate" {
		t.Errorf("operation = %q, want validate", env.Operation)
	}
	if env.Status != contract.StatusSuccess {
		t.Errorf("status = %q, want success (no-op skeleton)", env.Status)
	}
	if env.Findings == nil || len(env.Findings) != 0 {
		t.Errorf("no-op validate must report zero findings, got %v", env.Findings)
	}
	if code != int(contract.ExitSuccess) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSuccess))
	}
}

func TestValidateIsHumanReadableByDefault(t *testing.T) {
	stdout, _, code := exec(t, "validate")

	if strings.Contains(stdout, "{") {
		t.Errorf("default output must be human-readable, not JSON: %s", stdout)
	}
	if code != int(contract.ExitSuccess) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSuccess))
	}
}

// The LLM_WIKI_JSON env toggle selects JSON output just like the flag, so
// machine surfaces can opt in without editing the command line.
func TestJSONSelectedViaEnvVar(t *testing.T) {
	t.Setenv("LLM_WIKI_JSON", "1")
	stdout, _, code := exec(t, "version")

	env := decodeEnvelope(t, stdout)
	if env.Operation != "version" {
		t.Errorf("operation = %q, want version", env.Operation)
	}
	if code != int(contract.ExitSuccess) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSuccess))
	}
}

// The global --json flag is honored whether it appears before or after the
// subcommand — parsing must be deterministic regardless of position.
func TestJSONFlagPositionIndependent(t *testing.T) {
	before, _, _ := exec(t, "--json", "version")
	after, _, _ := exec(t, "version", "--json")

	if !strings.Contains(before, "{") {
		t.Errorf("--json before subcommand should emit JSON: %s", before)
	}
	if before != after {
		t.Errorf("--json position changed output:\n before: %s\n after: %s", before, after)
	}
}

func TestUnknownCommandIsInvalidInvocation(t *testing.T) {
	stdout, stderr, code := exec(t, "frobnicate")

	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitInvalidInvocation))
	}
	if stderr == "" {
		t.Error("an unknown command should explain the error on stderr")
	}
	if strings.Contains(stdout, "{") {
		t.Errorf("human-mode error should not print an envelope to stdout: %s", stdout)
	}
}

func TestUnknownCommandJSONEmitsInvalidInvocationEnvelope(t *testing.T) {
	stdout, _, code := exec(t, "frobnicate", "--json")

	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusInvalidInvocation {
		t.Errorf("status = %q, want %q", env.Status, contract.StatusInvalidInvocation)
	}
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitInvalidInvocation))
	}
}

func TestNoCommandIsInvalidInvocation(t *testing.T) {
	_, stderr, code := exec(t)

	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitInvalidInvocation))
	}
	if stderr == "" {
		t.Error("a bare invocation should print usage to stderr")
	}
}
