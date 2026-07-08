package index

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// --- GenerateIndex (ADR-011 sub-decisions 2 & 4) -------------------------------

// goldenIndex pins the LOCKED region format byte-for-byte (ADR-011 Consequences).
// It exercises: title link, em-dash description, description-omitted when empty,
// filename-stem title fallback (claim-002), grouping under one header, and the
// single blank line between sections.
const goldenIndex = "## claim\n" +
	"- [Claim One](claims/claim-001.md) — First claim\n" +
	"- [claim-002](claims/claim-002.md)\n" +
	"\n" +
	"## source\n" +
	"- [Source One](sources/source-001.md)\n"

func goldenPages() []PageMetadata {
	return []PageMetadata{
		{Type: "claim", Title: "Claim One", Description: "First claim", Path: "claims/claim-001.md"},
		{Type: "claim", Title: "claim-002", Description: "", Path: "claims/claim-002.md"},
		{Type: "source", Title: "Source One", Description: "", Path: "sources/source-001.md"},
	}
}

func TestGenerateIndexGoldenFormat(t *testing.T) {
	got := GenerateIndex(goldenPages())
	if got != goldenIndex {
		t.Errorf("golden mismatch:\n got=%q\nwant=%q", got, goldenIndex)
	}
}

func TestGenerateIndexIdempotent(t *testing.T) {
	pages := goldenPages()
	first := GenerateIndex(pages)
	second := GenerateIndex(pages)
	if first != second {
		t.Errorf("not byte-idempotent:\nfirst=%q\nsecond=%q", first, second)
	}
}

func TestGenerateIndexInputOrderIndependent(t *testing.T) {
	forward := GenerateIndex(goldenPages())

	shuffled := []PageMetadata{
		{Type: "source", Title: "Source One", Description: "", Path: "sources/source-001.md"},
		{Type: "claim", Title: "claim-002", Description: "", Path: "claims/claim-002.md"},
		{Type: "claim", Title: "Claim One", Description: "First claim", Path: "claims/claim-001.md"},
	}
	if got := GenerateIndex(shuffled); got != forward {
		t.Errorf("output depends on input order:\nshuffled=%q\nforward=%q", got, forward)
	}
}

func TestGenerateIndexSortsTypesThenPaths(t *testing.T) {
	// Types bytewise ascending (alpha before beta); within a type, full
	// bundle-relative path bytewise ascending — a/z.md before b/a.md, proving the
	// sort key is the whole path, not the basename.
	pages := []PageMetadata{
		{Type: "beta", Title: "B", Path: "beta/one.md"},
		{Type: "alpha", Title: "Z", Path: "a/z.md"},
		{Type: "alpha", Title: "A", Path: "b/a.md"},
	}
	want := "## alpha\n" +
		"- [Z](a/z.md)\n" +
		"- [A](b/a.md)\n" +
		"\n" +
		"## beta\n" +
		"- [B](beta/one.md)\n"
	if got := GenerateIndex(pages); got != want {
		t.Errorf("sort order wrong:\n got=%q\nwant=%q", got, want)
	}
}

func TestGenerateIndexGroupsUnderSingleHeader(t *testing.T) {
	pages := []PageMetadata{
		{Type: "claim", Title: "One", Path: "c/1.md"},
		{Type: "claim", Title: "Two", Path: "c/2.md"},
		{Type: "claim", Title: "Three", Path: "c/3.md"},
	}
	got := GenerateIndex(pages)
	if n := strings.Count(got, "## claim"); n != 1 {
		t.Errorf("header count = %d, want exactly 1 in:\n%s", n, got)
	}
}

func TestGenerateIndexIsLFOnly(t *testing.T) {
	if got := GenerateIndex(goldenPages()); strings.Contains(got, "\r") {
		t.Errorf("output contains a carriage return; must be LF-only:\n%q", got)
	}
}

func TestGenerateIndexEmptyInput(t *testing.T) {
	if got := GenerateIndex(nil); got != "" {
		t.Errorf("GenerateIndex(nil) = %q, want empty string", got)
	}
	if got := GenerateIndex([]PageMetadata{}); got != "" {
		t.Errorf("GenerateIndex([]) = %q, want empty string", got)
	}
}

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

// --- ExtractFencedRegion / Reconstruct (ADR-011 sub-decision 3) ----------------

func TestExtractFencedRegionHappyPath(t *testing.T) {
	content := []byte("# My Wiki\n\nHuman preamble.\n" +
		FenceStart + "\n" +
		"## claim\n- [x](x.md)\n" +
		FenceEnd + "\ntrailer text\n")

	before, fenced, after, err := ExtractFencedRegion(content)
	if err != nil {
		t.Fatalf("ExtractFencedRegion returned error: %v", err)
	}
	if got, want := string(fenced), "## claim\n- [x](x.md)\n"; got != want {
		t.Errorf("fenced = %q, want %q", got, want)
	}
	if !bytes.HasSuffix(before, []byte(FenceStart+"\n")) {
		t.Errorf("before must end with the start marker line + newline, got %q", before)
	}
	if !bytes.HasPrefix(after, []byte(FenceEnd)) {
		t.Errorf("after must begin with the end marker, got %q", after)
	}
	if got := Reconstruct(before, fenced, after); !bytes.Equal(got, content) {
		t.Errorf("round-trip mismatch:\n got=%q\nwant=%q", got, content)
	}
}

func TestExtractFencedRegionNoFence(t *testing.T) {
	content := []byte("# Just a doc\n\nno markers here\n")
	_, _, _, err := ExtractFencedRegion(content)
	if !errors.Is(err, ErrNoFence) {
		t.Fatalf("err = %v, want ErrNoFence", err)
	}
}

func TestExtractFencedRegionMalformed(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"two starts", FenceStart + "\n" + FenceStart + "\n" + FenceEnd + "\n"},
		{"two ends", FenceStart + "\n" + FenceEnd + "\n" + FenceEnd + "\n"},
		{"nested start,start,end,end", FenceStart + "\ninner\n" + FenceStart + "\n" + FenceEnd + "\n" + FenceEnd + "\n"},
		{"unterminated start no end", "preamble\n" + FenceStart + "\nbody\n"},
		{"orphan end no start", "preamble\n" + FenceEnd + "\ntrailer\n"},
		{"end before start", FenceEnd + "\nmiddle\n" + FenceStart + "\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := ExtractFencedRegion([]byte(tc.content))
			if !errors.Is(err, ErrMalformedFence) {
				t.Fatalf("err = %v, want ErrMalformedFence", err)
			}
		})
	}
}

func TestExtractFencedRegionCRLFMarkers(t *testing.T) {
	// Markers on CRLF lines are still detected; Reconstruct preserves the exact
	// bytes (including the CR) so the round trip is lossless.
	content := []byte("preamble\r\n" + FenceStart + "\r\ninner\r\n" + FenceEnd + "\r\ntrailer\r\n")
	before, fenced, after, err := ExtractFencedRegion(content)
	if err != nil {
		t.Fatalf("ExtractFencedRegion returned error on CRLF markers: %v", err)
	}
	if got := Reconstruct(before, fenced, after); !bytes.Equal(got, content) {
		t.Errorf("CRLF round-trip mismatch:\n got=%q\nwant=%q", got, content)
	}
}

func TestExtractFencedRegionPreservesHumanContent(t *testing.T) {
	// Preamble and trailer bytes survive extract → reconstruct untouched.
	content := []byte("# Title\n\n> a blockquote\n\n" +
		FenceStart + "\nold body\n" + FenceEnd + "\n\n## Appendix\nkeep me\n")
	before, fenced, after, err := ExtractFencedRegion(content)
	if err != nil {
		t.Fatalf("ExtractFencedRegion returned error: %v", err)
	}
	if !bytes.Contains(before, []byte("> a blockquote")) {
		t.Errorf("preamble content lost from before: %q", before)
	}
	if !bytes.Contains(after, []byte("## Appendix\nkeep me")) {
		t.Errorf("trailer content lost from after: %q", after)
	}
	if got := Reconstruct(before, fenced, after); !bytes.Equal(got, content) {
		t.Errorf("round-trip mismatch:\n got=%q\nwant=%q", got, content)
	}
}

func TestReconstructWithFreshBodyNonEmpty(t *testing.T) {
	content := []byte("preamble\n" + FenceStart + "\nstale\n" + FenceEnd + "\ntrailer\n")
	before, _, after, err := ExtractFencedRegion(content)
	if err != nil {
		t.Fatalf("ExtractFencedRegion returned error: %v", err)
	}
	body := GenerateIndex(goldenPages())
	rebuilt := Reconstruct(before, []byte(body), after)

	// The refreshed region reads back byte-identically, and doing it twice is a
	// fixed point (byte-idempotency across the whole file).
	before2, fenced2, after2, err := ExtractFencedRegion(rebuilt)
	if err != nil {
		t.Fatalf("re-extract returned error: %v", err)
	}
	if string(fenced2) != body {
		t.Errorf("refreshed region = %q, want %q", fenced2, body)
	}
	if got := Reconstruct(before2, []byte(GenerateIndex(goldenPages())), after2); !bytes.Equal(got, rebuilt) {
		t.Errorf("reconstruct not idempotent:\n first=%q\nsecond=%q", rebuilt, got)
	}
}

func TestReconstructWithFreshBodyEmpty(t *testing.T) {
	// Empty pages → empty body → the end marker sits immediately after the start
	// marker's newline (ADR-011 edge case).
	content := []byte("preamble\n" + FenceStart + "\nstale content\n" + FenceEnd + "\ntrailer\n")
	before, _, after, err := ExtractFencedRegion(content)
	if err != nil {
		t.Fatalf("ExtractFencedRegion returned error: %v", err)
	}
	rebuilt := Reconstruct(before, []byte(GenerateIndex(nil)), after)
	want := "preamble\n" + FenceStart + "\n" + FenceEnd + "\ntrailer\n"
	if string(rebuilt) != want {
		t.Errorf("empty-body reconstruct:\n got=%q\nwant=%q", rebuilt, want)
	}
	// And it re-extracts with an empty inner region.
	_, fenced2, _, err := ExtractFencedRegion(rebuilt)
	if err != nil {
		t.Fatalf("re-extract returned error: %v", err)
	}
	if len(fenced2) != 0 {
		t.Errorf("re-extracted inner region = %q, want empty", fenced2)
	}
}

func TestStalenessInvariant(t *testing.T) {
	// Documents the comparison #83/#85 rely on: bytes.Equal(fenced, GenerateIndex)
	// is true when the region is current and false once a page changes.
	body := GenerateIndex(goldenPages())
	content := []byte("preamble\n" + FenceStart + "\n" + body + FenceEnd + "\ntrailer\n")

	_, fenced, _, err := ExtractFencedRegion(content)
	if err != nil {
		t.Fatalf("ExtractFencedRegion returned error: %v", err)
	}
	if !bytes.Equal(fenced, []byte(GenerateIndex(goldenPages()))) {
		t.Errorf("current region reported stale:\nfenced=%q\n regen=%q", fenced, body)
	}

	changed := append(goldenPages(), PageMetadata{Type: "claim", Title: "New", Path: "claims/claim-003.md"})
	if bytes.Equal(fenced, []byte(GenerateIndex(changed))) {
		t.Error("region reported current after a page was added, want stale")
	}
}
