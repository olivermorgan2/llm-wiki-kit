# Troubleshooting

Organized by the **symptom you actually see**, not by cause. Find your
symptom heading, apply the fix. If you're mid-tutorial, this is the
page the "if something fails" note points to.

For install-specific errors (`gh: command not found`, wrong kit path),
see also the [install troubleshooting section](install.md#troubleshooting).

---

## A slash command does not autocomplete

You type `/prd` and Claude Code shows no matching skill.

- **You're not in the project root.** Project-local skills live under
  `.claude/skills/` and are discovered relative to where you launched
  Claude Code. Quit, `cd` to the project root (the folder containing
  `CLAUDE.md`), and run `claude` again.
- **The skills aren't installed.** Run `ls .claude/skills`. If it's
  empty or missing, the install didn't complete — re-run the
  installer (see [Quick start](../README.md#quick-start)).
- **`.claude/` is hidden.** It starts with a dot, so Finder/Explorer
  hides it. Verify from the terminal, not the file browser:
  `ls .claude/skills` should list the skill directories.

See also: [Skills are not discovered](#skills-are-not-discovered).

---

## Skills are not discovered

Claude Code starts but acts as if no kit skills exist.

- **Wrong working directory.** Run Claude Code from the project root.
  If you `cd` into a subdirectory the skills still resolve, but
  launching from an unrelated directory will not find them.
- **Confirm they're on disk:**

  ```bash
  ls .claude/skills        # should list idea-to-prd, prd-to-mvp, …
  ```

  If the directory is missing, re-run the installer. If it exists but
  Claude Code still can't see the skills, fully quit and restart
  Claude Code so it re-scans the project.

---

## GitHub project creation fails

`/issue-planner` reaches the project-board step and `gh project create`
fails — typically with a message about missing scopes or insufficient
permissions.

Project boards need a GitHub token scope that the default `gh auth
login` does not grant. Add it:

```bash
gh auth refresh -s project --hostname github.com
gh auth status        # confirm the project scope is now listed
```

Then re-run `/issue-planner`. Creating issues only needs the standard
`repo` scope; only the **project board** step needs `project`. If you
don't need a board, the issues themselves will still have been created
— check with `gh issue list`.

---

## The prepared prompt is stale

`/prepare-issue` (or the executor) warns that the prompt under
`prompts/issue-NNN-*.md` is out of date, or you edited the GitHub
issue after the prompt was generated.

This is expected and safe. The prompt is a generated artifact derived
from the issue body, the build-out plan, and linked ADRs — not a
hand-authored source of truth. When the issue changes, regenerate it:

```text
/prepare-issue <N>
```

Re-running overwrites the prompt with a fresh version. You lose nothing
of value: anything worth keeping lives in the issue, the ADRs, or the
build-out plan, all of which the regeneration reads back in.

---

## Re-running the installer changes files unexpectedly

You ran `bin/install-workflow-kit` a second time and it modified files
you didn't expect, or you expected it to and it didn't.

- **By default the installer is idempotent.** It skips files that
  already exist and makes no commit when nothing changed
  (`nothing to commit (idempotent re-run)`). A clean re-run should be
  a no-op.
- **`--force` is what changes things.** Passing `--force` re-renders
  `CLAUDE.md` and re-copies the skills on top of what's there. Only
  use it when you deliberately want to overwrite local edits.
- **`CLAUDE.md` was regenerated and lost your edits?** That only
  happens with `--force`. Recover your version from git
  (`git checkout CLAUDE.md` or restore from history) and re-apply
  edits without `--force`.
- **Unexpected `{{PLACEHOLDER}}` or `_TBD_` values?** The installer
  fills only the four required fields (project name, GitHub owner,
  GitHub repo, default branch) and marks every other field `_TBD_`.
  `_TBD_` means "unknown but acceptable" — fill it in when you know the
  answer. See [`docs/install.md`](install.md#3a-bootstrap-install-recommended).

---

## Still stuck

- Install and prerequisites: [`docs/install.md`](install.md)
- The end-to-end workflow: [`docs/workflow-guide.md`](workflow-guide.md)
- GitHub labels, milestones, scopes: [`docs/github-setup.md`](github-setup.md)
- What each skill does: [`docs/skills.md`](skills.md)
