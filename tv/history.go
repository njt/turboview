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
