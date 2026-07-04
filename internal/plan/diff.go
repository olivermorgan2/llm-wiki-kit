package plan

import (
	"fmt"
	"strings"
)

// diffContext is the number of unchanged lines shown around a change in the
// unified diff preview.
const diffContext = 3

// devNull is the conventional old-side label for a diff that adds a new file
// (an absent base), so a new-page plan renders as a full-file addition.
const devNull = "/dev/null"

// dline is one line of the line-level edit script: a tag (' ' context, '-'
// removed, '+' added) and the line text (newline excluded).
type dline struct {
	tag  byte
	text string
}

// unifiedDiff renders a unified diff from oldContent to newContent labelled with
// oldLabel/newLabel and diffContext lines of context. An absent base (new page)
// passes an empty oldContent and the devNull old label, so every staged line is
// an addition — the required full-file diff. Identical inputs yield "".
//
// The change region is emitted as a single merged hunk spanning the first change
// (minus context) through the last change (plus context). For page-sized files
// this is a valid, readable unified diff without the bookkeeping of splitting
// far-apart changes into separate hunks.
func unifiedDiff(oldLabel, newLabel string, oldContent, newContent []byte) string {
	a := splitLines(oldContent)
	b := splitLines(newContent)
	ops := diffOps(a, b)

	first, last := -1, -1
	for i, o := range ops {
		if o.tag != ' ' {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first < 0 {
		return "" // no changes
	}

	lo := first - diffContext
	if lo < 0 {
		lo = 0
	}
	hi := last + diffContext
	if hi > len(ops)-1 {
		hi = len(ops) - 1
	}

	// 1-based start line on each side is the count of that side's lines before lo.
	aStart, bStart := 1, 1
	for i := 0; i < lo; i++ {
		switch ops[i].tag {
		case ' ':
			aStart++
			bStart++
		case '-':
			aStart++
		case '+':
			bStart++
		}
	}

	aLen, bLen := 0, 0
	for i := lo; i <= hi; i++ {
		switch ops[i].tag {
		case ' ':
			aLen++
			bLen++
		case '-':
			aLen++
		case '+':
			bLen++
		}
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "--- %s\n", oldLabel)
	fmt.Fprintf(&buf, "+++ %s\n", newLabel)
	fmt.Fprintf(&buf, "@@ -%s +%s @@\n", hunkRange(aStart, aLen), hunkRange(bStart, bLen))
	for i := lo; i <= hi; i++ {
		buf.WriteByte(ops[i].tag)
		buf.WriteString(ops[i].text)
		buf.WriteByte('\n')
	}
	return buf.String()
}

// hunkRange formats one side of a hunk header. A zero-length side (a pure
// insertion or deletion) is reported as "start-1,0" per unified-diff convention,
// so a new file's old side is "0,0".
func hunkRange(start, length int) string {
	switch length {
	case 0:
		return fmt.Sprintf("%d,0", start-1)
	case 1:
		return fmt.Sprintf("%d", start)
	default:
		return fmt.Sprintf("%d,%d", start, length)
	}
}

// splitLines splits content into lines, dropping the single trailing newline
// that terminates a well-formed page so it does not produce a spurious empty
// final line. Empty content yields no lines.
func splitLines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
}

// diffOps computes a line-level edit script from a to b via a longest-common-
// subsequence table. Page-sized inputs make the O(m*n) table trivial, and the
// result is deterministic: at a tie the removal from a is emitted first.
func diffOps(a, b []string) []dline {
	m, n := len(a), len(b)
	// dp[i][j] = LCS length of a[i:] and b[j:].
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	var ops []dline
	i, j := 0, 0
	for i < m && j < n {
		switch {
		case a[i] == b[j]:
			ops = append(ops, dline{' ', a[i]})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			ops = append(ops, dline{'-', a[i]})
			i++
		default:
			ops = append(ops, dline{'+', b[j]})
			j++
		}
	}
	for ; i < m; i++ {
		ops = append(ops, dline{'-', a[i]})
	}
	for ; j < n; j++ {
		ops = append(ops, dline{'+', b[j]})
	}
	return ops
}
