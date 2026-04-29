package tv

// integration_phase9_batch2_test.go — Integration tests for Tasks 8–11.
//
// Covers the Dialog/Modal protocol built across Tasks 8–11:
//   Task  8: Dialog.HandleEvent: Escape → CmCancel (after group delegation)
//   Task  9: Dialog.HandleEvent: Enter → CmDefault broadcast (after group delegation)
//   Task 10: Dialog.HandleEvent: CmClose → CmCancel when SfModal
//   Task 11: Window.HandleEvent: CmClose → CmCancel when SfModal
//
// All five scenarios exercise REAL components (Dialog, Button, Group, ExecView)
// wired to a real Application + SimulationScreen so that events flow through
// PollEvent → ExecView modal loop → Dialog.HandleEvent, giving true end-to-end
// coverage.
//
// Test naming: TestIntegrationPhase9Batch2<DescriptiveSuffix>.
//
// Helpers re-used from sibling test files:
//   appWithDesktopAndScreen — integration_phase3_test.go
//   execViewStack           — dialog_test.go
//   newTestScreen           — application_test.go
//   enterKey / escapeKey    — integration_phase3_test.go / dialog_phase9_test.go

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Scenario 1: Dialog shown via ExecView returns CmCancel when Escape pressed
//
// Chain: screen.InjectKey(Escape) → app.PollEvent → ExecView routes to
// dialog.HandleEvent → Dialog converts KeyEscape → EvCommand/CmCancel →
// ExecView detects closing command → returns CmCancel.
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch2EscapeReturnsCmCancel verifies that pressing Escape
// inside a modal Dialog running via desktop.ExecView causes ExecView to return
// CmCancel.
//
// Task 8: "Escape → CmCancel (after group delegation)"
func TestIntegrationPhase9Batch2EscapeReturnsCmCancel(t *testing.T) {
	_, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(10, 5, 40, 12), "Escape Test")

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	// Give ExecView time to enter the modal poll loop.
	time.Sleep(50 * time.Millisecond)

	// Inject Escape — Dialog.HandleEvent converts it to EvCommand/CmCancel.
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)

	select {
	case cmd := <-result:
		if cmd != CmCancel {
			t.Errorf("ExecView returned %v after Escape, want CmCancel", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return within 2 s after Escape key injection")
	}
}

// escapeConsumingChild is a selectable view that only clears keyboard Escape
// events; other events (commands, broadcasts) pass through unmodified.
// This is used to verify that a focused child can veto Escape without also
// blocking the CmCancel PostCommand used to terminate the loop cleanly.
type escapeConsumingChild struct {
	BaseView
}

func (c *escapeConsumingChild) Draw(_ *DrawBuffer) {}
func (c *escapeConsumingChild) HandleEvent(event *Event) {
	if event.What == EvKeyboard && event.Key != nil && event.Key.Key == tcell.KeyEscape {
		event.Clear()
	}
	// All other events pass through to the dialog's post-delegation logic.
}

// TestIntegrationPhase9Batch2EscapeConsumedByChildDoesNotExitDialog verifies
// the falsifying case: when a focused child consumes (clears) the Escape event,
// the Dialog does NOT transform it to CmCancel and ExecView continues running.
//
// The dialog is terminated cleanly via PostCommand(CmCancel) after verifying
// that Escape alone did not close it.
func TestIntegrationPhase9Batch2EscapeConsumedByChildDoesNotExitDialog(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(10, 5, 40, 12), "Escape Consumed")

	// An Escape-consuming child clears the Escape keyboard event, preventing
	// Dialog from seeing it in its post-delegation check.
	consumer := &escapeConsumingChild{}
	consumer.SetBounds(NewRect(1, 1, 10, 1))
	consumer.SetState(SfVisible, true)
	consumer.SetOptions(OfSelectable, true)
	dlg.Insert(consumer)

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Inject Escape — child consumes it, dialog must NOT exit.
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)

	// Wait briefly; ExecView must still be running.
	time.Sleep(40 * time.Millisecond)

	select {
	case cmd := <-result:
		t.Errorf("ExecView returned %v after consumed Escape; dialog should still be open", cmd)
	default:
		// Expected: loop is still running.
	}

	// Terminate cleanly via PostCommand — escapeConsumingChild passes commands through.
	app.PostCommand(CmCancel, nil)
	select {
	case <-result:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not exit after CmCancel PostCommand")
	}
}

// ---------------------------------------------------------------------------
// Scenario 2: Dialog with default Button — Enter fires the button's command
//
// Chain: screen.InjectKey(Enter) → app.PollEvent → ExecView routes to
// dialog.HandleEvent → Dialog broadcasts EvBroadcast/CmDefault to children →
// Button with OfPostProcess (WithDefault) responds, calls press() → EvCommand/CmOK
// → Dialog clears the original Enter event → ExecView detects CmOK on a
// subsequent pass (button press happens in the broadcast, which transforms to
// EvCommand) → returns CmOK.
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch2EnterFiresDefaultButton verifies that pressing Enter
// while a default Button (WithDefault) is inside the dialog broadcasts CmDefault,
// the default button responds, and ExecView returns CmOK.
//
// Task 9: "Enter → CmDefault broadcast (after group delegation)"
func TestIntegrationPhase9Batch2EnterFiresDefaultButton(t *testing.T) {
	_, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(10, 5, 40, 12), "Enter Test")
	btn := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK, WithDefault())
	dlg.Insert(btn)

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Inject Enter — Dialog broadcasts CmDefault → default button fires CmOK.
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	select {
	case cmd := <-result:
		if cmd != CmOK {
			t.Errorf("ExecView returned %v after Enter with default button, want CmOK", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmOK within 2 s after Enter with default button")
	}
}

// TestIntegrationPhase9Batch2EnterWithNoDefaultButtonDoesNotExit verifies that
// pressing Enter when no child handles it (no default button, no consumer) causes
// Dialog to broadcast CmDefault, which goes unhandled, and ExecView continues
// running — Enter alone does not close the dialog.
func TestIntegrationPhase9Batch2EnterWithNoDefaultButtonDoesNotExit(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	// Dialog with no buttons — CmDefault broadcast is delivered but nothing closes.
	dlg := NewDialog(NewRect(10, 5, 40, 12), "Enter No Default")

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	// Brief wait — ExecView must still be running.
	time.Sleep(40 * time.Millisecond)

	select {
	case cmd := <-result:
		t.Errorf("ExecView returned %v after Enter with no default button; dialog must stay open", cmd)
	default:
		// Expected: still running.
	}

	// Terminate cleanly.
	app.PostCommand(CmCancel, nil)
	select {
	case <-result:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not exit after CmCancel PostCommand")
	}
}

// ---------------------------------------------------------------------------
// Scenario 3: Clicking a Button inside modal Dialog fires command and exits
//
// Chain: screen.InjectMouse at button's absolute screen position → app.PollEvent
// → ExecView clips mouse to dialog bounds, translates, → dialog.HandleEvent
// translates (-1,-1) for frame → group routes to button → button.press() →
// EvCommand/CmOK → ExecView returns CmOK.
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch2MouseClickOnButtonExitsModal verifies that a mouse
// click at the screen position of a button inside a modal Dialog causes ExecView
// to return the button's command.
//
// Dialog at (10, 5, 40, 12). Button at dialog-client (5, 3, 10, 1).
// Client offset = +1 for frame → button is at dialog-local (6, 4).
// Button absolute screen position: (10+6, 5+4) = (16, 9).
func TestIntegrationPhase9Batch2MouseClickOnButtonExitsModal(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(10, 5, 40, 12), "Click Test")
	btn := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK)
	dlg.Insert(btn)

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Click at absolute screen (16, 9):
	//   ExecView subtracts dialog.Bounds().A = (10,5) → dialog-local (6, 4)
	//   Dialog.HandleEvent: not on frame (6>0, 4>0) → subtracts (1,1) → group-local (5, 3)
	//   Group routes (5,3) → Button at (5,3) → press() → CmOK
	app.screen.PostEvent(tcell.NewEventMouse(16, 9, tcell.Button1, tcell.ModNone))

	select {
	case cmd := <-result:
		if cmd != CmOK {
			t.Errorf("mouse click on modal button returned %v, want CmOK", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmOK within 2 s after mouse click on modal button")
	}
}

// TestIntegrationPhase9Batch2MouseClickOutsideDialogIgnored verifies that a
// mouse click outside the modal dialog's bounds does not close the dialog.
func TestIntegrationPhase9Batch2MouseClickOutsideDialogIgnored(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	// Dialog occupies cols 10-49, rows 5-16.
	dlg := NewDialog(NewRect(10, 5, 40, 12), "Outside Click")
	btn := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK)
	dlg.Insert(btn)

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Click well outside the dialog (col 2, row 2 — desktop area).
	app.screen.PostEvent(tcell.NewEventMouse(2, 2, tcell.Button1, tcell.ModNone))

	// Give the loop time to process the outside click (should be discarded).
	time.Sleep(40 * time.Millisecond)

	select {
	case cmd := <-result:
		t.Errorf("ExecView returned %v after outside click; dialog must remain open", cmd)
	default:
		// Expected: loop still running.
	}

	// Terminate cleanly.
	app.PostCommand(CmCancel, nil)
	select {
	case <-result:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not exit after CmCancel PostCommand")
	}
}

// ---------------------------------------------------------------------------
// Scenario 4: CmClose on modal Dialog converts to CmCancel
//
// Chain: app.PostCommand(CmClose) → ExecView routes EvCommand to
// dialog.HandleEvent → Dialog (SfModal set) transforms CmClose → CmCancel →
// ExecView detects CmCancel → returns CmCancel.
//
// Task 10: "Dialog.HandleEvent: CmClose → CmCancel when SfModal"
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch2CmCloseOnModalDialogReturnsCmCancel verifies that
// posting CmClose to a running modal dialog causes ExecView to return CmCancel,
// not CmClose (per the Dialog's CmClose→CmCancel transformation).
func TestIntegrationPhase9Batch2CmCloseOnModalDialogReturnsCmCancel(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(10, 5, 40, 12), "CmClose Test")

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Verify SfModal is set while ExecView is running.
	if !dlg.HasState(SfModal) {
		t.Error("pre-condition: SfModal should be set on dialog while ExecView is running")
	}

	// Post CmClose — Dialog transforms it to CmCancel.
	app.PostCommand(CmClose, nil)

	select {
	case cmd := <-result:
		if cmd != CmCancel {
			t.Errorf("ExecView returned %v after CmClose on modal dialog, want CmCancel", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return within 2 s after CmClose posted to modal dialog")
	}
}

// TestIntegrationPhase9Batch2CmCloseOnNonModalDialogPassesThrough verifies that
// CmClose on a non-modal dialog (SfModal NOT set) is NOT transformed to CmCancel.
// We use a standalone dialog without ExecView so we can inspect the event directly.
//
// Task 10: "Dialog only transforms CmClose when SfModal is set."
func TestIntegrationPhase9Batch2CmCloseOnNonModalDialogPassesThrough(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 20), "Non-Modal")
	// SfModal is deliberately NOT set.

	ev := &Event{What: EvCommand, Command: CmClose}
	dlg.HandleEvent(ev)

	if ev.Command != CmClose {
		t.Errorf("non-modal dialog CmClose: Command = %v, want CmClose (unchanged)", ev.Command)
	}
}

// ---------------------------------------------------------------------------
// Scenario 5: Modal Window CmClose → CmCancel
//
// Window.HandleEvent: CmClose → CmCancel when SfModal is set.
//
// Task 11: "Window.HandleEvent: CmClose → CmCancel when SfModal"
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch2WindowCmCloseOnModalReturnsCmCancel verifies that
// a Window with SfModal set transforms CmClose to CmCancel in its HandleEvent.
// This is tested both via direct HandleEvent (unit-level) and via ExecView
// (integration-level).
func TestIntegrationPhase9Batch2WindowCmCloseOnModalReturnsCmCancel(t *testing.T) {
	// Direct HandleEvent path — no ExecView required.
	win := NewWindow(NewRect(0, 0, 40, 20), "Modal Win")
	win.SetState(SfModal, true)

	ev := &Event{What: EvCommand, Command: CmClose}
	win.HandleEvent(ev)

	if ev.Command != CmCancel {
		t.Errorf("modal Window CmClose: Command = %v, want CmCancel", ev.Command)
	}
}

// TestIntegrationPhase9Batch2WindowCmCloseNonModalPassesThrough verifies that
// CmClose on a non-modal Window is NOT transformed to CmCancel.
func TestIntegrationPhase9Batch2WindowCmCloseNonModalPassesThrough(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 40, 20), "Non-Modal Win")
	// SfModal is deliberately NOT set.

	ev := &Event{What: EvCommand, Command: CmClose}
	win.HandleEvent(ev)

	// Non-modal window dispatches CmClose to the group, which may handle it.
	// The critical constraint is that the command was NOT changed to CmCancel
	// by the Window's modal guard.
	if ev.Command == CmCancel {
		t.Errorf("non-modal Window CmClose was transformed to CmCancel — should only happen when SfModal is set")
	}
}

// TestIntegrationPhase9Batch2WindowExecViewCmCloseReturnsCmCancel verifies the
// full integration path: a Window running as a modal via desktop.ExecView
// receives CmClose and returns CmCancel (Window modal guard fires before
// ExecView sees the closing command).
func TestIntegrationPhase9Batch2WindowExecViewCmCloseReturnsCmCancel(t *testing.T) {
	screen := newTestScreen(t)
	app, err := NewApplication(
		WithScreen(screen),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		screen.Fini()
		t.Fatalf("NewApplication: %v", err)
	}
	defer screen.Fini()

	desktop := app.Desktop()
	win := NewWindow(NewRect(5, 5, 50, 15), "Modal Window")

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(win)
	}()

	time.Sleep(50 * time.Millisecond)

	// Verify SfModal was set by ExecView.
	if !win.HasState(SfModal) {
		t.Error("pre-condition: Window SfModal should be set while ExecView is running")
	}

	// Post CmClose — Window transforms it to CmCancel; ExecView returns CmCancel.
	app.PostCommand(CmClose, nil)

	select {
	case cmd := <-result:
		if cmd != CmCancel {
			t.Errorf("Window ExecView returned %v after CmClose, want CmCancel", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("Window ExecView did not return within 2 s after CmClose")
	}
}

// TestIntegrationPhase9Batch2WindowSfModalClearedAfterExecView verifies that
// SfModal is cleared on the Window after ExecView returns — both for CmClose
// (transformed to CmCancel) and normal closing commands.
func TestIntegrationPhase9Batch2WindowSfModalClearedAfterExecView(t *testing.T) {
	screen := newTestScreen(t)
	app, err := NewApplication(
		WithScreen(screen),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		screen.Fini()
		t.Fatalf("NewApplication: %v", err)
	}
	defer screen.Fini()

	desktop := app.Desktop()
	win := NewWindow(NewRect(5, 5, 50, 15), "Cleanup Win")

	done := make(chan CommandCode, 1)
	go func() {
		done <- desktop.ExecView(win)
	}()

	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmClose, nil) // transforms to CmCancel, exits loop

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}

	if win.HasState(SfModal) {
		t.Error("SfModal is still set on Window after ExecView returned — should be cleared")
	}
}

// ---------------------------------------------------------------------------
// Combined scenario: Escape then Enter in the same dialog session
//
// Verifies that Escape correctly terminates the modal loop even when Enter was
// previously pressed and the CmDefault broadcast went unhandled (dialog stayed
// open), ruling out any state corruption.
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch2EnterThenEscapeExitsModal exercises the full
// interaction: Enter is pressed (broadcasts CmDefault, stays open), then Escape
// closes the dialog with CmCancel.
func TestIntegrationPhase9Batch2EnterThenEscapeExitsModal(t *testing.T) {
	_, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	// Dialog with no default button: Enter → CmDefault broadcast → unhandled → stays open.
	dlg := NewDialog(NewRect(10, 5, 40, 12), "Enter Then Escape")

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Press Enter — no default button, dialog stays open.
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	time.Sleep(30 * time.Millisecond)

	// Verify dialog is still running.
	select {
	case cmd := <-result:
		t.Fatalf("ExecView returned %v after Enter with no default button; expected dialog to stay open", cmd)
	default:
	}

	// Press Escape — Dialog converts KeyEscape → CmCancel → ExecView exits.
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)

	select {
	case cmd := <-result:
		if cmd != CmCancel {
			t.Errorf("ExecView returned %v after Escape, want CmCancel", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmCancel within 2 s after Escape")
	}
}
