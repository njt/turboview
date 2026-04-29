package tv

import (
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var _ Widget = (*Button)(nil)

type Button struct {
	BaseView
	title     string
	command   CommandCode
	isDefault bool
}

type ButtonOption func(*Button)

func WithDefault() ButtonOption {
	return func(b *Button) {
		b.isDefault = true
		b.SetOptions(OfPostProcess, true)
	}
}

func NewButton(bounds Rect, title string, command CommandCode, opts ...ButtonOption) *Button {
	b := &Button{title: title, command: command}
	b.SetBounds(bounds)
	b.SetState(SfVisible, true)
	b.SetOptions(OfSelectable, true)
	for _, opt := range opts {
		opt(b)
	}
	b.SetSelf(b)
	return b
}

func (b *Button) Title() string        { return b.title }
func (b *Button) Command() CommandCode { return b.command }
func (b *Button) IsDefault() bool      { return b.isDefault }

func (b *Button) Draw(buf *DrawBuffer) {
	w, h := b.Bounds().Width(), b.Bounds().Height()
	if w < 4 || h < 1 {
		return
	}

	cs := b.ColorScheme()
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	shadowStyle := tcell.StyleDefault
	if cs != nil {
		if b.isDefault {
			normalStyle = cs.ButtonDefault
		} else {
			normalStyle = cs.ButtonNormal
		}
		shortcutStyle = cs.ButtonShortcut
		shadowStyle = cs.ButtonShadow
	}

	// Button face area (excluding shadow)
	faceW := w
	faceH := h
	if h >= 2 {
		faceW = w - 1
		faceH = h - 1
	}

	// Fill face
	buf.Fill(NewRect(0, 0, faceW, faceH), ' ', normalStyle)

	// Focus cursor
	if b.HasState(SfSelected) {
		buf.WriteChar(0, 0, '►', normalStyle)
	}

	// Bracket and title: "[ Title ]"
	segments := ParseTildeLabel(b.title)
	titleLen := 0
	for _, seg := range segments {
		titleLen += utf8.RuneCountInString(seg.Text)
	}
	bracketText := titleLen + 4 // "[ " + title + " ]"
	startX := (faceW - bracketText) / 2
	if startX < 0 {
		startX = 0
	}

	buf.WriteChar(startX, 0, '[', normalStyle)
	buf.WriteChar(startX+1, 0, ' ', normalStyle)
	x := startX + 2
	for _, seg := range segments {
		style := normalStyle
		if seg.Shortcut {
			style = shortcutStyle
		}
		buf.WriteStr(x, 0, seg.Text, style)
		x += utf8.RuneCountInString(seg.Text)
	}
	buf.WriteChar(x, 0, ' ', normalStyle)
	buf.WriteChar(x+1, 0, ']', normalStyle)

	// Shadow (if height >= 2)
	if h >= 2 {
		for y := 1; y < h; y++ {
			buf.WriteChar(w-1, y, ' ', shadowStyle)
		}
		for x := 1; x < w; x++ {
			buf.WriteChar(x, h-1, ' ', shadowStyle)
		}
	}
}

func (b *Button) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		if event.Mouse.Button&tcell.Button1 != 0 {
			b.press(event)
		}
		return
	}
	if event.What == EvKeyboard && event.Key != nil {
		switch event.Key.Key {
		case tcell.KeyEnter:
			b.press(event)
		case tcell.KeyRune:
			if event.Key.Rune == ' ' {
				b.press(event)
			}
		}
	}
}

func (b *Button) press(event *Event) {
	event.What = EvCommand
	event.Command = b.command
	event.Key = nil
	event.Mouse = nil
}
