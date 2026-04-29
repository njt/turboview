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

// TestLabelPhaseNewLabelDoesNotSetOfPreProcess verifies NewLabel no longer sets OfPreProcess.
func TestLabelPhaseNewLabelDoesNotSetOfPreProcess(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)

	if label.HasOption(OfPreProcess) {
		t.Error("NewLabel must not set OfPreProcess; label should run in post-process phase")
	}
}

// TestLabelPhasePostProcessDoesNotInterceptBeforeFocusedChild verifies that with
// OfPostProcess, a focused child handles Alt+shortcut before the label does.
// This is a behavioral inversion of the old OfPreProcess behavior.
func TestLabelPhasePostProcessDoesNotInterceptBeforeFocusedChild(t *testing.T) {
	// We use a spy to record whether the focused child received the event
	// before it was cleared by the label.
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)

	// A focused child that clears the event when it receives Alt+n,
	// simulating a widget that wants to handle the shortcut itself.
	interceptor := &altInterceptorView{intercept: 'n'}
	interceptor.SetBounds(NewRect(20, 0, 10, 1))
	interceptor.SetState(SfVisible, true)
	interceptor.SetOptions(OfSelectable, true)

	g.Insert(label)
	g.Insert(linked)
	g.Insert(interceptor) // interceptor gets focus

	if g.FocusedChild() != interceptor {
		t.Fatalf("precondition: FocusedChild() = %v, want interceptor", g.FocusedChild())
	}

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	// Interceptor cleared the event before the label's post-process phase ran,
	// so label should NOT have changed focus to linked.
	if g.FocusedChild() == linked {
		t.Errorf("FocusedChild() = linked; with OfPostProcess, focused child should handle Alt+n before label")
	}
	if !interceptor.gotEvent {
		t.Errorf("interceptor did not receive the Alt+n event")
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
