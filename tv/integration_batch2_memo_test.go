package tv

// integration_batch2_memo_test.go — Integration tests for TMemo core (Tasks 2-5).
//
// Verifies that the Memo subsystems — text editing, cursor movement, drawing,
// and viewport scrolling — work together as a coherent whole.
//
// Requirements covered:
//   Req 1  — Memo in a Window: keyboard events through Group dispatch reach the Memo
//   Req 2  — SetText then cursor movement produces correct positions
//   Req 3  — Drawing after text editing shows edited content
//   Req 4  — Viewport scrolling: cursor at row 15 in a 5-row Memo adjusts deltaY
//   Req 5  — Auto-indent: Enter after an indented line creates new indented line
//   Req 6  — Backspace at line start joins with previous line
//   Req 7  — Delete at line end joins with next line
//   Req 8  — Ctrl+Y deletes entire line

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Req 1 — Memo in a Window: keyboard events through Group dispatch reach Memo
// ---------------------------------------------------------------------------

// TestIntegrationMemoInWindowReceivesKeyboardEvents verifies that a Memo inserted
// into a Window responds to keyboard events dispatched to the Window.
// Wiring: Window.Insert(memo) → memo becomes focused child of the Window's Group.
// Window.HandleEvent(keyboard) → dispatches to group → to focused Memo.
func TestIntegrationMemoInWindowReceivesKeyboardEvents(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 30, 12), "Test")
	memo := NewMemo(NewRect(0, 0, 28, 10))
	win.Insert(memo)
	memo.SetText("hello")

	// Type 'X' via Window — must reach the Memo
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'X'}}
	win.HandleEvent(ev)

	if got := memo.Text(); got != "Xhello" {
		t.Errorf("After typing 'X' via Window: Text() = %q, want %q", got, "Xhello")
	}
}

// TestIntegrationMemoInWindowCursorMovesOnArrowKey verifies that arrow-key events
// dispatched to the Window update the Memo's cursor position.
func TestIntegrationMemoInWindowCursorMovesOnArrowKey(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 30, 12), "Test")
	memo := NewMemo(NewRect(0, 0, 28, 10))
	win.Insert(memo)
	memo.SetText("hello")

	// Send Right key via Window — cursor should advance
	win.HandleEvent(memoKeyEv(tcell.KeyRight))

	row, col := memo.CursorPos()
	if row != 0 || col != 1 {
		t.Errorf("After Right via Window: CursorPos() = (%d, %d), want (0, 1)", row, col)
	}
}

// TestIntegrationMemoInWindowKeyboardEventIsConsumed verifies the keyboard event
// is consumed (cleared) after the Memo handles it, confirming the event actually
// propagated through the Group's dispatch chain.
func TestIntegrationMemoInWindowKeyboardEventIsConsumed(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 30, 12), "Test")
	memo := NewMemo(NewRect(0, 0, 28, 10))
	win.Insert(memo)
	memo.SetText("hello")

	ev := memoKeyEv(tcell.KeyRight)
	win.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("KeyRight dispatched via Window: event should be cleared after Memo handles it")
	}
}

// TestIntegrationMemoInWindowTypeMultipleChars verifies multiple characters typed
// through the Window accumulate in the Memo correctly.
func TestIntegrationMemoInWindowTypeMultipleChars(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 30, 12), "Test")
	memo := NewMemo(NewRect(0, 0, 28, 10))
	win.Insert(memo)

	for _, r := range "abc" {
		ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}}
		win.HandleEvent(ev)
	}

	if got := memo.Text(); got != "abc" {
		t.Errorf("After typing \"abc\" via Window: Text() = %q, want %q", got, "abc")
	}
}

// ---------------------------------------------------------------------------
// Req 2 — SetText then cursor movement produces correct positions
// ---------------------------------------------------------------------------

// TestIntegrationSetTextThenMoveToEnd verifies that after SetText the cursor
// starts at (0,0) and can be moved to the end of the line.
func TestIntegrationSetTextThenMoveToEnd(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello world")

	// Move cursor to end of line
	m.HandleEvent(memoKeyEv(tcell.KeyEnd))

	row, col := m.CursorPos()
	if row != 0 || col != 11 {
		t.Errorf("After SetText+End: CursorPos() = (%d, %d), want (0, 11)", row, col)
	}
}

// TestIntegrationSetTextMultiLineThenNavigate verifies that SetText with multiple
// lines and subsequent cursor navigation produces correct positions.
func TestIntegrationSetTextMultiLineThenNavigate(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond\nthird")

	// Move to last line, end of line
	m.HandleEvent(memoCtrlKeyEv(tcell.KeyEnd))

	row, col := m.CursorPos()
	if row != 2 || col != 5 {
		t.Errorf("After SetText+Ctrl+End: CursorPos() = (%d, %d), want (2, 5)", row, col)
	}
}

// TestIntegrationSetTextThenMoveThenSetTextAgain verifies that SetText resets
// cursor to (0,0) regardless of prior movement.
func TestIntegrationSetTextThenMoveThenSetTextAgain(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	// Move cursor away
	repeatKey(m, tcell.KeyRight, 3)
	row, col := m.CursorPos()
	if row != 0 || col != 3 {
		t.Fatalf("Pre-condition: CursorPos() = (%d, %d), want (0, 3)", row, col)
	}

	// Second SetText must reset cursor
	m.SetText("world")
	row, col = m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("After second SetText: CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// TestIntegrationSetTextCursorMovementPreservesTextIntegrity verifies that
// moving the cursor does not alter the text content.
func TestIntegrationSetTextCursorMovementPreservesTextIntegrity(t *testing.T) {
	m := newTestMemo()
	m.SetText("line one\nline two")

	// Navigate around
	m.HandleEvent(memoKeyEv(tcell.KeyDown))
	repeatKey(m, tcell.KeyRight, 5)
	m.HandleEvent(memoKeyEv(tcell.KeyHome))
	m.HandleEvent(memoCtrlKeyEv(tcell.KeyHome))

	if got := m.Text(); got != "line one\nline two" {
		t.Errorf("After cursor navigation: Text() = %q, want %q (navigation must not alter text)", got, "line one\nline two")
	}
}

// ---------------------------------------------------------------------------
// Req 3 — Drawing after text editing shows edited content
// ---------------------------------------------------------------------------

// TestIntegrationDrawShowsTypedText verifies that typing characters and then
// calling Draw renders the updated text content.
func TestIntegrationDrawShowsTypedText(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 3))
	m.scheme = theme.BorlandBlue

	// Type "Hi"
	m.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'H'}})
	m.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'i'}})

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	if buf.GetCell(0, 0).Rune != 'H' {
		t.Errorf("Draw after typing 'H': cell(0,0).Rune = %q, want 'H'", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(1, 0).Rune != 'i' {
		t.Errorf("Draw after typing 'i': cell(1,0).Rune = %q, want 'i'", buf.GetCell(1, 0).Rune)
	}
}

// TestIntegrationDrawAfterBackspaceShowsResult verifies that Backspace removes
// the character and Draw reflects the change.
func TestIntegrationDrawAfterBackspaceShowsResult(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("abc")
	// Move to end and delete 'c'
	m.HandleEvent(memoKeyEv(tcell.KeyEnd))
	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	if buf.GetCell(0, 0).Rune != 'a' {
		t.Errorf("Draw after backspace: cell(0,0) = %q, want 'a'", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(1, 0).Rune != 'b' {
		t.Errorf("Draw after backspace: cell(1,0) = %q, want 'b'", buf.GetCell(1, 0).Rune)
	}
	// Cell 2 should be space (fill) since 'c' was deleted
	if buf.GetCell(2, 0).Rune != ' ' {
		t.Errorf("Draw after backspace: cell(2,0) = %q, want ' ' (fill after deletion)", buf.GetCell(2, 0).Rune)
	}
}

// TestIntegrationDrawAfterEnterShowsNewLine verifies that pressing Enter splits
// the line and Draw shows the second line on row 1.
func TestIntegrationDrawAfterEnterShowsNewLine(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("AB")
	// Move to col 1 and press Enter
	m.HandleEvent(memoKeyEv(tcell.KeyRight))
	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	// Row 0: 'A'
	if buf.GetCell(0, 0).Rune != 'A' {
		t.Errorf("Draw after Enter: cell(0,0) = %q, want 'A'", buf.GetCell(0, 0).Rune)
	}
	// Row 1: 'B'
	if buf.GetCell(0, 1).Rune != 'B' {
		t.Errorf("Draw after Enter: cell(0,1) = %q, want 'B'", buf.GetCell(0, 1).Rune)
	}
}

// TestIntegrationDrawAfterDeleteShowsJoinedLine verifies that Delete at end of
// line joins lines and Draw reflects the join.
func TestIntegrationDrawAfterDeleteShowsJoinedLine(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("AB\nCD")
	// Move to end of first line and delete
	m.HandleEvent(memoKeyEv(tcell.KeyEnd))
	m.HandleEvent(memoKeyEv(tcell.KeyDelete))

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	// Row 0 should show "ABCD"
	want := "ABCD"
	for i, ch := range want {
		if buf.GetCell(i, 0).Rune != ch {
			t.Errorf("Draw after Delete join: cell(%d,0) = %q, want %q", i, buf.GetCell(i, 0).Rune, ch)
		}
	}
	// Row 1 should be spaces (second line was merged up)
	if buf.GetCell(0, 1).Rune != ' ' {
		t.Errorf("Draw after Delete join: cell(0,1) = %q, want ' ' (row 1 now empty)", buf.GetCell(0, 1).Rune)
	}
}

// ---------------------------------------------------------------------------
// Req 4 — Viewport scrolling: cursor at row 15 in a 5-row Memo adjusts deltaY
// ---------------------------------------------------------------------------

// TestIntegrationViewportScrollingCursorPastViewport verifies that when text
// with 20 lines is loaded and the cursor is moved to row 15 in a 5-row Memo,
// Draw shows a line near row 15 at the top of the buffer (row 0), not line 0.
func TestIntegrationViewportScrollingCursorPastViewport(t *testing.T) {
	// 5-row Memo
	m := NewMemo(NewRect(0, 0, 20, 5))
	m.scheme = theme.BorlandBlue

	// Build 20 lines where each line's content encodes its row number
	var lines []string
	for i := 0; i < 20; i++ {
		switch i {
		case 0:
			lines = append(lines, "line-zero")
		case 15:
			lines = append(lines, "line-fifteen")
		default:
			lines = append(lines, "line-other")
		}
	}
	m.SetText(strings.Join(lines, "\n"))

	// Move cursor to row 15 via Down key
	repeatKey(m, tcell.KeyDown, 15)

	row, _ := m.CursorPos()
	if row != 15 {
		t.Fatalf("Pre-condition: cursor should be at row 15, got row %d", row)
	}

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	// Row 0 of the DrawBuffer should NOT show "line-zero" (which starts with 'l','i','n','e','-','z')
	// It should show something near row 15 instead.
	// "line-zero" has 'z' at position 5; "line-fifteen" has 'f' at position 5.
	// "line-other" has 'o' at position 5.
	// "line-zero" has unique char 'z' at col 5.
	cell := buf.GetCell(5, 0)
	if cell.Rune == 'z' {
		t.Errorf("Viewport scroll: after cursor at row 15, Draw row 0 shows 'z' (line-zero), "+
			"want a line near row 15 (viewport should have scrolled down)")
	}
}

// TestIntegrationViewportScrollingCursorVisibleInDraw verifies that after moving
// to row 15 in a 5-row Memo, the cursor's line is visible somewhere within the
// 5 rendered rows.
func TestIntegrationViewportScrollingCursorVisibleInDraw(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 5))
	m.scheme = theme.BorlandBlue

	// 20 lines; row 15 has unique content "FIFTEEN"
	var lines []string
	for i := 0; i < 20; i++ {
		if i == 15 {
			lines = append(lines, "FIFTEEN")
		} else {
			lines = append(lines, "other")
		}
	}
	m.SetText(strings.Join(lines, "\n"))
	repeatKey(m, tcell.KeyDown, 15)

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	// "FIFTEEN" should appear on one of the 5 visible rows
	found := false
	for bufRow := 0; bufRow < 5; bufRow++ {
		if buf.GetCell(0, bufRow).Rune == 'F' &&
			buf.GetCell(1, bufRow).Rune == 'I' &&
			buf.GetCell(2, bufRow).Rune == 'F' {
			found = true
			break
		}
	}
	if !found {
		t.Error("Viewport scroll: cursor line \"FIFTEEN\" is not visible in any of the 5 Draw rows; deltaY not adjusted")
	}
}

// TestIntegrationViewportScrollingCtrlHome verifies that after scrolling down,
// Ctrl+Home brings the viewport back to line 0.
func TestIntegrationViewportScrollingCtrlHome(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 5))
	m.scheme = theme.BorlandBlue

	// 20 lines; row 0 has unique content "ZERO"
	var lines []string
	lines = append(lines, "ZERO")
	for i := 1; i < 20; i++ {
		lines = append(lines, "other")
	}
	m.SetText(strings.Join(lines, "\n"))

	// Scroll down to row 15, then jump back to origin
	repeatKey(m, tcell.KeyDown, 15)
	m.HandleEvent(memoCtrlKeyEv(tcell.KeyHome))

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	// "ZERO" should now be visible at row 0
	if buf.GetCell(0, 0).Rune != 'Z' || buf.GetCell(1, 0).Rune != 'E' {
		t.Errorf("After Ctrl+Home: Draw row 0 starts with %q%q, want 'Z','E' (\"ZERO\")",
			buf.GetCell(0, 0).Rune, buf.GetCell(1, 0).Rune)
	}
}

// ---------------------------------------------------------------------------
// Req 5 — Auto-indent: Enter after an indented line creates new indented line
// ---------------------------------------------------------------------------

// TestIntegrationAutoIndentEnterCopiesLeadingSpaces verifies that pressing Enter
// at the end of a line with 4 leading spaces creates a new line that also starts
// with 4 spaces, and the cursor is placed after the indent.
func TestIntegrationAutoIndentEnterCopiesLeadingSpaces(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10), WithAutoIndent(true))
	m.SetText("    code") // 4 leading spaces

	// Move to end of line
	m.HandleEvent(memoKeyEv(tcell.KeyEnd))
	// Press Enter
	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Fatalf("Auto-indent: want 2 lines after Enter, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[1], "    ") {
		t.Errorf("Auto-indent: new line = %q, want 4 leading spaces", lines[1])
	}

	// Cursor should be at row 1, col 4 (after the indent)
	row, col := m.CursorPos()
	if row != 1 || col != 4 {
		t.Errorf("Auto-indent: CursorPos() = (%d, %d), want (1, 4)", row, col)
	}
}

// TestIntegrationAutoIndentDrawShowsIndentedLine verifies that after auto-indented
// Enter, Draw renders the indent spaces on the new line.
func TestIntegrationAutoIndentDrawShowsIndentedLine(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 5), WithAutoIndent(true))
	m.scheme = theme.BorlandBlue
	m.SetText("    hello")

	// End of line, then Enter
	m.HandleEvent(memoKeyEv(tcell.KeyEnd))
	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	buf := NewDrawBuffer(40, 5)
	m.Draw(buf)

	// Row 1 should show 4 spaces at columns 0-3
	for col := 0; col < 4; col++ {
		if buf.GetCell(col, 1).Rune != ' ' {
			t.Errorf("Auto-indent draw: cell(%d,1) = %q, want ' ' (indent space)", col, buf.GetCell(col, 1).Rune)
		}
	}
}

// TestIntegrationAutoIndentDisabledNoIndent verifies that with auto-indent
// disabled, Enter does not copy leading spaces.
func TestIntegrationAutoIndentDisabledNoIndent(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10), WithAutoIndent(false))
	m.SetText("    code")
	m.HandleEvent(memoKeyEv(tcell.KeyEnd))
	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Fatalf("No auto-indent: want 2 lines, got %d", len(lines))
	}
	if lines[1] != "" {
		t.Errorf("No auto-indent: new line = %q, want empty string", lines[1])
	}
}

// ---------------------------------------------------------------------------
// Req 6 — Backspace at line start joins with previous line
// ---------------------------------------------------------------------------

// TestIntegrationBackspaceAtLineStartJoinsLines verifies that pressing Backspace
// at the start of a non-first line joins it with the previous line.
func TestIntegrationBackspaceAtLineStartJoinsLines(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld")

	// Move to start of line 1
	m.HandleEvent(memoKeyEv(tcell.KeyDown))
	// Cursor is at (1, 0)
	row, col := m.CursorPos()
	if row != 1 || col != 0 {
		t.Fatalf("Pre-condition: CursorPos() = (%d, %d), want (1, 0)", row, col)
	}

	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))

	if got := m.Text(); got != "helloworld" {
		t.Errorf("Backspace at line start: Text() = %q, want %q", got, "helloworld")
	}
}

// TestIntegrationBackspaceAtLineStartCursorAtJoinPoint verifies the cursor is
// placed at the join point (end of the original previous line).
func TestIntegrationBackspaceAtLineStartCursorAtJoinPoint(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld") // "hello" is 5 chars

	m.HandleEvent(memoKeyEv(tcell.KeyDown))
	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))

	row, col := m.CursorPos()
	if row != 0 {
		t.Errorf("Backspace join: cursor row = %d, want 0", row)
	}
	if col != 5 {
		t.Errorf("Backspace join: cursor col = %d, want 5 (join point)", col)
	}
}

// TestIntegrationBackspaceAtLineStartReducesLineCount verifies the line count
// decreases by 1 when lines are joined.
func TestIntegrationBackspaceAtLineStartReducesLineCount(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond\nthird")

	// Move to start of line 2
	repeatKey(m, tcell.KeyDown, 2)

	before := len(strings.Split(m.Text(), "\n"))
	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))
	after := len(strings.Split(m.Text(), "\n"))

	if after != before-1 {
		t.Errorf("Backspace join: line count %d → %d, want %d", before, after, before-1)
	}
}

// ---------------------------------------------------------------------------
// Req 7 — Delete at line end joins with next line
// ---------------------------------------------------------------------------

// TestIntegrationDeleteAtLineEndJoinsLines verifies that Delete at the end of
// a line joins it with the next line.
func TestIntegrationDeleteAtLineEndJoinsLines(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld")

	// Move to end of line 0
	m.HandleEvent(memoKeyEv(tcell.KeyEnd))
	m.HandleEvent(memoKeyEv(tcell.KeyDelete))

	if got := m.Text(); got != "helloworld" {
		t.Errorf("Delete at line end: Text() = %q, want %q", got, "helloworld")
	}
}

// TestIntegrationDeleteAtLineEndCursorStaysAtJoinPoint verifies the cursor stays
// at the join point (end of original first line) after the join.
func TestIntegrationDeleteAtLineEndCursorStaysAtJoinPoint(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld") // "hello" is 5 chars

	m.HandleEvent(memoKeyEv(tcell.KeyEnd))
	m.HandleEvent(memoKeyEv(tcell.KeyDelete))

	row, col := m.CursorPos()
	if row != 0 {
		t.Errorf("Delete join: cursor row = %d, want 0", row)
	}
	if col != 5 {
		t.Errorf("Delete join: cursor col = %d, want 5 (join point)", col)
	}
}

// TestIntegrationDeleteAtLineEndDrawShowsJoinedLine verifies that after Delete
// joins two lines, Draw renders the joined content.
func TestIntegrationDeleteAtLineEndDrawShowsJoinedLine(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("AB\nCD")

	m.HandleEvent(memoKeyEv(tcell.KeyEnd))
	m.HandleEvent(memoKeyEv(tcell.KeyDelete))

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	want := "ABCD"
	for i, ch := range want {
		if buf.GetCell(i, 0).Rune != ch {
			t.Errorf("Delete join draw: cell(%d,0) = %q, want %q", i, buf.GetCell(i, 0).Rune, ch)
		}
	}
}

// ---------------------------------------------------------------------------
// Req 8 — Ctrl+Y deletes entire line
// ---------------------------------------------------------------------------

// TestIntegrationCtrlYDeletesLine verifies Ctrl+Y removes the current line.
func TestIntegrationCtrlYDeletesLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("alpha\nbeta\ngamma")
	m.HandleEvent(memoKeyEv(tcell.KeyDown)) // row 1 "beta"

	m.HandleEvent(memoKeyEv(tcell.KeyCtrlY))

	if got := m.Text(); got != "alpha\ngamma" {
		t.Errorf("Ctrl+Y: Text() = %q, want %q", got, "alpha\ngamma")
	}
}

// TestIntegrationCtrlYLineCountDecreases verifies Ctrl+Y reduces line count by 1.
func TestIntegrationCtrlYLineCountDecreases(t *testing.T) {
	m := newTestMemo()
	m.SetText("one\ntwo\nthree")

	before := len(strings.Split(m.Text(), "\n"))
	m.HandleEvent(memoKeyEv(tcell.KeyCtrlY))
	after := len(strings.Split(m.Text(), "\n"))

	if after != before-1 {
		t.Errorf("Ctrl+Y: line count %d → %d, want %d", before, after, before-1)
	}
}

// TestIntegrationCtrlYDrawShowsUpdatedContent verifies that after Ctrl+Y, Draw
// renders the content without the deleted line.
func TestIntegrationCtrlYDrawShowsUpdatedContent(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("alpha\nbeta\ngamma")
	// Delete row 0 "alpha"
	m.HandleEvent(memoKeyEv(tcell.KeyCtrlY))

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	// Row 0 should now show "beta"
	if buf.GetCell(0, 0).Rune != 'b' {
		t.Errorf("Ctrl+Y draw: cell(0,0) = %q, want 'b' (start of \"beta\")", buf.GetCell(0, 0).Rune)
	}
}

// TestIntegrationCtrlYOnOnlyLineEmptiesContent verifies Ctrl+Y on the only line
// clears it to empty (one empty line remains).
func TestIntegrationCtrlYOnOnlyLineEmptiesContent(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	m.HandleEvent(memoKeyEv(tcell.KeyCtrlY))

	if got := m.Text(); got != "" {
		t.Errorf("Ctrl+Y on only line: Text() = %q, want empty string", got)
	}
	// Cursor must still be valid
	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Ctrl+Y on only line: CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// TestIntegrationCtrlYViaWindowDispatch verifies Ctrl+Y dispatched through a
// Window reaches the Memo and removes the correct line.
func TestIntegrationCtrlYViaWindowDispatch(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 30, 12), "Test")
	memo := NewMemo(NewRect(0, 0, 28, 10))
	win.Insert(memo)
	memo.SetText("first\nsecond\nthird")

	// Send Ctrl+Y via Window — should delete "first" (row 0, the default cursor position)
	ev := memoKeyEv(tcell.KeyCtrlY)
	win.HandleEvent(ev)

	if got := memo.Text(); got != "second\nthird" {
		t.Errorf("Ctrl+Y via Window: Text() = %q, want %q", got, "second\nthird")
	}
}
