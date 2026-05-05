# MarkdownViewer Widget — Design Specification

## Motivation

A read-only widget that renders Markdown content with TUI-appropriate formatting. The primary use case is an AI chat interface where the AI emits Markdown and the widget renders it — but the widget is general-purpose, available to any TurboView application.

## Architecture

### Pipeline

```
SetMarkdown(text string)
  → goldmark parses text into CommonMark AST
  → AST walker converts to []mdBlock (internal IR)
  → IR stored in widget

Draw(buf *DrawBuffer)
  → walk []mdBlock
  → word-wrap runs to current widget width
  → apply style composition (block bg + inline fg/attrs)
  → WriteChar per cell through viewport offset (deltaX, deltaY)
```

Parsing happens once on `SetMarkdown`. Drawing walks the IR and performs word wrapping at render time, so resizing the widget is cheap (re-wrap, no re-parse). The IR is designed for efficient append to support future streaming.

### External Dependencies

- `github.com/yuin/goldmark` — CommonMark parser with extensions for tables, checkboxes, definition lists, and strikethrough. Standard Go Markdown library.

No other external dependencies. Rendering, style composition, scrolling, and table layout are all internal code.

## Intermediate Representation

The IR follows the codebase convention of flat structs with a `kind` discriminator (same pattern as `Event.What`), not interface hierarchies. All IR types are unexported.

```go
type mdBlockKind int

const (
    blockParagraph  mdBlockKind = iota
    blockHeader
    blockCodeBlock
    blockBulletList
    blockNumberList
    blockBlockquote
    blockTable
    blockHRule
    blockDefList
    blockCheckList
)

type mdBlock struct {
    kind     mdBlockKind
    level    int         // header level (1-6); numbered list start value
    runs     []mdRun     // inline content for paragraph, header
    language string      // code block language tag (e.g. "go", "python")
    code     []string    // code block raw lines (preserved for copy, line numbers)
    items    []mdItem    // list items (bullet, number, check, definition)
    children []mdBlock   // nested blocks (blockquote contents, sub-lists)
    headers  [][]mdRun   // table column headers
    rows     [][][]mdRun // table data rows, each row is [][]mdRun (cells)
}

type mdRunStyle int

const (
    runNormal     mdRunStyle = iota
    runBold
    runItalic
    runBoldItalic
    runCode
    runLink
    runStrikethrough
)

type mdRun struct {
    text  string
    style mdRunStyle
    url   string // populated only for runLink
}

type mdItem struct {
    runs     []mdRun   // item's inline content
    children []mdBlock // nested blocks (sub-lists, paragraphs, code blocks)
    checked  *bool     // nil = regular item; true/false = checkbox state
    term     []mdRun   // definition list: the term (runs is the definition)
}
```

### Key Design Decisions

- **`mdBlock` is a flat struct** with unused fields per kind, not a type hierarchy. Matches `Event` pattern.
- **Code blocks store raw `[]string` lines**, not styled runs. Preserves original text for clipboard copy, line numbers, and future syntax highlighting.
- **`mdItem` unifies list items, check items, and definition items.** `checked *bool` distinguishes checkboxes (nil = regular). `term []mdRun` holds the definition term when non-nil.
- **Blockquotes contain `[]mdBlock` children** — they're recursive and can hold any block type including nested blockquotes.
- **List nesting is unbounded.** `mdItem.children` can contain sub-lists, which contain sub-items, etc. Rendering indentation is `depth * indentWidth`.
- **Table cells contain `[]mdRun`** — cells can have inline formatting (bold, code, links).
- **Links store the URL** in `mdRun.url` for future link-following features. Rendered as styled text (not clickable in v1).
- **Image references** render as `[IMG: alt text]` using `MarkdownCode` style.

## Widget API

```go
type MarkdownViewer struct {
    BaseView
    blocks     []mdBlock
    source     string      // original markdown text
    deltaX     int         // horizontal scroll offset
    deltaY     int         // vertical scroll offset
    wrapText   bool        // word-wrap non-code content (default true)
    vScrollBar *ScrollBar
    hScrollBar *ScrollBar
}

// Constructor
func NewMarkdownViewer(bounds Rect) *MarkdownViewer

// Content
func (mv *MarkdownViewer) SetMarkdown(text string)
func (mv *MarkdownViewer) Markdown() string

// Scrollbar binding
func (mv *MarkdownViewer) SetVScrollBar(sb *ScrollBar)
func (mv *MarkdownViewer) SetHScrollBar(sb *ScrollBar)

// Word wrap toggle
func (mv *MarkdownViewer) SetWrapText(wrap bool)
func (mv *MarkdownViewer) WrapText() bool
```

### Constructor Behavior

`NewMarkdownViewer` sets:
- `SfVisible` — visible by default
- `OfSelectable` — can receive focus
- `OfFirstClick` — focusing click is also processed
- `SetSelf(mv)` — required for BaseView click-to-focus
- `wrapText = true` — word wrapping on by default

Matches `NewMemo`, `NewListViewer`, `NewOutlineViewer` patterns.

Note: `SetGrowMode` is NOT set by the constructor — it's the caller's responsibility, same as with Window and other container-placed widgets. The demo app should set `GfGrowHiX | GfGrowHiY` when placing the viewer in a resizable window.

### SetMarkdown

Parses the input string via goldmark, walks the AST to build `[]mdBlock`, stores both the IR and the original source string. Resets `deltaX` and `deltaY` to 0. Syncs scrollbars.

### Scrollbar Binding

`SetVScrollBar` and `SetHScrollBar` follow the established pattern:
- Set `sb.OnChange` to update `deltaY`/`deltaX`
- `syncScrollBars()` updates range, page size, and value after any state change
- `SetVScrollBar(nil)` / `SetHScrollBar(nil)` unbinds

## Markdown Elements

### In Scope (v1)

| Element | Markdown Syntax | Rendering |
|---------|----------------|-----------|
| Paragraph | Plain text | Word-wrapped to widget width, blank line between paragraphs |
| Header H1-H6 | `#` through `######` | Full line in header style, blank line after |
| Bold | `**text**` | Inline, composed with block style |
| Italic | `*text*` | Inline, composed with block style |
| Bold italic | `***text***` | Inline, composed with block style |
| Inline code | `` `code` `` | Inline with own background |
| Code block | ```` ``` ```` with optional language | Distinct background, never word-wrapped, horizontal scroll |
| Bulleted list | `- item` or `* item` | `•` marker in marker style, item text word-wrapped |
| Numbered list | `1. item` | Number + `.` in marker style, item text word-wrapped |
| Checkbox list | `- [ ]` / `- [x]` | `☐`/`☑` in marker style, item text word-wrapped |
| Nested lists | Indented sub-items | Unbounded depth, indentation = `depth * 2` chars |
| Blockquote | `> text` | `▌` left bar + dimmer text, recursive (can contain any block) |
| Horizontal rule | `---` | Full-width `─` characters in rule style |
| Link | `[text](url)` | Text in link style, URL stored but not clickable |
| Table | GFM table syntax | Box-drawing borders, cell wrapping with min-width heuristic |
| Definition list | `term\n: definition` | Term in bold style, definition indented 4 chars |
| Image | `![alt](url)` | Rendered as `[IMG: alt text]` in code style |
| Strikethrough | `~~text~~` | Rendered with `AttrStrikeThrough` attribute |

### Out of Scope

Footnotes, inline HTML tags, deeply nested blockquote styling beyond recursive rendering.

## Rendering Rules

### Block Spacing

- One blank line between consecutive blocks (paragraphs, headers, code blocks, lists, etc.)
- No blank line before the first block in the document
- List items within a list: no blank lines between siblings
- Code blocks: no surrounding blank lines beyond the standard inter-block gap
- Nested blocks (inside blockquotes, list items): indented, same spacing rules apply recursively

### Word Wrapping

- **Paragraphs, list items, blockquotes, definition bodies, table cells**: word-wrapped to available width (widget width minus indentation) when `wrapText` is true
- **Code blocks**: never wrapped regardless of `wrapText` setting — horizontal scroll handles overflow
- **Tables**: cells word-wrap if all columns fit at minimum width (longest word or 8 chars floor); if the table is too wide even at minimum widths, render at minimums and allow horizontal scrolling
- **Toggle**: `W` key or `SetWrapText()` flips wrapping for non-code content

### Style Composition

Block styles provide the background color and base foreground. Inline styles override the foreground and add attributes. The renderer composes them:

- **Normal inline text**: uses block style directly
- **Bold/Italic/BoldItalic/Strikethrough**: takes foreground and attributes from the inline style field, background from the current block style
- **Inline code**: uses its own background (overrides block background) — this is the exception
- **Links**: own foreground + underline, block background

This means bold text inside a blockquote inherits the blockquote's background but uses `MarkdownBold`'s foreground and bold attribute.

### Table Layout

1. Compute minimum column width for each column: the length of the longest word in that column, with a floor of 8 characters
2. If sum of minimum widths + borders fits within widget width: distribute remaining space proportionally and word-wrap cell contents
3. If sum of minimum widths + borders exceeds widget width: render at minimum widths, allow horizontal scrolling
4. Table borders use box-drawing characters (`┌─┬─┐`, `│`, `├─┼─┤`, `└─┴─┘`) in `MarkdownTableBorder` style
5. Header row (if present) uses `MarkdownBold` style for header cell text

### List Rendering

- Bullet marker: `•` in `MarkdownListMarker` style
- Number marker: `1.`, `2.`, etc. in `MarkdownListMarker` style
- Checkbox: `☐` (unchecked) or `☑` (checked) in `MarkdownListMarker` style
- Indent per nesting level: 2 characters
- Item text starts after marker + space, word-wrapped to remaining width
- Continuation lines align with the start of text, not the marker

### Blockquote Rendering

- Left bar: `▌` character in `MarkdownBlockquote` style at the left margin
- Content indented 2 characters from the bar
- Nested blockquotes stack: `▌ ▌ text` for depth 2
- All content inside the blockquote uses `MarkdownBlockquote` as the block style

### Definition List Rendering

- Term: rendered on its own line in `MarkdownDefTerm` style
- Definition: indented 4 characters on the next line(s), word-wrapped
- Multiple definitions for one term: each indented, separated by blank lines

### Horizontal Rule

Full widget width of `─` characters in `MarkdownHRule` style.

## Keyboard and Mouse Handling

Keyboard events are only handled when the widget has focus (`SfSelected`). Mouse events call `BaseView.HandleEvent` first for click-to-focus.

| Input | Action |
|-------|--------|
| Up | Scroll up 1 line |
| Down | Scroll down 1 line |
| PgUp | Scroll up viewport height |
| PgDn | Scroll down viewport height |
| Home | Jump to top of document |
| End | Jump to bottom of document |
| Left | Scroll left 1 column (when horizontal content exceeds width) |
| Right | Scroll right 1 column |
| W | Toggle word wrap for non-code content |
| Mouse wheel up/down | Scroll ±3 lines |
| Mouse wheel left/right | Scroll ±3 columns |
| Button1 click | Focus only (no text selection) |

All handled events are consumed (cleared). Unrecognized keys pass through.

Home/End go to document top/bottom directly (no Ctrl modifier needed) because the widget has no cursor — unlike Memo where Home/End go to line start/end.

## Color Scheme

18 new fields added to `theme.ColorScheme`:

```go
// Markdown viewer styles
MarkdownNormal     tcell.Style // Body text
MarkdownH1         tcell.Style // Header level 1
MarkdownH2         tcell.Style // Header level 2
MarkdownH3         tcell.Style // Header level 3
MarkdownH4         tcell.Style // Header level 4
MarkdownH5         tcell.Style // Header level 5
MarkdownH6         tcell.Style // Header level 6
MarkdownBold       tcell.Style // Bold inline text
MarkdownItalic     tcell.Style // Italic inline text
MarkdownBoldItalic tcell.Style // Bold + italic inline text
MarkdownCode       tcell.Style // Inline code
MarkdownCodeBlock  tcell.Style // Code block text
MarkdownBlockquote tcell.Style // Blockquote text
MarkdownLink       tcell.Style // Link text
MarkdownHRule      tcell.Style // Horizontal rule
MarkdownListMarker tcell.Style // Bullet, number, checkbox markers
MarkdownTableBorder tcell.Style // Table box-drawing characters
MarkdownDefTerm    tcell.Style // Definition list term
```

### BorlandBlue Defaults

| Field | Foreground | Background | Attributes |
|-------|-----------|------------|------------|
| MarkdownNormal | LightGray | Blue | — |
| MarkdownH1 | White | Blue | Bold, Underline |
| MarkdownH2 | Yellow | Blue | Bold |
| MarkdownH3 | Cyan | Blue | Bold |
| MarkdownH4 | Cyan | Blue | — |
| MarkdownH5 | LightGray | Blue | Bold |
| MarkdownH6 | LightGray | Blue | — |
| MarkdownBold | White | Blue | Bold |
| MarkdownItalic | LightGray | Blue | Italic |
| MarkdownBoldItalic | White | Blue | Bold, Italic |
| MarkdownCode | Cyan | DarkBlue | — |
| MarkdownCodeBlock | Green | DarkBlue | — |
| MarkdownBlockquote | DarkGray | Blue | — |
| MarkdownLink | Green | Blue | Underline |
| MarkdownHRule | DarkGray | Blue | — |
| MarkdownListMarker | Yellow | Blue | — |
| MarkdownTableBorder | DarkGray | Blue | — |
| MarkdownDefTerm | White | Blue | Bold |

## Scrolling

- `deltaY`: vertical offset in rendered lines (after word wrapping). Range: 0 to totalRenderedLines - viewportHeight.
- `deltaX`: horizontal offset in columns. Only affects content that exceeds widget width (code blocks, wide tables when not wrapping). Range: 0 to maxLineWidth - viewportWidth.
- Vertical scrollbar: always bound when provided. Range tracks total rendered height.
- Horizontal scrollbar: bound when provided. Range tracks maximum content width.
- `syncScrollBars()` called after `SetMarkdown`, `SetWrapText`, resize, and any scroll operation.
- Scrollbar range formula follows the `count - 1 + pageSize` convention established by ListViewer, so the scrollbar thumb reaches the end when scrolled to the bottom.

## Demo App Integration

Add a MarkdownViewer in a new window to the demo app (`e2e/testapp/basic/main.go`) with sample Markdown content exercising all supported elements. The window name should be "Markdown" for e2e test reference.

## Testing Strategy

### Unit Tests — IR Parsing

Test that Markdown strings produce the expected `[]mdBlock` structures. Tests the goldmark walker independently of rendering.

Examples:
- `"**bold** text"` → one `blockParagraph` with two runs: `{runBold, "bold"}`, `{runNormal, " text"}`
- `"# Title"` → one `blockHeader` with level 1
- Fenced code block → `blockCodeBlock` with language and raw lines preserved
- Nested list → `blockBulletList` with items whose children contain sub-lists
- Table → `blockTable` with headers and rows
- Definition list → `blockDefList` with term and definition

### Unit Tests — Rendering

Test that `[]mdBlock` renders correctly to a DrawBuffer at specific dimensions. Tests word wrapping, style application, indentation, table layout.

Examples:
- Paragraph wrapping at width 20
- Code block not wrapping, extending beyond width
- Table fitting vs. horizontal scroll
- Nested list indentation at depth 3
- Blockquote with left bar
- Style composition: bold inside blockquote gets blockquote bg + bold fg

### E2E Tests

Build the demo app, drive with tmux, assert visible output:
- Headers visible with expected text
- Code block content visible
- List markers (•) visible
- Table borders (┌─┬─┐) visible
- PgDn scrolls to show later content
- W toggles wrapping (code block unaffected)

## Future Work (IR-Ready)

The IR is designed to support these features without structural changes:

- **Streaming append**: `AppendMarkdown(chunk)` parses new content and appends blocks. Auto-scroll to bottom unless user scrolled up.
- **Block-level keyboard navigation**: Tab/Shift-Tab to jump between code blocks, headers, and other significant elements. IR blocks are addressable entities.
- **Code block copy**: Hotkey to copy focused code block's raw text. `code []string` preserves original content.
- **Syntax highlighting**: Language tag in `mdBlock.language` enables future colorization via chroma or similar.
- **Search (Ctrl+F)**: IR preserves text content for search matching; rendered line positions enable highlight.
- **Section folding**: Header hierarchy in IR enables collapse/expand by level.
- **Code block line numbers**: Optional gutter; `code []string` preserves line structure.
- **Link following**: URLs stored in `mdRun.url`; future command could open in system browser.
