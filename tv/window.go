package tv

import (
	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

var _ Container = (*Window)(nil)

type Window struct {
	BaseView
	group      *Group
	title      string
	number     int
	zoomed     bool
	zoomBounds Rect
	dragOff    Point
	resizing   bool
	resizeLeft bool
}

type WindowOption func(*Window)

func WithWindowNumber(n int) WindowOption {
	return func(w *Window) { w.number = n }
}

func NewWindow(bounds Rect, title string, opts ...WindowOption) *Window {
	w := &Window{
		title: title,
	}
	w.SetBounds(bounds)
	w.SetState(SfVisible, true)
	w.SetOptions(OfSelectable|OfTopSelect, true)

	cw := max(bounds.Width()-2, 0)
	ch := max(bounds.Height()-2, 0)
	w.group = NewGroup(NewRect(0, 0, cw, ch))
	w.group.SetFacade(w)

	for _, opt := range opts {
		opt(w)
	}
	return w
}

func (w *Window) Title() string     { return w.title }
func (w *Window) SetTitle(t string) { w.title = t }
func (w *Window) Number() int       { return w.number }
func (w *Window) SetColorScheme(cs *theme.ColorScheme) {
	w.scheme = cs
}

func (w *Window) Insert(v View)               { w.group.Insert(v) }
func (w *Window) Remove(v View)               { w.group.Remove(v) }
func (w *Window) Children() []View            { return w.group.Children() }
func (w *Window) FocusedChild() View          { return w.group.FocusedChild() }
func (w *Window) SetFocusedChild(v View)      { w.group.SetFocusedChild(v) }
func (w *Window) ExecView(v View) CommandCode { return w.group.ExecView(v) }

func (w *Window) SetBounds(r Rect) {
	w.BaseView.SetBounds(r)
	if w.group != nil {
		cw := max(r.Width()-2, 0)
		ch := max(r.Height()-2, 0)
		w.group.SetBounds(NewRect(0, 0, cw, ch))
	}
}

func (w *Window) Draw(buf *DrawBuffer) {
	width, height := w.Bounds().Width(), w.Bounds().Height()
	if width < 8 || height < 3 {
		return
	}

	cs := w.ColorScheme()
	active := w.HasState(SfSelected)

	frameStyle := tcell.StyleDefault
	titleStyle := tcell.StyleDefault
	bgStyle := tcell.StyleDefault
	if cs != nil {
		if active {
			frameStyle = cs.WindowFrameActive
		} else {
			frameStyle = cs.WindowFrameInactive
		}
		titleStyle = cs.WindowTitle
		bgStyle = cs.WindowBackground
	}

	// Client area background
	buf.Fill(NewRect(1, 1, width-2, height-2), ' ', bgStyle)

	// Frame
	if active {
		buf.WriteChar(0, 0, '╔', frameStyle)
		buf.WriteChar(width-1, 0, '╗', frameStyle)
		buf.WriteChar(0, height-1, '╚', frameStyle)
		buf.WriteChar(width-1, height-1, '╝', frameStyle)
		for x := 1; x < width-1; x++ {
			buf.WriteChar(x, 0, '═', frameStyle)
			buf.WriteChar(x, height-1, '═', frameStyle)
		}
		for y := 1; y < height-1; y++ {
			buf.WriteChar(0, y, '║', frameStyle)
			buf.WriteChar(width-1, y, '║', frameStyle)
		}
	} else {
		buf.WriteChar(0, 0, '┌', frameStyle)
		buf.WriteChar(width-1, 0, '┐', frameStyle)
		buf.WriteChar(0, height-1, '└', frameStyle)
		buf.WriteChar(width-1, height-1, '┘', frameStyle)
		for x := 1; x < width-1; x++ {
			buf.WriteChar(x, 0, '─', frameStyle)
			buf.WriteChar(x, height-1, '─', frameStyle)
		}
		for y := 1; y < height-1; y++ {
			buf.WriteChar(0, y, '│', frameStyle)
			buf.WriteChar(width-1, y, '│', frameStyle)
		}
	}

	// Close icon [×] at (1,0)-(3,0)
	buf.WriteChar(1, 0, '[', frameStyle)
	buf.WriteChar(2, 0, '×', frameStyle)
	buf.WriteChar(3, 0, ']', frameStyle)

	// Zoom icon [↑] at (width-4, 0)-(width-2, 0)
	buf.WriteChar(width-4, 0, '[', frameStyle)
	if w.zoomed {
		buf.WriteChar(width-3, 0, '↕', frameStyle)
	} else {
		buf.WriteChar(width-3, 0, '↑', frameStyle)
	}
	buf.WriteChar(width-2, 0, ']', frameStyle)

	// Title centered between icons
	availStart := 4
	availEnd := width - 4
	availW := availEnd - availStart
	if availW > 0 && len(w.title) > 0 {
		runes := []rune(w.title)
		if len(runes) > availW-2 {
			runes = runes[:availW-2]
		}
		padded := " " + string(runes) + " "
		runeLen := len([]rune(padded))
		titleX := availStart + (availW-runeLen)/2
		if titleX < availStart {
			titleX = availStart
		}
		buf.WriteStr(titleX, 0, padded, titleStyle)
	}

	// Draw children in client area
	clientBuf := buf.SubBuffer(NewRect(1, 1, width-2, height-2))
	w.group.Draw(clientBuf)
}

func (w *Window) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		w.handleMouseEvent(event)
		return
	}

	if event.What == EvBroadcast && event.Command == CmSelectWindowNum {
		if n, ok := event.Info.(int); ok && n == w.number && w.HasOption(OfSelectable) {
			if owner := w.Owner(); owner != nil {
				type fronter interface{ BringToFront(View) }
				if f, ok := owner.(fronter); ok {
					f.BringToFront(w)
				} else {
					owner.SetFocusedChild(w)
				}
			}
			event.Clear()
		}
		return
	}

	if event.What == EvKeyboard && event.Key != nil {
		if event.Key.Key == tcell.KeyTab && event.Key.Modifiers == 0 {
			w.group.FocusNext()
			event.Clear()
			return
		}
		if event.Key.Key == tcell.KeyBacktab {
			w.group.FocusPrev()
			event.Clear()
			return
		}
	}

	// Modal window: CmClose → CmCancel
	if event.What == EvCommand && event.Command == CmClose && w.HasState(SfModal) {
		event.Command = CmCancel
		return
	}

	if event.What == EvCommand {
		switch event.Command {
		case CmZoom:
			w.Zoom()
			event.Clear()
			return
		case CmResize:
			event.Clear()
			return
		}
	}

	w.group.HandleEvent(event)
}

// IsZoomed returns true if the window is currently zoomed.
func (w *Window) IsZoomed() bool { return w.zoomed }

// Zoom toggles the window's zoom state. When zooming in the current bounds
// are stored and the window expands to fill the owner's area; when zooming
// out the stored bounds are restored.
func (w *Window) Zoom() {
	if w.zoomed {
		w.SetBounds(w.zoomBounds)
		w.zoomed = false
	} else {
		w.zoomBounds = w.Bounds()
		if owner := w.Owner(); owner != nil {
			ob := owner.Bounds()
			w.SetBounds(NewRect(0, 0, ob.Width(), ob.Height()))
		}
		w.zoomed = true
	}
}

func (w *Window) handleMouseEvent(event *Event) {
	mx, my := event.Mouse.X, event.Mouse.Y
	width, height := w.Bounds().Width(), w.Bounds().Height()

	// During drag: coordinates are in Desktop-local space
	if w.HasState(SfDragging) {
		if event.Mouse.Button&tcell.Button1 != 0 {
			bounds := w.Bounds()
			newX := mx - w.dragOff.X
			newY := my - w.dragOff.Y
			w.SetBounds(NewRect(newX, newY, bounds.Width(), bounds.Height()))
		} else {
			w.SetState(SfDragging, false)
		}
		event.Clear()
		return
	}

	// During resize: coordinates are in Desktop-local space
	if w.resizing {
		if event.Mouse.Button&tcell.Button1 != 0 {
			bounds := w.Bounds()
			if w.resizeLeft {
				newX := mx
				newW := bounds.B.X - newX
				newH := my - bounds.A.Y + 1
				if newW < 10 {
					newW = 10
					newX = bounds.B.X - newW
				}
				if newH < 5 {
					newH = 5
				}
				w.SetBounds(NewRect(newX, bounds.A.Y, newW, newH))
			} else {
				newW := mx - bounds.A.X + 1
				newH := my - bounds.A.Y + 1
				if newW < 10 {
					newW = 10
				}
				if newH < 5 {
					newH = 5
				}
				w.SetBounds(NewRect(bounds.A.X, bounds.A.Y, newW, newH))
			}
		} else {
			w.resizing = false
		}
		event.Clear()
		return
	}

	// Frame clicks (window-local coordinates)
	if event.Mouse.Button&tcell.Button1 != 0 {
		// Top border
		if my == 0 {
			// Close icon [×] at (1-3, 0)
			if mx >= 1 && mx <= 3 {
				event.What = EvCommand
				event.Command = CmClose
				event.Mouse = nil
				return
			}
			// Zoom icon at (width-4 to width-2, 0)
			if mx >= width-4 && mx <= width-2 {
				w.Zoom()
				event.Clear()
				return
			}
			// Double-click on title bar: zoom
			if event.Mouse.ClickCount >= 2 {
				w.Zoom()
				event.Clear()
				return
			}
			// Title bar: start drag
			w.SetState(SfDragging, true)
			w.dragOff = NewPoint(mx, my)
			event.Clear()
			return
		}

		// Bottom-right corner: start resize
		if mx == width-1 && my == height-1 {
			w.resizing = true
			w.resizeLeft = false
			event.Clear()
			return
		}

		// Bottom-left corner: start resize (left edge)
		if mx == 0 && my == height-1 {
			w.resizing = true
			w.resizeLeft = true
			event.Clear()
			return
		}
	}

	// Client area: forward to group
	if mx > 0 && mx < width-1 && my > 0 && my < height-1 {
		event.Mouse.X -= 1
		event.Mouse.Y -= 1
		w.group.HandleEvent(event)
		return
	}
}
