package tv

// integration_phase13_batch1_test.go — Integration tests for Phase 13 Tasks 1–3:
// Navigation vs Selection distinction in ListViewer.
//
// Each test exercises multiple features across the navigation/selection pipeline
// to verify that the distinction between browsing (navigation keys, single-click)
// and confirming (Space, Enter, double-click) works correctly end-to-end.
//
// Features under test:
//   Task 1: Navigation keys (Up/Down/PgUp/PgDn/Home/End) do NOT fire OnSelect
//   Task 2: Space and Enter fire OnSelect with the correct index
//   Task 3: Single-click positions focus only; double-click fires OnSelect
//
// Test naming: TestIntegrationPhase13<DescriptiveSuffix>

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Scenario 1: Multi-step arrow navigation does NOT fire OnSelect at any point.
//
// Verifies that pressing Down multiple times in sequence never triggers OnSelect
// even as the selection moves through several items.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13MultiStepArrowNavNeverFiresOnSelect verifies that
// navigating through a list with Down arrows (0→1→2→3→4) never calls OnSelect
// at any intermediate step.
//
// This guards against an implementation that fires OnSelect on each move.
func TestIntegrationPhase13MultiStepArrowNavNeverFiresOnSelect(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e"})

	callCount := 0
	lv.OnSelect = func(index int) { callCount++ }

	// Navigate down 4 times: 0 → 1 → 2 → 3 → 4.
	for i := 0; i < 4; i++ {
		ev := listKeyEv(tcell.KeyDown)
		lv.HandleEvent(ev)
	}

	// Navigation must have worked.
	if lv.Selected() != 4 {
		t.Fatalf("prerequisite: after 4 Down presses, Selected()=%d, want 4", lv.Selected())
	}

	// OnSelect must not have fired at any point during navigation.
	if callCount != 0 {
		t.Errorf("task 1: OnSelect fired %d time(s) during arrow key navigation, want 0", callCount)
	}
}

// TestIntegrationPhase13MixedNavKeysNeverFiresOnSelect verifies that a sequence
// using different navigation keys (Down, PgDn, Up, Home, End) never fires OnSelect.
func TestIntegrationPhase13MixedNavKeysNeverFiresOnSelect(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"})

	callCount := 0
	lv.OnSelect = func(index int) { callCount++ }

	// Sequence of different navigation keys.
	navKeys := []tcell.Key{
		tcell.KeyDown,
		tcell.KeyDown,
		tcell.KeyPgDn,
		tcell.KeyUp,
		tcell.KeyHome,
		tcell.KeyEnd,
		tcell.KeyPgUp,
	}
	for _, key := range navKeys {
		ev := listKeyEv(key)
		lv.HandleEvent(ev)
	}

	if callCount != 0 {
		t.Errorf("task 1: OnSelect fired %d time(s) during mixed nav key sequence, want 0", callCount)
	}
}

// ---------------------------------------------------------------------------
// Scenario 2: Navigate to an item then press Space — fires OnSelect with correct index.
//
// Verifies the full navigate→confirm pipeline for Space.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13NavigateDownThenSpaceFiresOnSelectAtCorrectIndex verifies
// that after pressing Down three times (0→3), Space fires OnSelect with index 3.
func TestIntegrationPhase13NavigateDownThenSpaceFiresOnSelectAtCorrectIndex(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e"})

	// Navigate down 3 times without triggering OnSelect.
	for i := 0; i < 3; i++ {
		lv.HandleEvent(listKeyEv(tcell.KeyDown))
	}

	if lv.Selected() != 3 {
		t.Fatalf("prerequisite: after 3 Down presses, Selected()=%d, want 3", lv.Selected())
	}

	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	spaceEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	lv.HandleEvent(spaceEv)

	if got != 3 {
		t.Errorf("task 2: Space after navigating to index 3 must fire OnSelect(3); got %d", got)
	}
}

// TestIntegrationPhase13NavigateWithPgDnThenSpaceFiresOnSelectAtCorrectIndex verifies
// that after PgDn (from 0 to 5 in a 10-item list with height=5), Space fires with 5.
func TestIntegrationPhase13NavigateWithPgDnThenSpaceFiresOnSelectAtCorrectIndex(t *testing.T) {
	// newLV uses height=5; PgDn from 0 moves to min(0+5, count-1) = 5.
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"})

	lv.HandleEvent(listKeyEv(tcell.KeyPgDn))

	if lv.Selected() != 5 {
		t.Fatalf("prerequisite: after PgDn from 0, Selected()=%d, want 5", lv.Selected())
	}

	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	spaceEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	lv.HandleEvent(spaceEv)

	if got != 5 {
		t.Errorf("task 2: Space after PgDn to index 5 must fire OnSelect(5); got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Scenario 3: Navigate to an item then press Enter — fires OnSelect with correct index.
//
// Verifies the full navigate→confirm pipeline for Enter.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13NavigateDownThenEnterFiresOnSelectAtCorrectIndex verifies
// that after pressing Down twice (0→2), Enter fires OnSelect with index 2.
func TestIntegrationPhase13NavigateDownThenEnterFiresOnSelectAtCorrectIndex(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e"})

	for i := 0; i < 2; i++ {
		lv.HandleEvent(listKeyEv(tcell.KeyDown))
	}

	if lv.Selected() != 2 {
		t.Fatalf("prerequisite: after 2 Down presses, Selected()=%d, want 2", lv.Selected())
	}

	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	lv.HandleEvent(listKeyEv(tcell.KeyEnter))

	if got != 2 {
		t.Errorf("task 2: Enter after navigating to index 2 must fire OnSelect(2); got %d", got)
	}
}

// TestIntegrationPhase13NavigateToEndThenEnterFiresWithLastIndex verifies that
// pressing End followed by Enter fires OnSelect with the last item index.
func TestIntegrationPhase13NavigateToEndThenEnterFiresWithLastIndex(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	lv := newLVFocused(items)
	lastIdx := len(items) - 1 // 4

	lv.HandleEvent(listKeyEv(tcell.KeyEnd))

	if lv.Selected() != lastIdx {
		t.Fatalf("prerequisite: after End, Selected()=%d, want %d", lv.Selected(), lastIdx)
	}

	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	lv.HandleEvent(listKeyEv(tcell.KeyEnter))

	if got != lastIdx {
		t.Errorf("task 2: Enter after End must fire OnSelect(%d); got %d", lastIdx, got)
	}
}

// ---------------------------------------------------------------------------
// Scenario 4: Single-click positions focus, then Space fires OnSelect for that item.
//
// Verifies the click→focus + space→select pipeline.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13SingleClickThenSpaceFiresOnSelectForClickedItem verifies that
// a single click on row 2 positions selected to 2 (without OnSelect), then Space fires
// OnSelect with index 2.
func TestIntegrationPhase13SingleClickThenSpaceFiresOnSelectForClickedItem(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e"})
	lv.SetState(SfSelected, true)

	// Track OnSelect calls during the click phase.
	clickOnSelectCalled := false
	lv.OnSelect = func(index int) { clickOnSelectCalled = true }

	// Single click (ClickCount=1) on row 2.
	singleClickEv := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 2, Button: tcell.Button1, ClickCount: 1}}
	lv.HandleEvent(singleClickEv)

	// OnSelect must NOT have fired during the click.
	if clickOnSelectCalled {
		t.Error("task 3: OnSelect must NOT fire on single click (click should only position focus)")
	}

	// Selected must now be 2.
	if lv.Selected() != 2 {
		t.Fatalf("task 3: single click row 2: Selected()=%d, want 2 (focus positioned)", lv.Selected())
	}

	// Now set up OnSelect to capture the Space press.
	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	spaceEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	lv.HandleEvent(spaceEv)

	if got != 2 {
		t.Errorf("task 2+3: Space after single-click on row 2 must fire OnSelect(2); got %d", got)
	}
}

// TestIntegrationPhase13SingleClickRow1ThenSpaceFiresWithIndex1 verifies the same
// pipeline for row 1: click row 1 → Space → OnSelect(1).
func TestIntegrationPhase13SingleClickRow1ThenSpaceFiresWithIndex1(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e"})

	// Single click on row 1.
	singleClickEv := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1, ClickCount: 1}}
	lv.HandleEvent(singleClickEv)

	if lv.Selected() != 1 {
		t.Fatalf("prerequisite: single click row 1: Selected()=%d, want 1", lv.Selected())
	}

	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	spaceEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	lv.HandleEvent(spaceEv)

	if got != 1 {
		t.Errorf("task 2+3: Space after single-click row 1 must fire OnSelect(1); got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Scenario 5: Double-click fires OnSelect exactly once (not twice).
//
// Verifies that the double-click event results in exactly one OnSelect call,
// not two — guarding against implementations that fire once for "click" and
// once for "double-click confirm".
// ---------------------------------------------------------------------------

// TestIntegrationPhase13DoubleClickFiresOnSelectExactlyOnce verifies that a
// ClickCount=2 event results in exactly one OnSelect call.
func TestIntegrationPhase13DoubleClickFiresOnSelectExactlyOnce(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})

	callCount := 0
	lv.OnSelect = func(index int) { callCount++ }

	doubleClickEv := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1, ClickCount: 2}}
	lv.HandleEvent(doubleClickEv)

	if callCount != 1 {
		t.Errorf("task 3: double-click must fire OnSelect exactly once; called %d time(s)", callCount)
	}
}

// TestIntegrationPhase13DoubleClickFiresWithCorrectIndex verifies the double-click
// passes the correct row index to OnSelect.
func TestIntegrationPhase13DoubleClickFiresWithCorrectIndex(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})

	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	// topIndex=0; double-click row 3 → OnSelect(3).
	doubleClickEv := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 3, Button: tcell.Button1, ClickCount: 2}}
	lv.HandleEvent(doubleClickEv)

	if got != 3 {
		t.Errorf("task 3: double-click row 3 must fire OnSelect(3); got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Scenario 6: PgDn navigation then double-click fires OnSelect for the clicked item.
//
// Verifies that after scrolling via PgDn (which changes topIndex), a double-click
// on a visible row fires OnSelect for the row's absolute index (topIndex + clickY),
// not for the previously focused item.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13PgDnThenDoubleClickFiresOnSelectForClickedItem verifies
// that after PgDn navigation shifts topIndex, a double-click on row 0 fires
// OnSelect with the absolute index (topIndex + 0), not with the pre-PgDn selection.
func TestIntegrationPhase13PgDnThenDoubleClickFiresOnSelectForClickedItem(t *testing.T) {
	// 10 items, height=5. PgDn from 0 → selected=5, topIndex=1.
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"})

	lv.HandleEvent(listKeyEv(tcell.KeyPgDn))

	topIdx := lv.TopIndex()
	prevSelected := lv.Selected()
	if prevSelected == 0 {
		t.Fatal("prerequisite: PgDn should have moved selection away from 0")
	}

	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	// Double-click row 0 → absolute index = topIndex + 0.
	doubleClickEv := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1, ClickCount: 2}}
	lv.HandleEvent(doubleClickEv)

	wantIdx := topIdx + 0
	if got != wantIdx {
		t.Errorf("task 3: double-click row 0 with topIndex=%d must fire OnSelect(%d); got %d (was prevSelected=%d)",
			topIdx, wantIdx, got, prevSelected)
	}
}

// TestIntegrationPhase13PgDnThenDoubleClickRow2UsesTopIndexOffset verifies that
// after PgDn, double-clicking row 2 uses topIndex+2 as the OnSelect argument.
func TestIntegrationPhase13PgDnThenDoubleClickRow2UsesTopIndexOffset(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"})

	lv.HandleEvent(listKeyEv(tcell.KeyPgDn))

	topIdx := lv.TopIndex()

	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	doubleClickEv := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 2, Button: tcell.Button1, ClickCount: 2}}
	lv.HandleEvent(doubleClickEv)

	wantIdx := topIdx + 2
	if got != wantIdx {
		t.Errorf("task 3: double-click row 2 with topIndex=%d must fire OnSelect(%d); got %d",
			topIdx, wantIdx, got)
	}
}

// ---------------------------------------------------------------------------
// Scenario 7: Full sequence — navigate down 3 times → single-click item 1 → Space.
//
// Verifies that the full pipeline:
//   1. Arrow navigation does not fire OnSelect.
//   2. Single-click repositions selection (no OnSelect).
//   3. Space fires OnSelect with the single-click's index (1), not the nav index (3).
// ---------------------------------------------------------------------------

// TestIntegrationPhase13FullSequenceNavThenClickThenSpaceFiresForClickedItem verifies
// the complete sequence: navigate down 3 times → single-click item 1 → Space.
//
// Expected: OnSelect is called with index 1 (the clicked item), not 3 (the nav target).
func TestIntegrationPhase13FullSequenceNavThenClickThenSpaceFiresForClickedItem(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e"})

	onSelectCallsDuringNav := 0
	lv.OnSelect = func(index int) { onSelectCallsDuringNav++ }

	// Step 1: Navigate down 3 times. OnSelect must NOT fire.
	for i := 0; i < 3; i++ {
		lv.HandleEvent(listKeyEv(tcell.KeyDown))
	}

	if lv.Selected() != 3 {
		t.Fatalf("prerequisite: after 3 Down presses, Selected()=%d, want 3", lv.Selected())
	}
	if onSelectCallsDuringNav != 0 {
		t.Errorf("task 1: OnSelect fired %d time(s) during arrow navigation, want 0", onSelectCallsDuringNav)
	}

	// Step 2: Single-click on row 1. OnSelect must NOT fire, but selected must become 1.
	onSelectCallsDuringClick := 0
	lv.OnSelect = func(index int) { onSelectCallsDuringClick++ }

	singleClickEv := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1, ClickCount: 1}}
	lv.HandleEvent(singleClickEv)

	if lv.Selected() != 1 {
		t.Fatalf("task 3: single-click row 1: Selected()=%d, want 1", lv.Selected())
	}
	if onSelectCallsDuringClick != 0 {
		t.Errorf("task 3: OnSelect fired %d time(s) during single click, want 0", onSelectCallsDuringClick)
	}

	// Step 3: Press Space. OnSelect must fire with index 1 (the clicked item).
	var got int = -1
	lv.OnSelect = func(index int) { got = index }

	spaceEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	lv.HandleEvent(spaceEv)

	if got != 1 {
		t.Errorf("task 2+3: Space after full nav→click sequence must fire OnSelect(1); got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Scenario 8: ScrollBar sync is maintained throughout navigation and selection.
//
// Verifies that navigation and double-click both keep the bound ScrollBar in sync
// with the ListViewer's topIndex after the full pipeline.
// ---------------------------------------------------------------------------

// TestIntegrationPhase13NavigationKeepsScrollBarInSync verifies that arrow navigation
// that causes scrolling keeps the bound ScrollBar value equal to topIndex.
func TestIntegrationPhase13NavigationKeepsScrollBarInSync(t *testing.T) {
	items := make([]string, 15)
	for i := range items {
		items[i] = "x"
	}
	lv := newLVFocused(items) // height=5
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	// Navigate down 8 times — this will force scrolling.
	for i := 0; i < 8; i++ {
		lv.HandleEvent(listKeyEv(tcell.KeyDown))
	}

	if sb.Value() != lv.Selected() {
		t.Errorf("task 1+scrollbar: after navigation, ScrollBar.Value()=%d but Selected()=%d; must be equal",
			sb.Value(), lv.Selected())
	}
}

// TestIntegrationPhase13DoubleClickKeepsScrollBarInSync verifies that a double-click
// on a scrolled list does not break ScrollBar synchronisation.
func TestIntegrationPhase13DoubleClickKeepsScrollBarInSync(t *testing.T) {
	// Use 20 items so selected values stay within the scrollbar's effective range
	// (max=19, pageSize=5, effectiveMax=14).
	items := make([]string, 20)
	for i := range items {
		items[i] = "x"
	}
	lv := newLV(items) // height=5
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	// Scroll so topIndex=5 by selecting item 9.
	lv.SetSelected(9) // topIndex = 9 - 5 + 1 = 5

	topIdx := lv.TopIndex()
	if topIdx != 5 {
		t.Fatalf("prerequisite: TopIndex()=%d, want 5 after SetSelected(9)", topIdx)
	}

	lv.OnSelect = func(index int) {}
	// Double-click at y=2 → selected = topIndex + 2 = 5 + 2 = 7
	doubleClickEv := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 2, Button: tcell.Button1, ClickCount: 2}}
	lv.HandleEvent(doubleClickEv)

	// selected=7, which is within effectiveMax=14, so scrollbar value = selected.
	if sb.Value() != lv.Selected() {
		t.Errorf("task 3+scrollbar: after double-click, ScrollBar.Value()=%d but Selected()=%d; must be equal",
			sb.Value(), lv.Selected())
	}
}
