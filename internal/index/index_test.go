package index

import (
	"testing"
)

// --- ParsePageMetadata (ADR-011 sub-decision 2) --------------------------------

func TestParsePageMetadataFullFrontmatter(t *testing.T) {
	content := []byte("---\ntype: claim\ntitle: Claim One\ndescription: First claim\n---\nbody text\n")
	got, err := ParsePageMetadata(content, "claims/claim-001.md")
	if err != nil {
		t.Fatalf("ParsePageMetadata returned error: %v", err)
	}
	want := PageMetadata{
		Type:        "claim",
		Title:       "Claim One",
		Description: "First claim",
		Path:        "claims/claim-001.md",
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestParsePageMetadataTitleStemFallback(t *testing.T) {
	// Missing title → basename minus extension. Path recorded verbatim.
	content := []byte("---\ntype: claim\ndescription: no title here\n---\n")
	got, err := ParsePageMetadata(content, "claims/claim-002.md")
	if err != nil {
		t.Fatalf("ParsePageMetadata returned error: %v", err)
	}
	if got.Title != "claim-002" {
		t.Errorf("Title = %q, want stem fallback %q", got.Title, "claim-002")
	}
	if got.Path != "claims/claim-002.md" {
		t.Errorf("Path = %q, want verbatim relPath", got.Path)
	}
}

func TestParsePageMetadataEmptyTitleStemFallback(t *testing.T) {
	// An explicitly empty title is treated the same as absent.
	content := []byte("---\ntype: claim\ntitle: \"\"\n---\n")
	got, err := ParsePageMetadata(content, "notes/deep/thought.md")
	if err != nil {
		t.Fatalf("ParsePageMetadata returned error: %v", err)
	}
	if got.Title != "thought" {
		t.Errorf("Title = %q, want stem fallback %q", got.Title, "thought")
	}
}

func TestParsePageMetadataDescriptionOmitted(t *testing.T) {
	content := []byte("---\ntype: source\ntitle: Source One\n---\n")
	got, err := ParsePageMetadata(content, "sources/source-001.md")
	if err != nil {
		t.Fatalf("ParsePageMetadata returned error: %v", err)
	}
	if got.Description != "" {
		t.Errorf("Description = %q, want empty", got.Description)
	}
}

func TestParsePageMetadataTypeUnknownFallback(t *testing.T) {
	// Missing type is NOT a parse error — it falls back to "unknown".
	content := []byte("---\ntitle: Orphan\ndescription: no type\n---\n")
	got, err := ParsePageMetadata(content, "misc/orphan.md")
	if err != nil {
		t.Fatalf("ParsePageMetadata returned error (missing type must not error): %v", err)
	}
	if got.Type != "unknown" {
		t.Errorf("Type = %q, want %q", got.Type, "unknown")
	}
}

func TestParsePageMetadataEmptyTypeFallback(t *testing.T) {
	content := []byte("---\ntype: \"\"\ntitle: Orphan\n---\n")
	got, err := ParsePageMetadata(content, "misc/orphan.md")
	if err != nil {
		t.Fatalf("ParsePageMetadata returned error: %v", err)
	}
	if got.Type != "unknown" {
		t.Errorf("Type = %q, want %q", got.Type, "unknown")
	}
}

func TestParsePageMetadataMissingFrontmatter(t *testing.T) {
	// No leading `---` → structural failure → error (caller excludes the page).
	content := []byte("no frontmatter here\njust body\n")
	if _, err := ParsePageMetadata(content, "misc/bare.md"); err == nil {
		t.Fatal("ParsePageMetadata accepted content with no frontmatter, want error")
	}
}

func TestParsePageMetadataUnterminatedFrontmatter(t *testing.T) {
	content := []byte("---\ntype: claim\ntitle: never closed\n")
	if _, err := ParsePageMetadata(content, "misc/unterminated.md"); err == nil {
		t.Fatal("ParsePageMetadata accepted unterminated frontmatter, want error")
	}
}

func TestParsePageMetadataMalformedYAML(t *testing.T) {
	// An unterminated flow mapping is unambiguously malformed YAML; goccy
	// surfaces a decode error, and the caller excludes the page.
	content := []byte("---\ntitle: {broken\n---\n")
	if _, err := ParsePageMetadata(content, "misc/malformed.md"); err == nil {
		t.Fatal("ParsePageMetadata accepted malformed YAML, want error")
	}
}

func TestParsePageMetadataUnknownFieldsIgnored(t *testing.T) {
	content := []byte("---\ntype: claim\ntitle: Alpha\ndescription: d\ncustom_field: kept\ntags:\n  - a\n  - b\n---\n")
	got, err := ParsePageMetadata(content, "claims/alpha.md")
	if err != nil {
		t.Fatalf("ParsePageMetadata rejected unknown fields: %v", err)
	}
	if got.Type != "claim" || got.Title != "Alpha" || got.Description != "d" {
		t.Errorf("modeled fields wrong: %+v", got)
	}
}

func TestParsePageMetadataSubdirStemFallback(t *testing.T) {
	// Path keeps slash-form verbatim; the stem fallback uses the basename only.
	content := []byte("---\ntype: note\n---\n")
	got, err := ParsePageMetadata(content, "a/b/c/deep-note.md")
	if err != nil {
		t.Fatalf("ParsePageMetadata returned error: %v", err)
	}
	if got.Path != "a/b/c/deep-note.md" {
		t.Errorf("Path = %q, want slash-form verbatim", got.Path)
	}
	if got.Title != "deep-note" {
		t.Errorf("Title = %q, want basename stem %q", got.Title, "deep-note")
	}
}
