# Ideas & Future Work

## System Clipboard Integration

Make the app's internal clipboard be the system clipboard, so Ctrl+A / Ctrl+C in any widget (Memo, Editor, MarkdownViewer, etc.) puts content into the OS clipboard. This would let users copy AI responses, code blocks, etc. out of the TUI app.

## MarkdownViewer Search

Add Ctrl+F find to MarkdownViewer (similar to Editor's find dialog). Deferred from v1.

## MarkdownViewer — Streaming Append

`AppendMarkdown(chunk)` for incremental parsing/rendering as AI tokens arrive. Auto-scroll to bottom unless user has scrolled up. Design IR for efficient append from v1.

## MarkdownViewer — Block-Level Keyboard Navigation

Tab/Shift-Tab (or similar) to jump between code blocks, headers, and other significant elements. Combined with clipboard: navigate to a code block, press a key to copy just that block. Navigate to a section header, copy that section.

## MarkdownViewer — Code Block Copy

Hotkey to copy the currently-focused code block's raw text to clipboard. Depends on block-level nav and system clipboard integration.

## MarkdownViewer — Syntax Highlighting

Use chroma or similar to colorize code blocks by language tag. IR already preserves language tag and raw code text; this adds a render-time colorization pass.

## MarkdownViewer — Code Block Line Numbers

Optional line number gutter for code blocks. Toggled by keypress or API.

## MarkdownViewer — Section Folding

Collapse/expand sections by header level, similar to OutlineViewer's collapse behavior.

## MarkdownViewer — Link Following

Store URLs in IR. For file-viewing use cases, a "follow link" command could open in system browser or navigate to another document. Caution: this is the door to being a hypertext browser.
