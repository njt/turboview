package tv

// outline_concrete_test.go — Tests for Task 3: Outline concrete class (Batch 7).
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec requirement it verifies.
//
// Test organization:
//   Section 1  — Construction (NewOutline)
//   Section 2  — Root management (Root, SetRoot)
//   Section 3  — OnSelect callback
//   Section 4  — Update method
//   Section 5  — ForEach visitor
//   Section 6  — FirstThat visitor
//   Section 7  — adjust override (toggle Expanded, call Update)
//   Section 8  — adjustAll override (expand descendants, call Update)
//   Section 9  — selected override (calls OnSelect)
//   Section 10 — Widget interface
//   Section 11 — Integration tests

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// compile-time assertion: Outline must satisfy Widget.
var _ Widget = (*Outline)(nil)

// ---------------------------------------------------------------------------
// Test helpers for Outline
// ---------------------------------------------------------------------------

// newOutline creates an Outline with bounds 40x10 and BorlandBlue scheme.
func newOutline(root *TNode) *Outline {
	o := NewOutline(NewRect(0, 0, 40, 10), root)
	o.BaseView.scheme = theme.BorlandBlue
	return o
}

// newOutlineNil creates an Outline with nil root.
func newOutlineNil() *Outline {
	o := NewOutline(NewRect(0, 0, 40, 10), nil)
	o.BaseView.scheme = theme.BorlandBlue
	return o
}

// buildTestTree creates a tree: root -> child1 -> gc1, root.child1 -> child2
// Expanded true by default.
func buildTestTree() *TNode {
	gc1 := NewNode("gc1", nil, nil)
	child1 := NewNode("child1", gc1, nil)
	child2 := NewNode("child2", nil, nil)
	child1.Next = child2
	root := NewNode("root", child1, nil)
	return root
}

// buildFlatTree creates three top-level siblings: A -> B -> C.
func buildFlatTree() *TNode {
	a := NewNode("A", nil, nil)
	b := NewNode("B", nil, nil)
	c := NewNode("C", nil, nil)
	a.Next = b
	b.Next = c
	return a
}

// buildDeepTree creates: root -> child1 -> gc1 -> ggc1
// All expanded true by default.
func buildDeepTree() *TNode {
	ggc1 := NewNode("ggc1", nil, nil)
	gc1 := NewNode("gc1", ggc1, nil)
	child1 := NewNode("child1", gc1, nil)
	root := NewNode("root", child1, nil)
	return root
}

// ---------------------------------------------------------------------------
// Section 1 — Construction (NewOutline)
// ---------------------------------------------------------------------------

// TestNewOutlineCreatesOutlineViewer verifies NewOutline creates an Outline
// with an embedded OutlineViewer.
// Spec: "Embeds OutlineViewer (not pointer — value embedding so Outline IS an OutlineViewer)"
func TestNewOutlineCreatesOutlineViewer(t *testing.T) {
	root := buildTestTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Outline should be a Widget
	if _, ok := interface{}(o).(Widget); !ok {
		t.Error("NewOutline should create a type satisfying Widget")
	}

	// Verify value embedding: OutlineViewer field is accessible
	_ = o.OutlineViewer // compile-time proof of embedding
}

// TestNewOutlineSetsRoot verifies NewOutline sets the given root.
// Spec: "NewOutline(bounds Rect, root *TNode) *Outline — creates OutlineViewer, sets root"
func TestNewOutlineSetsRoot(t *testing.T) {
	root := buildTestTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	if o.Root() != root {
		t.Errorf("NewOutline: Root() = %p, want %p", o.Root(), root)
	}
}

// TestNewOutlineCallsUpdate verifies NewOutline calls Update().
// Spec: "NewOutline(bounds Rect, root *TNode) *Outline — calls Update()"
func TestNewOutlineCallsUpdate(t *testing.T) {
	root := buildTestTree() // 4 visible nodes
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// After Update(), focusedIdx should be clamped and scrollbar synced.
	// With a fresh outline and 4 visible nodes, focusedIdx=0 is valid.
	if o.focusedIdx != 0 {
		t.Errorf("NewOutline: focusedIdx = %d, want 0 after Update()", o.focusedIdx)
	}
}

// TestNewOutlineReturnsPointer verifies NewOutline returns *Outline.
// Spec: "Returns *Outline"
func TestNewOutlineReturnsPointer(t *testing.T) {
	root := buildTestTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	if o == nil {
		t.Error("NewOutline should return non-nil *Outline")
	}
}

// TestNewOutlineWithNilRoot verifies NewOutline works with nil root.
// Spec: implied — should handle nil gracefully
func TestNewOutlineWithNilRoot(t *testing.T) {
	o := NewOutline(NewRect(0, 0, 40, 10), nil)

	if o == nil {
		t.Error("NewOutline with nil root should return non-nil *Outline")
	}
	if o.Root() != nil {
		t.Error("NewOutline with nil root: Root() should be nil")
	}
}

// TestNewOutlineSetsUpBounds verifies NewOutline sets bounds correctly.
// Spec: implied — OutlineViewer inherits from BaseView
func TestNewOutlineSetsUpBounds(t *testing.T) {
	r := NewRect(2, 3, 40, 10)
	o := NewOutline(r, nil)

	if o.Bounds() != r {
		t.Errorf("NewOutline: Bounds() = %v, want %v", o.Bounds(), r)
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Root management (Root, SetRoot)
// ---------------------------------------------------------------------------

// TestRootReturnsRoot verifies Root() returns the root.
// Spec: "Root() *TNode — returns the root"
func TestRootReturnsRoot(t *testing.T) {
	root := buildTestTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	if o.Root() != root {
		t.Errorf("Root() = %p, want %p", o.Root(), root)
	}
}

// TestRootReturnsNilWhenNull verifies Root() returns nil if root is nil.
// Spec: "Root() *TNode — returns the root"
func TestRootReturnsNilWhenNull(t *testing.T) {
	o := NewOutline(NewRect(0, 0, 40, 10), nil)

	if o.Root() != nil {
		t.Error("Root() should return nil when root is nil")
	}
}

// TestSetRootReplacesRoot verifies SetRoot() replaces the root.
// Spec: "SetRoot(root *TNode) — sets root, resets focusedIdx=0, deltaY=0, calls Update()"
func TestSetRootReplacesRoot(t *testing.T) {
	oldRoot := buildTestTree()
	newRoot := buildFlatTree()
	o := NewOutline(NewRect(0, 0, 40, 10), oldRoot)

	o.SetRoot(newRoot)

	if o.Root() != newRoot {
		t.Errorf("SetRoot: Root() = %p, want %p", o.Root(), newRoot)
	}
}

// TestSetRootResetsFocusedIdx verifies SetRoot() resets focusedIdx to 0.
// Spec: "SetRoot(root *TNode) — resets focusedIdx=0"
func TestSetRootResetsFocusedIdx(t *testing.T) {
	oldRoot := buildTestTree() // 4 visible
	newRoot := buildFlatTree() // 3 visible
	o := NewOutline(NewRect(0, 0, 40, 10), oldRoot)

	// Move focus away from 0
	o.focusedIdx = 2

	// SetRoot should reset focusedIdx to 0
	o.SetRoot(newRoot)

	if o.focusedIdx != 0 {
		t.Errorf("SetRoot: focusedIdx = %d, want 0 (reset)", o.focusedIdx)
	}
}

// TestSetRootResetsDeltaY verifies SetRoot() resets deltaY to 0.
// Spec: "SetRoot(root *TNode) — resets deltaY=0"
func TestSetRootResetsDeltaY(t *testing.T) {
	oldRoot := buildTestTree()
	newRoot := buildFlatTree()
	o := NewOutline(NewRect(0, 0, 40, 10), oldRoot)

	// Set deltaY to non-zero
	o.deltaY = 5

	// SetRoot should reset deltaY to 0
	o.SetRoot(newRoot)

	if o.deltaY != 0 {
		t.Errorf("SetRoot: deltaY = %d, want 0 (reset)", o.deltaY)
	}
}

// TestSetRootCallsUpdate verifies SetRoot() calls Update().
// Spec: "SetRoot(root *TNode) — calls Update()"
func TestSetRootCallsUpdate(t *testing.T) {
	oldRoot := buildTestTree() // 4 visible
	newRoot := buildFlatTree() // 3 visible
	o := NewOutline(NewRect(0, 0, 40, 10), oldRoot)

	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	o.SetVScrollBar(sb)

	// SetRoot should call Update, which syncs scrollbar
	o.SetRoot(newRoot)

	// After Update, scrollbar should reflect the new tree (3 visible)
	if sb.Max() != 3 {
		t.Errorf("SetRoot: ScrollBar Max() = %d, want 3 (new tree size)", sb.Max())
	}
}

// TestSetRootWithNil verifies SetRoot() handles nil gracefully.
// Spec: implied — should not panic
func TestSetRootWithNil(t *testing.T) {
	o := NewOutline(NewRect(0, 0, 40, 10), buildTestTree())

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetRoot(nil) panicked: %v", r)
		}
	}()

	o.SetRoot(nil)

	if o.Root() != nil {
		t.Error("SetRoot(nil): Root() should be nil")
	}
}

// TestSetRootMultipleTimes verifies SetRoot() can be called multiple times.
// Spec: implied — independent state management
func TestSetRootMultipleTimes(t *testing.T) {
	root1 := buildTestTree()
	root2 := buildFlatTree()
	root3 := buildDeepTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root1)

	o.SetRoot(root2)
	if o.Root() != root2 {
		t.Errorf("SetRoot(root2): Root() = %p, want %p", o.Root(), root2)
	}

	o.SetRoot(root3)
	if o.Root() != root3 {
		t.Errorf("SetRoot(root3): Root() = %p, want %p", o.Root(), root3)
	}
}

// ---------------------------------------------------------------------------
// Section 3 — OnSelect callback
// ---------------------------------------------------------------------------

// TestOnSelectField verifies Outline has an OnSelect field.
// Spec: "Has OnSelect func(node *TNode) field"
func TestOnSelectField(t *testing.T) {
	o := newOutline(buildTestTree())

	// OnSelect should be accessible and callable
	if o.OnSelect != nil {
		t.Error("OnSelect should initially be nil (not set)")
	}
}

// TestSetOnSelectStoresCallback verifies SetOnSelect() stores the callback.
// Spec: "SetOnSelect(fn func(node *TNode)) — sets o.OnSelect = fn"
func TestSetOnSelectStoresCallback(t *testing.T) {
	o := newOutline(buildTestTree())

	called := false
	var callNode *TNode
	cb := func(node *TNode) {
		called = true
		callNode = node
	}

	o.SetOnSelect(cb)

	// Verify callback is stored
	if o.OnSelect == nil {
		t.Fatal("SetOnSelect: OnSelect should be set")
	}

	// Call the callback to verify it's the same function
	root := o.Root()
	o.OnSelect(root)

	if !called {
		t.Error("SetOnSelect: callback was not called")
	}
	if callNode != root {
		t.Errorf("SetOnSelect: callback received %p, want %p", callNode, root)
	}
}

// TestSetOnSelectWithNil verifies SetOnSelect(nil) clears the callback.
// Spec: implied — should allow nil
func TestSetOnSelectWithNil(t *testing.T) {
	o := newOutline(buildTestTree())

	cb := func(node *TNode) {}
	o.SetOnSelect(cb)

	if o.OnSelect == nil {
		t.Fatal("SetOnSelect: callback should be set")
	}

	o.SetOnSelect(nil)

	if o.OnSelect != nil {
		t.Error("SetOnSelect(nil): OnSelect should be nil")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Update method
// ---------------------------------------------------------------------------

// TestUpdateRecomputesVisibleCount verifies Update() recomputes visible count.
// Spec: "Update() — recomputes visible count and syncs scrollbar"
func TestUpdateRecomputesVisibleCount(t *testing.T) {
	root := buildTestTree() // 4 visible when expanded
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	o.SetVScrollBar(sb)

	// Collapse root
	root.Expanded = false

	// visibleCount is now 1, but scrollbar still shows 4
	if sb.Max() != 4 {
		t.Fatalf("setup: ScrollBar Max() = %d, want 4", sb.Max())
	}

	// Call Update to recompute
	o.Update()

	// Now scrollbar should be updated
	if sb.Max() != 1 {
		t.Errorf("Update: ScrollBar Max() = %d, want 1 (recomputed)", sb.Max())
	}
}

// TestUpdateClampsExcessiveFocusedIdx verifies Update() clamps focusedIdx if >= visibleCount().
// Spec: "If focusedIdx is now >= visibleCount(), clamps it to visibleCount()-1"
func TestUpdateClampsExcessiveFocusedIdx(t *testing.T) {
	root := buildTestTree() // 4 visible
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Set focusedIdx to 3 (valid)
	o.focusedIdx = 3

	// Collapse root: visibleCount becomes 1
	root.Expanded = false

	// focusedIdx=3 is now out of bounds (>= 1)
	if o.focusedIdx < 1 {
		t.Fatalf("setup: focusedIdx should be out of bounds after collapse")
	}

	// Update should clamp focusedIdx to 0
	o.Update()

	if o.focusedIdx != 0 {
		t.Errorf("Update: focusedIdx = %d, want 0 (clamped to visibleCount()-1)", o.focusedIdx)
	}
}

// TestUpdateCallsEnsureVisible verifies Update() calls ensureVisible().
// Spec: "Calls ensureVisible()"
func TestUpdateCallsEnsureVisible(t *testing.T) {
	root := buildTestTree() // 4 visible
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Set focusedIdx=3, deltaY=5 (inconsistent)
	o.focusedIdx = 0
	o.deltaY = 5

	// Update should call ensureVisible, which adjusts deltaY
	o.Update()

	// ensureVisible should have set deltaY to 0 (since focusedIdx < deltaY)
	if o.deltaY != 0 {
		t.Errorf("Update: deltaY = %d, want 0 (ensureVisible adjustment)", o.deltaY)
	}
}

// TestUpdateCallsSyncScrollBar verifies Update() calls syncScrollBars.
// Spec: "recomputes visible count and syncs scrollbar"
func TestUpdateCallsSyncScrollBar(t *testing.T) {
	root := buildTestTree() // 4 visible
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	o.SetVScrollBar(sb)

	// Manually change focusedIdx without sync
	o.focusedIdx = 2
	// Don't call syncScrollBars, so scrollbar value is stale

	// Call Update to re-sync
	o.Update()

	// Scrollbar value should match focusedIdx (via deltaY ensured by ensureVisible)
	if sb.Value() != o.deltaY {
		t.Errorf("Update: ScrollBar Value() = %d, want %d (deltaY)", sb.Value(), o.deltaY)
	}
}

// TestUpdateWithEmptyTree verifies Update() handles nil root.
// Spec: implied — should not panic
func TestUpdateWithEmptyTree(t *testing.T) {
	o := NewOutline(NewRect(0, 0, 40, 10), nil)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update() on empty tree panicked: %v", r)
		}
	}()

	o.Update()

	if o.focusedIdx != 0 {
		t.Errorf("Update on empty tree: focusedIdx = %d, want 0", o.focusedIdx)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — ForEach visitor
// ---------------------------------------------------------------------------

// TestForEachVisitsAllNodes verifies ForEach visits ALL nodes (including collapsed).
// Spec: "ForEach(fn func(node *TNode, level int)) — depth-first traversal of ALL nodes (not just visible)"
func TestForEachVisitsAllNodes(t *testing.T) {
	root := buildTestTree() // 4 visible: root, child1, gc1, child2
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Collapse child1: gc1 becomes invisible in the outline
	root.Children.Expanded = false

	visited := make([]*TNode, 0)
	o.ForEach(func(node *TNode, level int) {
		visited = append(visited, node)
	})

	// ForEach should visit ALL nodes: root, child1, gc1, child2
	if len(visited) != 4 {
		t.Errorf("ForEach: visited %d nodes, want 4 (all nodes including collapsed)", len(visited))
	}

	// Verify order: depth-first (root, child1, gc1, then child2)
	if visited[0].Text != "root" {
		t.Errorf("ForEach[0]: %q, want 'root'", visited[0].Text)
	}
	if visited[1].Text != "child1" {
		t.Errorf("ForEach[1]: %q, want 'child1'", visited[1].Text)
	}
	if visited[2].Text != "gc1" {
		t.Errorf("ForEach[2]: %q, want 'gc1'", visited[2].Text)
	}
	if visited[3].Text != "child2" {
		t.Errorf("ForEach[3]: %q, want 'child2'", visited[3].Text)
	}
}

// TestForEachLevelCorrect verifies ForEach passes correct level values.
// Spec: "Level starts at 0"
func TestForEachLevelCorrect(t *testing.T) {
	root := buildTestTree() // root, child1, gc1, child2
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	levels := make(map[string]int)
	o.ForEach(func(node *TNode, level int) {
		levels[node.Text] = level
	})

	if levels["root"] != 0 {
		t.Errorf("ForEach: root level = %d, want 0", levels["root"])
	}
	if levels["child1"] != 1 {
		t.Errorf("ForEach: child1 level = %d, want 1", levels["child1"])
	}
	if levels["gc1"] != 2 {
		t.Errorf("ForEach: gc1 level = %d, want 2", levels["gc1"])
	}
	if levels["child2"] != 1 {
		t.Errorf("ForEach: child2 level = %d, want 1", levels["child2"])
	}
}

// TestForEachDepthFirst verifies ForEach uses depth-first traversal.
// Spec: "Visits: node, then Children (if any), then Next (if any)"
func TestForEachDepthFirst(t *testing.T) {
	// Build: A has child B, B has children C and D (C.Next = D), A has next E
	// Expected depth-first order: A, B, C, D, E with levels: A=0, B=1, C=2, D=2, E=0
	d := NewNode("D", nil, nil)
	c := NewNode("C", nil, d)    // C.Next = D (siblings at level 2)
	b := NewNode("B", c, nil)     // B.Children = C
	e := NewNode("E", nil, nil)
	a := NewNode("A", b, e)       // A.Children = B, A.Next = E

	o := NewOutline(NewRect(0, 0, 40, 10), a)

	visited := make([]string, 0)
	levels := make([]int, 0)
	o.ForEach(func(node *TNode, level int) {
		visited = append(visited, node.Text)
		levels = append(levels, level)
	})

	expected := []string{"A", "B", "C", "D", "E"}
	expectedLevels := []int{0, 1, 2, 2, 0}
	if len(visited) != len(expected) {
		t.Errorf("ForEach: visited %d nodes, want %d", len(visited), len(expected))
	}
	for i, exp := range expected {
		if i >= len(visited) {
			t.Fatalf("ForEach: not enough nodes visited (want %d)", len(expected))
		}
		if visited[i] != exp {
			t.Errorf("ForEach[%d]: %q, want %q", i, visited[i], exp)
		}
		if levels[i] != expectedLevels[i] {
			t.Errorf("ForEach[%d] level: %d, want %d", i, levels[i], expectedLevels[i])
		}
	}
}

// TestForEachEmptyTree verifies ForEach on nil root does not panic.
// Spec: implied — should handle nil root
func TestForEachEmptyTree(t *testing.T) {
	o := NewOutline(NewRect(0, 0, 40, 10), nil)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ForEach on empty tree panicked: %v", r)
		}
	}()

	count := 0
	o.ForEach(func(node *TNode, level int) {
		count++
	})

	if count != 0 {
		t.Errorf("ForEach on empty tree: visited %d nodes, want 0", count)
	}
}

// TestForEachCollapsedNodesIncluded verifies ForEach visits collapsed children.
// Spec: "depth-first traversal of ALL nodes (not just visible). Collapsed nodes' children are visited."
func TestForEachCollapsedNodesIncluded(t *testing.T) {
	root := buildDeepTree() // root -> child1 -> gc1 -> ggc1
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Collapse child1
	root.Children.Expanded = false

	visited := make([]string, 0)
	o.ForEach(func(node *TNode, level int) {
		visited = append(visited, node.Text)
	})

	// Should visit: root, child1, gc1, ggc1 (not just root, child1)
	if len(visited) != 4 {
		t.Errorf("ForEach: visited %d nodes, want 4 (all, including collapsed)", len(visited))
	}

	expectedOrder := []string{"root", "child1", "gc1", "ggc1"}
	for i, exp := range expectedOrder {
		if visited[i] != exp {
			t.Errorf("ForEach[%d]: %q, want %q", i, visited[i], exp)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 6 — FirstThat visitor
// ---------------------------------------------------------------------------

// TestFirstThatFindsNode verifies FirstThat returns the first matching node.
// Spec: "FirstThat(fn func(node *TNode, level int) bool) *TNode — same traversal, returns first node where fn returns true, or nil"
func TestFirstThatFindsNode(t *testing.T) {
	root := buildTestTree() // root, child1, gc1, child2
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	found := o.FirstThat(func(node *TNode, level int) bool {
		return node.Text == "gc1"
	})

	if found == nil {
		t.Fatal("FirstThat: should find 'gc1' node")
	}
	if found.Text != "gc1" {
		t.Errorf("FirstThat: found %q, want 'gc1'", found.Text)
	}
}

// TestFirstThatReturnsNilWhenNotFound verifies FirstThat returns nil if no match.
// Spec: "returns first node where fn returns true, or nil"
func TestFirstThatReturnsNilWhenNotFound(t *testing.T) {
	root := buildTestTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	found := o.FirstThat(func(node *TNode, level int) bool {
		return node.Text == "nonexistent"
	})

	if found != nil {
		t.Errorf("FirstThat: found %p, want nil (no match)", found)
	}
}

// TestFirstThatReturnsFirstMatch verifies FirstThat returns the FIRST matching node.
// Spec: "returns first node where fn returns true"
func TestFirstThatReturnsFirstMatch(t *testing.T) {
	// Build two nodes with same text to verify first is returned
	a1 := NewNode("dup", nil, nil)
	a2 := NewNode("dup", nil, nil)
	a1.Next = a2
	o := NewOutline(NewRect(0, 0, 40, 10), a1)

	found := o.FirstThat(func(node *TNode, level int) bool {
		return node.Text == "dup"
	})

	// Should return the first "dup" node
	if found != a1 {
		t.Errorf("FirstThat: found %p, want %p (first match)", found, a1)
	}
}

// TestFirstThatPassesLevelCorrectly verifies FirstThat passes correct level to predicate.
// Spec: "same traversal, ... level int"
func TestFirstThatPassesLevelCorrectly(t *testing.T) {
	root := buildDeepTree() // root(0) -> child1(1) -> gc1(2) -> ggc1(3)
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	var foundLevel int
	o.FirstThat(func(node *TNode, level int) bool {
		if node.Text == "ggc1" {
			foundLevel = level
			return true
		}
		return false
	})

	if foundLevel != 3 {
		t.Errorf("FirstThat: ggc1 level = %d, want 3", foundLevel)
	}
}

// TestFirstThatEmptyTree verifies FirstThat on nil root returns nil.
// Spec: implied — should handle nil root
func TestFirstThatEmptyTree(t *testing.T) {
	o := NewOutline(NewRect(0, 0, 40, 10), nil)

	found := o.FirstThat(func(node *TNode, level int) bool {
		return true
	})

	if found != nil {
		t.Errorf("FirstThat on empty tree: found %p, want nil", found)
	}
}

// TestFirstThatDepthFirst verifies FirstThat uses depth-first order.
// Spec: "same traversal"
func TestFirstThatDepthFirst(t *testing.T) {
	// Build: root -> child1 -> gc1, child2
	// Depth-first order: root, child1, gc1, child2
	// If we search for anything after root, we should get child1 (not child2)
	gc1 := NewNode("gc1", nil, nil)
	child1 := NewNode("child1", gc1, nil)
	child2 := NewNode("child2", nil, nil)
	child1.Next = child2
	root := NewNode("root", child1, nil)
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Find the first node after root (should be child1, not child2)
	var firstAfterRoot *TNode
	o.FirstThat(func(node *TNode, level int) bool {
		if node.Text != "root" {
			firstAfterRoot = node
			return true
		}
		return false
	})

	if firstAfterRoot == nil || firstAfterRoot.Text != "child1" {
		if firstAfterRoot != nil {
			t.Errorf("FirstThat: first after root = %q, want 'child1'", firstAfterRoot.Text)
		} else {
			t.Error("FirstThat: no node after root found")
		}
	}
}

// TestFirstThatCollapsedNodesIncluded verifies FirstThat searches all nodes including collapsed.
// Spec: "depth-first traversal of ALL nodes"
func TestFirstThatCollapsedNodesIncluded(t *testing.T) {
	root := buildTestTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Collapse root: gc1 is no longer visible in outline
	root.Expanded = false

	found := o.FirstThat(func(node *TNode, level int) bool {
		return node.Text == "gc1"
	})

	if found == nil {
		t.Fatal("FirstThat: should find 'gc1' even though collapsed")
	}
	if found.Text != "gc1" {
		t.Errorf("FirstThat: found %q, want 'gc1'", found.Text)
	}
}

// ---------------------------------------------------------------------------
// Section 7 — adjust override (toggle Expanded, call Update)
// ---------------------------------------------------------------------------

// TestAdjustOverrideTogglesExpanded verifies adjust() toggles Expanded on focused node.
// Spec: "Override adjust(): Toggles Expanded on focused node, calls Update()"
func TestAdjustOverrideTogglesExpanded(t *testing.T) {
	root := buildTestTree() // root expanded
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	o.adjust()

	if root.Expanded != false {
		t.Errorf("adjust(): root.Expanded = %v, want false (toggled)", root.Expanded)
	}
}

// TestAdjustOverrideCallsUpdate verifies adjust() calls Update().
// Spec: "calls Update()"
func TestAdjustOverrideCallsUpdate(t *testing.T) {
	root := buildTestTree() // 4 visible
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	o.SetVScrollBar(sb)

	// Collapse root: visibleCount becomes 1
	o.adjust()

	// Update should have synced scrollbar
	if sb.Max() != 1 {
		t.Errorf("adjust(): ScrollBar Max() = %d, want 1 (Update called)", sb.Max())
	}
}

// TestAdjustOverrideOnLeaf verifies adjust() toggles Expanded even on leaf nodes.
// Spec: implied — same behavior as OutlineViewer.adjust()
func TestAdjustOverrideOnLeaf(t *testing.T) {
	root := buildTestTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Move focus to gc1 (leaf, no children)
	o.focusedIdx = 2
	gc1 := root.Children.Children
	prevExpanded := gc1.Expanded

	o.adjust()

	if gc1.Expanded == prevExpanded {
		t.Errorf("adjust() on leaf: Expanded = %v, still %v (should toggle)", prevExpanded, gc1.Expanded)
	}
}

// ---------------------------------------------------------------------------
// Section 8 — adjustAll override (expand descendants, call Update)
// ---------------------------------------------------------------------------

// TestAdjustAllOverrideExpandsDescendants verifies adjustAll() expands focused node and descendants.
// Spec: "Override adjustAll(): Expands focused node and all its descendants recursively, then calls Update()"
func TestAdjustAllOverrideExpandsDescendants(t *testing.T) {
	root := buildDeepTree() // root -> child1 -> gc1 -> ggc1
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Collapse everything
	root.Expanded = false
	root.Children.Expanded = false

	o.adjustAll()

	if !root.Expanded {
		t.Error("adjustAll(): root.Expanded should be true")
	}
	if !root.Children.Expanded {
		t.Error("adjustAll(): child1.Expanded should be true")
	}
	// gc1 and ggc1 should also be expanded
	if !root.Children.Children.Expanded {
		t.Error("adjustAll(): gc1.Expanded should be true")
	}
	if !root.Children.Children.Children.Expanded {
		t.Error("adjustAll(): ggc1.Expanded should be true")
	}
}

// TestAdjustAllOverrideCallsUpdate verifies adjustAll() calls Update().
// Spec: "then calls Update()"
func TestAdjustAllOverrideCallsUpdate(t *testing.T) {
	root := buildDeepTree() // 4 visible when all expanded
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	o.SetVScrollBar(sb)

	// Collapse everything
	root.Expanded = false
	root.Children.Expanded = false

	o.adjustAll()

	// After adjustAll, everything should be expanded -> scrollbar Max should reflect that
	if sb.Max() != 4 {
		t.Errorf("adjustAll(): ScrollBar Max() = %d, want 4 (fully expanded)", sb.Max())
	}
}

// TestAdjustAllOverrideDoesNotUnexpandOthers verifies adjustAll() only expands the focused node's tree.
// Spec: "expands focused node and all its descendants recursively"
func TestAdjustAllOverrideDoesNotUnexpandOthers(t *testing.T) {
	// Build: root -> (child1, child2)
	// child1 -> gc1, child2 is leaf, collapsed (collapsed node is sibling)
	gc1 := NewNode("gc1", nil, nil)
	child1 := NewNode("child1", gc1, nil)
	child2 := NewNode("child2", nil, nil)
	child1.Next = child2
	root := NewNode("root", child1, nil)
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Collapse child2 (a sibling to child1)
	child2.Expanded = false

	// Focus on child1 and call adjustAll
	o.focusedIdx = 1 // child1
	o.adjustAll()

	// child1 and gc1 should be expanded, child2.Expanded should remain unchanged
	if !child1.Expanded {
		t.Error("adjustAll(): child1 should be expanded")
	}
	if !gc1.Expanded {
		t.Error("adjustAll(): gc1 should be expanded")
	}
	// child2.Expanded should still be false (not a descendant of child1)
	if child2.Expanded != false {
		t.Errorf("adjustAll(): child2.Expanded = %v, want false (sibling, not affected)", child2.Expanded)
	}
}

// ---------------------------------------------------------------------------
// Section 9 — selected override (calls OnSelect)
// ---------------------------------------------------------------------------

// TestSelectedOverrideCallsOnSelect verifies selected() calls OnSelect with focused node.
// Spec: "Override selected(): finds the focused node, calls OnSelect if set"
func TestSelectedOverrideCallsOnSelect(t *testing.T) {
	root := buildTestTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Focus on child1 (idx 1)
	o.focusedIdx = 1

	called := false
	var callNode *TNode
	o.SetOnSelect(func(node *TNode) {
		called = true
		callNode = node
	})

	o.selected()

	if !called {
		t.Fatal("selected(): OnSelect should have been called")
	}

	// The called node should be child1
	if callNode != root.Children {
		t.Errorf("selected(): OnSelect called with %p, want %p (child1)", callNode, root.Children)
	}
}

// TestSelectedOverrideWithoutCallback verifies selected() does not panic if OnSelect is nil.
// Spec: "calls OnSelect if set"
func TestSelectedOverrideWithoutCallback(t *testing.T) {
	root := buildTestTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// OnSelect is nil by default

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("selected() with nil OnSelect panicked: %v", r)
		}
	}()

	o.selected()
}

// TestSelectedOverrideEmptyTree verifies selected() on empty tree does not panic.
// Spec: implied — should handle nil root
func TestSelectedOverrideEmptyTree(t *testing.T) {
	o := NewOutline(NewRect(0, 0, 40, 10), nil)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("selected() on empty tree panicked: %v", r)
		}
	}()

	o.SetOnSelect(func(node *TNode) {
		// This callback receives the focused node (or nothing)
	})

	o.selected()
}

// TestSelectedOverrideCallsCorrectNode verifies selected() calls OnSelect with the correct focused node.
// Spec: "finds the focused node"
func TestSelectedOverrideCallsCorrectNode(t *testing.T) {
	root := buildTestTree() // root, child1, gc1, child2
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Focus on each node and verify selected() calls with the correct node
	testCases := []struct {
		focusedIdx   int
		expectedNode *TNode
	}{
		{0, root},
		{1, root.Children},
		{2, root.Children.Children},
		{3, root.Children.Next},
	}

	for _, tc := range testCases {
		o.focusedIdx = tc.focusedIdx
		var callNode *TNode
		o.SetOnSelect(func(node *TNode) {
			callNode = node
		})

		o.selected()

		if callNode != tc.expectedNode {
			t.Errorf("selected() with focusedIdx=%d: called with %p, want %p",
				tc.focusedIdx, callNode, tc.expectedNode)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 10 — Widget interface
// ---------------------------------------------------------------------------

// TestOutlineIsWidget verifies Outline satisfies the Widget interface.
// Spec: implied — should implement Draw, HandleEvent, Bounds, SetBounds, etc.
func TestOutlineIsWidget(t *testing.T) {
	o := newOutline(buildTestTree())

	// Verify widget methods exist
	if o.Bounds() == (Rect{}) {
		// Zero rect is valid, just verify no panic
	}

	buf := NewDrawBuffer(40, 10)
	o.Draw(buf) // Should not panic

	ev := &Event{What: EvKeyboard}
	o.HandleEvent(ev) // Should not panic
}

// ---------------------------------------------------------------------------
// Section 11 — Integration tests
// ---------------------------------------------------------------------------

// TestOutlineIntegrationTreeNavigation verifies basic tree navigation and selection flow.
// Spec: combined test of Root, Update, ForEach, selected flow
func TestOutlineIntegrationTreeNavigation(t *testing.T) {
	root := buildTestTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Verify initial state
	if o.Root() != root {
		t.Fatal("Initial root should be set")
	}

	// Navigate and set callback
	selectedNodes := make([]*TNode, 0)
	o.SetOnSelect(func(node *TNode) {
		selectedNodes = append(selectedNodes, node)
	})

	// Select root
	o.focusedIdx = 0
	o.selected()

	// Select child1
	o.focusedIdx = 1
	o.selected()

	// Verify selections
	if len(selectedNodes) != 2 {
		t.Fatalf("Expected 2 selections, got %d", len(selectedNodes))
	}
	if selectedNodes[0] != root {
		t.Error("First selection should be root")
	}
	if selectedNodes[1] != root.Children {
		t.Error("Second selection should be child1")
	}
}

// TestOutlineIntegrationCollapseAndExpand verifies collapse/expand affects visibleCount.
// Spec: combined test of Update, adjust/adjustAll, and state management
func TestOutlineIntegrationCollapseAndExpand(t *testing.T) {
	root := buildTestTree() // 4 visible initially
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	initialCount := o.visibleCount()
	if initialCount != 4 {
		t.Fatalf("Initial visibleCount = %d, want 4", initialCount)
	}

	// Collapse root
	o.focusedIdx = 0
	o.adjust()
	if o.visibleCount() != 1 {
		t.Errorf("After collapse: visibleCount = %d, want 1", o.visibleCount())
	}

	// Expand root
	o.adjust()
	if o.visibleCount() != 4 {
		t.Errorf("After expand: visibleCount = %d, want 4", o.visibleCount())
	}
}

// TestOutlineIntegrationSetRootAndUpdate verifies SetRoot resets state properly.
// Spec: combined test of SetRoot, Root, Update
func TestOutlineIntegrationSetRootAndUpdate(t *testing.T) {
	oldRoot := buildTestTree()
	newRoot := buildFlatTree() // Different tree structure

	o := NewOutline(NewRect(0, 0, 40, 10), oldRoot)
	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	o.SetVScrollBar(sb)

	// Initial state
	if o.visibleCount() != 4 {
		t.Fatalf("Old tree visibleCount = %d, want 4", o.visibleCount())
	}

	// SetRoot changes tree and resets state
	o.focusedIdx = 2
	o.deltaY = 3
	o.SetRoot(newRoot)

	// After SetRoot, state should be reset and scrollbar synced
	if o.focusedIdx != 0 {
		t.Errorf("After SetRoot: focusedIdx = %d, want 0", o.focusedIdx)
	}
	if o.deltaY != 0 {
		t.Errorf("After SetRoot: deltaY = %d, want 0", o.deltaY)
	}
	if o.visibleCount() != 3 {
		t.Errorf("After SetRoot: visibleCount = %d, want 3 (new tree)", o.visibleCount())
	}
	if sb.Max() != 3 {
		t.Errorf("After SetRoot: ScrollBar Max() = %d, want 3", sb.Max())
	}
}

// TestOutlineIntegrationForEachAndFirstThat verifies traversal methods work on complex trees.
// Spec: combined test of ForEach, FirstThat with varied tree structures
func TestOutlineIntegrationForEachAndFirstThat(t *testing.T) {
	root := buildDeepTree() // root -> child1 -> gc1 -> ggc1
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// ForEach should visit all 4 nodes
	count := 0
	o.ForEach(func(node *TNode, level int) {
		count++
	})
	if count != 4 {
		t.Errorf("ForEach: visited %d nodes, want 4", count)
	}

	// FirstThat should find specific nodes
	found := o.FirstThat(func(node *TNode, level int) bool {
		return level == 2
	})
	if found == nil || found.Text != "gc1" {
		t.Error("FirstThat: should find gc1 at level 2")
	}

	// FirstThat should return nil for non-existent condition
	found = o.FirstThat(func(node *TNode, level int) bool {
		return level == 99
	})
	if found != nil {
		t.Error("FirstThat: should return nil for non-existent level")
	}
}

// ---------------------------------------------------------------------------
// Section 12 — Dispatch tests (function pointer callbacks)
// ---------------------------------------------------------------------------

// TestEnterKeyFiresOnSelectViaHandleEvent verifies that pressing Enter on an
// Outline dispatches through HandleEvent and fires the OnSelect callback.
func TestEnterKeyFiresOnSelectViaHandleEvent(t *testing.T) {
	root := buildTestTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)
	o.SetState(SfSelected, true)

	var calledWith *TNode
	o.SetOnSelect(func(n *TNode) {
		calledWith = n
	})

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	o.HandleEvent(ev)

	if calledWith == nil {
		t.Fatal("Enter via HandleEvent: OnSelect was not called")
	}
	if calledWith != root {
		t.Errorf("Enter via HandleEvent: OnSelect called with %q, want %q", calledWith.Text, root.Text)
	}
}

// TestPlusKeyUsesOutlineAdjustViaHandleEvent verifies that pressing '+' on an
// Outline dispatches through HandleEvent and uses Outline's adjust (with Update).
func TestPlusKeyUsesOutlineAdjustViaHandleEvent(t *testing.T) {
	root := buildTestTree() // root has children, Expanded=true
	o := NewOutline(NewRect(0, 0, 40, 10), root)
	o.SetState(SfSelected, true)
	sb := NewScrollBar(NewRect(40, 0, 1, 10), Vertical)
	o.SetVScrollBar(sb)

	// Press '+' to toggle expand/collapse via HandleEvent
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: '+'}}
	o.HandleEvent(ev)

	if root.Expanded {
		t.Error("'+' via HandleEvent: root should be collapsed")
	}
	// Verify Update() was called (scrollbar synced)
	if sb.Max() == 4 {
		t.Error("'+' via HandleEvent: scrollbar Max should have changed (Update called)")
	}
}
