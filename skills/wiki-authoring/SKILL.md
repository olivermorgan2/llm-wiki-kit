---
name: wiki-authoring
description: Author a new core-profile page in an llm-wiki bundle — draft it, validate it, preview the diff, and commit it through the engine. Use when adding or editing a page in a repository that contains an llm-wiki.yaml bundle and the llm-wiki binary. Every sourced claim gets a resolvable citation; the diff is always shown before any write; all mutation goes through llm-wiki page apply.
---

# Authoring an llm-wiki page

This skill drafts a wiki page and lands it through the `llm-wiki` engine. It is a
**thin adapter**: it holds no validation, hashing, or mutation logic of its own —
every decision is made by an engine command, and you relay the result. Never
write inside the bundle by hand; the engine is the only thing that mutates it.

The flow is: **draft → validate → plan (diff) → apply → confirm.**

## Hard rules

- **All mutation goes through `llm-wiki page apply`.** You never create, edit, or
  delete a file inside the bundle directly — not the page, not anything.
- **Show the diff before any write, always.** `page plan` produces the diff; the
  human sees it before you run `page apply`.
- **Every sourced claim carries a resolvable citation.** Put sourced claims in an
  evidence context and give each an ordinary inline-link citation. Never invent a
  citation: if a claim cannot be cited, rephrase it as inference or drop it.
- **Exit codes drive control flow** (frozen contract; branch on them, do not
  parse prose): `0` ok · `1` ok with warnings · `2` findings (fix the draft) ·
  `3` approval required (a reported citation loss; stop and ask) · `4` re-plan
  (stale plan or bad invocation) · `5` stop and report (system/boundary failure).

## Steps

### 1. Locate the bundle

Find the bundle root — the directory containing `llm-wiki.yaml`. Read
`wiki/templates/page-template.md` there as your drafting base if it exists.

### 2. Draft outside the bundle

Compose the page in a scratch file **outside the repository**, or pipe it on
stdin — never write a draft inside the bundle. Give the page valid core-profile
frontmatter (`type`, `title`, `description`, `timestamp`, `tags`, `aliases`,
`resource`) and a kebab-case filename.

Put every sourced claim under an **evidence context** — an `## Evidence` heading —
and cite it with an inline link to a resolvable target (an `https://…` URL or an
existing in-bundle page). Designate the evidence heading with the environment
variable `LLM_WIKI_EVIDENCE_SECTIONS` (interim channel until profile vocabulary
lands), e.g. `LLM_WIKI_EVIDENCE_SECTIONS=Evidence`.

Worked example (`wiki/photosynthesis.md`):

```markdown
---
type: concept
title: Photosynthesis
description: How plants convert light into chemical energy.
timestamp: 2026-07-04
tags:
  - biology
aliases:
  - light-reactions
resource: https://example.com/photosynthesis
---

# Photosynthesis

Plants convert light energy into chemical energy stored in sugars.

## Evidence

An overview lives in the bundle index at [overview](wiki/index.md).
Further mechanism detail comes from an [external source](https://example.com/calvin-cycle).
```

### 3. Validate the draft

```
LLM_WIKI_EVIDENCE_SECTIONS=Evidence \
  llm-wiki page inspect wiki/photosynthesis.md --content <draft-file> --json
```

`page inspect --content` validates the proposed bytes in full bundle context
without staging anything (the target need not exist yet). Iterate the draft until
there are **no error-severity findings** and **no `core-citation-*` findings** for
your sourced claims. A `core-citation-malformed` / `core-citation-unresolved`
finding means a claim's citation does not resolve — fix the target, rephrase the
claim as inference, or drop it. Re-run inspect freely; it never writes.

### 4. Stage and preview the diff

```
LLM_WIKI_EVIDENCE_SECTIONS=Evidence \
  llm-wiki page plan wiki/photosynthesis.md --content <draft-file> --json
```

- If `plan.noOp` is `true`, the page is already up to date — **report "already up
  to date" and stop** (no duplicate is created).
- Otherwise, **show `plan.diff` verbatim to the human before doing anything
  else.** Note the `plan.transaction` id — apply consumes it.
- If the envelope carries an `approval` block, the edit drops an existing
  citation. Surface exactly what is lost and get explicit human confirmation
  before proceeding.

### 5. Apply

After the human has seen the diff and confirmed:

```
llm-wiki page apply <transaction> --json
```

- Pass `--approve` **only** when the human explicitly approved a reported
  citation loss from step 4.
- Exit `3` (approval required): a citation loss you have not approved — stop and
  ask, do not add `--approve` on your own.
- Exit `4` (stale plan): the live tree moved since the plan. Re-run step 4 to
  re-plan, show the fresh diff, and apply the new transaction.

### 6. Confirm

```
llm-wiki validate <bundle-root> --json
```

Confirm the committed page validates clean. Report the committed path to the
human.

## Non-goals

This skill authors a single core-profile page. It does not enrich existing pages,
run hooks, apply academic-research profile rules, or package/install anything —
those are out of scope.
