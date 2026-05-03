# Missing Widgets Behavioral Spec

**Goal:** Add TListBox, TMemo, and THistory to TurboView, and validate the existing TLabel implementation against original Turbo Vision behavior.

**Reference authority:** The magiblot/tvision C++ source code (https://github.com/magiblot/tvision) is the definitive behavioral reference. The original Borland documentation provides API reference only — widget behavior is only fully described by the implementation.

**Three batches:**
1. TLabel validation (fix behavioral gaps vs original TV)
2. TListBox + TMemo (new widgets)
3. THistory (InputLine companion)

---

## 1. TLabel Validation

The existing `tv/label.go` has five behavioral gaps compared to original TV's TLabel (source: `tlabel.cpp`, `dialogs.h`).

### 1.1 Missing OfPreProcess

**Current:** Label sets only `OfPostProcess`.

**Original:** TLabel sets `ofPreProcess | ofPostProcess`.

**Required behavior:** Label must set both `OfPreProcess` and `OfPostProcess`. This ensures Alt+shortcut matching fires in preprocess (before the focused view), so the shortcut works reliably regardless of what the focused view might consume.

**Why it matters:** Currently works by accident because InputLine passes through Alt+Rune events. A future focusable widget that consumed Alt+letters would break Label shortcuts.

### 1.2 Text Starts at Column 0 Instead of Column 1

**Current:** `Draw()` starts rendering text at `x := 0`.

**Original:** `moveCStr(1, text, color)` — text starts at column 1. Column 0 is reserved for a monochrome marker (space on color displays).

**Required behavior:** Text must start at column 1. Column 0 is always a space (we do not implement monochrome markers, but the margin must exist for visual alignment with other dialog controls).

### 1.3 Missing Background Fill

**Current:** `Draw()` writes text characters only. Trailing cells are unfilled.

**Original:** `moveChar(0, ' ', color, size.x)` fills the entire widget width with spaces (in the appropriate background color) before drawing text.

**Required behavior:** `Draw()` must fill the entire bounds width with the normal/highlighted background color before rendering text. This ensures clean rendering when text changes to something shorter, and provides consistent background color across the full widget width.

### 1.4 Missing Plain-Letter Shortcut Matching in Postprocess

**Current:** Only matches `Alt+letter` shortcuts.

**Original:** Matches Alt+letter in any phase, AND matches the plain unmodified letter during the postprocess phase only. The postprocess check is: `owner.phase == phPostProcess && c == toupper(charCode)`.

**Required behavior:** In addition to Alt+letter matching, Label must also match the plain shortcut letter (case-insensitive) during postprocess. This fires only when the focused view did not consume the keystroke.

**Pragmatic decision:** The Group's three-phase dispatch does not currently expose which phase is active to child views, and adding phase tracking to support plain-letter matching would add complexity for marginal benefit. This spec takes the pragmatic approach: **only match Alt+letter** (skip plain-letter matching). This preserves 95% of original behavior without requiring Group infrastructure changes. Plain-letter matching can be added in a future phase if needed.

### 1.5 focusLink Must Check OfSelectable

**Current:** `HandleEvent` calls `owner.SetFocusedChild(l.link)` without checking whether the link is selectable.

**Original:** `focusLink` checks `link->options & ofSelectable` before calling `link->focus()`.

**Required behavior:** Before attempting to focus the linked view, Label must check `l.link.HasOption(OfSelectable)`. If the link is not selectable (e.g., disabled), the event is still cleared but no focus change occurs.

### 1.6 Palette Entries (no change needed)

Original TLabel has 4 palette entries: normal text, selected text, normal shortcut, selected shortcut. In the default Borland palette, entries 3 and 4 (both shortcut colors) map to the same index — the shortcut color does not change with highlight state.

Our existing ColorScheme has `LabelNormal`, `LabelHighlight`, and `LabelShortcut`. This is functionally equivalent — the shortcut style is the same regardless of highlight state, matching the original default. No change needed.

---

## 2. TListBox

A composite widget that bundles a ListViewer with a vertical ScrollBar, wired together as a single unit. In original TV, TListBox inherits from TListViewer; in our Go codebase, it is a Container using the Group facade pattern (like CheckBoxes, RadioButtons).

### 2.1 Architecture

TListBox implements the `Container` interface. Internally it owns a `*Group` and acts as its facade. The Group contains two children: a `*ListViewer` and a `*ScrollBar`.

```
TListBox (Container, facade)
  └── Group
       ├── ListViewer (fills width minus 1, full height)
       └── ScrollBar (vertical, width 1, right edge)
```

The ListViewer and ScrollBar are wired together via `ListViewer.SetScrollBar()`.

### 2.2 Constructor

```go
func NewListBox(bounds Rect, ds ListDataSource) *ListBox
```

Creates a ListBox with the given data source. Internally:
1. Creates a vertical ScrollBar at the right edge (width 1, full height)
2. Creates a ListViewer filling the remaining width
3. Wires them via `ListViewer.SetScrollBar(sb)`
4. Sets `OfSelectable` on the ListBox itself
5. Sets GrowMode defaults: ListViewer gets `GfGrowHiX | GfGrowHiY`, ScrollBar gets `GfGrowLoX | GfGrowHiY`

Convenience constructor for string lists:

```go
func NewStringListBox(bounds Rect, items []string) *ListBox
```

Creates a `StringList` data source and passes it to `NewListBox`.

### 2.3 Public API

```go
func (lb *ListBox) ListViewer() *ListViewer  // access internal ListViewer
func (lb *ListBox) ScrollBar() *ScrollBar    // access internal ScrollBar
func (lb *ListBox) Selected() int            // delegates to ListViewer.Selected()
func (lb *ListBox) SetSelected(index int)    // delegates to ListViewer.SetSelected()
func (lb *ListBox) DataSource() ListDataSource
func (lb *ListBox) SetDataSource(ds ListDataSource)
```

`SetDataSource` replaces the data source, resets selection to 0, updates scrollbar range, and triggers a redraw. This matches original `TListBox::newList` behavior.

### 2.4 Container Interface

Delegates to internal Group: `Insert`, `Remove`, `Children`, `FocusedChild`, `SetFocusedChild`, `ExecView`, `BringToFront`. This follows the same delegation pattern as CheckBoxes.

### 2.5 Drawing

TListBox.Draw delegates to its children — the ListViewer and ScrollBar each draw themselves into their respective SubBuffer regions. No custom drawing logic on TListBox itself.

### 2.6 Event Handling

TListBox routes events to the internal Group, which dispatches to the focused child (always the ListViewer — the ScrollBar is not selectable).

Mouse events are routed by position: clicks on the ScrollBar area go to the ScrollBar, clicks on the ListViewer area go to the ListViewer. This matches the Group's standard mouse routing.

### 2.7 Focus Behavior

The ListBox itself is the focusable unit in its parent container. When it receives `SfSelected`, the internal Group focuses the ListViewer (which gets `SfSelected`). The ScrollBar never receives focus (it is not selectable, matching the scrollbar fix from the previous session).

When the ListBox loses focus (`SfSelected` cleared), the ListViewer's focused-item highlighting follows the cluster-unfocused pattern: the focused item's visual differentiation disappears. However, unlike CheckBoxes/RadioButtons, the ListViewer already handles this correctly — it checks its own `SfSelected` state in its `Draw()` method and only uses `focusedStyle` when selected.

### 2.8 Scrollbar Synchronization

Handled by the existing `ListViewer.SetScrollBar` wiring:
- ListViewer updates ScrollBar on navigation (range, value, page size) via direct method calls
- ScrollBar fires its `OnChange` callback on user interaction
- ListViewer responds by updating `topIndex` and redrawing

No new synchronization code is needed — TListBox simply calls `ListViewer.SetScrollBar(sb)` during construction, which sets up the `OnChange` callback wiring.

### 2.9 Selection Model

Single selection. `Selected()` returns the focused item index. Space and double-click fire the `OnSelect` callback (if set) as a direct Go function call. This is existing ListViewer behavior — there is no command broadcast for item selection.

### 2.10 Empty List

When the data source has 0 items, the ListViewer displays `<empty>` in the first row using the normal color. This is existing ListViewer behavior (if not already implemented, it must be added to match original TV).

### 2.11 Color Scheme

Uses existing ListViewer palette entries: `ListNormal`, `ListSelected`, `ListFocused`. The ScrollBar uses `ScrollBar`, `ScrollThumb`.

No new color scheme entries needed for TListBox — it uses existing ListViewer and ScrollBar palette entries.

---

## 3. TMemo

A multi-line text editing widget for use in dialogs and other containers. This is a simplified implementation covering core editing behaviors. Advanced features (undo, find/replace, insert/overwrite toggle) are deferred to a future phase.

### 3.1 Architecture

TMemo is a plain Widget (implements `View`, not `Container`). It does not own sub-views. Scrollbars, if desired, are created externally and linked via setter methods — matching the original TV pattern where TEditor receives scrollbars from its parent window.

### 3.2 Buffer Model

Text is stored as `[][]rune` — a slice of lines, each line a slice of runes. This is:
- **Unicode-native**: rune-based, no byte-offset concerns
- **No artificial limits**: grows as needed, no pre-allocated buffer
- **Efficient for dialog-sized text**: line-based operations (split, join, insert) are O(line-length), which is fine for text that fits in a dialog

Each line stores only its content — no trailing newline characters. Line endings are implicit (the boundary between adjacent slices).

### 3.3 Cursor State

```
cursorRow int  // current line (0-based)
cursorCol int  // rune position within line (0-based)
```

The cursor position is always valid: `cursorRow` is clamped to `[0, len(lines)-1]`, `cursorCol` is clamped to `[0, len(lines[cursorRow])]` (can be at end of line for appending).

### 3.4 Selection Model

```
selStartRow, selStartCol int  // selection anchor
selEndRow, selEndCol int      // selection extent
```

Selection is tracked as two (row, col) positions. When no selection exists, start == end. The cursor is always at one end of the selection (either start or end, depending on direction of extension).

**Shift+arrow** extends the selection. Moving without Shift collapses it. `Ctrl+A` selects all text.

### 3.5 Viewport Scrolling

```
deltaX int  // horizontal scroll offset (columns)
deltaY int  // vertical scroll offset (lines)
```

The viewport shows `size.Height` lines starting from `deltaY`, scrolled horizontally by `deltaX` columns.

**Auto-scroll:** After every cursor movement, the viewport adjusts to keep the cursor visible:
- If cursor is above viewport: `deltaY = cursorRow`
- If cursor is below viewport: `deltaY = cursorRow - height + 1`
- If cursor is left of viewport: `deltaX = cursorCol`
- If cursor is right of viewport: `deltaX = cursorCol - width + 1`
- Otherwise: no scroll change (minimal scrolling)

### 3.6 Drawing

Each visible line is rendered independently:

1. For each row `y` in `[0, height)`:
   - Calculate the line index: `lineIdx = deltaY + y`
   - If `lineIdx >= len(lines)`: fill the row with spaces in normal color
   - Otherwise: render `lines[lineIdx]` starting from column `deltaX`
2. For each character in the visible portion of the line:
   - If the character position falls within the selection range: use selected color
   - Otherwise: use normal color
3. After the line content ends, fill remaining columns with spaces
   - If the selection extends past line end (selecting the newline), those trailing spaces use selected color

**Tab rendering:** Tab characters (`\t`) expand to the next multiple of 8 columns, rendered as spaces. Tab stops are at columns 0, 8, 16, 24, etc.

**No word wrap.** Lines longer than the view width are scrolled horizontally, matching original TEditor behavior.

### 3.7 Keyboard Handling

TMemo blocks the Tab key — it passes through for focus navigation, matching original TMemo behavior.

#### Cursor Movement

| Key | Action |
|-----|--------|
| Left | Move cursor one rune left; wraps to end of previous line |
| Right | Move cursor one rune right; wraps to start of next line |
| Up | Move cursor one line up, preserving column (clamp at line end) |
| Down | Move cursor one line down, preserving column (clamp at line end) |
| Home | Smart Home: move to first non-whitespace character on the line; if already there, move to column 0 |
| End | Move to end of current line |
| Ctrl+Home | Move to start of document (row 0, col 0) |
| Ctrl+End | Move to end of document (last row, last col) |
| PgUp | Move cursor up by `height - 1` lines |
| PgDn | Move cursor down by `height - 1` lines |
| Ctrl+Left | Move to start of previous word |
| Ctrl+Right | Move to start of next word |

**Word boundaries:** A word boundary is a transition between character classes: whitespace, punctuation (`!\"#$%&'()*+,-./:;<=>?@[\\]^{|}~`), and word characters (everything else). This matches original TEditor's `getCharType` four-class system (whitespace, line break, punctuation, word char), with line breaks treated as their own boundary.

**Vertical movement column preservation:** When moving up/down, the cursor attempts to land at the same column position. If the target line is shorter, the cursor clamps to end-of-line. The target column is recomputed from the current position on each vertical move (not persisted across multiple moves), matching original TEditor behavior.

All cursor movement keys support **Shift+ variants** to extend selection.

#### Text Editing

| Key | Action |
|-----|--------|
| Printable character | Insert at cursor (replaces selection if any) |
| Enter | Insert newline; if auto-indent enabled, copy leading whitespace from current line |
| Backspace | Delete selection if any; otherwise delete one rune before cursor (join lines if at start of line) |
| Delete | Delete selection if any; otherwise delete one rune after cursor (join lines if at end of line) |
| Ctrl+Backspace | Delete from cursor to start of previous word |
| Ctrl+Delete | Delete from cursor to start of next word |
| Ctrl+Y | Delete entire current line |

#### Clipboard

| Key | Action |
|-----|--------|
| Ctrl+C | Copy selection to clipboard |
| Ctrl+X | Cut selection to clipboard |
| Ctrl+V | Paste from clipboard at cursor (replaces selection if any) |
| Ctrl+A | Select all text |

Clipboard is shared with InputLine (the existing `clipboard` package-level variable in `tv/input_line.go`).

#### Not Handled (passes through)

- Tab (focus navigation)
- Alt+anything (menu shortcuts, label shortcuts)
- F-keys (window management)

### 3.8 Mouse Handling

| Action | Behavior |
|--------|----------|
| Click | Position cursor at clicked location; collapse selection |
| Drag | Select text from click point to drag point; auto-scroll when dragging outside bounds |
| Double-click | Select the word under the cursor |
| Triple-click | Select the entire line |

**Click-to-position:** Convert mouse (x, y) to buffer position: `row = deltaY + y`, `col = deltaX + x`, clamped to valid range. Tab characters affect column mapping (a click in the middle of a tab's visual span lands at the tab character's position).

**Auto-scroll during drag:** When the mouse moves outside the view bounds during a drag, scroll by 1 unit per auto-repeat event in the appropriate direction. This uses the Application's mouse-auto mechanism (existing `startMouseAuto`/`stopMouseAuto`).

### 3.9 Scrollbar Integration

TMemo does not own scrollbars. They are linked via setters:

```go
func (m *Memo) SetHScrollBar(sb *ScrollBar)
func (m *Memo) SetVScrollBar(sb *ScrollBar)
```

After any change (cursor movement, text edit, scroll), TMemo updates the linked scrollbars:
- Vertical: `SetRange(0, lineCount - 1)`, `SetPageSize(height - 1)`, `SetValue(deltaY)`
- Horizontal: `SetRange(0, maxLineWidth)`, `SetPageSize(width / 2)`, `SetValue(deltaX)`

TMemo sets `OnChange` callbacks on linked scrollbars to update `deltaX`/`deltaY` when the user interacts with the scrollbar directly. This follows the same callback pattern used by ListViewer.

When TMemo gains/loses active state (`SfSelected`), it shows/hides linked scrollbars, matching original TEditor behavior.

### 3.10 Options and Constructor

```go
type MemoOption func(*Memo)

func WithAutoIndent(enabled bool) MemoOption
func WithScrollBars(h, v *ScrollBar) MemoOption

func NewMemo(bounds Rect, opts ...MemoOption) *Memo
```

Defaults:
- Auto-indent: enabled (matches original TMemo)
- Scrollbars: none (must be linked explicitly)
- GrowMode: `GfGrowHiX | GfGrowHiY`
- Options: `OfSelectable` set
- Cursor: visible, at (0, 0)

### 3.11 Public API

```go
func (m *Memo) Text() string           // returns full text with \n line endings
func (m *Memo) SetText(s string)       // replaces all text, resets cursor to (0,0)
func (m *Memo) CursorPos() (row, col int)
func (m *Memo) Selection() (startRow, startCol, endRow, endCol int)
func (m *Memo) AutoIndent() bool
func (m *Memo) SetAutoIndent(enabled bool)
```

`SetText` splits the input on `\n` (and `\r\n`) into the internal `[][]rune` representation. `Text()` joins lines with `\n`.

### 3.12 Color Scheme

Two new entries:
- `MemoNormal` — normal text color
- `MemoSelected` — selected text color

These are separate from ListViewer entries because TMemo and ListViewer appear in different contexts with different default colors (in original TV, TMemo maps to dialog palette entries 26-27 while ListViewer maps to 26-29, and they resolve to different concrete colors in different palette levels).

---

## 4. THistory

A non-focusable companion view placed next to an InputLine, displaying a dropdown icon that opens a modal history list.

### 4.1 Architecture

THistory is a plain Widget (not Container, not focusable). It stores a reference to a linked InputLine and a numeric history ID. It uses `OfPostProcess` to intercept keyboard events that the focused InputLine does not consume.

### 4.2 Visual Appearance

Three characters wide, one row tall: `▐↓▌`

- `▐` (right half block, U+2590) and `▌` (left half block, U+258C): drawn in `HistorySides` color
- `↓` (down arrow, U+2193): drawn in `HistoryArrow` color

### 4.3 Constructor

```go
func NewHistory(bounds Rect, link *InputLine, historyID int) *History
```

Sets:
- `OfPostProcess` (sees events after focused view)
- Does NOT set `OfSelectable` (never receives focus)
- Event mask includes broadcasts

Typical positioning: immediately to the right of the InputLine, 3 chars wide, 1 row tall, same Y position.

### 4.4 Event Handling

#### Mouse Click

Any mouse click (`Button1`) on the History icon:
1. Attempt to focus the linked InputLine (`link.Owner().SetFocusedChild(link)`)
2. If the link is not selectable, clear the event and return
3. Otherwise, open the history dropdown (see 4.6)

#### Keyboard (Postprocess)

Down arrow key, only when the linked InputLine currently has focus (`link.HasState(SfSelected)`):
1. Open the history dropdown

All other keys: ignored (not consumed).

#### Broadcast

- `CmReleasedFocus` with `event.Info == link`: record the InputLine's current text to history
- `CmRecordHistory` (new command): record the InputLine's current text to history. This is broadcast by Button.press() so that all History views in a dialog record their linked InputLine contents when any button is pressed.

### 4.5 History Storage

A global `HistoryStore` manages history entries by ID.

```go
var DefaultHistory = NewHistoryStore(20)

type HistoryStore struct {
    maxPerID int
    entries  map[int][]string
}

func NewHistoryStore(maxPerID int) *HistoryStore
func (hs *HistoryStore) Add(id int, s string)
func (hs *HistoryStore) Entries(id int) []string
func (hs *HistoryStore) Clear()
```

**Add behavior:**
- Empty strings are never stored
- If the string equals the most recent entry for this ID, it is not added (deduplication)
- If a duplicate exists at any other position, that older entry is removed before adding the new one at the end
- If the entry count exceeds `maxPerID`, the oldest entry is evicted

**Entries** returns entries in chronological order (oldest first, newest last).

History IDs are integers. Multiple InputLines can share the same history ID (e.g., all filename inputs use ID 1). The ID is a convention, not enforced.

### 4.6 Dropdown Behavior

When triggered (by click or Down arrow):

1. **Record current InputLine text** to history via `DefaultHistory.Add(historyID, link.Text())`

2. **Calculate popup bounds** relative to the InputLine:
   - Width: InputLine width + 2 (1 char margin on each side for the frame)
   - Height: min(entry count + 2, 9) — at most 7 visible items plus frame
   - Position: overlapping the InputLine, extending downward
   - Clipped to the available screen area

3. **Create popup window**: A Window (no title, no number, close-only flags) containing a ListViewer with vertical scrollbar. The ListViewer shows history entries (most recent at top for user convenience — reversed from storage order).

4. **Execute modally**: The popup is inserted into the Desktop (not the current dialog) to avoid SubBuffer clipping. Coordinates are converted from dialog-local to desktop-absolute. `Desktop.ExecView(popup)` runs the modal loop.

5. **Handle result**:
   - `CmOK` (Enter or double-click): copy selected entry text into InputLine via `link.SetText(selectedText)`, select all text in InputLine, redraw InputLine
   - `CmCancel` (Escape, click outside, close button): InputLine unchanged

6. **Destroy popup window**.

### 4.7 Popup Interaction

The popup's ListViewer handles:
- Arrow keys, PgUp/PgDn for navigation
- Enter: confirm selection (`endModal(CmOK)`)
- Double-click: confirm selection
- Escape: cancel (`endModal(CmCancel)`)
- Click outside popup: cancel

Single click changes the focused item but does not close the popup.

### 4.8 CmRecordHistory Command

New command constant to add to the iota block in `tv/command.go`, after `CmSelectWindowNum`:

```go
CmRecordHistory
```

This must be broadcast by Button on press so all THistory views in the dialog record their linked InputLine contents when any button is pressed. This matches original TV where `TButton::press()` broadcasts `cmRecordHistory`.

### 4.9 Color Scheme

Two new entries:
- `HistoryArrow` — color for the `↓` character
- `HistorySides` — color for the `▐` and `▌` bracket characters

---

## 5. Color Scheme Summary

New entries to add to `theme.ColorScheme`:

| Entry | Widget | Purpose |
|-------|--------|---------|
| `MemoNormal` | TMemo | Normal text color |
| `MemoSelected` | TMemo | Selected text color |
| `HistoryArrow` | THistory | Down arrow icon color |
| `HistorySides` | THistory | Side bracket icon color |

All existing themes (BorlandBlue, BorlandCyan, BorlandGray, C64, Matrix) must be updated with entries for these new fields.

---

## 6. Demo App Updates

The demo app (`e2e/testapp/basic/main.go`) must be updated to exercise all new widgets:

### Window 1 "File Manager" changes:
- Add a TLabel linked to an existing control (e.g., `~F~ile type:` linked to the RadioButtons)
- Replace the manually-wired ListViewer+ScrollBar in Window 2 with a TListBox

### Window 2 "Editor" changes:
- Replace the ListViewer+ScrollBar with a TListBox
- Or: convert to a TMemo with sample text

### New Window 3 "Notes" (or add to existing dialog):
- TMemo with vertical scrollbar for multi-line text editing
- Demonstrates cursor movement, selection, scrolling

### Dialog changes:
- Add a THistory next to the InputLine in the "Open File" dialog
- Demonstrates dropdown history, entry recording, selection

### E2E test coverage:
Each new widget must have e2e tests that build the binary, launch in tmux, send keys, capture pane, and assert visible output:
- TLabel: verify shortcut focuses linked control, highlight changes on link focus
- TListBox: verify scrollbar appears, keyboard navigation works, selection fires
- TMemo: verify text entry, cursor movement, selection visible, scrolling
- THistory: verify icon visible, dropdown opens, selection applies to InputLine

---

## 7. Deferred Features (Future Phases)

These are explicitly out of scope for this spec:

- **Undo/redo** for TMemo (requires operation logging layer)
- **Find/replace** for TMemo
- **Insert/overwrite toggle** for TMemo
- **Multi-column ListViewer** (layout, divider rendering)
- **TEditor** (full-featured editor widget extending TMemo)
- **TFileDialog, TColorDialog** (composite dialogs using these widgets)
- **Word wrap** for TMemo (horizontal scrolling only, matching original TEditor)
