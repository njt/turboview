package tv

// file_dialog_test.go — Tests for Task 4 (Batch 7): FileInputLine + FileDialog.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the spec requirement it verifies.
//
// Test organisation:
//   Section 1  — FileInputLine: autofill on CmFileFocused (requirements 1-3)
//   Section 2  — FileInputLine: Clear and delegation (requirements 4-5)
//   Section 3  — FileDialog: construction and layout (requirements 6-8)
//   Section 4  — FileDialog: HandleEvent command handling (requirements 9, 16-18)
//   Section 5  — FileDialog: Valid method (requirements 10-14)
//   Section 6  — FileDialog: detectInput classification (requirement 15)
//   Section 7  — Falsifying: FileInputLine robustness (requirements 19-20)
//   Section 8  — Falsifying: FileDialog robustness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// =============================================================================
// Helper: newFileDialogTest creates a FileDialog for testing using a temp dir.
// =============================================================================

func newFileDialogTest(t *testing.T, flags FileDialogFlag) *FileDialog {
	t.Helper()
	dir := t.TempDir()
	return NewFileDialogInDir(dir, "*", "Test Dialog", flags)
}

// =============================================================================
// Helper: buttonWithTitle finds a Button child whose Title() contains substr.
// Returns nil if not found.
// =============================================================================

func buttonWithTitle(fd *FileDialog, substr string) *Button {
	for _, child := range fd.Children() {
		if btn, ok := child.(*Button); ok {
			if strings.Contains(btn.Title(), substr) {
				return btn
			}
		}
	}
	return nil
}

// =============================================================================
// Helper: newFileInputLine creates a FileInputLine for unit testing.
// =============================================================================

func newFileInputLine() *FileInputLine {
	return &FileInputLine{
		InputLine: NewInputLine(NewRect(1, 1, 32, 1), 256),
		wildcard:  "*",
	}
}

// =============================================================================
// Section 1 — FileInputLine: autofill on CmFileFocused (requirements 1-3)
// =============================================================================

// TestFileInputLine_Autofill_FileOnCmFileFocused verifies requirement 1:
// "When the user navigates the FileList (which broadcasts CmFileFocused), the
// FileInputLine auto-fills with the selected filename — but only when it does NOT
// have focus."
//
// Spec: "If EvBroadcast + CmFileFocused AND NOT focused (HasState(SfSelected) == false):
// ... If file: SetText(name)"
func TestFileInputLine_Autofill_FileOnCmFileFocused(t *testing.T) {
	fl := newFileInputLine()
	fl.SetState(SfSelected, false) // NOT focused

	entry := &FileEntry{Name: "hello.txt", IsDir: false, Path: "/tmp/hello.txt"}
	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: entry}

	fl.HandleEvent(ev)

	if fl.Text() != "hello.txt" {
		t.Errorf("Text() = %q, want %q (autofill with filename)", fl.Text(), "hello.txt")
	}
	if !ev.IsCleared() {
		t.Error("event must be cleared after autofill — no further processing")
	}
}

// TestFileInputLine_Autofill_DirectoryAppendsWildcard verifies requirement 2:
// "If directory: SetText(name + '/' + wildcard)"
//
// Spec: the wildcard is appended after a '/' separator to the directory name.
func TestFileInputLine_Autofill_DirectoryAppendsWildcard(t *testing.T) {
	fl := &FileInputLine{
		InputLine: NewInputLine(NewRect(1, 1, 32, 1), 256),
		wildcard:  "*.go",
	}
	fl.SetState(SfSelected, false)

	entry := &FileEntry{Name: "src", IsDir: true, Path: "/tmp/src"}
	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: entry}

	fl.HandleEvent(ev)

	expected := "src/*.go"
	if fl.Text() != expected {
		t.Errorf("Text() = %q, want %q (directory name + '/' + wildcard)", fl.Text(), expected)
	}
	if !ev.IsCleared() {
		t.Error("event must be cleared after autofill")
	}
}

// TestFileInputLine_Autofill_IgnoresWhenFocused verifies requirement 3:
// "but only when it does NOT have focus"
//
// Spec: "If EvBroadcast + CmFileFocused AND NOT focused ...
// Otherwise: delegate to InputLine.HandleEvent"
func TestFileInputLine_Autofill_IgnoresWhenFocused(t *testing.T) {
	fl := newFileInputLine()
	fl.SetText("existing text")
	fl.SetState(SfSelected, true) // IS focused — autofill must NOT happen

	entry := &FileEntry{Name: "should_not_use.txt", IsDir: false, Path: "/tmp/should_not_use.txt"}
	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: entry}

	fl.HandleEvent(ev)

	if fl.Text() != "existing text" {
		t.Errorf("Text() = %q, want %q (must not autofill when focused — "+
			"delegates to InputLine which ignores broadcasts)",
			fl.Text(), "existing text")
	}
}

// =============================================================================
// Section 2 — FileInputLine: Clear and delegation (requirements 4-5)
// =============================================================================

// TestFileInputLine_Clear_EmptiesText verifies requirement 4:
// "Clear() — sets text to '' via SetText('')"
func TestFileInputLine_Clear_EmptiesText(t *testing.T) {
	fl := newFileInputLine()
	fl.SetText("some content")

	fl.Clear()

	if fl.Text() != "" {
		t.Errorf("Text() = %q, want %q after Clear()", fl.Text(), "")
	}
}

// TestFileInputLine_NonBroadcastDelegatesToInputLine verifies requirement 5:
// "Otherwise: delegate to InputLine.HandleEvent"
//
// Non-broadcast events (e.g., keyboard rune) should be passed through to the
// embedded InputLine for normal processing.
func TestFileInputLine_NonBroadcastDelegatesToInputLine(t *testing.T) {
	fl := newFileInputLine()
	fl.SetState(SfSelected, true) // must be focused for InputLine to process chars

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'X'}}
	fl.HandleEvent(ev)

	if fl.Text() != "X" {
		t.Errorf("Text() = %q, want %q (keyboard event must be delegated to InputLine "+
			"which processes the rune)", fl.Text(), "X")
	}
}

// TestFileInputLine_NonBroadcastStillDelegatesEvenWhenNotFocused verifies that
// non-broadcast events are delegated to InputLine even when the FileInputLine
// is not selected (the InputLine itself decides whether to act on it).
// Spec: "Otherwise: delegate to InputLine.HandleEvent" — always, regardless of focus.
func TestFileInputLine_NonBroadcastStillDelegatesEvenWhenNotFocused(t *testing.T) {
	fl := newFileInputLine()
	fl.SetState(SfSelected, false) // NOT focused
	fl.SetText("before")

	// Send EvCommand — this should be delegated to InputLine
	ev := &Event{What: EvCommand, Command: CmOK}
	fl.HandleEvent(ev)

	// The event should reach InputLine (it won't be cleaned up by FileInputLine).
	// InputLine won't interact with it, but no panic should occur.
	// This test mainly verifies delegation rather than cleaning up the event.
	if ev.What == EvNothing {
		// If event was cleared it implies FileInputLine handled it (which would
		// be wrong for non-broadcast, non-CmFileFocused events).
		t.Error("non-broadcast EvCommand should be delegated to InputLine, not cleared by FileInputLine")
	}
}

// =============================================================================
// Section 3 — FileDialog: construction and layout (requirements 6-8)
// =============================================================================

// TestFileDialog_New_HasTitleChildrenFileNameEmpty verifies requirement 6:
// "FileDialog creation: has title, has children (>= 6), FileName() is empty initially"
func TestFileDialog_New_HasTitleChildrenFileNameEmpty(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	if fd.Title() != "Test Dialog" {
		t.Errorf("Title() = %q, want %q", fd.Title(), "Test Dialog")
	}

	n := len(fd.Children())
	if n < 6 {
		t.Errorf("Children() = %d, want >= 6 (labels, input, history, file list, info pane, buttons)", n)
	}

	if fd.FileName() != "" {
		t.Errorf("FileName() = %q, want %q initially", fd.FileName(), "")
	}
}

// TestFileDialog_New_FdOpenButtonShowsOpen verifies requirement 7:
// "FileDialog flag detection: FdOpenButton shows 'Open' button"
//
// Spec: "From flags: Open (CmFileOpen)" at FdOpenButton.
func TestFileDialog_New_FdOpenButtonShowsOpen(t *testing.T) {
	fd := newFileDialogTest(t, FdOpenButton)

	btn := buttonWithTitle(fd, "Open")
	if btn == nil {
		t.Error("FdOpenButton flag should create an 'Open' button, but none found")
	}
}

// TestFileDialog_New_NoOpenButtonWithoutFlag verifies the Open button is NOT
// present when FdOpenButton is not set (falsifying: implementation that always
// adds an Open button).
func TestFileDialog_New_NoOpenButtonWithoutFlag(t *testing.T) {
	fd := newFileDialogTest(t, 0) // no FdOpenButton

	btn := buttonWithTitle(fd, "Open")
	if btn != nil {
		t.Error("Open button should NOT be present without FdOpenButton flag")
	}
}

// TestFileDialog_New_NoFlagsDefaultsToOK verifies requirement 8:
// "FileDialog default OK button when no flags"
//
// Spec: "If no action button flags specified, adds 'OK' (CmOK) button by default"
func TestFileDialog_New_NoFlagsDefaultsToOK(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	btn := buttonWithTitle(fd, "OK")
	if btn == nil {
		t.Error("no action button flags: should default to an 'OK' button, but none found")
	}
}

// TestFileDialog_New_CancelButtonAlwaysPresent verifies Cancel is always
// appended at the end.
// Spec: "Cancel (CmCancel) always appended at the end"
func TestFileDialog_New_CancelButtonAlwaysPresent(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	btn := buttonWithTitle(fd, "Cancel")
	if btn == nil {
		t.Error("Cancel button should always be present")
	}

	// Cancel should be the LAST button (appended at the end).
	// Find the last Button child.
	var lastButton *Button
	for _, child := range fd.Children() {
		if b, ok := child.(*Button); ok {
			lastButton = b
		}
	}
	if lastButton == nil {
		t.Fatal("no buttons found in dialog")
	}
	if !strings.Contains(lastButton.Title(), "Cancel") {
		t.Errorf("last button title = %q, want Cancel (should be appended at end)",
			lastButton.Title())
	}
}

// TestFileDialog_New_WildcardDefault verifies that NewFileDialog defaults "" to "*".
// Spec: "If wildcard is '', defaults to '*'"
func TestFileDialog_New_WildcardDefault(t *testing.T) {
	dir := t.TempDir()
	fd := NewFileDialogInDir(dir, "", "Test Dialog", 0)

	if fd.fileInput.wildcard != "*" {
		t.Errorf("fileInput.wildcard = %q, want %q (default)", fd.fileInput.wildcard, "*")
	}
	if fd.fileInput.Text() != "*" {
		t.Errorf("fileInput.Text() = %q, want %q (initialized with default wildcard)", fd.fileInput.Text(), "*")
	}
}

// TestFileDialog_New_FileInputInitialTextIsWildcard verifies the FileInputLine
// starts with the wildcard text as its content.
// Spec: "FileInputLine at (1, 1, 32, 1) initialized with wildcard text"
func TestFileDialog_New_FileInputInitialTextIsWildcard(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	if fd.fileInput.Text() != "*" {
		t.Errorf("fileInput.Text() = %q, want %q (initialized with wildcard)", fd.fileInput.Text(), "*")
	}
}

// TestFileDialog_New_FirstButtonIsDefault verifies the first button gets WithDefault().
// Spec: "First button gets WithDefault()"
func TestFileDialog_New_FirstButtonIsDefault(t *testing.T) {
	fd := newFileDialogTest(t, FdOpenButton | FdOKButton)

	// Find first button child
	var firstButton *Button
	for _, child := range fd.Children() {
		if b, ok := child.(*Button); ok {
			if firstButton == nil {
				firstButton = b
			}
		}
	}
	if firstButton == nil {
		t.Fatal("no buttons found")
	}

	// The Open button should be first (from FdOpenButton flag)
	if !strings.Contains(firstButton.Title(), "Open") {
		t.Errorf("first button title = %q, want Open (first flag button)", firstButton.Title())
	}

	// Verify the first button is the default.
	if !firstButton.IsDefault() {
		t.Error("first button should have WithDefault() applied (IsDefault() should be true)")
	}
}

// =============================================================================
// Section 4 — FileDialog: HandleEvent command handling (requirements 9, 16-18)
// =============================================================================

// TestFileDialog_HandleEvent_CmFileClearClearsInput verifies requirement 9:
// "FileDialog Clear button clears filename"
//
// Spec: "CmFileClear: calls fileInput.Clear(), sets resultPath to '', clears event"
func TestFileDialog_HandleEvent_CmFileClearClearsInput(t *testing.T) {
	fd := newFileDialogTest(t, FdClearButton)
	fd.fileInput.SetText("something.txt")
	fd.resultPath = "/some/path" // non-empty to verify it is cleared

	ev := &Event{What: EvCommand, Command: CmFileClear}
	fd.HandleEvent(ev)

	if fd.fileInput.Text() != "" {
		t.Errorf("fileInput.Text() = %q, want %q after CmFileClear",
			fd.fileInput.Text(), "")
	}
	if fd.FileName() != "" {
		t.Errorf("FileName() = %q, want %q after CmFileClear (resultPath should be cleared)",
			fd.FileName(), "")
	}
	if !ev.IsCleared() {
		t.Error("CmFileClear event must be cleared (no further processing)")
	}
}

// TestFileDialog_HandleEvent_CmFileOpenRemapsToCmOK verifies requirement 16:
// "FileDialog HandleEvent: CmFileOpen remaps to CmOK"
//
// Spec: "CmFileOpen: remaps to CmOK (process as OK)"
func TestFileDialog_HandleEvent_CmFileOpenRemapsToCmOK(t *testing.T) {
	fd := newFileDialogTest(t, FdOpenButton)

	ev := &Event{What: EvCommand, Command: CmFileOpen}
	fd.HandleEvent(ev)

	if ev.Command != CmOK {
		t.Errorf("CmFileOpen: event.Command = %v, want CmOK (should be remapped)", ev.Command)
	}
	// The event should NOT be cleared — it continues through dispatch as CmOK.
	if ev.IsCleared() {
		t.Error("CmFileOpen remapped to CmOK: event must NOT be cleared (continues dispatch)")
	}
}

// TestFileDialog_HandleEvent_CmFileReplaceRemapsToCmOK verifies requirement 17:
// "FileDialog HandleEvent: CmFileReplace remaps to CmOK"
//
// Spec: "CmFileReplace: remaps to CmOK"
func TestFileDialog_HandleEvent_CmFileReplaceRemapsToCmOK(t *testing.T) {
	fd := newFileDialogTest(t, FdReplaceButton)

	ev := &Event{What: EvCommand, Command: CmFileReplace}
	fd.HandleEvent(ev)

	if ev.Command != CmOK {
		t.Errorf("CmFileReplace: event.Command = %v, want CmOK (should be remapped)", ev.Command)
	}
	if ev.IsCleared() {
		t.Error("CmFileReplace remapped to CmOK: event must NOT be cleared")
	}
}

// TestFileDialog_HandleEvent_DelegatesToDialog verifies requirement 18 indirectly:
// non-EvCommand events are delegated to Dialog.HandleEvent.
//
// Spec: "Delegates to Dialog.HandleEvent for everything else"
func TestFileDialog_HandleEvent_DelegatesToDialog(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	// Send an EvBroadcast — not CmFileClear, CmFileOpen, or CmFileReplace.
	// Should be delegated to Dialog.HandleEvent, which routes through Group.
	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: &FileEntry{Name: "test.txt", Path: "/tmp/test.txt"}}
	fd.HandleEvent(ev)

	// The event should remain unmodified (or routed through Group dispatch).
	// No panic should occur.
	// The key assertion: the event.What should not change to EvNothing
	// (which would indicate FileDialog consumed it when it shouldn't have).
	if ev.IsCleared() {
		t.Error("non-command broadcast event should not be cleared by FileDialog (delegated to Dialog)")
	}
}

// =============================================================================
// Section 5 — FileDialog: Valid method (requirements 10-14)
// =============================================================================

// TestFileDialog_Valid_CmCancelReturnsTrue verifies requirement 10:
// "FileDialog Valid(CmCancel) returns true"
//
// Spec: "CmCancel: always returns true (allow close)"
func TestFileDialog_Valid_CmCancelReturnsTrue(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	if !fd.Valid(CmCancel) {
		t.Error("Valid(CmCancel) should always return true")
	}
}

// TestFileDialog_Valid_NonCmOKReturnsTrue verifies that non-CmOK, non-Cancel
// commands return true (pass through).
// Spec: "For non-CmOK commands: returns true"
func TestFileDialog_Valid_NonCmOKReturnsTrue(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	// CmFileOpen is NOT CmOK, so it should return true (allow close).
	if !fd.Valid(CmFileOpen) {
		t.Error("Valid(non-CmOK, non-Cancel) should return true")
	}
}

// TestFileDialog_Valid_CmOKEmptyTextReturnsFalse verifies requirement 11:
// "FileDialog Valid(CmOK) with empty text returns false"
//
// Spec: "Trims whitespace from fileInput.Text() — If empty: returns false"
func TestFileDialog_Valid_CmOKEmptyTextReturnsFalse(t *testing.T) {
	fd := newFileDialogTest(t, 0)
	fd.fileInput.SetText("")

	if fd.Valid(CmOK) {
		t.Error("Valid(CmOK) with empty text should return false (can't close)")
	}
}

// TestFileDialog_Valid_CmOKWhitespaceOnlyReturnsFalse verifies whitespace trimming.
// Spec: "Trims whitespace from fileInput.Text() — If empty: returns false"
func TestFileDialog_Valid_CmOKWhitespaceOnlyReturnsFalse(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	for _, text := range []string{"   ", "\t", "  \t  "} {
		fd.fileInput.SetText(text)
		if fd.Valid(CmOK) {
			t.Errorf("Valid(CmOK) with whitespace-only text %q should return false", text)
		}
	}
}

// TestFileDialog_Valid_CmOKFilenameResolvesPath verifies requirement 12:
// "FileDialog Valid(CmOK) with filename resolves path and returns true"
//
// Spec: "inputFilename: resolves to absolute path (if relative, joins with
// FileList's current dir). Stores in resultPath. Returns true."
func TestFileDialog_Valid_CmOKFilenameResolvesPath(t *testing.T) {
	fd := newFileDialogTest(t, 0)
	initialDir := fd.fileList.Dir()

	fd.fileInput.SetText("hello.txt")

	if !fd.Valid(CmOK) {
		t.Error("Valid(CmOK) with a plain filename should return true")
	}

	expected := filepath.Join(initialDir, "hello.txt")
	if fd.FileName() != expected {
		t.Errorf("FileName() = %q, want %q (relative path resolved against fileList dir)",
			fd.FileName(), expected)
	}
}

// TestFileDialog_Valid_CmOKAbsolutePathFilename verifies that an absolute path
// is stored as-is without joining the current directory.
// Spec: "resolves to absolute path (if relative, joins with FileList's current dir)"
func TestFileDialog_Valid_CmOKAbsolutePathFilename(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	absPath := "/absolute/path/to/file.txt"
	fd.fileInput.SetText(absPath)

	if !fd.Valid(CmOK) {
		t.Error("Valid(CmOK) with an absolute path should return true")
	}

	if fd.FileName() != absPath {
		t.Errorf("FileName() = %q, want %q (absolute path stored as-is)", fd.FileName(), absPath)
	}
}

// TestFileDialog_Valid_CmOKWildcardUpdatesFilter verifies requirement 13:
// "FileDialog Valid(CmOK) with wildcard updates filter, returns false"
//
// Spec: "inputWildcard: updates wildcard, fileInput.wildcard, refreshes FileList
// with new wildcard. Broadcasts CmFileFilter to owner's children. Returns false."
func TestFileDialog_Valid_CmOKWildcardUpdatesFilter(t *testing.T) {
	fd := newFileDialogTest(t, 0)
	fd.fileInput.SetText("*.go")

	if fd.Valid(CmOK) {
		t.Error("Valid(CmOK) with wildcard text should return false (keep dialog open)")
	}

	if fd.wildcard != "*.go" {
		t.Errorf("fd.wildcard = %q, want %q", fd.wildcard, "*.go")
	}
	if fd.fileInput.wildcard != "*.go" {
		t.Errorf("fileInput.wildcard = %q, want %q", fd.fileInput.wildcard, "*.go")
	}
	if fd.fileList.Wildcard() != "*.go" {
		t.Errorf("fileList.Wildcard() = %q, want %q", fd.fileList.Wildcard(), "*.go")
	}
	// FileList data source should have been refreshed with the new wildcard.
	// (The data source changed, but we trust ReadDirectory was called.)
}

// TestFileDialog_Valid_CmOKDirectoryNavigates verifies requirement 14:
// "FileDialog Valid(CmOK) with directory navigates, returns false"
//
// Spec: "inputDirectory: navigates FileList to the directory, updates
// fileInput.wildcard. Returns false."
func TestFileDialog_Valid_CmOKDirectoryNavigates(t *testing.T) {
	baseDir := t.TempDir()
	subdir := filepath.Join(baseDir, "mydir")
	os.MkdirAll(subdir, 0755)

	fd := NewFileDialogInDir(baseDir, "*", "Test", 0)
	fd.fileInput.SetText(subdir)

	if fd.Valid(CmOK) {
		t.Error("Valid(CmOK) with existing directory should return false (keep dialog open)")
	}

	if fd.fileList.Dir() != subdir {
		t.Errorf("fileList.Dir() = %q, want %q (should navigate to the directory)",
			fd.fileList.Dir(), subdir)
	}

	// fileInput.wildcard should be updated to match
	if fd.fileInput.wildcard != "*" {
		t.Errorf("fileInput.wildcard = %q after directory navigation, want %q",
			fd.fileInput.wildcard, "*")
	}
}

// TestFileDialog_Valid_CmOKDirectoryUpdatesFileInputWildcard verifies the
// fileInput's wildcard is updated to the dialog's wildcard after directory
// navigation (not reset to something else).
// Spec: "updates fileInput.wildcard"
func TestFileDialog_Valid_CmOKDirectoryUpdatesFileInputWildcard(t *testing.T) {
	baseDir := t.TempDir()
	subdir := filepath.Join(baseDir, "docs")
	os.MkdirAll(subdir, 0755)

	fd := NewFileDialogInDir(baseDir, "*.txt", "Test", 0)
	fd.fileInput.SetText(subdir)

	fd.Valid(CmOK)

	// fileInput.wildcard must be updated to the current dialog wildcard
	if fd.fileInput.wildcard != "*.txt" {
		t.Errorf("fileInput.wildcard = %q, want %q after directory navigation",
			fd.fileInput.wildcard, "*.txt")
	}
}

// =============================================================================
// Section 6 — FileDialog: detectInput classification (requirement 15)
// =============================================================================

// TestFileDialog_DetectInput_WildcardDetection verifies requirement 15:
// "If text contains '*' or '?': inputWildcard"
//
// detectInput is unexported; we test it indirectly through Valid(CmOK).
func TestFileDialog_DetectInput_WildcardDetection(t *testing.T) {
	wildcardInputs := []string{
		"*",
		"*.go",
		"file*.txt",
		"test?.txt",
		"?test",
		"a?b*",
	}

	for _, input := range wildcardInputs {
		t.Run(input, func(t *testing.T) {
			fd := newFileDialogTest(t, 0)
			fd.fileInput.SetText(input)

			result := fd.Valid(CmOK)

			// Wildcard inputs should return false (keep dialog open)
			if result {
				t.Errorf("text %q contains wildcard char: Valid(CmOK) should return false (wildcard detected), got true", input)
			}
			// The dialog's wildcard should be updated to the input
			if fd.wildcard != input {
				t.Errorf("text %q: fd.wildcard = %q, want %q", input, fd.wildcard, input)
			}
			// FileName should remain empty (not a filename)
			if fd.FileName() != "" {
				t.Errorf("text %q: FileName() = %q, want %q (wildcard, not filename)",
					input, fd.FileName(), "")
			}
		})
	}
}

// TestFileDialog_DetectInput_DirectoryDetection verifies requirement 15:
// "If os.Stat(text) succeeds and IsDir(): inputDirectory"
//
// Tested indirectly through Valid(CmOK).
func TestFileDialog_DetectInput_DirectoryDetection(t *testing.T) {
	baseDir := t.TempDir()
	subdir := filepath.Join(baseDir, "projects")
	os.MkdirAll(subdir, 0755)

	fd := NewFileDialogInDir(baseDir, "*", "Test", 0)
	originalDir := fd.fileList.Dir()

	fd.fileInput.SetText(subdir)
	result := fd.Valid(CmOK)

	if result {
		t.Error("existing directory should be detected as inputDirectory: Valid(CmOK) should return false")
	}
	if fd.fileList.Dir() != subdir {
		t.Errorf("directory not navigated: fileList.Dir() = %q, want %q", fd.fileList.Dir(), subdir)
	}
	// FileName() should still be empty (not a filename)
	if fd.FileName() != "" {
		t.Errorf("FileName() = %q, want %q (directory, not filename)", fd.FileName(), "")
	}

	_ = originalDir
}

// TestFileDialog_DetectInput_FilenameDetection verifies requirement 15:
// "Otherwise: inputFilename"
//
// A plain string (no wildcards, not a directory) should be classified as a filename.
func TestFileDialog_DetectInput_FilenameDetection(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	fd.fileInput.SetText("report.pdf")
	result := fd.Valid(CmOK)

	if !result {
		t.Error("plain filename without wildcards should be detected as inputFilename: Valid(CmOK) should return true")
	}
	if fd.FileName() != filepath.Join(fd.fileList.Dir(), "report.pdf") {
		t.Errorf("FileName() = %q, want resolved absolute path", fd.FileName())
	}
}

// TestFileDialog_DetectInput_WildcardBeforeDirectory verifies that detectInput
// checks wildcards BEFORE directories. A directory named "src*" (if it existed)
// should still be treated as wildcard because '*' takes priority.
// Spec: "If text contains '*' or '?': inputWildcard" (checked first)
func TestFileDialog_DetectInput_WildcardBeforeDirectory(t *testing.T) {
	baseDir := t.TempDir()
	// Create a directory — but the input contains '*', so it's a wildcard.
	os.MkdirAll(filepath.Join(baseDir, "realdir"), 0755)

	fd := NewFileDialogInDir(baseDir, "*", "Test", 0)
	fd.fileInput.SetText("realdir*") // contains '*' so it's a wildcard, not a directory

	result := fd.Valid(CmOK)

	if result {
		t.Error("text with '*' should be detected as wildcard, not filename or directory")
	}
	// Should have been treated as wildcard, not directory navigation
	if fd.wildcard != "realdir*" {
		t.Errorf("wildcard = %q, want %q (wildcard takes priority over directory)", fd.wildcard, "realdir*")
	}
}

// =============================================================================
// Section 7 — Falsifying: FileInputLine robustness (requirements 19-20)
// =============================================================================

// TestFileInputLine_Falsifying_NilInfoDoesNotPanic verifies requirement 19:
// "Falsifying: FileInputLine doesn't panic on nil Info"
//
// If event.Info is nil, the type assertion to *FileEntry would fail. A correct
// implementation checks for nil before extracting.
func TestFileInputLine_Falsifying_NilInfoDoesNotPanic(t *testing.T) {
	fl := newFileInputLine()
	fl.SetState(SfSelected, false)

	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: nil}

	// Must not panic.
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("HandleEvent panicked with nil Info: %v", r)
			}
		}()
		fl.HandleEvent(ev)
	}()
}

// TestFileInputLine_Falsifying_WrongInfoTypeDoesNotPanic verifies requirement 20:
// "Falsifying: FileInputLine doesn't panic on wrong Info type"
//
// If event.Info is not a *FileEntry, the type assertion would fail. A correct
// implementation uses the comma-ok form to guard against this.
func TestFileInputLine_Falsifying_WrongInfoTypeDoesNotPanic(t *testing.T) {
	fl := newFileInputLine()
	fl.SetState(SfSelected, false)

	// Info is a string instead of *FileEntry.
	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: "not a FileEntry"}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("HandleEvent panicked with wrong Info type (string): %v", r)
			}
		}()
		fl.HandleEvent(ev)
	}()

	// Text should be unchanged (autofill didn't apply).
	if fl.Text() != "" {
		t.Errorf("Text() = %q, want %q (wrong Info type: autofill should not apply)", fl.Text(), "")
	}
}

// TestFileInputLine_Falsifying_WrongInfoTypeIntDoesNotPanic verifies that
// non-*FileEntry Info of different types (e.g., int) also doesn't panic.
func TestFileInputLine_Falsifying_WrongInfoTypeIntDoesNotPanic(t *testing.T) {
	fl := newFileInputLine()
	fl.SetState(SfSelected, false)

	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: 42}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("HandleEvent panicked with wrong Info type (int): %v", r)
			}
		}()
		fl.HandleEvent(ev)
	}()

	if fl.Text() != "" {
		t.Errorf("Text() = %q, want %q (non-*FileEntry Info: autofill should not apply)", fl.Text(), "")
	}
}

// =============================================================================
// Section 8 — Falsifying: FileDialog robustness
// =============================================================================

// TestFileDialog_Falsifying_ValidWildcardDoesNotPanicNoOwner verifies that
// Valid(CmOK) with wildcard text does not panic when the dialog has no owner
// (the CmFileFilter broadcast should silently no-op).
// Spec: "Broadcasts CmFileFilter to owner's children"
func TestFileDialog_Falsifying_ValidWildcardDoesNotPanicNoOwner(t *testing.T) {
	fd := newFileDialogTest(t, 0)
	fd.fileInput.SetText("*.test")

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Valid(CmOK) with wildcard panicked (no owner): %v", r)
			}
		}()
		fd.Valid(CmOK)
	}()
}

// TestFileDialog_Falsifying_ValidDirectoryDoesNotPanicNoOwner verifies
// Valid(CmOK) with directory text doesn't panic when there's no owner.
func TestFileDialog_Falsifying_ValidDirectoryDoesNotPanicNoOwner(t *testing.T) {
	baseDir := t.TempDir()
	subdir := filepath.Join(baseDir, "subdir")
	os.MkdirAll(subdir, 0755)

	fd := NewFileDialogInDir(baseDir, "*", "Test", 0)
	fd.fileInput.SetText(subdir)

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Valid(CmOK) with directory panicked (no owner): %v", r)
			}
		}()
		fd.Valid(CmOK)
	}()
}

// TestFileDialog_Falsifying_MultipleCallsToValidDoNotAccumulate verifies that
// calling Valid(CmOK) multiple times doesn't accumulate state incorrectly.
// Each call should independently evaluate the current text.
func TestFileDialog_Falsifying_MultipleCallsToValidDoNotAccumulate(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	// First call: wildcard → false
	fd.fileInput.SetText("*.md")
	if fd.Valid(CmOK) {
		t.Error("first call with wildcard should return false")
	}
	firstWildcard := fd.wildcard
	if firstWildcard != "*.md" {
		t.Errorf("first call: wildcard = %q, want '*.md'", firstWildcard)
	}

	// Second call: filename → true (should not remember wildcard state)
	fd.fileInput.SetText("readme.md")
	if !fd.Valid(CmOK) {
		t.Error("second call with filename should return true")
	}
	if fd.wildcard != "*.md" {
		t.Errorf("second call should not overwrite wildcard with filename: wildcard = %q, want '*.md'", fd.wildcard)
	}
}

// TestFileDialog_Falsifying_FileInputIsCorrectType verifies that fd.fileInput
// is the FileInputLine (not the bare InputLine inside it).
// Spec: "fileInput is the FileInputLine (not the InputLine inside it)"
func TestFileDialog_Falsifying_FileInputIsCorrectType(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	// The fileInput field must be a *FileInputLine (not *InputLine).
	// This is a compile-time check implicitly, but we verify its methods work:
	// FileInputLine has a wildcard field and Clear() method.
	if fd.fileInput.wildcard == "" {
		t.Error("fileInput.wildcard should be non-empty")
	}

	// fileInput.Clear() should work (FileInputLine method, not InputLine method)
	// InputLine doesn't have a Clear() method.
	fd.fileInput.SetText("data")
	fd.fileInput.Clear()
	if fd.fileInput.Text() != "" {
		t.Error("fileInput.Clear() should empty the text")
	}
}

// TestFileDialog_Falsifying_CmFileFilterBroadcastedOnWildcard verifies that
// when the dialog has an owner, CmFileFilter is broadcast to owner's children
// on wildcard detection.
// Spec: "Broadcasts CmFileFilter to owner's children"
func TestFileDialog_Falsifying_CmFileFilterBroadcastedOnWildcard(t *testing.T) {
	fd := newFileDialogTest(t, 0)
	spy := newNonSelectableSpyView()

	// Spy must be a sibling of FileDialog (inserted into FileDialog's owner),
	// not a child of FileDialog. Broadcasts go to owner's children.
	g := NewGroup(NewRect(0, 0, 80, 24))
	g.Insert(fd)
	g.Insert(spy)

	fd.fileInput.SetText("*.go")
	resetBroadcasts(spy)

	fd.Valid(CmOK)

	found := false
	for _, rec := range spy.broadcasts {
		if rec.command == CmFileFilter {
			found = true
			break
		}
	}
	if !found {
		t.Error("Valid(CmOK) with wildcard should broadcast CmFileFilter to owner's children (siblings)")
	}
}

// TestFileDialog_Falsifying_InitialSelectionIsIndexZero verifies the initial
// selection in the FileList is index 0.
// Spec: "Initial selection is index 0 in the FileList"
func TestFileDialog_Falsifying_InitialSelectionIsIndexZero(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	if fd.fileList.Selected() != 0 {
		t.Errorf("fileList.Selected() = %d, want 0 (initial selection)", fd.fileList.Selected())
	}
}

// TestFileDialog_Falsifying_AllChildTypesPresent verifies the complete set of
// expected child widget types are present in the dialog.
// Spec: labels, FileInputLine, History, FileList, FileInfoPane, buttons
func TestFileDialog_Falsifying_AllChildTypesPresent(t *testing.T) {
	fd := newFileDialogTest(t, 0)

	hasLabel := false
	hasFileInput := false
	hasHistory := false
	hasFileList := false
	hasFileInfo := false
	hasButton := false

	for _, child := range fd.Children() {
		switch child.(type) {
		case *Label:
			hasLabel = true
		case *FileInputLine:
			hasFileInput = true
		case *History:
			hasHistory = true
		case *FileList:
			hasFileList = true
		case *FileInfoPane:
			hasFileInfo = true
		case *Button:
			hasButton = true
		}
	}

	if !hasLabel {
		t.Error("missing Label child")
	}
	if !hasFileInput {
		t.Error("missing FileInputLine child")
	}
	if !hasHistory {
		t.Error("missing History child")
	}
	if !hasFileList {
		t.Error("missing FileList child")
	}
	if !hasFileInfo {
		t.Error("missing FileInfoPane child")
	}
	if !hasButton {
		t.Error("missing Button child")
	}
}
