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
