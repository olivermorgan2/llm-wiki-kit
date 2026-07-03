// Phase 2 (Install/init) acceptance corpus — the named, criterion-traceable
// gate evidence for the Phase 2 exit gate (design/build-out-plan.md §"Phase 2").
//
// These tests carry the stable TestAcceptance prefix so CI can name-select them
// (`go test ./cmd/llm-wiki -run '^TestAcceptance' -count=1 -v`) and print one
// legible PASS line per criterion per platform. Each test is doc-commented with
// the acceptance criterion and ADR it proves:
//
//	Criterion 1 (install half) — install --dry-run reports the full plan and
//	  mutates nothing; ADR-009 version-record manifest.
//	Criterion 2 — the CLI runs on every supported ADR-002 platform (the
//	  checksum/cross-compile half stays with the cross-compile-smoke job).
//	Criterion 3 — install into a new repo and into a non-empty repo with no
//	  file loss; collision refusal is zero-mutation (ADR-005/ADR-006).
//	Criterion 4-core — init with the core profile validates clean.
//
// The corpus deliberately overlaps some assertions with the per-command unit
// tests in cli_test.go / install_test.go: those give unit coverage, this gives
// criterion-mapped end-to-end journey evidence for the gate. It reuses the
// existing in-process harness (exec, decodeEnvelope, snapshotTree, sha256Hex,
// installTargets, initTargets) rather than a checked-in testdata/ fixture repo
// (Windows autocrlf would break byte-hash assertions and a .git/ dir cannot be
// checked in) — fixture repos are seeded programmatically with \n literals into
// t.TempDir(), which is platform-identical byte-for-byte.
//
// Portability: all envelope/manifest/approval assertions use slash-form
// literals; every disk touch goes through filepath.FromSlash/Join; snapshot
// keys are already ToSlash'd (see snapshotTree in install_test.go). Never
// compare filepath.Join output against an envelope path.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/manifest"
)

// seedTree writes each slash-form relative path→content pair under dir,
// creating parent directories as needed. Unlike writeFixtureDir (cli_test.go),
// it handles nested paths, so a fake repo tree can be materialized in one call.
func seedTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir for %q: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("write %q: %v", rel, err)
		}
	}
}

// seedNonEmptyRepo materializes a hermetic fake git repository in a fresh temp
// dir and returns its path. The files are written with literal \n bytes so the
// tree is byte-identical on every platform — no real `git init`, which would
// produce non-hermetic, platform-variant files and a real .git/ we cannot
// reason about. The .git/* entries let the no-file-loss proof assert that even
// version-control internals are preserved untouched.
func seedNonEmptyRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	seedTree(t, dir, map[string]string{
		".git/config":   "[core]\n\trepositoryformatversion = 0\n",
		".git/HEAD":     "ref: refs/heads/main\n",
		".gitignore":    "bin/\n*.log\n",
		"README.md":     "# Existing Project\n\nAn existing repo the kit installs into.\n",
		"Makefile":      "build:\n\tgo build ./...\n",
		"src/main.go":   "package main\n\nfunc main() {}\n",
		"docs/notes.md": "# Notes\n\nUser-authored notes.\n",
	})
	return dir
}

// TestAcceptanceCriterion2VersionRunsOnHostPlatform — criterion 2 (CLI runs on
// every supported platform), the "runs" half. When executed under the CI matrix
// this yields one named PASS per ADR-002 platform; the checksum/cross-compile
// half stays with the cross-compile-smoke job. Envelope shape per ADR-003.
func TestAcceptanceCriterion2VersionRunsOnHostPlatform(t *testing.T) {
	stdout, stderr, code := exec(t, "version", "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("version exit = %d, want 0 (stderr: %s)", code, stderr)
	}
	env := decodeEnvelope(t, stdout) // asserts the exact six ADR-003 fields
	if env.Operation != "version" {
		t.Errorf("operation = %q, want version", env.Operation)
	}
	if env.Status != contract.StatusSuccess {
		t.Errorf("status = %q, want success", env.Status)
	}
	if env.ContractVersion != "v1" {
		t.Errorf("contractVersion = %q, want v1", env.ContractVersion)
	}
}

// TestAcceptanceCriterion3InstallNewRepoThenValidateClean — criterion 3
// (install into a new repo) plus the ADR-009 version-record manifest and the
// criterion-1 outcome that the freshly installed bundle validates clean.
// Install into an empty dir writes exactly installTargets, the manifest
// catalogues exactly the three scaffold assets (plugin-owned, hashes matching
// on-disk bytes, never self-listed), and validate reports zero findings.
func TestAcceptanceCriterion3InstallNewRepoThenValidateClean(t *testing.T) {
	dir := t.TempDir()

	stdout, _, code := exec(t, "install", dir, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("install exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if len(env.AffectedPaths) != len(installTargets) {
		t.Fatalf("affectedPaths = %v, want %v", env.AffectedPaths, installTargets)
	}
	for i, want := range installTargets {
		if env.AffectedPaths[i] != want {
			t.Errorf("affectedPaths[%d] = %q, want %q", i, env.AffectedPaths[i], want)
		}
	}
	// All four planned files land on disk.
	for _, rel := range installTargets {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("installed file %q missing: %v", rel, err)
		}
	}

	// ADR-009 version-record manifest.
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
	if m.OKF == "" {
		t.Errorf("okf version must be recorded, got empty")
	}
	if m.Profile.ID != "core" {
		t.Errorf("profile id = %q, want core", m.Profile.ID)
	}
	if len(m.Assets) != 3 {
		t.Fatalf("assets = %d, want 3 (%+v)", len(m.Assets), m.Assets)
	}
	for _, a := range m.Assets {
		if a.Path == manifest.Path {
			t.Errorf("manifest must not list itself: %q", a.Path)
		}
		if string(a.Class) != "plugin-owned" {
			t.Errorf("asset %q class = %q, want plugin-owned", a.Path, a.Class)
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
	}

	// Criterion-1 outcome: the installed bundle validates clean.
	stdout, _, code = exec(t, "validate", dir, "--json")
	env = decodeEnvelope(t, stdout)
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

// TestAcceptanceCriterion3InstallNonEmptyRepoNoFileLoss — criterion 3 (install
// into a non-empty repo with no file loss). Install over a seeded fake repo
// succeeds; the post-install tree is exactly the pre-install tree plus the four
// scaffold files, every pre-existing file (including .git/*) is byte-identical,
// and the manifest catalogues only the three scaffold assets — never any user
// file (ADR-009). No validate-clean chain here: the seeded user .md files
// legitimately produce findings, so the no-file-loss proof is snapshot equality,
// not a clean validate (that lives in the fresh-dir journey above).
func TestAcceptanceCriterion3InstallNonEmptyRepoNoFileLoss(t *testing.T) {
	dir := seedNonEmptyRepo(t)
	before := snapshotTree(t, dir)

	stdout, _, code := exec(t, "install", dir, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("install exit = %d, want 0\n%s", code, stdout)
	}
	after := snapshotTree(t, dir)

	// Every seeded file survives byte-identical.
	for rel, b := range before {
		got, ok := after[rel]
		if !ok {
			t.Errorf("seeded file %q lost after install", rel)
			continue
		}
		if !bytes.Equal(got, b) {
			t.Errorf("seeded file %q modified by install", rel)
		}
	}
	// The post tree is exactly the pre tree plus the four scaffold files.
	if len(after) != len(before)+len(installTargets) {
		t.Errorf("file count = %d, want %d (pre) + %d (scaffold)", len(after), len(before), len(installTargets))
	}
	for _, rel := range installTargets {
		if _, ok := after[rel]; !ok {
			t.Errorf("scaffold file %q not present after install", rel)
		}
	}

	// The manifest catalogues only the three scaffold assets, no user files.
	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(manifest.Path)))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	m, err := manifest.Parse(raw)
	if err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	if len(m.Assets) != 3 {
		t.Fatalf("assets = %d, want 3 scaffold assets (%+v)", len(m.Assets), m.Assets)
	}
	scaffold := map[string]bool{}
	for _, rel := range installTargets {
		if rel != manifest.Path { // the manifest never lists itself
			scaffold[rel] = true
		}
	}
	for _, a := range m.Assets {
		if !scaffold[a.Path] {
			t.Errorf("manifest lists a non-scaffold asset %q (user files must not be catalogued)", a.Path)
		}
	}
}

// TestAcceptanceCriterion3CollisionRefusalIsZeroMutation — criterion 3's
// safety half: a collision with user-owned files refuses (exit 3) and mutates
// nothing. Over a seeded repo that already owns llm-wiki.yaml and wiki/index.md,
// install without --force refuses listing exactly those two conflicts (sorted,
// slash-form), the whole tree is byte-identical afterward, and .llm-wiki/ is
// never created — the refusal precedes txn.Begin (ADR-006), so no staging dir
// leaks.
func TestAcceptanceCriterion3CollisionRefusalIsZeroMutation(t *testing.T) {
	dir := seedNonEmptyRepo(t)
	seedTree(t, dir, map[string]string{
		"llm-wiki.yaml": "# user's own config\n",
		"wiki/index.md": "# user's own wiki index\n",
	})
	before := snapshotTree(t, dir)

	stdout, _, code := exec(t, "install", dir, "--json")
	if code != int(contract.ExitApprovalRequired) {
		t.Fatalf("install exit = %d, want 3\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusApprovalRequired {
		t.Errorf("status = %q, want approval-required", env.Status)
	}
	if env.Approval == nil || !env.Approval.Required {
		t.Fatalf("approval must be present and required, got %+v", env.Approval)
	}
	wantConflicts := []string{"llm-wiki.yaml", "wiki/index.md"}
	got := append([]string(nil), env.Approval.Paths...)
	sort.Strings(got)
	if len(got) != len(wantConflicts) {
		t.Fatalf("approval paths = %v, want exactly %v", env.Approval.Paths, wantConflicts)
	}
	for i, want := range wantConflicts {
		if got[i] != want {
			t.Errorf("approval path[%d] = %q, want %q", i, got[i], want)
		}
	}

	// Zero mutation: the whole tree is byte-identical and .llm-wiki/ is absent.
	after := snapshotTree(t, dir)
	if len(after) != len(before) {
		t.Errorf("refusal changed file count: before %d, after %d", len(before), len(after))
	}
	for rel, b := range before {
		if !bytes.Equal(after[rel], b) {
			t.Errorf("refusal mutated %q", rel)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".llm-wiki")); !os.IsNotExist(err) {
		t.Errorf(".llm-wiki/ must not be created on a refused install")
	}
}

// TestAcceptanceCriterion1InstallDryRunFullPlanNoOp — criterion 1's install
// half: install --dry-run over a non-empty repo reports the full four-path plan
// (exit 0) yet writes nothing. The whole tree is byte-identical afterward and
// .llm-wiki/ is absent — dry-run returns before txn.Begin.
func TestAcceptanceCriterion1InstallDryRunFullPlanNoOp(t *testing.T) {
	dir := seedNonEmptyRepo(t)
	before := snapshotTree(t, dir)

	stdout, _, code := exec(t, "install", "--dry-run", dir, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("dry-run exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if env.Status != contract.StatusSuccess {
		t.Errorf("status = %q, want success", env.Status)
	}
	if len(env.AffectedPaths) != len(installTargets) {
		t.Fatalf("affectedPaths = %v, want full plan %v", env.AffectedPaths, installTargets)
	}
	for i, want := range installTargets {
		if env.AffectedPaths[i] != want {
			t.Errorf("affectedPaths[%d] = %q, want %q", i, env.AffectedPaths[i], want)
		}
	}

	after := snapshotTree(t, dir)
	if len(after) != len(before) {
		t.Errorf("dry-run changed file count: before %d, after %d", len(before), len(after))
	}
	for rel, b := range before {
		if !bytes.Equal(after[rel], b) {
			t.Errorf("dry-run mutated %q", rel)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".llm-wiki")); !os.IsNotExist(err) {
		t.Errorf(".llm-wiki/ must not be created on a dry-run")
	}
}

// TestAcceptanceCriterion4InitCoreThenValidateClean — criterion 4-core: init
// with the core profile scaffolds exactly initTargets and the bundle validates
// clean. Init is distinct from install: it writes no ADR-009 version-record
// manifest (that is install's record; the shared ADR-006 txn layer may still
// leave an empty .llm-wiki/ working area). Core profile only —
// academic-research is Phase 4 and custom is Phase 7, both non-goals here.
func TestAcceptanceCriterion4InitCoreThenValidateClean(t *testing.T) {
	dir := t.TempDir()

	stdout, _, code := exec(t, "init", dir, "--json")
	if code != int(contract.ExitSuccess) {
		t.Fatalf("init exit = %d, want 0\n%s", code, stdout)
	}
	env := decodeEnvelope(t, stdout)
	if len(env.AffectedPaths) != len(initTargets) {
		t.Fatalf("affectedPaths = %v, want %v", env.AffectedPaths, initTargets)
	}
	for i, want := range initTargets {
		if env.AffectedPaths[i] != want {
			t.Errorf("affectedPaths[%d] = %q, want %q", i, env.AffectedPaths[i], want)
		}
	}
	// Exactly the three init targets on disk.
	for _, rel := range initTargets {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("scaffold file %q missing: %v", rel, err)
		}
	}
	// init is not install: it writes no ADR-009 version-record manifest. (The
	// ADR-006 transaction layer, shared by init and install, may leave an empty
	// .llm-wiki/ working area behind, so the meaningful distinction is the
	// absence of the manifest, not of .llm-wiki/ itself.)
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(manifest.Path))); !os.IsNotExist(err) {
		t.Errorf("init must not write the version-record manifest %q (that is install's record)", manifest.Path)
	}

	stdout, _, code = exec(t, "validate", dir, "--json")
	env = decodeEnvelope(t, stdout)
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
