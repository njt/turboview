package tv

// integration_phase3_test.go — Integration Checkpoint (Tasks 1-4).
// Each test verifies one requirement from the Phase 3 spec using REAL components
// wired together end-to-end: Application, Desktop, Window, Button, Label,
// StaticText inside a real Window/Desktop/Application chain.
//
// Test naming: TestIntegration<DescriptiveSuffix>.
//
// Screen coordinate convention:
//
//	Application is 80×25. Desktop is rows 0-24 (no status line in these tests).
//	A window at bounds (X, Y, W, H) has its top-left corner at screen (X, Y).
//	Window has a 1-cell border. Children at (0,0) in the window's group appear at
//	screen (winX+1, winY+1).

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// appWithWindow is a test helper that creates a standard Application (80×25) with
// BorlandBlue theme and no status line, inserts a Window at the given bounds, and
// returns all three.
func appWithWindow(t *testing.T, bounds Rect, title string) (*Application, *Desktop, *Window) {
	t.Helper()
	screen := newTestScreen(t)
	t.Cleanup(screen.Fini)
	app, err := NewApplication(
		WithScreen(screen),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		t.Fatalf("NewApplication returned error: %v", err)
	}
	desktop := app.Desktop()
	win := NewWindow(bounds, title)
	desktop.Insert(win)
	return app, desktop, win
}

// enterKey returns an EvKeyboard event for the Enter key with no modifiers.
func enterKey() *Event {
	return &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyEnter},
	}
}

// tabKey returns an EvKeyboard event for the Tab key with no modifiers.
func tabKey() *Event {
	return &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyTab},
	}
}

// shiftTabKey returns an EvKeyboard event for Shift+Tab (Backtab).
func shiftTabKey() *Event {
	return &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyBacktab},
	}
}

// ---------------------------------------------------------------------------
// Requirement 1: Button in Window fires command via app.handleEvent(Enter)
// ---------------------------------------------------------------------------

// TestIntegrationButtonInWindowFiresCommandOnEnter verifies that pressing Enter
// while a Button has focus inside a Window inside a Desktop inside an Application
// transforms the event to EvCommand with the button's command code.
//
// Spec req 1: "A Button inserted into a Window, inside a Desktop, inside an
// Application: injecting an Enter key via app.handleEvent reaches the Button and
// fires its command (transforms event to EvCommand)."
func TestIntegrationButtonInWindowFiresCommandOnEnter(t *testing.T) {
	app, _, win := appWithWindow(t, NewRect(5, 3, 30, 10), "Test")

	btn := NewButton(NewRect(0, 0, 10, 1), "OK", CmOK)
	win.Insert(btn)

	// Button should be the focused child of the window after insertion.
	if win.FocusedChild() != btn {
		t.Fatal("pre-condition: button should be focused after insertion")
	}

	ev := enterKey()
	app.handleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("after Enter, event.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("after Enter, event.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: Tab advances focus from button 1 to button 2
// ---------------------------------------------------------------------------

// TestIntegrationTabAdvancesFocusBetweenTwoButtons verifies that a Tab key
// advances focus from the first button to the second, setting SfSelected on the
// newly focused button and clearing it on the old one.
//
// Spec req 2: "Two Buttons in a Window: Tab key advances focus from button 1 to
// button 2. The newly focused button has SfSelected. The previously focused button
// loses SfSelected."
//
// Note: Tab is dispatched directly to the window because the Desktop's group
// intercepts Tab for window-level focus traversal before it reaches the window's
// internal group. Dispatching to win.HandleEvent replicates the path a real
// window takes once it receives keyboard events from the desktop.
func TestIntegrationTabAdvancesFocusBetweenTwoButtons(t *testing.T) {
	_, _, win := appWithWindow(t, NewRect(5, 3, 40, 15), "Focus")

	btn1 := NewButton(NewRect(0, 0, 10, 1), "One", CmOK)
	btn2 := NewButton(NewRect(12, 0, 10, 1), "Two", CmCancel)
	win.Insert(btn1)
	win.Insert(btn2)

	// After insertion, btn2 is focused (last inserted selectable).
	// Move focus back to btn1 so we can test Tab forward.
	win.SetFocusedChild(btn1)

	if win.FocusedChild() != btn1 {
		t.Fatal("pre-condition: btn1 should be focused before Tab")
	}
	if !btn1.HasState(SfSelected) {
		t.Fatal("pre-condition: btn1 should have SfSelected before Tab")
	}

	ev := tabKey()
	win.HandleEvent(ev)

	if win.FocusedChild() != btn2 {
		t.Errorf("after Tab, FocusedChild() = %v, want btn2", win.FocusedChild())
	}
	if !btn2.HasState(SfSelected) {
		t.Error("after Tab, btn2 should have SfSelected")
	}
	if btn1.HasState(SfSelected) {
		t.Error("after Tab, btn1 should have lost SfSelected")
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: Shift+Tab moves focus backward
// ---------------------------------------------------------------------------

// TestIntegrationShiftTabMovesFocusBackward verifies that Shift+Tab moves focus
// from button 2 back to button 1.
//
// Spec req 3: "Two Buttons in a Window: Shift+Tab moves focus backward (from
// button 2 to button 1)."
//
// Note: Shift+Tab is dispatched directly to the window for the same reason as
// TestIntegrationTabAdvancesFocusBetweenTwoButtons (Desktop's group intercepts
// it for window-level traversal before it reaches the window's inner group).
func TestIntegrationShiftTabMovesFocusBackward(t *testing.T) {
	_, _, win := appWithWindow(t, NewRect(5, 3, 40, 15), "FocusBack")

	btn1 := NewButton(NewRect(0, 0, 10, 1), "One", CmOK)
	btn2 := NewButton(NewRect(12, 0, 10, 1), "Two", CmCancel)
	win.Insert(btn1)
	win.Insert(btn2)

	// btn2 is focused after insertion (last inserted selectable).
	if win.FocusedChild() != btn2 {
		t.Fatal("pre-condition: btn2 should be focused before Shift+Tab")
	}

	ev := shiftTabKey()
	win.HandleEvent(ev)

	if win.FocusedChild() != btn1 {
		t.Errorf("after Shift+Tab, FocusedChild() = %v, want btn1", win.FocusedChild())
	}
	if !btn1.HasState(SfSelected) {
		t.Error("after Shift+Tab, btn1 should have SfSelected")
	}
	if btn2.HasState(SfSelected) {
		t.Error("after Shift+Tab, btn2 should have lost SfSelected")
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Tab skips non-selectable Label
// ---------------------------------------------------------------------------

// TestIntegrationTabSkipsNonSelectableLabel verifies that Tab traversal skips
// non-selectable widgets (like Label) and cycles only between selectable buttons.
//
// Spec req 4: "Tab with three items (Label non-selectable, Button A selectable,
// Button B selectable): Tab skips the Label and cycles only between the two buttons."
//
// Note: Tab is dispatched directly to the window (see TestIntegrationTabAdvancesFocusBetweenTwoButtons).
func TestIntegrationTabSkipsNonSelectableLabel(t *testing.T) {
	_, _, win := appWithWindow(t, NewRect(5, 3, 50, 15), "SkipLabel")

	// Label has no linked button in this test (we just need a non-selectable view).
	// NewLabel sets OfPreProcess but NOT OfSelectable.
	lbl := NewLabel(NewRect(0, 0, 8, 1), "Name:", nil)
	btnA := NewButton(NewRect(9, 0, 10, 1), "Alpha", CmOK)
	btnB := NewButton(NewRect(21, 0, 10, 1), "Beta", CmCancel)

	win.Insert(lbl)
	win.Insert(btnA)
	win.Insert(btnB)

	// Confirm label is not selectable.
	if lbl.HasOption(OfSelectable) {
		t.Fatal("pre-condition: Label should not be OfSelectable")
	}

	// btnB should be focused (last inserted selectable).
	if win.FocusedChild() != btnB {
		t.Fatal("pre-condition: btnB should be focused after insertion")
	}

	// Tab from btnB should wrap to btnA (skipping non-selectable label).
	ev := tabKey()
	win.HandleEvent(ev)

	focused := win.FocusedChild()
	if focused == lbl {
		t.Error("after Tab, Label (non-selectable) was focused — Tab should skip it")
	}
	if focused != btnA {
		t.Errorf("after Tab from btnB, FocusedChild() = %v, want btnA", focused)
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: Tab wraps from last button back to first
// ---------------------------------------------------------------------------

// TestIntegrationTabWrapsAroundFromLastToFirst verifies that Tab from the last
// button wraps back to the first button.
//
// Spec req 5: "Tab wraps around: with two buttons, Tab from the last goes back
// to the first."
//
// Note: Tab is dispatched directly to the window (see TestIntegrationTabAdvancesFocusBetweenTwoButtons).
func TestIntegrationTabWrapsAroundFromLastToFirst(t *testing.T) {
	_, _, win := appWithWindow(t, NewRect(5, 3, 40, 15), "Wrap")

	btn1 := NewButton(NewRect(0, 0, 10, 1), "First", CmOK)
	btn2 := NewButton(NewRect(12, 0, 10, 1), "Last", CmCancel)
	win.Insert(btn1)
	win.Insert(btn2)

	// btn2 is focused (last inserted selectable) — it is the last in tab order.
	if win.FocusedChild() != btn2 {
		t.Fatal("pre-condition: btn2 should be focused")
	}

	ev := tabKey()
	win.HandleEvent(ev)

	if win.FocusedChild() != btn1 {
		t.Errorf("after Tab from last button, FocusedChild() = %v, want btn1 (wrapped)", win.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: Default button fires via postprocess when non-button is focused
// ---------------------------------------------------------------------------

// TestIntegrationDefaultButtonFiresViaPostprocessWhenOtherWidgetFocused verifies
// that a default Button (WithDefault) responds to Enter in the postprocess phase
// even when a different (non-consuming) widget has focus.
//
// Spec req 6: "A default Button (WithDefault) in a Window responds to Enter in
// the postprocess phase even when a different widget has focus — meaning Enter
// pressed on a non-button focused view causes the default button to fire."
func TestIntegrationDefaultButtonFiresViaPostprocessWhenOtherWidgetFocused(t *testing.T) {
	app, _, win := appWithWindow(t, NewRect(5, 3, 40, 15), "Default")

	// Default button: gets OfPostProcess.
	defBtn := NewButton(NewRect(20, 0, 12, 1), "OK", CmOK, WithDefault())

	// A non-consuming selectable widget (BaseView does nothing with events).
	nonConsumer := &BaseView{}
	nonConsumer.SetBounds(NewRect(0, 0, 18, 1))
	nonConsumer.SetState(SfVisible, true)
	nonConsumer.SetOptions(OfSelectable, true)

	// Insert default button first, then the non-consumer (steals focus).
	win.Insert(defBtn)
	win.Insert(nonConsumer)

	if win.FocusedChild() != nonConsumer {
		t.Fatal("pre-condition: nonConsumer should be focused")
	}
	if !defBtn.HasOption(OfPostProcess) {
		t.Fatal("pre-condition: defBtn should have OfPostProcess")
	}

	ev := enterKey()
	app.handleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("default button did not fire via postprocess; ev.What = %v, want EvCommand", ev.What)
	}
	if ev.Command != CmOK {
		t.Errorf("default button postprocess fired wrong command; ev.Command = %v, want CmOK", ev.Command)
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: Label with link focuses the linked button via Alt+shortcut
// ---------------------------------------------------------------------------

// TestIntegrationLabelAltShortcutFocusesLinkedButton verifies that a Label with
// a tilde shortcut and a link to a Button causes Alt+shortcut to focus that Button.
// Label has OfPreProcess so it intercepts the Alt event before the focused child.
//
// Spec req 7: "A Label with link set to a Button: Alt+shortcut letter focuses the
// linked Button (Label is OfPreProcess, so it intercepts the Alt event before the
// focused child)."
func TestIntegrationLabelAltShortcutFocusesLinkedButton(t *testing.T) {
	app, _, win := appWithWindow(t, NewRect(5, 3, 50, 15), "LabelLink")

	btnTarget := NewButton(NewRect(10, 0, 12, 1), "Name", CmOK)

	// Label with '~N~ame:' → shortcut is 'n'. Link points to btnTarget.
	lbl := NewLabel(NewRect(0, 0, 7, 1), "~N~ame:", btnTarget)

	// Insert label first (it is OfPreProcess, not OfSelectable), then button.
	win.Insert(lbl)
	win.Insert(btnTarget)

	// btnTarget should be focused (it's the only selectable child).
	// Set focus to something else to verify the label re-focuses btnTarget.
	// Create a second selectable button to be the current focus.
	btnOther := NewButton(NewRect(25, 0, 12, 1), "Other", CmCancel)
	win.Insert(btnOther)

	// btnOther is now focused (last inserted selectable).
	if win.FocusedChild() != btnOther {
		t.Fatal("pre-condition: btnOther should be focused")
	}

	// Inject Alt+N — label intercepts this in preprocess and focuses btnTarget.
	ev := &Event{
		What: EvKeyboard,
		Key: &KeyEvent{
			Key:       tcell.KeyRune,
			Rune:      'n',
			Modifiers: tcell.ModAlt,
		},
	}
	app.handleEvent(ev)

	if win.FocusedChild() != btnTarget {
		t.Errorf("after Alt+N, FocusedChild() = %v, want btnTarget", win.FocusedChild())
	}
	if !btnTarget.HasState(SfSelected) {
		t.Error("after Alt+N, btnTarget should have SfSelected")
	}
	// Event should have been cleared by the label handler.
	if !ev.IsCleared() {
		t.Error("event should be cleared after label consumed Alt+N")
	}
}

// ---------------------------------------------------------------------------
// Requirement 8: StaticText drawn at correct position with word-wrapped text
// ---------------------------------------------------------------------------

// TestIntegrationStaticTextDrawnAtPositionWithWordWrap verifies that a StaticText
// widget inside a Window is drawn at its configured position, with its text
// content visible in the rendered output. Word-wrapping is implicitly exercised
// by using a narrow StaticText width with multi-word text.
//
// Spec req 8: "StaticText inside a Window is drawn at its position with
// word-wrapped text."
func TestIntegrationStaticTextDrawnAtPositionWithWordWrap(t *testing.T) {
	const bw, bh = 80, 25
	// Window at (2, 1, 30, 10). Client area starts at screen (3, 2).
	app, _, win := appWithWindow(t, NewRect(2, 1, 30, 10), "Text")

	// StaticText at client (0, 0) with narrow width (10) and two-word text.
	// "Hello World" — "Hello" (5) fits on row 0, "World" wraps to row 1.
	st := NewStaticText(NewRect(0, 0, 10, 3), "Hello World")
	win.Insert(st)

	buf := NewDrawBuffer(bw, bh)
	app.Draw(buf)

	// Client area origin in screen coords: winX+1=3, winY+1=2.
	// "Hello" should appear starting at screen (3, 2).
	clientX, clientY := 3, 2
	word := "Hello"
	for i, ch := range word {
		cell := buf.GetCell(clientX+i, clientY)
		if cell.Rune != ch {
			t.Errorf("StaticText at screen (%d,%d): rune = %q, want %q",
				clientX+i, clientY, cell.Rune, ch)
		}
	}

	// "World" should appear on the next line starting at screen (3, 3).
	word2 := "World"
	for i, ch := range word2 {
		cell := buf.GetCell(clientX+i, clientY+1)
		if cell.Rune != ch {
			t.Errorf("StaticText word-wrap at screen (%d,%d): rune = %q, want %q",
				clientX+i, clientY+1, cell.Rune, ch)
		}
	}
}

// ---------------------------------------------------------------------------
// Requirement 9: Drawing renders button bracket text at correct screen position
// ---------------------------------------------------------------------------

// TestIntegrationDrawRendersBracketTextAtCorrectPosition verifies that drawing an
// Application containing a Window with a Button renders the button's bracket text
// "[ OK ]" in the correct position within the window's client area.
//
// Spec req 9: "Drawing an Application containing a Window with a Button renders
// the button's bracket text [ OK ] in the correct position within the window's
// client area."
func TestIntegrationDrawRendersBracketTextAtCorrectPosition(t *testing.T) {
	const bw, bh = 80, 25
	// Window at (10, 5, 20, 8). Client area starts at screen (11, 6).
	app, _, win := appWithWindow(t, NewRect(10, 5, 20, 8), "Win")

	// Button at client (0, 0), width 6 exactly fits "[ OK ]".
	// "OK" title → bracketText = 2+4 = 6 chars → startX = (6-6)/2 = 0
	btn := NewButton(NewRect(0, 0, 6, 1), "OK", CmOK)
	win.Insert(btn)

	buf := NewDrawBuffer(bw, bh)
	app.Draw(buf)

	// Client area origin in screen coords: (10+1, 5+1) = (11, 6).
	// Button at client (0, 0) → screen (11, 6).
	// Expected layout: '[' ' ' 'O' 'K' ' ' ']' at x=11..16, y=6.
	clientX, clientY := 11, 6
	expected := []rune{'[', ' ', 'O', 'K', ' ', ']'}
	for i, want := range expected {
		cell := buf.GetCell(clientX+i, clientY)
		if cell.Rune != want {
			t.Errorf("button bracket at screen (%d,%d): rune = %q, want %q",
				clientX+i, clientY, cell.Rune, want)
		}
	}
}

// ---------------------------------------------------------------------------
// Task 9 integration tests — Dialog + ExecView
// ---------------------------------------------------------------------------

// appWithDesktopAndScreen creates an Application with the full Application→Desktop
// stack and returns the screen so callers can inject events. Unlike the phase-2
// helper appWithDesktop, this one returns the screen explicitly (no auto-cleanup)
// so ExecView goroutines can share the screen reference safely.
func appWithDesktopAndScreen(t *testing.T) (*Application, *Desktop, tcell.SimulationScreen) {
	t.Helper()
	screen := newTestScreen(t)
	app, err := NewApplication(
		WithScreen(screen),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		screen.Fini()
		t.Fatalf("NewApplication: %v", err)
	}
	return app, app.Desktop(), screen
}

// ---------------------------------------------------------------------------
// Requirement 1: Application→Desktop ExecView with OK button — Enter returns CmOK
// Spec req 1: "calling desktop.ExecView(dialog) where dialog has an OK button,
//
//	pressing Enter returns CmOK"
//
// ---------------------------------------------------------------------------

// TestIntegrationExecViewDesktopOKButtonEnterReturnsCmOK verifies the full
// Application→Desktop modal flow: ExecView running on the Desktop, with an OK
// button that has WithDefault(), returns CmOK when Enter is injected via
// screen.InjectKey so the event flows through Button.HandleEvent.
func TestIntegrationExecViewDesktopOKButtonEnterReturnsCmOK(t *testing.T) {
	_, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(10, 5, 40, 12), "Confirm")
	btn := NewButton(NewRect(5, 3, 12, 2), "OK", CmOK, WithDefault())
	dlg.Insert(btn)

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	// Wait for ExecView to enter its modal polling loop.
	time.Sleep(50 * time.Millisecond)

	// InjectKey sends a real tcell key event through the simulation screen's
	// PollEvent queue — Button.HandleEvent transforms Enter into CmOK in-place.
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	select {
	case cmd := <-result:
		if cmd != CmOK {
			t.Errorf("ExecView returned %v, want CmOK", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return within 2 s after Enter injection")
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: ExecView OK+Cancel — Tab moves focus to Cancel, Enter returns CmCancel
// Spec req 2: "ExecView with OK+Cancel: pressing Tab moves focus to Cancel,
//
//	pressing Enter returns CmCancel"
//
// ---------------------------------------------------------------------------

// TestIntegrationExecViewTabMovesFocusToCancelThenEnterReturnsCmCancel verifies
// that Tab inside the modal dialog moves focus from the default OK button to the
// Cancel button, and that a subsequent Enter returns CmCancel.
func TestIntegrationExecViewTabMovesFocusToCancelThenEnterReturnsCmCancel(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(5, 5, 50, 12), "Choose")
	okBtn := NewButton(NewRect(5, 3, 12, 2), "OK", CmOK, WithDefault())
	cancelBtn := NewButton(NewRect(20, 3, 14, 2), "Cancel", CmCancel)
	dlg.Insert(okBtn)
	dlg.Insert(cancelBtn)

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// After insertion the last-inserted selectable child (cancelBtn) is focused.
	// Tab from cancelBtn wraps to okBtn; to land on cancelBtn we need one Tab.
	// But insertion order means okBtn was inserted first — cancelBtn has focus.
	// One Tab advances focus: cancelBtn → okBtn (wrap). Two Tabs: okBtn → cancelBtn.
	// We want cancelBtn focused, so inject Tab once to move from cancelBtn → okBtn,
	// then Tab again to move back to cancelBtn.
	screen.InjectKey(tcell.KeyTab, 0, tcell.ModNone)
	time.Sleep(20 * time.Millisecond)
	screen.InjectKey(tcell.KeyTab, 0, tcell.ModNone)
	time.Sleep(20 * time.Millisecond)

	// Now cancelBtn is focused; Enter fires CmCancel.
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	select {
	case cmd := <-result:
		if cmd != CmCancel {
			t.Errorf("ExecView returned %v, want CmCancel", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return within 2 s after Tab+Enter")
	}

	_ = app // keep app alive
}

// ---------------------------------------------------------------------------
// Requirement 3: ExecView default button fires via postprocess when non-button focused
// Spec req 3: "ExecView with default button: pressing Enter when a non-button view
//
//	is focused fires the default button via postprocess"
//
// ---------------------------------------------------------------------------

// TestIntegrationExecViewDefaultButtonFiresViaPostprocessInsideDialog verifies
// that inside a modal Dialog, when a non-consuming selectable view has focus,
// pressing Enter causes the default button (OfPostProcess) to fire CmOK.
func TestIntegrationExecViewDefaultButtonFiresViaPostprocessInsideDialog(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(5, 5, 50, 12), "Postprocess")

	// Default button: receives Enter in Phase 3 (postprocess) only.
	defBtn := NewButton(NewRect(20, 3, 12, 2), "OK", CmOK, WithDefault())

	// Non-consuming selectable view that steals focus (inserted after defBtn).
	nonConsumer := &BaseView{}
	nonConsumer.SetBounds(NewRect(2, 1, 14, 1))
	nonConsumer.SetState(SfVisible, true)
	nonConsumer.SetOptions(OfSelectable, true)

	dlg.Insert(defBtn)
	dlg.Insert(nonConsumer) // now focused (last selectable inserted)

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Enter reaches nonConsumer (does nothing), then postprocess delivers to defBtn.
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	select {
	case cmd := <-result:
		if cmd != CmOK {
			t.Errorf("ExecView with default-button postprocess returned %v, want CmOK", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return within 2 s — default button postprocess may not be firing")
	}

	_ = app
}

// ---------------------------------------------------------------------------
// Requirement 4: MessageBox MbYes|MbNo returns CmYes when Enter pressed
// Spec req 4: "MessageBox with MbYes|MbNo returns CmYes when Enter pressed"
// ---------------------------------------------------------------------------

// TestIntegrationMessageBoxYesNoReturnsYesOnEnter verifies that MessageBox called
// with MbYes|MbNo uses ExecView and can return CmYes. MessageBox inserts Yes first
// (with WithDefault) and No second; after insertion No has focus. Tab moves focus
// back to Yes (wrap-around), and Enter then fires the Yes button returning CmYes.
func TestIntegrationMessageBoxYesNoReturnsYesOnEnter(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	result := make(chan CommandCode, 1)
	go func() {
		// MessageBox calls owner.ExecView internally; owner here is the Desktop.
		result <- MessageBox(desktop, "Question", "Delete file?", MbYes|MbNo)
	}()

	time.Sleep(50 * time.Millisecond)

	// MessageBox inserts Yes (default, i=0) then No (i=1). After insertion No is
	// focused (last selectable). Tab wraps: No → Yes. Enter then fires Yes → CmYes.
	screen.InjectKey(tcell.KeyTab, 0, tcell.ModNone)
	time.Sleep(20 * time.Millisecond)
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	select {
	case cmd := <-result:
		if cmd != CmYes {
			t.Errorf("MessageBox(MbYes|MbNo) Tab+Enter = %v, want CmYes", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("MessageBox did not return within 2 s after Tab+Enter")
	}

	_ = app
}

// ---------------------------------------------------------------------------
// Requirement 5: Dialog renders with DialogFrame style (double-line border)
// Spec req 5: "Dialog renders with DialogFrame style (double-line border, not single-line)"
// ---------------------------------------------------------------------------

// TestIntegrationDialogRendersWithDoubleLineBorder verifies that when a Dialog is
// drawn as part of the full Application→Desktop render stack (via app.Draw), its
// border characters are double-line (╔ ╗ ╚ ╝ ═ ║) and not single-line (┌ ┐ └ ┘).
func TestIntegrationDialogRendersWithDoubleLineBorder(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	// Dialog at absolute desktop position (10, 5, 30, 10).
	dlg := NewDialog(NewRect(10, 5, 30, 10), "Border Test")
	desktop.Insert(dlg)

	const bw, bh = 80, 25
	buf := NewDrawBuffer(bw, bh)
	app.Draw(buf)

	// Dialog occupies desktop rows 5-14, cols 10-39 (30 wide, 10 tall).
	// Check the four corners for double-line runes.
	corners := []struct {
		x, y int
		want rune
		desc string
	}{
		{10, 5, '╔', "top-left corner"},
		{39, 5, '╗', "top-right corner"},
		{10, 14, '╚', "bottom-left corner"},
		{39, 14, '╝', "bottom-right corner"},
	}
	for _, c := range corners {
		cell := buf.GetCell(c.x, c.y)
		if cell.Rune != c.want {
			t.Errorf("dialog %s at (%d,%d): rune = %q, want %q (double-line border)",
				c.desc, c.x, c.y, cell.Rune, c.want)
		}
	}

	// Sample top horizontal edge at a column that falls outside the title text.
	// Dialog is 30 wide at x=10; dialog-local col 2 → screen col 12.
	// Title "Border Test" (11 chars) with padding = 13; centered in 28 cols:
	// titleX (dialog-local) = 1 + (28-13)/2 = 8, so title occupies local cols 8-20.
	// Screen col 12 = dialog-local col 2, which is before the title — must be '═'.
	edgeCell := buf.GetCell(12, 5)
	if edgeCell.Rune != '═' {
		t.Errorf("dialog top edge at (12,5): rune = %q, want '═' (double-line)", edgeCell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: Dialog renders with DialogBackground style for client area
// Spec req 6: "Dialog renders with DialogBackground style for client area"
// ---------------------------------------------------------------------------

// TestIntegrationDialogRendersWithDialogBackgroundInClientArea verifies that the
// client area of a Dialog (one cell inside the frame) is filled with the
// DialogBackground style from the active ColorScheme.
func TestIntegrationDialogRendersWithDialogBackgroundInClientArea(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	// Dialog at (10, 5, 30, 10). Client area: cols 11-38, rows 6-13.
	dlg := NewDialog(NewRect(10, 5, 30, 10), "BG Test")
	desktop.Insert(dlg)

	const bw, bh = 80, 25
	buf := NewDrawBuffer(bw, bh)
	app.Draw(buf)

	// Sample a cell inside the client area (no child inserted, so it stays background).
	// Client origin in screen coords: (10+1, 5+1) = (11, 6).
	cell := buf.GetCell(11, 6)

	cs := theme.BorlandBlue
	if cell.Style != cs.DialogBackground {
		t.Errorf("dialog client area cell(11,6): style = %v, want DialogBackground %v",
			cell.Style, cs.DialogBackground)
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: Mouse click inside modal dialog fires command; outside discarded
// Spec req 7: "Mouse click inside modal dialog on a button fires the command;
//
//	mouse outside is discarded"
//
// ---------------------------------------------------------------------------

// TestIntegrationExecViewMouseClickInsideDialogOnButtonFiresCommand verifies that
// a mouse click at the screen position of a button inside the modal dialog causes
// ExecView to receive and process the EvCommand, returning CmOK.
func TestIntegrationExecViewMouseClickInsideDialogOnButtonFiresCommand(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	// Dialog at absolute desktop coords (10, 5, 40, 12).
	// Client area: cols 11-48, rows 6-15 (frame offset -1 each side).
	// Button at client (5, 3, 12, 2) → absolute (16, 9) through (27, 10).
	dlg := NewDialog(NewRect(10, 5, 40, 12), "Click Test")
	btn := NewButton(NewRect(5, 3, 12, 2), "OK", CmOK, WithDefault())
	dlg.Insert(btn)

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// ExecView receives mouse events in dialog-local coordinates (it subtracts
	// dialog.Bounds().A before forwarding to v.HandleEvent).
	// Dialog bounds: A=(10,5). A click at absolute (17, 9) is dialog-local (7, 4).
	// That is inside the frame (cols 1-38, rows 1-10 for a 40×12 dialog).
	// Dialog.HandleEvent then subtracts (1,1) → group-local (6, 3) hitting the button.
	app.screen.PostEvent(tcell.NewEventMouse(17, 9, tcell.Button1, tcell.ModNone))

	select {
	case cmd := <-result:
		if cmd != CmOK {
			t.Errorf("mouse click on modal button returned %v, want CmOK", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return within 2 s after mouse click on button")
	}
}

// TestIntegrationExecViewMouseClickOutsideDialogIsDiscarded verifies that a mouse
// click at a screen position outside the modal dialog's bounds does not close the
// dialog or forward the event to the dialog's children.
func TestIntegrationExecViewMouseClickOutsideDialogIsDiscarded(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	// Dialog at (20, 8, 30, 10) — occupies cols 20-49, rows 8-17.
	dlg := NewDialog(NewRect(20, 8, 30, 10), "Outside Click")
	btn := NewButton(NewRect(5, 3, 12, 2), "OK", CmOK, WithDefault())
	dlg.Insert(btn)

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Click at (2, 2) — well outside the dialog bounds.
	app.screen.PostEvent(tcell.NewEventMouse(2, 2, tcell.Button1, tcell.ModNone))

	// Give the loop time to process the outside click (it should be ignored).
	time.Sleep(30 * time.Millisecond)

	// The dialog should still be running. Terminate cleanly via PostCommand.
	app.PostCommand(CmCancel, nil)

	select {
	case cmd := <-result:
		// We expect CmCancel from our PostCommand, NOT from the outside click.
		if cmd != CmCancel {
			t.Errorf("expected CmCancel from PostCommand, got %v", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return within 2 s")
	}
}

// ---------------------------------------------------------------------------
// Requirement 8: Tab traversal works inside modal dialog (cycling between buttons)
// Spec req 8: "Tab traversal works inside the modal dialog (cycling between buttons)"
// ---------------------------------------------------------------------------

// TestIntegrationExecViewTabTraversalCyclesBetweenButtonsInDialog verifies that
// Tab keys injected while ExecView is running move focus among the dialog's buttons,
// so that after a full cycle the originally-focused button can be activated with Enter.
func TestIntegrationExecViewTabTraversalCyclesBetweenButtonsInDialog(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(5, 5, 50, 12), "Tab Cycle")
	// Insert three buttons; after insertion btn3 is focused (last selectable).
	btn1 := NewButton(NewRect(2, 3, 12, 2), "One", CmOK, WithDefault())
	btn2 := NewButton(NewRect(16, 3, 12, 2), "Two", CmCancel)
	btn3 := NewButton(NewRect(30, 3, 12, 2), "Three", CmClose)
	dlg.Insert(btn1)
	dlg.Insert(btn2)
	dlg.Insert(btn3) // focused after insertion

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Tab from btn3 → btn1 (wrap), Tab → btn2, Tab → btn3 (full cycle).
	// Then Enter fires btn3's command (CmClose).
	screen.InjectKey(tcell.KeyTab, 0, tcell.ModNone) // btn3 → btn1
	time.Sleep(20 * time.Millisecond)
	screen.InjectKey(tcell.KeyTab, 0, tcell.ModNone) // btn1 → btn2
	time.Sleep(20 * time.Millisecond)
	screen.InjectKey(tcell.KeyTab, 0, tcell.ModNone) // btn2 → btn3
	time.Sleep(20 * time.Millisecond)
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone) // fires btn3 → CmClose

	select {
	case cmd := <-result:
		if cmd != CmClose {
			t.Errorf("Tab cycle + Enter returned %v, want CmClose (btn3)", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return within 2 s after Tab cycle + Enter")
	}

	_ = app
}

// ---------------------------------------------------------------------------
// Requirement 9: After ExecView returns, dialog is no longer in owner's children list
// Spec req 9: "After ExecView returns, the dialog is no longer in owner's children list"
// ---------------------------------------------------------------------------

// TestIntegrationExecViewDialogRemovedFromDesktopChildrenAfterReturn verifies that
// once ExecView completes (either via a closing command or screen finalization),
// the dialog is absent from the Desktop's children list.
func TestIntegrationExecViewDialogRemovedFromDesktopChildrenAfterReturn(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(10, 5, 30, 10), "Cleanup")
	btn := NewButton(NewRect(5, 3, 12, 2), "OK", CmOK, WithDefault())
	dlg.Insert(btn)

	done := make(chan CommandCode, 1)
	go func() {
		done <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmOK, nil)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}

	for _, child := range desktop.Children() {
		if child == dlg {
			t.Error("dialog is still present in Desktop.Children() after ExecView returned — it should have been removed")
		}
	}
}

// TestIntegrationExecViewDialogRemovedFromWindowChildrenAfterReturn verifies the
// same cleanup guarantee when ExecView is called on a Window rather than a Desktop.
// This exercises the full Application→Desktop→Window owner-chain walk inside ExecView.
func TestIntegrationExecViewDialogRemovedFromWindowChildrenAfterReturn(t *testing.T) {
	app, _, win := appWithWindow(t, NewRect(5, 3, 60, 20), "Host")

	dlg := NewDialog(NewRect(2, 2, 30, 10), "Cleanup")
	btn := NewButton(NewRect(5, 3, 12, 2), "OK", CmOK, WithDefault())
	dlg.Insert(btn)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmOK, nil)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}

	for _, child := range win.Children() {
		if child == dlg {
			t.Error("dialog still in Window.Children() after ExecView returned — should have been removed")
		}
	}
}

// ---------------------------------------------------------------------------
// Additional coverage: MessageBox MbOK, Shift+Tab traversal, SfModal cleanup
// ---------------------------------------------------------------------------

// TestIntegrationMessageBoxOKReturnsCmOK verifies the simplest MessageBox case:
// MbOK shows a single OK button, and Enter returns CmOK.
func TestIntegrationMessageBoxOKReturnsCmOK(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(desktop, "Info", "Operation complete.", MbOK)
	}()

	time.Sleep(50 * time.Millisecond)
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	select {
	case cmd := <-result:
		if cmd != CmOK {
			t.Errorf("MessageBox(MbOK) + Enter = %v, want CmOK", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("MessageBox did not return within 2 s after Enter")
	}

	_ = app
}

// TestIntegrationExecViewShiftTabMovesBackwardInDialog verifies that Shift+Tab
// moves focus backward among dialog buttons, so the correct button fires on Enter.
func TestIntegrationExecViewShiftTabMovesBackwardInDialog(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(5, 5, 50, 12), "Shift-Tab")
	btn1 := NewButton(NewRect(2, 3, 12, 2), "Yes", CmYes, WithDefault())
	btn2 := NewButton(NewRect(16, 3, 12, 2), "No", CmNo)
	dlg.Insert(btn1)
	dlg.Insert(btn2) // btn2 focused after insertion

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Shift+Tab from btn2 → btn1.
	screen.InjectKey(tcell.KeyBacktab, 0, tcell.ModNone)
	time.Sleep(20 * time.Millisecond)
	// Enter fires btn1 → CmYes.
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	select {
	case cmd := <-result:
		if cmd != CmYes {
			t.Errorf("Shift+Tab then Enter returned %v, want CmYes", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return within 2 s after Shift+Tab + Enter")
	}

	_ = app
}

// TestIntegrationExecViewSfModalClearedAfterClose verifies that SfModal is no
// longer set on the dialog after ExecView returns — even when closed by a
// direct PostCommand rather than a key event.
func TestIntegrationExecViewSfModalClearedAfterClose(t *testing.T) {
	app, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(10, 5, 30, 10), "SfModal Cleanup")

	done := make(chan CommandCode, 1)
	go func() {
		done <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Verify SfModal is set while ExecView is running.
	if !dlg.HasState(SfModal) {
		t.Error("SfModal should be set on dialog while ExecView is running")
	}

	app.PostCommand(CmClose, nil)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}

	if dlg.HasState(SfModal) {
		t.Error("SfModal is still set after ExecView returned — should have been cleared")
	}
}
