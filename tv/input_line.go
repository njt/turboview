package tv

import "github.com/gdamore/tcell/v2"

var _ Widget = (*InputLine)(nil)

var clipboard string

type InputLine struct {
	BaseView
	text         []rune
	maxLen       int
	cursorPos    int
	scrollOffset int
	selStart     int
	selEnd       int
	overwrite    bool
	dragging     bool
	dragAnchor   int
}

type InputLineOption func(*InputLine)

func NewInputLine(bounds Rect, maxLen int, opts ...InputLineOption) *InputLine {
	il := &InputLine{maxLen: maxLen}
	il.SetBounds(bounds)
	il.SetState(SfVisible, true)
	il.SetOptions(OfSelectable, true)
	for _, opt := range opts {
		opt(il)
	}
	il.SetSelf(il)
	return il
}

func (il *InputLine) Text() string          { return string(il.text) }
func (il *InputLine) CursorPos() int        { return il.cursorPos }
func (il *InputLine) Selection() (int, int) { return il.selStart, il.selEnd }
func (il *InputLine) Overwrite() bool       { return il.overwrite }

func (il *InputLine) SelectAll() {
	il.selStart = 0
	il.selEnd = len(il.text)
	il.cursorPos = len(il.text)
}

func (il *InputLine) SetText(s string) {
	runes := []rune(s)
	if il.maxLen > 0 && len(runes) > il.maxLen {
		runes = runes[:il.maxLen]
	}
	il.text = runes
	il.cursorPos = len(runes)
	il.selStart = 0
	il.selEnd = 0
	il.adjustScroll()
}

// normalizedSel returns the selection range in order (lo, hi).
func (il *InputLine) normalizedSel() (int, int) {
	if il.selStart <= il.selEnd {
		return il.selStart, il.selEnd
	}
	return il.selEnd, il.selStart
}

// hasSelection returns true when a non-empty selection exists.
func (il *InputLine) hasSelection() bool {
	return il.selStart != il.selEnd
}

// deleteSelection removes the selected text, positions cursor at start of
// former selection, and clears the selection.
func (il *InputLine) deleteSelection() {
	lo, hi := il.normalizedSel()
	il.text = append(il.text[:lo], il.text[hi:]...)
	il.cursorPos = lo
	il.selStart = 0
	il.selEnd = 0
}

// adjustScroll ensures cursorPos is visible within the widget width.
func (il *InputLine) adjustScroll() {
	w := il.Bounds().Width()
	if w <= 0 {
		return
	}
	// Scroll right if cursor is past the visible area.
	if il.cursorPos >= il.scrollOffset+w {
		il.scrollOffset = il.cursorPos - w + 1
	}
	// Scroll left if cursor is before the scroll offset.
	if il.cursorPos < il.scrollOffset {
		il.scrollOffset = il.cursorPos
	}
}

// wordLeft returns the start of the word to the left of pos.
func (il *InputLine) wordLeft(pos int) int {
	if pos <= 0 {
		return 0
	}
	i := pos - 1
	for i > 0 && il.text[i] == ' ' {
		i--
	}
	for i > 0 && il.text[i-1] != ' ' {
		i--
	}
	return i
}

// wordRight returns the start of the next word to the right of pos.
func (il *InputLine) wordRight(pos int) int {
	n := len(il.text)
	if pos >= n {
		return n
	}
	i := pos
	for i < n && il.text[i] != ' ' {
		i++
	}
	for i < n && il.text[i] == ' ' {
		i++
	}
	return i
}

func (il *InputLine) Draw(buf *DrawBuffer) {
	w := il.Bounds().Width()
	cs := il.ColorScheme()
	normalStyle := tcell.StyleDefault
	selectionStyle := tcell.StyleDefault
	if cs != nil {
		normalStyle = cs.InputNormal
		selectionStyle = cs.InputSelection
	}

	// Fill the entire row with normal style.
	buf.Fill(NewRect(0, 0, w, 1), ' ', normalStyle)

	// Determine selection range.
	selLo, selHi := il.normalizedSel()
	isSelected := il.HasState(SfSelected)

	// Render text characters from scrollOffset.
	for col := 0; col < w; col++ {
		textIdx := col + il.scrollOffset
		if textIdx >= len(il.text) {
			break
		}
		ch := il.text[textIdx]
		style := normalStyle

		// Apply selection style if within selection range.
		if il.hasSelection() && textIdx >= selLo && textIdx < selHi {
			style = selectionStyle
		}

		buf.WriteChar(col, 0, ch, style)
	}

	// Draw cursor indicator when selected and no active selection.
	if isSelected && !il.hasSelection() {
		cursorCol := il.cursorPos - il.scrollOffset
		if cursorCol >= 0 && cursorCol < w {
			// Get current char at cursor (space if past text end).
			var ch rune = ' '
			if il.cursorPos < len(il.text) {
				ch = il.text[il.cursorPos]
			}
			buf.WriteChar(cursorCol, 0, ch, selectionStyle)
		}
	}

	// Scroll overflow indicators
	if il.scrollOffset > 0 {
		buf.WriteChar(0, 0, '◄', selectionStyle)
	}
	if il.scrollOffset+w < len(il.text) {
		buf.WriteChar(w-1, 0, '►', selectionStyle)
	}
}

func (il *InputLine) HandleEvent(event *Event) {
	switch event.What {
	case EvMouse:
		if event.Mouse == nil {
			return
		}
		col := event.Mouse.X - il.Bounds().A.X + il.scrollOffset
		if col < 0 {
			col = 0
		}
		if col > len(il.text) {
			col = len(il.text)
		}

		if event.Mouse.Button&tcell.Button1 != 0 {
			if event.Mouse.ClickCount >= 2 {
				// Double-click: select all text.
				il.selStart = 0
				il.selEnd = len(il.text)
				il.cursorPos = len(il.text)
				il.dragging = false
				il.adjustScroll()
				event.Clear()
				return
			}
			if !il.dragging {
				il.dragging = true
				il.dragAnchor = col
				il.cursorPos = col
				il.selStart = 0
				il.selEnd = 0
			} else {
				w := il.Bounds().Width()
				localX := event.Mouse.X - il.Bounds().A.X
				if localX <= 0 && il.scrollOffset > 0 {
					il.scrollOffset--
					col = il.scrollOffset
				} else if localX >= w-1 && il.scrollOffset+w < len(il.text) {
					il.scrollOffset++
					col = il.scrollOffset + w - 1
				}
				if col > len(il.text) {
					col = len(il.text)
				}
				il.cursorPos = col
				il.selStart = il.dragAnchor
				il.selEnd = il.cursorPos
			}
			il.adjustScroll()
			event.Clear()
		} else if il.dragging {
			il.dragging = false
		}

	case EvKeyboard:
		if event.Key == nil {
			return
		}
		ke := event.Key
		shift := ke.Modifiers&tcell.ModShift != 0

		switch ke.Key {
		case tcell.KeyLeft:
			ctrl := ke.Modifiers&tcell.ModCtrl != 0
			if ctrl {
				if shift {
					// Ctrl+Shift+Left: extend selection word-left.
					if !il.hasSelection() {
						il.selStart = il.cursorPos
						il.selEnd = il.cursorPos
					}
					il.cursorPos = il.wordLeft(il.cursorPos)
					il.selEnd = il.cursorPos
				} else {
					// Ctrl+Left: move cursor word-left, clear selection.
					il.selStart = 0
					il.selEnd = 0
					il.cursorPos = il.wordLeft(il.cursorPos)
				}
			} else if shift {
				// Start selection from current cursorPos if no selection.
				if !il.hasSelection() {
					il.selStart = il.cursorPos
					il.selEnd = il.cursorPos
				}
				if il.cursorPos > 0 {
					il.cursorPos--
					il.selEnd = il.cursorPos
				}
			} else {
				il.selStart = 0
				il.selEnd = 0
				if il.cursorPos > 0 {
					il.cursorPos--
				}
			}
			il.adjustScroll()
			event.Clear()

		case tcell.KeyRight:
			ctrl := ke.Modifiers&tcell.ModCtrl != 0
			if ctrl {
				if shift {
					// Ctrl+Shift+Right: extend selection word-right.
					if !il.hasSelection() {
						il.selStart = il.cursorPos
						il.selEnd = il.cursorPos
					}
					il.cursorPos = il.wordRight(il.cursorPos)
					il.selEnd = il.cursorPos
				} else {
					// Ctrl+Right: move cursor word-right, clear selection.
					il.selStart = 0
					il.selEnd = 0
					il.cursorPos = il.wordRight(il.cursorPos)
				}
			} else if shift {
				if !il.hasSelection() {
					il.selStart = il.cursorPos
					il.selEnd = il.cursorPos
				}
				if il.cursorPos < len(il.text) {
					il.cursorPos++
					il.selEnd = il.cursorPos
				}
			} else {
				il.selStart = 0
				il.selEnd = 0
				if il.cursorPos < len(il.text) {
					il.cursorPos++
				}
			}
			il.adjustScroll()
			event.Clear()

		case tcell.KeyHome:
			if shift {
				if !il.hasSelection() {
					il.selStart = il.cursorPos
					il.selEnd = il.cursorPos
				}
				il.cursorPos = 0
				il.selEnd = il.cursorPos
			} else {
				il.selStart = 0
				il.selEnd = 0
				il.cursorPos = 0
			}
			il.adjustScroll()
			event.Clear()

		case tcell.KeyEnd:
			if shift {
				if !il.hasSelection() {
					il.selStart = il.cursorPos
					il.selEnd = il.cursorPos
				}
				il.cursorPos = len(il.text)
				il.selEnd = il.cursorPos
			} else {
				il.selStart = 0
				il.selEnd = 0
				il.cursorPos = len(il.text)
			}
			il.adjustScroll()
			event.Clear()

		case tcell.KeyBackspace, tcell.KeyBackspace2:
			ctrl := ke.Modifiers&(tcell.ModCtrl|tcell.ModAlt) != 0
			if il.hasSelection() {
				il.deleteSelection()
			} else if ctrl {
				// Ctrl+Backspace: delete from wordLeft(cursorPos) to cursorPos.
				wl := il.wordLeft(il.cursorPos)
				if wl < il.cursorPos {
					il.text = append(il.text[:wl], il.text[il.cursorPos:]...)
					il.cursorPos = wl
				}
			} else if il.cursorPos > 0 {
				il.text = append(il.text[:il.cursorPos-1], il.text[il.cursorPos:]...)
				il.cursorPos--
			}
			il.adjustScroll()
			event.Clear()

		case tcell.KeyDelete:
			ctrl := ke.Modifiers&tcell.ModCtrl != 0
			if il.hasSelection() {
				il.deleteSelection()
			} else if ctrl {
				// Ctrl+Delete: delete from cursorPos to wordRight(cursorPos).
				wr := il.wordRight(il.cursorPos)
				if wr > il.cursorPos {
					il.text = append(il.text[:il.cursorPos], il.text[wr:]...)
				}
			} else if il.cursorPos < len(il.text) {
				il.text = append(il.text[:il.cursorPos], il.text[il.cursorPos+1:]...)
			}
			il.adjustScroll()
			event.Clear()

		case tcell.KeyRune:
			// Skip if Ctrl or Alt modifier is set.
			if ke.Modifiers&(tcell.ModCtrl|tcell.ModAlt) != 0 {
				return
			}
			// Replace selection if active.
			if il.hasSelection() {
				il.deleteSelection()
			}
			if il.overwrite && il.cursorPos < len(il.text) {
				// Overwrite mode: replace char at cursorPos (no maxLen check needed — not changing length).
				il.text[il.cursorPos] = ke.Rune
				il.cursorPos++
			} else {
				// Insert mode (or at end in overwrite mode): insert rune at cursor.
				// No-op if at maxLen.
				if il.maxLen > 0 && len(il.text) >= il.maxLen {
					event.Clear()
					return
				}
				il.text = append(il.text, 0)
				copy(il.text[il.cursorPos+1:], il.text[il.cursorPos:])
				il.text[il.cursorPos] = ke.Rune
				il.cursorPos++
			}
			il.adjustScroll()
			event.Clear()

		case tcell.KeyInsert:
			il.overwrite = !il.overwrite
			event.Clear()

		default:
			// Ctrl commands.
			switch ke.Key {
			case tcell.KeyCtrlA:
				il.selStart = 0
				il.selEnd = len(il.text)
				il.cursorPos = len(il.text)
				event.Clear()

			case tcell.KeyCtrlC:
				if il.hasSelection() {
					lo, hi := il.normalizedSel()
					clipboard = string(il.text[lo:hi])
				}
				event.Clear()

			case tcell.KeyCtrlX:
				if il.hasSelection() {
					lo, hi := il.normalizedSel()
					clipboard = string(il.text[lo:hi])
					il.deleteSelection()
					il.adjustScroll()
				}
				event.Clear()

			case tcell.KeyCtrlV:
				if il.hasSelection() {
					il.deleteSelection()
				}
				paste := []rune(clipboard)
				// Respect maxLen.
				if il.maxLen > 0 {
					remaining := il.maxLen - len(il.text)
					if remaining <= 0 {
						event.Clear()
						return
					}
					if len(paste) > remaining {
						paste = paste[:remaining]
					}
				}
				// Insert paste at cursor.
				newText := make([]rune, len(il.text)+len(paste))
				copy(newText, il.text[:il.cursorPos])
				copy(newText[il.cursorPos:], paste)
				copy(newText[il.cursorPos+len(paste):], il.text[il.cursorPos:])
				il.text = newText
				il.cursorPos += len(paste)
				il.adjustScroll()
				event.Clear()

			case tcell.KeyCtrlY:
				il.text = il.text[:0]
				il.cursorPos = 0
				il.selStart = 0
				il.selEnd = 0
				il.scrollOffset = 0
				event.Clear()

			// Tab, Escape, Enter and other unhandled keys pass through.
			}
		}
	}
}
