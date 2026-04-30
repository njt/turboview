package tv

import "github.com/gdamore/tcell/v2"

var _ Widget = (*ScrollBar)(nil)

type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

type ScrollBar struct {
	BaseView
	orientation Orientation
	min         int
	max         int
	value       int
	pageSize    int
	arStep      int
	OnChange    func(int)
}

func NewScrollBar(bounds Rect, orientation Orientation) *ScrollBar {
	sb := &ScrollBar{orientation: orientation, arStep: 1}
	sb.SetBounds(bounds)
	sb.SetState(SfVisible, true)
	sb.SetOptions(OfSelectable, true)
	sb.SetSelf(sb)
	return sb
}

func (sb *ScrollBar) ArStep() int      { return sb.arStep }
func (sb *ScrollBar) SetArStep(n int)  { sb.arStep = n }

func (sb *ScrollBar) Min() int      { return sb.min }
func (sb *ScrollBar) Max() int      { return sb.max }
func (sb *ScrollBar) Value() int    { return sb.value }
func (sb *ScrollBar) PageSize() int { return sb.pageSize }

func (sb *ScrollBar) SetRange(min, max int) {
	sb.min = min
	sb.max = max
	sb.clampValue()
}

func (sb *ScrollBar) SetValue(v int) {
	sb.value = v
	sb.clampValue()
}

func (sb *ScrollBar) SetPageSize(n int) {
	sb.pageSize = n
	sb.clampValue()
}

func (sb *ScrollBar) clampValue() {
	maxVal := sb.max - sb.pageSize
	if maxVal < sb.min {
		maxVal = sb.min
	}
	if sb.value < sb.min {
		sb.value = sb.min
	}
	if sb.value > maxVal {
		sb.value = maxVal
	}
}

func (sb *ScrollBar) trackLen() int {
	if sb.orientation == Vertical {
		return sb.Bounds().Height() - 2
	}
	return sb.Bounds().Width() - 2
}

func (sb *ScrollBar) thumbInfo() (pos, length int) {
	tl := sb.trackLen()
	if tl < 1 {
		return 0, 0
	}
	rng := sb.max - sb.min
	if rng <= 0 || sb.pageSize >= rng {
		return 0, tl
	}
	length = tl * sb.pageSize / rng
	if length < 1 {
		length = 1
	}
	scrollRange := rng - sb.pageSize
	if scrollRange <= 0 {
		return 0, length
	}
	pos = (sb.value - sb.min) * (tl - length) / scrollRange
	if pos < 0 {
		pos = 0
	}
	if pos > tl-length {
		pos = tl - length
	}
	return pos, length
}

func (sb *ScrollBar) Draw(buf *DrawBuffer) {
	cs := sb.ColorScheme()
	barStyle := tcell.StyleDefault
	thumbStyle := tcell.StyleDefault
	if cs != nil {
		barStyle = cs.ScrollBar
		thumbStyle = cs.ScrollThumb
	}

	if sb.orientation == Vertical {
		sb.drawVertical(buf, barStyle, thumbStyle)
	} else {
		sb.drawHorizontal(buf, barStyle, thumbStyle)
	}
}

func (sb *ScrollBar) drawVertical(buf *DrawBuffer, barStyle, thumbStyle tcell.Style) {
	h := sb.Bounds().Height()
	if h < 2 {
		return
	}
	buf.WriteChar(0, 0, '▲', barStyle)
	buf.WriteChar(0, h-1, '▼', barStyle)

	tl := h - 2
	for i := 0; i < tl; i++ {
		buf.WriteChar(0, i+1, '░', barStyle)
	}

	thumbPos, thumbLen := sb.thumbInfo()
	for i := 0; i < thumbLen && i+thumbPos < tl; i++ {
		buf.WriteChar(0, 1+thumbPos+i, '█', thumbStyle)
	}
}

func (sb *ScrollBar) drawHorizontal(buf *DrawBuffer, barStyle, thumbStyle tcell.Style) {
	w := sb.Bounds().Width()
	if w < 2 {
		return
	}
	buf.WriteChar(0, 0, '◄', barStyle)
	buf.WriteChar(w-1, 0, '►', barStyle)

	tl := w - 2
	for i := 0; i < tl; i++ {
		buf.WriteChar(i+1, 0, '░', barStyle)
	}

	thumbPos, thumbLen := sb.thumbInfo()
	for i := 0; i < thumbLen && i+thumbPos < tl; i++ {
		buf.WriteChar(1+thumbPos+i, 0, '█', thumbStyle)
	}
}

func (sb *ScrollBar) HandleEvent(event *Event) {
	if event.What == EvKeyboard && event.Key != nil && sb.HasState(SfSelected) {
		sb.handleKeyboard(event)
		return
	}

	if event.What != EvMouse || event.Mouse == nil {
		return
	}

	// Mouse wheel
	if event.Mouse.Button == tcell.WheelUp {
		sb.step(-3 * sb.arStep)
		event.Clear()
		return
	}
	if event.Mouse.Button == tcell.WheelDown {
		sb.step(3 * sb.arStep)
		event.Clear()
		return
	}

	if event.Mouse.Button&tcell.Button1 == 0 {
		return
	}

	if sb.orientation == Vertical {
		sb.handleVerticalClick(event)
	} else {
		sb.handleHorizontalClick(event)
	}
}

func (sb *ScrollBar) handleVerticalClick(event *Event) {
	my := event.Mouse.Y
	h := sb.Bounds().Height()

	if my == 0 {
		sb.step(-1)
		event.Clear()
		return
	}
	if my == h-1 {
		sb.step(1)
		event.Clear()
		return
	}

	trackY := my - 1
	thumbPos, _ := sb.thumbInfo()

	if trackY < thumbPos {
		sb.page(-1)
	} else {
		sb.page(1)
	}
	event.Clear()
}

func (sb *ScrollBar) handleHorizontalClick(event *Event) {
	mx := event.Mouse.X
	w := sb.Bounds().Width()

	if mx == 0 {
		sb.step(-1)
		event.Clear()
		return
	}
	if mx == w-1 {
		sb.step(1)
		event.Clear()
		return
	}

	trackX := mx - 1
	thumbPos, _ := sb.thumbInfo()

	if trackX < thumbPos {
		sb.page(-1)
	} else {
		sb.page(1)
	}
	event.Clear()
}

func (sb *ScrollBar) step(dir int) {
	old := sb.value
	sb.value += dir
	sb.clampValue()
	if sb.value != old && sb.OnChange != nil {
		sb.OnChange(sb.value)
	}
}

func (sb *ScrollBar) page(dir int) {
	old := sb.value
	sb.value += dir * sb.pageSize
	sb.clampValue()
	if sb.value != old && sb.OnChange != nil {
		sb.OnChange(sb.value)
	}
}

func (sb *ScrollBar) goToMin() {
	old := sb.value
	sb.value = sb.min
	sb.clampValue()
	if sb.value != old && sb.OnChange != nil {
		sb.OnChange(sb.value)
	}
}

func (sb *ScrollBar) goToMax() {
	old := sb.value
	sb.value = sb.max - sb.pageSize
	sb.clampValue()
	if sb.value != old && sb.OnChange != nil {
		sb.OnChange(sb.value)
	}
}

func (sb *ScrollBar) handleKeyboard(event *Event) {
	if sb.orientation == Vertical {
		sb.handleVerticalKeyboard(event)
	} else {
		sb.handleHorizontalKeyboard(event)
	}
}

func (sb *ScrollBar) handleVerticalKeyboard(event *Event) {
	ke := event.Key
	switch {
	case ke.Key == tcell.KeyUp:
		sb.step(-sb.arStep)
		event.Clear()
	case ke.Key == tcell.KeyDown:
		sb.step(sb.arStep)
		event.Clear()
	case ke.Key == tcell.KeyPgUp && ke.Modifiers&tcell.ModCtrl != 0:
		sb.goToMin()
		event.Clear()
	case ke.Key == tcell.KeyPgDn && ke.Modifiers&tcell.ModCtrl != 0:
		sb.goToMax()
		event.Clear()
	case ke.Key == tcell.KeyPgUp:
		sb.page(-1)
		event.Clear()
	case ke.Key == tcell.KeyPgDn:
		sb.page(1)
		event.Clear()
	}
}

func (sb *ScrollBar) handleHorizontalKeyboard(event *Event) {
	ke := event.Key
	switch {
	case ke.Key == tcell.KeyLeft && ke.Modifiers&tcell.ModCtrl != 0:
		sb.page(-1)
		event.Clear()
	case ke.Key == tcell.KeyRight && ke.Modifiers&tcell.ModCtrl != 0:
		sb.page(1)
		event.Clear()
	case ke.Key == tcell.KeyLeft:
		sb.step(-sb.arStep)
		event.Clear()
	case ke.Key == tcell.KeyRight:
		sb.step(sb.arStep)
		event.Clear()
	case ke.Key == tcell.KeyHome:
		sb.goToMin()
		event.Clear()
	case ke.Key == tcell.KeyEnd:
		sb.goToMax()
		event.Clear()
	}
}
