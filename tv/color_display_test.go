package tv

import (
	"testing"
)

// TestNewColorDisplayDefaults verifies default foreground and background colors.
func TestNewColorDisplayDefaults(t *testing.T) {
	cd := NewColorDisplay(NewRect(0, 0, 20, 3))
	if cd.Foreground() != 7 {
		t.Errorf("expected default foreground == 7, got %d", cd.Foreground())
	}
	if cd.Background() != 0 {
		t.Errorf("expected default background == 0, got %d", cd.Background())
	}
}

// TestColorDisplaySetColors verifies SetForeground and SetBackground update values.
func TestColorDisplaySetColors(t *testing.T) {
	cd := NewColorDisplay(NewRect(0, 0, 20, 3))

	cd.SetForeground(4)
	if cd.Foreground() != 4 {
		t.Errorf("expected foreground == 4, got %d", cd.Foreground())
	}

	cd.SetBackground(1)
	if cd.Background() != 1 {
		t.Errorf("expected background == 1, got %d", cd.Background())
	}
}

// TestColorDisplayForegroundBroadcast verifies CmColorForegroundChanged updates fg.
func TestColorDisplayForegroundBroadcast(t *testing.T) {
	cd := NewColorDisplay(NewRect(0, 0, 20, 3))

	ev := &Event{
		What:    EvBroadcast,
		Command: CmColorForegroundChanged,
		Info:    2,
	}
	cd.HandleEvent(ev)

	if cd.Foreground() != 2 {
		t.Errorf("expected foreground == 2 after CmColorForegroundChanged, got %d", cd.Foreground())
	}
	if !ev.IsCleared() {
		t.Error("expected event to be cleared after handling")
	}
}

// TestColorDisplayBackgroundBroadcast verifies CmColorBackgroundChanged updates bg.
func TestColorDisplayBackgroundBroadcast(t *testing.T) {
	cd := NewColorDisplay(NewRect(0, 0, 20, 3))

	ev := &Event{
		What:    EvBroadcast,
		Command: CmColorBackgroundChanged,
		Info:    5,
	}
	cd.HandleEvent(ev)

	if cd.Background() != 5 {
		t.Errorf("expected background == 5 after CmColorBackgroundChanged, got %d", cd.Background())
	}
	if !ev.IsCleared() {
		t.Error("expected event to be cleared after handling")
	}
}

// TestColorDisplayNonBroadcastDelegated verifies non-broadcast events are
// delegated to BaseView and not cleared by ColorDisplay.
func TestColorDisplayNonBroadcastDelegated(t *testing.T) {
	cd := NewColorDisplay(NewRect(0, 0, 20, 3))

	// A keyboard event is not a broadcast — it should pass through to BaseView
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: 0}, // just some key
	}
	cd.HandleEvent(ev)

	// Event should not be cleared by ColorDisplay (BaseView may or may not
	// clear it, depending on focus state)
	// The key point is: ColorDisplay didn't clear it
	// Since it's a keyboard event and BaseView doesn't clear keyboard events,
	// it should still be EvKeyboard
	if ev.IsCleared() {
		t.Error("ColorDisplay should not clear non-broadcast events")
	}
}

// TestColorDisplayWrongBroadcastCommand verifies that a broadcast with an
// unknown command is passed through to BaseView without clearing.
func TestColorDisplayWrongBroadcastCommand(t *testing.T) {
	cd := NewColorDisplay(NewRect(0, 0, 20, 3))

	ev := &Event{
		What:    EvBroadcast,
		Command: CmOK, // not a color-related command
		Info:    42,
	}
	cd.HandleEvent(ev)

	// Event should NOT be cleared — it's passed to BaseView
	if ev.IsCleared() {
		t.Error("unhandled broadcast should not be cleared")
	}

	// Verify foreground/background unchanged
	if cd.Foreground() != 7 || cd.Background() != 0 {
		t.Error("colors should not change for unrelated broadcast")
	}
}

// TestColorDisplayDrawDoesNotPanic verifies Draw runs without panicking.
func TestColorDisplayDrawDoesNotPanic(t *testing.T) {
	cd := NewColorDisplay(NewRect(0, 0, 20, 3))
	buf := NewDrawBuffer(20, 5)

	// Should not panic
	cd.Draw(buf)
}

// TestColorDisplayForegroundBroadcastNonIntIgnored verifies that
// CmColorForegroundChanged with non-int Info is ignored.
func TestColorDisplayForegroundBroadcastNonIntIgnored(t *testing.T) {
	cd := NewColorDisplay(NewRect(0, 0, 20, 3))

	ev := &Event{
		What:    EvBroadcast,
		Command: CmColorForegroundChanged,
		Info:    "not an int",
	}
	cd.HandleEvent(ev)

	// Foreground should remain at default
	if cd.Foreground() != 7 {
		t.Errorf("expected foreground unchanged when Info is not int, got %d", cd.Foreground())
	}
}
