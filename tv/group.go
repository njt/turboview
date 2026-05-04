package tv


const (
	PhPreProcess  = 0
	PhFocused     = 1
	PhPostProcess = 2
)

type Group struct {
	BaseView
	children []View
	focused  View
	facade   Container
	phase    int
}

func (g *Group) Phase() int { return g.phase }

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

	if oldW == 0 && oldH == 0 {
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
	if v == nil {
		return
	}
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

// setFocusSilent updates focused and SfSelected flags without broadcasting.
// Used internally when a broadcast-free focus update is needed.
func (g *Group) setFocusSilent(v View) {
	if g.focused == v {
		return
	}
	if g.focused != nil {
		g.focused.SetState(SfSelected, false)
	}
	g.focused = v
	if v != nil {
		v.SetState(SfSelected, true)
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

	// Compute the screen offset of this group's coordinate space so we can
	// translate screen-absolute mouse coordinates from PollEvent into group-local
	// coordinates before comparing against the dialog's bounds.
	offX, offY := g.screenOffset()

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
			localX := event.Mouse.X - offX
			localY := event.Mouse.Y - offY
			if vb.Contains(NewPoint(localX, localY)) {
				event.Mouse.X = localX - vb.A.X
				event.Mouse.Y = localY - vb.A.Y
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

// screenOffset computes the accumulated offset from screen-absolute coordinates
// to this group's local coordinate space. This walks up the owner chain, adding
// each ancestor's bounds origin and frame offset (for Window and Dialog frames).
func (g *Group) screenOffset() (int, int) {
	offX, offY := 0, 0
	var cur Container = g.facade
	if cur == nil {
		return 0, 0
	}
	for cur != nil {
		view, ok := cur.(View)
		if !ok {
			break
		}
		b := view.Bounds()
		offX += b.A.X
		offY += b.A.Y
		// Framed containers (Window, Dialog) offset their client area by (1,1)
		switch cur.(type) {
		case *Window, *Dialog:
			offX++
			offY++
		}
		owner := view.Owner()
		if owner == nil {
			break
		}
		cur = owner
	}
	return offX, offY
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

func (g *Group) canReceiveEvent(child View, eventType EventType) bool {
	if child.HasState(SfDisabled) {
		return false
	}
	mask := child.EventMask()
	if mask == 0 {
		return true
	}
	return eventType&mask != 0
}

func (g *Group) HandleEvent(event *Event) {
	if event.IsCleared() {
		return
	}

	// Mouse events: positional routing, back-to-front
	if event.What == EvMouse {
		mx, my := event.Mouse.X, event.Mouse.Y
		for i := len(g.children) - 1; i >= 0; i-- {
			child := g.children[i]
			if !child.HasState(SfVisible) || child.HasState(SfDisabled) {
				continue
			}
			if child.Bounds().Contains(NewPoint(mx, my)) {
				event.Mouse.X -= child.Bounds().A.X
				event.Mouse.Y -= child.Bounds().A.Y
				child.HandleEvent(event)
				return
			}
		}
		return
	}

	// Broadcast: deliver to all children (skip disabled; do NOT check EventMask)
	if event.What == EvBroadcast {
		for _, child := range g.children {
			if event.IsCleared() {
				return
			}
			if child.HasState(SfDisabled) {
				continue
			}
			child.HandleEvent(event)
		}
		return
	}

	// Three-phase dispatch for keyboard and command events

	// Phase 1: Preprocess
	g.phase = PhPreProcess
	for _, child := range g.children {
		if event.IsCleared() {
			return
		}
		if child != g.focused && child.HasOption(OfPreProcess) && g.canReceiveEvent(child, event.What) {
			child.HandleEvent(event)
		}
	}

	// Phase 2: Focused
	g.phase = PhFocused
	if !event.IsCleared() && g.focused != nil && g.canReceiveEvent(g.focused, event.What) {
		g.focused.HandleEvent(event)
	}

	// Phase 3: Postprocess
	g.phase = PhPostProcess
	for _, child := range g.children {
		if event.IsCleared() {
			return
		}
		if child != g.focused && child.HasOption(OfPostProcess) && g.canReceiveEvent(child, event.What) {
			child.HandleEvent(event)
		}
	}
}

func (g *Group) selectChild(v View) {
	old := g.focused
	if old == v {
		return
	}

	// Set focused BEFORE deselecting old so that blur-validation callbacks
	// (e.g., InputLine re-focusing itself on invalid input) can redirect focus.
	g.focused = v
	if old != nil {
		old.SetState(SfSelected, false)
		if g.focused != v {
			// Blur validation stole focus (e.g., InputLine re-focused itself).
			return
		}
	}
	if v != nil {
		v.SetState(SfSelected, true)
	}

	if old != nil {
		for _, child := range g.children {
			if child.HasState(SfDisabled) {
				continue
			}
			ev := &Event{What: EvBroadcast, Command: CmReleasedFocus, Info: old}
			child.HandleEvent(ev)
		}
		if v != nil {
			for _, child := range g.children {
				if child.HasState(SfDisabled) {
					continue
				}
				ev := &Event{What: EvBroadcast, Command: CmReceivedFocus, Info: v}
				child.HandleEvent(ev)
			}
		}
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

func (g *Group) FocusNext() { g.focusNext() }
func (g *Group) FocusPrev() { g.focusPrev() }

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
