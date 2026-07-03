# ADR-006: Staged mutation lifecycle and cross-file transaction model

**Status:** proposed
**Date:** 2026-07-03

## Context

Every managed engine mutation must be previewable, reviewable, and safe to
interrupt: the engine stages changes, a human (or CI) inspects a plan, and only
an unchanged plan is applied (PRD §8 FR1, FR7, FR11; §14 Reliability).
Acceptance criteria **11** (staged preview of changes), **12** (apply only what
was previewed), and **13** (reject a stale plan whose source/target changed
since preview) depend on this lifecycle. Interruption-recovery criteria **3**
and **20** additionally need a *multi-file* operation to be all-or-nothing, not
just per-file atomic.

[ADR-005](adr-005-safe-filesystem-layer.md) (accepted) deliberately narrowed its
atomicity claim to **per-file** write+rename inside the boundary and provided
the engine-managed `.llm-wiki/` staging location, but explicitly left the
**cross-file transaction model** — staging directory layout, plan manifest,
validate-before-commit ordering, commit sequencing, recovery/rollback,
partial-commit detection, and temp cleanup — **to this ADR**. The first real
consumer is Phase 2: installing into a **non-empty** repo and `init`-ing a
bundle are both multi-file mutations that must lose no pre-existing file and
must survive interruption. Phase 2 consumes the **transaction half** (all-or-
nothing commit on the staging area); the `page inspect/plan/apply` **CLI/UX
half** and its hash-bound stale-plan rejection are consumed in Phase 3.

Constraints: writes route only through the ADR-005 gate (the zero-out-of-
boundary hard gate stays provable); the engine is self-contained and offline
(ADR-002) so the transaction may not depend on an external VCS or daemon; plans
are carried in the ADR-003 JSON envelope; and cross-platform atomic-rename
semantics (notably Windows) must hold. Risks addressed
(`knowledge/risks.md`): partial multi-file writes on interruption (completes
what ADR-005 began), stale mutation plan applied after target changed
(hash-binding mechanism), and skills bypassing deterministic safety (mutation
stays staged and engine-committed).

## Options considered

### Option A: Engine-managed staged transaction — stage the full change set under `.llm-wiki/staging/<txn-id>/`, validate it whole, then commit by an ordered, journaled sequence of ADR-005 per-file atomic renames

- Pros: composes the exact primitives ADR-005 provides (bounded writes,
  per-file atomic write+rename, the `.llm-wiki/` staging area) into a *single*
  transaction chokepoint, so multi-file all-or-nothing is auditable at one
  place (criteria 3, 20). A staging **manifest** records each target path with
  its staged-content hash **and** the source/target base hashes captured at
  plan time, giving `apply` a deterministic **hash-bound stale-plan rejection**
  (criterion 13) and letting `inspect`/`plan` render the full preview
  (criteria 11, 12) without mutating the tree. Because a non-empty-repo commit
  can overwrite, delete, or metadata-modify a **pre-existing** target, the
  manifest also carries, for every such target, a **preimage record** — a
  durable backup of the prior bytes plus the metadata needed to restore it —
  so the transaction can return the tree to its exact pre-commit state rather
  than only its hash. A commit **journal** records progress so an interrupted
  commit is detectably **partial** and can be rolled forward (finish the
  remaining renames from staging) or rolled back to the pre-commit state (undo
  applied renames from the preimage records); temp/staging is cleaned on
  success or abort. Fully offline and self-contained — no external VCS or lock
  daemon.
- Cons: the manifest schema, commit ordering, and journal/recovery logic are
  real engineering the engine must own and test, including simulated
  interruption at each commit step; ordered atomic-rename recovery has subtle
  cross-platform (Windows) edges that need per-platform fixtures; the staging
  area consumes extra disk for the duration of a transaction.

### Option B: Best-effort sequential writes with compensating rollback

- Pros: least upfront machinery — write each target through the ADR-005 gate in
  turn and, on error, attempt to undo the writes already made.
- Cons: there is no point at which the batch is provably all-or-nothing — an
  interruption (crash, power loss) between two writes leaves a partial commit
  with **no journal to recover from**, exactly the failure criteria 3 and 20
  forbid; compensation is itself a mutation that can fail or be interrupted;
  and there is no staged artifact to preview or hash-bind, so criteria 11–13
  cannot be met. It pushes transaction concerns back to every call site,
  re-creating the drift ADR-005 closed for per-file safety.

### Option C: Delegate the transaction to the host VCS (git stash / worktree)

- Pros: reuses a battle-tested atomic mechanism; "rollback" is a `git` operation.
- Cons: requires a git working tree at a known state (install targets need not
  be git repos, and a dirty worktree is a guard case, not a transaction store);
  couples the engine to git internals and a specific CLI, breaking the
  self-contained/offline promise (ADR-002); cannot cover `.llm-wiki/` engine
  metadata that lives outside version control; and makes the safety of a core
  invariant depend on an external tool the engine does not control.

## Decision

Adopt **Option A** — an engine-managed staged transaction. A mutation stages its
entire change set under `.llm-wiki/staging/<txn-id>/`, writes a **staging
manifest** binding each target path to its staged-content hash plus the
source/target base hashes captured at plan time, **validates the whole staged
set before any commit**, and then **commits** by an ordered, **journaled**
sequence of ADR-005 per-file atomic renames from staging into the tree.
Interruption is recovered from the journal: a partial commit is detected and
either rolled forward (complete the remaining renames) or rolled back to the
pre-commit state; staging and temp files are cleaned on success or abort. This
is the only option that makes multi-file all-or-nothing **provable at one
chokepoint** while staying self-contained (Option C couples to git) and
interruption-safe (Option B has no recovery journal). `apply` refuses any plan
whose recorded base hashes no longer match the live source/target, satisfying
stale-plan rejection (criterion 13).

**Rollback is preimage-backed, not hash-inferred (required for correctness).**
A recorded hash proves *what a file was* but cannot *reconstruct* it, so
rollback to the pre-commit state is defined against durable **preimage
records**, not hashes. Before its first mutating rename, the commit captures —
under the same `.llm-wiki/staging/<txn-id>/` tree, through the ADR-005 gate — a
preimage record for **every existing target the plan will overwrite, delete, or
metadata-modify**. Each preimage record represents the pre-commit target as one
of:

- **absent** — an explicit sentinel meaning the target did not exist before the
  commit; rollback restores absence by removing the committed file;
- **regular file** — the prior bytes preserved as a durable staged backup
  (addressed by their content hash) plus the metadata ADR-005 manages (the mode
  bits it sets on atomic write); rollback restores the bytes and mode by the
  same atomic write+rename.

Rollback replays the preimage records in **reverse commit order** so each
atomic rename is undone before the one that preceded it, and completes only when
every mutated target has been restored; a rollback interrupted by a second crash
is itself resumable from the journal. Roll-forward remains available whenever the
staged postimages are intact and completing the transaction is preferable to
undoing it; the recovery driver chooses forward or back per the journal's
recorded commit point, and both endpoints are exact — either the fully committed
tree or the byte-and-metadata-identical pre-commit tree.

**Base-hash and target representation (consistent with ADR-005).** The
source/target base hashes and preimage records are defined only over **regular
files**. A missing source or target is represented by the same distinct
**absent sentinel** used above, never by the hash of empty content, so "file
absent" and "file empty" are distinguishable at plan and recovery time.
Consistent with the ADR-005 gate, which canonicalizes and resolves symlinks
before any write, a target that resolves to a **non-regular** file (symlink,
directory, device, socket) is **rejected at plan time** rather than staged,
overwritten, or backed up — the transaction commits and rolls back only regular
files inside the boundary. Directories are created as needed to hold committed
files but are not themselves transaction targets.

**Phase split (deliberate).** Phase 2 consumes only the **transaction
mechanism** — stage → validate → journaled atomic-rename commit → recover — as
the substrate for non-empty-repo install and `init`. The `page
inspect/plan/apply` **command surface and preview UX** built on the same
manifest, and dirty-worktree conflict UX, are Phase 3 work; this ADR fixes the
model they rest on, not their CLI ergonomics.

## Consequences

- Easier: multi-file operations become all-or-nothing at one auditable place,
  completing the interruption-safety criteria 3 and 20 that ADR-005's per-file
  atomicity only *began*; the staging manifest gives criteria 11 and 12 a
  concrete preview artifact and criterion 13 a deterministic hash-bound
  rejection; every write still flows through the ADR-005 gate, so the
  zero-out-of-boundary hard gate is unaffected.
- Harder: the engine must implement and test a manifest schema, deterministic
  commit ordering, a recovery journal, **preimage capture and reverse-order
  restore** (including the absent sentinel and non-regular-file rejection), and
  partial-commit detection, including simulated interruption at each commit and
  rollback step and per-platform atomic-rename fixtures (notably Windows).
- Maintain: the `.llm-wiki/staging/<txn-id>/` layout, the staging-manifest,
  **preimage-record**, and journal formats, the commit/recovery sequencing
  (roll-forward and preimage-backed rollback), and temp-cleanup — all composed
  strictly from ADR-005 primitives (this ADR adds no new write path).
- Deferred / validation implications: criteria 3, 11, 12, 13, 20 are the
  acceptance hooks; Phase 2 proves the **transaction** half under injected
  interruption, Phase 3 proves the `inspect/plan/apply` surface and stale-plan
  UX. Upgrade/uninstall preservation (criteria 3, 20 at lifecycle scope) runs
  *on* this transaction model but its ownership rules are
  [ADR-009](adr-009-install-upgrade-uninstall-ownership.md)'s domain; the
  provenance/citation-preservation content rules are out of scope here. The
  partial-multi-file-write and stale-plan risks in `knowledge/risks.md` stay
  `open` until the engine lands, with their mechanism now owned here.
