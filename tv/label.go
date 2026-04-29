package tv

import (
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var _ Widget = (*Label)(nil)

type Label struct {
	BaseView
	label    string
	link     View
	shortcut rune
}

func NewLabel(bounds Rect, label string, link View) *Label {
	l := &Label{
		label: label,
		link:  link,
	}
	l.SetBounds(bounds)
	l.SetState(SfVisible, true)
	l.SetOptions(OfPreProcess, true)

	segments := ParseTildeLabel(label)
	for _, seg := range segments {
		if seg.Shortcut && len(seg.Text) > 0 {
			l.shortcut, _ = utf8.DecodeRuneInString(seg.Text)
			break
		}
	}

	l.SetSelf(l)
	return l
}

func (l *Label) Draw(buf *DrawBuffer) {
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	if cs := l.ColorScheme(); cs != nil {
		normalStyle = cs.LabelNormal
		shortcutStyle = cs.LabelShortcut
	}

	x := 0
	segments := ParseTildeLabel(l.label)
	for _, seg := range segments {
		style := normalStyle
		if seg.Shortcut {
			style = shortcutStyle
		}
		buf.WriteStr(x, 0, seg.Text, style)
		x += utf8.RuneCountInString(seg.Text)
	}
}

func (l *Label) HandleEvent(event *Event) {
	if l.link == nil || l.shortcut == 0 {
		return
	}
	if event.What != EvKeyboard || event.Key == nil {
		return
	}
	if event.Key.Modifiers&tcell.ModAlt != 0 && event.Key.Key == tcell.KeyRune {
		if unicode.ToLower(event.Key.Rune) == unicode.ToLower(l.shortcut) {
			if owner := l.Owner(); owner != nil {
				owner.SetFocusedChild(l.link)
			}
			event.Clear()
		}
	}
}
