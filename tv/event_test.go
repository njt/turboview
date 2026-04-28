package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// === Event.Clear() and Event.IsCleared() ===

// TestEventClearSetsWhatToEvNothing verifies:
// "Event.Clear() sets What to EvNothing"
func TestEventClearSetsWhatToEvNothing(t *testing.T) {
	e := Event{What: EvKeyboard}
	e.Clear()
	if e.What != EvNothing {
		t.Errorf("After Clear(), What should be EvNothing, got %v", e.What)
	}
}

// TestEventClearedReturnsTrue verifies:
// "Event.IsCleared() returns true afterward" (after Clear())
func TestEventClearedReturnsTrue(t *testing.T) {
	e := Event{What: EvKeyboard}
	e.Clear()
	if !e.IsCleared() {
		t.Errorf("After Clear(), IsCleared() should return true")
	}
}

// TestEventNotClearedReturnsFalse verifies falsifying case:
// IsCleared() must return false when What is not EvNothing
func TestEventNotClearedReturnsFalse(t *testing.T) {
	e := Event{What: EvKeyboard}
	if e.IsCleared() {
		t.Errorf("Before Clear(), IsCleared() should return false")
	}
}

// TestEventClearMultipleTimesIsIdempotent verifies:
// Clear() can be called multiple times without error
func TestEventClearMultipleTimesIsIdempotent(t *testing.T) {
	e := Event{What: EvKeyboard}
	e.Clear()
	first := e.What
	e.Clear()
	second := e.What
	if first != second || second != EvNothing {
		t.Errorf("Clear() should be idempotent, got %v then %v", first, second)
	}
}

// TestEventClearPreservesMouse verifies:
// Clear() only sets What to EvNothing, other fields are untouched
func TestEventClearPreservesMouse(t *testing.T) {
	mouse := &MouseEvent{X: 5, Y: 10}
	e := Event{What: EvMouse, Mouse: mouse}
	e.Clear()
	if e.Mouse != mouse {
		t.Errorf("Clear() should not modify Mouse field")
	}
}

// TestEventClearPreservesKey verifies:
// Clear() only sets What to EvNothing, Key field unchanged
func TestEventClearPreservesKey(t *testing.T) {
	key := &KeyEvent{Key: tcell.KeyEnter}
	e := Event{What: EvKeyboard, Key: key}
	e.Clear()
	if e.Key != key {
		t.Errorf("Clear() should not modify Key field")
	}
}

// TestEventClearPreservesCommand verifies:
// Clear() only sets What to EvNothing, Command field unchanged
func TestEventClearPreservesCommand(t *testing.T) {
	e := Event{What: EvCommand, Command: CmOK}
	e.Clear()
	if e.Command != CmOK {
		t.Errorf("Clear() should not modify Command field")
	}
}

// === KbAlt() ===

// TestKbAltMatchesKeyEventWithAltModifier verifies:
// "KbAlt('X') matches a KeyEvent with Key=KeyRune, Rune='x', Modifiers=ModAlt"
func TestKbAltMatchesKeyEventWithAltModifier(t *testing.T) {
	kb := KbAlt('X')
	if kb.Key != tcell.KeyRune {
		t.Errorf("KbAlt('X') should have Key=KeyRune, got %v", kb.Key)
	}
	if kb.Rune != 'x' {
		t.Errorf("KbAlt('X') should have Rune='x', got %c", kb.Rune)
	}
	if kb.Mod != tcell.ModAlt {
		t.Errorf("KbAlt('X') should have Mod=ModAlt, got %v", kb.Mod)
	}
}

// TestKbAltIsCaseInsensitive verifies:
// "KbAlt is case-insensitive: KbAlt('X') matches both 'x' and 'X' key events"
// This test checks the KeyBinding itself; matching logic is in KeyEvent matching
func TestKbAltIsCaseInsensitive(t *testing.T) {
	kbUpper := KbAlt('X')
	kbLower := KbAlt('x')
	// Both should produce lowercase rune and ModAlt
	if kbUpper.Rune != 'x' || kbLower.Rune != 'x' {
		t.Errorf("KbAlt should normalize to lowercase: got %c and %c", kbUpper.Rune, kbLower.Rune)
	}
	if kbUpper.Mod != tcell.ModAlt || kbLower.Mod != tcell.ModAlt {
		t.Errorf("KbAlt should set ModAlt for both cases")
	}
}

// TestKbAltWithLowercaseInput verifies:
// KbAlt with lowercase input produces expected KeyBinding
func TestKbAltWithLowercaseInput(t *testing.T) {
	kb := KbAlt('x')
	if kb.Key != tcell.KeyRune || kb.Rune != 'x' || kb.Mod != tcell.ModAlt {
		t.Errorf("KbAlt('x') should produce (KeyRune, 'x', ModAlt)")
	}
}

// === KbFunc() ===

// TestKbFuncMatchesKeyEventWithFunctionKey verifies:
// "KbFunc(10) matches a KeyEvent with Key=KeyF10"
func TestKbFuncMatchesKeyEventWithFunctionKey(t *testing.T) {
	kb := KbFunc(10)
	if kb.Key != tcell.KeyF10 {
		t.Errorf("KbFunc(10) should have Key=KeyF10, got %v", kb.Key)
	}
}

// TestKbFuncF1 verifies:
// KbFunc(1) produces KeyF1
func TestKbFuncF1(t *testing.T) {
	kb := KbFunc(1)
	if kb.Key != tcell.KeyF1 {
		t.Errorf("KbFunc(1) should have Key=KeyF1, got %v", kb.Key)
	}
}

// TestKbFuncF12 verifies:
// KbFunc(12) produces KeyF12
func TestKbFuncF12(t *testing.T) {
	kb := KbFunc(12)
	if kb.Key != tcell.KeyF12 {
		t.Errorf("KbFunc(12) should have Key=KeyF12, got %v", kb.Key)
	}
}

// TestKbFuncNoModifiers verifies:
// KbFunc should not set any modifiers
func TestKbFuncNoModifiers(t *testing.T) {
	kb := KbFunc(5)
	if kb.Mod != tcell.ModNone {
		t.Errorf("KbFunc should have no modifiers, got %v", kb.Mod)
	}
}

// TestKbFuncRuneIsZero verifies:
// KbFunc should not set a rune
func TestKbFuncRuneIsZero(t *testing.T) {
	kb := KbFunc(5)
	if kb.Rune != 0 {
		t.Errorf("KbFunc should have Rune=0, got %c", kb.Rune)
	}
}

// === KbCtrl() ===

// TestKbCtrlMatchesCtrlNKeyEvent verifies:
// "KbCtrl('N') matches a KeyEvent with the Ctrl+N key code"
func TestKbCtrlMatchesCtrlNKeyEvent(t *testing.T) {
	kb := KbCtrl('N')
	// In tcell, Ctrl+N is encoded as tcell.KeyCtrlN (value 14)
	// which is tcell.Key('N' - 'A' + 1)
	expectedKey := tcell.Key('N' - 'A' + 1)
	if kb.Key != expectedKey {
		t.Errorf("KbCtrl('N') should have Key=%v (KeyCtrlN), got %v", expectedKey, kb.Key)
	}
	if kb.Mod != tcell.ModCtrl {
		t.Errorf("KbCtrl('N') should have Mod=ModCtrl, got %v", kb.Mod)
	}
}

// TestKbCtrlWithLowercaseInput verifies:
// KbCtrl('n') produces the same result as KbCtrl('N') since both map to Ctrl+N
func TestKbCtrlWithLowercaseInput(t *testing.T) {
	kbUpper := KbCtrl('N')
	kbLower := KbCtrl('n')
	// Both should produce the same key constant
	expectedKey := tcell.Key('N' - 'A' + 1)
	if kbUpper.Key != expectedKey {
		t.Errorf("KbCtrl('N') should have Key=%v", expectedKey)
	}
	if kbLower.Key != expectedKey {
		t.Errorf("KbCtrl('n') should have Key=%v (same as 'N')", expectedKey)
	}
	if kbUpper.Mod != tcell.ModCtrl || kbLower.Mod != tcell.ModCtrl {
		t.Errorf("Both KbCtrl('N') and KbCtrl('n') should have ModCtrl")
	}
}

// TestKbCtrlWithDifferentCharacters verifies:
// KbCtrl works with various characters, all mapping to proper key constants
func TestKbCtrlWithDifferentCharacters(t *testing.T) {
	tests := []rune{'A', 'Z', 'C', 'X'}
	for _, ch := range tests {
		kb := KbCtrl(ch)
		// Each character should map to its control key constant
		expectedKey := tcell.Key(ch - 'A' + 1)
		if kb.Key != expectedKey {
			t.Errorf("KbCtrl('%c') should have Key=%v, got %v", ch, expectedKey, kb.Key)
		}
		if kb.Mod != tcell.ModCtrl {
			t.Errorf("KbCtrl('%c') should have ModCtrl, got %v", ch, kb.Mod)
		}
	}
}

// === KbNone() ===

// TestKbNoneNeverMatches verifies:
// "KbNone() never matches any event"
// Tests the Matches() method to verify it always returns false
func TestKbNoneNeverMatches(t *testing.T) {
	kbNone := KbNone()

	// Test that it doesn't match various key events
	tests := []struct {
		name string
		ke   *KeyEvent
	}{
		{
			name: "regular key",
			ke:   &KeyEvent{Key: tcell.KeyRune, Rune: 'a'},
		},
		{
			name: "control key",
			ke:   &KeyEvent{Key: tcell.KeyCtrlA, Modifiers: tcell.ModCtrl},
		},
		{
			name: "function key",
			ke:   &KeyEvent{Key: tcell.KeyF1},
		},
		{
			name: "alt key",
			ke:   &KeyEvent{Key: tcell.KeyRune, Rune: 'x', Modifiers: tcell.ModAlt},
		},
		{
			name: "enter key",
			ke:   &KeyEvent{Key: tcell.KeyEnter},
		},
	}

	for _, test := range tests {
		if kbNone.Matches(test.ke) {
			t.Errorf("KbNone().Matches() should return false for %s", test.name)
		}
	}
}

// TestKbAltMatches verifies:
// KbAlt().Matches() returns true for matching alt events (case-insensitive)
func TestKbAltMatches(t *testing.T) {
	kbAlt := KbAlt('X')

	// Should match lowercase 'x' with alt
	if !kbAlt.Matches(&KeyEvent{Key: tcell.KeyRune, Rune: 'x', Modifiers: tcell.ModAlt}) {
		t.Errorf("KbAlt('X').Matches() should match KeyRune 'x' with ModAlt")
	}

	// Should match uppercase 'X' with alt (case-insensitive)
	if !kbAlt.Matches(&KeyEvent{Key: tcell.KeyRune, Rune: 'X', Modifiers: tcell.ModAlt}) {
		t.Errorf("KbAlt('X').Matches() should match KeyRune 'X' with ModAlt (case-insensitive)")
	}

	// Should NOT match without alt modifier
	if kbAlt.Matches(&KeyEvent{Key: tcell.KeyRune, Rune: 'x'}) {
		t.Errorf("KbAlt('X').Matches() should not match without ModAlt")
	}

	// Should NOT match different character
	if kbAlt.Matches(&KeyEvent{Key: tcell.KeyRune, Rune: 'y', Modifiers: tcell.ModAlt}) {
		t.Errorf("KbAlt('X').Matches() should not match 'y'")
	}
}

// TestKbFuncMatches verifies:
// KbFunc().Matches() returns true for matching function keys
func TestKbFuncMatches(t *testing.T) {
	kbFunc := KbFunc(10)

	// Should match F10
	if !kbFunc.Matches(&KeyEvent{Key: tcell.KeyF10}) {
		t.Errorf("KbFunc(10).Matches() should match KeyF10")
	}

	// Should NOT match F9
	if kbFunc.Matches(&KeyEvent{Key: tcell.KeyF9}) {
		t.Errorf("KbFunc(10).Matches() should not match KeyF9")
	}

	// Should NOT match F11
	if kbFunc.Matches(&KeyEvent{Key: tcell.KeyF11}) {
		t.Errorf("KbFunc(10).Matches() should not match KeyF11")
	}
}

// TestKbCtrlMatches verifies:
// KbCtrl().Matches() returns true for matching control key events
func TestKbCtrlMatches(t *testing.T) {
	kbCtrl := KbCtrl('N')
	expectedKey := tcell.Key('N' - 'A' + 1)

	// Should match the control key event
	if !kbCtrl.Matches(&KeyEvent{Key: expectedKey, Modifiers: tcell.ModCtrl}) {
		t.Errorf("KbCtrl('N').Matches() should match KeyCtrlN with ModCtrl")
	}

	// Should NOT match without ModCtrl
	if kbCtrl.Matches(&KeyEvent{Key: expectedKey}) {
		t.Errorf("KbCtrl('N').Matches() should not match without ModCtrl")
	}

	// Should NOT match different control key
	wrongKey := tcell.Key('C' - 'A' + 1)
	if kbCtrl.Matches(&KeyEvent{Key: wrongKey, Modifiers: tcell.ModCtrl}) {
		t.Errorf("KbCtrl('N').Matches() should not match KeyCtrlC")
	}
}

// === Event type constants ===

// TestEventTypeConstantsAreBitFlags verifies:
// EventType constants are defined as bit flags (1 << n)
func TestEventTypeConstantsAreBitFlags(t *testing.T) {
	if EvNothing != 0 {
		t.Errorf("EvNothing should be 0, got %v", EvNothing)
	}
	if EvMouse != 1 {
		t.Errorf("EvMouse should be 1, got %v", EvMouse)
	}
	if EvKeyboard != 2 {
		t.Errorf("EvKeyboard should be 2, got %v", EvKeyboard)
	}
	if EvCommand != 4 {
		t.Errorf("EvCommand should be 4, got %v", EvCommand)
	}
	if EvBroadcast != 8 {
		t.Errorf("EvBroadcast should be 8, got %v", EvBroadcast)
	}
}

// === MouseEvent structure ===

// TestMouseEventHasFields verifies:
// MouseEvent has X, Y, Button, Modifiers, ClickCount fields
func TestMouseEventHasFields(t *testing.T) {
	m := &MouseEvent{
		X:          10,
		Y:          20,
		Button:     tcell.Button1,
		Modifiers:  tcell.ModCtrl,
		ClickCount: 2,
	}
	if m.X != 10 || m.Y != 20 {
		t.Errorf("MouseEvent position incorrect")
	}
	if m.Button != tcell.Button1 {
		t.Errorf("MouseEvent button incorrect")
	}
	if m.Modifiers != tcell.ModCtrl {
		t.Errorf("MouseEvent modifiers incorrect")
	}
	if m.ClickCount != 2 {
		t.Errorf("MouseEvent click count incorrect")
	}
}

// === KeyEvent structure ===

// TestKeyEventHasFields verifies:
// KeyEvent has Key, Rune, Modifiers fields
func TestKeyEventHasFields(t *testing.T) {
	k := &KeyEvent{
		Key:       tcell.KeyEnter,
		Rune:      'a',
		Modifiers: tcell.ModShift,
	}
	if k.Key != tcell.KeyEnter {
		t.Errorf("KeyEvent key incorrect")
	}
	if k.Rune != 'a' {
		t.Errorf("KeyEvent rune incorrect")
	}
	if k.Modifiers != tcell.ModShift {
		t.Errorf("KeyEvent modifiers incorrect")
	}
}

// === Event structure ===

// TestEventHasAllFields verifies:
// Event has What, Mouse, Key, Command, Info fields
func TestEventHasAllFields(t *testing.T) {
	mouse := &MouseEvent{X: 1, Y: 2}
	key := &KeyEvent{Key: tcell.KeyEnter}
	info := "test"

	e := Event{
		What:    EvMouse,
		Mouse:   mouse,
		Key:     key,
		Command: CmOK,
		Info:    info,
	}

	if e.What != EvMouse {
		t.Errorf("Event.What incorrect")
	}
	if e.Mouse != mouse {
		t.Errorf("Event.Mouse incorrect")
	}
	if e.Key != key {
		t.Errorf("Event.Key incorrect")
	}
	if e.Command != CmOK {
		t.Errorf("Event.Command incorrect")
	}
	if e.Info != info {
		t.Errorf("Event.Info incorrect")
	}
}
