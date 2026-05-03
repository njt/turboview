// tv/editor_fileio_test.go
package tv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEditorLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello\nworld"), 0644)

	ed := newTestEditor("")
	err := ed.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if ed.Text() != "hello\nworld" {
		t.Fatalf("expected 'hello\\nworld' got %q", ed.Text())
	}
	if ed.FileName() != path {
		t.Fatalf("expected filename %q got %q", path, ed.FileName())
	}
	if ed.Modified() {
		t.Fatal("should not be modified after load")
	}
	if ed.CanUndo() {
		t.Fatal("should not have undo after load")
	}
}

func TestEditorSaveFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	ed := newTestEditor("save me")
	ed.modified = true
	err := ed.SaveFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "save me" {
		t.Fatalf("expected 'save me' got %q", string(data))
	}
	if ed.Modified() {
		t.Fatal("should not be modified after save")
	}
	if ed.FileName() != path {
		t.Fatalf("expected filename %q got %q", path, ed.FileName())
	}
}

func TestEditorLineEndingDetectionCRLF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "crlf.txt")
	os.WriteFile(path, []byte("hello\r\nworld\r\n"), 0644)

	ed := newTestEditor("")
	ed.LoadFile(path)
	if ed.Text() != "hello\nworld\n" {
		t.Fatalf("CRLF not normalized on load: %q", ed.Text())
	}

	outPath := filepath.Join(dir, "out.txt")
	ed.SaveFile(outPath)
	data, _ := os.ReadFile(outPath)
	if string(data) != "hello\r\nworld\r\n" {
		t.Fatalf("expected CRLF preserved on save, got %q", string(data))
	}
}

func TestEditorLineEndingDetectionLF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lf.txt")
	os.WriteFile(path, []byte("hello\nworld\n"), 0644)

	ed := newTestEditor("")
	ed.LoadFile(path)

	outPath := filepath.Join(dir, "out.txt")
	ed.SaveFile(outPath)
	data, _ := os.ReadFile(outPath)
	if string(data) != "hello\nworld\n" {
		t.Fatalf("expected LF preserved on save, got %q", string(data))
	}
}

func TestEditorLoadFileNotFound(t *testing.T) {
	ed := newTestEditor("")
	err := ed.LoadFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestEditorSaveFileError(t *testing.T) {
	ed := newTestEditor("data")
	err := ed.SaveFile("/nonexistent/dir/file.txt")
	if err == nil {
		t.Fatal("expected error for bad path")
	}
}

func TestEditorNewFileHasNoFilename(t *testing.T) {
	ed := newTestEditor("hello")
	if ed.FileName() != "" {
		t.Fatalf("new editor should have empty filename, got %q", ed.FileName())
	}
}
