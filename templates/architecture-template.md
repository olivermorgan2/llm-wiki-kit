<!--
  Template: Current architecture/design reference
  Filled by: the workflow-docs skill, or a human at architecture checkpoints
  Output in a target project: design/architecture.md
  Source: templates/architecture-template.md in the Claude Code Workflow Kit

  This document describes the current shape of the project. It is not an ADR
  log and not a raw implementation inventory. Keep it current after meaningful
  architecture changes, accepted ADRs, major feature work, and before releases.

  The workflow-docs skill wraps each generated section in marker fences of the
  form:
      <!-- workflow-docs:start:<section> -->
      ...content...
      <!-- workflow-docs:end:<section> -->
  On re-run, content inside fences is regenerated; content outside is
  preserved. Sections with no source data are omitted entirely.
-->

# {{PROJECT_NAME}} — Architecture

**Last updated:** {{YYYY-MM-DD}}

<!-- workflow-docs:start:overview -->
## Overview

{{ARCHITECTURE_OVERVIEW}}
<!-- workflow-docs:end:overview -->

<!-- workflow-docs:start:system-boundaries -->
## System boundaries

{{SYSTEM_BOUNDARIES}}
<!-- workflow-docs:end:system-boundaries -->

<!-- workflow-docs:start:major-components -->
## Major components

{{MAJOR_COMPONENTS}}
<!-- workflow-docs:end:major-components -->

<!-- workflow-docs:start:data-and-control-flow -->
## Data and control flow

{{DATA_AND_CONTROL_FLOW}}
<!-- workflow-docs:end:data-and-control-flow -->

<!-- workflow-docs:start:external-dependencies -->
## External dependencies

{{EXTERNAL_DEPENDENCIES}}
<!-- workflow-docs:end:external-dependencies -->

<!-- workflow-docs:start:key-constraints -->
## Key constraints

{{KEY_CONSTRAINTS}}
<!-- workflow-docs:end:key-constraints -->

<!-- workflow-docs:start:extension-points -->
## Extension points

{{EXTENSION_POINTS}}
<!-- workflow-docs:end:extension-points -->

<!-- workflow-docs:start:current-decisions -->
## Current decisions

{{CURRENT_DECISION_BULLETS}}

Decision history lives in [`design/adr/`](adr/) when ADRs are used. This file
summarizes the current architecture; it does not replace ADRs as rationale.
<!-- workflow-docs:end:current-decisions -->

<!-- workflow-docs:start:open-questions -->
## Open architecture questions

{{OPEN_ARCHITECTURE_QUESTIONS}}
<!-- workflow-docs:end:open-questions -->
