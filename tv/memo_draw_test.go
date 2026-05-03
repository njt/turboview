package tv

// memo_draw_test.go — Tests for Task 3: Memo Draw().
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the exact spec sentence it verifies.
//
// Spec authority: Task 3 "Memo Drawing" + section 3.6.
//
// Test organisation:
//   Section 1  — Text content rendered at correct column positions
//   Section 2  — MemoNormal style applied to text content
//   Section 3  — Fill spaces for short lines (trailing columns)
//   Section 4  — MemoNormal style applied to fill spaces
//   Section 5  — Rows beyond last line filled with spaces
//   Section 6  — MemoNormal style applied to beyond-last-line rows
//   Section 7  — Multi-line rendering
//   Section 8  — Edge cases (empty Memo, exact-width line)
//   Section 9  — Viewport scrolling stubs (require Task 4)

import (
	"testing"

	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Section 1 — Text content rendered at correct column positions
// ---------------------------------------------------------------------------

// TestDrawSingleCharAppearsAtColumn0 verifies that the first character of a line
// is rendered at column 0 in the default viewport (deltaX=0, deltaY=0).
// Spec: "render lines[lineIdx] starting from column deltaX" (deltaX=0 at default).
func TestDrawSingleCharAppearsAtColumn0(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("A")

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != 'A' {
		t.Errorf("Draw: cell(0,0).Rune = %q, want 'A'; first character must appear at column 0 (deltaX=0)", cell.Rune)
	}
}

// TestDrawLineContentRenderedConsecutively verifies that characters of a line
// are rendered at consecutive columns starting from 0.
// Spec: "render lines[lineIdx] starting from column deltaX" with deltaX=0.
func TestDrawLineContentRenderedConsecutively(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("Hello")

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	want := "Hello"
	for i, ch := range want {
		cell := buf.GetCell(i, 0)
		if cell.Rune != ch {
			t.Errorf("Draw: cell(%d,0).Rune = %q, want %q", i, cell.Rune, ch)
		}
	}
}

// TestDrawFirstLineAppearsAtRow0 verifies that the first line of text is rendered
// on row 0 of the buffer (default viewport, deltaY=0).
// Spec: "For each row y in [0, height): lineIdx = deltaY + y" (deltaY=0 means lineIdx=y).
func TestDrawFirstLineAppearsAtRow0(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("X")

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	// The first line "X" must be at row 0.
	cell := buf.GetCell(0, 0)
	if cell.Rune != 'X' {
		t.Errorf("Draw: cell(0,0).Rune = %q, want 'X'; first line must render on row 0", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Section 2 — MemoNormal style applied to text content
// ---------------------------------------------------------------------------

// TestDrawTextContentUsesMemoNormalStyle verifies that text characters are
// rendered with the MemoNormal style from the ColorScheme.
// Spec: "render line content starting from column deltaX in MemoNormal".
func TestDrawTextContentUsesMemoNormalStyle(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 3))
	scheme := theme.BorlandBlue
	m.scheme = scheme
	m.SetText("Hello")

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.MemoNormal {
		t.Errorf("Draw: cell(0,0).Style = %v, want MemoNormal %v; text characters must use MemoNormal", cell.Style, scheme.MemoNormal)
	}
}

// TestDrawTextContentAllCharactersUseMemoNormalStyle verifies that every
// character of the rendered line uses MemoNormal style.
// Spec: "render line content starting from column deltaX in MemoNormal".
func TestDrawTextContentAllCharactersUseMemoNormalStyle(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 3))
	scheme := theme.BorlandBlue
	m.scheme = scheme
	m.SetText("Hello")

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	for i := range "Hello" {
		cell := buf.GetCell(i, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Draw: cell(%d,0).Style = %v, want MemoNormal %v", i, cell.Style, scheme.MemoNormal)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Fill spaces for short lines (trailing columns)
// ---------------------------------------------------------------------------

// TestDrawShortLineTrailingColumnsAreSpaces verifies that columns after the
// end of a line's content are filled with spaces.
// Spec: "After line content ends, fill remaining columns with spaces in MemoNormal".
func TestDrawShortLineTrailingColumnsAreSpaces(t *testing.T) {
	// Line "Hi" is 2 chars; widget is 10 wide — columns 2-9 must be spaces.
	m := NewMemo(NewRect(0, 0, 10, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("Hi")

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	for col := 2; col < 10; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("Draw: cell(%d,0).Rune = %q, want ' '; trailing columns after line content must be spaces", col, cell.Rune)
		}
	}
}

// TestDrawShortLineFillStartsAfterContent verifies the fill begins immediately
// after the last character of line content.
// Spec: "After line content ends, fill remaining columns with spaces".
func TestDrawShortLineFillStartsAfterContent(t *testing.T) {
	// "A" is 1 char at col 0; col 1 must be the fill space.
	m := NewMemo(NewRect(0, 0, 10, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("A")

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Rune != ' ' {
		t.Errorf("Draw: cell(1,0).Rune = %q, want ' '; fill must start at column 1 (immediately after 'A')", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Section 4 — MemoNormal style applied to fill spaces
// ---------------------------------------------------------------------------

// TestDrawTrailingFillSpacesUseMemoNormalStyle verifies that the fill spaces
// after line content use MemoNormal style.
// Spec: "fill remaining columns with spaces in MemoNormal".
func TestDrawTrailingFillSpacesUseMemoNormalStyle(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 3))
	scheme := theme.BorlandBlue
	m.scheme = scheme
	m.SetText("Hi")

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	// Column 5 is well within the trailing fill area (line is 2 chars wide).
	cell := buf.GetCell(5, 0)
	if cell.Style != scheme.MemoNormal {
		t.Errorf("Draw: cell(5,0) fill space style = %v, want MemoNormal %v", cell.Style, scheme.MemoNormal)
	}
}

// TestDrawTrailingFillAllColumnsUseMemoNormalStyle verifies every trailing fill
// column uses MemoNormal, not the zero/default style.
// Spec: "fill remaining columns with spaces in MemoNormal".
func TestDrawTrailingFillAllColumnsUseMemoNormalStyle(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 3))
	scheme := theme.BorlandBlue
	m.scheme = scheme
	m.SetText("Hi") // 2 chars; cols 2-9 are fill

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	for col := 2; col < 10; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Draw: cell(%d,0) fill style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Rows beyond last line filled with spaces
// ---------------------------------------------------------------------------

// TestDrawRowBeyondLastLineFilledWithSpaces verifies that rows with no
// corresponding line (lineIdx >= len(lines)) are filled with spaces.
// Spec: "If lineIdx >= len(lines): fill row with spaces in MemoNormal".
func TestDrawRowBeyondLastLineFilledWithSpaces(t *testing.T) {
	// 1 line of text, widget height=3 → rows 1 and 2 have no line.
	m := NewMemo(NewRect(0, 0, 10, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("Hello")

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	// Row 1: lineIdx = deltaY(0) + 1 = 1 >= len(lines)(1) → must be all spaces.
	for col := 0; col < 10; col++ {
		cell := buf.GetCell(col, 1)
		if cell.Rune != ' ' {
			t.Errorf("Draw: cell(%d,1).Rune = %q, want ' '; row beyond last line must be filled with spaces", col, cell.Rune)
		}
	}
}

// TestDrawRowBeyondLastLineFillsEntireRow verifies the entire row width is
// filled when there is no corresponding line.
// Spec: "If lineIdx >= len(lines): fill row with spaces in MemoNormal".
func TestDrawRowBeyondLastLineFillsEntireRow(t *testing.T) {
	// 2 lines, widget height=5 → rows 2, 3, 4 have no line.
	m := NewMemo(NewRect(0, 0, 8, 5))
	m.scheme = theme.BorlandBlue
	m.SetText("one\ntwo")

	buf := NewDrawBuffer(8, 5)
	m.Draw(buf)

	for row := 2; row < 5; row++ {
		for col := 0; col < 8; col++ {
			cell := buf.GetCell(col, row)
			if cell.Rune != ' ' {
				t.Errorf("Draw: cell(%d,%d).Rune = %q, want ' '; row %d is beyond last line (index 1)", col, row, cell.Rune, row)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Section 6 — MemoNormal style applied to beyond-last-line rows
// ---------------------------------------------------------------------------

// TestDrawBeyondLastLineRowUsesMemoNormalStyle verifies that the fill spaces
// for rows beyond the last line use MemoNormal style.
// Spec: "If lineIdx >= len(lines): fill row with spaces in MemoNormal".
func TestDrawBeyondLastLineRowUsesMemoNormalStyle(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 3))
	scheme := theme.BorlandBlue
	m.scheme = scheme
	m.SetText("Hello") // 1 line; rows 1 and 2 are beyond.

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	// Row 1, col 0 — first cell of first beyond-last-line row.
	cell := buf.GetCell(0, 1)
	if cell.Style != scheme.MemoNormal {
		t.Errorf("Draw: cell(0,1) beyond-last-line row style = %v, want MemoNormal %v", cell.Style, scheme.MemoNormal)
	}
}

// TestDrawBeyondLastLineAllColumnsUseMemoNormalStyle verifies every column of
// every beyond-last-line row uses MemoNormal.
// Spec: "If lineIdx >= len(lines): fill row with spaces in MemoNormal".
func TestDrawBeyondLastLineAllColumnsUseMemoNormalStyle(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 6, 4))
	scheme := theme.BorlandBlue
	m.scheme = scheme
	m.SetText("one\ntwo") // 2 lines; rows 2 and 3 are beyond.

	buf := NewDrawBuffer(6, 4)
	m.Draw(buf)

	for row := 2; row < 4; row++ {
		for col := 0; col < 6; col++ {
			cell := buf.GetCell(col, row)
			if cell.Style != scheme.MemoNormal {
				t.Errorf("Draw: cell(%d,%d) beyond-last-line style = %v, want MemoNormal %v", col, row, cell.Style, scheme.MemoNormal)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Multi-line rendering
// ---------------------------------------------------------------------------

// TestDrawMultiLineEachLineAppearsOnCorrectRow verifies that each line of text
// is rendered on the correct row.
// Spec: "For each row y in [0, height): lineIdx = deltaY + y" with deltaY=0.
func TestDrawMultiLineEachLineAppearsOnCorrectRow(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("first\nsecond\nthird")

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	// Row 0 → line 0 ("first")
	if buf.GetCell(0, 0).Rune != 'f' {
		t.Errorf("Draw: cell(0,0).Rune = %q, want 'f' (start of \"first\")", buf.GetCell(0, 0).Rune)
	}
	// Row 1 → line 1 ("second")
	if buf.GetCell(0, 1).Rune != 's' {
		t.Errorf("Draw: cell(0,1).Rune = %q, want 's' (start of \"second\")", buf.GetCell(0, 1).Rune)
	}
	// Row 2 → line 2 ("third")
	if buf.GetCell(0, 2).Rune != 't' {
		t.Errorf("Draw: cell(0,2).Rune = %q, want 't' (start of \"third\")", buf.GetCell(0, 2).Rune)
	}
}

// TestDrawMultiLineLinesAreRenderedIndependently verifies that each visible
// line is rendered independently — content from one line does not bleed to
// adjacent rows.
// Spec: "Each visible line is rendered independently".
func TestDrawMultiLineLinesAreRenderedIndependently(t *testing.T) {
	// "AB" on row 0, "C" on row 1 — confirm col 1 on row 1 is a space (fill),
	// not 'B' from row 0.
	m := NewMemo(NewRect(0, 0, 10, 2))
	m.scheme = theme.BorlandBlue
	m.SetText("AB\nC")

	buf := NewDrawBuffer(10, 2)
	m.Draw(buf)

	// Row 0: 'A' at col 0, 'B' at col 1.
	if buf.GetCell(0, 0).Rune != 'A' {
		t.Errorf("Draw: cell(0,0).Rune = %q, want 'A'", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(1, 0).Rune != 'B' {
		t.Errorf("Draw: cell(1,0).Rune = %q, want 'B'", buf.GetCell(1, 0).Rune)
	}

	// Row 1: 'C' at col 0; col 1 must be fill space (not 'B' from row 0).
	if buf.GetCell(0, 1).Rune != 'C' {
		t.Errorf("Draw: cell(0,1).Rune = %q, want 'C'", buf.GetCell(0, 1).Rune)
	}
	if buf.GetCell(1, 1).Rune != ' ' {
		t.Errorf("Draw: cell(1,1).Rune = %q, want ' '; each line is independent — col 1 of row 1 must be fill, not 'B' from row 0", buf.GetCell(1, 1).Rune)
	}
}

// TestDrawMultiLineWithDifferentLineLengths verifies correct rendering when
// lines have different lengths.
// Spec: "No word wrap. Lines longer than the view width are scrolled horizontally."
// Spec: "After line content ends, fill remaining columns with spaces in MemoNormal".
func TestDrawMultiLineWithDifferentLineLengths(t *testing.T) {
	// "Hi" (2 chars), "Longer" (6 chars), "" (0 chars).
	m := NewMemo(NewRect(0, 0, 10, 3))
	m.scheme = theme.BorlandBlue
	m.SetText("Hi\nLonger\n")

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	// Row 0: 'H','i' then fill.
	if buf.GetCell(0, 0).Rune != 'H' {
		t.Errorf("Draw: row0 col0 = %q, want 'H'", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(2, 0).Rune != ' ' {
		t.Errorf("Draw: row0 col2 = %q, want ' ' (fill after \"Hi\")", buf.GetCell(2, 0).Rune)
	}

	// Row 1: 'L','o','n','g','e','r' then fill.
	if buf.GetCell(0, 1).Rune != 'L' {
		t.Errorf("Draw: row1 col0 = %q, want 'L'", buf.GetCell(0, 1).Rune)
	}
	if buf.GetCell(6, 1).Rune != ' ' {
		t.Errorf("Draw: row1 col6 = %q, want ' ' (fill after \"Longer\")", buf.GetCell(6, 1).Rune)
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Edge cases
// ---------------------------------------------------------------------------

// TestDrawEmptyMemoAllRowsFilledWithSpaces verifies that a Memo with no text
// (after NewMemo) fills all rows with spaces.
// Spec: "If lineIdx >= len(lines): fill row with spaces in MemoNormal".
// An empty Memo has 1 line ("") so row 0 renders an empty line (all fill);
// rows 1+ are beyond last line.
func TestDrawEmptyMemoAllRowsFilledWithSpaces(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 3))
	m.scheme = theme.BorlandBlue
	// No SetText — Memo starts with one empty line.

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	for row := 0; row < 3; row++ {
		for col := 0; col < 10; col++ {
			cell := buf.GetCell(col, row)
			if cell.Rune != ' ' {
				t.Errorf("Draw (empty Memo): cell(%d,%d).Rune = %q, want ' '", col, row, cell.Rune)
			}
		}
	}
}

// TestDrawEmptyMemoAllRowsUseMemoNormalStyle verifies that an empty Memo
// fills all rows with MemoNormal style.
// Spec: "fill row with spaces in MemoNormal" / "fill remaining columns with spaces in MemoNormal".
func TestDrawEmptyMemoAllRowsUseMemoNormalStyle(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 3))
	scheme := theme.BorlandBlue
	m.scheme = scheme

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	for row := 0; row < 3; row++ {
		for col := 0; col < 10; col++ {
			cell := buf.GetCell(col, row)
			if cell.Style != scheme.MemoNormal {
				t.Errorf("Draw (empty Memo): cell(%d,%d).Style = %v, want MemoNormal %v", col, row, cell.Style, scheme.MemoNormal)
			}
		}
	}
}

// TestDrawSingleLineExactWidthNoTrailingFill verifies that when a line's
// length exactly matches the widget width, there are no trailing fill spaces
// needed (all columns are occupied by content).
// Spec: "After line content ends, fill remaining columns with spaces in MemoNormal".
// — When no remaining columns exist, this is vacuously satisfied; the test
// confirms no panic and content renders correctly.
func TestDrawSingleLineExactWidthContent(t *testing.T) {
	// "ABCDE" is 5 chars; widget is 5 wide — no trailing fill needed.
	m := NewMemo(NewRect(0, 0, 5, 2))
	m.scheme = theme.BorlandBlue
	m.SetText("ABCDE")

	buf := NewDrawBuffer(5, 2)
	m.Draw(buf)

	want := "ABCDE"
	for i, ch := range want {
		cell := buf.GetCell(i, 0)
		if cell.Rune != ch {
			t.Errorf("Draw (exact-width line): cell(%d,0).Rune = %q, want %q", i, cell.Rune, ch)
		}
	}
}

// TestDrawSingleLineExactWidthStyleOnAllCells verifies style is MemoNormal
// on all cells when the line exactly fills the widget width.
// Spec: "render line content starting from column deltaX in MemoNormal".
func TestDrawSingleLineExactWidthStyleOnAllCells(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 5, 2))
	scheme := theme.BorlandBlue
	m.scheme = scheme
	m.SetText("ABCDE")

	buf := NewDrawBuffer(5, 2)
	m.Draw(buf)

	for col := 0; col < 5; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Draw (exact-width line): cell(%d,0).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// TestDrawWidgetMoreRowsThanLines verifies that a widget with more rows than
// text lines fills the extra rows with spaces in MemoNormal.
// Spec: "If lineIdx >= len(lines): fill row with spaces in MemoNormal".
func TestDrawWidgetMoreRowsThanLines(t *testing.T) {
	// 2 lines of text, widget has 5 rows → rows 2-4 are beyond last line.
	m := NewMemo(NewRect(0, 0, 10, 5))
	scheme := theme.BorlandBlue
	m.scheme = scheme
	m.SetText("one\ntwo")

	buf := NewDrawBuffer(10, 5)
	m.Draw(buf)

	for row := 2; row < 5; row++ {
		for col := 0; col < 10; col++ {
			cell := buf.GetCell(col, row)
			if cell.Rune != ' ' {
				t.Errorf("Draw: cell(%d,%d).Rune = %q, want ' ' (widget has more rows than lines)", col, row, cell.Rune)
			}
			if cell.Style != scheme.MemoNormal {
				t.Errorf("Draw: cell(%d,%d).Style = %v, want MemoNormal %v", col, row, cell.Style, scheme.MemoNormal)
			}
		}
	}
}

// TestDrawSingleCharOnLineWithFillAfter verifies that a single character is at
// column 0 and all remaining columns are filled with spaces.
// Spec: "render lines[lineIdx] starting from column deltaX" and
// "After line content ends, fill remaining columns with spaces in MemoNormal".
func TestDrawSingleCharOnLineWithFillAfter(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 5, 1))
	scheme := theme.BorlandBlue
	m.scheme = scheme
	m.SetText("Z")

	buf := NewDrawBuffer(5, 1)
	m.Draw(buf)

	// 'Z' at col 0.
	if buf.GetCell(0, 0).Rune != 'Z' {
		t.Errorf("Draw: cell(0,0).Rune = %q, want 'Z'", buf.GetCell(0, 0).Rune)
	}
	// Cols 1-4 are fill spaces.
	for col := 1; col < 5; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("Draw: cell(%d,0).Rune = %q, want ' ' (fill after single char)", col, cell.Rune)
		}
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Draw: cell(%d,0).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// TestDrawLineLongerThanWidthClipsToWidgetWidth verifies that when a line is
// longer than the widget width, only the first `width` characters are visible
// and no trailing fill spaces appear.
// Spec: "render lines[lineIdx] starting from column deltaX" — at deltaX=0,
// columns [0, width) show line[0..width-1]; excess characters are not rendered.
func TestDrawLineLongerThanWidthClipsToWidgetWidth(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 5, 1))
	scheme := theme.BorlandBlue
	m.scheme = scheme
	m.SetText("ABCDEFGHIJ") // 10 chars, widget is 5 wide

	buf := NewDrawBuffer(5, 1)
	m.Draw(buf)

	want := "ABCDE"
	for i, ch := range want {
		cell := buf.GetCell(i, 0)
		if cell.Rune != ch {
			t.Errorf("Draw (long line): cell(%d,0).Rune = %q, want %q", i, cell.Rune, ch)
		}
	}
}

// TestDrawCursorPositionNotVisuallyRendered verifies that the cursor position
// is NOT visually rendered by Draw — the cell at cursor position uses MemoNormal,
// not MemoSelected or any other distinctive style.
// Spec: "The cursor position is NOT visually rendered by Draw".
func TestDrawCursorPositionNotVisuallyRendered(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 3))
	scheme := theme.BorlandBlue
	m.scheme = scheme
	m.SetText("Hello")

	buf := NewDrawBuffer(10, 3)
	m.Draw(buf)

	// Cursor is at (0,0) after SetText. Cell at (0,0) must use MemoNormal.
	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.MemoNormal {
		t.Errorf("Draw: cell at cursor position (0,0) Style = %v, want MemoNormal %v; cursor must NOT be visually rendered", cell.Style, scheme.MemoNormal)
	}
	if cell.Style == scheme.MemoSelected {
		t.Errorf("Draw: cell at cursor position (0,0) uses MemoSelected; cursor must NOT be visually rendered")
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Viewport scrolling stubs (require Task 4: cursor movement)
// ---------------------------------------------------------------------------

// NOTE: The following tests CANNOT be written yet because deltaX and deltaY
// are private fields set only by cursor movement (Task 4). Once Task 4 is
// implemented and cursor movement causes viewport scrolling, the following
// behaviours should be tested:
//
// TestDrawDeltaXShiftsTextLeft:
//   When deltaX > 0, "text shifts left (first visible column is deltaX)".
//   Spec: "When deltaX > 0, text shifts left (first visible column is deltaX)".
//   Example: SetText("Hello"), move cursor to col 3 so deltaX=3; cell(0,0)
//   should show 'l' (the character at col 3 of "Hello").
//
// TestDrawDeltaYShiftsFirstLine:
//   When deltaY > 0, "first visible line is lines[deltaY]".
//   Spec: "When deltaY > 0, first visible line is lines[deltaY]".
//   Example: SetText("a\nb\nc"), scroll down so deltaY=1; cell(0,0) should
//   show 'b' (lines[1]), not 'a' (lines[0]).
//
// TestDrawDeltaYRowsAboveAreNotVisible:
//   Spec: "For each row y in [0, height): lineIdx = deltaY + y" — rows with
//   lineIdx < deltaY are not rendered.
//
// TestDrawDeltaXColumnsLeftOfDeltaXAreNotVisible:
//   Spec: "Drawing respects viewport: text scrolled by deltaX/deltaY".
