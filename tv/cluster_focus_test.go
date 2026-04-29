package tv

// cluster_focus_test.go — Tests for Task 1: Cluster Focus Indicator.
//
// Spec: CheckBoxes.Draw and RadioButtons.Draw must NOT suppress SfSelected on
// the focused child before drawing. The focused item shows the '►' cursor prefix;
// other items in the same cluster do not. The indicator persists even when the
// cluster itself is not the focused widget in a parent group.

import (
	"testing"

	"github.com/njt/turboview/theme"
)

// TestClusterFocusFocusedCheckBoxShowsCursor verifies that the focused CheckBox
// within a CheckBoxes cluster renders the '►' cursor at x=0, y=<item-row>.
// Spec: "CheckBoxes.Draw must NOT suppress SfSelected on the focused child —
// the focused checkbox shows the ► cursor prefix."
func TestClusterFocusFocusedCheckBoxShowsCursor(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})
	cbs.scheme = theme.BorlandBlue

	// Move focus to Item(1); the group sets SfSelected on it.
	cbs.SetFocusedChild(cbs.Item(1))

	buf := NewDrawBuffer(20, 3)
	cbs.Draw(buf)

	// Item(1) is at y=1.  Its '►' must appear at (0, 1).
	cell := buf.GetCell(0, 1)
	if cell.Rune != '►' {
		t.Errorf("focused CheckBox (item 1) at row 1: cell(0,1) = %q, want '►'", cell.Rune)
	}
}

// TestClusterFocusUnfocusedCheckBoxDoesNotShowCursor verifies that a CheckBox
// that is NOT the focused child does not show '►'.
// Spec: "Only the internally-focused item within the cluster shows ►."
func TestClusterFocusUnfocusedCheckBoxDoesNotShowCursor(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})
	cbs.scheme = theme.BorlandBlue

	// Focus item 1; items 0 and 2 must not show the cursor.
	cbs.SetFocusedChild(cbs.Item(1))

	buf := NewDrawBuffer(20, 3)
	cbs.Draw(buf)

	// Item(0) is at y=0; its first cell must be '[', not '►'.
	if buf.GetCell(0, 0).Rune == '►' {
		t.Errorf("unfocused CheckBox (item 0) at row 0: cell(0,0) = '►'; only the focused item may show the cursor")
	}

	// Item(2) is at y=2; its first cell must be '[', not '►'.
	if buf.GetCell(0, 2).Rune == '►' {
		t.Errorf("unfocused CheckBox (item 2) at row 2: cell(0,2) = '►'; only the focused item may show the cursor")
	}
}

// TestClusterFocusFocusedRadioButtonShowsCursor verifies that the focused
// RadioButton within a RadioButtons cluster renders the '►' cursor at its row.
// Spec: "RadioButtons.Draw must NOT suppress SfSelected on the focused child —
// the focused radio button shows the ► cursor prefix."
func TestClusterFocusFocusedRadioButtonShowsCursor(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"One", "Two", "Three"})
	rbs.scheme = theme.BorlandBlue

	// Move focus to Item(2); the group sets SfSelected on it.
	rbs.SetFocusedChild(rbs.Item(2))

	buf := NewDrawBuffer(20, 3)
	rbs.Draw(buf)

	// Item(2) is at y=2. Its '►' must appear at (0, 2).
	cell := buf.GetCell(0, 2)
	if cell.Rune != '►' {
		t.Errorf("focused RadioButton (item 2) at row 2: cell(0,2) = %q, want '►'", cell.Rune)
	}
}

// TestClusterFocusUnfocusedRadioButtonDoesNotShowCursor verifies that RadioButtons
// that are not the focused child do not display '►'.
// Spec: "Only the internally-focused item within the cluster shows ►."
func TestClusterFocusUnfocusedRadioButtonDoesNotShowCursor(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"One", "Two", "Three"})
	rbs.scheme = theme.BorlandBlue

	// Focus item 2; items 0 and 1 must not show the cursor.
	rbs.SetFocusedChild(rbs.Item(2))

	buf := NewDrawBuffer(20, 3)
	rbs.Draw(buf)

	// Item(0) is at y=0; must start with '(', not '►'.
	if buf.GetCell(0, 0).Rune == '►' {
		t.Errorf("unfocused RadioButton (item 0) at row 0: cell(0,0) = '►'; only the focused item may show the cursor")
	}

	// Item(1) is at y=1; must start with '(', not '►'.
	if buf.GetCell(0, 1).Rune == '►' {
		t.Errorf("unfocused RadioButton (item 1) at row 1: cell(0,1) = '►'; only the focused item may show the cursor")
	}
}

// TestClusterFocusIndicatorPersistsWhenClusterNotFocused verifies that the
// internal focus indicator (the '►' on the last-focused child) is still rendered
// even when the cluster itself does NOT have SfSelected in a parent group — i.e.,
// the cluster is not the active widget.
// Spec: "When the cluster itself is not focused (SfSelected not set on the cluster),
// the internal focus indicator still shows on whichever item was last focused."
func TestClusterFocusIndicatorPersistsWhenClusterNotFocused(t *testing.T) {
	// Build a parent group with two siblings: the cluster and a button.
	parent := NewGroup(NewRect(0, 0, 80, 25))

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})
	cbs.scheme = theme.BorlandBlue

	// A second selectable sibling so the cluster can be unfocused.
	sibling := NewButton(NewRect(0, 5, 10, 1), "OK", CmOK)

	parent.Insert(cbs)
	parent.Insert(sibling) // sibling inserted last → it gets focus, cluster loses SfSelected

	// Confirm the cluster itself is no longer the active (focused) widget in the parent.
	if parent.FocusedChild() == View(cbs) {
		t.Skip("cluster is still the focused child after sibling was inserted — test precondition not met")
	}

	// The cluster's internal focus should still be on its last-focused item (whichever
	// the group set before losing focus; at minimum one child has SfSelected from the
	// group's own focus management).  Confirm the cluster's own FocusedChild has
	// SfSelected set at the CheckBox level.
	focused := cbs.FocusedChild()
	if focused == nil {
		t.Fatal("cbs.FocusedChild() returned nil; cluster must always track an internal focus")
	}
	if !focused.HasState(SfSelected) {
		t.Fatal("cbs.FocusedChild() does not have SfSelected; internal focus state lost")
	}

	// Now Draw: the cluster is not the active widget in the parent, but its Draw
	// must still pass SfSelected through to the focused child, producing '►'.
	buf := NewDrawBuffer(20, 3)
	cbs.Draw(buf)

	focusedCB := focused.(*CheckBox)
	row := focusedCB.Bounds().A.Y

	cell := buf.GetCell(0, row)
	if cell.Rune != '►' {
		t.Errorf("cluster not focused in parent, but focused child (row %d) should still show '►'; got %q",
			row, cell.Rune)
	}
}
