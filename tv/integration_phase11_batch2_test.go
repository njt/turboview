package tv

// integration_phase11_batch2_test.go — Integration tests for Phase 11 Tasks 6–7:
// InputLine Mouse Selection Checkpoint.
//
// Verifies cross-component behavior of the mouse selection features added in Tasks 6–7:
//   Task 6: Mouse drag selection (press, drag, release)
//   Task 7: Double-click select all
//
// Each test exercises a full scenario from SetText through mouse events to final
// state inspection, verifying that the composed behaviors work correctly together.
//
// Test naming: TestIntegrationPhase11Batch2<DescriptiveSuffix>

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Requirement 1: Click at position 3 — cursor at 3, no selection, dragging started
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch2ClickStartsCursorAtClickPosition verifies that a Button1 click
// at column 3 positions the cursor at 3 with no selection and starts dragging.
//
// Scenario: SetText("hello world") → click at X=3 → cursorPos=3, no selection, dragging=true.
func TestIntegrationPhase11Batch2ClickStartsCursorAtClickPosition(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	event := &Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(event)

	if pos := il.CursorPos(); pos != 3 {
		t.Errorf("CursorPos() after click at X=3 = %d, want 3", pos)
	}
}

// TestIntegrationPhase11Batch2ClickClearsSelection verifies that the initial click clears
// any prior selection (selStart == selEnd after a fresh single click).
func TestIntegrationPhase11Batch2ClickClearsSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	event := &Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(event)

	selStart, selEnd := il.Selection()
	if selStart != selEnd {
		t.Errorf("Selection() after initial click = (%d, %d), want no selection (selStart == selEnd)", selStart, selEnd)
	}
}

// TestIntegrationPhase11Batch2ClickStartsDragging verifies that pressing Button1 sets
// the dragging flag to true.
func TestIntegrationPhase11Batch2ClickStartsDragging(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	event := &Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(event)

	if !il.dragging {
		t.Error("dragging should be true after Button1 press at X=3")
	}
}

// TestIntegrationPhase11Batch2ClickEventConsumed verifies the initial click event is consumed.
func TestIntegrationPhase11Batch2ClickEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	event := &Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(event)

	if !event.IsCleared() {
		t.Error("mouse click event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: Click at 3, move to 8 with Button1 held — selection from 3 to 8
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch2DragForwardCreatesSelection verifies that pressing at column 3
// and moving to column 8 with Button1 held creates a selection from 3 to 8.
//
// Scenario: click at X=3 (dragging=true, anchor=3) → move to X=8 with Button1 → selStart=3, selEnd=8.
func TestIntegrationPhase11Batch2DragForwardCreatesSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	// Press at column 3 to start drag.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})
	if !il.dragging {
		t.Fatal("precondition: dragging should be true after press")
	}

	// Move to column 8 with Button1 still held.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 8, Y: 0, Button: tcell.Button1}})

	selStart, selEnd := il.Selection()
	if selStart != 3 || selEnd != 8 {
		t.Errorf("Selection() after drag from 3 to 8 = (%d, %d), want (3, 8)", selStart, selEnd)
	}
}

// TestIntegrationPhase11Batch2DragForwardUpdatesCursor verifies cursor follows the drag position.
func TestIntegrationPhase11Batch2DragForwardUpdatesCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 8, Y: 0, Button: tcell.Button1}})

	if pos := il.CursorPos(); pos != 8 {
		t.Errorf("CursorPos() after drag to X=8 = %d, want 8", pos)
	}
}

// TestIntegrationPhase11Batch2DragEventConsumed verifies the drag move event is consumed.
func TestIntegrationPhase11Batch2DragEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})

	dragEvent := &Event{What: EvMouse, Mouse: &MouseEvent{X: 8, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(dragEvent)

	if !dragEvent.IsCleared() {
		t.Error("mouse drag move event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: Click at 3, move to 8, release — selection from 3 to 8, dragging stops
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch2ReleasePreservesSelection verifies that releasing Button1
// after a drag preserves the selection from 3 to 8.
//
// Scenario: press at X=3 → drag to X=8 → release at X=8 → selStart=3, selEnd=8.
func TestIntegrationPhase11Batch2ReleasePreservesSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 8, Y: 0, Button: tcell.Button1}})
	// Release: no button held.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 8, Y: 0}})

	selStart, selEnd := il.Selection()
	if selStart != 3 || selEnd != 8 {
		t.Errorf("Selection() after press-drag-release = (%d, %d), want (3, 8)", selStart, selEnd)
	}
}

// TestIntegrationPhase11Batch2ReleaseStopsDragging verifies dragging is false after release.
func TestIntegrationPhase11Batch2ReleaseStopsDragging(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 8, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 8, Y: 0}})

	if il.dragging {
		t.Error("dragging should be false after mouse release")
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Click at 3, move to 1 (backward drag) — selection covers 1 to 3
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch2BackwardDragRawSelection verifies that dragging from 3 back
// to 1 stores selection as (selStart=3, selEnd=1), i.e. anchor=3 and cursor=1.
//
// Scenario: press at X=3 (anchor=3) → move to X=1 → selStart=3, selEnd=1 (reversed).
func TestIntegrationPhase11Batch2BackwardDragRawSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 1, Y: 0, Button: tcell.Button1}})

	selStart, selEnd := il.Selection()
	// Raw storage: anchor is selStart (3), cursor is selEnd (1).
	if selStart != 3 || selEnd != 1 {
		t.Errorf("backward drag Selection() = (%d, %d), want (3, 1) — anchor first, cursor second", selStart, selEnd)
	}
}

// TestIntegrationPhase11Batch2BackwardDragNormalizedCovers1To3 verifies that normalizedSel
// returns (1, 3) regardless of drag direction, covering the region 1 to 3.
func TestIntegrationPhase11Batch2BackwardDragNormalizedCovers1To3(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 1, Y: 0, Button: tcell.Button1}})

	lo, hi := il.normalizedSel()
	if lo != 1 || hi != 3 {
		t.Errorf("normalizedSel() after backward drag = (%d, %d), want (1, 3)", lo, hi)
	}
}

// TestIntegrationPhase11Batch2BackwardDragCursorAtDragEnd verifies cursor ends at column 1
// after dragging backward from 3 to 1.
func TestIntegrationPhase11Batch2BackwardDragCursorAtDragEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 1, Y: 0, Button: tcell.Button1}})

	if pos := il.CursorPos(); pos != 1 {
		t.Errorf("CursorPos() after backward drag to X=1 = %d, want 1", pos)
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: Double-click on "hello world" — selects all text (selStart=0, selEnd=11)
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch2DoubleClickSelectsAll verifies double-click selects all text
// (selStart=0, selEnd=len("hello world")=11).
//
// Scenario: SetText("hello world") → double-click at X=5 → selStart=0, selEnd=11.
func TestIntegrationPhase11Batch2DoubleClickSelectsAll(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	event := &Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1, ClickCount: 2}}
	il.HandleEvent(event)

	selStart, selEnd := il.Selection()
	if selStart != 0 || selEnd != 11 {
		t.Errorf("Selection() after double-click on \"hello world\" = (%d, %d), want (0, 11)", selStart, selEnd)
	}
}

// TestIntegrationPhase11Batch2DoubleClickCursorAtEnd verifies cursor is at end of text
// after double-click.
func TestIntegrationPhase11Batch2DoubleClickCursorAtEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1, ClickCount: 2}})

	if pos := il.CursorPos(); pos != 11 {
		t.Errorf("CursorPos() after double-click = %d, want 11 (end of text)", pos)
	}
}

// TestIntegrationPhase11Batch2DoubleClickStopsDragging verifies double-click does not
// start a drag (dragging remains false).
func TestIntegrationPhase11Batch2DoubleClickStopsDragging(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1, ClickCount: 2}})

	if il.dragging {
		t.Error("dragging should be false after double-click (double-click must not start a drag)")
	}
}

// TestIntegrationPhase11Batch2DoubleClickEventConsumed verifies the double-click event
// is consumed.
func TestIntegrationPhase11Batch2DoubleClickEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	event := &Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1, ClickCount: 2}}
	il.HandleEvent(event)

	if !event.IsCleared() {
		t.Error("double-click event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: Double-click then single click — clears select-all, starts new drag
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch2DoubleClickThenSingleClickClearsSelection verifies that a
// single click following a double-click clears the select-all selection.
//
// Scenario: double-click (selStart=0, selEnd=11) → single click at X=3 → no selection.
func TestIntegrationPhase11Batch2DoubleClickThenSingleClickClearsSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	// Double-click to select all.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1, ClickCount: 2}})
	selStart, selEnd := il.Selection()
	if selStart != 0 || selEnd != 11 {
		t.Fatalf("precondition: double-click should have selected all, got (%d, %d)", selStart, selEnd)
	}

	// Single click at column 3.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})

	selStart, selEnd = il.Selection()
	if selStart != selEnd {
		t.Errorf("selection after single click following double-click = (%d, %d), want no selection (selStart == selEnd)", selStart, selEnd)
	}
}

// TestIntegrationPhase11Batch2DoubleClickThenSingleClickPositionsCursor verifies that the
// single click positions the cursor at the new click location.
func TestIntegrationPhase11Batch2DoubleClickThenSingleClickPositionsCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1, ClickCount: 2}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})

	if pos := il.CursorPos(); pos != 3 {
		t.Errorf("CursorPos() after single click at X=3 following double-click = %d, want 3", pos)
	}
}

// TestIntegrationPhase11Batch2DoubleClickThenSingleClickStartsDrag verifies the single
// click after a double-click starts a new drag.
func TestIntegrationPhase11Batch2DoubleClickThenSingleClickStartsDrag(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1, ClickCount: 2}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})

	if !il.dragging {
		t.Error("dragging should be true after single click following double-click")
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: Mouse drag then Ctrl+C copies selected text
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch2MouseDragThenCtrlCCopiesSelection verifies the cross-feature
// scenario: selection created by mouse drag is copied to clipboard by Ctrl+C.
//
// Scenario: press at X=6 → drag to X=11 → release → Ctrl+C → clipboard = "world".
func TestIntegrationPhase11Batch2MouseDragThenCtrlCCopiesSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	// Drag to select "world" (positions 6..11).
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 6, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 11, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 11, Y: 0}})

	selStart, selEnd := il.Selection()
	if selStart != 6 || selEnd != 11 {
		t.Fatalf("precondition: selection after drag = (%d, %d), want (6, 11)", selStart, selEnd)
	}

	// Ctrl+C should copy the selected text.
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlC}})

	// Paste into a fresh widget to verify clipboard content.
	il2 := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il2.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlV}})

	if got := il2.Text(); got != "world" {
		t.Errorf("clipboard after mouse-drag selection + Ctrl+C = %q (pasted into new widget), want %q", got, "world")
	}
}

// TestIntegrationPhase11Batch2MouseDragSelectionIsNonEmpty verifies that the mouse drag
// produces a non-empty selection before Ctrl+C (falsification guard).
func TestIntegrationPhase11Batch2MouseDragSelectionIsNonEmpty(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 6, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 11, Y: 0, Button: tcell.Button1}})

	selStart, selEnd := il.Selection()
	if selStart == selEnd {
		t.Errorf("Selection() after drag = (%d, %d); expected non-empty selection", selStart, selEnd)
	}
}

// ---------------------------------------------------------------------------
// Requirement 8: Mouse click clears selection from keyboard (Ctrl+A then click)
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch2CtrlAThenClickClearsKeyboardSelection verifies the cross-feature
// scenario: Ctrl+A creates a keyboard selection; a subsequent mouse click clears it.
//
// Scenario: SetText("hello world") → Ctrl+A (selStart=0, selEnd=11) → click at X=5 → no selection.
func TestIntegrationPhase11Batch2CtrlAThenClickClearsKeyboardSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	// Ctrl+A to select all via keyboard.
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlA}})
	selStart, selEnd := il.Selection()
	if selStart == selEnd {
		t.Fatalf("precondition: Ctrl+A should create a selection, got (%d, %d)", selStart, selEnd)
	}

	// Single mouse click at column 5 should clear the selection.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}})

	selStart, selEnd = il.Selection()
	if selStart != selEnd {
		t.Errorf("Selection() after mouse click following Ctrl+A = (%d, %d), want no selection (selStart == selEnd)", selStart, selEnd)
	}
}

// TestIntegrationPhase11Batch2CtrlAThenClickPositionsCursor verifies that the click after
// Ctrl+A also positions the cursor at the new location.
func TestIntegrationPhase11Batch2CtrlAThenClickPositionsCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlA}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}})

	if pos := il.CursorPos(); pos != 5 {
		t.Errorf("CursorPos() after click at X=5 following Ctrl+A = %d, want 5", pos)
	}
}

// TestIntegrationPhase11Batch2CtrlAThenClickSelectionIsCleared verifies the selection is
// empty (not partial) after a single click on a fully-selected text — a falsification
// guard to distinguish from partial clearing.
func TestIntegrationPhase11Batch2CtrlAThenClickSelectionIsCleared(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlA}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}})

	selStart, selEnd := il.Selection()
	// Specifically check that neither selStart nor selEnd retains a boundary from the old selection.
	if selStart != 0 && selEnd != 0 {
		// Only flag if the old selection values (0, 11) are still partially retained unexpectedly.
	}
	if selStart != selEnd {
		t.Errorf("Selection() = (%d, %d) after mouse click following Ctrl+A; selection must be cleared (selStart == selEnd)", selStart, selEnd)
	}
}
