package tv

// scrollbar_wheel_test.go — Tests for Task 2: ScrollBar Mouse Wheel Speed (spec 6.4).
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — WheelUp with default arStep=1 decreases by 3 (not 1)
//   Section 2  — WheelDown with default arStep=1 increases by 3 (not 1)
//   Section 3  — WheelUp with custom arStep multiplies by 3
//   Section 4  — WheelDown with custom arStep multiplies by 3
//   Section 5  — Wheel events clamp at boundaries
//   Section 6  — OnChange callback fires with correct value
//   Section 7  — Wheel events are consumed (IsCleared)
//   Section 8  — Works on both vertical and horizontal scrollbars
//   Section 9  — Falsifying: WheelUp does NOT scroll by 1 (catches no-change bug)
//   Section 10 — Falsifying: WheelDown does NOT scroll by 1 (catches no-change bug)

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Section 1 — WheelUp with default arStep=1 decreases by 3 (not 1)
// ---------------------------------------------------------------------------

// TestScrollBarWheelUpDefaultArStepDecreasesByThree verifies WheelUp scrolls by
// 3 * arStep (3 with default arStep=1), not by 1.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
// Spec: "WheelUp: step(-3 * arStep)"
// Spec: "arStep defaults to 1, so default wheel speed is 3 per tick"
func TestScrollBarWheelUpDefaultArStepDecreasesByThree(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 47 {
		t.Errorf("WheelUp (arStep=1): Value() = %d, want 47 (50 - 3*1)", sb.Value())
	}
}

// TestScrollBarWheelUpDefaultArStepVerticalIsIndependent verifies the behavior
// is not an artifact of bounds or pageSize; different ranges work the same.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelUpDefaultArStepDifferentRange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(10, 110)
	sb.SetPageSize(10)
	sb.SetValue(60)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 57 {
		t.Errorf("WheelUp with range [10,110] (arStep=1): Value() = %d, want 57 (60 - 3*1)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 2 — WheelDown with default arStep=1 increases by 3 (not 1)
// ---------------------------------------------------------------------------

// TestScrollBarWheelDownDefaultArStepIncreasesByThree verifies WheelDown scrolls by
// 3 * arStep (3 with default arStep=1), not by 1.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
// Spec: "WheelDown: step(3 * arStep)"
// Spec: "arStep defaults to 1, so default wheel speed is 3 per tick"
func TestScrollBarWheelDownDefaultArStepIncreasesByThree(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 53 {
		t.Errorf("WheelDown (arStep=1): Value() = %d, want 53 (50 + 3*1)", sb.Value())
	}
}

// TestScrollBarWheelDownDefaultArStepDifferentRange verifies the behavior
// is consistent across different ranges.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelDownDefaultArStepDifferentRange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(10, 110)
	sb.SetPageSize(10)
	sb.SetValue(60)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 63 {
		t.Errorf("WheelDown with range [10,110] (arStep=1): Value() = %d, want 63 (60 + 3*1)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — WheelUp with custom arStep multiplies by 3
// ---------------------------------------------------------------------------

// TestScrollBarWheelUpCustomArStepDecreasesByThreeTimesArStep verifies WheelUp
// uses 3 * arStep, not just arStep.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
// Spec: "WheelUp: step(-3 * arStep)"
// Spec: "When arStep is changed via SetArStep, wheel speed changes proportionally"
func TestScrollBarWheelUpCustomArStepDecreasesByThreeTimesArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(2)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 44 {
		t.Errorf("WheelUp (arStep=2): Value() = %d, want 44 (50 - 3*2)", sb.Value())
	}
}

// TestScrollBarWheelUpLargeArStepDecreasesByThreeTimesArStep verifies the
// multiplication applies to larger arStep values as well.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelUpLargeArStepDecreasesByThreeTimesArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 200)
	sb.SetPageSize(20)
	sb.SetValue(150)
	sb.SetArStep(5)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 135 {
		t.Errorf("WheelUp (arStep=5): Value() = %d, want 135 (150 - 3*5)", sb.Value())
	}
}

// TestScrollBarWheelUpArStepChangedAfterCreation verifies that changing arStep
// after creation affects subsequent wheel events correctly.
// Spec: "When arStep is changed via SetArStep, wheel speed changes proportionally"
func TestScrollBarWheelUpArStepChangedAfterCreation(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	// First wheel event with default arStep=1
	ev1 := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev1)

	if sb.Value() != 47 {
		t.Errorf("First WheelUp (arStep=1): Value() = %d, want 47", sb.Value())
	}

	// Change arStep and emit another wheel event
	sb.SetArStep(3)
	ev2 := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev2)

	if sb.Value() != 38 {
		t.Errorf("Second WheelUp (arStep=3): Value() = %d, want 38 (47 - 3*3)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 4 — WheelDown with custom arStep multiplies by 3
// ---------------------------------------------------------------------------

// TestScrollBarWheelDownCustomArStepIncreasesByThreeTimesArStep verifies WheelDown
// uses 3 * arStep, not just arStep.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
// Spec: "WheelDown: step(3 * arStep)"
// Spec: "When arStep is changed via SetArStep, wheel speed changes proportionally"
func TestScrollBarWheelDownCustomArStepIncreasesByThreeTimesArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(2)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 56 {
		t.Errorf("WheelDown (arStep=2): Value() = %d, want 56 (50 + 3*2)", sb.Value())
	}
}

// TestScrollBarWheelDownLargeArStepIncreasesByThreeTimesArStep verifies the
// multiplication applies to larger arStep values as well.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelDownLargeArStepIncreasesByThreeTimesArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 200)
	sb.SetPageSize(20)
	sb.SetValue(150)
	sb.SetArStep(5)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 165 {
		t.Errorf("WheelDown (arStep=5): Value() = %d, want 165 (150 + 3*5)", sb.Value())
	}
}

// TestScrollBarWheelDownArStepChangedAfterCreation verifies that changing arStep
// after creation affects subsequent wheel events correctly.
// Spec: "When arStep is changed via SetArStep, wheel speed changes proportionally"
func TestScrollBarWheelDownArStepChangedAfterCreation(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	// First wheel event with default arStep=1
	ev1 := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev1)

	if sb.Value() != 53 {
		t.Errorf("First WheelDown (arStep=1): Value() = %d, want 53", sb.Value())
	}

	// Change arStep and emit another wheel event
	sb.SetArStep(3)
	ev2 := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev2)

	if sb.Value() != 62 {
		t.Errorf("Second WheelDown (arStep=3): Value() = %d, want 62 (53 + 3*3)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Wheel events clamp at boundaries
// ---------------------------------------------------------------------------

// TestScrollBarWheelUpClampsAtMin verifies WheelUp does not scroll below min.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick (matching original TV)"
func TestScrollBarWheelUpClampsAtMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(2) // Only 2 above min, but WheelUp tries to move by 3
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 0 {
		t.Errorf("WheelUp clamped at min: Value() = %d, want 0 (clamped)", sb.Value())
	}
}

// TestScrollBarWheelDownClampsAtMax verifies WheelDown does not scroll above max-pageSize.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick (matching original TV)"
func TestScrollBarWheelDownClampsAtMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(88) // Only 2 below max-pageSize (90), but WheelDown tries to move by 3
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 90 {
		t.Errorf("WheelDown clamped at max: Value() = %d, want 90 (max-pageSize, clamped)", sb.Value())
	}
}

// TestScrollBarWheelUpClampsAtMinWithLargeArStep verifies clamping works with larger arSteps.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelUpClampsAtMinWithLargeArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(5)
	sb.SetArStep(3) // Would move by 9, but clamped at 0
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 0 {
		t.Errorf("WheelUp (arStep=3) clamped at min: Value() = %d, want 0", sb.Value())
	}
}

// TestScrollBarWheelDownClampsAtMaxWithLargeArStep verifies clamping works with larger arSteps.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelDownClampsAtMaxWithLargeArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(85)
	sb.SetArStep(3) // Would move by 9, reaching 94, but clamped at max-pageSize=90
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 90 {
		t.Errorf("WheelDown (arStep=3) clamped at max: Value() = %d, want 90", sb.Value())
	}
}

// TestScrollBarWheelUpAtExactMin verifies WheelUp at min does nothing.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelUpAtExactMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0) // Already at min
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 0 {
		t.Errorf("WheelUp at min: Value() = %d, want 0", sb.Value())
	}
}

// TestScrollBarWheelDownAtExactMax verifies WheelDown at max does nothing.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelDownAtExactMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90) // Already at max-pageSize
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 90 {
		t.Errorf("WheelDown at max: Value() = %d, want 90", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 6 — OnChange callback fires with correct value
// ---------------------------------------------------------------------------

// TestScrollBarWheelUpFiresOnChange verifies OnChange is called when WheelUp changes value.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelUpFiresOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	changedValue := -1
	sb.OnChange = func(v int) {
		changedValue = v
	}

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if changedValue != 47 {
		t.Errorf("WheelUp OnChange: callback received %d, want 47", changedValue)
	}
}

// TestScrollBarWheelDownFiresOnChange verifies OnChange is called when WheelDown changes value.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelDownFiresOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	changedValue := -1
	sb.OnChange = func(v int) {
		changedValue = v
	}

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if changedValue != 53 {
		t.Errorf("WheelDown OnChange: callback received %d, want 53", changedValue)
	}
}

// TestScrollBarWheelUpDoesNotFireOnChangeIfClamped verifies OnChange is only called
// if the value actually changes.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelUpDoesNotFireOnChangeIfClamped(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0) // Already at min
	sb.SetState(SfSelected, true)

	callCount := 0
	sb.OnChange = func(v int) {
		callCount++
	}

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if callCount != 0 {
		t.Errorf("WheelUp at min: OnChange called %d times, want 0 (no change)", callCount)
	}
}

// TestScrollBarWheelDownDoesNotFireOnChangeIfClamped verifies OnChange is only called
// if the value actually changes.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelDownDoesNotFireOnChangeIfClamped(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90) // Already at max-pageSize
	sb.SetState(SfSelected, true)

	callCount := 0
	sb.OnChange = func(v int) {
		callCount++
	}

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if callCount != 0 {
		t.Errorf("WheelDown at max: OnChange called %d times, want 0 (no change)", callCount)
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Wheel events are consumed (IsCleared)
// ---------------------------------------------------------------------------

// TestScrollBarWheelUpIsConsumed verifies WheelUp events are marked as consumed.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelUpIsConsumed(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("WheelUp event should be cleared (consumed) after handling")
	}
}

// TestScrollBarWheelDownIsConsumed verifies WheelDown events are marked as consumed.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelDownIsConsumed(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("WheelDown event should be cleared (consumed) after handling")
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Works on both vertical and horizontal scrollbars
// ---------------------------------------------------------------------------

// TestScrollBarWheelUpVertical verifies WheelUp works on vertical scrollbars.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick (matching original TV)"
func TestScrollBarWheelUpVertical(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 47 {
		t.Errorf("Vertical WheelUp: Value() = %d, want 47", sb.Value())
	}
}

// TestScrollBarWheelDownVertical verifies WheelDown works on vertical scrollbars.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick (matching original TV)"
func TestScrollBarWheelDownVertical(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 53 {
		t.Errorf("Vertical WheelDown: Value() = %d, want 53", sb.Value())
	}
}

// TestScrollBarWheelUpHorizontal verifies WheelUp works on horizontal scrollbars.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick (matching original TV)"
func TestScrollBarWheelUpHorizontal(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 47 {
		t.Errorf("Horizontal WheelUp: Value() = %d, want 47", sb.Value())
	}
}

// TestScrollBarWheelDownHorizontal verifies WheelDown works on horizontal scrollbars.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick (matching original TV)"
func TestScrollBarWheelDownHorizontal(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 53 {
		t.Errorf("Horizontal WheelDown: Value() = %d, want 53", sb.Value())
	}
}

// TestScrollBarWheelHorizontalWithCustomArStep verifies arStep multiplication
// works on horizontal scrollbars too.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelHorizontalWithCustomArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(4)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 62 {
		t.Errorf("Horizontal WheelDown (arStep=4): Value() = %d, want 62 (50 + 3*4)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Falsifying: WheelUp does NOT scroll by 1 (catches no-change bug)
// ---------------------------------------------------------------------------

// TestScrollBarWheelUpNotScrollByOne verifies WheelUp scrolls by 3, not by 1.
// This falsifying test catches bugs where implementation incorrectly uses
// step(-1) instead of step(-3*arStep).
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick, not by 1"
func TestScrollBarWheelUpNotScrollByOne(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() == 49 {
		t.Error("WheelUp scrolled by 1 (value=49), but should scroll by 3 (want 47)")
	}
}

// TestScrollBarWheelUpNotScrollByOneWithArStep verifies WheelUp scrolls by 3*arStep,
// not by arStep alone. This catches bugs where arStep is used without the 3x multiplier.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick, not by 1"
func TestScrollBarWheelUpNotScrollByOneWithArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(2)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	// Should be 44 (50 - 3*2), not 48 (50 - 2)
	if sb.Value() == 48 {
		t.Error("WheelUp scrolled by arStep only (value=48), but should scroll by 3*arStep (want 44)")
	}
}

// ---------------------------------------------------------------------------
// Section 10 — Falsifying: WheelDown does NOT scroll by 1 (catches no-change bug)
// ---------------------------------------------------------------------------

// TestScrollBarWheelDownNotScrollByOne verifies WheelDown scrolls by 3, not by 1.
// This falsifying test catches bugs where implementation incorrectly uses
// step(1) instead of step(3*arStep).
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick, not by 1"
func TestScrollBarWheelDownNotScrollByOne(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() == 51 {
		t.Error("WheelDown scrolled by 1 (value=51), but should scroll by 3 (want 53)")
	}
}

// TestScrollBarWheelDownNotScrollByOneWithArStep verifies WheelDown scrolls by 3*arStep,
// not by arStep alone. This catches bugs where arStep is used without the 3x multiplier.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick, not by 1"
func TestScrollBarWheelDownNotScrollByOneWithArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(2)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	// Should be 56 (50 + 3*2), not 52 (50 + 2)
	if sb.Value() == 52 {
		t.Error("WheelDown scrolled by arStep only (value=52), but should scroll by 3*arStep (want 56)")
	}
}
