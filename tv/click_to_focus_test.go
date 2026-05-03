package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func mouseClick(x, y int) *Event {
	return &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:          x,
			Y:          y,
			Button:     tcell.Button1,
			ClickCount: 1,
		},
	}
}

func TestListViewerClickAcquiresFocus(t *testing.T) {
	btn := NewButton(NewRect(0, 0, 10, 2), "OK", CmOK)
	lv := NewListViewer(NewRect(0, 3, 20, 5), NewStringList([]string{"A", "B"}))

	g := NewGroup(NewRect(0, 0, 30, 10))
	g.Insert(btn)
	g.Insert(lv)
	g.SetFocusedChild(btn)

	if g.FocusedChild() != btn {
		t.Fatal("precondition: button should be focused")
	}

	ev := mouseClick(5, 4) // within lv's bounds (y=3..7)
	g.HandleEvent(ev)

	if g.FocusedChild() != lv {
		t.Error("ListViewer should be focused after click")
	}
}

func TestListViewerFirstClickSelectsItem(t *testing.T) {
	lv := NewListViewer(NewRect(0, 0, 20, 5), NewStringList([]string{"A", "B", "C"}))
	btn := NewButton(NewRect(0, 6, 10, 2), "X", CmOK)

	g := NewGroup(NewRect(0, 0, 30, 10))
	g.Insert(lv)
	g.Insert(btn)
	g.SetFocusedChild(btn)

	ev := mouseClick(5, 2) // row 2 within lv → item index 2
	g.HandleEvent(ev)

	if lv.Selected() != 2 {
		t.Errorf("first click should select item 2, got %d", lv.Selected())
	}
}

func TestListBoxClickAcquiresFocus(t *testing.T) {
	btn := NewButton(NewRect(0, 0, 10, 2), "OK", CmOK)
	lb := NewStringListBox(NewRect(0, 3, 20, 5), []string{"A", "B", "C"})

	g := NewGroup(NewRect(0, 0, 30, 10))
	g.Insert(btn)
	g.Insert(lb)
	g.SetFocusedChild(btn)

	ev := mouseClick(5, 4) // within lb's bounds
	g.HandleEvent(ev)

	if g.FocusedChild() != lb {
		t.Error("ListBox should be focused after click")
	}
}

func TestCheckBoxesClickAcquiresFocus(t *testing.T) {
	btn := NewButton(NewRect(0, 0, 10, 2), "OK", CmOK)
	cbs := NewCheckBoxes(NewRect(0, 3, 20, 3), []string{"One", "Two", "Three"})

	g := NewGroup(NewRect(0, 0, 30, 10))
	g.Insert(btn)
	g.Insert(cbs)
	g.SetFocusedChild(btn)

	ev := mouseClick(5, 3) // first checkbox row
	g.HandleEvent(ev)

	if g.FocusedChild() != cbs {
		t.Error("CheckBoxes should be focused after click")
	}
}

func TestCheckBoxesClickTogglesOnFirstClick(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"One", "Two", "Three"})
	btn := NewButton(NewRect(0, 4, 10, 2), "OK", CmOK)

	g := NewGroup(NewRect(0, 0, 30, 10))
	g.Insert(cbs)
	g.Insert(btn)
	g.SetFocusedChild(btn)

	ev := mouseClick(5, 0) // click first checkbox
	g.HandleEvent(ev)

	if !cbs.items[0].Checked() {
		t.Error("first click should toggle checkbox")
	}
}

func TestRadioButtonsClickAcquiresFocus(t *testing.T) {
	btn := NewButton(NewRect(0, 0, 10, 2), "OK", CmOK)
	rbs := NewRadioButtons(NewRect(0, 3, 20, 3), []string{"A", "B", "C"})

	g := NewGroup(NewRect(0, 0, 30, 10))
	g.Insert(btn)
	g.Insert(rbs)
	g.SetFocusedChild(btn)

	ev := mouseClick(5, 4) // second radio button row
	g.HandleEvent(ev)

	if g.FocusedChild() != rbs {
		t.Error("RadioButtons should be focused after click")
	}
}

func TestRadioButtonsClickSelectsOnFirstClick(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	btn := NewButton(NewRect(0, 4, 10, 2), "OK", CmOK)

	g := NewGroup(NewRect(0, 0, 30, 10))
	g.Insert(rbs)
	g.Insert(btn)
	g.SetFocusedChild(btn)

	ev := mouseClick(5, 1) // click second radio
	g.HandleEvent(ev)

	if rbs.Value() != 1 {
		t.Errorf("first click should select radio 1, got %d", rbs.Value())
	}
}
