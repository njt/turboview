package tv

// integration_outline_test.go — Integration tests for the Outline widget pipeline.
//
// Verifies that TNode -> OutlineViewer -> Outline -> ScrollBar work together
// as a real component chain. Uses REAL components only (no mocks).
//
// Test naming: TestIntegrationOutline<DescriptiveSuffix>

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makeIntegrationTree builds a tree for integration tests:
//
//	Animals (Expanded=true)
//	  Mammals (Expanded=true)
//	    Cat
//	    Dog
//	  Birds (Expanded=true)
//	    Eagle
//	    Parrot
//	Plants (Expanded=true)
//	  Oak
//	  Pine
//
// Visible flattened (all expanded):
//
//	0: Animals      (level 0)
//	1: Mammals      (level 1)
//	2: Cat          (level 2)
//	3: Dog          (level 2)
//	4: Birds        (level 1)
//	5: Eagle        (level 2)
//	6: Parrot       (level 2)
//	7: Plants       (level 0)
//	8: Oak          (level 1)
//	9: Pine         (level 1)
//
// Total visible: 10
func makeIntegrationTree() *TNode {
	cat := NewNode("Cat", nil, nil)
	dog := NewNode("Dog", nil, nil)
	cat.Next = dog

	mammals := NewNode("Mammals", cat, nil)

	eagle := NewNode("Eagle", nil, nil)
	parrot := NewNode("Parrot", nil, nil)
	eagle.Next = parrot

	birds := NewNode("Birds", eagle, nil)
	mammals.Next = birds

	animals := NewNode("Animals", mammals, nil)

	oak := NewNode("Oak", nil, nil)
	pine := NewNode("Pine", nil, nil)
	oak.Next = pine

	plants := NewNode("Plants", oak, nil)

	animals.Next = plants

	return animals
}

// intOutlineKeyEv creates a keyboard event for a special key.
func intOutlineKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModNone}}
}

// intOutlineRuneEv creates a keyboard event for a rune key.
func intOutlineRuneEv(ch rune) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ch, Modifiers: tcell.ModNone}}
}

// newIntOutline creates an Outline with bounds 40x5 (small viewport to test
// scrolling), BorlandBlue color scheme, and SfSelected set.
func newIntOutline(root *TNode) *Outline {
	o := NewOutline(NewRect(0, 0, 40, 5), root)
	o.BaseView.scheme = theme.BorlandBlue
	o.SetState(SfSelected, true)
	return o
}

// ---------------------------------------------------------------------------
// Test 1: Creating an Outline with a tree produces correct visibleCount.
// Requirement: build a tree, create Outline, verify visibleCount matches
// expected count of visible nodes.
// ---------------------------------------------------------------------------

func TestIntegrationOutlineVisibleCountMatchesTree(t *testing.T) {
	root := makeIntegrationTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	got := o.visibleCount()
	if got != 10 {
		t.Errorf("visibleCount() = %d, want 10 (all nodes expanded)", got)
	}
}

// ---------------------------------------------------------------------------
// Test 2: nodeAt returns correct node and level at each visible index.
// Requirement: iterate through all visible indices, verify each returns the
// expected node text and nesting level.
// ---------------------------------------------------------------------------

func TestIntegrationOutlineNodeAtReturnsCorrectNodeAndLevel(t *testing.T) {
	root := makeIntegrationTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	type expect struct {
		text  string
		level int
	}
	expected := []expect{
		{"Animals", 0},
		{"Mammals", 1},
		{"Cat", 2},
		{"Dog", 2},
		{"Birds", 1},
		{"Eagle", 2},
		{"Parrot", 2},
		{"Plants", 0},
		{"Oak", 1},
		{"Pine", 1},
	}

	for i, exp := range expected {
		node, level := o.nodeAt(i)
		if node == nil {
			t.Fatalf("nodeAt(%d) returned nil, want %q at level %d", i, exp.text, exp.level)
		}
		if node.Text != exp.text {
			t.Errorf("nodeAt(%d).Text = %q, want %q", i, node.Text, exp.text)
		}
		if level != exp.level {
			t.Errorf("nodeAt(%d) level = %d, want %d", i, level, exp.level)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 3: Expanding/collapsing a node changes visibleCount.
// Requirement: collapse a node, verify count drops. Expand it, verify count
// restores.
// ---------------------------------------------------------------------------

func TestIntegrationOutlineCollapseExpandChangesVisibleCount(t *testing.T) {
	root := makeIntegrationTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	initial := o.visibleCount()
	if initial != 10 {
		t.Fatalf("setup: visibleCount() = %d, want 10", initial)
	}

	// Collapse "Mammals" (index 1) — hides Cat and Dog (2 nodes).
	o.focusedIdx = 1
	o.adjust()

	afterCollapse := o.visibleCount()
	if afterCollapse != 8 {
		t.Errorf("after collapsing Mammals: visibleCount() = %d, want 8", afterCollapse)
	}

	// Expand "Mammals" again — restores Cat and Dog.
	o.focusedIdx = 1
	o.adjust()

	afterExpand := o.visibleCount()
	if afterExpand != 10 {
		t.Errorf("after re-expanding Mammals: visibleCount() = %d, want 10", afterExpand)
	}
}

// ---------------------------------------------------------------------------
// Test 4: Keyboard Down moves focus, Up moves back, ensureVisible scrolls
// correctly.
// Requirement: send keyboard events through HandleEvent (Outline must have
// SfSelected set), verify focusedIdx changes and deltaY adjusts when
// scrolling past viewport.
// ---------------------------------------------------------------------------

func TestIntegrationOutlineKeyboardNavigationAndScrolling(t *testing.T) {
	root := makeIntegrationTree() // 10 visible nodes
	o := newIntOutline(root)      // viewport height = 5

	// Initial state.
	if o.focusedIdx != 0 {
		t.Fatalf("initial focusedIdx = %d, want 0", o.focusedIdx)
	}

	// Press Down 3 times: focusedIdx should advance to 3.
	for i := 0; i < 3; i++ {
		o.HandleEvent(intOutlineKeyEv(tcell.KeyDown))
	}
	if o.focusedIdx != 3 {
		t.Errorf("after 3 Down: focusedIdx = %d, want 3", o.focusedIdx)
	}
	// Still within viewport (height=5, deltaY=0 shows 0..4), no scroll needed.
	if o.deltaY != 0 {
		t.Errorf("after 3 Down: deltaY = %d, want 0 (still in viewport)", o.deltaY)
	}

	// Press Down 4 more times: focusedIdx=7, which is past viewport row 4.
	for i := 0; i < 4; i++ {
		o.HandleEvent(intOutlineKeyEv(tcell.KeyDown))
	}
	if o.focusedIdx != 7 {
		t.Errorf("after 7 Down total: focusedIdx = %d, want 7", o.focusedIdx)
	}
	// ensureVisible: deltaY = focusedIdx - height + 1 = 7 - 5 + 1 = 3.
	if o.deltaY != 3 {
		t.Errorf("after 7 Down total: deltaY = %d, want 3 (scrolled to keep focus visible)", o.deltaY)
	}

	// Press Up 2 times: focusedIdx=5.
	o.HandleEvent(intOutlineKeyEv(tcell.KeyUp))
	o.HandleEvent(intOutlineKeyEv(tcell.KeyUp))
	if o.focusedIdx != 5 {
		t.Errorf("after 2 Up: focusedIdx = %d, want 5", o.focusedIdx)
	}
	// deltaY should still be 3 (focusedIdx 5 is within 3..7).
	if o.deltaY != 3 {
		t.Errorf("after 2 Up: deltaY = %d, want 3 (focus still in viewport)", o.deltaY)
	}

	// Press Up 3 more times: focusedIdx=2, which is < deltaY=3.
	for i := 0; i < 3; i++ {
		o.HandleEvent(intOutlineKeyEv(tcell.KeyUp))
	}
	if o.focusedIdx != 2 {
		t.Errorf("after 5 Up total: focusedIdx = %d, want 2", o.focusedIdx)
	}
	// ensureVisible: deltaY = focusedIdx = 2.
	if o.deltaY != 2 {
		t.Errorf("after 5 Up total: deltaY = %d, want 2 (scrolled up to show focus)", o.deltaY)
	}
}

// ---------------------------------------------------------------------------
// Test 5: '+' expands, '-' collapses, and both Update() the visible row count.
// Requirement: use keyboard events through HandleEvent, verify Expanded toggles
// and visibleCount changes.
// ---------------------------------------------------------------------------

func TestIntegrationOutlinePlusMinusToggleExpandedAndUpdateCount(t *testing.T) {
	root := makeIntegrationTree() // 10 visible
	o := newIntOutline(root)

	// Focus on "Mammals" (index 1), which is expanded and has 2 children.
	o.HandleEvent(intOutlineKeyEv(tcell.KeyDown)) // focusedIdx = 1
	if o.focusedIdx != 1 {
		t.Fatalf("setup: focusedIdx = %d, want 1", o.focusedIdx)
	}

	// Press '-' to collapse Mammals. The adjust function toggles Expanded.
	o.HandleEvent(intOutlineRuneEv('-'))

	node, _ := o.nodeAt(1)
	if node == nil || node.Text != "Mammals" {
		t.Fatalf("nodeAt(1) should be Mammals after '-'")
	}
	if node.Expanded {
		t.Error("after '-': Mammals should be collapsed (Expanded=false)")
	}
	if o.visibleCount() != 8 {
		t.Errorf("after '-': visibleCount() = %d, want 8 (Cat+Dog hidden)", o.visibleCount())
	}

	// Press '+' to expand Mammals again. The adjust function toggles Expanded.
	o.HandleEvent(intOutlineRuneEv('+'))

	if !node.Expanded {
		t.Error("after '+': Mammals should be expanded (Expanded=true)")
	}
	if o.visibleCount() != 10 {
		t.Errorf("after '+': visibleCount() = %d, want 10 (Cat+Dog restored)", o.visibleCount())
	}
}

// ---------------------------------------------------------------------------
// Test 6: ScrollBar range and page size update when visibleCount changes.
// Requirement: attach a real ScrollBar, collapse/expand nodes, verify
// scrollbar's Min/Max/PageSize/Value update.
// ---------------------------------------------------------------------------

func TestIntegrationOutlineScrollBarUpdatesOnVisibleCountChange(t *testing.T) {
	root := makeIntegrationTree() // 10 visible
	o := newIntOutline(root)      // height = 5

	sb := NewScrollBar(NewRect(40, 0, 1, 5), Vertical)
	o.SetVScrollBar(sb)

	// Initial scrollbar state.
	if sb.Min() != 0 {
		t.Errorf("initial ScrollBar Min() = %d, want 0", sb.Min())
	}
	if sb.Max() != 10 {
		t.Errorf("initial ScrollBar Max() = %d, want 10", sb.Max())
	}
	if sb.PageSize() != 5 {
		t.Errorf("initial ScrollBar PageSize() = %d, want 5", sb.PageSize())
	}
	if sb.Value() != 0 {
		t.Errorf("initial ScrollBar Value() = %d, want 0", sb.Value())
	}

	// Collapse "Animals" (index 0) — hides everything under it (Mammals, Cat,
	// Dog, Birds, Eagle, Parrot = 6 nodes). Remaining visible: Animals, Plants,
	// Oak, Pine = 4.
	o.adjust() // focusedIdx = 0, toggles Animals
	if sb.Max() != 4 {
		t.Errorf("after collapse Animals: ScrollBar Max() = %d, want 4", sb.Max())
	}
	if sb.PageSize() != 5 {
		t.Errorf("after collapse Animals: ScrollBar PageSize() = %d, want 5 (unchanged)", sb.PageSize())
	}

	// Expand "Animals" again.
	o.adjust()
	if sb.Max() != 10 {
		t.Errorf("after expand Animals: ScrollBar Max() = %d, want 10", sb.Max())
	}

	// Navigate down and verify scrollbar value tracks deltaY.
	for i := 0; i < 7; i++ {
		o.HandleEvent(intOutlineKeyEv(tcell.KeyDown))
	}
	// focusedIdx=7, deltaY should be 3 (height=5, 7-5+1=3).
	if sb.Value() != o.deltaY {
		t.Errorf("after navigation: ScrollBar Value() = %d, want %d (deltaY)", sb.Value(), o.deltaY)
	}
}

// ---------------------------------------------------------------------------
// Test 7: ForEach visits all nodes regardless of expanded state.
// Requirement: collapse some nodes, call ForEach, verify it visits ALL nodes
// including collapsed children.
// ---------------------------------------------------------------------------

func TestIntegrationOutlineForEachVisitsAllNodesRegardlessOfExpanded(t *testing.T) {
	root := makeIntegrationTree() // 10 total nodes
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	// Collapse both parent nodes.
	root.Expanded = false               // collapse Animals -> hides 6 descendants
	root.Next.Expanded = false           // collapse Plants -> hides 2 descendants
	// visibleCount is now 2 (Animals and Plants).
	if o.visibleCount() != 2 {
		// Manually call Update because we mutated Expanded directly.
		o.Update()
	}

	visited := make([]string, 0, 10)
	o.ForEach(func(node *TNode, level int) {
		visited = append(visited, node.Text)
	})

	if len(visited) != 10 {
		t.Errorf("ForEach: visited %d nodes, want 10 (all nodes, regardless of collapse)", len(visited))
	}

	// Verify all expected nodes are present.
	expectedSet := map[string]bool{
		"Animals": true, "Mammals": true, "Cat": true, "Dog": true,
		"Birds": true, "Eagle": true, "Parrot": true,
		"Plants": true, "Oak": true, "Pine": true,
	}
	for _, name := range visited {
		if !expectedSet[name] {
			t.Errorf("ForEach: unexpected node %q", name)
		}
		delete(expectedSet, name)
	}
	for name := range expectedSet {
		t.Errorf("ForEach: missing node %q", name)
	}
}

// ---------------------------------------------------------------------------
// Test 8: FirstThat finds a node by text.
// Requirement: use FirstThat to search for a specific node, verify it's found.
// ---------------------------------------------------------------------------

func TestIntegrationOutlineFirstThatFindsByText(t *testing.T) {
	root := makeIntegrationTree()
	o := NewOutline(NewRect(0, 0, 40, 10), root)

	found := o.FirstThat(func(node *TNode, level int) bool {
		return node.Text == "Eagle"
	})

	if found == nil {
		t.Fatal("FirstThat: should find 'Eagle'")
	}
	if found.Text != "Eagle" {
		t.Errorf("FirstThat: found %q, want 'Eagle'", found.Text)
	}

	// Also verify a node that does not exist returns nil.
	notFound := o.FirstThat(func(node *TNode, level int) bool {
		return node.Text == "Fish"
	})
	if notFound != nil {
		t.Errorf("FirstThat: should return nil for 'Fish', got %q", notFound.Text)
	}
}

// ---------------------------------------------------------------------------
// Test 9: SetRoot replaces tree and resets state.
// Requirement: set focusedIdx and deltaY to non-zero, call SetRoot with new
// tree, verify both reset to 0 and visibleCount matches new tree.
// ---------------------------------------------------------------------------

func TestIntegrationOutlineSetRootResetsState(t *testing.T) {
	root := makeIntegrationTree() // 10 nodes
	o := newIntOutline(root)      // height=5

	// Navigate to put state in a non-default position.
	for i := 0; i < 7; i++ {
		o.HandleEvent(intOutlineKeyEv(tcell.KeyDown))
	}
	if o.focusedIdx == 0 || o.deltaY == 0 {
		t.Fatal("setup: focusedIdx and deltaY should be non-zero after navigation")
	}

	// Build a small new tree: X -> Y.
	y := NewNode("Y", nil, nil)
	x := NewNode("X", nil, y)

	o.SetRoot(x)

	if o.focusedIdx != 0 {
		t.Errorf("SetRoot: focusedIdx = %d, want 0 (reset)", o.focusedIdx)
	}
	if o.deltaY != 0 {
		t.Errorf("SetRoot: deltaY = %d, want 0 (reset)", o.deltaY)
	}
	if o.visibleCount() != 2 {
		t.Errorf("SetRoot: visibleCount() = %d, want 2 (X and Y)", o.visibleCount())
	}
	if o.Root() != x {
		t.Error("SetRoot: Root() should return the new root")
	}
}

// ---------------------------------------------------------------------------
// Test 10: OnSelect fires when Enter is pressed via HandleEvent.
// Requirement: set OnSelect callback, send Enter key through HandleEvent,
// verify callback fires with correct node.
// ---------------------------------------------------------------------------

func TestIntegrationOutlineOnSelectFiresOnEnter(t *testing.T) {
	root := makeIntegrationTree()
	o := newIntOutline(root)

	var selectedNode *TNode
	o.SetOnSelect(func(node *TNode) {
		selectedNode = node
	})

	// focusedIdx = 0, which is "Animals". Press Enter.
	o.HandleEvent(intOutlineKeyEv(tcell.KeyEnter))

	if selectedNode == nil {
		t.Fatal("OnSelect was not called on Enter")
	}
	if selectedNode.Text != "Animals" {
		t.Errorf("OnSelect called with %q, want 'Animals'", selectedNode.Text)
	}
}

// ---------------------------------------------------------------------------
// Test 11: Keyboard navigation + OnSelect end-to-end.
// Requirement: press Down 3 times, then Enter, verify OnSelect fires with the
// node at index 3.
// ---------------------------------------------------------------------------

func TestIntegrationOutlineNavigateThenSelect(t *testing.T) {
	root := makeIntegrationTree()
	o := newIntOutline(root)

	var selectedNode *TNode
	o.SetOnSelect(func(node *TNode) {
		selectedNode = node
	})

	// Press Down 3 times: focusedIdx 0 -> 1 -> 2 -> 3.
	for i := 0; i < 3; i++ {
		o.HandleEvent(intOutlineKeyEv(tcell.KeyDown))
	}
	if o.focusedIdx != 3 {
		t.Fatalf("after 3 Down: focusedIdx = %d, want 3", o.focusedIdx)
	}

	// Node at index 3 is "Dog".
	node, _ := o.nodeAt(3)
	if node == nil || node.Text != "Dog" {
		t.Fatalf("nodeAt(3) should be 'Dog', got %v", node)
	}

	// Press Enter.
	o.HandleEvent(intOutlineKeyEv(tcell.KeyEnter))

	if selectedNode == nil {
		t.Fatal("OnSelect was not called after navigation + Enter")
	}
	if selectedNode.Text != "Dog" {
		t.Errorf("OnSelect called with %q, want 'Dog'", selectedNode.Text)
	}
}

// ---------------------------------------------------------------------------
// Test 12: Collapse via keyboard, navigate, expand, verify tree state.
// Requirement: collapse a parent, navigate past it, expand it again, verify
// the full sequence produces correct tree state.
// ---------------------------------------------------------------------------

func TestIntegrationOutlineCollapseNavigateExpandVerifyState(t *testing.T) {
	root := makeIntegrationTree() // 10 visible
	o := newIntOutline(root)

	// Step 1: Navigate to "Mammals" (index 1) and collapse it.
	o.HandleEvent(intOutlineKeyEv(tcell.KeyDown)) // focusedIdx = 1
	if o.focusedIdx != 1 {
		t.Fatalf("step 1: focusedIdx = %d, want 1", o.focusedIdx)
	}

	o.HandleEvent(intOutlineRuneEv('-')) // toggle Mammals -> collapse

	mammalsNode, _ := o.nodeAt(1)
	if mammalsNode == nil || mammalsNode.Text != "Mammals" {
		t.Fatal("step 1: nodeAt(1) should be Mammals")
	}
	if mammalsNode.Expanded {
		t.Error("step 1: Mammals should be collapsed after '-'")
	}

	// After collapsing Mammals: visible = Animals, Mammals, Birds, Eagle,
	// Parrot, Plants, Oak, Pine = 8.
	if o.visibleCount() != 8 {
		t.Errorf("step 1: visibleCount() = %d, want 8", o.visibleCount())
	}

	// Step 2: Navigate past Mammals to "Birds" (now at index 2 after collapse).
	o.HandleEvent(intOutlineKeyEv(tcell.KeyDown)) // focusedIdx = 2
	birdsNode, _ := o.nodeAt(2)
	if birdsNode == nil || birdsNode.Text != "Birds" {
		t.Fatalf("step 2: nodeAt(2) should be 'Birds', got %v", birdsNode)
	}

	// Continue to Plants (index 5 after collapse: Animals, Mammals, Birds,
	// Eagle, Parrot, Plants).
	for i := 0; i < 3; i++ {
		o.HandleEvent(intOutlineKeyEv(tcell.KeyDown))
	}
	// focusedIdx should be 5.
	plantsNode, _ := o.nodeAt(o.focusedIdx)
	if plantsNode == nil || plantsNode.Text != "Plants" {
		t.Fatalf("step 2: focused node should be 'Plants', got %v (focusedIdx=%d)",
			plantsNode, o.focusedIdx)
	}

	// Step 3: Navigate back to Mammals and expand it.
	for o.focusedIdx > 1 {
		o.HandleEvent(intOutlineKeyEv(tcell.KeyUp))
	}
	if o.focusedIdx != 1 {
		t.Fatalf("step 3: focusedIdx = %d, want 1", o.focusedIdx)
	}

	o.HandleEvent(intOutlineRuneEv('+')) // toggle Mammals -> expand

	if !mammalsNode.Expanded {
		t.Error("step 3: Mammals should be expanded after '+'")
	}

	// After expanding Mammals: tree should be fully restored to 10 visible.
	if o.visibleCount() != 10 {
		t.Errorf("step 3: visibleCount() = %d, want 10", o.visibleCount())
	}

	// Verify the full order is restored.
	expectedOrder := []string{
		"Animals", "Mammals", "Cat", "Dog", "Birds",
		"Eagle", "Parrot", "Plants", "Oak", "Pine",
	}
	for i, exp := range expectedOrder {
		node, _ := o.nodeAt(i)
		if node == nil {
			t.Fatalf("step 3: nodeAt(%d) returned nil, want %q", i, exp)
		}
		if node.Text != exp {
			t.Errorf("step 3: nodeAt(%d).Text = %q, want %q", i, node.Text, exp)
		}
	}
}
