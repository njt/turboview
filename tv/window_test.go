package tv

// window_test.go — tests for Window (Task 2: Window Frame Drawing).
// Written against the spec before any implementation exists.
// Every assertion cites the spec sentence it verifies.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Compile-time interface check
// Spec: "Window satisfies the Container interface
//        (compile-time check: var _ Container = (*Window)(nil))"
// ---------------------------------------------------------------------------

var _ Container = (*Window)(nil)

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

// TestNewWindowSetsBounds verifies that NewWindow stores the given bounds.
// Spec: "NewWindow(bounds, title, opts...) creates a Window"
func TestNewWindowSetsBounds(t *testing.T) {
	r := NewRect(5, 10, 40, 20)
	w := NewWindow(r, "Test")

	if w.Bounds() != r {
		t.Errorf("NewWindow bounds = %v, want %v", w.Bounds(), r)
	}
}

// TestNewWindowSetsSfVisible verifies that NewWindow sets SfVisible.
// Spec: "NewWindow(bounds, title, opts...) creates a Window with SfVisible"
func TestNewWindowSetsSfVisible(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")

	if !w.HasState(SfVisible) {
		t.Errorf("NewWindow did not set SfVisible")
	}
}

// TestNewWindowMissingSfVisibleFailsSpec is a falsifying test: a Window without
// SfVisible would violate the spec, so a zero-state Window must fail this check.
// Spec: "NewWindow creates a Window with SfVisible"
func TestNewWindowMissingSfVisibleFailsSpec(t *testing.T) {
	// If we clear SfVisible the window should not have it.
	w := NewWindow(NewRect(0, 0, 40, 20), "Title")
	w.SetState(SfVisible, false)

	if w.HasState(SfVisible) {
		t.Errorf("SfVisible should be clearable; clearing it should result in HasState returning false")
	}
}

// TestNewWindowSetsOfSelectable verifies that NewWindow sets OfSelectable option.
// Spec: "NewWindow creates a Window with ... OfSelectable|OfTopSelect options"
func TestNewWindowSetsOfSelectable(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")

	if !w.HasOption(OfSelectable) {
		t.Errorf("NewWindow did not set OfSelectable option")
	}
}

// TestNewWindowSetsOfTopSelect verifies that NewWindow sets OfTopSelect option.
// Spec: "NewWindow creates a Window with ... OfSelectable|OfTopSelect options"
func TestNewWindowSetsOfTopSelect(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")

	if !w.HasOption(OfTopSelect) {
		t.Errorf("NewWindow did not set OfTopSelect option")
	}
}

// TestNewWindowTitleStored verifies NewWindow stores the title for retrieval.
// Spec: "Window.Title() returns the title string"
func TestNewWindowTitleStored(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "My Window")

	if w.Title() != "My Window" {
		t.Errorf("Title() = %q, want %q", w.Title(), "My Window")
	}
}

// TestNewWindowEmptyTitle verifies NewWindow accepts an empty title.
// Spec: "Window.Title() returns the title string" — title may be empty.
func TestNewWindowEmptyTitle(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "")

	if w.Title() != "" {
		t.Errorf("Title() = %q, want empty string", w.Title())
	}
}

// ---------------------------------------------------------------------------
// Title accessor
// ---------------------------------------------------------------------------

// TestSetTitleUpdatesTitle verifies SetTitle updates the title.
// Spec: "Window.SetTitle(t) updates the title"
func TestSetTitleUpdatesTitle(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Old Title")
	w.SetTitle("New Title")

	if w.Title() != "New Title" {
		t.Errorf("SetTitle: Title() = %q, want %q", w.Title(), "New Title")
	}
}

// TestSetTitleDoesNotRetainOldValue is a falsifying test: after SetTitle the old
// value must not be returned.
// Spec: "Window.SetTitle(t) updates the title"
func TestSetTitleDoesNotRetainOldValue(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Old")
	w.SetTitle("New")

	if w.Title() == "Old" {
		t.Errorf("SetTitle: Title() still returns old value %q", w.Title())
	}
}

// ---------------------------------------------------------------------------
// Window number
// ---------------------------------------------------------------------------

// TestNumberDefaultsToZero verifies Window.Number() returns 0 when not set.
// Spec: "Window.Number() returns the window number (0 if not set)"
func TestNumberDefaultsToZero(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")

	if w.Number() != 0 {
		t.Errorf("Number() without WithWindowNumber = %d, want 0", w.Number())
	}
}

// TestWithWindowNumberSetsNumber verifies WithWindowNumber option stores the number.
// Spec: "WindowOption type for functional options: WithWindowNumber(n int)"
//       "Window.Number() returns the window number"
func TestWithWindowNumberSetsNumber(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test", WithWindowNumber(7))

	if w.Number() != 7 {
		t.Errorf("Number() with WithWindowNumber(7) = %d, want 7", w.Number())
	}
}

// TestWithWindowNumberZeroExplicit verifies passing 0 explicitly keeps Number() at 0.
// Spec: "Window.Number() returns the window number (0 if not set)"
func TestWithWindowNumberZeroExplicit(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test", WithWindowNumber(0))

	if w.Number() != 0 {
		t.Errorf("Number() with WithWindowNumber(0) = %d, want 0", w.Number())
	}
}

// TestWithWindowNumberFalsify verifies that without the option the number is not non-zero.
// Spec: "Window.Number() returns the window number (0 if not set)"
func TestWithWindowNumberFalsify(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")

	// Without the option we must not get a non-zero number.
	if w.Number() != 0 {
		t.Errorf("Number() without option = %d, want 0", w.Number())
	}
}

// ---------------------------------------------------------------------------
// Container delegation
// ---------------------------------------------------------------------------

// TestWindowInsertDelegatesToGroup verifies Insert adds a child (via internal Group).
// Spec: "Window delegates Container methods (Insert, ...) to its internal Group"
func TestWindowInsertDelegatesToGroup(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	child := newMockView(NewRect(0, 0, 5, 5))

	w.Insert(child)

	children := w.Children()
	if len(children) != 1 || children[0] != child {
		t.Errorf("Insert: Children() = %v, want [child]", children)
	}
}

// TestWindowRemoveDelegatesToGroup verifies Remove removes a child (via internal Group).
// Spec: "Window delegates Container methods (..., Remove, ...) to its internal Group"
func TestWindowRemoveDelegatesToGroup(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	child := newMockView(NewRect(0, 0, 5, 5))
	w.Insert(child)

	w.Remove(child)

	for _, c := range w.Children() {
		if c == child {
			t.Errorf("Remove: child still present in Children()")
		}
	}
}

// TestWindowChildrenDelegatesToGroup verifies Children() returns group's children.
// Spec: "Window delegates Container methods (..., Children, ...) to its internal Group"
func TestWindowChildrenDelegatesToGroup(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	child1 := newMockView(NewRect(0, 0, 5, 5))
	child2 := newMockView(NewRect(5, 0, 5, 5))
	w.Insert(child1)
	w.Insert(child2)

	children := w.Children()

	if len(children) != 2 {
		t.Errorf("Children() len = %d, want 2", len(children))
	}
}

// TestWindowFocusedChildDelegatesToGroup verifies FocusedChild returns the group's focused child.
// Spec: "Window delegates Container methods (..., FocusedChild, ...) to its internal Group"
func TestWindowFocusedChildDelegatesToGroup(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 5, 5))
	w.Insert(child)

	got := w.FocusedChild()

	if got != child {
		t.Errorf("FocusedChild() = %v, want inserted selectable child", got)
	}
}

// TestWindowSetFocusedChildDelegatesToGroup verifies SetFocusedChild delegates to group.
// Spec: "Window delegates Container methods (..., SetFocusedChild, ...) to its internal Group"
func TestWindowSetFocusedChildDelegatesToGroup(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	first := newSelectableMockView(NewRect(0, 0, 5, 5))
	second := newSelectableMockView(NewRect(5, 0, 5, 5))
	w.Insert(first)
	w.Insert(second)

	w.SetFocusedChild(first)

	if w.FocusedChild() != first {
		t.Errorf("SetFocusedChild: FocusedChild() = %v, want first", w.FocusedChild())
	}
}


// TestWindowInsertSetsChildOwnerToWindow verifies that children's Owner is the Window
// (the facade), not the internal group.
// Spec: "an internal Group whose facade is set to the Window"
func TestWindowInsertSetsChildOwnerToWindow(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	child := newMockView(NewRect(0, 0, 5, 5))

	w.Insert(child)

	if child.Owner() != w {
		t.Errorf("child.Owner() = %v, want Window (facade)", child.Owner())
	}
}

// ---------------------------------------------------------------------------
// SetBounds
// ---------------------------------------------------------------------------

// TestSetBoundsUpdatesWindowBounds verifies SetBounds updates the Window's bounds.
// Spec: "Window.SetBounds(r) updates both Window's bounds and the internal Group's bounds"
func TestSetBoundsUpdatesWindowBounds(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	newBounds := NewRect(5, 5, 60, 30)

	w.SetBounds(newBounds)

	if w.Bounds() != newBounds {
		t.Errorf("SetBounds: Window.Bounds() = %v, want %v", w.Bounds(), newBounds)
	}
}

// TestSetBoundsUpdatesGroupClientArea verifies the internal Group's bounds become
// origin (0,0) with size (width-2, height-2).
// Spec: "Group uses origin 0,0, width-2, height-2 for the client area"
func TestSetBoundsUpdatesGroupClientArea(t *testing.T) {
	// Use a window large enough to inspect child placement.
	// We'll verify by inserting a child that fills the client area and checking
	// where it lands in Draw output.
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	// Insert a child that writes its own rune at (0,0) of whatever buffer it gets.
	sentinel := &drawOrderMockView{id: 'X', bounds: NewRect(0, 0, 38, 18)}
	sentinel.SetState(SfVisible, true)
	w.Insert(sentinel)

	buf := NewDrawBuffer(40, 20)
	w.Draw(buf)

	// The child draws 'X' at (0,0) of the client subbuffer, which corresponds to
	// absolute cell (1,1) in the window buffer.
	cell := buf.GetCell(1, 1)
	if cell.Rune != 'X' {
		t.Errorf("SetBounds client area: cell(1,1) = %q, want 'X' (child rendered in client subbuffer)", cell.Rune)
	}
}

// TestSetBoundsGroupOriginIsZeroZero verifies the group's origin is (0,0) not the
// window's origin.
// Spec: "Group uses origin 0,0, width-2, height-2 for the client area"
func TestSetBoundsGroupOriginIsZeroZero(t *testing.T) {
	// Verify by using SetBounds after construction with a non-zero origin.
	w := NewWindow(NewRect(10, 10, 40, 20), "Test")
	newBounds := NewRect(5, 5, 50, 25)
	w.SetBounds(newBounds)

	// The window itself has origin (5,5), but the group client area starts at (0,0)
	// relative to the window. A child at (0,0) in client coords should appear at
	// absolute (6,6) — that is, window origin (5,5) + border (1,1).
	sentinel := &drawOrderMockView{id: 'Z', bounds: NewRect(0, 0, 48, 23)}
	sentinel.SetState(SfVisible, true)
	w.Insert(sentinel)

	// Draw into a buffer sized for the full absolute area.
	buf := NewDrawBuffer(60, 35)
	w.Draw(buf)

	// Because DrawBuffer uses offsets from the containing context, and Window.Draw
	// will create a SubBuffer for the client area at (1,1), the child's (0,0) lands
	// at buf absolute (1,1) from the window's draw start (since NewWindow draws into
	// the buffer it receives without re-offsetting for its own position — the caller
	// is responsible for sub-buffering to the window's bounds).
	// So cell (1,1) in buf should be 'Z'.
	cell := buf.GetCell(1, 1)
	if cell.Rune != 'Z' {
		t.Errorf("Group origin: cell(1,1) = %q, want 'Z' (client area starts at (1,1))", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Draw: too-small no-op
// ---------------------------------------------------------------------------

// TestDrawNoOpWhenWidthTooSmall verifies Draw does nothing when width < 8.
// Spec: "If window is too small (width < 8 or height < 3), Draw is a no-op"
func TestDrawNoOpWhenWidthTooSmall(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 7, 10), "Hi")
	w.SetState(SfSelected, true)
	buf := NewDrawBuffer(7, 10)

	w.Draw(buf)

	// No frame characters should have been written; cell (0,0) should be the
	// default space rune from NewDrawBuffer.
	cell := buf.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Errorf("Draw no-op (width<8): cell(0,0) = %q, want ' ' (no-op)", cell.Rune)
	}
}

// TestDrawNoOpWhenHeightTooSmall verifies Draw does nothing when height < 3.
// Spec: "If window is too small (width < 8 or height < 3), Draw is a no-op"
func TestDrawNoOpWhenHeightTooSmall(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 2), "Hi")
	w.SetState(SfSelected, true)
	buf := NewDrawBuffer(20, 2)

	w.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Errorf("Draw no-op (height<3): cell(0,0) = %q, want ' ' (no-op)", cell.Rune)
	}
}

// TestDrawProceedsWhenExactlyMinimumSize verifies Draw renders for a window that
// is exactly at the minimum permitted size (width==8, height==3).
// Spec: "If window is too small (width < 8 or height < 3), Draw is a no-op" —
// a window of exactly 8×3 is NOT too small.
func TestDrawProceedsWhenExactlyMinimumSize(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 8, 3), "Hi")
	w.SetState(SfSelected, true)
	buf := NewDrawBuffer(8, 3)

	w.Draw(buf)

	// Should have written the top-left active frame corner '╔'.
	cell := buf.GetCell(0, 0)
	if cell.Rune != '╔' {
		t.Errorf("Draw at min size: cell(0,0) = %q, want '╔'", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Draw: active frame
// ---------------------------------------------------------------------------

// newWindowForDraw creates a 20×10 window with a colour scheme attached so that
// Draw tests can inspect frame characters and styles without needing a real owner.
func newWindowForDraw(title string) (*Window, *DrawBuffer, *theme.ColorScheme) {
	cs := &theme.ColorScheme{
		WindowBackground:    tcell.StyleDefault.Background(tcell.ColorNavy),
		WindowFrameActive:   tcell.StyleDefault.Foreground(tcell.ColorWhite),
		WindowFrameInactive: tcell.StyleDefault.Foreground(tcell.ColorSilver),
		WindowTitle:         tcell.StyleDefault.Foreground(tcell.ColorYellow),
	}
	w := NewWindow(NewRect(0, 0, 20, 10), title)
	w.scheme = cs
	buf := NewDrawBuffer(20, 10)
	return w, buf, cs
}

// TestDrawActiveFrameTopLeft verifies the top-left corner is '╔' when SfSelected is set.
// Spec: "Active frame (when SfSelected is set): double-line border using ╔═╗║╚═╝"
func TestDrawActiveFrameTopLeft(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.SetState(SfSelected, true)

	w.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != '╔' {
		t.Errorf("Active frame top-left: cell(0,0) = %q, want '╔'", cell.Rune)
	}
}

// TestDrawActiveFrameTopRight verifies the top-right corner is '╗' when SfSelected.
// Spec: "Active frame ... double-line border using ╔═╗║╚═╝"
func TestDrawActiveFrameTopRight(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.SetState(SfSelected, true)
	width := w.Bounds().Width()

	w.Draw(buf)

	cell := buf.GetCell(width-1, 0)
	if cell.Rune != '╗' {
		t.Errorf("Active frame top-right: cell(%d,0) = %q, want '╗'", width-1, cell.Rune)
	}
}

// TestDrawActiveFrameBottomLeft verifies the bottom-left corner is '╚' when SfSelected.
// Spec: "Active frame ... double-line border using ╔═╗║╚═╝"
func TestDrawActiveFrameBottomLeft(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.SetState(SfSelected, true)
	height := w.Bounds().Height()

	w.Draw(buf)

	cell := buf.GetCell(0, height-1)
	if cell.Rune != '╚' {
		t.Errorf("Active frame bottom-left: cell(0,%d) = %q, want '╚'", height-1, cell.Rune)
	}
}

// TestDrawActiveFrameBottomRight verifies the bottom-right corner is '╝' when SfSelected.
// Spec: "Active frame ... double-line border using ╔═╗║╚═╝"
func TestDrawActiveFrameBottomRight(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.SetState(SfSelected, true)
	bounds := w.Bounds()

	w.Draw(buf)

	cell := buf.GetCell(bounds.Width()-1, bounds.Height()-1)
	if cell.Rune != '╝' {
		t.Errorf("Active frame bottom-right: cell(%d,%d) = %q, want '╝'",
			bounds.Width()-1, bounds.Height()-1, cell.Rune)
	}
}

// TestDrawActiveFrameTopEdge verifies the top horizontal fill uses '═' when SfSelected.
// Spec: "Active frame ... double-line border using ╔═╗║╚═╝"
func TestDrawActiveFrameTopEdge(t *testing.T) {
	w, buf, _ := newWindowForDraw("")
	w.SetState(SfSelected, true)
	width := w.Bounds().Width()

	w.Draw(buf)

	// Cell (4,0) is past the close icon [×] at (1-3,0).
	// Use a position in the top edge that is not a corner and not occupied by icons.
	// Close icon occupies (1,2,3); zoom icon occupies (width-4, width-3, width-2).
	// Safe position: width/2 if it avoids icons. For width=20: cell (10,0).
	midX := width / 2
	cell := buf.GetCell(midX, 0)
	if cell.Rune != '═' {
		t.Errorf("Active frame top edge: cell(%d,0) = %q, want '═'", midX, cell.Rune)
	}
}

// TestDrawActiveFrameLeftEdge verifies the left vertical edge uses '║' when SfSelected.
// Spec: "Active frame ... double-line border using ╔═╗║╚═╝"
func TestDrawActiveFrameLeftEdge(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.SetState(SfSelected, true)

	w.Draw(buf)

	cell := buf.GetCell(0, 5)
	if cell.Rune != '║' {
		t.Errorf("Active frame left edge: cell(0,5) = %q, want '║'", cell.Rune)
	}
}

// TestDrawActiveFrameRightEdge verifies the right vertical edge uses '║' when SfSelected.
// Spec: "Active frame ... double-line border using ╔═╗║╚═╝"
func TestDrawActiveFrameRightEdge(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.SetState(SfSelected, true)
	width := w.Bounds().Width()

	w.Draw(buf)

	cell := buf.GetCell(width-1, 5)
	if cell.Rune != '║' {
		t.Errorf("Active frame right edge: cell(%d,5) = %q, want '║'", width-1, cell.Rune)
	}
}

// TestDrawActiveFrameBottomEdge verifies the bottom horizontal fill uses '═' when SfSelected.
// Spec: "Active frame ... double-line border using ╔═╗║╚═╝"
func TestDrawActiveFrameBottomEdge(t *testing.T) {
	w, buf, _ := newWindowForDraw("")
	w.SetState(SfSelected, true)
	bounds := w.Bounds()

	w.Draw(buf)

	cell := buf.GetCell(bounds.Width()/2, bounds.Height()-1)
	if cell.Rune != '═' {
		t.Errorf("Active frame bottom edge: cell(%d,%d) = %q, want '═'",
			bounds.Width()/2, bounds.Height()-1, cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Draw: inactive frame
// ---------------------------------------------------------------------------

// TestDrawInactiveFrameTopLeft verifies the top-left corner is '┌' when SfSelected is NOT set.
// Spec: "Inactive frame (when SfSelected is not set): single-line border using ┌─┐│└─┘"
func TestDrawInactiveFrameTopLeft(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	// SfSelected is NOT set (default)

	w.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != '┌' {
		t.Errorf("Inactive frame top-left: cell(0,0) = %q, want '┌'", cell.Rune)
	}
}

// TestDrawInactiveFrameTopRight verifies the top-right corner is '┐' when SfSelected is not set.
// Spec: "Inactive frame (when SfSelected is not set): single-line border using ┌─┐│└─┘"
func TestDrawInactiveFrameTopRight(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	width := w.Bounds().Width()

	w.Draw(buf)

	cell := buf.GetCell(width-1, 0)
	if cell.Rune != '┐' {
		t.Errorf("Inactive frame top-right: cell(%d,0) = %q, want '┐'", width-1, cell.Rune)
	}
}

// TestDrawInactiveFrameBottomLeft verifies the bottom-left corner is '└' when inactive.
// Spec: "Inactive frame ... single-line border using ┌─┐│└─┘"
func TestDrawInactiveFrameBottomLeft(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	height := w.Bounds().Height()

	w.Draw(buf)

	cell := buf.GetCell(0, height-1)
	if cell.Rune != '└' {
		t.Errorf("Inactive frame bottom-left: cell(0,%d) = %q, want '└'", height-1, cell.Rune)
	}
}

// TestDrawInactiveFrameBottomRight verifies the bottom-right corner is '┘' when inactive.
// Spec: "Inactive frame ... single-line border using ┌─┐│└─┘"
func TestDrawInactiveFrameBottomRight(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	bounds := w.Bounds()

	w.Draw(buf)

	cell := buf.GetCell(bounds.Width()-1, bounds.Height()-1)
	if cell.Rune != '┘' {
		t.Errorf("Inactive frame bottom-right: cell(%d,%d) = %q, want '┘'",
			bounds.Width()-1, bounds.Height()-1, cell.Rune)
	}
}

// TestDrawInactiveFrameTopEdge verifies the top horizontal fill uses '─' when inactive.
// Spec: "Inactive frame ... single-line border using ┌─┐│└─┘"
func TestDrawInactiveFrameTopEdge(t *testing.T) {
	w, buf, _ := newWindowForDraw("")
	width := w.Bounds().Width()

	w.Draw(buf)

	midX := width / 2
	cell := buf.GetCell(midX, 0)
	if cell.Rune != '─' {
		t.Errorf("Inactive frame top edge: cell(%d,0) = %q, want '─'", midX, cell.Rune)
	}
}

// TestDrawInactiveFrameLeftEdge verifies the left vertical edge uses '│' when inactive.
// Spec: "Inactive frame ... single-line border using ┌─┐│└─┘"
func TestDrawInactiveFrameLeftEdge(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")

	w.Draw(buf)

	cell := buf.GetCell(0, 5)
	if cell.Rune != '│' {
		t.Errorf("Inactive frame left edge: cell(0,5) = %q, want '│'", cell.Rune)
	}
}

// TestDrawInactiveFrameRightEdge verifies the right vertical edge uses '│' when inactive.
// Spec: "Inactive frame ... single-line border using ┌─┐│└─┘"
func TestDrawInactiveFrameRightEdge(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	width := w.Bounds().Width()

	w.Draw(buf)

	cell := buf.GetCell(width-1, 5)
	if cell.Rune != '│' {
		t.Errorf("Inactive frame right edge: cell(%d,5) = %q, want '│'", width-1, cell.Rune)
	}
}

// TestDrawInactiveFrameBottomEdge verifies the bottom horizontal fill uses '─' when inactive.
// Spec: "Inactive frame ... single-line border using ┌─┐│└─┘"
func TestDrawInactiveFrameBottomEdge(t *testing.T) {
	w, buf, _ := newWindowForDraw("")
	bounds := w.Bounds()

	w.Draw(buf)

	cell := buf.GetCell(bounds.Width()/2, bounds.Height()-1)
	if cell.Rune != '─' {
		t.Errorf("Inactive frame bottom edge: cell(%d,%d) = %q, want '─'",
			bounds.Width()/2, bounds.Height()-1, cell.Rune)
	}
}

// TestDrawActiveNotInactiveTopLeft verifies the active/inactive frames are distinct:
// an active window must NOT have the single-line '┌' corner.
// Spec: uses distinct characters for active vs inactive.
func TestDrawActiveNotInactiveTopLeft(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.SetState(SfSelected, true)

	w.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == '┌' {
		t.Errorf("Active frame: cell(0,0) = '┌' (inactive corner), want '╔'")
	}
}

// ---------------------------------------------------------------------------
// Draw: frame style
// ---------------------------------------------------------------------------

// TestDrawActiveFrameUsesWindowFrameActiveStyle verifies active frame cells use
// the WindowFrameActive colour scheme style.
// Spec: "Active frame ... in WindowFrameActive style"
func TestDrawActiveFrameUsesWindowFrameActiveStyle(t *testing.T) {
	w, buf, cs := newWindowForDraw("Test")
	w.SetState(SfSelected, true)

	w.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != cs.WindowFrameActive {
		t.Errorf("Active frame style: cell(0,0).Style = %v, want WindowFrameActive %v",
			cell.Style, cs.WindowFrameActive)
	}
}

// TestDrawInactiveFrameUsesWindowFrameInactiveStyle verifies inactive frame cells use
// the WindowFrameInactive style.
// Spec: "Inactive frame ... in WindowFrameInactive style"
func TestDrawInactiveFrameUsesWindowFrameInactiveStyle(t *testing.T) {
	w, buf, cs := newWindowForDraw("Test")
	// SfSelected NOT set → inactive

	w.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != cs.WindowFrameInactive {
		t.Errorf("Inactive frame style: cell(0,0).Style = %v, want WindowFrameInactive %v",
			cell.Style, cs.WindowFrameInactive)
	}
}

// ---------------------------------------------------------------------------
// Draw: close icon
// ---------------------------------------------------------------------------

// TestDrawCloseIconOpenBracket verifies '[' is at position (1,0).
// Spec: "Close icon [×] at positions (1,0)-(3,0) of top border"
func TestDrawCloseIconOpenBracket(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Rune != '[' {
		t.Errorf("Close icon '[': cell(1,0) = %q, want '['", cell.Rune)
	}
}

// TestDrawCloseIconX verifies '×' is at position (2,0).
// Spec: "Close icon [×] at positions (1,0)-(3,0) of top border"
func TestDrawCloseIconX(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.Draw(buf)

	cell := buf.GetCell(2, 0)
	if cell.Rune != '×' {
		t.Errorf("Close icon '×': cell(2,0) = %q, want '×'", cell.Rune)
	}
}

// TestDrawCloseIconCloseBracket verifies ']' is at position (3,0).
// Spec: "Close icon [×] at positions (1,0)-(3,0) of top border"
func TestDrawCloseIconCloseBracket(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.Draw(buf)

	cell := buf.GetCell(3, 0)
	if cell.Rune != ']' {
		t.Errorf("Close icon ']': cell(3,0) = %q, want ']'", cell.Rune)
	}
}

// TestDrawCloseIconUsesFrameStyle verifies the close icon is rendered in the same
// frame style as the rest of the border.
// Spec: "Close icon [×] at positions (1,0)-(3,0) of top border in frame style"
func TestDrawCloseIconUsesFrameStyle(t *testing.T) {
	w, buf, cs := newWindowForDraw("Test")
	// Inactive (SfSelected not set) so we compare against WindowFrameInactive.
	w.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Style != cs.WindowFrameInactive {
		t.Errorf("Close icon style: cell(1,0).Style = %v, want WindowFrameInactive", cell.Style)
	}
}

// TestDrawCloseIconActiveFrameStyle verifies the close icon uses WindowFrameActive
// when the window is selected.
// Spec: "Close icon [×] at positions (1,0)-(3,0) of top border in frame style"
func TestDrawCloseIconActiveFrameStyle(t *testing.T) {
	w, buf, cs := newWindowForDraw("Test")
	w.SetState(SfSelected, true)
	w.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Style != cs.WindowFrameActive {
		t.Errorf("Close icon active style: cell(1,0).Style = %v, want WindowFrameActive", cell.Style)
	}
}

// ---------------------------------------------------------------------------
// Draw: zoom icon
// ---------------------------------------------------------------------------

// TestDrawZoomIconOpenBracket verifies '[' is at position (width-4, 0).
// Spec: "Zoom icon [↑] at positions (width-4,0)-(width-2,0)"
func TestDrawZoomIconOpenBracket(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.Draw(buf)
	width := w.Bounds().Width()

	cell := buf.GetCell(width-4, 0)
	if cell.Rune != '[' {
		t.Errorf("Zoom icon '[': cell(%d,0) = %q, want '['", width-4, cell.Rune)
	}
}

// TestDrawZoomIconArrow verifies '↑' is at position (width-3, 0).
// Spec: "Zoom icon [↑] at positions (width-4,0)-(width-2,0)"
func TestDrawZoomIconArrow(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.Draw(buf)
	width := w.Bounds().Width()

	cell := buf.GetCell(width-3, 0)
	if cell.Rune != '↑' {
		t.Errorf("Zoom icon '↑': cell(%d,0) = %q, want '↑'", width-3, cell.Rune)
	}
}

// TestDrawZoomIconCloseBracket verifies ']' is at position (width-2, 0).
// Spec: "Zoom icon [↑] at positions (width-4,0)-(width-2,0)"
func TestDrawZoomIconCloseBracket(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	w.Draw(buf)
	width := w.Bounds().Width()

	cell := buf.GetCell(width-2, 0)
	if cell.Rune != ']' {
		t.Errorf("Zoom icon ']': cell(%d,0) = %q, want ']'", width-2, cell.Rune)
	}
}

// TestDrawZoomIconUsesFrameStyle verifies the zoom icon uses the frame style.
// Spec: "Zoom icon [↑] at positions (width-4,0)-(width-2,0) in frame style"
func TestDrawZoomIconUsesFrameStyle(t *testing.T) {
	w, buf, cs := newWindowForDraw("Test")
	// Inactive
	w.Draw(buf)
	width := w.Bounds().Width()

	cell := buf.GetCell(width-4, 0)
	if cell.Style != cs.WindowFrameInactive {
		t.Errorf("Zoom icon style: cell(%d,0).Style = %v, want WindowFrameInactive", width-4, cell.Style)
	}
}

// ---------------------------------------------------------------------------
// Draw: title
// ---------------------------------------------------------------------------

// TestDrawTitleRenderedInWindowTitleStyle verifies the title uses WindowTitle style.
// Spec: "Title centered between close and zoom icons, wrapped in spaces, in WindowTitle style"
func TestDrawTitleRenderedInWindowTitleStyle(t *testing.T) {
	w, buf, cs := newWindowForDraw("Hi")
	w.Draw(buf)

	// The title " Hi " starts after position (4,0) (past the close icon).
	// We find the 'H' character and check its style.
	width := w.Bounds().Width()
	foundH := false
	for x := 4; x < width-4; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune == 'H' {
			foundH = true
			if cell.Style != cs.WindowTitle {
				t.Errorf("Title rune 'H' at (%d,0): style = %v, want WindowTitle %v",
					x, cell.Style, cs.WindowTitle)
			}
			break
		}
	}
	if !foundH {
		t.Errorf("Title 'Hi' not found in top border between close and zoom icons")
	}
}

// TestDrawTitleCenteredBetweenIcons verifies the title is rendered between the
// close icon end (col 4) and the zoom icon start (width-4).
// Spec: "Title centered between close and zoom icons, wrapped in spaces"
func TestDrawTitleCenteredBetweenIcons(t *testing.T) {
	w, buf, _ := newWindowForDraw("A")
	w.Draw(buf)
	width := w.Bounds().Width()

	// The title must appear between columns 4 and width-4 (exclusive).
	// Look for the 'A' rune in that range.
	found := false
	for x := 4; x < width-4; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune == 'A' {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Title 'A' not found between close icon (col 4) and zoom icon (col %d)", width-4)
	}
}

// TestDrawTitleWrappedInSpaces verifies the title has a leading space before the
// first character and a trailing space after the last.
// Spec: "Title centered between close and zoom icons, wrapped in spaces, in WindowTitle style"
func TestDrawTitleWrappedInSpaces(t *testing.T) {
	w, buf, cs := newWindowForDraw("XY")
	w.Draw(buf)
	width := w.Bounds().Width()

	// Find "XY" in the title area and check that the immediately adjacent cells
	// (on both sides) are spaces in WindowTitle style.
	xPos := -1
	for x := 4; x < width-4; x++ {
		if buf.GetCell(x, 0).Rune == 'X' && buf.GetCell(x+1, 0).Rune == 'Y' {
			xPos = x
			break
		}
	}
	if xPos < 0 {
		t.Fatal("Title 'XY' not found in top border")
	}
	before := buf.GetCell(xPos-1, 0)
	after := buf.GetCell(xPos+2, 0)
	if before.Rune != ' ' || before.Style != cs.WindowTitle {
		t.Errorf("Title leading space: cell(%d,0) = %q style=%v, want ' ' WindowTitle",
			xPos-1, before.Rune, before.Style)
	}
	if after.Rune != ' ' || after.Style != cs.WindowTitle {
		t.Errorf("Title trailing space: cell(%d,0) = %q style=%v, want ' ' WindowTitle",
			xPos+2, after.Rune, after.Style)
	}
}

// TestDrawTitleNotDrawnOutsideIconBounds verifies the title does not bleed into
// or overwrite the close icon area (columns 1-3) or zoom icon area (width-4 .. width-2).
// This is a falsifying test against a naive implementation that ignores icon positions.
// Spec: "Title centered between close and zoom icons"
func TestDrawTitleNotDrawnOutsideIconBounds(t *testing.T) {
	w, buf, _ := newWindowForDraw("AAAAAAAAAAAAAAAA") // very long title
	w.Draw(buf)

	// Column 1 must be '[' (close icon), not a title char.
	cell := buf.GetCell(1, 0)
	if cell.Rune != '[' {
		t.Errorf("Close icon overwritten by title: cell(1,0) = %q, want '['", cell.Rune)
	}
	width := w.Bounds().Width()
	// Column width-4 must be '[' (zoom icon), not a title char.
	cell = buf.GetCell(width-4, 0)
	if cell.Rune != '[' {
		t.Errorf("Zoom icon overwritten by title: cell(%d,0) = %q, want '['", width-4, cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Draw: client area background
// ---------------------------------------------------------------------------

// TestDrawClientAreaBackground verifies the client area is filled with
// WindowBackground style and a space rune.
// Spec: "Client area background filled with ColorScheme().WindowBackground style
//        and space rune, at (1,1) with size (width-2, height-2)"
func TestDrawClientAreaBackground(t *testing.T) {
	w, buf, cs := newWindowForDraw("Test")
	w.Draw(buf)

	// Cell (1,1) is the top-left of the client area.
	cell := buf.GetCell(1, 1)
	if cell.Rune != ' ' {
		t.Errorf("Client area bg: cell(1,1) rune = %q, want ' '", cell.Rune)
	}
	if cell.Style != cs.WindowBackground {
		t.Errorf("Client area bg: cell(1,1) style = %v, want WindowBackground %v",
			cell.Style, cs.WindowBackground)
	}
}

// TestDrawClientAreaBackgroundFillsEntireArea verifies the fill covers the full
// client rectangle — not just one corner cell.
// Spec: "Client area background ... at (1,1) with size (width-2, height-2)"
func TestDrawClientAreaBackgroundFillsEntireArea(t *testing.T) {
	w, buf, cs := newWindowForDraw("Test")
	bounds := w.Bounds()
	w.Draw(buf)

	// Check an interior cell (not at (1,1) to confirm it's a fill, not a point write).
	interiorX := bounds.Width() / 2
	interiorY := bounds.Height() / 2
	cell := buf.GetCell(interiorX, interiorY)
	if cell.Style != cs.WindowBackground {
		t.Errorf("Client area bg interior: cell(%d,%d) style = %v, want WindowBackground",
			interiorX, interiorY, cell.Style)
	}
}

// TestDrawClientAreaDoesNotCoverBorder verifies the border cells are not filled
// with WindowBackground — the fill is at (1,1) not (0,0).
// Spec: "Client area background ... at (1,1) with size (width-2, height-2)"
func TestDrawClientAreaDoesNotCoverBorder(t *testing.T) {
	w, buf, cs := newWindowForDraw("Test")
	w.Draw(buf)

	// (0,0) is the frame corner; it must not have the WindowBackground style.
	cell := buf.GetCell(0, 0)
	if cell.Style == cs.WindowBackground {
		t.Errorf("Border cell (0,0) has WindowBackground style — fill extended past client area")
	}
}

// ---------------------------------------------------------------------------
// Draw: children in client area SubBuffer
// ---------------------------------------------------------------------------

// TestDrawChildrenDrawnIntoClientSubBuffer verifies that children are drawn
// into a SubBuffer starting at (1,1), so a child writing at its own (0,0)
// appears at absolute (1,1) in the window buffer.
// Spec: "Children drawn into a client-area SubBuffer at (1,1) with size
//        (width-2, height-2) via Group.Draw"
func TestDrawChildrenDrawnIntoClientSubBuffer(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	sentinel := &drawOrderMockView{id: 'Q', bounds: NewRect(0, 0, 18, 8)}
	sentinel.SetState(SfVisible, true)
	w.Insert(sentinel)

	w.Draw(buf)

	cell := buf.GetCell(1, 1)
	if cell.Rune != 'Q' {
		t.Errorf("Children in client SubBuffer: cell(1,1) = %q, want 'Q' (child's (0,0))", cell.Rune)
	}
}

// TestDrawChildrenNotRenderedOutsideClientArea verifies children cannot paint
// over the window border — they're clipped to the client SubBuffer.
// Spec: "Children drawn into a client-area SubBuffer at (1,1) with size (width-2, height-2)"
func TestDrawChildrenNotRenderedOutsideClientArea(t *testing.T) {
	w, buf, _ := newWindowForDraw("Test")
	// Child fills the entire window buffer — it would overwrite the border if
	// not clipped to the client SubBuffer.
	bigChild := &drawOrderMockView{id: 'F', bounds: NewRect(0, 0, 20, 10)}
	bigChild.SetState(SfVisible, true)
	w.Insert(bigChild)

	w.Draw(buf)

	// The top-left border corner should remain a frame character.
	cell := buf.GetCell(0, 0)
	if cell.Rune == 'F' {
		t.Errorf("Child wrote over border at (0,0): cell = %q, want frame char", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Draw: no ColorScheme fallback
// ---------------------------------------------------------------------------

// TestDrawNoColorSchemeFallsBackToStyleDefault verifies that when no ColorScheme
// is available, Draw uses tcell.StyleDefault for all styles without panicking.
// Spec: "When no ColorScheme is available, Draw uses tcell.StyleDefault for all styles"
func TestDrawNoColorSchemeFallsBackToStyleDefault(t *testing.T) {
	// No scheme attached, no owner → ColorScheme() returns nil.
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	buf := NewDrawBuffer(20, 10)

	// Must not panic.
	w.Draw(buf)

	// All rendered cells should use StyleDefault.
	cell := buf.GetCell(0, 0)
	if cell.Style != tcell.StyleDefault {
		t.Errorf("No ColorScheme: cell(0,0).Style = %v, want StyleDefault", cell.Style)
	}
}

// TestDrawNoColorSchemeFallsBackToStyleDefaultForBackground verifies the client
// area also uses StyleDefault when there is no scheme.
// Spec: "When no ColorScheme is available, Draw uses tcell.StyleDefault for all styles"
func TestDrawNoColorSchemeFallsBackToStyleDefaultForBackground(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 20, 10), "Test")
	buf := NewDrawBuffer(20, 10)

	w.Draw(buf)

	cell := buf.GetCell(1, 1)
	if cell.Style != tcell.StyleDefault {
		t.Errorf("No ColorScheme client area: cell(1,1).Style = %v, want StyleDefault", cell.Style)
	}
}

// ---------------------------------------------------------------------------
// HandleEvent delegation
// ---------------------------------------------------------------------------

// TestHandleEventDelegatesToGroup verifies HandleEvent forwards the event to
// the internal group (which routes it to the focused child).
// Spec: "Window.HandleEvent(event) delegates to the internal Group"
func TestHandleEventDelegatesToGroup(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 10, 5))
	w.Insert(child)
	event := &Event{What: EvKeyboard}

	w.HandleEvent(event)

	if child.eventHandled != event {
		t.Errorf("HandleEvent: focused child did not receive event")
	}
}

// TestHandleEventCommandDelegatesToGroup verifies command events are also
// forwarded via the group.
// Spec: "Window.HandleEvent(event) delegates to the internal Group"
func TestHandleEventCommandDelegatesToGroup(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 10, 5))
	w.Insert(child)
	event := &Event{What: EvCommand, Command: CmClose}

	w.HandleEvent(event)

	if child.eventHandled != event {
		t.Errorf("HandleEvent command: focused child did not receive event")
	}
}

// TestHandleEventNoChildDoesNotPanic verifies HandleEvent is safe with no children.
// Spec: "Window.HandleEvent(event) delegates to the internal Group" — must not panic.
func TestHandleEventNoChildDoesNotPanic(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	event := &Event{What: EvKeyboard}

	// Must not panic.
	w.HandleEvent(event)
}

func TestResizeBottomRightDoesNotShrinkOnInitialClick(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(10, 5, 35, 16), "Resize Test")
	d.Insert(w)
	origBounds := w.Bounds()

	// Click the bottom-right corner in Desktop-local space: (10+34, 5+15) = (44, 20)
	d.HandleEvent(mouseEvent(44, 20, tcell.Button1))
	if !w.resizing {
		t.Fatal("resize should have started")
	}

	// Mouse stays at the same Desktop-local position (stationary click)
	d.HandleEvent(mouseEvent(44, 20, tcell.Button1))

	newBounds := w.Bounds()
	if newBounds.Width() != origBounds.Width() || newBounds.Height() != origBounds.Height() {
		t.Errorf("window resized on stationary click: %dx%d → %dx%d",
			origBounds.Width(), origBounds.Height(),
			newBounds.Width(), newBounds.Height())
	}
}

func TestResizeEditWindowDoesNotShrinkOnStationary(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	ew := NewEditWindow(NewRect(45, 1, 35, 16), "")
	d.Insert(ew)
	origBounds := ew.Bounds()

	// Bottom-right corner at Desktop-local (45+34, 1+15) = (79, 16)
	brx := origBounds.B.X - 1
	bry := origBounds.B.Y - 1
	d.HandleEvent(mouseEvent(brx, bry, tcell.Button1))

	w := ew.Window
	if !w.resizing {
		t.Fatal("resize should have started on EditWindow's embedded Window")
	}

	// Mouse stays at same position
	d.HandleEvent(mouseEvent(brx, bry, tcell.Button1))

	newBounds := ew.Bounds()
	if newBounds.Width() != origBounds.Width() || newBounds.Height() != origBounds.Height() {
		t.Errorf("EditWindow resized on stationary click: %dx%d → %dx%d",
			origBounds.Width(), origBounds.Height(),
			newBounds.Width(), newBounds.Height())
	}
}

func TestResizeBottomRightGrowsByDelta(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 24))
	w := NewWindow(NewRect(10, 5, 35, 16), "Resize Test")
	d.Insert(w)

	// Bottom-right corner at Desktop-local (10+34, 5+15) = (44, 20)
	d.HandleEvent(mouseEvent(44, 20, tcell.Button1))
	if !w.resizing {
		t.Fatal("resize should have started")
	}

	// Drag 5 right and 3 down: to (49, 23)
	d.HandleEvent(mouseEvent(49, 23, tcell.Button1))

	newBounds := w.Bounds()
	wantW := 49 - 10 + 1 // mx - bounds.A.X + 1 = 40
	wantH := 23 - 5 + 1  // my - bounds.A.Y + 1 = 19
	if newBounds.Width() != wantW {
		t.Errorf("width after drag: got %d, want %d", newBounds.Width(), wantW)
	}
	if newBounds.Height() != wantH {
		t.Errorf("height after drag: got %d, want %d", newBounds.Height(), wantH)
	}
}
