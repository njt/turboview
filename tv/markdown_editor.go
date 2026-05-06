package tv

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// editKind classifies keyboard edit operations for undo coalescing.
// Consecutive edits of the same kind (editChar or editBackspace) are
// coalesced into a single undo unit. Other edit kinds always start a
// new undo unit and break the coalescing streak.
type editKind int

const (
	editNone      editKind = iota // non-edit key (arrow, click, etc.)
	editChar                      // single rune insert
	editBackspace                 // single char delete (Backspace, Delete without Ctrl)
	editOther                     // Enter, paste, format, etc. — always save
)

// MarkdownEditor is a markdown-aware text editor that combines the editing
// capabilities of Editor with live markdown parsing via parseMarkdown.
type MarkdownEditor struct {
	*Editor
	blocks       []mdBlock
	revealSpans  []revealSpan
	sourceCache  string
	showSource   bool
	lastEditKind editKind
	streakSaved  bool
	undoStack    []undoState
}

// NewMarkdownEditor creates a new MarkdownEditor with the given bounds.
func NewMarkdownEditor(bounds Rect) *MarkdownEditor {
	me := &MarkdownEditor{}
	me.Editor = NewEditor(bounds)
	me.Editor.SetOptions(OfSelectable|OfFirstClick, true)
	me.Editor.SetGrowMode(GfGrowHiX | GfGrowHiY)
	me.Editor.SetSelf(me)
	me.blocks = []mdBlock{}
	me.reparse()
	return me
}

// SetText overrides Editor.SetText: calls the embedded Editor.SetText,
// then calls reparse() to populate blocks.
func (me *MarkdownEditor) SetText(s string) {
	me.Editor.SetText(s)
	me.reparse()
}

// Text delegates to Editor.Text() (inherited from Memo).
func (me *MarkdownEditor) Text() string {
	return me.Editor.Text()
}

// ShowSource returns the current source toggle state.
func (me *MarkdownEditor) ShowSource() bool {
	return me.showSource
}

// SetShowSource sets the source toggle state.
func (me *MarkdownEditor) SetShowSource(on bool) {
	me.showSource = on
}

// reparse joins Memo.lines into a string, runs parseMarkdown, and stores the
// result in blocks. It is a no-op if the source text has not changed since
// the last parse.
func (me *MarkdownEditor) reparse() {
	src := me.Editor.Text()
	if src == me.sourceCache {
		return
	}
	me.sourceCache = src
	if src == "" {
		me.blocks = []mdBlock{}
		me.buildRevealSpans()
		return
	}
	me.blocks = parseMarkdown(src)
	if me.blocks == nil {
		me.blocks = []mdBlock{}
	}
	me.buildRevealSpans()
}

// classifyEvent returns the edit kind for an event, used by undo coalescing.
// Non-keyboard and non-edit keys return editNone.
// Word-boundary runes (space/punctuation after alphanumeric) return editOther
// to force a snapshot boundary.
func (me *MarkdownEditor) classifyEvent(event *Event) editKind {
	if event.What != EvKeyboard || event.Key == nil {
		return editNone
	}
	k := event.Key
	switch k.Key {
	case tcell.KeyRune:
		if me.isWordBoundary(k.Rune) && me.lastEditKind == editChar {
			return editOther
		}
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
	case tcell.KeyEnter:
		// Enter continues the character streak but the post-edit
		// handler resets streakSaved so the next edit starts fresh.
		return editChar
	case tcell.KeyCtrlX, tcell.KeyCtrlV, tcell.KeyCtrlY:
		return editOther
	}
	return editNone
}

// isWordBoundary returns true for runes that constitute a word completion
// boundary: space, tab, or common punctuation that separates words.
func (me *MarkdownEditor) isWordBoundary(r rune) bool {
	if r == ' ' || r == '\t' {
		return true
	}
	if unicode.IsPunct(r) {
		return true
	}
	return false
}

// pushUndo saves the current state onto the undo stack. If a coalescing
// streak is active, this is called only at streak boundaries (first edit
// of a new kind, or explicit boundary like Enter/format).
func (me *MarkdownEditor) pushUndo() {
	linesCopy := make([][]rune, len(me.Memo.lines))
	for i, line := range me.Memo.lines {
		linesCopy[i] = make([]rune, len(line))
		copy(linesCopy[i], line)
	}
	me.undoStack = append(me.undoStack, undoState{
		lines:       linesCopy,
		cursorRow:   me.Memo.cursorRow,
		cursorCol:   me.Memo.cursorCol,
		selStartRow: me.Memo.selStartRow,
		selStartCol: me.Memo.selStartCol,
		selEndRow:   me.Memo.selEndRow,
		selEndCol:   me.Memo.selEndCol,
		deltaX:      me.Memo.deltaX,
		deltaY:      me.Memo.deltaY,
	})
	// Keep Editor.undo_ in sync with top of stack for CanUndo().
	me.Editor.undo_ = &me.undoStack[len(me.undoStack)-1]
}

// popUndo restores the most recent undo state and removes it from the
// stack. Returns false if the stack is empty.
func (me *MarkdownEditor) popUndo() bool {
	if len(me.undoStack) == 0 {
		me.Editor.undo_ = nil
		return false
	}
	u := me.undoStack[len(me.undoStack)-1]
	me.undoStack = me.undoStack[:len(me.undoStack)-1]

	me.Memo.lines = u.lines
	me.Memo.cursorRow = u.cursorRow
	me.Memo.cursorCol = u.cursorCol
	me.Memo.selStartRow = u.selStartRow
	me.Memo.selStartCol = u.selStartCol
	me.Memo.selEndRow = u.selEndRow
	me.Memo.selEndCol = u.selEndCol
	me.Memo.deltaX = u.deltaX
	me.Memo.deltaY = u.deltaY
	me.Memo.syncScrollBars()

	// Update Editor.undo_ to new top (or nil).
	if len(me.undoStack) > 0 {
		me.Editor.undo_ = &me.undoStack[len(me.undoStack)-1]
	} else {
		me.Editor.undo_ = nil
	}
	return true
}

// clearUndoStack clears all undo history. Called by operations like
// SetText that reset the document to a known state.
func (me *MarkdownEditor) clearUndoStack() {
	me.undoStack = me.undoStack[:0]
	me.Editor.undo_ = nil
}
