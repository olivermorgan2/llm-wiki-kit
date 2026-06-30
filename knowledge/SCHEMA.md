# knowledge/ — schema & conventions

Canonical definition of the `llm-wiki-kit` **knowledge layer**: what each
file is, how entries are shaped, and what belongs where. This is the
authoritative spec for the layer; [`index.md`](index.md) is the live
navigational map, and [`README.md`](README.md) is the one-paragraph
orientation.

The knowledge layer is a **lightweight, human-curated** addition. It is
**not** shipped by the Claude Code Workflow Kit; it is our
collaboration-protocol layer on top of the kit's tool-maintained `design/`
artifacts.

## File inventory

| File | Role | Format |
|---|---|---|
| `SCHEMA.md` | This file — the layer's conventions and file roles. | Prose. |
| `index.md` | Live front door: every file + key `design/` pointers + current phase. | Tables + status line. |
| `README.md` | One-paragraph orientation that defers to `SCHEMA.md` / `index.md`. | Prose. |
| `project-brief.md` | Distilled product definition: what/problem/vision/users/jobs/spine. | Prose sections. |
| `risks.md` | Live risk register. | `risk → mitigation → status` tables. |
| `open-questions.md` | Unresolved product/technical decisions. | `# / question / owner / status` table. |
| `log.md` | Chronological decision & review log. | `date — decision — rationale — links` entries. |
| `reviews/` | Verbatim archival review artifacts (e.g. adversarial PRD review). | One dated file per review, kept as-produced. |
| `sources/` | Verbatim archival source material (e.g. `prd-original.md`). | Unmodified copies. |

## Entry conventions

- **`log.md`** — newest-relevant entries appended under a dated `###`
  heading. An item that becomes genuinely architectural is promoted to an
  ADR in `design/adr/` via `/adr-writer`, with a back-reference left here.
- **`risks.md`** — every risk carries a mitigation and a status:
  `open` (mitigation planned, not built), `mitigated-by-design`
  (addressed in PRD/architecture), `mitigated` (process step done), or
  `closed`. A bounding addendum does **not** flip a risk to closed if the
  underlying engineering work is unbuilt — it stays `open`.
- **`open-questions.md`** — each question has an owner (default **Oliver**)
  and a status: `open`, `assumption-locked for MVP` (a working assumption
  recorded in a PRD addendum that an owner can still override), `scoped to
  Phase N`, or `closed`. Resolutions are recorded in `log.md` and, if
  architectural, an ADR — then the question is marked `closed`.
- **`reviews/` & `sources/`** — verbatim. Do not edit archived artifacts to
  fix link drift; reconcile in the live files and let the archive stand as
  the historical record.

## What belongs here vs not

| Belongs in `knowledge/` | Belongs elsewhere |
|---|---|
| Distilled understanding that orients a human/agent fast: brief, live risks, open questions, decision log. | **Formal workflow artifacts** (PRD, normalized PRD, MVP, build-out plan, ADRs, architecture) → `design/` — the kit owns and regenerates these. |
| Curated summaries and pointers into `design/`. | **Freeform working notes / evaluation logs** → `notes/`. |
| Verbatim archival sources/reviews under `sources/` and `reviews/`. | **Per-issue prompts** → `prompts/`. |

## Relationship to `design/`

`design/` is authoritative and tool-maintained; `knowledge/` is the curated
human view that **links into** `design/` rather than duplicating it. When a
knowledge-layer item becomes a real architectural decision, promote it to an
ADR (`/adr-writer`) and leave a back-reference in `log.md`.

> Note: the kit also maintains its own `design/decisions.md` (a different,
> kit-owned file written by skills like `/clarify`). The knowledge layer's
> chronological log is `knowledge/log.md` — do not conflate the two.

## Update cadence

Curate, don't accrete. Update these files when a decision lands, a risk
changes status, or a question opens/closes. Keep each entry short and
current — stale curated knowledge is worse than none.
