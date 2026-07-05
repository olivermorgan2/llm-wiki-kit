// Package enrichment implements Phase 5's page enrichment functionality.
// It runs inspect/plan/apply on an existing page and returns validator findings.
package enrichment

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/validate"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// EnrichPage runs inspect/plan/apply on an existing page and returns validator findings
// from the inspect step. The function is deterministic and performs no model calls.
// It uses the shared validate engine to ensure findings are identical across all surfaces
// (criterion 15).
func EnrichPage(root, pagePath string, yamlAdapter yamladapter.Adapter, evidenceSections []string) ([]contract.Finding, error) {
	// Resolve absolute paths
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}
	absPagePath, err := filepath.Abs(pagePath)
	if err != nil {
		return nil, fmt.Errorf("resolve page path: %w", err)
	}

	// Ensure page is within root
	relPath, err := filepath.Rel(absRoot, absPagePath)
	if err != nil {
		return nil, fmt.Errorf("relative path: %w", err)
	}
	if relPath == "." || relPath == ".." || (len(relPath) > 0 && relPath[0] == '.') {
		return nil, fmt.Errorf("page path must be within root")
	}

	// Read page content
	content, err := os.ReadFile(absPagePath)
	if err != nil {
		return nil, fmt.Errorf("read page: %w", err)
	}

	// Create an in-memory filesystem with just this page
	inMemFS := &inMemoryFS{}
	inMemFS.files = make(map[string][]byte)
	inMemFS.files[relPath] = content

	// Run validation on the page to get findings
	// Use the shared validate engine for deterministic results
	v := validate.New(yamlAdapter)
	
	// Run validation
	findings := v.Run(inMemFS)
	return findings, nil
}

// inMemoryFS implements fs.FS for in-memory filesystems used for validation.
type inMemoryFS struct {
	files map[string][]byte
}

func (fsys *inMemoryFS) Open(name string) (fs.File, error) {
	if content, ok := fsys.files[name]; ok {
		return &inMemoryFile{bytes: content, name: name}, nil
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

type inMemoryFile struct {
	bytes []byte
	name  string
	pos   int
}

func (f *inMemoryFile) Read(p []byte) (int, error) {
	if f.pos >= len(f.bytes) {
		return 0, io.EOF
	}
	n := copy(p, f.bytes[f.pos:])
	f.pos += n
	return n, nil
}

func (f *inMemoryFile) Seek(offset int64, whence int) (int64, error) {
	newPos := int(offset)
	switch whence {
	case io.SeekStart:
		f.pos = newPos
	case io.SeekCurrent:
		f.pos += newPos
	case io.SeekEnd:
		f.pos = len(f.bytes) + newPos
	}
	return int64(f.pos), nil
}

func (f *inMemoryFile) Close() error {
	return nil
}

func (f *inMemoryFile) Stat() (fs.FileInfo, error) {
	return &inMemoryFileInfo{name: f.name, size: int64(len(f.bytes))}, nil
}

type inMemoryFileInfo struct {
	name string
	size int64
}

func (i *inMemoryFileInfo) Name() string       { return i.name }
func (i *inMemoryFileInfo) Size() int64        { return i.size }
func (i *inMemoryFileInfo) Mode() fs.FileMode  { return 0644 }
func (i *inMemoryFileInfo) ModTime() time.Time { return time.Time{} }
func (i *inMemoryFileInfo) IsDir() bool        { return false }
func (i *inMemoryFileInfo) Sys() interface{}   { return nil }