# TMemo Core Implementation Plan (Batch 2, Phase 2)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the core of TMemo — a multi-line text editing widget with buffer model, drawing, cursor movement, and basic text editing. Selection, clipboard, word operations, mouse, scrollbar integration, and tab rendering are deferred to Phase 3.

**Architecture:** TMemo is a plain Widget (not Container). Text stored as `[][]rune` — a slice of lines, each a slice of runes. Viewport scrolling via `deltaX`/`deltaY` with auto-scroll to keep cursor visible. Drawing renders visible lines with fill for empty space. Keyboard handling covers cursor movement and basic text editing.

**Tech Stack:** Go, tcell/v2, existing tv package (BaseView, DrawBuffer, Event, theme.ColorScheme)

---

## File Map

- **Modify:** `theme/scheme.go` — add MemoNormal, MemoSelected fields
- **Modify:** `theme/borland.go`, `theme/borland_cyan.go`, `theme/borland_gray.go`, `theme/c64.go`, `theme/matrix.go` — add MemoNormal/MemoSelected values
- **Create:** `tv/memo.go` — TMemo widget
- **Modify:** `e2e/testapp/basic/main.go` — add TMemo to demo app
- **Modify:** `e2e/e2e_test.go` — add TMemo e2e test

---

- [ ] ### Task 1: Color Scheme Additions

**Files:**
- Modify: `theme/scheme.go`
- Modify: `theme/borland.go`, `theme/borland_cyan.go`, `theme/borland_gray.go`, `theme/c64.go`, `theme/matrix.go`

**Requirements:**
- `ColorScheme` struct has two new fields: `MemoNormal tcell.Style` and `MemoSelected tcell.Style`
- All 5 theme files set these fields with appropriate colors:
  - BorlandBlue: MemoNormal = yellow on blue, MemoSelected = white on green (matching original TV dialog palette entries 26-27)
  - Other themes: follow each theme's color conventions

**Implementation:**

Add to `theme/scheme.go` ColorScheme struct:
```go
MemoNormal   tcell.Style
MemoSelected tcell.Style
```

Add to each theme's ColorScheme initialization. For BorlandBlue:
```go
MemoNormal:   tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlue),
MemoSelected: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorGreen),
```

**Run tests:** `go test ./theme/... -v`

**Commit:** `git commit -m "feat(theme): add MemoNormal and MemoSelected color scheme entries"`

---

- [ ] ### Task 2: Memo Struct, Constructor, and Buffer Model

**Files:**
- Create: `tv/memo.go`

**Requirements:**
- `Memo` struct with fields: `lines [][]rune`, `cursorRow int`, `cursorCol int`, `deltaX int`, `deltaY int`, `autoIndent bool`
- `var _ Widget = (*Memo)(nil)` compile-time assertion
- `NewMemo(bounds Rect, opts ...MemoOption) *Memo`:
  - Initializes with one empty line (`[][]rune{{}}`)
  - Sets SfVisible, OfSelectable
  - GrowMode: `GfGrowHiX | GfGrowHiY`
  - Auto-indent: true by default
  - Cursor at (0, 0)
- `MemoOption` functional options:
  - `WithAutoIndent(enabled bool) MemoOption`
- `Text() string` — joins lines with `\n`, returns full text
- `SetText(s string)` — splits on `\n` and `\r\n`, stores as `[][]rune`, resets cursor to (0, 0), resets delta to (0, 0)
- `CursorPos() (row, col int)` — returns current cursor position
- `AutoIndent() bool` — returns auto-indent state
- `SetAutoIndent(enabled bool)` — sets auto-indent
- Cursor clamping: `cursorRow` always in `[0, len(lines)-1]`, `cursorCol` in `[0, len(lines[cursorRow])]`
- An empty Memo (after `NewMemo`) has `Text() == ""` and `CursorPos() == (0, 0)`
- `SetText("hello\nworld")` results in 2 lines, `Text()` returns `"hello\nworld"`
- `SetText("")` results in 1 empty line (not 0 lines)

**Implementation:**

```go
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

func (m *Memo) CursorPos() (int, int)         { return m.cursorRow, m.cursorCol }
func (m *Memo) AutoIndent() bool               { return m.autoIndent }
func (m *Memo) SetAutoIndent(enabled bool)     { m.autoIndent = enabled }

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
```

Stub Draw and HandleEvent (will be filled in subsequent tasks):
```go
func (m *Memo) Draw(buf *DrawBuffer) {
	// Implemented in Task 3
}

func (m *Memo) HandleEvent(event *Event) {
	// Implemented in Tasks 4 and 5
}
```

**Run tests:** `go test ./tv/... -run TestMemo -v`

**Commit:** `git commit -m "feat: add Memo widget with buffer model and cursor state"`

---

- [ ] ### Task 3: Memo Drawing

**Files:**
- Modify: `tv/memo.go` — implement Draw method

**Requirements:**
- `Draw(buf)` renders visible lines from the viewport:
  - For each row y in [0, height): line index = deltaY + y
  - If lineIdx >= len(lines): fill row with spaces in MemoNormal
  - Otherwise: render line content starting from column deltaX in MemoNormal
  - After line content ends, fill remaining columns with spaces in MemoNormal
- No selection rendering in this task (deferred to Phase 3)
- No tab rendering in this task (deferred to Phase 3)
- Drawing respects viewport: text scrolled by deltaX/deltaY
- When deltaX > 0, text shifts left (first visible column is deltaX)
- When deltaY > 0, first visible line is lines[deltaY]
- The cursor position is NOT visually rendered by Draw (the Application sets the terminal cursor separately)

**Implementation:**

```go
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
```

**Run tests:** `go test ./tv/... -run TestMemo -v`

**Commit:** `git commit -m "feat(memo): implement Draw with viewport scrolling"`

---

- [ ] ### Task 4: Memo Cursor Movement

**Files:**
- Modify: `tv/memo.go` — implement cursor movement in HandleEvent

**Requirements:**
- Left: move cursor one rune left; at start of line wraps to end of previous line; at (0,0) does nothing
- Right: move cursor one rune right; at end of line wraps to start of next line; at end of document does nothing
- Up: move one line up, column clamped to target line length; at row 0 does nothing
- Down: move one line down, column clamped to target line length; at last row does nothing
- Home: smart Home — move to first non-whitespace char; if already there, move to column 0
- End: move to end of current line
- Ctrl+Home: move to (0, 0)
- Ctrl+End: move to (lastRow, len(lastLine))
- PgUp: move up by (height - 1) lines, clamp row to 0
- PgDn: move down by (height - 1) lines, clamp row to len(lines)-1
- After every cursor movement, call `ensureCursorVisible()` to auto-scroll viewport
- Tab key is NOT consumed (passes through for focus navigation)
- Alt+anything is NOT consumed (passes through for shortcuts)
- F-keys are NOT consumed (window management)
- All movement clears any selection (but selection doesn't exist yet — this is a no-op in Phase 2)
- Vertical movement column preservation: column attempts to stay at current position, clamped to target line end

**Implementation:**

```go
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
			// Ctrl+Left — deferred to Phase 3
			return
		}
		m.cursorLeft()
		event.Clear()
	case tcell.KeyRight:
		if k.Modifiers&tcell.ModCtrl != 0 {
			// Ctrl+Right — deferred to Phase 3
			return
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
```

**Run tests:** `go test ./tv/... -run TestMemo -v`

**Commit:** `git commit -m "feat(memo): implement cursor movement (arrows, Home, End, PgUp, PgDn)"`

---

- [ ] ### Task 5: Memo Text Editing

**Files:**
- Modify: `tv/memo.go` — implement text editing in HandleEvent

**Requirements:**
- Printable character (KeyRune): insert rune at cursor position, advance cursor by 1
- Enter: split current line at cursor position, creating a new line. If auto-indent is enabled, copy leading whitespace (spaces and tabs) from the current line to the new line.
- Backspace: at start of line → join with previous line (append current line content to previous line, remove current line, cursor at join point). Within line → delete rune before cursor.
- Delete: at end of line → join with next line (append next line content to current line, remove next line). Within line → delete rune after cursor.
- Ctrl+Y: delete entire current line. If it's the only line, clear it to empty. Cursor row stays the same (or decrements if at last line).
- All editing operations call `clampCursor()` and `ensureCursorVisible()` after modification.

**Implementation:**

Add to HandleEvent switch cases:
```go
case tcell.KeyRune:
	m.insertChar(k.Rune)
	event.Clear()
case tcell.KeyEnter:
	m.insertNewline()
	event.Clear()
case tcell.KeyBackspace, tcell.KeyBackspace2:
	if k.Modifiers&tcell.ModCtrl != 0 {
		// Ctrl+Backspace — deferred to Phase 3
		return
	}
	m.backspace()
	event.Clear()
case tcell.KeyDelete:
	if k.Modifiers&tcell.ModCtrl != 0 {
		// Ctrl+Delete — deferred to Phase 3
		return
	}
	m.deleteChar()
	event.Clear()
```

Add Ctrl+Y handling:
```go
// Inside HandleEvent, add before the default case:
case tcell.KeyRune:
	if k.Modifiers&tcell.ModCtrl != 0 && k.Rune == 'y' {
		// Actually Ctrl+Y is sent as a specific key, handle separately
	}
```

Note: Ctrl+Y in tcell may come as `tcell.KeyCtrlY`. Check how existing code handles it (InputLine uses `case tcell.KeyCtrlY`).

Editing helper methods:
```go
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
```

**Run tests:** `go test ./tv/... -run TestMemo -v`

**Commit:** `git commit -m "feat(memo): implement text editing (insert, Enter, Backspace, Delete, Ctrl+Y)"`

---

- [ ] ### Task 6: Integration Checkpoint — TMemo Core

**Purpose:** Verify Tasks 2-5 work together correctly.

**Requirements (for test writer):**
- A Memo in a Window receives keyboard events through Group dispatch and responds (type text, see cursor move, see text change)
- SetText followed by cursor movement produces correct positions
- Drawing after text editing shows the edited content
- Viewport scrolling: Set text with 20 lines in a 5-row Memo, cursor at row 15 → deltaY adjusts so line 15 is visible in Draw output
- Auto-indent: Enter after a line starting with "    " creates new line with 4 leading spaces
- Backspace at line start joins with previous line — verify via Text() and CursorPos()
- Delete at line end joins with next line — verify via Text()
- Ctrl+Y deletes entire line — verify via Text() and line count

**Components to wire up:** Memo, Window or Group (all real, no mocks)

**Run:** `go test ./tv/... -run TestIntegration -v`

---

- [ ] ### Task 7: E2E Test — TMemo in Demo App

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**
- Add a TMemo to the demo app (either in a new window "Notes" or replacing content in an existing window)
- The Memo should have sample text pre-loaded via SetText
- The e2e test verifies: sample text is visible, type additional text, the typed text appears on screen
- The e2e test verifies: press Down arrow multiple times, the cursor moves (verify by checking that the view content changes or scrolls)

**Implementation:**

Add a third window to the demo app:
```go
win3 := tv.NewWindow(tv.NewRect(45, 3, 30, 10), "Notes", tv.WithWindowNumber(3))
memo := tv.NewMemo(tv.NewRect(0, 0, 28, 8))
memo.SetText("Hello, World!\nThis is a memo.\nLine 3\nLine 4")
win3.Insert(memo)
app.Desktop().Insert(win3)
```

The e2e test infrastructure uses tmux. Follow existing patterns.

**Run tests:** `go test ./e2e/... -run TestMemo -v -timeout 60s`

**Commit:** `git commit -m "feat: add TMemo to demo app with e2e test"`
