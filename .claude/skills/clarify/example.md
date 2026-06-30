# clarify — worked example

A short walk-through of the clarify skill on the same fictional
team-collaboration tool used in `skills/planning/example.md`. Inputs
are abbreviated for readability — a real PRD, MVP, and planning.md
are longer.

---

## 1. Inputs

### `design/prd-normalized.md` (excerpt)

A team-collaboration tool for distributed product squads. Realtime
doc editing, threaded comments, shared task lists, integrations
(Slack, GitHub, Linear). Constraints: web + desktop, multi-tenant,
SSO required.

### `design/mvp.md` (excerpt)

In scope: realtime doc editing, threaded comments, shared task
lists, Slack integration only.

### `design/planning.md` (excerpt)

Three phases (Foundation → Editor → Slack). Open research questions:

- **Q1** — Does Yjs's relay model handle our auth pattern?
- **Q2** — What's the right cursor-presence cadence — 100ms or
  500ms?

### `design/adr/` (excerpt)

ADR-001 ("CRDT vs OT for realtime sync") — accepted. Locks the
realtime-sync question.

## 2. Re-run detection

`design/decisions.md` does not exist yet. This is a first run; no
earlier entries to preserve.

## 3. Accepted-ADR scan

Scanned `design/adr/`:

- ADR-001 (accepted, "CRDT vs OT for realtime sync") — topic
  "realtime sync mechanism" is locked. Any gray area mapping to
  this topic will be skipped.

## 4. Codebase scout

The repo has no `src/` yet (Phase 1 hasn't started). Skipping
codebase scout for this run.

## 5. Gray-area surface

Walked through PRD, MVP, planning.md. Surfaced 5 candidate gray
areas:

1. Cursor-presence cadence — 100ms or 500ms? (matches planning.md
   Q2)
2. SSO providers — Okta only, or Okta + Google?
3. Comments — threaded depth limit (none / 1 / 2)?
4. Task list assignee — required, optional, or absent in v1?
5. Slack integration direction — one-way (app → Slack) or
   bidirectional?

Filtered out:

- "Realtime sync mechanism" — locked by accepted ADR-001.
- "CRDT library choice (Yjs vs Automerge)" — would graduate to an
  ADR (reversal cost is structural — affects the entire editor
  Phase 2). Surfaced separately as a candidate ADR topic.

## 6. Checklist for the user

```
Gray areas surfaced. Pick which to resolve now (1-5, comma-separated, or "all"):

  1. Cursor-presence cadence (100ms vs 500ms)
  2. SSO providers (Okta only vs Okta + Google)
  3. Comments depth limit (none / 1 / 2)
  4. Task list assignee model (required / optional / absent)
  5. Slack integration direction (one-way / bidirectional)

Candidate ADR topics (will not be recorded here):
  - CRDT library choice (Yjs vs Automerge) — reversal cost is structural
```

User picks: `2, 3, 4`.

## 7. Deep-dive resolution

### Gray area 2 — SSO providers

- **Question:** Okta only for v1, or Okta + Google?
- **Options weighed:**
  - Okta only — simpler integration, faster Phase 1, leaves Google
    for v1.1.
  - Okta + Google — broader design-partner reach at Phase 1 launch,
    one extra integration.
- **Decision:** Okta + Google.
- **Rationale:** Three of five design partners use Google
  Workspace; shipping with Okta only would block them.

### Gray area 3 — Comments depth limit

- **Question:** Threaded comments — depth limit none / 1 / 2?
- **Options weighed:**
  - None — full threading, recursive UI, more complex.
  - 1 — top-level + replies, no nested replies.
  - 2 — replies + replies-to-replies.
- **Decision:** Depth 1.
- **Rationale:** Slack's threading is depth 1 and is the model
  users expect; deeper threads add UI complexity without clear
  benefit at MVP scope.

### Gray area 4 — Task list assignee

- **Question:** Inline task list — assignee required, optional, or
  absent in v1?
- **Options weighed:**
  - Required — every task has an owner.
  - Optional — assignee field exists but can be unset.
  - Absent — v1 ships without assignees, added in v1.1.
- **Decision:** Optional.
- **Rationale:** Required friction-blocks fast capture (a writer
  jotting a TODO mid-doc); absent forces a future migration of all
  v1 tasks. Optional matches both flows at low cost.

## 8. Output — `design/decisions.md`

```markdown
# Squadly — Decisions

**Last updated:** 2026-04-30
**Source PRD:** design/prd-normalized.md
**Source MVP:** design/mvp.md

## Purpose

Captures informal-but-settled context for the v1 build. Decisions
here are below ADR weight by design — reversal would be a routine
code change, not an architectural shift.

## Decisions

<!-- decisions:01:start -->

### D1 — SSO providers for v1

- **Question:** Okta only for v1, or Okta + Google?
- **Options weighed:**
  - Okta only — simpler integration, leaves Google for v1.1.
  - Okta + Google — broader design-partner reach.
- **Decision:** Okta + Google.
- **Rationale:** Three of five design partners use Google
  Workspace; shipping with Okta only would block them.

<!-- decisions:01:end -->

<!-- decisions:02:start -->

### D2 — Threaded comments depth limit

- **Question:** Threaded comments — depth limit none / 1 / 2?
- **Options weighed:**
  - None — full threading, recursive UI.
  - 1 — top-level + replies.
  - 2 — replies + replies-to-replies.
- **Decision:** Depth 1.
- **Rationale:** Matches Slack's mental model; deeper threads add
  UI complexity without clear MVP benefit.

<!-- decisions:02:end -->

<!-- decisions:03:start -->

### D3 — Task list assignee model

- **Question:** Inline task list — assignee required, optional, or
  absent in v1?
- **Options weighed:**
  - Required — every task has an owner.
  - Optional — assignee field, can be unset.
  - Absent — v1 ships without, added in v1.1.
- **Decision:** Optional.
- **Rationale:** Required friction-blocks fast capture; absent
  forces a future migration. Optional matches both flows.

<!-- decisions:03:end -->
```

## 9. Self-check

- ✓ Every entry has question, options weighed, decision, rationale.
- ✓ No entry duplicates a topic already in decisions.md (first run)
  or an accepted ADR (ADR-001's topic was filtered out at step 5).
- ✓ No entry would, on its own, warrant an ADR (each is reversible
  via routine code change).
- ✓ Marker fences balanced — three `:start` / three `:end` pairs.

## 10. Report

```
clarify: wrote design/decisions.md (3 new entries)
clarify: 2 gray areas left unresolved for next run (cursor-presence
         cadence, Slack integration direction)
clarify: 1 candidate ADR topic surfaced — CRDT library choice (Yjs
         vs Automerge). Hand to adr-writer when ready.
```

## 11. Re-run behaviour

If the user runs `/clarify` again later:

- D1, D2, D3 are preserved verbatim.
- The two unresolved gray areas (cursor-presence cadence, Slack
  direction) surface again in the checklist.
- Any new gray area surfaced by changes to PRD/MVP/planning.md
  shows up alongside.
- The CRDT-library candidate-ADR topic is reported again unless the
  user has accepted an ADR for it in the meantime.

The kit's append-only contract means every previous decision stays
visible; the log grows but never thrashes.
