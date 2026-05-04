package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestNewColorSelector16 verifies a 16-color selector has Selected()==0 and Kind()==0.
func TestNewColorSelector16(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 4), 16)
	if cs.Selected() != 0 {
		t.Errorf("expected Selected() == 0, got %d", cs.Selected())
	}
	if cs.Kind() != 0 {
		t.Errorf("expected Kind() == 0 for 16 colors, got %d", cs.Kind())
	}
}

// TestNewColorSelector8 verifies an 8-color selector has Kind()==1.
func TestNewColorSelector8(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 2), 8)
	if cs.Kind() != 1 {
		t.Errorf("expected Kind() == 1 for 8 colors, got %d", cs.Kind())
	}
}

// TestColorSelectorArrowWrappingLeft verifies Left from column 0 wraps to column 3.
func TestColorSelectorArrowWrappingLeft(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 4), 16)
	// Selection starts at index 0 (row 0, col 0)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	cs.HandleEvent(ev)
	// Should wrap to column 3, index = 0*4 + 3 = 3
	if cs.Selected() != 3 {
		t.Errorf("expected Left from 0 to wrap to 3, got %d", cs.Selected())
	}
}

// TestColorSelectorArrowWrappingUp verifies Up from row 0 wraps to last row.
func TestColorSelectorArrowWrappingUp(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 4), 16)
	// Selection at index 0 (row 0)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	cs.HandleEvent(ev)
	// numRows = 16/4 = 4, wrapping: (0-1+4)%4 = 3, so row 3, col 0 = index 12
	if cs.Selected() != 12 {
		t.Errorf("expected Up from 0 to wrap to 12, got %d", cs.Selected())
	}
}

// TestColorSelectorArrowWrappingDown verifies Down from last row wraps to row 0.
func TestColorSelectorArrowWrappingDown(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 4), 16)
	cs.SetSelected(12) // row 3, col 0
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cs.HandleEvent(ev)
	// numRows = 4, wrapping: (3+1)%4 = 0, row 0, col 0 = index 0
	if cs.Selected() != 0 {
		t.Errorf("expected Down from 12 to wrap to 0, got %d", cs.Selected())
	}
}

// TestColorSelectorArrowWrappingRight verifies Right from column 3 wraps to column 0.
func TestColorSelectorArrowWrappingRight(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 4), 16)
	cs.SetSelected(3) // row 0, col 3
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	cs.HandleEvent(ev)
	// col wraps: (3+1)%4 = 0, row 0, col 0 = index 0
	if cs.Selected() != 0 {
		t.Errorf("expected Right from 3 to wrap to 0, got %d", cs.Selected())
	}
}

// TestColorSelectorBroadcastFires verifies a broadcast is sent when selection changes.
func TestColorSelectorBroadcastFires(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 4), 16)
	g := NewGroup(NewRect(0, 0, 20, 10))
	g.Insert(cs)

	spy := newSpyView()
	g.Insert(spy)
	resetBroadcasts(spy)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cs.HandleEvent(ev)

	found := false
	for _, rec := range spy.broadcasts {
		if rec.command == CmColorForegroundChanged {
			found = true
			if idx, ok := rec.info.(int); !ok || idx != 4 {
				t.Errorf("expected CmColorForegroundChanged Info == 4, got %v", rec.info)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected CmColorForegroundChanged broadcast, got: %v", spy.broadcasts)
	}
}

// TestColorSelectorBackgroundBroadcast verifies Kind()==1 (8 colors) sends CmColorBackgroundChanged.
func TestColorSelectorBackgroundBroadcast(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 2), 8)
	g := NewGroup(NewRect(0, 0, 20, 10))
	g.Insert(cs)

	spy := newSpyView()
	g.Insert(spy)
	resetBroadcasts(spy)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cs.HandleEvent(ev)

	found := false
	for _, rec := range spy.broadcasts {
		if rec.command == CmColorBackgroundChanged {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected CmColorBackgroundChanged broadcast for 8-color selector, got: %v", spy.broadcasts)
	}
}

// TestColorSelectorMouseClick verifies mouse click selects the correct cell.
func TestColorSelectorMouseClick(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 4), 16)
	// Need SfSelected to prevent BaseView from clearing the mouse event
	// (BaseView clears mouse events for unselected selectable views without OfFirstClick)
	cs.SetState(SfSelected, true)

	// Click at pixel position corresponding to col=2, row=1 (index 1*4+2 = 6)
	// Each cell is 3px wide, so cell at col=2 starts at x=6
	mx := 7 // center of cell col 2
	my := 1 // row 1
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: mx, Y: my, Button: tcell.Button1}}
	cs.HandleEvent(ev)

	if cs.Selected() != 6 {
		t.Errorf("expected mouse click at (%d,%d) to select index 6, got %d", mx, my, cs.Selected())
	}
}

// TestColorSelectorMouseClamp verifies out-of-bounds mouse click does not change selection.
func TestColorSelectorMouseClamp(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 4), 16)
	cs.SetSelected(5)
	cs.SetState(SfSelected, true)

	// Click at a position that maps to an out-of-range index
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 50, Y: 10, Button: tcell.Button1}}
	cs.HandleEvent(ev)

	// Selection should still be 5 — out of bounds clicks don't change it
	if cs.Selected() != 5 {
		t.Errorf("expected no change for out-of-bounds click, got %d", cs.Selected())
	}
}

// TestColorSelectorSetSelectedClamps verifies SetSelected clamps to valid range.
func TestColorSelectorSetSelectedClamps(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 4), 16)

	cs.SetSelected(-5)
	if cs.Selected() != 0 {
		t.Errorf("expected SetSelected(-5) to clamp to 0, got %d", cs.Selected())
	}

	cs.SetSelected(100)
	if cs.Selected() != 15 {
		t.Errorf("expected SetSelected(100) to clamp to 15, got %d", cs.Selected())
	}

	cs.SetSelected(7)
	if cs.Selected() != 7 {
		t.Errorf("expected SetSelected(7) to be 7, got %d", cs.Selected())
	}
}

// TestColorSelectorUnknownKeyDoesNothing verifies an unhandled key does not
// change selection or send broadcast.
func TestColorSelectorUnknownKeyDoesNothing(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 4), 16)
	g := NewGroup(NewRect(0, 0, 20, 10))
	g.Insert(cs)

	spy := newSpyView()
	g.Insert(spy)
	resetBroadcasts(spy)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab}}
	cs.HandleEvent(ev)

	if cs.Selected() != 0 {
		t.Errorf("expected no change on Tab, got %d", cs.Selected())
	}
	if len(spy.broadcasts) > 0 {
		t.Errorf("expected no broadcasts on Tab, got: %v", spy.broadcasts)
	}
}

// TestColorSelectorNoDuplicateBroadcast verifies that pressing a key
// that produces no selection change (Up on a single row grid) does not
// fire a broadcast.
func TestColorSelectorNoDuplicateBroadcast(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 1), 4)
	g := NewGroup(NewRect(0, 0, 20, 10))
	g.Insert(cs)

	spy := newSpyView()
	g.Insert(spy)
	resetBroadcasts(spy)

	// Up on a 1-row grid: row wraps to same row, col stays, selection unchanged
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	cs.HandleEvent(ev)

	if cs.Selected() != 0 {
		t.Errorf("expected selection unchanged, got %d", cs.Selected())
	}
	if len(spy.broadcasts) > 0 {
		t.Errorf("expected no broadcasts when selection doesn't change, got: %v", spy.broadcasts)
	}
}

// TestColorSelectorDrawDoesNotPanic verifies Draw runs without panicking
// even when no ColorScheme is available.
func TestColorSelectorDrawDoesNotPanic(t *testing.T) {
	cs := NewColorSelector(NewRect(0, 0, 12, 4), 16)
	buf := NewDrawBuffer(15, 5)

	// Should not panic
	cs.Draw(buf)
}
