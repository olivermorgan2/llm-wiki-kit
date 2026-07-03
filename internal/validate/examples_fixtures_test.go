package validate

import (
	"os"
	"path"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// Acceptance-fixture corpus (issue #6). These tests assert the on-disk
// core-profile example pages against the real engine, so the fixtures that
// ADR-004 names as its backing evidence (criteria 5, 7, 8) stay honest as the
// rules evolve. Fixtures live under profiles/core/examples/{valid,invalid}/ and
// are loaded here with zero production changes: Run accepts any fs.FS.
const (
	validExamplesDir   = "../../profiles/core/examples/valid"
	invalidExamplesDir = "../../profiles/core/examples/invalid"
)

// invalidFixtureCodes is the source of truth mapping each invalid fixture to the
// single rule Code it must trip. Each fixture is a fully valid core page with
// exactly one mutation, so validated as a single-file bundle it yields exactly
// one finding whose Code is listed here. The invalid/README.md mirrors this map
// for humans.
var invalidFixtureCodes = map[string]string{
	"malformed-yaml.md":      CodeYAMLParse,      // criterion 7
	"missing-type.md":        codeOKFTypePresent, // criteria 5, 7
	"missing-title.md":       codeCoreReqTitle,   // criteria 5, 7
	"missing-description.md": codeCoreReqDesc,    // criterion 7
	"wrong-field-type.md":    codeCoreFieldType,  // criterion 7
	"broken-link.md":         codeCoreBrokenLink, // criterion 8
	"Not_Kebab.md":           codeCoreKebab,      // filename rule
}

// TestInvalidExamplesFailExactlyOneRule validates every invalid fixture on its
// own (a single-file fstest.MapFS) and asserts it produces exactly one finding
// whose Code matches the expected rule. Single-file isolation is what makes
// "fails exactly one identifiable rule" machine-checkable: e.g. broken-link.md's
// target is absent from the one-file bundle, and no incidental recommended-field
// or kebab finding can fire on a fixture that is otherwise complete.
func TestInvalidExamplesFailExactlyOneRule(t *testing.T) {
	for name, wantCode := range invalidFixtureCodes {
		name, wantCode := name, wantCode
		t.Run(name, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(invalidExamplesDir, name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			// The map key IS the filename under test (e.g. Not_Kebab.md), so the
			// filename-sensitive kebab rule sees the real name.
			fsys := fstest.MapFS{name: {Data: content}}
			got := engine().Run(fsys)
			if len(got) != 1 {
				t.Fatalf("want exactly one finding, got %d: %+v", len(got), got)
			}
			if got[0].Code != wantCode {
				t.Fatalf("finding Code = %q, want %q (%+v)", got[0].Code, wantCode, got[0])
			}
		})
	}
}

// TestInvalidExamplesDirMatchesTable guards against drift: every non-README .md
// file in the invalid dir must be covered by invalidFixtureCodes, and every
// table entry must exist on disk. A new fixture without an expected Code (or a
// deleted fixture) fails here rather than silently going unchecked.
func TestInvalidExamplesDirMatchesTable(t *testing.T) {
	entries, err := os.ReadDir(invalidExamplesDir)
	if err != nil {
		t.Fatalf("read invalid dir: %v", err)
	}
	onDisk := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".md" || e.Name() == "README.md" {
			continue
		}
		onDisk[e.Name()] = true
		if _, ok := invalidFixtureCodes[e.Name()]; !ok {
			t.Errorf("fixture %q has no expected Code in invalidFixtureCodes", e.Name())
		}
	}
	for name := range invalidFixtureCodes {
		if !onDisk[name] {
			t.Errorf("table lists %q but it is not on disk", name)
		}
	}
}

// TestValidExamplesHaveNoErrorFindings validates the valid pages as one bundle
// (so intra-wiki cross-links resolve without spurious broken-link warnings) and
// asserts the acceptance criterion: no error-severity findings. The bundle
// excludes README.md, which is human documentation rather than an OKF page.
func TestValidExamplesHaveNoErrorFindings(t *testing.T) {
	entries, err := os.ReadDir(validExamplesDir)
	if err != nil {
		t.Fatalf("read valid dir: %v", err)
	}
	fsys := fstest.MapFS{}
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".md" || e.Name() == "README.md" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(validExamplesDir, e.Name()))
		if err != nil {
			t.Fatalf("read valid fixture %q: %v", e.Name(), err)
		}
		fsys[e.Name()] = &fstest.MapFile{Data: content}
	}
	if len(fsys) == 0 {
		t.Fatal("no valid example pages found")
	}

	got := engine().Run(fsys)
	if len(got) != 0 {
		t.Fatalf("valid bundle should yield no findings, got %d: %+v", len(got), got)
	}
}
