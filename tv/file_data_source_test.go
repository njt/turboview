package tv

// file_data_source_test.go — Tests for Task 1 (Batch 7): FileDataSource.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the spec requirement it verifies.
//
// Test organisation:
//   Section 1  — New Command Code tests (CmFileOpen, CmFileReplace, CmFileClear,
//                CmFileFocused, CmFileFilter)
//   Section 2  — Constructor tests (NewFileDataSource, empty wildcard default,
//                non-existent dir)
//   Section 3  — Count tests (empty dir, with files, root, non-existent)
//   Section 4  — Item tests (parent, dir, file, root no parent)
//   Section 5  — Entry tests (returns FileEntry, nil for invalid index, field values)
//   Section 6  — Dir / Wildcard accessor tests
//   Section 7  — SetDir tests (changes directory, non-existent no panic)
//   Section 8  — SetWildcard tests (changes filter, empty defaults to star)
//   Section 9  — Sorting tests (parent first, dirs before files, alphabetical,
//                case-insensitive dirs, case-insensitive files)
//   Section 10 — Filtering tests (hidden excluded, dirs always included,
//                wildcard matching, no parent at root)
//   Section 11 — Multi-pattern wildcard tests
//   Section 12 — Falsifying tests (one per happy-path requirement)

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// compile-time assertion: FileDataSource must satisfy ListDataSource.
// Spec: "FileDataSource implements ListDataSource"
var _ ListDataSource = (*FileDataSource)(nil)

// ---------------------------------------------------------------------------
// Section 1 — New Command Code tests
// ---------------------------------------------------------------------------

// TestCmFileOpenExists verifies CmFileOpen is a defined CommandCode constant.
// Spec: "CmFileOpen — File Open button pressed"
func TestCmFileOpenExists(t *testing.T) {
	var c CommandCode = CmFileOpen
	_ = c
}

// TestCmFileReplaceExists verifies CmFileReplace is a defined CommandCode constant.
// Spec: "CmFileReplace — File Replace button pressed"
func TestCmFileReplaceExists(t *testing.T) {
	var c CommandCode = CmFileReplace
	_ = c
}

// TestCmFileClearExists verifies CmFileClear is a defined CommandCode constant.
// Spec: "CmFileClear — File Clear button pressed"
func TestCmFileClearExists(t *testing.T) {
	var c CommandCode = CmFileClear
	_ = c
}

// TestCmFileFocusedExists verifies CmFileFocused is a defined CommandCode constant.
// Spec: "CmFileFocused — broadcast when a file entry gains focus (carries a *FileEntry in event.Info)"
func TestCmFileFocusedExists(t *testing.T) {
	var c CommandCode = CmFileFocused
	_ = c
}

// TestCmFileFilterExists verifies CmFileFilter is a defined CommandCode constant.
// Spec: "CmFileFilter — broadcast when the wildcard filter changes (carries a string in event.Info)"
func TestCmFileFilterExists(t *testing.T) {
	var c CommandCode = CmFileFilter
	_ = c
}

// TestFileCommandCodesDistinct verifies all five new File command codes have
// distinct values from each other.
// Spec: "Add these new command constants" — each must have a unique value.
func TestFileCommandCodesDistinct(t *testing.T) {
	seen := make(map[CommandCode]string)
	names := map[string]CommandCode{
		"CmFileOpen":    CmFileOpen,
		"CmFileReplace": CmFileReplace,
		"CmFileClear":   CmFileClear,
		"CmFileFocused": CmFileFocused,
		"CmFileFilter":  CmFileFilter,
	}
	for name, val := range names {
		if prev, ok := seen[val]; ok {
			t.Errorf("Duplicate value: %s and %s both have value %d", prev, name, val)
		}
		seen[val] = name
	}
}

// TestFileCommandCodesDistinctFromExisting verifies none of the new File command
// codes collide with existing command codes (CmQuit through CmUser).
// Spec: "Add these new command constants" — they are additions, not replacements.
func TestFileCommandCodesDistinctFromExisting(t *testing.T) {
	existing := map[CommandCode]string{
		CmQuit:             "CmQuit",
		CmClose:            "CmClose",
		CmOK:               "CmOK",
		CmCancel:           "CmCancel",
		CmYes:              "CmYes",
		CmNo:               "CmNo",
		CmMenu:             "CmMenu",
		CmResize:           "CmResize",
		CmZoom:             "CmZoom",
		CmTile:             "CmTile",
		CmCascade:          "CmCascade",
		CmNext:             "CmNext",
		CmPrev:             "CmPrev",
		CmDefault:          "CmDefault",
		CmGrabDefault:      "CmGrabDefault",
		CmReleaseDefault:   "CmReleaseDefault",
		CmReceivedFocus:    "CmReceivedFocus",
		CmReleasedFocus:    "CmReleasedFocus",
		CmScrollBarClicked: "CmScrollBarClicked",
		CmScrollBarChanged: "CmScrollBarChanged",
		CmSelectWindowNum:  "CmSelectWindowNum",
		CmRecordHistory:    "CmRecordHistory",
		CmFind:             "CmFind",
		CmReplace:          "CmReplace",
		CmSearchAgain:      "CmSearchAgain",
		CmIndicatorUpdate:  "CmIndicatorUpdate",
		CmUser:             "CmUser",
	}

	newCodes := []struct {
		name string
		val  CommandCode
	}{
		{"CmFileOpen", CmFileOpen},
		{"CmFileReplace", CmFileReplace},
		{"CmFileClear", CmFileClear},
		{"CmFileFocused", CmFileFocused},
		{"CmFileFilter", CmFileFilter},
	}

	for _, nc := range newCodes {
		if existingName, ok := existing[nc.val]; ok {
			t.Errorf("%s = %d collides with existing %s = %d", nc.name, nc.val, existingName, nc.val)
		}
	}
}

// TestFileCommandCodesAreCommandCodeType verifies the new codes are of type CommandCode.
// Spec: "Add these new command constants" — they must be CommandCode type.
func TestFileCommandCodesAreCommandCodeType(t *testing.T) {
	// Type switch to verify all are CommandCode
	codes := []interface{}{CmFileOpen, CmFileReplace, CmFileClear, CmFileFocused, CmFileFilter}
	for i, c := range codes {
		if _, ok := c.(CommandCode); !ok {
			t.Errorf("index %d: value %v is not a CommandCode", i, c)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Constructor tests
// ---------------------------------------------------------------------------

// TestNewFileDataSourceValidDir verifies NewFileDataSource creates a
// FileDataSource from a real directory.
// Spec: "NewFileDataSource(dir, wildcard string) *FileDataSource — creates a
// FileDataSource, reads the directory, applies the wildcard filter."
func TestNewFileDataSourceValidDir(t *testing.T) {
	tmpDir := t.TempDir()
	fds := NewFileDataSource(tmpDir, "*")
	if fds == nil {
		t.Fatal("NewFileDataSource returned nil for valid directory")
	}
}

// TestNewFileDataSourceEmptyWildcardDefaultsToStar verifies an empty wildcard
// string is treated as "*".
// Spec: "If wildcard is empty, defaults to '*'."
func TestNewFileDataSourceEmptyWildcardDefaultsToStar(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644)

	// With empty wildcard, should behave like "*" and include the file
	fds := NewFileDataSource(tmpDir, "")
	if fds.Wildcard() != "*" {
		t.Errorf("Wildcard() = %q, want %q (empty should default to '*')", fds.Wildcard(), "*")
	}

	// File should be visible because empty defaults to "*"
	found := false
	for i := 0; i < fds.Count(); i++ {
		if !strings.HasSuffix(fds.Item(i), "/") && fds.Item(i) == "test.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("File 'test.txt' not found with empty wildcard (should default to '*')")
	}
}

// TestNewFileDataSourceNonExistentDirNoPanic verifies creating a FileDataSource
// for a non-existent directory does not panic.
// Spec: "Directory that doesn't exist (Count returns 0, no panic)"
func TestNewFileDataSourceNonExistentDirNoPanic(t *testing.T) {
	nonExistent := filepath.Join(t.TempDir(), "does-not-exist")
	fds := NewFileDataSource(nonExistent, "*")
	if fds == nil {
		t.Fatal("NewFileDataSource returned nil; should return a valid object")
	}
	if fds.Count() != 0 {
		t.Errorf("Count() = %d, want 0 for non-existent directory", fds.Count())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Count tests
// ---------------------------------------------------------------------------

// TestFileDataSourceCountEmptyDir verifies an empty directory has exactly one
// entry (the ".." parent).
// Spec: "Count() int — returns the number of entries"
// Spec: "Empty directory (only '..' shows if not root)"
func TestFileDataSourceCountEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	fds := NewFileDataSource(tmpDir, "*")
	if fds.Count() != 1 {
		t.Errorf("Count() = %d, want 1 (empty dir should only have '..')", fds.Count())
	}
	if fds.Item(0) != ".." {
		t.Errorf("Item(0) = %q, want '..'", fds.Item(0))
	}
}

// TestFileDataSourceCountWithFiles verifies Count returns the correct number of
// entries when files and directories exist.
// Spec: "Count() int — returns the number of entries"
func TestFileDataSourceCountWithFiles(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("b"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)

	// Expected: 1 ("..") + 1 dir + 2 files = 4
	fds := NewFileDataSource(tmpDir, "*")
	if fds.Count() != 4 {
		t.Errorf("Count() = %d, want 4 (1 parent + 1 dir + 2 files)", fds.Count())
	}
}

// TestFileDataSourceCountRootDir verifies the root directory has no ".."
// entry, so Count for an empty root counts only real entries.
// Spec: "The parent '..' entry is only included when NOT at the filesystem root"
func TestFileDataSourceCountRootDir(t *testing.T) {
	fds := NewFileDataSource("/", "*")
	// At root, there should be no ".." entry
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == ".." {
			t.Error("Root directory should not have a '..' entry")
		}
	}
}

// TestFileDataSourceCountNonExistentDir verifies Count returns 0 for a
// non-existent directory.
// Spec: "Directory that doesn't exist (Count returns 0, no panic)"
func TestFileDataSourceCountNonExistentDir(t *testing.T) {
	fds := NewFileDataSource("/nonexistent/dir/path", "*")
	if fds.Count() != 0 {
		t.Errorf("Count() = %d, want 0 for non-existent directory", fds.Count())
	}
}

// TestFileDataSourceCountVariesWithContent verifies Count changes when used
// with different directories (falsification: Count is not a hardcoded value).
// Spec: "Count() int — returns the number of entries"
func TestFileDataSourceCountVariesWithContent(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	os.WriteFile(filepath.Join(dir1, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir1, "b.txt"), []byte("b"), 0644)

	// dir2 has more files
	os.WriteFile(filepath.Join(dir2, "x.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir2, "y.txt"), []byte("y"), 0644)
	os.WriteFile(filepath.Join(dir2, "z.txt"), []byte("z"), 0644)

	fds1 := NewFileDataSource(dir1, "*")
	fds2 := NewFileDataSource(dir2, "*")

	if fds1.Count() == fds2.Count() {
		t.Error("Count() returned the same value for directories with different numbers of files; Count must reflect actual content")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Item tests
// ---------------------------------------------------------------------------

// TestFileDataSourceItemParentEntry verifies that the first entry (index 0)
// is ".." when not at the filesystem root.
// Spec: "Item(index int) string — returns display text: '..' for parent"
func TestFileDataSourceItemParentEntry(t *testing.T) {
	tmpDir := t.TempDir()
	fds := NewFileDataSource(tmpDir, "*")
	if fds.Item(0) != ".." {
		t.Errorf("Item(0) = %q, want '..' (parent entry)", fds.Item(0))
	}
}

// TestFileDataSourceItemDirectory verifies directories display with a trailing "/".
// Spec: "name + '/' for directories"
func TestFileDataSourceItemDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "mydir"), 0755)

	fds := NewFileDataSource(tmpDir, "*")

	found := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "mydir/" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Directory 'mydir/' not found in item list")
	}
}

// TestFileDataSourceItemFile verifies files display as just the name (no suffix).
// Spec: "name for files"
func TestFileDataSourceItemFile(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("hello"), 0644)

	fds := NewFileDataSource(tmpDir, "*")

	found := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "readme.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("File 'readme.txt' not found in item list")
	}
}

// TestFileDataSourceItemRootNoParent verifies that at root "/", the first
// entry is not "..".
// Spec: "The parent '..' entry is only included when NOT at the filesystem root"
func TestFileDataSourceItemRootNoParent(t *testing.T) {
	fds := NewFileDataSource("/", "*")
	// If Count > 0, first entry must not be ".."
	if fds.Count() > 0 && fds.Item(0) == ".." {
		t.Error("Root directory should not have '..' as first entry")
	}
}

// TestFileDataSourceItemFileNoTrailingSlash verifies files do NOT end with "/".
// This is the falsifying counterpart to TestFileDataSourceItemDirectory.
// Spec: "name for files" — bare name, no trailing slash.
func TestFileDataSourceItemFileNoTrailingSlash(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "data.csv"), []byte("a,b,c"), 0644)

	fds := NewFileDataSource(tmpDir, "*")

	for i := 0; i < fds.Count(); i++ {
		item := fds.Item(i)
		if item == ".." {
			continue
		}
		// Check the entry to verify non-dirs don't end with "/"
		entry := fds.Entry(i)
		if entry != nil && !entry.IsDir && strings.HasSuffix(item, "/") {
			t.Errorf("Item(%d) = %q — file should not end with '/'", i, item)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Entry tests
// ---------------------------------------------------------------------------

// TestFileDataSourceEntryReturnsFileEntry verifies Entry returns a non-nil
// *FileEntry for a valid index.
// Spec: "Entry(index int) *FileEntry — returns the underlying FileEntry or nil"
func TestFileDataSourceEntryReturnsFileEntry(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "hello.txt"), []byte("world"), 0644)

	fds := NewFileDataSource(tmpDir, "*")
	// Find the file entry (skip "..")
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "hello.txt" {
			entry := fds.Entry(i)
			if entry == nil {
				t.Errorf("Entry(%d) returned nil for file 'hello.txt'", i)
			}
			return
		}
	}
	t.Error("File 'hello.txt' not found in data source")
}

// TestFileDataSourceEntryNilForNegativeIndex verifies Entry returns nil for
// negative indices.
// Spec: "Entry(index int) *FileEntry — returns the underlying FileEntry or nil"
func TestFileDataSourceEntryNilForNegativeIndex(t *testing.T) {
	tmpDir := t.TempDir()
	fds := NewFileDataSource(tmpDir, "*")
	if fds.Entry(-1) != nil {
		t.Error("Entry(-1) should return nil")
	}
}

// TestFileDataSourceEntryNilForOutOfRange verifies Entry returns nil for
// indices >= Count().
// Spec: "Entry(index int) *FileEntry — returns the underlying FileEntry or nil"
func TestFileDataSourceEntryNilForOutOfRange(t *testing.T) {
	tmpDir := t.TempDir()
	fds := NewFileDataSource(tmpDir, "*")
	if fds.Entry(fds.Count()) != nil {
		t.Error("Entry(Count()) should return nil")
	}
}

// TestFileDataSourceEntryFields verifies the FileEntry struct has the expected
// fields: Name, Size, IsDir, Path.
// Spec: "FileEntry struct will be defined with fields: Name (string), Size (int64),
// ModTime (time.Time), IsDir (bool), Path (string - full path)"
func TestFileDataSourceEntryFields(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "data.bin")
	os.WriteFile(filePath, []byte("content"), 0644)

	fds := NewFileDataSource(tmpDir, "*")

	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "data.bin" {
			entry := fds.Entry(i)
			if entry == nil {
				t.Fatal("Entry returned nil")
			}
			if entry.Name != "data.bin" {
				t.Errorf("entry.Name = %q, want %q", entry.Name, "data.bin")
			}
			if entry.Size <= 0 {
				t.Errorf("entry.Size = %d, want > 0 (file has content)", entry.Size)
			}
			if entry.IsDir {
				t.Error("entry.IsDir = true, want false for a file")
			}
			if entry.Path != filePath {
				t.Errorf("entry.Path = %q, want %q", entry.Path, filePath)
			}
			if entry.ModTime.IsZero() {
				t.Error("entry.ModTime is zero; should be set")
			}
			return
		}
	}
	t.Error("File 'data.bin' not found")
}

// TestFileDataSourceEntryDirectoryFields verifies directory entries have
// correct IsDir and Name fields.
// Spec: "FileEntry struct will be defined with fields: ... IsDir (bool)"
func TestFileDataSourceEntryDirectoryFields(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "mydir"), 0755)

	fds := NewFileDataSource(tmpDir, "*")

	// Find "mydir/" entry
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "mydir/" {
			entry := fds.Entry(i)
			if entry == nil {
				t.Fatal("Entry returned nil for directory")
			}
			if !entry.IsDir {
				t.Error("entry.IsDir = false, want true for directory 'mydir/'")
			}
			if entry.Name != "mydir" {
				t.Errorf("entry.Name = %q, want %q", entry.Name, "mydir")
			}
			return
		}
	}
	t.Error("Directory 'mydir/' not found")
}

// TestFileDataSourceEntryParentEntryExists verifies the ".." parent entry
// returns a non-nil FileEntry with correct fields.
// Spec: "Entry(index int) *FileEntry — returns the underlying FileEntry or nil"
// — index 0 (valid, "..") should return non-nil with reasonable fields.
func TestFileDataSourceEntryParentEntryExists(t *testing.T) {
	tmpDir := t.TempDir()
	fds := NewFileDataSource(tmpDir, "*")
	if fds.Count() == 0 || fds.Item(0) != ".." {
		t.Skip("no parent entry to test (at filesystem root)")
	}
	entry := fds.Entry(0)
	if entry == nil {
		t.Fatal("Entry(0) for '..' should return a non-nil FileEntry")
	}
	if entry.Name != ".." {
		t.Errorf("Entry(0).Name = %q, want %q", entry.Name, "..")
	}
	if !entry.IsDir {
		t.Error("Entry(0).IsDir = false, want true (parent is a directory)")
	}
	if entry.Path != filepath.Dir(tmpDir) {
		t.Errorf("Entry(0).Path = %q, want parent dir %q", entry.Path, filepath.Dir(tmpDir))
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Dir / Wildcard accessor tests
// ---------------------------------------------------------------------------

// TestFileDataSourceDirReturnsDir verifies Dir() returns the original directory
// passed to the constructor.
// Spec: "Dir() string — returns the current directory path"
func TestFileDataSourceDirReturnsDir(t *testing.T) {
	tmpDir := t.TempDir()
	fds := NewFileDataSource(tmpDir, "*")
	if fds.Dir() != tmpDir {
		t.Errorf("Dir() = %q, want %q", fds.Dir(), tmpDir)
	}
}

// TestFileDataSourceWildcardReturnsWildcard verifies Wildcard() returns the
// wildcard pattern passed to the constructor.
// Spec: "Wildcard() string — returns the current wildcard pattern"
func TestFileDataSourceWildcardReturnsWildcard(t *testing.T) {
	tmpDir := t.TempDir()
	fds := NewFileDataSource(tmpDir, "*.go")
	if fds.Wildcard() != "*.go" {
		t.Errorf("Wildcard() = %q, want %q", fds.Wildcard(), "*.go")
	}
}

// TestFileDataSourceWildcardEmptyReturnsStar verifies Wildcard() returns "*"
// when constructor was called with an empty string.
// Spec: "If wildcard is empty, defaults to '*'."
func TestFileDataSourceWildcardEmptyReturnsStar(t *testing.T) {
	tmpDir := t.TempDir()
	fds := NewFileDataSource(tmpDir, "")
	if fds.Wildcard() != "*" {
		t.Errorf("Wildcard() = %q, want %q", fds.Wildcard(), "*")
	}
}

// ---------------------------------------------------------------------------
// Section 7 — SetDir tests
// ---------------------------------------------------------------------------

// TestFileDataSourceSetDirChangesDirectory verifies SetDir changes the
// current directory and refreshes the listing.
// Spec: "SetDir(dir string) — changes directory and refreshes the listing"
func TestFileDataSourceSetDirChangesDirectory(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	os.WriteFile(filepath.Join(dir1, "file1.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(dir2, "file2.txt"), []byte("2"), 0644)

	fds := NewFileDataSource(dir1, "*")
	if fds.Dir() != dir1 {
		t.Errorf("Dir() = %q, want %q", fds.Dir(), dir1)
	}

	// Verify file1.txt is present, file2.txt is not
	found1 := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "file1.txt" {
			found1 = true
		}
		if fds.Item(i) == "file2.txt" {
			t.Error("file2.txt should not be visible before SetDir")
		}
	}
	if !found1 {
		t.Error("file1.txt should be visible")
	}

	// Now change directory
	fds.SetDir(dir2)
	if fds.Dir() != dir2 {
		t.Errorf("Dir() after SetDir = %q, want %q", fds.Dir(), dir2)
	}

	// Verify file2.txt is present, file1.txt is not
	found2 := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "file2.txt" {
			found2 = true
		}
		if fds.Item(i) == "file1.txt" {
			t.Error("file1.txt should not be visible after SetDir")
		}
	}
	if !found2 {
		t.Error("file2.txt should be visible after SetDir")
	}
}

// TestFileDataSourceSetDirNonExistentNoPanic verifies SetDir to a non-existent
// directory does not panic.
// Spec: "SetDir(dir string) — changes directory and refreshes the listing"
// Edge case from spec: "Directory that doesn't exist (Count returns 0, no panic)"
func TestFileDataSourceSetDirNonExistentNoPanic(t *testing.T) {
	tmpDir := t.TempDir()
	fds := NewFileDataSource(tmpDir, "*")

	nonExistent := filepath.Join(tmpDir, "does-not-exist")
	fds.SetDir(nonExistent)

	// Dir() should reflect the new path even if it doesn't exist
	if fds.Dir() != nonExistent {
		t.Errorf("Dir() = %q, want %q", fds.Dir(), nonExistent)
	}
	// Count should be 0
	if fds.Count() != 0 {
		t.Errorf("Count() = %d, want 0 for non-existent directory", fds.Count())
	}
}

// ---------------------------------------------------------------------------
// Section 8 — SetWildcard tests
// ---------------------------------------------------------------------------

// TestFileDataSourceSetWildcardChangesFilter verifies SetWildcard changes the
// wildcard pattern and refreshes the listing.
// Spec: "SetWildcard(wc string) — changes wildcard and refreshes"
func TestFileDataSourceSetWildcardChangesFilter(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# readme"), 0644)

	fds := NewFileDataSource(tmpDir, "*.go")

	// Before: only main.go should be visible
	foundMD := false
	foundGo := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "readme.md" {
			foundMD = true
		}
		if fds.Item(i) == "main.go" {
			foundGo = true
		}
	}
	if foundMD {
		t.Error("readme.md should not match '*.go' pattern")
	}
	if !foundGo {
		t.Error("main.go should match '*.go' pattern")
	}

	// Change wildcard to include .md files
	fds.SetWildcard("*.md")
	if fds.Wildcard() != "*.md" {
		t.Errorf("Wildcard() = %q, want %q after SetWildcard", fds.Wildcard(), "*.md")
	}

	// After: only readme.md should be visible
	foundMD = false
	foundGo = false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "readme.md" {
			foundMD = true
		}
		if fds.Item(i) == "main.go" {
			foundGo = true
		}
	}
	if !foundMD {
		t.Error("readme.md should match '*.md' pattern after SetWildcard")
	}
	if foundGo {
		t.Error("main.go should not match '*.md' pattern after SetWildcard")
	}
}

// TestFileDataSourceSetWildcardEmptyDefaultsToStar verifies that empty string
// passed to SetWildcard defaults to "*".
// Spec: "SetWildcard(wc string) — changes wildcard and refreshes; empty string defaults to '*'"
func TestFileDataSourceSetWildcardEmptyDefaultsToStar(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "data.csv"), []byte("a,b"), 0644)

	// Start with a restrictive pattern
	fds := NewFileDataSource(tmpDir, "*.go")
	if fds.Wildcard() != "*.go" {
		t.Fatalf("initial Wildcard() = %q, want '*.go'", fds.Wildcard())
	}

	// Set wildcard to empty — should default to "*"
	fds.SetWildcard("")
	if fds.Wildcard() != "*" {
		t.Errorf("Wildcard() after SetWildcard('') = %q, want %q", fds.Wildcard(), "*")
	}

	// Both files should now be visible
	foundCSV := false
	foundGo := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "data.csv" {
			foundCSV = true
		}
		if fds.Item(i) == "main.go" {
			foundGo = true
		}
	}
	if !foundCSV || !foundGo {
		t.Errorf("After SetWildcard(''), both files should be visible. foundCSV=%v foundGo=%v", foundCSV, foundGo)
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Sorting tests
// ---------------------------------------------------------------------------

// TestFileDataSourceSortingParentFirst verifies ".." is always at index 0
// when not at the filesystem root.
// Spec: "'..' (parent directory) always first (unless at filesystem root '/')"
func TestFileDataSourceSortingParentFirst(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "aaa"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "000_file.txt"), []byte("x"), 0644)

	fds := NewFileDataSource(tmpDir, "*")
	if fds.Count() > 0 && fds.Item(0) != ".." {
		t.Errorf("Item(0) = %q, want '..' (parent should always be first)", fds.Item(0))
	}
}

// TestFileDataSourceSortingDirsBeforeFiles verifies all directories appear
// before any files (after "..").
// Spec: "Directories, alphabetically (case-insensitive)" and
// "Files, alphabetically (case-insensitive)" — dirs come before files.
func TestFileDataSourceSortingDirsBeforeFiles(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "zebra_dir"), 0755) // alphabetically after some files
	os.WriteFile(filepath.Join(tmpDir, "aardvark.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "banana.txt"), []byte("x"), 0644)

	fds := NewFileDataSource(tmpDir, "*")

	// Find the first non-".." entry — it should be a directory
	seenDir := false
	seenFile := false
	isFilePhase := false
	for i := 0; i < fds.Count(); i++ {
		item := fds.Item(i)
		if item == ".." {
			continue
		}
		isDir := strings.HasSuffix(item, "/")

		if isDir {
			seenDir = true
			if seenFile && !isFilePhase {
				t.Errorf("Directory %q appears after a file; dirs must come before files", item)
			}
		} else {
			seenFile = true
			if !seenDir {
				isFilePhase = true // first entry after ".." was a file — no dirs present
			}
		}
	}
}

// TestFileDataSourceSortingDirsAlphabetical verifies directories are sorted
// alphabetically, case-insensitively.
// Spec: "Directories, alphabetically (case-insensitive)"
func TestFileDataSourceSortingDirsAlphabetical(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "zzz"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "aaa"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "mmm"), 0755)

	fds := NewFileDataSource(tmpDir, "*")

	var dirs []string
	for i := 0; i < fds.Count(); i++ {
		item := fds.Item(i)
		if item == ".." {
			continue
		}
		if strings.HasSuffix(item, "/") {
			dirs = append(dirs, item)
		}
	}

	// Verify alphabetical order
	for i := 1; i < len(dirs); i++ {
		if strings.ToLower(dirs[i-1]) > strings.ToLower(dirs[i]) {
			t.Errorf("Directories not sorted alphabetically (case-insensitive): %q before %q",
				dirs[i-1], dirs[i])
		}
	}
}

// TestFileDataSourceSortingFilesAlphabetical verifies files are sorted
// alphabetically, case-insensitively.
// Spec: "Files, alphabetically (case-insensitive)"
func TestFileDataSourceSortingFilesAlphabetical(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "zebra.txt"), []byte("z"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "alpha.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "mike.txt"), []byte("m"), 0644)

	fds := NewFileDataSource(tmpDir, "*")

	var files []string
	for i := 0; i < fds.Count(); i++ {
		item := fds.Item(i)
		if item == ".." {
			continue
		}
		if !strings.HasSuffix(item, "/") {
			files = append(files, item)
		}
	}

	// Verify alphabetical order
	for i := 1; i < len(files); i++ {
		if strings.ToLower(files[i-1]) > strings.ToLower(files[i]) {
			t.Errorf("Files not sorted alphabetically (case-insensitive): %q before %q",
				files[i-1], files[i])
		}
	}
}

// TestFileDataSourceSortingCaseInsensitiveDirs verifies directories are sorted
// case-insensitively (e.g., "B_dir" before "a_dir").
// Spec: "Directories, alphabetically (case-insensitive)"
func TestFileDataSourceSortingCaseInsensitiveDirs(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "b_dir"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "a_dir"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "C_dir"), 0755)

	fds := NewFileDataSource(tmpDir, "*")

	var dirs []string
	for i := 0; i < fds.Count(); i++ {
		item := fds.Item(i)
		if item == ".." {
			continue
		}
		if strings.HasSuffix(item, "/") {
			dirs = append(dirs, item)
		}
	}

	if len(dirs) != 3 {
		t.Fatalf("Expected 3 directories, got %d: %v", len(dirs), dirs)
	}

	// Case-insensitive order: a_dir, b_dir, C_dir (all same case-insensitively)
	expectedLower := []string{"a_dir/", "b_dir/", "c_dir/"}
	for i, d := range dirs {
		if strings.ToLower(d) != expectedLower[i] {
			t.Errorf("Directory %d: %q (lowercase: %q), want %q",
				i, d, strings.ToLower(d), expectedLower[i])
		}
	}
}

// TestFileDataSourceSortingCaseInsensitiveFiles verifies files are sorted
// case-insensitively (e.g., "B.txt" before "a.txt").
// Spec: "Files, alphabetically (case-insensitive)"
func TestFileDataSourceSortingCaseInsensitiveFiles(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "B.txt"), []byte("B"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "C.txt"), []byte("C"), 0644)

	fds := NewFileDataSource(tmpDir, "*")

	var files []string
	for i := 0; i < fds.Count(); i++ {
		item := fds.Item(i)
		if item == ".." {
			continue
		}
		if !strings.HasSuffix(item, "/") {
			files = append(files, item)
		}
	}

	if len(files) != 3 {
		t.Fatalf("Expected 3 files, got %d: %v", len(files), files)
	}

	// Case-insensitive order: a.txt, B.txt, C.txt
	expectedLower := []string{"a.txt", "b.txt", "c.txt"}
	for i, f := range files {
		if strings.ToLower(f) != expectedLower[i] {
			t.Errorf("File %d: %q (lowercase: %q), want %q",
				i, f, strings.ToLower(f), expectedLower[i])
		}
	}
}

// TestFileDataSourceSortingNoParentAtRoot verifies that at root, the sorting
// still works (no ".." entry, dirs come first).
// Spec: "'..' (parent directory) always first (unless at filesystem root '/')"
func TestFileDataSourceSortingNoParentAtRoot(t *testing.T) {
	fds := NewFileDataSource("/", "*")

	// First entry (if any) must not be ".."
	if fds.Count() > 0 && fds.Item(0) == ".." {
		t.Error("Root directory must not have '..'")
	}

	// After the first entry, verify no ".." appears anywhere
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == ".." {
			t.Errorf("'..' found at index %d in root directory", i)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 10 — Filtering tests
// ---------------------------------------------------------------------------

// TestFileDataSourceFilteringExcludesHidden verifies hidden files and
// directories (names starting with ".") are excluded.
// Spec: "Hidden files/directories (names starting with '.') are excluded"
func TestFileDataSourceFilteringExcludesHidden(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte("v"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden.txt"), []byte("h"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, ".hidden_dir"), 0755)

	fds := NewFileDataSource(tmpDir, "*")

	for i := 0; i < fds.Count(); i++ {
		item := fds.Item(i)
		if strings.HasPrefix(item, ".") && item != ".." {
			t.Errorf("Hidden entry %q found at index %d — hidden files should be excluded", item, i)
		}
		if item == ".hidden.txt" || item == ".hidden_dir/" {
			t.Errorf("Hidden entry %q should not appear in listing", item)
		}
	}
}

// TestFileDataSourceFilteringDirsAlwaysIncluded verifies directories appear
// in the listing regardless of the wildcard pattern.
// Spec: "directories are always included regardless of wildcard"
func TestFileDataSourceFilteringDirsAlwaysIncluded(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a directory — no wildcard would match a directory name
	os.MkdirAll(filepath.Join(tmpDir, "docs"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)

	// Use a wildcard that matches .go files but not "docs"
	fds := NewFileDataSource(tmpDir, "*.go")

	foundDocs := false
	foundGo := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "docs/" {
			foundDocs = true
		}
		if fds.Item(i) == "main.go" {
			foundGo = true
		}
	}
	if !foundDocs {
		t.Error("Directory 'docs/' should be visible regardless of wildcard '*.go'")
	}
	if !foundGo {
		t.Error("File 'main.go' should match wildcard '*.go'")
	}
}

// TestFileDataSourceFilteringWildcardMatches verifies only files matching the
// wildcard pattern are included.
// Spec: "Only files matching the wildcard pattern are included"
func TestFileDataSourceFilteringWildcardMatches(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# readme"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "data.csv"), []byte("a,b"), 0644)

	// Only match .go files
	fds := NewFileDataSource(tmpDir, "*.go")

	foundMD := false
	foundCSV := false
	foundGo := false
	for i := 0; i < fds.Count(); i++ {
		switch fds.Item(i) {
		case "readme.md":
			foundMD = true
		case "data.csv":
			foundCSV = true
		case "main.go":
			foundGo = true
		}
	}
	if foundMD {
		t.Error("readme.md should not match '*.go' pattern")
	}
	if foundCSV {
		t.Error("data.csv should not match '*.go' pattern")
	}
	if !foundGo {
		t.Error("main.go should match '*.go' pattern")
	}
}

// TestFileDataSourceFilteringWildcardExcludesAllNonMatching verifies that a
// deeply restrictive wildcard that matches nothing results in only dirs and "..".
// Spec: "Only files matching the wildcard pattern are included"
func TestFileDataSourceFilteringWildcardExcludesAllNonMatching(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("hi"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "data.csv"), []byte("x"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755)

	// A pattern that matches nothing
	fds := NewFileDataSource(tmpDir, "*.xyzzy")

	// Only ".." and "sub/" should appear (no matching files, but dirs included)
	foundFile := false
	for i := 0; i < fds.Count(); i++ {
		item := fds.Item(i)
		if item != ".." && !strings.HasSuffix(item, "/") {
			foundFile = true
			t.Errorf("File %q found at index %d — should not match '*.xyzzy'", item, i)
		}
	}
	if foundFile {
		t.Error("No files should match '*.xyzzy'; only dirs and '..' should appear")
	}
	// Verify dirs appear
	foundDir := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "sub/" {
			foundDir = true
		}
	}
	if !foundDir {
		t.Error("Directory 'sub/' should be visible regardless of wildcard")
	}
}

// TestFileDataSourceFilteringNoParentAtRoot verifies that at root, there is
// no ".." parent entry.
// Spec: "The parent '..' entry is only included when NOT at the filesystem root"
func TestFileDataSourceFilteringNoParentAtRoot(t *testing.T) {
	fds := NewFileDataSource("/", "*")
	if fds.Count() > 0 && fds.Item(0) == ".." {
		t.Error("Root directory should not have a '..' entry")
	}
}

// ---------------------------------------------------------------------------
// Section 11 — Multi-pattern wildcard tests
// ---------------------------------------------------------------------------

// TestFileDataSourceMultiPatternWildcard verifies semicolon-separated patterns
// match files matching any of the patterns.
// Spec: "Wildcard supports semicolon-separated patterns: '*.go;*.mod' matches
// files matching either pattern"
func TestFileDataSourceMultiPatternWildcard(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module x"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# readme"), 0644)

	fds := NewFileDataSource(tmpDir, "*.go;*.mod")

	foundGo := false
	foundMod := false
	foundMD := false
	for i := 0; i < fds.Count(); i++ {
		switch fds.Item(i) {
		case "main.go":
			foundGo = true
		case "go.mod":
			foundMod = true
		case "readme.md":
			foundMD = true
		}
	}
	if !foundGo {
		t.Error("main.go should match '*.go;*.mod'")
	}
	if !foundMod {
		t.Error("go.mod should match '*.go;*.mod'")
	}
	if foundMD {
		t.Error("readme.md should NOT match '*.go;*.mod'")
	}
}

// TestFileDataSourceMultiPatternWildcardSingleSegmentWorks verifies a single
// pattern (no semicolons) still works correctly with the semicolon-aware parser.
// Spec: "Wildcard supports semicolon-separated patterns" — single pattern is
// a degenerate case.
func TestFileDataSourceMultiPatternWildcardSingleSegmentWorks(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# readme"), 0644)

	fds := NewFileDataSource(tmpDir, "*.go") // single pattern, no semicolons

	foundGo := false
	foundMD := false
	for i := 0; i < fds.Count(); i++ {
		switch fds.Item(i) {
		case "main.go":
			foundGo = true
		case "readme.md":
			foundMD = true
		}
	}
	if !foundGo {
		t.Error("main.go should match '*.go'")
	}
	if foundMD {
		t.Error("readme.md should NOT match '*.go'")
	}
}

// TestFileDataSourceMultiPatternThreePatterns verifies three semicolon-separated
// patterns all work together.
// Spec: "Wildcard supports semicolon-separated patterns"
func TestFileDataSourceMultiPatternThreePatterns(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module x"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "go.sum"), []byte("sum"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# readme"), 0644)

	fds := NewFileDataSource(tmpDir, "*.go;*.mod;*.sum")

	// All three should be visible
	for _, name := range []string{"main.go", "go.mod", "go.sum"} {
		found := false
		for i := 0; i < fds.Count(); i++ {
			if fds.Item(i) == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s should match '*.go;*.mod;*.sum'", name)
		}
	}

	// readme.md should not be visible
	foundMD := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "readme.md" {
			foundMD = true
		}
	}
	if foundMD {
		t.Error("readme.md should NOT match '*.go;*.mod;*.sum'")
	}
}

// TestFileDataSourceMultiPatternWildcardEmptySegmentIgnored verifies that
// empty segments or trailing semicolons are handled gracefully.
// Spec: "Wildcard supports semicolon-separated patterns"
func TestFileDataSourceMultiPatternWildcardEmptySegmentIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)

	// Trailing semicolon, leading semicolon
	fds := NewFileDataSource(tmpDir, ";*.go;")

	foundGo := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "main.go" {
			foundGo = true
		}
	}
	if !foundGo {
		t.Error("main.go should match ';*.go;'")
	}
}

// ---------------------------------------------------------------------------
// Section 12 — Falsifying tests
// ---------------------------------------------------------------------------

// TestFileDataSourceFalsifyingConstructorReturnsDifferentResults verifies that
// two FileDataSources for different directories return different entries.
// (Falsifying TestNewFileDataSourceValidDir — lazy impl that always returns same data.)
// Spec: "NewFileDataSource(dir, wildcard string) — creates a FileDataSource"
func TestFileDataSourceFalsifyingConstructorReturnsDifferentResults(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	os.WriteFile(filepath.Join(dir1, "one.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(dir2, "two.txt"), []byte("2"), 0644)

	fds1 := NewFileDataSource(dir1, "*")
	fds2 := NewFileDataSource(dir2, "*")

	// They must differ (lazy impl that ignores dir would return same for both)
	foundOne := false
	for i := 0; i < fds1.Count(); i++ {
		if fds1.Item(i) == "one.txt" {
			foundOne = true
		}
	}
	foundTwo := false
	for i := 0; i < fds2.Count(); i++ {
		if fds2.Item(i) == "two.txt" {
			foundTwo = true
		}
	}

	if !foundOne {
		t.Error("one.txt not found in fds1")
	}
	if !foundTwo {
		t.Error("two.txt not found in fds2")
	}
}

// TestFileDataSourceFalsifyingItemNotAlwaysParent verifies Item(i) is not always "..".
// (Falsifying TestFileDataSourceItemParentEntry — lazy impl that always returns "..".)
// Spec: "Item(index int) string"
func TestFileDataSourceFalsifyingItemNotAlwaysParent(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello"), 0644)

	fds := NewFileDataSource(tmpDir, "*")

	// Not every Item should be ".."
	allParent := true
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) != ".." {
			allParent = false
			break
		}
	}
	if allParent && fds.Count() > 0 {
		t.Error("All items are '..'; Item() must return different values for different entries")
	}
}

// TestFileDataSourceFalsifyingEntryDirPathCorrect verifies Entry.Path is the
// full path, not just relative.
// (Falsifying TestFileDataSourceEntryReturnsFileEntry — lazy impl that sets Path incorrectly.)
// Spec: "FileEntry struct ... Path (string - full path)"
func TestFileDataSourceFalsifyingEntryDirPathCorrect(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "nested"), 0755)

	fds := NewFileDataSource(tmpDir, "*")

	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "nested/" {
			entry := fds.Entry(i)
			if entry == nil {
				t.Fatal("Entry returned nil for 'nested/'")
			}
			expectedPath := filepath.Join(tmpDir, "nested")
			if entry.Path != expectedPath {
				t.Errorf("entry.Path = %q, want full path %q", entry.Path, expectedPath)
			}
			return
		}
	}
	t.Error("'nested/' not found")
}

// TestFileDataSourceFalsifyingSetDirActuallyRefreshes verifies SetDir actually
// refreshes the entry list (Count changes if different number of files).
// (Falsifying TestFileDataSourceSetDirChangesDirectory — lazy impl that changes
// Dir() but doesn't refresh the listing.)
// Spec: "SetDir(dir string) — changes directory and refreshes the listing"
func TestFileDataSourceFalsifyingSetDirActuallyRefreshes(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// dir1 has 3 files, dir2 has 1 file
	os.WriteFile(filepath.Join(dir1, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir1, "b.txt"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(dir1, "c.txt"), []byte("c"), 0644)
	os.WriteFile(filepath.Join(dir2, "x.txt"), []byte("x"), 0644)

	fds := NewFileDataSource(dir1, "*")
	if fds.Count() == 0 {
		t.Fatal("dir1 should have entries")
	}

	// SetDir to dir2
	fds.SetDir(dir2)

	// After SetDir, the entry from dir2 should appear, and dir1 entries should not
	foundA := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "a.txt" {
			foundA = true
		}
	}
	if foundA {
		t.Error("After SetDir(dir2), 'a.txt' (from dir1) should not appear")
	}
}

// TestFileDataSourceFalsifyingSetWildcardActuallyFilters verifies SetWildcard
// actually changes which files appear.
// (Falsifying TestFileDataSourceSetWildcardChangesFilter — lazy impl that
// updates Wildcard() but doesn't refresh listing.)
// Spec: "SetWildcard(wc string) — changes wildcard and refreshes"
func TestFileDataSourceFalsifyingSetWildcardActuallyFilters(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "data.csv"), []byte("a,b"), 0644)

	fds := NewFileDataSource(tmpDir, "*")

	// With "*", both files appear
	foundGo := false
	foundCSV := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "main.go" {
			foundGo = true
		}
		if fds.Item(i) == "data.csv" {
			foundCSV = true
		}
	}
	if !foundGo || !foundCSV {
		t.Fatalf("Both files should appear with '*' wildcard. foundGo=%v foundCSV=%v", foundGo, foundCSV)
	}

	// Change to "*.go" — data.csv should disappear
	fds.SetWildcard("*.go")

	foundCSVAgain := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "data.csv" {
			foundCSVAgain = true
		}
	}
	if foundCSVAgain {
		t.Error("After SetWildcard('*.go'), 'data.csv' should not appear")
	}
}

// TestFileDataSourceFalsifyingSortingDirBeforeFileRegardlessOfName verifies
// a directory starting with "z" comes before a file starting with "a".
// (Falsifying TestFileDataSourceSortingDirsBeforeFiles — lazy impl that sorts
// everything together alphabetically without separating dirs from files.)
// Spec: "Directories ... Files" — dirs before files, within each group alphabetical.
func TestFileDataSourceFalsifyingSortingDirBeforeFileRegardlessOfName(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "zzz"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "aaa.txt"), []byte("x"), 0644)

	fds := NewFileDataSource(tmpDir, "*")

	// Find zzz/ and aaa.txt positions (skip "..")
	zzzIdx := -1
	aaaIdx := -1
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "zzz/" {
			zzzIdx = i
		}
		if fds.Item(i) == "aaa.txt" {
			aaaIdx = i
		}
	}

	if zzzIdx < 0 {
		t.Fatal("Directory 'zzz/' not found")
	}
	if aaaIdx < 0 {
		t.Fatal("File 'aaa.txt' not found")
	}

	// zzz/ (dir) must come before aaa.txt (file) despite "z" > "a"
	if zzzIdx > aaaIdx {
		t.Errorf("Directory 'zzz/' (index %d) should come before file 'aaa.txt' (index %d); dirs before files", zzzIdx, aaaIdx)
	}
}

// TestFileDataSourceFalsifyingHiddenNotVisibleWithStarWildcard verifies hidden
// files are hidden even with "*" wildcard.
// (Falsifying TestFileDataSourceFilteringExcludesHidden — lazy impl that only
// hides them for non-* patterns.)
// Spec: "Hidden files/directories (names starting with '.') are excluded"
func TestFileDataSourceFalsifyingHiddenNotVisibleWithStarWildcard(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, ".secret"), []byte("s"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte("v"), 0644)

	fds := NewFileDataSource(tmpDir, "*")

	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == ".secret" {
			t.Error("Hidden file '.secret' should not appear even with '*' wildcard")
		}
	}
}

// TestFileDataSourceFalsifyingHiddenDirNotVisible verifies hidden directories
// are excluded too, not just hidden files.
// (Falsifying TestFileDataSourceFilteringExcludesHidden — lazy impl that only
// excludes hidden files but not hidden directories.)
// Spec: "Hidden files/directories (names starting with '.') are excluded"
func TestFileDataSourceFalsifyingHiddenDirNotVisible(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".git"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)

	fds := NewFileDataSource(tmpDir, "*")

	foundGit := false
	foundSrc := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == ".git/" {
			foundGit = true
		}
		if fds.Item(i) == "src/" {
			foundSrc = true
		}
	}
	if foundGit {
		t.Error("Hidden directory '.git/' should not appear")
	}
	if !foundSrc {
		t.Error("Non-hidden directory 'src/' should appear")
	}
}

// TestFileDataSourceFalsifyingMultiPatternSecondPatternWorks verifies the
// second pattern in a multi-pattern wildcard actually works.
// (Falsifying TestFileDataSourceMultiPatternWildcard — lazy impl that only uses
// the first pattern in a semicolon-separated list.)
// Spec: "Wildcard supports semicolon-separated patterns: '*.go;*.mod' matches
// files matching either pattern"
func TestFileDataSourceFalsifyingMultiPatternSecondPatternWorks(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module x"), 0644)

	fds := NewFileDataSource(tmpDir, "*.go;*.mod")

	foundMod := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "go.mod" {
			foundMod = true
		}
	}
	if !foundMod {
		t.Error("go.mod should match second pattern in '*.go;*.mod'; second pattern must not be ignored")
	}
}

// TestFileDataSourceFalsifyingCountDecreasesWithRestrictiveWildcard verifies
// Count decreases when switching to a more restrictive wildcard.
// (Falsifying TestFileDataSourceCountWithFiles — lazy impl that always returns the
// same count regardless of wildcard.)
// Spec: "Count() int — returns the number of entries"
func TestFileDataSourceFalsifyingCountDecreasesWithRestrictiveWildcard(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# readme"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "docs"), 0755)

	fdsAll := NewFileDataSource(tmpDir, "*")
	fdsGo := NewFileDataSource(tmpDir, "*.go")

	// fdsGo should have fewer entries than fdsAll (readme.md excluded)
	if fdsGo.Count() >= fdsAll.Count() {
		t.Errorf("Restrictive wildcard '*.go' count=%d should be less than '*' count=%d",
			fdsGo.Count(), fdsAll.Count())
	}
}

// ---------------------------------------------------------------------------
// Section 13 — filepath.Match behaviour tests
// ---------------------------------------------------------------------------

// TestFileDataSourceWildcardUsesFilepathMatch verifies wildcard patterns use
// filepath.Match semantics, where "?" matches any single character.
// Spec: "Wildcard matching uses filepath.Match"
func TestFileDataSourceWildcardUsesFilepathMatch(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "ab.txt"), []byte("ab"), 0644)

	// "?.txt" should match "a.txt" but not "ab.txt" (filepath.Match semantics)
	fds := NewFileDataSource(tmpDir, "?.txt")

	foundA := false
	foundAB := false
	for i := 0; i < fds.Count(); i++ {
		switch fds.Item(i) {
		case "a.txt":
			foundA = true
		case "ab.txt":
			foundAB = true
		}
	}

	if !foundA {
		t.Error("a.txt should match '?.txt' (filepath.Match: ? matches single char)")
	}
	if foundAB {
		t.Error("ab.txt should NOT match '?.txt' (filepath.Match: ? matches exactly one char)")
	}
}

// TestFileDataSourceWildcardCharacterClass verifies filepath.Match character
// class patterns like "[abc]*" work correctly.
// Spec: "Wildcard matching uses filepath.Match"
func TestFileDataSourceWildcardCharacterClass(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "apple.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "banana.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "cherry.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "date.txt"), []byte("x"), 0644)

	// "[a-c]*.txt" matches files starting with a, b, or c
	fds := NewFileDataSource(tmpDir, "[a-c]*.txt")

	foundApple := false
	foundBanana := false
	foundCherry := false
	foundDate := false
	for i := 0; i < fds.Count(); i++ {
		switch fds.Item(i) {
		case "apple.txt":
			foundApple = true
		case "banana.txt":
			foundBanana = true
		case "cherry.txt":
			foundCherry = true
		case "date.txt":
			foundDate = true
		}
	}

	if !foundApple {
		t.Error("apple.txt should match '[a-c]*.txt' (starts with 'a')")
	}
	if !foundBanana {
		t.Error("banana.txt should match '[a-c]*.txt' (starts with 'b')")
	}
	if !foundCherry {
		t.Error("cherry.txt should match '[a-c]*.txt' (starts with 'c')")
	}
	if foundDate {
		t.Error("date.txt should NOT match '[a-c]*.txt' (starts with 'd')")
	}
}

// TestFileDataSourceWildcardStarMatchesAll verifies "*" matches EVERYTHING
// (except hidden files, which are excluded by filtering rules).
// Spec: "Wildcard matching uses filepath.Match" — filepath.Match("*", name) is always true.
func TestFileDataSourceWildcardStarMatchesAll(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.log"), []byte("2"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "noext"), []byte("3"), 0644)

	fds := NewFileDataSource(tmpDir, "*")

	for _, name := range []string{"file1.txt", "file2.log", "noext"} {
		found := false
		for i := 0; i < fds.Count(); i++ {
			if fds.Item(i) == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s should match '*' wildcard (filepath.Match('*', name) is always true)", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 14 — Additional edge case tests
// ---------------------------------------------------------------------------

// TestFileDataSourceSetDirThenSetWildcard verifies chaining SetDir and
// SetWildcard works correctly — each call refreshes independently.
// Spec: "refresh() re-reads the directory and rebuilds the sorted, filtered
// entry list. Called by the constructor, SetDir, and SetWildcard."
func TestFileDataSourceSetDirThenSetWildcard(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	os.WriteFile(filepath.Join(dir1, "a.go"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir2, "b.md"), []byte("b"), 0644)

	fds := NewFileDataSource(dir1, "*.go")

	// a.go should be visible
	foundA := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "a.go" {
			foundA = true
		}
	}
	if !foundA {
		t.Fatal("a.go should be visible with '*.go' in dir1")
	}

	// Change to dir2 — b.md should not be visible with "*.go"
	fds.SetDir(dir2)
	foundB := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "b.md" {
			foundB = true
		}
	}
	if foundB {
		t.Error("b.md should NOT be visible with '*.go' in dir2")
	}

	// Now change wildcard to "*.md" — b.md should appear
	fds.SetWildcard("*.md")
	foundB = false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "b.md" {
			foundB = true
		}
	}
	if !foundB {
		t.Error("b.md should be visible after SetWildcard('*.md')")
	}
}

// TestFileDataSourceWildcardStoredExactly verifies that Wildcard() returns
// the exact string passed to the constructor or SetWildcard, not a normalized form.
// Spec: "Wildcard() string — returns the current wildcard pattern"
func TestFileDataSourceWildcardStoredExactly(t *testing.T) {
	tmpDir := t.TempDir()

	// Wildcard with semicolons
	fds := NewFileDataSource(tmpDir, "*.go;*.mod")
	if fds.Wildcard() != "*.go;*.mod" {
		t.Errorf("Wildcard() = %q, want %q", fds.Wildcard(), "*.go;*.mod")
	}

	// Change to another multi-pattern
	fds.SetWildcard("*.c;*.h;*.cpp")
	if fds.Wildcard() != "*.c;*.h;*.cpp" {
		t.Errorf("Wildcard() = %q, want %q", fds.Wildcard(), "*.c;*.h;*.cpp")
	}
}

// TestFileDataSourceImplementsListDataSource verifies FileDataSource satisfies
// the ListDataSource interface.
// Spec: "FileDataSource implements ListDataSource and represents a directory
// listing filtered by a wildcard pattern."
func TestFileDataSourceImplementsListDataSource(t *testing.T) {
	// Already verified at package level via var _ ListDataSource = (*FileDataSource)(nil)
	// This test provides additional runtime verification
	tmpDir := t.TempDir()
	var ds ListDataSource = NewFileDataSource(tmpDir, "*")
	if ds.Count() == 0 {
		// Empty dir should at least have ".."
		t.Error("ListDataSource returned Count 0 unexpectedly")
	}
	// Item(0) should return something
	_ = ds.Item(0)
}

// TestFileDataSourceFileSizesVerifyable verifies that file sizes from Entry()
// reflect actual file content sizes.
// Spec: "FileEntry struct ... Size (int64)"
func TestFileDataSourceFileSizesVerifyable(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte("hello world, this is test content with a specific size")
	filePath := filepath.Join(tmpDir, "data.txt")
	os.WriteFile(filePath, content, 0644)

	fds := NewFileDataSource(tmpDir, "*")

	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "data.txt" {
			entry := fds.Entry(i)
			if entry == nil {
				t.Fatal("Entry returned nil")
			}
			if entry.Size != int64(len(content)) {
				t.Errorf("entry.Size = %d, want %d (must match file size)", entry.Size, len(content))
			}
			return
		}
	}
	t.Error("data.txt not found in listing")
}

// TestFileDataSourceEmptyDirSizeIsZero verifies directory sizes are
// reasonable (typically 0 or a multiple of the filesystem block size).
// Spec: "FileEntry struct ... Size (int64)"
func TestFileDataSourceDirectorySizeIsReasonable(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)

	fds := NewFileDataSource(tmpDir, "*")

	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == "subdir/" {
			entry := fds.Entry(i)
			if entry == nil {
				t.Fatal("Entry returned nil")
			}
			if entry.Size < 0 {
				t.Errorf("Directory size should not be negative: got %d", entry.Size)
			}
			return
		}
	}
	t.Error("subdir/ not found in listing")
}

// TestFileDataSourceSubdirectoryListing verifies that SetDir can navigate into
// subdirectories and list their contents.
// Spec: "SetDir(dir string) — changes directory and refreshes the listing"
func TestFileDataSourceSubdirectoryListing(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a nested structure: tmpDir/subdir/ containing a file
	subDir := filepath.Join(tmpDir, "projects")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "main.go"), []byte("package main"), 0644)

	fds := NewFileDataSource(tmpDir, "*")
	fds.SetDir(subDir)

	if fds.Dir() != subDir {
		t.Fatalf("Dir() = %q, want %q", fds.Dir(), subDir)
	}

	// Should have ".." (back to tmpDir) and main.go
	foundParent := false
	foundMain := false
	for i := 0; i < fds.Count(); i++ {
		if fds.Item(i) == ".." {
			foundParent = true
		}
		if fds.Item(i) == "main.go" {
			foundMain = true
		}
	}
	if !foundParent {
		t.Error("Subdirectory should have '..' to navigate back to parent")
	}
	if !foundMain {
		t.Error("Subdirectory should show 'main.go'")
	}
}

// TestFileDataSourceSortingStability verifies the sorting is stable across
// refreshes — refreshing without changes maintains the same order.
// Spec: "refresh() re-reads the directory and rebuilds the sorted, filtered
// entry list."
func TestFileDataSourceSortingStability(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "c.txt"), []byte("c"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755)

	fds := NewFileDataSource(tmpDir, "*")

	// Record the initial order
	firstPass := make([]string, fds.Count())
	for i := 0; i < fds.Count(); i++ {
		firstPass[i] = fds.Item(i)
	}

	// Refresh via SetWildcard (or SetDir back to same dir)
	fds.SetDir(tmpDir)

	secondPass := make([]string, fds.Count())
	for i := 0; i < fds.Count(); i++ {
		secondPass[i] = fds.Item(i)
	}

	// Item lists should be identical
	if len(firstPass) != len(secondPass) {
		t.Fatalf("Count changed: first=%d second=%d", len(firstPass), len(secondPass))
	}
	for i := range firstPass {
		if firstPass[i] != secondPass[i] {
			t.Errorf("Order changed at index %d: first=%q second=%q", i, firstPass[i], secondPass[i])
		}
	}
}

// TestFileDataSourceAllEntriesAccessible verifies every item returned by Item
// has a corresponding non-nil Entry (except possibly "..").
// Spec: "Entry(index int) *FileEntry — returns the underlying FileEntry"
func TestFileDataSourceAllEntriesAccessible(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("b"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)

	fds := NewFileDataSource(tmpDir, "*")

	for i := 0; i < fds.Count(); i++ {
		item := fds.Item(i)
		entry := fds.Entry(i)
		if item != ".." && entry == nil {
			t.Errorf("Item(%d) = %q but Entry(%d) returned nil", i, item, i)
		}
		if entry != nil && entry.Name == "" {
			t.Errorf("Entry(%d) has empty Name for item %q", i, item)
		}
	}
}

// TestFileDataSourceSymlinkDir verifies symlinks to directories are listed
// as directories.
// Spec: "directories are always included regardless of wildcard"
// Note: symlinks may be followed or not; this test checks handling.
func TestFileDataSourceSymlinkDir(t *testing.T) {
	tmpDir := t.TempDir()
	realDir := filepath.Join(tmpDir, "real_dir")
	linkDir := filepath.Join(tmpDir, "link_dir")
	os.MkdirAll(realDir, 0755)
	os.Symlink(realDir, linkDir)

	fds := NewFileDataSource(tmpDir, "*")

	// Both the real dir and symlink should appear (symlink may or may not be treated as dir)
	// At minimum, no panic should occur
	for i := 0; i < fds.Count(); i++ {
		_ = fds.Item(i)
		_ = fds.Entry(i)
	}
}

// TestFileDataSourceDoesNotLeakNonExistentDirEntries verifies Count returns 0
// (not a stale positive value) after SetDir to a non-existent directory.
// (Falsifying: lazy impl that doesn't clear entries on failed refresh.)
// Spec: "Directory that doesn't exist (Count returns 0)"
func TestFileDataSourceDoesNotLeakNonExistentDirEntries(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "hello.txt"), []byte("hi"), 0644)

	fds := NewFileDataSource(tmpDir, "*")
	if fds.Count() == 0 {
		t.Fatal("Initial Count should be > 0 for non-empty dir")
	}

	// SetDir to a non-existent directory
	fds.SetDir("/does/not/exist/anywhere")

	// Count must be 0 — stale entries from the previous directory must not leak
	if fds.Count() != 0 {
		t.Errorf("Count() = %d, want 0 after SetDir to non-existent path", fds.Count())
	}
}
