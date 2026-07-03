package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/manifest"
)

// installTargets is the exact set install materializes, in the sorted slash-form
// order the success envelope reports. The manifest sorts first (leading dot).
var installTargets = []string{
	".llm-wiki/manifest.json",
	"llm-wiki.yaml",
	"wiki/index.md",
	"wiki/templates/page-template.md",
}

// snapshotTree records the on-disk bytes of every regular file under dir, keyed
// by slash-form path relative to dir, so a refusal or dry-run can be proven
// non-mutating across the whole tree (not just the scaffold set).
func snapshotTree(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	snap := map[string][]byte{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		snap[filepath.ToSlash(rel)] = b
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot tree: %v", err)
	}
	return snap
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// TestInstallNewRepoSucceedsAndValidatesClean: install into a fresh dir writes
// all four files (exit 0), and the bundle then validates with zero findings —
// validate skips non-.md files, so the manifest is invisible to it.
func TestInstallNewRepoSucceedsAndValidatesClean(t *testing.T) {
	dir := t.TempDir()

	_, _, code := exec(t, "install", dir, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("install exit = %d, want 0", code)
	}
	for _, rel := range installTargets {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("installed file %q missing: %v", rel, err)
		}
	}

	stdout, _, code := exec(t, "validate", dir, "--json")
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusSuccess {
		t.Errorf("validate status = %q, want success (%+v)", env.Status, env.Findings)
	}
	if len(env.Findings) != 0 {
		t.Errorf("installed bundle must validate with zero findings, got %+v", env.Findings)
	}
	if code != int(contract.ExitSuccess) {
		t.Errorf("validate exit = %d, want 0", code)
	}
}

// TestInstallJSONSuccessEnvelope: operation install, the four sorted
// affectedPaths, empty findings, nil approval.
func TestInstallJSONSuccessEnvelope(t *testing.T) {
	dir := t.TempDir()

	stdout, _, code := exec(t, "install", dir, "--json")
	env := decodeEnvelope(t, stdout)
	if env.Operation != "install" {
		t.Errorf("operation = %q, want install", env.Operation)
	}
	if env.Status != contract.StatusSuccess {
		t.Errorf("status = %q, want success", env.Status)
	}
	if len(env.AffectedPaths) != len(installTargets) {
		t.Fatalf("affectedPaths = %v, want %v", env.AffectedPaths, installTargets)
	}
	for i, want := range installTargets {
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

// TestInstallNonEmptyRepoPreservesUserFiles: pre-seeded unrelated files are
// byte-identical after install (data-loss check, criterion 3).
func TestInstallNonEmptyRepoPreservesUserFiles(t *testing.T) {
	dir := t.TempDir()
	seeded := map[string]string{
		"notes.txt":       "my working notes\n",
		"src/main.go":     "package main\n",
		"docs/readme.txt": "hello\n",
	}
	for rel, body := range seeded {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if _, _, code := exec(t, "install", dir, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("install exit = %d, want 0", code)
	}
	for rel, body := range seeded {
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("seeded file %q lost: %v", rel, err)
			continue
		}
		if string(got) != body {
			t.Errorf("seeded file %q modified: %q", rel, got)
		}
	}
}

// TestInstallWritesCorrectVersionRecord: parse the written manifest and confirm
// schema/versions/profile and per-asset hashes re-hashed against on-disk bytes,
// with okf equal to the okfVersion in the written llm-wiki.yaml.
func TestInstallWritesCorrectVersionRecord(t *testing.T) {
	dir := t.TempDir()
	if _, _, code := exec(t, "install", dir, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("install exit = %d, want 0", code)
	}

	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(manifest.Path)))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	m, err := manifest.Parse(raw)
	if err != nil {
		t.Fatalf("parse manifest: %v", err)
	}

	if m.SchemaVersion != manifest.SchemaVersion {
		t.Errorf("schemaVersion = %q, want %q", m.SchemaVersion, manifest.SchemaVersion)
	}
	if m.Plugin != version || m.CLI != version {
		t.Errorf("plugin/cli = %q/%q, want %q", m.Plugin, m.CLI, version)
	}
	if m.Profile.ID != "core" || m.Profile.Version != "0.1.0" {
		t.Errorf("profile = %+v, want {core 0.1.0}", m.Profile)
	}

	// okf matches the value written into llm-wiki.yaml.
	cfg, err := os.ReadFile(filepath.Join(dir, "llm-wiki.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfg), "okfVersion: \""+m.OKF+"\"") {
		t.Errorf("manifest okf %q not found in written config:\n%s", m.OKF, cfg)
	}

	// The three scaffold assets, each hash re-derived from on-disk bytes.
	if len(m.Assets) != 3 {
		t.Fatalf("assets = %d, want 3 (%+v)", len(m.Assets), m.Assets)
	}
	for _, a := range m.Assets {
		if a.Path == manifest.Path {
			t.Errorf("manifest must not list itself: %q", a.Path)
		}
		onDisk, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(a.Path)))
		if err != nil {
			t.Errorf("asset %q missing on disk: %v", a.Path, err)
			continue
		}
		want := sha256Hex(onDisk)
		if a.Hash != want || a.LastInstalledHash != want {
			t.Errorf("asset %q hashes = %q/%q, want %q", a.Path, a.Hash, a.LastInstalledHash, want)
		}
		if string(a.Class) != "plugin-owned" {
			t.Errorf("asset %q class = %q, want plugin-owned", a.Path, a.Class)
		}
	}
}

// TestInstallReinstallRefusesWithApprovalEnvelope: a second install without
// --force refuses (exit 3) listing all four conflicts, and the whole tree is
// unchanged.
func TestInstallReinstallRefusesWithApprovalEnvelope(t *testing.T) {
	dir := t.TempDir()
	if _, _, code := exec(t, "install", dir, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("first install exit = %d, want 0", code)
	}
	before := snapshotTree(t, dir)

	stdout, _, code := exec(t, "install", dir, "--json")
	if code != int(contract.ExitApprovalRequired) {
		t.Fatalf("re-install exit = %d, want 3", code)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusApprovalRequired {
		t.Errorf("status = %q, want approval-required", env.Status)
	}
	if env.Approval == nil || !env.Approval.Required {
		t.Fatalf("approval must be present and required, got %+v", env.Approval)
	}
	if len(env.Approval.Paths) != len(installTargets) {
		t.Errorf("approval paths = %v, want all four targets", env.Approval.Paths)
	}
	for i, want := range installTargets {
		if env.Approval.Paths[i] != want {
			t.Errorf("approval path[%d] = %q, want %q", i, env.Approval.Paths[i], want)
		}
	}
	after := snapshotTree(t, dir)
	if len(after) != len(before) {
		t.Errorf("refusal changed file count: before %d, after %d", len(before), len(after))
	}
	for rel, b := range before {
		if !bytes.Equal(after[rel], b) {
			t.Errorf("refusal mutated %q", rel)
		}
	}
}

// TestInstallPartialConflictRefusesAndPreservesUserFile: a user-owned
// llm-wiki.yaml pre-exists, so install refuses listing exactly it, the user's
// bytes are untouched, and neither .llm-wiki/ nor wiki/ is created.
func TestInstallPartialConflictRefusesAndPreservesUserFile(t *testing.T) {
	dir := t.TempDir()
	userBytes := []byte("# my own config\n")
	if err := os.WriteFile(filepath.Join(dir, "llm-wiki.yaml"), userBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := exec(t, "install", dir, "--json")
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
	if _, err := os.Stat(filepath.Join(dir, ".llm-wiki")); !os.IsNotExist(err) {
		t.Errorf(".llm-wiki/ must not be created on a refused install")
	}
	if _, err := os.Stat(filepath.Join(dir, "wiki")); !os.IsNotExist(err) {
		t.Errorf("wiki/ must not be created on a refused install")
	}
}

// TestInstallForceOverwritesAndRewritesManifest: corrupt a scaffold file, then
// --force reinstall rewrites it and the manifest hashes match disk again.
func TestInstallForceOverwritesAndRewritesManifest(t *testing.T) {
	dir := t.TempDir()
	if _, _, code := exec(t, "install", dir, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("first install exit = %d, want 0", code)
	}
	if err := os.WriteFile(filepath.Join(dir, "wiki", "index.md"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := exec(t, "install", dir, "--force", "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("forced install exit = %d, want 0\n%s", code, stdout)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "wiki", "index.md")); string(got) == "stale" {
		t.Errorf("--force did not rewrite wiki/index.md")
	}

	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(manifest.Path)))
	if err != nil {
		t.Fatal(err)
	}
	m, err := manifest.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range m.Assets {
		onDisk, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(a.Path)))
		if err != nil {
			t.Errorf("asset %q missing: %v", a.Path, err)
			continue
		}
		if a.Hash != sha256Hex(onDisk) {
			t.Errorf("asset %q hash does not match disk after --force", a.Path)
		}
	}
}

// TestInstallDryRunNewRepoIsCompleteNoOp: dry-run into a fresh dir reports exit
// 0 with the full plan in affectedPaths, and the directory is empty afterward
// (dry-run no-op proof — not even .llm-wiki/ is created).
func TestInstallDryRunNewRepoIsCompleteNoOp(t *testing.T) {
	dir := t.TempDir()

	stdout, _, code := exec(t, "install", "--dry-run", dir, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("dry-run exit = %d, want 0", code)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusSuccess {
		t.Errorf("status = %q, want success", env.Status)
	}
	if len(env.AffectedPaths) != len(installTargets) {
		t.Errorf("affectedPaths = %v, want full plan %v", env.AffectedPaths, installTargets)
	}
	for i, want := range installTargets {
		if env.AffectedPaths[i] != want {
			t.Errorf("affectedPaths[%d] = %q, want %q", i, env.AffectedPaths[i], want)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dry-run mutated the tree: %v", entries)
	}
}

// TestInstallDryRunConflictMirrorsRefusal: dry-run over an already-installed
// bundle mirrors the real refusal (exit 3) and changes nothing.
func TestInstallDryRunConflictMirrorsRefusal(t *testing.T) {
	dir := t.TempDir()
	if _, _, code := exec(t, "install", dir, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("first install exit = %d, want 0", code)
	}
	before := snapshotTree(t, dir)

	stdout, _, code := exec(t, "install", "--dry-run", dir, "--json")
	if code != int(contract.ExitApprovalRequired) {
		t.Fatalf("dry-run conflict exit = %d, want 3", code)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusApprovalRequired {
		t.Errorf("status = %q, want approval-required", env.Status)
	}
	after := snapshotTree(t, dir)
	for rel, b := range before {
		if !bytes.Equal(after[rel], b) {
			t.Errorf("dry-run refusal mutated %q", rel)
		}
	}
}

// TestInstallDryRunForceListsFullPlanWithoutWriting: dry-run + force over a
// conflicting tree lists the full plan (exit 0) but writes nothing.
func TestInstallDryRunForceListsFullPlanWithoutWriting(t *testing.T) {
	dir := t.TempDir()
	if _, _, code := exec(t, "install", dir, "--json"); code != int(contract.ExitSuccess) {
		t.Fatalf("first install exit = %d, want 0", code)
	}
	before := snapshotTree(t, dir)

	stdout, _, code := exec(t, "install", "--dry-run", "--force", dir, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("dry-run --force exit = %d, want 0", code)
	}
	env := decodeEnvelope(t, stdout)
	if len(env.AffectedPaths) != len(installTargets) {
		t.Errorf("affectedPaths = %v, want full plan", env.AffectedPaths)
	}
	after := snapshotTree(t, dir)
	for rel, b := range before {
		if !bytes.Equal(after[rel], b) {
			t.Errorf("dry-run --force mutated %q", rel)
		}
	}
	if len(after) != len(before) {
		t.Errorf("dry-run --force changed file count: before %d, after %d", len(before), len(after))
	}
}

func TestInstallUnknownProfileIsInvalidInvocation(t *testing.T) {
	dir := t.TempDir()
	stdout, _, code := exec(t, "install", dir, "--profile", "bogus", "--json")
	env := decodeEnvelope(t, stdout)
	if env.Operation != "install" {
		t.Errorf("operation = %q, want install", env.Operation)
	}
	if env.Status != contract.StatusInvalidInvocation {
		t.Errorf("status = %q, want invalid-invocation", env.Status)
	}
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit = %d, want 4", code)
	}
}

func TestInstallMissingTargetDirIsInvalidInvocation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")
	stdout, _, code := exec(t, "install", dir, "--json")
	env := decodeEnvelope(t, stdout)
	if env.Operation != "install" {
		t.Errorf("operation = %q, want install", env.Operation)
	}
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit = %d, want 4", code)
	}
}

func TestInstallUnknownFlagIsInvalidInvocation(t *testing.T) {
	dir := t.TempDir()
	stdout, _, code := exec(t, "install", dir, "--frobnicate", "--json")
	env := decodeEnvelope(t, stdout)
	if env.Operation != "install" {
		t.Errorf("operation = %q, want install", env.Operation)
	}
	if code != int(contract.ExitInvalidInvocation) {
		t.Errorf("exit = %d, want 4", code)
	}
}

// Human-mode success lists the created paths (including the manifest) on stdout
// and emits no JSON envelope.
func TestInstallHumanModeSuccessListsPaths(t *testing.T) {
	dir := t.TempDir()
	stdout, _, code := exec(t, "install", dir)
	if code != int(contract.ExitSuccess) {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.Contains(stdout, "{") {
		t.Errorf("human-mode success must not emit JSON: %s", stdout)
	}
	if !strings.Contains(stdout, "wiki/index.md") {
		t.Errorf("success output should list created paths: %s", stdout)
	}
	if !strings.Contains(stdout, manifest.Path) {
		t.Errorf("success output should list the manifest path: %s", stdout)
	}
}

// Human-mode dry-run says "would create" and writes nothing.
func TestInstallHumanModeDryRunSaysWouldCreate(t *testing.T) {
	dir := t.TempDir()
	stdout, _, code := exec(t, "install", "--dry-run", dir)
	if code != int(contract.ExitSuccess) {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.Contains(stdout, "{") {
		t.Errorf("human-mode dry-run must not emit JSON: %s", stdout)
	}
	if !strings.Contains(stdout, "would create") {
		t.Errorf("dry-run output should say 'would create': %s", stdout)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("dry-run mutated the tree: %v", entries)
	}
}

// Human-mode refusal prints the reason and --force hint to stderr, exit 3.
func TestInstallHumanModeRefusalPrintsHintToStderr(t *testing.T) {
	dir := t.TempDir()
	if _, _, code := exec(t, "install", dir); code != int(contract.ExitSuccess) {
		t.Fatalf("first install exit = %d, want 0", code)
	}
	stdout, stderr, code := exec(t, "install", dir)
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
