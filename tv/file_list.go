package tv

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
