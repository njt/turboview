package tv

import (
	"testing"
	"time"

	"github.com/njt/turboview/theme"
)

// =============================================================================
// Construction Tests
// =============================================================================

// TestFileInfoPane_Constructor_IsNotSelectable verifies requirement 1:
// "FileInfoPane is not selectable — HasOption(OfSelectable) returns false"
func TestFileInfoPane_Constructor_IsNotSelectable(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))

	if pane.HasOption(OfSelectable) {
		t.Error("NewFileInfoPane must not set OfSelectable — pane should be non-selectable")
	}
}

// TestFileInfoPane_Constructor_HasOfPostProcess verifies requirement 2:
// "FileInfoPane has OfPostProcess — HasOption(OfPostProcess) returns true, so it receives broadcasts"
func TestFileInfoPane_Constructor_HasOfPostProcess(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))

	if !pane.HasOption(OfPostProcess) {
		t.Error("NewFileInfoPane must set OfPostProcess so pane receives broadcast events")
	}
}

// TestFileInfoPane_Constructor_SetsSfVisible verifies spec:
// "Sets SfVisible to true"
func TestFileInfoPane_Constructor_SetsSfVisible(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))

	if !pane.HasState(SfVisible) {
		t.Error("NewFileInfoPane must set SfVisible to true")
	}
}

// TestFileInfoPane_Constructor_SetsBounds verifies spec:
// "Sets bounds"
func TestFileInfoPane_Constructor_SetsBounds(t *testing.T) {
	r := NewRect(5, 10, 80, 1)
	pane := NewFileInfoPane(r)

	if pane.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", pane.Bounds(), r)
	}
}

// =============================================================================
// SetEntry Tests — Requirement 10
// =============================================================================

// TestFileInfoPane_SetEntry_StoresEntry verifies requirement 10:
// "SetEntry — stores the entry for later retrieval"
func TestFileInfoPane_SetEntry_StoresEntry(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	entry := &FileEntry{
		Name:    "test.txt",
		Size:    1024,
		ModTime: time.Date(2026, time.January, 15, 14, 30, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/home/test.txt",
	}

	pane.SetEntry(entry)

	if pane.entry != entry {
		t.Fatal("SetEntry did not store the entry")
	}
	if pane.entry.Name != "test.txt" {
		t.Errorf("entry.Name = %q, want %q", pane.entry.Name, "test.txt")
	}
	if pane.entry.Size != 1024 {
		t.Errorf("entry.Size = %d, want %d", pane.entry.Size, 1024)
	}
}

// TestFileInfoPane_SetEntry_StoresNil verifies SetEntry can clear the entry:
// setting nil should be allowed.
func TestFileInfoPane_SetEntry_StoresNil(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	entry := &FileEntry{Name: "temp.txt", Path: "/tmp/temp.txt"}
	pane.SetEntry(entry)
	if pane.entry == nil {
		t.Fatal("SetEntry should have stored the entry")
	}

	pane.SetEntry(nil)
	if pane.entry != nil {
		t.Error("SetEntry(nil) should clear the entry; entry is not nil")
	}
}

// =============================================================================
// HandleEvent Tests — Requirements 3, 4, 5, 12
// =============================================================================

// TestFileInfoPane_HandleEvent_ReceivesCmFileFocused verifies requirement 3:
// "Receives CmFileFocused broadcast — stores the *FileEntry from event.Info"
func TestFileInfoPane_HandleEvent_ReceivesCmFileFocused(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	entry := &FileEntry{
		Name:    "readme.md",
		Size:    2048,
		ModTime: time.Date(2026, time.March, 3, 9, 15, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/docs/readme.md",
	}

	event := &Event{
		What:    EvBroadcast,
		Command: CmFileFocused,
		Info:    entry,
	}

	pane.HandleEvent(event)

	if pane.entry != entry {
		t.Fatal("HandleEvent should store the *FileEntry from event.Info")
	}
	if pane.entry.Name != "readme.md" {
		t.Errorf("entry.Name = %q, want %q", pane.entry.Name, "readme.md")
	}
}

// TestFileInfoPane_HandleEvent_IgnoresNonBroadcast verifies requirement 4:
// "Ignores non-broadcast events — does nothing for keyboard/mouse/command events"
func TestFileInfoPane_HandleEvent_IgnoresNonBroadcast(t *testing.T) {
	entry := &FileEntry{Name: "original.txt", Path: "/tmp/original.txt"}

	tests := []struct {
		name  string
		event *Event
	}{
		{
			name: "EvKeyboard",
			event: &Event{
				What:    EvKeyboard,
				Command: CmFileFocused,
				Info:    &FileEntry{Name: "keyboard.txt", Path: "/fake/keyboard.txt"},
			},
		},
		{
			name: "EvMouse",
			event: &Event{
				What:    EvMouse,
				Command: CmFileFocused,
				Info:    &FileEntry{Name: "mouse.txt", Path: "/fake/mouse.txt"},
			},
		},
		{
			name: "EvCommand",
			event: &Event{
				What:    EvCommand,
				Command: CmFileFocused,
				Info:    &FileEntry{Name: "command.txt", Path: "/fake/command.txt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
			pane.SetEntry(entry)

			pane.HandleEvent(tt.event)

			// Entry must remain unchanged — non-broadcast events are ignored.
			if pane.entry != entry {
				t.Errorf("HandleEvent modified entry on %s event; entry should remain unchanged", tt.name)
			}
		})
	}
}

// TestFileInfoPane_HandleEvent_IgnoresWrongBroadcast verifies requirement 5:
// "Ignores broadcasts with wrong command — only responds to CmFileFocused"
func TestFileInfoPane_HandleEvent_IgnoresWrongBroadcast(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	originalEntry := &FileEntry{Name: "original.txt", Path: "/tmp/original.txt"}
	pane.SetEntry(originalEntry)

	// Send a broadcast with a different command (CmOK).
	event := &Event{
		What:    EvBroadcast,
		Command: CmOK,
		Info:    &FileEntry{Name: "should-be-ignored.txt", Path: "/tmp/should-be-ignored.txt"},
	}

	pane.HandleEvent(event)

	if pane.entry != originalEntry {
		t.Error("HandleEvent should ignore broadcasts with command != CmFileFocused; entry was modified")
	}
}

// TestFileInfoPane_HandleEvent_DoesNotClearEvent verifies requirement 12 (falsifying):
// "HandleEvent does not clear the event — broadcast events should remain uncleared"
func TestFileInfoPane_HandleEvent_DoesNotClearEvent(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	entry := &FileEntry{Name: "test.txt", Path: "/tmp/test.txt"}

	event := &Event{
		What:    EvBroadcast,
		Command: CmFileFocused,
		Info:    entry,
	}

	pane.HandleEvent(event)

	if event.IsCleared() {
		t.Error("HandleEvent must NOT clear broadcast events — broadcasts are meant to reach all listeners")
	}
}

// =============================================================================
// Falsifying: HandleEvent type assertion safety — Requirement 3
// =============================================================================

// TestFileInfoPane_HandleEvent_IgnoresNilInfo verifies HandleEvent does not panic
// and does not store nil when event.Info is nil.
func TestFileInfoPane_HandleEvent_IgnoresNilInfo(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	originalEntry := &FileEntry{Name: "original.txt", Path: "/tmp/original.txt"}
	pane.SetEntry(originalEntry)

	event := &Event{
		What:    EvBroadcast,
		Command: CmFileFocused,
		Info:    nil,
	}

	// Must not panic.
	pane.HandleEvent(event)

	// Entry should remain unchanged since nil cannot be type-asserted to *FileEntry.
	if pane.entry != originalEntry {
		t.Error("HandleEvent should not modify entry when event.Info is nil")
	}
}

// TestFileInfoPane_HandleEvent_IgnoresWrongInfoType verifies HandleEvent does not panic
// and ignores event.Info that is not a *FileEntry.
func TestFileInfoPane_HandleEvent_IgnoresWrongInfoType(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	originalEntry := &FileEntry{Name: "original.txt", Path: "/tmp/original.txt"}
	pane.SetEntry(originalEntry)

	// Info is a string instead of *FileEntry.
	event := &Event{
		What:    EvBroadcast,
		Command: CmFileFocused,
		Info:    "not a FileEntry",
	}

	// Must not panic.
	pane.HandleEvent(event)

	// Entry should remain unchanged since Info is the wrong type.
	if pane.entry != originalEntry {
		t.Error("HandleEvent should not modify entry when event.Info is wrong type")
	}
}

// =============================================================================
// Draw Tests — Requirements 6, 7, 8, 9
// =============================================================================

// readRowAsString reads all runes from row y of buffer into a string.
func readRowAsString(buf *DrawBuffer, y int) string {
	w := buf.Width()
	runes := make([]rune, w)
	for x := 0; x < w; x++ {
		runes[x] = buf.GetCell(x, y).Rune
	}
	return string(runes)
}

// TestFileInfoPane_Draw_NilEntryReturnsEarly verifies requirement 6:
// "Draw with nil entry — draws nothing (no panic)"
func TestFileInfoPane_Draw_NilEntryReturnsEarly(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	pane.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(80, 1)
	pane.Draw(buf) // Must not panic.

	// All cells should remain at default style (nil entry → Draw returns early).
	for x := 0; x < 80; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune != ' ' {
			t.Errorf("cell(%d, 0) rune = %c, want space (nil entry should draw nothing)", x, cell.Rune)
		}
	}
}

// TestFileInfoPane_Draw_NarrowBoundsReturnsEarly verifies requirement 9:
// "Draw with narrow bounds — width < 4 returns early (no panic)"
func TestFileInfoPane_Draw_NarrowBoundsReturnsEarly(t *testing.T) {
	entry := &FileEntry{
		Name:    "file.txt",
		Size:    100,
		ModTime: time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/tmp/file.txt",
	}

	for _, width := range []int{0, 1, 2, 3} {
		pane := NewFileInfoPane(NewRect(0, 0, width, 1))
		pane.scheme = theme.BorlandBlue
		pane.SetEntry(entry)

		buf := NewDrawBuffer(max(width, 1), 1)
		pane.Draw(buf) // Must not panic.

		// Buffer should not be modified (Draw returns early).
		for x := 0; x < max(width, 1); x++ {
			cell := buf.GetCell(x, 0)
			if cell.Rune != ' ' && width > 0 {
				t.Errorf("width=%d: cell(%d, 0) rune = %c, want space (narrow bounds should draw nothing)", width, x, cell.Rune)
			}
		}
	}
}

// TestFileInfoPane_Draw_FillsBackgroundWithListNormal verifies the spec requirement:
// "Fills the background with spaces using ListNormal style"
func TestFileInfoPane_Draw_FillsBackgroundWithListNormal(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	pane.scheme = theme.BorlandBlue
	entry := &FileEntry{
		Name:    "test.txt",
		Size:    5000,
		ModTime: time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/tmp/test.txt",
	}
	pane.SetEntry(entry)

	buf := NewDrawBuffer(80, 1)
	pane.Draw(buf)

	listStyle := pane.scheme.ListNormal
	for x := 0; x < 80; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Style != listStyle {
			t.Errorf("cell(%d, 0) style = %v, want ListNormal %v — background should be filled with ListNormal", x, cell.Style, listStyle)
			return // One failure is enough to diagnose.
		}
	}
}

// TestFileInfoPane_Draw_DirectoryShowsDirAndTrailingSlash verifies requirement 7:
// "Draw directory entry — shows '<DIR>' and name with '/' suffix"
func TestFileInfoPane_Draw_DirectoryShowsDirAndTrailingSlash(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	pane.scheme = theme.BorlandBlue
	entry := &FileEntry{
		Name:    "Documents",
		Size:    0,
		ModTime: time.Date(2026, time.February, 15, 10, 30, 0, 0, time.UTC),
		IsDir:   true,
		Path:    "/home/Documents",
	}
	pane.SetEntry(entry)

	buf := NewDrawBuffer(80, 1)
	pane.Draw(buf)

	row := readRowAsString(buf, 0)

	if !containsString(row, "<DIR>") {
		t.Errorf("directory entry should show '<DIR>', but row content was: %q", row)
	}
	if !containsString(row, "Documents/") {
		t.Errorf("directory entry name should have '/' suffix, but row content was: %q", row)
	}
}

// TestFileInfoPane_Draw_FileShowsCommaFormattedSize verifies requirement 8:
// "Draw file entry — shows name, comma-formatted size, date, time"
func TestFileInfoPane_Draw_FileShowsCommaFormattedSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		wantSize string
	}{
		{"small file", 999, "999"},
		{"file with thousands", 1234, "1,234"},
		{"file with millions", 1234567, "1,234,567"},
		{"zero size", 0, "0"},
		{"single comma", 1000, "1,000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
			pane.scheme = theme.BorlandBlue
			entry := &FileEntry{
				Name:    "file.txt",
				Size:    tt.size,
				ModTime: time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC),
				IsDir:   false,
				Path:    "/tmp/file.txt",
			}
			pane.SetEntry(entry)

			buf := NewDrawBuffer(80, 1)
			pane.Draw(buf)

			row := readRowAsString(buf, 0)

			if !containsString(row, tt.wantSize) {
				t.Errorf("size %d should be formatted as %q in draw output; got row: %q", tt.size, tt.wantSize, row)
			}
		})
	}
}

// TestFileInfoPane_Draw_FileShowsName verifies the filename appears in the left side.
// Requirement 8: "Draw file entry — shows name"
func TestFileInfoPane_Draw_FileShowsName(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	pane.scheme = theme.BorlandBlue
	entry := &FileEntry{
		Name:    "README.md",
		Size:    2048,
		ModTime: time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/home/README.md",
	}
	pane.SetEntry(entry)

	buf := NewDrawBuffer(80, 1)
	pane.Draw(buf)

	row := readRowAsString(buf, 0)

	// Name should appear at the left side (after two leading spaces).
	// The two leading spaces mean the first two chars are spaces, then the name starts.
	if !containsString(row, "README.md") {
		t.Errorf("filename should appear in draw output; got row: %q", row)
	}
}

// TestFileInfoPane_Draw_LongFilenameTruncated verifies requirement 8's truncation:
// "If name exceeds half the width, truncate with '~' at end."
func TestFileInfoPane_Draw_LongFilenameTruncated(t *testing.T) {
	// Use narrow width so a long name exceeds half.
	const width = 40
	longName := "this-is-a-very-long-filename-that-exceeds-half-width.txt"

	pane := NewFileInfoPane(NewRect(0, 0, width, 1))
	pane.scheme = theme.BorlandBlue
	entry := &FileEntry{
		Name:    longName,
		Size:    100,
		ModTime: time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/tmp/" + longName,
	}
	pane.SetEntry(entry)

	buf := NewDrawBuffer(width, 1)
	pane.Draw(buf)

	row := readRowAsString(buf, 0)

	// The full name should NOT appear (it exceeds half the width and should be truncated).
	if containsString(row, longName) {
		t.Errorf("long filename should be truncated, but full name appeared in row: %q", row)
	}

	// The row should contain a '~' indicating truncation.
	if !containsString(row, "~") {
		t.Errorf("truncated filename should contain '~' marker; got row: %q", row)
	}

	// At least a prefix of the name should appear.
	if !containsString(row, "this") {
		t.Errorf("truncated filename should contain beginning of the name; got row: %q", row)
	}
}

// TestFileInfoPane_Draw_ShortFilenameNotTruncated verifies non-truncation:
// a short filename (well under half width) should NOT have a '~' in the name area.
func TestFileInfoPane_Draw_ShortFilenameNotTruncated(t *testing.T) {
	const width = 80
	pane := NewFileInfoPane(NewRect(0, 0, width, 1))
	pane.scheme = theme.BorlandBlue
	entry := &FileEntry{
		Name:    "short.txt",
		Size:    100,
		ModTime: time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/tmp/short.txt",
	}
	pane.SetEntry(entry)

	buf := NewDrawBuffer(width, 1)
	pane.Draw(buf)

	row := readRowAsString(buf, 0)

	// The full name should appear.
	if !containsString(row, "short.txt") {
		t.Errorf("short filename should appear fully in draw output; got row: %q", row)
	}
}

// TestFileInfoPane_Draw_ShowsDateAndTime verifies requirement 8:
// "date formatted as 'Jan  2, 2006', time formatted as '3:04pm'"
func TestFileInfoPane_Draw_ShowsDateAndTime(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	pane.scheme = theme.BorlandBlue
	entry := &FileEntry{
		Name:    "file.txt",
		Size:    100,
		ModTime: time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/tmp/file.txt",
	}
	pane.SetEntry(entry)

	buf := NewDrawBuffer(80, 1)
	pane.Draw(buf)

	row := readRowAsString(buf, 0)

	// Go's time.Format("Jan  2, 2006") produces "Jan  2, 2026" with double-space padding.
	if !containsString(row, "Jan  2, 2026") {
		t.Errorf("date should be formatted as 'Jan  2, 2026' (double-space padded day); got row: %q", row)
	}
	// Go's time.Format("3:04pm") produces "3:04pm".
	if !containsString(row, "3:04pm") {
		t.Errorf("time should be formatted as '3:04pm'; got row: %q", row)
	}
}

// TestFileInfoPane_Draw_HasLeadingSpaces verifies requirement 8:
// "Left side: two spaces, then the filename"
func TestFileInfoPane_Draw_HasLeadingSpaces(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	pane.scheme = theme.BorlandBlue
	entry := &FileEntry{
		Name:    "test.go",
		Size:    500,
		ModTime: time.Date(2026, time.February, 20, 9, 45, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/tmp/test.go",
	}
	pane.SetEntry(entry)

	buf := NewDrawBuffer(80, 1)
	pane.Draw(buf)

	// First two characters must be spaces.
	if buf.GetCell(0, 0).Rune != ' ' {
		t.Errorf("cell(0, 0) rune = %c, want space (leading padding)", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(1, 0).Rune != ' ' {
		t.Errorf("cell(1, 0) rune = %c, want space (leading padding)", buf.GetCell(1, 0).Rune)
	}
}

// TestFileInfoPane_Draw_DirectoryShowsDirNoSize verifies requirement 7:
// Directory entries show "<DIR>" instead of a numeric size.
func TestFileInfoPane_Draw_DirectoryShowsDirNoSize(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	pane.scheme = theme.BorlandBlue
	entry := &FileEntry{
		Name:    "Projects",
		Size:    4096, // Even with a size, directories show <DIR>
		ModTime: time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC),
		IsDir:   true,
		Path:    "/home/Projects",
	}
	pane.SetEntry(entry)

	buf := NewDrawBuffer(80, 1)
	pane.Draw(buf)

	row := readRowAsString(buf, 0)

	if !containsString(row, "<DIR>") {
		t.Errorf("directory entry should show '<DIR>'; got row: %q", row)
	}

	// Size value should NOT appear for directories.
	if containsString(row, "4,096") {
		t.Errorf("directory entry should show '<DIR>' not numeric size; got row: %q", row)
	}
}

// TestFileInfoPane_Draw_UsesSubStr verifies the Draw method only affects the
// pane's own bounds row, not other rows or outside the horizontal bounds.
func TestFileInfoPane_Draw_UsesSubBuffer(t *testing.T) {
	// Full buffer is 100 wide, 3 tall. Pane occupies x=10, y=1, w=80, h=1.
	buf := NewDrawBuffer(100, 3)

	// Pre-fill all cells with a known non-default style to detect overdraw.
	knownStyle := theme.BorlandBlue.LabelNormal
	for y := 0; y < 3; y++ {
		for x := 0; x < 100; x++ {
			buf.WriteChar(x, y, 'X', knownStyle)
		}
	}

	pane := NewFileInfoPane(NewRect(10, 1, 80, 1))
	pane.scheme = theme.BorlandBlue
	entry := &FileEntry{
		Name:    "test.txt",
		Size:    100,
		ModTime: time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/tmp/test.txt",
	}
	pane.SetEntry(entry)

	pane.Draw(buf)

	// Pane only affects columns 10-89 in row 1.
	// Rows 0 and 2 should remain unchanged.
	for y := 0; y < 3; y++ {
		for x := 0; x < 100; x++ {
			cell := buf.GetCell(x, y)
			if y == 1 && x >= 10 && x < 90 {
				// Within pane bounds — should have ListNormal style.
				if cell.Style != pane.scheme.ListNormal {
					t.Errorf("cell(%d, %d) style = %v, want ListNormal %v", x, y, cell.Style, pane.scheme.ListNormal)
				}
			} else {
				// Outside pane bounds — should retain original style.
				if cell.Style != knownStyle {
					t.Errorf("cell(%d, %d) style = %v, want original %v (pane over-drew outside its bounds)", x, y, cell.Style, knownStyle)
				}
			}
		}
	}
}

// =============================================================================
// Falsifying Tests
// =============================================================================

// TestFileInfoPane_Draw_NilColorSchemeDoesNotPanic verifies requirement 11 (falsifying):
// "Draw does not crash on nil color scheme — if ColorScheme() returns nil, should not panic"
func TestFileInfoPane_Draw_NilColorSchemeDoesNotPanic(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	// Do NOT set scheme; also do not set owner, so ColorScheme() returns nil.
	entry := &FileEntry{
		Name:    "test.txt",
		Size:    100,
		ModTime: time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/tmp/test.txt",
	}
	pane.SetEntry(entry)

	// Verify ColorScheme() returns nil.
	if pane.ColorScheme() != nil {
		t.Skip("ColorScheme() returned non-nil; cannot test nil scheme behavior")
	}

	buf := NewDrawBuffer(80, 1)
	// Must not panic.
	pane.Draw(buf)
}

// TestFileInfoPane_Draw_NilColorSchemeNoEntryDoesNotPanic verifies the combination
// of nil entry + nil scheme still does not panic.
func TestFileInfoPane_Draw_NilColorSchemeNoEntryDoesNotPanic(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 80, 1))
	// No scheme, no owner, no entry.

	buf := NewDrawBuffer(80, 1)
	// Must not panic.
	pane.Draw(buf)
}

// TestFileInfoPane_Draw_DoesNotPanicWithZeroWidth verifies Draw handles width=0 buffer.
func TestFileInfoPane_Draw_DoesNotPanicWithZeroWidth(t *testing.T) {
	pane := NewFileInfoPane(NewRect(0, 0, 0, 1))
	pane.scheme = theme.BorlandBlue
	entry := &FileEntry{
		Name:    "test.txt",
		Size:    100,
		ModTime: time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC),
		IsDir:   false,
		Path:    "/tmp/test.txt",
	}
	pane.SetEntry(entry)

	buf := NewDrawBuffer(1, 1)
	// Must not panic.
	pane.Draw(buf)
}

// =============================================================================
// Helpers
// =============================================================================

// containsString reports whether s contains substr.
func containsString(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
