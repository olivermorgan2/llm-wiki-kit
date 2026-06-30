<!--
  Template: Build-out plan
  Filled by: the prd-to-mvp skill (alongside the MVP statement), or a human
  Output in a target project: design/build-out-plan.md (recommended)
  The build-out plan sequences the MVP into phases, milestones, and an
  initial issue backlog. It should be lightweight enough to revise
  mid-project and concrete enough to drive GitHub issue creation.

  Phases (per ADR-032). The "## Phases" section below uses richer
  per-phase blocks: goal, scope bullets, ADR dependencies, and an exit
  criterion. Downstream skills read this:
    - issue-planner creates one GitHub milestone per phase.
    - workflow-docs surfaces phases in the README roadmap.
    - /release defaults to one phase per release.
    - adr-writer respects each ADR's optional `**Phase:**` frontmatter
      so the ADR index gains a Phase column when any ADR is phased.

  Single-phase fallback. A plan with only one `## Phase` section — or
  with no `## Phase` headings at all — is treated as one implicit
  phase by every downstream skill. Small projects can keep using the
  flat structure unchanged. See docs/workflow-guide.md for the
  fallback semantics.

  Granularity (per ADR-036). The `**Granularity:**` line below records
  the tier chosen by `prd-to-mvp` / `/planning` and is the canonical
  store for re-runs: explicit `--granularity=<tier>` flag wins, else
  the value below is read, else default `standard`. Allowed values:
  `coarse` (1–3 phases), `standard` (5–8, default), `fine` (8–12).
  Bands are targets, not hard caps — the planning skill picks the
  actual count and justifies it inline. Absence of the line is
  treated as `standard` for backwards compatibility.
-->

# {{PRODUCT_NAME}} — Build-Out Plan

**Last updated:** {{YYYY-MM-DD}}
**Granularity:** {{GRANULARITY}}

## Objective

{{One paragraph: what this build-out plan covers and what it produces
by the end. Link back to the MVP document.}}

## Build strategy

{{Describe the order of work in 4–8 steps. Each step should be an
outcome, not a task. Example:
1. Define repo structure and decisions.
2. Write foundational documentation.
3. Implement core capability A.
4. Implement core capability B.
5. Dry-run on a sample input.
6. Package for release.}}

## Scope

- In scope: {{what this plan commits to delivering}}
- Out of scope: {{what this plan explicitly defers}}
- Assumptions: {{anything the plan depends on — tools, access, prior decisions}}

## Success criteria

The plan is complete when a user can:

1. {{End-to-end outcome 1.}}
2. {{End-to-end outcome 2.}}
3. {{End-to-end outcome 3.}}

## Repository structure

```text
{{PROJECT_NAME}}/
  {{top-level folders with one-line purposes}}
```

## Phases

<!-- Each phase below is a delivery unit. issue-planner creates one
     GitHub milestone per phase; /release treats one phase as the
     default release boundary. Single-phase projects keep one block. -->

### Phase 1: {{NAME}}

- **Goal:** {{one line}}
- **Scope:**
  - {{in-scope item 1}}
  - {{in-scope item 2}}
  - {{in-scope item 3}}
- **ADR dependencies:** {{ADR-NNN, ADR-NNN, ... or "none"}}
- **Deliverables:** {{2–4 concrete artifacts}}
- **Exit criteria:** {{how you know the phase is done — observable, not a task list}}

### Phase 2: {{NAME}}

- **Goal:** {{one line}}
- **Scope:**
  - {{in-scope item 1}}
  - {{in-scope item 2}}
- **ADR dependencies:** {{...}}
- **Deliverables:** {{...}}
- **Exit criteria:** {{...}}

<!-- Add Phase 3+ as needed. Keep each phase to one readable block.
     For single-phase projects, keep just Phase 1 — every downstream
     skill treats a single-phase plan identically to a flat plan. -->

<!-- Phased ADR linkage. ADRs that belong to a specific phase should
     declare it via the optional `**Phase:**` frontmatter line in their
     ADR file. bin/sync-adr-index will surface a Phase column in
     design/adr/README.md whenever any ADR is phase-tagged. -->

## Milestone recommendation

| Milestone | Focus |
|---|---|
| {{M1}} | {{what this milestone is for}} |
| {{M2}} | {{...}} |
| {{M3}} | {{...}} |

## Initial issue backlog

### {{Milestone 1}}

- {{Issue title 1}}
- {{Issue title 2}}
- {{Issue title 3}}

### {{Milestone 2}}

- {{Issue title 1}}
- {{Issue title 2}}

<!-- The issue-planner skill (or a human) turns each of these into a
full GitHub issue body using templates/issue-template.md. -->

## Testing strategy

{{How the deliverables are validated — unit tests, integration tests,
dry-run walkthroughs, manual verification. Match this to the nature of
the project; not every project needs every category.}}

## Risks and mitigations

### Risk 1 — {{NAME}}

Mitigation: {{one line}}

### Risk 2 — {{NAME}}

Mitigation: {{one line}}

## Acceptance criteria for this document

This build-out plan is acceptable when it:

- matches the MVP statement,
- sequences work in realistic phases,
- identifies initial ADRs or decisions,
- and produces a practical milestone and issue structure.
