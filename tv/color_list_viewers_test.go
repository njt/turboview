package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestNewColorGroupListViewer creates a ColorGroupListViewer backed by StringList
// and confirms it is non-nil with a valid selected index.
func TestNewColorGroupListViewer(t *testing.T) {
	ds := NewStringList([]string{"Desktop", "Menus", "Dialogs"})
	cgl := NewColorGroupListViewer(NewRect(0, 0, 20, 5), ds)

	if cgl == nil {
		t.Fatal("NewColorGroupListViewer returned nil")
	}
	if cgl.Selected() != 0 {
		t.Errorf("expected Selected() == 0, got %d", cgl.Selected())
	}
	if cgl.DataSource() != ds {
		t.Error("DataSource should be the one passed in")
	}
	if cgl.lastSelected != -1 {
		t.Errorf("expected lastSelected == -1, got %d", cgl.lastSelected)
	}
}

// TestNewColorItemListViewer creates a ColorItemListViewer backed by StringList
// and confirms it is non-nil with the correct initial state.
func TestNewColorItemListViewer(t *testing.T) {
	ds := NewStringList([]string{"Normal", "Selected", "Focused"})
	cil := NewColorItemListViewer(NewRect(0, 0, 20, 5), ds)

	if cil == nil {
		t.Fatal("NewColorItemListViewer returned nil")
	}
	if cil.Selected() != 0 {
		t.Errorf("expected Selected() == 0, got %d", cil.Selected())
	}
	if cil.lastSelected != -1 {
		t.Errorf("expected lastSelected == -1, got %d", cil.lastSelected)
	}
}

// TestColorGroupListViewerBroadcastFires verifies that pressing Down arrow
// causes the viewer to send a CmNewColorGroup broadcast through its owner.
func TestColorGroupListViewerBroadcastFires(t *testing.T) {
	ds := NewStringList([]string{"Desktop", "Menus", "Dialogs", "Buttons"})
	cgl := NewColorGroupListViewer(NewRect(0, 0, 20, 5), ds)
	g := NewGroup(NewRect(0, 0, 20, 10))
	g.Insert(cgl)

	spy := newSpyView()
	g.Insert(spy)

	// Ensure the viewer has focus (SfSelected) so ListViewer processes keyboard events
	g.SetFocusedChild(cgl)
	resetBroadcasts(spy)

	// Press Down to move selection from 0 to 1
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cgl.HandleEvent(ev)

	// Verify broadcast was received
	found := false
	for _, rec := range spy.broadcasts {
		if rec.command == CmNewColorGroup {
			found = true
			// Info should be the new selection index (1)
			if idx, ok := rec.info.(int); !ok || idx != 1 {
				t.Errorf("expected CmNewColorGroup Info == 1, got %v", rec.info)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected CmNewColorGroup broadcast, got broadcasts: %v", spy.broadcasts)
	}
}

// TestColorItemListViewerBroadcastFires verifies that pressing Down arrow
// causes the viewer to send a CmNewColorIndex broadcast through its owner.
func TestColorItemListViewerBroadcastFires(t *testing.T) {
	ds := NewStringList([]string{"Normal", "Selected", "Focused"})
	cil := NewColorItemListViewer(NewRect(0, 0, 20, 5), ds)
	g := NewGroup(NewRect(0, 0, 20, 10))
	g.Insert(cil)

	spy := newSpyView()
	g.Insert(spy)

	g.SetFocusedChild(cil)
	resetBroadcasts(spy)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cil.HandleEvent(ev)

	found := false
	for _, rec := range spy.broadcasts {
		if rec.command == CmNewColorIndex {
			found = true
			if idx, ok := rec.info.(int); !ok || idx != 1 {
				t.Errorf("expected CmNewColorIndex Info == 1, got %v", rec.info)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected CmNewColorIndex broadcast, got broadcasts: %v", spy.broadcasts)
	}
}

// TestColorGroupListViewerNoDuplicateBroadcast verifies that calling
// HandleEvent with the same key when selection doesn't change does not
// produce duplicate broadcasts.
func TestColorGroupListViewerNoDuplicateBroadcast(t *testing.T) {
	ds := NewStringList([]string{"Desktop", "Menus"})
	cgl := NewColorGroupListViewer(NewRect(0, 0, 20, 5), ds)
	g := NewGroup(NewRect(0, 0, 20, 10))
	g.Insert(cgl)

	spy := newSpyView()
	g.Insert(spy)
	g.SetFocusedChild(cgl)

	// First key press: selection 0→1, broadcast expected
	ev1 := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cgl.HandleEvent(ev1)
	resetBroadcasts(spy)

	// Second key press: selection stays at 1 since only 2 items
	// (Down on last item should not change selection)
	ev2 := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cgl.HandleEvent(ev2)

	count := 0
	for _, rec := range spy.broadcasts {
		if rec.command == CmNewColorGroup {
			count++
		}
	}
	if count > 0 {
		t.Errorf("expected 0 CmNewColorGroup broadcasts on duplicate selection, got %d", count)
	}
}

// TestColorListViewerDelegatesNonKeyboard verifies that HandleEvent passes
// non-keyboard events through to the underlying ListViewer.
func TestColorListViewerDelegatesNonKeyboard(t *testing.T) {
	ds := NewStringList([]string{"Desktop", "Menus"})
	cgl := NewColorGroupListViewer(NewRect(0, 0, 20, 5), ds)

	// Mouse event should be handled normally by ListViewer
	// and should not cause a panic or error
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	cgl.HandleEvent(ev)

	// Verify selection hasn't changed — mouse click routing
	// depends on BaseView click-to-focus which requires the view to be in a container
	if cgl.Selected() < 0 {
		t.Error("selected should not be negative")
	}
}

// TestColorItemListViewerDoesNotBroadcastWhenNoOwner verifies that if the
// viewer has no owner, no broadcast is sent (no panic).
func TestColorItemListViewerNoOwner(t *testing.T) {
	ds := NewStringList([]string{"A", "B", "C"})
	cil := NewColorItemListViewer(NewRect(0, 0, 20, 5), ds)

	// Set SfSelected so ListViewer processes keyboard events
	cil.SetState(SfSelected, true)

	// No owner set — broadcast should be silently skipped
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	// This should not panic
	cil.HandleEvent(ev)

	if cil.Selected() != 1 {
		t.Errorf("expected selection to move to 1, got %d", cil.Selected())
	}
}
