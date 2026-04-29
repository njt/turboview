# Phase 3: Widgets, Focus Traversal & Dialog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add interactive widgets (Button, Label, StaticText), Tab/Shift+Tab focus traversal, and modal Dialog execution so the demo app can open a dialog with buttons and the user can Tab between them, press Enter to activate, and the dialog closes.

**Architecture:** Widgets embed `BaseView` and satisfy the `Widget` interface (marker). Focus traversal is added to `Group.HandleEvent` (intercepts Tab/Shift+Tab before three-phase dispatch). Dialog extends Window with modal execution via `ExecView`, which uses a new `EventSource` interface on Application to poll events in a nested loop. The entire dispatch chain (Application → Desktop → Window/Dialog → Button) is tested end-to-end.

**Tech Stack:** Go 1.22+, tcell/v2, existing tv/theme packages

---

## File Structure

| File | Responsibility | Status |
|------|----------------|--------|
| `tv/button.go` | Button widget with tilde label, shortcut, shadow, default-button behavior | Create |
| `tv/label.go` | Label widget with shortcut letter that focuses a linked view | Create |
| `tv/static_text.go` | StaticText widget — plain text display | Create |
| `tv/dialog.go` | Dialog (embeds BaseView + Group, modal with ExecView) | Create |
| `tv/group.go` | Add Tab/Shift+Tab focus traversal, replace ExecView panic | Modify |
| `tv/application.go` | Add EventSource (PollEvent), OnCommand callback | Modify |
| `tv/desktop.go` | No changes | — |
| `tv/window.go` | No changes (ExecView already delegates to group) | — |
| `e2e/testapp/basic/main.go` | Add dialog with buttons to demo app | Modify |
| `e2e/e2e_test.go` | Add dialog/button e2e tests | Modify |

---

### Task 1: Focus Traversal in Group

- [ ] Complete

**Files:**
- Modify: `tv/group.go`

**Requirements:**
- `Group.HandleEvent` intercepts `Tab` key (tcell.KeyTab, no modifiers) before the three-phase dispatch and advances focus to the next `OfSelectable` child (wrapping around). Clears the event.
- `Group.HandleEvent` intercepts `Shift+Tab` key (tcell.KeyBacktab) before the three-phase dispatch and moves focus to the previous `OfSelectable` child (wrapping around). Clears the event.
- `focusNext()` skips non-selectable children and wraps from end to beginning. If only one selectable child exists, it stays focused. If no selectable children exist, nothing happens.
- `focusPrev()` does the reverse — skips non-selectable children, wraps from beginning to end.
- Focus traversal uses `selectChild(v)` to set `SfSelected` on the new child and clear it on the old one — the same mechanism used by `Insert` and `BringToFront`.
- The Tab/Shift+Tab interception happens BEFORE the keyboard three-phase dispatch block (preprocess → focused → postprocess), so the focused child never sees the Tab event.
- Non-Tab keyboard events continue to use the existing three-phase dispatch unchanged.

**Implementation:**

```go
// tv/group.go — add to HandleEvent, before the three-phase dispatch block

func (g *Group) HandleEvent(event *Event) {
	if event.IsCleared() {
		return
	}

	// Mouse events: forward to focused child
	if event.What == EvMouse {
		if g.focused != nil {
			g.focused.HandleEvent(event)
		}
		return
	}

	// Broadcast: deliver to all children
	if event.What == EvBroadcast {
		for _, child := range g.children {
			if event.IsCleared() {
				return
			}
			child.HandleEvent(event)
		}
		return
	}

	// Tab/Shift+Tab focus traversal — before three-phase dispatch
	if event.What == EvKeyboard && event.Key != nil {
		if event.Key.Key == tcell.KeyTab && event.Key.Modifiers == 0 {
			g.focusNext()
			event.Clear()
			return
		}
		if event.Key.Key == tcell.KeyBacktab {
			g.focusPrev()
			event.Clear()
			return
		}
	}

	// Three-phase dispatch for keyboard and command events
	// (existing code unchanged)
	// Phase 1: Preprocess
	for _, child := range g.children {
		if event.IsCleared() {
			return
		}
		if child != g.focused && child.HasOption(OfPreProcess) {
			child.HandleEvent(event)
		}
	}

	// Phase 2: Focused
	if !event.IsCleared() && g.focused != nil {
		g.focused.HandleEvent(event)
	}

	// Phase 3: Postprocess
	for _, child := range g.children {
		if event.IsCleared() {
			return
		}
		if child != g.focused && child.HasOption(OfPostProcess) {
			child.HandleEvent(event)
		}
	}
}

func (g *Group) focusNext() {
	if len(g.children) == 0 {
		return
	}
	start := 0
	if g.focused != nil {
		for i, child := range g.children {
			if child == g.focused {
				start = i + 1
				break
			}
		}
	}
	n := len(g.children)
	for i := 0; i < n; i++ {
		idx := (start + i) % n
		if g.children[idx].HasOption(OfSelectable) && g.children[idx] != g.focused {
			g.selectChild(g.children[idx])
			return
		}
	}
}

func (g *Group) focusPrev() {
	if len(g.children) == 0 {
		return
	}
	start := len(g.children) - 1
	if g.focused != nil {
		for i, child := range g.children {
			if child == g.focused {
				start = i - 1
				break
			}
		}
	}
	n := len(g.children)
	for i := 0; i < n; i++ {
		idx := (start - i + n) % n
		if g.children[idx].HasOption(OfSelectable) && g.children[idx] != g.focused {
			g.selectChild(g.children[idx])
			return
		}
	}
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add Tab/Shift+Tab focus traversal to Group"`

---

### Task 2: StaticText Widget

- [ ] Complete

**Files:**
- Create: `tv/static_text.go`

**Requirements:**
- `StaticText` embeds `BaseView` and satisfies the `Widget` interface: `var _ Widget = (*StaticText)(nil)`.
- `NewStaticText(bounds Rect, text string) *StaticText` creates a StaticText with `SfVisible` set. It is NOT selectable (does not set `OfSelectable`).
- `StaticText.Text() string` returns the current text.
- `StaticText.SetText(t string)` updates the text.
- `StaticText.Draw(buf)` renders the text starting at (0, 0) within its bounds, wrapping to the next line at the bounds width. Uses `LabelNormal` style from `ColorScheme`. If the text exceeds the available area, it is clipped (DrawBuffer handles this naturally).
- `StaticText.HandleEvent(event)` is a no-op (BaseView default).
- Word wrapping splits on spaces: words that would exceed the line width start on the next line. A word longer than the line width is placed at the start of a line and clipped.

**Implementation:**

```go
// tv/static_text.go
package tv

import "github.com/gdamore/tcell/v2"

var _ Widget = (*StaticText)(nil)

type StaticText struct {
	BaseView
	text string
}

func NewStaticText(bounds Rect, text string) *StaticText {
	st := &StaticText{text: text}
	st.SetBounds(bounds)
	st.SetState(SfVisible, true)
	return st
}

func (st *StaticText) Text() string     { return st.text }
func (st *StaticText) SetText(t string) { st.text = t }

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

	x, y := 0, 0
	words := splitWords(st.text)
	for _, word := range words {
		runes := []rune(word)
		if x > 0 && x+len(runes) > w {
			x = 0
			y++
			if y >= h {
				return
			}
		}
		for _, r := range runes {
			if r == '\n' {
				x = 0
				y++
				if y >= h {
					return
				}
				continue
			}
			if x < w {
				buf.WriteChar(x, y, r, style)
			}
			x++
		}
		if x < w {
			x++ // space after word
		}
	}
}

func splitWords(s string) []string {
	var words []string
	current := ""
	for _, r := range s {
		if r == ' ' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add StaticText widget with word wrapping"`

---

### Task 3: Label Widget

- [ ] Complete

**Files:**
- Create: `tv/label.go`

**Requirements:**
- `Label` embeds `BaseView` and satisfies the `Widget` interface: `var _ Widget = (*Label)(nil)`.
- `NewLabel(bounds Rect, label string, link View) *Label` creates a Label. The `label` string uses tilde notation for the shortcut letter (e.g., `"~N~ame"`). `link` is the view to focus when the shortcut is activated. Sets `SfVisible`. NOT selectable.
- `Label.Draw(buf)` renders the label text with `LabelNormal` style for normal segments and `LabelShortcut` style for tilde-enclosed segments. Uses `ParseTildeLabel` (from `tv/tilde.go`).
- `Label.HandleEvent(event)` handles keyboard events: if the event is `Alt+<shortcut letter>` (where shortcut letter is the first rune of the tilde-enclosed segment), and `link` is not nil, the Label sets its owner's focused child to `link` by calling `owner.SetFocusedChild(link)`. Clears the event.
- The shortcut letter matching is case-insensitive.
- If `link` is nil, the Label renders normally but does not respond to keyboard events.
- Label sets `OfPreProcess` so it can intercept Alt+letter events before the focused child sees them.

**Implementation:**

```go
// tv/label.go
package tv

import (
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var _ Widget = (*Label)(nil)

type Label struct {
	BaseView
	label    string
	link     View
	shortcut rune
}

func NewLabel(bounds Rect, label string, link View) *Label {
	l := &Label{
		label: label,
		link:  link,
	}
	l.SetBounds(bounds)
	l.SetState(SfVisible, true)
	l.SetOptions(OfPreProcess, true)

	// Extract shortcut letter from tilde notation
	segments := ParseTildeLabel(label)
	for _, seg := range segments {
		if seg.Shortcut && len(seg.Text) > 0 {
			l.shortcut, _ = utf8.DecodeRuneInString(seg.Text)
			break
		}
	}

	return l
}

func (l *Label) Draw(buf *DrawBuffer) {
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	if cs := l.ColorScheme(); cs != nil {
		normalStyle = cs.LabelNormal
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

func (l *Label) HandleEvent(event *Event) {
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

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add Label widget with tilde shortcut and focus linking"`

---

### Task 4: Button Widget

- [ ] Complete

**Files:**
- Create: `tv/button.go`

**Requirements:**
- `Button` embeds `BaseView` and satisfies the `Widget` interface: `var _ Widget = (*Button)(nil)`.
- `NewButton(bounds Rect, title string, command CommandCode, opts ...ButtonOption) *Button` creates a Button. Sets `SfVisible` and `OfSelectable`. The `title` uses tilde notation for shortcut letters.
- `ButtonOption` is a function type `func(*Button)`. `WithDefault()` is a ButtonOption that marks the button as the default button (sets `OfPostProcess` and an internal `isDefault` flag).
- `Button.Draw(buf)` renders as `[ Title ]` (space-bracket-space pattern) centered within bounds:
  - Fill background with `ButtonNormal` style (or `ButtonDefault` style if `isDefault`).
  - Draw the bracket text `[ ]` around the title. Title segments use tilde parsing — shortcut letters render in `ButtonShortcut` style.
  - If the button has focus (`SfSelected` state), draw a `►` cursor at position (0, 0) to indicate focus.
  - Draw a 1-cell shadow below and to the right using `ButtonShadow` style: right column at (width-1, 1..height) and bottom row at (1, height-1..height), but only if bounds height >= 2.
- `Button.HandleEvent(event)` handles:
  - `Enter` key (tcell.KeyEnter): fires the command by setting `event.What = EvCommand`, `event.Command = b.command`, `event.Key = nil`.
  - `Space` key (tcell.KeyRune, rune ' '): same as Enter — fires the command.
  - Mouse click (EvMouse, Button1): fires the command (same transformation as Enter).
  - As a default button with `OfPostProcess`: responds to `Enter` in the postprocess phase (the focused child gets first crack; if it doesn't consume Enter, the default button fires).
- The shortcut letter in the title (tilde-enclosed) is handled by Label or StatusLine via `OfPreProcess`, not by the Button itself. The Button only responds to Enter/Space/Click when it has focus, or Enter via postprocess if it's the default.

**Implementation:**

```go
// tv/button.go
package tv

import (
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var _ Widget = (*Button)(nil)

type Button struct {
	BaseView
	title     string
	command   CommandCode
	isDefault bool
}

type ButtonOption func(*Button)

func WithDefault() ButtonOption {
	return func(b *Button) {
		b.isDefault = true
		b.SetOptions(OfPostProcess, true)
	}
}

func NewButton(bounds Rect, title string, command CommandCode, opts ...ButtonOption) *Button {
	b := &Button{
		title:   title,
		command: command,
	}
	b.SetBounds(bounds)
	b.SetState(SfVisible, true)
	b.SetOptions(OfSelectable, true)
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *Button) Title() string        { return b.title }
func (b *Button) Command() CommandCode { return b.command }
func (b *Button) IsDefault() bool      { return b.isDefault }

func (b *Button) Draw(buf *DrawBuffer) {
	w, h := b.Bounds().Width(), b.Bounds().Height()
	if w < 4 || h < 1 {
		return
	}

	cs := b.ColorScheme()
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	shadowStyle := tcell.StyleDefault
	if cs != nil {
		if b.isDefault {
			normalStyle = cs.ButtonDefault
		} else {
			normalStyle = cs.ButtonNormal
		}
		shortcutStyle = cs.ButtonShortcut
		shadowStyle = cs.ButtonShadow
	}

	// Button face area (excluding shadow)
	faceW := w
	faceH := h
	if h >= 2 {
		faceW = w - 1
		faceH = h - 1
	}

	// Fill face
	buf.Fill(NewRect(0, 0, faceW, faceH), ' ', normalStyle)

	// Focus cursor
	if b.HasState(SfSelected) {
		buf.WriteChar(0, 0, '►', normalStyle)
	}

	// Bracket and title: "[ Title ]"
	segments := ParseTildeLabel(b.title)
	titleLen := 0
	for _, seg := range segments {
		titleLen += utf8.RuneCountInString(seg.Text)
	}
	bracketText := titleLen + 4 // "[ " + title + " ]"
	startX := (faceW - bracketText) / 2
	if startX < 1 {
		startX = 1
	}

	buf.WriteChar(startX, 0, '[', normalStyle)
	buf.WriteChar(startX+1, 0, ' ', normalStyle)
	x := startX + 2
	for _, seg := range segments {
		style := normalStyle
		if seg.Shortcut {
			style = shortcutStyle
		}
		buf.WriteStr(x, 0, seg.Text, style)
		x += utf8.RuneCountInString(seg.Text)
	}
	buf.WriteChar(x, 0, ' ', normalStyle)
	buf.WriteChar(x+1, 0, ']', normalStyle)

	// Shadow (if height >= 2)
	if h >= 2 {
		// Right shadow column
		for y := 1; y < h; y++ {
			buf.WriteChar(w-1, y, ' ', shadowStyle)
		}
		// Bottom shadow row
		for x := 1; x < w; x++ {
			buf.WriteChar(x, h-1, ' ', shadowStyle)
		}
	}
}

func (b *Button) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		if event.Mouse.Button&tcell.Button1 != 0 {
			b.press(event)
		}
		return
	}

	if event.What == EvKeyboard && event.Key != nil {
		switch event.Key.Key {
		case tcell.KeyEnter:
			b.press(event)
		case tcell.KeyRune:
			if event.Key.Rune == ' ' {
				b.press(event)
			}
		}
	}
}

func (b *Button) press(event *Event) {
	event.What = EvCommand
	event.Command = b.command
	event.Key = nil
	event.Mouse = nil
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add Button widget with tilde shortcuts, shadow, and default-button support"`

---

### Task 5: Integration Checkpoint — Widgets + Focus

- [ ] Complete

**Purpose:** Verify that Button, Label, StaticText, and Tab focus traversal work together inside a real Window, with events flowing through the full Application → Desktop → Window → Widget chain.

**Requirements (for test writer):**
- A Button inserted into a Window, inside a Desktop, inside an Application: injecting an Enter key via `app.handleEvent` reaches the Button and fires its command (transforms event to EvCommand)
- Two Buttons in a Window: Tab key advances focus from button 1 to button 2. The newly focused button has `SfSelected`. The previously focused button loses `SfSelected`.
- Two Buttons in a Window: Shift+Tab moves focus backward (from button 2 to button 1).
- Tab with three items (Label non-selectable, Button A selectable, Button B selectable): Tab skips the Label and cycles only between the two buttons.
- Tab wraps around: with two buttons, Tab from the last goes back to the first.
- A default Button (WithDefault) in a Window responds to Enter in the postprocess phase even when a different widget has focus — meaning Enter pressed on a non-button focused view causes the default button to fire.
- A Label with `link` set to a Button: Alt+shortcut letter focuses the linked Button (Label is OfPreProcess, so it intercepts the Alt event before the focused child).
- StaticText inside a Window is drawn at its position with word-wrapped text.
- Drawing an Application containing a Window with a Button renders the button's bracket text `[ OK ]` in the correct position within the window's client area.

**Components to wire up:** Application, Desktop, Window, Button, Label, StaticText (all real, no mocks)

**Run tests:** `go test ./tv/... -v -run TestIntegration -count=1`

**Commit:** `git commit -m "test: add Phase 3 widget and focus integration tests"`

---

### Task 6: Application EventSource + OnCommand Callback

- [ ] Complete

**Files:**
- Modify: `tv/application.go`
- Modify: `tv/desktop.go`

**Requirements:**
- `Application` implements the `EventSource` interface by providing a `PollEvent() *Event` method.
- `PollEvent` loops on `app.screen.PollEvent()`, converts each tcell event via `convertEvent`, and returns the first non-nil result. Returns nil only when the screen is finalized (tcell returns nil). This loop ensures unrecognized tcell event types are silently skipped rather than terminating the event loop.
- Resize events trigger `app.layoutChildren()` inside PollEvent before returning the event.
- `Application.Run()` is refactored to use `app.PollEvent()` instead of calling `app.screen.PollEvent()` + `app.convertEvent()` directly.
- The refactored `Run()` must produce identical behavior to the existing implementation — all existing tests must continue to pass.
- `WithOnCommand(fn func(CommandCode, any) bool)` is a new `AppOption` that registers a callback for handling custom commands. The callback receives the command code and info, returns true if it handled the command (event should be cleared).
- In `handleCommand`, after the built-in `CmQuit` check, if the event is still not cleared and `onCommand` is set, call `app.onCommand(event.Command, event.Info)`. If it returns true, clear the event.
- **Desktop.app field:** Add an `app *Application` field to Desktop. Set it in `NewApplication` after creating Desktop: `app.desktop.app = app`. This enables `Group.ExecView` to find the Application by walking the owner chain up to Desktop and accessing `desktop.app`. Without this, the owner chain stops at Desktop (whose `Owner()` is nil since Application is not a Container) and ExecView would never find the EventSource.
- This enables the demo app (and any real app) to handle custom commands like `CmUser` without modifying the Application type.

**Implementation:**

```go
// tv/application.go — add to Application struct
type Application struct {
	bounds     Rect
	screen     tcell.Screen
	screenOwn  bool
	desktop    *Desktop
	statusLine *StatusLine
	scheme     *theme.ColorScheme
	quit       bool
	onCommand  func(CommandCode, any) bool
}

// add to AppOption functions
func WithOnCommand(fn func(CommandCode, any) bool) AppOption {
	return func(app *Application) {
		app.onCommand = fn
	}
}

// tv/desktop.go — add app field
type Desktop struct {
	BaseView
	group   *Group
	pattern rune
	app     *Application
}

// tv/application.go — in NewApplication, after creating desktop:
// app.desktop.app = app

// add PollEvent method — loops to skip unrecognized tcell events
func (app *Application) PollEvent() *Event {
	for {
		tcellEv := app.screen.PollEvent()
		if tcellEv == nil {
			return nil // screen finalized
		}
		if _, ok := tcellEv.(*tcell.EventResize); ok {
			w, h := app.screen.Size()
			app.bounds = NewRect(0, 0, w, h)
			app.layoutChildren()
		}
		if event := app.convertEvent(tcellEv); event != nil {
			return event
		}
		// Unknown event type — skip and poll again
	}
}

// refactor Run to use PollEvent
func (app *Application) Run() error {
	if app.screenOwn {
		if err := app.screen.Init(); err != nil {
			return err
		}
		defer app.screen.Fini()
	}

	app.screen.EnableMouse()
	app.screen.Clear()

	w, h := app.screen.Size()
	app.bounds = NewRect(0, 0, w, h)
	app.layoutChildren()
	app.drawAndFlush()

	for !app.quit {
		event := app.PollEvent()
		if event == nil {
			break
		}
		app.handleEvent(event)
		app.drawAndFlush()
	}

	return nil
}

// update handleCommand to call onCommand callback
func (app *Application) handleCommand(event *Event) {
	if event.What == EvCommand {
		switch event.Command {
		case CmQuit:
			app.quit = true
			event.Clear()
			return
		}
		if app.onCommand != nil {
			if app.onCommand(event.Command, event.Info) {
				event.Clear()
			}
		}
	}
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add Application.PollEvent() and OnCommand callback"`

---

### Task 7: Dialog with Modal ExecView

- [ ] Complete

**Files:**
- Create: `tv/dialog.go`
- Modify: `tv/group.go` (ExecView implementation)

**Requirements:**
- `Dialog` embeds `BaseView` directly (NOT Window — the spec requires "no embedding chains") and holds its own `*Group`. Satisfies Container: `var _ Container = (*Dialog)(nil)`.
- `NewDialog(bounds Rect, title string, opts ...DialogOption) *Dialog` creates a Dialog. Sets `SfVisible` and `OfSelectable`. Creates its own internal Group with facade set to the Dialog. Client area is `(width-2, height-2)`.
- `DialogOption` is a function type `func(*Dialog)`.
- Dialog implements all Container methods by delegating to its internal Group: `Insert`, `Remove`, `Children`, `FocusedChild`, `SetFocusedChild`, `ExecView`, plus `BringToFront`.
- `Dialog.SetBounds(r Rect)` updates both BaseView bounds and the internal Group bounds (client area).
- `Dialog.Title() string` returns the dialog title.
- `Dialog.Draw(buf)` draws a double-line frame using `DialogFrame` style, fills client area with `DialogBackground`, draws the title centered in the top border using `WindowTitle` style. No close/zoom icons (dialogs are dismissed via buttons, not frame icons). Draws children via the internal Group in a client-area sub-buffer.
- `Dialog.HandleEvent(event)` delegates keyboard/command events to its internal Group (which provides three-phase dispatch and Tab traversal). Mouse events are routed to the client area: coordinates are translated by (-1, -1) for the frame offset, then forwarded to the Group.
- **ExecView implementation in Group:**
  - `Group.ExecView(v View) CommandCode` implements modal execution (replaces the current `panic`):
    1. Inserts `v` into the Group
    2. Sets `SfModal` on `v`
    3. Walks up the `Owner()` chain from `g.facade` to find a `*Desktop`, then accesses `desktop.app` to get the `*Application`. This works because Application is NOT a Container (it doesn't implement View), so the owner chain naturally stops at Desktop. Desktop stores a direct `app *Application` reference set during `NewApplication` (added in Task 6).
    4. Enters a nested event loop: calls `app.PollEvent()`, routes the event to `v.HandleEvent(event)`, calls `app.drawAndFlush()`, checks if the event became a closing command
    5. On closing command (`CmOK`, `CmCancel`, `CmClose`, `CmYes`, `CmNo`): removes `v`, clears `SfModal`, returns the command code
    6. If PollEvent returns nil, removes `v`, clears `SfModal`, returns `CmCancel`
  - Window.ExecView already delegates to `w.group.ExecView(v)` — no change needed (the panic was in Group, not Window).
- **Modal event routing:** The nested event loop routes ALL events to the modal view only. Keyboard events go through `v.HandleEvent(event)` which (for Dialog) reaches the Dialog's Group three-phase dispatch, so Tab traversal and default-button postprocess work inside dialogs. Mouse events inside the modal view's bounds are translated and forwarded; mouse events outside are discarded.
- **Note:** The modal loop intentionally bypasses the Application's status line and desktop event routing. This is correct per spec: "all focused events are routed only to the modal view and its children." Status line shortcuts (like F2, Alt+X) do not function while a modal dialog is open.

**Implementation:**

```go
// tv/dialog.go
package tv

import "github.com/gdamore/tcell/v2"

var _ Container = (*Dialog)(nil)

type Dialog struct {
	BaseView
	group *Group
	title string
}

type DialogOption func(*Dialog)

func NewDialog(bounds Rect, title string, opts ...DialogOption) *Dialog {
	d := &Dialog{
		title: title,
	}
	d.SetBounds(bounds)
	d.SetState(SfVisible, true)
	d.SetOptions(OfSelectable, true)

	cw := max(bounds.Width()-2, 0)
	ch := max(bounds.Height()-2, 0)
	d.group = NewGroup(NewRect(0, 0, cw, ch))
	d.group.SetFacade(d)

	for _, opt := range opts {
		opt(d)
	}
	return d
}

func (d *Dialog) Title() string { return d.title }

func (d *Dialog) SetBounds(r Rect) {
	d.BaseView.SetBounds(r)
	if d.group != nil {
		cw := max(r.Width()-2, 0)
		ch := max(r.Height()-2, 0)
		d.group.SetBounds(NewRect(0, 0, cw, ch))
	}
}

func (d *Dialog) Insert(v View)               { d.group.Insert(v) }
func (d *Dialog) Remove(v View)               { d.group.Remove(v) }
func (d *Dialog) Children() []View            { return d.group.Children() }
func (d *Dialog) FocusedChild() View          { return d.group.FocusedChild() }
func (d *Dialog) SetFocusedChild(v View)      { d.group.SetFocusedChild(v) }
func (d *Dialog) ExecView(v View) CommandCode { return d.group.ExecView(v) }

func (d *Dialog) Draw(buf *DrawBuffer) {
	width, height := d.Bounds().Width(), d.Bounds().Height()
	if width < 8 || height < 3 {
		return
	}

	cs := d.ColorScheme()
	frameStyle := tcell.StyleDefault
	bgStyle := tcell.StyleDefault
	titleStyle := tcell.StyleDefault
	if cs != nil {
		frameStyle = cs.DialogFrame
		bgStyle = cs.DialogBackground
		titleStyle = cs.WindowTitle
	}

	// Client area background
	buf.Fill(NewRect(1, 1, width-2, height-2), ' ', bgStyle)

	// Frame — always double-line for dialogs
	buf.WriteChar(0, 0, '╔', frameStyle)
	buf.WriteChar(width-1, 0, '╗', frameStyle)
	buf.WriteChar(0, height-1, '╚', frameStyle)
	buf.WriteChar(width-1, height-1, '╝', frameStyle)
	for x := 1; x < width-1; x++ {
		buf.WriteChar(x, 0, '═', frameStyle)
		buf.WriteChar(x, height-1, '═', frameStyle)
	}
	for y := 1; y < height-1; y++ {
		buf.WriteChar(0, y, '║', frameStyle)
		buf.WriteChar(width-1, y, '║', frameStyle)
	}

	// Title centered in top border
	if len(d.title) > 0 {
		runes := []rune(d.title)
		availW := width - 2
		if len(runes) > availW-2 {
			runes = runes[:availW-2]
		}
		padded := " " + string(runes) + " "
		runeLen := len([]rune(padded))
		titleX := 1 + (availW-runeLen)/2
		if titleX < 1 {
			titleX = 1
		}
		buf.WriteStr(titleX, 0, padded, titleStyle)
	}

	// Draw children in client area
	clientBuf := buf.SubBuffer(NewRect(1, 1, width-2, height-2))
	d.group.Draw(clientBuf)
}

func (d *Dialog) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		width, height := d.Bounds().Width(), d.Bounds().Height()
		mx, my := event.Mouse.X, event.Mouse.Y
		// Client area: forward to group with frame offset
		if mx > 0 && mx < width-1 && my > 0 && my < height-1 {
			event.Mouse.X -= 1
			event.Mouse.Y -= 1
			d.group.HandleEvent(event)
		}
		return
	}
	d.group.HandleEvent(event)
}
```

```go
// tv/group.go — replace ExecView panic with implementation

func (g *Group) ExecView(v View) CommandCode {
	g.Insert(v)
	v.SetState(SfModal, true)

	// Walk owner chain from facade to find Application via Desktop.
	// Application is NOT a Container, so the chain stops at Desktop.
	// Desktop stores a direct reference to its Application (desktop.app).
	var app *Application
	var current Container = g.facade
	if current == nil {
		current = Container(g)
	}
	for current != nil {
		if d, ok := current.(*Desktop); ok && d.app != nil {
			app = d.app
			break
		}
		if view, ok := current.(View); ok {
			current = view.Owner()
		} else {
			break
		}
	}

	if app == nil {
		g.Remove(v)
		v.SetState(SfModal, false)
		return CmCancel
	}

	// Modal event loop
	var result CommandCode
	for {
		event := app.PollEvent()
		if event == nil {
			result = CmCancel
			break
		}

		// Route event to modal view only
		if event.What == EvMouse && event.Mouse != nil {
			vb := v.Bounds()
			mx, my := event.Mouse.X, event.Mouse.Y
			if vb.Contains(NewPoint(mx, my)) {
				event.Mouse.X -= vb.A.X
				event.Mouse.Y -= vb.A.Y
				v.HandleEvent(event)
			}
			// Outside clicks are discarded
		} else {
			v.HandleEvent(event)
		}

		// Check for closing command (Button.press transforms event in place)
		if event.What == EvCommand {
			switch event.Command {
			case CmOK, CmCancel, CmClose, CmYes, CmNo:
				result = event.Command
			}
		}

		app.drawAndFlush()

		if result != 0 {
			break
		}
	}

	g.Remove(v)
	v.SetState(SfModal, false)
	return result
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add Dialog with modal ExecView support"`

---

### Task 8: Standard Dialog Functions (MessageBox)

- [ ] Complete

**Files:**
- Modify: `tv/dialog.go`

**Requirements:**
- `MessageBox(owner Container, title, text string, buttons MsgBoxButton) CommandCode` creates a Dialog, inserts a StaticText with `text`, inserts buttons corresponding to the `buttons` bitmask, calls `owner.ExecView(dialog)`, and returns the result command code.
- `MsgBoxButton` is a bitmask type:
  - `MbOK MsgBoxButton = 1 << iota` → creates an "[ OK ]" button that fires `CmOK`
  - `MbCancel` → "[ Cancel ]" button that fires `CmCancel`
  - `MbYes` → "[ Yes ]" button that fires `CmYes`
  - `MbNo` → "[ No ]" button that fires `CmNo`
- Button labels do not use tilde shortcuts in MessageBox (plain text).
- The Dialog is auto-sized: width is `max(len(title)+4, len(text)+4, buttonRowWidth+4)` clamped to 60. Height is `textLines + 5` (frame + text + gap + buttons + frame). The dialog is centered in the owner's bounds.
- Buttons are arranged in a horizontal row centered at the bottom of the dialog (1 row above the bottom frame). Each button is 12 wide and 2 tall (face + shadow). Buttons are separated by 2 spaces.
- The first button in the set is the default button (marked with `WithDefault()`).
- If `buttons` is `MbOK|MbCancel`, two buttons appear: `[ OK ]` (default) and `[ Cancel ]`.

**Implementation:**

```go
// tv/dialog.go — add MessageBox and MsgBoxButton

type MsgBoxButton int

const (
	MbOK     MsgBoxButton = 1 << iota
	MbCancel
	MbYes
	MbNo
)

func MessageBox(owner Container, title, text string, buttons MsgBoxButton) CommandCode {
	type btnDef struct {
		label string
		cmd   CommandCode
	}
	var defs []btnDef
	if buttons&MbYes != 0 {
		defs = append(defs, btnDef{"Yes", CmYes})
	}
	if buttons&MbNo != 0 {
		defs = append(defs, btnDef{"No", CmNo})
	}
	if buttons&MbOK != 0 {
		defs = append(defs, btnDef{"OK", CmOK})
	}
	if buttons&MbCancel != 0 {
		defs = append(defs, btnDef{"Cancel", CmCancel})
	}
	if len(defs) == 0 {
		defs = append(defs, btnDef{"OK", CmOK})
	}

	// Auto-size
	textRunes := []rune(text)
	textW := len(textRunes)
	btnW := 12
	btnGap := 2
	buttonRowW := len(defs)*btnW + (len(defs)-1)*btnGap
	contentW := textW
	if buttonRowW > contentW {
		contentW = buttonRowW
	}
	titleW := len([]rune(title)) + 4
	if titleW > contentW {
		contentW = titleW
	}
	dialogW := contentW + 4 // 2 for frame + 2 for padding
	if dialogW > 60 {
		dialogW = 60
	}
	if dialogW < 20 {
		dialogW = 20
	}

	// Text wrapping height
	innerW := dialogW - 4
	textLines := 1
	lineLen := 0
	for _, r := range textRunes {
		if r == '\n' {
			textLines++
			lineLen = 0
			continue
		}
		lineLen++
		if lineLen > innerW {
			textLines++
			lineLen = 1
		}
	}

	dialogH := textLines + 5 // top frame + text rows + gap + button row + bottom frame
	if dialogH < 7 {
		dialogH = 7
	}

	// Center in owner
	ob := owner.Bounds()
	dx := (ob.Width() - dialogW) / 2
	dy := (ob.Height() - dialogH) / 2
	if dx < 0 {
		dx = 0
	}
	if dy < 0 {
		dy = 0
	}

	dlg := NewDialog(NewRect(dx, dy, dialogW, dialogH), title)

	// Insert static text
	st := NewStaticText(NewRect(1, 0, innerW, textLines), text)
	dlg.Insert(st)

	// Insert buttons
	btnY := textLines + 1
	totalBtnW := len(defs)*btnW + (len(defs)-1)*btnGap
	startX := (innerW - totalBtnW) / 2
	if startX < 0 {
		startX = 0
	}
	for i, def := range defs {
		x := startX + i*(btnW+btnGap)
		var opts []ButtonOption
		if i == 0 {
			opts = append(opts, WithDefault())
		}
		btn := NewButton(NewRect(x, btnY, btnW, 2), def.label, def.cmd, opts...)
		dlg.Insert(btn)
	}

	return owner.ExecView(dlg)
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add MessageBox standard dialog function"`

---

### Task 9: Integration Checkpoint — Dialog + ExecView

- [ ] Complete

**Purpose:** Verify that the full modal dialog flow works: Application → Desktop → ExecView → Dialog → Buttons → command result.

**Requirements (for test writer):**
- Creating an Application with a Desktop, calling `desktop.ExecView(dialog)` where `dialog` contains an OK button: pressing Enter returns `CmOK`.
- `desktop.ExecView(dialog)` with OK and Cancel buttons: pressing Tab moves focus to Cancel, pressing Enter returns `CmCancel`.
- `desktop.ExecView(dialog)` with a default button: pressing Enter when a non-button view is focused fires the default button via postprocess (returns the default button's command).
- `MessageBox(desktop, "Title", "Are you sure?", MbYes|MbNo)` displays the dialog and returns `CmYes` when Enter is pressed (Yes is the default/first button).
- Dialog renders with `DialogFrame` style (not WindowFrameActive/Inactive) — always double-line border.
- Dialog renders with `DialogBackground` style for the client area.
- Mouse click inside the modal dialog on a button fires the command. Mouse click outside the dialog bounds is discarded (does not crash, does not dismiss).
- Tab traversal works inside the modal dialog (cycling between buttons).
- After ExecView returns, the dialog is no longer in the owner's children list (it was removed).

**Components to wire up:** Application (with SimulationScreen + PostCommand for injecting events), Desktop, Dialog, Button, StaticText

**Note on testing ExecView:** ExecView blocks in a nested event loop calling `app.PollEvent()`, so tests must inject events from a goroutine. Two approaches:
- `sim.InjectKey(tcell.KeyEnter, 0, 0)` — injects a real key event that flows through the modal dialog's three-phase dispatch (Button receives Enter → transforms to EvCommand). Use this for realistic button-press testing.
- `sim.PostEvent(cmdTcellEvent)` or `app.PostCommand(CmOK, nil)` — injects a command event directly. The modal loop's closing-command check catches it after `v.HandleEvent` returns. This is a shortcut that bypasses the Button widget entirely.
Both require a `time.Sleep` or channel sync before injecting to ensure ExecView has entered its polling loop.

**Run tests:** `go test ./tv/... -v -run TestIntegration -count=1`

**Commit:** `git commit -m "test: add Phase 3 dialog and ExecView integration tests"`

---

### Task 10: Demo App + E2E Tests

- [ ] Complete

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**
- Demo app adds a new status item `"~F2~ Dialog"` that fires `CmUser` (command code 1000).
- Demo app handles `CmUser` using `WithOnCommand` callback (from Task 6). The callback calls `MessageBox(app.Desktop(), "Confirm", "Exit the application?", MbYes|MbNo)`. If result is `CmYes`, posts `CmQuit`. Returns true to indicate the command was handled.
- The demo app's Window 1 ("File Manager") now contains a StaticText "Press F2 for dialog" and two Buttons: "[ OK ]" (CmOK) and "[ Close ]" (CmClose).
- E2E test verifies: pressing F2 opens a dialog (box-drawing characters `╔` for dialog frame, "Confirm" title text visible), pressing Enter dismisses the dialog (dialog characters disappear), the app remains running (can still see desktop pattern).
- E2E test verifies: button text `[ OK ]` appears inside the window (rendered by the Button widget).
- All existing E2E tests continue to pass.

**Implementation:**

```go
// e2e/testapp/basic/main.go — updated
package main

import (
	"log"

	"github.com/njt/turboview/theme"
	"github.com/njt/turboview/tv"
)

func main() {
	statusLine := tv.NewStatusLine(
		tv.NewStatusItem("~Alt+X~ Exit", tv.KbAlt('X'), tv.CmQuit),
		tv.NewStatusItem("~F2~ Dialog", tv.KbFunc(2), tv.CmUser),
		tv.NewStatusItem("~F10~ Menu", tv.KbFunc(10), tv.CmMenu),
	)

	var app *tv.Application
	var err error

	app, err = tv.NewApplication(
		tv.WithStatusLine(statusLine),
		tv.WithTheme(theme.BorlandBlue),
		tv.WithOnCommand(func(cmd tv.CommandCode, info any) bool {
			if cmd == tv.CmUser {
				result := tv.MessageBox(app.Desktop(), "Confirm", "Exit the application?", tv.MbYes|tv.MbNo)
				if result == tv.CmYes {
					app.PostCommand(tv.CmQuit, nil)
				}
				return true
			}
			return false
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	win1 := tv.NewWindow(tv.NewRect(5, 2, 35, 15), "File Manager", tv.WithWindowNumber(1))
	st := tv.NewStaticText(tv.NewRect(1, 1, 30, 1), "Press F2 for dialog")
	win1.Insert(st)
	btnOK := tv.NewButton(tv.NewRect(1, 3, 12, 2), "OK", tv.CmOK)
	win1.Insert(btnOK)
	btnClose := tv.NewButton(tv.NewRect(15, 3, 12, 2), "Close", tv.CmClose)
	win1.Insert(btnClose)

	win2 := tv.NewWindow(tv.NewRect(20, 5, 40, 12), "Editor", tv.WithWindowNumber(2))

	app.Desktop().Insert(win1)
	app.Desktop().Insert(win2)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
```

```go
// e2e/e2e_test.go — extend TestBasicAppBoot and add TestDialogFlow

func TestBasicAppBoot(t *testing.T) {
	// ... existing test unchanged, but now also checks for button text
	root := projectRoot()
	binPath := filepath.Join(root, "e2e", "testapp", "basic", "basic")

	out, err := exec.Command("go", "build", "-o", binPath, filepath.Join(root, "e2e", "testapp", "basic")).CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	t.Cleanup(func() { exec.Command("rm", binPath).Run() })

	session := "tv3-e2e-basic"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)
	lines := tmuxCapture(t, session)

	// Desktop pattern visible
	desktopHasPattern := false
	for _, line := range lines {
		if strings.Contains(line, "░") {
			desktopHasPattern = true
			break
		}
	}
	if !desktopHasPattern {
		t.Error("desktop background pattern '░' not found")
	}

	// Window frame characters visible
	frameFound := false
	for _, line := range lines {
		if strings.Contains(line, "╔") || strings.Contains(line, "═") {
			frameFound = true
			break
		}
	}
	if !frameFound {
		t.Error("window frame characters not found")
	}

	// Window title visible
	titleFound := false
	for _, line := range lines {
		if strings.Contains(line, "File Manager") || strings.Contains(line, "Editor") {
			titleFound = true
			break
		}
	}
	if !titleFound {
		t.Error("window title text not found")
	}

	// Button text visible inside window
	buttonFound := false
	for _, line := range lines {
		if strings.Contains(line, "[ OK ]") || strings.Contains(line, "[ Close ]") {
			buttonFound = true
			break
		}
	}
	if !buttonFound {
		t.Error("button text '[ OK ]' or '[ Close ]' not found in rendered output")
	}

	// Status line contains "Alt+X"
	statusFound := false
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			if strings.Contains(lines[i], "Alt+X") {
				statusFound = true
			}
			break
		}
	}
	if !statusFound {
		t.Error("status line should contain 'Alt+X'")
	}

	// Alt+X exits the app
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

func TestDialogFlow(t *testing.T) {
	root := projectRoot()
	binPath := filepath.Join(root, "e2e", "testapp", "basic", "basic")

	out, err := exec.Command("go", "build", "-o", binPath, filepath.Join(root, "e2e", "testapp", "basic")).CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	t.Cleanup(func() { exec.Command("rm", binPath).Run() })

	session := "tv3-e2e-dialog"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Press F2 to open dialog
	tmuxSendKeys(t, session, "F2")
	lines := tmuxCapture(t, session)

	// Dialog frame should appear
	dialogFrameFound := false
	for _, line := range lines {
		if strings.Contains(line, "Confirm") {
			dialogFrameFound = true
			break
		}
	}
	if !dialogFrameFound {
		t.Error("dialog title 'Confirm' not found after F2")
	}

	// Press Enter to dismiss dialog (Yes is default)
	tmuxSendKeys(t, session, "Enter")
	lines = tmuxCapture(t, session)

	// App should have exited (because Yes → CmQuit)
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after confirming dialog with Enter")
	}
}
```

**Run tests:** `go test ./e2e/... -v -count=1 -timeout 30s`

**Commit:** `git commit -m "feat: add dialog demo and e2e tests for Phase 3"`
