package tv

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// newTestScreen creates and initialises a simulation screen sized 80x25.
// Callers are responsible for calling Fini() when done (typically via defer).
func newTestScreen(t *testing.T) tcell.SimulationScreen {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen.Init() failed: %v", err)
	}
	screen.SetSize(80, 25)
	return screen
}

// === NewApplication construction ===

// TestNewApplicationWithScreenSucceeds verifies that NewApplication with a
// WithScreen option returns a non-nil application and no error.
// Spec: "NewApplication(WithScreen(sim)) uses the injected screen"
func TestNewApplicationWithScreenSucceeds(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication(WithScreen) returned error: %v", err)
	}
	if app == nil {
		t.Fatal("NewApplication(WithScreen) returned nil application")
	}
}

// TestNewApplicationCreatesDesktop verifies that NewApplication always creates
// a Desktop automatically, accessible via Desktop().
// Spec: "Desktop is always created automatically"
func TestNewApplicationCreatesDesktop(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	if app.Desktop() == nil {
		t.Error("Desktop() should not be nil after NewApplication")
	}
}

// TestNewApplicationDefaultThemeIsBorlandBlue verifies that, when no WithTheme
// option is given, the application's colour scheme is theme.BorlandBlue.
// Spec: "Default theme is theme.BorlandBlue when no WithTheme is given"
func TestNewApplicationDefaultThemeIsBorlandBlue(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	if app.scheme != theme.BorlandBlue {
		t.Errorf("default scheme = %v, want theme.BorlandBlue", app.scheme)
	}
}

// TestNewApplicationWithThemeUsesProvidedScheme verifies that WithTheme causes
// the application to use the supplied colour scheme instead of the default.
// Spec: "NewApplication(WithTheme(scheme)) sets the application's color scheme"
func TestNewApplicationWithThemeUsesProvidedScheme(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	custom := &theme.ColorScheme{}
	app, err := NewApplication(WithScreen(screen), WithTheme(custom))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	if app.scheme != custom {
		t.Errorf("scheme = %v, want the custom scheme %v", app.scheme, custom)
	}
}

// TestNewApplicationDesktopGetsApplicationColorScheme verifies that the Desktop
// created by NewApplication is configured with the application's colour scheme.
// Spec: "Desktop's ColorScheme is set to the application's theme"
func TestNewApplicationDesktopGetsApplicationColorScheme(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	custom := &theme.ColorScheme{}
	app, err := NewApplication(WithScreen(screen), WithTheme(custom))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	got := app.Desktop().ColorScheme()
	if got != custom {
		t.Errorf("Desktop().ColorScheme() = %v, want custom scheme %v", got, custom)
	}
}

// TestNewApplicationWithStatusLineAttachesIt verifies that NewApplication with
// a WithStatusLine option makes the status line accessible via StatusLine().
// Spec: "NewApplication(WithStatusLine(sl)) attaches the status line"
func TestNewApplicationWithStatusLineAttachesIt(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	sl := NewStatusLine(NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit))
	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	if app.StatusLine() != sl {
		t.Errorf("StatusLine() = %v, want %v", app.StatusLine(), sl)
	}
}

// TestNewApplicationStatusLineGetsApplicationColorScheme verifies that, when a
// StatusLine is provided, its ColorScheme is set to the application's theme.
// Spec: "StatusLine's ColorScheme is set to the application's theme if a StatusLine is provided"
func TestNewApplicationStatusLineGetsApplicationColorScheme(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	custom := &theme.ColorScheme{}
	sl := NewStatusLine(NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit))
	app, err := NewApplication(WithScreen(screen), WithTheme(custom), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	got := app.StatusLine().ColorScheme()
	if got != custom {
		t.Errorf("StatusLine().ColorScheme() = %v, want custom scheme %v", got, custom)
	}
}

// === Screen accessor ===

// TestNewApplicationScreenAccessorReturnsInjectedScreen verifies that Screen()
// returns the same screen that was injected via WithScreen.
// Spec: accessor methods: Desktop() *Desktop, StatusLine() *StatusLine, Screen() tcell.Screen
func TestNewApplicationScreenAccessorReturnsInjectedScreen(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	if app.Screen() != screen {
		t.Errorf("Screen() = %v, want injected screen %v", app.Screen(), screen)
	}
}

// === StatusLine() returns nil when not provided ===

// TestNewApplicationStatusLineNilWhenNotProvided verifies that StatusLine()
// returns nil when no WithStatusLine option is given.
func TestNewApplicationStatusLineNilWhenNotProvided(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	if app.StatusLine() != nil {
		t.Errorf("StatusLine() = %v, want nil when not provided", app.StatusLine())
	}
}

// === Draw ===

// TestDrawRendersDesktopPatternInUpperArea verifies that Draw fills the desktop
// sub-buffer (rows 0..h-2) with the desktop pattern rune ('░').
// Spec: "Draw(buf) renders Desktop into a SubBuffer for rows 0..h-2"
func TestDrawRendersDesktopPatternInUpperArea(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	const w, h = 80, 25
	screen.SetSize(w, h)

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// Every cell in rows 0..h-2 should have the desktop pattern rune '░'.
	for y := 0; y < h-1; y++ {
		for x := 0; x < w; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != '░' {
				t.Errorf("Draw: cell (%d, %d) rune = %q, want '░' (desktop pattern)", x, y, cell.Rune)
				return
			}
		}
	}
}

// TestDrawRendersStatusLineInLastRow verifies that Draw renders the StatusLine
// in the last row of the buffer when a StatusLine is present.
// Spec: "Draw(buf) renders StatusLine into a SubBuffer for the last row"
func TestDrawRendersStatusLineInLastRow(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	const w, h = 80, 25
	screen.SetSize(w, h)

	// Use a label with no tilde shortcuts so the whole label uses StatusNormal style.
	sl := NewStatusLine(NewStatusItem("Quit", KbNone(), CmQuit))
	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// The last row must contain at least one cell that is not the desktop rune —
	// the status line fills it with spaces and label text (not '░').
	lastRow := h - 1
	allDesktop := true
	for x := 0; x < w; x++ {
		if buf.GetCell(x, lastRow).Rune != '░' {
			allDesktop = false
			break
		}
	}
	if allDesktop {
		t.Error("Draw: last row should be rendered by StatusLine, not desktop pattern")
	}
}

// TestDrawFillsFullHeightWithDesktopWhenNoStatusLine verifies that, when no
// StatusLine is provided, Draw renders desktop content in every row including
// the last.
// Spec: "Draw(buf) renders Desktop into a SubBuffer for rows 0..h-2,
//
//	StatusLine into a SubBuffer for the last row"
//
// — by implication, all rows are desktop when StatusLine is absent.
func TestDrawFillsFullHeightWithDesktopWhenNoStatusLine(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	const w, h = 80, 25
	screen.SetSize(w, h)

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// All h rows should have the desktop pattern rune.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != '░' {
				t.Errorf("Draw: cell (%d, %d) rune = %q, want '░' (desktop fills all rows without status line)", x, y, cell.Rune)
				return
			}
		}
	}
}

// === Run / PostCommand ===

// TestPostCommandCmQuitCausesRunToExit verifies that calling PostCommand(CmQuit,
// nil) after Run has started causes Run to exit.
// Spec: "PostCommand(CmQuit, nil) causes Run() to exit"
func TestPostCommandCmQuitCausesRunToExit(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
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
	case <-time.After(2 * time.Second):
		t.Error("Run() did not exit within 2 s after PostCommand(CmQuit)")
	}
}

// TestRunExitsCleanlyReturningNilOnCmQuit verifies that Run() returns nil (not
// an error) when it exits due to a CmQuit command.
// Spec: "Run() exits cleanly when a CmQuit command event is processed, returning nil"
func TestRunExitsCleanlyReturningNilOnCmQuit(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		app.PostCommand(CmQuit, nil)
	}()

	err = app.Run()
	if err != nil {
		t.Errorf("Run() = %v, want nil on clean CmQuit exit", err)
	}
}

// TestPostCommandIsSafeFromAnotherGoroutine verifies that PostCommand can be
// called concurrently from a different goroutine without data races.
// Spec: "PostCommand(cmd, info) is safe to call from any goroutine"
func TestPostCommandIsSafeFromAnotherGoroutine(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	// Fire several goroutines posting commands before the quit.
	for i := 0; i < 5; i++ {
		go func() {
			app.PostCommand(CmUser, nil)
		}()
	}

	go func() {
		time.Sleep(80 * time.Millisecond)
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
	case <-time.After(2 * time.Second):
		t.Error("Run() did not exit within 2 s")
	}
}

// === Event conversion ===

// TestEventConversionKeyEventBecomesEvKeyboard verifies that a tcell keyboard
// event is converted to an EvKeyboard application event and dispatched.
// Spec: "tcell.EventKey becomes EvKeyboard"
//
// Strategy: inject a key that the StatusLine translates to CmQuit, then verify
// Run exits (which proves the keyboard event was converted and dispatched).
func TestEventConversionKeyEventBecomesEvKeyboard(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	// Bind F10 → CmQuit via the status line so that a key press causes quit.
	sl := NewStatusLine(NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit))
	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		// Inject F10: the application must convert this tcell event to EvKeyboard,
		// the StatusLine will transform it to CmQuit, and Run will exit.
		screen.InjectKey(tcell.KeyF10, 0, tcell.ModNone)
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
	case <-time.After(2 * time.Second):
		t.Error("Run() did not exit after F10 key injection — key event may not have been converted to EvKeyboard")
	}
}

// TestEventConversionResizeEventBecomesEvCommandCmResize verifies that a tcell
// resize event is converted to an EvCommand/CmResize application event.
// Spec: "tcell.EventResize becomes EvCommand/CmResize"
//
// Strategy: inject a resize event followed by a quit command so Run can exit,
// then confirm Run did not return an error (i.e. the resize was handled without
// crashing).
func TestEventConversionResizeEventBecomesEvCommandCmResize(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		// Inject a resize event.
		screen.PostEvent(tcell.NewEventResize(100, 30))
		time.Sleep(50 * time.Millisecond)
		// Then quit so Run can return.
		app.PostCommand(CmQuit, nil)
	}()

	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() returned error after resize: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Run() did not exit within 2 s after resize + quit")
	}
}
