package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexUpdate(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	
	// Create test files
	files := []string{
		"test1.md",
		"test2.yaml",
		"subdir/test3.md",
		"subdir/test4.yaml",
	}
	
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	
	// Create index and update
	idx := NewIndex()
	if err := idx.Update(tmpDir); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	
	// Verify pages found
	if len(idx.Pages) != len(files) {
		t.Errorf("expected %d pages, got %d", len(files), len(idx.Pages))
	}
	
	// Verify sorting
	for i := 1; i < len(idx.Pages); i++ {
		if idx.Pages[i] < idx.Pages[i-1] {
			t.Errorf("pages not sorted: %s < %s", idx.Pages[i-1], idx.Pages[i])
		}
	}
	
	// Verify UpdatedAt set
	if idx.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestIndexWriteManifest(t *testing.T) {
	tmpDir := t.TempDir()
	
	idx := NewIndex()
	idx.Pages = []string{"a.md", "b.yaml"}
	idx.UpdatedAt = time.Now().UTC()
	
	manifestPath := filepath.Join(tmpDir, "manifest")
	if err := idx.WriteManifest(manifestPath); err != nil {
		t.Fatalf("WriteManifest failed: %v", err)
	}
	
	// Verify manifest exists
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("manifest not written: %v", err)
	}
	
	// Verify manifest parses
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid manifest JSON: %v", err)
	}
	
	if len(m.Pages) != 2 {
		t.Errorf("expected 2 pages in manifest, got %d", len(m.Pages))
	}
}