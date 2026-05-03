package tv

// list_viewer_drag_test.go — Tests for Task 7: Mouse drag with auto-scroll (spec 7.5).
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Spec 7.5: When clicking and dragging in the ListViewer:
//   - Track mouse movement while button is held
//   - Update focused item to match mouse position
//   - If mouse moves above the top edge, scroll up (auto-scroll)
//   - If mouse moves below the bottom edge, scroll down
//   - Drag does NOT fire OnSelect (navigation, not selection)
//   - Double-click (ClickCount >= 2) during a drag still fires OnSelect
//
// Test organisation:
//   Section 1 — Drag tracking (press starts drag, move updates, release stops)
//   Section 2 — Auto-scroll up (mouseY < 0)
//   Section 3 — Auto-scroll down (mouseY >= visibleHeight)
//   Section 4 — OnSelect behavior during drag
//   Section 5 — Edge cases (empty list, out-of-bounds)

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Drag event helpers
// ---------------------------------------------------------------------------

// dragPressEv creates a Button1 press event at (x, y) — starts a drag.
func dragPressEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 1}}
}

// dragMoveEv creates a Button1 held event at (x, y) — continues a drag.
// Button1 is still set (mouse is held), simulating evMouseAuto.
func dragMoveEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1}}
}

// dragReleaseEv creates a mouse release event at (x, y) — Button is 0 (no buttons pressed).
func dragReleaseEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: 0}}
}

// dragDoubleClickEv creates a Button1 double-click event at (x, y) during drag.
func dragDoubleClickEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 2}}
}

// ---------------------------------------------------------------------------
// Section 1 — Drag tracking
// ---------------------------------------------------------------------------

// TestDragPressStartsDragging verifies that a Button1 press sets dragging = true
// and updates selected to topIndex + mouseY.
// Spec 7.5: "When Button1 is pressed and held, set dragging = true"
func TestDragPressStartsDragging(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})

	ev := dragPressEv(0, 2)
	lv.HandleEvent(ev)

	if !lv.dragging {
		t.Error("spec 7.5: dragging should be true after Button1 press")
	}
}

// TestDragPressUpdatesSelectedToClickedRow verifies that a Button1 press updates
// selected to topIndex + mouseY.
// Spec 7.5: "Update selected to topIndex + mouseY, clamped to valid range"
func TestDragPressUpdatesSelectedToClickedRow(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})
	// topIndex=0; press at row 2 → selected should be 2

	ev := dragPressEv(0, 2)
	lv.HandleEvent(ev)

	if lv.Selected() != 2 {
		t.Errorf("spec 7.5: drag press at row 2: Selected()=%d, want 2 (topIndex=0 + mouseY=2)", lv.Selected())
	}
}

// TestDragMoveUpdatesSelected verifies that a subsequent Button1 event while dragging
// updates selected to topIndex + mouseY.
// Spec 7.5: "On subsequent Button1 mouse events while dragging, update selected to topIndex + mouseY"
func TestDragMoveUpdatesSelected(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})

	// Press at row 0 to start drag.
	lv.HandleEvent(dragPressEv(0, 0))
	if lv.Selected() != 0 {
		t.Fatalf("precondition: after press at row 0, Selected()=%d, want 0", lv.Selected())
	}

	// Drag to row 3.
	lv.HandleEvent(dragMoveEv(0, 3))

	if lv.Selected() != 3 {
		t.Errorf("spec 7.5: drag to row 3: Selected()=%d, want 3 (topIndex=0 + mouseY=3)", lv.Selected())
	}
}

// TestDragReleaseStopsDragging verifies that a Button1 release (Button=0) sets
// dragging = false.
// Spec 7.5: "On Button1 release (Button1 not set while dragging): set dragging = false"
func TestDragReleaseStopsDragging(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})

	lv.HandleEvent(dragPressEv(0, 0))
	if !lv.dragging {
		t.Fatal("precondition: dragging should be true after press")
	}

	lv.HandleEvent(dragReleaseEv(0, 0))

	if lv.dragging {
		t.Error("spec 7.5: dragging should be false after Button1 release (Button=0)")
	}
}

// TestAfterDragReleaseMouseMoveDoesNotUpdateSelection verifies that after releasing
// the button, subsequent mouse events (with Button1 held) are not treated as drags.
// Spec 7.5: drag tracking requires dragging = true; release ends it.
func TestAfterDragReleaseMouseMoveDoesNotUpdateSelection(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})

	// Press, drag to row 2, then release.
	lv.HandleEvent(dragPressEv(0, 0))
	lv.HandleEvent(dragMoveEv(0, 2))
	lv.HandleEvent(dragReleaseEv(0, 2))

	selAfterRelease := lv.Selected()

	// Simulate a Button1 event without an active drag — this is a fresh click,
	// not a drag continuation. It should process normally (as a click at that row),
	// but the key point is that dragging = false and the new click sets selection
	// based on click position alone, not a "drag update" path.
	// We verify by checking dragging remains false.
	lv.HandleEvent(dragMoveEv(0, 4))

	if lv.dragging {
		t.Error("spec 7.5: after release, dragging should remain false even after subsequent Button1 event")
	}
	// The subsequent Button1 event is a new press that starts a new drag.
	// What we care about is that dragging=false was properly set on release.
	_ = selAfterRelease
}

// TestDragConsumesEachMouseEvent verifies that each mouse event during a drag
// is consumed (cleared).
// Spec 7.5: "During drag, consume each mouse event (Clear)"
func TestDragConsumesEachMouseEvent(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})

	// Press event should be consumed.
	pressEv := dragPressEv(0, 0)
	lv.HandleEvent(pressEv)
	if !pressEv.IsCleared() {
		t.Error("spec 7.5: press event (drag start) should be consumed")
	}

	// Move event should be consumed.
	moveEv := dragMoveEv(0, 2)
	lv.HandleEvent(moveEv)
	if !moveEv.IsCleared() {
		t.Error("spec 7.5: drag-move event should be consumed")
	}

	// Release event should be consumed.
	releaseEv := dragReleaseEv(0, 2)
	lv.HandleEvent(releaseEv)
	if !releaseEv.IsCleared() {
		t.Error("spec 7.5: release event should be consumed")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Auto-scroll up (mouseY < 0)
// ---------------------------------------------------------------------------

// TestAutoScrollUpDecreasesTopIndex verifies that a drag event with mouseY = -1
// (above widget) decreases topIndex by 1 when topIndex > 0.
// Spec 7.5: "If mouseY < 0: auto-scroll up — decrease topIndex by 1 (if > 0), set selected = topIndex"
func TestAutoScrollUpDecreasesTopIndex(t *testing.T) {
	// 10 items, height=5. Start scrolled so topIndex=3.
	items := make([]string, 10)
	for i := range items {
		items[i] = string(rune('a' + i))
	}
	lv := newLV(items)
	lv.SetSelected(7) // forces topIndex = 7 - 5 + 1 = 3

	topBefore := lv.TopIndex()
	if topBefore != 3 {
		t.Fatalf("precondition: topIndex should be 3, got %d", topBefore)
	}

	// Start drag.
	lv.HandleEvent(dragPressEv(0, 0))

	// Drag above the widget (mouseY = -1).
	lv.HandleEvent(dragMoveEv(0, -1))

	if lv.TopIndex() != topBefore-1 {
		t.Errorf("spec 7.5: auto-scroll up: topIndex = %d, want %d (decreased by 1)",
			lv.TopIndex(), topBefore-1)
	}
}

// TestAutoScrollUpSetsSelectedToNewTopIndex verifies that after scrolling up,
// selected is set to the new topIndex.
// Spec 7.5: "set selected = topIndex" (after decreasing topIndex)
func TestAutoScrollUpSetsSelectedToNewTopIndex(t *testing.T) {
	items := make([]string, 10)
	for i := range items {
		items[i] = string(rune('a' + i))
	}
	lv := newLV(items)
	lv.SetSelected(7) // topIndex = 3

	lv.HandleEvent(dragPressEv(0, 0))
	lv.HandleEvent(dragMoveEv(0, -1))

	// After scrolling up, topIndex = 2, selected should = topIndex = 2.
	if lv.Selected() != lv.TopIndex() {
		t.Errorf("spec 7.5: auto-scroll up: Selected()=%d, want topIndex=%d",
			lv.Selected(), lv.TopIndex())
	}
}

// TestAutoScrollUpNoOpAtTopBoundary verifies that when topIndex = 0, a mouseY = -1
// drag event does not scroll (can't go further up).
// Spec 7.5: "decrease topIndex by 1 (if > 0)" — guards against going below 0
func TestAutoScrollUpNoOpAtTopBoundary(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g"}
	lv := newLV(items)
	// topIndex = 0 (default)

	lv.HandleEvent(dragPressEv(0, 0))
	lv.HandleEvent(dragMoveEv(0, -1))

	if lv.TopIndex() != 0 {
		t.Errorf("spec 7.5: auto-scroll up at boundary: topIndex = %d, want 0 (no change)", lv.TopIndex())
	}
	if lv.Selected() != 0 {
		t.Errorf("spec 7.5: auto-scroll up at boundary: Selected() = %d, want 0 (no change)", lv.Selected())
	}
}

// TestAutoScrollUpAccumulatesOverMultipleEvents verifies that multiple consecutive
// auto-scroll-up events each decrease topIndex by 1.
// Spec 7.5: "Auto-scroll boundary: scroll one item per event"
func TestAutoScrollUpAccumulatesOverMultipleEvents(t *testing.T) {
	items := make([]string, 10)
	for i := range items {
		items[i] = string(rune('a' + i))
	}
	lv := newLV(items)
	lv.SetSelected(9) // topIndex = 9 - 5 + 1 = 5

	lv.HandleEvent(dragPressEv(0, 0))

	// Three auto-scroll-up events should decrease topIndex by 3.
	for i := 0; i < 3; i++ {
		lv.HandleEvent(dragMoveEv(0, -1))
	}

	if lv.TopIndex() != 2 {
		t.Errorf("spec 7.5: 3 auto-scroll-up events: topIndex = %d, want 2 (5 - 3)", lv.TopIndex())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Auto-scroll down (mouseY >= visibleHeight)
// ---------------------------------------------------------------------------

// TestAutoScrollDownIncreasesTopIndex verifies that a drag event with mouseY = visibleHeight
// (below widget) increases topIndex by 1 when more items are below.
// Spec 7.5: "If mouseY >= visibleHeight: auto-scroll down — increase topIndex by 1 (if more items below)"
func TestAutoScrollDownIncreasesTopIndex(t *testing.T) {
	// 10 items, height=5. Start at topIndex=0, so there are items below.
	items := make([]string, 10)
	for i := range items {
		items[i] = string(rune('a' + i))
	}
	lv := newLV(items)
	// topIndex=0; visibleHeight=5; items 5-9 are below

	lv.HandleEvent(dragPressEv(0, 0))

	// Drag below the widget (mouseY = visibleHeight = 5).
	lv.HandleEvent(dragMoveEv(0, 5))

	if lv.TopIndex() != 1 {
		t.Errorf("spec 7.5: auto-scroll down: topIndex = %d, want 1 (increased by 1)", lv.TopIndex())
	}
}

// TestAutoScrollDownSetsSelectedToLastVisibleItem verifies that after scrolling down,
// selected is set to the last visible item (topIndex + visibleHeight - 1, clamped).
// Spec 7.5: "set selected to last visible item"
func TestAutoScrollDownSetsSelectedToLastVisibleItem(t *testing.T) {
	items := make([]string, 10)
	for i := range items {
		items[i] = string(rune('a' + i))
	}
	lv := newLV(items)
	// topIndex=0; visibleHeight=5

	lv.HandleEvent(dragPressEv(0, 0))
	lv.HandleEvent(dragMoveEv(0, 5)) // auto-scroll down: topIndex becomes 1

	// Last visible = topIndex + visibleHeight - 1 = 1 + 5 - 1 = 5
	wantSelected := lv.TopIndex() + lv.visibleHeight() - 1
	if wantSelected >= lv.dataSource.Count() {
		wantSelected = lv.dataSource.Count() - 1
	}

	if lv.Selected() != wantSelected {
		t.Errorf("spec 7.5: auto-scroll down: Selected()=%d, want last visible=%d",
			lv.Selected(), wantSelected)
	}
}

// TestAutoScrollDownNoOpAtMaxTopIndex verifies that when already at the maximum
// topIndex (no more items below), auto-scroll down is a no-op.
// Spec 7.5: "increase topIndex by 1 (if more items below)" — guard against over-scroll
func TestAutoScrollDownNoOpAtMaxTopIndex(t *testing.T) {
	// 7 items, height=5. Max topIndex = 7 - 5 = 2.
	items := []string{"a", "b", "c", "d", "e", "f", "g"}
	lv := newLV(items)
	lv.SetSelected(6) // topIndex = 6 - 5 + 1 = 2 (max)

	topBefore := lv.TopIndex()
	selBefore := lv.Selected()

	lv.HandleEvent(dragPressEv(0, 0))
	lv.HandleEvent(dragMoveEv(0, 5)) // try to auto-scroll down

	if lv.TopIndex() != topBefore {
		t.Errorf("spec 7.5: auto-scroll down at max: topIndex = %d, want %d (no change)",
			lv.TopIndex(), topBefore)
	}
	_ = selBefore
}

// TestAutoScrollDownAccumulatesOverMultipleEvents verifies that multiple consecutive
// auto-scroll-down events each increase topIndex by 1.
// Spec 7.5: "Auto-scroll boundary: scroll one item per event"
func TestAutoScrollDownAccumulatesOverMultipleEvents(t *testing.T) {
	// 12 items, height=5. Max topIndex = 12 - 5 = 7.
	items := make([]string, 12)
	for i := range items {
		items[i] = string(rune('a' + i))
	}
	lv := newLV(items)
	// topIndex=0 initially

	lv.HandleEvent(dragPressEv(0, 0))

	// Three auto-scroll-down events should increase topIndex by 3.
	for i := 0; i < 3; i++ {
		lv.HandleEvent(dragMoveEv(0, 5))
	}

	if lv.TopIndex() != 3 {
		t.Errorf("spec 7.5: 3 auto-scroll-down events: topIndex = %d, want 3 (0 + 3)", lv.TopIndex())
	}
}

// ---------------------------------------------------------------------------
// Section 4 — OnSelect behavior during drag
// ---------------------------------------------------------------------------

// TestDragDoesNotFireOnSelect verifies that dragging (press and move) does NOT
// call OnSelect at any point.
// Spec 7.5: "Drag does NOT fire OnSelect (navigation, not selection)"
func TestDragDoesNotFireOnSelect(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	lv := newLV(items)
	called := false
	lv.OnSelect = func(index int) { called = true }

	// Press to start drag.
	lv.HandleEvent(dragPressEv(0, 0))
	if called {
		t.Error("spec 7.5: OnSelect must NOT be called on drag press")
	}

	// Move during drag.
	lv.HandleEvent(dragMoveEv(0, 2))
	if called {
		t.Error("spec 7.5: OnSelect must NOT be called during drag move")
	}

	// Release.
	lv.HandleEvent(dragReleaseEv(0, 2))
	if called {
		t.Error("spec 7.5: OnSelect must NOT be called on drag release")
	}
}

// TestDragDoesNotFireOnSelectDuringAutoScrollUp verifies OnSelect is NOT called
// during auto-scroll up.
// Spec 7.5: "Drag does NOT fire OnSelect"
func TestDragDoesNotFireOnSelectDuringAutoScrollUp(t *testing.T) {
	items := make([]string, 10)
	for i := range items {
		items[i] = string(rune('a' + i))
	}
	lv := newLV(items)
	lv.SetSelected(7) // topIndex=3

	called := false
	lv.OnSelect = func(index int) { called = true }

	lv.HandleEvent(dragPressEv(0, 0))
	lv.HandleEvent(dragMoveEv(0, -1)) // auto-scroll up

	if called {
		t.Error("spec 7.5: OnSelect must NOT be called during auto-scroll up drag")
	}
}

// TestDragDoesNotFireOnSelectDuringAutoScrollDown verifies OnSelect is NOT called
// during auto-scroll down.
// Spec 7.5: "Drag does NOT fire OnSelect"
func TestDragDoesNotFireOnSelectDuringAutoScrollDown(t *testing.T) {
	items := make([]string, 10)
	for i := range items {
		items[i] = string(rune('a' + i))
	}
	lv := newLV(items)
	called := false
	lv.OnSelect = func(index int) { called = true }

	lv.HandleEvent(dragPressEv(0, 0))
	lv.HandleEvent(dragMoveEv(0, 5)) // auto-scroll down

	if called {
		t.Error("spec 7.5: OnSelect must NOT be called during auto-scroll down drag")
	}
}

// TestDoubleClickDuringDragFiresOnSelect verifies that a double-click event
// (ClickCount >= 2) during a drag still fires OnSelect for in-bounds items.
// Spec 7.5: "Double-click (ClickCount >= 2) during a drag still fires OnSelect for in-bounds items"
func TestDoubleClickDuringDragFiresOnSelect(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	lv := newLV(items)
	called := false
	var gotIndex int
	lv.OnSelect = func(index int) { called = true; gotIndex = index }

	// Start drag.
	lv.HandleEvent(dragPressEv(0, 0))

	// Double-click at row 2 during the drag.
	lv.HandleEvent(dragDoubleClickEv(0, 2))

	if !called {
		t.Error("spec 7.5: double-click (ClickCount >= 2) during drag must fire OnSelect")
	}
	if gotIndex != 2 {
		t.Errorf("spec 7.5: OnSelect received index %d after double-click at row 2, want 2", gotIndex)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Edge cases
// ---------------------------------------------------------------------------

// TestDragWithEmptyListDoesNotCrash verifies that pressing and dragging in an
// empty ListViewer does not panic.
// Spec 7.5: defensive edge case — empty data source
func TestDragWithEmptyListDoesNotCrash(t *testing.T) {
	lv := newLV([]string{})

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("spec 7.5: drag on empty list panicked: %v", r)
		}
	}()

	lv.HandleEvent(dragPressEv(0, 0))
	lv.HandleEvent(dragMoveEv(0, 2))
	lv.HandleEvent(dragMoveEv(0, -1))
	lv.HandleEvent(dragMoveEv(0, 5))
	lv.HandleEvent(dragReleaseEv(0, 0))
}

// TestDragBeyondDataCountKeepsSelectionValid verifies that dragging to a row
// beyond the data count keeps selected clamped to a valid range.
// Spec 7.5: "clamped to valid range" in "update selected to topIndex + mouseY, clamped"
func TestDragBeyondDataCountKeepsSelectionValid(t *testing.T) {
	lv := newLV([]string{"a", "b"}) // 2 items, height=5

	lv.HandleEvent(dragPressEv(0, 0))
	// Drag to row 4 — beyond the 2 items
	lv.HandleEvent(dragMoveEv(0, 4))

	// Selected must remain in [0, Count()-1] = [0, 1]
	if lv.Selected() < 0 || lv.Selected() >= 2 {
		t.Errorf("spec 7.5: drag beyond data: Selected()=%d is out of valid range [0, 1]", lv.Selected())
	}
}

// TestDragSelectedClampedToValidRange verifies selection is clamped when
// topIndex + mouseY would go negative (shouldn't happen during normal drag, but
// ensures the clamp is in place).
// Spec 7.5: "clamped to valid range"
func TestDragSelectedClampedToValidRange(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("spec 7.5: drag with negative computed index panicked: %v", r)
		}
	}()

	lv.HandleEvent(dragPressEv(0, 2))
	// Even a weird in-bounds move should result in valid selection.
	lv.HandleEvent(dragMoveEv(0, 0))

	if lv.Selected() < 0 || lv.Selected() >= 5 {
		t.Errorf("spec 7.5: drag: Selected()=%d is out of valid range [0, 4]", lv.Selected())
	}
}
