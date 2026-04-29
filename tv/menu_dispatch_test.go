package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newMenuDispatchApp creates an Application with a MenuBar and a Desktop.
// The SimulationScreen is pre-sized at 80x25 with BorlandBlue theme.
func newMenuDispatchApp(t *testing.T) (*Application, *MenuBar, tcell.SimulationScreen) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen.Init: %v", err)
	}
	screen.SetSize(80, 25)

	fileMenu := NewSubMenu("~F~ile",
		NewMenuItem("~N~ew", CmUser+1, KbNone()),
		NewMenuItem("~O~pen", CmUser+2, KbNone()),
	)
	windowMenu := NewSubMenu("~W~indow",
		NewMenuItem("~T~ile", CmTile, KbNone()),
		NewMenuItem("~C~ascade", CmCascade, KbNone()),
	)
	mb := NewMenuBar(fileMenu, windowMenu)

	app, err := NewApplication(
		WithScreen(screen),
		WithTheme(theme.BorlandBlue),
		WithMenuBar(mb),
	)
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	return app, mb, screen
}

// ---------------------------------------------------------------------------
// Test 1: checkPopupResult uses PostCommand (not direct handleCommand)
// ---------------------------------------------------------------------------

// TestMenuDispatchUsesPostCommand verifies that selecting a menu item posts the
// command through app.PostCommand so that subsequent PollEvent returns it.
// Spec: "checkPopupResult uses app.PostCommand(result, nil) instead of
// app.handleCommand(cmdEvent) for non-cancel results"
func TestMenuDispatchUsesPostCommand(t *testing.T) {
	app, mb, screen := newMenuDispatchApp(t)
	defer screen.Fini()

	var receivedCmd CommandCode
	app.onCommand = func(cmd CommandCode, info any) bool {
		receivedCmd = cmd
		return true
	}

	// Enter opens File menu popup; Enter selects first item (CmUser+1).
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	mb.ActivateAt(app, 0, false)

	// The command is posted to the event queue via PostCommand; drain it.
	if event := app.PollEvent(); event != nil {
		app.handleCommand(event)
	}

	// The command should have been received by onCommand after the drain.
	if receivedCmd != CmUser+1 {
		t.Errorf("PostCommand: got cmd=%v, want CmUser+1 (%v)", receivedCmd, CmUser+1)
	}
}

// ---------------------------------------------------------------------------
// Test 2: CmCancel still closes popup and stays active
// ---------------------------------------------------------------------------

// TestMenuDispatchCancelDoesNotPostCommand verifies that CmCancel from a popup
// closes the popup without posting any command.
// Spec: "If CmCancel, close popup (unchanged from before)"
func TestMenuDispatchCancelDoesNotPostCommand(t *testing.T) {
	app, mb, screen := newMenuDispatchApp(t)
	defer screen.Fini()

	commandPosted := false
	app.onCommand = func(cmd CommandCode, info any) bool {
		if cmd != CmCancel {
			commandPosted = true
		}
		return true
	}

	// Enter opens popup; Escape → CmCancel; F10 deactivates.
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	screen.InjectKey(tcell.KeyF10, 0, tcell.ModNone)
	mb.ActivateAt(app, 0, false)

	if commandPosted {
		t.Error("CmCancel should not result in a non-cancel command being posted")
	}
	if mb.IsActive() {
		t.Error("loop still active after Enter+Escape+F10")
	}
}

// ---------------------------------------------------------------------------
// Test 3: handleModalEvent EvCommand for non-CmMenu just returns
// ---------------------------------------------------------------------------

// TestMenuDispatchHandleModalEventNonCmMenuDoesNotCallHandleCommand verifies
// that during the modal loop, non-CmMenu EvCommand events are not processed
// directly (they are returned/dropped).
// Spec: "In handleModalEvent, change the EvCommand handling to not call
// app.handleCommand for non-CmMenu commands during the modal loop"
func TestMenuDispatchHandleModalEventNonCmMenuJustReturns(t *testing.T) {
	app, mb, screen := newMenuDispatchApp(t)
	defer screen.Fini()

	// Track if any external command handler is called during the modal loop.
	commandCalledDuringModal := false
	app.onCommand = func(cmd CommandCode, info any) bool {
		commandCalledDuringModal = true
		return true
	}

	// Inject an EvCommand event via PostEvent (simulates a command posted while
	// the modal loop is running). It should NOT be dispatched to handleCommand
	// inside the modal loop. Then F10 deactivates.
	// We do this by injecting the command event before activation.
	cmdEv := &cmdTcellEvent{cmd: CmUser + 5}
	cmdEv.SetEventNow()
	_ = screen.PostEvent(cmdEv)
	screen.InjectKey(tcell.KeyF10, 0, tcell.ModNone)
	mb.ActivateAt(app, 0, false)

	if commandCalledDuringModal {
		t.Error("handleModalEvent should not call handleCommand for non-CmMenu EvCommand events")
	}
}

// ---------------------------------------------------------------------------
// Test 4: Loop deactivates after menu selection
// ---------------------------------------------------------------------------

// TestMenuDispatchLoopExitsAfterSelection verifies the modal loop exits cleanly
// when a popup item is selected (regression guard: loop must not hang).
// Spec: "mb.closePopup() and mb.active = false are called after dispatch"
func TestMenuDispatchLoopExitsAfterSelection(t *testing.T) {
	app, mb, screen := newMenuDispatchApp(t)
	defer screen.Fini()

	app.onCommand = func(cmd CommandCode, info any) bool { return true }

	// Enter opens popup; Enter selects first item.
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	mb.ActivateAt(app, 0, false)

	if mb.IsActive() {
		t.Error("loop still active after popup item selected; loop must exit")
	}
	if mb.Popup() != nil {
		t.Error("popup not nil after loop exit; popup must be closed")
	}
}
