package validate

import (
	"bytes"
	"strings"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
)

// segment is a contiguous run of the page body tagged with whether it lies
// inside a profile-designated evidence context (ADR-008 sub-decision 1).
type segment struct {
	evidence bool
	text     []byte
}

// splitEvidenceContexts partitions body into evidence and navigational segments
// at ATX headings. A heading whose title (trimmed, optional closing-hash run
// stripped) exactly and case-sensitively matches an entry in sections opens an
// evidence context that extends to the next heading of the same-or-shallower
// level; nested sub-headings stay inside. Each matching heading opens its own
// context, so duplicate-citation scope is per-context. An empty sections yields a
// single navigational segment (the zero-cost default). Like the shipped inline
// link regex, this has no code-fence awareness — a `## Evidence` heading inside a
// fenced block still opens a context (reference-style links, autolinks, raw HTML,
// and fenced code remain deferred, mirroring links.go:11-16).
func splitEvidenceContexts(body []byte, sections []string) []segment {
	if len(sections) == 0 {
		return []segment{{evidence: false, text: body}}
	}
	set := make(map[string]bool, len(sections))
	for _, s := range sections {
		set[s] = true
	}

	var segs []segment
	segStart := 0
	curEvidence := false
	curLevel := 0
	flush := func(end int) {
		if end > segStart {
			segs = append(segs, segment{evidence: curEvidence, text: body[segStart:end]})
		}
	}

	pos := 0
	for pos < len(body) {
		lineEnd := pos + len(body[pos:])
		if nl := bytes.IndexByte(body[pos:], '\n'); nl >= 0 {
			lineEnd = pos + nl + 1 // include the newline in the line
		}
		if level, title, ok := parseATX(body[pos:lineEnd]); ok {
			if curEvidence && level <= curLevel {
				flush(pos) // the current evidence context ends before this heading
				curEvidence = false
				segStart = pos
			}
			if set[title] && !curEvidence {
				flush(pos) // start a fresh evidence context at this heading
				segStart = pos
				curEvidence = true
				curLevel = level
			}
		}
		pos = lineEnd
	}
	flush(len(body))
	return segs
}

// parseATX reports whether line is an ATX heading, returning its level (1-6) and
// its title with surrounding whitespace and any optional closing-hash run
// removed. A `#` run must be followed by a space/tab or the end of line.
func parseATX(line []byte) (int, string, bool) {
	s := strings.TrimRight(string(line), "\r\n")
	s = strings.TrimLeft(s, " ") // CommonMark allows minor leading indentation
	n := 0
	for n < len(s) && s[n] == '#' {
		n++
	}
	if n == 0 || n > 6 {
		return 0, "", false
	}
	rest := s[n:]
	if rest != "" && rest[0] != ' ' && rest[0] != '\t' {
		return 0, "", false // e.g. `#tag`, not a heading
	}
	return n, stripClosingHashes(strings.TrimSpace(rest)), true
}

// stripClosingHashes removes an optional trailing `#` closing sequence from an
// ATX heading title. The run must be at the very end and preceded by whitespace
// (or be the whole title), so a title like `C#` is preserved.
func stripClosingHashes(title string) string {
	trimmed := strings.TrimRight(title, "#")
	if trimmed == title {
		return title
	}
	if trimmed == "" || strings.HasSuffix(trimmed, " ") || strings.HasSuffix(trimmed, "\t") {
		return strings.TrimSpace(trimmed)
	}
	return title
}

// citationFindings classifies every inline link inside an evidence context and
// emits the aggregated core-citation-* findings (ADR-008 sub-decisions 2, 5, 7).
// Images are skipped (never citations). Per distinct dedupe key at most one
// finding fires, with precedence malformed > unresolved > duplicate: classify
// gives each key exactly one verdict, and duplicate fires only for otherwise
// resolved keys, so a malformed or unresolved target repeated within a context
// yields only its malformed/unresolved finding (ambiguity #2). Duplicate scope is
// per-context; malformed and unresolved dedupe page-wide. Each code aggregates
// into one finding per page (targets in trimmed original spelling, first-seen
// order) so its {ruleset, code, path} baseline fingerprint stays unique.
func (r *resolver) citationFindings(pagePath string, segments []segment) []contract.Finding {
	var malformed, unresolved, duplicate []string
	seenMalformed := map[string]bool{}
	seenUnresolved := map[string]bool{}
	seenDuplicate := map[string]bool{}

	for _, seg := range segments {
		if !seg.evidence {
			continue
		}
		counts := map[string]int{}
		spellingByKey := map[string]string{}
		var resolvedOrder []string
		for _, m := range inlineLink.FindAllSubmatch(seg.text, -1) {
			if len(m[1]) > 0 {
				continue // image link: never a citation (sub-decision 1).
			}
			target := strings.TrimSpace(string(m[2]))
			res := r.classify(target)
			switch {
			case res.class == classMalformed:
				if !seenMalformed[res.key] {
					seenMalformed[res.key] = true
					malformed = append(malformed, target)
				}
			case !res.resolved:
				if !seenUnresolved[res.key] {
					seenUnresolved[res.key] = true
					unresolved = append(unresolved, target)
				}
			default: // resolved: eligible for per-context duplicate detection.
				if _, ok := spellingByKey[res.key]; !ok {
					spellingByKey[res.key] = target
					resolvedOrder = append(resolvedOrder, res.key)
				}
				counts[res.key]++
			}
		}
		for _, k := range resolvedOrder {
			if counts[k] >= 2 && !seenDuplicate[k] {
				seenDuplicate[k] = true
				duplicate = append(duplicate, spellingByKey[k])
			}
		}
	}

	var out []contract.Finding
	if len(malformed) > 0 {
		out = append(out, contract.Finding{
			Ruleset:  contract.RulesetProfile,
			Severity: contract.SeverityWarning,
			Code:     codeCoreCitationMalformed,
			Message:  "malformed citation target(s): " + strings.Join(malformed, ", "),
			Path:     pagePath,
		})
	}
	if len(unresolved) > 0 {
		out = append(out, contract.Finding{
			Ruleset:  contract.RulesetProfile,
			Severity: contract.SeverityWarning,
			Code:     codeCoreCitationUnresolved,
			Message:  "unresolved citation target(s): " + strings.Join(unresolved, ", "),
			Path:     pagePath,
		})
	}
	if len(duplicate) > 0 {
		out = append(out, contract.Finding{
			Ruleset:  contract.RulesetProfile,
			Severity: contract.SeveritySuggestion,
			Code:     codeCoreCitationDuplicate,
			Message:  "duplicate citation target(s) in one evidence context: " + strings.Join(duplicate, ", "),
			Path:     pagePath,
		})
	}
	return out
}
