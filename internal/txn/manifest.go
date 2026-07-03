package txn

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

// manifestVersion is the on-disk schema version for both the staging manifest
// and the preimage set. A decode of any other version is refused so an older
// engine never misreads a newer layout.
const manifestVersion = 1

// errBadBaseRecord is returned when a base record violates the absent-XOR-hash
// invariant (ADR-006: "absent" and "empty" are distinct, and a record is
// exactly one of the two representations).
var errBadBaseRecord = errors.New("txn: base record must be exactly one of absent or hashed")

// baseRecord is the pre-commit representation of a single target, used both as
// the plan-time base captured in the manifest and as the preimage descriptor.
// It is either the absent sentinel (Absent true, no hash) or a regular file
// (SHA256 + Mode, Absent false) — never both, never neither. This keeps
// "file absent" distinguishable from "file empty" at plan and recovery time.
type baseRecord struct {
	Absent bool        `json:"absent,omitempty"`
	SHA256 string      `json:"sha256,omitempty"`
	Mode   fs.FileMode `json:"mode,omitempty"`
}

// absentBase returns the absent sentinel base record.
func absentBase() baseRecord { return baseRecord{Absent: true} }

// regularBase returns a base record for an existing regular file. Callers pass
// perm-only mode bits; an empty file uses hashBytes(nil), never the sentinel.
func regularBase(sha string, mode fs.FileMode) baseRecord {
	return baseRecord{SHA256: sha, Mode: mode.Perm()}
}

// validate enforces the absent-XOR-hash invariant.
func (b baseRecord) validate() error {
	if b.Absent && b.SHA256 != "" {
		return fmt.Errorf("%w: both absent and hashed", errBadBaseRecord)
	}
	if !b.Absent && b.SHA256 == "" {
		return fmt.Errorf("%w: neither absent nor hashed", errBadBaseRecord)
	}
	return nil
}

// manifestEntry binds a cleaned, slash-normalized boundary-relative target path
// to its staged postimage hash and mode plus the base record captured at plan
// time. The entry index in a sorted manifest is the commit step index and the
// staged file name (files/%04d).
type manifestEntry struct {
	Path   string      `json:"path"`
	Mode   fs.FileMode `json:"mode"`
	Staged string      `json:"staged"`
	Base   baseRecord  `json:"base"`
}

// manifest is the staging manifest written last at Begin (its presence means
// the change set is fully staged). Entries are sorted lexicographically by
// Path, and that order is the deterministic commit order.
type manifest struct {
	Version int             `json:"version"`
	Txn     string          `json:"txn"`
	Entries []manifestEntry `json:"entries"`
}

// marshalManifest sorts entries by path (the canonical commit order) and encodes
// the manifest as indented JSON. It validates each base record so a manifest is
// never written with a malformed sentinel.
func marshalManifest(m manifest) ([]byte, error) {
	sort.Slice(m.Entries, func(i, j int) bool { return m.Entries[i].Path < m.Entries[j].Path })
	for _, e := range m.Entries {
		if err := e.Base.validate(); err != nil {
			return nil, fmt.Errorf("txn: manifest entry %q: %w", e.Path, err)
		}
	}
	return json.MarshalIndent(m, "", "  ")
}

// decodeManifest parses and validates a manifest: the version must match, and
// every base record must satisfy the absent-XOR-hash invariant. Entries are
// re-sorted so a hand-edited or reordered file still yields canonical order.
func decodeManifest(data []byte) (manifest, error) {
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return manifest{}, fmt.Errorf("txn: decode manifest: %w", err)
	}
	if m.Version != manifestVersion {
		return manifest{}, fmt.Errorf("txn: unsupported manifest version %d (want %d)", m.Version, manifestVersion)
	}
	for _, e := range m.Entries {
		if err := e.Base.validate(); err != nil {
			return manifest{}, fmt.Errorf("txn: manifest entry %q: %w", e.Path, err)
		}
	}
	sort.Slice(m.Entries, func(i, j int) bool { return m.Entries[i].Path < m.Entries[j].Path })
	return m, nil
}

// preimageSet is the barrier record written after every preimage blob is
// captured; its presence means the durable backup of the pre-commit tree is
// complete. Blobs lists the content hashes stored under preimages/ (absent
// targets contribute none). The authoritative restore data is the manifest's
// base records plus these content-addressed blobs.
type preimageSet struct {
	Version int      `json:"version"`
	Txn     string   `json:"txn"`
	Blobs   []string `json:"blobs"`
}

// marshalPreimageSet encodes the preimage barrier record.
func marshalPreimageSet(p preimageSet) ([]byte, error) {
	sort.Strings(p.Blobs)
	return json.MarshalIndent(p, "", "  ")
}

// hashBytes returns the lowercase hex SHA-256 of b. It is the single hashing
// contract shared by staged postimages, base records, and preimage blob names.
func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// cleanTarget normalizes a boundary-relative target path to its canonical,
// slash-separated form. This is the manifest key and the sort/dedup key; two
// targets that clean to the same string are duplicates.
func cleanTarget(target string) string {
	return filepath.ToSlash(filepath.Clean(target))
}
