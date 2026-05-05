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
	lines         [][]rune
	cursorRow     int
	cursorCol     int
	deltaX        int
	deltaY        int
	autoIndent    bool
	selStartRow   int
	selStartCol   int
	selEndRow     int
	selEndCol     int
	dragging      bool
	dragAnchorRow int
	dragAnchorCol int
	hScrollBar    *ScrollBar
	vScrollBar    *ScrollBar
}

func NewMemo(bounds Rect, opts ...MemoOption) *Memo {
	m := &Memo{
		lines:      [][]rune{{}},
		autoIndent: true,
	}
	m.SetBounds(bounds)
	m.SetState(SfVisible, true)
	m.SetOptions(OfSelectable, true)
	m.SetOptions(OfFirstClick, true)
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
	m.syncScrollBars()
}

func (m *Memo) CursorPos() (int, int)      { return m.cursorRow, m.cursorCol }
func (m *Memo) AutoIndent() bool           { return m.autoIndent }
func (m *Memo) SetAutoIndent(enabled bool) { m.autoIndent = enabled }

func (m *Memo) SetVScrollBar(sb *ScrollBar) {
	if m.vScrollBar != nil {
		m.vScrollBar.OnChange = nil
	}
	m.vScrollBar = sb
	if sb != nil {
		sb.OnChange = func(val int) {
			m.deltaY = val
		}
		m.syncScrollBars()
	}
}

func (m *Memo) SetHScrollBar(sb *ScrollBar) {
	if m.hScrollBar != nil {
		m.hScrollBar.OnChange = nil
	}
	m.hScrollBar = sb
	if sb != nil {
		sb.OnChange = func(val int) {
			m.deltaX = val
		}
		m.syncScrollBars()
	}
}

func WithScrollBars(h, v *ScrollBar) MemoOption {
	return func(m *Memo) {
		m.SetHScrollBar(h)
		m.SetVScrollBar(v)
	}
}

func (m *Memo) syncScrollBars() {
	if m.vScrollBar != nil {
		m.vScrollBar.SetRange(0, len(m.lines)-1)
		m.vScrollBar.SetPageSize(m.Bounds().Height() - 1)
		m.vScrollBar.SetValue(m.deltaY)
	}
	if m.hScrollBar != nil {
		maxWidth := 0
		for _, line := range m.lines {
			if len(line) > maxWidth {
				maxWidth = len(line)
			}
		}
		m.hScrollBar.SetRange(0, maxWidth)
		m.hScrollBar.SetPageSize(m.Bounds().Width())
		m.hScrollBar.SetValue(m.deltaX)
	}
}

func (m *Memo) SetState(flag ViewState, on bool) {
	m.BaseView.SetState(flag, on)
	if flag&SfSelected != 0 {
		if m.vScrollBar != nil {
			m.vScrollBar.SetState(SfVisible, on)
		}
		if m.hScrollBar != nil {
			m.hScrollBar.SetState(SfVisible, on)
		}
	}
}

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

	cursorScreenX, cursorScreenY := -1, -1

	for y := 0; y < h; y++ {
		lineIdx := m.deltaY + y
		if lineIdx >= len(m.lines) {
			if lineIdx == m.cursorRow {
				cursorScreenX, cursorScreenY = 0, y
			}
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
			if lineIdx == m.cursorRow && runeIdx == m.cursorCol {
				cursorScreenX, cursorScreenY = x, y
			}

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

		if lineIdx == m.cursorRow && m.cursorCol >= len(line) && x < w {
			cursorScreenX, cursorScreenY = x, y
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

	if m.HasState(SfSelected) && !m.HasSelection() && cursorScreenX >= 0 {
		ch := rune(' ')
		if m.cursorRow < len(m.lines) && m.cursorCol < len(m.lines[m.cursorRow]) {
			ch = m.lines[m.cursorRow][m.cursorCol]
		}
		buf.WriteChar(cursorScreenX, cursorScreenY, ch, selectedStyle)
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

func (m *Memo) deleteSelection() {
	if !m.HasSelection() {
		return
	}
	sr, sc, er, ec := m.normalizedSelection()

	startLine := m.lines[sr]
	endLine := m.lines[er]

	merged := make([]rune, sc+len(endLine)-ec)
	copy(merged, startLine[:sc])
	copy(merged[sc:], endLine[ec:])

	m.lines[sr] = merged
	if er > sr {
		m.lines = append(m.lines[:sr+1], m.lines[er+1:]...)
	}

	m.cursorRow = sr
	m.cursorCol = sc
	m.clearSelection()
	m.syncScrollBars()
}

func (m *Memo) selectedText() string {
	if !m.HasSelection() {
		return ""
	}
	sr, sc, er, ec := m.normalizedSelection()
	if sr == er {
		return string(m.lines[sr][sc:ec])
	}
	var sb strings.Builder
	sb.WriteString(string(m.lines[sr][sc:]))
	for i := sr + 1; i < er; i++ {
		sb.WriteByte('\n')
		sb.WriteString(string(m.lines[i]))
	}
	sb.WriteByte('\n')
	sb.WriteString(string(m.lines[er][:ec]))
	return sb.String()
}

func (m *Memo) insertText(s string) {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	parts := strings.Split(s, "\n")
	if len(parts) == 1 {
		line := m.lines[m.cursorRow]
		runes := []rune(parts[0])
		newLine := make([]rune, len(line)+len(runes))
		copy(newLine, line[:m.cursorCol])
		copy(newLine[m.cursorCol:], runes)
		copy(newLine[m.cursorCol+len(runes):], line[m.cursorCol:])
		m.lines[m.cursorRow] = newLine
		m.cursorCol += len(runes)
	} else {
		line := m.lines[m.cursorRow]
		before := line[:m.cursorCol]
		after := line[m.cursorCol:]

		firstLine := make([]rune, len(before)+len([]rune(parts[0])))
		copy(firstLine, before)
		copy(firstLine[len(before):], []rune(parts[0]))

		lastPart := []rune(parts[len(parts)-1])
		lastLine := make([]rune, len(lastPart)+len(after))
		copy(lastLine, lastPart)
		copy(lastLine[len(lastPart):], after)

		newLines := make([][]rune, 0, len(m.lines)+len(parts)-1)
		newLines = append(newLines, m.lines[:m.cursorRow]...)
		newLines = append(newLines, firstLine)
		for i := 1; i < len(parts)-1; i++ {
			newLines = append(newLines, []rune(parts[i]))
		}
		newLines = append(newLines, lastLine)
		newLines = append(newLines, m.lines[m.cursorRow+1:]...)
		m.lines = newLines

		m.cursorRow += len(parts) - 1
		m.cursorCol = len(lastPart)
	}
	m.clearSelection()
	m.ensureCursorVisible()
	m.syncScrollBars()
}

// mouseToPos converts screen coordinates (relative to the Memo's top-left corner)
// to a (row, col) buffer position, accounting for tab expansion and viewport offsets.
func (m *Memo) mouseToPos(mx, my int) (int, int) {
	row := m.deltaY + my
	if row < 0 {
		row = 0
	}
	if row >= len(m.lines) {
		row = len(m.lines) - 1
	}

	line := m.lines[row]

	// Compute the visual column at deltaX (the left edge of the viewport).
	vcolAtDeltaX := 0
	for i := 0; i < m.deltaX && i < len(line); i++ {
		if line[i] == '\t' {
			vcolAtDeltaX += 8 - (vcolAtDeltaX % 8)
		} else {
			vcolAtDeltaX++
		}
	}

	targetVcol := vcolAtDeltaX + mx

	// Walk from start to find the rune at targetVcol.
	vcol := 0
	runeIdx := 0
	for runeIdx < len(line) && vcol < targetVcol {
		if line[runeIdx] == '\t' {
			tw := 8 - (vcol % 8)
			if vcol+tw > targetVcol {
				break // inside tab span — land on the tab rune
			}
			vcol += tw
		} else {
			vcol++
		}
		runeIdx++
	}

	if runeIdx > len(line) {
		runeIdx = len(line)
	}
	return row, runeIdx
}

// selectWordAtCursor selects the word (or whitespace/punctuation run) under the cursor.
func (m *Memo) selectWordAtCursor() {
	line := m.lines[m.cursorRow]
	if len(line) == 0 || m.cursorCol >= len(line) {
		return
	}
	cls := charClass(line[m.cursorCol])
	start := m.cursorCol
	for start > 0 && charClass(line[start-1]) == cls {
		start--
	}
	end := m.cursorCol
	for end < len(line) && charClass(line[end]) == cls {
		end++
	}
	m.selStartRow = m.cursorRow
	m.selStartCol = start
	m.cursorCol = end
	m.setSelectionEnd()
	m.ensureCursorVisible()
}

// selectLineAtCursor selects the entire current line.
func (m *Memo) selectLineAtCursor() {
	m.selStartRow = m.cursorRow
	m.selStartCol = 0
	m.cursorCol = len(m.lines[m.cursorRow])
	m.setSelectionEnd()
	m.ensureCursorVisible()
}

func (m *Memo) HandleEvent(event *Event) {
	// Handle mouse events before keyboard events.
	if event.What == EvMouse && event.Mouse != nil {
		m.BaseView.HandleEvent(event)
		if event.IsCleared() {
			return
		}
		me := event.Mouse
		if me.Button == tcell.WheelUp {
			m.deltaY -= 3
			if m.deltaY < 0 {
				m.deltaY = 0
			}
			m.syncScrollBars()
			event.Clear()
			return
		}
		if me.Button == tcell.WheelDown {
			maxDY := len(m.lines) - m.Bounds().Height()
			if maxDY < 0 {
				maxDY = 0
			}
			m.deltaY += 3
			if m.deltaY > maxDY {
				m.deltaY = maxDY
			}
			m.syncScrollBars()
			event.Clear()
			return
		}
		if me.Button == tcell.WheelLeft {
			m.deltaX -= 3
			if m.deltaX < 0 {
				m.deltaX = 0
			}
			m.syncScrollBars()
			event.Clear()
			return
		}
		if me.Button == tcell.WheelRight {
			maxWidth := 0
			for _, line := range m.lines {
				if len(line) > maxWidth {
					maxWidth = len(line)
				}
			}
			maxDX := maxWidth - m.Bounds().Width()
			if maxDX < 0 {
				maxDX = 0
			}
			m.deltaX += 3
			if m.deltaX > maxDX {
				m.deltaX = maxDX
			}
			m.syncScrollBars()
			event.Clear()
			return
		}
		if me.Button&tcell.Button1 != 0 {
			if m.dragging {
				// Continued drag (motion with button held).
				row, col := m.mouseToPos(me.X, me.Y)
				m.cursorRow = row
				m.cursorCol = col
				m.selStartRow = m.dragAnchorRow
				m.selStartCol = m.dragAnchorCol
				m.setSelectionEnd()
				m.ensureCursorVisible()
				m.syncScrollBars()
				event.Clear()
			} else {
				// Initial press — dispatch on click count.
				switch me.ClickCount {
				case 1:
					row, col := m.mouseToPos(me.X, me.Y)
					m.cursorRow = row
					m.cursorCol = col
					m.clearSelection()
					m.dragging = true
					m.dragAnchorRow = row
					m.dragAnchorCol = col
					m.ensureCursorVisible()
					m.syncScrollBars()
					event.Clear()
				case 2:
					row, col := m.mouseToPos(me.X, me.Y)
					m.cursorRow = row
					m.cursorCol = col
					m.selectWordAtCursor()
					m.syncScrollBars()
					event.Clear()
				case 3:
					row, col := m.mouseToPos(me.X, me.Y)
					m.cursorRow = row
					m.cursorCol = col
					m.selectLineAtCursor()
					m.syncScrollBars()
					event.Clear()
				}
			}
		} else if me.Button == tcell.ButtonNone && m.dragging {
			// Button released — stop dragging.
			m.dragging = false
			m.syncScrollBars()
			event.Clear()
		}
		return // mouse events don't fall through to keyboard handling
	}

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
			if shift {
				m.startSelectionIfNone()
			}
			m.wordLeft()
			if shift {
				m.setSelectionEnd()
			} else {
				m.clearSelection()
			}
			m.ensureCursorVisible()
			m.syncScrollBars()
			event.Clear()
			return
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
		m.syncScrollBars()
		event.Clear()
	case tcell.KeyRight:
		if k.Modifiers&tcell.ModCtrl != 0 {
			if shift {
				m.startSelectionIfNone()
			}
			m.wordRight()
			if shift {
				m.setSelectionEnd()
			} else {
				m.clearSelection()
			}
			m.ensureCursorVisible()
			m.syncScrollBars()
			event.Clear()
			return
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
		m.syncScrollBars()
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
		m.syncScrollBars()
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
		m.syncScrollBars()
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
		m.syncScrollBars()
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
		m.syncScrollBars()
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
		m.syncScrollBars()
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
		m.syncScrollBars()
		event.Clear()
	case tcell.KeyCtrlA:
		m.selStartRow = 0
		m.selStartCol = 0
		m.cursorRow = len(m.lines) - 1
		m.cursorCol = len(m.lines[m.cursorRow])
		m.setSelectionEnd()
		m.ensureCursorVisible()
		m.syncScrollBars()
		event.Clear()
	case tcell.KeyCtrlC:
		if m.HasSelection() {
			clipboard = m.selectedText()
		}
		m.syncScrollBars()
		event.Clear()
	case tcell.KeyCtrlX:
		if m.HasSelection() {
			clipboard = m.selectedText()
			m.deleteSelection()
		}
		m.syncScrollBars()
		event.Clear()
	case tcell.KeyCtrlV:
		if clipboard != "" {
			m.deleteSelection()
			m.insertText(clipboard)
		}
		m.syncScrollBars()
		event.Clear()
	case tcell.KeyRune:
		m.insertChar(k.Rune)
		m.syncScrollBars()
		event.Clear()
	case tcell.KeyEnter:
		m.insertNewline()
		m.syncScrollBars()
		event.Clear()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if k.Modifiers&tcell.ModCtrl != 0 {
			m.deleteWordLeft()
			m.ensureCursorVisible()
			m.syncScrollBars()
			event.Clear()
			return
		}
		m.backspace()
		m.syncScrollBars()
		event.Clear()
	case tcell.KeyDelete:
		if k.Modifiers&tcell.ModCtrl != 0 {
			m.deleteWordRight()
			m.ensureCursorVisible()
			m.syncScrollBars()
			event.Clear()
			return
		}
		m.deleteChar()
		m.syncScrollBars()
		event.Clear()
	case tcell.KeyCtrlY:
		m.deleteLine()
		m.syncScrollBars()
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
	m.syncScrollBars()
}

func (m *Memo) cursorRight() {
	if m.cursorCol < len(m.lines[m.cursorRow]) {
		m.cursorCol++
	} else if m.cursorRow < len(m.lines)-1 {
		m.cursorRow++
		m.cursorCol = 0
	}
	m.ensureCursorVisible()
	m.syncScrollBars()
}

func (m *Memo) cursorUp() {
	if m.cursorRow > 0 {
		m.cursorRow--
		m.clampCursor()
	}
	m.ensureCursorVisible()
	m.syncScrollBars()
}

func (m *Memo) cursorDown() {
	if m.cursorRow < len(m.lines)-1 {
		m.cursorRow++
		m.clampCursor()
	}
	m.ensureCursorVisible()
	m.syncScrollBars()
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
	m.syncScrollBars()
}

func (m *Memo) pageUp() {
	h := m.Bounds().Height()
	m.deltaY -= h - 1
	if m.deltaY < 0 {
		m.deltaY = 0
	}
	m.cursorRow -= h - 1
	if m.cursorRow < 0 {
		m.cursorRow = 0
	}
	m.clampCursor()
	m.ensureCursorVisible()
	m.syncScrollBars()
}

func (m *Memo) pageDown() {
	h := m.Bounds().Height()
	m.deltaY += h - 1
	maxDeltaY := len(m.lines) - h
	if maxDeltaY < 0 {
		maxDeltaY = 0
	}
	if m.deltaY > maxDeltaY {
		m.deltaY = maxDeltaY
	}
	m.cursorRow += h - 1
	if m.cursorRow >= len(m.lines) {
		m.cursorRow = len(m.lines) - 1
	}
	m.clampCursor()
	m.ensureCursorVisible()
	m.syncScrollBars()
}

func (m *Memo) insertChar(ch rune) {
	if m.HasSelection() {
		m.deleteSelection()
	}
	line := m.lines[m.cursorRow]
	newLine := make([]rune, len(line)+1)
	copy(newLine, line[:m.cursorCol])
	newLine[m.cursorCol] = ch
	copy(newLine[m.cursorCol+1:], line[m.cursorCol:])
	m.lines[m.cursorRow] = newLine
	m.cursorCol++
	m.ensureCursorVisible()
	m.clearSelection()
	m.syncScrollBars()
}

func (m *Memo) insertNewline() {
	// If a selection exists, capture the indent from the selection start line
	// before deleting the selection.
	var indentSourceLine []rune
	if m.HasSelection() {
		sr, _, _, _ := m.normalizedSelection()
		indentSourceLine = m.lines[sr]
		m.deleteSelection()
	} else {
		indentSourceLine = m.lines[m.cursorRow]
	}

	line := m.lines[m.cursorRow]
	before := make([]rune, m.cursorCol)
	copy(before, line[:m.cursorCol])
	after := make([]rune, len(line)-m.cursorCol)
	copy(after, line[m.cursorCol:])

	var indent []rune
	if m.autoIndent {
		for _, r := range indentSourceLine {
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
	m.clearSelection()
	m.syncScrollBars()
}

func (m *Memo) backspace() {
	if m.HasSelection() {
		m.deleteSelection()
		return
	}
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
	m.syncScrollBars()
}

func (m *Memo) deleteChar() {
	if m.HasSelection() {
		m.deleteSelection()
		return
	}
	line := m.lines[m.cursorRow]
	if m.cursorCol < len(line) {
		m.lines[m.cursorRow] = append(line[:m.cursorCol], line[m.cursorCol+1:]...)
	} else if m.cursorRow < len(m.lines)-1 {
		nextLine := m.lines[m.cursorRow+1]
		m.lines[m.cursorRow] = append(line, nextLine...)
		m.lines = append(m.lines[:m.cursorRow+1], m.lines[m.cursorRow+2:]...)
	}
	m.ensureCursorVisible()
	m.syncScrollBars()
}

// charClass classifies a rune for word-movement purposes.
// 0 = whitespace (space, tab)
// 1 = punctuation
// 2 = word character (everything else: letters, digits, underscore, etc.)
func charClass(r rune) int {
	if r == ' ' || r == '\t' {
		return 0
	}
	if strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^`{|}~", r) {
		return 1
	}
	return 2
}

func (m *Memo) wordLeft() {
	if m.cursorCol == 0 {
		if m.cursorRow > 0 {
			m.cursorRow--
			m.cursorCol = len(m.lines[m.cursorRow])
		}
		return
	}
	line := m.lines[m.cursorRow]
	col := m.cursorCol
	// skip whitespace
	for col > 0 && charClass(line[col-1]) == 0 {
		col--
	}
	if col == 0 {
		m.cursorCol = 0
		m.syncScrollBars()
		return
	}
	// skip current class
	cls := charClass(line[col-1])
	for col > 0 && charClass(line[col-1]) == cls {
		col--
	}
	m.cursorCol = col
	m.syncScrollBars()
}

func (m *Memo) wordRight() {
	line := m.lines[m.cursorRow]
	if m.cursorCol >= len(line) {
		if m.cursorRow < len(m.lines)-1 {
			m.cursorRow++
			m.cursorCol = 0
		}
		return
	}
	col := m.cursorCol
	// skip current class
	cls := charClass(line[col])
	for col < len(line) && charClass(line[col]) == cls {
		col++
	}
	// skip whitespace
	for col < len(line) && charClass(line[col]) == 0 {
		col++
	}
	m.cursorCol = col
	m.syncScrollBars()
}

func (m *Memo) deleteWordLeft() {
	if m.HasSelection() {
		m.deleteSelection()
		return
	}
	endRow, endCol := m.cursorRow, m.cursorCol
	m.wordLeft()
	if m.cursorRow == endRow && m.cursorCol == endCol {
		return // no movement, nothing to delete
	}
	// delete from (cursorRow, cursorCol) to (endRow, endCol)
	m.selStartRow, m.selStartCol = m.cursorRow, m.cursorCol
	m.selEndRow, m.selEndCol = endRow, endCol
	m.deleteSelection()
	m.syncScrollBars()
}

func (m *Memo) deleteWordRight() {
	if m.HasSelection() {
		m.deleteSelection()
		return
	}
	startRow, startCol := m.cursorRow, m.cursorCol
	m.wordRight()
	if m.cursorRow == startRow && m.cursorCol == startCol {
		return
	}
	m.selStartRow, m.selStartCol = startRow, startCol
	m.selEndRow, m.selEndCol = m.cursorRow, m.cursorCol
	m.cursorRow, m.cursorCol = startRow, startCol
	m.deleteSelection()
	m.syncScrollBars()
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
	m.clearSelection()
	m.syncScrollBars()
}
