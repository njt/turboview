package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// --- Keyboard (Alt+shortcut) guard ---

// TestLabelAltShortcutSelectableLinkSetsFocus verifies that when the linked view
// has OfSelectable, Alt+shortcut moves focus to the link (regression guard).
// Spec: "When the linked view IS selectable (has OfSelectable), the existing
// behavior is unchanged: focus moves to the link."
func TestLabelAltShortcutSelectableLinkSetsFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView() // has OfSelectable set
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other) // other gets focus last

	if g.FocusedChild() != other {
		t.Fatalf("precondition: FocusedChild() = %v, want other", g.FocusedChild())
	}
	if !linked.HasOption(OfSelectable) {
		t.Fatalf("precondition: linked must have OfSelectable")
	}

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() != linked {
		t.Errorf("FocusedChild() = %v, want linked after Alt+N when link is selectable", g.FocusedChild())
	}
}

// TestLabelAltShortcutNonSelectableLinkDoesNotFocus verifies that when the linked
// view does NOT have OfSelectable, Alt+shortcut does not change focus to the link.
// Spec: "When the linked view is NOT selectable (OfSelectable is false),
// Alt+shortcut clears the event but focus does NOT change."
func TestLabelAltShortcutNonSelectableLinkDoesNotFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	linked.SetOptions(OfSelectable, false) // make it non-selectable
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other) // other gets focus last

	if g.FocusedChild() != other {
		t.Fatalf("precondition: FocusedChild() = %v, want other", g.FocusedChild())
	}
	if linked.HasOption(OfSelectable) {
		t.Fatalf("precondition: linked must NOT have OfSelectable")
	}

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() == linked {
		t.Errorf("FocusedChild() = linked after Alt+N when link is non-selectable; focus must not change")
	}
}

// TestLabelAltShortcutNonSelectableLinkClearsEvent verifies that even when the
// linked view does NOT have OfSelectable, Alt+shortcut still clears the event.
// Spec: "When the linked view is NOT selectable (OfSelectable is false),
// Alt+shortcut clears the event but focus does NOT change."
func TestLabelAltShortcutNonSelectableLinkClearsEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	linked.SetOptions(OfSelectable, false) // make it non-selectable
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("event was not cleared after Alt+N on label with non-selectable link; event must still be consumed")
	}
}

// TestLabelAltShortcutGuardFalsification verifies that an implementation which
// skips the OfSelectable guard entirely (always focuses the link) would be caught.
// Specifically: when link is non-selectable, focus must stay on `other`, not move
// to `linked`. A guard-skipping implementation would fail this test.
// Spec: "Before calling owner.SetFocusedChild(l.link) in response to Alt+shortcut,
// Label must check l.link.HasOption(OfSelectable)."
func TestLabelAltShortcutGuardFalsification(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	linked.SetOptions(OfSelectable, false)
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	// An implementation that skips the guard would call SetFocusedChild(linked),
	// making FocusedChild() == linked. That must not happen.
	if g.FocusedChild() == linked {
		t.Errorf("FocusedChild() = linked; guard was skipped — non-selectable link must never receive focus via Alt+shortcut")
	}
	// Confirm the other view still holds focus.
	if g.FocusedChild() != other {
		t.Errorf("FocusedChild() = %v, want other; focus must remain on the previously focused view", g.FocusedChild())
	}
}

// --- Mouse (Button1 click) guard ---

// TestLabelClickButton1SelectableLinkSetsFocus verifies that a Button1 click on
// a Label whose link has OfSelectable focuses the link (regression guard).
// Spec: "When the linked view IS selectable (has OfSelectable), the existing
// behavior is unchanged: focus moves to the link."
func TestLabelClickButton1SelectableLinkSetsFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView() // has OfSelectable set
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other) // other gets focus last

	if g.FocusedChild() != other {
		t.Fatalf("precondition: FocusedChild() = %v, want other", g.FocusedChild())
	}
	if !linked.HasOption(OfSelectable) {
		t.Fatalf("precondition: linked must have OfSelectable")
	}

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.Button1},
	}
	label.HandleEvent(ev)

	if g.FocusedChild() != linked {
		t.Errorf("FocusedChild() = %v, want linked after Button1 click when link is selectable", g.FocusedChild())
	}
}

// TestLabelClickButton1NonSelectableLinkDoesNotFocus verifies that a Button1 click
// on a Label whose link does NOT have OfSelectable does not move focus to the link.
// Spec: "When the linked view is NOT selectable, mouse Button1 click clears the
// event but focus does NOT change."
func TestLabelClickButton1NonSelectableLinkDoesNotFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	linked.SetOptions(OfSelectable, false) // make it non-selectable
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other) // other gets focus last

	if g.FocusedChild() != other {
		t.Fatalf("precondition: FocusedChild() = %v, want other", g.FocusedChild())
	}
	if linked.HasOption(OfSelectable) {
		t.Fatalf("precondition: linked must NOT have OfSelectable")
	}

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.Button1},
	}
	label.HandleEvent(ev)

	if g.FocusedChild() == linked {
		t.Errorf("FocusedChild() = linked after Button1 click when link is non-selectable; focus must not change")
	}
}

// TestLabelClickButton1NonSelectableLinkClearsEvent verifies that even when the
// linked view does NOT have OfSelectable, a Button1 click still clears the event.
// Spec: "When the linked view is NOT selectable, mouse Button1 click clears the
// event but focus does NOT change."
func TestLabelClickButton1NonSelectableLinkClearsEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	linked.SetOptions(OfSelectable, false) // make it non-selectable
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
		t.Errorf("event was not cleared after Button1 click on label with non-selectable link; event must still be consumed")
	}
}
