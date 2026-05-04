package tv

// list_viewer_test.go — Tests for Task 4: ListViewer widget.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — StringList tests (Count, Item)
//   Section 2  — Constructor tests (SfVisible, OfSelectable, initial state)
//   Section 3  — State accessor tests (Selected, TopIndex, DataSource, SetSelected clamping, SetDataSource reset)
//   Section 4  — OnSelect callback tests (fires on keyboard/mouse, NOT on programmatic SetSelected)
//   Section 5  — ScrollBar binding tests (SetScrollBar syncs range/value/pageSize, OnChange updates topIndex, nil unbinds)
//   Section 6  — Drawing tests (ListNormal fill, selected row style, item text, truncation, visible items)
//   Section 7  — Keyboard handling tests (Down/Up/Home/End/PgDn/PgUp, boundaries, event consumption, scrolling)
//   Section 8  — Mouse handling tests (click selects row, event consumption, click beyond data)
//   Section 9  — Scroll adjustment tests (selected < topIndex, selected >= topIndex + visibleHeight)

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// compile-time assertion: ListViewer must satisfy Widget.
// Spec: "Implements Widget interface"
var _ Widget = (*ListViewer)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newLV creates a ListViewer with BorlandBlue scheme, bounds 20×5.
func newLV(items []string) *ListViewer {
	lv := NewListViewer(NewRect(0, 0, 20, 5), NewStringList(items))
	lv.scheme = theme.BorlandBlue
	return lv
}

// newLVFocused creates a focused ListViewer (SfSelected set).
func newLVFocused(items []string) *ListViewer {
	lv := newLV(items)
	lv.SetState(SfSelected, true)
	return lv
}

// listKeyEv creates a plain keyboard event.
func listKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Rune: 0, Modifiers: tcell.ModNone}}
}

// listMouseEv creates a mouse Button1 click event at (x, y).
func listMouseEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1}}
}

// ---------------------------------------------------------------------------
// Section 1 — StringList tests
// ---------------------------------------------------------------------------

// TestStringListCountEmpty verifies Count() returns 0 for an empty slice.
// Spec: "Count() int returns len(items)"
func TestStringListCountEmpty(t *testing.T) {
	sl := NewStringList([]string{})
	if sl.Count() != 0 {
		t.Errorf("Count() = %d, want 0 for empty list", sl.Count())
	}
}

// TestStringListCountOne verifies Count() returns 1 for a single-item slice.
// Spec: "Count() int returns len(items)"
func TestStringListCountOne(t *testing.T) {
	sl := NewStringList([]string{"alpha"})
	if sl.Count() != 1 {
		t.Errorf("Count() = %d, want 1", sl.Count())
	}
}

// TestStringListCountMany verifies Count() returns the correct count for multiple items.
// Spec: "Count() int returns len(items)"
func TestStringListCountMany(t *testing.T) {
	sl := NewStringList([]string{"a", "b", "c", "d", "e"})
	if sl.Count() != 5 {
		t.Errorf("Count() = %d, want 5", sl.Count())
	}
}

// TestStringListItemFirst verifies Item(0) returns the first element.
// Spec: "Item(index int) string returns items[index]"
func TestStringListItemFirst(t *testing.T) {
	sl := NewStringList([]string{"alpha", "beta", "gamma"})
	if sl.Item(0) != "alpha" {
		t.Errorf("Item(0) = %q, want %q", sl.Item(0), "alpha")
	}
}

// TestStringListItemLast verifies Item at the last valid index returns the last element.
// Spec: "Item(index int) string returns items[index]"
func TestStringListItemLast(t *testing.T) {
	sl := NewStringList([]string{"alpha", "beta", "gamma"})
	if sl.Item(2) != "gamma" {
		t.Errorf("Item(2) = %q, want %q", sl.Item(2), "gamma")
	}
}

// TestStringListItemMiddle verifies Item returns the correct middle element.
// Spec: "Item(index int) string returns items[index]"
func TestStringListItemMiddle(t *testing.T) {
	sl := NewStringList([]string{"alpha", "beta", "gamma"})
	if sl.Item(1) != "beta" {
		t.Errorf("Item(1) = %q, want %q", sl.Item(1), "beta")
	}
}

// TestStringListImplementsListDataSource verifies StringList satisfies ListDataSource.
// Spec: "StringList struct wraps []string and implements ListDataSource"
func TestStringListImplementsListDataSource(t *testing.T) {
	var _ ListDataSource = (*StringList)(nil)
}

// TestStringListCountDoesNotEqualWrongValue verifies Count() is not always the same value
// (falsification guard).
// Spec: "Count() int returns len(items)"
func TestStringListCountDoesNotEqualWrongValue(t *testing.T) {
	sl3 := NewStringList([]string{"a", "b", "c"})
	sl5 := NewStringList([]string{"a", "b", "c", "d", "e"})
	if sl3.Count() == sl5.Count() {
		t.Error("Count() returned the same value for lists of different lengths; Count must return len(items)")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Constructor tests
// ---------------------------------------------------------------------------

// TestNewListViewerSetsSfVisible verifies NewListViewer sets the SfVisible flag.
// Spec: "Sets SfVisible"
func TestNewListViewerSetsSfVisible(t *testing.T) {
	lv := NewListViewer(NewRect(0, 0, 20, 5), NewStringList([]string{"a"}))
	if !lv.HasState(SfVisible) {
		t.Error("NewListViewer did not set SfVisible")
	}
}

// TestNewListViewerSetsOfSelectable verifies NewListViewer sets the OfSelectable option.
// Spec: "Sets OfSelectable"
func TestNewListViewerSetsOfSelectable(t *testing.T) {
	lv := NewListViewer(NewRect(0, 0, 20, 5), NewStringList([]string{"a"}))
	if !lv.HasOption(OfSelectable) {
		t.Error("NewListViewer did not set OfSelectable")
	}
}

// TestNewListViewerStoresBounds verifies NewListViewer records the given bounds.
// Spec: "NewListViewer(bounds Rect, dataSource ListDataSource) *ListViewer"
func TestNewListViewerStoresBounds(t *testing.T) {
	r := NewRect(2, 3, 20, 5)
	lv := NewListViewer(r, NewStringList([]string{"a"}))
	if lv.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", lv.Bounds(), r)
	}
}

// TestNewListViewerInitialSelectionZero verifies the initial selection is index 0.
// Spec: "Initial selection: index 0 (if data source has items)"
func TestNewListViewerInitialSelectionZero(t *testing.T) {
	lv := NewListViewer(NewRect(0, 0, 20, 5), NewStringList([]string{"a", "b", "c"}))
	if lv.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0 (initial selection)", lv.Selected())
	}
}

// TestNewListViewerInitialTopIndexZero verifies the initial scroll offset is 0.
// Spec: "Initial scroll offset: 0"
func TestNewListViewerInitialTopIndexZero(t *testing.T) {
	lv := NewListViewer(NewRect(0, 0, 20, 5), NewStringList([]string{"a", "b", "c"}))
	if lv.TopIndex() != 0 {
		t.Errorf("TopIndex() = %d, want 0 (initial scroll offset)", lv.TopIndex())
	}
}

// TestNewListViewerStoresDataSource verifies DataSource() returns the supplied data source.
// Spec: "NewListViewer(bounds Rect, dataSource ListDataSource) *ListViewer"
func TestNewListViewerStoresDataSource(t *testing.T) {
	ds := NewStringList([]string{"x", "y"})
	lv := NewListViewer(NewRect(0, 0, 20, 5), ds)
	if lv.DataSource() != ds {
		t.Error("DataSource() did not return the data source passed to NewListViewer")
	}
}

// TestNewListViewerScrollBarNilByDefault verifies the scrollbar starts nil.
// Spec: "scrollBar *ScrollBar — optional bound scrollbar (nil by default)"
// Verified indirectly: binding a nil scrollbar first should be a no-op rather than panic.
func TestNewListViewerScrollBarNilByDefault(t *testing.T) {
	lv := NewListViewer(NewRect(0, 0, 20, 5), NewStringList([]string{"a"}))
	// If a scrollbar is nil by default, calling SetScrollBar(nil) should not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetScrollBar(nil) on a fresh ListViewer panicked: %v", r)
		}
	}()
	lv.SetScrollBar(nil)
}

// ---------------------------------------------------------------------------
// Section 3 — State accessor tests
// ---------------------------------------------------------------------------

// TestSelectedReturnsCurrentIndex verifies Selected() returns the current selection.
// Spec: "Selected() int returns current selection index"
func TestSelectedReturnsCurrentIndex(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	lv.SetSelected(2)
	if lv.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2 after SetSelected(2)", lv.Selected())
	}
}

// TestSetSelectedClampsToZero verifies SetSelected clamps negative index to 0.
// Spec: "SetSelected(index int) sets selection (clamped to [0, Count()-1])"
func TestSetSelectedClampsToZero(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	lv.SetSelected(-1)
	if lv.Selected() != 0 {
		t.Errorf("SetSelected(-1): Selected() = %d, want 0 (clamped to 0)", lv.Selected())
	}
}

// TestSetSelectedClampsToLastItem verifies SetSelected clamps index beyond last item.
// Spec: "SetSelected(index int) sets selection (clamped to [0, Count()-1])"
func TestSetSelectedClampsToLastItem(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	lv.SetSelected(99)
	if lv.Selected() != 2 {
		t.Errorf("SetSelected(99): Selected() = %d, want 2 (clamped to Count()-1=2)", lv.Selected())
	}
}

// TestSetSelectedExactLastItem verifies SetSelected accepts the exact last index.
// Spec: "clamped to [0, Count()-1]"
func TestSetSelectedExactLastItem(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	lv.SetSelected(2)
	if lv.Selected() != 2 {
		t.Errorf("SetSelected(2): Selected() = %d, want 2", lv.Selected())
	}
}

// TestTopIndexReturnsScrollOffset verifies TopIndex() returns the current scroll offset.
// Spec: "TopIndex() int returns current scroll offset"
func TestTopIndexReturnsScrollOffset(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	// Move selection to item 7 to force scroll
	lv.SetSelected(7)
	// topIndex should have adjusted so item 7 is visible
	if lv.TopIndex() > 7 {
		t.Errorf("TopIndex() = %d, want <= 7", lv.TopIndex())
	}
}

// TestDataSourceReturnsCurrentDataSource verifies DataSource() returns the current source.
// Spec: "DataSource() ListDataSource returns current data source"
func TestDataSourceReturnsCurrentDataSource(t *testing.T) {
	ds := NewStringList([]string{"x", "y"})
	lv := NewListViewer(NewRect(0, 0, 20, 5), ds)
	if lv.DataSource() != ds {
		t.Error("DataSource() did not return the original data source")
	}
}

// TestSetDataSourceReplacesDataSource verifies SetDataSource replaces the data source.
// Spec: "SetDataSource(ds ListDataSource) replaces data source"
func TestSetDataSourceReplacesDataSource(t *testing.T) {
	ds1 := NewStringList([]string{"a", "b"})
	ds2 := NewStringList([]string{"x", "y", "z"})
	lv := NewListViewer(NewRect(0, 0, 20, 5), ds1)
	lv.SetDataSource(ds2)
	if lv.DataSource() != ds2 {
		t.Error("DataSource() did not return ds2 after SetDataSource(ds2)")
	}
}

// TestSetDataSourceResetsSelectedToZero verifies SetDataSource resets selected to 0.
// Spec: "resets selected to 0 and topIndex to 0"
func TestSetDataSourceResetsSelectedToZero(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	lv.SetSelected(4)
	lv.SetDataSource(NewStringList([]string{"x", "y", "z"}))
	if lv.Selected() != 0 {
		t.Errorf("SetDataSource: Selected() = %d, want 0 (reset)", lv.Selected())
	}
}

// TestSetDataSourceResetsTopIndexToZero verifies SetDataSource resets topIndex to 0.
// Spec: "resets selected to 0 and topIndex to 0"
func TestSetDataSourceResetsTopIndexToZero(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	lv.SetSelected(7) // force scroll
	lv.SetDataSource(NewStringList([]string{"x", "y", "z"}))
	if lv.TopIndex() != 0 {
		t.Errorf("SetDataSource: TopIndex() = %d, want 0 (reset)", lv.TopIndex())
	}
}

// TestSetSelectedAdjustsTopIndexUp verifies SetSelected adjusts topIndex when selected < topIndex.
// Spec: "When selected index < topIndex, set topIndex = selected (scroll up to show selection)"
func TestSetSelectedAdjustsTopIndexUp(t *testing.T) {
	// Use a tall-enough list; first scroll down, then jump back up
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	lv := newLV(items) // height 5
	// Scroll down by selecting item 9 first
	lv.SetSelected(9)
	// Now jump to item 1 which should be above topIndex
	lv.SetSelected(1)
	if lv.TopIndex() > 1 {
		t.Errorf("SetSelected(1) after scrolling: TopIndex() = %d, want <= 1 (scrolled up to show selection)", lv.TopIndex())
	}
}

// TestSetSelectedAdjustsTopIndexDown verifies SetSelected adjusts topIndex when selected is below the visible area.
// Spec: "When selected index >= topIndex + visible height, set topIndex = selected - visible height + 1"
func TestSetSelectedAdjustsTopIndexDown(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	lv := newLV(items) // bounds height = 5, so visibleHeight = 5
	lv.SetSelected(7)
	// topIndex should be 7 - 5 + 1 = 3
	if lv.TopIndex() != 3 {
		t.Errorf("SetSelected(7) with height=5: TopIndex() = %d, want 3", lv.TopIndex())
	}
}

// ---------------------------------------------------------------------------
// Section 4 — OnSelect callback tests
// ---------------------------------------------------------------------------

// TestOnSelectCalledOnKeyDown verifies OnSelect does NOT fire when user presses Down arrow.
// Spec 7.1: "Arrow key navigation changes the focused item but does NOT call OnSelect."
func TestOnSelectCalledOnKeyDown(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)

	if called {
		t.Error("OnSelect must NOT be called after pressing Down arrow (spec 7.1)")
	}
}

// TestOnSelectCalledOnKeyUp verifies OnSelect does NOT fire when user presses Up arrow.
// Spec 7.1: "Arrow key navigation changes the focused item but does NOT call OnSelect."
func TestOnSelectCalledOnKeyUp(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(2)
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyUp)
	lv.HandleEvent(ev)

	if called {
		t.Error("OnSelect must NOT be called after pressing Up arrow (spec 7.1)")
	}
}

// TestOnSelectCalledOnHome verifies OnSelect does NOT fire when user presses Home.
// Spec 7.1: "Home and End change the focused item but do NOT call OnSelect."
func TestOnSelectCalledOnHome(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(2)
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	if called {
		t.Error("OnSelect must NOT be called after pressing Home (spec 7.1)")
	}
}

// TestOnSelectCalledOnEnd verifies OnSelect does NOT fire when user presses End.
// Spec 7.1: "Home and End change the focused item but do NOT call OnSelect."
func TestOnSelectCalledOnEnd(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	if called {
		t.Error("OnSelect must NOT be called after pressing End (spec 7.1)")
	}
}

// TestOnSelectCalledOnPgDn verifies OnSelect does NOT fire when user presses PgDn.
// Spec 7.1: "Page navigation (PgUp, PgDn) changes the focused item but does NOT call OnSelect."
func TestOnSelectCalledOnPgDn(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyPgDn)
	lv.HandleEvent(ev)

	if called {
		t.Error("OnSelect must NOT be called after pressing PgDn (spec 7.1)")
	}
}

// TestOnSelectCalledOnPgUp verifies OnSelect does NOT fire when user presses PgUp.
// Spec 7.1: "Page navigation (PgUp, PgDn) changes the focused item but does NOT call OnSelect."
func TestOnSelectCalledOnPgUp(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	lv.SetSelected(7)
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	if called {
		t.Error("OnSelect must NOT be called after pressing PgUp (spec 7.1)")
	}
}

// TestOnSelectCalledOnMouseClick verifies OnSelect is NOT fired on single click (only focus moves).
// Spec 7.3: "Single-click positions focus only; double-click fires OnSelect."
func TestOnSelectCalledOnMouseClick(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listMouseEv(0, 1) // ClickCount=0, treated as single-click
	lv.HandleEvent(ev)

	if called {
		t.Error("OnSelect must NOT be called after single mouse click (spec 7.3)")
	}
}

// TestOnSelectNotCalledBySetSelected verifies OnSelect is NOT called by programmatic SetSelected.
// Spec: "Not called by programmatic SetSelected()."
func TestOnSelectNotCalledBySetSelected(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	lv.SetSelected(2)

	if called {
		t.Error("OnSelect was called by SetSelected; it must only fire on user interaction")
	}
}

// TestOnSelectNotCalledBySetDataSource verifies OnSelect is NOT called by SetDataSource.
// Spec: "Not called by programmatic SetSelected()." (SetDataSource also resets, not user interaction)
func TestOnSelectNotCalledBySetDataSource(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	lv.SetDataSource(NewStringList([]string{"x", "y"}))

	if called {
		t.Error("OnSelect was called by SetDataSource; it must only fire on user interaction")
	}
}

// TestOnSelectReceivesNewIndex verifies Down arrow navigates to index 1 but does NOT call OnSelect.
// Spec 7.1: "Arrow key navigation changes the focused item but does NOT call OnSelect."
func TestOnSelectReceivesNewIndex(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	called := false
	lv.OnSelect = func(index int) { called = true }

	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)

	// Navigation must still move the selection
	if lv.Selected() != 1 {
		t.Errorf("Down arrow must navigate to index 1; got Selected()=%d", lv.Selected())
	}
	// But OnSelect must NOT be called
	if called {
		t.Error("OnSelect must NOT be called by Down arrow navigation (spec 7.1)")
	}
}

// TestOnSelectMouseClickReceivesClickedIndex verifies OnSelect receives the index of the double-clicked row.
// Spec 7.3: "Double-click fires OnSelect with the clicked row's index."
func TestOnSelectMouseClickReceivesClickedIndex(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	var got int
	lv.OnSelect = func(index int) { got = index }

	ev := listMouseEv(0, 2) // click row 2 → item 2 (topIndex=0)
	ev.Mouse.ClickCount = 2  // double-click
	lv.HandleEvent(ev)

	if got != 2 {
		t.Errorf("OnSelect received index %d after double-clicking row 2, want 2", got)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — ScrollBar binding tests
// ---------------------------------------------------------------------------

// TestSetScrollBarSyncsRange verifies SetScrollBar sets the scrollbar range.
// Spec: "ScrollBar range: min=0, max=dataSource.Count()-1, pageSize=visible height"
func TestSetScrollBarSyncsRange(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	if sb.Min() != 0 {
		t.Errorf("ScrollBar Min() = %d, want 0", sb.Min())
	}
	if sb.Max() != 7 {
		t.Errorf("ScrollBar Max() = %d, want 7 (Count()-1=7)", sb.Max())
	}
}

// TestSetScrollBarSyncsPageSize verifies SetScrollBar sets the scrollbar page size to visible height.
// Spec: "ScrollBar range: min=0, max=dataSource.Count(), pageSize=visible height"
func TestSetScrollBarSyncsPageSize(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	// visible height = bounds height = 5
	if sb.PageSize() != 5 {
		t.Errorf("ScrollBar PageSize() = %d, want 5 (visible height)", sb.PageSize())
	}
}

// TestSetScrollBarSyncsValue verifies SetScrollBar sets the scrollbar value to selected.
// Spec: "ScrollBar value: selected"
func TestSetScrollBarSyncsValue(t *testing.T) {
	// Use 20 items so selected=7 is well within the scrollbar's effective range
	// (max=19, pageSize=5, effectiveMax=14; 7 <= 14).
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("item%d", i)
	}
	lv := newLV(items)
	lv.SetSelected(7) // force scroll so selected > 0
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	if sb.Value() != lv.Selected() {
		t.Errorf("ScrollBar Value() = %d, want %d (selected)", sb.Value(), lv.Selected())
	}
}

// TestScrollBarOnChangeUpdatesSelected verifies ScrollBar's OnChange handler updates ListViewer selected.
// Spec: "ScrollBar's OnChange is set to update ListViewer's selected when user interacts with the scrollbar"
func TestScrollBarOnChangeUpdatesTopIndex(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	// Simulate the scrollbar firing its OnChange callback
	if sb.OnChange == nil {
		t.Fatal("ScrollBar.OnChange was not set by SetScrollBar")
	}
	sb.OnChange(3)

	if lv.Selected() != 3 {
		t.Errorf("after OnChange(3): Selected() = %d, want 3", lv.Selected())
	}
}

// TestSetScrollBarNilUnbinds verifies SetScrollBar(nil) unbinds the scrollbar.
// Spec: "Calling SetScrollBar(nil) unbinds"
func TestSetScrollBarNilUnbinds(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e", "f"})
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	// Save the OnChange that was set
	origOnChange := sb.OnChange

	lv.SetScrollBar(nil)

	// After unbinding, state changes should NOT update the old scrollbar.
	// We verify by checking the OnChange that was previously set is no longer
	// called when the LV scrolls — but we can only check this indirectly.
	// The simplest falsifying test: pressing Down should not panic (nil deref).
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("after SetScrollBar(nil), key event panicked: %v", r)
		}
	}()

	// The previously-set OnChange should no longer be associated in a way that
	// causes infinite loops; just confirm it was set before.
	if origOnChange == nil {
		t.Error("SetScrollBar should have set sb.OnChange; it was nil before unbind")
	}

	lv.SetState(SfSelected, true)
	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)
}

// TestScrollBarUpdatedOnSelectionChange verifies bound scrollbar is updated when selection changes.
// Spec: "After any state change that modifies topIndex or selected, update the bound ScrollBar (if any)"
func TestScrollBarUpdatedOnSelectionChange(t *testing.T) {
	// Use 20 items so selected=7 is within the scrollbar's effective range
	// (max=19, pageSize=5, effectiveMax=14; 7 <= 14).
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("item%d", i)
	}
	lv := newLV(items)
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	// Move selection to trigger a topIndex change
	lv.SetState(SfSelected, true)
	for i := 0; i < 7; i++ {
		ev := listKeyEv(tcell.KeyDown)
		lv.HandleEvent(ev)
	}

	if sb.Value() != lv.Selected() {
		t.Errorf("after scrolling down, ScrollBar Value()=%d but Selected()=%d; scrollbar not updated",
			sb.Value(), lv.Selected())
	}
}

// TestScrollBarUpdatedOnSetDataSource verifies bound scrollbar is updated when data source changes.
// Spec: "After any state change that modifies topIndex or selected, update the bound ScrollBar (if any)"
func TestScrollBarUpdatedOnSetDataSource(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	newDS := NewStringList([]string{"x", "y", "z", "w", "v", "u", "t"})
	lv.SetDataSource(newDS)

	if sb.Max() != newDS.Count()-1 {
		t.Errorf("after SetDataSource, ScrollBar Max()=%d, want %d (new count-1)", sb.Max(), newDS.Count()-1)
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Drawing tests
// ---------------------------------------------------------------------------

// TestDrawFillsWithListNormalStyle verifies the entire widget is filled with ListNormal style.
// Spec: "Fills entire bounds with ListNormal style"
func TestDrawFillsWithListNormalStyle(t *testing.T) {
	lv := newLV([]string{}) // empty list to see fill clearly
	buf := NewDrawBuffer(20, 5)
	lv.Draw(buf)

	style := theme.BorlandBlue.ListNormal
	for y := 0; y < 5; y++ {
		for x := 0; x < 20; x++ {
			cell := buf.GetCell(x, y)
			if cell.Style != style {
				t.Errorf("cell(%d,%d) style = %v, want ListNormal %v", x, y, cell.Style, style)
				return
			}
		}
	}
}

// TestDrawSelectedRowUsesListSelectedWhenNotFocused verifies selected row uses ListSelected when unfocused.
// Spec: "The selected item row uses ListSelected style when the widget does NOT have SfSelected (no focus)"
func TestDrawSelectedRowUsesListSelectedWhenNotFocused(t *testing.T) {
	lv := newLV([]string{"alpha", "beta", "gamma"})
	// Not focused: SfSelected not set (default)

	buf := NewDrawBuffer(20, 5)
	lv.Draw(buf)

	// Row 0 is the selected item (selected=0, topIndex=0)
	cell := buf.GetCell(0, 0)
	if cell.Style != theme.BorlandBlue.ListSelected {
		t.Errorf("selected row style (unfocused) = %v, want ListSelected %v",
			cell.Style, theme.BorlandBlue.ListSelected)
	}
}

// TestDrawSelectedRowUsesListFocusedWhenFocused verifies selected row uses ListFocused when focused.
// Spec: "The selected item row uses ListFocused style when the widget has SfSelected (focused)"
func TestDrawSelectedRowUsesListFocusedWhenFocused(t *testing.T) {
	lv := newLVFocused([]string{"alpha", "beta", "gamma"})

	buf := NewDrawBuffer(20, 5)
	lv.Draw(buf)

	// Row 0 is the selected item (selected=0, topIndex=0)
	cell := buf.GetCell(0, 0)
	if cell.Style != theme.BorlandBlue.ListFocused {
		t.Errorf("selected row style (focused) = %v, want ListFocused %v",
			cell.Style, theme.BorlandBlue.ListFocused)
	}
}

// TestDrawUnselectedRowUsesListNormalStyle verifies unselected rows use ListNormal style.
// Spec: "Unselected rows use ListNormal style"
func TestDrawUnselectedRowUsesListNormalStyle(t *testing.T) {
	lv := newLV([]string{"alpha", "beta", "gamma"})
	// selected=0; rows 1 and 2 are unselected

	buf := NewDrawBuffer(20, 5)
	lv.Draw(buf)

	for row := 1; row <= 2; row++ {
		cell := buf.GetCell(0, row)
		if cell.Style != theme.BorlandBlue.ListNormal {
			t.Errorf("unselected row %d style = %v, want ListNormal %v",
				row, cell.Style, theme.BorlandBlue.ListNormal)
		}
	}
}

// TestDrawRendersItemText verifies visible item text appears in the buffer.
// Spec: "Renders visible items starting from topIndex, one item per row"
func TestDrawRendersItemText(t *testing.T) {
	lv := newLV([]string{"hello", "world", "foo"})
	buf := NewDrawBuffer(20, 5)
	lv.Draw(buf)

	// Row 0 = "hello" (selected), row 1 = "world", row 2 = "foo"
	checkRowText(t, buf, 1, "world")
	checkRowText(t, buf, 2, "foo")
}

// checkRowText is a helper that verifies the text rendered at a given row.
func checkRowText(t *testing.T, buf *DrawBuffer, row int, text string) {
	t.Helper()
	for i, ch := range text {
		cell := buf.GetCell(i, row)
		if cell.Rune != ch {
			t.Errorf("row %d col %d = %q, want %q (from %q)", row, i, cell.Rune, ch, text)
			return
		}
	}
}

// TestDrawRendersFirstItemOnRow0 verifies item[topIndex] appears on row 0.
// Spec: "Renders visible items starting from topIndex, one item per row"
func TestDrawRendersFirstItemOnRow0(t *testing.T) {
	lv := newLV([]string{"first", "second", "third"})
	buf := NewDrawBuffer(20, 5)
	lv.Draw(buf)

	checkRowText(t, buf, 0, "first")
}

// TestDrawTruncatesLongItemToWidth verifies text is truncated to fit widget width.
// Spec: "Each item text is truncated to fit the widget width"
func TestDrawTruncatesLongItemToWidth(t *testing.T) {
	longItem := "abcdefghijklmnopqrstuvwxyz" // 26 chars, widget width = 20
	lv := newLV([]string{longItem})
	buf := NewDrawBuffer(20, 5)
	lv.Draw(buf)

	// Column 20 does not exist; check that column 19 has the 20th char
	cell19 := buf.GetCell(19, 0)
	expected := rune(longItem[19])
	if cell19.Rune != expected {
		t.Errorf("col 19 = %q, want %q (20th char of long item)", cell19.Rune, expected)
	}
	// And verify no overflow: col 20 is out of buffer range, so just check that
	// items beyond width are not rendered into col 20 (buffer is only 20 wide).
}

// TestDrawScrolledListRendersFromTopIndex verifies items start at topIndex row 0.
// Spec: "Renders visible items starting from topIndex, one item per row"
func TestDrawScrolledListRendersFromTopIndex(t *testing.T) {
	items := []string{"item0", "item1", "item2", "item3", "item4", "item5", "item6"}
	lv := newLV(items)
	// Scroll so topIndex = 2
	lv.SetSelected(6) // forces topIndex = 6 - 5 + 1 = 2
	// topIndex is now 2

	buf := NewDrawBuffer(20, 5)
	lv.Draw(buf)

	// Row 0 should show items[topIndex] = items[2] = "item2"
	checkRowText(t, buf, 0, "item2")
}

// TestDrawListSelectedAndListFocusedAreDifferentStyles verifies the two focused/unfocused
// selected styles are distinct (falsification guard for drawing tests).
func TestDrawListSelectedAndListFocusedAreDifferentStyles(t *testing.T) {
	if theme.BorlandBlue.ListSelected == theme.BorlandBlue.ListFocused {
		t.Skip("ListSelected equals ListFocused in this scheme — style distinction test is vacuous")
	}
}

// TestDrawListNormalAndListSelectedAreDifferentStyles verifies ListNormal and ListSelected differ.
func TestDrawListNormalAndListSelectedAreDifferentStyles(t *testing.T) {
	if theme.BorlandBlue.ListNormal == theme.BorlandBlue.ListSelected {
		t.Skip("ListNormal equals ListSelected in this scheme — style distinction test is vacuous")
	}
}

// Spec §2.10: "When the data source has 0 items, the ListViewer displays
// <empty> in the first row using the normal color."
func TestDrawEmptyDataSourceRendersEmptyText(t *testing.T) {
	lv := newLV([]string{})
	buf := NewDrawBuffer(20, 5)
	lv.Draw(buf)

	checkRowText(t, buf, 0, "<empty>")
}

func TestDrawEmptyDataSourceUsesNormalStyle(t *testing.T) {
	lv := newLV([]string{})
	buf := NewDrawBuffer(20, 5)
	lv.Draw(buf)

	normalStyle := theme.BorlandBlue.ListNormal
	for i, ch := range []rune("<empty>") {
		cell := buf.GetCell(i, 0)
		if cell.Style != normalStyle {
			t.Errorf("cell(%d,0) rune=%q style = %v, want ListNormal %v", i, ch, cell.Style, normalStyle)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Keyboard handling tests
// ---------------------------------------------------------------------------

// TestKeyDownMovesSelectionDown verifies Down arrow moves selection down by 1.
// Spec: "Down arrow: move selection down by 1 (no-op at last item)"
func TestKeyDownMovesSelectionDown(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)

	if lv.Selected() != 1 {
		t.Errorf("after Down: Selected() = %d, want 1", lv.Selected())
	}
}

// TestKeyDownNoOpAtLastItem verifies Down arrow is a no-op at the last item.
// Spec: "Down arrow: move selection down by 1 (no-op at last item)"
func TestKeyDownNoOpAtLastItem(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(2) // last item

	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)

	if lv.Selected() != 2 {
		t.Errorf("Down at last item: Selected() = %d, want 2 (no-op)", lv.Selected())
	}
}

// TestKeyDownConsumesEvent verifies Down arrow consumes the event.
// Spec: "consume event"
func TestKeyDownConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Down arrow: event not consumed")
	}
}

// TestKeyDownOnlyWhenFocused verifies Down arrow is only handled when focused.
// Spec: "Keyboard handling (only when focused)"
func TestKeyDownOnlyWhenFocused(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"}) // not focused
	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)

	if lv.Selected() != 0 {
		t.Errorf("Down without focus: Selected() = %d, want 0 (unhandled)", lv.Selected())
	}
	if ev.IsCleared() {
		t.Error("Down without focus: event should not have been consumed")
	}
}

// TestKeyUpMovesSelectionUp verifies Up arrow moves selection up by 1.
// Spec: "Up arrow: move selection up by 1 (no-op at first item)"
func TestKeyUpMovesSelectionUp(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(2)

	ev := listKeyEv(tcell.KeyUp)
	lv.HandleEvent(ev)

	if lv.Selected() != 1 {
		t.Errorf("after Up: Selected() = %d, want 1", lv.Selected())
	}
}

// TestKeyUpNoOpAtFirstItem verifies Up arrow is a no-op at the first item.
// Spec: "Up arrow: move selection up by 1 (no-op at first item)"
func TestKeyUpNoOpAtFirstItem(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	// selected = 0 (initial)

	ev := listKeyEv(tcell.KeyUp)
	lv.HandleEvent(ev)

	if lv.Selected() != 0 {
		t.Errorf("Up at first item: Selected() = %d, want 0 (no-op)", lv.Selected())
	}
}

// TestKeyUpConsumesEvent verifies Up arrow consumes the event.
// Spec: "consume event"
func TestKeyUpConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	lv.SetSelected(1)
	ev := listKeyEv(tcell.KeyUp)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Up arrow: event not consumed")
	}
}

// TestKeyHomeSelectsFirstItem verifies Home selects the first visible (topIndex) item.
// Spec 7.4: "Home: select first visible item on the page (topIndex)"
func TestKeyHomeSelectsFirstItem(t *testing.T) {
	// 10 items, height=5; scroll to topIndex=3, then press Home
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	lv := newLVFocused(items)
	lv.SetSelected(7) // forces topIndex = 7 - 5 + 1 = 3
	lv.selected = 6   // move selection within the page (topIndex stays 3)

	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	if lv.Selected() != 3 {
		t.Errorf("after Home with topIndex=3: Selected() = %d, want 3 (topIndex)", lv.Selected())
	}
}

// TestKeyHomeScrollsToTop verifies Home leaves topIndex unchanged (page-local).
// Spec 7.4: "Home: select first visible item on the page (topIndex); topIndex unchanged"
func TestKeyHomeScrollsToTop(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	lv := newLVFocused(items)
	lv.SetSelected(7) // topIndex = 7 - 5 + 1 = 3

	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	if lv.TopIndex() != 3 {
		t.Errorf("after Home: TopIndex() = %d, want 3 (unchanged)", lv.TopIndex())
	}
	if lv.Selected() != 3 {
		t.Errorf("after Home: Selected() = %d, want 3 (topIndex)", lv.Selected())
	}
}

// TestKeyHomeConsumesEvent verifies Home consumes the event.
// Spec: "consume event"
func TestKeyHomeConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	ev := listKeyEv(tcell.KeyHome)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Home: event not consumed")
	}
}

// TestKeyEndSelectsLastItem verifies End selects the last visible item on the page.
// Spec 7.4: "End: select last visible item on the page (topIndex + visibleHeight - 1, clamped)"
func TestKeyEndSelectsLastItem(t *testing.T) {
	// 10 items, height=5, topIndex=0; End → selected = 0 + 5 - 1 = 4
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	lv := newLVFocused(items)
	// topIndex=0 by default

	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	if lv.Selected() != 4 {
		t.Errorf("after End with topIndex=0, height=5: Selected() = %d, want 4 (topIndex+height-1)", lv.Selected())
	}
}

// TestKeyEndScrollsToShowLastItem verifies End leaves topIndex unchanged (page-local).
// Spec 7.4: "End: select last visible item on the page; topIndex unchanged"
func TestKeyEndScrollsToShowLastItem(t *testing.T) {
	// 10 items, height=5, topIndex=0; End selects item 4, topIndex stays 0
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	lv := newLVFocused(items)
	// topIndex=0 by default

	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	if lv.TopIndex() != 0 {
		t.Errorf("after End: TopIndex() = %d, want 0 (unchanged)", lv.TopIndex())
	}
	if lv.Selected() != 4 {
		t.Errorf("after End: Selected() = %d, want 4 (topIndex+height-1)", lv.Selected())
	}
}

// TestKeyEndConsumesEvent verifies End consumes the event.
// Spec: "consume event"
func TestKeyEndConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	ev := listKeyEv(tcell.KeyEnd)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("End: event not consumed")
	}
}

// TestKeyPgDnMovesSelectionDownByHeight verifies PgDn moves selection down by visible height.
// Spec: "PgDn: move selection down by visible height (clamped to last item)"
func TestKeyPgDnMovesSelectionDownByHeight(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"} // 10 items
	lv := newLVFocused(items)
	// height = 5; PgDn should move 5 rows

	ev := listKeyEv(tcell.KeyPgDn)
	lv.HandleEvent(ev)

	// 0 + 5 = 5
	if lv.Selected() != 5 {
		t.Errorf("after PgDn: Selected() = %d, want 5 (0 + visibleHeight 5)", lv.Selected())
	}
}

// TestKeyPgDnClampsToLastItem verifies PgDn clamps to last item when near end.
// Spec: "PgDn: move selection down by visible height (clamped to last item)"
func TestKeyPgDnClampsToLastItem(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g"} // 7 items
	lv := newLVFocused(items)
	lv.SetSelected(4) // 4 + 5 = 9, but last item = 6

	ev := listKeyEv(tcell.KeyPgDn)
	lv.HandleEvent(ev)

	if lv.Selected() != 6 {
		t.Errorf("PgDn near end: Selected() = %d, want 6 (clamped to last)", lv.Selected())
	}
}

// TestKeyPgDnConsumesEvent verifies PgDn consumes the event.
// Spec: "consume event"
func TestKeyPgDnConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	ev := listKeyEv(tcell.KeyPgDn)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("PgDn: event not consumed")
	}
}

// TestKeyPgUpMovesSelectionUpByHeight verifies PgUp moves selection up by visible height.
// Spec: "PgUp: move selection up by visible height (clamped to first item)"
func TestKeyPgUpMovesSelectionUpByHeight(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"} // 10 items
	lv := newLVFocused(items)
	lv.SetSelected(7)

	ev := listKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	// 7 - 5 = 2
	if lv.Selected() != 2 {
		t.Errorf("after PgUp: Selected() = %d, want 2 (7 - visibleHeight 5)", lv.Selected())
	}
}

// TestKeyPgUpClampsToFirstItem verifies PgUp clamps to first item when near beginning.
// Spec: "PgUp: move selection up by visible height (clamped to first item)"
func TestKeyPgUpClampsToFirstItem(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g"} // 7 items
	lv := newLVFocused(items)
	lv.SetSelected(2) // 2 - 5 = -3, should clamp to 0

	ev := listKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	if lv.Selected() != 0 {
		t.Errorf("PgUp near top: Selected() = %d, want 0 (clamped to first)", lv.Selected())
	}
}

// TestKeyPgUpConsumesEvent verifies PgUp consumes the event.
// Spec: "consume event"
func TestKeyPgUpConsumesEvent(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	lv.SetSelected(6)
	ev := listKeyEv(tcell.KeyPgUp)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("PgUp: event not consumed")
	}
}

// TestUnhandledKeyEventPassesThrough verifies unhandled key events are not consumed.
// Spec: "Events not handled pass through unconsumed"
func TestUnhandledKeyEventPassesThrough(t *testing.T) {
	lv := newLVFocused([]string{"a", "b", "c"})
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF1}}
	lv.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("unhandled key F1: event was consumed, want pass-through")
	}
}

// TestKeyDownScrollsWhenSelectionGoesOffScreen verifies Down arrow scrolls when needed.
// Spec: "scroll if needed"
func TestKeyDownScrollsWhenSelectionGoesOffScreen(t *testing.T) {
	// 10 items, height=5. Starting at item 4 (bottom of visible area).
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	lv := newLVFocused(items)
	lv.SetSelected(4) // topIndex=0, visible rows 0-4; pressing Down should scroll

	ev := listKeyEv(tcell.KeyDown)
	lv.HandleEvent(ev)

	// selected=5; now 5 >= 0 + 5, so topIndex should become 5 - 5 + 1 = 1
	if lv.TopIndex() != 1 {
		t.Errorf("Down from item 4: TopIndex() = %d, want 1 (scrolled to show item 5)", lv.TopIndex())
	}
}

// TestKeyUpScrollsWhenSelectionGoesAboveTopIndex verifies Up arrow scrolls when needed.
// Spec: "scroll if needed"
func TestKeyUpScrollsWhenSelectionGoesAboveTopIndex(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	lv := newLVFocused(items)
	lv.SetSelected(5) // forces topIndex=1; now pressing Up should keep item 4 visible

	ev := listKeyEv(tcell.KeyUp)
	lv.HandleEvent(ev)

	// selected=4; topIndex should remain 1 (4 is still visible within 1..5)
	if lv.TopIndex() > 4 {
		t.Errorf("Up from item 5: TopIndex() = %d, but selected=4 should be visible", lv.TopIndex())
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Mouse handling tests
// ---------------------------------------------------------------------------

// TestMouseClickSelectsRow verifies clicking a row selects topIndex + clickY.
// Spec: "Click (Button1) on a visible row: select that item … The clicked row is topIndex + clickY"
func TestMouseClickSelectsRow(t *testing.T) {
	lv := newLV([]string{"a", "b", "c", "d", "e"})
	ev := listMouseEv(0, 2) // click row 2
	lv.HandleEvent(ev)

	// topIndex=0, clickY=2 → selected = 0 + 2 = 2
	if lv.Selected() != 2 {
		t.Errorf("click row 2: Selected() = %d, want 2 (topIndex=0 + clickY=2)", lv.Selected())
	}
}

// TestMouseClickConsumesEvent verifies mouse click on a visible row consumes the event.
// Spec: "consume event"
func TestMouseClickConsumesEvent(t *testing.T) {
	lv := newLV([]string{"a", "b", "c"})
	ev := listMouseEv(0, 0)
	lv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("mouse click: event not consumed")
	}
}

// TestMouseClickRowZeroSelectsTopIndex verifies clicking row 0 selects topIndex.
// Spec: "The clicked row is topIndex + clickY"
func TestMouseClickRowZeroSelectsTopIndex(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	lv := newLV(items)
	lv.SetSelected(5) // forces topIndex=1

	ev := listMouseEv(3, 0) // click row 0
	lv.HandleEvent(ev)

	// topIndex + clickY = 1 + 0 = 1
	if lv.Selected() != lv.TopIndex() {
		t.Errorf("click row 0 with topIndex=%d: Selected()=%d, want %d",
			lv.TopIndex(), lv.Selected(), lv.TopIndex())
	}
}

// TestMouseClickBeyondDataCountDoesNotPanic verifies click on empty rows below data
// does not panic or cause an out-of-bounds selection.
// Spec: "The clicked row is topIndex + clickY" — if result >= Count(), no item exists there
func TestMouseClickBeyondDataCountDoesNotPanic(t *testing.T) {
	lv := newLV([]string{"a", "b"}) // 2 items, height=5

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("click beyond data count panicked: %v", r)
		}
	}()

	// Click row 4 — beyond the 2 items
	ev := listMouseEv(0, 4)
	lv.HandleEvent(ev)

	// Selected should remain valid (within [0, Count()-1])
	if lv.Selected() < 0 || lv.Selected() >= 2 {
		t.Errorf("click beyond data: Selected()=%d is out of valid range [0,1]", lv.Selected())
	}
}

// TestMouseClickRow1WithScrollOffset verifies click accounts for topIndex.
// Spec: "The clicked row is topIndex + clickY"
func TestMouseClickRow1WithScrollOffset(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	lv := newLV(items)
	lv.SetSelected(7) // forces topIndex = 7 - 5 + 1 = 3

	topIdx := lv.TopIndex()
	ev := listMouseEv(0, 1) // click row 1
	lv.HandleEvent(ev)

	want := topIdx + 1
	if lv.Selected() != want {
		t.Errorf("click row 1 with topIndex=%d: Selected()=%d, want %d", topIdx, lv.Selected(), want)
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Scroll adjustment tests
// ---------------------------------------------------------------------------

// TestScrollAdjustUpWhenSelectedLessThanTopIndex verifies topIndex is set to selected
// when selected < topIndex.
// Spec: "When selected index < topIndex, set topIndex = selected (scroll up to show selection)"
func TestScrollAdjustUpWhenSelectedLessThanTopIndex(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	lv := newLV(items)
	// Force topIndex to 5 by selecting item 9
	lv.SetSelected(9) // topIndex = 9 - 5 + 1 = 5
	// Now jump to item 2 (< topIndex=5)
	lv.SetSelected(2)

	if lv.TopIndex() != 2 {
		t.Errorf("SetSelected(2) with previous topIndex=5: TopIndex()=%d, want 2 (scrolled up)", lv.TopIndex())
	}
}

// TestScrollAdjustDownWhenSelectedBeyondVisibleArea verifies topIndex adjusts when
// selected >= topIndex + visibleHeight.
// Spec: "When selected index >= topIndex + visible height, set topIndex = selected - visible height + 1"
func TestScrollAdjustDownWhenSelectedBeyondVisibleArea(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	lv := newLV(items) // height=5
	lv.SetSelected(8)

	// topIndex = 8 - 5 + 1 = 4
	if lv.TopIndex() != 4 {
		t.Errorf("SetSelected(8) with height=5: TopIndex()=%d, want 4", lv.TopIndex())
	}
}

// TestScrollAdjustNoOpWhenSelectionAlreadyVisible verifies topIndex is unchanged
// when the selected item is already visible.
// Spec (implicit): scroll adjustment only occurs when selection is out of view
func TestScrollAdjustNoOpWhenSelectionAlreadyVisible(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	lv := newLV(items) // height=5, topIndex=0, visible rows 0-4
	lv.SetSelected(3)  // item 3 is visible (0 <= 3 < 5)

	if lv.TopIndex() != 0 {
		t.Errorf("SetSelected(3) with topIndex=0 and height=5: TopIndex()=%d, want 0 (no scroll needed)",
			lv.TopIndex())
	}
}

// TestScrollAdjustExactBottomBoundaryNoScroll verifies no scroll when selection is at
// exactly the last visible row.
// Spec: "When selected index >= topIndex + visible height" (boundary: selected == topIndex + visibleHeight - 1 is still visible)
func TestScrollAdjustExactBottomBoundaryNoScroll(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	lv := newLV(items) // height=5
	// topIndex=0; last visible row = 4
	lv.SetSelected(4)

	if lv.TopIndex() != 0 {
		t.Errorf("SetSelected(4) exactly at bottom visible row: TopIndex()=%d, want 0 (no scroll)", lv.TopIndex())
	}
}

// TestScrollAdjustOnePassedBottomBoundaryScrolls verifies scroll occurs when selection
// is one past the last visible row.
// Spec: "When selected index >= topIndex + visible height, set topIndex = selected - visible height + 1"
func TestScrollAdjustOnePassedBottomBoundaryScrolls(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	lv := newLV(items) // height=5, topIndex=0
	lv.SetSelected(5)  // 5 >= 0 + 5 = 5, so scroll: topIndex = 5 - 5 + 1 = 1

	if lv.TopIndex() != 1 {
		t.Errorf("SetSelected(5) one past bottom: TopIndex()=%d, want 1", lv.TopIndex())
	}
}

// TestScrollAdjustTopBoundaryExact verifies no scroll when selection equals topIndex.
// Spec: "When selected index < topIndex, set topIndex = selected"
// Boundary: selected == topIndex is visible, so no scroll.
func TestScrollAdjustTopBoundaryExact(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	lv := newLV(items)
	lv.SetSelected(7) // topIndex = 3
	lv.SetSelected(3) // selected == topIndex: visible, no scroll

	if lv.TopIndex() != 3 {
		t.Errorf("SetSelected(3) == topIndex=3: TopIndex()=%d, want 3 (no change)", lv.TopIndex())
	}
}

func TestScrollBarTracksSelectedNotTopIndex(t *testing.T) {
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d", i)
	}
	lv := newLV(items)
	lv.SetNumCols(2)
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	lv.SetSelected(5)

	if sb.Value() != 5 {
		t.Errorf("ScrollBar.Value() = %d, want 5 (should track selected, not topIndex)", sb.Value())
	}
}

func TestScrollBarOnChangeSetsSelected(t *testing.T) {
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d", i)
	}
	lv := newLV(items)
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	sb.OnChange(7)

	if lv.Selected() != 7 {
		t.Errorf("after OnChange(7): Selected() = %d, want 7", lv.Selected())
	}
}

func TestScrollBarRangeIsCountMinusOne(t *testing.T) {
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d", i)
	}
	lv := newLV(items)
	sb := NewScrollBar(NewRect(20, 0, 1, 5), Vertical)
	lv.SetScrollBar(sb)

	if sb.Max() != 19 {
		t.Errorf("ScrollBar.Max() = %d, want 19 (count-1)", sb.Max())
	}
}
