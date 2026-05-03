package tv

// checkbox_multicolumn_test.go — Tests for Task 2: Multi-column CheckBoxes with Enable/Disable.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Multi-column layout
//   Section 2  — SetEnabled / IsEnabled API
//   Section 3  — Enable/disable rendering
//   Section 4  — Navigation skips disabled items
//   Section 5  — Interaction blocked on disabled items
//   Section 6  — OfSelectable management (all-disabled removes it)
//   Section 7  — Plain-letter shortcut matching
//   Section 8  — SetBounds relayout
//   Section 9  — Falsifying tests

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Section 1 — Multi-column layout
// ---------------------------------------------------------------------------

// TestMulticolumnTwoColumnsWhenHeightLessThanItemCount verifies that 3 items in a
// bounds with height=2 produce a 2-column layout (items 0,1 in col 0; item 2 in col 1).
// Spec: "numCols = ceil(len(items) / bounds.Height()). If bounds.Height() >= len(items),
// everything fits in one column."
func TestMulticolumnTwoColumnsWhenHeightLessThanItemCount(t *testing.T) {
	// height=2, 3 items → ceil(3/2) = 2 columns
	cbs := NewCheckBoxes(NewRect(0, 0, 30, 2), []string{"Alpha", "Beta", "Gamma"})

	// Item 0: row=0%2=0, col=0/2=0 → x=0, y=0
	item0 := cbs.Item(0)
	if item0.Bounds().A.Y != 0 {
		t.Errorf("Item(0) row = %d, want 0 (first item, first column)", item0.Bounds().A.Y)
	}
	if item0.Bounds().A.X != 0 {
		t.Errorf("Item(0) col start x = %d, want 0 (first column)", item0.Bounds().A.X)
	}

	// Item 1: row=1%2=1, col=1/2=0 → x=0, y=1
	item1 := cbs.Item(1)
	if item1.Bounds().A.Y != 1 {
		t.Errorf("Item(1) row = %d, want 1 (second row, first column)", item1.Bounds().A.Y)
	}
	if item1.Bounds().A.X != 0 {
		t.Errorf("Item(1) col start x = %d, want 0 (still first column)", item1.Bounds().A.X)
	}

	// Item 2: row=2%2=0, col=2/2=1 → x=colWidth0, y=0
	item2 := cbs.Item(2)
	if item2.Bounds().A.Y != 0 {
		t.Errorf("Item(2) row = %d, want 0 (first row, second column)", item2.Bounds().A.Y)
	}
	if item2.Bounds().A.X <= 0 {
		t.Errorf("Item(2) x = %d, want > 0 (should be in second column)", item2.Bounds().A.X)
	}
}

// TestMulticolumnOneColumnWhenHeightEqualsItemCount verifies that 3 items in a
// bounds with height=3 produce a single column (backward compatible).
// Spec: "If bounds.Height() >= len(items), everything fits in one column (backward compatible)."
func TestMulticolumnOneColumnWhenHeightEqualsItemCount(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})

	// All items should be in column 0 (x=0) with y=index
	for i := 0; i < 3; i++ {
		item := cbs.Item(i)
		if item.Bounds().A.X != 0 {
			t.Errorf("Item(%d).Bounds().X = %d, want 0 (single column when height >= items)", i, item.Bounds().A.X)
		}
		if item.Bounds().A.Y != i {
			t.Errorf("Item(%d).Bounds().Y = %d, want %d (y=index in single column)", i, item.Bounds().A.Y, i)
		}
	}
}

// TestMulticolumnOneColumnWhenHeightExceedsItemCount verifies that extra height
// still produces a single column layout.
// Spec: "If bounds.Height() >= len(items), everything fits in one column."
func TestMulticolumnOneColumnWhenHeightExceedsItemCount(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 5), []string{"Alpha", "Beta"})

	for i := 0; i < 2; i++ {
		item := cbs.Item(i)
		if item.Bounds().A.X != 0 {
			t.Errorf("Item(%d).Bounds().X = %d, want 0 (single column when height > items)", i, item.Bounds().A.X)
		}
	}
}

// TestMulticolumnItemRowIsItemModHeight verifies row = item % height.
// Spec: "Items fill top-to-bottom within a column: row(item) = item % height."
func TestMulticolumnItemRowIsItemModHeight(t *testing.T) {
	// 4 items, height=2 → 2 columns. Item 2 is in col 1, row 0 (2%2=0).
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"A", "B", "C", "D"})

	// item 0: row=0, item 1: row=1, item 2: row=0, item 3: row=1
	expected := []int{0, 1, 0, 1}
	for i, want := range expected {
		got := cbs.Item(i).Bounds().A.Y
		if got != want {
			t.Errorf("Item(%d) row = %d, want %d (item %% height=%d)", i, got, want, 2)
		}
	}
}

// TestMulticolumnColumnXIsAccumulatedWidths verifies col 1 starts at column 0's width.
// Spec: "Column x-position = sum of widths of all previous columns."
func TestMulticolumnColumnXIsAccumulatedWidths(t *testing.T) {
	// 3 items, height=2. Column 0 contains items 0,1; column 1 contains item 2.
	// col0 width = widestLabel(items 0,1) + 6.
	// item 2's x should equal col0 width.
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"AB", "CD", "EF"})

	item0x := cbs.Item(0).Bounds().A.X
	item2x := cbs.Item(2).Bounds().A.X
	item0w := cbs.Item(0).Bounds().Width()

	if item0x != 0 {
		t.Errorf("Item(0) x = %d, want 0 (first column starts at 0)", item0x)
	}
	if item2x != item0w {
		t.Errorf("Item(2) x = %d, want %d (second column starts at col0 width)", item2x, item0w)
	}
}

// TestMulticolumnColumnWidthIncludesLabelPlusSix verifies colWidth = widestLabel + 6.
// Spec: "Column width = width of the widest label in that column + 6
// (1 focus + 1 '[' + 1 mark + 1 ']' + 1 space + 1 inter-column gap)."
func TestMulticolumnColumnWidthIncludesLabelPlusSix(t *testing.T) {
	// Single label "AB" (2 runes visible after tilde strip) → colWidth = 2 + 6 = 8.
	// Use a single-column layout (height >= items).
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 1), []string{"AB"})

	item := cbs.Item(0)
	want := tildeTextLen("AB") + 6 // 2 + 6 = 8
	if item.Bounds().Width() != want {
		t.Errorf("single-item colWidth = %d, want %d (tildeTextLen + 6)", item.Bounds().Width(), want)
	}
}

// TestMulticolumnWidestLabelDeterminesColumnWidth verifies that column width is based
// on the widest item in that column, not just the first.
// Spec: "Column width = width of the widest label in that column + 6."
func TestMulticolumnWidestLabelDeterminesColumnWidth(t *testing.T) {
	// 2 items in col 0: "A" (1 rune) and "LONGNAME" (8 runes).
	// height=2 so both fit in col 0.
	// colWidth = 8 + 6 = 14.
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"A", "LONGNAME", "X"})

	// Both items in col 0 should have the same width.
	w0 := cbs.Item(0).Bounds().Width()
	w1 := cbs.Item(1).Bounds().Width()

	if w0 != w1 {
		t.Errorf("Items in same column have different widths: Item(0)=%d, Item(1)=%d; all items in a column must share column width", w0, w1)
	}

	want := tildeTextLen("LONGNAME") + 6 // 8 + 6 = 14
	if w0 != want {
		t.Errorf("col0 width = %d, want %d (widest label + 6)", w0, want)
	}
}

// TestMulticolumnItemBoundsHeightIsOne verifies each item's bounds height is 1.
// Spec: "Each item's Bounds are set to NewRect(colX, row, colWidth, 1)."
func TestMulticolumnItemBoundsHeightIsOne(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"Alpha", "Beta", "Gamma"})

	for i := 0; i < 3; i++ {
		if cbs.Item(i).Bounds().Height() != 1 {
			t.Errorf("Item(%d).Bounds().Height() = %d, want 1", i, cbs.Item(i).Bounds().Height())
		}
	}
}

// TestMulticolumnCol1ItemInCorrectColumn verifies a 3-column layout positions
// items correctly across columns.
// Spec: "col(item) = item / height."
func TestMulticolumnCol1ItemInCorrectColumn(t *testing.T) {
	// 6 items, height=2 → 3 columns.
	// col 0: items 0,1; col 1: items 2,3; col 2: items 4,5.
	cbs := NewCheckBoxes(NewRect(0, 0, 60, 2), []string{"A", "B", "C", "D", "E", "F"})

	x0 := cbs.Item(0).Bounds().A.X
	x2 := cbs.Item(2).Bounds().A.X
	x4 := cbs.Item(4).Bounds().A.X

	// All three x-positions should be distinct and increasing.
	if x0 >= x2 {
		t.Errorf("Item(2).x=%d should be greater than Item(0).x=%d (in next column)", x2, x0)
	}
	if x2 >= x4 {
		t.Errorf("Item(4).x=%d should be greater than Item(2).x=%d (in next column)", x4, x2)
	}
	// Items in the same column share x.
	if cbs.Item(0).Bounds().A.X != cbs.Item(1).Bounds().A.X {
		t.Errorf("Items 0 and 1 in same column have different x: %d vs %d", cbs.Item(0).Bounds().A.X, cbs.Item(1).Bounds().A.X)
	}
}

// ---------------------------------------------------------------------------
// Section 2 — SetEnabled / IsEnabled API
// ---------------------------------------------------------------------------

// TestSetEnabledIsEnabledDefaultTrue verifies all items are enabled by default.
// Spec: "Internally stored as a bitmask enableMask uint32 (all bits set = all enabled by default)."
func TestSetEnabledIsEnabledDefaultTrue(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	for i := 0; i < 3; i++ {
		if !cbs.IsEnabled(i) {
			t.Errorf("IsEnabled(%d) = false on newly created CheckBoxes; all items must be enabled by default", i)
		}
	}
}

// TestSetEnabledFalseDisablesItem verifies SetEnabled(i, false) disables that item.
// Spec: "CheckBoxes has a SetEnabled(index int, enabled bool) method."
func TestSetEnabledFalseDisablesItem(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	cbs.SetEnabled(1, false)

	if cbs.IsEnabled(1) {
		t.Error("after SetEnabled(1, false), IsEnabled(1) = true; expected false")
	}
}

// TestSetEnabledTrueReEnablesItem verifies SetEnabled(i, true) re-enables a disabled item.
// Spec: "CheckBoxes has a SetEnabled(index int, enabled bool) method."
func TestSetEnabledTrueReEnablesItem(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetEnabled(0, false)

	cbs.SetEnabled(0, true)

	if !cbs.IsEnabled(0) {
		t.Error("after SetEnabled(0, true) following disable, IsEnabled(0) = false; expected true")
	}
}

// TestSetEnabledOnlyAffectsTargetItem verifies SetEnabled on one item does not
// affect other items.
// Spec: "SetEnabled(index int, enabled bool)."
func TestSetEnabledOnlyAffectsTargetItem(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	cbs.SetEnabled(1, false)

	if !cbs.IsEnabled(0) {
		t.Error("SetEnabled(1, false) changed IsEnabled(0); should only affect index 1")
	}
	if !cbs.IsEnabled(2) {
		t.Error("SetEnabled(1, false) changed IsEnabled(2); should only affect index 1")
	}
}

// TestIsEnabledReturnsFalseForDisabledItem verifies IsEnabled returns false after
// disabling.
// Spec: "CheckBoxes has an IsEnabled(index int) bool method."
func TestIsEnabledReturnsFalseForDisabledItem(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetEnabled(2, false)

	if cbs.IsEnabled(2) {
		t.Error("IsEnabled(2) = true after SetEnabled(2, false); expected false")
	}
}

// TestSetEnabledMultipleItemsIndependently verifies each item's enabled state is
// independent of others.
// Spec: "Internally stored as a bitmask enableMask uint32."
func TestSetEnabledMultipleItemsIndependently(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 4), []string{"A", "B", "C", "D"})

	cbs.SetEnabled(0, false)
	cbs.SetEnabled(2, false)

	if cbs.IsEnabled(0) {
		t.Error("IsEnabled(0) = true after SetEnabled(0, false)")
	}
	if !cbs.IsEnabled(1) {
		t.Error("IsEnabled(1) = false; should be enabled (was not disabled)")
	}
	if cbs.IsEnabled(2) {
		t.Error("IsEnabled(2) = true after SetEnabled(2, false)")
	}
	if !cbs.IsEnabled(3) {
		t.Error("IsEnabled(3) = false; should be enabled (was not disabled)")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Enable/disable rendering
// ---------------------------------------------------------------------------

// TestDisabledItemRendersWithClusterDisabledStyle verifies a disabled item uses
// ClusterDisabled style instead of CheckBoxNormal.
// Spec: "Disabled items render using ClusterDisabled color scheme style instead of CheckBoxNormal."
func TestDisabledItemRendersWithClusterDisabledStyle(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.ClusterDisabled == scheme.CheckBoxNormal {
		t.Skip("ClusterDisabled equals CheckBoxNormal in BorlandBlue — test would be vacuous")
	}

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})
	cbs.scheme = scheme
	cbs.SetEnabled(0, false)

	buf := NewDrawBuffer(20, 2)
	cbs.Draw(buf)

	// '[' at col 1, row 0 should use ClusterDisabled style for a disabled item.
	cell := buf.GetCell(1, 0)
	if cell.Style == scheme.CheckBoxNormal {
		t.Errorf("disabled item '[' uses CheckBoxNormal style; want ClusterDisabled style")
	}
	if cell.Style != scheme.ClusterDisabled {
		t.Errorf("disabled item '[' style = %v, want ClusterDisabled %v", cell.Style, scheme.ClusterDisabled)
	}
}

// TestEnabledItemRendersWithCheckBoxNormalStyle verifies an enabled item still uses
// CheckBoxNormal style (not ClusterDisabled).
// Spec: "Disabled items render using ClusterDisabled color scheme style instead of CheckBoxNormal."
func TestEnabledItemRendersWithCheckBoxNormalStyle(t *testing.T) {
	scheme := theme.BorlandBlue

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})
	cbs.scheme = scheme
	cbs.SetEnabled(0, false) // disable item 0 only
	// Item 1 remains enabled

	buf := NewDrawBuffer(20, 2)
	cbs.Draw(buf)

	// '[' at col 1, row 1 (item 1) should use CheckBoxNormal.
	cell := buf.GetCell(1, 1)
	if cell.Style != scheme.CheckBoxNormal {
		t.Errorf("enabled item '[' style = %v, want CheckBoxNormal %v", cell.Style, scheme.CheckBoxNormal)
	}
}

// TestDisabledItemFocusIndicatorIsSpace verifies that a disabled item never shows
// the '►' focus indicator at col 0, even when it has SfSelected.
// Spec: "Disabled items' focus indicator column shows a space (never ►)."
func TestDisabledItemFocusIndicatorIsSpace(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})
	cbs.scheme = theme.BorlandBlue
	cbs.SetEnabled(0, false)
	// Force SfSelected on the item to test that it's still suppressed.
	cbs.Item(0).SetState(SfSelected, true)

	buf := NewDrawBuffer(20, 2)
	cbs.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == '►' {
		t.Error("disabled item shows '►' focus indicator at col 0; disabled items must show space")
	}
	if cell.Rune != ' ' {
		t.Errorf("disabled item focus col = %q, want ' ' (space)", cell.Rune)
	}
}

// TestEnabledFocusedItemShowsArrowIndicator verifies that an enabled, focused item
// does show '►' at col 0 (not affected by disabled logic).
// Spec: "Disabled items' focus indicator column shows a space (never ►)." — enabled items still show it.
func TestEnabledFocusedItemShowsArrowIndicator(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})
	cbs.scheme = theme.BorlandBlue
	// Keep item 0 enabled (default).
	// Set CheckBoxes as focused so focus indicator appears.
	cbs.SetState(SfSelected, true)
	cbs.SetFocusedChild(cbs.Item(0))

	buf := NewDrawBuffer(20, 2)
	cbs.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != '►' {
		t.Errorf("enabled focused item: col 0 = %q, want '►'", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Navigation skips disabled items
// ---------------------------------------------------------------------------

// TestDownArrowSkipsDisabledItem verifies Down arrow skips over disabled items.
// Spec: "Keyboard navigation (Up/Down/Left/Right) skips disabled items."
func TestDownArrowSkipsDisabledItem(t *testing.T) {
	// Items: 0=enabled, 1=disabled, 2=enabled.
	// Down from 0 should skip 1 and land on 2.
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true)
	cbs.SetEnabled(1, false)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(2) {
		t.Errorf("Down with disabled item 1: FocusedChild() = %v, want Item(2)", cbs.FocusedChild())
	}
}

// TestUpArrowSkipsDisabledItem verifies Up arrow skips over disabled items.
// Spec: "Keyboard navigation (Up/Down/Left/Right) skips disabled items."
func TestUpArrowSkipsDisabledItem(t *testing.T) {
	// Items: 0=enabled, 1=disabled, 2=enabled.
	// Up from 2 should skip 1 and land on 0.
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(2))
	cbs.SetState(SfSelected, true)
	cbs.SetEnabled(1, false)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(0) {
		t.Errorf("Up with disabled item 1: FocusedChild() = %v, want Item(0)", cbs.FocusedChild())
	}
}

// TestDownArrowDoesNotMoveWhenNextIsDisabledAndPastEnd verifies Down at last enabled
// item does nothing when there's no next enabled item.
// Spec: "Up/Down arrows: same as current but skip disabled items."
func TestDownArrowDoesNotMoveWhenNextIsDisabledAndPastEnd(t *testing.T) {
	// Items: 0=enabled, 1=disabled. Down from 0 — next is disabled and there's
	// nothing after it, so no movement.
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true)
	cbs.SetEnabled(1, false)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(0) {
		t.Errorf("Down past all disabled: FocusedChild() = %v, want Item(0) (no move)", cbs.FocusedChild())
	}
}

// TestLeftArrowMovesToPreviousColumn verifies Left moves to item at currentIndex - height.
// Spec: "Left arrow: move focus to the item at currentIndex - height (previous column, same row)."
func TestLeftArrowMovesToPreviousColumn(t *testing.T) {
	// 4 items, height=2 → col 0: items 0,1; col 1: items 2,3.
	// Focus item 2 (col 1, row 0). Left should move to item 0 (col 0, row 0).
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"A", "B", "C", "D"})
	cbs.SetFocusedChild(cbs.Item(2))
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(0) {
		t.Errorf("Left from item 2 (col 1): FocusedChild() = %v, want Item(0) (col 0, same row)", cbs.FocusedChild())
	}
}

// TestLeftArrowDoesNothingAtFirstColumn verifies Left at column 0 does nothing
// (index - height would be < 0).
// Spec: "If that index is < 0 or disabled, don't move."
func TestLeftArrowDoesNothingAtFirstColumn(t *testing.T) {
	// 4 items, height=2. Focus item 0 (col 0, row 0). Left → index = 0 - 2 = -2 < 0.
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"A", "B", "C", "D"})
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(0) {
		t.Errorf("Left at first column: FocusedChild() = %v, want Item(0) (no move)", cbs.FocusedChild())
	}
}

// TestRightArrowMovesToNextColumn verifies Right moves to item at currentIndex + height.
// Spec: "Right arrow: move focus to the item at currentIndex + height (next column, same row)."
func TestRightArrowMovesToNextColumn(t *testing.T) {
	// 4 items, height=2 → col 0: items 0,1; col 1: items 2,3.
	// Focus item 0 (col 0, row 0). Right should move to item 2 (col 1, row 0).
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"A", "B", "C", "D"})
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(2) {
		t.Errorf("Right from item 0 (col 0): FocusedChild() = %v, want Item(2) (col 1, same row)", cbs.FocusedChild())
	}
}

// TestRightArrowDoesNothingAtLastColumn verifies Right at the last column does nothing.
// Spec: "If that index is >= len(items) or disabled, don't move."
func TestRightArrowDoesNothingAtLastColumn(t *testing.T) {
	// 4 items, height=2. Focus item 2 (col 1, row 0). Right → 2+2=4 >= 4.
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"A", "B", "C", "D"})
	cbs.SetFocusedChild(cbs.Item(2))
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(2) {
		t.Errorf("Right at last column: FocusedChild() = %v, want Item(2) (no move)", cbs.FocusedChild())
	}
}

// TestLeftArrowDoesNothingWhenTargetIsDisabled verifies Left does not move when
// the target item is disabled.
// Spec: "If that index is < 0 or disabled, don't move."
func TestLeftArrowDoesNothingWhenTargetIsDisabled(t *testing.T) {
	// 4 items, height=2. Item 0 disabled. Focus item 2. Left → item 0 (disabled) → don't move.
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"A", "B", "C", "D"})
	cbs.SetEnabled(0, false)
	cbs.SetFocusedChild(cbs.Item(2))
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(2) {
		t.Errorf("Left to disabled item 0: FocusedChild() = %v, want Item(2) (no move)", cbs.FocusedChild())
	}
}

// TestRightArrowDoesNothingWhenTargetIsDisabled verifies Right does not move when
// the target item is disabled.
// Spec: "If that index is >= len(items) or disabled, don't move."
func TestRightArrowDoesNothingWhenTargetIsDisabled(t *testing.T) {
	// 4 items, height=2. Item 2 disabled. Focus item 0. Right → item 2 (disabled) → don't move.
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"A", "B", "C", "D"})
	cbs.SetEnabled(2, false)
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(0) {
		t.Errorf("Right to disabled item 2: FocusedChild() = %v, want Item(0) (no move)", cbs.FocusedChild())
	}
}

// TestArrowKeysClearEventWhenHandled verifies Left/Right arrow events are cleared.
// Spec: (arrow keys are handled analogously to Up/Down — consumed when focused).
func TestLeftArrowClearsEvent(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"A", "B", "C", "D"})
	cbs.SetFocusedChild(cbs.Item(2))
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	cbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Left arrow did not clear event; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestRightArrowClearsEvent verifies Right arrow event is cleared when focused.
func TestRightArrowClearsEvent(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"A", "B", "C", "D"})
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	cbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Right arrow did not clear event; ev.What = %v, want EvNothing", ev.What)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Interaction blocked on disabled items
// ---------------------------------------------------------------------------

// TestSpaceBarOnDisabledItemDoesNotToggle verifies Space on a disabled item is a no-op.
// Spec: "Space bar on a disabled item does nothing (no toggle)."
func TestSpaceBarOnDisabledItemDoesNotToggle(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})
	cbs.SetEnabled(0, false)
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	cbs.HandleEvent(ev)

	if cbs.Item(0).Checked() {
		t.Error("Space on disabled item toggled it; disabled items must not toggle on Space")
	}
}

// TestMouseClickOnDisabledItemIsIgnored verifies a mouse click on a disabled item
// does not toggle its state.
// Spec: "Mouse clicks on disabled items are ignored (the click does nothing, event is still consumed)."
func TestMouseClickOnDisabledItemIsIgnored(t *testing.T) {
	// height=2, item 0 at row 0.
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})
	cbs.SetEnabled(0, false)

	// Click within item 0's bounds (row 0, any column).
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 0, Button: tcell.Button1}}
	cbs.HandleEvent(ev)

	if cbs.Item(0).Checked() {
		t.Error("mouse click on disabled item toggled it; clicks on disabled items must be ignored")
	}
}

// TestMouseClickOnDisabledItemEventIsConsumed verifies the click event is still
// consumed (not passed on) even though the action is ignored.
// Spec: "Mouse clicks on disabled items are ignored (the click does nothing, event is still consumed)."
func TestMouseClickOnDisabledItemEventIsConsumed(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})
	cbs.SetEnabled(0, false)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 0, Button: tcell.Button1}}
	cbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("click on disabled item was not consumed; spec says event is still consumed")
	}
}

// TestAltShortcutOnDisabledItemIsIgnored verifies Alt+shortcut for a disabled item
// is ignored.
// Spec: "If a shortcut key (Alt+letter) matches a disabled item, the shortcut is ignored."
func TestAltShortcutOnDisabledItemIsIgnored(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	cbs.SetEnabled(0, false) // disable '~S~ave'

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: tcell.ModAlt}}
	cbs.HandleEvent(ev)

	if cbs.Item(0).Checked() {
		t.Error("Alt+s on disabled item toggled it; shortcuts to disabled items must be ignored")
	}
}

// TestAltShortcutOnDisabledItemDoesNotMoveFocus verifies focus does not change when
// Alt+shortcut targets a disabled item.
// Spec: "If a shortcut key (Alt+letter) matches a disabled item, the shortcut is ignored."
func TestAltShortcutOnDisabledItemDoesNotMoveFocus(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	cbs.SetEnabled(0, false)
	cbs.SetFocusedChild(cbs.Item(1))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: tcell.ModAlt}}
	cbs.HandleEvent(ev)

	// Focus should remain on item 1.
	if cbs.FocusedChild() != cbs.Item(1) {
		t.Errorf("Alt+s to disabled item moved focus to %v, want Item(1) (no focus change)", cbs.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Section 6 — OfSelectable management
// ---------------------------------------------------------------------------

// TestAllDisabledRemovesOfSelectable verifies OfSelectable is cleared when all
// items are disabled.
// Spec: "If all items become disabled, CheckBoxes clears OfSelectable so it is
// skipped by Tab navigation."
func TestAllDisabledRemovesOfSelectable(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})

	cbs.SetEnabled(0, false)
	cbs.SetEnabled(1, false)

	if cbs.HasOption(OfSelectable) {
		t.Error("after disabling all items, OfSelectable should be cleared; CheckBoxes should be skipped by Tab")
	}
}

// TestOneEnabledRestoresOfSelectable verifies OfSelectable is restored when at
// least one item is re-enabled.
// Spec: "When at least one item is re-enabled, OfSelectable is restored."
func TestOneEnabledRestoresOfSelectable(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})
	cbs.SetEnabled(0, false)
	cbs.SetEnabled(1, false)

	cbs.SetEnabled(0, true)

	if !cbs.HasOption(OfSelectable) {
		t.Error("after re-enabling one item, OfSelectable should be restored")
	}
}

// TestPartialDisableKeepsOfSelectable verifies OfSelectable remains when at least
// one item is still enabled.
// Spec: "If all items become disabled, CheckBoxes clears OfSelectable…"
// — partial disable must NOT clear it.
func TestPartialDisableKeepsOfSelectable(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	cbs.SetEnabled(0, false)
	cbs.SetEnabled(1, false)
	// Item 2 remains enabled

	if !cbs.HasOption(OfSelectable) {
		t.Error("with one item still enabled, OfSelectable must remain set")
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Plain-letter shortcut matching
// ---------------------------------------------------------------------------

// TestPlainLetterShortcutWhenFocused verifies a plain letter activates the matching
// item when CheckBoxes is focused (SfSelected=true).
// Spec: "When CheckBoxes has focus (SfSelected is true), a plain letter keystroke
// (no Alt modifier) that matches a shortcut activates that item."
func TestPlainLetterShortcutWhenFocused(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: 0}}
	cbs.HandleEvent(ev)

	if !cbs.Item(0).Checked() {
		t.Error("plain 's' when focused: item 0 not toggled; plain shortcut should work when focused")
	}
}

// TestPlainLetterShortcutFocusesItem verifies a plain letter shortcut also focuses
// the matching item when CheckBoxes is focused.
// Spec: "plain letter keystroke … activates that item — same as Alt+letter but only when focused."
func TestPlainLetterShortcutFocusesItem(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	cbs.SetState(SfSelected, true)
	cbs.SetFocusedChild(cbs.Item(1))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: 0}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(0) {
		t.Errorf("plain 's' when focused: FocusedChild() = %v, want Item(0)", cbs.FocusedChild())
	}
}

// TestPlainLetterShortcutConsumesEventWhenFocused verifies the event is consumed
// when a plain shortcut is activated while focused.
// Spec: "plain letter keystroke … activates that item."
func TestPlainLetterShortcutConsumesEventWhenFocused(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: 0}}
	cbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("plain 's' when focused: event not consumed; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestPlainLetterShortcutOnDisabledItemIgnoredWhenFocused verifies that a plain
// letter shortcut targeting a disabled item is ignored, even when focused.
// Spec: "If a shortcut key … matches a disabled item, the shortcut is ignored."
func TestPlainLetterShortcutOnDisabledItemIgnoredWhenFocused(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	cbs.SetEnabled(0, false)
	cbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: 0}}
	cbs.HandleEvent(ev)

	if cbs.Item(0).Checked() {
		t.Error("plain 's' on disabled item when focused: item toggled; disabled items must be immune")
	}
}

// TestPlainLetterShortcutInPostProcessPhaseRequestsFocus verifies that when CheckBoxes
// is NOT focused but has OfPostProcess, a matching plain letter causes it to request
// focus and activate the item.
// Spec: "When CheckBoxes does NOT have focus but receives an event during the
// postProcess phase (OfPostProcess), it also matches plain-letter shortcuts.
// On match, it requests focus from its owner before activating."
func TestPlainLetterShortcutInPostProcessPhaseRequestsFocus(t *testing.T) {
	// Verify OfPostProcess is set.
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})

	if !cbs.HasOption(OfPostProcess) {
		t.Skip("CheckBoxes does not have OfPostProcess; cannot test post-process plain-letter shortcut")
	}

	// Not focused (SfSelected=false), but OfPostProcess is set.
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: 0}}
	cbs.HandleEvent(ev)

	// Without an owner, we can't verify focus request, but the item should be activated.
	if !cbs.Item(0).Checked() {
		t.Error("plain 's' in postProcess phase (not focused): item 0 not toggled")
	}
}

// TestCheckBoxesSetsOfPostProcess verifies NewCheckBoxes sets OfPostProcess so it can
// receive events in the post-process phase.
// Spec: "When CheckBoxes does NOT have focus but receives an event during the
// postProcess phase (OfPostProcess)…"
func TestCheckBoxesSetsOfPostProcess(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})

	if !cbs.HasOption(OfPostProcess) {
		t.Error("NewCheckBoxes did not set OfPostProcess; required for post-process plain-letter shortcut")
	}
}

// ---------------------------------------------------------------------------
// Section 8 — SetBounds relayout
// ---------------------------------------------------------------------------

// TestSetBoundsTriggersRelayout verifies that calling SetBounds with a new height
// recomputes column layout.
// Spec: "SetBounds triggers relayout — changing height can change column count."
func TestSetBoundsTriggersRelayout(t *testing.T) {
	// Start with height=3 (single column for 3 items).
	cbs := NewCheckBoxes(NewRect(0, 0, 30, 3), []string{"Alpha", "Beta", "Gamma"})

	// Verify single column initially.
	for i := 0; i < 3; i++ {
		if cbs.Item(i).Bounds().A.X != 0 {
			t.Fatalf("precondition: Item(%d).x = %d, want 0 (single column)", i, cbs.Item(i).Bounds().A.X)
		}
	}

	// Change height to 2 → should produce 2 columns.
	cbs.SetBounds(NewRect(0, 0, 30, 2))

	// Item 2 should now be in column 1 (x > 0).
	item2x := cbs.Item(2).Bounds().A.X
	if item2x <= 0 {
		t.Errorf("after SetBounds(height=2), Item(2).x = %d, want > 0 (moved to second column)", item2x)
	}
}

// TestSetBoundsToTallerHeightCollapsesToSingleColumn verifies that increasing height
// to >= numItems produces a single column.
// Spec: "SetBounds triggers relayout — changing height can change column count."
func TestSetBoundsToTallerHeightCollapsesToSingleColumn(t *testing.T) {
	// Start with height=2 (2 columns for 3 items).
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"Alpha", "Beta", "Gamma"})

	// Change height to 3 → single column.
	cbs.SetBounds(NewRect(0, 0, 40, 3))

	for i := 0; i < 3; i++ {
		if cbs.Item(i).Bounds().A.X != 0 {
			t.Errorf("after SetBounds(height=3), Item(%d).x = %d, want 0 (single column)", i, cbs.Item(i).Bounds().A.X)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Falsifying tests
// ---------------------------------------------------------------------------

// TestMulticolumnItem0And1NotInSameColumnWhenHeightIs1 verifies a single-row layout
// forces every item into its own column.
// Falsifies: an implementation that collapses to 1 column regardless of height.
func TestMulticolumnItem0And1NotInSameColumnWhenHeightIs1(t *testing.T) {
	// height=1, 3 items → 3 columns. Each item has its own column.
	cbs := NewCheckBoxes(NewRect(0, 0, 60, 1), []string{"A", "B", "C"})

	x0 := cbs.Item(0).Bounds().A.X
	x1 := cbs.Item(1).Bounds().A.X
	x2 := cbs.Item(2).Bounds().A.X

	if x0 == x1 || x1 == x2 || x0 == x2 {
		t.Errorf("height=1 with 3 items: items should be in 3 distinct columns, got x=[%d,%d,%d]", x0, x1, x2)
	}
}

// TestMulticolumnDisabledItemsNotSkippedByValue verifies that disabled items still
// contribute to Values() bitmask (they remain check-able programmatically).
// Falsifies: an implementation that omits disabled items from Values().
func TestMulticolumnDisabledItemsNotSkippedByValue(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetEnabled(1, false)
	cbs.Item(1).SetChecked(true) // programmatically check disabled item

	got := cbs.Values()
	if got&0b010 == 0 {
		t.Errorf("Values() = %b; disabled but checked item 1 should still set bit 1", got)
	}
}

// TestDisabledItemStyleDiffersFromEnabled verifies that disabled and enabled items
// actually render with different styles (guard against collapsed implementation).
// Falsifies: treating disabled the same as enabled in Draw.
func TestDisabledItemStyleDiffersFromEnabled(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.ClusterDisabled == scheme.CheckBoxNormal {
		t.Skip("ClusterDisabled equals CheckBoxNormal in BorlandBlue — test would be vacuous")
	}

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})
	cbs.scheme = scheme
	cbs.SetEnabled(0, false)

	buf := NewDrawBuffer(20, 2)
	cbs.Draw(buf)

	// '[' at (1,0) = disabled; '[' at (1,1) = enabled.
	disabledStyle := buf.GetCell(1, 0).Style
	enabledStyle := buf.GetCell(1, 1).Style

	if disabledStyle == enabledStyle {
		t.Errorf("disabled and enabled items render with same style %v; expected different styles", disabledStyle)
	}
}

// TestSetEnabledFalseDoesNotUncheckItem verifies that disabling an item does not
// change its checked state.
// Falsifies: an implementation that unchecks when disabling.
func TestSetEnabledFalseDoesNotUncheckItem(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})
	cbs.Item(0).SetChecked(true)

	cbs.SetEnabled(0, false)

	if !cbs.Item(0).Checked() {
		t.Error("SetEnabled(0, false) unchecked item 0; disabling must not change checked state")
	}
}

// TestSetEnabledIndependentOfCheckedState verifies enabling/disabling and
// checking/unchecking are orthogonal.
// Falsifies: linking enabled state to checked state.
func TestSetEnabledIndependentOfCheckedState(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})

	// Disable, then check via SetValues.
	cbs.SetEnabled(0, false)
	cbs.Item(0).SetChecked(true) // must still be possible programmatically

	if !cbs.Item(0).Checked() {
		t.Error("programmatic SetChecked(true) on disabled item failed; enable state must be orthogonal to checked state")
	}
	if cbs.IsEnabled(0) {
		t.Error("SetChecked changed enabled state; they must be independent")
	}
}

// TestAllDisabledThenReEnabledRestoresSelectability verifies the full cycle:
// all-disabled (OfSelectable cleared) → re-enable one → OfSelectable restored.
// Falsifies: forgetting to restore OfSelectable.
func TestAllDisabledThenReEnabledRestoresSelectability(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	cbs.SetEnabled(0, false)
	cbs.SetEnabled(1, false)
	cbs.SetEnabled(2, false)

	// Verify OfSelectable was cleared.
	if cbs.HasOption(OfSelectable) {
		t.Fatal("precondition failed: all-disabled should clear OfSelectable")
	}

	// Re-enable one.
	cbs.SetEnabled(1, true)

	if !cbs.HasOption(OfSelectable) {
		t.Error("after re-enabling item 1, OfSelectable not restored; CheckBoxes unreachable by Tab")
	}
}

// TestMulticolumnColumnWidthDoesNotUseRawLabelLength verifies column width uses
// tilde-stripped label length, not raw string length.
// Falsifies: using len(label) instead of tildeTextLen(label).
func TestMulticolumnColumnWidthDoesNotUseRawLabelLength(t *testing.T) {
	// "~A~BC" raw length=5, tilde-stripped text="ABC" length=3.
	// colWidth should be 3+6=9, not 5+6=11.
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 1), []string{"~A~BC"})

	want := tildeTextLen("~A~BC") + 6 // 3 + 6 = 9
	got := cbs.Item(0).Bounds().Width()

	if got != want {
		t.Errorf("colWidth for '~A~BC' = %d, want %d (must use tilde-stripped length)", got, want)
	}
}
