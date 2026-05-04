package tv

// TNode is a single node in a sibling-linked tree.
// Each node can have children (first child via Children, linked via Next)
// and siblings (linked via Next).
type TNode struct {
	Text     string
	Children *TNode
	Next     *TNode
	Expanded bool
}

// NewNode creates a node with the given text, children, and next sibling.
// Expanded defaults to true (children visible by default).
func NewNode(text string, children *TNode, next *TNode) *TNode {
	return &TNode{
		Text:     text,
		Children: children,
		Next:     next,
		Expanded: true,
	}
}
