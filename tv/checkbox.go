package tv

import (
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var _ Widget = (*CheckBox)(nil)
var _ Container = (*CheckBoxes)(nil)

// ---------------------------------------------------------------------------
// CheckBox
// ---------------------------------------------------------------------------

type CheckBox struct {
	BaseView
	label    string
	shortcut rune
	checked  bool
}

func NewCheckBox(bounds Rect, label string) *CheckBox {
	cb := &CheckBox{label: label}
	cb.SetBounds(bounds)
	cb.SetState(SfVisible, true)
	cb.SetOptions(OfSelectable, true)

	// Extract shortcut from tilde notation
	segments := ParseTildeLabel(label)
	for _, seg := range segments {
		if seg.Shortcut && len(seg.Text) > 0 {
			cb.shortcut, _ = utf8.DecodeRuneInString(seg.Text)
			break
		}
	}

	cb.SetSelf(cb)
	return cb
}

func (cb *CheckBox) Checked() bool          { return cb.checked }
func (cb *CheckBox) SetChecked(v bool)      { cb.checked = v }
func (cb *CheckBox) Label() string          { return cb.label }
func (cb *CheckBox) Shortcut() rune         { return cb.shortcut }

func (cb *CheckBox) Draw(buf *DrawBuffer) {
	cs := cb.ColorScheme()
	normalStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	if cs != nil {
		normalStyle = cs.CheckBoxNormal
		selectedStyle = cs.CheckBoxSelected
		shortcutStyle = cs.LabelShortcut
	}

	clusterFocused := cb.Owner() == nil || cb.Owner().HasState(SfSelected)
	focused := cb.HasState(SfSelected) && clusterFocused
	itemStyle := normalStyle
	if focused {
		itemStyle = selectedStyle
	}

	if focused {
		buf.WriteChar(0, 0, '►', itemStyle)
	} else {
		buf.WriteChar(0, 0, ' ', normalStyle)
	}

	// Bracket + mark always at columns 1-3
	buf.WriteChar(1, 0, '[', itemStyle)
	if cb.checked {
		buf.WriteChar(2, 0, 'X', itemStyle)
	} else {
		buf.WriteChar(2, 0, ' ', itemStyle)
	}
	buf.WriteChar(3, 0, ']', itemStyle)

	// Space before label at column 4
	buf.WriteChar(4, 0, ' ', itemStyle)

	// Label text starts at column 5
	x := 5
	segments := ParseTildeLabel(cb.label)
	for _, seg := range segments {
		style := itemStyle
		if seg.Shortcut {
			style = shortcutStyle
		}
		buf.WriteStr(x, 0, seg.Text, style)
		x += utf8.RuneCountInString(seg.Text)
	}
}

func (cb *CheckBox) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		if event.Mouse.Button == tcell.Button1 {
			cb.checked = !cb.checked
			event.Clear()
		}
		return
	}

	if event.What == EvKeyboard && event.Key != nil {
		if event.Key.Key == tcell.KeyRune && event.Key.Rune == ' ' {
			cb.checked = !cb.checked
			event.Clear()
		}
	}
}

// ---------------------------------------------------------------------------
// CheckBoxes
// ---------------------------------------------------------------------------

type CheckBoxes struct {
	BaseView
	group *Group
	items []*CheckBox
}

func NewCheckBoxes(bounds Rect, labels []string) *CheckBoxes {
	cbs := &CheckBoxes{}
	cbs.SetBounds(bounds)
	cbs.SetState(SfVisible, true)
	cbs.SetOptions(OfSelectable, true)
	cbs.SetOptions(OfPreProcess, true)

	cbs.group = NewGroup(bounds)
	cbs.group.SetFacade(cbs)

	for i, label := range labels {
		cb := NewCheckBox(NewRect(0, i, bounds.Width(), 1), label)
		cbs.items = append(cbs.items, cb)
		cbs.group.Insert(cb)
	}

	cbs.SetSelf(cbs)
	return cbs
}

func (cbs *CheckBoxes) Values() uint32 {
	var mask uint32
	for i, item := range cbs.items {
		if item.Checked() {
			mask |= 1 << uint(i)
		}
	}
	return mask
}

func (cbs *CheckBoxes) SetValues(mask uint32) {
	for i, item := range cbs.items {
		item.SetChecked(mask&(1<<uint(i)) != 0)
	}
}

func (cbs *CheckBoxes) Item(index int) *CheckBox {
	return cbs.items[index]
}

// Container interface — delegate to internal group

func (cbs *CheckBoxes) Insert(v View)               { cbs.group.Insert(v) }
func (cbs *CheckBoxes) Remove(v View)               { cbs.group.Remove(v) }
func (cbs *CheckBoxes) Children() []View            { return cbs.group.Children() }
func (cbs *CheckBoxes) FocusedChild() View          { return cbs.group.FocusedChild() }
func (cbs *CheckBoxes) SetFocusedChild(v View)      { cbs.group.SetFocusedChild(v) }
func (cbs *CheckBoxes) ExecView(v View) CommandCode { return cbs.group.ExecView(v) }
func (cbs *CheckBoxes) BringToFront(v View)         { cbs.group.BringToFront(v) }

func (cbs *CheckBoxes) Draw(buf *DrawBuffer) {
	for _, item := range cbs.items {
		childBounds := item.Bounds()
		sub := buf.SubBuffer(childBounds)
		item.Draw(sub)
	}
}

func (cbs *CheckBoxes) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		mx, my := event.Mouse.X, event.Mouse.Y
		for _, item := range cbs.items {
			if item.Bounds().Contains(NewPoint(mx, my)) {
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

	// Handle Alt+shortcut to focus and toggle matching checkbox
	if event.What == EvKeyboard && event.Key != nil &&
		event.Key.Key == tcell.KeyRune &&
		event.Key.Modifiers&tcell.ModAlt != 0 {

		r := unicode.ToLower(event.Key.Rune)
		for _, item := range cbs.items {
			if item.Shortcut() != 0 && unicode.ToLower(item.Shortcut()) == r {
				cbs.group.SetFocusedChild(item)
				item.SetChecked(!item.Checked())
				event.Clear()
				return
			}
		}
	}

	// Handle Up/Down arrow for focus navigation (does NOT toggle).
	// Only consume these keys when CheckBoxes itself has focus (SfSelected=true);
	// otherwise OfPostProcess siblings (e.g. History) would never see Down/Up
	// because CheckBoxes runs in Phase1 (OfPreProcess) before them.
	if event.What == EvKeyboard && event.Key != nil && cbs.HasState(SfSelected) {
		switch event.Key.Key {
		case tcell.KeyDown:
			cbs.moveNavigation(1)
			event.Clear()
			return
		case tcell.KeyUp:
			cbs.moveNavigation(-1)
			event.Clear()
			return
		}
	}

	// Delegate to group for Tab/Shift+Tab and other events
	cbs.group.HandleEvent(event)
}

func (cbs *CheckBoxes) moveNavigation(delta int) {
	current := -1
	for i, item := range cbs.items {
		if item.HasState(SfSelected) {
			current = i
			break
		}
	}
	if current < 0 {
		current = 0
	}
	next := current + delta
	if next < 0 || next >= len(cbs.items) {
		return
	}
	cbs.group.SetFocusedChild(cbs.items[next])
}
