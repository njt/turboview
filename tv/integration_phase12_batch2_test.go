package tv

// integration_phase12_batch2_test.go — Integration tests for Phase 12 Tasks 5–6:
// ScrollBar Thumb Drag and Broadcasts, verified together.
//
// These tests combine the drag mechanics (Task 5) with the broadcast system
// (Task 6), and also integrate with the keyboard/wheel/auto-repeat features
// from Batch 1 (Tasks 1–3). Every test exercises at least two features in
// combination so the full event flow is verified end-to-end.
//
// Geometry used throughout (unless noted):
//   Vertical scrollbar, height=12
//   up arrow Y=0, down arrow Y=11, track Y=1..10 (trackLen=10)
//   range [0,100], pageSize=10
//   thumbLen = max(1, 10*10/100) = 1
//   scrollRange = 100-10-0 = 90
//   thumbPos: value=0 → Y=1, value=45 → Y=5, value=90 → Y=10
//
// Test naming: TestIntegrationPhase12Batch2<DescriptiveSuffix>

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Scenario 1: Thumb drag down → CmScrollBarClicked on press + CmScrollBarChanged on move
//
// Spec: "When any scrollbar interaction begins: broadcast CmScrollBarClicked."
// Spec: "When the scrollbar value changes: broadcast CmScrollBarChanged."
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2ThumbDragDownBroadcastsClickedThenChanged verifies that
// pressing the thumb broadcasts CmScrollBarClicked, and dragging down broadcasts
// CmScrollBarChanged.
//
// Setup: value=0, thumb at Y=1; press Y=1 → CmScrollBarClicked; drag to Y=6 → CmScrollBarChanged.
func TestIntegrationPhase12Batch2ThumbDragDownBroadcastsClickedThenChanged(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb at Y=1 — starts drag, no value change yet
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("thumb press: spy did not receive CmScrollBarClicked; broadcasts: %v", spy.broadcasts)
	}
	if hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("thumb press without value change: spy must NOT receive CmScrollBarChanged; broadcasts: %v", spy.broadcasts)
	}

	resetBroadcasts(spy)

	// Drag to Y=6 (value becomes ~50) — value changes, so CmScrollBarChanged fires
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("drag to Y=6: spy did not receive CmScrollBarChanged; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() <= 0 {
		t.Errorf("drag to Y=6: Value() = %d, want > 0", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 2: Thumb drag + verify both OnChange AND CmScrollBarChanged fire
//
// Spec: "These broadcasts complement (don't replace) the existing OnChange callback."
// Spec: "Both broadcasts AND OnChange fire (not one or the other)."
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2ThumbDragFiresBothOnChangeAndCmScrollBarChanged verifies
// that a thumb drag which changes the value triggers both OnChange and CmScrollBarChanged,
// confirming neither mechanism suppresses the other.
//
// Setup: value=0; press Y=1; drag to Y=10 (value → 90).
// Both OnChange and CmScrollBarChanged must fire on the drag event.
func TestIntegrationPhase12Batch2ThumbDragFiresBothOnChangeAndCmScrollBarChanged(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	onChangeFired := false
	sb.OnChange = func(v int) { onChangeFired = true }

	// Press on thumb to start drag
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	resetBroadcasts(spy)
	onChangeFired = false // reset after press

	// Drag to bottom (Y=10) — value must become 90
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 10, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if !onChangeFired {
		t.Error("thumb drag to bottom: OnChange did not fire; both OnChange and broadcast must fire")
	}
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("thumb drag to bottom: CmScrollBarChanged not received by spy; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 90 {
		t.Errorf("thumb drag to bottom: Value() = %d, want 90", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 3: Keyboard event → CmScrollBarClicked AND CmScrollBarChanged
//
// Spec: "When any scrollbar interaction begins: broadcast CmScrollBarClicked."
// Spec: "When the scrollbar value changes: broadcast CmScrollBarChanged."
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2KeyboardEventBroadcastsBothCommands verifies that a
// keyboard Down arrow event broadcasts both CmScrollBarClicked (interaction) and
// CmScrollBarChanged (value changed), in that order.
//
// Setup: value=50, SfSelected=true; Down arrow → value 51.
// Both CmScrollBarClicked and CmScrollBarChanged must appear in spy.broadcasts.
func TestIntegrationPhase12Batch2KeyboardEventBroadcastsBothCommands(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	sb.HandleEvent(ev)

	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("keyboard Down: spy did not receive CmScrollBarClicked; broadcasts: %v", spy.broadcasts)
	}
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("keyboard Down: spy did not receive CmScrollBarChanged; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 51 {
		t.Errorf("keyboard Down: Value() = %d, want 51", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 4: Wheel event → CmScrollBarClicked AND CmScrollBarChanged
//
// Spec: "When any scrollbar interaction begins (mouse click, wheel, keyboard):
// broadcast CmScrollBarClicked."
// Spec: "When the scrollbar value changes: broadcast CmScrollBarChanged."
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2WheelEventBroadcastsBothCommands verifies that a
// WheelDown event broadcasts both CmScrollBarClicked and CmScrollBarChanged.
//
// Setup: value=50; WheelDown → value 53 (+3*arStep with arStep=1).
func TestIntegrationPhase12Batch2WheelEventBroadcastsBothCommands(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("WheelDown: spy did not receive CmScrollBarClicked; broadcasts: %v", spy.broadcasts)
	}
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("WheelDown: spy did not receive CmScrollBarChanged; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 53 {
		t.Errorf("WheelDown: Value() = %d, want 53 (50+3)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 5: Complete drag sequence — Clicked only on press, Changed on each move
//
// Spec: "When any scrollbar interaction begins: broadcast CmScrollBarClicked."
// (i.e., once, on the initial press — not on every drag motion event)
// Spec: "When the scrollbar value changes: broadcast CmScrollBarChanged."
// (i.e., each time the value changes during drag)
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2CompleteDragSequenceClickedOnlyOnPress verifies that
// across a full press → move → move → release sequence, CmScrollBarClicked is sent
// exactly once (on the initial press) while CmScrollBarChanged is sent for each
// drag event that changes the value.
//
// Setup: value=0, thumb at Y=1.
// Press Y=1 → CmScrollBarClicked (×1), no CmScrollBarChanged.
// Drag to Y=5 → CmScrollBarChanged (×1), value ~40.
// Drag to Y=10 → CmScrollBarChanged (×1 more), value 90.
// Release → no additional broadcasts.
func TestIntegrationPhase12Batch2CompleteDragSequenceClickedOnlyOnPress(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Phase 1: press on thumb
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	clickedAfterPress := countBroadcasts(spy, CmScrollBarClicked)
	changedAfterPress := countBroadcasts(spy, CmScrollBarChanged)
	if clickedAfterPress == 0 {
		t.Error("press on thumb: CmScrollBarClicked not received")
	}
	if changedAfterPress != 0 {
		t.Errorf("press on thumb (no value change): CmScrollBarChanged should be 0, got %d", changedAfterPress)
	}

	resetBroadcasts(spy)

	// Phase 2: drag to mid-track
	drag1 := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.Button1}}
	sb.HandleEvent(drag1)

	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("drag to Y=5: CmScrollBarChanged not received; broadcasts: %v", spy.broadcasts)
	}
	// CmScrollBarClicked should NOT re-fire during a drag move
	if hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("drag to Y=5: CmScrollBarClicked must NOT fire again during drag move; broadcasts: %v", spy.broadcasts)
	}

	resetBroadcasts(spy)

	// Phase 3: drag to bottom
	drag2 := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 10, Button: tcell.Button1}}
	sb.HandleEvent(drag2)

	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("drag to Y=10: CmScrollBarChanged not received; broadcasts: %v", spy.broadcasts)
	}

	resetBroadcasts(spy)

	// Phase 4: release — should not trigger additional broadcasts
	release := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 10, Button: 0}}
	sb.HandleEvent(release)

	if len(spy.broadcasts) != 0 {
		t.Errorf("release: unexpected broadcasts after drag release; broadcasts: %v", spy.broadcasts)
	}
}

// ---------------------------------------------------------------------------
// Scenario 6: Drag then keyboard on the same scrollbar
//
// Verifies that drag and keyboard inputs compose correctly and both trigger
// appropriate broadcasts. After a drag, a keyboard event should still fire
// CmScrollBarClicked + CmScrollBarChanged.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2DragThenKeyboardBothBroadcast verifies that after
// completing a thumb drag, subsequent keyboard events still fire both CmScrollBarClicked
// and CmScrollBarChanged.
//
// Setup: value=0; drag from Y=1 to Y=6 (value→50); release; Down arrow (value→51).
// Down arrow broadcasts CmScrollBarClicked and CmScrollBarChanged.
func TestIntegrationPhase12Batch2DragThenKeyboardBothBroadcast(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)
	sb.SetState(SfSelected, true)

	// Perform a complete drag to mid-track
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}})
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: tcell.Button1}})
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: 0}}) // release

	valueAfterDrag := sb.Value()
	if valueAfterDrag == 0 {
		t.Fatalf("drag did not change value; got %d; subsequent test would be invalid", valueAfterDrag)
	}

	resetBroadcasts(spy)

	// Now use keyboard Down arrow
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})

	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("keyboard Down after drag: CmScrollBarClicked not received; broadcasts: %v", spy.broadcasts)
	}
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("keyboard Down after drag: CmScrollBarChanged not received; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != valueAfterDrag+1 {
		t.Errorf("keyboard Down after drag: Value() = %d, want %d (+1)", sb.Value(), valueAfterDrag+1)
	}
}

// ---------------------------------------------------------------------------
// Scenario 7: Drag that does not change value → CmScrollBarClicked but NOT CmScrollBarChanged
//
// Spec: "When any scrollbar interaction begins: broadcast CmScrollBarClicked."
// Spec: "When the scrollbar value changes: broadcast CmScrollBarChanged." (only then)
// If the drag doesn't move to a different value position, CmScrollBarChanged must not fire.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2ThumbPressNoMoveClickedButNotChanged verifies that
// pressing on the thumb (starting a drag) broadcasts CmScrollBarClicked but does NOT
// broadcast CmScrollBarChanged, because the value has not changed at that point.
//
// Setup: value=0, thumb at Y=1; press Y=1 only (no drag motion).
func TestIntegrationPhase12Batch2ThumbPressNoMoveClickedButNotChanged(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb — starts drag but value unchanged
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	sb.HandleEvent(press)

	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("thumb press: CmScrollBarClicked must fire even without value change; broadcasts: %v", spy.broadcasts)
	}
	if hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("thumb press without drag: CmScrollBarChanged must NOT fire (value unchanged); broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 0 {
		t.Errorf("thumb press without drag: Value() = %d, want 0 (unchanged)", sb.Value())
	}
}

// TestIntegrationPhase12Batch2DragWithinThumbNoChangeClickedButNotChanged verifies
// that dragging the mouse within the thumb's current position (same Y, so no value
// change) broadcasts CmScrollBarClicked on press but does not repeat CmScrollBarChanged
// for no-op drag moves.
//
// Setup: value=40 → thumb at Y=5; press Y=5, drag to Y=5 (same position, value stays).
// value=40 survives the round-trip: thumbPos = 40*9/90 = 4, trackPos=4 → value = 4*90/9 = 40.
func TestIntegrationPhase12Batch2DragWithinThumbNoChangeClickedButNotChanged(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(40) // thumb at Y=5 (pos=4, Y=1+4=5); 40 round-trips cleanly

	// Press on thumb at Y=5
	press := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.Button1}}
	sb.HandleEvent(press)

	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("thumb press: CmScrollBarClicked must fire; broadcasts: %v", spy.broadcasts)
	}

	resetBroadcasts(spy)

	// Drag to same Y=5 — value should not change
	drag := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.Button1}}
	sb.HandleEvent(drag)

	if hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("drag to same position (value unchanged): CmScrollBarChanged must NOT fire; broadcasts: %v", spy.broadcasts)
	}
}

// ---------------------------------------------------------------------------
// Scenario 8: Multiple interaction types in sequence — click, drag, wheel, keyboard
//
// Verifies that across a mixed session all four interaction styles produce the
// expected combination of CmScrollBarClicked and CmScrollBarChanged broadcasts.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2AllInteractionTypesInSequenceBroadcastCorrectly verifies
// that click, thumb drag, wheel, and keyboard each fire CmScrollBarClicked, and those
// that change the value also fire CmScrollBarChanged.
//
// Sequence:
//   1. Click up arrow (Y=0): value 50→49; CmScrollBarClicked + CmScrollBarChanged
//   2. Thumb drag: press Y=1 (thumb at 0; first reset value); drag Y=6 (→50);
//      press: CmScrollBarClicked only; drag: CmScrollBarChanged
//   3. WheelDown: value →53; CmScrollBarClicked + CmScrollBarChanged
//   4. Keyboard Down: value →54; CmScrollBarClicked + CmScrollBarChanged
func TestIntegrationPhase12Batch2AllInteractionTypesInSequenceBroadcastCorrectly(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	sb.SetState(SfSelected, true)

	// Step 1: click up arrow at Y=0 (value 50→49)
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}})
	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("step 1 (up arrow click): CmScrollBarClicked not received; broadcasts: %v", spy.broadcasts)
	}
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("step 1 (up arrow click): CmScrollBarChanged not received; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 49 {
		t.Errorf("step 1 (up arrow click): Value() = %d, want 49", sb.Value())
	}
	resetBroadcasts(spy)

	// Step 2a: reset to value=0 for a clean drag demo, then press on thumb
	sb.SetValue(0)
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}})
	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("step 2a (thumb press): CmScrollBarClicked not received; broadcasts: %v", spy.broadcasts)
	}
	resetBroadcasts(spy)

	// Step 2b: drag to Y=6 (value→50)
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: tcell.Button1}})
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("step 2b (drag to Y=6): CmScrollBarChanged not received; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 50 {
		t.Errorf("step 2b (drag to Y=6): Value() = %d, want 50", sb.Value())
	}
	// Release drag
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: 0}})
	resetBroadcasts(spy)

	// Step 3: WheelDown (value 50→53)
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}})
	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("step 3 (WheelDown): CmScrollBarClicked not received; broadcasts: %v", spy.broadcasts)
	}
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("step 3 (WheelDown): CmScrollBarChanged not received; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 53 {
		t.Errorf("step 3 (WheelDown): Value() = %d, want 53 (50+3)", sb.Value())
	}
	resetBroadcasts(spy)

	// Step 4: keyboard Down (value 53→54)
	sb.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("step 4 (keyboard Down): CmScrollBarClicked not received; broadcasts: %v", spy.broadcasts)
	}
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("step 4 (keyboard Down): CmScrollBarChanged not received; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 54 {
		t.Errorf("step 4 (keyboard Down): Value() = %d, want 54 (53+1)", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 9: Drag to top → value 0, then drag to bottom → value 90
// Verifies full range sweep via drag, with broadcasts at each extremity.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2DragFullRangeSweepWithBroadcasts verifies a complete
// drag sweep from top to bottom, checking value extremes and broadcasts.
//
// Setup: value=90; press on thumb at Y=10; drag to Y=1 (min); CmScrollBarChanged fires.
// Then drag back to Y=10 (max); CmScrollBarChanged fires again.
func TestIntegrationPhase12Batch2DragFullRangeSweepWithBroadcasts(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90) // thumb starts at Y=10

	// Press on thumb at Y=10 to start drag
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 10, Button: tcell.Button1}})
	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("press at max: CmScrollBarClicked not received; broadcasts: %v", spy.broadcasts)
	}
	resetBroadcasts(spy)

	// Drag to Y=1 (top of track) → value must reach 0
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}})
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("drag to Y=1 (min): CmScrollBarChanged not received; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 0 {
		t.Errorf("drag to Y=1 (min): Value() = %d, want 0", sb.Value())
	}
	resetBroadcasts(spy)

	// Drag back to Y=10 (bottom of track) → value must reach 90
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 10, Button: tcell.Button1}})
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("drag to Y=10 (max): CmScrollBarChanged not received; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 90 {
		t.Errorf("drag to Y=10 (max): Value() = %d, want 90", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Scenario 10: OnChange count across drag moves — fires per-move, not per-press
// Verifies that OnChange accumulates across multiple drag positions and that
// CmScrollBarChanged mirrors each OnChange call.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2OnChangeAndBroadcastCountMatchDuringMultipleDrags verifies
// that OnChange and CmScrollBarChanged are both called the same number of times during
// a drag sequence with multiple distinct value positions.
//
// Setup: value=0; press Y=1; drag Y=4 (→30); drag Y=6 (→50); drag Y=8 (→70).
// OnChange must fire 3 times; CmScrollBarChanged must be in spy.broadcasts 3 times.
func TestIntegrationPhase12Batch2OnChangeAndBroadcastCountMatchDuringMultipleDrags(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	onChangeValues := []int{}
	sb.OnChange = func(v int) { onChangeValues = append(onChangeValues, v) }

	// Press on thumb to start drag
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}})
	resetBroadcasts(spy)

	// Three distinct drag positions
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 4, Button: tcell.Button1}}) // value→30
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: tcell.Button1}}) // value→50
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 8, Button: tcell.Button1}}) // value→70

	changedCount := countBroadcasts(spy, CmScrollBarChanged)

	if len(onChangeValues) != 3 {
		t.Errorf("OnChange called %d times, want 3 (once per drag position); values: %v", len(onChangeValues), onChangeValues)
	}
	if changedCount != 3 {
		t.Errorf("CmScrollBarChanged broadcast %d times, want 3 (once per drag position); broadcasts: %v", changedCount, spy.broadcasts)
	}
	// OnChange and CmScrollBarChanged should be symmetric
	if len(onChangeValues) != changedCount {
		t.Errorf("OnChange count (%d) != CmScrollBarChanged count (%d); they must be symmetric", len(onChangeValues), changedCount)
	}
}

// ---------------------------------------------------------------------------
// Scenario 11: Broadcast Info field is scrollbar for drag interactions
//
// Spec: "The Info field of the broadcast event should be the ScrollBar itself."
// Verifies this holds for both CmScrollBarClicked (on press) and CmScrollBarChanged
// (on drag).
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2DragBroadcastInfoIsScrollBar verifies that the Info
// field of both CmScrollBarClicked (from initial press) and CmScrollBarChanged (from
// drag) contains the ScrollBar that produced the event.
//
// Setup: value=0; press Y=1 (CmScrollBarClicked); drag Y=6 (CmScrollBarChanged).
func TestIntegrationPhase12Batch2DragBroadcastInfoIsScrollBar(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	// Press on thumb — emits CmScrollBarClicked
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}})

	foundClicked := false
	for _, rec := range spy.broadcasts {
		if rec.command == CmScrollBarClicked {
			if rec.info != sb {
				t.Errorf("CmScrollBarClicked Info = %v (%T), want ScrollBar (%p)", rec.info, rec.info, sb)
			}
			foundClicked = true
		}
	}
	if !foundClicked {
		t.Error("CmScrollBarClicked not received at all")
	}

	resetBroadcasts(spy)

	// Drag to Y=6 — emits CmScrollBarChanged
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 6, Button: tcell.Button1}})

	foundChanged := false
	for _, rec := range spy.broadcasts {
		if rec.command == CmScrollBarChanged {
			if rec.info != sb {
				t.Errorf("CmScrollBarChanged Info = %v (%T), want ScrollBar (%p)", rec.info, rec.info, sb)
			}
			foundChanged = true
		}
	}
	if !foundChanged {
		t.Error("CmScrollBarChanged not received after drag")
	}
}

// ---------------------------------------------------------------------------
// Scenario 12: Wheel + drag interleaved — broadcasts fire for each interaction
//
// Verifies that wheel events (which do not involve a thumb drag) still produce
// the right broadcasts when mixed with a drag session.
// ---------------------------------------------------------------------------

// TestIntegrationPhase12Batch2WheelThenDragBothBroadcastCorrectly verifies that a
// WheelDown followed by a thumb drag each fire CmScrollBarClicked, and the drag
// also fires CmScrollBarChanged when the value changes.
//
// Setup: value=20; WheelDown → 23; then press on new thumb position, drag to Y=10 → 90.
func TestIntegrationPhase12Batch2WheelThenDragBothBroadcastCorrectly(t *testing.T) {
	sb, _, spy := newScrollBarInGroup(NewRect(0, 0, 1, 12))
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(20)

	// WheelDown: value 20→23
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}})
	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("WheelDown: CmScrollBarClicked not received; broadcasts: %v", spy.broadcasts)
	}
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("WheelDown: CmScrollBarChanged not received; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 23 {
		t.Errorf("WheelDown: Value() = %d, want 23 (20+3)", sb.Value())
	}

	resetBroadcasts(spy)

	// Find the current thumb Y position.
	// value=23: thumbPos = 23*9/90 = 2 (integer division), so thumb at Y=1+2=3
	// Press on thumb at Y=3
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 3, Button: tcell.Button1}})
	if !hasBroadcast(spy, CmScrollBarClicked) {
		t.Errorf("thumb press after wheel: CmScrollBarClicked not received; broadcasts: %v", spy.broadcasts)
	}
	resetBroadcasts(spy)

	// Drag to bottom (Y=10) → value 90
	sb.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 10, Button: tcell.Button1}})
	if !hasBroadcast(spy, CmScrollBarChanged) {
		t.Errorf("drag to bottom after wheel: CmScrollBarChanged not received; broadcasts: %v", spy.broadcasts)
	}
	if sb.Value() != 90 {
		t.Errorf("drag to bottom after wheel: Value() = %d, want 90", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Helper: countBroadcasts counts how many times a given command appears in spy.broadcasts.
// ---------------------------------------------------------------------------

func countBroadcasts(spy *broadcastSpyView, cmd CommandCode) int {
	n := 0
	for _, rec := range spy.broadcasts {
		if rec.command == cmd {
			n++
		}
	}
	return n
}
