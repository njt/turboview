package tv

var _ Container = (*ListBox)(nil)

type ListBox struct {
	BaseView
	group     *Group
	viewer    *ListViewer
	scrollbar *ScrollBar
}

func NewListBox(bounds Rect, ds ListDataSource) *ListBox {
	lb := &ListBox{}
	lb.SetBounds(bounds)
	lb.SetState(SfVisible, true)
	lb.SetOptions(OfSelectable, true)

	lb.group = NewGroup(bounds)
	lb.group.SetFacade(lb)

	lb.viewer = NewListViewer(NewRect(0, 0, bounds.Width()-1, bounds.Height()), ds)
	lb.viewer.SetGrowMode(GfGrowHiX | GfGrowHiY)

	lb.scrollbar = NewScrollBar(NewRect(bounds.Width()-1, 0, 1, bounds.Height()), Vertical)
	lb.scrollbar.SetGrowMode(GfGrowLoX | GfGrowHiY)

	lb.viewer.SetScrollBar(lb.scrollbar)

	lb.group.Insert(lb.viewer)
	lb.group.Insert(lb.scrollbar)

	lb.SetSelf(lb)
	return lb
}

func NewStringListBox(bounds Rect, items []string) *ListBox {
	return NewListBox(bounds, NewStringList(items))
}

func (lb *ListBox) ListViewer() *ListViewer    { return lb.viewer }
func (lb *ListBox) ScrollBar() *ScrollBar      { return lb.scrollbar }
func (lb *ListBox) Selected() int              { return lb.viewer.Selected() }
func (lb *ListBox) SetSelected(index int)      { lb.viewer.SetSelected(index) }
func (lb *ListBox) DataSource() ListDataSource { return lb.viewer.DataSource() }

func (lb *ListBox) SetDataSource(ds ListDataSource) {
	lb.viewer.SetDataSource(ds)
}

func (lb *ListBox) Insert(v View)               { lb.group.Insert(v) }
func (lb *ListBox) Remove(v View)               { lb.group.Remove(v) }
func (lb *ListBox) Children() []View            { return lb.group.Children() }
func (lb *ListBox) FocusedChild() View          { return lb.group.FocusedChild() }
func (lb *ListBox) SetFocusedChild(v View)      { lb.group.SetFocusedChild(v) }
func (lb *ListBox) ExecView(v View) CommandCode { return lb.group.ExecView(v) }
func (lb *ListBox) BringToFront(v View)         { lb.group.BringToFront(v) }

func (lb *ListBox) Draw(buf *DrawBuffer) {
	for _, child := range lb.group.Children() {
		childBounds := child.Bounds()
		sub := buf.SubBuffer(childBounds)
		child.Draw(sub)
	}
}

func (lb *ListBox) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		for _, child := range lb.group.Children() {
			if child.Bounds().Contains(NewPoint(event.Mouse.X, event.Mouse.Y)) {
				origX, origY := event.Mouse.X, event.Mouse.Y
				event.Mouse.X -= child.Bounds().A.X
				event.Mouse.Y -= child.Bounds().A.Y
				child.HandleEvent(event)
				event.Mouse.X, event.Mouse.Y = origX, origY
				return
			}
		}
		return
	}
	lb.group.HandleEvent(event)
}
