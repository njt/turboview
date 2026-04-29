package tv

import (
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type StatusItem struct {
	Label      string
	KeyBinding KeyBinding
	Command    CommandCode
	HelpCtx    HelpContext
}

func NewStatusItem(label string, kb KeyBinding, cmd CommandCode) *StatusItem {
	return &StatusItem{
		Label:      label,
		KeyBinding: kb,
		Command:    cmd,
	}
}

type StatusLine struct {
	BaseView
	items      []*StatusItem
	activeCtx  HelpContext
	pressedIdx int // -1 means no item pressed
}

func (si *StatusItem) ForHelpCtx(hc HelpContext) *StatusItem {
	si.HelpCtx = hc
	return si
}

func (sl *StatusLine) SetActiveContext(hc HelpContext) {
	sl.activeCtx = hc
}

func NewStatusLine(items ...*StatusItem) *StatusLine {
	sl := &StatusLine{
		items:      items,
		pressedIdx: -1,
	}
	sl.SetState(SfVisible, true)
	return sl
}

// itemVisible reports whether item passes the current HelpCtx filter.
func (sl *StatusLine) itemVisible(item *StatusItem) bool {
	return item.HelpCtx == HcNoContext || item.HelpCtx == sl.activeCtx
}

// itemRanges computes the [startX, endX) ranges for each visible item,
// returning a slice parallel to sl.items where non-visible items have {-1,-1}.
type itemRange struct {
	start, end int
	idx        int // index into sl.items
}

func (sl *StatusLine) visibleItemRanges() []itemRange {
	var ranges []itemRange
	x := 1
	first := true
	for i, item := range sl.items {
		if !sl.itemVisible(item) {
			continue
		}
		if !first {
			x += 2
		}
		first = false
		w := 0
		for _, seg := range ParseTildeLabel(item.Label) {
			w += utf8.RuneCountInString(seg.Text)
		}
		ranges = append(ranges, itemRange{start: x, end: x + w, idx: i})
		x += w
	}
	return ranges
}

// itemAtX returns the index into sl.items for the visible item at x,
// or -1 if no item covers that x.
func (sl *StatusLine) itemAtX(mx int) int {
	for _, r := range sl.visibleItemRanges() {
		if mx >= r.start && mx < r.end {
			return r.idx
		}
	}
	return -1
}

func (sl *StatusLine) Draw(buf *DrawBuffer) {
	w := sl.Bounds().Width()
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault
	if cs := sl.ColorScheme(); cs != nil {
		normalStyle = cs.StatusNormal
		shortcutStyle = cs.StatusShortcut
		selectedStyle = cs.StatusSelected
	}

	buf.Fill(NewRect(0, 0, w, 1), ' ', normalStyle)

	x := 1
	first := true
	for i, item := range sl.items {
		if !sl.itemVisible(item) {
			continue
		}
		if !first {
			x += 2
		}
		first = false
		isPressed := (sl.pressedIdx == i)
		segments := ParseTildeLabel(item.Label)
		for _, seg := range segments {
			style := normalStyle
			if isPressed {
				style = selectedStyle
			} else if seg.Shortcut {
				style = shortcutStyle
			}
			buf.WriteStr(x, 0, seg.Text, style)
			x += utf8.RuneCountInString(seg.Text)
		}
	}
}

func (sl *StatusLine) HandleEvent(event *Event) {
	// Handle mouse events.
	if event.What == EvMouse && event.Mouse != nil {
		mx := event.Mouse.X
		held := event.Mouse.Button&tcell.Button1 != 0

		if held {
			// Press or drag: update pressedIdx to the item under cursor.
			sl.pressedIdx = sl.itemAtX(mx)
			return
		}

		// Button released.
		if sl.pressedIdx >= 0 {
			pressedIdx := sl.pressedIdx
			sl.pressedIdx = -1
			// Only fire if release is on the same item as the press.
			releaseIdx := sl.itemAtX(mx)
			if releaseIdx >= 0 && releaseIdx == pressedIdx {
				event.What = EvCommand
				event.Command = sl.items[releaseIdx].Command
				event.Mouse = nil
			}
		}
		return
	}

	if event.What != EvKeyboard || event.Key == nil {
		return
	}
	for _, item := range sl.items {
		if !sl.itemVisible(item) {
			continue
		}
		if item.KeyBinding.Matches(event.Key) {
			event.What = EvCommand
			event.Command = item.Command
			event.Key = nil
			return
		}
	}
}
