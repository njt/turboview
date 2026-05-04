package tv

type ColorGroupListViewer struct {
	*ListViewer
	lastSelected int
}

func NewColorGroupListViewer(bounds Rect, ds ListDataSource) *ColorGroupListViewer {
	lv := NewListViewer(bounds, ds)
	return &ColorGroupListViewer{ListViewer: lv, lastSelected: -1}
}

func (cgl *ColorGroupListViewer) HandleEvent(event *Event) {
	cgl.ListViewer.HandleEvent(event)
	sel := cgl.Selected()
	if sel != cgl.lastSelected {
		cgl.lastSelected = sel
		owner := cgl.Owner()
		if owner != nil {
			ev := &Event{What: EvBroadcast, Command: CmNewColorGroup, Info: sel}
			owner.HandleEvent(ev)
		}
	}
}

type ColorItemListViewer struct {
	*ListViewer
	lastSelected int
}

func NewColorItemListViewer(bounds Rect, ds ListDataSource) *ColorItemListViewer {
	lv := NewListViewer(bounds, ds)
	return &ColorItemListViewer{ListViewer: lv, lastSelected: -1}
}

func (cil *ColorItemListViewer) HandleEvent(event *Event) {
	cil.ListViewer.HandleEvent(event)
	sel := cil.Selected()
	if sel != cil.lastSelected {
		cil.lastSelected = sel
		owner := cil.Owner()
		if owner != nil {
			ev := &Event{What: EvBroadcast, Command: CmNewColorIndex, Info: sel}
			owner.HandleEvent(ev)
		}
	}
}
