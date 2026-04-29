package tv

// window_modal_test.go — tests for Task 11: Window CmClose on Modal.
// Written BEFORE any implementation; tests drive the spec.
//
// Spec:
//   When Window.HandleEvent receives EvCommand/CmClose and the window has
//   SfModal state, the event command is transformed to CmCancel and the
//   method returns without delegating to group.
//   When the window does NOT have SfModal, CmClose passes through unchanged.

import (
	"testing"
)

// modalCmdEvent creates an EvCommand event with the given command code.
// Named to avoid collision with cmdEvent defined in desktop_window_mgmt_test.go.
func modalCmdEvent(cmd CommandCode) *Event {
	return &Event{What: EvCommand, Command: cmd}
}

// newModalTestWindow returns a 20×10 Window at origin with SfModal set.
func newModalTestWindow() *Window {
	w := NewWindow(NewRect(0, 0, 20, 10), "Modal")
	w.SetState(SfModal, true)
	return w
}

// ---------------------------------------------------------------------------
// Test 1: Modal window converts CmClose to CmCancel
// ---------------------------------------------------------------------------

// TestModalWindowCmCloseTransformedToCmCancel verifies that a modal window
// converts a CmClose command to CmCancel and does not pass it to the group.
//
// Spec: "When Window.HandleEvent receives EvCommand with CmClose and the window
// has SfModal state: Transform the command to CmCancel."
func TestModalWindowCmCloseTransformedToCmCancel(t *testing.T) {
	w := newModalTestWindow()
	event := modalCmdEvent(CmClose)

	w.HandleEvent(event)

	if event.What != EvCommand {
		t.Fatalf("expected event.What=EvCommand, got %v", event.What)
	}
	if event.Command != CmCancel {
		t.Errorf("expected CmCancel after modal CmClose, got %v", event.Command)
	}
}

// ---------------------------------------------------------------------------
// Test 2: Non-modal window leaves CmClose unchanged
// ---------------------------------------------------------------------------

// TestNonModalWindowCmClosePassesThrough verifies that a non-modal window does
// not transform CmClose — it delegates to group unchanged.
//
// Spec: "When Window receives CmClose without SfModal, behavior is unchanged."
func TestNonModalWindowCmClosePassesThrough(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Plain")
	// SfModal is NOT set
	event := modalCmdEvent(CmClose)

	w.HandleEvent(event)

	// After passing to group (which has no handler for it), the command should
	// still be CmClose — it must not have been silently transformed.
	if event.Command != CmClose {
		t.Errorf("expected CmClose to remain CmClose on non-modal window, got %v", event.Command)
	}
}

// ---------------------------------------------------------------------------
// Test 3: Modal window: CmCancel is not further transformed (passes through)
// ---------------------------------------------------------------------------

// TestModalWindowCmCancelPassesThrough verifies that a CmCancel event on a
// modal window is not re-transformed; it delegates to group as-is.
//
// Spec: "Modal window: CmCancel is not further transformed (passes through)."
func TestModalWindowCmCancelPassesThrough(t *testing.T) {
	w := newModalTestWindow()
	event := modalCmdEvent(CmCancel)

	w.HandleEvent(event)

	// The event should still carry CmCancel, not be changed to something else.
	if event.Command != CmCancel {
		t.Errorf("expected CmCancel to remain CmCancel on modal window, got %v", event.Command)
	}
}

// ---------------------------------------------------------------------------
// Test 4: Modal window: non-CmClose commands pass through to group
// ---------------------------------------------------------------------------

// TestModalWindowOtherCommandsPassThrough verifies that commands other than
// CmClose on a modal window are not intercepted; they reach the group.
//
// Spec: "Modal window: non-CmClose commands pass through to group."
func TestModalWindowOtherCommandsPassThrough(t *testing.T) {
	w := newModalTestWindow()
	event := modalCmdEvent(CmOK)

	w.HandleEvent(event)

	// CmOK should be unchanged — the window must not have transformed it.
	if event.Command != CmOK {
		t.Errorf("expected CmOK to remain unchanged on modal window, got %v", event.Command)
	}
}
