// Package manifest builds and serializes the ADR-009 install version-record
// written at .llm-wiki/manifest.json. The manifest records the plugin, CLI, OKF,
// and profile versions plus, per install-managed asset, its path, ownership
// class, and content hashes — the record upgrade/uninstall (Phase 7) reconcile
// against the live tree to preserve user content and detect user-modified
// plugin-owned files.
//
// The manifest is itself plugin-owned engine metadata written through the
// ADR-006 transaction, but it never lists itself as an asset: a self-hash is a
// fixed point (hashing the manifest changes the manifest). It is still
// conflict-guarded like any other planned target and plugin-owned by
// construction — see cmd/llm-wiki/install.go. All hashing uses the same
// lowercase-hex SHA-256 contract as internal/txn, so a manifest hash matches the
// bytes install commits.
package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"

	"github.com/olivermorgan2/llm-wiki-kit/internal/txn"
)

// Path is the boundary-relative location of the version-record manifest. It
// lives inside the engine-managed .llm-wiki/ tree (ADR-005/ADR-006), never at
// the repo root or inside user content.
const Path = ".llm-wiki/manifest.json"

// SchemaVersion is the manifest schema version. It is independent of the plugin,
// CLI, and OKF versions the manifest records.
const SchemaVersion = "1"

// manifestMode is the permission applied to the committed manifest file.
const manifestMode fs.FileMode = 0o644

// Class is an asset's ADR-009 ownership class. Install writes only plugin-owned
// assets; repo-owned is recorded (and defined now) for Phase 7 upgrade/uninstall
// preservation, where install must reference existing repo content.
type Class string

const (
	// ClassPluginOwned marks an asset the plugin writes and may later replace or
	// remove.
	ClassPluginOwned Class = "plugin-owned"
	// ClassRepoOwned marks existing repo content install references but never
	// writes. Not produced by install today; defined for Phase 7.
	ClassRepoOwned Class = "repo-owned"
)

// Asset is one install-managed file recorded in the manifest. Hash is the
// current on-disk content hash at last write; LastInstalledHash is the hash the
// plugin last wrote. The two diverging is the user-modification signal
// upgrade/uninstall use (Phase 7); at install time they are equal.
type Asset struct {
	Path              string `json:"path"`
	Class             Class  `json:"class"`
	Hash              string `json:"hash"`
	LastInstalledHash string `json:"lastInstalledHash"`
}

// ProfileRecord is the ADR-007 profile reference recorded in the manifest: the
// profile id plus its pinned version.
type ProfileRecord struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// Versions is the version set an install records, assembled by the caller from
// the build version and the resolved profile.
type Versions struct {
	Plugin  string
	CLI     string
	OKF     string
	Profile ProfileRecord
}

// Manifest is the ADR-009 install version-record. Field order is the serialized
// order; the struct holds no maps, so Marshal is deterministic.
type Manifest struct {
	SchemaVersion string        `json:"schemaVersion"`
	Plugin        string        `json:"plugin"`
	CLI           string        `json:"cli"`
	OKF           string        `json:"okf"`
	Profile       ProfileRecord `json:"profile"`
	Assets        []Asset       `json:"assets"`
}

// Build records every change as a plugin-owned asset with Hash and
// LastInstalledHash both set to the SHA-256 of the change's bytes (equal at
// install time). Assets are sorted by path. The manifest never lists itself: a
// change targeting Path is skipped, since its hash cannot be known before it is
// written (self-hash fixed point). Hashes come from the in-memory change slice
// install commits, so they always match the committed bytes despite the
// date-stamped scaffold templates.
func Build(v Versions, changes []txn.FileChange) Manifest {
	assets := make([]Asset, 0, len(changes))
	for _, c := range changes {
		if c.Target == Path {
			continue
		}
		h := hashBytes(c.Data)
		assets = append(assets, Asset{
			Path:              c.Target,
			Class:             ClassPluginOwned,
			Hash:              h,
			LastInstalledHash: h,
		})
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].Path < assets[j].Path })
	return Manifest{
		SchemaVersion: SchemaVersion,
		Plugin:        v.Plugin,
		CLI:           v.CLI,
		OKF:           v.OKF,
		Profile:       v.Profile,
		Assets:        assets,
	}
}

// Marshal serializes the manifest as indented JSON with a trailing newline,
// matching the contract envelope's on-disk shape. Output is deterministic: the
// manifest holds no maps and Assets are pre-sorted by Build.
func (m Manifest) Marshal() ([]byte, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("manifest: marshal: %w", err)
	}
	return append(data, '\n'), nil
}

// Parse decodes manifest bytes. It is the read path Phase 7 upgrade/uninstall
// use; it is defined now so the round-trip is testable and stable.
func Parse(data []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("manifest: parse: %w", err)
	}
	return m, nil
}

// Change returns the txn.FileChange that commits the manifest at Path through
// the ADR-006 transaction, so install writes it in the same all-or-nothing set
// as the scaffold.
func (m Manifest) Change() (txn.FileChange, error) {
	data, err := m.Marshal()
	if err != nil {
		return txn.FileChange{}, err
	}
	return txn.FileChange{Target: Path, Data: data, Mode: manifestMode}, nil
}

// hashBytes returns the lowercase hex SHA-256 of b — the same hashing contract
// internal/txn uses for staged and base content, so a manifest hash matches the
// bytes the transaction commits.
func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
