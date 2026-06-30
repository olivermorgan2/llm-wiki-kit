# Skills Reference

A functional guide to every skill in the Claude Code Workflow Kit.
Skills are grouped by what they do in the workflow — not
alphabetically — so the order you'd encounter them on a real project
matches the order on this page.

For the action-oriented walkthrough that uses these skills, see
[`workflow-guide.md`](workflow-guide.md). For the canonical
inventory ordered by source-tree layout, see
[`../skills/README.md`](../skills/README.md). For the full
interface spec of any individual skill, follow its `SKILL.md` link.

**For agents:** the same inventory is available as structured data in
[`kit.json`](../kit.json) — skills, permission categories, inputs,
outputs, and handoffs — so you can build the workflow graph without
reading each `SKILL.md`. See [`agent-contract.md`](agent-contract.md).

Each skill has a permission category that tells Claude Code how to
treat it under auto-mode (when you've authorised the assistant to
skip per-tool approval prompts). The category appears as `cat-N` next
to the skill name throughout this document:

- **cat-1** — local, reversible work like editing files in your
  project. Auto-mode runs these without asking.
- **cat-2** — auto-mode proceeds but pauses to acknowledge a bypass
  in chat. Rare; only the executor's plan-mode gate falls here.
- **cat-3** — public or hard-to-reverse actions like creating PRs,
  pushing tags, or closing GitHub milestones. These always require
  an explicit `yes` from you, regardless of mode.

The full classification rules are in
[`workflow-guide.md` §7](workflow-guide.md#7-auto-mode-permission-contract-adr-041).

---

## Start here: the verb layer

You do not need to memorise all of the skills below. A small
**verb layer** sits in front of the inventory — reach for a verb and it
maps to the underlying skill(s):

| Verb | What you mean | Underlying skill(s) |
|---|---|---|
| `/start`, `/next` | "what do I run now?" | `start` (router) |
| `/decide` | "capture a decision" | `clarify` → `adr-writer` → `check-plan` |
| `/backlog` | "turn the plan into issues" | `issue-planner` |
| `/work` | "build the next issue" | `prepare-issue` → `claude-issue-executor` |
| `/ship` | "open the PR" | `pr-review-packager` |
| `/review-pr` | "AI-review a PR" | `review-pr` (dry-run, then approved publish) |
| `/finish-milestone` | "close out the milestone" | `audit-milestone` → `milestone-summary` → `complete-milestone` |
| `/feature` | "plan a major feature update" | `feature-prd` |
| `/release` | "cut a release" | `release` |
| `/resume`, `/pause` | "where were we?" / "save state" | the same-named skills |

New to the kit? Run `/start` — it inspects your project and tells you
which step to run next. The full reference below stays the source of
truth for each skill's exact interface, and agents read exact names from
[`kit.json`](../kit.json). The verb layer, the three **operating modes**
(interactive / assisted / autonomous), and the canonical **approval
gate** are defined together in
[`workflow-control.md`](workflow-control.md).

---

## 1. Scoping a project (idea → MVP)

Skills that turn a rough thought or external PRD into a scoped MVP
and a phased build-out plan. Run these once per **MVP scoping round**
— once on a new project, then only when starting a major scope round
in the steady state (see [workflow-guide §3](workflow-guide.md#3-after-the-first-release-the-steady-state-loop)).

### `/idea-to-prd` (cat-1)

Turn a rough idea — a paragraph, a conversation, a notes file — into
a lightweight PRD draft at `design/prd.md`. Output is intentionally
minimal; the next skill (`/prd-normalizer`) reshapes it into the
kit's canonical format.

- **Use when:** you have only an idea, no PRD or planning notes.
- **Input:** prose or notes from the user; no required files.
- **Output:** `design/prd.md`.
- **Skip when:** you already have a PRD or custom planning notes — go
  straight to `/prd-normalizer`.
- **Spec:** [`skills/idea-to-prd/SKILL.md`](../skills/idea-to-prd/SKILL.md).

### `/prd-normalizer` (cat-1)

Take a standard-shaped PRD or custom planning notes and rewrite them
into the kit's canonical Normalized PRD format. Downstream skills
(`/prd-to-mvp`, `/adr-writer`) read this single artefact, so they
don't have to branch on the input shape.

- **Use when:** you have a PRD or custom notes that need to be
  reshaped into the kit's format before scoping.
- **Input:** an existing PRD or custom planning notes on disk.
- **Output:** `design/prd-normalized.md`.
- **Spec:** [`skills/prd-normalizer/SKILL.md`](../skills/prd-normalizer/SKILL.md).

### `/prd-to-mvp` (cat-1)

Scope the normalized PRD into a written MVP and a phased build-out
plan. The MVP statement says what's in scope, what's explicitly out
of scope, and what success looks like; the build-out plan breaks the
in-scope work into phases that downstream skills (`/issue-planner`,
`/release`) consume.

- **Use when:** you have a normalized PRD and want to scope the first
  release.
- **Input:** `design/prd-normalized.md`.
- **Output:** `design/mvp.md` and `design/build-out-plan.md`.
- **Key flag:** `--granularity={coarse|standard|fine}` sets a target
  phase count (1–3 / 5–8 / 8–12). The choice is recorded in the
  build-out plan so re-runs are consistent.
- **Example:** `/prd-to-mvp --granularity=fine` for a multi-quarter
  project where each phase warrants its own milestone.
- **Spec:** [`skills/prd-to-mvp/SKILL.md`](../skills/prd-to-mvp/SKILL.md).

### `/feature-prd` (cat-1)

Capture a **major feature update** to a mature project as an additive
PRD addendum at `design/prd-addenda/NNN-<feature-name>.md` that extends
`design/prd.md` rather than replacing it. Identifies ADR impact
(create / revise-via-supersession), records affected assumptions and an
explicit "what does not change" section, and decomposes the feature into
phased issues. This is the steady-state **expansion** path — the
`/feature` verb — distinct from the first-definition intake above.

- **Use when:** a change adds a capability area, alters original-PRD
  assumptions, or needs multiple issues/ADRs. **Not** for one-off bugs,
  small fixes, or a single self-contained issue.
- **Input:** a feature description; reads `design/prd.md`, accepted
  ADRs, and `design/state.md` (none modified).
- **Output:** `design/prd-addenda/NNN-<feature-name>.md`.
- **Next:** `/adr-writer` for surfaced decisions, then `/issue-planner`.
- **Spec:** [`skills/feature-prd/SKILL.md`](../skills/feature-prd/SKILL.md).

### `/planning` (cat-1, opt-in)

Capture deeper planning context — requirements decomposition, risks,
assumptions, phase-sequencing rationale, open research questions —
into `design/planning.md` before any ADRs are drafted. Optional;
small projects skip it and go straight from `/prd-to-mvp` to
`/adr-writer`.

- **Use when:** the project is large or unfamiliar enough that you
  want to harden ambiguity into structured questions before writing
  ADRs.
- **Input:** `design/prd-normalized.md` and `design/mvp.md`.
- **Output:** `design/planning.md`.
- **Spec:** [`skills/planning/SKILL.md`](../skills/planning/SKILL.md).

### `/clarify` (cat-1, opt-in)

Surface unresolved implementation questions — "gray areas" that
don't warrant a full ADR but need to be settled before ADRs are
written. Conducts a deep-dive per question and appends settled
decisions to `design/decisions.md` (a below-ADR-weight log).

- **Use when:** scoping is done but you can feel a handful of small
  implementation questions hanging in the air.
- **Input:** `design/prd-normalized.md`, `design/mvp.md`, optionally
  `design/planning.md`.
- **Output:** appends to `design/decisions.md`.
- **Spec:** [`skills/clarify/SKILL.md`](../skills/clarify/SKILL.md).

---

## 2. Capturing decisions

Skills that turn scoped problems into documented architectural
decisions, with a quality gate.

### `/adr-writer` (cat-1)

Draft one or more ADRs from a list of architectural decision topics
— typically the "Decisions needing ADRs" list surfaced by
`/prd-to-mvp`, but also usable for any decision that comes up later.
Each ADR uses the kit's template (context, options, decision,
consequences) and starts with status `proposed`. Acceptance is a
human act and is not done by this skill.

- **Use when:** you have decisions to capture (after `/prd-to-mvp`,
  or ad-hoc when an architectural question comes up later).
- **Input:** a list of decision topics (chat input or a "Decisions
  needing ADRs" section in the MVP).
- **Output:** one or more `design/adr/adr-NNN-*.md` files, status
  `proposed`.
- **Key flag:** `--skip-check` opts out of the chained `/check-plan`
  quality gate.
- **Example:** `/adr-writer --skip-check` when you've already vetted
  the decision topics and want to skip the chained quality round.
- **Spec:** [`skills/adr-writer/SKILL.md`](../skills/adr-writer/SKILL.md).

### `/check-plan` (cat-1)

Validate an ADR or issue prompt against a version-locked checklist.
Returns either a pass or a list of specific revisions per failed
criterion, and iterates with the user up to three rounds. Chained
automatically from `/adr-writer` and `/prepare-issue` as a pre-commit
gate (skip with `--skip-check`).

- **Use when:** automatically chained — you usually don't run it
  directly. Standalone, run on any ADR or prompt already on disk.
- **Input:** an ADR or prompt (file path, or in-memory rendering when
  chained).
- **Output:** a pass / fail report; on fail, specific revisions per
  criterion.
- **Programmatic equivalent:** `bin/check-plan` for non-interactive
  chained use.
- **Spec:** [`skills/check-plan/SKILL.md`](../skills/check-plan/SKILL.md).

---

## 3. Generating the backlog

### `/issue-planner` (cat-3)

Turn `design/mvp.md` and `design/build-out-plan.md` into a reviewed
batch of GitHub issues, plus a Project board. Drafts the full batch
(titles, bodies, labels, milestones, ADR references), shows it for
your approval, and only then creates the issues via `gh issue create`
and adds them to a new GitHub Project board.

- **Use when:** the MVP and build-out plan exist and you're ready to
  turn them into a tracked GitHub backlog. Also re-runnable later
  when you're adding a new milestone-worth of issues.
- **Input:** `design/mvp.md`, `design/build-out-plan.md`, accepted
  ADRs.
- **Output:** GitHub issues + Project board + (one milestone per
  phase, or one for the whole project on single-phase plans).
- **Permission gate:** explicit `yes` required before any `gh issue
  create` runs.
- **Spec:** [`skills/issue-planner/SKILL.md`](../skills/issue-planner/SKILL.md).

---

## 4. Per-issue execution loop

Skills that take one issue from "filed on GitHub" to "merged on
`main`." This loop runs many times per project — once for each
issue.

### `/prepare-issue` (cat-1)

Pull a GitHub issue's title, body, labels, milestone, and linked ADR
references via `gh`, read the relevant build-out-plan section, and
fill the prompt template. Result is a per-issue prompt at
`prompts/issue-NNN-<short-title>.md` ready to paste into a fresh
Claude Code session.

- **Use when:** you have an issue number and want a primed prompt to
  hand to the executor. Auto-chained from `/claude-issue-executor`
  when no prompt exists for the issue.
- **Input:** an issue number; reads `gh`, `design/adr/`, and
  `design/build-out-plan.md`.
- **Output:** `prompts/issue-NNN-<short-title>.md`. Embeds
  carry-forward notes from a recently-merged PR if any (see
  [workflow-guide §6](workflow-guide.md#6-cross-skill-carry-forward-adr-040)).
- **Key flag:** `--skip-check` opts out of the chained `/check-plan`
  gate.
- **Example:** `/prepare-issue 42 --skip-check` to fill the prompt
  for issue #42 without running the quality gate.
- **Spec:** [`skills/prepare-issue/SKILL.md`](../skills/prepare-issue/SKILL.md).

### `/claude-issue-executor` (cat-2)

Run a full implementation session from a prepared prompt: parse the
prompt, propose a step-by-step plan, **block for explicit user
approval**, create a feature branch, implement in incremental commits
(referencing the ADR and issue), write tests alongside the code, and
emit an evaluation summary at the end.

- **Use when:** a prompt is ready (or you want to auto-chain
  `/prepare-issue` and execute in one go).
- **Input:** a prompt path, or an issue number (auto-chains
  `/prepare-issue`).
- **Output:** a feature branch with commits + tests, plus
  `notes/eval-issue-NNN.md` containing the eval summary and any
  structured `design-questions` YAML for cross-issue carry-forward.
- **Key flag:** `--no-prompt` for trivial issues (typo fixes,
  dependency bumps, ADR status flips) — skips the prompt artefact
  entirely.
- **Example:** `/claude-issue-executor 42 --no-prompt` for a typo
  fix or dependency bump where the issue body is enough context.
- **Plan-mode rhythm:** for significant sessions, the executor pauses
  and asks you to enter Claude Code plan mode before proposing the
  plan; trivial sessions skip the gate.
- **Spec:** [`skills/claude-issue-executor/SKILL.md`](../skills/claude-issue-executor/SKILL.md).

### `/pr-review-packager` (cat-3)

Package the current branch into a reviewable pull request. Reads the
PR template, fills `Closes #N` and ADR references from the branch
name and commit history, derives a change summary from `git
log main..HEAD`, shows the body for approval, and only then calls
`gh pr create`.

- **Use when:** the executor's branch is ready for review.
- **Input:** the current branch (must be ahead of `main`).
- **Output:** an open PR with a clean body. Adds a `## Notes for #N`
  section to the PR body for any cross-issue design questions
  surfaced by the executor (see
  [workflow-guide §6](workflow-guide.md#6-cross-skill-carry-forward-adr-040)).
- **Permission gate:** explicit `yes` required before `gh pr
  create`.
- **Spec:** [`skills/pr-review-packager/SKILL.md`](../skills/pr-review-packager/SKILL.md).

### `/review-pr` (cat-3)

AI-review an **already-open** PR. Generates a local dry-run review
artifact (Markdown + JSON) from the PR diff using a configured
OpenAI-compatible provider (OpenRouter first), classifies findings as
**blocking / non-blocking / question / praise**, and — only after you
preview the exact comments and give explicit approval — publishes a PR
review. Safe by default: generation posts nothing; publishing is the
cat-3 GitHub write, gated by a deterministic `--confirm publish-pr-N`
token and made idempotent by a receipt.

- **Use when:** you want a structured first-pass review of an open PR.
- **Input:** a PR number; provider config in `ai-review/config.json` with
  the API key in the env var it names (never pasted into chat or
  committed). See [`ai-review.md`](ai-review.md).
- **Output:** `ai-review/artifacts/pr-N-<hash>.{json,md}` always; a posted
  GitHub review only after explicit approval, plus a receipt.
- **Permission gate:** dry-run is free; posting always requires an
  explicit `yes` and the `--confirm` token, in every operating mode.
- **Tools:** [`bin/review-pr`](../bin/review-pr),
  [`bin/publish-review`](../bin/publish-review),
  [`bin/review-eval`](../bin/review-eval).
- **Spec:** [`skills/review-pr/SKILL.md`](../skills/review-pr/SKILL.md).

---

## 5. Closing milestones

Skills that wrap up a milestone (a phase, or a whole project on
single-phase setups) before cutting a release.

### `/audit-milestone` (cat-1)

Read a GitHub milestone, verify it is finishable, and return a
pass/fail report with a concrete gap list. Three checks: every issue
closed, every referenced ADR linked to a merged PR, and (when phased)
every `## Phase` block's exit criterion satisfied.

- **Use when:** before closing a milestone. The audit is
  **advisory** — gaps are surfaced but the audit never blocks the
  close.
- **Input:** a milestone number.
- **Output:** a pass / fail report in chat.
- **Spec:** [`skills/audit-milestone/SKILL.md`](../skills/audit-milestone/SKILL.md).

### `/milestone-summary` (cat-1)

Render a retrospective for a milestone into
`design/milestones/<N>-<slug>.md`. Filled from `git log` between
phase-boundary tags, the GitHub milestone, and accepted ADRs in the
date range. The `lessons` zone is **user-authored** and is preserved
verbatim across re-runs.

- **Use when:** before or just after closing a milestone, to capture
  what shipped, what was deferred, and lessons learned.
- **Input:** a milestone number.
- **Output:** `design/milestones/<N>-<slug>.md`.
- **Key flag:** `--overwrite` to regenerate (preserves the lessons
  zone verbatim).
- **Example:** `/milestone-summary 3 --overwrite` to regenerate the
  Phase 3 summary after late work has merged into the milestone.
- **Spec:** [`skills/milestone-summary/SKILL.md`](../skills/milestone-summary/SKILL.md).

### `/complete-milestone` (cat-3)

Close a milestone end-to-end: chains `/audit-milestone`, then
`/milestone-summary`, closes the GitHub milestone via `gh`, archives
state.md zones (clears in-flight, prepends to recent, sets
continue-here), and optionally chains `/release` with
`--milestone-phase=N`.

- **Use when:** the milestone's issues are merged and you're ready
  to close it cleanly.
- **Input:** a milestone number.
- **Output:** closed GitHub milestone, refreshed `design/state.md`,
  and optionally a tagged release.
- **Key flags:**
  - `--release` chains `/release --milestone-phase=N` after the
    close.
  - `--skip-summary` skips the chained summary.
- **Example:** `/complete-milestone 3 --release` to close Phase 3
  and immediately cut a release tied to it.
- **Permission gate:** explicit `yes` required before closing the
  GitHub milestone.
- **Spec:** [`skills/complete-milestone/SKILL.md`](../skills/complete-milestone/SKILL.md).

### Skills without a worked-example sidecar

`/complete-milestone` and `/milestone-summary` are the only two
skills in the cohort that ship without their own `example.md`. Both
are thin routing-orchestration layers: `/complete-milestone` chains
`/audit-milestone`, `/milestone-summary`, and optionally `/release`;
`/milestone-summary` reads `git log`, `gh`, and accepted ADRs to
compose one file. The full chain walkthrough lives in
[`skills/audit-milestone/example.md`](../skills/audit-milestone/example.md),
which exercises all three milestone-cluster skills against a phased
example project. The omission is intentional, not drift.

---

## 6. Cutting releases

### `/changelog` (cat-1)

Generate release notes from `git log` between two refs. Parses each
commit for verb prefix, ADR tokens, and issue references, groups
them into sections (Features, Fixes, Docs, Refactoring, Chores,
Other), and emits markdown to stdout, a file, or a GitHub Release
body.

- **Use when:** you want release notes — for a `CHANGELOG.md` entry,
  for a Release body, or for a quick preview before tagging.
- **Input:** two refs (defaults to `last-tag..HEAD`).
- **Output:** markdown release notes (stdout, file, or Release
  body).
- **Note:** this skill never tags or pushes; tagging is `/release`'s
  job.
- **Spec:** [`skills/changelog/SKILL.md`](../skills/changelog/SKILL.md).

### `/release` (cat-3)

Tag a semver release, generate notes via `/changelog`, push the tag,
and publish a GitHub Release. Every mutating step is gated behind a
single explicit approval.

- **Use when:** `main` is at a state worth shipping.
- **Input:** a bump (`patch` / `minor` / `major`); auto-detected on
  phased projects via `--milestone-phase=N`.
- **Output:** an annotated git tag (pushed), a published GitHub
  Release, and (on phased projects) the build-out plan's phase
  status row updated.
- **Key flags:**
  - `--milestone-phase=N` ties the release to a specific phase.
  - `--force-product-shape` / `--force-workflow-shape` override the
    project-shape auto-detection (see below).
- **Example:** `/release minor --milestone-phase=3` to cut a minor
  release for the Phase 3 milestone.
- **Project-shape detection:** auto-detects whether the project is
  software-shaped or workflow-shaped (research projects, books,
  curricula, content projects, design system docs, internal-policy
  documents) and adjusts the release-body framing.
- **Permission gate:** explicit `yes` required before any `git push`
  or `gh release create`.
- **Spec:** [`skills/release/SKILL.md`](../skills/release/SKILL.md).

---

## 7. Keeping docs in sync

### `/workflow-docs` (cat-1)

Re-render `README.md`, `design/architecture.md`, and
`design/ai-summary.md` from the project's current artefacts (PRD, MVP,
accepted ADRs, planning docs, `CLAUDE.md`). Generated
regions are wrapped in marker fences so manual edits outside those
regions are preserved across re-runs. Sections with no source data
are omitted entirely rather than left blank.

- **Use when:** the project's shape has changed meaningfully — new
  ADRs accepted, a major feature shipped, a pivot in scope. The
  architecture doc is the current human-readable system shape;
  `ai-summary.md` is what you paste into external AIs for
  second-opinion design reviews, so it's worth keeping current.
- **Input:** project artefacts (PRD, MVP, planning docs, decisions, ADRs, `CLAUDE.md`).
- **Output:** updated `README.md`, `design/architecture.md`, and `design/ai-summary.md`.
- **Spec:** [`skills/workflow-docs/SKILL.md`](../skills/workflow-docs/SKILL.md).

---

## 8. Cross-session continuity and routing

Skills that bridge multiple Claude Code sessions, so a fresh session
can pick up exactly where the last one left off, and route you to the
next step. Backed by a single committed pointer file, `design/state.md`.

### `/start` (cat-1)

The kit's front door and "what now?" router. Inspects project state —
`design/state.md` (especially the `next-action` zone), the PRD/MVP
artefacts, prepared prompts, and open branches — and recommends, or for
cat-1 targets invokes, the next appropriate skill. States the active
operating mode and asks one clarifying question when the next step is
ambiguous rather than dumping the skill list. Also answers to `/next`.

- **Use when:** at the start of a session, after finishing a step, or
  any time the next move is unclear. On an empty project it doubles as
  onboarding (routes toward `/idea-to-prd`).
- **Input:** an optional one-line goal hint; reads state and artefacts
  (never mutates).
- **Output:** a routing recommendation (and, for cat-1 targets, an offer
  to run them).
- **Spec:** [`skills/start/SKILL.md`](../skills/start/SKILL.md).

### `/resume` (cat-1)

Read `design/state.md` and emit a one-message brief: current phase,
in-flight issue (with its prompt and branch), the last few merged
PRs, blockers, and the "continue here" pointer. No `gh` calls on the
happy path; falls back to `gh pr list` / `gh issue list` if state.md
is missing, empty, or looks stale.

- **Use when:** at the start of every fresh Claude Code session in a
  project that uses `design/state.md`.
- **Input:** `design/state.md` (or `gh` fallback).
- **Output:** a chat brief; no file writes.
- **Spec:** [`skills/resume/SKILL.md`](../skills/resume/SKILL.md).

### `/pause` (cat-1)

Refresh `design/state.md` so it reflects right now — phase, in-flight
issue, recent work, blockers, continue-here. Optionally writes a
richer `notes/handoff-YYYY-MM-DD.md` for context-window-exhausting
session handoffs.

- **Use when:** before a context reset, end of day, after a
  non-trivial detour, or to seed `design/state.md` for the first
  time.
- **Input:** project state (build-out plan, prompts, branches, `gh
  pr list`).
- **Output:** refreshed `design/state.md`; with `--handoff`, also a
  handoff file in `notes/`.
- **Key flag:** `--handoff` for the richer context-window-exhausting
  case.
- **Example:** `/pause --handoff` before a long break or when the
  context window is about to roll over and you want a fresh session
  to pick up cleanly.
- **Spec:** [`skills/pause/SKILL.md`](../skills/pause/SKILL.md).

---

## Ancillary tooling (`bin/`)

The kit ships a few small command-line tools alongside the skills.
These are scaffolding rather than user-facing skills; most users
never invoke them directly.

| Tool | Purpose |
|---|---|
| `bin/check-plan` | Programmatic equivalent of `/check-plan` for chained, non-interactive use |
| `bin/check-state-cap` | CI guard rail: exits 1 if `design/state.md` exceeds the line cap (default 100) |
| `bin/sync-adr-index` | Rebuild `design/adr/README.md` from the ADRs in `design/adr/` |
| `bin/check-plan-criteria-drift` | Detect drift between `/check-plan` criteria and the canonical version-locked checklist |
| `bin/install-workflow-kit` | Install the kit into a target project |
| `bin/bootstrap-workflow-kit` | One-shot bootstrap helper for new projects |
| `bin/review-pr` | Generate a dry-run AI review artifact from a PR diff (backs `/review-pr`) |
| `bin/publish-review` | Preview, then post an AI review to a PR on explicit approval |
| `bin/review-eval` | Offline fixture harness for AI PR review quality |

The standard-envelope `bin/*` surfaces (including the three review tools)
are catalogued in [`agent-contract.md` §4](agent-contract.md#4-programmatic-surfaces);
the AI-review tools are documented for operators in [`ai-review.md`](ai-review.md).

---

## Alphabetical index

| Skill | Category |
|---|---|
| `/adr-writer` | §2 Capturing decisions |
| `/audit-milestone` | §5 Closing milestones |
| `/changelog` | §6 Cutting releases |
| `/check-plan` | §2 Capturing decisions |
| `/clarify` | §1 Scoping (opt-in) |
| `/claude-issue-executor` | §4 Per-issue execution |
| `/complete-milestone` | §5 Closing milestones |
| `/feature-prd` | §1 Scoping (expansion) |
| `/idea-to-prd` | §1 Scoping |
| `/issue-planner` | §3 Generating the backlog |
| `/milestone-summary` | §5 Closing milestones |
| `/pause` | §8 Cross-session continuity and routing |
| `/planning` | §1 Scoping (opt-in) |
| `/pr-review-packager` | §4 Per-issue execution |
| `/prd-normalizer` | §1 Scoping |
| `/prd-to-mvp` | §1 Scoping |
| `/prepare-issue` | §4 Per-issue execution |
| `/release` | §6 Cutting releases |
| `/review-pr` | §4 Per-issue execution |
| `/resume` | §8 Cross-session continuity and routing |
| `/start` | §8 Cross-session continuity and routing |
| `/workflow-docs` | §7 Keeping docs in sync |
