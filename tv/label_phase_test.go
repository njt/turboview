package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestLabelPhaseNewLabelSetsOfPostProcess verifies NewLabel sets the OfPostProcess option
// instead of OfPreProcess, so focused views get priority for Alt+shortcut.
func TestLabelPhaseNewLabelSetsOfPostProcess(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)

	if !label.HasOption(OfPostProcess) {
		t.Error("NewLabel did not set OfPostProcess")
	}
}

// TestLabelPhasePostProcessActivatesWhenFocusedChildDoesNotConsume verifies that
// when the focused child does NOT clear the Alt+shortcut event, the label's
// post-process handler still fires and focuses the link.
func TestLabelPhasePostProcessActivatesWhenFocusedChildDoesNotConsume(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() != linked {
		t.Errorf("FocusedChild() = %v, want linked; label post-process should activate when focused child ignores Alt+n", g.FocusedChild())
	}
}

// altInterceptorView is a test double that clears Alt+<rune> events it sees,
// simulating a focused widget that consumes the shortcut.
type altInterceptorView struct {
	BaseView
	intercept rune
	gotEvent  bool
}

func (a *altInterceptorView) HandleEvent(event *Event) {
	if event.What == EvKeyboard && event.Key != nil {
		if event.Key.Modifiers&tcell.ModAlt != 0 && event.Key.Key == tcell.KeyRune {
			if event.Key.Rune == a.intercept {
				a.gotEvent = true
				event.Clear()
			}
		}
	}
}
