// tv/edit_window_test.go
package tv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEditWindowCreation(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	if ew.Title() != "Untitled" {
		t.Fatalf("expected title 'Untitled' got %q", ew.Title())
	}
	if ew.Editor() == nil {
		t.Fatal("Editor should not be nil")
	}
}

func TestEditWindowWithFilename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0644)

	ew := NewEditWindow(NewRect(0, 0, 60, 20), path)
	if ew.Editor().Text() != "hello" {
		t.Fatalf("expected 'hello' got %q", ew.Editor().Text())
	}
	if ew.Title() != "test.txt" {
		t.Fatalf("expected title 'test.txt' got %q", ew.Title())
	}
}

func TestEditWindowHasAllChildren(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	children := ew.Children()
	hasVScroll := false
	hasHScroll := false
	hasEditor := false
	hasIndicator := false
	for _, child := range children {
		switch v := child.(type) {
		case *ScrollBar:
			if v.Bounds().Width() == 1 {
				hasVScroll = true
			} else {
				hasHScroll = true
			}
		case *Editor:
			hasEditor = true
		case *Indicator:
			hasIndicator = true
		}
	}
	if !hasVScroll {
		t.Fatal("missing vertical scrollbar")
	}
	if !hasHScroll {
		t.Fatal("missing horizontal scrollbar")
	}
	if !hasEditor {
		t.Fatal("missing editor")
	}
	if !hasIndicator {
		t.Fatal("missing indicator")
	}
}

func TestEditWindowIndicatorVisible(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	var ind *Indicator
	for _, child := range ew.Children() {
		if i, ok := child.(*Indicator); ok {
			ind = i
			break
		}
	}
	if ind == nil {
		t.Fatal("indicator not found")
	}
	clientH := 20 - 2
	if ind.Bounds().A.Y >= clientH {
		t.Fatalf("indicator at y=%d is outside client area (height %d)", ind.Bounds().A.Y, clientH)
	}
}

func TestEditWindowCloseIntercepted(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	ew.Editor().modified = true
	ev := &Event{What: EvCommand, Command: CmClose}
	ew.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Fatal("CmClose on modified EditWindow should be intercepted (cancelled)")
	}
}

func TestEditWindowEditorIsFocused(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	focused := ew.FocusedChild()
	if _, ok := focused.(*Editor); !ok {
		t.Fatal("Editor should be the focused child")
	}
}

func TestEditWindowIndicatorUpdates(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	ed := ew.Editor()
	ed.SetText("line1\nline2\nline3")

	ed.Memo.cursorRow = 1
	ed.Memo.cursorCol = 3
	ed.broadcastIndicator()

	var ind *Indicator
	for _, child := range ew.Children() {
		if i, ok := child.(*Indicator); ok {
			ind = i
			break
		}
	}
	if ind == nil {
		t.Fatal("indicator not found")
	}
	if ind.line != 2 || ind.col != 4 {
		t.Fatalf("expected indicator 2:4, got %d:%d", ind.line, ind.col)
	}
}

func TestEditWindowValidUnmodified(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	if !ew.Valid(CmClose) {
		t.Fatal("unmodified EditWindow should be valid for close")
	}
}

func TestEditWindowValidUnmodifiedQuit(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	if !ew.Valid(CmQuit) {
		t.Fatal("unmodified EditWindow should be valid for quit")
	}
}

func TestEditWindowMinSize(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 24, 6), "")
	b := ew.Bounds()
	if b.Width() < 24 || b.Height() < 6 {
		t.Fatalf("minimum size not enforced: %dx%d", b.Width(), b.Height())
	}
}

func TestEditWindowMinSizeEnforced(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 10, 3), "")
	b := ew.Bounds()
	if b.Width() < 24 || b.Height() < 6 {
		t.Fatalf("small bounds should be clamped to minimum: got %dx%d", b.Width(), b.Height())
	}
}
