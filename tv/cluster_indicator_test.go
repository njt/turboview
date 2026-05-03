package tv

import (
	"testing"
)

// cluster_indicator_test.go — Tests for fixed-position focus indicator in
// CheckBox and RadioButton clusters. In original Turbo Vision, the focus
// indicator does NOT shift content — items always render at fixed positions.

// TestCheckBoxDrawUnfocusedAtFixedPosition verifies that an unfocused
// checkbox draws its bracket at column 1 (column 0 reserved for indicator).
func TestCheckBoxDrawUnfocusedAtFixedPosition(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 30, 1), "Test")
	buf := NewDrawBuffer(30, 1)
	cb.Draw(buf)

	c0 := buf.GetCell(0, 0)
	if c0.Rune != ' ' {
		t.Errorf("unfocused checkbox column 0: got %q, want space (reserved for indicator)", string(c0.Rune))
	}

	c1 := buf.GetCell(1, 0)
	if c1.Rune != '[' {
		t.Errorf("unfocused checkbox column 1: got %q, want '['", string(c1.Rune))
	}
}

// TestCheckBoxDrawFocusedAtSamePosition verifies that a focused checkbox
// draws its bracket at the SAME column as unfocused (column 1).
func TestCheckBoxDrawFocusedAtSamePosition(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 30, 1), "Test")
	cb.SetState(SfSelected, true)
	buf := NewDrawBuffer(30, 1)
	cb.Draw(buf)

	c1 := buf.GetCell(1, 0)
	if c1.Rune != '[' {
		t.Errorf("focused checkbox column 1: got %q, want '[' (focus indicator must not shift content)", string(c1.Rune))
	}
}

// TestCheckBoxLabelPositionStable verifies that the label text starts at
// the same column regardless of focus state.
func TestCheckBoxLabelPositionStable(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 30, 1), "Alpha")

	buf1 := NewDrawBuffer(30, 1)
	cb.Draw(buf1)

	unfocusedCol := -1
	for x := 0; x < 30; x++ {
		c := buf1.GetCell(x, 0)
		if c.Rune == 'A' {
			unfocusedCol = x
			break
		}
	}

	cb.SetState(SfSelected, true)
	buf2 := NewDrawBuffer(30, 1)
	cb.Draw(buf2)

	focusedCol := -1
	for x := 0; x < 30; x++ {
		c := buf2.GetCell(x, 0)
		if c.Rune == 'A' {
			focusedCol = x
			break
		}
	}

	if unfocusedCol < 0 || focusedCol < 0 {
		t.Fatal("could not find label 'A' in either buffer")
	}

	if unfocusedCol != focusedCol {
		t.Errorf("label 'A' at column %d unfocused, column %d focused — content shifted by focus indicator", unfocusedCol, focusedCol)
	}
}

// TestRadioButtonDrawUnfocusedAtFixedPosition verifies unfocused radio button
// draws its paren at column 1.
func TestRadioButtonDrawUnfocusedAtFixedPosition(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 30, 1), "Test")
	buf := NewDrawBuffer(30, 1)
	rb.Draw(buf)

	c0 := buf.GetCell(0, 0)
	if c0.Rune != ' ' {
		t.Errorf("unfocused radio button column 0: got %q, want space", string(c0.Rune))
	}

	c1 := buf.GetCell(1, 0)
	if c1.Rune != '(' {
		t.Errorf("unfocused radio button column 1: got %q, want '('", string(c1.Rune))
	}
}

// TestRadioButtonDrawFocusedAtSamePosition verifies focused radio button
// keeps paren at column 1, with indicator at column 0.
func TestRadioButtonDrawFocusedAtSamePosition(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 30, 1), "Test")
	rb.SetState(SfSelected, true)
	buf := NewDrawBuffer(30, 1)
	rb.Draw(buf)

	c1 := buf.GetCell(1, 0)
	if c1.Rune != '(' {
		t.Errorf("focused radio button column 1: got %q, want '(' (indicator must not shift content)", string(c1.Rune))
	}
}

// TestRadioButtonLabelPositionStable verifies radio button label stays at
// same column regardless of focus.
func TestRadioButtonLabelPositionStable(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 30, 1), "Beta")

	buf1 := NewDrawBuffer(30, 1)
	rb.Draw(buf1)

	unfocusedCol := -1
	for x := 0; x < 30; x++ {
		c := buf1.GetCell(x, 0)
		if c.Rune == 'B' {
			unfocusedCol = x
			break
		}
	}

	rb.SetState(SfSelected, true)
	buf2 := NewDrawBuffer(30, 1)
	rb.Draw(buf2)

	focusedCol := -1
	for x := 0; x < 30; x++ {
		c := buf2.GetCell(x, 0)
		if c.Rune == 'B' {
			focusedCol = x
			break
		}
	}

	if unfocusedCol < 0 || focusedCol < 0 {
		t.Fatal("could not find label 'B' in either buffer")
	}

	if unfocusedCol != focusedCol {
		t.Errorf("label 'B' at column %d unfocused, column %d focused — content shifted", unfocusedCol, focusedCol)
	}
}

// TestCheckBoxFocusUsesSelectedStyle verifies the focused checkbox uses
// CheckBoxSelected color scheme entry (not just the same as unfocused).
func TestCheckBoxFocusUsesSelectedStyle(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 30, 1), "Test")
	cb.SetState(SfSelected, true)

	cs := cb.ColorScheme()
	if cs == nil {
		t.Skip("no color scheme set")
	}

	buf := NewDrawBuffer(30, 1)
	cb.Draw(buf)

	c1 := buf.GetCell(1, 0)
	if c1.Style == cs.CheckBoxNormal {
		t.Error("focused checkbox bracket uses CheckBoxNormal style — should use CheckBoxSelected for focus differentiation")
	}
}

// TestRadioButtonFocusUsesSelectedStyle verifies the focused radio button uses
// RadioButtonSelected color scheme entry.
func TestRadioButtonFocusUsesSelectedStyle(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 30, 1), "Test")
	rb.SetState(SfSelected, true)

	cs := rb.ColorScheme()
	if cs == nil {
		t.Skip("no color scheme set")
	}

	buf := NewDrawBuffer(30, 1)
	rb.Draw(buf)

	c1 := buf.GetCell(1, 0)
	if c1.Style == cs.RadioButtonNormal {
		t.Error("focused radio button paren uses RadioButtonNormal style — should use RadioButtonSelected for focus differentiation")
	}
}
