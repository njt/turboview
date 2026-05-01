# TMemo Advanced Features Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add selection model, clipboard, word operations, mouse handling, scrollbar integration, and tab rendering to the existing TMemo widget.

**Architecture:** The existing `Memo` struct gains selection state (`selStartRow/Col`, `selEndRow/Col`), two optional `*ScrollBar` pointers, and a `dragging` flag for mouse state. All editing operations become selection-aware (replace selection when one exists). Draw uses `MemoSelected` style for characters inside the selection range. Word boundaries follow original TEditor's `getCharType` classification.

**Tech Stack:** Go, tcell/v2, existing `tv` package types and patterns

---

## File Structure

| File | Responsibility |
|------|---------------|
| `tv/memo.go` | Modify: add selection fields, word movement, clipboard, mouse handling, scrollbar integration, tab rendering, selection-aware editing |
| `tv/memo_selection_test.go` | Create: tests for selection model, Shift+movement, Ctrl+A, selection-aware editing |
| `tv/memo_clipboard_test.go` | Create: tests for Ctrl+C/X/V, clipboard sharing with InputLine |
| `tv/memo_word_test.go` | Create: tests for Ctrl+Left/Right, Ctrl+Backspace/Delete, word boundary logic |
| `tv/memo_mouse_test.go` | Create: tests for click, drag, double-click, triple-click |
| `tv/memo_scrollbar_test.go` | Create: tests for SetHScrollBar/SetVScrollBar, sync, OnChange callbacks |
| `tv/memo_tab_test.go` | Create: tests for tab rendering in Draw |
| `tv/integration_batch2_memo_advanced_test.go` | Create: integration tests for advanced memo features |
| `e2e/testapp/basic/main.go` | Modify: add scrollbars to memo, demo selection |
| `e2e/e2e_test.go` | Modify: add e2e tests for memo advanced features |

---

## Phase 3 Tasks

- [ ] Task 1: Selection Model and Shift+Movement
- [ ] Task 2: Selection-Aware Draw and Tab Rendering
- [ ] Task 3: Clipboard Operations and Selection-Aware Editing
- [ ] Task 4: Word Movement and Word Deletion
- [ ] Task 5: Mouse Handling (Click, Drag, Double-Click, Triple-Click)
- [ ] Task 6: Scrollbar Integration
- [ ] Task 7: Integration Checkpoint — Advanced Memo Features
- [ ] Task 8: E2E Test — Advanced Memo in Demo App

---

### Task 1: Selection Model and Shift+Movement

**Files:**
- Modify: `tv/memo.go`
- Test: `tv/memo_selection_test.go`

**Requirements:**
- Memo struct gains four new fields: `selStartRow`, `selStartCol`, `selEndRow`, `selEndCol` (all int, default 0)
- `Selection() (startRow, startCol, endRow, endCol int)` returns the current selection range
- `HasSelection() bool` returns true when selection start != selection end
- `normalizedSelection()` returns selection endpoints in document order: (loRow, loCol, hiRow, hiCol) — the earlier position is always first regardless of selection direction
- `clearSelection()` sets all four fields to `cursorRow, cursorCol` (collapses to cursor)
- `setSelectionEnd()` sets `selEndRow, selEndCol` to `cursorRow, cursorCol`
- When no Shift is held, all cursor movement keys (Left, Right, Up, Down, Home, End, Ctrl+Home, Ctrl+End, PgUp, PgDn) collapse the selection before moving
- When Shift is held with any cursor movement key, the selection extends: the anchor (`selStart`) stays fixed, the extent (`selEnd`) follows the cursor
- Starting a Shift+movement from a collapsed selection sets the anchor at the current cursor position, then moves the cursor and updates the extent
- `Ctrl+A` selects all text: anchor at (0,0), cursor at end of last line, extent follows cursor
- After Ctrl+A, the cursor is at the end of the document (last row, last col)
- Shift+Left at position (0,0) does nothing (no negative selection)
- Shift+Right at end of document does nothing
- Shift+Up from row 0 moves cursor to col 0, extending selection
- Shift+Down from last row moves cursor to end of line, extending selection
- Shift+Home applies smart home logic while extending selection
- Shift+End extends selection to end of current line
- Shift+Ctrl+Home extends selection to document start
- Shift+Ctrl+End extends selection to document end
- Shift+PgUp extends selection up by page
- Shift+PgDn extends selection down by page
- After any non-Shift cursor key, `HasSelection()` returns false

**Implementation:**

Add selection fields to the Memo struct:

```go
type Memo struct {
    BaseView
    lines       [][]rune
    cursorRow   int
    cursorCol   int
    deltaX      int
    deltaY      int
    autoIndent  bool
    selStartRow int
    selStartCol int
    selEndRow   int
    selEndCol   int
}
```

Add selection methods:

```go
func (m *Memo) Selection() (int, int, int, int) {
    return m.selStartRow, m.selStartCol, m.selEndRow, m.selEndCol
}

func (m *Memo) HasSelection() bool {
    return m.selStartRow != m.selEndRow || m.selStartCol != m.selEndCol
}

func (m *Memo) normalizedSelection() (int, int, int, int) {
    sr, sc, er, ec := m.selStartRow, m.selStartCol, m.selEndRow, m.selEndCol
    if sr > er || (sr == er && sc > ec) {
        sr, sc, er, ec = er, ec, sr, sc
    }
    return sr, sc, er, ec
}

func (m *Memo) clearSelection() {
    m.selStartRow = m.cursorRow
    m.selStartCol = m.cursorCol
    m.selEndRow = m.cursorRow
    m.selEndCol = m.cursorCol
}

func (m *Memo) setSelectionEnd() {
    m.selEndRow = m.cursorRow
    m.selEndCol = m.cursorCol
}
```

Modify `HandleEvent` to detect Shift modifier and branch between selection-extending and selection-collapsing movement:

```go
case tcell.KeyLeft:
    if k.Modifiers&tcell.ModCtrl != 0 {
        // handled in Task 4
    }
    shift := k.Modifiers&tcell.ModShift != 0
    if shift {
        if !m.HasSelection() {
            m.selStartRow = m.cursorRow
            m.selStartCol = m.cursorCol
        }
        m.cursorLeft()
        m.setSelectionEnd()
    } else {
        m.cursorLeft()
        m.clearSelection()
    }
    event.Clear()
```

Apply the same shift/non-shift pattern to all cursor movement keys: Right, Up, Down, Home, End, Ctrl+Home, Ctrl+End, PgUp, PgDn.

Add Ctrl+A handler:

```go
case tcell.KeyCtrlA:
    m.selStartRow = 0
    m.selStartCol = 0
    m.cursorRow = len(m.lines) - 1
    m.cursorCol = len(m.lines[m.cursorRow])
    m.setSelectionEnd()
    m.ensureCursorVisible()
    event.Clear()
```

Update `SetText` to reset all four selection fields: `m.selStartRow = 0; m.selStartCol = 0; m.selEndRow = 0; m.selEndCol = 0` alongside the existing cursor/delta resets.

**Run tests:** `go test ./tv/ -run TestMemoSelection -v`

**Commit:** `git commit -m "feat(memo): add selection model with Shift+movement and Ctrl+A"`

---

### Task 2: Selection-Aware Draw and Tab Rendering

**Files:**
- Modify: `tv/memo.go` (Draw method)
- Test: `tv/memo_tab_test.go`
- Test: `tv/memo_selection_test.go` (additional draw tests in separate file or in selection test file)

**Requirements:**

**Selection-aware drawing:**
- When a selection exists, characters within the selection range use `MemoSelected` style
- Characters outside the selection use `MemoNormal` style
- Selection rendering works across multiple lines: the selection on intermediate lines covers the full line content plus trailing spaces to the widget edge (representing the selected newline)
- On the first line of the selection, only characters from `selStartCol` to end of line are selected
- On the last line of the selection, only characters from start of line to `selEndCol` are selected
- If selection start and end are on the same line, only characters between the two columns are selected
- Trailing spaces after line content on a selected intermediate line use `MemoSelected` style (the newline is "selected")
- Trailing spaces after line content on the last selection line use `MemoNormal` style (cursor is at the extent, not past it)
- When `selStartRow == selEndRow` but `selStartCol > selEndCol` (backward selection on same line), the smaller column is treated as the visual start
- Selection rendering uses `normalizedSelection()` to determine display order

**Tab rendering:**
- Tab characters (`\t`) expand visually to the next multiple of 8 columns
- A tab at visual column 0 expands to 8 spaces; at column 1, expands to 7 spaces; at column 7, expands to 1 space; at column 8, expands to 8 spaces
- Tab stops are at visual columns 0, 8, 16, 24, etc.
- **`deltaX` remains rune-based** (consistent with `cursorCol`, `ensureCursorVisible()`, and `clampCursor()`). Tab expansion is display-only: Draw walks runes from `deltaX` onward and tracks a visual column counter to compute tab widths, but `deltaX` itself is a rune offset into the line, not a visual column offset
- Cursor column still counts in runes, not visual columns — tab rendering is display-only
- Each space of a tab expansion uses the same style (normal or selected) as the tab character itself
- Tabs that extend past the right edge of the widget are clipped

**Implementation:**

Refactor Draw to handle selection highlighting and tab expansion. `deltaX` is rune-based (same as cursorCol), so the rune loop starts at `deltaX`. A separate `vcol` counter tracks the visual column for computing tab widths. Runes before `deltaX` are skipped but still walked to build up `vcol` for correct tab alignment:

```go
func (m *Memo) Draw(buf *DrawBuffer) {
    cs := m.ColorScheme()
    normalStyle := tcell.StyleDefault
    selectedStyle := tcell.StyleDefault
    if cs != nil {
        normalStyle = cs.MemoNormal
        selectedStyle = cs.MemoSelected
    }

    h := m.Bounds().Height()
    w := m.Bounds().Width()

    sr, sc, er, ec := m.normalizedSelection()

    for y := 0; y < h; y++ {
        lineIdx := m.deltaY + y
        if lineIdx >= len(m.lines) {
            buf.Fill(NewRect(0, y, w, 1), ' ', normalStyle)
            continue
        }
        line := m.lines[lineIdx]

        // Walk runes before deltaX to compute correct visual column for tab alignment
        vcol := 0  // visual column (accounts for tabs)
        for runeIdx := 0; runeIdx < m.deltaX && runeIdx < len(line); runeIdx++ {
            if line[runeIdx] == '\t' {
                vcol += 8 - (vcol % 8)
            } else {
                vcol++
            }
        }

        // Render visible runes starting from deltaX
        x := 0  // screen column
        for runeIdx := m.deltaX; runeIdx < len(line) && x < w; runeIdx++ {
            ch := line[runeIdx]
            inSel := m.posInSelection(lineIdx, runeIdx, sr, sc, er, ec)

            style := normalStyle
            if inSel {
                style = selectedStyle
            }

            if ch == '\t' {
                tabWidth := 8 - (vcol % 8)
                for i := 0; i < tabWidth && x < w; i++ {
                    buf.WriteChar(x, y, ' ', style)
                    x++
                }
                vcol += tabWidth
            } else {
                buf.WriteChar(x, y, ch, style)
                x++
                vcol++
            }
        }

        // trailing fill
        trailSelected := m.trailingSelected(lineIdx, len(line), sr, sc, er, ec)
        trailStyle := normalStyle
        if trailSelected {
            trailStyle = selectedStyle
        }
        for ; x < w; x++ {
            buf.WriteChar(x, y, ' ', trailStyle)
        }
    }
}
```

Helper to check if a rune position is in the selection:

```go
func (m *Memo) posInSelection(row, col, sr, sc, er, ec int) bool {
    if sr == er && sc == ec {
        return false // no selection
    }
    if row < sr || row > er {
        return false
    }
    if row == sr && row == er {
        return col >= sc && col < ec
    }
    if row == sr {
        return col >= sc
    }
    if row == er {
        return col < ec
    }
    return true // intermediate line
}
```

Helper for trailing space style:

```go
func (m *Memo) trailingSelected(row, lineLen, sr, sc, er, ec int) bool {
    if sr == er && sc == ec {
        return false
    }
    if row < sr || row >= er {
        return false
    }
    if row == sr {
        return lineLen >= sc // trailing of first sel line only if sel starts before/at end
    }
    return true // intermediate line: trailing is selected
}
```

**Run tests:** `go test ./tv/ -run "TestMemoTab|TestMemoDraw.*Selection" -v`

**Commit:** `git commit -m "feat(memo): selection-aware Draw with tab rendering"`

---

### Task 3: Clipboard Operations and Selection-Aware Editing

**Files:**
- Modify: `tv/memo.go`
- Test: `tv/memo_clipboard_test.go`

**Requirements:**

**Clipboard operations:**
- `Ctrl+C` copies the selected text to the package-level `clipboard` variable (same one used by InputLine in `tv/input_line.go`)
- `Ctrl+C` with no selection does nothing (does not clear clipboard)
- `Ctrl+X` cuts: copies selection to clipboard, then deletes the selection
- `Ctrl+X` with no selection does nothing
- `Ctrl+V` pastes: if a selection exists, it is replaced by the clipboard content; otherwise clipboard content is inserted at cursor
- `Ctrl+V` with empty clipboard does nothing
- Pasted text may contain newlines — these are split into multiple lines in the buffer
- After paste, cursor is at the end of the pasted text
- After paste, selection is cleared (collapsed at cursor)
- `Ctrl+A` selects all text (already implemented in Task 1)

**Selection-aware editing:**
- Typing a printable character while a selection exists replaces the selection with the typed character
- Enter while a selection exists replaces the selection with a newline (with auto-indent from the line where the selection started)
- Backspace while a selection exists deletes the selection (does not delete an additional character)
- Delete while a selection exists deletes the selection (does not delete an additional character)
- After deleting a selection, cursor is at the start (earlier position) of the former selection
- Ctrl+Y (delete line) is NOT selection-aware — it always deletes the current line regardless of selection, and clears the selection

**Implementation:**

Add `deleteSelection()`:

```go
func (m *Memo) deleteSelection() {
    if !m.HasSelection() {
        return
    }
    sr, sc, er, ec := m.normalizedSelection()

    startLine := m.lines[sr]
    endLine := m.lines[er]

    merged := make([]rune, sc+len(endLine)-ec)
    copy(merged, startLine[:sc])
    copy(merged[sc:], endLine[ec:])

    m.lines[sr] = merged
    if er > sr {
        m.lines = append(m.lines[:sr+1], m.lines[er+1:]...)
    }

    m.cursorRow = sr
    m.cursorCol = sc
    m.clearSelection()
}
```

Add `selectedText()`:

```go
func (m *Memo) selectedText() string {
    if !m.HasSelection() {
        return ""
    }
    sr, sc, er, ec := m.normalizedSelection()
    if sr == er {
        return string(m.lines[sr][sc:ec])
    }
    var sb strings.Builder
    sb.WriteString(string(m.lines[sr][sc:]))
    for i := sr + 1; i < er; i++ {
        sb.WriteByte('\n')
        sb.WriteString(string(m.lines[i]))
    }
    sb.WriteByte('\n')
    sb.WriteString(string(m.lines[er][:ec]))
    return sb.String()
}
```

Add clipboard handlers in HandleEvent:

```go
case tcell.KeyCtrlC:
    if m.HasSelection() {
        clipboard = m.selectedText()
    }
    event.Clear()
case tcell.KeyCtrlX:
    if m.HasSelection() {
        clipboard = m.selectedText()
        m.deleteSelection()
    }
    event.Clear()
case tcell.KeyCtrlV:
    if clipboard != "" {
        m.deleteSelection()
        m.insertText(clipboard)
    }
    event.Clear()
```

Add `insertText(s string)` that splits on newlines and inserts into the buffer at cursor:

```go
func (m *Memo) insertText(s string) {
    s = strings.ReplaceAll(s, "\r\n", "\n")
    parts := strings.Split(s, "\n")
    if len(parts) == 1 {
        line := m.lines[m.cursorRow]
        runes := []rune(parts[0])
        newLine := make([]rune, len(line)+len(runes))
        copy(newLine, line[:m.cursorCol])
        copy(newLine[m.cursorCol:], runes)
        copy(newLine[m.cursorCol+len(runes):], line[m.cursorCol:])
        m.lines[m.cursorRow] = newLine
        m.cursorCol += len(runes)
    } else {
        line := m.lines[m.cursorRow]
        before := line[:m.cursorCol]
        after := line[m.cursorCol:]

        firstLine := make([]rune, len(before)+len([]rune(parts[0])))
        copy(firstLine, before)
        copy(firstLine[len(before):], []rune(parts[0]))

        lastPart := []rune(parts[len(parts)-1])
        lastLine := make([]rune, len(lastPart)+len(after))
        copy(lastLine, lastPart)
        copy(lastLine[len(lastPart):], after)

        newLines := make([][]rune, 0, len(m.lines)+len(parts)-1)
        newLines = append(newLines, m.lines[:m.cursorRow]...)
        newLines = append(newLines, firstLine)
        for i := 1; i < len(parts)-1; i++ {
            newLines = append(newLines, []rune(parts[i]))
        }
        newLines = append(newLines, lastLine)
        newLines = append(newLines, m.lines[m.cursorRow+1:]...)
        m.lines = newLines

        m.cursorRow += len(parts) - 1
        m.cursorCol = len(lastPart)
    }
    m.clearSelection()
    m.ensureCursorVisible()
}
```

Modify `insertChar`, `insertNewline`, `backspace`, `deleteChar` to call `deleteSelection()` first when a selection exists. For backspace/delete, if a selection existed, skip the normal single-char operation (just delete the selection).

**Run tests:** `go test ./tv/ -run TestMemoClipboard -v`

**Commit:** `git commit -m "feat(memo): clipboard operations and selection-aware editing"`

---

### Task 4: Word Movement and Word Deletion

**Files:**
- Modify: `tv/memo.go`
- Test: `tv/memo_word_test.go`

**Requirements:**

**Word boundaries:**
- Character classification follows TEditor's `getCharType`: whitespace (space, tab), punctuation (`!"#$%&'()*+,-./:;<=>?@[\]^{|}~`), and word characters (everything else including letters, digits, underscore)
- A word boundary is a transition between different character classes
- Line breaks are their own boundary — moving across a line break always stops

**Ctrl+Left (previous word):**
- If at column 0 and row > 0: move to end of previous line (stop there)
- Otherwise: skip backward over characters of the current class, then skip backward over whitespace. The cursor lands at the start of the previous word.
- At document start (0,0): no movement
- If the character immediately before the cursor is whitespace, skip the whitespace first, then skip over the word/punctuation class

**Ctrl+Right (next word):**
- If at end of line and not on last row: move to start of next line (stop there)
- Otherwise: skip forward over characters of the current class, then skip forward over whitespace. The cursor lands at the start of the next word.
- At document end: no movement
- If the character at the cursor is whitespace, skip the whitespace first, then stop at the start of the next word

**Ctrl+Backspace (delete previous word):**
- Deletes from cursor position back to where Ctrl+Left would have moved the cursor
- If a selection exists, deletes the selection instead (same as regular Backspace with selection)

**Ctrl+Delete (delete next word):**
- Deletes from cursor position forward to where Ctrl+Right would have moved the cursor
- If a selection exists, deletes the selection instead (same as regular Delete with selection)

**Shift+Ctrl+Left and Shift+Ctrl+Right:**
- Extend the selection by word (same word movement logic, but extends selection instead of collapsing)

**Implementation:**

Add character classification:

```go
func charClass(r rune) int {
    if r == ' ' || r == '\t' {
        return 0 // whitespace
    }
    if strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^`{|}~", r) {
        return 1 // punctuation
    }
    return 2 // word character
}
```

Add word movement methods:

```go
func (m *Memo) wordLeft() {
    if m.cursorCol == 0 {
        if m.cursorRow > 0 {
            m.cursorRow--
            m.cursorCol = len(m.lines[m.cursorRow])
        }
        return
    }
    line := m.lines[m.cursorRow]
    col := m.cursorCol
    // skip whitespace
    for col > 0 && charClass(line[col-1]) == 0 {
        col--
    }
    if col == 0 {
        m.cursorCol = 0
        return
    }
    // skip current class
    cls := charClass(line[col-1])
    for col > 0 && charClass(line[col-1]) == cls {
        col--
    }
    m.cursorCol = col
}

func (m *Memo) wordRight() {
    line := m.lines[m.cursorRow]
    if m.cursorCol >= len(line) {
        if m.cursorRow < len(m.lines)-1 {
            m.cursorRow++
            m.cursorCol = 0
        }
        return
    }
    col := m.cursorCol
    // skip current class
    cls := charClass(line[col])
    for col < len(line) && charClass(line[col]) == cls {
        col++
    }
    // skip whitespace
    for col < len(line) && charClass(line[col]) == 0 {
        col++
    }
    m.cursorCol = col
}
```

Update HandleEvent for Ctrl+Left, Ctrl+Right (remove the existing `return` stubs), Ctrl+Backspace, Ctrl+Delete:

```go
case tcell.KeyLeft:
    shift := k.Modifiers&tcell.ModShift != 0
    if k.Modifiers&tcell.ModCtrl != 0 {
        if shift && !m.HasSelection() {
            m.selStartRow, m.selStartCol = m.cursorRow, m.cursorCol
        }
        m.wordLeft()
        if shift {
            m.setSelectionEnd()
        } else {
            m.clearSelection()
        }
    } else {
        // existing shift/non-shift cursorLeft logic
    }
    m.ensureCursorVisible()
    event.Clear()
```

For Ctrl+Backspace: find the target position via `wordLeft`, then delete the range. For Ctrl+Delete: find target via `wordRight`, then delete the range.

```go
func (m *Memo) deleteWordLeft() {
    if m.HasSelection() {
        m.deleteSelection()
        return
    }
    endRow, endCol := m.cursorRow, m.cursorCol
    m.wordLeft()
    if m.cursorRow == endRow && m.cursorCol == endCol {
        return // no movement, nothing to delete
    }
    // delete from (cursorRow, cursorCol) to (endRow, endCol)
    m.selStartRow, m.selStartCol = m.cursorRow, m.cursorCol
    m.selEndRow, m.selEndCol = endRow, endCol
    m.deleteSelection()
}

func (m *Memo) deleteWordRight() {
    if m.HasSelection() {
        m.deleteSelection()
        return
    }
    startRow, startCol := m.cursorRow, m.cursorCol
    m.wordRight()
    if m.cursorRow == startRow && m.cursorCol == startCol {
        return
    }
    m.selStartRow, m.selStartCol = startRow, startCol
    m.selEndRow, m.selEndCol = m.cursorRow, m.cursorCol
    m.cursorRow, m.cursorCol = startRow, startCol
    m.deleteSelection()
}
```

**Run tests:** `go test ./tv/ -run TestMemoWord -v`

**Commit:** `git commit -m "feat(memo): word movement (Ctrl+Left/Right) and word deletion (Ctrl+Backspace/Delete)"`

---

### Task 5: Mouse Handling (Click, Drag, Double-Click, Triple-Click)

**Files:**
- Modify: `tv/memo.go`
- Test: `tv/memo_mouse_test.go`

**Requirements:**

**Click (ClickCount == 1, Button1):**
- Positions cursor at clicked location: `row = deltaY + mouseY`, `col = deltaX + mouseX`
- Row is clamped to `[0, len(lines)-1]`
- Col is clamped to `[0, len(lines[row])]` (can be at end of line but not past)
- Collapses any existing selection
- Tab characters: a click in the visual span of a tab lands at the tab character's rune index (not the visual column)

**Drag (Button1 with motion, after initial click):**
- Sets `dragging` flag to true on button press
- On motion events while dragging: moves cursor to mouse position and extends selection
- The anchor point is the initial click position; the extent follows the mouse
- When mouse moves outside the widget bounds during drag: clamp cursor to valid range via `mouseToPos`, then `ensureCursorVisible()` scrolls the viewport. Continuous auto-scroll (using `startMouseAuto`/`stopMouseAuto`) is deferred — not implemented in this phase.
- Dragging stops when Button1 is released (ButtonNone)

**Double-click (ClickCount == 2):**
- Selects the word under the cursor (using the same word boundary logic from Task 4)
- If the cursor is on whitespace, selects the whitespace run
- If the cursor is at end of line, selects nothing (or the last word if cursor is immediately after it)

**Triple-click (ClickCount == 3):**
- Selects the entire line (from column 0 to end of line)
- Selection anchor at (row, 0), selection extent at (row, len(line))

**Mouse events that are NOT Button1 are ignored (not consumed).**
**Mouse events are consumed (event.Clear()) when handled.**

**Implementation:**

Add `dragging` field and `dragAnchorRow`, `dragAnchorCol` to Memo struct:

```go
dragging      bool
dragAnchorRow int
dragAnchorCol int
```

Add mouse-to-buffer position conversion. Since `deltaX` is rune-based, we must walk runes from line start, tracking visual columns (accounting for tab expansion), to find which rune index corresponds to the clicked screen X:

```go
func (m *Memo) mouseToPos(mx, my int) (int, int) {
    row := m.deltaY + my
    if row < 0 { row = 0 }
    if row >= len(m.lines) { row = len(m.lines) - 1 }

    line := m.lines[row]

    // Compute visual column at deltaX (the left edge of the viewport)
    vcolAtDeltaX := 0
    for i := 0; i < m.deltaX && i < len(line); i++ {
        if line[i] == '\t' {
            vcolAtDeltaX += 8 - (vcolAtDeltaX % 8)
        } else {
            vcolAtDeltaX++
        }
    }

    targetVcol := vcolAtDeltaX + mx

    // Walk from start to find the rune at targetVcol
    vcol := 0
    runeIdx := 0
    for runeIdx < len(line) && vcol < targetVcol {
        if line[runeIdx] == '\t' {
            tw := 8 - (vcol % 8)
            if vcol+tw > targetVcol {
                break // inside tab span — land on the tab rune
            }
            vcol += tw
        } else {
            vcol++
        }
        runeIdx++
    }

    if runeIdx > len(line) { runeIdx = len(line) }
    return row, runeIdx
}
```

Add mouse handling in HandleEvent (before keyboard handling). Motion events during drag arrive with Button1 still held — they are distinguished from the initial press by the `dragging` flag:

```go
if event.What == EvMouse && event.Mouse != nil {
    me := event.Mouse
    if me.Button&tcell.Button1 != 0 {
        if m.dragging {
            // Continued drag (motion with button held)
            row, col := m.mouseToPos(me.X, me.Y)
            m.cursorRow = row
            m.cursorCol = col
            m.selStartRow = m.dragAnchorRow
            m.selStartCol = m.dragAnchorCol
            m.setSelectionEnd()
            m.ensureCursorVisible()
            event.Clear()
        } else {
            // Initial press
            switch me.ClickCount {
            case 1:
                row, col := m.mouseToPos(me.X, me.Y)
                m.cursorRow = row
                m.cursorCol = col
                m.clearSelection()
                m.dragging = true
                m.dragAnchorRow = row
                m.dragAnchorCol = col
                m.ensureCursorVisible()
                event.Clear()
            case 2:
                row, col := m.mouseToPos(me.X, me.Y)
                m.cursorRow = row
                m.cursorCol = col
                m.selectWordAtCursor()
                event.Clear()
            case 3:
                row, col := m.mouseToPos(me.X, me.Y)
                m.cursorRow = row
                m.cursorCol = col
                m.selectLineAtCursor()
                event.Clear()
            }
        }
    } else if me.Button == tcell.ButtonNone && m.dragging {
        // Button released — stop dragging
        m.dragging = false
        event.Clear()
    }
    return
}
```

Word/line selection helpers:

```go
func (m *Memo) selectWordAtCursor() {
    line := m.lines[m.cursorRow]
    if len(line) == 0 || m.cursorCol >= len(line) {
        return
    }
    cls := charClass(line[m.cursorCol])
    start := m.cursorCol
    for start > 0 && charClass(line[start-1]) == cls {
        start--
    }
    end := m.cursorCol
    for end < len(line) && charClass(line[end]) == cls {
        end++
    }
    m.selStartRow = m.cursorRow
    m.selStartCol = start
    m.cursorCol = end
    m.setSelectionEnd()
    m.ensureCursorVisible()
}

func (m *Memo) selectLineAtCursor() {
    m.selStartRow = m.cursorRow
    m.selStartCol = 0
    m.cursorCol = len(m.lines[m.cursorRow])
    m.setSelectionEnd()
    m.ensureCursorVisible()
}
```

**Run tests:** `go test ./tv/ -run TestMemoMouse -v`

**Commit:** `git commit -m "feat(memo): mouse handling (click, drag, double-click, triple-click)"`

---

### Task 6: Scrollbar Integration

**Files:**
- Modify: `tv/memo.go`
- Test: `tv/memo_scrollbar_test.go`

**Requirements:**

**Linking scrollbars:**
- `SetVScrollBar(sb *ScrollBar)` links a vertical scrollbar
- `SetHScrollBar(sb *ScrollBar)` links a horizontal scrollbar
- If a scrollbar was previously linked, its `OnChange` callback is cleared before linking the new one
- Passing `nil` unlinks the scrollbar and clears its callback
- `WithScrollBars(h, v *ScrollBar) MemoOption` constructor option calls both setters

**Sync from Memo to scrollbar:**
- After any operation that changes cursor position, text content, or viewport offset, call `syncScrollBars()`
- Vertical scrollbar: `SetRange(0, len(lines)-1)`, `SetPageSize(height-1)`, `SetValue(deltaY)`
- Horizontal scrollbar: `SetRange(0, maxLineWidth)`, `SetPageSize(width/2)`, `SetValue(deltaX)`
- `maxLineWidth` is the length of the longest line in runes

**Sync from scrollbar to Memo:**
- Vertical scrollbar `OnChange` callback sets `deltaY` to the scrollbar's value
- Horizontal scrollbar `OnChange` callback sets `deltaX` to the scrollbar's value

**State-dependent visibility:**
- Override `SetState` on Memo
- When `SfSelected` changes: if gaining focus, show linked scrollbars (`SetState(SfVisible, true)`); if losing focus, hide them (`SetState(SfVisible, false)`)
- This matches original TEditor behavior where scrollbars only appear when the editor is active

**Implementation:**

Add scrollbar fields:

```go
hScrollBar *ScrollBar
vScrollBar *ScrollBar
```

Add setters:

```go
func (m *Memo) SetVScrollBar(sb *ScrollBar) {
    if m.vScrollBar != nil {
        m.vScrollBar.OnChange = nil
    }
    m.vScrollBar = sb
    if sb != nil {
        sb.OnChange = func(val int) {
            m.deltaY = val
        }
        m.syncScrollBars()
    }
}

func (m *Memo) SetHScrollBar(sb *ScrollBar) {
    if m.hScrollBar != nil {
        m.hScrollBar.OnChange = nil
    }
    m.hScrollBar = sb
    if sb != nil {
        sb.OnChange = func(val int) {
            m.deltaX = val
        }
        m.syncScrollBars()
    }
}
```

Add WithScrollBars option:

```go
func WithScrollBars(h, v *ScrollBar) MemoOption {
    return func(m *Memo) {
        m.SetHScrollBar(h)
        m.SetVScrollBar(v)
    }
}
```

Add sync method:

```go
func (m *Memo) syncScrollBars() {
    if m.vScrollBar != nil {
        m.vScrollBar.SetRange(0, len(m.lines)-1)
        m.vScrollBar.SetPageSize(m.Bounds().Height() - 1)
        m.vScrollBar.SetValue(m.deltaY)
    }
    if m.hScrollBar != nil {
        maxWidth := 0
        for _, line := range m.lines {
            if len(line) > maxWidth {
                maxWidth = len(line)
            }
        }
        m.hScrollBar.SetRange(0, maxWidth)
        m.hScrollBar.SetPageSize(m.Bounds().Width() / 2)
        m.hScrollBar.SetValue(m.deltaX)
    }
}
```

Call `m.syncScrollBars()` at the end of every method that modifies cursor/text/viewport: `cursorLeft`, `cursorRight`, `cursorUp`, `cursorDown`, `smartHome`, `pageUp`, `pageDown`, `insertChar`, `insertNewline`, `backspace`, `deleteChar`, `deleteLine`, `insertText`, `deleteSelection`, `deleteWordLeft`, `deleteWordRight`, `wordLeft`, `wordRight`, `SetText`, and the mouse handlers. **Do NOT call it inside `ensureCursorVisible()`** — `syncScrollBars` calls `SetValue` which triggers `OnChange` which writes back to `deltaX`/`deltaY`, creating a re-entrant modification during viewport adjustment. Instead, add explicit `m.syncScrollBars()` calls after `m.ensureCursorVisible()` in each method.

Override `SetState`:

```go
func (m *Memo) SetState(flag ViewState, on bool) {
    m.BaseView.SetState(flag, on)
    if flag&SfSelected != 0 {
        if m.vScrollBar != nil {
            m.vScrollBar.SetState(SfVisible, on)
        }
        if m.hScrollBar != nil {
            m.hScrollBar.SetState(SfVisible, on)
        }
    }
}
```

**Run tests:** `go test ./tv/ -run TestMemoScrollbar -v`

**Commit:** `git commit -m "feat(memo): scrollbar integration with SetHScrollBar/SetVScrollBar"`

---

### Task 7: Integration Checkpoint — Advanced Memo Features

**Purpose:** Verify that Tasks 1-6 work together: selection + draw, clipboard + editing, word ops + selection, mouse + selection, scrollbar + viewport.

**Files:**
- Create: `tv/integration_batch2_memo_advanced_test.go`

**Requirements (for test writer):**
- Shift+Right creates a selection, then Draw renders those characters in MemoSelected style
- Ctrl+A selects all, then Ctrl+C copies the full text to clipboard, then creating a new Memo and Ctrl+V pastes it
- Double-click selects a word, then Ctrl+X cuts it, verifying both clipboard content and buffer state
- Ctrl+Left followed by Shift+Ctrl+Right selects a word via keyboard
- Typing a character while a selection exists replaces the selection, and the resulting text is correct
- A vertical scrollbar linked to Memo updates its value after PgDn
- Mouse click positions cursor correctly, then Shift+Right extends selection from that position
- Tab characters render as expanded spaces in Draw output
- After mouse drag selection, Ctrl+C copies the correct text

**Components to wire up:** Memo (real), ScrollBar (real), DrawBuffer (real)

**Run:** `go test ./tv/ -run TestIntegration -v`

**Commit:** `git commit -m "test: integration tests for TMemo advanced features"`

---

### Task 8: E2E Test — Advanced Memo in Demo App

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**
- Add a vertical scrollbar to the Notes window's Memo in the demo app
- The Memo should have enough text to exercise scrolling (at least 15 lines)
- E2E test: focus the Notes window (Alt+3), type text, verify it appears on screen
- E2E test: type enough text to cause scrolling, verify viewport scrolls (first line no longer visible)
- E2E test: verify the vertical scrollbar is visible when Notes window is focused
- Existing e2e tests continue to pass

**Implementation:**

Update demo app to add scrollbar:

```go
win3 := tv.NewWindow(tv.NewRect(45, 1, 30, 12), "Notes", tv.WithWindowNumber(3))
vScroll := tv.NewScrollBar(tv.NewRect(28, 0, 1, 10), tv.Vertical)
memo := tv.NewMemo(tv.NewRect(0, 0, 27, 10), tv.WithScrollBars(nil, vScroll))
memo.SetText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\nLine 11\nLine 12\nLine 13\nLine 14\nLine 15")
win3.Insert(memo)
win3.Insert(vScroll)
```

Add e2e tests:

```go
func TestMemoTypingAdvanced(t *testing.T) {
    // Focus Notes (Alt+3), type characters, capture screen, verify text appears
}

func TestMemoScrollbar(t *testing.T) {
    // Focus Notes (Alt+3), verify scrollbar visible, PgDn, verify scroll
}
```

**Run tests:** `go test ./e2e/ -v`

**Commit:** `git commit -m "feat: add scrollbar to demo Memo, e2e tests for advanced memo features"`
