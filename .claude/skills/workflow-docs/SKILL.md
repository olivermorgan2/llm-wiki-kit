---
name: workflow-docs
description: Generate README.md, design/architecture.md, and design/ai-summary.md for a target project from PRD, MVP, ADRs, planning docs, and CLAUDE.md — re-runnable with marker-fenced sections that preserve manual edits. Use when generating or refreshing the project's user-facing and AI-facing docs.
permission-category: 1  # substitutable — generates README.md, design/architecture.md, and design/ai-summary.md locally, per workflow-guide §7
inputs: []
outputs:
  - artefact: "README.md"
    description: "Generated (marker-fenced)"
  - artefact: "design/architecture.md"
    description: "Current architecture/design reference (marker-fenced)"
  - artefact: "design/ai-summary.md"
    description: "AI-readable project summary (marker-fenced)"
next: []
---

# workflow-docs

Generate three project-level documents in a target project by reading the
artifacts the kit already produces and filling templates:

- `README.md` (public entry point for humans)
- `design/architecture.md` (current architecture/design reference)
- `design/ai-summary.md` (AI-readable project summary)

The skill is **re-runnable**. Generated regions are wrapped in comment
markers so a second run updates those regions without clobbering manual
edits outside them. Sections with no source data are **omitted**
entirely (heading + body) rather than left blank or filled with
placeholders.

See ADR-018 for the
decision to consolidate all outputs into one skill instead of two.

## When to use this skill

- After `prd-to-mvp` has produced `design/mvp.md`, to generate a real
  README, current architecture reference, and AI summary.
- After landing one or more ADRs, to refresh the current architecture and
  "Key decisions" sections in generated docs.
- Whenever `CLAUDE.md`, the PRD, the MVP statement, planning docs, or the
  implemented architecture has materially changed and the generated docs
  should reflect that change.

Typical invocation:

```
/workflow-docs
```

No arguments. The skill infers everything from files in the current
working directory (the target project root).

## What this skill does not do

- Does not modify the PRD, MVP statement, ADRs, or `CLAUDE.md`. These
  are **read-only inputs**.
- Does not invent content. A missing section source → the section is
  omitted, not filled with a placeholder or TODO.
- Does not overwrite files without explicit user approval. All generated
  files are shown in full before any write.
- Does not publish, deploy, or push. Output is local files only.
- Does not touch content outside its `<!-- workflow-docs:start:... -->`
  / `<!-- workflow-docs:end:... -->` markers on re-run.
- Does not manage CLAUDE.md — that is rendered from
  `templates/claude-md-template.md` at install time or by a human.

## Inputs (read-only)

All inputs live in the **target project** (the current working
directory when the skill runs). Every input is optional — the skill
degrades gracefully. Missing inputs cause their derived sections to be
omitted.

| Source                | What it contributes                                          |
|-----------------------|--------------------------------------------------------------|
| `design/prd.md` or `design/prd-normalized.md` | Project description, primary user, problem statement |
| `design/mvp.md`       | Product name, tagline, principles, In-scope, Out-of-scope, success criteria |
| `design/adr/*.md`     | "Key decisions" bullets — one line per ADR, newest first     |
| `design/build-out-plan.md` | Roadmap and phase context for architecture status           |
| `design/planning.md` / `design/decisions.md` | Current architecture context, constraints, open questions |
| `design/milestones/*.md` / `design/state.md` | Recent evolution and in-flight architecture status        |
| `CLAUDE.md`           | Tech stack, commands (install, dev, test), current phase, milestone |

If **none** of these exist, stop with a clear message pointing the user
at `prd-normalizer` / `prd-to-mvp` first. There is nothing useful to
generate from an empty kit.

## Outputs

Outputs are written to the **target project** root, next to the
inputs:

- `README.md` — rendered from
  [`templates/readme-template.md`](../../templates/readme-template.md).
- `design/architecture.md` — rendered from
  [`templates/architecture-template.md`](../../templates/architecture-template.md).
- `design/ai-summary.md` — rendered from
  [`templates/ai-summary-template.md`](../../templates/ai-summary-template.md).

On a first run, whole files are written. On a re-run, only the content
inside `<!-- workflow-docs:start:<section> -->` / `<!-- workflow-docs:end:<section> -->`
markers is regenerated; any other lines the user has added are left
untouched.

## Marker format (exact)

Every generated section is wrapped by two HTML comment markers on their
own lines:

```markdown
<!-- workflow-docs:start:overview -->
## Overview

One-paragraph description of the project, derived from the PRD and MVP.
<!-- workflow-docs:end:overview -->
```

Rules:

- Markers are always on their **own line**, with no leading whitespace.
- The section id (here `overview`) is kebab-case and stable across
  releases of this skill.
- The `start` marker precedes the section heading (not inside it), so
  the heading itself is regeneratable — this is what lets the skill
  cleanly **omit** a section on re-run if its source disappears.
- Content outside markers is never touched on re-run — `bin/docs-render`
  guarantees this byte-for-byte.
- If the user has moved markers or deleted the closing `end` marker,
  `bin/docs-render` exits 1 (malformed / no markers); the skill then
  flags the file as "manual edits — cannot safely update" and asks the
  user to either restore the markers or delete the generated file and
  regenerate.

## Section omission logic

Each generated section has a deterministic rule for when it appears:

### `README.md`

| Section         | Appears when                                                |
|-----------------|-------------------------------------------------------------|
| `tagline`       | `design/mvp.md` has a one-line description, **or** PRD has a tagline |
| `overview`      | `design/prd.md` / `prd-normalized.md` or `design/mvp.md` exists |
| `who-for`       | PRD / MVP identifies a primary user                         |
| `status`        | `CLAUDE.md` has a non-placeholder `{{CURRENT_PHASE}}` or milestone |
| `scope`         | `design/mvp.md` has an "In scope" and/or "Out of scope" section |
| `how-to-run`    | `CLAUDE.md` has non-placeholder commands                    |
| `key-decisions` | `design/adr/` contains at least one ADR                     |
| `roadmap`       | `design/build-out-plan.md` has 2+ `### Phase N` blocks (per ADR-032). Single-phase / no-phase plans omit this section. |
| `more`          | Always — static pointers to `CLAUDE.md`, `design/`          |

### `design/architecture.md`

| Section | Appears when |
|---|---|
| `overview` | PRD, MVP, `CLAUDE.md`, planning docs, or ADRs describe the system shape |
| `system-boundaries` | PRD/MVP/planning docs define in-scope, out-of-scope, users, integrations, or deployment boundaries |
| `major-components` | `CLAUDE.md`, planning docs, or repo structure identify components/modules/services |
| `data-and-control-flow` | Sources describe data flow, request flow, jobs, agents, pipelines, or state transitions |
| `external-dependencies` | Sources mention APIs, vendors, databases, hosted services, models, or infrastructure |
| `key-constraints` | MVP principles, PRD constraints, ADR consequences, or `CLAUDE.md` rules apply to architecture |
| `extension-points` | Sources identify plugin/module boundaries or safe places to add features |
| `current-decisions` | ADRs or `design/decisions.md` contain current load-bearing decisions |
| `open-questions` | Planning docs, state, milestones, or PR follow-ups list unresolved architecture questions |

### `design/ai-summary.md`

| Section          | Appears when                                              |
|------------------|-----------------------------------------------------------|
| `objectives`     | `design/mvp.md` has a product goal or success criteria    |
| `architecture`   | `CLAUDE.md` has a "Project structure" or tech stack block |
| `tech-stack`     | `CLAUDE.md` has any of runtime / framework / data layer   |
| `constraints`    | `design/mvp.md` has principles, or PRD has constraints    |
| `extension-points` | `CLAUDE.md` or PRD notes extension points (rare on first run — usually omitted) |
| `current-status` | `CLAUDE.md` has a current phase, or ADRs are dated        |
| `key-decisions`  | `design/adr/` contains at least one ADR                   |

A section is **omitted** simply by leaving it out of the `{id, body}`
array passed to `bin/docs-render`; the splicer removes its heading,
body, and both markers and collapses the gap so adjacent omissions
leave no stray blank lines. The skill decides *which* sections to omit
(the rules above); `docs-render` performs the removal.

### Roadmap section content

When the `roadmap` section appears, render a compact table sourced
from `design/build-out-plan.md`'s `### Phase N: <name>` blocks:

| # | Phase | Goal | Exit criterion |
|---|-------|------|----------------|
| 1 | Foundation | one-line goal | observable exit |
| 2 | Core feature | one-line goal | observable exit |
| 3 | Polish | one-line goal | observable exit |

Single-phase projects omit the roadmap section entirely — phase
metadata is implicit and adds no information to the README.

## Execution protocol

Run these steps in order. Stop on the first failure unless noted.

1. **Detect target project.** The current working directory must be the
   target project root. Confirm by checking for a `.git/` directory.
   If absent, ask the user to `cd` into the project root and stop.
2. **Collect inputs.** Read these files if they exist:
   - `design/prd.md` (fall back to `design/prd-normalized.md`)
   - `design/mvp.md`
   - `design/build-out-plan.md`
   - `design/planning.md`
   - `design/decisions.md`
   - Every file matching `design/adr/adr-*.md`
   - Every file matching `design/milestones/*.md`
   - `design/state.md`
   - `CLAUDE.md`

   If none exist, stop with the message "No PRD, MVP, ADRs, or
   CLAUDE.md found — run `prd-normalizer` / `prd-to-mvp` first."
3. **Parse inputs into a single context dict.** Extract the fields
   listed under "Template variables" below. Any field whose source is
   missing is left as `None`. Do **not** substitute placeholder text.
4. **Determine which sections to include.** Apply the omission rules
   above to all outputs. Produce an ordered list of `(section_id,
   rendered_content)` tuples for each file.
5. **Render each section body.** Read `templates/readme-template.md`,
   `templates/architecture-template.md`, and
   `templates/ai-summary-template.md` from the kit's
   `.claude/skills/workflow-docs/` copy (or the templates directory
   when running from the kit repo). For each section the omission rules
   **include**, fill its `{{PLACEHOLDERS}}` from the context dict to
   produce that section's body (heading included). Collect the included
   sections, in template order, as a JSON array of `{id, body}` objects
   per output file.
6. **Splice with `bin/docs-render`.** Hand each file's `{id, body}`
   array to the deterministic splicer rather than editing markers in
   the prompt:
   ```
   bin/docs-render --file <output> --sections <sections.json> --dry-run --format json
   ```
   It owns the marker mechanics: on a first run it wraps each section in
   its `<!-- workflow-docs:start:ID -->` / `:end` markers; on a re-run
   it replaces only the body between existing markers, omits the block
   (markers included) for any section the new run dropped, inserts new
   sections, and collapses blank-line runs — never touching content
   outside the markers. `--dry-run` returns the rendered result for the
   approval gate without writing. If `docs-render` exits 1 because the
   file exists with **no** markers (a hand-written file) or with
   **malformed** markers, do not overwrite: show the diff and ask the
   user to consent or restore the markers. `--check-markers` reports a
   file's marker health on its own.
7. **Show all generated file contents in chat** as fenced markdown blocks
   — or show a diff if the file already exists. Ask explicitly:
   "Write these files? (yes / edit / cancel)". Default: no.
8. **Write only after confirmation.** Re-run `bin/docs-render` for each
   file **without** `--dry-run` to write it atomically. Report the
   absolute paths and a one-line summary (from `outputs.sectionsReplaced`
   / `sectionsAdded` / `sectionsOmitted`) of what was written,
   regenerated, or omitted per file.

## Template variables

Filled into `{{PLACEHOLDER}}` slots from the sources below. Every
field is optional; missing fields drive section omission (see above).

### README (`templates/readme-template.md`)

| Placeholder            | Source                                                    |
|------------------------|-----------------------------------------------------------|
| `{{PROJECT_NAME}}`     | `design/mvp.md` "Product name", else `CLAUDE.md` first `#` heading, else repo basename |
| `{{PROJECT_TAGLINE}}`  | `design/mvp.md` "One-line description", else PRD tagline  |
| `{{PROJECT_DESCRIPTION}}` | `design/mvp.md` "Product goal", else PRD "Goal" / "Problem" paragraph |
| `{{PRIMARY_USER}}`     | `design/mvp.md` "Primary user", else PRD "Target user"    |
| `{{CURRENT_PHASE}}`    | `CLAUDE.md` "Current phase" section                       |
| `{{CURRENT_MILESTONE}}`| `CLAUDE.md` active milestone                              |
| `{{IN_SCOPE_BULLETS}}`     | `design/mvp.md` "In scope" bullets                    |
| `{{OUT_OF_SCOPE_BULLETS}}` | `design/mvp.md` "Out of scope" bullets                |
| `{{INSTALL_COMMAND}}`  | `CLAUDE.md` install command                               |
| `{{DEV_COMMAND}}`      | `CLAUDE.md` dev command                                   |
| `{{TEST_COMMAND}}`     | `CLAUDE.md` test command                                  |
| `{{KEY_ADR_BULLETS}}`  | One bullet per `design/adr/adr-NNN-*.md`, newest first: `- ADR-NNN: <one-line decision>` |

### Architecture (`templates/architecture-template.md`)

| Placeholder | Source |
|---|---|
| `{{PROJECT_NAME}}` | Same as README |
| `{{YYYY-MM-DD}}` | Today's date |
| `{{ARCHITECTURE_OVERVIEW}}` | PRD/MVP goal plus `CLAUDE.md` project structure and tech stack |
| `{{SYSTEM_BOUNDARIES}}` | MVP in/out-of-scope, PRD users/integrations, deploy target |
| `{{MAJOR_COMPONENTS}}` | `CLAUDE.md` repo structure, planning docs, implemented module/service names |
| `{{DATA_AND_CONTROL_FLOW}}` | Planning docs, PRD workflows, ADR consequences, relevant milestone summaries |
| `{{EXTERNAL_DEPENDENCIES}}` | Tech stack, APIs, databases, vendors, hosted services, infra |
| `{{KEY_CONSTRAINTS}}` | MVP principles, PRD constraints, ADR consequences, `CLAUDE.md` rules |
| `{{EXTENSION_POINTS}}` | PRD/CLAUDE/planning notes about module, plugin, API, or workflow boundaries |
| `{{CURRENT_DECISION_BULLETS}}` | Concise current decisions from ADRs and `design/decisions.md`; not every historical detail |
| `{{OPEN_ARCHITECTURE_QUESTIONS}}` | Open questions from planning docs, `design/state.md`, milestone summaries, or PR follow-ups |

### AI summary (`templates/ai-summary-template.md`)

| Placeholder            | Source                                                    |
|------------------------|-----------------------------------------------------------|
| `{{PROJECT_NAME}}`     | Same as README                                            |
| `{{YYYY-MM-DD}}`       | Today's date                                              |
| "Objectives" bullets   | `design/mvp.md` "Product goal" + "Success criteria"       |
| "Architecture" paragraph | `CLAUDE.md` "Project structure" + "Technology stack"    |
| "Tech stack" bullets   | `CLAUDE.md` runtime / framework / data layer / libraries / deploy target |
| "Constraints" bullets  | `design/mvp.md` "Product principles", PRD "Constraints"   |
| "Extension points" bullets | PRD / `CLAUDE.md` extension-point notes (usually omitted on first run) |
| "Current status" bullets | `CLAUDE.md` "Current phase", ADR dates, issue activity if available |
| "Key decisions" bullets | One bullet per ADR, newest first                         |

For the AI summary, also wrap each section of
`templates/ai-summary-template.md` in the same marker format before
rendering. Section ids mirror the template headings in kebab-case.

## Edge cases

- **No PRD, no MVP, no ADRs, no CLAUDE.md** → stop; tell the user to
  run `prd-normalizer` / `prd-to-mvp` first. Do not write empty docs.
- **Only `CLAUDE.md` exists** → generate a minimal README (tagline +
  how-to-run + more), a minimal architecture doc (system shape from
  project structure / tech stack), and a minimal AI summary (tech stack +
  current status). Omit all other sections.
- **Only ADRs exist** → generate only the `key-decisions` section in
  all outputs. Ask the user whether that is useful before writing.
- **Existing README has no markers** → treat as first run; show diff;
  require explicit confirmation; do not silently overwrite.
- **Existing README has markers but they are malformed** (missing
  `end`, duplicated `start`, moved out of order) → stop and ask the
  user to restore or delete the file.
- **Existing file has manual content inside a marker block** → that
  content is **regenerated** (markers are the contract for "this is
  regenerated"). Manual edits belong **outside** markers. Warn the user
  if the diff shows any non-trivial change inside a marker block.
- **ADR files missing a `# ADR-NNN:` heading** → skip that ADR with a
  warning; do not stop.
- **`CLAUDE.md` still has `{{PLACEHOLDER}}` tokens** → treat those
  fields as missing; drive section omission accordingly.
- **Multiple PRD files (`prd.md` + `prd-normalized.md`)** → prefer
  `prd-normalized.md` (the kit's canonical form).

## Review-before-write checkpoint

The user always sees all generated files (as full content on first
run, or as a diff against the existing file on re-run) before any
write. This is the approval gate for ADR-018.

Never skip this step, even if the content looks clean. Never write one
file while deferring the other — all go out together after one
confirmation, or neither.

## Self-check before writing

For each output file:

- [ ] Every `{{PLACEHOLDER}}` is either filled or its enclosing section
      has been omitted.
- [ ] Every included section is wrapped in matching `start:/end:`
      markers with the same section id.
- [ ] No omitted section has left a stray heading, body, or marker
      behind.
- [ ] On re-run: content outside markers in the existing file is
      byte-identical to what it was before.
- [ ] The user explicitly confirmed the write.

If any fail, fix and re-show before writing.

## Running against the kit repo itself

This repo (`workflow-generator`) has its own public architecture document at
`docs/architecture.md`. Do not use this skill to generate the kit repo's own
`docs/architecture.md`; maintain it directly as part of kit-release docs work.

This repo (`workflow-generator`) has `CLAUDE.md` and `design/adr/*.md`
but no `design/prd.md` or `design/mvp.md`. Running the skill here would
produce:

- **README.md**: the `scope`, `overview`, `who-for`, `tagline`, and
  `how-to-run` sections would be **omitted** (no PRD, no MVP, and
  `CLAUDE.md` does not carry commands in this repo). The `status`
  section may also be omitted if `CLAUDE.md` has no current-phase
  line. Only `key-decisions` (from ADRs) and `more` (static pointers)
  would be included — and because the kit repo already has a rich
  hand-written README, the skill would detect the missing markers and
  refuse to overwrite without an explicit "yes, replace".
- **design/architecture.md**: would be sparse and source-heavy because the kit
  repo does not follow the target-project PRD/MVP shape. Maintain
  `docs/architecture.md` directly instead.
- **design/ai-summary.md**: only `key-decisions` (from ADRs) and a
  minimal `current-status` section would be included. Objectives,
  architecture, tech stack, constraints, and extension points would
  all be omitted.

In practice this means the kit repo should **not** run `/workflow-docs`
on itself. The skill is aimed at target projects that have already run
`prd-to-mvp`.

## Handoff

Once the files are written, the user can edit outside markers
freely. Re-running the skill will only touch marked regions. If a
future run needs to add a new section, this skill's next release will
add a new marker id — unrecognised existing markers are left
untouched.

See [`example.md`](example.md) for a worked run on a small sample
project.
