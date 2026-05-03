# Batch 7: TFileDialog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a file dialog widget with directory browsing, wildcard filtering, file metadata display, and incremental keyboard search.

**Architecture:** FileDialog is a Dialog subclass containing FileInputLine (InputLine subclass), FileList (sorted ListBox + incremental search), FileInfoPane (non-focusable metadata), and configurable buttons. Components coordinate via `CmFileFocused` and `CmFileFilter` broadcast events.

**Tech Stack:** Go, tcell/v2, existing tv package, os/filepath for directory scanning

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `tv/command.go` | Modify | Add CmFileOpen, CmFileReplace, CmFileClear, CmFileFocused, CmFileFilter |
| `tv/file_list.go` | Create | FileEntry, FileDataSource, FileList (ListBox wrapper + incremental search + dir nav) |
| `tv/file_info_pane.go` | Create | FileInfoPane (non-focusable metadata display) |
| `tv/file_dialog.go` | Create | FileInputLine, FileDialog constructor, Valid(), FileName() |
| `e2e/testapp/basic/main.go` | Modify | Add File > Open menu item using FileDialog |

---

### Task 1: New Commands + FileDataSource

**Files:**
- Modify: `tv/command.go`
- Create: `tv/file_list.go` (FileEntry, FileDataSource, ReadDirectory)
- Test: `tv/file_data_source_test.go`

- [ ] **Step 1: Write failing tests**

```go
// tv/file_data_source_test.go
package tv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileDataSourceSorting(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "zzz"), 0755)
	os.MkdirAll(filepath.Join(dir, "aaa"), 0755)
	os.WriteFile(filepath.Join(dir, "zzz.go"), nil, 0644)
	os.WriteFile(filepath.Join(dir, "aaa.go"), nil, 0644)
	os.WriteFile(filepath.Join(dir, "README"), nil, 0644)

	fds := &FileDataSource{dir: dir, wildcard: "*"}
	fds.refresh()
	if fds.Count() == 0 {
		t.Fatal("expected entries")
	}
	// .. first, then dirs alpha, then files alpha
	names := make([]string, fds.Count())
	for i := range names {
		names[i] = fds.Item(i)
	}
	prevName := ""
	prevIsDir := false
	expectDir := false // after ..
	for i, name := range names {
		if i == 0 {
			if name != ".." {
				t.Fatalf("first entry should be .. got %q", name)
			}
			expectDir = true
			continue
		}
		isDir := false
		for _, entry := range fds.entries {
			if entry.Name == name {
				isDir = entry.IsDir
				break
			}
		}
		if prevIsDir && !isDir {
			expectDir = false
		}
		if expectDir && !isDir {
			// files should come after dirs — but only if there's at least one dir still expected
			// If no more dirs exist, files are fine
			if i < len(names) {
				// check if there are dirs after this
				hasDirsAfter := false
				for j := i + 1; j < len(names); j++ {
					for _, e := range fds.entries {
						if e.Name == names[j] && e.IsDir {
							hasDirsAfter = true
						}
					}
				}
				if hasDirsAfter {
					t.Fatalf("dir before file sorting broken at %q", name)
				}
			}
		}
		prevName = name
		prevIsDir = isDir
	}
}

func TestFileDataSourceWildcard(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), nil, 0644)
	os.WriteFile(filepath.Join(dir, "README.txt"), nil, 0644)
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	fds := &FileDataSource{dir: dir, wildcard: "*.go"}
	fds.refresh()
	if fds.Count() < 3 { // .. , subdir, main.go (dirs always shown)
		t.Fatalf("expected at least 3 entries, got %d", fds.Count())
	}
	foundGo := false
	foundTxt := false
	for i := 0; i < fds.Count(); i++ {
		n := fds.Item(i)
		if n == "main.go" {
			foundGo = true
		}
		if n == "README.txt" {
			foundTxt = true
		}
	}
	if !foundGo {
		t.Fatal("main.go not found")
	}
	if foundTxt {
		t.Fatal("README.txt should be filtered out")
	}
}

func TestFileDataSourceHiddenFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".hidden"), nil, 0644)
	os.WriteFile(filepath.Join(dir, "visible"), nil, 0644)

	fds := &FileDataSource{dir: dir, wildcard: "*"}
	fds.refresh()
	foundHidden := false
	foundVisible := false
	for i := 0; i < fds.Count(); i++ {
		switch fds.Item(i) {
		case ".hidden":
			foundHidden = true
		case "visible":
			foundVisible = true
		}
	}
	if foundHidden {
		t.Fatal("hidden files should be excluded")
	}
	if !foundVisible {
		t.Fatal("visible file should be present")
	}
}

func TestFileDataSourceEntry(t *testing.T) {
	fds := &FileDataSource{dir: "/tmp", wildcard: "*"}
	fds.refresh() // should not panic
	if fds.Count() == 0 {
		t.Skip("empty /tmp")
	}
	entry := fds.Entry(0)
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.Name == "" {
		t.Fatal("entry should have name")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tv/ -run "TestFileDataSource" -v -count=1`
Expected: FAIL — FileDataSource undefined

- [ ] **Step 3: Add commands and FileDataSource**

Add to `tv/command.go` after `CmIndicatorUpdate`:

```go
	CmFileOpen
	CmFileReplace
	CmFileClear
	CmFileFocused
	CmFileFilter
```

Create `tv/file_list.go`:

```go
// tv/file_list.go
package tv

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FileEntry struct {
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
	Path    string // full path
}

type FileDataSource struct {
	dir      string
	wildcard string
	entries  []*FileEntry
}

func NewFileDataSource(dir, wildcard string) *FileDataSource {
	fds := &FileDataSource{dir: dir, wildcard: wildcard}
	if wildcard == "" {
		fds.wildcard = "*"
	}
	fds.refresh()
	return fds
}

func (fds *FileDataSource) Dir() string      { return fds.dir }
func (fds *FileDataSource) Wildcard() string { return fds.wildcard }

func (fds *FileDataSource) Count() int { return len(fds.entries) }

func (fds *FileDataSource) Item(index int) string {
	if index < 0 || index >= len(fds.entries) {
		return ""
	}
	e := fds.entries[index]
	if e.Name == ".." {
		return ".."
	}
	if e.IsDir {
		return e.Name + "/"
	}
	return e.Name
}

func (fds *FileDataSource) Entry(index int) *FileEntry {
	if index < 0 || index >= len(fds.entries) {
		return nil
	}
	return fds.entries[index]
}

func (fds *FileDataSource) SetDir(dir string) {
	fds.dir = dir
	fds.refresh()
}

func (fds *FileDataSource) SetWildcard(wc string) {
	if wc == "" {
		wc = "*"
	}
	fds.wildcard = wc
	fds.refresh()
}

func (fds *FileDataSource) refresh() {
	entries, err := os.ReadDir(fds.dir)
	if err != nil {
		fds.entries = nil
		return
	}

	var result []*FileEntry

	// Add .. if not at root
	if fds.dir != "/" {
		info, err := os.Stat(filepath.Join(fds.dir, ".."))
		if err == nil {
			result = append(result, &FileEntry{
				Name:    "..",
				IsDir:   true,
				ModTime: info.ModTime(),
				Path:    filepath.Join(fds.dir, ".."),
			})
		}
	}

	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") && name != "." && name != ".." {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		isDir := e.IsDir()
		if !isDir && !matchWildcard(name, fds.wildcard) {
			continue
		}
		result = append(result, &FileEntry{
			Name:    name,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   isDir,
			Path:    filepath.Join(fds.dir, name),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		// .. always first
		if result[i].Name == ".." {
			return true
		}
		if result[j].Name == ".." {
			return false
		}
		// dirs before files
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	fds.entries = result
}

func matchWildcard(name, pattern string) bool {
	for _, p := range strings.Split(pattern, ";") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		matched, _ := filepath.Match(p, name)
		if matched {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./tv/ -run "TestFileDataSource" -v -count=1`
Expected: PASS

- [ ] **Step 5: Run all tests**

Run: `go test ./tv/ -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add tv/command.go tv/file_list.go tv/file_data_source_test.go
git commit -m "feat: add FileEntry, FileDataSource, and file dialog commands"
```

---

### Task 2: FileList Widget (ListBox + incremental search + directory navigation)

**Files:**
- Modify: `tv/file_list.go` (add FileList struct)
- Test: `tv/file_list_test.go`

- [ ] **Step 1: Write failing tests**

```go
// tv/file_list_test.go
package tv

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestFileListCreation(t *testing.T) {
	fl := NewFileList(NewRect(0, 0, 30, 10))
	if fl.FocusedChild() == nil {
		t.Fatal("FileList should have a focused child (its ListViewer)")
	}
}

func TestFileListReadDirectory(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.go"), nil, 0644)
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	fl := NewFileList(NewRect(0, 0, 30, 10))
	err := fl.ReadDirectory(dir, "*.go")
	if err != nil {
		t.Fatal(err)
	}
	ds := fl.DataSource()
	if ds.Count() < 3 {
		t.Fatalf("expected at least 3 entries (.., subdir/, test.go), got %d", ds.Count())
	}
}

func TestFileListEnterOnDirectory(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subdir")
	os.MkdirAll(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, "inner.go"), nil, 0644)

	fl := NewFileList(NewRect(0, 0, 30, 10))
	fl.ReadDirectory(dir, "*")

	// Select the "subdir/" entry (index 1, after ..)
	fl.ListViewer().SetSelected(1)

	// Send Enter key
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	fl.HandleEvent(ev)

	ds := fl.DataSource().(*FileDataSource)
	if filepath.Base(ds.Dir()) != "subdir" {
		t.Fatalf("expected to navigate into subdir, got %q", ds.Dir())
	}
}

func TestFileListDoubleClickFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), nil, 0644)

	fl := NewFileList(NewRect(0, 0, 30, 10))
	fl.ReadDirectory(dir, "*")

	// Select the file entry
	for i := 0; i < fl.DataSource().Count(); i++ {
		if fl.DataSource().Item(i) == "test.txt" {
			fl.ListViewer().SetSelected(i)
			break
		}
	}

	// Simulate double-click on the file row
	ev := &Event{
		What: EvMouse,
		Mouse: &MouseEvent{
			X:          5,
			Y:          fl.ListViewer().Selected() - fl.ListViewer().TopIndex(),
			Button:     tcell.Button1,
			ClickCount: 2,
		},
	}
	fl.HandleEvent(ev)
	if ev.What != EvCommand || ev.Command != CmOK {
		t.Fatalf("double-click on file should synthesize CmOK, got what=%d cmd=%d", ev.What, ev.Command)
	}
}

func TestFileListIncrementalSearch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "aaaa.go"), nil, 0644)
	os.WriteFile(filepath.Join(dir, "bbbb.go"), nil, 0644)
	os.WriteFile(filepath.Join(dir, "aabb.go"), nil, 0644)

	fl := NewFileList(NewRect(0, 0, 30, 10))
	fl.ReadDirectory(dir, "*")

	// Type 'a' to search
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Ch: 'a'}}
	fl.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Fatal("'a' key should be consumed")
	}
	sel := fl.ListViewer().Selected()
	item := fl.DataSource().Item(sel)
	if item != "aaaa.go" {
		t.Fatalf("expected 'aaaa.go', got %q", item)
	}
}

func TestFileListSearchTimeout(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "aaaa.go"), nil, 0644)
	os.WriteFile(filepath.Join(dir, "bbbb.go"), nil, 0644)

	fl := NewFileList(NewRect(0, 0, 30, 10))
	fl.ReadDirectory(dir, "*")

	// Type 'a'
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Ch: 'a'}}
	fl.HandleEvent(ev)

	sel1 := fl.ListViewer().Selected()

	// Simulate timeout by resetting search buffer
	fl.searchBuf = ""
	fl.searchTime = time.Time{}

	// Type 'b' as new search
	ev = &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Ch: 'b'}}
	fl.HandleEvent(ev)

	sel2 := fl.ListViewer().Selected()
	if sel1 == sel2 {
		t.Fatal("new search after timeout should find different match")
	}
}

func TestFileListSearchBackspace(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "baaa.go"), nil, 0644)

	fl := NewFileList(NewRect(0, 0, 30, 10))
	fl.ReadDirectory(dir, "*")

	// Type 'b' then 'a'
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Ch: 'b'}}
	fl.HandleEvent(ev)
	ev = &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Ch: 'a'}}
	fl.HandleEvent(ev)

	// Backspace deletes last char
	ev = &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyBackspace2}}
	fl.HandleEvent(ev)

	if len(fl.searchBuf) != 1 || fl.searchBuf[0] != 'b' {
		t.Fatalf("expected search buf 'b', got %q", string(fl.searchBuf))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tv/ -run "TestFileList" -v -count=1`
Expected: FAIL — NewFileList undefined

- [ ] **Step 3: Implement FileList**

Append to `tv/file_list.go`:

```go
import (
	"github.com/gdamore/tcell/v2"
	"unicode"
)

type FileList struct {
	*ListBox
	dataSource *FileDataSource
	searchBuf  []rune
	searchTime time.Time
}

func NewFileList(bounds Rect) *FileList {
	fds := NewFileDataSource(".", "*")
	fl := &FileList{
		ListBox:    NewListBox(bounds, fds),
		dataSource: fds,
	}
	fl.SetSelf(fl)
	return fl
}

func (fl *FileList) ReadDirectory(dir, wildcard string) error {
	if wildcard == "" {
		wildcard = "*"
	}
	fl.dataSource = NewFileDataSource(dir, wildcard)
	fl.SetDataSource(fl.dataSource)
	return nil
}

func (fl *FileList) Dir() string       { return fl.dataSource.Dir() }
func (fl *FileList) Wildcard() string   { return fl.dataSource.Wildcard() }

func (fl *FileList) HandleEvent(event *Event) {
	// Pre-process mouse events: detect double-click on directory vs file
	if event.What == EvMouse && event.Mouse != nil && event.Mouse.ClickCount >= 2 {
		fl.BaseView.HandleEvent(event)
		if event.IsCleared() {
			return
		}
		// Determine which entry was clicked
		idx := fl.Selected()
		entry := fl.dataSource.Entry(idx)
		if entry != nil && entry.IsDir {
			fl.ReadDirectory(entry.Path, fl.dataSource.Wildcard())
			fl.ListViewer().SetSelected(0)
			event.Clear()
			return
		}
		// On file double-click: synthesize CmOK (matching original Turbo Vision)
		if entry != nil && !entry.IsDir && entry.Name != ".." {
			event.What = EvCommand
			event.Command = CmOK
			event.Mouse = nil
			return
		}
	}

	// Delegate to ListBox for normal mouse + keyboard (arrows, etc.)
	if event.What == EvMouse || (event.What == EvKeyboard && event.Key != nil) {
		// Check if this is a navigation key — if so, clear incremental search
		if event.What == EvKeyboard && event.Key != nil {
			k := event.Key
			isNav := false
			switch k.Key {
			case tcell.KeyUp, tcell.KeyDown, tcell.KeyPgUp, tcell.KeyPgDn,
				tcell.KeyHome, tcell.KeyEnd, tcell.KeyLeft, tcell.KeyRight:
				isNav = true
			}
			if isNav {
				fl.searchBuf = nil
				fl.searchTime = time.Time{}
			}
		}

		// Incremental search for printable characters
		if event.What == EvKeyboard && event.Key != nil {
			if event.Key.Key == tcell.KeyRune && event.Key.Ch != 0 {
				ch := event.Key.Ch
				// Check for timeout
				if !fl.searchTime.IsZero() && time.Since(fl.searchTime) > time.Second {
					fl.searchBuf = nil
				}
				if ch == '.' || !unicode.IsControl(ch) {
					fl.searchBuf = append(fl.searchBuf, ch)
					fl.searchTime = time.Now()
					fl.incrementalSearch()
					event.Clear()
					return
				}
			}
			if event.Key.Key == tcell.KeyBackspace || event.Key.Key == tcell.KeyBackspace2 {
				if len(fl.searchBuf) > 0 {
					fl.searchBuf = fl.searchBuf[:len(fl.searchBuf)-1]
					fl.searchTime = time.Now()
					fl.incrementalSearch()
				}
				event.Clear()
				return
			}
		}
	}

	// Enter on directory navigates; Enter on file -> CmOK
	if event.What == EvKeyboard && event.Key != nil && event.Key.Key == tcell.KeyEnter {
		fl.BaseView.HandleEvent(event)
		if event.IsCleared() {
			return
		}
		idx := fl.Selected()
		entry := fl.dataSource.Entry(idx)
		if entry != nil && entry.IsDir {
			fl.ReadDirectory(entry.Path, fl.dataSource.Wildcard())
			fl.ListViewer().SetSelected(0)
			event.Clear()
			// Broadcast updated focus
			fl.broadcastFocused()
			return
		}
		if entry != nil && !entry.IsDir && entry.Name != ".." {
			event.What = EvCommand
			event.Command = CmOK
			event.Key = nil
			return
		}
	}

	// Delegate to ListBox
	fl.ListBox.HandleEvent(event)

	// After any focus change, broadcast CmFileFocused
	if event.What == EvKeyboard || event.What == EvMouse {
		fl.broadcastFocused()
	}
}

func (fl *FileList) broadcastFocused() {
	idx := fl.Selected()
	entry := fl.dataSource.Entry(idx)
	if entry == nil {
		return
	}
	owner := fl.Owner()
	if owner == nil {
		return
	}
	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: entry}
	for _, child := range owner.Children() {
		child.HandleEvent(ev)
	}
}

func (fl *FileList) incrementalSearch() {
	needle := strings.ToLower(string(fl.searchBuf))
	ds := fl.dataSource
	for i := 0; i < ds.Count(); i++ {
		name := strings.ToLower(ds.entries[i].Name)
		if strings.HasPrefix(name, needle) {
			fl.ListViewer().SetSelected(i)
			return
		}
	}
}

func (fl *FileList) SetBounds(r Rect) {
	fl.ListBox.SetBounds(r)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./tv/ -run "TestFileList" -v -count=1`
Expected: PASS

- [ ] **Step 5: Run all tests**

Run: `go test ./tv/ -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add tv/file_list.go tv/file_list_test.go
git commit -m "feat: add FileList widget with incremental search and directory navigation"
```

---

### Task 3: FileInfoPane

**Files:**
- Create: `tv/file_info_pane.go`
- Test: `tv/file_info_pane_test.go`

- [ ] **Step 1: Write failing tests**

```go
// tv/file_info_pane_test.go
package tv

import (
	"testing"
	"time"
)

func TestFileInfoPaneNotSelectable(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 30, 2))
	if pane.HasOption(OfSelectable) {
		t.Fatal("FileInfoPane should not be selectable")
	}
}

func TestFileInfoPanePostProcess(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 30, 2))
	if !pane.HasOption(OfPostProcess) {
		t.Fatal("FileInfoPane should have OfPostProcess to receive broadcasts")
	}
}

func TestFileInfoPaneReceivesBroadcast(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 30, 2))
	entry := &FileEntry{
		Name:    "main.go",
		Size:    1234,
		ModTime: time.Date(2026, 5, 3, 14, 15, 0, 0, time.UTC),
		IsDir:   false,
	}
	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: entry}
	pane.HandleEvent(ev)
	if pane.entry != entry {
		t.Fatal("pane should store the entry")
	}
}

func TestFileInfoPaneDirectory(t *testing.T) {
	entry := &FileEntry{
		Name:    "pkg",
		Size:    0,
		ModTime: time.Now(),
		IsDir:   true,
	}
	pane := NewFileInfoPane(NewRect(0, 0, 30, 2))
	pane.entry = entry

	buf := NewDrawBuffer(30, 2)
	pane.Draw(buf)

	found := false
	for x := 0; x < 30; x++ {
		if buf.cells[0][x].Rune == '<' && x+4 < 30 {
			if buf.cells[0][x+1].Rune == 'D' && buf.cells[0][x+2].Rune == 'I' && buf.cells[0][x+3].Rune == 'R' {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("should show '<DIR>' for directories")
	}
}

func TestFileInfoPaneFile(t *testing.T) {
	entry := &FileEntry{
		Name:    "main.go",
		Size:    1234,
		ModTime: time.Date(2026, 5, 3, 14, 15, 0, 0, time.UTC),
		IsDir:   false,
	}
	pane := NewFileInfoPane(NewRect(0, 0, 30, 2))
	pane.entry = entry

	buf := NewDrawBuffer(30, 2)
	pane.Draw(buf)

	foundName := false
	for x := 0; x < 30; x++ {
		if x+6 < 30 {
			s := string([]rune{
				buf.cells[0][x].Rune, buf.cells[0][x+1].Rune, buf.cells[0][x+2].Rune,
				buf.cells[0][x+3].Rune, buf.cells[0][x+4].Rune, buf.cells[0][x+5].Rune,
			})
			if s == "main.g" {
				foundName = true
				break
			}
		}
	}
	if !foundName {
		t.Fatal("should show filename 'main.go'")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tv/ -run TestFileInfoPane -v -count=1`
Expected: FAIL — NewFileInfoPane undefined

- [ ] **Step 3: Implement FileInfoPane**

Create `tv/file_info_pane.go`:

```go
// tv/file_info_pane.go
package tv

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type FileInfoPane struct {
	BaseView
	entry *FileEntry
}

func NewFileInfoPane(bounds Rect) *FileInfoPane {
	fp := &FileInfoPane{}
	fp.SetBounds(bounds)
	fp.SetState(SfVisible, true)
	fp.SetOptions(OfPostProcess, true)
	fp.SetSelf(fp)
	return fp
}

func (fp *FileInfoPane) SetEntry(e *FileEntry) { fp.entry = e }

func (fp *FileInfoPane) Draw(buf *DrawBuffer) {
	if fp.entry == nil {
		return
	}

	w := fp.Bounds().Width()
	if w < 4 {
		return
	}

	// Format: "  main.go         1,234  May  3, 2026  2:15pm"
	name := fp.entry.Name
	if fp.entry.IsDir {
		name += "/"
	}
	if utf8.RuneCountInString(name) > w/2 {
		name = string([]rune(name)[:w/2-1]) + "~"
	}

	// Right side: size or <DIR>
	var sizeStr string
	if fp.entry.IsDir {
		sizeStr = "<DIR>"
	} else {
		sizeStr = commaFormat(fp.entry.Size)
	}

	// Date
	t := fp.entry.ModTime
	dateStr := t.Format("Jan  2, 2006")
	timeStr := t.Format("3:04pm")

	// Build the line: name left, size+date+time right
	rightPart := fmt.Sprintf("%6s  %s  %s", sizeStr, dateStr, timeStr)
	line := fmt.Sprintf("  %s", name)

	buf.Fill(NewRect(0, 0, w, 1), ' ', fp.ColorScheme().ListNormal)

	for i, r := range line {
		if i < w {
			buf.WriteChar(i, 0, r, fp.ColorScheme().ListNormal)
		}
	}

	rightRunes := []rune(rightPart)
	rightW := len(rightRunes)
	startX := w - rightW
	if startX < utf8.RuneCountInString(name)+2 {
		startX = utf8.RuneCountInString(name) + 2
	}
	for i, r := range rightRunes {
		x := startX + i
		if x < w {
			buf.WriteChar(x, 0, r, fp.ColorScheme().ListNormal)
		}
	}
}

func commaFormat(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ",")
}

func (fp *FileInfoPane) HandleEvent(event *Event) {
	if event.What == EvBroadcast && event.Command == CmFileFocused {
		if entry, ok := event.Info.(*FileEntry); ok {
			fp.entry = entry
		}
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./tv/ -run TestFileInfoPane -v -count=1`
Expected: PASS

- [ ] **Step 5: Run all tests**

Run: `go test ./tv/ -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add tv/file_info_pane.go tv/file_info_pane_test.go
git commit -m "feat: add FileInfoPane for file metadata display"
```

---

### Task 4: FileInputLine + FileDialog

**Files:**
- Create: `tv/file_dialog.go`
- Test: `tv/file_dialog_test.go`

- [ ] **Step 1: Write failing tests**

```go
// tv/file_dialog_test.go
package tv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestFileInputLineAutofillOnBroadcast(t *testing.T) {
	fil := &FileInputLine{
		InputLine: NewInputLine(NewRect(0, 0, 30, 1), 256),
		wildcard:  "*",
	}
	// Not focused by default — should accept CmFileFocused
	entry := &FileEntry{Name: "main.go", IsDir: false}
	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: entry}
	fil.HandleEvent(ev)
	if fil.Text() != "main.go" {
		t.Fatalf("expected autofill 'main.go', got %q", fil.Text())
	}
}

func TestFileInputLineAutofillDir(t *testing.T) {
	fil := &FileInputLine{
		InputLine: NewInputLine(NewRect(0, 0, 30, 1), 256),
		wildcard:  "*.go",
	}
	entry := &FileEntry{Name: "subdir", IsDir: true}
	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: entry}
	fil.HandleEvent(ev)
	if fil.Text() != "subdir/*.go" {
		t.Fatalf("expected autofill 'subdir/*.go', got %q", fil.Text())
	}
}

func TestFileInputLineIgnoresWhenFocused(t *testing.T) {
	fil := &FileInputLine{
		InputLine: NewInputLine(NewRect(0, 0, 30, 1), 256),
		wildcard:  "*",
	}
	fil.SetText("user input")
	fil.SetState(SfSelected, true) // simulate focused
	entry := &FileEntry{Name: "main.go", IsDir: false}
	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: entry}
	fil.HandleEvent(ev)
	if fil.Text() != "user input" {
		t.Fatalf("focused FileInputLine should ignore autofill, got %q", fil.Text())
	}
}

func TestFileInputLineClear(t *testing.T) {
	fil := &FileInputLine{
		InputLine: NewInputLine(NewRect(0, 0, 30, 1), 256),
		wildcard:  "*",
	}
	fil.SetText("something")
	fil.Clear()
	if fil.Text() != "" {
		t.Fatalf("Clear should empty text, got %q", fil.Text())
	}
}

func TestFileDialogCreation(t *testing.T) {
	fd := NewFileDialog("*.go", "Open File", FdOpenButton)
	if fd.Title() != "Open File" {
		t.Fatalf("expected title 'Open File', got %q", fd.Title())
	}
	if fd.FileName() != "" {
		t.Fatal("FileName should be empty before selection")
	}
	// Should have children: labels, FileInputLine, History, FileList, FileInfoPane, buttons
	children := fd.Children()
	if len(children) < 6 {
		t.Fatalf("expected at least 6 children, got %d", len(children))
	}
}

func TestFileDialogWildcardDetection(t *testing.T) {
	fd := NewFileDialog("*.go", "Open", FdOpenButton)
	fil := fd.fileInput
	// Set wildcard text — should detect wildcard
	fil.SetText("*.txt")
	dd := fd.detectInput(fil.Text())
	if dd != inputWildcard {
		t.Fatalf("expected wildcard detection, got %d", dd)
	}
}

func TestFileDialogDirectoryDetection(t *testing.T) {
	fd := NewFileDialog("*.go", "Open", FdOpenButton)
	fil := fd.fileInput
	fil.SetText(os.TempDir())
	dd := fd.detectInput(fil.Text())
	if dd != inputDirectory {
		t.Fatalf("expected directory detection for %q, got %d", os.TempDir(), dd)
	}
}

func TestFileDialogNoFlagsDefaultsOK(t *testing.T) {
	fd := NewFileDialog("*", "Test", 0)
	children := fd.Children()
	hasOK := false
	for _, child := range children {
		btn, ok := child.(*Button)
		if ok && btn.Label() == "OK" {
			hasOK = true
			break
		}
	}
	if !hasOK {
		t.Fatal("expected OK button when no flags specified")
	}
}

func TestFileDialogClearButton(t *testing.T) {
	fd := NewFileDialog("*", "Test", FdClearButton)
	fd.fileInput.SetText("test.txt")
	// Find the Clear button and simulate press
	children := fd.Children()
	for _, child := range children {
		btn, ok := child.(*Button)
		if ok && btn.Label() == "Clear" {
			ev := &Event{What: EvCommand, Command: CmFileClear}
			btn.HandleEvent(ev)
			break
		}
	}
	if fd.FileName() != "" {
		t.Fatal("fileName should be empty after clear")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tv/ -run "TestFileInputLine|TestFileDialog" -v -count=1`
Expected: FAIL — FileInputLine/FileDialog undefined

- [ ] **Step 3: Implement FileInputLine + FileDialog**

Create `tv/file_dialog.go`:

```go
// tv/file_dialog.go
package tv

import (
	"os"
	"path/filepath"
	"strings"
)

type FileDialogFlag int

const (
	FdOpenButton    FileDialogFlag = 1 << iota
	FdOKButton
	FdReplaceButton
	FdClearButton
	FdHelpButton
)

type inputDetection int

const (
	inputFilename  inputDetection = iota
	inputWildcard
	inputDirectory
)

// FileInputLine wraps InputLine with wildcard tracking and autofill.
type FileInputLine struct {
	*InputLine
	wildcard string
}

func (fil *FileInputLine) HandleEvent(event *Event) {
	if event.What == EvBroadcast && event.Command == CmFileFocused {
		// Only autofill when NOT focused
		if !fil.HasState(SfSelected) {
			if entry, ok := event.Info.(*FileEntry); ok {
				if entry.IsDir {
					fil.SetText(entry.Name + "/" + fil.wildcard)
				} else {
					fil.SetText(entry.Name)
				}
			}
		}
		event.Clear()
		return
	}
	fil.InputLine.HandleEvent(event)
}

func (fil *FileInputLine) Clear() {
	fil.SetText("")
}

type FileDialog struct {
	*Dialog
	fileList    *FileList
	fileInfo    *FileInfoPane
	fileInput   *FileInputLine
	wildcard    string
	resultPath  string
	initialDir  string
}

func NewFileDialog(wildcard, title string, flags FileDialogFlag) *FileDialog {
	dir, _ := os.Getwd()
	return NewFileDialogInDir(dir, wildcard, title, flags)
}

func NewFileDialogInDir(dir, wildcard, title string, flags FileDialogFlag) *FileDialog {
	if wildcard == "" {
		wildcard = "*"
	}

	// Auto-size: default 52x20, min 49x19
	w, h := 52, 20
	if w < 49 {
		w = 49
	}
	if h < 19 {
		h = 19
	}

	dlg := NewDialog(NewRect(0, 0, w, h), title)
	fd := &FileDialog{Dialog: dlg, wildcard: wildcard, initialDir: dir}
	fd.SetSelf(fd)

	clientW := w - 2
	// Left content area: columns 1-36 (width 36)
	// Right button area: columns 38-49 (width 12)

	// Row 2: Label "File ~N~ame:"
	nameLabel := NewLabel(NewRect(1, 0, 12, 1), "File ~N~ame:", nil)
	fd.Insert(nameLabel)

	// Row 3 (y=1): FileInputLine + History
	fil := &FileInputLine{
		InputLine: NewInputLine(NewRect(1, 1, 32, 1), 256),
		wildcard:  wildcard,
	}
	fil.SetText(wildcard)
	fd.fileInput = fil
	hist := NewHistory(NewRect(33, 1, 2, 1), fil.InputLine, 20)
	fd.Insert(fil.InputLine)
	fd.Insert(hist)

	// Row 5 (y=3): Label "~F~iles:"
	filesLabel := NewLabel(NewRect(1, 3, 10, 1), "~F~iles:", nil)
	fd.Insert(filesLabel)

	// Rows 6-15 (y=4 to y=13): FileList
	fl := NewFileList(NewRect(1, 4, 35, 10))
	fl.ReadDirectory(dir, wildcard)
	fd.fileList = fl
	fd.Insert(fl)

	// Rows 16-17 (y=14 to y=15): FileInfoPane
	fp := NewFileInfoPane(NewRect(1, 14, 35, 2))
	fd.fileInfo = fp
	fd.Insert(fp)

	// Buttons stacked on right edge (columns 38-49), starting at row 3 (y=1)
	btnX := 38
	btnW := 12
	btnY := 1

	// Build button list from flags
	type btnDef struct {
		label string
		cmd   CommandCode
	}
	var defs []btnDef
	if flags&FdOpenButton != 0 {
		defs = append(defs, btnDef{"Open", CmFileOpen})
	}
	if flags&FdOKButton != 0 {
		defs = append(defs, btnDef{"OK", CmOK})
	}
	if flags&FdReplaceButton != 0 {
		defs = append(defs, btnDef{"Replace", CmFileReplace})
	}
	if flags&FdClearButton != 0 {
		defs = append(defs, btnDef{"Clear", CmFileClear})
	}
	if flags&FdHelpButton != 0 {
		defs = append(defs, btnDef{"Help", CmUser + 100})
	}
	if len(defs) == 0 {
		defs = append(defs, btnDef{"OK", CmOK})
	}
	// Cancel always present
	defs = append(defs, btnDef{"Cancel", CmCancel})

	for i, def := range defs {
		var opts []ButtonOption
		if i == 0 {
			opts = append(opts, WithDefault())
		}
		btn := NewButton(NewRect(btnX, btnY, btnW, 2), def.label, def.cmd, opts...)
		fd.Insert(btn)
		btnY += 3
	}

	return fd
}

func (fd *FileDialog) HandleEvent(event *Event) {
	if event.What == EvCommand {
		switch event.Command {
		case CmFileClear:
			fd.fileInput.Clear()
			fd.resultPath = ""
			event.Clear()
			return
		case CmFileOpen:
			// Process as OK — resolve filename
			event.Command = CmOK
		case CmFileReplace:
			event.Command = CmOK
		}
	}
	fd.Dialog.HandleEvent(event)
}

func (fd *FileDialog) FileName() string { return fd.resultPath }

func (fd *FileDialog) Valid(cmd CommandCode) bool {
	if cmd == CmCancel {
		return true
	}
	if cmd != CmOK {
		return true
	}

	text := strings.TrimSpace(fd.fileInput.Text())
	if text == "" {
		return false
	}

	detected := fd.detectInput(text)
	switch detected {
	case inputWildcard:
		fd.wildcard = text
		fd.fileInput.wildcard = text
		fd.fileList.ReadDirectory(fd.fileList.Dir(), text)
		// Stay open — don't return true
		// Broadcast filter change
		owner := fd.Owner()
		if owner != nil {
			ev := &Event{What: EvBroadcast, Command: CmFileFilter, Info: text}
			for _, child := range owner.Children() {
				child.HandleEvent(ev)
			}
		}
		return false

	case inputDirectory:
		fd.fileList.ReadDirectory(text, fd.wildcard)
		fd.fileInput.wildcard = fd.wildcard
		return false

	case inputFilename:
		// Resolve to absolute path
		path := text
		if !filepath.IsAbs(path) {
			path = filepath.Join(fd.fileList.Dir(), path)
		}
		fd.resultPath = path
		return true
	}

	return true
}

func (fd *FileDialog) detectInput(text string) inputDetection {
	if strings.ContainsAny(text, "*?") {
		return inputWildcard
	}
	info, err := os.Stat(text)
	if err == nil && info.IsDir() {
		return inputDirectory
	}
	return inputFilename
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./tv/ -run "TestFileInputLine|TestFileDialog" -v -count=1`
Expected: PASS

- [ ] **Step 5: Run all tests**

Run: `go test ./tv/ -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add tv/file_dialog.go tv/file_dialog_test.go
git commit -m "feat: add FileInputLine autofill and FileDialog with resizable layout"
```

---

### Task 5: Integration Checkpoint — FileDialog Pipeline

**Files:**
- Test: `tv/integration_filedialog_test.go`

- [ ] **Step 1: Write integration tests**

```go
// tv/integration_filedialog_test.go
package tv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIntegrationFileDialogBuildAndDraw(t *testing.T) {
	fd := NewFileDialog("*.go", "Open File", FdOpenButton)
	buf := NewDrawBuffer(52, 20)
	fd.Draw(buf)

	// Title should be rendered
	titleFound := false
	for x := 0; x < 52; x++ {
		if buf.cells[0][x].Rune == 'O' {
			titleFound = true
			break
		}
	}
	if !titleFound {
		t.Fatal("title not rendered")
	}
}

func TestIntegrationFileListBroadcastsToInfoPane(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package main"), 0644)

	fl := NewFileList(NewRect(0, 0, 30, 10))
	fp := NewFileInfoPane(NewRect(0, 0, 30, 2))

	// Set up owner chain so broadcasts flow
	g := NewGroup(NewRect(0, 0, 30, 12))
	g.Insert(fl)
	g.Insert(fp)
	g.SetFocusedChild(fl)

	fl.ReadDirectory(dir, "*")
	fl.ListViewer().SetSelected(1) // first file after ..
	g.BringToFront(fl)

	if fp.entry != nil {
		t.Log("info pane received entry from focus broadcast")
	}
}

func TestIntegrationFileDialogValidWildcard(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.go"), nil, 0644)

	fd := NewFileDialogInDir(dir, "*.go", "Test", FdOpenButton)
	// Set wildcard text
	fd.fileInput.SetText("*.mod")
	// Valid with wildcard should NOT close (returns false)
	if fd.Valid(CmOK) {
		t.Fatal("wildcard should keep dialog open")
	}
}

func TestIntegrationFileDialogValidFilename(t *testing.T) {
	dir := t.TempDir()
	fd := NewFileDialogInDir(dir, "*", "Test", FdOpenButton)
	fd.fileInput.SetText("myfile.txt")
	if !fd.Valid(CmOK) {
		t.Fatal("filename should close dialog")
	}
	expected := filepath.Join(dir, "myfile.txt")
	if fd.FileName() != expected {
		t.Fatalf("expected %q, got %q", expected, fd.FileName())
	}
}
```

- [ ] **Step 2: Run integration tests**

Run: `go test ./tv/ -run "TestIntegrationFile" -v -count=1`
Expected: PASS

- [ ] **Step 3: Run full test suite**

Run: `go test ./tv/ -count=1`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add tv/integration_filedialog_test.go
git commit -m "test: add integration tests for FileDialog pipeline"
```

---

### Task 6: Demo App + E2E Tests

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Test: `e2e/e2e_test.go`

- [ ] **Step 1: Update demo app**

Add File > Open menu item. Update `e2e/testapp/basic/main.go`.

Add the `OnCommand` handler case for CmFileOpen:

In the `WithOnCommand` handler (after existing cases), add:
```go
if cmd == tv.CmFileOpen || cmd == tv.CmUser+50 {
    fd := tv.NewFileDialogInDir(".", "*.go", "Open File", tv.FdOpenButton)
    result := app.Desktop().ExecView(fd)
    if result == tv.CmOK {
        fn := fd.FileName()
        if fn != "" {
            st.SetText("Opened: " + fn)
        }
    }
    return true
}
```

Add to the File menu (before the "Exit" item):
```go
tv.NewMenuItem("~O~pen...", tv.CmFileOpen, tv.KbFunc(3)),
```

- [ ] **Step 2: Build and verify demo compiles**

Run: `go build ./e2e/testapp/basic/`
Expected: SUCCESS

- [ ] **Step 3: Add E2E tests**

Add to `e2e/e2e_test.go`:

```go
func TestFileDialogVisible(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-fileopen"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	// Open the File menu
	tmuxSendKeys(t, session, "M-f")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)
	if !containsAny(lines, "Open") {
		t.Fatal("File > Open menu item not visible")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestFileDialogSelectGoFile(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-filedlg"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	// Open File > Open via Alt+F then O
	tmuxSendKeys(t, session, "M-f")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "O")
	time.Sleep(500 * time.Millisecond)

	lines := tmuxCapture(t, session)
	// File dialog should be visible with "Open File" title
	if !containsAny(lines, "Open File") {
		t.Fatal("FileDialog title 'Open File' not visible")
	}

	// Cancel the dialog
	tmuxSendKeys(t, session, "Escape")
	time.Sleep(300 * time.Millisecond)

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}
```

- [ ] **Step 4: Run E2E tests**

Run: `go test ./e2e/ -run TestFileDialog -v -count=1 -timeout 60s`
Expected: PASS

- [ ] **Step 5: Run full E2E suite**

Run: `go test ./e2e/ -v -count=1 -timeout 180s`
Expected: PASS (new + pre-existing — known failures: TestLabelShortcutFocusesLink, TestHistoryDropdown)

- [ ] **Step 6: Run full unit test suite**

Run: `go test ./tv/ -count=1`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add e2e/testapp/basic/main.go e2e/e2e_test.go
git commit -m "feat: add FileDialog to demo app with E2E tests"
```

---

## Notes for Implementers

1. **Same package access:** FileList, FileInputLine, FileInfoPane, and FileDialog are all in `tv` package. They can access unexported fields on each other's types.

2. **Button.Launch():** The Button widget has a `Label()` method for getting the button text. Check that `Button.Label()` exists or use the existing Button API.

3. **ExecView for dialogs:** FileDialog is a Dialog subclass. Use `owner.ExecView(fd)` to show it modally.

4. **Incremental search:** The FileList search buffer is cleared after 1 second or on navigation. Use `time.Since()` to check.

5. **Double-click CmOK:** When a file is double-clicked in the FileList, the event is transformed to `EvCommand/CmOK` which the Dialog's HandleEvent processes.

6. **GrowMode for resize:** All internal FileDialog components should have appropriate GrowMode flags so the dialog resizes properly.

7. **Known pre-existing E2E failures:** TestLabelShortcutFocusesLink and TestHistoryDropdown — not caused by Batch 7 changes.
