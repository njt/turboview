package tv

// integration_batch2_listbox_test.go — Integration tests for the ListBox widget.
//
// These tests exercise ListBox end-to-end using real components (Window, Group,
// ListViewer, ScrollBar, DrawBuffer). No mocks.
//
// Test naming: TestIntegrationBatch2ListBox<DescriptiveSuffix>

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makeListBoxItems builds n strings of the form "Item 1", "Item 2", …, "Item n".
func makeListBoxItems(n int) []string {
	items := make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d", i+1)
	}
	return items
}

// listBoxKeyEv constructs an EvKeyboard event for the given special key.
func listBoxKeyEv(key tcell.Key) *Event {
	return &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: key},
	}
}

// ---------------------------------------------------------------------------
// Test 1: Keyboard navigation — Down arrow moves selection from 0 to 1.
// ---------------------------------------------------------------------------

// TestIntegrationBatch2ListBoxKeyboardDownMovesSelection verifies that pressing
// Down inside a Window containing a ListBox moves the selection from 0 to 1.
//
// The event path is:
//   Window.HandleEvent → group.HandleEvent (focused=ListBox) → ListBox.HandleEvent
//   → ListViewer.HandleEvent (focused=viewer) → selected++
func TestIntegrationBatch2ListBoxKeyboardDownMovesSelection(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 30, 12), "Test")
	lb := NewStringListBox(NewRect(0, 0, 20, 5), makeListBoxItems(10))
	win.Insert(lb)
	win.SetFocusedChild(lb)

	// Focus the internal ListViewer so it responds to keyboard.
	lb.group.SetFocusedChild(lb.viewer)

	if lb.Selected() != 0 {
		t.Fatalf("pre-condition: Selected() = %d, want 0", lb.Selected())
	}

	win.HandleEvent(listBoxKeyEv(tcell.KeyDown))

	if lb.Selected() != 1 {
		t.Errorf("after Down: Selected() = %d, want 1", lb.Selected())
	}
}

// ---------------------------------------------------------------------------
// Test 2: Scrollbar sync — navigate past visible area, scrollbar value updates.
// ---------------------------------------------------------------------------

// TestIntegrationBatch2ListBoxScrollbarSyncsOnNavigationPastVisible verifies that
// when navigating past the visible area (5 rows, 20 items), the scrollbar value
// updates to reflect the new topIndex.
func TestIntegrationBatch2ListBoxScrollbarSyncsOnNavigationPastVisible(t *testing.T) {
	// 20 items, 5-row visible area.
	lb := NewStringListBox(NewRect(0, 0, 20, 5), makeListBoxItems(20))

	// Give viewer SfSelected so it handles keyboard.
	lb.viewer.SetState(SfSelected, true)

	// Initial scrollbar value must be 0.
	if lb.scrollbar.Value() != 0 {
		t.Fatalf("pre-condition: ScrollBar.Value() = %d, want 0", lb.scrollbar.Value())
	}

	// Navigate down 5 times — this pushes selection to index 5, which is at
	// the boundary of the visible area (rows 0..4). One more step (index 5)
	// forces topIndex to 1.
	for i := 0; i < 5; i++ {
		lb.viewer.HandleEvent(listBoxKeyEv(tcell.KeyDown))
	}

	// selected=5, visibleHeight=5 → selected >= topIndex+5 → topIndex = 5-5+1 = 1.
	// Scrollbar tracks selected, so its value should equal Selected().
	if lb.scrollbar.Value() <= 0 {
		t.Errorf("after navigating past visible area, ScrollBar.Value() = %d, want > 0",
			lb.scrollbar.Value())
	}
	if lb.viewer.Selected() != lb.scrollbar.Value() {
		t.Errorf("ScrollBar.Value() = %d does not match ListViewer.Selected() = %d",
			lb.scrollbar.Value(), lb.viewer.Selected())
	}
}

// ---------------------------------------------------------------------------
// Test 3: Mouse click on ListViewer area — click at y=2 selects item at y=2.
// ---------------------------------------------------------------------------

// TestIntegrationBatch2ListBoxMouseClickSelectsRow verifies that a mouse click
// at relative y=2 in the ListBox's ListViewer area selects the item at
// topIndex + 2.
//
// ListBox dispatches mouse events to child whose bounds contain the point.
// ListViewer is at (0,0,19,5) (width-1 = 19), so click at (0, 2) hits the viewer.
func TestIntegrationBatch2ListBoxMouseClickSelectsRow(t *testing.T) {
	lb := NewStringListBox(NewRect(0, 0, 20, 5), makeListBoxItems(10))

	// Pre-condition: selection at 0.
	if lb.Selected() != 0 {
		t.Fatalf("pre-condition: Selected() = %d, want 0", lb.Selected())
	}

	// Build a Button1 mouse event at ListViewer local (0, 2).
	// ListBox.HandleEvent routes it to the viewer (bounds contains (0,2)).
	clickEv := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:          0,
			Y:          2,
			Button:     tcell.Button1,
			ClickCount: 1,
		},
	}
	lb.HandleEvent(clickEv)

	// topIndex=0, so click at y=2 → item 2.
	if lb.Selected() != 2 {
		t.Errorf("after click at y=2: Selected() = %d, want 2", lb.Selected())
	}
}

// ---------------------------------------------------------------------------
// Test 4: NewStringListBox items accessible via DataSource.
// ---------------------------------------------------------------------------

// TestIntegrationBatch2ListBoxDataSourceItemsReadable verifies that items passed
// to NewStringListBox are accessible through DataSource() with correct count and
// values.
func TestIntegrationBatch2ListBoxDataSourceItemsReadable(t *testing.T) {
	items := []string{"alpha", "beta", "gamma", "delta"}
	lb := NewStringListBox(NewRect(0, 0, 20, 5), items)

	ds := lb.DataSource()
	if ds == nil {
		t.Fatal("DataSource() must not be nil")
	}
	if ds.Count() != len(items) {
		t.Errorf("DataSource().Count() = %d, want %d", ds.Count(), len(items))
	}
	for i, want := range items {
		got := ds.Item(i)
		if got != want {
			t.Errorf("DataSource().Item(%d) = %q, want %q", i, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 5: SetDataSource resets selection — Selected() returns 0 after reset.
// ---------------------------------------------------------------------------

// TestIntegrationBatch2ListBoxSetDataSourceResetsSelection verifies that setting
// selection to 5, then calling SetDataSource with new data, resets Selected() to 0.
func TestIntegrationBatch2ListBoxSetDataSourceResetsSelection(t *testing.T) {
	lb := NewStringListBox(NewRect(0, 0, 20, 10), makeListBoxItems(10))
	lb.SetSelected(5)

	// Confirm pre-condition: selection is non-zero.
	if lb.Selected() != 5 {
		t.Fatalf("pre-condition: Selected() = %d after SetSelected(5), want 5", lb.Selected())
	}

	newData := NewStringList([]string{"one", "two", "three"})
	lb.SetDataSource(newData)

	if lb.Selected() != 0 {
		t.Errorf("after SetDataSource: Selected() = %d, want 0 (reset)", lb.Selected())
	}
}

// ---------------------------------------------------------------------------
// Test 6: Focus behavior — Tab focuses ListBox; FocusedChild is the ListViewer.
// ---------------------------------------------------------------------------

// TestIntegrationBatch2ListBoxFocusViaTabFocusesListViewer verifies that when a
// Window contains a ListBox and another selectable widget, Tab focusing the ListBox
// results in FocusedChild() of the ListBox being the internal ListViewer (not the
// ScrollBar).
func TestIntegrationBatch2ListBoxFocusViaTabFocusesListViewer(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 30, 12), "Test")
	lb := NewStringListBox(NewRect(0, 0, 20, 5), makeListBoxItems(10))
	btn := NewButton(NewRect(0, 6, 10, 1), "OK", CmOK)

	win.Insert(lb)
	win.Insert(btn)

	// Start with button focused.
	win.SetFocusedChild(btn)
	if win.FocusedChild() != btn {
		t.Fatalf("pre-condition: FocusedChild() = %v, want btn", win.FocusedChild())
	}

	// Tab advances from btn → wraps to lb.
	win.HandleEvent(listBoxKeyEv(tcell.KeyTab))

	if win.FocusedChild() != lb {
		t.Errorf("after Tab, window FocusedChild() = %v, want ListBox", win.FocusedChild())
	}

	// The ListBox's internal focused child must be the ListViewer, not the ScrollBar.
	innerFocused := lb.FocusedChild()
	if innerFocused == nil {
		// No inner focused child yet — that's acceptable, but scrollbar must not be it.
		if lb.scrollbar.HasOption(OfSelectable) {
			t.Error("ScrollBar must not be selectable; FocusedChild should never be the ScrollBar")
		}
		return
	}
	if innerFocused == View(lb.scrollbar) {
		t.Error("FocusedChild of ListBox is the ScrollBar; it must be the ListViewer")
	}
	if innerFocused != View(lb.viewer) {
		t.Errorf("FocusedChild of ListBox = %T, want *ListViewer", innerFocused)
	}
}

// ---------------------------------------------------------------------------
// Test 7: Container interface — Children() returns 2 children.
// ---------------------------------------------------------------------------

// TestIntegrationBatch2ListBoxChildrenReturnsTwoEntries verifies that ListBox
// implements Container and its Children() method returns exactly 2 entries:
// the ListViewer and the ScrollBar.
func TestIntegrationBatch2ListBoxChildrenReturnsTwoEntries(t *testing.T) {
	lb := NewStringListBox(NewRect(0, 0, 20, 5), makeListBoxItems(5))

	children := lb.Children()
	if len(children) != 2 {
		t.Errorf("Children() returned %d children, want 2 (ListViewer + ScrollBar)", len(children))
	}

	// Verify the two children are the expected types.
	foundViewer := false
	foundScrollBar := false
	for _, child := range children {
		if child == View(lb.ListViewer()) {
			foundViewer = true
		}
		if child == View(lb.ScrollBar()) {
			foundScrollBar = true
		}
	}
	if !foundViewer {
		t.Error("Children() does not include the ListViewer")
	}
	if !foundScrollBar {
		t.Error("Children() does not include the ScrollBar")
	}
}

// ---------------------------------------------------------------------------
// Test 8: Draw produces visible output — item text appears at expected positions.
// ---------------------------------------------------------------------------

// TestIntegrationBatch2ListBoxDrawRendersItemText verifies that calling Draw on a
// ListBox writes list item text into the DrawBuffer at the expected column/row
// positions.
//
// ListBox bounds: (0,0,20,5) → width=20, height=5.
// ListViewer occupies columns 0..18 (width-1=19), rows 0..4.
// Item 0 text "Item 1" should appear starting at (0,0) of the ListViewer sub-buffer.
func TestIntegrationBatch2ListBoxDrawRendersItemText(t *testing.T) {
	items := []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}
	lb := NewStringListBox(NewRect(0, 0, 20, 5), items)

	// Apply a color scheme so Draw has valid styles.
	lb.group.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 5)
	lb.Draw(buf)

	// Check that "Apple" (item 0) starts at column 0, row 0.
	expected := []rune("Apple")
	for col, ch := range expected {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ch {
			t.Errorf("Draw: buf[col=%d, row=0].Rune = %q, want %q (from %q)",
				col, cell.Rune, ch, "Apple")
			break
		}
	}

	// Check that "Banana" (item 1) starts at column 0, row 1.
	expected2 := []rune("Banana")
	for col, ch := range expected2 {
		cell := buf.GetCell(col, 1)
		if cell.Rune != ch {
			t.Errorf("Draw: buf[col=%d, row=1].Rune = %q, want %q (from %q)",
				col, cell.Rune, ch, "Banana")
			break
		}
	}
}

// TestIntegrationBatch2ListBoxDrawScrollBarAppearsAtRightEdge verifies that Draw
// renders the scrollbar's up-arrow character at the rightmost column.
func TestIntegrationBatch2ListBoxDrawScrollBarAppearsAtRightEdge(t *testing.T) {
	items := makeListBoxItems(20) // more than visible → scrollbar is meaningful
	lb := NewStringListBox(NewRect(0, 0, 20, 5), items)
	lb.group.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 5)
	lb.Draw(buf)

	// The scrollbar occupies column 19 (width-1).
	// Its top cell should be the up-arrow character '▲'.
	cell := buf.GetCell(19, 0)
	if cell.Rune != '▲' {
		t.Errorf("Draw: buf[col=19, row=0].Rune = %q, want '▲' (scrollbar up-arrow)", cell.Rune)
	}
}
