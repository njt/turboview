package tv

// dialog_phase9_test.go — Tests for Tasks 8, 9, and 10.
//
//   Task  8: Escape key in Dialog → EvCommand/CmCancel (after group delegation).
//   Task  9: Enter key in Dialog  → broadcast EvBroadcast/CmDefault to all children.
//   Task 10: Dialog modal termination — CmClose → CmCancel when SfModal is set.
//
// Test naming: TestDialog<Task><DescriptiveSuffix>.
//
// Helpers re-used from this package:
//   enterKey()           — integration_phase3_test.go
//   newSelectableMockView() — group_test.go
//   broadcastSpyView / newSpyView() — group_focus_broadcast_test.go
//
// Additional helpers defined here:
//   escapeKey()                — KeyEscape event
//   consumingMockView          — HandleEvent clears the event, simulating a
//                                focused child that handles the key itself

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// escapeKey returns an EvKeyboard event for the Escape key.
func escapeKey() *Event {
	return &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyEscape},
	}
}

// consumingMockView is a selectable view whose HandleEvent clears the event,
// simulating a focused child that fully handles (consumes) the key.
type consumingMockView struct {
	BaseView
}

func (c *consumingMockView) Draw(_ *DrawBuffer) {}
func (c *consumingMockView) HandleEvent(event *Event) {
	event.Clear()
}

func newConsumingMockView(bounds Rect) *consumingMockView {
	v := &consumingMockView{}
	v.SetBounds(bounds)
	v.SetState(SfVisible, true)
	v.SetOptions(OfSelectable, true)
	return v
}

// ---------------------------------------------------------------------------
// Task 8: Escape key → CmCancel
// ---------------------------------------------------------------------------

// TestDialogEscapeTransformsToCmCancel verifies that sending KeyEscape to a
// Dialog transforms the event into EvCommand/CmCancel when no child handles it.
//
// Spec (Task 8): "If event is NOT cleared and is KeyEscape: transform to
// EvCommand/CmCancel."
func TestDialogEscapeTransformsToCmCancel(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	// Insert a non-consuming child so focus exists but Escape is not handled.
	child := newSelectableMockView(NewRect(0, 0, 10, 1))
	d.Insert(child)

	ev := escapeKey()
	d.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("after Escape: event.What = %v, want EvCommand", ev.What)
	}
	if ev.Command != CmCancel {
		t.Errorf("after Escape: event.Command = %v, want CmCancel", ev.Command)
	}
	if ev.Key != nil {
		t.Error("after Escape→CmCancel transform: event.Key should be nil")
	}
}

// TestDialogEscapeConsumedByChildNotTransformed verifies that when a focused
// child clears the Escape event, Dialog does NOT transform it.
//
// Spec (Task 8): "If focused child handled Escape (cleared it), Dialog does
// NOT transform."
func TestDialogEscapeConsumedByChildNotTransformed(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	consumer := newConsumingMockView(NewRect(0, 0, 10, 1))
	d.Insert(consumer)

	ev := escapeKey()
	d.HandleEvent(ev)

	// Event was cleared by the child; Dialog must not re-transform it.
	if ev.What != EvNothing {
		t.Errorf("after child consumed Escape: event.What = %v, want EvNothing (cleared)", ev.What)
	}
	if ev.Command == CmCancel {
		t.Error("after child consumed Escape: event.Command should not be CmCancel")
	}
}

// ---------------------------------------------------------------------------
// Task 9: Enter key → broadcast CmDefault
// ---------------------------------------------------------------------------

// TestDialogEnterBroadcastsCmDefault verifies that when Enter is sent to a
// Dialog and no child handles it, Dialog broadcasts EvBroadcast/CmDefault to
// all children and clears the event.
//
// Spec (Task 9): "If event is NOT cleared and is KeyEnter: broadcast
// EvBroadcast/CmDefault to all children via group, then clear the event."
func TestDialogEnterBroadcastsCmDefault(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")

	// Two spy children to confirm broadcast reaches all children.
	spy1 := newSpyView()
	spy2 := newSpyView()
	d.Insert(spy1)
	d.Insert(spy2)
	resetBroadcasts(spy1, spy2)

	ev := enterKey()
	d.HandleEvent(ev)

	// Both spies must have received CmDefault.
	hasCmDefault := func(spy *broadcastSpyView) bool {
		for _, rec := range spy.broadcasts {
			if rec.command == CmDefault {
				return true
			}
		}
		return false
	}
	if !hasCmDefault(spy1) {
		t.Errorf("spy1 did not receive CmDefault broadcast; broadcasts: %v", spy1.broadcasts)
	}
	if !hasCmDefault(spy2) {
		t.Errorf("spy2 did not receive CmDefault broadcast; broadcasts: %v", spy2.broadcasts)
	}
}

// TestDialogEnterClearsEventAfterBroadcast verifies the Enter event is cleared
// after the CmDefault broadcast.
//
// Spec (Task 9): "broadcast EvBroadcast/CmDefault to all children via group,
// then clear the event."
func TestDialogEnterClearsEventAfterBroadcast(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	d.Insert(newSelectableMockView(NewRect(0, 0, 10, 1)))

	ev := enterKey()
	d.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("after Enter broadcast: event should be cleared, got What=%v", ev.What)
	}
}

// TestDialogEnterConsumedByChildNoBroadcast verifies that when a focused child
// clears the Enter event, Dialog does NOT broadcast CmDefault.
//
// Spec (Task 9): "If focused child handled Enter (cleared it), Dialog does NOT
// broadcast."
func TestDialogEnterConsumedByChildNoBroadcast(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	consumer := newConsumingMockView(NewRect(0, 0, 10, 1))
	spy := newSpyView()
	d.Insert(consumer) // focused (last selectable)
	d.Insert(spy)
	resetBroadcasts(spy)

	ev := enterKey()
	d.HandleEvent(ev)

	for _, rec := range spy.broadcasts {
		if rec.command == CmDefault {
			t.Errorf("CmDefault broadcast delivered even though child consumed Enter; broadcasts: %v", spy.broadcasts)
		}
	}
}

// ---------------------------------------------------------------------------
// Task 10: Modal termination — CmClose → CmCancel
// ---------------------------------------------------------------------------

// TestDialogModalCmCloseBecomesCmCancel verifies that when a Dialog has SfModal
// set and receives EvCommand/CmClose, the command is transformed to CmCancel.
//
// Spec (Task 10): "CmClose → transform to CmCancel."
func TestDialogModalCmCloseBecomesCmCancel(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	d.SetState(SfModal, true)

	ev := &Event{What: EvCommand, Command: CmClose}
	d.HandleEvent(ev)

	if ev.Command != CmCancel {
		t.Errorf("modal dialog CmClose: got Command=%v, want CmCancel", ev.Command)
	}
}

// TestDialogNonModalCmCloseUnchanged verifies that a Dialog without SfModal does
// NOT transform CmClose to CmCancel.
//
// Spec (Task 10): Dialog only applies CmClose→CmCancel when it "has SfModal".
func TestDialogNonModalCmCloseUnchanged(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	// SfModal deliberately NOT set.

	ev := &Event{What: EvCommand, Command: CmClose}
	d.HandleEvent(ev)

	if ev.Command != CmClose {
		t.Errorf("non-modal dialog CmClose: got Command=%v, want CmClose (unchanged)", ev.Command)
	}
}

// TestDialogModalCmOKPassesThrough verifies CmOK is not transformed when modal.
//
// Spec (Task 10): "CmOK, CmCancel, CmYes, CmNo: leave as-is."
func TestDialogModalCmOKPassesThrough(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	d.SetState(SfModal, true)

	ev := &Event{What: EvCommand, Command: CmOK}
	d.HandleEvent(ev)

	if ev.Command != CmOK {
		t.Errorf("modal dialog CmOK: got Command=%v, want CmOK", ev.Command)
	}
}

// TestDialogModalCmCancelPassesThrough verifies CmCancel is not transformed
// further when modal.
//
// Spec (Task 10): "CmCancel: leave as-is."
func TestDialogModalCmCancelPassesThrough(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	d.SetState(SfModal, true)

	ev := &Event{What: EvCommand, Command: CmCancel}
	d.HandleEvent(ev)

	if ev.Command != CmCancel {
		t.Errorf("modal dialog CmCancel: got Command=%v, want CmCancel", ev.Command)
	}
}

// TestDialogModalCmYesPassesThrough verifies CmYes is not transformed when
// modal.
func TestDialogModalCmYesPassesThrough(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	d.SetState(SfModal, true)

	ev := &Event{What: EvCommand, Command: CmYes}
	d.HandleEvent(ev)

	if ev.Command != CmYes {
		t.Errorf("modal dialog CmYes: got Command=%v, want CmYes", ev.Command)
	}
}

// TestDialogModalCmNoPassesThrough verifies CmNo is not transformed when modal.
func TestDialogModalCmNoPassesThrough(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	d.SetState(SfModal, true)

	ev := &Event{What: EvCommand, Command: CmNo}
	d.HandleEvent(ev)

	if ev.Command != CmNo {
		t.Errorf("modal dialog CmNo: got Command=%v, want CmNo", ev.Command)
	}
}
