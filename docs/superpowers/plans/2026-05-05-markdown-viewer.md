# MarkdownViewer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a read-only MarkdownViewer widget that parses Markdown via goldmark and renders it with TUI-appropriate formatting, scrolling, and word wrapping.

**Architecture:** Parse-on-set pipeline: `SetMarkdown(text)` runs goldmark to build an internal IR (`[]mdBlock`), then `Draw()` walks the IR with word wrapping at render time. The IR is a flat struct with `kind` discriminator (matching the `Event.What` pattern). Style composition overlays inline styles (fg/attrs) on block styles (bg).

**Tech Stack:** Go, goldmark (CommonMark parser), tcell v2 (terminal), existing TurboView framework

**Spec:** `docs/superpowers/specs/2026-05-05-markdown-viewer-design.md`

---

## File Structure

| File | Purpose |
|------|---------|
| `tv/markdown_ir.go` | IR types (`mdBlock`, `mdRun`, `mdItem`, enums) and goldmark AST-to-IR walker |
| `tv/markdown_viewer.go` | `MarkdownViewer` widget: constructor, API, `Draw`, `HandleEvent`, scrolling, scrollbar sync |
| `tv/markdown_render.go` | Rendering helpers: word wrap, style composition, table layout, list/blockquote indent |
| `theme/scheme.go` | Add 18 `Markdown*` fields to `ColorScheme` |
| `theme/borland.go` | BorlandBlue defaults for all 18 Markdown styles |
| `e2e/testapp/basic/main.go` | Add Window 5 ("Markdown") with sample content |

Test files (created by test writers, not in plan):
- `tv/markdown_ir_test.go` — IR parsing tests
- `tv/markdown_render_test.go` — rendering/layout tests
- `tv/markdown_viewer_test.go` — widget behavior tests (scrolling, keyboard, scrollbar)
- `e2e/e2e_test.go` — e2e tests appended

---

## Phase 1: Core Rendering

**After this phase:** The demo app shows a "Markdown" window (Window 5) with rendered Markdown — headers in color, bold/italic text, inline code, code blocks with distinct background, paragraphs word-wrapped, horizontal rules. PgDn scrolls through the document. The `W` key toggles word wrapping.

### Task 1: Color Scheme — Add Markdown Style Fields

**Files:**
- Modify: `theme/scheme.go` — add 18 fields to `ColorScheme` struct
- Modify: `theme/borland.go` — add BorlandBlue defaults

**Requirements:**
- `ColorScheme` has 18 new fields: `MarkdownNormal`, `MarkdownH1` through `MarkdownH6`, `MarkdownBold`, `MarkdownItalic`, `MarkdownBoldItalic`, `MarkdownCode`, `MarkdownCodeBlock`, `MarkdownBlockquote`, `MarkdownLink`, `MarkdownHRule`, `MarkdownListMarker`, `MarkdownTableBorder`, `MarkdownDefTerm`
- All fields are `tcell.Style` type
- BorlandBlue assigns all 18 fields with the values from the spec's BorlandBlue Defaults table
- Existing tests still pass (no existing field positions broken — Go struct fields are named, not positional)

**Implementation:**

Add to `theme/scheme.go` after the `OutlineCollapsed` field:

```go
// Markdown viewer
MarkdownNormal      tcell.Style
MarkdownH1          tcell.Style
MarkdownH2          tcell.Style
MarkdownH3          tcell.Style
MarkdownH4          tcell.Style
MarkdownH5          tcell.Style
MarkdownH6          tcell.Style
MarkdownBold        tcell.Style
MarkdownItalic      tcell.Style
MarkdownBoldItalic  tcell.Style
MarkdownCode        tcell.Style
MarkdownCodeBlock   tcell.Style
MarkdownBlockquote  tcell.Style
MarkdownLink        tcell.Style
MarkdownHRule       tcell.Style
MarkdownListMarker  tcell.Style
MarkdownTableBorder tcell.Style
MarkdownDefTerm     tcell.Style
```

Add to `theme/borland.go` in the `BorlandBlue` initialization, after the Outline fields. Use the existing `s` helper for simple fg+bg, and add attributes where needed:

```go
MarkdownNormal:      s(tcell.ColorLightGray, tcell.ColorBlue),
MarkdownH1:          s(tcell.ColorWhite, tcell.ColorBlue).Bold(true).Underline(true),
MarkdownH2:          s(tcell.ColorYellow, tcell.ColorBlue).Bold(true),
MarkdownH3:          s(tcell.ColorDarkCyan, tcell.ColorBlue).Bold(true),
MarkdownH4:          s(tcell.ColorDarkCyan, tcell.ColorBlue),
MarkdownH5:          s(tcell.ColorLightGray, tcell.ColorBlue).Bold(true),
MarkdownH6:          s(tcell.ColorLightGray, tcell.ColorBlue),
MarkdownBold:        s(tcell.ColorWhite, tcell.ColorBlue).Bold(true),
MarkdownItalic:      s(tcell.ColorLightGray, tcell.ColorBlue).Italic(true),
MarkdownBoldItalic:  s(tcell.ColorWhite, tcell.ColorBlue).Bold(true).Italic(true),
MarkdownCode:        s(tcell.ColorDarkCyan, tcell.ColorNavy),
MarkdownCodeBlock:   s(tcell.ColorGreen, tcell.ColorNavy),
MarkdownBlockquote:  s(tcell.ColorDarkGray, tcell.ColorBlue),
MarkdownLink:        s(tcell.ColorGreen, tcell.ColorBlue).Underline(true),
MarkdownHRule:       s(tcell.ColorDarkGray, tcell.ColorBlue),
MarkdownListMarker:  s(tcell.ColorYellow, tcell.ColorBlue),
MarkdownTableBorder: s(tcell.ColorDarkGray, tcell.ColorBlue),
MarkdownDefTerm:     s(tcell.ColorWhite, tcell.ColorBlue).Bold(true),
```

**Run tests:** `go test ./theme/ -v`

**Commit:** `git commit -m "feat: add 18 Markdown style fields to ColorScheme"`

---

### Task 2: IR Types and Goldmark Parser

**Files:**
- Create: `tv/markdown_ir.go`
- Modify: `go.mod` — add goldmark dependency

**Requirements:**
- `mdBlockKind` enum with 10 values: `blockParagraph`, `blockHeader`, `blockCodeBlock`, `blockBulletList`, `blockNumberList`, `blockBlockquote`, `blockTable`, `blockHRule`, `blockDefList`, `blockCheckList`
- `mdRunStyle` enum with 7 values: `runNormal`, `runBold`, `runItalic`, `runBoldItalic`, `runCode`, `runLink`, `runStrikethrough`
- `mdBlock` struct with fields: `kind`, `level`, `runs`, `language`, `code`, `items`, `children`, `headers`, `rows`
- `mdRun` struct with fields: `text`, `style`, `url`
- `mdItem` struct with fields: `runs`, `children`, `checked`, `term`
- `parseMarkdown(src string) []mdBlock` function that:
  - Parses `src` via goldmark with GFM extensions (table, strikethrough, task list) and definition list extension
  - Walks the goldmark AST to produce `[]mdBlock`
  - Paragraphs: collects inline children into `[]mdRun` with appropriate styles
  - Headers: sets `level` (1-6) and collects inline runs
  - Code blocks: stores language tag and raw lines as `[]string` (no trailing empty line)
  - Bullet lists: each item gets `runs` from inline content, `children` from nested blocks
  - Ordered lists: same as bullet, plus `level` field stores start number
  - Blockquotes: `children` contains the recursive block contents
  - Tables: `headers` is `[][]mdRun`, `rows` is `[][][]mdRun`
  - Horizontal rules: empty block with `kind = blockHRule`
  - Definition lists: `items` with `term` for the term, `runs` for the definition
  - Task list items: `checked` is `*bool` (non-nil)
  - Bold text → `runBold`, italic → `runItalic`, both → `runBoldItalic`
  - Inline code → `runCode`
  - Links → `runLink` with URL in `url` field, display text in `text`
  - Images → `runCode` with text `[IMG: alt]`
  - Strikethrough → `runStrikethrough`
  - Adjacent runs with the same style are merged into a single run
  - Nested emphasis is resolved: bold inside italic = `runBoldItalic`

**Implementation:**

First add goldmark:
```bash
go get github.com/yuin/goldmark
```

The goldmark definition list extension is at `github.com/yuin/goldmark/extension` (same module, no extra dependency).

```go
package tv

import (
    "strings"

    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/ast"
    east "github.com/yuin/goldmark/extension/ast"
    "github.com/yuin/goldmark/extension"
    "github.com/yuin/goldmark/text"
)

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

type mdRunStyle int

const (
    runNormal mdRunStyle = iota
    runBold
    runItalic
    runBoldItalic
    runCode
    runLink
    runStrikethrough
)

type mdBlock struct {
    kind     mdBlockKind
    level    int
    runs     []mdRun
    language string
    code     []string
    items    []mdItem
    children []mdBlock
    headers  [][]mdRun
    rows     [][][]mdRun
}

type mdRun struct {
    text  string
    style mdRunStyle
    url   string
}

type mdItem struct {
    runs     []mdRun
    children []mdBlock
    checked  *bool
    term     []mdRun
}

func parseMarkdown(src string) []mdBlock {
    md := goldmark.New(
        goldmark.WithExtensions(
            extension.GFM,             // tables, strikethrough, task list
            extension.DefinitionList,  // definition lists
        ),
    )
    source := []byte(src)
    reader := text.NewReader(source)
    doc := md.Parser().Parse(reader)
    return walkBlocks(doc, source)
}

// walkBlocks iterates over children of a goldmark AST node and
// converts each to an mdBlock.
func walkBlocks(parent ast.Node, source []byte) []mdBlock {
    var blocks []mdBlock
    for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
        if b, ok := convertNode(child, source); ok {
            blocks = append(blocks, b)
        }
    }
    return blocks
}

// convertNode converts a single goldmark AST node to an mdBlock.
// Returns false if the node type is not supported.
func convertNode(node ast.Node, source []byte) (mdBlock, bool) {
    switch n := node.(type) {
    case *ast.Paragraph:
        return mdBlock{kind: blockParagraph, runs: collectInlineRuns(n, source, runNormal)}, true

    case *ast.Heading:
        return mdBlock{kind: blockHeader, level: n.Level, runs: collectInlineRuns(n, source, runNormal)}, true

    case *ast.FencedCodeBlock:
        lang := ""
        if n.Info != nil {
            lang = strings.TrimSpace(string(n.Info.Text(source)))
            // strip anything after first space (```go title="foo")
            if idx := strings.IndexByte(lang, ' '); idx >= 0 {
                lang = lang[:idx]
            }
        }
        var lines []string
        for i := 0; i < n.Lines().Len(); i++ {
            seg := n.Lines().At(i)
            line := string(seg.Value(source))
            line = strings.TrimRight(line, "\n")
            lines = append(lines, line)
        }
        return mdBlock{kind: blockCodeBlock, language: lang, code: lines}, true

    case *ast.CodeBlock:
        var lines []string
        for i := 0; i < n.Lines().Len(); i++ {
            seg := n.Lines().At(i)
            line := strings.TrimRight(string(seg.Value(source)), "\n")
            lines = append(lines, line)
        }
        return mdBlock{kind: blockCodeBlock, code: lines}, true

    case *ast.List:
        return convertList(n, source), true

    case *ast.Blockquote:
        return mdBlock{kind: blockBlockquote, children: walkBlocks(n, source)}, true

    case *east.Table:
        return convertTable(n, source), true

    case *ast.ThematicBreak:
        return mdBlock{kind: blockHRule}, true

    case *east.DefinitionList:
        return convertDefList(n, source), true
    }
    return mdBlock{}, false
}

// convertList converts an ast.List (ordered or unordered) to an mdBlock.
// Detects task list items by checking for east.TaskCheckBox children.
func convertList(list *ast.List, source []byte) mdBlock {
    kind := blockBulletList
    if list.IsOrdered() {
        kind = blockNumberList
    }
    // Detect if this is a task/check list: first item has a TaskCheckBox
    isCheckList := false
    if first := list.FirstChild(); first != nil {
        for c := first.FirstChild(); c != nil; c = c.NextSibling() {
            if _, ok := c.(*ast.TextBlock); ok {
                for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
                    if _, ok := gc.(*east.TaskCheckBox); ok {
                        isCheckList = true
                        break
                    }
                }
            }
            if isCheckList {
                break
            }
        }
    }
    if isCheckList {
        kind = blockCheckList
    }

    var items []mdItem
    for child := list.FirstChild(); child != nil; child = child.NextSibling() {
        if li, ok := child.(*ast.ListItem); ok {
            items = append(items, convertListItem(li, source, isCheckList))
        }
    }

    b := mdBlock{kind: kind, items: items}
    if list.IsOrdered() {
        b.level = list.Start
    }
    return b
}

// convertListItem converts an ast.ListItem to an mdItem.
func convertListItem(li *ast.ListItem, source []byte, isCheckList bool) mdItem {
    item := mdItem{}

    // Collect inline content from first TextBlock/Paragraph child
    // and nested block children (sub-lists, code blocks, etc.)
    for child := li.FirstChild(); child != nil; child = child.NextSibling() {
        switch c := child.(type) {
        case *ast.TextBlock:
            runs := collectInlineRuns(c, source, runNormal)
            // Check for TaskCheckBox — it will appear as a child
            if isCheckList {
                for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
                    if cb, ok := gc.(*east.TaskCheckBox); ok {
                        checked := cb.IsChecked
                        item.checked = &checked
                        break
                    }
                }
            }
            item.runs = append(item.runs, runs...)
        case *ast.Paragraph:
            item.runs = append(item.runs, collectInlineRuns(c, source, runNormal)...)
        default:
            if b, ok := convertNode(child, source); ok {
                item.children = append(item.children, b)
            }
        }
    }
    return item
}

// convertTable converts an east.Table to an mdBlock.
func convertTable(table *east.Table, source []byte) mdBlock {
    b := mdBlock{kind: blockTable}
    for child := table.FirstChild(); child != nil; child = child.NextSibling() {
        switch row := child.(type) {
        case *east.TableHeader:
            // Header row — single row with cells
            for tr := row.FirstChild(); tr != nil; tr = tr.NextSibling() {
                if tableRow, ok := tr.(*east.TableRow); ok {
                    var cells [][]mdRun
                    for cell := tableRow.FirstChild(); cell != nil; cell = cell.NextSibling() {
                        if tc, ok := cell.(*east.TableCell); ok {
                            cells = append(cells, collectInlineRuns(tc, source, runNormal))
                        }
                    }
                    b.headers = cells
                }
            }
            // If no TableRow children, cells are direct children of TableHeader
            if b.headers == nil {
                for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
                    if tc, ok := cell.(*east.TableCell); ok {
                        b.headers = append(b.headers, collectInlineRuns(tc, source, runNormal))
                    }
                }
            }
        case *east.TableBody:
            for tr := row.FirstChild(); tr != nil; tr = tr.NextSibling() {
                if tableRow, ok := tr.(*east.TableRow); ok {
                    var cells [][]mdRun
                    for cell := tableRow.FirstChild(); cell != nil; cell = cell.NextSibling() {
                        if tc, ok := cell.(*east.TableCell); ok {
                            cells = append(cells, collectInlineRuns(tc, source, runNormal))
                        }
                    }
                    b.rows = append(b.rows, cells)
                }
            }
        }
    }
    return b
}

// convertDefList converts an east.DefinitionList to an mdBlock.
func convertDefList(dl *east.DefinitionList, source []byte) mdBlock {
    b := mdBlock{kind: blockDefList}
    var currentItem *mdItem
    for child := dl.FirstChild(); child != nil; child = child.NextSibling() {
        switch c := child.(type) {
        case *east.DefinitionTerm:
            if currentItem != nil {
                b.items = append(b.items, *currentItem)
            }
            currentItem = &mdItem{
                term: collectInlineRuns(c, source, runNormal),
            }
        case *east.DefinitionDescription:
            if currentItem == nil {
                currentItem = &mdItem{}
            }
            // Description can contain block-level content
            for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
                if para, ok := gc.(*ast.Paragraph); ok {
                    currentItem.runs = append(currentItem.runs, collectInlineRuns(para, source, runNormal)...)
                } else if blk, ok := convertNode(gc, source); ok {
                    currentItem.children = append(currentItem.children, blk)
                }
            }
        }
    }
    if currentItem != nil {
        b.items = append(b.items, *currentItem)
    }
    return b
}

// collectInlineRuns walks inline children of a node and builds styled runs.
// parentStyle is the inherited style (e.g., runNormal inside a paragraph,
// but could be runBold if we're inside a bold span).
func collectInlineRuns(node ast.Node, source []byte, parentStyle mdRunStyle) []mdRun {
    var runs []mdRun
    for child := node.FirstChild(); child != nil; child = child.NextSibling() {
        switch c := child.(type) {
        case *ast.Text:
            t := string(c.Text(source))
            if c.SoftLineBreak() {
                t += " "
            }
            if t != "" {
                runs = append(runs, mdRun{text: t, style: parentStyle})
            }
        case *ast.String:
            t := string(c.Value)
            if t != "" {
                runs = append(runs, mdRun{text: t, style: parentStyle})
            }
        case *ast.CodeSpan:
            var buf strings.Builder
            for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
                if t, ok := gc.(*ast.Text); ok {
                    buf.Write(t.Text(source))
                }
            }
            if buf.Len() > 0 {
                runs = append(runs, mdRun{text: buf.String(), style: runCode})
            }
        case *ast.Emphasis:
            innerStyle := runItalic
            if c.Level == 2 {
                innerStyle = runBold
            }
            // Handle nesting: bold inside italic = boldItalic
            if parentStyle == runBold && innerStyle == runItalic {
                innerStyle = runBoldItalic
            } else if parentStyle == runItalic && innerStyle == runBold {
                innerStyle = runBoldItalic
            }
            runs = append(runs, collectInlineRuns(c, source, innerStyle)...)
        case *ast.Link:
            var buf strings.Builder
            for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
                if t, ok := gc.(*ast.Text); ok {
                    buf.Write(t.Text(source))
                }
            }
            runs = append(runs, mdRun{
                text:  buf.String(),
                style: runLink,
                url:   string(c.Destination),
            })
        case *ast.Image:
            alt := string(c.Text(source))
            if alt == "" {
                alt = "image"
            }
            runs = append(runs, mdRun{text: "[IMG: " + alt + "]", style: runCode})
        case *ast.AutoLink:
            url := string(c.URL(source))
            runs = append(runs, mdRun{text: url, style: runLink, url: url})
        case *east.Strikethrough:
            runs = append(runs, collectInlineRuns(c, source, runStrikethrough)...)
        case *east.TaskCheckBox:
            // Handled at list item level, skip here
        default:
            // Recurse into unknown inline containers
            runs = append(runs, collectInlineRuns(child, source, parentStyle)...)
        }
    }
    return mergeRuns(runs)
}

// mergeRuns combines adjacent runs with the same style into single runs.
func mergeRuns(runs []mdRun) []mdRun {
    if len(runs) <= 1 {
        return runs
    }
    merged := []mdRun{runs[0]}
    for _, r := range runs[1:] {
        last := &merged[len(merged)-1]
        if last.style == r.style && last.url == r.url {
            last.text += r.text
        } else {
            merged = append(merged, r)
        }
    }
    return merged
}
```

**Run tests:** `go test ./tv/ -run TestMarkdown -v`

**Commit:** `git commit -m "feat: goldmark-based Markdown IR parser"`

---

### Task 3: Rendering Helpers — Word Wrap, Style Composition, Block Layout

**Files:**
- Create: `tv/markdown_render.go`

**Requirements:**

**Word wrapping (`wrapRuns`):**
- Given `[]mdRun` and a maximum width, returns `[][]mdRun` (one slice per visual line)
- Wraps at word boundaries (spaces). If a single word exceeds the width, it is broken at the width boundary
- Preserves run styles across line breaks — a bold word that wraps keeps its bold style
- An empty input returns a single empty line

**Style composition (`composeStyle`):**
- Given a block `tcell.Style` and an `mdRunStyle`, returns the composed `tcell.Style`
- For `runNormal`: returns the block style unchanged
- For `runBold`, `runItalic`, `runBoldItalic`, `runStrikethrough`: takes foreground and attributes from the corresponding `ColorScheme` field, keeps background from the block style
- For `runCode`: returns `MarkdownCode` directly (own background)
- For `runLink`: takes foreground and underline from `MarkdownLink`, keeps background from block style

**Rendered height calculation (`renderedHeight`):**
- Given `[]mdBlock`, widget width, wrap setting, and nesting depth, returns total number of rendered lines
- Accounts for: word-wrapped paragraphs, code block line count, inter-block blank lines, list item indentation reducing available width, blockquote indentation, table row count (with cell wrapping)

**Rendered line retrieval (`renderLine`):**
- Given `[]mdBlock`, a target line number `y`, widget width, wrap setting, returns the characters and styles for that line
- This is the core rendering function called by `Draw` for each visible row
- Must handle the viewport offset: Draw calls `renderLine(blocks, deltaY + row, width, wrapText)` for each row in the viewport

**Table layout (`layoutTable`):**
- Given a table block and available width, returns column widths
- Minimum column width: longest word in the column or 8 chars, whichever is greater
- If minimums + borders fit: distribute extra space proportionally to content width
- If minimums + borders don't fit: use minimums (horizontal scroll handles overflow)
- Border overhead: numCols + 1 characters (one `│` per column boundary plus edges)

**Implementation:**

```go
package tv

import (
    "fmt"
    "strings"
    "unicode"

    "github.com/gdamore/tcell/v2"
    "github.com/njt/turboview/theme"
)

// wrapRuns word-wraps a slice of runs to fit within maxWidth columns.
// Returns one []mdRun per visual line. Code runs are never broken mid-word
// but other text wraps at word boundaries.
func wrapRuns(runs []mdRun, maxWidth int) [][]mdRun {
    if maxWidth <= 0 {
        return [][]mdRun{nil}
    }
    if len(runs) == 0 {
        return [][]mdRun{nil}
    }

    var lines [][]mdRun
    var curLine []mdRun
    col := 0

    for _, run := range runs {
        words := splitWords(run.text)
        for _, word := range words {
            wLen := len([]rune(word))
            if wLen == 0 {
                continue
            }
            // Space handling: if word is just spaces, add to current line
            trimmed := strings.TrimSpace(word)
            if trimmed == "" {
                if col > 0 && col+wLen <= maxWidth {
                    curLine = append(curLine, mdRun{text: word, style: run.style, url: run.url})
                    col += wLen
                }
                continue
            }
            // If word fits on current line
            if col+wLen <= maxWidth {
                curLine = append(curLine, mdRun{text: word, style: run.style, url: run.url})
                col += wLen
            } else if col == 0 {
                // Word too long for any line — force break
                runes := []rune(word)
                for len(runes) > 0 {
                    take := maxWidth
                    if take > len(runes) {
                        take = len(runes)
                    }
                    curLine = append(curLine, mdRun{text: string(runes[:take]), style: run.style, url: run.url})
                    lines = append(lines, mergeRuns(curLine))
                    curLine = nil
                    col = 0
                    runes = runes[take:]
                }
            } else {
                // Wrap to next line
                lines = append(lines, mergeRuns(curLine))
                curLine = []mdRun{{text: word, style: run.style, url: run.url}}
                col = wLen
            }
        }
    }
    if curLine != nil || len(lines) == 0 {
        lines = append(lines, mergeRuns(curLine))
    }
    return lines
}

// splitWords splits text into alternating sequences of non-space and space tokens.
func splitWords(text string) []string {
    var words []string
    runes := []rune(text)
    i := 0
    for i < len(runes) {
        if unicode.IsSpace(runes[i]) {
            j := i
            for j < len(runes) && unicode.IsSpace(runes[j]) {
                j++
            }
            words = append(words, string(runes[i:j]))
            i = j
        } else {
            j := i
            for j < len(runes) && !unicode.IsSpace(runes[j]) {
                j++
            }
            words = append(words, string(runes[i:j]))
            i = j
        }
    }
    return words
}

// composeStyle combines a block background style with an inline run style.
// Block style provides the background; inline style provides foreground and attributes.
func composeStyle(blockStyle tcell.Style, runStyle mdRunStyle, cs *theme.ColorScheme) tcell.Style {
    if cs == nil {
        return blockStyle
    }
    switch runStyle {
    case runNormal:
        return blockStyle
    case runBold:
        return overlayFgAttrs(blockStyle, cs.MarkdownBold)
    case runItalic:
        return overlayFgAttrs(blockStyle, cs.MarkdownItalic)
    case runBoldItalic:
        return overlayFgAttrs(blockStyle, cs.MarkdownBoldItalic)
    case runCode:
        return cs.MarkdownCode // own background
    case runLink:
        return overlayFgAttrs(blockStyle, cs.MarkdownLink)
    case runStrikethrough:
        return overlayFgAttrs(blockStyle, cs.MarkdownBold).StrikeThrough(true)
    }
    return blockStyle
}

// overlayFgAttrs takes the foreground color and attributes from src
// and applies them to dst, keeping dst's background.
func overlayFgAttrs(dst, src tcell.Style) tcell.Style {
    fg, _, _ := src.Decompose()
    _, bg, _ := dst.Decompose()
    result := tcell.StyleDefault.Foreground(fg).Background(bg)
    // Copy attributes from src
    if _, _, attrs := src.Decompose(); attrs != 0 {
        result = result.Attributes(attrs)
    }
    return result
}

// mdRenderer holds state for rendering blocks into a flat sequence of styled lines.
type mdRenderer struct {
    blocks   []mdBlock
    width    int
    wrapText bool
    cs       *theme.ColorScheme
}

// renderedHeight returns the total number of visual lines when the blocks
// are rendered at the given width.
func (r *mdRenderer) renderedHeight() int {
    return r.blocksHeight(r.blocks, 0)
}

// blocksHeight returns the height of a slice of blocks at a given indent depth.
func (r *mdRenderer) blocksHeight(blocks []mdBlock, depth int) int {
    h := 0
    for i, b := range blocks {
        if i > 0 {
            h++ // blank line between blocks
        }
        h += r.blockHeight(b, depth)
    }
    return h
}

// blockHeight returns the height of a single block.
func (r *mdRenderer) blockHeight(b mdBlock, depth int) int {
    indent := depth * 2
    avail := r.width - indent
    if avail < 1 {
        avail = 1
    }

    switch b.kind {
    case blockParagraph, blockHeader:
        if r.wrapText {
            return len(wrapRuns(b.runs, avail))
        }
        return 1

    case blockCodeBlock:
        if len(b.code) == 0 {
            return 1
        }
        return len(b.code)

    case blockBulletList, blockNumberList, blockCheckList:
        return r.listHeight(b, depth)

    case blockBlockquote:
        return r.blocksHeight(b.children, depth+1)

    case blockTable:
        return r.tableHeight(b, avail)

    case blockHRule:
        return 1

    case blockDefList:
        return r.defListHeight(b, depth)
    }
    return 1
}

func (r *mdRenderer) listHeight(b mdBlock, depth int) int {
    h := 0
    for _, item := range b.items {
        // Item text
        markerWidth := 4 // "  • " or "  1. "
        indent := depth*2 + markerWidth
        avail := r.width - indent
        if avail < 1 {
            avail = 1
        }
        if r.wrapText {
            h += len(wrapRuns(item.runs, avail))
        } else {
            h++
        }
        // Nested children
        if len(item.children) > 0 {
            h += r.blocksHeight(item.children, depth+1)
        }
    }
    return h
}

func (r *mdRenderer) defListHeight(b mdBlock, depth int) int {
    h := 0
    for i, item := range b.items {
        if i > 0 {
            h++ // blank line between definition items
        }
        h++ // term line
        defAvail := r.width - depth*2 - 4 // 4-char indent for definition
        if defAvail < 1 {
            defAvail = 1
        }
        if r.wrapText {
            h += len(wrapRuns(item.runs, defAvail))
        } else {
            h++
        }
    }
    return h
}

func (r *mdRenderer) tableHeight(b mdBlock, avail int) int {
    if len(b.headers) == 0 && len(b.rows) == 0 {
        return 0
    }
    colWidths := layoutTable(b, avail)
    h := 1 // top border
    if len(b.headers) > 0 {
        h += r.tableRowHeight(b.headers, colWidths) // header cells
        h++ // separator line
    }
    for _, row := range b.rows {
        h += r.tableRowHeight(row, colWidths)
    }
    h++ // bottom border
    return h
}

func (r *mdRenderer) tableRowHeight(cells [][]mdRun, colWidths []int) int {
    maxH := 1
    for i, cell := range cells {
        if i >= len(colWidths) {
            break
        }
        if r.wrapText {
            lines := wrapRuns(cell, colWidths[i])
            if len(lines) > maxH {
                maxH = len(lines)
            }
        }
    }
    return maxH
}

// layoutTable computes column widths for a table block.
func layoutTable(b mdBlock, availWidth int) []int {
    numCols := len(b.headers)
    if numCols == 0 && len(b.rows) > 0 {
        numCols = len(b.rows[0])
    }
    if numCols == 0 {
        return nil
    }

    // Compute minimum widths: longest word in each column, floor of 8
    minWidths := make([]int, numCols)
    contentWidths := make([]int, numCols) // total content width for proportional distribution
    for i := range minWidths {
        minWidths[i] = 8
    }

    measureCell := func(cell []mdRun, col int) {
        totalLen := 0
        for _, run := range cell {
            totalLen += len([]rune(run.text))
            for _, word := range strings.Fields(run.text) {
                wl := len([]rune(word))
                if wl > minWidths[col] {
                    minWidths[col] = wl
                }
            }
        }
        if totalLen > contentWidths[col] {
            contentWidths[col] = totalLen
        }
    }

    for i, cell := range b.headers {
        if i < numCols {
            measureCell(cell, i)
        }
    }
    for _, row := range b.rows {
        for i, cell := range row {
            if i < numCols {
                measureCell(cell, i)
            }
        }
    }

    borderOverhead := numCols + 1 // │ per boundary
    totalMin := borderOverhead
    for _, w := range minWidths {
        totalMin += w
    }

    if totalMin >= availWidth {
        return minWidths // horizontal scroll will handle overflow
    }

    // Distribute extra space proportionally to content width
    extra := availWidth - totalMin
    totalContent := 0
    for _, cw := range contentWidths {
        totalContent += cw
    }

    result := make([]int, numCols)
    copy(result, minWidths)
    if totalContent > 0 && extra > 0 {
        distributed := 0
        for i, cw := range contentWidths {
            share := extra * cw / totalContent
            result[i] += share
            distributed += share
        }
        // Distribute remainder to first columns
        for i := 0; distributed < extra && i < numCols; i++ {
            result[i]++
            distributed++
        }
    }

    return result
}
```

**Run tests:** `go test ./tv/ -run TestMarkdown -v`

**Commit:** `git commit -m "feat: Markdown rendering helpers — word wrap, style composition, table layout"`

---

### Task 4: Integration Checkpoint — IR + Rendering

**Purpose:** Verify that Tasks 2 and 3 work together: parse Markdown, compute rendered height, retrieve rendered lines with correct styles.

**Requirements (for test writer):**
- Parse `"# Hello\n\nWorld"` → rendered height is 3 (header + blank line + paragraph) at width 40
- Parse a paragraph with `**bold**` → `composeStyle` with `runBold` on `MarkdownNormal` background produces `MarkdownBold` foreground with `MarkdownNormal` background
- Parse a fenced code block → rendered height equals number of lines in the block
- Parse a nested list (2 levels) → rendered height accounts for indentation reducing available width
- `wrapRuns` on a paragraph that exceeds width → returns multiple lines, all runs preserve their styles
- `layoutTable` on a 3-column table within a 40-char width → returns column widths that sum + borders ≤ 40

**Components to wire up:** `parseMarkdown`, `mdRenderer`, `wrapRuns`, `composeStyle`, `layoutTable` (all real, no mocks)

**Run:** `go test ./tv/ -run TestIntegration -v`

---

### Task 5: MarkdownViewer Widget — Constructor, Draw, HandleEvent

**Files:**
- Create: `tv/markdown_viewer.go`

**Requirements:**

**Constructor:**
- `NewMarkdownViewer(bounds Rect)` returns a `*MarkdownViewer`
- Sets `SfVisible`, `OfSelectable`, `OfFirstClick`
- Calls `SetSelf(mv)`
- `wrapText` defaults to `true`
- `deltaX` and `deltaY` default to 0

**SetMarkdown:**
- `SetMarkdown(text string)` parses via `parseMarkdown` and stores blocks and source
- Resets `deltaX` and `deltaY` to 0
- Calls `syncScrollBars()`

**Markdown getter:**
- `Markdown()` returns the original source string passed to `SetMarkdown`

**Draw:**
- Fills entire bounds with `MarkdownNormal` style
- Walks blocks using `mdRenderer`, rendering lines `deltaY` through `deltaY + height - 1`
- Each rendered line writes styled characters via `buf.WriteChar`
- Code blocks use `MarkdownCodeBlock` as background, filling the full widget width
- Headers use the corresponding `MarkdownH1`-`MarkdownH6` style
- Horizontal rules render as `─` characters across the full width in `MarkdownHRule` style
- Lists render with markers (`•`, `1.`, `☐`, `☑`) in `MarkdownListMarker` style
- Blockquotes render with `▌` left bar in `MarkdownBlockquote` style
- Tables render with box-drawing borders in `MarkdownTableBorder` style
- Definition terms render in `MarkdownDefTerm` style
- Horizontal scroll (deltaX) offsets all content — only relevant when content exceeds width

**HandleEvent (keyboard — only when SfSelected):**
- Up: `deltaY--` (clamped to 0)
- Down: `deltaY++` (clamped to max)
- PgUp: `deltaY -= viewport height`
- PgDn: `deltaY += viewport height`
- Home: `deltaY = 0`
- End: `deltaY = max`
- Left: `deltaX--` (clamped to 0)
- Right: `deltaX++` (clamped to max)
- W (rune 'w' or 'W'): toggle `wrapText`, reset `deltaX` to 0
- All handled keys are consumed (event.Clear())
- Unrecognized keys pass through

**HandleEvent (mouse):**
- Calls `BaseView.HandleEvent` first (click-to-focus)
- WheelUp/WheelDown: `deltaY ∓ 3`
- WheelLeft/WheelRight: `deltaX ∓ 3`
- Button1: consumed (focus only, handled by BaseView)

**Scrollbar binding:**
- `SetVScrollBar(sb)`: binds `sb.OnChange` to update `deltaY`, calls `syncScrollBars()`
- `SetHScrollBar(sb)`: binds `sb.OnChange` to update `deltaX`, calls `syncScrollBars()`
- `SetVScrollBar(nil)` / `SetHScrollBar(nil)`: clears `OnChange` on old scrollbar
- `syncScrollBars()`: updates range (`totalHeight - 1 + pageSize` for vertical, `maxWidth - 1 + pageSize` for horizontal), page size (viewport dimensions), value (current delta)

**SetBounds override:**
- `SetBounds(bounds)` calls `BaseView.SetBounds(bounds)` then `syncScrollBars()` so resize recalculates scrollbar ranges

**SetState override:**
- When `SfSelected` changes, toggle scrollbar visibility (same pattern as Memo)

**WrapText:**
- `SetWrapText(wrap bool)` sets `wrapText`, resets `deltaX` to 0, calls `syncScrollBars()`
- `WrapText()` returns current value

**Implementation:**

```go
package tv

import "github.com/gdamore/tcell/v2"

var _ Widget = (*MarkdownViewer)(nil)

type MarkdownViewer struct {
    BaseView
    blocks     []mdBlock
    source     string
    deltaX     int
    deltaY     int
    wrapText   bool
    vScrollBar *ScrollBar
    hScrollBar *ScrollBar
}

func NewMarkdownViewer(bounds Rect) *MarkdownViewer {
    mv := &MarkdownViewer{wrapText: true}
    mv.SetBounds(bounds)
    mv.SetState(SfVisible, true)
    mv.SetOptions(OfSelectable|OfFirstClick, true)
    mv.SetSelf(mv)
    return mv
}

func (mv *MarkdownViewer) Markdown() string    { return mv.source }
func (mv *MarkdownViewer) WrapText() bool      { return mv.wrapText }
func (mv *MarkdownViewer) DeltaX() int         { return mv.deltaX }
func (mv *MarkdownViewer) DeltaY() int         { return mv.deltaY }

func (mv *MarkdownViewer) SetMarkdown(text string) {
    mv.source = text
    mv.blocks = parseMarkdown(text)
    mv.deltaX = 0
    mv.deltaY = 0
    mv.syncScrollBars()
}

func (mv *MarkdownViewer) SetWrapText(wrap bool) {
    mv.wrapText = wrap
    mv.deltaX = 0
    mv.syncScrollBars()
}

func (mv *MarkdownViewer) SetVScrollBar(sb *ScrollBar) {
    if mv.vScrollBar != nil {
        mv.vScrollBar.OnChange = nil
    }
    mv.vScrollBar = sb
    if sb != nil {
        sb.OnChange = func(val int) {
            mv.deltaY = val
        }
        mv.syncScrollBars()
    }
}

func (mv *MarkdownViewer) SetHScrollBar(sb *ScrollBar) {
    if mv.hScrollBar != nil {
        mv.hScrollBar.OnChange = nil
    }
    mv.hScrollBar = sb
    if sb != nil {
        sb.OnChange = func(val int) {
            mv.deltaX = val
        }
        mv.syncScrollBars()
    }
}

func (mv *MarkdownViewer) SetBounds(bounds Rect) {
    mv.BaseView.SetBounds(bounds)
    mv.syncScrollBars()
}

func (mv *MarkdownViewer) SetState(flag ViewState, on bool) {
    mv.BaseView.SetState(flag, on)
    if flag&SfSelected != 0 {
        if mv.vScrollBar != nil {
            mv.vScrollBar.SetState(SfVisible, on)
        }
        if mv.hScrollBar != nil {
            mv.hScrollBar.SetState(SfVisible, on)
        }
    }
}

func (mv *MarkdownViewer) renderer() *mdRenderer {
    return &mdRenderer{
        blocks:   mv.blocks,
        width:    mv.Bounds().Width(),
        wrapText: mv.wrapText,
        cs:       mv.ColorScheme(),
    }
}

func (mv *MarkdownViewer) syncScrollBars() {
    r := mv.renderer()
    totalH := r.renderedHeight()
    vpH := mv.Bounds().Height()

    // Clamp deltaY
    maxDY := totalH - vpH
    if maxDY < 0 {
        maxDY = 0
    }
    if mv.deltaY > maxDY {
        mv.deltaY = maxDY
    }
    if mv.deltaY < 0 {
        mv.deltaY = 0
    }

    if mv.vScrollBar != nil {
        maxRange := totalH - 1 + vpH
        if maxRange < 0 {
            maxRange = 0
        }
        mv.vScrollBar.SetRange(0, maxRange)
        mv.vScrollBar.SetPageSize(vpH)
        mv.vScrollBar.SetValue(mv.deltaY)
    }

    // Horizontal: compute max content width
    maxW := r.maxContentWidth()
    vpW := mv.Bounds().Width()
    maxDX := maxW - vpW
    if maxDX < 0 {
        maxDX = 0
    }
    if mv.deltaX > maxDX {
        mv.deltaX = maxDX
    }
    if mv.deltaX < 0 {
        mv.deltaX = 0
    }

    if mv.hScrollBar != nil {
        maxRange := maxW - 1 + vpW
        if maxRange < 0 {
            maxRange = 0
        }
        mv.hScrollBar.SetRange(0, maxRange)
        mv.hScrollBar.SetPageSize(vpW)
        mv.hScrollBar.SetValue(mv.deltaX)
    }
}

func (mv *MarkdownViewer) Draw(buf *DrawBuffer) {
    w := mv.Bounds().Width()
    h := mv.Bounds().Height()
    cs := mv.ColorScheme()
    normalStyle := tcell.StyleDefault
    if cs != nil {
        normalStyle = cs.MarkdownNormal
    }
    buf.Fill(NewRect(0, 0, w, h), ' ', normalStyle)

    if len(mv.blocks) == 0 {
        return
    }

    r := mv.renderer()
    for row := 0; row < h; row++ {
        lineY := mv.deltaY + row
        r.renderLineInto(buf, lineY, row, mv.deltaX, w)
    }
}

func (mv *MarkdownViewer) HandleEvent(event *Event) {
    if event.What == EvMouse && event.Mouse != nil {
        mv.BaseView.HandleEvent(event)
        if event.IsCleared() {
            return
        }
        switch {
        case event.Mouse.Button == tcell.WheelUp:
            mv.deltaY -= 3
            if mv.deltaY < 0 {
                mv.deltaY = 0
            }
            mv.syncScrollBars()
            event.Clear()
        case event.Mouse.Button == tcell.WheelDown:
            mv.deltaY += 3
            mv.syncScrollBars()
            event.Clear()
        case event.Mouse.Button == tcell.WheelLeft:
            mv.deltaX -= 3
            if mv.deltaX < 0 {
                mv.deltaX = 0
            }
            mv.syncScrollBars()
            event.Clear()
        case event.Mouse.Button == tcell.WheelRight:
            mv.deltaX += 3
            mv.syncScrollBars()
            event.Clear()
        case event.Mouse.Button&tcell.Button1 != 0:
            event.Clear()
        }
        return
    }

    if event.What != EvKeyboard || event.Key == nil {
        return
    }
    if !mv.HasState(SfSelected) {
        return
    }

    r := mv.renderer()
    totalH := r.renderedHeight()
    vpH := mv.Bounds().Height()

    // W toggle
    if event.Key.Key == tcell.KeyRune && (event.Key.Rune == 'w' || event.Key.Rune == 'W') {
        mv.SetWrapText(!mv.wrapText)
        event.Clear()
        return
    }

    switch event.Key.Key {
    case tcell.KeyUp:
        if mv.deltaY > 0 {
            mv.deltaY--
            mv.syncScrollBars()
        }
        event.Clear()
    case tcell.KeyDown:
        maxDY := totalH - vpH
        if maxDY < 0 {
            maxDY = 0
        }
        if mv.deltaY < maxDY {
            mv.deltaY++
            mv.syncScrollBars()
        }
        event.Clear()
    case tcell.KeyPgUp:
        mv.deltaY -= vpH
        if mv.deltaY < 0 {
            mv.deltaY = 0
        }
        mv.syncScrollBars()
        event.Clear()
    case tcell.KeyPgDn:
        mv.deltaY += vpH
        mv.syncScrollBars()
        event.Clear()
    case tcell.KeyHome:
        mv.deltaY = 0
        mv.syncScrollBars()
        event.Clear()
    case tcell.KeyEnd:
        maxDY := totalH - vpH
        if maxDY < 0 {
            maxDY = 0
        }
        mv.deltaY = maxDY
        mv.syncScrollBars()
        event.Clear()
    case tcell.KeyLeft:
        if mv.deltaX > 0 {
            mv.deltaX--
            mv.syncScrollBars()
        }
        event.Clear()
    case tcell.KeyRight:
        mv.deltaX++
        mv.syncScrollBars()
        event.Clear()
    }
}
```

The `renderLineInto` method on `mdRenderer` is the core Draw loop helper — it computes which block and sub-line a given visual line `y` falls into, resolves the indent/style, and writes characters into the DrawBuffer. This method should be implemented in `markdown_render.go` and will need the full block-walking logic. Given its complexity, the implementation should build rendered lines by walking the block tree with a running line counter, stopping when it reaches the target line.

```go
// Add to markdown_render.go:

// maxContentWidth returns the widest line in the rendered output.
func (r *mdRenderer) maxContentWidth() int {
    return r.blocksMaxWidth(r.blocks, 0)
}

func (r *mdRenderer) blocksMaxWidth(blocks []mdBlock, depth int) int {
    maxW := 0
    for _, b := range blocks {
        w := r.blockMaxWidth(b, depth)
        if w > maxW {
            maxW = w
        }
    }
    return maxW
}

func (r *mdRenderer) blockMaxWidth(b mdBlock, depth int) int {
    indent := depth * 2
    switch b.kind {
    case blockCodeBlock:
        maxW := 0
        for _, line := range b.code {
            w := len([]rune(line)) + indent
            if w > maxW {
                maxW = w
            }
        }
        return maxW
    case blockTable:
        colWidths := layoutTable(b, r.width-indent)
        total := indent + len(colWidths) + 1
        for _, cw := range colWidths {
            total += cw
        }
        return total
    case blockBlockquote:
        return r.blocksMaxWidth(b.children, depth) + 2
    default:
        if !r.wrapText {
            w := indent + runsWidth(b.runs)
            return w
        }
    }
    return r.width
}

func runsWidth(runs []mdRun) int {
    w := 0
    for _, run := range runs {
        w += len([]rune(run.text))
    }
    return w
}

// renderLineInto renders visual line `lineY` into the draw buffer at screen row `screenY`,
// applying horizontal scroll offset `dx` within available width `w`.
func (r *mdRenderer) renderLineInto(buf *DrawBuffer, lineY, screenY, dx, w int) {
    // Walk blocks, tracking current visual line number.
    // When we reach lineY, render that line and return.
    cur := 0
    r.renderBlocksLine(buf, r.blocks, 0, lineY, screenY, dx, w, &cur, r.cs.MarkdownNormal)
}

// renderBlocksLine walks blocks looking for the target lineY.
// cur tracks the current visual line position. Returns true if lineY was rendered.
func (r *mdRenderer) renderBlocksLine(buf *DrawBuffer, blocks []mdBlock, depth int, lineY, screenY, dx, w int, cur *int, blockStyle tcell.Style) bool {
    for i, b := range blocks {
        if i > 0 {
            if *cur == lineY {
                return true // blank separator line — already filled with bg
            }
            *cur++
        }
        if r.renderBlockLine(buf, b, depth, lineY, screenY, dx, w, cur, blockStyle) {
            return true
        }
    }
    return false
}

// renderBlockLine renders a single block's contribution at lineY.
func (r *mdRenderer) renderBlockLine(buf *DrawBuffer, b mdBlock, depth int, lineY, screenY, dx, w int, cur *int, parentBlockStyle tcell.Style) bool {
    cs := r.cs
    if cs == nil {
        return false
    }
    indent := depth * 2

    switch b.kind {
    case blockParagraph:
        return r.renderParagraphLine(buf, b.runs, parentBlockStyle, indent, lineY, screenY, dx, w, cur)

    case blockHeader:
        style := r.headerStyle(b.level)
        return r.renderParagraphLine(buf, b.runs, style, indent, lineY, screenY, dx, w, cur)

    case blockCodeBlock:
        if len(b.code) == 0 {
            // Empty code block: render one blank line with code block background
            if *cur == lineY {
                for x := 0; x < w; x++ {
                    buf.WriteChar(x, screenY, ' ', cs.MarkdownCodeBlock)
                }
                return true
            }
            *cur++
        } else {
            for _, line := range b.code {
                if *cur == lineY {
                    for x := 0; x < w; x++ {
                        buf.WriteChar(x, screenY, ' ', cs.MarkdownCodeBlock)
                    }
                    runes := []rune(line)
                    for i, ch := range runes {
                        x := indent + i - dx
                        if x >= 0 && x < w {
                            buf.WriteChar(x, screenY, ch, cs.MarkdownCodeBlock)
                        }
                    }
                    return true
                }
                *cur++
            }
        }

    case blockBulletList, blockNumberList, blockCheckList:
        return r.renderListLine(buf, b, depth, lineY, screenY, dx, w, cur, parentBlockStyle)

    case blockBlockquote:
        bqStyle := cs.MarkdownBlockquote
        for i, child := range b.children {
            if i > 0 {
                if *cur == lineY {
                    // Blank line in blockquote — draw bar
                    x := indent - dx
                    if x >= 0 && x < w {
                        buf.WriteChar(x, screenY, '▌', bqStyle)
                    }
                    return true
                }
                *cur++
            }
            if r.renderBlockLine(buf, child, depth+1, lineY, screenY, dx, w, cur, bqStyle) {
                // Draw blockquote bar for this line
                x := indent - dx
                if x >= 0 && x < w {
                    buf.WriteChar(x, screenY, '▌', bqStyle)
                }
                return true
            }
        }

    case blockTable:
        return r.renderTableLine(buf, b, indent, lineY, screenY, dx, w, cur)

    case blockHRule:
        if *cur == lineY {
            for x := 0; x < w; x++ {
                buf.WriteChar(x, screenY, '─', cs.MarkdownHRule)
            }
            return true
        }
        *cur++

    case blockDefList:
        return r.renderDefListLine(buf, b, depth, lineY, screenY, dx, w, cur, parentBlockStyle)
    }

    return false
}

func (r *mdRenderer) headerStyle(level int) tcell.Style {
    if r.cs == nil {
        return tcell.StyleDefault
    }
    switch level {
    case 1: return r.cs.MarkdownH1
    case 2: return r.cs.MarkdownH2
    case 3: return r.cs.MarkdownH3
    case 4: return r.cs.MarkdownH4
    case 5: return r.cs.MarkdownH5
    case 6: return r.cs.MarkdownH6
    }
    return r.cs.MarkdownH6
}

// renderParagraphLine renders a word-wrapped paragraph/header line.
func (r *mdRenderer) renderParagraphLine(buf *DrawBuffer, runs []mdRun, blockStyle tcell.Style, indent int, lineY, screenY, dx, w int, cur *int) bool {
    avail := r.width - indent
    if avail < 1 {
        avail = 1
    }
    var lines [][]mdRun
    if r.wrapText {
        lines = wrapRuns(runs, avail)
    } else {
        lines = [][]mdRun{runs}
    }
    for _, lineRuns := range lines {
        if *cur == lineY {
            x := indent - dx
            for _, run := range lineRuns {
                style := composeStyle(blockStyle, run.style, r.cs)
                for _, ch := range run.text {
                    if x >= 0 && x < w {
                        buf.WriteChar(x, screenY, ch, style)
                    }
                    x++
                }
            }
            return true
        }
        *cur++
    }
    return false
}

// renderListLine renders list items with markers and indentation.
func (r *mdRenderer) renderListLine(buf *DrawBuffer, b mdBlock, depth int, lineY, screenY, dx, w int, cur *int, blockStyle tcell.Style) bool {
    cs := r.cs
    for itemIdx, item := range b.items {
        markerWidth := 4
        indent := depth*2 + markerWidth
        avail := r.width - indent
        if avail < 1 {
            avail = 1
        }

        var lines [][]mdRun
        if r.wrapText {
            lines = wrapRuns(item.runs, avail)
        } else {
            lines = [][]mdRun{item.runs}
        }

        for lineIdx, lineRuns := range lines {
            if *cur == lineY {
                x := depth*2 - dx
                // Draw marker on first line
                if lineIdx == 0 {
                    marker := r.listMarker(b.kind, itemIdx, b.level, item.checked)
                    for _, ch := range marker {
                        if x >= 0 && x < w {
                            buf.WriteChar(x, screenY, ch, cs.MarkdownListMarker)
                        }
                        x++
                    }
                } else {
                    x += markerWidth // continuation indent
                }
                // Draw item text
                for _, run := range lineRuns {
                    style := composeStyle(blockStyle, run.style, cs)
                    for _, ch := range run.text {
                        if x >= 0 && x < w {
                            buf.WriteChar(x, screenY, ch, style)
                        }
                        x++
                    }
                }
                return true
            }
            *cur++
        }

        // Nested children
        if len(item.children) > 0 {
            if r.renderBlocksLine(buf, item.children, depth+1, lineY, screenY, dx, w, cur, blockStyle) {
                return true
            }
        }
    }
    return false
}

func (r *mdRenderer) listMarker(kind mdBlockKind, itemIdx, start int, checked *bool) string {
    switch kind {
    case blockBulletList:
        return "  • "
    case blockNumberList:
        num := start + itemIdx
        if num < 1 {
            num = itemIdx + 1
        }
        s := fmt.Sprintf("%d. ", num)
        for len(s) < 4 {
            s = " " + s
        }
        return s
    case blockCheckList:
        if checked != nil && *checked {
            return "  ☑ "
        }
        return "  ☐ "
    }
    return "  • "
}

// renderTableLine and renderDefListLine follow the same pattern:
// walk rows/items, compute line counts, render when *cur matches lineY.
// Implementation follows the same structure as renderListLine.

func (r *mdRenderer) renderTableLine(buf *DrawBuffer, b mdBlock, indent int, lineY, screenY, dx, w int, cur *int) bool {
    cs := r.cs
    avail := r.width - indent
    colWidths := layoutTable(b, avail)
    if len(colWidths) == 0 {
        return false
    }

    borderStyle := cs.MarkdownTableBorder

    // Top border
    if *cur == lineY {
        r.drawTableBorder(buf, colWidths, '┌', '─', '┬', '┐', indent, screenY, dx, w, borderStyle)
        return true
    }
    *cur++

    // Header row
    if len(b.headers) > 0 {
        headerH := r.tableRowHeight(b.headers, colWidths)
        for row := 0; row < headerH; row++ {
            if *cur == lineY {
                r.drawTableDataRow(buf, b.headers, colWidths, row, indent, screenY, dx, w, cs.MarkdownBold, borderStyle)
                return true
            }
            *cur++
        }
        // Header separator
        if *cur == lineY {
            r.drawTableBorder(buf, colWidths, '├', '─', '┼', '┤', indent, screenY, dx, w, borderStyle)
            return true
        }
        *cur++
    }

    // Data rows
    for _, dataRow := range b.rows {
        rowH := r.tableRowHeight(dataRow, colWidths)
        for row := 0; row < rowH; row++ {
            if *cur == lineY {
                r.drawTableDataRow(buf, dataRow, colWidths, row, indent, screenY, dx, w, cs.MarkdownNormal, borderStyle)
                return true
            }
            *cur++
        }
    }

    // Bottom border
    if *cur == lineY {
        r.drawTableBorder(buf, colWidths, '└', '─', '┴', '┘', indent, screenY, dx, w, borderStyle)
        return true
    }
    *cur++

    return false
}

func (r *mdRenderer) drawTableBorder(buf *DrawBuffer, colWidths []int, left, mid, cross, right rune, indent, screenY, dx, w int, style tcell.Style) {
    x := indent - dx
    writeAt := func(ch rune) {
        if x >= 0 && x < w {
            buf.WriteChar(x, screenY, ch, style)
        }
        x++
    }
    writeAt(left)
    for i, cw := range colWidths {
        for j := 0; j < cw; j++ {
            writeAt(mid)
        }
        if i < len(colWidths)-1 {
            writeAt(cross)
        }
    }
    writeAt(right)
}

func (r *mdRenderer) drawTableDataRow(buf *DrawBuffer, cells [][]mdRun, colWidths []int, subRow, indent, screenY, dx, w int, cellStyle tcell.Style, borderStyle tcell.Style) {
    x := indent - dx
    writeAt := func(ch rune, style tcell.Style) {
        if x >= 0 && x < w {
            buf.WriteChar(x, screenY, ch, style)
        }
        x++
    }
    writeAt('│', borderStyle)
    for i, cw := range colWidths {
        var cellRuns []mdRun
        if i < len(cells) {
            cellRuns = cells[i]
        }
        var lines [][]mdRun
        if r.wrapText {
            lines = wrapRuns(cellRuns, cw)
        } else {
            lines = [][]mdRun{cellRuns}
        }
        var lineRuns []mdRun
        if subRow < len(lines) {
            lineRuns = lines[subRow]
        }
        col := 0
        for _, run := range lineRuns {
            style := composeStyle(cellStyle, run.style, r.cs)
            for _, ch := range run.text {
                if col < cw {
                    writeAt(ch, style)
                    col++
                }
            }
        }
        for col < cw {
            writeAt(' ', cellStyle)
            col++
        }
        writeAt('│', borderStyle)
    }
}

func (r *mdRenderer) renderDefListLine(buf *DrawBuffer, b mdBlock, depth int, lineY, screenY, dx, w int, cur *int, blockStyle tcell.Style) bool {
    cs := r.cs
    for i, item := range b.items {
        if i > 0 {
            if *cur == lineY {
                return true // blank separator
            }
            *cur++
        }
        // Term line
        if *cur == lineY {
            x := depth*2 - dx
            for _, run := range item.term {
                style := cs.MarkdownDefTerm
                for _, ch := range run.text {
                    if x >= 0 && x < w {
                        buf.WriteChar(x, screenY, ch, style)
                    }
                    x++
                }
            }
            return true
        }
        *cur++

        // Definition lines (indented 4)
        defIndent := depth*2 + 4
        defAvail := r.width - defIndent
        if defAvail < 1 {
            defAvail = 1
        }
        var lines [][]mdRun
        if r.wrapText {
            lines = wrapRuns(item.runs, defAvail)
        } else {
            lines = [][]mdRun{item.runs}
        }
        for _, lineRuns := range lines {
            if *cur == lineY {
                x := defIndent - dx
                for _, run := range lineRuns {
                    style := composeStyle(blockStyle, run.style, cs)
                    for _, ch := range run.text {
                        if x >= 0 && x < w {
                            buf.WriteChar(x, screenY, ch, style)
                        }
                        x++
                    }
                }
                return true
            }
            *cur++
        }
    }
    return false
}
```

**Run tests:** `go test ./tv/ -run TestMarkdown -v`

**Commit:** `git commit -m "feat: MarkdownViewer widget — constructor, Draw, HandleEvent, scrolling"`

---

### Task 6: Integration Checkpoint — Full Widget

**Purpose:** Verify that the complete MarkdownViewer widget works end-to-end: parse Markdown, draw to buffer, handle keyboard and mouse events, sync scrollbars.

**Requirements (for test writer):**
- `NewMarkdownViewer` sets SfVisible, OfSelectable, OfFirstClick, wrapText=true
- `SetMarkdown("# Hello\n\nBody text")` populates blocks; `Markdown()` returns the original string
- Drawing a 40×10 MarkdownViewer with `"# Hello"` shows "Hello" at row 0 using `MarkdownH1` style
- Drawing with `"**bold**"` shows "bold" with `MarkdownBold` foreground and `MarkdownNormal` background
- Drawing with a fenced code block shows code lines with `MarkdownCodeBlock` style spanning full width
- Drawing with `"---"` shows `─` characters in `MarkdownHRule` style across the full width
- Drawing with `"- item"` shows `•` in `MarkdownListMarker` style
- Down arrow when focused increments deltaY and consumes the event
- Down arrow when unfocused does nothing and does not consume the event
- PgDn scrolls by viewport height
- Home sets deltaY to 0; End sets deltaY to max
- W key toggles wrapText
- Mouse wheel scrolls ±3 lines
- Binding a vertical scrollbar: scrollbar value tracks deltaY after keyboard scrolling
- SetVScrollBar(nil) unbinds without panic
- SetState(SfSelected, false) hides bound scrollbars

**Components to wire up:** MarkdownViewer, ScrollBar, DrawBuffer, ColorScheme (all real)

**Run:** `go test ./tv/ -run TestIntegration -v`

---

### Task 7: Demo App + E2E Tests

**Files:**
- Modify: `e2e/testapp/basic/main.go` — add Window 5 ("Markdown")
- Modify: `e2e/e2e_test.go` — add e2e tests

**Requirements:**

**Demo app:**
- Window 5 titled "Markdown" at position `(3, 3)` size `(45, 18)` with window number 5
- Contains a MarkdownViewer filling the client area (43×16) with a vertical scrollbar
- Sample Markdown content exercises: H1, H2, H3, paragraph with bold/italic/inline code, fenced code block with language, bullet list, numbered list, checkbox list, blockquote, horizontal rule, table, definition list, link, image alt text, strikethrough
- MarkdownViewer and scrollbar have GrowMode set for resizing
- Window is inserted into Desktop

**E2E tests (requirements for test writer):**
- `TestMarkdownViewerVisible`: switch to Window 5, capture screen, verify header text is visible
- `TestMarkdownViewerCodeBlock`: verify code block content is visible on screen
- `TestMarkdownViewerList`: verify bullet marker `•` is visible
- `TestMarkdownViewerTable`: verify table border character `┌` or `│` is visible
- `TestMarkdownViewerScroll`: press PgDn, verify later content becomes visible (content that was below the fold)
- `TestMarkdownViewerWrapToggle`: press W, verify layout changes (capture before and after — different line counts or positions)

**Implementation for demo app:**

```go
// Window 5 — Markdown viewer
win5 := tv.NewWindow(tv.NewRect(3, 3, 45, 18), "Markdown", tv.WithWindowNumber(5))
mdViewer := tv.NewMarkdownViewer(tv.NewRect(0, 0, 43, 16))
mdViewer.SetMarkdown(`# MarkdownViewer Demo

This is **bold**, *italic*, ***bold italic***, and ~~strikethrough~~.
Here is ` + "`inline code`" + ` in a sentence.
A [link](https://example.com) and an ![image](photo.jpg).

## Code Block

` + "```go" + `
func main() {
    fmt.Println("Hello, TurboView!")
}
` + "```" + `

## Lists

- First **bullet** item
- Second item with ` + "`code`" + `
- Third item

1. Numbered one
2. Numbered two

- [x] Task complete
- [ ] Task pending

## Blockquote

> This is a blockquote with **bold** text.
> It can span multiple lines.

---

## Table

| Name    | Type   | Description         |
|---------|--------|---------------------|
| width   | int    | Widget width        |
| height  | int    | Widget height       |
| visible | bool   | Visibility flag     |

## Definition List

Markdown
: A lightweight markup language

TurboView
: A TUI framework reimplementing Borland Turbo Vision

### H3 Heading
#### H4 Heading
##### H5 Heading
###### H6 Heading
`)

mdVSB := tv.NewScrollBar(tv.NewRect(43, 0, 1, 16), tv.Vertical)
mdViewer.SetVScrollBar(mdVSB)
mdViewer.SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)
mdVSB.SetGrowMode(tv.GfGrowLoX | tv.GfGrowHiX | tv.GfGrowHiY)
win5.Insert(mdViewer)
win5.Insert(mdVSB)
```

Insert win5 into Desktop alongside existing windows:
```go
app.Desktop().Insert(win5)
```

**Run tests:** `go test ./e2e/ -timeout 180s -run TestMarkdown -v`

**Commit:** `git commit -m "feat: add MarkdownViewer to demo app with e2e tests"`

---

## Summary

| Task | What it builds | Key files |
|------|---------------|-----------|
| 1 | ColorScheme fields + BorlandBlue defaults | theme/scheme.go, theme/borland.go |
| 2 | IR types + goldmark parser | tv/markdown_ir.go, go.mod |
| 3 | Rendering helpers — wrap, compose, layout | tv/markdown_render.go |
| 4 | Integration checkpoint — IR + rendering | tests only |
| 5 | MarkdownViewer widget — full implementation | tv/markdown_viewer.go |
| 6 | Integration checkpoint — full widget | tests only |
| 7 | Demo app + e2e tests | e2e/testapp/basic/main.go, e2e/e2e_test.go |
