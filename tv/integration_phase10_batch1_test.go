package tv

// integration_phase10_batch1_test.go — Integration tests for Phase 10 Tasks 1–4:
// Cluster Behavior Checkpoint.
//
// Verifies that cluster widgets (CheckBoxes, RadioButtons) work correctly inside
// real Dialogs with Buttons.
//
//   Task 1: Focused checkbox in a Dialog shows ► prefix when drawn (DrawBuffer)
//   Task 2: Focused radio button in a Dialog shows ► prefix when drawn
//   Task 3: Down arrow in CheckBoxes moves focus without toggling checked state
//   Task 4: Left/Right arrows in RadioButtons change selection same as Up/Down
//   Task 5: Enter in a Dialog with CheckBoxes + default Button fires the Button,
//            not any checkbox (states unchanged, event becomes EvCommand)
//   Task 6: Enter in a Dialog with RadioButtons + default Button fires the Button,
//            not changing radio selection
//   Task 7: Tab from a cluster in a Dialog moves focus to the next sibling (Button),
//            not within the cluster's internal items
//
// Test naming: TestIntegrationPhase10Batch1<DescriptiveSuffix>
//
// Helpers used from sibling test files:
//   enterKey() — integration_phase3_test.go
//   tabKey()   — integration_phase3_test.go

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Test 1: Focused checkbox in a Dialog shows ► prefix in DrawBuffer
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch1FocusedCheckBoxShowsCursorInDialog verifies that
// when a CheckBoxes cluster is inside a Dialog, the internally-focused CheckBox
// renders the ► prefix in the draw output.
//
// Dialog at (0,0,30,8) with 1-pixel frame → client area starts at (1,1).
// CheckBoxes at client-relative (0,0,20,3) → dialog buffer rows 1–3, cols 1–20.
// Item(1) is at y=1 within the cluster → dialog buffer row 2, col 1.
// After dlg.Draw, cell at (1,2) must be '►'.
func TestIntegrationPhase10Batch1FocusedCheckBoxShowsCursorInDialog(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 30, 8), "Checkboxes")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})
	cbs.scheme = theme.BorlandBlue
	dlg.Insert(cbs)

	// Move internal focus to Item(1).
	cbs.SetFocusedChild(cbs.Item(1))

	buf := NewDrawBuffer(30, 8)
	dlg.Draw(buf)

	// Dialog frame is at row 0 and col 0; client area begins at (1,1).
	// Cluster is at client (0,0), so Item(1) at cluster y=1 → dialog buffer y=2, x=1.
	cell := buf.GetCell(1, 2)
	if cell.Rune != '►' {
		t.Errorf("focused CheckBox (item 1) drawn inside Dialog: cell(1,2) = %q, want '►'", cell.Rune)
	}
}

// TestIntegrationPhase10Batch1UnfocusedCheckBoxDoesNotShowCursorInDialog confirms
// that unfocused checkboxes do not render the ► prefix inside a Dialog.
func TestIntegrationPhase10Batch1UnfocusedCheckBoxDoesNotShowCursorInDialog(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 30, 8), "Checkboxes")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})
	cbs.scheme = theme.BorlandBlue
	dlg.Insert(cbs)

	// Item(1) is focused; items 0 and 2 must not show ►.
	cbs.SetFocusedChild(cbs.Item(1))

	buf := NewDrawBuffer(30, 8)
	dlg.Draw(buf)

	// Item(0) → dialog buffer y=1, x=1.
	if buf.GetCell(1, 1).Rune == '►' {
		t.Errorf("unfocused CheckBox (item 0) at dialog cell(1,1) shows '►'; only focused item may show cursor")
	}
	// Item(2) → dialog buffer y=3, x=1.
	if buf.GetCell(1, 3).Rune == '►' {
		t.Errorf("unfocused CheckBox (item 2) at dialog cell(1,3) shows '►'; only focused item may show cursor")
	}
}

// ---------------------------------------------------------------------------
// Test 2: Focused radio button in a Dialog shows ► prefix in DrawBuffer
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch1FocusedRadioButtonShowsCursorInDialog verifies that
// when a RadioButtons cluster is inside a Dialog, the internally-focused RadioButton
// renders the ► prefix in the draw output.
//
// Dialog at (0,0,30,8), RadioButtons at client (0,0,20,3).
// Item(2) at cluster y=2 → dialog buffer y=3, x=1.
func TestIntegrationPhase10Batch1FocusedRadioButtonShowsCursorInDialog(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 30, 8), "Radiobuttons")

	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"One", "Two", "Three"})
	rbs.scheme = theme.BorlandBlue
	dlg.Insert(rbs)

	// Move internal focus to Item(2).
	rbs.SetFocusedChild(rbs.Item(2))

	buf := NewDrawBuffer(30, 8)
	dlg.Draw(buf)

	// Item(2) at cluster y=2 → dialog buffer y=3, x=1.
	cell := buf.GetCell(1, 3)
	if cell.Rune != '►' {
		t.Errorf("focused RadioButton (item 2) drawn inside Dialog: cell(1,3) = %q, want '►'", cell.Rune)
	}
}

// TestIntegrationPhase10Batch1UnfocusedRadioButtonDoesNotShowCursorInDialog confirms
// that unfocused radio buttons do not render the ► prefix inside a Dialog.
func TestIntegrationPhase10Batch1UnfocusedRadioButtonDoesNotShowCursorInDialog(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 30, 8), "Radiobuttons")

	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"One", "Two", "Three"})
	rbs.scheme = theme.BorlandBlue
	dlg.Insert(rbs)

	// Focus item 2; items 0 and 1 must not show ►.
	rbs.SetFocusedChild(rbs.Item(2))

	buf := NewDrawBuffer(30, 8)
	dlg.Draw(buf)

	// Item(0) → dialog buffer y=1, x=1.
	if buf.GetCell(1, 1).Rune == '►' {
		t.Errorf("unfocused RadioButton (item 0) at dialog cell(1,1) shows '►'; only focused item may show cursor")
	}
	// Item(1) → dialog buffer y=2, x=1.
	if buf.GetCell(1, 2).Rune == '►' {
		t.Errorf("unfocused RadioButton (item 1) at dialog cell(1,2) shows '►'; only focused item may show cursor")
	}
}

// ---------------------------------------------------------------------------
// Test 3: Down arrow in CheckBoxes moves focus without toggling checked state
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch1DownArrowInDialogCheckBoxesMovesFocusOnly verifies
// that pressing Down in a CheckBoxes cluster inside a Dialog moves internal focus
// to the next item but does NOT toggle any checkbox.
//
// This is the integration version of the unit test in checkbox_arrows_test.go,
// exercised through the full Dialog → group → cluster dispatch chain.
func TestIntegrationPhase10Batch1DownArrowInDialogCheckBoxesMovesFocusOnly(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 30, 10), "CheckBoxes Down")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	dlg.Insert(cbs)
	dlg.SetFocusedChild(cbs)

	// Focus Item(0) within the cluster.
	cbs.SetFocusedChild(cbs.Item(0))

	// Record initial checked states (all false).
	for i := 0; i < 3; i++ {
		if cbs.Item(i).Checked() {
			t.Fatalf("pre-condition: Item(%d) should not be checked initially", i)
		}
	}

	// Deliver Down arrow through the Dialog's event chain.
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	dlg.HandleEvent(ev)

	// Internal focus must have moved to Item(1).
	if cbs.FocusedChild() != cbs.Item(1) {
		t.Errorf("Down arrow: cbs.FocusedChild() = %v, want Item(1)", cbs.FocusedChild())
	}

	// No item may have been toggled.
	for i := 0; i < 3; i++ {
		if cbs.Item(i).Checked() {
			t.Errorf("Down arrow toggled Item(%d); arrow key must not toggle any checkbox", i)
		}
	}
}

// TestIntegrationPhase10Batch1DownArrowCheckedCheckBoxNotToggled is the falsified
// case: a checked item stays checked after Down arrow passes through it.
func TestIntegrationPhase10Batch1DownArrowCheckedCheckBoxNotToggled(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 30, 10), "CheckBoxes Down False")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	dlg.Insert(cbs)
	dlg.SetFocusedChild(cbs)

	// Pre-check Item(1) so we can verify it stays checked after a Down from Item(0).
	cbs.Item(1).SetChecked(true)
	cbs.SetFocusedChild(cbs.Item(0))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	dlg.HandleEvent(ev)

	// Item(1) must still be checked (Down arrow must not have toggled it).
	if !cbs.Item(1).Checked() {
		t.Errorf("Down arrow toggled (unchecked) Item(1) which was checked; arrow key must not change checked state")
	}
}

// ---------------------------------------------------------------------------
// Test 4: Left/Right arrows in RadioButtons change selection same as Up/Down
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch1RightArrowInDialogRadioButtonsChangesSelection verifies
// that pressing Right in a RadioButtons cluster inside a Dialog changes the selected
// radio button to the next one (same as Down), exercised through the Dialog dispatch chain.
func TestIntegrationPhase10Batch1RightArrowInDialogRadioButtonsChangesSelection(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 30, 10), "RadioButtons Right")

	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"X", "Y", "Z"})
	dlg.Insert(rbs)
	dlg.SetFocusedChild(rbs)

	// Item(0) is selected by default.
	if rbs.Value() != 0 {
		t.Fatalf("pre-condition: RadioButtons.Value() = %d, want 0 (first item selected)", rbs.Value())
	}

	// Deliver Right arrow through the Dialog's event chain.
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	dlg.HandleEvent(ev)

	if rbs.Value() != 1 {
		t.Errorf("Right arrow: RadioButtons.Value() = %d, want 1 (second item)", rbs.Value())
	}
	if !rbs.Item(1).Selected() {
		t.Errorf("Right arrow: Item(1).Selected() = false, want true")
	}
	if rbs.Item(0).Selected() {
		t.Errorf("Right arrow: Item(0).Selected() = true, want false (deselected)")
	}
}

// TestIntegrationPhase10Batch1LeftArrowInDialogRadioButtonsChangesSelection verifies
// that pressing Left in a RadioButtons cluster inside a Dialog changes selection to
// the previous item (same as Up), exercised through the Dialog dispatch chain.
func TestIntegrationPhase10Batch1LeftArrowInDialogRadioButtonsChangesSelection(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 30, 10), "RadioButtons Left")

	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"X", "Y", "Z"})
	dlg.Insert(rbs)
	dlg.SetFocusedChild(rbs)

	// Start with Item(2) selected.
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))

	// Deliver Left arrow through the Dialog.
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	dlg.HandleEvent(ev)

	if rbs.Value() != 1 {
		t.Errorf("Left arrow: RadioButtons.Value() = %d, want 1 (second item)", rbs.Value())
	}
	if !rbs.Item(1).Selected() {
		t.Errorf("Left arrow: Item(1).Selected() = false, want true")
	}
}

// ---------------------------------------------------------------------------
// Test 5: Enter in a Dialog with CheckBoxes + default Button fires the Button
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch1EnterInDialogWithCheckBoxesFiresDefaultButton verifies
// that pressing Enter in a Dialog containing CheckBoxes and a default Button fires
// the Button's command (via CmDefault broadcast) and does NOT toggle any checkbox.
//
// Chain: Dialog.HandleEvent(Enter) → group delegates to focused cluster → cluster
// does not handle Enter → returns to Dialog post-delegation → Dialog broadcasts
// CmDefault → default Button responds via press() → event becomes EvCommand/CmOK.
func TestIntegrationPhase10Batch1EnterInDialogWithCheckBoxesFiresDefaultButton(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Enter CheckBoxes")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})
	btn := NewButton(NewRect(0, 4, 10, 1), "OK", CmOK, WithDefault())

	dlg.Insert(cbs)
	dlg.Insert(btn)

	// Focus the cluster so the Button is not focused (cluster is focused child).
	dlg.SetFocusedChild(cbs)

	// Pre-condition: no checkbox should be checked initially.
	for i := 0; i < 3; i++ {
		if cbs.Item(i).Checked() {
			t.Fatalf("pre-condition: Item(%d) should not be checked initially", i)
		}
	}

	ev := enterKey()
	dlg.HandleEvent(ev)

	// Enter must have fired the default Button, not toggled any checkbox.
	if ev.What != EvCommand {
		t.Errorf("Enter with default Button present: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("Enter with default Button present: ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}

	// Checkbox states must be unchanged.
	for i := 0; i < 3; i++ {
		if cbs.Item(i).Checked() {
			t.Errorf("Enter key toggled Item(%d); Enter must not toggle any checkbox", i)
		}
	}
}

// TestIntegrationPhase10Batch1EnterDoesNotToggleCheckedCheckBox is the falsified
// case: even when a checkbox is already checked, Enter must not untoggle it.
func TestIntegrationPhase10Batch1EnterDoesNotToggleCheckedCheckBox(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Enter No Toggle")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})
	btn := NewButton(NewRect(0, 4, 10, 1), "OK", CmOK, WithDefault())

	dlg.Insert(cbs)
	dlg.Insert(btn)
	dlg.SetFocusedChild(cbs)

	// Pre-check Item(0).
	cbs.Item(0).SetChecked(true)

	ev := enterKey()
	dlg.HandleEvent(ev)

	// Item(0) must remain checked after Enter.
	if !cbs.Item(0).Checked() {
		t.Errorf("Enter key untoggled Item(0) which was checked; Enter must not change checkbox state")
	}
}

// ---------------------------------------------------------------------------
// Test 6: Enter in a Dialog with RadioButtons + default Button fires the Button
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch1EnterInDialogWithRadioButtonsFiresDefaultButton verifies
// that pressing Enter in a Dialog containing RadioButtons and a default Button fires
// the Button's command and does NOT change radio selection.
//
// This confirms that RadioButton's removal of Enter handling (Task 4) means Enter
// propagates out of the cluster to the Dialog's CmDefault broadcast path.
func TestIntegrationPhase10Batch1EnterInDialogWithRadioButtonsFiresDefaultButton(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Enter RadioButtons")

	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"One", "Two", "Three"})
	btn := NewButton(NewRect(0, 4, 10, 1), "OK", CmOK, WithDefault())

	dlg.Insert(rbs)
	dlg.Insert(btn)

	// Focus the cluster; Item(0) is selected by default.
	dlg.SetFocusedChild(rbs)

	initialValue := rbs.Value()
	if initialValue != 0 {
		t.Fatalf("pre-condition: RadioButtons.Value() = %d, want 0", initialValue)
	}

	ev := enterKey()
	dlg.HandleEvent(ev)

	// Enter must have fired the default Button.
	if ev.What != EvCommand {
		t.Errorf("Enter with default Button present: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("Enter with default Button present: ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}

	// Radio selection must not have changed.
	if rbs.Value() != initialValue {
		t.Errorf("Enter changed radio selection from %d to %d; Enter must not change selection", initialValue, rbs.Value())
	}
}

// TestIntegrationPhase10Batch1EnterDoesNotChangeRadioSelectionFalsified confirms
// that the radio selection was indeed stable — the second item was NOT accidentally
// selected (which would only happen if Enter was routed as a Down arrow or similar).
func TestIntegrationPhase10Batch1EnterDoesNotChangeRadioSelectionFalsified(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Enter Radio False")

	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"One", "Two", "Three"})
	btn := NewButton(NewRect(0, 4, 10, 1), "OK", CmOK, WithDefault())

	dlg.Insert(rbs)
	dlg.Insert(btn)
	dlg.SetFocusedChild(rbs)

	// Start with Item(1) selected.
	rbs.SetValue(1)
	rbs.SetFocusedChild(rbs.Item(1))

	ev := enterKey()
	dlg.HandleEvent(ev)

	// Item(1) must still be selected.
	if !rbs.Item(1).Selected() {
		t.Errorf("Enter changed selection away from Item(1); Enter must preserve radio selection")
	}
	// Item(0) and Item(2) must not have been selected.
	if rbs.Item(0).Selected() {
		t.Errorf("Enter selected Item(0); Enter must not change radio selection")
	}
	if rbs.Item(2).Selected() {
		t.Errorf("Enter selected Item(2); Enter must not change radio selection")
	}
}

// ---------------------------------------------------------------------------
// Test 7: Tab from a cluster navigates to the next sibling in the Dialog
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch1TabFromCheckBoxesClusterMovesToButton verifies that
// pressing Tab while a CheckBoxes cluster is the focused child of a Dialog moves
// focus to the next selectable sibling (a Button), NOT within the cluster's items.
//
// The Dialog's group handles Tab at its own level (before routing to the focused
// child), so Tab always moves focus between Dialog-level children.
func TestIntegrationPhase10Batch1TabFromCheckBoxesClusterMovesToButton(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Tab CheckBoxes")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	btn := NewButton(NewRect(0, 4, 10, 1), "OK", CmOK)

	dlg.Insert(cbs)
	dlg.Insert(btn)

	// Put focus on the cluster at the Dialog level.
	dlg.SetFocusedChild(cbs)

	if dlg.FocusedChild() != cbs {
		t.Fatalf("pre-condition: Dialog FocusedChild() = %v, want cbs", dlg.FocusedChild())
	}

	// Deliver Tab through the Dialog.
	dlg.HandleEvent(tabKey())

	// Focus must have moved from the cluster to the Button at the dialog level.
	if dlg.FocusedChild() != btn {
		t.Errorf("Tab from CheckBoxes cluster: Dialog.FocusedChild() = %v, want btn", dlg.FocusedChild())
	}
}

// TestIntegrationPhase10Batch1TabFromRadioButtonsClusterMovesToButton verifies the
// same Tab behavior with a RadioButtons cluster: Tab moves focus to the next sibling
// Button at the Dialog level, not within the cluster's internal items.
func TestIntegrationPhase10Batch1TabFromRadioButtonsClusterMovesToButton(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Tab RadioButtons")

	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"X", "Y", "Z"})
	btn := NewButton(NewRect(0, 4, 10, 1), "OK", CmOK)

	dlg.Insert(rbs)
	dlg.Insert(btn)

	// Put focus on the cluster at the Dialog level.
	dlg.SetFocusedChild(rbs)

	if dlg.FocusedChild() != rbs {
		t.Fatalf("pre-condition: Dialog FocusedChild() = %v, want rbs", dlg.FocusedChild())
	}

	// Deliver Tab through the Dialog.
	dlg.HandleEvent(tabKey())

	// Focus must have moved from the cluster to the Button.
	if dlg.FocusedChild() != btn {
		t.Errorf("Tab from RadioButtons cluster: Dialog.FocusedChild() = %v, want btn", dlg.FocusedChild())
	}
}

// TestIntegrationPhase10Batch1TabFromButtonMovesBackToCluster verifies that Tab
// cycles: after Tab from cluster to Button, another Tab wraps back to the cluster.
func TestIntegrationPhase10Batch1TabFromButtonMovesBackToCluster(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Tab Cycle")

	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	btn := NewButton(NewRect(0, 4, 10, 1), "OK", CmOK)

	dlg.Insert(cbs)
	dlg.Insert(btn)

	// Start with focus on the cluster.
	dlg.SetFocusedChild(cbs)

	// Tab → Button.
	dlg.HandleEvent(tabKey())
	if dlg.FocusedChild() != btn {
		t.Fatalf("first Tab: FocusedChild = %v, want btn", dlg.FocusedChild())
	}

	// Tab again → should wrap back to cluster (only two selectable children).
	dlg.HandleEvent(tabKey())
	if dlg.FocusedChild() != cbs {
		t.Errorf("second Tab (wrap): FocusedChild = %v, want cbs", dlg.FocusedChild())
	}
}
