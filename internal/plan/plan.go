// Package plan is the page addressing/load/parse/report substrate for the Phase
// 3 page commands (ADR-006). Issue #42 landed its first, read-only half — the
// `page inspect` report — and `page plan` (Plan, in pageplan.go) builds on the
// same addressing and hashing to stage a whole-page change set: it never mutates
// live files, staging the proposed content under .llm-wiki/staging/<txn-id>/ via
// the ADR-006 transaction and rendering a unified diff preview. `page apply`
// reuses this addressing and hashing rather than reimplementing them.
//
// The read-only substrate does three things: it resolves a user-supplied page path to a
// bundle-relative address through the ADR-005 fsafe boundary gate (so no path
// outside the bundle is ever read), it captures the page's content hash
// (lowercase-hex SHA-256, the base-hash input ADR-006 plan/apply binds
// against), and it reports validation findings for exactly that page by
// filtering the shared validate engine's whole-bundle run. Reusing the one
// engine keeps `page inspect` and `validate` reporting identical findings for
// the same page (criterion 15) with zero rule reimplementation.
package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/olivermorgan2/llm-wiki-kit/internal/contract"
	"github.com/olivermorgan2/llm-wiki-kit/internal/fsafe"
	"github.com/olivermorgan2/llm-wiki-kit/internal/validate"
	"github.com/olivermorgan2/llm-wiki-kit/internal/yamladapter"
)

// Sentinel addressing errors. Boundary-escape errors are not defined here:
// ResolvePage passes fsafe.ErrOutsideBoundary / fsafe.ErrSymlinkEscape through
// unwrapped so callers can errors.Is against fsafe directly and apply its
// documented system-failure mapping.
var (
	// ErrPageNotFound means the path resolves inside the boundary but no regular
	// file exists there (missing, or a directory / other non-regular target).
	ErrPageNotFound = errors.New("plan: page not found")
	// ErrNotMarkdown means the page path does not end in .md; page addresses are
	// markdown pages only.
	ErrNotMarkdown = errors.New("plan: page path must end in .md")
)

// PageRef is a resolved page address. Rel is the bundle-relative, slash-form
// path — the same form the validate engine tags its findings with, so it is the
// key page-scoped filtering matches on. Abs is the boundary-checked absolute
// path the page loads from.
type PageRef struct {
	Root string
	Rel  string
	Abs  string
}

// ResolvePage resolves arg (relative to root, or absolute-if-contained) to a
// bundle-relative page address, enforcing the ADR-005 boundary gate. It returns:
//   - fsafe.ErrOutsideBoundary / fsafe.ErrSymlinkEscape (passed through) when the
//     path escapes the boundary lexically or via a symlink;
//   - ErrNotMarkdown when the path does not end in .md;
//   - ErrPageNotFound when nothing regular exists at the resolved path.
//
// The .md check runs before touching the filesystem so a non-page path is
// rejected as such rather than as missing.
func ResolvePage(root, arg string) (PageRef, error) {
	gate, err := fsafe.New(root)
	if err != nil {
		return PageRef{}, fmt.Errorf("plan: open boundary: %w", err)
	}

	if filepath.Ext(arg) != ".md" {
		return PageRef{}, ErrNotMarkdown
	}

	abs, err := gate.Resolve(arg)
	if err != nil {
		// fsafe sentinels pass through unwrapped for the caller's errors.Is.
		return PageRef{}, err
	}

	info, err := os.Stat(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return PageRef{}, ErrPageNotFound
		}
		return PageRef{}, fmt.Errorf("plan: stat page: %w", err)
	}
	if !info.Mode().IsRegular() {
		return PageRef{}, ErrPageNotFound
	}

	rel, err := relForRoot(root, abs)
	if err != nil {
		return PageRef{}, err
	}
	return PageRef{Root: root, Rel: rel, Abs: abs}, nil
}

// relForRoot computes the bundle-relative slash-form path of abs under root,
// canonicalizing root the same way fsafe does so the relative path matches the
// validate engine's os.DirFS(root) paths even when root is given via a symlink.
func relForRoot(root, abs string) (string, error) {
	base, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("plan: canonicalize root: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(base); err == nil {
		base = resolved
	}
	rel, err := filepath.Rel(base, abs)
	if err != nil {
		return "", fmt.Errorf("plan: relativize page: %w", err)
	}
	return filepath.ToSlash(rel), nil
}

// Report is the read-only inspect result for one page: its bundle-relative
// path, whether its frontmatter parsed, the content hash, the page-scoped
// findings, and the envelope status those findings reduce to.
type Report struct {
	Path        string
	Parsed      bool
	ContentHash string
	Findings    []contract.Finding
	Status      contract.Status
}

// Inspect resolves arg, hashes the page bytes, and reports the page-scoped
// validation findings. Findings come from running the shared validate engine
// over the whole bundle (os.DirFS(root)) and filtering to ref.Rel — the
// broken-link rule needs the full bundle file set, and reusing the one engine
// keeps inspect and validate findings identical for the same page. Overrides
// apply the ADR-004 profile-override layer exactly as runValidate does. The
// page is Parsed unless it carries the never-suppressible okf-yaml-parse
// finding; malformed YAML still hashes, since the bytes exist regardless of
// parseability (ADR-006 base-hash capture).
//
// evidenceSections designates the ATX headings that open evidence contexts
// (ADR-008 sub-decision 1); it is wired identically to runValidate and page plan
// so all three agree on which links are citations (criterion 15). Empty means no
// link is a citation — the pre-#37 behavior.
func Inspect(root, arg string, yaml yamladapter.Adapter, evidenceSections []string, overrides map[string]contract.Severity) (*Report, error) {
	ref, err := ResolvePage(root, arg)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(ref.Abs)
	if err != nil {
		return nil, fmt.Errorf("plan: read page: %w", err)
	}

	// Anchor the ADR-008 repo-path resolution class identically to runValidate so
	// inspect and validate agree on in-repo `../` links (criterion 15); ok=false
	// falls back to the zero Options (pre-citation behavior). The evidence sections
	// come from the caller so inspect and validate see the same citations.
	opts, _ := validate.AnchorRepo(root)
	opts.EvidenceSections = evidenceSections
	all := validate.NewWithOptions(yaml, opts).Run(os.DirFS(root))
	all = validate.Resolve(all, overrides)

	findings := []contract.Finding{}
	parsed := true
	for _, f := range all {
		if f.Path != ref.Rel {
			continue
		}
		if f.Code == validate.CodeYAMLParse {
			parsed = false
		}
		findings = append(findings, f)
	}

	return &Report{
		Path:        ref.Rel,
		Parsed:      parsed,
		ContentHash: hashBytes(content),
		Findings:    findings,
		Status:      validate.StatusFor(findings),
	}, nil
}

// InspectContent reports the page-scoped validation findings for proposed
// content at arg without that content being on disk — the validate step of the
// authoring flow (draft → validate → plan diff → apply, issue #38). It mirrors
// Inspect exactly, with two differences: the target may be absent (a new-page
// draft is allowed; a non-.md path or a boundary escape is still rejected as in
// ResolvePage), and the shared validate engine runs over an overlay of
// os.DirFS(root) that serves the proposed bytes at the target and injects the
// draft into the bundle file set. Running the one engine over the whole bundle
// keeps content-inspect, live inspect, and validate reporting identical findings
// for the same content (criterion 15) — broken-link and citation membership see
// the draft as part of the bundle.
//
// ContentHash is the SHA-256 of the proposed bytes exactly as read
// (pre-normalization; frontmatter canonicalization stays page plan's job). A
// draft whose frontmatter fails to parse is still hashed and reported unparsed,
// exactly as Inspect treats a malformed live page. evidenceSections and overrides
// are wired identically to Inspect.
func InspectContent(root, arg string, proposed []byte, yaml yamladapter.Adapter, evidenceSections []string, overrides map[string]contract.Severity) (*Report, error) {
	ref, err := resolveContentTarget(root, arg)
	if err != nil {
		return nil, err
	}

	// Anchor the ADR-008 repo-path resolution class identically to Inspect/runValidate
	// so all surfaces agree on in-repo `../` links (criterion 15); the evidence
	// sections come from the caller so they see the same citations.
	opts, _ := validate.AnchorRepo(root)
	opts.EvidenceSections = evidenceSections
	fsys := newOverlayFS(root, ref.Rel, proposed)
	all := validate.NewWithOptions(yaml, opts).Run(fsys)
	all = validate.Resolve(all, overrides)

	findings := []contract.Finding{}
	parsed := true
	for _, f := range all {
		if f.Path != ref.Rel {
			continue
		}
		if f.Code == validate.CodeYAMLParse {
			parsed = false
		}
		findings = append(findings, f)
	}

	return &Report{
		Path:        ref.Rel,
		Parsed:      parsed,
		ContentHash: hashBytes(proposed),
		Findings:    findings,
		Status:      validate.StatusFor(findings),
	}, nil
}

// resolveContentTarget resolves arg to a bundle-relative page address for
// content-inspect: it enforces the ADR-005 boundary gate and the .md check
// exactly as ResolvePage does, but permits an absent target (a new-page draft).
// An existing non-regular target (a directory or other special file) is rejected
// as ErrPageNotFound — nothing regular lives there to inspect. The .md check runs
// before touching the filesystem, matching ResolvePage's ordering.
func resolveContentTarget(root, arg string) (PageRef, error) {
	gate, err := fsafe.New(root)
	if err != nil {
		return PageRef{}, fmt.Errorf("plan: open boundary: %w", err)
	}
	if filepath.Ext(arg) != ".md" {
		return PageRef{}, ErrNotMarkdown
	}
	abs, err := gate.Resolve(arg)
	if err != nil {
		// fsafe sentinels pass through unwrapped for the caller's errors.Is.
		return PageRef{}, err
	}
	info, err := os.Lstat(abs)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// A new-page draft: the target does not exist yet, which is allowed here.
	case err != nil:
		return PageRef{}, fmt.Errorf("plan: stat page: %w", err)
	case !info.Mode().IsRegular():
		return PageRef{}, ErrPageNotFound
	}
	rel, err := relForRoot(root, abs)
	if err != nil {
		return PageRef{}, err
	}
	return PageRef{Root: root, Rel: rel, Abs: abs}, nil
}

// hashBytes returns the lowercase hex SHA-256 of b. This is the single hashing
// contract shared across llm-wiki (staged postimages, base records, manifest
// entries); the per-package duplication is deliberate repo precedent so each
// package owns its hash contract without a cross-cutting import.
func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
