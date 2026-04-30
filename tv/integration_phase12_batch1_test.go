package tv

// integration_phase12_batch1_test.go — Integration tests for Phase 12 Tasks 1–3:
// ScrollBar Keyboard Handling, Mouse Wheel Speed, and Arrow Auto-Repeat.
//
// Each test exercises at least two features in combination to verify the
// interactions between keyboard navigation, wheel scrolling, arStep configuration,
// and auto-repeat mouse clicks work correctly as a unit.
//
// Features under test:
//   Task 1: Keyboard handling (Up/Down/PgUp/PgDn/Ctrl+PgUp/Ctrl+PgDn/Left/Right/Home/End)
//   Task 2: Mouse wheel speed (WheelUp/WheelDown scroll by 3 * arStep)
//   Task 3: Arrow click auto-repeat (Button1 on arrow positions steps by 1 per click)
//
// Test naming: TestIntegrationPhase12Batch1<DescriptiveSuffix>

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Scenario 1: Keyboard + wheel on the same vertical scrollbar
// Verify that keyboard and wheel events each move the scrollbar correctly and
// their effects accumulate on the shared value.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch1KeyboardThenWheelVertical verifies that a Down arrow
// followed by a WheelDown each advance the value independently — total +4.
//
// Setup: range [0,100], pageSize=10, value=50, arStep=1 (default)
// Down arrow → value 51; WheelDown → value 51+3 = 54.
func TestIntegrationPhase12Batch1KeyboardThenWheelVertical(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	// Task 1: keyboard Down arrow (+1)
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if sb.Value() != 51 {
		t.Errorf("after Down arrow: Value() = %d, want 51", sb.Value())
	}

	// Task 2: wheel down (+3*1)
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}})
	if sb.Value() != 54 {
		t.Errorf("after Down+WheelDown: Value() = %d, want 54 (51+3)", sb.Value())
	}
}

// TestIntegrationPhase12Batch1WheelThenKeyboardVertical verifies that WheelUp followed
// by Up arrow both decrease the value — total -4.
//
// Setup: range [0,100], pageSize=10, value=50, arStep=1
// WheelUp → value 47; Up arrow → value 46.
func TestIntegrationPhase12Batch1WheelThenKeyboardVertical(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	// Task 2: wheel up (−3)
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}})
	if sb.Value() != 47 {
		t.Errorf("after WheelUp: Value() = %d, want 47", sb.Value())
	}

	// Task 1: keyboard Up arrow (−1)
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}})
	if sb.Value() != 46 {
		t.Errorf("after WheelUp+Up arrow: Value() = %d, want 46 (47-1)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 2: arStep change affects both keyboard and wheel
// Changing arStep mid-session must immediately affect wheel (3*arStep) but
// must NOT affect keyboard arrows (they always use arStep via the key handler).
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch1ArStepChangeAffectsBothKeyboardAndWheel verifies that
// after SetArStep(2), Down arrow uses 2 and WheelDown uses 3*2=6.
//
// Setup: range [0,100], pageSize=10, value=20
// SetArStep(2); Down → 22; WheelDown → 28.
func TestIntegrationPhase12Batch1ArStepChangeAffectsBothKeyboardAndWheel(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(20)
	sb.SetArStep(2)
	sb.SetState(SfSelected, true)

	// Task 1: keyboard Down (arStep=2 → +2)
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if sb.Value() != 22 {
		t.Errorf("after Down (arStep=2): Value() = %d, want 22", sb.Value())
	}

	// Task 2: wheel down (3 * arStep=2 → +6)
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}})
	if sb.Value() != 28 {
		t.Errorf("after Down+WheelDown (arStep=2): Value() = %d, want 28 (22+6)", sb.Value())
	}
}

// TestIntegrationPhase12Batch1ArStepChangedMidSessionAffectsWheelOnly verifies
// that changing arStep between events correctly updates wheel speed while
// subsequent keyboard key steps also use the new arStep.
//
// Setup: range [0,100], pageSize=10, value=50, arStep starts at 1
// WheelDown (3*1=+3) → 53; SetArStep(4); WheelDown (3*4=+12) → 65;
// Down arrow (arStep=4 → +4) → 69.
func TestIntegrationPhase12Batch1ArStepChangedMidSessionAffectsWheelOnly(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	// First wheel with arStep=1
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}})
	if sb.Value() != 53 {
		t.Errorf("WheelDown (arStep=1): Value() = %d, want 53", sb.Value())
	}

	// Change arStep and verify both wheel and keyboard use new value
	sb.SetArStep(4)

	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}})
	if sb.Value() != 65 {
		t.Errorf("WheelDown (arStep=4): Value() = %d, want 65 (53+12)", sb.Value())
	}

	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if sb.Value() != 69 {
		t.Errorf("Down arrow (arStep=4): Value() = %d, want 69 (65+4)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 3: Auto-repeat mouse clicks + keyboard navigation combined
// Arrow clicks step by 1 (regardless of arStep); keyboard arrows step by arStep.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch1AutoRepeatThenKeyboard verifies that 5 mouse clicks
// on the down arrow (each +1) followed by a keyboard Down arrow (+arStep) sum correctly.
//
// Setup: range [0,100], pageSize=10, value=50, arStep=1
// 5×Button1 on Y=9 → 55; Down keyboard → 56.
func TestIntegrationPhase12Batch1AutoRepeatThenKeyboard(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	// Task 3: 5 auto-repeat clicks on down arrow (Y=9), each steps by 1
	for i := 0; i < 5; i++ {
		sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}})
	}
	if sb.Value() != 55 {
		t.Errorf("after 5 clicks on down arrow: Value() = %d, want 55", sb.Value())
	}

	// Task 1: keyboard Down (+arStep=1)
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if sb.Value() != 56 {
		t.Errorf("after 5 clicks + Down arrow: Value() = %d, want 56", sb.Value())
	}
}

// TestIntegrationPhase12Batch1AutoRepeatArStepDoesNotAffectMouseArrows verifies that
// mouse clicks on arrows always step by 1, not by arStep, even when arStep is large.
// Then keyboard Down should step by arStep.
//
// Setup: range [0,100], pageSize=10, value=50, arStep=5
// 5×Button1 on Y=9 → 55 (not 75); Down keyboard → 60.
func TestIntegrationPhase12Batch1AutoRepeatArStepDoesNotAffectMouseArrows(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetArStep(5)
	sb.SetState(SfSelected, true)

	// Task 3: 5 auto-repeat clicks — arrow clicks always step by 1, not arStep
	for i := 0; i < 5; i++ {
		sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}})
	}
	if sb.Value() != 55 {
		t.Errorf("after 5 mouse clicks (arStep=5): Value() = %d, want 55 (mouse arrows always step by 1)", sb.Value())
	}

	// Task 1: keyboard Down uses arStep=5
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if sb.Value() != 60 {
		t.Errorf("after 5 clicks + Down arrow (arStep=5): Value() = %d, want 60 (55+5)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 4: Keyboard Page Down then auto-repeat up arrows
// Verifies page step (Task 1) and mouse auto-repeat (Task 3) interact correctly.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch1PageDownThenAutoRepeatUp verifies that PgDn moves by
// pageSize, then 3 mouse clicks on the up arrow each step back by 1.
//
// Setup: range [0,100], pageSize=10, value=50
// PgDn → 60; 3×Button1 on Y=0 → 57.
func TestIntegrationPhase12Batch1PageDownThenAutoRepeatUp(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	// Task 1: Page Down (+pageSize=10)
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgDn}})
	if sb.Value() != 60 {
		t.Errorf("after PgDn: Value() = %d, want 60", sb.Value())
	}

	// Task 3: 3 clicks on up arrow (Y=0), each -1
	for i := 0; i < 3; i++ {
		sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}})
	}
	if sb.Value() != 57 {
		t.Errorf("after PgDn + 3 up clicks: Value() = %d, want 57 (60-3)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 5: Ctrl+PgDn to max, then wheel and keyboard back toward min
// Verifies Ctrl+PgDn (jump to max) interacts with wheel and keyboard.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch1CtrlPgDnThenWheelUp verifies Ctrl+PgDn jumps to
// max-pageSize (90), then WheelUp scrolls back by 3.
//
// Setup: range [0,100], pageSize=10, value=20
// Ctrl+PgDn → 90; WheelUp → 87.
func TestIntegrationPhase12Batch1CtrlPgDnThenWheelUp(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(20)
	sb.SetState(SfSelected, true)

	// Task 1: Ctrl+PgDn jumps to max (90)
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgDn, Modifiers: tcell.ModCtrl}})
	if sb.Value() != 90 {
		t.Errorf("after Ctrl+PgDn: Value() = %d, want 90 (max-pageSize)", sb.Value())
	}

	// Task 2: WheelUp (−3*1)
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}})
	if sb.Value() != 87 {
		t.Errorf("after Ctrl+PgDn + WheelUp: Value() = %d, want 87 (90-3)", sb.Value())
	}
}

// TestIntegrationPhase12Batch1CtrlPgDnThenKeyboardBack verifies Ctrl+PgDn to max,
// then keyboard PgUp scrolls back by pageSize.
//
// Setup: range [0,100], pageSize=10, value=20
// Ctrl+PgDn → 90; PgUp → 80.
func TestIntegrationPhase12Batch1CtrlPgDnThenKeyboardBack(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(20)
	sb.SetState(SfSelected, true)

	// Task 1: Ctrl+PgDn → max=90
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgDn, Modifiers: tcell.ModCtrl}})
	if sb.Value() != 90 {
		t.Errorf("after Ctrl+PgDn: Value() = %d, want 90", sb.Value())
	}

	// Task 1: PgUp (−pageSize=10) → 80
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp}})
	if sb.Value() != 80 {
		t.Errorf("after Ctrl+PgDn + PgUp: Value() = %d, want 80 (90-10)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 6: Horizontal scrollbar — keyboard + wheel + auto-repeat together
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch1HorizontalKeyboardWheelAutoRepeat verifies that on a
// horizontal scrollbar, Home (→ min), WheelDown (+3), and 2 right-arrow clicks (+2)
// compose correctly.
//
// Setup: range [0,100], pageSize=10, value=40
// Home → 0; WheelDown → 3; 2×Button1 at X=11 → 5.
func TestIntegrationPhase12Batch1HorizontalKeyboardWheelAutoRepeat(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(40)
	sb.SetState(SfSelected, true)

	// Task 1: Home → min=0
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}})
	if sb.Value() != 0 {
		t.Errorf("after Home: Value() = %d, want 0", sb.Value())
	}

	// Task 2: WheelDown (+3*1)
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.WheelDown}})
	if sb.Value() != 3 {
		t.Errorf("after Home+WheelDown: Value() = %d, want 3", sb.Value())
	}

	// Task 3: 2 right-arrow clicks (X=11, each +1)
	for i := 0; i < 2; i++ {
		sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 11, Y: 0, Button: tcell.Button1}})
	}
	if sb.Value() != 5 {
		t.Errorf("after Home+WheelDown+2 right clicks: Value() = %d, want 5 (3+2)", sb.Value())
	}
}

// TestIntegrationPhase12Batch1HorizontalEndThenAutoRepeatLeft verifies End (→ max)
// then left-arrow auto-repeat decrements correctly.
//
// Setup: range [0,100], pageSize=10, value=20
// End → 90; 5×Button1 at X=0 → 85.
func TestIntegrationPhase12Batch1HorizontalEndThenAutoRepeatLeft(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(20)
	sb.SetState(SfSelected, true)

	// Task 1: End → max-pageSize=90
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd}})
	if sb.Value() != 90 {
		t.Errorf("after End: Value() = %d, want 90", sb.Value())
	}

	// Task 3: 5 left-arrow clicks (X=0, each -1)
	for i := 0; i < 5; i++ {
		sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}})
	}
	if sb.Value() != 85 {
		t.Errorf("after End+5 left clicks: Value() = %d, want 85 (90-5)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 7: OnChange fires for all input methods in sequence
// Verifies that the callback wires correctly across keyboard, wheel, and auto-repeat.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch1OnChangeFiredByAllThreeMethods verifies that OnChange
// is invoked by keyboard, wheel, and auto-repeat in a single session.
//
// Setup: Down arrow (×1), WheelDown (×1), Button1 on Y=9 (×1) → 3 OnChange calls.
func TestIntegrationPhase12Batch1OnChangeFiredByAllThreeMethods(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	callCount := 0
	sb.OnChange = func(v int) { callCount++ }

	// Task 1: keyboard
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	// Task 2: wheel
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}})
	// Task 3: auto-repeat click
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}})

	if callCount != 3 {
		t.Errorf("OnChange called %d times, want 3 (once per input method)", callCount)
	}
}

// TestIntegrationPhase12Batch1OnChangeReceivesCorrectSequenceOfValues verifies the exact
// values delivered to OnChange across keyboard, wheel, and auto-repeat events.
//
// Setup: value=50; Down(+1)→51; WheelDown(+3)→54; 2×down click(+1,+1)→55,56.
func TestIntegrationPhase12Batch1OnChangeReceivesCorrectSequenceOfValues(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	var got []int
	sb.OnChange = func(v int) { got = append(got, v) }

	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}})
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}})
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}})

	want := []int{51, 54, 55, 56}
	if len(got) != len(want) {
		t.Fatalf("OnChange called %d times, want %d; got values: %v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("OnChange call %d: got %d, want %d", i+1, got[i], w)
		}
	}
}

// ---------------------------------------------------------------------------
// Scenario 8: Wheel speed scales with arStep; keyboard also uses new arStep
// Combined scenario: change arStep, then interleave wheel and keyboard.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch1WheelSpeedAndKeyboardBothScaleWithArStep verifies that
// when arStep=3, WheelDown=+9 and Down arrow=+3, and auto-repeat clicks remain +1.
//
// Setup: range [0,100], pageSize=10, value=30, arStep=3
// WheelDown → 39; Down → 42; Button1 on Y=9 → 43.
func TestIntegrationPhase12Batch1WheelSpeedAndKeyboardBothScaleWithArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(30)
	sb.SetArStep(3)
	sb.SetState(SfSelected, true)

	// Task 2: WheelDown (+3*3=9)
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}})
	if sb.Value() != 39 {
		t.Errorf("WheelDown (arStep=3): Value() = %d, want 39 (30+9)", sb.Value())
	}

	// Task 1: keyboard Down (+arStep=3)
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if sb.Value() != 42 {
		t.Errorf("Down arrow (arStep=3): Value() = %d, want 42 (39+3)", sb.Value())
	}

	// Task 3: mouse click on down arrow (always +1, not +arStep)
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}})
	if sb.Value() != 43 {
		t.Errorf("down arrow click (arStep=3): Value() = %d, want 43 (42+1, mouse clicks always step by 1)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 9: Clamping across all three input methods at the boundary
// Ensures no method can push the value past max-pageSize.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch1AllMethodsClampAtMax verifies that wheel, keyboard, and
// auto-repeat all clamp at max-pageSize when the value is already at the boundary.
//
// Setup: range [0,100], pageSize=10, value=89
// WheelDown (+3) → 90 (not 92); Down arrow → 90 (stays); Button1 at Y=9 → 90.
func TestIntegrationPhase12Batch1AllMethodsClampAtMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(89)
	sb.SetState(SfSelected, true)

	// Task 2: WheelDown — would reach 92, clamped to 90
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}})
	if sb.Value() != 90 {
		t.Errorf("WheelDown near max: Value() = %d, want 90 (clamped)", sb.Value())
	}

	// Task 1: keyboard Down — already at max, stays
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if sb.Value() != 90 {
		t.Errorf("Down arrow at max: Value() = %d, want 90 (clamped)", sb.Value())
	}

	// Task 3: mouse click on down arrow — also stays at max
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}})
	if sb.Value() != 90 {
		t.Errorf("down arrow click at max: Value() = %d, want 90 (clamped)", sb.Value())
	}
}

// TestIntegrationPhase12Batch1AllMethodsClampAtMin verifies that wheel, keyboard, and
// auto-repeat all clamp at min when the value is already at the boundary.
//
// Setup: range [0,100], pageSize=10, value=1
// WheelUp (−3) → 0; Up arrow → 0; Button1 at Y=0 → 0.
func TestIntegrationPhase12Batch1AllMethodsClampAtMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(1)
	sb.SetState(SfSelected, true)

	// Task 2: WheelUp — would reach −2, clamped to 0
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}})
	if sb.Value() != 0 {
		t.Errorf("WheelUp near min: Value() = %d, want 0 (clamped)", sb.Value())
	}

	// Task 1: keyboard Up — already at min, stays
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}})
	if sb.Value() != 0 {
		t.Errorf("Up arrow at min: Value() = %d, want 0 (clamped)", sb.Value())
	}

	// Task 3: mouse click on up arrow — also stays at min
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}})
	if sb.Value() != 0 {
		t.Errorf("up arrow click at min: Value() = %d, want 0 (clamped)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 10: Keyboard navigation then auto-repeat on horizontal scrollbar
// Combines Right-arrow keyboard (Task 1) with auto-repeat mouse right-arrow (Task 3)
// and verifies that arStep affects keyboard but not mouse arrow clicks.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch1HorizontalKeyboardAutoRepeatWithCustomArStep verifies
// that on a horizontal scrollbar with arStep=4, Right arrow uses arStep=4 but
// mouse right-arrow clicks use 1.
//
// Setup: range [0,100], pageSize=10, value=30, arStep=4
// Right arrow → 34; 3×Button1 at X=11 → 37.
func TestIntegrationPhase12Batch1HorizontalKeyboardAutoRepeatWithCustomArStep(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(30)
	sb.SetArStep(4)
	sb.SetState(SfSelected, true)

	// Task 1: keyboard Right (+arStep=4)
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}})
	if sb.Value() != 34 {
		t.Errorf("Right arrow (arStep=4): Value() = %d, want 34 (30+4)", sb.Value())
	}

	// Task 3: 3 mouse right-arrow clicks (each +1, not +arStep)
	for i := 0; i < 3; i++ {
		sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 11, Y: 0, Button: tcell.Button1}})
	}
	if sb.Value() != 37 {
		t.Errorf("Right arrow + 3 right clicks (arStep=4): Value() = %d, want 37 (34+3, clicks always +1)", sb.Value())
	}
}
