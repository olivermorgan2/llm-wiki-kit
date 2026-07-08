package index

import (
	"errors"
	"sort"
	"strings"
)

// Fence markers delimit the engine-managed region of the OKF-reserved index.md
// (ADR-011 sub-decision 3). They are immutable: the engine rewrites only the
// bytes between them, never the markers themselves. A marker is recognized only
// as a whole line (after trimming a trailing CR, so CRLF files parse).
const (
	FenceStart = "<!-- llm-wiki:index:start -->"
	FenceEnd   = "<!-- llm-wiki:index:end -->"
)

// Sentinel errors from ExtractFencedRegion. Two suffice for the whole edge-case
// space (ADR-011 sub-decision 3): #83 maps a missing fence to its append-at-EOF
// path, and #85 maps every malformed variant to the single core-index-unmanaged
// finding — no finer sub-classification is needed now.
var (
	// ErrNoFence reports that content contains no fence markers at all. Callers
	// (#83) treat this as "append a fresh fenced region at EOF", never as an error
	// to surface.
	ErrNoFence = errors.New("index: no fenced region found")
	// ErrMalformedFence reports any structurally ambiguous fence: duplicate
	// starts or ends, a start with no end, an end with no start, or an end that
	// precedes its start (which also covers the nested start,start,end,end case).
	// The engine refuses to guess the authoritative region.
	ErrMalformedFence = errors.New("index: malformed fenced region")
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

// ExtractFencedRegion splits content around exactly one fence pair:
//
//   - before: bytes up to and including the start-marker line and its newline.
//   - fenced: the inner bytes (empty when the end marker immediately follows the
//     start). This is what the caller compares against GenerateIndex to decide
//     staleness.
//   - after:  the end-marker line and everything after it.
//
// The identity Reconstruct(before, fenced, after) == content always holds, so
// human content outside the region and the markers themselves are byte-preserved.
// Markers are matched only as whole lines (a trailing CR is trimmed on the marker
// line, so CRLF files parse; inner bytes are taken verbatim). It returns
// ErrNoFence when no markers are present and ErrMalformedFence for any ambiguous
// arrangement (see the sentinel docs).
func ExtractFencedRegion(content []byte) (before, fenced, after []byte, err error) {
	lines := strings.Split(string(content), "\n")

	var startIdxs, endIdxs []int
	for i, line := range lines {
		switch strings.TrimRight(line, "\r") {
		case FenceStart:
			startIdxs = append(startIdxs, i)
		case FenceEnd:
			endIdxs = append(endIdxs, i)
		}
	}

	if len(startIdxs) == 0 && len(endIdxs) == 0 {
		return nil, nil, nil, ErrNoFence
	}
	if len(startIdxs) != 1 || len(endIdxs) != 1 || startIdxs[0] >= endIdxs[0] {
		return nil, nil, nil, ErrMalformedFence
	}
	startIdx, endIdx := startIdxs[0], endIdxs[0]

	before = []byte(strings.Join(lines[:startIdx+1], "\n") + "\n")
	after = []byte(strings.Join(lines[endIdx:], "\n"))
	if endIdx == startIdx+1 {
		fenced = []byte{} // empty inner region: end marker follows start directly
	} else {
		fenced = []byte(strings.Join(lines[startIdx+1:endIdx], "\n") + "\n")
	}
	return before, fenced, after, nil
}

// Reconstruct assembles before + body + after into the full file bytes. Passing
// GenerateIndex(pages) as body rewrites only the fenced region while preserving
// the markers and all surrounding human content (this covers phase-5-plan's
// "MergeIndex" role). Reconstruct(ExtractFencedRegion(content)) round-trips to
// content exactly.
func Reconstruct(before, body, after []byte) []byte {
	out := make([]byte, 0, len(before)+len(body)+len(after))
	out = append(out, before...)
	out = append(out, body...)
	out = append(out, after...)
	return out
}
