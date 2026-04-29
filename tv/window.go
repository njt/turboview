package tv

import "github.com/gdamore/tcell/v2"

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
	buf.WriteChar(width-3, 0, '↑', frameStyle)
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
	w.group.HandleEvent(event)
}
