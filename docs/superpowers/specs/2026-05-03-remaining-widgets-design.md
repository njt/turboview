# Remaining Turbo Vision Widgets — Behavioral Spec

**Goal:** Implement all remaining original Turbo Vision widgets in TurboView: input validation, multi-column clusters, full text editor, file dialog, color dialog, and tree view.

**Reference authority:** The magiblot/tvision C++ source code (https://github.com/magiblot/tvision) is the definitive behavioral reference. Clone to `/tmp/tvision` for research.

**Relationship to existing spec:** This spec covers Batches 4-9. Batches 1-3 (TLabel validation, TListBox/TMemo fixes, THistory fixes) are covered by `2026-05-01-missing-widgets-design.md` and should be executed first.

**Batch sequence:**

| Batch | Capability | Depends On |
|-------|-----------|-----------|
| 4 | Input validation | InputLine (existing) |
| 5 | Multi-column clusters + list | CheckBoxes, RadioButtons, ListViewer (existing) |
| 6 | TEditor + TEditWindow | Memo (existing + Batch 1-3 fixes) |
| 7 | TFileDialog | ListBox, InputLine, Label, Dialog (existing) |
| 8 | TColorDialog | ListViewer, multi-column clusters (Batch 5) |
| 9 | TOutline | ScrollBar (existing) |

---

## 4. Input Validation (TValidator)

Reference: `validate.h`, `tvalidat.cpp`, `tinputli.cpp`

### 4.1 Architecture

TValidator is an interface that InputLine optionally holds. Validators implement two-phase validation: `IsValidInput()` on every keystroke (rejects invalid characters in real-time), and `IsValid()` on blur/commit (rejects incomplete input). When `IsValid()` fails, `Error()` shows a message box.

```
Validator (interface)
├── FilterValidator      — character whitelist
│   └── RangeValidator   — numeric min/max (extends FilterValidator)
├── PictureValidator     — pattern-based format (e.g., phone numbers)
└── StringLookupValidator — validates against a list of allowed strings
```

### 4.2 Validator Interface

```go
type Validator interface {
    IsValid(s string) bool
    IsValidInput(s string, noAutoFill bool) bool
    Error()
}
```

- `IsValidInput(s, noAutoFill)` — called on every keystroke, paste, delete. Returns true if the current partial input is acceptable so far. If false, the keystroke is rejected (InputLine reverts the change). For PictureValidator, when `noAutoFill` is false and auto-fill is enabled, it may insert literal characters (parentheses, dashes) automatically.
- `IsValid(s)` — called when InputLine loses focus or when the dialog OK button is pressed. Returns true if the input is complete and valid.
- `Error()` — displays an error message box describing why validation failed. Each validator type has its own message.

### 4.3 InputLine Integration

InputLine gets a new optional field:

```go
func (il *InputLine) SetValidator(v Validator)
func (il *InputLine) Validator() Validator
```

Validation hooks:
- **On keystroke/paste/delete:** Before modifying the text buffer, InputLine saves its entire state (text, cursor position, selection). After mutation, call `IsValidInput(newText, false)`. If it returns false, **roll back the entire InputLine state** to the saved snapshot — not just the text, but cursor and selection too. This rollback mechanism is critical: without it, the cursor drifts on rejected keystrokes. For PictureValidator with auto-fill, `IsValidInput` may modify the text string (inserting literal characters) — InputLine must accept the modified string.
- **On blur (`SfSelected` cleared):** Call `IsValid(text)`. If false, call `Error()`, then re-focus the InputLine (prevent leaving an invalid field). This uses the `validate()` method pattern from the original.
- **On dialog OK:** Dialog's `Valid()` method iterates all children. Each InputLine with a validator calls `IsValid()`. If any fails, that InputLine gets focus, `Error()` shows, and the dialog stays open.
- **On `CmCancel`:** Skip validation entirely — the user is cancelling, not committing.

### 4.4 FilterValidator

Accepts only characters from a whitelist string.

```go
func NewFilterValidator(validChars string) *FilterValidator
```

- `IsValidInput`: every rune in `s` must be in `validChars`. Uses rune-level matching.
- `IsValid`: same check (if all chars are valid individually, the string is valid).
- `Error`: message box "Invalid character in input"

### 4.5 RangeValidator

Extends FilterValidator with numeric range checking.

```go
func NewRangeValidator(min, max int) *RangeValidator
```

- Inherits FilterValidator. Auto-selects character set based on range: if `min >= 0`, uses `validChars = "+0123456789"` (unsigned); otherwise uses `validChars = "+-0123456789"` (signed). This matches original `tvalidat.cpp:661-665`.
- `IsValidInput`: delegates to FilterValidator (allows digits and signs while typing). Also accepts empty string and lone `+`/`-` (partial input during typing).
- `IsValid`: parses the string as an integer, checks `min <= value <= max`. Empty string fails.
- `Error`: message box "Value not in the range %d to %d"

### 4.6 PictureValidator

Pattern-based validation with optional auto-fill.

```go
func NewPictureValidator(pic string, autoFill bool) *PictureValidator
```

Picture format codes (matching original `tvalidat.cpp`):
- `#` — digit required
- `?` — letter required
- `&` — letter required, auto-uppercase
- `!` — any char, auto-uppercase
- `@` — any char as-is
- `*N{group}` — repeat group N times (or unlimited if no number follows `*`)
- `[...]` — optional group
- `{...}` — required group
- `;` — escape next character as literal
- `,` — alternation (try different sub-patterns)

`IsValidInput` walks the picture and input simultaneously, matching each input character against the current picture position. Returns true if input matches the picture so far (may be incomplete). With auto-fill enabled and `noAutoFill` false, literal characters in the picture that the user hasn't typed yet are inserted automatically into the string.

`IsValid` requires the picture to be completely satisfied — all required positions filled. Returns false if required groups remain unfilled.

`Error`: message box "Input does not conform to picture: %s" showing the picture string.

### 4.7 StringLookupValidator

Validates against a fixed list of allowed strings.

```go
func NewStringLookupValidator(items []string) *StringLookupValidator
```

- `IsValidInput`: always returns true (any partial input is acceptable — validation is only on commit).
- `IsValid`: exact case-sensitive match against any item in the list.
- `Error`: message box "Input not in valid list"

### 4.8 Color Scheme

No new color scheme entries needed. Validators use message box dialogs which use existing dialog colors.

### 4.9 Demo Integration

Add to the "File Manager" dialog:
- A "Port:" InputLine with `NewRangeValidator(1, 65535)`
- The existing filename InputLine gets `NewFilterValidator` rejecting null bytes

---

## 5. Multi-column Clusters + Multi-column ListViewer

Reference: `tcluster.cpp` (draw, handleEvent, column, row, findSel), `tlistviewer.cpp`

### 5.1 TCluster Multi-column Layout

Column count is determined dynamically from the widget's bounds height. Items fill top-to-bottom within a column, then overflow to the next column. This matches the original algorithm in `tcluster.cpp`.

**Algorithm:**

- `row(item) = item % height` — which row an item occupies
- `col(item) = item / height` — which column
- Column width = widest label in that column + prefix (5 chars: 1 focus indicator + `[ ] ` or `( ) `)
- Column x-position = sum of widths of all columns to the left

**Example:** 6 items, bounds height = 3:

```
Column 0          Column 1
►[ ] Read only    ( ) Hex
 [ ] Hidden       ( ) Binary  
 [ ] System       ( ) Text
```

### 5.2 Item Enable/Disable

Individual items can be disabled via a bitmask, matching original `tcluster.cpp`'s `enableMask`:

```go
func (c *CheckBoxes) SetEnabled(index int, enabled bool)
func (c *CheckBoxes) IsEnabled(index int) bool
// Same for RadioButtons
```

- Disabled items render in a dimmed color (use existing `ButtonDisabled` or add `ClusterDisabled`)
- Keyboard navigation skips disabled items (arrow keys jump over them)
- Mouse clicks on disabled items are ignored
- Space bar on a disabled item does nothing
- If all items are disabled, the cluster clears `OfSelectable`

### 5.3 Changes to CheckBoxes and RadioButtons

Both currently place items at fixed `NewRect(0, i, width, 1)` — single column only. Changes:

1. Add a `relayoutItems()` method called on construction and `SetBounds` that computes item positions using the column algorithm
2. Each item's bounds become `NewRect(colX, row, colWidth, 1)`
3. Add Left/Right keyboard navigation in `HandleEvent`: Right adds `height` to the focused index (next column), Left subtracts `height`. Clamp to valid range, don't wrap. Skip disabled items.
4. **Space bar** activates (toggles/selects) the currently focused item
5. **Hot key matching:** Alt+letter matches from anywhere. Plain letter matches only when the cluster is focused or during postProcess phase. On match, the cluster takes focus before activating the item.
6. Mouse routing already works via `item.Bounds().Contains()` — correct bounds are sufficient
7. Items in the rightmost column get `GfGrowHiX`; other items get fixed width

### 5.4 Multi-column ListViewer

ListViewer gains a `numCols` parameter. Each column shows a contiguous range of items; the scrollbar navigates all columns together.

```go
func (lv *ListViewer) SetNumCols(n int)
func (lv *ListViewer) NumCols() int
```

**Layout:**
- Column width = `bounds.Width() / numCols`
- Items per visible page = `numCols * bounds.Height()`
- Column `c` shows items `topIndex + c*height` through `topIndex + (c+1)*height - 1`
- Single-character vertical divider `│` drawn between columns

**Keyboard:** Up/Down move within column. Left/Right move between columns (shift focus by `height` items). If moving right would exceed item count, don't move.

**Mouse:** Click x-position determines column (`col = x / colWidth`), y-position determines row. Item index = `topIndex + col*height + row`.

**Scrollbar range:** Adjusted so each "page" scrolls by `numCols * height` items. `maxValue = ceil(itemCount / (numCols * height))`.

**Draw:** Iterate columns left-to-right. Each column draws its items as a vertical slice. Focused item highlight spans only its column width. Divider lines drawn in normal color.

### 5.5 Color Scheme

New entry:

| Entry | Purpose |
|-------|---------|
| `ClusterDisabled` | Disabled cluster item text |

All five themes updated.

### 5.6 Demo Integration

- Change CheckBoxes in Window 1 to bounds height 2 with 3 items → forces 2 columns
- Change RadioButtons similarly
- Optionally show a 2-column ListViewer

---

## 6. TEditor + TEditWindow

Reference: `editors.h`, `teditor1.cpp`, `teditor2.cpp`, `teditor3.cpp`

### 6.1 Architecture

Editor extends the existing Memo with undo, find/replace, indicator linkage, and file I/O. It keeps the line-based `[][]rune` buffer — no gap buffer rewrite.

```
Memo (existing — cursor, selection, scrollbars, keyboard/mouse editing)
 └── Editor (embeds *Memo — adds undo, find/replace, file I/O, indicator)

EditWindow (Window subclass — wraps Editor + Indicator + ScrollBars)
```

### 6.2 Undo

Single-level undo matching original TV's behavior (not a full undo stack):

```go
type undoState struct {
    lines    [][]rune
    cursorRow, cursorCol int
    selStartRow, selStartCol, selEndRow, selEndCol int
    deltaX, deltaY int
}
```

- Before each edit operation (character insert, delete, paste, cut, line delete, Enter), snapshot the current state into a single `*undoState` field
- `Ctrl+Z`: restores the snapshot, then clears it (one undo only, matching original)
- Next edit after undo overwrites the snapshot
- `CanUndo() bool` — true if a snapshot exists

### 6.3 Find/Replace

Two modal dialogs:

| Key | Action |
|-----|--------|
| Ctrl+F | Open Find dialog |
| Ctrl+H | Open Replace dialog |
| Ctrl+L or F3 | Search again (repeat last find) |

**Find dialog** contains:
- InputLine "Search for:" with history (historyID for search terms)
- CheckBoxes: `[ ] Case sensitive`, `[ ] Whole words only`
- Buttons: OK, Cancel

**Replace dialog** extends Find dialog with:
- InputLine "Replace with:" with history
- Additional checkbox: `[ ] Prompt on replace`
- Buttons: OK, Replace All, Cancel

**Search behavior:**
- Linear scan forward from cursor position
- If not found forward, wraps to start of document and searches to cursor position
- Case-insensitive by default
- Whole-word matching checks character class boundaries (same word boundary logic as Memo's Ctrl+Left/Right)
- No regex support (matching original)
- On match: select the matched text, scroll to make it visible
- Replace: replace selected text, advance to next match
- Replace All: replace all occurrences, report count
- "Prompt on replace": each match highlights and shows a Yes/No/Cancel confirm dialog. Yes replaces and continues, No skips and continues, Cancel aborts.

**Search state** is stored on the Editor:

```go
type searchState struct {
    findText    string
    replaceText string
    caseSensitive bool
    wholeWords    bool
    promptOnReplace bool
}
```

### 6.4 TIndicator

A small non-focusable view showing editor state, placed in the EditWindow's frame footer:

```
 2:15  *
```

- Line:column display (1-indexed)
- `*` character when modified (text differs from last save/load)
- Not selectable, sets `OfPostProcess`

```go
type Indicator struct {
    BaseView
    line, col int
    modified  bool
}

func NewIndicator(bounds Rect) *Indicator
func (ind *Indicator) SetValue(line, col int, modified bool)
```

Editor broadcasts `CmIndicatorUpdate` (new command) after every cursor move or edit. Indicator listens in postprocess and calls `SetValue`.

### 6.5 File I/O

```go
func (e *Editor) LoadFile(path string) error
func (e *Editor) SaveFile(path string) error
func (e *Editor) FileName() string
func (e *Editor) Modified() bool
```

- `LoadFile`: reads file via `os.ReadFile`, calls `SetText()`, clears undo state, stores filename, clears modified flag
- `SaveFile`: writes `Text()` to disk via `os.WriteFile`, clears modified flag
- Line ending detection on load: detects `\r\n` vs `\n`, stores preference, writes back with the detected ending
- Modified flag: set to true on any edit, cleared on save/load

### 6.6 Close Behavior (Save Prompt)

When the EditWindow receives `CmClose` and the editor has unsaved changes (`Modified() == true`):

1. Show a modal dialog: "Save changes to [filename]?" with Yes / No / Cancel buttons
2. **Yes**: call `SaveFile()` (prompt for filename if untitled via a FileDialog), then close
3. **No**: discard changes, close the window
4. **Cancel**: abort the close, window stays open

This matches original `TFileEditor::valid()` at `tfiledtr.cpp:264-291`. The `Valid()` method on EditWindow implements this logic — it returns false to block closing when the user chooses Cancel.

### 6.7 Pragmatic Omission: WordStar Chord Sequences

The original TEditor supports WordStar-compatible two-key chord sequences (Ctrl+K for block operations, Ctrl+Q for quick movement). These are a legacy input method largely superseded by the Ctrl+C/X/V/F/H shortcuts we already support. Implementing chord state tracking (a "pending chord prefix" mode that changes the meaning of the next keystroke) adds complexity for minimal modern value. **Omitted from this spec.** If needed later, it can be added as a standalone enhancement.

### 6.8 TEditWindow

Window subclass that constructs and wires everything:

```go
func NewEditWindow(bounds Rect, filename string) *EditWindow
```

Internally creates:
- Vertical ScrollBar (right edge, `GfGrowLoX | GfGrowHiX | GfGrowHiY`)
- Horizontal ScrollBar (bottom edge, `GfGrowLoY | GfGrowHiY | GfGrowHiX`)
- Editor filling the client area, linked to both scrollbars
- Indicator in the frame footer area (left side)
- Window title = filename basename, or "Untitled" if empty
- Minimum window size: 24×6

If `filename` is non-empty, calls `Editor.LoadFile(filename)` during construction.

### 6.9 Shared Clipboard

The existing Memo clipboard is a package-level `var clipboard string`. Editor reuses this — all Memo and Editor instances share the same clipboard. No change needed.

### 6.10 New Commands

```go
CmFind
CmReplace
CmSearchAgain
CmIndicatorUpdate
```

### 6.11 Color Scheme

No new entries. Editor uses Memo's existing `MemoNormal` and `MemoSelected`. Indicator uses `WindowTitle` style for its text.

### 6.12 Demo Integration

- Replace Window 3 "Notes" (currently plain Memo) with an EditWindow containing an Editor pre-loaded with sample text
- Add menu items: Edit > Find (Ctrl+F), Edit > Replace (Ctrl+H), Edit > Search Again (F3)
- Indicator visible in the window footer showing line:col and modified flag

---

## 7. TFileDialog

Reference: `stddlg.h`, file dialog source files (`tfiledlg.cpp`, `tfilecol.cpp`, `tfillist.cpp`, `tdirlist.cpp`)

### 7.1 Architecture

TFileDialog is a modal Dialog containing specialized sub-widgets that coordinate via broadcast events.

```
FileDialog (Dialog)
├── Label "File ~N~ame:" → linked to FileInputLine
├── FileInputLine (InputLine + history)
├── History (linked to FileInputLine)
├── Label "~F~iles:" → linked to FileList
├── FileList (sorted ListBox of directory entries)
├── FileInfoPane (file metadata display)
└── Buttons: Open/OK/Replace/Clear, Cancel (configurable)
```

### 7.2 FileInputLine

InputLine subclass. Detects whether the user's input is a wildcard pattern, a directory path, or a plain filename, and handles each differently on Enter.

```go
type FileInputLine struct {
    *InputLine
}
```

Behavior on Enter (via the dialog's `valid()` method):
- If text contains wildcard characters (`*`, `?`): store as new wildcard filter, broadcast `CmFileFilter`, rescan directory — do NOT close dialog
- If text names a directory (detectable via `os.Stat`): change to that directory, rescan with current wildcard — do NOT close dialog
- If text is a plain filename: resolve to absolute path, close dialog with `CmOK`

### 7.3 FileList

A sorted ListBox displaying directory entries.

```go
type FileEntry struct {
    Name    string
    Size    int64
    ModTime time.Time
    IsDir   bool
}

func NewFileList(bounds Rect) *FileList
func (fl *FileList) ReadDirectory(dir, wildcard string) error
```

Display format: directories show `name/`, files show `name` (left-aligned). Sorted: `..` always first, then directories alphabetically, then files alphabetically.

`ReadDirectory` scans using `os.ReadDir`:
- Filters files by wildcard using `filepath.Match`
- Directories are always shown (not filtered by wildcard)
- Adds `..` entry if not at filesystem root
- Hidden files (starting with `.`) are excluded by default

On focus change: broadcasts `CmFileFocused` with the focused `*FileEntry` as `event.Info`.
On Enter or double-click on a directory: changes into it, calls `ReadDirectory` with current wildcard.
On Enter or double-click on a file: synthesizes a `CmOK` event (matching original's `cmFileDoubleClicked` → `putEvent(cmOK)` pattern), which triggers the dialog's `valid()` method.

**Incremental keyboard search:** When the FileList is focused and the user types printable characters, the list incrementally searches for a matching filename. Backspace undoes the last search character. `.` jumps to the file extension area. This matches original `TSortedListBox` behavior in `stddlg.cpp:110-185`. The search buffer is cleared after 1 second of inactivity or when a navigation key is pressed.

### 7.4 FileInfoPane

Non-focusable display showing metadata of the currently focused file.

```go
type FileInfoPane struct {
    BaseView
    entry *FileEntry
}
```

Display format:
```
  main.go         1,234  May  3, 2026  2:15pm
```

- Filename left-aligned
- Size in bytes with comma separators (or `<DIR>` for directories), right-aligned
- Modification date and time
- Listens for `CmFileFocused` broadcasts to update
- Not selectable

### 7.5 FileInputLine Auto-fill Coordination

When the FileInputLine does NOT have focus, it listens for `CmFileFocused` broadcasts from the FileList. On receiving one:
- If the focused entry is a file: sets the InputLine text to the filename
- If the focused entry is a directory: sets the InputLine text to `directoryName/` + current wildcard

This means navigating the file list with arrow keys automatically updates the filename input field, matching the original behavior in `stddlg.cpp:75-91`.

When the FileInputLine HAS focus, `CmFileFocused` is ignored (the user is typing, don't overwrite).

### 7.6 FileDialog Constructor

```go
type FileDialogFlag int
const (
    FdOpenButton    FileDialogFlag = 1 << iota
    FdOKButton
    FdReplaceButton
    FdClearButton
    FdHelpButton
)

func NewFileDialog(wildcard, title string, flags FileDialogFlag) *FileDialog
func NewFileDialogInDir(dir, wildcard, title string, flags FileDialogFlag) *FileDialog
func (fd *FileDialog) FileName() string  // returns selected absolute path
```

Default size: 52×20. Layout:
- Row 2: Label "File ~N~ame:"
- Row 3: FileInputLine (columns 3-34) + History (columns 35-37)
- Row 5: Label "~F~iles:"
- Rows 6-15: FileList (columns 3-35) with scrollbar
- Rows 16-17: FileInfoPane (columns 3-35)
- Buttons stacked on right edge (columns 38-49), starting at row 3

First button in the flags gets the `BfDefault` flag. Cancel is always present.

### 7.7 Resizability

The FileDialog is resizable (`WfGrow` flag set on the window). All internal components have appropriate GrowMode flags so they resize with the dialog. Minimum size: 49×19. The dialog auto-sizes based on available screen dimensions, matching original `tfildlg.cpp:62,141-167`.

### 7.8 Wildcard Handling

Uses Go's `filepath.Match` for pattern matching. Supports `*` and `?` wildcards. Multiple patterns separated by `;` (e.g., `*.go;*.mod`) are tried in order — a file matches if it matches any pattern. When no wildcard is specified, defaults to `*`.

### 7.9 New Commands

```go
CmFileOpen      // Open button pressed
CmFileReplace   // Replace button pressed  
CmFileClear     // Clear button pressed
CmFileFocused   // FileList focus changed (Info field: *FileEntry)
CmFileFilter    // Wildcard pattern changed
```

### 7.10 Cross-platform

Unix/macOS only. No drive enumeration. Dialog starts in the working directory or a caller-specified directory. Path separator is always `/`.

### 7.11 Color Scheme

No new entries. Uses existing Dialog, Label, InputLine, ListBox, Button colors.

### 7.12 Demo Integration

- Add menu item: File > Open (F3) that opens `NewFileDialog("*.go", "Open File", FdOpenButton)`
- Selected filename shown in status line or loaded into the Editor window
- Demonstrates: directory navigation, wildcard filtering, file metadata display

---

## 8. TColorDialog

Reference: `colorsel.h`, color selector source files

### 8.1 Architecture

TColorDialog is a modal Dialog where cascading selections flow left-to-right: pick a group → pick an item → pick foreground/background colors → see preview. Components coordinate via broadcast events.

```
ColorDialog (Dialog)
├── Label "~G~roup:" → linked to ColorGroupList
├── ColorGroupList (ListViewer of color groups)
├── Label "~I~tem:" → linked to ColorItemList
├── ColorItemList (ListViewer of items within selected group)
├── Label "Foreground"
├── ColorSelector (4×4 grid, 16 foreground colors)
├── Label "Background"
├── ColorSelector (4×2 grid, 8 background colors)
├── ColorDisplay (preview pane)
└── Buttons: OK, Cancel
```

### 8.2 Data Model

```go
type ColorItem struct {
    Name  string
    Index int    // palette entry index this item controls
}

type ColorGroup struct {
    Name  string
    Items []ColorItem
}
```

Groups organize items semantically: "Desktop" group contains "Background"; "Windows" group contains "Frame Active", "Frame Inactive", "Title", etc.

```go
func NewColorDialog(groups []ColorGroup, palette []tcell.Style) *ColorDialog
func (cd *ColorDialog) Palette() []tcell.Style  // returns modified palette on OK
```

### 8.3 ColorGroupList

A ListViewer showing group names. On focus change, broadcasts `CmNewColorGroup` with the group index as `event.Info`. ColorItemList listens and swaps its displayed items to show the newly selected group's items.

### 8.4 ColorItemList

A ListViewer showing item names within the currently selected group. On focus change, broadcasts `CmNewColorIndex` with the selected item's palette index as `event.Info`. The ColorSelectors and ColorDisplay listen and update to show that palette entry's current foreground/background color values.

### 8.5 ColorSelector

A grid widget showing terminal colors. Not a ListViewer — a custom focusable widget that draws colored cells.

**Foreground selector:** 4 columns × 4 rows = 16 colors.
**Background selector:** 4 columns × 2 rows = 8 colors (terminals traditionally support only 8 background colors).

```go
type ColorSelector struct {
    BaseView
    numColors int  // 16 for foreground, 8 for background
    selected  int  // currently highlighted color index
    kind      int  // 0 = foreground, 1 = background
}

func NewColorSelector(bounds Rect, kind int) *ColorSelector
```

- **Draw:** Each cell is 3 chars wide. Selected cell shows a centered marker character. Cells are filled with their respective background color.
- **Mouse:** Click selects the color at that grid position. Grid position: `col = x / 3`, `row = y`, `index = row * 4 + col`.
- **Keyboard:** Arrow keys navigate the grid. Up/Down change rows, Left/Right change columns. **All directions wrap:** left from column 0 goes to last column, up from row 0 goes to last row, and vice versa. Foreground wraps in a 4×4 grid (indices 0-15), background in a 4×2 grid (indices 0-7). Matching original `colorsel.cpp:182-216`.
- **On change:** Broadcasts `CmColorForegroundChanged` or `CmColorBackgroundChanged` with the new color index.

### 8.6 TMonoSelector

A cluster (radio group) with 4 fixed monochrome attributes, shown instead of the color selectors when `showMarkers` mode is active. Matches original `colorsel.cpp:267-328`.

```go
type MonoSelector struct {
    BaseView
    selected int  // 0-3
}

func NewMonoSelector(bounds Rect) *MonoSelector
```

Four fixed options:
- Normal (attribute 0x07)
- Highlight (attribute 0x0F)
- Underline (attribute 0x01)
- Inverse (attribute 0x70)

Keyboard: Up/Down to navigate, Space/Enter to select. On change, broadcasts `CmColorForegroundChanged` with the attribute value.

**Pragmatic note:** Terminal monochrome displays are rare today. We implement MonoSelector for completeness but the ColorDialog defaults to color mode. A `SetMonoMode(bool)` method on the dialog can toggle between color selectors and the mono selector.

### 8.7 Per-Group Focus Memory

When the user switches between groups, the focused item within each group is remembered. Switching back to a previously visited group restores the previously focused item. This matches original `colorsel.cpp:506-512,632-641`.

Each `ColorGroup` stores its last-focused item index. `CmNewColorGroup` triggers saving the current group's focused index before loading the new group's items.

This focus memory persists for the lifetime of the dialog (not across dialog invocations — keeping it simple).

### 8.8 ColorDisplay

Non-focusable preview pane showing sample text in the currently selected foreground+background combination.

```go
type ColorDisplay struct {
    BaseView
    fg, bg int
}
```

- Draws repeating sample text `"Text Text Text"` using a `tcell.Style` constructed from `fg` and `bg` color indices
- 3 rows tall
- Listens for `CmColorForegroundChanged` and `CmColorBackgroundChanged` to update

### 8.9 Broadcast Flow

```
ColorGroupList ──CmNewColorGroup──→ ColorItemList (swap items)
ColorItemList  ──CmNewColorIndex──→ ColorSelector × 2 (load current fg/bg from palette)
                                  → ColorDisplay (load current color)
ColorSelector  ──CmColorFgChanged─→ ColorDisplay (update preview)
                                  → palette entry updated
ColorSelector  ──CmColorBgChanged─→ ColorDisplay (update preview)
                                  → palette entry updated
```

### 8.10 New Commands

```go
CmNewColorGroup
CmNewColorIndex
CmColorForegroundChanged
CmColorBackgroundChanged
```

### 8.11 Color Scheme

New entries for the color selector grid:

| Entry | Purpose |
|-------|---------|
| `ColorSelectorNormal` | Unselected color cell border/text |
| `ColorSelectorCursor` | Selected color cell marker |
| `ColorDisplayText` | Preview pane text (overridden by live fg/bg) |

All five themes updated.

### 8.12 Demo Integration

- Add menu item: Options > Colors that opens `NewColorDialog(defaultGroups, currentPalette)`
- `defaultGroups` derived from the current theme's ColorScheme field names, organized by widget category
- On OK: apply the modified palette to the running application for live theme change
- On Cancel: no change

---

## 9. TOutline (Tree View)

Reference: `outline.h`, outline source files

### 9.1 Architecture

TOutline is a scrollable tree view. It flattens a node tree into a vertical list for display. Builds on `OutlineViewer` (display, scrolling, input) with a concrete `Outline` that owns a `TNode` tree.

```
OutlineViewer (base — flattened tree display, scrolling, keyboard/mouse)
 └── Outline (concrete — owns TNode tree, implements abstract methods)
```

### 9.2 TNode

Sibling-linked tree structure matching original:

```go
type TNode struct {
    Text     string
    Children *TNode  // first child (nil if leaf)
    Next     *TNode  // next sibling (nil if last)
    Expanded bool
}

func NewNode(text string, children *TNode, next *TNode) *TNode
```

Nodes form a **forest**: the `root` parameter is the first of potentially many top-level sibling nodes linked by `Next`. Each can have `Children` which are themselves sibling-linked lists. `Expanded` controls whether children are visible. The outline supports multiple root-level nodes — this is not a single-root tree but a sibling chain at every level, matching original `toutline.cpp:310-319`.

### 9.3 OutlineViewer

Handles visual representation and interaction.

```go
type OutlineViewer struct {
    BaseView
    root       *TNode
    focusedIdx int      // index in flattened visible list
    deltaY     int      // vertical scroll offset
    vScrollBar *ScrollBar
    hScrollBar *ScrollBar
}

func (ov *OutlineViewer) SetVScrollBar(sb *ScrollBar)
func (ov *OutlineViewer) SetHScrollBar(sb *ScrollBar)
```

**Flattening algorithm:** Depth-first traversal. Visit a node (count it as a visible row), then if expanded, recursively visit its children, then visit its next sibling. Only expanded nodes' children contribute to the visible row count.

```go
func (ov *OutlineViewer) visibleCount() int
func (ov *OutlineViewer) nodeAt(idx int) (*TNode, int)  // returns node and nesting level
```

### 9.4 Drawing

Each visible row renders two parts:

**Graph prefix** — tree structure lines (3 characters per nesting level):
- `│  ` — vertical connector for ancestors that have more siblings below
- `├──` — non-last sibling connector
- `└──` — last sibling connector
- After the connector: `+` if node has children but is collapsed, `─` if expanded, ` ` if leaf

**Node text** — after the graph prefix

Example:
```
├─── Project
│  ├── src
│  │  ├── main.go
│  │  └── util.go
│  └─+ docs
└── LICENSE
```

The `+` on "docs" indicates it has children but is collapsed.

**Colors:** 4 palette entries:
- `OutlineNormal` — normal node text
- `OutlineFocused` — focused row highlight
- `OutlineSelected` — selected node
- `OutlineCollapsed` — collapsed node that has hidden children (visual cue that expansion is possible)

### 9.5 Keyboard Handling

| Key | Action |
|-----|--------|
| Up | Move focus to previous visible row |
| Down | Move focus to next visible row |
| Right | If focused node has children and is collapsed: expand. If already expanded: move focus to first child |
| Left | If focused node is expanded: collapse. If collapsed or leaf: move focus to parent |
| Enter | Toggle expand/collapse |
| `+` | Expand focused node |
| `-` | Collapse focused node |
| `*` | Expand all descendants of focused node recursively |
| PgUp | Move focus up by `bounds.Height() - 1` rows |
| PgDn | Move focus down by `bounds.Height() - 1` rows |
| Home | Move focus to first visible node |
| End | Move focus to last visible node |

After expand/collapse, the focused node remains focused (its flattened index may change but the node identity is preserved). Scroll position adjusts to keep the focused node visible.

### 9.6 Mouse Handling

- **Click on graph area** (first `level * 3 + 3` characters): calls `adjust()` — toggles expand/collapse of that node. This is distinct from `selected()`.
- **Click on text area**: move focus to that row
- **Double-click on text area**: calls `selected()` callback (the "item activated" action). If the node has children, also toggles expand/collapse.

**Distinction between `adjust()` and `selected()`:** `adjust()` only toggles expand/collapse state. `selected()` is the "user chose this item" callback — it fires `OnSelect`. Keyboard `+`/`-`/`*` call `adjust()`. Enter calls `selected()`. This matches original `toutline.cpp:516-517,547`.

### 9.7 Scrollbar Integration

Same pattern as Memo and ListViewer:
- Vertical scrollbar: range = `visibleCount()`, page = `bounds.Height()`, value = `deltaY`. Updated after every expand/collapse/navigation.
- Horizontal scrollbar: range = max line width across all visible rows (graph prefix + text), page = `bounds.Width() / 2`, value = `deltaX`.

### 9.8 Outline (Concrete)

```go
type Outline struct {
    OutlineViewer
}

func NewOutline(bounds Rect, root *TNode) *Outline
func (o *Outline) Root() *TNode
func (o *Outline) SetRoot(root *TNode)
func (o *Outline) SetOnSelect(fn func(node *TNode))
```

`SetRoot` replaces the tree, resets focus to index 0, resets scroll, calls `Update()`.

### 9.9 Update After Mutations

```go
func (o *Outline) Update()
```

Must be called after any external modification to the node tree (adding/removing nodes, changing text). Recomputes the visible row count and maximum line width, updates scrollbar ranges, and adjusts focus if it's now out of bounds. Matching original `toutline.cpp:587-594`.

This is the caller's responsibility — the Outline does not detect tree mutations automatically. `SetRoot()`, `adjust()` (expand/collapse), and all keyboard/mouse handlers call `Update()` internally.

### 9.10 Visitor Methods

Matching original TV's `firstThat` and `forEach`:

```go
func (o *Outline) ForEach(fn func(node *TNode, level int))
func (o *Outline) FirstThat(fn func(node *TNode, level int) bool) *TNode
```

`ForEach` visits every node in the tree (depth-first, regardless of expanded state). `FirstThat` returns the first node where `fn` returns true.

### 9.11 New Color Scheme Entries

| Entry | Purpose |
|-------|---------|
| `OutlineNormal` | Normal node text |
| `OutlineFocused` | Focused row |
| `OutlineSelected` | Selected node |
| `OutlineCollapsed` | Collapsed node with children |

All five themes updated.

### 9.12 Demo Integration

- Add Window 4 "Outline" containing an Outline widget with a sample tree:
  ```
  Project
  ├── src
  │  ├── main.go
  │  └── util.go
  ├── docs
  │  ├── README.md
  │  └── DESIGN.md
  ├── tests
  │  └── e2e_test.go
  └── go.mod
  ```
- Vertical scrollbar linked
- Pre-expand "src" to show initial tree state

---

## 10. New Commands Summary

Commands to add to `tv/command.go`:

```go
// File dialog
CmFileOpen
CmFileReplace
CmFileClear
CmFileFocused
CmFileFilter

// Editor
CmFind
CmReplace
CmSearchAgain
CmIndicatorUpdate

// Color dialog
CmNewColorGroup
CmNewColorIndex
CmColorForegroundChanged
CmColorBackgroundChanged
```

---

## 11. Color Scheme Summary

New entries across all five themes:

| Entry | Widget | Purpose |
|-------|--------|---------|
| `ClusterDisabled` | CheckBoxes/RadioButtons | Disabled item text |
| `ColorSelectorNormal` | ColorSelector | Unselected cell |
| `ColorSelectorCursor` | ColorSelector | Selected cell marker |
| `ColorDisplayText` | ColorDisplay | Preview text |
| `OutlineNormal` | Outline | Normal node |
| `OutlineFocused` | Outline | Focused row |
| `OutlineSelected` | Outline | Selected node |
| `OutlineCollapsed` | Outline | Collapsed expandable node |

---

## 12. Demo App Updates Summary

| Window/Dialog | Changes |
|--------------|---------|
| Window 1 "File Manager" | Add validated InputLine (Port with RangeValidator). CheckBoxes/RadioButtons use multi-column layout. |
| Window 2 "Editor" | Optionally show 2-column ListViewer |
| Window 3 "Notes" → EditWindow | Replace plain Memo with EditWindow (Editor + Indicator + ScrollBars) |
| New Window 4 "Outline" | Outline widget with sample tree |
| File > Open menu | Opens TFileDialog |
| Options > Colors menu | Opens TColorDialog with live palette editing |
| Edit menu | Find (Ctrl+F), Replace (Ctrl+H), Search Again (F3) |

---

## 13. Deferred Features (Future Work)

Explicitly out of scope:
- **Multi-level undo** for Editor (only single-level, matching original TV)
- **TCollection / TSortedCollection** (Go slices + sort.Slice cover these use cases)
- **Resource/stream system** (Go's encoding/gob or encoding/json replaces serialization)
- **Word wrap** for Editor/Memo (horizontal scrolling only, matching original TEditor)
- **TDirListBox** as a separate widget (directory navigation is handled within TFileDialog)
