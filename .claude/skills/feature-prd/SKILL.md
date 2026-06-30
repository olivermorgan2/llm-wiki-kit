---
name: feature-prd
description: Capture a major feature update to a mature project as an additive PRD addendum under design/prd-addenda/, extending the original PRD rather than replacing it. Identifies ADR impact and decomposes the feature into phased issues. Use when a change adds a capability area, alters original-PRD assumptions, or needs multiple issues/ADRs — not for one-off bugs or small fixes.
permission-category: 1  # substitutable — drafts a local addendum file reviewed before any gh action, per workflow-guide §7
inputs:
  - name: "feature"
    required: true
    description: "The major feature or capability area to plan (a description, notes, or a roadmap reference)"
outputs:
  - artefact: "design/prd-addenda/NNN-feature-name.md"
    description: "Numbered PRD addendum extending design/prd.md"
next:
  - skill: adr-writer
    when: "the addendum surfaces decisions needing ADRs"
  - skill: issue-planner
    when: "the decomposition feeds the GitHub backlog"
---

# feature-prd

Plan a **major feature update** to a project the kit already owns. The
output is an additive PRD addendum at
`design/prd-addenda/NNN-feature-name.md` that **extends** the original
`design/prd.md` — it never overwrites it. This is the expansion path
that complements the first-definition intake flow (`idea-to-prd` →
`prd-normalizer` → `prd-to-mvp`), governed by
ADR-049. It is the
`/feature` verb in the
[verb layer](../../docs/workflow-control.md#3-the-human-facing-verb-layer).

## When to use this skill

Use a follow-up PRD when the change:

- adds a **new capability area** rather than a small enhancement,
- **changes assumptions** baked into the original PRD,
- needs **multiple issues or phases**,
- requires **one or more ADRs**, or
- introduces **compatibility, migration, or permission-boundary**
  concerns.

## When NOT to use this skill

Do **not** use a follow-up PRD for one-off bugs, small docs fixes,
obvious refactors, or a single self-contained issue. Those stay on the
ordinary steady-state loop ([`workflow-guide.md` §3](../../docs/workflow-guide.md#3-after-the-first-release-the-steady-state-loop)):
file an issue and run `/prepare-issue` → `/claude-issue-executor` →
`/pr-review-packager`. Forcing an addendum onto small work is ceremony
without insight; forcing a large feature through ordinary issues
under-specifies it. If unsure which side a change falls on, it is a
follow-up PRD when you cannot name its scope in a single issue title.

## Inputs

- **Required:** a description of the feature — prose, notes, or a
  pointer to a roadmap entry.
- **Read (never modified):** `design/prd.md` (the original definition),
  the accepted ADRs in `design/adr/`, `design/build-out-plan.md`, and
  `design/state.md` for current milestone/state. The original PRD is
  **read-only** to this skill.

## Output

- **File:** `design/prd-addenda/NNN-<feature-name>.md`, where `NNN` is
  the next free sequential number in `design/prd-addenda/` (create the
  directory if absent; start at `001`). `<feature-name>` is a short
  kebab-case slug.
- **Shape:** the sections in [`templates/prd-addendum-template.md`](../../templates/prd-addendum-template.md):
  Context (original PRD, current milestone/state, trigger), Problem,
  Goals, Non-goals, What changes, **What does not change**, Affected
  assumptions, ADR impact, User stories, Requirements, Migration and
  compatibility notes, Issue decomposition, Success metrics, Open
  questions.

## What the addendum must establish

1. **What does not change.** An explicit section preserving existing
   scope, so reviewers see the feature is additive. Do not leave it
   empty — name the parts of the product the feature deliberately
   leaves alone.
2. **ADR impact.** Identify which existing ADRs the feature requires to
   be **created**, **revised**, or **explicitly superseded**. Per the
   kit's rule, ADRs are never edited in place — a revision means a new
   ADR that supersedes the old one.
3. **Affected assumptions.** Name the original-PRD assumptions the
   feature changes, so the divergence from the founding definition is
   recorded rather than silent.
4. **Issue decomposition.** Break the feature into phased issues with
   dependencies and per-issue non-goals, so the addendum feeds
   `/issue-planner` directly.

## Protocol

1. **Read** `design/prd.md`, accepted ADRs, the build-out plan, and
   `design/state.md`. Establish the current milestone/state for the
   Context section.
2. **Draft** the addendum from the template, filling every section.
   Mark genuine unknowns as `[TBD — ask user]` and batch them into at
   most five questions per turn; skip TBDs with an obvious default.
3. **Resolve ADR impact** — list each affected ADR with the action
   (create / revise-via-supersession / no change but relevant).
4. **Decompose** into phased issues with dependencies and non-goals.
5. **Run the self-check** below; iterate steps 2–4 until it passes.
6. **Write** `design/prd-addenda/NNN-<feature-name>.md` and report the
   path, the ADR impact summary, and the recommended next step.

## Self-check before writing

- [ ] The original `design/prd.md` was read and is **not** modified.
- [ ] "What does not change" is non-empty and concrete.
- [ ] ADR impact names each affected ADR and its action.
- [ ] Affected assumptions are listed (or explicitly "none").
- [ ] The feature is decomposed into at least one phased issue with
  named non-goals.
- [ ] No `[TBD — …]` placeholders remain.

## Handoff

The addendum feeds the existing machinery unchanged. Recommend the next
step from the ADR impact:

- decisions need ADRs → `/adr-writer` (then `/check-plan`),
- the decomposition is ready for the backlog → `/issue-planner`.

Recommended full flow:

```text
/feature-prd → /adr-writer → /issue-planner → /prepare-issue
            → /claude-issue-executor → /pr-review-packager
```

See [`example.md`](example.md) for a worked feature → addendum pair.
