# Bug Audit Fixes — Design Spec

> Derived from [bug audit](../bug-audits/2026-05-06-markdown-viewer.md) recommendations 1, 2, 3, 5, 6. Recommendation 4 (shared document walker) deferred.

## 1. Regression Tests for Critical Bugs

### 1a. defListHeight includes nested children

**Bug:** `defListHeight` never added height of `item.children`, so nested content inside definition items was invisible.

**Fix (already applied):** Added `blocksHeight(item.children, depth+1)` call.

**Regression test:** Create a definition list where one item has a child paragraph block. Verify `renderedHeight()` is greater than without the child. Verify the child content renders in `Draw()` output.

### 1b. defListLine blank lines between items

**Bug:** `defListHeight` accounts for blank lines between items, but `renderDefListLine` did not — the height counter and render cursor diverged for multi-item definition lists.

**Fix (already applied):** Added blank-line handling matching height calculation.

**Regression test:** Create a definition list with 2+ items. Render and verify blank lines appear between items. Verify the second term renders at the correct row (not overlapping the first definition).

### 1c. Code blocks inside blockquotes inherit bar

**Bug:** Block styles must flow through nesting hierarchy. No explicit test verifies that a code block inside a blockquote renders both the blockquote bar and code block background.

**Regression test:** Create a blockquote containing a code block. Render and verify: (a) the blockquote bar character (▌) appears at the correct indent, (b) the code block background fill is present, (c) code text renders on top.

### 1d. Table header cells use MarkdownBold

**Bug:** Header cells rendered with `MarkdownNormal` instead of `MarkdownBold`.

**Fix (already applied):** Pass `cs.MarkdownBold` for header row.

**Regression test:** Render a table with headers. Verify header cell style uses `MarkdownBold` foreground color/attributes. Verify body cell style uses `MarkdownNormal`. Verify the two are distinct.

---

## 2. Standardize Clamping Contract

**Bug:** `KeyDown`/`KeyRight` rely on `syncScrollBars` to clamp delta; `KeyPgDn`/`KeyEnd` compute bounds explicitly. `KeyEnd` lacks the `len(mv.blocks) > 0` guard that `KeyPgDn` has.

**Design:** Remove explicit bound computation from `KeyPgDn` and `KeyEnd`. Both should follow the same pattern as `KeyDown`/`KeyRight`: mutate delta, then call `syncScrollBars()` which already has comprehensive clamping with all guards.

- `KeyPgDn`: `mv.deltaY += vpH; mv.syncScrollBars()` — no manual clamping
- `KeyEnd`: `mv.deltaY = totalH; mv.syncScrollBars()` — set to max, let syncScrollBars clamp

This also fixes the missing `len(mv.blocks) > 0` guard in `KeyEnd` since `syncScrollBars` already has it.

---

## 3. Set GrowMode in Constructor

**Bug:** Memo, OutlineViewer, and ListBox all set `GfGrowHiX | GfGrowHiY` in their constructors. MarkdownViewer does not, so it doesn't grow when its parent window resizes unless the caller explicitly sets it.

**Fix:** Add `mv.SetGrowMode(GfGrowHiX | GfGrowHiY)` to `NewMarkdownViewer`, matching all other scrollable widgets.

---

## 5. Block Renderer Checklist

**Bug:** Several bugs (code block bg fill ignoring depth indent, style propagation through nesting, height/render blank-line mismatch) share the pattern: new block type renderers miss requirements that aren't explicitly documented.

**Fix:** Add a checklist comment in `markdown_render.go` above the `renderBlockLine` dispatch function. The checklist covers the 6 things every block renderer must handle:

1. Height function — account for depth indent
2. Max-width function — account for depth indent + content
3. Render function — propagate `parentBlockStyle` through `composeStyle`
4. Background fill — use depth indent as start offset
5. Nested children — call `blocksHeight`/`renderBlocksInto` at `depth+1`
6. Blank lines — height and renderer must agree on blank-line placement

---

## 6. Style Audit Test

**Bug:** Weak style assertions in existing tests verify text content but not that styles are visually distinct. A theme change that made all Markdown styles identical would not be caught.

**Test:** Using the BorlandBlue theme, verify that key style pairs are visually distinct. Use full `Decompose()` comparison (foreground + background + attributes), not just foreground color — styles can differ by Bold, Italic, background, etc. while sharing the same foreground.

Verify these styles differ from `MarkdownNormal` (LightGray on Blue, no attrs):
- `MarkdownBold` — same fg/bg, Bold attr
- `MarkdownItalic` — same fg/bg, Italic attr
- `MarkdownCode` — different background (Cyan)
- `MarkdownH1` through `MarkdownH5` — various fg/attribute differences
- `MarkdownLink` — Yellow foreground
- `MarkdownBlockquote` — DarkGray background
- `MarkdownTableBorder` — DarkCyan foreground, Bold
- `MarkdownHRule` — DarkCyan foreground
- `MarkdownListMarker` — DarkCyan foreground
- `MarkdownDefTerm` — White foreground, Bold

Note: `MarkdownH6` in BorlandBlue is intentionally identical to `MarkdownNormal` (matches original TV convention where H6 uses the same formatting as body text). This is not tested as "distinct."
