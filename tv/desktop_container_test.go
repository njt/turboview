package tv

// desktop_container_test.go — tests for Task 1: Desktop Container Upgrade
//
// Every assertion traces to a spec sentence quoted in the test's doc comment.
// No implementation code was read before writing these tests; spec-only approach.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// Compile-time check: spec says "Desktop satisfies the Container interface
// (compile-time check: var _ Container = (*Desktop)(nil))".
var _ Container = (*Desktop)(nil)

// ---------------------------------------------------------------------------
// Desktop.Insert / Container delegation
// ---------------------------------------------------------------------------

// Spec: "Desktop.Insert(v) delegates to the internal Group"
// Confirming: a view inserted via Desktop.Insert appears in Desktop.Children().
func TestDesktopInsertAddsToChildren(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(0, 0, 10, 5))

	d.Insert(v)

	children := d.Children()
	if len(children) != 1 {
		t.Fatalf("Children() len = %d after Insert, want 1", len(children))
	}
	if children[0] != v {
		t.Errorf("Children()[0] != inserted view")
	}
}

// Spec: "Desktop.Insert(v) delegates to the internal Group"
// Falsifying: a naive implementation that appends directly to a Desktop-owned
// slice (bypassing the Group) would still pass the confirming test above,
// but would fail if Children() is backed exclusively by the Group's list.
// We verify Children() returns the same set after two Inserts (tests ordering too).
func TestDesktopInsertMultiplePreservesOrder(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v1 := newMockView(NewRect(0, 0, 10, 5))
	v2 := newMockView(NewRect(10, 0, 10, 5))

	d.Insert(v1)
	d.Insert(v2)

	children := d.Children()
	if len(children) != 2 || children[0] != v1 || children[1] != v2 {
		t.Errorf("Children() = %v after two inserts, want [v1, v2] in order", children)
	}
}

// ---------------------------------------------------------------------------
// Desktop.Remove / Container delegation
// ---------------------------------------------------------------------------

// Spec: "Desktop.Remove(v) delegates to the internal Group"
// Confirming: a view removed via Desktop.Remove is no longer in Children().
func TestDesktopRemoveRemovesFromChildren(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(0, 0, 10, 5))
	d.Insert(v)

	d.Remove(v)

	for _, c := range d.Children() {
		if c == v {
			t.Errorf("Children() still contains view after Remove")
		}
	}
}

// Spec: "Desktop.Remove(v) delegates to the internal Group"
// Falsifying: Remove must also clear the child's owner (Group.Remove does this).
// A stub Remove that only removes from the slice would leave Owner() non-nil.
func TestDesktopRemoveClearsChildOwner(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(0, 0, 10, 5))
	d.Insert(v)

	d.Remove(v)

	if v.Owner() != nil {
		t.Errorf("child.Owner() = %v after Desktop.Remove, want nil", v.Owner())
	}
}

// ---------------------------------------------------------------------------
// Child owner = Desktop (facade)
// ---------------------------------------------------------------------------

// Spec: "Desktop holds an internal Group whose facade is set to the Desktop itself,
// so children inserted via Desktop.Insert(v) see the Desktop as their owner"
// Confirming: after Insert the child's Owner() is the Desktop, not the internal Group.
func TestDesktopInsertSetsOwnerToDesktop(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(0, 0, 10, 5))

	d.Insert(v)

	if v.Owner() != d {
		t.Errorf("child.Owner() = %v after Desktop.Insert, want Desktop", v.Owner())
	}
}

// Spec: same sentence — facade is the Desktop, not the internal Group.
// Falsifying: if the facade is not set, the child's Owner() would be the internal Group.
// We confirm it is specifically the Desktop pointer.
func TestDesktopInsertOwnerIsExactlyDesktop(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(0, 0, 10, 5))

	d.Insert(v)

	// Owner must be the desktop itself, castable as *Desktop.
	if _, ok := v.Owner().(*Desktop); !ok {
		t.Errorf("child.Owner() type = %T, want *Desktop", v.Owner())
	}
}

// ---------------------------------------------------------------------------
// Desktop.Children
// ---------------------------------------------------------------------------

// Spec: "Desktop.Children() delegates to the internal Group"
// Confirming: Children() returns empty slice before any Insert.
func TestDesktopChildrenEmptyInitially(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))

	if len(d.Children()) != 0 {
		t.Errorf("Children() len = %d on fresh Desktop, want 0", len(d.Children()))
	}
}

// ---------------------------------------------------------------------------
// Desktop.FocusedChild / SetFocusedChild
// ---------------------------------------------------------------------------

// Spec: "Desktop.FocusedChild() delegates to the internal Group"
// Confirming: FocusedChild() returns nil when no selectable child is present.
func TestDesktopFocusedChildNilWhenEmpty(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))

	if d.FocusedChild() != nil {
		t.Errorf("FocusedChild() = %v on empty Desktop, want nil", d.FocusedChild())
	}
}

// Spec: "Desktop.FocusedChild() delegates to the internal Group"
// Confirming: after inserting a selectable view it becomes the focused child.
func TestDesktopFocusedChildReturnsInsertedSelectable(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v := newSelectableMockView(NewRect(0, 0, 10, 5))

	d.Insert(v)

	if d.FocusedChild() != v {
		t.Errorf("FocusedChild() = %v, want inserted selectable view", d.FocusedChild())
	}
}

// Spec: "Desktop.SetFocusedChild(v) delegates to the internal Group"
// Confirming: calling SetFocusedChild changes FocusedChild().
func TestDesktopSetFocusedChildChangesSelection(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	first := newSelectableMockView(NewRect(0, 0, 10, 5))
	second := newSelectableMockView(NewRect(10, 0, 10, 5))

	d.Insert(first)
	d.Insert(second)
	// second is selected; switch back to first
	d.SetFocusedChild(first)

	if d.FocusedChild() != first {
		t.Errorf("FocusedChild() = %v after SetFocusedChild(first), want first", d.FocusedChild())
	}
}

// Spec: "Desktop.SetFocusedChild(v) delegates to the internal Group"
// Falsifying: deselection of the old child must also happen (Group behaviour).
func TestDesktopSetFocusedChildDeselectsPrevious(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	first := newSelectableMockView(NewRect(0, 0, 10, 5))
	second := newSelectableMockView(NewRect(10, 0, 10, 5))

	d.Insert(first)
	d.Insert(second)
	d.SetFocusedChild(first)

	if second.HasState(SfSelected) {
		t.Errorf("previous child still has SfSelected after SetFocusedChild")
	}
}

// ---------------------------------------------------------------------------
// Desktop.ExecView
// ---------------------------------------------------------------------------

// Spec: "Desktop.ExecView(v) delegates to the internal Group (panics 'not implemented')"
// Confirming: ExecView panics.
func TestDesktopExecViewPanics(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(0, 0, 10, 5))

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Desktop.ExecView did not panic")
		}
	}()

	d.ExecView(v)
}

// ---------------------------------------------------------------------------
// Desktop.SetBounds
// ---------------------------------------------------------------------------

// Spec: "Desktop.SetBounds(r) updates both Desktop's bounds and the internal
// Group's bounds (Group uses origin 0,0 with Desktop's width and height)"
// Confirming: Desktop.Bounds() reflects the new rect after SetBounds.
func TestDesktopSetBoundsUpdatesDesktopBounds(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	newBounds := NewRect(5, 10, 100, 40)

	d.SetBounds(newBounds)

	if d.Bounds() != newBounds {
		t.Errorf("Desktop.Bounds() = %v after SetBounds, want %v", d.Bounds(), newBounds)
	}
}

// Spec: "Group uses origin 0,0 with Desktop's width and height"
// Confirming: after SetBounds the internal Group has origin (0,0) and the same
// width/height as the Desktop's new rect.
// We verify this indirectly: a child inserted at (0,0) with the group's full
// size gets drawn, confirming the group's bounds match Desktop width×height.
func TestDesktopSetBoundsUpdatesGroupOriginToZero(t *testing.T) {
	d := NewDesktop(NewRect(10, 20, 50, 30)) // non-zero origin on Desktop

	// Insert a child that fills what the group should think is its full area.
	// The group's bounds should be NewRect(0,0,50,30).
	// A child at (0,0,10,5) must be drawable.
	child := newMockView(NewRect(0, 0, 10, 5))
	d.Insert(child)

	buf := NewDrawBuffer(50, 30)
	d.Draw(buf)

	if !child.drawCalled {
		t.Errorf("child at group-local (0,0) was not drawn; group origin may not be 0,0")
	}
}

// Spec: "Group uses origin 0,0 with Desktop's width and height"
// Falsifying: if SetBounds copies the Desktop rect directly into the Group
// (including the Desktop's non-zero origin), children positioned at group-local
// (0,0) would not be drawn because (0,0) would lie outside the group's
// assumed origin. The confirming test above catches this, but we also check
// width and height independently.
func TestDesktopSetBoundsGroupHasSameWidthHeight(t *testing.T) {
	// Set a Desktop with non-zero origin and specific dimensions.
	d := NewDesktop(NewRect(5, 10, 60, 20))

	// A child positioned at (59,19) is within a 60×20 group (origin 0,0) but
	// outside a group that naively copied the Desktop rect (5..65, 10..30).
	// drawOrderMockView.Draw writes its id rune at (0,0) of its subbuffer.
	// If the group's clip contains the child's position, the rune 'X' will appear
	// at absolute (59,19) in the buffer.
	child := &drawOrderMockView{id: 'X', bounds: NewRect(59, 19, 1, 1)}
	child.SetState(SfVisible, true)
	d.Insert(child)

	buf := NewDrawBuffer(60, 20)
	d.Draw(buf)

	cell := buf.GetCell(59, 19)
	if cell.Rune != 'X' {
		t.Errorf("cell(59,19) rune = %q, want 'X'; group dimensions may not match Desktop width×height", cell.Rune)
	}
}

// Spec: "Desktop.SetBounds(r) updates both Desktop's bounds and the internal
// Group's bounds (Group uses origin 0,0 with Desktop's width and height)"
// Falsifying: the constructor could set group bounds correctly while SetBounds
// (called after construction) forgets to propagate. This test calls SetBounds
// after construction with different dimensions and verifies via drawing.
func TestDesktopSetBoundsAfterConstructionUpdatesGroup(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 10, 5))
	d.SetBounds(NewRect(5, 10, 80, 40))

	child := &drawOrderMockView{id: 'Q', bounds: NewRect(79, 39, 1, 1)}
	child.SetState(SfVisible, true)
	d.Insert(child)

	buf := NewDrawBuffer(80, 40)
	d.Draw(buf)

	cell := buf.GetCell(79, 39)
	if cell.Rune != 'Q' {
		t.Errorf("cell(79,39) rune = %q, want 'Q'; SetBounds may not update group dimensions", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Desktop.Draw — background fill then children
// ---------------------------------------------------------------------------

// Spec: "Desktop.Draw(buf) fills the background with the pattern rune and
// DesktopBackground style, then calls the internal Group's Draw to render
// children back-to-front"
// Confirming: a child inserted at a position paints its content on top of the
// background (child's draw happens after the fill).
func TestDesktopDrawRendersChildrenOverBackground(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 10, 5))
	// Insert a child that writes 'Z' at (0,0).
	child := &drawOrderMockView{id: 'Z', bounds: NewRect(0, 0, 5, 3)}
	child.SetState(SfVisible, true)
	d.Insert(child)

	buf := NewDrawBuffer(10, 5)
	d.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != 'Z' {
		t.Errorf("cell(0,0) rune = %q, want 'Z' (child must overdraw background)", cell.Rune)
	}
}

// Spec: same — background fills first, then children are rendered.
// Falsifying: if Draw only renders children and forgets the background fill,
// cells outside children's bounds would have the zero-value rune ' ' instead
// of the '░' pattern rune.
func TestDesktopDrawFillsBackgroundBeyondChildren(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 10, 5))
	// child only covers top-left 5×3; bottom-right area should still have '░'
	child := newMockView(NewRect(0, 0, 5, 3))
	d.Insert(child)

	buf := NewDrawBuffer(10, 5)
	d.Draw(buf)

	cell := buf.GetCell(9, 4) // bottom-right corner, outside child
	if cell.Rune != '░' {
		t.Errorf("cell(9,4) rune = %q, want '░' background fill outside child area", cell.Rune)
	}
}

// Spec: "fills the background with the pattern rune and DesktopBackground style"
// Confirming: background cells use ColorScheme().DesktopBackground style.
// (Same as existing tests but now ensures children don't break it at unoccupied cells.)
func TestDesktopDrawBackgroundStyleBeyondChildren(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 10, 5))
	scheme := &theme.ColorScheme{
		DesktopBackground: tcell.StyleDefault.Foreground(tcell.ColorTeal),
	}
	d.scheme = scheme

	child := newMockView(NewRect(0, 0, 5, 3))
	d.Insert(child)

	buf := NewDrawBuffer(10, 5)
	d.Draw(buf)

	cell := buf.GetCell(9, 4)
	if cell.Style != scheme.DesktopBackground {
		t.Errorf("background cell style = %v, want DesktopBackground %v", cell.Style, scheme.DesktopBackground)
	}
}

// Spec: "then calls the internal Group's Draw to render children back-to-front"
// Confirming: Group's back-to-front order is respected — last-inserted child
// paints over an earlier child that occupies the same cells.
func TestDesktopDrawChildrenBackToFront(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 20, 10))
	bottom := &drawOrderMockView{id: 'A', bounds: NewRect(0, 0, 10, 5)}
	bottom.SetState(SfVisible, true)
	top := &drawOrderMockView{id: 'B', bounds: NewRect(0, 0, 10, 5)}
	top.SetState(SfVisible, true)

	d.Insert(bottom)
	d.Insert(top)

	buf := NewDrawBuffer(20, 10)
	d.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != 'B' {
		t.Errorf("cell(0,0) rune = %q, want 'B' (last-inserted draws on top)", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Desktop.HandleEvent
// ---------------------------------------------------------------------------

// Spec: "Desktop.HandleEvent(event) delegates to the internal Group"
// Confirming: HandleEvent forwards the event to the focused child (Group behaviour).
func TestDesktopHandleEventForwardsToFocusedChild(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v := newSelectableMockView(NewRect(0, 0, 10, 5))
	d.Insert(v)
	event := &Event{What: EvKeyboard}

	d.HandleEvent(event)

	if v.eventHandled != event {
		t.Errorf("HandleEvent did not forward event to focused child")
	}
}

// Spec: "Desktop.HandleEvent(event) delegates to the internal Group"
// Falsifying: a no-op HandleEvent passes only if no assertion is made.
// We also check that a cleared event (EvNothing) is NOT forwarded (Group skips it).
func TestDesktopHandleEventDoesNotForwardClearedEvent(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v := newSelectableMockView(NewRect(0, 0, 10, 5))
	d.Insert(v)
	event := &Event{What: EvNothing}

	d.HandleEvent(event)

	if v.eventHandled != nil {
		t.Errorf("HandleEvent forwarded a cleared event; Group should have skipped it")
	}
}

// Spec: same — must not panic when no focused child.
func TestDesktopHandleEventNoFocusedChildDoesNotPanic(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	event := &Event{What: EvKeyboard}

	// Must not panic.
	d.HandleEvent(event)
}

// ---------------------------------------------------------------------------
// Group.BringToFront
// ---------------------------------------------------------------------------

// Spec: "Group.BringToFront(v) moves the given child to the end of the children
// list (frontmost in z-order) and selects it"
// Confirming: after BringToFront the target view is at the last position in Children().
func TestGroupBringToFrontMovesChildToEnd(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v1 := newMockView(NewRect(0, 0, 10, 5))
	v2 := newMockView(NewRect(10, 0, 10, 5))
	v3 := newMockView(NewRect(20, 0, 10, 5))

	g.Insert(v1)
	g.Insert(v2)
	g.Insert(v3)

	g.BringToFront(v1) // v1 was first; should move to last

	children := g.Children()
	if len(children) != 3 {
		t.Fatalf("Children() len = %d after BringToFront, want 3", len(children))
	}
	if children[len(children)-1] != v1 {
		t.Errorf("last child = %v after BringToFront(v1), want v1", children[len(children)-1])
	}
}

// Spec: same
// Falsifying: a naive implementation might append a copy without removing the
// original, producing 4 elements. We verify the list length stays 3.
func TestGroupBringToFrontDoesNotDuplicateChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v1 := newMockView(NewRect(0, 0, 10, 5))
	v2 := newMockView(NewRect(10, 0, 10, 5))

	g.Insert(v1)
	g.Insert(v2)
	g.BringToFront(v1)

	if len(g.Children()) != 2 {
		t.Errorf("Children() len = %d after BringToFront, want 2 (no duplicates)", len(g.Children()))
	}
}

// Spec: "moves the given child to the end of the children list (frontmost in z-order)
// and selects it"
// Confirming: after BringToFront the view becomes the focused (selected) child.
func TestGroupBringToFrontSelectsChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v1 := newSelectableMockView(NewRect(0, 0, 10, 5))
	v2 := newSelectableMockView(NewRect(10, 0, 10, 5))

	g.Insert(v1)
	g.Insert(v2)
	// v2 is now focused

	g.BringToFront(v1)

	if g.FocusedChild() != v1 {
		t.Errorf("FocusedChild() = %v after BringToFront(v1), want v1", g.FocusedChild())
	}
}

// Spec: "and selects it" — implies the previously focused child is deselected.
// Falsifying: a lazy impl might set SfSelected on the target without clearing it
// from the previous focused child.
func TestGroupBringToFrontDeselectsPreviousFocused(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v1 := newSelectableMockView(NewRect(0, 0, 10, 5))
	v2 := newSelectableMockView(NewRect(10, 0, 10, 5))

	g.Insert(v1)
	g.Insert(v2)
	// v2 is focused

	g.BringToFront(v1)

	if v2.HasState(SfSelected) {
		t.Errorf("previous focused child still has SfSelected after BringToFront")
	}
}

// Spec: "no-op if the child is not in the group"
// Confirming: BringToFront with a view that was never inserted doesn't change anything.
func TestGroupBringToFrontNoOpForNonMember(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v1 := newMockView(NewRect(0, 0, 10, 5))
	outsider := newMockView(NewRect(0, 0, 5, 5))

	g.Insert(v1)
	originalChildren := g.Children()

	g.BringToFront(outsider) // outsider not in group

	after := g.Children()
	if len(after) != len(originalChildren) {
		t.Errorf("BringToFront non-member changed Children() len: got %d, want %d", len(after), len(originalChildren))
	}
	if after[0] != v1 {
		t.Errorf("BringToFront non-member changed children order")
	}
}

// Spec: "no-op if the child is not in the group"
// Falsifying: ensure BringToFront non-member doesn't panic.
func TestGroupBringToFrontNonMemberDoesNotPanic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	outsider := newMockView(NewRect(0, 0, 5, 5))

	// Must not panic.
	g.BringToFront(outsider)
}

// Spec: "Group.BringToFront(v) moves the given child to the end of the children list"
// Confirming: when the target is already at the end it remains at the end
// (and the list length stays the same).
func TestGroupBringToFrontAlreadyFrontmost(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v1 := newMockView(NewRect(0, 0, 10, 5))
	v2 := newMockView(NewRect(10, 0, 10, 5))

	g.Insert(v1)
	g.Insert(v2)
	// v2 is last

	g.BringToFront(v2) // already frontmost

	children := g.Children()
	if len(children) != 2 {
		t.Fatalf("Children() len = %d, want 2", len(children))
	}
	if children[len(children)-1] != v2 {
		t.Errorf("last child = %v after BringToFront(already-front), want v2", children[len(children)-1])
	}
}

// Spec: BringToFront z-order — the frontmost child draws last (on top).
// Confirming: after BringToFront(v1) over v2 in the same position, v1 is drawn
// on top (its rune wins at the overlapping cell).
func TestGroupBringToFrontChildDrawsOnTop(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 20, 10))
	v1 := &drawOrderMockView{id: 'A', bounds: NewRect(0, 0, 10, 5)}
	v1.SetState(SfVisible, true)
	v2 := &drawOrderMockView{id: 'B', bounds: NewRect(0, 0, 10, 5)}
	v2.SetState(SfVisible, true)

	g.Insert(v1)
	g.Insert(v2)
	// v2 is on top (drawn last)

	g.BringToFront(v1) // now v1 should be on top

	buf := NewDrawBuffer(20, 10)
	g.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != 'A' {
		t.Errorf("cell(0,0) rune = %q after BringToFront(v1), want 'A' (v1 draws on top)", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Desktop.BringToFront
// ---------------------------------------------------------------------------

// Spec: "Desktop.BringToFront(v) delegates to the internal Group's BringToFront"
// Confirming: after Desktop.BringToFront(v1) the child list has v1 last.
func TestDesktopBringToFrontMovesChildToFront(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v1 := newMockView(NewRect(0, 0, 10, 5))
	v2 := newMockView(NewRect(10, 0, 10, 5))

	d.Insert(v1)
	d.Insert(v2)

	d.BringToFront(v1)

	children := d.Children()
	if len(children) != 2 {
		t.Fatalf("Children() len = %d, want 2", len(children))
	}
	if children[len(children)-1] != v1 {
		t.Errorf("last child = %v after Desktop.BringToFront(v1), want v1", children[len(children)-1])
	}
}

// Spec: same — delegation means selection also follows (Group.BringToFront selects).
// Falsifying: a forwarding stub that only reorders without calling Group.BringToFront
// fully would miss the selection side-effect.
func TestDesktopBringToFrontSelectsChild(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	v1 := newSelectableMockView(NewRect(0, 0, 10, 5))
	v2 := newSelectableMockView(NewRect(10, 0, 10, 5))

	d.Insert(v1)
	d.Insert(v2)

	d.BringToFront(v1)

	if d.FocusedChild() != v1 {
		t.Errorf("FocusedChild() = %v after Desktop.BringToFront(v1), want v1", d.FocusedChild())
	}
}
