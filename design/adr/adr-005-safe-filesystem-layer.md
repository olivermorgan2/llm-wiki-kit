# ADR-005: Mandatory engine filesystem-safety gate

**Status:** accepted
**Date:** 2026-06-30

## Context

All deterministic engine operations must restrict writes to the repository and
configured bundle/metadata paths, resolve and validate symlinks, use atomic
writes, avoid partial multi-file updates, treat imported content as data (never
instructions), restrict staging to an engine-managed `.llm-wiki/` location, and
prevent skills from bypassing approval policy via lower-level mutating commands
(PRD §8 FR10; §14 Security/Reliability). Acceptance criterion 17 (pre-write
guards reject tested path traversal and symlink escapes) is governed by the
project's **single hard quantitative release gate: zero out-of-boundary writes**
(`design/mvp.md`; addendum 001).

Risks addressed (`knowledge/risks.md`): path-traversal/symlink escape (fully
here), partial-write-on-interruption at the **single-file** level (fully here;
the *multi-file* transaction that makes a batch all-or-nothing is completed by
ADR-006), untrusted source material as instructions (contributed to here at the
filesystem boundary), and skills bypassing deterministic safety. This layer is
foundational to Phase 1 and is used by every mutation across all phases.

## Options considered

### Option A: A single mandatory engine-level filesystem-safety gate every write routes through — canonicalize + boundary-check + symlink-resolve before write, per-file atomic write-and-rename, engine-managed `.llm-wiki/` staging area

- Pros: one chokepoint makes "zero out-of-boundary writes" (criterion 17)
  auditable and testable with traversal/symlink fixtures; per-file atomic
  write+rename guarantees no single file is ever left half-written, so an
  interruption leaves each file either fully old or fully new (a **necessary**
  primitive for the interruption-recovery reliability behind criteria 3, 20);
  routing all mutation through the engine stops skills bypassing approval
  (FR10); treating imported content as data (never instructions) **contributes
  to** mitigating the untrusted-instruction risk at the filesystem boundary.
- Cons: every write path must go through the gate (no shortcut convenience
  writes); per-file atomic-rename semantics differ subtly across the five
  platforms (notably Windows) and need per-platform tests; per-file atomicity
  alone does **not** make a multi-file operation atomic — cross-file
  all-or-nothing commit is a separate transaction concern (see Decision).

### Option B: Per-operation ad-hoc path checks at each call site

- Pros: less upfront plumbing; each command guards only what it touches.
- Cons: a single missed call site breaks the zero-out-of-boundary gate
  (criterion 17) — exactly the failure the hard gate forbids; no central place to
  prove safety; invites drift and bypass; partial-write safety becomes per-command
  and inconsistent.

## Decision

Adopt **Option A** — one mandatory engine-level filesystem-safety gate that all
writes route through, performing canonicalization, boundary enforcement, symlink
resolution, **per-file atomic write+rename**, and providing the engine-managed
`.llm-wiki/` staging area. Because zero out-of-boundary writes is the project's
only hard quantitative release gate, a single auditable chokepoint is the only
design that makes the gate provable; per-call-site checks (Option B) cannot
guarantee it.

**Scope of the atomicity claim (deliberately narrowed).** ADR-005 guarantees
**per-file atomicity** only: every individual write is a write-to-temp +
fsync + atomic rename within the boundary, so no file is ever observed
half-written. It does **not** define a **cross-file transaction model** — the
staging directory layout, plan manifest, validate-before-commit ordering,
commit sequencing, recovery/rollback, temp-file cleanup, and partial-commit
detection that make a *multi-file* operation all-or-nothing. That transaction
model rides on the staging area this gate provides but is **owned by ADR-006**
(the staged mutation `inspect/plan/apply` layer), alongside hash-bound
stale-plan rejection. This ADR provides the primitives (bounded writes,
per-file atomicity, the staging location); ADR-006 composes them into
interruption-safe multi-file transactions.

## Consequences

- Easier: criterion 17 is provable at one gate with traversal/symlink fixtures;
  per-file atomicity gives criteria 3 and 20 a **necessary** interruption-safety
  primitive (full multi-file interruption safety is completed by ADR-006's
  transaction model, not by this gate alone); approval-bypass is structurally
  closed, and inputs-as-data **contributes to** closing the untrusted-input
  risk at the filesystem boundary.
- Harder: all mutation funnels through the gate; cross-platform per-file
  atomic-rename semantics (notably Windows) need dedicated per-platform tests.
- Maintain: the path-canonicalization/boundary/symlink logic, the per-file
  atomic-write primitive and the `.llm-wiki/` staging location, and the
  platform-specific fixtures that hold the zero-out-of-boundary gate.
- Deferred / validation implications: criterion 17 is the hard release gate
  proved here; criteria 3 and 20 (interruption recovery, upgrade/uninstall
  preservation) **depend on** this layer's per-file atomicity **plus** ADR-006's
  cross-file transaction model — neither is fully satisfied by ADR-005 alone.
  The cross-file transaction model (staging manifest, commit ordering,
  recovery/rollback, partial-commit detection), hash-bound stale-plan rejection
  (criterion 13), and inspect/plan/apply ergonomics are deferred to ADR-006.
  Dirty-worktree conflict detection is in scope as a guard, but its UX surfaces
  via ADR-006.
