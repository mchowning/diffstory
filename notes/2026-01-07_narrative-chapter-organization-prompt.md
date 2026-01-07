---
date: 2026-01-07T17:01:24-05:00
git_commit: 3fee8737c57e08d565d280f4d6dfe125f38cb0da
branch: main
repository: diffstory
topic: "Narrative Chapter Organization Prompt"
tags: [implementation, llm-prompt, tui]
last_updated: 2026-01-07
---

# Narrative Chapter Organization Prompt

## Summary

Updated the LLM classification prompt to guide chapter organization by functional purpose rather than file path, addressing the original success metric of presenting changes as a coherent narrative story.

## Overview

The LLM prompt for classifying diff hunks into chapters and sections was producing file-based groupings (e.g., "Changes to auth.go") and overly fragmented sections with single trivial hunks. This undermined the tool's core purpose of presenting code changes as a readable narrative.

The fix adds explicit functional grouping guidance to the fixed prompt template and replaces the fragmentation-encouraging default instructions with balanced section sizing guidance. The changes are split between non-editable constraints (which would break the tool if changed) and user-editable style preferences.

## Technical Details

The implementation required careful separation between fixed prompt content and user-editable defaults, based on existing architecture that exposes reviewer instructions in the UI (see commit `7b74d7f`).

### Fixed Prompt Template (`internal/tui/generate.go`)

Added a "Grouping Philosophy" section with core constraints that should not be user-editable:

```go
## Grouping Philosophy

Group hunks by FUNCTIONAL PURPOSE, not by file path. Hunks that work together to achieve a goal belong in the same section, even if they span multiple files.

DO NOT:
- Group by file path. "Changes to auth.go" is never a good chapter title.
- Separate documentation into its own chapter. Docs belong with their related code.
```

The anti-file-grouping constraint is in the fixed section because file-based chapters would undermine the tool's core purpose. This is a structural requirement, not a style preference.

The existing guidelines were reorganized under a "## Format Requirements" header for clarity, with the "small, focused" qualifier removed from the section description since it was encouraging fragmentation.

### Editable Default Instructions (`internal/tui/model.go`)

Replaced the previous default that encouraged fragmentation:

```go
// Before
const DefaultReviewerInstructions = "Prefer many small, focused sections over fewer large ones. Keep section narratives to 1-2 sentences explaining what and why. "

// After
const DefaultReviewerInstructions = `Section sizing: Combine trivial hunks (imports, formatting, small fixes) with the substantial changes they support. Don't leave trivial hunks isolated, but also don't combine unrelated hunks just to reduce section count. A substantial hunk can stand alone; trivial hunks should join their related work.

Keep section narratives to 1-2 sentences explaining what and why. `
```

The new default introduces the "trivial vs substantial" distinction to provide balanced guidance:
- Trivial hunks (imports, formatting) should be combined with related substantial changes
- Substantial hunks can stand alone
- Unrelated hunks should not be combined just to reduce section count

This prevents both over-fragmentation (many single-hunk sections) and over-combination (unrelated hunks lumped together).

### Design Decision: Fixed vs. Editable Split

| Fixed (non-editable) | Editable (user can modify) |
|---------------------|---------------------------|
| JSON format requirements | Section sizing preferences |
| "MUST classify ALL hunks" | Trivial vs substantial distinction |
| Anti-file-grouping constraint | Narrative style (1-2 sentences) |
| Documentation coupling rule | Additional context |
| Title length guidance | |
| Importance/isTest requirements | |

## Git References

**Branch**: `main`

**Commit Range**: Single commit

**Commits Documented**:

**3fee8737c57e08d565d280f4d6dfe125f38cb0da** (2026-01-07)
Improve LLM prompt for narrative chapter organization

Replace file-based grouping guidance with functional cohesion principles.
Addresses original success metric: "Present changes as a coherent story,
not grouped by file."

Fixed template (non-editable):
- Add "Grouping Philosophy" section emphasizing functional purpose
- Explicit "DO NOT" constraints: no file-path chapters, no isolated docs
- Clarify that chapters contain sections

Editable defaults:
- Replace "prefer many small sections" (caused fragmentation)
- Add balanced section sizing: combine trivial hunks, don't over-combine
- Clarify trivial vs substantial hunk distinction
