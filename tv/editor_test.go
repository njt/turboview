package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func newTestEditor(text string) *Editor {
	ed := NewEditor(NewRect(0, 0, 40, 10))
	ed.SetText(text)
	return ed
}

func sendEditorKey(v View, key tcell.Key, r rune, mod tcell.ModMask) *Event {
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: key, Rune: r, Modifiers: mod},
	}
	v.HandleEvent(ev)
	return ev
}

func TestEditorEmbedsMemo(t *testing.T) {
	ed := newTestEditor("hello")
	if ed.Text() != "hello" {
		t.Fatalf("expected 'hello' got %q", ed.Text())
	}
	row, col := ed.CursorPos()
	if row != 0 || col != 0 {
		t.Fatalf("cursor at %d:%d", row, col)
	}
}

func TestEditorModifiedFlagAfterEdit(t *testing.T) {
	ed := newTestEditor("hello")
	if ed.Modified() {
		t.Fatal("should not be modified initially")
	}
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	if !ed.Modified() {
		t.Fatal("should be modified after typing")
	}
}

func TestEditorUndoRestoresText(t *testing.T) {
	ed := newTestEditor("hello")
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	if ed.Text() != "xhello" {
		t.Fatalf("expected 'xhello' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello" {
		t.Fatalf("after undo expected 'hello' got %q", ed.Text())
	}
}

func TestEditorUndoRestoresCursor(t *testing.T) {
	ed := newTestEditor("hello")
	sendEditorKey(ed, tcell.KeyEnd, 0, 0)
	_, col := ed.CursorPos()
	if col != 5 {
		t.Fatalf("expected col 5, got %d", col)
	}
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	_, col = ed.CursorPos()
	if col != 6 {
		t.Fatalf("expected col 6, got %d", col)
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	row, col := ed.CursorPos()
	if row != 0 || col != 5 {
		t.Fatalf("after undo expected 0:5, got %d:%d", row, col)
	}
}

func TestEditorUndoSingleLevel(t *testing.T) {
	ed := newTestEditor("abc")
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	sendEditorKey(ed, tcell.KeyRune, 'y', 0)
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "xabc" {
		t.Fatalf("expected 'xabc' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "xabc" {
		t.Fatalf("second undo should be no-op, got %q", ed.Text())
	}
}

func TestEditorUndoAfterDelete(t *testing.T) {
	ed := newTestEditor("hello")
	sendEditorKey(ed, tcell.KeyEnd, 0, 0)
	sendEditorKey(ed, tcell.KeyBackspace2, 0, 0)
	if ed.Text() != "hell" {
		t.Fatalf("expected 'hell' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello" {
		t.Fatalf("after undo expected 'hello' got %q", ed.Text())
	}
}

func TestEditorUndoAfterEnter(t *testing.T) {
	ed := newTestEditor("hello")
	ed.Memo.cursorCol = 3
	sendEditorKey(ed, tcell.KeyEnter, 0, 0)
	if ed.Text() != "hel\nlo" {
		t.Fatalf("expected 'hel\\nlo' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello" {
		t.Fatalf("after undo expected 'hello' got %q", ed.Text())
	}
}

func TestEditorUndoAfterPaste(t *testing.T) {
	ed := newTestEditor("hello")
	clipboard = "PASTED"
	sendEditorKey(ed, tcell.KeyCtrlV, 0, tcell.ModCtrl)
	if ed.Text() != "PASTEDhello" {
		t.Fatalf("expected 'PASTEDhello' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello" {
		t.Fatalf("after undo expected 'hello' got %q", ed.Text())
	}
}

func TestEditorUndoAfterCut(t *testing.T) {
	ed := newTestEditor("hello")
	sendEditorKey(ed, tcell.KeyCtrlA, 0, tcell.ModCtrl)
	sendEditorKey(ed, tcell.KeyCtrlX, 0, tcell.ModCtrl)
	if ed.Text() != "" {
		t.Fatalf("expected empty after cut, got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello" {
		t.Fatalf("after undo expected 'hello' got %q", ed.Text())
	}
}

func TestEditorUndoAfterDeleteLine(t *testing.T) {
	ed := newTestEditor("line1\nline2\nline3")
	ed.Memo.cursorRow = 1
	sendEditorKey(ed, tcell.KeyCtrlY, 0, tcell.ModCtrl)
	if ed.Text() != "line1\nline3" {
		t.Fatalf("expected 'line1\\nline3' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "line1\nline2\nline3" {
		t.Fatalf("after undo expected original, got %q", ed.Text())
	}
}

func TestEditorCanUndo(t *testing.T) {
	ed := newTestEditor("hello")
	if ed.CanUndo() {
		t.Fatal("should not be able to undo initially")
	}
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	if !ed.CanUndo() {
		t.Fatal("should be able to undo after edit")
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.CanUndo() {
		t.Fatal("should not be able to undo after undoing")
	}
}

func TestEditorNavigationDoesNotSnapshot(t *testing.T) {
	ed := newTestEditor("hello\nworld")
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	sendEditorKey(ed, tcell.KeyDown, 0, 0)
	sendEditorKey(ed, tcell.KeyRight, 0, 0)
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello\nworld" {
		t.Fatalf("expected original after undo, got %q", ed.Text())
	}
}

func TestEditorSetTextClearsUndo(t *testing.T) {
	ed := newTestEditor("hello")
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	ed.SetText("fresh")
	if ed.CanUndo() {
		t.Fatal("SetText should clear undo")
	}
	if ed.Modified() {
		t.Fatal("SetText should clear modified")
	}
}

func TestFindNextAdvancesThroughMatches(t *testing.T) {
	// 20 lines: "the" appears on rows 1, 15, 18. Editor height 10 rows.
	lines := []string{
		"Line 0: nothing here",         // 0
		"Line 1: the first match",      // 1 — "the" at col 9
		"Line 2: no match",             // 2
		"Line 3: no match",             // 3
		"Line 4: no match",             // 4
		"Line 5: no match",             // 5
		"Line 6: no match",             // 6
		"Line 7: no match",             // 7
		"Line 8: no match",             // 8
		"Line 9: no match",             // 9
		"Line 10: no match",            // 10
		"Line 11: no match",            // 11
		"Line 12: no match",            // 12
		"Line 13: no match",            // 13
		"Line 14: no match",            // 14
		"Line 15: the second match",    // 15 — "the" at col 10
		"Line 16: no match",            // 16
		"Line 17: no match",            // 17
		"Line 18: the third match",     // 18 — "the" at col 10
		"Line 19: no match",            // 19
	}
	text := ""
	for i, l := range lines {
		if i > 0 {
			text += "\n"
		}
		text += l
	}

	ed := NewEditor(NewRect(0, 0, 40, 10))
	ed.SetText(text)
	ed.search.findText = "the"

	// First findNext: should find row 1, "the" at col 8, cursor at col 11
	if !ed.findNext() {
		t.Fatal("findNext should find first match")
	}
	row, col := ed.CursorPos()
	if row != 1 || col != 11 {
		t.Fatalf("after 1st find: got %d:%d, want 1:11", row, col)
	}

	// Second findNext: should advance to row 15, "the" at col 9, cursor at 12
	if !ed.findNext() {
		t.Fatal("findNext should find second match")
	}
	row, col = ed.CursorPos()
	if row != 15 || col != 12 {
		t.Fatalf("after 2nd find: got %d:%d, want 15:12", row, col)
	}

	// Viewport should have scrolled to show row 15
	if ed.Memo.deltaY+ed.Bounds().Height() <= 15 {
		t.Fatalf("row 15 not visible: deltaY=%d, height=%d", ed.Memo.deltaY, ed.Bounds().Height())
	}
	if ed.Memo.deltaY > 15 {
		t.Fatalf("row 15 scrolled past top: deltaY=%d", ed.Memo.deltaY)
	}

	// Third findNext: should advance to row 18, "the" at col 9, cursor at 12
	if !ed.findNext() {
		t.Fatal("findNext should find third match")
	}
	row, col = ed.CursorPos()
	if row != 18 || col != 12 {
		t.Fatalf("after 3rd find: got %d:%d, want 18:12", row, col)
	}

	// Fourth findNext: should wrap back to row 1, "the" at col 8, cursor at 11
	if !ed.findNext() {
		t.Fatal("findNext should wrap to first match")
	}
	row, col = ed.CursorPos()
	if row != 1 || col != 11 {
		t.Fatalf("after 4th find (wrap): got %d:%d, want 1:11", row, col)
	}
}

func TestFindNextScrollsSelectionStartVisible(t *testing.T) {
	// After jumping to a match far to the right, then wrapping back to a
	// short-line match, the selection start must be visible — not scrolled
	// off the left edge.
	ed := NewEditor(NewRect(0, 0, 30, 10))
	text := "the match here\n" +
		"no match on line 2\n" +
		"padding padding padding padding padding padding the far match"
	ed.SetText(text)
	ed.search.findText = "the"

	// First match: row 0, col 0
	ed.findNext()
	row, col := ed.CursorPos()
	if row != 0 || col != 3 {
		t.Fatalf("first match: got %d:%d, want 0:3", row, col)
	}

	// Second match: row 2, far to the right — "the" at col 50
	ed.findNext()
	row, col = ed.CursorPos()
	if row != 2 {
		t.Fatalf("second match: got row %d, want 2", row)
	}
	// deltaX should have scrolled right

	// Third match: wraps to row 0, col 0
	ed.findNext()
	row, _ = ed.CursorPos()
	if row != 0 {
		t.Fatalf("third match: got row %d, want 0", row)
	}
	// The selection starts at col 0. It must be visible (deltaX <= 0).
	sr, sc, _, _ := ed.Memo.Selection()
	if sc < ed.Memo.deltaX {
		t.Fatalf("selection start col %d is left of viewport (deltaX=%d) on row %d",
			sc, ed.Memo.deltaX, sr)
	}
}
