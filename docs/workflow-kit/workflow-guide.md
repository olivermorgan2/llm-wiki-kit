# Workflow Guide — Idea to Shipped Release

A practical walkthrough of the full Claude Code Workflow Kit flow.
The kit covers two phases of a project's life:

- **The MVP path** (§2) — the first pass through, taking you from a
  rough idea to a tagged first release. Run end-to-end the first
  time you use the kit on a new project.
- **The steady-state loop** (§3) — how you keep iterating after the
  first release. Most of your time post-MVP is spent here, returning
  to §2's earlier steps only for major scope rounds.

Both phases share most of the same skills; the difference is which
steps repeat and how often.

**Who this is for:** a solo developer or small team who has installed
the kit into a new target project and wants to ship a feature without
piecing the flow together from individual skill docs.

**What you end up with:** at the end of the MVP path, a tagged first
release on GitHub with an accompanying PRD, MVP scope, ADRs, issues,
and PRs. From the steady-state loop onward, a continuously evolving
project with versioned releases (per the MAJOR/MINOR/PATCH policy in
§3), a continuously maintained architecture document, and a documented
idea backlog (§4) feeding the next round of work.

This guide focuses on *what to run next*. For the kit's current architecture, see [`docs/architecture.md`](architecture.md).
For detailed decision history in a target project, see that project's
`design/adr/` directory.
Skills are referenced by their slash name (e.g. `/idea-to-prd`); each
one has its own `SKILL.md` under `.claude/skills/<name>/` in your
target project with the full interface spec.

---

## At a glance

The diagram below shows the **MVP path (§2)** — the first pass through,
from idea to tagged first release. Iteration after the first release
is described in §3 (the steady-state loop), and the input pipeline
that feeds it is §4 (the idea backlog).

Six numbered phases. Steps 1–3 happen once per MVP scoping round;
step 4 loops per issue; step 5 closes a phase milestone; step 6 ships
the release. `⚠cat-3` marks skills that write to GitHub and require
explicit approval even under auto-mode (§7).

```
IDEA OR PRD
   │
   ▼
1. SCOPING  (once per MVP round)
   ├─ /idea-to-prd          (rough idea → PRD)
   ├─ /prd-normalizer       (external PRD → kit format)
   ├─ /prd-to-mvp [--granularity=coarse|standard|fine]
   │     → design/mvp.md + design/build-out-plan.md
   └─ /planning             (optional, deeper plans → design/planning.md)
   │
   ▼
2. DECISIONS  (per architectural choice)
   ├─ /clarify              (optional, scopes the question
   │                         → design/decisions.md)
   ├─ /adr-writer           → design/adr/adr-NNN-*.md
   │                         (proposed → accepted)
   └─ /check-plan or bin/check-plan   (advisory gate)
   │
   ▼
3. BACKLOG
   └─ /issue-planner  ⚠cat-3
         → GitHub issues + Project board
         + one milestone per phase
   │
   ▼
4. PER-ISSUE LOOP  ◄────────────── repeat for each issue in the milestone
   │
   ├─ /prepare-issue N
   │     → prompts/issue-NNN-*.md
   │       (auto-chained from executor if no prompt exists)
   │       (embeds carry-forward "## Notes for #N" if a prior PR left any)
   │
   ├─ /claude-issue-executor
   │     1. significance check
   │     2. plan mode → approve
   │     3. branch + commits + tests
   │     → notes/eval-issue-NNN.md
   │       (incl. design-questions YAML when relevant)
   │
   ├─ /pr-review-packager  ⚠cat-3
   │     → PR body (summary, ADR link, test results,
   │       "## Notes for #M" carry-forward to future issues)
   │
   ├─ self-review + CodeRabbit + optional AI review
   │
   └─ gh pr merge --squash --delete-branch
         │
         ├─ next issue?  → loop back to /prepare-issue
         └─ milestone done? → continue to step 5
   │
   ▼
5. MILESTONE CLOSE  (per phase)
   ├─ /audit-milestone N    (advisory pass/fail report)
   ├─ /milestone-summary N  → design/milestones/N-slug.md
   │                         (lessons zone is user-authored)
   └─ /complete-milestone N [--release]  ⚠cat-3
         (closes GH milestone; chains /release if --release)
   │
   ▼
6. RELEASE
   ├─ /changelog            → release notes (grouped by verb / ADR / issue)
   ├─ /release  ⚠cat-3      → git tag + gh release create
   │                          (auto-detects product vs workflow shape)
   └─ /workflow-docs        → README.md + design/architecture.md
                              + design/ai-summary.md
   │
   ▼
TAGGED RELEASE ON GITHUB
```

### Side-channels (run alongside the main flow)

**Cross-session state (§5).** Long-running projects use a
committed pointer file so each new session knows where to pick up.

```
/pause [--handoff]   → writes design/state.md
/resume              → reads design/state.md at session start

design/state.md zones:
   phase | in-flight | recent | blockers | continue-here | next-action
```

**Cross-skill carry-forward (§6).** Design questions raised
during one issue propagate to a later, dependent issue:

```
/claude-issue-executor (issue M)
   writes structured design-questions YAML
   in notes/eval-issue-M.md  (under ## Follow-ups)
        │
        ▼
/pr-review-packager renders them as
   "## Notes for #N" in M's PR body
        │
        ▼ (after the PR merges)
/prepare-issue N reads the merged PR's Notes for #N
   and embeds them in the prompt as
   "## Design questions carried forward from PR #M"
```

**Permission categories (§7).** Every skill is classified for how
auto-mode treats it:

```
cat-1   local, reversible        → auto-mode proceeds without prompts
        (most skills: /idea-to-prd, /adr-writer, /prepare-issue,
         /changelog, /resume, /pause, etc.)

cat-2   operator-acknowledged    → auto-mode proceeds but must
        bypass                     state the bypass in chat output
        (only /claude-issue-executor's plan-mode gate)

cat-3   public / hard-to-reverse → explicit approval ALWAYS required,
                                   regardless of mode
        (/issue-planner, /pr-review-packager, /release,
         /complete-milestone)
```

---

## 1. Prerequisites

You should have already:

- Installed Git, GitHub CLI, and Claude Code, and authenticated `gh`.
- Created the target project on GitHub and cloned it locally.
- Copied the kit's skills into `.claude/skills/` and rendered
  `CLAUDE.md` at the project root.

If any of that is missing, stop here and complete
[`docs/install.md`](install.md) first. The rest of this guide assumes
you are sitting inside the target project's root with Claude Code open.

---

## 2. The MVP path: idea to first release

One full pass takes a new project from idea to its first tagged
release. Steps (a) through (c) are scoping ceremonies — done once
per MVP scoping round on a new project, and revisited only for major
scope rounds in the steady state (§3). Steps (d) through (g) loop
per issue. Steps (h) through (j) close out a milestone and cut the
release.

### 2.a From idea or PRD to MVP

**Why this step exists.** This is where the dream becomes a plan you
can execute. PRD = what you eventually want; MVP = the smallest first
cut that delivers value; issues = the concrete work. Without a
discrete MVP, scope tends to bleed and every feature ends up in the
first release. The MVP doc is also the input `/issue-planner` reads
to size the backlog — there's no path from PRD straight to issues.

Pick the skill that matches what you have in hand:

| You have… | Run | You produce |
|---|---|---|
| A rough idea, no doc | `/idea-to-prd` | A lightweight PRD in `design/` |
| A standard or custom PRD | `/prd-normalizer` | A normalized PRD in the kit's internal format |
| A normalized PRD | `/prd-to-mvp` | An MVP scope statement: in-scope, out-of-scope, success signals |

Run them in sequence if you have a rough idea: `/idea-to-prd` →
`/prd-normalizer` (if you later want to reshape it) → `/prd-to-mvp`.

**Granularity.** `prd-to-mvp` and `/planning` accept
an optional `--granularity={coarse|standard|fine}` flag that sets a
target band for phase count: `coarse` aims for 1–3 phases, `standard`
(the default) for 5–8, `fine` for 8–12. Bands are *targets*, not hard
caps — the skill picks the actual count for your project and includes
an inline justification in the rendered build-out-plan. The choice is
recorded as a `**Granularity:**` line in `design/build-out-plan.md`,
so re-runs are consistent without you having to remember the flag.
Precedence (highest first): explicit flag, then the stored value in
the build-out-plan, then the default `standard`. Pick `coarse` for
weekend or one-week projects, `standard` for typical multi-month
builds, `fine` for large multi-quarter scope where each phase warrants
its own milestone.

**What unblocks next:** a written MVP statement that scopes the first
release. You now know what decisions you need to lock in.

### 2.b Design decisions to ADRs

**Why this step exists.** ADRs are how a vague idea ("we should use
JWT") becomes a documented decision ("we chose JWT over sessions
because X, with these consequences"). One ADR per decision, each one
standing alone, so you can review, reject, or supersede them
individually — and so the rationale is still discoverable six months
later when you (or a future contributor) ask why the project is
shaped the way it is.

For each meaningful architectural choice surfaced by the MVP — a library
pick, an integration, a data model, a deployment target — write an ADR.

- **Run:** `/adr-writer`
- **What you see:** a drafted ADR in `design/adr/adr-NNN-<short-title>.md`
  following the kit's template (context, options, decision, consequences).
- **What you produce:** one accepted ADR per decision.

One ADR per decision. If the skill drafts several, accept them one at a
time — each ADR should stand alone.

**What unblocks next:** a small set of accepted ADRs that define the
backlog's shape.

### 2.c Plan to backlog

**Why this step exists.** Until the work is on GitHub, it's a
wishlist, not a backlog. Issues give each task a stable identifier, a
comment thread, a label and milestone home, and — crucially —
something a PR can close on merge. The phased breakdown from the
build-out plan becomes one milestone per phase, so progress is
visible and releases have natural cut points.

Turn the MVP plus ADRs into a tracked GitHub backlog.

- **Run:** `/issue-planner`
- **What you see:** a batch of drafted issues (title, body, labels,
  milestone, ADR reference) presented for your approval before any are
  created.
- **What you produce:** GitHub issues created via `gh issue create`,
  plus a GitHub Project board organising them.

The skill is GitHub-first: issues live on GitHub, not in local files.
The Project board gives you a kanban view of the backlog.

**Phased vs single-phase plans.** When
`design/build-out-plan.md` contains multiple `## Phase N: <name>`
blocks, `/issue-planner` creates one GitHub milestone per phase and
assigns each issue to its phase's milestone; `/release` defaults to
one tag per phase. When the plan has only one Phase block — or no
Phase headings at all — every downstream skill behaves identically
to a flat path: one milestone for the whole project,
one release at the end. The single-phase fallback is the right shape
for small projects (weekend or one-week scope); multi-phase shape
fits multi-month builds where each phase delivers standalone value.
See [`examples/projects/phased-podcast-pipeline/`](../examples/projects/phased-podcast-pipeline/)
for a worked 3-phase example, or [`examples/projects/kb-lookup/`](../examples/projects/kb-lookup/)
and [`examples/projects/slug-utils/`](../examples/projects/slug-utils/)
for the single-phase shape.

**Anti-pattern: tuning granularity to dodge scope.** The
`--granularity` flag (documented in step 2.a) is a knob for *how
many phases your scope decomposes into*, not a knob for shipping a
large project faster. If you find yourself reaching for `coarse` on
a multi-month project so the build-out plan looks shorter, the right
move is to cut MVP scope (re-run `prd-to-mvp` with a tighter scope
question), not to compress the build-out into fewer phases. Phases
are how the work is *paced*; scope is what is *delivered*. Don't
confuse the two. ADR-036's "Consequences" section calls this out as
the explicit anti-pattern this section guards against.

**What unblocks next:** a list of issue numbers you can work through
one at a time.

### 2.d One issue at a time to a prompt

**Why this step exists.** Pasting an issue body straight into Claude
Code skips half the context: the linked ADR, the relevant build-out
plan section, the kit's conventions. `/prepare-issue` collects all
of that into one prompt file you can review before the build session
starts. It's a deliberate pause — the prompt is written, but
execution starts separately, so you can sanity-check the framing
before any files get edited.

Before writing any code, generate a per-issue Claude Code session
prompt.

- **Run:** `/prepare-issue <issue-number>`
- **What you see:** the skill pulls the issue body and its linked ADR
  via `gh`, reads the relevant build-out plan section, and fills the
  prompt template.
- **What you produce:** `prompts/issue-NNN-<short-title>.md`, ready
  to paste into a fresh Claude Code session.

This is a deliberate handoff — the prompt is written but execution
starts separately, so you get a review checkpoint between planning and
building.

**Auto-chain.** This step is *optional*:
`/claude-issue-executor <issue-number>` will auto-invoke
`/prepare-issue` itself when no `prompts/issue-NNN-*.md` exists for
the issue, logging the prep step prominently. You can still run
`/prepare-issue` explicitly if you want a review checkpoint between
prompt generation and execution — that is the recommended path for
significant work. **Stale prompts** (the issue body or any linked ADR
has been updated since the prompt was last written) trigger a
regeneration prompt with confirmation; the executor never silently
ignores a stale prompt.

**What unblocks next:** a prompt file you can kick off whenever you
have time to build — or skip this step and let the executor
auto-chain.

**Cross-issue carry-forward (see §6).** When a recently-merged
PR contains a `## Notes for #<this-issue>` section (left by the
upstream issue's `pr-review-packager`), `prepare-issue` embeds it in
the generated prompt as a *Design questions carried forward from PR
#M* subsection. This is how cross-issue design-coherence questions
reach the executor as deterministic context rather than depending on
chain-aware authoring — see §6.

### 2.e Prompt to implementation

**Why this step exists.** This is where code actually gets written.
The plan-first model is the whole point: the executor proposes a
step-by-step plan and waits for your approval before touching any
files. Approving a plan is much cheaper than reverting
bad code. The evaluation summary at the end — files changed, tests
run, manual verification needed — is your quality gate before the
work moves to PR review.

Open a fresh Claude Code session and hand it the prompt.

- **Run:** `/claude-issue-executor <path-to-prompt>`
- **What you see:** the skill enters plan mode, proposes a step-by-step
  plan, and waits for your approval. Only after you approve does it
  create a feature branch and start implementing.
- **What you produce:** a branch with incremental commits (each
  referencing the ADR and issue number), tests written alongside code,
  and an evaluation summary at the end.

This is the plan-first execution model. The evaluation
summary is your quality gate — it lists files changed, test results,
and any manual verification steps. Do not skip reading it.

**Plan-mode rhythm.** For significant sessions, the
executor pauses at the start and asks you to enter Claude Code plan
mode — the harness-level lock that gates all mutating tools — before
proposing the plan. The full rhythm:

1. The executor reports *"Significant — toggle plan mode when ready."*
2. You toggle plan mode (`shift+tab shift+tab`).
3. The executor proposes the plan inside plan mode.
4. You approve and exit plan mode (`shift+tab` once).
5. Optionally, toggle auto-accept edits (`shift+tab` again) so the
   execution phase runs without per-tool approval prompts.
6. The executor implements; if a new significant boundary appears
   mid-session it pauses and re-flags before crossing it.

Plan mode is requested automatically when the session crosses the
"significant" threshold defined in
[`skills/claude-issue-executor/SKILL.md`](../skills/claude-issue-executor/SKILL.md)
(edits a `skills/*/SKILL.md`, modifies 3+ files, edits `templates/*`,
etc.). Trivial sessions (single typo, ADR status flip, doc tweak)
skip plan mode and rely on the chat plan-gate alone.

**`--no-prompt` mode.** Genuinely-trivial issues —
typo fixes, dependency bumps, doc tweaks, ADR status flips — don't
earn ADR-shaped ceremony. Pass `--no-prompt` to skip prompt
generation entirely; the executor reads the issue body directly and
leaves a one-line breadcrumb in the first commit's message
(`issue executed without prompt per ADR-038`) for the audit trail.
The criteria for when `--no-prompt` is appropriate are exactly the
**Trivial checklist** in ADR-039 (single source of truth — see
`skills/claude-issue-executor/SKILL.md`). The executor *auto-detects*
candidates conservatively: zero `ADR-NNN` references in the issue
body **and** label in `chore` / `docs` / `bugfix-trivial`. When both
hold, it suggests `--no-prompt` and asks for confirmation; explicit
`--no-prompt` overrides the auto-detect without asking. Plan-mode
rhythm still applies — `--no-prompt` skips the prompt step, not the
significance gate.

**Branch naming:** the skill creates branches following the kit's
convention (see [GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow)) —
e.g. `add-auth-middleware` or `NN-add-auth-middleware`.

**What unblocks next:** a working, tested branch ready for PR.

**Carry-forward output (see §6).** When the executor's eval
summary surfaces a design question that affects an upcoming issue,
it records the question as a structured `design-questions:` entry in
`notes/eval-issue-NNN.md`. The packager picks this up at PR time —
see §6 for the schema, when-to-populate rule, and downstream flow.

### 2.f Branch to PR

**Why this step exists.** A PR is the review surface — for you, for
CodeRabbit, for any human reviewer. `/pr-review-packager` assembles
the body (summary, ADR link, file-level changelog, test results,
`Closes` trailer) so you review the package once instead of writing
PR descriptions from scratch every time. It's also where
carry-forward design questions (§6) get parked in the PR body
so a future issue's prompt can pick them up.

Package the branch into a reviewable pull request.

- **Run:** `/pr-review-packager`
- **What you see:** the skill drafts a PR body using the kit's PR
  template — summary, ADR link, file-level changelog, test results,
  `Closes #NN` trailer — and shows it to you before creating the PR.
- **What you produce:** a PR opened via `gh pr create`, with a clean
  body and the linked issue set to auto-close on merge.

For work that is not yet done, open as a draft PR (`gh pr create
--draft`) to get early feedback; convert with `gh pr ready <number>`
when implementation and tests are complete.

**What unblocks next:** a PR you can self-review and merge.

**Carry-forward propagation (see §6).** When the executor's
`notes/eval-issue-NNN.md` contains `design-questions:` entries, the
packager emits a `## Notes for #N` section in the PR body for each
unique target-issue, so the carried-forward context survives in PR
history and reaches `prepare-issue` deterministically when the
target issue is later prepared — see §6.

### 2.g Merge, then loop back

**Why this step exists.** Squash merge keeps `main`'s history
readable as a sequence of features and fixes rather than a sequence
of "wip" commits. Deleting the branch on merge keeps the repo tidy.
After this, you're back at step 2.d for the next issue — the loop
is what turns a backlog into shipped work, one issue at a time.

Review the diff on GitHub, resolve any CodeRabbit or human review
comments, then merge.

```bash
gh pr merge <number> --squash --delete-branch
git checkout main
git pull
```

- Squash merge is the default — it keeps `main`'s history clean.
- Use a regular merge only when the per-commit history of the branch
  carries real value.
- Delete the branch on merge so the repo stays tidy.

**Loop back to step 2.d** for the next issue. Keep going until the
milestone's issues are all merged.

### 2.h Close out the milestone

**Why this step exists.** Without an explicit close-out step,
milestones drift: issues linger half-done, retrospectives never get
written, and lessons from one phase don't reach the next. The three
skills here split that work cleanly — audit surfaces gaps before
they become surprises, summary captures lessons while the context
is fresh, and complete archives state and (optionally) cuts the
release.

Before cutting a release, close the GitHub milestone the merged
issues belong to. Three composable skills handle this, chained in
order:

| Step | Run | Produces |
|---|---|---|
| 1. Verify the milestone is finishable | `/audit-milestone <N>` | Pass/fail report — open issues, ADRs without merged PRs, phases still `in-progress` |
| 2. Draft the retrospective | `/milestone-summary <N>` | `design/milestones/<N>-<slug>.md` filled from `git log`, the GitHub milestone, and accepted ADRs in the date range; `lessons` zone left for you to author |
| 3. Close and (optionally) release | `/complete-milestone <N> [--release]` | GitHub milestone closed; `design/state.md` archived (in-flight cleared, recent prepended, continue-here updated); chains `/release --milestone-phase=N` if `--release` is set |

The audit is **advisory**, not gating — gaps surface clearly but
the user has the final say. The summary file's `lessons` zone is
**user-authored** and preserved verbatim across re-runs (even with
`--overwrite`). All three skills run standalone or in any subset;
`/complete-milestone` chains the previous two as a convenience.

**One milestone == one release (default).**
`/issue-planner` creates one GitHub milestone per phase, and
`/release` defaults to one tag per phase via `--milestone-phase=N`.
Pass `--release` to `/complete-milestone` to chain `/release` after
the close, or omit it to close the milestone without cutting a tag
(e.g. when bundling multiple phases into one release).

For a worked walkthrough — including the fail case where Phase 2 has
an open issue and an ADR without a merged PR — see
[`skills/audit-milestone/example.md`](../skills/audit-milestone/example.md).

**What unblocks next:** a closed milestone with an archived state
pointer and a retrospective on disk; either a release just cut (if
`--release` chained) or a clean state to run `/release` separately.

### 2.i Tagged release

**Why this step exists.** A tag is the snapshot stakeholders read.
Without a release step, "what changed in v1.2?" has no good answer
— you're reading the commit log. `/changelog` groups changes by
verb, ADR, and issue so the notes tell a story; `/release` tags,
pushes, and publishes. On phased projects, this is where one
milestone becomes one release.

Once a milestone is closed (or as a standalone step on a single-phase
project), cut a release.

| Step | Run | Produces |
|---|---|---|
| 1. Draft release notes | `/changelog` | Markdown grouped by verb, ADR, and issue, written to stdout, `CHANGELOG.md`, or a GitHub Release body |
| 2. Tag and publish | `/release` | Semver-bumped annotated git tag, pushed to origin, GitHub Release created via `gh release create` |

`/release` calls `/changelog` internally, so you can run it directly
once you are happy with the tag and bump (patch / minor / major). It
can also mark the milestone phase complete in your build-out plan.

**Project-shape detection.** `/release` auto-detects
whether the project is software-shaped or workflow-shaped (research
projects, books, curricula, content projects, design system docs,
internal-policy documents, etc.) by scanning for non-product
indicators in the PRD, build-out plan, and repo root. When two or
more signals fire, the release body leads with a workflow-shape
clarifier banner instead of standard product framing. The
**canonical signal list** lives in
[`skills/release/SKILL.md` § Project-shape detection](../skills/release/SKILL.md)
— that file is the single source of truth; do not maintain a
parallel list here. Operators on borderline projects can override
auto-detection with `--force-product-shape` or `--force-workflow-shape`.

**What unblocks next:** a tagged, published release that stakeholders
can read and that future `/changelog` runs will pick up as the lower
bound.

### 2.j Docs sync

**Why this step exists.** `README.md` and `design/ai-summary.md`
drift. Each shipped feature, each accepted ADR, each scope pivot
makes them slightly less accurate. `/workflow-docs` re-renders both
from the current state of the project, so the docs people actually
read — and the AI summary you paste into external LLMs for second
opinions — stay aligned with reality.

Before closing out the release, regenerate the docs that summarise the
current state of the project.

- **Run:** `/workflow-docs`
- **What you see:** the skill reads the PRD, MVP spec, accepted ADRs,
  and `CLAUDE.md`, then re-renders `README.md` and
  `design/ai-summary.md` from the kit's templates.
- **What you produce:** an updated `README.md` and
  `design/ai-summary.md` committed to `main`.

Re-run this whenever the project's shape has changed meaningfully — new
ADRs accepted, a major feature shipped, a pivot in scope. The `ai-summary.md`
in particular is what you paste into external AIs for second-opinion
design reviews.

---

## 3. After the first release: the steady-state loop

**Why this section exists.** §2 takes a new project from idea to its
first release. After that, the work doesn't stop — you keep adding
features, fixing bugs, and shipping versioned releases. Most projects
spend most of their time here. The steady-state loop reuses §2's
per-issue and per-milestone skills, but skips the one-off scoping
ceremonies at the front (§2.a, §2.c) unless you're starting a major
scope round.

### The loop

For each new piece of work after the first release:

1. **Pick the next item.** From `notes/feature-ideas.md` (§4), a
   stakeholder ask, a bug report, or a PR follow-up note.
2. **(Optional) Write an ADR.** Per §2.b — only if it's an
   architectural choice. Skip for trivial work (§8).
3. **File the issue.** `gh issue create` directly for one-offs;
   re-run `/issue-planner` only when adding a new milestone-worth
   of issues.
4. **Loop §2.d–§2.g.** `/prepare-issue` → `/claude-issue-executor`
   → `/pr-review-packager` → merge. One issue at a time.
5. **Cut the release.** Once enough merged work has accumulated, run
   §2.h–§2.j (audit milestone → summary → complete → changelog →
   release → docs sync).

### When to revisit §2.a (PRD/MVP scoping)

Only for **major scope rounds** — a v2 charter, a new product line,
a research project's next major question, a pivot in audience.
Day-to-day features add to the existing build-out plan; they do not
need a fresh `/prd-to-mvp` round. If you find yourself reaching for
`/prd-to-mvp` per feature, you have probably mis-scoped the original
MVP — see ADR-036's anti-pattern note in §2.c.

**Major feature updates: use `/feature-prd`, not a fresh PRD.** When a
mature project needs a major feature — a new capability area, a change
to original-PRD assumptions, work spanning multiple issues and ADRs —
do **not** overwrite `design/prd.md` and do not force it through
ordinary one-off issues. Run `/feature` (`/feature-prd`) to capture it
as an **additive PRD addendum** under `design/prd-addenda/NNN-*.md` that
extends the original PRD, names ADR impact, and decomposes the feature
into phased issues. The addendum then feeds the normal flow:
`/feature-prd` → `/adr-writer` → `/issue-planner` → `/prepare-issue` →
`/claude-issue-executor` → `/pr-review-packager`. See
ADR-049. One-off
bugs, small docs fixes, and self-contained issues stay on the ordinary
loop above — no addendum.

### Versioning your releases (ADR-026)

`/release` follows a documented MAJOR/MINOR/PATCH policy. Apply the
spirit of it to your project; what counts as "breaking" depends on
what your downstream users — other developers, deployment scripts,
API consumers, library users — rely on.

| Bump | Triggers | Example |
|---|---|---|
| **MAJOR** | Breaking changes for downstream users | Removed feature, renamed public API, schema change, CLI default change |
| **MINOR** | Additive — new capability that doesn't break existing usage | New feature, new endpoint, new optional CLI flag |
| **PATCH** | Fixes and refinements with no behavioural change | Bug fix, doc fix, internal refactor, ADR status flip |

`/release` reads accepted ADRs and merged PRs since the last tag,
suggests a bump, and waits for your confirmation. The full
kit-internal classification table — including edge cases like "heading
rename in a template that no skill parses" — is in
ADR-026. The same
logic applies to target projects, with your domain replacing the
kit's templates/skills/installer surface.

### Pre-1.0 note

The kit is past `v1.0`, so the standard pre-1.0 SemVer permission to
make breaking changes in MINOR releases does **not** apply here. For
your own project, decide once whether you want pre-1.0 latitude;
record the choice in your `CLAUDE.md` so it's not relitigated.

---

## 4. The idea backlog

**Why this section exists.** Not every idea is ready to be an ADR or
an issue. Some surface mid-session, some during stakeholder
conversations, some from PR follow-ups (the carry-forward branch in
§6 where there's no specific dependent issue). Without a parking
lot, ideas either get filed too early — cluttering the backlog with
half-baked work — or are forgotten entirely.

### Where ideas live

The kit ships `notes/feature-ideas.md`
as the parking lot. Each entry captures:

- **A sketch** — one paragraph on what the idea looks like.
- **Options in mind** — alternatives worth weighing, or "TBD."
- **Open questions** — what you would need to decide before writing
  the ADR.
- **Consequences to think through** — likely tradeoffs, things that
  get harder, maintenance burden.

Entries are short — a few sentences per field is plenty. The file is
ordered by ADR dependency: foundational decisions first, then
features that build on them.

### Status lifecycle

```
idea  →  ready-for-adr  →  adr-drafted  →  shipped  |  dropped
```

- `idea` — captured, not yet ripe.
- `ready-for-adr` — you've thought enough about the tradeoffs that an
  ADR can be written.
- `adr-drafted` — `/adr-writer` has produced a draft; link to the ADR
  file from the entry.
- `shipped` — the ADR was accepted and the work merged.
- `dropped` — the idea was rejected. The entry stays as a record so
  the same idea doesn't get re-litigated next time it surfaces.

### Target tags

- `v-next` — already a candidate for the next phase or scope round;
  will likely graduate to an ADR soon.
- `future` — later consideration. Bigger scope, or it would reopen
  an existing ADR.

### Graduating an idea

When an entry reaches `ready-for-adr`, hand it to `/adr-writer`. The
skill produces a draft ADR in `design/adr/`. From there it follows
§2.b → §2.c (or steady-state filing per §3) → §2.d–§2.j like any
other ADR. Update the entry's status to `adr-drafted` with a link
to the ADR file.

### Relationship to ADR-040 carry-forward (§6)

Carry-forward (§6) is for design questions that affect a **specific
upcoming issue** that's already filed or planned. The idea backlog is
for ideas that don't have a specific dependent issue yet — they're
parked until a use case sharpens them. The two are siblings: when a
session surfaces a question with no dependent issue, it goes into
`feature-ideas.md`, not into PR Notes for #N.

---

## 5. Across sessions: `design/state.md`, `/resume`, `/pause`

Long-running projects span many Claude Code sessions. Without a
canonical "where are we?" pointer, every fresh session reconstructs
state from `gh` calls and prompt files. The kit closes that gap with
a single committed artefact, `design/state.md`, plus two skills.

### The artefact

`design/state.md` is a small (under ~100 lines), committed pointer
holding six zones, each wrapped in marker fences so individual
skills can rewrite their zone idempotently without disturbing the
others:

| Zone | Holds | Updated by |
|---|---|---|
| `phase` | Current phase name or `single` | `/pause` |
| `in-flight` | Issue number, prompt, branch, status | `/prepare-issue`, `/claude-issue-executor`, `/pr-review-packager`, `/pause` |
| `recent` | Rolling list of the last 5 merged PRs | `/pr-review-packager`, `/pause` |
| `blockers` | One line per blocker, or `none` | `/pause` |
| `continue-here` | One paragraph naming the next concrete action | `/prepare-issue`, `/pr-review-packager`, `/pause` |
| `next-action` | Structured next step (`skill`, `args`, `preconditions`, `blocked-by`) — the executable complement to `continue-here`, read by `/resume` and `/start` (ADR-048; [workflow-control §4](workflow-control.md#4-finding-the-next-step)) | `/prepare-issue`, `/claude-issue-executor`, `/pr-review-packager`, `/pause` |

The format spec is [`templates/state-template.md`](../templates/state-template.md).
The kit ships the template; each target project instantiates its
own `design/state.md` (typically by running `/pause` for the first
time).

### Reading it: `/resume`

- **Run:** `/resume` at the start of a fresh session.
- **What you see:** a one-message brief — phase, in-flight issue,
  last five PRs, blockers, "continue here" — with no `gh` calls on
  the happy path.
- **Fallback:** if `design/state.md` is missing, empty, or looks
  suspect (in-flight issue already merged, file mtime stale),
  `/resume` falls back to `gh pr list` / `gh issue list` and
  recommends running `/pause` to seed the file.

### Refreshing it: `/pause`

- **Run:** `/pause` (or `/pause --handoff`) before a context reset,
  end of day, after a non-trivial detour, or to seed the file for
  the first time.
- **What you see:** each zone is recomputed from current truth
  (build-out plan for phase, prompt + branch for in-flight, `gh pr
  list --state merged` for recent, you for blockers and continue-
  here) and written back. With `--handoff`, an additional richer
  `notes/handoff-YYYY-MM-DD.md` is written for context-window-
  exhausting handoffs.
- **What you produce:** a refreshed `design/state.md` ready to
  commit.

### Conflict-resolution rule

`design/state.md` is committed, so parallel branches can produce
merge conflicts in it. The rule: **the most recently merged PR's
version wins for the conflicting zone.** If the result looks
suspect after the merge, `/resume` re-derives from `gh` and points
you at `/pause` to refresh.

### Optional CI line-cap check

`bin/check-state-cap` exits 1 if `design/state.md` exceeds the
line cap (default 100). Wire it into CI alongside `bin/sync-adr-index`
if you want a guard rail; it exits 0 silently when the file is
absent, so adopting it is no-op for projects that don't use
`design/state.md`.

## 6. Cross-skill carry-forward (ADR-040)

The implementation skill chain — `claude-issue-executor` →
`pr-review-packager` → `prepare-issue` — runs as a unit when a
session's eval summary raises a design question whose answer
belongs in an upcoming issue. ADR-040 codifies this carry-forward as
a kit guarantee: structured data flows producer → preserver →
consumer regardless of session author or chain-aware authoring.
This section is the canonical schema and rule set; the three skill
specs cross-reference it without restating.

### Canonical schema (version 1)

The machine-readable mirror of this subsection is
[`schemas/design-questions.v1.yaml`](../schemas/design-questions.v1.yaml):
it declares the required fields, the `target-issue` pattern, and the
no-extra-fields rule in structured form, and is what
[`bin/validate-carry-forward`](../bin/validate-carry-forward) enforces.
This prose and that file must agree — `bin/check-consistency` fails if
the file is not referenced here. When the schema changes, update both
(see [Schema versioning](#schema-versioning) below).

The carry-forward unit is one or more entries in a
`design-questions` block. The block lives under `## Follow-ups` of
the executor's eval summary file (`notes/eval-issue-NNN.md`) as a
fenced YAML code block under a `### design-questions` subheading:

````markdown
### design-questions

```yaml
- title: <one-line problem statement>
  target-issue: "#<N>"
  context: |
    <one-paragraph context note explaining the design question and
    why it affects the target issue specifically>
```
````

Field semantics:

- **`title`** — a one-line problem statement, written so a reader of
  the target issue's prompt can grasp the question without opening
  the source PR.
- **`target-issue`** — the issue number whose plan or implementation
  the question affects. Quoted (`"#5"`) so YAML does not lex the `#`
  as a comment marker. Reference one issue per entry; create
  separate entries if a single design question affects multiple
  upcoming issues.
- **`context`** — a one-paragraph block (YAML literal `|` so embedded
  newlines and markdown are preserved). Captures *why* the question
  matters for the target issue — the cross-issue coupling, not the
  question's general background.

The block is **omitted entirely** when the eval summary has no
entries (do not emit `design-questions: []`). This keeps clean
sessions clean.

### When to populate

Add a `design-questions` entry when *all three* of the following
hold for a question raised during the executor session:

1. The question concerns a **load-bearing constraint** that another
   upcoming issue will need to resolve (architecture, contract,
   shared schema, naming convention, file location, behaviour at a
   skill boundary).
2. The constraint is **shared with at least one specific upcoming
   issue** that is already filed or planned (i.e. you can name the
   target issue number — even if loosely).
3. The answer is **not fully determined by this issue's commits** —
   the upstream issue's executor saw the question but did not have
   the scope to answer it.

### When NOT to populate

Skip the `design-questions` entry — even if a design question came
up — when *any* of the following hold:

1. **Self-resolved.** The question was raised mid-session but the
   executor's commits resolve it within the current issue's scope.
   No follow-up reaches the target issue.
2. **No upcoming dependent issue.** The question is interesting but
   no specific filed-or-planned issue depends on the answer. Capture
   in `notes/feature-ideas.md` instead, where it will surface again
   when an issue is filed against it.
3. **Tactics, not architecture.** The question is purely about
   implementation tactics within one issue (whether to use map vs.
   filter, exact log message wording, internal helper-function
   factoring). These belong in the issue's commits or `## Follow-ups`
   prose, not the structured carry-forward.
4. **Already covered by an ADR or `design/decisions.md`.** The
   question has a documented answer — link to it from the issue's
   prompt instead. Carrying it forward as if undecided would create
   a duplicate authority.

If a question is borderline, prefer **omitting** the carry-forward
entry. False positives cost more (PR-body noise on every downstream
issue, prompt clutter for unaffected executors) than false
negatives (one extra round-trip when the question resurfaces in the
target issue's session).

### File and section names — single source of truth

| Stage | File / location | Format |
|---|---|---|
| Producer | `notes/eval-issue-NNN.md` (executor session output) | `### design-questions` heading + fenced YAML block under `## Follow-ups` |
| Preserver | PR body | One `## Notes for #<N>` section per unique `target-issue`; entries rendered as `- **<title>**: <context paragraph>` bullets |
| Consumer | Generated prompt | One `## Design questions carried forward from PR #M` subsection inserted before `## Requirements`, embedding the PR's Notes verbatim with an instruction to the executor to address each in its plan |

The three formats share the same canonical content (the YAML
fields) but render in the surface most natural for each stage:
machine-readable on the producer, human-readable on the preserver
and consumer. The **issue numbers are kit-wide** — `#69`, `#70`
etc. — there is no per-skill or per-stage renumbering.

### Schema versioning

The schema above is **version 1**. When it evolves (new fields,
renamed fields, semantic changes to existing fields), update **§6
first** as the canonical spec, then update its machine-readable mirror
[`schemas/design-questions.v1.yaml`](../schemas/design-questions.v1.yaml)
and the three SKILL.md cross-references in lockstep within the same
change. A breaking change adds `schemas/design-questions.v2.yaml`
beside the v1 file rather than editing v1 in place. Spec drift between
§6 and any one SKILL.md silently breaks the carry-forward loop;
`bin/validate-carry-forward` now mechanically checks producer blocks
against the schema file, and PR review remains the enforcement point
for the prose↔SKILL.md cross-references.

### `--no-prompt` interaction

The `--no-prompt` mode skips `prepare-issue` entirely on
genuinely-trivial issues (typo fixes, ADR status flips, dependency
bumps). Such issues by definition do not raise cross-issue design
questions, so bypassing the consumer stage is acceptable: a
`--no-prompt` invocation on a trivial issue with `## Notes for #N`
in a recently-merged PR will not embed those notes. If you find
yourself in that situation, reclassify the issue as non-trivial and
run `prepare-issue` explicitly.

### Pointers

- ADR: `design/adr/adr-040-cross-skill-design-question-carry-forward.md`
- Canonical schema file: [`schemas/design-questions.v1.yaml`](../schemas/design-questions.v1.yaml)
- Validator: [`bin/validate-carry-forward`](../bin/validate-carry-forward)
- Producer skill: [`skills/claude-issue-executor/SKILL.md`](../skills/claude-issue-executor/SKILL.md)
- Preserver skill: [`skills/pr-review-packager/SKILL.md`](../skills/pr-review-packager/SKILL.md)
- Consumer skill: [`skills/prepare-issue/SKILL.md`](../skills/prepare-issue/SKILL.md)
- Worked round-trip example: [`skills/pr-review-packager/example.md`](../skills/pr-review-packager/example.md) §7

## 7. Auto-mode permission contract (ADR-041)

Auto-mode reduces ceremony by letting the assistant proceed without
per-tool approval prompts. That is desirable for routine work and
dangerous for hard-to-reverse work. ADR-041 codifies a kit-wide
**permission contract** that names what auto-mode is allowed to
substitute for, what it is not, and how operators authorize
substitutions when the contract permits them. The contract turns
F24-class regressions (silent bypass of significant-task plan mode)
and F23-class regressions (strict-mode-vs-runtime mismatch in
`/pr-review-packager`) from author-discipline failures into
structural impossibilities.

This section is the single source of truth for the permission
**categories** — a property of each skill *operation*. The named
**operating modes** (`interactive` / `assisted` / `autonomous`) — a
property of the *session* — layer over these categories and are defined
in [`workflow-control.md` §1](workflow-control.md#1-operating-modes);
`assisted` and `autonomous` are the two shapes this section calls
"auto-mode". No mode ever relaxes cat-3. Every shipped skill's
`SKILL.md` declares its `permission-category` in front-matter and
cross-references this section without restating it. New skill
operations must be classified at merge time.

### The three categories

**Category 1 — Substitutable.** Auto-mode may proceed without
explicit operator approval. Local, reversible, low-blast-radius
operations live here. Examples: file reads, repo scans, lint
checks, format passes, ADR / `feature-ideas.md` status flips,
single-typo fixes, generation of artefacts that another step
reviews before publishing.

**Category 2 — Operator-acknowledged-bypass.** Auto-mode may proceed
but must explicitly state in the skill's chat output that the
bypass is happening, citing this section. The bypass is
operator-acknowledged, never silent. Examples: significant-task
plan mode (ADR-039) when the operator has pre-authorized auto-mode
for the session via an explicit toggle.

**Category 3 — Non-substitutable.** Auto-mode never substitutes for
explicit operator approval on these operations regardless of mode.
Public-visibility or hard-to-reverse blast radius. Examples:
`git push`, `gh pr create`, `gh release create`, `git tag`,
creating GitHub issues or comments, closing GitHub milestones,
running migrations, modifying `.claude/settings*.json`, modifying
`bin/*` scripts.

### Per-skill classification (canonical)

Exhaustive at time of writing. New skills are added to this table
in the same PR that ships the skill.

| Cat | Skill | Gating operation |
|---|---|---|
| 1 | `/adr-writer` | Drafts ADRs locally; user accepts manually |
| 1 | `/audit-milestone` | Read-only `gh` queries; advisory pass/fail report |
| 1 | `/changelog` | Renders to stdout / file / Release-body draft (publishing is `/release`'s job) |
| 1 | `/check-plan` | Validates an ADR or prompt against criteria; advisory only |
| 1 | `bin/check-plan` | Programmatic equivalent of `/check-plan`; read-only against the criteria list, no mutating side effects; invoked by chained skills under auto-mode |
| 1 | `/clarify` | Local conversation; appends to `design/decisions.md` |
| 1 | `/idea-to-prd` | Drafts a PRD locally |
| 1 | `/milestone-summary` | Writes `design/milestones/N-summary.md` locally |
| 1 | `/pause` | Refreshes local `design/state.md` |
| 1 | `/planning` | Writes `design/planning.md` locally |
| 1 | `/prd-normalizer` | Local doc rewrite |
| 1 | `/prd-to-mvp` | Local doc creation (`design/mvp.md`, `design/build-out-plan.md`) |
| 1 | `/prepare-issue` | Reads `gh` and ADRs (non-mutating); writes prompt file locally |
| 1 | `/resume` | Reads `design/state.md` and emits a summary; falls back to `gh` reads |
| 1 | `/workflow-docs` | Generates `README.md`, `design/architecture.md`, and `design/ai-summary.md` locally |
| **2** | `/claude-issue-executor` | Significant-task plan-mode gate (ADR-039); local file edits and `git` commits but no push |
| **3** | `/complete-milestone` | Closes a GitHub milestone via `gh`; chains `/release` when `--release` is passed |
| **3** | `/issue-planner` | Calls `gh issue create` for each issue and creates a Project board |
| **3** | `/pr-review-packager` | Calls `gh pr create` — public-visibility, hard-to-reverse |
| **3** | `/release` | `git tag`, `git push`, `gh release create` — maximum public visibility |

### Category 2 — significant-task plan mode (ADR-039 instance)

The cat-2 reference instance is `/claude-issue-executor`'s
significant-task plan-mode gate. The canonical "significant"
checklist lives in
[`skills/claude-issue-executor/SKILL.md`](../skills/claude-issue-executor/SKILL.md)
and is the per-skill instance of this category. The two checklists
(this category's framing and ADR-039's checklist) must move
together: when the contract category text is amended, ADR-039's
checklist is reviewed in the same change, and vice versa.
**Alignment is enforced at PR review time** — there is no
machine-checkable enforcement until ADR-034's plan-checker
structural-rule framework ships (see §6 of this guide and the
*Schema-drift enforcement* subsection below).

The cat-2 contract requires the executor to *ask once* at session
start under auto-mode rather than auto-classify silently:

- *"Enter plan mode for this task? yes / no / decide-from-scope."*
- `yes` / `decide-from-scope` → use the significance checklist as
  documented.
- `no` → write a one-line acknowledgement in chat output before any
  mutating edit: *"Plan mode bypassed by operator (cat-2
  operator-acknowledged bypass per workflow-guide §7)."* The
  acknowledgement is mandatory; without it the bypass is silent and
  the cat-2 contract is violated.

### Category 3 — PR creation, releases, GitHub state changes

The cat-3 reference instance is `/pr-review-packager`'s explicit-yes
gate before `gh pr create`. The skill has always asked for explicit
`yes` before opening a PR (steps 12–13 of its execution protocol);
ADR-041 promotes that convention to a contract: **explicit `yes` is
required regardless of mode; auto-mode never substitutes for it**.
The same rule applies to `/release`, `/issue-planner`, and
`/complete-milestone` — all four cat-3 skills require an explicit
operator gate that auto-mode does not satisfy.

If a future skill adds an operation in this category — anything
public-visibility or hard-to-reverse — its `SKILL.md` MUST declare
`permission-category: 3` in front-matter, cross-reference this
section, and gate the operation behind an explicit-yes prompt that
auto-mode cannot substitute for.

### `--no-prompt` interaction

`--no-prompt` mode (see §2.e) skips prompt generation on
genuinely-trivial issues. It is itself a **category-1
operator-pre-authorization for the trivial-issue path**: the
operator has named the issue trivial by passing the flag, and
trivial issues by definition do not raise cross-issue design
questions or touch hard-to-reverse operations.

`--no-prompt` does **not** bypass category-3 operations. Even under
`--no-prompt`, the executor cannot push, tag, or open a PR without
explicit approval. The handoff to `/pr-review-packager` (cat-3) is
unchanged; `--no-prompt` only affects the prompt artefact.

**Future-proofing note.** If a future skill or mode introduces
*non-trivial* prompt-generation bypass paths (e.g. CI-driven
implementations, bulk migrations), that mode must be classified at
the time it is added — likely category 2, since cross-issue
continuity (§6 carry-forward) is too important to let auto-mode
silently skip on non-trivial work. The current `--no-prompt` flag's
cat-1 classification is correct because its eligibility criteria
are exactly the trivial checklist — it cannot reach non-trivial
work by construction.

### Schema-drift enforcement

The contract above (categories, classifications, cross-references)
is enforced at **PR review** today. A skill spec that classifies a
category-3 operation under auto-mode, or a front-matter
`permission-category` line that drifts from the canonical table,
must be caught at review time.

Machine-checkable enforcement is deferred to issue #72 (ADR-034
plan-checker), where the generic structural-rule framework will
also subsume §6's design-questions schema-drift check (per
ADR-040's *Maintain* paragraph). Building a single-purpose
permission-contract checker before #72 lands the generic framework
would create a throwaway artefact and pre-empt the better
solution.

### Pointers

- ADR: `design/adr/adr-041-auto-mode-permission-contract.md`
- Cat-2 instance: [`skills/claude-issue-executor/SKILL.md`](../skills/claude-issue-executor/SKILL.md) — *Plan-mode rhythm* section
- Cat-3 instance: [`skills/pr-review-packager/SKILL.md`](../skills/pr-review-packager/SKILL.md) — *Auto-mode permission category* section
- Related: ADR-039 (per-skill rule that ADR-041 generalises), ADR-038 (`--no-prompt` interaction), ADR-040 (schema-drift home alignment)

---

## 8. When you don't need an ADR

Not every change is an architectural decision. Bug fixes, chores,
dependency bumps, CI/CD tweaks, and doc-only changes generally do not
warrant an ADR — forcing one creates ceremony without insight.

**ADR-free changes still follow the same GitHub flow:**

| Stage | ADR-driven work | ADR-free work |
|---|---|---|
| Decision | Write ADR, accept | Skip |
| Issue | `/issue-planner` or manually via `gh issue create` | Manually via `gh issue create` |
| Branch | Feature branch from `main` | Feature branch from `main` |
| Prompt | `/prepare-issue <N>` | `/prepare-issue <N>` (still fills the template; ADR section is marked "none") |
| Build | `/claude-issue-executor` | `/claude-issue-executor` |
| PR | `/pr-review-packager` | `/pr-review-packager` |
| Merge | `gh pr merge --squash` | `gh pr merge --squash` |

So: still need an issue. Still need a branch. Still need a PR. You just
skip `/adr-writer` and mark the ADR section of the issue as "none".

The kit's issue template explicitly supports this — see the HTML
comment under the `## ADR` heading in
[`templates/issue-template.md`](../templates/issue-template.md): *"If
no ADR applies, write 'none' and say why."*

**Examples of ADR-free work:**

| Change | Why no ADR |
|---|---|
| Bug fix in an existing module | Restoring intended behaviour, not deciding new architecture |
| Dependency version bump | Unless it forces a behaviour change, it's a chore |
| CI/CD tweak (add a lint step, change a runner) | Operational, not architectural |
| README typo fix, docs rephrasing | Doc-only |
| Refactor with no behaviour change | Unless it reshapes the architecture, it's cleanup |
| Config tweak (log level, timeouts) | Tuning, not deciding |

If you are unsure, err on the side of writing an ADR — they are cheap,
and a rejected ADR is still useful as a record of what was considered.

---

## 9. Where to go deeper

| You want to… | Go to |
|---|---|
| Learn how to run Claude Code skills from inside a target project | [`claude-code-guide.md`](claude-code-guide.md) |
| Be told which skill to run next, or learn the verb layer and operating modes | `/start`, or [`workflow-control.md`](workflow-control.md) |
| Understand a specific skill's interface, inputs, and outputs | `.claude/skills/<name>/SKILL.md` in your target project, or [`skills/README.md`](../skills/README.md) in the kit |
| Revisit a design decision that shaped this workflow | `design/adr/` |
| Set up GitHub labels, milestones, and branch protection | [`docs/github-setup.md`](github-setup.md) |

If the flow above doesn't match the skills you have installed, check
[`skills/README.md`](../skills/README.md) for the current inventory —
some skills land in later milestones and the guide lists the full v-next
set.

