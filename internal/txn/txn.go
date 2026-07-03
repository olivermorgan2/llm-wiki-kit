// Package txn implements the ADR-006 cross-file transaction model on top of the
// ADR-005 filesystem-safety gate (internal/fsafe). A mutation stages its entire
// change set under .llm-wiki/staging/<txn-id>/, records a staging manifest
// binding each target to its staged-content hash plus the pre-commit base hash,
// validates the whole set, then commits by an ordered, journaled sequence of
// per-file atomic renames. Interruption is recovered from the journal: a partial
// commit is rolled forward (finish the remaining renames from staging) when the
// staged postimages are intact, or rolled back to the byte-and-metadata-identical
// pre-commit tree from durable preimage records. This makes multi-file mutation
// all-or-nothing at one auditable chokepoint (acceptance criteria 3, 20).
//
// Every mutating byte routes through the fsafe.Gate — WriteFile for staged and
// committed content and preimage restore, Remove for absent-preimage rollback —
// so the ADR-005 "zero out-of-boundary writes" guarantee is unaffected; this
// package adds no new write path.
//
// Caveats (accepted MVP trade-offs):
//   - Double-write commit: committing a step re-reads and re-writes the staged
//     postimage through the gate (there is no rename primitive on the gate), so a
//     step's staged bytes stay intact for roll-forward and every step is
//     idempotent, at the cost of writing each payload twice.
//   - No cross-process locking: concurrent transactions on the same boundary are
//     not serialized here; the consumer (install/init) runs one at a time.
//   - Single filesystem: the atomic-rename assumption is inherited from fsafe —
//     the boundary must live on one filesystem.
package txn

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
)

// internalMode is the permission applied to every engine-internal staging file
// (staged postimages, manifest, preimage blobs, journal markers).
const internalMode fs.FileMode = 0o644

// Exported sentinel errors. fsafe sentinels (ErrOutsideBoundary, ErrSymlinkEscape)
// pass through wrapped and remain matchable with errors.Is.
var (
	// ErrEmptyChangeSet is returned by Begin when the change set is empty.
	ErrEmptyChangeSet = errors.New("txn: empty change set")
	// ErrDuplicateTarget is returned by Begin when two changes clean to the
	// same target path.
	ErrDuplicateTarget = errors.New("txn: duplicate target in change set")
	// ErrReservedPath is returned by Begin when a target falls under a reserved
	// engine path (.llm-wiki/staging/ or .llm-wiki/tmp/).
	ErrReservedPath = errors.New("txn: target under reserved engine path")
	// ErrNonRegularTarget is returned when a target resolves to a non-regular
	// file (symlink, directory, device, socket), at plan time or commit
	// revalidation.
	ErrNonRegularTarget = errors.New("txn: target is not a regular file")
	// ErrStale is returned by Commit when a live target or staged postimage no
	// longer matches what Begin recorded; the tree is left unmutated and the
	// transaction is still Abort-able.
	ErrStale = errors.New("txn: staged plan is stale")
	// ErrTxnDone is returned by Commit/Abort after the transaction has already
	// been committed or aborted.
	ErrTxnDone = errors.New("txn: transaction already committed or aborted")
)

// errPostimageDamaged signals that a staged postimage no longer matches its
// recorded hash, so roll-forward is impossible and the recovery driver must roll
// back instead. It is internal to the package.
var errPostimageDamaged = errors.New("txn: staged postimage damaged")

// txnIDRe bounds a transaction id to 16 lowercase hex chars. removeTxnDir refuses
// to RemoveAll a directory whose name does not match, confining recursive delete
// to engine-owned staging dirs.
var txnIDRe = regexp.MustCompile(`^[0-9a-f]{16}$`)

// stagingRootRel is the boundary-relative staging root (.llm-wiki/staging).
var stagingRootRel = filepath.Join(fsafe.StagingDir, "staging")

// Reserved target prefixes (slash form). Targets elsewhere under .llm-wiki (e.g.
// .llm-wiki/manifest.json, which ADR-009's install writes through a transaction)
// are allowed; only these two engine subtrees are refused.
var (
	stagingReservedPrefix = filepath.ToSlash(stagingRootRel) + "/"
	tmpReservedPrefix     = filepath.ToSlash(filepath.Join(fsafe.StagingDir, "tmp")) + "/"
)

// FileChange is one target mutation in a transaction: the boundary-relative
// target path, the exact bytes to commit, and the mode to set on the committed
// file. A nil/empty Data commits an empty file (distinct from not staging the
// target at all).
type FileChange struct {
	Target string
	Data   []byte
	Mode   fs.FileMode
}

// Txn is a staged, not-yet-committed transaction. It is not safe for concurrent
// use. Once Commit or Abort returns, the Txn is done and further calls yield
// ErrTxnDone.
type Txn struct {
	gate     fsafe.Gate
	root     string // canonical boundary root (gate.Resolve("."))
	id       string // 16 hex chars
	manifest manifest
	done     bool

	// stepHook, when non-nil, is invoked before each mutating disk op during
	// commit, rollback, and cleanup. A non-nil return simulates a crash: the
	// operation returns immediately with no compensation, leaving the disk
	// exactly as a kill at that instant would. Test-only; production callers
	// leave it nil (Recover passes nil to recoverWithHook).
	stepHook func(stepOp) error
}

// plannedChange couples a manifest entry with the staged bytes to write for it,
// so the two stay aligned through the sort that fixes commit order.
type plannedChange struct {
	entry manifestEntry
	data  []byte
}

// Begin stages changes under a fresh transaction dir and validates the whole set
// at plan time, writing the manifest last so its presence means the change set is
// fully staged. It mutates nothing in the user tree. A rejected change set
// (escape, symlink component, non-regular target, duplicate, reserved path,
// empty set) returns before any staging is written.
func Begin(boundary string, changes []FileChange) (*Txn, error) {
	if len(changes) == 0 {
		return nil, ErrEmptyChangeSet
	}
	g, err := fsafe.New(boundary)
	if err != nil {
		return nil, err
	}
	root, err := g.Resolve(".")
	if err != nil {
		return nil, err
	}
	id, err := newTxnID()
	if err != nil {
		return nil, err
	}
	t := &Txn{gate: g, root: root, id: id}

	planned, err := t.planChanges(changes)
	if err != nil {
		return nil, err
	}

	entries := make([]manifestEntry, len(planned))
	for i, p := range planned {
		if err := g.WriteFile(t.stagedRel(i), p.data, internalMode); err != nil {
			t.bestEffortRemoveDir()
			return nil, err
		}
		entries[i] = p.entry
	}

	m := manifest{Version: manifestVersion, Txn: id, Entries: entries}
	data, err := marshalManifest(m)
	if err != nil {
		t.bestEffortRemoveDir()
		return nil, err
	}
	if err := g.WriteFile(t.manifestRel(), data, internalMode); err != nil {
		t.bestEffortRemoveDir()
		return nil, err
	}
	t.manifest = m
	return t, nil
}

// Abort discards a staged transaction: the staging dir is removed and the user
// tree is left exactly as it was (no commit has occurred). Further Commit/Abort
// calls return ErrTxnDone.
func (t *Txn) Abort() error {
	if t.done {
		return ErrTxnDone
	}
	t.done = true
	return t.cleanup()
}

// planChanges validates every change and returns the planned changes in
// canonical (lexicographic) commit order.
func (t *Txn) planChanges(changes []FileChange) ([]plannedChange, error) {
	seen := make(map[string]bool, len(changes))
	planned := make([]plannedChange, 0, len(changes))
	for _, c := range changes {
		cleaned := cleanTarget(c.Target)
		if seen[cleaned] {
			return nil, fmt.Errorf("%w: %q", ErrDuplicateTarget, cleaned)
		}
		seen[cleaned] = true

		if isReservedTarget(cleaned) {
			return nil, fmt.Errorf("%w: %q", ErrReservedPath, cleaned)
		}

		// Boundary + symlink guard: Resolve rejects escapes; any in-boundary
		// symlink component makes the resolved path differ from the lexical
		// join, which the commit's fd-relative gate would refuse mid-commit, so
		// reject it now at plan time.
		safe, err := t.gate.Resolve(c.Target)
		if err != nil {
			return nil, err
		}
		lexical := filepath.Join(t.root, filepath.FromSlash(cleaned))
		if safe != lexical {
			// A symlink was traversed. If the final component itself is the
			// symlink it is a non-regular target (ADR-006 rejects those); an
			// intermediate symlink dir component is an escape-class rejection.
			if info, lerr := os.Lstat(lexical); lerr == nil && info.Mode()&os.ModeSymlink != 0 {
				return nil, fmt.Errorf("%w: %q", ErrNonRegularTarget, cleaned)
			}
			return nil, fmt.Errorf("%w: %q traverses a symlink component", fsafe.ErrSymlinkEscape, cleaned)
		}

		base, err := captureBase(lexical, cleaned)
		if err != nil {
			return nil, err
		}
		planned = append(planned, plannedChange{
			entry: manifestEntry{
				Path:   cleaned,
				Mode:   c.Mode.Perm(),
				Staged: hashBytes(c.Data),
				Base:   base,
			},
			data: c.Data,
		})
	}
	sort.Slice(planned, func(i, j int) bool { return planned[i].entry.Path < planned[j].entry.Path })
	return planned, nil
}

// captureBase records the pre-commit state of the target at lexical: absent, an
// existing regular file (bytes hashed + perm mode), or a rejected non-regular
// file. An empty existing file hashes empty content, never the absent sentinel.
func captureBase(lexical, cleaned string) (baseRecord, error) {
	info, err := os.Lstat(lexical)
	if errors.Is(err, fs.ErrNotExist) {
		return absentBase(), nil
	}
	if err != nil {
		return baseRecord{}, err
	}
	if !info.Mode().IsRegular() {
		return baseRecord{}, fmt.Errorf("%w: %q", ErrNonRegularTarget, cleaned)
	}
	b, err := os.ReadFile(lexical)
	if err != nil {
		return baseRecord{}, err
	}
	return regularBase(hashBytes(b), info.Mode()), nil
}

// isReservedTarget reports whether a cleaned slash-form target lies under a
// reserved engine subtree.
func isReservedTarget(cleaned string) bool {
	for _, prefix := range []string{stagingReservedPrefix, tmpReservedPrefix} {
		if cleaned == strings.TrimSuffix(prefix, "/") || strings.HasPrefix(cleaned, prefix) {
			return true
		}
	}
	return false
}

// --- staging layout helpers ---

func (t *Txn) txnDirRel() string          { return filepath.Join(stagingRootRel, t.id) }
func (t *Txn) manifestRel() string        { return filepath.Join(t.txnDirRel(), "manifest.json") }
func (t *Txn) stagedRel(i int) string     { return filepath.Join(t.txnDirRel(), "files", pad4(i)) }
func (t *Txn) preimageSetRel() string     { return filepath.Join(t.txnDirRel(), "preimages.json") }
func (t *Txn) preimageBlobRel(s string) string {
	return filepath.Join(t.txnDirRel(), "preimages", s)
}
func (t *Txn) journalRel(name string) string { return filepath.Join(t.txnDirRel(), "journal", name) }

// abs joins a boundary-relative staging path onto the canonical root for os
// reads (reads are outside the ADR-005 write gate).
func (t *Txn) abs(rel string) string { return filepath.Join(t.root, rel) }

// cleanup removes the transaction's staging area, deleting the manifest FIRST so
// a crash mid-RemoveAll leaves a provably-terminal (manifest-less) dir that
// Recover garbage-collects, never a dir that could be misread as in-flight.
func (t *Txn) cleanup() error {
	if err := t.checkpoint(stepOp{kind: opCleanupManifest}); err != nil {
		return err
	}
	if err := t.gate.Remove(t.manifestRel()); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := t.checkpoint(stepOp{kind: opCleanupRemoveAll}); err != nil {
		return err
	}
	return t.removeTxnDir()
}

// removeTxnDir os.RemoveAll's the transaction dir after confirming its id is a
// valid 16-hex token and it is not a symlink. The area is engine-owned and
// RemoveAll does not follow symlinks, so no recursive-delete gate primitive is
// needed.
func (t *Txn) removeTxnDir() error {
	if !txnIDRe.MatchString(t.id) {
		return fmt.Errorf("txn: refusing to remove staging dir with invalid id %q", t.id)
	}
	dir := t.abs(t.txnDirRel())
	info, err := os.Lstat(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("txn: staging dir %q is a symlink", dir)
	}
	return os.RemoveAll(dir)
}

// bestEffortRemoveDir cleans a partially-staged dir on a Begin error path,
// ignoring failures (a manifest-less remnant is terminal garbage Recover will
// collect regardless).
func (t *Txn) bestEffortRemoveDir() { _ = t.removeTxnDir() }

func newTxnID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("txn: generate id: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

func pad4(i int) string { return fmt.Sprintf("%04d", i) }
