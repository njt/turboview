package tv

// integration_phase14_batch2_test.go — Integration tests for Phase 14 Batch 2:
// architectural corrections: CmTile/CmCascade live in Application, Alt+N lives
// in Application (via CmSelectWindowNum broadcast), and Tab moves focus at the
// Window/Dialog level rather than inside Group or CheckBoxes.
//
// These tests exercise the full event pipeline end-to-end — real Application,
// real Desktop, real Window/Dialog/CheckBoxes — to verify that the three moved
// behaviours work correctly together in realistic scenarios.
//
// Features under test (per spec):
//   13.1 (CmTile/CmCascade): Application.handleCommand tiles / cascades.
//   13.2 (Alt+N):            Application.handleEvent broadcasts CmSelectWindowNum.
//   13.3 (Tab):              Window and Dialog cycle focus; CheckBoxes does NOT.
//
// Test naming: TestIntegrationPhase14Batch2_<DescriptiveSuffix>

import "testing"

// ---------------------------------------------------------------------------
// Local helpers
// ---------------------------------------------------------------------------

// newAppTwoWindows builds an Application backed by a simulation screen and
// inserts two numbered, non-overlapping Windows into its Desktop.
// The desktop is 80×24 (no menu bar, no status line).
func newAppTwoWindows(t *testing.T) (*Application, *Window, *Window) {
	t.Helper()
	s := newTestScreen(t)
	s.SetSize(80, 24)
	t.Cleanup(func() { s.Fini() })

	app, err := NewApplication(WithScreen(s))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	w1 := NewWindow(NewRect(0, 0, 10, 5), "W1", WithWindowNumber(1))
	w2 := NewWindow(NewRect(0, 0, 10, 5), "W2", WithWindowNumber(2))
	app.Desktop().Insert(w1)
	app.Desktop().Insert(w2)
	return app, w1, w2
}

// ---------------------------------------------------------------------------
// Scenario 1: CmTile reaches Application and tiles windows
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch2_CmTileTilesWindowsViaApp verifies that sending a
// CmTile command event through Application.handleEvent tiles the desktop windows,
// giving them distinct, non-overlapping origins.
//
// This confirms that Application.handleCommand (not Desktop.HandleEvent) now owns
// CmTile.
func TestIntegrationPhase14Batch2_CmTileTilesWindowsViaApp(t *testing.T) {
	app, w1, w2 := newAppTwoWindows(t)

	// Both windows start at the same origin — any tile operation must separate them.
	if w1.Bounds().A != w2.Bounds().A {
		t.Logf("note: windows already at different origins before CmTile")
	}

	ev := &Event{What: EvCommand, Command: CmTile}
	app.handleEvent(ev)

	b1, b2 := w1.Bounds(), w2.Bounds()
	if b1.A == b2.A {
		t.Errorf("CmTile via Application.handleEvent: both windows still at same origin %v; Tile() was not called", b1.A)
	}
}

// TestIntegrationPhase14Batch2_CmTileEventConsumedByApp verifies that the CmTile
// event is cleared (consumed) by Application.handleCommand.
func TestIntegrationPhase14Batch2_CmTileEventConsumedByApp(t *testing.T) {
	app, _, _ := newAppTwoWindows(t)

	ev := &Event{What: EvCommand, Command: CmTile}
	app.handleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmTile via Application: event not cleared (IsCleared = false)")
	}
}

// TestIntegrationPhase14Batch2_CmTileWindowsHaveDifferentSizes verifies that
// tiling two windows on an 80×24 desktop assigns them side-by-side cells that
// together fill the full desktop area, giving each distinct dimensions.
func TestIntegrationPhase14Batch2_CmTileWindowsHaveDifferentPositions(t *testing.T) {
	app, w1, w2 := newAppTwoWindows(t)

	ev := &Event{What: EvCommand, Command: CmTile}
	app.handleEvent(ev)

	// With 2 windows on an 80×24 desktop, Tile() creates 2 columns.
	// w1 gets col=0, w2 gets col=1 — their X origins must differ.
	if w1.Bounds().A.X == w2.Bounds().A.X {
		t.Errorf("CmTile: w1.X=%d == w2.X=%d; windows should be in separate columns",
			w1.Bounds().A.X, w2.Bounds().A.X)
	}
}

// ---------------------------------------------------------------------------
// Scenario 2: CmCascade reaches Application and cascades windows
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch2_CmCascadeCascadesWindowsViaApp verifies that
// sending a CmCascade event through Application.handleEvent applies cascade
// offsets to the desktop windows.
//
// On an 80×24 desktop with 2 windows, Cascade() places w1 at (0,0) and w2 at (2,1).
func TestIntegrationPhase14Batch2_CmCascadeCascadesWindowsViaApp(t *testing.T) {
	app, w1, w2 := newAppTwoWindows(t)

	ev := &Event{What: EvCommand, Command: CmCascade}
	app.handleEvent(ev)

	b1, b2 := w1.Bounds(), w2.Bounds()
	if b1.A.X != 0 || b1.A.Y != 0 {
		t.Errorf("CmCascade: w1 origin = (%d,%d), want (0,0)", b1.A.X, b1.A.Y)
	}
	if b2.A.X != 2 || b2.A.Y != 1 {
		t.Errorf("CmCascade: w2 origin = (%d,%d), want (2,1)", b2.A.X, b2.A.Y)
	}
}

// TestIntegrationPhase14Batch2_CmCascadeEventConsumedByApp verifies that the
// CmCascade event is cleared after Application.handleCommand processes it.
func TestIntegrationPhase14Batch2_CmCascadeEventConsumedByApp(t *testing.T) {
	app, _, _ := newAppTwoWindows(t)

	ev := &Event{What: EvCommand, Command: CmCascade}
	app.handleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmCascade via Application: event not cleared (IsCleared = false)")
	}
}

// ---------------------------------------------------------------------------
// Scenario 3: Alt+1 broadcasts CmSelectWindowNum and focuses the correct window
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch2_AltOneSelectsWindow1 verifies that an Alt+1
// keyboard event dispatched through Application.handleEvent focuses the window
// numbered 1, even when window 3 is currently focused.
//
// This exercises the full path: Application.handleEvent → broadcast
// CmSelectWindowNum → Desktop group dispatch → Window.HandleEvent (matching).
func TestIntegrationPhase14Batch2_AltOneSelectsWindow1(t *testing.T) {
	s := newTestScreen(t)
	s.SetSize(80, 24)
	t.Cleanup(func() { s.Fini() })

	app, err := NewApplication(WithScreen(s))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1", WithWindowNumber(1))
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2", WithWindowNumber(2))
	w3 := NewWindow(NewRect(40, 0, 20, 10), "W3", WithWindowNumber(3))
	app.Desktop().Insert(w1)
	app.Desktop().Insert(w2)
	app.Desktop().Insert(w3)

	// Focus w3 so we can verify a real focus change.
	app.Desktop().BringToFront(w3)
	if app.Desktop().FocusedChild() != w3 {
		t.Fatalf("precondition: FocusedChild() = %v, want w3", app.Desktop().FocusedChild())
	}

	ev := appAltNKeyEvent('1')
	app.handleEvent(ev)

	if app.Desktop().FocusedChild() != w1 {
		t.Errorf("Alt+1 via Application: FocusedChild() = %v, want w1 (number 1)",
			app.Desktop().FocusedChild())
	}
}

// TestIntegrationPhase14Batch2_AltOneKeyboardEventConsumed verifies that the
// original Alt+1 keyboard event is cleared when a matching window is found.
func TestIntegrationPhase14Batch2_AltOneKeyboardEventConsumed(t *testing.T) {
	s := newTestScreen(t)
	s.SetSize(80, 24)
	t.Cleanup(func() { s.Fini() })

	app, err := NewApplication(WithScreen(s))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1", WithWindowNumber(1))
	app.Desktop().Insert(w1)

	ev := appAltNKeyEvent('1')
	app.handleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Alt+1 via Application: keyboard event not cleared after matching window focused")
	}
}

// TestIntegrationPhase14Batch2_AltNNoMatchLeavesEventUncleared verifies that an
// Alt+N event where no window has the given number is NOT consumed.
func TestIntegrationPhase14Batch2_AltNNoMatchLeavesEventUncleared(t *testing.T) {
	s := newTestScreen(t)
	s.SetSize(80, 24)
	t.Cleanup(func() { s.Fini() })

	app, err := NewApplication(WithScreen(s))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Only window number 1 exists; Alt+9 should not match anything.
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1", WithWindowNumber(1))
	app.Desktop().Insert(w1)

	ev := appAltNKeyEvent('9')
	app.handleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("Alt+9 via Application: event was cleared even though no window number 9 exists")
	}
}

// ---------------------------------------------------------------------------
// Scenario 4: Tab in a Window cycles focus between child widgets
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch2_TabInWindowCyclesFocus verifies that sending a
// Tab keyboard event to a Window with three selectable children advances focus
// from the first child to the second.
func TestIntegrationPhase14Batch2_TabInWindowCyclesFocus(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 15), "W")

	v1 := newSelectableMockView(NewRect(0, 0, 10, 3))
	v2 := newSelectableMockView(NewRect(0, 3, 10, 3))
	v3 := newSelectableMockView(NewRect(0, 6, 10, 3))
	w.Insert(v1)
	w.Insert(v2)
	w.Insert(v3)

	// After the last insert, v3 is focused. Bring v1 to front to make state predictable.
	w.SetFocusedChild(v1)
	if w.FocusedChild() != v1 {
		t.Fatalf("precondition: FocusedChild() = %v, want v1", w.FocusedChild())
	}

	w.HandleEvent(tabEvent())

	if w.FocusedChild() != v2 {
		t.Errorf("Tab in Window: FocusedChild() = %v, want v2 (next after v1)", w.FocusedChild())
	}
}

// TestIntegrationPhase14Batch2_TabInWindowEventConsumed verifies that Tab is
// cleared (consumed) after Window.HandleEvent processes it.
func TestIntegrationPhase14Batch2_TabInWindowEventConsumed(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 15), "W")
	v1 := newSelectableMockView(NewRect(0, 0, 10, 3))
	v2 := newSelectableMockView(NewRect(0, 3, 10, 3))
	w.Insert(v1)
	w.Insert(v2)
	w.SetFocusedChild(v1)

	ev := tabEvent()
	w.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Tab in Window: event not cleared (should be consumed by Window.HandleEvent)")
	}
}

// TestIntegrationPhase14Batch2_TabInWindowWrapsAround verifies that Tab from
// the last child in a Window wraps focus back to the first child.
func TestIntegrationPhase14Batch2_TabInWindowWrapsAround(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 15), "W")
	v1 := newSelectableMockView(NewRect(0, 0, 10, 3))
	v2 := newSelectableMockView(NewRect(0, 3, 10, 3))
	w.Insert(v1)
	w.Insert(v2)
	// v2 is last, so focus it and verify Tab wraps to v1.
	w.SetFocusedChild(v2)
	if w.FocusedChild() != v2 {
		t.Fatalf("precondition: FocusedChild() = %v, want v2", w.FocusedChild())
	}

	w.HandleEvent(tabEvent())

	if w.FocusedChild() != v1 {
		t.Errorf("Tab from last child in Window: FocusedChild() = %v, want v1 (wrap-around)", w.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Scenario 5: Tab in a Dialog cycles focus between child widgets
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch2_TabInDialogCyclesFocus verifies that sending a
// Tab event to a Dialog advances focus from the first child to the second.
func TestIntegrationPhase14Batch2_TabInDialogCyclesFocus(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "D")

	v1 := newSelectableMockView(NewRect(0, 0, 10, 3))
	v2 := newSelectableMockView(NewRect(0, 3, 10, 3))
	d.Insert(v1)
	d.Insert(v2)
	d.SetFocusedChild(v1)
	if d.FocusedChild() != v1 {
		t.Fatalf("precondition: FocusedChild() = %v, want v1", d.FocusedChild())
	}

	d.HandleEvent(tabEvent())

	if d.FocusedChild() != v2 {
		t.Errorf("Tab in Dialog: FocusedChild() = %v, want v2 (next after v1)", d.FocusedChild())
	}
}

// TestIntegrationPhase14Batch2_TabInDialogEventConsumed verifies that Tab is
// consumed by Dialog.HandleEvent.
func TestIntegrationPhase14Batch2_TabInDialogEventConsumed(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "D")
	v1 := newSelectableMockView(NewRect(0, 0, 10, 3))
	v2 := newSelectableMockView(NewRect(0, 3, 10, 3))
	d.Insert(v1)
	d.Insert(v2)
	d.SetFocusedChild(v1)

	ev := tabEvent()
	d.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Tab in Dialog: event not cleared (should be consumed by Dialog.HandleEvent)")
	}
}

// ---------------------------------------------------------------------------
// Scenario 6: Tab does NOT cycle within a CheckBoxes cluster
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch2_TabInCheckBoxesClusterDoesNotCycleItems verifies
// that sending Tab directly to a CheckBoxes cluster does not change the focused
// item within the cluster.
//
// This is the key spec-13.3 requirement: CheckBoxes uses Up/Down for internal
// navigation; Tab is intentionally not consumed, so it propagates to the parent
// (Window or Dialog) for inter-widget focus traversal.
func TestIntegrationPhase14Batch2_TabInCheckBoxesClusterDoesNotCycleItems(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})

	// After construction the last-inserted item is focused. Explicitly set Item(0)
	// as the focused item so the test has a deterministic starting state.
	item0 := cbs.Item(0)
	cbs.SetFocusedChild(item0)
	if cbs.FocusedChild() != item0 {
		t.Fatalf("precondition: FocusedChild() = %v, want Item(0)", cbs.FocusedChild())
	}

	ev := tabEvent()
	cbs.HandleEvent(ev)

	// Tab must NOT shift focus from item0 to item1 within the cluster.
	if cbs.FocusedChild() != item0 {
		t.Errorf("Tab in CheckBoxes: FocusedChild() = %v, want Item(0) (Tab must not cycle within cluster)",
			cbs.FocusedChild())
	}
}

// TestIntegrationPhase14Batch2_TabInCheckBoxesNotConsumedByCluster verifies that
// Tab is not consumed by the CheckBoxes cluster itself, so it can propagate to
// the enclosing Dialog or Window for inter-widget traversal.
func TestIntegrationPhase14Batch2_TabInCheckBoxesNotConsumedByCluster(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})

	ev := tabEvent()
	cbs.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("Tab in CheckBoxes: event was cleared by the cluster; it should propagate to the parent")
	}
}

// TestIntegrationPhase14Batch2_TabViaDialogSkipsCheckBoxesClusterInternals verifies
// that Tab dispatched at the Dialog level moves focus from the CheckBoxes cluster
// to the next sibling widget — not to a CheckBox item inside the cluster.
func TestIntegrationPhase14Batch2_TabViaDialogSkipsCheckBoxesClusterInternals(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "D")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})
	next := newSelectableMockView(NewRect(0, 4, 20, 3)) // sibling widget after the cluster

	d.Insert(cbs)
	d.Insert(next)

	// Focus the CheckBoxes cluster (it is selectable, so Insert focuses it already).
	d.SetFocusedChild(cbs)
	if d.FocusedChild() != cbs {
		t.Fatalf("precondition: FocusedChild() = %v, want cbs", d.FocusedChild())
	}

	d.HandleEvent(tabEvent())

	// After Tab, focus must move to 'next' (the sibling widget), not stay on cbs
	// or move to an internal CheckBox item.
	if d.FocusedChild() != next {
		t.Errorf("Tab from CheckBoxes via Dialog: FocusedChild() = %v, want next sibling widget",
			d.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Scenario 7: Full round-trip — Alt+N switches window, then Tab navigates within it
// ---------------------------------------------------------------------------

// TestIntegrationPhase14Batch2_AltNThenTabNavigatesWithinWindow verifies the
// combined round-trip: Alt+2 switches focus to window 2, then Tab advances focus
// between window 2's child widgets.
func TestIntegrationPhase14Batch2_AltNThenTabNavigatesWithinWindow(t *testing.T) {
	s := newTestScreen(t)
	s.SetSize(80, 24)
	t.Cleanup(func() { s.Fini() })

	app, err := NewApplication(WithScreen(s))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Window 1: focused initially.
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1", WithWindowNumber(1))
	// Window 2: has two selectable children.
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2", WithWindowNumber(2))
	v2a := newSelectableMockView(NewRect(0, 0, 10, 3))
	v2b := newSelectableMockView(NewRect(0, 3, 10, 3))
	w2.Insert(v2a)
	w2.Insert(v2b)
	// After inserts, v2b is focused inside w2.
	w2.SetFocusedChild(v2a) // set explicitly for determinism

	app.Desktop().Insert(w1)
	app.Desktop().Insert(w2)

	// Bring w1 to front so we can observe the switch.
	app.Desktop().BringToFront(w1)
	if app.Desktop().FocusedChild() != w1 {
		t.Fatalf("precondition: Desktop FocusedChild() = %v, want w1", app.Desktop().FocusedChild())
	}
	if w2.FocusedChild() != v2a {
		t.Fatalf("precondition: w2.FocusedChild() = %v, want v2a", w2.FocusedChild())
	}

	// Step 1: Alt+2 via Application.handleEvent — switches desktop focus to w2.
	app.handleEvent(appAltNKeyEvent('2'))

	if app.Desktop().FocusedChild() != w2 {
		t.Errorf("Alt+2: Desktop FocusedChild() = %v, want w2", app.Desktop().FocusedChild())
	}

	// Step 2: Tab via w2.HandleEvent — advances focus from v2a to v2b.
	w2.HandleEvent(tabEvent())

	if w2.FocusedChild() != v2b {
		t.Errorf("Tab in w2 after Alt+2: w2.FocusedChild() = %v, want v2b", w2.FocusedChild())
	}
}
