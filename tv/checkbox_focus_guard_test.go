package tv

// checkbox_focus_guard_test.go — Tests that CheckBoxes does NOT consume
// Down/Up arrow events when it is not the focused widget (SfSelected=false).
//
// CheckBoxes has OfPreProcess=true, which means it receives keyboard events
// in Phase1 of Group dispatch, BEFORE the focused child sees them (Phase2)
// and BEFORE OfPostProcess widgets like History see them (Phase3).
// If CheckBoxes unconditionally clears Down/Up it prevents History (Phase3)
// from ever opening the dropdown. The fix: only handle Down/Up when the
// CheckBoxes itself has SfSelected (i.e. it is the focused widget in the
// enclosing group).

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestCheckBoxArrowsDownNotConsumedWhenNotFocused verifies that a Down arrow
// event is NOT cleared by CheckBoxes when SfSelected is false (not focused).
// This is the Phase1/Phase3 interaction scenario: CheckBoxes is a sibling of
// History; when InputLine is focused, CheckBoxes must not steal Down.
func TestCheckBoxArrowsDownNotConsumedWhenNotFocused(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Do NOT set SfSelected — CheckBoxes is not the focused widget.

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cbs.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Down arrow was consumed by unfocused CheckBoxes (SfSelected=false); " +
			"it must pass through so OfPostProcess widgets (e.g. History) can see it")
	}
}

// TestCheckBoxArrowsUpNotConsumedWhenNotFocused verifies that an Up arrow
// event is NOT cleared by CheckBoxes when SfSelected is false (not focused).
func TestCheckBoxArrowsUpNotConsumedWhenNotFocused(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Do NOT set SfSelected — CheckBoxes is not the focused widget.

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	cbs.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Up arrow was consumed by unfocused CheckBoxes (SfSelected=false); " +
			"it must pass through so other widgets can see it")
	}
}

// TestCheckBoxArrowsDownConsumedWhenFocused verifies that Down IS still
// consumed (and moves focus) when CheckBoxes has SfSelected=true.
func TestCheckBoxArrowsDownConsumedWhenFocused(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true) // CheckBoxes is the focused widget

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Down arrow was NOT consumed by focused CheckBoxes (SfSelected=true); it should be")
	}
	if cbs.FocusedChild() != cbs.Item(1) {
		t.Errorf("Down arrow: FocusedChild() = %v, want Item(1)", cbs.FocusedChild())
	}
}

// TestCheckBoxArrowsUpConsumedWhenFocused verifies that Up IS still
// consumed (and moves focus) when CheckBoxes has SfSelected=true.
func TestCheckBoxArrowsUpConsumedWhenFocused(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(2))
	cbs.SetState(SfSelected, true) // CheckBoxes is the focused widget

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	cbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Up arrow was NOT consumed by focused CheckBoxes (SfSelected=true); it should be")
	}
	if cbs.FocusedChild() != cbs.Item(1) {
		t.Errorf("Up arrow: FocusedChild() = %v, want Item(1)", cbs.FocusedChild())
	}
}
