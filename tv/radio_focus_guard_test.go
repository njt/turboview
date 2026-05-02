package tv

// radio_focus_guard_test.go — Tests that RadioButtons does NOT consume
// Down/Up/Right/Left arrow events when it is not the focused widget (SfSelected=false).
//
// RadioButtons has OfPreProcess=true, which means it receives keyboard events
// in Phase1 of Group dispatch, BEFORE the focused child sees them (Phase2)
// and BEFORE OfPostProcess widgets like History see them (Phase3).
// If RadioButtons unconditionally clears Down/Up/Right/Left it prevents History (Phase3)
// from ever opening the dropdown. The fix: only handle these keys when the
// RadioButtons itself has SfSelected (i.e. it is the focused widget in the
// enclosing group).

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestRadioArrowsDownNotConsumedWhenNotFocused verifies that a Down arrow
// event is NOT cleared by RadioButtons when SfSelected is false (not focused).
func TestRadioArrowsDownNotConsumedWhenNotFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Do NOT set SfSelected — RadioButtons is not the focused widget.

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Down arrow was consumed by unfocused RadioButtons (SfSelected=false); " +
			"it must pass through so OfPostProcess widgets (e.g. History) can see it")
	}
}

// TestRadioArrowsUpNotConsumedWhenNotFocused verifies that an Up arrow
// event is NOT cleared by RadioButtons when SfSelected is false.
func TestRadioArrowsUpNotConsumedWhenNotFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(1)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	rbs.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Up arrow was consumed by unfocused RadioButtons (SfSelected=false); " +
			"it must pass through")
	}
}

// TestRadioArrowsDownConsumedWhenFocused verifies that Down IS still
// consumed (and moves selection) when RadioButtons has SfSelected=true.
func TestRadioArrowsDownConsumedWhenFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetState(SfSelected, true) // RadioButtons is the focused widget

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Down arrow was NOT consumed by focused RadioButtons (SfSelected=true); it should be")
	}
	if !rbs.Item(1).Selected() {
		t.Error("Down arrow: expected Item(1) to be selected after Down from Item(0)")
	}
}

// TestRadioArrowsUpConsumedWhenFocused verifies that Up IS still
// consumed (and moves selection) when RadioButtons has SfSelected=true.
func TestRadioArrowsUpConsumedWhenFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true) // RadioButtons is the focused widget

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Up arrow was NOT consumed by focused RadioButtons (SfSelected=true); it should be")
	}
	if !rbs.Item(1).Selected() {
		t.Error("Up arrow: expected Item(1) to be selected after Up from Item(2)")
	}
}
