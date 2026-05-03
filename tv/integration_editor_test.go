// tv/integration_editor_test.go
package tv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

func TestIntegrationEditorUndoFlow(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	ew.SetColorScheme(theme.BorlandBlue)
	ed := ew.Editor()
	ed.SetText("original")

	sendEditorKey(ed, tcell.KeyRune, 'X', 0)
	if ed.Text() != "Xoriginal" {
		t.Fatalf("expected 'Xoriginal' got %q", ed.Text())
	}
	if !ed.Modified() {
		t.Fatal("should be modified")
	}

	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "original" {
		t.Fatalf("expected 'original' after undo, got %q", ed.Text())
	}
}

func TestIntegrationEditorSearchReplace(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	ew.SetColorScheme(theme.BorlandBlue)
	ed := ew.Editor()
	ed.SetText("foo bar foo baz foo")

	ed.search.findText = "foo"
	ed.search.replaceText = "qux"

	found := ed.findNext()
	if !found {
		t.Fatal("should find 'foo'")
	}
	_, sc, _, _ := ed.Selection()
	if sc != 0 {
		t.Fatalf("first match at col 0, got %d", sc)
	}

	ed.replaceCurrent()
	if ed.Text() != "qux bar foo baz foo" {
		t.Fatalf("expected 'qux bar foo baz foo' got %q", ed.Text())
	}

	count := ed.replaceAll()
	if count != 2 {
		t.Fatalf("expected 2 more replacements, got %d", count)
	}
	if ed.Text() != "qux bar qux baz qux" {
		t.Fatalf("expected all replaced, got %q", ed.Text())
	}
}

func TestIntegrationEditorFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello\nworld"), 0644)

	ew := NewEditWindow(NewRect(0, 0, 60, 20), path)
	ed := ew.Editor()

	if ed.Text() != "hello\nworld" {
		t.Fatalf("load failed: %q", ed.Text())
	}
	if ew.Title() != "test.txt" {
		t.Fatalf("title: %q", ew.Title())
	}

	sendEditorKey(ed, tcell.KeyRune, '!', 0)
	if !ed.Modified() {
		t.Fatal("should be modified")
	}

	outPath := filepath.Join(dir, "out.txt")
	err := ed.SaveFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if ed.Modified() {
		t.Fatal("should not be modified after save")
	}

	data, _ := os.ReadFile(outPath)
	if string(data) != "!hello\nworld" {
		t.Fatalf("saved content: %q", string(data))
	}
}

func TestIntegrationIndicatorTracksEditor(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	ew.SetColorScheme(theme.BorlandBlue)
	ed := ew.Editor()
	ed.SetText("line1\nline2\nline3")

	var ind *Indicator
	for _, child := range ew.Children() {
		if i, ok := child.(*Indicator); ok {
			ind = i
			break
		}
	}
	if ind == nil {
		t.Fatal("no indicator")
	}

	sendEditorKey(ed, tcell.KeyDown, 0, 0)
	if ind.line != 2 || ind.col != 1 {
		t.Fatalf("expected 2:1, got %d:%d", ind.line, ind.col)
	}

	sendEditorKey(ed, tcell.KeyRune, 'X', 0)
	if !ind.modified {
		t.Fatal("indicator should show modified")
	}
}

func TestIntegrationEditWindowDraw(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 40, 12), "")
	ew.SetColorScheme(theme.BorlandBlue)
	ew.Editor().SetText("Hello Editor")

	buf := NewDrawBuffer(40, 12)
	ew.Draw(buf)

	titleFound := false
	for x := 0; x < 40; x++ {
		if buf.cells[0][x].Rune == 'U' {
			titleFound = true
			break
		}
	}
	if !titleFound {
		t.Fatal("title 'Untitled' not found in frame")
	}
}

func TestIntegrationEditWindowCloseUnmodified(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	if !ew.Valid(CmClose) {
		t.Fatal("unmodified window should allow close")
	}
}
