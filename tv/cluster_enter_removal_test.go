package tv

// cluster_enter_removal_test.go — Tests for Task 4: Remove Enter Key from CheckBox
// and RadioButton.
//
// CheckBox.HandleEvent does NOT handle KeyEnter — pressing Enter does not toggle.
// RadioButton.HandleEvent does NOT handle KeyEnter — pressing Enter does not select.
// Space still works (existing behavior).
// Mouse click still works (existing behavior).

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// CheckBox — Enter key removal
// ---------------------------------------------------------------------------

// TestClusterEnterRemovalCheckBoxEnterDoesNotToggle verifies that pressing Enter on
// a CheckBox does NOT toggle its checked state.
// Spec: "CheckBox.HandleEvent does NOT handle KeyEnter."
func TestClusterEnterRemovalCheckBoxEnterDoesNotToggle(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	cb.HandleEvent(ev)

	if cb.Checked() {
		t.Error("Enter key toggled CheckBox; Enter must not toggle after removal")
	}
}

// TestClusterEnterRemovalCheckBoxEnterDoesNotConsumeEvent verifies that pressing
// Enter on a CheckBox does NOT consume the event.
// Spec: "CheckBox.HandleEvent does NOT handle KeyEnter."
func TestClusterEnterRemovalCheckBoxEnterDoesNotConsumeEvent(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	cb.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Enter key was consumed by CheckBox; Enter must not be consumed after removal")
	}
}

// TestClusterEnterRemovalCheckBoxSpaceStillToggles verifies Space still toggles the
// CheckBox (existing behavior preserved).
// Spec: "Space still works (existing behavior)."
func TestClusterEnterRemovalCheckBoxSpaceStillToggles(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	cb.HandleEvent(ev)

	if !cb.Checked() {
		t.Error("Space did not toggle CheckBox; Space must still toggle after Enter removal")
	}
}

// TestClusterEnterRemovalCheckBoxMouseClickStillToggles verifies Button1 click still
// toggles the CheckBox (existing behavior preserved).
// Spec: "Mouse click still works (existing behavior)."
func TestClusterEnterRemovalCheckBoxMouseClickStillToggles(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 10, 1), "OK")

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	cb.HandleEvent(ev)

	if !cb.Checked() {
		t.Error("Button1 click did not toggle CheckBox; mouse click must still work after Enter removal")
	}
}

// ---------------------------------------------------------------------------
// RadioButton — Enter key removal
// ---------------------------------------------------------------------------

// TestClusterEnterRemovalRadioButtonEnterDoesNotSelect verifies that pressing Enter
// on a RadioButton does NOT select it.
// Spec: "RadioButton.HandleEvent does NOT handle KeyEnter."
func TestClusterEnterRemovalRadioButtonEnterDoesNotSelect(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")
	rb.SetSelected(false)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	rb.HandleEvent(ev)

	if rb.Selected() {
		t.Error("Enter key selected RadioButton; Enter must not select after removal")
	}
}

// TestClusterEnterRemovalRadioButtonEnterDoesNotConsumeEvent verifies that pressing
// Enter on a RadioButton does NOT consume the event.
// Spec: "RadioButton.HandleEvent does NOT handle KeyEnter."
func TestClusterEnterRemovalRadioButtonEnterDoesNotConsumeEvent(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	rb.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Enter key was consumed by RadioButton; Enter must not be consumed after removal")
	}
}

// TestClusterEnterRemovalRadioButtonSpaceStillSelects verifies Space still selects
// the RadioButton (existing behavior preserved).
// Spec: "Space still works (existing behavior)."
func TestClusterEnterRemovalRadioButtonSpaceStillSelects(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")
	rb.SetSelected(false)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	rb.HandleEvent(ev)

	if !rb.Selected() {
		t.Error("Space did not select RadioButton; Space must still select after Enter removal")
	}
}

// TestClusterEnterRemovalRadioButtonMouseClickStillSelects verifies Button1 click
// still selects the RadioButton (existing behavior preserved).
// Spec: "Mouse click still works (existing behavior)."
func TestClusterEnterRemovalRadioButtonMouseClickStillSelects(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")
	rb.SetSelected(false)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 0, Button: tcell.Button1}}
	rb.HandleEvent(ev)

	if !rb.Selected() {
		t.Error("Button1 click did not select RadioButton; mouse click must still work after Enter removal")
	}
}
