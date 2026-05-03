package tv

// list_viewer_homeend_test.go — Tests for Task 5: Home/End page-local, Ctrl+PgUp/PgDn absolute.
//
// Written BEFORE any implementation exists; all tests drive spec 7.4.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Spec 7.4:
//   - Home: go to topIndex (first visible item). Do NOT change topIndex.
//   - End:  go to topIndex + visibleHeight - 1 clamped to count-1 (last visible item). Do NOT change topIndex.
//   - Ctrl+PgUp: go to item 0 (absolute start). Set topIndex to 0.
//   - Ctrl+PgDn: go to count-1 (absolute end). Call ensureVisible to adjust topIndex.
//
// Test organisation:
//   Section 1  — Home tests (page-local)
//   Section 2  — End tests (page-local)
//   Section 3  — Ctrl+PgUp tests (absolute start)
//   Section 4  — Ctrl+PgDn tests (absolute end)
//   Section 5  — Falsification tests

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Helpers — Ctrl+key events (do NOT redeclare newLV, newLVFocused, listKeyEv)
// ---------------------------------------------------------------------------

// listCtrlKeyEv creates a keyboard event with the Ctrl modifier set.
// Use this for Ctrl+PgUp and Ctrl+PgDn.
func listCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Rune: 0, Modifiers: tcell.ModCtrl}}
}

// ---------------------------------------------------------------------------
// Section 1 — Home tests (page-local)
// ---------------------------------------------------------------------------

// TestHomeFromScrolledPositionSelectsTopIndex verifies Home navigates to topIndex, not item 0.
// Spec 7.4: "Home: go to topIndex (first visible item)."
func TestHomeFromScrolledPositionSelectsTopIndex(t *testing.T) {
	// 10 items; scroll so topIndex=3, then press Home from a lower position.
	items := make10Items()
	lv := newLVFocused(items)
	// Force topIndex=3 by selecting item 7 (7-5+1=3), then move back to item 5 without scrolling.
	lv.SetSelected(7) // topIndex becomes 3
	lv.selected = 5   // place cursor mid-page without triggering ensureVisible

	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	// Spec 7.4: Home → selected = topIndex (3), NOT 0.
	if lv.Selected() != 3 {
		t.Errorf("spec 7.4: Home with topIndex=3: Selected()=%d, want 3 (topIndex, not item 0)", lv.Selected())
	}
}

// TestHomeDoesNotChangeTopIndex verifies Home does not scroll — topIndex stays put.
// Spec 7.4: "Home: go to topIndex (first visible item). Do NOT change topIndex."
func TestHomeDoesNotChangeTopIndex(t *testing.T) {
	items := make10Items()
	lv := newLVFocused(items)
	lv.SetSelected(7) // topIndex becomes 3
	lv.selected = 5   // cursor mid-page

	topBefore := lv.TopIndex() // should be 3

	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	if lv.TopIndex() != topBefore {
		t.Errorf("spec 7.4: Home must not change topIndex; was %d, now %d", topBefore, lv.TopIndex())
	}
}

// TestHomeWhenAlreadyAtTopIndexIsNoOp verifies Home when cursor is already at topIndex
// leaves selected unchanged (no-op).
// Spec 7.4: "Home when already at topIndex: no-op."
func TestHomeWhenAlreadyAtTopIndexIsNoOp(t *testing.T) {
	items := make10Items()
	lv := newLVFocused(items)
	lv.SetSelected(7) // topIndex becomes 3; selected=3 after ensureVisible would be 3 but SetSelected sets it to 7
	// Now place cursor at topIndex exactly.
	lv.selected = lv.TopIndex() // cursor == topIndex == 3

	selectedBefore := lv.Selected()

	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	if lv.Selected() != selectedBefore {
		t.Errorf("spec 7.4: Home at topIndex is no-op; selected was %d, now %d", selectedBefore, lv.Selected())
	}
}

// TestHomeConsumesEvent verifies Home always clears the event.
// Spec 7.4: "All four keys consume the event (Clear)."
func TestHomeConsumesEvent(t *testing.T) {
	lv := newLVFocused(make10Items())
	lv.SetSelected(5)

	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.4: Home must consume (clear) the event")
	}
}

// TestHomeDoesNotCallOnSelect verifies Home does not fire OnSelect.
// Spec 7.1 / 7.4: "None of these keys call OnSelect (navigation, not selection)."
func TestHomeDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused(make10Items())
	lv.SetSelected(7) // scroll so topIndex=3
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.4: Home must NOT call OnSelect (navigation key, not selection)")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — End tests (page-local)
// ---------------------------------------------------------------------------

// TestEndFromScrolledPositionSelectsLastVisible verifies End navigates to the last
// visible item (topIndex + visibleHeight - 1), not the absolute last item.
// Spec 7.4: "End: go to topIndex + pageSize - 1 (last visible item)."
// Setup: topIndex=3, height=5, 10 items → last visible = 3+4 = 7, NOT 9.
func TestEndFromScrolledPositionSelectsLastVisible(t *testing.T) {
	items := make10Items()
	lv := newLVFocused(items)
	lv.SetSelected(7) // topIndex becomes 3
	lv.selected = 3   // put cursor at topIndex (first visible)

	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	// visibleHeight=5, topIndex=3 → last visible = 3 + 5 - 1 = 7 (NOT 9)
	if lv.Selected() != 7 {
		t.Errorf("spec 7.4: End with topIndex=3,height=5,count=10: Selected()=%d, want 7 (last visible, not 9)", lv.Selected())
	}
}

// TestEndDoesNotChangeTopIndex verifies End does not scroll — topIndex stays put.
// Spec 7.4: "End: go to topIndex + pageSize - 1 (last visible item). Do NOT change topIndex."
func TestEndDoesNotChangeTopIndex(t *testing.T) {
	items := make10Items()
	lv := newLVFocused(items)
	lv.SetSelected(7) // topIndex becomes 3
	lv.selected = 3   // cursor at first visible

	topBefore := lv.TopIndex() // 3

	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	if lv.TopIndex() != topBefore {
		t.Errorf("spec 7.4: End must not change topIndex; was %d, now %d", topBefore, lv.TopIndex())
	}
}

// TestEndWithFewerItemsThanHeight verifies End clamps to count-1 when the list is shorter
// than the visible area.
// Spec 7.4: "End: go to min(topIndex + visibleHeight - 1, count - 1)."
// Setup: 3 items, height=5, topIndex=0 → last visible = min(4, 2) = 2.
func TestEndWithFewerItemsThanHeight(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"}) // 3 items, height=5

	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	// min(0+5-1, 3-1) = min(4, 2) = 2
	if lv.Selected() != 2 {
		t.Errorf("spec 7.4: End with 3 items, height=5: Selected()=%d, want 2 (count-1)", lv.Selected())
	}
}

// TestEndWhenAlreadyAtLastVisibleIsNoOp verifies End when cursor is already at the last
// visible item leaves selected unchanged.
// Spec 7.4: "End when already at last visible item: no-op."
func TestEndWhenAlreadyAtLastVisibleIsNoOp(t *testing.T) {
	items := make10Items()
	lv := newLVFocused(items)
	lv.SetSelected(7) // topIndex=3
	// cursor is at 7, which is exactly topIndex+visibleHeight-1 = 3+5-1 = 7

	selectedBefore := lv.Selected() // 7

	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	if lv.Selected() != selectedBefore {
		t.Errorf("spec 7.4: End at last visible is no-op; selected was %d, now %d", selectedBefore, lv.Selected())
	}
}

// TestEndConsumesEvent verifies End always clears the event.
// Spec 7.4: "All four keys consume the event (Clear)."
func TestEndConsumesEvent(t *testing.T) {
	lv := newLVFocused(make10Items())

	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.4: End must consume (clear) the event")
	}
}

// TestEndDoesNotCallOnSelect verifies End does not fire OnSelect.
// Spec 7.1 / 7.4: "None of these keys call OnSelect (navigation, not selection)."
func TestEndDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused(make10Items())
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.4: End must NOT call OnSelect (navigation key, not selection)")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Ctrl+PgUp tests (absolute start)
// ---------------------------------------------------------------------------

// TestCtrlPgUpFromScrolledPositionSelectsItemZero verifies Ctrl+PgUp jumps to absolute
// start: selected=0 and topIndex=0, regardless of current scroll position.
// Spec 7.4: "Ctrl+PgUp: go to item 0 (absolute start). Set topIndex to 0."
func TestCtrlPgUpFromScrolledPositionSelectsItemZero(t *testing.T) {
	items := make10Items()
	lv := newLVFocused(items)
	lv.SetSelected(7) // topIndex becomes 3, selected=7

	ev := listCtrlKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	if lv.Selected() != 0 {
		t.Errorf("spec 7.4: Ctrl+PgUp: Selected()=%d, want 0 (absolute start)", lv.Selected())
	}
	if lv.TopIndex() != 0 {
		t.Errorf("spec 7.4: Ctrl+PgUp: TopIndex()=%d, want 0 (absolute start)", lv.TopIndex())
	}
}

// TestCtrlPgUpConsumesEvent verifies Ctrl+PgUp always clears the event.
// Spec 7.4: "All four keys consume the event (Clear)."
func TestCtrlPgUpConsumesEvent(t *testing.T) {
	lv := newLVFocused(make10Items())
	lv.SetSelected(5)

	ev := listCtrlKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.4: Ctrl+PgUp must consume (clear) the event")
	}
}

// TestCtrlPgUpDoesNotCallOnSelect verifies Ctrl+PgUp does not fire OnSelect.
// Spec 7.1 / 7.4: "None of these keys call OnSelect (navigation, not selection)."
func TestCtrlPgUpDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused(make10Items())
	lv.SetSelected(7)
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listCtrlKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.4: Ctrl+PgUp must NOT call OnSelect (navigation key, not selection)")
	}
}

// TestCtrlPgUpWhenAlreadyAtZeroIsNoOpButConsumed verifies Ctrl+PgUp when cursor is
// already at item 0 and topIndex is 0: selected stays 0, event still consumed.
// Spec 7.4: "Ctrl+PgUp when already at 0: no-op, event consumed."
func TestCtrlPgUpWhenAlreadyAtZeroIsNoOpButConsumed(t *testing.T) {
	lv := newLVFocused(make10Items())
	// Initial state: selected=0, topIndex=0 (default)

	ev := listCtrlKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	if lv.Selected() != 0 {
		t.Errorf("spec 7.4: Ctrl+PgUp at item 0: Selected()=%d, want 0 (no-op)", lv.Selected())
	}
	if lv.TopIndex() != 0 {
		t.Errorf("spec 7.4: Ctrl+PgUp at item 0: TopIndex()=%d, want 0 (no-op)", lv.TopIndex())
	}
	if !ev.IsCleared() {
		t.Error("spec 7.4: Ctrl+PgUp at item 0 must still consume (clear) the event")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Ctrl+PgDn tests (absolute end)
// ---------------------------------------------------------------------------

// TestCtrlPgDnSelectsLastItem verifies Ctrl+PgDn jumps to the absolute last item.
// Spec 7.4: "Ctrl+PgDn: go to last item (absolute end). Set selected to count-1."
func TestCtrlPgDnSelectsLastItem(t *testing.T) {
	items := make10Items() // 10 items; last = index 9
	lv := newLVFocused(items)

	ev := listCtrlKeyEv(tcell.KeyPgDn)
	lv.HandleEvent(ev)

	if lv.Selected() != 9 {
		t.Errorf("spec 7.4: Ctrl+PgDn: Selected()=%d, want 9 (count-1=9)", lv.Selected())
	}
}

// TestCtrlPgDnAdjustsTopIndexToShowLastItem verifies Ctrl+PgDn calls ensureVisible so
// the last item is scrolled into view.
// Spec 7.4: "Ctrl+PgDn: call ensureVisible (absolute end)."
// With 10 items, height=5: topIndex must be 10-5=5 to show item 9.
func TestCtrlPgDnAdjustsTopIndexToShowLastItem(t *testing.T) {
	items := make10Items() // count=10, height=5
	lv := newLVFocused(items)

	ev := listCtrlKeyEv(tcell.KeyPgDn)
	lv.HandleEvent(ev)

	// ensureVisible with selected=9, height=5: topIndex = 9 - 5 + 1 = 5
	if lv.TopIndex() != 5 {
		t.Errorf("spec 7.4: Ctrl+PgDn with 10 items, height=5: TopIndex()=%d, want 5 (last item visible)", lv.TopIndex())
	}
}

// TestCtrlPgDnConsumesEvent verifies Ctrl+PgDn always clears the event.
// Spec 7.4: "All four keys consume the event (Clear)."
func TestCtrlPgDnConsumesEvent(t *testing.T) {
	lv := newLVFocused(make10Items())

	ev := listCtrlKeyEv(tcell.KeyPgDn)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("spec 7.4: Ctrl+PgDn must consume (clear) the event")
	}
}

// TestCtrlPgDnDoesNotCallOnSelect verifies Ctrl+PgDn does not fire OnSelect.
// Spec 7.1 / 7.4: "None of these keys call OnSelect (navigation, not selection)."
func TestCtrlPgDnDoesNotCallOnSelect(t *testing.T) {
	lv := newLVFocused(make10Items())
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listCtrlKeyEv(tcell.KeyPgDn)
	lv.HandleEvent(ev)

	if called {
		t.Error("spec 7.4: Ctrl+PgDn must NOT call OnSelect (navigation key, not selection)")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Falsification tests
// ---------------------------------------------------------------------------

// TestHomeGoesToTopIndexNotItemZero proves Home is page-local: when scrolled so
// topIndex=3, pressing Home must select item 3, NOT item 0.
// Spec 7.4: "Home: go to topIndex (first visible item)." — falsifies old absolute behaviour.
func TestHomeGoesToTopIndexNotItemZero(t *testing.T) {
	items := make10Items()
	lv := newLVFocused(items)
	lv.SetSelected(7) // topIndex=3
	lv.selected = 6   // cursor somewhere below topIndex

	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	// Must be topIndex (3), NOT 0.
	if lv.Selected() == 0 {
		t.Error("spec 7.4 FALSIFICATION: Home went to item 0 (absolute); it must go to topIndex=3 (page-local)")
	}
	if lv.Selected() != 3 {
		t.Errorf("spec 7.4 FALSIFICATION: Home: Selected()=%d, want 3 (topIndex, not 0)", lv.Selected())
	}
}

// TestEndGoesToLastVisibleNotLastItem proves End is page-local: when scrolled so
// topIndex=3 (visible items 3-7) in a 10-item list, pressing End must select item 7,
// NOT item 9.
// Spec 7.4: "End: go to topIndex + pageSize - 1 (last visible item)." — falsifies old absolute behaviour.
func TestEndGoesToLastVisibleNotLastItem(t *testing.T) {
	items := make10Items()
	lv := newLVFocused(items)
	lv.SetSelected(7) // topIndex=3, selected=7
	lv.selected = 3   // cursor at top of visible window

	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	// Must be 7 (topIndex+height-1 = 3+5-1), NOT 9 (count-1).
	if lv.Selected() == 9 {
		t.Error("spec 7.4 FALSIFICATION: End went to item 9 (absolute); it must go to last visible (7)")
	}
	if lv.Selected() != 7 {
		t.Errorf("spec 7.4 FALSIFICATION: End: Selected()=%d, want 7 (last visible, not 9)", lv.Selected())
	}
}

// TestCtrlPgUpGoesToItemZeroNotTopIndex proves Ctrl+PgUp is absolute: it sets selected
// to 0 and topIndex to 0, unlike plain Home which only sets selected to topIndex.
// Spec 7.4: "Ctrl+PgUp: go to item 0 (absolute start)." — falsifies page-local behaviour.
func TestCtrlPgUpGoesToItemZeroNotTopIndex(t *testing.T) {
	items := make10Items()
	lv := newLVFocused(items)
	lv.SetSelected(7) // topIndex=3, selected=7
	lv.selected = 5   // cursor mid-page

	topBefore := lv.TopIndex() // 3 — would stay unchanged if this were Home

	ev := listCtrlKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	// selected must be 0 (not topIndex=3)
	if lv.Selected() == topBefore {
		t.Errorf("spec 7.4 FALSIFICATION: Ctrl+PgUp went to topIndex=%d (page-local); it must go to 0 (absolute)", topBefore)
	}
	if lv.Selected() != 0 {
		t.Errorf("spec 7.4 FALSIFICATION: Ctrl+PgUp: Selected()=%d, want 0 (absolute start)", lv.Selected())
	}
	// topIndex must also be 0 (scrolled to top)
	if lv.TopIndex() != 0 {
		t.Errorf("spec 7.4 FALSIFICATION: Ctrl+PgUp: TopIndex()=%d, want 0 (absolute start scrolls to top)", lv.TopIndex())
	}
}

// TestPlainPgUpWithoutCtrlStillWorksAsPageNavigation proves the Ctrl modifier
// differentiates Ctrl+PgUp from plain PgUp: without Ctrl, PgUp is still page-based
// (moves by visibleHeight), not absolute.
// Spec 7.4: plain PgUp is unchanged; Ctrl modifier activates absolute navigation.
func TestPlainPgUpWithoutCtrlStillWorksAsPageNavigation(t *testing.T) {
	items := make10Items()
	lv := newLVFocused(items)
	lv.SetSelected(7) // selected=7, topIndex=3

	// Plain PgUp (no Ctrl modifier) — should move by height=5, not jump to 0.
	ev := listKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	// 7 - 5 = 2
	if lv.Selected() == 0 {
		t.Error("spec 7.4: plain PgUp (no Ctrl) behaved like Ctrl+PgUp (jumped to 0); Ctrl modifier must be required for absolute navigation")
	}
	if lv.Selected() != 2 {
		t.Errorf("spec 7.4: plain PgUp: Selected()=%d, want 2 (7 - visibleHeight 5 = 2)", lv.Selected())
	}
}

// ---------------------------------------------------------------------------
// Shared test helpers (local to this file)
// ---------------------------------------------------------------------------

// make10Items returns a 10-element string slice for use in Home/End/Ctrl tests.
// Bounds are 20x5 so visibleHeight=5.
func make10Items() []string {
	return []string{"item0", "item1", "item2", "item3", "item4",
		"item5", "item6", "item7", "item8", "item9"}
}
