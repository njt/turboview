# Markdown Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an interactive MarkdownEditor widget with reveal-on-cursor syntax editing, smart list continuation, auto-format, and keyboard shortcuts.

**Architecture:** MarkdownEditor embeds Editor→Memo for source editing (cursor, selection, undo, scroll). On each edit, goldmark re-parses source to []mdBlock. A custom Draw renders formatted blocks with syntax markers revealed when cursor is in scope. Reveal scope is computed per-parse based on cursor position.

**Tech Stack:** Go, tcell/v2, goldmark (already integrated), existing tv package widget framework

---

## File Structure

| File | Purpose | Phase |
|---|---|---|
| `tv/markdown_editor.go` | MarkdownEditor struct, constructor, public API | 1 |
| `tv/markdown_editor_draw.go` | Custom Draw — renders blocks with formatted output | 1 |
| `tv/markdown_editor_reveal.go` | Reveal mapper, revealSpan, cursor-in-scope checks | 2 |
| `tv/markdown_editor_autoformat.go` | Auto-format detection, smart list continuation, keyboard shortcuts | 3 |
| `tv/markdown_editor_paste.go` | Paste detection and dispatch | 3 |
| `tv/markdown_editor_test.go` | Unit tests (written by test writer) | 1-3 |
| `tv/markdown_editor_integration_test.go` | Integration tests (written by test writer) | 1-3 |
| `e2e/markdown_editor_test.go` | E2e tests (written by test writer) | 1-3 |
| `e2e/testapp/basic/main.go` | Demo app — add MarkdownEditor window | 1 |

---

## Phase 1: Core Editing with Formatted Rendering

**Capability:** Type markdown source in an editor window and see it rendered as formatted output. The cursor moves through source lines. Re-parsing happens on each edit. No syntax reveal yet — just formatted rendering of the source.

### Task 1: MarkdownEditor struct and constructor

**Files:**
- Create: `tv/markdown_editor.go`

**Requirements:**
- `MarkdownEditor` struct embeds `*Editor` (which embeds `*Memo`), giving it cursor, selection, scroll, undo, and keyboard input
- `MarkdownEditor` stores `blocks []mdBlock` (parsed output), `sourceCache string` (to detect source changes), and `showSource bool` (source toggle, wired in Phase 3)
- `NewMarkdownEditor(bounds Rect) *MarkdownEditor` creates the widget, initializes Editor→Memo, sets `OfSelectable|OfFirstClick`, sets `GrowMode` to `GfGrowHiX|GfGrowHiY`
- `SetText(s string)` overrides Editor.SetText: calls the embedded Editor.SetText, then calls `reparse()` to populate blocks
- `Text() string` delegates to Editor.Text() (inherited)
- `reparse()` joins `Memo.lines` into a string, runs goldmark parse, stores result in `blocks`
- `reparse()` is called after every edit (not every keystroke — it's called from HandleEvent after Memo processes the edit)
- If the source text hasn't changed (compared to `sourceCache`), `reparse()` is a no-op
- `reparse()` handles empty source: `blocks` is set to empty slice `[]mdBlock{}`

**Implementation:**

```go
package tv

var _ Widget = (*MarkdownEditor)(nil)

type MarkdownEditor struct {
    *Editor
    blocks      []mdBlock
    sourceCache string
    showSource  bool
}

func NewMarkdownEditor(bounds Rect) *MarkdownEditor {
    me := &MarkdownEditor{}
    me.Editor = NewEditor(bounds)
    me.SetSelf(me)
    me.reparse()
    return me
}

func (me *MarkdownEditor) SetText(s string) {
    me.Editor.SetText(s)
    me.reparse()
}

func (me *MarkdownEditor) Text() string {
    return me.Editor.Text()
}

func (me *MarkdownEditor) reparse() {
    src := me.Editor.Text()
    if src == me.sourceCache {
        return
    }
    me.sourceCache = src
    if len(src) == 0 {
        me.blocks = []mdBlock{}
        return
    }
    me.blocks = parseMarkdown(src)
    if me.blocks == nil {
        me.blocks = []mdBlock{}
    }
}

// ShowSource returns the current source toggle state.
func (me *MarkdownEditor) ShowSource() bool { return me.showSource }

// SetShowSource sets the source toggle state.
func (me *MarkdownEditor) SetShowSource(on bool) { me.showSource = on }
```

**Run tests:** `go test ./tv/ -run MarkdownEditor -v`

**Commit:** `git commit -m "feat: add MarkdownEditor struct and constructor"`

---

### Task 2: Custom Draw with formatted markdown rendering

**Files:**
- Create: `tv/markdown_editor_draw.go`
- Modify: `tv/markdown_editor.go` (add HandleEvent that triggers reparse)

**Requirements:**
- `MarkdownEditor.Draw(buf *DrawBuffer)` renders formatted markdown using the existing `mdRenderer`
- When `showSource` is true, Draw delegates entirely to `Memo.Draw(buf)` — raw source view
- When `showSource` is false, Draw renders blocks through `mdRenderer.renderLineInto` for each visible line
- Draw uses the existing color scheme markdown styles via `me.ColorScheme()`
- Draw fills the background with `MarkdownNormal` style
- An `mdRenderer` is constructed with `blocks`, `width` (bounds width), `wrapText=true`, and `cs` (color scheme)
- Scroll position (`deltaX`, `deltaY` from Memo) controls which rendered lines are visible
- Draw only renders lines from `deltaY` to `deltaY + viewportHeight + 5` (overscan buffer)
- If `blocks` is empty, fill the viewport with background and return (no crash)
- `MarkdownEditor.HandleEvent(event *Event)` intercepts edit keys to trigger reparse:
  1. If `showSource`, delegate completely to `Editor.HandleEvent(event)` and return
  2. Call `Editor.HandleEvent(event)` (Memo handles the edit)
  3. If the event was cleared (consumed by Memo), call `reparse()` to refresh blocks
  4. If the event is `EvBroadcast` with `CmIndicatorUpdate`, forward to Editor (for status line updates)
- Scrollbar sync: override `syncScrollBars()` to use rendered content height from `mdRenderer` instead of raw line count, matching how `MarkdownViewer.syncScrollBars()` works

**Implementation for syncScrollBars:**

```go
func (me *MarkdownEditor) syncScrollBars() {
    if me.showSource {
        me.Editor.Memo.syncScrollBars()
        return
    }
    r := me.renderer()
    totalH := r.renderedHeight()
    vpH := me.Bounds().Height()

    if len(me.blocks) > 0 {
        maxDY := totalH - vpH
        if maxDY < 0 {
            maxDY = 0
        }
        if me.Memo.deltaY > maxDY {
            me.Memo.deltaY = maxDY
        }
    }
    if me.Memo.deltaY < 0 {
        me.Memo.deltaY = 0
    }

    if me.Memo.vScrollBar != nil {
        maxRange := totalH - 1 + vpH
        if maxRange < 0 {
            maxRange = 0
        }
        me.Memo.vScrollBar.SetRange(0, maxRange)
        me.Memo.vScrollBar.SetPageSize(vpH)
        me.Memo.vScrollBar.SetValue(me.Memo.deltaY)
    }

    maxW := r.maxContentWidth()
    vpW := me.Bounds().Width()

    if len(me.blocks) > 0 {
        maxDX := maxW - vpW
        if maxDX < 0 {
            maxDX = 0
        }
        if me.Memo.deltaX > maxDX {
            me.Memo.deltaX = maxDX
        }
    }
    if me.Memo.deltaX < 0 {
        me.Memo.deltaX = 0
    }

    if me.Memo.hScrollBar != nil {
        maxRange := maxW - 1 + vpW
        if maxRange < 0 {
            maxRange = 0
        }
        me.Memo.hScrollBar.SetRange(0, maxRange)
        me.Memo.hScrollBar.SetPageSize(vpW)
        me.Memo.hScrollBar.SetValue(me.Memo.deltaX)
    }
}
```

**Implementation for Draw:**

```go
func (me *MarkdownEditor) Draw(buf *DrawBuffer) {
    if me.showSource {
        me.Editor.Memo.Draw(buf)
        return
    }

    w := me.Bounds().Width()
    h := me.Bounds().Height()
    cs := me.ColorScheme()
    normalStyle := tcell.StyleDefault
    if cs != nil {
        normalStyle = cs.MarkdownNormal
    }
    buf.Fill(NewRect(0, 0, w, h), ' ', normalStyle)

    if len(me.blocks) == 0 {
        // Draw cursor on empty doc
        if me.HasState(SfSelected) {
            buf.WriteChar(0, 0, ' ', cs.MemoSelected)
        }
        return
    }

    r := me.renderer()
    for row := 0; row < h; row++ {
        lineY := me.Memo.deltaY + row
        r.renderLineInto(buf, lineY, row, me.Memo.deltaX, w)
    }

    // Draw cursor at source position mapped to rendered position
    me.drawCursor(buf, cs)
}

func (me *MarkdownEditor) renderer() *mdRenderer {
    return &mdRenderer{
        blocks:   me.blocks,
        width:    me.Bounds().Width(),
        wrapText: true,
        cs:       me.ColorScheme(),
    }
}
```

**Cursor rendering in formatted mode:** Since the cursor lives in source coordinates (`cursorRow`, `cursorCol`), we need to map it to a screen position for the block cursor. In Phase 1, the simplest approach: compute which rendered line corresponds to the source cursor line by walking blocks and counting rendered lines per block. The cursor is drawn at that screen position.

**Implementation for drawCursor:**

```go
func (me *MarkdownEditor) drawCursor(buf *DrawBuffer, cs *theme.ColorScheme) {
    if !me.HasState(SfSelected) || me.HasSelection() {
        return
    }
    // Walk blocks to find the rendered line/col for (cursorRow, cursorCol)
    screenY, screenX := me.sourceToScreen(me.Memo.cursorRow, me.Memo.cursorCol)
    if screenY >= 0 && screenY < me.Bounds().Height() && screenX >= 0 && screenX < me.Bounds().Width() {
        ch := ' '
        if me.Memo.cursorRow < len(me.Memo.lines) && me.Memo.cursorCol < len(me.Memo.lines[me.Memo.cursorRow]) {
            ch = me.Memo.lines[me.Memo.cursorRow][me.Memo.cursorCol]
        }
        buf.WriteChar(screenX, screenY, ch, cs.MemoSelected)
    }
}
```

**The `sourceToScreen` method** — maps a source (row, col) to a screen (x, y) in the rendered output. It walks blocks, tracking current rendered line, and for the block containing the source line, computes the column within that block's rendered content. In Phase 1, a simple implementation: walk blocks, count lines per block, and when we reach the block containing the source line, map the source column through the block's rendered layout. For paragraph/header blocks, the screen column equals the source column (plus indent). For list items, add marker width. For blockquotes, add indent. For code blocks, map directly.

**HandleEvent:**

```go
func (me *MarkdownEditor) HandleEvent(event *Event) {
    if me.showSource {
        me.Editor.HandleEvent(event)
        return
    }

    if event.What == EvBroadcast && event.Command == CmIndicatorUpdate {
        me.Editor.HandleEvent(event)
        return
    }

    me.Editor.HandleEvent(event)

    if event.IsCleared() {
        me.reparse()
    }
}
```

**Run tests:** `go test ./tv/ -run MarkdownEditor -v`

**Commit:** `git commit -m "feat: add MarkdownEditor Draw with formatted rendering"`

---

### Task 3: Integration Checkpoint — Core Editing Loop

**Purpose:** Verify that Tasks 1-2 produce a working edit→parse→render loop.

**Requirements (for test writer):**
- Typing `# Hello` into a MarkdownEditor renders a heading (H1 style applied)
- Typing `**bold**` renders text with bold style
- The cursor moves through source coordinates and is visible on screen
- Scrolling works: typing enough lines to overflow the viewport, the scrollbar adjusts
- `SetText` followed by a Draw renders the expected formatted output
- Empty document renders without crashing

**Components to wire up:** MarkdownEditor, no mocks

**Run:** `go test ./tv/ -run TestMarkdownEditor -v`

**Commit:** `git commit -m "test: add integration tests for MarkdownEditor core loop"`

---

### Task 4: End-to-End Test — Phase 1

**Purpose:** Verify the MarkdownEditor works in a real running application.

**Requirements (for test writer):**
- Launch demo app with a MarkdownEditor window
- Type markdown text and verify it renders formatted on screen
- Verify cursor is visible and moves with arrow keys
- Verify scroll works with content that overflows the viewport

**Demo app update:** Add an EditWindow-like window containing a MarkdownEditor to `e2e/testapp/basic/main.go`. Wire it to a menu item or F-key.

**Run:** `go test ./e2e/ -run TestMarkdownEditor -timeout 180s`

**Commit:** `git commit -m "test: add e2e tests for MarkdownEditor Phase 1"`

---

## Phase 2: Syntax Reveal

**Capability:** Moving the cursor into a markdown construct reveals its syntax markers. Moving out hides them. Block-level reveal for headings, blockquotes, lists, code fences. Inline-level reveal for bold, italic, code, strikethrough. Text shifts to accommodate revealed markers.

### Task 5: Reveal mapper — block-level

**Files:**
- Create: `tv/markdown_editor_reveal.go`
- Modify: `tv/markdown_editor_draw.go` (integrate reveal spans into Draw)
- Modify: `tv/markdown_editor.go` (call reveal mapper from reparse)

**Requirements:**
- `revealSpan` struct with fields: `startRow, startCol int` (source position of first marker), `endRow, endCol int` (source position after last marker), `markerOpen string` (marker text at start), `markerClose string` (marker text at end), `kind revealKind` (block or inline)
- `revealKind` type: `revealBlock` or `revealInline`
- `buildRevealSpans(blocks []mdBlock, source [][]rune, cursorRow, cursorCol int) []revealSpan` produces reveal spans by walking blocks and checking cursor position
- Block-level reveal triggers when `cursorRow` falls within a block's source line range:
  - **Heading (#):** markerOpen = strings.Repeat("#", level) + " ", markerClose = ""
  - **Blockquote (>):** for each child block line, markerOpen = "> ", markerClose = ""
  - **Bullet list (-):** markerOpen = "- ", markerClose = ""
  - **Numbered list (1.):** markerOpen = "1. " (with correct number), markerClose = ""
  - **Checklist (- [ ]):** markerOpen = "- [ ] " or "- [x] ", markerClose = ""
  - **Code fence (```):** markerOpen = "```" + language on first line, markerClose = "```" on last line
  - **Horizontal rule (---):** markerOpen = "---" or "***", markerClose = ""
  - **Table:** markerOpen renders pipe and dash syntax for the row containing the cursor
- Block reveal scope is the ENTIRE block, not just the line with the cursor
- For blocks that span multiple source lines (blockquote children, list items with nested content, code blocks), the reveal spans cover all lines of the block
- Source positions for block markers: the marker text is "virtual" — it occupies column 0 of the rendered line (before the content), with the marker's source position anchored at (blockStartRow, 0)
- `inRevealSpan(spans []revealSpan, row, col int) *revealSpan` returns the span containing (row, col) or nil
- Empty source or empty blocks: `buildRevealSpans` returns nil
- The reveal mapper is called from `reparse()` after parsing, and the spans are stored in `MarkdownEditor.revealSpans`
- The reveal mapper does NOT mutate source; it only annotates
- **Escape sequences:** Backslash-escaped markdown characters (`\*`, `\#`, `\-`, etc.) must NOT generate reveal spans. Goldmark's AST already handles this — escaped characters are parsed as literal text, not as emphasis/heading/list markers. The reveal mapper inherits this naturally since it only produces spans for blocks/inline runs that goldmark identifies as markdown constructs.

**Implementation for building block-level reveal spans:**

```go
type revealKind int

const (
    revealBlock revealKind = iota
    revealInline
)

type revealSpan struct {
    startRow, startCol int
    endRow, endCol     int
    markerOpen         string
    markerClose        string
    kind               revealKind
}

func (me *MarkdownEditor) buildRevealSpans() {
    me.revealSpans = nil
    if len(me.blocks) == 0 {
        return
    }
    source := me.Editor.Memo.lines
    curRow, curCol := me.Memo.cursorRow, me.Memo.cursorCol

    // Walk blocks tracking current source row. For each block, check if
    // cursor is within its source line range. If so, generate reveal spans.
    row := 0
    me.revealSpans = me.collectRevealSpans(me.blocks, source, curRow, curCol, 0, &row)
}

// collectRevealSpans walks blocks recursively, tracking source row position
// and generating revealSpan records for blocks containing the cursor.
func (me *MarkdownEditor) collectRevealSpans(blocks []mdBlock, source [][]rune, curRow, curCol, depth int, row *int) []revealSpan {
    var spans []revealSpan
    for _, b := range blocks {
        blockStartRow := *row

        // Determine how many source lines this block occupies.
        // Paragraphs/headers: 1 line. Code blocks: len(code)+1 lines (fences).
        // Lists: sum of item lines + nested children. Blockquotes: sum of children.
        blockEndRow := blockStartRow + blockSourceLineCount(b, source, blockStartRow)

        if curRow >= blockStartRow && curRow < blockEndRow {
            spans = append(spans, blockRevealSpan(b, source, blockStartRow, depth)...)
        }

        // Walk nested children (blockquote children, list item children)
        *row = blockStartRow
        for _, item := range b.items {
            // item runs occupy one source line per item
            *row++ // marker line
            if len(item.children) > 0 {
                spans = append(spans, me.collectRevealSpans(item.children, source, curRow, curCol, depth+1, row)...)
            }
        }
        for _, child := range b.children {
            spans = append(spans, me.collectRevealSpans([]mdBlock{child}, source, curRow, curCol, depth+1, row)...)
        }

        *row = blockEndRow
    }
    return spans
}

// blockSourceLineCount returns the number of source lines a block occupies.
func blockSourceLineCount(b mdBlock, source [][]rune, startRow int) int {
    switch b.kind {
    case blockParagraph, blockHeader:
        return 1
    case blockCodeBlock:
        if len(b.code) == 0 {
            return 2 // opening and closing fences
        }
        return len(b.code) + 2
    case blockHRule:
        return 1
    case blockBulletList, blockNumberList, blockCheckList:
        count := 0
        for _, item := range b.items {
            count++ // marker line
            count += len(item.children) // nested blocks
        }
        return count
    case blockBlockquote:
        count := 0
        for _, child := range b.children {
            count += blockSourceLineCount(child, source, startRow+count)
        }
        return count
    case blockTable:
        // header row + separator row + data rows
        return len(b.rows) + 2
    default:
        return 1
    }
}

// blockRevealSpan generates reveal spans for a block when cursor is inside it.
// Reads actual marker characters from source lines rather than hardcoding.
func blockRevealSpan(b mdBlock, source [][]rune, blockRow, depth int) []revealSpan {
    indent := strings.Repeat("  ", depth)
    switch b.kind {
    case blockHeader:
        return []revealSpan{{
            startRow: blockRow, startCol: 0,
            endRow: blockRow + 1, endCol: 0,
            markerOpen: strings.Repeat("#", b.level) + " ",
            markerClose: "",
            kind:       revealBlock,
        }}
    case blockBulletList, blockNumberList, blockCheckList:
        var spans []revealSpan
        row := blockRow
        for _, item := range b.items {
            if row < len(source) {
                line := string(source[row])
                marker, _, _ := detectListMarker(line)
                if marker != "" {
                    spans = append(spans, revealSpan{
                        startRow: row, startCol: 0,
                        endRow: row + 1, endCol: 0,
                        markerOpen: indent + marker,
                        markerClose: "",
                        kind:       revealBlock,
                    })
                }
            }
            row++
        }
        return spans
    case blockBlockquote:
        var spans []revealSpan
        row := blockRow
        for _, child := range b.children {
            childCount := blockSourceLineCount(child, source, row)
            for i := 0; i < childCount; i++ {
                spans = append(spans, revealSpan{
                    startRow: row + i, startCol: 0,
                    endRow: row + i + 1, endCol: 0,
                    markerOpen: indent + "> ",
                    markerClose: "",
                    kind:       revealBlock,
                })
            }
            row += childCount
        }
        return spans
    case blockCodeBlock:
        return []revealSpan{{
            startRow: blockRow, startCol: 0,
            endRow: blockRow + 1, endCol: 0,
            markerOpen: "```" + b.language,
            markerClose: "",
            kind:       revealBlock,
        }, {
            startRow: blockRow + blockSourceLineCount(b, source, blockRow) - 1, startCol: 0,
            endRow: blockRow + blockSourceLineCount(b, source, blockRow), endCol: 0,
            markerOpen: "```",
            markerClose: "",
            kind:       revealBlock,
        }}
    case blockTable:
        // Reveal pipe-and-dash table syntax for the row containing the cursor.
        // Build a marker showing the column structure from headers.
        if len(b.headers) > 0 {
            var parts []string
            for _, h := range b.headers {
                parts = append(parts, strings.Repeat("-", len(h)+2))
            }
            return []revealSpan{{
                startRow: blockRow, startCol: 0,
                endRow: blockRow + 1, endCol: 0,
                markerOpen: "|" + strings.Join(parts, "|") + "|",
                markerClose: "",
                kind:       revealBlock,
            }}
        }
    case blockHRule:
        return []revealSpan{{
            startRow: blockRow, startCol: 0,
            endRow: blockRow + 1, endCol: 0,
            markerOpen: "---",
            markerClose: "",
            kind:       revealBlock,
        }}
    }
    return nil
}
```

**Draw integration:** Modify `markdown_editor_draw.go` to consult `revealSpans` during rendering.

**Important — `detectListMarker` placement:** `blockRevealSpan` calls `detectListMarker(line)` to read the actual marker from source (so `* `, `- `, `+ `, `1. `, `- [ ] ` all work). This helper and its dependencies (`isNumberedList`, `isChecklist`, `isDigit`, `incrementListNumber`, `uncheckedVersion`) must be defined in this task (Phase 2, `markdown_editor_reveal.go`). Task 10 (Phase 3) reuses them rather than redefining.

**The approach for rendering with reveal:** Instead of trying to inject markers into the existing block renderer (which would require modifying all render functions), we add a pass AFTER the formatted render that overlays markers:

1. Draw formatted blocks as before (no reveal)
2. For each revealed span, determine its screen position and draw the marker text over the formatted content, shifting content right as needed
3. Since the user accepted text-shift behavior, we write markers at the span's screen position, and content to the right shifts

Simpler approach that matches the spec better: during the Draw pass, for each visible rendered line, check whether any reveal spans apply to that line. If a block-level reveal is active for the block on this line, render the marker text at the beginning of the line (at the appropriate indent), then render the formatted content after the marker.

For Phase 2, we modify the Draw to:
1. Before rendering each block's line, check if a block-reveal span covers this line
2. If yes: first draw the marker text (e.g., "# ") using a dimmed style, then draw the content shifted right by the marker width
3. If no: draw the formatted content as normal

This is most naturally done by adding a new render method `renderBlockLineWithReveal` that wraps the existing `renderBlockLine` but prepends markers when appropriate.

**Theme note:** Revealed syntax markers use `MarkdownBlockquote` style (already defined in all themes, dimmed gray) — no new theme style needed. If visual distinction from blockquote content is insufficient, add a `MarkdownSyntaxMarker` style to each theme file.

**Run tests:** `go test ./tv/ -run MarkdownEditor -v`

**Commit:** `git commit -m "feat: add block-level syntax reveal to MarkdownEditor"`

---

### Task 6: Reveal mapper — inline-level

**Files:**
- Modify: `tv/markdown_editor_reveal.go` (add inline reveal logic)

**Requirements:**
- Inline-level reveal triggers when `(cursorRow, cursorCol)` falls inside or directly adjacent to an inline span in source
- Inline spans are extracted from goldmark's AST during parsing (the positions of `**`, `*`, `` ` ``, `~~` markers around text)
- The existing `parseMarkdown` function doesn't preserve source positions for inline markers — we need to extend it or add a second pass that scans source for inline marker positions
- **Alternative approach:** Since we have the source as `[][]rune` and the parsed blocks with `mdRun` styles, we can scan the source text for markdown syntax patterns to find marker positions, then match them to the parse tree's inline runs
- Inline reveal spans:
  - **Bold (`**text**`):** markerOpen = "**", markerClose = "**", spanning from the opening `**` to the closing `**`
  - **Italic (`*text*`):** markerOpen = "*", markerClose = "*"
  - **Italic (`_text_`):** markerOpen = "_", markerClose = "_"
  - **Code (`` `text` ``):** markerOpen = "`", markerClose = "`"
  - **Strikethrough (`~~text~~`):** markerOpen = "~~", markerClose = "~~"
- Inline reveal only applies to the specific span containing the cursor, not all spans in the block
- "Adjacent" means the cursor is on the marker character itself, or immediately before/after the marker
- Inline reveal must NOT cause a full document re-layout — text shifts locally within the line
- If cursor is in a block that has block-level reveal active, inline reveal within that block still works independently
- Implementation approach for scanning: walk source lines, find markdown syntax patterns using a state machine, match found positions with the parsed block structure
- Store inline marker positions during the parse pass by extending `mdBlock` with source position metadata, OR compute them in a separate scan of source for each block

**Simpler implementation (per spec):** Instead of modifying the goldmark parse to preserve inline marker positions, add a `scanInlineMarkers(source [][]rune, blocks []mdBlock) []revealSpan` function that walks each block's source lines and finds inline syntax markers. For each block:

1. Get the source text for the block's lines
2. Scan for `**`, `*`, `` ` ``, `~~` patterns
3. Match found markers to the block's `mdRun` inline styles (a bold run between positions X and Y means the `**` markers are at X and Y)
4. Generate revealSpan records for each inline construct

This is a pragmatic approach — it uses the parsed block structure to know WHERE inline formatting exists, then scans source to find the marker positions.

**Draw integration:** Inline markers are drawn in the inline content rendering, shifting surrounding text right within the line. When the renderer encounters a position that is the start or end of an inline reveal span, it writes the marker characters.

**Run tests:** `go test ./tv/ -run MarkdownEditor -v`

**Commit:** `git commit -m "feat: add inline-level syntax reveal to MarkdownEditor"`

---

### Task 7: Integration Checkpoint — Reveal Behavior

**Purpose:** Verify block and inline reveal work together correctly.

**Requirements (for test writer):**
- Cursor entering a heading line reveals `# ` prefix; leaving hides it
- Cursor entering a bold span reveals `**`; leaving hides it
- Block and inline reveal can be active simultaneously (cursor in bold inside a heading)
- Only the inline span containing the cursor reveals its markers, not other spans in the same block
- Reveal does not mutate source text
- Cursor position is stable during reveal/hide transitions

**Components to wire up:** MarkdownEditor with multi-block source text

**Run:** `go test ./tv/ -run TestMarkdownEditor -v`

**Commit:** `git commit -m "test: add integration tests for reveal behavior"`

---

### Task 8: End-to-End Test — Phase 2

**Purpose:** Verify reveal behavior in a running application.

**Requirements (for test writer):**
- Launch demo app with MarkdownEditor containing headings, bold, italic, lists, blockquotes
- Navigate cursor through different constructs and verify markers appear/disappear on screen
- Verify block-level reveal: entering a heading shows `#`, leaving hides it
- Verify inline reveal: entering bold shows `**`, leaving hides it

**Run:** `go test ./e2e/ -run TestMarkdownEditor -timeout 180s`

**Commit:** `git commit -m "test: add e2e tests for reveal behavior"`

---

## Phase 3: Interactive Features

**Capability:** Smart list continuation, auto-format on typing, keyboard shortcuts (Ctrl+B/I/K/T), link dialog, paste handling, source toggle.

### Task 9: Auto-format detection

**Files:**
- Create: `tv/markdown_editor_autoformat.go`
- Modify: `tv/markdown_editor.go` (call auto-format check after reparse)

**Requirements:**
- Auto-format does NOT mutate source — it only triggers re-render by calling `reparse()`. The source already contains the markdown syntax the user typed; the re-parse picks it up and the renderer formats it.
- Auto-format triggers are detected by checking what the user just typed (the last character(s) inserted):
  - `# ` typed at column 0 of a line → reparse already handles this
  - `**` typed as the closing of a bold construct (i.e., `**text**` was just completed)
  - `*` typed as the closing of italic
  - `` ` `` typed as the closing of inline code
  - `- ` typed at column 0 of a line
  - `> ` typed at column 0 of a line
  - `1. ` typed at column 0 of a line
- Since we already re-parse on every edit, auto-format is largely "free" — the re-parse detects the new structure. The main addition is list continuation, which DOES mutate source.

**Implementation note:** Most "auto-format" behavior is already handled by the re-parse-on-every-edit loop. When the user types `**text**`, the source contains `**text**`, and goldmark parses it as bold text. The renderer shows it as bold. When the cursor moves into the span, the reveal mapper shows the `**` markers. No additional auto-format code is needed for the basic cases in the spec table.

**Coverage of spec auto-format triggers:**

| Trigger | How it's covered |
|---|---|
| `# ` at line start | Reparse detects heading block. Rendered as H1-H6. |
| `**text**` after closing `**` | Reparse detects bold inline run. Rendered bold. |
| `*text*` after closing `*` | Reparse detects italic run. Rendered italic. |
| `` `text` `` after closing `` ` `` | Reparse detects code run. Rendered as code. |
| `- ` at line start | Reparse detects bullet list. Rendered with bullet. |
| `> ` at line start | Reparse detects blockquote. Rendered with indent bar. |
| ` ``` ` + Enter | Reparse detects fenced code block. Rendered as code. |
| `1. ` at line start | Reparse detects numbered list. Rendered with number. |

All 8 triggers are handled by goldmark's existing parser — no special detection code needed. The re-parse fires on every edit (Task 2's HandleEvent), so the formatted view updates immediately after each keystroke that completes a construct.

The auto-format file primarily handles:
1. Smart list continuation (Task 10) — the one case where source IS mutated
2. Keyboard format shortcuts (Task 11) — Ctrl+B/I toggles that mutate source

**Run tests:** `go test ./tv/ -run MarkdownEditor -v`

**Commit:** `git commit -m "feat: add auto-format detection to MarkdownEditor"`

---

### Task 10: Smart list continuation

**Files:**
- Modify: `tv/markdown_editor_autoformat.go` (add list continuation)

**Requirements:**
- When the user presses Enter at the end of a non-empty list item, insert a new line with the same marker:
  - `- item` + Enter → new line with `- ` prefix
  - `1. item` + Enter → new line with `2. ` prefix (incremented number)
  - `- [ ] item` + Enter → new line with `- [ ] ` prefix
  - `- [x] item` + Enter → new line with `- [ ] ` prefix (new items default unchecked)
- Detection: after Memo processes Enter, check if the PREVIOUS line (cursorRow - 1) is a list item by scanning its text for the marker pattern
- Insert the marker text at the start of the new line
- When the user presses Enter on an empty list item (line is just the marker, e.g., "- " or "1. "), delete the marker and replace with a blank line (exit the list)
- Tab at a list item: add `  ` prefix before the marker (`- ` → `  - `)
- Shift-Tab at an indented list item: remove `  ` prefix (`  - ` → `- `). If no indent, do nothing.
- List continuation works for bullet lists, numbered lists, and checklists
- Nested lists: Tab on a list item at depth 0 indents it by adding `  ` prefix; Shift-Tab outdents

**Implementation:**

```go
// handleListIndent handles Tab/Shift-Tab for list indent/outdent.
// Called BEFORE Editor processes the event so we can intercept Tab.
func (me *MarkdownEditor) handleListIndent(event *Event) bool {
    if event.What != EvKeyboard || event.Key == nil {
        return false
    }
    k := event.Key
    if k.Key == tcell.KeyTab && k.Modifiers == 0 {
        return me.listIndent()
    }
    if k.Key == tcell.KeyTab && k.Modifiers&tcell.ModShift != 0 {
        return me.listOutdent()
    }
    return false
}

// listEnterContinuation handles Enter for smart list continuation.
// Called AFTER Editor has processed Enter (new line created, cursor moved).
// Returns true if source was mutated (caller must reparse).
func (me *MarkdownEditor) listEnterContinuation() bool {
    lines := me.Memo.lines
    row := me.Memo.cursorRow
    if row == 0 {
        return false
    }
    prevLine := string(lines[row-1])
    curLine := string(lines[row])

    marker, _, isListItem := detectListMarker(prevLine)
    if !isListItem {
        return false
    }

    // If current line is empty marker, exit list
    curMarker, _, curIsListItem := detectListMarker(curLine)
    if curIsListItem && strings.TrimSpace(curLine) == strings.TrimSpace(curMarker) {
        lines[row] = []rune{}
        me.Memo.cursorCol = 0
        return true
    }

    // Insert marker on current line (the new line created by Enter)
    if strings.TrimSpace(curLine) == "" {
        newMarker := marker
        if isNumberedList(marker) {
            newMarker = incrementListNumber(marker)
        }
        if isChecklist(marker) {
            newMarker = uncheckedVersion(marker)
        }
        lines[row] = []rune(newMarker)
        me.Memo.cursorCol = len([]rune(newMarker))
        return true
    }

    return false
}

**Helpers:**

```go
// detectListMarker checks if a line starts with a list marker.
// Returns (marker, contentAfter, isListItem).
func detectListMarker(line string) (string, string, bool) {
    // Check for "- [ ] " or "- [x] "
    if strings.HasPrefix(line, "- [ ] ") {
        return "- [ ] ", line[6:], true
    }
    if strings.HasPrefix(line, "- [x] ") {
        return "- [x] ", line[6:], true
    }
    // Check for "- "
    if strings.HasPrefix(line, "- ") {
        return "- ", line[2:], true
    }
    // Check for "* "
    if strings.HasPrefix(line, "* ") {
        return "* ", line[2:], true
    }
    // Check for "N. " (numbered)
    if len(line) >= 3 && isDigit(line[0]) && strings.HasPrefix(line[1:], ". ") {
        end := strings.Index(line, ". ")
        return line[:end+2], line[end+2:], true
    }
    // Check for indented variants
    trimmed := line
    indent := 0
    for strings.HasPrefix(trimmed, "  ") {
        indent += 2
        trimmed = trimmed[2:]
    }
    if m, rest, ok := detectListMarker(trimmed); ok {
        return line[:indent] + m, rest, true
    }
    return "", line, false
}
```

**Run tests:** `go test ./tv/ -run MarkdownEditor -v`

**Commit:** `git commit -m "feat: add smart list continuation to MarkdownEditor"`

---

### Task 11: Keyboard shortcuts (Ctrl+B, Ctrl+I, Ctrl+K, Ctrl+T)

**Files:**
- Modify: `tv/markdown_editor_autoformat.go` (add shortcut handling)

**Requirements:**
- Ctrl+B: toggle bold on selection, or insert `****` with cursor between them
  - With selection: wrap selection in `**`. If already wrapped in `**`, remove the markers.
  - Without selection: insert `****` and place cursor between the inner `**` (position = original + 2)
  - After action, call `reparse()`
- Ctrl+I: toggle italic, same pattern with `*`
  - With selection: wrap in `*`. If already wrapped, remove.
  - Without selection: insert `**` with cursor between
- Ctrl+K: open link dialog
  - With selection: open dialog with selection text as the link text, empty URL. On OK: insert `[text](url)` replacing selection. On Cancel: no change.
  - Without selection, cursor on a link: open dialog with current link text and URL. On OK: update the link in source. On Cancel: no change. On Remove: delete the link syntax, keeping the text.
  - Without selection, not on a link: open dialog with empty fields. On OK: insert `[text](url)` at cursor position.
- **Enter while cursor is on a link** also opens the link dialog (same as Ctrl+K with cursor on link).
- **Status line hints:** When the cursor position falls within a link in the source, broadcast an indicator update with a hint like "Enter — edit link". This uses the existing `broadcastIndicator()` mechanism. The hint is cleared when the cursor leaves the link.
- The dialog uses two sequential `InputBox` calls (text, then URL) followed by a `MessageBox` for the confirm step OR a custom `Dialog` with two input fields and OK/Cancel/Remove buttons. The sequential InputBox approach is simpler and matches Turbo Vision patterns.
- Ctrl+T: toggle `showSource`
  - Flips `me.showSource`
  - When toggling back to formatted mode, call `reparse()`
  - Status line hint update via `broadcastIndicator()`
- Shortcuts are handled in `MarkdownEditor.HandleEvent` BEFORE delegating to `Editor.HandleEvent`

**Implementation for toggle format:**

```go
func (me *MarkdownEditor) toggleFormat(marker string) {
    if me.HasSelection() {
        sel := me.selectedText()
        // Check if already wrapped
        if strings.HasPrefix(sel, marker) && strings.HasSuffix(sel, marker) {
            // Remove markers
            unwrapped := sel[len(marker):len(sel)-len(marker)]
            me.deleteSelection()
            me.insertText(unwrapped)
        } else {
            // Wrap
            me.deleteSelection()
            wrapped := marker + sel + marker
            me.insertText(wrapped)
        }
    } else {
        // Insert empty markers
        me.insertText(marker + marker)
        // Move cursor between markers
        me.Memo.cursorCol -= len([]rune(marker))
    }
    me.reparse()
}
```

**Run tests:** `go test ./tv/ -run MarkdownEditor -v`

**Commit:** `git commit -m "feat: add keyboard shortcuts to MarkdownEditor"`

---

### Task 12: Link dialog

**Files:**
- Modify: `tv/markdown_editor_autoformat.go` (add link dialog)

**Requirements:**
- `openLinkDialog()` opens a modal dialog for link editing
- Uses the existing `tv.InputBox` pattern: two sequential InputBox calls (one for text, one for URL)
- When cursor is on a link in the source: pre-fill text and URL from the link's mdRun data
- When there's a selection: pre-fill text from selection, URL empty
- When neither: both fields empty
- OK: replace/create the link in source. Use `[text](url)` markdown syntax.
- Cancel: no change
- Additional "Remove" option: strips link syntax, keeps the text
- After the dialog, call `reparse()`

**Implementation:**

```go
// linkSpan holds the source positions and content of a markdown link.
type linkSpan struct {
    row                   int    // source row
    fullStart, fullEnd    int    // positions of entire [text](url) construct
    textStart, textEnd    int    // positions of link text (between [ and ])
    urlStart, urlEnd      int    // positions of URL (between ( and ))
    text, url             string // extracted content
}

// findLinkAt scans the source at (row, col) for a markdown link pattern
// [text](url) and returns a linkSpan if the position falls within the link text.
// This is a source-level scan — simpler and more robust than walking parsed blocks.
func (me *MarkdownEditor) findLinkAt(row, col int) *linkSpan {
    if row < 0 || row >= len(me.Memo.lines) {
        return nil
    }
    line := string(me.Memo.lines[row])
    // Find all [text](url) patterns on this line
    for i := 0; i < len(line); i++ {
        if line[i] == '[' {
            closeBracket := strings.Index(line[i:], "](")
            if closeBracket < 0 {
                continue
            }
            closeBracket += i
            closeParen := strings.Index(line[closeBracket+2:], ")")
            if closeParen < 0 {
                continue
            }
            closeParen += closeBracket + 2
            ls := &linkSpan{
                row:       row,
                fullStart: i,
                fullEnd:   closeParen + 1,
                textStart: i + 1,
                textEnd:   closeBracket,
                urlStart:  closeBracket + 2,
                urlEnd:    closeParen,
                text:      line[i+1 : closeBracket],
                url:       line[closeBracket+2 : closeParen],
            }
            if col >= ls.textStart && col < ls.textEnd {
                return ls
            }
            i = closeParen
        }
    }
    return nil
}

func (me *MarkdownEditor) openLinkDialog() {
    desktop := me.findDesktop()
    if desktop == nil {
        return
    }

    linkText := ""
    linkURL := ""
    var existingLink *linkSpan

    // Check if cursor is on a link
    r, c := me.Memo.cursorRow, me.Memo.cursorCol
    existingLink = me.findLinkAt(r, c)
    if existingLink != nil {
        linkText = existingLink.text
        linkURL = existingLink.url
    } else if me.HasSelection() {
        linkText = me.selectedText()
    }

    text, ok := InputBox(desktop, "Link Text", "~T~ext:", linkText)
    if !ok {
        return
    }
    url, ok := InputBox(desktop, "Link URL", "~U~RL:", linkURL)
    if !ok {
        return
    }

    // If editing an existing link, replace it in source
    if existingLink != nil {
        if url == "" && text == existingLink.text {
            return // nothing changed
        }
        line := me.Memo.lines[existingLink.row]
        before := string(line[:existingLink.fullStart])
        after := string(line[existingLink.fullEnd:])
        var replacement string
        if url != "" {
            replacement = "[" + text + "](" + url + ")"
        } else {
            replacement = text // remove link, keep text
        }
        me.Memo.lines[existingLink.row] = []rune(before + replacement + after)
        me.Memo.cursorRow = existingLink.row
        me.Memo.cursorCol = existingLink.fullStart + len(replacement)
        me.reparse()
        return
    }

    // Creating a new link from selection or at cursor
    if me.HasSelection() {
        me.deleteSelection()
    }

    if url != "" {
        me.insertText("[" + text + "](" + url + ")")
    } else if text != "" {
        me.insertText(text) // plain text, no link
    }

    me.reparse()
}
```

**Run tests:** `go test ./tv/ -run MarkdownEditor -v`

**Commit:** `git commit -m "feat: add link dialog to MarkdownEditor"`

---

### Task 13: Undo/Redo with meaningful boundaries

**Files:**
- Modify: `tv/markdown_editor.go` (override saveSnapshot behavior)

**Requirements:**
- Editor's default `saveSnapshot()` fires before every edit key, producing character-level undo. MarkdownEditor overrides this for word-level undo.
- Snapshots fire at meaningful boundaries only:
  - Word completion: space or punctuation typed after alphanumeric characters
  - Enter key pressed
  - Format command applied (Ctrl+B, Ctrl+I, Ctrl+K)
  - Paste completed
  - Delete word (Ctrl+Backspace, Ctrl+Delete)
  - Delete line (Ctrl+Y)
- Consecutive single-character inserts coalesce into one undo unit. A "character insert streak" is broken by: arrow key movement, mouse click, Enter, or any non-rune key.
- Consecutive single-character backspaces/deletes similarly coalesce.
- The coalescing is implemented by tracking `lastEditCoalesced bool` — set true after a snapshot for a coalesced edit. On the next coalesceable edit, skip the snapshot (it was already saved before the streak started).
- Undo (Ctrl+Z) restores the snapshot from the beginning of the current edit streak, plus cursor position.
- The behavior matches the spec: "Typing a word is one undo unit; applying a format is one undo unit; pasting is one undo unit."
- Undo/Redo in source toggle mode delegates to Editor's default behavior.

**Implementation:**

```go
// In MarkdownEditor:
type MarkdownEditor struct {
    *Editor
    blocks        []mdBlock
    sourceCache   string
    showSource    bool
    revealSpans   []revealSpan
    lastEditKind  editKind
    streakSaved   bool // snapshot already saved for current streak
}

type editKind int

const (
    editNone editKind = iota
    editChar           // single rune insert
    editBackspace      // single char delete
    editOther          // Enter, paste, format, etc — always save
)

func (me *MarkdownEditor) HandleEvent(event *Event) {
    if me.showSource {
        me.Editor.HandleEvent(event)
        return
    }

    if event.What == EvBroadcast && event.Command == CmIndicatorUpdate {
        me.Editor.HandleEvent(event)
        return
    }

    // Determine edit kind before processing
    kind := me.classifyEvent(event)

    // Save snapshot at meaningful boundaries
    if kind == editOther {
        me.saveSnapshot()
        me.streakSaved = true
    } else if (kind == editChar || kind == editBackspace) && !me.streakSaved {
        me.saveSnapshot()
        me.streakSaved = true
    }
    // If streakSaved already true and this is a continuation, skip save

    me.Editor.HandleEvent(event)

    if event.IsCleared() {
        if kind == editNone {
            // Non-edit key consumed (arrow, click) — breaks the streak
            me.streakSaved = false
            me.lastEditKind = editNone
        } else {
            me.lastEditKind = kind
            me.reparse()
        }
    }
}

func (me *MarkdownEditor) classifyEvent(event *Event) editKind {
    if event.What != EvKeyboard || event.Key == nil {
        return editNone
    }
    k := event.Key
    switch k.Key {
    case tcell.KeyRune:
        return editChar
    case tcell.KeyBackspace, tcell.KeyBackspace2:
        if k.Modifiers&tcell.ModCtrl != 0 {
            return editOther
        }
        return editBackspace
    case tcell.KeyDelete:
        if k.Modifiers&tcell.ModCtrl != 0 {
            return editOther
        }
        return editBackspace
    case tcell.KeyEnter, tcell.KeyCtrlX, tcell.KeyCtrlV, tcell.KeyCtrlY:
        return editOther
    }
    return editNone
}
```

**Run tests:** `go test ./tv/ -run MarkdownEditor -v`

**Commit:** `git commit -m "feat: add word-level undo/redo to MarkdownEditor"`

---

### Task 14: Paste handling

**Files:**
- Create: `tv/markdown_editor_paste.go`
- Modify: `tv/markdown_editor.go` (wire paste handling in HandleEvent)

**Requirements:**
- Ctrl+V pastes clipboard content with markdown detection:
  1. If clipboard starts with markdown patterns (`#`, `**`, `- `, `> `, `` ` ``, `[`) — treat as markdown, insert verbatim
  2. Otherwise — treat as plain text, insert verbatim
- Ctrl+Shift+V forces plain text paste (no markdown interpretation)
- HTML clipboard detection: check if clipboard starts with `<` and contains HTML tags. If yes, convert to markdown:
  - `<h1>text</h1>` → `# text`
  - `<h2>text</h2>` → `## text`
  - `<strong>text</strong>` or `<b>text</b>` → `**text**`
  - `<em>text</em>` or `<i>text</i>` → `*text*`
  - `<code>text</code>` → `` `text` ``
  - `<ul><li>text</li></ul>` → `- text`
  - `<ol><li>text</li></ol>` → `1. text`
  - `<a href="url">text</a>` → `[text](url)`
- After paste, call `reparse()`

**Implementation:**

```go
func (me *MarkdownEditor) handlePaste(forcePlain bool) {
    if clipboard == "" {
        return
    }
    
    text := clipboard
    if !forcePlain {
        text = me.convertClipboard(clipboard)
    }
    
    if me.HasSelection() {
        me.deleteSelection()
    }
    me.insertText(text)
    me.reparse()
}

func (me *MarkdownEditor) convertClipboard(s string) string {
    // HTML detection
    if looksLikeHTML(s) {
        return htmlToMarkdown(s)
    }
    return s // markdown or plain text, insert as-is
}
```

**Run tests:** `go test ./tv/ -run MarkdownEditor -v`

**Commit:** `git commit -m "feat: add paste handling to MarkdownEditor"`

---

### Task 15: Consolidated HandleEvent reference

**Purpose:** Merge all HandleEvent intercepts from Tasks 2, 5, 11, 12, 13, 14 into a single dispatch order. This is a reference for the implementer — not a new file.

The final HandleEvent dispatch order after all Phase 3 tasks:

```go
func (me *MarkdownEditor) HandleEvent(event *Event) {
    // 1. Ctrl+T toggles source mode — MUST be before showSource guard
    //    so it can toggle back FROM source mode.
    if event.What == EvKeyboard && event.Key != nil && event.Key.Key == tcell.KeyCtrlT {
        me.showSource = !me.showSource
        if !me.showSource {
            me.reparse()
        }
        me.broadcastIndicator()
        event.Clear()
        return
    }

    // 2. If in source mode, delegate completely to Editor
    if me.showSource {
        me.Editor.HandleEvent(event)
        return
    }

    // 3. Forward broadcast events to Editor (for status line)
    if event.What == EvBroadcast && event.Command == CmIndicatorUpdate {
        me.Editor.HandleEvent(event)
        return
    }

    // 4. Keyboard shortcuts (before Editor consumes the keystroke)
    if event.What == EvKeyboard && event.Key != nil {
        k := event.Key
        // 4a. Ctrl+B, Ctrl+I, Ctrl+K toggle format / open link dialog
        if k.Key == tcell.KeyCtrlB {
            me.toggleFormat("**")
            event.Clear()
            return
        }
        if k.Key == tcell.KeyCtrlI {
            me.toggleFormat("*")
            event.Clear()
            return
        }
        if k.Key == tcell.KeyCtrlK {
            me.openLinkDialog()
            event.Clear()
            return
        }
        // 4b. Enter on a link opens link dialog (intercepted before Editor)
        if k.Key == tcell.KeyEnter && me.findLinkAt(me.Memo.cursorRow, me.Memo.cursorCol) != nil {
            me.openLinkDialog()
            event.Clear()
            return
        }
        // 4c. Ctrl+V paste handling — both plain and force-plain
        if k.Key == tcell.KeyCtrlV {
            forcePlain := k.Modifiers&tcell.ModShift != 0
            me.handlePaste(forcePlain)
            event.Clear()
            return
        }
    }

    // 5. Smart list indent/outdent (Tab/Shift-Tab, before Editor)
    //    Enter continuation is handled post-edit (step 9b) because it needs
    //    Editor to create the new line first.
    if me.handleListIndent(event) {
        event.Clear()
        return
    }

    // 6. Classify event for undo coalescing
    kind := me.classifyEvent(event)

    // 7. Save snapshot at meaningful boundaries
    if kind == editOther || ((kind == editChar || kind == editBackspace) && !me.streakSaved) {
        me.saveSnapshot()
        me.streakSaved = true
    }

    // 8. Delegate to Editor (Memo handles cursor, selection, typing, Enter)
    me.Editor.HandleEvent(event)

    // 9. Post-edit: reparse, list continuation, or reset streak
    if event.IsCleared() {
        if kind == editNone {
            // 9a. Non-edit key (arrow, click) — breaks the streak
            me.streakSaved = false
            me.lastEditKind = editNone
        } else {
            me.lastEditKind = kind
            me.reparse()
            // 9b. After Enter, check list continuation (new line now exists).
            //     This mutates source, so reparse again to pick up the new marker.
            if event.What == EvKeyboard && event.Key != nil && event.Key.Key == tcell.KeyEnter {
                if me.listEnterContinuation() {
                    me.reparse()
                }
            }
        }
    }
}
```

**Key design decisions in this dispatch order:**

- **Ctrl+T first (step 1):** Must precede the `showSource` guard so it can toggle back from source mode. When toggling to formatted mode, `reparse()` refreshes the block tree. `broadcastIndicator()` is inherited from `Editor` — it sends `CmIndicatorUpdate` to update the status line.
- **Enter-on-link before Editor (step 4b):** Intercepted before Memo consumes Enter, since we want to open a dialog instead of inserting a newline.
- **Ctrl+V both variants (step 4c):** Plain Ctrl+V (`Modifiers == 0`) does markdown-aware paste; Ctrl+Shift+V forces plain text.
- **Tab/Shift-Tab before Editor (step 5):** Intercepted before Memo processes them, since we're modifying list indentation.
- **Enter list continuation after Editor (step 9b):** Editor must process Enter first — it creates the new blank line and moves the cursor. Only then can `listEnterContinuation()` inspect the previous line's marker and insert a new marker on the current line.
- **Undo snapshots before Editor (step 7):** Saved before the edit so undo can restore to pre-edit state.

**Commit:** `git commit -m "refactor: consolidate MarkdownEditor HandleEvent dispatch"`

---

### Task 16: Integration Checkpoint — Interactive Features

**Purpose:** Verify list continuation, shortcuts, link dialog, and paste work together.

**Requirements (for test writer):**
- Typing `- item` then Enter creates a new `- ` line
- Typing Enter on an empty `- ` line exits the list
- Ctrl+B with selection wraps it in `**`; pressed again unwraps
- Ctrl+T toggles between source and formatted view
- Ctrl+K with selection opens link dialog; filling it in inserts `[text](url)`

**Components to wire up:** MarkdownEditor, no mocks

**Run:** `go test ./tv/ -run TestMarkdownEditor -v`

**Commit:** `git commit -m "test: add integration tests for interactive features"`

---

### Task 17: End-to-End Test — Phase 3

**Purpose:** Verify all interactive features in a running application.

**Requirements (for test writer):**
- Type markdown and see it render formatted in real time
- Navigate into a heading with cursor, see `# ` markers appear
- Navigate into bold, see `**` markers appear
- Ctrl+T toggles raw source view
- Ctrl+B wraps selection in `**`
- Enter in a list continues the list
- Enter on empty list item exits the list

**Run:** `go test ./e2e/ -run TestMarkdownEditor -timeout 180s`

**Commit:** `git commit -m "test: add e2e tests for Phase 3"`

---

### Task 18: Demo app integration

**Files:**
- Modify: `e2e/testapp/basic/main.go`

**Requirements:**
- Add a menu item (e.g., `File > New Markdown` or an F-key) that opens a window with a MarkdownEditor
- The MarkdownEditor is in a scrollable window (like EditWindow but simpler)
- The window has scrollbars bound to the MarkdownEditor
- The MarkdownEditor has a status line hint showing Ctrl+T for source toggle
- The demo app builds and runs without errors

**Implementation:**

```go
// In main.go, add to menu:
tv.NewMenuItem("~M~arkdown Editor", tv.CmUser+40, tv.KbCtrl('M')),

// In OnCommand:
if cmd == tv.CmUser+40 {
    w := tv.NewWindow(tv.NewRect(2, 1, 60, 20), "Markdown Editor", tv.WithWindowNumber(7))

    // Layout: editor fills window interior, scrollbars at right/bottom edges.
    // Window frame is 1 char on each side, so interior starts at (1,1).
    iw, ih := w.Bounds().Width()-2, w.Bounds().Height()-2
    editor := tv.NewMarkdownEditor(tv.NewRect(1, 1, iw-1, ih-1))
    editor.SetText("# Welcome\n\nType **markdown** here.\n\n- item one\n- item two")

    vScroll := tv.NewScrollBar(tv.NewRect(iw, 1, 1, ih-1), tv.SbVertical)
    hScroll := tv.NewScrollBar(tv.NewRect(1, ih, iw-1, 1), tv.SbHorizontal)
    editor.SetVScrollBar(vScroll)
    editor.SetHScrollBar(hScroll)

    w.Insert(editor)
    w.Insert(vScroll)
    w.Insert(hScroll)
    app.Desktop().Insert(w)
    return true
}
```

**Run:** `go run ./e2e/testapp/basic`

**Commit:** `git commit -m "feat: add MarkdownEditor to demo app"`

---
