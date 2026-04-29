# Phase 6: ScrollBar, ListViewer, and SetColorScheme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add scrollable list capability — ScrollBar, ListViewer (with concrete StringList), and per-window color scheme assignment — so the demo app can display a navigable, scrolling list in a differently-themed window.

**Architecture:** ScrollBar is a standalone Widget that renders a vertical or horizontal track with thumb, arrows, and page areas. ListViewer is a Widget that displays a scrollable list with keyboard navigation and binds to a ScrollBar for visual feedback. SetColorScheme is a simple method on Window and Dialog that sets the BaseView.scheme field, enabling per-window theming via the existing ColorScheme() owner-chain walk.

**Tech Stack:** Go 1.22+, tcell/v2, existing tv framework patterns (BaseView embedding, Widget interface)

---

## File Map

| File | Responsibility |
|------|---------------|
| `tv/scrollbar.go` (create) | ScrollBar widget — vertical/horizontal, thumb, arrows, page areas, mouse interaction |
| `tv/list_viewer.go` (create) | ListViewer widget — abstract scrollable list with keyboard navigation, selection, ScrollBar binding; StringList concrete implementation |
| `tv/window.go` (modify) | Add `SetColorScheme(*theme.ColorScheme)` method |
| `tv/dialog.go` (modify) | Add `SetColorScheme(*theme.ColorScheme)` method |
| `e2e/testapp/basic/main.go` (modify) | Add a ListViewer with ScrollBar in a differently-themed window |
| `e2e/e2e_test.go` (modify) | E2E tests for scrollable list visibility and navigation |

---

## Batch 1: ScrollBar + SetColorScheme (Tasks 1-2)

### Task 1: ScrollBar Widget

**Files:**
- Create: `tv/scrollbar.go`
- Test: `tv/scrollbar_test.go`

**Requirements:**

Constructor:
- `NewScrollBar(bounds Rect, orientation Orientation) *ScrollBar`
- `Orientation` is a new type: `const (Horizontal Orientation = iota; Vertical)`
- Sets `SfVisible` by default; does NOT set `OfSelectable` (scrollbars don't receive focus)
- Implements the `Widget` interface

State:
- `min, max, value, pageSize int` — the scrollable range
- `SetRange(min, max int)` sets the scrollable range
- `SetValue(v int)` sets current scroll position (clamped to [min, max-pageSize])
- `SetPageSize(n int)` sets the visible page size (used for thumb proportional sizing and page up/down)
- `Value() int` returns current value
- `Min() int`, `Max() int`, `PageSize() int` return current settings

Callback:
- `OnChange func(value int)` — called when user interaction changes the value (mouse click on arrows, page area, or thumb drag)

Drawing (Vertical, width=1):
- Row 0: up arrow `▲` in `ScrollBar` style
- Rows 1 to height-2: track area filled with `░` in `ScrollBar` style
- Thumb position within track area rendered as `█` in `ScrollThumb` style
- Last row: down arrow `▼` in `ScrollBar` style
- Thumb position calculation: if max <= min or pageSize >= (max-min), thumb fills entire track. Otherwise:
  - trackLen = height - 2 (excluding arrows)
  - thumbLen = max(1, trackLen * pageSize / (max - min))
  - thumbPos = (value - min) * (trackLen - thumbLen) / (max - min - pageSize) (clamped to [0, trackLen-thumbLen])
- When trackLen < 1 (widget too small), draw nothing between arrows

Drawing (Horizontal, height=1):
- Col 0: left arrow `◄` in `ScrollBar` style
- Cols 1 to width-2: track area with `░` in `ScrollBar` style
- Thumb rendered as `█` in `ScrollThumb` style (same proportional calculation, using width instead of height)
- Last col: right arrow `►` in `ScrollBar` style

Mouse handling:
- Click on up arrow (vertical) or left arrow (horizontal): decrement value by 1, clamp, call OnChange
- Click on down arrow or right arrow: increment value by 1, clamp, call OnChange
- Click in track area above/left of thumb: decrement value by pageSize, clamp, call OnChange
- Click in track area below/right of thumb: increment value by pageSize, clamp, call OnChange
- Mouse wheel up (`WheelUp`): decrement value by 1, clamp, call OnChange
- Mouse wheel down (`WheelDown`): increment value by 1, clamp, call OnChange
- All mouse interactions consume the event

**Implementation:**

```go
package tv

import "github.com/gdamore/tcell/v2"

var _ Widget = (*ScrollBar)(nil)

type Orientation int

const (
    Horizontal Orientation = iota
    Vertical
)

type ScrollBar struct {
    BaseView
    orientation Orientation
    min         int
    max         int
    value       int
    pageSize    int
    OnChange    func(int)
}

// Note: Thumb drag is deferred. It requires a mouse capture mechanism (where the
// widget continues receiving mouse events even when the cursor leaves its bounds).
// The framework currently dispatches mouse events based on cursor position. Without
// capture, dragging the thumb 1 pixel outside the scrollbar's bounds silently drops
// the drag. The ScrollBar is fully usable without drag via arrows, page clicks, and
// mouse wheel. Drag will be added when a mouse capture mechanism is implemented.

func NewScrollBar(bounds Rect, orientation Orientation) *ScrollBar {
    sb := &ScrollBar{orientation: orientation}
    sb.SetBounds(bounds)
    sb.SetState(SfVisible, true)
    return sb
}

func (sb *ScrollBar) Min() int          { return sb.min }
func (sb *ScrollBar) Max() int          { return sb.max }
func (sb *ScrollBar) Value() int        { return sb.value }
func (sb *ScrollBar) PageSize() int     { return sb.pageSize }

func (sb *ScrollBar) SetRange(min, max int) {
    sb.min = min
    sb.max = max
    sb.clampValue()
}

func (sb *ScrollBar) SetValue(v int) {
    sb.value = v
    sb.clampValue()
}

func (sb *ScrollBar) SetPageSize(n int) {
    sb.pageSize = n
    sb.clampValue()
}

func (sb *ScrollBar) clampValue() {
    maxVal := sb.max - sb.pageSize
    if maxVal < sb.min {
        maxVal = sb.min
    }
    if sb.value < sb.min {
        sb.value = sb.min
    }
    if sb.value > maxVal {
        sb.value = maxVal
    }
}

func (sb *ScrollBar) trackLen() int {
    if sb.orientation == Vertical {
        return sb.Bounds().Height() - 2
    }
    return sb.Bounds().Width() - 2
}

func (sb *ScrollBar) thumbInfo() (pos, length int) {
    tl := sb.trackLen()
    if tl < 1 {
        return 0, 0
    }
    rng := sb.max - sb.min
    if rng <= 0 || sb.pageSize >= rng {
        return 0, tl
    }
    length = tl * sb.pageSize / rng
    if length < 1 {
        length = 1
    }
    scrollRange := rng - sb.pageSize
    if scrollRange <= 0 {
        return 0, length
    }
    pos = (sb.value - sb.min) * (tl - length) / scrollRange
    if pos < 0 {
        pos = 0
    }
    if pos > tl-length {
        pos = tl - length
    }
    return pos, length
}

func (sb *ScrollBar) Draw(buf *DrawBuffer) {
    cs := sb.ColorScheme()
    barStyle := tcell.StyleDefault
    thumbStyle := tcell.StyleDefault
    if cs != nil {
        barStyle = cs.ScrollBar
        thumbStyle = cs.ScrollThumb
    }

    if sb.orientation == Vertical {
        sb.drawVertical(buf, barStyle, thumbStyle)
    } else {
        sb.drawHorizontal(buf, barStyle, thumbStyle)
    }
}

func (sb *ScrollBar) drawVertical(buf *DrawBuffer, barStyle, thumbStyle tcell.Style) {
    h := sb.Bounds().Height()
    if h < 2 {
        return
    }
    buf.WriteChar(0, 0, '▲', barStyle)
    buf.WriteChar(0, h-1, '▼', barStyle)

    tl := h - 2
    for i := 0; i < tl; i++ {
        buf.WriteChar(0, i+1, '░', barStyle)
    }

    thumbPos, thumbLen := sb.thumbInfo()
    for i := 0; i < thumbLen && i+thumbPos < tl; i++ {
        buf.WriteChar(0, 1+thumbPos+i, '█', thumbStyle)
    }
}

func (sb *ScrollBar) drawHorizontal(buf *DrawBuffer, barStyle, thumbStyle tcell.Style) {
    w := sb.Bounds().Width()
    if w < 2 {
        return
    }
    buf.WriteChar(0, 0, '◄', barStyle)
    buf.WriteChar(w-1, 0, '►', barStyle)

    tl := w - 2
    for i := 0; i < tl; i++ {
        buf.WriteChar(i+1, 0, '░', barStyle)
    }

    thumbPos, thumbLen := sb.thumbInfo()
    for i := 0; i < thumbLen && i+thumbPos < tl; i++ {
        buf.WriteChar(1+thumbPos+i, 0, '█', thumbStyle)
    }
}

func (sb *ScrollBar) HandleEvent(event *Event) {
    if event.What != EvMouse || event.Mouse == nil {
        return
    }

    // Mouse wheel
    if event.Mouse.Button == tcell.WheelUp {
        sb.step(-1)
        event.Clear()
        return
    }
    if event.Mouse.Button == tcell.WheelDown {
        sb.step(1)
        event.Clear()
        return
    }

    if event.Mouse.Button&tcell.Button1 == 0 {
        return
    }

    if sb.orientation == Vertical {
        sb.handleVerticalClick(event)
    } else {
        sb.handleHorizontalClick(event)
    }
}

func (sb *ScrollBar) handleVerticalClick(event *Event) {
    my := event.Mouse.Y
    h := sb.Bounds().Height()

    if my == 0 {
        sb.step(-1)
        event.Clear()
        return
    }
    if my == h-1 {
        sb.step(1)
        event.Clear()
        return
    }

    trackY := my - 1
    thumbPos, _ := sb.thumbInfo()

    if trackY < thumbPos {
        sb.page(-1)
    } else {
        sb.page(1)
    }
    event.Clear()
}

func (sb *ScrollBar) handleHorizontalClick(event *Event) {
    mx := event.Mouse.X
    w := sb.Bounds().Width()

    if mx == 0 {
        sb.step(-1)
        event.Clear()
        return
    }
    if mx == w-1 {
        sb.step(1)
        event.Clear()
        return
    }

    trackX := mx - 1
    thumbPos, _ := sb.thumbInfo()

    if trackX < thumbPos {
        sb.page(-1)
    } else {
        sb.page(1)
    }
    event.Clear()
}

func (sb *ScrollBar) step(dir int) {
    old := sb.value
    sb.value += dir
    sb.clampValue()
    if sb.value != old && sb.OnChange != nil {
        sb.OnChange(sb.value)
    }
}

func (sb *ScrollBar) page(dir int) {
    old := sb.value
    sb.value += dir * sb.pageSize
    sb.clampValue()
    if sb.value != old && sb.OnChange != nil {
        sb.OnChange(sb.value)
    }
}
```

**Run tests:** `go test ./tv/... -run TestScrollBar -v`

**Commit:** `git commit -m "feat(tv): add ScrollBar widget with vertical and horizontal orientation"`

---

### Task 2: SetColorScheme on Window and Dialog

**Files:**
- Modify: `tv/window.go`
- Modify: `tv/dialog.go`
- Test: `tv/colorscheme_test.go`

**Requirements:**

Window:
- `SetColorScheme(cs *theme.ColorScheme)` sets the `BaseView.scheme` field on the Window
- After calling `SetColorScheme`, `window.ColorScheme()` returns the set scheme (not the owner-chain scheme)
- After calling `SetColorScheme(nil)`, `window.ColorScheme()` falls back to the owner-chain scheme
- Children of the window inherit the window's scheme via `ColorScheme()` owner-chain walk

Dialog:
- `SetColorScheme(cs *theme.ColorScheme)` sets the `BaseView.scheme` field on the Dialog
- Same inheritance behavior as Window

Note: `BaseView` already has a `scheme` field and `ColorScheme()` walks the owner chain. The only thing missing is a public setter. The implementation is one line per type.

**Implementation:**

Add to `tv/window.go`:
```go
func (w *Window) SetColorScheme(cs *theme.ColorScheme) {
    w.scheme = cs
}
```

Add to `tv/dialog.go`:
```go
func (d *Dialog) SetColorScheme(cs *theme.ColorScheme) {
    d.scheme = cs
}
```

**Run tests:** `go test ./tv/... -run TestColorScheme -v`

**Commit:** `git commit -m "feat(tv): add SetColorScheme on Window and Dialog for per-window theming"`

---

### Task 3: Integration Checkpoint — ScrollBar and ColorScheme

**Purpose:** Verify ScrollBar works inside a Window with proper color scheme inheritance, and that SetColorScheme produces different visual output.

**Requirements (for test writer):**
- A ScrollBar inside a Window uses the Window's color scheme for rendering (ScrollBar and ScrollThumb styles)
- A ScrollBar inside a Window with a custom ColorScheme (via SetColorScheme) uses the custom scheme's styles
- Two windows with different color schemes render their ScrollBars with different styles
- ScrollBar mouse click events (arrow clicks, page clicks) work when the ScrollBar is inside a Window (events route through Window's client area offset)
- ScrollBar OnChange callback fires when clicked via the real event dispatch chain inside a Window

**Components to wire up:** Application (with SimulationScreen), Desktop, Window, ScrollBar (all real, no mocks)

**Run:** `go test ./tv/... -run TestIntegrationPhase6 -v`

**Commit:** `git commit -m "test(tv): add Phase 6 integration tests for ScrollBar and ColorScheme"`

---

## Batch 2: ListViewer (Tasks 4-5)

### Task 4: ListViewer Widget

**Files:**
- Create: `tv/list_viewer.go`
- Test: `tv/list_viewer_test.go`

**Requirements:**

DataSource interface:
- `ListDataSource` interface: `Count() int`, `Item(index int) string`
- This is how ListViewer gets its data — any type implementing this interface can supply list items

StringList (concrete data source):
- `StringList` struct wraps `[]string` and implements `ListDataSource`
- `NewStringList(items []string) *StringList`
- `Count() int` returns len(items)
- `Item(index int) string` returns items[index]

**Note on multiple columns:** The spec mentions "Multiple columns" for ListViewer. This is deferred to a future phase. Multiple columns (where a list wraps items into side-by-side columns like Turbo Vision's original TListViewer.numCols) is an independent layout concern that doesn't affect the core scrolling/selection/data-source architecture. The single-column implementation built here is the foundation — adding `numCols` later requires only a Draw change (wrap items across columns) and a keyboard change (Left/Right to move between columns). It does not change the ListDataSource interface, ScrollBar binding, or selection model.

Constructor:
- `NewListViewer(bounds Rect, dataSource ListDataSource) *ListViewer`
- Sets `SfVisible` and `OfSelectable`
- Implements `Widget` interface
- Initial selection: index 0 (if data source has items)
- Initial scroll offset: 0

State:
- `selected int` — currently highlighted item index
- `topIndex int` — index of first visible item (scroll offset)
- `scrollBar *ScrollBar` — optional bound scrollbar (nil by default)
- `Selected() int` returns current selection index
- `SetSelected(index int)` sets selection (clamped to [0, Count()-1]), adjusts topIndex to keep selected visible
- `TopIndex() int` returns current scroll offset
- `DataSource() ListDataSource` returns current data source
- `SetDataSource(ds ListDataSource)` replaces data source, resets selected to 0 and topIndex to 0

Callback:
- `OnSelect func(index int)` — called when user interaction changes the selection (keyboard navigation, mouse click). Not called by programmatic `SetSelected()`.
- This makes the ListViewer useful in interactive applications where the owning code needs to respond to selection changes (e.g., updating a detail view)

ScrollBar binding:
- `SetScrollBar(sb *ScrollBar)` binds a ScrollBar to this ListViewer
- When bound: ListViewer updates ScrollBar range/value/pageSize on any state change (selection, scroll, data source change)
- ScrollBar range: min=0, max=dataSource.Count(), pageSize=visible height
- ScrollBar value: topIndex
- ScrollBar's OnChange is set to update ListViewer's topIndex when user interacts with the scrollbar
- Calling `SetScrollBar(nil)` unbinds

Drawing:
- Fills entire bounds with `ListNormal` style
- Renders visible items starting from `topIndex`, one item per row
- Each item text is truncated to fit the widget width
- The selected item row uses `ListSelected` style when the widget does NOT have `SfSelected` (no focus)
- The selected item row uses `ListFocused` style when the widget has `SfSelected` (focused)
- Unselected rows use `ListNormal` style

Keyboard handling (only when focused):
- `Down arrow`: move selection down by 1 (no-op at last item), scroll if needed, call OnSelect if set, consume event
- `Up arrow`: move selection up by 1 (no-op at first item), scroll if needed, call OnSelect if set, consume event
- `Home`: select first item, scroll to top, call OnSelect if set, consume event
- `End`: select last item, scroll to show last item, call OnSelect if set, consume event
- `PgDn`: move selection down by visible height (clamped to last item), scroll accordingly, call OnSelect if set, consume event
- `PgUp`: move selection up by visible height (clamped to first item), scroll accordingly, call OnSelect if set, consume event
- Events not handled pass through unconsumed

Mouse handling:
- Click (Button1) on a visible row: select that item, call OnSelect if set, consume event
- The clicked row is `topIndex + clickY` (where clickY is the mouse Y coordinate within the widget)

Scroll adjustment:
- When selected index < topIndex, set topIndex = selected (scroll up to show selection)
- When selected index >= topIndex + visible height, set topIndex = selected - visible height + 1 (scroll down)
- After any state change that modifies topIndex or selected, update the bound ScrollBar (if any)

**Implementation:**

```go
package tv

import "github.com/gdamore/tcell/v2"

var _ Widget = (*ListViewer)(nil)

type ListDataSource interface {
    Count() int
    Item(index int) string
}

type StringList struct {
    items []string
}

func NewStringList(items []string) *StringList {
    return &StringList{items: items}
}

func (sl *StringList) Count() int            { return len(sl.items) }
func (sl *StringList) Item(index int) string { return sl.items[index] }

type ListViewer struct {
    BaseView
    dataSource ListDataSource
    selected   int
    topIndex   int
    scrollBar  *ScrollBar
    OnSelect   func(int)
}

func NewListViewer(bounds Rect, dataSource ListDataSource) *ListViewer {
    lv := &ListViewer{dataSource: dataSource}
    lv.SetBounds(bounds)
    lv.SetState(SfVisible, true)
    lv.SetOptions(OfSelectable, true)
    return lv
}

func (lv *ListViewer) Selected() int              { return lv.selected }
func (lv *ListViewer) TopIndex() int              { return lv.topIndex }
func (lv *ListViewer) DataSource() ListDataSource { return lv.dataSource }

func (lv *ListViewer) SetSelected(index int) {
    count := lv.dataSource.Count()
    if count == 0 {
        lv.selected = 0
        return
    }
    if index < 0 {
        index = 0
    }
    if index >= count {
        index = count - 1
    }
    lv.selected = index
    lv.ensureVisible()
    lv.syncScrollBar()
}

func (lv *ListViewer) SetDataSource(ds ListDataSource) {
    lv.dataSource = ds
    lv.selected = 0
    lv.topIndex = 0
    lv.syncScrollBar()
}

func (lv *ListViewer) SetScrollBar(sb *ScrollBar) {
    lv.scrollBar = sb
    if sb != nil {
        sb.OnChange = func(val int) {
            lv.topIndex = val
            lv.clampTopIndex()
        }
        lv.syncScrollBar()
    }
}

func (lv *ListViewer) visibleHeight() int {
    return lv.Bounds().Height()
}

func (lv *ListViewer) ensureVisible() {
    vh := lv.visibleHeight()
    if vh <= 0 {
        return
    }
    if lv.selected < lv.topIndex {
        lv.topIndex = lv.selected
    }
    if lv.selected >= lv.topIndex+vh {
        lv.topIndex = lv.selected - vh + 1
    }
    lv.clampTopIndex()
}

func (lv *ListViewer) clampTopIndex() {
    count := lv.dataSource.Count()
    vh := lv.visibleHeight()
    maxTop := count - vh
    if maxTop < 0 {
        maxTop = 0
    }
    if lv.topIndex > maxTop {
        lv.topIndex = maxTop
    }
    if lv.topIndex < 0 {
        lv.topIndex = 0
    }
}

func (lv *ListViewer) syncScrollBar() {
    if lv.scrollBar == nil {
        return
    }
    count := lv.dataSource.Count()
    lv.scrollBar.SetRange(0, count)
    lv.scrollBar.SetPageSize(lv.visibleHeight())
    lv.scrollBar.SetValue(lv.topIndex)
}

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
    hasFocus := lv.HasState(SfSelected)

    for row := 0; row < vh; row++ {
        idx := lv.topIndex + row
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
            // Fill entire row with selection style
            for x := 0; x < w; x++ {
                buf.WriteChar(x, row, ' ', style)
            }
        }

        text := lv.dataSource.Item(idx)
        runes := []rune(text)
        if len(runes) > w {
            runes = runes[:w]
        }
        for i, ch := range runes {
            buf.WriteChar(i, row, ch, style)
        }
    }
}

func (lv *ListViewer) HandleEvent(event *Event) {
    if event.What == EvMouse && event.Mouse != nil {
        if event.Mouse.Button&tcell.Button1 != 0 {
            clickIdx := lv.topIndex + event.Mouse.Y
            if clickIdx >= 0 && clickIdx < lv.dataSource.Count() {
                lv.selected = clickIdx
                lv.ensureVisible()
                lv.syncScrollBar()
                if lv.OnSelect != nil {
                    lv.OnSelect(lv.selected)
                }
            }
            event.Clear()
        }
        return
    }

    if event.What != EvKeyboard || event.Key == nil {
        return
    }

    count := lv.dataSource.Count()
    if count == 0 {
        return
    }

    switch event.Key.Key {
    case tcell.KeyDown:
        if lv.selected < count-1 {
            lv.selected++
            lv.ensureVisible()
            lv.syncScrollBar()
            if lv.OnSelect != nil {
                lv.OnSelect(lv.selected)
            }
        }
        event.Clear()

    case tcell.KeyUp:
        if lv.selected > 0 {
            lv.selected--
            lv.ensureVisible()
            lv.syncScrollBar()
            if lv.OnSelect != nil {
                lv.OnSelect(lv.selected)
            }
        }
        event.Clear()

    case tcell.KeyHome:
        lv.selected = 0
        lv.topIndex = 0
        lv.syncScrollBar()
        if lv.OnSelect != nil {
            lv.OnSelect(lv.selected)
        }
        event.Clear()

    case tcell.KeyEnd:
        lv.selected = count - 1
        lv.ensureVisible()
        lv.syncScrollBar()
        if lv.OnSelect != nil {
            lv.OnSelect(lv.selected)
        }
        event.Clear()

    case tcell.KeyPgDn:
        lv.selected += lv.visibleHeight()
        if lv.selected >= count {
            lv.selected = count - 1
        }
        lv.ensureVisible()
        lv.syncScrollBar()
        if lv.OnSelect != nil {
            lv.OnSelect(lv.selected)
        }
        event.Clear()

    case tcell.KeyPgUp:
        lv.selected -= lv.visibleHeight()
        if lv.selected < 0 {
            lv.selected = 0
        }
        lv.ensureVisible()
        lv.syncScrollBar()
        if lv.OnSelect != nil {
            lv.OnSelect(lv.selected)
        }
        event.Clear()
    }
}
```

**Run tests:** `go test ./tv/... -run "TestListViewer|TestStringList" -v`

**Commit:** `git commit -m "feat(tv): add ListViewer widget with StringList data source and ScrollBar binding"`

---

### Task 5: Integration Checkpoint — ListViewer with ScrollBar in Window

**Purpose:** Verify ListViewer and ScrollBar work together inside a Window, with proper event routing, scroll synchronization, and rendering.

**Requirements (for test writer):**
- A ListViewer and ScrollBar inside a Window: clicking the ScrollBar's down arrow updates the ListViewer's topIndex
- Keyboard navigation on a focused ListViewer (Down, PgDn) updates the bound ScrollBar's value
- A ListViewer with more items than visible height shows correct items after scrolling via keyboard
- A ListViewer inside a Window with a custom ColorScheme renders items with that scheme's ListNormal/ListSelected/ListFocused styles
- Tab between a ListViewer and a Button inside the same Window changes focus correctly
- Mouse click on a visible ListViewer row inside a Window selects that item (coordinates route through Window's client area)

**Components to wire up:** Application (with SimulationScreen), Desktop, Window, ListViewer, ScrollBar, Button, StringList (all real, no mocks)

**Run:** `go test ./tv/... -run TestIntegrationPhase6List -v`

**Commit:** `git commit -m "test(tv): add Phase 6 integration tests for ListViewer with ScrollBar in Window"`

---

## Task 6: Demo App Update and E2E Tests

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`
- Test: `e2e/e2e_test.go`

**Requirements:**

**Demo app changes:**
- Add a ListViewer with a ScrollBar to win2 ("Editor" window)
- The ListViewer contains 20 items: "Item 1" through "Item 20"
- The ScrollBar is placed at the right edge of win2's client area (vertical, width=1)
- The ListViewer fills the remaining width
- win2 gets a different color scheme via `SetColorScheme` — use a custom scheme with green-on-black ListNormal, white-on-green ListSelected, yellow-on-green ListFocused (to make the difference visually obvious)
- The ListViewer is pre-focused so it responds to keyboard input immediately

**New E2E tests:**

1. `TestListViewerVisible` — After boot, win2 contains "Item 1" text visible in the window
2. `TestScrollBarVisible` — After boot, win2 contains scrollbar arrow characters (`▲` and `▼`)
3. `TestListViewerNavigation` — Click on win2 to focus it, press Down arrow multiple times, verify "Item 5" or later items become visible (scrolling works)
4. `TestListViewerDifferentTheme` — win2's list items render visibly different from win1's widgets (this is a visual smoke test — verify the list is visible and win2 doesn't use the same background as win1)

**E2E test approach:**
- Build binary via `buildBasicApp(t)`
- Launch in tmux session
- Send keystrokes via `tmuxSendKeys`
- Capture pane content via `tmuxCapture`
- Assert on visible text content

**Implementation guidance for demo app:**

Note: the code below uses `fmt.Sprintf` — ensure `"fmt"` is in the import block of `main.go`.

```go
// In main.go, modify win2 section:

// Create a custom scheme for win2
editorScheme := &theme.ColorScheme{}
*editorScheme = *theme.BorlandBlue // copy base
editorScheme.ListNormal = tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)
editorScheme.ListSelected = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorGreen)
editorScheme.ListFocused = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorGreen)
editorScheme.WindowBackground = tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)

win2 := tv.NewWindow(tv.NewRect(20, 5, 40, 12), "Editor", tv.WithWindowNumber(2))
win2.SetColorScheme(editorScheme)

// ListViewer fills client area minus scrollbar column
clientW := 40 - 2 // window width minus frame
clientH := 12 - 2 // window height minus frame

items := make([]string, 20)
for i := range items {
    items[i] = fmt.Sprintf("Item %d", i+1)
}

lv := tv.NewListViewer(tv.NewRect(0, 0, clientW-1, clientH), tv.NewStringList(items))
sb := tv.NewScrollBar(tv.NewRect(clientW-1, 0, 1, clientH), tv.Vertical)
lv.SetScrollBar(sb)

win2.Insert(lv)
win2.Insert(sb)
```

**Run tests:** `cd e2e && go test -v -timeout 120s`

**Commit:** `git commit -m "feat(e2e): add ListViewer with ScrollBar to demo app and e2e tests for Phase 6"`
