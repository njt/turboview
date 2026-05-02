package tv

import "github.com/gdamore/tcell/v2"

var _ Widget = (*History)(nil)

type History struct {
	BaseView
	link      *InputLine
	historyID int
}

func NewHistory(bounds Rect, link *InputLine, historyID int) *History {
	h := &History{
		link:      link,
		historyID: historyID,
	}
	h.SetBounds(bounds)
	h.SetState(SfVisible, true)
	h.SetOptions(OfPostProcess, true)
	h.SetSelf(h)
	return h
}

func (h *History) Link() *InputLine { return h.link }
func (h *History) HistoryID() int   { return h.historyID }

func (h *History) Draw(buf *DrawBuffer) {
	cs := h.ColorScheme()
	sidesStyle := tcell.StyleDefault
	arrowStyle := tcell.StyleDefault
	if cs != nil {
		sidesStyle = cs.HistorySides
		arrowStyle = cs.HistoryArrow
	}
	buf.WriteChar(0, 0, '▐', sidesStyle)
	buf.WriteChar(1, 0, '↓', arrowStyle)
	buf.WriteChar(2, 0, '▌', sidesStyle)
}

func (h *History) HandleEvent(event *Event) {
	if event.What == EvBroadcast {
		switch event.Command {
		case CmReleasedFocus:
			if event.Info == h.link {
				DefaultHistory.Add(h.historyID, h.link.Text())
			}
		case CmRecordHistory:
			DefaultHistory.Add(h.historyID, h.link.Text())
		}
		return
	}

	if event.What == EvMouse && event.Mouse != nil {
		if event.Mouse.Button&tcell.Button1 != 0 {
			if h.link.Owner() != nil {
				h.link.Owner().SetFocusedChild(h.link)
			}
			if !h.link.HasOption(OfSelectable) {
				event.Clear()
				return
			}
			h.openDropdown(event)
			return
		}
		return
	}

	if event.What == EvKeyboard && event.Key != nil {
		if event.Key.Key == tcell.KeyDown && h.link.HasState(SfSelected) {
			h.openDropdown(event)
			return
		}
	}
}

func (h *History) openDropdown(event *Event) {
	DefaultHistory.Add(h.historyID, h.link.Text())
	entries := DefaultHistory.Entries(h.historyID)
	if len(entries) == 0 {
		event.Clear()
		return
	}

	reversed := make([]string, len(entries))
	for i, e := range entries {
		reversed[len(entries)-1-i] = e
	}

	app := findApp(h)
	desktop := findDesktop(h)
	if app == nil || desktop == nil {
		event.Clear()
		return
	}

	linkAbsX, linkAbsY := viewToDesktop(h.link)
	popupW := h.link.Bounds().Width() + 2
	popupH := len(reversed) + 2
	if popupH > 9 {
		popupH = 9
	}
	popupX := linkAbsX - 1
	popupY := linkAbsY

	popup := newHistoryPopup(NewRect(popupX, popupY, popupW, popupH), reversed)
	desktop.Insert(popup)
	popup.propagateScheme()
	app.drawAndFlush()

	result := h.runPopupLoop(app, popup)

	desktop.Remove(popup)

	if result == CmOK {
		sel := popup.viewer.Selected()
		if sel >= 0 && sel < len(reversed) {
			h.link.SetText(reversed[sel])
			h.link.SelectAll()
		}
	}

	app.drawAndFlush()
	event.Clear()
}

func (h *History) runPopupLoop(app *Application, popup *historyPopup) CommandCode {
	desktop := findDesktop(h)
	desktopOrigin := desktop.Bounds().A
	for {
		ev := app.PollEvent()
		if ev == nil {
			return CmCancel
		}

		if ev.What == EvKeyboard && ev.Key != nil {
			switch ev.Key.Key {
			case tcell.KeyEnter:
				return CmOK
			case tcell.KeyEscape:
				return CmCancel
			default:
				popup.viewer.HandleEvent(ev)
			}
		} else if ev.What == EvMouse && ev.Mouse != nil {
			pb := popup.Bounds()
			mx := ev.Mouse.X - desktopOrigin.X
			my := ev.Mouse.Y - desktopOrigin.Y
			if pb.Contains(NewPoint(mx, my)) {
				ev.Mouse.X = mx - pb.A.X - 1
				ev.Mouse.Y = my - pb.A.Y - 1
				popup.viewer.HandleEvent(ev)
				if popup.confirmed {
					return CmOK
				}
			} else if ev.Mouse.Button&tcell.Button1 != 0 {
				return CmCancel
			}
		}

		app.drawAndFlush()
	}
}

// ---------------------------------------------------------------------------
// historyPopup
// ---------------------------------------------------------------------------

type historyPopup struct {
	BaseView
	viewer    *ListViewer
	scrollbar *ScrollBar
	confirmed bool
}

func newHistoryPopup(bounds Rect, entries []string) *historyPopup {
	hp := &historyPopup{}
	hp.SetBounds(bounds)
	hp.SetState(SfVisible, true)

	clientW := bounds.Width() - 2
	clientH := bounds.Height() - 2

	needsScroll := clientH < len(entries)
	viewerW := clientW
	if needsScroll {
		viewerW = clientW - 1
	}

	ds := NewStringList(entries)
	hp.viewer = NewListViewer(NewRect(0, 0, viewerW, clientH), ds)
	hp.viewer.SetState(SfVisible, true)
	hp.viewer.SetOptions(OfSelectable, true)
	hp.viewer.SetState(SfSelected, true)
	hp.viewer.OnSelect = func(int) { hp.confirmed = true }

	if needsScroll {
		hp.scrollbar = NewScrollBar(NewRect(clientW-1, 0, 1, clientH), Vertical)
		hp.viewer.SetScrollBar(hp.scrollbar)
	}

	hp.SetSelf(hp)
	return hp
}

func (hp *historyPopup) propagateScheme() {
	cs := hp.ColorScheme()
	if cs != nil {
		hp.viewer.scheme = cs
		if hp.scrollbar != nil {
			hp.scrollbar.scheme = cs
		}
	}
}

func (hp *historyPopup) Draw(buf *DrawBuffer) {
	cs := hp.ColorScheme()
	frameStyle := tcell.StyleDefault
	bgStyle := tcell.StyleDefault
	if cs != nil {
		frameStyle = cs.WindowFrameActive
		bgStyle = cs.WindowBackground
	}

	w := hp.Bounds().Width()
	h := hp.Bounds().Height()

	buf.Fill(NewRect(0, 0, w, h), ' ', bgStyle)

	buf.WriteChar(0, 0, '┌', frameStyle)
	buf.WriteChar(w-1, 0, '┐', frameStyle)
	buf.WriteChar(0, h-1, '└', frameStyle)
	buf.WriteChar(w-1, h-1, '┘', frameStyle)
	for x := 1; x < w-1; x++ {
		buf.WriteChar(x, 0, '─', frameStyle)
		buf.WriteChar(x, h-1, '─', frameStyle)
	}
	for y := 1; y < h-1; y++ {
		buf.WriteChar(0, y, '│', frameStyle)
		buf.WriteChar(w-1, y, '│', frameStyle)
	}

	viewerW := w - 2
	if hp.scrollbar != nil {
		viewerW = w - 3
	}
	clientBuf := buf.SubBuffer(NewRect(1, 1, viewerW, h-2))
	hp.viewer.Draw(clientBuf)

	if hp.scrollbar != nil {
		sbBuf := buf.SubBuffer(NewRect(w-2, 1, 1, h-2))
		hp.scrollbar.Draw(sbBuf)
	}
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func viewToDesktop(v View) (int, int) {
	x, y := v.Bounds().A.X, v.Bounds().A.Y
	owner := v.Owner()
	for owner != nil {
		if _, isDesktop := owner.(*Desktop); isDesktop {
			break
		}
		if view, ok := owner.(View); ok {
			b := view.Bounds()
			x += b.A.X
			y += b.A.Y
			if _, isWindow := owner.(*Window); isWindow {
				x += 1
				y += 1
			}
			owner = view.Owner()
		} else {
			break
		}
	}
	return x, y
}

func findDesktop(v View) *Desktop {
	var current Container = v.Owner()
	for current != nil {
		if d, ok := current.(*Desktop); ok {
			return d
		}
		if view, ok := current.(View); ok {
			current = view.Owner()
		} else {
			break
		}
	}
	return nil
}

func findApp(v View) *Application {
	d := findDesktop(v)
	if d != nil {
		return d.App()
	}
	return nil
}
