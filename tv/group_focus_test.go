package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// stubView is a test double for View that records the last event it received.
type stubView struct {
	BaseView
	lastEvent *Event
}

func (s *stubView) HandleEvent(event *Event) {
	s.lastEvent = event
}

// newSelectableView creates a visible, selectable stubView.
func newSelectableView() *stubView {
	v := &stubView{}
	v.SetBounds(NewRect(0, 0, 10, 1))
	v.SetState(SfVisible, true)
	v.SetOptions(OfSelectable, true)
	return v
}

// newNonSelectableView creates a visible, non-selectable stubView.
func newNonSelectableView() *stubView {
	v := &stubView{}
	v.SetBounds(NewRect(0, 0, 10, 1))
	v.SetState(SfVisible, true)
	return v
}

// tabEvent returns a keyboard event for the Tab key with no modifiers.
func tabEvent() *Event {
	return &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyTab},
	}
}

// shiftTabEvent returns a keyboard event for Shift+Tab (Backtab).
func shiftTabEvent() *Event {
	return &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyBacktab},
	}
}

// ── Tab is no longer intercepted by Group ──────────────────────────────────────
// Tab/Shift+Tab handling was moved to Window (spec 13.3). Group no longer
// intercepts Tab before three-phase dispatch.

// TestTabSkipFalsifiedNonSelectableIsNotFocused catches an implementation that
// incorrectly focuses non-selectable children.
// Spec: "focusNext() skips non-selectable children".
func TestTabSkipFalsifiedNonSelectableIsNotFocused(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	nonSel := newNonSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(nonSel)
	g.Insert(second)
	g.SetFocusedChild(first)

	g.HandleEvent(tabEvent())

	if g.FocusedChild() == nonSel {
		t.Errorf("after Tab, FocusedChild() is the non-selectable child — skipping failed")
	}
}

// TestTabWithOnlyOneSelectableChildKeepsFocus verifies that Tab with a single
// selectable child leaves that child focused.
// Spec: "If only one selectable child exists, it stays focused."
func TestTabWithOnlyOneSelectableChildKeepsFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	only := newSelectableView()
	g.Insert(only)

	g.HandleEvent(tabEvent())

	if g.FocusedChild() != only {
		t.Errorf("Tab with single selectable: FocusedChild() = %v, want the only selectable child", g.FocusedChild())
	}
}

// TestTabWithNoSelectableChildrenDoesNothing verifies that Tab with no selectable
// children does not crash or change anything.
// Spec: "If no selectable children exist, nothing happens."
func TestTabWithNoSelectableChildrenDoesNothing(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	nonSel := newNonSelectableView()
	g.Insert(nonSel)

	// Must not panic, and focused must remain nil.
	g.HandleEvent(tabEvent())

	if g.FocusedChild() != nil {
		t.Errorf("Tab with no selectable children: FocusedChild() = %v, want nil", g.FocusedChild())
	}
}

// ── Shift+Tab is no longer intercepted by Group ─────────────────────────────

// TestShiftTabSkipFalsifiedNonSelectableIsNotFocused catches an implementation
// that incorrectly stops at a non-selectable child during backward traversal.
// Spec: "focusPrev() ... skips non-selectable children".
func TestShiftTabSkipFalsifiedNonSelectableIsNotFocused(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	nonSel := newNonSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(nonSel)
	g.Insert(second)
	// second is auto-focused.

	g.HandleEvent(shiftTabEvent())

	if g.FocusedChild() == nonSel {
		t.Errorf("after Shift+Tab, FocusedChild() is the non-selectable child — skipping failed")
	}
}

// TestShiftTabWithOnlyOneSelectableChildKeepsFocus verifies that Shift+Tab with
// a single selectable child leaves that child focused.
// Spec: "If only one selectable child exists, it stays focused."
func TestShiftTabWithOnlyOneSelectableChildKeepsFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	only := newSelectableView()
	g.Insert(only)

	g.HandleEvent(shiftTabEvent())

	if g.FocusedChild() != only {
		t.Errorf("Shift+Tab with single selectable: FocusedChild() = %v, want the only selectable child", g.FocusedChild())
	}
}

// TestShiftTabWithNoSelectableChildrenDoesNothing verifies Shift+Tab with no
// selectable children does not crash.
// Spec: "If no selectable children exist, nothing happens."
func TestShiftTabWithNoSelectableChildrenDoesNothing(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	nonSel := newNonSelectableView()
	g.Insert(nonSel)

	// Must not panic, and focused must remain nil.
	g.HandleEvent(shiftTabEvent())

	if g.FocusedChild() != nil {
		t.Errorf("Shift+Tab with no selectable children: FocusedChild() = %v, want nil", g.FocusedChild())
	}
}

// ── Non-Tab keyboard events still use three-phase dispatch ───────────────────

// TestNonTabKeyboardEventStillDispatchedToFocusedChild verifies that non-Tab
// keyboard events continue to use the existing three-phase dispatch.
// Spec: "Non-Tab keyboard events continue to use the existing three-phase dispatch
// unchanged."
func TestNonTabKeyboardEventStillDispatchedToFocusedChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	g.Insert(first)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyEnter},
	}
	g.HandleEvent(ev)

	if first.lastEvent != ev {
		t.Errorf("non-Tab keyboard event not dispatched to focused child — three-phase dispatch broken")
	}
}

// TestNonTabKeyboardEventIsNotCleared verifies that a non-Tab keyboard event is
// not incorrectly cleared by the Tab interception logic.
// Spec: "Non-Tab keyboard events continue to use the existing three-phase dispatch
// unchanged."
func TestNonTabKeyboardEventIsNotCleared(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	g.Insert(first)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyEnter},
	}
	g.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("non-Tab keyboard event was cleared by Tab interception logic")
	}
}

// TestNonTabKeyboardEventDoesNotChangeFocus verifies that a non-Tab key does not
// accidentally trigger focus traversal.
// Spec: only Tab and Shift+Tab are intercepted.
func TestNonTabKeyboardEventDoesNotChangeFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyEnter},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() != first {
		t.Errorf("non-Tab key changed focus: FocusedChild() = %v, want first", g.FocusedChild())
	}
}

// ── Tab traversal is symmetric ───────────────────────────────────────────────

// TestTabAndShiftTabAreSymmetric verifies that Tab followed by Shift+Tab returns
// to the original child (forward and reverse are truly opposite).
// Spec: "focusPrev() does the reverse — skips non-selectable children, wraps
// from beginning to end."
func TestTabAndShiftTabAreSymmetric(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	third := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	g.Insert(third)
	g.SetFocusedChild(first)

	g.HandleEvent(tabEvent())
	// Should now be on second.
	g.HandleEvent(shiftTabEvent())
	// Should be back on first.

	if g.FocusedChild() != first {
		t.Errorf("Tab then Shift+Tab: FocusedChild() = %v, want first (symmetry broken)", g.FocusedChild())
	}
}

// TestShiftTabAndTabAreSymmetric verifies that Shift+Tab followed by Tab returns
// to the original child.
// Spec: "focusPrev() does the reverse".
func TestShiftTabAndTabAreSymmetric(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	third := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	g.Insert(third)
	// third is auto-focused.

	g.HandleEvent(shiftTabEvent())
	// Should now be on second.
	g.HandleEvent(tabEvent())
	// Should be back on third.

	if g.FocusedChild() != third {
		t.Errorf("Shift+Tab then Tab: FocusedChild() = %v, want third (symmetry broken)", g.FocusedChild())
	}
}

// TestTabWithModifiersIsNotIntercepted verifies that a Tab key WITH modifiers
// (e.g., Ctrl+Tab) is not treated as focus-traversal Tab.
// Spec: "intercepts Tab key (tcell.KeyTab, no modifiers)".
func TestTabWithModifiersIsNotIntercepted(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	// Ctrl+Tab — has a modifier, should NOT be intercepted.
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyTab, Modifiers: tcell.ModCtrl},
	}
	g.HandleEvent(ev)

	// Focus must not have advanced.
	if g.FocusedChild() != first {
		t.Errorf("Ctrl+Tab (with modifier) changed focus — only plain Tab should be intercepted")
	}
	// Event must not have been cleared by Tab logic.
	if ev.IsCleared() {
		t.Errorf("Ctrl+Tab event was cleared — only plain Tab should be intercepted")
	}
}

// ── Empty-group edge cases ────────────────────────────────────────────────────

// TestTabOnEmptyGroupDoesNotPanic verifies that sending Tab to a Group with zero
// children does not panic. Tab is no longer intercepted by Group (spec 13.3).
func TestTabOnEmptyGroupDoesNotPanic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	// No children inserted — the children slice is empty.
	// Must not panic; the event is NOT cleared (Group no longer intercepts Tab).
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Tab on empty group panicked: %v", r)
		}
	}()
	g.HandleEvent(tabEvent())
}

// TestShiftTabOnEmptyGroupDoesNotPanic verifies that Shift+Tab on an empty
// Group does not panic. Tab handling was moved to Window (spec 13.3).
func TestShiftTabOnEmptyGroupDoesNotPanic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Shift+Tab on empty group panicked: %v", r)
		}
	}()
	g.HandleEvent(shiftTabEvent())
}
