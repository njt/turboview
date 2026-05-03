package tv

import (
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var _ Widget = (*RadioButton)(nil)
var _ Container = (*RadioButtons)(nil)

// ---------------------------------------------------------------------------
// RadioButton
// ---------------------------------------------------------------------------

type RadioButton struct {
	BaseView
	label    string
	shortcut rune
	selected bool
	disabled bool
}

func NewRadioButton(bounds Rect, label string) *RadioButton {
	rb := &RadioButton{label: label}
	rb.SetBounds(bounds)
	rb.SetState(SfVisible, true)
	rb.SetOptions(OfSelectable|OfFirstClick, true)

	segments := ParseTildeLabel(label)
	for _, seg := range segments {
		if seg.Shortcut && len(seg.Text) > 0 {
			rb.shortcut, _ = utf8.DecodeRuneInString(seg.Text)
			break
		}
	}

	rb.SetSelf(rb)
	return rb
}

func (rb *RadioButton) Selected() bool     { return rb.selected }
func (rb *RadioButton) SetSelected(v bool) { rb.selected = v }
func (rb *RadioButton) Label() string      { return rb.label }
func (rb *RadioButton) Shortcut() rune     { return rb.shortcut }

func (rb *RadioButton) Draw(buf *DrawBuffer) {
	cs := rb.ColorScheme()
	normalStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	if cs != nil {
		normalStyle = cs.RadioButtonNormal
		selectedStyle = cs.RadioButtonSelected
		shortcutStyle = cs.LabelShortcut
	}

	if rb.disabled {
		disabledStyle := tcell.StyleDefault
		if cs != nil {
			disabledStyle = cs.ClusterDisabled
		}
		// Disabled: never show ►, use disabledStyle for all rendering.
		buf.WriteChar(0, 0, ' ', disabledStyle)
		buf.WriteChar(1, 0, '(', disabledStyle)
		if rb.selected {
			buf.WriteChar(2, 0, '*', disabledStyle)
		} else {
			buf.WriteChar(2, 0, ' ', disabledStyle)
		}
		buf.WriteChar(3, 0, ')', disabledStyle)
		buf.WriteChar(4, 0, ' ', disabledStyle)
		x := 5
		segments := ParseTildeLabel(rb.label)
		for _, seg := range segments {
			buf.WriteStr(x, 0, seg.Text, disabledStyle)
			x += utf8.RuneCountInString(seg.Text)
		}
		return
	}

	clusterFocused := rb.Owner() == nil || rb.Owner().HasState(SfSelected)
	focused := rb.HasState(SfSelected) && clusterFocused
	itemStyle := normalStyle
	if focused {
		itemStyle = selectedStyle
	}

	if focused {
		buf.WriteChar(0, 0, '►', itemStyle)
	} else {
		buf.WriteChar(0, 0, ' ', normalStyle)
	}

	// Paren + mark always at columns 1-3
	buf.WriteChar(1, 0, '(', itemStyle)
	if rb.selected {
		buf.WriteChar(2, 0, '*', itemStyle)
	} else {
		buf.WriteChar(2, 0, ' ', itemStyle)
	}
	buf.WriteChar(3, 0, ')', itemStyle)

	// Space before label at column 4
	buf.WriteChar(4, 0, ' ', itemStyle)

	// Label text starts at column 5
	x := 5
	segments := ParseTildeLabel(rb.label)
	for _, seg := range segments {
		style := itemStyle
		if seg.Shortcut {
			style = shortcutStyle
		}
		buf.WriteStr(x, 0, seg.Text, style)
		x += utf8.RuneCountInString(seg.Text)
	}
}

func (rb *RadioButton) HandleEvent(event *Event) {
	if rb.disabled {
		return
	}

	if event.What == EvMouse && event.Mouse != nil {
		rb.BaseView.HandleEvent(event)
		if event.IsCleared() {
			return
		}
		if event.Mouse.Button == tcell.Button1 {
			rb.selectInCluster()
			event.Clear()
		}
		return
	}

	if event.What == EvKeyboard && event.Key != nil {
		if event.Key.Key == tcell.KeyRune && event.Key.Rune == ' ' {
			rb.selectInCluster()
			event.Clear()
		}
	}
}

func (rb *RadioButton) selectInCluster() {
	// Notify the cluster (owner) to handle exclusive selection
	if cluster, ok := rb.Owner().(*RadioButtons); ok {
		for _, item := range cluster.items {
			item.selected = (item == rb)
		}
	} else {
		rb.selected = true
	}
}

// ---------------------------------------------------------------------------
// RadioButtons
// ---------------------------------------------------------------------------

type RadioButtons struct {
	BaseView
	group      *Group
	items      []*RadioButton
	enableMask uint32
}

func NewRadioButtons(bounds Rect, labels []string) *RadioButtons {
	rbs := &RadioButtons{}
	rbs.enableMask = ^uint32(0)
	rbs.SetBounds(bounds)
	rbs.SetState(SfVisible, true)
	rbs.SetOptions(OfSelectable|OfFirstClick, true)
	rbs.SetOptions(OfPreProcess|OfPostProcess, true)

	rbs.group = NewGroup(bounds)
	rbs.group.SetFacade(rbs)

	for _, label := range labels {
		rb := NewRadioButton(NewRect(0, 0, bounds.Width(), 1), label)
		rbs.items = append(rbs.items, rb)
		rbs.group.Insert(rb)
	}

	// Select the first item by default
	if len(rbs.items) > 0 {
		rbs.items[0].selected = true
	}

	rbs.relayoutItems()

	rbs.SetSelf(rbs)
	return rbs
}

func (rbs *RadioButtons) relayoutItems() {
	h := rbs.Bounds().Height()
	if h <= 0 {
		h = len(rbs.items)
	}
	if h <= 0 {
		return
	}
	numCols := (len(rbs.items) + h - 1) / h
	colWidths := make([]int, numCols)
	for i, item := range rbs.items {
		col := i / h
		w := labelDisplayWidth(item.label) + 6
		if w > colWidths[col] {
			colWidths[col] = w
		}
	}
	lastCol := numCols - 1
	for i, item := range rbs.items {
		col := i / h
		row := i % h
		colX := 0
		for c := 0; c < col; c++ {
			colX += colWidths[c]
		}
		item.SetBounds(NewRect(colX, row, colWidths[col], 1))
		if col == lastCol {
			item.SetGrowMode(GfGrowHiX)
		} else {
			item.SetGrowMode(0)
		}
	}
}

func (rbs *RadioButtons) SetBounds(r Rect) {
	rbs.BaseView.SetBounds(r)
	if rbs.group != nil {
		rbs.group.SetBounds(NewRect(0, 0, r.Width(), r.Height()))
		rbs.relayoutItems()
	}
}

func (rbs *RadioButtons) SetEnabled(index int, enabled bool) {
	if index < 0 || index >= len(rbs.items) {
		return
	}
	if enabled {
		rbs.enableMask |= 1 << uint(index)
	} else {
		rbs.enableMask &^= 1 << uint(index)
	}
	// Keep item.disabled in sync so HandleEvent guards work without requiring Draw first.
	rbs.items[index].disabled = !enabled
	anyEnabled := rbs.enableMask&((1<<uint(len(rbs.items)))-1) != 0
	rbs.SetOptions(OfSelectable, anyEnabled)
}

func (rbs *RadioButtons) IsEnabled(index int) bool {
	if index < 0 || index >= len(rbs.items) {
		return false
	}
	return rbs.enableMask&(1<<uint(index)) != 0
}

func (rbs *RadioButtons) Value() int {
	for i, item := range rbs.items {
		if item.Selected() {
			return i
		}
	}
	return -1
}

func (rbs *RadioButtons) SetValue(index int) {
	for i, item := range rbs.items {
		item.selected = (i == index)
	}
}

func (rbs *RadioButtons) Item(index int) *RadioButton {
	return rbs.items[index]
}

// Container interface — delegate to internal group

func (rbs *RadioButtons) Insert(v View)               { rbs.group.Insert(v) }
func (rbs *RadioButtons) Remove(v View)               { rbs.group.Remove(v) }
func (rbs *RadioButtons) Children() []View            { return rbs.group.Children() }
func (rbs *RadioButtons) FocusedChild() View          { return rbs.group.FocusedChild() }
func (rbs *RadioButtons) SetFocusedChild(v View)      { rbs.group.SetFocusedChild(v) }
func (rbs *RadioButtons) ExecView(v View) CommandCode { return rbs.group.ExecView(v) }
func (rbs *RadioButtons) BringToFront(v View)         { rbs.group.BringToFront(v) }

func (rbs *RadioButtons) Draw(buf *DrawBuffer) {
	for i, item := range rbs.items {
		item.disabled = !rbs.IsEnabled(i)
		childBounds := item.Bounds()
		sub := buf.SubBuffer(childBounds)
		item.Draw(sub)
	}
}

func (rbs *RadioButtons) HandleEvent(event *Event) {
	// Route mouse events by position to the correct child.
	if event.What == EvMouse && event.Mouse != nil {
		rbs.BaseView.HandleEvent(event)
		if event.IsCleared() {
			return
		}
		mx, my := event.Mouse.X, event.Mouse.Y
		for i, item := range rbs.items {
			if item.Bounds().Contains(NewPoint(mx, my)) {
				if !rbs.IsEnabled(i) {
					event.Clear()
					return
				}
				// Adjust coordinates to child-local space and forward.
				origX, origY := event.Mouse.X, event.Mouse.Y
				event.Mouse.X -= item.Bounds().A.X
				event.Mouse.Y -= item.Bounds().A.Y
				item.HandleEvent(event)
				event.Mouse.X, event.Mouse.Y = origX, origY
				return
			}
		}
		return
	}

	// Plain-letter shortcut matching (before SfSelected guard).
	if event.What == EvKeyboard && event.Key != nil &&
		event.Key.Key == tcell.KeyRune && event.Key.Modifiers == 0 {
		r := unicode.ToLower(event.Key.Rune)
		for i, item := range rbs.items {
			if item.Shortcut() != 0 && unicode.ToLower(item.Shortcut()) == r {
				if !rbs.IsEnabled(i) {
					continue
				}
				if owner, ok := rbs.Owner().(Container); ok && !rbs.HasState(SfSelected) {
					owner.SetFocusedChild(rbs)
				}
				rbs.group.SetFocusedChild(item)
				rbs.SetValue(i)
				event.Clear()
				return
			}
		}
	}

	// Handle Alt+shortcut to focus and select matching radio button
	if event.What == EvKeyboard && event.Key != nil &&
		event.Key.Key == tcell.KeyRune &&
		event.Key.Modifiers&tcell.ModAlt != 0 {

		r := unicode.ToLower(event.Key.Rune)
		for i, item := range rbs.items {
			if item.Shortcut() != 0 && unicode.ToLower(item.Shortcut()) == r {
				if !rbs.IsEnabled(i) {
					continue
				}
				rbs.group.SetFocusedChild(item)
				rbs.SetValue(i)
				event.Clear()
				return
			}
		}
	}

	// Handle Up/Down/Right/Left arrow for selection (not just focus).
	// Only consume these keys when RadioButtons itself has focus (SfSelected=true);
	// otherwise OfPostProcess siblings (e.g. History) would never see Down/Up
	// because RadioButtons runs in Phase1 (OfPreProcess) before them.
	if event.What == EvKeyboard && event.Key != nil && rbs.HasState(SfSelected) {
		switch event.Key.Key {
		case tcell.KeyDown:
			rbs.moveSelection(1)
			event.Clear()
			return
		case tcell.KeyUp:
			rbs.moveSelection(-1)
			event.Clear()
			return
		case tcell.KeyRight:
			h := rbs.Bounds().Height()
			if h <= 0 {
				h = len(rbs.items)
			}
			rbs.moveSelection(h)
			event.Clear()
			return
		case tcell.KeyLeft:
			h := rbs.Bounds().Height()
			if h <= 0 {
				h = len(rbs.items)
			}
			rbs.moveSelection(-h)
			event.Clear()
			return
		}
	}

	// Delegate to group only when focused — prevents PostProcess from
	// forwarding Space/etc. to the internal RadioButton items.
	if rbs.HasState(SfSelected) {
		rbs.group.HandleEvent(event)
	}
}

func (rbs *RadioButtons) moveSelection(delta int) {
	current := rbs.Value()
	if current < 0 {
		current = 0
	}
	next := current + delta
	for next >= 0 && next < len(rbs.items) && !rbs.IsEnabled(next) {
		next += delta
	}
	if next < 0 || next >= len(rbs.items) {
		return
	}
	rbs.SetValue(next)
	rbs.group.SetFocusedChild(rbs.items[next])
}
