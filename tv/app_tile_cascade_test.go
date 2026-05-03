package tv

// app_tile_cascade_test.go — Tests for Phase 14 Task 4:
// Moving CmTile and CmCascade handling from Desktop.HandleEvent to
// Application.handleCommand.
//
// Architectural requirement (spec 13.1):
//   - Application.handleCommand handles CmTile → calls app.desktop.Tile(), clears event
//   - Application.handleCommand handles CmCascade → calls app.desktop.Cascade(), clears event
//   - Desktop.HandleEvent NO LONGER handles CmTile or CmCascade
//   - Desktop still handles CmNext and CmPrev (regression tests)
//   - The tile and cascade functionality itself is unchanged
//
// All tests go through app.handleEvent (unexported, accessible in package tv).

import "testing"

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// appWithWindows creates an Application backed by a SimulationScreen and
// inserts n windows into its desktop. Each window uses a fixed initial bounds
// that guarantees all windows overlap, so any tiling/cascade layout will be
// detectable.
//
// The screen is 80×24. The desktop gets the full area (no menu bar or status
// line is attached).
func appWithWindows(t *testing.T, n int) (*Application, []*Window) {
	t.Helper()
	s := newTestScreen(t)
	s.SetSize(80, 24)
	app, err := NewApplication(WithScreen(s))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	windows := make([]*Window, n)
	for i := 0; i < n; i++ {
		// All windows start at the same position so we can detect movement.
		w := NewWindow(NewRect(0, 0, 10, 5), "W")
		app.Desktop().Insert(w)
		windows[i] = w
	}
	return app, windows
}

// appCmd builds an EvCommand event.
func appCmd(cmd CommandCode) *Event {
	return &Event{What: EvCommand, Command: cmd}
}

// ---------------------------------------------------------------------------
// CmTile via Application.handleEvent
// ---------------------------------------------------------------------------

// Spec: "Application.handleCommand handles CmTile: calls app.desktop.Tile(),
// consumes event."
//
// Send CmTile through app.handleEvent and verify that two windows end up with
// distinct, non-overlapping origins — proof that Tile() was called.
func TestAppTileCascade_CmTileTilesWindows(t *testing.T) {
	app, windows := appWithWindows(t, 2)
	w1, w2 := windows[0], windows[1]

	ev := appCmd(CmTile)
	app.handleEvent(ev)

	// After tiling, windows must have different origins.
	if w1.Bounds().A == w2.Bounds().A {
		t.Errorf("CmTile via handleEvent: both windows still at same origin %v; Tile() was not called", w1.Bounds().A)
	}
}

// Spec: event is consumed (cleared) after CmTile is handled by Application.
func TestAppTileCascade_CmTileEventConsumed(t *testing.T) {
	app, _ := appWithWindows(t, 1)

	ev := appCmd(CmTile)
	app.handleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmTile via handleEvent: event not cleared (IsCleared=%v)", ev.IsCleared())
	}
}

// Falsification: CmTile actually changes window bounds, not merely clearing the
// event. Verify the size is consistent with a tiled layout (one window fills the
// desktop when n=1).
func TestAppTileCascade_CmTileActuallyChangesBounds(t *testing.T) {
	s := newTestScreen(t)
	s.SetSize(80, 24)
	app, err := NewApplication(WithScreen(s))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Place the window far from (0,0) with small dimensions.
	w := NewWindow(NewRect(10, 10, 10, 5), "W")
	app.Desktop().Insert(w)

	original := w.Bounds()

	ev := appCmd(CmTile)
	app.handleEvent(ev)

	if w.Bounds() == original {
		t.Errorf("CmTile via handleEvent: window bounds unchanged (%v); Tile() had no effect", original)
	}
	// With 1 window, Tile() fills the desktop. Desktop is 80×24 (no menu/status).
	got := w.Bounds()
	if got.A.X != 0 || got.A.Y != 0 || got.Width() != 80 || got.Height() != 24 {
		t.Errorf("CmTile via handleEvent: 1-window tile bounds = %v, want (0,0,80,24)", got)
	}
}

// ---------------------------------------------------------------------------
// CmCascade via Application.handleEvent
// ---------------------------------------------------------------------------

// Spec: "Application.handleCommand handles CmCascade: calls app.desktop.Cascade(),
// consumes event."
//
// Send CmCascade through app.handleEvent and verify that two windows end up with
// the expected cascade offsets — proof that Cascade() was called.
func TestAppTileCascade_CmCascadeCascadesWindows(t *testing.T) {
	app, windows := appWithWindows(t, 2)
	w1, w2 := windows[0], windows[1]

	ev := appCmd(CmCascade)
	app.handleEvent(ev)

	// After cascading, windows must have been repositioned. With 2 windows on
	// an 80×24 desktop, Cascade() places w1 at (0,0) and w2 at (2,1).
	b1, b2 := w1.Bounds(), w2.Bounds()
	if b1.A.X != 0 || b1.A.Y != 0 {
		t.Errorf("CmCascade via handleEvent: w1 origin = (%d,%d), want (0,0)", b1.A.X, b1.A.Y)
	}
	if b2.A.X != 2 || b2.A.Y != 1 {
		t.Errorf("CmCascade via handleEvent: w2 origin = (%d,%d), want (2,1)", b2.A.X, b2.A.Y)
	}
}

// Spec: event is consumed (cleared) after CmCascade is handled by Application.
func TestAppTileCascade_CmCascadeEventConsumed(t *testing.T) {
	app, _ := appWithWindows(t, 1)

	ev := appCmd(CmCascade)
	app.handleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmCascade via handleEvent: event not cleared (IsCleared=%v)", ev.IsCleared())
	}
}

// Falsification: CmCascade actually resizes windows, not just consuming the event.
// With 1 window on an 80×24 desktop, Cascade() sets size to 3/4 of desktop = 60×18.
func TestAppTileCascade_CmCascadeActuallyChangesBounds(t *testing.T) {
	s := newTestScreen(t)
	s.SetSize(80, 24)
	app, err := NewApplication(WithScreen(s))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	w := NewWindow(NewRect(5, 5, 10, 5), "W")
	app.Desktop().Insert(w)
	original := w.Bounds()

	ev := appCmd(CmCascade)
	app.handleEvent(ev)

	if w.Bounds() == original {
		t.Errorf("CmCascade via handleEvent: window bounds unchanged (%v); Cascade() had no effect", original)
	}
	// With 1 window, 80*3/4=60 wide, 24*3/4=18 tall, at origin (0,0).
	got := w.Bounds()
	wantW, wantH := 80*3/4, 24*3/4
	if got.Width() != wantW || got.Height() != wantH {
		t.Errorf("CmCascade via handleEvent: 1-window cascade size = (%d,%d), want (%d,%d)", got.Width(), got.Height(), wantW, wantH)
	}
}

// ---------------------------------------------------------------------------
// Desktop does NOT handle CmTile directly (post-move)
// ---------------------------------------------------------------------------

// Spec: "Desktop.HandleEvent NO LONGER handles CmTile."
//
// Send CmTile directly to Desktop.HandleEvent with a real window present.
// The event must pass through to the group UNCONSUMED — Desktop should not
// clear it.
func TestAppTileCascade_DesktopDoesNotHandleCmTile(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(0, 0, 10, 5), "W")
	d.Insert(w)

	ev := appCmd(CmTile)
	d.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("CmTile sent directly to Desktop was consumed; Desktop should no longer handle CmTile")
	}
}

// Additional confirmation: when Desktop does not handle CmTile, the window bounds
// must remain unchanged (since nobody else in the desktop chain handles it).
func TestAppTileCascade_DesktopDoesNotTileOnCmTile(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 5, 10, 5), "W")
	d.Insert(w)

	original := w.Bounds()

	ev := appCmd(CmTile)
	d.HandleEvent(ev)

	if w.Bounds() != original {
		t.Errorf("CmTile sent directly to Desktop changed window bounds from %v to %v; Desktop should not tile", original, w.Bounds())
	}
}

// ---------------------------------------------------------------------------
// Desktop does NOT handle CmCascade directly (post-move)
// ---------------------------------------------------------------------------

// Spec: "Desktop.HandleEvent NO LONGER handles CmCascade."
//
// Send CmCascade directly to Desktop.HandleEvent. The event must pass through
// UNCONSUMED — Desktop should not clear it.
func TestAppTileCascade_DesktopDoesNotHandleCmCascade(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(0, 0, 10, 5), "W")
	d.Insert(w)

	ev := appCmd(CmCascade)
	d.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("CmCascade sent directly to Desktop was consumed; Desktop should no longer handle CmCascade")
	}
}

// Additional confirmation: when Desktop does not handle CmCascade, the window
// bounds remain unchanged.
func TestAppTileCascade_DesktopDoesNotCascadeOnCmCascade(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(5, 5, 10, 5), "W")
	d.Insert(w)

	original := w.Bounds()

	ev := appCmd(CmCascade)
	d.HandleEvent(ev)

	if w.Bounds() != original {
		t.Errorf("CmCascade sent directly to Desktop changed window bounds from %v to %v; Desktop should not cascade", original, w.Bounds())
	}
}

// ---------------------------------------------------------------------------
// Regression: Desktop still handles CmNext and CmPrev
// ---------------------------------------------------------------------------

// Spec: "Desktop still handles CmNext and CmPrev."
//
// Regression: moving CmTile/CmCascade must not accidentally remove CmNext handling.
func TestAppTileCascade_DesktopStillHandlesCmNext(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	d.Insert(w1)
	d.Insert(w2)
	// After inserts: w2 is last (focused). CmNext from w2 wraps to w1.

	ev := appCmd(CmNext)
	d.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmNext regression: Desktop no longer clears CmNext event")
	}
	if d.FocusedChild() != w1 {
		t.Errorf("CmNext regression: FocusedChild() = %v, want w1 (SelectNextWindow)", d.FocusedChild())
	}
}

// Spec: "Desktop still handles CmNext and CmPrev."
//
// Regression: moving CmTile/CmCascade must not accidentally remove CmPrev handling.
func TestAppTileCascade_DesktopStillHandlesCmPrev(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	w3 := NewWindow(NewRect(40, 0, 20, 10), "W3")
	d.Insert(w1)
	d.Insert(w2)
	d.Insert(w3)
	// After inserts: w3 is last (focused). CmPrev from w3 moves to w2.

	ev := appCmd(CmPrev)
	d.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmPrev regression: Desktop no longer clears CmPrev event")
	}
	if d.FocusedChild() != w2 {
		t.Errorf("CmPrev regression: FocusedChild() = %v, want w2 (SelectPrevWindow)", d.FocusedChild())
	}
}
