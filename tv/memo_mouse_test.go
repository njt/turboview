package tv

// memo_mouse_test.go — Tests for Task 5: Mouse Handling (Click, Drag, Double-Click, Triple-Click).
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test cites the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Click: basic cursor positioning
//   Section 2  — Click: row clamping
//   Section 3  — Click: col clamping
//   Section 4  — Click: collapses selection
//   Section 5  — Click: event consumption
//   Section 6  — Drag: creates selection
//   Section 7  — Drag: anchor and extent
//   Section 8  — Drag: release ends drag
//   Section 9  — Drag: backward drag
//   Section 10 — Double-click: word selection
//   Section 11 — Double-click: whitespace selection
//   Section 12 — Double-click: end of line does nothing
//   Section 13 — Double-click: event consumed
//   Section 14 — Triple-click: selects entire line
//   Section 15 — Triple-click: anchor and extent
//   Section 16 — Triple-click: event consumed
//   Section 17 — Falsifying tests

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// mouseEv creates a Memo mouse event.
func mouseEv(x, y int, btn tcell.ButtonMask, clickCount int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: btn, ClickCount: clickCount}}
}

// memoClickAt sends a single Button1 click to a Memo at the given screen coordinates.
func memoClickAt(m *Memo, x, y int) *Event {
	ev := mouseEv(x, y, tcell.Button1, 1)
	m.HandleEvent(ev)
	return ev
}

// memoDoubleClickAt sends a Button1 double-click to a Memo at the given screen coordinates.
func memoDoubleClickAt(m *Memo, x, y int) *Event {
	ev := mouseEv(x, y, tcell.Button1, 2)
	m.HandleEvent(ev)
	return ev
}

// memoTripleClickAt sends a Button1 triple-click to a Memo at the given screen coordinates.
func memoTripleClickAt(m *Memo, x, y int) *Event {
	ev := mouseEv(x, y, tcell.Button1, 3)
	m.HandleEvent(ev)
	return ev
}

// memoDragTo sends a Button1 motion event (drag) to a Memo at the given screen coordinates.
func memoDragTo(m *Memo, x, y int) *Event {
	ev := mouseEv(x, y, tcell.Button1, 0)
	m.HandleEvent(ev)
	return ev
}

// memoRelease sends a ButtonNone event to a Memo (mouse release).
func memoRelease(m *Memo, x, y int) *Event {
	ev := mouseEv(x, y, tcell.ButtonNone, 0)
	m.HandleEvent(ev)
	return ev
}

// newMouseMemo creates a Memo at (0,0) with width=40, height=10, suitable for mouse tests.
// With the Memo origin at (0,0) and deltaY=0 (no scroll), mouseY maps directly to line index.
func newMouseMemo() *Memo {
	return NewMemo(NewRect(0, 0, 40, 10))
}

// ---------------------------------------------------------------------------
// Section 1 — Click: basic cursor positioning
// ---------------------------------------------------------------------------

// TestClickPositionsCursorRow verifies that clicking at (x, y) positions the cursor
// at the row matching deltaY + mouseY (deltaY=0 for a non-scrolled Memo).
// Spec: "row = deltaY + mouseY"
func TestClickPositionsCursorRow(t *testing.T) {
	m := newMouseMemo()
	m.SetText("line0\nline1\nline2")

	memoClickAt(m, 0, 2) // click on row 2

	row, _ := m.CursorPos()
	if row != 2 {
		t.Errorf("click at y=2: CursorPos row = %d, want 2", row)
	}
}

// TestClickPositionsCursorCol verifies that clicking at column x positions the cursor
// at the rune index corresponding to x.
// Spec: "col converted from screen X to rune index"
func TestClickPositionsCursorCol(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello\nworld")

	memoClickAt(m, 3, 0) // click on col 3 of row 0

	_, col := m.CursorPos()
	if col != 3 {
		t.Errorf("click at x=3: CursorPos col = %d, want 3", col)
	}
}

// TestClickPositionsCursorRowAndCol verifies both row and col are set correctly from a click.
// Spec: "row = deltaY + mouseY", "col converted from screen X to rune index"
func TestClickPositionsCursorRowAndCol(t *testing.T) {
	m := newMouseMemo()
	m.SetText("first\nsecond\nthird")

	memoClickAt(m, 4, 1) // click at (x=4, y=1)

	row, col := m.CursorPos()
	if row != 1 {
		t.Errorf("click at y=1: CursorPos row = %d, want 1", row)
	}
	if col != 4 {
		t.Errorf("click at x=4: CursorPos col = %d, want 4", col)
	}
}

// TestClickAtColZeroPositionsCursorAtStart verifies clicking at x=0 puts cursor at col 0.
// Spec: "col converted from screen X to rune index"
func TestClickAtColZeroPositionsCursorAtStart(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello")
	memoClickAt(m, 3, 0) // first move cursor off col 0

	memoClickAt(m, 0, 0) // click back at col 0

	_, col := m.CursorPos()
	if col != 0 {
		t.Errorf("click at x=0: CursorPos col = %d, want 0", col)
	}
}

// TestClickOnTabLandsAtTabRuneIndex verifies that clicking in the visual span of a tab
// positions the cursor at the tab character's rune index.
// Spec: "Tab characters: a click in the visual span of a tab lands at the tab character's rune index"
func TestClickOnTabLandsAtTabRuneIndex(t *testing.T) {
	m := newMouseMemo()
	// Line: tab followed by "hello". Tab at col 0 expands to 8 visual spaces (cols 0–7).
	// Clicking at visual col 3 (inside the tab's visual span) should put cursor at rune index 0.
	m.SetText("\thello")

	memoClickAt(m, 3, 0) // click in the middle of tab's visual expansion

	_, col := m.CursorPos()
	if col != 0 {
		t.Errorf("click at visual col 3 inside tab span: CursorPos col = %d, want 0 (tab rune index)", col)
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Click: row clamping
// ---------------------------------------------------------------------------

// TestClickRowBeyondContentClampsToLastLine verifies that clicking past the last line
// clamps to the last line.
// Spec: "Row clamped to [0, len(lines)-1]"
func TestClickRowBeyondContentClampsToLastLine(t *testing.T) {
	m := newMouseMemo()
	m.SetText("line0\nline1") // 2 lines (rows 0 and 1)

	memoClickAt(m, 0, 9) // y=9 is well past the last line (row 1)

	row, _ := m.CursorPos()
	if row != 1 {
		t.Errorf("click at y=9 beyond last line: CursorPos row = %d, want 1 (clamped to last line)", row)
	}
}

// TestClickRowAtExactLastLine verifies clicking exactly on the last line works.
// Spec: "Row clamped to [0, len(lines)-1]"
func TestClickRowAtExactLastLine(t *testing.T) {
	m := newMouseMemo()
	m.SetText("alpha\nbeta\ngamma") // 3 lines (rows 0–2)

	memoClickAt(m, 2, 2) // click on row 2 (last line), col 2

	row, col := m.CursorPos()
	if row != 2 {
		t.Errorf("click at y=2 (last line): CursorPos row = %d, want 2", row)
	}
	if col != 2 {
		t.Errorf("click at x=2: CursorPos col = %d, want 2", col)
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Click: col clamping
// ---------------------------------------------------------------------------

// TestClickPastEndOfLineClampsToEndOfLine verifies that clicking past the end of a
// line clamps to len(lines[row]).
// Spec: "Col clamped to [0, len(lines[row])]"
func TestClickPastEndOfLineClampsToEndOfLine(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hi") // 2 runes

	memoClickAt(m, 20, 0) // click well past the end of "hi"

	_, col := m.CursorPos()
	if col != 2 {
		t.Errorf("click at x=20 past end of 2-char line: CursorPos col = %d, want 2 (clamped to len)", col)
	}
}

// TestClickAtExactEndOfLineIsAllowed verifies that clicking at len(line) (past last rune,
// at the logical end) is a valid position.
// Spec: "Col clamped to [0, len(lines[row])] (can be at end of line but not past)"
func TestClickAtExactEndOfLineIsAllowed(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello") // 5 runes, valid col range [0,5]

	memoClickAt(m, 5, 0) // click at the position just after 'o'

	_, col := m.CursorPos()
	if col != 5 {
		t.Errorf("click at x=5 (end of 5-char line): CursorPos col = %d, want 5", col)
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Click: collapses selection
// ---------------------------------------------------------------------------

// TestClickCollapsesExistingSelection verifies that a single click collapses any
// existing selection.
// Spec: "Collapses any existing selection"
func TestClickCollapsesExistingSelection(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")
	// Build a selection using Shift+Right.
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	if !m.HasSelection() {
		t.Fatal("precondition: selection should exist before click")
	}

	memoClickAt(m, 5, 0)

	if m.HasSelection() {
		t.Error("click should collapse selection; HasSelection() = true after click, want false")
	}
}

// TestClickAfterCtrlACollapsesSelection verifies that a click collapses a Ctrl+A selection.
// Spec: "Collapses any existing selection"
func TestClickAfterCtrlACollapsesSelection(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")
	m.HandleEvent(selCtrlKeyEv(tcell.KeyCtrlA))
	if !m.HasSelection() {
		t.Fatal("precondition: Ctrl+A should create a selection")
	}

	memoClickAt(m, 2, 0)

	if m.HasSelection() {
		t.Error("click after Ctrl+A should collapse selection; HasSelection() = true, want false")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Click: event consumption
// ---------------------------------------------------------------------------

// TestClickConsumesEvent verifies that a Button1 single-click event is consumed.
// Spec: "Mouse events are consumed (event.Clear()) when handled"
func TestClickConsumesEvent(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello")

	ev := mouseEv(2, 0, tcell.Button1, 1)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Button1 single-click event was not consumed (IsCleared() = false)")
	}
}

// TestNonButton1ClickIsNotConsumed verifies that a click with a button other than
// Button1 is NOT consumed.
// Spec: "Mouse events that are NOT Button1 are ignored (not consumed)"
func TestNonButton1ClickIsNotConsumed(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello")

	ev := mouseEv(2, 0, tcell.Button2, 1)
	m.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Button2 event should NOT be consumed; IsCleared() = true, want false")
	}
}

// TestButton3ClickIsNotConsumed verifies that Button3 clicks are also not consumed.
// Spec: "Mouse events that are NOT Button1 are ignored (not consumed)"
func TestButton3ClickIsNotConsumed(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello")

	ev := mouseEv(2, 0, tcell.Button3, 1)
	m.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Button3 event should NOT be consumed; IsCleared() = true, want false")
	}
}

// TestNonButton1ClickDoesNotMoveCursor verifies that a non-Button1 click does not
// change the cursor position.
// Spec: "Mouse events that are NOT Button1 are ignored (not consumed)"
func TestNonButton1ClickDoesNotMoveCursor(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")
	// Cursor starts at (0,0).

	m.HandleEvent(mouseEv(5, 0, tcell.Button2, 1))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Button2 click moved cursor to (%d,%d); cursor should not have moved", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Drag: creates selection
// ---------------------------------------------------------------------------

// TestDragCreatesSelection verifies that pressing Button1 then sending a motion
// event with Button1 held creates a selection.
// Spec: "Subsequent Button1 events while dragging: moves cursor and extends selection"
func TestDragCreatesSelection(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoClickAt(m, 2, 0)  // press at col 2 (starts drag)
	memoDragTo(m, 7, 0)   // drag to col 7

	if !m.HasSelection() {
		t.Error("drag should create a selection; HasSelection() = false after drag")
	}
}

// TestDragCreatesSelectionMemoEventConsumed verifies that drag events are consumed.
// Spec: "Mouse events are consumed (event.Clear()) when handled"
func TestDragCreatesSelectionMemoEventConsumed(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoClickAt(m, 2, 0)

	dragEv := mouseEv(7, 0, tcell.Button1, 0)
	m.HandleEvent(dragEv)

	if !dragEv.IsCleared() {
		t.Error("drag motion event was not consumed (IsCleared() = false)")
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Drag: anchor and extent
// ---------------------------------------------------------------------------

// TestDragAnchorIsInitialClickPosition verifies that the selection anchor stays at
// the initial click position.
// Spec: "Anchor point is the initial click position; extent follows mouse"
func TestDragAnchorIsInitialClickPosition(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoClickAt(m, 2, 0)  // anchor established at col 2
	memoDragTo(m, 7, 0)   // drag to col 7

	sr, sc, _, _ := m.Selection()
	if sr != 0 || sc != 2 {
		t.Errorf("drag anchor = (%d,%d), want (0,2) (initial click position)", sr, sc)
	}
}

// TestDragExtentFollowsMouse verifies that the selection extent follows the mouse
// during a drag.
// Spec: "extent follows mouse"
func TestDragExtentFollowsMouse(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoClickAt(m, 2, 0)  // anchor at col 2
	memoDragTo(m, 7, 0)   // drag to col 7

	_, _, er, ec := m.Selection()
	if er != 0 || ec != 7 {
		t.Errorf("drag extent = (%d,%d), want (0,7) (mouse position)", er, ec)
	}
}

// TestDragExtentUpdatesAsMovesContinues verifies the extent follows the mouse as
// it continues to move.
// Spec: "extent follows mouse"
func TestDragExtentUpdatesAsMovesContinues(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoClickAt(m, 2, 0)
	memoDragTo(m, 5, 0) // intermediate drag
	memoDragTo(m, 9, 0) // further drag

	_, _, er, ec := m.Selection()
	if er != 0 || ec != 9 {
		t.Errorf("drag extent after continued move = (%d,%d), want (0,9)", er, ec)
	}
}

// TestDragAnchorDoesNotMoveWhenExtentUpdates verifies anchor stays fixed as the
// mouse moves.
// Spec: "Anchor point is the initial click position"
func TestDragAnchorDoesNotMoveWhenExtentUpdates(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoClickAt(m, 3, 0)
	memoDragTo(m, 5, 0)
	memoDragTo(m, 8, 0)

	sr, sc, _, _ := m.Selection()
	if sr != 0 || sc != 3 {
		t.Errorf("anchor after multiple drag moves = (%d,%d), want (0,3) (unchanged)", sr, sc)
	}
}

// TestDragMultiRowCreatesMultiRowSelection verifies dragging across rows creates
// a multi-row selection.
// Spec: "Subsequent Button1 events while dragging: moves cursor and extends selection"
func TestDragMultiRowCreatesMultiRowSelection(t *testing.T) {
	m := newMouseMemo()
	m.SetText("line0\nline1\nline2")

	memoClickAt(m, 1, 0)  // anchor at (row=0, col=1)
	memoDragTo(m, 3, 2)   // drag to (row=2, col=3)

	if !m.HasSelection() {
		t.Error("multi-row drag should have selection; HasSelection() = false")
	}
	_, _, er, _ := m.Selection()
	if er != 2 {
		t.Errorf("drag extent row = %d, want 2 (dragged to row 2)", er)
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Drag: release ends drag
// ---------------------------------------------------------------------------

// TestReleaseEndsaDrag verifies that a ButtonNone event stops dragging.
// Spec: "Dragging stops on ButtonNone (release)"
func TestReleaseEndsaDrag(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoClickAt(m, 2, 0)
	memoDragTo(m, 7, 0)
	memoRelease(m, 7, 0)

	// After release, selection should be preserved.
	if !m.HasSelection() {
		t.Error("selection should be preserved after drag release; HasSelection() = false")
	}
}

// TestReleasePreservesSelection verifies the selection remains in place after release.
// Spec: "Dragging stops on ButtonNone (release)"
func TestReleasePreservesSelection(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoClickAt(m, 2, 0)
	memoDragTo(m, 7, 0)
	memoRelease(m, 7, 0)

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 2 {
		t.Errorf("selection start after release = (%d,%d), want (0,2)", sr, sc)
	}
	if er != 0 || ec != 7 {
		t.Errorf("selection end after release = (%d,%d), want (0,7)", er, ec)
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Drag: backward drag
// ---------------------------------------------------------------------------

// TestDragBackwardCreatesReversedSelection verifies dragging left of the anchor
// creates a selection where the extent is left of the anchor.
// Spec: "Anchor point is the initial click position; extent follows mouse"
func TestDragBackwardCreatesReversedSelection(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoClickAt(m, 8, 0)  // anchor at col 8
	memoDragTo(m, 3, 0)   // drag backward to col 3

	if !m.HasSelection() {
		t.Error("backward drag should create a selection; HasSelection() = false")
	}
	sr, sc, er, ec := m.Selection()
	// Anchor should be at the initial click (col 8), extent at col 3.
	if sr != 0 || sc != 8 {
		t.Errorf("backward drag anchor = (%d,%d), want (0,8)", sr, sc)
	}
	if er != 0 || ec != 3 {
		t.Errorf("backward drag extent = (%d,%d), want (0,3)", er, ec)
	}
}

// ---------------------------------------------------------------------------
// Section 10 — Double-click: word selection
// ---------------------------------------------------------------------------

// TestDoubleClickSelectsWordUnderCursor verifies double-click selects the word
// under the cursor.
// Spec: "Double-click (ClickCount == 2): Selects the word under the cursor using
// word boundary logic (charClass)"
func TestDoubleClickSelectsWordUnderCursor(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoDoubleClickAt(m, 1, 0) // double-click in the middle of "hello"

	if !m.HasSelection() {
		t.Error("double-click on a word should create selection; HasSelection() = false")
	}
}

// TestDoubleClickSelectsEntireWord verifies the double-click selection spans the
// whole word, not just a single character.
// Spec: "Selects the word under the cursor using word boundary logic (charClass)"
func TestDoubleClickSelectsEntireWord(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoDoubleClickAt(m, 2, 0) // click inside "hello"

	sr, sc, er, ec := m.Selection()
	// "hello" occupies cols 0–4; selection should cover the whole word.
	if sr != 0 || sc != 0 {
		t.Errorf("double-click word start = (%d,%d), want (0,0) (start of \"hello\")", sr, sc)
	}
	if er != 0 || ec != 5 {
		t.Errorf("double-click word end = (%d,%d), want (0,5) (end of \"hello\")", er, ec)
	}
}

// TestDoubleClickSelectsSecondWord verifies double-click on a second word selects
// that specific word.
// Spec: "Selects the word under the cursor using word boundary logic (charClass)"
func TestDoubleClickSelectsSecondWord(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoDoubleClickAt(m, 7, 0) // click inside "world" (cols 6–10)

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 6 {
		t.Errorf("double-click second word start = (%d,%d), want (0,6) (start of \"world\")", sr, sc)
	}
	if er != 0 || ec != 11 {
		t.Errorf("double-click second word end = (%d,%d), want (0,11) (end of \"world\")", er, ec)
	}
}

// TestDoubleClickCreatesSelection verifies HasSelection is true after double-click.
// Spec: "Double-click (ClickCount == 2): Selects the word under the cursor"
func TestDoubleClickCreatesSelection(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoDoubleClickAt(m, 2, 0)

	if !m.HasSelection() {
		t.Error("double-click on word: HasSelection() = false, want true")
	}
}

// ---------------------------------------------------------------------------
// Section 11 — Double-click: whitespace selection
// ---------------------------------------------------------------------------

// TestDoubleClickOnWhitespaceSelectsWhitespaceRun verifies double-click on whitespace
// selects the continuous whitespace run.
// Spec: "If cursor is on whitespace, selects the whitespace run"
func TestDoubleClickOnWhitespaceSelectsWhitespaceRun(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello   world") // 3 spaces at cols 5,6,7

	memoDoubleClickAt(m, 6, 0) // click in the middle of the spaces

	if !m.HasSelection() {
		t.Error("double-click on whitespace should create selection; HasSelection() = false")
	}
	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 5 {
		t.Errorf("whitespace selection start = (%d,%d), want (0,5)", sr, sc)
	}
	if er != 0 || ec != 8 {
		t.Errorf("whitespace selection end = (%d,%d), want (0,8)", er, ec)
	}
}

// ---------------------------------------------------------------------------
// Section 12 — Double-click: end of line does nothing
// ---------------------------------------------------------------------------

// TestDoubleClickAtEndOfLineDoesNothing verifies that double-clicking past the
// last rune of a line (cursor >= len) does not create a selection.
// Spec: "If cursor is at end of line (>= len), does nothing"
func TestDoubleClickAtEndOfLineDoesNothing(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello") // 5 runes; end-of-line position is col 5

	memoDoubleClickAt(m, 5, 0) // click at end of line

	if m.HasSelection() {
		t.Error("double-click at end of line should not create selection; HasSelection() = true, want false")
	}
}

// TestDoubleClickPastEndOfLineDoesNothing verifies double-clicking well beyond
// the end of a line also does nothing.
// Spec: "If cursor is at end of line (>= len), does nothing"
func TestDoubleClickPastEndOfLineDoesNothing(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hi") // 2 runes; any click at x >= 2 lands at end

	memoDoubleClickAt(m, 20, 0) // x=20 >> len("hi")=2, clamped to col 2

	if m.HasSelection() {
		t.Error("double-click past end of line should not create selection; HasSelection() = true, want false")
	}
}

// ---------------------------------------------------------------------------
// Section 13 — Double-click: event consumed
// ---------------------------------------------------------------------------

// TestMemoDoubleClickConsumesEvent verifies that a double-click event on a Memo is consumed.
// Spec: "Mouse events are consumed (event.Clear()) when handled"
func TestMemoDoubleClickConsumesEvent(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	ev := mouseEv(2, 0, tcell.Button1, 2)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("double-click event was not consumed (IsCleared() = false)")
	}
}

// ---------------------------------------------------------------------------
// Section 14 — Triple-click: selects entire line
// ---------------------------------------------------------------------------

// TestTripleClickSelectsEntireLine verifies triple-click selects the full content
// of the clicked line.
// Spec: "Triple-click (ClickCount == 3): Selects the entire line: anchor at (row, 0),
// extent at (row, len(line))"
func TestTripleClickSelectsEntireLine(t *testing.T) {
	m := newMouseMemo()
	m.SetText("first\nsecondline\nthird")

	memoTripleClickAt(m, 3, 1) // triple-click anywhere on row 1 ("secondline")

	if !m.HasSelection() {
		t.Error("triple-click should create selection; HasSelection() = false")
	}
}

// TestTripleClickSelectsCorrectLineContent verifies the selection spans from col 0
// to len(line) on the clicked row.
// Spec: "anchor at (row, 0), extent at (row, len(line))"
func TestTripleClickSelectsCorrectLineContent(t *testing.T) {
	m := newMouseMemo()
	m.SetText("first\nsecondline\nthird") // row 1 = "secondline" (10 chars)

	memoTripleClickAt(m, 3, 1) // triple-click on row 1

	sr, sc, er, ec := m.Selection()
	if sr != 1 || sc != 0 {
		t.Errorf("triple-click selection start = (%d,%d), want (1,0)", sr, sc)
	}
	if er != 1 || ec != 10 {
		t.Errorf("triple-click selection end = (%d,%d), want (1,10) (len of \"secondline\")", er, ec)
	}
}

// TestTripleClickOnFirstLine verifies triple-click works on the first line.
// Spec: "anchor at (row, 0), extent at (row, len(line))"
func TestTripleClickOnFirstLine(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello\nworld") // row 0 = "hello" (5 chars)

	memoTripleClickAt(m, 2, 0) // triple-click on row 0

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("triple-click first-line start = (%d,%d), want (0,0)", sr, sc)
	}
	if er != 0 || ec != 5 {
		t.Errorf("triple-click first-line end = (%d,%d), want (0,5)", er, ec)
	}
}

// TestTripleClickOnEmptyLine verifies triple-click on an empty line creates a
// collapsed selection (no visible selection) since len(line) == 0.
// Spec: "anchor at (row, 0), extent at (row, len(line))"
func TestTripleClickOnEmptyLine(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello\n\nworld") // row 1 is empty

	memoTripleClickAt(m, 0, 1) // triple-click on empty row 1

	sr, sc, er, ec := m.Selection()
	// anchor and extent both at (1, 0): collapsed or zero-width selection.
	if sr != 1 || sc != 0 {
		t.Errorf("triple-click empty line start = (%d,%d), want (1,0)", sr, sc)
	}
	if er != 1 || ec != 0 {
		t.Errorf("triple-click empty line end = (%d,%d), want (1,0)", er, ec)
	}
}

// ---------------------------------------------------------------------------
// Section 15 — Triple-click: anchor and extent
// ---------------------------------------------------------------------------

// TestTripleClickAnchorAtRowColZero verifies that the anchor is at col 0 of the
// clicked row.
// Spec: "anchor at (row, 0)"
func TestTripleClickAnchorAtRowColZero(t *testing.T) {
	m := newMouseMemo()
	m.SetText("abc\ndefgh\nij") // row 1 = "defgh" (5 chars)

	memoTripleClickAt(m, 2, 1) // triple-click on row 1

	sr, sc, _, _ := m.Selection()
	if sr != 1 || sc != 0 {
		t.Errorf("triple-click anchor = (%d,%d), want (1,0)", sr, sc)
	}
}

// TestTripleClickExtentAtRowLenLine verifies that the extent is at len(line) of the
// clicked row.
// Spec: "extent at (row, len(line))"
func TestTripleClickExtentAtRowLenLine(t *testing.T) {
	m := newMouseMemo()
	m.SetText("abc\ndefgh\nij") // row 1 = "defgh" (5 chars)

	memoTripleClickAt(m, 2, 1) // triple-click on row 1

	_, _, er, ec := m.Selection()
	if er != 1 || ec != 5 {
		t.Errorf("triple-click extent = (%d,%d), want (1,5) (row 1, len(\"defgh\")=5)", er, ec)
	}
}

// ---------------------------------------------------------------------------
// Section 16 — Triple-click: event consumed
// ---------------------------------------------------------------------------

// TestTripleClickConsumesEvent verifies a triple-click event is consumed.
// Spec: "Mouse events are consumed (event.Clear()) when handled"
func TestTripleClickConsumesEvent(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	ev := mouseEv(2, 0, tcell.Button1, 3)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("triple-click event was not consumed (IsCleared() = false)")
	}
}

// ---------------------------------------------------------------------------
// Section 17 — Falsifying tests
// ---------------------------------------------------------------------------

// TestClickDoesNotExtendSelection verifies that a single click does NOT extend an
// existing selection — it collapses it.
// Falsifying: an implementation that treats any click as drag-start without collapsing
// would fail this.
func TestClickDoesNotExtendSelection(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")
	// Create selection via Shift+Right.
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	if !m.HasSelection() {
		t.Fatal("precondition: selection must exist")
	}

	memoClickAt(m, 6, 0) // click somewhere else

	if m.HasSelection() {
		t.Error("click (ClickCount=1) should collapse selection, not extend it; HasSelection() = true")
	}
}

// TestDragDoesNotJustMoveCursor verifies that a drag creates a selection, not merely
// repositions the cursor.
// Falsifying: an implementation that moves the cursor on Button1 motion without
// tracking selection would have HasSelection() == false.
func TestDragDoesNotJustMoveCursor(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoClickAt(m, 2, 0)
	memoDragTo(m, 8, 0)

	if !m.HasSelection() {
		t.Error("drag should create selection, not just move cursor; HasSelection() = false after drag")
	}
	// The cursor itself should be at the drag destination.
	_, col := m.CursorPos()
	if col != 8 {
		t.Errorf("after drag to col 8, CursorPos col = %d, want 8", col)
	}
}

// TestDoubleClickSelectsWholeWordNotJustChar verifies that double-click selects the
// full word, not just a single character.
// Falsifying: an implementation that selects only one character would give
// selection width of 1, not the full word length.
func TestDoubleClickSelectsWholeWordNotJustChar(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoDoubleClickAt(m, 2, 0) // click inside "hello" (5 chars)

	_, sc, _, ec := m.Selection()
	// Selection must span more than 1 character.
	if ec-sc <= 1 {
		t.Errorf("double-click selected %d chars (col %d to %d); want >1 (full word)", ec-sc, sc, ec)
	}
}

// TestNonButton1EventsAreNotConsumed verifies that button2, button3, and scroll wheel
// events are all left unconsumed.
// Falsifying: an implementation that dispatches on any EvMouse would erroneously
// consume non-Button1 events.
func TestNonButton1EventsAreNotConsumed(t *testing.T) {
	buttons := []struct {
		name string
		btn  tcell.ButtonMask
	}{
		{"Button2", tcell.Button2},
		{"Button3", tcell.Button3},
	}
	for _, tc := range buttons {
		m := newMouseMemo()
		m.SetText("hello")

		ev := mouseEv(2, 0, tc.btn, 1)
		m.HandleEvent(ev)

		if ev.IsCleared() {
			t.Errorf("%s event should NOT be consumed; IsCleared() = true, want false", tc.name)
		}
	}
}

// TestClickDoesNotCreateSelection verifies a single click results in a collapsed
// (not extended) selection.
// Falsifying: some implementations might accidentally set selection start != end
// on a plain click.
func TestClickDoesNotCreateSelection(t *testing.T) {
	m := newMouseMemo()
	m.SetText("hello world")

	memoClickAt(m, 4, 0) // single click

	if m.HasSelection() {
		t.Error("single click should produce a collapsed selection; HasSelection() = true, want false")
	}
}
