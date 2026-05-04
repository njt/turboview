package tv

import "github.com/gdamore/tcell/v2"

type monoOption struct {
	name string
	attr int
}

var monoOptions = []monoOption{
	{"Normal", 0x07},
	{"Highlight", 0x0F},
	{"Underline", 0x01},
	{"Inverse", 0x70},
}

type MonoSelector struct {
	BaseView
	selected int
}

func NewMonoSelector(bounds Rect) *MonoSelector {
	ms := &MonoSelector{selected: 0}
	ms.SetBounds(bounds)
	ms.SetState(SfVisible, true)
	ms.SetOptions(OfSelectable, true)
	ms.SetSelf(ms)
	return ms
}

func (ms *MonoSelector) Selected() int       { return ms.selected }
func (ms *MonoSelector) SetSelected(i int)   { if i >= 0 && i < 4 { ms.selected = i } }

func (ms *MonoSelector) Draw(buf *DrawBuffer) {
	scheme := ms.ColorScheme()
	normalStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault
	if scheme != nil {
		normalStyle = scheme.RadioButtonNormal
		selectedStyle = scheme.RadioButtonSelected
	}
	for i, opt := range monoOptions {
		style := normalStyle
		marker := "( )"
		if i == ms.selected {
			style = selectedStyle
			marker = "(•)"
		}
		attrStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
		switch opt.attr {
		case 0x07:
			attrStyle = tcell.StyleDefault.Foreground(tcell.ColorLightGray).Background(tcell.ColorBlack)
		case 0x0F:
			attrStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
		case 0x01:
			attrStyle = tcell.StyleDefault.Foreground(tcell.ColorLightGray).Background(tcell.ColorBlack).Underline(true)
		case 0x70:
			attrStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorLightGray)
		}
		_ = style // placeholder - use scheme.RadioButtonNormal/Selected for future styling
		buf.WriteStr(0, i, marker+" "+opt.name, attrStyle)
	}
}

func (ms *MonoSelector) HandleEvent(event *Event) {
	ms.BaseView.HandleEvent(event)
	if event.IsCleared() {
		return
	}
	if event.What == EvKeyboard && event.Key != nil {
		old := ms.selected
		switch event.Key.Key {
		case tcell.KeyUp:
			ms.selected = (ms.selected - 1 + 4) % 4
		case tcell.KeyDown:
			ms.selected = (ms.selected + 1) % 4
		default:
			return
		}
		if ms.selected != old {
			owner := ms.Owner()
			if owner != nil {
				ev := &Event{What: EvBroadcast, Command: CmColorForegroundChanged, Info: monoOptions[ms.selected].attr}
				owner.HandleEvent(ev)
			}
		}
		event.Clear()
	}
}
