package tv

import (
	"testing"

	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Tests for StaticText centering with \x03 prefix.
// ---------------------------------------------------------------------------

// TestStaticTextCenteringCenteredLineIscentered verifies that a line beginning
// with \x03 is drawn centered within the view width.
// Spec: "When a line begins with \x03, center that line within the view width"
func TestStaticTextCenteringCenteredLineIsCentered(t *testing.T) {
	// Width=20, text="\x03Hi" (2 chars). Expected startX = (20-2)/2 = 9.
	st := NewStaticText(NewRect(0, 0, 20, 3), "\x03Hi")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 3)
	st.Draw(buf)

	if buf.GetCell(9, 0).Rune != 'H' {
		t.Errorf("centered 'H': got %q at (9,0), want 'H'", buf.GetCell(9, 0).Rune)
	}
	if buf.GetCell(10, 0).Rune != 'i' {
		t.Errorf("centered 'i': got %q at (10,0), want 'i'", buf.GetCell(10, 0).Rune)
	}
}

// TestStaticTextCenteringPrefixIsConsumed verifies that \x03 itself is not drawn.
// Spec: "\x03 is consumed (not displayed)"
func TestStaticTextCenteringPrefixIsConsumed(t *testing.T) {
	// If \x03 were drawn as a rune it would appear somewhere on row 0.
	// We check no cell on row 0 holds rune 3.
	st := NewStaticText(NewRect(0, 0, 20, 3), "\x03Hi")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 3)
	st.Draw(buf)

	for x := 0; x < 20; x++ {
		if buf.GetCell(x, 0).Rune == '\x03' {
			t.Errorf("Draw drew \\x03 at (%d,0); it must be consumed", x)
		}
	}
}

// TestStaticTextCenteringSubsequentLineIsLeftAligned verifies that a line NOT
// prefixed with \x03 is left-aligned even when it follows a centered line.
// Spec: "Centering applies only to that line — subsequent lines are left-aligned
// unless they also start with \x03"
func TestStaticTextCenteringSubsequentLineIsLeftAligned(t *testing.T) {
	// Two lines: first centered, second normal.
	st := NewStaticText(NewRect(0, 0, 20, 3), "\x03Hi\nHello")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 3)
	st.Draw(buf)

	// Second line "Hello" must start at x=0.
	if buf.GetCell(0, 1).Rune != 'H' {
		t.Errorf("second line: got %q at (0,1), want 'H' (left-aligned)", buf.GetCell(0, 1).Rune)
	}
}

// TestStaticTextCenteringBothLinesCentered verifies that two consecutive lines
// each prefixed with \x03 are both centered independently.
// Spec: "Centering applies only to that line"
func TestStaticTextCenteringBothLinesCentered(t *testing.T) {
	// Width=20. Line1 "\x03AB" → (20-2)/2=9. Line2 "\x03ABCDE" → (20-5)/2=7.
	st := NewStaticText(NewRect(0, 0, 20, 3), "\x03AB\n\x03ABCDE")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 3)
	st.Draw(buf)

	// Row 0: "AB" starts at x=9.
	if buf.GetCell(9, 0).Rune != 'A' {
		t.Errorf("row0 'A': got %q at (9,0), want 'A'", buf.GetCell(9, 0).Rune)
	}

	// Row 1: "ABCDE" starts at x=7.
	if buf.GetCell(7, 1).Rune != 'A' {
		t.Errorf("row1 'A': got %q at (7,1), want 'A'", buf.GetCell(7, 1).Rune)
	}
}

// TestStaticTextCenteringClampsToZeroWhenTextExceedsWidth verifies that
// centering does not produce a negative x offset.
// Spec: "Centering: startX = (viewWidth - textWidth) / 2, clamped to 0"
func TestStaticTextCenteringClampsToZeroWhenTextExceedsWidth(t *testing.T) {
	// Width=4, text "Hello" (5 chars) → (4-5)/2 = -0.5 → clamped to 0.
	st := NewStaticText(NewRect(0, 0, 4, 2), "\x03Hello")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(4, 2)
	st.Draw(buf)

	// Must start at x=0 (clamped) and not panic.
	if buf.GetCell(0, 0).Rune != 'H' {
		t.Errorf("clamped center: got %q at (0,0), want 'H'", buf.GetCell(0, 0).Rune)
	}
}

// TestStaticTextCenteringWordWrappingPreservedForNonCenteredLine verifies that
// word-wrapping still works for non-centered lines (regression guard).
// Spec: "Must preserve existing word-wrapping behavior"
func TestStaticTextCenteringWordWrappingPreservedForNonCenteredLine(t *testing.T) {
	// Width=5. "Hello World": wraps as before.
	st := NewStaticText(NewRect(0, 0, 5, 2), "Hello World")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(5, 2)
	st.Draw(buf)

	for i, want := range "Hello" {
		if buf.GetCell(i, 0).Rune != want {
			t.Errorf("row0 cell(%d) = %q, want %q", i, buf.GetCell(i, 0).Rune, want)
		}
	}
	if buf.GetCell(0, 1).Rune != 'W' {
		t.Errorf("row1 cell(0) = %q, want 'W'", buf.GetCell(0, 1).Rune)
	}
}
