// Package index implements deterministic index maintenance for Phase 5.
// It tracks all page files (.md/.yaml) in the repository for fast lookup
// and CI validation. No model calls are made (ADR-010).
package index

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Manifest represents the index file written to disk.
type Manifest struct {
	Version   string   `json:"version"`
	UpdatedAt string   `json:"updated_at"`
	Pages     []string `json:"pages"`
}

// Index keeps sorted paths to all page files for fast lookup.
type Index struct {
	Pages     []string
	UpdatedAt time.Time
}

// NewIndex creates a fresh empty index.
func NewIndex() *Index {
	return &Index{Pages: []string{}, UpdatedAt: time.Time{}}
}

// Update builds the index from a repo root, finding all .md and .yaml files.
func (i *Index) Update(root string) error {
	pages := []string{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return nil // skip errors (safe-filesystem approach)
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".md" || ext == ".yaml" {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			pages = append(pages, rel)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("index walk: %w", err)
	}

	sort.Strings(pages)
	i.Pages = pages
	i.UpdatedAt = time.Now().UTC()
	return nil
}

// WriteManifest writes the index manifest to the specified path.
func (i *Index) WriteManifest(path string) error {
	m := Manifest{
		Version:   "v1",
		UpdatedAt: i.UpdatedAt.Format(time.RFC3339),
		Pages:     i.Pages,
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// LoadManifest reads an index manifest from the specified path.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshal manifest: %w", err)
	}
	return &m, nil
}