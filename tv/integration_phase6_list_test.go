package tv

// integration_phase6_list_test.go — Integration tests for Phase 6 Batch 2:
// ListViewer with ScrollBar inside a Window.
//
// Each test verifies one requirement from the Phase 6 plan using REAL components
// wired end-to-end: Window → ListViewer + ScrollBar.
//
// Test naming: TestIntegrationPhase6List<DescriptiveSuffix>.
//
// Requirements covered:
//   1. ScrollBar down-arrow click updates ListViewer topIndex via OnChange callback.
//   2. Keyboard navigation (Down, PgDn) on focused ListViewer updates bound ScrollBar value.
//   3. ListViewer with more items than visible height shows correct items after scroll.
//   4. ListViewer inside Window with custom ColorScheme renders with scheme's List styles.
//   5. Tab between ListViewer and Button inside same Window changes focus correctly.
//   6. Mouse click on a visible ListViewer row inside a Window selects that item.

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makeItems builds a slice of n strings "Item 1", "Item 2", … "Item n".
func makeItems(n int) []string {
	items := make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d", i+1)
	}
	return items
}

// makeListWindow creates a Window(0,0,30,12) with a ListViewer(0,0,27,8) and a
// vertical ScrollBar(27,0,1,8) wired together.  The ListViewer is focused.
// Returns all three so tests can inspect and interact with them.
func makeListWindow(items []string) (*Window, *ListViewer, *ScrollBar) {
	win := NewWindow(NewRect(0, 0, 30, 12), "List", WithWindowNumber(1))
	lv := NewListViewer(NewRect(0, 0, 27, 8), NewStringList(items))
	sb := NewScrollBar(NewRect(27, 0, 1, 8), Vertical)
	lv.SetScrollBar(sb)
	win.Insert(lv)
	win.Insert(sb)
	win.SetFocusedChild(lv)
	return win, lv, sb
}

// ---------------------------------------------------------------------------
// Requirement 1: ScrollBar down-arrow click updates ListViewer topIndex.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6ListScrollBarDownArrowUpdatesTopIndex verifies that a
// mouse click on the ScrollBar's down arrow — routed through the Window's client
// area dispatch chain — updates the ListViewer's selected via the OnChange
// callback that SetScrollBar installs.
//
// syncScrollBar sets arStep = 1 for single-column, so one down-arrow click
// steps selected by 1.
//
// Layout (window-local coordinates):
//   - Client area starts at (1,1) in window coords (1-cell frame).
//   - ScrollBar is at (27,0) in client → x=28 in window coords.
//   - ScrollBar height 8, down arrow at y=7 in SB local → client y=7 → window y=8.
func TestIntegrationPhase6ListScrollBarDownArrowUpdatesTopIndex(t *testing.T) {
	// Use 20 items so there is content to scroll.
	win, lv, sb := makeListWindow(makeItems(20))

	// Preconditions: selected must start at 0.
	if lv.Selected() != 0 {
		t.Fatalf("pre-condition: ListViewer.Selected() = %d, want 0", lv.Selected())
	}

	// To route the mouse click to the ScrollBar, the ScrollBar must be the focused
	// child of the window's group (Group forwards mouse events to the focused child).
	win.SetFocusedChild(sb)

	// ScrollBar is at client-area position (27,0).  Down arrow is at scrollbar-local
	// y = height-1 = 7.  In client coordinates: x=27, y=7.  In window coords
	// (add 1 for frame): x=28, y=8.
	downArrowEv := clickEvent(28, 8, tcell.Button1)
	win.HandleEvent(downArrowEv)

	// SetScrollBar wires sb.OnChange to call lv.SetSelected(val).
	// syncScrollBar sets arStep = 1 for single-column, so one step-down moves by 1.
	if lv.Selected() != 1 {
		t.Errorf("after ScrollBar down-arrow click, ListViewer.Selected() = %d, want 1 (arStep=1)", lv.Selected())
	}

	// ScrollBar value should also be 1 (tracks selected).
	if sb.Value() != 1 {
		t.Errorf("after ScrollBar down-arrow click, ScrollBar.Value() = %d, want 1", sb.Value())
	}
}

// TestIntegrationPhase6ListScrollBarMultipleDownArrowsAccumulate verifies that
// three down-arrow clicks accumulate correctly.
// syncScrollBar sets arStep = 1 for single-column, so each click steps by 1.
// Range is [0, 19] (count-1). 3 clicks from 0 → selected=3.
func TestIntegrationPhase6ListScrollBarMultipleDownArrowsAccumulate(t *testing.T) {
	win, lv, sb := makeListWindow(makeItems(20))
	win.SetFocusedChild(sb)

	// Click down arrow 3 times (each steps by arStep=1).
	for i := 0; i < 3; i++ {
		win.HandleEvent(clickEvent(28, 8, tcell.Button1))
	}

	if lv.Selected() != 3 {
		t.Errorf("after 3 ScrollBar down-arrow clicks, Selected() = %d, want 3", lv.Selected())
	}
	if sb.Value() != 3 {
		t.Errorf("after 3 ScrollBar down-arrow clicks, ScrollBar.Value() = %d, want 3", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: Keyboard navigation updates the bound ScrollBar value.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6ListKeyboardDownUpdatesBoundScrollBar verifies that
// pressing the Down arrow key while a ListViewer is focused inside a Window
// advances the selection and syncs the ScrollBar (which tracks selected).
func TestIntegrationPhase6ListKeyboardDownUpdatesBoundScrollBar(t *testing.T) {
	win, lv, sb := makeListWindow(makeItems(20))
	// ListViewer is already focused (makeListWindow sets it).

	// Initial state: selected=0, scrollBar value=0.
	if lv.Selected() != 0 {
		t.Fatalf("pre-condition: Selected() = %d, want 0", lv.Selected())
	}

	// Press Down — event sent to Window, routed to focused ListViewer.
	win.HandleEvent(listKeyEv(tcell.KeyDown))

	if lv.Selected() != 1 {
		t.Errorf("after Down, Selected() = %d, want 1", lv.Selected())
	}

	// ScrollBar tracks selected, so value should be 1.
	if sb.Value() != 1 {
		t.Errorf("after Down, ScrollBar.Value() = %d, want 1 (tracks selected)", sb.Value())
	}

	// Press Down 7 more times to move selection to index 8 (beyond visible range).
	for i := 0; i < 7; i++ {
		win.HandleEvent(listKeyEv(tcell.KeyDown))
	}
	// selected=8, scrollbar should track selected.
	if lv.Selected() != 8 {
		t.Errorf("after 8 Down presses, Selected() = %d, want 8", lv.Selected())
	}
	if sb.Value() != 8 {
		t.Errorf("after 8 Down presses, ScrollBar.Value() = %d, want 8 (tracks selected)", sb.Value())
	}
}

// TestIntegrationPhase6ListKeyboardPgDnUpdatesBoundScrollBar verifies that
// PgDn advances selection by visibleHeight and syncs the ScrollBar (which tracks selected).
func TestIntegrationPhase6ListKeyboardPgDnUpdatesBoundScrollBar(t *testing.T) {
	win, lv, sb := makeListWindow(makeItems(20))

	// PgDn from selected=0 should jump to index 8 (0 + visibleHeight=8).
	win.HandleEvent(listKeyEv(tcell.KeyPgDn))

	if lv.Selected() != 8 {
		t.Errorf("after PgDn, Selected() = %d, want 8", lv.Selected())
	}
	// topIndex should scroll so that index 8 is visible at the bottom:
	// ensureVisible: selected(8) >= topIndex(0)+vh(8) → topIndex = 8-8+1 = 1.
	if lv.TopIndex() != 1 {
		t.Errorf("after PgDn, TopIndex() = %d, want 1", lv.TopIndex())
	}
	// ScrollBar tracks selected, so value should be 8.
	if sb.Value() != 8 {
		t.Errorf("after PgDn, ScrollBar.Value() = %d, want 8 (tracks selected)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: ListViewer shows correct items after keyboard scroll.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6ListShowsCorrectItemsAfterKeyboardScroll verifies that
// after scrolling via keyboard, Draw renders the items at the updated topIndex.
func TestIntegrationPhase6ListShowsCorrectItemsAfterKeyboardScroll(t *testing.T) {
	items := makeItems(20)
	win, lv, _ := makeListWindow(items)

	// Scroll down 5 positions using PgDn + Down presses.
	// After PgDn: selected=8, topIndex=1.
	win.HandleEvent(listKeyEv(tcell.KeyPgDn))
	// After PgDn again: selected=16, topIndex = 16-8+1 = 9.
	win.HandleEvent(listKeyEv(tcell.KeyPgDn))

	topIdx := lv.TopIndex()
	if topIdx < 1 {
		t.Fatalf("expected topIndex > 0 after two PgDns, got %d", topIdx)
	}

	// Draw into a buffer wide enough for the window.
	// Window is (0,0,30,12): client area is 28 wide, 10 tall.
	buf := NewDrawBuffer(30, 12)
	win.Draw(buf)

	// The ListViewer occupies client rows 0..7, client cols 0..26.
	// In window-buffer coordinates: row r of LV = buf row (1+r), col c = buf col (1+c).
	//
	// The first visible item should be items[topIdx].
	// We check that the first few runes of items[topIdx] appear at row (1+0) = 1.
	expectedText := items[topIdx]
	for col, ch := range []rune(expectedText) {
		if col >= 27 { // LV width
			break
		}
		cell := buf.GetCell(1+col, 1)
		if cell.Rune != ch {
			t.Errorf("after scroll, buf[col=%d, row=1].Rune = %q, want %q (from %q)",
				1+col, cell.Rune, ch, expectedText)
			break
		}
	}

	// Row 1 (second visible row) should show items[topIdx+1].
	if topIdx+1 < len(items) {
		expectedText2 := items[topIdx+1]
		for col, ch := range []rune(expectedText2) {
			if col >= 27 {
				break
			}
			cell := buf.GetCell(1+col, 2)
			if cell.Rune != ch {
				t.Errorf("after scroll, buf[col=%d, row=2].Rune = %q, want %q (from %q)",
					1+col, cell.Rune, ch, expectedText2)
				break
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Custom ColorScheme renders with scheme's List styles.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6ListColorSchemeListNormalApplied verifies that a
// ListViewer inside a Window with a custom ColorScheme renders unselected items
// using that scheme's ListNormal style.
func TestIntegrationPhase6ListColorSchemeListNormalApplied(t *testing.T) {
	customScheme := &theme.ColorScheme{}
	*customScheme = *theme.BorlandBlue
	customScheme.ListNormal = tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorNavy)
	customScheme.ListSelected = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlue)
	customScheme.ListFocused = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed)

	win := NewWindow(NewRect(0, 0, 30, 12), "ColorScheme Test")
	win.SetColorScheme(customScheme)

	items := makeItems(10)
	lv := NewListViewer(NewRect(0, 0, 27, 8), NewStringList(items))
	sb := NewScrollBar(NewRect(27, 0, 1, 8), Vertical)
	lv.SetScrollBar(sb)
	win.Insert(lv)
	win.Insert(sb)
	win.SetFocusedChild(lv)

	// Verify the ListViewer inherits the scheme.
	cs := lv.ColorScheme()
	if cs == nil {
		t.Fatal("ListViewer.ColorScheme() returned nil; expected inheritance from Window")
	}
	if cs != customScheme {
		t.Errorf("ListViewer.ColorScheme() != customScheme; scheme inheritance broken")
	}

	// Draw the window.
	buf := NewDrawBuffer(30, 12)
	win.Draw(buf)

	// The focused child (lv) has SfSelected set, so selected item (index 0) uses ListFocused.
	// All other rows use ListNormal.
	// Check a non-selected row — row 1 of LV (buf row 2) should use ListNormal.
	cell := buf.GetCell(1, 2) // col=1 (first char of LV at client col 0), row=2 (LV row 1)
	if cell.Style != customScheme.ListNormal {
		t.Errorf("non-selected row: cell style = %v, want customScheme.ListNormal %v",
			cell.Style, customScheme.ListNormal)
	}
}

// TestIntegrationPhase6ListColorSchemeListFocusedApplied verifies that the
// focused+selected row uses the scheme's ListFocused style.
func TestIntegrationPhase6ListColorSchemeListFocusedApplied(t *testing.T) {
	customScheme := &theme.ColorScheme{}
	*customScheme = *theme.BorlandBlue
	customScheme.ListNormal = tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorNavy)
	customScheme.ListFocused = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed)

	win := NewWindow(NewRect(0, 0, 30, 12), "ColorScheme Test")
	win.SetColorScheme(customScheme)

	items := makeItems(10)
	lv := NewListViewer(NewRect(0, 0, 27, 8), NewStringList(items))
	win.Insert(lv)
	win.SetFocusedChild(lv)

	buf := NewDrawBuffer(30, 12)
	win.Draw(buf)

	// The selected item is index 0 and the LV has SfSelected (focused) → ListFocused style.
	// In the buf: LV row 0 = buf row 1 (frame + client offset).
	// The selected row is filled with spaces in the focused style before text is drawn.
	// Check a cell within the row past the text length or check the style of a text cell.
	// "Item 1" is 6 chars; check cell at col 7 (within LV width, past the text).
	cell := buf.GetCell(8, 1) // buf col 8 = LV client col 7, buf row 1 = LV row 0
	if cell.Style != customScheme.ListFocused {
		t.Errorf("focused+selected row (blank area): style = %v, want ListFocused %v",
			cell.Style, customScheme.ListFocused)
	}
}

// TestIntegrationPhase6ListColorSchemeListSelectedApplied verifies that when the
// ListViewer is NOT focused (SfSelected cleared), the selected row uses ListSelected.
func TestIntegrationPhase6ListColorSchemeListSelectedApplied(t *testing.T) {
	customScheme := &theme.ColorScheme{}
	*customScheme = *theme.BorlandBlue
	customScheme.ListNormal = tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorNavy)
	customScheme.ListSelected = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlue)
	customScheme.ListFocused = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed)

	win := NewWindow(NewRect(0, 0, 30, 12), "ColorScheme Test")
	win.SetColorScheme(customScheme)

	items := makeItems(10)
	lv := NewListViewer(NewRect(0, 0, 27, 8), NewStringList(items))
	win.Insert(lv)

	// Do NOT call SetFocusedChild(lv) — it is not focused, so SfSelected is not set.
	// However Insert() calls selectChild if OfSelectable is set, so we need to
	// explicitly unfocus it.
	lv.SetState(SfSelected, false)

	buf := NewDrawBuffer(30, 12)
	win.Draw(buf)

	// selected=0, not focused → style = ListSelected.
	// buf col 8, row 1 (within the padding of the selected row).
	cell := buf.GetCell(8, 1)
	if cell.Style != customScheme.ListSelected {
		t.Errorf("unfocused selected row: style = %v, want ListSelected %v",
			cell.Style, customScheme.ListSelected)
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: Tab between ListViewer and Button changes focus correctly.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6ListTabFromListViewerToButton verifies that Tab sent to
// a Window moves focus from a ListViewer to a Button.
func TestIntegrationPhase6ListTabFromListViewerToButton(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 30, 12), "Tab Test")

	lv := NewListViewer(NewRect(0, 0, 20, 6), NewStringList(makeItems(10)))
	btn := NewButton(NewRect(0, 7, 10, 1), "OK", CmOK)

	win.Insert(lv)
	win.Insert(btn)
	win.SetFocusedChild(lv)

	if win.FocusedChild() != lv {
		t.Fatal("pre-condition: ListViewer should be focused initially")
	}

	// Tab → should advance focus to Button.
	win.HandleEvent(tabKey())

	if win.FocusedChild() != btn {
		t.Errorf("after Tab, FocusedChild() = %v, want Button", win.FocusedChild())
	}

	// ListViewer should no longer have SfSelected.
	if lv.HasState(SfSelected) {
		t.Error("after Tab, ListViewer still has SfSelected set")
	}

	// Button should have SfSelected.
	if !btn.HasState(SfSelected) {
		t.Error("after Tab, Button should have SfSelected set")
	}
}

// TestIntegrationPhase6ListTabFromButtonBackToListViewer verifies that a second
// Tab (or Shift+Tab) from the Button returns focus to the ListViewer.
func TestIntegrationPhase6ListTabFromButtonBackToListViewer(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 30, 12), "Tab Test")

	lv := NewListViewer(NewRect(0, 0, 20, 6), NewStringList(makeItems(10)))
	btn := NewButton(NewRect(0, 7, 10, 1), "OK", CmOK)

	win.Insert(lv)
	win.Insert(btn)
	win.SetFocusedChild(lv)

	// Tab forward: lv → btn.
	win.HandleEvent(tabKey())
	if win.FocusedChild() != btn {
		t.Fatal("pre-condition: Button should be focused after first Tab")
	}

	// Tab forward again: btn → lv (wraps).
	win.HandleEvent(tabKey())
	if win.FocusedChild() != lv {
		t.Errorf("after second Tab (wrap), FocusedChild() = %v, want ListViewer", win.FocusedChild())
	}
}

// TestIntegrationPhase6ListShiftTabGoesBackward verifies that Shift+Tab goes
// backward in focus order inside the Window.
func TestIntegrationPhase6ListShiftTabGoesBackward(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 30, 12), "ShiftTab Test")

	lv := NewListViewer(NewRect(0, 0, 20, 6), NewStringList(makeItems(10)))
	btn := NewButton(NewRect(0, 7, 10, 1), "OK", CmOK)

	win.Insert(lv)
	win.Insert(btn)

	// Start with focus on Button (last inserted selectable).
	win.SetFocusedChild(btn)
	if win.FocusedChild() != btn {
		t.Fatal("pre-condition: Button should be focused")
	}

	// Shift+Tab from Button → should go to ListViewer (previous selectable).
	win.HandleEvent(shiftTabKey())

	if win.FocusedChild() != lv {
		t.Errorf("after Shift+Tab from Button, FocusedChild() = %v, want ListViewer", win.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: Mouse click on a visible ListViewer row selects that item.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6ListMouseClickSelectsRow verifies that a mouse click on
// a visible row in a ListViewer inside a Window selects the correct item.
//
// Coordinate math (window-local):
//   - Frame is 1 cell, so client area starts at (1,1).
//   - ListViewer is at client (0,0), so LV row r = window y (1+r).
//   - Clicking window-local (1, 1) = LV row 0 = item at topIndex+0.
//   - Clicking window-local (1, 3) = LV row 2 = item at topIndex+2.
func TestIntegrationPhase6ListMouseClickSelectsRow(t *testing.T) {
	items := makeItems(10)
	win, lv, _ := makeListWindow(items)

	// Pre-condition: selected = 0.
	if lv.Selected() != 0 {
		t.Fatalf("pre-condition: Selected() = %d, want 0", lv.Selected())
	}

	// Click on row 2 of the ListViewer (window-local y=3, x=1).
	// The group routes mouse events to the focused child (lv).
	// Window.handleMouseEvent adjusts: clientX = x-1 = 0, clientY = y-1 = 2.
	// Group forwards to focused child (lv) with those coords.
	// ListViewer HandleEvent: clickIdx = topIndex(0) + Y(2) = 2.
	win.HandleEvent(clickEvent(1, 3, tcell.Button1))

	if lv.Selected() != 2 {
		t.Errorf("after click on row 2, Selected() = %d, want 2", lv.Selected())
	}
}

// TestIntegrationPhase6ListMouseClickOnFirstRowSelectsFirstItem verifies clicking
// the first row of the ListViewer.
func TestIntegrationPhase6ListMouseClickOnFirstRowSelectsFirstItem(t *testing.T) {
	items := makeItems(10)
	win, lv, _ := makeListWindow(items)

	// Start with selection at 5 to confirm a click on row 0 changes it.
	lv.SetSelected(5)

	// Click row 0: window-local (1, 1).
	win.HandleEvent(clickEvent(1, 1, tcell.Button1))

	if lv.Selected() != 0 {
		t.Errorf("after click on row 0, Selected() = %d, want 0", lv.Selected())
	}
}

// TestIntegrationPhase6ListMouseClickFiresOnSelect verifies that double-clicking a row
// triggers the OnSelect callback with the correct index (spec 7.3).
func TestIntegrationPhase6ListMouseClickFiresOnSelect(t *testing.T) {
	items := makeItems(10)
	win, lv, _ := makeListWindow(items)

	selectFired := false
	selectIdx := -1
	lv.OnSelect = func(idx int) {
		selectFired = true
		selectIdx = idx
	}

	// Double-click row 4: window-local (1, 5).
	win.HandleEvent(doubleClickEvent(1, 5, tcell.Button1))

	if !selectFired {
		t.Error("OnSelect was not called after mouse double-click through Window dispatch chain")
	}
	if selectIdx != 4 {
		t.Errorf("OnSelect called with index %d, want 4", selectIdx)
	}
	if lv.Selected() != 4 {
		t.Errorf("after double-click on row 4, Selected() = %d, want 4", lv.Selected())
	}
}

// TestIntegrationPhase6ListMouseClickAfterScrollSelectsCorrectItem verifies that
// after keyboard scrolling shifts topIndex, a mouse click still maps to the
// correct absolute item index (topIndex + row).
func TestIntegrationPhase6ListMouseClickAfterScrollSelectsCorrectItem(t *testing.T) {
	items := makeItems(20)
	win, lv, _ := makeListWindow(items)

	// Scroll down so topIndex = 5 by pressing Down 5+7=12 times (after 8th Down, topIndex=1;
	// continue until topIndex=5).
	// Simpler: use SetSelected to force topIndex.
	lv.SetSelected(12) // selected=12, topIndex = 12-8+1 = 5
	if lv.TopIndex() != 5 {
		t.Fatalf("pre-condition: TopIndex() = %d, want 5 after SetSelected(12)", lv.TopIndex())
	}

	// Now click row 0 in the LV (window-local (1,1)).
	// Expected: clickIdx = topIndex(5) + Y(0) = 5.
	win.HandleEvent(clickEvent(1, 1, tcell.Button1))

	if lv.Selected() != 5 {
		t.Errorf("after scroll+click row 0, Selected() = %d, want 5 (topIndex+0)", lv.Selected())
	}

	// Click row 3 (window-local (1, 4)).
	// Expected: clickIdx = topIndex(5) + Y(3) = 8.
	win.HandleEvent(clickEvent(1, 4, tcell.Button1))

	if lv.Selected() != 8 {
		t.Errorf("after scroll+click row 3, Selected() = %d, want 8 (topIndex+3)", lv.Selected())
	}
}
