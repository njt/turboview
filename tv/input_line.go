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
	return il
}

func (il *InputLine) Text() string          { return string(il.text) }
func (il *InputLine) CursorPos() int        { return il.cursorPos }
func (il *InputLine) Selection() (int, int) { return il.selStart, il.selEnd }

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
}

func (il *InputLine) HandleEvent(event *Event) {
	switch event.What {
	case EvMouse:
		if event.Mouse != nil && event.Mouse.Button&tcell.Button1 != 0 {
			col := event.Mouse.X - il.Bounds().A.X + il.scrollOffset
			if col < 0 {
				col = 0
			}
			if col > len(il.text) {
				col = len(il.text)
			}
			il.cursorPos = col
			il.selStart = 0
			il.selEnd = 0
			il.adjustScroll()
			event.Clear()
		}

	case EvKeyboard:
		if event.Key == nil {
			return
		}
		ke := event.Key
		shift := ke.Modifiers&tcell.ModShift != 0

		switch ke.Key {
		case tcell.KeyLeft:
			if shift {
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
			if shift {
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
			if il.hasSelection() {
				il.deleteSelection()
			} else if il.cursorPos > 0 {
				il.text = append(il.text[:il.cursorPos-1], il.text[il.cursorPos:]...)
				il.cursorPos--
			}
			il.adjustScroll()
			event.Clear()

		case tcell.KeyDelete:
			if il.hasSelection() {
				il.deleteSelection()
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
			// No-op if at maxLen.
			if il.maxLen > 0 && len(il.text) >= il.maxLen {
				event.Clear()
				return
			}
			// Insert rune at cursor.
			il.text = append(il.text, 0)
			copy(il.text[il.cursorPos+1:], il.text[il.cursorPos:])
			il.text[il.cursorPos] = ke.Rune
			il.cursorPos++
			il.adjustScroll()
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

			// Tab, Escape, Enter and other unhandled keys pass through.
			}
		}
	}
}
