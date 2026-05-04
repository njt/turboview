package tv

import "github.com/gdamore/tcell/v2"

type ColorDisplay struct {
	BaseView
	fg   int
	bg   int
	text string
}

func NewColorDisplay(bounds Rect) *ColorDisplay {
	cd := &ColorDisplay{
		fg:   7,
		bg:   0,
		text: " Text Text Text ",
	}
	cd.SetBounds(bounds)
	cd.SetState(SfVisible, true)
	cd.SetSelf(cd)
	return cd
}

func (cd *ColorDisplay) Foreground() int     { return cd.fg }
func (cd *ColorDisplay) Background() int     { return cd.bg }
func (cd *ColorDisplay) SetForeground(i int) { cd.fg = i }
func (cd *ColorDisplay) SetBackground(i int) { cd.bg = i }

func (cd *ColorDisplay) Draw(buf *DrawBuffer) {
	style := tcell.StyleDefault.
		Foreground(tcell.PaletteColor(cd.fg)).
		Background(tcell.PaletteColor(cd.bg))
	for y := 0; y < cd.Bounds().Height(); y++ {
		buf.WriteStr(0, y, cd.text, style)
	}
}

func (cd *ColorDisplay) HandleEvent(event *Event) {
	if event.What == EvBroadcast {
		switch event.Command {
		case CmColorForegroundChanged:
			if idx, ok := event.Info.(int); ok {
				cd.fg = idx
			}
			event.Clear()
			return
		case CmColorBackgroundChanged:
			if idx, ok := event.Info.(int); ok {
				cd.bg = idx
			}
			event.Clear()
			return
		}
	}
	cd.BaseView.HandleEvent(event)
}
