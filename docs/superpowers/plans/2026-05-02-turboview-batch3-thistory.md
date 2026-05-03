# THistory Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add THistory widget — a non-focusable companion to InputLine that stores and presents a dropdown history list.

**Architecture:** THistory is a plain widget (not a Container) placed next to an InputLine. It renders a `▐↓▌` icon. Clicking the icon or pressing Down while the linked InputLine is focused opens a modal popup with history entries. The popup uses a custom modal loop (like MenuBar) with `app.PollEvent()` — not `ExecView` — to support outside-click dismissal. A global `HistoryStore` manages entries by numeric ID with deduplication and eviction. Button presses broadcast `CmRecordHistory` so all History views record their linked InputLine's text.

**Tech Stack:** Go, tcell/v2, existing tv package (BaseView, Group, ListViewer, ScrollBar, InputLine, Application)

---

## File Structure

| File | Responsibility |
|------|---------------|
| `tv/history_store.go` | HistoryStore type, Add/Entries/Clear, global DefaultHistory |
| `tv/history.go` | History widget: constructor, Draw, HandleEvent, dropdown logic |
| `tv/command.go` | Add CmRecordHistory constant |
| `tv/button.go` | Broadcast CmRecordHistory in press() |
| `tv/input_line.go` | Add SelectAll() method |
| `theme/scheme.go` | Add HistoryArrow, HistorySides fields |
| `theme/borland.go` | Add History colors to BorlandBlue |
| `theme/borland_cyan.go` | Add History colors |
| `theme/borland_gray.go` | Add History colors |
| `theme/c64.go` | Add History colors |
| `theme/matrix.go` | Add History colors |
| `e2e/testapp/basic/main.go` | Add History next to InputLine in win1 |
| `tv/desktop.go` | Add `App()` accessor if not already exported |
| `e2e/e2e_test.go` | E2E tests for History |

---

## Tasks

- [ ] Task 1: HistoryStore
- [ ] Task 2: Color Scheme, Command Constants, and InputLine.SelectAll
- [ ] Task 3: History Widget — Draw and Constructor
- [ ] Task 4: History Widget — Event Handling and Dropdown
- [ ] Task 5: Button CmRecordHistory Broadcast
- [ ] Task 6: Integration Checkpoint
- [ ] Task 7: E2E Test

---

### Task 1: HistoryStore

**Files:**
- Create: `tv/history_store.go`
- Test: `tv/history_store_test.go`

**Requirements:**

**Constructor:**
- `NewHistoryStore(maxPerID int)` creates a store with the given max entries per ID
- `var DefaultHistory = NewHistoryStore(20)` is the package-level default

**Add behavior:**
- `Add(id int, s string)` adds an entry for the given history ID
- Empty strings are never stored (no-op)
- If the string equals the most recent entry for this ID, it is not added (dedup at tail)
- If a duplicate exists at any other position, that older entry is removed before adding the new one at the end
- If the entry count exceeds `maxPerID`, the oldest entry is evicted
- Adding to an ID that doesn't exist yet creates the list

**Entries behavior:**
- `Entries(id int)` returns entries in chronological order (oldest first, newest last)
- For an unknown ID, returns nil (or empty slice)

**Clear behavior:**
- `Clear()` removes all entries for all IDs

**Implementation:**

```go
package tv

type HistoryStore struct {
    maxPerID int
    entries  map[int][]string
}

func NewHistoryStore(maxPerID int) *HistoryStore {
    return &HistoryStore{
        maxPerID: maxPerID,
        entries:  make(map[int][]string),
    }
}

var DefaultHistory = NewHistoryStore(20)

func (hs *HistoryStore) Add(id int, s string) {
    if s == "" {
        return
    }
    list := hs.entries[id]
    if len(list) > 0 && list[len(list)-1] == s {
        return
    }
    for i := 0; i < len(list); i++ {
        if list[i] == s {
            list = append(list[:i], list[i+1:]...)
            break
        }
    }
    list = append(list, s)
    if len(list) > hs.maxPerID {
        list = list[len(list)-hs.maxPerID:]
    }
    hs.entries[id] = list
}

func (hs *HistoryStore) Entries(id int) []string {
    return hs.entries[id]
}

func (hs *HistoryStore) Clear() {
    hs.entries = make(map[int][]string)
}
```

**Run tests:** `go test ./tv/ -run TestHistoryStore -v`

**Commit:** `git commit -m "feat: add HistoryStore for THistory entry management"`

---

### Task 2: Color Scheme, Command Constants, and InputLine.SelectAll

**Files:**
- Modify: `theme/scheme.go`
- Modify: `theme/borland.go`
- Modify: `theme/borland_cyan.go`
- Modify: `theme/borland_gray.go`
- Modify: `theme/c64.go`
- Modify: `theme/matrix.go`
- Modify: `tv/command.go`
- Modify: `tv/input_line.go`
- Test: `tv/history_colors_test.go`

**Requirements:**

**Color scheme:**
- Add `HistoryArrow tcell.Style` to `ColorScheme` struct — color for the `↓` character
- Add `HistorySides tcell.Style` to `ColorScheme` struct — color for the `▐` and `▌` bracket characters
- All five themes must include values for these new fields

**Command constant:**
- Add `CmRecordHistory` to the iota block in `tv/command.go`, after `CmSelectWindowNum`

**InputLine.SelectAll:**
- Add `SelectAll()` method to InputLine that selects all text: sets `selStart = 0`, `selEnd = len(il.text)`, `cursorPos = len(il.text)`
- This is needed by History's dropdown to select all text after pasting the selected entry

**Theme colors (reference Borland TV defaults):**
- BorlandBlue: arrow = green on cyan, sides = cyan on blue
- BorlandCyan: arrow = dark cyan on white, sides = white on cyan
- BorlandGray: arrow = dark gray on white, sides = white on dark gray
- C64: arrow = light blue on blue, sides = blue on light blue
- Matrix: arrow = white on dark green, sides = green on black

**Implementation:**

Add to `theme/scheme.go` after `MemoSelected`:
```go
HistoryArrow tcell.Style
HistorySides tcell.Style
```

Add to `tv/command.go` after `CmSelectWindowNum`:
```go
CmRecordHistory
```

Add to `tv/input_line.go`:
```go
func (il *InputLine) SelectAll() {
    il.selStart = 0
    il.selEnd = len(il.text)
    il.cursorPos = len(il.text)
}
```

Add to each theme file, the two new fields with appropriate colors.

**Run tests:** `go test ./tv/ -run "TestHistoryColor|TestInputLine" -v && go test ./theme/... -v`

**Commit:** `git commit -m "feat: add HistoryArrow/HistorySides colors, CmRecordHistory, InputLine.SelectAll"`

---

### Task 3: History Widget — Draw and Constructor

**Files:**
- Create: `tv/history.go`
- Test: `tv/history_test.go`

**Requirements:**

**Constructor:**
- `NewHistory(bounds Rect, link *InputLine, historyID int) *History`
- Sets `OfPostProcess` (sees events after focused view)
- Does NOT set `OfSelectable` (never receives focus)
- Sets `SfVisible`
- Stores reference to linked InputLine and history ID

**History struct:**
- Fields: `link *InputLine`, `historyID int`
- Embeds `BaseView`

**Draw:**
- Renders exactly 3 characters: `▐↓▌`
- `▐` (U+2590) at x=0 in `HistorySides` color
- `↓` (U+2193) at x=1 in `HistoryArrow` color
- `▌` (U+258C) at x=2 in `HistorySides` color
- If no color scheme is available, use `tcell.StyleDefault`

**Accessors:**
- `Link() *InputLine` returns the linked InputLine
- `HistoryID() int` returns the history ID

**Implementation:**

```go
package tv

import "github.com/gdamore/tcell/v2"

var _ Widget = (*History)(nil)

type History struct {
    BaseView
    link      *InputLine
    historyID int
}

func NewHistory(bounds Rect, link *InputLine, historyID int) *History {
    h := &History{
        link:      link,
        historyID: historyID,
    }
    h.SetBounds(bounds)
    h.SetState(SfVisible, true)
    h.SetOptions(OfPostProcess, true)
    h.SetSelf(h)
    return h
}

func (h *History) Link() *InputLine { return h.link }
func (h *History) HistoryID() int   { return h.historyID }

func (h *History) Draw(buf *DrawBuffer) {
    cs := h.ColorScheme()
    sidesStyle := tcell.StyleDefault
    arrowStyle := tcell.StyleDefault
    if cs != nil {
        sidesStyle = cs.HistorySides
        arrowStyle = cs.HistoryArrow
    }
    buf.WriteChar(0, 0, '▐', sidesStyle)
    buf.WriteChar(1, 0, '↓', arrowStyle)
    buf.WriteChar(2, 0, '▌', sidesStyle)
}
```

**Run tests:** `go test ./tv/ -run TestHistory -v`

**Commit:** `git commit -m "feat: add History widget with constructor and Draw"`

---

### Task 4: History Widget — Event Handling and Dropdown

**Files:**
- Modify: `tv/history.go`
- Test: `tv/history_event_test.go`

**Requirements:**

**Mouse click handling:**
- Any click (`Button1`) on the History icon:
  1. Attempt to focus the linked InputLine: `h.link.Owner().SetFocusedChild(h.link)`
  2. If the link does not have `OfSelectable`, clear the event and return
  3. Otherwise, open the history dropdown
- Non-Button1 mouse events: ignored (not consumed)

**Keyboard handling (PostProcess):**
- Down arrow key, only when `h.link.HasState(SfSelected)`: open the history dropdown
- All other keys: not consumed (return without clearing event)

**Broadcast handling:**
- `CmReleasedFocus` with `event.Info == h.link`: record InputLine text via `DefaultHistory.Add(h.historyID, h.link.Text())`
- `CmRecordHistory`: record InputLine text via `DefaultHistory.Add(h.historyID, h.link.Text())`
- All other broadcasts: ignored

**Dropdown behavior (openDropdown method):**
1. Record current InputLine text: `DefaultHistory.Add(h.historyID, h.link.Text())`
2. Get entries: `DefaultHistory.Entries(h.historyID)` — if empty, return
3. Reverse entries for display (most recent at top)
4. Calculate popup bounds:
   - Width: InputLine width + 2 (frame margins)
   - Height: min(len(entries) + 2, 9) — at most 7 visible items plus frame
   - Position: convert InputLine bounds to desktop-absolute coords via `viewToDesktop()`. Popup top-left at `(inputAbsX - 1, inputAbsY)` so the frame aligns with the InputLine's left edge
5. Create a `historyPopup` (custom view with ListViewer + ScrollBar + frame drawing), insert into Desktop, propagate ColorScheme to internal ListViewer/ScrollBar (they have no Owner so can't walk up the chain)
6. Run a custom modal loop (like MenuBar): call `app.PollEvent()` in a loop, route events, handle Enter/Escape/outside-clicks directly
7. **Mouse coordinate conversion:** `PollEvent()` returns screen-absolute mouse coords. Convert to Desktop-local by subtracting `desktop.Bounds().A` before comparing against popup bounds (which are Desktop-local). This is critical when MenuBar is present (Desktop starts at y=1).
8. On confirm: `h.link.SetText(selectedEntry)` then `h.link.SelectAll()`
9. Clear the event after opening dropdown

**Custom modal loop (not ExecView):**

The dropdown uses its own modal loop, like MenuBar does, to support outside-click dismissal. This avoids ExecView which discards outside clicks silently.

The popup is a View temporarily inserted into the Desktop (so the normal draw pipeline picks it up). The modal loop calls `app.PollEvent()` with its own event routing, then removes the popup when done.

```go
func (h *History) openDropdown(event *Event) {
    DefaultHistory.Add(h.historyID, h.link.Text())
    entries := DefaultHistory.Entries(h.historyID)
    if len(entries) == 0 {
        return
    }

    reversed := make([]string, len(entries))
    for i, e := range entries {
        reversed[len(entries)-1-i] = e
    }

    app := findApp(h)
    desktop := findDesktop(h)
    if app == nil || desktop == nil {
        return
    }

    linkAbsX, linkAbsY := viewToDesktop(h.link)
    popupW := h.link.Bounds().Width() + 2
    popupH := len(reversed) + 2
    if popupH > 9 {
        popupH = 9
    }
    popupX := linkAbsX - 1
    popupY := linkAbsY

    popup := newHistoryPopup(NewRect(popupX, popupY, popupW, popupH), reversed)
    desktop.Insert(popup)
    popup.propagateScheme()
    app.drawAndFlush()

    result := h.runPopupLoop(app, popup)

    desktop.Remove(popup)

    if result == CmOK {
        sel := popup.viewer.Selected()
        if sel >= 0 && sel < len(reversed) {
            h.link.SetText(reversed[sel])
            h.link.SelectAll()
        }
    }

    app.drawAndFlush()
    event.Clear()
}

func (h *History) runPopupLoop(app *Application, popup *historyPopup) CommandCode {
    desktop := findDesktop(h)
    desktopOrigin := desktop.Bounds().A // screen offset of Desktop (e.g., y=1 when MenuBar present)
    for {
        ev := app.PollEvent()
        if ev == nil {
            return CmCancel
        }

        if ev.What == EvKeyboard && ev.Key != nil {
            switch ev.Key.Key {
            case tcell.KeyEnter:
                return CmOK
            case tcell.KeyEscape:
                return CmCancel
            default:
                popup.viewer.HandleEvent(ev)
            }
        } else if ev.What == EvMouse && ev.Mouse != nil {
            pb := popup.Bounds()
            // PollEvent mouse coords are screen-absolute; convert to Desktop-local
            mx := ev.Mouse.X - desktopOrigin.X
            my := ev.Mouse.Y - desktopOrigin.Y
            if pb.Contains(NewPoint(mx, my)) {
                ev.Mouse.X = mx - pb.A.X - 1
                ev.Mouse.Y = my - pb.A.Y - 1
                popup.viewer.HandleEvent(ev)
                if popup.confirmed {
                    return CmOK
                }
            } else if ev.Mouse.Button&tcell.Button1 != 0 {
                return CmCancel
            }
        }

        app.drawAndFlush()
    }
}
```

**historyPopup type** (private, defined in `tv/history.go`):

```go
type historyPopup struct {
    BaseView
    viewer    *ListViewer
    scrollbar *ScrollBar
    confirmed bool
}

func newHistoryPopup(bounds Rect, entries []string) *historyPopup {
    hp := &historyPopup{}
    hp.SetBounds(bounds)
    hp.SetState(SfVisible, true)

    clientW := bounds.Width() - 2
    clientH := bounds.Height() - 2

    needsScroll := clientH < len(entries)
    viewerW := clientW
    if needsScroll {
        viewerW = clientW - 1
    }

    ds := NewStringList(entries)
    hp.viewer = NewListViewer(NewRect(0, 0, viewerW, clientH), ds)
    hp.viewer.SetState(SfVisible, true)
    hp.viewer.SetOptions(OfSelectable, true)
    hp.viewer.SetState(SfSelected, true)
    hp.viewer.OnSelect = func(int) { hp.confirmed = true }

    if needsScroll {
        hp.scrollbar = NewScrollBar(NewRect(clientW-1, 0, 1, clientH), Vertical)
        hp.viewer.SetScrollBar(hp.scrollbar)
    }

    hp.SetSelf(hp)
    return hp
}

// propagateScheme passes the popup's ColorScheme to internal components
// that are not inserted into a Container (so have no Owner to walk up to).
// Must be called after the popup is inserted into the Desktop.
// NOTE: BaseView has an unexported `scheme` field but no SetColorScheme.
// Window has SetColorScheme. For ListViewer/ScrollBar (which embed BaseView),
// the implementer should either: (a) add a SetColorScheme to BaseView, or
// (b) set the field directly since history.go is in the same package.
func (hp *historyPopup) propagateScheme() {
    cs := hp.ColorScheme()
    if cs != nil {
        hp.viewer.scheme = cs
        if hp.scrollbar != nil {
            hp.scrollbar.scheme = cs
        }
    }
}

func (hp *historyPopup) Draw(buf *DrawBuffer) {
    cs := hp.ColorScheme()
    frameStyle := tcell.StyleDefault
    bgStyle := tcell.StyleDefault
    if cs != nil {
        frameStyle = cs.WindowFrameActive
        bgStyle = cs.WindowBackground
    }

    w := hp.Bounds().Width()
    h := hp.Bounds().Height()

    // Fill background
    buf.Fill(NewRect(0, 0, w, h), ' ', bgStyle)

    // Single-line frame
    buf.WriteChar(0, 0, '┌', frameStyle)
    buf.WriteChar(w-1, 0, '┐', frameStyle)
    buf.WriteChar(0, h-1, '└', frameStyle)
    buf.WriteChar(w-1, h-1, '┘', frameStyle)
    for x := 1; x < w-1; x++ {
        buf.WriteChar(x, 0, '─', frameStyle)
        buf.WriteChar(x, h-1, '─', frameStyle)
    }
    for y := 1; y < h-1; y++ {
        buf.WriteChar(0, y, '│', frameStyle)
        buf.WriteChar(w-1, y, '│', frameStyle)
    }

    // Draw ListViewer into client area (narrower when scrollbar present)
    viewerW := w - 2
    if hp.scrollbar != nil {
        viewerW = w - 3
    }
    clientBuf := buf.SubBuffer(NewRect(1, 1, viewerW, h-2))
    hp.viewer.Draw(clientBuf)

    // Draw scrollbar if present
    if hp.scrollbar != nil {
        sbBuf := buf.SubBuffer(NewRect(w-2, 1, 1, h-2))
        hp.scrollbar.Draw(sbBuf)
    }
}
```

**Helper functions** (in `tv/history.go`):

```go
// viewToDesktop converts a View's position to Desktop-local coordinates.
// Walks the owner chain, accumulating bounds offsets and Window frame offsets.
// Stops at the Desktop level (does NOT include Desktop.Bounds().A) because
// the popup is a Desktop child and its bounds must be Desktop-local.
func viewToDesktop(v View) (int, int) {
    x, y := v.Bounds().A.X, v.Bounds().A.Y
    owner := v.Owner()
    for owner != nil {
        if _, isDesktop := owner.(*Desktop); isDesktop {
            break
        }
        if view, ok := owner.(View); ok {
            b := view.Bounds()
            x += b.A.X
            y += b.A.Y
            // Window has a 1-cell frame; client area starts at (1,1)
            if _, isWindow := owner.(*Window); isWindow {
                x += 1
                y += 1
            }
            owner = view.Owner()
        } else {
            break
        }
    }
    return x, y
}

func findDesktop(v View) *Desktop {
    var current Container = v.Owner()
    for current != nil {
        if d, ok := current.(*Desktop); ok {
            return d
        }
        if view, ok := current.(View); ok {
            current = view.Owner()
        } else {
            break
        }
    }
    return nil
}

func findApp(v View) *Application {
    d := findDesktop(v)
    if d != nil {
        return d.App()
    }
    return nil
}
```

**NOTE:** `Desktop.App()` may not exist yet. If so, add a public accessor:
```go
func (d *Desktop) App() *Application { return d.app }
```

Also, `Application.PollEvent()` and `Application.drawAndFlush()` may be unexported. Check their visibility. If `drawAndFlush` is unexported, it must be used from the same package (`tv`), which is the case since History is in the `tv` package. If `PollEvent` is unexported, add a public accessor. Check `tv/application.go` for current export status.

**Implementation:**

The HandleEvent method on History:

```go
func (h *History) HandleEvent(event *Event) {
    // Broadcast handling
    if event.What == EvBroadcast {
        switch event.Command {
        case CmReleasedFocus:
            if event.Info == h.link {
                DefaultHistory.Add(h.historyID, h.link.Text())
            }
        case CmRecordHistory:
            DefaultHistory.Add(h.historyID, h.link.Text())
        }
        return
    }

    // Mouse click
    if event.What == EvMouse && event.Mouse != nil {
        if event.Mouse.Button&tcell.Button1 != 0 {
            if h.link.Owner() != nil {
                h.link.Owner().SetFocusedChild(h.link)
            }
            if !h.link.HasOption(OfSelectable) {
                event.Clear()
                return
            }
            h.openDropdown(event)
            return
        }
        return
    }

    // Keyboard (PostProcess) — Down arrow only when link is focused
    if event.What == EvKeyboard && event.Key != nil {
        if event.Key.Key == tcell.KeyDown && h.link.HasState(SfSelected) {
            h.openDropdown(event)
            return
        }
    }
}
```

**Run tests:** `go test ./tv/ -run TestHistory -v`

**Commit:** `git commit -m "feat: History event handling and dropdown popup"`

---

### Task 5: Button CmRecordHistory Broadcast

**Files:**
- Modify: `tv/button.go`
- Test: `tv/history_button_test.go`

**Requirements:**
- When `Button.press()` fires, it broadcasts `CmRecordHistory` to its owner BEFORE transforming the event to the button's command
- This matches original TV where `TButton::press()` broadcasts `cmRecordHistory`
- The broadcast causes all History views in the dialog to record their linked InputLine contents

**Implementation:**

```go
func (b *Button) press(event *Event) {
    b.broadcastToOwner(CmRecordHistory)
    event.What = EvCommand
    event.Command = b.command
    event.Key = nil
    event.Mouse = nil
}
```

**Run tests:** `go test ./tv/ -run TestHistoryButton -v`

**Commit:** `git commit -m "feat: Button broadcasts CmRecordHistory on press"`

---

### Task 6: Integration Checkpoint — THistory

**Purpose:** Verify that Tasks 1-5 work together: HistoryStore + History widget + InputLine + Button + broadcast chain.

**Files:**
- Create: `tv/integration_batch3_history_test.go`

**Requirements (for test writer):**
- HistoryStore Add then Entries round-trip: add multiple entries for same ID, verify order and dedup
- History widget draws `▐↓▌` with correct characters at correct positions
- History broadcast handler: simulate CmReleasedFocus with link as Info, verify entry recorded in DefaultHistory
- History broadcast handler: simulate CmRecordHistory, verify entry recorded in DefaultHistory
- Button press broadcasts CmRecordHistory, which causes a History view in the same group to record its InputLine's text
- History Down arrow only opens dropdown when link has SfSelected (not when link is unfocused)
- History mouse click focuses the linked InputLine

**Components to wire up:** HistoryStore (real), History (real), InputLine (real), Button (real), Group/Window (real)

**Run:** `go test ./tv/ -run TestIntegration -v`

**Commit:** `git commit -m "test: integration tests for THistory"`

---

### Task 7: E2E Test — THistory in Demo App

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**

**Demo app changes:**
Add a History widget next to the existing InputLine in win1:

```go
history := tv.NewHistory(tv.NewRect(31, 12, 3, 1), inputLine, 1)
win1.Insert(history)
```

**E2E tests:**
- `TestHistoryIconVisible`: Focus win1 (Alt+1), verify `↓` character is visible on screen
- `TestHistoryDropdown`: Focus win1 (Alt+1), Tab to the InputLine, type "test1", Tab away (triggers CmReleasedFocus recording), Tab back to InputLine, press Down arrow (opens history dropdown), verify dropdown appears with "test1" visible, press Escape to dismiss, verify app still running
- Existing e2e tests continue to pass

**Run tests:** `go test ./e2e/ -v -timeout 300s`

**Commit:** `git commit -m "feat: add History to demo app, e2e tests for THistory"`
