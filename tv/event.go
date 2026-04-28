package tv

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
)

type EventType uint16

const (
	EvNothing   EventType = 0
	EvMouse     EventType = 1 << 0
	EvKeyboard  EventType = 1 << 1
	EvCommand   EventType = 1 << 2
	EvBroadcast EventType = 1 << 3
)

type MouseEvent struct {
	X, Y       int
	Button     tcell.ButtonMask
	Modifiers  tcell.ModMask
	ClickCount int
}

type KeyEvent struct {
	Key       tcell.Key
	Rune      rune
	Modifiers tcell.ModMask
}

type Event struct {
	What    EventType
	Mouse   *MouseEvent
	Key     *KeyEvent
	Command CommandCode
	Info    any
}

func (e *Event) Clear() {
	e.What = EvNothing
}

func (e *Event) IsCleared() bool {
	return e.What == EvNothing
}

type KeyBinding struct {
	Key  tcell.Key
	Rune rune
	Mod  tcell.ModMask
}

func KbCtrl(ch rune) KeyBinding {
	ch = unicode.ToUpper(ch)
	return KeyBinding{Key: tcell.Key(ch - 'A' + 1), Mod: tcell.ModCtrl}
}

func KbAlt(ch rune) KeyBinding {
	return KeyBinding{Key: tcell.KeyRune, Rune: unicode.ToLower(ch), Mod: tcell.ModAlt}
}

func KbFunc(n int) KeyBinding {
	return KeyBinding{Key: tcell.Key(int(tcell.KeyF1) + n - 1)}
}

func KbNone() KeyBinding {
	return KeyBinding{}
}

func (kb KeyBinding) Matches(ke *KeyEvent) bool {
	if kb == (KeyBinding{}) {
		return false
	}
	if kb.Key == tcell.KeyRune {
		return unicode.ToLower(ke.Rune) == unicode.ToLower(kb.Rune) &&
			ke.Modifiers&kb.Mod == kb.Mod
	}
	return ke.Key == kb.Key && ke.Modifiers&kb.Mod == kb.Mod
}
