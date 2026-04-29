package tv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// TestIntegrationPhase7HelpContextStatusLineFiltering verifies that a
// StatusLine containing both HcNoContext items and context-specific items
// correctly filters its output based on which window is focused.
//
// When window1 (HelpCtx=1) is focused, only HcNoContext items and HelpCtx=1
// items appear in the last row. When window2 (HelpCtx=2) is focused, only
// HcNoContext items and HelpCtx=2 items appear.
func TestIntegrationPhase7HelpContextStatusLineFiltering(t *testing.T) {
	const w, h = 80, 25

	noCtxItem := NewStatusItem("Always", KbNone(), CmUser)
	ctx1Item := NewStatusItem("~F1~Ctx1", KbFunc(1), CmUser).ForHelpCtx(1)
	ctx2Item := NewStatusItem("~F2~Ctx2", KbFunc(2), CmUser).ForHelpCtx(2)

	sl := NewStatusLine(noCtxItem, ctx1Item, ctx2Item)

	screen := newTestScreen(t)
	defer screen.Fini()
	screen.SetSize(w, h)

	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	win1 := NewWindow(NewRect(0, 0, 30, 10), "Win1")
	win1.SetHelpCtx(1)

	win2 := NewWindow(NewRect(5, 5, 30, 10), "Win2")
	win2.SetHelpCtx(2)

	// Insert both windows; Insert calls selectChild so win2 ends up focused.
	app.Desktop().Insert(win1)
	app.Desktop().Insert(win2)

	// Focus win1 first.
	app.Desktop().BringToFront(win1)

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// Verify HelpCtx=1 label text "Ctx1" appears in last row.
	lastRow := h - 1
	foundCtx1 := false
	foundCtx2 := false
	for x := 0; x < w-3; x++ {
		if buf.GetCell(x, lastRow).Rune == 'C' &&
			buf.GetCell(x+1, lastRow).Rune == 't' &&
			buf.GetCell(x+2, lastRow).Rune == 'x' &&
			buf.GetCell(x+3, lastRow).Rune == '1' {
			foundCtx1 = true
		}
		if buf.GetCell(x, lastRow).Rune == 'C' &&
			buf.GetCell(x+1, lastRow).Rune == 't' &&
			buf.GetCell(x+2, lastRow).Rune == 'x' &&
			buf.GetCell(x+3, lastRow).Rune == '2' {
			foundCtx2 = true
		}
	}
	if !foundCtx1 {
		t.Error("window1 focused (HelpCtx=1): expected 'Ctx1' label in status row, not found")
	}
	if foundCtx2 {
		t.Error("window1 focused (HelpCtx=1): did not expect 'Ctx2' label in status row, but found it")
	}

	// Verify always-visible item "Always" also appears.
	foundAlways := false
	for x := 0; x < w-5; x++ {
		if buf.GetCell(x, lastRow).Rune == 'A' &&
			buf.GetCell(x+1, lastRow).Rune == 'l' &&
			buf.GetCell(x+2, lastRow).Rune == 'w' &&
			buf.GetCell(x+3, lastRow).Rune == 'a' &&
			buf.GetCell(x+4, lastRow).Rune == 'y' &&
			buf.GetCell(x+5, lastRow).Rune == 's' {
			foundAlways = true
		}
	}
	if !foundAlways {
		t.Error("window1 focused: expected HcNoContext item 'Always' in status row, not found")
	}

	// Now switch focus to win2 and redraw.
	app.Desktop().BringToFront(win2)

	buf2 := NewDrawBuffer(w, h)
	app.Draw(buf2)

	foundCtx1 = false
	foundCtx2 = false
	for x := 0; x < w-3; x++ {
		if buf2.GetCell(x, lastRow).Rune == 'C' &&
			buf2.GetCell(x+1, lastRow).Rune == 't' &&
			buf2.GetCell(x+2, lastRow).Rune == 'x' &&
			buf2.GetCell(x+3, lastRow).Rune == '1' {
			foundCtx1 = true
		}
		if buf2.GetCell(x, lastRow).Rune == 'C' &&
			buf2.GetCell(x+1, lastRow).Rune == 't' &&
			buf2.GetCell(x+2, lastRow).Rune == 'x' &&
			buf2.GetCell(x+3, lastRow).Rune == '2' {
			foundCtx2 = true
		}
	}
	if !foundCtx2 {
		t.Error("window2 focused (HelpCtx=2): expected 'Ctx2' label in status row, not found")
	}
	if foundCtx1 {
		t.Error("window2 focused (HelpCtx=2): did not expect 'Ctx1' label in status row, but found it")
	}
}

// TestIntegrationPhase7DeepestHelpCtxWins verifies that resolveHelpCtx walks
// the full focus chain and returns the deepest non-zero HelpCtx. If a Window
// has HelpCtx=1 and its focused Button has HelpCtx=5, resolveHelpCtx must
// return 5, and the StatusLine must reflect that.
func TestIntegrationPhase7DeepestHelpCtxWins(t *testing.T) {
	const w, h = 80, 25

	ctx1Item := NewStatusItem("~F1~Win1", KbFunc(1), CmUser).ForHelpCtx(1)
	ctx5Item := NewStatusItem("~F5~Deep5", KbFunc(5), CmUser).ForHelpCtx(5)

	sl := NewStatusLine(ctx1Item, ctx5Item)

	screen := newTestScreen(t)
	defer screen.Fini()
	screen.SetSize(w, h)

	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	win := NewWindow(NewRect(0, 0, 40, 15), "TestWin")
	win.SetHelpCtx(1)

	btn := NewButton(NewRect(1, 1, 10, 3), "OK", CmOK)
	btn.SetHelpCtx(5)
	win.Insert(btn)

	app.Desktop().Insert(win)

	// resolveHelpCtx should walk Desktop -> Window (HelpCtx=1) -> Button (HelpCtx=5) -> return 5.
	got := app.resolveHelpCtx()
	if got != HelpContext(5) {
		t.Errorf("resolveHelpCtx() = %d, want 5 (deepest child HelpCtx)", got)
	}

	// The StatusLine should render the HelpCtx=5 item but not the HelpCtx=1 item.
	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	lastRow := h - 1
	foundDeep5 := false
	foundWin1 := false
	for x := 0; x < w-4; x++ {
		if buf.GetCell(x, lastRow).Rune == 'D' &&
			buf.GetCell(x+1, lastRow).Rune == 'e' &&
			buf.GetCell(x+2, lastRow).Rune == 'e' &&
			buf.GetCell(x+3, lastRow).Rune == 'p' &&
			buf.GetCell(x+4, lastRow).Rune == '5' {
			foundDeep5 = true
		}
		if buf.GetCell(x, lastRow).Rune == 'W' &&
			buf.GetCell(x+1, lastRow).Rune == 'i' &&
			buf.GetCell(x+2, lastRow).Rune == 'n' &&
			buf.GetCell(x+3, lastRow).Rune == '1' {
			foundWin1 = true
		}
	}
	if !foundDeep5 {
		t.Error("deepest HelpCtx=5: expected 'Deep5' label in status row, not found")
	}
	if foundWin1 {
		t.Error("deepest HelpCtx=5: HelpCtx=1 item 'Win1' should be suppressed, but was found")
	}
}

// TestIntegrationPhase7AllThemesWindowBackground verifies that all 5 registered
// themes (borland-blue, borland-cyan, borland-gray, matrix, c64) produce the
// correct WindowBackground colour on cells drawn inside the window client area.
func TestIntegrationPhase7AllThemesWindowBackground(t *testing.T) {
	const screenW, h = 80, 25

	tests := []struct {
		name string
	}{
		{"borland-blue"},
		{"borland-cyan"},
		{"borland-gray"},
		{"matrix"},
		{"c64"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			scheme := theme.Get(tc.name)
			if scheme == nil {
				t.Fatalf("theme.Get(%q) returned nil — theme not registered", tc.name)
			}

			screen := newTestScreen(t)
			defer screen.Fini()
			screen.SetSize(screenW, h)

			app, err := NewApplication(WithScreen(screen), WithTheme(scheme))
			if err != nil {
				t.Fatalf("NewApplication failed: %v", err)
			}

			// Window large enough to have a visible client area; placed at origin.
			win := NewWindow(NewRect(0, 0, 40, 15), "ThemeTest")
			app.Desktop().Insert(win)

			buf := NewDrawBuffer(screenW, h)
			app.Draw(buf)

			// Client area starts at (1,1) relative to the window's top-left.
			// The window's top-left in the desktop is (0,0), but the desktop
			// itself starts at row 0 (no menu bar).  Client cell (1,1).
			clientX, clientY := 1, 1
			cell := buf.GetCell(clientX, clientY)

			_, gotBG, _ := cell.Style.Decompose()
			_, wantBG, _ := scheme.WindowBackground.Decompose()

			if gotBG != wantBG {
				t.Errorf("theme %q: client cell (%d,%d) background = %v, want %v (WindowBackground)",
					tc.name, clientX, clientY, gotBG, wantBG)
			}
		})
	}
}

// TestIntegrationPhase7ConfigLoadedThemeOverride verifies the end-to-end flow
// of: write a JSON config, load it via theme.LoadConfig, create an Application
// with that scheme, draw, and confirm the window background uses the overridden
// colour.
func TestIntegrationPhase7ConfigLoadedThemeOverride(t *testing.T) {
	const screenW, h = 80, 25

	// Build a temp config: base matrix, override WindowBackground to #ff0000.
	cfg := map[string]interface{}{
		"base": "matrix",
		"overrides": map[string]string{
			"WindowBackground": "#ff0000:#0000ff",
		},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "theme.json")
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}

	scheme, err := theme.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("theme.LoadConfig: %v", err)
	}
	if scheme == nil {
		t.Fatal("theme.LoadConfig returned nil scheme")
	}

	// Verify the override was applied: background should be #0000ff, not matrix's black.
	wantBGColor := tcell.NewRGBColor(0, 0, 0xff)
	_, gotBG, _ := scheme.WindowBackground.Decompose()
	if gotBG != wantBGColor {
		t.Errorf("after override, WindowBackground bg = %v, want #0000ff (%v)", gotBG, wantBGColor)
	}

	// Confirm the original Matrix theme is unaffected.
	matrixScheme := theme.Get("matrix")
	_, matrixBG, _ := matrixScheme.WindowBackground.Decompose()
	if matrixBG == wantBGColor {
		t.Error("LoadConfig mutated the base Matrix theme — it should have made a copy")
	}

	screen := newTestScreen(t)
	defer screen.Fini()
	screen.SetSize(screenW, h)

	app, err := NewApplication(WithScreen(screen), WithTheme(scheme))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	win := NewWindow(NewRect(0, 0, 40, 15), "ConfigThemeTest")
	app.Desktop().Insert(win)

	buf := NewDrawBuffer(screenW, h)
	app.Draw(buf)

	// Client area cell at (1,1).
	cell := buf.GetCell(1, 1)
	_, cellBG, _ := cell.Style.Decompose()
	if cellBG != wantBGColor {
		t.Errorf("window client cell background = %v, want overridden #0000ff (%v)", cellBG, wantBGColor)
	}
}

// TestIntegrationPhase7StatusLineKeybindingContextFiltering verifies that a
// StatusLine keybinding bound to a specific HelpCtx only fires when the
// Application's focused window has that HelpCtx. When focus moves to a window
// with no HelpCtx, the same key must NOT trigger the command.
func TestIntegrationPhase7StatusLineKeybindingContextFiltering(t *testing.T) {
	const screenW, h = 80, 25

	const cmdTest CommandCode = CmUser + 1

	fired := 0
	ctx5Item := NewStatusItem("~F5~Ctx5Cmd", KbFunc(5), cmdTest).ForHelpCtx(5)
	sl := NewStatusLine(ctx5Item)

	screen := newTestScreen(t)
	defer screen.Fini()
	screen.SetSize(screenW, h)

	app, err := NewApplication(
		WithScreen(screen),
		WithStatusLine(sl),
		WithOnCommand(func(cmd CommandCode, _ any) bool {
			if cmd == cmdTest {
				fired++
				return true
			}
			return false
		}),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	winWithCtx := NewWindow(NewRect(0, 0, 30, 10), "WithCtx")
	winWithCtx.SetHelpCtx(5)

	winNoCtx := NewWindow(NewRect(5, 5, 30, 10), "NoCtx")
	// winNoCtx intentionally left at HcNoContext (default 0).

	app.Desktop().Insert(winWithCtx)
	app.Desktop().Insert(winNoCtx)

	// Focus winWithCtx.
	app.Desktop().BringToFront(winWithCtx)

	// Draw first so that Application.Draw calls SetActiveContext on the StatusLine,
	// which is what resolves the active HelpCtx before event handling. Without a
	// Draw the activeCtx remains 0 (HcNoContext) and context-specific bindings
	// would never fire.
	drawBuf := NewDrawBuffer(screenW, h)
	app.Draw(drawBuf)

	// Simulate an F5 keyboard event routed through the application's handler.
	keyEvent := &Event{
		What: EvKeyboard,
		Key: &KeyEvent{
			Key:       tcell.KeyF5,
			Rune:      0,
			Modifiers: tcell.ModNone,
		},
	}
	app.handleEvent(keyEvent)

	if fired != 1 {
		t.Errorf("HelpCtx=5 focused: command fired %d times after F5, want 1", fired)
	}

	// Now switch focus to window with no HelpCtx, redraw to update activeCtx,
	// then inject the same key — the command must NOT fire.
	app.Desktop().BringToFront(winNoCtx)

	drawBuf2 := NewDrawBuffer(screenW, h)
	app.Draw(drawBuf2)

	keyEvent2 := &Event{
		What: EvKeyboard,
		Key: &KeyEvent{
			Key:       tcell.KeyF5,
			Rune:      0,
			Modifiers: tcell.ModNone,
		},
	}
	app.handleEvent(keyEvent2)

	if fired != 1 {
		t.Errorf("HcNoContext focused: command should not fire on F5 (HelpCtx mismatch), but fired again (total=%d)", fired)
	}
}
