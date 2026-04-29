package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

func TestScrollIndicatorRightOnlyWhenTextOverflowsRight(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop") // 16 chars, width 10
	il.cursorPos = 0
	il.scrollOffset = 0

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	if buf.GetCell(0, 0).Rune == '◄' {
		t.Error("should not show ◄ at col 0 when scrollOffset=0")
	}
	if buf.GetCell(9, 0).Rune != '►' {
		t.Errorf("should show ► at col 9 when text overflows right, got %q", string(buf.GetCell(9, 0).Rune))
	}
}

func TestScrollIndicatorBothWhenScrolledMiddle(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 8
	il.scrollOffset = 5

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	if buf.GetCell(0, 0).Rune != '◄' {
		t.Errorf("should show ◄ at col 0 when scrollOffset=5, got %q", string(buf.GetCell(0, 0).Rune))
	}
	if buf.GetCell(9, 0).Rune != '►' {
		t.Errorf("should show ► at col 9 when text overflows right, got %q", string(buf.GetCell(9, 0).Rune))
	}
}

func TestScrollIndicatorLeftOnlyWhenScrolledToEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 15
	il.scrollOffset = 6 // chars 6-15 visible, scrollOffset+w=16=len(text)

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	if buf.GetCell(0, 0).Rune != '◄' {
		t.Errorf("should show ◄ at col 0 when scrollOffset=6, got %q", string(buf.GetCell(0, 0).Rune))
	}
	if buf.GetCell(9, 0).Rune == '►' {
		t.Error("should not show ► when all remaining text fits")
	}
}

func TestScrollIndicatorNoneWhenTextFits(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abc")
	il.cursorPos = 0

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	if buf.GetCell(0, 0).Rune == '◄' {
		t.Error("should not show ◄ when text fits entirely")
	}
	if buf.GetCell(9, 0).Rune == '►' {
		t.Error("should not show ► when text fits entirely")
	}
}

func TestScrollIndicatorUsesSelectionStyle(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	il.SetText("abcdefghijklmnop")
	il.cursorPos = 8
	il.scrollOffset = 3

	selStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlue)
	cs := &theme.ColorScheme{
		InputNormal:    tcell.StyleDefault,
		InputSelection: selStyle,
	}
	il.scheme = cs

	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	if buf.GetCell(0, 0).Style != selStyle {
		t.Error("◄ indicator should use InputSelection style")
	}
	if buf.GetCell(9, 0).Style != selStyle {
		t.Error("► indicator should use InputSelection style")
	}
}
