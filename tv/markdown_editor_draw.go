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
		if me.HasState(SfSelected) && cs != nil {
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

	// Overlay revealed markers
	me.overlayRevealSpans(buf, w, h, cs)
}

// HandleEvent overrides Editor.HandleEvent to add markdown-aware keyboard
// shortcuts, smart list continuation, undo coalescing, paste handling, and
// auto-reparse on edit.
//
// Dispatch order:
//  1. Ctrl+T - toggle source mode (must be before showSource guard)
//  2. showSource guard - delegate entirely to Editor
//  3. Broadcast forward (CmIndicatorUpdate)
//  4. Keyboard shortcuts (Ctrl+B, Ctrl+I, Ctrl+K, Enter on link, Ctrl+V) - before Memo
//  5. Smart list indent/outdent (Tab/Shift-Tab) - before Memo
//  6. Undo coalescing: classify, save, dispatch to Memo directly
//  7. Post-edit: reparse, list continuation, and link indicator
func (me *MarkdownEditor) HandleEvent(event *Event) {
	// 1. Ctrl+T toggles source mode — MUST be before showSource guard
	if event.What == EvKeyboard && event.Key != nil && event.Key.Key == tcell.KeyCtrlT {
		me.showSource = !me.showSource
		if !me.showSource {
			me.reparse()
		}
		event.Clear()
		return
	}

	// 2. If in source mode, delegate completely to Editor
	if me.showSource {
		me.Editor.HandleEvent(event)
		return
	}

	// 3. Forward broadcast events to Editor (for status line)
	if event.What == EvBroadcast && event.Command == CmIndicatorUpdate {
		me.Editor.HandleEvent(event)
		return
	}

	// 4. Keyboard shortcuts (before Memo consumes the keystroke)
	if event.What == EvKeyboard && event.Key != nil {
		k := event.Key
		// 4a. Ctrl+B / Ctrl+I toggle format
		if k.Key == tcell.KeyCtrlB && k.Modifiers == tcell.ModNone {
			me.toggleFormat("**")
			event.Clear()
			return
		}
		if k.Key == tcell.KeyCtrlI && k.Modifiers == tcell.ModNone {
			me.toggleFormat("*")
			event.Clear()
			return
		}
		// 4b. Ctrl+K — open link dialog
		if k.Key == tcell.KeyCtrlK && k.Modifiers == tcell.ModNone {
			me.openLinkDialog()
			event.Clear()
			return
		}
		// 4c. Enter on a link opens link dialog (before Memo inserts newline)
		if k.Key == tcell.KeyEnter && me.findLinkAt(me.Memo.cursorRow, me.Memo.cursorCol) != nil {
			me.openLinkDialog()
			event.Clear()
			return
		}
		// 4d. Ctrl+V / Ctrl+Shift+V paste handling
		if k.Key == tcell.KeyCtrlV {
			forcePlain := k.Modifiers&tcell.ModShift != 0
			me.pushUndo()
			me.streakSaved = false
			me.lastEditKind = editOther
			me.handlePaste(forcePlain)
			me.reparse()
			me.updateLinkIndicator()
			event.Clear()
			return
		}
	}

	// 5. Smart list indent/outdent (Tab/Shift-Tab, before Memo)
	if me.handleListIndent(event) {
		event.Clear()
		return
	}

	// 6. Undo coalescing and dispatch to Memo.
	//
	// We call Memo.HandleEvent directly rather than Editor.HandleEvent because
	// Editor.HandleEvent unconditionally calls saveSnapshot() before every edit
	// key — which would defeat our coalescing. Instead we manage save/restore
	// ourselves via the undo stack and only save at meaningful boundaries.
	//
	// Keyboard events that are not edit keys (arrows, Home, End, etc.) are still
	// dispatched to Memo, which handles cursor movement and selection. Non-edit
	// keys break the coalescing streak.
	if event.What == EvKeyboard && event.Key != nil && !event.IsCleared() {
		k := event.Key

		// 6a. Ctrl+Z undo — restore most recent snapshot
		if k.Key == tcell.KeyCtrlZ {
			me.popUndo()
			me.reparse()
			me.updateLinkIndicator()
			me.streakSaved = false
			me.lastEditKind = editNone
			event.Clear()
			return
		}

		// 6b. Classify the event for undo coalescing
		kind := me.classifyEvent(event)

		// 6c. Save snapshot at meaningful boundaries:
		//   - editOther always saves (Enter, paste, format, word boundaries)
		//   - editChar/editBackspace save only when starting a new streak
		shouldSave := false
		if kind == editOther {
			shouldSave = true
			me.streakSaved = false // reset so next editChar starts a new streak
		} else if (kind == editChar || kind == editBackspace) && !me.streakSaved {
			shouldSave = true
			me.streakSaved = true
		}
		if shouldSave {
			me.pushUndo()
		}

		// 6d. Dispatch to Memo directly (bypass Editor to avoid its saveSnapshot)
		me.Memo.HandleEvent(event)

		// 6e. Post-edit
		if event.IsCleared() {
			if kind == editNone {
				// Non-edit key (arrow, click) — breaks the coalescing streak
				me.streakSaved = false
				me.lastEditKind = editNone
			} else {
				me.lastEditKind = kind
				me.reparse()
				// List continuation after Enter.
				// Also reset the coalescing streak so the next edit
				// (on the new line) starts a fresh undo unit.
				if k.Key == tcell.KeyEnter {
					me.streakSaved = false
					if me.listEnterContinuation() {
						me.reparse()
					}
				}
				me.updateLinkIndicator()
			}
		}
		return
	}

	// 7. For non-keyboard events (mouse, etc.), delegate to Memo.
	me.Memo.HandleEvent(event)

	// Post-edit reparse for non-keyboard events that were consumed (e.g. mouse
	// clicks that move cursor).
	if event.IsCleared() {
		me.reparse()
		me.updateLinkIndicator()
	}
}

// syncScrollBars overrides Memo.syncScrollBars to use rendered content
// dimensions from mdRenderer instead of raw Memo line counts. This matches
// the MarkdownViewer.syncScrollBars pattern.
func (me *MarkdownEditor) syncScrollBars() {
	if me.showSource {
		me.Editor.Memo.syncScrollBars()
		return
	}

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
// and the color scheme from the widget. revealIndent is computed from reveal
// spans so that rendered content does not overlap with revealed markers.
// Reveal indent is only applied when the widget is selected (SfSelected).
func (me *MarkdownEditor) renderer() *mdRenderer {
	revealIndent := 0
	if me.HasState(SfSelected) {
		for _, s := range me.revealSpans {
			if s.kind == revealBlock {
				revealIndent = len([]rune(s.markerOpen))
				break
			}
		}
	}
	return &mdRenderer{
		blocks:       me.blocks,
		width:        me.Bounds().Width(),
		wrapText:     true,
		cs:           me.ColorScheme(),
		revealIndent: revealIndent,
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

// overlayRevealSpans draws revealed syntax markers on top of rendered content.
// It handles two kinds of spans:
//   - revealBlock: drawn at the start of a visual line (left margin).
//   - revealInline: drawn at their source position within the content line.
//
// Reveal is only active when the widget is selected (SfSelected).
func (me *MarkdownEditor) overlayRevealSpans(buf *DrawBuffer, w, h int, cs *theme.ColorScheme) {
	if len(me.revealSpans) == 0 || cs == nil || !me.HasState(SfSelected) {
		return
	}

	style := cs.MarkdownBlockquote

	// ---- Block-level markers (drawn at visual-line starts) ----
	spanBySourceRow := make(map[int]string)
	for _, s := range me.revealSpans {
		if s.kind == revealBlock {
			spanBySourceRow[s.startRow] = s.markerOpen
		}
	}

	if len(spanBySourceRow) > 0 {
		me.walkRevealVisual(me.blocks, 0, func(visLine, srcRow, indent int) bool {
			screenY := visLine - me.Memo.deltaY
			if screenY < 0 {
				return true // continue
			}
			if screenY >= h {
				return false // stop
			}

			marker, ok := spanBySourceRow[srcRow]
			if !ok {
				return true
			}

			x := indent - me.Memo.deltaX
			for _, ch := range marker {
				if x >= 0 && x < w {
					buf.WriteChar(x, screenY, ch, style)
				}
				x++
			}
			return true
		})
	}

	// ---- Inline markers (drawn at source-position screen coordinates) ----
	for i := range me.revealSpans {
		s := &me.revealSpans[i]
		if s.kind != revealInline {
			continue
		}

		// Draw opening marker at the span's start position.
		scrY, scrX := me.sourceToScreen(s.startRow, s.startCol)
		for _, ch := range s.markerOpen {
			if scrX >= 0 && scrX < w && scrY >= 0 && scrY < h {
				buf.WriteChar(scrX, scrY, ch, style)
			}
			scrX++
		}

		// Draw closing marker.  endCol is inclusive, so the closing marker
		// starts at (endCol - len(markerClose) + 1).
		closeStart := s.endCol - len([]rune(s.markerClose)) + 1
		scrY, scrX = me.sourceToScreen(s.endRow, closeStart)
		for _, ch := range s.markerClose {
			if scrX >= 0 && scrX < w && scrY >= 0 && scrY < h {
				buf.WriteChar(scrX, scrY, ch, style)
			}
			scrX++
		}
	}
}

// walkRevealVisual walks blocks tracking visual line position and calls fn for
// each visual line with its corresponding source row and indentation level.
// It stops when fn returns false.
func (me *MarkdownEditor) walkRevealVisual(blocks []mdBlock, depth int, fn func(visLine, srcRow, indent int) bool) {
	cur := 0
	_ = me.walkBlocksForReveal(blocks, depth, &cur, fn)
}

func (me *MarkdownEditor) walkBlocksForReveal(blocks []mdBlock, depth int, cur *int, fn func(visLine, srcRow, indent int) bool) bool {
	srcRow := 0

	// Walk source to find non-blank lines and map them to blocks
	blockIdx := 0
	for srcRow < len(me.Memo.lines) {
		if strings.TrimSpace(string(me.Memo.lines[srcRow])) == "" {
			srcRow++
			continue
		}
		if blockIdx >= len(blocks) {
			break
		}

		b := blocks[blockIdx]
		if !me.walkBlockForReveal(b, srcRow, depth, cur, fn) {
			return false
		}
		srcRow += blockSourceLineCount(b, me.Memo.lines, srcRow)
		blockIdx++
	}
	return true
}

func (me *MarkdownEditor) walkBlockForReveal(b mdBlock, srcRow, depth int, cur *int, fn func(visLine, srcRow, indent int) bool) bool {
	indent := depth * 2

	switch b.kind {
	case blockParagraph, blockHeader:
		r := me.renderer()
		runs := b.runs
		avail := r.width - indent
		if avail < 1 {
			avail = 1
		}
		var lines [][]mdRun
		if r.wrapText && len(runs) > 0 {
			lines = wrapRuns(runs, avail)
		} else {
			lines = [][]mdRun{runs}
		}
		for range lines {
			if !fn(*cur, srcRow, indent) {
				return false
			}
			*cur++
		}
		return true

	case blockCodeBlock:
		if len(b.code) == 0 {
			if !fn(*cur, srcRow, indent) {
				return false
			}
			*cur++
			return true
		}
		for range b.code {
			if !fn(*cur, srcRow, indent) {
				return false
			}
			*cur++
			srcRow++
		}
		return true

	case blockBulletList, blockNumberList, blockCheckList:
		markerWidth := 4
		itemIndent := depth*2 + markerWidth
		r := me.renderer()
		avail := r.width - itemIndent
		if avail < 1 {
			avail = 1
		}

		itemSrcRow := srcRow
		for _, item := range b.items {
			var lines [][]mdRun
			if r.wrapText && len(item.runs) > 0 {
				lines = wrapRuns(item.runs, avail)
			} else {
				lines = [][]mdRun{item.runs}
			}
			for lineIdx := range lines {
				// Report the source row only on the first visual line of the item
				reportedSrcRow := -1
				if lineIdx == 0 {
					reportedSrcRow = itemSrcRow
				}
				if !fn(*cur, reportedSrcRow, itemIndent) {
					return false
				}
				*cur++
			}
			itemSrcRow++

			// Nested children
			if len(item.children) > 0 {
				childSrcRow := itemSrcRow
				for _, child := range item.children {
					if !me.walkBlockForReveal(child, childSrcRow, depth+1, cur, fn) {
						return false
					}
					childSrcRow += blockSourceLineCount(child, me.Memo.lines, childSrcRow)
				}
				itemSrcRow = childSrcRow
			}
		}
		return true

	case blockBlockquote:
		childSrcRow := srcRow
		for i, child := range b.children {
			if i > 0 {
				if !fn(*cur, -1, indent) {
					return false
				}
				*cur++
			}
			if !me.walkBlockForReveal(child, childSrcRow, depth+1, cur, fn) {
				return false
			}
			childSrcRow += blockSourceLineCount(child, me.Memo.lines, childSrcRow)
		}
		return true

	case blockTable:
		r := me.renderer()
		avail := r.width - indent
		if avail < 1 {
			avail = 1
		}
		colWidths := layoutTable(b, avail)
		if colWidths == nil {
			return true
		}
		eff := computeEffectiveWidths(colWidths, avail, r.wrapText)

		// Top border
		if !fn(*cur, srcRow, indent) {
			return false
		}
		*cur++

		// Headers
		headerH := r.tableRowHeight(b.headers, eff)
		for h := 0; h < headerH; h++ {
			if !fn(*cur, srcRow, indent) {
				return false
			}
			*cur++
		}
		srcRow++

		// Header separator
		if !fn(*cur, srcRow, indent) {
			return false
		}
		*cur++
		srcRow++

		// Data rows
		for _, row := range b.rows {
			rowH := r.tableRowHeight(row, eff)
			for h := 0; h < rowH; h++ {
				if !fn(*cur, srcRow, indent) {
					return false
				}
				*cur++
			}
			srcRow++
		}

		// Bottom border
		if !fn(*cur, -1, indent) {
			return false
		}
		*cur++
		return true

	case blockHRule:
		return fn(*cur, srcRow, indent)

	case blockDefList:
		defIndent := 4
		r := me.renderer()
		defAvail := r.width - depth*2 - defIndent
		if defAvail < 1 {
			defAvail = 1
		}

		itemSrcRow := srcRow
		for i, item := range b.items {
			if i > 0 {
				if !fn(*cur, -1, indent) {
					return false
				}
				*cur++
			}
			// Term
			if !fn(*cur, itemSrcRow, depth*2) {
				return false
			}
			*cur++
			itemSrcRow++

			// Definition
			var defLines [][]mdRun
			if r.wrapText && len(item.runs) > 0 {
				defLines = wrapRuns(item.runs, defAvail)
			} else {
				defLines = [][]mdRun{item.runs}
			}
			for lineIdx := range defLines {
				reportedSrcRow := -1
				if lineIdx == 0 {
					reportedSrcRow = itemSrcRow
				}
				if !fn(*cur, reportedSrcRow, indent+defIndent) {
					return false
				}
				*cur++
			}
			itemSrcRow++

			// Children
			if len(item.children) > 0 {
				childSrcRow := itemSrcRow
				for _, child := range item.children {
					if !me.walkBlockForReveal(child, childSrcRow, depth+1, cur, fn) {
						return false
					}
					childSrcRow += blockSourceLineCount(child, me.Memo.lines, childSrcRow)
				}
				itemSrcRow = childSrcRow
			}
		}
		return true
	}

	return true
}
