// Package scaffold builds the deterministic change set that `llm-wiki init`
// materializes for a new wiki bundle. It produces a minimal, immediately-valid
// core-profile bundle: a root config recording the active profile (the ADR-007
// profile *reference* — never a copy of rule data), a starter page, and an
// authoring template. The change set is handed to internal/txn for an
// all-or-nothing write (ADR-006); scaffold itself performs no I/O beyond the
// lstat conflict probe.
package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
	"github.com/olivermorgan2/llm-wiki-kit/internal/txn"
)

// scaffoldMode is the permission applied to every scaffolded file.
const scaffoldMode fs.FileMode = 0o644

// timestampLayout renders the injected clock into the OKF `timestamp` field.
const timestampLayout = "2006-01-02"

// configTemplate is the bundle config written at the bundle root. It is a static
// literal — the profile reference (id + version) is interpolated from the
// resolved Profile, but no YAML library marshals it (ADR-001 confines YAML to
// the adapter, and a byte literal invokes none). `okfVersion` pins the OKF
// target version the bundle targets (design/prd.md); issue #19's version record
// will reference it. The `profile:` block is the ADR-007 reference — init copies
// no rule data.
const configTemplate = `# llm-wiki bundle configuration.
# Records the bundle format and the active validation profile. This is a
# reference to the shipped profile, not a copy of its rule data (ADR-007).
bundleFormat: 1
okfVersion: "0.1"
profile:
  id: %s
  version: "%s"
`

// indexTemplate is the starter page. Its frontmatter carries every required and
// recommended core-profile field so the fresh bundle validates with zero
// findings. Its single intra-wiki link is written bundle-root-relative
// (`wiki/templates/page-template.md`) because links resolve against the bundle
// root, not the page (internal/validate/links.go); a page-relative target would
// be flagged as a broken link.
const indexTemplate = `---
type: guide
title: Home
description: Starting page for this wiki bundle.
timestamp: %s
tags: [wiki]
aliases: [home]
resource: https://example.com/
---
# Home

Welcome to your wiki bundle. Author new pages by copying the
[page template](wiki/templates/page-template.md), keeping the frontmatter
fields filled in and using kebab-case filenames.
`

// pageTemplateTemplate is the authoring template. It is itself a real page and
// is validated like any other, so it carries fully conformant frontmatter. It
// holds no intra-wiki links, so nothing on it can resolve broken.
const pageTemplateTemplate = `---
type: concept
title: Page Template
description: Copy this file to start a new page in the wiki.
timestamp: %s
tags: [template]
aliases: [page-template]
resource: https://example.com/
---
# Page Title

Replace this heading and body with your content. Keep every frontmatter field
above present and filled in, and give new files kebab-case names.
`

// Plan returns the deterministic, lexically-sorted change set for a bundle
// scaffolded with profile p at time now. Targets are boundary-relative slash
// paths, all mode 0o644. The set is exactly the bundle config, the starter page,
// and the authoring template — no README (a frontmatter-less `.md` would fail
// validation) and no rule-data copy.
func Plan(p profile.Profile, now time.Time) []txn.FileChange {
	stamp := now.UTC().Format(timestampLayout)
	changes := []txn.FileChange{
		{
			Target: "llm-wiki.yaml",
			Data:   []byte(fmt.Sprintf(configTemplate, p.ID, p.Version)),
			Mode:   scaffoldMode,
		},
		{
			Target: "wiki/index.md",
			Data:   []byte(fmt.Sprintf(indexTemplate, stamp)),
			Mode:   scaffoldMode,
		},
		{
			Target: "wiki/templates/page-template.md",
			Data:   []byte(fmt.Sprintf(pageTemplateTemplate, stamp)),
			Mode:   scaffoldMode,
		},
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Target < changes[j].Target })
	return changes
}

// Conflicts lstats each planned target under root and returns the sorted
// slash-form targets that already exist as any filesystem entry — regular file,
// directory, or symlink. A silent overwrite is refused (ADR-009), so any
// pre-existing target is a conflict. It errors only on a real lstat failure; a
// non-existent target is simply not a conflict.
func Conflicts(root string, changes []txn.FileChange) ([]string, error) {
	var conflicts []string
	for _, c := range changes {
		p := filepath.Join(root, filepath.FromSlash(c.Target))
		if _, err := os.Lstat(p); err == nil {
			conflicts = append(conflicts, c.Target)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	sort.Strings(conflicts)
	return conflicts, nil
}
