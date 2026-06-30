# start — worked routings

Three runs of `/start` against different project states. Each shows the
inputs the skill reads and the brief it emits.

---

## 1. Fresh project (onboarding)

**State.** New repo. No `design/prd.md`, no `design/` artefacts, no
prompts, clean tree on `main`.

**`/start`**

> **Mode:** interactive (default).
> **Where you stand:** brand-new project — only an idea, no PRD yet.
> **Next:** `/idea-to-prd` — turn your idea into a lightweight PRD.
> This is cat-1 (local). Want me to start it now? Tell me the idea in a
> sentence or two and I'll draft `design/prd.md`.

Rule matched: #2 (no PRD, no scoping artefacts).

---

## 2. Mid-flow, healthy `state.md`

**State.** `design/state.md` present with a `next-action` zone:

```yaml
skill: prepare-issue
args: "95"
preconditions:
  - "issue #95 exists and is open"
blocked-by: none
```

**`/next`**

> **Mode:** assisted.
> **Where you stand:** issue #95 selected, no prompt on disk yet.
> **Next:** `/prepare-issue 95` — fill the implementation prompt from
> the issue and its linked ADR. cat-1; running it now.

Rule matched: #1 (actionable `next-action` zone). Because the active
mode is `assisted` and the target is cat-1, `/start` proceeds to invoke
`/prepare-issue 95` rather than only recommending it.

---

## 3. Ambiguous — one clarifying question

**State.** A prepared prompt `prompts/issue-095-export-csv.md` exists,
**and** a feature branch `fix-94-pagination` is ahead of `main` with no
open PR. No `next-action` zone (older `state.md`).

**`/start`**

> **Mode:** interactive.
> Two things are in flight: an open branch `fix-94-pagination` ready to
> ship, and a prepared prompt for #95 ready to build. Which first —
> open the PR for #94, or start building #95?

One question, then route: a "ship #94" answer → recommend `/ship`
(`/pr-review-packager`, cat-3 — recommend, do not invoke); a "build #95"
answer → recommend `/work` (`/claude-issue-executor`).

Rules #6 and #7 both matched, so the skill asked one question rather than
guessing.
