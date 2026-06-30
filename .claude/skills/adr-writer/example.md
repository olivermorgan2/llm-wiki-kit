# adr-writer — worked example

A single batch run that drafts two ADRs for the Pace Drift project:
one **genuinely contentious** decision (real trade-offs to weigh) and
one **near-trivial** decision (the obvious answer). Both get an ADR
because the kit's convention is "decisions are recorded, not assumed."

Input is the "Decisions needing ADRs" list from
[`prd-to-mvp/example.md`](../prd-to-mvp/example.md).

---

## 1. Input — decision batch

From the Pace Drift build-out plan:

1. **GPX parsing location: browser-side vs. server-side.** Affects
   privacy, file-size limits, cost, and the static-site assumption.
2. **Test framework.** Vitest, Jest, or node:test. Affects developer
   ergonomics; mostly a defaults question.

Plus context: `design/prd-normalized.md` (constraints: web-only,
laptop-first, single-user, local session) and `design/mvp.md`
(principles: local session only, no server-side persistence).

## 2. Confirmation turn

The skill shows the numbered list back and asks one question:

> Two ADRs to draft: (1) GPX parsing location, (2) test framework.
> Anything to remove, reframe, or add before I draft? For the parsing
> question, the constraint "no server-side persistence" leans
> toward browser-side, but I want to make sure the trade-offs are
> spelled out — confirm?

User: confirms, and asks the parsing-location ADR to also note the
offline-use angle.

## 3. Output 1 (contentious) — `design/adr/adr-001-gpx-parsing-location.md`

```markdown
# ADR-001: Parse GPX files in the browser, not on a server

**Status:** proposed
**Date:** 2026-04-15

## Context

Pace Drift loads a GPX file from the user, parses it, and computes
pace drift against a target pace. The parsing and computation could
run in the user's browser or on a server.

The MVP constraints from design/prd-normalized.md and design/mvp.md
are relevant:

- Single-user, local session.
- No server-side persistence.
- Laptop-first; users open a recent race file from their machine.
- Verbatim data — no AI smoothing or interpretation.

A GPX from a marathon is typically a few MB; from an ultra, larger
but still well within browser memory.

## Options considered

### Option A: Parse and compute in the browser

- Pros: no backend to run or pay for; the GPX never leaves the
  user's machine (privacy by default); trivially works offline once
  the page is loaded; matches the "no server-side persistence"
  principle without further effort; deployable as a static site.
- Cons: large or unusual GPX files load entirely into browser memory;
  parsing-library choice is constrained to ones that work in the
  browser; debugging happens in browser devtools rather than server
  logs.

### Option B: Parse and compute on a server

- Pros: heavier file-size ceiling; richer choice of parsing
  libraries; centralised logging if anything goes wrong.
- Cons: adds infrastructure to run and pay for; uploads the user's
  race file to a third party (privacy regression); breaks the
  static-site deployment model; offline use requires extra work;
  contradicts the "no server-side persistence" principle even
  if uploads are not stored.

## Decision

Parse and compute in the browser. The MVP's local-session, no-backend,
privacy-leaning constraints all push the same direction, and there is
no current evidence that file size will exceed browser limits for the
race files this product targets. The whole product can ship as a
static site.

## Consequences

- Easier: deployment is "host the static files anywhere"; no backend
  to run; offline use works once the page is loaded; user files stay
  on the user's machine.
- Harder: very large GPX files (multi-day ultras with high-frequency
  recording) may hit browser memory limits — out of scope for the
  MVP but noted.
- Maintain: the parsing library choice must remain browser-compatible.
- Deferred: any feature that genuinely needs a server (saved history,
  multi-race comparison, sharing) is out of MVP scope and would
  re-open this decision.
```

## 4. Output 2 (trivial) — `design/adr/adr-002-test-framework.md`

```markdown
# ADR-002: Use Vitest as the test framework

**Status:** proposed
**Date:** 2026-04-15

## Context

Pace Drift is a small TypeScript browser app. It needs a unit-test
runner for the GPX parser and the drift-computation module (Phase 1
of the build-out plan). The choice has no architectural weight; it is
a defaults question, but the kit's convention is to record it anyway.

## Options considered

### Option A: Vitest

- Pros: works out of the box with TypeScript and ESM; familiar API
  (Jest-compatible); fast; widely used in 2026 browser-app projects.
- Cons: one more dev dependency to keep current.

### Option B: node:test

- Pros: no dependency to install; ships with Node.
- Cons: less ergonomic for browser-flavoured code; smaller community
  for tooling around it.

## Decision

Use Vitest. Default-of-defaults for a TypeScript browser app in 2026,
and the marginal effort to install it pays back immediately in
ergonomics.

## Consequences

- Easier: writing and running tests is a one-line `npx vitest run`.
- Harder: None new.
- Maintain: keep Vitest current with each major release.
- Deferred: None new.
```

## 5. Self-check trace

For ADR-001:

- [x] NNN is `001` (no existing ADRs in the target project).
- [x] Status is `proposed`.
- [x] Two options listed, each with Pros and Cons.
- [x] Decision names Option A.
- [x] Consequences covers all four bullets.
- [x] Date is today's date.
- [x] No `{{...}}` placeholders remain.

For ADR-002:

- [x] NNN is `002` (next after the just-drafted ADR-001).
- [x] Status is `proposed`.
- [x] Two options listed, each with Pros and Cons.
- [x] Decision names Option A.
- [x] Consequences covers all four bullets — "Harder" and "Deferred"
      use "None new" rather than being omitted.
- [x] Date is today's date.
- [x] No `{{...}}` placeholders remain.

All checks pass for both ADRs. The skill reports back:

> Drafted 2 ADRs:
> - design/adr/adr-001-gpx-parsing-location.md (proposed)
> - design/adr/adr-002-test-framework.md (proposed)
>
> Review and change status to `accepted` when ready.

The user reviews and accepts in their own time. Acceptance is a human
act and is not done by this skill.
