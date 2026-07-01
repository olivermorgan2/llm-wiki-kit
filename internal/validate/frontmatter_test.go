package validate

import "testing"

func TestSplitFrontmatterExtractsYAMLAndBody(t *testing.T) {
	in := []byte("---\ntype: concept\ntitle: Alpha\n---\n# Alpha\n\nBody.\n")
	yaml, body, err := splitFrontmatter(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(yaml) != "type: concept\ntitle: Alpha" {
		t.Errorf("yaml = %q", string(yaml))
	}
	if string(body) != "# Alpha\n\nBody.\n" {
		t.Errorf("body = %q", string(body))
	}
}

func TestSplitFrontmatterEmptyBlock(t *testing.T) {
	yaml, _, err := splitFrontmatter([]byte("---\n---\nbody\n"))
	if err != nil {
		t.Fatalf("empty frontmatter should not error: %v", err)
	}
	if string(yaml) != "" {
		t.Errorf("yaml = %q, want empty", string(yaml))
	}
}

func TestSplitFrontmatterMissingBlockErrors(t *testing.T) {
	if _, _, err := splitFrontmatter([]byte("# no frontmatter\n")); err == nil {
		t.Fatal("file without a leading '---' should error")
	}
}

func TestSplitFrontmatterUnterminatedErrors(t *testing.T) {
	if _, _, err := splitFrontmatter([]byte("---\ntype: concept\n")); err == nil {
		t.Fatal("unterminated frontmatter block should error")
	}
}
