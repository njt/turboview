package tv

import (
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type StatusItem struct {
	Label      string
	KeyBinding KeyBinding
	Command    CommandCode
	HelpCtx    HelpContext
}

func NewStatusItem(label string, kb KeyBinding, cmd CommandCode) *StatusItem {
	return &StatusItem{
		Label:      label,
		KeyBinding: kb,
		Command:    cmd,
	}
}

type StatusLine struct {
	BaseView
	items []*StatusItem
}

func NewStatusLine(items ...*StatusItem) *StatusLine {
	sl := &StatusLine{
		items: items,
	}
	sl.SetState(SfVisible, true)
	return sl
}

func (sl *StatusLine) Draw(buf *DrawBuffer) {
	w := sl.Bounds().Width()
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	if cs := sl.ColorScheme(); cs != nil {
		normalStyle = cs.StatusNormal
		shortcutStyle = cs.StatusShortcut
	}

	buf.Fill(NewRect(0, 0, w, 1), ' ', normalStyle)

	x := 1
	for i, item := range sl.items {
		if i > 0 {
			x += 2
		}
		segments := ParseTildeLabel(item.Label)
		for _, seg := range segments {
			style := normalStyle
			if seg.Shortcut {
				style = shortcutStyle
			}
			buf.WriteStr(x, 0, seg.Text, style)
			x += utf8.RuneCountInString(seg.Text)
		}
	}
}

func (sl *StatusLine) HandleEvent(event *Event) {
	if event.What != EvKeyboard || event.Key == nil {
		return
	}
	for _, item := range sl.items {
		if item.KeyBinding.Matches(event.Key) {
			event.What = EvCommand
			event.Command = item.Command
			event.Key = nil
			return
		}
	}
}
