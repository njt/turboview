package tv

import "github.com/gdamore/tcell/v2"

type Group struct {
	BaseView
	children []View
	focused  View
	facade   Container
}

func NewGroup(bounds Rect) *Group {
	g := &Group{}
	g.SetBounds(bounds)
	g.SetState(SfVisible, true)
	return g
}

func (g *Group) SetBounds(r Rect) {
	oldW := g.size.X
	oldH := g.size.Y
	g.BaseView.SetBounds(r)
	newW := r.Width()
	newH := r.Height()

	deltaW := newW - oldW
	deltaH := newH - oldH

	if deltaW == 0 && deltaH == 0 {
		return
	}

	for _, child := range g.children {
		gm := child.GrowMode()
		if gm == 0 {
			continue
		}

		cb := child.Bounds()

		if gm&GfGrowRel != 0 {
			ax, ay := cb.A.X, cb.A.Y
			bx, by := cb.B.X, cb.B.Y
			if oldW > 0 {
				ax = ax * newW / oldW
				bx = bx * newW / oldW
			}
			if oldH > 0 {
				ay = ay * newH / oldH
				by = by * newH / oldH
			}
			child.SetBounds(Rect{A: Point{ax, ay}, B: Point{bx, by}})
			continue
		}

		ax, ay := cb.A.X, cb.A.Y
		bx, by := cb.B.X, cb.B.Y
		if gm&GfGrowLoX != 0 {
			ax += deltaW
			if gm&GfGrowHiX == 0 {
				bx += deltaW
			}
		}
		if gm&GfGrowHiX != 0 {
			bx += deltaW
		}
		if gm&GfGrowLoY != 0 {
			ay += deltaH
			if gm&GfGrowHiY == 0 {
				by += deltaH
			}
		}
		if gm&GfGrowHiY != 0 {
			by += deltaH
		}
		child.SetBounds(Rect{A: Point{ax, ay}, B: Point{bx, by}})
	}
}

func (g *Group) SetFacade(c Container) {
	g.facade = c
}

func (g *Group) Insert(v View) {
	owner := Container(g)
	if g.facade != nil {
		owner = g.facade
	}
	v.SetOwner(owner)
	g.children = append(g.children, v)
	if v.HasOption(OfSelectable) {
		g.selectChild(v)
	}
}

func (g *Group) Remove(v View) {
	for i, child := range g.children {
		if child == v {
			g.children = append(g.children[:i], g.children[i+1:]...)
			v.SetOwner(nil)
			if g.focused == v {
				g.focused = nil
				g.selectPrevious()
			}
			return
		}
	}
}

func (g *Group) Children() []View {
	return g.children
}

func (g *Group) FocusedChild() View {
	return g.focused
}

func (g *Group) SetFocusedChild(v View) {
	g.selectChild(v)
}

func (g *Group) ExecView(v View) CommandCode {
	g.Insert(v)
	v.SetState(SfModal, true)

	// Walk owner chain from facade to find Application via Desktop.
	var app *Application
	var current Container = g.facade
	if current == nil {
		current = Container(g)
	}
	for current != nil {
		if d, ok := current.(*Desktop); ok && d.app != nil {
			app = d.app
			break
		}
		if view, ok := current.(View); ok {
			current = view.Owner()
		} else {
			break
		}
	}

	if app == nil {
		g.Remove(v)
		v.SetState(SfModal, false)
		return CmCancel
	}

	// Draw immediately so the modal view is visible before the first event
	app.drawAndFlush()

	// Modal event loop
	var result CommandCode
	for {
		event := app.PollEvent()
		if event == nil {
			result = CmCancel
			break
		}

		// Route event to modal view only
		if event.What == EvMouse && event.Mouse != nil {
			vb := v.Bounds()
			mx, my := event.Mouse.X, event.Mouse.Y
			if vb.Contains(NewPoint(mx, my)) {
				event.Mouse.X -= vb.A.X
				event.Mouse.Y -= vb.A.Y
				v.HandleEvent(event)
			}
			// Outside clicks are discarded
		} else {
			v.HandleEvent(event)
		}

		// Check for closing command (Button.press transforms event in place)
		if event.What == EvCommand {
			switch event.Command {
			case CmOK, CmCancel, CmClose, CmYes, CmNo:
				result = event.Command
			}
		}

		app.drawAndFlush()

		if result != 0 {
			break
		}
	}

	g.Remove(v)
	v.SetState(SfModal, false)
	return result
}

func (g *Group) Draw(buf *DrawBuffer) {
	for _, child := range g.children {
		if !child.HasState(SfVisible) {
			continue
		}
		childBounds := child.Bounds()
		sub := buf.SubBuffer(childBounds)
		child.Draw(sub)
	}
}

func (g *Group) HandleEvent(event *Event) {
	if event.IsCleared() {
		return
	}

	// Mouse events: forward to focused child (positional routing done by caller)
	if event.What == EvMouse {
		if g.focused != nil {
			g.focused.HandleEvent(event)
		}
		return
	}

	// Broadcast: deliver to all children
	if event.What == EvBroadcast {
		for _, child := range g.children {
			if event.IsCleared() {
				return
			}
			child.HandleEvent(event)
		}
		return
	}

	// Tab/Shift+Tab focus traversal — before three-phase dispatch
	if event.What == EvKeyboard && event.Key != nil {
		if event.Key.Key == tcell.KeyTab && event.Key.Modifiers == 0 {
			g.focusNext()
			event.Clear()
			return
		}
		if event.Key.Key == tcell.KeyBacktab {
			g.focusPrev()
			event.Clear()
			return
		}
	}

	// Three-phase dispatch for keyboard and command events

	// Phase 1: Preprocess
	for _, child := range g.children {
		if event.IsCleared() {
			return
		}
		if child != g.focused && child.HasOption(OfPreProcess) {
			child.HandleEvent(event)
		}
	}

	// Phase 2: Focused
	if !event.IsCleared() && g.focused != nil {
		g.focused.HandleEvent(event)
	}

	// Phase 3: Postprocess
	for _, child := range g.children {
		if event.IsCleared() {
			return
		}
		if child != g.focused && child.HasOption(OfPostProcess) {
			child.HandleEvent(event)
		}
	}
}

func (g *Group) selectChild(v View) {
	if g.focused != nil && g.focused != v {
		g.focused.SetState(SfSelected, false)
	}
	g.focused = v
	if v != nil {
		v.SetState(SfSelected, true)
	}
}

func (g *Group) selectPrevious() {
	for i := len(g.children) - 1; i >= 0; i-- {
		if g.children[i].HasOption(OfSelectable) {
			g.selectChild(g.children[i])
			return
		}
	}
}

func (g *Group) focusNext() {
	if len(g.children) == 0 {
		return
	}
	start := 0
	if g.focused != nil {
		for i, child := range g.children {
			if child == g.focused {
				start = i + 1
				break
			}
		}
	}
	n := len(g.children)
	for i := 0; i < n; i++ {
		idx := (start + i) % n
		if g.children[idx].HasOption(OfSelectable) && g.children[idx] != g.focused {
			g.selectChild(g.children[idx])
			return
		}
	}
}

func (g *Group) focusPrev() {
	if len(g.children) == 0 {
		return
	}
	start := len(g.children) - 1
	if g.focused != nil {
		for i, child := range g.children {
			if child == g.focused {
				start = i - 1
				break
			}
		}
	}
	n := len(g.children)
	for i := 0; i < n; i++ {
		idx := (start - i + n) % n
		if g.children[idx].HasOption(OfSelectable) && g.children[idx] != g.focused {
			g.selectChild(g.children[idx])
			return
		}
	}
}

func (g *Group) BringToFront(v View) {
	for i, child := range g.children {
		if child == v {
			g.children = append(append(g.children[:i:i], g.children[i+1:]...), v)
			g.selectChild(v)
			return
		}
	}
}
