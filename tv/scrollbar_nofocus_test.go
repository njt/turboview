package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// scrollbar_nofocus_test.go — Tests for correct ScrollBar focus behavior.
//
// In original Turbo Vision, TScrollBar does NOT have ofSelectable set.
// Tab never lands on it. The associated TListViewer owns all keyboard
// navigation. The scrollbar is a passive visual indicator and mouse target.

// TestTabSkipsScrollBarInWindow verifies that Tab navigation in a Window
// skips the scrollbar entirely. When a Window contains a ListViewer and
// a ScrollBar, Tab from the ListViewer should NOT land on the scrollbar.
func TestTabSkipsScrollBarInWindow(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 40, 15), "Test")

	items := []string{"A", "B", "C"}
	lv := NewListViewer(NewRect(0, 0, 30, 10), NewStringList(items))
	sb := NewScrollBar(NewRect(30, 0, 1, 10), Vertical)
	lv.SetScrollBar(sb)
	btn := NewButton(NewRect(0, 11, 10, 2), "OK", CmOK)

	win.Insert(lv)
	win.Insert(sb)
	win.Insert(btn)

	// Focus the listviewer
	win.SetFocusedChild(lv)
	if !lv.HasState(SfSelected) {
		t.Fatal("ListViewer should have SfSelected after SetFocusedChild")
	}

	// Tab should skip scrollbar and land on button (or wrap to next selectable)
	tabEvent := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab}}
	win.HandleEvent(tabEvent)

	if sb.HasState(SfSelected) {
		t.Error("Tab should NOT land on ScrollBar — scrollbar must not be focusable")
	}
}

// TestListViewerKeepsFocusNotScrollBar verifies that in a window with both
// a ListViewer and ScrollBar, the ListViewer receives focus, not the scrollbar.
func TestListViewerKeepsFocusNotScrollBar(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 40, 15), "Test")

	items := []string{"A", "B", "C", "D", "E"}
	lv := NewListViewer(NewRect(0, 0, 30, 3), NewStringList(items))
	sb := NewScrollBar(NewRect(30, 0, 1, 3), Vertical)
	lv.SetScrollBar(sb)

	win.Insert(lv)
	win.Insert(sb)

	// Focus the listviewer
	win.SetFocusedChild(lv)

	// Down arrow should move selection in listviewer (not be eaten by scrollbar)
	downEvent := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	win.HandleEvent(downEvent)

	if lv.Selected() != 1 {
		t.Errorf("Down arrow with ListViewer focused: Selected() = %d, want 1", lv.Selected())
	}
	if sb.HasState(SfSelected) {
		t.Error("ScrollBar should never gain SfSelected through normal interaction")
	}
}

// TestScrollBarMouseStillWorksWithoutFocus verifies that clicking the scrollbar
// still works even though it's not focusable. Mouse events are routed by
// position, not focus.
func TestScrollBarMouseStillWorksWithoutFocus(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	// Click the down arrow (bottom of scrollbar)
	ev := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:      0,
			Y:      9, // bottom arrow
			Button: tcell.Button1,
		},
	}
	sb.HandleEvent(ev)

	if sb.Value() != 51 {
		t.Errorf("Mouse click on scrollbar arrow: Value() = %d, want 51 (mouse should work without OfSelectable)", sb.Value())
	}
}
