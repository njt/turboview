# TListBox Implementation Plan (Batch 2, Phase 1)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build TListBox — a composite widget that bundles a ListViewer with a vertical ScrollBar, using the Group facade pattern.

**Architecture:** TListBox implements the Container interface, owns an internal Group as facade, and delegates all container operations. The ListViewer fills width-1, the ScrollBar occupies the right edge. They are wired via `ListViewer.SetScrollBar()`. This follows the same pattern as CheckBoxes and RadioButtons.

**Tech Stack:** Go, tcell/v2, existing tv package (ListViewer, ScrollBar, Group, BaseView)

---

## File Map

- **Create:** `tv/list_box.go` — TListBox widget
- **Modify:** `e2e/testapp/basic/main.go` — replace manual ListViewer+ScrollBar with ListBox
- **Modify:** `e2e/e2e_test.go` — add ListBox e2e test

---

- [ ] ### Task 1: ListBox Widget

**Files:**
- Create: `tv/list_box.go`

**Requirements:**
- `ListBox` implements `Container` interface
- `NewListBox(bounds Rect, ds ListDataSource) *ListBox` creates a ListBox with:
  - A vertical ScrollBar at the right edge (width 1, full height of bounds)
  - A ListViewer filling the remaining width (bounds width - 1, full height)
  - Wired via `ListViewer.SetScrollBar(sb)`
  - OfSelectable set on the ListBox itself
  - GrowMode: ListViewer gets `GfGrowHiX | GfGrowHiY`, ScrollBar gets `GfGrowLoX | GfGrowHiY`
  - SfVisible set
- `NewStringListBox(bounds Rect, items []string) *ListBox` is a convenience that wraps `NewStringList` + `NewListBox`
- Public API:
  - `ListViewer() *ListViewer` — access internal ListViewer
  - `ScrollBar() *ScrollBar` — access internal ScrollBar
  - `Selected() int` — delegates to ListViewer.Selected()
  - `SetSelected(index int)` — delegates to ListViewer.SetSelected()
  - `DataSource() ListDataSource` — returns the ListViewer's data source
  - `SetDataSource(ds ListDataSource)` — replaces data source, resets selection to 0, updates scrollbar range
- Container delegation: `Insert`, `Remove`, `Children`, `FocusedChild`, `SetFocusedChild`, `ExecView`, `BringToFront` — all delegate to internal Group
- `Draw(buf)` iterates children and draws each into its SubBuffer (same pattern as CheckBoxes.Draw)
- `HandleEvent(event)`:
  - Mouse events: route by position (clicks on ScrollBar area go to ScrollBar, clicks on ListViewer area go to ListViewer) — adjust coordinates to child-local space
  - Keyboard events: delegate to Group for three-phase dispatch (the ListViewer handles arrows/etc., ScrollBar is non-selectable)
- When ListBox receives `SfSelected`, the internal ListViewer gets focus via the Group
- The ScrollBar is NOT selectable (it responds to mouse but never gets keyboard focus)

**Implementation:**

```go
package tv

var _ Container = (*ListBox)(nil)

type ListBox struct {
	BaseView
	group    *Group
	viewer   *ListViewer
	scrollbar *ScrollBar
}

func NewListBox(bounds Rect, ds ListDataSource) *ListBox {
	lb := &ListBox{}
	lb.SetBounds(bounds)
	lb.SetState(SfVisible, true)
	lb.SetOptions(OfSelectable, true)

	lb.group = NewGroup(bounds)
	lb.group.SetFacade(lb)

	lb.viewer = NewListViewer(NewRect(0, 0, bounds.Width()-1, bounds.Height()), ds)
	lb.viewer.SetGrowMode(GfGrowHiX | GfGrowHiY)

	lb.scrollbar = NewScrollBar(NewRect(bounds.Width()-1, 0, 1, bounds.Height()), Vertical)
	lb.scrollbar.SetGrowMode(GfGrowLoX | GfGrowHiY)

	lb.viewer.SetScrollBar(lb.scrollbar)

	lb.group.Insert(lb.viewer)
	lb.group.Insert(lb.scrollbar)

	lb.SetSelf(lb)
	return lb
}

func NewStringListBox(bounds Rect, items []string) *ListBox {
	return NewListBox(bounds, NewStringList(items))
}

func (lb *ListBox) ListViewer() *ListViewer    { return lb.viewer }
func (lb *ListBox) ScrollBar() *ScrollBar      { return lb.scrollbar }
func (lb *ListBox) Selected() int              { return lb.viewer.Selected() }
func (lb *ListBox) SetSelected(index int)      { lb.viewer.SetSelected(index) }
func (lb *ListBox) DataSource() ListDataSource { return lb.viewer.DataSource() }

func (lb *ListBox) SetDataSource(ds ListDataSource) {
	lb.viewer.SetDataSource(ds)
}

// Container delegation
func (lb *ListBox) Insert(v View)               { lb.group.Insert(v) }
func (lb *ListBox) Remove(v View)               { lb.group.Remove(v) }
func (lb *ListBox) Children() []View            { return lb.group.Children() }
func (lb *ListBox) FocusedChild() View          { return lb.group.FocusedChild() }
func (lb *ListBox) SetFocusedChild(v View)      { lb.group.SetFocusedChild(v) }
func (lb *ListBox) ExecView(v View) CommandCode { return lb.group.ExecView(v) }
func (lb *ListBox) BringToFront(v View)         { lb.group.BringToFront(v) }

func (lb *ListBox) Draw(buf *DrawBuffer) {
	for _, child := range lb.group.Children() {
		childBounds := child.Bounds()
		sub := buf.SubBuffer(childBounds)
		child.Draw(sub)
	}
}

func (lb *ListBox) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		mx := event.Mouse.X
		for _, child := range lb.group.Children() {
			if child.Bounds().Contains(NewPoint(mx, event.Mouse.Y)) {
				origX, origY := event.Mouse.X, event.Mouse.Y
				event.Mouse.X -= child.Bounds().A.X
				event.Mouse.Y -= child.Bounds().A.Y
				child.HandleEvent(event)
				event.Mouse.X, event.Mouse.Y = origX, origY
				return
			}
		}
		return
	}
	lb.group.HandleEvent(event)
}
```

**Run tests:** `go test ./tv/... -run TestListBox -v`

**Commit:** `git commit -m "feat: add ListBox composite widget (ListViewer + ScrollBar)"`

---

- [ ] ### Task 2: Integration Checkpoint — ListBox

**Purpose:** Verify ListBox works correctly as a composite with real ListViewer and ScrollBar.

**Requirements (for test writer):**
- `NewListBox` creates a ListBox where `ListViewer()` and `ScrollBar()` return non-nil children
- `Selected()` and `SetSelected()` correctly delegate to the internal ListViewer
- Keyboard navigation (Down/Up arrows) through Group dispatch changes the selected item in the ListViewer
- ScrollBar value updates when the ListViewer selection moves past the visible area (scrollbar synchronization)
- Mouse clicks in the ListViewer area change selection; mouse clicks in the ScrollBar area scroll
- `NewStringListBox` convenience constructor produces a working ListBox with the expected items
- `SetDataSource` replaces items and resets selection to 0
- ListBox inserted into a Window/Group can be focused via Tab and responds to keyboard events

**Components to wire up:** ListBox, Window or Group (all real, no mocks)

**Run:** `go test ./tv/... -run TestIntegration -v`

---

- [ ] ### Task 3: E2E Test — ListBox in Demo App

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**
- Replace the manually-wired ListViewer+ScrollBar in Window 2 ("Editor") with a single `NewStringListBox` call
- The existing e2e tests that reference Window 2 behavior must still pass (scrollbar visible, list navigation works)
- Add a new e2e test `TestListBoxNavigation` that:
  - Switches to win2
  - Navigates down with arrow keys
  - Verifies the selection indicator moves
  - Verifies the scrollbar thumb position changes when scrolling past visible area

**Implementation:**

In `e2e/testapp/basic/main.go`, replace the manual ListViewer+ScrollBar setup (lines ~88-107) with:

```go
listBox := tv.NewStringListBox(tv.NewRect(0, 0, clientW, clientH), items)
win2.Insert(listBox)
listBox.SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)
```

Remove the separate `lv` and `sb` variables and their individual insertions.

The e2e test infrastructure uses tmux. Session names must be unique. Follow existing test patterns.

**Run tests:** `go test ./e2e/... -run TestListBox -v -timeout 60s`

**Commit:** `git commit -m "feat: replace manual ListViewer+ScrollBar with ListBox in demo app"`
