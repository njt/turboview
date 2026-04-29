package tv

import "testing"

// mouseSpyView is a test double that records whether HandleEvent was called and
// what mouse coordinates were delivered (in the coordinate space the event
// carried at call time).
type mouseSpyView struct {
	BaseView
	called     bool
	lastEvent  *Event
	lastMouseX int
	lastMouseY int
}

func (s *mouseSpyView) HandleEvent(event *Event) {
	s.called = true
	s.lastEvent = event
	if event.Mouse != nil {
		s.lastMouseX = event.Mouse.X
		s.lastMouseY = event.Mouse.Y
	}
}

// newMouseSpy creates a visible mouseSpyView positioned at the given bounds.
func newMouseSpy(bounds Rect) *mouseSpyView {
	v := &mouseSpyView{}
	v.SetBounds(bounds)
	v.SetState(SfVisible, true)
	return v
}

// newSelectableMouseSpy creates a visible, selectable mouseSpyView.
func newSelectableMouseSpy(bounds Rect) *mouseSpyView {
	v := newMouseSpy(bounds)
	v.SetOptions(OfSelectable, true)
	return v
}

// mouseEventAt builds a mouse event with the given absolute coordinates.
func mouseEventAt(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y}}
}

// ── Test 1: Mouse goes to child containing the point, NOT to focused child ────

// TestMouseRoutedToContainingChildNotFocused verifies that a mouse event is
// delivered to the child whose bounds contain the mouse point, regardless of
// which child is focused.
// Spec: "dispatch mouse events to the child whose bounds contain the mouse
// point … It must NOT send mouse events to the focused child."
func TestMouseRoutedToContainingChildNotFocused(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// focused is selectable and gets focus via Insert; it is placed at x=40
	focused := newSelectableMouseSpy(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	// target is not selectable, placed at x=0; we click inside it
	target := newMouseSpy(NewRect(0, 0, 20, 10))
	g.Insert(target)

	g.HandleEvent(mouseEventAt(5, 5))

	if !target.called {
		t.Errorf("mouse event was not delivered to the child whose bounds contain the point")
	}
	if focused.called {
		t.Errorf("mouse event was delivered to the focused child (positional routing must bypass focused dispatch)")
	}
}

// TestMouseRoutingShortcutFalsified confirms the test above would catch an
// implementation that always sends mouse events to the focused child instead
// of doing positional lookup. The focused child's bounds do NOT contain the
// click point, so it must not be called.
// Spec: "It must NOT send mouse events to the focused child."
func TestMouseRoutingShortcutFalsified(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newSelectableMouseSpy(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	other := newMouseSpy(NewRect(0, 0, 20, 10))
	g.Insert(other)

	// Click inside other, not inside focused.
	g.HandleEvent(mouseEventAt(5, 5))

	if focused.called {
		t.Errorf("shortcut check: implementation sent mouse event to focused child instead of performing positional routing")
	}
}

// ── Test 2: Coordinates translated to child-local space ───────────────────────

// TestMouseCoordinatesTranslatedToChildLocalSpace verifies that the X/Y
// coordinates in the delivered event are translated to child-local space
// (i.e. relative to the child's top-left corner).
// Spec: "Translate the mouse coordinates to child-local space."
func TestMouseCoordinatesTranslatedToChildLocalSpace(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// Child starts at (10, 5) in group space; it occupies columns 10–29, rows 5–14.
	child := newMouseSpy(NewRect(10, 5, 20, 10))
	g.Insert(child)

	// Click at (15, 8) in group space → local (5, 3).
	g.HandleEvent(mouseEventAt(15, 8))

	if !child.called {
		t.Fatalf("child was not called; cannot verify coordinate translation")
	}
	wantX, wantY := 5, 3
	if child.lastMouseX != wantX || child.lastMouseY != wantY {
		t.Errorf("translated coordinates = (%d, %d), want (%d, %d)",
			child.lastMouseX, child.lastMouseY, wantX, wantY)
	}
}

// TestMouseCoordinatesTranslationShortcutFalsified confirms that passing
// un-translated (absolute) coordinates would fail this test.
// Spec: "Translate the mouse coordinates to child-local space."
func TestMouseCoordinatesTranslationShortcutFalsified(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// Child at (10, 5); click at (15, 8) → local (5, 3).
	child := newMouseSpy(NewRect(10, 5, 20, 10))
	g.Insert(child)

	g.HandleEvent(mouseEventAt(15, 8))

	if !child.called {
		t.Fatalf("child was not called; cannot verify translation shortcut check")
	}
	// If the implementation forgot to translate, the child would see (15, 8).
	// We assert it does NOT see the absolute coords.
	if child.lastMouseX == 15 && child.lastMouseY == 8 {
		t.Errorf("child received absolute coordinates (15, 8) instead of translated child-local coordinates (5, 3)")
	}
}

// ── Test 3: Overlapping children — last (topmost) child wins ─────────────────

// TestMouseOverlappingChildrenTopmostReceivesEvent verifies that when two
// children overlap, the last child in insertion order (topmost in z-order)
// receives the event.
// Spec: "Iterate children in reverse order (front-to-back z-order)" and
// "Find the first visible child whose bounds contain the mouse point."
func TestMouseOverlappingChildrenTopmostReceivesEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// Both children fully overlap at (0,0)–(20,10).
	bottom := newMouseSpy(NewRect(0, 0, 20, 10))
	top := newMouseSpy(NewRect(0, 0, 20, 10))

	g.Insert(bottom) // inserted first = lower z-order
	g.Insert(top)    // inserted last  = higher z-order (topmost)

	g.HandleEvent(mouseEventAt(5, 5))

	if !top.called {
		t.Errorf("topmost (last-inserted) overlapping child did not receive mouse event")
	}
	if bottom.called {
		t.Errorf("bottom (first-inserted) overlapping child received mouse event; topmost should have taken it")
	}
}

// TestMouseOverlappingZOrderShortcutFalsified confirms that a naive
// front-to-back (forward) iteration would get the wrong child.
// Spec: "searching back-to-front in children (last child = topmost)."
func TestMouseOverlappingZOrderShortcutFalsified(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	bottom := newMouseSpy(NewRect(0, 0, 20, 10))
	top := newMouseSpy(NewRect(0, 0, 20, 10))

	g.Insert(bottom)
	g.Insert(top)

	g.HandleEvent(mouseEventAt(5, 5))

	// A wrong implementation that iterates forward would call bottom first.
	if bottom.called {
		t.Errorf("z-order shortcut: forward iteration chose bottom child instead of topmost (last-inserted)")
	}
}

// ── Test 4: Invisible child skipped even if it contains the point ─────────────

// TestMouseInvisibleChildSkipped verifies that a child without SfVisible is
// skipped during positional routing even if its bounds contain the mouse point.
// Spec: "Find the first visible child whose bounds contain the mouse point."
func TestMouseInvisibleChildSkipped(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	invisible := newMouseSpy(NewRect(0, 0, 20, 10))
	invisible.SetState(SfVisible, false) // explicitly invisible

	g.Insert(invisible)

	g.HandleEvent(mouseEventAt(5, 5))

	if invisible.called {
		t.Errorf("invisible child received a mouse event; invisible children must be skipped")
	}
}

// TestMouseInvisibleChildSkippedFallsThrough verifies that when the topmost
// child containing the point is invisible, the next visible child underneath
// receives the event instead.
// Spec: "Find the first visible child whose bounds contain the mouse point."
func TestMouseInvisibleChildSkippedFallsThrough(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	visible := newMouseSpy(NewRect(0, 0, 20, 10))

	invisible := newMouseSpy(NewRect(0, 0, 20, 10))
	invisible.SetState(SfVisible, false)

	g.Insert(visible)    // lower z-order but visible
	g.Insert(invisible)  // higher z-order but invisible → must be skipped

	g.HandleEvent(mouseEventAt(5, 5))

	if !visible.called {
		t.Errorf("visible child beneath invisible one did not receive mouse event after invisible child was skipped")
	}
	if invisible.called {
		t.Errorf("invisible child received mouse event despite not being visible")
	}
}

// ── Test 5: No child contains point → no child receives event ─────────────────

// TestMouseNoChildContainsPointNoDelivery verifies that when no child's bounds
// contain the mouse point, no child receives the event.
// Spec: "If no child contains the mouse point, the event is not delivered to
// any child."
func TestMouseNoChildContainsPointNoDelivery(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// Child is in the top-left; click is far outside it.
	child := newMouseSpy(NewRect(0, 0, 10, 5))
	g.Insert(child)

	g.HandleEvent(mouseEventAt(50, 20)) // well outside child bounds

	if child.called {
		t.Errorf("child received mouse event even though the click point was outside its bounds")
	}
}

// TestMouseEmptyGroupNoDelivery verifies that a mouse event on an empty group
// does not panic and delivers to nobody.
// Spec: "If no child contains the mouse point, the event is not delivered to
// any child." — the empty-group case is a degenerate but valid instance.
func TestMouseEmptyGroupNoDelivery(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// Must not panic.
	g.HandleEvent(mouseEventAt(5, 5))
}

// ── Test 6: Non-mouse events follow normal (focused) dispatch ─────────────────

// TestKeyboardEventStillGoesToFocusedChild verifies that keyboard events are
// not subject to positional routing and still reach the focused child via the
// normal dispatch path, regardless of where the focused child is positioned.
// Spec: (implicit) positional routing is mouse-only; non-mouse events use the
// existing focused/three-phase dispatch.
func TestKeyboardEventStillGoesToFocusedChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// focused is selectable and receives focus on Insert; it is placed far from
	// where a hypothetical mouse click would land.
	focused := newSelectableMouseSpy(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	// A second visible child is inserted but not focused.
	other := newMouseSpy(NewRect(0, 0, 20, 10))
	g.Insert(other)

	g.HandleEvent(&Event{What: EvKeyboard})

	if !focused.called {
		t.Errorf("keyboard event did not reach the focused child; non-mouse events must use normal dispatch")
	}
}

// TestKeyboardEventDoesNotUsePositionalRouting verifies that a keyboard event
// is not positionally dispatched to a non-focused child.
// Spec: positional routing applies to mouse events only.
func TestKeyboardEventDoesNotUsePositionalRouting(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newSelectableMouseSpy(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	// other is at x=0; a positional router for keyboard events would wrongly
	// pick this child.
	other := newMouseSpy(NewRect(0, 0, 20, 10))
	g.Insert(other)

	g.HandleEvent(&Event{What: EvKeyboard})

	if other.called {
		t.Errorf("keyboard event was positionally routed to a non-focused child; positional routing is mouse-only")
	}
}
