package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// =============================================================================
// Integration tests — IR parser + rendering helpers working together
// =============================================================================

// TestIntegration_ParseAndHeight verifies that parseMarkdown and renderedHeight
// work together correctly for a simple document.
//
//	"# Hello\n\nWorld" → header (1) + blank (1) + paragraph (1) = 3 lines
func TestIntegration_ParseAndHeight(t *testing.T) {
	blocks := parseMarkdown("# Hello\n\nWorld")
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}

	r := mdRenderer{
		blocks:   blocks,
		width:    40,
		wrapText: true,
	}

	h := r.renderedHeight()
	if h != 3 {
		t.Errorf("height = %d, want 3 (header 1 + blank 1 + paragraph 1)", h)
	}
}

// TestIntegration_BoldStyleComposition verifies composeStyle correctly combines
// the MarkdownNormal block background with the MarkdownBold foreground from a
// real parsed bold run, using theme.BorlandBlue.
//
//	Spec: runBold → overlayFgAttrs(blockStyle, cs.MarkdownBold)
//	foreground from MarkdownBold, background from MarkdownNormal
func TestIntegration_BoldStyleComposition(t *testing.T) {
	blocks := parseMarkdown("**bold**")
	if len(blocks) != 1 || len(blocks[0].runs) != 1 {
		t.Fatalf("got %d blocks / %d runs, want 1 block with 1 run", len(blocks), len(blocks[0].runs))
	}

	run := blocks[0].runs[0]
	if run.style != runBold {
		t.Fatalf("run style = %v, want runBold", run.style)
	}

	cs := theme.BorlandBlue
	blockStyle := cs.MarkdownNormal
	result := composeStyle(blockStyle, run.style, cs)

	// Foreground should come from MarkdownBold
	fg, _, _ := result.Decompose()
	wantFG, _, _ := cs.MarkdownBold.Decompose()
	if fg != wantFG {
		t.Errorf("foreground = %v, want %v (from MarkdownBold)", fg, wantFG)
	}

	// Background should be preserved from MarkdownNormal
	_, bg, _ := result.Decompose()
	_, wantBG, _ := blockStyle.Decompose()
	if bg != wantBG {
		t.Errorf("background = %v, want %v (from MarkdownNormal)", bg, wantBG)
	}

	// Bold attribute should be set
	_, _, attrs := result.Decompose()
	if attrs&tcell.AttrBold == 0 {
		t.Error("composed style should have Bold=true")
	}
}

// TestIntegration_CodeBlockHeight verifies that a fenced code block
// parsed from markdown renders with height equal to its line count.
//
//	Spec: Code block height = number of code lines
func TestIntegration_CodeBlockHeight(t *testing.T) {
	blocks := parseMarkdown("```go\nline1\nline2\nline3\n```")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockCodeBlock {
		t.Fatalf("block kind = %v, want blockCodeBlock", blocks[0].kind)
	}
	if len(blocks[0].code) != 3 {
		t.Fatalf("got %d code lines, want 3", len(blocks[0].code))
	}

	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: true,
	}

	h := r.renderedHeight()
	if h != 3 {
		t.Errorf("height = %d, want 3", h)
	}
}

// TestIntegration_NestedListHeight verifies that a nested list produces
// a greater renderedHeight at a narrow width than at a wide width,
// because indentation reduces available width for wrapping.
//
//	Spec: nested list height narrow(30) > wide(80)
func TestIntegration_NestedListHeight(t *testing.T) {
	// Nested list: outer item with long text + inner item with long text.
	// Goldmark requires 4-space indent for nesting.
	src := "- Outer item with long text that will wrap at narrow width\n" +
		"    - Nested item also with long text that wraps at narrow width"

	blocks := parseMarkdown(src)
	if len(blocks) == 0 {
		t.Fatal("no blocks parsed")
	}
	if blocks[0].kind != blockBulletList {
		t.Fatalf("block kind = %v, want blockBulletList", blocks[0].kind)
	}

	rNarrow := mdRenderer{
		blocks:   blocks,
		width:    30,
		wrapText: true,
	}
	rWide := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: true,
	}

	hNarrow := rNarrow.renderedHeight()
	hWide := rWide.renderedHeight()

	if hNarrow <= hWide {
		t.Errorf("narrow height (%d) <= wide height (%d), want narrow > wide due to wrapping", hNarrow, hWide)
	}
}

// TestIntegration_WrapRunsPreservesStyles verifies that wrapRuns returns
// multiple lines for a wide paragraph at a narrow width, and that all
// resulting runs preserve their original styles from the parsed IR.
//
//	Spec: "All resulting runs preserve their styles from the original paragraph"
func TestIntegration_WrapRunsPreservesStyles(t *testing.T) {
	src := "This is **bold text** and *italic text* and some more text here to definitely make it wrap around"
	blocks := parseMarkdown(src)
	if len(blocks) != 1 || blocks[0].kind != blockParagraph {
		t.Fatalf("got %d blocks (kind=%v), want 1 paragraph", len(blocks), blocks[0].kind)
	}
	paragraphRuns := blocks[0].runs

	// Verify the parser produced mixed styles (bold and italic)
	foundBold := false
	foundItalic := false
	for _, r := range paragraphRuns {
		if r.style == runBold {
			foundBold = true
		}
		if r.style == runItalic {
			foundItalic = true
		}
	}
	if !foundBold || !foundItalic {
		t.Fatal("expected both bold and italic runs in parsed output")
	}

	// Wrap at a narrow width to force multiple output lines
	lines := wrapRuns(paragraphRuns, 20)
	if len(lines) < 2 {
		t.Fatalf("wrapRuns produced %d lines, want >= 2 (narrow width should force wrapping)", len(lines))
	}

	// Verify every run in every wrapped line preserves its style from the original
	// by checking that each run's style is one of the known input styles.
	originalStyles := map[mdRunStyle]bool{}
	for _, r := range paragraphRuns {
		originalStyles[r.style] = true
	}

	for i, line := range lines {
		for j, r := range line {
			if !originalStyles[r.style] {
				t.Errorf("line %d, run %d: style=%v not present in original paragraph styles", i, j, r.style)
			}
			if r.style == runNormal && r.text == "" {
				continue // empty text runs are OK
			}
		}
	}
}

// TestIntegration_TableLayout verifies that layoutTable returns column widths
// for a 3-column table where the total (sum of widths + borders) fits within
// the available width.
//
//	Spec: "layoutTable at width 40: column widths sum + borders ≤ 40"
func TestIntegration_TableLayout(t *testing.T) {
	src := "| H1 | H2 | H3 |\n| -- | -- | -- |\n| c1 | c2 | c3 |"
	blocks := parseMarkdown(src)
	if len(blocks) != 1 || blocks[0].kind != blockTable {
		t.Fatalf("got %d blocks (kind=%v), want 1 blockTable", len(blocks), blocks[0].kind)
	}

	table := blocks[0]
	if len(table.headers) != 3 {
		t.Fatalf("got %d header cells, want 3", len(table.headers))
	}
	if len(table.rows) != 1 || len(table.rows[0]) != 3 {
		t.Fatalf("got %d rows / %d cols, want 1 row with 3 cols", len(table.rows), len(table.rows[0]))
	}

	colWidths := layoutTable(table, 40)
	if colWidths == nil {
		t.Fatal("layoutTable returned nil")
	}
	if len(colWidths) != 3 {
		t.Fatalf("got %d column widths, want 3", len(colWidths))
	}

	// Sum of column widths + borders (4 for 3 columns: | + | + | + |)
	total := 4 // numCols + 1 borders
	for _, w := range colWidths {
		total += w
	}
	if total > 40 {
		t.Errorf("total table width = %d (col widths %v + 4 borders), exceeds availWidth 40", total, colWidths)
	}
}

// TestIntegration_FullDocumentHeight verifies renderedHeight for a document
// with multiple block types, and that narrow width produces greater height
// than wide width due to text wrapping.
//
//	Spec: "renderedHeight > 0 and height at narrow(20) > height at wide(80)"
func TestIntegration_FullDocumentHeight(t *testing.T) {
	src := "# Document Title\n\n" +
		"This is a paragraph with text that should wrap at narrow widths.\n\n" +
		"```\ncode line one\ncode line two\ncode line three\n```\n\n" +
		"- Bullet item one with enough text to wrap\n" +
		"- Bullet item two also with wrapping text"

	blocks := parseMarkdown(src)
	if len(blocks) < 2 {
		t.Fatalf("got %d blocks, want at least 2", len(blocks))
	}

	rNarrow := mdRenderer{
		blocks:   blocks,
		width:    20,
		wrapText: true,
	}
	rWide := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: true,
	}

	hNarrow := rNarrow.renderedHeight()
	hWide := rWide.renderedHeight()

	if hNarrow <= 0 {
		t.Errorf("narrow height = %d, want > 0", hNarrow)
	}
	if hWide <= 0 {
		t.Errorf("wide height = %d, want > 0", hWide)
	}
	if hNarrow <= hWide {
		t.Errorf("narrow height (%d) <= wide height (%d), want narrow > wide due to wrapping", hNarrow, hWide)
	}
}
