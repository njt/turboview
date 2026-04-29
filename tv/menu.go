package tv

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type MenuItem struct {
	Label    string
	Command  CommandCode
	Accel    KeyBinding
	Disabled bool
}

func NewMenuItem(label string, cmd CommandCode, accel KeyBinding) *MenuItem {
	return &MenuItem{Label: label, Command: cmd, Accel: accel}
}

type SubMenu struct {
	Label string
	Items []any
}

func NewSubMenu(label string, items ...any) *SubMenu {
	return &SubMenu{Label: label, Items: items}
}

type MenuSeparator struct{}

func NewMenuSeparator() *MenuSeparator {
	return &MenuSeparator{}
}

func FormatAccel(kb KeyBinding) string {
	if kb == (KeyBinding{}) {
		return ""
	}
	// KbCtrl stores Key as tcell.Key(ch - 'A' + 1), not KeyRune
	// The letter is always displayed uppercase per spec.
	if kb.Mod&tcell.ModCtrl != 0 && kb.Key >= 1 && kb.Key <= 26 {
		ch := rune(kb.Key) + 'A' - 1
		return fmt.Sprintf("Ctrl+%c", ch)
	}
	if kb.Mod&tcell.ModAlt != 0 && kb.Key == tcell.KeyRune {
		return fmt.Sprintf("Alt+%c", unicode.ToUpper(kb.Rune))
	}
	if kb.Key >= tcell.KeyF1 && kb.Key <= tcell.KeyF12 {
		n := int(kb.Key-tcell.KeyF1) + 1
		return fmt.Sprintf("F%d", n)
	}
	return ""
}

func tildeTextLen(label string) int {
	segments := ParseTildeLabel(label)
	n := 0
	for _, seg := range segments {
		n += utf8.RuneCountInString(seg.Text)
	}
	return n
}

func menuItemWidth(item *MenuItem) int {
	w := tildeTextLen(item.Label)
	accel := FormatAccel(item.Accel)
	if accel != "" {
		w += 2 + utf8.RuneCountInString(accel)
	}
	return w
}

func popupWidth(items []any) int {
	maxW := 0
	for _, item := range items {
		if mi, ok := item.(*MenuItem); ok {
			w := menuItemWidth(mi)
			if w > maxW {
				maxW = w
			}
		}
	}
	return maxW + 4
}

func popupHeight(items []any) int {
	return len(items) + 2
}
