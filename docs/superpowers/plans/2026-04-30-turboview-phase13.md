# Phase 13: ListViewer Corrections — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Correct ListViewer behavior to match Turbo Vision: separate navigation from selection, add Space/Enter/double-click to select, fix Home/End semantics, and add mouse drag with auto-scroll (spec 7.1–7.5).

**Architecture:** All changes touch `tv/list_viewer.go`. The ListViewer already has keyboard handling (Up/Down/Home/End/PgUp/PgDn), single-click mouse handling, an OnSelect callback, and scrollbar integration. We remove OnSelect calls from navigation keys and single-click, add Space/Enter to fire OnSelect, add double-click detection using the existing `MouseEvent.ClickCount` field, change Home/End to page-local with Ctrl+PgUp/PgDn for absolute, and add mouse drag tracking with auto-scroll support.

**Tech Stack:** Go, tcell/v2, existing ListViewer/Event/ScrollBar APIs

**Critical:** The existing `list_viewer_test.go` has tests (Section 4) that assert OnSelect fires on arrow keys and single-click. These tests will FAIL after Phase 13 implementation — they test the OLD behavior. The implementer must update these tests to match the new spec (OnSelect must NOT fire on navigation keys or single-click). Similarly, Section 7 tests for Home/End assert absolute navigation (Home→item 0, End→last item) which changes to page-local navigation.

---

## File Structure

| File | Responsibility |
|------|---------------|
| `tv/list_viewer.go` | All ListViewer logic — navigation/selection separation, Space/Enter/double-click, Home/End semantics, mouse drag |

---

## Batch 1: Navigation vs Selection (Tasks 1–3)

### Task 1: Remove OnSelect from Navigation Keys (spec 7.1)

**Files:**
- Modify: `tv/list_viewer.go` (HandleEvent keyboard section)
- Modify: `tv/list_viewer_test.go` (Section 4 — update tests to expect OnSelect NOT firing on navigation)

**Requirements:**
- Arrow key navigation (Up, Down) changes the focused item and redraws but does NOT call OnSelect
- Page navigation (PgUp, PgDn) changes the focused item but does NOT call OnSelect
- Home and End change the focused item but do NOT call OnSelect
- All navigation keys still consume the event (Clear)
- All navigation keys still update selected, ensureVisible, and syncScrollBar
- OnSelect is nil-safe (no change from current behavior — just removing the calls from navigation)

**Implementation:**

Remove all `if lv.OnSelect != nil { lv.OnSelect(lv.selected) }` blocks from the keyboard switch cases for KeyDown, KeyUp, KeyHome, KeyEnd, KeyPgDn, KeyPgUp. The navigation logic (selected++, ensureVisible, syncScrollBar, event.Clear) remains unchanged.

```go
case tcell.KeyDown:
    if lv.selected < count-1 {
        lv.selected++
        lv.ensureVisible()
        lv.syncScrollBar()
    }
    event.Clear()

case tcell.KeyUp:
    if lv.selected > 0 {
        lv.selected--
        lv.ensureVisible()
        lv.syncScrollBar()
    }
    event.Clear()
```

Apply the same pattern to Home, End, PgUp, PgDn — remove OnSelect calls but keep everything else.

**Existing test updates required:** Tests in Section 4 (`TestOnSelectCalledOnKeyDown`, `TestOnSelectCalledOnKeyUp`, `TestOnSelectCalledOnHome`, `TestOnSelectCalledOnEnd`, `TestOnSelectCalledOnPgDn`, `TestOnSelectCalledOnPgUp`) must be inverted — they should assert OnSelect is NOT called. `TestOnSelectReceivesNewIndex` (which uses KeyDown to verify the index passed to OnSelect) must be rewritten to navigate with KeyDown then press Space, and verify OnSelect receives the navigated-to index. The mouse-click OnSelect tests (`TestOnSelectCalledOnMouseClick`, `TestOnSelectMouseClickReceivesClickedIndex`) also need updating — single-click should NOT fire OnSelect (see Task 3).

**Run tests:** `go test ./tv/... -run TestOnSelect -v`

**Commit:** `git commit -m "feat(listviewer): remove OnSelect from navigation keys (spec 7.1)"`

---

### Task 2: Space and Enter Key to Select (spec 7.2, 7.1)

**Files:**
- Modify: `tv/list_viewer.go` (HandleEvent keyboard section)

**Requirements:**
- When Space is pressed and a valid item is focused (count > 0, has SfSelected), call `OnSelect(selected)` and consume the event
- When Enter is pressed and a valid item is focused, call `OnSelect(selected)` and consume the event
- Space/Enter with nil OnSelect callback: no panic, still consume event
- Space/Enter with empty list (count == 0): no-op, do not call OnSelect, but still consume event
- Space/Enter when not focused (no SfSelected): not handled, event passes through
- The Enter key fires OnSelect on the ListViewer — if a Dialog intercepts Enter first via CmDefault broadcast, it never reaches the ListViewer. The ListViewer doesn't need to worry about this; it just handles Enter if it arrives.

**Implementation:**

Space and Enter must be handled BEFORE the `count == 0` early return so they consume the event even with an empty list, but must guard the OnSelect call on `count > 0`. Both are also after the existing `!lv.HasState(SfSelected)` early return, so the unfocused case is automatically covered.

For Space, tcell delivers it as `KeyRune` with Rune=' ', so add a check after the SfSelected guard but BEFORE the `count == 0` early return:

```go
// Space to select
if event.Key.Key == tcell.KeyRune && event.Key.Rune == ' ' {
    if count > 0 && lv.OnSelect != nil {
        lv.OnSelect(lv.selected)
    }
    event.Clear()
    return
}
```

For Enter, also add before the `count == 0` early return:

```go
// Enter to select
if event.Key.Key == tcell.KeyEnter {
    if count > 0 && lv.OnSelect != nil {
        lv.OnSelect(lv.selected)
    }
    event.Clear()
    return
}
```

The `count` variable must be declared before these checks. Move the `count := lv.dataSource.Count()` line up to before the Space/Enter checks.

The resulting order in HandleEvent's keyboard section is:
1. `if !lv.HasState(SfSelected) { return }` (existing)
2. `count := lv.dataSource.Count()` (moved up)
3. Space check (new — consumes event even if count==0, only calls OnSelect if count > 0)
4. Enter check (new — same pattern)
5. `if count == 0 { return }` (existing)
6. Switch statement for navigation keys (existing, minus OnSelect calls)

**Run tests:** `go test ./tv/... -run TestListViewer -v`

**Commit:** `git commit -m "feat(listviewer): add Space and Enter to fire OnSelect (spec 7.1, 7.2)"`

---

### Task 3: Single-Click Focus Only, Double-Click to Select (spec 7.3)

**Files:**
- Modify: `tv/list_viewer.go` (HandleEvent mouse section)
- Modify: `tv/list_viewer_test.go` (Section 8 — update tests for single-click not firing OnSelect)

**Requirements:**
- Single mouse click (Button1, ClickCount < 2) positions focus: sets selected to `topIndex + clickY`, calls ensureVisible, syncScrollBar, consumes event — but does NOT call OnSelect
- Double mouse click (Button1, ClickCount >= 2) positions focus AND calls OnSelect
- Click beyond data count (topIndex + clickY >= count): no change, no OnSelect, no panic
- Double-click beyond data count: no OnSelect, no panic
- ClickCount is already provided by the MouseEvent struct — the Application layer sets it from tcell
- Single-click still consumes the event (Clear)
- Double-click still consumes the event (Clear)

**Implementation:**

Update the mouse handling block in HandleEvent:

```go
if event.What == EvMouse && event.Mouse != nil {
    if event.Mouse.Button&tcell.Button1 != 0 {
        clickIdx := lv.topIndex + event.Mouse.Y
        if clickIdx >= 0 && clickIdx < lv.dataSource.Count() {
            lv.selected = clickIdx
            lv.ensureVisible()
            lv.syncScrollBar()
            if event.Mouse.ClickCount >= 2 && lv.OnSelect != nil {
                lv.OnSelect(lv.selected)
            }
        }
        event.Clear()
    }
    return
}
```

**Existing test updates required:** `TestOnSelectCalledOnMouseClick` should assert OnSelect is NOT called on single click. `TestOnSelectMouseClickReceivesClickedIndex` should be converted to a double-click test.

**Run tests:** `go test ./tv/... -run TestMouse -v`

**Commit:** `git commit -m "feat(listviewer): single-click focus only, double-click to select (spec 7.3)"`

---

### Task 4: Integration Checkpoint — Navigation vs Selection

**Purpose:** Verify that components from Tasks 1-3 work together correctly.

**Requirements (for test writer):**
- Navigating with arrow keys through a list does NOT fire OnSelect at any point during a multi-step navigation sequence
- After navigating to an item, pressing Space fires OnSelect with the correct index
- After navigating to an item, pressing Enter fires OnSelect with the correct index
- Single-clicking an item then pressing Space fires OnSelect for the clicked item (verifying click→focus + space→select pipeline)
- Double-clicking fires OnSelect exactly once (not twice — once for click focus, once for double-click)
- Navigating with PgUp/PgDn then double-clicking fires OnSelect for the clicked item, not the previously focused item
- A full sequence: navigate down 3 times → single-click item 1 → Space → verify OnSelect received index 1

**Components to wire up:** ListViewer with real ScrollBar, OnSelect callback tracking

**Run:** `go test ./tv/... -run TestIntegrationPhase13 -v`

---

## Batch 2: Home/End Semantics (Tasks 5–6)

### Task 5: Home/End Page-Local, Ctrl+PgUp/PgDn Absolute (spec 7.4)

**Files:**
- Modify: `tv/list_viewer.go` (HandleEvent keyboard section)
- Modify: `tv/list_viewer_test.go` (Section 7 — update Home/End tests for page-local behavior)

**Requirements:**
- Home: set selected to topIndex (first visible item on current page). Do NOT change topIndex.
- End: set selected to min(topIndex + visibleHeight - 1, count - 1) (last visible item on current page). Do NOT change topIndex.
- Ctrl+PgUp: set selected to 0 (absolute start), set topIndex to 0. This is the old Home behavior.
- Ctrl+PgDn: set selected to count - 1 (absolute end), call ensureVisible. This is the old End behavior.
- All four keys consume the event (Clear)
- All four keys call syncScrollBar after updating state
- None of these keys call OnSelect (they are navigation, not selection — per spec 7.1)
- Home when already at topIndex: no-op (selected doesn't change), event still consumed
- End when already at last visible item: no-op, event still consumed
- Ctrl+PgUp with modifier detection: `event.Key.Modifiers & tcell.ModCtrl != 0` combined with KeyPgUp
- Ctrl+PgDn with modifier detection: `event.Key.Modifiers & tcell.ModCtrl != 0` combined with KeyPgDn

**Implementation:**

Replace the Home/End cases and add Ctrl+PgUp/Ctrl+PgDn before PgUp/PgDn in the switch:

```go
case tcell.KeyHome:
    lv.selected = lv.topIndex
    lv.syncScrollBar()
    event.Clear()

case tcell.KeyEnd:
    lastVisible := lv.topIndex + lv.visibleHeight() - 1
    if lastVisible >= count {
        lastVisible = count - 1
    }
    lv.selected = lastVisible
    lv.syncScrollBar()
    event.Clear()
```

For Ctrl+PgUp/PgDn, check modifiers before the plain PgUp/PgDn cases. Since Go switch/case doesn't fall through by default, put the Ctrl variants first using a preliminary check:

```go
// Before the switch, check for Ctrl+PgUp/PgDn (these are after the count==0 early return)
ke := event.Key
if ke.Key == tcell.KeyPgUp && ke.Modifiers&tcell.ModCtrl != 0 {
    lv.selected = 0
    lv.topIndex = 0
    lv.syncScrollBar()
    event.Clear()
    return
}
if ke.Key == tcell.KeyPgDn && ke.Modifiers&tcell.ModCtrl != 0 {
    lv.selected = count - 1
    if lv.selected < 0 {
        lv.selected = 0
    }
    lv.ensureVisible()
    lv.syncScrollBar()
    event.Clear()
    return
}
```

Note: Both Ctrl+PgUp and Ctrl+PgDn checks are placed after the `count == 0` early return, so `count >= 1` is guaranteed and `count - 1 >= 0`. The defensive clamp `if lv.selected < 0` is belt-and-suspenders safety.

**Existing test updates required:**
- `TestKeyHomeSelectsFirstItem`: Must be rewritten with a scrolled list (topIndex > 0) to verify Home goes to topIndex, not item 0. Example: 10 items, height 5, scroll so topIndex=3, press Home, assert Selected()=3 (not 0).
- `TestKeyHomeScrollsToTop`: Must be rewritten to verify Home does NOT change topIndex. Example: set selected=7 to force topIndex=3, press Home, assert TopIndex() remains 3 (unchanged) and Selected() becomes 3.
- `TestKeyEndSelectsLastItem`: Must be rewritten to verify End goes to last visible item (topIndex + visibleHeight - 1), not the absolute last item. Example: 10 items, height 5, topIndex=0, press End, assert Selected()=4 (not 9).
- `TestKeyEndScrollsToShowLastItem`: Must be rewritten to verify End does NOT scroll. Example: 10 items, height 5, topIndex=0, press End, assert TopIndex() remains 0 and Selected()=4.
- New tests needed for Ctrl+PgUp (absolute start: selected=0, topIndex=0) and Ctrl+PgDn (absolute end: selected=count-1, ensureVisible adjusts topIndex).

**Run tests:** `go test ./tv/... -run "TestKeyHome|TestKeyEnd|TestKeyPg" -v`

**Commit:** `git commit -m "feat(listviewer): Home/End page-local, Ctrl+PgUp/PgDn absolute (spec 7.4)"`

---

### Task 6: Integration Checkpoint — Home/End with Scrolling

**Purpose:** Verify Home/End/Ctrl+PgUp/Ctrl+PgDn interact correctly with scrollbar and visible window.

**Requirements (for test writer):**
- With a scrolled list (topIndex > 0), pressing Home selects topIndex, not item 0
- With a scrolled list, pressing End selects the last visible item, not the last item in the list
- After pressing Home, pressing End moves to last visible item on the same page (topIndex unchanged)
- Ctrl+PgUp from a scrolled position resets both selected and topIndex to 0
- Ctrl+PgDn from the top moves selected to last item and scrolls to show it
- After Ctrl+PgDn, pressing Home selects topIndex (which is now scrolled to show the last page)
- ScrollBar value stays in sync with topIndex after all operations
- After Ctrl+PgUp, scrollBar value is 0
- After Ctrl+PgDn, scrollBar value equals the new topIndex

**Components to wire up:** ListViewer with real ScrollBar (20 items, height 5)

**Run:** `go test ./tv/... -run TestIntegrationPhase13 -v`

---

## Batch 3: Mouse Drag (Tasks 7–8)

### Task 7: Mouse Drag with Auto-Scroll (spec 7.5)

**Files:**
- Modify: `tv/list_viewer.go` (HandleEvent mouse section)

**Requirements:**
- When Button1 is pressed and held, track the drag state: set `dragging = true`
- On subsequent Button1 mouse events while dragging, update selected to `topIndex + mouseY`, clamped to valid range [0, count-1]
- If mouseY < 0 (above the widget), auto-scroll up: decrease topIndex by 1 (if topIndex > 0), set selected = topIndex
- If mouseY >= visibleHeight (below the widget), auto-scroll down: increase topIndex by 1 (if more items below), set selected = topIndex + visibleHeight - 1 (clamped to count-1)
- On Button1 release (mouse event with Button1 not set while dragging): set dragging = false
- During drag, call ensureVisible and syncScrollBar after each update
- During drag, consume each mouse event (Clear)
- Drag does NOT fire OnSelect — it's navigation, not selection (per spec 7.1)
- Mouse auto-repeat (evMouseAuto) is handled by the Application layer — it posts synthetic EvMouse events with Button1 while the button is held. The ListViewer just processes each EvMouse event normally; if mouseY is outside bounds, it auto-scrolls.
- The `dragging` field prevents the first-click from double-counting: first Button1 press sets selected AND dragging=true, subsequent Button1 events (from evMouseAuto or real mouse move) just update selected.
- Auto-scroll boundary: when mouseY < 0, scroll up one item per event. When mouseY >= visibleHeight, scroll down one item per event.

**Implementation:**

Add a `dragging` field to ListViewer:

```go
type ListViewer struct {
    BaseView
    dataSource ListDataSource
    selected   int
    topIndex   int
    scrollBar  *ScrollBar
    dragging   bool
    OnSelect   func(int)
}
```

Rework the mouse handling in HandleEvent:

```go
if event.What == EvMouse && event.Mouse != nil {
    if event.Mouse.Button&tcell.Button1 != 0 {
        count := lv.dataSource.Count()
        my := event.Mouse.Y

        if my < 0 {
            // Auto-scroll up
            if lv.topIndex > 0 {
                lv.topIndex--
                lv.selected = lv.topIndex
            }
        } else if my >= lv.visibleHeight() {
            // Auto-scroll down
            maxTop := count - lv.visibleHeight()
            if maxTop < 0 {
                maxTop = 0
            }
            if lv.topIndex < maxTop {
                lv.topIndex++
            }
            lastVisible := lv.topIndex + lv.visibleHeight() - 1
            if lastVisible >= count {
                lastVisible = count - 1
            }
            lv.selected = lastVisible
        } else {
            // Normal click/drag within bounds
            clickIdx := lv.topIndex + my
            if clickIdx >= 0 && clickIdx < count {
                lv.selected = clickIdx
            }
        }

        if !lv.dragging {
            lv.dragging = true
        }

        lv.ensureVisible()
        lv.syncScrollBar()

        // Double-click fires OnSelect (only on normal in-bounds click, not during drag-scroll)
        if event.Mouse.ClickCount >= 2 && my >= 0 && my < lv.visibleHeight() && lv.OnSelect != nil {
            lv.OnSelect(lv.selected)
        }

        event.Clear()
    } else if lv.dragging {
        lv.dragging = false
        event.Clear()
    }
    return
}
```

**Run tests:** `go test ./tv/... -run TestListViewer -v`

**Commit:** `git commit -m "feat(listviewer): mouse drag with auto-scroll (spec 7.5)"`

---

### Task 8: Integration Checkpoint — Mouse Drag and Auto-Scroll

**Purpose:** Verify mouse drag and auto-scroll work with the full ListViewer + ScrollBar chain.

**Requirements (for test writer):**
- Pressing Button1 on row 2, then moving to row 4 (Button1 still held) updates selected from topIndex+2 to topIndex+4
- Dragging above the widget (mouseY = -1) with topIndex > 0 scrolls up and selects the new topIndex
- Dragging below the widget (mouseY = visibleHeight) with items below scrolls down and selects the last visible item
- Releasing Button1 after a drag stops tracking (subsequent mouse move does not update selection)
- ScrollBar value stays in sync during the entire drag sequence
- Auto-scroll up from topIndex=3 with mouseY=-1: topIndex becomes 2, selected becomes 2
- Auto-scroll down stops when topIndex reaches maxTop (no over-scroll)
- Drag does not fire OnSelect at any point during the drag sequence
- Double-click during a drag (ClickCount >= 2) fires OnSelect for the in-bounds item

**Components to wire up:** ListViewer with real ScrollBar (20 items, height 5)

**Run:** `go test ./tv/... -run TestIntegrationPhase13 -v`

---

## Batch 4: E2E Test (Task 9)

### Task 9: E2E Test — ListViewer Selection Behavior

**Files:**
- Modify: `e2e/e2e_test.go`

**Requirements:**
- Extend the e2e test suite with a test that verifies the ListViewer selection behavior in the running app
- Navigate to win2 (which has the ListViewer), use arrow keys to navigate without triggering selection, then verify the app is still responsive
- Use Down arrow to move through items, verifying visual update (different items become visible)
- All existing e2e tests must continue to pass (regression)

**Implementation:**

Add a new test to `e2e/e2e_test.go`:

```go
func TestListViewerSpaceSelect(t *testing.T) {
    binPath := buildBasicApp(t)
    session := "tv3-e2e-listspace"
    exec.Command("tmux", "kill-session", "-t", session).Run()
    startTmux(t, session, binPath)

    // Navigate to win2's list viewer
    tmuxSendKeys(t, session, "Tab")
    time.Sleep(500 * time.Millisecond)

    // Arrow down a few times (navigation only, no selection)
    for i := 0; i < 3; i++ {
        tmuxSendKeys(t, session, "Down")
    }
    time.Sleep(500 * time.Millisecond)

    lines := tmuxCapture(t, session)

    // Item 4 should be focused (started at Item 1, moved down 3)
    if !containsAny(lines, "Item 4") {
        t.Error("Item 4 not visible after navigating down 3 times")
    }

    // App still responsive — clean exit
    tmuxSendKeys(t, session, "M-x")
    exited := false
    for i := 0; i < 15; i++ {
        if !tmuxHasSession(session) {
            exited = true
            break
        }
        time.Sleep(200 * time.Millisecond)
    }
    if !exited {
        t.Error("app did not exit after Alt+X")
    }
}
```

**Run tests:** `cd e2e && go test -v -timeout 120s`

**Commit:** `git commit -m "test(e2e): add ListViewer navigation and selection e2e test"`
