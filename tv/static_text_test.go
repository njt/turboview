package tv

import (
	"testing"

	"github.com/njt/turboview/theme"
)

// compile-time assertion: StaticText must satisfy Widget.
// Spec: "StaticText embeds BaseView and satisfies the Widget interface: var _ Widget = (*StaticText)(nil)"
var _ Widget = (*StaticText)(nil)

// --- Construction ---

// TestNewStaticTextSetsSfVisible verifies NewStaticText sets the SfVisible state flag.
// Spec: "NewStaticText(bounds, text) creates a StaticText with SfVisible set."
func TestNewStaticTextSetsSfVisible(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 20, 3), "hello")

	if !st.HasState(SfVisible) {
		t.Error("NewStaticText did not set SfVisible")
	}
}

// TestNewStaticTextIsNotSelectable verifies NewStaticText does NOT set OfSelectable.
// Spec: "It is NOT selectable (does not set OfSelectable)."
func TestNewStaticTextIsNotSelectable(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 20, 3), "hello")

	if st.HasOption(OfSelectable) {
		t.Error("NewStaticText must not set OfSelectable")
	}
}

// TestNewStaticTextStoresBounds verifies NewStaticText records the given bounds.
// Spec: "NewStaticText(bounds Rect, text string) *StaticText"
func TestNewStaticTextStoresBounds(t *testing.T) {
	r := NewRect(5, 10, 30, 4)
	st := NewStaticText(r, "test")

	if st.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", st.Bounds(), r)
	}
}

// TestNewStaticTextStoresText verifies NewStaticText records the initial text.
// Spec: "NewStaticText(bounds Rect, text string) *StaticText"
func TestNewStaticTextStoresText(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 20, 3), "initial text")

	if st.Text() != "initial text" {
		t.Errorf("Text() = %q, want %q", st.Text(), "initial text")
	}
}

// --- Text / SetText ---

// TestStaticTextTextReturnsCurrentText verifies Text() returns the current text.
// Spec: "StaticText.Text() string returns the current text."
func TestStaticTextTextReturnsCurrentText(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 20, 3), "original")

	if got := st.Text(); got != "original" {
		t.Errorf("Text() = %q, want %q", got, "original")
	}
}

// TestStaticTextSetTextUpdatesText verifies SetText changes the value returned by Text.
// Spec: "StaticText.SetText(t string) updates the text."
func TestStaticTextSetTextUpdatesText(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 20, 3), "original")
	st.SetText("updated")

	if got := st.Text(); got != "updated" {
		t.Errorf("Text() after SetText = %q, want %q", got, "updated")
	}
}

// TestStaticTextSetTextToEmpty verifies SetText accepts an empty string.
// Spec: "StaticText.SetText(t string) updates the text."
func TestStaticTextSetTextToEmpty(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 20, 3), "something")
	st.SetText("")

	if got := st.Text(); got != "" {
		t.Errorf("Text() after SetText(\"\") = %q, want %q", got, "")
	}
}

// TestStaticTextSetTextReplacesNotAppends verifies SetText replaces, not appends.
// Spec: "StaticText.SetText(t string) updates the text."
func TestStaticTextSetTextReplacesNotAppends(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 20, 3), "first")
	st.SetText("second")
	st.SetText("third")

	if got := st.Text(); got != "third" {
		t.Errorf("Text() after two SetText calls = %q, want %q", got, "third")
	}
}

// --- HandleEvent ---

// TestStaticTextHandleEventIsNoOp verifies HandleEvent does nothing (BaseView default).
// Spec: "StaticText.HandleEvent(event) is a no-op (BaseView default)."
func TestStaticTextHandleEventIsNoOp(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 20, 3), "hello")
	event := &Event{What: EvKeyboard}

	// Must not panic; event must be unmodified.
	st.HandleEvent(event)

	if event.What != EvKeyboard {
		t.Errorf("HandleEvent changed event.What = %v, want EvKeyboard", event.What)
	}
}

// --- Draw: style ---

// TestStaticTextDrawUsesLabelNormalStyle verifies Draw uses the LabelNormal style
// from the ColorScheme for rendered characters.
// Spec: "Uses LabelNormal style from ColorScheme."
func TestStaticTextDrawUsesLabelNormalStyle(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 20, 3), "Hi")
	scheme := theme.BorlandBlue
	st.scheme = scheme

	buf := NewDrawBuffer(20, 3)
	st.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.LabelNormal {
		t.Errorf("Draw cell(0,0) style = %v, want LabelNormal %v", cell.Style, scheme.LabelNormal)
	}
}

// --- Draw: position ---

// TestStaticTextDrawStartsAtOrigin verifies the first character is at (0, 0).
// Spec: "renders the text starting at (0, 0) within its bounds"
func TestStaticTextDrawStartsAtOrigin(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 20, 3), "Hello")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 3)
	st.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != 'H' {
		t.Errorf("Draw cell(0,0) = %q, want 'H'", cell.Rune)
	}
}

// TestStaticTextDrawRendersAllCharactersOnFirstLine verifies a short text is
// rendered consecutively from (0, 0).
// Spec: "renders the text starting at (0, 0) within its bounds"
func TestStaticTextDrawRendersAllCharactersOnFirstLine(t *testing.T) {
	text := "ABC"
	st := NewStaticText(NewRect(0, 0, 20, 3), text)
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 3)
	st.Draw(buf)

	for i, want := range text {
		got := buf.GetCell(i, 0).Rune
		if got != want {
			t.Errorf("Draw cell(%d, 0) = %q, want %q", i, got, want)
		}
	}
}

// --- Draw: word wrapping ---

// TestStaticTextDrawWrapsWordToNextLine verifies that a word that would exceed
// the line width starts on the next line.
// Spec: "wrapping to the next line at the bounds width. Word wrapping splits on spaces:
// words that would exceed the line width start on the next line."
func TestStaticTextDrawWrapsWordToNextLine(t *testing.T) {
	// Width=5. "Hello World": "Hello" fits line 0. "World" would push past width, goes to line 1.
	st := NewStaticText(NewRect(0, 0, 5, 2), "Hello World")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(5, 2)
	st.Draw(buf)

	// "Hello" on row 0
	for i, want := range "Hello" {
		got := buf.GetCell(i, 0).Rune
		if got != want {
			t.Errorf("row0 cell(%d) = %q, want %q", i, got, want)
		}
	}

	// "World" starts at (0, 1)
	for i, want := range "World" {
		got := buf.GetCell(i, 1).Rune
		if got != want {
			t.Errorf("row1 cell(%d) = %q, want %q", i, got, want)
		}
	}
}

// TestStaticTextDrawWrapsMultipleWords verifies wrapping across several words.
// Spec: "words that would exceed the line width start on the next line."
func TestStaticTextDrawWrapsMultipleWords(t *testing.T) {
	// Width=6. "one two three": "one" (3) fits, space+two=7 > 6 → "two" to row 1.
	st := NewStaticText(NewRect(0, 0, 6, 3), "one two three")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(6, 3)
	st.Draw(buf)

	// Row 0: starts with "one"
	if buf.GetCell(0, 0).Rune != 'o' || buf.GetCell(1, 0).Rune != 'n' || buf.GetCell(2, 0).Rune != 'e' {
		t.Errorf("row0 = %q%q%q, want \"one\"",
			buf.GetCell(0, 0).Rune, buf.GetCell(1, 0).Rune, buf.GetCell(2, 0).Rune)
	}

	// Row 1 must start with "two" (the word that didn't fit)
	if buf.GetCell(0, 1).Rune != 't' {
		t.Errorf("row1 cell(0) = %q, want 't' (start of \"two\")", buf.GetCell(0, 1).Rune)
	}
}

// TestStaticTextDrawLongWordPlacedAtLineStartAndClipped verifies that a single word
// longer than the line width is placed at the start of the next line and clipped by
// the DrawBuffer.
// Spec: "A word longer than the line width is placed at the start of a line and clipped."
func TestStaticTextDrawLongWordPlacedAtLineStart(t *testing.T) {
	// Width=4. "ab toolong": "ab" fits row 0; "toolong" (7) > 4, placed at start of row 1.
	st := NewStaticText(NewRect(0, 0, 4, 2), "ab toolong")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(4, 2)
	st.Draw(buf)

	// "ab" on row 0
	if buf.GetCell(0, 0).Rune != 'a' || buf.GetCell(1, 0).Rune != 'b' {
		t.Errorf("row0 = %q%q, want \"ab\"", buf.GetCell(0, 0).Rune, buf.GetCell(1, 0).Rune)
	}

	// "tool" on row 1 (first 4 chars of "toolong", rest clipped by DrawBuffer)
	for i, want := range "tool" {
		got := buf.GetCell(i, 1).Rune
		if got != want {
			t.Errorf("row1 cell(%d) = %q, want %q (long word at start, clipped)", i, got, want)
		}
	}
}

// TestStaticTextDrawLongWordAloneOnFirstLine verifies a lone oversized word starts at (0,0).
// Spec: "A word longer than the line width is placed at the start of a line and clipped."
func TestStaticTextDrawLongWordAloneOnFirstLine(t *testing.T) {
	// Width=3. Single word "Hello" (5) > 3, placed at (0,0), clipped after 3 chars.
	st := NewStaticText(NewRect(0, 0, 3, 2), "Hello")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(3, 2)
	st.Draw(buf)

	// First char at (0,0)
	if buf.GetCell(0, 0).Rune != 'H' {
		t.Errorf("row0 cell(0) = %q, want 'H'", buf.GetCell(0, 0).Rune)
	}

	// DrawBuffer clips anything beyond width=3; chars at columns 3+ are not written.
	// Only verify the characters that are within bounds.
	for i, want := range "Hel" {
		got := buf.GetCell(i, 0).Rune
		if got != want {
			t.Errorf("row0 cell(%d) = %q, want %q", i, got, want)
		}
	}
}

// --- Draw: clipping ---

// TestStaticTextDrawClipsTextBeyondBoundsHeight verifies that text exceeding the
// available vertical area is clipped (DrawBuffer handles this naturally).
// Spec: "If the text exceeds the available area, it is clipped (DrawBuffer handles this naturally)."
func TestStaticTextDrawClipsTextBeyondBoundsHeight(t *testing.T) {
	// 3 words, each on its own line (width=5), but only 2 rows available.
	st := NewStaticText(NewRect(0, 0, 5, 2), "one two three")
	st.scheme = theme.BorlandBlue

	// Buffer has 2 rows; third word ("three") must not appear.
	buf := NewDrawBuffer(5, 2)
	st.Draw(buf)

	// Must not panic; row 1 exists but row 2 would be out of bounds.
	// Verify the buffer height is still 2 (unchanged).
	if buf.Height() != 2 {
		t.Errorf("DrawBuffer height changed to %d, want 2", buf.Height())
	}
}

// TestStaticTextDrawEmptyTextLeavesBufferUnchanged verifies Draw with empty text
// does not write any characters.
// Spec: "renders the text starting at (0, 0) within its bounds"
// — empty text renders nothing.
func TestStaticTextDrawEmptyTextLeavesBufferUnchanged(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 10, 3), "")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(10, 3)
	// Record initial state of first cell.
	initial := buf.GetCell(0, 0)

	st.Draw(buf)

	// With empty text, the first cell should be unchanged (still the default space).
	after := buf.GetCell(0, 0)
	if after.Rune != ' ' && after.Rune != initial.Rune {
		// An implementation may or may not write styled spaces; it must not
		// write a non-space character when text is empty.
		t.Errorf("Draw with empty text wrote %q at (0,0), want space or unchanged", after.Rune)
	}
}

// TestStaticTextDrawWordWrapDoesNotSplitMidWord verifies wrapping never breaks
// a word in the middle — the wrapped content on row 1 starts at the word boundary.
// Spec: "Word wrapping splits on spaces: words that would exceed the line width start on the next line."
func TestStaticTextDrawWordWrapDoesNotSplitMidWord(t *testing.T) {
	// Width=4. "Hi there": "Hi" (2) fits row 0; " there" pushes "there" to row 1.
	st := NewStaticText(NewRect(0, 0, 4, 2), "Hi there")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(4, 2)
	st.Draw(buf)

	// The character at (2, 0) must NOT be 't' (start of "there"), confirming no mid-word split.
	// It should be a space or an empty cell since "Hi" ends at column 1.
	cell := buf.GetCell(2, 0)
	if cell.Rune == 't' {
		t.Errorf("Draw split word mid-line: 't' appeared at (2,0) on row with 'Hi'")
	}

	// Row 1 must start with 't' of "there".
	if buf.GetCell(0, 1).Rune != 't' {
		t.Errorf("row1 cell(0) = %q, want 't' (start of \"there\" after wrap)", buf.GetCell(0, 1).Rune)
	}
}
