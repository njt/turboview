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

// ── Tab advances focus ────────────────────────────────────────────────────────

// TestTabAdvancesFocusToNextSelectableChild verifies that sending a Tab event
// moves focus from the first selectable child to the second.
// Spec: "Group.HandleEvent intercepts Tab key ... and advances focus to the next
// OfSelectable child (wrapping around). Clears the event."
func TestTabAdvancesFocusToNextSelectableChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	// Insert auto-focuses the last selectable; set focus back to first.
	g.SetFocusedChild(first)

	g.HandleEvent(tabEvent())

	if g.FocusedChild() != second {
		t.Errorf("after Tab, FocusedChild() = %v, want second", g.FocusedChild())
	}
}

// TestTabFocusFalsifiedDoesNotStayOnSameChild catches an implementation that
// ignores Tab entirely and leaves focus unchanged.
// Spec: "advances focus to the next OfSelectable child".
func TestTabFocusFalsifiedDoesNotStayOnSameChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	g.HandleEvent(tabEvent())

	if g.FocusedChild() == first {
		t.Errorf("after Tab, FocusedChild() is still first — Tab did not advance focus")
	}
}

// TestTabWrapsFromLastToFirst verifies that Tab on the last selectable child
// wraps focus back to the first selectable child.
// Spec: "advances focus to the next OfSelectable child (wrapping around)".
func TestTabWrapsFromLastToFirst(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	// second is auto-focused last; it is the last selectable child.

	g.HandleEvent(tabEvent())

	if g.FocusedChild() != first {
		t.Errorf("Tab from last selectable: FocusedChild() = %v, want first (wrap)", g.FocusedChild())
	}
}

// TestTabWrapFalsifiedDoesNotStopAtEnd catches an implementation that stops at
// the last child instead of wrapping.
// Spec: "(wrapping around)".
func TestTabWrapFalsifiedDoesNotStopAtEnd(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	// second is focused (last inserted selectable).

	g.HandleEvent(tabEvent())

	if g.FocusedChild() == second {
		t.Errorf("Tab from last selectable: FocusedChild() is still second — wrap did not occur")
	}
}

// TestTabSkipsNonSelectableChildren verifies that non-selectable children are
// not candidates for focus during Tab traversal.
// Spec: "focusNext() skips non-selectable children".
func TestTabSkipsNonSelectableChildren(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	nonSel := newNonSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(nonSel)
	g.Insert(second)
	g.SetFocusedChild(first)

	g.HandleEvent(tabEvent())

	if g.FocusedChild() != second {
		t.Errorf("after Tab, FocusedChild() = %v, want second (skipped non-selectable)", g.FocusedChild())
	}
}

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

// TestTabClearsTheEvent verifies that Tab interception clears the event so that
// children do not see it.
// Spec: "Clears the event."
func TestTabClearsTheEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	ev := tabEvent()
	g.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("after Tab, event.IsCleared() = false, want true")
	}
}

// TestTabSetsNewChildSfSelected verifies that after Tab, the new focused child
// gains SfSelected state.
// Spec: "Focus traversal uses selectChild(v) to set SfSelected on the new child
// and clear it on the old one."
func TestTabSetsNewChildSfSelected(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	g.HandleEvent(tabEvent())

	if !second.HasState(SfSelected) {
		t.Errorf("after Tab, new focused child does not have SfSelected")
	}
}

// TestTabClearsSfSelectedOnOldChild verifies that after Tab, the previously
// focused child loses SfSelected state.
// Spec: "Focus traversal uses selectChild(v) to set SfSelected on the new child
// and clear it on the old one."
func TestTabClearsSfSelectedOnOldChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	g.HandleEvent(tabEvent())

	if first.HasState(SfSelected) {
		t.Errorf("after Tab, old focused child still has SfSelected")
	}
}

// ── Shift+Tab moves focus backward ──────────────────────────────────────────

// TestShiftTabMovesFocusToPreviousSelectableChild verifies that Shift+Tab moves
// focus from the second selectable child back to the first.
// Spec: "Group.HandleEvent intercepts Shift+Tab key (tcell.KeyBacktab) ... and
// moves focus to the previous OfSelectable child (wrapping around). Clears the
// event."
func TestShiftTabMovesFocusToPreviousSelectableChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	// second is auto-focused.

	g.HandleEvent(shiftTabEvent())

	if g.FocusedChild() != first {
		t.Errorf("after Shift+Tab, FocusedChild() = %v, want first", g.FocusedChild())
	}
}

// TestShiftTabFalsifiedDoesNotStayOnSameChild catches an implementation that
// ignores Shift+Tab.
// Spec: "moves focus to the previous OfSelectable child".
func TestShiftTabFalsifiedDoesNotStayOnSameChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	// second is auto-focused.

	g.HandleEvent(shiftTabEvent())

	if g.FocusedChild() == second {
		t.Errorf("after Shift+Tab, FocusedChild() is still second — Shift+Tab did not move focus")
	}
}

// TestShiftTabWrapsFromFirstToLast verifies that Shift+Tab on the first
// selectable child wraps focus to the last selectable child.
// Spec: "focusPrev() ... wraps from beginning to end."
func TestShiftTabWrapsFromFirstToLast(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	g.HandleEvent(shiftTabEvent())

	if g.FocusedChild() != second {
		t.Errorf("Shift+Tab from first selectable: FocusedChild() = %v, want second (wrap)", g.FocusedChild())
	}
}

// TestShiftTabWrapFalsifiedDoesNotStopAtBeginning catches an implementation
// that stops instead of wrapping from the beginning.
// Spec: "wraps from beginning to end".
func TestShiftTabWrapFalsifiedDoesNotStopAtBeginning(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	g.HandleEvent(shiftTabEvent())

	if g.FocusedChild() == first {
		t.Errorf("Shift+Tab from first selectable: FocusedChild() is still first — wrap did not occur")
	}
}

// TestShiftTabSkipsNonSelectableChildren verifies that non-selectable children
// are skipped when traversing backwards.
// Spec: "focusPrev() ... skips non-selectable children".
func TestShiftTabSkipsNonSelectableChildren(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	nonSel := newNonSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(nonSel)
	g.Insert(second)
	// second is auto-focused.

	g.HandleEvent(shiftTabEvent())

	if g.FocusedChild() != first {
		t.Errorf("after Shift+Tab, FocusedChild() = %v, want first (skipped non-selectable)", g.FocusedChild())
	}
}

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

// TestShiftTabClearsTheEvent verifies that Shift+Tab interception clears the event.
// Spec: "Clears the event."
func TestShiftTabClearsTheEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	// second is auto-focused.

	ev := shiftTabEvent()
	g.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("after Shift+Tab, event.IsCleared() = false, want true")
	}
}

// TestShiftTabSetsNewChildSfSelected verifies the new focused child gains
// SfSelected after Shift+Tab.
// Spec: "Focus traversal uses selectChild(v) to set SfSelected on the new child
// and clear it on the old one."
func TestShiftTabSetsNewChildSfSelected(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	// second is auto-focused.

	g.HandleEvent(shiftTabEvent())

	if !first.HasState(SfSelected) {
		t.Errorf("after Shift+Tab, new focused child does not have SfSelected")
	}
}

// TestShiftTabClearsSfSelectedOnOldChild verifies the old focused child loses
// SfSelected after Shift+Tab.
// Spec: "Focus traversal uses selectChild(v) to set SfSelected on the new child
// and clear it on the old one."
func TestShiftTabClearsSfSelectedOnOldChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	// second is auto-focused.

	g.HandleEvent(shiftTabEvent())

	if second.HasState(SfSelected) {
		t.Errorf("after Shift+Tab, old focused child still has SfSelected")
	}
}

// ── Interception happens before three-phase dispatch ─────────────────────────

// TestTabNotDeliveredToFocusedChild verifies that the focused child never sees
// the Tab event because interception happens before three-phase dispatch.
// Spec: "The Tab/Shift+Tab interception happens BEFORE the keyboard three-phase
// dispatch block, so the focused child never sees the Tab event."
func TestTabNotDeliveredToFocusedChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	g.HandleEvent(tabEvent())

	// Neither child should have received a Tab event via HandleEvent.
	if first.lastEvent != nil && first.lastEvent.Key != nil && first.lastEvent.Key.Key == tcell.KeyTab {
		t.Errorf("old focused child received the Tab event — interception did not happen before dispatch")
	}
	if second.lastEvent != nil && second.lastEvent.Key != nil && second.lastEvent.Key.Key == tcell.KeyTab {
		t.Errorf("new focused child received the Tab event — interception did not happen before dispatch")
	}
}

// TestShiftTabNotDeliveredToFocusedChild verifies the focused child never sees
// the Shift+Tab event.
// Spec: "so the focused child never sees the Tab event."
func TestShiftTabNotDeliveredToFocusedChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(first)
	g.Insert(second)
	// second is auto-focused.

	g.HandleEvent(shiftTabEvent())

	if second.lastEvent != nil && second.lastEvent.Key != nil && second.lastEvent.Key.Key == tcell.KeyBacktab {
		t.Errorf("old focused child received the Shift+Tab event — interception did not happen before dispatch")
	}
	if first.lastEvent != nil && first.lastEvent.Key != nil && first.lastEvent.Key.Key == tcell.KeyBacktab {
		t.Errorf("new focused child received the Shift+Tab event — interception did not happen before dispatch")
	}
}

// TestTabNotDeliveredToPreprocessChild verifies that a preprocess child does not
// receive the Tab event.
// Spec: "The Tab/Shift+Tab interception happens BEFORE the keyboard three-phase
// dispatch block."
func TestTabNotDeliveredToPreprocessChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	pre := &stubView{}
	pre.SetBounds(NewRect(0, 0, 10, 1))
	pre.SetState(SfVisible, true)
	pre.SetOptions(OfPreProcess, true)
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(pre)
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	g.HandleEvent(tabEvent())

	if pre.lastEvent != nil && pre.lastEvent.Key != nil && pre.lastEvent.Key.Key == tcell.KeyTab {
		t.Errorf("preprocess child received Tab event — interception must happen before three-phase dispatch")
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
// children does not panic and clears the event.
// Spec: "If no selectable children exist, nothing happens." A group with zero
// children has no selectable children, so the implementation must guard against
// indexing into an empty slice.
func TestTabOnEmptyGroupDoesNotPanic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	// No children inserted — the children slice is empty.

	ev := tabEvent()
	g.HandleEvent(ev) // must not panic

	if !ev.IsCleared() {
		t.Errorf("Tab on empty group: event.IsCleared() = false, want true")
	}
}

// TestShiftTabOnEmptyGroupDoesNotPanic is the Shift+Tab counterpart to
// TestTabOnEmptyGroupDoesNotPanic. A group with zero children must handle
// Shift+Tab without panicking and must clear the event.
// Spec: "If no selectable children exist, nothing happens."
func TestShiftTabOnEmptyGroupDoesNotPanic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	// No children inserted — the children slice is empty.

	ev := shiftTabEvent()
	g.HandleEvent(ev) // must not panic

	if !ev.IsCleared() {
		t.Errorf("Shift+Tab on empty group: event.IsCleared() = false, want true")
	}
}

// TestShiftTabNotDeliveredToPreprocessChild verifies that a preprocess child does
// not receive the Shift+Tab event. This is symmetric to
// TestTabNotDeliveredToPreprocessChild and confirms that the Shift+Tab
// interception also occurs before the three-phase dispatch that would otherwise
// deliver the event to OfPreProcess children.
// Spec: "The Tab/Shift+Tab interception happens BEFORE the keyboard three-phase
// dispatch block."
func TestShiftTabNotDeliveredToPreprocessChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	pre := &stubView{}
	pre.SetBounds(NewRect(0, 0, 10, 1))
	pre.SetState(SfVisible, true)
	pre.SetOptions(OfPreProcess, true)
	first := newSelectableView()
	second := newSelectableView()
	g.Insert(pre)
	g.Insert(first)
	g.Insert(second)
	// second is auto-focused (last inserted selectable).

	g.HandleEvent(shiftTabEvent())

	if pre.lastEvent != nil && pre.lastEvent.Key != nil && pre.lastEvent.Key.Key == tcell.KeyBacktab {
		t.Errorf("preprocess child received Shift+Tab event — interception must happen before three-phase dispatch")
	}
}
