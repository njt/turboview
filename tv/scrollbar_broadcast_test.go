package tv

// scrollbar_broadcast_test.go — Tests for Task 6: ScrollBar Broadcasts (spec 6.5).
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — CmScrollBarClicked broadcasts (mouse click, wheel, keyboard)
//   Section 2  — CmScrollBarChanged broadcasts (value-changing interactions)
//   Section 3  — Complementary with OnChange: both fire, neither replaces the other
//   Section 4  — Edge cases: clicked-without-change, no-changed-without-change, thumb drag
//   Section 5  — Falsifying: CmScrollBarClicked and CmScrollBarChanged are distinct

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newScrollBarInGroup creates a vertical ScrollBar and inserts it into a Group,
// alongside a broadcastSpyView. It returns the scrollbar, the group, and the spy.
// The spy receives all EvBroadcast events dispatched by the Group to its children.
func newScrollBarInGroup(bounds Rect) (*ScrollBar, *Group, *broadcastSpyView) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	sb := NewScrollBar(bounds, Vertical)
	spy := newNonSelectableSpyView()
	g.Insert(sb)
	g.Insert(spy)
	resetBroadcasts(spy)
	return sb, g, spy
}

// newHorizontalScrollBarInGroup creates a horizontal ScrollBar in a Group with a spy.
func newHorizontalScrollBarInGroup(bounds Rect) (*ScrollBar, *Group, *broadcastSpyView) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	sb := NewScrollBar(bounds, Horizontal)
	spy := newNonSelectableSpyView()
	g.Insert(sb)
	g.Insert(spy)
	resetBroadcasts(spy)
	return sb, g, spy
}

// hasBroadcast reports whether the spy received at least one broadcast with the given command.
func hasBroadcast(spy *broadcastSpyView, cmd CommandCode) bool {
	for _, rec := range spy.broadcasts {
		if rec.command == cmd {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Section 1 — CmScrollBarClicked broadcasts
// ---------------------------------------------------------------------------

// TestScrollBarMouseClickBroadcastsCmScrollBarClicked verifies that a mouse click
// on an arrow broadcasts CmScrollBarClicked to the owner.
// Spec: "When any scrollbar interaction begins (mouse click, wheel, keyboard):
// broadcast CmScrollBarClicked (new constant) to the owner."
func TestScrollBarMouseClickBroadcastsCmScrollBarClicked(t *testing.T) {
	// height=10, up arrow at Y=0
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("mouse click on arrow: spy did not receive CmScrollBarClicked; broadcasts: %v", spy.broadcasts)
	}
}

// TestScrollBarMouseWheelBroadcastsCmScrollBarClicked verifies that a mouse wheel
// event broadcasts CmScrollBarClicked to the owner.
// Spec: "When any scrollbar interaction begins (mouse click, wheel, keyboard):
// broadcast CmScrollBarClicked (new constant) to the owner."
func TestScrollBarMouseWheelBroadcastsCmScrollBarClicked(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("mouse wheel: spy did not receive CmScrollBarClicked; broadcasts: %v", spy.broadcasts)
	}
}

// TestScrollBarKeyboardEventBroadcastsCmScrollBarClicked verifies that a keyboard
// event broadcasts CmScrollBarClicked to the owner.
// Spec: "When any scrollbar interaction begins (mouse click, wheel, keyboard):
// broadcast CmScrollBarClicked (new constant) to the owner."
func TestScrollBarKeyboardEventBroadcastsCmScrollBarClicked(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	sb.HandleEvent(ev)

	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("keyboard event: spy did not receive CmScrollBarClicked; broadcasts: %v", spy.broadcasts)
	}
}

// TestScrollBarClickedBroadcastInfoIsScrollBar verifies that the Info field of
// the CmScrollBarClicked broadcast is the ScrollBar itself.
// Spec: "The Info field of the broadcast event should be the ScrollBar itself
// (so the owner can identify which scrollbar changed)."
func TestScrollBarClickedBroadcastInfoIsScrollBar(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	for _, rec := range spy.broadcasts {
		if rec.command == CmScrollBarClicked {
			if rec.info != sb {
				t.Errorf("CmScrollBarClicked Info = %v (%T), want the ScrollBar itself (%p)", rec.info, rec.info, sb)
			}
			return
		}
	}
	t.Error("CmScrollBarClicked broadcast not received at all")
}

// ---------------------------------------------------------------------------
// Section 2 — CmScrollBarChanged broadcasts
// ---------------------------------------------------------------------------

// TestScrollBarMouseClickThatChangesValueBroadcastsCmScrollBarChanged verifies
// that a mouse click which changes the scrollbar value broadcasts CmScrollBarChanged.
// Spec: "When the scrollbar value changes, broadcast CmScrollBarChanged (new constant)
// to the owner."
func TestScrollBarMouseClickThatChangesValueBroadcastsCmScrollBarChanged(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	// Click up arrow at Y=0 — value changes from 50 to 49
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("mouse click (value changed): spy did not receive CmScrollBarChanged; broadcasts: %v", spy.broadcasts)
	}
}

// TestScrollBarKeyboardThatChangesValueBroadcastsCmScrollBarChanged verifies
// that a keyboard event which changes the scrollbar value broadcasts CmScrollBarChanged.
// Spec: "When the scrollbar value changes, broadcast CmScrollBarChanged (new constant)
// to the owner."
func TestScrollBarKeyboardThatChangesValueBroadcastsCmScrollBarChanged(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	sb.HandleEvent(ev)

	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("keyboard event (value changed): spy did not receive CmScrollBarChanged; broadcasts: %v", spy.broadcasts)
	}
}

// TestScrollBarMouseWheelThatChangesValueBroadcastsCmScrollBarChanged verifies
// that a wheel event which changes the scrollbar value broadcasts CmScrollBarChanged.
// Spec: "When the scrollbar value changes, broadcast CmScrollBarChanged (new constant)
// to the owner."
func TestScrollBarMouseWheelThatChangesValueBroadcastsCmScrollBarChanged(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("mouse wheel (value changed): spy did not receive CmScrollBarChanged; broadcasts: %v", spy.broadcasts)
	}
}

// TestScrollBarChangedBroadcastInfoIsScrollBar verifies that the Info field of
// the CmScrollBarChanged broadcast is the ScrollBar itself.
// Spec: "The Info field of the broadcast event should be the ScrollBar itself
// (so the owner can identify which scrollbar changed)."
func TestScrollBarChangedBroadcastInfoIsScrollBar(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	// Down arrow at Y=9 changes value 50 → 51
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	for _, rec := range spy.broadcasts {
		if rec.command == CmScrollBarChanged {
			if rec.info != sb {
				t.Errorf("CmScrollBarChanged Info = %v (%T), want the ScrollBar itself (%p)", rec.info, rec.info, sb)
			}
			return
		}
	}
	t.Error("CmScrollBarChanged broadcast not received at all")
}

// ---------------------------------------------------------------------------
// Section 3 — Complementary with OnChange
// ---------------------------------------------------------------------------

// TestScrollBarBothOnChangeAndCmScrollBarChangedFireOnValueChange verifies that
// both OnChange AND CmScrollBarChanged are triggered when the value changes.
// Spec: "These broadcasts complement (don't replace) the existing OnChange callback."
// Spec: "Both broadcasts AND OnChange fire (not one or the other)."
func TestScrollBarBothOnChangeAndCmScrollBarChangedFireOnValueChange(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	onChangeFired := false
	sb.OnChange = func(v int) { onChangeFired = true }

	// Click up arrow — value changes from 50 to 49
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if !onChangeFired {
		t.Error("OnChange did not fire after value-changing interaction")
	}
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Error("CmScrollBarChanged broadcast did not fire after value-changing interaction")
	}
}

// TestScrollBarOnChangeIsNotReplacedByBroadcast verifies that setting up a
// broadcast mechanism does not suppress OnChange.
// Spec: "These broadcasts complement (don't replace) the existing OnChange callback."
func TestScrollBarOnChangeIsNotReplacedByBroadcast(t *testing.T) {
	sb, _, _ := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	callCount := 0
	sb.OnChange = func(v int) { callCount++ }

	// Trigger two value-changing interactions
	ev1 := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}} // up arrow → 49
	ev2 := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}} // down arrow → 50
	sb.HandleEvent(ev1)
	sb.HandleEvent(ev2)

	if callCount != 2 {
		t.Errorf("OnChange called %d time(s), want 2 — OnChange must not be suppressed by broadcast", callCount)
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Edge cases
// ---------------------------------------------------------------------------

// TestScrollBarClickedFiresEvenWhenValueDoesNotChange verifies CmScrollBarClicked
// is broadcast even when the interaction does not change the value (e.g., clicking
// up arrow while already at minimum).
// Spec: "When any scrollbar interaction begins (mouse click, wheel, keyboard):
// broadcast CmScrollBarClicked (new constant) to the owner."
// Note: the spec says "interaction begins", not "value changes", so it fires always.
func TestScrollBarClickedFiresEvenWhenValueDoesNotChange(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0) // already at minimum

	// Click up arrow — value is already at min and will not change
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("CmScrollBarClicked should fire even when value does not change (boundary click); broadcasts: %v", spy.broadcasts)
	}
}

// TestScrollBarChangedDoesNotFireWhenValueDoesNotChange verifies CmScrollBarChanged
// is NOT broadcast when an interaction does not change the value.
// Spec: "When the scrollbar value changes, broadcast CmScrollBarChanged (new constant)."
// Value unchanged → no CmScrollBarChanged.
func TestScrollBarChangedDoesNotFireWhenValueDoesNotChange(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0) // already at minimum

	// Click up arrow — value is already at min and will not change
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("CmScrollBarChanged must NOT fire when value does not change; broadcasts: %v", spy.broadcasts)
	}
}

// TestScrollBarThumbDragInitialClickBroadcastsCmScrollBarClicked verifies that
// the initial mouse press on the thumb (starting a drag) broadcasts CmScrollBarClicked.
// Spec: "When any scrollbar interaction begins (mouse click, wheel, keyboard):
// broadcast CmScrollBarClicked (new constant) to the owner."
// A thumb-press is still an interaction begin.
func TestScrollBarThumbDragInitialClickBroadcastsCmScrollBarClicked(t *testing.T) {
	// height=12, value=0 → thumb at Y=1
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on the thumb — this starts a drag without immediately changing value
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("thumb drag initial click: spy did not receive CmScrollBarClicked; broadcasts: %v", spy.broadcasts)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Falsifying: distinct commands
// ---------------------------------------------------------------------------

// TestCmScrollBarClickedAndCmScrollBarChangedAreDistinct verifies that
// CmScrollBarClicked and CmScrollBarChanged are different command constants.
// A value-changing click broadcasts both; they must be distinguishable.
// Spec: "CmScrollBarClicked (new constant)" vs "CmScrollBarChanged (new constant)"
func TestCmScrollBarClickedAndCmScrollBarChangedAreDistinct(t *testing.T) {
	if CmScrollBarClicked == CmScrollBarChanged {
		t.Errorf("CmScrollBarClicked (%d) == CmScrollBarChanged (%d); they must be distinct constants",
			CmScrollBarClicked, CmScrollBarChanged)
	}
}

// TestCmScrollBarClickedBroadcastIsDistinguishableFromChanged verifies that a
// boundary click (value unchanged) produces CmScrollBarClicked but NOT
// CmScrollBarChanged — confirming the two commands carry different semantics.
func TestCmScrollBarClickedBroadcastIsDistinguishableFromChanged(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0) // at minimum — up arrow click will not change value

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	gotClicked := hasBroadcast(spy, CmScrollBarClicked)
	gotChanged := hasBroadcast(spy, CmScrollBarChanged)

	if !gotClicked {
		t.Error("boundary click: expected CmScrollBarClicked, but did not receive it")
	}
	if gotChanged {
		t.Error("boundary click: received CmScrollBarChanged, but value did not change — these commands must be distinct")
	}
}

// TestScrollBarChangedNotEmittedWhenOnlyClickedIsExpected verifies that for a
// value-changing interaction, CmScrollBarChanged references a different command
// code than CmScrollBarClicked (both are present but as separate records).
func TestScrollBarChangedNotEmittedWhenOnlyClickedIsExpected(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 10))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	// Down arrow changes value; both CmScrollBarClicked and CmScrollBarChanged fire.
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	clickedCount := 0
	changedCount := 0
	for _, rec := range spy.broadcasts {
		if rec.command == CmScrollBarClicked {
			clickedCount++
		}
		if rec.command == CmScrollBarChanged {
			changedCount++
		}
	}

	if clickedCount == 0 {
		t.Error("value-changing click: CmScrollBarClicked not received")
	}
	if changedCount == 0 {
		t.Error("value-changing click: CmScrollBarChanged not received")
	}
	// Confirm they are recorded as distinct broadcasts (not the same entry matching both).
	// If CmScrollBarClicked == CmScrollBarChanged the test above would fail first.
	if clickedCount > 0 && changedCount > 0 && CmScrollBarClicked == CmScrollBarChanged {
		t.Error("CmScrollBarClicked and CmScrollBarChanged have the same constant value — they must be distinct")
	}
}
