<!--
  Template: AI-readable project summary
  Filled by: the workflow-docs skill, or a human at design checkpoints
  Output in a target project: design/ai-summary.md
  This document is written for AI consumption. External AIs (Perplexity,
  ChatGPT, other Claude instances) read it when helping to design a
  feature or review an ADR. Prioritise structure, concrete facts, and
  unambiguous language over readability for humans.
  Pair with design/architecture.md: the architecture doc is the current
  human-readable shape; this summary is the compact AI handoff.
  Regenerate both after each significant architectural change.
-->

# {{PROJECT_NAME}} — AI Summary

**Last updated:** {{YYYY-MM-DD}}

## Objectives

- {{Primary objective of the project in one line.}}
- {{Secondary objectives, one per bullet.}}

## Architecture

{{High-level shape: the main components, how they communicate, where
data lives. One paragraph or a short bullet list — no marketing prose.}}

## Tech stack

- Runtime / language: {{...}}
- Framework: {{...}}
- Data layer: {{...}}
- External services / APIs: {{...}}
- Deployment target: {{...}}

## Constraints

- {{Hard constraints — things that must be true (e.g. offline-capable,
  no third-party trackers, sub-200ms p95).}}
- {{Soft constraints — strong preferences (e.g. TypeScript preferred,
  keep bundle under 200KB).}}
- {{Non-functional constraints (performance, cost, compliance).}}

## Extension points

- {{Where new functionality typically plugs in — modules, hooks, plugin
  boundaries, config surfaces.}}
- {{Deliberate "do not extend here" boundaries, if any.}}

## Current status

- Current phase: {{e.g. MVP, Beta, v1.0}}
- In progress: {{short list of active workstreams}}
- Recently landed: {{2–3 notable recent changes}}
- Known gaps / risks: {{things an AI reviewer should flag in design suggestions}}

## Key decisions

<!-- Ordered newest first. One line per ADR. -->
- ADR-{{NNN}}: {{one-line decision summary}}
- ADR-{{NNN}}: {{one-line decision summary}}
