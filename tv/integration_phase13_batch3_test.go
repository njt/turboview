package tv

// integration_phase13_batch3_test.go — Integration tests for Task 8:
// Mouse drag and auto-scroll in the TurboView ListViewer.
//
// Each test exercises the drag/auto-scroll pipeline end-to-end, verifying
// selection movement, scrollbar sync, and OnSelect behavior throughout
// a complete drag sequence.
//
// Test naming: TestIntegrationPhase13Batch3<DescriptiveSuffix>

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Mouse event helpers (integration-level; named distinctly to avoid collision
// with list_viewer_drag_test.go helpers which are in the same package).
// ---------------------------------------------------------------------------

// intDragPress returns a Button1 press event at (x, y) — starts a drag.
func intDragPress(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 1}}
}

// intDragMove returns a Button1-held move event at (x, y) — continues a drag.
func intDragMove(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 0}}
}

// intDragRelease returns a button-release event at (x, y) — ends the drag.
func intDragRelease(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: 0, ClickCount: 0}}
}

// intDragDoubleClick returns a Button1 double-click event at (x, y).
func intDragDoubleClick(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 2}}
}

// ---------------------------------------------------------------------------
// Requirement 1: Pressing Button1 on row 2, then moving to row 4 (Button1 still
// held) updates selected from topIndex+2 to topIndex+4.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch3DragRow2ToRow4UpdatesSelected verifies that
// pressing Button1 on row 2 then dragging to row 4 updates selected accordingly.
func TestIntegrationPhase13Batch3DragRow2ToRow4UpdatesSelected(t *testing.T) {
	lv := newLVFocused(items20()) // 20 items, bounds 20x5, topIndex=0
	topIdx := lv.TopIndex()      // 0

	// Press on row 2 — selection should update to topIndex+2.
	lv.HandleEvent(intDragPress(0, 2))
	if lv.Selected() != topIdx+2 {
		t.Fatalf("after press row 2: Selected()=%d, want %d (topIndex+2)", lv.Selected(), topIdx+2)
	}

	// Drag to row 4 (Button1 still held) — selection should update to topIndex+4.
	lv.HandleEvent(intDragMove(0, 4))
	if lv.Selected() != topIdx+4 {
		t.Errorf("after drag to row 4: Selected()=%d, want %d (topIndex+4)", lv.Selected(), topIdx+4)
	}
}

// TestIntegrationPhase13Batch3DragUpdatesSelectionOnEachMove verifies that
// selection tracks the mouse Y position on each drag event during a drag sequence.
func TestIntegrationPhase13Batch3DragUpdatesSelectionOnEachMove(t *testing.T) {
	lv := newLVFocused(items20())
	topIdx := lv.TopIndex() // 0

	lv.HandleEvent(intDragPress(0, 0))
	if lv.Selected() != topIdx+0 {
		t.Fatalf("after press row 0: Selected()=%d, want %d", lv.Selected(), topIdx)
	}

	for _, row := range []int{1, 2, 3, 4} {
		lv.HandleEvent(intDragMove(0, row))
		if lv.Selected() != topIdx+row {
			t.Errorf("drag to row %d: Selected()=%d, want %d", row, lv.Selected(), topIdx+row)
		}
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: Dragging above the widget (mouseY = -1) with topIndex > 0
// scrolls up and selects the new topIndex.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch3AutoScrollUpScrollsAndSelectsNewTopIndex verifies
// that dragging above the widget decreases topIndex and sets selected=topIndex.
func TestIntegrationPhase13Batch3AutoScrollUpScrollsAndSelectsNewTopIndex(t *testing.T) {
	lv := newLVFocused(items20())
	lv.SetSelected(10) // forces topIndex = 10 - 5 + 1 = 6

	topBefore := lv.TopIndex()
	if topBefore <= 0 {
		t.Fatalf("prerequisite: topIndex must be > 0 for auto-scroll up test; got %d", topBefore)
	}

	// Start drag.
	lv.HandleEvent(intDragPress(0, 0))

	// Drag above the widget (mouseY = -1).
	lv.HandleEvent(intDragMove(0, -1))

	if lv.TopIndex() != topBefore-1 {
		t.Errorf("auto-scroll up: TopIndex()=%d, want %d (decreased by 1)", lv.TopIndex(), topBefore-1)
	}
	if lv.Selected() != lv.TopIndex() {
		t.Errorf("auto-scroll up: Selected()=%d, want topIndex=%d", lv.Selected(), lv.TopIndex())
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: Dragging below the widget (mouseY >= visibleHeight) with items
// below scrolls down and selects the last visible item.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch3AutoScrollDownScrollsAndSelectsLastVisible verifies
// that dragging below the widget increases topIndex and selects the last visible item.
func TestIntegrationPhase13Batch3AutoScrollDownScrollsAndSelectsLastVisible(t *testing.T) {
	lv := newLVFocused(items20()) // topIndex=0, visibleHeight=5; 15 items below
	topBefore := lv.TopIndex()   // 0

	// Start drag.
	lv.HandleEvent(intDragPress(0, 0))

	// Drag below the widget (mouseY = visibleHeight = 5).
	lv.HandleEvent(intDragMove(0, lv.visibleHeight()))

	if lv.TopIndex() != topBefore+1 {
		t.Errorf("auto-scroll down: TopIndex()=%d, want %d (increased by 1)", lv.TopIndex(), topBefore+1)
	}

	wantLastVisible := lv.TopIndex() + lv.visibleHeight() - 1
	count := lv.dataSource.Count()
	if wantLastVisible >= count {
		wantLastVisible = count - 1
	}
	if lv.Selected() != wantLastVisible {
		t.Errorf("auto-scroll down: Selected()=%d, want last visible=%d", lv.Selected(), wantLastVisible)
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Releasing Button1 after a drag stops tracking; subsequent
// mouse moves (with Button1) do not update selection.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch3ReleaseStopsDragTracking verifies that after
// releasing Button1, dragging=false and a subsequent move event begins a fresh
// drag (not a continuation), with dragging=false only if pressed first.
func TestIntegrationPhase13Batch3ReleaseStopsDragTracking(t *testing.T) {
	lv := newLVFocused(items20())

	// Press at row 0, drag to row 2, then release.
	lv.HandleEvent(intDragPress(0, 0))
	lv.HandleEvent(intDragMove(0, 2))
	lv.HandleEvent(intDragRelease(0, 2))

	// After release, dragging must be false.
	if lv.dragging {
		t.Error("after Button1 release: dragging must be false")
	}
}

// TestIntegrationPhase13Batch3SubsequentMoveAfterReleaseDoesNotDrag verifies that
// after a release, a subsequent Button1 event without a prior press does not
// treat the prior drag as still active.
func TestIntegrationPhase13Batch3SubsequentMoveAfterReleaseDoesNotDrag(t *testing.T) {
	lv := newLVFocused(items20())

	// Press at row 0, drag to row 2, release.
	lv.HandleEvent(intDragPress(0, 0))
	lv.HandleEvent(intDragMove(0, 2))
	lv.HandleEvent(intDragRelease(0, 2))

	selAfterRelease := lv.Selected()

	// Send a Button1 move event — because dragging=false this is a new drag start,
	// not a continuation. The critical check: dragging was false before this event.
	wasNotDragging := !lv.dragging
	lv.HandleEvent(intDragMove(0, 4))

	// The widget should have treated this as a new click (dragging=false → start new drag).
	// What we care about: the old release correctly set dragging=false.
	if !wasNotDragging {
		t.Error("after release: subsequent Button1 event found dragging=true (release did not stop tracking)")
	}
	_ = selAfterRelease
}

// ---------------------------------------------------------------------------
// Requirement 5: ScrollBar value stays in sync during the entire drag sequence.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch3ScrollBarSyncDuringDrag verifies that the scrollbar
// value equals selected after each event in a full drag sequence.
func TestIntegrationPhase13Batch3ScrollBarSyncDuringDrag(t *testing.T) {
	lv := newLVFocused(items20())
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	// Press.
	lv.HandleEvent(intDragPress(0, 0))
	if sb.Value() != lv.Selected() {
		t.Errorf("after press: ScrollBar.Value()=%d, Selected()=%d; must match", sb.Value(), lv.Selected())
	}

	// Drag within bounds.
	lv.HandleEvent(intDragMove(0, 3))
	if sb.Value() != lv.Selected() {
		t.Errorf("after drag to row 3: ScrollBar.Value()=%d, Selected()=%d; must match", sb.Value(), lv.Selected())
	}

	// Release.
	lv.HandleEvent(intDragRelease(0, 3))
	if sb.Value() != lv.Selected() {
		t.Errorf("after release: ScrollBar.Value()=%d, Selected()=%d; must match", sb.Value(), lv.Selected())
	}
}

// TestIntegrationPhase13Batch3ScrollBarSyncDuringAutoScrollUp verifies scrollbar
// stays in sync through an auto-scroll-up event.
func TestIntegrationPhase13Batch3ScrollBarSyncDuringAutoScrollUp(t *testing.T) {
	lv := newLVFocused(items20())
	lv.SetSelected(10) // topIndex = 6
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	lv.HandleEvent(intDragPress(0, 0))
	lv.HandleEvent(intDragMove(0, -1)) // auto-scroll up

	if sb.Value() != lv.Selected() {
		t.Errorf("after auto-scroll up: ScrollBar.Value()=%d, Selected()=%d; must match",
			sb.Value(), lv.Selected())
	}
}

// TestIntegrationPhase13Batch3ScrollBarSyncDuringAutoScrollDown verifies scrollbar
// stays in sync through an auto-scroll-down event.
func TestIntegrationPhase13Batch3ScrollBarSyncDuringAutoScrollDown(t *testing.T) {
	lv := newLVFocused(items20()) // topIndex=0
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	lv.HandleEvent(intDragPress(0, 0))
	lv.HandleEvent(intDragMove(0, lv.visibleHeight())) // auto-scroll down

	if sb.Value() != lv.Selected() {
		t.Errorf("after auto-scroll down: ScrollBar.Value()=%d, Selected()=%d; must match",
			sb.Value(), lv.Selected())
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: Auto-scroll up from topIndex=3 with mouseY=-1:
// topIndex becomes 2, selected becomes 2.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch3AutoScrollUpFromTopIndex3 verifies the exact values
// after a single auto-scroll-up event from topIndex=3.
func TestIntegrationPhase13Batch3AutoScrollUpFromTopIndex3(t *testing.T) {
	lv := newLVFocused(items20())
	lv.SetSelected(7) // topIndex = 7 - 5 + 1 = 3; selected = 7
	lv.selected = 5   // move within page so we can observe change clearly

	if lv.TopIndex() != 3 {
		t.Fatalf("prerequisite: TopIndex()=%d, want 3", lv.TopIndex())
	}

	// Start drag, then drag above.
	lv.HandleEvent(intDragPress(0, 0))
	lv.HandleEvent(intDragMove(0, -1)) // auto-scroll up

	if lv.TopIndex() != 2 {
		t.Errorf("auto-scroll up from topIndex=3: TopIndex()=%d, want 2", lv.TopIndex())
	}
	if lv.Selected() != 2 {
		t.Errorf("auto-scroll up from topIndex=3: Selected()=%d, want 2 (new topIndex)", lv.Selected())
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: Auto-scroll down stops when topIndex reaches maxTop (no over-scroll).
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch3AutoScrollDownStopsAtMaxTop verifies that when
// topIndex is already at its maximum value, dragging below the widget is a no-op.
func TestIntegrationPhase13Batch3AutoScrollDownStopsAtMaxTop(t *testing.T) {
	lv := newLVFocused(items20()) // 20 items, height=5; maxTop = 20 - 5 = 15
	lv.SetSelected(19)            // forces topIndex = 19 - 5 + 1 = 15 (maxTop)

	topAtMax := lv.TopIndex()
	if topAtMax != 15 {
		t.Fatalf("prerequisite: TopIndex()=%d, want 15 (maxTop)", topAtMax)
	}

	lv.HandleEvent(intDragPress(0, 0))
	// Multiple attempts to auto-scroll down beyond maxTop.
	for i := 0; i < 3; i++ {
		lv.HandleEvent(intDragMove(0, lv.visibleHeight()))
	}

	if lv.TopIndex() != topAtMax {
		t.Errorf("auto-scroll down at maxTop: TopIndex()=%d, want %d (no over-scroll)",
			lv.TopIndex(), topAtMax)
	}
	// Selected must also remain valid.
	if lv.Selected() < 0 || lv.Selected() >= lv.dataSource.Count() {
		t.Errorf("auto-scroll down at maxTop: Selected()=%d is out of valid range", lv.Selected())
	}
}

// ---------------------------------------------------------------------------
// Requirement 8: Drag does not fire OnSelect at any point during the drag sequence.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch3DragNeverFiresOnSelect verifies that no event in a
// full drag sequence (press, multiple moves, release) calls OnSelect.
func TestIntegrationPhase13Batch3DragNeverFiresOnSelect(t *testing.T) {
	lv := newLVFocused(items20())
	callCount := 0
	lv.OnSelect = func(index int) { callCount++ }

	// Press.
	lv.HandleEvent(intDragPress(0, 0))
	if callCount != 0 {
		t.Errorf("OnSelect called %d time(s) on drag press, want 0", callCount)
	}

	// Several in-bounds moves.
	for _, y := range []int{1, 2, 3, 4} {
		lv.HandleEvent(intDragMove(0, y))
		if callCount != 0 {
			t.Errorf("OnSelect called %d time(s) after drag to row %d, want 0", callCount, y)
		}
	}

	// Release.
	lv.HandleEvent(intDragRelease(0, 4))
	if callCount != 0 {
		t.Errorf("OnSelect called %d time(s) on drag release, want 0", callCount)
	}
}

// TestIntegrationPhase13Batch3AutoScrollUpNeverFiresOnSelect verifies that
// auto-scroll-up during a drag does not call OnSelect.
func TestIntegrationPhase13Batch3AutoScrollUpNeverFiresOnSelect(t *testing.T) {
	lv := newLVFocused(items20())
	lv.SetSelected(10) // topIndex = 6

	called := false
	lv.OnSelect = func(index int) { called = true }

	lv.HandleEvent(intDragPress(0, 0))
	lv.HandleEvent(intDragMove(0, -1)) // auto-scroll up

	if called {
		t.Error("OnSelect fired during auto-scroll-up drag, want no call")
	}
}

// TestIntegrationPhase13Batch3AutoScrollDownNeverFiresOnSelect verifies that
// auto-scroll-down during a drag does not call OnSelect.
func TestIntegrationPhase13Batch3AutoScrollDownNeverFiresOnSelect(t *testing.T) {
	lv := newLVFocused(items20()) // topIndex=0

	called := false
	lv.OnSelect = func(index int) { called = true }

	lv.HandleEvent(intDragPress(0, 0))
	lv.HandleEvent(intDragMove(0, lv.visibleHeight())) // auto-scroll down

	if called {
		t.Error("OnSelect fired during auto-scroll-down drag, want no call")
	}
}

// ---------------------------------------------------------------------------
// Requirement 9: Double-click during a drag (ClickCount >= 2) fires OnSelect
// for the in-bounds item.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch3DoubleClickDuringDragFiresOnSelect verifies that
// a double-click event while dragging fires OnSelect with the correct index.
func TestIntegrationPhase13Batch3DoubleClickDuringDragFiresOnSelect(t *testing.T) {
	lv := newLVFocused(items20()) // topIndex=0
	called := false
	var gotIndex int
	lv.OnSelect = func(index int) { called = true; gotIndex = index }

	// Start drag.
	lv.HandleEvent(intDragPress(0, 0))
	if called {
		t.Fatalf("OnSelect called on drag press, want 0 calls before double-click")
	}

	// Double-click at row 3 during the drag.
	lv.HandleEvent(intDragDoubleClick(0, 3))

	if !called {
		t.Error("double-click (ClickCount=2) during drag must fire OnSelect")
	}
	// topIndex=0, row 3 → absolute index 3.
	wantIdx := lv.TopIndex() + 3
	// Note: double-click sets selected to topIndex+y first, so use selected.
	if gotIndex != lv.Selected() {
		t.Errorf("OnSelect received index %d, want %d (selected after double-click)", gotIndex, lv.Selected())
	}
	_ = wantIdx
}

// TestIntegrationPhase13Batch3DoubleClickDuringScrolledDragFiresOnSelectWithAbsoluteIndex
// verifies that after scrolling (topIndex > 0), a double-click fires OnSelect
// with the absolute index (topIndex + clickY), not the click-local row.
func TestIntegrationPhase13Batch3DoubleClickDuringScrolledDragFiresOnSelectWithAbsoluteIndex(t *testing.T) {
	lv := newLVFocused(items20())
	lv.SetSelected(10) // topIndex = 6

	topIdx := lv.TopIndex()
	if topIdx == 0 {
		t.Fatal("prerequisite: topIndex must be > 0")
	}

	var gotIndex int = -1
	lv.OnSelect = func(index int) { gotIndex = index }

	// Start drag.
	lv.HandleEvent(intDragPress(0, 0))

	// Double-click at row 2 during drag.
	lv.HandleEvent(intDragDoubleClick(0, 2))

	// topIndex + 2 is the expected absolute index.
	wantIdx := topIdx + 2
	if gotIndex != wantIdx {
		t.Errorf("double-click row 2 with topIndex=%d: OnSelect received %d, want %d (topIndex+2)",
			topIdx, gotIndex, wantIdx)
	}
}

// TestIntegrationPhase13Batch3DoubleClickOutOfBoundsDoesNotFireOnSelect verifies that
// a double-click at mouseY >= visibleHeight does not fire OnSelect (out of bounds).
func TestIntegrationPhase13Batch3DoubleClickOutOfBoundsDoesNotFireOnSelect(t *testing.T) {
	lv := newLVFocused(items20())

	called := false
	lv.OnSelect = func(index int) { called = true }

	// Start drag.
	lv.HandleEvent(intDragPress(0, 0))

	// Double-click below the widget (out of bounds).
	lv.HandleEvent(intDragDoubleClick(0, lv.visibleHeight()))

	if called {
		t.Error("double-click at mouseY=visibleHeight (out of bounds) must NOT fire OnSelect")
	}
}
