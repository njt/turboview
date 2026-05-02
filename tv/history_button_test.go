package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// history_button_test.go — Tests for Task 5: Button.press() broadcasts CmRecordHistory.
//
// Spec: When Button.press() fires, it broadcasts CmRecordHistory to its owner
// BEFORE transforming the event to the button's command. This matches original TV
// where TButton::press() broadcasts cmRecordHistory. The broadcast causes all
// History views in the dialog to record their linked InputLine contents.
//
// Test organisation:
//   TestButtonPressRecordsHistoryEntry         — Space on focused button saves InputLine text
//   TestButtonPressRecordsBeforeEventTransform — CmRecordHistory fires before event becomes EvCommand
//   TestButtonPressWithoutHistoryNoPanic       — press() on a button with no History widget works
//   TestButtonPressCommandStillCorrect         — the button's command is preserved after press
//   TestButtonPressMouseRecordsHistory         — mouse click also triggers the broadcast

// newButtonWithHistory builds a Window containing a Button, an InputLine, and a
// History widget linked to that InputLine. The button has focus (SfSelected).
// historyID 99 is used so tests can query DefaultHistory.Entries(99).
func newButtonWithHistory() (*Window, *Button, *InputLine, *History) {
	win := NewWindow(NewRect(0, 0, 40, 10), "test")
	il := NewInputLine(NewRect(1, 1, 20, 1), 80)
	h := NewHistory(NewRect(21, 1, 3, 1), il, 99)
	btn := NewButton(NewRect(1, 3, 10, 1), "~O~K", CmOK)

	win.Insert(il)
	win.Insert(h)
	win.Insert(btn)
	win.SetFocusedChild(btn)
	btn.SetState(SfSelected, true)

	return win, btn, il, h
}

// TestButtonPressRecordsHistoryEntry verifies that pressing Space on a focused
// button (which calls press()) broadcasts CmRecordHistory, causing the History
// widget to record the linked InputLine's current text.
// Spec: "broadcasts CmRecordHistory to its owner BEFORE transforming the event"
func TestButtonPressRecordsHistoryEntry(t *testing.T) {
	DefaultHistory.Clear()

	_, btn, il, _ := newButtonWithHistory()
	il.SetText("my search term")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	btn.HandleEvent(ev)

	entries := DefaultHistory.Entries(99)
	if len(entries) == 0 {
		t.Fatal("after Button.press(), expected history entry from CmRecordHistory broadcast, got none")
	}
	if entries[len(entries)-1] != "my search term" {
		t.Errorf("recorded entry = %q, want %q", entries[len(entries)-1], "my search term")
	}
}

// TestButtonPressRecordsBeforeEventTransform verifies that CmRecordHistory is
// broadcast before the event is transformed. We verify this indirectly: the
// History widget receives CmRecordHistory (records text), and the event is
// subsequently transformed to EvCommand. Both must hold after a single press.
// Spec: "broadcasts CmRecordHistory BEFORE transforming the event"
func TestButtonPressRecordsBeforeEventTransform(t *testing.T) {
	DefaultHistory.Clear()

	_, btn, il, _ := newButtonWithHistory()
	il.SetText("query")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	btn.HandleEvent(ev)

	// History entry must exist (broadcast happened).
	entries := DefaultHistory.Entries(99)
	if len(entries) == 0 {
		t.Error("CmRecordHistory broadcast did not record history entry")
	}

	// Event must have been transformed to EvCommand (press also fired).
	if ev.What != EvCommand {
		t.Errorf("after press(), ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
}

// TestButtonPressWithoutHistoryNoPanic verifies that press() works normally when
// there is no History widget in the owner — the CmRecordHistory broadcast finds
// no handler and press() still transforms the event.
// Spec: "Without a History widget present, press() still works normally (no panic)"
func TestButtonPressWithoutHistoryNoPanic(t *testing.T) {
	DefaultHistory.Clear()

	win := NewWindow(NewRect(0, 0, 40, 10), "test")
	btn := NewButton(NewRect(1, 1, 10, 1), "~O~K", CmOK)
	win.Insert(btn)
	win.SetFocusedChild(btn)
	btn.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}

	// Must not panic.
	btn.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("press() without History: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("press() without History: ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
}

// TestButtonPressCommandStillCorrect verifies that after press() broadcasts
// CmRecordHistory, the button's own command is still set correctly on the event.
// Spec: "the button command itself is still correct after press"
func TestButtonPressCommandStillCorrect(t *testing.T) {
	DefaultHistory.Clear()

	win := NewWindow(NewRect(0, 0, 40, 10), "test")
	il := NewInputLine(NewRect(1, 1, 20, 1), 80)
	h := NewHistory(NewRect(21, 1, 3, 1), il, 99)
	customCmd := CmUser + 7
	btn := NewButton(NewRect(1, 3, 12, 1), "~A~pply", customCmd)
	win.Insert(il)
	win.Insert(h)
	win.Insert(btn)
	win.SetFocusedChild(btn)
	btn.SetState(SfSelected, true)

	il.SetText("value")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	btn.HandleEvent(ev)

	if ev.Command != customCmd {
		t.Errorf("after press(), ev.Command = %v, want custom command %v", ev.Command, customCmd)
	}
}

// TestButtonPressMouseRecordsHistory verifies that a mouse click (which also
// calls press()) also triggers the CmRecordHistory broadcast.
// Spec: "broadcasts CmRecordHistory to its owner BEFORE transforming the event"
func TestButtonPressMouseRecordsHistory(t *testing.T) {
	DefaultHistory.Clear()

	_, btn, il, _ := newButtonWithHistory()
	il.SetText("mouse triggered")

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 3, Y: 0, Button: tcell.Button1},
	}
	btn.HandleEvent(ev)

	entries := DefaultHistory.Entries(99)
	if len(entries) == 0 {
		t.Fatal("mouse click press(): expected history entry from CmRecordHistory broadcast, got none")
	}
	if entries[len(entries)-1] != "mouse triggered" {
		t.Errorf("recorded entry = %q, want %q", entries[len(entries)-1], "mouse triggered")
	}
}
