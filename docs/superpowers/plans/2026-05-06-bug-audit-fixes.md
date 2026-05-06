# Bug Audit Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply 5 low-risk improvements from the MarkdownViewer bug audit: regression tests for critical bugs, standardized clamping, GrowMode consistency, block renderer checklist, and style audit test.

**Architecture:** All changes are within the existing MarkdownViewer subsystem (3 files). No new files, no new dependencies, no API changes. Each change is independent and small (1-30 lines).

**Tech Stack:** Go, tcell v2, goldmark (unchanged)

**Note:** No e2e task — these are internal hardening changes with no new user-visible capability. Existing e2e tests (`TestMarkdownViewer*`) cover regression.

---

### Task 1: Code Fixes (GrowMode, Clamping, Checklist)

**Files:**
- Modify: `tv/markdown_viewer.go:22-29` (constructor), `tv/markdown_viewer.go:268-292` (KeyPgDn/KeyEnd)
- Modify: `tv/markdown_render.go:669-690` (renderBlockLine dispatch — add checklist comment)

**Requirements:**

1. `NewMarkdownViewer` sets `GfGrowHiX | GfGrowHiY` on the widget, matching the pattern used by Memo (memo.go:45), OutlineViewer (outline.go:486), and ListBox (list_box.go:22).
2. `KeyPgDn` does NOT compute `maxDY` explicitly — it increments `deltaY` by `vpH` and delegates clamping to `syncScrollBars()`, matching the `KeyDown` pattern.
3. `KeyEnd` does NOT compute `maxDY` explicitly — it sets `deltaY = totalH` (maximum possible) and delegates clamping to `syncScrollBars()`, matching the `KeyDown` pattern. This also fixes the missing `len(mv.blocks) > 0` guard.
4. A checklist comment block is added above `renderBlockLine` in `markdown_render.go` documenting the 6 responsibilities every block renderer must handle.

**Implementation:**

In `NewMarkdownViewer` (after `mv.SetSelf(mv)`):
```go
mv.SetGrowMode(GfGrowHiX | GfGrowHiY)
```

In `HandleEvent`, replace `KeyPgDn` case:
```go
case tcell.KeyPgDn:
	mv.deltaY += vpH
	mv.syncScrollBars()
	event.Clear()
```

In `HandleEvent`, replace `KeyEnd` case:
```go
case tcell.KeyEnd:
	mv.deltaY = totalH
	mv.syncScrollBars()
	event.Clear()
```

In `markdown_render.go`, add before `func (r *mdRenderer) renderBlockLine`:
```go
// Block renderer checklist — every new mdBlockKind must handle:
//   1. Height: blockHeight switch case accounting for depth indent
//   2. MaxWidth: blockMaxWidth switch case accounting for depth indent + content width
//   3. Render: renderBlockLine dispatch + render function using composeStyle for inline runs
//   4. Background fill: start at (depth*2 - dx) clamped to [0, w), not x=0
//   5. Nested children: blocksHeight / renderBlocksInto at depth+1 for any child blocks
//   6. Blank lines: height function and render function must agree on blank-line placement
```

**Run tests:** `go test ./tv/ -run TestMarkdownViewer -count=1`

**Commit:** `git commit -m "fix: set GrowMode, standardize clamping, add block renderer checklist"`

---

### Task 2: Regression Tests + Style Audit

**Files:**
- Modify: `tv/markdown_viewer_test.go` (append new tests)

**Requirements:**

1. **defListHeight includes nested children:** Create a definition list where one item has a child paragraph block. `renderedHeight()` must be greater than the same list without children. The child content must appear in `Draw()` output.

2. **defListLine blank lines between items:** Create a definition list with 2 items ("Term1"/"Definition one" and "Term2"/"Definition two"). Render at width 40. Verify a blank line (all spaces) appears between the first definition and the second term.

3. **Code block inside blockquote:** Create a blockquote containing a code block with text "code in blockquote". Render and verify: (a) the blockquote bar character '▌' is at the correct x position, (b) the code block background fill is present (code block style, not normal style), (c) the code text "code in blockquote" is in the rendered output.

4. **Table header style:** Create a table with headers "Name" and "Type" and one body row "foo"/"bar". Render and verify: (a) a cell in the header row has style `MarkdownBold`, (b) a cell in the body row has style `MarkdownNormal`, (c) `MarkdownBold` and `MarkdownNormal` foreground colors differ (using `Decompose()`).

5. **Style audit (BorlandBlue):** Using the BorlandBlue theme, verify the following styles are visually distinct from `MarkdownNormal` via full `Decompose()` comparison (foreground + background + attributes). The implementation must use `Decompose()` — do NOT hardcode expected colors. The parenthetical notes below describe the actual BorlandBlue theme, but assertions must use Decompose() comparison, not hardcoded values:
   - `MarkdownBold` — Bold attribute present
   - `MarkdownItalic` — Italic attribute present
   - `MarkdownCode` — background differs
   - `MarkdownH1` through `MarkdownH5` — each has different fg/bg/attrs
   - `MarkdownLink` — foreground + Underline differ
   - `MarkdownBlockquote` — foreground differs
   - `MarkdownTableBorder` — foreground + Bold differ
   - `MarkdownHRule` — foreground differs
   - `MarkdownListMarker` — foreground differs
   - `MarkdownDefTerm` — foreground + Bold differ
   Note: `MarkdownH6` is intentionally identical to `MarkdownNormal` and is excluded.

**Implementation:**

Test setup follows existing patterns in `markdown_viewer_test.go`:
- Tests 1-4: Use `parseMarkdown` to create IR, construct `mdRenderer` with `wrapText: true`, use `DrawBuffer` for rendering output verification, `Decompose()` for style assertions.
- Test 5: Use `theme.BorlandBlue` directly, iterate style pairs, call `Decompose()` on each, compare foreground/background/attributes.

**Run tests:** `go test ./tv/ -run "TestMarkdownViewer" -count=1`

**Commit:** `git commit -m "test: add regression tests for defList, blockquote-code, table header style, and style audit"`
