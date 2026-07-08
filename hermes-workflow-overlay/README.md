# hermes-workflow-overlay

A personal hardening layer applied *on top of* a vanilla
[claude-workflow-kit](https://github.com/olivermorgan2/claude-workflow-kit)
install. The kit stays simple and generic; this overlay adds the gates that
the llm-wiki-kit postmortem (July 2026) showed are load-bearing:
adversarial reviews that cannot be silently skipped, a knowledge layer,
and mechanical (not procedural) enforcement of the PR flow.

**This folder is a standalone repo-in-waiting.** Move it out of any project
checkout (`mv hermes-workflow-overlay ~/`) and `git init` it, or keep it
wherever you keep personal tooling.

## What's in it

| File | Purpose |
|---|---|
| `CLAUDE-overlay.md` | Hard rules to append to a project's `CLAUDE.md` after kit install. |
| `.github/workflows/workflow-guard.yml` | CI process-invariant checks (PR hygiene, no binaries, review receipts, knowledge-layer consistency). |
| `scripts/install-overlay.sh` | One-command apply: guard workflow + knowledge skeleton + CLAUDE.md section. |
| `scripts/protect-branch.sh` | One-shot GitHub branch-protection setup (blocks direct/force pushes to `main`, requires status checks, applies to admins). |
| `knowledge/` | Day-one knowledge-layer skeleton (`SCHEMA.md`, `index.md`, `log.md`, `project-brief.md`, `risks.md`, `open-questions.md`, `reviews/`). |

## Applying to a project

1. Bootstrap the project with the vanilla kit as usual.
2. Apply the overlay:
   ```bash
   scripts/install-overlay.sh /path/to/project
   ```
3. Commit: directly to `main` on a fresh repo (protection isn't on yet);
   via PR when retrofitting an existing protected repo.
4. After the guard workflow has run once (so GitHub knows the check name),
   enable protection — **last**, after any history surgery:
   ```bash
   scripts/protect-branch.sh owner/repo "guard" "<test-matrix job names...>"
   ```
5. Audit collaborators: only the intended account(s) should have write
   access. An unexpected committer identity on `main` is a
   stop-the-line event.

## Design notes

- Branch protection is the only gate an unsupervised agent cannot talk
  itself past. Everything in `CLAUDE-overlay.md` is defense-in-depth
  around that.
- PR review-approval is deliberately **not** required (an autonomous agent
  cannot self-approve, and Oliver ratifies async). The guard checks +
  test matrix are the merge gate; ratification debt is capped by rule
  instead (see CLAUDE-overlay.md §Ratification debt).
