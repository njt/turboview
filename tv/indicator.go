package tv

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
)

type Indicator struct {
	BaseView
	line     int
	col      int
	modified bool
}

func NewIndicator(bounds Rect) *Indicator {
	ind := &Indicator{
		line: 1,
		col:  1,
	}
	ind.SetBounds(bounds)
	ind.SetState(SfVisible, true)
	ind.SetOptions(OfPostProcess, true)
	return ind
}

func (ind *Indicator) SetValue(line, col int, modified bool) {
	ind.line = line
	ind.col = col
	ind.modified = modified
}

func (ind *Indicator) Draw(buf *DrawBuffer) {
	cs := ind.ColorScheme()
	style := tcell.StyleDefault
	if cs != nil {
		style = cs.WindowTitle
	}
	w := ind.Bounds().Width()
	buf.Fill(NewRect(0, 0, w, 1), ' ', style)
	text := fmt.Sprintf(" %d:%d", ind.line, ind.col)
	if ind.modified {
		text += "  *"
	}
	for i, ch := range []rune(text) {
		if i < w {
			buf.WriteChar(i, 0, ch, style)
		}
	}
}

func (ind *Indicator) HandleEvent(event *Event) {
	if event.What == EvBroadcast && event.Command == CmIndicatorUpdate {
		if ed, ok := event.Info.(*Editor); ok {
			row, col := ed.Memo.CursorPos()
			ind.SetValue(row+1, col+1, ed.Modified())
		}
	}
}
