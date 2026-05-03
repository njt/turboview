package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// TestMemoHasOfFirstClick verifies that Memo has OfFirstClick set so clicking
// an unfocused Memo both focuses it and positions the cursor.
func TestMemoHasOfFirstClick(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 5))
	if !m.HasOption(OfFirstClick) {
		t.Error("Memo should have OfFirstClick set")
	}
}

// TestMemoFirstClickPositionsCursor verifies that clicking an unfocused Memo
// focuses it AND processes the click (positions cursor), not just focuses.
func TestMemoFirstClickPositionsCursor(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	g.scheme = theme.BorlandBlue

	btn := NewButton(NewRect(0, 0, 10, 2), "OK", CmOK)
	m := NewMemo(NewRect(0, 3, 20, 5))
	m.SetText("Hello World")
	g.Insert(btn)
	g.Insert(m)

	// Focus the button, not the memo.
	g.SetFocusedChild(btn)
	if m.HasState(SfSelected) {
		t.Fatal("Memo should not be focused initially")
	}

	// Click on the Memo at position (5, 3) — middle of "Hello World".
	ev := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:          5,
			Y:          3,
			Button:     tcell.Button1,
			ClickCount: 1,
		},
	}
	g.HandleEvent(ev)

	if !m.HasState(SfSelected) {
		t.Logf("Memo bounds: %v, owner=%v, self=%v, OfSelectable=%v, OfFirstClick=%v",
			m.Bounds(), m.Owner() != nil, m.self != nil, m.HasOption(OfSelectable), m.HasOption(OfFirstClick))
		t.Logf("Event cleared: %v", ev.IsCleared())
		t.Error("Memo should be focused after click")
	}

	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("cursor at (%d,%d), want (0,5)", row, col)
	}
}

// TestMemoMouseWheelScrolls verifies that mouse wheel events scroll the Memo.
func TestMemoMouseWheelScrolls(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6")

	// Wheel down should increase deltaY.
	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 5, Y: 1, Button: tcell.WheelDown},
	}
	m.HandleEvent(ev)

	_, deltaY := m.CursorPos()
	_ = deltaY
	// Check deltaY directly.
	if m.deltaY == 0 {
		t.Error("WheelDown should scroll Memo (deltaY > 0)")
	}
}

// TestInputLineHasOfFirstClick verifies that InputLine has OfFirstClick set.
func TestInputLineHasOfFirstClick(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 40)
	if !il.HasOption(OfFirstClick) {
		t.Error("InputLine should have OfFirstClick set")
	}
}
