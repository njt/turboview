package tv

// integration_phase14_batch1_test.go — Integration tests for Phase 14 Batch 1:
// CmZoom, CmSelectWindowNum, and CmResize working together in realistic scenarios.
//
// Each test creates real Desktop and Window instances — no mocks — and exercises
// the command pipeline end-to-end to verify that the three new features interact
// correctly.
//
// Features under test:
//   Task 1 (CmZoom):          Window.HandleEvent toggles zoom state on CmZoom
//   Task 2 (CmSelectWindowNum): Broadcast focuses the matching, selectable window
//   Task 3 (CmResize):        CmResize is consumed; bounds are unchanged
//
// Test naming: TestIntegrationPhase14Batch1_<DescriptiveSuffix>

import "testing"

// ---------------------------------------------------------------------------
// Scenario 1: Desktop with two windows — CmZoom via command dispatch
//
// Create a desktop with two windows. Send CmZoom to the focused window.
// Verify it zooms and fills the desktop area.
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch1_CmZoomFillsDesktopArea verifies that sending
// CmZoom to a window that has a Desktop as its owner causes the window to expand
// to fill the desktop's bounds.
//
// This exercises the window-owns-desktop relationship and confirms that Zoom()
// correctly stores the pre-zoom bounds and expands to owner area.
func TestIntegrationPhase14Batch1_CmZoomFillsDesktopArea(t *testing.T) {
	desktopBounds := NewRect(0, 0, 80, 24)
	d := NewDesktop(desktopBounds)

	w1 := NewWindow(NewRect(0, 0, 30, 10), "W1")
	w2 := NewWindow(NewRect(40, 0, 30, 10), "W2")
	d.Insert(w1)
	d.Insert(w2)
	// w2 is focused after the last insert.

	if w1.IsZoomed() {
		t.Fatal("precondition: w1 must start un-zoomed")
	}

	w1.HandleEvent(&Event{What: EvCommand, Command: CmZoom})

	if !w1.IsZoomed() {
		t.Error("after CmZoom: w1.IsZoomed() = false, want true")
	}

	// A zoomed window should fill the desktop's client area.
	got := w1.Bounds()
	if got != desktopBounds {
		t.Errorf("after CmZoom: w1.Bounds() = %v, want desktop bounds %v", got, desktopBounds)
	}
}

// TestIntegrationPhase14Batch1_CmZoomEventConsumedOnDesktop verifies that the
// CmZoom event is cleared (consumed) after being handled by a window inside a
// desktop, so it does not propagate further up the view hierarchy.
func TestIntegrationPhase14Batch1_CmZoomEventConsumedOnDesktop(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 30, 10), "W")
	d.Insert(w)

	ev := &Event{What: EvCommand, Command: CmZoom}
	w.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmZoom event not cleared after handling: What = %v, want EvNothing", ev.What)
	}
}

// ---------------------------------------------------------------------------
// Scenario 2: Desktop with three windows — CmSelectWindowNum targets correct window
//
// Create a desktop with 3 numbered windows. Focus window 1. Broadcast
// CmSelectWindowNum with number 3. Verify window 3 becomes focused.
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch1_SelectWindowNumFocusesCorrectInThree verifies that
// broadcasting CmSelectWindowNum with number 3 focuses window 3 in a 3-window
// desktop where window 1 is currently focused.
func TestIntegrationPhase14Batch1_SelectWindowNumFocusesCorrectInThree(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w1 := NewWindow(NewRect(0, 0, 25, 10), "W1", WithWindowNumber(1))
	w2 := NewWindow(NewRect(27, 0, 25, 10), "W2", WithWindowNumber(2))
	w3 := NewWindow(NewRect(54, 0, 25, 10), "W3", WithWindowNumber(3))
	d.Insert(w1)
	d.Insert(w2)
	d.Insert(w3)
	// Explicitly focus w1 so the starting state is deterministic.
	d.BringToFront(w1)

	if d.FocusedChild() != w1 {
		t.Fatalf("precondition: FocusedChild() = %v, want w1", d.FocusedChild())
	}

	// Deliver the broadcast directly to w3 (simulates desktop group dispatch).
	ev := &Event{What: EvBroadcast, Command: CmSelectWindowNum, Info: 3}
	w3.HandleEvent(ev)

	if d.FocusedChild() != w3 {
		t.Errorf("after CmSelectWindowNum(3): FocusedChild() = %v, want w3", d.FocusedChild())
	}
	if !ev.IsCleared() {
		t.Errorf("CmSelectWindowNum(3) event not cleared after matching w3")
	}
}

// TestIntegrationPhase14Batch1_SelectWindowNumNonMatchingDoesNotChangeFocus verifies
// that broadcasting CmSelectWindowNum with a number that does not match a window
// leaves focus unchanged.
func TestIntegrationPhase14Batch1_SelectWindowNumNonMatchingDoesNotChangeFocus(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w1 := NewWindow(NewRect(0, 0, 25, 10), "W1", WithWindowNumber(1))
	w2 := NewWindow(NewRect(27, 0, 25, 10), "W2", WithWindowNumber(2))
	d.Insert(w1)
	d.Insert(w2)
	d.BringToFront(w1)

	// Deliver a broadcast for number 2 to w1 (number 1) — w1 must not steal focus.
	ev := &Event{What: EvBroadcast, Command: CmSelectWindowNum, Info: 2}
	w1.HandleEvent(ev)

	// Focus must remain w1; w1 must not mis-identify itself as number 2.
	if d.FocusedChild() != w1 {
		t.Errorf("CmSelectWindowNum(2) on w1(num=1): FocusedChild() changed to %v, want w1", d.FocusedChild())
	}
	// Event must not be cleared — w1 did not consume it.
	if ev.IsCleared() {
		t.Errorf("CmSelectWindowNum(2) on w1(num=1): event was cleared, want it to pass through")
	}
}

// ---------------------------------------------------------------------------
// Scenario 3: CmZoom then CmSelectWindowNum — zoom and switch
//
// Zoom window 1, then broadcast CmSelectWindowNum for window 2.
// Verify window 2 is focused and window 1 stays zoomed.
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch1_ZoomThenSelectOtherWindowStaysZoomed verifies that
// zooming window 1 and then selecting window 2 via CmSelectWindowNum does not
// un-zoom window 1. The two commands affect independent state.
func TestIntegrationPhase14Batch1_ZoomThenSelectOtherWindowStaysZoomed(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w1 := NewWindow(NewRect(0, 0, 30, 10), "W1", WithWindowNumber(1))
	w2 := NewWindow(NewRect(40, 0, 30, 10), "W2", WithWindowNumber(2))
	d.Insert(w1)
	d.Insert(w2)
	d.BringToFront(w1)

	// Step 1: Zoom w1.
	w1.HandleEvent(&Event{What: EvCommand, Command: CmZoom})
	if !w1.IsZoomed() {
		t.Fatal("precondition: CmZoom on w1 must set IsZoomed() = true")
	}

	// Step 2: Select w2 via broadcast.
	ev := &Event{What: EvBroadcast, Command: CmSelectWindowNum, Info: 2}
	w2.HandleEvent(ev)

	// w2 must be focused.
	if d.FocusedChild() != w2 {
		t.Errorf("after CmSelectWindowNum(2): FocusedChild() = %v, want w2", d.FocusedChild())
	}
	// w1 must still be zoomed — CmSelectWindowNum must not affect zoom state.
	if !w1.IsZoomed() {
		t.Errorf("after CmSelectWindowNum(2): w1.IsZoomed() = false, want true (zoom state unchanged)")
	}
}

// TestIntegrationPhase14Batch1_ZoomThenSelectSetsEventCleared verifies that the
// CmSelectWindowNum broadcast event is cleared when the matching window handles it,
// regardless of the zoom state of other windows.
func TestIntegrationPhase14Batch1_ZoomThenSelectSetsEventCleared(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w1 := NewWindow(NewRect(0, 0, 30, 10), "W1", WithWindowNumber(1))
	w2 := NewWindow(NewRect(40, 0, 30, 10), "W2", WithWindowNumber(2))
	d.Insert(w1)
	d.Insert(w2)

	// Zoom w1 first.
	w1.HandleEvent(&Event{What: EvCommand, Command: CmZoom})

	// Select w2.
	ev := &Event{What: EvBroadcast, Command: CmSelectWindowNum, Info: 2}
	w2.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmSelectWindowNum(2) after w1 zoom: event not cleared")
	}
}

// ---------------------------------------------------------------------------
// Scenario 4: CmSelectWindowNum skips non-selectable
//
// Two windows, make one non-selectable. Broadcast its number.
// Verify focus doesn't change.
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch1_SelectNumSkipsNonSelectableWindow verifies that
// broadcasting CmSelectWindowNum for a window that has had OfSelectable disabled
// does not change focus to that window.
func TestIntegrationPhase14Batch1_SelectNumSkipsNonSelectableWindow(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	anchor := NewWindow(NewRect(0, 0, 30, 10), "Anchor", WithWindowNumber(1))
	target := NewWindow(NewRect(40, 0, 30, 10), "Target", WithWindowNumber(2))
	d.Insert(anchor)
	d.Insert(target)
	// Bring anchor to front so it is the focused window.
	d.BringToFront(anchor)

	if d.FocusedChild() != anchor {
		t.Fatalf("precondition: FocusedChild() = %v, want anchor", d.FocusedChild())
	}

	// Disable selectability on target.
	target.SetOptions(OfSelectable, false)

	// Broadcast target's number.
	ev := &Event{What: EvBroadcast, Command: CmSelectWindowNum, Info: 2}
	target.HandleEvent(ev)

	// Focus must remain on anchor.
	if d.FocusedChild() != anchor {
		t.Errorf("CmSelectWindowNum on non-selectable target: FocusedChild() = %v, want anchor", d.FocusedChild())
	}
	// Event must not be cleared.
	if ev.IsCleared() {
		t.Errorf("CmSelectWindowNum on non-selectable target: event was cleared, want it to pass through")
	}
}

// TestIntegrationPhase14Batch1_SelectNumSelectableWindowIsFocused verifies the
// affirmative counterpart: a selectable window with the matching number IS focused.
// This guards against an implementation that ignores OfSelectable for all windows.
func TestIntegrationPhase14Batch1_SelectNumSelectableWindowIsFocused(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	anchor := NewWindow(NewRect(0, 0, 30, 10), "Anchor", WithWindowNumber(1))
	selectable := NewWindow(NewRect(40, 0, 30, 10), "Selectable", WithWindowNumber(2))
	d.Insert(anchor)
	d.Insert(selectable)
	d.BringToFront(anchor)

	// selectable has OfSelectable enabled by default (NewWindow default).
	ev := &Event{What: EvBroadcast, Command: CmSelectWindowNum, Info: 2}
	selectable.HandleEvent(ev)

	if d.FocusedChild() != selectable {
		t.Errorf("CmSelectWindowNum on selectable window: FocusedChild() = %v, want selectable", d.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Scenario 5: CmResize on focused window
//
// Send CmResize to a window, verify bounds unchanged and event consumed.
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch1_CmResizeBoundsUnchanged verifies that sending
// CmResize to a window in a desktop does not change the window's bounds. The
// keyboard-driven resize feature is not yet implemented; the event is consumed
// as a placeholder.
func TestIntegrationPhase14Batch1_CmResizeBoundsUnchanged(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	original := NewRect(10, 5, 30, 12)
	w := NewWindow(original, "W")
	d.Insert(w)

	w.HandleEvent(&Event{What: EvCommand, Command: CmResize})

	if w.Bounds() != original {
		t.Errorf("CmResize changed bounds: got %v, want %v (unchanged)", w.Bounds(), original)
	}
}

// TestIntegrationPhase14Batch1_CmResizeEventConsumed verifies that the CmResize
// event is cleared (consumed) and does not propagate to the desktop or other views.
func TestIntegrationPhase14Batch1_CmResizeEventConsumed(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(10, 5, 30, 12), "W")
	d.Insert(w)

	ev := &Event{What: EvCommand, Command: CmResize}
	w.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmResize event not cleared: What = %v, want EvNothing", ev.What)
	}
}

// TestIntegrationPhase14Batch1_CmResizeAfterZoomBoundsUnchanged verifies that
// sending CmResize after zooming a window does not alter the zoomed bounds. The
// two commands interact on the same window, and CmResize must remain a no-op for
// state changes regardless of zoom status.
func TestIntegrationPhase14Batch1_CmResizeAfterZoomBoundsUnchanged(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 30, 10), "W")
	d.Insert(w)

	// Zoom first so the window now has expanded (desktop-filling) bounds.
	w.HandleEvent(&Event{What: EvCommand, Command: CmZoom})
	if !w.IsZoomed() {
		t.Fatal("precondition: window must be zoomed")
	}
	zoomedBounds := w.Bounds()

	// Now send CmResize — bounds must be unchanged.
	ev := &Event{What: EvCommand, Command: CmResize}
	w.HandleEvent(ev)

	if w.Bounds() != zoomedBounds {
		t.Errorf("CmResize after zoom changed bounds: got %v, want %v (zoomed bounds unchanged)", w.Bounds(), zoomedBounds)
	}
	if !ev.IsCleared() {
		t.Errorf("CmResize after zoom: event not cleared")
	}
}

// TestIntegrationPhase14Batch1_CmZoomUnzoomRestoresBoundsInDesktop verifies the
// full round-trip of CmZoom in a realistic desktop: zoom in, then zoom out, and
// confirm the original bounds are restored.
//
// This is an integration-level complement to the unit test
// TestWindowCmZoom_Falsify_DoubleCmZoomRestoresBounds.
func TestIntegrationPhase14Batch1_CmZoomUnzoomRestoresBoundsInDesktop(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	original := NewRect(5, 3, 30, 10)
	w := NewWindow(original, "W")
	d.Insert(w)

	// First CmZoom: zoom in.
	w.HandleEvent(&Event{What: EvCommand, Command: CmZoom})
	if !w.IsZoomed() {
		t.Fatal("precondition: first CmZoom must zoom the window")
	}

	// Second CmZoom: zoom out.
	w.HandleEvent(&Event{What: EvCommand, Command: CmZoom})

	if w.IsZoomed() {
		t.Error("after second CmZoom: IsZoomed() = true, want false (toggle)")
	}
	if w.Bounds() != original {
		t.Errorf("after double CmZoom: Bounds() = %v, want original %v", w.Bounds(), original)
	}
}
