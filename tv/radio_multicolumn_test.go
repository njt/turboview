package tv

// radio_multicolumn_test.go — Tests for Task 3: Multi-column RadioButtons with Enable/Disable.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test verifies one named requirement.
//
// Test organisation:
//   Section 1  — Multi-column layout
//   Section 2  — SetEnabled / IsEnabled API
//   Section 3  — Enable/disable rendering
//   Section 4  — Navigation: Up/Down = row (delta ±1), Left/Right = column (delta ±height)
//   Section 5  — Navigation skips disabled items
//   Section 6  — Interaction blocked on disabled items
//   Section 7  — Selected item remains when disabled; OfSelectable management
//   Section 8  — Plain-letter shortcut matching
//   Section 9  — SetBounds relayout
//   Section 10 — Falsifying tests

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Section 1 — Multi-column layout
// ---------------------------------------------------------------------------

// TestRBMulticolumnTwoColumnsWhenHeightLessThanItemCount verifies that 3 items in a
// bounds with height=2 produce a 2-column layout.
// Spec: "numCols = ceil(len(items) / height). If bounds.Height() >= len(items), single column."
func TestRBMulticolumnTwoColumnsWhenHeightLessThanItemCount(t *testing.T) {
	// height=2, 3 items → ceil(3/2) = 2 columns.
	// col 0: items 0,1; col 1: item 2.
	rbs := NewRadioButtons(NewRect(0, 0, 40, 2), []string{"Alpha", "Beta", "Gamma"})

	// Item 0: row=0, col=0 → x=0, y=0.
	if rbs.Item(0).Bounds().A.Y != 0 {
		t.Errorf("Item(0) row = %d, want 0", rbs.Item(0).Bounds().A.Y)
	}
	if rbs.Item(0).Bounds().A.X != 0 {
		t.Errorf("Item(0) x = %d, want 0 (first column)", rbs.Item(0).Bounds().A.X)
	}

	// Item 1: row=1, col=0 → x=0, y=1.
	if rbs.Item(1).Bounds().A.Y != 1 {
		t.Errorf("Item(1) row = %d, want 1", rbs.Item(1).Bounds().A.Y)
	}
	if rbs.Item(1).Bounds().A.X != 0 {
		t.Errorf("Item(1) x = %d, want 0 (still first column)", rbs.Item(1).Bounds().A.X)
	}

	// Item 2: row=0, col=1 → x>0, y=0.
	if rbs.Item(2).Bounds().A.Y != 0 {
		t.Errorf("Item(2) row = %d, want 0 (first row, second column)", rbs.Item(2).Bounds().A.Y)
	}
	if rbs.Item(2).Bounds().A.X <= 0 {
		t.Errorf("Item(2) x = %d, want > 0 (second column)", rbs.Item(2).Bounds().A.X)
	}
}

// TestRBMulticolumnOneColumnWhenHeightEqualsItemCount verifies that height=len(items)
// produces a single-column layout (backward compatible).
// Spec: "Backward compatible: if bounds.Height() >= len(labels), single column as before."
func TestRBMulticolumnOneColumnWhenHeightEqualsItemCount(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})

	for i := 0; i < 3; i++ {
		if rbs.Item(i).Bounds().A.X != 0 {
			t.Errorf("Item(%d).x = %d, want 0 (single column when height >= items)", i, rbs.Item(i).Bounds().A.X)
		}
		if rbs.Item(i).Bounds().A.Y != i {
			t.Errorf("Item(%d).y = %d, want %d (y=index in single column)", i, rbs.Item(i).Bounds().A.Y, i)
		}
	}
}

// TestRBMulticolumnOneColumnWhenHeightExceedsItemCount verifies extra height still
// produces a single-column layout.
// Spec: "Backward compatible: if bounds.Height() >= len(labels), single column as before."
func TestRBMulticolumnOneColumnWhenHeightExceedsItemCount(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 5), []string{"Alpha", "Beta"})

	for i := 0; i < 2; i++ {
		if rbs.Item(i).Bounds().A.X != 0 {
			t.Errorf("Item(%d).x = %d, want 0 (single column when height > items)", i, rbs.Item(i).Bounds().A.X)
		}
	}
}

// TestRBMulticolumnItemRowIsItemModHeight verifies row = item % height.
// Spec: "Items fill top-to-bottom within a column."
func TestRBMulticolumnItemRowIsItemModHeight(t *testing.T) {
	// 4 items, height=2 → items 0,1 in col0; items 2,3 in col1.
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})

	expected := []int{0, 1, 0, 1}
	for i, want := range expected {
		got := rbs.Item(i).Bounds().A.Y
		if got != want {
			t.Errorf("Item(%d) row = %d, want %d (item %% height=2)", i, got, want)
		}
	}
}

// TestRBMulticolumnColumnXIsAccumulatedWidths verifies col1 starts at col0's width.
// Spec: "Column x-position = sum of widths of all previous columns."
func TestRBMulticolumnColumnXIsAccumulatedWidths(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"AB", "CD", "EF"})

	item0x := rbs.Item(0).Bounds().A.X
	item0w := rbs.Item(0).Bounds().Width()
	item2x := rbs.Item(2).Bounds().A.X

	if item0x != 0 {
		t.Errorf("Item(0) x = %d, want 0", item0x)
	}
	if item2x != item0w {
		t.Errorf("Item(2) x = %d, want %d (col1 starts at col0 width)", item2x, item0w)
	}
}

// TestRBMulticolumnColumnWidthIncludesLabelPlusSix verifies colWidth = widestLabel + 6.
// Spec: "column width = widest label + 6."
func TestRBMulticolumnColumnWidthIncludesLabelPlusSix(t *testing.T) {
	// Single label "AB" (2 visible runes) → colWidth = 2+6 = 8.
	rbs := NewRadioButtons(NewRect(0, 0, 40, 1), []string{"AB"})

	want := tildeTextLen("AB") + 6 // 2+6=8
	got := rbs.Item(0).Bounds().Width()
	if got != want {
		t.Errorf("colWidth = %d, want %d (tildeTextLen + 6)", got, want)
	}
}

// TestRBMulticolumnWidestLabelDeterminesColumnWidth verifies the column width is
// determined by the widest label in that column.
// Spec: "column width = widest label + 6."
func TestRBMulticolumnWidestLabelDeterminesColumnWidth(t *testing.T) {
	// 2 items in col0 (height=2): "A" and "LONGNAME".
	// colWidth = len("LONGNAME")+6 = 14.
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "LONGNAME", "X"})

	w0 := rbs.Item(0).Bounds().Width()
	w1 := rbs.Item(1).Bounds().Width()

	if w0 != w1 {
		t.Errorf("Items 0 and 1 in same column have different widths: %d vs %d", w0, w1)
	}

	want := tildeTextLen("LONGNAME") + 6
	if w0 != want {
		t.Errorf("col0 width = %d, want %d (widest label + 6)", w0, want)
	}
}

// TestRBMulticolumnItemBoundsHeightIsOne verifies each item's bounds height is 1.
// Spec: "Each item's Bounds are set to NewRect(colX, row, colWidth, 1)."
func TestRBMulticolumnItemBoundsHeightIsOne(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"Alpha", "Beta", "Gamma"})

	for i := 0; i < 3; i++ {
		if rbs.Item(i).Bounds().Height() != 1 {
			t.Errorf("Item(%d).Bounds().Height() = %d, want 1", i, rbs.Item(i).Bounds().Height())
		}
	}
}

// TestRBMulticolumnThreeColumnsWithSixItems verifies a 3-column layout with 6 items.
// Spec: "col(item) = item / height."
func TestRBMulticolumnThreeColumnsWithSixItems(t *testing.T) {
	// 6 items, height=2 → 3 columns: (0,1), (2,3), (4,5).
	rbs := NewRadioButtons(NewRect(0, 0, 90, 2), []string{"A", "B", "C", "D", "E", "F"})

	x0 := rbs.Item(0).Bounds().A.X
	x2 := rbs.Item(2).Bounds().A.X
	x4 := rbs.Item(4).Bounds().A.X

	if x0 >= x2 {
		t.Errorf("Item(2).x=%d should be > Item(0).x=%d (next column)", x2, x0)
	}
	if x2 >= x4 {
		t.Errorf("Item(4).x=%d should be > Item(2).x=%d (next column)", x4, x2)
	}
	// Same-column items share x.
	if rbs.Item(0).Bounds().A.X != rbs.Item(1).Bounds().A.X {
		t.Errorf("Items 0 and 1 in same column have different x: %d vs %d",
			rbs.Item(0).Bounds().A.X, rbs.Item(1).Bounds().A.X)
	}
}

// TestRBMulticolumnColumnWidthUsesTildeStrippedLength verifies that column width
// uses the tilde-stripped label length, not the raw string length.
// Spec: "column width = widest label + 6."
func TestRBMulticolumnColumnWidthUsesTildeStrippedLength(t *testing.T) {
	// "~A~BC" raw=5, stripped="ABC"=3 → colWidth = 3+6 = 9, not 5+6 = 11.
	rbs := NewRadioButtons(NewRect(0, 0, 40, 1), []string{"~A~BC"})

	want := tildeTextLen("~A~BC") + 6 // 3+6=9
	got := rbs.Item(0).Bounds().Width()
	if got != want {
		t.Errorf("colWidth for '~A~BC' = %d, want %d (must use tilde-stripped length)", got, want)
	}
}

// ---------------------------------------------------------------------------
// Section 2 — SetEnabled / IsEnabled API
// ---------------------------------------------------------------------------

// TestRBSetEnabledIsEnabledDefaultTrue verifies all items are enabled by default.
// Spec: "Uses enableMask uint32 (all enabled by default)."
func TestRBSetEnabledIsEnabledDefaultTrue(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	for i := 0; i < 3; i++ {
		if !rbs.IsEnabled(i) {
			t.Errorf("IsEnabled(%d) = false on newly created RadioButtons; must be enabled by default", i)
		}
	}
}

// TestRBSetEnabledFalseDisablesItem verifies SetEnabled(i, false) disables that item.
// Spec: "RadioButtons has SetEnabled(index int, enabled bool)."
func TestRBSetEnabledFalseDisablesItem(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	rbs.SetEnabled(1, false)

	if rbs.IsEnabled(1) {
		t.Error("after SetEnabled(1, false), IsEnabled(1) = true; expected false")
	}
}

// TestRBSetEnabledTrueReEnablesItem verifies SetEnabled(i, true) re-enables a disabled item.
// Spec: "RadioButtons has SetEnabled(index int, enabled bool)."
func TestRBSetEnabledTrueReEnablesItem(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetEnabled(0, false)

	rbs.SetEnabled(0, true)

	if !rbs.IsEnabled(0) {
		t.Error("after SetEnabled(0, true) following disable, IsEnabled(0) = false; expected true")
	}
}

// TestRBSetEnabledOnlyAffectsTargetItem verifies SetEnabled on one item does not
// affect other items.
// Spec: "RadioButtons has SetEnabled(index int, enabled bool)."
func TestRBSetEnabledOnlyAffectsTargetItem(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	rbs.SetEnabled(1, false)

	if !rbs.IsEnabled(0) {
		t.Error("SetEnabled(1, false) changed IsEnabled(0); should only affect index 1")
	}
	if !rbs.IsEnabled(2) {
		t.Error("SetEnabled(1, false) changed IsEnabled(2); should only affect index 1")
	}
}

// TestRBSetEnabledMultipleItemsIndependently verifies independent enable masks per item.
// Spec: "Uses enableMask uint32."
func TestRBSetEnabledMultipleItemsIndependently(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 4), []string{"A", "B", "C", "D"})

	rbs.SetEnabled(0, false)
	rbs.SetEnabled(2, false)

	if rbs.IsEnabled(0) {
		t.Error("IsEnabled(0) = true after SetEnabled(0, false)")
	}
	if !rbs.IsEnabled(1) {
		t.Error("IsEnabled(1) = false; should be enabled (not disabled)")
	}
	if rbs.IsEnabled(2) {
		t.Error("IsEnabled(2) = true after SetEnabled(2, false)")
	}
	if !rbs.IsEnabled(3) {
		t.Error("IsEnabled(3) = false; should be enabled (not disabled)")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Enable/disable rendering
// ---------------------------------------------------------------------------

// TestRBDisabledItemRendersWithClusterDisabledStyle verifies a disabled item uses
// ClusterDisabled style, not RadioButtonNormal.
// Spec: "Disabled items render with ClusterDisabled style."
func TestRBDisabledItemRendersWithClusterDisabledStyle(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.ClusterDisabled == scheme.RadioButtonNormal {
		t.Skip("ClusterDisabled equals RadioButtonNormal in BorlandBlue — test would be vacuous")
	}

	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})
	rbs.scheme = scheme
	rbs.SetEnabled(0, false)

	buf := NewDrawBuffer(20, 2)
	rbs.Draw(buf)

	// '(' at col 1, row 0 for a disabled item should use ClusterDisabled.
	cell := buf.GetCell(1, 0)
	if cell.Style == scheme.RadioButtonNormal {
		t.Errorf("disabled item '(' uses RadioButtonNormal; want ClusterDisabled")
	}
	if cell.Style != scheme.ClusterDisabled {
		t.Errorf("disabled item '(' style = %v, want ClusterDisabled %v", cell.Style, scheme.ClusterDisabled)
	}
}

// TestRBEnabledItemRendersWithRadioButtonNormalStyle verifies that an enabled item
// still uses RadioButtonNormal (not ClusterDisabled).
// Spec: "Disabled items render with ClusterDisabled style." — enabled items must not be affected.
func TestRBEnabledItemRendersWithRadioButtonNormalStyle(t *testing.T) {
	scheme := theme.BorlandBlue

	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})
	rbs.scheme = scheme
	rbs.SetEnabled(0, false) // only item 0 disabled; item 1 stays enabled

	buf := NewDrawBuffer(20, 2)
	rbs.Draw(buf)

	// '(' at col 1, row 1 (item 1) should use RadioButtonNormal.
	cell := buf.GetCell(1, 1)
	if cell.Style != scheme.RadioButtonNormal {
		t.Errorf("enabled item '(' style = %v, want RadioButtonNormal %v", cell.Style, scheme.RadioButtonNormal)
	}
}

// TestRBDisabledItemFocusIndicatorIsSpace verifies a disabled item never shows '►'
// at col 0, even when it has SfSelected.
// Spec: "Disabled items render with ClusterDisabled style."
func TestRBDisabledItemFocusIndicatorIsSpace(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})
	rbs.scheme = theme.BorlandBlue
	rbs.SetEnabled(0, false)
	rbs.Item(0).SetState(SfSelected, true)

	buf := NewDrawBuffer(20, 2)
	rbs.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == '►' {
		t.Error("disabled item shows '►' focus indicator; disabled items must show space at col 0")
	}
	if cell.Rune != ' ' {
		t.Errorf("disabled item col 0 = %q, want ' '", cell.Rune)
	}
}

// TestRBEnabledFocusedItemShowsArrowIndicator verifies an enabled, focused item
// shows '►' at col 0.
// Spec: "Disabled items render with ClusterDisabled style." — enabled items still show '►'.
func TestRBEnabledFocusedItemShowsArrowIndicator(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})
	rbs.scheme = theme.BorlandBlue
	rbs.SetState(SfSelected, true)
	rbs.SetFocusedChild(rbs.Item(0))

	buf := NewDrawBuffer(20, 2)
	rbs.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != '►' {
		t.Errorf("enabled focused item: col 0 = %q, want '►'", cell.Rune)
	}
}

// TestRBDisabledAndEnabledRenderWithDifferentStyles verifies disabled and enabled items
// actually render with different styles (guard against collapsed implementation).
// Falsifies: treating disabled the same as enabled in Draw.
func TestRBDisabledAndEnabledRenderWithDifferentStyles(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.ClusterDisabled == scheme.RadioButtonNormal {
		t.Skip("ClusterDisabled equals RadioButtonNormal in BorlandBlue — test would be vacuous")
	}

	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})
	rbs.scheme = scheme
	rbs.SetEnabled(0, false)

	buf := NewDrawBuffer(20, 2)
	rbs.Draw(buf)

	disabledStyle := buf.GetCell(1, 0).Style
	enabledStyle := buf.GetCell(1, 1).Style
	if disabledStyle == enabledStyle {
		t.Errorf("disabled and enabled items render with same style %v; expected different styles", disabledStyle)
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Navigation: Up/Down = row (delta ±1), Left/Right = column (delta ±height)
// ---------------------------------------------------------------------------

// TestRBUpDownDoRowNavigation verifies Up/Down moves by 1 item (row navigation).
// Spec: "Up/Down: move by 1 item (row navigation)."
func TestRBUpDownDoRowNavigation(t *testing.T) {
	// 4 items, height=2. Focus item 0. Down should move to item 1 (next row in same col).
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 1 {
		t.Errorf("Down from item 0: Value() = %d, want 1 (row nav delta=1)", rbs.Value())
	}
	if rbs.FocusedChild() != rbs.Item(1) {
		t.Errorf("Down from item 0: FocusedChild() = %v, want Item(1)", rbs.FocusedChild())
	}
}

// TestRBUpDoesRowNavigation verifies Up moves by 1 item backward (row navigation).
// Spec: "Up/Down: move by 1 item (row navigation)."
func TestRBUpDoesRowNavigation(t *testing.T) {
	// 4 items, height=2. Focus item 1. Up should move to item 0 (prev row same col).
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetValue(1)
	rbs.SetFocusedChild(rbs.Item(1))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("Up from item 1: Value() = %d, want 0 (row nav delta=1)", rbs.Value())
	}
}

// TestRBRightDoesColumnNavigation verifies Right moves by height items (column navigation).
// Spec: "Left/Right: move by height items (column navigation)."
func TestRBRightDoesColumnNavigation(t *testing.T) {
	// 4 items, height=2 → col0: (0,1), col1: (2,3).
	// Focus item 0. Right should move to item 0+height=2 (same row, next column).
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 2 {
		t.Errorf("Right from item 0 (height=2): Value() = %d, want 2 (column nav delta=height=2)", rbs.Value())
	}
	if rbs.FocusedChild() != rbs.Item(2) {
		t.Errorf("Right from item 0: FocusedChild() = %v, want Item(2)", rbs.FocusedChild())
	}
}

// TestRBLeftDoesColumnNavigation verifies Left moves by height items backward (column navigation).
// Spec: "Left/Right: move by height items (column navigation)."
func TestRBLeftDoesColumnNavigation(t *testing.T) {
	// 4 items, height=2. Focus item 2. Left → 2-height=0.
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("Left from item 2 (height=2): Value() = %d, want 0 (column nav delta=height=2)", rbs.Value())
	}
	if rbs.FocusedChild() != rbs.Item(0) {
		t.Errorf("Left from item 2: FocusedChild() = %v, want Item(0)", rbs.FocusedChild())
	}
}

// TestRBRightColumnNavAtLastColumnIsNoOp verifies Right at the last column does nothing.
// Spec: "Left/Right: move by height items (column navigation)."
func TestRBRightColumnNavAtLastColumnIsNoOp(t *testing.T) {
	// 4 items, height=2. Focus item 2 (col1, row0). Right → 2+2=4 >= 4 → no-op.
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 2 {
		t.Errorf("Right at last column: Value() = %d, want 2 (no-op)", rbs.Value())
	}
}

// TestRBLeftColumnNavAtFirstColumnIsNoOp verifies Left at the first column does nothing.
// Spec: "Left/Right: move by height items (column navigation)."
func TestRBLeftColumnNavAtFirstColumnIsNoOp(t *testing.T) {
	// 4 items, height=2. Focus item 0 (col0). Left → 0-2=-2 < 0 → no-op.
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("Left at first column: Value() = %d, want 0 (no-op)", rbs.Value())
	}
}

// TestRBDownAtLastRowInColumnIsNoOp verifies Down at the bottom of a column does nothing.
// Spec: "Up/Down: move by 1 item (row navigation)."
func TestRBDownAtLastRowInColumnIsNoOp(t *testing.T) {
	// Single column, 3 items. Focus item 2 (last). Down → 3 >= 3 → no-op.
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 2 {
		t.Errorf("Down at last item: Value() = %d, want 2 (no-op)", rbs.Value())
	}
}

// TestRBNavigationMovesSelectionWithFocus verifies that navigation moves BOTH
// focus AND selection (unlike CheckBoxes where navigation only moves focus).
// Spec: "Navigation moves both focus AND selection for RadioButtons."
func TestRBNavigationMovesSelectionWithFocus(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	// Both focus and selection must have moved.
	if rbs.FocusedChild() != rbs.Item(1) {
		t.Errorf("Down: FocusedChild = %v, want Item(1)", rbs.FocusedChild())
	}
	if rbs.Value() != 1 {
		t.Errorf("Down: Value() = %d, want 1; selection must move with focus", rbs.Value())
	}
	if rbs.Item(0).Selected() {
		t.Error("Down: Item(0) still selected; navigation must deselect previous item")
	}
}

// TestRBRightArrowClearsEvent verifies Right arrow clears the event when focused.
// Spec: (events consumed when focused, consistent with Up/Down/Left).
func TestRBRightArrowClearsEventWhenFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Right arrow (focused) did not clear event; ev.What = %v", ev.What)
	}
}

// TestRBLeftArrowClearsEventWhenFocused verifies Left arrow clears the event when focused.
func TestRBLeftArrowClearsEventWhenFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Left arrow (focused) did not clear event; ev.What = %v", ev.What)
	}
}

// TestRBRightArrowNotConsumedWhenNotFocused verifies Right arrow is NOT consumed
// when RadioButtons is not focused.
// Spec: "Only consume these keys when RadioButtons itself has focus (SfSelected=true)."
func TestRBRightArrowNotConsumedWhenNotFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	// Do NOT set SfSelected.

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Right arrow consumed by unfocused RadioButtons; must pass through")
	}
}

// TestRBLeftArrowNotConsumedWhenNotFocused verifies Left arrow is NOT consumed
// when RadioButtons is not focused.
func TestRBLeftArrowNotConsumedWhenNotFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetValue(2)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	rbs.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Left arrow consumed by unfocused RadioButtons; must pass through")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Navigation skips disabled items
// ---------------------------------------------------------------------------

// TestRBDownArrowSkipsDisabledItem verifies Down skips disabled items.
// Spec: "Navigation skips disabled items."
func TestRBDownArrowSkipsDisabledItem(t *testing.T) {
	// Items 0=enabled, 1=disabled, 2=enabled. Down from 0 → skip 1 → land on 2.
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetEnabled(1, false)
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 2 {
		t.Errorf("Down with disabled item 1: Value() = %d, want 2 (skip disabled)", rbs.Value())
	}
}

// TestRBUpArrowSkipsDisabledItem verifies Up skips disabled items.
// Spec: "Navigation skips disabled items."
func TestRBUpArrowSkipsDisabledItem(t *testing.T) {
	// Items 0=enabled, 1=disabled, 2=enabled. Up from 2 → skip 1 → land on 0.
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetEnabled(1, false)
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("Up with disabled item 1: Value() = %d, want 0 (skip disabled)", rbs.Value())
	}
}

// TestRBDownArrowDoesNotMoveWhenOnlyDisabledItemsRemain verifies Down does nothing
// when no enabled item exists beyond the current position.
// Spec: "Navigation skips disabled items." — with no valid target, don't move.
func TestRBDownArrowDoesNotMoveWhenOnlyDisabledItemsRemain(t *testing.T) {
	// Items 0=enabled, 1=disabled. Down from 0 → only disabled ahead → no-op.
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"A", "B"})
	rbs.SetEnabled(1, false)
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("Down with only disabled ahead: Value() = %d, want 0 (no-op)", rbs.Value())
	}
}

// TestRBRightColumnNavSkipsDisabledItem verifies Right (column nav) skips
// disabled items when the target column item is disabled.
// Spec: "Navigation skips disabled items."
func TestRBRightColumnNavSkipsDisabledTarget(t *testing.T) {
	// 4 items, height=2. col0: (0,1), col1: (2,3). Disable item 2.
	// Focus item 0. Right → target=2 (disabled) → no-op.
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetEnabled(2, false)
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("Right to disabled item 2: Value() = %d, want 0 (no-op)", rbs.Value())
	}
}

// TestRBLeftColumnNavSkipsDisabledItem verifies Left (column nav) skips
// disabled items when the target column item is disabled.
// Spec: "Navigation skips disabled items."
func TestRBLeftColumnNavSkipsDisabledTarget(t *testing.T) {
	// 4 items, height=2. Disable item 0. Focus item 2. Left → target=0 (disabled) → no-op.
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetEnabled(0, false)
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 2 {
		t.Errorf("Left to disabled item 0: Value() = %d, want 2 (no-op)", rbs.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Interaction blocked on disabled items
// ---------------------------------------------------------------------------

// TestRBSpaceBarOnDisabledItemDoesNotSelect verifies Space on a disabled item is a no-op.
// Spec: "Space/Enter on a disabled item does nothing."
func TestRBSpaceBarOnDisabledItemDoesNotSelect(t *testing.T) {
	// Item 1 is initially selected. Item 0 is disabled and focused.
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"A", "B"})
	rbs.SetValue(1)
	rbs.SetEnabled(0, false)
	rbs.SetFocusedChild(rbs.Item(0))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	rbs.HandleEvent(ev)

	// Value should remain 1; the disabled item 0 must not become selected.
	if rbs.Value() != 1 {
		t.Errorf("Space on disabled item 0: Value() = %d, want 1 (no change)", rbs.Value())
	}
}

// TestRBMouseClickOnDisabledItemIsIgnored verifies a mouse click on a disabled item
// does not change the selection.
// Spec: "Mouse clicks on disabled items are ignored."
func TestRBMouseClickOnDisabledItemIsIgnored(t *testing.T) {
	// Single-column, 2 items. Item 1 selected. Item 0 disabled.
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"A", "B"})
	rbs.SetValue(1)
	rbs.SetEnabled(0, false)

	// Click on item 0's row (row=0).
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 0, Button: tcell.Button1}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 1 {
		t.Errorf("mouse click on disabled item 0: Value() = %d, want 1 (no change)", rbs.Value())
	}
}

// TestRBMouseClickOnDisabledItemIsConsumed verifies the click event is still consumed
// even though the action is ignored.
// Spec: "Mouse clicks on disabled items are ignored."
func TestRBMouseClickOnDisabledItemIsConsumed(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"A", "B"})
	rbs.SetEnabled(0, false)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 0, Button: tcell.Button1}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("click on disabled item was not consumed; spec says event is still consumed")
	}
}

// TestRBAltShortcutOnDisabledItemIsIgnored verifies Alt+shortcut for a disabled item
// does not select it.
// Spec: "Alt+shortcut for disabled items is ignored."
func TestRBAltShortcutOnDisabledItemIsIgnored(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	rbs.SetValue(1) // item 1 selected
	rbs.SetEnabled(0, false)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: tcell.ModAlt}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 1 {
		t.Errorf("Alt+s on disabled item 0: Value() = %d, want 1 (no change)", rbs.Value())
	}
}

// TestRBAltShortcutOnDisabledItemDoesNotMoveFocus verifies focus does not change
// when Alt+shortcut targets a disabled item.
// Spec: "Alt+shortcut for disabled items is ignored."
func TestRBAltShortcutOnDisabledItemDoesNotMoveFocus(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	rbs.SetEnabled(0, false)
	rbs.SetFocusedChild(rbs.Item(1))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: tcell.ModAlt}}
	rbs.HandleEvent(ev)

	if rbs.FocusedChild() != rbs.Item(1) {
		t.Errorf("Alt+s to disabled item: focus moved to %v, want Item(1) (no change)", rbs.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Selected item remains when disabled; OfSelectable management
// ---------------------------------------------------------------------------

// TestRBDisablingSelectedItemDoesNotClearSelection verifies that disabling the currently
// selected item does NOT change the selection.
// Spec: "When the currently selected radio button becomes disabled, the selection does NOT change."
func TestRBDisablingSelectedItemDoesNotClearSelection(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(1) // item 1 is selected

	rbs.SetEnabled(1, false) // disable the currently selected item

	if rbs.Value() != 1 {
		t.Errorf("disabling selected item 1: Value() = %d, want 1 (selection must not change)", rbs.Value())
	}
}

// TestRBAllDisabledRemovesOfSelectable verifies OfSelectable is cleared when all
// items are disabled.
// Spec: "If all items become disabled, RadioButtons clears OfSelectable so it is
// skipped by Tab navigation."
func TestRBAllDisabledRemovesOfSelectable(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"A", "B"})

	rbs.SetEnabled(0, false)
	rbs.SetEnabled(1, false)

	if rbs.HasOption(OfSelectable) {
		t.Error("after disabling all items, OfSelectable should be cleared; RadioButtons must be skipped by Tab")
	}
}

// TestRBOneReEnabledRestoresOfSelectable verifies OfSelectable is restored when at
// least one item is re-enabled.
// Spec: "When at least one item is re-enabled, OfSelectable is restored."
func TestRBOneReEnabledRestoresOfSelectable(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"A", "B"})
	rbs.SetEnabled(0, false)
	rbs.SetEnabled(1, false)

	rbs.SetEnabled(0, true)

	if !rbs.HasOption(OfSelectable) {
		t.Error("after re-enabling one item, OfSelectable should be restored")
	}
}

// TestRBPartialDisableKeepsOfSelectable verifies OfSelectable remains when at least
// one item is still enabled.
// Spec: "If all items become disabled, RadioButtons clears OfSelectable."
// — partial disable must NOT clear it.
func TestRBPartialDisableKeepsOfSelectable(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	rbs.SetEnabled(0, false)
	rbs.SetEnabled(1, false)
	// Item 2 remains enabled.

	if !rbs.HasOption(OfSelectable) {
		t.Error("with one item still enabled, OfSelectable must remain set")
	}
}

// TestRBAllDisabledThenReEnabledRestoresSelectability verifies the full cycle:
// all-disabled (OfSelectable cleared) → re-enable one → OfSelectable restored.
// Spec: "When at least one item is re-enabled, OfSelectable is restored."
func TestRBAllDisabledThenReEnabledRestoresSelectability(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	rbs.SetEnabled(0, false)
	rbs.SetEnabled(1, false)
	rbs.SetEnabled(2, false)

	if rbs.HasOption(OfSelectable) {
		t.Fatal("precondition: all-disabled should clear OfSelectable")
	}

	rbs.SetEnabled(1, true)

	if !rbs.HasOption(OfSelectable) {
		t.Error("after re-enabling item 1, OfSelectable not restored; RadioButtons unreachable by Tab")
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Plain-letter shortcut matching
// ---------------------------------------------------------------------------

// TestRBPlainLetterShortcutWhenFocused verifies a plain letter selects the matching
// item when RadioButtons is focused (SfSelected=true).
// Spec: "When RadioButtons has focus (SfSelected is true), a plain letter keystroke
// (no Alt modifier) that matches a shortcut activates that item."
func TestRBPlainLetterShortcutWhenFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	rbs.SetValue(1) // start with item 1 selected
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: 0}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("plain 's' when focused: Value() = %d, want 0", rbs.Value())
	}
}

// TestRBPlainLetterShortcutFocusesItem verifies a plain letter shortcut also moves
// focus to the matching item when RadioButtons is focused.
// Spec: "plain letter keystroke … activates that item — same as Alt+letter but only when focused."
func TestRBPlainLetterShortcutFocusesItem(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	rbs.SetState(SfSelected, true)
	rbs.SetFocusedChild(rbs.Item(1))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: 0}}
	rbs.HandleEvent(ev)

	if rbs.FocusedChild() != rbs.Item(0) {
		t.Errorf("plain 's' when focused: FocusedChild() = %v, want Item(0)", rbs.FocusedChild())
	}
}

// TestRBPlainLetterShortcutConsumesEventWhenFocused verifies the event is consumed
// when a plain shortcut is activated while focused.
func TestRBPlainLetterShortcutConsumesEventWhenFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: 0}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("plain 's' when focused: event not consumed; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestRBPlainLetterShortcutOnDisabledItemIgnoredWhenFocused verifies that a plain
// letter shortcut targeting a disabled item is ignored.
// Spec: "Alt+shortcut for disabled items is ignored." — same applies to plain-letter.
func TestRBPlainLetterShortcutOnDisabledItemIgnoredWhenFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	rbs.SetValue(1)
	rbs.SetEnabled(0, false)
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: 0}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 1 {
		t.Errorf("plain 's' on disabled item 0: Value() = %d, want 1 (no change)", rbs.Value())
	}
}

// TestRBPlainLetterShortcutInPostProcessPhaseActivatesItem verifies that when
// RadioButtons is NOT focused but has OfPostProcess, a matching plain letter
// activates the item.
// Spec: "When RadioButtons does NOT have focus but receives an event during the
// postProcess phase (OfPostProcess), it also matches plain-letter shortcuts."
func TestRBPlainLetterShortcutInPostProcessPhaseActivatesItem(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})

	if !rbs.HasOption(OfPostProcess) {
		t.Skip("RadioButtons does not have OfPostProcess; cannot test post-process plain-letter shortcut")
	}

	// Not focused (SfSelected=false).
	rbs.SetValue(1)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: 0}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("plain 's' in postProcess phase (not focused): Value() = %d, want 0", rbs.Value())
	}
}

// TestRBSetsOfPostProcess verifies NewRadioButtons sets OfPostProcess.
// Spec: "When RadioButtons does NOT have focus but receives an event during the
// postProcess phase (OfPostProcess)…"
func TestRBSetsOfPostProcess(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 2), []string{"A", "B"})

	if !rbs.HasOption(OfPostProcess) {
		t.Error("NewRadioButtons did not set OfPostProcess; required for post-process plain-letter shortcut")
	}
}

// ---------------------------------------------------------------------------
// Section 9 — SetBounds relayout
// ---------------------------------------------------------------------------

// TestRBSetBoundsTriggersRelayout verifies that calling SetBounds with a new height
// recomputes the column layout.
// Spec: "SetBounds triggers relayout."
func TestRBSetBoundsTriggersRelayout(t *testing.T) {
	// Start with height=3 (single column for 3 items).
	rbs := NewRadioButtons(NewRect(0, 0, 40, 3), []string{"Alpha", "Beta", "Gamma"})

	// Confirm single column.
	for i := 0; i < 3; i++ {
		if rbs.Item(i).Bounds().A.X != 0 {
			t.Fatalf("precondition: Item(%d).x = %d, want 0 (single column)", i, rbs.Item(i).Bounds().A.X)
		}
	}

	// Shrink height → 2 columns.
	rbs.SetBounds(NewRect(0, 0, 40, 2))

	if rbs.Item(2).Bounds().A.X <= 0 {
		t.Errorf("after SetBounds(height=2), Item(2).x = %d, want > 0 (moved to second column)", rbs.Item(2).Bounds().A.X)
	}
}

// TestRBSetBoundsToTallerHeightCollapsesToSingleColumn verifies that increasing height
// to >= numItems produces a single-column layout.
// Spec: "SetBounds triggers relayout."
func TestRBSetBoundsToTallerHeightCollapsesToSingleColumn(t *testing.T) {
	// Start with height=2 (2 columns for 3 items).
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"Alpha", "Beta", "Gamma"})

	rbs.SetBounds(NewRect(0, 0, 60, 3))

	for i := 0; i < 3; i++ {
		if rbs.Item(i).Bounds().A.X != 0 {
			t.Errorf("after SetBounds(height=3), Item(%d).x = %d, want 0 (single column)", i, rbs.Item(i).Bounds().A.X)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 10 — Falsifying tests
// ---------------------------------------------------------------------------

// TestRBMulticolumnItem0And1NotInSameColumnWhenHeightIs1 verifies height=1 puts
// every item in its own column.
// Falsifies: an implementation that collapses to 1 column regardless of height.
func TestRBMulticolumnItem0And1NotInSameColumnWhenHeightIs1(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 90, 1), []string{"A", "B", "C"})

	x0 := rbs.Item(0).Bounds().A.X
	x1 := rbs.Item(1).Bounds().A.X
	x2 := rbs.Item(2).Bounds().A.X

	if x0 == x1 || x1 == x2 || x0 == x2 {
		t.Errorf("height=1 with 3 items: must be in 3 distinct columns, got x=[%d,%d,%d]", x0, x1, x2)
	}
}

// TestRBRightDoesNotDoRowNavWhenMulticolumn verifies Right uses column delta (height),
// not row delta (1), in a multi-column layout.
// Falsifies: an implementation that treats Right the same as Down in all cases.
func TestRBRightDoesNotDoRowNavWhenMulticolumn(t *testing.T) {
	// height=2. Item 0 at col0/row0. Right → item 2 (col1/row0), NOT item 1 (col0/row1).
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	if rbs.Value() == 1 {
		t.Errorf("Right moved to item 1 (row nav delta=1); should move to item 2 (column nav delta=height=2)")
	}
	if rbs.Value() != 2 {
		t.Errorf("Right from item 0: Value() = %d, want 2 (column nav)", rbs.Value())
	}
}

// TestRBLeftDoesNotDoRowNavWhenMulticolumn verifies Left uses column delta (height),
// not row delta (1), in a multi-column layout.
// Falsifies: an implementation that treats Left the same as Up.
func TestRBLeftDoesNotDoRowNavWhenMulticolumn(t *testing.T) {
	// height=2. Item 3 at col1/row1. Left → item 1 (col0/row1), NOT item 2 (col1/row0).
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D"})
	rbs.SetValue(3)
	rbs.SetFocusedChild(rbs.Item(3))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	rbs.HandleEvent(ev)

	if rbs.Value() == 2 {
		t.Errorf("Left moved to item 2 (row nav delta=1); should move to item 1 (column nav delta=height=2)")
	}
	if rbs.Value() != 1 {
		t.Errorf("Left from item 3: Value() = %d, want 1 (column nav)", rbs.Value())
	}
}

// TestRBDisablingSelectedItemKeepsSelection verifies that disabling the selected item
// does NOT change Value().
// Falsifies: an implementation that clears selection when an item is disabled.
func TestRBDisablingSelectedItemKeepsSelection(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(2)

	rbs.SetEnabled(2, false)

	if rbs.Value() != 2 {
		t.Errorf("disabling selected item 2: Value() = %d, want 2 (selection must not auto-change)", rbs.Value())
	}
}

// TestRBSetEnabledFalseDoesNotChangeSelectedValue verifies that disabling a non-selected
// item does not change the selection.
// Falsifies: an implementation that reassigns selection when any item is disabled.
func TestRBSetEnabledFalseDoesNotChangeSelectedValue(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(0)

	rbs.SetEnabled(1, false) // disable a non-selected item

	if rbs.Value() != 0 {
		t.Errorf("SetEnabled(1, false) changed Value() to %d; must not change selection", rbs.Value())
	}
}

// TestRBUpDownAreDistinctFromLeftRightInSingleColumn verifies that in a single-column
// layout with height >= len(items), Up/Down (delta=1) and Left/Right (delta=height=len(items))
// behave differently — Left/Right at the first or last item is a no-op, while
// Up/Down navigates row by row.
// Falsifies: collapsing Left/Right and Up/Down to the same operation.
func TestRBUpDownAreDistinctFromLeftRightInSingleColumn(t *testing.T) {
	// Single column: height=3, 3 items. height == len(items), so Left/Right delta = 3.
	// From item 0: Down moves to item 1; Right (delta=3) → index 3 >= 3 → no-op.
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))
	rbs.SetState(SfSelected, true)

	// Down moves by 1.
	downEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(downEv)
	if rbs.Value() != 1 {
		t.Errorf("Down from item 0 (single col): Value() = %d, want 1", rbs.Value())
	}

	// Reset to item 0.
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))

	// Right (delta=height=3) from item 0 → index 3 >= 3 → no-op.
	rightEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(rightEv)
	if rbs.Value() != 0 {
		t.Errorf("Right from item 0 (single col, delta=height=3): Value() = %d, want 0 (no-op)", rbs.Value())
	}
}
