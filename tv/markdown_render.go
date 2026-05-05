package tv

import (
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

	// Determine effective column widths for wrapping.  When the total
	// table width exceeds the available width and wrapping is enabled,
	// scale column widths down proportionally so that cells wrap to fit
	// within the viewport instead of at their ideal (overflow) widths.
	effectiveWidths := make([]int, len(colWidths))
	copy(effectiveWidths, colWidths)

	totalTableWidth := len(colWidths) + 1 // border overhead
	for _, w := range colWidths {
		totalTableWidth += w
	}
	if r.wrapText && totalTableWidth > availWidth && availWidth > len(colWidths)+1 {
		contentAvail := availWidth - (len(colWidths) + 1)
		totalCol := 0
		for _, w := range colWidths {
			totalCol += w
		}
		allocated := 0
		for i := range effectiveWidths {
			effectiveWidths[i] = colWidths[i] * contentAvail / totalCol
			if effectiveWidths[i] < 1 {
				effectiveWidths[i] = 1
			}
			allocated += effectiveWidths[i]
		}
		for i := 0; allocated < contentAvail && i < len(effectiveWidths); i++ {
			effectiveWidths[i]++
			allocated++
		}
	}

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
