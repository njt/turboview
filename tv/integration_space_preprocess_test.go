package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// TestSpaceReachesInputLineWithCheckBoxesInGroup verifies that Space typed into
// a focused InputLine inserts a space character, even when an unfocused
// CheckBoxes widget with OfPreProcess is present in the same Group.
func TestSpaceReachesInputLineWithCheckBoxesInGroup(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 60, 20))
	g.scheme = theme.BorlandBlue

	il := NewInputLine(NewRect(0, 0, 20, 1), 40)
	cbs := NewCheckBoxes(NewRect(0, 2, 25, 3), []string{"Read only", "Hidden"})
	g.Insert(il)
	g.Insert(cbs)

	// Focus the InputLine.
	g.SetFocusedChild(il)
	if !il.HasState(SfSelected) {
		t.Fatal("InputLine should be focused")
	}

	// Type "a b" — a, space, b.
	for _, r := range []rune{'a', ' ', 'b'} {
		ev := &Event{
			What: EvKeyboard,
			Key:  &KeyEvent{Key: tcell.KeyRune, Rune: r, Modifiers: tcell.ModNone},
		}
		g.HandleEvent(ev)
	}

	got := il.Text()
	if got != "a b" {
		t.Errorf("InputLine.Text() = %q, want %q", got, "a b")
	}
}

// TestSpaceReachesInputLineWithRadioButtonsInGroup is the same test but with
// RadioButtons instead of CheckBoxes.
func TestSpaceReachesInputLineWithRadioButtonsInGroup(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 60, 20))
	g.scheme = theme.BorlandBlue

	il := NewInputLine(NewRect(0, 0, 20, 1), 40)
	rbs := NewRadioButtons(NewRect(0, 2, 25, 3), []string{"Text", "Binary"})
	g.Insert(il)
	g.Insert(rbs)

	g.SetFocusedChild(il)

	for _, r := range []rune{'a', ' ', 'b'} {
		ev := &Event{
			What: EvKeyboard,
			Key:  &KeyEvent{Key: tcell.KeyRune, Rune: r, Modifiers: tcell.ModNone},
		}
		g.HandleEvent(ev)
	}

	got := il.Text()
	if got != "a b" {
		t.Errorf("InputLine.Text() = %q, want %q", got, "a b")
	}
}

// TestCheckBoxSpaceStillWorksWhenFocused verifies Space still toggles a
// CheckBox when CheckBoxes is focused.
func TestCheckBoxSpaceStillWorksWhenFocused(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 60, 20))
	g.scheme = theme.BorlandBlue

	cbs := NewCheckBoxes(NewRect(0, 0, 25, 3), []string{"Read only", "Hidden"})
	il := NewInputLine(NewRect(0, 5, 20, 1), 40)
	g.Insert(cbs)
	g.Insert(il)

	g.SetFocusedChild(cbs)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: ' ', Modifiers: tcell.ModNone},
	}
	g.HandleEvent(ev)

	// Last inserted CheckBox ("Hidden", index 1) is focused by default.
	if cbs.Values() != 2 {
		t.Errorf("CheckBoxes.Values() = %d, want 2 (second item toggled)", cbs.Values())
	}
}
