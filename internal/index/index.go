package index

import (
	"sort"
	"strings"
)

// GenerateIndex renders the deterministic INNER region body — the only bytes the
// engine may rewrite (fence markers are immutable, ADR-011 sub-decision 3). The
// output is sorted, LF-only, and byte-idempotent (ADR-011 sub-decision 4): the
// same input always yields the same bytes, independent of the input slice order.
//
// Format (LOCKED here, fixtured in the tests as ADR-011 Consequences require):
//
//  1. Pages are grouped by Type. Sections appear in bytewise-ascending Type
//     order. A single blank line separates consecutive sections.
//  2. Within a section, entries are sorted by Path bytewise ascending (Title is
//     display-only and never a sort key — ADR-011 sub-decision 4). Path is also
//     the structural dedup key: one entry per bundle-relative path.
//  3. Each section is a `## <Type>` header line followed by one line per entry:
//     `- [<Title>](<Path>)` with ` — <Description>` appended when Description is
//     non-empty.
//  4. Empty input → "".
func GenerateIndex(pages []PageMetadata) string {
	if len(pages) == 0 {
		return ""
	}

	// Group by type without mutating the caller's slice.
	byType := make(map[string][]PageMetadata)
	for _, p := range pages {
		byType[p.Type] = append(byType[p.Type], p)
	}

	types := make([]string, 0, len(byType))
	for t := range byType {
		types = append(types, t)
	}
	sort.Strings(types) // bytewise ascending

	var b strings.Builder
	for i, t := range types {
		if i > 0 {
			b.WriteString("\n") // blank line between sections
		}
		entries := byType[t]
		sort.Slice(entries, func(a, c int) bool {
			return entries[a].Path < entries[c].Path // bytewise ascending on Path
		})

		b.WriteString("## ")
		b.WriteString(t)
		b.WriteString("\n")
		for _, e := range entries {
			b.WriteString("- [")
			b.WriteString(e.Title)
			b.WriteString("](")
			b.WriteString(e.Path)
			b.WriteString(")")
			if e.Description != "" {
				b.WriteString(" — ")
				b.WriteString(e.Description)
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}
