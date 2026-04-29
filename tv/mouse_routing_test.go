package tv

// mouse_routing_test.go — tests for Task 4: Mouse Event Routing.
// Written against the spec BEFORE any implementation exists.
// Each test cites the spec requirement it verifies.
//
// Spec summary:
//   Application.handleEvent routes mouse events by position (StatusLine bounds
//   first, then Desktop bounds), translating coordinates to the target's local
//   space. Non-mouse events use the existing path unchanged.
//
//   Desktop.HandleEvent for mouse events: iterates windows front-to-back (last
//   child first), hit-tests using child.Bounds().Contains(point), translates
//   coordinates to window-local space, and forwards. If the hit window has
//   OfTopSelect and a button is pressed (Button1), brings it to front first.
//
//   Window.HandleEvent for mouse events: if the click is in the client area
//   (x>0 && x<width-1 && y>0 && y<height-1) translates to client-local coords
//   and forwards to the internal Group. Otherwise the click is on the frame
//   (Task 5 — currently no-op).

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// mouseEvent constructs an EvMouse Event with absolute screen coordinates.
func mouseEvent(x, y int, button tcell.ButtonMask) *Event {
	return &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:      x,
			Y:      y,
			Button: button,
		},
	}
}

// keyboardEvent constructs a simple EvKeyboard Event.
func keyboardEvent() *Event {
	return &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'a'},
	}
}

// ---------------------------------------------------------------------------
// Desktop.HandleEvent — mouse routing
// ---------------------------------------------------------------------------

// TestDesktopMouseRoutingClickOnFrontmostWindow verifies that a mouse click
// at a position covered by the frontmost (last-inserted) window is forwarded
// to that window with coordinates translated to window-local space.
//
// Spec: "Desktop.HandleEvent for mouse events: iterates windows front-to-back
// (last child first), hit-tests using child.Bounds().Contains(point),
// translates coordinates to window-local space, and forwards."
func TestDesktopMouseRoutingClickOnFrontmostWindow(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))

	// Window occupies cols 10-29, rows 2-11 (20×10).
	win := NewWindow(NewRect(10, 2, 20, 10), "Front")
	child := newSelectablePhaseView("child", NewRect(1, 1, 18, 8))
	win.Insert(child)
	d.Insert(win)

	// Click at desktop-absolute (15, 5) → window-local (5, 3).
	ev := mouseEvent(15, 5, tcell.Button1)
	d.HandleEvent(ev)

	// The window should have received the event (it will route internally).
	// We check via the child inside the window — if Window.HandleEvent forwarded
	// to the Group and the phaseView was focused, it receives the event.
	// For Desktop routing correctness we just need to know the event reached the
	// window's boundary. We track this via a phaseTestView inserted *directly*
	// as a Desktop child that records receipt.
	// Reset: use a simpler approach — insert a phaseTestView directly as the
	// Desktop child so we can check it directly.
	d2 := NewDesktop(NewRect(0, 0, 80, 24))
	pv := newSelectablePhaseView("pv", NewRect(10, 2, 20, 10))
	d2.Insert(pv)

	ev2 := mouseEvent(15, 5, tcell.Button1)
	d2.HandleEvent(ev2)

	if pv.handleCount == 0 {
		t.Error("Desktop: frontmost window/view did not receive mouse click inside its bounds")
	}
	if pv.lastEvent == nil {
		t.Fatal("Desktop: lastEvent is nil")
	}
	// Coordinates must be translated to local space: 15-10=5, 5-2=3.
	if pv.lastEvent.Mouse.X != 5 || pv.lastEvent.Mouse.Y != 3 {
		t.Errorf("Desktop: translated coords = (%d,%d), want (5,3)",
			pv.lastEvent.Mouse.X, pv.lastEvent.Mouse.Y)
	}
}

// TestDesktopMouseRoutingOverlapGoesToFrontmost verifies that when two children
// overlap at the click position, the event goes to the frontmost (last in
// Children() list), not the one behind it.
//
// Spec: "iterates windows front-to-back (last child first), hit-tests using
// child.Bounds().Contains(point)"
func TestDesktopMouseRoutingOverlapGoesToFrontmost(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))

	// Both views cover (5,5). back is first-inserted; front is last-inserted.
	back := newSelectablePhaseView("back", NewRect(0, 0, 20, 10))
	front := newPhaseView("front", NewRect(0, 0, 20, 10))
	front.SetState(SfVisible, true)
	d.Insert(back)
	d.Insert(front)

	ev := mouseEvent(5, 5, tcell.Button1)
	d.HandleEvent(ev)

	if front.handleCount == 0 {
		t.Error("Desktop: frontmost child did not receive mouse click at overlapping position")
	}
	if back.handleCount != 0 {
		t.Error("Desktop: back child should NOT receive mouse click when front child is on top")
	}
}

// TestDesktopMouseRoutingClickOnBackWindow verifies that a click at a position
// covered only by a back window (not overlapped by any frontmost child) is
// forwarded to that back window.
//
// Spec: "iterates windows front-to-back (last child first), hit-tests using
// child.Bounds().Contains(point)"
func TestDesktopMouseRoutingClickOnBackWindow(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))

	// back covers cols 0-19; front covers cols 40-59; click at (5,5) hits only back.
	back := newPhaseView("back", NewRect(0, 0, 20, 10))
	back.SetState(SfVisible, true)
	front := newPhaseView("front", NewRect(40, 0, 20, 10))
	front.SetState(SfVisible, true)
	d.Insert(back)
	d.Insert(front)

	ev := mouseEvent(5, 5, tcell.Button1)
	d.HandleEvent(ev)

	if back.handleCount == 0 {
		t.Error("Desktop: back window did not receive click at its exclusive area")
	}
	if front.handleCount != 0 {
		t.Error("Desktop: front window should NOT receive click that hits only the back window")
	}
}

// TestDesktopMouseRoutingClickOutsideAllWindowsNotForwarded verifies that a
// mouse click outside all children's bounds is not forwarded to any child.
//
// Spec: "hit-tests using child.Bounds().Contains(point)" — if no child
// contains the point, no child receives the event.
func TestDesktopMouseRoutingClickOutsideAllWindowsNotForwarded(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))

	pv := newPhaseView("pv", NewRect(10, 10, 20, 10))
	pv.SetState(SfVisible, true)
	d.Insert(pv)

	// Click far away from the window.
	ev := mouseEvent(5, 5, tcell.Button1)
	d.HandleEvent(ev)

	if pv.handleCount != 0 {
		t.Errorf("Desktop: child received mouse click at (%d,%d) which is outside its bounds",
			ev.Mouse.X, ev.Mouse.Y)
	}
}

// TestDesktopMouseRoutingBringsOfTopSelectToFront verifies that clicking a
// window with OfTopSelect while Button1 is pressed brings it to front before
// forwarding the event.
//
// Spec: "If the hit window has OfTopSelect and a button is pressed (Button1),
// brings it to front first."
func TestDesktopMouseRoutingBringsOfTopSelectToFront(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))

	// first has OfTopSelect (default for NewWindow), second is on top.
	// Both share position (5,5).
	first := NewWindow(NewRect(0, 0, 20, 10), "First")
	second := NewWindow(NewRect(0, 0, 20, 10), "Second")
	d.Insert(first)
	d.Insert(second)
	// second is now frontmost (last).

	// Click at (5,5) with Button1 — hits second (frontmost), which already is front.
	// To test BringToFront behaviour, use non-overlapping windows and click back one.
	d2 := NewDesktop(NewRect(0, 0, 80, 24))
	backWin := NewWindow(NewRect(0, 0, 30, 15), "Back")
	frontWin := NewWindow(NewRect(40, 0, 30, 15), "Front")
	d2.Insert(backWin)
	d2.Insert(frontWin)
	// backWin is at index 0, frontWin at index 1.

	// Click at (5, 5) — inside backWin, outside frontWin.
	ev := mouseEvent(5, 5, tcell.Button1)
	d2.HandleEvent(ev)

	// After the click with Button1, backWin (which has OfTopSelect) should be
	// brought to front — it should now be last in Children().
	children := d2.Children()
	if len(children) < 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[len(children)-1] != backWin {
		t.Error("Desktop: clicking OfTopSelect window with Button1 should bring it to front")
	}
}

// TestDesktopMouseRoutingButton1NotPressedDoesNotBringToFront verifies that a
// mouse-move (ButtonNone) on a window with OfTopSelect does NOT bring it to front.
//
// Spec: "If the hit window has OfTopSelect and a button is pressed (Button1),
// brings it to front first." — ButtonNone does not count as a press.
func TestDesktopMouseRoutingButton1NotPressedDoesNotBringToFront(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))

	backWin := NewWindow(NewRect(0, 0, 30, 15), "Back")
	frontWin := NewWindow(NewRect(40, 0, 30, 15), "Front")
	d.Insert(backWin)
	d.Insert(frontWin)
	// backWin is at index 0, frontWin at index 1.

	// Mouse-move (no button) over backWin.
	ev := mouseEvent(5, 5, tcell.ButtonNone)
	d.HandleEvent(ev)

	// frontWin should still be last (frontmost).
	children := d.Children()
	if len(children) < 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[len(children)-1] != frontWin {
		t.Error("Desktop: mouse-move (ButtonNone) on OfTopSelect window should NOT bring it to front")
	}
}

// TestDesktopMouseRoutingCoordinateTranslation verifies that the event
// coordinates forwarded to the hit child are translated to child-local space
// (subtract the child's origin).
//
// Spec: "translates coordinates to window-local space" (subtract bounds.A.X,
// bounds.A.Y from the event coordinates).
func TestDesktopMouseRoutingCoordinateTranslation(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))

	// Child at origin (15, 7).
	pv := newPhaseView("pv", NewRect(15, 7, 20, 10))
	pv.SetState(SfVisible, true)
	d.Insert(pv)

	// Click at absolute (20, 10) → local (20-15=5, 10-7=3).
	ev := mouseEvent(20, 10, tcell.Button1)
	d.HandleEvent(ev)

	if pv.handleCount == 0 {
		t.Fatal("Desktop: child did not receive the mouse event")
	}
	if pv.lastEvent.Mouse.X != 5 || pv.lastEvent.Mouse.Y != 3 {
		t.Errorf("Desktop: translated coords = (%d,%d), want (5,3)",
			pv.lastEvent.Mouse.X, pv.lastEvent.Mouse.Y)
	}
}

// TestDesktopMouseRoutingInvisibleWindowSkipped verifies that an invisible
// child is not hit-tested during mouse routing.
//
// Spec: "iterates windows front-to-back … hit-tests" — invisible children
// should be skipped (consistent with Draw skipping invisible children).
func TestDesktopMouseRoutingInvisibleWindowSkipped(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))

	// Invisible child covering the click area.
	invisible := &phaseTestView{name: "invisible"}
	invisible.SetBounds(NewRect(0, 0, 20, 10))
	// SfVisible intentionally NOT set
	d.Insert(invisible)

	// Visible child at different position.
	visible := newPhaseView("visible", NewRect(30, 0, 20, 10))
	visible.SetState(SfVisible, true)
	d.Insert(visible)

	// Click inside invisible child's bounds.
	ev := mouseEvent(5, 5, tcell.Button1)
	d.HandleEvent(ev)

	if invisible.handleCount != 0 {
		t.Error("Desktop: invisible child should be skipped during hit-testing")
	}
}

// ---------------------------------------------------------------------------
// Window.HandleEvent — mouse routing
// ---------------------------------------------------------------------------

// TestWindowMouseRoutingClientAreaForwardsToGroup verifies that a mouse click
// inside the client area (x>0, x<width-1, y>0, y<height-1) is forwarded to
// the internal Group with client-local coordinates (x-1, y-1).
//
// Spec: "Window.HandleEvent for mouse events: if the click is in the client
// area (x>0 && x<width-1 && y>0 && y<height-1), translates to client-local
// coordinates and forwards to the internal Group."
func TestWindowMouseRoutingClientAreaForwardsToGroup(t *testing.T) {
	// 20×10 window; client area is cols 1-18, rows 1-8.
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	child := newSelectablePhaseView("child", NewRect(0, 0, 18, 8))
	w.Insert(child)

	// Window-local click at (5, 4) — inside client area (x>0, x<19, y>0, y<9).
	// Client-local coords: (5-1=4, 4-1=3).
	ev := mouseEvent(5, 4, tcell.Button1)
	w.HandleEvent(ev)

	if child.handleCount == 0 {
		t.Error("Window: client-area click was not forwarded to internal Group")
	}
	if child.lastEvent == nil {
		t.Fatal("Window: child lastEvent is nil")
	}
	if child.lastEvent.Mouse.X != 4 || child.lastEvent.Mouse.Y != 3 {
		t.Errorf("Window: client-local coords = (%d,%d), want (4,3)",
			child.lastEvent.Mouse.X, child.lastEvent.Mouse.Y)
	}
}

// TestWindowMouseRoutingFrameClickNotForwardedToGroup verifies that a mouse
// click on the frame (top border, y==0) is NOT forwarded to the internal Group.
//
// Spec: "Otherwise the click is on the frame (handled in Task 5 — currently
// a no-op)."
func TestWindowMouseRoutingFrameClickNotForwardedToGroup(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	child := newSelectablePhaseView("child", NewRect(0, 0, 18, 8))
	w.Insert(child)

	// Click on top border (y==0).
	ev := mouseEvent(5, 0, tcell.Button1)
	w.HandleEvent(ev)

	if child.handleCount != 0 {
		t.Error("Window: frame click (y==0) should NOT be forwarded to the Group")
	}
}

// TestWindowMouseRoutingBottomBorderNotForwarded verifies a click on the
// bottom border (y==height-1) is not forwarded to the Group.
//
// Spec: "Otherwise the click is on the frame (handled in Task 5 — currently
// a no-op)."
func TestWindowMouseRoutingBottomBorderNotForwarded(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	child := newSelectablePhaseView("child", NewRect(0, 0, 18, 8))
	w.Insert(child)

	// Bottom border: y == height-1 == 9.
	ev := mouseEvent(5, 9, tcell.Button1)
	w.HandleEvent(ev)

	if child.handleCount != 0 {
		t.Error("Window: frame click (y==height-1) should NOT be forwarded to the Group")
	}
}

// TestWindowMouseRoutingLeftBorderNotForwarded verifies a click on the left
// border (x==0) is not forwarded to the Group.
//
// Spec: "Otherwise the click is on the frame (handled in Task 5 — currently
// a no-op)."
func TestWindowMouseRoutingLeftBorderNotForwarded(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	child := newSelectablePhaseView("child", NewRect(0, 0, 18, 8))
	w.Insert(child)

	// Left border: x == 0.
	ev := mouseEvent(0, 5, tcell.Button1)
	w.HandleEvent(ev)

	if child.handleCount != 0 {
		t.Error("Window: frame click (x==0) should NOT be forwarded to the Group")
	}
}

// TestWindowMouseRoutingRightBorderNotForwarded verifies a click on the right
// border (x==width-1) is not forwarded to the Group.
//
// Spec: "Otherwise the click is on the frame (handled in Task 5 — currently
// a no-op)."
func TestWindowMouseRoutingRightBorderNotForwarded(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	child := newSelectablePhaseView("child", NewRect(0, 0, 18, 8))
	w.Insert(child)

	// Right border: x == width-1 == 19.
	ev := mouseEvent(19, 5, tcell.Button1)
	w.HandleEvent(ev)

	if child.handleCount != 0 {
		t.Error("Window: frame click (x==width-1) should NOT be forwarded to the Group")
	}
}

// TestWindowMouseRoutingNonMouseEventGoesToGroup verifies that keyboard and
// command events still use the existing path (group.HandleEvent three-phase
// dispatch) and are not affected by mouse routing logic.
//
// Spec: "All existing keyboard-event tests pass unchanged — mouse routing only
// activates for EvMouse events."
func TestWindowMouseRoutingNonMouseEventGoesToGroup(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	child := newSelectablePhaseView("child", NewRect(0, 0, 18, 8))
	w.Insert(child)

	ev := keyboardEvent()
	w.HandleEvent(ev)

	if child.handleCount == 0 {
		t.Error("Window: keyboard event should still reach the focused child via Group dispatch")
	}
	if child.lastEvent != ev {
		t.Error("Window: keyboard event should be the exact event forwarded to the child")
	}
}

// TestWindowMouseRoutingClientAreaExactBoundaryMin verifies that x==1 and
// y==1 (the very first client-area cells) ARE forwarded to the Group.
//
// Spec: "x>0 && x<width-1 && y>0 && y<height-1"
func TestWindowMouseRoutingClientAreaExactBoundaryMin(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	child := newSelectablePhaseView("child", NewRect(0, 0, 18, 8))
	w.Insert(child)

	// x==1, y==1: just inside the client area.
	ev := mouseEvent(1, 1, tcell.Button1)
	w.HandleEvent(ev)

	if child.handleCount == 0 {
		t.Error("Window: click at (1,1) should be in client area and forwarded to Group")
	}
	// Client-local: (1-1=0, 1-1=0).
	if child.lastEvent.Mouse.X != 0 || child.lastEvent.Mouse.Y != 0 {
		t.Errorf("Window: client-local coords at boundary = (%d,%d), want (0,0)",
			child.lastEvent.Mouse.X, child.lastEvent.Mouse.Y)
	}
}

// TestWindowMouseRoutingClientAreaExactBoundaryMax verifies that
// x==width-2 and y==height-2 (the very last client-area cells) ARE forwarded.
//
// Spec: "x>0 && x<width-1 && y>0 && y<height-1"
func TestWindowMouseRoutingClientAreaExactBoundaryMax(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	child := newSelectablePhaseView("child", NewRect(0, 0, 18, 8))
	w.Insert(child)

	// x==18 (width-2), y==8 (height-2): last client-area cells.
	ev := mouseEvent(18, 8, tcell.Button1)
	w.HandleEvent(ev)

	if child.handleCount == 0 {
		t.Error("Window: click at (width-2, height-2) should be in client area and forwarded")
	}
}

// ---------------------------------------------------------------------------
// Application.handleEvent — mouse routing
// ---------------------------------------------------------------------------

// TestApplicationMouseRoutingDesktopAreaToDesktop verifies that a mouse click
// in the Desktop area (rows 0..h-2 in an 80×25 screen with StatusLine) is
// routed to the Desktop with coordinates translated to Desktop-local space.
//
// Spec: "Application.handleEvent routes mouse events by position: checks
// StatusLine bounds first, then Desktop bounds, translating coordinates to
// the target's local space."
//
// Desktop bounds are (0,0,80,24) when there's a StatusLine at row 24.
// Coordinate translation: subtract Desktop.Bounds().A (which is (0,0)) —
// so desktop-local coords equal the screen coords for clicks in the desktop area.
func TestApplicationMouseRoutingDesktopAreaToDesktop(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	sl := NewStatusLine(NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit))
	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication error: %v", err)
	}

	// Insert a phaseTestView into the Desktop so we can detect receipt.
	// Place it to cover the click target.
	pv := newPhaseView("pv", NewRect(5, 5, 20, 10))
	pv.SetState(SfVisible, true)
	app.Desktop().Insert(pv)

	// Click at (10, 7) — in the Desktop area (y < 24).
	ev := mouseEvent(10, 7, tcell.Button1)
	app.handleEvent(ev)

	if pv.handleCount == 0 {
		t.Error("Application: mouse click in Desktop area was not routed to Desktop")
	}
}

// TestApplicationMouseRoutingStatusLineAreaToStatusLine verifies that a mouse
// click in the StatusLine's row (row 24 in an 80×25 screen) is routed to the
// StatusLine, not the Desktop.
//
// Spec: "checks StatusLine bounds first, then Desktop bounds"
// StatusLine is at (0,24,80,1) in an 80×25 screen.
func TestApplicationMouseRoutingStatusLineAreaToStatusLine(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	// We use a custom StatusLine implementation via a phaseTestView embedded in
	// the application's event routing. However, StatusLine is a concrete type.
	// Instead we verify indirectly: a click on row 24 should NOT reach the Desktop.

	sl := NewStatusLine(NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit))
	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication error: %v", err)
	}

	// Put a phase view in the Desktop that covers row 24 in desktop-absolute coords.
	// (Desktop height is only 24 rows, so row 24 is OUTSIDE the Desktop — but we
	// want to verify the Desktop never sees the event.)
	pv := newPhaseView("pv", NewRect(0, 0, 80, 24))
	pv.SetState(SfVisible, true)
	app.Desktop().Insert(pv)

	// Click at row 24 (the StatusLine's row).
	ev := mouseEvent(10, 24, tcell.Button1)
	app.handleEvent(ev)

	// The Desktop should NOT have received a click for row 24 (it's outside Desktop bounds).
	// If mouse routing is correct, the event goes to the StatusLine instead.
	// The phaseView at (0,0,80,24) in Desktop-local space does NOT contain (10,24)
	// because Desktop local space is rows 0..23. So handleCount should be 0.
	if pv.handleCount != 0 {
		t.Error("Application: mouse click on StatusLine row should NOT reach Desktop children")
	}
}

// TestApplicationMouseRoutingKeyboardEventUsesExistingPath verifies that
// keyboard events still use the existing StatusLine→Desktop path unchanged,
// with no mouse-routing logic applied.
//
// Spec: "Non-mouse events use the existing path (StatusLine → Desktop →
// CmQuit check)."
func TestApplicationMouseRoutingKeyboardEventUsesExistingPath(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	// Bind F10→CmQuit via the status line.
	sl := NewStatusLine(NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit))
	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication error: %v", err)
	}

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyF10},
	}
	app.handleEvent(ev)

	// The StatusLine should have transformed the F10 key to CmQuit, and then
	// handleEvent's CmQuit check should have set app.quit = true.
	if !app.quit {
		t.Error("Application: F10 keyboard event should still trigger CmQuit via existing path")
	}
}

// TestApplicationMouseRoutingCommandEventUsesExistingPath verifies that
// command events still use the existing path (not the mouse-routing path).
//
// Spec: "Non-mouse events use the existing path (StatusLine → Desktop →
// CmQuit check)."
func TestApplicationMouseRoutingCommandEventUsesExistingPath(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication error: %v", err)
	}

	ev := &Event{What: EvCommand, Command: CmQuit}
	app.handleEvent(ev)

	if !app.quit {
		t.Error("Application: CmQuit command event should still set app.quit via existing path")
	}
}

// TestApplicationMouseRoutingNoStatusLineDesktopGetsFullHeight verifies that
// when there is no StatusLine, the Desktop spans the full screen height and
// a click on the bottom row is routed to the Desktop.
//
// Spec: "Desktop bounds are (0,0,80,24) when there's a StatusLine at row 24
// in an 80×25 screen." — without a StatusLine, Desktop is (0,0,80,25).
func TestApplicationMouseRoutingNoStatusLineDesktopGetsFullHeight(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication error: %v", err)
	}

	pv := newPhaseView("pv", NewRect(0, 0, 80, 25))
	pv.SetState(SfVisible, true)
	app.Desktop().Insert(pv)

	// Click on the last row (24) — should reach Desktop since there's no StatusLine.
	ev := mouseEvent(10, 24, tcell.Button1)
	app.handleEvent(ev)

	if pv.handleCount == 0 {
		t.Error("Application: without StatusLine, bottom row click should reach Desktop")
	}
}

// ---------------------------------------------------------------------------
// Backward compatibility — existing keyboard dispatch unchanged
// ---------------------------------------------------------------------------

// TestMouseRoutingBackwardCompatKeyboardToDesktopFocusedChild verifies that
// keyboard events continue to reach the Desktop's focused child via the
// existing three-phase group dispatch — unaffected by Task 4 changes.
//
// Spec: "All existing keyboard-event tests pass unchanged — mouse routing only
// activates for EvMouse events."
func TestMouseRoutingBackwardCompatKeyboardToDesktopFocusedChild(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	focused := newSelectablePhaseView("focused", NewRect(0, 0, 40, 12))
	d.Insert(focused)

	ev := keyboardEvent()
	d.HandleEvent(ev)

	if focused.handleCount == 0 {
		t.Error("backward compat: keyboard event should still reach Desktop's focused child")
	}
	if focused.lastEvent != ev {
		t.Error("backward compat: Desktop should forward the exact keyboard event to focused child")
	}
}

// TestMouseRoutingBackwardCompatCommandToDesktopFocusedChild verifies that
// command events still reach the Desktop's focused child via existing dispatch.
//
// Spec: "All existing keyboard-event tests pass unchanged."
func TestMouseRoutingBackwardCompatCommandToDesktopFocusedChild(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	focused := newSelectablePhaseView("focused", NewRect(0, 0, 40, 12))
	d.Insert(focused)

	ev := &Event{What: EvCommand, Command: CmClose}
	d.HandleEvent(ev)

	if focused.handleCount == 0 {
		t.Error("backward compat: command event should still reach Desktop's focused child")
	}
}

// TestMouseRoutingBackwardCompatKeyboardToWindowFocusedChild verifies that
// keyboard events sent directly to a Window continue to reach its focused
// child via the internal Group's three-phase dispatch.
//
// Spec: "All existing keyboard-event tests pass unchanged."
func TestMouseRoutingBackwardCompatKeyboardToWindowFocusedChild(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	child := newSelectablePhaseView("child", NewRect(0, 0, 38, 18))
	w.Insert(child)

	ev := keyboardEvent()
	w.HandleEvent(ev)

	if child.handleCount == 0 {
		t.Error("backward compat: keyboard event should still reach Window's focused child")
	}
	if child.lastEvent != ev {
		t.Error("backward compat: Window should forward the exact keyboard event to focused child")
	}
}

// TestMouseRoutingBackwardCompatDesktopHandleEventForwardsClearedEvent verifies
// that a cleared (EvNothing) event sent to Desktop is still not forwarded to
// any child — unchanged from before Task 4.
//
// Spec: "All existing keyboard-event tests pass unchanged."
func TestMouseRoutingBackwardCompatDesktopHandleEventForwardsClearedEvent(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	focused := newSelectablePhaseView("focused", NewRect(0, 0, 40, 12))
	d.Insert(focused)

	ev := &Event{What: EvNothing}
	d.HandleEvent(ev)

	if focused.handleCount != 0 {
		t.Error("backward compat: cleared event should NOT be forwarded to any child")
	}
}
