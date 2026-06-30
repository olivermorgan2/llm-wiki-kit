# planning — worked example

A short walk-through of the planning skill on a fictional multi-month
project. Inputs are abbreviated for readability — a real PRD and MVP
are longer.

---

## 1. Inputs

### `design/prd-normalized.md` (excerpt)

A team-collaboration tool for distributed product squads. Eleven
canonical fields. Core capabilities: realtime doc editing, threaded
comments, shared task lists, integrations (Slack, GitHub, Linear).
Constraints: web + desktop, multi-tenant, must integrate with
existing SSO. Goal: ship a usable v1 in three months.

### `design/mvp.md` (excerpt)

In scope: realtime doc editing, threaded comments, shared task
lists, Slack integration only. Out of scope: GitHub / Linear
integrations, mobile, offline mode. Three principles: realtime is
non-negotiable; SSO from day one; ship one integration deeply, not
three shallowly.

### `design/build-out-plan.md` (excerpt)

Three phases. Phase 1 — Foundation (auth, SSO, doc storage).
Phase 2 — Editor (realtime sync, comments, task lists).
Phase 3 — Slack integration and polish.

## 2. Elicitation by the skill

Batched, one turn (≤ 5 questions):

1. Realtime sync — confirm CRDT-based (per the open question in
   the PRD) or are we still considering OT?
2. SSO providers — Okta only for v1, or Okta + Google?
3. Comments — threaded depth limit (none / 1 / 2)?
4. Task lists — assignee required, optional, or absent in v1?
5. Slack integration — direction (Slack → app only, or bidirectional)?

User answers: CRDT decided; Okta + Google; depth limit 1; assignee
optional; bidirectional.

## 3. Output — `design/planning.md` (excerpt)

```markdown
# Squadly — Planning

**Last updated:** 2026-04-30
**Source PRD:** design/prd-normalized.md
**Source MVP:** design/mvp.md

## Objective

Decompose the v1 MVP into requirements, surface delivery risks,
record the assumptions the three-month timeline depends on, justify
the foundation → editor → Slack phase order, and list the
architectural decisions that need ADRs before Phase 1 starts.

## Requirements decomposition

<!-- planning:requirements:start -->

### Auth and SSO

- **R1 — Okta SSO end-to-end.** Login redirect, callback,
  session storage, logout. Required for any non-public surface.
- **R2 — Google SSO alongside Okta.** Same flow, second provider.

### Realtime doc editing

- **R3 — CRDT-backed doc storage.** Per-doc Yjs document, server
  is a relay, clients reconcile.
- **R4 — Cursor presence.** Other users' cursors visible in real
  time.

### Comments

- **R5 — Threaded comments, depth-1.** Top-level comment +
  replies. No nested replies under replies.

### Task lists

- **R6 — Inline task list block.** Checkbox + text + optional
  assignee.

### Slack

- **R7 — Bidirectional Slack notifications.** App → Slack on
  comment / mention; Slack → app on reply.

<!-- planning:requirements:end -->

## Risks

<!-- planning:risks:start -->

### Risk 1 — CRDT complexity blows the schedule

- **Impact:** high
- **Likelihood:** medium
- **Mitigation:** Yjs is mature; use it directly, do not roll our
  own. Phase 1 includes a CRDT spike to validate the relay model
  before Phase 2 commits to it.

### Risk 2 — Slack bidirectional flow needs OAuth scopes that
require workspace admin approval

- **Impact:** medium
- **Likelihood:** high
- **Mitigation:** Document the required scopes in the install
  guide; surface scope-failure errors clearly; gracefully degrade
  to one-way Slack → app if bidirectional is blocked.

<!-- planning:risks:end -->

## Assumptions

<!-- planning:assumptions:start -->

- **A1 — Okta tenants are available for v1 design partners.** If
  this is wrong: Google SSO ships first and Okta moves to v1.1.
- **A2 — A single relay process per workspace can handle 50
  concurrent editors.** If this is wrong: shard by document, add
  presence-only relays, raise infra costs.
- **A3 — Slack apps under our namespace can request the
  `chat:write` and `channels:history` scopes without enterprise
  review.** If this is wrong: bidirectional Slack moves to v1.1.

<!-- planning:assumptions:end -->

## Sequencing rationale

<!-- planning:sequencing:start -->

### Phase 1 first because foundation gates everything

Auth, SSO, and doc storage are dependencies for every other
surface. Without them, Phase 2's editor has nowhere to save and
no user identity to attach to documents. Phase 1 also includes
the CRDT spike — if that fails, the whole plan changes.

### Phase 2 follows because the editor is the product

Realtime editing, comments, and task lists are the core user
value. Phase 2 ships them on top of Phase 1's storage and auth.
Comments and task lists piggyback on the same doc-mutation pipe.

### Phase 3 follows because integrations need a working app first

Slack notifications only make sense once there are doc events
worth notifying about. Bidirectional flow needs a stable doc API
for Slack replies to land into. Doing Slack earlier risks
designing the integration around an editor that hasn't shipped.

<!-- planning:sequencing:end -->

## Open research questions

<!-- planning:research:start -->

- **Q1 — Does Yjs's relay model handle our auth pattern, or do
  we need a custom server?** Owner: backend lead. Target:
  2026-05-08.
- **Q2 — What's the right cursor-presence cadence — 100ms or
  500ms?** Owner: realtime engineer. Target: 2026-05-15.

<!-- planning:research:end -->

## Decisions needing ADRs

- Realtime sync: CRDT (Yjs) vs. OT — confirm CRDT formally.
- SSO: Okta + Google in v1, or stage Google to v1.1?
- Slack integration: bidirectional vs. one-way only.
- Task list assignee model: required, optional, or absent.
```

## 4. Handoff

The "Decisions needing ADRs" list (four items) goes to `adr-writer`
as a batch. `issue-planner` reads the sequencing-rationale section
when ordering the issue backlog so Phase 1 issues land before
Phase 2 issues automatically.
