package tv

// history_event_test.go — Tests for Task 4: History Widget — Event Handling.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec behaviour it verifies.
//
// Test organisation:
//   Section 1  — Mouse: Button1 focuses the linked InputLine
//   Section 2  — Mouse: Button1 on non-selectable link
//   Section 3  — Mouse: Button1 on selectable link
//   Section 4  — Mouse: non-Button1 buttons are ignored
//   Section 5  — Keyboard (PostProcess): Down arrow
//   Section 6  — Broadcast: CmReleasedFocus
//   Section 7  — Broadcast: CmRecordHistory
//   Section 8  — Broadcast: other commands ignored

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// mouseEvent constructs an EvMouse event for the given button.
func mouseButton1Event() *Event {
	return &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.Button1},
	}
}

func mouseButtonEvent(btn tcell.ButtonMask) *Event {
	return &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: btn},
	}
}

func downArrowEvent() *Event {
	return &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyDown},
	}
}

func upArrowEvent() *Event {
	return &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyUp},
	}
}

func broadcastEvent(cmd CommandCode, info any) *Event {
	return &Event{
		What:    EvBroadcast,
		Command: cmd,
		Info:    info,
	}
}

// newHistoryInWindow creates a Window containing both a History and its linked
// InputLine, ready for focus and mouse tests.
func newHistoryInWindow() (*Window, *History, *InputLine) {
	win := NewWindow(NewRect(0, 0, 40, 10), "test")
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(20, 0, 3, 1), il, 1)
	win.Insert(il)
	win.Insert(h)
	return win, h, il
}

// ---------------------------------------------------------------------------
// Section 1 — Mouse: Button1 focuses the linked InputLine
// ---------------------------------------------------------------------------

// TestHistoryMouseClickFocusesLink verifies that a Button1 click on History
// calls SetFocusedChild(h.link) on the owner, making the InputLine focused.
// Spec: "Mouse click (Button1): Focus linked InputLine via h.link.Owner().SetFocusedChild(h.link)"
func TestHistoryMouseClickFocusesLink(t *testing.T) {
	win, h, il := newHistoryInWindow()

	// Insert a second selectable view and focus it, so the InputLine is not focused.
	other := NewInputLine(NewRect(0, 2, 20, 1), 0)
	win.Insert(other)
	win.SetFocusedChild(other) // shift focus away from il

	if win.FocusedChild() == il {
		t.Fatal("precondition: InputLine should not be focused before click")
	}

	ev := mouseButton1Event()
	h.HandleEvent(ev)

	if win.FocusedChild() != il {
		t.Errorf("after Button1 click, FocusedChild = %v, want the linked InputLine", win.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Mouse: Button1 on non-selectable link
// ---------------------------------------------------------------------------

// TestHistoryMouseClickNonSelectableClearsEvent verifies that when the linked
// InputLine lacks OfSelectable, a Button1 click still clears the event
// (no dropdown attempt, but event is consumed).
// Spec: "If link lacks OfSelectable: clear event, return (no dropdown)"
func TestHistoryMouseClickNonSelectableClearsEvent(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 40, 10), "test")
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetOptions(OfSelectable, false) // remove OfSelectable
	h := NewHistory(NewRect(20, 0, 3, 1), il, 1)
	win.Insert(il)
	win.Insert(h)

	ev := mouseButton1Event()
	h.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Button1 click with non-selectable link: event should be cleared")
	}
}

// TestHistoryMouseClickNonSelectableDoesNotRecord verifies that a Button1 click
// when the link lacks OfSelectable does not add any history entry.
// Spec: "If link lacks OfSelectable: clear event, return (no dropdown)"
// (no recording implied by "return" — only broadcasts record)
func TestHistoryMouseClickNonSelectableDoesNotRecord(t *testing.T) {
	DefaultHistory.Clear()

	win := NewWindow(NewRect(0, 0, 40, 10), "test")
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetOptions(OfSelectable, false)
	il.SetText("hello")
	h := NewHistory(NewRect(20, 0, 3, 1), il, 1)
	win.Insert(il)
	win.Insert(h)

	ev := mouseButton1Event()
	h.HandleEvent(ev)

	entries := DefaultHistory.Entries(1)
	if len(entries) != 0 {
		t.Errorf("Button1 click with non-selectable link: unexpected history entries %v", entries)
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Mouse: Button1 on selectable link
// ---------------------------------------------------------------------------

// TestHistoryMouseClickSelectableClearsEvent verifies that when the linked
// InputLine has OfSelectable, a Button1 click clears the event (after opening
// the dropdown attempt, which returns early without an Application).
// Spec: "If link has OfSelectable: open dropdown (clear event)"
func TestHistoryMouseClickSelectableClearsEvent(t *testing.T) {
	_, h, _ := newHistoryInWindow()
	// InputLine has OfSelectable by default (set in NewInputLine).

	ev := mouseButton1Event()
	h.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Button1 click with selectable link: event should be cleared")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Mouse: non-Button1 buttons are ignored
// ---------------------------------------------------------------------------

// TestHistoryMouseButton2NotConsumed verifies that a Button2 click does NOT
// clear the event — History only handles Button1.
// Spec: "Non-Button1 mouse: ignored (event NOT cleared)"
func TestHistoryMouseButton2NotConsumed(t *testing.T) {
	_, h, _ := newHistoryInWindow()

	ev := mouseButtonEvent(tcell.Button2)
	h.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Button2 click should NOT clear the event")
	}
}

// TestHistoryMouseButton3NotConsumed verifies that a Button3 click does NOT
// clear the event.
// Spec: "Non-Button1 mouse: ignored (event NOT cleared)"
func TestHistoryMouseButton3NotConsumed(t *testing.T) {
	_, h, _ := newHistoryInWindow()

	ev := mouseButtonEvent(tcell.Button3)
	h.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Button3 click should NOT clear the event")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Keyboard (PostProcess): Down arrow
// ---------------------------------------------------------------------------

// TestHistoryDownArrowWhenLinkFocusedClearsEvent verifies that pressing Down
// when the linked InputLine has SfSelected clears the event (opens dropdown).
// Spec: "Down arrow when h.link.HasState(SfSelected): open dropdown (clear event)"
func TestHistoryDownArrowWhenLinkFocusedClearsEvent(t *testing.T) {
	_, h, il := newHistoryInWindow()
	il.SetState(SfSelected, true) // mark link as selected/focused

	ev := downArrowEvent()
	h.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Down arrow when link is SfSelected: event should be cleared")
	}
}

// TestHistoryDownArrowWhenLinkNotFocusedNotConsumed verifies that pressing Down
// when the linked InputLine does NOT have SfSelected leaves the event untouched.
// Spec: "Down arrow when link NOT focused: not consumed"
func TestHistoryDownArrowWhenLinkNotFocusedNotConsumed(t *testing.T) {
	_, h, il := newHistoryInWindow()
	il.SetState(SfSelected, false) // link is not focused

	ev := downArrowEvent()
	h.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Down arrow when link is NOT SfSelected: event should NOT be cleared")
	}
}

// TestHistoryOtherKeyNotConsumed verifies that keys other than Down arrow are
// not consumed by History (event is left uncleared).
// Spec: "Other keys: not consumed"
func TestHistoryOtherKeyNotConsumed(t *testing.T) {
	_, h, il := newHistoryInWindow()
	il.SetState(SfSelected, true) // even when focused, only Down is handled

	ev := upArrowEvent()
	h.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Up arrow should NOT be consumed by History")
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Broadcast: CmReleasedFocus
// ---------------------------------------------------------------------------

// TestHistoryBroadcastReleasedFocusRecordsText verifies that a CmReleasedFocus
// broadcast whose Info equals h.link records the link's current text.
// Spec: "CmReleasedFocus with event.Info == h.link: record text via DefaultHistory.Add(historyID, link.Text())"
func TestHistoryBroadcastReleasedFocusRecordsText(t *testing.T) {
	DefaultHistory.Clear()

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	h := NewHistory(NewRect(20, 0, 3, 1), il, 42)

	ev := broadcastEvent(CmReleasedFocus, il)
	h.HandleEvent(ev)

	entries := DefaultHistory.Entries(42)
	if len(entries) == 0 {
		t.Fatal("CmReleasedFocus with matching Info: expected history entry, got none")
	}
	if entries[len(entries)-1] != "hello" {
		t.Errorf("CmReleasedFocus: recorded %q, want %q", entries[len(entries)-1], "hello")
	}
}

// TestHistoryBroadcastReleasedFocusWrongInfoIgnored verifies that a
// CmReleasedFocus broadcast whose Info is a different view is ignored — no
// history entry is recorded.
// Spec: "CmReleasedFocus with event.Info != h.link: ignored"
func TestHistoryBroadcastReleasedFocusWrongInfoIgnored(t *testing.T) {
	DefaultHistory.Clear()

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	h := NewHistory(NewRect(20, 0, 3, 1), il, 42)

	otherView := NewInputLine(NewRect(0, 2, 20, 1), 0)
	ev := broadcastEvent(CmReleasedFocus, otherView)
	h.HandleEvent(ev)

	entries := DefaultHistory.Entries(42)
	if len(entries) != 0 {
		t.Errorf("CmReleasedFocus with wrong Info: unexpected history entries %v", entries)
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Broadcast: CmRecordHistory
// ---------------------------------------------------------------------------

// TestHistoryBroadcastRecordHistoryRecordsText verifies that a CmRecordHistory
// broadcast records the link's current text regardless of Info.
// Spec: "CmRecordHistory: record text via DefaultHistory.Add(historyID, link.Text())"
func TestHistoryBroadcastRecordHistoryRecordsText(t *testing.T) {
	DefaultHistory.Clear()

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("world")
	h := NewHistory(NewRect(20, 0, 3, 1), il, 7)

	ev := broadcastEvent(CmRecordHistory, nil)
	h.HandleEvent(ev)

	entries := DefaultHistory.Entries(7)
	if len(entries) == 0 {
		t.Fatal("CmRecordHistory: expected history entry, got none")
	}
	if entries[len(entries)-1] != "world" {
		t.Errorf("CmRecordHistory: recorded %q, want %q", entries[len(entries)-1], "world")
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Broadcast: other commands and edge cases
// ---------------------------------------------------------------------------

// TestHistoryBroadcastOtherCommandIgnored verifies that broadcast commands
// other than CmReleasedFocus and CmRecordHistory do not record history.
// Spec: "Other broadcasts: ignored"
func TestHistoryBroadcastOtherCommandIgnored(t *testing.T) {
	DefaultHistory.Clear()

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("should not be recorded")
	h := NewHistory(NewRect(20, 0, 3, 1), il, 5)

	ev := broadcastEvent(CmClose, nil)
	h.HandleEvent(ev)

	entries := DefaultHistory.Entries(5)
	if len(entries) != 0 {
		t.Errorf("CmClose broadcast: unexpected history entries %v", entries)
	}
}

// TestHistoryBroadcastReleasedFocusEmptyTextNotRecorded verifies that when the
// linked InputLine's text is empty, CmReleasedFocus does not add an entry —
// HistoryStore.Add is a no-op for empty strings.
// Spec: "CmReleasedFocus with event.Info == h.link: record text via DefaultHistory.Add(historyID, link.Text())"
// Per HistoryStore: "Empty strings are never stored (no-op)"
func TestHistoryBroadcastReleasedFocusEmptyTextNotRecorded(t *testing.T) {
	DefaultHistory.Clear()

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	// text is empty by default
	h := NewHistory(NewRect(20, 0, 3, 1), il, 3)

	ev := broadcastEvent(CmReleasedFocus, il)
	h.HandleEvent(ev)

	entries := DefaultHistory.Entries(3)
	if len(entries) != 0 {
		t.Errorf("CmReleasedFocus with empty text: unexpected history entries %v", entries)
	}
}
