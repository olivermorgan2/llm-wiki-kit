# knowledge/ — schema and conventions

The knowledge layer is the project's curated memory: what was decided, why,
what's still open, and what could go wrong. It complements (never replaces)
the tool-maintained `design/` artifacts.

## Files

| File | Role | Rules |
|---|---|---|
| `index.md` | Live front door: current phase, next action, pointers. | Must never claim more than `log.md` records — if they disagree, `log.md` wins. Changes together with `log.md` (guard-enforced). |
| `log.md` | Append-only chronological decision & review log. | Format: `### date — decision`, then rationale + links. Never rewrite past entries. Architectural items get promoted to an ADR with a back-reference. |
| `project-brief.md` | Distilled product definition and architectural spine. | Updated when scope/architecture genuinely shifts. |
| `risks.md` | Live risk register. | An accepted ADR *bounds* a risk; it does not retire it before the engineering exists. Rows carry status: open / mitigated / retired. |
| `open-questions.md` | Numbered questions with status. | Statuses: open / assumption-locked / closed (with the closing decision linked). Only a human decision or an accepted ADR closes a question. |
| `reviews/` | Verbatim adversarial-review artifacts. | One file per review, named `YYYY-MM-DD-<reviewer>-<subject>.md`, archived verbatim. Every ADR acceptance and phase closeout references one. |
| `sources/` | Verbatim inputs (original PRD, etc.). | Never edited. |

## Invariants

- Closeouts are atomic: `index.md`, `log.md`, and `design/state.md` update
  in the same PR.
- Every "accepted" claim links a review artifact in `reviews/`.
- Evidence honesty: no unverified numbers, verdicts, or links anywhere in
  the layer.
