# Phase 10: Widget Corrections Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix behavioral fidelity issues in cluster widgets (CheckBoxes, RadioButtons), Label, StatusLine, StaticText, and menu command dispatch to match original Turbo Vision behavior.

**Architecture:** Each fix is isolated to one or two files. Cluster fixes correct focus indicator suppression, arrow key navigation, and Enter key removal. Label gets click-to-focus-linked-view, visual focus tracking via broadcasts, and phase change. StatusLine gets mouse click support. StaticText gets `\x03` line centering. Menu command dispatch routes through the desktop instead of bypassing it.

**Tech Stack:** Go, tcell/v2, package `tv` and `theme`

---

## File Structure

| File | Changes |
|------|---------|
| `tv/checkbox.go` | Remove SfSelected suppression in CheckBoxes.Draw, add arrow key navigation to CheckBoxes.HandleEvent, remove Enter handler from CheckBox.HandleEvent |
| `tv/radio.go` | Remove SfSelected suppression in RadioButtons.Draw, add Left/Right arrow keys to RadioButtons.HandleEvent, remove Enter handler from RadioButton.HandleEvent |
| `tv/label.go` | Add `light` field, click handler, CmReceivedFocus/CmReleasedFocus broadcast handler, change OfPreProcess to OfPostProcess, add LabelHighlight color |
| `tv/status.go` | Add mouse event handling — hit-test status items, highlight on press, fire command on release |
| `tv/static_text.go` | Add `\x03` centering logic in Draw |
| `tv/menu_bar.go` | Change `checkPopupResult` to use `app.PostCommand` instead of `app.handleCommand` |
| `theme/scheme.go` | Add `LabelHighlight` and `StatusSelected` fields |
| `theme/borland.go` | Set `LabelHighlight` and `StatusSelected` styles |

---

## Batch 1: Cluster Widget Fixes (Tasks 1–4)

### Task 1: Cluster Focus Indicator — Remove SfSelected Suppression

**Files:**
- Modify: `tv/checkbox.go` (CheckBoxes.Draw, lines 176–192)
- Modify: `tv/radio.go` (RadioButtons.Draw, lines 190–203)

**Requirements:**
- `CheckBoxes.Draw` must NOT suppress `SfSelected` on the focused child before drawing — the focused checkbox shows the `►` cursor prefix
- `RadioButtons.Draw` must NOT suppress `SfSelected` on the focused child — the focused radio button shows the `►` cursor prefix
- Only the internally-focused item within the cluster shows `►` — other items in the same cluster do not
- When the cluster itself is not focused (SfSelected not set on the cluster), the internal focus indicator still shows on whichever item was last focused within the cluster

**Implementation:**

Replace `CheckBoxes.Draw`:

```go
func (cbs *CheckBoxes) Draw(buf *DrawBuffer) {
	for _, item := range cbs.items {
		childBounds := item.Bounds()
		sub := buf.SubBuffer(childBounds)
		item.Draw(sub)
	}
}
```

Replace `RadioButtons.Draw`:

```go
func (rbs *RadioButtons) Draw(buf *DrawBuffer) {
	for _, item := range rbs.items {
		childBounds := item.Bounds()
		sub := buf.SubBuffer(childBounds)
		item.Draw(sub)
	}
}
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "fix: show focus indicator in cluster widgets"`

---

### Task 2: Arrow Key Navigation in CheckBoxes

**Files:**
- Modify: `tv/checkbox.go` (CheckBoxes struct and HandleEvent)

**Requirements:**
- Up arrow moves internal focus to the previous checkbox item (does NOT toggle)
- Down arrow moves internal focus to the next checkbox item (does NOT toggle)
- At the first item, Up does nothing (no wrap)
- At the last item, Down does nothing (no wrap)
- Arrow keys clear the event after handling
- Arrow navigation only moves focus — it does NOT change the checked state of any checkbox

**Implementation:**

Add `moveNavigation` method to CheckBoxes and arrow handling:

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
	if next < 0 || next >= len(cbs.items) {
		return
	}
	cbs.group.SetFocusedChild(cbs.items[next])
}
```

In `CheckBoxes.HandleEvent`, add before the group delegation:

```go
if event.What == EvKeyboard && event.Key != nil {
    switch event.Key.Key {
    case tcell.KeyDown:
        cbs.moveNavigation(1)
        event.Clear()
        return
    case tcell.KeyUp:
        cbs.moveNavigation(-1)
        event.Clear()
        return
    }
}
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: add arrow key navigation to CheckBoxes"`

---

### Task 3: Left/Right Arrow Keys in RadioButtons

**Files:**
- Modify: `tv/radio.go` (RadioButtons.HandleEvent)

**Requirements:**
- Left arrow moves selection to the previous radio button (same as Up)
- Right arrow moves selection to the next radio button (same as Down)
- At the first item, Left does nothing (no wrap, same as Up)
- At the last item, Right does nothing (no wrap, same as Down)
- Left/Right both change selection AND move focus (same behavior as Up/Down in RadioButtons)

**Implementation:**

In `RadioButtons.HandleEvent`, extend the arrow key handling:

```go
if event.What == EvKeyboard && event.Key != nil {
    switch event.Key.Key {
    case tcell.KeyDown, tcell.KeyRight:
        rbs.moveSelection(1)
        event.Clear()
        return
    case tcell.KeyUp, tcell.KeyLeft:
        rbs.moveSelection(-1)
        event.Clear()
        return
    }
}
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: add Left/Right arrow keys to RadioButtons"`

---

### Task 4: Remove Enter Key from CheckBox and RadioButton

**Files:**
- Modify: `tv/checkbox.go` (CheckBox.HandleEvent, lines 93–114)
- Modify: `tv/radio.go` (RadioButton.HandleEvent, lines 92–113)

**Requirements:**
- CheckBox.HandleEvent does NOT handle `tcell.KeyEnter` — the `case tcell.KeyEnter` block is removed entirely
- RadioButton.HandleEvent does NOT handle `tcell.KeyEnter` — the `case tcell.KeyEnter` block is removed entirely
- Space still toggles/selects (unchanged)
- Mouse click still toggles/selects (unchanged)
- Enter is now handled exclusively by Dialog's CmDefault broadcast (Phase 9)

**Implementation:**

CheckBox.HandleEvent — remove the Enter case:

```go
func (cb *CheckBox) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		if event.Mouse.Button == tcell.Button1 {
			cb.checked = !cb.checked
			event.Clear()
		}
		return
	}

	if event.What == EvKeyboard && event.Key != nil {
		if event.Key.Key == tcell.KeyRune && event.Key.Rune == ' ' {
			cb.checked = !cb.checked
			event.Clear()
		}
	}
}
```

RadioButton.HandleEvent — remove the Enter case:

```go
func (rb *RadioButton) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		if event.Mouse.Button == tcell.Button1 {
			rb.selectInCluster()
			event.Clear()
		}
		return
	}

	if event.What == EvKeyboard && event.Key != nil {
		if event.Key.Key == tcell.KeyRune && event.Key.Rune == ' ' {
			rb.selectInCluster()
			event.Clear()
		}
	}
}
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "fix: remove Enter handler from CheckBox and RadioButton"`

---

### Task 5: Integration Checkpoint — Cluster Behavior

**Purpose:** Verify that Tasks 1–4 work together: focus indicators show in clusters, arrow navigation works, Enter doesn't activate cluster items, and clusters work correctly inside real Dialogs.

**Requirements (for test writer):**
- In a Dialog with a CheckBoxes cluster: the focused checkbox shows `►` prefix when drawn
- In a Dialog with a RadioButtons cluster: the focused radio button shows `►` prefix when drawn
- Down arrow in CheckBoxes moves focus without toggling; Up arrow moves focus back
- Left/Right arrows in RadioButtons change selection same as Up/Down
- Enter in a Dialog with a CheckBoxes cluster does NOT toggle any checkbox — it fires the default button instead (via CmDefault broadcast)
- Enter in a Dialog with a RadioButtons cluster does NOT change selection — it fires the default button
- Tab from a cluster in a dialog navigates to the next sibling widget (e.g., a Button), not within the cluster's internal items

**Components to wire up:** Dialog, CheckBoxes, RadioButtons, Button (all real, no mocks), DrawBuffer for rendering assertions

**Run:** `go test ./tv/... -run TestIntegration -v`

---

## Batch 2: Label Behavior (Tasks 6–8)

### Task 6: Label Click to Focus Linked View

**Files:**
- Modify: `tv/label.go`

**Requirements:**
- When Label receives `EvMouse` with Button1 and has a linked view: call `owner.SetFocusedChild(link)` to focus the linked view
- Clear the event after focusing (the click is consumed)
- If the label has no linked view (`link == nil`), mouse events pass through unchanged
- If the label has no owner, mouse events pass through unchanged

**Implementation:**

Add mouse handling at the start of `Label.HandleEvent`:

```go
func (l *Label) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		if event.Mouse.Button&tcell.Button1 != 0 && l.link != nil {
			if owner := l.Owner(); owner != nil {
				owner.SetFocusedChild(l.link)
			}
			event.Clear()
		}
		return
	}

	if l.link == nil || l.shortcut == 0 {
		return
	}
	if event.What != EvKeyboard || event.Key == nil {
		return
	}
	if event.Key.Modifiers&tcell.ModAlt != 0 && event.Key.Key == tcell.KeyRune {
		if unicode.ToLower(event.Key.Rune) == unicode.ToLower(l.shortcut) {
			if owner := l.Owner(); owner != nil {
				owner.SetFocusedChild(l.link)
			}
			event.Clear()
		}
	}
}
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: Label click focuses linked view"`

---

### Task 7: Label Visual Focus Tracking

**Files:**
- Modify: `tv/label.go`
- Modify: `theme/scheme.go`
- Modify: `theme/borland.go`

**Requirements:**
- Label has a `light bool` field that tracks whether its linked view is currently focused
- When Label receives `EvBroadcast` with `CmReceivedFocus` and `Info` matches the linked view: set `light = true`
- When Label receives `EvBroadcast` with `CmReleasedFocus` and `Info` matches the linked view: set `light = false`
- When `light == true`, Label.Draw uses `LabelHighlight` style instead of `LabelNormal` for regular (non-shortcut) text
- When `light == false`, Label.Draw uses `LabelNormal` (existing behavior)
- The shortcut character always uses `LabelShortcut` regardless of light state
- `LabelHighlight` is a new field in `theme.ColorScheme`
- Borland Blue theme sets `LabelHighlight` to white on cyan (same as label but brighter — use `tcell.ColorWhite` foreground)

**Implementation:**

Add `light` field and broadcast handler to Label:

```go
type Label struct {
	BaseView
	label    string
	link     View
	shortcut rune
	light    bool
}
```

Add broadcast handling in HandleEvent (before existing keyboard handling):

```go
if event.What == EvBroadcast && l.link != nil {
    switch event.Command {
    case CmReceivedFocus:
        if event.Info == l.link {
            l.light = true
        }
    case CmReleasedFocus:
        if event.Info == l.link {
            l.light = false
        }
    }
    return
}
```

Update Draw to use light state:

```go
func (l *Label) Draw(buf *DrawBuffer) {
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	if cs := l.ColorScheme(); cs != nil {
		if l.light {
			normalStyle = cs.LabelHighlight
		} else {
			normalStyle = cs.LabelNormal
		}
		shortcutStyle = cs.LabelShortcut
	}

	x := 0
	segments := ParseTildeLabel(l.label)
	for _, seg := range segments {
		style := normalStyle
		if seg.Shortcut {
			style = shortcutStyle
		}
		buf.WriteStr(x, 0, seg.Text, style)
		x += utf8.RuneCountInString(seg.Text)
	}
}
```

Add to `theme/scheme.go`:

```go
LabelHighlight tcell.Style
```

Add to `theme/borland.go` (in the BorlandBlue initialization):

```go
LabelHighlight: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorCyan),
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: Label visual focus tracking via broadcasts"`

---

### Task 8: Label Alt+Shortcut Phase Change

**Files:**
- Modify: `tv/label.go` (NewLabel constructor)

**Requirements:**
- Label sets `OfPostProcess` instead of `OfPreProcess` in its constructor
- This means the focused view gets priority for Alt+shortcut keys before the label's handler runs
- Alt+shortcut still focuses the linked view (existing behavior, unchanged)
- Label no longer sets `OfPreProcess`

**Implementation:**

In `NewLabel`, change:

```go
l.SetOptions(OfPreProcess, true)
```

to:

```go
l.SetOptions(OfPostProcess, true)
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "fix: Label uses OfPostProcess for Alt+shortcut"`

---

### Task 9: Integration Checkpoint — Label Behavior

**Purpose:** Verify that Tasks 6–8 work together with real components.

**Requirements (for test writer):**
- In a Group with a Label linked to an InputLine: clicking the Label focuses the InputLine
- In a Group with a Label linked to a CheckBoxes cluster: clicking the Label focuses the cluster
- When the linked view gains focus (via SetFocusedChild), the Label's `light` field becomes true
- When focus moves away from the linked view, the Label's `light` field becomes false
- Label draws with LabelHighlight style when light=true (verify via DrawBuffer)
- Label's Alt+shortcut works in postprocess phase (focused view gets priority)
- In a Group where the focused view handles Alt+N: Label's Alt+N does NOT fire (postprocess means focused view wins)

**Components to wire up:** Group, Label, InputLine (or Button), DrawBuffer

**Run:** `go test ./tv/... -run TestIntegration -v`

---

## Batch 3: StatusLine, StaticText, Menu Dispatch (Tasks 10–13)

### Task 10: StatusLine Mouse Support

**Files:**
- Modify: `tv/status.go`
- Modify: `theme/scheme.go`
- Modify: `theme/borland.go`

**Requirements:**
- StatusLine.HandleEvent handles `EvMouse` events
- On mouse Button1 press: hit-test the X position against visible status items to find which item (if any) is under the cursor
- Hit-testing uses the same X positions as Draw (items are rendered left-to-right with gaps)
- On mouse release (Button1 not pressed, after a previous press): if the mouse is still over the same item, transform the event to `EvCommand` with that item's command
- If the mouse moves off the item before release, no command fires
- Status items that are filtered out by HelpCtx (not visible) cannot be clicked
- `StatusSelected` is a new field in `theme.ColorScheme` for the pressed highlight
- Borland Blue theme sets `StatusSelected` (e.g., black on white)

**Implementation:**

Add pressed state tracking to StatusLine:

```go
type StatusLine struct {
	BaseView
	items      []*StatusItem
	activeCtx  HelpContext
	pressedIdx int // -1 when not pressed
}

func NewStatusLine(items ...*StatusItem) *StatusLine {
	sl := &StatusLine{
		items:      items,
		pressedIdx: -1,
	}
	sl.SetState(SfVisible, true)
	return sl
}
```

Add a helper to compute item positions and hit-test:

```go
func (sl *StatusLine) itemRanges() []struct{ start, end int } {
	var ranges []struct{ start, end int }
	x := 1
	first := true
	for _, item := range sl.items {
		if item.HelpCtx != HcNoContext && item.HelpCtx != sl.activeCtx {
			ranges = append(ranges, struct{ start, end int }{-1, -1})
			continue
		}
		if !first {
			x += 2
		}
		first = false
		start := x
		segments := ParseTildeLabel(item.Label)
		for _, seg := range segments {
			x += utf8.RuneCountInString(seg.Text)
		}
		ranges = append(ranges, struct{ start, end int }{start, x})
	}
	return ranges
}

func (sl *StatusLine) itemAtX(mx int) int {
	ranges := sl.itemRanges()
	for i, r := range ranges {
		if r.start >= 0 && mx >= r.start && mx < r.end {
			return i
		}
	}
	return -1
}
```

Add mouse handling to HandleEvent:

```go
func (sl *StatusLine) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		realButtons := event.Mouse.Button & (tcell.Button1 | tcell.Button2 | tcell.Button3)
		if realButtons&tcell.Button1 != 0 {
			// Button pressed — track which item is under the cursor
			sl.pressedIdx = sl.itemAtX(event.Mouse.X)
			event.Clear()
		} else if realButtons != 0 && sl.pressedIdx >= 0 {
			// Mouse move while held — update highlight to track cursor position
			sl.pressedIdx = sl.itemAtX(event.Mouse.X)
			event.Clear()
		} else if sl.pressedIdx >= 0 {
			// Release — fire command if still over the same item
			idx := sl.itemAtX(event.Mouse.X)
			pressed := sl.pressedIdx
			sl.pressedIdx = -1
			if idx == pressed && pressed >= 0 && pressed < len(sl.items) {
				item := sl.items[pressed]
				if item.HelpCtx == HcNoContext || item.HelpCtx == sl.activeCtx {
					event.What = EvCommand
					event.Command = item.Command
					event.Mouse = nil
				}
			}
		}
		return
	}

	if event.What != EvKeyboard || event.Key == nil {
		return
	}
	for _, item := range sl.items {
		if item.HelpCtx != HcNoContext && item.HelpCtx != sl.activeCtx {
			continue
		}
		if item.KeyBinding.Matches(event.Key) {
			event.What = EvCommand
			event.Command = item.Command
			event.Key = nil
			return
		}
	}
}
```

Update Draw to show pressed highlight:

```go
func (sl *StatusLine) Draw(buf *DrawBuffer) {
	w := sl.Bounds().Width()
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault
	if cs := sl.ColorScheme(); cs != nil {
		normalStyle = cs.StatusNormal
		shortcutStyle = cs.StatusShortcut
		selectedStyle = cs.StatusSelected
	}

	buf.Fill(NewRect(0, 0, w, 1), ' ', normalStyle)

	x := 1
	first := true
	itemIdx := 0
	for _, item := range sl.items {
		if item.HelpCtx != HcNoContext && item.HelpCtx != sl.activeCtx {
			itemIdx++
			continue
		}
		if !first {
			x += 2
		}
		first = false

		isPressed := sl.pressedIdx == itemIdx
		segments := ParseTildeLabel(item.Label)
		for _, seg := range segments {
			style := normalStyle
			if isPressed {
				style = selectedStyle
			} else if seg.Shortcut {
				style = shortcutStyle
			}
			buf.WriteStr(x, 0, seg.Text, style)
			x += utf8.RuneCountInString(seg.Text)
		}
		itemIdx++
	}
}
```

Add to `theme/scheme.go`:

```go
StatusSelected tcell.Style
```

Add to `theme/borland.go`:

```go
StatusSelected: tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite),
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: StatusLine mouse click support"`

---

### Task 11: StaticText Centering

**Files:**
- Modify: `tv/static_text.go` (Draw method)

**Requirements:**
- When a line of text begins with `\x03` (byte value 3, ASCII ETX): center that line horizontally within the view width
- The `\x03` character is consumed (not displayed) — it is a formatting directive
- Centering applies only to the line it appears on — subsequent lines are left-aligned unless they also begin with `\x03`
- A `\n` resets to the next line (existing behavior); if the next line starts with `\x03`, it is also centered
- Centering calculation: `startX = (viewWidth - lineTextWidth) / 2`, clamped to 0
- Lines without `\x03` render left-aligned as before (no behavior change)
- Empty centered lines (`"\x03\n"`) render as blank lines

**Implementation:**

Modify `StaticText.Draw` to detect `\x03` at the start of each rendered line. Keep the existing `splitWords` word-wrapping logic but add centering support. The approach: process the text in two passes per visual line — first determine the words and wrapping to compute line content, then render with centering if the line starts with `\x03`.

```go
func (st *StaticText) Draw(buf *DrawBuffer) {
	w := st.Bounds().Width()
	h := st.Bounds().Height()
	if w <= 0 || h <= 0 {
		return
	}

	style := tcell.StyleDefault
	if cs := st.ColorScheme(); cs != nil {
		style = cs.LabelNormal
	}

	// Build visual lines preserving word-wrap semantics
	lines := wrapText(st.text, w)
	for y, line := range lines {
		if y >= h {
			break
		}
		centered := false
		runes := []rune(line)
		if len(runes) > 0 && runes[0] == '\x03' {
			centered = true
			runes = runes[1:]
		}
		x := 0
		if centered {
			x = (w - len(runes)) / 2
			if x < 0 {
				x = 0
			}
		}
		for _, r := range runes {
			if x < w {
				buf.WriteChar(x, y, r, style)
			}
			x++
		}
	}
}
```

Add `wrapText` helper that preserves word-wrapping and `\x03` markers. Keep `splitWords` unchanged — `wrapText` uses it internally:

```go
func wrapText(text string, width int) []string {
	// Split on explicit newlines first
	rawLines := splitOnNewlines(text)
	var result []string
	for _, raw := range rawLines {
		// Check for centering prefix
		prefix := ""
		content := raw
		if len(raw) > 0 && raw[0] == '\x03' {
			prefix = "\x03"
			content = raw[1:]
		}
		// Word-wrap the content
		words := splitWords(content)
		if len(words) == 0 {
			result = append(result, prefix)
			continue
		}
		line := prefix
		lineLen := 0
		for i, word := range words {
			wLen := len([]rune(word))
			if lineLen > 0 && lineLen+1+wLen > width {
				result = append(result, line)
				line = word
				lineLen = wLen
			} else {
				if lineLen > 0 {
					line += " "
					lineLen++
				}
				line += word
				lineLen += wLen
			}
		}
		result = append(result, line)
	}
	return result
}

func splitOnNewlines(s string) []string {
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	lines = append(lines, current)
	return lines
}
```

The old `splitWords` function stays unchanged — it is still used by `wrapText`.

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "feat: StaticText \\x03 line centering"`

---

### Task 12: Menu Command Dispatch via Event Queue

**Files:**
- Modify: `tv/menu_bar.go` (checkPopupResult method, lines 245–262)

**Requirements:**
- When a menu item is selected (result is not 0 and not CmCancel), the command must be posted to the application's event queue via `app.PostCommand(result, nil)` instead of calling `app.handleCommand` directly
- This ensures the command flows through the normal event dispatch chain: `app.handleEvent` → statusLine → desktop.HandleEvent → app.handleCommand
- Commands like CmTile and CmCascade (handled by Desktop) now reach the Desktop instead of being sent only to handleCommand
- CmQuit still works (reaches handleCommand after desktop doesn't handle it)
- CmCancel still closes the popup without dispatching (existing behavior)
- The menu bar's modal loop still exits (mb.active = false) before the posted command is processed — this is correct because the posted command will be picked up by the main Application.Run loop on its next iteration

**Implementation:**

Replace `checkPopupResult`:

```go
func (mb *MenuBar) checkPopupResult(app *Application) {
	if mb.popup == nil {
		return
	}
	result := mb.popup.Result()
	if result == 0 {
		return
	}
	if result == CmCancel {
		mb.closePopup()
		return
	}
	mb.closePopup()
	mb.active = false
	app.PostCommand(result, nil)
}
```

Also update `handleModalEvent` to not call `app.handleCommand` for non-CmMenu commands during the modal loop. This is intentional: the menu's modal loop should only handle CmMenu (to close itself). Any other command events that arrive during the modal loop (e.g., from mouse auto-repeat cleanup) are silently discarded — the important commands are the ones dispatched via `checkPopupResult`→`PostCommand`, which are processed by the main event loop after the modal exits:

```go
func (mb *MenuBar) handleModalEvent(event *Event, app *Application) {
	if event.What == EvCommand {
		if event.Command == CmMenu {
			mb.active = false
		}
		return
	}
	// ... rest unchanged
```

**Run tests:** `go test ./tv/... -count=1`

**Commit:** `git commit -m "fix: menu commands dispatch through event queue to reach Desktop"`

---

### Task 13: Integration Checkpoint — StatusLine, StaticText, Menu Dispatch

**Purpose:** Verify that Tasks 10–12 work together and with the rest of the system.

**Requirements (for test writer):**
- StatusLine mouse click on a visible status item fires the item's command (via event transformation)
- StatusLine mouse click on a hidden item (filtered by HelpCtx) does nothing
- StatusLine mouse press then release on a different item does not fire (drag-off cancels)
- StaticText with `"\x03Hello"` renders "Hello" centered in the view width
- StaticText with mixed centered and non-centered lines renders correctly
- Menu selection of CmTile via the real menu→popup→checkPopupResult path reaches Desktop.HandleEvent (verifiable by checking that Desktop.Tile() was called — test with real Application + Desktop + MenuBar)

**Components to wire up:** Application, Desktop, MenuBar, StatusLine, StaticText, DrawBuffer

**Run:** `go test ./tv/... -run TestIntegration -v`

---

## Batch 4: E2E Test Update

### Task 14: E2E Test — Widget Corrections

**Purpose:** Update the e2e test suite to verify Phase 10 functionality through the real tmux-driven application.

**Requirements:**
- Build the demo app binary
- Launch in tmux
- Test 1: In win1, focus on checkboxes cluster, verify `►` cursor is visible on the focused checkbox item
- Test 2: Open File menu → select Tile → verify windows rearrange (both window titles visible in a tiled layout)
- Test 3: Click on a StatusLine item (e.g., click on "Alt+X Exit" text area) — verify the app exits (status line mouse support)
- Test 4: All previous e2e tests still pass (regression)

**Update demo app if needed:** The demo app should have checkboxes visible in win1 (already does). Tile should be testable via menu (already has Window → Tile).

**Run:** `go test ./e2e/... -v -timeout 120s`

**Commit:** `git commit -m "test: e2e tests for widget corrections (Phase 10)"`
