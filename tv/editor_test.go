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
