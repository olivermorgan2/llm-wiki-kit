# Decision & review log — llm-wiki-kit

Chronological log of decisions and review outcomes. Lightweight; an item
that is genuinely architectural should be promoted to an ADR in
`design/adr/` via `/adr-writer`, with a back-reference here.

Format: date — decision — rationale — links.

---

### 2026-06-30 — Bootstrap from `claude-workflow-kit` v5.0.1

Bootstrapped `llm-wiki-kit` into `/Users/hermes/llm-wiki-kit` (mirrors the
existing public repo `olivermorgan2/llm-wiki-kit`) using the public
release's `bootstrap-workflow-kit` installer at pinned tag **v5.0.1**.

- **Visibility:** public (repo already existed, public, and empty; cloned
  and initialized in place — not deleted/recreated).
- **Flags used:** `--with-docs`, `--license=mit`,
  `--license-holder="Oliver Morgan"`. **Omitted `--with-ai-review`** — our
  review mechanism is Codex on the PRD, not the kit's OpenRouter PR-review
  runtime.
- **Day-one CLAUDE.md fields set:** `PROJECT_NAME=llm-wiki-kit`,
  `GITHUB_OWNER=olivermorgan2`, `GITHUB_REPO=llm-wiki-kit`,
  `DEFAULT_BRANCH=main`.

Rationale: the PRD is a finished, standard-shaped artifact, so the flow
skips `idea-to-prd` and starts at PRD normalization.

### 2026-06-30 — Start at PRD normalization; PRD placed at `design/prd.md`

The source PRD ("Claude Code Kit for OKF Wikis", updated 2026-06-20) was
copied verbatim to `design/prd.md` (the kit's expected input location) and
archived verbatim at `knowledge/sources/prd-original.md`.

### 2026-06-30 — Normalized PRD generated

Ran the `prd-normalizer` procedure against `design/prd.md`, producing the
canonical 11-field `design/prd-normalized.md`. The source PRD is rich, so
all hard-required fields resolved without invention; the only `[TBD]`
carries the pending Codex review. `design/prd.md` was left untouched per
the skill contract.

### 2026-06-30 — License = MIT, holder "Oliver Morgan"

Used the installer's `--license=mit` option (ADR-025/ADR-030 in the kit).
Note: PRD §19 still lists the *plugin/marketplace* license as an open
product decision (see [`open-questions.md`](open-questions.md) Q1) — this
decision covers the repository scaffold's LICENSE file.

### 2026-06-30 — Revision-target convention = `design/prd-addenda/`

Codex review fixes/addenda go in `design/prd-addenda/NNN-*.md` (additive,
non-destructive) rather than in-place edits to `design/prd-normalized.md`,
unless a kit command itself must regenerate normalized output. See
[`../design/prd-addenda/README.md`](../design/prd-addenda/README.md).

### 2026-06-30 — Codex adversarial PRD review — PENDING

Next gate: adversarial Codex review of `design/prd-normalized.md` (with the
source PRD for context). **Not run in this session.** Findings will be
recorded here and split into `open-questions.md` / `risks.md`, with
accepted fixes landing under `design/prd-addenda/`. Only after that loop
closes does `/prd-to-mvp` begin.
