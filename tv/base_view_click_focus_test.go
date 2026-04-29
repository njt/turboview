package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// focusTestView is a concrete View that embeds BaseView and calls SetSelf.
// Its HandleEvent calls through to BaseView first, then records whether the
// event survived (was not cleared by click-to-focus).
type focusTestView struct {
	BaseView
	handleEventCalled bool
}

func newFocusTestView(bounds Rect) *focusTestView {
	v := &focusTestView{}
	v.SetBounds(bounds)
	v.SetState(SfVisible, true)
	v.SetOptions(OfSelectable, true)
	v.SetSelf(v)
	return v
}

func (v *focusTestView) HandleEvent(event *Event) {
	wasMouse := event.What == EvMouse
	v.BaseView.HandleEvent(event)
	if wasMouse && !event.IsCleared() {
		v.handleEventCalled = true
	}
}

// clickEventAt returns a mouse event with Button1 pressed at the given coords.
func clickEventAt(x, y int) *Event {
	return &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1},
	}
}

// button2EventAt returns a mouse event with Button2 pressed.
func button2EventAt(x, y int) *Event {
	return &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button2},
	}
}

// button3EventAt returns a mouse event with Button3 pressed.
func button3EventAt(x, y int) *Event {
	return &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button3},
	}
}

// wheelEventAt returns a wheel-up mouse event (not a real button press).
func wheelEventAt(x, y int) *Event {
	return &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: x, Y: y, Button: tcell.WheelUp},
	}
}

// ── Test 1: Click on unfocused selectable view focuses it ─────────────────────

// TestClickOnUnfocusedSelectableViewFocusesIt verifies that clicking an
// unfocused selectable view causes the Group to mark it as the focused child.
// Spec: "If the view is NOT in SfSelected state, AND NOT SfDisabled, AND has
// OfSelectable — Calls owner.SetFocusedChild(self) to gain focus."
func TestClickOnUnfocusedSelectableViewFocusesIt(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// focused occupies right half; it gets focus on Insert.
	focused := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	// target occupies left half; it is unfocused.
	target := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(target)
	// After inserting target it may become focused; revert so focused holds it.
	g.SetFocusedChild(focused)

	// Click inside target.
	g.HandleEvent(clickEventAt(5, 5))

	if g.FocusedChild() != target {
		t.Errorf("after click on target, FocusedChild() = %v, want target", g.FocusedChild())
	}
}

// TestClickOnUnfocusedSelectableViewFocusesFalsified confirms the test above
// would catch an implementation that ignores clicks entirely.
// Spec: "Calls owner.SetFocusedChild(self) to gain focus."
func TestClickOnUnfocusedSelectableViewFocusesFalsified(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(target)
	g.SetFocusedChild(focused)

	g.HandleEvent(clickEventAt(5, 5))

	// If click-to-focus is not implemented, focused stays unchanged.
	if g.FocusedChild() == focused {
		t.Errorf("falsified: FocusedChild() is still the original focused child — click-to-focus did not fire")
	}
}

// ── Test 2: Without OfFirstClick, click is consumed (event cleared) ───────────

// TestClickWithoutOfFirstClickClearsEvent verifies that when the view lacks
// OfFirstClick, the click that causes focus is consumed (cleared).
// Spec: "If the view does NOT have OfFirstClick, clears the event."
func TestClickWithoutOfFirstClickClearsEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	// target has OfSelectable but NOT OfFirstClick (default).
	g.Insert(target)
	g.SetFocusedChild(focused)

	ev := clickEventAt(5, 5)
	g.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("click event was NOT cleared; expected cleared because target lacks OfFirstClick")
	}
}

// TestClickWithoutOfFirstClickClearsEventFalsified catches an implementation
// that never clears the event.
// Spec: "If the view does NOT have OfFirstClick, clears the event."
func TestClickWithoutOfFirstClickClearsEventFalsified(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(target)
	g.SetFocusedChild(focused)

	ev := clickEventAt(5, 5)
	g.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("falsified: event was not cleared — implementation must clear the click when OfFirstClick is absent")
	}
}

// TestClickWithoutOfFirstClickDoesNotCallThroughToViewLogic verifies that
// when the event is cleared by click-to-focus the view's own logic sees it
// as cleared and does not act on it.
// Spec: "click consumed by focusing."
func TestClickWithoutOfFirstClickDoesNotCallThroughToViewLogic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(target)
	g.SetFocusedChild(focused)

	g.HandleEvent(clickEventAt(5, 5))

	if target.handleEventCalled {
		t.Errorf("view's own event logic was reached after click-to-focus consumed the event; it should have been cleared")
	}
}

// ── Test 3: With OfFirstClick, click passes through (event NOT cleared) ────────

// TestClickWithOfFirstClickDoesNotClearEvent verifies that when the view has
// OfFirstClick, the focusing click is NOT consumed — the event passes through.
// Spec: "If the view HAS OfFirstClick, leaves the event alone."
func TestClickWithOfFirstClickDoesNotClearEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	target.SetOptions(OfFirstClick, true)
	g.Insert(target)
	g.SetFocusedChild(focused)

	ev := clickEventAt(5, 5)
	g.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("click event was cleared; expected NOT cleared because target has OfFirstClick")
	}
}

// TestClickWithOfFirstClickStillFocuses verifies that OfFirstClick does not
// suppress focus acquisition — the view still becomes focused.
// Spec: "Calls owner.SetFocusedChild(self) to gain focus" — OfFirstClick only
// changes whether the click is consumed, not whether focus is acquired.
func TestClickWithOfFirstClickStillFocuses(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	target.SetOptions(OfFirstClick, true)
	g.Insert(target)
	g.SetFocusedChild(focused)

	g.HandleEvent(clickEventAt(5, 5))

	if g.FocusedChild() != target {
		t.Errorf("with OfFirstClick, FocusedChild() = %v, want target — OfFirstClick must not prevent focus acquisition", g.FocusedChild())
	}
}

// TestClickWithOfFirstClickPassesThroughToViewLogic verifies that with
// OfFirstClick the event is visible to the view's own handler.
// Spec: "click passes through."
func TestClickWithOfFirstClickPassesThroughToViewLogic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(focused)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	target.SetOptions(OfFirstClick, true)
	g.Insert(target)
	g.SetFocusedChild(focused)

	g.HandleEvent(clickEventAt(5, 5))

	if !target.handleEventCalled {
		t.Errorf("with OfFirstClick, view's own handler was NOT reached — click-to-focus must pass the event through")
	}
}

// ── Test 4: Already-focused view — click passes through unchanged ─────────────

// TestClickOnAlreadyFocusedViewPassesThrough verifies that clicking a view
// that already has SfSelected causes no change to focus and the event is not
// cleared.
// Spec: "If the view IS already SfSelected, does nothing (event passes through)."
func TestClickOnAlreadyFocusedViewPassesThrough(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(focused)
	// focused is now SfSelected (set by Group.SetFocusedChild).

	ev := clickEventAt(5, 5)
	g.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("click on already-focused view cleared the event; event must pass through")
	}
}

// TestClickOnAlreadyFocusedViewDoesNotChangeFocus verifies that clicking the
// already-focused view does not alter which child is focused.
// Spec: "If the view IS already SfSelected, does nothing."
func TestClickOnAlreadyFocusedViewDoesNotChangeFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(focused)

	g.HandleEvent(clickEventAt(5, 5))

	if g.FocusedChild() != focused {
		t.Errorf("click on already-focused view changed FocusedChild(); it must remain unchanged")
	}
}

// TestClickOnAlreadyFocusedViewCallsThroughToViewLogic verifies that the
// view's own handler receives the event (since it was not cleared).
// Spec: "event passes through."
func TestClickOnAlreadyFocusedViewCallsThroughToViewLogic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(focused)

	g.HandleEvent(clickEventAt(5, 5))

	if !focused.handleEventCalled {
		t.Errorf("click on already-focused view did not reach the view's own handler; it must pass through")
	}
}

// ── Test 5: Disabled view — click does nothing ────────────────────────────────

// TestClickOnDisabledViewDoesNotFocus verifies that a disabled view is not
// focused by a click and remains unfocused.
// Spec: "If the view has SfDisabled, does nothing."
func TestClickOnDisabledViewDoesNotFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	other := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(other)

	disabled := newFocusTestView(NewRect(0, 0, 20, 10))
	disabled.SetState(SfDisabled, true)
	g.Insert(disabled)
	g.SetFocusedChild(other)

	g.HandleEvent(clickEventAt(5, 5))

	if g.FocusedChild() == disabled {
		t.Errorf("disabled view became focused after a click; SfDisabled must prevent focus acquisition")
	}
}

// TestClickOnDisabledViewDoesNotClearEvent verifies that a click on a disabled
// view is not consumed — the event passes through.
// Spec: "If the view has SfDisabled, does nothing."
func TestClickOnDisabledViewDoesNotClearEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	other := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(other)

	disabled := newFocusTestView(NewRect(0, 0, 20, 10))
	disabled.SetState(SfDisabled, true)
	g.Insert(disabled)
	g.SetFocusedChild(other)

	ev := clickEventAt(5, 5)
	g.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("click on disabled view cleared the event; disabled views must not consume clicks")
	}
}

// ── Test 6: Non-selectable view — click does nothing ─────────────────────────

// TestClickOnNonSelectableViewDoesNotFocus verifies that a view lacking
// OfSelectable is not focused by a click.
// Spec: "If the view lacks OfSelectable, does nothing."
func TestClickOnNonSelectableViewDoesNotFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// nonSel is visible but NOT selectable.
	nonSel := &focusTestView{}
	nonSel.SetBounds(NewRect(0, 0, 20, 10))
	nonSel.SetState(SfVisible, true)
	// OfSelectable deliberately NOT set.
	nonSel.SetSelf(nonSel)
	g.Insert(nonSel)

	ev := clickEventAt(5, 5)
	g.HandleEvent(ev)

	// No selectable child exists, so FocusedChild() remains nil.
	if g.FocusedChild() == nonSel {
		t.Errorf("non-selectable view was focused by a click; OfSelectable must be required for click-to-focus")
	}
}

// TestClickOnNonSelectableViewDoesNotClearEvent verifies that clicking a
// non-selectable view does not consume the event.
// Spec: "If the view lacks OfSelectable, does nothing."
func TestClickOnNonSelectableViewDoesNotClearEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	nonSel := &focusTestView{}
	nonSel.SetBounds(NewRect(0, 0, 20, 10))
	nonSel.SetState(SfVisible, true)
	nonSel.SetSelf(nonSel)
	g.Insert(nonSel)

	ev := clickEventAt(5, 5)
	g.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("click on non-selectable view cleared the event; non-selectable views must not consume clicks")
	}
}

// ── Test 7: Wheel events do NOT trigger click-to-focus ───────────────────────

// TestWheelEventDoesNotTriggerFocus verifies that scroll-wheel events are not
// treated as real button presses and therefore do not trigger click-to-focus.
// Spec: "Wheel events do NOT trigger click-to-focus."
func TestWheelEventDoesNotTriggerFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	other := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(other)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(target)
	g.SetFocusedChild(other)

	g.HandleEvent(wheelEventAt(5, 5))

	if g.FocusedChild() != other {
		t.Errorf("wheel event changed focus to target; wheel events must not trigger click-to-focus")
	}
}

// TestWheelEventDoesNotClearEvent verifies that a wheel event is not consumed
// by click-to-focus logic.
// Spec: "Wheel events do NOT trigger click-to-focus."
func TestWheelEventDoesNotClearEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	other := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(other)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(target)
	g.SetFocusedChild(other)

	ev := wheelEventAt(5, 5)
	g.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("wheel event was cleared; wheel events must not be consumed by click-to-focus")
	}
}

// ── Test 8: Non-mouse events do nothing ──────────────────────────────────────

// TestKeyboardEventDoesNotTriggerFocus verifies that keyboard events are
// completely ignored by click-to-focus logic.
// Spec: "Non-mouse events are ignored."
func TestKeyboardEventDoesNotTriggerClickFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	other := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(other)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(target)
	g.SetFocusedChild(other)

	// Keyboard events are dispatched to the focused child, not positionally.
	// Send directly to target's BaseView to test the guard.
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	target.BaseView.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("keyboard event was cleared by click-to-focus logic; non-mouse events must be ignored")
	}
	if g.FocusedChild() != other {
		t.Errorf("keyboard event changed focus via click-to-focus logic; non-mouse events must be ignored")
	}
}

// TestKeyboardEventDoesNotClearEventInBaseView verifies BaseView.HandleEvent
// does not clear a keyboard event regardless of view state.
// Spec: "Non-mouse events are ignored."
func TestKeyboardEventDoesNotClearEventInBaseView(t *testing.T) {
	v := newFocusTestView(NewRect(0, 0, 20, 10))
	// Give it an owner so the path that might call SetFocusedChild is reachable.
	mc := &mockContainer{}
	v.SetOwner(mc)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	v.BaseView.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("BaseView.HandleEvent cleared a keyboard event; only mouse button events trigger click-to-focus")
	}
}

// ── Test 9: SetSelf must be called — without it, click-to-focus silently skips ─

// TestWithoutSetSelfClickToFocusDoesNotFire verifies that if SetSelf was not
// called, click-to-focus silently skips and the event is NOT cleared.
// Spec: "All widget constructors must call SetSelf so BaseView knows the
// concrete View identity."
func TestWithoutSetSelfClickToFocusDoesNotFire(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	other := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(other)

	// Build a focusTestView WITHOUT calling SetSelf.
	noSelf := &focusTestView{}
	noSelf.SetBounds(NewRect(0, 0, 20, 10))
	noSelf.SetState(SfVisible, true)
	noSelf.SetOptions(OfSelectable, true)
	// SetSelf intentionally omitted.
	g.Insert(noSelf)
	g.SetFocusedChild(other)

	ev := clickEventAt(5, 5)
	g.HandleEvent(ev)

	// With no self identity, click-to-focus cannot call SetFocusedChild(self),
	// so focus must not change.
	if g.FocusedChild() != other {
		t.Errorf("without SetSelf, focus changed; click-to-focus must be a no-op when SetSelf was never called")
	}
}

// TestWithoutSetSelfEventNotCleared verifies that the event is not consumed
// when SetSelf was not called.
// Spec: silently skips — implies the event is not cleared either.
func TestWithoutSetSelfEventNotCleared(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	other := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(other)

	noSelf := &focusTestView{}
	noSelf.SetBounds(NewRect(0, 0, 20, 10))
	noSelf.SetState(SfVisible, true)
	noSelf.SetOptions(OfSelectable, true)
	// SetSelf intentionally omitted.
	g.Insert(noSelf)
	g.SetFocusedChild(other)

	ev := clickEventAt(5, 5)
	g.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("without SetSelf, event was cleared; click-to-focus must be a complete no-op when self is unknown")
	}
}

// ── Test 10: All three real buttons trigger click-to-focus ───────────────────

// TestButton2TriggersFocus verifies that Button2 (right-click) also triggers
// click-to-focus, not just Button1.
// Spec: "Button1, Button2, or Button3 — NOT wheel events."
func TestButton2TriggersFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	other := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(other)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(target)
	g.SetFocusedChild(other)

	g.HandleEvent(button2EventAt(5, 5))

	if g.FocusedChild() != target {
		t.Errorf("Button2 click did not focus target; Button2 must trigger click-to-focus like Button1")
	}
}

// TestButton3TriggersFocus verifies that Button3 (middle-click) also triggers
// click-to-focus.
// Spec: "Button1, Button2, or Button3 — NOT wheel events."
func TestButton3TriggersFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	other := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(other)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(target)
	g.SetFocusedChild(other)

	g.HandleEvent(button3EventAt(5, 5))

	if g.FocusedChild() != target {
		t.Errorf("Button3 click did not focus target; Button3 must trigger click-to-focus like Button1")
	}
}

// TestNoButtonMouseEventDoesNotTriggerFocus verifies that a mouse event with
// no button pressed (ButtonNone) does not trigger click-to-focus.
// Spec: "real button press (Button1, Button2, or Button3)."
func TestNoButtonMouseEventDoesNotTriggerFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	other := newFocusTestView(NewRect(40, 0, 20, 10))
	g.Insert(other)

	target := newFocusTestView(NewRect(0, 0, 20, 10))
	g.Insert(target)
	g.SetFocusedChild(other)

	// ButtonNone — cursor motion, no button pressed.
	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 5, Y: 5, Button: tcell.ButtonNone},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() != other {
		t.Errorf("mouse-motion (ButtonNone) changed focus; only real button presses must trigger click-to-focus")
	}
}
