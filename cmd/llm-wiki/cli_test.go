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
	// The six ADR-003 fields are always present; page-scoped operations add
	// exactly one optional seventh payload field — "page" (inspect) or "plan"
	// (plan) — and nothing else.
	switch len(generic) {
	case 6:
	case 7:
		_, hasPage := generic["page"]
		_, hasPlan := generic["plan"]
		if !hasPage && !hasPlan {
			t.Errorf("7-field envelope's only permitted extra is \"page\" or \"plan\": %s", stdout)
		}
		if hasPage && hasPlan {
			t.Errorf("envelope must not carry both \"page\" and \"plan\": %s", stdout)
		}
	default:
		t.Errorf("envelope must carry 6 fields (or 7 with page/plan), got %d: %s", len(generic), stdout)
	}
	var env contract.Envelope
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return env
}

// initTargets is the exact scaffold the init command materializes, in the
// sorted slash-form order the success envelope reports.
var initTargets = []string{"llm-wiki.yaml", "wiki/index.md", "wiki/templates/page-template.md"}

// AC1 end-to-end: a fresh init produces a bundle that validates with ZERO
// findings, and all three files land on disk.
func TestInitProducesBundleThatValidatesClean(t *testing.T) {
	dir := t.TempDir()

	_, _, code := exec(t, "init", dir, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("init exit = %d, want 0", code)
	}
	for _, rel := range initTargets {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("scaffold file %q missing: %v", rel, err)
		}
	}

	stdout, _, code := exec(t, "validate", dir, "--json")
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusSuccess {
		t.Errorf("validate status = %q, want success (%+v)", env.Status, env.Findings)
	}
	if len(env.Findings) != 0 {
		t.Errorf("scaffolded bundle must validate with zero findings, got %+v", env.Findings)
	}
	if code != int(contract.ExitSuccess) {
		t.Errorf("validate exit = %d, want 0", code)
	}
}

// AC3 success envelope shape: operation init, status success, the three sorted
// affectedPaths, empty findings, nil approval.
func TestInitJSONSuccessEnvelope(t *testing.T) {
	dir := t.TempDir()

	stdout, _, code := exec(t, "init", dir, "--json")
	env := decodeEnvelope(t, stdout)
	if env.Operation != "init" {
		t.Errorf("operation = %q, want init", env.Operation)
	}
	if env.Status != contract.StatusSuccess {
		t.Errorf("status = %q, want success", env.Status)
	}
	if len(env.AffectedPaths) != len(initTargets) {
		t.Fatalf("affectedPaths = %v, want %v", env.AffectedPaths, initTargets)
	}
	for i, want := range initTargets {
		if env.AffectedPaths[i] != want {
			t.Errorf("affectedPaths[%d] = %q, want %q", i, env.AffectedPaths[i], want)
		}
	}
	if len(env.Findings) != 0 {
		t.Errorf("findings = %v, want empty", env.Findings)
	}
	if env.Approval != nil {
		t.Errorf("approval = %+v, want nil on success", env.Approval)
	}
	if code != int(contract.ExitSuccess) {
		t.Errorf("exit = %d, want 0", code)
	}
}

// snapshotBundle records the on-disk bytes of the scaffold files so a refusal
// can be proven non-mutating.
func snapshotBundle(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	snap := map[string][]byte{}
	for _, rel := range initTargets {
		if b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel))); err == nil {
			snap[rel] = b
		}
	}
	return snap
}

// AC2 re-init refusal: a second init without --force refuses with an
// approval-required envelope listing all conflicts and mutates nothing.
func TestInitReinitRefusesWithApprovalEnvelope(t *testing.T) {
	dir := t.TempDir()
	if _, _, code := exec(t, "init", dir, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("first init exit = %d, want 0", code)
	}
	before := snapshotBundle(t, dir)

	stdout, _, code := exec(t, "init", dir, "--json")
	if code != int(contract.ExitApprovalRequired) {
		t.Fatalf("re-init exit = %d, want 3", code)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusApprovalRequired {
		t.Errorf("status = %q, want approval-required", env.Status)
	}
	if env.Approval == nil || !env.Approval.Required {
		t.Fatalf("approval must be present and required, got %+v", env.Approval)
	}
	if len(env.Approval.Paths) != len(initTargets) {
		t.Errorf("approval paths = %v, want all three targets", env.Approval.Paths)
	}
	for i, want := range initTargets {
		if env.Approval.Paths[i] != want {
			t.Errorf("approval path[%d] = %q, want %q", i, env.Approval.Paths[i], want)
		}
	}
	after := snapshotBundle(t, dir)
	for rel, b := range before {
		if !bytes.Equal(after[rel], b) {
			t.Errorf("refusal mutated %q", rel)
		}
	}
}

// AC2 partial conflict: only a user-owned llm-wiki.yaml pre-exists, so the
// refusal lists exactly it, the user's bytes are untouched, and wiki/ is never
// created.
func TestInitPartialConflictRefusesAndPreservesUserFile(t *testing.T) {
	dir := t.TempDir()
	userBytes := []byte("# my own file\n")
	if err := os.WriteFile(filepath.Join(dir, "llm-wiki.yaml"), userBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := exec(t, "init", dir, "--json")
	if code != int(contract.ExitApprovalRequired) {
		t.Fatalf("exit = %d, want 3", code)
	}
	env := decodeEnvelope(t, stdout)
	if len(env.Approval.Paths) != 1 || env.Approval.Paths[0] != "llm-wiki.yaml" {
		t.Errorf("approval paths = %v, want [llm-wiki.yaml]", env.Approval.Paths)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "llm-wiki.yaml")); !bytes.Equal(got, userBytes) {
		t.Errorf("user file was modified: %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "wiki")); !os.IsNotExist(err) {
		t.Errorf("wiki/ must not be created on a refused init")
	}
}

// --force grants the approval: init over an existing bundle succeeds and
// rewrites the files.
func TestInitForceOverwritesExistingBundle(t *testing.T) {
	dir := t.TempDir()
	if _, _, code := exec(t, "init", dir, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("first init exit = %d, want 0", code)
	}
	// Corrupt one file so we can prove --force rewrote it.
	if err := os.WriteFile(filepath.Join(dir, "wiki", "index.md"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := exec(t, "init", dir, "--force", "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("forced init exit = %d, want 0\n%s", code, stdout)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "wiki", "index.md")); string(got) == "stale" {
		t.Errorf("--force did not rewrite wiki/index.md")
	}
}

func TestInitExplicitCoreProfileSucceeds(t *testing.T) {
	dir := t.TempDir()
	_, _, code := exec(t, "init", dir, "--profile", "core", "--json")
	if code != int(contract.ExitSuccess) {
		t.Errorf("init --profile core exit = %d, want 0", code)
	}
}

func TestInitUnknownProfileIsInvalidInvocation(t *testing.T) {
	dir := t.TempDir()
	stdout, _, code := exec(t, "init", dir, "--profile", "bogus", "--json")
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusInvalidInvocation {
		t.Errorf("status = %q, want invalid-invocation", env.Status)
	}
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit = %d, want 4", code)
	}
}

func TestInitMissingTargetDirIsInvalidInvocation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")
	_, _, code := exec(t, "init", dir, "--json")
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit = %d, want 4", code)
	}
}

func TestInitUnknownFlagIsInvalidInvocation(t *testing.T) {
	dir := t.TempDir()
	_, _, code := exec(t, "init", dir, "--frobnicate", "--json")
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit = %d, want 4", code)
	}
}

// Human-mode success prints the created paths to stdout and no JSON envelope.
func TestInitHumanModeSuccessListsPaths(t *testing.T) {
	dir := t.TempDir()
	stdout, _, code := exec(t, "init", dir)
	if code != int(contract.ExitSuccess) {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.Contains(stdout, "{") {
		t.Errorf("human-mode success must not emit JSON: %s", stdout)
	}
	if !strings.Contains(stdout, "wiki/index.md") {
		t.Errorf("success output should list created paths: %s", stdout)
	}
}

// Human-mode refusal prints the reason and --force hint to stderr, exit 3.
func TestInitHumanModeRefusalPrintsHintToStderr(t *testing.T) {
	dir := t.TempDir()
	if _, _, code := exec(t, "init", dir); code != int(contract.ExitSuccess) {
		t.Fatalf("first init exit = %d, want 0", code)
	}
	stdout, stderr, code := exec(t, "init", dir)
	if code != int(contract.ExitApprovalRequired) {
		t.Fatalf("exit = %d, want 3", code)
	}
	if strings.Contains(stdout, "{") {
		t.Errorf("human-mode refusal must not emit JSON to stdout: %s", stdout)
	}
	if !strings.Contains(stderr, "--force") {
		t.Errorf("refusal should hint --force on stderr: %s", stderr)
	}
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

// brokenLinkFixturePage is complete and kebab-named, so its only finding is a
// single core-broken-link (an intra-wiki link to a page absent from the bundle).
const brokenLinkFixturePage = "---\ntype: concept\ntitle: Alpha\ndescription: A page.\ntimestamp: 2026-01-01\ntags: [x]\naliases: [y]\nresource: r\n---\n# Alpha\nSee [gone](missing-target.md).\n"

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

// A page whose only issue is a broken intra-wiki link is, by default, a warning:
// success-with-warnings, exit 1 (ADR-004 FR8 default severity).
func TestValidateBrokenLinkDefaultIsWarning(t *testing.T) {
	dir := writeFixtureDir(t, map[string]string{"alpha-page.md": brokenLinkFixturePage})
	stdout, _, code := exec(t, "validate", dir, "--json")

	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusSuccessWithWarnings {
		t.Fatalf("status = %q, want success-with-warnings (%+v)", env.Status, env.Findings)
	}
	if code != int(contract.ExitSuccessWithWarnings) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitSuccessWithWarnings))
	}
}

// Configuring core-broken-link at error severity (LLM_WIKI_SEVERITY) promotes the
// warning to a validation failure end-to-end: the emitted envelope flips to
// validation-failure and the process exits 2. This is the runtime override path
// the engine-level Resolve unit test could not exercise (ADR-004 core-default →
// profile-override precedence, #4).
func TestValidateBrokenLinkPromotedToErrorViaConfig(t *testing.T) {
	t.Setenv("LLM_WIKI_SEVERITY", "core-broken-link=error")
	dir := writeFixtureDir(t, map[string]string{"alpha-page.md": brokenLinkFixturePage})
	stdout, _, code := exec(t, "validate", dir, "--json")

	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusValidationFailure {
		t.Fatalf("status = %q, want validation-failure (%+v)", env.Status, env.Findings)
	}
	if code != int(contract.ExitValidationFailure) {
		t.Errorf("exit code = %d, want %d", code, int(contract.ExitValidationFailure))
	}
	var saw bool
	for _, f := range env.Findings {
		if f.Code == "core-broken-link" {
			saw = true
			if f.Severity != contract.SeverityError {
				t.Errorf("core-broken-link severity = %q, want error (promoted)", f.Severity)
			}
		}
	}
	if !saw {
		t.Errorf("expected a core-broken-link finding, got %+v", env.Findings)
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

// severityOverrides parses configured code=severity pairs into a Resolve map.
func TestSeverityOverridesParsesConfiguredPairs(t *testing.T) {
	got := severityOverrides("core-broken-link=error, core-kebab-filename=warning")
	if got["core-broken-link"] != contract.SeverityError {
		t.Errorf("core-broken-link = %q, want error", got["core-broken-link"])
	}
	if got["core-kebab-filename"] != contract.SeverityWarning {
		t.Errorf("core-kebab-filename = %q, want warning", got["core-kebab-filename"])
	}
}

// An empty config yields nil, the identity Resolve treats as a core-only run.
func TestSeverityOverridesEmptyIsNil(t *testing.T) {
	if got := severityOverrides(""); got != nil {
		t.Errorf("empty config should yield nil overrides, got %v", got)
	}
}

// A pair whose severity is not one of the three levels is ignored rather than
// crashing or silently coercing.
func TestSeverityOverridesIgnoresInvalidSeverity(t *testing.T) {
	got := severityOverrides("core-broken-link=bogus")
	if _, ok := got["core-broken-link"]; ok {
		t.Errorf("invalid severity value must be ignored, got %v", got)
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
