package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// mockView is a test double for View that tracks Draw and HandleEvent calls.
type mockView struct {
	BaseView
	drawCalled   bool
	drawBuf      *DrawBuffer
	eventHandled *Event
}

func (m *mockView) Draw(buf *DrawBuffer) {
	m.drawCalled = true
	m.drawBuf = buf
}

func (m *mockView) HandleEvent(event *Event) {
	m.eventHandled = event
}

// mockFacadeContainer is a test double for Container used as a facade.
type mockFacadeContainer struct {
	BaseView
}

func (m *mockFacadeContainer) Insert(child View)            {}
func (m *mockFacadeContainer) Remove(child View)            {}
func (m *mockFacadeContainer) Children() []View             { return nil }
func (m *mockFacadeContainer) FocusedChild() View           { return nil }
func (m *mockFacadeContainer) SetFocusedChild(child View)   {}
func (m *mockFacadeContainer) ExecView(v View) CommandCode  { return CmCancel }

// newMockView creates a visible mock view with the given bounds.
func newMockView(bounds Rect) *mockView {
	v := &mockView{}
	v.SetBounds(bounds)
	v.SetState(SfVisible, true)
	return v
}

// newSelectableMockView creates a visible, selectable mock view with the given bounds.
func newSelectableMockView(bounds Rect) *mockView {
	v := newMockView(bounds)
	v.SetOptions(OfSelectable, true)
	return v
}

// TestNewGroupSetsBounds verifies that NewGroup stores the given bounds.
// Spec: "NewGroup(bounds) creates a Group with the given bounds".
func TestNewGroupSetsBounds(t *testing.T) {
	r := NewRect(5, 10, 40, 20)
	g := NewGroup(r)

	if g.Bounds() != r {
		t.Errorf("NewGroup bounds = %v, want %v", g.Bounds(), r)
	}
}

// TestNewGroupSetsSfVisible verifies that NewGroup sets SfVisible state.
// Spec: "NewGroup(bounds) creates a Group with the given bounds, SfVisible state set".
func TestNewGroupSetsSfVisible(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	if !g.HasState(SfVisible) {
		t.Errorf("NewGroup did not set SfVisible")
	}
}

// TestInsertAddsChildToChildrenList verifies Insert appends the child to the group's child list.
// Spec: "Insert(view) adds a child to the end of the child list".
func TestInsertAddsChildToChildrenList(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(0, 0, 10, 5))

	g.Insert(v)

	children := g.Children()
	if len(children) != 1 {
		t.Fatalf("Children() len = %d, want 1", len(children))
	}
	if children[0] != v {
		t.Errorf("Children()[0] = %v, want %v", children[0], v)
	}
}

// TestInsertSetsChildOwnerToGroup verifies Insert sets the child's owner to the group.
// Spec: "Insert(view) adds a child to the end of the child list and sets the child's owner
// to this Group (or the facade if set)".
func TestInsertSetsChildOwnerToGroup(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(0, 0, 10, 5))

	g.Insert(v)

	if v.Owner() != g {
		t.Errorf("child.Owner() = %v, want group", v.Owner())
	}
}

// TestInsertWithFacadeSetsChildOwnerToFacade verifies that when a facade is set,
// inserted children see the facade as their owner.
// Spec: "when set, inserted children see the facade as their owner instead of the Group".
func TestInsertWithFacadeSetsChildOwnerToFacade(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	facade := &mockFacadeContainer{}
	g.SetFacade(facade)
	v := newMockView(NewRect(0, 0, 10, 5))

	g.Insert(v)

	if v.Owner() != facade {
		t.Errorf("child.Owner() with facade = %v, want facade", v.Owner())
	}
}

// TestInsertSelectableChildSelectsIt verifies that inserting an OfSelectable view
// sets SfSelected on that child.
// Spec: "Insert(view) on an OfSelectable view selects it (sets SfSelected)".
func TestInsertSelectableChildSelectsIt(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := newSelectableMockView(NewRect(0, 0, 10, 5))

	g.Insert(v)

	if !v.HasState(SfSelected) {
		t.Errorf("selectable child after Insert does not have SfSelected")
	}
}

// TestInsertSelectableChildDeselectionsPrevious verifies that inserting a second
// selectable view deselects the previously selected child.
// Spec: "Insert(view) on an OfSelectable view selects it (sets SfSelected)
// and deselects the previous selection".
func TestInsertSelectableChildDeselectsPreviousSelection(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableMockView(NewRect(0, 0, 10, 5))
	second := newSelectableMockView(NewRect(10, 0, 10, 5))

	g.Insert(first)
	g.Insert(second)

	if first.HasState(SfSelected) {
		t.Errorf("first selectable child still has SfSelected after inserting second")
	}
	if !second.HasState(SfSelected) {
		t.Errorf("second selectable child does not have SfSelected")
	}
}

// TestInsertNonSelectableChildDoesNotChangeFocused verifies inserting a non-selectable
// child does not alter the current selection.
// Spec: "Insert(view) on an OfSelectable view selects it" — by contrapositive, a
// non-selectable insert must not affect the focused child.
func TestInsertNonSelectableChildDoesNotChangeFocused(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	selectable := newSelectableMockView(NewRect(0, 0, 10, 5))
	nonSelectable := newMockView(NewRect(10, 0, 10, 5))

	g.Insert(selectable)
	g.Insert(nonSelectable)

	if !selectable.HasState(SfSelected) {
		t.Errorf("selectable child lost SfSelected after inserting non-selectable child")
	}
}

// TestInsertMultipleMaintainsInsertionOrder verifies that multiple inserts preserve order.
// Spec: "Insert(view) adds a child to the end of the child list".
func TestInsertMultipleMaintainsInsertionOrder(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v1 := newMockView(NewRect(0, 0, 5, 5))
	v2 := newMockView(NewRect(5, 0, 5, 5))
	v3 := newMockView(NewRect(10, 0, 5, 5))

	g.Insert(v1)
	g.Insert(v2)
	g.Insert(v3)

	children := g.Children()
	if len(children) != 3 {
		t.Fatalf("Children() len = %d, want 3", len(children))
	}
	if children[0] != v1 || children[1] != v2 || children[2] != v3 {
		t.Errorf("Children() order = [%v, %v, %v], want [v1, v2, v3]", children[0], children[1], children[2])
	}
}

// TestRemoveRemovesChildFromList verifies Remove takes the child out of the children list.
// Spec: "Remove(view) removes the child and clears its owner".
func TestRemoveRemovesChildFromList(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(0, 0, 10, 5))
	g.Insert(v)

	g.Remove(v)

	children := g.Children()
	for _, c := range children {
		if c == v {
			t.Errorf("Children() still contains removed view")
		}
	}
}

// TestRemoveClearsChildOwner verifies Remove clears the child's owner.
// Spec: "Remove(view) removes the child and clears its owner".
func TestRemoveClearsChildOwner(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(0, 0, 10, 5))
	g.Insert(v)

	g.Remove(v)

	if v.Owner() != nil {
		t.Errorf("removed child.Owner() = %v, want nil", v.Owner())
	}
}

// TestRemoveFocusedChildSelectsPrevious verifies that removing the focused child
// selects the previous selectable child.
// Spec: "if it was focused, selects the previous child".
func TestRemoveFocusedChildSelectsPreviousSelectable(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	prev := newSelectableMockView(NewRect(0, 0, 10, 5))
	focused := newSelectableMockView(NewRect(10, 0, 10, 5))

	g.Insert(prev)
	g.Insert(focused)
	// focused is now selected (last inserted selectable)

	g.Remove(focused)

	if !prev.HasState(SfSelected) {
		t.Errorf("previous selectable child does not have SfSelected after removing focused child")
	}
}

// TestRemoveNonFocusedChildDoesNotChangeFocused verifies that removing a non-focused
// child leaves the current selection intact.
// Spec: "if it was focused, selects the previous child" — a non-focused removal must
// not disturb the focused child.
func TestRemoveNonFocusedChildDoesNotChangeFocused(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	focused := newSelectableMockView(NewRect(0, 0, 10, 5))
	other := newMockView(NewRect(10, 0, 10, 5))

	g.Insert(focused)
	g.Insert(other)

	g.Remove(other)

	if !focused.HasState(SfSelected) {
		t.Errorf("focused child lost SfSelected after removing an unrelated child")
	}
}

// TestChildrenReturnsInsertionOrder verifies Children() returns children in insertion order.
// Spec: "Children() returns the child list in insertion order".
func TestChildrenReturnsInsertionOrder(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v1 := newMockView(NewRect(0, 0, 5, 5))
	v2 := newMockView(NewRect(5, 0, 5, 5))

	g.Insert(v1)
	g.Insert(v2)

	children := g.Children()
	if len(children) < 2 || children[0] != v1 || children[1] != v2 {
		t.Errorf("Children() = %v, want [v1, v2] in insertion order", children)
	}
}

// TestChildrenEmptyGroupReturnsEmpty verifies Children() returns an empty or nil slice
// for a group with no children.
// Spec: "Children() returns the child list in insertion order" — with zero children
// the list is empty.
func TestChildrenEmptyGroupReturnsEmpty(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	children := g.Children()

	if len(children) != 0 {
		t.Errorf("Children() on empty group len = %d, want 0", len(children))
	}
}

// TestFocusedChildNilWhenNoChildren verifies FocusedChild returns nil for an empty group.
// Spec: "FocusedChild() returns the child with SfSelected state, or nil".
func TestFocusedChildNilWhenNoChildren(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	if g.FocusedChild() != nil {
		t.Errorf("FocusedChild() on empty group = %v, want nil", g.FocusedChild())
	}
}

// TestFocusedChildReturnsSelectedChild verifies FocusedChild returns the child
// that has SfSelected.
// Spec: "FocusedChild() returns the child with SfSelected state, or nil".
func TestFocusedChildReturnsSelectedChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := newSelectableMockView(NewRect(0, 0, 10, 5))
	g.Insert(v)

	got := g.FocusedChild()

	if got != v {
		t.Errorf("FocusedChild() = %v, want inserted selectable child", got)
	}
}

// TestSetFocusedChildSelectsNewAndDeselectsOld verifies SetFocusedChild transitions
// the selection from the current child to the given child.
// Spec: "SetFocusedChild(view) deselects the current selection and selects the given view".
func TestSetFocusedChildSelectsNewAndDeselectsOld(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	first := newSelectableMockView(NewRect(0, 0, 10, 5))
	second := newSelectableMockView(NewRect(10, 0, 10, 5))

	g.Insert(first)
	g.Insert(second)
	// second is focused after two inserts; switch back to first
	g.SetFocusedChild(first)

	if !first.HasState(SfSelected) {
		t.Errorf("SetFocusedChild: new child does not have SfSelected")
	}
	if second.HasState(SfSelected) {
		t.Errorf("SetFocusedChild: old child still has SfSelected")
	}
}

// TestDrawCallsDrawOnVisibleChildren verifies Draw invokes Draw on each visible child.
// Spec: "Draw(buf) iterates children back-to-front; for each visible child, creates a
// SubBuffer clipped to the child's bounds and calls child.Draw(sub)".
func TestDrawCallsDrawOnVisibleChildren(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(0, 0, 10, 5))
	g.Insert(v)
	buf := NewDrawBuffer(80, 25)

	g.Draw(buf)

	if !v.drawCalled {
		t.Errorf("Draw did not call Draw on visible child")
	}
}

// TestDrawPassesSubBuffer verifies Draw passes a SubBuffer (not the full buffer)
// to each visible child.
// Spec: "for each visible child, creates a SubBuffer clipped to the child's bounds
// and calls child.Draw(sub)".
func TestDrawPassesSubBuffer(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := newMockView(NewRect(5, 3, 10, 5))
	g.Insert(v)
	buf := NewDrawBuffer(80, 25)

	g.Draw(buf)

	if v.drawBuf == nil {
		t.Fatal("Draw passed nil buffer to child")
	}
	if v.drawBuf == buf {
		t.Errorf("Draw passed the full buffer instead of a SubBuffer to child")
	}
}

// TestDrawSkipsInvisibleChildren verifies Draw does not call Draw on children
// that lack SfVisible.
// Spec: "for each visible child, creates a SubBuffer ... and calls child.Draw(sub)" —
// only visible children are drawn.
func TestDrawSkipsInvisibleChildren(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := &mockView{}
	v.SetBounds(NewRect(0, 0, 10, 5))
	// SfVisible is NOT set
	g.Insert(v)
	buf := NewDrawBuffer(80, 25)

	g.Draw(buf)

	if v.drawCalled {
		t.Errorf("Draw called Draw on invisible child")
	}
}

// TestDrawIteratesBackToFront verifies that children drawn later (higher index) paint
// over children drawn earlier (lower index), i.e. back-to-front order.
// Spec: "Draw(buf) iterates children back-to-front; ... last inserted draws last = on top".
func TestDrawIteratesBackToFront(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 20, 10))
	buf := NewDrawBuffer(20, 10)

	// Both children occupy the same cell (0,0).
	// The one drawn last wins. With back-to-front, the last-inserted child draws last.
	bottom := &drawOrderMockView{id: 'A', bounds: NewRect(0, 0, 10, 5)}
	bottom.SetState(SfVisible, true)
	top := &drawOrderMockView{id: 'B', bounds: NewRect(0, 0, 10, 5)}
	top.SetState(SfVisible, true)

	g.Insert(bottom)
	g.Insert(top)

	g.Draw(buf)

	// The last-inserted (top) should have drawn after bottom, so its rune wins.
	cell := buf.GetCell(0, 0)
	if cell.Rune != 'B' {
		t.Errorf("Draw order: cell(0,0) rune = %q, want 'B' (last-inserted draws on top)", cell.Rune)
	}
}

// drawOrderMockView writes its id rune into position (0,0) of the subbuffer it receives.
type drawOrderMockView struct {
	BaseView
	id     rune
	bounds Rect
}

func (d *drawOrderMockView) Draw(buf *DrawBuffer) {
	buf.WriteChar(0, 0, d.id, tcell.StyleDefault)
}

func (d *drawOrderMockView) HandleEvent(event *Event) {}

func (d *drawOrderMockView) Bounds() Rect {
	return d.bounds
}

// TestHandleEventForwardsToFocusedChild verifies HandleEvent sends the event to
// the focused child.
// Spec: "HandleEvent(event) forwards keyboard/command events to the focused child".
func TestHandleEventForwardsToFocusedChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := newSelectableMockView(NewRect(0, 0, 10, 5))
	g.Insert(v)
	event := &Event{What: EvKeyboard}

	g.HandleEvent(event)

	if v.eventHandled != event {
		t.Errorf("HandleEvent did not forward event to focused child")
	}
}

// TestHandleEventCommandForwardsToFocusedChild verifies HandleEvent forwards
// command events as well as keyboard events.
// Spec: "forwards keyboard/command events to the focused child".
func TestHandleEventCommandForwardsToFocusedChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := newSelectableMockView(NewRect(0, 0, 10, 5))
	g.Insert(v)
	event := &Event{What: EvCommand, Command: CmClose}

	g.HandleEvent(event)

	if v.eventHandled != event {
		t.Errorf("HandleEvent did not forward command event to focused child")
	}
}

// TestHandleEventNoFocusedChildDoesNotPanic verifies HandleEvent is safe when
// no child is focused.
// Spec: "HandleEvent(event) forwards keyboard/command events to the focused child" —
// the nil case must not panic.
func TestHandleEventNoFocusedChildDoesNotPanic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	event := &Event{What: EvKeyboard}

	// Must not panic.
	g.HandleEvent(event)
}

// TestHandleEventClearedEventNotForwarded verifies that a cleared event is not
// forwarded to the focused child.
// Spec: "HandleEvent with cleared event doesn't forward".
func TestHandleEventClearedEventNotForwarded(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	v := newSelectableMockView(NewRect(0, 0, 10, 5))
	g.Insert(v)
	event := &Event{What: EvNothing}

	g.HandleEvent(event)

	if v.eventHandled != nil {
		t.Errorf("HandleEvent forwarded a cleared event to focused child")
	}
}

