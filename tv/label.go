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
	light    bool
}

func NewLabel(bounds Rect, label string, link View) *Label {
	l := &Label{
		label: label,
		link:  link,
	}
	l.SetBounds(bounds)
	l.SetState(SfVisible, true)
	l.SetOptions(OfPreProcess, true)
	l.SetOptions(OfPostProcess, true)

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
		if l.light {
			normalStyle = cs.LabelHighlight
		} else {
			normalStyle = cs.LabelNormal
		}
		shortcutStyle = cs.LabelShortcut
	}

	// Background fill: entire bounds width with normal/highlight style
	b := l.Bounds()
	buf.Fill(NewRect(0, 0, b.Width(), 1), ' ', normalStyle)

	// Text starts at column 1 (column 0 is monochrome marker margin)
	x := 1
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
	if event.What == EvMouse && event.Mouse != nil {
		if event.Mouse.Button&tcell.Button1 != 0 && l.link != nil {
			if l.link.HasOption(OfSelectable) {
				if owner := l.Owner(); owner != nil {
					owner.SetFocusedChild(l.link)
				}
			}
			event.Clear()
		}
		return
	}

	if event.What == EvBroadcast && l.link != nil {
		switch event.Command {
		case CmReceivedFocus:
			if event.Info == l.link {
				l.light = true
			}
		case CmReleasedFocus:
			if event.Info == l.link {
				l.light = false
			}
		}
		return
	}

	if l.link == nil || l.shortcut == 0 {
		return
	}
	if event.What != EvKeyboard || event.Key == nil {
		return
	}
	if event.Key.Modifiers&tcell.ModAlt != 0 && event.Key.Key == tcell.KeyRune {
		if unicode.ToLower(event.Key.Rune) == unicode.ToLower(l.shortcut) {
			if l.link.HasOption(OfSelectable) {
				if owner := l.Owner(); owner != nil {
					owner.SetFocusedChild(l.link)
				}
			}
			event.Clear()
		}
	}
}
