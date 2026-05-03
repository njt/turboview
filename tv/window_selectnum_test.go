package tv

// window_selectnum_test.go — tests for Task 2: CmSelectWindowNum broadcast handling in Window.
// Written BEFORE any implementation exists; all tests drive the spec.
//
// Spec summary (Task 2, section 11.2):
//   - New constant CmSelectWindowNum in command.go, before CmUser.
//   - Window.HandleEvent handles EvBroadcast/CmSelectWindowNum:
//       if event.Info is an int matching w.number AND the window HasOption(OfSelectable),
//       call owner.SetFocusedChild(w) and clear the event.
//   - If the number does not match, the event passes through unconsumed.
//   - If window is not selectable (OfSelectable disabled), ignore.
//   - Wrong Info type (not int): ignore.
//   - Window returns unconditionally after EvBroadcast/CmSelectWindowNum — does NOT
//     forward to w.group.HandleEvent.

import "testing"

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// broadcastSelectNum returns an EvBroadcast event for CmSelectWindowNum with
// the given number as Info.
func broadcastSelectNum(n int) *Event {
	return &Event{What: EvBroadcast, Command: CmSelectWindowNum, Info: n}
}

// newSelectableWindow creates a Window with the given number inserted into a
// Desktop. The window has OfSelectable by default (NewWindow default).
// It returns the window and the desktop so callers can inspect FocusedChild.
func newSelectableWindow(bounds Rect, num int) (*Window, *Desktop) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w := NewWindow(bounds, "W", WithWindowNumber(num))
	d.Insert(w)
	return w, d
}

// ---------------------------------------------------------------------------
// Section 1 — Matching number, selectable window: focuses and clears
// ---------------------------------------------------------------------------

// TestWindowSelectNumMatchingNumberFocusesWindow verifies that a CmSelectWindowNum
// broadcast whose Info int matches the window's number causes the window to
// become the focused child of its owner.
// Spec: "if event.Info matches w.number and HasOption(OfSelectable), call
// owner.SetFocusedChild(w)"
func TestWindowSelectNumMatchingNumberFocusesWindow(t *testing.T) {
	w, d := newSelectableWindow(NewRect(0, 0, 30, 10), 3)

	event := broadcastSelectNum(3)
	w.HandleEvent(event)

	if d.FocusedChild() != w {
		t.Errorf("CmSelectWindowNum matching number: FocusedChild() = %v, want w (%p)", d.FocusedChild(), w)
	}
}

// TestWindowSelectNumMatchingNumberClearsEvent verifies that a CmSelectWindowNum
// broadcast with a matching number clears the event (marks it consumed).
// Spec: "consume event" when the number matches and window is selectable.
func TestWindowSelectNumMatchingNumberClearsEvent(t *testing.T) {
	w, _ := newSelectableWindow(NewRect(0, 0, 30, 10), 5)

	event := broadcastSelectNum(5)
	w.HandleEvent(event)

	if !event.IsCleared() {
		t.Errorf("CmSelectWindowNum matching number: event not cleared (IsCleared = false), want true")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Non-matching number: no focus change, event passes through
// ---------------------------------------------------------------------------

// TestWindowSelectNumNonMatchingNumberDoesNotFocus verifies that a CmSelectWindowNum
// broadcast whose Info int does not match the window's number does NOT change focus.
// Spec: "If the number doesn't match, event passes through (other windows get to check)"
func TestWindowSelectNumNonMatchingNumberDoesNotFocus(t *testing.T) {
	_, d := newSelectableWindow(NewRect(0, 0, 30, 10), 3)
	// Insert a second window so the desktop has a valid initial focus that is NOT w.
	other := NewWindow(NewRect(30, 0, 30, 10), "Other", WithWindowNumber(7))
	d.Insert(other)
	// BringToFront(other) to ensure other is the focused child before the event.
	d.BringToFront(other)

	// Deliver a broadcast for number 7 to w (number 3) directly — w should not steal focus.
	w := d.Children()[0].(*Window) // number 3
	event := broadcastSelectNum(7)
	w.HandleEvent(event)

	if d.FocusedChild() != other {
		t.Errorf("CmSelectWindowNum non-matching: FocusedChild() changed to %v, want other (%p)", d.FocusedChild(), other)
	}
}

// TestWindowSelectNumNonMatchingNumberDoesNotClearEvent verifies that a
// CmSelectWindowNum broadcast with a non-matching number does NOT clear the event.
// Spec: "If the number doesn't match, event passes through"
func TestWindowSelectNumNonMatchingNumberDoesNotClearEvent(t *testing.T) {
	w, _ := newSelectableWindow(NewRect(0, 0, 30, 10), 3)

	event := broadcastSelectNum(99)
	w.HandleEvent(event)

	if event.IsCleared() {
		t.Errorf("CmSelectWindowNum non-matching: event was cleared, want it to pass through")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Non-selectable window: no focus change, event passes through
// ---------------------------------------------------------------------------

// TestWindowSelectNumNonSelectableDoesNotFocus verifies that a CmSelectWindowNum
// broadcast with a matching number does NOT focus a window that has OfSelectable
// disabled.
// Spec: "if HasOption(OfSelectable) [is false], ignore"
func TestWindowSelectNumNonSelectableDoesNotFocus(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	// Create two windows so d has a selectable focused child distinct from w.
	anchor := NewWindow(NewRect(0, 0, 30, 10), "Anchor", WithWindowNumber(1))
	d.Insert(anchor)
	w := NewWindow(NewRect(30, 0, 30, 10), "W", WithWindowNumber(2))
	d.Insert(w)
	d.BringToFront(anchor) // anchor is now focused

	// Disable selectability on w.
	w.SetOptions(OfSelectable, false)

	event := broadcastSelectNum(2)
	w.HandleEvent(event)

	if d.FocusedChild() != anchor {
		t.Errorf("CmSelectWindowNum non-selectable: FocusedChild() changed to %v, want anchor (%p)", d.FocusedChild(), anchor)
	}
}

// TestWindowSelectNumNonSelectableDoesNotClearEvent verifies that a
// CmSelectWindowNum broadcast with a matching number on a non-selectable window
// does NOT clear the event.
// Spec: "if HasOption(OfSelectable) [is false], ignore" — event is not consumed.
func TestWindowSelectNumNonSelectableDoesNotClearEvent(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w := NewWindow(NewRect(0, 0, 30, 10), "W", WithWindowNumber(4))
	d.Insert(w)
	w.SetOptions(OfSelectable, false)

	event := broadcastSelectNum(4)
	w.HandleEvent(event)

	if event.IsCleared() {
		t.Errorf("CmSelectWindowNum non-selectable: event was cleared, want it to pass through")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Wrong Info type: no focus change, event passes through
// ---------------------------------------------------------------------------

// TestWindowSelectNumWrongInfoTypeDoesNotFocus verifies that a CmSelectWindowNum
// broadcast whose Info is not an int (e.g. a string) does NOT focus the window,
// even if the string happens to represent the right number.
// Spec: "if event.Info is an int matching w.number" — type check is required.
func TestWindowSelectNumWrongInfoTypeDoesNotFocus(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	anchor := NewWindow(NewRect(0, 0, 30, 10), "Anchor", WithWindowNumber(1))
	d.Insert(anchor)
	w := NewWindow(NewRect(30, 0, 30, 10), "W", WithWindowNumber(2))
	d.Insert(w)
	d.BringToFront(anchor) // anchor is focused

	// Info is a string, not an int.
	event := &Event{What: EvBroadcast, Command: CmSelectWindowNum, Info: "2"}
	w.HandleEvent(event)

	if d.FocusedChild() != anchor {
		t.Errorf("CmSelectWindowNum wrong Info type (string): FocusedChild() changed to %v, want anchor (%p)", d.FocusedChild(), anchor)
	}
}

// TestWindowSelectNumWrongInfoTypeDoesNotClearEvent verifies that a
// CmSelectWindowNum broadcast with a non-int Info does NOT clear the event.
// Spec: "if event.Info is an int matching w.number" — type assertion must fail
// gracefully and leave the event unconsumed.
func TestWindowSelectNumWrongInfoTypeDoesNotClearEvent(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w := NewWindow(NewRect(0, 0, 30, 10), "W", WithWindowNumber(6))
	d.Insert(w)

	event := &Event{What: EvBroadcast, Command: CmSelectWindowNum, Info: "6"}
	w.HandleEvent(event)

	if event.IsCleared() {
		t.Errorf("CmSelectWindowNum wrong Info type (string): event was cleared, want it to pass through")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Falsification: focus really moves to the matched window
// ---------------------------------------------------------------------------

// TestWindowSelectNumFocusedChildIsActuallyWindow is a falsification test that
// confirms the matched window becomes the focused child of its owner, not merely
// that the event was cleared.
// Spec: "call owner.SetFocusedChild(w)" — the owner's focused child must be w.
func TestWindowSelectNumFocusedChildIsActuallyWindow(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	first := NewWindow(NewRect(0, 0, 30, 10), "First", WithWindowNumber(1))
	second := NewWindow(NewRect(30, 0, 30, 10), "Second", WithWindowNumber(2))
	d.Insert(first)
	d.Insert(second)
	// second is focused because Insert keeps track of last selectable inserted.
	// Explicitly focus first so we can then select second via broadcast.
	d.BringToFront(first)
	if d.FocusedChild() != first {
		t.Fatalf("setup: FocusedChild() = %v, want first (%p)", d.FocusedChild(), first)
	}

	event := broadcastSelectNum(2)
	second.HandleEvent(event)

	if d.FocusedChild() != second {
		t.Errorf("after CmSelectWindowNum match: FocusedChild() = %v, want second (%p)", d.FocusedChild(), second)
	}
	if d.FocusedChild() == first {
		t.Errorf("after CmSelectWindowNum match: FocusedChild() is still first — focus did not change")
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Non-matching broadcast is NOT forwarded to the window's group
// ---------------------------------------------------------------------------

// TestWindowSelectNumDoesNotForwardToGroup verifies that Window.HandleEvent
// returns unconditionally after handling EvBroadcast/CmSelectWindowNum,
// regardless of whether the number matches.  Specifically, a non-matching
// broadcast must NOT be forwarded to w.group.HandleEvent — the child view
// inside the window must never see the event.
//
// Spec: "Window returns unconditionally after EvBroadcast/CmSelectWindowNum —
// do NOT forward to w.group.HandleEvent."
func TestWindowSelectNumDoesNotForwardToGroup(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w := NewWindow(NewRect(0, 0, 30, 10), "W", WithWindowNumber(3))
	d.Insert(w)

	// Add a selectable child inside the window so that if the window ever
	// calls w.group.HandleEvent the child's eventHandled field will be set.
	child := newSelectableMockView(NewRect(0, 0, 10, 3))
	w.Insert(child)

	// Send a broadcast for a number that does NOT match w (w is 3, we send 99).
	event := broadcastSelectNum(99)
	w.HandleEvent(event)

	// The child must not have received the event.
	if child.eventHandled != nil {
		t.Errorf("CmSelectWindowNum non-matching: child view received the event — window forwarded to group instead of returning unconditionally")
	}
}

// TestWindowSelectNumTwoWindowsBroadcastMatchesCorrectOne is a falsification
// test with two windows in a desktop. Sending CmSelectWindowNum with the number
// of window B must focus B and leave A unfocused — proving that matching is
// per-window and not global.
// Spec: "if event.Info matches w.number" — the matching is per-window.
func TestWindowSelectNumTwoWindowsBroadcastMatchesCorrectOne(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	wA := NewWindow(NewRect(0, 0, 30, 10), "A", WithWindowNumber(1))
	wB := NewWindow(NewRect(30, 0, 30, 10), "B", WithWindowNumber(2))
	d.Insert(wA)
	d.Insert(wB)
	d.BringToFront(wA) // A is focused

	// Deliver the broadcast directly to wB (simulating group dispatch to all children).
	event := broadcastSelectNum(2)
	wB.HandleEvent(event)

	// wB should now be the focused child.
	if d.FocusedChild() != wB {
		t.Errorf("two-window broadcast: FocusedChild() = %v, want wB (%p)", d.FocusedChild(), wB)
	}
	// wA must not be focused.
	if d.FocusedChild() == wA {
		t.Errorf("two-window broadcast: wA is still focused — broadcast matched the wrong window")
	}
	// The event must be cleared (wB consumed it).
	if !event.IsCleared() {
		t.Errorf("two-window broadcast: event not cleared after wB handled it")
	}
}
