package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

// pageContentHash returns the lowercase-hex SHA-256 of the file at path — the
// independent check that page inspect reports the on-disk content hash.
func pageContentHash(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// initBundle scaffolds a fresh core bundle into a temp dir and returns its root.
func initBundle(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if _, _, code := exec(t, "init", dir, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("init exit = %d, want 0", code)
	}
	return dir
}

// Dispatch: page with no subcommand, an unknown subcommand, and inspect with no
// path are all invalid-invocation (exit 4) and carry that status in JSON.
func TestPageDispatchInvalid(t *testing.T) {
	cases := [][]string{
		{"page"},
		{"page", "bogus"},
		{"page", "inspect"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			_, _, code := exec(t, args...)
			if code != int(contract.ExitInvalidInvocation) {
				t.Errorf("%v exit = %d, want 4", args, code)
			}
			jsonArgs := append(append([]string{}, args...), "--json")
			stdout, _, jcode := exec(t, jsonArgs...)
			if jcode != int(contract.ExitInvalidInvocation) {
				t.Errorf("%v exit = %d, want 4", jsonArgs, jcode)
			}
			env := decodeEnvelope(t, stdout)
			if env.Status != contract.StatusInvalidInvocation {
				t.Errorf("%v status = %q, want invalid-invocation", jsonArgs, env.Status)
			}
			if env.Page != nil {
				t.Errorf("%v must not carry a page payload: %+v", jsonArgs, env.Page)
			}
		})
	}
}

// Happy path: inspect a clean scaffolded page → exit 0, operation "page
// inspect", empty findings, page payload with the on-disk hash.
func TestPageInspectHappyPath(t *testing.T) {
	dir := initBundle(t)

	stdout, _, code := exec(t, "page", "inspect", "--root", dir, "wiki/index.md", "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("page inspect exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Operation != "page inspect" {
		t.Errorf("operation = %q, want \"page inspect\"", env.Operation)
	}
	if env.Status != contract.StatusSuccess {
		t.Errorf("status = %q, want success (%+v)", env.Status, env.Findings)
	}
	if len(env.Findings) != 0 {
		t.Errorf("findings = %+v, want none", env.Findings)
	}
	if env.Page == nil {
		t.Fatal("page payload missing")
	}
	if env.Page.Path != "wiki/index.md" {
		t.Errorf("page.path = %q, want wiki/index.md", env.Page.Path)
	}
	if !env.Page.Parsed {
		t.Error("page.parsed = false, want true")
	}
	want := pageContentHash(t, filepath.Join(dir, "wiki", "index.md"))
	if env.Page.ContentHash != want {
		t.Errorf("page.contentHash = %q, want %q", env.Page.ContentHash, want)
	}
}

// Acceptance — malformed YAML: exit 2, validation-failure, okf-yaml-parse
// finding, page.parsed false, hash still present. Human mode prints the finding.
func TestPageInspectMalformedYAML(t *testing.T) {
	dir := initBundle(t)
	bad := filepath.Join(dir, "wiki", "bad.md")
	if err := os.WriteFile(bad, []byte("---\ntype: concept\ntitle: {broken\n---\n\n# Bad\n"), 0o644); err != nil {
		t.Fatalf("write bad page: %v", err)
	}

	stdout, _, code := exec(t, "page", "inspect", "--root", dir, "wiki/bad.md", "--json")
	if code != int(contract.ExitValidationFailure) {
		t.Fatalf("exit = %d, want 2\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusValidationFailure {
		t.Errorf("status = %q, want validation-failure", env.Status)
	}
	if env.Page == nil || env.Page.Parsed {
		t.Errorf("page.parsed should be false: %+v", env.Page)
	}
	if env.Page != nil && env.Page.ContentHash == "" {
		t.Error("content hash must be populated even for malformed YAML")
	}
	hasParse := false
	for _, f := range env.Findings {
		if f.Code == "okf-yaml-parse" {
			hasParse = true
		}
	}
	if !hasParse {
		t.Errorf("expected an okf-yaml-parse finding, got %+v", env.Findings)
	}

	// Human mode prints the finding line to stdout.
	humanOut, _, hcode := exec(t, "page", "inspect", "--root", dir, "wiki/bad.md")
	if hcode != int(contract.ExitValidationFailure) {
		t.Errorf("human exit = %d, want 2", hcode)
	}
	if !strings.Contains(humanOut, "okf-yaml-parse") {
		t.Errorf("human output should print the finding; got:\n%s", humanOut)
	}
	if !strings.Contains(humanOut, "parse: failed") {
		t.Errorf("human output should show parse: failed; got:\n%s", humanOut)
	}
}

// Acceptance — boundary refusal: a file outside the root reached via ../ is
// refused with exit 5, system-or-filesystem-failure, the standard six-field
// envelope, and no leaked content.
func TestPageInspectBoundaryRefusal(t *testing.T) {
	dir := initBundle(t)
	parent := filepath.Dir(dir)
	secret := filepath.Join(parent, "secret.md")
	if err := os.WriteFile(secret, []byte("---\ntype: secret\n---\nTOPSECRET\n"), 0o644); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	t.Cleanup(func() { os.Remove(secret) })

	stdout, stderr, code := exec(t, "page", "inspect", "--root", dir, "../secret.md", "--json")
	if code != int(contract.ExitSystemFailure) {
		t.Fatalf("exit = %d, want 5\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusSystemFailure {
		t.Errorf("status = %q, want system-or-filesystem-failure", env.Status)
	}
	if env.Page != nil {
		t.Errorf("refusal must not carry a page payload: %+v", env.Page)
	}
	// Standard six-field envelope: no page key.
	var generic map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &generic); err != nil {
		t.Fatalf("stdout not JSON: %v", err)
	}
	if len(generic) != 6 {
		t.Errorf("refusal envelope must be exactly six fields, got %d: %s", len(generic), stdout)
	}
	if strings.Contains(stdout, "TOPSECRET") || strings.Contains(stderr, "TOPSECRET") {
		t.Error("boundary refusal leaked out-of-bundle content")
	}
}

// Missing page → invalid-invocation (exit 4).
func TestPageInspectMissingPage(t *testing.T) {
	dir := initBundle(t)
	_, _, code := exec(t, "page", "inspect", "--root", dir, "wiki/nope.md")
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit = %d, want 4", code)
	}
}

// Envelope shape invariant: a page-inspect JSON run carries exactly seven keys
// (six + page), while a validate JSON run in the same test still carries exactly
// six — guarding the omitempty behavior end to end.
func TestPageInspectEnvelopeShapeVsValidate(t *testing.T) {
	dir := initBundle(t)

	pageOut, _, _ := exec(t, "page", "inspect", "--root", dir, "wiki/index.md", "--json")
	var pageKeys map[string]json.RawMessage
	if err := json.Unmarshal([]byte(pageOut), &pageKeys); err != nil {
		t.Fatalf("page json: %v", err)
	}
	if len(pageKeys) != 7 {
		t.Errorf("page inspect envelope must have 7 keys, got %d: %s", len(pageKeys), pageOut)
	}
	if _, ok := pageKeys["page"]; !ok {
		t.Errorf("page inspect envelope missing page key: %s", pageOut)
	}

	valOut, _, _ := exec(t, "validate", dir, "--json")
	var valKeys map[string]json.RawMessage
	if err := json.Unmarshal([]byte(valOut), &valKeys); err != nil {
		t.Fatalf("validate json: %v", err)
	}
	if len(valKeys) != 6 {
		t.Errorf("validate envelope must have 6 keys, got %d: %s", len(valKeys), valOut)
	}
	if _, ok := valKeys["page"]; ok {
		t.Errorf("validate envelope must not carry a page key: %s", valOut)
	}
}

// Cross-surface parity (criterion 15): for one page, validate's findings
// filtered to that page deep-equal page inspect's findings.
func TestPageInspectParityWithValidate(t *testing.T) {
	dir := initBundle(t)
	// A page with a broken link so there is at least one finding to compare.
	linker := "---\ntype: concept\ntitle: Linker\ndescription: Links out.\n" +
		"timestamp: 2026-07-03\ntags: []\naliases: []\nresource: https://x\n---\n\n" +
		"See [ghost](wiki/ghost.md).\n"
	if err := os.WriteFile(filepath.Join(dir, "wiki", "linker.md"), []byte(linker), 0o644); err != nil {
		t.Fatalf("write linker: %v", err)
	}

	pageOut, _, _ := exec(t, "page", "inspect", "--root", dir, "wiki/linker.md", "--json")
	pageEnv := decodeEnvelope(t, pageOut)

	valOut, _, _ := exec(t, "validate", dir, "--json")
	valEnv := decodeEnvelope(t, valOut)

	var filtered []contract.Finding
	for _, f := range valEnv.Findings {
		if f.Path == "wiki/linker.md" {
			filtered = append(filtered, f)
		}
	}
	if len(filtered) == 0 {
		t.Fatal("expected at least one validate finding for wiki/linker.md")
	}
	if !reflect.DeepEqual(filtered, pageEnv.Findings) {
		t.Errorf("findings differ between surfaces:\n validate: %+v\n page:     %+v", filtered, pageEnv.Findings)
	}
}
