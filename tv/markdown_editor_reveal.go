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
		for _, item := range b.items {
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
	return me.revealSpans
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
