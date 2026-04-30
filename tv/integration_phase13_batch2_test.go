package tv

// integration_phase13_batch2_test.go — Integration tests for Task 6:
// Home/End with scrolling in TurboView ListViewer.
//
// Each test exercises Home, End, Ctrl+PgUp, or Ctrl+PgDn with a scrolled
// list (topIndex > 0), verifying both selection movement and topIndex/scrollbar sync.
//
// Test naming: TestIntegrationPhase13Batch2<DescriptiveSuffix>

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ctrlPgUpEv returns a Ctrl+PgUp keyboard event.
func ctrlPgUpEv() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp, Modifiers: tcell.ModCtrl}}
}

// ctrlPgDnEv returns a Ctrl+PgDn keyboard event.
func ctrlPgDnEv() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgDn, Modifiers: tcell.ModCtrl}}
}

// items20 returns a 20-item string slice ("item0" … "item19").
// With height=5 (the default from newLVFocused), the list scrolls and leaves
// enough room to verify scrolling semantics in all directions.
func items20() []string {
	items := make([]string, 20)
	for i := range items {
		items[i] = "item" + string(rune('0'+i/10)) + string(rune('0'+i%10))
	}
	return items
}

// scrolledLV returns a focused ListViewer with 20 items, height=5,
// scrolled so that topIndex=10 (items 10–14 visible).
// selected is placed at 12 (mid-page) so Home and End can move meaningfully.
func scrolledLV() *ListViewer {
	lv := newLVFocused(items20())
	// SetSelected(14) gives topIndex = 14-5+1 = 10; then move selected to 12.
	lv.SetSelected(14)
	lv.selected = 12 // direct field access (same package) — topIndex stays 10
	return lv
}

// ---------------------------------------------------------------------------
// Requirement 1: With a scrolled list (topIndex > 0), Home selects topIndex, not 0.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch2HomeSelectsTopIndexNotZero verifies that pressing
// Home on a scrolled list (topIndex=10) selects item 10 (topIndex), not item 0.
func TestIntegrationPhase13Batch2HomeSelectsTopIndexNotZero(t *testing.T) {
	lv := scrolledLV()

	if lv.TopIndex() != 10 {
		t.Fatalf("prerequisite: TopIndex()=%d, want 10", lv.TopIndex())
	}

	lv.HandleEvent(listKeyEv(tcell.KeyHome))

	if lv.Selected() == 0 {
		t.Errorf("Home on scrolled list: Selected()=0, but should be topIndex=10 (not absolute start)")
	}
	if lv.Selected() != lv.TopIndex() {
		t.Errorf("Home: Selected()=%d, want %d (topIndex)", lv.Selected(), lv.TopIndex())
	}
}

// TestIntegrationPhase13Batch2HomeDoesNotChangeTopIndex verifies that pressing
// Home does not alter topIndex (page-local navigation only).
func TestIntegrationPhase13Batch2HomeDoesNotChangeTopIndex(t *testing.T) {
	lv := scrolledLV()
	topBefore := lv.TopIndex() // 10

	lv.HandleEvent(listKeyEv(tcell.KeyHome))

	if lv.TopIndex() != topBefore {
		t.Errorf("Home changed TopIndex: got %d, want %d (unchanged)", lv.TopIndex(), topBefore)
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: With a scrolled list, End selects the last visible item, not the last in the list.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch2EndSelectsLastVisibleNotLastItem verifies that pressing
// End on a scrolled list (topIndex=10, height=5) selects item 14, not item 19.
func TestIntegrationPhase13Batch2EndSelectsLastVisibleNotLastItem(t *testing.T) {
	lv := scrolledLV()

	if lv.TopIndex() != 10 {
		t.Fatalf("prerequisite: TopIndex()=%d, want 10", lv.TopIndex())
	}

	lv.HandleEvent(listKeyEv(tcell.KeyEnd))

	lastItem := lv.dataSource.Count() - 1 // 19
	if lv.Selected() == lastItem {
		t.Errorf("End on scrolled list: Selected()=%d (last item in list), but should be last visible item %d",
			lv.Selected(), lv.TopIndex()+lv.visibleHeight()-1)
	}
	wantLastVisible := lv.TopIndex() + lv.visibleHeight() - 1 // 10+5-1=14
	if lv.Selected() != wantLastVisible {
		t.Errorf("End: Selected()=%d, want %d (topIndex+height-1=%d+%d-1)",
			lv.Selected(), wantLastVisible, lv.TopIndex(), lv.visibleHeight())
	}
}

// TestIntegrationPhase13Batch2EndDoesNotChangeTopIndex verifies that pressing
// End does not alter topIndex.
func TestIntegrationPhase13Batch2EndDoesNotChangeTopIndex(t *testing.T) {
	lv := scrolledLV()
	topBefore := lv.TopIndex() // 10

	lv.HandleEvent(listKeyEv(tcell.KeyEnd))

	if lv.TopIndex() != topBefore {
		t.Errorf("End changed TopIndex: got %d, want %d (unchanged)", lv.TopIndex(), topBefore)
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: After pressing Home, pressing End moves to last visible item on the
// same page (topIndex unchanged).
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch2HomeThenEndStaysOnSamePage verifies the sequence
// Home → End leaves topIndex unchanged and End selects the last visible item.
func TestIntegrationPhase13Batch2HomeThenEndStaysOnSamePage(t *testing.T) {
	lv := scrolledLV()
	topBefore := lv.TopIndex() // 10

	lv.HandleEvent(listKeyEv(tcell.KeyHome))

	if lv.TopIndex() != topBefore {
		t.Fatalf("Home changed TopIndex: got %d, want %d", lv.TopIndex(), topBefore)
	}
	if lv.Selected() != topBefore {
		t.Fatalf("Home: Selected()=%d, want %d (topIndex)", lv.Selected(), topBefore)
	}

	lv.HandleEvent(listKeyEv(tcell.KeyEnd))

	if lv.TopIndex() != topBefore {
		t.Errorf("Home→End changed TopIndex: got %d, want %d (unchanged)", lv.TopIndex(), topBefore)
	}
	wantLastVisible := topBefore + lv.visibleHeight() - 1 // 10+5-1=14
	if lv.Selected() != wantLastVisible {
		t.Errorf("Home→End: Selected()=%d, want %d (last visible on same page)", lv.Selected(), wantLastVisible)
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Ctrl+PgUp from a scrolled position resets both selected and topIndex to 0.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch2CtrlPgUpResetsSelectedToZero verifies that Ctrl+PgUp
// from topIndex=10 sets selected to 0.
func TestIntegrationPhase13Batch2CtrlPgUpResetsSelectedToZero(t *testing.T) {
	lv := scrolledLV()

	lv.HandleEvent(ctrlPgUpEv())

	if lv.Selected() != 0 {
		t.Errorf("Ctrl+PgUp: Selected()=%d, want 0 (absolute start)", lv.Selected())
	}
}

// TestIntegrationPhase13Batch2CtrlPgUpResetsTopIndexToZero verifies that Ctrl+PgUp
// from topIndex=10 sets topIndex to 0.
func TestIntegrationPhase13Batch2CtrlPgUpResetsTopIndexToZero(t *testing.T) {
	lv := scrolledLV()

	lv.HandleEvent(ctrlPgUpEv())

	if lv.TopIndex() != 0 {
		t.Errorf("Ctrl+PgUp: TopIndex()=%d, want 0 (absolute start)", lv.TopIndex())
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: Ctrl+PgDn from the top moves selected to last item and scrolls to show it.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch2CtrlPgDnSelectsLastItem verifies that Ctrl+PgDn
// from the top (topIndex=0, selected=0) moves selected to the last item (19).
func TestIntegrationPhase13Batch2CtrlPgDnSelectsLastItem(t *testing.T) {
	lv := newLVFocused(items20()) // 20 items, height=5, topIndex=0

	lv.HandleEvent(ctrlPgDnEv())

	lastItem := lv.dataSource.Count() - 1 // 19
	if lv.Selected() != lastItem {
		t.Errorf("Ctrl+PgDn: Selected()=%d, want %d (last item)", lv.Selected(), lastItem)
	}
}

// TestIntegrationPhase13Batch2CtrlPgDnScrollsToShowLastItem verifies that Ctrl+PgDn
// adjusts topIndex so the last item is visible.
func TestIntegrationPhase13Batch2CtrlPgDnScrollsToShowLastItem(t *testing.T) {
	lv := newLVFocused(items20()) // 20 items, height=5

	lv.HandleEvent(ctrlPgDnEv())

	lastItem := lv.dataSource.Count() - 1         // 19
	wantTopIndex := lastItem - lv.visibleHeight() + 1 // 19-5+1=15

	if lv.TopIndex() != wantTopIndex {
		t.Errorf("Ctrl+PgDn: TopIndex()=%d, want %d (scrolled to show last item)", lv.TopIndex(), wantTopIndex)
	}
	// Confirm last item is within the visible range.
	if lastItem < lv.TopIndex() || lastItem >= lv.TopIndex()+lv.visibleHeight() {
		t.Errorf("Ctrl+PgDn: last item %d not visible in range [%d, %d)",
			lastItem, lv.TopIndex(), lv.TopIndex()+lv.visibleHeight())
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: After Ctrl+PgDn, pressing Home selects topIndex (now showing the last page).
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch2CtrlPgDnThenHomeSelectsNewTopIndex verifies the
// sequence Ctrl+PgDn → Home: topIndex is now 15 (last page), so Home selects 15.
func TestIntegrationPhase13Batch2CtrlPgDnThenHomeSelectsNewTopIndex(t *testing.T) {
	lv := newLVFocused(items20()) // 20 items, height=5

	lv.HandleEvent(ctrlPgDnEv())

	topAfterCtrlPgDn := lv.TopIndex() // should be 15

	lv.HandleEvent(listKeyEv(tcell.KeyHome))

	if lv.Selected() != topAfterCtrlPgDn {
		t.Errorf("Ctrl+PgDn→Home: Selected()=%d, want %d (new topIndex after Ctrl+PgDn)",
			lv.Selected(), topAfterCtrlPgDn)
	}
	// topIndex must still equal topAfterCtrlPgDn (Home does not change it).
	if lv.TopIndex() != topAfterCtrlPgDn {
		t.Errorf("Ctrl+PgDn→Home: TopIndex()=%d, want %d (unchanged by Home)",
			lv.TopIndex(), topAfterCtrlPgDn)
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: ScrollBar value stays in sync with topIndex after all operations.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch2ScrollBarSyncAfterHome verifies scrollBar.Value()
// equals topIndex after pressing Home on a scrolled list.
func TestIntegrationPhase13Batch2ScrollBarSyncAfterHome(t *testing.T) {
	lv := scrolledLV()
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	lv.HandleEvent(listKeyEv(tcell.KeyHome))

	if sb.Value() != lv.TopIndex() {
		t.Errorf("Home: ScrollBar.Value()=%d, TopIndex()=%d; must be equal", sb.Value(), lv.TopIndex())
	}
}

// TestIntegrationPhase13Batch2ScrollBarSyncAfterEnd verifies scrollBar.Value()
// equals topIndex after pressing End on a scrolled list.
func TestIntegrationPhase13Batch2ScrollBarSyncAfterEnd(t *testing.T) {
	lv := scrolledLV()
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	lv.HandleEvent(listKeyEv(tcell.KeyEnd))

	if sb.Value() != lv.TopIndex() {
		t.Errorf("End: ScrollBar.Value()=%d, TopIndex()=%d; must be equal", sb.Value(), lv.TopIndex())
	}
}

// ---------------------------------------------------------------------------
// Requirement 8: After Ctrl+PgUp, scrollBar value is 0.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch2ScrollBarZeroAfterCtrlPgUp verifies that after
// Ctrl+PgUp from a scrolled list, the scrollbar value is 0.
func TestIntegrationPhase13Batch2ScrollBarZeroAfterCtrlPgUp(t *testing.T) {
	lv := scrolledLV()
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	lv.HandleEvent(ctrlPgUpEv())

	if sb.Value() != 0 {
		t.Errorf("Ctrl+PgUp: ScrollBar.Value()=%d, want 0", sb.Value())
	}
	if lv.TopIndex() != 0 {
		t.Errorf("Ctrl+PgUp: TopIndex()=%d, want 0", lv.TopIndex())
	}
	if sb.Value() != lv.TopIndex() {
		t.Errorf("Ctrl+PgUp: ScrollBar.Value()=%d != TopIndex()=%d; must be equal",
			sb.Value(), lv.TopIndex())
	}
}

// ---------------------------------------------------------------------------
// Requirement 9: After Ctrl+PgDn, scrollBar value equals the new topIndex.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch2ScrollBarEqualsNewTopIndexAfterCtrlPgDn verifies
// that after Ctrl+PgDn, scrollBar.Value() equals the new topIndex (15 for 20 items, height=5).
func TestIntegrationPhase13Batch2ScrollBarEqualsNewTopIndexAfterCtrlPgDn(t *testing.T) {
	lv := newLVFocused(items20()) // 20 items, height=5, topIndex=0
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	lv.HandleEvent(ctrlPgDnEv())

	wantTopIndex := lv.dataSource.Count() - lv.visibleHeight() // 20-5=15
	if lv.TopIndex() != wantTopIndex {
		t.Errorf("Ctrl+PgDn: TopIndex()=%d, want %d", lv.TopIndex(), wantTopIndex)
	}
	if sb.Value() != lv.TopIndex() {
		t.Errorf("Ctrl+PgDn: ScrollBar.Value()=%d != TopIndex()=%d; must be equal",
			sb.Value(), lv.TopIndex())
	}
	if sb.Value() != wantTopIndex {
		t.Errorf("Ctrl+PgDn: ScrollBar.Value()=%d, want %d", sb.Value(), wantTopIndex)
	}
}

// ---------------------------------------------------------------------------
// Bonus: OnSelect is never called by any of these navigation keys.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13Batch2NoOnSelectFromNavKeys verifies that Home, End,
// Ctrl+PgUp, and Ctrl+PgDn never fire OnSelect.
func TestIntegrationPhase13Batch2NoOnSelectFromNavKeys(t *testing.T) {
	for _, tc := range []struct {
		name string
		ev   *Event
	}{
		{"Home", listKeyEv(tcell.KeyHome)},
		{"End", listKeyEv(tcell.KeyEnd)},
		{"Ctrl+PgUp", ctrlPgUpEv()},
		{"Ctrl+PgDn", ctrlPgDnEv()},
	} {
		lv := newLVFocused(items20())
		lv.SetSelected(12) // scroll to a mid-list position

		called := false
		lv.OnSelect = func(index int) { called = true }

		lv.HandleEvent(tc.ev)

		if called {
			t.Errorf("%s: OnSelect was called (navigation must not fire OnSelect)", tc.name)
		}
	}
}
