package tv

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

var _ Widget = (*Memo)(nil)

type MemoOption func(*Memo)

func WithAutoIndent(enabled bool) MemoOption {
	return func(m *Memo) { m.autoIndent = enabled }
}

type Memo struct {
	BaseView
	lines       [][]rune
	cursorRow   int
	cursorCol   int
	deltaX      int
	deltaY      int
	autoIndent  bool
	selStartRow int
	selStartCol int
	selEndRow   int
	selEndCol   int
}

func NewMemo(bounds Rect, opts ...MemoOption) *Memo {
	m := &Memo{
		lines:      [][]rune{{}},
		autoIndent: true,
	}
	m.SetBounds(bounds)
	m.SetState(SfVisible, true)
	m.SetOptions(OfSelectable, true)
	m.SetGrowMode(GfGrowHiX | GfGrowHiY)

	for _, opt := range opts {
		opt(m)
	}

	m.SetSelf(m)
	return m
}

func (m *Memo) Text() string {
	strs := make([]string, len(m.lines))
	for i, line := range m.lines {
		strs[i] = string(line)
	}
	return strings.Join(strs, "\n")
}

func (m *Memo) SetText(s string) {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	parts := strings.Split(s, "\n")
	m.lines = make([][]rune, len(parts))
	for i, p := range parts {
		m.lines[i] = []rune(p)
	}
	if len(m.lines) == 0 {
		m.lines = [][]rune{{}}
	}
	m.cursorRow = 0
	m.cursorCol = 0
	m.deltaX = 0
	m.deltaY = 0
	m.selStartRow = 0
	m.selStartCol = 0
	m.selEndRow = 0
	m.selEndCol = 0
}

func (m *Memo) CursorPos() (int, int)      { return m.cursorRow, m.cursorCol }
func (m *Memo) AutoIndent() bool           { return m.autoIndent }
func (m *Memo) SetAutoIndent(enabled bool) { m.autoIndent = enabled }

func (m *Memo) Selection() (int, int, int, int) {
	return m.selStartRow, m.selStartCol, m.selEndRow, m.selEndCol
}

func (m *Memo) HasSelection() bool {
	return m.selStartRow != m.selEndRow || m.selStartCol != m.selEndCol
}

func (m *Memo) normalizedSelection() (int, int, int, int) {
	sr, sc, er, ec := m.selStartRow, m.selStartCol, m.selEndRow, m.selEndCol
	if sr > er || (sr == er && sc > ec) {
		sr, sc, er, ec = er, ec, sr, sc
	}
	return sr, sc, er, ec
}

func (m *Memo) clearSelection() {
	m.selStartRow = m.cursorRow
	m.selStartCol = m.cursorCol
	m.selEndRow = m.cursorRow
	m.selEndCol = m.cursorCol
}

func (m *Memo) setSelectionEnd() {
	m.selEndRow = m.cursorRow
	m.selEndCol = m.cursorCol
}

func (m *Memo) startSelectionIfNone() {
	if !m.HasSelection() {
		m.selStartRow = m.cursorRow
		m.selStartCol = m.cursorCol
	}
}

func (m *Memo) clampCursor() {
	if m.cursorRow < 0 {
		m.cursorRow = 0
	}
	if m.cursorRow >= len(m.lines) {
		m.cursorRow = len(m.lines) - 1
	}
	if m.cursorCol < 0 {
		m.cursorCol = 0
	}
	if m.cursorCol > len(m.lines[m.cursorRow]) {
		m.cursorCol = len(m.lines[m.cursorRow])
	}
}

func (m *Memo) ensureCursorVisible() {
	h := m.Bounds().Height()
	w := m.Bounds().Width()
	if m.cursorRow < m.deltaY {
		m.deltaY = m.cursorRow
	}
	if m.cursorRow >= m.deltaY+h {
		m.deltaY = m.cursorRow - h + 1
	}
	if m.cursorCol < m.deltaX {
		m.deltaX = m.cursorCol
	}
	if m.cursorCol >= m.deltaX+w {
		m.deltaX = m.cursorCol - w + 1
	}
}

func (m *Memo) Draw(buf *DrawBuffer) {
	cs := m.ColorScheme()
	normalStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault
	if cs != nil {
		normalStyle = cs.MemoNormal
		selectedStyle = cs.MemoSelected
	}

	h := m.Bounds().Height()
	w := m.Bounds().Width()

	sr, sc, er, ec := m.normalizedSelection()

	for y := 0; y < h; y++ {
		lineIdx := m.deltaY + y
		if lineIdx >= len(m.lines) {
			buf.Fill(NewRect(0, y, w, 1), ' ', normalStyle)
			continue
		}
		line := m.lines[lineIdx]

		// Walk runes before deltaX to compute correct visual column for tab alignment.
		vcol := 0
		for runeIdx := 0; runeIdx < m.deltaX && runeIdx < len(line); runeIdx++ {
			if line[runeIdx] == '\t' {
				vcol += 8 - (vcol % 8)
			} else {
				vcol++
			}
		}

		// Render visible runes starting from deltaX.
		x := 0
		for runeIdx := m.deltaX; runeIdx < len(line) && x < w; runeIdx++ {
			ch := line[runeIdx]
			inSel := m.posInSelection(lineIdx, runeIdx, sr, sc, er, ec)

			style := normalStyle
			if inSel {
				style = selectedStyle
			}

			if ch == '\t' {
				tabWidth := 8 - (vcol % 8)
				for i := 0; i < tabWidth && x < w; i++ {
					buf.WriteChar(x, y, ' ', style)
					x++
				}
				vcol += tabWidth
			} else {
				buf.WriteChar(x, y, ch, style)
				x++
				vcol++
			}
		}

		// Trailing fill.
		trailSelected := m.trailingSelected(lineIdx, len(line), sr, sc, er, ec)
		trailStyle := normalStyle
		if trailSelected {
			trailStyle = selectedStyle
		}
		for ; x < w; x++ {
			buf.WriteChar(x, y, ' ', trailStyle)
		}
	}
}

// posInSelection reports whether the rune at (row, col) falls within the
// normalized selection [sr,sc)→(er,ec).
func (m *Memo) posInSelection(row, col, sr, sc, er, ec int) bool {
	if sr == er && sc == ec {
		return false // no selection
	}
	if row < sr || row > er {
		return false
	}
	if row == sr && row == er {
		return col >= sc && col < ec
	}
	if row == sr {
		return col >= sc
	}
	if row == er {
		return col < ec
	}
	return true // intermediate line
}

// trailingSelected reports whether the trailing fill spaces on (row) should
// use MemoSelected style. This is true for intermediate lines of a selection
// (the newline character is considered selected).
func (m *Memo) trailingSelected(row, lineLen, sr, sc, er, ec int) bool {
	if sr == er && sc == ec {
		return false
	}
	if row < sr || row >= er {
		return false
	}
	if row == sr {
		return lineLen >= sc // trailing of first sel line only if sel starts before/at end
	}
	return true // intermediate line: trailing is selected
}

func (m *Memo) HandleEvent(event *Event) {
	if event.What != EvKeyboard || event.Key == nil {
		return
	}
	k := event.Key

	// Don't consume Alt+anything or F-keys
	if k.Modifiers&tcell.ModAlt != 0 {
		return
	}

	shift := k.Modifiers&tcell.ModShift != 0

	switch k.Key {
	case tcell.KeyLeft:
		if k.Modifiers&tcell.ModCtrl != 0 {
			return // Ctrl+Left deferred to Task 4
		}
		if shift {
			m.startSelectionIfNone()
			m.cursorLeft()
			m.setSelectionEnd()
		} else {
			if m.HasSelection() {
				m.clearSelection()
			} else {
				m.cursorLeft()
				m.clearSelection()
			}
		}
		event.Clear()
	case tcell.KeyRight:
		if k.Modifiers&tcell.ModCtrl != 0 {
			return // Ctrl+Right deferred to Task 4
		}
		if shift {
			m.startSelectionIfNone()
			m.cursorRight()
			m.setSelectionEnd()
		} else {
			if m.HasSelection() {
				m.clearSelection()
			} else {
				m.cursorRight()
				m.clearSelection()
			}
		}
		event.Clear()
	case tcell.KeyUp:
		if shift {
			m.startSelectionIfNone()
			if m.cursorRow == 0 {
				// Shift+Up from row 0: move cursor to col 0
				m.cursorCol = 0
				m.ensureCursorVisible()
			} else {
				m.cursorUp()
			}
			m.setSelectionEnd()
		} else {
			if m.HasSelection() {
				m.clearSelection()
			} else {
				m.cursorUp()
				m.clearSelection()
			}
		}
		event.Clear()
	case tcell.KeyDown:
		if shift {
			m.startSelectionIfNone()
			if m.cursorRow == len(m.lines)-1 {
				// Shift+Down from last row: move cursor to end of line
				m.cursorCol = len(m.lines[m.cursorRow])
				m.ensureCursorVisible()
			} else {
				m.cursorDown()
			}
			m.setSelectionEnd()
		} else {
			if m.HasSelection() {
				m.clearSelection()
			} else {
				m.cursorDown()
				m.clearSelection()
			}
		}
		event.Clear()
	case tcell.KeyHome:
		if shift {
			m.startSelectionIfNone()
			if k.Modifiers&tcell.ModCtrl != 0 {
				m.cursorRow = 0
				m.cursorCol = 0
			} else {
				m.smartHome()
			}
			m.ensureCursorVisible()
			m.setSelectionEnd()
		} else {
			if k.Modifiers&tcell.ModCtrl != 0 {
				m.cursorRow = 0
				m.cursorCol = 0
			} else {
				m.smartHome()
			}
			m.ensureCursorVisible()
			m.clearSelection()
		}
		event.Clear()
	case tcell.KeyEnd:
		if shift {
			m.startSelectionIfNone()
			if k.Modifiers&tcell.ModCtrl != 0 {
				m.cursorRow = len(m.lines) - 1
				m.cursorCol = len(m.lines[m.cursorRow])
			} else {
				m.cursorCol = len(m.lines[m.cursorRow])
			}
			m.ensureCursorVisible()
			m.setSelectionEnd()
		} else {
			if k.Modifiers&tcell.ModCtrl != 0 {
				m.cursorRow = len(m.lines) - 1
				m.cursorCol = len(m.lines[m.cursorRow])
			} else {
				m.cursorCol = len(m.lines[m.cursorRow])
			}
			m.ensureCursorVisible()
			m.clearSelection()
		}
		event.Clear()
	case tcell.KeyPgUp:
		if shift {
			m.startSelectionIfNone()
			m.pageUp()
			m.setSelectionEnd()
		} else {
			if m.HasSelection() {
				m.clearSelection()
			} else {
				m.pageUp()
				m.clearSelection()
			}
		}
		event.Clear()
	case tcell.KeyPgDn:
		if shift {
			m.startSelectionIfNone()
			m.pageDown()
			m.setSelectionEnd()
		} else {
			if m.HasSelection() {
				m.clearSelection()
			} else {
				m.pageDown()
				m.clearSelection()
			}
		}
		event.Clear()
	case tcell.KeyCtrlA:
		m.selStartRow = 0
		m.selStartCol = 0
		m.cursorRow = len(m.lines) - 1
		m.cursorCol = len(m.lines[m.cursorRow])
		m.setSelectionEnd()
		m.ensureCursorVisible()
		event.Clear()
	case tcell.KeyRune:
		m.insertChar(k.Rune)
		event.Clear()
	case tcell.KeyEnter:
		m.insertNewline()
		event.Clear()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if k.Modifiers&tcell.ModCtrl != 0 {
			return // Ctrl+Backspace deferred to Task 4
		}
		m.backspace()
		event.Clear()
	case tcell.KeyDelete:
		if k.Modifiers&tcell.ModCtrl != 0 {
			return // Ctrl+Delete deferred to Task 4
		}
		m.deleteChar()
		event.Clear()
	case tcell.KeyCtrlY:
		m.deleteLine()
		event.Clear()
	default:
		// Tab, F-keys, etc. — not consumed
	}
}

func (m *Memo) cursorLeft() {
	if m.cursorCol > 0 {
		m.cursorCol--
	} else if m.cursorRow > 0 {
		m.cursorRow--
		m.cursorCol = len(m.lines[m.cursorRow])
	}
	m.ensureCursorVisible()
}

func (m *Memo) cursorRight() {
	if m.cursorCol < len(m.lines[m.cursorRow]) {
		m.cursorCol++
	} else if m.cursorRow < len(m.lines)-1 {
		m.cursorRow++
		m.cursorCol = 0
	}
	m.ensureCursorVisible()
}

func (m *Memo) cursorUp() {
	if m.cursorRow > 0 {
		m.cursorRow--
		m.clampCursor()
	}
	m.ensureCursorVisible()
}

func (m *Memo) cursorDown() {
	if m.cursorRow < len(m.lines)-1 {
		m.cursorRow++
		m.clampCursor()
	}
	m.ensureCursorVisible()
}

func (m *Memo) smartHome() {
	line := m.lines[m.cursorRow]
	firstNonWS := 0
	for firstNonWS < len(line) && (line[firstNonWS] == ' ' || line[firstNonWS] == '\t') {
		firstNonWS++
	}
	if m.cursorCol == firstNonWS {
		m.cursorCol = 0
	} else {
		m.cursorCol = firstNonWS
	}
}

func (m *Memo) pageUp() {
	h := m.Bounds().Height()
	m.cursorRow -= h - 1
	if m.cursorRow < 0 {
		m.cursorRow = 0
	}
	m.clampCursor()
	m.ensureCursorVisible()
}

func (m *Memo) pageDown() {
	h := m.Bounds().Height()
	m.cursorRow += h - 1
	if m.cursorRow >= len(m.lines) {
		m.cursorRow = len(m.lines) - 1
	}
	m.clampCursor()
	m.ensureCursorVisible()
}

func (m *Memo) insertChar(ch rune) {
	line := m.lines[m.cursorRow]
	newLine := make([]rune, len(line)+1)
	copy(newLine, line[:m.cursorCol])
	newLine[m.cursorCol] = ch
	copy(newLine[m.cursorCol+1:], line[m.cursorCol:])
	m.lines[m.cursorRow] = newLine
	m.cursorCol++
	m.ensureCursorVisible()
}

func (m *Memo) insertNewline() {
	line := m.lines[m.cursorRow]
	before := make([]rune, m.cursorCol)
	copy(before, line[:m.cursorCol])
	after := make([]rune, len(line)-m.cursorCol)
	copy(after, line[m.cursorCol:])

	var indent []rune
	if m.autoIndent {
		for _, r := range line {
			if r == ' ' || r == '\t' {
				indent = append(indent, r)
			} else {
				break
			}
		}
	}

	newAfter := make([]rune, len(indent)+len(after))
	copy(newAfter, indent)
	copy(newAfter[len(indent):], after)

	m.lines[m.cursorRow] = before
	newLines := make([][]rune, len(m.lines)+1)
	copy(newLines, m.lines[:m.cursorRow+1])
	newLines[m.cursorRow+1] = newAfter
	copy(newLines[m.cursorRow+2:], m.lines[m.cursorRow+1:])
	m.lines = newLines

	m.cursorRow++
	m.cursorCol = len(indent)
	m.ensureCursorVisible()
}

func (m *Memo) backspace() {
	if m.cursorCol > 0 {
		line := m.lines[m.cursorRow]
		m.lines[m.cursorRow] = append(line[:m.cursorCol-1], line[m.cursorCol:]...)
		m.cursorCol--
	} else if m.cursorRow > 0 {
		prevLine := m.lines[m.cursorRow-1]
		curLine := m.lines[m.cursorRow]
		joinCol := len(prevLine)
		m.lines[m.cursorRow-1] = append(prevLine, curLine...)
		m.lines = append(m.lines[:m.cursorRow], m.lines[m.cursorRow+1:]...)
		m.cursorRow--
		m.cursorCol = joinCol
	}
	m.ensureCursorVisible()
}

func (m *Memo) deleteChar() {
	line := m.lines[m.cursorRow]
	if m.cursorCol < len(line) {
		m.lines[m.cursorRow] = append(line[:m.cursorCol], line[m.cursorCol+1:]...)
	} else if m.cursorRow < len(m.lines)-1 {
		nextLine := m.lines[m.cursorRow+1]
		m.lines[m.cursorRow] = append(line, nextLine...)
		m.lines = append(m.lines[:m.cursorRow+1], m.lines[m.cursorRow+2:]...)
	}
	m.ensureCursorVisible()
}

func (m *Memo) deleteLine() {
	if len(m.lines) == 1 {
		m.lines[0] = []rune{}
		m.cursorCol = 0
	} else {
		m.lines = append(m.lines[:m.cursorRow], m.lines[m.cursorRow+1:]...)
		if m.cursorRow >= len(m.lines) {
			m.cursorRow = len(m.lines) - 1
		}
	}
	m.clampCursor()
	m.ensureCursorVisible()
}
