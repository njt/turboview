package tv

// list_viewer_nav_select_test.go — Tests for Phase 13 Task 1: Remove OnSelect from navigation keys.
//
// Written BEFORE any implementation exists; all tests drive spec 7.1.
// Each test has a doc comment citing spec 7.1.
//
// Spec 7.1: OnSelect must NOT fire on every arrow key press. It should only fire on
// Space bar press, double-click, and Enter key. Arrow key navigation (Up, Down,
// PgUp, PgDn, Home, End) changes the focused item and redraws, but does NOT call
// OnSelect. This distinction separates browsing from confirming.
//
// Test organisation:
//   Section 1  — OnSelect NOT fired on navigation keys (confirming tests)
//   Section 2  — Navigation still works (falsification — prove non-trivial absence)
//   Section 3  — Events still consumed by navigation keys
//   Section 4  — Nil OnSelect safety

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Section 1 — OnSelect NOT fired on navigation keys
// ---------------------------------------------------------------------------

// TestNavDownDoesNotCallOnSelect verifies Down arrow does NOT call OnSelect.
// Spec 7.1: "Arrow key navigation changes the focused item and redraws, but does NOT call OnSelect."
func TestNavDownDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.1: OnSelect must NOT fire on Down arrow key navigation")
	}
}

// TestNavUpDoesNotCallOnSelect verifies Up arrow does NOT call OnSelect.
// Spec 7.1: "Arrow key navigation changes the focused item and redraws, but does NOT call OnSelect."
func TestNavUpDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(2)
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyUp)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.1: OnSelect must NOT fire on Up arrow key navigation")
	}
}

// TestNavHomeDoesNotCallOnSelect verifies Home does NOT call OnSelect.
// Spec 7.1: "Home and End change the focused item but do NOT call OnSelect."
func TestNavHomeDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(2)
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.1: OnSelect must NOT fire on Home key navigation")
	}
}

// TestNavEndDoesNotCallOnSelect verifies End does NOT call OnSelect.
// Spec 7.1: "Home and End change the focused item but do NOT call OnSelect."
func TestNavEndDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.1: OnSelect must NOT fire on End key navigation")
	}
}

// TestNavPgDnDoesNotCallOnSelect verifies PgDn does NOT call OnSelect.
// Spec 7.1: "Page navigation (PgUp, PgDn) changes the focused item but does NOT call OnSelect."
func TestNavPgDnDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyPgDn)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.1: OnSelect must NOT fire on PgDn key navigation")
	}
}

// TestNavPgUpDoesNotCallOnSelect verifies PgUp does NOT call OnSelect.
// Spec 7.1: "Page navigation (PgUp, PgDn) changes the focused item but does NOT call OnSelect."
func TestNavPgUpDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	lv.SetSelected(7)
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.1: OnSelect must NOT fire on PgUp key navigation")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Navigation still works (falsification)
// ---------------------------------------------------------------------------

// TestNavDownStillMovesSelection verifies Down arrow still moves selected from 0 to 1.
// Spec 7.1: "changes the focused item and redraws" — navigation must not be broken by removing OnSelect.
func TestNavDownStillMovesSelection(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})

	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)

	if lv.Selected() != 1 {
		t.Errorf("spec 7.1: Down arrow must still move selection; got Selected()=%d, want 1", lv.Selected())
	}
}

// TestNavUpStillMovesSelection verifies Up arrow still moves selected from 2 to 1.
// Spec 7.1: "changes the focused item and redraws" — navigation must not be broken by removing OnSelect.
func TestNavUpStillMovesSelection(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(2)

	ev := listKeyEv(tcell.KeyUp)
	lv.HandleEvent(ev)

	if lv.Selected() != 1 {
		t.Errorf("spec 7.1: Up arrow must still move selection; got Selected()=%d, want 1", lv.Selected())
	}
}

// TestNavMultipleDownArrowsAccumulate verifies multiple Down presses accumulate: selected goes 0→1→2→3.
// Spec 7.1: navigation is not broken — absence of OnSelect must not prevent normal cursor movement.
func TestNavMultipleDownArrowsAccumulate(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e"})

	for i := 0; i < 3; i++ {
		ev := listKeyEv(tcell.KeyDown)
		lv.HandleEvent(ev)
	}

	if lv.Selected() != 3 {
		t.Errorf("spec 7.1: three Down presses from 0 must reach Selected()=3; got %d", lv.Selected())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Events still consumed by navigation keys
// ---------------------------------------------------------------------------

// TestNavDownEventIsConsumed verifies Down arrow still consumes (clears) the event.
// Spec 7.1: "All navigation keys still consume the event (Clear)."
func TestNavDownEventIsConsumed(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.1: Down arrow must still consume (clear) the event")
	}
}

// TestNavUpEventIsConsumed verifies Up arrow still consumes (clears) the event.
// Spec 7.1: "All navigation keys still consume the event (Clear)."
func TestNavUpEventIsConsumed(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(1)
	ev := listKeyEv(tcell.KeyUp)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.1: Up arrow must still consume (clear) the event")
	}
}

// TestNavHomeEventIsConsumed verifies Home still consumes (clears) the event.
// Spec 7.1: "All navigation keys still consume the event (Clear)."
func TestNavHomeEventIsConsumed(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.1: Home must still consume (clear) the event")
	}
}

// TestNavEndEventIsConsumed verifies End still consumes (clears) the event.
// Spec 7.1: "All navigation keys still consume the event (Clear)."
func TestNavEndEventIsConsumed(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.1: End must still consume (clear) the event")
	}
}

// TestNavPgDnEventIsConsumed verifies PgDn still consumes (clears) the event.
// Spec 7.1: "All navigation keys still consume the event (Clear)."
func TestNavPgDnEventIsConsumed(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	ev := listKeyEv(tcell.KeyPgDn)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.1: PgDn must still consume (clear) the event")
	}
}

// TestNavPgUpEventIsConsumed verifies PgUp still consumes (clears) the event.
// Spec 7.1: "All navigation keys still consume the event (Clear)."
func TestNavPgUpEventIsConsumed(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	lv.SetSelected(6)
	ev := listKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.1: PgUp must still consume (clear) the event")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Nil OnSelect safety
// ---------------------------------------------------------------------------

// TestNavDownWithNilOnSelectDoesNotPanic verifies Down arrow with nil OnSelect does not panic.
// Spec 7.1: "OnSelect is nil-safe (no change from current behavior — just removing the calls
// from navigation)." Nil OnSelect must never be dereferenced.
func TestNavDownWithNilOnSelectDoesNotPanic(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	// OnSelect is nil by default — do not set it

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("spec 7.1: Down arrow with nil OnSelect panicked: %v", r)
		}
	}()

	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)
}
