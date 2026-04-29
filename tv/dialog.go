package tv

import "github.com/gdamore/tcell/v2"

type MsgBoxButton int

const (
	MbOK     MsgBoxButton = 1 << iota
	MbCancel
	MbYes
	MbNo
)

func MessageBox(owner Container, title, text string, buttons MsgBoxButton) CommandCode {
	type btnDef struct {
		label string
		cmd   CommandCode
	}
	var defs []btnDef
	if buttons&MbYes != 0 {
		defs = append(defs, btnDef{"Yes", CmYes})
	}
	if buttons&MbNo != 0 {
		defs = append(defs, btnDef{"No", CmNo})
	}
	if buttons&MbOK != 0 {
		defs = append(defs, btnDef{"OK", CmOK})
	}
	if buttons&MbCancel != 0 {
		defs = append(defs, btnDef{"Cancel", CmCancel})
	}
	if len(defs) == 0 {
		defs = append(defs, btnDef{"OK", CmOK})
	}

	// Auto-size
	textRunes := []rune(text)
	textW := len(textRunes)
	btnW := 12
	btnGap := 2
	buttonRowW := len(defs)*btnW + (len(defs)-1)*btnGap
	contentW := textW
	if buttonRowW > contentW {
		contentW = buttonRowW
	}
	titleW := len([]rune(title)) + 4
	if titleW > contentW {
		contentW = titleW
	}
	dialogW := contentW + 4 // 2 for frame + 2 for padding
	if dialogW > 60 {
		dialogW = 60
	}
	if dialogW < 20 {
		dialogW = 20
	}

	// Text wrapping height
	innerW := dialogW - 4
	textLines := 1
	lineLen := 0
	for _, r := range textRunes {
		if r == '\n' {
			textLines++
			lineLen = 0
			continue
		}
		lineLen++
		if lineLen > innerW {
			textLines++
			lineLen = 1
		}
	}

	dialogH := textLines + 5 // top frame + text rows + gap + button row + bottom frame
	if dialogH < 6 {
		dialogH = 6
	}

	// Center in owner
	ob := owner.Bounds()
	dx := (ob.Width() - dialogW) / 2
	dy := (ob.Height() - dialogH) / 2
	if dx < 0 {
		dx = 0
	}
	if dy < 0 {
		dy = 0
	}

	dlg := NewDialog(NewRect(dx, dy, dialogW, dialogH), title)

	// Insert static text
	st := NewStaticText(NewRect(1, 0, innerW, textLines), text)
	dlg.Insert(st)

	// Insert buttons
	btnY := textLines + 1
	totalBtnW := len(defs)*btnW + (len(defs)-1)*btnGap
	startX := (innerW - totalBtnW) / 2
	if startX < 0 {
		startX = 0
	}
	var firstBtn *Button
	for i, def := range defs {
		x := startX + i*(btnW+btnGap)
		var opts []ButtonOption
		if i == 0 {
			opts = append(opts, WithDefault())
		}
		btn := NewButton(NewRect(x, btnY, btnW, 2), def.label, def.cmd, opts...)
		dlg.Insert(btn)
		if i == 0 {
			firstBtn = btn
		}
	}
	if firstBtn != nil {
		dlg.SetFocusedChild(firstBtn)
	}

	return owner.ExecView(dlg)
}

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
