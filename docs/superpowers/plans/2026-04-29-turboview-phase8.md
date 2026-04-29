# Phase 8: GrowMode Resize Cascade and Context Menus

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement GrowMode-based resize cascade so views automatically reposition and resize when their owner's bounds change (on terminal resize, Tile, Cascade, or manual window resize), and add context menu support for right-click popup menus.

**Architecture:** Group.SetBounds calculates the delta between old and new bounds and adjusts each child's bounds based on their GrowMode flags. This cascades recursively through the view tree. Context menus reuse the existing MenuPopup with a dedicated modal loop on Application, rendering the popup as an overlay on top of all other content.

**Tech Stack:** Go 1.22+, tcell/v2

---

## File Structure

### Modified Files
- `tv/group.go` — Override `SetBounds` to cascade GrowMode adjustments to children
- `tv/application.go` — Add `contextPopup` field, overlay rendering in `Draw`, `ContextMenu` method
- `e2e/testapp/basic/main.go` — Set GrowMode on windows and widgets, add right-click context menu demo
- `e2e/e2e_test.go` — New e2e tests for terminal resize and context menu

---

### Task 1: GrowMode Resize Cascade

**Files:**
- Modify: `tv/group.go`

**Requirements:**

GrowMode cascade on SetBounds:
- When `Group.SetBounds` is called with a new Rect that has different dimensions than the current bounds, each child's bounds are adjusted based on the child's GrowMode flags
- When the delta is zero (same width and height), no child bounds are modified
- `GfGrowLoX`: child's left edge (A.X) shifts by the width delta (owner grew wider → child shifts right)
- `GfGrowHiX`: child's right edge (B.X) shifts by the width delta (owner grew wider → child's right edge moves right, making child wider)
- `GfGrowLoY`: child's top edge (A.Y) shifts by the height delta (owner grew taller → child shifts down)
- `GfGrowHiY`: child's bottom edge (B.Y) shifts by the height delta (owner grew taller → child's bottom edge moves down, making child taller)
- `GfGrowAll` (all four flags): both edges on both axes shift by delta — the child shifts position but maintains its size
- A child with `GfGrowHiX | GfGrowHiY` at position (5,5) with size (30,15): when owner grows by deltaW=20, deltaH=5, the child becomes position (5,5) size (50,20) — it stretches
- A child with `GfGrowLoX | GfGrowLoY` at position (5,5) with size (10,3): when owner grows by deltaW=20, deltaH=5, the child becomes position (25,10) size (10,3) — it shifts to stay at the bottom-right region
- `GfGrowRel`: all four edges scale proportionally. If owner was 40x20 and becomes 80x40 (2x), a child at (10,5,20,10) becomes (20,10,40,20). Division by zero is safe — if old width or height is 0, skip proportional for that axis
- A child with GrowMode=0 is not moved or resized
- The cascade is recursive: if a child is a Container (Window, Dialog), and its `SetBounds` triggers its internal group's `SetBounds`, that group cascades to its own children

Window integration:
- When `Window.SetBounds` is called (it calls `group.SetBounds`), the GrowMode cascade propagates to widgets inside the window
- When `Desktop.SetBounds` is called (it calls `group.SetBounds`), the GrowMode cascade propagates to windows inside the desktop
- On terminal resize, `Application.layoutChildren` → `Desktop.SetBounds` → cascade to all windows with GrowMode → cascade to their widget children

**Implementation:**

Override `SetBounds` in `tv/group.go`:

```go
func (g *Group) SetBounds(r Rect) {
	oldW := g.size.X
	oldH := g.size.Y
	g.BaseView.SetBounds(r)
	newW := r.Width()
	newH := r.Height()

	deltaW := newW - oldW
	deltaH := newH - oldH

	if deltaW == 0 && deltaH == 0 {
		return
	}

	for _, child := range g.children {
		gm := child.GrowMode()
		if gm == 0 {
			continue
		}

		cb := child.Bounds()

		if gm&GfGrowRel != 0 {
			ax, ay := cb.A.X, cb.A.Y
			bx, by := cb.B.X, cb.B.Y
			if oldW > 0 {
				ax = ax * newW / oldW
				bx = bx * newW / oldW
			}
			if oldH > 0 {
				ay = ay * newH / oldH
				by = by * newH / oldH
			}
			child.SetBounds(Rect{A: Point{ax, ay}, B: Point{bx, by}})
			continue
		}

		ax, ay := cb.A.X, cb.A.Y
		bx, by := cb.B.X, cb.B.Y
		if gm&GfGrowLoX != 0 {
			ax += deltaW
		}
		if gm&GfGrowHiX != 0 {
			bx += deltaW
		}
		if gm&GfGrowLoY != 0 {
			ay += deltaH
		}
		if gm&GfGrowHiY != 0 {
			by += deltaH
		}
		child.SetBounds(Rect{A: Point{ax, ay}, B: Point{bx, by}})
	}
}
```

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: add GrowMode resize cascade in Group.SetBounds"`

---

### Task 2: Context Menu

**Files:**
- Modify: `tv/application.go`

**Requirements:**

Application.ContextMenu:
- `app.ContextMenu(x, y int, items ...any) CommandCode` creates a MenuPopup at position (x,y) and enters a modal event loop
- Keyboard events are forwarded to the popup: arrow keys navigate, Enter selects, Escape dismisses
- Mouse events inside the popup bounds are forwarded (with coordinates translated to popup-local space)
- Mouse click (Button1) outside the popup bounds dismisses the popup, returning CmCancel
- When an item is selected, the popup closes and the method returns the selected item's CommandCode
- When Escape is pressed, the method returns CmCancel
- The context popup is rendered on top of all other content during its modal loop (it appears visually above windows, desktop, menu bar, and status line)
- After the method returns, the popup is removed from rendering and the screen redraws cleanly
- Right-click (Button3) does not automatically trigger a context menu — views must explicitly call ContextMenu in their event handlers

Application.Draw overlay:
- When a `contextPopup` is active, it is drawn last in `Application.Draw`, on top of everything else
- The popup uses the Application's color scheme

**Implementation:**

Add `contextPopup` field to Application struct:

```go
type Application struct {
	bounds      Rect
	screen      tcell.Screen
	screenOwn   bool
	desktop     *Desktop
	statusLine  *StatusLine
	menuBar     *MenuBar
	scheme      *theme.ColorScheme
	quit        bool
	onCommand   func(CommandCode, any) bool
	configFile  string
	contextPopup *MenuPopup
}
```

Add overlay rendering at the end of `Application.Draw`:

```go
	// Context menu overlay (drawn last, on top of everything)
	if app.contextPopup != nil {
		pb := app.contextPopup.Bounds()
		app.contextPopup.Draw(buf.SubBuffer(pb), app.scheme)
	}
```

Add the ContextMenu method:

```go
func (app *Application) ContextMenu(x, y int, items ...any) CommandCode {
	popup := NewMenuPopup(items, x, y)
	app.contextPopup = popup
	app.drawAndFlush()

	var result CommandCode
	for result == 0 {
		event := app.PollEvent()
		if event == nil {
			result = CmCancel
			break
		}

		if event.What == EvKeyboard && event.Key != nil {
			popup.HandleEvent(event)
			if r := popup.Result(); r != 0 {
				result = r
				break
			}
		} else if event.What == EvMouse && event.Mouse != nil {
			pb := popup.Bounds()
			mx, my := event.Mouse.X, event.Mouse.Y
			if pb.Contains(NewPoint(mx, my)) {
				localEvent := *event
				localMouse := *event.Mouse
				localMouse.X = mx - pb.A.X
				localMouse.Y = my - pb.A.Y
				localEvent.Mouse = &localMouse
				popup.HandleEvent(&localEvent)
				if r := popup.Result(); r != 0 {
					result = r
					break
				}
			} else if event.Mouse.Button&tcell.Button1 != 0 {
				result = CmCancel
				break
			}
		}

		app.drawAndFlush()
	}

	app.contextPopup = nil
	app.drawAndFlush()

	if result == CmCancel {
		return CmCancel
	}
	return result
}
```

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: add context menu support via Application.ContextMenu"`

---

### Task 3: Integration Checkpoint — GrowMode + Context Menu

**Purpose:** Verify that GrowMode cascade works through the real view tree and that context menus render and dismiss correctly.

**Requirements (for test writer):**

GrowMode integration:
- Application with Desktop containing a Window. Window has a Button at position (2,2,10,2) with GfGrowLoX|GfGrowLoY. Resize the Desktop (via `app.Desktop().SetBounds(larger)` then letting it cascade). After resize, the Button's position has shifted right and down by the delta — it stays at the bottom-right region
- Application with a Window containing a widget with GfGrowHiX. Resize by calling Window.SetBounds with a wider window. Widget's width increases by the width delta
- GfGrowRel: a child at (10,5,20,10) in a 40x20 container. Container resizes to 80x40. Child becomes (20,10,40,20) — proportional scaling
- GrowMode=0 child is unchanged after container resize
- Recursive cascade: Desktop → Window (GfGrowHiX|GfGrowHiY) → Button (GfGrowLoX|GfGrowLoY). Resize Desktop. Window stretches. Button inside window shifts.

Context menu integration:
- Verify that `app.contextPopup` renders in the draw buffer at the correct position when set
- Verify that after ContextMenu returns, the contextPopup is nil and the overlay is gone

**Components to wire up:** Application (with SimulationScreen), Desktop, Window, Button (all real)

**Run:** `go test ./tv/... -run TestIntegration -v`

---

### Task 4: E2E Test

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**

Demo app changes:
- win2 gets `SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)` so it stretches when the desktop grows
- win2's ListViewer gets `SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)` so it stretches with the window
- win2's ScrollBar gets `SetGrowMode(tv.GfGrowLoX | tv.GfGrowHiY)` so it sticks to the right edge and stretches vertically

E2E tests:
- TestTerminalResize: Boot the app in an 80x25 tmux pane. Resize the pane to 100x30 using `tmux resize-window`. Wait briefly. Capture the output. Verify that the desktop pattern fills the new area (the desktop should be wider/taller). Verify the app is still running and responsive (Alt+X exits cleanly)
- TestContextMenu: This test verifies the build succeeds with the ContextMenu method available (a smoke test that the API compiles). Full interactive testing of context menus via tmux is fragile, so this test just verifies the app boots and exits cleanly with the new code

**Implementation:**

Add GrowMode to win2 and its children in `e2e/testapp/basic/main.go`:

```go
	win2.SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)
	lv.SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)
	sb.SetGrowMode(tv.GfGrowLoX | tv.GfGrowHiY)
```

Add resize e2e test to `e2e/e2e_test.go`.

**Run tests:** `cd /Users/gnat/Source/Personal/tv3 && go test ./e2e/... -v -timeout 120s`

**Commit:** `git commit -m "feat: add e2e tests for terminal resize with GrowMode"`
