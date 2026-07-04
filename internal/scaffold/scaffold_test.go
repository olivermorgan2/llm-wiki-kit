package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
	"github.com/olivermorgan2/llm-wiki-kit/internal/txn"
	"github.com/olivermorgan2/llm-wiki-kit/internal/validate"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// fixedNow is an arbitrary injected clock so tests never depend on wall-clock
// time (Date.now is banned in determinism-sensitive code paths).
var fixedNow = time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)

func coreProfile(t *testing.T) profile.Profile {
	t.Helper()
	p, err := profile.Resolve(profile.CoreID)
	if err != nil {
		t.Fatalf("resolve core profile: %v", err)
	}
	return p
}

// The plan materializes exactly three targets, lexically sorted, in slash form,
// all mode 0o644.
func TestPlanTargetsAreExactlyTheThreeFiles(t *testing.T) {
	changes := Plan(coreProfile(t), fixedNow)

	var got []string
	for _, c := range changes {
		got = append(got, c.Target)
		if c.Mode != 0o644 {
			t.Errorf("target %q mode = %o, want 0644", c.Target, c.Mode)
		}
	}
	want := []string{"llm-wiki.yaml", "wiki/index.md", "wiki/templates/page-template.md"}
	if len(got) != len(want) {
		t.Fatalf("targets = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("target[%d] = %q, want %q (must be sorted, slash-form)", i, got[i], want[i])
		}
	}
}

// The bundle config is a static literal, but it must remain syntactically valid
// YAML that carries the ADR-007 profile reference. Unmarshal via the adapter
// guards against literal syntax rot.
func TestPlanConfigUnmarshalsAndCarriesProfileReference(t *testing.T) {
	changes := Plan(coreProfile(t), fixedNow)

	var config []byte
	for _, c := range changes {
		if c.Target == "llm-wiki.yaml" {
			config = c.Data
		}
	}
	if config == nil {
		t.Fatal("no llm-wiki.yaml in change set")
	}

	var cfg struct {
		BundleFormat int    `yaml:"bundleFormat"`
		OKFVersion   string `yaml:"okfVersion"`
		Profile      struct {
			ID      string `yaml:"id"`
			Version string `yaml:"version"`
		} `yaml:"profile"`
	}
	if err := yamladapter.New().Unmarshal(config, &cfg); err != nil {
		t.Fatalf("config does not unmarshal: %v\n%s", err, config)
	}
	if cfg.BundleFormat != 1 {
		t.Errorf("bundleFormat = %d, want 1", cfg.BundleFormat)
	}
	if cfg.OKFVersion != "0.1" {
		t.Errorf("okfVersion = %q, want 0.1", cfg.OKFVersion)
	}
	if cfg.Profile.ID != "core" {
		t.Errorf("profile.id = %q, want core", cfg.Profile.ID)
	}
	if cfg.Profile.Version != "0.1.0" {
		t.Errorf("profile.version = %q, want 0.1.0", cfg.Profile.Version)
	}
}

// Two Plan calls with the same injected clock yield byte-identical change sets,
// and the timestamp is rendered from the injected clock.
func TestPlanIsDeterministicForFixedNow(t *testing.T) {
	a := Plan(coreProfile(t), fixedNow)
	b := Plan(coreProfile(t), fixedNow)
	if len(a) != len(b) {
		t.Fatalf("len mismatch: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].Target != b[i].Target || !bytes.Equal(a[i].Data, b[i].Data) || a[i].Mode != b[i].Mode {
			t.Errorf("change[%d] not deterministic", i)
		}
	}
	stamp := fixedNow.Format("2006-01-02")
	for _, c := range a {
		if filepath.Ext(c.Target) == ".md" && !bytes.Contains(c.Data, []byte(stamp)) {
			t.Errorf("%q does not render the injected timestamp %q", c.Target, stamp)
		}
	}
}

// AC1: the scaffolded bundle validates with ZERO findings — stricter than a
// success status, since suggestions (core-recommended-missing) count here.
func TestScaffoldValidatesWithZeroFindings(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, Plan(coreProfile(t), fixedNow))

	findings := validate.New(yamladapter.New()).Run(os.DirFS(dir))
	if len(findings) != 0 {
		t.Fatalf("scaffold must validate with zero findings, got %d:\n%+v", len(findings), findings)
	}
}

func academicProfile(t *testing.T) profile.Profile {
	t.Helper()
	p, err := profile.Resolve("academic-research")
	if err != nil {
		t.Fatalf("resolve academic-research profile: %v", err)
	}
	return p
}

// The academic-research scaffold carries the config, a home page, and one
// authoring template per profiled type — sorted, slash-form, mode 0644.
func TestAcademicPlanTargetsIncludePerTypeTemplates(t *testing.T) {
	changes := Plan(academicProfile(t), fixedNow)
	var got []string
	for _, c := range changes {
		got = append(got, c.Target)
		if c.Mode != 0o644 {
			t.Errorf("target %q mode = %o, want 0644", c.Target, c.Mode)
		}
	}
	want := []string{
		"llm-wiki.yaml",
		"wiki/index.md",
		"wiki/templates/claim.md",
		"wiki/templates/method.md",
		"wiki/templates/question.md",
		"wiki/templates/source.md",
		"wiki/templates/synthesis.md",
	}
	if len(got) != len(want) {
		t.Fatalf("targets = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("target[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// The config records the academic-research profile reference (id + version).
func TestAcademicPlanConfigRecordsProfileReference(t *testing.T) {
	changes := Plan(academicProfile(t), fixedNow)
	var config []byte
	for _, c := range changes {
		if c.Target == "llm-wiki.yaml" {
			config = c.Data
		}
	}
	var cfg struct {
		Profile struct {
			ID      string `yaml:"id"`
			Version string `yaml:"version"`
		} `yaml:"profile"`
	}
	if err := yamladapter.New().Unmarshal(config, &cfg); err != nil {
		t.Fatalf("config does not unmarshal: %v", err)
	}
	if cfg.Profile.ID != "academic-research" || cfg.Profile.Version != "1.0" {
		t.Errorf("profile ref = %q/%q, want academic-research/1.0", cfg.Profile.ID, cfg.Profile.Version)
	}
}

// AC1 for academic-research: the scaffolded bundle validates with ZERO findings
// UNDER the academic-research profile — the per-type templates are valid
// instances (no missing required field/section, no enum/list-min violation, no
// citation obligation, no broken template link).
func TestAcademicScaffoldValidatesWithZeroFindings(t *testing.T) {
	p := academicProfile(t)
	dir := t.TempDir()
	writePlan(t, dir, Plan(p, fixedNow))

	findings := validate.NewWithOptions(yamladapter.New(), validate.Options{Profile: p}).Run(os.DirFS(dir))
	if len(findings) != 0 {
		t.Fatalf("academic scaffold must validate clean under its profile, got %d:\n%+v", len(findings), findings)
	}
}

func TestConflictsEmptyDirIsNil(t *testing.T) {
	dir := t.TempDir()
	got, err := Conflicts(dir, Plan(coreProfile(t), fixedNow))
	if err != nil {
		t.Fatalf("Conflicts: %v", err)
	}
	if got != nil {
		t.Errorf("empty dir must report no conflicts, got %v", got)
	}
}

func TestConflictsReportsPreexistingFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "llm-wiki.yaml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Conflicts(dir, Plan(coreProfile(t), fixedNow))
	if err != nil {
		t.Fatalf("Conflicts: %v", err)
	}
	if len(got) != 1 || got[0] != "llm-wiki.yaml" {
		t.Errorf("conflicts = %v, want [llm-wiki.yaml]", got)
	}
}

// A pre-existing directory where a page target belongs is still a conflict — any
// filesystem entry blocks the scaffold, not just regular files.
func TestConflictsReportsPreexistingDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "wiki", "index.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Conflicts(dir, Plan(coreProfile(t), fixedNow))
	if err != nil {
		t.Fatalf("Conflicts: %v", err)
	}
	var saw bool
	for _, c := range got {
		if c == "wiki/index.md" {
			saw = true
		}
	}
	if !saw {
		t.Errorf("a directory at wiki/index.md must be a conflict, got %v", got)
	}
}

// writePlan writes a change set to disk (creating parent dirs) so a validation
// pass can run over the scaffolded tree.
func writePlan(t *testing.T, root string, changes []txn.FileChange) {
	t.Helper()
	for _, c := range changes {
		p := filepath.Join(root, filepath.FromSlash(c.Target))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, c.Data, c.Mode); err != nil {
			t.Fatal(err)
		}
	}
}
