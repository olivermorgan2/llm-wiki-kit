<!--
  Template: README for a kit-installed target project
  Filled by:  the workflow-docs skill, or a human at project bootstrap
  Output in a target project: README.md at the repo root
  Source:     templates/readme-template.md in the Claude Code Workflow Kit

  The README is the public entry point for humans. It explains what the
  project is, who it is for, how to run it, and where to find the deeper
  design docs. Keep it short — link out to design/ and docs/ for detail.

  ------------------------------------------------------------------------
  PLACEHOLDERS (all use {{UPPER_SNAKE}} syntax)
  ------------------------------------------------------------------------

  Required:
    {{PROJECT_NAME}}          Human-readable project name.
    {{PROJECT_TAGLINE}}       One-line elevator pitch.
    {{PROJECT_DESCRIPTION}}   One paragraph describing what the project
                              does and for whom (from PRD / MVP goal).
    {{PRIMARY_USER}}          Who this is for first.
    {{INSTALL_COMMAND}}       Example: "pnpm install"
    {{DEV_COMMAND}}           Example: "pnpm dev"
    {{TEST_COMMAND}}          Example: "pnpm test"

  Optional (sections using only-optional placeholders are omitted
  entirely by the workflow-docs skill if no source data is available):
    {{CURRENT_PHASE}}         Short line describing what is active now.
    {{CURRENT_MILESTONE}}     Example: "MVP", "v0.2"
    {{IN_SCOPE_BULLETS}}      Bulleted "In scope" list from design/mvp.md.
    {{OUT_OF_SCOPE_BULLETS}}  Bulleted "Out of scope" list from design/mvp.md.
    {{KEY_ADR_BULLETS}}       One bullet per ADR, newest first.
    design/architecture.md     Generated current architecture/design reference.

  The workflow-docs skill wraps each generated section in marker fences
  of the form:
      <!-- workflow-docs:start:<section> -->
      ...content...
      <!-- workflow-docs:end:<section> -->
  On re-run, content inside fences is regenerated; content outside is
  preserved. Sections with no source data are omitted entirely (heading
  and body, including the fence).
-->

# {{PROJECT_NAME}}

<!-- workflow-docs:start:tagline -->
> {{PROJECT_TAGLINE}}
<!-- workflow-docs:end:tagline -->

<!-- workflow-docs:start:overview -->
## Overview

{{PROJECT_DESCRIPTION}}
<!-- workflow-docs:end:overview -->

<!-- workflow-docs:start:who-for -->
## Who this is for

{{PRIMARY_USER}}
<!-- workflow-docs:end:who-for -->

<!-- workflow-docs:start:status -->
## Status

Current phase: {{CURRENT_PHASE}}.
Active milestone: `{{CURRENT_MILESTONE}}`.
<!-- workflow-docs:end:status -->

<!-- workflow-docs:start:scope -->
## What is in scope

{{IN_SCOPE_BULLETS}}

### Out of scope

{{OUT_OF_SCOPE_BULLETS}}
<!-- workflow-docs:end:scope -->

<!-- workflow-docs:start:how-to-run -->
## How to run

```bash
{{INSTALL_COMMAND}}
{{DEV_COMMAND}}
{{TEST_COMMAND}}
```
<!-- workflow-docs:end:how-to-run -->

<!-- workflow-docs:start:key-decisions -->
## Key decisions

{{KEY_ADR_BULLETS}}

See [`design/adr/`](design/adr/) for the full list.
<!-- workflow-docs:end:key-decisions -->

<!-- workflow-docs:start:more -->
## More

- [`CLAUDE.md`](CLAUDE.md) — project rules for Claude Code
- [`design/mvp.md`](design/mvp.md) — MVP statement
- [`design/architecture.md`](design/architecture.md) — current architecture/design reference
- [`design/ai-summary.md`](design/ai-summary.md) — AI-readable project summary
- [`design/adr/`](design/adr/) — decision history and rationale
<!-- workflow-docs:end:more -->
