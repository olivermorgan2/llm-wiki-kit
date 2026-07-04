package txn

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
)

// ErrTxnNotFound is returned by Load when no staged transaction with the given
// id exists under .llm-wiki/staging/ — it was never staged, was already
// committed and cleaned, or the id is not a valid transaction token.
var ErrTxnNotFound = errors.New("txn: staged transaction not found")

// Load reconstructs a staged-but-not-committed transaction from its staging dir
// under boundary, so a separate invocation (the ADR-006 page plan / apply split)
// can Commit or Abort the plan a prior Begin left on disk. The returned Txn is
// positioned exactly as the one Begin returned: Commit re-verifies every recorded
// base hash against the live tree (ErrStale on drift) before mutating, and Abort
// discards the staging. A missing, unreadable, or mis-bound manifest — or an id
// that is not a valid transaction token — is ErrTxnNotFound; no user file is
// touched on any of these paths.
func Load(boundary, id string) (*Txn, error) {
	if !txnIDRe.MatchString(id) {
		return nil, fmt.Errorf("%w: invalid id %q", ErrTxnNotFound, id)
	}
	g, err := fsafe.New(boundary)
	if err != nil {
		return nil, err
	}
	root, err := g.Resolve(".")
	if err != nil {
		return nil, err
	}
	t := &Txn{gate: g, root: root, id: id}

	// Prove the .llm-wiki/staging chain is real non-symlink dirs before reading,
	// matching Recover; an absent chain means nothing is staged for this id.
	if err := t.verifyStagingRoot(); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %q", ErrTxnNotFound, id)
		}
		return nil, err
	}

	data, err := os.ReadFile(t.abs(t.manifestRel()))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %q", ErrTxnNotFound, id)
		}
		return nil, err
	}
	m, err := decodeManifest(data)
	if err != nil {
		return nil, err
	}
	if m.Txn != id {
		// A manifest not bound to this dir cannot be trusted to drive in-boundary
		// mutations (see recoverOne); treat it as no loadable transaction.
		return nil, fmt.Errorf("%w: manifest not bound to %q", ErrTxnNotFound, id)
	}
	t.manifest = m
	return t, nil
}

// StagingDir returns the absolute path of this transaction's staging directory
// (.llm-wiki/staging/<id>/ under the canonical boundary root). A plan/apply layer
// uses it to co-locate plan-scoped sidecars (e.g. an approval record) with the
// transaction, without duplicating the staging layout.
func (t *Txn) StagingDir() string { return t.abs(t.txnDirRel()) }

// Targets returns the transaction's committed target paths in canonical
// (lexicographic) commit order — the affectedPaths an apply reports.
func (t *Txn) Targets() []string {
	paths := make([]string, len(t.manifest.Entries))
	for i, e := range t.manifest.Entries {
		paths[i] = e.Path
	}
	return paths
}
