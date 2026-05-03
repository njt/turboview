// tv/editor_search_test.go
package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestEditorFindForwardBasic(t *testing.T) {
	ed := newTestEditor("hello world hello")
	ed.search.findText = "hello"
	found := ed.findNext()
	if !found {
		t.Fatal("expected to find 'hello'")
	}
	sr, sc, er, ec := ed.Selection()
	if sr != 0 || sc != 0 || er != 0 || ec != 5 {
		t.Fatalf("expected selection 0:0-0:5, got %d:%d-%d:%d", sr, sc, er, ec)
	}
}

func TestEditorFindForwardFromCursor(t *testing.T) {
	ed := newTestEditor("hello world hello")
	ed.search.findText = "hello"
	ed.Memo.cursorCol = 1
	found := ed.findNext()
	if !found {
		t.Fatal("expected to find second 'hello'")
	}
	sr, sc, er, ec := ed.Selection()
	if sr != 0 || sc != 12 || er != 0 || ec != 17 {
		t.Fatalf("expected selection 0:12-0:17, got %d:%d-%d:%d", sr, sc, er, ec)
	}
}

func TestEditorFindWraps(t *testing.T) {
	ed := newTestEditor("hello world")
	ed.search.findText = "hello"
	ed.Memo.cursorCol = 6
	found := ed.findNext()
	if !found {
		t.Fatal("expected to find 'hello' by wrapping")
	}
	sr, sc, _, _ := ed.Selection()
	if sr != 0 || sc != 0 {
		t.Fatalf("expected wrap to 0:0, got %d:%d", sr, sc)
	}
}

func TestEditorFindNotFound(t *testing.T) {
	ed := newTestEditor("hello world")
	ed.search.findText = "xyz"
	found := ed.findNext()
	if found {
		t.Fatal("should not find 'xyz'")
	}
}

func TestEditorFindCaseInsensitive(t *testing.T) {
	ed := newTestEditor("Hello World")
	ed.search.findText = "hello"
	ed.search.caseSensitive = false
	found := ed.findNext()
	if !found {
		t.Fatal("case-insensitive search should find 'Hello'")
	}
}

func TestEditorFindCaseSensitive(t *testing.T) {
	ed := newTestEditor("Hello World")
	ed.search.findText = "hello"
	ed.search.caseSensitive = true
	found := ed.findNext()
	if found {
		t.Fatal("case-sensitive search should not find 'hello' in 'Hello'")
	}
}

func TestEditorFindWholeWords(t *testing.T) {
	ed := newTestEditor("helloworld hello world")
	ed.search.findText = "hello"
	ed.search.wholeWords = true
	found := ed.findNext()
	if !found {
		t.Fatal("expected to find whole word 'hello'")
	}
	_, sc, _, ec := ed.Selection()
	if sc != 11 || ec != 16 {
		t.Fatalf("expected col 11-16, got %d-%d", sc, ec)
	}
}

func TestEditorFindMultiline(t *testing.T) {
	ed := newTestEditor("line1\nfoo bar\nline3")
	ed.search.findText = "foo"
	found := ed.findNext()
	if !found {
		t.Fatal("expected to find 'foo'")
	}
	sr, sc, er, ec := ed.Selection()
	if sr != 1 || sc != 0 || er != 1 || ec != 3 {
		t.Fatalf("expected 1:0-1:3, got %d:%d-%d:%d", sr, sc, er, ec)
	}
}

func TestEditorFindScrollsToMatch(t *testing.T) {
	lines := "line0\n"
	for i := 1; i <= 30; i++ {
		lines += "filler\n"
	}
	lines += "target here"
	ed := newTestEditor(lines)
	ed.search.findText = "target"
	ed.findNext()
	row, _ := ed.CursorPos()
	if row < 30 {
		t.Fatalf("cursor should be at target row, got %d", row)
	}
	if ed.Memo.deltaY == 0 {
		t.Fatal("viewport should have scrolled to make match visible")
	}
}

func TestEditorReplaceBasic(t *testing.T) {
	ed := newTestEditor("hello world")
	ed.search.findText = "hello"
	ed.search.replaceText = "hi"
	ed.findNext()
	ed.replaceCurrent()
	if ed.Text() != "hi world" {
		t.Fatalf("expected 'hi world' got %q", ed.Text())
	}
}

func TestEditorReplaceAll(t *testing.T) {
	ed := newTestEditor("hello world hello")
	ed.search.findText = "hello"
	ed.search.replaceText = "hi"
	count := ed.replaceAll()
	if count != 2 {
		t.Fatalf("expected 2 replacements, got %d", count)
	}
	if ed.Text() != "hi world hi" {
		t.Fatalf("expected 'hi world hi' got %q", ed.Text())
	}
}

func TestEditorReplaceAllCaseInsensitive(t *testing.T) {
	ed := newTestEditor("Hello hello HELLO")
	ed.search.findText = "hello"
	ed.search.replaceText = "hi"
	ed.search.caseSensitive = false
	count := ed.replaceAll()
	if count != 3 {
		t.Fatalf("expected 3 replacements, got %d", count)
	}
	if ed.Text() != "hi hi hi" {
		t.Fatalf("expected 'hi hi hi' got %q", ed.Text())
	}
}

func TestEditorCtrlFClearsEvent(t *testing.T) {
	ed := newTestEditor("hello")
	ev := sendEditorKey(ed, tcell.KeyCtrlF, 0, tcell.ModCtrl)
	if !ev.IsCleared() {
		t.Fatal("Ctrl+F should be consumed by Editor")
	}
}

func TestEditorCtrlHClearsEvent(t *testing.T) {
	ed := newTestEditor("hello")
	ev := sendEditorKey(ed, tcell.KeyCtrlH, 0, tcell.ModCtrl)
	if !ev.IsCleared() {
		t.Fatal("Ctrl+H should be consumed by Editor")
	}
}

func TestEditorSearchAgainNoOp(t *testing.T) {
	ed := newTestEditor("hello")
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyF3},
	}
	ed.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Fatal("F3 should be consumed")
	}
}

func TestEditorSearchAgainRepeats(t *testing.T) {
	ed := newTestEditor("aaa bbb aaa bbb")
	ed.search.findText = "bbb"
	ed.findNext()
	_, sc1, _, _ := ed.Selection()
	ed.findNext()
	_, sc2, _, _ := ed.Selection()
	if sc1 == sc2 {
		t.Fatal("search again should advance to next match")
	}
}
