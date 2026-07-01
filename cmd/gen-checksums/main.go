// Command gen-checksums is the local developer/CI tool that produces the
// bin/SHA256SUMS manifest ADR-002 requires: it walks the per-platform binaries
// under <root>/bin, computes each file's SHA-256, and writes the manifest the
// engine's selfcheck gate consumes. It is deliberately small and offline; the
// full multi-platform release pipeline (building the five binaries) is deferred
// to a later infra issue.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/olivermorgan2/llm-wiki-kit/internal/platform"
)

// manifestName is the checksum file gen-checksums writes and selfcheck reads.
const manifestName = "SHA256SUMS"

func main() {
	root := flag.String("root", ".", "bundle root containing the bin/ directory")
	flag.Parse()
	os.Exit(gen(*root, os.Stdout, os.Stderr))
}

// gen builds the manifest for the bundle at root and writes it to
// <root>/bin/SHA256SUMS. It returns a process exit code and never writes an
// empty manifest — an empty bin/ is treated as an error so a broken build can't
// masquerade as "verified".
func gen(root string, stdout, stderr io.Writer) int {
	m, err := buildManifest(root)
	if err != nil {
		fmt.Fprintf(stderr, "gen-checksums: %v\n", err)
		return 1
	}
	if len(m) == 0 {
		fmt.Fprintf(stderr, "gen-checksums: no binaries found under %s\n", filepath.Join(root, "bin"))
		return 1
	}

	out := filepath.Join(root, "bin", manifestName)
	f, err := os.Create(out)
	if err != nil {
		fmt.Fprintf(stderr, "gen-checksums: %v\n", err)
		return 1
	}
	if err := platform.WriteManifest(f, m); err != nil {
		f.Close()
		fmt.Fprintf(stderr, "gen-checksums: %v\n", err)
		return 1
	}
	if err := f.Close(); err != nil {
		fmt.Fprintf(stderr, "gen-checksums: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "wrote %s (%d entr%s)\n", out, len(m), plural(len(m)))
	return 0
}

// buildManifest walks <root>/bin and returns a Manifest of every regular file
// (except the SHA256SUMS output itself), keyed by its slash-separated path
// relative to root so the keys match Platform.ArtifactPath on every OS.
func buildManifest(root string) (platform.Manifest, error) {
	binDir := filepath.Join(root, "bin")
	m := platform.Manifest{}
	err := filepath.WalkDir(binDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == manifestName {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		sum, err := platform.Sum(f)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		m[filepath.ToSlash(rel)] = sum
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
