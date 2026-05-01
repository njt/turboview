package tv

import "strings"

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
	// Implemented in Task 3
}

func (m *Memo) HandleEvent(event *Event) {
	// Implemented in Tasks 4 and 5
}
