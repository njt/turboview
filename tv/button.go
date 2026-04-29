package tv

import (
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var _ Widget = (*Button)(nil)

type Button struct {
	BaseView
	title     string
	shortcut  rune
	command   CommandCode
	bfDefault bool
	amDefault bool
}

type ButtonOption func(*Button)

func WithDefault() ButtonOption {
	return func(b *Button) {
		b.bfDefault = true
		b.amDefault = true
		b.SetOptions(OfPostProcess, true)
	}
}

func NewButton(bounds Rect, title string, command CommandCode, opts ...ButtonOption) *Button {
	b := &Button{title: title, command: command}
	b.SetBounds(bounds)
	b.SetState(SfVisible, true)
	b.SetOptions(OfSelectable, true)
	b.SetOptions(OfFirstClick, true)  // click fires even on first (focus) click
	b.SetOptions(OfPostProcess, true) // all buttons get OfPostProcess

	segments := ParseTildeLabel(title)
	for _, seg := range segments {
		if seg.Shortcut && len(seg.Text) > 0 {
			b.shortcut, _ = utf8.DecodeRuneInString(seg.Text)
			break
		}
	}

	for _, opt := range opts {
		opt(b)
	}
	b.SetSelf(b)
	return b
}

func (b *Button) Title() string        { return b.title }
func (b *Button) Command() CommandCode { return b.command }
func (b *Button) IsDefault() bool      { return b.amDefault }

func (b *Button) SetState(flag ViewState, on bool) {
	b.BaseView.SetState(flag, on)
	if flag&SfSelected != 0 && !b.bfDefault {
		if on {
			b.broadcastToOwner(CmGrabDefault)
			b.amDefault = true
		} else {
			b.broadcastToOwner(CmReleaseDefault)
			b.amDefault = false
		}
	}
}

func (b *Button) broadcastToOwner(cmd CommandCode) {
	if b.owner == nil {
		return
	}
	ev := &Event{What: EvBroadcast, Command: cmd, Info: b}
	for _, child := range b.owner.Children() {
		if child.HasState(SfDisabled) {
			continue
		}
		child.HandleEvent(ev)
	}
}

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
		if b.amDefault {
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
	// Click-to-focus (from BaseView) then fire on Button1
	if event.What == EvMouse && event.Mouse != nil {
		b.BaseView.HandleEvent(event)
		if event.IsCleared() {
			return
		}
		if event.Mouse.Button&tcell.Button1 != 0 {
			b.press(event)
		}
		return
	}

	// Broadcast handling
	if event.What == EvBroadcast {
		switch event.Command {
		case CmDefault:
			if b.amDefault {
				b.press(event)
			}
		case CmGrabDefault:
			if b.bfDefault {
				b.amDefault = false
			}
		case CmReleaseDefault:
			if b.bfDefault {
				b.amDefault = true
			}
		}
		return
	}

	if event.What == EvKeyboard && event.Key != nil {
		switch event.Key.Key {
		case tcell.KeyRune:
			if event.Key.Rune == ' ' && b.HasState(SfSelected) {
				b.press(event)
			} else if event.Key.Modifiers&tcell.ModAlt != 0 && b.shortcut != 0 {
				if unicode.ToLower(event.Key.Rune) == unicode.ToLower(b.shortcut) {
					b.press(event)
				}
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
