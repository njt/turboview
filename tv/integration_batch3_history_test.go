package tv

// integration_batch3_history_test.go — Integration tests verifying Tasks 1–5
// of the THistory batch work together end-to-end.
//
// Tests use REAL components; no mocks.
//
// Test list:
//   1. TestIntegrationHistoryStoreAddThenEntries
//   2. TestIntegrationHistoryDrawsCorrectCharacters
//   3. TestIntegrationHistoryBroadcastReleasedFocusRecords
//   4. TestIntegrationHistoryBroadcastRecordHistoryRecords
//   5. TestIntegrationButtonPressTriggersHistoryRecording
//   6. TestIntegrationHistoryDownArrowOnlyWhenLinkSelected
//   7. TestIntegrationHistoryMouseClickFocusesInputLine

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// 1. HistoryStore: Add then Entries — chronological order and dedup
// ---------------------------------------------------------------------------

// TestIntegrationHistoryStoreAddThenEntries verifies that adding multiple
// entries for the same ID across separate Add calls produces a chronologically
// ordered list with deduplication applied consistently.
//
// Specifically:
//   - New unique values are appended in insertion order.
//   - A duplicate of an existing entry is moved to the end (not added twice).
//   - The final Entries slice reflects all of the above.
func TestIntegrationHistoryStoreAddThenEntries(t *testing.T) {
	store := NewHistoryStore(20)
	const id = 1

	store.Add(id, "alpha")
	store.Add(id, "beta")
	store.Add(id, "gamma")

	entries := store.Entries(id)
	if len(entries) != 3 {
		t.Fatalf("after 3 unique adds: len(Entries) = %d, want 3", len(entries))
	}
	if entries[0] != "alpha" || entries[1] != "beta" || entries[2] != "gamma" {
		t.Errorf("entries = %v, want [alpha beta gamma]", entries)
	}

	// Re-add "alpha" — it should move to the end.
	store.Add(id, "alpha")

	entries = store.Entries(id)
	if len(entries) != 3 {
		t.Fatalf("after re-adding alpha: len(Entries) = %d, want 3 (no duplicate)", len(entries))
	}
	if entries[len(entries)-1] != "alpha" {
		t.Errorf("after re-add: last entry = %q, want %q", entries[len(entries)-1], "alpha")
	}

	// Add a fourth unique value.
	store.Add(id, "delta")

	entries = store.Entries(id)
	if len(entries) != 4 {
		t.Fatalf("after 4 distinct values: len(Entries) = %d, want 4", len(entries))
	}
	if entries[len(entries)-1] != "delta" {
		t.Errorf("last entry = %q, want delta", entries[len(entries)-1])
	}

	// Empty strings must not be stored.
	store.Add(id, "")
	if len(store.Entries(id)) != 4 {
		t.Errorf("empty Add should not grow entries; len = %d, want 4", len(store.Entries(id)))
	}
}

// ---------------------------------------------------------------------------
// 2. History.Draw — correct characters at positions 0, 1, 2
// ---------------------------------------------------------------------------

// TestIntegrationHistoryDrawsCorrectCharacters verifies that History.Draw
// writes the three-character sequence ▐↓▌ at buffer positions 0, 1, 2.
// A real DrawBuffer is used; no mocks.
func TestIntegrationHistoryDrawsCorrectCharacters(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 80)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	want := []rune{'▐', '↓', '▌'}
	for x, wantRune := range want {
		got := buf.GetCell(x, 0).Rune
		if got != wantRune {
			t.Errorf("position %d: rune = %q (U+%04X), want %q (U+%04X)",
				x, got, got, wantRune, wantRune)
		}
	}
}

// ---------------------------------------------------------------------------
// 3. CmReleasedFocus broadcast → records entry in DefaultHistory
// ---------------------------------------------------------------------------

// TestIntegrationHistoryBroadcastReleasedFocusRecords wires History + InputLine
// in a Window, sends a CmReleasedFocus broadcast with the InputLine as Info,
// and verifies an entry appears in DefaultHistory.
func TestIntegrationHistoryBroadcastReleasedFocusRecords(t *testing.T) {
	DefaultHistory.Clear()

	win := NewWindow(NewRect(0, 0, 40, 10), "test")
	il := NewInputLine(NewRect(0, 0, 20, 1), 80)
	h := NewHistory(NewRect(20, 0, 3, 1), il, 11)
	win.Insert(il)
	win.Insert(h)

	il.SetText("search term")

	ev := &Event{
		What:    EvBroadcast,
		Command: CmReleasedFocus,
		Info:    il,
	}
	h.HandleEvent(ev)

	entries := DefaultHistory.Entries(11)
	if len(entries) == 0 {
		t.Fatal("CmReleasedFocus with link as Info: expected history entry, got none")
	}
	if entries[len(entries)-1] != "search term" {
		t.Errorf("recorded entry = %q, want %q", entries[len(entries)-1], "search term")
	}
}

// ---------------------------------------------------------------------------
// 4. CmRecordHistory broadcast → records entry in DefaultHistory
// ---------------------------------------------------------------------------

// TestIntegrationHistoryBroadcastRecordHistoryRecords wires History + InputLine
// in a Window, sends a CmRecordHistory broadcast, and verifies an entry appears
// in DefaultHistory.
func TestIntegrationHistoryBroadcastRecordHistoryRecords(t *testing.T) {
	DefaultHistory.Clear()

	win := NewWindow(NewRect(0, 0, 40, 10), "test")
	il := NewInputLine(NewRect(0, 0, 20, 1), 80)
	h := NewHistory(NewRect(20, 0, 3, 1), il, 22)
	win.Insert(il)
	win.Insert(h)

	il.SetText("query text")

	ev := &Event{
		What:    EvBroadcast,
		Command: CmRecordHistory,
	}
	h.HandleEvent(ev)

	entries := DefaultHistory.Entries(22)
	if len(entries) == 0 {
		t.Fatal("CmRecordHistory broadcast: expected history entry, got none")
	}
	if entries[len(entries)-1] != "query text" {
		t.Errorf("recorded entry = %q, want %q", entries[len(entries)-1], "query text")
	}
}

// ---------------------------------------------------------------------------
// 5. Button Enter press → CmRecordHistory broadcast → History records
// ---------------------------------------------------------------------------

// TestIntegrationButtonPressTriggersHistoryRecording wires Button + InputLine +
// History in a Window, sets text in InputLine, sends Enter to the focused Button
// (which calls press() → broadcasts CmRecordHistory → History records), and
// verifies DefaultHistory has the entry.
func TestIntegrationButtonPressTriggersHistoryRecording(t *testing.T) {
	DefaultHistory.Clear()

	win := NewWindow(NewRect(0, 0, 40, 10), "test")
	il := NewInputLine(NewRect(1, 1, 20, 1), 80)
	h := NewHistory(NewRect(21, 1, 3, 1), il, 33)
	btn := NewButton(NewRect(1, 3, 10, 1), "~O~K", CmOK)

	win.Insert(il)
	win.Insert(h)
	win.Insert(btn)
	win.SetFocusedChild(btn)
	btn.SetState(SfSelected, true)

	il.SetText("entered value")

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: ' '},
	}
	btn.HandleEvent(ev)

	entries := DefaultHistory.Entries(33)
	if len(entries) == 0 {
		t.Fatal("Button Enter press: expected history entry from CmRecordHistory broadcast, got none")
	}
	if entries[len(entries)-1] != "entered value" {
		t.Errorf("recorded entry = %q, want %q", entries[len(entries)-1], "entered value")
	}
}

// ---------------------------------------------------------------------------
// 6. Down arrow consumed only when InputLine has SfSelected
// ---------------------------------------------------------------------------

// TestIntegrationHistoryDownArrowOnlyWhenLinkSelected verifies that a Down
// arrow event is NOT consumed by History when the linked InputLine does not
// have SfSelected, but IS consumed (cleared) when SfSelected is set.
func TestIntegrationHistoryDownArrowOnlyWhenLinkSelected(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 40, 10), "test")
	il := NewInputLine(NewRect(0, 0, 20, 1), 80)
	h := NewHistory(NewRect(20, 0, 3, 1), il, 44)
	win.Insert(il)
	win.Insert(h)

	// Without SfSelected: Down arrow must NOT be consumed.
	il.SetState(SfSelected, false)

	evDown := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyDown},
	}
	h.HandleEvent(evDown)

	if evDown.IsCleared() {
		t.Error("Down arrow without SfSelected on InputLine: event should NOT be consumed")
	}

	// With SfSelected: Down arrow IS consumed (openDropdown clears it; no
	// Application is available so openDropdown returns early after clearing).
	il.SetState(SfSelected, true)

	evDown2 := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyDown},
	}
	h.HandleEvent(evDown2)

	if !evDown2.IsCleared() {
		t.Error("Down arrow with SfSelected on InputLine: event should be consumed (cleared)")
	}
}

// ---------------------------------------------------------------------------
// 7. Mouse click on History focuses the linked InputLine
// ---------------------------------------------------------------------------

// TestIntegrationHistoryMouseClickFocusesInputLine wires History + InputLine +
// Button in a Window. Button is the initial focused child. A Button1 click on
// the History widget should cause the InputLine to become the focused child of
// the Window.
func TestIntegrationHistoryMouseClickFocusesInputLine(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 40, 10), "test")
	il := NewInputLine(NewRect(0, 0, 20, 1), 80)
	h := NewHistory(NewRect(20, 0, 3, 1), il, 55)
	btn := NewButton(NewRect(0, 3, 10, 1), "~O~K", CmOK)

	win.Insert(il)
	win.Insert(h)
	win.Insert(btn)
	win.SetFocusedChild(btn)

	if win.FocusedChild() != btn {
		t.Fatalf("precondition: FocusedChild = %v, want Button", win.FocusedChild())
	}

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.Button1},
	}
	h.HandleEvent(ev)

	if win.FocusedChild() != il {
		t.Errorf("after Button1 click on History: FocusedChild = %v, want InputLine", win.FocusedChild())
	}
}
