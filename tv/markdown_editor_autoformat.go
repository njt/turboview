package tv

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// handleListIndent intercepts Tab/Shift-Tab to indent/outdent list items.
// Called BEFORE Editor processes the event in HandleEvent.
// Returns true if the event was consumed.
func (me *MarkdownEditor) handleListIndent(event *Event) bool {
	if event.What != EvKeyboard || event.Key == nil {
		return false
	}
	k := event.Key
	if k.Key == tcell.KeyTab && k.Modifiers == tcell.ModNone {
		return me.listIndent()
	}
	if k.Key == tcell.KeyTab && k.Modifiers&tcell.ModShift != 0 {
		return me.listOutdent()
	}
	return false
}

// listIndent adds a "  " prefix to the current line if it is a list item.
// Returns true and clears the event if indentation was performed.
func (me *MarkdownEditor) listIndent() bool {
	me.pushUndo()
	me.streakSaved = false
	me.lastEditKind = editOther
	row := me.Memo.cursorRow
	if row < 0 || row >= len(me.Memo.lines) {
		return false
	}
	line := string(me.Memo.lines[row])
	_, _, isListItem := detectListMarker(line)
	if !isListItem {
		return false
	}
	// Insert "  " before the line
	me.Memo.lines[row] = []rune("  " + line)
	me.Memo.cursorCol += 2
	return true
}

// listOutdent removes a "  " prefix from the current line if it is an indented
// list item. If the line is not indented or not a list item, it is a no-op.
// Returns true and clears the event if outdentation was performed.
func (me *MarkdownEditor) listOutdent() bool {
	me.pushUndo()
	me.streakSaved = false
	me.lastEditKind = editOther
	row := me.Memo.cursorRow
	if row < 0 || row >= len(me.Memo.lines) {
		return false
	}
	line := string(me.Memo.lines[row])
	if !strings.HasPrefix(line, "  ") {
		return false
	}
	// Only outdent if the indented line is itself a list item
	trimmed := line[2:]
	_, _, isListItem := detectListMarker(trimmed)
	if !isListItem {
		return false
	}
	me.Memo.lines[row] = []rune(trimmed)
	if me.Memo.cursorCol >= 2 {
		me.Memo.cursorCol -= 2
	} else {
		me.Memo.cursorCol = 0
	}
	return true
}

// listEnterContinuation handles Enter for smart list continuation.
// Called AFTER Editor has processed Enter (new line created, cursor moved).
// Returns true if source was mutated (caller must reparse).
func (me *MarkdownEditor) listEnterContinuation() bool {
	lines := me.Memo.lines
	row := me.Memo.cursorRow // Current row (the new line created by Enter)
	if row == 0 {
		return false
	}
	prevLine := string(lines[row-1])
	curLine := string(lines[row])

	marker, rest, isListItem := detectListMarker(prevLine)
	if !isListItem {
		return false
	}

	// If previous line was an empty list item (marker only, no content),
	// clear the marker to exit the list.
	if strings.TrimSpace(rest) == "" {
		lines[row-1] = []rune{}
		return true
	}

	// If current line is non-empty, don't interfere (user typed content)
	if strings.TrimSpace(curLine) != "" {
		return false
	}

	// Insert marker on the current (new) line
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

// isNumberedList returns true if the marker is a numbered list marker
// (starting with a digit, e.g., "1. " or "99. ").
func isNumberedList(marker string) bool {
	trimmed := strings.TrimLeft(marker, " ")
	if len(trimmed) < 3 || trimmed[0] < '0' || trimmed[0] > '9' {
		return false
	}
	dotIdx := strings.Index(trimmed, ". ")
	return dotIdx > 0
}

// incrementListNumber parses the number in the marker and returns a new marker
// with the number incremented by 1. Preserves leading indent.
// Returns the original marker unchanged if parsing fails.
func incrementListNumber(marker string) string {
	indent := ""
	rest := marker
	for strings.HasPrefix(rest, "  ") {
		indent += "  "
		rest = rest[2:]
	}
	numEnd := strings.Index(rest, ". ")
	if numEnd < 0 {
		return marker
	}
	num, err := strconv.Atoi(rest[:numEnd])
	if err != nil {
		return marker
	}
	return indent + strconv.Itoa(num+1) + ". "
}

// isChecklist returns true if the marker contains checklist syntax
// (both '[' and ']' characters, e.g., "- [ ] " or "- [x] ").
func isChecklist(marker string) bool {
	trimmed := strings.TrimLeft(marker, " ")
	return strings.Contains(trimmed, "[ ]") || strings.Contains(trimmed, "[x]") || strings.Contains(trimmed, "[X]")
}

// uncheckedVersion returns the unchecked variant of a checklist marker.
// Always returns "- [ ] " regardless of whether the original is checked or not,
// preserving any leading indent.
func uncheckedVersion(marker string) string {
	indent := ""
	rest := marker
	for strings.HasPrefix(rest, "  ") {
		indent += "  "
		rest = rest[2:]
	}
	return indent + "- [ ] "
}

// toggleFormat toggles inline formatting markers on the selection or at the
// cursor position. With a selection: wraps/unwraps the selected text. Without
// a selection: inserts empty markers and places the cursor between them.
// Also checks the context surrounding the selection for existing markers
// (to handle cases where the user selects only the inner content of an
// already-formatted span).
func (me *MarkdownEditor) toggleFormat(marker string) {
	me.pushUndo()
	me.streakSaved = false
	me.lastEditKind = editOther
	markerLen := len([]rune(marker))
	if me.Memo.HasSelection() {
		sel := me.Memo.selectedText()

		// Check if selected text itself is wrapped in markers
		if strings.HasPrefix(sel, marker) && strings.HasSuffix(sel, marker) {
			unwrapped := sel[markerLen : len(sel)-markerLen]
			me.Memo.deleteSelection()
			me.Memo.insertText(unwrapped)
			me.reparse()
			return
		}

		// Check surrounding context for markers (single-line selections only)
		sr, sc, er, ec := me.Memo.normalizedSelection()
		if sr == er {
			line := me.Memo.lines[sr]
			if sc >= markerLen && ec+markerLen <= len(line) {
				before := string(line[sc-markerLen : sc])
				after := string(line[ec : ec+markerLen])
				if before == marker && after == marker {
					// Expand selection to include surrounding markers,
					// then delete and insert the inner text only.
					me.Memo.selStartCol = sc - markerLen
					me.Memo.selEndCol = ec + markerLen
					me.Memo.deleteSelection()
					me.Memo.insertText(sel)
					me.reparse()
					return
				}
			}
		}

		// Otherwise, wrap the selection
		wrapped := marker + sel + marker
		me.Memo.deleteSelection()
		me.Memo.insertText(wrapped)
	} else {
		// No selection: insert empty markers
		me.Memo.insertText(marker + marker)
		me.Memo.cursorCol -= markerLen
	}
	me.reparse()
}
