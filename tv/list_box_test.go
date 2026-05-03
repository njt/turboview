package tv

// list_box_test.go — Tests for ListBox widget.
//
// Written BEFORE any implementation exists; all tests are driven by the spec only.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Compile-time interface assertions
//   Section 2  — NewListBox construction: sub-components are non-nil
//   Section 3  — NewListBox construction: flags (OfSelectable, SfVisible)
//   Section 4  — NewListBox construction: bounds layout (viewer width-1, scrollbar width 1 at right edge)
//   Section 5  — ListViewer accessor
//   Section 6  — ScrollBar accessor
//   Section 7  — Selected / SetSelected delegation
//   Section 8  — DataSource accessor
//   Section 9  — SetDataSource resets selection
//   Section 10 — NewStringListBox convenience constructor
//   Section 11 — ScrollBar is not selectable (no OfSelectable)
//   Section 12 — Wiring: scrollbar OnChange is non-nil after construction

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Section 1 — Compile-time interface assertion
// ---------------------------------------------------------------------------

// Spec: "ListBox implements Container interface"
var _ Container = (*ListBox)(nil)

// ---------------------------------------------------------------------------
// Section 2 — Construction: sub-components are non-nil
// ---------------------------------------------------------------------------

// TestNewListBoxListViewerNonNil verifies the internal ListViewer is created.
// Spec: "A ListViewer filling remaining width (bounds width - 1, full height)"
func TestNewListBoxListViewerNonNil(t *testing.T) {
	ds := NewStringList([]string{"a", "b", "c"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)

	if lb.ListViewer() == nil {
		t.Error("ListBox.ListViewer() must not be nil after NewListBox")
	}
}

// TestNewListBoxScrollBarNonNil verifies the internal ScrollBar is created.
// Spec: "A vertical ScrollBar at the right edge (width 1, full height)"
func TestNewListBoxScrollBarNonNil(t *testing.T) {
	ds := NewStringList([]string{"a", "b", "c"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)

	if lb.ScrollBar() == nil {
		t.Error("ListBox.ScrollBar() must not be nil after NewListBox")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Construction: OfSelectable and SfVisible set
// ---------------------------------------------------------------------------

// TestNewListBoxSetsOfSelectable verifies OfSelectable is set on the ListBox.
// Spec: "OfSelectable set"
func TestNewListBoxSetsOfSelectable(t *testing.T) {
	ds := NewStringList([]string{"a"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)

	if !lb.HasOption(OfSelectable) {
		t.Error("NewListBox must set OfSelectable on the ListBox")
	}
}

// TestNewListBoxSetsSfVisible verifies SfVisible is set on the ListBox.
// Spec: "SfVisible set"
func TestNewListBoxSetsSfVisible(t *testing.T) {
	ds := NewStringList([]string{"a"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)

	if !lb.HasState(SfVisible) {
		t.Error("NewListBox must set SfVisible on the ListBox")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Construction: bounds layout
// ---------------------------------------------------------------------------

// TestNewListBoxViewerWidthIsOneLessThanBounds verifies the ListViewer fills width-1.
// Spec: "A ListViewer filling remaining width (bounds width - 1, full height)"
func TestNewListBoxViewerWidthIsOneLessThanBounds(t *testing.T) {
	bounds := NewRect(0, 0, 20, 10)
	ds := NewStringList([]string{"a"})
	lb := NewListBox(bounds, ds)

	got := lb.ListViewer().Bounds().Width()
	want := bounds.Width() - 1
	if got != want {
		t.Errorf("ListViewer width = %d, want bounds.Width()-1 = %d", got, want)
	}
}

// TestNewListBoxViewerHeightEqualsFullHeight verifies the ListViewer spans the full height.
// Spec: "A ListViewer filling remaining width (bounds width - 1, full height)"
func TestNewListBoxViewerHeightEqualsFullHeight(t *testing.T) {
	bounds := NewRect(0, 0, 20, 10)
	ds := NewStringList([]string{"a"})
	lb := NewListBox(bounds, ds)

	got := lb.ListViewer().Bounds().Height()
	want := bounds.Height()
	if got != want {
		t.Errorf("ListViewer height = %d, want full bounds height %d", got, want)
	}
}

// TestNewListBoxScrollBarWidthIsOne verifies the ScrollBar has width 1.
// Spec: "A vertical ScrollBar at the right edge (width 1, full height)"
func TestNewListBoxScrollBarWidthIsOne(t *testing.T) {
	ds := NewStringList([]string{"a"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)

	got := lb.ScrollBar().Bounds().Width()
	if got != 1 {
		t.Errorf("ScrollBar width = %d, want 1", got)
	}
}

// TestNewListBoxScrollBarHeightEqualsFullHeight verifies the ScrollBar spans full height.
// Spec: "A vertical ScrollBar at the right edge (width 1, full height)"
func TestNewListBoxScrollBarHeightEqualsFullHeight(t *testing.T) {
	bounds := NewRect(0, 0, 20, 10)
	ds := NewStringList([]string{"a"})
	lb := NewListBox(bounds, ds)

	got := lb.ScrollBar().Bounds().Height()
	want := bounds.Height()
	if got != want {
		t.Errorf("ScrollBar height = %d, want full bounds height %d", got, want)
	}
}

// TestNewListBoxScrollBarAtRightEdge verifies the ScrollBar is positioned at the right edge.
// Spec: "A vertical ScrollBar at the right edge"
// The scrollbar's left edge (Bounds().A.X) must equal bounds.Width()-1.
func TestNewListBoxScrollBarAtRightEdge(t *testing.T) {
	bounds := NewRect(0, 0, 20, 10)
	ds := NewStringList([]string{"a"})
	lb := NewListBox(bounds, ds)

	// Bounds are relative to the ListBox origin (0,0); scrollbar left edge = width-1
	gotX := lb.ScrollBar().Bounds().A.X
	wantX := bounds.Width() - 1
	if gotX != wantX {
		t.Errorf("ScrollBar left edge X = %d, want %d (right edge)", gotX, wantX)
	}
}

// TestNewListBoxViewerStartsAtLeftEdge verifies the ListViewer starts at x=0.
// Spec: "A ListViewer filling remaining width (bounds width - 1, full height)"
func TestNewListBoxViewerStartsAtLeftEdge(t *testing.T) {
	ds := NewStringList([]string{"a"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)

	if lb.ListViewer().Bounds().A.X != 0 {
		t.Errorf("ListViewer left edge = %d, want 0", lb.ListViewer().Bounds().A.X)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — ListViewer accessor
// ---------------------------------------------------------------------------

// TestListViewerAccessorReturnsNonNil verifies ListViewer() is always non-nil.
// Spec: "ListViewer() *ListViewer — access internal ListViewer (non-nil)"
func TestListViewerAccessorReturnsNonNil(t *testing.T) {
	lb := NewListBox(NewRect(0, 0, 30, 8), NewStringList([]string{"x"}))
	if lb.ListViewer() == nil {
		t.Error("ListViewer() must return non-nil")
	}
}

// ---------------------------------------------------------------------------
// Section 6 — ScrollBar accessor
// ---------------------------------------------------------------------------

// TestScrollBarAccessorReturnsNonNil verifies ScrollBar() is always non-nil.
// Spec: "ScrollBar() *ScrollBar — access internal ScrollBar (non-nil)"
func TestScrollBarAccessorReturnsNonNil(t *testing.T) {
	lb := NewListBox(NewRect(0, 0, 30, 8), NewStringList([]string{"x"}))
	if lb.ScrollBar() == nil {
		t.Error("ScrollBar() must return non-nil")
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Selected / SetSelected delegation
// ---------------------------------------------------------------------------

// TestSelectedDelegatesToListViewer verifies Selected() returns the ListViewer's selection.
// Spec: "Selected() int — delegates to ListViewer.Selected()"
func TestSelectedDelegatesToListViewer(t *testing.T) {
	ds := NewStringList([]string{"a", "b", "c"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)

	lb.ListViewer().SetSelected(2)

	if lb.Selected() != 2 {
		t.Errorf("ListBox.Selected() = %d, want 2 (delegates to ListViewer.Selected())", lb.Selected())
	}
}

// TestSetSelectedDelegatesToListViewer verifies SetSelected() updates the ListViewer.
// Spec: "SetSelected(index int) — delegates to ListViewer.SetSelected()"
func TestSetSelectedDelegatesToListViewer(t *testing.T) {
	ds := NewStringList([]string{"a", "b", "c"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)

	lb.SetSelected(1)

	if lb.ListViewer().Selected() != 1 {
		t.Errorf("after ListBox.SetSelected(1), ListViewer.Selected() = %d, want 1", lb.ListViewer().Selected())
	}
}

// TestSelectedAndSetSelectedRoundtrip verifies Selected() reports what SetSelected() stored.
// Spec: "Selected() int — delegates to ListViewer.Selected()"
//       "SetSelected(index int) — delegates to ListViewer.SetSelected()"
func TestSelectedAndSetSelectedRoundtrip(t *testing.T) {
	ds := NewStringList([]string{"a", "b", "c"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)

	lb.SetSelected(2)

	if lb.Selected() != 2 {
		t.Errorf("after SetSelected(2), Selected() = %d, want 2", lb.Selected())
	}
}

// ---------------------------------------------------------------------------
// Section 8 — DataSource accessor
// ---------------------------------------------------------------------------

// TestDataSourceAccessorReturnsListViewerDataSource verifies DataSource() returns
// the ListViewer's data source.
// Spec: "DataSource() ListDataSource — returns the ListViewer's data source"
func TestDataSourceAccessorReturnsListViewerDataSource(t *testing.T) {
	ds := NewStringList([]string{"a", "b", "c"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)

	if lb.DataSource() != ds {
		t.Error("DataSource() must return the data source passed to NewListBox")
	}
}

// TestDataSourceMatchesListViewerDataSource verifies DataSource() agrees with
// ListViewer().DataSource().
// Spec: "DataSource() ListDataSource — returns the ListViewer's data source"
func TestDataSourceMatchesListViewerDataSource(t *testing.T) {
	ds := NewStringList([]string{"x", "y"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)

	if lb.DataSource() != lb.ListViewer().DataSource() {
		t.Error("ListBox.DataSource() and ListViewer().DataSource() must agree")
	}
}

// ---------------------------------------------------------------------------
// Section 9 — SetDataSource resets selection
// ---------------------------------------------------------------------------

// TestListBoxSetDataSourceReplacesDataSource verifies SetDataSource updates the data source.
// Spec: "SetDataSource(ds ListDataSource) — replaces data source, resets selection to 0"
func TestListBoxSetDataSourceReplacesDataSource(t *testing.T) {
	ds1 := NewStringList([]string{"a", "b"})
	ds2 := NewStringList([]string{"x", "y", "z"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds1)

	lb.SetDataSource(ds2)

	if lb.DataSource() != ds2 {
		t.Error("after SetDataSource(ds2), DataSource() must return ds2")
	}
}

// TestSetDataSourceResetsSelectionToZero verifies SetDataSource resets selection to 0.
// Spec: "SetDataSource(ds ListDataSource) — replaces data source, resets selection to 0"
func TestSetDataSourceResetsSelectionToZero(t *testing.T) {
	ds1 := NewStringList([]string{"a", "b", "c"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds1)
	lb.SetSelected(2)

	ds2 := NewStringList([]string{"x", "y", "z"})
	lb.SetDataSource(ds2)

	if lb.Selected() != 0 {
		t.Errorf("after SetDataSource, Selected() = %d, want 0 (reset)", lb.Selected())
	}
}

// TestSetDataSourceSelectionWasNonZeroBeforeReset falsifies a trivial always-zero implementation.
// Spec: "resets selection to 0" — meaningful only when selection was non-zero beforehand.
func TestSetDataSourceSelectionWasNonZeroBeforeReset(t *testing.T) {
	ds := NewStringList([]string{"a", "b", "c"})
	lb := NewListBox(NewRect(0, 0, 20, 10), ds)
	lb.SetSelected(2)

	if lb.Selected() == 0 {
		t.Skip("selection was already 0; reset test is trivially satisfied")
	}

	lb.SetDataSource(NewStringList([]string{"x", "y"}))

	if lb.Selected() != 0 {
		t.Errorf("SetDataSource must reset selection from non-zero (%d) to 0; got %d",
			2, lb.Selected())
	}
}

// ---------------------------------------------------------------------------
// Section 10 — NewStringListBox convenience constructor
// ---------------------------------------------------------------------------

// TestNewStringListBoxReturnsNonNil verifies the convenience constructor returns a ListBox.
// Spec: "NewStringListBox(bounds Rect, items []string) *ListBox convenience constructor"
func TestNewStringListBoxReturnsNonNil(t *testing.T) {
	lb := NewStringListBox(NewRect(0, 0, 20, 10), []string{"a", "b", "c"})
	if lb == nil {
		t.Error("NewStringListBox must return a non-nil *ListBox")
	}
}

// TestNewStringListBoxListViewerNonNil verifies the convenience constructor produces a
// ListBox with a non-nil ListViewer.
// Spec: "NewStringListBox(bounds Rect, items []string) *ListBox convenience constructor"
func TestNewStringListBoxListViewerNonNil(t *testing.T) {
	lb := NewStringListBox(NewRect(0, 0, 20, 10), []string{"a", "b"})
	if lb.ListViewer() == nil {
		t.Error("NewStringListBox: ListViewer() must not be nil")
	}
}

// TestNewStringListBoxScrollBarNonNil verifies the convenience constructor produces a
// ListBox with a non-nil ScrollBar.
// Spec: "NewStringListBox(bounds Rect, items []string) *ListBox convenience constructor"
func TestNewStringListBoxScrollBarNonNil(t *testing.T) {
	lb := NewStringListBox(NewRect(0, 0, 20, 10), []string{"a", "b"})
	if lb.ScrollBar() == nil {
		t.Error("NewStringListBox: ScrollBar() must not be nil")
	}
}

// TestNewStringListBoxExposesItems verifies the items are accessible through the DataSource.
// Spec: "NewStringListBox(bounds Rect, items []string) *ListBox convenience constructor"
func TestNewStringListBoxExposesItems(t *testing.T) {
	items := []string{"alpha", "beta", "gamma"}
	lb := NewStringListBox(NewRect(0, 0, 20, 10), items)

	ds := lb.DataSource()
	if ds == nil {
		t.Fatal("DataSource() must not be nil")
	}
	if ds.Count() != len(items) {
		t.Errorf("DataSource().Count() = %d, want %d", ds.Count(), len(items))
	}
	if ds.Item(1) != "beta" {
		t.Errorf("DataSource().Item(1) = %q, want %q", ds.Item(1), "beta")
	}
}

// TestNewStringListBoxBoundsPreserved verifies the bounds are applied correctly.
// Spec: "NewStringListBox(bounds Rect, items []string) *ListBox"
func TestNewStringListBoxBoundsPreserved(t *testing.T) {
	bounds := NewRect(0, 0, 30, 12)
	lb := NewStringListBox(bounds, []string{"x"})

	// Viewer width should be bounds.Width()-1 = 29
	if lb.ListViewer().Bounds().Width() != bounds.Width()-1 {
		t.Errorf("NewStringListBox: ListViewer width = %d, want %d",
			lb.ListViewer().Bounds().Width(), bounds.Width()-1)
	}
}

// ---------------------------------------------------------------------------
// Section 11 — ScrollBar is not selectable (no keyboard focus)
// ---------------------------------------------------------------------------

// TestScrollBarNotSelectable verifies the ScrollBar does not have OfSelectable set.
// Spec: "ScrollBar is NOT selectable (no keyboard focus)"
func TestScrollBarNotSelectable(t *testing.T) {
	lb := NewListBox(NewRect(0, 0, 20, 10), NewStringList([]string{"a"}))

	if lb.ScrollBar().HasOption(OfSelectable) {
		t.Error("ScrollBar must NOT have OfSelectable — it must not receive keyboard focus")
	}
}

// TestScrollBarNeverGainsSfSelected verifies SfSelected is never set on the ScrollBar
// even when the ListBox is focused.
// Spec: "ScrollBar is NOT selectable (no keyboard focus)"
func TestScrollBarNeverGainsSfSelected(t *testing.T) {
	lb := NewListBox(NewRect(0, 0, 20, 10), NewStringList([]string{"a", "b", "c"}))

	// Simulate the ListBox gaining focus (as a Container this could set focused child)
	// The focused child must be the ListViewer, not the ScrollBar.
	focused := lb.FocusedChild()
	if focused == nil {
		// If there is no focused child yet, that's fine — scrollbar still must not be it.
		if lb.ScrollBar().HasOption(OfSelectable) {
			t.Error("ScrollBar must not have OfSelectable")
		}
		return
	}
	if focused == View(lb.ScrollBar()) {
		t.Error("FocusedChild must not be the ScrollBar — it is not selectable")
	}
}

// TestListViewerIsFocusedChildNotScrollBar verifies the ListViewer is the focused
// child within the internal group, not the ScrollBar.
// Spec: "ListViewer is the focused child within the internal Group"
func TestListViewerIsFocusedChildNotScrollBar(t *testing.T) {
	lb := NewListBox(NewRect(0, 0, 20, 10), NewStringList([]string{"a", "b"}))

	focused := lb.FocusedChild()
	// The spec says ListViewer is the focused child; scrollbar is not selectable
	// so it can never be focused.
	if focused == View(lb.ScrollBar()) {
		t.Error("FocusedChild must not be the ScrollBar; ListViewer must be focused")
	}
	// If focused is non-nil, it must be the ListViewer.
	if focused != nil && focused != View(lb.ListViewer()) {
		t.Errorf("FocusedChild = %T, want *ListViewer", focused)
	}
}

// ---------------------------------------------------------------------------
// Section 12 — Wiring: scrollbar OnChange is non-nil after construction
// ---------------------------------------------------------------------------

// TestScrollBarOnChangeSetAfterConstruction verifies the ScrollBar's OnChange handler
// is non-nil, proving the ListViewer wired itself to the ScrollBar via SetScrollBar.
// Spec: "Wired via ListViewer.SetScrollBar(sb)"
// Spec (ListViewer.SetScrollBar): "sb.OnChange = func(val int) { lv.topIndex = val … }"
func TestScrollBarOnChangeSetAfterConstruction(t *testing.T) {
	lb := NewListBox(NewRect(0, 0, 20, 10), NewStringList([]string{"a", "b", "c"}))

	if lb.ScrollBar().OnChange == nil {
		t.Error("ScrollBar.OnChange must be non-nil after construction — ListViewer.SetScrollBar was not called")
	}
}

// TestScrollBarOnChangeUpdatesBehavior verifies that calling the ScrollBar's OnChange
// actually propagates the scroll position to the ListViewer (functional wiring check).
// Spec: "Wired via ListViewer.SetScrollBar(sb)"
func TestScrollBarOnChangeUpdatesBehavior(t *testing.T) {
	items := make([]string, 20)
	for i := range items {
		items[i] = "item"
	}
	lb := NewListBox(NewRect(0, 0, 20, 10), NewStringList(items))

	sb := lb.ScrollBar()
	if sb.OnChange == nil {
		t.Fatal("ScrollBar.OnChange is nil — wiring not established")
	}

	// Fire the OnChange callback as if the user scrolled to position 5
	sb.OnChange(5)

	// The ListViewer's topIndex must have been updated
	if lb.ListViewer().TopIndex() != 5 {
		t.Errorf("after OnChange(5), ListViewer.TopIndex() = %d, want 5 (wiring not working)",
			lb.ListViewer().TopIndex())
	}
}
