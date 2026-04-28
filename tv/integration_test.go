package tv

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// TestIntegrationFullDrawRendering verifies that an Application with a
// StatusLine draws the Desktop pattern into rows 0-23 and the StatusLine into
// row 24 on an 80x25 SimulationScreen.
func TestIntegrationFullDrawRendering(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	const w, h = 80, 25

	sl := NewStatusLine(NewStatusItem("~Alt+X~ Exit", KbAlt('x'), CmQuit))
	app, err := NewApplication(
		WithScreen(screen),
		WithStatusLine(sl),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	desktopStyle := theme.BorlandBlue.DesktopBackground

	// Rows 0..23 must be filled with '░' in DesktopBackground style.
	for y := 0; y < h-1; y++ {
		for x := 0; x < w; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != '░' {
				t.Errorf("row %d col %d: rune = %q, want '░'", y, x, cell.Rune)
				return
			}
			if cell.Style != desktopStyle {
				t.Errorf("row %d col %d: style mismatch (desktop)", y, x)
				return
			}
		}
	}

	// Row 24 must NOT be the desktop pattern — StatusLine owns it.
	allDesktop := true
	for x := 0; x < w; x++ {
		if buf.GetCell(x, h-1).Rune != '░' {
			allDesktop = false
			break
		}
	}
	if allDesktop {
		t.Error("row 24 should be rendered by StatusLine, not desktop pattern")
	}
}

// TestIntegrationStatusLineContent verifies that the status row contains the
// shortcut text "Alt+X" in StatusShortcut style and " Exit" in StatusNormal
// style when the label is "~Alt+X~ Exit".
func TestIntegrationStatusLineContent(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	const w, h = 80, 25

	sl := NewStatusLine(NewStatusItem("~Alt+X~ Exit", KbAlt('x'), CmQuit))
	app, err := NewApplication(
		WithScreen(screen),
		WithStatusLine(sl),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// StatusLine.Draw starts at x=1 for the first item.
	// "~Alt+X~ Exit" parses to [{Text:"Alt+X", Shortcut:true}, {Text:" Exit", Shortcut:false}]
	// So "Alt+X" occupies columns 1..5 in StatusShortcut style,
	// and " Exit" occupies columns 6..10 in StatusNormal style.

	shortcutStyle := theme.BorlandBlue.StatusShortcut
	normalStyle := theme.BorlandBlue.StatusNormal
	lastRow := h - 1

	shortcut := "Alt+X"
	for i, ch := range shortcut {
		cell := buf.GetCell(1+i, lastRow)
		if cell.Rune != ch {
			t.Errorf("shortcut col %d: rune = %q, want %q", 1+i, cell.Rune, ch)
		}
		if cell.Style != shortcutStyle {
			t.Errorf("shortcut col %d: expected StatusShortcut style", 1+i)
		}
	}

	normal := " Exit"
	normalStart := 1 + len(shortcut)
	for i, ch := range normal {
		cell := buf.GetCell(normalStart+i, lastRow)
		if cell.Rune != ch {
			t.Errorf("normal col %d: rune = %q, want %q", normalStart+i, cell.Rune, ch)
		}
		if cell.Style != normalStyle {
			t.Errorf("normal col %d: expected StatusNormal style", normalStart+i)
		}
	}
}

// TestIntegrationKeyEventAltXTriggersQuit verifies that injecting an Alt+X
// keyboard event via handleEvent transforms it to EvCommand/CmQuit and sets
// app.quit to true.
func TestIntegrationKeyEventAltXTriggersQuit(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	sl := NewStatusLine(NewStatusItem("~Alt+X~ Exit", KbAlt('x'), CmQuit))
	app, err := NewApplication(
		WithScreen(screen),
		WithStatusLine(sl),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	if app.quit {
		t.Fatal("app.quit should be false before any event")
	}

	event := &Event{
		What: EvKeyboard,
		Key: &KeyEvent{
			Key:       tcell.KeyRune,
			Rune:      'x',
			Modifiers: tcell.ModAlt,
		},
	}

	app.handleEvent(event)

	if !app.quit {
		t.Error("app.quit should be true after Alt+X event")
	}
	if event.What != EvNothing {
		t.Errorf("event.What = %v after handling, want EvNothing (cleared)", event.What)
	}
}

// TestIntegrationNonMatchingKeyDoesNotQuit verifies that injecting an
// unbound key ('a') does NOT set app.quit to true.
func TestIntegrationNonMatchingKeyDoesNotQuit(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	sl := NewStatusLine(NewStatusItem("~Alt+X~ Exit", KbAlt('x'), CmQuit))
	app, err := NewApplication(
		WithScreen(screen),
		WithStatusLine(sl),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	event := &Event{
		What: EvKeyboard,
		Key: &KeyEvent{
			Key:       tcell.KeyRune,
			Rune:      'a',
			Modifiers: tcell.ModNone,
		},
	}

	app.handleEvent(event)

	if app.quit {
		t.Error("app.quit should remain false after an unbound key event")
	}
}

// TestIntegrationResizeRecalculatesBounds verifies that after updating
// app.bounds and calling layoutChildren(), the Desktop occupies rows 0..h-2
// and the StatusLine occupies the last row at the new size.
func TestIntegrationResizeRecalculatesBounds(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	sl := NewStatusLine(NewStatusItem("~Alt+X~ Exit", KbAlt('x'), CmQuit))
	app, err := NewApplication(
		WithScreen(screen),
		WithStatusLine(sl),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	// Simulate a resize to 40x10.
	const newW, newH = 40, 10
	app.bounds = NewRect(0, 0, newW, newH)
	app.layoutChildren()

	desktopBounds := app.desktop.Bounds()
	if desktopBounds.Height() != newH-1 {
		t.Errorf("Desktop height after resize = %d, want %d", desktopBounds.Height(), newH-1)
	}
	if desktopBounds.Width() != newW {
		t.Errorf("Desktop width after resize = %d, want %d", desktopBounds.Width(), newW)
	}

	slBounds := app.statusLine.Bounds()
	if slBounds.A.Y != newH-1 {
		t.Errorf("StatusLine top row after resize = %d, want %d", slBounds.A.Y, newH-1)
	}
	if slBounds.Width() != newW {
		t.Errorf("StatusLine width after resize = %d, want %d", slBounds.Width(), newW)
	}
}

// TestIntegrationPostCommandCmQuitExitsRun verifies that PostCommand(CmQuit,
// nil) causes Run() to exit cleanly.
func TestIntegrationPostCommandCmQuitExitsRun(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	sl := NewStatusLine(NewStatusItem("~Alt+X~ Exit", KbAlt('x'), CmQuit))
	app, err := NewApplication(
		WithScreen(screen),
		WithStatusLine(sl),
		WithTheme(theme.BorlandBlue),
	)
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

// TestIntegrationColorSchemePropagation verifies that both Desktop and
// StatusLine return the ColorScheme set by the Application via WithTheme.
func TestIntegrationColorSchemePropagation(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	sl := NewStatusLine(NewStatusItem("~Alt+X~ Exit", KbAlt('x'), CmQuit))
	app, err := NewApplication(
		WithScreen(screen),
		WithStatusLine(sl),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}

	if got := app.Desktop().ColorScheme(); got != theme.BorlandBlue {
		t.Errorf("Desktop().ColorScheme() = %v, want theme.BorlandBlue", got)
	}

	if got := app.StatusLine().ColorScheme(); got != theme.BorlandBlue {
		t.Errorf("StatusLine().ColorScheme() = %v, want theme.BorlandBlue", got)
	}
}
