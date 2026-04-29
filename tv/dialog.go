package tv

import "github.com/gdamore/tcell/v2"

var _ Container = (*Dialog)(nil)

type Dialog struct {
	BaseView
	group *Group
	title string
}

type DialogOption func(*Dialog)

func NewDialog(bounds Rect, title string, opts ...DialogOption) *Dialog {
	d := &Dialog{
		title: title,
	}
	d.SetBounds(bounds)
	d.SetState(SfVisible, true)
	d.SetOptions(OfSelectable, true)

	cw := max(bounds.Width()-2, 0)
	ch := max(bounds.Height()-2, 0)
	d.group = NewGroup(NewRect(0, 0, cw, ch))
	d.group.SetFacade(d)

	for _, opt := range opts {
		opt(d)
	}
	return d
}

func (d *Dialog) Title() string { return d.title }

func (d *Dialog) SetBounds(r Rect) {
	d.BaseView.SetBounds(r)
	if d.group != nil {
		cw := max(r.Width()-2, 0)
		ch := max(r.Height()-2, 0)
		d.group.SetBounds(NewRect(0, 0, cw, ch))
	}
}

func (d *Dialog) Insert(v View)               { d.group.Insert(v) }
func (d *Dialog) Remove(v View)               { d.group.Remove(v) }
func (d *Dialog) Children() []View            { return d.group.Children() }
func (d *Dialog) FocusedChild() View          { return d.group.FocusedChild() }
func (d *Dialog) SetFocusedChild(v View)      { d.group.SetFocusedChild(v) }
func (d *Dialog) ExecView(v View) CommandCode { return d.group.ExecView(v) }
func (d *Dialog) BringToFront(v View)         { d.group.BringToFront(v) }

func (d *Dialog) Draw(buf *DrawBuffer) {
	width, height := d.Bounds().Width(), d.Bounds().Height()
	if width < 8 || height < 3 {
		return
	}

	cs := d.ColorScheme()
	frameStyle := tcell.StyleDefault
	bgStyle := tcell.StyleDefault
	titleStyle := tcell.StyleDefault
	if cs != nil {
		frameStyle = cs.DialogFrame
		bgStyle = cs.DialogBackground
		titleStyle = cs.WindowTitle
	}

	// Client area background
	buf.Fill(NewRect(1, 1, width-2, height-2), ' ', bgStyle)

	// Frame — always double-line for dialogs
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

	// Title centered in top border
	if len(d.title) > 0 {
		runes := []rune(d.title)
		availW := width - 2
		if len(runes) > availW-2 {
			runes = runes[:availW-2]
		}
		padded := " " + string(runes) + " "
		runeLen := len([]rune(padded))
		titleX := 1 + (availW-runeLen)/2
		if titleX < 1 {
			titleX = 1
		}
		buf.WriteStr(titleX, 0, padded, titleStyle)
	}

	// Draw children in client area
	clientBuf := buf.SubBuffer(NewRect(1, 1, width-2, height-2))
	d.group.Draw(clientBuf)
}

func (d *Dialog) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		width, height := d.Bounds().Width(), d.Bounds().Height()
		mx, my := event.Mouse.X, event.Mouse.Y
		// Client area: forward to group with frame offset
		if mx > 0 && mx < width-1 && my > 0 && my < height-1 {
			event.Mouse.X -= 1
			event.Mouse.Y -= 1
			d.group.HandleEvent(event)
		}
		return
	}
	d.group.HandleEvent(event)
}
