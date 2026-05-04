package tv

import (
	"testing"
)

// ---------------------------------------------------------------------------
// TNode tests — Batch 9 Task 1: TNode data type for outline tree nodes.
//
// Every assertion cites the spec requirement it verifies.
// Each test covers exactly one behaviour.
//
// Spec requirements tested:
//   1. TNode struct has fields: Text string, Children *TNode, Next *TNode, Expanded bool
//   2. NewNode(text string, children *TNode, next *TNode) *TNode creates a node with initial values
//   3. Nodes form a sibling-linked forest: Next chains siblings
//   4. Children are themselves sibling-linked lists: Children → child.Next → child.Next.Next
//   5. Expanded defaults to true (visibility of children)
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Requirement 1 — NewNode sets all fields from its arguments.
// Spec: "NewNode(text string, children *TNode, next *TNode) *TNode creates a node with initial values"
// ---------------------------------------------------------------------------

// TestNewNodeSetsFields verifies NewNode creates a TNode and sets Text, Children, Next, and Expanded.
func TestNewNodeSetsFields(t *testing.T) {
	child := &TNode{Text: "child"}
	next := &TNode{Text: "sibling"}

	n := NewNode("root", child, next)

	if n.Text != "root" {
		t.Errorf("Text = %q, want %q", n.Text, "root")
	}
	if n.Children != child {
		t.Errorf("Children = %v, want %v", n.Children, child)
	}
	if n.Next != next {
		t.Errorf("Next = %v, want %v", n.Next, next)
	}
	if !n.Expanded {
		t.Error("Expanded should default to true")
	}
}

// ---------------------------------------------------------------------------
// Requirement 2 — NewNode handles nil children and next.
// Spec: "NewNode(text string, children *TNode, next *TNode) creates a node with initial values"
// Corner case: nil arguments should be stored as nil.
// ---------------------------------------------------------------------------

// TestNewNodeNilArgs verifies NewNode accepts nil for Children and Next.
func TestNewNodeNilArgs(t *testing.T) {
	n := NewNode("leaf", nil, nil)

	if n.Text != "leaf" {
		t.Errorf("Text = %q, want %q", n.Text, "leaf")
	}
	if n.Children != nil {
		t.Error("Children should be nil when passed nil")
	}
	if n.Next != nil {
		t.Error("Next should be nil when passed nil")
	}
	if !n.Expanded {
		t.Error("Expanded should default to true even with nil args")
	}
}

// ---------------------------------------------------------------------------
// Requirement 3 — Nodes form a sibling-linked forest via Next.
// Spec: "the root parameter is the first of potentially many top-level sibling nodes linked by Next"
// ---------------------------------------------------------------------------

// TestTNodeSiblingChain verifies Next chains siblings in a linked list.
func TestTNodeSiblingChain(t *testing.T) {
	// Build:  A → B → C → nil
	a := NewNode("A", nil, nil)
	b := NewNode("B", nil, nil)
	c := NewNode("C", nil, nil)
	a.Next = b
	b.Next = c

	// Verify the chain
	if a.Next != b {
		t.Error("A.Next should be B")
	}
	if b.Next != c {
		t.Error("B.Next should be C")
	}
	if c.Next != nil {
		t.Error("C.Next should be nil (end of chain)")
	}

	// Traverse the chain and verify Text values
	if a.Text != "A" {
		t.Errorf("A.Text = %q, want %q", a.Text, "A")
	}
	if a.Next.Text != "B" {
		t.Errorf("A.Next.Text = %q, want %q", a.Next.Text, "B")
	}
	if a.Next.Next.Text != "C" {
		t.Errorf("A.Next.Next.Text = %q, want %q", a.Next.Next.Text, "C")
	}
}

// ---------------------------------------------------------------------------
// Requirement 4 — Children are sibling-linked lists.
// Spec: "A node's children are linked through Children → child.Next → child.Next.Next, etc."
// ---------------------------------------------------------------------------

// TestTNodeChildrenChain verifies a node's Children form a sibling-linked list.
func TestTNodeChildrenChain(t *testing.T) {
	// Build children:  child1 → child2 → child3 → nil
	child1 := NewNode("child1", nil, nil)
	child2 := NewNode("child2", nil, nil)
	child3 := NewNode("child3", nil, nil)
	child1.Next = child2
	child2.Next = child3

	root := NewNode("root", child1, nil)

	// Verify the children chain
	if root.Children != child1 {
		t.Error("root.Children should point to child1")
	}
	if root.Children.Next != child2 {
		t.Error("root.Children.Next should be child2")
	}
	if root.Children.Next.Next != child3 {
		t.Error("root.Children.Next.Next should be child3")
	}
	if root.Children.Next.Next.Next != nil {
		t.Error("child3.Next should be nil (end of children list)")
	}

	// Verify children text values
	if root.Children.Text != "child1" {
		t.Errorf("child1.Text = %q, want %q", root.Children.Text, "child1")
	}
	if root.Children.Next.Text != "child2" {
		t.Errorf("child2.Text = %q, want %q", root.Children.Next.Text, "child2")
	}
}

// ---------------------------------------------------------------------------
// Requirement 5 — Expanded defaults to true.
// Spec: "Expanded controls whether children are visible" and the struct includes "Expanded bool".
// The default value of a bool in Go is false, but the spec requires true for visibility.
// We verify NewNode sets Expanded to true.
// ---------------------------------------------------------------------------

// TestTNodeExpandedDefault verifies Expanded defaults to true via NewNode.
func TestTNodeExpandedDefault(t *testing.T) {
	n := NewNode("test", nil, nil)

	if !n.Expanded {
		t.Error("Expanded should default to true (children visible by default)")
	}
}

// ---------------------------------------------------------------------------
// Falsifying test: verify Expanded is not accidentally set to false.
// Spec: "Expanded controls whether children are visible" — children should be visible by default.
// ---------------------------------------------------------------------------

// TestTNodeExpandedNotFalseByDefault verifies Expanded is not defaulting to the Go zero value (false).
func TestTNodeExpandedNotFalseByDefault(t *testing.T) {
	n := NewNode("test", nil, nil)

	if n.Expanded != true {
		t.Error("Expanded should be true, not false. Children are visible by default per spec.")
	}
}

// ---------------------------------------------------------------------------
// TestTNodeCanHaveBothChildrenAndNext verifies a node can simultaneously have
// children (downward) and a next sibling (rightward) — the full tree shape.
// ---------------------------------------------------------------------------

// TestTNodeCanHaveBothChildrenAndNext verifies sibling-linked forest: a node
// can have both Children and Next.
func TestTNodeCanHaveBothChildrenAndNext(t *testing.T) {
	grandChild := NewNode("gc", nil, nil)
	child := NewNode("c", grandChild, nil)
	sibling := NewNode("s", nil, nil)

	root := NewNode("r", child, sibling)

	// Children downward
	if root.Children != child {
		t.Error("root.Children should be child")
	}
	if root.Children.Children != grandChild {
		t.Error("root.Children.Children should be grandChild")
	}

	// Next rightward
	if root.Next != sibling {
		t.Error("root.Next should be sibling")
	}

	// Path: root → root.Children (child) → root.Children.Children (grandChild)
	gc := root.Children.Children
	if gc.Text != "gc" {
		t.Errorf("grandChild.Text = %q, want %q", gc.Text, "gc")
	}
}

// ---------------------------------------------------------------------------
// TestNewNodeExpandedAlwaysTrue verifies Expanded is always true regardless of
// arguments — falsifying: that Expanded depends on arguments.
// ---------------------------------------------------------------------------

// TestNewNodeExpandedAlwaysTrue verifies Expanded is set to true even when
// children or next are provided (not just nil).
func TestNewNodeExpandedAlwaysTrue(t *testing.T) {
	n1 := NewNode("a", &TNode{Text: "child"}, nil)
	n2 := NewNode("b", nil, &TNode{Text: "next"})
	n3 := NewNode("c", &TNode{Text: "c1"}, &TNode{Text: "s1"})

	if !n1.Expanded {
		t.Error("Expanded should be true with Children set")
	}
	if !n2.Expanded {
		t.Error("Expanded should be true with Next set")
	}
	if !n3.Expanded {
		t.Error("Expanded should be true with both Children and Next set")
	}
}
