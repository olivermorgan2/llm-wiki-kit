# Workflow-kit dogfood notes

Running notes on how well `claude-workflow-kit` (v5.0.1) works while
bootstrapping `llm-wiki-kit`. Concrete observations and friction so we can
feed fixes back upstream.

## 2026-06-30 — Bootstrap

**What was done:** downloaded `bootstrap-workflow-kit` (release asset,
v5.0.1), audited it, cloned the kit at the pinned tag, and ran
`bin/install-workflow-kit` directly against `/Users/hermes/llm-wiki-kit`
(an already-cloned, empty public repo) with `--with-docs --license=mit
--license-holder="Oliver Morgan" --non-interactive` and the four `--set`
placeholders.

### What worked well

- **Audit-first path is first-class.** The bootstrap script's own header
  documents the `gh release download` → inspect → run flow, so we never had
  to pipe curl into bash. Good for a user who dislikes pipe-to-bash.
- **Installer is well-instrumented.** Every copied file is logged; the
  `skip (ai-review disabled): .claude/skills/review-pr` line made it
  obvious that omitting `--with-ai-review` did the right thing.
- **Placeholder rendering is clean.** `PROJECT_NAME`, `GITHUB_OWNER`,
  `GITHUB_REPO`, `DEFAULT_BRANCH` all rendered; no `{{...}}` left in
  `CLAUDE.md`. Repo + branch landed correctly as
  `olivermorgan2/llm-wiki-kit` / `main`.
- **License option works.** `--license=mit --license-holder="Oliver
  Morgan"` produced a standard MIT `LICENSE` (`Copyright (c) 2026 Oliver
  Morgan`) — no need to hand-write one.
- **Works into a pre-existing git repo.** Target was already a clone with
  `origin` set; installer detected the repo, skipped `git init`, and made
  the initial commit without complaint.
- **`--help` is accurate.** Documented flags matched actual behavior,
  including the license flags (ADR-025/030) and `--with-docs` (ADR-010).

### Friction / things to watch

- **`_TBD_` noise in `CLAUDE.md`.** The rendered `CLAUDE.md` carries ~25
  `_TBD_` placeholders (stack, testing, code style, "Tests stay green. Run
  `_TBD_`", definitions of done). They're intentional ("unknown but
  acceptable") and clearly labeled, but a brand-new repo's load-bearing
  rules file ships with `Run _TBD_ before declaring a task done`, which is a
  weak default. Not a blocker; will fill in once the Go stack is decided.
  Worth considering whether the kit should mark these more aggressively as
  must-fill vs optional.
- **No `knowledge/` layer shipped.** Expected — the kit doesn't provide a
  curated knowledge layer; we added our own `knowledge/` per protocol. Flag
  for upstream only if we want it standardized.
- **`prd-normalizer` is a skill, not a CLI command.** There's no
  `llm-wiki`-style binary entry for normalization; it's a Claude Code skill
  under `.claude/skills/prd-normalizer/`. In this supervised session the
  slash-command UI isn't the invocation path, so we executed the skill's
  documented `SKILL.md` procedure by hand to produce
  `design/prd-normalized.md`. Output follows the canonical 11-field shape;
  source `design/prd.md` left untouched per the skill contract. Next
  interactive session can re-run `/prd-normalizer` to confirm parity if
  desired.
- **Initial commit message is generic.** `chore: install workflow kit
  (project-local)` — fine, but doesn't reference the project. Minor.

### Open follow-ups

- Fill `CLAUDE.md` `_TBD_` fields once the Go toolchain/test commands are
  decided (ties to open questions Q2/Q3).
- Run the Codex adversarial review of `design/prd-normalized.md` (separate
  session) and capture findings under `design/prd-addenda/`.
- Confirm `/prd-normalizer` (interactive) produces output consistent with
  the hand-run normalization, then proceed to `/prd-to-mvp`.
