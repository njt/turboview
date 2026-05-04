package tv

import "github.com/gdamore/tcell/v2"

type ColorSelector struct {
	BaseView
	numColors int
	selected  int
}

func NewColorSelector(bounds Rect, numColors int) *ColorSelector {
	cs := &ColorSelector{
		numColors: numColors,
		selected:  0,
	}
	cs.SetBounds(bounds)
	cs.SetState(SfVisible, true)
	cs.SetOptions(OfSelectable, true)
	cs.SetSelf(cs)
	return cs
}

func (cs *ColorSelector) Selected() int       { return cs.selected }
func (cs *ColorSelector) Kind() int           { if cs.numColors == 16 { return 0 }; return 1 }
func (cs *ColorSelector) SetSelected(idx int) {
	if idx < 0 { idx = 0 }
	if idx >= cs.numColors { idx = cs.numColors - 1 }
	cs.selected = idx
}

func (cs *ColorSelector) broadcast() {
	cmd := CmColorForegroundChanged
	if cs.Kind() == 1 {
		cmd = CmColorBackgroundChanged
	}
	owner := cs.Owner()
	if owner != nil {
		ev := &Event{What: EvBroadcast, Command: cmd, Info: cs.selected}
		owner.HandleEvent(ev)
	}
}

func (cs *ColorSelector) Draw(buf *DrawBuffer) {
	scheme := cs.ColorScheme()
	cursor := tcell.StyleDefault
	normal := tcell.StyleDefault
	if scheme != nil {
		cursor = scheme.ColorSelectorCursor
		normal = scheme.ColorSelectorNormal
	}

	for i := 0; i < cs.numColors; i++ {
		col := i % 4
		row := i / 4
		x := col * 3
		y := row

		color := tcell.PaletteColor(i)
		cellStyle := normal.Foreground(color).Background(color)

		if i == cs.selected {
			buf.WriteChar(x+1, y, '█', cursor)
		} else {
			buf.WriteStr(x, y, "   ", cellStyle)
		}
	}
}

func (cs *ColorSelector) HandleEvent(event *Event) {
	cs.BaseView.HandleEvent(event)
	if event.IsCleared() {
		return
	}

	numRows := cs.numColors / 4
	numCols := 4
	oldSelected := cs.selected

	switch {
	case event.What == EvKeyboard && event.Key != nil:
		k := event.Key
		row := cs.selected / 4
		col := cs.selected % 4

		switch k.Key {
		case tcell.KeyLeft:
			col = (col - 1 + numCols) % numCols
		case tcell.KeyRight:
			col = (col + 1) % numCols
		case tcell.KeyUp:
			row = (row - 1 + numRows) % numRows
		case tcell.KeyDown:
			row = (row + 1) % numRows
		default:
			return
		}
		cs.selected = row*4 + col
		if cs.selected >= cs.numColors {
			cs.selected = cs.numColors - 1
		}
		if cs.selected != oldSelected {
			cs.broadcast()
		}
		event.Clear()

	case event.What == EvMouse && event.Mouse != nil:
		mx := event.Mouse.X - cs.Bounds().A.X
		my := event.Mouse.Y - cs.Bounds().A.Y
		col := mx / 3
		row := my
		idx := row*4 + col
		if idx >= 0 && idx < cs.numColors {
			cs.selected = idx
		}
		if cs.selected != oldSelected {
			cs.broadcast()
		}
		event.Clear()
	}
}
