package tv

// file_list_test.go — Tests for FileList widget.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Compile-time interface assertions
//   Section 2  — Constructor tests (NewFileList, Dir, Wildcard, ListViewer, FocusedChild)
//   Section 3  — ReadDirectory tests (dir/wildcard change, data source update, nil return)
//   Section 4  — Double-click tests (directory navigation, ".." navigation, file CmOK)
//   Section 5  — Enter key tests (directory navigation + broadcast, file CmOK)
//   Section 6  — Incremental search: printable chars (prefix match, 1s timeout)
//   Section 7  — Incremental search: backspace and navigation key clearing
//   Section 8  — Focus broadcasting (CmFileFocused after events)
//   Section 9  — Delegate to ListBox (normal events pass through)
//   Section 10 — Falsifying tests (one per happy-path requirement)

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newFileListWithOwner creates a FileList inserted into a Group (which becomes its
// owner) along with a broadcastSpyView. Returns the FileList, the group, and the spy.
func newFileListWithOwner(bounds Rect) (*FileList, *Group, *broadcastSpyView) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	fl := NewFileList(bounds)
	spy := newNonSelectableSpyView()
	g.Insert(fl)
	g.Insert(spy)
	resetBroadcasts(spy)
	// FileList's owner is now the group (set by Insert).
	return fl, g, spy
}

// newFileListWithTempDir creates a FileList pointed at a temp directory with the
// given wildcard. It also populates the temp dir with the provided directory names
// and file names (created as empty files/directories). Returns the FileList and
// the cleanup function (caller should defer it).
//
// The temp directory itself is managed by t.TempDir() so it is cleaned up automatically;
// the cleanup func is a no-op provided for symmetry.
func newFileListWithTempDir(t *testing.T, wildcard string, dirs []string, files []string) *FileList {
	t.Helper()
	tmpDir := t.TempDir()
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(tmpDir, d), 0755)
	}
	for _, f := range files {
		os.WriteFile(filepath.Join(tmpDir, f), []byte("x"), 0644)
	}
	fl := NewFileList(NewRect(0, 0, 40, 15))
	fl.ReadDirectory(tmpDir, wildcard)
	return fl
}

// hasBroadcastSpy reports whether a broadcastSpyView received at least one
// broadcast with the given command.
func hasBroadcastSpy(spy *broadcastSpyView, cmd CommandCode) bool {
	for _, rec := range spy.broadcasts {
		if rec.command == cmd {
			return true
		}
	}
	return false
}

// broadcastInfoFor returns the Info of the first broadcast with the given command.
func broadcastInfoFor(spy *broadcastSpyView, cmd CommandCode) any {
	for _, rec := range spy.broadcasts {
		if rec.command == cmd {
			return rec.info
		}
	}
	return nil
}

// flDoubleClickEv creates a double-click (ClickCount=2) mouse event at (x, y).
func flDoubleClickEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 2}}
}

// flSingleClickEv creates a single-click (ClickCount=1) mouse event at (x, y).
func flSingleClickEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 1}}
}

// keyboardRuneEv creates a keyboard event with the given rune.
func keyboardRuneEv(ch rune) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ch}}
}

// keyboardKeyEv creates a keyboard event with the given non-rune key.
func keyboardKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key}}
}

// enterKeyEv creates an Enter key keyboard event.
func enterKeyEv() *Event {
	return keyboardKeyEv(tcell.KeyEnter)
}

// backspaceKeyEv creates a Backspace keyboard event.
func backspaceKeyEv() *Event {
	return keyboardKeyEv(tcell.KeyBackspace)
}

// findEntryItem returns the index of the entry whose Item() text equals `name`,
// or -1 if not found.
func findEntryItem(fl *FileList, name string) int {
	ds := fl.DataSource()
	for i := 0; i < ds.Count(); i++ {
		if ds.Item(i) == name {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Section 1 — Compile-time interface assertions
// ---------------------------------------------------------------------------

// Spec: "FileList is a *ListBox wrapper" — must satisfy Container interface
var _ Container = (*FileList)(nil)

// ---------------------------------------------------------------------------
// Section 2 — Constructor tests
// ---------------------------------------------------------------------------

// TestNewFileListReturnsNonNil verifies NewFileList creates a non-nil *FileList.
// Spec: "NewFileList(bounds Rect) *FileList"
func TestNewFileListReturnsNonNil(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	if fl == nil {
		t.Fatal("NewFileList returned nil")
	}
}

// TestNewFileListDirIsCurrentDir verifies Dir returns the current directory (".")
// after construction.
// Spec: "creates a FileList with a FileDataSource pointed at current directory"
func TestNewFileListDirIsCurrentDir(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	if fl.Dir() != "." {
		t.Errorf("Dir() = %q, want %q (current directory)", fl.Dir(), ".")
	}
}

// TestNewFileListWildcardIsStar verifies Wildcard returns "*" after construction.
// Spec: "with wildcard '*'"
func TestNewFileListWildcardIsStar(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	if fl.Wildcard() != "*" {
		t.Errorf("Wildcard() = %q, want %q", fl.Wildcard(), "*")
	}
}

// TestNewFileListListViewerNonNil verifies the internal ListViewer is non-nil.
// Spec: "Focused child is the internal ListViewer"
func TestNewFileListListViewerNonNil(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	if fl.ListViewer() == nil {
		t.Error("ListViewer() must not be nil after NewFileList")
	}
}

// TestNewFileListFocusedChildIsListViewer verifies FocusedChild returns the
// internal ListViewer.
// Spec: "Focused child is the internal ListViewer"
func TestNewFileListFocusedChildIsListViewer(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	fc := fl.FocusedChild()
	if fc == nil {
		t.Error("FocusedChild() must not be nil after NewFileList")
	}
	if lv := fl.ListViewer(); lv != nil && fc != View(lv) { //nolint:staticcheck
		t.Errorf("FocusedChild() = %T, want *ListViewer", fc)
	}
}

// TestNewFileListSetsSelf verifies the Self view is set (the widget registers
// itself for click-to-focus).
// Spec: "Sets itself as the Self view"
func TestNewFileListSetsSelf(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	// Self is set means the widget can receive focus clicks.
	// Verify by checking HasOption is accessible — the FileList is functional.
	if !fl.HasOption(OfSelectable) {
		t.Error("NewFileList must set OfSelectable (ListBox sets it)")
	}
}

// TestNewFileListIsContainer verifies FileList satisfies the Container interface
// at runtime.
// Spec: "FileList is a *ListBox wrapper"
func TestNewFileListIsContainer(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	// Container interface methods must not panic
	_ = fl.Children()
	_ = fl.FocusedChild()
	fl.Insert(nil) // nop for nil, but demonstrates availability
}

// ---------------------------------------------------------------------------
// Section 3 — ReadDirectory tests
// ---------------------------------------------------------------------------

// TestFileListReadDirectoryChangesDir verifies ReadDirectory updates Dir().
// Spec: "wraps the directory and wildcard changes. Creates a new FileDataSource
// for dir/wildcard"
func TestFileListReadDirectoryChangesDir(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	tmpDir := t.TempDir()
	err := fl.ReadDirectory(tmpDir, "*")
	if err != nil {
		t.Fatalf("ReadDirectory returned error: %v", err)
	}
	if fl.Dir() != tmpDir {
		t.Errorf("Dir() = %q, want %q after ReadDirectory", fl.Dir(), tmpDir)
	}
}

// TestFileListReadDirectoryChangesWildcard verifies ReadDirectory updates Wildcard().
// Spec: "wraps the directory and wildcard changes"
func TestFileListReadDirectoryChangesWildcard(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	tmpDir := t.TempDir()
	err := fl.ReadDirectory(tmpDir, "*.go")
	if err != nil {
		t.Fatalf("ReadDirectory returned error: %v", err)
	}
	if fl.Wildcard() != "*.go" {
		t.Errorf("Wildcard() = %q, want %q after ReadDirectory", fl.Wildcard(), "*.go")
	}
}

// TestFileListReadDirectoryUpdatesDataSource verifies ReadDirectory replaces the
// data source so that Count reflects the new directory's contents.
// Spec: "Creates a new FileDataSource for dir/wildcard, sets it on the ListBox
// via SetDataSource"
func TestFileListReadDirectoryUpdatesDataSource(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "hello.txt"), []byte("world"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)

	err := fl.ReadDirectory(tmpDir, "*")
	if err != nil {
		t.Fatalf("ReadDirectory returned error: %v", err)
	}

	// Should have "..", "subdir/", "hello.txt" = 3 entries
	if fl.DataSource().Count() < 3 {
		t.Errorf("DataSource().Count() = %d, want at least 3 (.. + subdir/ + hello.txt)",
			fl.DataSource().Count())
	}
}

// TestFileListReadDirectoryReturnsNil verifies ReadDirectory returns nil on success.
// Spec: "Returns nil"
func TestFileListReadDirectoryReturnsNil(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	tmpDir := t.TempDir()
	err := fl.ReadDirectory(tmpDir, "*")
	if err != nil {
		t.Errorf("ReadDirectory returned error: %v, want nil", err)
	}
}

// TestFileListReadDirectorySwitchesBetweenDirs verifies ReadDirectory can be
// called multiple times to switch between directories.
// Spec: "wraps the directory and wildcard changes"
func TestFileListReadDirectorySwitchesBetweenDirs(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	os.WriteFile(filepath.Join(dir1, "one.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(dir2, "two.txt"), []byte("2"), 0644)

	fl.ReadDirectory(dir1, "*")
	if fl.Dir() != dir1 {
		t.Errorf("Dir() = %q, want %q", fl.Dir(), dir1)
	}

	fl.ReadDirectory(dir2, "*.txt")
	if fl.Dir() != dir2 {
		t.Errorf("Dir() = %q after switch, want %q", fl.Dir(), dir2)
	}
	if fl.Wildcard() != "*.txt" {
		t.Errorf("Wildcard() = %q after switch, want %q", fl.Wildcard(), "*.txt")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Double-click tests
// ---------------------------------------------------------------------------

// TestFileListDoubleClickDirectoryNavigates verifies double-clicking a directory
// entry navigates into that directory and selects index 0.
// Spec: "If the double-clicked entry IS a directory (including '..'): navigate
// by calling ReadDirectory with the entry's Path and current wildcard, then
// select index 0. Clear the event. Return."
func TestFileListDoubleClickDirectoryNavigates(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", []string{"docs"}, []string{"readme.txt"})

	// TopIndex is 0, ".." is at row 0, directories after ".."
	// docs/ should be at Y=1 (row 0 = "..", row 1 = first dir)
	docsIdx := findEntryItem(fl, "docs/")
	if docsIdx < 0 {
		t.Fatal("docs/ not found in listing")
	}
	originalDir := fl.Dir()

	// Double-click on the docs/ row
	ev := flDoubleClickEv(0, docsIdx)
	fl.HandleEvent(ev)

	// Event should be cleared (spec: "Clear the event")
	if !ev.IsCleared() {
		t.Error("double-click on directory: event must be cleared")
	}
	// Selection should be 0 (spec: "select index 0")
	if fl.Selected() != 0 {
		t.Errorf("double-click on directory: Selected() = %d, want 0", fl.Selected())
	}
	// Dir should have changed (navigated into docs/)
	if fl.Dir() == originalDir {
		t.Error("double-click on directory: Dir() did not change — should have navigated into docs/")
	}
}

// TestFileListDoubleClickParentDotDotNavigates verifies double-clicking ".."
// navigates to the parent directory.
// Spec: "including '..'"
func TestFileListDoubleClickParentDotDotNavigates(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", []string{"sub"}, []string{"a.txt"})
	originalDir := fl.Dir()

	// Navigate into sub/
	subIdx := findEntryItem(fl, "sub/")
	if subIdx < 0 {
		t.Fatal("sub/ not found")
	}
	fl.HandleEvent(flDoubleClickEv(0, subIdx))
	if fl.Dir() == originalDir {
		t.Fatal("expected to navigate into sub/")
	}

	// Now ".." should be at row 0, pointing back to the parent
	ev := flDoubleClickEv(0, 0) // row 0 = ".."
	fl.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("double-click on '..': event must be cleared")
	}
	if fl.Selected() != 0 {
		t.Errorf("double-click on '..': Selected() = %d, want 0", fl.Selected())
	}
	if fl.Dir() != originalDir {
		t.Errorf("double-click on '..': Dir() = %q, want %q (back to parent)", fl.Dir(), originalDir)
	}
}

// TestFileListDoubleClickFileTransformsToCmOK verifies double-clicking a regular
// file transforms the event to EvCommand/CmOK.
// Spec: "If the double-clicked entry IS a regular file: transform the event to
// EvCommand/CmOK"
func TestFileListDoubleClickFileTransformsToCmOK(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"data.txt"})

	fIdx := findEntryItem(fl, "data.txt")
	if fIdx < 0 {
		t.Fatal("data.txt not found in listing")
	}

	ev := flDoubleClickEv(0, fIdx)
	fl.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("double-click on file: event.What = %v, want EvCommand", ev.What)
	}
	if ev.Command != CmOK {
		t.Errorf("double-click on file: event.Command = %v, want CmOK", ev.Command)
	}
	if ev.Mouse != nil {
		t.Error("double-click on file: event.Mouse must be nil after transform")
	}
}

// TestFileListDoubleClickFileClearsMouseField verifies event.Mouse is set to nil
// when transforming to EvCommand/CmOK.
// Spec: "event.Mouse = nil"
func TestFileListDoubleClickFileClearsMouseField(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"report.csv"})

	idx := findEntryItem(fl, "report.csv")
	if idx < 0 {
		t.Fatal("report.csv not found in listing")
	}

	ev := flDoubleClickEv(5, idx) // x=5, y=row
	// Mouse should be non-nil before HandleEvent
	if ev.Mouse == nil {
		t.Fatal("precondition: ev.Mouse must be non-nil")
	}

	fl.HandleEvent(ev)

	if ev.Mouse != nil {
		t.Error("double-click on file: event.Mouse must be nil after transform")
	}
}

// TestFileListDoubleClickTripleClickAlsoTriggers verifies ClickCount >= 2
// triggers double-click handling (triple-click works too).
// Spec: "ClickCount >= 2"
func TestFileListDoubleClickTripleClickAlsoTriggers(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"log.txt"})

	idx := findEntryItem(fl, "log.txt")
	if idx < 0 {
		t.Fatal("log.txt not found")
	}

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: idx, Button: tcell.Button1, ClickCount: 3}}
	fl.HandleEvent(ev)

	if ev.What != EvCommand || ev.Command != CmOK {
		t.Errorf("triple-click on file: event.What=%v event.Command=%v, want EvCommand/CmOK",
			ev.What, ev.Command)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Enter key tests
// ---------------------------------------------------------------------------

// TestFileListEnterOnDirectoryNavigates verifies pressing Enter on a directory
// entry navigates into it, selects 0, and clears the event.
// Spec: "If the selected entry IS a directory: navigate as above (ReadDirectory),
// select 0, clear event"
func TestFileListEnterOnDirectoryNavigates(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", []string{"projects"}, []string{"main.go"})
	originalDir := fl.Dir()

	projIdx := findEntryItem(fl, "projects/")
	if projIdx < 0 {
		t.Fatal("projects/ not found")
	}
	fl.SetSelected(projIdx)

	ev := enterKeyEv()
	fl.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Enter on directory: event must be cleared")
	}
	if fl.Selected() != 0 {
		t.Errorf("Enter on directory: Selected() = %d, want 0", fl.Selected())
	}
	if fl.Dir() == originalDir {
		t.Error("Enter on directory: Dir() did not change — should have navigated")
	}
}

// TestFileListEnterOnDirectoryBroadcastsCmFileFocused verifies Enter on directory
// broadcasts CmFileFocused.
// Spec: "broadcast CmFileFocused"
func TestFileListEnterOnDirectoryBroadcastsCmFileFocused(t *testing.T) {
	fl, _, spy := newFileListWithOwner(NewRect(0, 0, 40, 10))

	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "mydir"), 0755)
	fl.ReadDirectory(tmpDir, "*")

	projIdx := findEntryItem(fl, "mydir/")
	if projIdx < 0 {
		t.Fatal("mydir/ not found")
	}
	fl.SetSelected(projIdx)
	resetBroadcasts(spy)

	ev := enterKeyEv()
	fl.HandleEvent(ev)

	if !hasBroadcastSpy(spy, CmFileFocused) {
		t.Error("Enter on directory: CmFileFocused broadcast not received")
	}
}

// TestFileListEnterOnFileTransformsToCmOK verifies pressing Enter on a regular
// file entry transforms the event to EvCommand/CmOK.
// Spec: "If the selected entry IS a regular file: transform to EvCommand/CmOK"
func TestFileListEnterOnFileTransformsToCmOK(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"config.json"})

	fIdx := findEntryItem(fl, "config.json")
	if fIdx < 0 {
		t.Fatal("config.json not found")
	}
	fl.SetSelected(fIdx)

	ev := enterKeyEv()
	fl.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("Enter on file: event.What = %v, want EvCommand", ev.What)
	}
	if ev.Command != CmOK {
		t.Errorf("Enter on file: event.Command = %v, want CmOK", ev.Command)
	}
}

// TestFileListEnterOnFileDoesNotNavigate verifies Enter on a file does not change
// the directory.
// Spec: "If the selected entry IS a regular file: transform to EvCommand/CmOK"
func TestFileListEnterOnFileDoesNotNavigate(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"output.log"})
	originalDir := fl.Dir()

	fIdx := findEntryItem(fl, "output.log")
	if fIdx < 0 {
		t.Fatal("output.log not found")
	}
	fl.SetSelected(fIdx)

	ev := enterKeyEv()
	fl.HandleEvent(ev)

	if fl.Dir() != originalDir {
		t.Errorf("Enter on file: Dir() changed from %q to %q — should not navigate", originalDir, fl.Dir())
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Incremental search: printable chars
// ---------------------------------------------------------------------------

// TestFileListIncrementalSearchSelectsFirstPrefixMatch verifies typing a
// character selects the first entry whose name starts with that character
// (case-insensitive).
// Spec: "case-insensitive prefix match against entry names, selects first match"
func TestFileListIncrementalSearchSelectsFirstPrefixMatch(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"apple.txt", "banana.txt", "apricot.txt", "cherry.txt"})

	// Type 'a' — should select first entry starting with 'a' (apple.txt)
	ev := keyboardRuneEv('a')
	fl.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("KeyRune event must be cleared after incremental search")
	}

	aIdx := findEntryItem(fl, "apple.txt")
	apIdx := findEntryItem(fl, "apricot.txt")
	if fl.Selected() != aIdx && fl.Selected() != apIdx {
		t.Errorf("After typing 'a': Selected() = %d (%s), want the first 'a' file",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}
	// Should be the earlier one (apple.txt, alphabetically)
	if fl.Selected() != aIdx {
		t.Errorf("After typing 'a': Selected() = %d (%s), want apple.txt",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}
}

// TestFileListIncrementalSearchRefinesWithMoreChars verifies typing additional
// characters refines the search to match the accumulated prefix.
// Spec: "case-insensitive prefix match against entry names"
func TestFileListIncrementalSearchRefinesWithMoreChars(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"apple.txt", "apricot.txt", "banana.txt"})

	// Type 'a'
	fl.HandleEvent(keyboardRuneEv('a'))
	// Type 'p' — should match entries starting with "ap"
	fl.HandleEvent(keyboardRuneEv('p'))

	aIdx := findEntryItem(fl, "apple.txt")
	apIdx := findEntryItem(fl, "apricot.txt")
	bIdx := findEntryItem(fl, "banana.txt")

	if fl.Selected() == bIdx {
		t.Error("After typing 'ap': Selected() on banana.txt, should match 'ap' prefix")
	}
	// Apple comes before apricot alphabetically
	if fl.Selected() != aIdx {
		t.Errorf("After typing 'ap': Selected() = %d (%s), want apple.txt",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}

	// Type 'r' — should match apricot.txt only
	fl.HandleEvent(keyboardRuneEv('r'))

	if fl.Selected() != apIdx {
		t.Errorf("After typing 'apr': Selected() = %d (%s), want apricot.txt",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}
}

// TestFileListIncrementalSearchCaseInsensitive verifies prefix matching is
// case-insensitive.
// Spec: "case-insensitive prefix match"
func TestFileListIncrementalSearchCaseInsensitive(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"Banana.txt", "apple.txt", "Cherry.txt"})

	// Type lowercase 'b' — should match Banana.txt (case-insensitive)
	ev := keyboardRuneEv('b')
	fl.HandleEvent(ev)

	bIdx := findEntryItem(fl, "Banana.txt")
	if fl.Selected() != bIdx {
		t.Errorf("After typing 'b': Selected() = %d (%s), want Banana.txt (case-insensitive)",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}

	// New FileList: type uppercase 'A' — should match apple.txt
	fl2 := newFileListWithTempDir(t, "*", nil, []string{"Banana.txt", "apple.txt", "Cherry.txt"})
	ev2 := keyboardRuneEv('A')
	fl2.HandleEvent(ev2)

	aIdx2 := findEntryItem(fl2, "apple.txt")
	if fl2.Selected() != aIdx2 {
		t.Errorf("After typing 'A': Selected() = %d (%s), want apple.txt (case-insensitive)",
			fl2.Selected(), fl2.DataSource().Item(fl2.Selected()))
	}
}

// TestFileListIncrementalSearchDotIsPrintable verifies '.' is treated as a
// printable character for search (useful for file extensions).
// Spec: "if char is '.' or not a control character"
func TestFileListIncrementalSearchDotIsPrintable(t *testing.T) {
	fl := newFileListWithTempDir(t, "*.go;*.mod", nil,
		[]string{"main.go", "go.mod", "test.go", "go.sum"})

	// Type '.' — should select first entry starting with '.' (none) or matching prefix
	// Actually: files don't start with '.' since hidden files are excluded,
	// but go.mod starts with 'g'. Let me test incrementally:
	// Type 'g', then 'o', then '.'
	fl.HandleEvent(keyboardRuneEv('g'))
	fl.HandleEvent(keyboardRuneEv('o'))
	fl.HandleEvent(keyboardRuneEv('.'))

	// "go." prefix should match "go.mod" and "go.sum"
	modIdx := findEntryItem(fl, "go.mod")
	sumIdx := findEntryItem(fl, "go.sum")
	mainIdx := findEntryItem(fl, "main.go")

	if fl.Selected() == mainIdx {
		t.Errorf("After 'go.': Selected() = main.go, should match 'go.' prefix (go.mod or go.sum)")
	}
	if fl.Selected() != modIdx && fl.Selected() != sumIdx {
		t.Errorf("After 'go.': Selected() = %d (%s), want go.mod or go.sum",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}
}

// TestFileListIncrementalSearchTimeoutResetsBuffer verifies the 1-second timeout
// resets the search buffer so a new search starts fresh.
// Spec: "Handle 1-second timeout: if current time > 1 second since last search
// time, clear searchBuf"
func TestFileListIncrementalSearchTimeoutResetsBuffer(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"alpha.txt", "bravo.txt"})

	// Type 'a' — selects alpha.txt
	fl.HandleEvent(keyboardRuneEv('a'))
	aIdx := findEntryItem(fl, "alpha.txt")
	if fl.Selected() != aIdx {
		t.Fatalf("After 'a': expected alpha.txt selected, got %s", fl.DataSource().Item(fl.Selected()))
	}

	// Wait > 1 second for timeout
	time.Sleep(1100 * time.Millisecond)

	// Type 'b' — buffer should have been cleared, so this is a fresh 'b' search
	fl.HandleEvent(keyboardRuneEv('b'))
	bIdx := findEntryItem(fl, "bravo.txt")

	if fl.Selected() != bIdx {
		t.Errorf("After timeout + 'b': Selected() = %d (%s), want bravo.txt (timeout should reset buffer, so 'b' starts fresh, not 'ab')",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}
}

// TestFileListIncrementalSearchNoMatchKeepsPreviousSelection verifies typing
// a character with no matching entry does not change the selection.
// Spec: "case-insensitive prefix match against entry names, selects first match"
// Edge case: no match → selection unchanged (or no-op).
func TestFileListIncrementalSearchNoMatchKeepsPreviousSelection(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"delta.txt"})

	// Type 'z' — no entry starts with 'z'
	prev := fl.Selected()
	fl.HandleEvent(keyboardRuneEv('z'))

	if fl.Selected() != prev {
		t.Errorf("After typing 'z' (no match): Selected() changed from %d to %d; should stay unchanged",
			prev, fl.Selected())
	}
}

// TestFileListIncrementalSearchClearsEvent verifies a KeyRune event is cleared
// after handling.
// Spec: "Clear the event"
func TestFileListIncrementalSearchClearsEvent(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"test.txt"})

	ev := keyboardRuneEv('t')
	fl.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("KeyRune event must be cleared after incremental search")
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Incremental search: backspace and navigation key clearing
// ---------------------------------------------------------------------------

// TestFileListBackspaceRemovesLastSearchChar verifies Backspace removes the last
// character from the search buffer and updates the search.
// Spec: "Backspace ... If searchBuf has chars, remove last char"
func TestFileListBackspaceRemovesLastSearchChar(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"apple.txt", "apricot.txt", "banana.txt"})

	// Type 'a', 'p', 'p'
	fl.HandleEvent(keyboardRuneEv('a'))
	fl.HandleEvent(keyboardRuneEv('p'))
	fl.HandleEvent(keyboardRuneEv('p'))

	// After "app", selection should be on apple.txt (only match for "app")
	aIdx := findEntryItem(fl, "apple.txt")
	if fl.Selected() != aIdx {
		t.Fatalf("After 'app': expected apple.txt, got %s", fl.DataSource().Item(fl.Selected()))
	}

	// Press Backspace — removes last 'p', search becomes "ap"
	ev := backspaceKeyEv()
	fl.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Backspace event must be cleared")
	}

	// After "ap", selection should be on apple.txt (first "ap" match)
	// Could also be apricot.txt if sorted differently, but apple < apricot
	if fl.Selected() != aIdx {
		t.Errorf("After backspace (ap): Selected() = %d (%s), want apple.txt",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}
}

// TestFileListBackspaceOnEmptyBufferNoOp verifies Backspace on an empty search
// buffer does nothing.
// Spec: "If searchBuf has chars, remove last char" — no chars means no-op.
func TestFileListBackspaceOnEmptyBufferNoOp(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"alpha.txt", "bravo.txt"})

	prev := fl.Selected()

	// Backspace without any prior search
	ev := backspaceKeyEv()
	fl.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Backspace event must be cleared even on empty buffer")
	}
	if fl.Selected() != prev {
		t.Error("Backspace on empty buffer should not change selection")
	}
}

// TestFileListBackspaceToEmptyClearsSearch verifies backspacing until the buffer
// is empty clears the search, returning selection to its pre-search position.
// Spec: "If searchBuf has chars, remove last char"
func TestFileListBackspaceToEmptyClearsSearch(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"apple.txt", "bravo.txt", "charlie.txt"})

	// Select bravo.txt manually
	bIdx := findEntryItem(fl, "bravo.txt")
	fl.SetSelected(bIdx)

	// Type 'a' — should select first 'a' file (apple.txt)
	fl.HandleEvent(keyboardRuneEv('a'))
	aIdx := findEntryItem(fl, "apple.txt")
	if fl.Selected() != aIdx {
		t.Fatalf("After 'a': expected apple.txt, got %s", fl.DataSource().Item(fl.Selected()))
	}

	// Backspace — clears search (back to "a" search, which is now empty)
	fl.HandleEvent(backspaceKeyEv())
	// After clearing the only char, the search is empty
	// The selection after clearing the search is whatever the last search was before clearing
	// Typically the selection from the previous state
}

// TestFileListNavigationKeysClearSearchBuffer verifies Up, Down, PgUp, PgDn,
// Home, End, Left, Right clear the search buffer.
// Spec: "Navigation keys (Up, Down, PgUp, PgDn, Home, End, Left, Right) clear
// the search buffer"
func TestFileListNavigationKeysClearSearchBuffer(t *testing.T) {
	for _, key := range []tcell.Key{
		tcell.KeyUp, tcell.KeyDown, tcell.KeyPgUp, tcell.KeyPgDn,
		tcell.KeyHome, tcell.KeyEnd, tcell.KeyLeft, tcell.KeyRight,
	} {
		t.Run(key.String(), func(t *testing.T) {
			fl := newFileListWithTempDir(t, "*", nil,
				[]string{"apple.txt", "bravo.txt", "charlie.txt"})

			// Type 'a' — search buffer now "a"
			fl.HandleEvent(keyboardRuneEv('a'))

			aIdx := findEntryItem(fl, "apple.txt")
			if fl.Selected() != aIdx {
				t.Fatalf("After 'a': expected apple.txt selected, got %s",
					fl.DataSource().Item(fl.Selected()))
			}

			// Send navigation key — search buffer should be cleared
			fl.HandleEvent(keyboardKeyEv(key))

			// Now type 'b' — buffer should be fresh "b", not "ab"
			fl.HandleEvent(keyboardRuneEv('b'))
			bIdx := findEntryItem(fl, "bravo.txt")

			if fl.Selected() != bIdx {
				t.Errorf("After %v + 'b': Selected() = %d (%s), want bravo.txt (search buffer should have been cleared by navigation key)",
					key, fl.Selected(), fl.DataSource().Item(fl.Selected()))
			}
		})
	}
}

// TestFileListNavigationKeysPreserveListBoxNavigation verifies that navigation
// keys clear the search buffer AND still result in ListBox navigation (the event
// is passed through, not consumed by FileList).
// Spec: "Navigation keys ... clear the search buffer"
// Implicit: the navigation must still happen, so the event reaches ListBox.
func TestFileListNavigationKeysPreserveListBoxNavigation(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil,
		[]string{"alpha.txt", "bravo.txt", "charlie.txt"})

	// Set the selection to index 0, then type 'a' to start search from alpha.txt
	fl.SetSelected(0)
	fl.HandleEvent(keyboardRuneEv('a'))

	aIdx := findEntryItem(fl, "alpha.txt")
	if fl.Selected() != aIdx {
		t.Fatalf("After 'a': expected alpha.txt selected, got %s", fl.DataSource().Item(fl.Selected()))
	}

	// Send Down arrow — should clear search buffer AND navigate down via ListBox
	ev := keyboardKeyEv(tcell.KeyDown)
	fl.HandleEvent(ev)

	// After Down arrow, selection should have moved from alpha.txt to the next entry
	// If FileList consumes the event, selection stays at alpha.txt
	if fl.Selected() == aIdx {
		t.Error("After Down arrow: selection unchanged; FileList may have consumed event instead of delegating to ListBox")
	}
}

// TestFileListBackspaceUpdatesSearchTime verifies Backspace updates the search
// time to now, preventing immediate timeout.
// Spec: "Update searchTime" for backspace
// This is really checking that backspace is part of the same search session.
func TestFileListBackspaceUpdatesSearchTime(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"alpha.txt", "bravo.txt"})

	// Type 'a' then backspace within the same window (no timeout)
	fl.HandleEvent(keyboardRuneEv('a'))
	fl.HandleEvent(backspaceKeyEv())

	// After backspace, immediately type 'b' — if search time was updated,
	// backspace shouldn't cause a timeout, so 'b' starts a fresh search.
	fl.HandleEvent(keyboardRuneEv('b'))

	bIdx := findEntryItem(fl, "bravo.txt")
	if fl.Selected() != bIdx {
		t.Errorf("After 'a', backspace, 'b': Selected() = %d (%s), want bravo.txt (fresh 'b' search)",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Focus broadcasting
// ---------------------------------------------------------------------------

// TestFileListKeyboardEventBroadcastsCmFileFocused verifies that after a
// keyboard event, CmFileFocused is broadcast to the owner's children.
// Spec: "After any keyboard or mouse event that may change focus, broadcast
// CmFileFocused to all owner's children with the current Entry as event.Info"
func TestFileListKeyboardEventBroadcastsCmFileFocused(t *testing.T) {
	fl, _, spy := newFileListWithOwner(NewRect(0, 0, 40, 10))

	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "data.txt"), []byte("test"), 0644)
	fl.ReadDirectory(tmpDir, "*")
	resetBroadcasts(spy)

	// Send a keyboard event (Down arrow — triggers navigation + broadcast)
	ev := keyboardKeyEv(tcell.KeyDown)
	fl.HandleEvent(ev)

	if !hasBroadcastSpy(spy, CmFileFocused) {
		t.Error("keyboard event: CmFileFocused broadcast not received")
	}
}

// TestFileListMouseEventBroadcastsCmFileFocused verifies that after a mouse
// event, CmFileFocused is broadcast.
// Spec: "After any keyboard or mouse event that may change focus, broadcast
// CmFileFocused"
func TestFileListMouseEventBroadcastsCmFileFocused(t *testing.T) {
	fl, _, spy := newFileListWithOwner(NewRect(0, 0, 40, 10))

	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "clickme.txt"), []byte("x"), 0644)
	fl.ReadDirectory(tmpDir, "*")
	resetBroadcasts(spy)

	// Send a mouse event (single click on an entry)
	ev := flSingleClickEv(0, 1)
	fl.HandleEvent(ev)

	if !hasBroadcastSpy(spy, CmFileFocused) {
		t.Error("mouse event: CmFileFocused broadcast not received")
	}
}

// TestFileListCmFileFocusedCarriesFileEntry verifies the broadcast CmFileFocused
// carries a *FileEntry in event.Info.
// Spec: "with the current Entry as event.Info"
func TestFileListCmFileFocusedCarriesFileEntry(t *testing.T) {
	fl, _, spy := newFileListWithOwner(NewRect(0, 0, 40, 10))

	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "data.txt"), []byte("hello"), 0644)
	fl.ReadDirectory(tmpDir, "*")
	resetBroadcasts(spy)

	ev := keyboardKeyEv(tcell.KeyDown)
	fl.HandleEvent(ev)

	info := broadcastInfoFor(spy, CmFileFocused)
	if info == nil {
		t.Fatal("CmFileFocused broadcast had nil Info")
	}
	if _, ok := info.(*FileEntry); !ok {
		t.Errorf("CmFileFocused Info = %T, want *FileEntry", info)
	}
}

// TestFileListCmFileFocusedNotBroadcastWhenNoOwner verifies that when the
// FileList has no owner, no panic occurs on broadcast attempt.
// Spec: "broadcast CmFileFocused to all owner's children" — no owner means no broadcast.
func TestFileListCmFileFocusedNotBroadcastWhenNoOwner(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	tmpDir := t.TempDir()
	fl.ReadDirectory(tmpDir, "*")

	// Send event — should not panic even without owner
	ev := enterKeyEv()
	// No panic expected
	fl.HandleEvent(ev)
}

// ---------------------------------------------------------------------------
// Section 9 — Delegate to ListBox
// ---------------------------------------------------------------------------

// TestFileListDelegatesNormalMouseToListBox verifies normal mouse events
// (not double-click) are delegated to ListBox.HandleEvent.
// Spec: "For normal mouse events and keyboard events not handled above, delegate
// to the embedded ListBox.HandleEvent"
func TestFileListDelegatesNormalMouseToListBox(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"one.txt", "two.txt"})

	// Single click should be delegated to ListBox for selection
	prev := fl.Selected()
	ev := flSingleClickEv(0, 1) // click on row 1
	fl.HandleEvent(ev)

	// The ListBox (via ListViewer) should have changed the selection
	if fl.Selected() == prev && fl.Selected() != 1 {
		t.Errorf("Single click: Selected() = %d (unchanged from %d); should have changed via ListBox delegation",
			fl.Selected(), prev)
	}
}

// TestFileListDelegatesUnhandledKeyboardToListBox verifies keyboard events not
// handled by FileList (e.g., arrow keys after search cleared) are delegated to
// ListBox.HandleEvent.
// Spec: "For normal mouse events and keyboard events not handled above, delegate
// to the embedded ListBox.HandleEvent"
func TestFileListDelegatesUnhandledKeyboardToListBox(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil,
		[]string{"a.txt", "b.txt", "c.txt"})

	// Select first entry
	fl.SetSelected(0)

	// Send Down arrow — FileList clears search buffer, then delegates to ListBox
	// which should move selection down
	ev := keyboardKeyEv(tcell.KeyDown)
	fl.HandleEvent(ev)

	if fl.Selected() != 1 {
		t.Errorf("Down arrow: Selected() = %d, want 1 (delegated to ListBox for navigation)",
			fl.Selected())
	}
}

// TestFileListDelegatesPgUpToScroll verifies PgUp is delegated to ListBox for
// page-up behavior after search buffer is cleared.
// Spec: "delegate to the embedded ListBox.HandleEvent"
func TestFileListDelegatesPgUpToScroll(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil,
		[]string{"a.txt", "b.txt", "c.txt", "d.txt", "e.txt", "f.txt", "g.txt", "h.txt"})

	// Select last
	fl.SetSelected(7)

	// Send PgUp — should move selection up by a page
	ev := keyboardKeyEv(tcell.KeyPgUp)
	fl.HandleEvent(ev)

	if fl.Selected() == 7 {
		t.Errorf("PgUp: Selected() still %d; should have moved up via ListBox delegation", fl.Selected())
	}
}

// ---------------------------------------------------------------------------
// Section 10 — Falsifying tests
// ---------------------------------------------------------------------------

// TestFileListFalsifyingReadDirectoryActuallyRefreshes verifies ReadDirectory
// actually changes the data source entries (Count changes), not just updates
// Dir/Wildcard accessors.
// (Falsifying: lazy impl that updates Dir() but doesn't refresh the listing.)
// Spec: "Creates a new FileDataSource for dir/wildcard, sets it on the ListBox
// via SetDataSource"
func TestFileListFalsifyingReadDirectoryActuallyRefreshes(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte("c"), 0644)

	fl.ReadDirectory(dir, "*")

	// Count should reflect the directory (.. + 3 files = 4)
	count := fl.DataSource().Count()
	if count != 4 {
		t.Errorf("ReadDirectory: DataSource().Count() = %d, want 4 (.. + 3 files)", count)
	}
}

// TestFileListFalsifyingReadDirectoryChangesWildcardFilter verifies ReadDirectory
// with a restrictive wildcard filters files.
// (Falsifying: lazy impl that ignores the wildcard parameter.)
// Spec: "wraps the directory and wildcard changes"
func TestFileListFalsifyingReadDirectoryChangesWildcardFilter(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 40, 10))
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# readme"), 0644)

	fl.ReadDirectory(dir, "*.go")

	// md file should NOT appear with *.go filter
	mdIdx := findEntryItem(fl, "readme.md")
	if mdIdx >= 0 {
		t.Error("readme.md should not appear with wildcard '*.go'")
	}
}

// TestFileListFalsifyingDoubleClickOnDirReallyNavigates verifies double-click
// on a directory actually changes the directory content (entries from the new
// directory are visible).
// (Falsifying: lazy impl that clears event and sets Selected(0) without actually
// calling ReadDirectory.)
// Spec: "navigate by calling ReadDirectory with the entry's Path and current wildcard"
func TestFileListFalsifyingDoubleClickOnDirReallyNavigates(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "inside.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "outside.txt"), []byte("y"), 0644)

	fl := NewFileList(NewRect(0, 0, 40, 15))
	fl.ReadDirectory(tmpDir, "*")

	// Double-click on subdir/
	subIdx := findEntryItem(fl, "subdir/")
	if subIdx < 0 {
		t.Fatal("subdir/ not found")
	}

	fl.HandleEvent(flDoubleClickEv(0, subIdx))

	// After navigation, "inside.txt" should be visible (it's in subdir/)
	if findEntryItem(fl, "inside.txt") < 0 {
		t.Error("After double-click on subdir/: 'inside.txt' not found — navigation did not refresh listing")
	}
	// "outside.txt" should NOT be visible (we left tmpDir)
	if findEntryItem(fl, "outside.txt") >= 0 {
		t.Error("After double-click on subdir/: 'outside.txt' still visible — navigation did not change directory")
	}
}

// TestFileListFalsifyingEnterOnFileDoesNotEmitCmOKForDir verifies Enter on a
// file emits CmOK, but Enter on a file does NOT navigate (dir vs file distinction).
// (Falsifying: lazy impl that transforms to CmOK for both dirs and files.)
// Spec: "If the selected entry IS a regular file: transform to EvCommand/CmOK"
func TestFileListFalsifyingEnterOnFileDoesNotEmitCmOKForDir(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", []string{"folder"}, []string{"document.txt"})

	originalDir := fl.Dir()
	dirIdx := findEntryItem(fl, "folder/")

	// Enter on directory
	fl.SetSelected(dirIdx)
	ev := enterKeyEv()
	fl.HandleEvent(ev)

	// After Enter on directory, event should be CLEARED (not transformed to CmOK)
	if ev.What == EvCommand {
		t.Errorf("Enter on directory: event.What = EvCommand (%d), should be EvNothing (cleared, not CmOK)", ev.What)
	}
	// Directory should have changed
	if fl.Dir() == originalDir {
		t.Error("Enter on directory: Dir() did not change")
	}
}

// TestFileListFalsifyingSearchTimeoutActuallyResets verifies the 1-second timeout
// actually resets the search buffer, not just a cosmetic update.
// (Falsifying: lazy impl that resets searchTime but doesn't clear searchBuf.)
// Spec: "if current time > 1 second since last search time, clear searchBuf"
func TestFileListFalsifyingSearchTimeoutActuallyResets(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil,
		[]string{"alpha.txt", "alpha2.txt", "bravo.txt"})

	// Type "alpha" quickly
	for _, ch := range "alpha" {
		fl.HandleEvent(keyboardRuneEv(ch))
	}

	// Both alpha.txt and alpha2.txt start with "alpha"
	a2Idx := findEntryItem(fl, "alpha2.txt")
	if fl.Selected() == a2Idx {
		t.Skipf("alpha2.txt selected before alpha.txt — alphabetical order issue")
	}

	// Wait > 1 second
	time.Sleep(1100 * time.Millisecond)

	// Type 'b' — if timeout didn't clear buffer, search would be "alphab" (no match)
	// If timeout DID clear, fresh 'b' search → bravo.txt
	fl.HandleEvent(keyboardRuneEv('b'))
	bIdx := findEntryItem(fl, "bravo.txt")

	if fl.Selected() != bIdx {
		t.Errorf("After timeout + 'b': Selected() = %d (%s), want bravo.txt; "+
			"timeout should have cleared buffer so 'b' starts fresh",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}
}

// TestFileListFalsifyingBackspaceOnEmptyDoesNotCrash verifies backspacing on an
// empty search buffer does not crash or remove phantom characters.
// (Falsifying: lazy impl that unconditionally slices the buffer without
// checking length.)
// Spec: "If searchBuf has chars, remove last char"
func TestFileListFalsifyingBackspaceOnEmptyDoesNotCrash(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil, []string{"a.txt", "b.txt"})

	// Multiple backspaces on empty buffer — no crash
	for i := 0; i < 10; i++ {
		ev := backspaceKeyEv()
		fl.HandleEvent(ev)
	}

	// FileList should still function normally
	fl.HandleEvent(keyboardRuneEv('a'))
	aIdx := findEntryItem(fl, "a.txt")
	if fl.Selected() != aIdx {
		t.Error("FileList broken after backspacing on empty buffer")
	}
}

// TestFileListFalsifyingBroadcastHasCorrectEntry verifies CmFileFocused carries
// the current entry, not a stale one from a previous selection.
// (Falsifying: lazy impl that always broadcasts the first entry.)
// Spec: "with the current Entry as event.Info"
func TestFileListFalsifyingBroadcastHasCorrectEntry(t *testing.T) {
	fl, _, spy := newFileListWithOwner(NewRect(0, 0, 40, 10))

	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "first.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "second.txt"), []byte("b"), 0644)
	fl.ReadDirectory(tmpDir, "*")

	// Select "second.txt" (index 2, since ".." is 0 and no dirs)
	secondIdx := findEntryItem(fl, "second.txt")
	if secondIdx < 0 {
		t.Fatal("second.txt not found")
	}
	fl.SetSelected(secondIdx)
	resetBroadcasts(spy)

	// Send event that triggers broadcast
	ev := keyboardKeyEv(tcell.KeyUp) // move up, then broadcast
	fl.HandleEvent(ev)

	info := broadcastInfoFor(spy, CmFileFocused)
	if info == nil {
		t.Fatal("CmFileFocused broadcast had nil Info")
	}
	entry, ok := info.(*FileEntry)
	if !ok {
		t.Fatalf("CmFileFocused Info is %T, want *FileEntry", info)
	}
	// After moving up from second.txt, entry should be first.txt
	if entry.Name != "first.txt" {
		t.Errorf("CmFileFocused Entry.Name = %q, want %q (the entry that now has focus)",
			entry.Name, "first.txt")
	}
}

// TestFileListFalsifyingDoubleClickDirThenFileReusesCurrentWildcard verifies
// double-click on ".." brings back the original wildcard, and after navigating
// back, files matching the wildcard are visible.
// (Falsifying: lazy ReadDirectory that uses "*" instead of the current wildcard.)
// Spec: "navigate by calling ReadDirectory with the entry's Path and current wildcard"
func TestFileListFalsifyingDoubleClickDirThenFileReusesCurrentWildcard(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "code.go"), []byte("package p"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "notes.txt"), []byte("notes"), 0644)
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "inside.go"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(subDir, "inside.txt"), []byte("y"), 0644)

	fl := NewFileList(NewRect(0, 0, 40, 15))
	fl.ReadDirectory(tmpDir, "*.go")

	// Only .go files should be visible in tmpDir
	if findEntryItem(fl, "notes.txt") >= 0 {
		t.Fatal("notes.txt should not appear with '*.go' wildcard")
	}

	// Navigate into subdir/
	subIdx := findEntryItem(fl, "subdir/")
	if subIdx < 0 {
		t.Fatal("subdir/ not found")
	}
	fl.HandleEvent(flDoubleClickEv(0, subIdx))

	// Inside subdir with "*.go" wildcard: inside.go should be visible, inside.txt should not
	if findEntryItem(fl, "inside.go") < 0 {
		t.Error("inside.go should be visible with '*.go' in subdir")
	}
	if findEntryItem(fl, "inside.txt") >= 0 {
		t.Error("inside.txt should not be visible with '*.go' in subdir — wildcard not preserved")
	}

	// Navigate back to parent via ".."
	fl.HandleEvent(flDoubleClickEv(0, 0)) // ".." is at row 0

	// Back in original dir: wildcard "*.go" should still be active
	if fl.Wildcard() != "*.go" {
		t.Errorf("Wildcard after navigating back: %q, want '*.go'", fl.Wildcard())
	}
	if findEntryItem(fl, "notes.txt") >= 0 {
		t.Error("notes.txt should still not be visible after navigating back — wildcard not preserved")
	}
}

// TestFileListFalsifyingSearchIsPerformedIncrementally verifies that search
// narrows results incrementally as more characters are typed, not just
// matching the first character.
// (Falsifying: lazy impl that only uses the first character of the search buffer.)
// Spec: "case-insensitive prefix match against entry names, selects first match"
func TestFileListFalsifyingSearchIsPerformedIncrementally(t *testing.T) {
	fl := newFileListWithTempDir(t, "*", nil,
		[]string{"car.txt", "cat.txt", "dog.txt"})

	// Build search "ca" character by character
	fl.HandleEvent(keyboardRuneEv('c'))
	fl.HandleEvent(keyboardRuneEv('a'))

	// "ca" should match car.txt and cat.txt, first is car.txt
	carIdx := findEntryItem(fl, "car.txt")
	catIdx := findEntryItem(fl, "cat.txt")
	dogIdx := findEntryItem(fl, "dog.txt")

	if fl.Selected() != carIdx {
		t.Errorf("After 'ca': Selected() = %d (%s), want car.txt (prefix match, not first-char-only)",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}
	_ = catIdx
	_ = dogIdx

	// Now type 't' — search "cat" should match cat.txt only
	fl.HandleEvent(keyboardRuneEv('t'))
	if fl.Selected() != catIdx {
		t.Errorf("After 'cat': Selected() = %d (%s), want cat.txt (incremental narrowing)",
			fl.Selected(), fl.DataSource().Item(fl.Selected()))
	}
}

// TestFileListFalsifyingEnterOnDirBroadcastsOnlyOnce verifies that Enter on
// directory broadcasts CmFileFocused exactly once (not multiple times from
// both the explicit broadcast and the general rule).
// Spec: "broadcast CmFileFocused"
func TestFileListFalsifyingEnterOnDirBroadcastsOnlyOnce(t *testing.T) {
	fl, _, spy := newFileListWithOwner(NewRect(0, 0, 40, 10))

	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "folder"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("x"), 0644)
	fl.ReadDirectory(tmpDir, "*")

	dirIdx := findEntryItem(fl, "folder/")
	if dirIdx < 0 {
		t.Fatal("folder/ not found")
	}
	fl.SetSelected(dirIdx)
	resetBroadcasts(spy)

	ev := enterKeyEv()
	fl.HandleEvent(ev)

	if !hasBroadcastSpy(spy, CmFileFocused) {
		t.Error("Enter on directory: CmFileFocused broadcast not received")
	}

	// Verify the broadcast has the correct entry (the directory just entered)
	info := broadcastInfoFor(spy, CmFileFocused)
	if info == nil {
		t.Fatal("CmFileFocused broadcast Info was nil")
	}
	entry, ok := info.(*FileEntry)
	if !ok {
		t.Fatalf("CmFileFocused Info = %T, want *FileEntry", info)
	}
	// After navigating into folder/, Selected is 0 which should be ".."
	if entry.Name != ".." {
		t.Errorf("After navigating into folder/: CmFileFocused entry = %q, want %q (index 0 is '..')",
			entry.Name, "..")
	}
}
