// Package index builds the deterministic derived view of page frontmatter that
// ADR-011 writes into the HTML-comment fenced region of the OKF-reserved
// index.md. It is a PURE package: it performs no file I/O, no directory walk,
// no staging, and — by construction — makes no model/client calls. It imports
// only the Go standard library and the internal yamladapter seam (ADR-001), so
// the no-model-call invariant (ADR-011 sub-decision 8) holds structurally.
//
// The concrete markdown format of the generated region is locked here behind
// golden fixtures in the package tests, as ADR-011 Consequences require. Higher
// layers compose these primitives: the CLI walk and staged write (#83), the
// validate findings core-index-stale / core-index-unmanaged (#85), and the init
// fence scaffolding (#86) all live outside this package.
package index

import (
	"fmt"
	"path"
	"strings"

	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// PageMetadata is one indexable page, derived solely from frontmatter + path
// (ADR-011 sub-decision 2). No body scraping, no new parser.
type PageMetadata struct {
	Type        string // frontmatter `type`; empty → "unknown"
	Title       string // frontmatter `title`; empty → filename stem
	Description string // frontmatter `description`; may be empty (omitted at render)
	Path        string // bundle-relative, slash-form; the sort + dedup key
}

// frontmatterDelim opens and closes the YAML frontmatter of an OKF page. It
// mirrors internal/validate so index parses the same page shape.
const frontmatterDelim = "---"

// typeUnknown is the fallback section for a page whose frontmatter carries no
// (or an empty) `type` (ADR-011 sub-decision 2).
const typeUnknown = "unknown"

// ParsePageMetadata parses one page's raw bytes into PageMetadata. PURE: no file
// I/O. relPath is recorded verbatim as Path and used for the title-stem
// fallback. It returns an error on missing/unterminated frontmatter or YAML
// parse failure; the #83 walk EXCLUDES such pages (ADR-011 sub-decision 2). An
// absent `type` on an otherwise well-formed page is NOT an error — it falls back
// to "unknown".
func ParsePageMetadata(content []byte, relPath string) (PageMetadata, error) {
	fm, _, err := splitFrontmatter(content)
	if err != nil {
		return PageMetadata{}, fmt.Errorf("index: parse %s: %w", relPath, err)
	}

	var page struct {
		Type        string `yaml:"type"`
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
	}
	if err := yamladapter.New().Unmarshal(fm, &page); err != nil {
		return PageMetadata{}, fmt.Errorf("index: parse %s: %w", relPath, err)
	}

	meta := PageMetadata{
		Type:        page.Type,
		Title:       page.Title,
		Description: page.Description,
		Path:        relPath,
	}
	if meta.Type == "" {
		meta.Type = typeUnknown
	}
	if meta.Title == "" {
		meta.Title = titleStem(relPath)
	}
	return meta, nil
}

// titleStem is the filename-stem fallback for a missing title: the basename of
// the bundle-relative path with its extension removed.
func titleStem(relPath string) string {
	base := path.Base(relPath)
	return strings.TrimSuffix(base, path.Ext(base))
}

// splitFrontmatter separates a leading `---`-fenced YAML frontmatter block from
// the markdown body, returning the raw YAML bytes (delimiters removed) and the
// body. A file that does not open with a `---` line, or whose block is never
// closed, is a structural failure and returns an error. This mirrors
// internal/validate's unexported helper; ADR-011 sub-decision 8 keeps the index
// package on stdlib + yamladapter only, so it is re-implemented locally rather
// than importing across the validate seam (a shared helper is possible
// follow-up DRY work, out of scope for #82).
func splitFrontmatter(data []byte) (yaml []byte, body []byte, err error) {
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != frontmatterDelim {
		return nil, nil, fmt.Errorf("missing frontmatter: file must begin with a %q line", frontmatterDelim)
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == frontmatterDelim {
			yaml = []byte(strings.Join(lines[1:i], "\n"))
			body = []byte(strings.Join(lines[i+1:], "\n"))
			return yaml, body, nil
		}
	}
	return nil, nil, fmt.Errorf("unterminated frontmatter: no closing %q line", frontmatterDelim)
}
