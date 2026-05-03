package tv

import (
	"testing"

	"github.com/njt/turboview/theme"
)

// cluster_unfocused_test.go — Tests for original TV behavior: when a cluster
// (CheckBoxes/RadioButtons) loses focus, ALL visual indicators disappear.
// The focused item within the cluster retains its SfSelected state internally,
// but Draw must not show the ► indicator or selected color when the parent
// cluster itself doesn't have SfSelected.

func makeClusterTestScheme() *theme.ColorScheme {
	return theme.BorlandBlue
}

// --- CheckBoxes unfocused ---

// TestCheckBoxesUnfocusedHidesIndicator verifies that when the CheckBoxes
// cluster does NOT have SfSelected (focus has moved away), no item shows
// the ► indicator — even the internally-focused item.
func TestCheckBoxesUnfocusedHidesIndicator(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 30, 3), []string{"Alpha", "Beta", "Gamma"})

	// The last item is the focused child (inserted last, gets SfSelected).
	// Verify internal state: one item has SfSelected.
	focusedIdx := -1
	for i, item := range []View{cbs.Item(0), cbs.Item(1), cbs.Item(2)} {
		if item.HasState(SfSelected) {
			focusedIdx = i
		}
	}
	if focusedIdx < 0 {
		t.Fatal("expected one checkbox to have SfSelected internally")
	}

	// The cluster itself does NOT have SfSelected (simulating focus elsewhere).
	cbs.SetState(SfSelected, false)

	buf := NewDrawBuffer(30, 3)
	cbs.Draw(buf)

	// No row should show the ► indicator.
	for row := 0; row < 3; row++ {
		c := buf.GetCell(0, row)
		if c.Rune == '►' {
			t.Errorf("row %d shows ► indicator when cluster is unfocused — should be space", row)
		}
	}
}

// TestCheckBoxesFocusedShowsIndicator verifies that when the cluster HAS
// SfSelected, the focused item's ► indicator is visible.
func TestCheckBoxesFocusedShowsIndicator(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 30, 3), []string{"Alpha", "Beta", "Gamma"})

	// Give cluster focus.
	cbs.SetState(SfSelected, true)

	// Focus the second item.
	cbs.SetFocusedChild(cbs.Item(1))

	buf := NewDrawBuffer(30, 3)
	cbs.Draw(buf)

	// Row 1 (Beta) should show ►.
	c := buf.GetCell(0, 1)
	if c.Rune != '►' {
		t.Errorf("focused item row 1: got %q at col 0, want '►'", string(c.Rune))
	}

	// Rows 0 and 2 should NOT show ►.
	for _, row := range []int{0, 2} {
		c := buf.GetCell(0, row)
		if c.Rune == '►' {
			t.Errorf("unfocused item row %d shows ► — only focused item should", row)
		}
	}
}

// TestCheckBoxesUnfocusedUsesNormalColor verifies that when the cluster
// loses focus, all items use CheckBoxNormal style (not CheckBoxSelected).
func TestCheckBoxesUnfocusedUsesNormalColor(t *testing.T) {
	cs := makeClusterTestScheme()
	cbs := NewCheckBoxes(NewRect(0, 0, 30, 3), []string{"Alpha", "Beta", "Gamma"})
	cbs.BaseView.scheme = cs

	// Set up: cluster unfocused, but internal item has SfSelected.
	cbs.SetState(SfSelected, false)

	buf := NewDrawBuffer(30, 3)
	cbs.Draw(buf)

	// All brackets should use CheckBoxNormal style.
	for row := 0; row < 3; row++ {
		c := buf.GetCell(1, row)
		if c.Style == cs.CheckBoxSelected {
			t.Errorf("row %d bracket uses CheckBoxSelected when cluster is unfocused", row)
		}
	}
}

// --- RadioButtons unfocused ---

// TestRadioButtonsUnfocusedHidesIndicator verifies that when the RadioButtons
// cluster does NOT have SfSelected, no item shows the ► indicator.
func TestRadioButtonsUnfocusedHidesIndicator(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 30, 3), []string{"Red", "Green", "Blue"})

	// Cluster loses focus.
	rbs.SetState(SfSelected, false)

	buf := NewDrawBuffer(30, 3)
	rbs.Draw(buf)

	for row := 0; row < 3; row++ {
		c := buf.GetCell(0, row)
		if c.Rune == '►' {
			t.Errorf("row %d shows ► indicator when radio cluster is unfocused", row)
		}
	}
}

// TestRadioButtonsFocusedShowsIndicator verifies that when the radio cluster
// HAS SfSelected, the focused item shows ►.
func TestRadioButtonsFocusedShowsIndicator(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 30, 3), []string{"Red", "Green", "Blue"})

	// Give cluster focus.
	rbs.SetState(SfSelected, true)

	// Focus the first item.
	rbs.SetFocusedChild(rbs.Item(0))

	buf := NewDrawBuffer(30, 3)
	rbs.Draw(buf)

	c := buf.GetCell(0, 0)
	if c.Rune != '►' {
		t.Errorf("focused radio item row 0: got %q at col 0, want '►'", string(c.Rune))
	}
}

// TestRadioButtonsUnfocusedUsesNormalColor verifies that when the radio
// cluster loses focus, all items use RadioButtonNormal (not Selected).
func TestRadioButtonsUnfocusedUsesNormalColor(t *testing.T) {
	cs := makeClusterTestScheme()
	rbs := NewRadioButtons(NewRect(0, 0, 30, 3), []string{"Red", "Green", "Blue"})
	rbs.BaseView.scheme = cs

	rbs.SetState(SfSelected, false)

	buf := NewDrawBuffer(30, 3)
	rbs.Draw(buf)

	for row := 0; row < 3; row++ {
		c := buf.GetCell(1, row)
		if c.Style == cs.RadioButtonSelected {
			t.Errorf("row %d paren uses RadioButtonSelected when cluster is unfocused", row)
		}
	}
}

// --- Transition test ---

// TestCheckBoxesIndicatorDisappearsOnFocusLoss verifies the transition:
// cluster focused → cluster unfocused → indicator disappears.
func TestCheckBoxesIndicatorDisappearsOnFocusLoss(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 30, 3), []string{"Alpha", "Beta", "Gamma"})
	cbs.SetState(SfSelected, true)
	cbs.SetFocusedChild(cbs.Item(0))

	// While focused: should show indicator.
	buf1 := NewDrawBuffer(30, 3)
	cbs.Draw(buf1)
	c := buf1.GetCell(0, 0)
	if c.Rune != '►' {
		t.Fatalf("precondition: focused item should show ► but got %q", string(c.Rune))
	}

	// Remove cluster focus.
	cbs.SetState(SfSelected, false)

	// After losing focus: indicator must disappear.
	buf2 := NewDrawBuffer(30, 3)
	cbs.Draw(buf2)
	c = buf2.GetCell(0, 0)
	if c.Rune == '►' {
		t.Error("indicator ► still visible after cluster lost focus — should be space")
	}
}
