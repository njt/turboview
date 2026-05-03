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
	numCols    int // default 1
	scrollBar  *ScrollBar
	dragging   bool
	OnSelect   func(int)
}

func NewListViewer(bounds Rect, dataSource ListDataSource) *ListViewer {
	lv := &ListViewer{dataSource: dataSource, numCols: 1}
	lv.SetBounds(bounds)
	lv.SetState(SfVisible, true)
	lv.SetOptions(OfSelectable|OfFirstClick, true)
	lv.SetSelf(lv)
	return lv
}

func (lv *ListViewer) Selected() int              { return lv.selected }
func (lv *ListViewer) TopIndex() int              { return lv.topIndex }
func (lv *ListViewer) DataSource() ListDataSource { return lv.dataSource }
func (lv *ListViewer) NumCols() int               { return lv.numCols }

func (lv *ListViewer) SetNumCols(n int) {
	if n < 1 {
		n = 1
	}
	lv.numCols = n
	lv.ensureVisible()
	lv.syncScrollBar()
}

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

func (lv *ListViewer) itemsPerPage() int {
	return lv.numCols * lv.visibleHeight()
}

func (lv *ListViewer) colWidth() int {
	if lv.numCols <= 1 {
		return lv.Bounds().Width()
	}
	return lv.Bounds().Width() / lv.numCols
}

func (lv *ListViewer) ensureVisible() {
	vh := lv.visibleHeight()
	if vh <= 0 {
		return
	}
	ipp := lv.itemsPerPage()
	if lv.selected < lv.topIndex {
		lv.topIndex = lv.selected
	}
	if lv.selected >= lv.topIndex+ipp {
		lv.topIndex = lv.selected - ipp + 1
	}
	lv.clampTopIndex()
}

func (lv *ListViewer) clampTopIndex() {
	count := lv.dataSource.Count()
	ipp := lv.itemsPerPage()
	maxTop := count - ipp
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
	lv.scrollBar.SetPageSize(lv.itemsPerPage())
	lv.scrollBar.SetValue(lv.topIndex)
	vh := lv.visibleHeight()
	if vh < 1 {
		vh = 1
	}
	// Only set arStep if it hasn't been customised (i.e. still at the default of 1).
	if lv.scrollBar.ArStep() == 1 {
		lv.scrollBar.SetArStep(vh)
	}
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
	if count == 0 {
		text := "<empty>"
		for i, ch := range text {
			if i < w {
				buf.WriteChar(i, 0, ch, normalStyle)
			}
		}
		return
	}
	hasFocus := lv.HasState(SfSelected)
	cw := lv.colWidth()

	for col := 0; col < lv.numCols; col++ {
		colX := col * cw
		isLastCol := col == lv.numCols-1
		// Draw width for item text: last col uses full colWidth, others leave room for divider
		drawWidth := cw
		if !isLastCol {
			drawWidth = cw - 1
		}

		for row := 0; row < vh; row++ {
			idx := lv.topIndex + col*vh + row
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
				// Highlight spans only this column's width
				for x := 0; x < drawWidth; x++ {
					buf.WriteChar(colX+x, row, ' ', style)
				}
			}

			text := lv.dataSource.Item(idx)
			runes := []rune(text)
			if len(runes) > drawWidth {
				runes = runes[:drawWidth]
			}
			for i, ch := range runes {
				buf.WriteChar(colX+i, row, ch, style)
			}
		}

		// Draw divider after non-last columns
		if !isLastCol {
			divX := colX + cw - 1
			for row := 0; row < vh; row++ {
				buf.WriteChar(divX, row, '│', normalStyle)
			}
		}
	}
}

func (lv *ListViewer) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		lv.BaseView.HandleEvent(event)
		if event.IsCleared() {
			return
		}
		if event.Mouse.Button&tcell.Button1 != 0 {
			count := lv.dataSource.Count()
			my := event.Mouse.Y
			mx := event.Mouse.X

			if my < 0 {
				// Auto-scroll up
				if lv.topIndex > 0 {
					lv.topIndex--
					lv.selected = lv.topIndex
				}
			} else if my >= lv.visibleHeight() {
				// Auto-scroll down
				maxTop := count - lv.visibleHeight()
				if maxTop < 0 {
					maxTop = 0
				}
				if lv.topIndex < maxTop {
					lv.topIndex++
				}
				lastVisible := lv.topIndex + lv.visibleHeight() - 1
				if lastVisible >= count {
					lastVisible = count - 1
				}
				if lastVisible >= 0 {
					lv.selected = lastVisible
				}
			} else {
				// Normal click/drag within bounds — determine column from x
				cw := lv.colWidth()
				col := 0
				if cw > 0 {
					col = mx / cw
				}
				if col >= lv.numCols {
					col = lv.numCols - 1
				}
				clickIdx := lv.topIndex + col*lv.visibleHeight() + my
				if clickIdx >= 0 && clickIdx < count {
					lv.selected = clickIdx
				}
			}

			if !lv.dragging && event.Mouse.ClickCount >= 1 {
				lv.dragging = true
			}

			lv.ensureVisible()
			lv.syncScrollBar()

			// Double-click fires OnSelect (only on in-bounds click within data range)
			if event.Mouse.ClickCount >= 2 && my >= 0 && my < lv.visibleHeight() &&
				lv.topIndex+my < count && lv.OnSelect != nil {
				lv.OnSelect(lv.selected)
			}

			event.Clear()
		} else if lv.dragging {
			lv.dragging = false
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

	ke := event.Key
	if ke.Key == tcell.KeyPgUp && ke.Modifiers&tcell.ModCtrl != 0 {
		lv.selected = 0
		lv.topIndex = 0
		lv.syncScrollBar()
		event.Clear()
		return
	}
	if ke.Key == tcell.KeyPgDn && ke.Modifiers&tcell.ModCtrl != 0 {
		lv.selected = count - 1
		if lv.selected < 0 {
			lv.selected = 0
		}
		lv.ensureVisible()
		lv.syncScrollBar()
		event.Clear()
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

	case tcell.KeyLeft:
		vh := lv.visibleHeight()
		if lv.selected >= vh {
			lv.selected -= vh
			lv.ensureVisible()
			lv.syncScrollBar()
		}
		event.Clear()

	case tcell.KeyRight:
		vh := lv.visibleHeight()
		if lv.selected+vh < count {
			lv.selected += vh
			lv.ensureVisible()
			lv.syncScrollBar()
		}
		event.Clear()

	case tcell.KeyHome:
		lv.selected = lv.topIndex
		lv.syncScrollBar()
		event.Clear()

	case tcell.KeyEnd:
		lastVisible := lv.topIndex + lv.itemsPerPage() - 1
		if lastVisible >= count {
			lastVisible = count - 1
		}
		lv.selected = lastVisible
		lv.syncScrollBar()
		event.Clear()

	case tcell.KeyPgDn:
		lv.selected += lv.itemsPerPage()
		if lv.selected >= count {
			lv.selected = count - 1
		}
		lv.ensureVisible()
		lv.syncScrollBar()
		event.Clear()

	case tcell.KeyPgUp:
		lv.selected -= lv.itemsPerPage()
		if lv.selected < 0 {
			lv.selected = 0
		}
		lv.ensureVisible()
		lv.syncScrollBar()
		event.Clear()
	}
}
