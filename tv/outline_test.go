package tv

// outline_test.go — Tests for Batch 9 Task 2: OutlineViewer widget.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Construction & initial state (NewOutlineViewer)
//   Section 2  — visibleCount (depth-first flattened visible rows)
//   Section 3  — nodeAt (node and level at a given flattened index)
//   Section 4  — ScrollBar integration (SetVScrollBar, syncScrollBars, ensureVisible, SetState)
//   Section 5  — Keyboard handling (Up/Down/Left/Right/Enter/+/-/*/PgUp/PgDn/Ctrl+arrows/Home/End)
//   Section 6  — Mouse handling (click to focus, graph vs text area, double-click)
//   Section 7  — Drawing (graph prefix characters, style selection)
//   Section 8  — adjust / adjustAll / selected placeholders

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// compile-time assertion: OutlineViewer must satisfy Widget.
var _ Widget = (*OutlineViewer)(nil)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// newOV creates an OutlineViewer with BorlandBlue scheme, bounds 40x10.
func newOV() *OutlineViewer {
	ov := NewOutlineViewer(NewRect(0, 0, 40, 10))
	ov.BaseView.scheme = theme.BorlandBlue
	return ov
}

// newOVFocused creates an OutlineViewer with SfSelected set.
func newOVFocused() *OutlineViewer {
	ov := newOV()
	ov.SetState(SfSelected, true)
	return ov
}

// ovKeyEv creates a plain keyboard event.
func ovKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Rune: 0, Modifiers: tcell.ModNone}}
}

// ovRuneEv creates a keyboard event for a rune.
func ovRuneEv(ch rune) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ch, Modifiers: tcell.ModNone}}
}

// ovCtrlKeyEv creates a keyboard event with Ctrl modifier.
func ovCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Rune: 0, Modifiers: tcell.ModCtrl}}
}

// ovMouseEv creates a mouse Button1 event at (x, y).
func ovMouseEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1}}
}

// ovMouseDblEv creates a mouse Button1 double-click event at (x, y).
func ovMouseDblEv(x, y int) *Event {
	return &Event{What: EvMouse, Mouse: &MouseEvent{X: x, Y: y, Button: tcell.Button1, ClickCount: 2}}
}

// makeSimpleTree creates a tree:
//   root (Expanded=true)
//     child1 (Expanded=true)
//       gc1 (Expanded=true, leaf)
//     child2 (Expanded=true, leaf)
//
// Visible flattened: [root, child1, gc1, child2]
func makeSimpleTree() *TNode {
	// gc1: leaf
	gc1 := NewNode("gc1", nil, nil)

	// child1: has child gc1
	child1 := NewNode("child1", gc1, nil)

	// child2: leaf, Next of child1
	child2 := NewNode("child2", nil, nil)
	child1.Next = child2

	// root: parent of child1
	root := NewNode("root", child1, nil)
	return root
}

// makeFlatSiblingTree creates a tree with three top-level siblings:
//   A → B → C  (all leaves)
// Visible flattened: [A, B, C]
func makeFlatSiblingTree() *TNode {
	a := NewNode("A", nil, nil)
	b := NewNode("B", nil, nil)
	c := NewNode("C", nil, nil)
	a.Next = b
	b.Next = c
	return a
}

// makeTallTree creates a flat sibling chain of 15 nodes (n0..n14).
// All 15 are visible, which exceeds the default viewer height of 10.
func makeTallTree() *TNode {
	var first, prev *TNode
	for i := 0; i < 15; i++ {
		n := NewNode(fmt.Sprintf("n%d", i), nil, nil)
		if first == nil {
			first = n
		}
		if prev != nil {
			prev.Next = n
		}
		prev = n
	}
	return first
}

// makeDeepTree creates a 3-level deep tree:
//   root (Expanded=true)
//     child1 (Expanded=true)
//       gc1 (Expanded=true, leaf)
//
// Set root on a viewer and gc1 collapsed.
func makeDeepTree() *TNode {
	gc1 := NewNode("gc1", nil, nil)
	child1 := NewNode("child1", gc1, nil)
	root := NewNode("root", child1, nil)
	return root
}

// ---------------------------------------------------------------------------
// Section 1 — Construction & initial state
// ---------------------------------------------------------------------------

// TestNewOutlineViewerSetsSfVisible verifies NewOutlineViewer sets SfVisible.
// Spec: "Sets SfVisible"
func TestNewOutlineViewerSetsSfVisible(t *testing.T) {
	ov := NewOutlineViewer(NewRect(0, 0, 40, 10))
	if !ov.HasState(SfVisible) {
		t.Error("NewOutlineViewer did not set SfVisible")
	}
}

// TestNewOutlineViewerSetsOfSelectable verifies NewOutlineViewer sets OfSelectable.
// Spec: "Sets OfSelectable"
func TestNewOutlineViewerSetsOfSelectable(t *testing.T) {
	ov := NewOutlineViewer(NewRect(0, 0, 40, 10))
	if !ov.HasOption(OfSelectable) {
		t.Error("NewOutlineViewer did not set OfSelectable")
	}
}

// TestNewOutlineViewerSetsOfFirstClick verifies NewOutlineViewer sets OfFirstClick.
// Spec: "Sets OfFirstClick"
func TestNewOutlineViewerSetsOfFirstClick(t *testing.T) {
	ov := NewOutlineViewer(NewRect(0, 0, 40, 10))
	if !ov.HasOption(OfFirstClick) {
		t.Error("NewOutlineViewer did not set OfFirstClick")
	}
}

// TestNewOutlineViewerStoresBounds verifies NewOutlineViewer records the given bounds.
// Spec: "NewOutlineViewer(bounds Rect) *OutlineViewer"
func TestNewOutlineViewerStoresBounds(t *testing.T) {
	r := NewRect(2, 3, 40, 10)
	ov := NewOutlineViewer(r)
	if ov.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", ov.Bounds(), r)
	}
}

// TestNewOutlineViewerInitialState verifies initial focusedIdx=0, deltaY=0, root=nil.
// Spec: "creates a viewer with no root (nil), focusedIdx=0, deltaY=0"
func TestNewOutlineViewerInitialState(t *testing.T) {
	ov := NewOutlineViewer(NewRect(0, 0, 40, 10))
	if ov.root != nil {
		t.Error("initial root should be nil")
	}
	if ov.focusedIdx != 0 {
		t.Errorf("initial focusedIdx = %d, want 0", ov.focusedIdx)
	}
	if ov.deltaY != 0 {
		t.Errorf("initial deltaY = %d, want 0", ov.deltaY)
	}
}

// TestNewOutlineViewerCallsSetSelf verifies NewOutlineViewer calls SetSelf.
// Spec: "Calls SetSelf"
func TestNewOutlineViewerCallsSetSelf(t *testing.T) {
	ov := NewOutlineViewer(NewRect(0, 0, 40, 10))
	// SetSelf stores the view as self; verify self == ov
	if ov.self != ov {
		t.Error("NewOutlineViewer should call SetSelf (ov.self should equal ov)")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — visibleCount
// ---------------------------------------------------------------------------

// TestVisibleCountNilRoot verifies visibleCount returns 0 for a nil root.
// Spec: "An empty tree (nil root) returns 0"
func TestVisibleCountNilRoot(t *testing.T) {
	ov := newOV()
	// root is nil by default
	if ov.visibleCount() != 0 {
		t.Errorf("visibleCount() = %d, want 0 (nil root)", ov.visibleCount())
	}
}

// TestVisibleCountSingleNode verifies visibleCount returns 1 for a single node.
// Spec: "Depth-first traversal counting only visible nodes"
func TestVisibleCountSingleNode(t *testing.T) {
	ov := newOV()
	ov.root = NewNode("single", nil, nil)
	if ov.visibleCount() != 1 {
		t.Errorf("visibleCount() = %d, want 1 (single node)", ov.visibleCount())
	}
}

// TestVisibleCountSiblingChain verifies visibleCount counts all siblings.
// Spec: "Depth-first traversal counting only visible nodes (root -> children if expanded -> next sibling)"
func TestVisibleCountSiblingChain(t *testing.T) {
	ov := newOV()
	ov.root = makeFlatSiblingTree() // A -> B -> C
	if ov.visibleCount() != 3 {
		t.Errorf("visibleCount() = %d, want 3 (A, B, C siblings)", ov.visibleCount())
	}
}

// TestVisibleCountExpandedChildren verifies visibleCount includes children when node is expanded.
// Spec: "root -> children if expanded -> next sibling"
func TestVisibleCountExpandedChildren(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // root + child1 + gc1 + child2 = 4
	if ov.visibleCount() != 4 {
		t.Errorf("visibleCount() = %d, want 4 (root expanded, child1 expanded)", ov.visibleCount())
	}
}

// TestVisibleCountCollapsedChildren verifies visibleCount excludes children when node is collapsed.
// Spec: "root -> children if expanded -> next sibling"
// Collapsed nodes: children are NOT visited, traversal goes to Next sibling.
func TestVisibleCountCollapsedChildren(t *testing.T) {
	ov := newOV()
	root := makeSimpleTree()       // root + child1 + gc1 + child2 = 4
	root.Children.Expanded = false // collapse child1: gc1 hidden
	ov.root = root
	// Now visible: root, child1, child2 = 3
	if ov.visibleCount() != 3 {
		t.Errorf("visibleCount() = %d, want 3 (child1 collapsed, gc1 hidden)", ov.visibleCount())
	}
}

// TestVisibleCountCollapsedRoot verifies collapsing root hides all children.
// Spec: "root -> children if expanded"
func TestVisibleCountCollapsedRoot(t *testing.T) {
	ov := newOV()
	root := makeSimpleTree()
	root.Expanded = false // collapse root
	ov.root = root
	// Now visible: only root = 1
	if ov.visibleCount() != 1 {
		t.Errorf("visibleCount() = %d, want 1 (root collapsed)", ov.visibleCount())
	}
}

// TestVisibleCountMultiLevelExpanded verifies visibleCount for a deep expanded tree.
// Spec: "Depth-first traversal counting only visible nodes"
func TestVisibleCountMultiLevelExpanded(t *testing.T) {
	ov := newOV()
	ov.root = makeDeepTree() // root + child1 + gc1 = 3
	if ov.visibleCount() != 3 {
		t.Errorf("visibleCount() = %d, want 3 (root + child1 + gc1)", ov.visibleCount())
	}
}

// TestVisibleCountChangesAfterToggle verifies visibleCount changes after toggling Expanded.
// Spec: "Expanded controls whether children are visible"
func TestVisibleCountChangesAfterToggle(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // 4 visible
	if ov.visibleCount() != 4 {
		t.Fatalf("setup: visibleCount() = %d, want 4", ov.visibleCount())
	}
	// Toggle root collapsed
	ov.root.Expanded = false
	if ov.visibleCount() != 1 {
		t.Errorf("after collapsing root, visibleCount() = %d, want 1", ov.visibleCount())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — nodeAt
// ---------------------------------------------------------------------------

// TestNodeAtNilRoot verifies nodeAt returns nil, 0 for nil root.
// Spec: "Returns nil, 0 if idx is out of bounds or root is nil"
func TestNodeAtNilRoot(t *testing.T) {
	ov := newOV()
	node, level := ov.nodeAt(0)
	if node != nil {
		t.Error("nodeAt(0) with nil root: node should be nil")
	}
	if level != 0 {
		t.Errorf("nodeAt(0) with nil root: level = %d, want 0", level)
	}
}

// TestNodeAtIndex0IsRoot verifies nodeAt(0) returns the root node at level 0.
// Spec: "Returns the node at flattened index idx and its nesting level (root = level 0)"
func TestNodeAtIndex0IsRoot(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree()
	node, level := ov.nodeAt(0)
	if node == nil {
		t.Fatal("nodeAt(0) should not be nil")
	}
	if node.Text != "root" {
		t.Errorf("nodeAt(0).Text = %q, want %q", node.Text, "root")
	}
	if level != 0 {
		t.Errorf("nodeAt(0): level = %d, want 0 (root)", level)
	}
}

// TestNodeAtIndex1IsFirstChild verifies nodeAt returns the first child at level 1.
// Spec: "Same depth-first traversal as visibleCount()"
func TestNodeAtIndex1IsFirstChild(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // [root, child1, gc1, child2]
	node, level := ov.nodeAt(1)
	if node == nil {
		t.Fatal("nodeAt(1) should not be nil")
	}
	if node.Text != "child1" {
		t.Errorf("nodeAt(1).Text = %q, want %q", node.Text, "child1")
	}
	if level != 1 {
		t.Errorf("nodeAt(1): level = %d, want 1", level)
	}
}

// TestNodeAtGrandchildLevel verifies nodeAt returns the grandchild at level 2.
// Spec: "Same depth-first traversal as visibleCount()"
func TestNodeAtGrandchildLevel(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // [root, child1, gc1, child2]
	node, level := ov.nodeAt(2)
	if node == nil {
		t.Fatal("nodeAt(2) should not be nil")
	}
	if node.Text != "gc1" {
		t.Errorf("nodeAt(2).Text = %q, want %q", node.Text, "gc1")
	}
	if level != 2 {
		t.Errorf("nodeAt(2): level = %d, want 2", level)
	}
}

// TestNodeAtLastSibling verifies nodeAt returns the last sibling.
// Spec: "Same depth-first traversal as visibleCount()"
func TestNodeAtLastSibling(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // [root, child1, gc1, child2]
	node, level := ov.nodeAt(3)
	if node == nil {
		t.Fatal("nodeAt(3) should not be nil")
	}
	if node.Text != "child2" {
		t.Errorf("nodeAt(3).Text = %q, want %q", node.Text, "child2")
	}
	if level != 1 {
		t.Errorf("nodeAt(3): level = %d, want 1", level)
	}
}

// TestNodeAtSkipsCollapsedChildren verifies nodeAt skips children of collapsed nodes.
// Spec: "root -> children if expanded"
func TestNodeAtSkipsCollapsedChildren(t *testing.T) {
	ov := newOV()
	root := makeSimpleTree()
	root.Children.Expanded = false // collapse child1: gc1 hidden
	ov.root = root
	// Flattened: [root, child1, child2]
	node, _ := ov.nodeAt(1)
	if node == nil {
		t.Fatal("nodeAt(1) should not be nil")
	}
	if node.Text != "child1" {
		t.Errorf("nodeAt(1).Text = %q, want %q (child1)", node.Text, "child1")
	}
	node, _ = ov.nodeAt(2)
	if node == nil {
		t.Fatal("nodeAt(2) should not be nil")
	}
	if node.Text != "child2" {
		t.Errorf("nodeAt(2).Text = %q, want %q (child2, skipping gc1)", node.Text, "child2")
	}
}

// TestNodeAtOutOfBounds verifies nodeAt returns nil, 0 for out-of-bounds index.
// Spec: "Returns nil, 0 if idx is out of bounds or root is nil"
func TestNodeAtOutOfBounds(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // 4 visible
	node, level := ov.nodeAt(99)
	if node != nil {
		t.Error("nodeAt(99) should return nil (out of bounds)")
	}
	if level != 0 {
		t.Errorf("nodeAt(99): level = %d, want 0", level)
	}
}

// TestNodeAtNegativeIndex verifies nodeAt handles negative index as out of bounds.
// Spec: "Returns nil, 0 if idx is out of bounds"
func TestNodeAtNegativeIndex(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree()
	node, level := ov.nodeAt(-1)
	if node != nil {
		t.Error("nodeAt(-1) should return nil (out of bounds)")
	}
	if level != 0 {
		t.Errorf("nodeAt(-1): level = %d, want 0", level)
	}
}

// TestNodeAtFlatSiblingTree verifies nodeAt returns siblings at same level.
func TestNodeAtFlatSiblingTree(t *testing.T) {
	ov := newOV()
	ov.root = makeFlatSiblingTree() // [A, B, C] all level 0
	for i, text := range []string{"A", "B", "C"} {
		node, level := ov.nodeAt(i)
		if node == nil {
			t.Fatalf("nodeAt(%d) should not be nil", i)
		}
		if node.Text != text {
			t.Errorf("nodeAt(%d).Text = %q, want %q", i, node.Text, text)
		}
		if level != 0 {
			t.Errorf("nodeAt(%d): level = %d, want 0 (flat sibling)", i, level)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 4 — ScrollBar integration
// ---------------------------------------------------------------------------

// TestSetVScrollBarHooksOnChange verifies SetVScrollBar sets up the OnChange callback.
// Spec: "sets up scrollbar with OnChange callback that sets deltaY"
func TestSetVScrollBarHooksOnChange(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // 4 visible
	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	ov.SetVScrollBar(sb)

	if sb.OnChange == nil {
		t.Fatal("SetVScrollBar should set sb.OnChange")
	}

	// Simulate scrollbar change
	sb.OnChange(1)
	if ov.deltaY != 1 {
		t.Errorf("after OnChange(1): deltaY = %d, want 1", ov.deltaY)
	}
}

// TestSetVScrollBarNilClears verifies SetVScrollBar(nil) clears the existing scrollbar.
// Spec: "SetVScrollBar with nil should clear the existing scrollbar (set OnChange to nil first)"
func TestSetVScrollBarNilClears(t *testing.T) {
	ov := newOV()
	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	ov.SetVScrollBar(sb)

	// Verify OnChange was set
	if sb.OnChange == nil {
		t.Fatal("SetVScrollBar should have set sb.OnChange")
	}

	ov.SetVScrollBar(nil)

	// After nil, sb.OnChange should be cleared
	if sb.OnChange != nil {
		t.Error("after SetVScrollBar(nil), sb.OnChange should be nil")
	}
}

// TestSetVScrollBarSyncsScrollBars verifies SetVScrollBar calls syncScrollBars.
// Spec: "Calls syncScrollBars()"
func TestSetVScrollBarSyncsScrollBars(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // 4 visible
	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	ov.SetVScrollBar(sb)

	// syncScrollBars sets range to 0..visibleCount()-1, pageSize=bounds.Height(), value=deltaY
	if sb.Min() != 0 {
		t.Errorf("ScrollBar Min() = %d, want 0", sb.Min())
	}
	if sb.Max() != 4 {
		t.Errorf("ScrollBar Max() = %d, want 4 (visibleCount=4)", sb.Max())
	}
	if sb.PageSize() != 10 {
		t.Errorf("ScrollBar PageSize() = %d, want 10 (bounds height)", sb.PageSize())
	}
	if sb.Value() != 0 {
		t.Errorf("ScrollBar Value() = %d, want 0 (deltaY)", sb.Value())
	}
}

// TestSyncScrollBarsUpdatesWithDeltaY verifies syncScrollBars uses current deltaY.
// Spec: "syncScrollBars() — sets scrollbar ... value to deltaY"
func TestSyncScrollBarsUpdatesWithDeltaY(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // 4 visible
	ov.deltaY = 2
	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	ov.SetVScrollBar(sb)

	if sb.Value() != 2 {
		t.Errorf("ScrollBar Value() = %d, want 2 (deltaY)", sb.Value())
	}
}

// TestEnsureVisibleAdjustsDeltaYBelow verifies ensureVisible when focusedIdx < deltaY.
// Spec: "if focusedIdx < deltaY, set deltaY = focusedIdx"
func TestEnsureVisibleAdjustsDeltaYBelow(t *testing.T) {
	ov := newOV()
	ov.root = makeTallTree() // 15 visible items, height=10 → maxDelta=5
	ov.focusedIdx = 1
	ov.deltaY = 3
	ov.ensureVisible()

	if ov.deltaY != 1 {
		t.Errorf("deltaY = %d, want 1 (focusedIdx < deltaY)", ov.deltaY)
	}
}

// TestEnsureVisibleAdjustsDeltaYAbove verifies ensureVisible when focusedIdx >= deltaY + height.
// Spec: "If focusedIdx >= deltaY + bounds.Height(), set deltaY = focusedIdx - bounds.Height() + 1"
func TestEnsureVisibleAdjustsDeltaYAbove(t *testing.T) {
	ov := newOV()
	// bounds height = 10, 15 visible → maxDelta=5
	ov.root = makeTallTree() // 15 visible
	ov.focusedIdx = 12
	ov.deltaY = 2
	// 12 >= 2 + 10 = 12, so deltaY = 12 - 10 + 1 = 3
	ov.ensureVisible()

	if ov.deltaY != 3 {
		t.Errorf("deltaY = %d, want 3 (focusedIdx >= deltaY + height)", ov.deltaY)
	}
}

// TestEnsureVisibleNoScrollWhenVisible verifies ensureVisible does nothing when focusedIdx is in view.
// Spec: implicit — only adjusts when out of bounds
func TestEnsureVisibleNoScrollWhenVisible(t *testing.T) {
	ov := newOV()
	ov.root = makeTallTree() // 15 visible, height=10 → maxDelta=5
	ov.focusedIdx = 3
	ov.deltaY = 2 // visible rows: 2..11, focusedIdx=3 is visible
	ov.ensureVisible()

	if ov.deltaY != 2 {
		t.Errorf("deltaY changed from 2 to %d, should stay 2 (focusedIdx visible)", ov.deltaY)
	}
}

// TestEnsureVisibleClampsDeltaY verifies ensureVisible clamps deltaY.
// Spec: "Clamp deltaY and sync scrollbar"
func TestEnsureVisibleClampsDeltaY(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // 4 visible
	// bounds height=10, visibleCount=4
	// max deltaY = visibleCount - height = 4 - 10 = -6, clamped to 0
	ov.focusedIdx = 50 // should be out of bounds
	ov.deltaY = 0
	ov.ensureVisible()

	// deltaY should be clamped to max valid value
	// With visibleCount=4 and height=10, max deltaY = max(0, 4-10) = 0
	if ov.deltaY != 0 {
		t.Errorf("deltaY = %d, want 0 (clamped to max valid)", ov.deltaY)
	}
}

// TestSetStateTogglesVScrollBarVisibility verifies SetState toggles vScrollBar visibility.
// Spec: "when SfSelected changes, toggle vScrollBar visibility to match"
func TestSetStateTogglesVScrollBarVisibility(t *testing.T) {
	ov := newOV()
	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	ov.SetVScrollBar(sb)

	// Initially not selected -> scrollbar should not be visible
	if sb.HasState(SfVisible) {
		t.Error("vScrollBar should not be SfVisible when viewer is not selected")
	}

	// Set selected -> scrollbar should become visible
	ov.SetState(SfSelected, true)
	if !sb.HasState(SfVisible) {
		t.Error("vScrollBar should be SfVisible when viewer is selected")
	}

	// Clear selected -> scrollbar should hide
	ov.SetState(SfSelected, false)
	if sb.HasState(SfVisible) {
		t.Error("vScrollBar should not be SfVisible when viewer is not selected")
	}
}

// TestSetVScrollBarNilDoesNotPanic verifies SetVScrollBar(nil) does not panic on fresh viewer.
// Spec: "SetVScrollBar with nil should clear the existing scrollbar"
func TestSetVScrollBarNilDoesNotPanic(t *testing.T) {
	ov := newOV()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetVScrollBar(nil) on fresh viewer panicked: %v", r)
		}
	}()
	ov.SetVScrollBar(nil)
}

// ---------------------------------------------------------------------------
// Section 5 — Keyboard handling
// ---------------------------------------------------------------------------

// TestKeyDownMovesFocusedIdx verifies Down arrow moves focusedIdx to next visible row.
// Spec: "Down / tcell.KeyDown: move focusedIdx to next visible row (clamped to visibleCount()-1)"
func TestKeyDownMovesFocusedIdx(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree() // 4 visible
	ev := ovKeyEv(tcell.KeyDown)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 1 {
		t.Errorf("after Down: focusedIdx = %d, want 1", ov.focusedIdx)
	}
}

// TestKeyDownClampedAtEnd verifies Down arrow does not go past last visible row.
// Spec: "clamped to visibleCount()-1"
func TestKeyDownClampedAtEnd(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree() // 4 visible
	ov.focusedIdx = 3 // last visible
	ev := ovKeyEv(tcell.KeyDown)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 3 {
		t.Errorf("Down at last: focusedIdx = %d, want 3 (no change)", ov.focusedIdx)
	}
}

// TestKeyUpMovesFocusedIdx verifies Up arrow moves focusedIdx to previous visible row.
// Spec: "Up / tcell.KeyUp: move focusedIdx to previous visible row (clamped to 0)"
func TestKeyUpMovesFocusedIdx(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree() // 4 visible
	ov.focusedIdx = 2
	ev := ovKeyEv(tcell.KeyUp)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 1 {
		t.Errorf("after Up: focusedIdx = %d, want 1", ov.focusedIdx)
	}
}

// TestKeyUpClampedAtStart verifies Up arrow does not go below 0.
// Spec: "clamped to 0"
func TestKeyUpClampedAtStart(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree() // 4 visible
	// focusedIdx is already 0
	ev := ovKeyEv(tcell.KeyUp)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 0 {
		t.Errorf("Up at first: focusedIdx = %d, want 0 (no change)", ov.focusedIdx)
	}
}

// TestKeyRightAliasForDown verifies Right arrow is an alias for Down.
// Spec: "Right / tcell.KeyRight: alias for Down"
func TestKeyRightAliasForDown(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree() // 4 visible
	ev := ovKeyEv(tcell.KeyRight)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 1 {
		t.Errorf("after Right: focusedIdx = %d, want 1 (alias for Down)", ov.focusedIdx)
	}
}

// TestKeyLeftAliasForUp verifies Left arrow is an alias for Up.
// Spec: "Left / tcell.KeyLeft: alias for Up"
func TestKeyLeftAliasForUp(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree() // 4 visible
	ov.focusedIdx = 2
	ev := ovKeyEv(tcell.KeyLeft)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 1 {
		t.Errorf("after Left: focusedIdx = %d, want 1 (alias for Up)", ov.focusedIdx)
	}
}

// TestKeyEnterIsConsumed verifies Enter is consumed by HandleEvent.
// Spec: "Enter / tcell.KeyEnter: call selected() on focused node (placeholder)"
// selected() is a no-op placeholder; the key should be consumed regardless of what selected() does.
func TestKeyEnterIsConsumed(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree() // 4 visible
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	ov.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Enter key: event should be consumed")
	}
}

// TestKeyPlusTogglesExpand verifies '+' toggles Expanded via adjust().
// Spec: "+ (rune '+'): call adjust() — if node has children, toggle Expanded"
func TestKeyPlusTogglesExpand(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree() // root focused (idx 0), root has children, Expanded=true
	ev := ovRuneEv('+')
	ov.HandleEvent(ev)

	if ov.root.Expanded != false {
		t.Errorf("after '+': root.Expanded = %v, want false (toggled)", ov.root.Expanded)
	}
}

// TestKeyPlusTogglesLeafExpanded verifies '+' calls adjust() which always toggles Expanded.
// Spec: "+ (rune '+'): call adjust() — toggles Expanded on focused node"
func TestKeyPlusTogglesLeafExpanded(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree() // 4 visible
	// Move focus to gc1 (idx 2, leaf with no children)
	ov.focusedIdx = 2
	gc1 := ov.root.Children.Children // gc1 is a leaf
	prevExpanded := gc1.Expanded

	ev := ovRuneEv('+')
	ov.HandleEvent(ev)

	// adjust() always toggles, even on leaves. The toggle has no visible effect
	// since a leaf has no children, but the flag itself is toggled.
	if gc1.Expanded != !prevExpanded {
		t.Errorf("'+' on leaf: Expanded = %v, want %v (adjust always toggles)", gc1.Expanded, !prevExpanded)
	}
}

// TestKeyMinusTogglesExpand verifies '-' toggles Expanded via adjust().
// Spec: "- (rune '-'): call adjust() — if node has children, toggle Expanded"
func TestKeyMinusTogglesExpand(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree() // root has children, Expanded=true
	ev := ovRuneEv('-')
	ov.HandleEvent(ev)

	if ov.root.Expanded != false {
		t.Errorf("after '-': root.Expanded = %v, want false (toggled)", ov.root.Expanded)
	}
}

// TestKeyStarExpandsAll verifies '*' expands node and all descendants.
// Spec: "* (rune '*'): call adjustAll() — expand node and all descendants"
func TestKeyStarExpandsAll(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree()
	// Collapse everything first
	ov.root.Expanded = false
	ov.root.Children.Expanded = false

	ev := ovRuneEv('*')
	ov.HandleEvent(ev)

	if !ov.root.Expanded {
		t.Error("after '*': root.Expanded should be true")
	}
	if !ov.root.Children.Expanded {
		t.Error("after '*': child1.Expanded should be true")
	}
}

// TestKeyPgUpMovesByHeight verifies PgUp moves focusedIdx up by bounds.Height()-1.
// Spec: "PgUp: move focusedIdx up by bounds.Height()-1 rows (one row overlap between pages)"
func TestKeyPgUpMovesByHeight(t *testing.T) {
	ov := newOVFocused()
	// Need enough visible rows; use a flat chain of siblings
	ov.root = makeSiblingList(20) // 20 siblings
	ov.focusedIdx = 15
	ev := ovKeyEv(tcell.KeyPgUp)
	ov.HandleEvent(ev)

	// 15 - 9 = 6 (height-1 = 9)
	if ov.focusedIdx != 6 {
		t.Errorf("after PgUp: focusedIdx = %d, want 6 (15 - 9)", ov.focusedIdx)
	}
}

// TestKeyPgUpClampedToZero verifies PgUp clamps to 0.
// Spec: "PgUp: move focusedIdx up by bounds.Height()-1 rows" (implicit clamped)
func TestKeyPgUpClampedToZero(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSiblingList(20)
	ov.focusedIdx = 5
	ev := ovKeyEv(tcell.KeyPgUp)
	ov.HandleEvent(ev)

	// 5 - 9 = -4, should clamp to 0
	if ov.focusedIdx != 0 {
		t.Errorf("after PgUp from 5: focusedIdx = %d, want 0 (clamped)", ov.focusedIdx)
	}
}

// TestKeyPgDnMovesByHeight verifies PgDn moves focusedIdx down by bounds.Height()-1.
// Spec: "PgDn: move focusedIdx down by bounds.Height()-1 rows (one row overlap between pages)"
func TestKeyPgDnMovesByHeight(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSiblingList(20) // 20 siblings
	ov.focusedIdx = 2
	ev := ovKeyEv(tcell.KeyPgDn)
	ov.HandleEvent(ev)

	// 2 + 9 = 11 (height-1 = 9)
	if ov.focusedIdx != 11 {
		t.Errorf("after PgDn: focusedIdx = %d, want 11 (2 + 9)", ov.focusedIdx)
	}
}

// TestKeyPgDnClampedToLast verifies PgDn clamps to visibleCount()-1.
// Spec: "PgDn: move focusedIdx down by bounds.Height()-1 rows" (implicit clamped)
func TestKeyPgDnClampedToLast(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSiblingList(20) // 20 siblings
	ov.focusedIdx = 15
	ev := ovKeyEv(tcell.KeyPgDn)
	ov.HandleEvent(ev)

	// 15 + 9 = 24, clamped to 19
	if ov.focusedIdx != 19 {
		t.Errorf("after PgDn from 15: focusedIdx = %d, want 19 (clamped to last)", ov.focusedIdx)
	}
}

// TestKeyCtrlPgUpMovesToTop verifies Ctrl+PgUp moves focusedIdx to 0.
// Spec: "Ctrl+PgUp: move focusedIdx to 0 (absolute top)"
func TestKeyCtrlPgUpMovesToTop(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSiblingList(20)
	ov.focusedIdx = 7
	ev := ovCtrlKeyEv(tcell.KeyPgUp)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 0 {
		t.Errorf("after Ctrl+PgUp: focusedIdx = %d, want 0 (absolute top)", ov.focusedIdx)
	}
}

// TestKeyCtrlPgDnMovesToBottom verifies Ctrl+PgDn moves focusedIdx to visibleCount()-1.
// Spec: "Ctrl+PgDn: move focusedIdx to visibleCount()-1 (absolute bottom)"
func TestKeyCtrlPgDnMovesToBottom(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSiblingList(20) // 20 siblings
	ov.focusedIdx = 3
	ev := ovCtrlKeyEv(tcell.KeyPgDn)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 19 {
		t.Errorf("after Ctrl+PgDn: focusedIdx = %d, want 19 (absolute bottom)", ov.focusedIdx)
	}
}

// TestKeyHomeMovesToDeltaY verifies Home moves focusedIdx to deltaY (first visible row).
// Spec: "Home: move focusedIdx to deltaY (first visible row in viewport)"
func TestKeyHomeMovesToDeltaY(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSiblingList(20)
	ov.deltaY = 5
	ov.focusedIdx = 10
	ev := ovKeyEv(tcell.KeyHome)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 5 {
		t.Errorf("after Home: focusedIdx = %d, want 5 (deltaY)", ov.focusedIdx)
	}
}

// TestKeyEndMovesToLastVisibleRow verifies End moves focusedIdx to deltaY + height - 1.
// Spec: "End: move focusedIdx to deltaY + bounds.Height() - 1, clamped to last visible"
func TestKeyEndMovesToLastVisibleRow(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSiblingList(20) // 20 siblings
	ov.deltaY = 5
	ov.focusedIdx = 6
	ev := ovKeyEv(tcell.KeyEnd)
	ov.HandleEvent(ev)

	// deltaY + height - 1 = 5 + 10 - 1 = 14
	if ov.focusedIdx != 14 {
		t.Errorf("after End: focusedIdx = %d, want 14 (deltaY + height - 1)", ov.focusedIdx)
	}
}

// TestKeyEndClampedToLast verifies End clamps to visibleCount()-1 when viewport extends past end.
// Spec: "clamped to last visible"
func TestKeyEndClampedToLast(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSiblingList(20) // 20 siblings
	ov.deltaY = 15
	// deltaY + height - 1 = 15 + 10 - 1 = 24, but only 20 visible
	ev := ovKeyEv(tcell.KeyEnd)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 19 {
		t.Errorf("after End with deltaY=15: focusedIdx = %d, want 19 (clamped to last)", ov.focusedIdx)
	}
}

// TestMovementCallsEnsureVisible verifies movement keys call ensureVisible and syncScrollBars.
// Spec: "All movement handlers call ensureVisible() and syncScrollBars() to scroll and update scrollbar"
func TestMovementCallsEnsureVisible(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSiblingList(20) // 20 siblings
	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	ov.SetVScrollBar(sb)

	// Scroll down past the visible area
	for i := 0; i < 15; i++ {
		ev := ovKeyEv(tcell.KeyDown)
		ov.HandleEvent(ev)
	}

	// focusedIdx=15, ensureVisible should adjust deltaY
	if ov.focusedIdx != 15 {
		t.Errorf("after 15 Down: focusedIdx = %d, want 15", ov.focusedIdx)
	}
	// deltaY should have been adjusted: 15 >= 0+10, so deltaY = 15-10+1 = 6
	if ov.deltaY != 6 {
		t.Errorf("after 15 Down: deltaY = %d, want 6 (scrolled to show focused row)", ov.deltaY)
	}
	// Scrollbar should be synced
	if sb.Value() != ov.deltaY {
		t.Errorf("ScrollBar Value() = %d, want %d (deltaY)", sb.Value(), ov.deltaY)
	}
}

// TestKeyboardOnlyWhenSelected verifies keyboard events are ignored when SfSelected is not set.
// Spec: "Only handles events when SfSelected is set"
func TestKeyboardOnlyWhenSelected(t *testing.T) {
	ov := newOV() // not focused
	ov.root = makeSimpleTree() // 4 visible
	ev := ovKeyEv(tcell.KeyDown)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 0 {
		t.Errorf("Down without SfSelected: focusedIdx = %d, want 0 (unhandled)", ov.focusedIdx)
	}
	if ev.IsCleared() {
		t.Error("Down without SfSelected: event should not have been consumed")
	}
}

// TestKeyboardEmptyTree does not panic when keyboard used on empty tree.
// Spec: "Only handles events when SfSelected is set" — but even when selected, empty tree should not crash
func TestKeyboardEmptyTree(t *testing.T) {
	ov := newOVFocused()
	// root is nil, visibleCount = 0
	// Down on empty tree should be a no-op (focusedIdx clamped to 0, which equals visibleCount()-1 = -1)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Down on empty tree panicked: %v", r)
		}
	}()

	ev := ovKeyEv(tcell.KeyDown)
	ov.HandleEvent(ev)
	// focusedIdx should remain 0 (clamped)
	if ov.focusedIdx != 0 {
		t.Errorf("Down on empty tree: focusedIdx = %d, want 0", ov.focusedIdx)
	}
}

// TestKeyPgUpOnEmptyTree does not panic on empty tree.
func TestKeyPgUpOnEmptyTree(t *testing.T) {
	ov := newOVFocused()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PgUp on empty tree panicked: %v", r)
		}
	}()
	ev := ovKeyEv(tcell.KeyPgUp)
	ov.HandleEvent(ev)
}

// TestKeyMinusOnEmptyTree does not panic on empty tree.
func TestKeyMinusOnEmptyTree(t *testing.T) {
	ov := newOVFocused()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("'-' on empty tree panicked: %v", r)
		}
	}()
	ev := ovRuneEv('-')
	ov.HandleEvent(ev)
}

// TestUnhandledKeyPassesThrough verifies unhandled keys are not consumed.
// Spec: implicit — only known keys are consumed
func TestUnhandledKeyPassesThrough(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree()
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF1}}
	ov.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("unhandled key F1: event was consumed, want pass-through")
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Mouse handling
// ---------------------------------------------------------------------------

// TestOutlineMouseClickSelectsRow verifies clicking on a visible row moves focus.
// Spec: "Compute which visible row was clicked from mouse Y relative to bounds.A.Y + deltaY"
// "If row is valid, move focus to that row"
func TestOutlineMouseClickSelectsRow(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // 4 visible
	// Click at (5, 2) — Y=2 means row 2 (deltaY=0), which is gc1
	ev := ovMouseEv(5, 2)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 2 {
		t.Errorf("click on row 2: focusedIdx = %d, want 2", ov.focusedIdx)
	}
}

// TestMouseClickWithScrollOffset verifies click accounts for deltaY.
// Spec: "Y relative to bounds.A.Y + deltaY"
func TestMouseClickWithScrollOffset(t *testing.T) {
	ov := newOV()
	ov.root = makeSiblingList(20)
	ov.deltaY = 5
	// Click at Y=3: actual row = deltaY + clickY = 5 + 3 = 8
	ev := ovMouseEv(5, 3)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 8 {
		t.Errorf("click at Y=3 with deltaY=5: focusedIdx = %d, want 8", ov.focusedIdx)
	}
}

// TestMouseClickOutsideBoundsIgnored verifies click outside widget bounds is ignored.
// Spec: must call BaseView.HandleEvent first for click-to-focus
// Click outside bounds should not change focusedIdx
func TestMouseClickOutsideBoundsIgnored(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree()
	// Click at Y=15 (outside bounds height of 10)
	ev := ovMouseEv(5, 15)
	ov.HandleEvent(ev)

	// focusedIdx should not change (or if it does, it's via BaseView which we can't override here)
	// The key test: the event may be handled by BaseView for click-to-focus,
	// but if the mouse is out of bounds, the idx calculation should not apply.
	// We just verify no panic and focusedIdx doesn't go out of range.
	if ov.focusedIdx < 0 || ov.focusedIdx >= ov.visibleCount() && ov.visibleCount() > 0 {
		t.Errorf("click outside bounds: focusedIdx = %d, should be in range [0, %d)",
			ov.focusedIdx, ov.visibleCount())
	}
}

// TestMouseClickGraphAreaTogglesExpand verifies clicking on graph prefix area calls adjust().
// Spec: "Graph area = the prefix region (first few characters): call adjust() (toggle expand/collapse)"
func TestMouseClickGraphAreaTogglesExpand(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // root focused (idx 0)
	// root at level 0: prefix = connector(3) + status(1) = 4 chars of graph area
	// Click at X=0 (first char of graph area)
	ev := ovMouseEv(0, 0)
	ov.HandleEvent(ev)

	// Focus should move to row 0 (already there) AND graph area click should toggle expand
	// root.Expanded starts true, graph click should toggle it
	if ov.root.Expanded != false {
		t.Errorf("click on graph area (X=0): root.Expanded = %v, want false (toggled)", ov.root.Expanded)
	}
}

// TestMouseClickTextAreaOnlyMovesFocus verifies clicking on text area just moves focus, no toggle.
// Spec: "Text area = beyond prefix: just move focus"
func TestMouseClickTextAreaOnlyMovesFocus(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // 4 visible
	// Move focus to row 1 (child1, level 1)
	// child1 at level 1: prefix = "   " (3 for root without Next) + "├──" (3, has Next) + "─" (1, expanded) + " " = 8 chars
	// Wait, let me just click far right (X=30) which is definitely in text area
	ov.focusedIdx = 1
	child1 := ov.root.Children
	prevExpanded := child1.Expanded

	// Double-click at row 0 in text area. But single click should NOT toggle.
	// Actually, re-reading the spec: single click in graph area toggles, single click in text just moves focus.
	// Let's verify single click in text area doesn't toggle.
	ev := ovMouseEv(30, 1) // click in text area of row 1
	ov.HandleEvent(ev)

	if ov.focusedIdx != 1 {
		t.Errorf("click in text area row 1: focusedIdx = %d, want 1", ov.focusedIdx)
	}
	if child1.Expanded != prevExpanded {
		t.Errorf("click in text area: Expanded changed from %v to %v, should NOT toggle", prevExpanded, child1.Expanded)
	}
}

// TestMouseDoubleClickCallsSelected verifies double-click in text area calls selected().
// Spec: "Double-click in text area: move focus, then call selected()"
func TestMouseDoubleClickCallsSelected(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // 4 visible
	// Double-click on row 0 in text area (X=30)
	ev := ovMouseDblEv(30, 0)
	ov.HandleEvent(ev)

	// Focus should move to row 0
	// selected() is a no-op placeholder; we can't verify it was called directly
	// but we can verify the event was consumed and focus moved
	if !ev.IsCleared() {
		t.Error("double-click in text area: event should be consumed")
	}
	if ov.focusedIdx != 0 {
		t.Errorf("double-click row 0: focusedIdx = %d, want 0", ov.focusedIdx)
	}
}

// TestMouseDoubleClickWithChildrenTogglesExpand verifies double-click on a node with children also calls adjust().
// Spec: "Double-click in text area: move focus, then call selected(). If node has children, also call adjust()."
func TestMouseDoubleClickWithChildrenTogglesExpand(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // root has children
	// Double-click on root (row 0) in text area (X=30)
	ev := ovMouseDblEv(30, 0)
	ov.HandleEvent(ev)

	// Root has children, so adjust() should toggle Expanded
	if ov.root.Expanded != false {
		t.Errorf("double-click on root (has children): root.Expanded = %v, want false (toggled)", ov.root.Expanded)
	}
}

// TestMouseClickEmptyTree verifies mouse click on empty tree does not panic.
func TestMouseClickEmptyTree(t *testing.T) {
	ov := newOV()
	// root is nil
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("mouse click on empty tree panicked: %v", r)
		}
	}()
	ev := ovMouseEv(5, 2)
	ov.HandleEvent(ev)
}

// TestMouseDoubleClickEmptyTree verifies double-click on empty tree does not panic.
func TestMouseDoubleClickEmptyTree(t *testing.T) {
	ov := newOV()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("double-click on empty tree panicked: %v", r)
		}
	}()
	ev := ovMouseDblEv(5, 2)
	ov.HandleEvent(ev)
}

// TestMouseClickRowZero verifies clicking row 0 with no scroll offset selects deltaY.
// Spec: "Compute which visible row was clicked from mouse Y relative to bounds.A.Y + deltaY"
func TestMouseClickRowZero(t *testing.T) {
	ov := newOV()
	ov.root = makeSiblingList(20) // 20 siblings
	ov.deltaY = 5
	// Click Y=0: actual idx = 5 + 0 = 5
	ev := ovMouseEv(5, 0)
	ov.HandleEvent(ev)

	if ov.focusedIdx != 5 {
		t.Errorf("click row 0 with deltaY=5: focusedIdx = %d, want 5", ov.focusedIdx)
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Drawing
// ---------------------------------------------------------------------------

// TestDrawIncludesTreeChars verifies drawn buffer contains tree-drawing characters.
// Spec: "Each visible row draws a graph prefix + node text"
// "Node connector: ├── (non-last) or └── (last)"
func TestDrawIncludesTreeChars(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // root + child1 + gc1 + child2
	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	// Check for presence of tree-drawing characters
	hasConnector := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 40; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune == '├' || cell.Rune == '└' || cell.Rune == '─' || cell.Rune == '│' {
				hasConnector = true
			}
		}
	}
	if !hasConnector {
		t.Error("Draw buffer should contain tree-drawing characters (├, └, ─, or │)")
	}
}

// TestDrawRendersNodeText verifies node text appears in the buffer.
// Spec: "Each visible row draws a graph prefix + node text"
func TestDrawRendersNodeText(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree()
	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	// Check that root text "root" appears somewhere (after the graph prefix)
	foundRoot := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 40-4; x++ {
			if buf.GetCell(x, y).Rune == 'r' &&
				buf.GetCell(x+1, y).Rune == 'o' &&
				buf.GetCell(x+2, y).Rune == 'o' &&
				buf.GetCell(x+3, y).Rune == 't' {
				foundRoot = true
			}
		}
	}
	if !foundRoot {
		t.Error("Draw buffer should contain 'root' text")
	}
}

// TestDrawRendersChildNodeText verifies child node text appears after its prefix.
func TestDrawRendersChildNodeText(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree()
	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	// Check for "child1" text
	foundChild := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 40-6; x++ {
			if buf.GetCell(x, y).Rune == 'c' &&
				buf.GetCell(x+1, y).Rune == 'h' &&
				buf.GetCell(x+2, y).Rune == 'i' &&
				buf.GetCell(x+3, y).Rune == 'l' &&
				buf.GetCell(x+4, y).Rune == 'd' &&
				buf.GetCell(x+5, y).Rune == '1' {
				foundChild = true
			}
		}
	}
	if !foundChild {
		t.Error("Draw buffer should contain 'child1' text")
	}
}

// TestDrawOnlyViewportRows verifies only rows within the viewport (deltaY..deltaY+height-1) are drawn.
// Spec: "Only draws rows within the viewport: rows deltaY through deltaY + bounds.Height() - 1"
func TestDrawOnlyViewportRows(t *testing.T) {
	ov := newOV()
	ov.root = makeSiblingList(20) // 20 siblings
	ov.deltaY = 5                 // should draw rows 5 through 14
	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	// Row 0 should show sibling[5], not sibling[0]
	// Check for "item5" at row 0
	foundItem5 := false
	for x := 0; x < 35; x++ {
		if buf.GetCell(x, 0).Rune == 'i' &&
			buf.GetCell(x+1, 0).Rune == 't' &&
			buf.GetCell(x+2, 0).Rune == 'e' &&
			buf.GetCell(x+3, 0).Rune == 'm' &&
			buf.GetCell(x+4, 0).Rune == '5' {
			foundItem5 = true
		}
	}
	if !foundItem5 {
		t.Error("Row 0 should show 'item5' when deltaY=5 (viewport-relative drawing)")
	}
}

// TestDrawGraphPrefixNonLastRoot verifies the graph prefix for a non-last root
// draws with ├── connector (not └──) and has no ancestor lines.
func TestDrawGraphPrefixNonLastRoot(t *testing.T) {
	// Tree: root -> child2 (siblings). Root is not last (has Next).
	root := NewNode("root", nil, nil)
	child2 := NewNode("child2", nil, nil)
	root.Next = child2
	ov := newOV()
	ov.root = root
	ov.focusedIdx = 1 // focus child2 so root gets normal style
	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	// root at (0,0): should be ├── (not └── since root.Next != nil)
	if buf.GetCell(0, 0).Rune != '├' {
		t.Errorf("root (non-last) at (0,0): got %q, want '├'", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(1, 0).Rune != '─' {
		t.Errorf("root connector at (1,0): got %q, want '─'", buf.GetCell(1, 0).Rune)
	}
	if buf.GetCell(2, 0).Rune != '─' {
		t.Errorf("root connector at (2,0): got %q, want '─'", buf.GetCell(2, 0).Rune)
	}
}

// TestDrawGraphPrefixLastRoot verifies the graph prefix for a last-sibling root
// draws with └── connector.
func TestDrawGraphPrefixLastRoot(t *testing.T) {
	// Single root, no Next -> last sibling
	root := NewNode("lastRoot", nil, nil)
	ov := newOV()
	ov.root = root
	ov.focusedIdx = 99 // focus away from root
	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	// root at (0,0): should be └── (since root.Next == nil)
	if buf.GetCell(0, 0).Rune != '└' {
		t.Errorf("root (last) at (0,0): got %q, want '└'", buf.GetCell(0, 0).Rune)
	}
}

// TestDrawGraphPrefixAncestorLines verifies ancestor vertical lines for nested nodes.
func TestDrawGraphPrefixAncestorLines(t *testing.T) {
	// tree: root -> child1 -> gc1. root.Next != nil (set below), child1.Next=nil
	gc1 := NewNode("gc1", nil, nil)
	child1 := NewNode("child1", gc1, nil)
	root := NewNode("root", child1, nil)
	sibling := NewNode("sibling", nil, nil)
	root.Next = sibling // root has Next -> not last
	ov := newOV()
	ov.root = root
	ov.focusedIdx = 99 // normal style for all

	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	// gc1 is at row 2, level 2
	// Child1 is not last (no Next), but root IS last? No — root.Next=sibling.
	// Level 0 (root): Next != nil -> │ should appear at ancestor pos for child1
	// Level 1 (child1): Next == nil -> spaces at ancestor pos for gc1
	// gc1: child1 is last -> └── connector
	if buf.GetCell(0, 2).Rune != '│' {
		t.Errorf("gc1 ancestor (root level, not last): got %q at (0,2), want '│'", buf.GetCell(0, 2).Rune)
	}
	if buf.GetCell(1, 2).Rune != ' ' {
		t.Errorf("gc1 ancestor spacing at (1,2): got %q, want ' '", buf.GetCell(1, 2).Rune)
	}
	if buf.GetCell(2, 2).Rune != ' ' {
		t.Errorf("gc1 ancestor spacing at (2,2): got %q, want ' '", buf.GetCell(2, 2).Rune)
	}
	// Level 1 ancestor (child1, last sibling): spaces
	if buf.GetCell(3, 2).Rune != ' ' {
		t.Errorf("gc1 ancestor (child1 level, last): got %q at (3,2), want ' '", buf.GetCell(3, 2).Rune)
	}
	if buf.GetCell(4, 2).Rune != ' ' {
		t.Errorf("gc1 ancestor spacing at (4,2): got %q, want ' '", buf.GetCell(4, 2).Rune)
	}
	if buf.GetCell(5, 2).Rune != ' ' {
		t.Errorf("gc1 ancestor spacing at (5,2): got %q, want ' '", buf.GetCell(5, 2).Rune)
	}
	// gc1 connector (last sibling, child1.Next == nil): └──
	if buf.GetCell(6, 2).Rune != '└' {
		t.Errorf("gc1 connector (last) at (6,2): got %q, want '└'", buf.GetCell(6, 2).Rune)
	}
}

// TestDrawGraphPrefixStatusChars verifies status characters: +, ─, space.
func TestDrawGraphPrefixStatusChars(t *testing.T) {
	// collapsedRoot (has children, Expanded=false): status should be '+'
	// expandedRoot (has children, Expanded=true): status should be '─'
	// leaf (no children): status should be ' '
	child1 := NewNode("child", nil, nil)
	child2 := NewNode("child", nil, nil)
	collapsedRoot := NewNode("collapsedRoot", child1, nil)
	expandedRoot := NewNode("expandedRoot", child2, nil)
	leaf := NewNode("leaf", nil, nil)
	collapsedRoot.Expanded = false
	expandedRoot.Expanded = false // collapsed so child2 is hidden, keeping rows predictable
	collapsedRoot.Next = expandedRoot
	expandedRoot.Next = leaf
	ov := newOV()
	ov.root = collapsedRoot
	ov.focusedIdx = 99
	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	// Visible rows: collapsedRoot (row 0), expandedRoot (row 1), leaf (row 2)
	// All are level 0 siblings.
	// Status char is at position 3 (after connector ├── or └──)
	// collapsedRoot (row 0): '+' at pos 3 (has children, collapsed)
	if buf.GetCell(3, 0).Rune != '+' {
		t.Errorf("collapsed root status at (3,0): got %q, want '+'", buf.GetCell(3, 0).Rune)
	}
	// expandedRoot (row 1): '+' at pos 3 (has children, collapsed)
	if buf.GetCell(3, 1).Rune != '+' {
		t.Errorf("collapsed expandedRoot status at (3,1): got %q, want '+'", buf.GetCell(3, 1).Rune)
	}
	// leaf (row 2): ' ' at pos 3 (no children)
	if buf.GetCell(3, 2).Rune != ' ' {
		t.Errorf("leaf status at (3,2): got %q, want ' '", buf.GetCell(3, 2).Rune)
	}

	// Now expand expandedRoot to verify '─' status char
	expandedRoot.Expanded = true
	buf2 := NewDrawBuffer(40, 10)
	ov.Draw(buf2)
	// Visible rows: collapsedRoot (0), expandedRoot (1), child (2), leaf (3)
	if buf2.GetCell(3, 1).Rune != '─' {
		t.Errorf("expanded root status at (3,1): got %q, want '─'", buf2.GetCell(3, 1).Rune)
	}
}

// TestDrawFocusedRowStyle verifies the focused row uses OutlineFocused style (when SfSelected).
// Spec: "Focused row (flatIdx == focusedIdx): uses scheme.OutlineFocused style"
func TestDrawFocusedRowStyle(t *testing.T) {
	ov := newOVFocused()
	ov.root = makeSimpleTree() // 4 visible, focusedIdx=0
	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	// Row 0 should use OutlineFocused style (at least for the text portion)
	focusedStyle := theme.BorlandBlue.OutlineFocused
	foundFocusedStyle := false
	for x := 0; x < 40; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Style == focusedStyle && cell.Rune != ' ' {
			foundFocusedStyle = true
			break
		}
	}
	if !foundFocusedStyle {
		t.Error("Focused row (0) should use OutlineFocused style")
	}
}

// TestDrawCollapsedNodeStyle verifies collapsed nodes use OutlineCollapsed style.
// Spec: "Collapsed node with children (node.Children != nil && !node.Expanded, not focused): scheme.OutlineCollapsed"
func TestDrawCollapsedNodeStyle(t *testing.T) {
	// Build: collapsedParent (collapsed, has children) -> leaf (focused)
	// Both top-level siblings. collapsedParent at idx 0, leaf at idx 1.
	// FocusedIdx=1 means collapsedParent gets OutlineCollapsed, not OutlineFocused.
	child := NewNode("child", nil, nil)
	collapsedParent := NewNode("collapsedParent", child, nil)
	collapsedParent.Expanded = false
	leaf := NewNode("leaf", nil, nil)
	collapsedParent.Next = leaf

	ov := newOV() // not focused
	ov.root = collapsedParent
	ov.focusedIdx = 1 // focus leaf (idx 1), collapsedParent at idx 0 is not focused

	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	collapsedStyle := theme.BorlandBlue.OutlineCollapsed
	// collapsedParent on row 0 should use OutlineCollapsed style
	foundCollapsedStyle := false
	for x := 0; x < 40; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Style == collapsedStyle && cell.Rune != ' ' {
			foundCollapsedStyle = true
			break
		}
	}
	if !foundCollapsedStyle {
		t.Error("Collapsed node on row 0 should use OutlineCollapsed style")
	}
}

// TestDrawNormalNodeStyle verifies normal (unfocused, non-collapsed) nodes use OutlineNormal style.
// Spec: "Normal: scheme.OutlineNormal"
func TestDrawNormalNodeStyle(t *testing.T) {
	ov := newOV() // not focused
	root := makeSimpleTree()
	// root is expanded, focusedIdx=0 (focused row is root)
	// Row 1 (child1) is not focused, expanded (not collapsed)
	ov.root = root
	ov.focusedIdx = 0

	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	normalStyle := theme.BorlandBlue.OutlineNormal
	// Row 1 (child1) should use OutlineNormal
	foundNormalStyle := false
	for x := 0; x < 40; x++ {
		cell := buf.GetCell(x, 1)
		if cell.Style == normalStyle && cell.Rune != ' ' {
			foundNormalStyle = true
			break
		}
	}
	if !foundNormalStyle {
		t.Error("Normal expanded node on row 1 should use OutlineNormal style")
	}
}

// TestDrawUsesWriteChar verifies drawing uses WriteChar (checking by drawing and reading back).
// Spec: "IMPORTANT: Uses buf.WriteChar(x, y, ch, style) to draw each character -- NOT SetCell"
// We verify by the fact that cells have the correct runes and styles after Draw.
func TestDrawUsesWriteChar(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree()
	buf := NewDrawBuffer(40, 10)
	ov.Draw(buf)

	// If SetCell is used (which sets Style but not Rune), the cell would have rune=0
	// WriteChar sets both Rune and Style. Verify cells have non-zero runes.
	foundNonZero := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 40; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != 0 && cell.Rune != ' ' {
				foundNonZero = true
			}
		}
	}
	if !foundNonZero {
		t.Error("Draw should use WriteChar (cells should have non-zero runes)")
	}
}

// TestDrawEmptyTreeNilRoot verifies Draw does not panic with nil root.
// Spec: "An empty tree (nil root) returns 0"
func TestDrawEmptyTreeNilRoot(t *testing.T) {
	ov := newOV()
	buf := NewDrawBuffer(40, 10)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Draw with nil root panicked: %v", r)
		}
	}()

	ov.Draw(buf)
	// No assertion needed beyond "does not panic"
}

// ---------------------------------------------------------------------------
// Section 8 — adjust / adjustAll / selected placeholders
// ---------------------------------------------------------------------------

// TestAdjustTogglesExpanded verifies adjust() toggles the focused node's Expanded flag.
// Spec: "adjust() — finds focused node, toggles Expanded"
func TestAdjustTogglesExpanded(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // root has children, Expanded=true, focusedIdx=0
	ov.adjust()

	if ov.root.Expanded != false {
		t.Errorf("after adjust(): root.Expanded = %v, want false (toggled)", ov.root.Expanded)
	}
}

// TestAdjustOnLeafTogglesExpandedButHasNoVisibleEffect verifies adjust() toggles Expanded
// even on leaf nodes (no children), but the toggle has no structural effect.
// Spec: "adjust() — finds focused node, toggles Expanded" (always toggles)
func TestAdjustOnLeafTogglesExpandedButHasNoVisibleEffect(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree()
	// Move focus to gc1 (idx 2, leaf, no children)
	ov.focusedIdx = 2
	gc1 := ov.root.Children.Children
	prevExpanded := gc1.Expanded

	ov.adjust()

	// adjust() always toggles, even on leaves
	if gc1.Expanded == prevExpanded {
		t.Errorf("adjust() should toggle Expanded even on leaf nodes; was %v, still %v", prevExpanded, gc1.Expanded)
	}
	// Since gc1 has no children, toggling has no effect on visibleCount
}

// TestAdjustCallsSyncScrollBars verifies adjust() calls syncScrollBars.
// Spec: "calls syncScrollBars and ensureVisible"
func TestAdjustCallsSyncScrollBars(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // 4 visible
	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	ov.SetVScrollBar(sb)

	// Adjust toggles root.Expanded from true to false -> visibleCount goes from 4 to 1
	ov.adjust()

	// syncScrollBars should have updated scrollbar range
	if sb.Max() != 1 {
		t.Errorf("after adjust(): ScrollBar Max() = %d, want 1 (visibleCount after collapse)", sb.Max())
	}
}

// TestAdjustCallsEnsureVisible verifies adjust() calls ensureVisible.
// Spec: "calls syncScrollBars and ensureVisible"
func TestAdjustCallsEnsureVisible(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree() // 4 visible
	ov.deltaY = 3 // scroll past the visible range after collapse? Let's check.

	// Collapse root: visibleCount goes from 4 to 1
	// focusedIdx=0, deltaY=3 -> ensureVisible should set deltaY=0 (focusedIdx < deltaY)
	ov.adjust()

	if ov.deltaY != 0 {
		t.Errorf("after adjust() collapses root: deltaY = %d, want 0 (focusedIdx < deltaY)", ov.deltaY)
	}
}

// TestAdjustAllExpandsDescendants verifies adjustAll() expands the focused node and all descendants.
// Spec: "adjustAll() — expands focused node and all descendants recursively"
func TestAdjustAllExpandsDescendants(t *testing.T) {
	ov := newOV()
	ov.root = makeDeepTree() // root -> child1 -> gc1
	// Collapse everything
	ov.root.Expanded = false
	ov.root.Children.Expanded = false

	ov.adjustAll()

	if !ov.root.Expanded {
		t.Error("after adjustAll(): root.Expanded should be true")
	}
	if !ov.root.Children.Expanded {
		t.Error("after adjustAll(): child1.Expanded should be true")
	}
	// gc1 is a leaf (no children), so its Expanded is not critical
	// But if it were collapsible, it would also be expanded
}

// TestAdjustAllOnEmptyTree verifies adjustAll() does not panic when root is nil.
func TestAdjustAllOnEmptyTree(t *testing.T) {
	ov := newOV()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("adjustAll() on empty tree panicked: %v", r)
		}
	}()
	ov.adjustAll()
}

// TestAdjustAllCallsSyncScrollBars verifies adjustAll() calls syncScrollBars.
// Spec: "calls syncScrollBars and ensureVisible"
func TestAdjustAllCallsSyncScrollBars(t *testing.T) {
	ov := newOV()
	ov.root = makeDeepTree() // 3 visible when expanded
	ov.root.Expanded = false
	ov.root.Children.Expanded = false
	// Currently visibleCount = 1 (only root visible, collapsed)
	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	ov.SetVScrollBar(sb)

	ov.adjustAll() // expands all -> visibleCount = 3

	if sb.Max() != 3 {
		t.Errorf("after adjustAll(): ScrollBar Max() = %d, want 3 (all expanded)", sb.Max())
	}
}

// TestSelectedIsNoOp verifies selected() is a no-op placeholder.
// Spec: "selected() — does nothing (placeholder)"
func TestSelectedIsNoOp(t *testing.T) {
	ov := newOV()
	ov.root = makeSimpleTree()
	// Just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("selected() panicked: %v", r)
		}
	}()
	ov.selected()
}

// ---------------------------------------------------------------------------
// Helpers for building test trees
// ---------------------------------------------------------------------------

// makeSiblingList creates a flat list of n sibling nodes.
// Item text is "item0", "item1", ..., "itemN-1".
func makeSiblingList(n int) *TNode {
	if n == 0 {
		return nil
	}
	nodes := make([]*TNode, n)
	for i := 0; i < n; i++ {
		text := "item" + string(rune('0'+i%10))
		if i >= 10 {
			text = "item" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		}
		nodes[i] = NewNode(text, nil, nil)
	}
	for i := 0; i < n-1; i++ {
		nodes[i].Next = nodes[i+1]
	}
	return nodes[0]
}
