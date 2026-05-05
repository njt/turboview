package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// =============================================================================
// Helpers
// =============================================================================

// mkRun builds a simple mdRun for test data.
func mkRun(text string, style mdRunStyle) mdRun {
	return mdRun{text: text, style: style}
}

// testMarkdownCS returns a ColorScheme with distinct styles for each markdown role,
// so that composeStyle tests can tell them apart.
func testMarkdownCS() *theme.ColorScheme {
	return &theme.ColorScheme{
		MarkdownNormal:    tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack),
		MarkdownBold:      tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
		MarkdownItalic:    tcell.StyleDefault.Foreground(tcell.ColorGreen).Italic(true),
		MarkdownBoldItalic: tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true).Italic(true),
		MarkdownCode:      tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorGray),
		MarkdownLink:      tcell.StyleDefault.Foreground(tcell.ColorTeal).Underline(true),
	}
}

// =============================================================================
// wrapRuns tests
// =============================================================================

// TestWrapRuns_EmptyInput verifies empty input returns a single empty line.
// Spec: "An empty input returns a single empty line"
func TestWrapRuns_EmptyInput(t *testing.T) {
	result := wrapRuns(nil, 80)
	if len(result) != 1 {
		t.Fatalf("got %d lines, want 1", len(result))
	}
	if result[0] != nil {
		t.Fatalf("got non-nil slice %v, want nil", result[0])
	}
}

// TestWrapRuns_ZeroMaxWidth verifies maxWidth of zero returns a single empty line.
// Spec: "maxWidth <= 0 returns a single empty line"
func TestWrapRuns_ZeroMaxWidth(t *testing.T) {
	runs := []mdRun{mkRun("hello world", runNormal)}
	result := wrapRuns(runs, 0)
	if len(result) != 1 {
		t.Fatalf("got %d lines, want 1", len(result))
	}
	if result[0] != nil {
		t.Fatalf("got non-nil slice, want nil")
	}
}

// TestWrapRuns_NegativeMaxWidth verifies negative maxWidth returns a single empty line.
// Spec: "maxWidth <= 0 returns a single empty line"
func TestWrapRuns_NegativeMaxWidth(t *testing.T) {
	runs := []mdRun{mkRun("hello world", runNormal)}
	result := wrapRuns(runs, -5)
	if len(result) != 1 {
		t.Fatalf("got %d lines, want 1", len(result))
	}
	if result[0] != nil {
		t.Fatalf("got non-nil slice, want nil")
	}
}

// TestWrapRuns_ShortTextFitsOnOneLine verifies text within maxWidth is not wrapped.
// Spec: "wraps at word boundaries (spaces)" — when text fits, no wrap needed.
func TestWrapRuns_ShortTextFitsOnOneLine(t *testing.T) {
	runs := []mdRun{mkRun("hi", runNormal)}
	result := wrapRuns(runs, 10)
	if len(result) != 1 {
		t.Fatalf("got %d lines, want 1", len(result))
	}
	// All runs should be on the single line, preserving style and order.
	if len(result[0]) != 1 {
		t.Fatalf("got %d runs on line, want 1", len(result[0]))
	}
	if result[0][0].text != "hi" {
		t.Errorf("text = %q, want %q", result[0][0].text, "hi")
	}
	if result[0][0].style != runNormal {
		t.Errorf("style = %d, want runNormal (%d)", result[0][0].style, runNormal)
	}
}

// TestWrapRuns_WordBoundary verifies text wraps at a space between words.
// Spec: "Wraps at word boundaries (spaces)."
func TestWrapRuns_WordBoundary(t *testing.T) {
	// "hello world" = 11 chars, maxWidth=6 forces a wrap after "hello "
	runs := []mdRun{mkRun("hello world", runNormal)}
	result := wrapRuns(runs, 6)
	if len(result) != 2 {
		t.Fatalf("got %d lines, want 2", len(result))
	}
	// First line: "hello " (includes the space)
	if len(result[0]) != 1 || result[0][0].text != "hello " {
		t.Errorf("line 0: got runs %v, want [\"hello \"]", runTexts(result[0]))
	}
	// Second line: "world"
	if len(result[1]) != 1 || result[1][0].text != "world" {
		t.Errorf("line 1: got runs %v, want [\"world\"]", runTexts(result[1]))
	}
}

// TestWrapRuns_StylePreservedAcrossWrappedLines verifies bold style is preserved
// when a bold word wraps to the next line.
// Spec: "Preserves run styles across line breaks — a bold word that wraps keeps its bold style"
func TestWrapRuns_StylePreservedAcrossWrappedLines(t *testing.T) {
	runs := []mdRun{mkRun("boldtext more", runBold)}
	result := wrapRuns(runs, 6)
	// "boldtext" = 8 chars > 6, hard-breaks at 6: "boldte" + "xt more"
	// "xt " takes the space, "more" wraps
	// Line 0: "boldte" or "boldte" (hard break), line 1: "xt more"
	// Both lines must have runBold style on ALL runs.
	for lineIdx, line := range result {
		for runIdx, r := range line {
			if r.style != runBold {
				t.Errorf("line %d, run %d: style = %d, want runBold (%d)",
					lineIdx, runIdx, r.style, runBold)
			}
		}
	}
	if len(result) < 2 {
		t.Fatalf("got %d lines, want at least 2", len(result))
	}
}

// TestWrapRuns_SingleLongWordBrokenAtWidth verifies a word longer than maxWidth
// is broken at the width boundary (not dropped or truncated silently).
// Spec: "If a single word exceeds the width, it is broken at the width boundary"
func TestWrapRuns_SingleLongWordBrokenAtWidth(t *testing.T) {
	runs := []mdRun{mkRun("ABCDEFGHIJ", runNormal)}
	result := wrapRuns(runs, 4)
	// 10-char word, maxWidth=4 -> should produce at least 3 lines
	if len(result) < 2 {
		t.Fatalf("single long word with maxWidth=4: got %d lines, want at least 2", len(result))
	}
	// Verify no line exceeds maxWidth in character count
	for i, line := range result {
		lineLen := lineRuneLen(line)
		if lineLen > 4 {
			t.Errorf("line %d has %d chars, exceeds maxWidth=4", i, lineLen)
		}
	}
}

// TestWrapRuns_MixedStylesPreservedOnWrap verifies that when mixed-style runs
// are wrapped, styles are preserved across line breaks.
// Spec: "Preserves run styles across line breaks"
func TestWrapRuns_MixedStylesPreservedOnWrap(t *testing.T) {
	// "aaaaa" (5) + "bbbbbbbb" (8) = 13 chars, maxWidth=8
	// Col 0-4: runNormal "aaaaa" (5 chars)
	// Col 5+: runBold "bbbbbbbb" (8 chars) — first 3 chars fit on line 0, rest wraps
	runs := []mdRun{mkRun("aaaaa", runNormal), mkRun("bbbbbbbb", runBold)}
	result := wrapRuns(runs, 8)

	if len(result) < 2 {
		t.Fatalf("got %d lines, want at least 2 (text should wrap)", len(result))
	}

	// First line should end with runBold text (bold word started mid-line)
	firstLine := result[0]
	lastRun := firstLine[len(firstLine)-1]
	if lastRun.style != runBold {
		t.Errorf("first line ends with style=%d, want runBold (%d)", lastRun.style, runBold)
	}

	// Second line should start with runBold text (style preserved across break)
	secondLine := result[1]
	firstRun := secondLine[0]
	if firstRun.style != runBold {
		t.Errorf("second line starts with style=%d, want runBold (%d)", firstRun.style, runBold)
	}
}

// =============================================================================
// composeStyle tests
// =============================================================================

// TestComposeStyle_RunNormal verifies runNormal returns the block style unchanged.
// Spec: "For runNormal: returns the block style unchanged"
func TestComposeStyle_RunNormal(t *testing.T) {
	blockStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	cs := testMarkdownCS()
	result := composeStyle(blockStyle, runNormal, cs)

	fg, bg, attrs := result.Decompose()
	wantFG, wantBG, _ := blockStyle.Decompose()

	if fg != wantFG {
		t.Errorf("foreground = %v, want %v", fg, wantFG)
	}
	if bg != wantBG {
		t.Errorf("background = %v, want %v", bg, wantBG)
	}
	if attrs&tcell.AttrBold != 0 {
		t.Error("runNormal should not add Bold")
	}
	if attrs&tcell.AttrItalic != 0 {
		t.Error("runNormal should not add Italic")
	}
	if attrs&tcell.AttrUnderline != 0 {
		t.Error("runNormal should not add Underline")
	}
}

// TestComposeStyle_RunBold verifies runBold uses MarkdownBold fg+attrs, block bg.
// Spec: "For runBold ... takes foreground and attributes from the corresponding
// ColorScheme field, keeps background from the block style"
func TestComposeStyle_RunBold(t *testing.T) {
	blockStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	cs := testMarkdownCS()
	result := composeStyle(blockStyle, runBold, cs)

	// Foreground from MarkdownBold (Yellow)
	fg, _, _ := result.Decompose()
	wantFG, _, _ := cs.MarkdownBold.Decompose()
	if fg != wantFG {
		t.Errorf("foreground = %v, want %v (from MarkdownBold)", fg, wantFG)
	}
	// Background from block style (Black)
	_, bg, _ := result.Decompose()
	_, wantBG, _ := blockStyle.Decompose()
	if bg != wantBG {
		t.Errorf("background = %v, want %v (from block style)", bg, wantBG)
	}
	// Bold from MarkdownBold
	_, _, attrs := result.Decompose()
	if attrs&tcell.AttrBold == 0 {
		t.Error("runBold should have Bold=true")
	}
}

// TestComposeStyle_RunItalic verifies runItalic uses MarkdownItalic fg+attrs, block bg.
// Spec: "For ... runItalic ... takes foreground and attributes from the corresponding
// ColorScheme field, keeps background from the block style"
func TestComposeStyle_RunItalic(t *testing.T) {
	blockStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	cs := testMarkdownCS()
	result := composeStyle(blockStyle, runItalic, cs)

	fg, _, _ := result.Decompose()
	wantFG, _, _ := cs.MarkdownItalic.Decompose()
	if fg != wantFG {
		t.Errorf("foreground = %v, want %v (from MarkdownItalic)", fg, wantFG)
	}
	_, bg, _ := result.Decompose()
	_, wantBG, _ := blockStyle.Decompose()
	if bg != wantBG {
		t.Errorf("background = %v, want %v (from block style)", bg, wantBG)
	}
	_, _, attrs := result.Decompose()
	if attrs&tcell.AttrItalic == 0 {
		t.Error("runItalic should have Italic=true")
	}
}

// TestComposeStyle_RunBoldItalic verifies runBoldItalic uses MarkdownBoldItalic fg+attrs, block bg.
// Spec: "For ... runBoldItalic ... takes foreground and attributes from the corresponding
// ColorScheme field, keeps background from the block style"
func TestComposeStyle_RunBoldItalic(t *testing.T) {
	blockStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	cs := testMarkdownCS()
	result := composeStyle(blockStyle, runBoldItalic, cs)

	fg, _, _ := result.Decompose()
	wantFG, _, _ := cs.MarkdownBoldItalic.Decompose()
	if fg != wantFG {
		t.Errorf("foreground = %v, want %v (from MarkdownBoldItalic)", fg, wantFG)
	}
	_, bg, _ := result.Decompose()
	_, wantBG, _ := blockStyle.Decompose()
	if bg != wantBG {
		t.Errorf("background = %v, want %v (from block style)", bg, wantBG)
	}
	_, _, attrs := result.Decompose()
	if attrs&tcell.AttrBold == 0 {
		t.Error("runBoldItalic should have Bold=true")
	}
	if attrs&tcell.AttrItalic == 0 {
		t.Error("runBoldItalic should have Italic=true")
	}
}

// TestComposeStyle_RunCode verifies runCode returns MarkdownCode directly (own background).
// Spec: "For runCode: returns MarkdownCode directly (own background)"
func TestComposeStyle_RunCode(t *testing.T) {
	blockStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	cs := testMarkdownCS()
	result := composeStyle(blockStyle, runCode, cs)

	// Should match MarkdownCode entirely (foreground AND background)
	fg, bg, _ := result.Decompose()
	wantFG, wantBG, _ := cs.MarkdownCode.Decompose()

	if fg != wantFG {
		t.Errorf("foreground = %v, want %v (from MarkdownCode)", fg, wantFG)
	}
	if bg != wantBG {
		t.Errorf("background = %v, want %v (from MarkdownCode, NOT block style)", bg, wantBG)
	}
}

// TestComposeStyle_RunLink verifies runLink uses MarkdownLink fg+attrs (underline), block bg.
// Spec: "For runLink: takes foreground and underline from MarkdownLink, keeps background
// from block style"
func TestComposeStyle_RunLink(t *testing.T) {
	blockStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	cs := testMarkdownCS()
	result := composeStyle(blockStyle, runLink, cs)

	fg, _, _ := result.Decompose()
	wantFG, _, _ := cs.MarkdownLink.Decompose()
	if fg != wantFG {
		t.Errorf("foreground = %v, want %v (from MarkdownLink)", fg, wantFG)
	}
	_, bg, _ := result.Decompose()
	_, wantBG, _ := blockStyle.Decompose()
	if bg != wantBG {
		t.Errorf("background = %v, want %v (from block style)", bg, wantBG)
	}
	_, _, attrs := result.Decompose()
	if attrs&tcell.AttrUnderline == 0 {
		t.Error("runLink should have Underline=true")
	}
}

// TestComposeStyle_RunStrikethrough verifies runStrikethrough uses MarkdownBold fg+attrs
// plus StrikeThrough, and block bg.
// Implementation note: uses MarkdownBold fg+attrs since no dedicated MarkdownStrikethrough
// field exists in the ColorScheme. The StrikeThrough attribute is added separately.
func TestComposeStyle_RunStrikethrough(t *testing.T) {
	blockStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	cs := testMarkdownCS()
	result := composeStyle(blockStyle, runStrikethrough, cs)

	// Foreground from MarkdownBold (Yellow)
	fg, _, _ := result.Decompose()
	wantFG, _, _ := cs.MarkdownBold.Decompose()
	if fg != wantFG {
		t.Errorf("foreground = %v, want %v (from MarkdownBold)", fg, wantFG)
	}
	// Background from block style
	_, bg, _ := result.Decompose()
	_, wantBG, _ := blockStyle.Decompose()
	if bg != wantBG {
		t.Errorf("background = %v, want %v (from block style)", bg, wantBG)
	}
	// Must have StrikeThrough
	_, _, attrs := result.Decompose()
	if attrs&tcell.AttrStrikeThrough == 0 {
		t.Error("runStrikethrough should have StrikeThrough=true")
	}
}

// TestComposeStyle_NilColorScheme verifies a nil ColorScheme returns the block style
// unchanged for all run styles.
// Spec: "When cs is nil, returns the block style unchanged for all run styles"
func TestComposeStyle_NilColorScheme(t *testing.T) {
	blockStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	allRunStyles := []mdRunStyle{
		runNormal, runBold, runItalic, runBoldItalic,
		runCode, runLink, runStrikethrough,
	}

	for _, rs := range allRunStyles {
		result := composeStyle(blockStyle, rs, nil)
		// Compare via Decompose for complete style equality check
		rf, rb, ra := result.Decompose()
		bf, bb, ba := blockStyle.Decompose()
		if rf != bf || rb != bb || ra != ba {
			t.Errorf("runStyle=%d: style differs from block style when cs is nil", rs)
		}
	}
}

// TestComposeStyle_DefaultBlockStyle verifies that when blockStyle is StyleDefault
// (no explicit colors set), composeStyle properly applies the run style's foreground
// from the ColorScheme without introducing a spurious non-default background.
// Spec: "takes foreground and attributes from the corresponding ColorScheme field,
// keeps background from the block style"
func TestComposeStyle_DefaultBlockStyle(t *testing.T) {
	blockStyle := tcell.StyleDefault
	cs := testMarkdownCS()
	result := composeStyle(blockStyle, runBold, cs)

	// Foreground should come from MarkdownBold
	fg, _, _ := result.Decompose()
	wantFG, _, _ := cs.MarkdownBold.Decompose()
	if fg != wantFG {
		t.Errorf("foreground = %v, want %v (from MarkdownBold)", fg, wantFG)
	}

	// Background should be tcell.ColorDefault (preserved from StyleDefault)
	_, bg, _ := result.Decompose()
	if bg != tcell.ColorDefault {
		t.Errorf("background = %v, want tcell.ColorDefault (preserved from StyleDefault)", bg)
	}

	// Should have Bold from MarkdownBold
	_, _, attrs := result.Decompose()
	if attrs&tcell.AttrBold == 0 {
		t.Error("runBold should have Bold=true")
	}
}

// =============================================================================
// renderedHeight tests
// =============================================================================

// TestRenderedHeight_EmptyBlocks verifies zero blocks produce height 0.
// Spec: "Empty blocks returns 0"
func TestRenderedHeight_EmptyBlocks(t *testing.T) {
	r := mdRenderer{
		blocks:   nil,
		width:    80,
		wrapText: true,
	}
	if h := r.renderedHeight(); h != 0 {
		t.Errorf("nil blocks: height = %d, want 0", h)
	}

	r.blocks = []mdBlock{}
	if h := r.renderedHeight(); h != 0 {
		t.Errorf("empty blocks: height = %d, want 0", h)
	}
}

// TestRenderedHeight_SingleUnwrappedParagraph verifies an unwrapped paragraph has height 1.
// Spec: "A single paragraph with no wrapping returns 1"
func TestRenderedHeight_SingleUnwrappedParagraph(t *testing.T) {
	r := mdRenderer{
		blocks:   []mdBlock{parseMarkdown("hello world")[0]},
		width:    80,
		wrapText: false, // no wrapping
	}
	if h := r.renderedHeight(); h != 1 {
		t.Errorf("height = %d, want 1", h)
	}
}

// TestRenderedHeight_WrappedParagraph verifies a paragraph that word-wraps has
// height equal to the number of wrapped lines.
// Spec: "A paragraph that word-wraps to N lines returns N"
func TestRenderedHeight_WrappedParagraph(t *testing.T) {
	r := mdRenderer{
		blocks:   []mdBlock{parseMarkdown("hello world")[0]},
		width:    6,        // "hello " = 6 chars, "world" wraps → 2 lines
		wrapText: true,
	}
	h := r.renderedHeight()
	if h < 2 {
		t.Errorf("height = %d, want at least 2 (text wraps at width 6)", h)
	}
}

// TestRenderedHeight_TwoBlocksBlankLine verifies a blank line is inserted
// between consecutive blocks.
// Spec: "Two consecutive blocks: blank line between them (block1 height + 1 + block2 height)"
func TestRenderedHeight_TwoBlocksBlankLine(t *testing.T) {
	blocks := parseMarkdown("first\n\nsecond")
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	// Each paragraph = 1, blank = 1 → total = 3
	if h != 3 {
		t.Errorf("height = %d, want 3 (1 + 1 blank + 1)", h)
	}
}

// TestRenderedHeight_CodeBlockLines verifies a code block's height equals its line count.
// Spec: "Code block: height = number of code lines (or 1 if empty)"
func TestRenderedHeight_CodeBlockLines(t *testing.T) {
	blocks := parseMarkdown("```\na\nb\nc\n```")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	if h != 3 {
		t.Errorf("height = %d, want 3 (three code lines)", h)
	}
}

// TestRenderedHeight_EmptyCodeBlock verifies an empty code block still reports height 1.
// Spec: "Empty code block returns height 1"
func TestRenderedHeight_EmptyCodeBlock(t *testing.T) {
	b := mdBlock{
		kind: blockCodeBlock,
		code: nil, // no code lines
	}
	r := mdRenderer{
		blocks:   []mdBlock{b},
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	if h != 1 {
		t.Errorf("height = %d, want 1 (empty code block)", h)
	}
}

// TestRenderedHeight_Header verifies a header returns height 1 when not wrapping.
// Spec: "Header: height = 1 (with wrap) or number of wrapped lines"
func TestRenderedHeight_Header(t *testing.T) {
	blocks := parseMarkdown("# Title")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	if h != 1 {
		t.Errorf("height = %d, want 1 (unwrapped header)", h)
	}
}

// TestRenderedHeight_HeaderWrapped verifies a header that wraps produces >1 height.
// Spec: "Header: height = 1 (with wrap) or number of wrapped lines"
func TestRenderedHeight_HeaderWrapped(t *testing.T) {
	blocks := parseMarkdown("# A very long header title that will wrap")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    10,
		wrapText: true,
	}
	h := r.renderedHeight()
	if h <= 1 {
		t.Errorf("height = %d, want > 1 (header text should wrap at width 10)", h)
	}
}

// TestRenderedHeight_HRule verifies a horizontal rule has height 1.
// Spec: "HRule: height = 1"
func TestRenderedHeight_HRule(t *testing.T) {
	blocks := parseMarkdown("---")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	if h := r.renderedHeight(); h != 1 {
		t.Errorf("height = %d, want 1", h)
	}
}

// TestRenderedHeight_BulletListItems verifies each unwrapped bullet list item
// contributes one line.
// Spec: "Bullet/numbered/check list: height includes item text (wrapped) plus nested children"
func TestRenderedHeight_BulletListItems(t *testing.T) {
	blocks := parseMarkdown("- one\n- two\n- three")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	if h != 3 {
		t.Errorf("height = %d, want 3 (three unwrapped list items)", h)
	}
}

// TestRenderedHeight_NumberedList verifies each unwrapped numbered list item
// contributes one line.
// Spec: "Bullet/numbered/check list: height includes item text plus nested children"
func TestRenderedHeight_NumberedList(t *testing.T) {
	blocks := parseMarkdown("1. first\n2. second")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	if h != 2 {
		t.Errorf("height = %d, want 2 (two unwrapped numbered list items)", h)
	}
}

// TestRenderedHeight_CheckList verifies each unwrapped checklist item contributes one line.
// Spec: "Bullet/numbered/check list: height includes item text plus nested children"
func TestRenderedHeight_CheckList(t *testing.T) {
	blocks := parseMarkdown("- [x] checked\n- [ ] unchecked")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	if h != 2 {
		t.Errorf("height = %d, want 2 (two unwrapped checklist items)", h)
	}
}

// TestRenderedHeight_WrappedListItem verifies that a list item with long text
// contributes more than one line when wrapping is enabled with a narrow width.
// Spec: "height includes item text (wrapped) plus nested children"
func TestRenderedHeight_WrappedListItem(t *testing.T) {
	blocks := parseMarkdown("- this list item text is long enough to wrap at narrow width\n- short")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    20,
		wrapText: true,
	}
	h := r.renderedHeight()
	// At width 20 with marker space (~4 chars), the long item text should wrap
	if h <= 2 {
		t.Errorf("height = %d, want > 2 (long item should wrap, short item = 1 line)", h)
	}
}

// TestRenderedHeight_NestedList verifies nested children increase total height.
// Spec: "Bullet/numbered/check list: height includes item text (wrapped) plus nested children"
func TestRenderedHeight_NestedList(t *testing.T) {
	blocks := parseMarkdown("- parent\n  - child")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	// parent = 1 line, child = 1 line → total = 2
	if h != 2 {
		t.Errorf("height = %d, want 2 (parent + nested child)", h)
	}
}

// TestRenderedHeight_Blockquote verifies a blockquote's height equals its children's height.
// Spec: "Blockquote: height = children height with +1 depth"
func TestRenderedHeight_Blockquote(t *testing.T) {
	blocks := parseMarkdown("> quoted text")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	if h != 1 {
		t.Errorf("height = %d, want 1 (single paragraph inside blockquote)", h)
	}
}

// TestRenderedHeight_BlockquoteDepthReducesWidth verifies that nesting inside a
// blockquote reduces available width, causing wrapping.
// Spec: "blockquote: height = children height with +1 depth"
// At depth 1, available width is reduced, so a paragraph that barely fits at the
// top level must wrap inside a blockquote.
func TestRenderedHeight_BlockquoteDepthReducesWidth(t *testing.T) {
	// Create a paragraph whose text exactly fills the available width at depth 0
	const w = 10
	textWidth := w // text exactly fills width at depth 0
	text := ""
	for i := 0; i < textWidth; i++ {
		text += "x"
	}
	paraBlock := mdBlock{kind: blockParagraph, runs: []mdRun{mkRun(text, runNormal)}}

	// At depth 0, the paragraph should fit in 1 line
	r0 := mdRenderer{blocks: []mdBlock{paraBlock}, width: w, wrapText: true}
	if h := r0.renderedHeight(); h != 1 {
		t.Errorf("depth 0 (width=%d, text=%d chars): height = %d, want 1", w, textWidth, h)
	}

	// Same paragraph inside a blockquote (depth +1) has reduced width → must wrap
	bq := mdBlock{kind: blockBlockquote, children: []mdBlock{paraBlock}}
	r1 := mdRenderer{blocks: []mdBlock{bq}, width: w, wrapText: true}
	if h := r1.renderedHeight(); h <= 1 {
		t.Errorf("depth 1 (blockquote): height = %d, want > 1 (indent should reduce available width)", h)
	}
}

// TestRenderedHeight_TableHeight verifies table height includes borders, header,
// separator, and data rows.
// Spec: "Table height includes top border, header row(s), separator, data rows, bottom border"
func TestRenderedHeight_TableHeight(t *testing.T) {
	blocks := parseMarkdown("| A | B |\n|---|---|\n| 1 | 2 |")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	// top border (1) + header (1) + separator (1) + data row (1) + bottom border (1) = 5
	if h != 5 {
		t.Errorf("height = %d, want 5 (top + header + sep + row + bottom)", h)
	}
}

// TestRenderedHeight_TableMultipleRows verifies multiple data rows each contribute.
// Spec: "Table height includes ... data rows"
func TestRenderedHeight_TableMultipleRows(t *testing.T) {
	blocks := parseMarkdown("| X |\n|---|\n| a |\n| b |\n| c |")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	// top (1) + header (1) + sep (1) + 3 data rows (3) + bottom (1) = 7
	if h != 7 {
		t.Errorf("height = %d, want 7 (top + header + sep + 3 rows + bottom)", h)
	}
}

// TestRenderedHeight_TableWithCellWrapping verifies that when wrapText is enabled
// with a narrow width, cells that wrap produce a taller table than the same
// table rendered without wrapping at a wide width.
// Spec: "Table height includes ... cell wrapping when wrapText is enabled"
func TestRenderedHeight_TableWithCellWrapping(t *testing.T) {
	blocks := parseMarkdown("| Col | VeryWideCellContentHere |\n|-----|------------------------|\n| a   | b                      |")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}

	rWide := mdRenderer{blocks: blocks, width: 80, wrapText: false}
	rNarrow := mdRenderer{blocks: blocks, width: 20, wrapText: true}

	wideH := rWide.renderedHeight()
	narrowH := rNarrow.renderedHeight()

	// Narrow width must produce more height due to cell wrapping
	if narrowH <= wideH {
		t.Errorf("narrow width (20, wrap) height = %d, wide width (80, nowrap) height = %d — narrow should be taller",
			narrowH, wideH)
	}
}

// TestRenderedHeight_DefinitionList verifies definition list height accounts for
// term lines, definition lines, and blank lines between items.
// Spec: "Definition list: height includes term line + wrapped definition lines +
// blank lines between items"
func TestRenderedHeight_DefinitionList(t *testing.T) {
	blocks := parseMarkdown("First\n: Definition one\n\nSecond\n: Definition two")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	r := mdRenderer{
		blocks:   blocks,
		width:    80,
		wrapText: false,
	}
	h := r.renderedHeight()
	// Item 0: term (1) + definition (1) = 2
	// Blank between items: 1
	// Item 1: term (1) + definition (1) = 2
	// Total: 5
	if h != 5 {
		t.Errorf("height = %d, want 5 (term + def + blank + term + def)", h)
	}
}

// TestRenderedHeight_WidthAffectsWrapping verifies that reducing width increases
// height because of word wrapping.
// Spec: "Width affects wrapping and thus height"
func TestRenderedHeight_WidthAffectsWrapping(t *testing.T) {
	blocks := parseMarkdown("this is a paragraph that will wrap when the width is small")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}

	rWide := mdRenderer{blocks: blocks, width: 80, wrapText: true}
	rNarrow := mdRenderer{blocks: blocks, width: 10, wrapText: true}

	wideH := rWide.renderedHeight()
	narrowH := rNarrow.renderedHeight()

	if narrowH <= wideH {
		t.Errorf("narrow width (%d) height = %d, wide width (80) height = %d — narrow should be taller",
			10, narrowH, wideH)
	}
}

// =============================================================================
// layoutTable tests
// =============================================================================

// TestLayoutTable_EmptyTable verifies an empty table (no headers, no rows) returns nil.
// Spec: "Empty table (no headers, no rows) returns nil"
func TestLayoutTable_EmptyTable(t *testing.T) {
	b := mdBlock{kind: blockTable}
	result := layoutTable(b, 80)
	if result != nil {
		t.Errorf("got %v, want nil", result)
	}
}

// TestLayoutTable_SimpleTwoColumn verifies a 2-column table that fits within width
// distributes extra space beyond minimums.
// Spec: "If minimums + borders fit: distribute extra space proportionally to content width"
func TestLayoutTable_SimpleTwoColumn(t *testing.T) {
	b := mdBlock{
		kind:    blockTable,
		headers: [][]mdRun{{{text: "ColA", style: runNormal}}, {{text: "ColumnB", style: runNormal}}},
		rows: [][][]mdRun{
			{{{text: "x", style: runNormal}}, {{text: "y", style: runNormal}}},
		},
	}
	result := layoutTable(b, 50)
	if len(result) != 2 {
		t.Fatalf("got %d columns, want 2", len(result))
	}
	// Border overhead: numCols + 1 = 3
	// Minimums: [max(4, 8)=8, max(7, 8)=8]
	// totalMin = 8 + 8 + 3 = 19 <= 50, so extra space distributed
	// sum(cols) + 3 should = 50
	colSum := result[0] + result[1]
	if colSum+3 != 50 {
		t.Errorf("column widths %v sum to %d, +3 borders = %d, want %d",
			result, colSum, colSum+3, 50)
	}
	// Each column should be at least its minimum (8)
	for i, w := range result {
		if w < 8 {
			t.Errorf("column %d: width = %d, want at least 8 (minimum floor)", i, w)
		}
	}
}

// TestLayoutTable_ProportionalDistribution verifies that columns with wider content
// receive proportionally more extra space than columns with narrow content.
// Spec: "If minimums + borders fit: distribute extra space proportionally to content width"
func TestLayoutTable_ProportionalDistribution(t *testing.T) {
	b := mdBlock{
		kind:    blockTable,
		headers: [][]mdRun{{{text: "A", style: runNormal}}, {{text: "VeryLongColumnName", style: runNormal}}},
		rows: [][][]mdRun{
			{{{text: "x", style: runNormal}}, {{text: "y", style: runNormal}}},
		},
	}
	// Sufficient width for both columns to exceed minimums
	result := layoutTable(b, 100)
	if len(result) != 2 {
		t.Fatalf("got %d columns, want 2", len(result))
	}
	// Column 1 has much longer content, so it should receive more extra space
	if result[1] <= result[0] {
		t.Errorf("column widths %v: column 1 (longer content) should be wider than column 0", result)
	}
}

// TestLayoutTable_TooWide verifies that when minimums + borders exceed available
// width, columns are set to their minimum widths.
// Spec: "If minimums + borders don't fit: use minimums (horizontal scroll handles overflow)"
func TestLayoutTable_TooWide(t *testing.T) {
	b := mdBlock{
		kind:    blockTable,
		headers: [][]mdRun{{{text: "WideColumn", style: runNormal}}, {{text: "AnotherWide", style: runNormal}}},
		rows: [][][]mdRun{
			{{{text: "data", style: runNormal}}, {{text: "more", style: runNormal}}},
		},
	}
	result := layoutTable(b, 15)
	if len(result) != 2 {
		t.Fatalf("got %d columns, want 2", len(result))
	}
	// Minimums: [max(10, 8)=10, max(11, 8)=11], border = 3, total = 24 > 15
	// So columns should be [10, 11] (minimums only)
	if result[0] != 10 {
		t.Errorf("column 0 = %d, want 10 (minimum width)", result[0])
	}
	if result[1] != 11 {
		t.Errorf("column 1 = %d, want 11 (minimum width)", result[1])
	}
}

// TestLayoutTable_SingleColumn verifies a single-column table works.
// Spec: "layoutTable ... returns column widths"
func TestLayoutTable_SingleColumn(t *testing.T) {
	b := mdBlock{
		kind:    blockTable,
		headers: [][]mdRun{{{text: "Col", style: runNormal}}},
		rows: [][][]mdRun{
			{{{text: "a", style: runNormal}}},
		},
	}
	result := layoutTable(b, 50)
	if len(result) != 1 {
		t.Fatalf("got %d columns, want 1", len(result))
	}
	// Minimum: max(3, 8) = 8, border = 2, total = 10 <= 50
	// Column should get all remaining space: 50 - 2 = 48
	if result[0] != 48 {
		t.Errorf("column 0 = %d, want 48 (full content width)", result[0])
	}
}

// TestLayoutTable_ColumnFloorEight verifies columns default to at least 8 characters
// even when content is very short.
// Spec: "Minimum column width: longest word in the column or 8 chars, whichever is greater"
func TestLayoutTable_ColumnFloorEight(t *testing.T) {
	b := mdBlock{
		kind:    blockTable,
		headers: [][]mdRun{{{text: "A", style: runNormal}}, {{text: "B", style: runNormal}}, {{text: "C", style: runNormal}}},
		rows: [][][]mdRun{
			{{{text: "x", style: runNormal}}, {{text: "y", style: runNormal}}, {{text: "z", style: runNormal}}},
		},
	}
	result := layoutTable(b, 28)
	// 3 columns, border = 4, mins = [8, 8, 8], total = 28
	// Fits exactly, columns get minimums
	if len(result) != 3 {
		t.Fatalf("got %d columns, want 3", len(result))
	}
	for i, w := range result {
		if w < 8 {
			t.Errorf("column %d: width = %d, want at least 8 (floor)", i, w)
		}
	}
}

// TestLayoutTable_ColumnLongestWord determines minimum width, not total cell length.
// Falsifying test: catches an implementation that uses total cell text length
// instead of longest word as the minimum width metric.
// Spec: "Minimum column width: longest word in the column or 8 chars, whichever is greater"
func TestLayoutTable_ColumnLongestWord(t *testing.T) {
	b := mdBlock{
		kind:    blockTable,
		headers: [][]mdRun{{{text: "short longword", style: runNormal}}},
		rows: [][][]mdRun{
			{{{text: "a", style: runNormal}}},
		},
	}
	// Total text = 15 chars, but longest word is "longword" = 8 chars
	// Minimum = max(8, 8) = 8
	result := layoutTable(b, 9)
	// 1 column, border = 2, total if using total-length minimum (15) = 17 > 9 → use minimums = [15]
	// total if using longest-word minimum (8) = 10 > 9 → use minimums = [8]
	// If implementation incorrectly uses total length, result[0] would be 15
	// If correct (longest word), result[0] = 8
	if len(result) == 0 {
		t.Fatal("got empty result")
	}
	if result[0] > 8 {
		t.Errorf("column 0 = %d, want 8 (longest word based, not total cell length)", result[0])
	}
}

// =============================================================================
// Helpers used only by tests
// =============================================================================

// runTexts returns the text values from a slice of mdRun, for error messages.
func runTexts(runs []mdRun) []string {
	texts := make([]string, len(runs))
	for i, r := range runs {
		texts[i] = r.text
	}
	return texts
}

// lineRuneLen returns the total rune count of all run texts on a line.
func lineRuneLen(line []mdRun) int {
	n := 0
	for _, r := range line {
		n += len([]rune(r.text))
	}
	return n
}
