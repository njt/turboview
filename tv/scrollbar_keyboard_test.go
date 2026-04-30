package tv

// scrollbar_keyboard_test.go — Tests for Task 1: ScrollBar Keyboard Handling (spec 6.1).
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — OfSelectable set by NewScrollBar
//   Section 2  — arStep field: default, SetArStep, ArStep
//   Section 3  — Vertical scrollbar: Up/Down arrow keys
//   Section 4  — Vertical scrollbar: Page Up/Down
//   Section 5  — Vertical scrollbar: Ctrl+Page Up/Down (goto min/max)
//   Section 6  — Horizontal scrollbar: Left/Right arrow keys
//   Section 7  — Horizontal scrollbar: Ctrl+Left/Right (page step)
//   Section 8  — Horizontal scrollbar: Home/End (goto min/max)
//   Section 9  — Event consumed after handling
//   Section 10 — Keyboard ignored when not SfSelected
//   Section 11 — OnChange callback
//   Section 12 — Falsifying tests

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Section 1 — OfSelectable set by NewScrollBar
// ---------------------------------------------------------------------------

// TestNewScrollBarSetsOfSelectable verifies NewScrollBar sets OfSelectable so the
// scrollbar can receive focus.
// Spec: "SetOptions(OfSelectable, true) in NewScrollBar so scrollbar can receive focus"
func TestNewScrollBarSetsOfSelectable(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)

	if !sb.HasOption(OfSelectable) {
		t.Error("NewScrollBar must set OfSelectable so the scrollbar can receive focus")
	}
}

// TestNewScrollBarSetsOfSelectableHorizontal verifies the same for horizontal orientation.
// Spec: "SetOptions(OfSelectable, true) in NewScrollBar so scrollbar can receive focus"
func TestNewScrollBarSetsOfSelectableHorizontal(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)

	if !sb.HasOption(OfSelectable) {
		t.Error("NewScrollBar (horizontal) must set OfSelectable so the scrollbar can receive focus")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — arStep field: default value, SetArStep, ArStep
// ---------------------------------------------------------------------------

// TestScrollBarArStepDefaultIsOne verifies a new ScrollBar has arStep == 1.
// Spec: "Add arStep int field (default arStep=1)"
func TestScrollBarArStepDefaultIsOne(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)

	if sb.ArStep() != 1 {
		t.Errorf("ArStep() = %d, want 1 (default)", sb.ArStep())
	}
}

// TestScrollBarSetArStepStoresValue verifies SetArStep stores the given step size.
// Spec: "Add setter: SetArStep(n)"
func TestScrollBarSetArStepStoresValue(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetArStep(5)

	if sb.ArStep() != 5 {
		t.Errorf("ArStep() = %d, want 5 after SetArStep(5)", sb.ArStep())
	}
}

// TestScrollBarSetArStepToOne verifies SetArStep(1) is accepted (minimum useful step).
// Spec: "Add setter: SetArStep(n)"
func TestScrollBarSetArStepToOne(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetArStep(10)
	sb.SetArStep(1)

	if sb.ArStep() != 1 {
		t.Errorf("ArStep() = %d, want 1 after SetArStep(1)", sb.ArStep())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Vertical scrollbar: Up arrow / Down arrow
// ---------------------------------------------------------------------------

// TestScrollBarVerticalUpArrowDecreasesByArStep verifies Up arrow decreases value by arStep.
// Spec: "Up arrow: step(-arStep) (decrease value)"
func TestScrollBarVerticalUpArrowDecreasesByArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 49 {
		t.Errorf("Up arrow (arStep=1): Value() = %d, want 49", sb.Value())
	}
}

// TestScrollBarVerticalUpArrowDecreasesByCustomArStep verifies Up arrow uses the
// configured arStep, not a hard-coded 1.
// Spec: "Up arrow: step(-arStep) (decrease value)"
func TestScrollBarVerticalUpArrowDecreasesByCustomArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(3)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 47 {
		t.Errorf("Up arrow (arStep=3): Value() = %d, want 47 (50 - 3)", sb.Value())
	}
}

// TestScrollBarVerticalUpArrowClampsAtMin verifies Up arrow does not go below min.
// Spec: "Up arrow: step(-arStep) (decrease value)"
func TestScrollBarVerticalUpArrowClampsAtMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0) // already at min
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 0 {
		t.Errorf("Up arrow at min: Value() = %d, want 0 (clamped at min)", sb.Value())
	}
}

// TestScrollBarVerticalDownArrowIncreasesByArStep verifies Down arrow increases value by arStep.
// Spec: "Down arrow: step(arStep) (increase value)"
func TestScrollBarVerticalDownArrowIncreasesByArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 51 {
		t.Errorf("Down arrow (arStep=1): Value() = %d, want 51", sb.Value())
	}
}

// TestScrollBarVerticalDownArrowIncreasesByCustomArStep verifies Down arrow uses the
// configured arStep.
// Spec: "Down arrow: step(arStep) (increase value)"
func TestScrollBarVerticalDownArrowIncreasesByCustomArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(4)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 54 {
		t.Errorf("Down arrow (arStep=4): Value() = %d, want 54 (50 + 4)", sb.Value())
	}
}

// TestScrollBarVerticalDownArrowClampsAtMax verifies Down arrow does not exceed max-pageSize.
// Spec: "Down arrow: step(arStep) (increase value)"
func TestScrollBarVerticalDownArrowClampsAtMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90) // already at max-pageSize
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 90 {
		t.Errorf("Down arrow at max: Value() = %d, want 90 (clamped at max-pageSize)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Vertical scrollbar: Page Up / Page Down
// ---------------------------------------------------------------------------

// TestScrollBarVerticalPageUpDecreasesByPageSize verifies Page Up decreases value by pageSize.
// Spec: "Page Up: page(-1) (decrease by pageSize, using existing page method)"
func TestScrollBarVerticalPageUpDecreasesByPageSize(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(15)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 35 {
		t.Errorf("Page Up (pageSize=15): Value() = %d, want 35 (50 - 15)", sb.Value())
	}
}

// TestScrollBarVerticalPageUpClampsAtMin verifies Page Up does not go below min.
// Spec: "Page Up: page(-1) (decrease by pageSize, using existing page method)"
func TestScrollBarVerticalPageUpClampsAtMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(20)
	sb.SetValue(10) // 10 - 20 would be below min
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 0 {
		t.Errorf("Page Up clamped: Value() = %d, want 0 (clamped at min)", sb.Value())
	}
}

// TestScrollBarVerticalPageDownIncreasesByPageSize verifies Page Down increases value by pageSize.
// Spec: "Page Down: page(1) (increase by pageSize)"
func TestScrollBarVerticalPageDownIncreasesByPageSize(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(15)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgDn}}
	sb.HandleEvent(ev)

	if sb.Value() != 65 {
		t.Errorf("Page Down (pageSize=15): Value() = %d, want 65 (50 + 15)", sb.Value())
	}
}

// TestScrollBarVerticalPageDownClampsAtMax verifies Page Down does not exceed max-pageSize.
// Spec: "Page Down: page(1) (increase by pageSize)"
func TestScrollBarVerticalPageDownClampsAtMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(20)
	sb.SetValue(75) // 75 + 20 = 95 > max-pageSize=80
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgDn}}
	sb.HandleEvent(ev)

	if sb.Value() != 80 {
		t.Errorf("Page Down clamped: Value() = %d, want 80 (clamped at max-pageSize = 100-20)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Vertical scrollbar: Ctrl+Page Up / Ctrl+Page Down (goto min/max)
// ---------------------------------------------------------------------------

// TestScrollBarVerticalCtrlPageUpGoesToMin verifies Ctrl+Page Up sets value to min.
// Spec: "Ctrl+Page Up: go to minVal"
func TestScrollBarVerticalCtrlPageUpGoesToMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(10, 100)
	sb.SetPageSize(10)
	sb.SetValue(75)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp, Modifiers: tcell.ModCtrl}}
	sb.HandleEvent(ev)

	if sb.Value() != 10 {
		t.Errorf("Ctrl+Page Up: Value() = %d, want 10 (min)", sb.Value())
	}
}

// TestScrollBarVerticalCtrlPageUpGoesToZeroMin verifies Ctrl+Page Up uses actual min,
// not assuming zero.
// Spec: "Ctrl+Page Up: go to minVal"
func TestScrollBarVerticalCtrlPageUpGoesToZeroMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 50)
	sb.SetPageSize(5)
	sb.SetValue(30)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp, Modifiers: tcell.ModCtrl}}
	sb.HandleEvent(ev)

	if sb.Value() != 0 {
		t.Errorf("Ctrl+Page Up: Value() = %d, want 0 (min=0)", sb.Value())
	}
}

// TestScrollBarVerticalCtrlPageDownGoesToMax verifies Ctrl+Page Down sets value to max-pageSize.
// Spec: "Ctrl+Page Down: go to maxVal"
func TestScrollBarVerticalCtrlPageDownGoesToMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(20)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgDn, Modifiers: tcell.ModCtrl}}
	sb.HandleEvent(ev)

	// "go to maxVal" — maxVal for the scrollbar position is max-pageSize = 90
	if sb.Value() != 90 {
		t.Errorf("Ctrl+Page Down: Value() = %d, want 90 (max-pageSize = 100-10)", sb.Value())
	}
}

// TestScrollBarVerticalCtrlPageDownGoesToMaxWithNonZeroMin verifies Ctrl+Page Down
// uses actual max, not assuming 100.
// Spec: "Ctrl+Page Down: go to maxVal"
func TestScrollBarVerticalCtrlPageDownGoesToMaxWithNonZeroMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(5, 55)
	sb.SetPageSize(10)
	sb.SetValue(15)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgDn, Modifiers: tcell.ModCtrl}}
	sb.HandleEvent(ev)

	// max-pageSize = 55-10 = 45
	if sb.Value() != 45 {
		t.Errorf("Ctrl+Page Down: Value() = %d, want 45 (max-pageSize = 55-10)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Horizontal scrollbar: Left / Right arrow keys
// ---------------------------------------------------------------------------

// TestScrollBarHorizontalLeftArrowDecreasesByArStep verifies Left arrow decreases value by arStep.
// Spec: "Left arrow: decrease by arStep"
func TestScrollBarHorizontalLeftArrowDecreasesByArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	sb.HandleEvent(ev)

	if sb.Value() != 49 {
		t.Errorf("Left arrow (arStep=1): Value() = %d, want 49", sb.Value())
	}
}

// TestScrollBarHorizontalLeftArrowDecreasesByCustomArStep verifies Left arrow uses
// the configured arStep.
// Spec: "Left arrow: decrease by arStep"
func TestScrollBarHorizontalLeftArrowDecreasesByCustomArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(7)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	sb.HandleEvent(ev)

	if sb.Value() != 43 {
		t.Errorf("Left arrow (arStep=7): Value() = %d, want 43 (50 - 7)", sb.Value())
	}
}

// TestScrollBarHorizontalLeftArrowClampsAtMin verifies Left arrow does not go below min.
// Spec: "Left arrow: decrease by arStep"
func TestScrollBarHorizontalLeftArrowClampsAtMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	sb.HandleEvent(ev)

	if sb.Value() != 0 {
		t.Errorf("Left arrow at min: Value() = %d, want 0 (clamped)", sb.Value())
	}
}

// TestScrollBarHorizontalRightArrowIncreasesByArStep verifies Right arrow increases value by arStep.
// Spec: "Right arrow: increase by arStep"
func TestScrollBarHorizontalRightArrowIncreasesByArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	sb.HandleEvent(ev)

	if sb.Value() != 51 {
		t.Errorf("Right arrow (arStep=1): Value() = %d, want 51", sb.Value())
	}
}

// TestScrollBarHorizontalRightArrowIncreasesByCustomArStep verifies Right arrow uses
// the configured arStep.
// Spec: "Right arrow: increase by arStep"
func TestScrollBarHorizontalRightArrowIncreasesByCustomArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(6)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	sb.HandleEvent(ev)

	if sb.Value() != 56 {
		t.Errorf("Right arrow (arStep=6): Value() = %d, want 56 (50 + 6)", sb.Value())
	}
}

// TestScrollBarHorizontalRightArrowClampsAtMax verifies Right arrow does not exceed max-pageSize.
// Spec: "Right arrow: increase by arStep"
func TestScrollBarHorizontalRightArrowClampsAtMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	sb.HandleEvent(ev)

	if sb.Value() != 90 {
		t.Errorf("Right arrow at max: Value() = %d, want 90 (clamped at max-pageSize)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Horizontal scrollbar: Ctrl+Left / Ctrl+Right (page step)
// ---------------------------------------------------------------------------

// TestScrollBarHorizontalCtrlLeftDecreasesByPageSize verifies Ctrl+Left decreases value by pageSize.
// Spec: "Ctrl+Left: page(-1) (decrease by pageSize)"
func TestScrollBarHorizontalCtrlLeftDecreasesByPageSize(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(20)
	sb.SetValue(60)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}}
	sb.HandleEvent(ev)

	if sb.Value() != 40 {
		t.Errorf("Ctrl+Left (pageSize=20): Value() = %d, want 40 (60 - 20)", sb.Value())
	}
}

// TestScrollBarHorizontalCtrlLeftClampsAtMin verifies Ctrl+Left does not go below min.
// Spec: "Ctrl+Left: page(-1) (decrease by pageSize)"
func TestScrollBarHorizontalCtrlLeftClampsAtMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(30)
	sb.SetValue(15) // 15 - 30 < 0
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}}
	sb.HandleEvent(ev)

	if sb.Value() != 0 {
		t.Errorf("Ctrl+Left clamped: Value() = %d, want 0 (clamped at min)", sb.Value())
	}
}

// TestScrollBarHorizontalCtrlRightIncreasesByPageSize verifies Ctrl+Right increases value by pageSize.
// Spec: "Ctrl+Right: page(1) (increase by pageSize)"
func TestScrollBarHorizontalCtrlRightIncreasesByPageSize(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(20)
	sb.SetValue(40)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}}
	sb.HandleEvent(ev)

	if sb.Value() != 60 {
		t.Errorf("Ctrl+Right (pageSize=20): Value() = %d, want 60 (40 + 20)", sb.Value())
	}
}

// TestScrollBarHorizontalCtrlRightClampsAtMax verifies Ctrl+Right does not exceed max-pageSize.
// Spec: "Ctrl+Right: page(1) (increase by pageSize)"
func TestScrollBarHorizontalCtrlRightClampsAtMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(20)
	sb.SetValue(75) // 75 + 20 = 95 > max-pageSize=80
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}}
	sb.HandleEvent(ev)

	if sb.Value() != 80 {
		t.Errorf("Ctrl+Right clamped: Value() = %d, want 80 (max-pageSize = 100-20)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Horizontal scrollbar: Home / End (goto min/max)
// ---------------------------------------------------------------------------

// TestScrollBarHorizontalHomeGoesToMin verifies Home sets value to min.
// Spec: "Home: go to minVal"
func TestScrollBarHorizontalHomeGoesToMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(5, 100)
	sb.SetPageSize(10)
	sb.SetValue(70)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}}
	sb.HandleEvent(ev)

	if sb.Value() != 5 {
		t.Errorf("Home: Value() = %d, want 5 (min)", sb.Value())
	}
}

// TestScrollBarHorizontalHomeGoesToZeroMin verifies Home uses the actual min value.
// Spec: "Home: go to minVal"
func TestScrollBarHorizontalHomeGoesToZeroMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 80)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}}
	sb.HandleEvent(ev)

	if sb.Value() != 0 {
		t.Errorf("Home: Value() = %d, want 0 (min=0)", sb.Value())
	}
}

// TestScrollBarHorizontalEndGoesToMax verifies End sets value to max-pageSize.
// Spec: "End: go to maxVal"
func TestScrollBarHorizontalEndGoesToMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(20)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd}}
	sb.HandleEvent(ev)

	// maxVal for position is max-pageSize = 90
	if sb.Value() != 90 {
		t.Errorf("End: Value() = %d, want 90 (max-pageSize = 100-10)", sb.Value())
	}
}

// TestScrollBarHorizontalEndGoesToMaxWithNonZeroMin verifies End uses actual max.
// Spec: "End: go to maxVal"
func TestScrollBarHorizontalEndGoesToMaxWithNonZeroMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(10, 60)
	sb.SetPageSize(5)
	sb.SetValue(20)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd}}
	sb.HandleEvent(ev)

	// max-pageSize = 60-5 = 55
	if sb.Value() != 55 {
		t.Errorf("End: Value() = %d, want 55 (max-pageSize = 60-5)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Event consumed after handling
// ---------------------------------------------------------------------------

// TestScrollBarKeyboardEventClearedVerticalKeys verifies each handled vertical
// keyboard event is consumed (IsCleared returns true).
// Spec: "Event is cleared after handling"
func TestScrollBarKeyboardEventClearedVerticalKeys(t *testing.T) {
	keys := []struct {
		name string
		key  tcell.Key
		mod  tcell.ModMask
	}{
		{"Up", tcell.KeyUp, 0},
		{"Down", tcell.KeyDown, 0},
		{"PgUp", tcell.KeyPgUp, 0},
		{"PgDn", tcell.KeyPgDn, 0},
		{"Ctrl+PgUp", tcell.KeyPgUp, tcell.ModCtrl},
		{"Ctrl+PgDn", tcell.KeyPgDn, tcell.ModCtrl},
	}

	for _, tt := range keys {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
			sb.SetRange(0, 100)
			sb.SetPageSize(10)
			sb.SetValue(50)
			sb.SetState(SfSelected, true)

			ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tt.key, Modifiers: tt.mod}}
			sb.HandleEvent(ev)

			if !ev.IsCleared() {
				t.Errorf("%s: event not cleared after handling (IsCleared=false)", tt.name)
			}
		})
	}
}

// TestScrollBarKeyboardEventClearedHorizontalKeys verifies each handled horizontal
// keyboard event is consumed.
// Spec: "Event is cleared after handling"
func TestScrollBarKeyboardEventClearedHorizontalKeys(t *testing.T) {
	keys := []struct {
		name string
		key  tcell.Key
		mod  tcell.ModMask
	}{
		{"Left", tcell.KeyLeft, 0},
		{"Right", tcell.KeyRight, 0},
		{"Ctrl+Left", tcell.KeyLeft, tcell.ModCtrl},
		{"Ctrl+Right", tcell.KeyRight, tcell.ModCtrl},
		{"Home", tcell.KeyHome, 0},
		{"End", tcell.KeyEnd, 0},
	}

	for _, tt := range keys {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
			sb.SetRange(0, 100)
			sb.SetPageSize(10)
			sb.SetValue(50)
			sb.SetState(SfSelected, true)

			ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tt.key, Modifiers: tt.mod}}
			sb.HandleEvent(ev)

			if !ev.IsCleared() {
				t.Errorf("%s: event not cleared after handling (IsCleared=false)", tt.name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Section 10 — Keyboard ignored when not SfSelected
// ---------------------------------------------------------------------------

// TestScrollBarKeyboardIgnoredWhenNotSelected verifies that keyboard events are
// NOT handled when the scrollbar does not have SfSelected state.
// Spec: "ScrollBar.HandleEvent handles EvKeyboard events when the scrollbar has SfSelected state"
func TestScrollBarVerticalKeyboardIgnoredWhenNotSelected(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	// Do NOT set SfSelected

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 50 {
		t.Errorf("Up arrow without SfSelected: Value() = %d, want 50 (unchanged)", sb.Value())
	}
}

// TestScrollBarVerticalKeyboardEventNotClearedWhenNotSelected verifies that the event
// is NOT consumed when the scrollbar does not have SfSelected.
// Spec: "ScrollBar.HandleEvent handles EvKeyboard events when the scrollbar has SfSelected state"
func TestScrollBarVerticalKeyboardEventNotClearedWhenNotSelected(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	// Do NOT set SfSelected

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	sb.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("event was cleared without SfSelected; keyboard should be ignored when not selected")
	}
}

// TestScrollBarHorizontalKeyboardIgnoredWhenNotSelected verifies horizontal scrollbar
// also ignores keys without SfSelected.
// Spec: "ScrollBar.HandleEvent handles EvKeyboard events when the scrollbar has SfSelected state"
func TestScrollBarHorizontalKeyboardIgnoredWhenNotSelected(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	// Do NOT set SfSelected

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	sb.HandleEvent(ev)

	if sb.Value() != 50 {
		t.Errorf("Right arrow without SfSelected: Value() = %d, want 50 (unchanged)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 11 — OnChange callback fires on keyboard-driven value changes
// ---------------------------------------------------------------------------

// TestScrollBarKeyboardUpArrowCallsOnChange verifies OnChange is called when Up arrow
// changes the value.
// Spec: "OnChange func(int) - callback when value changes" (implied by all value-changing operations)
func TestScrollBarKeyboardUpArrowCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	called := false
	sb.OnChange = func(v int) { called = true }

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called after Up arrow key")
	}
}

// TestScrollBarKeyboardUpArrowPassesNewValueToOnChange verifies OnChange receives
// the post-change value.
// Spec: "OnChange func(int) - callback when value changes"
func TestScrollBarKeyboardUpArrowPassesNewValueToOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	var got int
	sb.OnChange = func(v int) { got = v }

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	sb.HandleEvent(ev)

	if got != 49 {
		t.Errorf("OnChange received %d, want 49", got)
	}
}

// TestScrollBarKeyboardDownArrowCallsOnChange verifies OnChange is called when Down
// arrow changes the value.
// Spec: "OnChange func(int) - callback when value changes"
func TestScrollBarKeyboardDownArrowCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	called := false
	sb.OnChange = func(v int) { called = true }

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called after Down arrow key")
	}
}

// TestScrollBarKeyboardPageUpCallsOnChange verifies OnChange is called after Page Up.
// Spec: "OnChange func(int) - callback when value changes"
func TestScrollBarKeyboardPageUpCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	called := false
	sb.OnChange = func(v int) { called = true }

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called after Page Up key")
	}
}

// TestScrollBarKeyboardCtrlPageUpCallsOnChange verifies OnChange is called after
// Ctrl+Page Up jumps to min.
// Spec: "OnChange func(int) - callback when value changes"
func TestScrollBarKeyboardCtrlPageUpCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	called := false
	sb.OnChange = func(v int) { called = true }

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp, Modifiers: tcell.ModCtrl}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called after Ctrl+Page Up key")
	}
}

// TestScrollBarKeyboardHomeCallsOnChange verifies OnChange is called after Home.
// Spec: "OnChange func(int) - callback when value changes"
func TestScrollBarKeyboardHomeCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	called := false
	sb.OnChange = func(v int) { called = true }

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called after Home key")
	}
}

// TestScrollBarKeyboardEndCallsOnChange verifies OnChange is called after End.
// Spec: "OnChange func(int) - callback when value changes"
func TestScrollBarKeyboardEndCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	called := false
	sb.OnChange = func(v int) { called = true }

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called after End key")
	}
}

// ---------------------------------------------------------------------------
// Section 12 — Falsifying tests
// ---------------------------------------------------------------------------

// TestScrollBarVerticalUpDoesNotActAsDown verifies Up arrow decreases (not increases) value.
// Falsification guard: catches a swap of Up/Down handlers.
// Spec: "Up arrow: step(-arStep) (decrease value)"
func TestScrollBarVerticalUpDoesNotActAsDown(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	sb.HandleEvent(ev)

	if sb.Value() >= 50 {
		t.Errorf("Up arrow: Value() = %d, want < 50 (Up must decrease, not increase)", sb.Value())
	}
}

// TestScrollBarVerticalDownDoesNotActAsUp verifies Down arrow increases (not decreases) value.
// Falsification guard: catches a swap of Down/Up handlers.
// Spec: "Down arrow: step(arStep) (increase value)"
func TestScrollBarVerticalDownDoesNotActAsUp(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	sb.HandleEvent(ev)

	if sb.Value() <= 50 {
		t.Errorf("Down arrow: Value() = %d, want > 50 (Down must increase, not decrease)", sb.Value())
	}
}

// TestScrollBarVerticalPageUpStepsDifferentlyFromArrowUp verifies Page Up steps by
// pageSize, not arStep — they are distinct operations.
// Spec: "Page Up: page(-1) (decrease by pageSize)" vs "Up arrow: step(-arStep)"
func TestScrollBarVerticalPageUpStepsDifferentlyFromArrowUp(t *testing.T) {
	// arStep=1, pageSize=20: Page Up should step 20, not 1.
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(20)
	sb.SetArStep(1)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp}}
	sb.HandleEvent(ev)

	if sb.Value() == 49 {
		t.Error("Page Up stepped by arStep=1 instead of pageSize=20; these must be distinct")
	}
	if sb.Value() != 30 {
		t.Errorf("Page Up (pageSize=20): Value() = %d, want 30 (50 - 20)", sb.Value())
	}
}

// TestScrollBarVerticalCtrlPageUpGoesToMinNotStepOne verifies Ctrl+Page Up jumps to
// min outright, not just steps by one page.
// Spec: "Ctrl+Page Up: go to minVal"
func TestScrollBarVerticalCtrlPageUpGoesToMinNotStepOne(t *testing.T) {
	// If value is 50 and pageSize=10, one page-step gives 40. Min is 0.
	// Must reach 0, not 40.
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp, Modifiers: tcell.ModCtrl}}
	sb.HandleEvent(ev)

	if sb.Value() == 40 {
		t.Error("Ctrl+Page Up stepped by one page (40) instead of jumping to min (0)")
	}
	if sb.Value() != 0 {
		t.Errorf("Ctrl+Page Up: Value() = %d, want 0 (must jump to min, not step)", sb.Value())
	}
}

// TestScrollBarHorizontalCtrlRightIsNotPlainRight verifies Ctrl+Right steps by
// pageSize, not arStep.
// Spec: "Ctrl+Right: page(1) (increase by pageSize)" vs "Right arrow: increase by arStep"
func TestScrollBarHorizontalCtrlRightIsNotPlainRight(t *testing.T) {
	// arStep=1, pageSize=25: Ctrl+Right should step 25, not 1.
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(25)
	sb.SetArStep(1)
	sb.SetValue(30)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}}
	sb.HandleEvent(ev)

	if sb.Value() == 31 {
		t.Error("Ctrl+Right stepped by arStep=1 instead of pageSize=25; these must be distinct")
	}
	if sb.Value() != 55 {
		t.Errorf("Ctrl+Right (pageSize=25): Value() = %d, want 55 (30 + 25)", sb.Value())
	}
}

// TestScrollBarHorizontalEndGoesToMaxNotStepPageSize verifies End jumps to max,
// not just steps by one page.
// Spec: "End: go to maxVal"
func TestScrollBarHorizontalEndGoesToMaxNotStepPageSize(t *testing.T) {
	// value=30, pageSize=10: one Ctrl+Right gives 40, End must give 90 (max-pageSize).
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(30)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd}}
	sb.HandleEvent(ev)

	if sb.Value() == 40 {
		t.Error("End stepped by one page (40) instead of jumping to max (90)")
	}
	if sb.Value() != 90 {
		t.Errorf("End: Value() = %d, want 90 (max-pageSize)", sb.Value())
	}
}

// TestScrollBarVerticalKeyboardOrientationSensitive verifies vertical keys (Up/Down)
// have no effect on a horizontal scrollbar.
// This is a structural sanity check: vertical and horizontal scrollbars respond
// to orientation-appropriate keys.
// Spec: "For vertical scrollbars: Up arrow … Down arrow …" / "For horizontal scrollbars: Left arrow … Right arrow …"
func TestScrollBarVerticalKeyboardOrientationSensitive(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	// Up/Down are vertical-only keys — should not affect a horizontal scrollbar
	evUp := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	sb.HandleEvent(evUp)

	if sb.Value() != 50 {
		t.Errorf("Up key on horizontal scrollbar changed value to %d; vertical keys must not affect horizontal scrollbar", sb.Value())
	}

	evDown := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	sb.HandleEvent(evDown)

	if sb.Value() != 50 {
		t.Errorf("Down key on horizontal scrollbar changed value to %d; vertical keys must not affect horizontal scrollbar", sb.Value())
	}
}

// TestScrollBarHorizontalKeyboardOrientationSensitive verifies horizontal keys
// (Left/Right without modifiers) have no effect on a vertical scrollbar.
// Spec: "For vertical scrollbars: Up arrow … Down arrow …" / "For horizontal scrollbars: Left arrow … Right arrow …"
func TestScrollBarHorizontalKeyboardOrientationSensitive(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	// Left/Right are horizontal-only keys — should not affect a vertical scrollbar
	evLeft := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	sb.HandleEvent(evLeft)

	if sb.Value() != 50 {
		t.Errorf("Left key on vertical scrollbar changed value to %d; horizontal keys must not affect vertical scrollbar", sb.Value())
	}

	evRight := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	sb.HandleEvent(evRight)

	if sb.Value() != 50 {
		t.Errorf("Right key on vertical scrollbar changed value to %d; horizontal keys must not affect vertical scrollbar", sb.Value())
	}
}

// TestScrollBarArStepDefaultIsNotZero verifies arStep defaults to 1 (not zero),
// ensuring Up/Down arrows actually move the scrollbar.
// Spec: "Add arStep int field (default arStep=1)"
func TestScrollBarArStepDefaultIsNotZero(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	sb.HandleEvent(ev)

	if sb.Value() == 50 {
		t.Error("Down arrow with default arStep did not change value; default arStep must be 1 (not 0)")
	}
}
