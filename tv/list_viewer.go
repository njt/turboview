package tv

import "github.com/gdamore/tcell/v2"

var _ Widget = (*ListViewer)(nil)

type ListDataSource interface {
	Count() int
	Item(index int) string
}

type StringList struct {
	items []string
}

func NewStringList(items []string) *StringList {
	return &StringList{items: items}
}

func (sl *StringList) Count() int            { return len(sl.items) }
func (sl *StringList) Item(index int) string { return sl.items[index] }

type ListViewer struct {
	BaseView
	dataSource ListDataSource
	selected   int
	topIndex   int
	scrollBar  *ScrollBar
	OnSelect   func(int)
}

func NewListViewer(bounds Rect, dataSource ListDataSource) *ListViewer {
	lv := &ListViewer{dataSource: dataSource}
	lv.SetBounds(bounds)
	lv.SetState(SfVisible, true)
	lv.SetOptions(OfSelectable, true)
	lv.SetSelf(lv)
	return lv
}

func (lv *ListViewer) Selected() int              { return lv.selected }
func (lv *ListViewer) TopIndex() int              { return lv.topIndex }
func (lv *ListViewer) DataSource() ListDataSource { return lv.dataSource }

func (lv *ListViewer) SetSelected(index int) {
	count := lv.dataSource.Count()
	if count == 0 {
		lv.selected = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	lv.selected = index
	lv.ensureVisible()
	lv.syncScrollBar()
}

func (lv *ListViewer) SetDataSource(ds ListDataSource) {
	lv.dataSource = ds
	lv.selected = 0
	lv.topIndex = 0
	lv.syncScrollBar()
}

func (lv *ListViewer) SetScrollBar(sb *ScrollBar) {
	if lv.scrollBar != nil {
		lv.scrollBar.OnChange = nil
	}
	lv.scrollBar = sb
	if sb != nil {
		sb.OnChange = func(val int) {
			lv.topIndex = val
			lv.clampTopIndex()
		}
		lv.syncScrollBar()
	}
}

func (lv *ListViewer) visibleHeight() int {
	return lv.Bounds().Height()
}

func (lv *ListViewer) ensureVisible() {
	vh := lv.visibleHeight()
	if vh <= 0 {
		return
	}
	if lv.selected < lv.topIndex {
		lv.topIndex = lv.selected
	}
	if lv.selected >= lv.topIndex+vh {
		lv.topIndex = lv.selected - vh + 1
	}
	lv.clampTopIndex()
}

func (lv *ListViewer) clampTopIndex() {
	count := lv.dataSource.Count()
	vh := lv.visibleHeight()
	maxTop := count - vh
	if maxTop < 0 {
		maxTop = 0
	}
	if lv.topIndex > maxTop {
		lv.topIndex = maxTop
	}
	if lv.topIndex < 0 {
		lv.topIndex = 0
	}
}

func (lv *ListViewer) syncScrollBar() {
	if lv.scrollBar == nil {
		return
	}
	count := lv.dataSource.Count()
	lv.scrollBar.SetRange(0, count)
	lv.scrollBar.SetPageSize(lv.visibleHeight())
	lv.scrollBar.SetValue(lv.topIndex)
}

func (lv *ListViewer) Draw(buf *DrawBuffer) {
	w := lv.Bounds().Width()
	vh := lv.visibleHeight()
	cs := lv.ColorScheme()
	normalStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault
	focusedStyle := tcell.StyleDefault
	if cs != nil {
		normalStyle = cs.ListNormal
		selectedStyle = cs.ListSelected
		focusedStyle = cs.ListFocused
	}

	buf.Fill(NewRect(0, 0, w, vh), ' ', normalStyle)

	count := lv.dataSource.Count()
	hasFocus := lv.HasState(SfSelected)

	for row := 0; row < vh; row++ {
		idx := lv.topIndex + row
		if idx >= count {
			break
		}

		style := normalStyle
		if idx == lv.selected {
			if hasFocus {
				style = focusedStyle
			} else {
				style = selectedStyle
			}
			for x := 0; x < w; x++ {
				buf.WriteChar(x, row, ' ', style)
			}
		}

		text := lv.dataSource.Item(idx)
		runes := []rune(text)
		if len(runes) > w {
			runes = runes[:w]
		}
		for i, ch := range runes {
			buf.WriteChar(i, row, ch, style)
		}
	}
}

func (lv *ListViewer) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		if event.Mouse.Button&tcell.Button1 != 0 {
			clickIdx := lv.topIndex + event.Mouse.Y
			if clickIdx >= 0 && clickIdx < lv.dataSource.Count() {
				lv.selected = clickIdx
				lv.ensureVisible()
				lv.syncScrollBar()
				if lv.OnSelect != nil {
					lv.OnSelect(lv.selected)
				}
			}
			event.Clear()
		}
		return
	}

	if event.What != EvKeyboard || event.Key == nil {
		return
	}

	if !lv.HasState(SfSelected) {
		return
	}

	count := lv.dataSource.Count()

	// Space to select
	if event.Key.Key == tcell.KeyRune && event.Key.Rune == ' ' {
		if count > 0 && lv.OnSelect != nil {
			lv.OnSelect(lv.selected)
		}
		event.Clear()
		return
	}

	// Enter to select
	if event.Key.Key == tcell.KeyEnter {
		if count > 0 && lv.OnSelect != nil {
			lv.OnSelect(lv.selected)
		}
		event.Clear()
		return
	}

	if count == 0 {
		return
	}

	switch event.Key.Key {
	case tcell.KeyDown:
		if lv.selected < count-1 {
			lv.selected++
			lv.ensureVisible()
			lv.syncScrollBar()
		}
		event.Clear()

	case tcell.KeyUp:
		if lv.selected > 0 {
			lv.selected--
			lv.ensureVisible()
			lv.syncScrollBar()
		}
		event.Clear()

	case tcell.KeyHome:
		lv.selected = 0
		lv.topIndex = 0
		lv.syncScrollBar()
		event.Clear()

	case tcell.KeyEnd:
		lv.selected = count - 1
		lv.ensureVisible()
		lv.syncScrollBar()
		event.Clear()

	case tcell.KeyPgDn:
		lv.selected += lv.visibleHeight()
		if lv.selected >= count {
			lv.selected = count - 1
		}
		lv.ensureVisible()
		lv.syncScrollBar()
		event.Clear()

	case tcell.KeyPgUp:
		lv.selected -= lv.visibleHeight()
		if lv.selected < 0 {
			lv.selected = 0
		}
		lv.ensureVisible()
		lv.syncScrollBar()
		event.Clear()
	}
}
