
## Hermes hardened workflow (overlay — non-negotiable)

These rules extend the Workflow rules above. They exist because of the
llm-wiki-kit Phase 5–7 failure (July 2026): a review outage plus an
"autonomous mandate" let three phases land on `main` unreviewed, unbuildable,
and with fabricated closeout docs. No session may waive these.

### Roles and model assignments

Roles are fixed; models are substitutable operators of a role. A model
substitution must be equal-or-higher capability for that role and is
recorded in `knowledge/log.md`. Roles never collapse into one identity —
above all, no author ever reviews its own work.

| Role | Assigned | Fallback | Notes |
|---|---|---|---|
| Owner / ratifier | Oliver (human) | none — cannot be delegated | Approves plans, ratifies ADR acceptances, holds admin credentials, receives halt reports. |
| Planner / orchestrator | Claude Fable 5 | Claude Opus 4.8 if Fable is unavailable | PRD → MVP → ADR drafting, issue planning, closeouts, session driving. |
| Builder | Claude Opus 4.8 | — | Implementation sessions: one issue, one branch, one PR; tests green. |
| Knowledge-layer curator | Claude Haiku | planner model | Maintains `knowledge/` (log, index, risks, open questions) per `SCHEMA.md`. Curates and records; closeout content still passes the closeout gates. |
| Adversarial reviewer | OpenAI Codex | another **non-Anthropic** reviewer of equivalent rigor | Must be cross-vendor and context-independent from the author. Unavailable → HALT (see below); the author's own model family re-reading the artifact is not a review. |
| Guard + branch protection | no model (deterministic CI + GitHub settings) | — | Deliberately non-agentic; cannot be argued with. |

### Gates cannot be skipped, only failed

- **Adversarial review is a hard gate.** Every ADR, PRD/normalization
  output, and phase closeout gets an independent adversarial review (Codex
  or equivalent) reaching `READY` before acceptance/merge. If the reviewer
  is unavailable (usage limit, outage, missing CLI): **HALT the phase and
  report to Oliver.** "Review deferred, covered by tests" is exactly the
  rationalization that caused the failure — it is forbidden.
- **No mandate may remove a gate.** An autonomous-phase mandate can
  delegate *who operates* a gate, never whether it runs. Plan-first
  approval, issue-per-PR, adversarial review, and green tests apply in
  every session, supervised or not.

### Ratification debt

- At most **one phase** of ADRs may be "accepted under mandate, awaiting
  Oliver's async ratification" at any time. The next phase's issues may
  not be filed until the backlog is ratified. (In the failure, five
  unratified ADRs had normalized "accept now, ask later".)

### Mechanical enforcement assumed

- `main` has branch protection: PRs only, required status checks
  (`guard` + full test matrix), no force pushes, enforced for admins.
  Never attempt to bypass it, and treat its absence as a setup bug to fix
  before feature work.
- An unexpected committer identity in `git log main` is a stop-the-line
  event: halt, report, do not build on top of it.

### Phase and closeout discipline

- A phase's prerequisite ADRs are drafted **and accepted** before its
  implementation issues open. No implementation commit may cite an ADR
  that does not exist.
- Closeouts are atomic: `design/state.md`, `knowledge/log.md`, and
  `knowledge/index.md` update in the **same** closeout PR, and a phase is
  "closed" only when that PR merges. If these three files disagree, the
  most conservative one is true and the disagreement is a bug to fix first.
- **Evidence honesty.** Never write a coverage number, CI verdict, PR
  link, or "criterion satisfied" claim that was not directly observed;
  cite the run ID / commit. A `pull/new/...` URL is not a PR.

### Repo hygiene

- No compiled binaries or files >1 MB in commits (guard-enforced).
- No new runtime dependency without an ADR. Replacing a dependency an
  accepted ADR chose (e.g. the YAML library) requires a superseding ADR
  first — never edit or contradict an accepted ADR in place.
- Rewrites that delete exported API used elsewhere in the repo are
  architecture changes: ADR first, then a plan, then a PR.
