# ADR-005: Mandatory engine filesystem-safety gate

**Status:** proposed
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

Risks addressed (`knowledge/risks.md`): path-traversal/symlink escape, partial
multi-file writes on interruption, untrusted source material as instructions, and
skills bypassing deterministic safety. This layer is foundational to Phase 1 and
is used by every mutation across all phases.

## Options considered

### Option A: A single mandatory engine-level filesystem-safety gate every write routes through — canonicalize + boundary-check + symlink-resolve before write, atomic write-and-rename, all-or-nothing multi-file staging under `.llm-wiki/`

- Pros: one chokepoint makes "zero out-of-boundary writes" (criterion 17)
  auditable and testable with traversal/symlink fixtures; atomic write+rename and
  staged multi-file commit prevent partial/corrupt state on interruption
  (criteria 3, 20 reliability); routing all mutation through the engine stops
  skills bypassing approval (FR10); treating inputs as data closes the
  untrusted-instruction risk.
- Cons: every write path must go through the gate (no shortcut convenience
  writes); atomic-rename + staging semantics differ subtly across the five
  platforms (notably Windows) and need per-platform tests.

### Option B: Per-operation ad-hoc path checks at each call site

- Pros: less upfront plumbing; each command guards only what it touches.
- Cons: a single missed call site breaks the zero-out-of-boundary gate
  (criterion 17) — exactly the failure the hard gate forbids; no central place to
  prove safety; invites drift and bypass; partial-write safety becomes per-command
  and inconsistent.

## Decision

Adopt **Option A** — one mandatory engine-level filesystem-safety gate that all
writes route through, performing canonicalization, boundary enforcement, symlink
resolution, atomic write+rename, and all-or-nothing multi-file staging under
`.llm-wiki/`. Because zero out-of-boundary writes is the project's only hard
quantitative release gate, a single auditable chokepoint is the only design that
makes the gate provable; per-call-site checks (Option B) cannot guarantee it.
Staged mutation `inspect/plan/apply` mechanics that sit on top of this layer are
specified in ADR-006.

## Consequences

- Easier: criterion 17 is provable at one gate with traversal/symlink fixtures;
  interruption-safety (criteria 3, 20) follows from atomic+staged writes;
  approval-bypass and untrusted-input risks are structurally closed.
- Harder: all mutation funnels through the gate; cross-platform atomic-rename and
  staging semantics (notably Windows) need dedicated per-platform tests.
- Maintain: the path-canonicalization/boundary/symlink logic, the atomic-write +
  staging primitives under `.llm-wiki/`, and the platform-specific fixtures that
  hold the zero-out-of-boundary gate.
- Deferred / validation implications: criterion 17 is the hard release gate;
  criteria 3 and 20 (interruption recovery, upgrade/uninstall preservation) lean
  on this layer. Hash-bound stale-plan rejection (criterion 13) and
  inspect/plan/apply ergonomics are deferred to ADR-006. Dirty-worktree conflict
  detection is in scope as a guard, but its UX surfaces via ADR-006.
