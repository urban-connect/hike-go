---
name: create-pull-request
description: Create a new branch, commit all changes, push, and open a PR.
---

# Create Pull Request

Commit, push, and open a PR.

## Context

Current branch: !`git branch --show-current`
Working tree status: !`git status --short`

## Steps

1. **Create a new branch** from the current branch
   - Pick a short, readable name describing the changes
   - Never use a prefix (no `feat/`, `fix/`, `chore/`, etc.) — just the descriptive name
2. **Stage and commit** all changes
   - Stage specific files — never use `git add -A` or `git add .`
   - Write a concise commit message (1-2 sentences) focusing on the "why"
3. **Push** the branch with `-u origin <branch>`
4. **Open a PR** with `gh pr create`
   - Title under 70 characters
   - Body must follow the **PR Description Structure** below (no test cases)
   - Assign to `@me`
5. **Stop here** — return the PR URL to the user

## PR Description Structure

Every PR body must contain exactly these three `##` (level-2) sections, in this order. If a section has no content, write `None` under it.

```markdown
## Summary

<Brief summary of what has been done and why.>

## Impact on consumers

<hike-go is a library. Note any breaking API change that would require consumers
to update. If the change is additive or internal, write `None`.>

## Links

<Links to relevant PRs or tickets. If none, write `None`.>
```

## Rules

- **Do not merge unless the user explicitly asks to merge** — if asked, use the `merge-pull-request` skill
- **Never prefix branch names** (no `feat/`, `fix/`, `chore/`, etc.) — use a short, readable descriptive name only
- Follow all Git & PR conventions from CLAUDE.md
- Never force push
- Never add Claude as a co-author
- Always assign the PR to its author
