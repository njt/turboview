# Phase 4: Menu System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a pull-down menu system to the TUI framework — MenuBar at row 0, with keyboard and mouse-activated pull-down menus.

**Architecture:** MenuBar sits at Application row 0, holds SubMenu definitions. When activated (F10/CmMenu or click), MenuBar enters its own modal event loop (same pattern as ExecView but without using ExecView — the menu system manages its own loop to handle left/right navigation between menus). MenuPopup is a rendering/navigation component that draws a bordered dropdown and handles up/down/Enter/Escape. Application.Draw() renders the popup on top of the desktop after drawing it, ensuring the popup appears above window content.

**Tech Stack:** Go 1.22+, tcell/v2, existing tv framework types

---

## File Structure

| File | Responsibility |
|------|---------------|
| `tv/menu.go` | Menu data types (MenuItem, SubMenu, MenuSeparator), constructors, accelerator formatting |
| `tv/menu_popup.go` | MenuPopup struct, Draw (bordered dropdown), HandleEvent (keyboard/mouse navigation) |
| `tv/menu_bar.go` | MenuBar struct, Draw (horizontal bar), Activate (modal event loop), popup management |
| `tv/application.go` | Modified: add menuBar field, WithMenuBar option, update layout/draw/event routing |
| `e2e/testapp/basic/main.go` | Modified: add menu bar to demo app |
| `e2e/e2e_test.go` | Modified: add menu system e2e tests |

---

### Task 1: Menu Data Types

- [ ] Complete

**Files:**
- Create: `tv/menu.go`

**Requirements:**
- `MenuItem` struct has fields: `Label string`, `Command CommandCode`, `Accel KeyBinding`, `Disabled bool`.
- `NewMenuItem(label string, cmd CommandCode, accel KeyBinding) *MenuItem` creates a MenuItem. Label uses tilde notation for shortcut letters (e.g., `"~N~ew"`).
- `SubMenu` struct has fields: `Label string`, `Items []any`. Items can be `*MenuItem`, `*MenuSeparator`, or `*SubMenu` (nested).
- `NewSubMenu(label string, items ...any) *SubMenu` creates a SubMenu. Label uses tilde notation.
- `MenuSeparator` is a struct marker type (no fields).
- `NewMenuSeparator() *MenuSeparator` creates a separator.
- `FormatAccel(kb KeyBinding) string` returns a human-readable string for a KeyBinding:
  - `KbCtrl('N')` → `"Ctrl+N"`
  - `KbAlt('X')` → `"Alt+X"`
  - `KbFunc(10)` → `"F10"`
  - `KbNone()` → `""` (empty string)
  - For `KbCtrl`, the letter is always uppercase in the display string.
- `menuItemWidth(item *MenuItem) int` (unexported) returns the display width of a menu item: `runeCount(label without tildes) + 2 (gap) + runeCount(accel string)`. If accel is KbNone, no gap or accel is counted.
- `popupWidth(items []any) int` (unexported) returns the width needed for a popup: `max(menuItemWidth for all MenuItems) + 4` (2 for borders + 2 for left/right padding).
- `popupHeight(items []any) int` (unexported) returns the height: `len(items) + 2` (2 for top/bottom borders).

**Implementation:**

```go
// tv/menu.go
package tv

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type MenuItem struct {
	Label    string
	Command  CommandCode
	Accel    KeyBinding
	Disabled bool
}

func NewMenuItem(label string, cmd CommandCode, accel KeyBinding) *MenuItem {
	return &MenuItem{Label: label, Command: cmd, Accel: accel}
}

type SubMenu struct {
	Label string
	Items []any
}

func NewSubMenu(label string, items ...any) *SubMenu {
	return &SubMenu{Label: label, Items: items}
}

type MenuSeparator struct{}

func NewMenuSeparator() *MenuSeparator {
	return &MenuSeparator{}
}

func FormatAccel(kb KeyBinding) string {
	if kb == (KeyBinding{}) {
		return ""
	}
	// KbCtrl stores Key as tcell.Key(ch - 'A' + 1), not KeyRune
	if kb.Mod&tcell.ModCtrl != 0 && kb.Key >= 1 && kb.Key <= 26 {
		ch := rune(kb.Key) + 'A' - 1
		return fmt.Sprintf("Ctrl+%c", ch)
	}
	if kb.Mod&tcell.ModAlt != 0 && kb.Key == tcell.KeyRune {
		return fmt.Sprintf("Alt+%c", unicode.ToUpper(kb.Rune))
	}
	if kb.Key >= tcell.KeyF1 && kb.Key <= tcell.KeyF12 {
		n := int(kb.Key-tcell.KeyF1) + 1
		return fmt.Sprintf("F%d", n)
	}
	return ""
}

func tildeTextLen(label string) int {
	segments := ParseTildeLabel(label)
	n := 0
	for _, seg := range segments {
		n += utf8.RuneCountInString(seg.Text)
	}
	return n
}

func menuItemWidth(item *MenuItem) int {
	w := tildeTextLen(item.Label)
	accel := FormatAccel(item.Accel)
	if accel != "" {
		w += 2 + utf8.RuneCountInString(accel)
	}
	return w
}

func popupWidth(items []any) int {
	maxW := 0
	for _, item := range items {
		if mi, ok := item.(*MenuItem); ok {
			w := menuItemWidth(mi)
			if w > maxW {
				maxW = w
			}
		}
	}
	return maxW + 4
}

func popupHeight(items []any) int {
	return len(items) + 2
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add menu data types and constructors"`

---

### Task 2: MenuPopup — Rendering and Navigation

- [ ] Complete

**Files:**
- Create: `tv/menu_popup.go`

**Requirements:**
- `MenuPopup` struct has fields: `items []any` (from SubMenu.Items), `selected int` (index of highlighted item, -1 if none), `result CommandCode` (command of selected item, 0 if not yet selected), `bounds Rect` (screen-coordinates position/size).
- `NewMenuPopup(items []any, x, y int) *MenuPopup` creates a popup positioned at screen coordinates (x, y). Width and height are auto-computed from items using `popupWidth` and `popupHeight`.
- `MenuPopup.Bounds() Rect` returns the popup's bounds in screen coordinates.
- `MenuPopup.Result() CommandCode` returns the selected command (0 if none).
- `MenuPopup.Draw(buf *DrawBuffer)` renders the popup:
  - Single-line border using `MenuNormal` style: `┌─┐│└─┘` characters.
  - Each `*MenuItem`: left-aligned label (tilde-parsed, shortcut in `MenuShortcut` style, rest in `MenuNormal`), right-aligned accelerator text in `MenuNormal` style. If the item is the selected item (`selected == index`), use `MenuSelected` style for the entire row. If item is disabled, use `MenuDisabled` style.
  - Each `*MenuSeparator`: horizontal line `├─┤` across the full popup width using `MenuNormal` style.
  - Background fill with `MenuNormal` style for all item rows.
  - Draw uses the popup's own bounds dimensions (not screen coordinates — the caller provides a SubBuffer).
- `MenuPopup.HandleEvent(event *Event)` handles keyboard and mouse:
  - **Down arrow**: moves `selected` to next non-separator item, wrapping around.
  - **Up arrow**: moves `selected` to previous non-separator item, wrapping around.
  - **Enter**: if `selected` points to a non-disabled MenuItem, sets `result` to that item's Command.
  - **Escape**: sets `result` to `CmCancel` (the caller checks `result != 0` to exit).
  - **Rune key**: if a menu item's tilde shortcut matches the rune (case-insensitive), selects and fires that item (sets `result`).
  - **Mouse click (Button1)**: if click is on an item row (y=1 to y=height-2 in popup-local coords), selects that item. If the item is a non-disabled MenuItem, fires it (sets `result`). Clicks on borders or separators are ignored.
  - **Mouse move (ButtonNone, no buttons pressed)**: if mouse is on an item row, highlight that item (update `selected`). This provides hover-highlight behavior.
- The initial `selected` is 0 (first item). If the first item is a separator, advance to the next non-separator.
- Items that are `*SubMenu` (nested menus) are NOT supported in Phase 4 — they are treated as disabled menu items showing the label with a `►` suffix. This defers nested menu complexity.

**Implementation:**

```go
// tv/menu_popup.go
package tv

import (
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type MenuPopup struct {
	items    []any
	selected int
	result   CommandCode
	bounds   Rect
}

func NewMenuPopup(items []any, x, y int) *MenuPopup {
	w := popupWidth(items)
	h := popupHeight(items)
	mp := &MenuPopup{
		items:    items,
		selected: -1,
		bounds:   NewRect(x, y, w, h),
	}
	mp.selectNext(true)
	return mp
}

func (mp *MenuPopup) Bounds() Rect       { return mp.bounds }
func (mp *MenuPopup) Result() CommandCode { return mp.result }
func (mp *MenuPopup) Selected() int       { return mp.selected }

func (mp *MenuPopup) Draw(buf *DrawBuffer, cs *theme.ColorScheme) {
	w, h := mp.bounds.Width(), mp.bounds.Height()
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault
	disabledStyle := tcell.StyleDefault
	if cs != nil {
		normalStyle = cs.MenuNormal
		shortcutStyle = cs.MenuShortcut
		selectedStyle = cs.MenuSelected
		disabledStyle = cs.MenuDisabled
	}

	// Border
	buf.WriteChar(0, 0, '┌', normalStyle)
	buf.WriteChar(w-1, 0, '┐', normalStyle)
	buf.WriteChar(0, h-1, '└', normalStyle)
	buf.WriteChar(w-1, h-1, '┘', normalStyle)
	for x := 1; x < w-1; x++ {
		buf.WriteChar(x, 0, '─', normalStyle)
		buf.WriteChar(x, h-1, '─', normalStyle)
	}
	for y := 1; y < h-1; y++ {
		buf.WriteChar(0, y, '│', normalStyle)
		buf.WriteChar(w-1, y, '│', normalStyle)
	}

	innerW := w - 2

	for i, item := range mp.items {
		row := i + 1 // skip top border
		switch it := item.(type) {
		case *MenuSeparator:
			buf.WriteChar(0, row, '├', normalStyle)
			for x := 1; x < w-1; x++ {
				buf.WriteChar(x, row, '─', normalStyle)
			}
			buf.WriteChar(w-1, row, '┤', normalStyle)

		case *MenuItem:
			isSelected := i == mp.selected
			style := normalStyle
			scStyle := shortcutStyle
			if it.Disabled {
				style = disabledStyle
				scStyle = disabledStyle
			} else if isSelected {
				style = selectedStyle
				scStyle = selectedStyle
			}

			// Fill row background
			buf.Fill(NewRect(1, row, innerW, 1), ' ', style)

			// Draw label with tilde parsing
			x := 1
			segments := ParseTildeLabel(it.Label)
			for _, seg := range segments {
				s := style
				if seg.Shortcut && !it.Disabled && !isSelected {
					s = scStyle
				}
				buf.WriteStr(x, row, seg.Text, s)
				x += utf8.RuneCountInString(seg.Text)
			}

			// Draw accelerator right-aligned
			accel := FormatAccel(it.Accel)
			if accel != "" {
				ax := 1 + innerW - utf8.RuneCountInString(accel)
				buf.WriteStr(ax, row, accel, style)
			}

		case *SubMenu:
			isSelected := i == mp.selected
			style := disabledStyle // nested menus not supported in Phase 4
			if isSelected {
				style = selectedStyle
			}
			buf.Fill(NewRect(1, row, innerW, 1), ' ', style)
			segments := ParseTildeLabel(it.Label)
			x := 1
			for _, seg := range segments {
				buf.WriteStr(x, row, seg.Text, style)
				x += utf8.RuneCountInString(seg.Text)
			}
			buf.WriteStr(1+innerW-1, row, "►", style)
		}
	}
}

func (mp *MenuPopup) HandleEvent(event *Event) {
	if event.What == EvKeyboard && event.Key != nil {
		switch event.Key.Key {
		case tcell.KeyDown:
			mp.selectNext(false)
		case tcell.KeyUp:
			mp.selectPrev()
		case tcell.KeyEnter:
			mp.fireSelected()
		case tcell.KeyEscape:
			mp.result = CmCancel
		case tcell.KeyRune:
			mp.matchShortcut(event.Key.Rune)
		}
		return
	}

	if event.What == EvMouse && event.Mouse != nil {
		my := event.Mouse.Y
		if my >= 1 && my < mp.bounds.Height()-1 {
			idx := my - 1
			if idx >= 0 && idx < len(mp.items) {
				if _, ok := mp.items[idx].(*MenuSeparator); !ok {
					mp.selected = idx
					if event.Mouse.Button&tcell.Button1 != 0 {
						mp.fireSelected()
					}
				}
			}
		}
	}
}

func (mp *MenuPopup) selectNext(initial bool) {
	n := len(mp.items)
	if n == 0 {
		return
	}
	start := mp.selected + 1
	if initial && mp.selected < 0 {
		start = 0
	}
	for i := 0; i < n; i++ {
		idx := (start + i) % n
		if _, ok := mp.items[idx].(*MenuSeparator); !ok {
			mp.selected = idx
			return
		}
	}
}

func (mp *MenuPopup) selectPrev() {
	n := len(mp.items)
	if n == 0 {
		return
	}
	start := mp.selected - 1
	if start < 0 {
		start = n - 1
	}
	for i := 0; i < n; i++ {
		idx := (start - i + n) % n
		if _, ok := mp.items[idx].(*MenuSeparator); !ok {
			mp.selected = idx
			return
		}
	}
}

func (mp *MenuPopup) fireSelected() {
	if mp.selected < 0 || mp.selected >= len(mp.items) {
		return
	}
	if mi, ok := mp.items[mp.selected].(*MenuItem); ok && !mi.Disabled {
		mp.result = mi.Command
	}
}

func (mp *MenuPopup) matchShortcut(r rune) {
	r = unicode.ToLower(r)
	for i, item := range mp.items {
		mi, ok := item.(*MenuItem)
		if !ok || mi.Disabled {
			continue
		}
		segments := ParseTildeLabel(mi.Label)
		for _, seg := range segments {
			if seg.Shortcut && len(seg.Text) > 0 {
				sc, _ := utf8.DecodeRuneInString(seg.Text)
				if unicode.ToLower(sc) == r {
					mp.selected = i
					mp.result = mi.Command
					return
				}
			}
		}
	}
}
```

**Note:** `MenuPopup.Draw` takes a `*theme.ColorScheme` parameter rather than looking it up from a View hierarchy, because the popup is not inserted into any Container — it's drawn directly by Application.Draw() at screen coordinates. The caller passes the Application's color scheme.

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add MenuPopup with rendering and keyboard/mouse navigation"`

---

### Task 3: MenuBar — Rendering and Activation

- [ ] Complete

**Files:**
- Create: `tv/menu_bar.go`

**Requirements:**
- `MenuBar` embeds `BaseView` and has fields: `menus []*SubMenu`, `active bool`, `activeIndex int`, `popup *MenuPopup`, `app *Application`.
- `NewMenuBar(menus ...*SubMenu) *MenuBar` creates a MenuBar. Sets `SfVisible`. Does not set `OfPreProcess` or `OfPostProcess` — the Application routes events to the MenuBar explicitly.
- `MenuBar.Draw(buf *DrawBuffer)` renders the menu bar:
  - Fill the entire row with `MenuNormal` style.
  - Draw each SubMenu's label horizontally with 2-space gaps, using tilde parsing for shortcuts. Shortcut letters use `MenuShortcut` style, rest use `MenuNormal`.
  - If `active` is true and `activeIndex` matches a menu, highlight that menu's label using `MenuSelected` style (both shortcut and non-shortcut text).
  - Each menu item starts at a computed X position. Store these positions for popup placement.
- `MenuBar.Popup() *MenuPopup` returns the currently open popup (nil if none). Used by Application.Draw() to render the popup.
- `MenuBar.IsActive() bool` returns whether the menu bar is currently in its modal loop.
- `MenuBar.Activate(app *Application)` is a convenience wrapper that calls `ActivateAt(app, 0, false)`.
- `MenuBar.ActivateAt(app *Application, index int, openPopup bool)` enters a modal event loop:
  1. Sets `active = true`, `activeIndex = index`. If `openPopup` is true, opens the popup for that index.
  2. Calls `app.drawAndFlush()` to show the highlighted menu bar.
  3. Loops: `app.PollEvent()`, handle the event, `app.drawAndFlush()`, check if still active.
  4. On exit: sets `active = false`, `popup = nil`.
- **Keyboard handling inside Activate's modal loop:**
  - **Escape**: if popup is open, close it (set `popup = nil`). If popup is already closed, deactivate the menu bar (exit loop).
  - **F10 or CmMenu command**: deactivate (exit loop).
  - **Left arrow**: move `activeIndex` to previous menu (wrapping). If popup is open, close it and open popup for the new menu.
  - **Right arrow**: move `activeIndex` to next menu (wrapping). Same popup behavior.
  - **Enter or Down arrow**: if no popup open, open popup for current menu. If popup is open, forward to popup.
  - **Up arrow**: if popup is open, forward to popup.
  - **Any other key**: if popup is open, forward to popup. If popup is closed, check if the key matches a menu label's shortcut (case-insensitive) — if so, open that menu.
  - After forwarding to popup: check `popup.Result()`. If non-zero and not CmCancel, post the command via `app.PostCommand(cmd, nil)` and deactivate. If CmCancel, close popup but stay active.
- **Mouse handling inside Activate's modal loop:**
  - Click on menu bar row (Y=0 in screen coords): find which menu item was clicked, set `activeIndex`, open popup.
  - Click inside popup bounds: translate coordinates to popup-local (subtract popup.Bounds().A), forward to popup. Check result.
  - Click outside both menu bar and popup: deactivate (exit loop).
- **Opening a popup:** `mp.openPopup()` creates a `NewMenuPopup(menus[activeIndex].Items, menuX, 1)` where `menuX` is the X position of the active menu label and Y=1 is the row below the menu bar (in screen coordinates). Assigns to `mp.popup`.
- The `menuXPositions` are computed during Draw (or in a helper) and stored for lookup.

**Implementation:**

```go
// tv/menu_bar.go
package tv

import (
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type MenuBar struct {
	BaseView
	menus       []*SubMenu
	active      bool
	activeIndex int
	popup       *MenuPopup
	app         *Application
	menuXPos    []int // computed X position of each menu label
}

func NewMenuBar(menus ...*SubMenu) *MenuBar {
	mb := &MenuBar{
		menus: menus,
	}
	mb.SetState(SfVisible, true)
	return mb
}

func (mb *MenuBar) Popup() *MenuPopup { return mb.popup }
func (mb *MenuBar) IsActive() bool    { return mb.active }
func (mb *MenuBar) Menus() []*SubMenu { return mb.menus }

func (mb *MenuBar) Draw(buf *DrawBuffer) {
	w := mb.Bounds().Width()
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault
	if cs := mb.ColorScheme(); cs != nil {
		normalStyle = cs.MenuNormal
		shortcutStyle = cs.MenuShortcut
		selectedStyle = cs.MenuSelected
	}

	buf.Fill(NewRect(0, 0, w, 1), ' ', normalStyle)

	mb.menuXPos = make([]int, len(mb.menus))
	x := 1
	for i, menu := range mb.menus {
		if i > 0 {
			x += 1
		}
		mb.menuXPos[i] = x

		isActive := mb.active && i == mb.activeIndex
		segments := ParseTildeLabel(menu.Label)
		for _, seg := range segments {
			style := normalStyle
			if isActive {
				style = selectedStyle
			} else if seg.Shortcut {
				style = shortcutStyle
			}
			buf.WriteStr(x, 0, seg.Text, style)
			x += utf8.RuneCountInString(seg.Text)
		}
	}
}

func (mb *MenuBar) menuEndX(i int) int {
	if i < 0 || i >= len(mb.menus) {
		return 0
	}
	return mb.menuXPos[i] + tildeTextLen(mb.menus[i].Label)
}

func (mb *MenuBar) menuIndexAtX(x int) int {
	for i := range mb.menus {
		start := mb.menuXPos[i]
		end := mb.menuEndX(i)
		if x >= start && x < end {
			return i
		}
	}
	return -1
}

func (mb *MenuBar) openPopup() {
	if mb.activeIndex < 0 || mb.activeIndex >= len(mb.menus) {
		return
	}
	menu := mb.menus[mb.activeIndex]
	x := mb.menuXPos[mb.activeIndex]
	mb.popup = NewMenuPopup(menu.Items, x, 1) // Y=1: row below menu bar
}

func (mb *MenuBar) closePopup() {
	mb.popup = nil
}

func (mb *MenuBar) ActivateAt(app *Application, index int, openPopup bool) {
	mb.app = app
	mb.active = true
	mb.activeIndex = index
	mb.popup = nil
	if openPopup {
		mb.openPopup()
	}

	app.drawAndFlush()

	for mb.active {
		event := app.PollEvent()
		if event == nil {
			break
		}

		mb.handleModalEvent(event, app)
		app.drawAndFlush()
	}

	mb.active = false
	mb.popup = nil
	mb.app = nil
}

func (mb *MenuBar) Activate(app *Application) {
	mb.ActivateAt(app, 0, false)
}

func (mb *MenuBar) handleModalEvent(event *Event, app *Application) {
	// Command events
	if event.What == EvCommand && event.Command == CmMenu {
		mb.active = false
		return
	}

	// Keyboard events
	if event.What == EvKeyboard && event.Key != nil {
		switch event.Key.Key {
		case tcell.KeyEscape:
			if mb.popup != nil {
				mb.closePopup()
			} else {
				mb.active = false
			}
			return

		case tcell.KeyF10:
			mb.active = false
			return

		case tcell.KeyLeft:
			mb.activeIndex = (mb.activeIndex - 1 + len(mb.menus)) % len(mb.menus)
			if mb.popup != nil {
				mb.openPopup()
			}
			return

		case tcell.KeyRight:
			mb.activeIndex = (mb.activeIndex + 1) % len(mb.menus)
			if mb.popup != nil {
				mb.openPopup()
			}
			return

		case tcell.KeyEnter, tcell.KeyDown:
			if mb.popup == nil {
				mb.openPopup()
			} else {
				mb.popup.HandleEvent(event)
				mb.checkPopupResult(app)
			}
			return

		case tcell.KeyUp:
			if mb.popup != nil {
				mb.popup.HandleEvent(event)
				mb.checkPopupResult(app)
			}
			return

		default:
			if mb.popup != nil {
				mb.popup.HandleEvent(event)
				mb.checkPopupResult(app)
			} else if event.Key.Key == tcell.KeyRune {
				mb.matchMenuShortcut(event.Key.Rune)
			}
			return
		}
	}

	// Mouse events
	if event.What == EvMouse && event.Mouse != nil {
		mx, my := event.Mouse.X, event.Mouse.Y

		// Click on menu bar row
		if my == 0 && event.Mouse.Button&tcell.Button1 != 0 {
			idx := mb.menuIndexAtX(mx)
			if idx >= 0 {
				mb.activeIndex = idx
				mb.openPopup()
			}
			return
		}

		// Inside popup bounds
		if mb.popup != nil {
			pb := mb.popup.Bounds()
			if pb.Contains(NewPoint(mx, my)) {
				localEvent := *event
				localMouse := *event.Mouse
				localMouse.X = mx - pb.A.X
				localMouse.Y = my - pb.A.Y
				localEvent.Mouse = &localMouse
				mb.popup.HandleEvent(&localEvent)
				mb.checkPopupResult(app)
				return
			}
		}

		// Click outside menu bar and popup — dismiss
		if event.Mouse.Button&tcell.Button1 != 0 {
			mb.active = false
		}
		return
	}
}

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
	// Selected a real command
	app.PostCommand(result, nil)
	mb.closePopup()
	mb.active = false
}

func (mb *MenuBar) matchMenuShortcut(r rune) {
	r = unicode.ToLower(r)
	for i, menu := range mb.menus {
		segments := ParseTildeLabel(menu.Label)
		for _, seg := range segments {
			if seg.Shortcut && len(seg.Text) > 0 {
				sc, _ := utf8.DecodeRuneInString(seg.Text)
				if unicode.ToLower(sc) == r {
					mb.activeIndex = i
					mb.openPopup()
					return
				}
			}
		}
	}
}

func (mb *MenuBar) HandleEvent(event *Event) {
	// Non-modal event handling: F10/CmMenu activates the menu bar.
	// Mouse click on the bar also activates.
	// The actual modal loop is in Activate(), called by Application.
	if event.What == EvCommand && event.Command == CmMenu {
		if mb.app != nil {
			mb.Activate(mb.app)
		}
		event.Clear()
		return
	}
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add MenuBar with rendering, activation, and modal event loop"`

---

### Task 4: Application Integration

- [ ] Complete

**Files:**
- Modify: `tv/application.go`

**Requirements:**
- `Application` struct adds a `menuBar *MenuBar` field.
- `WithMenuBar(mb *MenuBar) AppOption` registers a MenuBar with the Application.
- In `NewApplication`, after creating the desktop, if `menuBar` is set: set `menuBar.scheme = app.scheme` and `menuBar.app = app`.
- `Application.MenuBar() *MenuBar` returns the menu bar (nil if none).
- **layoutChildren** is updated:
  - If menuBar is present: menuBar at row 0 (bounds `(0, 0, w, 1)`), desktop at rows 1 through `h-statusH-1`, statusLine at last row.
  - If no menuBar: desktop starts at row 0 (unchanged from current behavior).
  - Desktop top is `menuH` where `menuH = 1` if menuBar is set, else `menuH = 0`.
- **Draw** is updated:
  - Draw order: desktop first, then menu bar, then popup (if open), then status line. This ensures the popup appears on top of the desktop.
  - Desktop is drawn into `SubBuffer(0, menuH, w, desktopH)`.
  - Menu bar is drawn into `SubBuffer(0, 0, w, 1)`.
  - If `menuBar.Popup()` is non-nil, draw the popup at its screen-coordinates bounds using `buf.SubBuffer(popup.Bounds())` and call `popup.Draw(popupBuf, app.scheme)`.
  - Status line is drawn into `SubBuffer(0, h-1, w, 1)`.
- **handleEvent** is updated:
  - After StatusLine transforms F10 into CmMenu, and before routing to Desktop: if event is `EvCommand` with `CmMenu` and `menuBar` is not nil, call `menuBar.Activate(app)`, clear the event, and return. This enters the modal menu loop.
- **routeMouseEvent** is updated:
  - Add menu bar hit-testing before status line and desktop: if mouse Y is within menuBar bounds and a menu label is clicked, call `menuBar.ActivateAt(app, clickedIndex, true)` to enter the modal loop with that menu open.
  - Clicks on the bar outside any menu label are ignored.

**Implementation changes to tv/application.go:**

```go
// Add to Application struct:
type Application struct {
	bounds     Rect
	screen     tcell.Screen
	screenOwn  bool
	desktop    *Desktop
	menuBar    *MenuBar
	statusLine *StatusLine
	scheme     *theme.ColorScheme
	quit       bool
	onCommand  func(CommandCode, any) bool
}

// Add WithMenuBar option:
func WithMenuBar(mb *MenuBar) AppOption {
	return func(app *Application) {
		app.menuBar = mb
	}
}

// Add accessor:
func (app *Application) MenuBar() *MenuBar { return app.menuBar }

// In NewApplication, after desktop setup:
if app.menuBar != nil {
	app.menuBar.scheme = app.scheme
	app.menuBar.app = app
}

// Updated layoutChildren:
func (app *Application) layoutChildren() {
	w, h := app.bounds.Width(), app.bounds.Height()
	menuH := 0

	if app.menuBar != nil {
		app.menuBar.SetBounds(NewRect(0, 0, w, 1))
		menuH = 1
	}

	desktopBottom := h
	if app.statusLine != nil {
		statusRow := h - 1
		if statusRow < menuH {
			statusRow = menuH
		}
		app.statusLine.SetBounds(NewRect(0, statusRow, w, 1))
		desktopBottom = statusRow
	}

	if app.desktop != nil {
		desktopTop := menuH
		desktopH := desktopBottom - desktopTop
		if desktopH < 0 {
			desktopH = 0
		}
		app.desktop.SetBounds(NewRect(0, 0, w, desktopH))
	}
}

// Updated Draw:
func (app *Application) Draw(buf *DrawBuffer) {
	h := app.bounds.Height()
	w := app.bounds.Width()
	menuH := 0
	if app.menuBar != nil {
		menuH = 1
	}

	desktopBottom := h
	if app.statusLine != nil {
		desktopBottom = h - 1
	}
	desktopH := desktopBottom - menuH

	// Draw desktop
	if app.desktop != nil && desktopH > 0 {
		desktopBuf := buf.SubBuffer(NewRect(0, menuH, w, desktopH))
		app.desktop.Draw(desktopBuf)
	}

	// Draw menu bar at row 0
	if app.menuBar != nil {
		menuBuf := buf.SubBuffer(NewRect(0, 0, w, 1))
		app.menuBar.Draw(menuBuf)

		// Draw popup on top of desktop if open
		if popup := app.menuBar.Popup(); popup != nil {
			pb := popup.Bounds()
			popupBuf := buf.SubBuffer(pb)
			popup.Draw(popupBuf, app.scheme)
		}
	}

	// Draw status line
	if app.statusLine != nil && h > 0 {
		statusBuf := buf.SubBuffer(NewRect(0, h-1, w, 1))
		app.statusLine.Draw(statusBuf)
	}
}

// Updated handleEvent — add CmMenu check after StatusLine:
func (app *Application) handleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		app.routeMouseEvent(event)
		if !event.IsCleared() && event.What == EvCommand {
			app.handleCommand(event)
		}
		return
	}

	if app.statusLine != nil {
		app.statusLine.HandleEvent(event)
	}

	// CmMenu activates the menu bar
	if !event.IsCleared() && event.What == EvCommand && event.Command == CmMenu {
		if app.menuBar != nil {
			app.menuBar.Activate(app)
			event.Clear()
			return
		}
	}

	if !event.IsCleared() && app.desktop != nil {
		app.desktop.HandleEvent(event)
	}

	if !event.IsCleared() {
		app.handleCommand(event)
	}
}

// Updated routeMouseEvent — add menu bar:
func (app *Application) routeMouseEvent(event *Event) {
	mx, my := event.Mouse.X, event.Mouse.Y

	// Menu bar
	if app.menuBar != nil {
		mbBounds := app.menuBar.Bounds()
		if mbBounds.Contains(NewPoint(mx, my)) {
			if event.Mouse.Button&tcell.Button1 != 0 {
				idx := app.menuBar.menuIndexAtX(mx)
				if idx >= 0 {
					app.menuBar.ActivateAt(app, idx, true)
				}
			}
			return
		}
	}

	// Status line
	if app.statusLine != nil {
		slBounds := app.statusLine.Bounds()
		if slBounds.Contains(NewPoint(mx, my)) {
			event.Mouse.X -= slBounds.A.X
			event.Mouse.Y -= slBounds.A.Y
			app.statusLine.HandleEvent(event)
			return
		}
	}

	// Desktop
	if app.desktop != nil {
		menuH := 0
		if app.menuBar != nil {
			menuH = 1
		}
		// Desktop is at screen row menuH..desktopBottom
		// Translate screen coordinates to desktop-local
		event.Mouse.Y -= menuH
		app.desktop.HandleEvent(event)
	}
}
```

**Important coordinate detail:** Desktop's bounds are `(0, 0, w, desktopH)` — it thinks it starts at (0,0). But on screen, the desktop is drawn at row `menuH`. So mouse events from the screen have Y coordinates in screen space, and we need to subtract `menuH` before forwarding to the desktop. This is done in `routeMouseEvent`.

Similarly, the popup's bounds are in screen coordinates (e.g., `(5, 1, 20, 8)`). When Application.Draw() calls `buf.SubBuffer(popup.Bounds())`, the SubBuffer is at the correct screen position.

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: integrate MenuBar into Application layout, drawing, and event routing"`

---

### Task 5: Integration Checkpoint — Menu System

- [ ] Complete

**Purpose:** Verify that the full menu flow works: Application → MenuBar activation → popup open → item selection → command posted.

**Requirements (for test writer):**
- Creating an Application with a MenuBar and StatusLine (F10→CmMenu), pressing F10 activates the menu bar (MenuBar.IsActive() is true during the modal loop). Pressing F10 again or Escape deactivates.
- With menu bar active, pressing Down or Enter opens a popup. The popup renders with single-line border characters (`┌┐└┘─│`).
- Pressing Down/Up inside the popup navigates between items (updates selected index). Separators are skipped.
- Pressing Enter on a menu item fires that item's command: the command appears as a posted event.
- Pressing Left/Right while popup is open switches to the adjacent menu and opens its popup.
- Pressing Escape while popup is open closes the popup but keeps the menu bar active. A second Escape deactivates the bar.
- Mouse click on a menu bar label opens that menu directly.
- Mouse click outside the menu bar and popup deactivates the menu.
- After menu deactivation, the desktop is fully functional (events route correctly).
- The menu bar renders at row 0, desktop starts at row 1, status line at the last row.
- Tab traversal inside windows still works after adding a menu bar (no regression).
- A dialog opened via ExecView while a menu bar exists works correctly (the modal dialog event loop functions independently of the menu system).

**Components to wire up:** Application (with SimulationScreen), MenuBar, SubMenu, MenuItem, MenuPopup, Desktop, Window, Button, StatusLine

**Run:** `go test ./tv/... -run TestIntegration -v -count=1`

**Commit:** `git commit -m "test: add Phase 4 menu system integration tests"`

---

### Task 6: Demo App + E2E Tests

- [ ] Complete

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**
- Demo app creates a MenuBar with two menus:
  - `"~F~ile"` containing: `NewMenuItem("~N~ew", CmUser+1, KbCtrl('N'))`, `NewMenuItem("~O~pen...", CmUser+2, KbCtrl('O'))`, `NewMenuSeparator()`, `NewMenuItem("E~x~it", CmQuit, KbAlt('X'))`.
  - `"~W~indow"` containing: `NewMenuItem("~T~ile", CmTile, KbNone())`, `NewMenuItem("~C~ascade", CmCascade, KbNone())`.
- Demo app passes the MenuBar via `WithMenuBar(menuBar)`.
- E2E test `TestMenuBarVisible`: after boot, the menu bar text "File" and "Window" is visible on the screen.
- E2E test `TestMenuOpenAndSelect`: press F10 to activate menu, press Enter to open File menu, verify "New" and "Open..." are visible (popup rendered), press Escape twice to dismiss. Verify app still runs (desktop pattern visible).
- E2E test `TestMenuSelectExit`: press F10, Enter (open File), press `x` (shortcut for E~x~it which fires CmQuit). Verify app exits.
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
	menuBar := tv.NewMenuBar(
		tv.NewSubMenu("~F~ile",
			tv.NewMenuItem("~N~ew", tv.CmUser+1, tv.KbCtrl('N')),
			tv.NewMenuItem("~O~pen...", tv.CmUser+2, tv.KbCtrl('O')),
			tv.NewMenuSeparator(),
			tv.NewMenuItem("E~x~it", tv.CmQuit, tv.KbAlt('X')),
		),
		tv.NewSubMenu("~W~indow",
			tv.NewMenuItem("~T~ile", tv.CmTile, tv.KbNone()),
			tv.NewMenuItem("~C~ascade", tv.CmCascade, tv.KbNone()),
		),
	)

	statusLine := tv.NewStatusLine(
		tv.NewStatusItem("~Alt+X~ Exit", tv.KbAlt('X'), tv.CmQuit),
		tv.NewStatusItem("~F2~ Dialog", tv.KbFunc(2), tv.CmUser),
		tv.NewStatusItem("~F10~ Menu", tv.KbFunc(10), tv.CmMenu),
	)

	var app *tv.Application
	var err error

	app, err = tv.NewApplication(
		tv.WithMenuBar(menuBar),
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

**E2E test requirements (for test writer):**
- `TestMenuBarVisible`: After boot, the first screen row contains "File" and "Window" text. App exits cleanly on Alt+X.
- `TestMenuOpenAndSelect`: Press F10 to activate menu, Enter to open File menu. Screen shows "New" and "Open..." text (popup visible). Escape twice dismisses. Desktop pattern (`░`) is visible afterward. App still runs.
- `TestMenuSelectExit`: Press F10, Enter (open File), then `x` (shortcut for E~x~it which fires CmQuit). App exits.
- All existing E2E tests continue to pass (regression).

**Run tests:** `go test ./e2e/... -v -count=1 -timeout 30s`

**Commit:** `git commit -m "feat: add menu bar to demo app and e2e tests for Phase 4"`
