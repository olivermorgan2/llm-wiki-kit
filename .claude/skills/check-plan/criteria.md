<!--
  Document: /check-plan criteria
  Filled by: humans (this is hand-curated; not generated)
  Used by: skills/check-plan/SKILL.md
  Source-of-truth status: this file IS the source of truth for what
    /check-plan validates. The SKILL.md references criteria here by
    stable IDs (e.g. ADR-C1, PROMPT-C3) and never duplicates the
    text. Adding a criterion = appending a new row; renaming a
    criterion is FORBIDDEN — use a new ID instead.

  Version-lock contract (per ADR-034). When templates/adr-template.md
  or prompts/_template.md changes, this file must be reviewed and
  updated in the same PR. The optional bin/check-plan-criteria-drift
  script flags drift via mtime comparison; CI can wire that as a
  warning if desired. The contract holds even without the script.

  Determinism flag. Each criterion is either:
    - DETERMINISTIC — a structural / textual check that can be made
      pass/fail with no judgement (e.g. "section X exists", "no
      {{...}} placeholders left"). A failure on a deterministic
      criterion blocks the gate.
    - WARNING — a heuristic or judgement check whose result is
      surfaced to the user but does not block the gate (e.g. "fits
      build-out-plan phase", "no apparent conflict with accepted
      ADRs"). Warnings are advisory; the user decides.
  Per ADR-034: the checker emits warnings (not hard errors) for
  dimensions it cannot verify deterministically.
-->

# /check-plan — Criteria

This file lists every check `/check-plan` runs. Each criterion has a
**stable ID**, a one-line **what to check**, a **determinism flag**
(deterministic | warning), and a **fix hint** the skill prints when
the criterion fails.

The two checklists below are version-locked to two upstream files:

| Checklist | Validates | Locked to |
|---|---|---|
| ADR criteria | `design/adr/adr-NNN-*.md` | `templates/adr-template.md` |
| Prompt criteria | `prompts/issue-NNN-*.md` | `prompts/_template.md` |

When either upstream file changes, audit the matching checklist and
update or extend it in the same PR. `bin/check-plan-criteria-drift`
flags mtime drift; ADR-034's "Maintain" line is the authoritative
contract.

## ADR criteria

Validates ADRs against `templates/adr-template.md`. Per ADR-034:
"clear context / decision / consequences; references the right ADRs;
doesn't conflict with accepted ones; options have both pros and
cons; decision names one of the listed options."

| ID | Determinism | What to check | Fix hint on fail |
|---|---|---|---|
| `ADR-C1` | deterministic | The ADR has all of `## Context`, `## Decision`, `## Consequences` headings, each with non-empty body. | Add the missing section. The template orders them: Context → Options considered → Decision → Consequences. |
| `ADR-C2` | deterministic | `## Options considered` contains at least two `### Option ` sub-headings. | Add a second option or remove the Options-considered section if the decision is genuinely uncontested (rare; ADRs without alternatives usually mask undisclosed trade-offs). |
| `ADR-C3` | deterministic | Every `### Option ...` block has both a `Pros:` line and a `Cons:` line (matching the template's `- Pros:` / `- Cons:` shape). | For each option missing one, add the line and at least one bullet. The template's "Option A: …" block is the reference shape. |
| `ADR-C4` | deterministic | The `## Decision` body names one of the option labels declared in `## Options considered` (e.g. "Adopt **Option B**" matches `### Option B:`). | Either name an option explicitly in the Decision, or rename the chosen option to match. |
| `ADR-C5` | warning | Every `ADR-NNN` token in the body resolves to a file in `design/adr/`. | Listed unresolved IDs; either fix the token (typo) or create the missing ADR. Warning only — sometimes ADRs reference ones not yet drafted. |
| `ADR-C6` | warning | The Decision does not textually contradict an accepted ADR (best-effort substring check against accepted-ADR Decision sections). | Listed candidate-conflict ADRs. The user is the final judge — if the contradiction is intentional, this ADR should declare itself a *supersedes* of the older one. |

## Prompt criteria

Validates issue prompts against `prompts/_template.md` (per ADR-008).
Per ADR-034: "acceptance criteria present; correct ADR links; fits
the build-out-plan phase (per ADR-032 if shipped); single-PR scope;
no ambiguous TBD placeholders."

| ID | Determinism | What to check | Fix hint on fail |
|---|---|---|---|
| `PROMPT-C1` | deterministic | The `Acceptance criteria` section is present and contains at least one bullet. | Add an Acceptance criteria section listing observable end-state outcomes (not implementation tasks). The template's example is the reference shape. |
| `PROMPT-C2` | deterministic | The `ADR:` section is present and either resolves to a file in `design/adr/` (via `File:` line) or explicitly states `ADR: none` with a reason. | Either link the governing ADR, or replace the section with `ADR: none — <reason>` per the template's HTML-comment guidance. |
| `PROMPT-C3` | deterministic | No remaining `{{...}}` template placeholders. No unresolved `<!-- TODO: fill in -->` markers. | List the unfilled slots with line numbers; fill or explicitly delete the line if the slot is irrelevant to this issue. |
| `PROMPT-C4` | warning | When `design/build-out-plan.md` exists with `## Phase N` blocks, the prompt's stated scope fits within one of those phases. When the build-out-plan is absent or single-phase, this check is silently skipped (not warned). | Listed phase candidates the scope spans. Either narrow the scope to one phase, or split into multiple issues. |
| `PROMPT-C5` | warning | Heuristic single-PR scope: combined `Requirements` + `Scope and constraints` body fits a soft cap (≤ ~25 bullets total, ≤ ~80 lines). Exceeds → warning. | Either split the issue, or accept the warning if the work is genuinely cohesive (e.g. a single skill add with many small artefacts, like ADR-035's #44). |
| `PROMPT-C6` | deterministic | All required sections from `prompts/_template.md` are present: Context, ADR, GitHub Issue, Goal, Why it matters, Requirements, Acceptance criteria, Scope and constraints, Evaluation & testing requirements, Instructions for you. | Add the missing sections. Re-run `/prepare-issue` if the prompt was hand-authored and drifted from the template. |

## How `/check-plan` reads this file

The skill loads this file once per invocation and parses each table
row into a criterion record. The skill's protocol references rows
by ID, never by row position — so reordering rows is safe.

Adding a criterion: append a new row in the appropriate table with
the next available ID (e.g. `ADR-C7`). Update `skills/check-plan/SKILL.md`
only if the new criterion needs special handling beyond the
table-row contract; otherwise the SKILL.md picks it up automatically.

Removing a criterion: leave the ID gap (e.g. delete `ADR-C2` →
later `ADR-C7` is `C7`, not the gap-filler). The gap signals the
historical removal and prevents accidental ID reuse.

Renaming an ID is **forbidden**. Tooling and any external references
(commit messages, PR bodies citing a failure) treat IDs as stable
identifiers.
