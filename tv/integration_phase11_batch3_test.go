package tv

// integration_phase11_batch3_test.go — Integration tests for Phase 11 Tasks 9–10:
// InputLine Scroll Indicators and Edge Auto-Scroll Checkpoint.
//
// Verifies cross-component behavior of the scroll features added in Tasks 9–10:
//   Task 9: Scroll indicators (◄ and ►)
//   Task 10: Edge auto-scroll during mouse drag
//
// Each test exercises a full scenario from SetText through rendering or mouse events
// to final state inspection, verifying that the composed behaviors work correctly together.
//
// Test naming: TestIntegrationPhase11Batch3<DescriptiveSuffix>

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Requirement 1: scrollOffset=0 — no ◄ at col 0, ► at col 9
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch3ScrollOffset0NoLeftIndicator verifies that when
// scrollOffset=0 (text starts at the left edge), no ◄ indicator appears at col 0.
//
// Scenario: InputLine width=10, text="abcdefghijklmnop" (16 chars), scrollOffset=0
// → Draw → cell(0,0).Rune != '◄'.
func TestIntegrationPhase11Batch3ScrollOffset0NoLeftIndicator(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 0
	il.scrollOffset = 0

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	if buf.GetCell(0, 0).Rune == '◄' {
		t.Error("col 0 shows ◄ when scrollOffset=0; left indicator must not appear at the left edge")
	}
}

// TestIntegrationPhase11Batch3ScrollOffset0RightIndicatorPresent verifies that when
// text overflows to the right and scrollOffset=0, a ► indicator appears at col 9.
//
// Scenario: InputLine width=10, text="abcdefghijklmnop" (16 chars), scrollOffset=0
// → Draw → cell(9,0).Rune == '►'.
func TestIntegrationPhase11Batch3ScrollOffset0RightIndicatorPresent(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 0
	il.scrollOffset = 0

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	if buf.GetCell(9, 0).Rune != '►' {
		t.Errorf("col 9 does not show ► when text overflows right with scrollOffset=0, got %q",
			string(buf.GetCell(9, 0).Rune))
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: scrollOffset=5 — both ◄ at col 0 and ► at col 9
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch3ScrollOffset5LeftIndicatorPresent verifies that when
// scrollOffset=5 (some text is scrolled off left), ◄ appears at col 0.
//
// Scenario: InputLine width=10, text="abcdefghijklmnop" (16 chars), scrollOffset=5
// → Draw → cell(0,0).Rune == '◄'.
func TestIntegrationPhase11Batch3ScrollOffset5LeftIndicatorPresent(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 8
	il.scrollOffset = 5

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	if buf.GetCell(0, 0).Rune != '◄' {
		t.Errorf("col 0 does not show ◄ when scrollOffset=5, got %q",
			string(buf.GetCell(0, 0).Rune))
	}
}

// TestIntegrationPhase11Batch3ScrollOffset5RightIndicatorPresent verifies that when
// scrollOffset=5 with 16-char text and width=10, ► also appears at col 9 because
// chars 15 is not yet visible (scrollOffset+10=15 < 16).
//
// Scenario: InputLine width=10, text="abcdefghijklmnop" (16 chars), scrollOffset=5
// → Draw → cell(9,0).Rune == '►'.
func TestIntegrationPhase11Batch3ScrollOffset5RightIndicatorPresent(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 8
	il.scrollOffset = 5

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	if buf.GetCell(9, 0).Rune != '►' {
		t.Errorf("col 9 does not show ► when scrollOffset=5 and text extends further right, got %q",
			string(buf.GetCell(9, 0).Rune))
	}
}

// TestIntegrationPhase11Batch3ScrollOffset5BothIndicatorsPresent is a combined check
// that verifies both ◄ and ► appear simultaneously when the view is mid-scroll.
func TestIntegrationPhase11Batch3ScrollOffset5BothIndicatorsPresent(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 8
	il.scrollOffset = 5

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	leftOk := buf.GetCell(0, 0).Rune == '◄'
	rightOk := buf.GetCell(9, 0).Rune == '►'

	if !leftOk || !rightOk {
		t.Errorf("expected both indicators with scrollOffset=5: ◄ at col 0 = %v (got %q), ► at col 9 = %v (got %q)",
			leftOk, string(buf.GetCell(0, 0).Rune),
			rightOk, string(buf.GetCell(9, 0).Rune))
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: scrollOffset=6 (last char visible) — ◄ at col 0, no ► at col 9
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch3ScrollOffset6LeftIndicatorPresent verifies that when
// scrollOffset=6 with 16-char text and width=10 (scrollOffset+w=16=len(text)),
// ◄ still appears at col 0.
//
// Scenario: InputLine width=10, text="abcdefghijklmnop" (16 chars), scrollOffset=6
// → Draw → cell(0,0).Rune == '◄'.
func TestIntegrationPhase11Batch3ScrollOffset6LeftIndicatorPresent(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 15
	il.scrollOffset = 6

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	if buf.GetCell(0, 0).Rune != '◄' {
		t.Errorf("col 0 does not show ◄ when scrollOffset=6, got %q",
			string(buf.GetCell(0, 0).Rune))
	}
}

// TestIntegrationPhase11Batch3ScrollOffset6NoRightIndicator verifies that when all
// remaining text fits within the visible area (scrollOffset+width == len(text)),
// no ► indicator appears at col 9.
//
// Scenario: InputLine width=10, text="abcdefghijklmnop" (16 chars), scrollOffset=6
// → scrollOffset+w=16=len(text) → Draw → cell(9,0).Rune != '►'.
func TestIntegrationPhase11Batch3ScrollOffset6NoRightIndicator(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 15
	il.scrollOffset = 6

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	if buf.GetCell(9, 0).Rune == '►' {
		t.Error("col 9 shows ► when all remaining text fits within the view (scrollOffset+w == len(text))")
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Mouse drag to right edge with Button1 held — scrollOffset increases,
// selection extends
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch3DragToRightEdgeIncreasesScrollOffset verifies that when
// the user drags to the rightmost column (X=9) with Button1 held during a drag and
// there is still text to the right, scrollOffset increases by 1.
//
// Scenario: press at X=5 (anchor=5), drag to X=9 (right edge, w-1) with text extending
// beyond the view → scrollOffset goes from 0 to 1.
func TestIntegrationPhase11Batch3DragToRightEdgeIncreasesScrollOffset(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop") // 16 chars, width 10
	il.cursorPos = 5
	il.scrollOffset = 0

	// Start drag.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}})
	if !il.dragging {
		t.Fatal("precondition: dragging should be true after press")
	}

	// Drag to right edge.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 9, Y: 0, Button: tcell.Button1}})

	if il.scrollOffset != 1 {
		t.Errorf("scrollOffset = %d after drag to right edge, want 1", il.scrollOffset)
	}
}

// TestIntegrationPhase11Batch3DragToRightEdgeExtendsSelection verifies that dragging
// to the right edge with Button1 creates a non-empty selection anchored at the press position.
func TestIntegrationPhase11Batch3DragToRightEdgeExtendsSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 5
	il.scrollOffset = 0

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 9, Y: 0, Button: tcell.Button1}})

	selStart, selEnd := il.Selection()
	if selStart == selEnd {
		t.Error("selection should be non-empty after drag to right edge")
	}
	if selStart != 5 {
		t.Errorf("selStart = %d, want 5 (the drag anchor from the initial press)", selStart)
	}
}

// TestIntegrationPhase11Batch3DragToRightEdgeSelectionIsNonEmpty is a focused check
// that the selection is non-empty after a right-edge scroll event.
func TestIntegrationPhase11Batch3DragToRightEdgeSelectionIsNonEmpty(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 5
	il.scrollOffset = 0

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 9, Y: 0, Button: tcell.Button1}})

	if !il.hasSelection() {
		t.Error("hasSelection() should be true after dragging to the right edge with text overflow")
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: Mouse drag to left edge with scrollOffset > 0 — scrollOffset decreases,
// selection extends
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch3DragToLeftEdgeDecreasesScrollOffset verifies that when
// the user drags to column 0 (left edge) with Button1 held during a drag and
// scrollOffset > 0, scrollOffset decreases by 1.
//
// Scenario: set scrollOffset=5, press at local X=3 (absolute text pos 8), drag to X=0
// (left edge) → scrollOffset goes from 5 to 4.
func TestIntegrationPhase11Batch3DragToLeftEdgeDecreasesScrollOffset(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 8
	il.scrollOffset = 5

	// Start drag at local X=3 (scrollOffset=5 → text pos 3+5=8).
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})
	if !il.dragging {
		t.Fatal("precondition: dragging should be true after press")
	}

	// Drag to left edge (X=0).
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}})

	if il.scrollOffset != 4 {
		t.Errorf("scrollOffset = %d after drag to left edge, want 4", il.scrollOffset)
	}
}

// TestIntegrationPhase11Batch3DragToLeftEdgeExtendsSelection verifies that dragging
// to the left edge extends the selection from the drag anchor.
func TestIntegrationPhase11Batch3DragToLeftEdgeExtendsSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 8
	il.scrollOffset = 5

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}})

	selStart, selEnd := il.Selection()
	if selStart == selEnd {
		t.Error("selection should be non-empty after drag to left edge")
	}
}

// TestIntegrationPhase11Batch3DragToLeftEdgeDoesNotScrollBelowZero verifies that
// repeated drags to the left edge do not push scrollOffset below zero.
func TestIntegrationPhase11Batch3DragToLeftEdgeDoesNotScrollBelowZero(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 3
	il.scrollOffset = 0

	// Start drag.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})

	// Drag to left edge when already at scrollOffset=0 — must not go negative.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}})

	if il.scrollOffset < 0 {
		t.Errorf("scrollOffset = %d, must not be negative after drag to left edge at scrollOffset=0", il.scrollOffset)
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: After edge scroll, cursor position is valid and within text bounds
// ---------------------------------------------------------------------------

// TestIntegrationPhase11Batch3AfterRightEdgeScrollCursorIsValid verifies that after
// a right-edge auto-scroll event, the cursor position is a valid rune index (>= 0
// and <= len(text)).
func TestIntegrationPhase11Batch3AfterRightEdgeScrollCursorIsValid(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop") // 16 chars
	il.cursorPos = 5
	il.scrollOffset = 0

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 9, Y: 0, Button: tcell.Button1}})

	pos := il.CursorPos()
	textLen := len([]rune(il.Text()))
	if pos < 0 || pos > textLen {
		t.Errorf("CursorPos() = %d after right-edge scroll is out of bounds [0, %d]", pos, textLen)
	}
}

// TestIntegrationPhase11Batch3AfterLeftEdgeScrollCursorIsValid verifies that after
// a left-edge auto-scroll event, the cursor position is a valid rune index (>= 0
// and <= len(text)).
func TestIntegrationPhase11Batch3AfterLeftEdgeScrollCursorIsValid(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 8
	il.scrollOffset = 5

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}})

	pos := il.CursorPos()
	textLen := len([]rune(il.Text()))
	if pos < 0 || pos > textLen {
		t.Errorf("CursorPos() = %d after left-edge scroll is out of bounds [0, %d]", pos, textLen)
	}
}

// TestIntegrationPhase11Batch3AfterMultipleRightScrollsCursorStaysInBounds verifies
// that repeated right-edge drags never push cursorPos past the end of the text.
func TestIntegrationPhase11Batch3AfterMultipleRightScrollsCursorStaysInBounds(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop") // 16 chars
	il.cursorPos = 5
	il.scrollOffset = 0

	// Start drag.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}})

	// Repeatedly drag to right edge — more times than necessary to exhaust the scroll range.
	textLen := len([]rune(il.Text()))
	for i := 0; i < 20; i++ {
		il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 9, Y: 0, Button: tcell.Button1}})
	}

	pos := il.CursorPos()
	if pos < 0 || pos > textLen {
		t.Errorf("CursorPos() = %d after repeated right-edge scrolls is out of bounds [0, %d]", pos, textLen)
	}
}

// TestIntegrationPhase11Batch3AfterEdgeScrollCursorPositionIsNonNegative is a focused
// guard that cursor never goes negative after any edge scroll sequence.
func TestIntegrationPhase11Batch3AfterEdgeScrollCursorPositionIsNonNegative(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 3
	il.scrollOffset = 0

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}})
	// Try left edge even though scrollOffset is 0.
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}})

	if pos := il.CursorPos(); pos < 0 {
		t.Errorf("CursorPos() = %d is negative after left-edge drag with scrollOffset=0", pos)
	}
}

// TestIntegrationPhase11Batch3ScrollOffsetConsistentWithCursorAfterRightEdge verifies
// that after a right-edge scroll, the cursor remains within the visible window
// (scrollOffset <= cursorPos < scrollOffset+width).
func TestIntegrationPhase11Batch3ScrollOffsetConsistentWithCursorAfterRightEdge(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop") // 16 chars
	il.cursorPos = 5
	il.scrollOffset = 0

	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}})
	il.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 9, Y: 0, Button: tcell.Button1}})

	pos := il.CursorPos()
	so := il.scrollOffset
	w := 10
	if pos < so || pos >= so+w {
		t.Errorf("after right-edge scroll: cursorPos=%d is outside visible window [%d, %d)",
			pos, so, so+w)
	}
}
