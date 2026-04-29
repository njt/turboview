package tv

import "github.com/gdamore/tcell/v2"

var _ Container = (*Desktop)(nil)

type Desktop struct {
	BaseView
	group   *Group
	pattern rune
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
	if cs := d.ColorScheme(); cs != nil {
		style = cs.DesktopBackground
	}
	buf.Fill(NewRect(0, 0, w, h), d.pattern, style)
	d.group.Draw(buf)
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

	// Delegate to group (three-phase dispatch reaches focused child).
	d.group.HandleEvent(event)

	// After dispatch, handle CmClose at the Desktop level if not yet cleared.
	if event.What == EvCommand && event.Command == CmClose {
		if focused := d.group.FocusedChild(); focused != nil {
			d.Remove(focused)
			event.Clear()
		}
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
