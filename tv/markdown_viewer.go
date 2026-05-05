package tv

import (
	"github.com/gdamore/tcell/v2"
)

var _ Widget = (*MarkdownViewer)(nil)

// MarkdownViewer renders parsed markdown blocks with scroll support.
type MarkdownViewer struct {
	BaseView
	blocks     []mdBlock
	source     string
	deltaX     int
	deltaY     int
	wrapText   bool
	vScrollBar *ScrollBar
	hScrollBar *ScrollBar
}

// NewMarkdownViewer creates a new MarkdownViewer with the given bounds.
func NewMarkdownViewer(bounds Rect) *MarkdownViewer {
	mv := &MarkdownViewer{wrapText: true}
	mv.SetBounds(bounds)
	mv.SetState(SfVisible, true)
	mv.SetOptions(OfSelectable|OfFirstClick, true)
	mv.SetSelf(mv)
	return mv
}

// Markdown returns the original source string passed to SetMarkdown.
func (mv *MarkdownViewer) Markdown() string { return mv.source }

// WrapText returns whether word wrapping is enabled.
func (mv *MarkdownViewer) WrapText() bool { return mv.wrapText }

// DeltaX returns the current horizontal scroll offset.
func (mv *MarkdownViewer) DeltaX() int { return mv.deltaX }

// DeltaY returns the current vertical scroll offset.
func (mv *MarkdownViewer) DeltaY() int { return mv.deltaY }

// SetMarkdown parses and stores the markdown content, resetting scroll position.
func (mv *MarkdownViewer) SetMarkdown(text string) {
	mv.source = text
	mv.blocks = parseMarkdown(text)
	if mv.blocks == nil {
		mv.blocks = []mdBlock{}
	}
	mv.deltaX = 0
	mv.deltaY = 0
	mv.syncScrollBars()
}

// SetWrapText enables or disables word wrapping and resets horizontal scroll.
func (mv *MarkdownViewer) SetWrapText(wrap bool) {
	mv.wrapText = wrap
	mv.deltaX = 0
	mv.syncScrollBars()
}

// SetVScrollBar binds a vertical scrollbar to this viewer.
func (mv *MarkdownViewer) SetVScrollBar(sb *ScrollBar) {
	if mv.vScrollBar != nil {
		mv.vScrollBar.OnChange = nil
	}
	mv.vScrollBar = sb
	if sb != nil {
		sb.OnChange = func(val int) { mv.deltaY = val }
		mv.syncScrollBars()
	}
}

// SetHScrollBar binds a horizontal scrollbar to this viewer.
func (mv *MarkdownViewer) SetHScrollBar(sb *ScrollBar) {
	if mv.hScrollBar != nil {
		mv.hScrollBar.OnChange = nil
	}
	mv.hScrollBar = sb
	if sb != nil {
		sb.OnChange = func(val int) { mv.deltaX = val }
		mv.syncScrollBars()
	}
}

// SetBounds updates the viewer bounds and syncs scrollbars.
func (mv *MarkdownViewer) SetBounds(bounds Rect) {
	mv.BaseView.SetBounds(bounds)
	mv.syncScrollBars()
}

// SetState propagates focus state to bound scrollbars.
func (mv *MarkdownViewer) SetState(flag ViewState, on bool) {
	mv.BaseView.SetState(flag, on)
	if flag == SfSelected {
		if mv.vScrollBar != nil {
			mv.vScrollBar.SetState(SfVisible, on)
		}
		if mv.hScrollBar != nil {
			mv.hScrollBar.SetState(SfVisible, on)
		}
	}
}

// renderer returns an mdRenderer configured from the current viewer state.
func (mv *MarkdownViewer) renderer() *mdRenderer {
	return &mdRenderer{
		blocks:   mv.blocks,
		width:    mv.Bounds().Width(),
		wrapText: mv.wrapText,
		cs:       mv.ColorScheme(),
	}
}

// syncScrollBars updates the bound scrollbar ranges based on current content.
func (mv *MarkdownViewer) syncScrollBars() {
	r := mv.renderer()
	totalH := r.renderedHeight()
	vpH := mv.Bounds().Height()

	// Clamp deltaY within valid range when content exists
	if len(mv.blocks) > 0 {
		maxDY := totalH - vpH
		if maxDY < 0 {
			maxDY = 0
		}
		if mv.deltaY > maxDY {
			mv.deltaY = maxDY
		}
	}
	if mv.deltaY < 0 {
		mv.deltaY = 0
	}

	if mv.vScrollBar != nil {
		maxRange := totalH - 1 + vpH
		if maxRange < 0 {
			maxRange = 0
		}
		mv.vScrollBar.SetRange(0, maxRange)
		mv.vScrollBar.SetPageSize(vpH)
		mv.vScrollBar.SetValue(mv.deltaY)
	}

	maxW := r.maxContentWidth()
	vpW := mv.Bounds().Width()

	// Clamp deltaX within valid range when content exists
	if len(mv.blocks) > 0 {
		maxDX := maxW - vpW
		if maxDX < 0 {
			maxDX = 0
		}
		if mv.deltaX > maxDX {
			mv.deltaX = maxDX
		}
	}
	if mv.deltaX < 0 {
		mv.deltaX = 0
	}

	if mv.hScrollBar != nil {
		maxRange := maxW - 1 + vpW
		if maxRange < 0 {
			maxRange = 0
		}
		mv.hScrollBar.SetRange(0, maxRange)
		mv.hScrollBar.SetPageSize(vpW)
		mv.hScrollBar.SetValue(mv.deltaX)
	}
}

// Draw renders the visible portion of the markdown content.
func (mv *MarkdownViewer) Draw(buf *DrawBuffer) {
	w := mv.Bounds().Width()
	h := mv.Bounds().Height()
	cs := mv.ColorScheme()
	normalStyle := tcell.StyleDefault
	if cs != nil {
		normalStyle = cs.MarkdownNormal
	}
	buf.Fill(NewRect(0, 0, w, h), ' ', normalStyle)

	if len(mv.blocks) == 0 {
		return
	}

	r := mv.renderer()
	for row := 0; row < h; row++ {
		lineY := mv.deltaY + row
		r.renderLineInto(buf, lineY, row, mv.deltaX, w)
	}
}

// HandleEvent handles mouse and keyboard events for scrolling and interaction.
func (mv *MarkdownViewer) HandleEvent(event *Event) {
	// Mouse handling
	if event.What == EvMouse && event.Mouse != nil {
		mv.BaseView.HandleEvent(event)
		if event.IsCleared() {
			return
		}

		switch {
		case event.Mouse.Button&tcell.WheelUp != 0:
			mv.deltaY -= 3
			if mv.deltaY < 0 {
				mv.deltaY = 0
			}
			mv.syncScrollBars()
			event.Clear()
		case event.Mouse.Button&tcell.WheelDown != 0:
			mv.deltaY += 3
			mv.syncScrollBars()
			event.Clear()
		case event.Mouse.Button&tcell.WheelLeft != 0:
			mv.deltaX -= 3
			if mv.deltaX < 0 {
				mv.deltaX = 0
			}
			mv.syncScrollBars()
			event.Clear()
		case event.Mouse.Button&tcell.WheelRight != 0:
			mv.deltaX += 3
			mv.syncScrollBars()
			event.Clear()
		}
		return
	}

	// Keyboard handling
	if event.What != EvKeyboard || event.Key == nil {
		return
	}
	if !mv.HasState(SfSelected) {
		return
	}

	r := mv.renderer()
	totalH := r.renderedHeight()
	vpH := mv.Bounds().Height()

	// W toggle
	if event.Key.Key == tcell.KeyRune && (event.Key.Rune == 'w' || event.Key.Rune == 'W') {
		mv.SetWrapText(!mv.wrapText)
		event.Clear()
		return
	}

	switch event.Key.Key {
	case tcell.KeyUp:
		if mv.deltaY > 0 {
			mv.deltaY--
			mv.syncScrollBars()
		}
		event.Clear()
	case tcell.KeyDown:
		mv.deltaY++
		mv.syncScrollBars()
		event.Clear()
	case tcell.KeyPgUp:
		mv.deltaY -= vpH
		if mv.deltaY < 0 {
			mv.deltaY = 0
		}
		mv.syncScrollBars()
		event.Clear()
	case tcell.KeyPgDn:
		mv.deltaY += vpH
		if len(mv.blocks) > 0 {
			maxDY := totalH - vpH
			if maxDY < 0 {
				maxDY = 0
			}
			if mv.deltaY > maxDY {
				mv.deltaY = maxDY
			}
		}
		mv.syncScrollBars()
		event.Clear()
	case tcell.KeyHome:
		mv.deltaY = 0
		mv.syncScrollBars()
		event.Clear()
	case tcell.KeyEnd:
		maxDY := totalH - vpH
		if maxDY < 0 {
			maxDY = 0
		}
		mv.deltaY = maxDY
		mv.syncScrollBars()
		event.Clear()
	case tcell.KeyLeft:
		if mv.deltaX > 0 {
			mv.deltaX--
			mv.syncScrollBars()
		}
		event.Clear()
	case tcell.KeyRight:
		mv.deltaX++
		mv.syncScrollBars()
		event.Clear()
	}
}
