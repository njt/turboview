package tv

// scrollbar_drag_test.go — Tests for Task 5: ScrollBar Thumb Drag (spec 6.3).
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Vertical thumb drag: start, move, release
//   Section 2  — Horizontal thumb drag: start, move, release
//   Section 3  — Edge cases: track click, clamping, event guard
//   Section 4  — Falsifying tests: proportionality, release truly stops drag

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Geometry constants for vertical scrollbar (height=12)
//
// up arrow at Y=0, down arrow at Y=11, track Y=1..10 (trackLen=10)
// range [0,100], pageSize=10
// thumbLen = max(1, 10*10/100) = 1
// scrollRange = 100-10-0 = 90
// thumbPos formula: pos = value * (trackLen-thumbLen) / scrollRange = value * 9 / 90
//   value=0  → pos=0, thumb at Y=1
//   value=45 → pos=4 (45*9/90=4.5→4), thumb at Y=5
//   value=90 → pos=9, thumb at Y=10
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Geometry constants for horizontal scrollbar (width=22)
//
// left arrow at X=0, right arrow at X=21, track X=1..20 (trackLen=20)
// range [0,100], pageSize=10
// thumbLen = max(1, 20*10/100) = 2
// scrollRange = 90
// thumbPos formula: pos = value * (trackLen-thumbLen) / scrollRange = value * 18 / 90
//   value=0  → pos=0, thumb at X=1..2
//   value=90 → pos=18, thumb at X=19..20
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Section 1 — Vertical thumb drag
// ---------------------------------------------------------------------------

// TestScrollBarVerticalDragClickOnThumbStartsDrag verifies that a Button1 press
// on the thumb area begins a drag without immediately changing value.
// Spec: "When mouse press (Button1) hits the thumb area: start drag, record offset within thumb"
func TestScrollBarVerticalDragClickOnThumbStartsDrag(t *testing.T) {
	// height=12, value=0 → thumb at Y=1
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	initialValue := sb.Value()

	// Press on the thumb (Y=1)
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Value must not change on the initial thumb press
	if sb.Value() != initialValue {
		t.Errorf("thumb press: Value() = %d, want %d (unchanged on drag start)", sb.Value(), initialValue)
	}
}

// TestScrollBarVerticalDragDownMovesValueHigher verifies that dragging the thumb
// downward proportionally increases the value.
// Spec: "During drag (Button1 held): calculate new value proportionally from mouse position in track"
func TestScrollBarVerticalDragDownMovesValueHigher(t *testing.T) {
	// height=12, value=0 → thumb starts at Y=1
	// Drag to Y=5: mouseTrackPos = 5-1 = 4; value = 0 + 4*90/(10-1) = 40
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb to start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag down (Button1 still held) to Y=5
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if sb.Value() <= 0 {
		t.Errorf("drag down: Value() = %d, want > 0 (should increase proportionally)", sb.Value())
	}
}

// TestScrollBarVerticalDragToBottomReachesMaxValue verifies dragging the thumb
// to the bottom of the track yields a value at or near max-pageSize (90).
// Spec: "Thumb position maps linearly from minVal to maxVal across available track space"
func TestScrollBarVerticalDragToBottomReachesMaxValue(t *testing.T) {
	// height=12, track Y=1..10, thumb at Y=10 means pos=9 → value=90
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb at Y=1 to start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag to bottom of track (Y=10)
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 10, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if sb.Value() != 90 {
		t.Errorf("drag to bottom: Value() = %d, want 90 (max-pageSize)", sb.Value())
	}
}

// TestScrollBarVerticalDragToTopReachesMinValue verifies dragging the thumb
// to the top of the track yields a value at or near min (0).
// Spec: "Thumb position maps linearly from minVal to maxVal across available track space"
func TestScrollBarVerticalDragToTopReachesMinValue(t *testing.T) {
	// Start with value=90 (thumb at Y=10), drag to Y=1
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90)

	// Press on thumb at Y=10 to start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 10, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag to top of track (Y=1)
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if sb.Value() != 0 {
		t.Errorf("drag to top: Value() = %d, want 0 (min)", sb.Value())
	}
}

// TestScrollBarVerticalDragReleaseStopsDrag verifies that releasing the mouse button
// ends the drag, so subsequent mouse movement without Button1 does not change value.
// Spec: "On release (no Button1 held while thumbDragging): stop drag"
func TestScrollBarVerticalDragReleaseStopsDrag(t *testing.T) {
	// height=12, value=0 → thumb at Y=1
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb to start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag to mid-track
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	valueAfterDrag := sb.Value()

	// Release (Button = 0)
	release := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: 0}}
	sb.HandleEvent(release)

	// Move mouse without button (should NOT change value after release)
	move := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: 0}}
	sb.HandleEvent(move)

	if sb.Value() != valueAfterDrag {
		t.Errorf("after release, mouse move changed value to %d; drag must stop on release (was %d)", sb.Value(), valueAfterDrag)
	}
}

// TestScrollBarVerticalDragOnChangeCalledDuringDrag verifies that OnChange is called
// each time the value updates during a drag.
// Spec: "Value updates continuously during drag, calling OnChange"
func TestScrollBarVerticalDragOnChangeCalledDuringDrag(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	callCount := 0
	sb.OnChange = func(v int) { callCount++ }

	// Press on thumb to start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag to a new position (value must change, so OnChange should fire)
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if callCount == 0 {
		t.Error("OnChange was not called during drag; it must be called when the value changes during drag")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Horizontal thumb drag
// ---------------------------------------------------------------------------

// TestScrollBarHorizontalDragClickOnThumbStartsDrag verifies that a Button1 press
// on the thumb of a horizontal scrollbar begins a drag without changing value.
// Spec: "When mouse press (Button1) hits the thumb area: start drag, record offset within thumb"
func TestScrollBarHorizontalDragClickOnThumbStartsDrag(t *testing.T) {
	// width=22, value=0 → thumb at X=1..2
	sb := NewScrollBar(NewRect(0, 0, 22, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	initialValue := sb.Value()

	// Press on thumb (X=1)
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 1, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(press)

	if sb.Value() != initialValue {
		t.Errorf("horizontal thumb press: Value() = %d, want %d (unchanged on drag start)", sb.Value(), initialValue)
	}
}

// TestScrollBarHorizontalDragRightIncreasesValue verifies that dragging the thumb
// rightward proportionally increases the value.
// Spec: "During drag (Button1 held): calculate new value proportionally from mouse position in track"
func TestScrollBarHorizontalDragRightIncreasesValue(t *testing.T) {
	// width=22, value=0 → thumb at X=1..2
	// Drag to X=10: mouseTrackPos = 10-1 = 9; value = 9*90/(20-2) = 9*90/18 = 45
	sb := NewScrollBar(NewRect(0, 0, 22, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb to start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 1, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag right (Button1 still held) to X=10
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 10, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if sb.Value() <= 0 {
		t.Errorf("horizontal drag right: Value() = %d, want > 0 (should increase proportionally)", sb.Value())
	}
}

// TestScrollBarHorizontalDragToRightmostReachesMaxValue verifies dragging the thumb
// to the rightmost track position yields the maximum scrollable value (90).
// Spec: "Thumb position maps linearly from minVal to maxVal across available track space"
func TestScrollBarHorizontalDragToRightmostReachesMaxValue(t *testing.T) {
	// width=22, track X=1..20, thumbLen=2
	// thumbPos at value=90: pos=90*18/90=18, thumb at X=19..20
	// Drag to X=19 (start of thumb at max)
	sb := NewScrollBar(NewRect(0, 0, 22, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb at X=1 to start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 1, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag to rightmost track position (X=19)
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 19, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if sb.Value() != 90 {
		t.Errorf("horizontal drag to rightmost: Value() = %d, want 90 (max-pageSize)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Edge cases
// ---------------------------------------------------------------------------

// TestScrollBarVerticalTrackClickDoesNotStartDrag verifies that clicking in the
// track area (not on the thumb) does NOT start a drag — it pages instead.
// Spec: "When mouse press (Button1) hits the thumb area: start drag"
// (by implication: clicks not on the thumb should not start a drag)
func TestScrollBarVerticalTrackClickDoesNotStartDrag(t *testing.T) {
	// height=12, value=0 → thumb at Y=1
	// Click at Y=8 (well below the thumb, in the track) → should page down, not drag
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Click in track below thumb (Y=8)
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 8, Button: tcell.Button1}}
	sb.HandleEvent(press)

	valueAfterTrackClick := sb.Value()

	// If drag started, a subsequent move with Button1 would change value proportionally.
	// After a page click (no drag), the value should have jumped by pageSize, not be
	// continuously variable. Test: a further mouse move without Button1 should NOT
	// alter the value further (drag did not start).
	move := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 3, Button: 0}}
	sb.HandleEvent(move)

	if sb.Value() != valueAfterTrackClick {
		t.Errorf("track click followed by button-released move changed value from %d to %d; track click must not start drag", valueAfterTrackClick, sb.Value())
	}
}

// TestScrollBarVerticalDragBeyondBottomClampsToMax verifies that dragging the thumb
// below the track bottom clamps the value to max-pageSize (90) rather than
// producing an out-of-range value.
// Spec: "Clamp to valid range"
func TestScrollBarVerticalDragBeyondBottomClampsToMax(t *testing.T) {
	// height=12, value=0 → thumb at Y=1
	// Drag to Y=50 (far beyond track end at Y=10)
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb to start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag way beyond the track
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 50, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if sb.Value() > 90 {
		t.Errorf("drag beyond track bottom: Value() = %d, want <= 90 (clamped to max-pageSize)", sb.Value())
	}
	if sb.Value() != 90 {
		t.Errorf("drag beyond track bottom: Value() = %d, want exactly 90 (max-pageSize)", sb.Value())
	}
}

// TestScrollBarVerticalDragBeyondTopClampsToMin verifies that dragging the thumb
// above the track top clamps the value to min (0).
// Spec: "Clamp to valid range"
func TestScrollBarVerticalDragBeyondTopClampsToMin(t *testing.T) {
	// height=12, value=90 → thumb at Y=10
	// Drag to Y=-50 (far above track start at Y=1)
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90)

	// Press on thumb to start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 10, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag way above the track
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: -50, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if sb.Value() < 0 {
		t.Errorf("drag beyond track top: Value() = %d, want >= 0 (clamped to min)", sb.Value())
	}
	if sb.Value() != 0 {
		t.Errorf("drag beyond track top: Value() = %d, want exactly 0 (min)", sb.Value())
	}
}

// TestScrollBarVerticalReleaseEventReachesHandlerWhenDragging verifies that when a
// drag is in progress, a Button-released event (Button=0) is not blocked by the
// Button1 guard and actually reaches the drag-stop code.
// Spec: "The Button1 guard in HandleEvent must allow events through when thumbDragging is true"
func TestScrollBarVerticalReleaseEventReachesHandlerWhenDragging(t *testing.T) {
	// height=12, value=0 → thumb at Y=1
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb to start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag partway
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	valueAfterDrag := sb.Value()

	// Release — this must NOT be silently ignored by the Button1 guard
	release := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: 0}}
	sb.HandleEvent(release)

	// Now move without button — if release was ignored, drag is still active and value changes.
	// If release was processed, drag stopped and value stays.
	move := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 10, Button: 0}}
	sb.HandleEvent(move)

	if sb.Value() != valueAfterDrag {
		t.Errorf("release event was not processed (Button1 guard blocked it): value changed from %d to %d after release+move; release must stop the drag", valueAfterDrag, sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Falsifying tests
// ---------------------------------------------------------------------------

// TestScrollBarVerticalDragIsProportionalNotIncremental verifies the drag uses
// proportional position mapping, not incremental stepping by 1 per pixel moved.
// Spec: "value = min + (mouseTrackPos * scrollRange) / (trackLen - thumbLen)"
//
// If the implementation just incremented by 1 per drag event, dragging from Y=1
// directly to Y=6 (skipping intermediate events) would only change by ~1.
// The correct implementation maps the absolute mouse position to a value.
func TestScrollBarVerticalDragIsProportionalNotIncremental(t *testing.T) {
	// height=12, value=0 → thumb at Y=1
	// Direct drag from Y=1 to Y=6:
	// mouseTrackPos = 6-1 = 5 (0-indexed in track)
	// value = 0 + 5 * 90 / (10-1) = 450/9 = 50
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb at Y=1
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Single drag event jumping directly to Y=6 (no intermediate Y=2, Y=3, Y=4, Y=5)
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	// With proportional mapping, value should be approximately 50, definitely > 2.
	// If implementation increments by 1 per event, value would be 1 (wrong).
	if sb.Value() <= 2 {
		t.Errorf("drag to Y=6 in one step: Value() = %d; drag must be proportional (want ~50), not incremental-by-1 (would give ~1)", sb.Value())
	}
}

// TestScrollBarVerticalDragReleaseActuallyStopsDrag verifies that releasing stops
// the drag such that subsequent mouse events without Button1 do NOT continue
// changing the value.
// Spec: "On release (no Button1 held while thumbDragging): stop drag"
//
// This is a stronger version of TestScrollBarVerticalDragReleaseStopsDrag that
// uses an extreme position to make any continued drag very obvious.
func TestScrollBarVerticalDragReleaseActuallyStopsDrag(t *testing.T) {
	// height=12, value=45 → thumb at Y=5
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(45)

	// Press on thumb at Y=5
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag to Y=5 (same position — value stays ~45)
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	// Release at Y=5
	release := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: 0}}
	sb.HandleEvent(release)

	valueAfterRelease := sb.Value()

	// Move to the far bottom — if drag is still active this would set value=90
	farMove := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 10, Button: 0}}
	sb.HandleEvent(farMove)

	if sb.Value() != valueAfterRelease {
		t.Errorf("after release, buttonless move to Y=10 changed value from %d to %d; release must fully stop drag (value should stay at %d)", valueAfterRelease, sb.Value(), valueAfterRelease)
	}

	// Specifically: if drag continued, value would be ~90 (the far-bottom value).
	// Verify the value did NOT jump to 90.
	if sb.Value() == 90 {
		t.Error("after release, buttonless move caused value to reach 90 (max); drag did not stop on release")
	}
}

// TestScrollBarVerticalDragMidpointValueIsProportional verifies that the value at
// the midpoint of the track is the midpoint of the scroll range, not some other value.
// Spec: "value = min + (mouseTrackPos * scrollRange) / (trackLen - thumbLen)"
//
// For height=12, track Y=1..10 (trackLen=10), thumbLen=1, scrollRange=90:
//   Midpoint mouse Y=5 (midpoint of track):
//   mouseTrackPos = 5-1 = 4 (relative to track start at Y=1)
//   Wait — the midpoint of the track offset range [0, trackLen-thumbLen] = [0,9]
//   is at offset 4 or 5. At offset 4: value = 4*90/9 = 40. At offset 5: value = 5*90/9 = 50.
//   Mouse Y=6 → offset 5 → value 50. This is the exact midpoint of scroll range [0,90].
func TestScrollBarVerticalDragMidpointValueIsProportional(t *testing.T) {
	// height=12, value=0 → thumb at Y=1
	// Drag to Y=6: mouseTrackPos = 6-1=5; value = 5*90/9 = 50
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb at Y=1
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	// Drag to Y=6 (should yield value=50 = midpoint of scroll range)
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if sb.Value() != 50 {
		t.Errorf("drag to midpoint Y=6: Value() = %d, want 50 (midpoint of scroll range 0..90)", sb.Value())
	}
}

// TestScrollBarVerticalThumbInfoReturnsCorrectPositionAtValue0 verifies thumbInfo()
// returns (pos=0, len=1) for value=0 with the test geometry.
// Spec: "thumbPos at value=0: pos = 0*(10-1)/(100-10) = 0, thumb at Y=1"
func TestScrollBarVerticalThumbInfoReturnsCorrectPositionAtValue0(t *testing.T) {
	// height=12, trackLen=10, range [0,100], pageSize=10
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	pos, length := sb.thumbInfo()

	if pos != 0 {
		t.Errorf("thumbInfo() pos = %d, want 0 at value=0", pos)
	}
	if length < 1 {
		t.Errorf("thumbInfo() length = %d, want >= 1", length)
	}
}

// TestScrollBarVerticalThumbInfoReturnsCorrectPositionAtValue90 verifies thumbInfo()
// returns pos=9 for value=90 with the test geometry.
// Spec: "thumbPos at value=90: pos = 90*9/90 = 9, thumb at Y=10"
func TestScrollBarVerticalThumbInfoReturnsCorrectPositionAtValue90(t *testing.T) {
	// height=12, trackLen=10, range [0,100], pageSize=10
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90)

	pos, length := sb.thumbInfo()

	if pos != 9 {
		t.Errorf("thumbInfo() pos = %d, want 9 at value=90", pos)
	}
	if length < 1 {
		t.Errorf("thumbInfo() length = %d, want >= 1", length)
	}
}

// TestScrollBarHorizontalThumbInfoReturnsCorrectPositionAtValue0 verifies thumbInfo()
// returns (pos=0, len=2) for value=0 on the horizontal test geometry.
// Spec: "thumbPos at value=0: pos = 0*(20-2)/(90) = 0, thumb at X=1..2"
func TestScrollBarHorizontalThumbInfoReturnsCorrectPositionAtValue0(t *testing.T) {
	// width=22, trackLen=20, range [0,100], pageSize=10, thumbLen=2
	sb := NewScrollBar(NewRect(0, 0, 22, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	pos, length := sb.thumbInfo()

	if pos != 0 {
		t.Errorf("horizontal thumbInfo() pos = %d, want 0 at value=0", pos)
	}
	if length != 2 {
		t.Errorf("horizontal thumbInfo() length = %d, want 2 (20*10/100=2)", length)
	}
}

// TestScrollBarHorizontalThumbInfoReturnsCorrectPositionAtValue90 verifies thumbInfo()
// returns pos=18 for value=90 on the horizontal test geometry.
// Spec: "thumbPos at value=90: pos = 90*18/90 = 18, thumb at X=19..20"
func TestScrollBarHorizontalThumbInfoReturnsCorrectPositionAtValue90(t *testing.T) {
	// width=22, trackLen=20, range [0,100], pageSize=10, thumbLen=2
	sb := NewScrollBar(NewRect(0, 0, 22, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90)

	pos, length := sb.thumbInfo()

	if pos != 18 {
		t.Errorf("horizontal thumbInfo() pos = %d, want 18 at value=90", pos)
	}
	if length != 2 {
		t.Errorf("horizontal thumbInfo() length = %d, want 2", length)
	}
}
