# Phase 11: InputLine Enhancements — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement all 8 InputLine behavioral fidelity items (spec 5.1–5.8): word movement, word deletion, insert/overwrite mode, Ctrl+Y clear, mouse drag selection, double-click select-all, scroll indicators, and edge auto-scroll.

**Architecture:** All changes touch `tv/input_line.go`. The InputLine already has cursor movement, selection (selStart/selEnd), scrollOffset, and adjustScroll. We add word-boundary helpers, insert mode tracking, mouse drag state, and scroll indicator drawing. Edge auto-scroll uses the existing evMouseAuto system (implemented in Phase 1).

**Tech Stack:** Go, tcell/v2, existing InputLine/DrawBuffer/Event APIs

---

## File Structure

| File | Responsibility |
|------|---------------|
| `tv/input_line.go` | All InputLine logic — word boundaries, insert mode, mouse drag, scroll indicators |
| `tv/input_line_test.go` | Existing unit tests (may need updates for changed behaviors) |
| `tv/event.go` | No changes needed — MouseEvent.ClickCount already exists |
| `tv/application.go` | No changes — evMouseAuto already implemented |

---

## Batch 1: Keyboard Enhancements (Tasks 1–5)

### Task 1: Word Boundary Helpers

**Files:**
- Modify: `tv/input_line.go`

**Requirements:**
- Add `wordLeft(pos int) int` method: from position `pos`, move left to the start of the previous word. A word boundary is a transition from space to non-space when scanning leftward. If already at a word start, move to the start of the word before it. Returns 0 if no word boundary is found.
- Add `wordRight(pos int) int` method: from position `pos`, move right to the start of the next word. Skip non-space characters, then skip spaces, arriving at the start of the next word. Returns `len(text)` if no word boundary is found.
- Word boundaries treat any non-space rune as a word character (no distinction between alphanumeric and punctuation — matching original TV behavior).
- These are internal helpers used by Tasks 2 and 3.

**Implementation:**

```go
func (il *InputLine) wordLeft(pos int) int {
	if pos <= 0 {
		return 0
	}
	i := pos - 1
	// Skip spaces
	for i > 0 && il.text[i] == ' ' {
		i--
	}
	// Skip word characters
	for i > 0 && il.text[i-1] != ' ' {
		i--
	}
	return i
}

func (il *InputLine) wordRight(pos int) int {
	n := len(il.text)
	if pos >= n {
		return n
	}
	i := pos
	// Skip current word characters
	for i < n && il.text[i] != ' ' {
		i++
	}
	// Skip spaces
	for i < n && il.text[i] == ' ' {
		i++
	}
	return i
}
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: InputLine word boundary helpers"`

---

### Task 2: Word Movement — Ctrl+Left, Ctrl+Right (spec 5.1)

**Files:**
- Modify: `tv/input_line.go` (HandleEvent, KeyLeft/KeyRight cases)

**Requirements:**
- When Ctrl+Left is pressed (no Shift): move cursor to `wordLeft(cursorPos)`, clear selection
- When Ctrl+Right is pressed (no Shift): move cursor to `wordRight(cursorPos)`, clear selection
- When Ctrl+Shift+Left is pressed: extend selection to `wordLeft(cursorPos)` — if no selection exists, set selStart=cursorPos before moving
- When Ctrl+Shift+Right is pressed: extend selection to `wordRight(cursorPos)` — if no selection exists, set selStart=cursorPos before moving
- Plain Left/Right (no Ctrl) behavior is unchanged
- Shift+Left/Right (no Ctrl) behavior is unchanged
- `adjustScroll()` is called after each movement
- Event is cleared after handling

**Implementation:**

In the `case tcell.KeyLeft:` block, check for Ctrl modifier first:

```go
case tcell.KeyLeft:
	ctrl := ke.Modifiers&tcell.ModCtrl != 0
	if ctrl {
		newPos := il.wordLeft(il.cursorPos)
		if shift {
			if !il.hasSelection() {
				il.selStart = il.cursorPos
				il.selEnd = il.cursorPos
			}
			il.cursorPos = newPos
			il.selEnd = il.cursorPos
		} else {
			il.selStart = 0
			il.selEnd = 0
			il.cursorPos = newPos
		}
	} else if shift {
		if !il.hasSelection() {
			il.selStart = il.cursorPos
			il.selEnd = il.cursorPos
		}
		if il.cursorPos > 0 {
			il.cursorPos--
			il.selEnd = il.cursorPos
		}
	} else {
		il.selStart = 0
		il.selEnd = 0
		if il.cursorPos > 0 {
			il.cursorPos--
		}
	}
	il.adjustScroll()
	event.Clear()
```

Similar pattern for `case tcell.KeyRight:` using `wordRight`.

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: InputLine Ctrl+Left/Right word movement"`

---

### Task 3: Word Deletion — Ctrl+Backspace, Ctrl+Delete (spec 5.2)

**Files:**
- Modify: `tv/input_line.go` (HandleEvent, Backspace/Delete cases)

**Requirements:**
- When Ctrl+Backspace (or Alt+Backspace) is pressed: if selection exists, delete it. Otherwise, delete from `wordLeft(cursorPos)` to `cursorPos`.
- When Ctrl+Delete is pressed: if selection exists, delete it. Otherwise, delete from `cursorPos` to `wordRight(cursorPos)`.
- Plain Backspace/Delete behavior is unchanged (character-at-a-time deletion)
- `adjustScroll()` is called after deletion
- Event is cleared after handling

**Implementation:**

In the `case tcell.KeyBackspace, tcell.KeyBackspace2:` block:

```go
case tcell.KeyBackspace, tcell.KeyBackspace2:
	ctrl := ke.Modifiers&(tcell.ModCtrl|tcell.ModAlt) != 0
	if il.hasSelection() {
		il.deleteSelection()
	} else if ctrl {
		newPos := il.wordLeft(il.cursorPos)
		if newPos < il.cursorPos {
			il.text = append(il.text[:newPos], il.text[il.cursorPos:]...)
			il.cursorPos = newPos
		}
	} else if il.cursorPos > 0 {
		il.text = append(il.text[:il.cursorPos-1], il.text[il.cursorPos:]...)
		il.cursorPos--
	}
	il.adjustScroll()
	event.Clear()
```

In the `case tcell.KeyDelete:` block:

```go
case tcell.KeyDelete:
	ctrl := ke.Modifiers&tcell.ModCtrl != 0
	if il.hasSelection() {
		il.deleteSelection()
	} else if ctrl {
		newPos := il.wordRight(il.cursorPos)
		if newPos > il.cursorPos {
			il.text = append(il.text[:il.cursorPos], il.text[newPos:]...)
		}
	} else if il.cursorPos < len(il.text) {
		il.text = append(il.text[:il.cursorPos], il.text[il.cursorPos+1:]...)
	}
	il.adjustScroll()
	event.Clear()
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: InputLine Ctrl+Backspace/Delete word deletion"`

---

### Task 4: Insert/Overwrite Mode + Ctrl+Y Clear (spec 5.3, 5.4)

**Files:**
- Modify: `tv/input_line.go`

**Requirements:**
- Add `overwrite bool` field to InputLine struct
- Insert key toggles `overwrite` mode: `il.overwrite = !il.overwrite`
- In overwrite mode, typing a rune replaces the character at cursorPos instead of inserting. If cursorPos is at end of text, append instead of replace.
- `Overwrite() bool` getter method
- Ctrl+Y clears the entire input: sets text to empty, cursor to 0, clears selection
- In Draw: when overwrite mode is active and the view is selected, the cursor indicator uses a different visual — use the selection style on the character under cursor (same as current insert cursor, but this ensures the cursor is visible in both modes; the visual distinction is that overwrite mode has a block cursor on the character, while insert mode has the block cursor too — the distinction comes from the behavioral difference, not visual)
- Event is cleared after handling Insert key and Ctrl+Y

**Implementation:**

Add field:
```go
type InputLine struct {
	BaseView
	text         []rune
	maxLen       int
	cursorPos    int
	scrollOffset int
	selStart     int
	selEnd       int
	overwrite    bool
}

func (il *InputLine) Overwrite() bool { return il.overwrite }
```

In HandleEvent, add Insert key case in the default/Ctrl commands section:
```go
case tcell.KeyInsert:
	il.overwrite = !il.overwrite
	event.Clear()
```

Add Ctrl+Y:
```go
case tcell.KeyCtrlY:
	il.text = nil
	il.cursorPos = 0
	il.selStart = 0
	il.selEnd = 0
	il.scrollOffset = 0
	event.Clear()
```

Modify rune insertion to support overwrite:
```go
case tcell.KeyRune:
	if ke.Modifiers&(tcell.ModCtrl|tcell.ModAlt) != 0 {
		return
	}
	if il.hasSelection() {
		il.deleteSelection()
	}
	if il.overwrite && il.cursorPos < len(il.text) {
		il.text[il.cursorPos] = ke.Rune
		il.cursorPos++
	} else {
		if il.maxLen > 0 && len(il.text) >= il.maxLen {
			event.Clear()
			return
		}
		il.text = append(il.text, 0)
		copy(il.text[il.cursorPos+1:], il.text[il.cursorPos:])
		il.text[il.cursorPos] = ke.Rune
		il.cursorPos++
	}
	il.adjustScroll()
	event.Clear()
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: InputLine insert/overwrite mode and Ctrl+Y clear"`

---

### Task 5: Integration Checkpoint — Keyboard Enhancements

**Purpose:** Verify that Tasks 1–4 work together with real InputLine instances and event dispatch.

**Requirements (for test writer):**
- Typing "hello world" then Ctrl+Left moves cursor to start of "world" (position 6)
- Ctrl+Left twice from end moves cursor to start of "hello" (position 0)
- Ctrl+Right from position 0 with "hello world" moves to position 6 (start of "world")
- Ctrl+Shift+Left selects the previous word: cursor at end of "hello world", Ctrl+Shift+Left selects "world" (selStart=11, selEnd=6)
- Ctrl+Backspace from end of "hello world" deletes "world" leaving "hello "
- Ctrl+Delete from start of "hello world" deletes "hello " (word + trailing space) leaving "world"
- Ctrl+Backspace with active selection deletes the selection (not the word)
- Insert key toggles overwrite mode: type "abc", Home, press Insert, type "X" — text becomes "Xbc"
- Overwrite at end of text appends: text "ab", cursor at end, Insert mode, type "c" — text becomes "abc"
- Ctrl+Y clears all text, cursor at 0, no selection
- Ctrl+Y on non-empty text with selection: everything cleared

**Components to wire up:** InputLine, Event, direct method calls

**Run:** `go test ./tv/... -run TestIntegration -v`

---

## Batch 2: Mouse Selection (Tasks 6–8)

### Task 6: Mouse Drag Selection (spec 5.5)

**Files:**
- Modify: `tv/input_line.go` (HandleEvent mouse section, add dragging state)

**Requirements:**
- Add `dragging bool` and `dragAnchor int` fields to InputLine
- On Button1 press: set dragging=true, position cursor at click location, set dragAnchor=cursorPos, clear any existing selection. Clear event.
- On mouse move with Button1 held (dragging=true): update cursor position based on mouse X, set selection from dragAnchor to cursorPos. Clear event.
- On mouse release (no Button1, dragging=true): set dragging=false. Don't clear event (let it pass through).
- Mouse X is translated to text position: `col := mouseX + scrollOffset`. Clamp to [0, len(text)].
- The existing mouse click handler positions cursor but doesn't set up drag state — replace it with the new drag-aware handler.
- `adjustScroll()` is called on every mouse interaction.

**Implementation:**

Add fields:
```go
type InputLine struct {
	BaseView
	text         []rune
	maxLen       int
	cursorPos    int
	scrollOffset int
	selStart     int
	selEnd       int
	overwrite    bool
	dragging     bool
	dragAnchor   int
}
```

Replace the mouse handling block. Note: keep the `- il.Bounds().A.X` origin subtraction from the existing code — Group translates mouse coordinates to parent-local space, but the InputLine bounds may have a non-zero origin within its parent. The existing test `TestMouseClickAccountsForWidgetOriginX` verifies this.

**Drag-release limitation:** If the mouse leaves the InputLine's bounds during a drag and is released outside, the release event goes to whatever view is under the mouse — not the InputLine. The `dragging` state could get stuck. This is acceptable because (a) evMouseAuto events stop on release so no auto-scroll continues, and (b) the next click on the InputLine will reset dragging via the press branch. Full mouse capture (TV's original solution) is a separate architectural feature.

```go
case EvMouse:
	if event.Mouse == nil {
		return
	}
	col := event.Mouse.X - il.Bounds().A.X + il.scrollOffset
	if col < 0 {
		col = 0
	}
	if col > len(il.text) {
		col = len(il.text)
	}

	if event.Mouse.Button&tcell.Button1 != 0 {
		if !il.dragging {
			// Button1 press — start drag
			il.dragging = true
			il.dragAnchor = col
			il.cursorPos = col
			il.selStart = 0
			il.selEnd = 0
		} else {
			// Button1 held — extend selection
			il.cursorPos = col
			il.selStart = il.dragAnchor
			il.selEnd = il.cursorPos
		}
		il.adjustScroll()
		event.Clear()
	} else if il.dragging {
		// Button released
		il.dragging = false
	}
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: InputLine mouse drag selection"`

---

### Task 7: Double-Click Select All (spec 5.6)

**Files:**
- Modify: `tv/input_line.go` (HandleEvent mouse section)

**Requirements:**
- When a double-click is detected (ClickCount >= 2), select all text: set selStart=0, selEnd=len(text), cursorPos=len(text)
- Double-click detection uses `event.Mouse.ClickCount` which is already populated by the Application's mouse event conversion
- Double-click overrides the normal click behavior (don't start a drag)
- `adjustScroll()` is called
- Event is cleared

**Implementation:**

In the mouse handler, check ClickCount before the drag logic:

```go
if event.Mouse.Button&tcell.Button1 != 0 {
	if event.Mouse.ClickCount >= 2 {
		// Double-click: select all
		il.selStart = 0
		il.selEnd = len(il.text)
		il.cursorPos = len(il.text)
		il.dragging = false
		il.adjustScroll()
		event.Clear()
		return
	}
	if !il.dragging {
		// ... drag start logic
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: InputLine double-click select all"`

---

### Task 8: Integration Checkpoint — Mouse Selection

**Purpose:** Verify that Tasks 6–7 work correctly with realistic event sequences.

**Requirements (for test writer):**
- Click at position 3 in "hello world": cursor at 3, no selection, dragging started
- Click at position 3, then mouse move to position 8 with Button1 held: selection from 3 to 8
- Click at position 3, move to position 8, release: selection from 3 to 8, dragging=false
- Click at position 3, move to position 1 (backward drag): selection covers 1 to 3
- Double-click on "hello world": selects all text (selStart=0, selEnd=11)
- Double-click then single click: clears the select-all, starts new drag

**Components to wire up:** InputLine, mouse events with Button1, ClickCount

**Run:** `go test ./tv/... -run TestIntegration -v`

---

## Batch 3: Scroll Indicators and Edge Auto-Scroll (Tasks 9–12)

### Task 9: Scroll Indicators (spec 5.7)

**Files:**
- Modify: `tv/input_line.go` (Draw method)

**Requirements:**
- When text is scrolled off the left edge (scrollOffset > 0): draw `◄` at column 0, row 0
- When text extends beyond the right edge (scrollOffset + viewWidth < len(text)): draw `►` at column viewWidth-1, row 0
- The indicators replace the text/space character at those positions
- Use the `InputSelection` style for the indicators (no new theme field needed)
- When there is no overflow in a direction, the normal text or space is drawn (no indicator)
- The cursor can still be at position 0 or viewWidth-1 — indicators show overflow, cursor shows position

**Implementation:**

At the end of the Draw method, after rendering text and cursor, overlay the indicators:

```go
// Scroll indicators
indicatorStyle := normalStyle
if cs != nil {
	indicatorStyle = cs.InputSelection
}

if il.scrollOffset > 0 {
	buf.WriteChar(0, 0, '◄', indicatorStyle)
}
if il.scrollOffset+w < len(il.text) {
	buf.WriteChar(w-1, 0, '►', indicatorStyle)
}
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: InputLine scroll overflow indicators"`

---

### Task 10: Edge Auto-Scroll (spec 5.8)

**Files:**
- Modify: `tv/input_line.go` (HandleEvent mouse section)

**Requirements:**
- During a mouse drag (dragging=true, Button1 held), when the mouse X is at or past the left edge (X <= 0) and scrollOffset > 0: scroll left by 1 (decrease scrollOffset) and extend selection
- During a mouse drag, when the mouse X is at or past the right edge (X >= viewWidth - 1) and text extends beyond the view: scroll right by 1 (increase scrollOffset) and extend selection
- This works with evMouseAuto: the Application generates periodic EvMouse events while the button is held, so the auto-scroll happens continuously as long as the mouse stays at the edge
- The cursor position is clamped to valid range after scrolling
- `adjustScroll()` is called after scrolling

**Implementation:**

In the mouse handler, when dragging with Button1 held, add edge detection. Use the local X (after origin subtraction) for edge comparisons:

```go
if event.Mouse.Button&tcell.Button1 != 0 {
	// ... double-click check ...
	
	localX := event.Mouse.X - il.Bounds().A.X
	
	if !il.dragging {
		// ... drag start ...
	} else {
		// Button1 held — extend selection with edge auto-scroll
		w := il.Bounds().Width()
		if localX <= 0 && il.scrollOffset > 0 {
			il.scrollOffset--
			col = il.scrollOffset
		} else if localX >= w-1 && il.scrollOffset+w < len(il.text) {
			il.scrollOffset++
			col = il.scrollOffset + w - 1
		}
		if col > len(il.text) {
			col = len(il.text)
		}
		il.cursorPos = col
		il.selStart = il.dragAnchor
		il.selEnd = il.cursorPos
	}
	il.adjustScroll()
	event.Clear()
}
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: InputLine edge auto-scroll during drag"`

---

### Task 11: Integration Checkpoint — Scroll Features

**Purpose:** Verify that Tasks 9–10 work together with realistic scenarios.

**Requirements (for test writer):**
- InputLine with width=10, text "abcdefghijklmnop" (16 chars): at scroll offset 0, no ◄ indicator at col 0, but ► indicator at col 9
- Same InputLine scrolled so scrollOffset=5: both ◄ at col 0 and ► at col 9 are drawn
- Same InputLine scrolled to end (scrollOffset=6 so last char at col 9): ◄ at col 0, no ► at col 9
- Mouse drag to right edge with Button1 held: scrollOffset increases, selection extends
- Mouse drag to left edge with scrollOffset > 0: scrollOffset decreases, selection extends
- After edge scroll, cursor position is valid and within text bounds

**Components to wire up:** InputLine, DrawBuffer, mouse events

**Run:** `go test ./tv/... -run TestIntegration -v`

---

### Task 12: E2E Test — InputLine Enhancements

**Purpose:** Update the e2e test suite to verify Phase 11 functionality through the real tmux-driven application.

**Requirements:**
- Build the demo app binary
- Launch in tmux
- Test 1: Open InputBox (F3 in win1), type "hello world", press Ctrl+Y to clear, then type "new" and Enter — verify static text shows "File: new" (proves Ctrl+Y cleared the input before typing)
- Test 2: Open InputBox (F3 in win1), type "hello world", use Home to go to start, type "X" — verify "X" appears in the result (proves basic insert mode works; overwrite mode tested via unit/integration tests)
- Test 3: All previous e2e tests still pass (regression)

**Update demo app if needed:** The demo app already has F3 → InputBox. No changes needed.

**Run:** `go test ./e2e/... -v -timeout 120s`

**Commit:** `git commit -m "test: e2e tests for InputLine enhancements (Phase 11)"`
