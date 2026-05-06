package tv

import (
	"strings"
)

// revealKind distinguishes between block-level and inline-level reveal spans.
type revealKind int

const (
	revealBlock  revealKind = iota
	revealInline
)

// revealSpan describes a region of source text where a markdown syntax marker
// should be revealed (displayed in dimmed style).
type revealSpan struct {
	startRow, startCol int
	endRow, endCol     int
	markerOpen         string
	markerClose        string
	kind               revealKind
}

// detectListMarker detects list marker syntax at the start of a line.
// Supported patterns:
//   - Bullet: "- ", "* ", "+ "
//   - Numbered: "N. " (e.g., "1. ", "99. ")
//   - Checklist: "- [ ] ", "- [x] ", "- [X] "
//   - Indented: pairs of "  " before the marker
//
// Returns the marker text (including indent), remaining content, and
// whether the line is a list item.
func detectListMarker(line string) (marker string, rest string, isListItem bool) {
	// Collect leading indent (pairs of two spaces)
	indent := ""
	rest2 := line
	for strings.HasPrefix(rest2, "  ") {
		indent += "  "
		rest2 = rest2[2:]
	}

	trimmed := rest2

	// Check checklist: "- [ ] ", "- [x] ", "- [X] "
	if len(trimmed) >= 6 &&
		trimmed[0] == '-' && trimmed[1] == ' ' && trimmed[2] == '[' &&
		(trimmed[3] == ' ' || trimmed[3] == 'x' || trimmed[3] == 'X') &&
		trimmed[4] == ']' && trimmed[5] == ' ' {
		marker = indent + "- [" + string(trimmed[3]) + "] "
		rest = trimmed[6:]
		return marker, rest, true
	}

	// Check bullet: "- ", "* ", "+ "
	if len(trimmed) >= 2 {
		c := trimmed[0]
		if (c == '-' || c == '*' || c == '+') && trimmed[1] == ' ' {
			marker = indent + string(c) + " "
			rest = trimmed[2:]
			return marker, rest, true
		}
	}

	// Check numbered: "N. " where N is one or more digits
	if len(trimmed) >= 3 && trimmed[0] >= '0' && trimmed[0] <= '9' {
		for i := 1; i < len(trimmed); i++ {
			if trimmed[i] == '.' && i+1 < len(trimmed) && trimmed[i+1] == ' ' {
				marker = indent + trimmed[:i+2]
				rest = trimmed[i+2:]
				return marker, rest, true
			}
			if trimmed[i] < '0' || trimmed[i] > '9' {
				break
			}
		}
	}

	return "", "", false
}

// blockSourceLineCount returns the number of source lines a block occupies
// in the raw source text, starting at startRow.
func blockSourceLineCount(b mdBlock, source [][]rune, startRow int) int {
	switch b.kind {
	case blockParagraph, blockHeader:
		return 1
	case blockCodeBlock:
		return len(b.code) + 2
	case blockHRule:
		return 1
	case blockBulletList, blockNumberList, blockCheckList:
		count := 0
		row := startRow
		for _, item := range b.items {
			count++ // this item's source line
			row++
			for _, child := range item.children {
				childCount := blockSourceLineCount(child, source, row)
				count += childCount
				row += childCount
			}
		}
		return count
	case blockBlockquote:
		// Count consecutive source lines that belong to this blockquote.
		// goldmark may merge multiple "> " lines into a single paragraph
		// with soft line breaks, so counting children gives 1 instead of N.
		// Walk the source directly: a line belongs to the blockquote if it
		// starts with ">" (after optional indent), until a blank line.
		count := 0
		for i := startRow; i < len(source); i++ {
			rawLine := string(source[i])
			if strings.TrimSpace(rawLine) == "" {
				break
			}
			trimmed := strings.TrimLeft(rawLine, " ")
			if strings.HasPrefix(trimmed, ">") {
				count++
			} else {
				break
			}
		}
		return count
	case blockTable:
		return len(b.rows) + 2
	case blockDefList:
		count := 0
		row := startRow
		for i, item := range b.items {
			if i > 0 {
				count++ // blank line between items
				row++
			}
			count++ // term line
			row++
			count++ // definition line
			row++
			for _, child := range item.children {
				childCount := blockSourceLineCount(child, source, row)
				count += childCount
				row += childCount
			}
		}
		return count
	default:
		return 1
	}
}

// blockRevealSpan returns reveal spans for a single block and all its nested
// content, anchored at blockRow in the source.
func blockRevealSpan(b mdBlock, source [][]rune, blockRow, depth int) []revealSpan {
	switch b.kind {
	case blockHeader:
		marker := strings.Repeat("#", b.level) + " "
		return []revealSpan{{
			startRow:    blockRow,
			startCol:    0,
			endRow:      blockRow + 1,
			endCol:      0,
			markerOpen:  marker,
			markerClose: "",
			kind:        revealBlock,
		}}

	case blockCodeBlock:
		openMarker := "```"
		if b.language != "" {
			openMarker += b.language
		}
		closeRow := blockRow + len(b.code) + 1
		return []revealSpan{
			{
				startRow:    blockRow,
				startCol:    0,
				endRow:      blockRow + 1,
				endCol:      0,
				markerOpen:  openMarker,
				markerClose: "",
				kind:        revealBlock,
			},
			{
				startRow:    closeRow,
				startCol:    0,
				endRow:      closeRow + 1,
				endCol:      0,
				markerOpen:  "```",
				markerClose: "",
				kind:        revealBlock,
			},
		}

	case blockHRule:
		return []revealSpan{{
			startRow:    blockRow,
			startCol:    0,
			endRow:      blockRow + 1,
			endCol:      0,
			markerOpen:  "---",
			markerClose: "",
			kind:        revealBlock,
		}}

	case blockBulletList, blockNumberList, blockCheckList:
		var spans []revealSpan
		itemRow := blockRow
		for _, item := range b.items {
			marker, _, _ := detectListMarker(string(source[itemRow]))
			spans = append(spans, revealSpan{
				startRow:    itemRow,
				startCol:    0,
				endRow:      itemRow + 1,
				endCol:      0,
				markerOpen:  marker,
				markerClose: "",
				kind:        revealBlock,
			})
			itemRow++
			for _, child := range item.children {
				childRow := itemRow
				spans = append(spans, blockRevealSpan(child, source, childRow, depth+1)...)
				itemRow += blockSourceLineCount(child, source, childRow)
			}
		}
		return spans

	case blockBlockquote:
		var spans []revealSpan
		// Count ">" prefixes in source to handle nesting depth.
		// goldmark may merge multiple blockquote lines into a single paragraph,
		// so we count markers directly from source rather than from children.
		for i := blockRow; i < len(source); i++ {
			rawLine := string(source[i])
			if strings.TrimSpace(rawLine) == "" {
				break
			}
			trimmed := strings.TrimLeft(rawLine, " ")
			if !strings.HasPrefix(trimmed, ">") {
				break
			}
			// Count consecutive ">" prefixes to determine nesting depth
			nestingCount := 0
			rest := trimmed
			for strings.HasPrefix(rest, ">") {
				nestingCount++
				rest = rest[1:]
				rest = strings.TrimLeft(rest, " ")
			}
			// Generate one "> " marker per nesting level
			for n := 0; n < nestingCount; n++ {
				spans = append(spans, revealSpan{
					startRow:    i,
					startCol:    0,
					endRow:      i + 1,
					endCol:      0,
					markerOpen:  "> ",
					markerClose: "",
					kind:        revealBlock,
				})
			}
		}
		return spans

	case blockTable:
		var spans []revealSpan
		row := blockRow
		// Header row
		spans = append(spans, revealSpan{
			startRow:    row,
			startCol:    0,
			endRow:      row + 1,
			endCol:      0,
			markerOpen:  "| ",
			markerClose: "",
			kind:        revealBlock,
		})
		row++
		// Separator row
		spans = append(spans, revealSpan{
			startRow:    row,
			startCol:    0,
			endRow:      row + 1,
			endCol:      0,
			markerOpen:  "| ",
			markerClose: "",
			kind:        revealBlock,
		})
		row++
		// Data rows
		for range b.rows {
			spans = append(spans, revealSpan{
				startRow:    row,
				startCol:    0,
				endRow:      row + 1,
				endCol:      0,
				markerOpen:  "| ",
				markerClose: "",
				kind:        revealBlock,
			})
			row++
		}
		return spans

	case blockParagraph:
		return nil

	default:
		return nil
	}
}

// collectRevealSpans walks blocks recursively, tracking source row position.
// When curRow falls within a block's source range, it calls blockRevealSpan
// to generate all spans for that block (which handles nested recursion internally).
// Only the block containing the cursor produces spans.
func collectRevealSpans(blocks []mdBlock, source [][]rune, curRow, curCol, depth int, row *int) []revealSpan {
	srcRow := 0
	blockIdx := 0

	for srcRow < len(source) {
		// Skip blank source lines (they separate blocks)
		if strings.TrimSpace(string(source[srcRow])) == "" {
			srcRow++
			continue
		}

		if blockIdx >= len(blocks) {
			break
		}

		lineCount := blockSourceLineCount(blocks[blockIdx], source, srcRow)
		endRow := srcRow + lineCount

		if curRow >= srcRow && curRow < endRow {
			// Cursor is within this block's source range
			spans := blockRevealSpan(blocks[blockIdx], source, srcRow, depth)
			*row = endRow
			return spans
		}

		srcRow += lineCount
		blockIdx++
	}

	return nil
}

// buildRevealSpans clears and rebuilds the revealSpans field based on the
// current cursor position and parsed blocks. Returns the spans for use by
// callers that need the value directly.
func (me *MarkdownEditor) buildRevealSpans() []revealSpan {
	me.revealSpans = nil

	if len(me.blocks) == 0 || len(me.Memo.lines) == 0 {
		return nil
	}

	var endRow int
	me.revealSpans = collectRevealSpans(me.blocks, me.Memo.lines, me.Memo.cursorRow, me.Memo.cursorCol, 0, &endRow)
	if me.revealSpans == nil {
		me.revealSpans = []revealSpan{}
	}

	// Append inline-level spans from scanInlineMarkers.
	inlineSpans := scanInlineMarkers(me.Memo.lines, me.blocks, me.Memo.cursorRow, me.Memo.cursorCol)
	me.revealSpans = append(me.revealSpans, inlineSpans...)

	return me.revealSpans
}

// scanInlineMarkers walks blocks to find inline syntax markers (bold, italic,
// code, strikethrough) in the source text. It matches found markers to the
// block's mdRun inline styles from goldmark's parse tree and returns only the
// revealSpan for the span containing (or adjacent to) the cursor position.
//
// Adjacent means the cursor is within 1 character of the span bounds: on the
// marker itself, immediately before the opening marker, or immediately after
// the closing marker.
//
// If the cursor is not inside or adjacent to any inline span, returns nil.
func scanInlineMarkers(source [][]rune, blocks []mdBlock, cursorRow int, cursorCol int) []revealSpan {
	if len(source) == 0 || len(blocks) == 0 {
		return nil
	}

	srcRow := 0
	blockIdx := 0

	for srcRow < len(source) && blockIdx < len(blocks) {
		// Skip blank source lines (they separate blocks).
		if strings.TrimSpace(string(source[srcRow])) == "" {
			srcRow++
			continue
		}

		b := blocks[blockIdx]
		lineCount := blockSourceLineCount(b, source, srcRow)

		// Check if cursorRow falls within this block's source range.
		if cursorRow >= srcRow && cursorRow < srcRow+lineCount {
			spans := blockInlineSpans(b, source, srcRow)
			for _, s := range spans {
				if cursorRow >= s.startRow && cursorRow <= s.endRow {
					if cursorCol >= s.startCol-1 && cursorCol <= s.endCol+1 {
						return []revealSpan{s}
					}
				}
			}
			return nil
		}

		srcRow += lineCount
		blockIdx++
	}

	return nil
}

// blockquotePrefixSkip computes the number of columns to skip at the start
// of a blockquote source line to get past ">" prefixes and whitespace.
// Example: "  > > text" -> skip skips past "  > > " (the leading spaces and
// all ">" prefixes and their trailing spaces).
func blockquotePrefixSkip(line []rune) int {
	rawLine := string(line)
	trimmed := strings.TrimLeft(rawLine, " ")
	skip := len(rawLine) - len(trimmed) // leading spaces
	rest := trimmed
	for strings.HasPrefix(rest, ">") {
		rest = rest[1:]
		skip++
		if strings.HasPrefix(rest, " ") {
			rest = rest[1:]
			skip++
		}
	}
	return skip
}

// defListPrefixSkip computes the number of columns to skip for a definition
// list definition line (": " or "~ " prefix plus leading whitespace).
func defListPrefixSkip(line []rune) int {
	rawLine := string(line)
	trimmed := strings.TrimLeft(rawLine, " ")
	skip := len(rawLine) - len(trimmed)
	if strings.HasPrefix(trimmed, ": ") || strings.HasPrefix(trimmed, "~ ") {
		skip += 2
	} else if strings.HasPrefix(trimmed, ":") || strings.HasPrefix(trimmed, "~") {
		skip++
	}
	return skip
}

// blockInlineSpans returns all inline reveal spans for a block, anchored at
// blockSrcRow in the source. It dispatches to the appropriate handler based
// on block kind.
func blockInlineSpans(b mdBlock, source [][]rune, blockSrcRow int) []revealSpan {
	switch b.kind {
	case blockParagraph:
		if blockSrcRow >= len(source) {
			return nil
		}
		return lineInlineSpans(b.runs, source[blockSrcRow], blockSrcRow, 0)

	case blockHeader:
		if blockSrcRow >= len(source) {
			return nil
		}
		// Header marker is "#"*level + " "; skip it to find inline content.
		skip := b.level + 1
		return lineInlineSpans(b.runs, source[blockSrcRow], blockSrcRow, skip)

	case blockBulletList, blockNumberList, blockCheckList:
		var allSpans []revealSpan
		itemRow := blockSrcRow
		for _, item := range b.items {
			if itemRow < len(source) && len(item.runs) > 0 {
				marker, _, _ := detectListMarker(string(source[itemRow]))
				skip := len([]rune(marker))
				allSpans = append(allSpans,
					lineInlineSpans(item.runs, source[itemRow], itemRow, skip)...)
			}
			itemRow++
			for _, child := range item.children {
				allSpans = append(allSpans,
					blockInlineSpans(child, source, itemRow)...)
				itemRow += blockSourceLineCount(child, source, itemRow)
			}
		}
		return allSpans

	case blockBlockquote:
		var allSpans []revealSpan
		childRow := blockSrcRow
		for _, child := range b.children {
			if childRow >= len(source) {
				break
			}
			skip := blockquotePrefixSkip(source[childRow])

			// Dispatch based on child kind, adjusting skip for the
			// blockquote ">" prefix.
			switch child.kind {
			case blockParagraph:
				allSpans = append(allSpans,
					lineInlineSpans(child.runs, source[childRow], childRow, skip)...)
			case blockHeader:
				// Header marker comes after the ">" prefix.
				headerSkip := skip + child.level + 1
				allSpans = append(allSpans,
					lineInlineSpans(child.runs, source[childRow], childRow, headerSkip)...)
			default:
				allSpans = append(allSpans,
					blockInlineSpans(child, source, childRow)...)
			}
			childRow += blockSourceLineCount(child, source, childRow)
		}
		return allSpans

	case blockTable:
		var allSpans []revealSpan
		row := blockSrcRow

		// Header row (if present).
		if len(b.headers) > 0 && row < len(source) {
			allSpans = append(allSpans,
				tableRowInlineSpans(b.headers, source[row], row)...)
			row++
		}
		// Separator row.
		row++

		// Body rows.
		for _, bodyRow := range b.rows {
			if row >= len(source) {
				break
			}
			allSpans = append(allSpans,
				tableRowInlineSpans(bodyRow, source[row], row)...)
			row++
		}
		return allSpans

	case blockDefList:
		var allSpans []revealSpan
		itemRow := blockSrcRow
		for i, item := range b.items {
			if i > 0 {
				itemRow++ // blank line between items
			}
			// Term line.
			if itemRow < len(source) && len(item.term) > 0 {
				allSpans = append(allSpans,
					lineInlineSpans(item.term, source[itemRow], itemRow, 0)...)
			}
			itemRow++
			// Definition line.
			if itemRow < len(source) && len(item.runs) > 0 {
				skip := defListPrefixSkip(source[itemRow])
				allSpans = append(allSpans,
					lineInlineSpans(item.runs, source[itemRow], itemRow, skip)...)
			}
			itemRow++
			// Nested children.
			for _, child := range item.children {
				allSpans = append(allSpans,
					blockInlineSpans(child, source, itemRow)...)
				itemRow += blockSourceLineCount(child, source, itemRow)
			}
		}
		return allSpans

	default:
		return nil
	}
}

// tableRowInlineSpans scans a table row source line for inline syntax markers
// in each cell, using the parsed cell runs.
func tableRowInlineSpans(cells [][]mdRun, line []rune, srcRow int) []revealSpan {
	var allSpans []revealSpan
	pos := 0

	// Skip the opening "|".
	for pos < len(line) && line[pos] == '|' {
		pos++
	}

	for _, cell := range cells {
		// Skip whitespace before cell content.
		for pos < len(line) && (line[pos] == ' ' || line[pos] == '\t') {
			pos++
		}

		cellStart := pos
		if cellStart < len(line) && len(cell) > 0 {
			allSpans = append(allSpans,
				lineInlineSpans(cell, line, srcRow, cellStart)...)
		}

		// Advance past this cell to the next "|" separator.
		// Walk past the rendered cell content and any inline markers.
		for pos < len(line) && line[pos] != '|' {
			pos++
		}
		if pos < len(line) {
			pos++ // skip "|"
		}
	}
	return allSpans
}

// lineInlineSpans scans a single source line (as []rune) for inline syntax
// markers that correspond to the given runs. The skip parameter is the number
// of columns to skip at the start of the line (e.g., for header markers).
// Returned spans have column positions relative to the full source line.
func lineInlineSpans(runs []mdRun, line []rune, srcRow, skip int) []revealSpan {
	if skip >= len(line) {
		return nil
	}

	var spans []revealSpan
	srcPos := 0 // position within the content part of the line (after skip)

	for _, run := range runs {
		// Allocate enough runes for the run text to avoid re-allocation.
		runRunes := []rune(run.text)
		runLen := len(runRunes)

		if run.style == runNormal || run.style == runLink {
			// Plain text: advance past it in the source.  The content
			// starting at (skip + srcPos) must match run.text.
			if skip+srcPos+runLen <= len(line) {
				srcPos += runLen
			} else {
				// Run text exceeds remaining line — stop scanning.
				break
			}
			continue
		}

		// Styled run: detect the opening marker at the current source position.
		openStart := srcPos
		markerOpen := detectInlineMarker(line, skip+srcPos, run.style)
		if markerOpen == "" {
			continue
		}
		srcPos += len([]rune(markerOpen))

		// Content text — skip past it.
		srcPos += runLen

		// Detect the closing marker after the content text.
		// Goldmark may strip trailing whitespace from content text
		// (e.g., "**text  **" -> content is "text", not "text  "),
		// and nested formatting (e.g., "**bold *italic***")
		// puts inner markers before the outer closing marker.
		// Scan forward when the closing marker is not at the
		// expected position.
		closeAbsPos := skip + srcPos
		markerClose := detectInlineMarker(line, closeAbsPos, run.style)
		if markerClose == "" {
			for closeAbsPos < len(line) {
				closeAbsPos++
				markerClose = detectInlineMarker(line, closeAbsPos, run.style)
				if markerClose != "" {
					break
				}
			}
		}
		if markerClose == "" {
			continue
		}
		srcPos = closeAbsPos - skip // realign to found position
		closeLen := len([]rune(markerClose))
		closeEnd := srcPos + closeLen - 1 // inclusive end column in skip-relative coords

		spans = append(spans, revealSpan{
			startRow:    srcRow,
			startCol:    skip + openStart,
			endRow:      srcRow,
			endCol:      skip + closeEnd,
			markerOpen:  markerOpen,
			markerClose: markerClose,
			kind:        revealInline,
		})

		srcPos = closeEnd + 1
	}

	return spans
}

// detectInlineMarker reads the opening or closing marker at the given position
// in the source line for the specified run style. Returns the marker string,
// or "" if no matching marker is found.
func detectInlineMarker(line []rune, pos int, style mdRunStyle) string {
	if pos >= len(line) {
		return ""
	}

	switch style {
	case runBold:
		if pos+2 <= len(line) && line[pos] == '*' && line[pos+1] == '*' {
			return "**"
		}
		if pos+2 <= len(line) && line[pos] == '_' && line[pos+1] == '_' {
			return "__"
		}

	case runItalic:
		if line[pos] == '*' {
			return "*"
		}
		if line[pos] == '_' {
			return "_"
		}

	case runBoldItalic:
		if pos+3 <= len(line) && line[pos] == '*' && line[pos+1] == '*' && line[pos+2] == '*' {
			return "***"
		}
		if pos+3 <= len(line) && line[pos] == '_' && line[pos+1] == '_' && line[pos+2] == '_' {
			return "___"
		}

	case runCode:
		if line[pos] == '`' {
			return "`"
		}

	case runStrikethrough:
		if pos+2 <= len(line) && line[pos] == '~' && line[pos+1] == '~' {
			return "~~"
		}
	}

	return ""
}

// inRevealSpan returns a pointer to the revealSpan containing the given
// (row, col) position, or nil if no span matches.
// When endCol is 0, it is treated as startCol + len(markerOpen).
func inRevealSpan(spans []revealSpan, row, col int) *revealSpan {
	for i := range spans {
		s := &spans[i]
		if row >= s.startRow && row < s.endRow {
			endCol := s.endCol
			if endCol == 0 {
				endCol = s.startCol + len([]rune(s.markerOpen))
			}
			if col >= s.startCol && col <= endCol {
				return s
			}
		}
	}
	return nil
}
