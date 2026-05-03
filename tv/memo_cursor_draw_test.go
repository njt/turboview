package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

func TestMemoCursorVisibleWhenFocused(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 5))
	m.SetText("Hello")
	m.SetState(SfSelected, true)
	m.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	cell := buf.GetCell(0, 0)
	normalFg, normalBg, _ := theme.BorlandBlue.MemoNormal.Decompose()
	cellFg, cellBg, _ := cell.Style.Decompose()
	if cellFg == normalFg && cellBg == normalBg {
		t.Error("cursor at (0,0) should use MemoSelected style, not MemoNormal")
	}
}

func TestMemoCursorNotVisibleWhenUnfocused(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 5))
	m.SetText("Hello")
	m.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	cell := buf.GetCell(0, 0)
	normalFg, normalBg, _ := theme.BorlandBlue.MemoNormal.Decompose()
	cellFg, cellBg, _ := cell.Style.Decompose()
	if cellFg != normalFg || cellBg != normalBg {
		t.Error("cursor should not be highlighted when memo is unfocused")
	}
}

func TestMemoCursorAtEndOfLine(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 5))
	m.SetText("Hi")
	m.SetState(SfSelected, true)
	m.scheme = theme.BorlandBlue
	m.cursorCol = 2

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	cell := buf.GetCell(2, 0)
	if cell.Rune != ' ' {
		t.Errorf("cursor at end of line should show space, got %q", cell.Rune)
	}
	normalFg, normalBg, _ := theme.BorlandBlue.MemoNormal.Decompose()
	cellFg, cellBg, _ := cell.Style.Decompose()
	if cellFg == normalFg && cellBg == normalBg {
		t.Error("cursor at end of line should use cursor style, not normal")
	}
}

func TestMemoCursorHiddenDuringSelection(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 5))
	m.SetText("Hello World")
	m.SetState(SfSelected, true)
	m.scheme = theme.BorlandBlue
	m.selStartRow, m.selStartCol = 0, 0
	m.selEndRow, m.selEndCol = 0, 5
	m.cursorRow, m.cursorCol = 0, 5

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	// 'W' at col 6 is past the selection and past the cursor — should be normal
	cell6 := buf.GetCell(6, 0)
	normalFg, normalBg, _ := theme.BorlandBlue.MemoNormal.Decompose()
	fg6, bg6, _ := cell6.Style.Decompose()
	if fg6 != normalFg || bg6 != normalBg {
		t.Error("character past selection end should use normal style, not cursor/selection")
	}
}

func TestMemoCursorShowsCharacterUnderIt(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 20, 5))
	m.SetText("ABCDE")
	m.SetState(SfSelected, true)
	m.scheme = &theme.ColorScheme{
		MemoNormal:   tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack),
		MemoSelected: tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite),
	}
	m.cursorCol = 2

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	cell := buf.GetCell(2, 0)
	if cell.Rune != 'C' {
		t.Errorf("cursor should show character under it ('C'), got %q", cell.Rune)
	}
}
