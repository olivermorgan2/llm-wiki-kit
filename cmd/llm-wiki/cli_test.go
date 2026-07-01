package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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

// writeFixtureDir creates a temp directory holding the given filename→content
// pages and returns its path, so validate can run over a controlled wiki.
func writeFixtureDir(t *testing.T, pages map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range pages {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}
	return dir
}

const validFixturePage = "---\ntype: concept\ntitle: Alpha\ndescription: A page.\ntimestamp: 2026-01-01\ntags: [x]\naliases: [y]\nresource: r\n---\n# Alpha\n"

// A complete, conformant page validates clean: success, exit 0, no findings.
func TestValidateJSONCleanPageIsSuccess(t *testing.T) {
	dir := writeFixtureDir(t, map[string]string{"alpha-page.md": validFixturePage})
	stdout, _, code := exec(t, "validate", dir, "--json")

	env := decodeEnvelope(t, stdout)
	if env.Operation != "validate" {
		t.Errorf("operation = %q, want validate", env.Operation)
	}
	if env.Status != contract.StatusSuccess {
		t.Errorf("status = %q, want success", env.Status)
	}
	if env.Findings == nil || len(env.Findings) != 0 {
		t.Errorf("clean page must report zero findings, got %v", env.Findings)
	}
	if code != int(contract.ExitSuccess) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSuccess))
	}
}

// Malformed YAML fails validation unconditionally: validation-failure, exit 2,
// tagged as the OKF parse error (criterion 7).
func TestValidateJSONMalformedYAMLIsValidationFailure(t *testing.T) {
	dir := writeFixtureDir(t, map[string]string{"bad.md": "---\ntype: concept\ntitle: {broken\n---\n"})
	stdout, _, code := exec(t, "validate", dir, "--json")

	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusValidationFailure {
		t.Errorf("status = %q, want validation-failure", env.Status)
	}
	if code != int(contract.ExitValidationFailure) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitValidationFailure))
	}
	var sawParse bool
	for _, f := range env.Findings {
		if f.Ruleset == contract.RulesetOKF && f.Code == "okf-yaml-parse" {
			sawParse = true
		}
	}
	if !sawParse {
		t.Errorf("expected an OKF okf-yaml-parse finding, got %+v", env.Findings)
	}
}

// A page whose only issue is a non-kebab filename is a warning: the run reports
// success-with-warnings and exit 1 (warnings do not fail the gate).
func TestValidateJSONWarningOnlyIsSuccessWithWarnings(t *testing.T) {
	// Complete frontmatter (no errors, no missing recommended fields) but a
	// non-kebab filename → a single warning.
	dir := writeFixtureDir(t, map[string]string{"Not_Kebab.md": validFixturePage})
	stdout, _, code := exec(t, "validate", dir, "--json")

	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusSuccessWithWarnings {
		t.Fatalf("status = %q, want success-with-warnings (%+v)", env.Status, env.Findings)
	}
	if code != int(contract.ExitSuccessWithWarnings) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSuccessWithWarnings))
	}
}

func TestValidateIsHumanReadableByDefault(t *testing.T) {
	dir := writeFixtureDir(t, map[string]string{"alpha-page.md": validFixturePage})
	stdout, _, code := exec(t, "validate", dir)

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

// A bare invocation under JSON mode must still receive the v1 envelope on
// stdout, not human usage on stderr — machine surfaces contract on the
// envelope even for invalid invocation (ADR-003, #1).
func TestNoCommandJSONFlagEmitsInvalidInvocationEnvelope(t *testing.T) {
	stdout, stderr, code := exec(t, "--json")

	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusInvalidInvocation {
		t.Errorf("status = %q, want %q", env.Status, contract.StatusInvalidInvocation)
	}
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitInvalidInvocation))
	}
	if stderr != "" {
		t.Errorf("JSON-mode error must not print human usage to stderr: %s", stderr)
	}
}

// The LLM_WIKI_JSON env toggle selects the envelope for a bare invocation just
// like the flag does.
func TestNoCommandJSONEnvEmitsInvalidInvocationEnvelope(t *testing.T) {
	t.Setenv("LLM_WIKI_JSON", "1")
	stdout, stderr, code := exec(t)

	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusInvalidInvocation {
		t.Errorf("status = %q, want %q", env.Status, contract.StatusInvalidInvocation)
	}
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitInvalidInvocation))
	}
	if stderr != "" {
		t.Errorf("JSON-mode error must not print human usage to stderr: %s", stderr)
	}
}
