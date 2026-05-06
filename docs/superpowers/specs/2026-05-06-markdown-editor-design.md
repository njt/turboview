# Markdown Editor — Design Spec

## Overview

Interactive markdown editor widget for the Turbo Vision Go TUI framework. Uses a "reveal-on-cursor" hybrid model: the editor renders markdown as formatted text by default, and syntax markers become visible only when the cursor enters their scope.

The markdown source is the single source of truth. The rendered view is a presentation layer — all edits, undo/redo, copy/paste, and persistence operate on the source.

## Architecture

`MarkdownEditor` embeds `Editor` (which embeds `Memo`). This inherits: cursor management, selection, scrolling, undo snapshots, keyboard input, scrollbar binding, and file I/O.

The editor adds:

- **Parser layer** — re-parses source `[][]rune` to `[]mdBlock` via goldmark on each edit
- **Reveal mapper** — after parsing, computes which syntax markers are "in scope" based on cursor position, producing a set of `revealSpan` records
- **EditorDraw** — a custom `Draw` that walks blocks, renders them formatted, and overlays revealed syntax markers when the cursor is in scope
- **Auto-format** — detects when typed text completes a markdown construct and triggers re-render (source unchanged)
- **Paste handler** — detects paste content type and dispatches to plain/markdown/HTML branches

**Data flow:** Keystroke → Memo mutates source → goldmark parses → reveal mapper annotates with cursor context → Draw renders formatted blocks + revealed markers.

**Cursor model:** Source-coordinate cursor. Memo's `(cursorRow, cursorCol)` are source positions. The draw layer maps these to screen positions. When the cursor is inside a bold span in the source, the renderer reveals the `**` markers. Cursor never jumps during reveal/hide transitions — the source position is unchanged, only the rendering around it changes.

## Syntax Reveal Model

### revealSpan records

After each parse, the reveal mapper produces:

```
revealSpan {
    startRow, startCol  // source position of first marker char
    endRow, endCol      // source position after last marker char
    markerOpen          // text to show at start ("**", "# ", "> ", etc.)
    markerClose         // text to show at end ("**", "", etc.)
    kind                // block or inline
}
```

### Block-level reveal

Triggers when `cursorRow` falls within a block's source lines. The entire block reveals: `# ` prefix on headings, `▌` bar plus `> ` text on blockquotes, `- ` on list items, backtick fences on code blocks, etc. Multi-line blocks reveal as a unit.

### Inline-level reveal

Triggers when `(cursorRow, cursorCol)` falls inside or directly adjacent to an inline span's source range. Only that specific span reveals its markers — `**` for bold, `*` for italic, `` ` `` for code, `~~` for strikethrough. Adjacent means cursor is on the marker character itself or immediately next to it in source. Text shifts to accommodate revealed markers (no space reservation).

### Draw-time

The custom Draw walks blocks line by line using the existing renderer's line-height layout. For each rendered character, checks: is this position inside a revealed span? If so, draw marker characters using a dimmed style. Otherwise, draw the formatted content.

## Typing and Auto-Format

Auto-format triggers when typed syntax becomes complete. Source is NOT mutated — only rendering updates:

| Typed | When | Result |
|---|---|---|
| `# ` at line start | After space | Line renders as heading |
| `**text**` | After closing `**` | Text renders bold |
| `*text*` | After closing `*` | Text renders italic |
| `` `text` `` | After closing `` ` `` | Text renders as code |
| `- ` at line start | After space | Line renders as list item |
| `> ` at line start | After space | Line renders as blockquote |
| ` ``` ` + Enter | After Enter | Opens fenced code block |
| `1. ` at line start | After space | Line renders as numbered list |

## Smart List Continuation

These DO mutate source:

- **Enter** at end of a non-empty list item → new line with same marker type. Bullet and checklist markers are copied verbatim (`- `, `- [ ] `). Numbered list markers are incremented (`1. ` → `2. `, `99. ` → `100. `).
- **Enter** on an empty list item → delete empty marker, exit list, insert blank line
- **Tab** at list item → indent (add `  ` before marker, e.g., `- ` → `  - `)
- **Shift-Tab** at indented list item → outdent (remove `  ` prefix)

## Keyboard Shortcuts

| Shortcut | Action |
|---|---|
| Ctrl+B | Toggle `**bold**` on selection or insert empty markers |
| Ctrl+I | Toggle `*italic*` on selection or insert empty markers |
| Ctrl+K | Open link dialog (create/edit/remove) |
| Ctrl+T | Toggle raw source view |

Toggle behavior: with selection, wraps/unwraps. Without selection, inserts empty markers and places cursor between them.

## Link Interaction

Links render as formatted text (green, underlined). When cursor is on link text:

- Status line hints available action ("Enter to edit link")
- Pressing Enter opens sequential InputBox dialogs for link text and URL (standard Turbo Vision pattern), with OK/Cancel on each step. Clearing the URL field and pressing OK removes the link (keeps the text).
- Ctrl+K with a selection opens the same dialog to create a new link

## Paste Behavior

Three branches based on clipboard content:

1. **Plain text** — inserted verbatim, no markdown interpretation
2. **Markdown text** — detected by presence of markdown syntax or `text/markdown` MIME. Inserted as source, rendered accordingly
3. **Rich text / HTML** — converted to markdown before insertion (headings → `#`, bold → `**`, italic → `*`, lists → `-`, links → `[text](url)`). Platform-limited in a TUI; degrades gracefully

`Ctrl+Shift+V` forces "paste as plain text" regardless of type.

## Undo/Redo

Reuses `Editor`'s existing snapshot mechanism with two additions:

- Snapshots fire at meaningful boundaries: word completion (space/punctuation after alphanumerics), Enter, format command, paste, delete-word
- Consecutive single-character inserts/deletes coalesce into one undo unit
- Undo restores both source and cursor position (already handled by Editor's snapshot)

## Source Toggle

Per-document `showSource` bool, toggled by `Ctrl+T` and a `View > Show Source` menu item. When active, `Draw` delegates entirely to `Memo.Draw` — raw markdown text, no formatting, no reveal. Source and cursor shared between modes; toggling doesn't lose state.

## Rendering Scope

CommonMark core plus GFM extensions via goldmark (already integrated): tables, task lists, strikethrough, fenced code blocks with language tags, autolinks, definition lists. Math and footnotes deferred.

## Performance

Viewport-only rendering with overscan buffer. The existing `mdRenderer` already computes line-level height and renders line-by-line. For editing, bounded to `[deltaY - buffer, deltaY + viewportHeight + buffer]`. Source stays fully in memory. Goldmark re-parses full source on each edit — fast enough (< 1ms) for documents under ~10K lines.

## Files

All in `tv/`:

| File | Purpose |
|---|---|
| `markdown_editor.go` | `MarkdownEditor` struct, constructor, public API, embed of Editor |
| `markdown_editor_reveal.go` | Reveal mapper, `revealSpan`, cursor-in-scope checks, block/inline reveal logic |
| `markdown_editor_draw.go` | Custom `Draw` — renders blocks with syntax reveal overlay |
| `markdown_editor_autoformat.go` | Auto-format detection, smart list continuation, keyboard format shortcuts |
| `markdown_editor_paste.go` | Paste detection and dispatch |
| `markdown_editor_test.go` | Unit tests |
| `markdown_editor_integration_test.go` | Integration/e2e tests |

## Theme

No new theme styles required — existing 17 `Markdown*` styles cover all block and inline types. Revealed syntax markers use a dimmed style; add `MarkdownSyntaxMarker` if existing styles aren't distinct enough.

## What the Editor Does Not Do

- No autocorrect, autocapitalize, or smart-quote (silently corrupts source)
- No hiding of deliberately-typed markdown (escape sequences like `\*` must round-trip)
- No separate "rich" representation that could drift from source
- No lossy conversion of pasted markdown (preserves `*` vs `_` choice)
