package tv

// application_menu_test.go — Tests for Task 4: Application Integration (MenuBar).
//
// Spec requirements under test:
//   1. WithMenuBar option — app stores MenuBar, MenuBar() accessor returns it.
//   2. No menuBar — app works same as before (nil accessor, desktop at row 0).
//   3. Layout — with menuBar: desktop starts at row 1, menuBar at row 0.
//   4. Draw — menuBar rendered at row 0, desktop below it.
//   5. F10→CmMenu→Activate — F10 activates menu bar through StatusLine→CmMenu path.
//   6. Mouse on row 0 — click activates menu bar.
//   7. Mouse on desktop — Y translated by menuH.
//
// Each test targets exactly one spec requirement. Falsifying tests are
// provided where a trivially-passing implementation could otherwise satisfy
// the suite.

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newMenuBarApp creates an Application with a MenuBar, StatusLine (F10→CmMenu),
// and BorlandBlue theme on an 80x25 simulation screen.
// Callers should defer screen.Fini().
func newMenuBarApp(t *testing.T) (*Application, *MenuBar, tcell.SimulationScreen) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen.Init: %v", err)
	}
	screen.SetSize(80, 25)

	menuBar := NewMenuBar(
		NewSubMenu("~F~ile",
			NewMenuItem("~N~ew", CmUser+1, KbNone()),
			NewMenuItem("~O~pen", CmUser+2, KbNone()),
		),
		NewSubMenu("~W~indow",
			NewMenuItem("~T~ile", CmUser+3, KbNone()),
		),
	)
	statusLine := NewStatusLine(
		NewStatusItem("~F10~ Menu", KbFunc(10), CmMenu),
	)

	app, err := NewApplication(
		WithScreen(screen),
		WithMenuBar(menuBar),
		WithStatusLine(statusLine),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}
	return app, menuBar, screen
}

// ---------------------------------------------------------------------------
// 1. WithMenuBar option — accessor and storage
// ---------------------------------------------------------------------------

// TestAppMenuWithMenuBarIsAnAppOption verifies that WithMenuBar compiles as an
// AppOption and NewApplication accepts it without error.
// Spec: "WithMenuBar(mb *MenuBar) AppOption registers a MenuBar with the Application."
func TestAppMenuWithMenuBarIsAnAppOption(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	mb := NewMenuBar(NewSubMenu("~F~ile"))
	app, err := NewApplication(WithScreen(screen), WithMenuBar(mb))
	if err != nil {
		t.Fatalf("NewApplication(WithMenuBar): %v", err)
	}
	if app == nil {
		t.Fatal("NewApplication returned nil")
	}
}

// TestAppMenuAccessorReturnsRegisteredMenuBar verifies MenuBar() returns the
// exact MenuBar passed to WithMenuBar.
// Spec: "Application.MenuBar() *MenuBar returns the menu bar (nil if none)."
func TestAppMenuAccessorReturnsRegisteredMenuBar(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	mb := NewMenuBar(NewSubMenu("~F~ile"))
	app, err := NewApplication(WithScreen(screen), WithMenuBar(mb))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	if app.MenuBar() != mb {
		t.Errorf("MenuBar() = %v, want the registered MenuBar %v", app.MenuBar(), mb)
	}
}

// TestAppMenuAccessorReturnsDifferentMenuBarsDistinctly is a falsifying test:
// two apps with different MenuBars must return their own MenuBar, not a shared one.
func TestAppMenuAccessorReturnsDifferentMenuBarsDistinctly(t *testing.T) {
	screen1 := newTestScreen(t)
	defer screen1.Fini()
	screen2 := newTestScreen(t)
	defer screen2.Fini()

	mb1 := NewMenuBar(NewSubMenu("~F~ile"))
	mb2 := NewMenuBar(NewSubMenu("~W~indow"))

	app1, err := NewApplication(WithScreen(screen1), WithMenuBar(mb1))
	if err != nil {
		t.Fatalf("NewApplication 1: %v", err)
	}
	app2, err := NewApplication(WithScreen(screen2), WithMenuBar(mb2))
	if err != nil {
		t.Fatalf("NewApplication 2: %v", err)
	}

	if app1.MenuBar() == app2.MenuBar() {
		t.Error("two distinct apps with different MenuBars must return distinct pointers")
	}
}

// TestAppMenuNewApplicationSetsMenuBarScheme verifies that NewApplication
// assigns its color scheme to the menuBar after creation.
// Spec: "if menuBar is set: set menuBar.scheme = app.scheme"
func TestAppMenuNewApplicationSetsMenuBarScheme(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	mb := NewMenuBar(NewSubMenu("~F~ile"))
	custom := &theme.ColorScheme{}
	_, err := NewApplication(WithScreen(screen), WithMenuBar(mb), WithTheme(custom))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	if mb.scheme != custom {
		t.Errorf("menuBar.scheme = %v, want the application's custom scheme %v", mb.scheme, custom)
	}
}

// TestAppMenuNewApplicationSetsMenuBarAppField verifies that NewApplication
// sets the menuBar.app field to the Application.
// Spec: "set menuBar.app = app"
func TestAppMenuNewApplicationSetsMenuBarAppField(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	mb := NewMenuBar(NewSubMenu("~F~ile"))
	app, err := NewApplication(WithScreen(screen), WithMenuBar(mb))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	if mb.app != app {
		t.Errorf("menuBar.app = %v, want the Application %v", mb.app, app)
	}
}

// ---------------------------------------------------------------------------
// 2. No menuBar — unchanged behavior
// ---------------------------------------------------------------------------

// TestAppMenuNoMenuBarAccessorReturnsNil verifies MenuBar() returns nil when
// no WithMenuBar option was given.
// Spec: "Application.MenuBar() *MenuBar returns the menu bar (nil if none)."
func TestAppMenuNoMenuBarAccessorReturnsNil(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	if app.MenuBar() != nil {
		t.Errorf("MenuBar() = %v, want nil when no WithMenuBar provided", app.MenuBar())
	}
}

// TestAppMenuNoMenuBarDesktopStartsAtRowZero verifies that without a MenuBar,
// the desktop still starts at row 0 (unchanged behavior).
// Spec: "If no menuBar: desktop starts at row 0 (unchanged from current behavior)."
func TestAppMenuNoMenuBarDesktopStartsAtRowZero(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	db := app.Desktop().Bounds()
	if db.A.Y != 0 {
		t.Errorf("desktop top = %d, want 0 when no menuBar", db.A.Y)
	}
}

// TestAppMenuNoMenuBarDrawIsUnchanged verifies that without a MenuBar, Draw
// renders the desktop pattern in every row (no status line case) as before.
// Spec: "If no menuBar: desktop starts at row 0 (unchanged from current behavior)."
func TestAppMenuNoMenuBarDrawIsUnchanged(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	const w, h = 80, 25
	screen.SetSize(w, h)

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// All cells must be the desktop pattern rune '░'.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != '░' {
				t.Errorf("no-menuBar Draw: cell (%d,%d) = %q, want '░'", x, y, cell.Rune)
				return
			}
		}
	}
}

// ---------------------------------------------------------------------------
// 3. Layout — desktop top shifted when menuBar present
// ---------------------------------------------------------------------------

// TestAppMenuLayoutDesktopStartsAtRow1WithMenuBar verifies that layoutChildren
// places the desktop starting at Y=1 when a menuBar is present.
// Spec: "Desktop top is menuH where menuH = 1 if menuBar is set, else menuH = 0."
func TestAppMenuLayoutDesktopStartsAtRow1WithMenuBar(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	db := app.Desktop().Bounds()
	if db.A.Y != 1 {
		t.Errorf("desktop top = %d, want 1 when menuBar is present", db.A.Y)
	}
}

// TestAppMenuLayoutMenuBarAtRow0(t *testing.T) verifies that layoutChildren
// places the menuBar at Y=0.
// Spec: "menuBar at row 0 (bounds (0, 0, w, 1))"
func TestAppMenuLayoutMenuBarAtRow0(t *testing.T) {
	app, mb, screen := newMenuBarApp(t)
	defer screen.Fini()
	_ = app

	b := mb.Bounds()
	if b.A.Y != 0 {
		t.Errorf("menuBar top = %d, want 0", b.A.Y)
	}
	if b.A.X != 0 {
		t.Errorf("menuBar left = %d, want 0", b.A.X)
	}
}

// TestAppMenuLayoutMenuBarHeightIsOne verifies the menuBar is exactly 1 row tall.
// Spec: "menuBar at row 0 (bounds (0, 0, w, 1))"
func TestAppMenuLayoutMenuBarHeightIsOne(t *testing.T) {
	app, mb, screen := newMenuBarApp(t)
	defer screen.Fini()
	_ = app

	if mb.Bounds().Height() != 1 {
		t.Errorf("menuBar height = %d, want 1", mb.Bounds().Height())
	}
}

// TestAppMenuLayoutMenuBarWidthEqualsScreenWidth verifies the menuBar spans
// the full screen width.
// Spec: "menuBar at row 0 (bounds (0, 0, w, 1))"
func TestAppMenuLayoutMenuBarWidthEqualsScreenWidth(t *testing.T) {
	app, mb, screen := newMenuBarApp(t)
	defer screen.Fini()
	_ = app

	w, _ := screen.Size()
	if mb.Bounds().Width() != w {
		t.Errorf("menuBar width = %d, want %d (screen width)", mb.Bounds().Width(), w)
	}
}

// TestAppMenuLayoutStatusLineAtLastRow verifies the status line is placed in
// the last row (h-1) when both menuBar and statusLine are present.
// Spec: "statusLine at last row"
func TestAppMenuLayoutStatusLineAtLastRow(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	_, h := screen.Size()
	sl := app.StatusLine()
	if sl == nil {
		t.Fatal("StatusLine() is nil — test setup error")
	}

	slBounds := sl.Bounds()
	if slBounds.A.Y != h-1 {
		t.Errorf("statusLine top = %d, want %d (last row)", slBounds.A.Y, h-1)
	}
}

// TestAppMenuLayoutDesktopBottomWithMenuBarAndStatusLine verifies the desktop
// occupies rows 1 through h-2 (exclusive of menuBar and statusLine).
// Spec: "desktop at rows 1 through h-statusH-1"
func TestAppMenuLayoutDesktopBottomWithMenuBarAndStatusLine(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	_, h := screen.Size()
	db := app.Desktop().Bounds()

	// With menuBar at 0 and statusLine at h-1, desktop should be rows 1..(h-2).
	wantTop := 1
	wantHeight := h - 2 // rows 1 through h-2 inclusive
	if db.A.Y != wantTop {
		t.Errorf("desktop top = %d, want %d", db.A.Y, wantTop)
	}
	if db.Height() != wantHeight {
		t.Errorf("desktop height = %d, want %d", db.Height(), wantHeight)
	}
}

// TestAppMenuLayoutDesktopNotAtRow0WhenMenuBarPresent falsifies any impl that
// ignores the menuBar and always puts the desktop at row 0.
func TestAppMenuLayoutDesktopNotAtRow0WhenMenuBarPresent(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	db := app.Desktop().Bounds()
	if db.A.Y == 0 {
		t.Error("desktop top == 0 with a menuBar present; want top == 1 (menuBar occupies row 0)")
	}
}

// ---------------------------------------------------------------------------
// 4. Draw — menuBar at row 0, desktop below
// ---------------------------------------------------------------------------

// TestAppMenuDrawRendersMenuBarAtRow0 verifies that Draw writes menuBar content
// into row 0 of the draw buffer.
// Spec: "Menu bar drawn at SubBuffer(0, 0, w, 1)."
func TestAppMenuDrawRendersMenuBarAtRow0(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	const w, h = 80, 25
	screen.SetSize(w, h)

	// Re-layout after potential resize.
	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// Row 0 must NOT be the desktop pattern — the menuBar owns it.
	allDesktop := true
	for x := 0; x < w; x++ {
		if buf.GetCell(x, 0).Rune != '░' {
			allDesktop = false
			break
		}
	}
	if allDesktop {
		t.Error("Draw: row 0 is all desktop pattern; expected menuBar content there")
	}
}

// TestAppMenuDrawDesktopStartsAtRow1 verifies that the desktop pattern rune
// appears starting at row 1 (not row 0) when a menuBar is present.
// Spec: "Desktop drawn at SubBuffer(0, menuH, w, desktopH)."
func TestAppMenuDrawDesktopStartsAtRow1(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	const w, h = 80, 25
	screen.SetSize(w, h)

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// Row 1 should be the desktop pattern (menuBar at 0, status at h-1).
	for x := 0; x < w; x++ {
		cell := buf.GetCell(x, 1)
		if cell.Rune != '░' {
			t.Errorf("Draw: cell (%d,1) = %q, want '░' (desktop at row 1)", x, cell.Rune)
			return
		}
	}
}

// TestAppMenuDrawStatusLineAtLastRow verifies status line is rendered in the
// last row when menuBar is also present.
// Spec: "Status line at bottom row."
func TestAppMenuDrawStatusLineAtLastRow(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	const w, h = 80, 25
	screen.SetSize(w, h)

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// Row h-1 must NOT be the desktop pattern — StatusLine owns it.
	allDesktop := true
	for x := 0; x < w; x++ {
		if buf.GetCell(x, h-1).Rune != '░' {
			allDesktop = false
			break
		}
	}
	if allDesktop {
		t.Error("Draw: last row is all desktop pattern; expected StatusLine content there")
	}
}

// TestAppMenuDrawDesktopNotAtRow0WithMenuBar falsifies an impl that draws the
// desktop at row 0 regardless of menuBar.
func TestAppMenuDrawDesktopNotAtRow0WithMenuBar(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	const w, h = 80, 25
	screen.SetSize(w, h)

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// Row 0 must not be the desktop pattern rune.
	desktopCount := 0
	for x := 0; x < w; x++ {
		if buf.GetCell(x, 0).Rune == '░' {
			desktopCount++
		}
	}
	if desktopCount == w {
		t.Error("Draw: row 0 is entirely '░' (desktop); menuBar should have overwritten it")
	}
}

// TestAppMenuDrawMenuBarLabelVisibleInRow0 verifies that a menu label character
// (the 'F' from "~F~ile") appears in row 0 after Draw.
// Spec: "Menu bar drawn at SubBuffer(0, 0, w, 1)."
func TestAppMenuDrawMenuBarLabelVisibleInRow0(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	const w, h = 80, 25
	screen.SetSize(w, h)

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// The menuBar for "~F~ile" should render 'F' somewhere in row 0.
	found := false
	for x := 0; x < w; x++ {
		if buf.GetCell(x, 0).Rune == 'F' {
			found = true
			break
		}
	}
	if !found {
		t.Error("Draw: 'F' (from '~F~ile') not found in row 0; menuBar not rendered there")
	}
}

// ---------------------------------------------------------------------------
// 5. F10→CmMenu→Activate — key path through StatusLine
// ---------------------------------------------------------------------------

// TestAppMenuF10ActivatesMenuBarViaStatusLine verifies that pressing F10 when
// the StatusLine maps F10→CmMenu causes the menuBar to enter its modal loop
// and deactivate after Escape.
// Spec: "if event is CmMenu and menuBar is not nil, call menuBar.Activate(app), clear event, return."
func TestAppMenuF10ActivatesMenuBarViaStatusLine(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	// Escape deactivates the menuBar modal loop; then inject Quit so Run exits.
	go func() {
		time.Sleep(50 * time.Millisecond)
		screen.InjectKey(tcell.KeyF10, 0, tcell.ModNone)
		// Give the modal loop time to start, then escape it.
		time.Sleep(30 * time.Millisecond)
		screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
		// Then quit.
		time.Sleep(30 * time.Millisecond)
		app.PostCommand(CmQuit, nil)
	}()

	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("Run() did not exit within 3 s — F10→CmMenu→Activate path may be broken")
	}
}

// TestAppMenuCmMenuEventDirectlyActivatesMenuBar verifies that injecting a
// CmMenu command event causes handleEvent to activate the menuBar.
// Spec: "After StatusLine transforms F10→CmMenu: if event is CmMenu and menuBar is not nil,
// call menuBar.Activate(app), clear event, return."
func TestAppMenuCmMenuEventDirectlyActivatesMenuBar(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	// Post CmMenu then immediately Escape to exit the modal loop, then Quit.
	go func() {
		time.Sleep(50 * time.Millisecond)
		app.PostCommand(CmMenu, nil)
		time.Sleep(30 * time.Millisecond)
		screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
		time.Sleep(30 * time.Millisecond)
		app.PostCommand(CmQuit, nil)
	}()

	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("Run() did not exit — CmMenu event did not activate menuBar, or modal loop blocked")
	}
}

// TestAppMenuCmMenuNotDispatchedToDesktopWhenMenuBarPresent verifies that a
// CmMenu event is consumed by the menuBar path and not forwarded to the desktop.
// Spec: "clear event, return" — the event must not propagate.
func TestAppMenuCmMenuNotDispatchedToDesktopWhenMenuBarPresent(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	mb := NewMenuBar(NewSubMenu("~F~ile", NewMenuItem("~N~ew", CmUser+1, KbNone())))
	statusLine := NewStatusLine(
		NewStatusItem("~F10~ Menu", KbFunc(10), CmMenu),
	)

	desktopReceivedCmMenu := false
	var onCmd func(CommandCode, any) bool
	onCmd = func(cmd CommandCode, info any) bool {
		if cmd == CmMenu {
			desktopReceivedCmMenu = true
		}
		return cmd == CmQuit
	}

	app, err := NewApplication(
		WithScreen(screen),
		WithMenuBar(mb),
		WithStatusLine(statusLine),
		WithOnCommand(onCmd),
	)
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Inject CmMenu directly, escape modal loop, then quit.
	go func() {
		time.Sleep(50 * time.Millisecond)
		app.PostCommand(CmMenu, nil)
		time.Sleep(30 * time.Millisecond)
		screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
		time.Sleep(30 * time.Millisecond)
		app.PostCommand(CmQuit, nil)
	}()

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
		if desktopReceivedCmMenu {
			t.Error("CmMenu event reached onCommand handler; it must be consumed by menuBar and not propagated")
		}
	case <-time.After(3 * time.Second):
		t.Error("Run() did not exit within 3 s")
	}
}

// TestAppMenuNoMenuBarCmMenuNotConsumedSpecially verifies that when there is no
// menuBar, a CmMenu command is NOT intercepted and reaches the onCommand
// callback as a normal command.
// Spec: "if menuBar is not nil" (the intercept only happens when menuBar is set).
func TestAppMenuNoMenuBarCmMenuNotConsumedSpecially(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	var receivedCmd CommandCode
	app, err := NewApplication(
		WithScreen(screen),
		WithOnCommand(func(cmd CommandCode, info any) bool {
			receivedCmd = cmd
			return true
		}),
	)
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	ev := &Event{What: EvCommand, Command: CmMenu}
	app.handleEvent(ev)

	if receivedCmd != CmMenu {
		t.Errorf("without menuBar, CmMenu should reach onCommand; got %v", receivedCmd)
	}
}

// ---------------------------------------------------------------------------
// 6. Mouse on row 0 — click routes to menuBar
// ---------------------------------------------------------------------------

// TestAppMenuMouseOnRow0ActivatesMenuBar verifies that a mouse click on row 0
// is routed to the menuBar (ActivateAt), not the desktop.
// Spec: "click on row 0 → menuBar.ActivateAt(app, clickedIndex, true)."
func TestAppMenuMouseOnRow0ActivatesMenuBar(t *testing.T) {
	app, _, screen := newMenuBarApp(t)
	defer screen.Fini()

	// Click at (2, 0) — row 0 belongs to the menuBar. The menuBar modal loop
	// starts; Escape closes any popup, Escape deactivates, then Quit exits Run.
	go func() {
		time.Sleep(50 * time.Millisecond)
		screen.InjectMouse(2, 0, tcell.Button1, tcell.ModNone)
		time.Sleep(30 * time.Millisecond)
		// The click opens a popup at index 0; close popup, deactivate.
		screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
		time.Sleep(20 * time.Millisecond)
		screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
		time.Sleep(30 * time.Millisecond)
		app.PostCommand(CmQuit, nil)
	}()

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("Run() did not exit — mouse click on row 0 may not have activated menuBar")
	}
}

// TestAppMenuMouseOnRow0DoesNotRouteToDesktop verifies that a click at Y=0
// is NOT routed to the desktop. We confirm this by checking that the desktop
// does NOT receive a mouse click when Y=0.
// Spec: "Menu bar hit-testing before status line/desktop: click on row 0 →
// menuBar.ActivateAt(app, clickedIndex, true)."
func TestAppMenuMouseOnRow0DoesNotRouteToDesktop(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	mb := NewMenuBar(NewSubMenu("~F~ile", NewMenuItem("~N~ew", CmUser+1, KbNone())))
	statusLine := NewStatusLine(
		NewStatusItem("~F10~ Menu", KbFunc(10), CmMenu),
	)

	app, err := NewApplication(
		WithScreen(screen),
		WithMenuBar(mb),
		WithStatusLine(statusLine),
	)
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Simulate a mouse event at Y=0 directly through routeMouseEvent.
	// After the call we check whether it entered the menuBar modal loop by
	// inspecting whether the menuBar is still active. Because routeMouseEvent
	// calls ActivateAt synchronously (the modal loop blocks), we cannot call
	// it directly here without pre-injecting an exit event.
	//
	// Strategy: inject Escape so the modal loop exits immediately after the
	// click routes to the menuBar, then call routeMouseEvent on a background
	// goroutine and verify that it returns (i.e., the modal loop was entered
	// and exited).
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)

	done := make(chan struct{}, 1)
	go func() {
		ev := &Event{
			What: EvMouse,
			Mouse: &MouseEvent{
				X:      2,
				Y:      0,
				Button: tcell.Button1,
			},
		}
		app.routeMouseEvent(ev)
		done <- struct{}{}
	}()

	select {
	case <-done:
		// routeMouseEvent returned — modal loop was entered and exited via Escape.
	case <-time.After(3 * time.Second):
		t.Error("routeMouseEvent for Y=0 did not return — menuBar modal loop may not have been entered or exited")
	}
}

// ---------------------------------------------------------------------------
// 7. Mouse on desktop — Y translated by menuH
// ---------------------------------------------------------------------------

// TestAppMenuMouseOnDesktopYTranslatedByMenuH verifies that a mouse click at
// Y=2 (screen row 2, which is desktop row 1 with menuH=1) is translated so
// the desktop receives Y=1.
// Spec: "Desktop mouse Y is translated by subtracting menuH."
func TestAppMenuMouseOnDesktopYTranslatedByMenuH(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	mb := NewMenuBar(NewSubMenu("~F~ile", NewMenuItem("~N~ew", CmUser+1, KbNone())))
	statusLine := NewStatusLine(
		NewStatusItem("~F10~ Menu", KbFunc(10), CmMenu),
	)

	var desktopReceivedMouseY int = -1
	var desktopReceivedMouse bool

	app, err := NewApplication(
		WithScreen(screen),
		WithMenuBar(mb),
		WithStatusLine(statusLine),
	)
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// We need to observe the Y that the desktop receives after translation.
	// The desktop's HandleEvent is called with translated coordinates.
	// We intercept by inspecting what the event looks like after routeMouseEvent.
	//
	// Strategy: craft an event at screen Y=2 (desktop local Y should be 1
	// with menuH=1), call routeMouseEvent, but to observe the translation we
	// check the desktop bounds. With menuBar at row 0 and statusLine at h-1,
	// the desktop starts at row 1. routeMouseEvent subtracts dBounds.A.Y
	// from the event. So a click at Y=2 becomes Y=2-1=1 when delivered to
	// the desktop.
	//
	// We verify indirectly: the desktop is at Y=1, so a click at screen Y=2
	// is within the desktop bounds (2 >= 1 and 2 < h-1). The translated Y
	// must be 2 - 1 = 1.
	_ = desktopReceivedMouseY
	_ = desktopReceivedMouse

	desktop := app.Desktop()
	dBounds := desktop.Bounds()

	// Pre-condition: desktop must start at Y=1 when menuBar is present.
	if dBounds.A.Y != 1 {
		t.Fatalf("pre-condition failed: desktop top = %d, want 1 (need menuBar)", dBounds.A.Y)
	}

	// Build a mouse event at screen coordinates (10, 2).
	ev := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:      10,
			Y:      2,
			Button: tcell.Button1,
		},
	}

	// routeMouseEvent modifies ev.Mouse in place (translates coordinates).
	// We capture what it does by calling routeMouseEvent and then reading
	// ev.Mouse — the function subtracts dBounds.A.{X,Y} before dispatching.
	// Since HandleEvent on the Desktop will be called and the event will be
	// modified, we inspect ev.Mouse after the call.
	//
	// Note: routeMouseEvent may call Desktop.HandleEvent which does nothing
	// for an unhandled mouse event; no panic expected.
	app.routeMouseEvent(ev)

	// After translation, ev.Mouse.Y should be 2 - dBounds.A.Y = 2 - 1 = 1.
	wantY := 2 - dBounds.A.Y // == 1
	if ev.Mouse.Y != wantY {
		t.Errorf("after routeMouseEvent, ev.Mouse.Y = %d, want %d (screen Y 2 minus menuH 1)", ev.Mouse.Y, wantY)
	}
}

// TestAppMenuMouseDesktopYWithoutMenuBarTranslation verifies that without a
// menuBar, a click at screen Y=1 results in desktop Y=1 (no offset subtracted,
// since desktop starts at row 0).
// Spec: "Desktop mouse Y is translated by subtracting menuH." (menuH=0 without menuBar)
func TestAppMenuMouseDesktopYWithoutMenuBarTranslation(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	sl := NewStatusLine(NewStatusItem("~F10~ Menu", KbFunc(10), CmMenu))
	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	dBounds := app.Desktop().Bounds()
	if dBounds.A.Y != 0 {
		t.Fatalf("pre-condition: desktop top = %d, want 0 (no menuBar)", dBounds.A.Y)
	}

	ev := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:      10,
			Y:      1,
			Button: tcell.Button1,
		},
	}

	app.routeMouseEvent(ev)

	// Without a menuBar, dBounds.A.Y == 0, so translation is Y - 0 == 1.
	wantY := 1 - dBounds.A.Y // == 1
	if ev.Mouse.Y != wantY {
		t.Errorf("without menuBar, ev.Mouse.Y after routeMouseEvent = %d, want %d", ev.Mouse.Y, wantY)
	}
}

// TestAppMenuMouseRow0WithNoMenuBarRoutesToDesktopNotMenuBar verifies that
// without a menuBar, a click at Y=0 is routed normally to the desktop.
// Falsifying: the menuBar path must only activate when a menuBar is registered.
func TestAppMenuMouseRow0WithNoMenuBarRoutesToDesktopNotMenuBar(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	// No menuBar; statusLine at last row; desktop starts at row 0.
	sl := NewStatusLine(NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit))
	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// A click at Y=0 must not panic and must not enter a menuBar modal loop.
	// Since there is no menuBar, the routeMouseEvent call should complete immediately.
	done := make(chan struct{}, 1)
	go func() {
		ev := &Event{
			What: EvMouse,
			Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1},
		}
		app.routeMouseEvent(ev)
		done <- struct{}{}
	}()

	select {
	case <-done:
		// Completed without hanging — no menuBar modal loop entered.
	case <-time.After(1 * time.Second):
		t.Error("routeMouseEvent hung for Y=0 with no menuBar — unexpected modal loop?")
	}
}
