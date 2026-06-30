<!--
  Issue session prompt template — one issue, one session.

  How to use this file:
  1. Copy it to `prompts/issue-NNN-short-title.md` (zero-padded issue number,
     kebab-case title; e.g. `prompts/issue-012-prompts-folder.md`).
  2. Replace every `{{PLACEHOLDER}}` with the specific value for this issue.
     Delete any optional lines that do not apply.
  3. Paste the filled file into a fresh Claude Code session.
  4. Keep the filled prompt in `prompts/` — it is part of the project's
     history and makes it easy to resume or re-run a session later.

  Underlying decisions:
  - ADR-006 (design/adr/adr-006-claude-code-execution-model.md) — plan-first,
    one issue per session.
  - ADR-008 (design/adr/adr-008-dedicated-prompts-folder.md) — prompts live
    in `prompts/`, not `notes/`. The `notes/` directory is for freeform
    working notes only.

  Section reference (keep this order): Context, ADR, GitHub Issue, Goal,
  Requirements, Acceptance criteria, Scope, Evaluation, Instructions.

  Placeholder reference (when filling by hand instead of via
  `/prepare-issue`). The GitHub-issue placeholders are self-evident
  and omitted; everything else maps as follows:
    | Placeholder | Source |
    |---|---|
    | {{PROJECT_NAME}} | The repo name. |
    | {{ONE_LINE_PROJECT_DESCRIPTION}} | The one-line description in `design/mvp.md` (or `design/prd-normalized.md` if no MVP yet). |
    | {{WORKFLOW_DOC_PATH}} | Usually `docs/workflow-guide.md`. For repos that customise this, point at the actual workflow doc. |
    | {{ADR_FILE}} / {{ADR_ONE_LINE_SUMMARY}} | The ADR(s) driving this issue. Repeat the two-line block per ADR. If none, write `ADR: none — {{REASON}}`. |
    | {{PRIMARY_FOLDERS}} / {{AVOID_FOLDERS}} | Folders the work should and should not touch. Be specific; the avoid list prevents scope creep. |

  Content boundary (ADR-038, audited in notes/adr-038-alignment-review.md).
  The prompt is per-issue and immutable — written once by `/prepare-issue`,
  consumed once by `/claude-issue-executor`. Project- or session-scoped
  artefacts are read-only context here; do not duplicate their content into
  the prompt:
    - design/planning.md (ADR-031): project-wide decomposition, risks,
      sequencing rationale. Reference; do not restate.
    - design/build-out-plan.md `## Phase N` blocks (ADR-032): phase scope,
      deliverables, exit criteria. Reference by phase name.
    - design/decisions.md (ADR-033): informal-but-settled decisions
      below ADR weight. Reference when a decision constrains this issue.
    - design/state.md (ADR-035): session-mutable pointer (in-flight
      issue, recent PRs, blockers). Never mirror it here.
    - skills/check-plan/criteria.md (ADR-034): structural validation
      rules. Orthogonal to this prompt's Acceptance criteria (which
      capture end-state outcomes, not validation rules).
    - /milestone-summary outputs (ADR-037): cross-issue retrospectives.
      Out of scope for a per-issue prompt.
-->

You are working in my `{{PROJECT_NAME}}` repository.

Context:
- {{ONE_LINE_PROJECT_DESCRIPTION}}
- Follow the rules in `CLAUDE.md`.
- The workflow model is described in `{{WORKFLOW_DOC_PATH}}`.

ADR:
- File: `design/adr/{{ADR_FILE}}`
- Decision: {{ADR_ONE_LINE_SUMMARY}}
<!-- If multiple ADRs apply, repeat the two lines per ADR.
     If no ADR applies, replace this section with: "ADR: none — {{REASON}}." -->

GitHub Issue:
- Title: {{ISSUE_TITLE}}
- Number: #{{ISSUE_NUMBER}}
- Milestone: {{MILESTONE}}
- Labels: {{LABELS}}

Goal
{{ONE_OR_TWO_SENTENCES}}

Why it matters
{{ONE_PARAGRAPH}}

Requirements
- {{REQUIREMENT_1}}
- {{REQUIREMENT_2}}
- {{REQUIREMENT_3}}

Acceptance criteria
- {{CRITERION_1}}
- {{CRITERION_2}}
- {{CRITERION_3}}

Scope and constraints
- Primary folders to touch: {{PRIMARY_FOLDERS}}
- Folders to avoid unless absolutely necessary: {{AVOID_FOLDERS}}
- {{PROJECT_SPECIFIC_CONSTRAINT_OR_DELETE_THIS_LINE}}

Evaluation & testing requirements
- {{EVALUATION_REQUIREMENT_1}}
- {{EVALUATION_REQUIREMENT_2}}
- All existing tests must continue to pass.
- If a change cannot be unit tested, document the manual verification.

Instructions for you
1. Read the relevant docs and existing files:
   - `CLAUDE.md`
   - `design/adr/{{ADR_FILE}}`
   - any existing modules under {{PRIMARY_FOLDERS}}
   - any existing tests related to the modules you will change
2. Propose a short, step-by-step implementation PLAN for this issue, including:
   - new files or modules to create,
   - existing files to modify,
   - key functions or structures,
   - your verification or test plan.
3. Wait for my approval of the plan before making any edits.
4. After I approve, implement the plan:
   - keep changes focused on this issue's scope,
   - commit incrementally with messages referencing the ADR and issue
     (e.g. "feat(scope): add thing (ADR-NNN, #NN)").
5. At the end, provide an evaluation summary:
   - what changed,
   - verification steps performed,
   - any follow-up work needed for later issues,
   - exact commands I should run to inspect the result myself.

Do not start editing files until I explicitly approve your plan.
