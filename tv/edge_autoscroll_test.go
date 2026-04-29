package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestEdgeAutoScrollRightIncreasesScrollOffset(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop") // 16 chars, width 10
	il.cursorPos = 5
	il.scrollOffset = 0

	// Start drag at pos 5
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(press)

	// Drag to right edge (X=9, which is w-1)
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 9, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(drag)

	if il.scrollOffset != 1 {
		t.Errorf("scrollOffset = %d, want 1 after dragging to right edge", il.scrollOffset)
	}
}

func TestEdgeAutoScrollLeftDecreasesScrollOffset(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 8
	il.scrollOffset = 5

	// Start drag at local X=3 (absolute 3, scrollOffset=5 → pos 8)
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(press)

	// Drag to left edge (X=0)
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(drag)

	if il.scrollOffset != 4 {
		t.Errorf("scrollOffset = %d, want 4 after dragging to left edge", il.scrollOffset)
	}
}

func TestEdgeAutoScrollLeftDoesNotScrollPastZero(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 3
	il.scrollOffset = 0

	// Start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(press)

	// Drag to left edge — scrollOffset already 0
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(drag)

	if il.scrollOffset != 0 {
		t.Errorf("scrollOffset = %d, want 0 (should not go below 0)", il.scrollOffset)
	}
}

func TestEdgeAutoScrollExtendsSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 5
	il.scrollOffset = 0

	// Start drag at pos 5
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(press)

	// Drag to right edge
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 9, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(drag)

	selStart, selEnd := il.Selection()
	if selStart == selEnd {
		t.Error("selection should be non-empty after edge auto-scroll")
	}
	if selStart != 5 {
		t.Errorf("selStart = %d, want 5 (drag anchor)", selStart)
	}
}

func TestEdgeAutoScrollMultipleEventsAccumulate(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop") // 16 chars
	il.cursorPos = 5
	il.scrollOffset = 0

	// Start drag at pos 5
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(press)

	// Multiple auto-scroll events at right edge
	for i := 0; i < 3; i++ {
		drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 9, Y: 0, Button: tcell.Button1}}
		il.HandleEvent(drag)
	}

	if il.scrollOffset < 3 {
		t.Errorf("scrollOffset = %d, want >= 3 after 3 auto-scroll events", il.scrollOffset)
	}
}

func TestEdgeAutoScrollRightStopsWhenTextEnds(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop") // 16 chars
	il.cursorPos = 5
	il.scrollOffset = 0

	// Start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1}}
	il.HandleEvent(press)

	// Scroll 10 times — should stop when all text is visible
	for i := 0; i < 10; i++ {
		drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 9, Y: 0, Button: tcell.Button1}}
		il.HandleEvent(drag)
	}

	// Max scrollOffset: 16 - 10 = 6
	if il.scrollOffset > 6 {
		t.Errorf("scrollOffset = %d, want <= 6 (should not scroll past text end)", il.scrollOffset)
	}
}
