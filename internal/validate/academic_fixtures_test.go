package validate

import (
	"os"
	"path"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// Acceptance-fixture corpus for the shipped academic-research profile (issue #57
// / I5). It mirrors the core examples_fixtures_test pattern but validates each
// fixture under the resolved academic-research profile, so the addendum-003
// fixture table stays honest as the rules evolve (criteria 5, 7, 8, 9 for the
// profile). Fixtures live under profiles/academic-research/examples/{valid,invalid}/.
const (
	acValidExamplesDir   = "../../profiles/academic-research/examples/valid"
	acInvalidExamplesDir = "../../profiles/academic-research/examples/invalid"
)

// academicProfile resolves the shipped profile once for the corpus.
func academicProfile(t *testing.T) profile.Profile {
	t.Helper()
	p, err := profile.Resolve("academic-research")
	if err != nil {
		t.Fatalf("resolve academic-research: %v", err)
	}
	return p
}

func academicEngine(t *testing.T) *Engine {
	return NewWithOptions(yamladapter.New(), Options{Profile: academicProfile(t)})
}

// acInvalidFixtureCodes maps each invalid fixture to the single rule Code it must
// trip. The invalid/README.md mirrors this for humans (addendum-003 table).
var acInvalidFixtureCodes = map[string]string{
	"source-missing-authors.md":         codeProfileRequiredField,
	"source-no-doi-or-url.md":           codeProfileRecommendedPair,
	"source-bad-source-type.md":         codeProfileFieldEnum,
	"source-empty-authors.md":           codeProfileListMin,
	"claim-supported-no-citation.md":    codeProfileCitationRequired,
	"claim-bad-confidence.md":           codeProfileFieldEnum,
	"method-missing-limitations.md":     codeProfileRequiredSection,
	"claim-cites-question.md":           codeProfileCitationTargetType,
	"synthesis-missing-disagreement.md": codeProfileRequiredSection,
}

// acFixtureCompanions supplies extra in-bundle files a fixture needs for its one
// intended finding to fire. claim-cites-question.md must cite a RESOLVABLE
// question page so it trips profile-citation-target-type (not
// core-citation-unresolved); the companion is the valid question fixture placed
// at the cited path.
var acFixtureCompanions = map[string]map[string]string{
	"claim-cites-question.md": {"question.md": "valid/question.md"},
}

// TestAcademicInvalidExamplesFailExactlyOneRule validates every invalid fixture
// (with any required companion) under the academic-research profile and asserts
// exactly one finding whose Code matches the addendum-003 table. Single-fixture
// isolation is what makes "fails exactly one identifiable rule" machine-checkable.
func TestAcademicInvalidExamplesFailExactlyOneRule(t *testing.T) {
	for name, wantCode := range acInvalidFixtureCodes {
		name, wantCode := name, wantCode
		t.Run(name, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(acInvalidExamplesDir, name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			fsys := fstest.MapFS{name: {Data: content}}
			// Add any companion files (e.g. a resolvable question target).
			for dst, src := range acFixtureCompanions[name] {
				companion, err := os.ReadFile(filepath.Join(acValidExamplesDir, "..", src))
				if err != nil {
					t.Fatalf("read companion %q: %v", src, err)
				}
				fsys[dst] = &fstest.MapFile{Data: companion}
			}

			got := academicEngine(t).Run(fsys)
			if len(got) != 1 {
				t.Fatalf("want exactly one finding, got %d: %+v", len(got), got)
			}
			if got[0].Code != wantCode {
				t.Fatalf("finding Code = %q, want %q (%+v)", got[0].Code, wantCode, got[0])
			}
		})
	}
}

// TestAcademicInvalidExamplesDirMatchesTable guards against drift: every non-README
// .md in the invalid dir is covered by the table, and every table entry exists.
func TestAcademicInvalidExamplesDirMatchesTable(t *testing.T) {
	entries, err := os.ReadDir(acInvalidExamplesDir)
	if err != nil {
		t.Fatalf("read invalid dir: %v", err)
	}
	onDisk := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".md" || e.Name() == "README.md" {
			continue
		}
		onDisk[e.Name()] = true
		if _, ok := acInvalidFixtureCodes[e.Name()]; !ok {
			t.Errorf("fixture %q has no expected Code in acInvalidFixtureCodes", e.Name())
		}
	}
	for name := range acInvalidFixtureCodes {
		if !onDisk[name] {
			t.Errorf("table lists %q but it is not on disk", name)
		}
	}
}

// TestAcademicValidExamplesHaveNoFindings validates the valid pages as one bundle
// (so cross-links resolve) under the academic-research profile and asserts zero
// findings — OKF, core, and profile all clean. README.md is excluded (human doc).
func TestAcademicValidExamplesHaveNoFindings(t *testing.T) {
	entries, err := os.ReadDir(acValidExamplesDir)
	if err != nil {
		t.Fatalf("read valid dir: %v", err)
	}
	fsys := fstest.MapFS{}
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".md" || e.Name() == "README.md" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(acValidExamplesDir, e.Name()))
		if err != nil {
			t.Fatalf("read valid fixture %q: %v", e.Name(), err)
		}
		fsys[e.Name()] = &fstest.MapFile{Data: content}
	}
	if len(fsys) == 0 {
		t.Fatal("no valid example pages found")
	}

	got := academicEngine(t).Run(fsys)
	if len(got) != 0 {
		t.Fatalf("valid academic bundle should yield no findings, got %d: %+v", len(got), got)
	}
}

// TestAcademicValidExamplesCoverEveryProfiledType ensures the corpus ships a valid
// fixture for each of the five added/tightened types (>=1 valid per type).
func TestAcademicValidExamplesCoverEveryProfiledType(t *testing.T) {
	want := map[string]bool{"source": false, "claim": false, "method": false, "question": false, "synthesis": false}
	entries, err := os.ReadDir(acValidExamplesDir)
	if err != nil {
		t.Fatalf("read valid dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".md" || e.Name() == "README.md" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(acValidExamplesDir, e.Name()))
		if err != nil {
			t.Fatalf("read %q: %v", e.Name(), err)
		}
		fm, _, err := splitFrontmatter(content)
		if err != nil {
			t.Fatalf("split %q: %v", e.Name(), err)
		}
		var m map[string]any
		if err := yamladapter.New().Unmarshal(fm, &m); err != nil {
			t.Fatalf("parse %q: %v", e.Name(), err)
		}
		if ty, _ := m["type"].(string); want[ty] == false {
			if _, ok := want[ty]; ok {
				want[ty] = true
			}
		}
	}
	for ty, covered := range want {
		if !covered {
			t.Errorf("no valid fixture for profiled type %q", ty)
		}
	}
}
