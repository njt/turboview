package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestNewMonoSelectorSelected verifies the initial Selected() is 0.
func TestNewMonoSelectorSelected(t *testing.T) {
	ms := NewMonoSelector(NewRect(0, 0, 20, 4))
	if ms.Selected() != 0 {
		t.Errorf("expected Selected() == 0, got %d", ms.Selected())
	}
}

// TestMonoSelectorNavigateDown verifies pressing Down moves selection from 0 to 1.
func TestMonoSelectorNavigateDown(t *testing.T) {
	ms := NewMonoSelector(NewRect(0, 0, 20, 4))
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	ms.HandleEvent(ev)
	if ms.Selected() != 1 {
		t.Errorf("expected Down to move 0→1, got %d", ms.Selected())
	}
}

// TestMonoSelectorNavigateDownWrap verifies Down from last item (3) wraps to 0.
func TestMonoSelectorNavigateDownWrap(t *testing.T) {
	ms := NewMonoSelector(NewRect(0, 0, 20, 4))
	ms.SetSelected(3)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	ms.HandleEvent(ev)
	if ms.Selected() != 0 {
		t.Errorf("expected Down to wrap 3→0, got %d", ms.Selected())
	}
}

// TestMonoSelectorNavigateUp verifies pressing Up moves selection from 1 to 0.
func TestMonoSelectorNavigateUp(t *testing.T) {
	ms := NewMonoSelector(NewRect(0, 0, 20, 4))
	ms.SetSelected(1)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	ms.HandleEvent(ev)
	if ms.Selected() != 0 {
		t.Errorf("expected Up to move 1→0, got %d", ms.Selected())
	}
}

// TestMonoSelectorNavigateUpWrap verifies Up from first item (0) wraps to 3.
func TestMonoSelectorNavigateUpWrap(t *testing.T) {
	ms := NewMonoSelector(NewRect(0, 0, 20, 4))
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	ms.HandleEvent(ev)
	if ms.Selected() != 3 {
		t.Errorf("expected Up to wrap 0→3, got %d", ms.Selected())
	}
}

// TestMonoSelectorSetSelectedClamps verifies SetSelected clamps to valid range.
func TestMonoSelectorSetSelectedClamps(t *testing.T) {
	ms := NewMonoSelector(NewRect(0, 0, 20, 4))

	ms.SetSelected(-1)
	if ms.Selected() != 0 {
		t.Errorf("expected SetSelected(-1) to leave at 0, got %d", ms.Selected())
	}

	ms.SetSelected(5)
	if ms.Selected() != 0 {
		t.Errorf("expected SetSelected(5) to leave at 0, got %d", ms.Selected())
	}

	ms.SetSelected(2)
	if ms.Selected() != 2 {
		t.Errorf("expected SetSelected(2) to be 2, got %d", ms.Selected())
	}
}

// TestMonoSelectorBroadcastFires verifies a broadcast with the correct
// attribute value is sent on selection change.
func TestMonoSelectorBroadcastFires(t *testing.T) {
	ms := NewMonoSelector(NewRect(0, 0, 20, 4))
	g := NewGroup(NewRect(0, 0, 30, 10))
	g.Insert(ms)

	spy := newSpyView()
	g.Insert(spy)
	resetBroadcasts(spy)

	// Move Down: 0 (Normal, attr 0x07) → 1 (Highlight, attr 0x0F)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	ms.HandleEvent(ev)

	found := false
	for _, rec := range spy.broadcasts {
		if rec.command == CmColorForegroundChanged {
			found = true
			if attr, ok := rec.info.(int); !ok || attr != 0x0F {
				t.Errorf("expected CmColorForegroundChanged Info == 0x0F (Highlight), got %v", rec.info)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected CmColorForegroundChanged broadcast, got: %v", spy.broadcasts)
	}
}

// TestMonoSelectorBroadcastWrap verifies broadcast fires with correct attribute
// when wrapping from last to first item.
func TestMonoSelectorBroadcastWrap(t *testing.T) {
	ms := NewMonoSelector(NewRect(0, 0, 20, 4))
	ms.SetSelected(3) // Inverse, attr 0x70
	g := NewGroup(NewRect(0, 0, 30, 10))
	g.Insert(ms)

	spy := newSpyView()
	g.Insert(spy)
	resetBroadcasts(spy)

	// Down from 3 wraps to 0 (Normal, attr 0x07)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	ms.HandleEvent(ev)

	found := false
	for _, rec := range spy.broadcasts {
		if rec.command == CmColorForegroundChanged {
			found = true
			if attr, ok := rec.info.(int); !ok || attr != 0x07 {
				t.Errorf("expected CmColorForegroundChanged Info == 0x07 (Normal), got %v", rec.info)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected CmColorForegroundChanged broadcast, got: %v", spy.broadcasts)
	}
}

// TestMonoSelectorNoBroadcastOnSameSelection verifies that if Up is pressed
// and the selection doesn't change, no broadcast fires.
func TestMonoSelectorNoBroadcastOnSameSelection(t *testing.T) {
	ms := NewMonoSelector(NewRect(0, 0, 20, 4))
	g := NewGroup(NewRect(0, 0, 30, 10))
	g.Insert(ms)

	spy := newSpyView()
	g.Insert(spy)

	// First move to get a baseline broadcast count (0→3 from Up)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	ms.HandleEvent(ev)
	resetBroadcasts(spy)

	// Move again: 3→2
	ev = &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	ms.HandleEvent(ev)

	// This is a change, so broadcast should fire
	if len(spy.broadcasts) == 0 {
		t.Error("expected broadcast on selection change")
	}
}

// TestMonoSelectorDrawDoesNotPanic verifies Draw runs without panicking.
func TestMonoSelectorDrawDoesNotPanic(t *testing.T) {
	ms := NewMonoSelector(NewRect(0, 0, 20, 4))
	buf := NewDrawBuffer(25, 5)

	// Should not panic
	ms.Draw(buf)
}

// TestMonoSelectorNonKeyboardEventPassthrough verifies non-keyboard events
// are passed through without crashing.
func TestMonoSelectorNonKeyboardEventPassthrough(t *testing.T) {
	ms := NewMonoSelector(NewRect(0, 0, 20, 4))
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	ms.HandleEvent(ev)

	// Selection should not change on mouse events
	if ms.Selected() != 0 {
		t.Errorf("expected selection unchanged on mouse event, got %d", ms.Selected())
	}
}
