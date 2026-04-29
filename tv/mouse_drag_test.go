package tv

// Tests for InputLine mouse drag selection (spec 5.5).
//
// Each test cites the spec sentence it verifies.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// mousePressEv creates a mouse Button1 press event (used to begin a drag).
func mousePressEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1}}
}

// mouseMoveEv creates a mouse move event with Button1 held (drag-in-progress).
func mouseMoveEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1}}
}

// mouseReleaseEv creates a mouse release event (no button held).
func mouseReleaseEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: 0}}
}

// ---------------------------------------------------------------------------
// Press starts drag
// ---------------------------------------------------------------------------

// TestMousePressStartsDragging verifies that pressing Button1 sets dragging=true.
// Spec 5.5: "On Button1 press: set dragging=true."
func TestMousePressStartsDragging(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(mousePressEv(3, 0))

	if !il.dragging {
		t.Error("dragging should be true after Button1 press")
	}
}

// TestMousePressPositionsCursor verifies Button1 press sets cursor at clicked column.
// Spec 5.5: "On Button1 press: position cursor at click location."
func TestMousePressPositionsCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(mousePressEv(3, 0))

	if pos := il.CursorPos(); pos != 3 {
		t.Errorf("CursorPos() after press at col 3 = %d, want 3", pos)
	}
}

// TestMousePressSetsAnchor verifies Button1 press sets dragAnchor to cursorPos.
// Spec 5.5: "On Button1 press: set dragAnchor=cursorPos."
func TestMousePressSetsAnchor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(mousePressEv(5, 0))

	if il.dragAnchor != 5 {
		t.Errorf("dragAnchor after press at col 5 = %d, want 5", il.dragAnchor)
	}
}

// TestMousePressClearsExistingSelection verifies Button1 press clears any existing selection.
// Spec 5.5: "On Button1 press: clear any existing selection."
func TestMousePressClearsExistingSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")
	// Establish a selection via Ctrl+A.
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA))

	selStart, selEnd := il.Selection()
	if selStart == selEnd {
		t.Fatal("precondition: Ctrl+A should have created a selection")
	}

	il.HandleEvent(mousePressEv(3, 0))

	selStart, selEnd = il.Selection()
	if selStart != selEnd {
		t.Errorf("selection after press should be cleared (selStart=%d, selEnd=%d)", selStart, selEnd)
	}
}

// TestMousePressEventConsumed verifies Button1 press event is consumed.
// Spec 5.5: "Clear event."
func TestMousePressEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")

	ev := mousePressEv(2, 0)
	il.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("mouse press event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Drag extends selection
// ---------------------------------------------------------------------------

// TestMouseDragForwardCreatesSelection verifies moving right while Button1 held
// creates a selection from anchor to cursor.
// Spec 5.5: "On mouse move with Button1 held: set selection from dragAnchor to cursorPos."
func TestMouseDragForwardCreatesSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	// Press at col 3, then drag to col 8.
	il.HandleEvent(mousePressEv(3, 0))
	il.HandleEvent(mouseMoveEv(8, 0))

	selStart, selEnd := il.Selection()
	if selStart != 3 || selEnd != 8 {
		t.Errorf("selection after drag from 3 to 8 = (%d, %d), want (3, 8)", selStart, selEnd)
	}
}

// TestMouseDragUpdatesCursorPosition verifies cursor tracks the drag position.
// Spec 5.5: "On mouse move with Button1 held: update cursor position."
func TestMouseDragUpdatesCursorPosition(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(mousePressEv(3, 0))
	il.HandleEvent(mouseMoveEv(8, 0))

	if pos := il.CursorPos(); pos != 8 {
		t.Errorf("CursorPos() after drag to col 8 = %d, want 8", pos)
	}
}

// TestMouseDragEventConsumed verifies drag-move event is consumed.
// Spec 5.5: "Clear event."
func TestMouseDragEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(mousePressEv(3, 0))
	ev := mouseMoveEv(8, 0)
	il.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("mouse drag-move event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Release ends drag
// ---------------------------------------------------------------------------

// TestMouseReleaseStopsDragging verifies releasing Button1 sets dragging=false.
// Spec 5.5: "On mouse release (Button1 not held, dragging was true): set dragging=false."
func TestMouseReleaseStopsDragging(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(mousePressEv(3, 0))
	il.HandleEvent(mouseReleaseEv(8, 0))

	if il.dragging {
		t.Error("dragging should be false after mouse release")
	}
}

// TestMouseReleasePreservesSelection verifies the selection remains after release.
// Spec 5.5: drag establishes selection; release does not clear it.
func TestMouseReleasePreservesSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(mousePressEv(3, 0))
	il.HandleEvent(mouseMoveEv(8, 0))
	il.HandleEvent(mouseReleaseEv(8, 0))

	selStart, selEnd := il.Selection()
	if selStart != 3 || selEnd != 8 {
		t.Errorf("selection after press-drag-release = (%d, %d), want (3, 8)", selStart, selEnd)
	}
}

// ---------------------------------------------------------------------------
// Backward drag
// ---------------------------------------------------------------------------

// TestMouseDragBackwardCreatesReversedSelection verifies dragging left creates
// a selection where selStart > selEnd (reversed, normalized by normalizedSel).
// Spec 5.5: "set selection from dragAnchor to cursorPos" (order preserved as-is).
func TestMouseDragBackwardCreatesReversedSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	// Press at col 8, drag back to col 3.
	il.HandleEvent(mousePressEv(8, 0))
	il.HandleEvent(mouseMoveEv(3, 0))

	selStart, selEnd := il.Selection()
	// selStart is the anchor (8), selEnd is cursor (3); reversed.
	if selStart != 8 || selEnd != 3 {
		t.Errorf("backward drag selection = (%d, %d), want (8, 3)", selStart, selEnd)
	}
	// normalizedSel should still give (3, 8).
	lo, hi := il.normalizedSel()
	if lo != 3 || hi != 8 {
		t.Errorf("normalizedSel() after backward drag = (%d, %d), want (3, 8)", lo, hi)
	}
}

// ---------------------------------------------------------------------------
// Clamping
// ---------------------------------------------------------------------------

// TestMousePressClampsBelowZero verifies click X < widget origin is clamped to 0.
// Spec 5.5: "Mouse X is translated to text position … Clamp to [0, len(text)]."
func TestMousePressClampsBelowZero(t *testing.T) {
	// Widget at X=5; a click at absolute X=3 gives col=-2, clamped to 0.
	il := NewInputLine(NewRect(5, 0, 15, 1), 0)
	il.SetText("hello")

	il.HandleEvent(mousePressEv(3, 0))

	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("CursorPos() after click before widget origin = %d, want 0", pos)
	}
}

// TestMousePressClampsAboveTextLength verifies click past text end is clamped to len(text).
// Spec 5.5: "Clamp to [0, len(text)]."
func TestMousePressClampsAboveTextLength(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hi") // 2 runes

	il.HandleEvent(mousePressEv(15, 0)) // well past text end

	if pos := il.CursorPos(); pos != 2 {
		t.Errorf("CursorPos() after click past text end = %d, want 2 (clamped to len)", pos)
	}
}

// ---------------------------------------------------------------------------
// NilMouse guard
// ---------------------------------------------------------------------------

// TestMouseEventWithNilMouseIsIgnored verifies HandleEvent with EvMouse but nil
// Mouse pointer does not panic.
// Spec 5.5: defensive nil check in mouse handler.
func TestMouseEventWithNilMouseIsIgnored(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")

	// Must not panic.
	il.HandleEvent(&Event{What: EvMouse, Mouse: nil})
}
