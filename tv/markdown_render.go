package tv

import (
	"fmt"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// =============================================================================
// wrapRuns - word-wrap runs to fit within maxWidth
// =============================================================================

// wrapRuns word-wraps a slice of runs to fit within maxWidth columns.
// Returns one []mdRun per visual line.
//
// - Wraps at word boundaries (spaces)
// - If a single word exceeds width, break at width boundary
// - Preserves run styles (and url) across line breaks
// - Empty input -> [][]mdRun{nil}
// - maxWidth <= 0 -> [][]mdRun{nil}
func wrapRuns(runs []mdRun, maxWidth int) [][]mdRun {
	if maxWidth <= 0 {
		return [][]mdRun{nil}
	}

	var lines [][]mdRun
	var curLine []mdRun
	col := 0

	// placeRun appends text to the current line, merging with the last run
	// if it has the same style and url.
	placeRun := func(text string, style mdRunStyle, url string) {
		if len(curLine) > 0 && curLine[len(curLine)-1].style == style && curLine[len(curLine)-1].url == url {
			curLine[len(curLine)-1].text += text
		} else {
			curLine = append(curLine, mdRun{text: text, style: style, url: url})
		}
	}

	finalizeLine := func() {
		lines = append(lines, curLine)
		curLine = nil
		col = 0
	}

	for _, run := range runs {
		tokens := splitWordTokens(run.text)
		for _, tok := range tokens {
			isSpace := isAllSpaces(tok)
			tokRunes := []rune(tok)

			if isSpace && col == 0 {
				continue // drop leading spaces
			}

			for len(tokRunes) > 0 {
				tokLen := len(tokRunes)
				if col+tokLen <= maxWidth {
					// Token fits entirely on the current line
					placeRun(string(tokRunes), run.style, run.url)
					col += tokLen
					tokRunes = nil
				} else if col > 0 {
					// Token doesn't fit; try to fill what remains of the current line
					avail := maxWidth - col
					if avail > 0 {
						placeRun(string(tokRunes[:avail]), run.style, run.url)
						tokRunes = tokRunes[avail:]
					}
					// Start a new line
					finalizeLine()
					// Drop leading spaces on the new line
					if isSpace {
						tokRunes = nil
					}
				} else {
					// col == 0: word is at the start of a line but too long
					// Hard-break at the width boundary
					placeRun(string(tokRunes[:maxWidth]), run.style, run.url)
					tokRunes = tokRunes[maxWidth:]
					finalizeLine()
				}
			}
		}
	}

	// Append the final (possibly empty) line
	if curLine != nil || len(lines) == 0 {
		lines = append(lines, curLine)
	}

	return lines
}

// splitWordTokens splits text into alternating non-space / space tokens.
// Examples:
//
//	"hello world"  -> ["hello", " ", "world"]
//	"  hi  there " -> ["  ", "hi", "  ", "there", " "]
//	"word"         -> ["word"]
func splitWordTokens(text string) []string {
	var tokens []string
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		// Collect a non-space token
		if !unicode.IsSpace(runes[i]) {
			start := i
			for i < len(runes) && !unicode.IsSpace(runes[i]) {
				i++
			}
			tokens = append(tokens, string(runes[start:i]))
			continue
		}
		// Collect a space token
		start := i
		for i < len(runes) && unicode.IsSpace(runes[i]) {
			i++
		}
		tokens = append(tokens, string(runes[start:i]))
	}
	return tokens
}

// isAllSpaces returns true if every rune in s is a Unicode space character.
func isAllSpaces(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// =============================================================================
// composeStyle - combine block style with inline run style
// =============================================================================

// composeStyle combines a block-level background Style with inline
// foreground/attributes from the ColorScheme based on the run style.
//
// - nil cs -> returns blockStyle unchanged
// - runNormal -> blockStyle unchanged
// - runBold -> overlayFgAttrs(blockStyle, cs.MarkdownBold)
// - runItalic -> overlayFgAttrs(blockStyle, cs.MarkdownItalic)
// - runBoldItalic -> overlayFgAttrs(blockStyle, cs.MarkdownBoldItalic)
// - runCode -> cs.MarkdownCode (own background)
// - runLink -> overlayFgAttrs(blockStyle, cs.MarkdownLink)
// - runStrikethrough -> overlayFgAttrs(blockStyle, cs.MarkdownBold).StrikeThrough(true)
func composeStyle(blockStyle tcell.Style, runStyle mdRunStyle, cs *theme.ColorScheme) tcell.Style {
	if cs == nil {
		return blockStyle
	}

	switch runStyle {
	case runNormal:
		return blockStyle
	case runBold:
		return overlayFgAttrs(blockStyle, cs.MarkdownBold)
	case runItalic:
		return overlayFgAttrs(blockStyle, cs.MarkdownItalic)
	case runBoldItalic:
		return overlayFgAttrs(blockStyle, cs.MarkdownBoldItalic)
	case runCode:
		return cs.MarkdownCode
	case runLink:
		return overlayFgAttrs(blockStyle, cs.MarkdownLink)
	case runStrikethrough:
		return overlayFgAttrs(blockStyle, cs.MarkdownBold).StrikeThrough(true)
	default:
		return blockStyle
	}
}

// overlayFgAttrs copies the foreground and text attributes from src onto dst,
// preserving dst's background.
func overlayFgAttrs(dst, src tcell.Style) tcell.Style {
	fg, _, srcAttrs := src.Decompose()
	_, bg, _ := dst.Decompose()
	return tcell.StyleDefault.
		Foreground(fg).
		Background(bg).
		Attributes(srcAttrs)
}

// =============================================================================
// mdRenderer - computes layout metrics for rendered markdown blocks
// =============================================================================

// mdRenderer computes layout metrics for a set of parsed markdown blocks.
type mdRenderer struct {
	blocks   []mdBlock
	width    int
	wrapText bool
	cs       *theme.ColorScheme
}

// renderedHeight returns the total height (in lines) needed to render all blocks.
func (r *mdRenderer) renderedHeight() int {
	return r.blocksHeight(r.blocks, 0)
}

// blocksHeight returns the total height for a slice of blocks at a given depth.
func (r *mdRenderer) blocksHeight(blocks []mdBlock, depth int) int {
	h := 0
	for i, b := range blocks {
		if i > 0 {
			h++ // blank line between blocks
		}
		h += r.blockHeight(b, depth)
	}
	return h
}

// blockHeight returns the height for a single block at a given depth.
func (r *mdRenderer) blockHeight(b mdBlock, depth int) int {
	indent := depth * 2
	avail := r.width - indent
	if avail < 1 {
		avail = 1
	}

	switch b.kind {
	case blockParagraph, blockHeader:
		if r.wrapText {
			return len(wrapRuns(b.runs, avail))
		}
		return 1
	case blockCodeBlock:
		if len(b.code) == 0 {
			return 1
		}
		return len(b.code)
	case blockBulletList, blockNumberList, blockCheckList:
		return r.listHeight(b, depth)
	case blockBlockquote:
		return r.blocksHeight(b.children, depth+1)
	case blockTable:
		return r.tableHeight(b, avail)
	case blockHRule:
		return 1
	case blockDefList:
		return r.defListHeight(b, depth)
	}
	return 1
}

// listHeight returns the height for a list block (bullet, numbered, or checklist).
func (r *mdRenderer) listHeight(b mdBlock, depth int) int {
	markerWidth := 4
	itemIndent := depth*2 + markerWidth
	avail := r.width - itemIndent
	if avail < 1 {
		avail = 1
	}

	h := 0
	for _, item := range b.items {
		if r.wrapText {
			lines := wrapRuns(item.runs, avail)
			h += len(lines)
		} else {
			h++ // one line per item
		}
		// Add nested children at depth+1
		h += r.blocksHeight(item.children, depth+1)
	}
	return h
}

// defListHeight returns the height for a definition list block.
func (r *mdRenderer) defListHeight(b mdBlock, depth int) int {
	defIndent := 4
	defAvail := r.width - depth*2 - defIndent
	if defAvail < 1 {
		defAvail = 1
	}

	h := 0
	for i, item := range b.items {
		if i > 0 {
			h++ // blank line between items
		}
		// Term: always 1 line
		h++
		// Definition: wrapped if enabled
		if r.wrapText {
			h += len(wrapRuns(item.runs, defAvail))
		} else {
			h++
		}
		if len(item.children) > 0 {
			h += r.blocksHeight(item.children, depth+1)
		}
	}
	return h
}

// tableHeight returns the height for a table block.
func (r *mdRenderer) tableHeight(b mdBlock, availWidth int) int {
	colWidths := layoutTable(b, availWidth)
	if colWidths == nil {
		return 0
	}
	effectiveWidths := computeEffectiveWidths(colWidths, availWidth, r.wrapText)

	h := 2 // top + bottom borders

	if len(b.headers) > 0 {
		h += r.tableRowHeight(b.headers, effectiveWidths)
		h++ // separator
	}

	for _, row := range b.rows {
		h += r.tableRowHeight(row, effectiveWidths)
	}

	return h
}

// tableRowHeight returns the height of a single table row given column widths.
// When not wrapping, every row is 1 line.  When wrapping, it is the maximum
// number of wrapped lines across all cells in the row.
func (r *mdRenderer) tableRowHeight(row [][]mdRun, colWidths []int) int {
	if !r.wrapText {
		return 1
	}
	rowH := 0
	for j, cell := range row {
		if j >= len(colWidths) {
			break
		}
		n := len(wrapRuns(cell, colWidths[j]))
		if n > rowH {
			rowH = n
		}
	}
	if rowH == 0 {
		rowH = 1
	}
	return rowH
}

// =============================================================================
// layoutTable - compute column widths for a table
// =============================================================================

// layoutTable determines column widths for a table block.
//
// - Empty table (no headers, no rows) returns nil.
// - Compute minWidths (floor 8) based on longest word in each column.
// - Compute contentWidths based on maximum cell text length in each column.
// - If minimums + borders fit: distribute extra space proportionally to content widths.
// - If minimums + borders don't fit: return minimums.
func layoutTable(b mdBlock, availWidth int) []int {
	numCols := 0
	if len(b.headers) > 0 {
		numCols = len(b.headers)
	} else if len(b.rows) > 0 {
		numCols = len(b.rows[0])
	}
	if numCols == 0 {
		return nil
	}

	// Compute per-column metrics: longestWord and maxContentLength.
	longestWord := make([]int, numCols)
	contentLength := make([]int, numCols)

	// Helper: update metrics for a cell.
	updateCell := func(j int, cell []mdRun) {
		text := allRunText(cell)
		// Longest word in this cell
		lw := longestWordInText(text)
		if lw > longestWord[j] {
			longestWord[j] = lw
		}
		// Total content length
		cl := len([]rune(text))
		if cl > contentLength[j] {
			contentLength[j] = cl
		}
	}

	// Scan headers.
	for j, cell := range b.headers {
		updateCell(j, cell)
	}
	// Scan rows.
	for _, row := range b.rows {
		for j, cell := range row {
			if j < numCols {
				updateCell(j, cell)
			}
		}
	}

	// Compute minimum widths (floor 8).
	minWidths := make([]int, numCols)
	for j := 0; j < numCols; j++ {
		mw := longestWord[j]
		if mw < 8 {
			mw = 8
		}
		minWidths[j] = mw
	}

	borderOverhead := numCols + 1
	totalMin := borderOverhead
	for _, mw := range minWidths {
		totalMin += mw
	}

	if totalMin >= availWidth {
		// Not enough space; return minimum widths.
		return minWidths
	}

	// Distribute extra space proportionally by content width.
	extra := availWidth - totalMin

	totalContent := 0
	for _, cl := range contentLength {
		totalContent += cl
	}

	result := make([]int, numCols)
	distributed := 0
	for j := 0; j < numCols; j++ {
		var share int
		if totalContent > 0 {
			share = extra * contentLength[j] / totalContent
		}
		result[j] = minWidths[j] + share
		distributed += share
	}

	// Distribute remainder (from integer division) to first columns.
	remainder := extra - distributed
	for j := 0; j < numCols && remainder > 0; j++ {
		result[j]++
		remainder--
	}

	return result
}

// longestWordInText returns the length of the longest contiguous non-space word
// in the given text.
func longestWordInText(text string) int {
	maxLen := 0
	currentLen := 0
	for _, r := range text {
		if unicode.IsSpace(r) {
			if currentLen > maxLen {
				maxLen = currentLen
			}
			currentLen = 0
		} else {
			currentLen++
		}
	}
	if currentLen > maxLen {
		maxLen = currentLen
	}
	return maxLen
}

// allRunText concatenates all run texts into a single string.
func allRunText(runs []mdRun) string {
	var s string
	for _, r := range runs {
		s += r.text
	}
	return s
}

// runLen returns the total rune count of all runs combined.
func runLen(runs []mdRun) int {
	n := 0
	for _, r := range runs {
		n += len([]rune(r.text))
	}
	return n
}

// =============================================================================
// maxContentWidth — maximum rendered line width for horizontal scroll
// =============================================================================

// maxContentWidth returns the maximum rendered content width in characters,
// used to determine horizontal scroll range.
func (r *mdRenderer) maxContentWidth() int {
	return r.blocksMaxWidth(r.blocks, 0)
}

// blocksMaxWidth computes max width across a slice of blocks.
func (r *mdRenderer) blocksMaxWidth(blocks []mdBlock, depth int) int {
	maxW := 0
	for _, b := range blocks {
		w := r.blockMaxWidth(b, depth)
		if w > maxW {
			maxW = w
		}
	}
	return maxW
}

// blockMaxWidth computes the maximum rendered width for a single block.
func (r *mdRenderer) blockMaxWidth(b mdBlock, depth int) int {
	indent := depth * 2
	avail := r.width - indent
	if avail < 1 {
		avail = 1
	}

	switch b.kind {
	case blockParagraph:
		if r.wrapText {
			maxW := indent
			for _, line := range wrapRuns(b.runs, avail) {
				w := indent + runLen(line)
				if w > maxW {
					maxW = w
				}
			}
			return maxW
		}
		if len(b.runs) == 0 {
			return indent
		}
		return indent + runLen(b.runs)

	case blockHeader:
		if r.wrapText {
			maxW := indent
			for _, line := range wrapRuns(b.runs, avail) {
				w := indent + runLen(line)
				if w > maxW {
					maxW = w
				}
			}
			return maxW
		}
		return indent + runLen(b.runs)

	case blockCodeBlock:
		maxW := indent
		for _, line := range b.code {
			w := indent + len([]rune(line))
			if w > maxW {
				maxW = w
			}
		}
		return maxW

	case blockBulletList, blockNumberList, blockCheckList:
		markerWidth := 4
		itemIndent := depth*2 + markerWidth
		itemAvail := r.width - itemIndent
		if itemAvail < 1 {
			itemAvail = 1
		}
		maxW := 0
		for _, item := range b.items {
			if len(item.runs) > 0 {
				if r.wrapText {
					for _, line := range wrapRuns(item.runs, itemAvail) {
						w := itemIndent + runLen(line)
						if w > maxW {
							maxW = w
						}
					}
				} else {
					w := itemIndent + runLen(item.runs)
					if w > maxW {
						maxW = w
					}
				}
			}
			cw := r.blocksMaxWidth(item.children, depth+1)
			if cw > maxW {
				maxW = cw
			}
		}
		return maxW

	case blockBlockquote:
		return r.blocksMaxWidth(b.children, depth+1)

	case blockTable:
		colWidths := layoutTable(b, avail)
		if colWidths == nil {
			return indent
		}
		eff := computeEffectiveWidths(colWidths, avail, r.wrapText)
		totalW := len(colWidths) + 1 // borders
		for _, cw := range eff {
			totalW += cw
		}
		return indent + totalW

	case blockHRule:
		return indent + 1

	case blockDefList:
		defIndent := 4
		defAvail := r.width - depth*2 - defIndent
		if defAvail < 1 {
			defAvail = 1
		}
		maxW := 0
		for _, item := range b.items {
			w := depth*2 + runLen(item.term)
			if w > maxW {
				maxW = w
			}
			if len(item.runs) > 0 {
				if r.wrapText {
					for _, line := range wrapRuns(item.runs, defAvail) {
						w := depth*2 + defIndent + runLen(line)
						if w > maxW {
							maxW = w
						}
					}
				} else {
					w := depth*2 + defIndent + runLen(item.runs)
					if w > maxW {
						maxW = w
					}
				}
			}
			cw := r.blocksMaxWidth(item.children, depth+1)
			if cw > maxW {
				maxW = cw
			}
		}
		return maxW
	}
	return 0
}

// =============================================================================
// renderLineInto — renders a single visual line into a DrawBuffer
// =============================================================================

// renderLineInto renders a single visual line (document line index lineY) into
// the DrawBuffer at the given screen row, applying horizontal scroll offset dx.
func (r *mdRenderer) renderLineInto(buf *DrawBuffer, lineY, screenY, dx, w int) {
	if r.cs == nil {
		return
	}
	cur := 0
	for i, b := range r.blocks {
		if i > 0 {
			if cur == lineY {
				return // blank separator line; already background-filled
			}
			cur++
		}
		if r.renderBlockLine(buf, b, lineY, screenY, dx, w, 0, &cur) {
			return
		}
	}
}

// renderBlockLine dispatches to the appropriate block-type renderer.
func (r *mdRenderer) renderBlockLine(buf *DrawBuffer, b mdBlock, lineY, screenY, dx, w, depth int, cur *int) bool {
	switch b.kind {
	case blockParagraph:
		return r.renderParagraphLine(buf, b, lineY, screenY, dx, w, depth, cur)
	case blockHeader:
		return r.renderHeaderLine(buf, b, lineY, screenY, dx, w, depth, cur)
	case blockCodeBlock:
		return r.renderCodeBlockLine(buf, b, lineY, screenY, dx, w, depth, cur)
	case blockBulletList, blockNumberList, blockCheckList:
		return r.renderListLine(buf, b, lineY, screenY, dx, w, depth, cur)
	case blockBlockquote:
		return r.renderBlockquoteLine(buf, b, lineY, screenY, dx, w, depth, cur)
	case blockTable:
		return r.renderTableLine(buf, b, lineY, screenY, dx, w, depth, cur)
	case blockHRule:
		return r.renderHRuleLine(buf, b, lineY, screenY, dx, w, depth, cur)
	case blockDefList:
		return r.renderDefListLine(buf, b, lineY, screenY, dx, w, depth, cur)
	}
	return false
}

// =============================================================================
// Paragraph rendering
// =============================================================================

func (r *mdRenderer) renderParagraphLine(buf *DrawBuffer, b mdBlock, lineY, screenY, dx, w, depth int, cur *int) bool {
	indent := depth * 2
	avail := w - indent
	if avail < 1 {
		avail = 1
	}

	var lines [][]mdRun
	if r.wrapText && len(b.runs) > 0 {
		lines = wrapRuns(b.runs, avail)
	} else {
		lines = [][]mdRun{b.runs}
	}

	normalStyle := r.cs.MarkdownNormal

	for _, line := range lines {
		if *cur == lineY {
			x := indent - dx
			for _, run := range line {
				s := composeStyle(normalStyle, run.style, r.cs)
				for _, ch := range run.text {
					if x >= 0 && x < w {
						buf.WriteChar(x, screenY, ch, s)
					}
					x++
				}
			}
			return true
		}
		*cur++
	}
	return false
}

// =============================================================================
// Header rendering
// =============================================================================

func (r *mdRenderer) renderHeaderLine(buf *DrawBuffer, b mdBlock, lineY, screenY, dx, w, depth int, cur *int) bool {
	indent := depth * 2
	avail := w - indent
	if avail < 1 {
		avail = 1
	}

	var lines [][]mdRun
	if r.wrapText && len(b.runs) > 0 {
		lines = wrapRuns(b.runs, avail)
	} else {
		lines = [][]mdRun{b.runs}
	}

	headerStyle := r.cs.MarkdownNormal
	switch b.level {
	case 1:
		headerStyle = r.cs.MarkdownH1
	case 2:
		headerStyle = r.cs.MarkdownH2
	case 3:
		headerStyle = r.cs.MarkdownH3
	case 4:
		headerStyle = r.cs.MarkdownH4
	case 5:
		headerStyle = r.cs.MarkdownH5
	case 6:
		headerStyle = r.cs.MarkdownH6
	}

	for _, line := range lines {
		if *cur == lineY {
			x := indent - dx
			for _, run := range line {
				s := composeStyle(headerStyle, run.style, r.cs)
				for _, ch := range run.text {
					if x >= 0 && x < w {
						buf.WriteChar(x, screenY, ch, s)
					}
					x++
				}
			}
			return true
		}
		*cur++
	}
	return false
}

// =============================================================================
// Code block rendering
// =============================================================================

func (r *mdRenderer) renderCodeBlockLine(buf *DrawBuffer, b mdBlock, lineY, screenY, dx, w, depth int, cur *int) bool {
	indent := depth * 2
	codeStyle := r.cs.MarkdownCodeBlock

	// Fill row with code block background, starting at indent offset
	startX := indent - dx
	if startX < 0 {
		startX = 0
	}
	for x := startX; x < w; x++ {
		buf.WriteChar(x, screenY, ' ', codeStyle)
	}

	// If no code lines, still render one background row
	if len(b.code) == 0 {
		if *cur == lineY {
			return true
		}
		*cur++
		return false
	}

	for _, codeLine := range b.code {
		if *cur == lineY {
			// Render code text on top of background
			x := indent - dx
			for _, ch := range codeLine {
				if x >= 0 && x < w {
					buf.WriteChar(x, screenY, ch, codeStyle)
				}
				x++
			}
			return true
		}
		*cur++
	}
	return false
}

// =============================================================================
// List rendering
// =============================================================================

func (r *mdRenderer) renderListLine(buf *DrawBuffer, b mdBlock, lineY, screenY, dx, w, depth int, cur *int) bool {
	markerWidth := 4
	itemIndent := depth*2 + markerWidth
	avail := w - itemIndent
	if avail < 1 {
		avail = 1
	}

	markerStyle := r.cs.MarkdownListMarker
	normalStyle := r.cs.MarkdownNormal

	for itemNum, item := range b.items {
		var lines [][]mdRun
		if r.wrapText && len(item.runs) > 0 {
			lines = wrapRuns(item.runs, avail)
		} else {
			lines = [][]mdRun{item.runs}
		}

		for lineIdx, line := range lines {
			if *cur == lineY {
				// Render marker on first line only
				if lineIdx == 0 {
					x := depth*2 - dx
					marker := listMarkerText(b.kind, itemNum, item)
					for _, ch := range marker {
						if x >= 0 && x < w {
							buf.WriteChar(x, screenY, ch, markerStyle)
						}
						x++
					}
				}
				// Render content
				x := itemIndent - dx
				for _, run := range line {
					s := composeStyle(normalStyle, run.style, r.cs)
					for _, ch := range run.text {
						if x >= 0 && x < w {
							buf.WriteChar(x, screenY, ch, s)
						}
						x++
					}
				}
				return true
			}
			*cur++
		}

		// Nested children
		if len(item.children) > 0 {
			if r.renderBlocksInto(buf, item.children, lineY, screenY, dx, w, depth+1, cur) {
				return true
			}
		}
	}
	return false
}

// listMarkerText returns the marker string for a list item.
func listMarkerText(kind mdBlockKind, itemNum int, item mdItem) string {
	switch kind {
	case blockBulletList:
		return "• "
	case blockNumberList:
		return fmt.Sprintf("%d. ", itemNum+1)
	case blockCheckList:
		if item.checked != nil && *item.checked {
			return "☑ "
		}
		return "☐ "
	}
	return ""
}

// =============================================================================
// Blockquote rendering
// =============================================================================

func (r *mdRenderer) renderBlockquoteLine(buf *DrawBuffer, b mdBlock, lineY, screenY, dx, w, depth int, cur *int) bool {
	barStyle := r.cs.MarkdownBlockquote

	for i, child := range b.children {
		if i > 0 {
			if *cur == lineY {
				// Draw bar on separator line
				x := depth*2 - dx
				if x >= 0 && x < w {
					buf.WriteChar(x, screenY, '▌', barStyle)
				}
				return true
			}
			*cur++
		}
		if r.renderBlockLine(buf, child, lineY, screenY, dx, w, depth+1, cur) {
			// Overlay bar on the rendered content
			x := depth*2 - dx
			if x >= 0 && x < w {
				buf.WriteChar(x, screenY, '▌', barStyle)
			}
			return true
		}
	}
	return false
}

// =============================================================================
// Horizontal rule rendering
// =============================================================================

func (r *mdRenderer) renderHRuleLine(buf *DrawBuffer, b mdBlock, lineY, screenY, dx, w, depth int, cur *int) bool {
	if *cur == lineY {
		hrStyle := r.cs.MarkdownHRule
		indent := depth * 2
		for x := 0; x < w; x++ {
			// The hrule character starts after indent, but scroll offset affects position
			// We fill with hrule chars considering the indent
			absX := x - (indent - dx)
			if absX >= 0 {
				buf.WriteChar(x, screenY, '─', hrStyle)
			}
		}
		return true
	}
	*cur++
	return false
}

// =============================================================================
// Definition list rendering
// =============================================================================

func (r *mdRenderer) renderDefListLine(buf *DrawBuffer, b mdBlock, lineY, screenY, dx, w, depth int, cur *int) bool {
	defIndent := 4
	defAvail := w - depth*2 - defIndent
	if defAvail < 1 {
		defAvail = 1
	}

	termStyle := r.cs.MarkdownDefTerm
	normalStyle := r.cs.MarkdownNormal

	for i, item := range b.items {
		if i > 0 {
			if *cur == lineY {
				return true // blank separator line
			}
			*cur++
		}
		// Term: 1 line
		if *cur == lineY {
			x := depth*2 - dx
			for _, run := range item.term {
				s := composeStyle(termStyle, run.style, r.cs)
				for _, ch := range run.text {
					if x >= 0 && x < w {
						buf.WriteChar(x, screenY, ch, s)
					}
					x++
				}
			}
			return true
		}
		*cur++

		// Definition: wrapped lines
		var defLines [][]mdRun
		if r.wrapText && len(item.runs) > 0 {
			defLines = wrapRuns(item.runs, defAvail)
		} else {
			defLines = [][]mdRun{item.runs}
		}
		for _, line := range defLines {
			if *cur == lineY {
				x := depth*2 + defIndent - dx
				for _, run := range line {
					s := composeStyle(normalStyle, run.style, r.cs)
					for _, ch := range run.text {
						if x >= 0 && x < w {
							buf.WriteChar(x, screenY, ch, s)
						}
						x++
					}
				}
				return true
			}
			*cur++
		}

		// Nested children
		if len(item.children) > 0 {
			if r.renderBlocksInto(buf, item.children, lineY, screenY, dx, w, depth+1, cur) {
				return true
			}
		}
	}
	return false
}

// =============================================================================
// Table rendering
// =============================================================================

func (r *mdRenderer) renderTableLine(buf *DrawBuffer, b mdBlock, lineY, screenY, dx, w, depth int, cur *int) bool {
	indent := depth * 2
	avail := w - indent
	if avail < 1 {
		avail = 1
	}

	colWidths := layoutTable(b, avail)
	if colWidths == nil {
		return false
	}

	eff := computeEffectiveWidths(colWidths, avail, r.wrapText)

	borderStyle := r.cs.MarkdownTableBorder
	normalStyle := r.cs.MarkdownNormal

	// Top border
	if *cur == lineY {
		renderTableBorder(buf, screenY, dx, indent, eff, borderStyle, '┌', '┬', '┐')
		return true
	}
	*cur++

	// Headers
	if len(b.headers) > 0 {
		headerH := r.tableRowHeight(b.headers, eff)
		for h := 0; h < headerH; h++ {
			if *cur == lineY {
				renderTableDataRow(buf, screenY, dx, indent, b.headers, eff, h, borderStyle, r.cs.MarkdownBold, r.wrapText, r.cs)
				return true
			}
			*cur++
		}

		// Header separator
		if *cur == lineY {
			renderTableBorder(buf, screenY, dx, indent, eff, borderStyle, '├', '┼', '┤')
			return true
		}
		*cur++
	}

	// Body rows
	for _, row := range b.rows {
		rowH := r.tableRowHeight(row, eff)
		for h := 0; h < rowH; h++ {
			if *cur == lineY {
				renderTableDataRow(buf, screenY, dx, indent, row, eff, h, borderStyle, normalStyle, r.wrapText, r.cs)
				return true
			}
			*cur++
		}
	}

	// Bottom border
	if *cur == lineY {
		renderTableBorder(buf, screenY, dx, indent, eff, borderStyle, '└', '┴', '┘')
		return true
	}
	*cur++

	return false
}

// renderTableBorder draws a horizontal table border line.
func renderTableBorder(buf *DrawBuffer, screenY, dx, indent int, colWidths []int, style tcell.Style, left, sep, right rune) {
	w := buf.Width()
	x := indent - dx

	// Left corner
	if x >= 0 && x < w {
		buf.WriteChar(x, screenY, left, style)
	}
	x++

	for j, cw := range colWidths {
		for k := 0; k < cw; k++ {
			if x >= 0 && x < w {
				buf.WriteChar(x, screenY, '─', style)
			}
			x++
		}
		if j < len(colWidths)-1 {
			if x >= 0 && x < w {
				buf.WriteChar(x, screenY, sep, style)
			}
		} else {
			if x >= 0 && x < w {
				buf.WriteChar(x, screenY, right, style)
			}
		}
		x++
	}
}

// renderTableDataRow renders one visual line of table cell content.
func renderTableDataRow(buf *DrawBuffer, screenY, dx, indent int, row [][]mdRun, colWidths []int, lineIdx int, borderStyle, normalStyle tcell.Style, wrapText bool, cs *theme.ColorScheme) {
	w := buf.Width()
	x := indent - dx

	// Left border
	if x >= 0 && x < w {
		buf.WriteChar(x, screenY, '│', borderStyle)
	}
	x++

	for j, cell := range row {
		if j >= len(colWidths) {
			break
		}
		cw := colWidths[j]

		// Get cell runs for this visual line
		var cellRuns []mdRun
		if wrapText && len(cell) > 0 {
			lines := wrapRuns(cell, cw)
			if lineIdx < len(lines) {
				cellRuns = lines[lineIdx]
			}
		} else if !wrapText && lineIdx == 0 {
			cellRuns = cell
		}

		// Render content
		rendered := 0
		for _, run := range cellRuns {
			s := composeStyle(normalStyle, run.style, cs)
			for _, ch := range run.text {
				if x >= 0 && x < w {
					buf.WriteChar(x, screenY, ch, s)
				}
				x++
				rendered++
			}
		}

		// Pad remaining cell width
		for rendered < cw {
			if x >= 0 && x < w {
				buf.WriteChar(x, screenY, ' ', normalStyle)
			}
			x++
			rendered++
		}

		// Column separator (or right border for last column)
		if x >= 0 && x < w {
			buf.WriteChar(x, screenY, '│', borderStyle)
		}
		x++
	}
}

// =============================================================================
// Blocks-level rendering (handles blank lines between blocks)
// =============================================================================

// renderBlocksInto walks a slice of blocks, incrementing cur for each visual
// line, and renders the target lineY when reached. Returns true when rendered.
func (r *mdRenderer) renderBlocksInto(buf *DrawBuffer, blocks []mdBlock, lineY, screenY, dx, w, depth int, cur *int) bool {
	for i, b := range blocks {
		if i > 0 {
			if *cur == lineY {
				return true // blank separator line; background already filled
			}
			*cur++
		}
		if r.renderBlockLine(buf, b, lineY, screenY, dx, w, depth, cur) {
			return true
		}
	}
	return false
}

// =============================================================================
// computeEffectiveWidths — scaled column widths for tables
// =============================================================================

// computeEffectiveWidths scales column widths down when the table exceeds the
// available width and wrapping is enabled, similar to tableHeight logic.
func computeEffectiveWidths(colWidths []int, availWidth int, wrapText bool) []int {
	effective := make([]int, len(colWidths))
	copy(effective, colWidths)

	if !wrapText {
		return effective
	}

	totalTableWidth := len(colWidths) + 1 // border overhead
	for _, w := range colWidths {
		totalTableWidth += w
	}
	if totalTableWidth <= availWidth {
		return effective
	}
	if availWidth <= len(colWidths)+1 {
		return effective
	}

	contentAvail := availWidth - (len(colWidths) + 1)
	totalCol := 0
	for _, w := range colWidths {
		totalCol += w
	}
	if totalCol == 0 {
		return effective
	}

	allocated := 0
	for i := range effective {
		effective[i] = colWidths[i] * contentAvail / totalCol
		if effective[i] < 1 {
			effective[i] = 1
		}
		allocated += effective[i]
	}
	for i := 0; allocated < contentAvail && i < len(effective); i++ {
		effective[i]++
		allocated++
	}

	return effective
}
