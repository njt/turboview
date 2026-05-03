# Batch 5: Multi-column Clusters + Multi-column ListViewer

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add dynamic multi-column layout to CheckBoxes/RadioButtons and multi-column display to ListViewer, matching original Turbo Vision cluster layout behavior.

**Architecture:** CheckBoxes and RadioButtons gain a `relayoutItems()` method that computes column positions from `bounds.Height()`. Items fill top-to-bottom within a column, then overflow right. ListViewer gains a `numCols` parameter that splits the visible area into equal-width columns. Both add per-item enable/disable support. A new `ClusterDisabled` color scheme entry across all 5 themes supports dimmed rendering.

**Tech Stack:** Go, tcell/v2, existing tv/theme packages

---

## File Structure

| File | Responsibility | Action |
|------|---------------|--------|
| `theme/scheme.go` | ColorScheme struct | Add `ClusterDisabled` field |
| `theme/borland.go` | BorlandBlue theme | Add `ClusterDisabled` value |
| `theme/borland_cyan.go` | BorlandCyan theme | Add `ClusterDisabled` value |
| `theme/borland_gray.go` | BorlandGray theme | Add `ClusterDisabled` value |
| `theme/matrix.go` | Matrix theme | Add `ClusterDisabled` value |
| `theme/c64.go` | C64 theme | Add `ClusterDisabled` value |
| `tv/checkbox.go` | CheckBox + CheckBoxes | Multi-column layout, enable/disable, Left/Right nav |
| `tv/radio.go` | RadioButton + RadioButtons | Multi-column layout, enable/disable, Left/Right nav |
| `tv/list_viewer.go` | ListViewer | Multi-column display, column navigation |
| `e2e/testapp/basic/main.go` | Demo app | Update to show multi-column clusters |
| `e2e/e2e_test.go` | E2e tests | New tests for multi-column rendering |

---

## Tasks

- [ ] Task 1: Add ClusterDisabled to color scheme
- [ ] Task 2: Multi-column CheckBoxes with enable/disable
- [ ] Task 3: Multi-column RadioButtons with enable/disable
- [ ] Task 4: Integration checkpoint — Multi-column clusters
- [ ] Task 5: Multi-column ListViewer
- [ ] Task 6: Demo app + E2e tests

---

### Task 1: Add ClusterDisabled to Color Scheme

**Files:**
- Modify: `theme/scheme.go`
- Modify: `theme/borland.go`
- Modify: `theme/borland_cyan.go`
- Modify: `theme/borland_gray.go`
- Modify: `theme/matrix.go`
- Modify: `theme/c64.go`

**Requirements:**
- `ColorScheme` has a `ClusterDisabled` field of type `tcell.Style`
- Each of the 5 themes sets `ClusterDisabled` to a non-zero style
- `ClusterDisabled` has a visually dimmed appearance — lower contrast than the corresponding `CheckBoxNormal`/`RadioButtonNormal` in each theme
- All existing theme tests continue to pass

**Implementation:**

Add to `theme/scheme.go` after the `RadioButtonSelected` line:

```go
ClusterDisabled     tcell.Style
```

Add to each theme file's `ColorScheme` literal. Color choices per theme:

| Theme | ClusterDisabled fg | ClusterDisabled bg | Rationale |
|-------|-------------------|-------------------|-----------|
| BorlandBlue | DarkGray | Teal | Dimmed variant of CheckBoxNormal (Black on Teal) |
| BorlandCyan | DarkGray | Silver | Dimmed variant of CheckBoxNormal (Black on Silver) |
| BorlandGray | DarkGray | Silver | Same as Cyan |
| Matrix | DarkGreen (color #22) | Black | Dimmed Green on Black |
| C64 | Gray | Purple | Dimmed variant of CheckBoxNormal (White on Purple) |

For BorlandBlue, add after `RadioButtonSelected`:
```go
ClusterDisabled:     s(tcell.ColorDarkGray, tcell.ColorTeal),
```

For BorlandCyan and BorlandGray:
```go
ClusterDisabled:     s(tcell.ColorDarkGray, tcell.ColorSilver),
```

For Matrix:
```go
ClusterDisabled:     s(tcell.ColorDarkGreen, tcell.ColorBlack),
```

For C64:
```go
ClusterDisabled:     s(tcell.ColorGray, tcell.ColorPurple),
```

**Run tests:** `go test ./theme/ -v`

**Commit:** `git commit -m "feat: add ClusterDisabled color to all themes"`

---

### Task 2: Multi-column CheckBoxes with Enable/Disable

**Files:**
- Modify: `tv/checkbox.go`

**Requirements:**

**Multi-column layout:**
- Column count is determined dynamically: `numCols = ceil(len(items) / bounds.Height())`. If `bounds.Height() >= len(items)`, everything fits in one column (backward compatible).
- Items fill top-to-bottom within a column: `row(item) = item % height`, `col(item) = item / height`
- Column width = width of the widest label in that column + 6 (1 for focus indicator + 1 for `[` + 1 for mark + 1 for `]` + 1 for space + 1 for inter-column gap). This matches original `tcluster.cpp` which uses `width + 6`.
- Column x-position = sum of widths of all previous columns
- `NewCheckBoxes(bounds, labels)` with 3 labels and `bounds.Height() == 2` produces 2 columns: items 0,1 in column 0; item 2 in column 1
- `NewCheckBoxes(bounds, labels)` with 3 labels and `bounds.Height() >= 3` produces 1 column (backward compatible with existing tests)
- `SetBounds` triggers relayout — changing height can change column count
- Each item's Bounds are set to `NewRect(colX, row, colWidth, 1)` where colWidth is the computed width for that column

**Enable/disable:**
- `CheckBoxes` has a `SetEnabled(index int, enabled bool)` method
- `CheckBoxes` has an `IsEnabled(index int) bool` method
- Internally stored as a bitmask `enableMask uint32` (all bits set = all enabled by default)
- Disabled items render using `ClusterDisabled` color scheme style instead of `CheckBoxNormal`
- Disabled items' focus indicator column shows a space (never `►`)
- Keyboard navigation (Up/Down/Left/Right) skips disabled items — if the next item in the navigation direction is disabled, keep searching in that direction until an enabled item is found or the boundary is reached
- Mouse clicks on disabled items are ignored (the click does nothing, event is still consumed)
- Space bar on a disabled item does nothing (no toggle)
- If a shortcut key (Alt+letter) matches a disabled item, the shortcut is ignored
- If all items become disabled, `CheckBoxes` clears `OfSelectable` so it is skipped by Tab navigation. When at least one item is re-enabled, `OfSelectable` is restored.

**Plain-letter shortcut matching:**
- When CheckBoxes has focus (SfSelected is true), a plain letter keystroke (no Alt modifier) that matches a shortcut activates that item — same as Alt+letter but only when focused
- When CheckBoxes does NOT have focus but receives an event during the postProcess phase (OfPostProcess), it also matches plain-letter shortcuts. On match, it requests focus from its owner before activating the item. This allows typing a shortcut letter when another widget in the same Group is focused.
- To enable this, add `OfPostProcess` alongside the existing `OfPreProcess` in the constructor
- This matches original TV behavior where `tcluster::handleEvent` catches unmodified shortcut letters during focused and postProcess phases

**Keyboard navigation:**
- Left arrow: move focus to the item at `currentIndex - height` (previous column, same row). If that index is < 0 or disabled, don't move.
- Right arrow: move focus to the item at `currentIndex + height` (next column, same row). If that index is >= len(items) or disabled, don't move.
- Up/Down arrows: same as current but skip disabled items

**Implementation:**

Add fields to `CheckBoxes`:
```go
type CheckBoxes struct {
    BaseView
    group      *Group
    items      []*CheckBox
    enableMask uint32
}
```

Initialize `enableMask` to `^uint32(0)` (all enabled) in the constructor.

Add `OfPostProcess` to the constructor (alongside existing `OfPreProcess`):
```go
cbs.SetOptions(OfPostProcess, true)
```

Add a `relayoutItems()` method to `CheckBoxes`:

```go
func (cbs *CheckBoxes) relayoutItems() {
    h := cbs.Bounds().Height()
    if h <= 0 {
        h = len(cbs.items)
    }
    if h <= 0 {
        return
    }

    numCols := (len(cbs.items) + h - 1) / h

    // Compute column widths
    colWidths := make([]int, numCols)
    for i, item := range cbs.items {
        col := i / h
        w := labelDisplayWidth(item.label) + 6 // focus + [ + mark + ] + space + gap
        if w > colWidths[col] {
            colWidths[col] = w
        }
    }

    // Position each item and set grow mode
    lastCol := numCols - 1
    for i, item := range cbs.items {
        col := i / h
        row := i % h
        colX := 0
        for c := 0; c < col; c++ {
            colX += colWidths[c]
        }
        item.SetBounds(NewRect(colX, row, colWidths[col], 1))
        if col == lastCol {
            item.SetGrowMode(GfGrowHiX)
        } else {
            item.SetGrowMode(0) // fixed width for non-rightmost columns
        }
    }
}
```

Add a package-level helper to compute label display width (stripping tilde markers):
```go
func labelDisplayWidth(label string) int {
    w := 0
    segments := ParseTildeLabel(label)
    for _, seg := range segments {
        w += utf8.RuneCountInString(seg.Text)
    }
    return w
}
```

This helper is used by both CheckBoxes and RadioButtons (Tasks 2 and 3). Place it in `checkbox.go` since that's the first task to use it; Task 3 reuses it directly.

Call `relayoutItems()` at the end of `NewCheckBoxes` (after all items inserted) and in `SetBounds` (after updating group bounds).

Add enable/disable methods:
```go
func (cbs *CheckBoxes) SetEnabled(index int, enabled bool) {
    if index < 0 || index >= len(cbs.items) {
        return
    }
    if enabled {
        cbs.enableMask |= 1 << uint(index)
    } else {
        cbs.enableMask &^= 1 << uint(index)
    }
    // If all items disabled, clear OfSelectable; restore when any enabled
    anyEnabled := cbs.enableMask & ((1 << uint(len(cbs.items))) - 1) != 0
    cbs.SetOptions(OfSelectable, anyEnabled)
}

func (cbs *CheckBoxes) IsEnabled(index int) bool {
    if index < 0 || index >= len(cbs.items) {
        return false
    }
    return cbs.enableMask&(1<<uint(index)) != 0
}
```

Modify `CheckBox.Draw` to accept a disabled flag. The cleanest approach: add a `disabled` field to CheckBox that CheckBoxes sets before drawing. Or, pass through via the color scheme lookup. Simplest: add a `disabled bool` field to `CheckBox`.

```go
type CheckBox struct {
    BaseView
    label    string
    shortcut rune
    checked  bool
    disabled bool
}
```

In `CheckBox.HandleEvent`, add an early return when disabled so Space/click do nothing:
```go
func (cb *CheckBox) HandleEvent(event *Event) {
    if cb.disabled {
        return
    }
    // ... existing logic unchanged
}
```

In `CheckBox.Draw`, when `cb.disabled` is true:
- Use `cs.ClusterDisabled` style for all rendering (indicator, bracket, mark, label)
- Never show `►` even if focused

In `CheckBoxes`, sync the disabled state before drawing:
```go
func (cbs *CheckBoxes) Draw(buf *DrawBuffer) {
    for i, item := range cbs.items {
        item.disabled = !cbs.IsEnabled(i)
        childBounds := item.Bounds()
        sub := buf.SubBuffer(childBounds)
        item.Draw(sub)
    }
}
```

Modify `moveNavigation` to skip disabled items:
```go
func (cbs *CheckBoxes) moveNavigation(delta int) {
    current := -1
    for i, item := range cbs.items {
        if item.HasState(SfSelected) {
            current = i
            break
        }
    }
    if current < 0 {
        current = 0
    }
    next := current + delta
    // Skip disabled items in the delta direction
    for next >= 0 && next < len(cbs.items) && !cbs.IsEnabled(next) {
        next += delta
    }
    if next < 0 || next >= len(cbs.items) {
        return
    }
    cbs.group.SetFocusedChild(cbs.items[next])
}
```

Add Left/Right arrow handling in `HandleEvent` (alongside existing Up/Down):
```go
case tcell.KeyLeft:
    h := cbs.Bounds().Height()
    if h <= 0 { h = len(cbs.items) }
    cbs.moveNavigation(-h)
    event.Clear()
    return
case tcell.KeyRight:
    h := cbs.Bounds().Height()
    if h <= 0 { h = len(cbs.items) }
    cbs.moveNavigation(h)
    event.Clear()
    return
```

Modify Alt+shortcut handling to skip disabled items:
```go
if item.Shortcut() != 0 && unicode.ToLower(item.Shortcut()) == r {
    if !cbs.IsEnabled(i) {  // add index 'i' to the loop
        continue  // skip disabled
    }
    cbs.group.SetFocusedChild(item)
    item.SetChecked(!item.Checked())
    event.Clear()
    return
}
```

Note: need to change the Alt+shortcut loop to `for i, item := range cbs.items` (add index variable).

Add plain-letter shortcut matching. This fires both when focused (Phase 2) and during postProcess (Phase 3, unfocused). Place this BEFORE the `cbs.HasState(SfSelected)` guard so it runs even when unfocused:
```go
// Plain letter shortcut matching (focused OR postProcess phase)
if event.What == EvKeyboard && event.Key != nil &&
    event.Key.Key == tcell.KeyRune &&
    event.Key.Modifiers == 0 {

    r := unicode.ToLower(event.Key.Rune)
    for i, item := range cbs.items {
        if item.Shortcut() != 0 && unicode.ToLower(item.Shortcut()) == r {
            if !cbs.IsEnabled(i) {
                continue
            }
            // Grab focus from parent first (important for postProcess path)
            if owner, ok := cbs.Owner().(Container); ok && !cbs.HasState(SfSelected) {
                owner.SetFocusedChild(cbs)
            }
            cbs.group.SetFocusedChild(item)
            item.SetChecked(!item.Checked())
            event.Clear()
            return
        }
    }
}
```

Modify mouse click handling to skip disabled items. In the mouse routing section:
```go
for i, item := range cbs.items {
    if item.Bounds().Contains(NewPoint(mx, my)) {
        if !cbs.IsEnabled(i) {
            event.Clear() // consume but ignore
            return
        }
        // ... forward to item
    }
}
```

**Run tests:** `go test ./tv/ -run "CheckBox" -v`

**Commit:** `git commit -m "feat: add multi-column layout and enable/disable to CheckBoxes"`

---

### Task 3: Multi-column RadioButtons with Enable/Disable

**Files:**
- Modify: `tv/radio.go`

**Requirements:**

**Multi-column layout:**
- Same algorithm as CheckBoxes Task 2: `numCols = ceil(len(items) / height)`, items fill top-to-bottom, column width = widest label + 6
- `NewRadioButtons(bounds, labels)` with `bounds.Height() < len(labels)` produces multiple columns
- `SetBounds` triggers relayout
- Backward compatible: if `bounds.Height() >= len(labels)`, single column as before

**Enable/disable:**
- `RadioButtons` has `SetEnabled(index int, enabled bool)` and `IsEnabled(index int) bool`
- Uses `enableMask uint32` (all enabled by default)
- Disabled items render with `ClusterDisabled` style
- Navigation skips disabled items
- Mouse clicks on disabled items are ignored
- Space/Enter on a disabled item does nothing
- Alt+shortcut for disabled items is ignored
- When the currently selected radio button becomes disabled, the selection does NOT change (the radio can still be visually selected but dimmed)
- If all items become disabled, `RadioButtons` clears `OfSelectable` so it is skipped by Tab navigation. When at least one item is re-enabled, `OfSelectable` is restored.

**Plain-letter shortcut matching:**
- When RadioButtons has focus (SfSelected is true), a plain letter keystroke (no Alt modifier) that matches a shortcut activates that item — same as Alt+letter but only when focused
- When RadioButtons does NOT have focus but receives an event during the postProcess phase (OfPostProcess), it also matches plain-letter shortcuts. On match, it requests focus from its owner before activating.
- To enable this, add `OfPostProcess` alongside the existing `OfPreProcess` in the constructor
- This matches original TV behavior where `tcluster::handleEvent` catches unmodified shortcut letters during focused and postProcess phases

**Keyboard navigation:**
- Left/Right: move by `height` items (column navigation), skip disabled
- Up/Down: move by 1 item (row navigation), skip disabled
- Navigation moves both focus AND selection for RadioButtons (unlike CheckBoxes where navigation only moves focus)

**Implementation:**

Very similar to Task 2. Key differences:

Add `enableMask uint32` to `RadioButtons`, initialize to `^uint32(0)`.

Add `OfPostProcess` to the constructor:
```go
rbs.SetOptions(OfPostProcess, true)
```

Add `relayoutItems()` — identical algorithm to CheckBoxes including `GfGrowHiX` for rightmost-column items only (others get `GfGrowMode(0)`). Reuse the `labelDisplayWidth()` package-level function added in Task 2.

Add `disabled bool` field to `RadioButton`. Sync in `RadioButtons.Draw`.

In `RadioButton.HandleEvent`, add an early return when disabled so Space/Enter/click do nothing:
```go
func (rb *RadioButton) HandleEvent(event *Event) {
    if rb.disabled {
        return
    }
    // ... existing logic unchanged
}
```

In `RadioButton.Draw`, when disabled: use `ClusterDisabled`, never show `►`.

Modify `moveSelection` to skip disabled:
```go
func (rbs *RadioButtons) moveSelection(delta int) {
    current := rbs.Value()
    if current < 0 {
        current = 0
    }
    next := current + delta
    for next >= 0 && next < len(rbs.items) && !rbs.IsEnabled(next) {
        next += delta
    }
    if next < 0 || next >= len(rbs.items) {
        return
    }
    rbs.SetValue(next)
    rbs.group.SetFocusedChild(rbs.items[next])
}
```

**Replace the entire switch block** in `RadioButtons.HandleEvent` (currently at lines 251-261 in `tv/radio.go`). The existing code combines `KeyDown, KeyRight` and `KeyUp, KeyLeft` — this must be split so Left/Right do column navigation:

```go
if event.What == EvKeyboard && event.Key != nil && rbs.HasState(SfSelected) {
    switch event.Key.Key {
    case tcell.KeyDown:
        rbs.moveSelection(1)
        event.Clear()
        return
    case tcell.KeyUp:
        rbs.moveSelection(-1)
        event.Clear()
        return
    case tcell.KeyRight:
        h := rbs.Bounds().Height()
        if h <= 0 { h = len(rbs.items) }
        rbs.moveSelection(h)
        event.Clear()
        return
    case tcell.KeyLeft:
        h := rbs.Bounds().Height()
        if h <= 0 { h = len(rbs.items) }
        rbs.moveSelection(-h)
        event.Clear()
        return
    }
}
```

Add plain-letter shortcut matching (focused OR postProcess phase — same pattern as CheckBoxes Task 2). Place BEFORE the `rbs.HasState(SfSelected)` guard:
```go
if event.What == EvKeyboard && event.Key != nil &&
    event.Key.Key == tcell.KeyRune &&
    event.Key.Modifiers == 0 {

    r := unicode.ToLower(event.Key.Rune)
    for i, item := range rbs.items {
        if item.Shortcut() != 0 && unicode.ToLower(item.Shortcut()) == r {
            if !rbs.IsEnabled(i) {
                continue
            }
            // Grab focus from parent first (important for postProcess path)
            if owner, ok := rbs.Owner().(Container); ok && !rbs.HasState(SfSelected) {
                owner.SetFocusedChild(rbs)
            }
            rbs.group.SetFocusedChild(item)
            rbs.SetValue(i)
            event.Clear()
            return
        }
    }
}
```

Mouse routing: add disabled check (same pattern as CheckBoxes).

Alt+shortcut: add disabled check (same pattern as CheckBoxes).

**Run tests:** `go test ./tv/ -run "Radio" -v`

**Commit:** `git commit -m "feat: add multi-column layout and enable/disable to RadioButtons"`

---

### Task 4: Integration Checkpoint — Multi-column Clusters

**Purpose:** Verify that multi-column CheckBoxes and RadioButtons work correctly in real Container hierarchies (Window, Dialog) with proper focus, keyboard, and mouse interaction.

**Requirements (for test writer):**
- A CheckBoxes with 6 items in a bounds of height 3 produces 2 columns, with items 0-2 in column 0 and items 3-5 in column 1
- Right arrow from item 0 (column 0, row 0) focuses item 3 (column 1, row 0)
- Left arrow from item 3 (column 1, row 0) focuses item 0 (column 0, row 0)
- Down arrow from item 0 focuses item 1 (same column, next row)
- A disabled item is skipped by arrow navigation: if item 1 is disabled, Down from item 0 goes to item 2
- Mouse click on a disabled item does not toggle its state
- A CheckBoxes inside a Window receives keyboard events through the Window's three-phase dispatch
- A RadioButtons with 4 items in height 2 produces 2 columns: items 0,1 in column 0; items 2,3 in column 1
- Right arrow in RadioButtons moves selection (not just focus) to the next column
- Alt+shortcut on a disabled RadioButton item is ignored
- SetBounds on a CheckBoxes that changes height from 3 to 2 relays out items from 1 column to 2 columns

**Components to wire up:** CheckBoxes, RadioButtons, Window, Group (all real, no mocks)

**Run:** `go test ./tv/ -run TestIntegrationMultiCol -v`

**Commit:** `git commit -m "test: add integration tests for multi-column clusters"`

---

### Task 5: Multi-column ListViewer

**Files:**
- Modify: `tv/list_viewer.go`
- Modify: `tv/scrollbar.go` (fix arrow click to use `arStep`)

**Requirements:**

**Configuration:**
- `ListViewer` has a `numCols` field (default 1 for backward compatibility)
- `SetNumCols(n int)` sets the number of columns (minimum 1)
- `NumCols() int` returns the current column count

**Layout:**
- Column width = `bounds.Width() / numCols`
- Items per visible page = `numCols * bounds.Height()`
- Column `c` shows items starting at `topIndex + c * height` through `topIndex + (c+1) * height - 1`
- A 1-character vertical divider `│` is drawn between columns in normal style

**Drawing:**
- Iterate columns left-to-right
- Each column draws its items as a vertical slice at x-offset `col * columnWidth`
- Focused/selected item highlight spans only its column width (not the full viewer width)
- The divider `│` is drawn at `colWidth - 1` for columns 0 through numCols-2
- Item text is truncated to `colWidth - 1` (leaving room for divider, except last column which uses full colWidth)

**Keyboard:**
- Up/Down: move within the current column (same as current, delta ±1)
- Left: move focus by `-height` items (previous column, same row). If index would go below 0, don't move.
- Right: move focus by `+height` items (next column, same row). If index would exceed count, don't move.
- PgUp/PgDn: move by `numCols * height` (one full page)
- Home/End: first/last visible item on current page
- Ctrl+PgUp/PgDn: absolute first/last item

**Mouse:**
- Click x-position determines column: `col = x / colWidth`
- Click y-position determines row within column
- Item index = `topIndex + col * height + row`
- If computed index >= count, ignore the click

**Scrollbar:**
- Range: `(0, itemCount)`, value: `topIndex`
- Page step = `numCols * height`
- Arrow step = `height` via `SetArStep(height)` (matches original `tlstview.cpp:48-58` — clicking the scrollbar arrow scrolls one row across all columns)
- **Bug fix needed:** The scrollbar's click handlers (`handleVerticalClick`, `handleHorizontalClick`) currently use `sb.step(-1)` / `sb.step(1)` instead of `sb.step(-sb.arStep)` / `sb.step(sb.arStep)`. Fix these 4 call sites in `tv/scrollbar.go` so that clicking the arrow buttons respects `arStep`.

**Implementation:**

Add fields:
```go
type ListViewer struct {
    BaseView
    dataSource ListDataSource
    selected   int
    topIndex   int
    numCols    int  // NEW: default 1
    scrollBar  *ScrollBar
    dragging   bool
    OnSelect   func(int)
}
```

Initialize `numCols: 1` in `NewListViewer`.

Add accessors:
```go
func (lv *ListViewer) NumCols() int { return lv.numCols }
func (lv *ListViewer) SetNumCols(n int) {
    if n < 1 { n = 1 }
    lv.numCols = n
    lv.ensureVisible()
    lv.syncScrollBar()
}
```

Modify `visibleHeight` — it stays the same (returns bounds.Height()). Add:
```go
func (lv *ListViewer) itemsPerPage() int {
    return lv.numCols * lv.visibleHeight()
}

func (lv *ListViewer) colWidth() int {
    if lv.numCols <= 1 {
        return lv.Bounds().Width()
    }
    return lv.Bounds().Width() / lv.numCols
}
```

Modify `Draw`:
```go
func (lv *ListViewer) Draw(buf *DrawBuffer) {
    w := lv.Bounds().Width()
    vh := lv.visibleHeight()
    cs := lv.ColorScheme()
    normalStyle := tcell.StyleDefault
    selectedStyle := tcell.StyleDefault
    focusedStyle := tcell.StyleDefault
    if cs != nil {
        normalStyle = cs.ListNormal
        selectedStyle = cs.ListSelected
        focusedStyle = cs.ListFocused
    }

    buf.Fill(NewRect(0, 0, w, vh), ' ', normalStyle)

    count := lv.dataSource.Count()
    if count == 0 {
        text := "<empty>"
        for i, ch := range text {
            if i < w { buf.WriteChar(i, 0, ch, normalStyle) }
        }
        return
    }

    hasFocus := lv.HasState(SfSelected)
    cw := lv.colWidth()

    for col := 0; col < lv.numCols; col++ {
        colX := col * cw
        drawW := cw
        if col < lv.numCols-1 {
            drawW = cw - 1 // leave room for divider
        }

        for row := 0; row < vh; row++ {
            idx := lv.topIndex + col*vh + row
            if idx >= count {
                break
            }

            style := normalStyle
            if idx == lv.selected {
                if hasFocus {
                    style = focusedStyle
                } else {
                    style = selectedStyle
                }
                for x := 0; x < drawW; x++ {
                    buf.WriteChar(colX+x, row, ' ', style)
                }
            }

            text := lv.dataSource.Item(idx)
            runes := []rune(text)
            if len(runes) > drawW {
                runes = runes[:drawW]
            }
            for i, ch := range runes {
                buf.WriteChar(colX+i, row, ch, style)
            }
        }

        // Draw column divider
        if col < lv.numCols-1 {
            for row := 0; row < vh; row++ {
                buf.WriteChar(colX+cw-1, row, '│', normalStyle)
            }
        }
    }
}
```

Modify keyboard handling — add Left/Right:
```go
case tcell.KeyLeft:
    vh := lv.visibleHeight()
    if lv.selected >= vh {
        lv.selected -= vh
        lv.ensureVisible()
        lv.syncScrollBar()
    }
    event.Clear()

case tcell.KeyRight:
    vh := lv.visibleHeight()
    if lv.selected+vh < count {
        lv.selected += vh
        lv.ensureVisible()
        lv.syncScrollBar()
    }
    event.Clear()
```

Update PgUp/PgDn to use `itemsPerPage()`:
```go
case tcell.KeyPgDn:
    lv.selected += lv.itemsPerPage()
    if lv.selected >= count { lv.selected = count - 1 }
    lv.ensureVisible()
    lv.syncScrollBar()
    event.Clear()

case tcell.KeyPgUp:
    lv.selected -= lv.itemsPerPage()
    if lv.selected < 0 { lv.selected = 0 }
    lv.ensureVisible()
    lv.syncScrollBar()
    event.Clear()
```

Update End to use `itemsPerPage()` (existing code uses `visibleHeight()` which only covers column 0):
```go
case tcell.KeyEnd:
    lastVisible := lv.topIndex + lv.itemsPerPage() - 1
    if lastVisible >= count {
        lastVisible = count - 1
    }
    lv.selected = lastVisible
    lv.syncScrollBar()
    event.Clear()
```

Modify mouse handling:
```go
// In the "Normal click/drag within bounds" section:
cw := lv.colWidth()
col := 0
if cw > 0 {
    col = mx / cw
}
if col >= lv.numCols {
    col = lv.numCols - 1
}
clickIdx := lv.topIndex + col*lv.visibleHeight() + my
if clickIdx >= 0 && clickIdx < count {
    lv.selected = clickIdx
}
```

Modify `syncScrollBar` — note `SetArStep` for multi-column scrolling:
```go
func (lv *ListViewer) syncScrollBar() {
    if lv.scrollBar == nil { return }
    count := lv.dataSource.Count()
    lv.scrollBar.SetRange(0, count)
    lv.scrollBar.SetPageSize(lv.itemsPerPage())
    lv.scrollBar.SetValue(lv.topIndex)
    vh := lv.visibleHeight()
    if vh < 1 { vh = 1 }
    lv.scrollBar.SetArStep(vh)
}
```

Modify `ensureVisible` for multi-column awareness:
```go
func (lv *ListViewer) ensureVisible() {
    vh := lv.visibleHeight()
    if vh <= 0 { return }
    ipp := lv.itemsPerPage()
    if lv.selected < lv.topIndex {
        lv.topIndex = lv.selected
    }
    if lv.selected >= lv.topIndex+ipp {
        // Selected is beyond current page — scroll so it's on the last page
        lv.topIndex = lv.selected - ipp + 1
    }
    lv.clampTopIndex()
}
```

Modify `clampTopIndex`:
```go
func (lv *ListViewer) clampTopIndex() {
    count := lv.dataSource.Count()
    ipp := lv.itemsPerPage()
    maxTop := count - ipp
    if maxTop < 0 { maxTop = 0 }
    if lv.topIndex > maxTop { lv.topIndex = maxTop }
    if lv.topIndex < 0 { lv.topIndex = 0 }
}
```

**Fix scrollbar arrow click handlers** in `tv/scrollbar.go`:

In `handleVerticalClick`, find the two arrow-button cases (`my == 0` and `my == h-1`) and change:
- `sb.step(-1)` → `sb.step(-sb.arStep)`
- `sb.step(1)` → `sb.step(sb.arStep)`

In `handleHorizontalClick`, find the two arrow-button cases (`mx == 0` and `mx == w-1`) and change:
- `sb.step(-1)` → `sb.step(-sb.arStep)`
- `sb.step(1)` → `sb.step(sb.arStep)`

This ensures that clicking the scrollbar arrow buttons respects the configured `arStep`, which defaults to 1 (preserving backward compatibility) but is set to `height` by multi-column ListViewer.

**Important backward compatibility:** When `numCols == 1`, all behavior is identical to the current single-column ListViewer. The Left/Right keys are new but don't conflict since single-column ListViewer currently ignores them.

**Run tests:** `go test ./tv/ -run "ListViewer" -v`

**Commit:** `git commit -m "feat: add multi-column display to ListViewer"`

---

### Task 6: Demo App + E2e Tests

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**

**Demo app changes:**
- Change CheckBoxes in Window 1 to `bounds.Height() == 2` with 3 items → forces 2 columns (items "Read only", "Hidden" in column 0; "System" in column 1)
- Change RadioButtons similarly to `bounds.Height() == 2` with 3 items → forces 2 columns
- Window 1 height may need adjustment to fit the new layout (items take fewer rows now)
- All existing widgets in Window 1 still render correctly
- Set the existing ListBox in Window 2 to 2 columns via `listBox.SetNumCols(2)` to demonstrate multi-column ListViewer (spec 5.6 suggests this)

**E2e tests:**
- Build the binary, launch in tmux
- Navigate to Window 1 (Alt+1)
- Verify CheckBoxes render in 2 columns: both "Read only" and "System" visible on the same or adjacent rows (they're in different columns)
- Verify RadioButtons render in 2 columns: both "Text" and "Hex" visible on adjacent rows
- Close overlapping windows if needed to see win1 clearly
- All existing e2e tests must continue to pass

**Implementation:**

In `e2e/testapp/basic/main.go`, change:
```go
// Current: height 3, single column
checkBoxes := tv.NewCheckBoxes(tv.NewRect(1, 5, 25, 3), []string{"~R~ead only", "~H~idden", "~S~ystem"})
// New: height 2, two columns
checkBoxes := tv.NewCheckBoxes(tv.NewRect(1, 5, 30, 2), []string{"~R~ead only", "~H~idden", "~S~ystem"})
```

```go
// Current: height 3, single column
radioButtons := tv.NewRadioButtons(tv.NewRect(1, 9, 25, 3), []string{"~T~ext", "~B~inary", "~H~ex"})
// New: height 2, two columns (adjust y-position since checkboxes now take 2 rows instead of 3)
radioButtons := tv.NewRadioButtons(tv.NewRect(1, 8, 30, 2), []string{"~T~ext", "~B~inary", "~H~ex"})
```

Adjust subsequent y-positions since items now take less vertical space:
- RadioButtons moved from y=9 to y=8 (since checkboxes shrunk from 3 to 2 rows)
- InputLine moved from y=12 to y=11
- History, nameLabel, portInput, portLabel: adjust y accordingly
- Window 1 may shrink from height 16 to height 15

Set the existing ListBox in Window 2 to 2 columns. After `listBox` is created:
```go
listBox.ListViewer().SetNumCols(2)
```

E2e test: verify the multi-column layout is visible. After closing overlapping windows and focusing win1, check that checkbox labels from different columns appear on the same row.

**Run tests:** `go test ./e2e/ -timeout 300s -v`

**Commit:** `git commit -m "feat: update demo app with multi-column clusters and e2e tests"`
