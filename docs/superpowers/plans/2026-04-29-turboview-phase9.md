# Phase 9: Event Dispatch Foundations + Dialog/Button Core

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the foundational event dispatch bugs and implement the Dialog/Button behavioral protocols that everything else depends on.

**Architecture:** Group's mouse routing changes from "send to focused" to "positional hit-test back-to-front." BaseView gains click-to-focus. Dialog gains Escape→CmCancel and Enter→cmDefault broadcast. Button gains the grab/release default protocol and loses its Enter handler.

**Tech Stack:** Go, tcell/v2, module `github.com/njt/turboview`, package `tv`

---

## File Structure

| File | Responsibility |
|------|---------------|
| `tv/command.go` | Add new command constants (CmDefault, CmGrabDefault, CmReleaseDefault, CmReceivedFocus, CmReleasedFocus) |
| `tv/group.go` | Fix mouse routing to positional, add disabled/eventmask filtering, add focus change broadcasts |
| `tv/base_view.go` | Add click-to-focus in BaseView.HandleEvent |
| `tv/application.go` | Add mouse auto-repeat generation |
| `tv/dialog.go` | Add Escape→CmCancel, Enter→cmDefault broadcast, modal termination |
| `tv/window.go` | Add CmClose→CmCancel for modal windows |
| `tv/button.go` | Replace isDefault with amDefault, add grab/release protocol, add Alt+shortcut, remove Enter, guard Space |

---

## Batch 1: Event Dispatch Foundations

### Task 1: New Command Constants

**Files:**
- Modify: `tv/command.go`

**Requirements:**
- `CmDefault` constant exists with a unique value after the existing constants
- `CmGrabDefault` constant exists
- `CmReleaseDefault` constant exists
- `CmReceivedFocus` constant exists
- `CmReleasedFocus` constant exists
- All new constants have distinct values
- All existing constants (`CmQuit` through `CmUser`) retain their current values

**Implementation:**

```go
const (
	CmQuit    CommandCode = iota + 1
	CmClose
	CmOK
	CmCancel
	CmYes
	CmNo
	CmMenu
	CmResize
	CmZoom
	CmTile
	CmCascade
	CmNext
	CmPrev

	CmDefault
	CmGrabDefault
	CmReleaseDefault
	CmReceivedFocus
	CmReleasedFocus

	CmUser CommandCode = 1000
)
```

**Run tests:** `go test ./tv/... -run TestCommand -v`

**Commit:** `git commit -m "feat: add command constants for default button protocol and focus broadcasts"`

---

### Task 2: Mouse Positional Routing in Group

**Files:**
- Modify: `tv/group.go:211-222`

**Requirements:**
- When Group receives an EvMouse event, it iterates children in reverse order (last child = topmost, checked first)
- It finds the first visible child whose bounds contain the mouse point `(event.Mouse.X, event.Mouse.Y)`
- It translates mouse coordinates to child-local space before delivery
- It delivers the event to that child only
- If no child contains the mouse point, the event is NOT delivered to any child (no fallback to focused)
- Non-mouse events (keyboard, command, broadcast) are NOT affected — they still use the three-phase dispatch
- Children with `SfDragging` state are NOT given special treatment here (Desktop handles drag capture separately)

**Implementation:**

Replace the mouse block in `Group.HandleEvent` (lines 217-222):

```go
// Mouse events: positional routing, back-to-front
if event.What == EvMouse {
    mx, my := event.Mouse.X, event.Mouse.Y
    for i := len(g.children) - 1; i >= 0; i-- {
        child := g.children[i]
        if !child.HasState(SfVisible) {
            continue
        }
        if child.Bounds().Contains(NewPoint(mx, my)) {
            event.Mouse.X -= child.Bounds().A.X
            event.Mouse.Y -= child.Bounds().A.Y
            child.HandleEvent(event)
            return
        }
    }
    return
}
```

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "fix: route mouse events positionally in Group instead of to focused child"`

---

### Task 3: Click-to-Focus in BaseView

**Files:**
- Modify: `tv/base_view.go:83`

**Requirements:**
- `BaseView.HandleEvent` handles EvMouse events with a real button press (Button1, Button2, or Button3 — NOT wheel events)
- If the view is NOT in `SfSelected` state, AND NOT `SfDisabled`, AND has `OfSelectable`:
  - Calls `owner.SetFocusedChild(self)` to gain focus (where `self` is the view that owns this BaseView — use the facade/owner pattern)
  - If the view does NOT have `OfFirstClick`, clears the event (click consumed by focusing)
  - If the view HAS `OfFirstClick`, leaves the event alone (click passes through)
- If the view IS already `SfSelected`, does nothing (event passes through for subclass handling)
- If the view has `SfDisabled`, does nothing
- If the view lacks `OfSelectable`, does nothing
- Non-mouse events are ignored by BaseView (passed through unchanged)
- Wheel events (WheelUp, WheelDown, etc.) do NOT trigger click-to-focus
- All widget constructors must call `SetSelf` so that BaseView knows the concrete View identity

**Implementation:**

Add a `self` field to BaseView. Each concrete type sets `self` during construction so BaseView can pass the concrete View to `SetFocusedChild`.

```go
// Add to BaseView struct:
type BaseView struct {
    // ... existing fields ...
    self View
}

// Add method:
func (b *BaseView) SetSelf(v View) { b.self = v }

// Add import for tcell at top of base_view.go:
import "github.com/gdamore/tcell/v2"

// Update HandleEvent:
func (b *BaseView) HandleEvent(event *Event) {
    if event.What == EvMouse && event.Mouse != nil {
        realButtons := event.Mouse.Button & (tcell.Button1 | tcell.Button2 | tcell.Button3)
        if realButtons != 0 && !b.HasState(SfSelected) && !b.HasState(SfDisabled) && b.HasOption(OfSelectable) {
            if b.owner != nil && b.self != nil {
                b.owner.SetFocusedChild(b.self)
            }
            if !b.HasOption(OfFirstClick) {
                event.Clear()
            }
        }
    }
}
```

**All widget constructors must add `SetSelf` calls.** This task includes updating every widget constructor:

- `NewButton` in `tv/button.go`: add `b.SetSelf(b)` before return
- `NewCheckBox` in `tv/checkbox.go`: add `cb.SetSelf(cb)` before return
- `NewCheckBoxes` in `tv/checkbox.go`: add `cbs.SetSelf(cbs)` before return
- `NewRadioButton` in `tv/radio.go`: add `rb.SetSelf(rb)` before return
- `NewRadioButtons` in `tv/radio.go`: add `rbs.SetSelf(rbs)` before return
- `NewInputLine` in `tv/input_line.go`: add `il.SetSelf(il)` before return
- `NewScrollBar` in `tv/scrollbar.go`: add `sb.SetSelf(sb)` before return
- `NewListViewer` in `tv/list_viewer.go`: add `lv.SetSelf(lv)` before return
- `NewLabel` in `tv/label.go`: add `l.SetSelf(l)` before return
- `NewStaticText` in `tv/static_text.go`: add `st.SetSelf(st)` before return

Window and Dialog don't need SetSelf — they handle mouse events themselves and delegate to their internal group.

For widgets whose HandleEvent overrides BaseView: they should call `b.BaseView.HandleEvent(event)` at the start, then check `event.IsCleared()` before their own handling. Button is updated in Task 13 of this plan. Other widgets (CheckBox, RadioButton, InputLine, etc.) are updated in later phases when their HandleEvent is corrected.

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: add click-to-focus behavior in BaseView.HandleEvent"`

---

### Task 4: Disabled View and EventMask Filtering in Group

**Files:**
- Modify: `tv/group.go:249-274`

**Requirements:**
- During three-phase dispatch (preProcess, focused, postProcess phases):
  - Skip children with `SfDisabled` state
  - Only deliver events whose `EventType` is included in the child's `EventMask` (if EventMask is 0, deliver all events — 0 means "accept everything", matching TV's default)
- During positional mouse routing:
  - Skip children with `SfDisabled` state (already visible-only check exists, add disabled check)
- A disabled child is completely excluded from event dispatch (not just keyboard — mouse too)
- EventMask filtering: `child.EventMask() == 0` means accept all; otherwise `event.What & child.EventMask() != 0` must be true
- Broadcast events (`EvBroadcast`) skip disabled children but do NOT check EventMask (broadcasts always deliver)

**Implementation:**

Add a helper method:

```go
func (g *Group) canReceiveEvent(child View, eventType EventType) bool {
    if child.HasState(SfDisabled) {
        return false
    }
    mask := child.EventMask()
    if mask == 0 {
        return true
    }
    return eventType&mask != 0
}
```

Update mouse routing:
```go
if event.What == EvMouse {
    mx, my := event.Mouse.X, event.Mouse.Y
    for i := len(g.children) - 1; i >= 0; i-- {
        child := g.children[i]
        if !child.HasState(SfVisible) || child.HasState(SfDisabled) {
            continue
        }
        if child.Bounds().Contains(NewPoint(mx, my)) {
            event.Mouse.X -= child.Bounds().A.X
            event.Mouse.Y -= child.Bounds().A.Y
            child.HandleEvent(event)
            return
        }
    }
    return
}
```

Update broadcast:
```go
if event.What == EvBroadcast {
    for _, child := range g.children {
        if event.IsCleared() {
            return
        }
        if child.HasState(SfDisabled) {
            continue
        }
        child.HandleEvent(event)
    }
    return
}
```

Update three-phase dispatch:
```go
// Phase 1: Preprocess
for _, child := range g.children {
    if event.IsCleared() {
        return
    }
    if child != g.focused && child.HasOption(OfPreProcess) && g.canReceiveEvent(child, event.What) {
        child.HandleEvent(event)
    }
}

// Phase 2: Focused
if !event.IsCleared() && g.focused != nil && g.canReceiveEvent(g.focused, event.What) {
    g.focused.HandleEvent(event)
}

// Phase 3: Postprocess
for _, child := range g.children {
    if event.IsCleared() {
        return
    }
    if child != g.focused && child.HasOption(OfPostProcess) && g.canReceiveEvent(child, event.What) {
        child.HandleEvent(event)
    }
}
```

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: add disabled view and EventMask filtering in Group dispatch"`

---

### Task 5: Focus Change Broadcasts

**Files:**
- Modify: `tv/group.go:277-285` (selectChild method)

**Requirements:**
- When `selectChild` changes focus from old child to new child:
  - If old child exists and is different from new child: broadcast `EvBroadcast` with `CmReleasedFocus` and old child as `Info` to all children
  - If new child exists: broadcast `EvBroadcast` with `CmReceivedFocus` and new child as `Info` to all children
- When selectChild is called with the same child that's already focused, no broadcasts are sent
- When selectChild is called with nil (clearing focus), only `CmReleasedFocus` is broadcast for the old child
- Broadcast delivery: iterate all children, skip disabled children, deliver unconditionally (do NOT stop on Clear — every non-disabled child receives the notification)
- The broadcasts happen AFTER the state change (SfSelected flags are already updated)

**Implementation:**

```go
func (g *Group) selectChild(v View) {
    old := g.focused
    if old == v {
        return
    }

    if old != nil {
        old.SetState(SfSelected, false)
    }
    g.focused = v
    if v != nil {
        v.SetState(SfSelected, true)
    }

    // Broadcast focus changes — deliver to ALL non-disabled children unconditionally
    if old != nil {
        for _, child := range g.children {
            if child.HasState(SfDisabled) {
                continue
            }
            ev := &Event{What: EvBroadcast, Command: CmReleasedFocus, Info: old}
            child.HandleEvent(ev)
        }
    }
    if v != nil {
        for _, child := range g.children {
            if child.HasState(SfDisabled) {
                continue
            }
            ev := &Event{What: EvBroadcast, Command: CmReceivedFocus, Info: v}
            child.HandleEvent(ev)
        }
    }
}
```

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: broadcast CmReceivedFocus/CmReleasedFocus on focus changes"`

---

### Task 6: Mouse Auto-Repeat (evMouseAuto)

**Files:**
- Modify: `tv/application.go`

**Requirements:**
- After the Application receives a mouse event with a button press, it starts tracking for auto-repeat
- While the button remains held (no release event), the Application generates synthetic EvMouse events at approximately 50ms intervals
- Synthetic events have the same button flags and current mouse position as the original press
- When a mouse release event arrives (button mask becomes 0 or changes), auto-repeat stops
- The synthetic events use `EvMouse` type (not a separate event type) — views handle them identically to real mouse events
- Auto-repeat does NOT start for mouse wheel events (WheelUp/WheelDown) — only for Button1/Button2/Button3
- Auto-repeat works during modal loops (ExecView, MenuBar.ActivateAt) — these call PollEvent which is where the synthetic events are injected

**Implementation:**

Add tracking state to Application:

```go
type Application struct {
    // ... existing fields ...
    mouseAutoMu   sync.Mutex
    mouseAutoBtn  tcell.ButtonMask
    mouseAutoX    int
    mouseAutoY    int
    mouseAutoChan chan struct{}
}
```

Use `screen.PostEvent` from a goroutine to inject synthetic events into tcell's event queue. Protect shared mouse position with a mutex since the goroutine reads coordinates that the main goroutine writes.

```go
type mouseAutoEvent struct {
    tcell.EventTime
    x, y   int
    button tcell.ButtonMask
}

func (app *Application) startMouseAuto(x, y int, button tcell.ButtonMask) {
    app.stopMouseAuto()
    app.mouseAutoMu.Lock()
    app.mouseAutoBtn = button
    app.mouseAutoX = x
    app.mouseAutoY = y
    done := make(chan struct{})
    app.mouseAutoChan = done
    app.mouseAutoMu.Unlock()
    go func() {
        ticker := time.NewTicker(50 * time.Millisecond)
        defer ticker.Stop()
        for {
            select {
            case <-done:
                return
            case <-ticker.C:
                app.mouseAutoMu.Lock()
                ev := &mouseAutoEvent{x: app.mouseAutoX, y: app.mouseAutoY, button: app.mouseAutoBtn}
                app.mouseAutoMu.Unlock()
                ev.SetEventNow()
                _ = app.screen.PostEvent(ev)
            }
        }
    }()
}

func (app *Application) stopMouseAuto() {
    app.mouseAutoMu.Lock()
    defer app.mouseAutoMu.Unlock()
    if app.mouseAutoChan != nil {
        close(app.mouseAutoChan)
        app.mouseAutoChan = nil
    }
    app.mouseAutoBtn = 0
}
```

Update `convertEvent` to handle `mouseAutoEvent`:
```go
case *mouseAutoEvent:
    return &Event{
        What: EvMouse,
        Mouse: &MouseEvent{
            X:      ev.x,
            Y:      ev.y,
            Button: ev.button,
        },
    }
```

Update `convertEvent` for regular mouse events — track button state:
```go
case *tcell.EventMouse:
    x, y := ev.Position()
    buttons := ev.Buttons()
    realButtons := buttons & (tcell.Button1 | tcell.Button2 | tcell.Button3)
    if realButtons != 0 && app.mouseAutoBtn == 0 {
        app.startMouseAuto(x, y, realButtons)
    } else if realButtons != 0 {
        app.mouseAutoMu.Lock()
        app.mouseAutoX = x
        app.mouseAutoY = y
        app.mouseAutoMu.Unlock()
    } else if realButtons == 0 && app.mouseAutoBtn != 0 {
        app.stopMouseAuto()
    }
    return &Event{
        What: EvMouse,
        Mouse: &MouseEvent{
            X:         x,
            Y:         y,
            Button:    buttons,
            Modifiers: ev.Modifiers(),
        },
    }
```

Add `import "sync"` and `import "time"` to application.go.

Also ensure `stopMouseAuto()` is called in `Run()` cleanup (before `screen.Fini()`).

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: add mouse auto-repeat for held button events"`

---

### Task 7: Integration Checkpoint — Event Dispatch Chain

**Purpose:** Verify that Tasks 1-6 work together: positional mouse routing, click-to-focus, disabled filtering, and focus broadcasts all cooperate correctly.

**Requirements (for test writer):**
- A mouse click on an unfocused Button inside a Group focuses that Button (positional routing + click-to-focus)
- A mouse click on Button A (unfocused, no OfFirstClick) focuses it and consumes the click — the button does NOT fire its command
- A mouse click on Button B (unfocused, WITH OfFirstClick) focuses it AND fires its command
- A disabled view inside a Group does not receive mouse events even when the mouse is over it
- A disabled view inside a Group does not receive keyboard events during three-phase dispatch
- A view with EventMask set to EvMouse only does not receive keyboard events
- When focus changes from child A to child B, CmReleasedFocus is broadcast with child A as Info
- When focus changes from child A to child B, CmReceivedFocus is broadcast with child B as Info
- Clicking on empty space in a Group (no child under mouse) does not crash and does not change focus

**Components to wire up:** Group, BaseView (via Button), command constants (all real, no mocks)

**Run:** `go test ./tv/... -run TestIntegration -v`

---

## Batch 2: Dialog and Modal Behavior

### Task 8: Escape Key in Dialog

**Files:**
- Modify: `tv/dialog.go:232-245`

**Requirements:**
- When Dialog receives an EvKeyboard event with KeyEscape:
  - The event is first delegated to the group (focused child gets first crack)
  - If the event is NOT cleared after group delegation, Dialog transforms it: sets `event.What = EvCommand`, `event.Command = CmCancel`, clears `event.Key`
- If the focused child handles Escape (clears the event), Dialog does NOT transform it
- The transformed CmCancel command causes the modal loop (ExecView) to exit with CmCancel result
- Non-Escape keyboard events pass through to the group unchanged
- Mouse events are handled as before (no change to mouse routing)

**Implementation:**

Replace `Dialog.HandleEvent`:

```go
func (d *Dialog) HandleEvent(event *Event) {
    if event.What == EvMouse && event.Mouse != nil {
        width, height := d.Bounds().Width(), d.Bounds().Height()
        mx, my := event.Mouse.X, event.Mouse.Y
        if mx > 0 && mx < width-1 && my > 0 && my < height-1 {
            event.Mouse.X -= 1
            event.Mouse.Y -= 1
            d.group.HandleEvent(event)
        }
        return
    }

    // Delegate to group first
    d.group.HandleEvent(event)

    // Escape → CmCancel (if group didn't handle it)
    if !event.IsCleared() && event.What == EvKeyboard && event.Key != nil {
        if event.Key.Key == tcell.KeyEscape {
            event.What = EvCommand
            event.Command = CmCancel
            event.Key = nil
        }
    }
}
```

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: Dialog converts Escape to CmCancel after group delegation"`

---

### Task 9: Enter Key in Dialog — cmDefault Broadcast

**Files:**
- Modify: `tv/dialog.go` (HandleEvent method, same method as Task 8)

**Requirements:**
- When Dialog receives an EvKeyboard event with KeyEnter:
  - The event is first delegated to the group (focused child gets first crack)
  - If the event is NOT cleared after group delegation, Dialog broadcasts `EvBroadcast` with `CmDefault` to all children via the group
  - Then clears the original event
- If the focused child handles Enter (clears the event), Dialog does NOT broadcast
- The CmDefault broadcast is delivered to ALL children (standard broadcast — the default button responds by calling press())
- This works regardless of which widget is currently focused

**Implementation:**

Add to the HandleEvent after the Escape handling:

```go
func (d *Dialog) HandleEvent(event *Event) {
    if event.What == EvMouse && event.Mouse != nil {
        width, height := d.Bounds().Width(), d.Bounds().Height()
        mx, my := event.Mouse.X, event.Mouse.Y
        if mx > 0 && mx < width-1 && my > 0 && my < height-1 {
            event.Mouse.X -= 1
            event.Mouse.Y -= 1
            d.group.HandleEvent(event)
        }
        return
    }

    // Delegate to group first
    d.group.HandleEvent(event)

    if !event.IsCleared() && event.What == EvKeyboard && event.Key != nil {
        switch event.Key.Key {
        case tcell.KeyEscape:
            event.What = EvCommand
            event.Command = CmCancel
            event.Key = nil

        case tcell.KeyEnter:
            // Broadcast CmDefault to all children — the default button responds
            broadcast := &Event{What: EvBroadcast, Command: CmDefault}
            d.group.HandleEvent(broadcast)
            event.Clear()
        }
    }
}
```

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: Dialog broadcasts CmDefault on Enter for default button protocol"`

---

### Task 10: Dialog Modal Termination

**Files:**
- Modify: `tv/dialog.go`

**Requirements:**
- When Dialog is modal (`SfModal` state) and receives an EvCommand event:
  - For `CmOK`, `CmCancel`, `CmYes`, `CmNo`: transform the event so ExecView's modal loop sees it and exits
  - For `CmClose`: transform to `CmCancel` (close = cancel for modal dialogs)
- When Dialog is NOT modal, command events pass through unchanged
- The ExecView modal loop in group.go already checks for these commands after HandleEvent — Dialog just needs to NOT clear them, and ensure CmClose is mapped to CmCancel
- EvCommand events should be handled AFTER group delegation (so buttons inside the dialog can set the command first)

**Implementation:**

Add command handling to Dialog.HandleEvent:

```go
func (d *Dialog) HandleEvent(event *Event) {
    if event.What == EvMouse && event.Mouse != nil {
        width, height := d.Bounds().Width(), d.Bounds().Height()
        mx, my := event.Mouse.X, event.Mouse.Y
        if mx > 0 && mx < width-1 && my > 0 && my < height-1 {
            event.Mouse.X -= 1
            event.Mouse.Y -= 1
            d.group.HandleEvent(event)
        }
        return
    }

    // Delegate to group first
    d.group.HandleEvent(event)

    if !event.IsCleared() && event.What == EvKeyboard && event.Key != nil {
        switch event.Key.Key {
        case tcell.KeyEscape:
            event.What = EvCommand
            event.Command = CmCancel
            event.Key = nil

        case tcell.KeyEnter:
            broadcast := &Event{What: EvBroadcast, Command: CmDefault}
            d.group.HandleEvent(broadcast)
            event.Clear()
        }
    }

    // Modal termination: CmClose → CmCancel
    if !event.IsCleared() && event.What == EvCommand && d.HasState(SfModal) {
        if event.Command == CmClose {
            event.Command = CmCancel
        }
    }
}
```

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: Dialog handles modal termination and CmClose→CmCancel"`

---

### Task 11: Window CmClose on Modal

**Files:**
- Modify: `tv/window.go:161-167`

**Requirements:**
- When Window receives `EvCommand` with `CmClose` and the window has `SfModal` state:
  - Transform the command to `CmCancel` (do NOT close the window)
- When Window receives `CmClose` without `SfModal`, behavior is unchanged (event passes through for Desktop to handle)
- This prevents a modal window from being closed — the close converts to a cancel instead

**Implementation:**

Update `Window.HandleEvent`:

```go
func (w *Window) HandleEvent(event *Event) {
    if event.What == EvMouse && event.Mouse != nil {
        w.handleMouseEvent(event)
        return
    }

    // Modal window: CmClose → CmCancel
    if event.What == EvCommand && event.Command == CmClose && w.HasState(SfModal) {
        event.Command = CmCancel
        return
    }

    w.group.HandleEvent(event)
}
```

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: Window converts CmClose to CmCancel when modal"`

---

### Task 12: Integration Checkpoint — Dialog Modal Protocol

**Purpose:** Verify that Tasks 8-11 work together: Dialog's Escape/Enter handling, modal termination, and Window modal close.

**Requirements (for test writer):**
- A Dialog shown via ExecView returns CmCancel when Escape is pressed
- A Dialog with a default Button returns the button's command when Enter is pressed (Enter→CmDefault broadcast→button press)
- A Dialog with a focused widget that handles Escape first: the focused widget's handler runs, Dialog does NOT transform to CmCancel
- Clicking a Button inside a modal Dialog fires the button's command and the modal loop exits with that command
- CmClose sent to a modal Dialog is converted to CmCancel — ExecView returns CmCancel
- A Window with SfModal converts CmClose to CmCancel

**Components to wire up:** Dialog, Group, Button, ExecView modal loop (all real, no mocks)

**Run:** `go test ./tv/... -run TestIntegration -v`

---

## Batch 3: Button Protocol

### Task 13: Button Default Protocol (amDefault, Grab/Release)

**Files:**
- Modify: `tv/button.go`

**Requirements:**
- `isDefault` field is replaced by two fields: `bfDefault bool` (was created with WithDefault — permanent trait) and `amDefault bool` (currently acting as default — dynamic)
- `WithDefault()` sets both `bfDefault = true` and `amDefault = true`, and sets `OfPostProcess`
- `IsDefault()` returns `amDefault` (for rendering)
- `Button.Draw` uses `amDefault` to decide the button style (already uses `isDefault` — just rename)
- On `CmGrabDefault` broadcast: if `bfDefault` is true, set `amDefault = false`
- On `CmReleaseDefault` broadcast: if `bfDefault` is true, set `amDefault = true`
- On `CmDefault` broadcast: if `amDefault` is true, call `press(event)` (press transforms the broadcast event to EvCommand — the modal loop picks it up)
- When Button gains focus (`SfSelected` becomes true via SetState): if `bfDefault` is false, broadcast `CmGrabDefault` to owner's children, then set `amDefault = true`
- When Button loses focus (`SfSelected` becomes false via SetState): if `bfDefault` is false, broadcast `CmReleaseDefault` to owner's children, then set `amDefault = false`
- Override `SetState` on Button to intercept SfSelected changes and trigger grab/release
- All buttons set `OfPostProcess` (not just default buttons) so they can receive broadcasts

**Implementation:**

```go
type Button struct {
    BaseView
    title     string
    command   CommandCode
    bfDefault bool
    amDefault bool
}

func WithDefault() ButtonOption {
    return func(b *Button) {
        b.bfDefault = true
        b.amDefault = true
        b.SetOptions(OfPostProcess, true)
    }
}

func NewButton(bounds Rect, title string, command CommandCode, opts ...ButtonOption) *Button {
    b := &Button{title: title, command: command}
    b.SetBounds(bounds)
    b.SetState(SfVisible, true)
    b.SetOptions(OfSelectable, true)
    b.SetOptions(OfPostProcess, true)
    for _, opt := range opts {
        opt(b)
    }
    b.SetSelf(b)
    return b
}

func (b *Button) IsDefault() bool { return b.amDefault }

func (b *Button) SetState(flag ViewState, on bool) {
    b.BaseView.SetState(flag, on)
    if flag&SfSelected != 0 && !b.bfDefault {
        if on {
            b.broadcastToOwner(CmGrabDefault)
            b.amDefault = true
        } else {
            b.broadcastToOwner(CmReleaseDefault)
            b.amDefault = false
        }
    }
}

func (b *Button) broadcastToOwner(cmd CommandCode) {
    if b.owner == nil {
        return
    }
    ev := &Event{What: EvBroadcast, Command: cmd, Info: b}
    for _, child := range b.owner.Children() {
        if child != View(b) {
            child.HandleEvent(ev)
        }
    }
}

func (b *Button) HandleEvent(event *Event) {
    // Click-to-focus
    if event.What == EvMouse && event.Mouse != nil {
        b.BaseView.HandleEvent(event)
        if event.IsCleared() {
            return
        }
        if event.Mouse.Button&tcell.Button1 != 0 {
            b.press(event)
        }
        return
    }

    // Broadcast handling
    if event.What == EvBroadcast {
        switch event.Command {
        case CmDefault:
            if b.amDefault {
                b.press(event)
            }
        case CmGrabDefault:
            if b.bfDefault {
                b.amDefault = false
            }
        case CmReleaseDefault:
            if b.bfDefault {
                b.amDefault = true
            }
        }
        return
    }

    if event.What == EvKeyboard && event.Key != nil {
        if event.Key.Key == tcell.KeyRune && event.Key.Rune == ' ' {
            if b.HasState(SfSelected) {
                b.press(event)
            }
        }
    }
}
```

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: implement Button default protocol with grab/release broadcasts"`

---

### Task 14: Button Alt+Shortcut

**Files:**
- Modify: `tv/button.go`

**Requirements:**
- Button parses its tilde-marked shortcut character from the title (same as RadioButton/CheckBox)
- Store the shortcut rune as a field
- When Button receives EvKeyboard with ModAlt and the rune matches the shortcut (case-insensitive):
  - Call `press(event)`
- This works in the postProcess phase (OfPostProcess is set on all buttons), so it activates even when the button isn't focused
- If no tilde shortcut is present in the title, no Alt+key handling occurs

**Implementation:**

Add shortcut field and parsing:

```go
type Button struct {
    BaseView
    title     string
    shortcut  rune
    command   CommandCode
    bfDefault bool
    amDefault bool
}

func NewButton(bounds Rect, title string, command CommandCode, opts ...ButtonOption) *Button {
    b := &Button{title: title, command: command}
    b.SetBounds(bounds)
    b.SetState(SfVisible, true)
    b.SetOptions(OfSelectable, true)
    b.SetOptions(OfPostProcess, true)

    segments := ParseTildeLabel(title)
    for _, seg := range segments {
        if seg.Shortcut && len(seg.Text) > 0 {
            b.shortcut, _ = utf8.DecodeRuneInString(seg.Text)
            break
        }
    }

    for _, opt := range opts {
        opt(b)
    }
    b.SetSelf(b)
    return b
}
```

Add to HandleEvent keyboard section:

```go
if event.What == EvKeyboard && event.Key != nil {
    switch event.Key.Key {
    case tcell.KeyRune:
        if event.Key.Rune == ' ' && b.HasState(SfSelected) {
            b.press(event)
        } else if event.Key.Modifiers&tcell.ModAlt != 0 && b.shortcut != 0 {
            if unicode.ToLower(event.Key.Rune) == unicode.ToLower(b.shortcut) {
                b.press(event)
            }
        }
    }
}
```

Add `"unicode"` to imports.

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: add Alt+shortcut support to Button"`

---

### Task 15: Button Space Guard and Enter Removal

**Files:**
- Modify: `tv/button.go` (HandleEvent)

**Requirements:**
- Space bar ONLY activates the button when the button has `SfSelected` state (is focused)
- If Space is pressed and button is NOT focused (SfSelected), the event passes through unchanged
- Enter key does NOT activate the button — Enter is handled by Dialog's CmDefault broadcast
- The `case tcell.KeyEnter` is completely removed from Button.HandleEvent
- Mouse click still works regardless of focus state

**Implementation:**

This is already covered in the HandleEvent from Task 13 — the Enter case is removed, and Space checks `b.HasState(SfSelected)`. Confirm these are in place.

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "fix: Button Space requires focus, remove Enter handler"`

---

### Task 16: Integration Checkpoint — Default Button Protocol End-to-End

**Purpose:** Verify that Tasks 13-15 work together with Dialog from Tasks 8-10.

**Requirements (for test writer):**
- In a Dialog with OK (default) and Cancel buttons: Enter activates OK via the cmDefault broadcast chain (Dialog→broadcast→Button.amDefault→press)
- In a Dialog, when focus moves from OK (default) to Cancel: Cancel broadcasts CmGrabDefault, OK's amDefault becomes false, Cancel's amDefault becomes true
- In a Dialog, when focus returns to OK: Cancel broadcasts CmReleaseDefault, OK's amDefault becomes true, Cancel's amDefault becomes false
- Alt+shortcut on a Button fires the button's command even when a different widget is focused
- Space on a focused Button fires its command; Space on an unfocused Button does nothing
- Enter does NOT fire a non-default button — only the default button responds via broadcast
- The full flow: Dialog with 3 buttons (Save=default, Cancel, Help), focus on Cancel, press Enter → Save fires (not Cancel)

**Components to wire up:** Dialog, Group, Button (all real, no mocks)

**Run:** `go test ./tv/... -run TestIntegration -v`

---

## Batch 4: E2E Test Update

### Task 17: E2E Test — Event Dispatch and Dialog Protocol

**Purpose:** Update the e2e test suite to verify Phase 9 functionality through the real application.

**Requirements:**
- Build the demo app binary
- Launch in tmux
- Test 1: Click on an unfocused widget — verify it gains focus (screen shows focus indicator)
- Test 2: Open a dialog (via menu or keyboard) → press Escape → verify dialog closes (desktop visible again)
- Test 3: Open a dialog with OK/Cancel → press Enter → verify OK fires (dialog closes with OK result, observable in app behavior)
- Test 4: Open a dialog → Tab to Cancel → verify Cancel now has default appearance → press Enter → verify OK still fires (default follows focus but Enter fires true default)
  - Actually: with the grab/release protocol, Enter fires whichever button currently has amDefault. When Cancel is focused, Cancel grabs default, so Enter fires Cancel. This is correct TV behavior.
- Test 5: All previous e2e tests still pass (regression)

**Update demo app:** The demo app should have a dialog that can be opened via a menu command. The dialog should have at least two buttons (OK and Cancel). Menu → "File" → "New" (or similar) could open a dialog.

**Run:** `go test ./e2e/... -v -timeout 120s`

**Commit:** `git commit -m "test: e2e tests for event dispatch and dialog protocol"`
