package main

import (
	"os"
	"path/filepath"
	"testing"
)

// (G31) With a repo anchored at the nearest .llm-wiki/ marker, a bundle page
// whose only link is an in-repo `../` reference to an existing file resolves
// through the ADR-008 repo-path class — no core-broken-link. This exercises the
// runValidate wiring of AnchorRepo + NewWithOptions.
func TestValidateInRepoRelativeLinkResolvesWithAnchor(t *testing.T) {
	repo := t.TempDir()
	mkdirAll(t, filepath.Join(repo, ".llm-wiki"))
	mkdirAll(t, filepath.Join(repo, "wiki"))
	mkdirAll(t, filepath.Join(repo, "shared"))
	writeFile(t, filepath.Join(repo, "shared", "doc.md"), validFixturePage)
	writeFile(t, filepath.Join(repo, "wiki", "alpha-page.md"),
		validFixturePage+"See [up](../shared/doc.md).\n")

	stdout, _, _ := exec(t, "validate", filepath.Join(repo, "wiki"), "--json")
	env := decodeEnvelope(t, stdout)
	for _, f := range env.Findings {
		if f.Code == "core-broken-link" {
			t.Errorf("in-repo ../ link should resolve with an anchor, got %+v", f)
		}
	}
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
