package tv

// integration_batch2_memo_advanced_test.go — Integration tests for advanced Memo
// features working together (Task 7: Integration Checkpoint).
//
// Each test exercises at least two subsystems interacting:
//   - Selection model + Draw (selection style rendering)
//   - Ctrl+A + Ctrl+C + Ctrl+V (copy / paste round-trip)
//   - Double-click + Ctrl+X (mouse word select + cut)
//   - Ctrl+Left + Shift+Ctrl+Right (word movement + keyboard word selection)
//   - Character typing over a selection (replace semantics)
//   - Vertical scrollbar + PgDn (scrollbar sync after page navigation)
//   - Mouse click + Shift+Right (mouse cursor positioning + keyboard selection)
//   - Tab rendering in Draw output
//   - Mouse drag + Ctrl+C (drag selection + copy)

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Helpers local to this file
// ---------------------------------------------------------------------------

// advShiftKeyEv creates a keyboard event with Shift held.
func advShiftKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModShift}}
}

// advCtrlKeyEv creates a keyboard event with Ctrl held.
func advCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModCtrl}}
}

// advShiftCtrlKeyEv creates a keyboard event with Shift+Ctrl held.
func advShiftCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModShift | tcell.ModCtrl}}
}

// advPlainKeyEv creates a plain keyboard event with no modifiers.
func advPlainKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key}}
}

// advRuneEv creates a keyboard event for a printable rune.
func advRuneEv(r rune) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}}
}

// advMouseClickEv creates a Button1 click event.
func advMouseClickEv(x, y, clickCount int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: clickCount}}
}

// advMouseDragEv creates a Button1 motion event (drag, no ClickCount).
func advMouseDragEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1}}
}

// advMouseReleaseEv creates a ButtonNone (release) event.
func advMouseReleaseEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.ButtonNone}}
}

// newAdvMemo creates a Memo with a BorlandBlue colour scheme sized w×h.
func newAdvMemo(w, h int) *Memo {
	m := NewMemo(NewRect(0, 0, w, h))
	m.scheme = theme.BorlandBlue
	return m
}

// ---------------------------------------------------------------------------
// Test 1 — Selection model + Draw: Shift+Right creates selection rendered in
//           MemoSelected style.
// ---------------------------------------------------------------------------

// TestIntegrationShiftRightSelectionRenderedInSelectedStyle verifies that after
// Shift+Right the selected characters are rendered with MemoSelected style in
// Draw, and the unselected characters use MemoNormal.
// Features: selection model (Shift+Right) + selection-aware Draw.
func TestIntegrationShiftRightSelectionRenderedInSelectedStyle(t *testing.T) {
	m := newAdvMemo(20, 3)
	m.SetText("hello")

	// Shift+Right three times — select "hel" (rune indices 0, 1, 2).
	m.HandleEvent(advShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(advShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(advShiftKeyEv(tcell.KeyRight))

	if !m.HasSelection() {
		t.Fatal("precondition: HasSelection() should be true after Shift+Right x3")
	}
	_, sc, _, ec := m.Selection()
	if sc != 0 || ec != 3 {
		t.Fatalf("precondition: selection cols should be [0,3), got [%d,%d)", sc, ec)
	}

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	scheme := theme.BorlandBlue

	// Characters at cols 0, 1, 2 ("hel") must use MemoSelected.
	for col := 0; col < 3; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoSelected {
			t.Errorf("Selected char at cell(%d,0).Style = %v, want MemoSelected %v", col, cell.Style, scheme.MemoSelected)
		}
	}

	// Characters at cols 3, 4 ("lo") must use MemoNormal.
	for col := 3; col < 5; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Unselected char at cell(%d,0).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 2 — Ctrl+A + Ctrl+C + Ctrl+V: select-all, copy, paste into new Memo.
// Features: Ctrl+A (select all) + Ctrl+C (copy) + Ctrl+V (paste) clipboard round-trip.
// ---------------------------------------------------------------------------

// TestIntegrationCtrlACtrlCCtrlVRoundTrip verifies that Ctrl+A selects all text,
// Ctrl+C copies it to the clipboard, and then Ctrl+V in a new Memo pastes the
// full text correctly.
func TestIntegrationCtrlACtrlCCtrlVRoundTrip(t *testing.T) {
	clipboard = ""
	src := NewMemo(NewRect(0, 0, 40, 10))
	src.SetText("hello\nworld")

	// Ctrl+A selects all.
	src.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlA}})
	if !src.HasSelection() {
		t.Fatal("Ctrl+A should select all; HasSelection() = false")
	}

	// Ctrl+C copies selection to clipboard.
	src.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlC}})
	if clipboard != "hello\nworld" {
		t.Fatalf("clipboard after Ctrl+A+C = %q, want %q", clipboard, "hello\nworld")
	}

	// Ctrl+V pastes into a new Memo.
	dst := NewMemo(NewRect(0, 0, 40, 10))
	dst.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlV}})

	if got := dst.Text(); got != "hello\nworld" {
		t.Errorf("Ctrl+V paste into new Memo: Text() = %q, want %q", got, "hello\nworld")
	}
}

// ---------------------------------------------------------------------------
// Test 3 — Double-click word select + Ctrl+X: verifies clipboard and buffer state.
// Features: mouse double-click (word selection) + Ctrl+X (cut).
// ---------------------------------------------------------------------------

// TestIntegrationDoubleClickWordSelectThenCtrlX verifies that double-clicking on
// a word selects it, and Ctrl+X then cuts it — leaving the correct text in both
// the clipboard and the buffer.
func TestIntegrationDoubleClickWordSelectThenCtrlX(t *testing.T) {
	clipboard = ""
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("hello world")

	// Double-click inside "hello" (col 2, row 0).
	m.HandleEvent(advMouseClickEv(2, 0, 2))

	if !m.HasSelection() {
		t.Fatal("double-click should select a word; HasSelection() = false")
	}
	_, sc, _, ec := m.Selection()
	// "hello" spans cols 0–5 (selEndCol == 5 exclusive).
	if sc != 0 || ec != 5 {
		t.Fatalf("double-click word selection = [%d,%d), want [0,5)", sc, ec)
	}

	// Ctrl+X cuts the selection.
	m.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlX}})

	if clipboard != "hello" {
		t.Errorf("clipboard after Ctrl+X = %q, want %q", clipboard, "hello")
	}
	if got := m.Text(); got != " world" {
		t.Errorf("buffer after Ctrl+X = %q, want %q", got, " world")
	}
	if m.HasSelection() {
		t.Error("HasSelection() should be false after Ctrl+X")
	}
}

// ---------------------------------------------------------------------------
// Test 4 — Ctrl+Left then Shift+Ctrl+Right: keyboard word navigation + word
//           selection interact correctly.
// Features: word movement (Ctrl+Left) + word selection (Shift+Ctrl+Right).
// ---------------------------------------------------------------------------

// TestIntegrationCtrlLeftThenShiftCtrlRightSelectsWord verifies that Ctrl+Left
// moves the cursor back one word and Shift+Ctrl+Right then selects forward one
// word from the new position, producing the expected selection.
func TestIntegrationCtrlLeftThenShiftCtrlRightSelectsWord(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("hello world")

	// Position cursor at end of "world" (col 11).
	m.HandleEvent(advPlainKeyEv(tcell.KeyEnd))
	row, col := m.CursorPos()
	if row != 0 || col != 11 {
		t.Fatalf("precondition: cursor should be at (0,11) after End, got (%d,%d)", row, col)
	}

	// Ctrl+Left — move back one word (to start of "world", col 6).
	m.HandleEvent(advCtrlKeyEv(tcell.KeyLeft))
	_, col = m.CursorPos()
	if col != 6 {
		t.Fatalf("after Ctrl+Left: cursor col = %d, want 6 (start of \"world\")", col)
	}

	// There should be no selection yet.
	if m.HasSelection() {
		t.Fatal("no selection should exist after Ctrl+Left alone")
	}

	// Shift+Ctrl+Right — extend selection forward one word.
	m.HandleEvent(advShiftCtrlKeyEv(tcell.KeyRight))

	if !m.HasSelection() {
		t.Fatal("Shift+Ctrl+Right should create a selection; HasSelection() = false")
	}
	_, sc, _, ec := m.Selection()
	// "world" runs from col 6 to col 11; selStart at 6, selEnd at 11.
	if sc != 6 {
		t.Errorf("selection start col = %d, want 6 (start of \"world\")", sc)
	}
	if ec != 11 {
		t.Errorf("selection end col = %d, want 11 (end of \"world\")", ec)
	}
}

// ---------------------------------------------------------------------------
// Test 5 — Type character while selection active: selection is replaced.
// Features: selection model + selection-aware character insertion.
// ---------------------------------------------------------------------------

// TestIntegrationTypingCharOverSelectionReplacesAndVerifiesText verifies that
// typing a character while a selection exists replaces the selected text, and
// the resulting buffer content is correct.
func TestIntegrationTypingCharOverSelectionReplacesAndVerifiesText(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("hello world")

	// Select "hello" by pressing Shift+Right five times from col 0.
	for i := 0; i < 5; i++ {
		m.HandleEvent(advShiftKeyEv(tcell.KeyRight))
	}
	if !m.HasSelection() {
		t.Fatal("precondition: selection should exist")
	}

	// Type 'X' — must replace "hello".
	m.HandleEvent(advRuneEv('X'))

	if got := m.Text(); got != "X world" {
		t.Errorf("after typing 'X' over selection: Text() = %q, want %q", got, "X world")
	}

	// Cursor should be right after 'X' at col 1, row 0.
	row, col := m.CursorPos()
	if row != 0 || col != 1 {
		t.Errorf("after replace-type: CursorPos() = (%d,%d), want (0,1)", row, col)
	}

	// No residual selection.
	if m.HasSelection() {
		t.Error("HasSelection() should be false after typing over selection")
	}
}

// ---------------------------------------------------------------------------
// Test 6 — Vertical scrollbar + PgDn: scrollbar value updates after page navigation.
// Features: scrollbar integration (SetVScrollBar) + PgDn navigation.
// ---------------------------------------------------------------------------

// TestIntegrationVScrollBarUpdatesAfterPgDn verifies that a linked vertical
// scrollbar has its value updated after PgDn scrolls the Memo's viewport.
func TestIntegrationVScrollBarUpdatesAfterPgDn(t *testing.T) {
	// Memo height = 3 so a single PgDn causes viewport to shift.
	m := NewMemo(NewRect(0, 0, 40, 3))
	// Build enough lines that PgDn moves the viewport (height-1 = 2 lines).
	m.SetText("line0\nline1\nline2\nline3\nline4\nline5")

	sb := NewScrollBar(NewRect(39, 0, 1, 3), Vertical)
	m.SetVScrollBar(sb)

	// Verify starting value is 0.
	if sb.Value() != 0 {
		t.Fatalf("precondition: scrollbar value = %d, want 0", sb.Value())
	}

	// Press PgDn — should advance viewport (deltaY > 0) and sync scrollbar.
	m.HandleEvent(advPlainKeyEv(tcell.KeyPgDn))

	if sb.Value() == 0 {
		t.Error("scrollbar value still 0 after PgDn; expected deltaY > 0 to be reflected in scrollbar")
	}
}

// ---------------------------------------------------------------------------
// Test 7 — Mouse click positions cursor, then Shift+Right extends selection.
// Features: mouse click (cursor positioning) + Shift+Right (selection extension).
// ---------------------------------------------------------------------------

// TestIntegrationMouseClickThenShiftRightExtendsSelection verifies that a mouse
// click positions the cursor and that Shift+Right correctly extends a selection
// starting at the clicked position.
func TestIntegrationMouseClickThenShiftRightExtendsSelection(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("hello world")

	// Single click at (x=3, y=0) — cursor lands at (row=0, col=3).
	m.HandleEvent(advMouseClickEv(3, 0, 1))
	row, col := m.CursorPos()
	if row != 0 || col != 3 {
		t.Fatalf("click at (3,0): CursorPos() = (%d,%d), want (0,3)", row, col)
	}
	if m.HasSelection() {
		t.Fatal("single click should not create a selection")
	}

	// Shift+Right three times — extends selection from col 3 to col 6.
	for i := 0; i < 3; i++ {
		m.HandleEvent(advShiftKeyEv(tcell.KeyRight))
	}

	if !m.HasSelection() {
		t.Fatal("Shift+Right should create a selection; HasSelection() = false")
	}
	sr, sc, er, ec := m.Selection()
	// selStart should be at the original click position (0,3).
	if sr != 0 || sc != 3 {
		t.Errorf("selection start = (%d,%d), want (0,3) (click position)", sr, sc)
	}
	// selEnd should be three positions right at (0,6).
	if er != 0 || ec != 6 {
		t.Errorf("selection end = (%d,%d), want (0,6) (click+3 rights)", er, ec)
	}
}

// ---------------------------------------------------------------------------
// Test 8 — Tab characters render as expanded spaces in Draw output.
// Features: tab rendering (Draw) + SetText with tab content.
// ---------------------------------------------------------------------------

// TestIntegrationTabRendersAsExpandedSpacesInDraw verifies that a line containing
// a tab followed by text renders the tab as expanded spaces in the Draw output,
// and that the text following the tab appears at the correct visual column.
func TestIntegrationTabRendersAsExpandedSpacesInDraw(t *testing.T) {
	m := newAdvMemo(20, 3)
	// "\tABC": tab at rune 0 (visual col 0 → expands to 8 spaces), then 'A' at visual col 8.
	m.SetText("\tABC")

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	// Visual cols 0–7 must be spaces (tab expansion).
	for col := 0; col < 8; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("tab expansion: cell(%d,0).Rune = %q, want ' ' (tab expands to 8 spaces)", col, cell.Rune)
		}
	}

	// 'A' must appear at visual col 8.
	if buf.GetCell(8, 0).Rune != 'A' {
		t.Errorf("after tab expansion: cell(8,0).Rune = %q, want 'A'", buf.GetCell(8, 0).Rune)
	}
	if buf.GetCell(9, 0).Rune != 'B' {
		t.Errorf("after tab expansion: cell(9,0).Rune = %q, want 'B'", buf.GetCell(9, 0).Rune)
	}
	if buf.GetCell(10, 0).Rune != 'C' {
		t.Errorf("after tab expansion: cell(10,0).Rune = %q, want 'C'", buf.GetCell(10, 0).Rune)
	}
}

// TestIntegrationTabWithSelectionRendersExpandedSpacesInMemoSelectedStyle verifies
// that a tab within a selection renders all its expanded spaces with MemoSelected
// style — integrating both tab expansion and selection-aware Draw.
// Features: tab rendering + selection-aware Draw.
func TestIntegrationTabWithSelectionRendersExpandedSpacesInMemoSelectedStyle(t *testing.T) {
	m := newAdvMemo(20, 3)
	// "\tX": tab at rune 0, 'X' at rune 1.
	m.SetText("\tX")

	// Ctrl+A selects both runes (selStart=(0,0), selEnd=(0,2)).
	m.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlA}})

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// The tab at rune 0 is within the selection; its 8 expanded spaces use MemoSelected.
	for col := 0; col < 8; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoSelected {
			t.Errorf("tab in selection: cell(%d,0).Style = %v, want MemoSelected %v", col, cell.Style, scheme.MemoSelected)
		}
		if cell.Rune != ' ' {
			t.Errorf("tab in selection: cell(%d,0).Rune = %q, want ' '", col, cell.Rune)
		}
	}

	// 'X' at visual col 8 is also selected.
	cellX := buf.GetCell(8, 0)
	if cellX.Rune != 'X' {
		t.Errorf("after tab: cell(8,0).Rune = %q, want 'X'", cellX.Rune)
	}
	if cellX.Style != scheme.MemoSelected {
		t.Errorf("'X' in selection: cell(8,0).Style = %v, want MemoSelected %v", cellX.Style, scheme.MemoSelected)
	}
}

// ---------------------------------------------------------------------------
// Test 9 — Mouse drag selection + Ctrl+C: copies correct dragged text.
// Features: mouse drag (selection creation) + Ctrl+C (copy).
// ---------------------------------------------------------------------------

// TestIntegrationMouseDragSelectionThenCtrlCCopiesCorrectText verifies that after
// selecting text by dragging the mouse, Ctrl+C copies exactly the dragged
// selection (not the whole line, not an empty string).
func TestIntegrationMouseDragSelectionThenCtrlCCopiesCorrectText(t *testing.T) {
	clipboard = ""
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("hello world")

	// Press Button1 at col 6, row 0 — establishes anchor.
	m.HandleEvent(advMouseClickEv(6, 0, 1))
	// Drag to col 11, row 0 — extends selection to cover "world".
	m.HandleEvent(advMouseDragEv(11, 0))
	// Release.
	m.HandleEvent(advMouseReleaseEv(11, 0))

	if !m.HasSelection() {
		t.Fatal("drag should have created a selection; HasSelection() = false")
	}

	// Ctrl+C copies the selection.
	m.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlC}})

	if clipboard != "world" {
		t.Errorf("clipboard after drag+Ctrl+C = %q, want %q", clipboard, "world")
	}

	// Source buffer must be unchanged.
	if got := m.Text(); got != "hello world" {
		t.Errorf("Ctrl+C should not modify buffer: Text() = %q, want %q", got, "hello world")
	}
}

// TestIntegrationMouseDragMultiRowThenCtrlCCopiesMultipleLines verifies that a
// drag spanning two rows and Ctrl+C copies both lines correctly.
// Features: multi-row mouse drag + clipboard copy.
func TestIntegrationMouseDragMultiRowThenCtrlCCopiesMultipleLines(t *testing.T) {
	clipboard = ""
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("abc\ndef")

	// Press Button1 at (0,0) — anchor at (row=0, col=0).
	m.HandleEvent(advMouseClickEv(0, 0, 1))
	// Drag to (3,1) — end of "def" on row 1.
	m.HandleEvent(advMouseDragEv(3, 1))
	m.HandleEvent(advMouseReleaseEv(3, 1))

	if !m.HasSelection() {
		t.Fatal("multi-row drag: HasSelection() = false after drag")
	}

	// Ctrl+C copies.
	m.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlC}})

	if clipboard != "abc\ndef" {
		t.Errorf("multi-row drag+Ctrl+C clipboard = %q, want %q", clipboard, "abc\ndef")
	}
}
