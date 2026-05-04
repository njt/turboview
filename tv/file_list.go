package tv

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// FileEntry represents a single file or directory entry in the file listing.
type FileEntry struct {
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
	Path    string // full path
}

// FileDataSource implements ListDataSource and represents a directory listing
// filtered by a wildcard pattern.
type FileDataSource struct {
	dir      string
	wildcard string
	entries  []*FileEntry
}

// NewFileDataSource creates a FileDataSource for the given directory and
// wildcard pattern. If wildcard is empty, it defaults to "*".
func NewFileDataSource(dir, wildcard string) *FileDataSource {
	if wildcard == "" {
		wildcard = "*"
	}
	fds := &FileDataSource{
		dir:      dir,
		wildcard: wildcard,
	}
	fds.refresh()
	return fds
}

// Count returns the number of entries in the file listing.
func (fds *FileDataSource) Count() int {
	return len(fds.entries)
}

// Item returns the display text for the entry at the given index.
// Returns ".." for the parent entry, name + "/" for directories,
// name for files, and "" for out-of-bounds indices.
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

// Entry returns the FileEntry at the given index, or nil for out-of-bounds.
func (fds *FileDataSource) Entry(index int) *FileEntry {
	if index < 0 || index >= len(fds.entries) {
		return nil
	}
	return fds.entries[index]
}

// Dir returns the current directory path.
func (fds *FileDataSource) Dir() string {
	return fds.dir
}

// Wildcard returns the current wildcard pattern.
func (fds *FileDataSource) Wildcard() string {
	return fds.wildcard
}

// SetDir changes the current directory and refreshes the listing.
func (fds *FileDataSource) SetDir(dir string) {
	fds.dir = dir
	fds.refresh()
}

// SetWildcard changes the wildcard pattern and refreshes the listing.
// If wc is empty, it defaults to "*".
func (fds *FileDataSource) SetWildcard(wc string) {
	if wc == "" {
		wc = "*"
	}
	fds.wildcard = wc
	fds.refresh()
}

// refresh re-reads the directory and rebuilds the sorted, filtered entry list.
func (fds *FileDataSource) refresh() {
	fds.entries = nil

	entries, err := os.ReadDir(fds.dir)
	if err != nil {
		return
	}

	isRoot := filepath.Clean(fds.dir) == "/"

	var parent []*FileEntry
	var dirs []*FileEntry
	var files []*FileEntry

	// Add ".." parent entry (only if not at filesystem root)
	if !isRoot {
		parent = append(parent, &FileEntry{
			Name:  "..",
			IsDir: true,
			Path:  filepath.Dir(fds.dir),
		})
	}

	// Split wildcard into patterns
	patterns := strings.Split(fds.wildcard, ";")

	for _, entry := range entries {
		name := entry.Name()

		// Exclude hidden files and directories
		if strings.HasPrefix(name, ".") {
			continue
		}

		if entry.IsDir() {
			// Directories are always included regardless of wildcard
			dirs = append(dirs, &FileEntry{
				Name:  name,
				IsDir: true,
				Path:  filepath.Join(fds.dir, name),
			})
		} else {
			// Files must match at least one wildcard pattern
			matches := false
			for _, pattern := range patterns {
				if ok, _ := filepath.Match(pattern, name); ok {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}

			// Build file entry with size and mod time
			fe := &FileEntry{
				Name:  name,
				IsDir: false,
				Path:  filepath.Join(fds.dir, name),
			}
			if info, err := entry.Info(); err == nil {
				fe.Size = info.Size()
				fe.ModTime = info.ModTime()
			}
			files = append(files, fe)
		}
	}

	// Sort directories case-insensitively
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})

	// Sort files case-insensitively
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	// Combine: parent first, then dirs, then files
	fds.entries = parent
	fds.entries = append(fds.entries, dirs...)
	fds.entries = append(fds.entries, files...)
}

// FileList is a *ListBox wrapper that displays a directory listing with
// incremental search, double-click directory navigation, and file selection
// via CmOK.
type FileList struct {
	*ListBox
	searchBuf  []rune
	searchTime time.Time
}

// NewFileList creates a FileList with a FileDataSource pointed at the current
// directory with wildcard "*".
func NewFileList(bounds Rect) *FileList {
	fds := NewFileDataSource(".", "*")
	lb := NewListBox(bounds, fds)
	fl := &FileList{ListBox: lb}
	fl.SetSelf(fl)
	return fl
}

// ReadDirectory reads the given directory filtered by wildcard and updates the
// internal data source.
func (fl *FileList) ReadDirectory(dir, wildcard string) error {
	fds := NewFileDataSource(dir, wildcard)
	fl.SetDataSource(fds)
	return nil
}

// Dir returns the current directory being displayed.
func (fl *FileList) Dir() string {
	return fl.DataSource().(*FileDataSource).Dir()
}

// Wildcard returns the current wildcard filter pattern.
func (fl *FileList) Wildcard() string {
	return fl.DataSource().(*FileDataSource).Wildcard()
}

// dataSource returns the internal data source cast to *FileDataSource.
func (fl *FileList) dataSource() *FileDataSource {
	return fl.DataSource().(*FileDataSource)
}

// HandleEvent processes events for the FileList, handling double-click
// navigation, incremental search, and delegating other events to the embedded
// ListBox.
func (fl *FileList) HandleEvent(event *Event) {
	wasKbOrMouse := event.What == EvKeyboard || event.What == EvMouse

	// 1. Double-click handling (BEFORE delegation)
	if event.What == EvMouse && event.Mouse != nil && event.Mouse.ClickCount >= 2 {
		fl.BaseView.HandleEvent(event) // click-to-focus
		if event.IsCleared() {
			return
		}

		idx := fl.ListViewer().TopIndex() + event.Mouse.Y
		entry := fl.dataSource().Entry(idx)

		if entry != nil && entry.IsDir {
			fl.ReadDirectory(entry.Path, fl.dataSource().Wildcard())
			fl.ListViewer().SetSelected(0)
			event.Clear()
			fl.broadcastFocused()
			return
		}

		if entry != nil && !entry.IsDir && entry.Name != ".." {
			fl.broadcastFocused()
			event.What = EvCommand
			event.Command = CmOK
			event.Mouse = nil
			return
		}
	}

	// 2. For keyboard events, check incremental search first
	if event.What == EvKeyboard && event.Key != nil {
		k := event.Key

		// Clear search on navigation keys BEFORE handling (so they still work)
		switch k.Key {
		case tcell.KeyUp, tcell.KeyDown, tcell.KeyPgUp, tcell.KeyPgDn,
			tcell.KeyHome, tcell.KeyEnd, tcell.KeyLeft, tcell.KeyRight:
			fl.searchBuf = nil
			fl.searchTime = time.Time{}
			// fall through to delegation
		}

		// Printable rune -> incremental search
		if k.Key == tcell.KeyRune && k.Rune != 0 {
			if '.' == k.Rune || !unicode.IsControl(k.Rune) {
				// 1-second timeout
				if !fl.searchTime.IsZero() && time.Since(fl.searchTime) > time.Second {
					fl.searchBuf = nil
				}
				fl.searchBuf = append(fl.searchBuf, k.Rune)
				fl.searchTime = time.Now()
				fl.incrementalSearch()
				event.Clear()
				fl.broadcastFocused()
				return
			}
		}

		// Backspace -> remove last search char
		if k.Key == tcell.KeyBackspace || k.Key == tcell.KeyBackspace2 {
			if len(fl.searchBuf) > 0 {
				fl.searchBuf = fl.searchBuf[:len(fl.searchBuf)-1]
				fl.searchTime = time.Now()
				fl.incrementalSearch()
			}
			event.Clear()
			fl.broadcastFocused()
			return
		}

		// Enter key -> navigate or CmOK
		if k.Key == tcell.KeyEnter {
			fl.BaseView.HandleEvent(event)
			if event.IsCleared() {
				return
			}

			idx := fl.Selected()
			entry := fl.dataSource().Entry(idx)

			if entry != nil && entry.IsDir {
				fl.ReadDirectory(entry.Path, fl.dataSource().Wildcard())
				fl.ListViewer().SetSelected(0)
				event.Clear()
				fl.broadcastFocused()
				return
			}

			if entry != nil && !entry.IsDir && entry.Name != ".." {
				fl.broadcastFocused()
				event.What = EvCommand
				event.Command = CmOK
				event.Key = nil
				return
			}
		}
	}

	// 3. Delegate to ListBox
	fl.ListBox.HandleEvent(event)

	// 4. Broadcast CmFileFocused after any keyboard or mouse event
	if wasKbOrMouse {
		fl.broadcastFocused()
	}
}

// broadcastFocused broadcasts a CmFileFocused event to the owner's children
// with the currently selected FileEntry as event.Info.
func (fl *FileList) broadcastFocused() {
	idx := fl.Selected()
	entry := fl.dataSource().Entry(idx)
	if entry == nil {
		return
	}
	owner := fl.Owner()
	if owner == nil {
		return
	}
	ev := &Event{What: EvBroadcast, Command: CmFileFocused, Info: entry}
	owner.HandleEvent(ev)
}

// incrementalSearch searches for the first entry whose name starts with the
// accumulated search buffer (case-insensitive) and selects it.
func (fl *FileList) incrementalSearch() {
	needle := strings.ToLower(string(fl.searchBuf))
	ds := fl.dataSource()
	for i := 0; i < ds.Count(); i++ {
		name := strings.ToLower(ds.entries[i].Name)
		if strings.HasPrefix(name, needle) {
			fl.ListViewer().SetSelected(i)
			return
		}
	}
}
