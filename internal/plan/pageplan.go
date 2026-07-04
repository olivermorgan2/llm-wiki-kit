package plan

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
	"github.com/olivermorgan2/llm-wiki-kit/internal/txn"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// ErrTargetNotRegular means the plan target already exists but is not a regular
// file (a directory, symlink, or other special file); ADR-006 stages and
// commits only regular files, so such a target is rejected rather than planned.
var ErrTargetNotRegular = errors.New("plan: target is not a regular file")

// frontmatterDelim opens and closes a page's YAML frontmatter block.
const frontmatterDelim = "---"

// PlanResult is the preview of a staged whole-page change set (ADR-006). It
// records the target's bundle-relative path, whether the change is a no-op
// (proposed content already matches the live page), the staging transaction id
// binding the plan on disk (empty for a no-op), the captured base state
// (BaseAbsent marks a new page's absent-target sentinel; BaseHash is the live
// page's content hash otherwise), the staged content and its hash, and a
// unified diff preview. Planning never mutates live repository files: on a real
// change the staged bytes and manifest live under .llm-wiki/staging/<TxnID>/
// until a later apply; a no-op stages nothing at all.
type PlanResult struct {
	Path       string
	NoOp       bool
	TxnID      string
	BaseAbsent bool
	BaseHash   string
	Staged     []byte
	StagedHash string
	Diff       string
}

// Plan builds a staged whole-page change set for target (relative to root, or an
// absolute path contained by it) from the proposed page content and returns a
// preview. It enforces the ADR-005 boundary gate, captures the source/target
// base hash — the absent-target sentinel when the page is new — normalizes the
// proposed frontmatter through the yamladapter round-trip so unknown fields
// survive (ADR-001, criterion 6), and renders a unified diff.
//
// When the normalized content already equals the live page byte for byte the
// plan is a no-op: it stages nothing (no staging dir is created) and returns
// with NoOp true. Otherwise the change set is staged under
// .llm-wiki/staging/<TxnID>/ via one ADR-006 transaction that is left staged,
// not committed — Plan mutates no live file. A boundary escape passes through
// the fsafe sentinel unwrapped; a non-.md path is ErrNotMarkdown; an existing
// non-regular target is ErrTargetNotRegular.
func Plan(root, target string, proposed []byte, yaml yamladapter.Adapter) (*PlanResult, error) {
	if filepath.Ext(target) != ".md" {
		return nil, ErrNotMarkdown
	}

	gate, err := fsafe.New(root)
	if err != nil {
		return nil, fmt.Errorf("plan: open boundary: %w", err)
	}
	abs, err := gate.Resolve(target)
	if err != nil {
		// fsafe sentinels pass through unwrapped for the caller's errors.Is.
		return nil, err
	}
	rel, err := relForRoot(root, abs)
	if err != nil {
		return nil, err
	}

	baseAbsent, baseHash, existing, mode, err := captureTargetBase(abs, rel)
	if err != nil {
		return nil, err
	}

	staged := normalizePage(proposed, yaml)
	result := &PlanResult{
		Path:       rel,
		BaseAbsent: baseAbsent,
		BaseHash:   baseHash,
		Staged:     staged,
		StagedHash: hashBytes(staged),
	}

	// No-op: committing would change nothing on disk, so nothing is staged.
	if !baseAbsent && bytes.Equal(existing, staged) {
		result.NoOp = true
		return result, nil
	}

	tx, err := txn.Begin(root, []txn.FileChange{{Target: rel, Data: staged, Mode: mode}})
	if err != nil {
		return nil, err
	}
	// The transaction is left staged (not committed, not aborted): the plan is a
	// durable preview a later apply consumes. No live file has been touched.
	result.TxnID = tx.ID()

	oldLabel := devNull
	if !baseAbsent {
		oldLabel = "a/" + rel
	}
	result.Diff = unifiedDiff(oldLabel, "b/"+rel, existing, staged)
	return result, nil
}

// captureTargetBase reads the plan target's pre-commit state at abs: absent (a
// new page), an existing regular file (its bytes, content hash, and perm mode),
// or a rejected non-regular file. The mode of an existing file is preserved so a
// later apply does not silently change permissions; a new page commits 0644.
func captureTargetBase(abs, rel string) (absent bool, hash string, content []byte, mode fs.FileMode, err error) {
	info, statErr := os.Lstat(abs)
	switch {
	case errors.Is(statErr, fs.ErrNotExist):
		return true, "", nil, 0o644, nil
	case statErr != nil:
		return false, "", nil, 0, fmt.Errorf("plan: stat target: %w", statErr)
	case !info.Mode().IsRegular():
		return false, "", nil, 0, fmt.Errorf("%w: %q", ErrTargetNotRegular, rel)
	}
	content, err = os.ReadFile(abs)
	if err != nil {
		return false, "", nil, 0, fmt.Errorf("plan: read target: %w", err)
	}
	return false, hashBytes(content), content, info.Mode().Perm(), nil
}

// normalizePage returns the proposed page with its frontmatter re-serialized
// through the yamladapter round-trip, preserving key order and every field —
// including ones the engine does not model — so unknown frontmatter survives the
// plan cycle (ADR-001, criterion 6). Content whose frontmatter is missing or
// unparseable is staged verbatim: normalization is a canonicalization, never a
// validation gate (validation of the staged set is a later apply concern).
func normalizePage(content []byte, yaml yamladapter.Adapter) []byte {
	fm, body, ok := splitFrontmatter(content)
	if !ok {
		return content
	}
	var m yamladapter.OrderedMap
	if err := yaml.Unmarshal(fm, &m); err != nil {
		return content
	}
	out, err := yaml.Marshal(m)
	if err != nil {
		return content
	}
	return assemblePage(out, body)
}

// splitFrontmatter separates a leading `---`-fenced YAML frontmatter block from
// the markdown body, returning the raw block bytes (delimiters removed) and the
// exact body bytes. It reports ok=false when the file does not open with a
// delimiter line or the block is never closed. Splitting and rejoining on "\n"
// is a byte-exact inverse, so the body region is preserved unchanged.
func splitFrontmatter(content []byte) (fm, body []byte, ok bool) {
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != frontmatterDelim {
		return nil, nil, false
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == frontmatterDelim {
			return []byte(strings.Join(lines[1:i], "\n")), []byte(strings.Join(lines[i+1:], "\n")), true
		}
	}
	return nil, nil, false
}

// assemblePage rebuilds a page from a marshaled frontmatter block and a body.
// The block is fenced with `---` lines; a canonical frontmatter block (which the
// marshaler terminates with a newline) reassembles to a byte-exact fixed point,
// so re-planning an already-normalized page is a no-op.
func assemblePage(fm, body []byte) []byte {
	var b bytes.Buffer
	b.WriteString(frontmatterDelim)
	b.WriteByte('\n')
	b.Write(fm)
	if len(fm) == 0 || fm[len(fm)-1] != '\n' {
		b.WriteByte('\n')
	}
	b.WriteString(frontmatterDelim)
	b.WriteByte('\n')
	b.Write(body)
	return b.Bytes()
}
