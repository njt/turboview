package tv

// list_viewer_space_enter_test.go — Tests for Task 2: Space and Enter key to select (spec 7.1, 7.2).
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Spec 7.1: OnSelect should fire on: Space bar press, Double-click, Enter key.
// Spec 7.2: When Space is pressed and a valid item is focused (count > 0, has SfSelected),
//           call OnSelect(focused) and consume the event.
//
// Test organisation:
//   Section 1  — Space key fires OnSelect
//   Section 2  — Space key edge cases (nil callback, empty list, not focused)
//   Section 3  — Space key receives the correct index (navigate then select)
//   Section 4  — Enter key fires OnSelect
//   Section 5  — Enter key edge cases (nil callback, empty list, not focused)
//   Section 6  — Enter key receives the correct index
//   Section 7  — Falsification tests (Space vs Down distinction; navigate+select pipeline)

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// spaceKeyEv creates a keyboard event for the Space bar.
// Space is delivered by tcell as KeyRune with Rune == ' '.
func spaceKeyEv() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' ', Modifiers: tcell.ModNone}}
}

// ---------------------------------------------------------------------------
// Section 1 — Space key fires OnSelect
// ---------------------------------------------------------------------------

// TestSpaceFiresOnSelectWhenFocused verifies Space calls OnSelect with the current index.
// Spec 7.2: "when Space is pressed and a valid item is focused, call OnSelect(focused)"
func TestSpaceFiresOnSelectWhenFocused(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := spaceKeyEv()
	lv.HandleEvent(ev)

	if !called {
		t.Error("spec 7.2: Space must call OnSelect when widget is focused and count > 0")
	}
}

// TestSpaceFiresOnSelectWithCorrectIndex verifies Space passes the currently selected index.
// Spec 7.2: "call OnSelect(focused)" — the focused item index is what gets reported.
func TestSpaceFiresOnSelectWithCorrectIndex(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(1)
	var got int
	lv.OnSelect = func(index int) { got = index }

	ev := spaceKeyEv()
	lv.HandleEvent(ev)

	if got != 1 {
		t.Errorf("spec 7.2: Space must pass selected index to OnSelect; got %d, want 1", got)
	}
}

// TestSpaceConsumesEvent verifies Space bar consumes the event when focused and count > 0.
// Spec 7.2: "consume the event"
func TestSpaceConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	ev := spaceKeyEv()
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.2: Space must consume (clear) the event when focused and count > 0")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Space key edge cases
// ---------------------------------------------------------------------------

// TestSpaceWithNilOnSelectDoesNotPanic verifies Space with nil OnSelect does not panic.
// Spec 7.2: event is still consumed; nil OnSelect is safe.
func TestSpaceWithNilOnSelectDoesNotPanic(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	// OnSelect is nil by default

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("spec 7.2: Space with nil OnSelect panicked: %v", r)
		}
	}()

	ev := spaceKeyEv()
	lv.HandleEvent(ev)
}

// TestSpaceWithNilOnSelectStillConsumesEvent verifies Space still consumes the event
// even when OnSelect is nil.
// Spec 7.2: "still consume event" — the action is confirmed regardless of callback presence.
func TestSpaceWithNilOnSelectStillConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	// OnSelect is nil by default

	ev := spaceKeyEv()
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.2: Space with nil OnSelect must still consume (clear) the event")
	}
}

// TestSpaceWithEmptyListDoesNotCallOnSelect verifies Space with count == 0 does not call OnSelect.
// Spec 7.2: "a valid item is focused (count > 0)" — empty list has no valid item.
func TestSpaceWithEmptyListDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused([]string{})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := spaceKeyEv()
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.2: Space with empty list must NOT call OnSelect (no valid item)")
	}
}

// TestSpaceWithEmptyListStillConsumesEvent verifies Space still consumes event on empty list.
// Spec 7.2: "still consume event" — the key is recognised even if no action is taken.
func TestSpaceWithEmptyListStillConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{})

	ev := spaceKeyEv()
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.2: Space on empty focused list must still consume (clear) the event")
	}
}

// TestSpaceWhenNotFocusedDoesNotConsumeEvent verifies Space is not handled when not focused.
// Spec 7.2: "a valid item is focused (has SfSelected)" — without focus the event passes through.
func TestSpaceWhenNotFocusedDoesNotConsumeEvent(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"}) // not focused: no SfSelected
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := spaceKeyEv()
	lv.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("spec 7.2: Space without focus must NOT consume the event (pass-through)")
	}
	if called {
		t.Error("spec 7.2: Space without focus must NOT call OnSelect")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Space receives the correct index (navigate then select)
// ---------------------------------------------------------------------------

// TestSpaceAfterNavigatingToIndex3FiresWithIndex3 verifies the index passed to OnSelect
// reflects navigation that happened before the Space press.
// Spec 7.2: OnSelect receives the currently focused item at the moment Space is pressed.
func TestSpaceAfterNavigatingToIndex3FiresWithIndex3(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e"})
	lv.SetSelected(3)

	var got int
	lv.OnSelect = func(index int) { got = index }

	ev := spaceKeyEv()
	lv.HandleEvent(ev)

	if got != 3 {
		t.Errorf("spec 7.2: Space after navigating to index 3 must pass 3 to OnSelect; got %d", got)
	}
}

// TestSpaceDoesNotChangeSelectedIndex verifies Space does not move the selection cursor.
// Spec 7.2: Space is a confirm action, not navigation — selected must remain unchanged.
func TestSpaceDoesNotChangeSelectedIndex(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(2)
	lv.OnSelect = func(index int) {}

	before := lv.Selected()

	ev := spaceKeyEv()
	lv.HandleEvent(ev)

	if lv.Selected() != before {
		t.Errorf("spec 7.2: Space must not change selected index; was %d, now %d", before, lv.Selected())
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Enter key fires OnSelect
// ---------------------------------------------------------------------------

// TestEnterFiresOnSelectWhenFocused verifies Enter calls OnSelect with the current index.
// Spec 7.1: "OnSelect should fire on: … Enter key."
func TestEnterFiresOnSelectWhenFocused(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyEnter)
	lv.HandleEvent(ev)

	if !called {
		t.Error("spec 7.1: Enter must call OnSelect when widget is focused and count > 0")
	}
}

// TestEnterFiresOnSelectWithCorrectIndex verifies Enter passes the currently selected index.
// Spec 7.1: "OnSelect should fire on: … Enter key."
func TestEnterFiresOnSelectWithCorrectIndex(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(2)
	var got int
	lv.OnSelect = func(index int) { got = index }

	ev := listKeyEv(tcell.KeyEnter)
	lv.HandleEvent(ev)

	if got != 2 {
		t.Errorf("spec 7.1: Enter must pass selected index to OnSelect; got %d, want 2", got)
	}
}

// TestEnterConsumesEvent verifies Enter consumes the event when focused and count > 0.
// Spec 7.1: Enter handling must consume the event.
func TestEnterConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	ev := listKeyEv(tcell.KeyEnter)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.1: Enter must consume (clear) the event when focused and count > 0")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Enter key edge cases
// ---------------------------------------------------------------------------

// TestEnterWithNilOnSelectDoesNotPanic verifies Enter with nil OnSelect does not panic.
// Spec 7.1: nil OnSelect must be safe for Enter just as for Space.
func TestEnterWithNilOnSelectDoesNotPanic(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	// OnSelect is nil by default

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("spec 7.1: Enter with nil OnSelect panicked: %v", r)
		}
	}()

	ev := listKeyEv(tcell.KeyEnter)
	lv.HandleEvent(ev)
}

// TestEnterWithNilOnSelectStillConsumesEvent verifies Enter still consumes the event
// even when OnSelect is nil.
// Spec 7.1: event consumption is independent of whether a callback is registered.
func TestEnterWithNilOnSelectStillConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	// OnSelect is nil by default

	ev := listKeyEv(tcell.KeyEnter)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.1: Enter with nil OnSelect must still consume (clear) the event")
	}
}

// TestEnterWithEmptyListDoesNotCallOnSelect verifies Enter with count == 0 does not call OnSelect.
// Spec 7.1: no valid item exists in an empty list; OnSelect must not fire.
func TestEnterWithEmptyListDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused([]string{})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyEnter)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.1: Enter with empty list must NOT call OnSelect (no valid item)")
	}
}

// TestEnterWithEmptyListStillConsumesEvent verifies Enter still consumes event on empty list.
// Spec 7.1: the key is recognised even if no action is taken.
func TestEnterWithEmptyListStillConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{})

	ev := listKeyEv(tcell.KeyEnter)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.1: Enter on empty focused list must still consume (clear) the event")
	}
}

// TestEnterWhenNotFocusedDoesNotConsumeEvent verifies Enter is not handled when not focused.
// Spec 7.1: keyboard handling only applies when the widget has SfSelected (is focused).
func TestEnterWhenNotFocusedDoesNotConsumeEvent(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"}) // not focused
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyEnter)
	lv.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("spec 7.1: Enter without focus must NOT consume the event (pass-through)")
	}
	if called {
		t.Error("spec 7.1: Enter without focus must NOT call OnSelect")
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Enter receives the correct index
// ---------------------------------------------------------------------------

// TestEnterAfterNavigatingToIndex2FiresWithIndex2 verifies the index passed to OnSelect
// reflects the selection at the moment Enter is pressed.
// Spec 7.1: Enter confirms the currently focused item.
func TestEnterAfterNavigatingToIndex2FiresWithIndex2(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e"})
	lv.SetSelected(2)

	var got int
	lv.OnSelect = func(index int) { got = index }

	ev := listKeyEv(tcell.KeyEnter)
	lv.HandleEvent(ev)

	if got != 2 {
		t.Errorf("spec 7.1: Enter after navigating to index 2 must pass 2 to OnSelect; got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Falsification tests
// ---------------------------------------------------------------------------

// TestSpaceFiresOnSelectButDownDoesNot proves the distinction between selection and navigation.
// This guards against an implementation that calls OnSelect on every key event or no key event.
// Spec 7.1: "Down arrow changes focused item but does NOT call OnSelect; Space DOES."
func TestSpaceFiresOnSelectButDownDoesNot(t *testing.T) {
	// Down arrow: OnSelect must NOT fire.
	lvNav := newLVFocused([]string{"a", "b", "c"})
	navCalled := false
	lvNav.OnSelect = func(index int) { navCalled = true }

	navEv := listKeyEv(tcell.KeyDown)
	lvNav.HandleEvent(navEv)

	if navCalled {
		t.Error("spec 7.1: Down arrow must NOT call OnSelect — cannot be 'always fire'")
	}

	// Space bar: OnSelect MUST fire.
	lvSel := newLVFocused([]string{"a", "b", "c"})
	selCalled := false
	lvSel.OnSelect = func(index int) { selCalled = true }

	selEv := spaceKeyEv()
	lvSel.HandleEvent(selEv)

	if !selCalled {
		t.Error("spec 7.2: Space must call OnSelect — cannot be 'never fire'")
	}
}

// TestNavigateDownThreeTimesTheSpaceFiresWithIndex3 verifies the full
// navigation-then-selection pipeline: three Down presses followed by Space must
// fire OnSelect with index 3.
// Spec 7.2: OnSelect receives the item focused at the moment Space is pressed.
func TestNavigateDownThreeTimesTheSpaceFiresWithIndex3(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e"})

	// Navigate down three times: 0 → 1 → 2 → 3.
	for i := 0; i < 3; i++ {
		ev := listKeyEv(tcell.KeyDown)
		lv.HandleEvent(ev)
	}

	if lv.Selected() != 3 {
		t.Fatalf("prerequisite failed: after 3 Down presses Selected()=%d, want 3", lv.Selected())
	}

	var got int
	lv.OnSelect = func(index int) { got = index }

	spaceEv := spaceKeyEv()
	lv.HandleEvent(spaceEv)

	if got != 3 {
		t.Errorf("spec 7.2: Space after navigating to index 3 must fire OnSelect(3); got %d", got)
	}
}
