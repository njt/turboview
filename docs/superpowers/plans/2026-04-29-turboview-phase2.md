# TurboView Phase 2 Implementation Plan — Windows

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add overlapping windows with frames, drag/resize/zoom/close interaction, shadow rendering, z-ordering (click to front), Alt+N window switching, and tile/cascade arrangement. After this phase, the demo app shows 2-3 windows on the desktop that the user can interact with.

**Architecture:** Window wraps an internal Group (facade pattern) with a frame, title, and interaction logic. Desktop is upgraded from a simple pattern-drawer to a Container (also via internal Group + facade), managing windows with z-ordering. Mouse events are routed by position (hit-testing) through Application → Desktop → Window. Group.HandleEvent gains three-phase dispatch (preprocess/focused/postprocess) for keyboard/command events.

**Tech Stack:** Go 1.22+, tcell/v2 for terminal abstraction, tmux for e2e tests.

---

## File Structure

```
tv/
  desktop.go         — MODIFY: upgrade to Container via internal Group, shadow rendering, window management
  group.go           — MODIFY: add BringToFront, three-phase dispatch
  window.go          — CREATE: Window with frame, drag, resize, zoom, close
  application.go     — MODIFY: add mouse event routing by position
  draw_buffer.go     — MODIFY: add SetCellStyle for shadow rendering
e2e/
  testapp/basic/main.go  — MODIFY: add windows to demo app
  e2e_test.go            — MODIFY: add window tests
```

---

### Task 1: Desktop Container Upgrade

- [ ] Complete

**Files:**
- Modify: `tv/desktop.go`
- Modify: `tv/group.go`

**Requirements:**
- `Desktop` satisfies the `Container` interface (compile-time check: `var _ Container = (*Desktop)(nil)`)
- Desktop holds an internal Group whose facade is set to the Desktop itself, so children inserted via `Desktop.Insert(v)` see the Desktop as their owner
- `Desktop.Insert(v)` delegates to the internal Group
- `Desktop.Remove(v)` delegates to the internal Group
- `Desktop.Children()` delegates to the internal Group
- `Desktop.FocusedChild()` delegates to the internal Group
- `Desktop.SetFocusedChild(v)` delegates to the internal Group
- `Desktop.ExecView(v)` delegates to the internal Group (panics "not implemented")
- `Desktop.SetBounds(r)` updates both Desktop's bounds and the internal Group's bounds (Group uses origin 0,0 with Desktop's width and height)
- `Desktop.Draw(buf)` fills the background with the pattern rune and DesktopBackground style, then calls the internal Group's Draw to render children back-to-front
- `Desktop.HandleEvent(event)` delegates to the internal Group
- All existing Desktop tests pass without modification (no children = same behavior as before)
- `Group.BringToFront(v)` moves the given child to the end of the children list (frontmost in z-order) and selects it; no-op if the child is not in the group
- `Desktop.BringToFront(v)` delegates to the internal Group's BringToFront

**Implementation:**

```go
// tv/group.go — add BringToFront method

func (g *Group) BringToFront(v View) {
	for i, child := range g.children {
		if child == v {
			g.children = append(g.children[:i], g.children[i+1:]...)
			g.children = append(g.children, v)
			g.selectChild(v)
			return
		}
	}
}
```

```go
// tv/desktop.go — full rewrite

package tv

import "github.com/gdamore/tcell/v2"

var _ Container = (*Desktop)(nil)

type Desktop struct {
	BaseView
	group   *Group
	pattern rune
}

func NewDesktop(bounds Rect) *Desktop {
	d := &Desktop{
		pattern: '░',
	}
	d.SetBounds(bounds)
	d.SetState(SfVisible, true)
	d.SetGrowMode(GfGrowAll)
	d.group = NewGroup(NewRect(0, 0, bounds.Width(), bounds.Height()))
	d.group.SetFacade(d)
	return d
}

func (d *Desktop) SetBounds(r Rect) {
	d.BaseView.SetBounds(r)
	if d.group != nil {
		d.group.SetBounds(NewRect(0, 0, r.Width(), r.Height()))
	}
}

func (d *Desktop) Insert(v View)                  { d.group.Insert(v) }
func (d *Desktop) Remove(v View)                  { d.group.Remove(v) }
func (d *Desktop) Children() []View               { return d.group.Children() }
func (d *Desktop) FocusedChild() View              { return d.group.FocusedChild() }
func (d *Desktop) SetFocusedChild(v View)          { d.group.SetFocusedChild(v) }
func (d *Desktop) ExecView(v View) CommandCode     { return d.group.ExecView(v) }
func (d *Desktop) BringToFront(v View)             { d.group.BringToFront(v) }

func (d *Desktop) Draw(buf *DrawBuffer) {
	w, h := d.Bounds().Width(), d.Bounds().Height()
	style := tcell.StyleDefault
	if cs := d.ColorScheme(); cs != nil {
		style = cs.DesktopBackground
	}
	buf.Fill(NewRect(0, 0, w, h), d.pattern, style)
	d.group.Draw(buf)
}

func (d *Desktop) HandleEvent(event *Event) {
	d.group.HandleEvent(event)
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: upgrade Desktop to Container with internal Group and BringToFront"`

---

### Task 2: Window Frame Drawing

- [ ] Complete

**Files:**
- Create: `tv/window.go`

**Requirements:**
- `Window` satisfies the `Container` interface (compile-time check: `var _ Container = (*Window)(nil)`)
- `NewWindow(bounds, title, opts...)` creates a Window with `SfVisible`, `OfSelectable|OfTopSelect` options, and an internal Group whose facade is set to the Window
- `WindowOption` type for functional options: `WithWindowNumber(n int)`
- `Window.Title()` returns the title string
- `Window.SetTitle(t)` updates the title
- `Window.Number()` returns the window number (0 if not set)
- Window delegates Container methods (`Insert`, `Remove`, `Children`, `FocusedChild`, `SetFocusedChild`, `ExecView`) to its internal Group
- `Window.SetBounds(r)` updates both Window's bounds and the internal Group's bounds (Group uses origin 0,0, width-2, height-2 for the client area)
- `Window.Draw(buf)` renders:
  - Client area background filled with `ColorScheme().WindowBackground` style and space rune, at (1,1) with size (width-2, height-2)
  - Active frame (when `SfSelected` is set): double-line border using `╔═╗║╚═╝` in `WindowFrameActive` style
  - Inactive frame (when `SfSelected` is not set): single-line border using `┌─┐│└─┘` in `WindowFrameInactive` style
  - Close icon `[×]` at positions (1,0)-(3,0) of top border in frame style
  - Zoom icon `[↑]` at positions (width-4,0)-(width-2,0) in frame style
  - Title centered between close and zoom icons, wrapped in spaces, in `WindowTitle` style
  - Children drawn into a client-area SubBuffer at (1,1) with size (width-2, height-2) via Group.Draw
- If window is too small (width < 8 or height < 3), Draw is a no-op
- When no ColorScheme is available, Draw uses `tcell.StyleDefault` for all styles
- `Window.HandleEvent(event)` delegates to the internal Group

**Implementation:**

```go
// tv/window.go
package tv

import "github.com/gdamore/tcell/v2"

var _ Container = (*Window)(nil)

type Window struct {
	BaseView
	group      *Group
	title      string
	number     int
	zoomed     bool
	zoomBounds Rect
	dragOff    Point
	resizing   bool
	resizeLeft bool
}

type WindowOption func(*Window)

func WithWindowNumber(n int) WindowOption {
	return func(w *Window) { w.number = n }
}

func NewWindow(bounds Rect, title string, opts ...WindowOption) *Window {
	w := &Window{
		title: title,
	}
	w.SetBounds(bounds)
	w.SetState(SfVisible, true)
	w.SetOptions(OfSelectable|OfTopSelect, true)

	cw := max(bounds.Width()-2, 0)
	ch := max(bounds.Height()-2, 0)
	w.group = NewGroup(NewRect(0, 0, cw, ch))
	w.group.SetFacade(w)

	for _, opt := range opts {
		opt(w)
	}
	return w
}

func (w *Window) Title() string      { return w.title }
func (w *Window) SetTitle(t string)  { w.title = t }
func (w *Window) Number() int        { return w.number }

func (w *Window) Insert(v View)                  { w.group.Insert(v) }
func (w *Window) Remove(v View)                  { w.group.Remove(v) }
func (w *Window) Children() []View               { return w.group.Children() }
func (w *Window) FocusedChild() View              { return w.group.FocusedChild() }
func (w *Window) SetFocusedChild(v View)          { w.group.SetFocusedChild(v) }
func (w *Window) ExecView(v View) CommandCode     { return w.group.ExecView(v) }

func (w *Window) SetBounds(r Rect) {
	w.BaseView.SetBounds(r)
	if w.group != nil {
		cw := max(r.Width()-2, 0)
		ch := max(r.Height()-2, 0)
		w.group.SetBounds(NewRect(0, 0, cw, ch))
	}
}

func (w *Window) Draw(buf *DrawBuffer) {
	width, height := w.Bounds().Width(), w.Bounds().Height()
	if width < 8 || height < 3 {
		return
	}

	cs := w.ColorScheme()
	active := w.HasState(SfSelected)

	frameStyle := tcell.StyleDefault
	titleStyle := tcell.StyleDefault
	bgStyle := tcell.StyleDefault
	if cs != nil {
		if active {
			frameStyle = cs.WindowFrameActive
		} else {
			frameStyle = cs.WindowFrameInactive
		}
		titleStyle = cs.WindowTitle
		bgStyle = cs.WindowBackground
	}

	// Client area background
	buf.Fill(NewRect(1, 1, width-2, height-2), ' ', bgStyle)

	// Frame
	if active {
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
	} else {
		buf.WriteChar(0, 0, '┌', frameStyle)
		buf.WriteChar(width-1, 0, '┐', frameStyle)
		buf.WriteChar(0, height-1, '└', frameStyle)
		buf.WriteChar(width-1, height-1, '┘', frameStyle)
		for x := 1; x < width-1; x++ {
			buf.WriteChar(x, 0, '─', frameStyle)
			buf.WriteChar(x, height-1, '─', frameStyle)
		}
		for y := 1; y < height-1; y++ {
			buf.WriteChar(0, y, '│', frameStyle)
			buf.WriteChar(width-1, y, '│', frameStyle)
		}
	}

	// Close icon [×] at (1,0)-(3,0)
	buf.WriteChar(1, 0, '[', frameStyle)
	buf.WriteChar(2, 0, '×', frameStyle)
	buf.WriteChar(3, 0, ']', frameStyle)

	// Zoom icon [↑] at (width-4, 0)-(width-2, 0)
	buf.WriteChar(width-4, 0, '[', frameStyle)
	buf.WriteChar(width-3, 0, '↑', frameStyle)
	buf.WriteChar(width-2, 0, ']', frameStyle)

	// Title centered between icons
	availStart := 4
	availEnd := width - 4
	availW := availEnd - availStart
	if availW > 0 && len(w.title) > 0 {
		t := w.title
		if len(t) > availW-2 {
			t = t[:availW-2]
		}
		padded := " " + t + " "
		titleX := availStart + (availW-len(padded))/2
		if titleX < availStart {
			titleX = availStart
		}
		buf.WriteStr(titleX, 0, padded, titleStyle)
	}

	// Draw children in client area
	clientBuf := buf.SubBuffer(NewRect(1, 1, width-2, height-2))
	w.group.Draw(clientBuf)
}

func (w *Window) HandleEvent(event *Event) {
	w.group.HandleEvent(event)
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add Window with frame drawing, title, close and zoom icons"`

---

### Task 3: Three-Phase Event Dispatch in Group

- [ ] Complete

**Files:**
- Modify: `tv/group.go`

**Requirements:**
- `Group.HandleEvent` implements three-phase dispatch for non-mouse, non-cleared events:
  1. **Preprocess**: iterate all children — for each child with `OfPreProcess` that is NOT the focused child, call `child.HandleEvent(event)`. Stop if event is cleared.
  2. **Focused**: if not cleared, forward to the focused child.
  3. **Postprocess**: iterate all children — for each child with `OfPostProcess` that is NOT the focused child, call `child.HandleEvent(event)`. Stop if event is cleared.
- Mouse events (`EvMouse`) skip three-phase dispatch and forward directly to the focused child (position-based routing is handled by the caller)
- Cleared events (EvNothing) are not forwarded at all
- When no children have `OfPreProcess` or `OfPostProcess`, behavior is identical to the existing single-phase dispatch (existing tests pass without modification)
- Broadcast events (`EvBroadcast`) are delivered to ALL children, not just focused

**Implementation:**

```go
// tv/group.go — replace HandleEvent

func (g *Group) HandleEvent(event *Event) {
	if event.IsCleared() {
		return
	}

	// Mouse events: forward to focused child (positional routing done by caller)
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

	// Three-phase dispatch for keyboard and command events

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
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add three-phase event dispatch (preprocess/focused/postprocess) to Group"`

---

### Task 4: Mouse Event Routing

- [ ] Complete

**Files:**
- Modify: `tv/application.go`
- Modify: `tv/desktop.go`
- Modify: `tv/window.go`

**Requirements:**
- `Application.handleEvent` routes mouse events by position: checks StatusLine bounds first, then Desktop bounds, translating coordinates to the target's local space. Non-mouse events use the existing path (StatusLine → Desktop → CmQuit check).
- `Desktop.HandleEvent` for mouse events: iterates windows front-to-back (last child first), hit-tests using `child.Bounds().Contains(point)`, translates coordinates to window-local space, and forwards. If the hit window has `OfTopSelect` and a button is pressed, brings it to front first.
- `Window.HandleEvent` for mouse events: if the click is in the client area (x > 0 && x < width-1 && y > 0 && y < height-1), translates to client-local coordinates and forwards to the internal Group. Otherwise the click is on the frame (handled in Task 5).
- All existing keyboard-event tests pass unchanged — mouse routing only activates for `EvMouse` events

**Implementation:**

```go
// tv/application.go — modify handleEvent, add routeMouseEvent

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

	if !event.IsCleared() && app.desktop != nil {
		app.desktop.HandleEvent(event)
	}

	if !event.IsCleared() {
		app.handleCommand(event)
	}
}

func (app *Application) handleCommand(event *Event) {
	if event.What == EvCommand {
		switch event.Command {
		case CmQuit:
			app.quit = true
			event.Clear()
		}
	}
}

func (app *Application) routeMouseEvent(event *Event) {
	mx, my := event.Mouse.X, event.Mouse.Y

	if app.statusLine != nil {
		slBounds := app.statusLine.Bounds()
		if slBounds.Contains(NewPoint(mx, my)) {
			event.Mouse.X -= slBounds.A.X
			event.Mouse.Y -= slBounds.A.Y
			app.statusLine.HandleEvent(event)
			return
		}
	}

	if app.desktop != nil {
		dBounds := app.desktop.Bounds()
		if dBounds.Contains(NewPoint(mx, my)) {
			event.Mouse.X -= dBounds.A.X
			event.Mouse.Y -= dBounds.A.Y
			app.desktop.HandleEvent(event)
		}
	}
}
```

```go
// tv/desktop.go — modify HandleEvent, add routeMouseEvent

func (d *Desktop) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		d.routeMouseEvent(event)
		return
	}
	d.group.HandleEvent(event)
}

func (d *Desktop) routeMouseEvent(event *Event) {
	mx, my := event.Mouse.X, event.Mouse.Y

	children := d.group.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		if !child.HasState(SfVisible) {
			continue
		}
		bounds := child.Bounds()
		if bounds.Contains(NewPoint(mx, my)) {
			if child.HasOption(OfTopSelect) && event.Mouse.Button&tcell.Button1 != 0 {
				d.BringToFront(child)
			}
			event.Mouse.X -= bounds.A.X
			event.Mouse.Y -= bounds.A.Y
			child.HandleEvent(event)
			return
		}
	}
}
```

```go
// tv/window.go — modify HandleEvent

func (w *Window) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		w.handleMouseEvent(event)
		return
	}
	w.group.HandleEvent(event)
}

func (w *Window) handleMouseEvent(event *Event) {
	mx, my := event.Mouse.X, event.Mouse.Y
	width, height := w.Bounds().Width(), w.Bounds().Height()

	// Client area: forward to group with translated coordinates
	if mx > 0 && mx < width-1 && my > 0 && my < height-1 {
		event.Mouse.X -= 1
		event.Mouse.Y -= 1
		w.group.HandleEvent(event)
		return
	}

	// Frame clicks handled in Task 5
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add position-based mouse event routing with hit-testing"`

---

### Task 5: Window Interaction (Drag, Resize, Zoom, Close)

- [ ] Complete

**Files:**
- Modify: `tv/window.go`
- Modify: `tv/desktop.go`

**Requirements:**
- **Drag**: left-click (`Button1`) on the title bar (y=0, x not on close icon 1-3 or zoom icon width-4 to width-2) sets `SfDragging` and records the drag offset (window-local mouse position). Subsequent mouse events while `Button1` is held update the window's position. Releasing `Button1` clears `SfDragging`.
- **Mouse capture during drag**: Desktop routes ALL mouse events to a dragging window (the one with `SfDragging` set) regardless of hit-testing, and does NOT translate coordinates to window-local (passes Desktop-local coordinates so the window can compute its new origin)
- **Resize**: left-click on the bottom-right corner (x == width-1 && y == height-1) OR bottom-left corner (x == 0 && y == height-1) starts resizing. During resize, mouse movement adjusts the window's size. Bottom-right resize adjusts width and height. Bottom-left resize adjusts the left edge and height (keeping right edge fixed). Minimum window size: width 10, height 5. Releasing `Button1` ends resize.
- **Mouse capture during resize**: same as drag — Desktop routes to the resizing window without coordinate translation
- **Close**: left-click on the close icon `[×]` at (1-3, 0) transforms the event to `EvCommand/CmClose`
- **Zoom**: left-click on the zoom icon at (width-4 to width-2, 0), OR double-click on the title bar (y=0, ClickCount >= 2, not on close/zoom icons), toggles between normal and zoomed state. When zooming in, stores current bounds and sets bounds to fill the owner's area. When zooming out, restores stored bounds. Zoom icon changes to `↕` when zoomed.
- `Window.Zoom()` toggles zoom state (public method for programmatic use)
- `Window.IsZoomed()` returns the current zoom state
- Desktop handles `CmClose`: removes the focused child and clears the event
- Desktop handles mouse capture: checks for any child with `SfDragging` or `resizing` before normal hit-testing

**Implementation:**

```go
// tv/window.go — extend handleMouseEvent, add Zoom/IsZoomed

func (w *Window) IsZoomed() bool { return w.zoomed }

func (w *Window) Zoom() {
	if w.zoomed {
		w.SetBounds(w.zoomBounds)
		w.zoomed = false
	} else {
		w.zoomBounds = w.Bounds()
		if owner := w.Owner(); owner != nil {
			ob := owner.Bounds()
			w.SetBounds(NewRect(0, 0, ob.Width(), ob.Height()))
		}
		w.zoomed = true
	}
}

func (w *Window) handleMouseEvent(event *Event) {
	mx, my := event.Mouse.X, event.Mouse.Y
	width, height := w.Bounds().Width(), w.Bounds().Height()

	// During drag: coordinates are in Desktop-local space
	if w.HasState(SfDragging) {
		if event.Mouse.Button&tcell.Button1 != 0 {
			bounds := w.Bounds()
			newX := mx - w.dragOff.X
			newY := my - w.dragOff.Y
			w.SetBounds(NewRect(newX, newY, bounds.Width(), bounds.Height()))
		} else {
			w.SetState(SfDragging, false)
		}
		event.Clear()
		return
	}

	// During resize: coordinates are in Desktop-local space
	if w.resizing {
		if event.Mouse.Button&tcell.Button1 != 0 {
			bounds := w.Bounds()
			if w.resizeLeft {
				// Bottom-left resize: adjust left edge and height, keep right edge fixed
				newX := mx
				newW := bounds.B.X - newX
				newH := my - bounds.A.Y + 1
				if newW < 10 {
					newW = 10
					newX = bounds.B.X - newW
				}
				if newH < 5 {
					newH = 5
				}
				w.SetBounds(NewRect(newX, bounds.A.Y, newW, newH))
			} else {
				// Bottom-right resize: adjust width and height
				newW := mx - bounds.A.X + 1
				newH := my - bounds.A.Y + 1
				if newW < 10 {
					newW = 10
				}
				if newH < 5 {
					newH = 5
				}
				w.SetBounds(NewRect(bounds.A.X, bounds.A.Y, newW, newH))
			}
		} else {
			w.resizing = false
		}
		event.Clear()
		return
	}

	// Frame clicks (window-local coordinates)
	if event.Mouse.Button&tcell.Button1 != 0 {
		// Top border
		if my == 0 {
			// Close icon [×] at (1-3, 0)
			if mx >= 1 && mx <= 3 {
				event.What = EvCommand
				event.Command = CmClose
				event.Mouse = nil
				return
			}
			// Zoom icon at (width-4 to width-2, 0)
			if mx >= width-4 && mx <= width-2 {
				w.Zoom()
				event.Clear()
				return
			}
			// Double-click on title bar: zoom
			if event.Mouse.ClickCount >= 2 {
				w.Zoom()
				event.Clear()
				return
			}
			// Title bar: start drag
			w.SetState(SfDragging, true)
			w.dragOff = NewPoint(mx, my)
			event.Clear()
			return
		}

		// Bottom-right corner: start resize
		if mx == width-1 && my == height-1 {
			w.resizing = true
			w.resizeLeft = false
			event.Clear()
			return
		}

		// Bottom-left corner: start resize (left edge)
		if mx == 0 && my == height-1 {
			w.resizing = true
			w.resizeLeft = true
			event.Clear()
			return
		}
	}

	// Client area: forward to group
	if mx > 0 && mx < width-1 && my > 0 && my < height-1 {
		event.Mouse.X -= 1
		event.Mouse.Y -= 1
		w.group.HandleEvent(event)
		return
	}
}
```

```go
// tv/window.go — update Draw to show correct zoom icon

// In Draw method, replace the zoom icon section:
if w.zoomed {
	buf.WriteChar(width-3, 0, '↕', frameStyle)
} else {
	buf.WriteChar(width-3, 0, '↑', frameStyle)
}
```

```go
// tv/desktop.go — modify routeMouseEvent for mouse capture, add CmClose handling

func (d *Desktop) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		d.routeMouseEvent(event)
		return
	}

	// Handle desktop-level commands before delegating to group
	if event.What == EvCommand {
		switch event.Command {
		case CmClose:
			if focused := d.group.FocusedChild(); focused != nil {
				d.Remove(focused)
				event.Clear()
				return
			}
		}
	}

	d.group.HandleEvent(event)
}

func (d *Desktop) routeMouseEvent(event *Event) {
	mx, my := event.Mouse.X, event.Mouse.Y

	// Mouse capture: if any child is being dragged or resized,
	// route all mouse events to it WITHOUT translating coordinates
	for _, child := range d.group.Children() {
		if child.HasState(SfDragging) {
			child.HandleEvent(event)
			return
		}
		if w, ok := child.(*Window); ok && w.resizing {
			child.HandleEvent(event)
			return
		}
	}

	// Normal hit-testing: front-to-back
	children := d.group.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		if !child.HasState(SfVisible) {
			continue
		}
		bounds := child.Bounds()
		if bounds.Contains(NewPoint(mx, my)) {
			if child.HasOption(OfTopSelect) && event.Mouse.Button&tcell.Button1 != 0 {
				d.BringToFront(child)
			}
			event.Mouse.X -= bounds.A.X
			event.Mouse.Y -= bounds.A.Y
			child.HandleEvent(event)
			return
		}
	}
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add window drag, resize, zoom, and close interaction"`

---

### Task 6: Shadow Rendering

- [ ] Complete

**Note:** This task replaces Task 1's `d.group.Draw(buf)` call in `Desktop.Draw` with a manual child iteration loop that interleaves shadow rendering after each window. The Group's draw method is no longer used for Desktop's children — Desktop now owns the draw loop.

**Files:**
- Modify: `tv/draw_buffer.go`
- Modify: `tv/desktop.go`

**Requirements:**
- `DrawBuffer.SetCellStyle(x, y, style)` updates only the style of the cell at (x, y), preserving its rune and combining characters. Out-of-bounds or outside-clip writes are no-ops.
- Desktop draws shadows during its draw pass, after drawing each visible window
- Shadow extends 2 cells to the right and 1 cell below the window
- Right shadow: column range [window.B.X, window.B.X+2), row range [window.A.Y+1, window.B.Y+1)
- Bottom shadow: column range [window.A.X+2, window.B.X+2), row = window.B.Y
- Shadow uses `ColorScheme().WindowShadow` style
- Shadow preserves the existing character (rune) in each cell and only replaces the style
- Shadow does not render outside the Desktop's bounds

**Implementation:**

```go
// tv/draw_buffer.go — add SetCellStyle

func (db *DrawBuffer) SetCellStyle(x, y int, style tcell.Style) {
	ax, ay := x+db.offset.X, y+db.offset.Y
	if !db.clip.Contains(NewPoint(ax, ay)) {
		return
	}
	if ay < 0 || ay >= len(db.cells) || ax < 0 || ax >= len(db.cells[0]) {
		return
	}
	db.cells[ay][ax].Style = style
}
```

```go
// tv/desktop.go — modify Draw to render shadows

func (d *Desktop) Draw(buf *DrawBuffer) {
	w, h := d.Bounds().Width(), d.Bounds().Height()
	style := tcell.StyleDefault
	shadowStyle := tcell.StyleDefault
	if cs := d.ColorScheme(); cs != nil {
		style = cs.DesktopBackground
		shadowStyle = cs.WindowShadow
	}

	buf.Fill(NewRect(0, 0, w, h), d.pattern, style)

	for _, child := range d.group.Children() {
		if !child.HasState(SfVisible) {
			continue
		}
		cb := child.Bounds()

		// Draw the window
		sub := buf.SubBuffer(cb)
		child.Draw(sub)

		// Draw shadow (2 right, 1 down)
		// Right shadow
		for y := cb.A.Y + 1; y < cb.B.Y+1; y++ {
			for x := cb.B.X; x < cb.B.X+2; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					buf.SetCellStyle(x, y, shadowStyle)
				}
			}
		}
		// Bottom shadow
		for x := cb.A.X + 2; x < cb.B.X+2; x++ {
			y := cb.B.Y
			if x >= 0 && x < w && y >= 0 && y < h {
				buf.SetCellStyle(x, y, shadowStyle)
			}
		}
	}
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add window shadow rendering to Desktop draw pass"`

---

### Task 7: Desktop Window Management

- [ ] Complete

**Note:** This task's `Desktop.HandleEvent` is the final version — it supersedes the versions from Tasks 1, 4, and 5 by incorporating mouse routing, command handling, and keyboard shortcuts into a single coherent method. Implementers should write the complete HandleEvent, not append to a previous version.

**Files:**
- Modify: `tv/desktop.go`

**Requirements:**
- Desktop handles Alt+1 through Alt+9 keyboard events: selects the window whose `Number()` matches the digit. Clears the event on match.
- `Desktop.SelectNextWindow()` moves focus to the next window in the children list (wraps around). Brings the newly focused window to front. Note: `BringToFront` moves the window to the end of the children list, so subsequent `SelectNextWindow` calls cycle in z-order, not insertion order. This is intentional — the cycling tracks the most-recently-used window stack.
- `Desktop.SelectPrevWindow()` moves focus to the previous window (wraps around). Brings it to front. Same z-order cycling behavior as SelectNextWindow.
- Desktop handles `CmNext` command by calling `SelectNextWindow()` and clearing the event
- Desktop handles `CmPrev` command by calling `SelectPrevWindow()` and clearing the event
- `Desktop.Tile()` arranges all visible windows in a non-overlapping grid:
  - Computes a grid of `cols × rows` cells where `cols = ceil(sqrt(n))` and `rows = ceil(n / cols)`
  - Each window gets one cell; the last column/row absorbs remaining space
  - Minimum window size per cell: 10 wide, 5 tall
- `Desktop.Cascade()` arranges visible windows in a diagonal stack:
  - Each window is 3/4 of Desktop width and 3/4 of Desktop height (minimum 10×5)
  - Windows offset by (2, 1) from the previous, wrapping if they'd exceed Desktop bounds
- Desktop handles `CmTile` and `CmCascade` commands
- `CmNext` and `CmPrev` handlers are available for programmatic use and will be wired to keyboard bindings (F6/Shift+F6) via the menu system in Phase 4. No keyboard binding is added in Phase 2.
- All window management keyboard events are handled BEFORE delegating to Group

**Implementation:**

```go
// tv/desktop.go — extend HandleEvent, add management methods

func (d *Desktop) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		d.routeMouseEvent(event)
		return
	}

	// Alt+N window switching
	if event.What == EvKeyboard && event.Key != nil {
		if event.Key.Modifiers&tcell.ModAlt != 0 && event.Key.Key == tcell.KeyRune {
			n := int(event.Key.Rune - '0')
			if n >= 1 && n <= 9 {
				d.selectWindowByNumber(n)
				event.Clear()
				return
			}
		}
	}

	// Desktop-level commands
	if event.What == EvCommand {
		switch event.Command {
		case CmClose:
			if focused := d.group.FocusedChild(); focused != nil {
				d.Remove(focused)
				event.Clear()
				return
			}
		case CmNext:
			d.SelectNextWindow()
			event.Clear()
			return
		case CmPrev:
			d.SelectPrevWindow()
			event.Clear()
			return
		case CmTile:
			d.Tile()
			event.Clear()
			return
		case CmCascade:
			d.Cascade()
			event.Clear()
			return
		}
	}

	d.group.HandleEvent(event)
}

func (d *Desktop) selectWindowByNumber(n int) {
	for _, child := range d.group.Children() {
		if w, ok := child.(*Window); ok && w.Number() == n {
			d.BringToFront(w)
			return
		}
	}
}

func (d *Desktop) SelectNextWindow() {
	children := d.group.Children()
	if len(children) == 0 {
		return
	}
	current := d.group.FocusedChild()
	if current == nil {
		d.BringToFront(children[0])
		return
	}
	for i, child := range children {
		if child == current {
			next := children[(i+1)%len(children)]
			d.BringToFront(next)
			return
		}
	}
}

func (d *Desktop) SelectPrevWindow() {
	children := d.group.Children()
	if len(children) == 0 {
		return
	}
	current := d.group.FocusedChild()
	if current == nil {
		d.BringToFront(children[len(children)-1])
		return
	}
	for i, child := range children {
		if child == current {
			prev := children[(i-1+len(children))%len(children)]
			d.BringToFront(prev)
			return
		}
	}
}

func (d *Desktop) visibleWindows() []*Window {
	var windows []*Window
	for _, child := range d.group.Children() {
		if w, ok := child.(*Window); ok && w.HasState(SfVisible) {
			windows = append(windows, w)
		}
	}
	return windows
}

func (d *Desktop) Tile() {
	windows := d.visibleWindows()
	n := len(windows)
	if n == 0 {
		return
	}
	dw, dh := d.Bounds().Width(), d.Bounds().Height()

	cols := 1
	for cols*cols < n {
		cols++
	}
	rows := (n + cols - 1) / cols

	cellW := dw / cols
	cellH := dh / rows

	for i, win := range windows {
		col := i % cols
		row := i / cols
		x := col * cellW
		y := row * cellH
		w := cellW
		h := cellH
		if col == cols-1 {
			w = dw - x
		}
		if row == rows-1 || i == n-1 {
			h = dh - y
		}
		if w < 10 {
			w = 10
		}
		if h < 5 {
			h = 5
		}
		win.SetBounds(NewRect(x, y, w, h))
		win.zoomed = false
	}
}

func (d *Desktop) Cascade() {
	windows := d.visibleWindows()
	n := len(windows)
	if n == 0 {
		return
	}
	dw, dh := d.Bounds().Width(), d.Bounds().Height()
	winW := dw * 3 / 4
	winH := dh * 3 / 4
	if winW < 10 {
		winW = 10
	}
	if winH < 5 {
		winH = 5
	}

	for i, win := range windows {
		x := i * 2
		y := i
		if x+winW > dw {
			x = 0
		}
		if y+winH > dh {
			y = 0
		}
		win.SetBounds(NewRect(x, y, winW, winH))
		win.zoomed = false
	}
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add Desktop window management with Alt+N switching, Tile, and Cascade"`

---

### Task 8: Integration Checkpoint — Window Management System

- [ ] Complete

**Purpose:** Verify that Application, Desktop, Window, three-phase dispatch, mouse routing, drag, resize, zoom, close, shadow, and window management work together end-to-end.

**Requirements (for test writer):**
- Creating an Application with a Desktop, inserting two Windows into the Desktop, and calling `app.Draw(buf)` renders: desktop pattern as background, both windows with frames visible, frontmost window drawn on top (overwrites background and back window where they overlap)
- Window frame uses double-line characters (`╔═╗║╚═╝`) for the selected window and single-line characters (`┌─┐│└─┘`) for the non-selected window
- Window title appears centered in the top border between the close and zoom icons
- Close icon `[×]` appears at positions (1-3, 0) of the window's frame
- Zoom icon appears at positions (width-4, width-2) of the window's frame
- Shadow cells (2 right, 1 below each window) have `WindowShadow` style with preserved character runes from the underlying content
- Clicking (Button1 press) on a back window brings it to front (OfTopSelect): after the event, it becomes the focused child and draws with active frame style
- Clicking the close icon `[×]` on a window results in `EvCommand/CmClose`, and Desktop removes the window
- Alt+1 selects window number 1 if such a window exists; Alt+2 selects window number 2
- `Desktop.Tile()` with 2 windows arranges them side-by-side (each gets half the desktop width)
- `Desktop.Cascade()` with 2 windows arranges them diagonally offset
- Three-phase dispatch: a child with `OfPreProcess` receives the event before the focused child; if it clears the event, the focused child does not see it
- Injecting a drag sequence (Button1 press on title bar → Button1 move → Button1 release) repositions the window
- During drag, mouse events outside the window bounds are still routed to the dragging window (mouse capture)
- Injecting a resize sequence (Button1 press on bottom-right corner → Button1 move → Button1 release) changes the window's size
- Double-click on the title bar (ClickCount=2) toggles zoom state

**Components to wire up:** Application, Desktop, Window, StatusLine, BorlandBlue (all real, no mocks for framework components)

**Run tests:** `go test ./tv/... -v -run TestIntegration -count=1`

**Commit:** `git commit -m "test: add Phase 2 integration tests for window management"`

---

### Task 9: Demo App + E2E Tests

- [ ] Complete

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**
- Demo app creates 2 windows: Window 1 "File Manager" at (5, 2, 35, 15) with number 1, Window 2 "Editor" at (20, 5, 40, 12) with number 2. Both inserted into the Desktop.
- Demo app retains existing StatusLine with "~Alt+X~ Exit" and "~F10~ Menu" items
- E2E test verifies window frames appear in the captured pane (look for box-drawing characters `╔`, `═`, or `┌`, `─`)
- E2E test verifies window title text ("File Manager" or "Editor") appears in the captured pane
- E2E test verifies desktop pattern `░` is still visible (in uncovered areas)
- E2E test sends Alt+X and verifies the app exits
- All existing E2E tests continue to pass (desktop pattern visible, status line contains "Alt+X", Alt+X exits)

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
		tv.NewStatusItem("~F10~ Menu", tv.KbFunc(10), tv.CmMenu),
	)

	app, err := tv.NewApplication(
		tv.WithStatusLine(statusLine),
		tv.WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		log.Fatal(err)
	}

	win1 := tv.NewWindow(tv.NewRect(5, 2, 35, 15), "File Manager", tv.WithWindowNumber(1))
	win2 := tv.NewWindow(tv.NewRect(20, 5, 40, 12), "Editor", tv.WithWindowNumber(2))

	app.Desktop().Insert(win1)
	app.Desktop().Insert(win2)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
```

```go
// e2e/e2e_test.go — extend TestBasicAppBoot

func TestBasicAppBoot(t *testing.T) {
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

	// Window frame characters visible (double-line border for active window)
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
```

**Run tests:** `go test ./e2e/... -v -count=1 -timeout 30s`

**Commit:** `git commit -m "feat: add windows to demo app and extend e2e tests"`
