package tv

// application_task6_test.go — Tests for Task 6: Application EventSource + OnCommand Callback.
//
// Requirements under test:
//   1. Application.PollEvent() satisfies the EventSource interface.
//   2. PollEvent converts tcell key events to EvKeyboard Events.
//   3. PollEvent returns nil when the screen is finalized (tcell returns nil).
//   4. PollEvent silently skips unrecognized tcell event types and returns the next valid one.
//   5. PollEvent triggers layoutChildren on resize before returning the event.
//   6. WithOnCommand registers an onCommand callback (AppOption).
//   7. OnCommand callback receives the CommandCode and Info from the event.
//   8. OnCommand returning true causes the event to be cleared.
//   9. OnCommand returning false leaves the event uncleared.
//  10. OnCommand is NOT called for CmQuit (CmQuit is handled first, clears event).
//  11. Desktop.app is set to the Application after NewApplication.
//  12. Run still works correctly (PollEvent refactor preserves existing behavior).

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Requirement 1: Application satisfies the EventSource interface
// ---------------------------------------------------------------------------

// TestApplicationImplementsEventSource confirms that *Application implements
// the EventSource interface (has a PollEvent() *Event method).
// Falsifying shortcut: if PollEvent is missing the test won't compile.
func TestApplicationImplementsEventSource(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Compile-time check: *Application must satisfy EventSource.
	var _ EventSource = app
}

// ---------------------------------------------------------------------------
// Requirement 2: PollEvent converts tcell key event to EvKeyboard
// ---------------------------------------------------------------------------

// TestPollEventConvertsKeyEventToEvKeyboard verifies that when a key event is
// injected into the simulation screen, PollEvent returns an EvKeyboard Event
// with the correct key fields populated.
func TestPollEventConvertsKeyEventToEvKeyboard(t *testing.T) {
	screen := newTestScreen(t)
	// Do not defer Fini — we finalize manually to unblock PollEvent.

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Inject a recognizable key before PollEvent blocks.
	screen.InjectKey(tcell.KeyF1, 0, tcell.ModNone)

	result := make(chan *Event, 1)
	go func() {
		result <- app.PollEvent()
	}()

	select {
	case ev := <-result:
		if ev == nil {
			t.Fatal("PollEvent returned nil for an injected key event")
		}
		if ev.What != EvKeyboard {
			t.Errorf("PollEvent: ev.What = %v, want EvKeyboard", ev.What)
		}
		if ev.Key == nil {
			t.Fatal("PollEvent: ev.Key is nil for a key event")
		}
		if ev.Key.Key != tcell.KeyF1 {
			t.Errorf("PollEvent: ev.Key.Key = %v, want tcell.KeyF1", ev.Key.Key)
		}
	case <-time.After(2 * time.Second):
		t.Error("PollEvent did not return within 2 s after key injection")
	}

	screen.Fini()
}

// TestPollEventConvertsKeyRuneToEvKeyboard verifies that a rune key event
// (printable character) is correctly converted with Rune populated.
func TestPollEventConvertsKeyRuneToEvKeyboard(t *testing.T) {
	screen := newTestScreen(t)

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	screen.InjectKey(tcell.KeyRune, 'x', tcell.ModNone)

	result := make(chan *Event, 1)
	go func() {
		result <- app.PollEvent()
	}()

	select {
	case ev := <-result:
		if ev == nil {
			t.Fatal("PollEvent returned nil for a rune key event")
		}
		if ev.What != EvKeyboard {
			t.Errorf("PollEvent: ev.What = %v, want EvKeyboard", ev.What)
		}
		if ev.Key == nil {
			t.Fatal("PollEvent: ev.Key is nil")
		}
		if ev.Key.Rune != 'x' {
			t.Errorf("PollEvent: ev.Key.Rune = %q, want 'x'", ev.Key.Rune)
		}
	case <-time.After(2 * time.Second):
		t.Error("PollEvent did not return within 2 s")
	}

	screen.Fini()
}

// ---------------------------------------------------------------------------
// Requirement 3: PollEvent returns nil when screen is finalized
// ---------------------------------------------------------------------------

// TestPollEventReturnsNilWhenScreenFinalized verifies that when the tcell
// screen is finalized (returns nil from its PollEvent), Application.PollEvent
// also returns nil without blocking.
func TestPollEventReturnsNilWhenScreenFinalized(t *testing.T) {
	screen := newTestScreen(t)

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	result := make(chan *Event, 1)
	go func() {
		result <- app.PollEvent()
	}()

	// Finalize the screen. tcell simulation screen's PollEvent unblocks with nil.
	screen.Fini()

	select {
	case ev := <-result:
		if ev != nil {
			t.Errorf("PollEvent returned %v after screen finalization, want nil", ev)
		}
	case <-time.After(2 * time.Second):
		t.Error("PollEvent did not return nil within 2 s after screen.Fini()")
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: PollEvent skips unrecognized event types
// ---------------------------------------------------------------------------

// TestPollEventSkipsUnrecognizedEventsAndReturnsNextValid verifies that
// PollEvent's loop skips tcell events that convertEvent cannot handle (returns
// nil for) and continues until it finds a valid event.
//
// The simulation screen only produces recognized types, so we verify the
// behavior indirectly: inject a command event via PostCommand (which produces
// a cmdTcellEvent, a recognized type) and confirm it is returned. The skip
// path is exercised internally whenever an unknown event type appears — the
// key guarantee is that PollEvent does NOT return nil prematurely and DOES
// return the next valid event.
func TestPollEventSkipsUnrecognizedEventsAndReturnsNextValid(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Post a command event so PollEvent has something to return.
	app.PostCommand(CmUser, "payload")

	result := make(chan *Event, 1)
	go func() {
		result <- app.PollEvent()
	}()

	select {
	case ev := <-result:
		if ev == nil {
			t.Fatal("PollEvent returned nil for a posted command event")
		}
		if ev.What != EvCommand {
			t.Errorf("PollEvent: ev.What = %v, want EvCommand", ev.What)
		}
		if ev.Command != CmUser {
			t.Errorf("PollEvent: ev.Command = %v, want CmUser", ev.Command)
		}
		if ev.Info != "payload" {
			t.Errorf("PollEvent: ev.Info = %v, want \"payload\"", ev.Info)
		}
	case <-time.After(2 * time.Second):
		t.Error("PollEvent did not return within 2 s after PostCommand")
	}
}

// TestPollEventDoesNotReturnNilForValidEvent is a falsifying test: a naive
// implementation that always returns nil would fail here.  We confirm the
// returned event is non-nil and has the right type.
func TestPollEventDoesNotReturnNilForValidEvent(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	result := make(chan *Event, 1)
	go func() {
		result <- app.PollEvent()
	}()

	select {
	case ev := <-result:
		if ev == nil {
			t.Error("PollEvent returned nil for a valid injected key — skipping all events incorrectly")
		}
	case <-time.After(2 * time.Second):
		t.Error("PollEvent blocked indefinitely on a valid event")
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: PollEvent triggers layoutChildren on resize
// ---------------------------------------------------------------------------

// TestPollEventTriggersLayoutOnResize verifies that when a resize event is
// received, PollEvent calls layoutChildren so the desktop and status line
// bounds are updated before the event is returned to the caller.
func TestPollEventTriggersLayoutOnResize(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	sl := NewStatusLine(NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit))
	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Post a resize event then a quit so we can stop.
	newW, newH := 100, 30
	screen.PostEvent(tcell.NewEventResize(newW, newH))
	// Also post a key so PollEvent returns after the resize.
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)

	// Drain the first event (resize).
	result := make(chan *Event, 1)
	go func() {
		result <- app.PollEvent()
	}()

	select {
	case ev := <-result:
		if ev == nil {
			t.Fatal("PollEvent returned nil for resize event")
		}
		// The returned event should be the CmResize command.
		if ev.What != EvCommand || ev.Command != CmResize {
			t.Errorf("PollEvent: resize event = {What:%v Command:%v}, want EvCommand/CmResize", ev.What, ev.Command)
		}
		// After layout, the desktop bounds should reflect the new screen size.
		// With a status line, desktop height = newH - 1.
		desktop := app.Desktop()
		db := desktop.Bounds()
		if db.Width() != newW {
			t.Errorf("after resize PollEvent, desktop width = %d, want %d", db.Width(), newW)
		}
		if db.Height() != newH-1 {
			t.Errorf("after resize PollEvent, desktop height = %d, want %d", db.Height(), newH-1)
		}
	case <-time.After(2 * time.Second):
		t.Error("PollEvent did not return resize event within 2 s")
	}
}

// TestPollEventResizeLayoutUpdatesAppBounds is a falsifying test: without the
// layoutChildren call inside PollEvent, the desktop bounds would remain at
// the pre-resize dimensions.
func TestPollEventResizeLayoutUpdatesAppBounds(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	initialBounds := app.Desktop().Bounds()
	if initialBounds.Width() != 80 || initialBounds.Height() != 25 {
		t.Fatalf("pre-condition: expected 80x25 desktop, got %dx%d", initialBounds.Width(), initialBounds.Height())
	}

	// Inject resize to a different size.
	screen.PostEvent(tcell.NewEventResize(120, 40))

	result := make(chan *Event, 1)
	go func() {
		result <- app.PollEvent()
	}()

	select {
	case ev := <-result:
		if ev == nil {
			t.Fatal("PollEvent returned nil")
		}
		db := app.Desktop().Bounds()
		// Without layoutChildren in PollEvent, this would still be 80x25.
		if db.Width() == 80 && db.Height() == 25 {
			t.Error("PollEvent did not call layoutChildren on resize — desktop bounds unchanged")
		}
		if db.Width() != 120 || db.Height() != 40 {
			t.Errorf("after resize, desktop bounds = %dx%d, want 120x40", db.Width(), db.Height())
		}
	case <-time.After(2 * time.Second):
		t.Error("PollEvent did not return within 2 s")
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: WithOnCommand registers the callback
// ---------------------------------------------------------------------------

// TestWithOnCommandIsAnAppOption verifies that WithOnCommand is an AppOption
// (i.e., returns func(*Application)) and can be passed to NewApplication
// without error.
func TestWithOnCommandIsAnAppOption(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	called := false
	cb := func(cmd CommandCode, info any) bool {
		called = true
		return false
	}

	// WithOnCommand must compile and work as an AppOption.
	app, err := NewApplication(WithScreen(screen), WithOnCommand(cb))
	if err != nil {
		t.Fatalf("NewApplication(WithOnCommand(...)): %v", err)
	}
	if app == nil {
		t.Fatal("NewApplication returned nil")
	}
	_ = called // callback not invoked yet — just verifying registration compiles
}

// TestWithOnCommandRegistersCallback verifies that the callback registered via
// WithOnCommand is stored on the application (observable effect: it is invoked
// when a non-CmQuit command is processed through handleCommand).
func TestWithOnCommandRegistersCallback(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	var receivedCmd CommandCode
	var callCount int
	cb := func(cmd CommandCode, info any) bool {
		callCount++
		receivedCmd = cmd
		return false
	}

	app, err := NewApplication(WithScreen(screen), WithOnCommand(cb))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	ev := &Event{What: EvCommand, Command: CmUser}
	app.handleCommand(ev)

	if callCount != 1 {
		t.Errorf("onCommand callback called %d times, want 1", callCount)
	}
	if receivedCmd != CmUser {
		t.Errorf("onCommand received cmd = %v, want CmUser", receivedCmd)
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: OnCommand callback receives CommandCode and Info
// ---------------------------------------------------------------------------

// TestOnCommandCallbackReceivesCommandCodeAndInfo verifies that the callback
// is passed the exact CommandCode and Info from the event.
func TestOnCommandCallbackReceivesCommandCodeAndInfo(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	type payload struct{ value int }
	p := &payload{value: 42}
	customCmd := CmUser + 5

	var gotCmd CommandCode
	var gotInfo any
	cb := func(cmd CommandCode, info any) bool {
		gotCmd = cmd
		gotInfo = info
		return false
	}

	app, err := NewApplication(WithScreen(screen), WithOnCommand(cb))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	ev := &Event{What: EvCommand, Command: customCmd, Info: p}
	app.handleCommand(ev)

	if gotCmd != customCmd {
		t.Errorf("callback cmd = %v, want %v", gotCmd, customCmd)
	}
	if gotInfo != p {
		t.Errorf("callback info = %v, want %v", gotInfo, p)
	}
}

// ---------------------------------------------------------------------------
// Requirement 8: OnCommand returning true clears the event
// ---------------------------------------------------------------------------

// TestOnCommandReturnTrueClearsEvent verifies that when the onCommand callback
// returns true, the event is cleared (What == EvNothing).
func TestOnCommandReturnTrueClearsEvent(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	cb := func(cmd CommandCode, info any) bool {
		return true // handled
	}

	app, err := NewApplication(WithScreen(screen), WithOnCommand(cb))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	ev := &Event{What: EvCommand, Command: CmUser}
	app.handleCommand(ev)

	if !ev.IsCleared() {
		t.Errorf("after onCommand returning true, ev.IsCleared() = false, want true; ev.What = %v", ev.What)
	}
}

// TestOnCommandReturnTrueClearsEventFalsifier is a falsifying test: a naive
// implementation that never clears the event would fail this check.
func TestOnCommandReturnTrueClearsEventFalsifier(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	handledCount := 0
	cb := func(cmd CommandCode, info any) bool {
		handledCount++
		return true
	}

	app, err := NewApplication(WithScreen(screen), WithOnCommand(cb))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	ev := &Event{What: EvCommand, Command: CmUser + 1}
	app.handleCommand(ev)

	if handledCount != 1 {
		t.Errorf("callback called %d times, want 1", handledCount)
	}
	// The event must be cleared.
	if ev.What != EvNothing {
		t.Errorf("ev.What = %v after callback returned true, want EvNothing", ev.What)
	}
}

// ---------------------------------------------------------------------------
// Requirement 9: OnCommand returning false leaves event uncleared
// ---------------------------------------------------------------------------

// TestOnCommandReturnFalseLeavesEventUncleared verifies that when the callback
// returns false, the event What remains EvCommand (not cleared).
func TestOnCommandReturnFalseLeavesEventUncleared(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	cb := func(cmd CommandCode, info any) bool {
		return false // not handled
	}

	app, err := NewApplication(WithScreen(screen), WithOnCommand(cb))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	ev := &Event{What: EvCommand, Command: CmUser}
	app.handleCommand(ev)

	if ev.IsCleared() {
		t.Errorf("after onCommand returning false, ev.IsCleared() = true, want false; callback should not clear the event")
	}
	if ev.What != EvCommand {
		t.Errorf("after onCommand returning false, ev.What = %v, want EvCommand", ev.What)
	}
}

// ---------------------------------------------------------------------------
// Requirement 10: OnCommand NOT called for CmQuit
// ---------------------------------------------------------------------------

// TestOnCommandNotCalledForCmQuit verifies that CmQuit is fully handled by
// the built-in switch before the onCommand callback is reached, so the
// callback is never invoked for CmQuit.
func TestOnCommandNotCalledForCmQuit(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	callCount := 0
	cb := func(cmd CommandCode, info any) bool {
		callCount++
		return true
	}

	app, err := NewApplication(WithScreen(screen), WithOnCommand(cb))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	ev := &Event{What: EvCommand, Command: CmQuit}
	app.handleCommand(ev)

	if callCount != 0 {
		t.Errorf("onCommand called %d times for CmQuit, want 0 (CmQuit must be handled before callback)", callCount)
	}
	// CmQuit should clear the event (sets quit flag) and the callback never fires.
	if !ev.IsCleared() {
		t.Error("CmQuit should clear the event in the built-in handler")
	}
}

// TestOnCommandCalledForNonQuitCommands is a paired falsifying test ensuring
// the callback IS called for commands other than CmQuit (proving the previous
// test isn't vacuous).
func TestOnCommandCalledForNonQuitCommands(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	callCount := 0
	cb := func(cmd CommandCode, info any) bool {
		callCount++
		return false
	}

	app, err := NewApplication(WithScreen(screen), WithOnCommand(cb))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	for _, cmd := range []CommandCode{CmUser, CmUser + 1, CmClose} {
		ev := &Event{What: EvCommand, Command: cmd}
		app.handleCommand(ev)
	}

	if callCount != 3 {
		t.Errorf("onCommand called %d times for 3 non-CmQuit commands, want 3", callCount)
	}
}

// TestOnCommandNotCalledWhenNilCallback verifies that handleCommand does not
// panic when no onCommand callback is registered (nil-safety).
func TestOnCommandNotCalledWhenNilCallback(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	// No WithOnCommand — callback should be nil.
	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Should not panic.
	ev := &Event{What: EvCommand, Command: CmUser}
	app.handleCommand(ev)

	// Event should be uncleared (no built-in handler for CmUser, no callback).
	if ev.IsCleared() {
		t.Error("event should not be cleared when there is no callback and the command is not CmQuit")
	}
}

// ---------------------------------------------------------------------------
// Requirement 11: Desktop.app is set after NewApplication
// ---------------------------------------------------------------------------

// TestDesktopAppFieldSetAfterNewApplication verifies that after NewApplication
// completes, the Desktop's app back-pointer is set to the Application.
// This enables ExecView's owner chain to reach Application.
func TestDesktopAppFieldSetAfterNewApplication(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	desktop := app.Desktop()
	if desktop == nil {
		t.Fatal("Desktop() returned nil")
	}

	// Access the unexported app field directly (same package).
	if desktop.app == nil {
		t.Error("Desktop.app is nil after NewApplication, want the Application")
	}
	if desktop.app != app {
		t.Errorf("Desktop.app = %p, want app %p", desktop.app, app)
	}
}

// TestDesktopAppFieldIsNotNilFalsifier is a falsifying test: if the field is
// never set, desktop.app would be nil and the test above would catch it.
// This test additionally ensures a freshly constructed Desktop (without going
// through NewApplication) does NOT have the field set, proving NewApplication
// is responsible for wiring it.
func TestDesktopAppFieldIsNotNilFalsifier(t *testing.T) {
	// A bare Desktop created directly must NOT have app set.
	bare := NewDesktop(NewRect(0, 0, 80, 25))
	if bare.app != nil {
		t.Error("bare NewDesktop() should have app == nil; only NewApplication should set it")
	}
}

// ---------------------------------------------------------------------------
// Requirement 12: Run still works correctly after PollEvent refactor
// ---------------------------------------------------------------------------

// TestRunStillExitsOnCmQuitAfterPollEventRefactor verifies that Run() still
// exits cleanly when CmQuit is posted, confirming the PollEvent refactor does
// not break existing behavior.
func TestRunStillExitsOnCmQuitAfterPollEventRefactor(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
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
		t.Error("Run() did not exit within 2 s after PostCommand(CmQuit) — PollEvent refactor may have broken Run")
	}
}

// TestRunWithOnCommandCallbackExitsOnCmQuit verifies that registering an
// onCommand callback does not prevent Run from exiting on CmQuit (the built-in
// handler runs before the callback).
func TestRunWithOnCommandCallbackExitsOnCmQuit(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	callbackCalledForQuit := false
	cb := func(cmd CommandCode, info any) bool {
		if cmd == CmQuit {
			callbackCalledForQuit = true
		}
		return false
	}

	app, err := NewApplication(WithScreen(screen), WithOnCommand(cb))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
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
		if callbackCalledForQuit {
			t.Error("onCommand callback was called for CmQuit — it must not be")
		}
	case <-time.After(2 * time.Second):
		t.Error("Run() did not exit within 2 s — onCommand callback may have interfered")
	}
}

// TestRunKeyEventStillDispatchedAfterPollEventRefactor verifies that key events
// injected through the screen still reach the component tree when Run uses
// PollEvent internally.
func TestRunKeyEventStillDispatchedAfterPollEventRefactor(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	// Bind F10 → CmQuit via the status line; Run should exit on F10.
	sl := NewStatusLine(NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit))
	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
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
		t.Error("Run() did not exit after F10 key — key events may not be dispatched after PollEvent refactor")
	}
}

// ---------------------------------------------------------------------------
// Table-driven tests: OnCommand callback with various command codes
// ---------------------------------------------------------------------------

// TestOnCommandCallbackCommandCodeVariants uses a table of command codes to
// verify that the callback is called for each non-CmQuit command exactly once
// and receives the correct code.
func TestOnCommandCallbackCommandCodeVariants(t *testing.T) {
	cmds := []struct {
		name string
		cmd  CommandCode
	}{
		{"CmClose", CmClose},
		{"CmUser", CmUser},
		{"CmUser+1", CmUser + 1},
		{"CmUser+99", CmUser + 99},
	}

	for _, tc := range cmds {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			screen := newTestScreen(t)
			defer screen.Fini()

			var gotCmd CommandCode
			cb := func(cmd CommandCode, info any) bool {
				gotCmd = cmd
				return true // clear it
			}

			app, err := NewApplication(WithScreen(screen), WithOnCommand(cb))
			if err != nil {
				t.Fatalf("NewApplication: %v", err)
			}

			ev := &Event{What: EvCommand, Command: tc.cmd}
			app.handleCommand(ev)

			if gotCmd != tc.cmd {
				t.Errorf("callback received cmd %v, want %v", gotCmd, tc.cmd)
			}
			if !ev.IsCleared() {
				t.Errorf("event not cleared after callback returned true for %v", tc.cmd)
			}
		})
	}
}

// TestOnCommandCallbackNotCalledForNonCommandEventType verifies that
// handleCommand does not invoke the callback when the event type is not
// EvCommand (e.g., a keyboard event that somehow reaches handleCommand).
func TestOnCommandCallbackNotCalledForNonCommandEventType(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	callCount := 0
	cb := func(cmd CommandCode, info any) bool {
		callCount++
		return true
	}

	app, err := NewApplication(WithScreen(screen), WithOnCommand(cb))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Keyboard event — handleCommand should ignore it.
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	app.handleCommand(ev)

	if callCount != 0 {
		t.Errorf("onCommand called %d times for a non-EvCommand event, want 0", callCount)
	}
}
