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
	lines      [][]rune
	cursorRow  int
	cursorCol  int
	deltaX     int
	deltaY     int
	autoIndent bool
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
}

func (m *Memo) CursorPos() (int, int)     { return m.cursorRow, m.cursorCol }
func (m *Memo) AutoIndent() bool          { return m.autoIndent }
func (m *Memo) SetAutoIndent(enabled bool) { m.autoIndent = enabled }

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
	if cs != nil {
		normalStyle = cs.MemoNormal
	}

	h := m.Bounds().Height()
	w := m.Bounds().Width()

	for y := 0; y < h; y++ {
		lineIdx := m.deltaY + y
		if lineIdx >= len(m.lines) {
			buf.Fill(NewRect(0, y, w, 1), ' ', normalStyle)
			continue
		}
		line := m.lines[lineIdx]
		x := 0
		for col := m.deltaX; col < len(line) && x < w; col++ {
			buf.WriteChar(x, y, line[col], normalStyle)
			x++
		}
		for ; x < w; x++ {
			buf.WriteChar(x, y, ' ', normalStyle)
		}
	}
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

	switch k.Key {
	case tcell.KeyLeft:
		if k.Modifiers&tcell.ModCtrl != 0 {
			return // Ctrl+Left deferred to Phase 3
		}
		m.cursorLeft()
		event.Clear()
	case tcell.KeyRight:
		if k.Modifiers&tcell.ModCtrl != 0 {
			return // Ctrl+Right deferred to Phase 3
		}
		m.cursorRight()
		event.Clear()
	case tcell.KeyUp:
		m.cursorUp()
		event.Clear()
	case tcell.KeyDown:
		m.cursorDown()
		event.Clear()
	case tcell.KeyHome:
		if k.Modifiers&tcell.ModCtrl != 0 {
			m.cursorRow = 0
			m.cursorCol = 0
		} else {
			m.smartHome()
		}
		m.ensureCursorVisible()
		event.Clear()
	case tcell.KeyEnd:
		if k.Modifiers&tcell.ModCtrl != 0 {
			m.cursorRow = len(m.lines) - 1
			m.cursorCol = len(m.lines[m.cursorRow])
		} else {
			m.cursorCol = len(m.lines[m.cursorRow])
		}
		m.ensureCursorVisible()
		event.Clear()
	case tcell.KeyPgUp:
		m.pageUp()
		event.Clear()
	case tcell.KeyPgDn:
		m.pageDown()
		event.Clear()
	case tcell.KeyRune:
		// Printable character — handled in Task 5
	case tcell.KeyEnter:
		// Enter — handled in Task 5
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// Backspace — handled in Task 5
	case tcell.KeyDelete:
		// Delete — handled in Task 5
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
