package platform

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Manifest maps a bundle-relative artifact path to its expected lowercase-hex
// SHA-256 digest. It is the parsed form of the SHA256SUMS file CI ships
// alongside the per-platform binaries.
type Manifest map[string]string

// hashLen is the hex length of a SHA-256 digest.
const hashLen = 64

// Sum returns the lowercase-hex SHA-256 digest of everything read from r. It is
// the one hashing routine shared by manifest generation and verification, so
// both sides agree on algorithm and encoding.
func Sum(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ParseManifest reads a GNU coreutils sha256sum-style document: each non-blank
// line is "<64-hex-digest><space><type><path>", where type is ' ' (text) or
// '*' (binary). Digests are normalised to lowercase. Any malformed line is a
// hard error so a corrupt manifest can never silently drop an entry.
func ParseManifest(r io.Reader) (Manifest, error) {
	m := Manifest{}
	sc := bufio.NewScanner(r)
	line := 0
	for sc.Scan() {
		line++
		raw := strings.TrimRight(sc.Text(), "\r")
		if strings.TrimSpace(raw) == "" {
			continue
		}
		// Need digest (64) + separator (1) + type (1) + at least one path char.
		if len(raw) < hashLen+3 || raw[hashLen] != ' ' {
			return nil, fmt.Errorf("manifest line %d: malformed entry: %q", line, raw)
		}
		flag := raw[hashLen+1]
		if flag != ' ' && flag != '*' {
			return nil, fmt.Errorf("manifest line %d: bad checksum-type indicator %q", line, string(flag))
		}
		digest := strings.ToLower(raw[:hashLen])
		if !isHex(digest) {
			return nil, fmt.Errorf("manifest line %d: digest is not 64 hex chars: %q", line, raw[:hashLen])
		}
		path := raw[hashLen+2:]
		if path == "" {
			return nil, fmt.Errorf("manifest line %d: missing path", line)
		}
		m[path] = digest
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

// WriteManifest writes m in canonical, deterministic form: entries sorted by
// path, text mode ("<digest>  <path>"), one per line. Deterministic output
// keeps CI-generated manifests diffable and reproducible.
func WriteManifest(w io.Writer, m Manifest) error {
	paths := make([]string, 0, len(m))
	for p := range m {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	bw := bufio.NewWriter(w)
	for _, p := range paths {
		if _, err := fmt.Fprintf(bw, "%s  %s\n", m[p], p); err != nil {
			return err
		}
	}
	return bw.Flush()
}

// isHex reports whether s is entirely lowercase hexadecimal digits.
func isHex(s string) bool {
	if len(s) != hashLen {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
