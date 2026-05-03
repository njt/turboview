package tv

// cluster_focus_test.go — Tests for Cluster Focus Indicator.
//
// Original TV behavior (confirmed from magiblot/tvision source): the ► indicator
// and selected color only show when the cluster itself has SfSelected (focus).
// When focus leaves the cluster, all items render uniformly with normal color.

import (
	"testing"

	"github.com/njt/turboview/theme"
)

// TestClusterFocusFocusedCheckBoxShowsCursor verifies that the focused CheckBox
// within a focused CheckBoxes cluster renders the '►' cursor at x=0, y=<item-row>.
func TestClusterFocusFocusedCheckBoxShowsCursor(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})
	cbs.scheme = theme.BorlandBlue

	cbs.SetState(SfSelected, true)
	cbs.SetFocusedChild(cbs.Item(1))

	buf := NewDrawBuffer(20, 3)
	cbs.Draw(buf)

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
// RadioButton within a focused RadioButtons cluster renders the '►' cursor.
func TestClusterFocusFocusedRadioButtonShowsCursor(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"One", "Two", "Three"})
	rbs.scheme = theme.BorlandBlue

	rbs.SetState(SfSelected, true)
	rbs.SetFocusedChild(rbs.Item(2))

	buf := NewDrawBuffer(20, 3)
	rbs.Draw(buf)

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

// TestClusterFocusIndicatorHidesWhenClusterNotFocused verifies original TV
// behavior: when the cluster loses focus in a parent group, the ► indicator
// disappears from all items. The internal focus position is preserved but has
// no visual effect.
func TestClusterFocusIndicatorHidesWhenClusterNotFocused(t *testing.T) {
	parent := NewGroup(NewRect(0, 0, 80, 25))

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})
	cbs.scheme = theme.BorlandBlue

	sibling := NewButton(NewRect(0, 5, 10, 1), "OK", CmOK)

	parent.Insert(cbs)
	parent.Insert(sibling)

	if parent.FocusedChild() == View(cbs) {
		t.Skip("cluster is still the focused child — test precondition not met")
	}

	focused := cbs.FocusedChild()
	if focused == nil {
		t.Fatal("cbs.FocusedChild() returned nil; cluster must always track an internal focus")
	}
	if !focused.HasState(SfSelected) {
		t.Fatal("cbs.FocusedChild() does not have SfSelected; internal focus state lost")
	}

	buf := NewDrawBuffer(20, 3)
	cbs.Draw(buf)

	for row := 0; row < 3; row++ {
		cell := buf.GetCell(0, row)
		if cell.Rune == '►' {
			t.Errorf("row %d shows '►' when cluster is unfocused — original TV hides all indicators", row)
		}
	}
}
