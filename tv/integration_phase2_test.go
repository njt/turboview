package tv

// integration_phase2_test.go — Integration Checkpoint (Task 8).
// Each test verifies one requirement from the Phase 2 spec using REAL components
// wired together end-to-end: Application, Desktop, Window, StatusLine,
// BorlandBlue theme. No mocks for framework components.
//
// Test naming: TestIntegration<DescriptiveSuffix>.
//
// Screen coordinate convention used throughout:
//
//	Application is 80×25. StatusLine owns row 24. Desktop is rows 0-23.
//	A window at bounds (X, Y, W, H) has its top-left corner at screen (X, Y)
//	(identical to Desktop-local, since Desktop origin is (0,0)).
//	To click window-local position (wx, wy), inject screen coordinates (X+wx, Y+wy).

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// helper: newTestScreen is already defined in integration_test.go
// ---------------------------------------------------------------------------

// appWithDesktop creates a standard Application (80×25) with a StatusLine and
// BorlandBlue theme, returning the app and its Desktop.
func appWithDesktop(t *testing.T) (*Application, *Desktop) {
	t.Helper()
	screen := newTestScreen(t)
	t.Cleanup(screen.Fini)
	sl := NewStatusLine(NewStatusItem("~Alt+X~ Exit", KbAlt('x'), CmQuit))
	app, err := NewApplication(
		WithScreen(screen),
		WithStatusLine(sl),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}
	return app, app.Desktop()
}

// ---------------------------------------------------------------------------
// Requirement 1: Draw renders desktop background + both windows + frontmost on top
// ---------------------------------------------------------------------------

// TestIntegrationDrawRendersDesktopAndTwoWindows verifies that after inserting
// two windows into the Desktop and calling app.Draw, the desktop background
// pattern, both window frames, and the frontmost window drawn on top are all
// present in the output buffer.
//
// Spec req 1: "Creating an Application with a Desktop, inserting two Windows into
// the Desktop, and calling app.Draw(buf) renders: desktop pattern as background,
// both windows with frames visible, frontmost window drawn on top."
func TestIntegrationDrawRendersDesktopAndTwoWindows(t *testing.T) {
	const w, h = 80, 25
	app, desktop := appWithDesktop(t)

	// w1 at (0,0,30,12); w2 at (0,0,30,12) — deliberately overlapping.
	// w2 is inserted last so it is frontmost and drawn on top.
	w1 := NewWindow(NewRect(0, 0, 30, 12), "Back")
	w2 := NewWindow(NewRect(2, 2, 30, 12), "Front")
	desktop.Insert(w1)
	desktop.Insert(w2)

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// Desktop background must appear in an area not covered by any window.
	// Position (60, 15) is outside both windows' bounds.
	bgCell := buf.GetCell(60, 15)
	if bgCell.Rune != '░' {
		t.Errorf("desktop background at (60,15): rune = %q, want '░'", bgCell.Rune)
	}
	if bgCell.Style != theme.BorlandBlue.DesktopBackground {
		t.Errorf("desktop background at (60,15): wrong style")
	}

	// w1's top-left corner '┌' at (0,0) — w1 is NOT selected so single-line frame.
	w1CornerCell := buf.GetCell(0, 0)
	if w1CornerCell.Rune != '┌' {
		t.Errorf("w1 (back, inactive) top-left: rune = %q, want '┌'", w1CornerCell.Rune)
	}

	// w2's top-left corner '╔' at (2,2) — w2 IS selected (frontmost), double-line.
	w2CornerCell := buf.GetCell(2, 2)
	if w2CornerCell.Rune != '╔' {
		t.Errorf("w2 (front, active) top-left: rune = %q, want '╔'", w2CornerCell.Rune)
	}

	// Where w2 overlaps w1 (e.g. cell (3,3) is in w2's frame, overwriting w1's interior).
	// w2's left border '║' at x=2, rows 3..13. Check (2,3).
	overlapCell := buf.GetCell(2, 3)
	if overlapCell.Rune != '║' {
		t.Errorf("overlap area at (2,3): rune = %q, want '║' (w2 on top of w1)", overlapCell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: Frame characters — double for selected, single for non-selected
// ---------------------------------------------------------------------------

// TestIntegrationWindowFrameCharsActiveVsInactive verifies that the selected
// (frontmost) window uses double-line frame characters and the non-selected
// window uses single-line characters.
//
// Spec req 2: "Window frame uses double-line characters (╔═╗║╚═╝) for the
// selected window and single-line characters (┌─┐│└─┘) for the non-selected window."
func TestIntegrationWindowFrameCharsActiveVsInactive(t *testing.T) {
	const bw, bh = 80, 25
	app, desktop := appWithDesktop(t)

	// Non-overlapping windows so we can inspect each independently.
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1") // will be inactive (back)
	w2 := NewWindow(NewRect(30, 0, 20, 10), "W2") // will be active (front)
	desktop.Insert(w1)
	desktop.Insert(w2) // w2 is focused/selected

	buf := NewDrawBuffer(bw, bh)
	app.Draw(buf)

	// w1 is inactive: top-left must be '┌', top horizontal must be '─', side '│'.
	if c := buf.GetCell(0, 0); c.Rune != '┌' {
		t.Errorf("inactive window top-left: %q, want '┌'", c.Rune)
	}
	if c := buf.GetCell(1, 0); c.Rune != '─' {
		// x=1 is the close icon '[' — try further right to skip icon
		_ = c // close icon bracket; skip this cell
	}
	if c := buf.GetCell(0, 5); c.Rune != '│' {
		t.Errorf("inactive window left side at (0,5): %q, want '│'", c.Rune)
	}
	if c := buf.GetCell(19, 0); c.Rune != '┐' {
		t.Errorf("inactive window top-right: %q, want '┐'", c.Rune)
	}
	if c := buf.GetCell(0, 9); c.Rune != '└' {
		t.Errorf("inactive window bottom-left: %q, want '└'", c.Rune)
	}
	if c := buf.GetCell(19, 9); c.Rune != '┘' {
		t.Errorf("inactive window bottom-right: %q, want '┘'", c.Rune)
	}

	// w2 is active: double-line characters.
	if c := buf.GetCell(30, 0); c.Rune != '╔' {
		t.Errorf("active window top-left: %q, want '╔'", c.Rune)
	}
	if c := buf.GetCell(30, 5); c.Rune != '║' {
		t.Errorf("active window left side at (30,5): %q, want '║'", c.Rune)
	}
	if c := buf.GetCell(49, 0); c.Rune != '╗' {
		t.Errorf("active window top-right: %q, want '╗'", c.Rune)
	}
	if c := buf.GetCell(30, 9); c.Rune != '╚' {
		t.Errorf("active window bottom-left: %q, want '╚'", c.Rune)
	}
	if c := buf.GetCell(49, 9); c.Rune != '╝' {
		t.Errorf("active window bottom-right: %q, want '╝'", c.Rune)
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: Title centered between close and zoom icons
// ---------------------------------------------------------------------------

// TestIntegrationWindowTitleAppearsInTopBorder verifies that the window title
// is rendered somewhere in the top border row of the window, centered between
// the close icon and zoom icon.
//
// Spec req 3: "Window title appears centered in the top border between the
// close and zoom icons."
func TestIntegrationWindowTitleAppearsInTopBorder(t *testing.T) {
	const bw, bh = 80, 25
	app, desktop := appWithDesktop(t)

	// Single window at (5, 3, 30, 10); title = "Hello".
	// Title bar row in screen coords: y=3.
	// Available width between close icon (ends at x=8, i.e. screen x=5+3=8)
	// and zoom icon (starts at screen x=5+(30-4)=31); centered within that.
	win := NewWindow(NewRect(5, 3, 30, 10), "Hello")
	desktop.Insert(win)

	buf := NewDrawBuffer(bw, bh)
	app.Draw(buf)

	// Scan the title bar row (screen y=3) between close icon end (x=9) and zoom
	// icon start (x=5+30-4=31) looking for the 'H' rune from "Hello".
	titleRow := 3
	found := false
	for x := 9; x < 31; x++ {
		cell := buf.GetCell(x, titleRow)
		if cell.Rune == 'H' {
			found = true
			break
		}
	}
	if !found {
		t.Error("window title 'Hello' first character 'H' not found in expected region of title bar")
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Close icon [×] at positions (1-3, 0)
// ---------------------------------------------------------------------------

// TestIntegrationCloseIconPositionInFrame verifies that the close icon '[×]'
// occupies window-local (1,0), (2,0), (3,0) which map to specific screen cells.
//
// Spec req 4: "Close icon [×] appears at positions (1-3, 0) of the window's frame."
func TestIntegrationCloseIconPositionInFrame(t *testing.T) {
	const bw, bh = 80, 25
	app, desktop := appWithDesktop(t)

	// Window at screen origin (5, 2, 20, 10).
	// Close icon at window-local (1,0), (2,0), (3,0)
	// → screen coords (6,2), (7,2), (8,2).
	win := NewWindow(NewRect(5, 2, 20, 10), "Test")
	desktop.Insert(win)

	buf := NewDrawBuffer(bw, bh)
	app.Draw(buf)

	if c := buf.GetCell(6, 2); c.Rune != '[' {
		t.Errorf("close icon '[' at screen (6,2): rune = %q, want '['", c.Rune)
	}
	if c := buf.GetCell(7, 2); c.Rune != '×' {
		t.Errorf("close icon '×' at screen (7,2): rune = %q, want '×'", c.Rune)
	}
	if c := buf.GetCell(8, 2); c.Rune != ']' {
		t.Errorf("close icon ']' at screen (8,2): rune = %q, want ']'", c.Rune)
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: Zoom icon at (width-4, width-2) of the window's frame
// ---------------------------------------------------------------------------

// TestIntegrationZoomIconPositionInFrame verifies that the zoom icon '[↑]'
// occupies window-local (width-4, 0) through (width-2, 0).
//
// Spec req 5: "Zoom icon appears at positions (width-4, width-2) of the window's frame."
func TestIntegrationZoomIconPositionInFrame(t *testing.T) {
	const bw, bh = 80, 25
	app, desktop := appWithDesktop(t)

	// Window at (5, 2, 20, 10). width=20.
	// Zoom icon at window-local (16,0), (17,0), (18,0)
	// → screen coords (21,2), (22,2), (23,2).
	win := NewWindow(NewRect(5, 2, 20, 10), "Test")
	desktop.Insert(win)

	buf := NewDrawBuffer(bw, bh)
	app.Draw(buf)

	if c := buf.GetCell(21, 2); c.Rune != '[' {
		t.Errorf("zoom icon '[' at screen (21,2): rune = %q, want '['", c.Rune)
	}
	if c := buf.GetCell(22, 2); c.Rune != '↑' {
		t.Errorf("zoom icon '↑' at screen (22,2): rune = %q, want '↑'", c.Rune)
	}
	if c := buf.GetCell(23, 2); c.Rune != ']' {
		t.Errorf("zoom icon ']' at screen (23,2): rune = %q, want ']'", c.Rune)
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: Shadow cells have WindowShadow style with preserved rune
// ---------------------------------------------------------------------------

// TestIntegrationShadowStyleAndPreservedRune verifies that the two columns to
// the right and one row below a window have WindowShadow style applied, while
// their rune content (from the desktop background drawn underneath) is preserved.
//
// Spec req 6: "Shadow cells (2 right, 1 below each window) have WindowShadow style
// with preserved character runes from the underlying content."
func TestIntegrationShadowStyleAndPreservedRune(t *testing.T) {
	const bw, bh = 80, 25
	app, desktop := appWithDesktop(t)

	// Window at (5, 3, 20, 10). Right edge at x=25. Bottom edge at y=13.
	// Right shadow: x=25,26, rows 4..13.
	// Bottom shadow: y=13, x=7..26.
	win := NewWindow(NewRect(5, 3, 20, 10), "Shadow")
	desktop.Insert(win)

	buf := NewDrawBuffer(bw, bh)
	app.Draw(buf)

	shadowStyle := theme.BorlandBlue.WindowShadow

	// Check right shadow column at x=25, y=4 (first shadow row on right side).
	rightShadow := buf.GetCell(25, 4)
	if rightShadow.Style != shadowStyle {
		t.Errorf("right shadow at (25,4): style mismatch, want WindowShadow")
	}
	// The rune should be preserved (the desktop background '░' drawn first, then style replaced).
	if rightShadow.Rune != '░' {
		t.Errorf("right shadow at (25,4): rune = %q, want '░' (preserved from desktop background)", rightShadow.Rune)
	}

	// Check bottom shadow row at y=13, x=7.
	bottomShadow := buf.GetCell(7, 13)
	if bottomShadow.Style != shadowStyle {
		t.Errorf("bottom shadow at (7,13): style mismatch, want WindowShadow")
	}
	if bottomShadow.Rune != '░' {
		t.Errorf("bottom shadow at (7,13): rune = %q, want '░' (preserved from desktop background)", bottomShadow.Rune)
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: Clicking a back window brings it to front (OfTopSelect)
// ---------------------------------------------------------------------------

// TestIntegrationClickBackWindowBringsToFront verifies that clicking (Button1)
// on a back window brings it to front: it becomes the focused child and draws
// with the active (double-line) frame style.
//
// Spec req 7: "Clicking (Button1 press) on a back window brings it to front
// (OfTopSelect): after the event, it becomes the focused child and draws with
// active frame style."
func TestIntegrationClickBackWindowBringsToFront(t *testing.T) {
	const bw, bh = 80, 25
	app, desktop := appWithDesktop(t)

	// w1 is at (0,0,30,12) — back window after w2 is inserted.
	// w2 is at (35,0,30,12) — front window (non-overlapping).
	w1 := NewWindow(NewRect(0, 0, 30, 12), "Back")
	w2 := NewWindow(NewRect(35, 0, 30, 12), "Front")
	desktop.Insert(w1)
	desktop.Insert(w2) // w2 is focused

	if desktop.FocusedChild() != w2 {
		t.Fatal("pre-condition: w2 should be focused before click")
	}

	// Click on w1 at window-local (5, 5) → screen coords (5, 5).
	// This is in the client area, but what matters is Desktop sees it and brings
	// w1 to front via OfTopSelect.
	ev := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:      5,
			Y:      5,
			Button: tcell.Button1,
		},
	}
	app.handleEvent(ev)

	if desktop.FocusedChild() != w1 {
		t.Errorf("after clicking back window: FocusedChild() = %v, want w1", desktop.FocusedChild())
	}

	// Verify w1 now draws with active (double-line) frame.
	buf := NewDrawBuffer(bw, bh)
	app.Draw(buf)

	if c := buf.GetCell(0, 0); c.Rune != '╔' {
		t.Errorf("after bring-to-front: w1 top-left = %q, want '╔' (active frame)", c.Rune)
	}
}

// ---------------------------------------------------------------------------
// Requirement 8: Clicking close icon results in EvCommand/CmClose and Desktop removes window
// ---------------------------------------------------------------------------

// TestIntegrationClickCloseIconRemovesWindow verifies that clicking the close
// icon '[×]' on a window results in EvCommand/CmClose, which causes the Desktop
// to remove the window.
//
// Spec req 8: "Clicking the close icon [×] on a window results in EvCommand/CmClose,
// and Desktop removes the window."
func TestIntegrationClickCloseIconRemovesWindow(t *testing.T) {
	_, desktop := appWithDesktop(t)

	// Window at (5, 2, 20, 10).
	// Close icon at window-local (1-3, 0) → screen (6-8, 2).
	// Click at screen (7, 2) hits '×'.
	win := NewWindow(NewRect(5, 2, 20, 10), "CloseMe")
	desktop.Insert(win)

	if desktop.FocusedChild() != win {
		t.Fatal("pre-condition: window should be focused")
	}

	// Inject via Desktop directly (Desktop-local coordinates equal screen coords
	// here since Desktop origin is (0,0)).
	ev := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:          7,  // screen x = window.A.X(5) + local(2)
			Y:          2,  // screen y = window.A.Y(2) + local(0)
			Button:     tcell.Button1,
			ClickCount: 1,
		},
	}
	desktop.HandleEvent(ev)

	// Window should have been removed.
	for _, child := range desktop.Children() {
		if child == win {
			t.Error("close icon click should have removed the window from Desktop")
		}
	}
}

// ---------------------------------------------------------------------------
// Requirement 9: Alt+1 / Alt+2 keyboard shortcuts select windows by number
// ---------------------------------------------------------------------------

// TestIntegrationAltNumberSelectsWindowByNumber verifies that Alt+1 selects
// window number 1 and Alt+2 selects window number 2.
//
// Spec req 9: "Alt+1 selects window number 1 if such a window exists; Alt+2
// selects window number 2."
func TestIntegrationAltNumberSelectsWindowByNumber(t *testing.T) {
	app, desktop := appWithDesktop(t)

	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1", WithWindowNumber(1))
	w2 := NewWindow(NewRect(25, 0, 20, 10), "W2", WithWindowNumber(2))
	desktop.Insert(w1)
	desktop.Insert(w2) // w2 is focused

	// Press Alt+2 — w2 is already focused, verify via Alt+1 switching.
	alt1 := &Event{
		What: EvKeyboard,
		Key: &KeyEvent{
			Key:       tcell.KeyRune,
			Rune:      '1',
			Modifiers: tcell.ModAlt,
		},
	}
	app.handleEvent(alt1)

	if desktop.FocusedChild() != w1 {
		t.Errorf("Alt+1: FocusedChild() = %v, want w1 (number 1)", desktop.FocusedChild())
	}
	if !w1.HasState(SfSelected) {
		t.Error("Alt+1: w1 should have SfSelected after being brought to front")
	}

	// Now press Alt+2 to switch back to w2.
	alt2 := &Event{
		What: EvKeyboard,
		Key: &KeyEvent{
			Key:       tcell.KeyRune,
			Rune:      '2',
			Modifiers: tcell.ModAlt,
		},
	}
	app.handleEvent(alt2)

	if desktop.FocusedChild() != w2 {
		t.Errorf("Alt+2: FocusedChild() = %v, want w2 (number 2)", desktop.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Requirement 10: Desktop.Tile() with 2 windows arranges them side-by-side
// ---------------------------------------------------------------------------

// TestIntegrationTileTwoWindowsSideBySide verifies that Tile() with 2 windows
// arranges them side-by-side, each getting half the desktop width.
//
// Spec req 10: "Desktop.Tile() with 2 windows arranges them side-by-side
// (each gets half the desktop width)."
func TestIntegrationTileTwoWindowsSideBySide(t *testing.T) {
	app, desktop := appWithDesktop(t)

	w1 := NewWindow(NewRect(0, 0, 10, 5), "W1")
	w2 := NewWindow(NewRect(0, 0, 10, 5), "W2")
	desktop.Insert(w1)
	desktop.Insert(w2)

	desktop.Tile()

	// Desktop is 80 wide (screen 80, no horizontal offset). With 2 windows:
	// cols=ceil(sqrt(2))=2, so each gets 80/2=40 columns.
	dw := desktop.Bounds().Width()
	halfW := dw / 2

	b1 := w1.Bounds()
	b2 := w2.Bounds()

	// w1 at col 0, w2 at col 1.
	if b1.A.X != 0 {
		t.Errorf("Tile: w1.A.X = %d, want 0", b1.A.X)
	}
	if b1.Width() != halfW {
		t.Errorf("Tile: w1 width = %d, want %d", b1.Width(), halfW)
	}
	if b2.A.X != halfW {
		t.Errorf("Tile: w2.A.X = %d, want %d", b2.A.X, halfW)
	}

	// They should together fill the full desktop width.
	totalW := b1.Width() + b2.Width()
	if totalW != dw {
		t.Errorf("Tile: total width = %d, want %d (full desktop width)", totalW, dw)
	}

	// Same top row.
	if b1.A.Y != 0 || b2.A.Y != 0 {
		t.Errorf("Tile: y offsets = (%d, %d), want both 0 (side-by-side)", b1.A.Y, b2.A.Y)
	}

	// app.Draw should not panic with the new layout.
	buf := NewDrawBuffer(80, 25)
	app.Draw(buf)
}

// ---------------------------------------------------------------------------
// Requirement 11: Desktop.Cascade() with 2 windows arranges them diagonally
// ---------------------------------------------------------------------------

// TestIntegrationCascadeTwoWindowsDiagonalOffset verifies that Cascade() with
// 2 windows arranges them diagonally: the second window is offset (2, 1) from
// the first.
//
// Spec req 11: "Desktop.Cascade() with 2 windows arranges them diagonally offset."
func TestIntegrationCascadeTwoWindowsDiagonalOffset(t *testing.T) {
	app, desktop := appWithDesktop(t)

	w1 := NewWindow(NewRect(5, 5, 10, 5), "W1")
	w2 := NewWindow(NewRect(5, 5, 10, 5), "W2")
	desktop.Insert(w1)
	desktop.Insert(w2)

	desktop.Cascade()

	b1 := w1.Bounds()
	b2 := w2.Bounds()

	// First window at (0,0).
	if b1.A.X != 0 || b1.A.Y != 0 {
		t.Errorf("Cascade: w1 origin = (%d,%d), want (0,0)", b1.A.X, b1.A.Y)
	}

	// Second window offset by (2,1).
	dx := b2.A.X - b1.A.X
	dy := b2.A.Y - b1.A.Y
	if dx != 2 || dy != 1 {
		t.Errorf("Cascade: offset between w1 and w2 = (%d,%d), want (2,1)", dx, dy)
	}

	// Both windows should have the 3/4 desktop size.
	dw := desktop.Bounds().Width()
	dh := desktop.Bounds().Height()
	wantW := max(dw*3/4, 10)
	wantH := max(dh*3/4, 5)
	if b1.Width() != wantW || b1.Height() != wantH {
		t.Errorf("Cascade: w1 size = (%d,%d), want (%d,%d)", b1.Width(), b1.Height(), wantW, wantH)
	}

	// app.Draw should succeed without panic.
	buf := NewDrawBuffer(80, 25)
	app.Draw(buf)
}

// ---------------------------------------------------------------------------
// Requirement 12: Three-phase dispatch — OfPreProcess child runs before focused
// ---------------------------------------------------------------------------

// TestIntegrationThreePhasePreProcessBlocksFocused verifies that a child with
// OfPreProcess inserted into a window receives events before the focused child,
// and if it clears the event, the focused child does not see it.
//
// Spec req 12: "Three-phase dispatch: a child with OfPreProcess receives the
// event before the focused child; if it clears the event, the focused child
// does not see it."
func TestIntegrationThreePhasePreProcessBlocksFocused(t *testing.T) {
	app, desktop := appWithDesktop(t)

	win := NewWindow(NewRect(5, 3, 40, 15), "Dispatch")
	desktop.Insert(win)

	// Insert a preprocess observer that clears the event.
	pre := newPhaseView("pre", NewRect(0, 0, 5, 3))
	pre.SetOptions(OfPreProcess, true)
	pre.clearOnHandle = true

	// Insert a selectable focused child.
	focused := newSelectablePhaseView("focused", NewRect(0, 0, 20, 10))

	win.Insert(pre)
	win.Insert(focused)

	// Send a keyboard event through the full app pipeline.
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'z'},
	}
	app.handleEvent(ev)

	// pre should have seen the event.
	if pre.handleCount == 0 {
		t.Error("preprocess child should have received the keyboard event")
	}
	// focused should NOT have seen it (pre cleared it).
	if focused.handleCount != 0 {
		t.Errorf("focused child should NOT have received event when preprocess cleared it; handleCount = %d", focused.handleCount)
	}
	// Event should be cleared.
	if !ev.IsCleared() {
		t.Error("event should be cleared after preprocess handler consumed it")
	}
}

// ---------------------------------------------------------------------------
// Requirement 13: Drag sequence repositions the window
// ---------------------------------------------------------------------------

// TestIntegrationDragSequenceRepositionsWindow verifies that injecting a drag
// sequence (Button1 press on title bar → Button1 move → Button1 release)
// repositions the window to the new location.
//
// Spec req 13: "Injecting a drag sequence (Button1 press on title bar →
// Button1 move → Button1 release) repositions the window."
func TestIntegrationDragSequenceRepositionsWindow(t *testing.T) {
	app, desktop := appWithDesktop(t)

	// Window at (10, 5, 20, 10). Title bar at y=5 in screen coords.
	// Drag start at window-local (8, 0) → screen (18, 5).
	win := NewWindow(NewRect(10, 5, 20, 10), "Drag")
	desktop.Insert(win)

	// Step 1: Button1 press on title bar → starts drag.
	// Window-local title bar position (8, 0) → screen (10+8=18, 5+0=5).
	press := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X: 18, Y: 5,
			Button:     tcell.Button1,
			ClickCount: 1,
		},
	}
	app.handleEvent(press)

	if !win.HasState(SfDragging) {
		t.Fatal("after title bar press: SfDragging should be set")
	}

	// Step 2: Button1 move to screen (25, 8).
	// Expected new position: (25 - dragOff.X, 8 - dragOff.Y) = (25-8, 8-0) = (17, 8).
	move := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X: 25, Y: 8,
			Button: tcell.Button1,
		},
	}
	app.handleEvent(move)

	// Step 3: Button1 release → ends drag.
	release := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X: 25, Y: 8,
			Button: tcell.ButtonNone,
		},
	}
	app.handleEvent(release)

	if win.HasState(SfDragging) {
		t.Error("after release: SfDragging should be cleared")
	}

	gotBounds := win.Bounds()
	wantX := 25 - 8 // desktop-local move x minus dragOffset x
	wantY := 8 - 0  // desktop-local move y minus dragOffset y
	if gotBounds.A.X != wantX || gotBounds.A.Y != wantY {
		t.Errorf("after drag: window origin = (%d,%d), want (%d,%d)",
			gotBounds.A.X, gotBounds.A.Y, wantX, wantY)
	}
}

// ---------------------------------------------------------------------------
// Requirement 14: Mouse capture — events outside window still routed to dragging window
// ---------------------------------------------------------------------------

// TestIntegrationDragCaptureSticksToWindow verifies that during a drag, mouse
// events with coordinates outside the window's bounds are still routed to the
// dragging window (mouse capture).
//
// Spec req 14: "During drag, mouse events outside the window bounds are still
// routed to the dragging window (mouse capture)."
func TestIntegrationDragCaptureSticksToWindow(t *testing.T) {
	app, desktop := appWithDesktop(t)

	// Window at (10, 5, 20, 10). Bystander far away.
	win := NewWindow(NewRect(10, 5, 20, 10), "Capture")
	bystander := newSelectablePhaseView("bystander", NewRect(60, 0, 15, 15))
	desktop.Insert(win)
	desktop.Insert(bystander)
	// Make win the focused window.
	desktop.BringToFront(win)

	// Start drag via title bar press at screen (18, 5).
	press := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X: 18, Y: 5,
			Button:     tcell.Button1,
			ClickCount: 1,
		},
	}
	app.handleEvent(press)

	if !win.HasState(SfDragging) {
		t.Fatal("drag should have started")
	}

	bystanderCountBefore := bystander.handleCount

	// Move mouse to bystander's area (65, 5) — outside win's bounds.
	move := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X: 65, Y: 5,
			Button: tcell.Button1,
		},
	}
	app.handleEvent(move)

	// Bystander should NOT have received the event.
	if bystander.handleCount != bystanderCountBefore {
		t.Error("bystander should NOT receive events during drag capture")
	}

	// The window's position should have been updated (drag move applied).
	// new pos = (65 - dragOff.X=8, 5 - dragOff.Y=0) = (57, 5).
	gotBounds := win.Bounds()
	if gotBounds.A.X != 65-8 {
		t.Errorf("after captured drag: window x = %d, want %d", gotBounds.A.X, 65-8)
	}
}

// ---------------------------------------------------------------------------
// Requirement 15: Resize sequence changes window size
// ---------------------------------------------------------------------------

// TestIntegrationResizeSequenceChangesWindowSize verifies that injecting a
// resize sequence (Button1 press on bottom-right corner → Button1 move →
// Button1 release) changes the window's size.
//
// Spec req 15: "Injecting a resize sequence (Button1 press on bottom-right
// corner → Button1 move → Button1 release) changes the window's size."
func TestIntegrationResizeSequenceChangesWindowSize(t *testing.T) {
	app, desktop := appWithDesktop(t)

	// Window at (5, 3, 20, 10). Bottom-right corner at window-local (19, 9)
	// → screen (24, 12).
	win := NewWindow(NewRect(5, 3, 20, 10), "Resize")
	desktop.Insert(win)

	// Step 1: Press on bottom-right corner.
	press := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X: 24, Y: 12,
			Button:     tcell.Button1,
			ClickCount: 1,
		},
	}
	app.handleEvent(press)

	if !win.resizing {
		t.Fatal("after bottom-right corner press: resizing should be true")
	}

	// Step 2: Move to (30, 16) with Button1 held.
	// New width  = (30 - 5) + 1 = 26
	// New height = (16 - 3) + 1 = 14
	move := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X: 30, Y: 16,
			Button: tcell.Button1,
		},
	}
	app.handleEvent(move)

	// Step 3: Release.
	release := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X: 30, Y: 16,
			Button: tcell.ButtonNone,
		},
	}
	app.handleEvent(release)

	if win.resizing {
		t.Error("after release: resizing should be false")
	}

	got := win.Bounds()
	if got.Width() != 26 {
		t.Errorf("after resize: width = %d, want 26", got.Width())
	}
	if got.Height() != 14 {
		t.Errorf("after resize: height = %d, want 14", got.Height())
	}
	// Origin should not have moved (right-side resize).
	if got.A.X != 5 || got.A.Y != 3 {
		t.Errorf("after resize: origin = (%d,%d), want (5,3) (unchanged)", got.A.X, got.A.Y)
	}
}

// ---------------------------------------------------------------------------
// Requirement 16: Double-click on title bar toggles zoom state
// ---------------------------------------------------------------------------

// TestIntegrationDoubleClickTitleBarTogglesZoom verifies that a double-click
// (ClickCount=2) on the title bar toggles the window's zoom state.
//
// Spec req 16: "Double-click on the title bar (ClickCount=2) toggles zoom state."
func TestIntegrationDoubleClickTitleBarTogglesZoom(t *testing.T) {
	app, desktop := appWithDesktop(t)

	// Window at (5, 3, 30, 10). Title bar at y=3 in screen coords.
	// Safe title bar position: window-local (8, 0) → screen (13, 3).
	// Not on close icon (1-3) or zoom icon (26-28).
	win := NewWindow(NewRect(5, 3, 30, 10), "Zoom")
	desktop.Insert(win)

	if win.IsZoomed() {
		t.Fatal("pre-condition: window should not be zoomed initially")
	}

	// Double-click on title bar at screen (13, 3).
	dblClick := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:          13,
			Y:          3,
			Button:     tcell.Button1,
			ClickCount: 2,
		},
	}
	app.handleEvent(dblClick)

	if !win.IsZoomed() {
		t.Error("after double-click on title bar: window should be zoomed")
	}

	// Draw should reflect zoom (window fills desktop, zoom icon '↕').
	buf := NewDrawBuffer(80, 25)
	app.Draw(buf)

	// When zoomed, window fills the desktop (0,0,80,24).
	// Zoom icon middle at screen (80-3, 0) = (77, 0).
	zoomCell := buf.GetCell(77, 0)
	if zoomCell.Rune != '↕' {
		t.Errorf("zoomed window: zoom icon at (77,0) = %q, want '↕'", zoomCell.Rune)
	}
}
