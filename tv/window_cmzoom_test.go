package tv

// window_cmzoom_test.go — tests for Phase 14 Task 1: CmZoom and CmResize
// command handling in Window.HandleEvent.
//
// Spec (section 11.1):
//   Window.HandleEvent must handle EvCommand/CmZoom and EvCommand/CmResize
//   BEFORE delegating to the group, so the window gets first chance.
//
//   CmZoom:   call w.Zoom(), clear the event (consume it).
//   CmResize: clear the event (keyboard-driven resize deferred to later task).
//
//   Existing behaviour that must be preserved:
//   - Modal window CmClose → CmCancel (does NOT clear the event).
//   - Non-modal window passes CmClose through to the group unchanged.
//   - Unknown commands reach the group (CmZoom/CmResize must not break
//     general passthrough).

import "testing"

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// cmZoomEvent returns a fresh EvCommand/CmZoom event.
func cmZoomEvent() *Event {
	return &Event{What: EvCommand, Command: CmZoom}
}

// cmResizeEvent returns a fresh EvCommand/CmResize event.
func cmResizeEvent() *Event {
	return &Event{What: EvCommand, Command: CmResize}
}

// newZoomTestWindow returns a 20×10 Window at origin — no owner attached, so
// Zoom() keeps the existing bounds when zooming in (owner branch is skipped).
func newZoomTestWindow() *Window {
	return NewWindow(NewRect(0, 0, 20, 10), "ZoomTest")
}

// newZoomTestWindowWithOwner creates a Desktop (80×24), inserts a 20×10 window,
// and returns both so that Zoom() can fill the owner's area.
func newZoomTestWindowWithOwner() (*Window, *Desktop) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 3, 20, 10), "ZoomOwner")
	d.Insert(w)
	return w, d
}

// ---------------------------------------------------------------------------
// Test 1: CmZoom toggles zoom state from un-zoomed to zoomed
// ---------------------------------------------------------------------------

// TestWindowCmZoom_TogglesZoomState verifies that sending CmZoom to an
// un-zoomed window sets IsZoomed() to true.
//
// Spec: "CmZoom: call w.Zoom() — toggles zoom state."
func TestWindowCmZoom_TogglesZoomState(t *testing.T) {
	w, _ := newZoomTestWindowWithOwner()

	if w.IsZoomed() {
		t.Fatal("precondition: window must start un-zoomed")
	}

	w.HandleEvent(cmZoomEvent())

	if !w.IsZoomed() {
		t.Error("after CmZoom: IsZoomed() = false, want true")
	}
}

// ---------------------------------------------------------------------------
// Test 2: CmZoom on already-zoomed window un-zooms it (toggle)
// ---------------------------------------------------------------------------

// TestWindowCmZoom_Unzooms_WhenAlreadyZoomed verifies the second CmZoom
// toggles zoom off — the zoom is a proper toggle, not a one-way flag.
//
// Spec: "CmZoom: call w.Zoom() — Zoom() is a toggle."
func TestWindowCmZoom_Unzooms_WhenAlreadyZoomed(t *testing.T) {
	w, _ := newZoomTestWindowWithOwner()
	w.Zoom() // zoom in first

	if !w.IsZoomed() {
		t.Fatal("precondition: window must be zoomed before this test")
	}

	w.HandleEvent(cmZoomEvent())

	if w.IsZoomed() {
		t.Error("after second CmZoom on zoomed window: IsZoomed() = true, want false")
	}
}

// ---------------------------------------------------------------------------
// Test 3: CmZoom event is cleared (consumed)
// ---------------------------------------------------------------------------

// TestWindowCmZoom_EventIsCleared verifies that after CmZoom is handled the
// event is cleared (What == EvNothing) so it does not propagate further.
//
// Spec: "CmZoom: consume event (call event.Clear())."
func TestWindowCmZoom_EventIsCleared(t *testing.T) {
	w := newZoomTestWindow()
	ev := cmZoomEvent()

	w.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmZoom: event.What = %v after HandleEvent, want EvNothing (cleared)", ev.What)
	}
}

// ---------------------------------------------------------------------------
// Test 4: CmResize event is cleared (consumed)
// ---------------------------------------------------------------------------

// TestWindowCmResize_EventIsCleared verifies that CmResize is consumed by
// Window.HandleEvent — the event is cleared and does not reach the group.
//
// Spec: "CmResize: consume event."
func TestWindowCmResize_EventIsCleared(t *testing.T) {
	w := newZoomTestWindow()
	ev := cmResizeEvent()

	w.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmResize: event.What = %v after HandleEvent, want EvNothing (cleared)", ev.What)
	}
}

// ---------------------------------------------------------------------------
// Test 5: CmResize does not change window bounds
// ---------------------------------------------------------------------------

// TestWindowCmResize_BoundsUnchanged verifies that handling CmResize does not
// move or resize the window — keyboard-driven resize is deferred.
//
// Spec: "CmResize: consume event (keyboard-driven resize deferred)."
func TestWindowCmResize_BoundsUnchanged(t *testing.T) {
	original := NewRect(5, 3, 20, 10)
	w := NewWindow(original, "ResizeTest")

	w.HandleEvent(cmResizeEvent())

	if w.Bounds() != original {
		t.Errorf("CmResize changed bounds: got %v, want %v (unchanged)", w.Bounds(), original)
	}
}

// ---------------------------------------------------------------------------
// Test 6: Modal CmClose→CmCancel still works (regression)
// ---------------------------------------------------------------------------

// TestWindowCmZoom_ModalCmCloseRegressionStillConverts verifies that adding
// CmZoom/CmResize handling does not break the existing modal-window behaviour:
// a modal window must still convert CmClose to CmCancel.
//
// Spec: "Modal window: CmClose → CmCancel — must be preserved."
func TestWindowCmZoom_ModalCmCloseRegressionStillConverts(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Modal")
	w.SetState(SfModal, true)
	ev := &Event{What: EvCommand, Command: CmClose}

	w.HandleEvent(ev)

	if ev.Command != CmCancel {
		t.Errorf("modal CmClose regression: Command = %v, want CmCancel", ev.Command)
	}
	// The spec says it does NOT clear — the event remains EvCommand so the
	// converted CmCancel can still propagate.
	if ev.IsCleared() {
		t.Error("modal CmClose regression: event should NOT be cleared (CmCancel must propagate)")
	}
}

// ---------------------------------------------------------------------------
// Test 7: Non-modal window passes CmClose through to group (regression)
// ---------------------------------------------------------------------------

// TestWindowCmZoom_NonModalCmClosePassesThrough verifies that a non-modal
// window does not transform CmClose — it reaches the group unchanged.
//
// Spec: "Non-modal window: CmClose must not be transformed."
func TestWindowCmZoom_NonModalCmClosePassesThrough(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Plain")
	child := newSelectableMockView(NewRect(0, 0, 18, 8))
	w.Insert(child)

	ev := &Event{What: EvCommand, Command: CmClose}
	w.HandleEvent(ev)

	// The group received it, so the child should have seen the event.
	if child.eventHandled == nil {
		t.Fatal("non-modal CmClose regression: group child did not receive event at all")
	}
	if child.eventHandled.Command != CmClose {
		t.Errorf("non-modal CmClose regression: child received Command=%v, want CmClose",
			child.eventHandled.Command)
	}
}

// ---------------------------------------------------------------------------
// Test 8: Unknown commands still reach the group
// ---------------------------------------------------------------------------

// TestWindowCmZoom_UnknownCommandsReachGroup verifies that CmZoom/CmResize
// handling does not silently swallow unrelated commands — they must still
// be dispatched to the group.
//
// Spec: "CmZoom/CmResize handling must be specific; other commands reach group."
func TestWindowCmZoom_UnknownCommandsReachGroup(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "PassThrough")
	child := newSelectableMockView(NewRect(0, 0, 18, 8))
	w.Insert(child)

	ev := &Event{What: EvCommand, Command: CmOK}
	w.HandleEvent(ev)

	if child.eventHandled == nil {
		t.Fatal("CmOK should have reached the group child, but child.eventHandled is nil")
	}
	if child.eventHandled.Command != CmOK {
		t.Errorf("CmOK passthrough: child received Command=%v, want CmOK",
			child.eventHandled.Command)
	}
}

// ---------------------------------------------------------------------------
// Test 9 (Falsification): CmZoom actually changes IsZoomed() — not just
// clearing the event
// ---------------------------------------------------------------------------

// TestWindowCmZoom_Falsify_ZoomStateChanges is a falsification test: it
// verifies that CmZoom changes IsZoomed() — an implementation that only clears
// the event without calling Zoom() would fail this test.
//
// Spec: "CmZoom must call w.Zoom(), not just consume the event."
func TestWindowCmZoom_Falsify_ZoomStateChanges(t *testing.T) {
	w, _ := newZoomTestWindowWithOwner()
	before := w.IsZoomed()

	w.HandleEvent(cmZoomEvent())

	after := w.IsZoomed()
	if before == after {
		t.Errorf("CmZoom did not change IsZoomed(): was %v, still %v — Zoom() was not called",
			before, after)
	}
}

// ---------------------------------------------------------------------------
// Test 10 (Falsification): Double CmZoom returns to original bounds
// ---------------------------------------------------------------------------

// TestWindowCmZoom_Falsify_DoubleCmZoomRestoresBounds verifies that two
// CmZoom commands restore the window to its original bounds — confirming that
// Zoom() is a genuine toggle, not a one-way state setter.
//
// Spec: "w.Zoom() stores bounds on zoom-in and restores them on zoom-out."
func TestWindowCmZoom_Falsify_DoubleCmZoomRestoresBounds(t *testing.T) {
	original := NewRect(5, 3, 20, 10)
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(original, "BoundsRestore")
	d.Insert(w)

	// First CmZoom: zoom in (fills owner area, bounds change).
	w.HandleEvent(cmZoomEvent())

	afterZoomIn := w.Bounds()
	if afterZoomIn == original {
		// If bounds didn't change (e.g. no owner propagated), warn but don't
		// fail — the important assertion is the round-trip below.
		t.Logf("note: zoom-in did not change bounds (owner area may equal original); proceeding")
	}

	// Second CmZoom: zoom out (restores original bounds).
	w.HandleEvent(cmZoomEvent())

	afterZoomOut := w.Bounds()
	if afterZoomOut != original {
		t.Errorf("double CmZoom did not restore original bounds: got %v, want %v",
			afterZoomOut, original)
	}

	// Confirm the window is no longer in the zoomed state.
	if w.IsZoomed() {
		t.Error("after double CmZoom: IsZoomed() = true, want false")
	}
}
