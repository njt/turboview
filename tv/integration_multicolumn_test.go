package tv

// integration_multicolumn_test.go — Integration tests for multi-column CheckBoxes
// and RadioButtons in real Container hierarchies (Window, Group) with focus,
// keyboard, and mouse interaction.
//
// Test organisation:
//   Section 1 — CheckBoxes column layout in a 6-item / height-3 cluster
//   Section 2 — CheckBoxes arrow navigation (left/right across columns, down within column)
//   Section 3 — CheckBoxes navigation skips disabled items
//   Section 4 — CheckBoxes mouse interaction on disabled items
//   Section 5 — CheckBoxes inside a Window — keyboard through three-phase dispatch
//   Section 6 — RadioButtons column layout (4 items / height 2)
//   Section 7 — RadioButtons Right arrow moves selection (not just focus)
//   Section 8 — RadioButtons Alt+shortcut on disabled item is ignored
//   Section 9 — SetBounds relayout (height change triggers column change)

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Section 1 — CheckBoxes: 6 items / height 3 → 2 columns
// ---------------------------------------------------------------------------

// TestIntegrationMultiColCBSixItemsHeight3TwoColumns verifies that a CheckBoxes
// with 6 items and bounds height=3 produces 2 columns, with items 0-2 in column 0
// and items 3-5 in column 1.
func TestIntegrationMultiColCBSixItemsHeight3TwoColumns(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 60, 3), []string{"A", "B", "C", "D", "E", "F"})

	// Column 0 contains items 0, 1, 2 (indices 0..2 with height=3: i/3 == 0)
	for i := 0; i < 3; i++ {
		item := cbs.Item(i)
		if item.Bounds().A.X != 0 {
			t.Errorf("item %d: x = %d, want 0 (column 0)", i, item.Bounds().A.X)
		}
		if item.Bounds().A.Y != i {
			t.Errorf("item %d: y = %d, want %d (row %d)", i, item.Bounds().A.Y, i, i)
		}
	}

	// Column 1 contains items 3, 4, 5 (i/3 == 1)
	col1X := cbs.Item(3).Bounds().A.X
	if col1X == 0 {
		t.Fatal("item 3: x = 0, want > 0 (must be in column 1)")
	}
	for i := 3; i <= 5; i++ {
		item := cbs.Item(i)
		if item.Bounds().A.X != col1X {
			t.Errorf("item %d: x = %d, want %d (column 1 x)", i, item.Bounds().A.X, col1X)
		}
		wantRow := i - 3
		if item.Bounds().A.Y != wantRow {
			t.Errorf("item %d: y = %d, want %d (row in column 1)", i, item.Bounds().A.Y, wantRow)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 2 — CheckBoxes arrow navigation across columns
// ---------------------------------------------------------------------------

// TestIntegrationMultiColCBRightFromItem0FocusesItem3 verifies that Right arrow
// from item 0 (column 0, row 0) in a 6-item / height-3 CheckBoxes focuses
// item 3 (column 1, row 0).
func TestIntegrationMultiColCBRightFromItem0FocusesItem3(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 60, 3), []string{"A", "B", "C", "D", "E", "F"})
	cbs.SetState(SfSelected, true)
	cbs.SetFocusedChild(cbs.Item(0))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(3) {
		t.Errorf("Right from item 0: FocusedChild = %v, want Item(3)", cbs.FocusedChild())
	}
	if !ev.IsCleared() {
		t.Errorf("Right arrow event not cleared; ev.What = %v", ev.What)
	}
}

// TestIntegrationMultiColCBLeftFromItem3FocusesItem0 verifies that Left arrow
// from item 3 (column 1, row 0) in a 6-item / height-3 CheckBoxes focuses
// item 0 (column 0, row 0).
func TestIntegrationMultiColCBLeftFromItem3FocusesItem0(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 60, 3), []string{"A", "B", "C", "D", "E", "F"})
	cbs.SetState(SfSelected, true)
	cbs.SetFocusedChild(cbs.Item(3))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(0) {
		t.Errorf("Left from item 3: FocusedChild = %v, want Item(0)", cbs.FocusedChild())
	}
	if !ev.IsCleared() {
		t.Errorf("Left arrow event not cleared; ev.What = %v", ev.What)
	}
}

// TestIntegrationMultiColCBDownFromItem0FocusesItem1 verifies that Down arrow
// from item 0 focuses item 1 (same column, next row).
func TestIntegrationMultiColCBDownFromItem0FocusesItem1(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 60, 3), []string{"A", "B", "C", "D", "E", "F"})
	cbs.SetState(SfSelected, true)
	cbs.SetFocusedChild(cbs.Item(0))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(1) {
		t.Errorf("Down from item 0: FocusedChild = %v, want Item(1)", cbs.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — CheckBoxes navigation skips disabled items
// ---------------------------------------------------------------------------

// TestIntegrationMultiColCBDownSkipsDisabledItem1 verifies that when item 1 is
// disabled, Down from item 0 skips item 1 and lands on item 2.
func TestIntegrationMultiColCBDownSkipsDisabledItem1(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 60, 3), []string{"A", "B", "C", "D", "E", "F"})
	cbs.SetEnabled(1, false)
	cbs.SetState(SfSelected, true)
	cbs.SetFocusedChild(cbs.Item(0))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(2) {
		t.Errorf("Down with item 1 disabled: FocusedChild = %v, want Item(2)", cbs.FocusedChild())
	}
}

// TestIntegrationMultiColCBLeftColumnBlockedWhenTargetDisabled verifies that
// Left arrow from item 3 does not move when item 0 (the target in column 0,
// same row) is disabled.
func TestIntegrationMultiColCBLeftColumnBlockedWhenTargetDisabled(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 60, 3), []string{"A", "B", "C", "D", "E", "F"})
	cbs.SetEnabled(0, false) // target of Left from item 3
	cbs.SetState(SfSelected, true)
	cbs.SetFocusedChild(cbs.Item(3))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(3) {
		t.Errorf("Left to disabled item 0: FocusedChild = %v, want Item(3) (no move)", cbs.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Section 4 — CheckBoxes mouse interaction on disabled items
// ---------------------------------------------------------------------------

// TestIntegrationMultiColCBMouseClickOnDisabledItemDoesNotToggle verifies that a
// mouse click on a disabled item does not toggle its checked state.
func TestIntegrationMultiColCBMouseClickOnDisabledItemDoesNotToggle(t *testing.T) {
	// height=3, 6 items. Item 0 is at row 0, column 0.
	cbs := NewCheckBoxes(NewRect(0, 0, 60, 3), []string{"A", "B", "C", "D", "E", "F"})
	cbs.SetEnabled(0, false)

	// Click within item 0's bounds (row=0, col=0 area — x=2, y=0 is safe).
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 0, Button: tcell.Button1}}
	cbs.HandleEvent(ev)

	if cbs.Item(0).Checked() {
		t.Error("mouse click on disabled item 0 toggled it; disabled items must not respond to clicks")
	}
	// The event should be consumed to prevent further propagation.
	if !ev.IsCleared() {
		t.Error("click on disabled item was not consumed; it should be silently swallowed")
	}
}

// TestIntegrationMultiColCBMouseClickOnEnabledItemToggles verifies that a mouse
// click on an enabled item does toggle its checked state (sanity check).
func TestIntegrationMultiColCBMouseClickOnEnabledItemToggles(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 60, 3), []string{"A", "B", "C", "D", "E", "F"})
	// All items enabled by default.

	// Click within item 1's bounds: row=1, x within column 0.
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 1, Button: tcell.Button1}}
	cbs.HandleEvent(ev)

	if !cbs.Item(1).Checked() {
		t.Error("mouse click on enabled item 1 did not toggle it; enabled items must respond to clicks")
	}
}

// TestIntegrationMultiColCBMouseClickOnColumn1DisabledItemIgnored verifies that
// a mouse click on a disabled item in column 1 is also ignored.
func TestIntegrationMultiColCBMouseClickOnColumn1DisabledItemIgnored(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 60, 3), []string{"A", "B", "C", "D", "E", "F"})
	cbs.SetEnabled(3, false) // item 3 is in column 1, row 0

	col1X := cbs.Item(3).Bounds().A.X
	row := cbs.Item(3).Bounds().A.Y // = 0
	// Click at a position within item 3's bounds.
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: col1X + 2, Y: row, Button: tcell.Button1}}
	cbs.HandleEvent(ev)

	if cbs.Item(3).Checked() {
		t.Error("mouse click on disabled item 3 (column 1) toggled it; must be ignored")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — CheckBoxes inside a Window: keyboard through three-phase dispatch
// ---------------------------------------------------------------------------

// TestIntegrationMultiColCBInsideWindowReceivesKeyboard verifies that a CheckBoxes
// inside a Window receives keyboard events through the Window's three-phase dispatch.
// The Window delegates keyboard events to its internal Group, which uses three-phase
// dispatch. CheckBoxes has OfPostProcess set, so it receives events in phase 3.
func TestIntegrationMultiColCBInsideWindowReceivesKeyboard(t *testing.T) {
	// Create a Window large enough to hold a CheckBoxes widget.
	// Window at (0,0) width=40, height=8. Client area is (width-2=38) x (height-2=6).
	win := NewWindow(NewRect(0, 0, 40, 8), "Test")

	// Place CheckBoxes inside the window client area.
	// Client area origin is (0,0) in client-local coords.
	cbs := NewCheckBoxes(NewRect(1, 1, 30, 3), []string{"~A~lpha", "~B~eta", "Gamma", "~D~elta", "~E~psilon", "Zeta"})
	win.Insert(cbs)

	// Focus the window (normally the Desktop would do this, but here we do it
	// directly).
	win.SetState(SfSelected, true)
	// Focus the CheckBoxes inside the window.
	win.SetFocusedChild(cbs)
	cbs.SetState(SfSelected, true)
	// Focus item 0 inside the CheckBoxes.
	cbs.SetFocusedChild(cbs.Item(0))

	// Send a Down arrow key event to the Window; it should pass through to the
	// focused CheckBoxes via three-phase dispatch.
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	win.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(1) {
		t.Errorf("Down through Window dispatch: CheckBoxes FocusedChild = %v, want Item(1)", cbs.FocusedChild())
	}
	if !ev.IsCleared() {
		t.Errorf("Down arrow event not cleared after Window dispatch; ev.What = %v", ev.What)
	}
}

// TestIntegrationMultiColCBInsideWindowAltShortcutActivatesItem verifies that an
// Alt+shortcut event delivered through the Window activates the matching CheckBox
// item even when CheckBoxes is not the focused child of the Window.
func TestIntegrationMultiColCBInsideWindowAltShortcutActivatesItem(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 40, 8), "Test")
	cbs := NewCheckBoxes(NewRect(1, 1, 30, 3), []string{"~A~lpha", "~B~eta", "Gamma", "~D~elta", "~E~psilon", "Zeta"})
	win.Insert(cbs)

	// CheckBoxes is in the window but does NOT have focus (SfSelected=false).
	// CheckBoxes has OfPostProcess, so it receives events in phase 3.
	win.SetState(SfSelected, true)

	// Send Alt+a — should toggle item 0 ("~A~lpha") even without focus.
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'a', Modifiers: tcell.ModAlt}}
	win.HandleEvent(ev)

	if !cbs.Item(0).Checked() {
		t.Error("Alt+a through Window dispatch did not toggle item 0 (~A~lpha); Alt+shortcut must work in post-process phase")
	}
}

// TestIntegrationMultiColCBInsideWindowRightArrowMovesFocus verifies Right arrow
// delivered through a Window moves focus within the CheckBoxes from column 0 to
// column 1.
func TestIntegrationMultiColCBInsideWindowRightArrowMovesFocus(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 40, 8), "Test")
	cbs := NewCheckBoxes(NewRect(1, 1, 30, 3), []string{"A", "B", "C", "D", "E", "F"})
	win.Insert(cbs)

	win.SetState(SfSelected, true)
	win.SetFocusedChild(cbs)
	cbs.SetState(SfSelected, true)
	cbs.SetFocusedChild(cbs.Item(0))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	win.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(3) {
		t.Errorf("Right through Window: FocusedChild = %v, want Item(3)", cbs.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Section 6 — RadioButtons: 4 items / height 2 → 2 columns
// ---------------------------------------------------------------------------

// TestIntegrationMultiColRBFourItemsHeight2TwoColumns verifies that a RadioButtons
// with 4 items and bounds height=2 produces 2 columns: items 0,1 in column 0;
// items 2,3 in column 1.
func TestIntegrationMultiColRBFourItemsHeight2TwoColumns(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"Alpha", "Beta", "Gamma", "Delta"})

	// Column 0: items 0 and 1 (i/2 == 0)
	for i := 0; i < 2; i++ {
		if rbs.Item(i).Bounds().A.X != 0 {
			t.Errorf("item %d: x = %d, want 0 (column 0)", i, rbs.Item(i).Bounds().A.X)
		}
		if rbs.Item(i).Bounds().A.Y != i {
			t.Errorf("item %d: y = %d, want %d (row %d)", i, rbs.Item(i).Bounds().A.Y, i, i)
		}
	}

	// Column 1: items 2 and 3 (i/2 == 1)
	col1X := rbs.Item(2).Bounds().A.X
	if col1X == 0 {
		t.Fatal("item 2: x = 0, want > 0 (must be in column 1)")
	}
	for i := 2; i < 4; i++ {
		if rbs.Item(i).Bounds().A.X != col1X {
			t.Errorf("item %d: x = %d, want %d (column 1 x)", i, rbs.Item(i).Bounds().A.X, col1X)
		}
		wantRow := i - 2
		if rbs.Item(i).Bounds().A.Y != wantRow {
			t.Errorf("item %d: y = %d, want %d (row in column 1)", i, rbs.Item(i).Bounds().A.Y, wantRow)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 7 — RadioButtons: Right arrow moves selection AND focus
// ---------------------------------------------------------------------------

// TestIntegrationMultiColRBRightArrowMovesSelectionToNextColumn verifies that
// Right arrow in a RadioButtons (4 items, height 2) moves both focus AND selection
// to item 2 (column 1, row 0) when starting at item 0 (column 0, row 0).
func TestIntegrationMultiColRBRightArrowMovesSelectionToNextColumn(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"Alpha", "Beta", "Gamma", "Delta"})
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	// Focus must have moved to item 2 (column 1, row 0).
	if rbs.FocusedChild() != rbs.Item(2) {
		t.Errorf("Right arrow: FocusedChild = %v, want Item(2)", rbs.FocusedChild())
	}
	// Selection (Value) must also have moved — RadioButtons navigation moves selection.
	if rbs.Value() != 2 {
		t.Errorf("Right arrow: Value() = %d, want 2 (selection must follow focus)", rbs.Value())
	}
	// Item 0 must no longer be selected.
	if rbs.Item(0).Selected() {
		t.Error("Right arrow: Item(0) still selected; navigation must deselect previous item")
	}
	// Item 2 must now be selected.
	if !rbs.Item(2).Selected() {
		t.Error("Right arrow: Item(2) not selected; navigation must select new item")
	}
}

// TestIntegrationMultiColRBRightArrowIsNoOpAtLastColumn verifies that Right arrow
// at the last column in a 2-column RadioButtons is a no-op.
func TestIntegrationMultiColRBRightArrowIsNoOpAtLastColumn(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"Alpha", "Beta", "Gamma", "Delta"})
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 2 {
		t.Errorf("Right at last column: Value() = %d, want 2 (no-op)", rbs.Value())
	}
	if rbs.FocusedChild() != rbs.Item(2) {
		t.Errorf("Right at last column: FocusedChild = %v, want Item(2) (no move)", rbs.FocusedChild())
	}
}

// TestIntegrationMultiColRBLeftArrowMovesSelectionToPrevColumn verifies Left
// arrow from item 2 (column 1) moves selection back to item 0 (column 0).
func TestIntegrationMultiColRBLeftArrowMovesSelectionToPrevColumn(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"Alpha", "Beta", "Gamma", "Delta"})
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("Left from item 2: Value() = %d, want 0", rbs.Value())
	}
	if rbs.FocusedChild() != rbs.Item(0) {
		t.Errorf("Left from item 2: FocusedChild = %v, want Item(0)", rbs.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Section 8 — RadioButtons: Alt+shortcut on disabled item is ignored
// ---------------------------------------------------------------------------

// TestIntegrationMultiColRBAltShortcutOnDisabledItemIsIgnored verifies that
// Alt+shortcut targeting a disabled RadioButton item is ignored — selection
// does not change and focus does not move.
func TestIntegrationMultiColRBAltShortcutOnDisabledItemIsIgnored(t *testing.T) {
	// 4 items height=2: col0=(0,1) col1=(2,3).
	// Item 0 has shortcut 'a' (via tilde notation). Disable item 0.
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"~A~lpha", "~B~eta", "~G~amma", "~D~elta"})
	rbs.SetEnabled(0, false)
	rbs.SetValue(1) // item 1 is selected
	rbs.SetFocusedChild(rbs.Item(1))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'a', Modifiers: tcell.ModAlt}}
	rbs.HandleEvent(ev)

	// Selection must not have changed.
	if rbs.Value() != 1 {
		t.Errorf("Alt+a on disabled item 0: Value() = %d, want 1 (no change)", rbs.Value())
	}
	// Focus must not have moved to item 0.
	if rbs.FocusedChild() != rbs.Item(1) {
		t.Errorf("Alt+a on disabled item 0: FocusedChild = %v, want Item(1) (no focus change)", rbs.FocusedChild())
	}
}

// TestIntegrationMultiColRBAltShortcutOnEnabledItemWorks verifies that
// Alt+shortcut on an enabled RadioButton item works correctly (selection moves).
func TestIntegrationMultiColRBAltShortcutOnEnabledItemWorks(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"~A~lpha", "~B~eta", "~G~amma", "~D~elta"})
	rbs.SetValue(1) // item 1 initially selected

	// Alt+g targets item 2 (~G~amma) which is enabled.
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'g', Modifiers: tcell.ModAlt}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 2 {
		t.Errorf("Alt+g on enabled item 2: Value() = %d, want 2", rbs.Value())
	}
}

// TestIntegrationMultiColRBAltShortcutOnDisabledItemInColumn1IsIgnored verifies
// that Alt+shortcut targeting a disabled item in column 1 is also ignored.
func TestIntegrationMultiColRBAltShortcutOnDisabledItemInColumn1IsIgnored(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 2), []string{"~A~lpha", "~B~eta", "~G~amma", "~D~elta"})
	rbs.SetEnabled(2, false) // disable item 2 (~G~amma) in column 1
	rbs.SetValue(0)
	rbs.SetFocusedChild(rbs.Item(0))

	// Alt+g would target item 2 (disabled).
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'g', Modifiers: tcell.ModAlt}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("Alt+g on disabled item 2 (col 1): Value() = %d, want 0 (no change)", rbs.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 9 — SetBounds relayout: height change triggers column change
// ---------------------------------------------------------------------------

// TestIntegrationMultiColCBSetBoundsHeight3to2RelaysOut verifies that calling
// SetBounds on a CheckBoxes that changes height from 3 to 2 re-lays out items
// from 1 column to 2 columns.
func TestIntegrationMultiColCBSetBoundsHeight3to2RelaysOut(t *testing.T) {
	// Start with height=3: 6 items fit in 2 columns (3 per column).
	// Actually with height=3 and 6 items: ceil(6/3)=2 columns.
	// Let us use 3 items with height=3 to get 1 column, then shrink to height=2
	// to get 2 columns.
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 3), []string{"Alpha", "Beta", "Gamma"})

	// Confirm single column at height=3.
	for i := 0; i < 3; i++ {
		if cbs.Item(i).Bounds().A.X != 0 {
			t.Fatalf("precondition: item %d x=%d, want 0 (single column at height=3)", i, cbs.Item(i).Bounds().A.X)
		}
	}

	// Shrink height to 2 → ceil(3/2) = 2 columns.
	cbs.SetBounds(NewRect(0, 0, 40, 2))

	// Item 2 (index 2: 2/2=col1) should now be in column 1 (x > 0).
	if cbs.Item(2).Bounds().A.X == 0 {
		t.Errorf("after SetBounds(height=2), item 2 x = %d, want > 0 (moved to column 1)", cbs.Item(2).Bounds().A.X)
	}
	// Items 0 and 1 should remain in column 0.
	for i := 0; i < 2; i++ {
		if cbs.Item(i).Bounds().A.X != 0 {
			t.Errorf("after SetBounds(height=2), item %d x = %d, want 0 (still column 0)", i, cbs.Item(i).Bounds().A.X)
		}
	}
	// Item 0 should be at row 0, item 1 at row 1.
	if cbs.Item(0).Bounds().A.Y != 0 {
		t.Errorf("after SetBounds, item 0 row = %d, want 0", cbs.Item(0).Bounds().A.Y)
	}
	if cbs.Item(1).Bounds().A.Y != 1 {
		t.Errorf("after SetBounds, item 1 row = %d, want 1", cbs.Item(1).Bounds().A.Y)
	}
	// Item 2 should be at row 0 in column 1.
	if cbs.Item(2).Bounds().A.Y != 0 {
		t.Errorf("after SetBounds, item 2 row = %d, want 0 (row 0 in column 1)", cbs.Item(2).Bounds().A.Y)
	}
}

// TestIntegrationMultiColCBSetBoundsHeight2to3CollapsesToSingleColumn verifies
// that increasing height from 2 to 3 collapses back to a single column.
func TestIntegrationMultiColCBSetBoundsHeight2to3CollapsesToSingleColumn(t *testing.T) {
	// Start with height=2: 3 items → 2 columns.
	cbs := NewCheckBoxes(NewRect(0, 0, 40, 2), []string{"Alpha", "Beta", "Gamma"})

	// Confirm two columns at height=2.
	if cbs.Item(2).Bounds().A.X == 0 {
		t.Fatal("precondition: item 2 should be in column 1 at height=2")
	}

	// Increase height to 3 → single column.
	cbs.SetBounds(NewRect(0, 0, 40, 3))

	for i := 0; i < 3; i++ {
		if cbs.Item(i).Bounds().A.X != 0 {
			t.Errorf("after SetBounds(height=3), item %d x = %d, want 0 (single column)", i, cbs.Item(i).Bounds().A.X)
		}
		if cbs.Item(i).Bounds().A.Y != i {
			t.Errorf("after SetBounds(height=3), item %d y = %d, want %d", i, cbs.Item(i).Bounds().A.Y, i)
		}
	}
}

// TestIntegrationMultiColRBSetBoundsRelayoutHeight3to2 verifies that SetBounds
// on a RadioButtons changes column layout when height shrinks from 3 to 2.
func TestIntegrationMultiColRBSetBoundsRelayoutHeight3to2(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 60, 3), []string{"Alpha", "Beta", "Gamma"})

	// Confirm single column at height=3.
	for i := 0; i < 3; i++ {
		if rbs.Item(i).Bounds().A.X != 0 {
			t.Fatalf("precondition: item %d x=%d, want 0 (single column)", i, rbs.Item(i).Bounds().A.X)
		}
	}

	// Shrink height to 2 → 2 columns.
	rbs.SetBounds(NewRect(0, 0, 60, 2))

	if rbs.Item(2).Bounds().A.X == 0 {
		t.Errorf("after SetBounds(height=2), item 2 x = %d, want > 0 (column 1)", rbs.Item(2).Bounds().A.X)
	}
	// Items 0 and 1 stay in column 0.
	for i := 0; i < 2; i++ {
		if rbs.Item(i).Bounds().A.X != 0 {
			t.Errorf("after SetBounds(height=2), item %d x = %d, want 0 (column 0)", i, rbs.Item(i).Bounds().A.X)
		}
	}
}
