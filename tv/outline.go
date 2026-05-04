package tv

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// OutlineViewer is a tree outline viewer widget that displays a hierarchical
// tree of TNodes with expand/collapse, keyboard navigation, and mouse support.
type OutlineViewer struct {
	BaseView
	root        *TNode
	focusedIdx  int
	deltaY      int
	vScrollBar  *ScrollBar
	selectedFn  func()
	adjustFn    func()
	adjustAllFn func()
}

// NewOutlineViewer creates a new OutlineViewer with the given bounds.
func NewOutlineViewer(bounds Rect) *OutlineViewer {
	ov := &OutlineViewer{
		root:       nil,
		focusedIdx: 0,
		deltaY:     0,
	}
	ov.SetBounds(bounds)
	ov.SetState(SfVisible, true)
	ov.SetOptions(OfSelectable, true)
	ov.SetOptions(OfFirstClick, true)
	ov.SetSelf(ov)
	ov.selectedFn = ov.selected
	ov.adjustFn = ov.adjust
	ov.adjustAllFn = ov.adjustAll
	return ov
}

// visibleCount returns the number of visible nodes in a depth-first traversal.
// Collapsed nodes' children are not counted.
func (ov *OutlineViewer) visibleCount() int {
	if ov.root == nil {
		return 0
	}
	count := 0
	var walk func(n *TNode)
	walk = func(n *TNode) {
		if n == nil {
			return
		}
		count++
		if n.Expanded {
			walk(n.Children)
		}
		walk(n.Next)
	}
	walk(ov.root)
	return count
}

// nodeAt returns the node and level (root=0) at the given flattened index.
// Returns nil, 0 if idx is out of bounds or root is nil.
func (ov *OutlineViewer) nodeAt(idx int) (*TNode, int) {
	if ov.root == nil || idx < 0 {
		return nil, 0
	}
	var result *TNode
	var resultLevel int
	count := 0
	found := false

	var walk func(n *TNode, lvl int)
	walk = func(n *TNode, lvl int) {
		if n == nil || found {
			return
		}
		if count == idx {
			result = n
			resultLevel = lvl
			found = true
			return
		}
		count++
		if n.Expanded {
			walk(n.Children, lvl+1)
		}
		if !found {
			walk(n.Next, lvl)
		}
	}
	walk(ov.root, 0)
	return result, resultLevel
}

// graphPrefix returns the tree-drawing prefix string, node, and level for the
// node at the given flattened index. The prefix includes ancestor lines,
// connector, and status character.
func (ov *OutlineViewer) graphPrefix(targetIdx int) (prefix string, node *TNode, level int) {
	if ov.root == nil {
		return "", nil, 0
	}
	pathEntries := make([]bool, 0, 8)
	var result *TNode
	var resultLevel int
	count := 0
	found := false

	var walk func(n *TNode, lvl int)
	walk = func(n *TNode, lvl int) {
		if n == nil || found {
			return
		}
		for len(pathEntries) <= lvl {
			pathEntries = append(pathEntries, false)
		}
		pathEntries[lvl] = (n.Next == nil)

		if count == targetIdx {
			result = n
			resultLevel = lvl
			found = true
			return
		}
		count++
		if n.Expanded {
			walk(n.Children, lvl+1)
		}
		if !found {
			walk(n.Next, lvl)
		}
	}
	walk(ov.root, 0)
	if result == nil {
		return "", nil, 0
	}

	var sb strings.Builder
	for i := 0; i < resultLevel; i++ {
		if pathEntries[i] {
			sb.WriteString("   ")
		} else {
			sb.WriteString("│  ")
		}
	}
	if pathEntries[resultLevel] {
		sb.WriteString("└──")
	} else {
		sb.WriteString("├──")
	}
	if result.Children != nil && !result.Expanded {
		sb.WriteRune('+')
	} else if result.Children != nil {
		sb.WriteRune('─')
	} else {
		sb.WriteByte(' ')
	}
	return sb.String(), result, resultLevel
}

// getStyle returns the appropriate style for a node at the given flat index.
func (ov *OutlineViewer) getStyle(flatIdx int, node *TNode) tcell.Style {
	cs := ov.ColorScheme()
	if cs == nil {
		return tcell.StyleDefault
	}
	if flatIdx == ov.focusedIdx && ov.HasState(SfSelected) {
		return cs.OutlineFocused
	}
	if node.Children != nil && !node.Expanded {
		return cs.OutlineCollapsed
	}
	return cs.OutlineNormal
}

// Draw renders the outline tree into the draw buffer.
func (ov *OutlineViewer) Draw(buf *DrawBuffer) {
	if ov.root == nil {
		return
	}
	h := ov.Bounds().Height()
	vc := ov.visibleCount()

	for flatIdx := ov.deltaY; flatIdx < ov.deltaY+h && flatIdx < vc; flatIdx++ {
		y := flatIdx - ov.deltaY
		prefix, node, _ := ov.graphPrefix(flatIdx)
		if node == nil {
			break
		}
		style := ov.getStyle(flatIdx, node)

		col := 0
		for _, ch := range prefix {
			buf.WriteChar(col, y, ch, style)
			col++
		}
		for _, ch := range node.Text {
			buf.WriteChar(col, y, ch, style)
			col++
		}
	}
}

// HandleEvent dispatches keyboard and mouse events.
func (ov *OutlineViewer) HandleEvent(event *Event) {
	if event.What == EvMouse {
		ov.BaseView.HandleEvent(event)
		if event.IsCleared() {
			return
		}
	}

	if event.What == EvKeyboard {
		if !ov.HasState(SfSelected) {
			return
		}
		ov.handleKeyboard(event)
		return
	}

	if event.What == EvMouse && event.Mouse != nil {
		ov.handleMouse(event)
	}
}

func (ov *OutlineViewer) handleKeyboard(event *Event) {
	k := event.Key
	vc := ov.visibleCount()

	switch k.Key {
	case tcell.KeyUp:
		if ov.focusedIdx > 0 {
			ov.focusedIdx--
		}
		ov.ensureVisible()
		ov.syncScrollBars()
		event.Clear()

	case tcell.KeyDown:
		if vc > 0 && ov.focusedIdx < vc-1 {
			ov.focusedIdx++
		}
		ov.ensureVisible()
		ov.syncScrollBars()
		event.Clear()

	case tcell.KeyRight:
		if vc > 0 && ov.focusedIdx < vc-1 {
			ov.focusedIdx++
		}
		ov.ensureVisible()
		ov.syncScrollBars()
		event.Clear()

	case tcell.KeyLeft:
		if ov.focusedIdx > 0 {
			ov.focusedIdx--
		}
		ov.ensureVisible()
		ov.syncScrollBars()
		event.Clear()

	case tcell.KeyEnter:
		ov.selectedFn()
		event.Clear()

	case tcell.KeyPgUp:
		if k.Modifiers&tcell.ModCtrl != 0 {
			ov.focusedIdx = 0
		} else {
			ov.focusedIdx -= ov.Bounds().Height() - 1
			if ov.focusedIdx < 0 {
				ov.focusedIdx = 0
			}
		}
		ov.ensureVisible()
		ov.syncScrollBars()
		event.Clear()

	case tcell.KeyPgDn:
		if k.Modifiers&tcell.ModCtrl != 0 {
			if vc > 0 {
				ov.focusedIdx = vc - 1
			}
		} else {
			ov.focusedIdx += ov.Bounds().Height() - 1
			if vc > 0 && ov.focusedIdx >= vc {
				ov.focusedIdx = vc - 1
			}
		}
		ov.ensureVisible()
		ov.syncScrollBars()
		event.Clear()

	case tcell.KeyHome:
		ov.focusedIdx = ov.deltaY
		ov.ensureVisible()
		ov.syncScrollBars()
		event.Clear()

	case tcell.KeyEnd:
		ov.focusedIdx = ov.deltaY + ov.Bounds().Height() - 1
		if vc > 0 && ov.focusedIdx >= vc {
			ov.focusedIdx = vc - 1
		}
		if ov.focusedIdx < 0 {
			ov.focusedIdx = 0
		}
		ov.ensureVisible()
		ov.syncScrollBars()
		event.Clear()

	case tcell.KeyRune:
		switch k.Rune {
		case '+':
			ov.adjustFn()
			event.Clear()
		case '-':
			ov.adjustFn()
			event.Clear()
		case '*':
			ov.adjustAllFn()
			event.Clear()
		}
	}
}

func (ov *OutlineViewer) handleMouse(event *Event) {
	m := event.Mouse
	// Only handle Button1.
	if m.Button != tcell.Button1 {
		return
	}

	b := ov.Bounds()
	row := ov.deltaY + (m.Y - b.A.Y)
	vc := ov.visibleCount()
	if row < 0 || vc == 0 || row >= vc {
		return
	}

	// Move focus to the clicked row.
	ov.focusedIdx = row

	_, node, level := ov.graphPrefix(row)
	if node == nil {
		return
	}

	graphWidth := level*3 + 4
	clickX := m.X - b.A.X

	if m.ClickCount >= 2 {
		// Double-click: select, then toggle if has children.
		ov.selectedFn()
		if node.Children != nil {
			ov.adjustFn()
		}
		event.Clear()
		return
	}

	if clickX < graphWidth {
		// Click in graph area: toggle expand/collapse.
		ov.adjustFn()
	}
	event.Clear()
}

// SetVScrollBar sets or clears the vertical scrollbar.
func (ov *OutlineViewer) SetVScrollBar(sb *ScrollBar) {
	if ov.vScrollBar != nil {
		ov.vScrollBar.OnChange = nil
	}
	ov.vScrollBar = sb
	if sb != nil {
		sb.OnChange = func(v int) {
			ov.deltaY = v
		}
		// Sync scrollbar visibility to match current selected state
		sb.SetState(SfVisible, ov.HasState(SfSelected))
	}
	ov.syncScrollBars()
}

// syncScrollBars updates the scrollbar's range, page size, and value.
func (ov *OutlineViewer) syncScrollBars() {
	if ov.vScrollBar == nil {
		return
	}
	max := ov.visibleCount()
	if max < 0 {
		max = 0
	}
	ov.vScrollBar.SetRange(0, max)
	ov.vScrollBar.SetPageSize(ov.Bounds().Height())
	ov.vScrollBar.SetValue(ov.deltaY)
}

// ensureVisible scrolls the viewport so that the focused row is visible.
func (ov *OutlineViewer) ensureVisible() {
	h := ov.Bounds().Height()
	vc := ov.visibleCount()

	if ov.focusedIdx < ov.deltaY {
		ov.deltaY = ov.focusedIdx
	}
	if ov.focusedIdx >= ov.deltaY+h {
		ov.deltaY = ov.focusedIdx - h + 1
	}

	maxDelta := vc - h
	if maxDelta < 0 {
		maxDelta = 0
	}
	if ov.deltaY < 0 {
		ov.deltaY = 0
	}
	if ov.deltaY > maxDelta {
		ov.deltaY = maxDelta
	}

	ov.syncScrollBars()
}

// SetState updates view state and toggles scrollbar visibility when SfSelected changes.
func (ov *OutlineViewer) SetState(flag ViewState, on bool) {
	ov.BaseView.SetState(flag, on)
	if flag == SfSelected && ov.vScrollBar != nil {
		ov.vScrollBar.SetState(SfVisible, on)
	}
}

// adjust toggles the Expanded flag of the focused node and updates the view.
func (ov *OutlineViewer) adjust() {
	node, _ := ov.nodeAt(ov.focusedIdx)
	if node != nil {
		node.Expanded = !node.Expanded
	}
	ov.syncScrollBars()
	ov.ensureVisible()
}

// adjustAll expands the focused node and all its descendants, then updates the view.
func (ov *OutlineViewer) adjustAll() {
	node, _ := ov.nodeAt(ov.focusedIdx)
	if node != nil {
		expandAll(node)
	}
	ov.syncScrollBars()
	ov.ensureVisible()
}

// selected is a placeholder overridden by Outline.
func (ov *OutlineViewer) selected() {}

// expandAll recursively sets Expanded=true on n and all descendants.
func expandAll(n *TNode) {
	if n == nil {
		return
	}
	n.Expanded = true
	for child := n.Children; child != nil; child = child.Next {
		expandAll(child)
	}
}

// ---------------------------------------------------------------------------
// Outline — concrete class that owns a node tree
// ---------------------------------------------------------------------------

// Outline is the concrete outline widget that owns a node tree and provides
// visitor methods and an OnSelect callback. It embeds OutlineViewer by value.
type Outline struct {
	OutlineViewer
	OnSelect func(node *TNode)
}

// NewOutline creates an Outline with the given bounds and root node.
func NewOutline(bounds Rect, root *TNode) *Outline {
	o := &Outline{}
	o.SetBounds(bounds)
	o.SetState(SfVisible, true)
	o.SetOptions(OfSelectable, true)
	o.SetOptions(OfFirstClick, true)
	o.SetGrowMode(GfGrowHiX | GfGrowHiY)
	o.SetSelf(o)
	o.selectedFn = o.selected
	o.adjustFn = o.adjust
	o.adjustAllFn = o.adjustAll
	o.root = root
	o.Update()
	return o
}

// Root returns the root node of the outline tree.
func (o *Outline) Root() *TNode { return o.root }

// SetRoot replaces the root node, resets focusedIdx and deltaY, and calls Update.
func (o *Outline) SetRoot(root *TNode) {
	o.root = root
	o.focusedIdx = 0
	o.deltaY = 0
	o.Update()
}

// SetOnSelect sets the OnSelect callback.
func (o *Outline) SetOnSelect(fn func(node *TNode)) {
	o.OnSelect = fn
}

// Update recomputes the visible count and syncs the scrollbar.
// If focusedIdx >= visibleCount(), it is clamped to visibleCount()-1.
func (o *Outline) Update() {
	total := o.visibleCount()
	if total > 0 && o.focusedIdx >= total {
		o.focusedIdx = total - 1
	}
	o.ensureVisible()
	o.syncScrollBars()
}

// selected finds the focused node and calls OnSelect if set.
func (o *Outline) selected() {
	node, _ := o.nodeAt(o.focusedIdx)
	if node != nil && o.OnSelect != nil {
		o.OnSelect(node)
	}
}

// adjust toggles the Expanded flag of the focused node and calls Update.
func (o *Outline) adjust() {
	node, _ := o.nodeAt(o.focusedIdx)
	if node != nil {
		node.Expanded = !node.Expanded
	}
	o.Update()
}

// adjustAll expands the focused node and all its descendants, then calls Update.
func (o *Outline) adjustAll() {
	node, _ := o.nodeAt(o.focusedIdx)
	if node != nil {
		expandAll(node)
	}
	o.Update()
}

// ForEach performs a depth-first traversal of ALL nodes (including collapsed).
// Visits: node, then Children, then Next. Level starts at 0.
func (o *Outline) ForEach(fn func(node *TNode, level int)) {
	forEachNode(o.root, 0, fn)
}

func forEachNode(n *TNode, level int, fn func(*TNode, int)) {
	if n == nil {
		return
	}
	fn(n, level)
	forEachNode(n.Children, level+1, fn)
	forEachNode(n.Next, level, fn)
}

// FirstThat performs a depth-first traversal of ALL nodes (including collapsed)
// and returns the first node where fn returns true, or nil.
func (o *Outline) FirstThat(fn func(node *TNode, level int) bool) *TNode {
	return firstThatNode(o.root, 0, fn)
}

func firstThatNode(n *TNode, level int, fn func(*TNode, int) bool) *TNode {
	if n == nil {
		return nil
	}
	if fn(n, level) {
		return n
	}
	if result := firstThatNode(n.Children, level+1, fn); result != nil {
		return result
	}
	return firstThatNode(n.Next, level, fn)
}
