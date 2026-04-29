package tv

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
	panic("ExecView not implemented")
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

func (g *Group) BringToFront(v View) {
	for i, child := range g.children {
		if child == v {
			g.children = append(append(g.children[:i:i], g.children[i+1:]...), v)
			g.selectChild(v)
			return
		}
	}
}
