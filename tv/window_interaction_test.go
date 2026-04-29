package tv

// window_interaction_test.go — tests for Task 5: Window Interaction.
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment quoting the relevant spec requirement.
//
// Spec summary:
//   Drag: title bar Button1 sets SfDragging, records drag offset, subsequent
//   Button1 events update window position, release clears SfDragging, event
//   is cleared. Close/zoom icon clicks do not start drag. Client area does not.
//
//   Mouse capture (drag/resize): Desktop routes ALL mouse events to the window
//   with SfDragging set (or resizing==true) WITHOUT coordinate translation.
//
//   Resize: bottom-right corner starts resize (resizeLeft=false); bottom-left
//   starts resize (resizeLeft=true). Button1 held adjusts size; minimum 10×5.
//   Release ends resize. Event is cleared.
//
//   Close: Button1 on close icon [×] at (1-3, 0) transforms to EvCommand/CmClose.
//   Desktop handles CmClose: removes focused child, clears event.
//
//   Zoom: zoom icon (width-4 to width-2, 0) or double-click on title bar toggles
//   zoom. Zoom in fills owner's area and stores old bounds; zoom out restores.
//   Icon changes to '↕' when zoomed. Window.Zoom() is a public toggle.
//   Window.IsZoomed() returns current state.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// clickEvent creates an EvMouse Event with ClickCount=1 (single click).
func clickEvent(x, y int, button tcell.ButtonMask) *Event {
	return &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:          x,
			Y:          y,
			Button:     button,
			ClickCount: 1,
		},
	}
}

// doubleClickEvent creates an EvMouse Event with ClickCount=2.
func doubleClickEvent(x, y int, button tcell.ButtonMask) *Event {
	return &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:          x,
			Y:          y,
			Button:     button,
			ClickCount: 2,
		},
	}
}

// newTestWindow returns a 20×10 Window at origin (0,0) — width=20, height=10.
// close icon: x=1,2,3  y=0
// zoom icon:  x=16,17,18  y=0  (width-4=16, width-2=18)
// title bar:  y=0, x=4..15 (safe drag zone)
// bottom-right corner: x=19, y=9
// bottom-left  corner: x=0,  y=9
func newTestWindow() *Window {
	return NewWindow(NewRect(0, 0, 20, 10), "Test")
}

// ---------------------------------------------------------------------------
// Drag — 8+ tests
// ---------------------------------------------------------------------------

// TestDragTitleBarSetsSfDragging verifies that a Button1 press on the title bar
// (y=0, x not on close icon 1-3 or zoom icon width-4 to width-2) sets SfDragging.
//
// Spec: "Left-click (Button1) on the title bar (y=0, x not on close icon 1-3
// or zoom icon width-4 to width-2) sets SfDragging."
func TestDragTitleBarSetsSfDragging(t *testing.T) {
	w := newTestWindow()

	// x=8 is on the title bar, not on any icon.
	ev := clickEvent(8, 0, tcell.Button1)
	w.HandleEvent(ev)

	if !w.HasState(SfDragging) {
		t.Error("title bar Button1 click should set SfDragging")
	}
}

// TestDragTitleBarRecordsDragOffset verifies that the drag offset is the
// window-local mouse position at the time of the title bar click.
//
// Spec: "Left-click (Button1) on the title bar … records the drag offset
// (window-local mouse position)."
func TestDragTitleBarRecordsDragOffset(t *testing.T) {
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")

	// Send window-local click at (8, 0) — window-local because handleMouseEvent
	// receives coords already in window-local space (Desktop translates before
	// forwarding to the window via normal hit-test path).
	ev := clickEvent(8, 0, tcell.Button1)
	w.HandleEvent(ev)

	// dragOff should equal the window-local mouse position at click time.
	if w.dragOff.X != 8 || w.dragOff.Y != 0 {
		t.Errorf("drag offset = (%d,%d), want (8,0)", w.dragOff.X, w.dragOff.Y)
	}
}

// TestDragButton1HeldUpdatesWindowPosition verifies that subsequent mouse events
// with Button1 held update the window's position using the Desktop-local
// coordinates and the stored drag offset.
//
// Spec: "Subsequent mouse events while Button1 is held update the window's
// position." and "Desktop does NOT translate coordinates to window-local
// during capture (passes Desktop-local coordinates)."
func TestDragButton1HeldUpdatesWindowPosition(t *testing.T) {
	// Window starts at Desktop position (5, 3).
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	// Step 1: start drag — click title bar at window-local (8,0).
	// Desktop translates (5+8=13, 3+0=3) to window-local; window sees (8,0).
	startEv := mouseEvent(13, 3, tcell.Button1)
	d.HandleEvent(startEv)

	if !w.HasState(SfDragging) {
		t.Fatal("drag should have started; SfDragging not set")
	}

	// Step 2: move mouse to Desktop-local (20, 10) with Button1 still held.
	// Expected new position: (20 - dragOff.X, 10 - dragOff.Y) = (20-8, 10-0) = (12, 10).
	moveEv := mouseEvent(20, 10, tcell.Button1)
	d.HandleEvent(moveEv)

	got := w.Bounds()
	wantX, wantY := 20-8, 10-0 // 12, 10
	if got.A.X != wantX || got.A.Y != wantY {
		t.Errorf("after drag move: origin = (%d,%d), want (%d,%d)",
			got.A.X, got.A.Y, wantX, wantY)
	}
}

// TestDragReleaseClearsSfDragging verifies that releasing Button1 (ButtonNone)
// ends the drag by clearing SfDragging.
//
// Spec: "Releasing Button1 clears SfDragging."
func TestDragReleaseClearsSfDragging(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	// Start drag.
	d.HandleEvent(mouseEvent(13, 3, tcell.Button1))
	if !w.HasState(SfDragging) {
		t.Fatal("drag should have started")
	}

	// Release.
	d.HandleEvent(mouseEvent(13, 3, tcell.ButtonNone))

	if w.HasState(SfDragging) {
		t.Error("releasing Button1 should clear SfDragging")
	}
}

// TestDragEventIsCleared verifies that the mouse event that starts dragging
// is cleared (set to EvNothing) so it is not processed further.
//
// Spec: "Drag events clear the event (set to EvNothing)."
func TestDragEventIsCleared(t *testing.T) {
	w := newTestWindow()

	ev := clickEvent(8, 0, tcell.Button1)
	w.HandleEvent(ev)

	if ev.What != EvNothing {
		t.Errorf("drag start event.What = %v, want EvNothing", ev.What)
	}
}

// TestDragMoveEventIsCleared verifies that mouse move events during drag are
// also cleared.
//
// Spec: "Drag events clear the event (set to EvNothing)."
func TestDragMoveEventIsCleared(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	// Start drag.
	d.HandleEvent(mouseEvent(13, 3, tcell.Button1))

	// Move event during drag.
	moveEv := mouseEvent(20, 10, tcell.Button1)
	d.HandleEvent(moveEv)

	if moveEv.What != EvNothing {
		t.Errorf("drag move event.What = %v, want EvNothing", moveEv.What)
	}
}

// TestDragCloseIconDoesNotStartDrag verifies that clicking on the close icon
// (x=1,2,3, y=0) does NOT set SfDragging.
//
// Spec: "Left-click on the title bar (y=0, x not on close icon 1-3 …) sets
// SfDragging." — contrapositive: close icon click must not start drag.
func TestDragCloseIconDoesNotStartDrag(t *testing.T) {
	w := newTestWindow()

	for _, x := range []int{1, 2, 3} {
		ev := clickEvent(x, 0, tcell.Button1)
		w.HandleEvent(ev)
		if w.HasState(SfDragging) {
			t.Errorf("close icon click at x=%d should NOT set SfDragging", x)
		}
	}
}

// TestDragZoomIconDoesNotStartDrag verifies that clicking on the zoom icon
// (x=width-4 to width-2, y=0) does NOT set SfDragging.
//
// Spec: "Left-click on the title bar (y=0, x not on … zoom icon width-4 to
// width-2) sets SfDragging."
func TestDragZoomIconDoesNotStartDrag(t *testing.T) {
	w := newTestWindow() // width=20: zoom icon at x=16,17,18
	for _, x := range []int{16, 17, 18} {
		w.SetState(SfDragging, false) // reset between iterations
		ev := clickEvent(x, 0, tcell.Button1)
		w.HandleEvent(ev)
		if w.HasState(SfDragging) {
			t.Errorf("zoom icon click at x=%d should NOT set SfDragging", x)
		}
	}
}

// TestDragClientAreaDoesNotSetSfDragging verifies that a click on the client
// area (not the title bar) does NOT set SfDragging.
//
// Spec: "Left-click (Button1) on the title bar (y=0, …) sets SfDragging." —
// y>0 is the client area, not the title bar.
func TestDragClientAreaDoesNotSetSfDragging(t *testing.T) {
	w := newTestWindow()

	// Client area: y=5, x=8.
	ev := clickEvent(8, 5, tcell.Button1)
	w.HandleEvent(ev)

	if w.HasState(SfDragging) {
		t.Error("client area click should NOT set SfDragging")
	}
}

// ---------------------------------------------------------------------------
// Mouse capture during drag — 3+ tests
// ---------------------------------------------------------------------------

// TestDesktopRoutesDragCaptureToWindow verifies that Desktop routes ALL mouse
// events to the dragging window, regardless of where the cursor is.
//
// Spec: "Desktop routes ALL mouse events to a dragging window (the one with
// SfDragging set) regardless of hit-testing."
func TestDesktopRoutesDragCaptureToWindow(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	bystander := newSelectablePhaseView("bystander", NewRect(40, 0, 20, 10))
	d.Insert(w)
	d.Insert(bystander)

	// Start drag on w.
	d.HandleEvent(mouseEvent(13, 3, tcell.Button1))
	if !w.HasState(SfDragging) {
		t.Fatal("drag should have started")
	}

	// Reset bystander count.
	bystander.handleCount = 0

	// Move mouse to bystander's area — the Desktop should still route to w.
	d.HandleEvent(mouseEvent(50, 5, tcell.Button1))

	if bystander.handleCount != 0 {
		t.Error("bystander should NOT receive events during drag capture")
	}
}

// TestDesktopDragCaptureNoCoordinateTranslation verifies that during drag
// capture the Desktop passes Desktop-local (not window-local) coordinates.
//
// Spec: "Desktop does NOT translate coordinates to window-local during capture
// (passes Desktop-local coordinates)."
func TestDesktopDragCaptureNoCoordinateTranslation(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	// Window at Desktop position (5,3); width=20, height=10.
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	// Start drag at Desktop-local (13,3) → window-local (8,0).
	d.HandleEvent(mouseEvent(13, 3, tcell.Button1))

	// Move to Desktop-local (20, 10). During capture, window gets (20,10).
	// If Desktop translated, window would get (20-5, 10-3)=(15,7).
	// The new position should be: Desktop-local minus dragOff = (20-8, 10-0)=(12,10).
	moveEv := mouseEvent(20, 10, tcell.Button1)
	d.HandleEvent(moveEv)

	bounds := w.Bounds()
	// If translation was wrongly applied, origin would be wrong.
	if bounds.A.X != 12 || bounds.A.Y != 10 {
		t.Errorf("after captured drag: window origin = (%d,%d), want (12,10) — no coord translation during capture",
			bounds.A.X, bounds.A.Y)
	}
}

// TestDesktopDragCaptureEndsWhenSfDraggingCleared verifies that once
// SfDragging is cleared (Button1 released), the Desktop resumes normal
// hit-testing.
//
// Spec: "Desktop routes ALL mouse events to a dragging window … regardless of
// hit-testing" — ends when SfDragging is cleared.
func TestDesktopDragCaptureEndsWhenSfDraggingCleared(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	bystander := newSelectablePhaseView("bystander", NewRect(40, 0, 20, 10))
	d.Insert(w)
	d.Insert(bystander)

	// Start and then release drag.
	d.HandleEvent(mouseEvent(13, 3, tcell.Button1))
	d.HandleEvent(mouseEvent(13, 3, tcell.ButtonNone)) // release

	if w.HasState(SfDragging) {
		t.Fatal("SfDragging should be cleared after release")
	}

	// Now click inside bystander — should go to bystander via normal routing.
	bystander.handleCount = 0
	d.HandleEvent(mouseEvent(50, 5, tcell.Button1))
	if bystander.handleCount == 0 {
		t.Error("after drag ends, normal hit-testing should resume and bystander should receive the click")
	}
}

// ---------------------------------------------------------------------------
// Resize — 8+ tests
// ---------------------------------------------------------------------------

// TestResizeBottomRightCornerStartsResize verifies that Button1 on the
// bottom-right corner (x==width-1, y==height-1) starts resizing with
// resizeLeft=false.
//
// Spec: "Left-click on bottom-right corner (x == width-1 && y == height-1)
// starts resizing (resizeLeft=false)."
func TestResizeBottomRightCornerStartsResize(t *testing.T) {
	w := newTestWindow() // width=20, height=10

	ev := clickEvent(19, 9, tcell.Button1) // (width-1, height-1)
	w.HandleEvent(ev)

	if !w.resizing {
		t.Error("bottom-right corner click should start resizing (resizing=true)")
	}
	if w.resizeLeft {
		t.Error("bottom-right corner resize should set resizeLeft=false")
	}
}

// TestResizeBottomLeftCornerStartsResize verifies that Button1 on the
// bottom-left corner (x==0, y==height-1) starts resizing with resizeLeft=true.
//
// Spec: "Left-click on bottom-left corner (x == 0 && y == height-1) starts
// resizing (resizeLeft=true)."
func TestResizeBottomLeftCornerStartsResize(t *testing.T) {
	w := newTestWindow() // height=10

	ev := clickEvent(0, 9, tcell.Button1) // (0, height-1)
	w.HandleEvent(ev)

	if !w.resizing {
		t.Error("bottom-left corner click should start resizing (resizing=true)")
	}
	if !w.resizeLeft {
		t.Error("bottom-left corner resize should set resizeLeft=true")
	}
}

// TestResizeBottomRightAdjustsWidthAndHeight verifies that during a
// bottom-right resize, subsequent Button1 events adjust both width and height.
//
// Spec: "During resize with Button1 held: bottom-right adjusts width and height."
func TestResizeBottomRightAdjustsWidthAndHeight(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	// Window at (5,3) with width=20, height=10.
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	// Start resize at bottom-right corner — Desktop-local (5+19, 3+9) = (24,12).
	d.HandleEvent(mouseEvent(24, 12, tcell.Button1))
	if !w.resizing {
		t.Fatal("resize should have started")
	}

	// Move to Desktop-local (30, 16) with Button1 held.
	// New width  = (30 - window.A.X) + 1 = (30-5)+1 = 26
	// New height = (16 - window.A.Y) + 1 = (16-3)+1 = 14
	d.HandleEvent(mouseEvent(30, 16, tcell.Button1))

	bounds := w.Bounds()
	if bounds.Width() != 26 {
		t.Errorf("resize width = %d, want 26", bounds.Width())
	}
	if bounds.Height() != 14 {
		t.Errorf("resize height = %d, want 14", bounds.Height())
	}
	// Left edge must not have moved.
	if bounds.A.X != 5 {
		t.Errorf("resize left edge = %d, want 5 (unchanged)", bounds.A.X)
	}
}

// TestResizeBottomLeftAdjustsLeftEdgeKeepsRightFixed verifies that during a
// bottom-left resize, the left edge moves while the right edge stays fixed.
//
// Spec: "bottom-left adjusts left edge and height keeping right edge fixed."
func TestResizeBottomLeftAdjustsLeftEdgeKeepsRightFixed(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	// Window at (10,3) with width=20, height=10. Right edge at x=30.
	w := NewWindow(NewRect(10, 3, 20, 10), "Test")
	d.Insert(w)

	// Start resize at bottom-left corner — Desktop-local (10+0, 3+9) = (10,12).
	d.HandleEvent(mouseEvent(10, 12, tcell.Button1))
	if !w.resizing {
		t.Fatal("resize should have started")
	}

	// Move to Desktop-local (7, 16) with Button1 held.
	// New left  = 7 (Desktop-local x)
	// Right stays at 30.
	// New width  = 30 - 7 = 23
	// New height = (16 - 3) + 1 = 14
	d.HandleEvent(mouseEvent(7, 16, tcell.Button1))

	bounds := w.Bounds()
	if bounds.A.X != 7 {
		t.Errorf("bottom-left resize: left edge = %d, want 7", bounds.A.X)
	}
	if bounds.B.X != 30 {
		t.Errorf("bottom-left resize: right edge = %d, want 30 (fixed)", bounds.B.X)
	}
	if bounds.Height() != 14 {
		t.Errorf("bottom-left resize: height = %d, want 14", bounds.Height())
	}
}

// TestResizeMinimumWidthEnforced verifies that resizing cannot make the window
// narrower than 10 columns.
//
// Spec: "Minimum window size: width 10, height 5."
func TestResizeMinimumWidthEnforced(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	// Start bottom-right resize.
	d.HandleEvent(mouseEvent(24, 12, tcell.Button1))

	// Try to resize to width=3 (below minimum).
	// New width = (6 - 5) + 1 = 2 → clamped to 10.
	d.HandleEvent(mouseEvent(6, 12, tcell.Button1))

	if w.Bounds().Width() < 10 {
		t.Errorf("width = %d, want at least 10 (minimum enforced)", w.Bounds().Width())
	}
}

// TestResizeMinimumHeightEnforced verifies that resizing cannot make the window
// shorter than 5 rows.
//
// Spec: "Minimum window size: width 10, height 5."
func TestResizeMinimumHeightEnforced(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	// Start bottom-right resize.
	d.HandleEvent(mouseEvent(24, 12, tcell.Button1))

	// Try to resize to height=2 (below minimum).
	// New height = (5 - 3) + 1 = 3 → clamped to 5.
	d.HandleEvent(mouseEvent(24, 5, tcell.Button1))

	if w.Bounds().Height() < 5 {
		t.Errorf("height = %d, want at least 5 (minimum enforced)", w.Bounds().Height())
	}
}

// TestResizeReleaseEndsResize verifies that releasing Button1 ends the resize.
//
// Spec: "Releasing Button1 ends resize."
func TestResizeReleaseEndsResize(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	d.HandleEvent(mouseEvent(24, 12, tcell.Button1))
	if !w.resizing {
		t.Fatal("resize should have started")
	}

	d.HandleEvent(mouseEvent(24, 12, tcell.ButtonNone))

	if w.resizing {
		t.Error("releasing Button1 should end resize (resizing=false)")
	}
}

// TestResizeEventIsCleared verifies that resize-related events are cleared.
//
// Spec: "Resize events clear the event."
func TestResizeEventIsCleared(t *testing.T) {
	w := newTestWindow()

	ev := clickEvent(19, 9, tcell.Button1) // bottom-right corner
	w.HandleEvent(ev)

	if ev.What != EvNothing {
		t.Errorf("resize start event.What = %v, want EvNothing", ev.What)
	}
}

// TestResizeMoveEventIsCleared verifies that move events during resize are cleared.
//
// Spec: "Resize events clear the event."
func TestResizeMoveEventIsCleared(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	d.HandleEvent(mouseEvent(24, 12, tcell.Button1))

	moveEv := mouseEvent(30, 16, tcell.Button1)
	d.HandleEvent(moveEv)

	if moveEv.What != EvNothing {
		t.Errorf("resize move event.What = %v, want EvNothing", moveEv.What)
	}
}

// ---------------------------------------------------------------------------
// Mouse capture during resize — 2+ tests
// ---------------------------------------------------------------------------

// TestDesktopRoutesResizeCaptureToWindow verifies that Desktop routes ALL mouse
// events to the resizing window regardless of cursor position.
//
// Spec: "Same as drag — Desktop routes to resizing window without coordinate
// translation."
func TestDesktopRoutesResizeCaptureToWindow(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	bystander := newSelectablePhaseView("bystander", NewRect(40, 0, 20, 10))
	d.Insert(w)
	d.Insert(bystander)

	// Start resize on w (bottom-right corner: Desktop-local 24,12).
	d.HandleEvent(mouseEvent(24, 12, tcell.Button1))
	if !w.resizing {
		t.Fatal("resize should have started")
	}

	bystander.handleCount = 0

	// Move mouse to bystander area — Desktop should still route to w.
	d.HandleEvent(mouseEvent(50, 5, tcell.Button1))

	if bystander.handleCount != 0 {
		t.Error("bystander should NOT receive events during resize capture")
	}
}

// TestDesktopResizeCaptureEndsOnRelease verifies that after Button1 release
// the Desktop returns to normal hit-testing.
//
// Spec: "Capture ends when resize finishes."
func TestDesktopResizeCaptureEndsOnRelease(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	bystander := newSelectablePhaseView("bystander", NewRect(40, 0, 20, 10))
	d.Insert(w)
	d.Insert(bystander)

	// Start and release resize.
	d.HandleEvent(mouseEvent(24, 12, tcell.Button1))
	d.HandleEvent(mouseEvent(24, 12, tcell.ButtonNone))

	if w.resizing {
		t.Fatal("resizing should be false after release")
	}

	// Normal routing should be restored.
	bystander.handleCount = 0
	d.HandleEvent(mouseEvent(50, 5, tcell.Button1))
	if bystander.handleCount == 0 {
		t.Error("after resize ends, bystander should receive click via normal routing")
	}
}

// ---------------------------------------------------------------------------
// Close — 3+ tests
// ---------------------------------------------------------------------------

// TestCloseIconTransformsEventToCmClose verifies that Button1 on the close icon
// (x=1,2,3, y=0) transforms the event to EvCommand/CmClose.
//
// Spec: "Left-click on close icon [×] at (1-3, 0) transforms event to
// EvCommand/CmClose."
func TestCloseIconTransformsEventToCmClose(t *testing.T) {
	for _, x := range []int{1, 2, 3} {
		ev := clickEvent(x, 0, tcell.Button1)
		w := newTestWindow()
		w.HandleEvent(ev)

		if ev.What != EvCommand {
			t.Errorf("close icon click at x=%d: event.What = %v, want EvCommand", x, ev.What)
		}
		if ev.Command != CmClose {
			t.Errorf("close icon click at x=%d: event.Command = %v, want CmClose", x, ev.Command)
		}
	}
}

// TestDesktopHandlesCmCloseRemovesFocusedChild verifies that Desktop handles
// CmClose by removing the focused child.
//
// Spec: "Desktop handles CmClose: removes the focused child."
func TestDesktopHandlesCmCloseRemovesFocusedChild(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	if d.FocusedChild() != w {
		t.Fatal("window should be the focused child before close")
	}

	// Click close icon — Desktop sees the transformed CmClose.
	ev := clickEvent(5+1, 3+0, tcell.Button1) // Desktop-local close icon position
	d.HandleEvent(ev)

	// w should have been removed from the Desktop.
	for _, child := range d.Children() {
		if child == w {
			t.Error("CmClose should have removed the focused child from Desktop")
		}
	}
}

// TestDesktopCmCloseClearsEvent verifies that Desktop clears the event after
// handling CmClose.
//
// Spec: "Desktop handles CmClose … clears the event."
func TestDesktopCmCloseClearsEvent(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	// Click close icon via Desktop-local coordinates.
	ev := clickEvent(5+1, 3+0, tcell.Button1)
	d.HandleEvent(ev)

	if ev.What != EvNothing {
		t.Errorf("after CmClose handling: event.What = %v, want EvNothing", ev.What)
	}
}

// ---------------------------------------------------------------------------
// Zoom — 8+ tests
// ---------------------------------------------------------------------------

// TestZoomIsZoomedFalseInitially verifies that a newly created Window is not
// zoomed.
//
// Spec: "Window.IsZoomed() returns current zoom state."
func TestZoomIsZoomedFalseInitially(t *testing.T) {
	w := newTestWindow()

	if w.IsZoomed() {
		t.Error("new window should not be zoomed initially")
	}
}

// TestZoomIconClickTogglesZoom verifies that Button1 on the zoom icon
// (x=width-4 to width-2, y=0) toggles zoom.
//
// Spec: "Left-click on zoom icon at (width-4 to width-2, 0) toggles zoom."
func TestZoomIconClickTogglesZoom(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	// Zoom icon at window-local x=16 (width-4), Desktop-local x=5+16=21.
	ev := clickEvent(21, 3, tcell.Button1)
	d.HandleEvent(ev)

	if !w.IsZoomed() {
		t.Error("zoom icon click should toggle zoom on (IsZoomed=true)")
	}

	// Click again to zoom out.
	ev2 := clickEvent(21, 3, tcell.Button1)
	// Window has moved; use absolute position again.
	// After zoom-in the window fills the desktop (0,0,80,24).
	// Zoom icon is now at x=80-4=76, y=0 — Desktop-local same since window is at 0,0.
	ev2 = clickEvent(76, 0, tcell.Button1)
	d.HandleEvent(ev2)

	if w.IsZoomed() {
		t.Error("second zoom icon click should toggle zoom off (IsZoomed=false)")
	}
}

// TestZoomInSetsIsZoomedTrue verifies IsZoomed() returns true after zooming in.
//
// Spec: "Window.IsZoomed() returns current zoom state." and zoom-in enables zoom.
func TestZoomInSetsIsZoomedTrue(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	w.Zoom()

	if !w.IsZoomed() {
		t.Error("after Zoom(): IsZoomed() should return true")
	}
}

// TestZoomInSetsBoundsToOwnerArea verifies that zooming in sets the window's
// bounds to fill the owner's area (origin 0,0, owner's width/height).
//
// Spec: "Zoom in: stores current bounds, sets bounds to fill owner's area
// (origin 0,0, owner's width/height)."
func TestZoomInSetsBoundsToOwnerArea(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	w.Zoom()

	bounds := w.Bounds()
	ownerBounds := d.Bounds()
	if bounds.A.X != 0 || bounds.A.Y != 0 {
		t.Errorf("zoomed origin = (%d,%d), want (0,0)", bounds.A.X, bounds.A.Y)
	}
	if bounds.Width() != ownerBounds.Width() || bounds.Height() != ownerBounds.Height() {
		t.Errorf("zoomed size = (%d,%d), want (%d,%d)",
			bounds.Width(), bounds.Height(),
			ownerBounds.Width(), ownerBounds.Height())
	}
}

// TestZoomOutRestoresOriginalBounds verifies that zooming out (second call to
// Zoom()) restores the bounds stored before zooming in.
//
// Spec: "Zoom out: restores stored bounds."
func TestZoomOutRestoresOriginalBounds(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	originalBounds := NewRect(5, 3, 20, 10)
	w := NewWindow(originalBounds, "Test")
	d.Insert(w)

	w.Zoom() // zoom in
	w.Zoom() // zoom out

	bounds := w.Bounds()
	if bounds != originalBounds {
		t.Errorf("after zoom out: bounds = %v, want %v (original restored)", bounds, originalBounds)
	}
}

// TestZoomDoubleClickOnTitleBarTogglesZoom verifies that a double-click
// (ClickCount >= 2) on the title bar (y=0, not on icons) toggles zoom.
//
// Spec: "Double-click on title bar (y=0, ClickCount >= 2, not on close/zoom
// icons) toggles zoom."
func TestZoomDoubleClickOnTitleBarTogglesZoom(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	// Double-click on title bar at window-local x=8, Desktop-local x=13.
	ev := doubleClickEvent(13, 3, tcell.Button1)
	d.HandleEvent(ev)

	if !w.IsZoomed() {
		t.Error("double-click on title bar should toggle zoom on")
	}
}

// TestZoomDoubleClickOnCloseIconDoesNotToggle verifies that a double-click on
// the close icon does NOT toggle zoom (it triggers close instead).
//
// Spec: "Double-click on title bar (y=0, ClickCount >= 2, not on close/zoom
// icons) toggles zoom." — close icon is excluded.
func TestZoomDoubleClickOnCloseIconDoesNotToggle(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	// Double-click at close icon x=2 (window-local), Desktop-local x=7.
	ev := doubleClickEvent(7, 3, tcell.Button1)
	d.HandleEvent(ev)

	// Window may have been removed (CmClose); either way zoom must not be true.
	if w.IsZoomed() {
		t.Error("double-click on close icon should NOT toggle zoom")
	}
}

// TestZoomIconDrawnAs↕WhenZoomed verifies that Draw renders the zoom icon as
// '↕' (rather than '↑') when the window is zoomed.
//
// Spec: "Zoom icon changes to '↕' when zoomed (in Draw)."
func TestZoomIconDrawnAs_WhenZoomed(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	d.Insert(w)

	w.Zoom() // zoom in

	buf := NewDrawBuffer(80, 24)
	w.Draw(buf)

	// After zooming, window fills owner (80×24). Zoom icon middle at width-3.
	// New width is 80; zoom icon middle at x=80-3=77.
	zoomMid := buf.GetCell(80-3, 0)
	if zoomMid.Rune != '↕' {
		t.Errorf("zoomed window Draw: zoom icon rune = %q, want '↕'", zoomMid.Rune)
	}
}

// TestZoomIconDrawnAsUpWhenNotZoomed verifies that Draw renders '↑' when not zoomed.
//
// Spec: "Zoom icon changes to '↕' when zoomed." — contrapositive: not zoomed shows '↑'.
func TestZoomIconDrawnAsUpWhenNotZoomed(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	buf := NewDrawBuffer(20, 10)
	w.Draw(buf)

	// Zoom icon middle at width-3=17.
	zoomMid := buf.GetCell(17, 0)
	if zoomMid.Rune != '↑' {
		t.Errorf("unzoomed window Draw: zoom icon rune = %q, want '↑'", zoomMid.Rune)
	}
}

// TestZoomPublicMethodToggles verifies Window.Zoom() as a public toggle method.
//
// Spec: "Window.Zoom() toggles zoom state (public method)."
func TestZoomPublicMethodToggles(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "Test")
	d.Insert(w)

	w.Zoom()
	if !w.IsZoomed() {
		t.Error("Zoom(): first call should set IsZoomed=true")
	}

	w.Zoom()
	if w.IsZoomed() {
		t.Error("Zoom(): second call should set IsZoomed=false")
	}
}

// TestZoomWithNoOwnerStillSetsZoomed verifies that calling Zoom() when the
// window has no owner still toggles the zoomed flag without panicking.
//
// Spec: "Zoom when no owner still sets zoomed=true."
func TestZoomWithNoOwnerStillSetsZoomed(t *testing.T) {
	w := newTestWindow() // no owner (not inserted into Desktop)

	// Must not panic.
	w.Zoom()

	if !w.IsZoomed() {
		t.Error("Zoom() with no owner should still set zoomed=true")
	}
}

// ---------------------------------------------------------------------------
// Desktop CmClose handling — 2+ additional focused tests
// ---------------------------------------------------------------------------

// TestDesktopCmCloseCommandEventRemovesFocusedChild verifies that when Desktop
// receives an EvCommand/CmClose event directly (not via mouse), it removes the
// focused child.
//
// Spec: "Desktop handles CmClose: removes the focused child."
func TestDesktopCmCloseCommandEventRemovesFocusedChild(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "First")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "Second")
	d.Insert(w1)
	d.Insert(w2)
	// w2 is now focused.

	ev := &Event{What: EvCommand, Command: CmClose}
	d.HandleEvent(ev)

	for _, child := range d.Children() {
		if child == w2 {
			t.Error("Desktop.HandleEvent CmClose should have removed the focused child (w2)")
		}
	}
	// w1 should still be present.
	found := false
	for _, child := range d.Children() {
		if child == w1 {
			found = true
		}
	}
	if !found {
		t.Error("Desktop.HandleEvent CmClose should only remove the focused child, not all children")
	}
}

// TestDesktopCmCloseDirectEventClears verifies that a directly delivered
// EvCommand/CmClose event is cleared after handling.
//
// Spec: "Desktop handles CmClose … clears the event after removing."
func TestDesktopCmCloseDirectEventClears(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(0, 0, 20, 10), "W")
	d.Insert(w)

	ev := &Event{What: EvCommand, Command: CmClose}
	d.HandleEvent(ev)

	if ev.What != EvNothing {
		t.Errorf("after CmClose: event.What = %v, want EvNothing", ev.What)
	}
}
