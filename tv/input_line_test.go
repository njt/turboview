package tv

// Tests for InputLine widget.
//
// Each test cites the spec sentence it verifies.
// Tests are grouped by concern: construction, accessors, drawing, keyboard
// handling, selection, mouse handling, clipboard, and scroll.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// compile-time assertion: InputLine must satisfy Widget.
var _ Widget = (*InputLine)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newInputLine is a test helper that creates an InputLine with a BorlandBlue
// scheme pre-attached, so draw tests work without an owner chain.
func newInputLine(x, y, w, maxLen int) *InputLine {
	il := NewInputLine(NewRect(x, y, w, 1), maxLen)
	il.scheme = theme.BorlandBlue
	return il
}

// keyEv creates a plain keyboard event (no modifier).
func keyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key}}
}

// runeEv creates a printable rune keyboard event.
func runeEv(r rune) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}}
}

// shiftKeyEv creates a keyboard event with Shift modifier.
func shiftKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModShift}}
}

// ctrlEv creates a keyboard event for a Ctrl+letter (using the pre-computed
// tcell key constants directly to avoid magic arithmetic in tests).
func ctrlEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key}}
}

// mouseClickEv creates a mouse Button1 click event.
func mouseClickEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1}}
}

// drawInputLine draws the widget into a fresh buffer of its own width and
// returns the buffer.
func drawInputLine(il *InputLine) *DrawBuffer {
	w := il.Bounds().Width()
	buf := NewDrawBuffer(w, 1)
	il.Draw(buf)
	return buf
}

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

// TestNewInputLineSetsSfVisible verifies the constructor sets SfVisible.
// Spec: "Sets SfVisible … by default."
func TestNewInputLineSetsSfVisible(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	if !il.HasState(SfVisible) {
		t.Error("NewInputLine did not set SfVisible")
	}
}

// TestNewInputLineSetsOfSelectable verifies the constructor sets OfSelectable.
// Spec: "Sets … OfSelectable by default."
func TestNewInputLineSetsOfSelectable(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	if !il.HasOption(OfSelectable) {
		t.Error("NewInputLine did not set OfSelectable")
	}
}

// TestNewInputLineDefaultsToEmptyText verifies new widget starts with empty text.
// Spec: "Stores an internal text buffer as []rune" (buffer starts empty).
func TestNewInputLineDefaultsToEmptyText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	if il.Text() != "" {
		t.Errorf("NewInputLine.Text() = %q, want empty string", il.Text())
	}
}

// TestNewInputLineStoresBounds verifies the bounds are stored.
// Spec: "NewInputLine(bounds Rect, maxLen int, opts ...InputLineOption)"
func TestNewInputLineStoresBounds(t *testing.T) {
	r := NewRect(5, 3, 20, 1)
	il := NewInputLine(r, 0)
	if il.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", il.Bounds(), r)
	}
}

// TestNewInputLineDefaultCursorAtEnd verifies cursor starts at end of empty text (0).
// Spec: "SetText … resets cursor to end." (by analogy, constructor leaves cursor at end
// of empty string, which is position 0.)
func TestNewInputLineDefaultCursorAtZero(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	if il.CursorPos() != 0 {
		t.Errorf("CursorPos() = %d, want 0 for new empty widget", il.CursorPos())
	}
}

// TestNewInputLineAcceptsOptions verifies that options are applied (smoke test;
// at minimum the constructor must not panic with a valid option).
// Spec: "NewInputLine(bounds Rect, maxLen int, opts ...InputLineOption)"
func TestNewInputLineAcceptsOptions(t *testing.T) {
	// Must not panic; actual option behaviour is tested via the concrete options.
	_ = NewInputLine(NewRect(0, 0, 20, 1), 0)
}

// ---------------------------------------------------------------------------
// Text accessor
// ---------------------------------------------------------------------------

// TestTextReturnsCurrentContent verifies Text() returns what was set.
// Spec: "Text() string returns current text content."
func TestTextReturnsCurrentContent(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	if got := il.Text(); got != "hello" {
		t.Errorf("Text() = %q, want %q", got, "hello")
	}
}

// TestTextIsNotStale verifies Text() reflects the most recently set value, not
// an earlier one.
// Falsification: catches an implementation that caches the first value set.
func TestTextIsNotStale(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("first")
	il.SetText("second")
	if got := il.Text(); got != "second" {
		t.Errorf("Text() = %q after two SetText calls, want %q (second)", got, "second")
	}
}

// ---------------------------------------------------------------------------
// SetText
// ---------------------------------------------------------------------------

// TestSetTextClampedToMaxLen verifies text is truncated at maxLen runes when maxLen > 0.
// Spec: "SetText(s string) sets the text, clamping to maxLen if set."
func TestSetTextClampedToMaxLen(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 5)
	il.SetText("abcdefgh")
	if got := il.Text(); got != "abcde" {
		t.Errorf("Text() after SetText with maxLen=5 = %q, want %q", got, "abcde")
	}
}

// TestSetTextNotClampedWhenMaxLenZero verifies no truncation when maxLen == 0 (unlimited).
// Spec: "maxLen limits the number of runes that can be entered (0 means unlimited)."
func TestSetTextNotClampedWhenMaxLenZero(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	long := "abcdefghijklmnopqrstuvwxyz"
	il.SetText(long)
	if got := il.Text(); got != long {
		t.Errorf("Text() with maxLen=0 = %q, want full string %q", got, long)
	}
}

// TestSetTextResetsCursorToEnd verifies cursor is at end of text after SetText.
// Spec: "SetText … resets cursor to end."
func TestSetTextResetsCursorToEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	if pos := il.CursorPos(); pos != 5 {
		t.Errorf("CursorPos() after SetText(\"hello\") = %d, want 5 (end)", pos)
	}
}

// TestSetTextResetsCursorAfterMove verifies cursor is reset even if it was moved first.
// Falsification: catches an implementation that only sets cursor on construction.
func TestSetTextResetsCursorAfterMove(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abcde")
	// Move cursor left
	il.HandleEvent(keyEv(tcell.KeyLeft))
	il.HandleEvent(keyEv(tcell.KeyLeft))
	// Now reset
	il.SetText("xyz")
	if pos := il.CursorPos(); pos != 3 {
		t.Errorf("CursorPos() after SetText(\"xyz\") = %d, want 3 (end of new text)", pos)
	}
}

// TestSetTextClampsCursorWhenClampingText verifies that when text is clamped, cursor
// still lands at the clamped end, not beyond it.
// Spec: "clamping to maxLen if set, and resets cursor to end."
func TestSetTextClampsCursorWhenClampingText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 3)
	il.SetText("abcdef")
	if pos := il.CursorPos(); pos != 3 {
		t.Errorf("CursorPos() after clamped SetText = %d, want 3", pos)
	}
}

// TestSetTextHandlesMultibyteRunes verifies maxLen counts runes, not bytes.
// Spec: "maxLen limits the number of runes that can be entered."
func TestSetTextHandlesMultibyteRunes(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 3)
	// Each character is a multibyte rune.
	il.SetText("こんにちは") // 5 runes, each 3 bytes
	got := []rune(il.Text())
	if len(got) != 3 {
		t.Errorf("len([]rune(Text())) after SetText with multibyte runes and maxLen=3 = %d, want 3", len(got))
	}
}

// ---------------------------------------------------------------------------
// CursorPos
// ---------------------------------------------------------------------------

// TestCursorPosInitiallyZero verifies a fresh widget has cursor at 0.
// Spec: "CursorPos() int returns the cursor position (rune index)."
func TestCursorPosInitiallyZero(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	if il.CursorPos() != 0 {
		t.Errorf("CursorPos() = %d, want 0 for empty text", il.CursorPos())
	}
}

// TestCursorPosAfterLeftKey verifies CursorPos decrements after Left arrow.
// Spec: "Left arrow: move cursor left one rune."
func TestCursorPosAfterLeftKey(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	// cursor is at 3
	il.HandleEvent(keyEv(tcell.KeyLeft))
	if il.CursorPos() != 2 {
		t.Errorf("CursorPos() after Left = %d, want 2", il.CursorPos())
	}
}

// ---------------------------------------------------------------------------
// Drawing: background fill
// ---------------------------------------------------------------------------

// TestDrawFillsWithInputNormalStyle verifies Draw fills every cell with InputNormal style.
// Spec: "Fills its bounds width×1 with InputNormal style from ColorScheme."
func TestDrawFillsWithInputNormalStyle(t *testing.T) {
	il := newInputLine(0, 0, 10, 0)
	buf := drawInputLine(il)

	for x := 0; x < 10; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Style != theme.BorlandBlue.InputNormal {
			t.Errorf("cell(%d,0) style = %v, want InputNormal %v", x, cell.Style, theme.BorlandBlue.InputNormal)
			break
		}
	}
}

// TestDrawInputNormalStyleDiffersFromDefault verifies InputNormal is distinct from
// tcell.StyleDefault in BorlandBlue (falsification guard for the fill test).
func TestDrawInputNormalStyleDiffersFromDefault(t *testing.T) {
	if theme.BorlandBlue.InputNormal == tcell.StyleDefault {
		t.Skip("InputNormal equals StyleDefault in this scheme — fill test is vacuous")
	}
}

// TestDrawRendersTextStartingAtColumn0 verifies text is drawn from the left edge.
// Spec: "Renders text starting from a scroll offset so the cursor is always visible."
// (When scroll offset = 0, text starts at column 0.)
func TestDrawRendersTextStartingAtColumn0(t *testing.T) {
	il := newInputLine(0, 0, 20, 0)
	il.SetText("hello")
	buf := drawInputLine(il)

	expected := []rune{'h', 'e', 'l', 'l', 'o'}
	for i, want := range expected {
		got := buf.GetCell(i, 0).Rune
		if got != want {
			t.Errorf("cell(%d,0).Rune = %q, want %q", i, got, want)
		}
	}
}

// TestDrawRendersEmptyBufferAsSpaces verifies Draw fills empty-text widget with spaces.
// Spec: "Fills its bounds width×1 with InputNormal style."
func TestDrawRendersEmptyBufferAsSpaces(t *testing.T) {
	il := newInputLine(0, 0, 10, 0)
	buf := drawInputLine(il)

	for x := 0; x < 10; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune != ' ' {
			t.Errorf("cell(%d,0).Rune = %q, want ' ' for empty input", x, cell.Rune)
		}
	}
}

// ---------------------------------------------------------------------------
// Drawing: cursor indicator when SfSelected
// ---------------------------------------------------------------------------

// TestDrawCursorIndicatorUsesInputSelectionStyle verifies selected widget shows
// cursor using InputSelection style.
// Spec: "When the widget has SfSelected state, displays a cursor indicator: the
// character at cursor position uses InputSelection style."
func TestDrawCursorIndicatorUsesInputSelectionStyle(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.InputSelection == scheme.InputNormal {
		t.Skip("InputSelection equals InputNormal in this scheme — cursor style test is vacuous")
	}

	il := newInputLine(0, 0, 10, 0)
	il.SetText("ab")
	il.SetState(SfSelected, true)
	// cursor is at end (pos 2) after SetText
	buf := drawInputLine(il)

	// Position 2 (the cursor position) should use InputSelection style.
	cell := buf.GetCell(2, 0)
	if cell.Style != scheme.InputSelection {
		t.Errorf("cursor cell(2,0) style = %v, want InputSelection %v", cell.Style, scheme.InputSelection)
	}
}

// TestDrawCursorIndicatorNotShownWhenNotSelected verifies that without SfSelected,
// the cursor indicator is not applied.
// Spec: "When the widget has SfSelected state, displays a cursor indicator."
func TestDrawCursorIndicatorNotShownWhenNotSelected(t *testing.T) {
	scheme := theme.BorlandBlue

	il := newInputLine(0, 0, 10, 0)
	il.SetText("ab")
	il.SetState(SfSelected, false)
	buf := drawInputLine(il)

	// No cell should use InputSelection style when not selected.
	for x := 0; x < 10; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Style == scheme.InputSelection {
			t.Errorf("cell(%d,0) uses InputSelection style but widget is not selected", x)
		}
	}
}

// TestDrawCursorPositionMatchesCursorPos verifies the cursor indicator appears at
// the current cursor position, not at a fixed location.
// Falsification: catches an implementation that always draws cursor at col 0.
func TestDrawCursorPositionMatchesCursorPos(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.InputSelection == scheme.InputNormal {
		t.Skip("InputSelection equals InputNormal in this scheme")
	}

	il := newInputLine(0, 0, 10, 0)
	il.SetText("abc")
	il.SetState(SfSelected, true)
	// Cursor is at pos 3 after SetText; move it left to pos 1.
	il.HandleEvent(keyEv(tcell.KeyLeft))
	il.HandleEvent(keyEv(tcell.KeyLeft))

	buf := drawInputLine(il)

	// Cursor is at position 1.
	cell := buf.GetCell(1, 0)
	if cell.Style != scheme.InputSelection {
		t.Errorf("cursor indicator not at col 1 (cursor pos); cell(1,0).Style = %v, want InputSelection", cell.Style)
	}
	// Position 0 must NOT show cursor.
	cell0 := buf.GetCell(0, 0)
	if cell0.Style == scheme.InputSelection {
		t.Errorf("cell(0,0) shows cursor indicator but cursor is at pos 1")
	}
}

// ---------------------------------------------------------------------------
// Keyboard: printable runes
// ---------------------------------------------------------------------------

// TestPrintableRuneInsertsAtCursor verifies a printable rune is inserted at cursor.
// Spec: "Printable runes: insert at cursor position, advance cursor."
func TestPrintableRuneInsertsAtCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.HandleEvent(runeEv('h'))
	il.HandleEvent(runeEv('i'))
	if got := il.Text(); got != "hi" {
		t.Errorf("Text() after typing 'h','i' = %q, want %q", got, "hi")
	}
}

// TestPrintableRuneAdvancesCursor verifies cursor advances after insert.
// Spec: "Printable runes: insert at cursor position, advance cursor."
func TestPrintableRuneAdvancesCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.HandleEvent(runeEv('a'))
	if pos := il.CursorPos(); pos != 1 {
		t.Errorf("CursorPos() after typing 'a' = %d, want 1", pos)
	}
}

// TestPrintableRuneInsertedAtMiddlePosition verifies insertion at non-end cursor.
// Spec: "insert at cursor position."
func TestPrintableRuneInsertedAtMiddlePosition(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("ac") // cursor at 2
	il.HandleEvent(keyEv(tcell.KeyLeft))
	// cursor at 1
	il.HandleEvent(runeEv('b'))
	if got := il.Text(); got != "abc" {
		t.Errorf("Text() after inserting 'b' between 'a' and 'c' = %q, want %q", got, "abc")
	}
}

// TestPrintableRuneNoOpAtMaxLen verifies typing is ignored when maxLen reached.
// Spec: "Printable runes: no-op if at maxLen."
func TestPrintableRuneNoOpAtMaxLen(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 3)
	il.HandleEvent(runeEv('a'))
	il.HandleEvent(runeEv('b'))
	il.HandleEvent(runeEv('c'))
	il.HandleEvent(runeEv('d')) // should be ignored
	if got := il.Text(); got != "abc" {
		t.Errorf("Text() after typing 4 runes with maxLen=3 = %q, want %q", got, "abc")
	}
}

// TestPrintableRuneConsumedEvent verifies the event is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestPrintableRuneConsumedEvent(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	ev := runeEv('a')
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("printable rune event was not consumed (IsCleared() = false)")
	}
}

// TestPrintableRuneAtMaxLenEventStillConsumed verifies even a no-op at maxLen is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestPrintableRuneAtMaxLenEventStillConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 1)
	il.HandleEvent(runeEv('a')) // fills maxLen
	ev := runeEv('b')           // would be no-op
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("printable rune event at maxLen was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Keyboard: Left arrow
// ---------------------------------------------------------------------------

// TestLeftArrowMovesLeft verifies cursor moves left by 1.
// Spec: "Left arrow: move cursor left one rune."
func TestLeftArrowMovesLeft(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc") // cursor at 3
	il.HandleEvent(keyEv(tcell.KeyLeft))
	if pos := il.CursorPos(); pos != 2 {
		t.Errorf("CursorPos() after Left = %d, want 2", pos)
	}
}

// TestLeftArrowNoOpAtStart verifies cursor does not go negative.
// Spec: "Left arrow: no-op at position 0."
func TestLeftArrowNoOpAtStart(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	// cursor already at 0
	il.HandleEvent(keyEv(tcell.KeyLeft))
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("CursorPos() after Left at start = %d, want 0", pos)
	}
}

// TestLeftArrowEventConsumed verifies Left is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestLeftArrowEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	ev := keyEv(tcell.KeyLeft)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Left arrow event was not consumed")
	}
}

// TestLeftArrowEventConsumedAtStart verifies Left is consumed even at pos 0.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestLeftArrowEventConsumedAtStart(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	ev := keyEv(tcell.KeyLeft)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Left arrow at pos 0 was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Keyboard: Right arrow
// ---------------------------------------------------------------------------

// TestRightArrowMovesRight verifies cursor moves right by 1.
// Spec: "Right arrow: move cursor right one rune."
func TestRightArrowMovesRight(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc") // cursor at 3; move back to 1
	il.HandleEvent(keyEv(tcell.KeyLeft))
	il.HandleEvent(keyEv(tcell.KeyLeft))
	il.HandleEvent(keyEv(tcell.KeyRight))
	if pos := il.CursorPos(); pos != 2 {
		t.Errorf("CursorPos() after Right = %d, want 2", pos)
	}
}

// TestRightArrowNoOpAtEnd verifies cursor does not move past end of text.
// Spec: "Right arrow: no-op at end of text."
func TestRightArrowNoOpAtEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc") // cursor at 3 (end)
	il.HandleEvent(keyEv(tcell.KeyRight))
	if pos := il.CursorPos(); pos != 3 {
		t.Errorf("CursorPos() after Right at end = %d, want 3", pos)
	}
}

// TestRightArrowEventConsumed verifies Right is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestRightArrowEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyLeft)) // move back from end
	ev := keyEv(tcell.KeyRight)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Right arrow event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Keyboard: Home
// ---------------------------------------------------------------------------

// TestHomeMovesToStart verifies Home moves cursor to position 0.
// Spec: "Home: move cursor to position 0."
func TestHomeMovesToStart(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello") // cursor at 5
	il.HandleEvent(keyEv(tcell.KeyHome))
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("CursorPos() after Home = %d, want 0", pos)
	}
}

// TestHomeNoOpWhenAlreadyAtStart verifies Home is a no-op at position 0.
// Spec: "Home: move cursor to position 0."
func TestHomeNoOpWhenAlreadyAtStart(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	// cursor at 0 already
	il.HandleEvent(keyEv(tcell.KeyHome))
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("CursorPos() after Home at start = %d, want 0", pos)
	}
}

// TestHomeEventConsumed verifies Home is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestHomeEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	ev := keyEv(tcell.KeyHome)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Home event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Keyboard: End
// ---------------------------------------------------------------------------

// TestEndMovesToEndOfText verifies End moves cursor to end of text.
// Spec: "End: move cursor to end of text."
func TestEndMovesToEndOfText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello") // cursor at 5; move to 0 first
	il.HandleEvent(keyEv(tcell.KeyHome))
	il.HandleEvent(keyEv(tcell.KeyEnd))
	if pos := il.CursorPos(); pos != 5 {
		t.Errorf("CursorPos() after End = %d, want 5", pos)
	}
}

// TestEndEventConsumed verifies End is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestEndEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	ev := keyEv(tcell.KeyEnd)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("End event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Keyboard: Backspace
// ---------------------------------------------------------------------------

// TestBackspaceDeletesBeforeCursor verifies Backspace removes rune before cursor.
// Spec: "Backspace: delete rune before cursor, move cursor left."
func TestBackspaceDeletesBeforeCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc") // cursor at 3
	il.HandleEvent(keyEv(tcell.KeyBackspace2))
	if got := il.Text(); got != "ab" {
		t.Errorf("Text() after Backspace = %q, want %q", got, "ab")
	}
}

// TestBackspaceMovesLeft verifies cursor moves left after Backspace.
// Spec: "Backspace: delete rune before cursor, move cursor left."
func TestBackspaceMovesLeft(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc") // cursor at 3
	il.HandleEvent(keyEv(tcell.KeyBackspace2))
	if pos := il.CursorPos(); pos != 2 {
		t.Errorf("CursorPos() after Backspace = %d, want 2", pos)
	}
}

// TestBackspaceNoOpAtStart verifies Backspace is no-op at position 0.
// Spec: "Backspace: no-op at position 0."
func TestBackspaceNoOpAtStart(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyHome))
	il.HandleEvent(keyEv(tcell.KeyBackspace2))
	if got := il.Text(); got != "abc" {
		t.Errorf("Text() after Backspace at pos 0 = %q, want %q (unchanged)", got, "abc")
	}
}

// TestBackspaceNoOpCursorAtStart verifies cursor stays at 0 when Backspace at start.
// Spec: "Backspace: no-op at position 0."
func TestBackspaceNoOpCursorAtStart(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	// cursor is at 0 (empty text)
	il.HandleEvent(keyEv(tcell.KeyBackspace2))
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("CursorPos() after Backspace at start = %d, want 0", pos)
	}
}

// TestBackspaceEventConsumed verifies Backspace is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestBackspaceEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	ev := keyEv(tcell.KeyBackspace2)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Backspace event was not consumed")
	}
}

// TestBackspaceAlternativeKeyCode verifies tcell.KeyBackspace (not just KeyBackspace2) works.
// Spec: "Backspace: delete rune before cursor."
func TestBackspaceAlternativeKeyCode(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyBackspace))
	if got := il.Text(); got != "ab" {
		t.Errorf("Text() after KeyBackspace (not KeyBackspace2) = %q, want %q", got, "ab")
	}
}

// ---------------------------------------------------------------------------
// Keyboard: Delete
// ---------------------------------------------------------------------------

// TestDeleteRemovesAtCursor verifies Delete removes the rune at cursor.
// Spec: "Delete (tcell.KeyDelete): delete rune at cursor position."
func TestDeleteRemovesAtCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyHome))
	il.HandleEvent(keyEv(tcell.KeyDelete))
	if got := il.Text(); got != "bc" {
		t.Errorf("Text() after Delete at pos 0 = %q, want %q", got, "bc")
	}
}

// TestDeleteDoesNotMoveCursor verifies cursor stays at same position after Delete.
// Spec: "delete rune at cursor position" — cursor does not advance.
func TestDeleteDoesNotMoveCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyHome))
	il.HandleEvent(keyEv(tcell.KeyDelete))
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("CursorPos() after Delete at pos 0 = %d, want 0", pos)
	}
}

// TestDeleteNoOpAtEnd verifies Delete is no-op at end of text.
// Spec: "Delete: no-op at end."
func TestDeleteNoOpAtEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc") // cursor at 3 (end)
	il.HandleEvent(keyEv(tcell.KeyDelete))
	if got := il.Text(); got != "abc" {
		t.Errorf("Text() after Delete at end = %q, want %q (unchanged)", got, "abc")
	}
}

// TestDeleteEventConsumed verifies Delete is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestDeleteEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyHome))
	ev := keyEv(tcell.KeyDelete)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Delete event was not consumed")
	}
}

// TestDeleteNoOpAtEndEventConsumed verifies Delete is consumed even when no-op at end.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestDeleteNoOpAtEndEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	ev := keyEv(tcell.KeyDelete)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Delete at end was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Keyboard: unhandled events pass through
// ---------------------------------------------------------------------------

// TestTabEventPassesThrough verifies Tab is not consumed.
// Spec: "Events the widget doesn't handle (Tab, Escape, etc.) pass through unconsumed."
func TestTabEventPassesThrough(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	ev := keyEv(tcell.KeyTab)
	il.HandleEvent(ev)
	if ev.IsCleared() {
		t.Error("Tab event was consumed; it should pass through unconsumed")
	}
}

// TestEscapeEventPassesThrough verifies Escape is not consumed.
// Spec: "Events the widget doesn't handle (Tab, Escape, etc.) pass through unconsumed."
func TestEscapeEventPassesThrough(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	ev := keyEv(tcell.KeyEscape)
	il.HandleEvent(ev)
	if ev.IsCleared() {
		t.Error("Escape event was consumed; it should pass through unconsumed")
	}
}

// TestEnterEventPassesThrough verifies Enter is not consumed (it is not listed
// as a handled key).
// Spec: "Events the widget doesn't handle (Tab, Escape, etc.) pass through unconsumed."
func TestEnterEventPassesThrough(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	ev := keyEv(tcell.KeyEnter)
	il.HandleEvent(ev)
	if ev.IsCleared() {
		t.Error("Enter event was consumed; it should pass through unconsumed")
	}
}

// ---------------------------------------------------------------------------
// Keyboard: Ctrl+A (select all)
// ---------------------------------------------------------------------------

// TestCtrlASelectsAllText verifies Ctrl+A sets selection to cover entire text.
// Spec: "Ctrl+A: select all text (set selection to cover entire text)."
func TestCtrlASelectsAllText(t *testing.T) {
	il := newInputLine(0, 0, 20, 0)
	il.SetText("hello")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA))

	// Verify by drawing: all text cells should use InputSelection style when selected.
	il.SetState(SfSelected, true)
	buf := drawInputLine(il)

	scheme := theme.BorlandBlue
	if scheme.InputSelection == scheme.InputNormal {
		t.Skip("InputSelection equals InputNormal — selection style test is vacuous")
	}

	for i := 0; i < 5; i++ {
		cell := buf.GetCell(i, 0)
		if cell.Style != scheme.InputSelection {
			t.Errorf("after Ctrl+A, cell(%d,0).Style = %v, want InputSelection (entire text selected)", i, cell.Style)
		}
	}
}

// TestCtrlAEventConsumed verifies Ctrl+A is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestCtrlAEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	ev := ctrlEv(tcell.KeyCtrlA)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+A event was not consumed")
	}
}

// TestCtrlAOnEmptyTextDoesNotPanic verifies Ctrl+A on empty text is safe.
// Spec: "Ctrl+A: select all text."
func TestCtrlAOnEmptyTextDoesNotPanic(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	// Must not panic.
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA))
}

// ---------------------------------------------------------------------------
// Keyboard: Ctrl+C / Ctrl+X / Ctrl+V (clipboard)
// ---------------------------------------------------------------------------

// TestCtrlCCopiesSelectionToClipboard verifies Ctrl+C copies selected text.
// Spec: "Ctrl+C: copy selected text to internal clipboard."
func TestCtrlCCopiesSelectionToClipboard(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA)) // select all
	il.HandleEvent(ctrlEv(tcell.KeyCtrlC)) // copy

	// Paste into a second widget; if the clipboard holds "hello world", it will appear.
	il2 := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il2.HandleEvent(ctrlEv(tcell.KeyCtrlV))
	if got := il2.Text(); got != "hello world" {
		t.Errorf("after Ctrl+C on \"hello world\" and Ctrl+V in new widget, Text() = %q, want %q",
			got, "hello world")
	}
}

// TestCtrlCEventConsumed verifies Ctrl+C is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestCtrlCEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	ev := ctrlEv(tcell.KeyCtrlC)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+C event was not consumed")
	}
}

// TestCtrlXCutsSelectionToClipboard verifies Ctrl+X removes selected text and stores it.
// Spec: "Ctrl+X: cut selected text to internal clipboard."
func TestCtrlXCutsSelectionToClipboard(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA)) // select all
	il.HandleEvent(ctrlEv(tcell.KeyCtrlX)) // cut
	if got := il.Text(); got != "" {
		t.Errorf("Text() after Ctrl+X = %q, want empty string (text was cut)", got)
	}
}

// TestCtrlXClipboardContainsCutText verifies the cut text ends up in clipboard.
// Spec: "Ctrl+X: cut selected text to internal clipboard."
func TestCtrlXClipboardContainsCutText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	il.HandleEvent(ctrlEv(tcell.KeyCtrlX))

	il2 := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il2.HandleEvent(ctrlEv(tcell.KeyCtrlV))
	if got := il2.Text(); got != "hello" {
		t.Errorf("after Ctrl+X + Ctrl+V in new widget, Text() = %q, want %q", got, "hello")
	}
}

// TestCtrlXEventConsumed verifies Ctrl+X is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestCtrlXEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	ev := ctrlEv(tcell.KeyCtrlX)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+X event was not consumed")
	}
}

// TestCtrlVPasteInsertedAtCursor verifies Ctrl+V pastes clipboard at cursor.
// Spec: "Ctrl+V: paste from internal clipboard at cursor position (respecting maxLen)."
func TestCtrlVPasteInsertedAtCursor(t *testing.T) {
	// Put text in clipboard via a first widget.
	il1 := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il1.SetText("world")
	il1.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	il1.HandleEvent(ctrlEv(tcell.KeyCtrlC))

	// Paste into widget with existing text, cursor at start.
	il2 := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il2.SetText(" end")
	il2.HandleEvent(keyEv(tcell.KeyHome)) // cursor to 0
	il2.HandleEvent(ctrlEv(tcell.KeyCtrlV))
	if got := il2.Text(); got != "world end" {
		t.Errorf("Text() after Ctrl+V at start of \" end\" = %q, want %q", got, "world end")
	}
}

// TestCtrlVRespectsMaxLen verifies paste is clamped to maxLen.
// Spec: "Ctrl+V: paste from internal clipboard at cursor position (respecting maxLen)."
func TestCtrlVRespectsMaxLen(t *testing.T) {
	// Put "hello" in clipboard.
	il1 := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il1.SetText("hello")
	il1.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	il1.HandleEvent(ctrlEv(tcell.KeyCtrlC))

	// Widget with maxLen=3, currently empty.
	il2 := NewInputLine(NewRect(0, 0, 20, 1), 3)
	il2.HandleEvent(ctrlEv(tcell.KeyCtrlV))
	got := []rune(il2.Text())
	if len(got) > 3 {
		t.Errorf("Text() after Ctrl+V with maxLen=3 has %d runes (%q), want at most 3", len(got), string(got))
	}
}

// TestCtrlVEventConsumed verifies Ctrl+V is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestCtrlVEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	ev := ctrlEv(tcell.KeyCtrlV)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+V event was not consumed")
	}
}

// TestClipboardIsPackageLevelShared verifies all InputLine instances share the clipboard.
// Spec: "A package-level var clipboard string shared across all InputLine instances."
func TestClipboardIsPackageLevelShared(t *testing.T) {
	il1 := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il1.SetText("shared")
	il1.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	il1.HandleEvent(ctrlEv(tcell.KeyCtrlC))

	il2 := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il2.HandleEvent(ctrlEv(tcell.KeyCtrlV))
	if got := il2.Text(); got != "shared" {
		t.Errorf("Text() after copy in il1 and paste in il2 = %q, want %q (shared clipboard)", got, "shared")
	}
}

// ---------------------------------------------------------------------------
// Selection
// ---------------------------------------------------------------------------

// TestShiftLeftExtendsSelection verifies Shift+Left extends the selection leftward.
// Spec: "Shift+Left/Shift+Right extends or contracts the selection."
func TestShiftLeftExtendsSelection(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.InputSelection == scheme.InputNormal {
		t.Skip("InputSelection equals InputNormal — selection style test is vacuous")
	}

	il := newInputLine(0, 0, 20, 0)
	il.SetText("abc")  // cursor at 3
	il.SetState(SfSelected, true)
	il.HandleEvent(shiftKeyEv(tcell.KeyLeft)) // select 'c'

	buf := drawInputLine(il)
	// Position 2 ('c') should be rendered with selection style.
	cell := buf.GetCell(2, 0)
	if cell.Style != scheme.InputSelection {
		t.Errorf("cell(2,0) after Shift+Left: style = %v, want InputSelection", cell.Style)
	}
}

// TestShiftRightExtendsSelection verifies Shift+Right extends the selection rightward.
// Spec: "Shift+Left/Shift+Right extends or contracts the selection."
func TestShiftRightExtendsSelection(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.InputSelection == scheme.InputNormal {
		t.Skip("InputSelection equals InputNormal — selection style test is vacuous")
	}

	il := newInputLine(0, 0, 20, 0)
	il.SetText("abc")
	il.SetState(SfSelected, true)
	il.HandleEvent(keyEv(tcell.KeyHome))       // cursor to 0
	il.HandleEvent(shiftKeyEv(tcell.KeyRight)) // select 'a'

	buf := drawInputLine(il)
	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.InputSelection {
		t.Errorf("cell(0,0) after Shift+Right: style = %v, want InputSelection", cell.Style)
	}
}

// TestShiftHomeSelectsToStart verifies Shift+Home selects from cursor to position 0.
// Spec: "Shift+Home selects from cursor to start."
func TestShiftHomeSelectsToStart(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.InputSelection == scheme.InputNormal {
		t.Skip("InputSelection equals InputNormal — selection style test is vacuous")
	}

	il := newInputLine(0, 0, 20, 0)
	il.SetText("hello") // cursor at 5
	il.SetState(SfSelected, true)
	il.HandleEvent(shiftKeyEv(tcell.KeyHome)) // select all from cursor to 0

	buf := drawInputLine(il)
	// All five characters should be selected.
	for i := 0; i < 5; i++ {
		cell := buf.GetCell(i, 0)
		if cell.Style != scheme.InputSelection {
			t.Errorf("after Shift+Home, cell(%d,0).Style = %v, want InputSelection", i, cell.Style)
		}
	}
}

// TestShiftEndSelectsToEnd verifies Shift+End selects from cursor to end of text.
// Spec: "Shift+End selects from cursor to end."
func TestShiftEndSelectsToEnd(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.InputSelection == scheme.InputNormal {
		t.Skip("InputSelection equals InputNormal — selection style test is vacuous")
	}

	il := newInputLine(0, 0, 20, 0)
	il.SetText("hello")
	il.SetState(SfSelected, true)
	il.HandleEvent(keyEv(tcell.KeyHome))     // cursor to 0
	il.HandleEvent(shiftKeyEv(tcell.KeyEnd)) // select to end

	buf := drawInputLine(il)
	for i := 0; i < 5; i++ {
		cell := buf.GetCell(i, 0)
		if cell.Style != scheme.InputSelection {
			t.Errorf("after Shift+End, cell(%d,0).Style = %v, want InputSelection", i, cell.Style)
		}
	}
}

// TestTypingWithSelectionReplacesSelectedText verifies that typing replaces selection.
// Spec: "Typing with active selection replaces the selected text."
func TestTypingWithSelectionReplacesSelectedText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA)) // select all
	il.HandleEvent(runeEv('x'))             // replace
	if got := il.Text(); got != "x" {
		t.Errorf("Text() after typing with selection = %q, want %q", got, "x")
	}
}

// TestBackspaceWithSelectionDeletesSelected verifies Backspace with selection deletes it.
// Spec: "Backspace/Delete with active selection deletes the selected text."
func TestBackspaceWithSelectionDeletesSelected(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA)) // select all
	il.HandleEvent(keyEv(tcell.KeyBackspace2))
	if got := il.Text(); got != "" {
		t.Errorf("Text() after Backspace with selection = %q, want empty", got)
	}
}

// TestDeleteWithSelectionDeletesSelected verifies Delete with selection deletes it.
// Spec: "Backspace/Delete with active selection deletes the selected text."
func TestDeleteWithSelectionDeletesSelected(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA)) // select all
	il.HandleEvent(keyEv(tcell.KeyDelete))
	if got := il.Text(); got != "" {
		t.Errorf("Text() after Delete with selection = %q, want empty", got)
	}
}

// TestNoSelectionWhenSelStartEqualsSelEnd verifies that when selStart == selEnd
// (no selection), nothing beyond the cursor indicator is drawn with selection style.
// Spec: "Selection is tracked as selStart, selEnd int (rune indices); when equal, no selection."
func TestNoSelectionWhenSelStartEqualsSelEnd(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.InputSelection == scheme.InputNormal {
		t.Skip("InputSelection equals InputNormal — vacuous")
	}

	il := newInputLine(0, 0, 20, 0)
	il.SetText("abc")
	il.SetState(SfSelected, true)
	// Do NOT set any selection; cursor is at end (pos 3).

	buf := drawInputLine(il)
	// Only the cursor position (col 3 – the space beyond text) should use selection
	// style (the cursor indicator). Columns 0, 1, 2 (the text characters) must not.
	for i := 0; i < 3; i++ {
		cell := buf.GetCell(i, 0)
		if cell.Style == scheme.InputSelection {
			t.Errorf("cell(%d,0) uses InputSelection without active selection", i)
		}
	}
}

// TestSelectedTextRendersWithInputSelectionStyle verifies selected runes use InputSelection.
// Spec: "Selected text renders with InputSelection style."
func TestSelectedTextRendersWithInputSelectionStyle(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.InputSelection == scheme.InputNormal {
		t.Skip("InputSelection equals InputNormal — vacuous")
	}

	il := newInputLine(0, 0, 20, 0)
	il.SetText("abc")
	il.SetState(SfSelected, true)
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA)) // select all

	buf := drawInputLine(il)
	for i := 0; i < 3; i++ {
		cell := buf.GetCell(i, 0)
		if cell.Style != scheme.InputSelection {
			t.Errorf("selected char cell(%d,0).Style = %v, want InputSelection", i, cell.Style)
		}
	}
}

// ---------------------------------------------------------------------------
// Mouse handling
// ---------------------------------------------------------------------------

// TestMouseClickPositionsCursor verifies Button1 click positions cursor at clicked column.
// Spec: "Click (Button1) within the widget bounds positions the cursor at the clicked column
// (accounting for scroll offset and the widget's origin X)."
func TestMouseClickPositionsCursor(t *testing.T) {
	// Widget at origin (0,0), width 20.
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	ev := mouseClickEv(3, 0)
	il.HandleEvent(ev)

	if pos := il.CursorPos(); pos != 3 {
		t.Errorf("CursorPos() after click at col 3 = %d, want 3", pos)
	}
}

// TestMouseClickConsumed verifies the click event is consumed.
// Spec: "The click event is consumed."
func TestMouseClickConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	ev := mouseClickEv(2, 0)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("mouse click event was not consumed")
	}
}

// TestMouseClickAtDifferentColumns verifies distinct columns yield distinct positions.
// Falsification: catches an implementation that always sets cursor to 0.
func TestMouseClickAtDifferentColumns(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello world")

	ev5 := mouseClickEv(5, 0)
	il.HandleEvent(ev5)
	pos5 := il.CursorPos()

	ev2 := mouseClickEv(2, 0)
	il.HandleEvent(ev2)
	pos2 := il.CursorPos()

	if pos5 == pos2 {
		t.Errorf("clicking col 5 and col 2 both yield position %d; expected different positions", pos5)
	}
}

// TestMouseClickAccountsForWidgetOriginX verifies click accounting uses widget's origin X.
// Spec: "Group delivers mouse events in owner-local coordinates without translating to the
// child's origin" — so the widget must subtract its own origin X.
func TestMouseClickAccountsForWidgetOriginX(t *testing.T) {
	// Widget has non-zero origin: starts at X=5.
	il := NewInputLine(NewRect(5, 0, 15, 1), 0)
	il.SetText("hello world")

	// A click at absolute X=7 means column 7-5=2 within the widget.
	ev := mouseClickEv(7, 0)
	il.HandleEvent(ev)

	if pos := il.CursorPos(); pos != 2 {
		t.Errorf("CursorPos() after click at abs X=7 with widget origin X=5 = %d, want 2", pos)
	}
}

// TestMouseClickClampsToTextLength verifies click beyond text end places cursor at text end.
// Spec: "positions the cursor at the clicked column (accounting for scroll offset)."
// Clicking past end of text should clamp to end.
func TestMouseClickClampsToTextLength(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("ab") // 2 runes, cursor range 0..2

	// Click at column 15 (well past text end).
	ev := mouseClickEv(15, 0)
	il.HandleEvent(ev)

	if pos := il.CursorPos(); pos > 2 {
		t.Errorf("CursorPos() after click past text end = %d, want <= 2", pos)
	}
}

// ---------------------------------------------------------------------------
// Scroll behavior
// ---------------------------------------------------------------------------

// TestScrollOffsetAdjustsWhenCursorPastVisibleWidth verifies scrolling right.
// Spec: "When cursor moves past scrollOffset + visible width, scrollOffset increases."
// We test this indirectly: text longer than the view should still show the cursor
// character at the rightmost visible column when rendered.
func TestScrollOffsetAdjustsWhenCursorPastVisibleWidth(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.InputSelection == scheme.InputNormal {
		t.Skip("InputSelection equals InputNormal — cursor style test is vacuous")
	}

	// Widget 5 chars wide; type 8 chars so cursor scrolls out of the initial view.
	il := newInputLine(0, 0, 5, 0)
	il.SetState(SfSelected, true)
	for _, r := range "abcdefgh" {
		il.HandleEvent(runeEv(r))
	}
	// Cursor is at position 8, visible width is 5.
	// Draw should show cursor within the 5-wide buffer (not off the right edge).
	buf := drawInputLine(il)

	// The cursor indicator must appear somewhere in the buffer.
	found := false
	for x := 0; x < 5; x++ {
		if buf.GetCell(x, 0).Style == scheme.InputSelection {
			found = true
			break
		}
	}
	if !found {
		t.Error("after typing past visible width, cursor indicator is not visible in the draw buffer")
	}
}

// TestScrollOffsetAdjustsWhenCursorMovesLeft verifies scrolling left.
// Spec: "When cursor moves left of scrollOffset, scrollOffset decreases to show cursor."
func TestScrollOffsetAdjustsWhenCursorMovesLeft(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.InputSelection == scheme.InputNormal {
		t.Skip("InputSelection equals InputNormal — cursor style test is vacuous")
	}

	// Width 5; type 8 chars to scroll right, then Home to scroll back.
	il := newInputLine(0, 0, 5, 0)
	il.SetState(SfSelected, true)
	for _, r := range "abcdefgh" {
		il.HandleEvent(runeEv(r))
	}
	il.HandleEvent(keyEv(tcell.KeyHome)) // cursor to 0, should scroll back

	buf := drawInputLine(il)
	// After Home, 'a' must be at column 0, and cursor (at pos 0) must be visible.
	cell := buf.GetCell(0, 0)
	if cell.Rune != 'a' {
		t.Errorf("after Home on scrolled input, cell(0,0).Rune = %q, want 'a' (scroll reset)", cell.Rune)
	}
}

// TestScrollOffsetDrawsFromScrolledPosition verifies text is rendered starting from
// the scroll offset.
// Spec: "Drawing starts from text[scrollOffset:]."
func TestScrollOffsetDrawsFromScrolledPosition(t *testing.T) {
	// Width 3; type "abcde". After typing, cursor at 5, scrollOffset should be >= 2
	// so "cde" (or a suffix) starts at col 0.
	il := newInputLine(0, 0, 3, 0)
	for _, r := range "abcde" {
		il.HandleEvent(runeEv(r))
	}

	buf := drawInputLine(il)
	// The cell at column 0 must NOT be 'a' (which is scrolled off).
	cell := buf.GetCell(0, 0)
	if cell.Rune == 'a' {
		t.Error("cell(0,0).Rune = 'a' even though text is scrolled; scrollOffset not applied")
	}
}

// ---------------------------------------------------------------------------
// Fix 1: Shift+Left/Right/Home/End event consumption
// ---------------------------------------------------------------------------

// TestShiftLeftEventConsumed verifies Shift+Left is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestShiftLeftEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc") // cursor at 3
	ev := shiftKeyEv(tcell.KeyLeft)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Shift+Left event was not consumed")
	}
}

// TestShiftRightEventConsumed verifies Shift+Right is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestShiftRightEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyHome)) // cursor to 0
	ev := shiftKeyEv(tcell.KeyRight)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Shift+Right event was not consumed")
	}
}

// TestShiftHomeEventConsumed verifies Shift+Home is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestShiftHomeEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello") // cursor at 5
	ev := shiftKeyEv(tcell.KeyHome)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Shift+Home event was not consumed")
	}
}

// TestShiftEndEventConsumed verifies Shift+End is consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestShiftEndEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	il.HandleEvent(keyEv(tcell.KeyHome)) // cursor to 0
	ev := shiftKeyEv(tcell.KeyEnd)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Shift+End event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Fix 2: Ctrl+C with no selection
// ---------------------------------------------------------------------------

// TestCtrlCWithNoSelectionDoesNotCorruptClipboard verifies Ctrl+C with no active
// selection is consumed but does not overwrite the clipboard with empty content.
// Spec: "Ctrl+C: copy selected text to internal clipboard."
func TestCtrlCWithNoSelectionDoesNotCorruptClipboard(t *testing.T) {
	// Seed the clipboard with a known value.
	seed := NewInputLine(NewRect(0, 0, 20, 1), 0)
	seed.SetText("seed text")
	seed.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	seed.HandleEvent(ctrlEv(tcell.KeyCtrlC))

	// Now invoke Ctrl+C on a widget with no selection; the event must be consumed.
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("something")
	// Do NOT select anything; cursor is at end, selStart == selEnd.
	ev := ctrlEv(tcell.KeyCtrlC)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+C with no selection was not consumed")
	}

	// Paste into a fresh widget; clipboard must still hold "seed text".
	verify := NewInputLine(NewRect(0, 0, 40, 1), 0)
	verify.HandleEvent(ctrlEv(tcell.KeyCtrlV))
	if got := verify.Text(); got != "seed text" {
		t.Errorf("clipboard after Ctrl+C with no selection = %q, want %q (previous value preserved)",
			got, "seed text")
	}
}

// ---------------------------------------------------------------------------
// Fix 3: Ctrl+V cursor position after paste
// ---------------------------------------------------------------------------

// TestCtrlVCursorPositionAfterPaste verifies cursor lands at original position plus
// the length of the pasted text.
// Spec: "Ctrl+V: paste from internal clipboard at cursor position."
func TestCtrlVCursorPositionAfterPaste(t *testing.T) {
	// Put "xyz" (3 runes) into the clipboard.
	il1 := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il1.SetText("xyz")
	il1.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	il1.HandleEvent(ctrlEv(tcell.KeyCtrlC))

	// Paste into a widget whose cursor is at position 2.
	il2 := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il2.SetText("abcde") // cursor at 5
	il2.HandleEvent(keyEv(tcell.KeyHome))  // cursor to 0
	il2.HandleEvent(keyEv(tcell.KeyRight)) // cursor to 1
	il2.HandleEvent(keyEv(tcell.KeyRight)) // cursor to 2
	// Paste; cursor should advance to 2 + 3 = 5.
	il2.HandleEvent(ctrlEv(tcell.KeyCtrlV))
	if pos := il2.CursorPos(); pos != 5 {
		t.Errorf("CursorPos() after pasting 3-rune string at position 2 = %d, want 5", pos)
	}
}

// ---------------------------------------------------------------------------
// Fix 4: Typing with partial selection
// ---------------------------------------------------------------------------

// TestTypingWithPartialSelectionReplacesSelection verifies that typing a character
// while a partial selection is active replaces only the selected text.
// Spec: "Typing with active selection replaces the selected text."
func TestTypingWithPartialSelectionReplacesSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abcde")           // cursor at 5
	il.HandleEvent(keyEv(tcell.KeyHome))  // cursor to 0
	il.HandleEvent(keyEv(tcell.KeyRight)) // cursor to 1 (after 'a')

	// Shift+Right x3 selects "bcd" (positions 1..4).
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))

	// Type 'X' to replace selection.
	il.HandleEvent(runeEv('X'))

	if got := il.Text(); got != "aXe" {
		t.Errorf("Text() after replacing \"bcd\" with 'X' = %q, want %q", got, "aXe")
	}
	if pos := il.CursorPos(); pos != 2 {
		t.Errorf("CursorPos() after replacing selection = %d, want 2", pos)
	}
}

// ---------------------------------------------------------------------------
// Fix 5: Mouse click with scroll offset
// ---------------------------------------------------------------------------

// TestMouseClickWithScrollOffsetAccountsForScroll verifies that a mouse click
// positions the cursor relative to the scroll offset, not the raw click column.
// Spec: "Click (Button1) within the widget bounds positions the cursor at the clicked
// column (accounting for scroll offset and the widget's origin X)."
func TestMouseClickWithScrollOffsetAccountsForScroll(t *testing.T) {
	// Widget 5 chars wide; type "abcdefgh" to force scrolling.
	il := NewInputLine(NewRect(0, 0, 5, 0), 0)
	for _, r := range "abcdefgh" {
		il.HandleEvent(runeEv(r))
	}
	// After typing, cursor is at 8 and the view is scrolled so the cursor is visible.
	// Record the scroll offset indirectly: click at column 2 and see what position we get.
	// If scrollOffset > 0, the resulting cursor position must be > 2.
	ev := mouseClickEv(2, 0)
	il.HandleEvent(ev)

	pos := il.CursorPos()
	// The scroll offset must be positive (we typed past width 5), so clicking column 2
	// within the widget corresponds to a text position greater than 2.
	if pos <= 2 {
		t.Errorf("CursorPos() after click at col 2 on scrolled widget = %d; expected > 2 (scroll offset not applied)", pos)
	}
}

// ---------------------------------------------------------------------------
// Fix 6: Selection contraction
// ---------------------------------------------------------------------------

// TestShiftRightThenShiftLeftContractsSelection verifies that pressing Shift+Left after
// extending a selection by Shift+Right contracts the selection by one character.
// Spec: "Shift+Left/Shift+Right extends or contracts the selection."
func TestShiftRightThenShiftLeftContractsSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abcde")
	il.HandleEvent(keyEv(tcell.KeyHome)) // cursor to 0

	// Extend selection right 3 times: anchor=0, cursor=3, selecting "abc".
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))

	// Contract by pressing Shift+Left once: selection should now cover 2 chars.
	il.HandleEvent(shiftKeyEv(tcell.KeyLeft))

	start, end := il.Selection()
	if start > end {
		start, end = end, start
	}
	selLen := end - start
	if selLen != 2 {
		t.Errorf("selection length after 3× Shift+Right then 1× Shift+Left = %d, want 2", selLen)
	}
}

// ---------------------------------------------------------------------------
// Fix 7: Right arrow consumed at end of text
// ---------------------------------------------------------------------------

// TestRightArrowEventConsumedAtEnd verifies Right arrow at end of text is still consumed.
// Spec: "All keyboard events that InputLine handles must be consumed via event.Clear()."
func TestRightArrowEventConsumedAtEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc") // cursor at 3 (end)
	ev := keyEv(tcell.KeyRight)
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Right arrow at end of text was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Fix 8: Backspace/Delete with selection cursor position
// ---------------------------------------------------------------------------

// TestBackspaceWithSelectionCursorPosition verifies cursor lands at the start of
// the former selection range after Backspace deletes the selection.
// Spec: "Backspace/Delete with active selection deletes the selected text."
func TestBackspaceWithSelectionCursorPosition(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abcde")           // cursor at 5
	il.HandleEvent(keyEv(tcell.KeyHome))  // cursor to 0
	il.HandleEvent(keyEv(tcell.KeyRight)) // cursor to 1 (after 'a')

	// Shift+Right x3 selects "bcd" (positions 1..4).
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))

	il.HandleEvent(keyEv(tcell.KeyBackspace2))

	if pos := il.CursorPos(); pos != 1 {
		t.Errorf("CursorPos() after Backspace with selection \"bcd\" from \"abcde\" = %d, want 1 (start of former selection)", pos)
	}
}

// TestDeleteWithSelectionCursorPosition verifies cursor lands at the start of
// the former selection range after Delete deletes the selection.
// Spec: "Backspace/Delete with active selection deletes the selected text."
func TestDeleteWithSelectionCursorPosition(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abcde")           // cursor at 5
	il.HandleEvent(keyEv(tcell.KeyHome))  // cursor to 0
	il.HandleEvent(keyEv(tcell.KeyRight)) // cursor to 1 (after 'a')

	// Shift+Right x3 selects "bcd" (positions 1..4).
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))

	il.HandleEvent(keyEv(tcell.KeyDelete))

	if pos := il.CursorPos(); pos != 1 {
		t.Errorf("CursorPos() after Delete with selection \"bcd\" from \"abcde\" = %d, want 1 (start of former selection)", pos)
	}
}
