package tv

// scrollbar_autorepeat_test.go — Tests for Task 3: ScrollBar Arrow Click Auto-Repeat (spec 6.2).
//
// Written to verify auto-repeat functionality already works correctly.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// The spec states: "When the user clicks and holds on a scrollbar arrow,
// the scroll action must repeat. This uses evMouseAuto (section 1.4):
// while the mouse button is held and the cursor is still on the arrow,
// repeat the step action."
//
// The existing code handles EvMouse events with Button1 held at arrow positions.
// The evMouseAuto system generates additional EvMouse events with Button1 still held,
// which the existing handler processes as additional clicks. So auto-repeat should work.
//
// Test organisation:
//   Section 1  — Vertical arrows: multiple clicks accumulate steps
//   Section 2  — Horizontal arrows: multiple clicks accumulate steps
//   Section 3  — Custom arStep values: steps are multiplied correctly
//   Section 4  — OnChange fires on each repeat
//   Section 5  — Events are consumed
//   Section 6  — Falsifying: value changes incrementally (not all-at-once)

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Section 1 — Vertical arrows: multiple clicks accumulate steps
// ---------------------------------------------------------------------------

// TestScrollBarAutoRepeatVerticalUpArrowFiveClicks verifies that 5 consecutive
// Button1 events on the vertical up arrow decrease value by 5*arStep.
// Spec: "Multiple EvMouse events with Button1 on the same arrow accumulate steps"
func TestScrollBarAutoRepeatVerticalUpArrowFiveClicks(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(1)

	// Simulate 5 consecutive Button1 events on up arrow (Y=0)
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if sb.Value() != 45 {
		t.Errorf("after 5 clicks on up arrow: Value() = %d, want 45 (50 - 5*1)", sb.Value())
	}
}

// TestScrollBarAutoRepeatVerticalDownArrowFiveClicks verifies that 5 consecutive
// Button1 events on the vertical down arrow increase value by 5*arStep.
// Spec: "Multiple EvMouse events with Button1 on the same arrow accumulate steps"
func TestScrollBarAutoRepeatVerticalDownArrowFiveClicks(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(1)

	// Simulate 5 consecutive Button1 events on down arrow (Y=height-1=9)
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if sb.Value() != 55 {
		t.Errorf("after 5 clicks on down arrow: Value() = %d, want 55 (50 + 5*1)", sb.Value())
	}
}

// TestScrollBarAutoRepeatVerticalUpArrowTenClicks verifies that even more
// consecutive events accumulate correctly.
// Spec: "Multiple EvMouse events with Button1 on the same arrow accumulate steps"
func TestScrollBarAutoRepeatVerticalUpArrowTenClicks(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(60)
	sb.SetArStep(1)

	// Simulate 10 consecutive Button1 events on up arrow
	for i := 0; i < 10; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if sb.Value() != 50 {
		t.Errorf("after 10 clicks on up arrow: Value() = %d, want 50 (60 - 10*1)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Horizontal arrows: multiple clicks accumulate steps
// ---------------------------------------------------------------------------

// TestScrollBarAutoRepeatHorizontalLeftArrowFiveClicks verifies that 5 consecutive
// Button1 events on the horizontal left arrow decrease value by 5*arStep.
// Spec: "Multiple EvMouse events with Button1 on the same arrow accumulate steps"
func TestScrollBarAutoRepeatHorizontalLeftArrowFiveClicks(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(1)

	// Simulate 5 consecutive Button1 events on left arrow (X=0)
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if sb.Value() != 45 {
		t.Errorf("after 5 clicks on left arrow: Value() = %d, want 45 (50 - 5*1)", sb.Value())
	}
}

// TestScrollBarAutoRepeatHorizontalRightArrowFiveClicks verifies that 5 consecutive
// Button1 events on the horizontal right arrow increase value by 5*arStep.
// Spec: "Multiple EvMouse events with Button1 on the same arrow accumulate steps"
func TestScrollBarAutoRepeatHorizontalRightArrowFiveClicks(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(1)

	// Simulate 5 consecutive Button1 events on right arrow (X=width-1=11)
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 11, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if sb.Value() != 55 {
		t.Errorf("after 5 clicks on right arrow: Value() = %d, want 55 (50 + 5*1)", sb.Value())
	}
}

// TestScrollBarAutoRepeatHorizontalLeftArrowTenClicks verifies that more
// consecutive events on horizontal left arrow accumulate correctly.
// Spec: "Multiple EvMouse events with Button1 on the same arrow accumulate steps"
func TestScrollBarAutoRepeatHorizontalLeftArrowTenClicks(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(60)
	sb.SetArStep(1)

	// Simulate 10 consecutive Button1 events on left arrow
	for i := 0; i < 10; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if sb.Value() != 50 {
		t.Errorf("after 10 clicks on left arrow: Value() = %d, want 50 (60 - 10*1)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Custom arStep values: steps are multiplied correctly
// ---------------------------------------------------------------------------

// TestScrollBarAutoRepeatVerticalUpArrowAlwaysStepByOne verifies that arrow clicks
// always step by 1, regardless of arStep setting (arStep only affects wheel events).
// This test verifies that 5 clicks always decrease by 5, not by 5*arStep.
// Spec: "Each EvMouse with Button1 on the up arrow decrements by arStep (which for arrows is always 1)"
func TestScrollBarAutoRepeatVerticalUpArrowAlwaysStepByOne(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(3) // arStep is 3, but arrow clicks ignore it

	// Simulate 5 consecutive Button1 events on up arrow
	// Arrow clicks step by 1, not by arStep
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if sb.Value() != 45 {
		t.Errorf("after 5 clicks on up arrow (arStep=3): Value() = %d, want 45 (50 - 5*1, arrows always step by 1)", sb.Value())
	}
}

// TestScrollBarAutoRepeatVerticalDownArrowAlwaysStepByOne verifies that down arrow
// clicks always step by 1, regardless of arStep setting.
// Spec: "Each EvMouse with Button1 on the down arrow increments by arStep (which for arrows is always 1)"
func TestScrollBarAutoRepeatVerticalDownArrowAlwaysStepByOne(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(3) // arStep is 3, but arrow clicks ignore it

	// Simulate 5 consecutive Button1 events on down arrow
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if sb.Value() != 55 {
		t.Errorf("after 5 clicks on down arrow (arStep=3): Value() = %d, want 55 (50 + 5*1, arrows always step by 1)", sb.Value())
	}
}

// TestScrollBarAutoRepeatHorizontalLeftArrowAlwaysStepByOne verifies that left arrow
// clicks always step by 1, regardless of arStep setting.
// Spec: "Same for horizontal left/right arrows"
func TestScrollBarAutoRepeatHorizontalLeftArrowAlwaysStepByOne(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(3) // arStep is 3, but arrow clicks ignore it

	// Simulate 5 consecutive Button1 events on left arrow
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if sb.Value() != 45 {
		t.Errorf("after 5 clicks on left arrow (arStep=3): Value() = %d, want 45 (50 - 5*1, arrows always step by 1)", sb.Value())
	}
}

// TestScrollBarAutoRepeatHorizontalRightArrowAlwaysStepByOne verifies that right arrow
// clicks always step by 1, regardless of arStep setting.
// Spec: "Same for horizontal left/right arrows"
func TestScrollBarAutoRepeatHorizontalRightArrowAlwaysStepByOne(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(3) // arStep is 3, but arrow clicks ignore it

	// Simulate 5 consecutive Button1 events on right arrow
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 11, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if sb.Value() != 55 {
		t.Errorf("after 5 clicks on right arrow (arStep=3): Value() = %d, want 55 (50 + 5*1, arrows always step by 1)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 4 — OnChange fires on each repeat
// ---------------------------------------------------------------------------

// TestScrollBarAutoRepeatVerticalUpArrowOnChangeFiresOnEachClick verifies that
// OnChange is called for each click, not just once at the end.
// Spec: "OnChange fires on each step"
func TestScrollBarAutoRepeatVerticalUpArrowOnChangeFiresOnEachClick(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(1)

	callCount := 0
	var receivedValues []int
	sb.OnChange = func(v int) {
		callCount++
		receivedValues = append(receivedValues, v)
	}

	// Simulate 5 consecutive Button1 events on up arrow
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if callCount != 5 {
		t.Errorf("OnChange called %d times, want 5 (once per click)", callCount)
	}

	// Verify the sequence of values: 49, 48, 47, 46, 45
	expectedValues := []int{49, 48, 47, 46, 45}
	for i, expected := range expectedValues {
		if i < len(receivedValues) && receivedValues[i] != expected {
			t.Errorf("OnChange call %d received %d, want %d", i+1, receivedValues[i], expected)
		}
	}
}

// TestScrollBarAutoRepeatVerticalDownArrowOnChangeFiresOnEachClick verifies that
// OnChange is called for each down arrow click.
// Spec: "OnChange fires on each step"
func TestScrollBarAutoRepeatVerticalDownArrowOnChangeFiresOnEachClick(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(1)

	callCount := 0
	var receivedValues []int
	sb.OnChange = func(v int) {
		callCount++
		receivedValues = append(receivedValues, v)
	}

	// Simulate 5 consecutive Button1 events on down arrow
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if callCount != 5 {
		t.Errorf("OnChange called %d times, want 5 (once per click)", callCount)
	}

	// Verify the sequence of values: 51, 52, 53, 54, 55
	expectedValues := []int{51, 52, 53, 54, 55}
	for i, expected := range expectedValues {
		if i < len(receivedValues) && receivedValues[i] != expected {
			t.Errorf("OnChange call %d received %d, want %d", i+1, receivedValues[i], expected)
		}
	}
}

// TestScrollBarAutoRepeatHorizontalLeftArrowOnChangeFiresOnEachClick verifies that
// OnChange is called for each left arrow click.
// Spec: "OnChange fires on each step"
func TestScrollBarAutoRepeatHorizontalLeftArrowOnChangeFiresOnEachClick(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(1)

	callCount := 0
	sb.OnChange = func(v int) {
		callCount++
	}

	// Simulate 5 consecutive Button1 events on left arrow
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if callCount != 5 {
		t.Errorf("OnChange called %d times, want 5 (once per click)", callCount)
	}
}

// TestScrollBarAutoRepeatHorizontalRightArrowOnChangeFiresOnEachClick verifies that
// OnChange is called for each right arrow click.
// Spec: "OnChange fires on each step"
func TestScrollBarAutoRepeatHorizontalRightArrowOnChangeFiresOnEachClick(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(1)

	callCount := 0
	sb.OnChange = func(v int) {
		callCount++
	}

	// Simulate 5 consecutive Button1 events on right arrow
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 11, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	if callCount != 5 {
		t.Errorf("OnChange called %d times, want 5 (once per click)", callCount)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Events are consumed
// ---------------------------------------------------------------------------

// TestScrollBarAutoRepeatEventsAreConsumed verifies that each EvMouse event
// with Button1 on an arrow is consumed (cleared).
// Spec: "Events are consumed"
func TestScrollBarAutoRepeatEventsAreConsumed(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	// Simulate 5 consecutive Button1 events on up arrow
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)

		if !ev.IsCleared() {
			t.Errorf("event %d: not cleared; ev.What = %v, want EvNothing", i+1, ev.What)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Falsifying: value changes incrementally (not all-at-once)
// ---------------------------------------------------------------------------

// TestScrollBarAutoRepeatValueChangesIncrementally verifies that the value changes
// incrementally with each click, not all-at-once at the end. This falsifies the
// hypothesis that all steps happen in a single transaction.
// Spec: "Falsifying: accumulation not all-at-once (value changes incrementally)"
func TestScrollBarAutoRepeatValueChangesIncrementally(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(1)

	var valueSequence []int
	sb.OnChange = func(v int) {
		valueSequence = append(valueSequence, v)
	}

	// Simulate 5 consecutive Button1 events on up arrow
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	// If value changed all-at-once, we'd see only one callback with value 45.
	// If it changed incrementally, we'd see 5 callbacks with values 49, 48, 47, 46, 45.

	if len(valueSequence) != 5 {
		t.Errorf("value changed %d times, want 5 (incremental changes); if only 1, then all-at-once", len(valueSequence))
	}

	// Verify they are sequential, not all the same
	if len(valueSequence) > 1 {
		for i := 1; i < len(valueSequence); i++ {
			if valueSequence[i] == valueSequence[i-1] {
				t.Errorf("value at callback %d is %d (same as previous); expected incremental changes",
					i+1, valueSequence[i])
			}
		}
	}

	// Verify final value is 45, not something else
	if sb.Value() != 45 {
		t.Errorf("final Value() = %d, want 45", sb.Value())
	}
}

// TestScrollBarAutoRepeatValueChangesIncrementallyArrowSteps verifies incremental
// changes with multiple arrow clicks. Arrow clicks always step by 1 (not arStep).
// Each click should change value by 1, and changes should be incremental.
// Spec: "Falsifying: accumulation not all-at-once (value changes incrementally)"
func TestScrollBarAutoRepeatValueChangesIncrementallyArrowSteps(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(3) // arStep does not affect arrow clicks

	var valueSequence []int
	sb.OnChange = func(v int) {
		valueSequence = append(valueSequence, v)
	}

	// Simulate 5 consecutive Button1 events on up arrow
	// Arrow clicks always step by 1, even with arStep=3
	for i := 0; i < 5; i++ {
		ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
		sb.HandleEvent(ev)
	}

	// Each click should decrease by 1: 49, 48, 47, 46, 45
	expectedSequence := []int{49, 48, 47, 46, 45}

	if len(valueSequence) != 5 {
		t.Errorf("OnChange called %d times, want 5 (incremental)", len(valueSequence))
		return
	}

	for i, expected := range expectedSequence {
		if valueSequence[i] != expected {
			t.Errorf("callback %d: value = %d, want %d (each arrow click should decrease by 1)",
				i+1, valueSequence[i], expected)
		}
	}
}
