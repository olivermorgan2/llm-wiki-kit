# knowledge/ — curated project knowledge layer

This directory is a **lightweight, human-curated** knowledge layer for
`llm-wiki-kit`. It is **not** shipped by the Claude Code Workflow Kit; it
is our collaboration-protocol addition, layered on top of the kit's
`design/` artifacts.

## What belongs here

- Durable, distilled understanding of the project that helps a human or an
  agent get oriented fast: the brief, the live risks, the open questions,
  and the decision log.
- Curated summaries and pointers — **not** transcript dumps, not raw chat
  logs, not generated output.

## What does NOT belong here

- **Formal workflow artifacts** → `design/` (PRD, normalized PRD, ADRs,
  architecture, MVP scope, build-out plan). The kit owns those and its
  skills regenerate them.
- **Freeform working notes / evaluation logs** → `notes/`.
- **Per-issue prompts** → `prompts/`.
- **Archival source material** → `knowledge/sources/` (kept verbatim;
  e.g. the original PRD).

## Relationship to `design/`

`design/` is authoritative and tool-maintained. `knowledge/` is the
curated human view: it links into `design/` rather than duplicating it.
When a `knowledge/` item becomes a real architectural decision, promote it
to an ADR in `design/adr/` via `/adr-writer` and leave a back-reference in
`knowledge/decisions.md`.

## Files

| File | Purpose |
|---|---|
| `README.md` | This file — what the layer is and what belongs where. |
| `project-brief.md` | Distilled product definition, problem, users, jobs. |
| `risks.md` | Live risk register (risk → mitigation → status). |
| `open-questions.md` | Unresolved product/technical decisions, with owner/status. |
| `decisions.md` | Chronological decision & review log. |
| `sources/` | Verbatim archival source material (e.g. `prd-original.md`). |

## Update cadence

Curate, don't accrete. Update these files when a decision lands, a risk
changes status, or a question opens/closes — keep each entry short and
current. Stale curated knowledge is worse than none.
