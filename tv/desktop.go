package tv

import "github.com/gdamore/tcell/v2"

var _ Container = (*Desktop)(nil)

type Desktop struct {
	BaseView
	group   *Group
	pattern rune
	app     *Application
}

func NewDesktop(bounds Rect) *Desktop {
	d := &Desktop{
		pattern: '░',
	}
	d.SetBounds(bounds)
	d.SetState(SfVisible, true)
	d.SetGrowMode(GfGrowAll)
	d.group = NewGroup(NewRect(0, 0, bounds.Width(), bounds.Height()))
	d.group.SetFacade(d)
	return d
}

func (d *Desktop) SetBounds(r Rect) {
	d.BaseView.SetBounds(r)
	if d.group != nil {
		d.group.SetBounds(NewRect(0, 0, r.Width(), r.Height()))
	}
}

func (d *Desktop) Insert(v View)               { d.group.Insert(v) }
func (d *Desktop) Remove(v View)               { d.group.Remove(v) }
func (d *Desktop) Children() []View            { return d.group.Children() }
func (d *Desktop) FocusedChild() View          { return d.group.FocusedChild() }
func (d *Desktop) SetFocusedChild(v View)      { d.group.SetFocusedChild(v) }
func (d *Desktop) ExecView(v View) CommandCode { return d.group.ExecView(v) }
func (d *Desktop) BringToFront(v View)         { d.group.BringToFront(v) }

func (d *Desktop) Draw(buf *DrawBuffer) {
	w, h := d.Bounds().Width(), d.Bounds().Height()
	style := tcell.StyleDefault
	shadowStyle := tcell.StyleDefault
	if cs := d.ColorScheme(); cs != nil {
		style = cs.DesktopBackground
		shadowStyle = cs.WindowShadow
	}

	buf.Fill(NewRect(0, 0, w, h), d.pattern, style)

	for _, child := range d.group.Children() {
		if !child.HasState(SfVisible) {
			continue
		}
		cb := child.Bounds()

		// Draw the window
		sub := buf.SubBuffer(cb)
		child.Draw(sub)

		// Draw shadow (2 right, 1 down)
		// Right shadow
		for y := cb.A.Y + 1; y < cb.B.Y+1; y++ {
			for x := cb.B.X; x < cb.B.X+2; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					buf.SetCellStyle(x, y, shadowStyle)
				}
			}
		}
		// Bottom shadow
		for x := cb.A.X + 2; x < cb.B.X+2; x++ {
			y := cb.B.Y
			if x >= 0 && x < w && y >= 0 && y < h {
				buf.SetCellStyle(x, y, shadowStyle)
			}
		}
	}
}

func (d *Desktop) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		d.routeMouseEvent(event)
		// A child may have transformed the mouse event into a command (e.g. CmClose).
		if event.What == EvCommand && event.Command == CmClose {
			if focused := d.group.FocusedChild(); focused != nil {
				d.Remove(focused)
				event.Clear()
			}
		}
		return
	}

	// Alt+N window switching
	if event.What == EvKeyboard && event.Key != nil {
		if event.Key.Modifiers&tcell.ModAlt != 0 && event.Key.Key == tcell.KeyRune {
			n := int(event.Key.Rune - '0')
			if n >= 1 && n <= 9 {
				if d.selectWindowByNumber(n) {
					event.Clear()
				}
				return
			}
		}
		// Tab/Shift+Tab: forward to focused window for widget traversal.
		// Desktop uses CmNext/CmPrev (F6) for window cycling, not Tab.
		if event.Key.Key == tcell.KeyTab || event.Key.Key == tcell.KeyBacktab {
			if focused := d.group.FocusedChild(); focused != nil {
				focused.HandleEvent(event)
			}
			return
		}
	}

	// Desktop-level commands
	if event.What == EvCommand {
		switch event.Command {
		case CmNext:
			d.SelectNextWindow()
			event.Clear()
			return
		case CmPrev:
			d.SelectPrevWindow()
			event.Clear()
			return
		case CmTile:
			d.Tile()
			event.Clear()
			return
		case CmCascade:
			d.Cascade()
			event.Clear()
			return
		}
	}

	d.group.HandleEvent(event)

	// After group dispatch, handle CmClose at Desktop level if not yet cleared.
	if event.What == EvCommand && event.Command == CmClose {
		if focused := d.group.FocusedChild(); focused != nil {
			d.Remove(focused)
			event.Clear()
		}
	}
}

func (d *Desktop) selectWindowByNumber(n int) bool {
	for _, child := range d.group.Children() {
		if w, ok := child.(*Window); ok && w.Number() == n {
			d.BringToFront(w)
			return true
		}
	}
	return false
}

func (d *Desktop) SelectNextWindow() {
	children := d.group.Children()
	if len(children) == 0 {
		return
	}
	current := d.group.FocusedChild()
	if current == nil {
		d.BringToFront(children[0])
		return
	}
	for i, child := range children {
		if child == current {
			next := children[(i+1)%len(children)]
			d.BringToFront(next)
			return
		}
	}
}

func (d *Desktop) SelectPrevWindow() {
	children := d.group.Children()
	if len(children) == 0 {
		return
	}
	current := d.group.FocusedChild()
	if current == nil {
		d.BringToFront(children[len(children)-1])
		return
	}
	for i, child := range children {
		if child == current {
			prev := children[(i-1+len(children))%len(children)]
			d.BringToFront(prev)
			return
		}
	}
}

func (d *Desktop) visibleWindows() []*Window {
	var windows []*Window
	for _, child := range d.group.Children() {
		if w, ok := child.(*Window); ok && w.HasState(SfVisible) {
			windows = append(windows, w)
		}
	}
	return windows
}

func (d *Desktop) Tile() {
	windows := d.visibleWindows()
	n := len(windows)
	if n == 0 {
		return
	}
	dw, dh := d.Bounds().Width(), d.Bounds().Height()

	cols := 1
	for cols*cols < n {
		cols++
	}
	rows := (n + cols - 1) / cols

	cellW := dw / cols
	cellH := dh / rows

	for i, win := range windows {
		col := i % cols
		row := i / cols
		x := col * cellW
		y := row * cellH
		w := cellW
		h := cellH
		if col == cols-1 {
			w = dw - x
		}
		// Last window in its column: absorb remaining height.
		nextInSameCol := i + cols
		if row == rows-1 || nextInSameCol >= n {
			h = dh - y
		}
		if w < 10 {
			w = 10
		}
		if h < 5 {
			h = 5
		}
		win.SetBounds(NewRect(x, y, w, h))
		win.zoomed = false
	}
}

func (d *Desktop) Cascade() {
	windows := d.visibleWindows()
	n := len(windows)
	if n == 0 {
		return
	}
	dw, dh := d.Bounds().Width(), d.Bounds().Height()
	winW := dw * 3 / 4
	winH := dh * 3 / 4
	if winW < 10 {
		winW = 10
	}
	if winH < 5 {
		winH = 5
	}

	for i, win := range windows {
		x := i * 2
		y := i
		if x+winW > dw {
			x = 0
		}
		if y+winH > dh {
			y = 0
		}
		win.SetBounds(NewRect(x, y, winW, winH))
		win.zoomed = false
	}
}

func (d *Desktop) routeMouseEvent(event *Event) {
	mx, my := event.Mouse.X, event.Mouse.Y

	// Mouse capture: if any child is being dragged or resized,
	// route all mouse events to it WITHOUT translating coordinates
	for _, child := range d.group.Children() {
		if child.HasState(SfDragging) {
			child.HandleEvent(event)
			return
		}
		if w, ok := child.(*Window); ok && w.resizing {
			child.HandleEvent(event)
			return
		}
	}

	// Normal hit-testing: front-to-back
	children := d.group.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		if !child.HasState(SfVisible) {
			continue
		}
		bounds := child.Bounds()
		if bounds.Contains(NewPoint(mx, my)) {
			if child.HasOption(OfTopSelect) && event.Mouse.Button&tcell.Button1 != 0 {
				d.BringToFront(child)
			}
			event.Mouse.X -= bounds.A.X
			event.Mouse.Y -= bounds.A.Y
			child.HandleEvent(event)
			return
		}
	}
}
