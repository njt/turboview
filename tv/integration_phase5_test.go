package tv

// integration_phase5_test.go — Integration tests for Phase 5 form widgets in Dialogs.
//
// Each test verifies one requirement from the Phase 5 plan using REAL components
// wired end-to-end: Application → Desktop → Window → Dialog → form widget.
//
// Test naming: TestIntegrationPhase5<DescriptiveSuffix>.
//
// Conventions inherited from other integration test files in this package:
//   - newTestScreen(t)       — 80×25 SimulationScreen helper from application_test.go
//   - execViewStack(t)       — Application+Desktop+Window stack from dialog_test.go
//   - enterKey / tabKey / shiftTabKey — helpers from integration_phase3_test.go
//   - clickEvent             — helper from window_interaction_test.go

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Test 1: InputLine inside a Dialog receives keystrokes through the full
// dispatch chain (Application → Desktop → Dialog → Group → InputLine).
// ---------------------------------------------------------------------------

// TestIntegrationPhase5InputLineReceivesKeystrokes verifies that when an InputLine
// is focused inside a Dialog which is being run via ExecView, typing a rune
// reaches the InputLine through the full dispatch chain.
func TestIntegrationPhase5InputLineReceivesKeystrokes(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(5, 3, 50, 10), "Input Test")
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	dialog.Insert(il)

	// Focus the InputLine inside the dialog.
	dialog.SetFocusedChild(il)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	// Give ExecView time to enter its loop.
	time.Sleep(30 * time.Millisecond)

	// Inject keystrokes via the screen — they travel through the full Application
	// event-processing chain into the modal ExecView loop.
	app.screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	app.screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	time.Sleep(30 * time.Millisecond)

	// Verify text reached the InputLine.
	got := il.Text()
	if got != "hi" {
		t.Errorf("InputLine.Text() = %q after typing 'h','i' through dispatch chain, want %q", got, "hi")
	}

	// Clean up.
	app.PostCommand(CmCancel, nil)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}
}

// ---------------------------------------------------------------------------
// Test 2: Tab cycles focus between InputLine, CheckBoxes, and Button inside
// a Dialog.
// ---------------------------------------------------------------------------

// TestIntegrationPhase5TabCyclesBetweenFormWidgets verifies that Tab key
// advances focus through InputLine → CheckBoxes → Button inside a Dialog.
func TestIntegrationPhase5TabCyclesBetweenFormWidgets(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 60, 20), "Form")

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	cbs := NewCheckBoxes(NewRect(0, 2, 20, 2), []string{"~A~lpha", "~B~eta"})
	btn := NewButton(NewRect(0, 5, 10, 1), "OK", CmOK)

	d.Insert(il)
	d.Insert(cbs)
	d.Insert(btn)

	// Start with focus on InputLine.
	d.SetFocusedChild(il)
	if d.FocusedChild() != il {
		t.Fatal("pre-condition: InputLine should be focused")
	}

	// Tab → should move to CheckBoxes.
	d.HandleEvent(tabKey())
	if d.FocusedChild() != cbs {
		t.Errorf("after first Tab, FocusedChild() = %v, want CheckBoxes", d.FocusedChild())
	}

	// Tab → should move to Button.
	d.HandleEvent(tabKey())
	if d.FocusedChild() != btn {
		t.Errorf("after second Tab, FocusedChild() = %v, want Button", d.FocusedChild())
	}

	// Tab → wraps back to InputLine.
	d.HandleEvent(tabKey())
	if d.FocusedChild() != il {
		t.Errorf("after third Tab (wrap), FocusedChild() = %v, want InputLine", d.FocusedChild())
	}
}

// TestIntegrationPhase5ShiftTabCyclesBackward verifies that Shift+Tab moves
// focus backward through form widgets inside a Dialog.
func TestIntegrationPhase5ShiftTabCyclesBackward(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 60, 20), "Form")

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	cbs := NewCheckBoxes(NewRect(0, 2, 20, 2), []string{"~A~lpha", "~B~eta"})
	btn := NewButton(NewRect(0, 5, 10, 1), "OK", CmOK)

	d.Insert(il)
	d.Insert(cbs)
	d.Insert(btn)

	// Start at Button (last), Shift+Tab should go to CheckBoxes.
	d.SetFocusedChild(btn)
	d.HandleEvent(shiftTabKey())
	if d.FocusedChild() != cbs {
		t.Errorf("after Shift+Tab from Button, FocusedChild() = %v, want CheckBoxes", d.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Test 3: CheckBox inside a Dialog toggles when Space is pressed while focused.
// ---------------------------------------------------------------------------

// TestIntegrationPhase5CheckBoxTogglesOnSpace verifies that a Space key press
// reaches the focused CheckBox inside a Dialog and toggles its state.
func TestIntegrationPhase5CheckBoxTogglesOnSpace(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "Check Test")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	d.Insert(cbs)
	d.SetFocusedChild(cbs)

	// Identify the currently focused CheckBox (last inserted by default).
	focusedCB := cbs.FocusedChild()
	if focusedCB == nil {
		t.Fatal("CheckBoxes.FocusedChild() returned nil — no CheckBox is focused")
	}
	cb, ok := focusedCB.(*CheckBox)
	if !ok {
		t.Fatalf("FocusedChild() is not a *CheckBox: %T", focusedCB)
	}

	// Pre-condition: focused CheckBox must be unchecked.
	if cb.Checked() {
		t.Fatal("pre-condition: focused CheckBox should start unchecked")
	}

	// Space on the CheckBoxes cluster: the cluster forwards Space to its focused CheckBox.
	spaceEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	d.HandleEvent(spaceEv)

	if !cb.Checked() {
		t.Error("Space through Dialog → CheckBoxes did not toggle the focused CheckBox")
	}
}

// TestIntegrationPhase5CheckBoxSecondToggleRestoresState verifies that a second
// Space press untoggled the CheckBox (round-trip through the Dialog dispatch chain).
func TestIntegrationPhase5CheckBoxSecondToggleRestoresState(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "Check Test")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	d.Insert(cbs)
	d.SetFocusedChild(cbs)

	for i := 0; i < 2; i++ {
		spaceEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
		d.HandleEvent(spaceEv)
	}

	if cbs.Item(0).Checked() {
		t.Error("two Space presses through Dialog should leave CheckBox unchecked (toggle twice)")
	}
}

// ---------------------------------------------------------------------------
// Test 4: RadioButtons cluster inside a Dialog allows Up/Down selection.
// ---------------------------------------------------------------------------

// TestIntegrationPhase5RadioButtonsDownArrowSelectsNext verifies Down arrow
// reaches the RadioButtons cluster through the Dialog dispatch chain and selects
// the next button.
func TestIntegrationPhase5RadioButtonsDownArrowSelectsNext(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "Radio Test")

	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"~A~lpha", "~B~eta", "~G~amma"})
	d.Insert(rbs)
	d.SetFocusedChild(rbs)

	// Item(0) is selected by default. Down should select Item(1).
	downEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	d.HandleEvent(downEv)

	if !rbs.Item(1).Selected() {
		t.Error("Down arrow through Dialog → RadioButtons did not select Item(1)")
	}
	if rbs.Item(0).Selected() {
		t.Error("Item(0) should be deselected after Down arrow")
	}
}

// TestIntegrationPhase5RadioButtonsUpArrowSelectsPrevious verifies Up arrow
// travels through the Dialog dispatch chain and selects the previous radio button.
func TestIntegrationPhase5RadioButtonsUpArrowSelectsPrevious(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "Radio Test")

	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"~A~lpha", "~B~eta", "~G~amma"})
	rbs.SetValue(2) // start at last item
	rbs.SetFocusedChild(rbs.Item(2))

	d.Insert(rbs)
	d.SetFocusedChild(rbs)

	upEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	d.HandleEvent(upEv)

	if !rbs.Item(1).Selected() {
		t.Error("Up arrow through Dialog → RadioButtons did not select Item(1)")
	}
}

// TestIntegrationPhase5RadioButtonsUpDownSequenceKeepsExclusiveSelection verifies
// that after a sequence of Up/Down arrow presses through the Dialog dispatch
// chain, exactly one RadioButton remains selected.
func TestIntegrationPhase5RadioButtonsUpDownSequenceKeepsExclusiveSelection(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "Radio Test")

	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	d.Insert(rbs)
	d.SetFocusedChild(rbs)

	down := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	up := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}

	for i := 0; i < 3; i++ {
		d.HandleEvent(down)
	}
	for i := 0; i < 2; i++ {
		d.HandleEvent(up)
	}

	selected := 0
	for i := 0; i < 3; i++ {
		if rbs.Item(i).Selected() {
			selected++
		}
	}
	if selected != 1 {
		t.Errorf("after Up/Down through Dialog, %d RadioButtons selected; want exactly 1", selected)
	}
}

// ---------------------------------------------------------------------------
// Test 5: Alt+shortcut on a Label linked to an InputLine focuses the InputLine
// inside a Dialog.
// ---------------------------------------------------------------------------

// TestIntegrationPhase5LabelAltShortcutFocusesInputLine verifies that
// Alt+<shortcut> on a Label linked to an InputLine, when dispatched through
// a Dialog, moves focus to that InputLine.
func TestIntegrationPhase5LabelAltShortcutFocusesInputLine(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 60, 20), "Label Test")

	il := NewInputLine(NewRect(10, 0, 20, 1), 0)
	lbl := NewLabel(NewRect(0, 0, 9, 1), "~N~ame:", il)
	btn := NewButton(NewRect(0, 2, 10, 1), "OK", CmOK)

	d.Insert(lbl)
	d.Insert(il)
	d.Insert(btn)

	// Start with focus on Button.
	d.SetFocusedChild(btn)
	if d.FocusedChild() != btn {
		t.Fatal("pre-condition: Button should be focused initially")
	}

	// Send Alt+N — should trigger Label's shortcut handler and focus the InputLine.
	altNEv := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	d.HandleEvent(altNEv)

	if d.FocusedChild() != il {
		t.Errorf("after Alt+N, FocusedChild() = %v, want InputLine (linked to Label '~N~ame:')", d.FocusedChild())
	}
}

// TestIntegrationPhase5LabelAltShortcutCaseInsensitive verifies that Alt+N
// (uppercase) also focuses the InputLine linked by the lowercase-shortcut label.
func TestIntegrationPhase5LabelAltShortcutCaseInsensitive(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 60, 20), "Label Test")

	il := NewInputLine(NewRect(10, 0, 20, 1), 0)
	lbl := NewLabel(NewRect(0, 0, 9, 1), "~N~ame:", il)
	btn := NewButton(NewRect(0, 2, 10, 1), "OK", CmOK)

	d.Insert(lbl)
	d.Insert(il)
	d.Insert(btn)

	d.SetFocusedChild(btn)

	// Alt+N uppercase.
	altNEv := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'N', Modifiers: tcell.ModAlt},
	}
	d.HandleEvent(altNEv)

	if d.FocusedChild() != il {
		t.Errorf("after Alt+N (uppercase), FocusedChild() = %v, want InputLine", d.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Test 6: ColorScheme inheritance — InputLine inside a Dialog with a custom
// ColorScheme uses that scheme's InputNormal/InputSelection styles.
// ---------------------------------------------------------------------------

// TestIntegrationPhase5InputLineInheritsDialogColorScheme verifies that an
// InputLine inserted into a Dialog inherits the Dialog's ColorScheme for
// rendering (InputNormal style).
func TestIntegrationPhase5InputLineInheritsDialogColorScheme(t *testing.T) {
	// Build a custom ColorScheme with a distinctive InputNormal style.
	customScheme := &theme.ColorScheme{
		InputNormal:      tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlue),
		InputSelection:   tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorRed),
		DialogBackground: tcell.StyleDefault.Background(tcell.ColorNavy),
		DialogFrame:      tcell.StyleDefault.Foreground(tcell.ColorWhite),
		WindowTitle:      tcell.StyleDefault.Foreground(tcell.ColorYellow),
	}

	d := NewDialog(NewRect(0, 0, 40, 15), "Scheme Test")
	d.scheme = customScheme

	// InputLine with no scheme of its own — should inherit from Dialog.
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	d.Insert(il)

	// ColorScheme() should propagate from Dialog to InputLine via Owner chain.
	cs := il.ColorScheme()
	if cs == nil {
		t.Fatal("InputLine.ColorScheme() returned nil; expected inheritance from Dialog")
	}
	if cs != customScheme {
		t.Errorf("InputLine.ColorScheme() = %v, want Dialog's custom scheme %v", cs, customScheme)
	}
}

// TestIntegrationPhase5InputLineRendersWithInheritedScheme verifies that
// Draw actually uses the inherited scheme by checking the style of drawn cells.
// We use BorlandBlue as the Dialog's scheme since we know its InputNormal style
// is distinct from tcell.StyleDefault, avoiding any color encoding ambiguity.
func TestIntegrationPhase5InputLineRendersWithInheritedScheme(t *testing.T) {
	inheritedScheme := theme.BorlandBlue

	if inheritedScheme.InputNormal == tcell.StyleDefault {
		t.Skip("BorlandBlue.InputNormal equals StyleDefault — style inheritance test would be vacuous")
	}

	d := NewDialog(NewRect(0, 0, 40, 15), "Scheme Test")
	d.scheme = inheritedScheme

	// InputLine with no scheme of its own — should inherit from Dialog.
	il := NewInputLine(NewRect(0, 0, 10, 1), 0)
	d.Insert(il)

	// Verify scheme inheritance: ColorScheme() on the InputLine returns the dialog's scheme.
	cs := il.ColorScheme()
	if cs == nil {
		t.Fatal("InputLine.ColorScheme() returned nil after being inserted into a Dialog with a scheme")
	}
	if cs != inheritedScheme {
		t.Errorf("InputLine.ColorScheme() != dialog's scheme; inheritance is broken")
	}

	// Draw the InputLine and verify it uses the inherited InputNormal style.
	// Note: the InputLine was inserted into the Dialog so it has SfSelected set
	// (it's the focused child). With an empty text and cursor at position 0,
	// cell 0 uses InputSelection (the cursor indicator). Cells 1+ use InputNormal.
	buf := NewDrawBuffer(10, 1)
	il.Draw(buf)

	// Check cells beyond the cursor position (col 0 has the cursor indicator).
	for x := 1; x < 10; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Style != inheritedScheme.InputNormal {
			t.Errorf("cell(%d,0) style = %v, want BorlandBlue.InputNormal %v", x, cell.Style, inheritedScheme.InputNormal)
			break
		}
	}
}

// TestIntegrationPhase5CheckBoxInheritsDialogColorScheme verifies that a
// CheckBox inside a Dialog inherits the Dialog's ColorScheme.
func TestIntegrationPhase5CheckBoxInheritsDialogColorScheme(t *testing.T) {
	customScheme := &theme.ColorScheme{
		CheckBoxNormal:   tcell.StyleDefault.Foreground(tcell.ColorGreen),
		DialogBackground: tcell.StyleDefault.Background(tcell.ColorNavy),
		DialogFrame:      tcell.StyleDefault.Foreground(tcell.ColorWhite),
		WindowTitle:      tcell.StyleDefault.Foreground(tcell.ColorYellow),
	}

	d := NewDialog(NewRect(0, 0, 40, 15), "Scheme Test")
	d.scheme = customScheme

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"Option 1", "Option 2"})
	d.Insert(cbs)

	// The CheckBoxes cluster (and its children) should inherit the dialog's scheme.
	cs := cbs.ColorScheme()
	if cs == nil {
		t.Fatal("CheckBoxes.ColorScheme() returned nil; expected inheritance from Dialog")
	}
	if cs != customScheme {
		t.Errorf("CheckBoxes.ColorScheme() = %v, want Dialog's custom scheme", cs)
	}
}

// ---------------------------------------------------------------------------
// Test 7: ExecView on a Dialog containing form widgets returns the correct
// command code when OK/Cancel is pressed.
// ---------------------------------------------------------------------------

// TestIntegrationPhase5ExecViewWithFormWidgetsReturnsCmOK verifies that
// ExecView returns CmOK when the OK button in a Dialog with form widgets is
// activated.
func TestIntegrationPhase5ExecViewWithFormWidgetsReturnsCmOK(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(5, 3, 55, 18), "Form Dialog")

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	cbs := NewCheckBoxes(NewRect(0, 2, 20, 2), []string{"~S~ave", "~L~oad"})
	rbs := NewRadioButtons(NewRect(0, 5, 20, 3), []string{"~A~lpha", "~B~eta", "~G~amma"})
	okBtn := NewButton(NewRect(0, 9, 10, 1), "OK", CmOK)
	cancelBtn := NewButton(NewRect(12, 9, 12, 1), "Cancel", CmCancel)

	dialog.Insert(il)
	dialog.Insert(cbs)
	dialog.Insert(rbs)
	dialog.Insert(okBtn)
	dialog.Insert(cancelBtn)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmOK, nil)

	select {
	case code := <-done:
		if code != CmOK {
			t.Errorf("ExecView returned %v, want CmOK", code)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmOK within 2 s")
	}
}

// TestIntegrationPhase5ExecViewWithFormWidgetsReturnsCmCancel verifies that
// ExecView returns CmCancel when the Cancel button in a form Dialog is pressed.
func TestIntegrationPhase5ExecViewWithFormWidgetsReturnsCmCancel(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(5, 3, 55, 18), "Form Dialog")

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	cbs := NewCheckBoxes(NewRect(0, 2, 20, 2), []string{"~S~ave", "~L~oad"})
	okBtn := NewButton(NewRect(0, 5, 10, 1), "OK", CmOK)
	cancelBtn := NewButton(NewRect(12, 5, 12, 1), "Cancel", CmCancel)

	dialog.Insert(il)
	dialog.Insert(cbs)
	dialog.Insert(okBtn)
	dialog.Insert(cancelBtn)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmCancel, nil)

	select {
	case code := <-done:
		if code != CmCancel {
			t.Errorf("ExecView returned %v, want CmCancel", code)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmCancel within 2 s")
	}
}

// ---------------------------------------------------------------------------
// Test 8: InputLine text is preserved during modal dialog execution.
// Type text, press OK, verify Text() before dialog closes.
// ---------------------------------------------------------------------------

// TestIntegrationPhase5InputLineTextPreservedDuringExecView verifies that text
// typed into an InputLine inside an ExecView modal loop is still accessible on
// the InputLine widget after the modal command fires (before the caller discards
// the dialog).
func TestIntegrationPhase5InputLineTextPreservedDuringExecView(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(5, 3, 50, 12), "Text Dialog")
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	okBtn := NewButton(NewRect(0, 2, 10, 1), "OK", CmOK)

	dialog.Insert(il)
	dialog.Insert(okBtn)

	// Focus the InputLine in the dialog before launching ExecView.
	dialog.SetFocusedChild(il)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	// Give ExecView time to enter the modal loop.
	time.Sleep(30 * time.Millisecond)

	// Inject keystrokes that should reach the InputLine through the ExecView loop.
	app.screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	app.screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	app.screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	app.screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	app.screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	time.Sleep(30 * time.Millisecond)

	// Capture the text before ExecView exits.
	textBeforeClose := il.Text()

	// Now close the dialog.
	app.PostCommand(CmOK, nil)

	select {
	case code := <-done:
		if code != CmOK {
			t.Errorf("ExecView returned %v, want CmOK", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}

	// Text should have been "hello" at the time of closing.
	if textBeforeClose != "hello" {
		t.Errorf("InputLine.Text() before dialog close = %q, want %q", textBeforeClose, "hello")
	}

	// Text must also still be accessible after ExecView returns (dialog is not destroyed).
	if il.Text() != "hello" {
		t.Errorf("InputLine.Text() after ExecView returned = %q, want %q", il.Text(), "hello")
	}
}

// TestIntegrationPhase5InputLineTextNotAffectedByCancelClose verifies that the
// text typed into an InputLine is preserved even when Cancel closes the dialog.
// (The caller is responsible for ignoring the data on Cancel; the widget must
// not clear itself on modal close.)
func TestIntegrationPhase5InputLineTextNotAffectedByCancelClose(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(5, 3, 50, 12), "Text Dialog")
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	// Pre-set text programmatically to avoid relying on keystroke injection timing.
	il.SetText("preset")

	cancelBtn := NewButton(NewRect(0, 2, 12, 1), "Cancel", CmCancel)

	dialog.Insert(il)
	dialog.Insert(cancelBtn)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmCancel, nil)

	select {
	case code := <-done:
		if code != CmCancel {
			t.Errorf("ExecView returned %v, want CmCancel", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}

	// The text must be preserved after the modal closes.
	if il.Text() != "preset" {
		t.Errorf("InputLine.Text() after Cancel close = %q, want %q", il.Text(), "preset")
	}
}

// ---------------------------------------------------------------------------
// Additional integration tests: wiring correctness for the full component chain
// ---------------------------------------------------------------------------

// TestIntegrationPhase5DialogDispatchChainReachesInputLine verifies the dispatch
// chain works without a running ExecView — HandleEvent directly on the Dialog
// reaches the focused InputLine.
func TestIntegrationPhase5DialogDispatchChainReachesInputLine(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "Dispatch Test")
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	d.Insert(il)
	d.SetFocusedChild(il)

	// Send rune events directly.
	d.HandleEvent(runeEv('a'))
	d.HandleEvent(runeEv('b'))
	d.HandleEvent(runeEv('c'))

	if got := il.Text(); got != "abc" {
		t.Errorf("Dialog dispatch to InputLine: Text() = %q, want %q", got, "abc")
	}
}

// TestIntegrationPhase5DialogDispatchChainCursorMovement verifies cursor movement
// events travel through the Dialog dispatch chain to the InputLine.
func TestIntegrationPhase5DialogDispatchChainCursorMovement(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "Cursor Test")
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello") // cursor at position 5
	d.Insert(il)
	d.SetFocusedChild(il)

	// Home key through Dialog.
	d.HandleEvent(keyEv(tcell.KeyHome))

	if il.CursorPos() != 0 {
		t.Errorf("after Home through Dialog, CursorPos() = %d, want 0", il.CursorPos())
	}
}

// TestIntegrationPhase5RadioButtonsInsideDialogExclusiveSelection verifies that
// the exclusion invariant is maintained when selecting RadioButtons inside a Dialog.
func TestIntegrationPhase5RadioButtonsInsideDialogExclusiveSelection(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "Radio Dialog")

	rbs := NewRadioButtons(NewRect(0, 0, 20, 4), []string{"One", "Two", "Three", "Four"})
	d.Insert(rbs)
	d.SetFocusedChild(rbs)

	// Navigate Down three times from Item(0).
	for i := 0; i < 3; i++ {
		d.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	}

	// Only one RadioButton should be selected.
	selected := 0
	for i := 0; i < 4; i++ {
		if rbs.Item(i).Selected() {
			selected++
		}
	}
	if selected != 1 {
		t.Errorf("after 3 Down presses, %d RadioButtons selected; want exactly 1", selected)
	}
	// The selected one should be Item(3).
	if !rbs.Item(3).Selected() {
		t.Errorf("Item(3) should be selected after 3 Down presses from Item(0)")
	}
}

// TestIntegrationPhase5CheckBoxValuesReflectedAfterDialogInteraction verifies
// that Values() correctly reflects checkbox state set via Dialog event dispatch.
func TestIntegrationPhase5CheckBoxValuesReflectedAfterDialogInteraction(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "Checkbox Dialog")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"~A~lpha", "~B~eta", "~G~amma"})
	d.Insert(cbs)
	d.SetFocusedChild(cbs)

	// Focus Item(0) and press Space to check it.
	cbs.SetFocusedChild(cbs.Item(0))
	spaceEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	d.HandleEvent(spaceEv)

	// Use Down arrow to move to Item(1) inside the cluster.
	// (Tab no longer cycles within clusters per spec 13.3 — use arrow keys instead.)
	downEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	d.HandleEvent(downEv)
	spaceEv2 := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	d.HandleEvent(spaceEv2)

	// Item(0) should be checked (bit 0), and Down+Space checked Item(1) (bit 1).
	values := cbs.Values()
	if values&1 == 0 {
		t.Errorf("Values() = %b; bit 0 should be set after Space on Item(0) through Dialog", values)
	}
	if values&2 == 0 {
		t.Errorf("Values() = %b; bit 1 should be set after Down+Space on Item(1)", values)
	}
}
