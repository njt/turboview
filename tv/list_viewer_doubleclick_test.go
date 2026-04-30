package tv

// list_viewer_doubleclick_test.go — Tests for Task 3: Single-click focus only,
// double-click to select (spec 7.3).
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Spec 7.3: Mouse single-click should position focus (set `selected` to the
// clicked item) but NOT call OnSelect. Double-click should position focus AND
// call OnSelect. Double-click detection uses MouseEvent.ClickCount >= 2.
//
// Test organisation:
//   Section 1  — Single-click does NOT call OnSelect (confirming)
//   Section 2  — Single-click still positions focus and consumes event (confirming)
//   Section 3  — Double-click calls OnSelect (confirming)
//   Section 4  — Double-click still consumes event and handles edge cases
//   Section 5  — Falsification tests (prove the single/double distinction exists)

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// listSingleClickEv creates a mouse Button1 event with ClickCount=1 (explicit single-click).
func listSingleClickEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 1}}
}

// listDoubleClickEv creates a mouse Button1 event with ClickCount=2 (double-click).
func listDoubleClickEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 2}}
}

// listTripleClickEv creates a mouse Button1 event with ClickCount=3 (triple-click).
func listTripleClickEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 3}}
}

// ---------------------------------------------------------------------------
// Section 1 — Single-click does NOT call OnSelect
// ---------------------------------------------------------------------------

// TestSingleClickDoesNotCallOnSelect verifies a single click (ClickCount=1) does NOT
// call OnSelect.
// Spec 7.3: "Mouse single-click should position focus … but NOT call OnSelect."
func TestSingleClickDoesNotCallOnSelect(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listSingleClickEv(0, 1)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.3: OnSelect must NOT be called on single click (ClickCount=1)")
	}
}

// TestSingleClickZeroCountDoesNotCallOnSelect verifies a click with ClickCount=0 (Go zero
// value, as produced by listMouseEv) does NOT call OnSelect.
// Spec 7.3: single-click means ClickCount < 2; ClickCount=0 is also a single click.
func TestSingleClickZeroCountDoesNotCallOnSelect(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	// listMouseEv produces ClickCount=0 (Go zero value)
	ev := listMouseEv(0, 1)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.3: OnSelect must NOT be called when ClickCount=0 (single-click zero value)")
	}
}

// TestSingleClickBeyondDataCountDoesNotCallOnSelect verifies a single click on a row beyond
// the data count does not call OnSelect and does not panic.
// Spec 7.3: click beyond data count — no change, no OnSelect, no panic.
func TestSingleClickBeyondDataCountDoesNotCallOnSelect(t *testing.T) {
	lv := newLV([]string{"a", "b"}) // 2 items, height=5
	called := false
	lv.OnSelect = func(index int) { called = true }

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("spec 7.3: single click beyond data panicked: %v", r)
		}
	}()

	ev := listSingleClickEv(0, 4) // row 4 — beyond the 2 items
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.3: OnSelect must NOT be called when clicking beyond data count")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Single-click still positions focus and consumes event
// ---------------------------------------------------------------------------

// TestSingleClickSetsSelectedToClickedRow verifies a single click sets selected to
// topIndex + clickY.
// Spec 7.3: "single-click … sets selected to the clicked item."
func TestSingleClickSetsSelectedToClickedRow(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})
	// topIndex=0; click row 2 → selected should be 2
	ev := listSingleClickEv(0, 2)
	lv.HandleEvent(ev)

	if lv.Selected() != 2 {
		t.Errorf("spec 7.3: single click row 2: Selected()=%d, want 2 (topIndex=0 + clickY=2)", lv.Selected())
	}
}

// TestSingleClickConsumesEvent verifies a single click (ClickCount=1) consumes the event.
// Spec 7.3: single click "consumes event."
func TestSingleClickConsumesEvent(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	ev := listSingleClickEv(0, 0)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.3: single click must consume (clear) the event")
	}
}

// TestSingleClickZeroCountConsumesEvent verifies a click with ClickCount=0 still consumes
// the event.
// Spec 7.3: event is consumed regardless of click count.
func TestSingleClickZeroCountConsumesEvent(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	ev := listMouseEv(0, 0) // ClickCount=0
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.3: click with ClickCount=0 must still consume (clear) the event")
	}
}

// TestSingleClickBeyondDataCountConsumesEvent verifies a single click beyond the data count
// still consumes the event.
// Spec 7.3: "single-click … consumes event" — even when no selection change occurs.
func TestSingleClickBeyondDataCountConsumesEvent(t *testing.T) {
	lv := newLV([]string{"a", "b"}) // 2 items, height=5
	ev := listSingleClickEv(0, 4)   // beyond data

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("spec 7.3: single click beyond data panicked: %v", r)
		}
	}()

	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.3: single click beyond data must still consume (clear) the event")
	}
}

// TestSingleClickDoesNotChangeSelectionBeyondData verifies selection remains unchanged
// when clicking beyond the data count.
// Spec 7.3: "click beyond data count: no change."
func TestSingleClickDoesNotChangeSelectionBeyondData(t *testing.T) {
	lv := newLV([]string{"a", "b"}) // 2 items, height=5
	lv.SetSelected(1)               // start at item 1

	ev := listSingleClickEv(0, 4) // row 4 — beyond the 2 items
	lv.HandleEvent(ev)

	if lv.Selected() != 1 {
		t.Errorf("spec 7.3: click beyond data: Selected()=%d, want 1 (unchanged)", lv.Selected())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Double-click calls OnSelect
// ---------------------------------------------------------------------------

// TestDoubleClickCallsOnSelect verifies a double click (ClickCount=2) calls OnSelect.
// Spec 7.3: "Double-click should position focus AND call OnSelect."
func TestDoubleClickCallsOnSelect(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listDoubleClickEv(0, 1)
	lv.HandleEvent(ev)

	if !called {
		t.Error("spec 7.3: OnSelect must be called on double click (ClickCount=2)")
	}
}

// TestDoubleClickPassesCorrectIndexToOnSelect verifies the double click passes
// topIndex + clickY as the argument to OnSelect.
// Spec 7.3: "Double-click should position focus AND call OnSelect."
func TestDoubleClickPassesCorrectIndexToOnSelect(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})
	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	// topIndex=0; click row 2 → OnSelect should receive 2
	ev := listDoubleClickEv(0, 2)
	lv.HandleEvent(ev)

	if got != 2 {
		t.Errorf("spec 7.3: OnSelect received index %d after double-clicking row 2, want 2", got)
	}
}

// TestDoubleClickPassesCorrectIndexWithScrollOffset verifies double-click index
// accounts for topIndex.
// Spec 7.3: "The clicked row is topIndex + clickY."
func TestDoubleClickPassesCorrectIndexWithScrollOffset(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	lv := newLV(items)
	lv.SetSelected(7) // forces topIndex = 7 - 5 + 1 = 3

	topIdx := lv.TopIndex()
	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	ev := listDoubleClickEv(0, 1) // row 1 → topIndex + 1
	lv.HandleEvent(ev)

	want := topIdx + 1
	if got != want {
		t.Errorf("spec 7.3: double-click row 1 with topIndex=%d: OnSelect received %d, want %d",
			topIdx, got, want)
	}
}

// TestDoubleClickAlsoSetsSelected verifies double-click positions focus (sets selected).
// Spec 7.3: "Double-click should position focus AND call OnSelect."
func TestDoubleClickAlsoSetsSelected(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})
	lv.OnSelect = func(index int) {} // set but ignore

	ev := listDoubleClickEv(0, 3)
	lv.HandleEvent(ev)

	if lv.Selected() != 3 {
		t.Errorf("spec 7.3: double-click row 3: Selected()=%d, want 3", lv.Selected())
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Double-click consumes event and handles edge cases
// ---------------------------------------------------------------------------

// TestDoubleClickConsumesEvent verifies a double click consumes the event.
// Spec 7.3: double-click "consumes event."
func TestDoubleClickConsumesEvent(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	lv.OnSelect = func(index int) {}

	ev := listDoubleClickEv(0, 1)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.3: double click must consume (clear) the event")
	}
}

// TestDoubleClickBeyondDataCountDoesNotCallOnSelect verifies a double click on a row
// beyond the data count does not call OnSelect.
// Spec 7.3: "double-click beyond data count: no OnSelect, no panic."
func TestDoubleClickBeyondDataCountDoesNotCallOnSelect(t *testing.T) {
	lv := newLV([]string{"a", "b"}) // 2 items, height=5
	called := false
	lv.OnSelect = func(index int) { called = true }

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("spec 7.3: double click beyond data panicked: %v", r)
		}
	}()

	ev := listDoubleClickEv(0, 4) // row 4 — beyond the 2 items
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.3: OnSelect must NOT be called when double-clicking beyond data count")
	}
}

// TestDoubleClickBeyondDataCountConsumesEvent verifies a double click beyond the data
// count still consumes the event.
// Spec 7.3: double-click "consumes event" — even when no selection change occurs.
func TestDoubleClickBeyondDataCountConsumesEvent(t *testing.T) {
	lv := newLV([]string{"a", "b"}) // 2 items, height=5

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("spec 7.3: double click beyond data panicked: %v", r)
		}
	}()

	ev := listDoubleClickEv(0, 4)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.3: double click beyond data must still consume (clear) the event")
	}
}

// TestDoubleClickWithNilOnSelectDoesNotPanic verifies double-click with nil OnSelect
// does not panic and still consumes the event.
// Spec 7.3: nil OnSelect is safe — no panic.
func TestDoubleClickWithNilOnSelectDoesNotPanic(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	// OnSelect intentionally left nil

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("spec 7.3: double click with nil OnSelect panicked: %v", r)
		}
	}()

	ev := listDoubleClickEv(0, 1)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.3: double click with nil OnSelect must still consume (clear) the event")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Falsification tests (prove the single/double distinction exists)
// ---------------------------------------------------------------------------

// TestSingleClickVsDoubleClickDistinction verifies that single-click does NOT fire
// OnSelect but double-click on the same row DOES — proving the distinction is real.
// Spec 7.3: This is the core distinction: single-click = focus-only, double-click = confirm.
func TestSingleClickVsDoubleClickDistinction(t *testing.T) {
	items := []string{"a", "b", "c"}

	// Single click on row 1
	lvSingle := newLV(items)
	singleCalled := false
	lvSingle.OnSelect = func(index int) { singleCalled = true }
	lvSingle.HandleEvent(listSingleClickEv(0, 1))

	// Double click on row 1
	lvDouble := newLV(items)
	doubleCalled := false
	lvDouble.OnSelect = func(index int) { doubleCalled = true }
	lvDouble.HandleEvent(listDoubleClickEv(0, 1))

	if singleCalled {
		t.Error("spec 7.3: single-click must NOT fire OnSelect")
	}
	if !doubleCalled {
		t.Error("spec 7.3: double-click must fire OnSelect")
	}
}

// TestTripleClickAlsoFiresOnSelect verifies triple-click (ClickCount=3) also calls OnSelect
// because the spec uses ClickCount >= 2, not == 2.
// Spec 7.3: "ClickCount >= 2 means double-click" — triple-click satisfies >= 2.
func TestTripleClickAlsoFiresOnSelect(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listTripleClickEv(0, 1)
	lv.HandleEvent(ev)

	if !called {
		t.Error("spec 7.3: triple click (ClickCount=3) must call OnSelect because ClickCount >= 2")
	}
}

// TestSingleClickWithClickCountOneDoesNotFireOnSelectButZeroDoesNot verifies
// that both ClickCount=0 and ClickCount=1 are single-clicks that do not fire OnSelect,
// ruling out any off-by-one where only ClickCount=0 is treated as single-click.
// Spec 7.3: ClickCount < 2 means single-click.
func TestBothClickCount0And1AreSingleClicks(t *testing.T) {
	items := []string{"a", "b", "c"}

	for _, count := range []int{0, 1} {
		lv := newLV(items)
		called := false
		lv.OnSelect = func(index int) { called = true }

		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1, ClickCount: count}}
		lv.HandleEvent(ev)

		if called {
			t.Errorf("spec 7.3: click with ClickCount=%d must NOT fire OnSelect (it is a single-click)", count)
		}
	}
}
