package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestLabelClickButton1WithLinkSetsFocus verifies that a Button1 mouse click on
// a Label with a linked view calls owner.SetFocusedChild(link).
func TestLabelClickButton1WithLinkSetsFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other) // other gets focus last

	if g.FocusedChild() != other {
		t.Fatalf("precondition: FocusedChild() = %v, want other", g.FocusedChild())
	}

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.Button1},
	}
	label.HandleEvent(ev)

	if g.FocusedChild() != linked {
		t.Errorf("FocusedChild() = %v, want linked after Button1 click on label", g.FocusedChild())
	}
}

// TestLabelClickButton1WithLinkClearsEvent verifies that the mouse event is cleared
// after a Button1 click on a label with a linked view.
func TestLabelClickButton1WithLinkClearsEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.Button1},
	}
	label.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("event was not cleared after Button1 click on label with link")
	}
}

// TestLabelClickNilLinkDoesNotPanic verifies that a mouse click on a Label with
// no linked view does not panic and does not clear the event.
func TestLabelClickNilLinkDoesNotPanic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", nil)
	g.Insert(label)

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.Button1},
	}
	// Must not panic.
	label.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("event was cleared when label has no link; mouse events should pass through")
	}
}

// TestLabelClickNoOwnerDoesNotPanic verifies that a Button1 click on a label
// with a link but no owner does not panic.
func TestLabelClickNoOwnerDoesNotPanic(t *testing.T) {
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	// label has no owner (not inserted into a group)

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.Button1},
	}
	// Must not panic.
	label.HandleEvent(ev)
}

// TestLabelClickButton2WithLinkDoesNotFocus verifies that a non-Button1 mouse
// click does not activate the label's link focus behavior.
func TestLabelClickButton2WithLinkDoesNotFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.Button2},
	}
	label.HandleEvent(ev)

	if g.FocusedChild() == linked {
		t.Errorf("FocusedChild() = linked after Button2 click; only Button1 should focus the link")
	}
	if ev.IsCleared() {
		t.Errorf("event was cleared on Button2 click; only Button1 should clear the event")
	}
}
