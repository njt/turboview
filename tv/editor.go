package tv

import "github.com/gdamore/tcell/v2"

type undoState struct {
	lines                                          [][]rune
	cursorRow, cursorCol                           int
	selStartRow, selStartCol, selEndRow, selEndCol int
	deltaX, deltaY                                 int
}

type searchState struct {
	findText        string
	replaceText     string
	caseSensitive   bool
	wholeWords      bool
	promptOnReplace bool
}

type Editor struct {
	*Memo
	undo_      *undoState
	modified   bool
	search     searchState
	filename   string
	lineEnding string
}

func NewEditor(bounds Rect) *Editor {
	ed := &Editor{
		lineEnding: "\n",
	}
	ed.Memo = NewMemo(bounds)
	return ed
}

func (e *Editor) Modified() bool { return e.modified }
func (e *Editor) CanUndo() bool  { return e.undo_ != nil }

func (e *Editor) SetText(s string) {
	e.Memo.SetText(s)
	e.undo_ = nil
	e.modified = false
}

func (e *Editor) saveSnapshot() {
	linesCopy := make([][]rune, len(e.Memo.lines))
	for i, line := range e.Memo.lines {
		linesCopy[i] = make([]rune, len(line))
		copy(linesCopy[i], line)
	}
	e.undo_ = &undoState{
		lines:       linesCopy,
		cursorRow:   e.Memo.cursorRow,
		cursorCol:   e.Memo.cursorCol,
		selStartRow: e.Memo.selStartRow,
		selStartCol: e.Memo.selStartCol,
		selEndRow:   e.Memo.selEndRow,
		selEndCol:   e.Memo.selEndCol,
		deltaX:      e.Memo.deltaX,
		deltaY:      e.Memo.deltaY,
	}
}

func (e *Editor) restoreSnapshot() {
	if e.undo_ == nil {
		return
	}
	u := e.undo_
	e.Memo.lines = u.lines
	e.Memo.cursorRow = u.cursorRow
	e.Memo.cursorCol = u.cursorCol
	e.Memo.selStartRow = u.selStartRow
	e.Memo.selStartCol = u.selStartCol
	e.Memo.selEndRow = u.selEndRow
	e.Memo.selEndCol = u.selEndCol
	e.Memo.deltaX = u.deltaX
	e.Memo.deltaY = u.deltaY
	e.undo_ = nil
	e.Memo.syncScrollBars()
}

func (e *Editor) isEditKey(event *Event) bool {
	if event.What != EvKeyboard || event.Key == nil {
		return false
	}
	k := event.Key
	switch k.Key {
	case tcell.KeyRune, tcell.KeyEnter,
		tcell.KeyBackspace, tcell.KeyBackspace2,
		tcell.KeyDelete, tcell.KeyCtrlX, tcell.KeyCtrlV,
		tcell.KeyCtrlY:
		return true
	}
	return false
}

func (e *Editor) HandleEvent(event *Event) {
	if event.IsCleared() {
		return
	}

	// Handle command events from menus
	if event.What == EvCommand {
		switch event.Command {
		case CmFind:
			e.openFindDialog()
			event.Clear()
			return
		case CmReplace:
			e.openReplaceDialog()
			event.Clear()
			return
		case CmSearchAgain:
			if e.search.findText != "" {
				e.findNext()
			}
			e.broadcastIndicator()
			event.Clear()
			return
		}
	}

	if event.What == EvKeyboard && event.Key != nil {
		k := event.Key
		if k.Key == tcell.KeyCtrlZ {
			e.restoreSnapshot()
			e.broadcastIndicator()
			event.Clear()
			return
		}
		if k.Key == tcell.KeyCtrlF {
			e.openFindDialog()
			event.Clear()
			return
		}
		if k.Key == tcell.KeyCtrlH {
			e.openReplaceDialog()
			event.Clear()
			return
		}
		if k.Key == tcell.KeyF3 || k.Key == tcell.KeyCtrlL {
			if e.search.findText != "" {
				e.findNext()
			}
			e.broadcastIndicator()
			event.Clear()
			return
		}
	}

	isEdit := e.isEditKey(event)
	if isEdit {
		e.saveSnapshot()
	}

	e.Memo.HandleEvent(event)

	if event.IsCleared() {
		if isEdit {
			e.modified = true
		}
		e.broadcastIndicator()
	}
}

func (e *Editor) broadcastIndicator() {
	owner := e.Owner()
	if owner == nil {
		return
	}
	ev := &Event{
		What:    EvBroadcast,
		Command: CmIndicatorUpdate,
		Info:    e,
	}
	for _, child := range owner.Children() {
		child.HandleEvent(ev)
	}
}

// Stubs for find/replace - implemented in find_dialog.go (Task 3)
func (e *Editor) openFindDialog()    {}
func (e *Editor) openReplaceDialog() {}
func (e *Editor) findNext() bool     { return false }
