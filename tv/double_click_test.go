package tv

// Tests for InputLine double-click select-all (spec 5.6).
//
// Each test cites the spec sentence it verifies.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// mouseDoubleClickEv creates a mouse Button1 double-click event.
func mouseDoubleClickEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 2}}
}

// ---------------------------------------------------------------------------
// Double-click selects all
// ---------------------------------------------------------------------------

// TestDoubleClickSelectsAllText verifies double-click sets selection over entire text.
// Spec 5.6: "When ClickCount >= 2: select all text (selStart=0, selEnd=len(text))."
func TestDoubleClickSelectsAllText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world") // 11 runes

	il.HandleEvent(mouseDoubleClickEv(5, 0))

	selStart, selEnd := il.Selection()
	if selStart != 0 || selEnd != 11 {
		t.Errorf("selection after double-click = (%d, %d), want (0, 11)", selStart, selEnd)
	}
}

// TestDoubleClickMovescursorToEnd verifies double-click places cursor at end of text.
// Spec 5.6: "cursorPos=len(text)."
func TestDoubleClickMovesCursorToEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world") // 11 runes

	il.HandleEvent(mouseDoubleClickEv(2, 0))

	if pos := il.CursorPos(); pos != 11 {
		t.Errorf("CursorPos() after double-click on \"hello world\" = %d, want 11", pos)
	}
}

// TestDoubleClickStopsDragging verifies double-click sets dragging=false.
// Spec 5.6: "Set dragging=false (don't start a drag on double-click)."
func TestDoubleClickStopsDragging(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(mouseDoubleClickEv(5, 0))

	if il.dragging {
		t.Error("dragging should be false after double-click")
	}
}

// TestDoubleClickEventConsumed verifies the double-click event is consumed.
// Spec 5.6: "adjustScroll() called, event cleared."
func TestDoubleClickEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")

	ev := mouseDoubleClickEv(2, 0)
	il.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("double-click event was not consumed")
	}
}

// TestDoubleClickOnEmptyText verifies double-click on empty text is safe.
// Spec 5.6: "selStart=0, selEnd=len(text)" — when text is empty, both are 0.
func TestDoubleClickOnEmptyText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	// text is empty

	il.HandleEvent(mouseDoubleClickEv(0, 0))

	selStart, selEnd := il.Selection()
	if selStart != 0 || selEnd != 0 {
		t.Errorf("selection after double-click on empty text = (%d, %d), want (0, 0)", selStart, selEnd)
	}
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("CursorPos() after double-click on empty text = %d, want 0", pos)
	}
}

// ---------------------------------------------------------------------------
// Double-click then single click
// ---------------------------------------------------------------------------

// TestDoubleClickThenSingleClickClearsSelectAll verifies a subsequent single click
// starts a new drag (clearing the select-all).
// Spec 5.5/5.6: single click clears existing selection and starts drag anchor.
func TestDoubleClickThenSingleClickClearsSelectAll(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	// Double-click to select all.
	il.HandleEvent(mouseDoubleClickEv(5, 0))
	selStart, selEnd := il.Selection()
	if selStart != 0 || selEnd != 11 {
		t.Fatalf("precondition: double-click did not select all (got %d, %d)", selStart, selEnd)
	}

	// Single click at col 3 should clear selection and start a new drag.
	il.HandleEvent(mousePressEv(3, 0))

	selStart, selEnd = il.Selection()
	if selStart != selEnd {
		t.Errorf("after single click following double-click, selection should be cleared (got %d, %d)", selStart, selEnd)
	}
	if pos := il.CursorPos(); pos != 3 {
		t.Errorf("CursorPos() after single click at col 3 = %d, want 3", pos)
	}
}
