package tv

// integration_phase9_batch3_test.go — Integration Checkpoint: Default Button Protocol
//
// Verifies that the Button default protocol (Tasks 13–15) works end-to-end with
// Dialog (Tasks 8–10). All tests use REAL components: Dialog, Group, Button.
// No mocks.
//
// Requirements covered:
//   Req 1: Dialog + OK (default) + Cancel: Enter activates OK via the CmDefault
//           broadcast chain (Dialog→broadcast→Button.amDefault→press).
//   Req 2: Focus moves from OK to Cancel: Cancel broadcasts CmGrabDefault →
//           OK.amDefault = false, Cancel.amDefault = true.
//   Req 3: Focus returns to OK: Cancel broadcasts CmReleaseDefault →
//           OK.amDefault = true, Cancel.amDefault = false.
//   Req 4: Alt+shortcut on a Button fires the button's command even when a
//           different widget is focused.
//   Req 5: Space on a focused Button fires its command; Space on unfocused does
//           nothing.
//   Req 6: Enter does NOT fire a non-default button — only the button with
//           amDefault responds via broadcast.
//   Req 7: Full flow — Dialog with 3 buttons (Save=bfDefault, Cancel, Help),
//           focus on Cancel, Enter → Cancel (amDefault=true) fires, not Save.
//
// Test naming: TestIntegrationPhase9Batch3<DescriptiveSuffix>

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Requirement 1: Enter activates the default button via CmDefault broadcast
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch3EnterActivatesDefaultButtonViaDialog verifies that
// pressing Enter in a Dialog with an OK (default) and Cancel button fires the OK
// button's command via the chain:
//   Dialog.HandleEvent(Enter) → broadcasts CmDefault → OK.amDefault=true → press()
//   → event transformed to EvCommand/CmOK.
//
// The Enter event is sent directly to Dialog.HandleEvent (no ExecView) so we can
// inspect the event after the call.
func TestIntegrationPhase9Batch3EnterActivatesDefaultButtonViaDialog(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Req1")

	ok := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(ok)
	dlg.Insert(cancel) // cancel gets focus last (silently, no broadcasts yet)

	// Move focus back to ok so that ok has SfSelected and cancel does not.
	// We need ok.amDefault to be true for the test's intent.
	// After Insert(cancel), cancel gained focus silently — let's put focus back on ok
	// via SetFocusedChild so broadcasts fire and ok.amDefault is re-established.
	dlg.SetFocusedChild(ok)

	// Pre-condition: ok should be amDefault after regaining focus via CmReleaseDefault.
	if !ok.IsDefault() {
		t.Fatal("pre-condition: ok.amDefault should be true when ok is focused")
	}

	ev := enterKey()
	dlg.HandleEvent(ev)

	// The CmDefault broadcast reaches ok (amDefault=true) → press() transforms the
	// broadcast event; Dialog copies What/Command back to the original event.
	if ev.What != EvCommand {
		t.Errorf("Enter in Dialog with default ok button: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("Enter in Dialog with default ok button: ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
}

// TestIntegrationPhase9Batch3EnterWithNoDefaultButtonDoesNotProduceCommand
// falsifies Req 1: when no button has amDefault=true, Enter's CmDefault broadcast
// is unhandled and the event is cleared (not transformed to EvCommand).
//
// An empty Dialog (no buttons) guarantees no child has amDefault=true, so the
// CmDefault broadcast goes completely unhandled and Enter produces no command.
func TestIntegrationPhase9Batch3EnterWithNoDefaultButtonDoesNotProduceCommand(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Req1-false")
	// No buttons inserted — no child can have amDefault=true.

	ev := enterKey()
	dlg.HandleEvent(ev)

	// No amDefault button → broadcast unhandled → event is cleared (not EvCommand).
	if ev.What == EvCommand {
		t.Errorf("Enter with no buttons (no amDefault): ev.What = EvCommand (%v), want cleared (EvNothing)", ev.What)
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: Focus moves from OK (default) to Cancel — Cancel becomes amDefault
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch3FocusMoveToNonDefaultBroadcastsGrabDefault verifies
// that when a non-bfDefault button (Cancel) gains focus via SetFocusedChild:
//   - Cancel broadcasts CmGrabDefault to siblings.
//   - OK (bfDefault) receives CmGrabDefault → ok.amDefault = false.
//   - Cancel itself gains amDefault = true (it called broadcastToOwner then set amDefault).
func TestIntegrationPhase9Batch3FocusMoveToNonDefaultBroadcastsGrabDefault(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Req2")

	ok := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(ok)
	// After Insert(ok): ok gets silent focus, ok.amDefault = true (bfDefault).

	dlg.Insert(cancel)
	// After Insert(cancel): cancel gets silent focus (setFocusSilent, no broadcasts).
	// cancel.amDefault stays false and ok.amDefault stays true at this point.

	// Now trigger real focus change with broadcasts.
	dlg.SetFocusedChild(ok) // put ok in focus first (with real broadcast)
	// ok re-gains focus → cancel loses SfSelected → cancel broadcasts CmReleaseDefault
	// → ok.amDefault = true (restored). ok should now be amDefault.

	if !ok.IsDefault() {
		t.Fatal("pre-condition: ok.amDefault should be true when ok is the focused default")
	}
	if cancel.IsDefault() {
		t.Fatal("pre-condition: cancel.amDefault should be false when ok is focused")
	}

	// Now move focus to cancel — this triggers the real SetState(SfSelected, true) on
	// cancel which broadcasts CmGrabDefault, making ok.amDefault = false.
	dlg.SetFocusedChild(cancel)

	// Spec Req 2: "Cancel broadcasts CmGrabDefault, OK's amDefault becomes false,
	// Cancel's amDefault becomes true."
	if ok.IsDefault() {
		t.Errorf("after focus moved to cancel: ok.amDefault = true, want false (CmGrabDefault should have cleared it)")
	}
	if !cancel.IsDefault() {
		t.Errorf("after focus moved to cancel: cancel.amDefault = false, want true (focused non-default gains amDefault)")
	}
}

// TestIntegrationPhase9Batch3FocusMoveToNonDefaultFalsified confirms that
// ok.amDefault remains true when focus has NOT been moved to cancel.
func TestIntegrationPhase9Batch3FocusMoveToNonDefaultFalsified(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Req2-false")

	ok := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(ok)
	dlg.Insert(cancel)

	// Put focus on ok using real focus change (broadcasts fire).
	dlg.SetFocusedChild(ok)

	// ok should be amDefault; cancel should not be amDefault.
	if !ok.IsDefault() {
		t.Errorf("ok.amDefault should be true when ok holds focus and was not displaced")
	}
	if cancel.IsDefault() {
		t.Errorf("cancel.amDefault should be false when cancel does not hold focus")
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: Focus returns to OK — Cancel releases default, OK regains it
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch3FocusReturnToDefaultRestoresAmDefault verifies that
// when focus returns to ok (bfDefault) after having been on cancel:
//   - cancel loses SfSelected → broadcasts CmReleaseDefault.
//   - ok (bfDefault) receives CmReleaseDefault → ok.amDefault = true.
//   - cancel.amDefault = false (it lost SfSelected).
func TestIntegrationPhase9Batch3FocusReturnToDefaultRestoresAmDefault(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Req3")

	ok := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(ok)
	dlg.Insert(cancel)

	// Establish state: focus on ok so broadcasts are clean.
	dlg.SetFocusedChild(ok)

	// Move focus to cancel — cancel grabs default.
	dlg.SetFocusedChild(cancel)

	// Pre-condition: cancel should be amDefault now.
	if !cancel.IsDefault() {
		t.Fatal("pre-condition: cancel.amDefault should be true after gaining focus")
	}
	if ok.IsDefault() {
		t.Fatal("pre-condition: ok.amDefault should be false after cancel grabbed default")
	}

	// Now return focus to ok — cancel releases default, ok regains it.
	dlg.SetFocusedChild(ok)

	// Spec Req 3: "Cancel broadcasts CmReleaseDefault, OK's amDefault becomes true,
	// Cancel's amDefault becomes false."
	if !ok.IsDefault() {
		t.Errorf("after focus returned to ok: ok.amDefault = false, want true (CmReleaseDefault should restore it)")
	}
	if cancel.IsDefault() {
		t.Errorf("after focus returned to ok: cancel.amDefault = true, want false (cancel released default)")
	}
}

// TestIntegrationPhase9Batch3FocusReturnToDefaultFalsified confirms the state is
// still "cancel owns default" if focus has NOT been returned to ok.
func TestIntegrationPhase9Batch3FocusReturnToDefaultFalsified(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Req3-false")

	ok := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(ok)
	dlg.Insert(cancel)
	dlg.SetFocusedChild(ok)
	dlg.SetFocusedChild(cancel) // cancel grabs default; leave it there

	// Confirm cancel still owns amDefault — ok has not regained it.
	if ok.IsDefault() {
		t.Errorf("ok.amDefault should still be false while cancel holds focus")
	}
	if !cancel.IsDefault() {
		t.Errorf("cancel.amDefault should still be true while cancel holds focus")
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Alt+shortcut fires button command even when another widget is focused
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch3AltShortcutFiresUnfocusedButton verifies that
// pressing Alt+shortcut triggers the button's command via the Group's postprocess
// phase, even when a different button is focused.
//
// All buttons have OfPostProcess set, so they receive keyboard events in Phase 3
// (postprocess) regardless of focus.
func TestIntegrationPhase9Batch3AltShortcutFiresUnfocusedButton(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Req4")

	ok := NewButton(NewRect(5, 3, 10, 1), "~OK", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(ok)
	dlg.Insert(cancel)

	// Focus is on ok (set via SetFocusedChild to trigger real broadcasts).
	dlg.SetFocusedChild(ok)

	if dlg.FocusedChild() != ok {
		t.Fatal("pre-condition: ok should be the focused button")
	}

	// Alt+C should fire cancel's command even though ok is focused.
	// Group will route the keyboard event to the focused child (ok) in Phase 2,
	// then to cancel (OfPostProcess) in Phase 3.
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'C', Modifiers: tcell.ModAlt},
	}
	dlg.HandleEvent(ev)

	// Spec Req 4: Alt+shortcut on a Button fires the button's command even when a
	// different widget is focused.
	if ev.What != EvCommand {
		t.Errorf("Alt+C with ok focused: ev.What = %v, want EvCommand (%v) from cancel's alt+shortcut", ev.What, EvCommand)
	}
	if ev.Command != CmCancel {
		t.Errorf("Alt+C with ok focused: ev.Command = %v, want CmCancel (%v)", ev.Command, CmCancel)
	}
}

// TestIntegrationPhase9Batch3AltShortcutUnfocusedFalsified confirms that a
// shortcut WITHOUT Alt does not fire a non-focused button.
func TestIntegrationPhase9Batch3AltShortcutUnfocusedFalsified(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Req4-false")

	ok := NewButton(NewRect(5, 3, 10, 1), "~OK", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(ok)
	dlg.Insert(cancel)
	dlg.SetFocusedChild(ok)

	// Plain 'C' without Alt — must not fire cancel.
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'C', Modifiers: 0},
	}
	dlg.HandleEvent(ev)

	if ev.What == EvCommand && ev.Command == CmCancel {
		t.Errorf("plain 'C' (no Alt): fired cancel command, but Alt is required for the shortcut")
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: Space fires focused button; Space does nothing to unfocused button
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch3SpaceFiresFocusedButton verifies that Space on a
// focused button (SfSelected) fires the button's command.
func TestIntegrationPhase9Batch3SpaceFiresFocusedButton(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Req5-focused")

	ok := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(ok)
	dlg.Insert(cancel)
	dlg.SetFocusedChild(cancel)

	if !cancel.HasState(SfSelected) {
		t.Fatal("pre-condition: cancel must have SfSelected (be focused)")
	}

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	dlg.HandleEvent(ev)

	// Spec Req 5: Space on a focused Button fires its command.
	if ev.What != EvCommand {
		t.Errorf("Space on focused cancel: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmCancel {
		t.Errorf("Space on focused cancel: ev.Command = %v, want CmCancel (%v)", ev.Command, CmCancel)
	}
}

// TestIntegrationPhase9Batch3SpaceOnUnfocusedButtonDoesNotFire verifies that
// Space does NOT fire a button when that button is not focused.
func TestIntegrationPhase9Batch3SpaceOnUnfocusedButtonDoesNotFire(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Req5-unfocused")

	ok := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(ok)
	dlg.Insert(cancel)
	// ok is focused — cancel is unfocused.
	dlg.SetFocusedChild(ok)

	if cancel.HasState(SfSelected) {
		t.Fatal("pre-condition: cancel must NOT have SfSelected (it is not focused)")
	}

	// Space — should fire ok (focused), not cancel (unfocused).
	// We specifically check cancel did not fire its own command.
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	dlg.HandleEvent(ev)

	// Spec Req 5: Space on an unfocused Button does nothing (for that button).
	// If space fired, it fired ok's command (CmOK), not cancel's (CmCancel).
	if ev.What == EvCommand && ev.Command == CmCancel {
		t.Errorf("Space with ok focused: fired cancel's command — unfocused button must not respond to Space")
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: Enter does NOT fire a non-default button
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch3EnterDoesNotFireNonDefaultButton verifies that a
// button without amDefault=true does NOT fire when Enter is pressed. Only the
// button with amDefault=true responds to the CmDefault broadcast.
func TestIntegrationPhase9Batch3EnterDoesNotFireNonDefaultButton(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 10), "Req6")

	ok := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(ok)
	dlg.Insert(cancel)

	// Focus ok so that ok.amDefault = true and cancel.amDefault = false.
	dlg.SetFocusedChild(ok)

	if !ok.IsDefault() {
		t.Fatal("pre-condition: ok must be amDefault=true")
	}
	if cancel.IsDefault() {
		t.Fatal("pre-condition: cancel must be amDefault=false")
	}

	ev := enterKey()
	dlg.HandleEvent(ev)

	// Spec Req 6: Enter does NOT fire a non-default button — only the default
	// button responds via the CmDefault broadcast.
	// The event should carry CmOK (from ok), not CmCancel (from cancel).
	if ev.What == EvCommand && ev.Command == CmCancel {
		t.Errorf("Enter fired cancel (non-default button); Enter must only fire the button with amDefault=true")
	}

	// Positive check: ok's command fires.
	if ev.What != EvCommand {
		t.Errorf("Enter: ev.What = %v, want EvCommand (%v) from the default button (ok)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("Enter: ev.Command = %v, want CmOK (%v) from the default button", ev.Command, CmOK)
	}
}

// TestIntegrationPhase9Batch3EnterDoesNotFireNonDefaultFocusedButton verifies
// that Enter does not directly fire a button even when that button is focused
// and non-default — it may only fire via the CmDefault broadcast if amDefault=true.
func TestIntegrationPhase9Batch3EnterDoesNotDirectlyFireButton(t *testing.T) {
	// Standalone button — no Dialog, no Group. Mimics direct Enter delivery.
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)
	b.SetState(SfSelected, true) // focused

	ev := enterKey()
	b.HandleEvent(ev)

	// Spec Req 6: Button.HandleEvent no longer has an Enter handler — Enter must
	// not directly fire the button.
	if ev.What == EvCommand {
		t.Error("Enter key directly fired button command — Enter handler must not exist on Button; use CmDefault broadcast instead")
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: Full flow — 3 buttons, focus on Cancel, Enter fires Cancel
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch3FullFlowFocusedNonDefaultButtonFiresOnEnter verifies
// the complete end-to-end scenario:
//
//	Dialog with 3 buttons: Save (bfDefault), Cancel (plain), Help (plain).
//	Initial focus is moved to Cancel via SetFocusedChild (triggers broadcasts).
//	Cancel broadcasts CmGrabDefault → Save.amDefault=false, Cancel.amDefault=true.
//	Pressing Enter → Dialog broadcasts CmDefault → Cancel (amDefault=true) fires →
//	event becomes EvCommand/CmCancel (not CmOK from Save).
//
// Spec Req 7: "focus on Cancel, press Enter → the button with amDefault (Cancel)
// fires — not Save"
func TestIntegrationPhase9Batch3FullFlowFocusedNonDefaultButtonFiresOnEnter(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 60, 10), "Req7")

	save := NewButton(NewRect(2, 3, 10, 1), "~Save", CmOK, WithDefault())
	cancel := NewButton(NewRect(14, 3, 12, 1), "~Cancel", CmCancel)
	help := NewButton(NewRect(28, 3, 10, 1), "~Help", CmUser)

	dlg.Insert(save)
	dlg.Insert(cancel)
	dlg.Insert(help)

	// Establish clean focus on save first so the broadcast chain is deterministic.
	dlg.SetFocusedChild(save)

	// Move focus to cancel via the real broadcast path.
	dlg.SetFocusedChild(cancel)

	// Spec Req 7, pre-condition: after focus moves to cancel, cancel should be
	// amDefault=true and save should be amDefault=false.
	if save.IsDefault() {
		t.Fatalf("pre-condition: save.amDefault should be false after cancel grabbed default; save.IsDefault()=%v", save.IsDefault())
	}
	if !cancel.IsDefault() {
		t.Fatalf("pre-condition: cancel.amDefault should be true after cancel grabbed default; cancel.IsDefault()=%v", cancel.IsDefault())
	}

	// Press Enter — Dialog broadcasts CmDefault.
	ev := enterKey()
	dlg.HandleEvent(ev)

	// Spec Req 7: the button with amDefault (cancel, because it grabbed default when
	// focused) fires — not save.
	if ev.What != EvCommand {
		t.Errorf("Enter with cancel as amDefault: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmCancel {
		t.Errorf("Enter with cancel as amDefault: ev.Command = %v, want CmCancel (%v) — the focused (amDefault) button must fire, not save", ev.Command, CmCancel)
	}
}

// TestIntegrationPhase9Batch3FullFlowSaveFiresWhenFocused verifies the symmetric
// case: when focus is on save (bfDefault), Enter fires save's command (CmOK).
func TestIntegrationPhase9Batch3FullFlowSaveFiresWhenFocused(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 60, 10), "Req7-sym")

	save := NewButton(NewRect(2, 3, 10, 1), "~Save", CmOK, WithDefault())
	cancel := NewButton(NewRect(14, 3, 12, 1), "~Cancel", CmCancel)
	help := NewButton(NewRect(28, 3, 10, 1), "~Help", CmUser)

	dlg.Insert(save)
	dlg.Insert(cancel)
	dlg.Insert(help)

	// Focus on save — save should be amDefault.
	dlg.SetFocusedChild(save)

	if !save.IsDefault() {
		t.Fatal("pre-condition: save.amDefault should be true when save is focused")
	}

	ev := enterKey()
	dlg.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("Enter with save as amDefault: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("Enter with save as amDefault: ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
}

// TestIntegrationPhase9Batch3FullFlowTabCycleChangesDefaultOwner verifies that
// cycling through all three buttons via Tab changes which button owns amDefault
// at each step, and Enter always fires the current amDefault holder.
func TestIntegrationPhase9Batch3FullFlowTabCycleChangesDefaultOwner(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 60, 10), "Req7-tab")

	save := NewButton(NewRect(2, 3, 10, 1), "~Save", CmOK, WithDefault())
	cancel := NewButton(NewRect(14, 3, 12, 1), "~Cancel", CmCancel)
	help := NewButton(NewRect(28, 3, 10, 1), "~Help", CmUser)

	dlg.Insert(save)
	dlg.Insert(cancel)
	dlg.Insert(help)

	// Start with save focused.
	dlg.SetFocusedChild(save)

	// Tab → focus moves to cancel (next selectable after save).
	dlg.HandleEvent(tabKey())
	if dlg.FocusedChild() != cancel {
		t.Fatalf("after Tab from save: FocusedChild = %v, want cancel", dlg.FocusedChild())
	}

	// cancel should own amDefault now; save should not.
	if save.IsDefault() {
		t.Errorf("after Tab to cancel: save.amDefault = true, want false")
	}
	if !cancel.IsDefault() {
		t.Errorf("after Tab to cancel: cancel.amDefault = false, want true")
	}

	// Tab again → focus moves to help.
	dlg.HandleEvent(tabKey())
	if dlg.FocusedChild() != help {
		t.Fatalf("after second Tab: FocusedChild = %v, want help", dlg.FocusedChild())
	}

	// help should own amDefault now; cancel should not.
	if cancel.IsDefault() {
		t.Errorf("after Tab to help: cancel.amDefault = true, want false")
	}
	if !help.IsDefault() {
		t.Errorf("after Tab to help: help.amDefault = false, want true")
	}

	// Tab again → focus wraps back to save.
	dlg.HandleEvent(tabKey())
	if dlg.FocusedChild() != save {
		t.Fatalf("after third Tab (wrap): FocusedChild = %v, want save", dlg.FocusedChild())
	}

	// save (bfDefault) should have amDefault restored by CmReleaseDefault from help.
	if !save.IsDefault() {
		t.Errorf("after Tab wrap to save: save.amDefault = false, want true (bfDefault restored on CmReleaseDefault)")
	}
}

// ---------------------------------------------------------------------------
// ExecView end-to-end — Enter fires default button returns CmOK via modal loop
// ---------------------------------------------------------------------------

// TestIntegrationPhase9Batch3ExecViewEnterFiresDefaultAndExits verifies the
// complete ExecView end-to-end flow: Dialog with default OK button running in
// the real modal loop, Enter key injected via screen → ExecView returns CmOK.
func TestIntegrationPhase9Batch3ExecViewEnterFiresDefaultAndExits(t *testing.T) {
	_, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(10, 5, 40, 10), "E2E Default")
	ok := NewButton(NewRect(5, 3, 10, 1), "OK", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(ok)
	dlg.Insert(cancel)
	// Focus on ok (the default button) before ExecView.
	dlg.SetFocusedChild(ok)

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Inject Enter — Dialog broadcasts CmDefault → ok (amDefault=true) fires → CmOK.
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	select {
	case cmd := <-result:
		if cmd != CmOK {
			t.Errorf("ExecView with default ok focused: returned %v after Enter, want CmOK", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmOK within 2 s after Enter with default ok")
	}
}

// TestIntegrationPhase9Batch3ExecViewEnterFiresFocusedNonDefaultAndExits verifies
// the Req 7 scenario end-to-end via ExecView: Dialog with Save (bfDefault) and
// Cancel; focus moved to Cancel before Enter is pressed. ExecView must return
// CmCancel (not CmOK), because Cancel owns amDefault when focused.
func TestIntegrationPhase9Batch3ExecViewEnterFiresFocusedNonDefaultAndExits(t *testing.T) {
	_, desktop, screen := appWithDesktopAndScreen(t)
	defer screen.Fini()

	dlg := NewDialog(NewRect(10, 5, 40, 10), "E2E NonDefault")
	save := NewButton(NewRect(5, 3, 10, 1), "~Save", CmOK, WithDefault())
	cancel := NewButton(NewRect(18, 3, 12, 1), "~Cancel", CmCancel)

	dlg.Insert(save)
	dlg.Insert(cancel)
	// Move focus to cancel — cancel grabs amDefault, save loses it.
	dlg.SetFocusedChild(save) // clean start
	dlg.SetFocusedChild(cancel)

	if !cancel.IsDefault() {
		t.Fatal("pre-condition: cancel.amDefault must be true before starting ExecView")
	}

	result := make(chan CommandCode, 1)
	go func() {
		result <- desktop.ExecView(dlg)
	}()

	time.Sleep(50 * time.Millisecond)

	// Enter → CmDefault broadcast → cancel (amDefault=true) fires → CmCancel.
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	select {
	case cmd := <-result:
		if cmd != CmCancel {
			t.Errorf("ExecView with cancel as amDefault: returned %v after Enter, want CmCancel", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmCancel within 2 s after Enter with cancel as amDefault")
	}
}
