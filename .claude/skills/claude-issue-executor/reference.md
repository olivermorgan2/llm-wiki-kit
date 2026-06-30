# claude-issue-executor — reference

This file contains edge-case handling, design-questions populate
rules, and the end-of-session alignment checklist referenced from
[`SKILL.md`](SKILL.md). The session protocol, plan gate, and
evaluation summary structure stay in SKILL.md.

## Edge cases

- **Prompt file not found.** Search `prompts/` and `notes/` for the
  closest match and ask the user which they meant. Never guess silently.
- **Prompt file malformed.** Report the missing sections by name and
  ask whether to fix the prompt or to proceed with explicit user
  guidance for the gaps.
- **Working tree dirty on entry.** Refuse to proceed. Do not auto-stash.
  Ask the user what to do.
- **Target branch already exists.** Ask whether to (a) switch to the
  existing branch and continue, (b) pick a new name, or (c) delete the
  existing one (only with explicit confirmation and only when the user
  is sure nothing there is needed). Never force-delete silently.
- **User denies the plan.** Offer to revise. If the user wants to stop
  entirely, stop — do not keep pushing variants.
- **Tests fail during implementation.** Stop, report, and ask — do not
  paper over failing tests to keep the commit cadence.
- **ADR or issue number missing from the prompt.** Malformed — see
  **Prompt validation** in SKILL.md. Do not substitute a placeholder
  in commit messages.

## design-questions populate rules

**When to populate.** Add an entry only when *all three* hold: the
question concerns a load-bearing constraint, a specific upcoming
issue depends on the answer, and this issue's commits do not fully
resolve it. (Full rule in §6 of the workflow guide — when adding a
new entry, follow §6's rule, not a restatement here.)

**When NOT to populate.** Skip the entry — even if a design
question came up — when *any* of the following hold: (1) the
question was self-resolved by this issue's commits; (2) no
upcoming filed-or-planned issue depends on the answer (capture in
`notes/feature-ideas.md` instead); (3) the question is purely
implementation tactics with no cross-issue coupling; (4) the
answer is already in an ADR or `design/decisions.md`.

If borderline, **omit** the entry. False positives cost more than
false negatives — see §6's rationale.

**Empty case.** If there are no entries, **omit the entire
`### design-questions` block**. Do not emit `design-questions: []`
or an empty heading.

## Alignment check

Before finishing the session, confirm:

- [ ] The plan was proposed and explicitly approved before any edits.
- [ ] A feature branch was created from `main` (not committed to `main`
  directly).
- [ ] Every commit message includes `ADR-NNN` and `#issue` from the
  prompt.
- [ ] Tests, if applicable to the project, landed with the code.
- [ ] An evaluation summary was printed **and persisted to
  `notes/eval-issue-NNN.md`**.
- [ ] If the session raised a cross-issue design question matching
  the §6 when-to-populate rule, a `### design-questions` block
  was added to the persisted eval summary; otherwise the block was
  omitted.
- [ ] `/pr-review-packager` was suggested, not auto-invoked.

If any box is unchecked, the skill has drifted — say so in the
evaluation summary rather than hiding it.
