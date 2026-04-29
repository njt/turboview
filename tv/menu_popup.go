package tv

import (
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

type MenuPopup struct {
	items    []any
	selected int
	result   CommandCode
	bounds   Rect
}

func NewMenuPopup(items []any, x, y int) *MenuPopup {
	w := popupWidth(items)
	h := popupHeight(items)
	mp := &MenuPopup{
		items:    items,
		selected: -1,
		bounds:   NewRect(x, y, w, h),
	}
	mp.selectNext(true)
	return mp
}

func (mp *MenuPopup) Bounds() Rect        { return mp.bounds }
func (mp *MenuPopup) Result() CommandCode { return mp.result }
func (mp *MenuPopup) Selected() int       { return mp.selected }

func (mp *MenuPopup) Draw(buf *DrawBuffer, cs *theme.ColorScheme) {
	w, h := mp.bounds.Width(), mp.bounds.Height()
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault
	disabledStyle := tcell.StyleDefault
	if cs != nil {
		normalStyle = cs.MenuNormal
		shortcutStyle = cs.MenuShortcut
		selectedStyle = cs.MenuSelected
		disabledStyle = cs.MenuDisabled
	}

	// Border
	buf.WriteChar(0, 0, '┌', normalStyle)
	buf.WriteChar(w-1, 0, '┐', normalStyle)
	buf.WriteChar(0, h-1, '└', normalStyle)
	buf.WriteChar(w-1, h-1, '┘', normalStyle)
	for x := 1; x < w-1; x++ {
		buf.WriteChar(x, 0, '─', normalStyle)
		buf.WriteChar(x, h-1, '─', normalStyle)
	}
	for y := 1; y < h-1; y++ {
		buf.WriteChar(0, y, '│', normalStyle)
		buf.WriteChar(w-1, y, '│', normalStyle)
	}

	innerW := w - 2

	for i, item := range mp.items {
		row := i + 1
		switch it := item.(type) {
		case *MenuSeparator:
			buf.WriteChar(0, row, '├', normalStyle)
			for x := 1; x < w-1; x++ {
				buf.WriteChar(x, row, '─', normalStyle)
			}
			buf.WriteChar(w-1, row, '┤', normalStyle)

		case *MenuItem:
			isSelected := i == mp.selected
			style := normalStyle
			scStyle := shortcutStyle
			if it.Disabled {
				style = disabledStyle
				scStyle = disabledStyle
			} else if isSelected {
				style = selectedStyle
				scStyle = selectedStyle
			}

			buf.Fill(NewRect(1, row, innerW, 1), ' ', style)

			x := 1
			segments := ParseTildeLabel(it.Label)
			for _, seg := range segments {
				s := style
				if seg.Shortcut && !it.Disabled && !isSelected {
					s = scStyle
				}
				buf.WriteStr(x, row, seg.Text, s)
				x += utf8.RuneCountInString(seg.Text)
			}

			accel := FormatAccel(it.Accel)
			if accel != "" {
				ax := 1 + innerW - utf8.RuneCountInString(accel)
				buf.WriteStr(ax, row, accel, style)
			}

		case *SubMenu:
			isSelected := i == mp.selected
			style := disabledStyle
			if isSelected {
				style = selectedStyle
			}
			buf.Fill(NewRect(1, row, innerW, 1), ' ', style)
			segments := ParseTildeLabel(it.Label)
			x := 1
			for _, seg := range segments {
				buf.WriteStr(x, row, seg.Text, style)
				x += utf8.RuneCountInString(seg.Text)
			}
			buf.WriteStr(1+innerW-1, row, "►", style)
		}
	}
}

func (mp *MenuPopup) HandleEvent(event *Event) {
	if event.What == EvKeyboard && event.Key != nil {
		switch event.Key.Key {
		case tcell.KeyDown:
			mp.selectNext(false)
		case tcell.KeyUp:
			mp.selectPrev()
		case tcell.KeyEnter:
			mp.fireSelected()
		case tcell.KeyEscape:
			mp.result = CmCancel
		case tcell.KeyRune:
			mp.matchShortcut(event.Key.Rune)
		}
		return
	}

	if event.What == EvMouse && event.Mouse != nil {
		my := event.Mouse.Y
		if my >= 1 && my < mp.bounds.Height()-1 {
			idx := my - 1
			if idx >= 0 && idx < len(mp.items) {
				if _, ok := mp.items[idx].(*MenuSeparator); !ok {
					mp.selected = idx
					if event.Mouse.Button&tcell.Button1 != 0 {
						mp.fireSelected()
					}
				}
			}
		}
	}
}

func (mp *MenuPopup) selectNext(initial bool) {
	n := len(mp.items)
	if n == 0 {
		return
	}
	start := mp.selected + 1
	if initial && mp.selected < 0 {
		start = 0
	}
	for i := 0; i < n; i++ {
		idx := (start + i) % n
		if _, ok := mp.items[idx].(*MenuSeparator); !ok {
			mp.selected = idx
			return
		}
	}
}

func (mp *MenuPopup) selectPrev() {
	n := len(mp.items)
	if n == 0 {
		return
	}
	start := mp.selected - 1
	if start < 0 {
		start = n - 1
	}
	for i := 0; i < n; i++ {
		idx := (start - i + n) % n
		if _, ok := mp.items[idx].(*MenuSeparator); !ok {
			mp.selected = idx
			return
		}
	}
}

func (mp *MenuPopup) fireSelected() {
	if mp.selected < 0 || mp.selected >= len(mp.items) {
		return
	}
	if mi, ok := mp.items[mp.selected].(*MenuItem); ok && !mi.Disabled {
		mp.result = mi.Command
	}
}

func (mp *MenuPopup) matchShortcut(r rune) {
	r = unicode.ToLower(r)
	for i, item := range mp.items {
		mi, ok := item.(*MenuItem)
		if !ok || mi.Disabled {
			continue
		}
		segments := ParseTildeLabel(mi.Label)
		for _, seg := range segments {
			if seg.Shortcut && len(seg.Text) > 0 {
				sc, _ := utf8.DecodeRuneInString(seg.Text)
				if unicode.ToLower(sc) == r {
					mp.selected = i
					mp.result = mi.Command
					return
				}
			}
		}
	}
}
