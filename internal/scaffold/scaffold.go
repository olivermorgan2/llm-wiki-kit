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
	"strings"
	"time"

	"github.com/olivermorgan2/llm-wiki-kit/internal/profile"
	"github.com/olivermorgan2/llm-wiki-kit/internal/txn"
)

// scaffoldMode is the permission applied to every scaffolded file.
const scaffoldMode fs.FileMode = 0o644

// timestampLayout renders the injected clock into the OKF `timestamp` field.
const timestampLayout = "2006-01-02"

// OKFVersion pins the OKF target version a scaffolded bundle targets
// (design/prd.md). It is interpolated into the bundle config's `okfVersion`
// field and single-sources the value issue #19's version record (ADR-009)
// records in the install manifest, so the manifest's `okf` field and the
// written config never drift.
const OKFVersion = "0.1"

// configTemplate is the bundle config written at the bundle root. It is a static
// literal — the profile reference (id + version) and OKFVersion are interpolated,
// but no YAML library marshals it (ADR-001 confines YAML to the adapter, and a
// byte literal invokes none). The `profile:` block is the ADR-007 reference —
// init copies no rule data.
const configTemplate = `# llm-wiki bundle configuration.
# Records the bundle format and the active validation profile. This is a
# reference to the shipped profile, not a copy of its rule data (ADR-007).
bundleFormat: 1
okfVersion: "%s"
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
// paths, all mode 0o644. Every scaffold carries the bundle config (the ADR-007
// profile reference — never a rule-data copy) and a home page. A profile that
// declares per-type rules additionally gets one valid authoring template per
// profiled type, generated from the profile data (so scaffold stays generic — it
// hardcodes no profile); a typeless profile (core) keeps the single generic
// page-template, byte-identical to the pre-Phase-4 scaffold. No README (a
// frontmatter-less `.md` would fail validation).
func Plan(p profile.Profile, now time.Time) []txn.FileChange {
	stamp := now.UTC().Format(timestampLayout)
	changes := []txn.FileChange{{
		Target: "llm-wiki.yaml",
		Data:   []byte(fmt.Sprintf(configTemplate, OKFVersion, p.ID, p.Version)),
		Mode:   scaffoldMode,
	}}

	if len(p.Types) == 0 {
		// Core / typeless profile: the pre-Phase-4 scaffold, byte-identical.
		changes = append(changes,
			txn.FileChange{Target: "wiki/index.md", Data: []byte(fmt.Sprintf(indexTemplate, stamp)), Mode: scaffoldMode},
			txn.FileChange{Target: "wiki/templates/page-template.md", Data: []byte(fmt.Sprintf(pageTemplateTemplate, stamp)), Mode: scaffoldMode},
		)
	} else {
		changes = append(changes, txn.FileChange{
			Target: "wiki/index.md",
			Data:   profileIndex(p, stamp),
			Mode:   scaffoldMode,
		})
		for _, typeName := range sortedTypeNames(p) {
			changes = append(changes, txn.FileChange{
				Target: "wiki/templates/" + typeName + ".md",
				Data:   typeTemplate(typeName, p.Types[typeName], stamp),
				Mode:   scaffoldMode,
			})
		}
	}

	sort.Slice(changes, func(i, j int) bool { return changes[i].Target < changes[j].Target })
	return changes
}

// sortedTypeNames returns p's profiled type names in lexical order for a
// deterministic change set.
func sortedTypeNames(p profile.Profile) []string {
	names := make([]string, 0, len(p.Types))
	for name := range p.Types {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// profileIndex renders a home page linking to each per-type template. It is an
// unprofiled `guide` page (validated by core rules only); its links resolve
// against the templates the same Plan materializes.
func profileIndex(p profile.Profile, stamp string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "---\ntype: guide\ntitle: Home\ndescription: Starting page for this %s wiki bundle.\n", p.ID)
	fmt.Fprintf(&b, "timestamp: %s\ntags: [wiki]\naliases: [home]\nresource: https://example.com/\n---\n", stamp)
	fmt.Fprintf(&b, "# Home\n\nWelcome to your %s wiki bundle. Author a new page by copying the\ntemplate for its type and replacing the placeholder content:\n\n", p.ID)
	for _, name := range sortedTypeNames(p) {
		fmt.Fprintf(&b, "- [%s](wiki/templates/%s.md)\n", name, name)
	}
	return []byte(b.String())
}

// typeTemplate renders one valid authoring template for a profiled type from its
// rule data: core required/recommended fields, each profile-added required field
// with a valid placeholder value (a list of the required length for a list-min
// field; a non-obligation-triggering enum member for an enum field), one member
// of each recommended any-of group (so no suggestion fires), and every required
// section as an empty heading. The result validates with zero findings under the
// profile — proved by TestAcademicScaffoldValidatesWithZeroFindings.
func typeTemplate(typeName string, rules profile.TypeRules, stamp string) []byte {
	title := capitalize(typeName)
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "type: %s\n", typeName)
	fmt.Fprintf(&b, "title: %s template\n", title)
	fmt.Fprintf(&b, "description: A starter %s page. Replace this content.\n", typeName)
	fmt.Fprintf(&b, "timestamp: %s\ntags: [template]\naliases: [%s-template]\nresource: https://example.com/\n", stamp, typeName)
	for _, field := range rules.Required {
		fmt.Fprintf(&b, "%s: %s\n", field, placeholderValue(field, rules))
	}
	for _, group := range rules.RecommendedAnyOf {
		if len(group) > 0 {
			fmt.Fprintf(&b, "%s: REPLACE\n", group[0])
		}
	}
	b.WriteString("---\n")
	fmt.Fprintf(&b, "# %s Title\n\nReplace this heading and body with your content.\n", title)
	for _, sec := range rules.RequiredSections {
		fmt.Fprintf(&b, "\n## %s\n\nReplace this section.\n", sec)
	}
	return []byte(b.String())
}

// placeholderValue renders a valid YAML value for a profile-added required field:
// a list of the required minimum length for a list-min field, a valid enum member
// that does not trigger a citation obligation for an enum field, else a scalar
// placeholder.
func placeholderValue(field string, rules profile.TypeRules) string {
	if min, ok := rules.ListMin[field]; ok {
		n := max(min, 1)
		items := make([]string, n)
		for i := range items {
			items[i] = "Replace Me"
		}
		return "[" + strings.Join(items, ", ") + "]"
	}
	if vals, ok := rules.Enums[field]; ok && len(vals) > 0 {
		return nonTriggeringEnum(field, vals, rules)
	}
	return "REPLACE"
}

// nonTriggeringEnum picks a valid enum member for field that does not satisfy a
// citation.requireWhen trigger on that field (so a generated `claim` template is
// not obliged to carry a citation). Falls back to the first member.
func nonTriggeringEnum(field string, vals []string, rules profile.TypeRules) string {
	trigger := ""
	if rules.Citation != nil {
		trigger = rules.Citation.RequireWhen[field]
	}
	for _, v := range vals {
		if v != trigger {
			return v
		}
	}
	return vals[0]
}

// capitalize upper-cases the first byte of s (ASCII type names only).
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
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
