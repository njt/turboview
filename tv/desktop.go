package tv

import "github.com/gdamore/tcell/v2"

type Desktop struct {
	BaseView
	pattern rune
}

func NewDesktop(bounds Rect) *Desktop {
	d := &Desktop{
		pattern: '░',
	}
	d.SetBounds(bounds)
	d.SetState(SfVisible, true)
	d.SetGrowMode(GfGrowAll)
	return d
}

func (d *Desktop) Draw(buf *DrawBuffer) {
	w, h := d.Bounds().Width(), d.Bounds().Height()
	style := tcell.StyleDefault
	if cs := d.ColorScheme(); cs != nil {
		style = cs.DesktopBackground
	}
	buf.Fill(NewRect(0, 0, w, h), d.pattern, style)
}
