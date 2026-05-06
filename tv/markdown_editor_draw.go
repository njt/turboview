package tv

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// Draw renders the MarkdownEditor. When showSource is true it delegates to
// the underlying Memo. Otherwise it fills the viewport with the MarkdownNormal
// background and renders formatted markdown blocks through an mdRenderer.
func (me *MarkdownEditor) Draw(buf *DrawBuffer) {
	if me.showSource {
		me.Editor.Memo.Draw(buf)
		return
	}

	w := me.Bounds().Width()
	h := me.Bounds().Height()
	cs := me.ColorScheme()
	normalStyle := tcell.StyleDefault
	if cs != nil {
		normalStyle = cs.MarkdownNormal
	}

	buf.Fill(NewRect(0, 0, w, h), ' ', normalStyle)

	if len(me.blocks) == 0 {
		if me.HasState(SfSelected) {
			buf.WriteChar(0, 0, ' ', cs.MemoSelected)
		}
		return
	}

	r := me.renderer()

	// Render viewport + overscan buffer (+5)
	for row := 0; row < h+5; row++ {
		lineY := me.Memo.deltaY + row
		r.renderLineInto(buf, lineY, row, me.Memo.deltaX, w)
	}

	me.drawCursor(buf, cs)
}

// HandleEvent overrides Editor.HandleEvent. When showSource is true it
// delegates entirely to Editor. Otherwise it delegates to Editor for edit
// processing and calls reparse after consumed events to refresh blocks.
func (me *MarkdownEditor) HandleEvent(event *Event) {
	if me.showSource {
		me.Editor.HandleEvent(event)
		return
	}
	if event.What == EvBroadcast && event.Command == CmIndicatorUpdate {
		me.Editor.HandleEvent(event)
		return
	}
	me.Editor.HandleEvent(event)
	if event.IsCleared() {
		me.reparse()
	}
}

// syncScrollBars overrides Memo.syncScrollBars to use rendered content
// dimensions from mdRenderer instead of raw Memo line counts. This matches
// the MarkdownViewer.syncScrollBars pattern.
func (me *MarkdownEditor) syncScrollBars() {
	r := me.renderer()
	totalH := r.renderedHeight()
	vpH := me.Bounds().Height()

	// Clamp deltaY within valid range when content exists
	if len(me.blocks) > 0 {
		maxDY := totalH - vpH
		if maxDY < 0 {
			maxDY = 0
		}
		if me.Memo.deltaY > maxDY {
			me.Memo.deltaY = maxDY
		}
	}
	if me.Memo.deltaY < 0 {
		me.Memo.deltaY = 0
	}

	if me.Memo.vScrollBar != nil {
		maxRange := totalH - 1 + vpH
		if maxRange < 0 {
			maxRange = 0
		}
		me.Memo.vScrollBar.SetRange(0, maxRange)
		me.Memo.vScrollBar.SetPageSize(vpH)
		me.Memo.vScrollBar.SetValue(me.Memo.deltaY)
	}

	maxW := r.maxContentWidth()
	vpW := me.Bounds().Width()

	// Clamp deltaX within valid range when content exists
	if len(me.blocks) > 0 {
		maxDX := maxW - vpW
		if maxDX < 0 {
			maxDX = 0
		}
		if me.Memo.deltaX > maxDX {
			me.Memo.deltaX = maxDX
		}
	}
	if me.Memo.deltaX < 0 {
		me.Memo.deltaX = 0
	}

	if me.Memo.hScrollBar != nil {
		maxRange := maxW - 1 + vpW
		if maxRange < 0 {
			maxRange = 0
		}
		me.Memo.hScrollBar.SetRange(0, maxRange)
		me.Memo.hScrollBar.SetPageSize(vpW)
		me.Memo.hScrollBar.SetValue(me.Memo.deltaX)
	}
}

// renderer returns an mdRenderer configured from the current editor state.
// Blocks come from the last parse, width from bounds, wrapText is always true,
// and the color scheme from the widget.
func (me *MarkdownEditor) renderer() *mdRenderer {
	return &mdRenderer{
		blocks:   me.blocks,
		width:    me.Bounds().Width(),
		wrapText: true,
		cs:       me.ColorScheme(),
	}
}

// sourceToScreen maps a source (row, col) position in Memo.lines to a screen
// position in the rendered markdown output. It walks blocks to find the
// rendered line that contains the given source row, then computes the screen
// column based on block type and indent.
//
// The mapping uses the following strategy:
//   - Count non-blank source lines before the target row to determine which
//     parsed block contains the source position. Each non-blank source line
//     starts a new top-level block (paragraph, header, hrule, etc.).
//   - Walk blocks, accumulating rendered line counts (including blank
//     separators between blocks) to find the screen Y offset.
//   - Screen X is computed as: indent + sourceCol - deltaX, where indent
//     depends on block type (0 for paragraph/header, 4 for lists, 2 for
//     blockquotes).
func (me *MarkdownEditor) sourceToScreen(row, col int) (screenY, screenX int) {
	// Clamp inputs to valid ranges.
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	if len(me.Memo.lines) == 0 || len(me.blocks) == 0 {
		return 0, 0
	}
	if row >= len(me.Memo.lines) {
		row = len(me.Memo.lines) - 1
	}
	if row < 0 {
		return 0, 0
	}
	if col > len(me.Memo.lines[row]) {
		col = len(me.Memo.lines[row])
	}

	// Count non-blank source lines before this row to determine block index.
	nonBlankBefore := 0
	for i := 0; i < row; i++ {
		if strings.TrimSpace(string(me.Memo.lines[i])) != "" {
			nonBlankBefore++
		}
	}

	// If the current source line is blank, back up to the preceding block.
	blockIdx := nonBlankBefore
	if strings.TrimSpace(string(me.Memo.lines[row])) == "" && blockIdx > 0 {
		blockIdx--
	}
	if blockIdx >= len(me.blocks) {
		blockIdx = len(me.blocks) - 1
	}
	if blockIdx < 0 {
		return 0, 0
	}

	// Walk blocks to accumulate the rendered line offset for this block.
	r := me.renderer()
	renderedLine := 0
	for i := 0; i < blockIdx; i++ {
		if i > 0 {
			renderedLine++ // blank separator between blocks
		}
		renderedLine += r.blockHeight(me.blocks[i], 0)
	}
	if blockIdx > 0 {
		renderedLine++ // blank separator before current block
	}

	// Compute indent based on block type.
	indent := 0
	b := me.blocks[blockIdx]
	switch b.kind {
	case blockBulletList, blockNumberList, blockCheckList:
		indent = 4 // marker width
	case blockBlockquote:
		indent = 2 // blockquote indent
	}

	screenY = renderedLine - me.Memo.deltaY
	screenX = indent + col - me.Memo.deltaX

	if screenY < 0 {
		screenY = 0
	}
	if screenX < 0 {
		screenX = 0
	}

	return screenY, screenX
}

// drawCursor renders the block cursor at the mapped screen position. It is
// a no-op when the widget is not selected or when a selection is active.
func (me *MarkdownEditor) drawCursor(buf *DrawBuffer, cs *theme.ColorScheme) {
	if !me.HasState(SfSelected) || me.Memo.HasSelection() {
		return
	}

	screenY, screenX := me.sourceToScreen(me.Memo.cursorRow, me.Memo.cursorCol)

	w := me.Bounds().Width()
	h := me.Bounds().Height()

	if screenX >= 0 && screenX < w && screenY >= 0 && screenY < h {
		ch := rune(' ')
		if me.Memo.cursorRow < len(me.Memo.lines) &&
			me.Memo.cursorCol < len(me.Memo.lines[me.Memo.cursorRow]) {
			ch = me.Memo.lines[me.Memo.cursorRow][me.Memo.cursorCol]
		}
		style := tcell.StyleDefault
		if cs != nil {
			style = cs.MemoSelected
		}
		buf.WriteChar(screenX, screenY, ch, style)
	}
}
