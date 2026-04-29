package tv

// integration_phase10_batch3_test.go — Integration tests for Phase 10 Tasks 10–12:
// StatusLine mouse support, StaticText centering, and Menu command dispatch.
//
// These tests exercise cross-component behavior by combining real structs (no
// mocks) and verifying observable outcomes:
//
//   Task 10: StatusLine click → EvCommand; hidden items not clickable; drag-off cancels
//   Task 11: StaticText \x03 centering in real DrawBuffer
//   Task 12: Menu PostCommand reaches Desktop via Application event queue
//
// Test naming: TestIntegrationPhase10Batch3<DescriptiveSuffix>

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// StatusLine helpers (local, not reusing status_mouse_test.go helpers which
// use different bounds and scheme)
// ---------------------------------------------------------------------------

// newBatch3StatusLine builds a 80-wide StatusLine with a plain default-style
// scheme (no scheme set; status_mouse_test.go helpers set scheme; we don't
// need style checking here, just command dispatch).
func newBatch3StatusLine(items ...*StatusItem) *StatusLine {
	sl := NewStatusLine(items...)
	sl.SetBounds(NewRect(0, 0, 80, 1))
	return sl
}

// batch3Press returns a Button1 press event at (x, 0).
func batch3Press(x int) *Event {
	return &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: x, Y: 0, Button: tcell.Button1},
	}
}

// batch3Release returns a release (no button) event at (x, 0).
func batch3Release(x int) *Event {
	return &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: x, Y: 0, Button: 0},
	}
}

// ---------------------------------------------------------------------------
// Test 1: StatusLine mouse click fires EvCommand
//
// Press and release Button1 on the same visible item → release event is
// transformed to EvCommand with the item's command code.
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch3StatusLineClickFiresCommand verifies the full
// press-then-release flow on a StatusLine item produces EvCommand.
//
// Chain: press at item x → sl.HandleEvent → pressedIdx set;
//        release at same x → sl.HandleEvent → event transformed to EvCommand.
func TestIntegrationPhase10Batch3StatusLineClickFiresCommand(t *testing.T) {
	const myCmd = CmUser + 42
	item := NewStatusItem("Click", KbNone(), myCmd)
	sl := newBatch3StatusLine(item)

	// "Click" is 5 chars; first item starts at x=1, spans x=1..5.
	pressEv := batch3Press(1)
	sl.HandleEvent(pressEv)

	releaseEv := batch3Release(1)
	sl.HandleEvent(releaseEv)

	if releaseEv.What != EvCommand {
		t.Fatalf("release on same item: What = %v, want EvCommand", releaseEv.What)
	}
	if releaseEv.Command != myCmd {
		t.Errorf("release on same item: Command = %v, want %v", releaseEv.Command, myCmd)
	}
}

// ---------------------------------------------------------------------------
// Test 2: StatusLine hidden item is not clickable
//
// An item whose HelpCtx is non-zero but does not match the StatusLine's
// activeCtx should be invisible. Clicking its would-be position fires nothing.
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch3StatusLineHiddenItemNotClickable verifies that
// a HelpCtx-filtered item cannot be activated by mouse click.
//
// Chain: item.HelpCtx = 5, sl.activeCtx = 0 → item invisible;
//        press+release at x=1 → no item at x → no EvCommand.
func TestIntegrationPhase10Batch3StatusLineHiddenItemNotClickable(t *testing.T) {
	const myCmd = CmUser + 43
	item := NewStatusItem("Hi", KbNone(), myCmd).ForHelpCtx(HelpContext(5))
	sl := newBatch3StatusLine(item)
	// activeCtx stays at 0 (HcNoContext) — item is filtered out.

	pressEv := batch3Press(1)
	sl.HandleEvent(pressEv)

	releaseEv := batch3Release(1)
	sl.HandleEvent(releaseEv)

	if releaseEv.What == EvCommand {
		t.Errorf("hidden item click: got EvCommand (cmd=%v), want no command", releaseEv.Command)
	}
}

// TestIntegrationPhase10Batch3StatusLineActiveCtxShowsHiddenItem verifies that
// when the activeCtx matches the item's HelpCtx, the item becomes clickable.
// (complementary check — the item IS there once context matches)
func TestIntegrationPhase10Batch3StatusLineActiveCtxShowsHiddenItem(t *testing.T) {
	const myCmd = CmUser + 44
	const hc = HelpContext(7)
	item := NewStatusItem("Hi", KbNone(), myCmd).ForHelpCtx(hc)
	sl := newBatch3StatusLine(item)
	sl.SetActiveContext(hc) // now the item is visible

	pressEv := batch3Press(1)
	sl.HandleEvent(pressEv)

	releaseEv := batch3Release(1)
	sl.HandleEvent(releaseEv)

	if releaseEv.What != EvCommand {
		t.Fatalf("active-ctx item: What = %v, want EvCommand", releaseEv.What)
	}
	if releaseEv.Command != myCmd {
		t.Errorf("active-ctx item: Command = %v, want %v", releaseEv.Command, myCmd)
	}
}

// ---------------------------------------------------------------------------
// Test 3: StatusLine drag-off cancels (press one item, release on another)
//
// Press on item1 at x=1, release on item2 at x=5 → pressedIdx was item1 but
// release is over item2 → no command fires.
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch3StatusLineDragOffCancels verifies that releasing
// the mouse over a different item than was pressed does not fire a command.
//
// Chain: press at item1 x → pressedIdx = item1;
//        release at item2 x → releaseIdx ≠ pressedIdx → no transform.
func TestIntegrationPhase10Batch3StatusLineDragOffCancels(t *testing.T) {
	item1 := NewStatusItem("AB", KbNone(), CmUser+50)
	item2 := NewStatusItem("CD", KbNone(), CmUser+51)
	sl := newBatch3StatusLine(item1, item2)

	// item1 "AB" starts at x=1 (width 2, spans 1..2).
	// item2 "CD" starts at x=5 (gap=2, then width 2 => 1+2+2=5).
	pressEv := batch3Press(1)
	sl.HandleEvent(pressEv)

	releaseEv := batch3Release(5)
	sl.HandleEvent(releaseEv)

	if releaseEv.What == EvCommand {
		t.Errorf("drag-off: got EvCommand (cmd=%v), want no command (drag-off should cancel)", releaseEv.Command)
	}
}

// ---------------------------------------------------------------------------
// Test 4: StaticText \x03 centering
//
// A StaticText with "\x03Hello" drawn in a 20-wide view should place 'H' at
// x = (20-5)/2 = 7.
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch3StaticTextCenteringHello verifies that a line
// prefixed with \x03 is centered within the draw buffer.
//
// Chain: NewStaticText → st.Draw(buf) → buf.GetCell(7, 0).Rune == 'H'.
func TestIntegrationPhase10Batch3StaticTextCenteringHello(t *testing.T) {
	// Width=20, text "Hello" (5 chars). Expected startX = (20-5)/2 = 7.
	st := NewStaticText(NewRect(0, 0, 20, 5), "\x03Hello")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 5)
	st.Draw(buf)

	const wantX = 7 // (20-5)/2
	got := buf.GetCell(wantX, 0).Rune
	if got != 'H' {
		t.Errorf("centered 'H': got %q at x=%d, want 'H'", got, wantX)
	}

	// Verify the full word "Hello" is in place.
	for i, want := range "Hello" {
		cell := buf.GetCell(wantX+i, 0)
		if cell.Rune != want {
			t.Errorf("centered char %d: got %q at x=%d, want %q", i, cell.Rune, wantX+i, want)
		}
	}
}

// TestIntegrationPhase10Batch3StaticTextCenteringPrefixNotDrawn verifies that
// the \x03 byte itself is never rendered as a visible character.
func TestIntegrationPhase10Batch3StaticTextCenteringPrefixNotDrawn(t *testing.T) {
	st := NewStaticText(NewRect(0, 0, 20, 5), "\x03Hello")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 5)
	st.Draw(buf)

	for x := 0; x < 20; x++ {
		if buf.GetCell(x, 0).Rune == '\x03' {
			t.Errorf("\\x03 drawn at x=%d; must be consumed, not displayed", x)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 5: StaticText mixed centered and non-centered lines
//
// Text "\x03Centered\nLeft" — first line centered, second line left-aligned.
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch3StaticTextMixedCenteringAndLeft verifies that
// centering applies only to the \x03-prefixed line; the second line starts at x=0.
//
// Chain: Draw → row 0 centered, row 1 left-aligned.
func TestIntegrationPhase10Batch3StaticTextMixedCenteringAndLeft(t *testing.T) {
	// Width=20.
	// "Centered" = 8 chars → startX = (20-8)/2 = 6.
	// "Left" starts at x=0.
	st := NewStaticText(NewRect(0, 0, 20, 5), "\x03Centered\nLeft")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 5)
	st.Draw(buf)

	// First line: "Centered" centered at x=6.
	const centeredStartX = 6 // (20-8)/2
	if buf.GetCell(centeredStartX, 0).Rune != 'C' {
		t.Errorf("centered line row 0: got %q at x=%d, want 'C'",
			buf.GetCell(centeredStartX, 0).Rune, centeredStartX)
	}
	// Verify 'C' is NOT at x=0 (it should be centered, not left-aligned).
	if buf.GetCell(0, 0).Rune == 'C' {
		t.Errorf("centered line row 0: 'C' found at x=0; line should be centered, not left-aligned")
	}

	// Second line: "Left" starts at x=0.
	if buf.GetCell(0, 1).Rune != 'L' {
		t.Errorf("left-aligned line row 1: got %q at x=0, want 'L'",
			buf.GetCell(0, 1).Rune)
	}
}

// TestIntegrationPhase10Batch3StaticTextBothLinesCentered verifies that two
// consecutive \x03-prefixed lines are each centered independently.
func TestIntegrationPhase10Batch3StaticTextBothLinesCentered(t *testing.T) {
	// Width=20.
	// Row 0: "\x03Hello" (5 chars) → startX = (20-5)/2 = 7.
	// Row 1: "\x03World" (5 chars) → startX = (20-5)/2 = 7.
	st := NewStaticText(NewRect(0, 0, 20, 5), "\x03Hello\n\x03World")
	st.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 5)
	st.Draw(buf)

	if buf.GetCell(7, 0).Rune != 'H' {
		t.Errorf("row0 'H': got %q at x=7, want 'H'", buf.GetCell(7, 0).Rune)
	}
	if buf.GetCell(7, 1).Rune != 'W' {
		t.Errorf("row1 'W': got %q at x=7, want 'W'", buf.GetCell(7, 1).Rune)
	}
}

// ---------------------------------------------------------------------------
// Test 6: Menu PostCommand reaches Desktop (CmTile tiles windows)
//
// Create an Application with a MenuBar containing a CmTile item.
// Open the menu, select the item, drain the event queue, verify Desktop.Tile()
// was called by checking that windows were repositioned.
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch3MenuCmTileReachesDesktop verifies that selecting
// the Tile item from the Window menu posts CmTile through the event queue, the
// Application picks it up, routes it to the Desktop, and Desktop.Tile() runs —
// observable as changed window bounds.
//
// Chain: menu item selected → app.PostCommand(CmTile, nil) →
//        app.PollEvent() returns EvCommand{CmTile} →
//        app.handleEvent → desktop.HandleEvent(CmTile) → desktop.Tile().
func TestIntegrationPhase10Batch3MenuCmTileReachesDesktop(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen.Init: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(80, 25)

	windowMenu := NewSubMenu("~W~indow",
		NewMenuItem("~T~ile", CmTile, KbNone()),
		NewMenuItem("~C~ascade", CmCascade, KbNone()),
	)
	mb := NewMenuBar(windowMenu)

	app, err := NewApplication(
		WithScreen(screen),
		WithTheme(theme.BorlandBlue),
		WithMenuBar(mb),
	)
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Add two windows to the Desktop so Tile() has something to rearrange.
	// Place them at irregular positions so we can confirm they moved.
	w1 := NewWindow(NewRect(0, 0, 20, 10), "Win1")
	w2 := NewWindow(NewRect(30, 5, 20, 10), "Win2")
	app.Desktop().Insert(w1)
	app.Desktop().Insert(w2)

	// Record initial bounds.
	initialW1 := w1.Bounds()
	initialW2 := w2.Bounds()

	// Select the Tile item via the modal menu loop:
	// Enter opens the Window popup; Enter again selects the first item (~T~ile).
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone) // open popup
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone) // select Tile
	mb.ActivateAt(app, 0, false)

	// The menu has posted CmTile via app.PostCommand. Drain the event queue
	// so the Application handles the command and routes it to the Desktop.
	if event := app.PollEvent(); event != nil {
		app.handleEvent(event)
	}

	// Tile() should have repositioned the windows.
	// After tiling 2 windows in a 80×24 desktop (80 wide, 25 - 1 menu row = 24 rows):
	// cols=2, rows=1, each window gets ~40 wide and 24 tall.
	finalW1 := w1.Bounds()
	finalW2 := w2.Bounds()

	if finalW1 == initialW1 && finalW2 == initialW2 {
		t.Error("after CmTile: window bounds unchanged; Desktop.Tile() was not called")
	}

	// Sanity: windows should now be adjacent (tiled), starting from x=0.
	if finalW1.A.X != 0 || finalW1.A.Y != 0 {
		t.Errorf("w1 after Tile: origin = (%d,%d), want (0,0)",
			finalW1.A.X, finalW1.A.Y)
	}
}

// TestIntegrationPhase10Batch3MenuCmCascadeReachesDesktop is a companion to the
// Tile test, verifying that CmCascade also flows through the event queue to the
// Desktop.  We detect it by checking that windows move from their initial
// positions after the cascade command is processed.
func TestIntegrationPhase10Batch3MenuCmCascadeReachesDesktop(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen.Init: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(80, 25)

	windowMenu := NewSubMenu("~W~indow",
		NewMenuItem("~T~ile", CmTile, KbNone()),
		NewMenuItem("~C~ascade", CmCascade, KbNone()),
	)
	mb := NewMenuBar(windowMenu)

	app, err := NewApplication(
		WithScreen(screen),
		WithTheme(theme.BorlandBlue),
		WithMenuBar(mb),
	)
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	w1 := NewWindow(NewRect(0, 0, 20, 10), "Win1")
	w2 := NewWindow(NewRect(30, 5, 20, 10), "Win2")
	app.Desktop().Insert(w1)
	app.Desktop().Insert(w2)

	initialW1 := w1.Bounds()
	initialW2 := w2.Bounds()

	// Navigate to Cascade: Down to move to second item, then Enter to select.
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone) // open popup
	screen.InjectKey(tcell.KeyDown, 0, tcell.ModNone)  // move to Cascade
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone) // select Cascade
	mb.ActivateAt(app, 0, false)

	if event := app.PollEvent(); event != nil {
		app.handleEvent(event)
	}

	finalW1 := w1.Bounds()
	finalW2 := w2.Bounds()

	if finalW1 == initialW1 && finalW2 == initialW2 {
		t.Error("after CmCascade: window bounds unchanged; Desktop.Cascade() was not called")
	}
}
