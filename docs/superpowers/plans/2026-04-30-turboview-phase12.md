# Phase 12: ScrollBar Enhancements — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement all 5 ScrollBar behavioral fidelity items (spec 6.1–6.5): keyboard handling, arrow click auto-repeat, thumb drag, mouse wheel speed, and scrollbar broadcasts.

**Architecture:** All changes touch `tv/scrollbar.go` and `tv/commands.go`. The ScrollBar already has mouse click handling (arrows, page areas), wheel events, step/page methods, and an OnChange callback. We add keyboard dispatch, auto-repeat via evMouseAuto, thumb drag tracking, wheel speed multiplier, and owner broadcasts.

**Tech Stack:** Go, tcell/v2, existing ScrollBar/Event/Group APIs

---

## File Structure

| File | Responsibility |
|------|---------------|
| `tv/scrollbar.go` | All ScrollBar logic — keyboard handling, auto-repeat, thumb drag, wheel speed, broadcasts |
| `tv/commands.go` | New command constants: CmScrollBarClicked, CmScrollBarChanged |

---

## Batch 1: Keyboard and Wheel (Tasks 1–4)

### Task 1: ScrollBar Keyboard Handling (spec 6.1)

**Files:**
- Modify: `tv/scrollbar.go` (HandleEvent)

**Requirements:**
- ScrollBar.HandleEvent handles EvKeyboard events when the scrollbar has SfSelected state
- Add `arStep int` field (default arStep=1)
- Add setter: `SetArStep(n)`
- For vertical scrollbars:
  - Up arrow: step(-arStep) (decrease value)
  - Down arrow: step(arStep) (increase value)
  - Page Up: page(-1) (decrease by pageSize, using existing page method)
  - Page Down: page(1) (increase by pageSize)
  - Ctrl+Page Up: go to min
  - Ctrl+Page Down: go to max
- For horizontal scrollbars:
  - Left arrow: step(-arStep)
  - Right arrow: step(arStep)
  - Ctrl+Left: page(-1) (decrease by pageSize)
  - Ctrl+Right: page(1) (increase by pageSize)
  - Home: go to min
  - End: go to max
- Event is cleared after handling
- SetOptions(OfSelectable, true) in NewScrollBar so scrollbar can receive focus

**Implementation:**

Add fields and setters:
```go
type ScrollBar struct {
	BaseView
	orientation Orientation
	min         int
	max         int
	value       int
	pageSize    int
	arStep      int
	OnChange    func(int)
}

func NewScrollBar(bounds Rect, orientation Orientation) *ScrollBar {
	sb := &ScrollBar{
		orientation: orientation,
		arStep:      1,
	}
	sb.SetBounds(bounds)
	sb.SetState(SfVisible, true)
	sb.SetOptions(OfSelectable, true)
	sb.SetSelf(sb)
	return sb
}

func (sb *ScrollBar) SetArStep(n int) { sb.arStep = n }
func (sb *ScrollBar) ArStep() int     { return sb.arStep }
```

Add keyboard handling at the top of HandleEvent:
```go
func (sb *ScrollBar) HandleEvent(event *Event) {
	if event.What == EvKeyboard && event.Key != nil && sb.HasState(SfSelected) {
		sb.handleKeyboard(event)
		return
	}
	// ... existing mouse handling ...
}

func (sb *ScrollBar) handleKeyboard(event *Event) {
	ke := event.Key
	ctrl := ke.Modifiers&tcell.ModCtrl != 0

	if sb.orientation == Vertical {
		switch ke.Key {
		case tcell.KeyUp:
			sb.step(-sb.arStep)
			event.Clear()
		case tcell.KeyDown:
			sb.step(sb.arStep)
			event.Clear()
		case tcell.KeyPgUp:
			if ctrl {
				sb.goToMin()
			} else {
				sb.page(-1)
			}
			event.Clear()
		case tcell.KeyPgDn:
			if ctrl {
				sb.goToMax()
			} else {
				sb.page(1)
			}
			event.Clear()
		}
	} else {
		switch ke.Key {
		case tcell.KeyLeft:
			if ctrl {
				sb.page(-1)
			} else {
				sb.step(-sb.arStep)
			}
			event.Clear()
		case tcell.KeyRight:
			if ctrl {
				sb.page(1)
			} else {
				sb.step(sb.arStep)
			}
			event.Clear()
		case tcell.KeyHome:
			sb.goToMin()
			event.Clear()
		case tcell.KeyEnd:
			sb.goToMax()
			event.Clear()
		}
	}
}

func (sb *ScrollBar) goToMin() {
	old := sb.value
	sb.value = sb.min
	if sb.value != old && sb.OnChange != nil {
		sb.OnChange(sb.value)
	}
}

func (sb *ScrollBar) goToMax() {
	old := sb.value
	sb.value = sb.max - sb.pageSize
	sb.clampValue()
	if sb.value != old && sb.OnChange != nil {
		sb.OnChange(sb.value)
	}
}
```

Modify `step` to support multi-step:
```go
func (sb *ScrollBar) step(amount int) {
	old := sb.value
	sb.value += amount
	sb.clampValue()
	if sb.value != old && sb.OnChange != nil {
		sb.OnChange(sb.value)
	}
}
```

And update `page` similarly — the existing `page(dir)` multiplies by pageSize, which is correct.

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: ScrollBar keyboard handling for vertical and horizontal"`

---

### Task 2: Mouse Wheel Speed (spec 6.4)

**Files:**
- Modify: `tv/scrollbar.go` (HandleEvent wheel section)

**Requirements:**
- Mouse wheel should scroll by `3 * arStep` per wheel tick, not by 1
- WheelUp: step(-3 * arStep)
- WheelDown: step(3 * arStep)

**Implementation:**

Replace the existing wheel handling:
```go
if event.Mouse.Button == tcell.WheelUp {
	sb.step(-3 * sb.arStep)
	event.Clear()
	return
}
if event.Mouse.Button == tcell.WheelDown {
	sb.step(3 * sb.arStep)
	event.Clear()
	return
}
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "fix: ScrollBar mouse wheel scrolls 3x arStep per tick"`

---

### Task 3: Arrow Click Auto-Repeat (spec 6.2)

**Files:**
- Modify: `tv/scrollbar.go` (HandleEvent mouse section)

**Requirements:**
- When the mouse button is held on a scrollbar arrow, the evMouseAuto system generates repeat EvMouse events
- Each repeat event that hits the same arrow should trigger another step
- The existing code already handles single clicks on arrows — it just needs to NOT return early so repeat events also fire
- The current code already does this correctly because it checks `Button1 != 0` on every mouse event, and evMouseAuto sends events with Button1 held. So arrows already auto-repeat.
- Verify with tests that multiple mouse events on the same arrow accumulate steps

**Implementation:**

No code changes needed — the existing arrow click handling already works with evMouseAuto because it processes every EvMouse with Button1 held at the arrow position. The step function is idempotent per call.

Just verify with tests.

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "test: verify ScrollBar arrow auto-repeat works with evMouseAuto"`

---

### Task 4: Integration Checkpoint — Keyboard and Wheel

**Purpose:** Verify that Tasks 1–3 work together.

**Requirements (for test writer):**
- Vertical ScrollBar with range [0, 100], pageSize=10, arStep=1: Down arrow increases value by 1
- Same scrollbar: Page Down increases by pageSize (default 1 page)
- Same scrollbar: Ctrl+PgDn goes to max (90 with pageSize=10)
- Horizontal ScrollBar: Right arrow increases by arStep, Left decreases
- Horizontal ScrollBar: Home goes to min, End goes to max
- Mouse wheel on vertical scrollbar: WheelDown increases by 3*arStep
- Arrow auto-repeat: 5 EvMouse events with Button1 on the down arrow should increase value by 5*arStep
- OnChange fires for keyboard events

**Components to wire up:** ScrollBar, keyboard events, mouse events

**Run:** `go test ./tv/... -run TestIntegration -v`

---

## Batch 2: Thumb Drag and Broadcasts (Tasks 5–8)

### Task 5: Thumb Drag (spec 6.3)

**Files:**
- Modify: `tv/scrollbar.go`

**Requirements:**
- Add `thumbDragging bool` and `thumbDragOffset int` fields
- When mouse press hits the thumb area: start drag, record offset within thumb
- During drag (Button1 held): calculate new value proportionally from mouse position in track
- On release (no Button1 held while thumbDragging): stop drag
- Value updates continuously during drag, calling OnChange
- Thumb position maps linearly: value = min + (mouseTrackPos * scrollRange) / (trackLen - thumbLen)
- Clamp to valid range

**Implementation:**

Add fields:
```go
type ScrollBar struct {
	// ... existing ...
	thumbDragging   bool
	thumbDragOffset int
}
```

**Critical:** Modify the Button1 guard in HandleEvent to allow release events through when dragging:
```go
if event.Mouse.Button&tcell.Button1 == 0 && !sb.thumbDragging {
	return
}
```

In handleVerticalClick/handleHorizontalClick, detect if click is on the thumb:
```go
func (sb *ScrollBar) handleVerticalClick(event *Event) {
	my := event.Mouse.Y
	h := sb.Bounds().Height()

	// During drag
	if sb.thumbDragging {
		if event.Mouse.Button&tcell.Button1 != 0 {
			trackPos := my - 1 - sb.thumbDragOffset
			sb.setValueFromTrackPos(trackPos)
		} else {
			sb.thumbDragging = false
		}
		event.Clear()
		return
	}

	if my == 0 {
		sb.step(-sb.arStep)
		event.Clear()
		return
	}
	if my == h-1 {
		sb.step(sb.arStep)
		event.Clear()
		return
	}

	trackY := my - 1
	thumbPos, thumbLen := sb.thumbInfo()

	// Click on thumb — start drag
	if trackY >= thumbPos && trackY < thumbPos+thumbLen {
		sb.thumbDragging = true
		sb.thumbDragOffset = trackY - thumbPos
		event.Clear()
		return
	}

	// Click on track — page
	if trackY < thumbPos {
		sb.page(-1)
	} else {
		sb.page(1)
	}
	event.Clear()
}
```

Add value-from-track-position method:
```go
func (sb *ScrollBar) setValueFromTrackPos(trackPos int) {
	tl := sb.trackLen()
	_, thumbLen := sb.thumbInfo()
	scrollRange := sb.max - sb.pageSize - sb.min
	if scrollRange <= 0 || tl <= thumbLen {
		return
	}
	availableTrack := tl - thumbLen
	if availableTrack <= 0 {
		return
	}
	if trackPos < 0 {
		trackPos = 0
	}
	if trackPos > availableTrack {
		trackPos = availableTrack
	}
	old := sb.value
	sb.value = sb.min + trackPos*scrollRange/availableTrack
	sb.clampValue()
	if sb.value != old && sb.OnChange != nil {
		sb.OnChange(sb.value)
	}
}
```

Similar changes for handleHorizontalClick using event.Mouse.X.

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: ScrollBar thumb drag with proportional value mapping"`

---

### Task 6: ScrollBar Broadcasts (spec 6.5)

**Files:**
- Modify: `tv/scrollbar.go`
- Modify: `tv/commands.go` (add new constants)

**Requirements:**
- Add `CmScrollBarClicked` and `CmScrollBarChanged` command constants
- When any scrollbar interaction begins (mouse click, wheel, keyboard): broadcast CmScrollBarClicked to the owner
- When the scrollbar value changes: broadcast CmScrollBarChanged to the owner
- Broadcasts complement (don't replace) the existing OnChange callback
- Broadcast uses the same mechanism as focus broadcasts: `owner.HandleEvent(&Event{What: EvBroadcast, Command: cmd, Info: sb})`

**Implementation:**

In `tv/commands.go`, add:
```go
CmScrollBarClicked CommandCode = CmUser - 7
CmScrollBarChanged CommandCode = CmUser - 8
```

Add broadcast helper:
```go
func (sb *ScrollBar) broadcastToOwner(cmd CommandCode) {
	if owner := sb.Owner(); owner != nil {
		bcast := &Event{What: EvBroadcast, Command: cmd, Info: sb}
		owner.HandleEvent(bcast)
	}
}
```

In `step`, `page`, `goToMin`, `goToMax`, and `setValueFromTrackPos`: after value changes, call `sb.broadcastToOwner(CmScrollBarChanged)`.

At the start of `handleKeyboard`, `handleVerticalClick`/`handleHorizontalClick`, and wheel handling: call `sb.broadcastToOwner(CmScrollBarClicked)`.

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: ScrollBar CmScrollBarClicked/Changed broadcasts to owner"`

---

### Task 7: Integration Checkpoint — Thumb Drag and Broadcasts

**Purpose:** Verify Tasks 5–6 work together with real components.

**Requirements (for test writer):**
- Vertical ScrollBar with range [0, 100], pageSize=10: click on thumb, drag down, value increases proportionally
- Drag to bottom of track → value near max (90)
- Drag to top of track → value near min (0)
- CmScrollBarClicked broadcast is sent on first click
- CmScrollBarChanged broadcast is sent when value changes during drag
- CmScrollBarClicked is sent on wheel event
- CmScrollBarChanged is sent on keyboard event that changes value
- Both broadcasts AND OnChange fire (not one or the other)

**Components to wire up:** ScrollBar, Group (as owner for broadcasts), mouse/keyboard events

**Run:** `go test ./tv/... -run TestIntegration -v`

---

### Task 8: E2E Test — ScrollBar Enhancements

**Purpose:** Update the e2e test suite to verify Phase 12 functionality.

**Requirements:**
- Build the demo app binary
- Launch in tmux
- Test 1: In win2 (ListViewer with ScrollBar), press Down arrow multiple times, verify scrollbar thumb moves (the list scrolls and later items become visible — existing test `TestListViewerNavigation` already covers this, so just verify no regression)
- Test 2: All previous e2e tests still pass (regression)

**Run:** `go test ./e2e/... -v -timeout 120s`

**Commit:** `git commit -m "test: e2e regression check for ScrollBar enhancements (Phase 12)"`
