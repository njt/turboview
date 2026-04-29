package tv

import (
	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

type ViewState uint16

const (
	SfVisible  ViewState = 1 << iota
	SfFocused
	SfSelected
	SfModal
	SfDisabled
	SfExposed
	SfDragging
)

type ViewOptions uint16

const (
	OfSelectable  ViewOptions = 1 << iota
	OfTopSelect
	OfFirstClick
	OfPreProcess
	OfPostProcess
	OfCentered
)

type HelpContext uint16

const HcNoContext HelpContext = 0

type View interface {
	Draw(buf *DrawBuffer)
	HandleEvent(event *Event)
	Bounds() Rect
	SetBounds(Rect)
	GrowMode() GrowFlag
	SetGrowMode(GrowFlag)
	Owner() Container
	SetOwner(Container)
	State() ViewState
	SetState(ViewState, bool)
	EventMask() EventType
	SetEventMask(EventType)
	Options() ViewOptions
	SetOptions(ViewOptions, bool)
	HasState(ViewState) bool
	HasOption(ViewOptions) bool
	ColorScheme() *theme.ColorScheme
}

type Container interface {
	View
	Insert(View)
	Remove(View)
	Children() []View
	FocusedChild() View
	SetFocusedChild(View)
	ExecView(View) CommandCode
}

type Widget interface {
	View
}

type EventSource interface {
	PollEvent() *Event
}

type BaseView struct {
	origin    Point
	size      Point
	growMode  GrowFlag
	state     ViewState
	eventMask EventType
	options   ViewOptions
	owner     Container
	scheme    *theme.ColorScheme
	helpCtx   HelpContext
	self      View
}

func (b *BaseView) SetSelf(v View) { b.self = v }

func (b *BaseView) Draw(buf *DrawBuffer) {}
func (b *BaseView) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		realButtons := event.Mouse.Button & (tcell.Button1 | tcell.Button2 | tcell.Button3)
		if realButtons != 0 && !b.HasState(SfSelected) && !b.HasState(SfDisabled) && b.HasOption(OfSelectable) && b.self != nil {
			if b.owner != nil {
				b.owner.SetFocusedChild(b.self)
			}
			if !b.HasOption(OfFirstClick) {
				event.Clear()
			}
		}
	}
}

func (b *BaseView) Bounds() Rect {
	return Rect{
		A: b.origin,
		B: Point{X: b.origin.X + b.size.X, Y: b.origin.Y + b.size.Y},
	}
}

func (b *BaseView) SetBounds(r Rect) {
	b.origin = r.A
	b.size = Point{X: r.Width(), Y: r.Height()}
}

func (b *BaseView) GrowMode() GrowFlag      { return b.growMode }
func (b *BaseView) SetGrowMode(gm GrowFlag) { b.growMode = gm }

func (b *BaseView) Owner() Container      { return b.owner }
func (b *BaseView) SetOwner(c Container)  { b.owner = c }

func (b *BaseView) State() ViewState { return b.state }
func (b *BaseView) SetState(flag ViewState, on bool) {
	if on {
		b.state |= flag
	} else {
		b.state &^= flag
	}
}

func (b *BaseView) EventMask() EventType       { return b.eventMask }
func (b *BaseView) SetEventMask(mask EventType) { b.eventMask = mask }

func (b *BaseView) Options() ViewOptions { return b.options }
func (b *BaseView) SetOptions(flag ViewOptions, on bool) {
	if on {
		b.options |= flag
	} else {
		b.options &^= flag
	}
}

func (b *BaseView) HasState(flag ViewState) bool    { return b.state&flag != 0 }
func (b *BaseView) HasOption(flag ViewOptions) bool  { return b.options&flag != 0 }

func (b *BaseView) HelpCtx() HelpContext        { return b.helpCtx }
func (b *BaseView) SetHelpCtx(hc HelpContext)    { b.helpCtx = hc }

func (b *BaseView) ColorScheme() *theme.ColorScheme {
	if b.scheme != nil {
		return b.scheme
	}
	if b.owner != nil {
		return b.owner.ColorScheme()
	}
	return nil
}
