package tv

import "strings"

// linkSpan describes a markdown link found in source text:
// [text](url) where text is the displayed content and url is
// the target destination.
type linkSpan struct {
	row, col           int    // start of link text in source
	text, url          string // extracted content
	fullStart, fullEnd int    // positions of entire [text](url) in source line
}

// findLinkAt scans the source line at row for [text](url) patterns.
// Returns a linkSpan if col falls within the link text portion
// (between '[' and ']'), nil otherwise.
func (me *MarkdownEditor) findLinkAt(row, col int) *linkSpan {
	if row < 0 || row >= len(me.Memo.lines) {
		return nil
	}
	line := string(me.Memo.lines[row])
	for i := 0; i < len(line); i++ {
		if line[i] == '[' {
			// Check for escaped bracket
			if i > 0 && line[i-1] == '\\' {
				continue
			}
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
			// col is in link text: after '[' and before ']'
			if col > i && col < closeBracket {
				return &linkSpan{
					row:       row,
					col:       i + 1,
					text:      line[i+1 : closeBracket],
					url:       line[closeBracket+2 : closeParen],
					fullStart: i,
					fullEnd:   closeParen + 1,
				}
			}
			i = closeParen
		}
	}
	return nil
}

// openLinkDialog opens a modal dialog to create or edit a markdown link.
// Uses InputBox for text and URL fields. When the cursor is on an existing
// link, the dialog pre-fills with the current values. When a selection is
// active, the selected text is used as the link text.
func (me *MarkdownEditor) openLinkDialog() {
	desktop := findDesktop(me)
	if desktop == nil {
		return // No desktop available (unit test environment)
	}

	linkText := ""
	linkURL := ""
	var existingLink *linkSpan

	r, c := me.Memo.cursorRow, me.Memo.cursorCol
	existingLink = me.findLinkAt(r, c)
	if existingLink != nil {
		linkText = existingLink.text
		linkURL = existingLink.url
	} else if me.Memo.HasSelection() {
		linkText = me.Memo.selectedText()
	}

	text, cmd := InputBox(desktop, "Link Text", "~T~ext:", linkText)
	if cmd != CmOK {
		return
	}
	url, cmd := InputBox(desktop, "Link URL", "~U~RL:", linkURL)
	if cmd != CmOK {
		return
	}

	// Save undo snapshot before source mutation
	me.pushUndo()
	me.streakSaved = false
	me.lastEditKind = editOther

	// Replace existing link
	if existingLink != nil {
		line := me.Memo.lines[existingLink.row]
		before := string(line[:existingLink.fullStart])
		after := string(line[existingLink.fullEnd:])
		var replacement string
		if url != "" {
			replacement = "[" + text + "](" + url + ")"
		} else {
			replacement = text
		}
		me.Memo.lines[existingLink.row] = []rune(before + replacement + after)
		me.Memo.cursorRow = existingLink.row
		me.Memo.cursorCol = existingLink.fullStart + len(replacement)
		me.reparse()
		me.updateLinkIndicator()
		return
	}

	// Create new link
	if me.Memo.HasSelection() {
		me.Memo.deleteSelection()
	}
	if url != "" {
		me.Memo.insertText("[" + text + "](" + url + ")")
	} else if text != "" {
		me.Memo.insertText(text)
	}
	me.reparse()
	me.updateLinkIndicator()
}

// updateLinkIndicator broadcasts an indicator update via the Editor's
// broadcastIndicator mechanism. This allows status-line widgets to display
// link-related hints (e.g., "Enter to edit link") when the cursor is on a link.
func (me *MarkdownEditor) updateLinkIndicator() {
	me.Editor.broadcastIndicator()
}
