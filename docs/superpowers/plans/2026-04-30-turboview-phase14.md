# Phase 14: Window Commands and Architectural Corrections — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add CmZoom/CmResize command handling to Window, add CmSelectWindowNum broadcast, and correct the responsibility distribution: move CmTile/CmCascade to Application, move Alt+N to Application with broadcast, move Tab from Group to Window.

**Architecture:** Changes span `tv/window.go`, `tv/command.go`, `tv/application.go`, `tv/desktop.go`, and `tv/group.go`. Each change is a focused responsibility move — the behavior already works, we're correcting where it lives. Dependencies: 11.2 (CmSelectWindowNum) must be done before 13.2 (Alt+N broadcasts it), and 13.3 (Tab in Window) requires removing Tab from Group.

**Tech Stack:** Go, tcell/v2, existing Window/Desktop/Group/Application APIs

**Critical:** Sections 13.1-13.3 are architectural moves — behavior must remain identical. The e2e tests for tile, cascade, Alt+N, and Tab navigation are the safety net. All 25 existing e2e tests must continue to pass after these changes.

---

## File Structure

| File | Responsibility |
|------|---------------|
| `tv/window.go` | Add CmZoom/CmResize command handling, CmSelectWindowNum broadcast response, Tab/Shift+Tab handling |
| `tv/command.go` | Add CmSelectWindowNum constant |
| `tv/application.go` | Move CmTile/CmCascade handling here, move Alt+N here with CmSelectWindowNum broadcast |
| `tv/desktop.go` | Remove CmTile/CmCascade, remove Alt+N, remove Tab forwarding |
| `tv/group.go` | Remove Tab/Shift+Tab handling |

---

## Batch 1: Window Commands (Tasks 1–3)

### Task 1: CmZoom and CmResize in Window (spec 11.1)

**Files:**
- Modify: `tv/window.go` (HandleEvent)

**Requirements:**
- Window.HandleEvent handles EvCommand with CmZoom: call `w.Zoom()`, consume event
- Window.HandleEvent handles EvCommand with CmResize: consume event (keyboard-driven resize is deferred — spec 11.1 is partially implemented here, as the resize UI does not exist yet)
- CmZoom and CmResize handling must be BEFORE the group dispatch (`w.group.HandleEvent(event)`) so the window gets first chance
- Existing CmClose→CmCancel handling for modal windows must be preserved
- **Spec deviation:** The spec mentions checking `wfZoom` capability. No window flags system exists in this codebase — all windows can zoom. The capability check is deferred.

**Implementation:**

In `Window.HandleEvent`, after the mouse check and the modal CmClose check, add:

```go
if event.What == EvCommand {
    switch event.Command {
    case CmZoom:
        w.Zoom()
        event.Clear()
        return
    case CmResize:
        event.Clear()
        return
    }
}

w.group.HandleEvent(event)
```

The existing structure is:
```go
func (w *Window) HandleEvent(event *Event) {
    if event.What == EvMouse && event.Mouse != nil {
        w.handleMouseEvent(event)
        return
    }
    // Modal: CmClose → CmCancel
    if event.What == EvCommand && event.Command == CmClose && w.HasState(SfModal) {
        event.Command = CmCancel
        return
    }
    w.group.HandleEvent(event)
}
```

Insert the CmZoom/CmResize handling after the CmClose→CmCancel block and before the group dispatch.

**Run tests:** `go test ./tv/... -run TestWindow -v`

**Commit:** `git commit -m "feat(window): handle CmZoom and CmResize commands (spec 11.1)"`

---

### Task 2: CmSelectWindowNum Broadcast (spec 11.2)

**Files:**
- Modify: `tv/command.go` (add CmSelectWindowNum constant)
- Modify: `tv/window.go` (HandleEvent — respond to broadcast)

**Requirements:**
- New constant `CmSelectWindowNum` in command.go, before CmUser
- Window.HandleEvent handles EvBroadcast with CmSelectWindowNum: if `event.Info` is an `int` matching `w.number`, and the window is selectable (`HasOption(OfSelectable)`), call focus on self, consume event
- If the number doesn't match, event passes through (other windows get to check)
- EvBroadcast events go through the group's three-phase dispatch to reach windows, so the broadcast handling should be in the Window directly

**Implementation:**

Add to `tv/command.go`:
```go
CmSelectWindowNum
```

In `Window.HandleEvent`, add broadcast handling before the group dispatch:

```go
if event.What == EvBroadcast && event.Command == CmSelectWindowNum {
    if n, ok := event.Info.(int); ok && n == w.number && w.HasOption(OfSelectable) {
        if owner := w.Owner(); owner != nil {
            owner.SetFocusedChild(w)
        }
        event.Clear()
    }
    // Unconditional return: broadcasts are delivered to all children by the Group's
    // broadcast dispatch loop, so we must NOT forward to w.group.HandleEvent here
    // (that would re-deliver to child widgets, which should not respond to this).
    return
}
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat(window): respond to CmSelectWindowNum broadcast (spec 11.2)"`

---

### Task 3: Integration Checkpoint — Window Commands

**Purpose:** Verify CmZoom and CmSelectWindowNum work together.

**Requirements (for test writer):**
- Sending CmZoom command to a Window toggles its zoom state
- Sending CmZoom to a zoomed window un-zooms it
- CmSelectWindowNum broadcast with matching number focuses the window
- CmSelectWindowNum broadcast with non-matching number does not focus
- CmSelectWindowNum broadcast to a non-selectable window does not focus
- CmResize command is consumed (event cleared) but no state change

**Components to wire up:** Window with number, Desktop with multiple windows

**Run:** `go test ./tv/... -run TestIntegrationPhase14 -v`

---

## Batch 2: Architectural Corrections (Tasks 4–7)

### Task 4: Move CmTile/CmCascade to Application (spec 13.1)

**Files:**
- Modify: `tv/application.go` (handleCommand — add CmTile/CmCascade)
- Modify: `tv/desktop.go` (HandleEvent — remove CmTile/CmCascade cases)

**Requirements:**
- Application.handleCommand handles CmTile: calls `app.desktop.Tile()`, consumes event
- Application.handleCommand handles CmCascade: calls `app.desktop.Cascade()`, consumes event
- Desktop.HandleEvent no longer handles CmTile or CmCascade — removes those cases
- Desktop still handles CmNext and CmPrev
- The tile and cascade functionality itself is unchanged — just the responsibility for dispatching it moves
- All existing e2e tests for tile must continue to pass

**Implementation:**

In `tv/application.go`, `handleCommand`, add before the CmQuit case:
```go
case CmTile:
    if app.desktop != nil {
        app.desktop.Tile()
    }
    event.Clear()
    return
case CmCascade:
    if app.desktop != nil {
        app.desktop.Cascade()
    }
    event.Clear()
    return
```

In `tv/desktop.go`, `HandleEvent`, remove the CmTile and CmCascade cases from the EvCommand switch.

**Run tests:** `go test ./tv/... -count=1 && cd e2e && go test -run TestMenuTile -timeout 60s -count=1`

**Commit:** `git commit -m "refactor: move CmTile/CmCascade from Desktop to Application (spec 13.1)"`

---

### Task 5: Move Alt+N to Application with CmSelectWindowNum (spec 13.2)

**Files:**
- Modify: `tv/application.go` (handleEvent — add Alt+N handling)
- Modify: `tv/desktop.go` (HandleEvent — remove Alt+N handling)

**Requirements:**
- Application.handleEvent handles Alt+1-9 keyboard events: broadcast CmSelectWindowNum to the desktop with the window number as Info
- The broadcast goes through `app.desktop.HandleEvent` with an EvBroadcast event, which propagates to all windows via the group's three-phase dispatch
- Desktop.HandleEvent removes the Alt+N keyboard handling and the `selectWindowByNumber` method is no longer needed for this code path (but keep the method — it may be used elsewhere)
- The `selectWindowByNumber` helper in Desktop can be kept as-is; the Application will use broadcasting instead
- All existing e2e tests for Alt+N window switching must continue to pass

**Implementation:**

In `tv/application.go`, in `handleEvent`, add Alt+N handling BEFORE the desktop dispatch:

```go
if event.What == EvKeyboard && event.Key != nil {
    if event.Key.Modifiers&tcell.ModAlt != 0 && event.Key.Key == tcell.KeyRune {
        n := int(event.Key.Rune - '0')
        if n >= 1 && n <= 9 {
            bcast := &Event{What: EvBroadcast, Command: CmSelectWindowNum, Info: n}
            app.desktop.HandleEvent(bcast)
            if bcast.IsCleared() {
                event.Clear()
            }
            return
        }
    }
}
```

In `tv/desktop.go`, remove ONLY the inner Alt+N block (lines 96-103 — the `if event.Key.Modifiers&tcell.ModAlt...` through its closing brace). Keep the outer `if event.What == EvKeyboard` guard at line 95 — it's shared with the Tab forwarding block below. Do NOT remove the Tab/Shift+Tab forwarding in this task — that's Task 6.

**Run tests:** `go test ./tv/... -count=1 && cd e2e && go test -run TestHelpContext -timeout 60s -count=1`

**Commit:** `git commit -m "refactor: move Alt+N window switching from Desktop to Application (spec 13.2)"`

---

### Task 6: Move Tab from Group to Window (spec 13.3)

**Files:**
- Modify: `tv/window.go` (HandleEvent — add Tab/Shift+Tab handling)
- Modify: `tv/group.go` (HandleEvent — remove Tab/Shift+Tab handling)
- Modify: `tv/desktop.go` (HandleEvent — remove Tab forwarding)

**Requirements:**
- Window.HandleEvent handles Tab (KeyTab with no modifiers): calls `w.group.FocusNext()`, consumes event
- Window.HandleEvent handles Shift+Tab (KeyBacktab): calls `w.group.FocusPrev()`, consumes event
- Tab/Shift+Tab handling in Window must be BEFORE the group dispatch (so Tab is handled by the Window, not by its group)
- Group.HandleEvent removes Tab/Shift+Tab handling entirely (lines 275-287 of group.go — comment on 275, code on 276-287)
- Desktop.HandleEvent removes the Tab/Shift+Tab forwarding to focused window (lines 105-112 of desktop.go) — the Desktop should NOT handle Tab at all. After Task 5 removed Alt+N, the outer `if event.What == EvKeyboard` guard at line 95 may only contain the Tab block — remove the entire keyboard guard block if Tab was the only remaining content.
- The effect: Tab cycles widgets within a Window but does NOT cycle within a cluster's internal Group
- **Known behavioral change:** In the current code, Tab goes through Group's three-phase dispatch, so a focused child widget could consume Tab in Phase 2. After this change, Window intercepts Tab before it reaches the group, matching original Turbo Vision behavior. No existing widgets in this codebase consume Tab, so this is safe.
- All existing e2e tests for Tab navigation must continue to pass
- Group must expose `FocusNext()` and `FocusPrev()` as public methods if they aren't already

**Implementation:**

First, check if `focusNext`/`focusPrev` are public. They are currently lowercase (unexported). Need to either export them or add public wrappers.

Add to `tv/group.go`:
```go
func (g *Group) FocusNext() { g.focusNext() }
func (g *Group) FocusPrev() { g.focusPrev() }
```

In `tv/window.go`, add Tab handling in HandleEvent, after mouse check and before command handling:
```go
if event.What == EvKeyboard && event.Key != nil {
    if event.Key.Key == tcell.KeyTab && event.Key.Modifiers == 0 {
        w.group.FocusNext()
        event.Clear()
        return
    }
    if event.Key.Key == tcell.KeyBacktab {
        w.group.FocusPrev()
        event.Clear()
        return
    }
}
```

In `tv/group.go`, remove lines 275-287 (comment "Tab/Shift+Tab focus traversal" on line 275, code on 276-287).

In `tv/desktop.go`, remove the Tab/Shift+Tab forwarding block (lines 105-112 after Task 5's changes). If the outer keyboard guard `if event.What == EvKeyboard` is now empty after removing both Alt+N (Task 5) and Tab (this task), remove the entire guard block.

**Run tests:** `go test ./tv/... -count=1 && cd e2e && go test -run TestTabFocus -timeout 60s -count=1`

**Commit:** `git commit -m "refactor: move Tab/Shift+Tab from Group to Window (spec 13.3)"`

---

### Task 7: Integration Checkpoint — Architectural Corrections

**Purpose:** Verify all architectural moves work correctly end-to-end.

**Requirements (for test writer):**
- CmTile command reaches Application and tiles windows (Desktop no longer handles it directly)
- CmCascade command reaches Application and cascades windows
- Alt+1 broadcasts CmSelectWindowNum and focuses the correct window
- Tab in a Window cycles focus between its child widgets
- Tab in a Dialog cycles focus between its child widgets (Dialog extends Window)
- Tab does NOT cycle between items within a cluster's internal group (the whole point of spec 13.3)

**Components to wire up:** Application with Desktop, two Windows with child widgets, at least one cluster

**Run:** `go test ./tv/... -run TestIntegrationPhase14 -v`

---

## Batch 3: E2E Test (Task 8)

### Task 8: E2E Test — Architectural Corrections Regression

**Files:**
- Run existing e2e tests (no new tests needed — the existing tile, tab, and Alt+N tests are the regression suite)

**Requirements:**
- All 25 existing e2e tests must pass
- Specifically verify: TestMenuTileRearrangesWindows, TestTabFocusNavigation, TestHelpContextFiltering (which uses Alt+N)

**Implementation:**

Run the full e2e suite. If any test fails, fix the underlying issue.

**Run tests:** `cd e2e && go test -v -timeout 120s -count=1`

**Commit:** Only if changes are needed to fix regressions.
