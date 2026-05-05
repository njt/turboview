package tv

import (
	"testing"
)

// =============================================================================
// Compile-time type checks
// =============================================================================

// compile-time check: mdBlockKind is an int-based enum.
// Spec: "mdBlockKind enum with 10 values"
var _ = blockParagraph + blockHeader + blockCodeBlock + blockBulletList + blockNumberList + blockBlockquote + blockTable + blockHRule + blockDefList + blockCheckList

// compile-time check: mdRunStyle is an int-based enum.
// Spec: "mdRunStyle enum with 7 values"
var _ = runNormal + runBold + runItalic + runBoldItalic + runCode + runLink + runStrikethrough

// =============================================================================
// mdBlockKind enum iota tests
// =============================================================================

// TestMdBlockKindIotaValues verifies mdBlockKind values follow iota ordering (0-9).
// Spec: "mdBlockKind enum with 10 values: blockParagraph, blockHeader, blockCodeBlock,
//
//	blockBulletList, blockNumberList, blockBlockquote, blockTable, blockHRule,
//	blockDefList, blockCheckList"
func TestMdBlockKindIotaValues(t *testing.T) {
	if blockParagraph != 0 {
		t.Errorf("blockParagraph = %d, want 0 (first iota)", blockParagraph)
	}
	if blockHeader != 1 {
		t.Errorf("blockHeader = %d, want 1", blockHeader)
	}
	if blockCodeBlock != 2 {
		t.Errorf("blockCodeBlock = %d, want 2", blockCodeBlock)
	}
	if blockBulletList != 3 {
		t.Errorf("blockBulletList = %d, want 3", blockBulletList)
	}
	if blockNumberList != 4 {
		t.Errorf("blockNumberList = %d, want 4", blockNumberList)
	}
	if blockBlockquote != 5 {
		t.Errorf("blockBlockquote = %d, want 5", blockBlockquote)
	}
	if blockTable != 6 {
		t.Errorf("blockTable = %d, want 6", blockTable)
	}
	if blockHRule != 7 {
		t.Errorf("blockHRule = %d, want 7", blockHRule)
	}
	if blockDefList != 8 {
		t.Errorf("blockDefList = %d, want 8", blockDefList)
	}
	if blockCheckList != 9 {
		t.Errorf("blockCheckList = %d, want 9 (last iota)", blockCheckList)
	}
}

// TestMdBlockKindDistinctValues verifies all mdBlockKind values are distinct.
// This catches bugs where iota is accidentally interrupted or values are duplicated.
// Spec: "mdBlockKind enum with 10 values" — each must be unique.
func TestMdBlockKindDistinctValues(t *testing.T) {
	seen := map[mdBlockKind]bool{}
	kinds := []mdBlockKind{
		blockParagraph, blockHeader, blockCodeBlock, blockBulletList,
		blockNumberList, blockBlockquote, blockTable, blockHRule,
		blockDefList, blockCheckList,
	}
	for _, k := range kinds {
		if seen[k] {
			t.Errorf("mdBlockKind value %d appears more than once", k)
		}
		seen[k] = true
	}
	if len(seen) != 10 {
		t.Errorf("expected 10 distinct mdBlockKind values, got %d", len(seen))
	}
}

// =============================================================================
// mdRunStyle enum iota tests
// =============================================================================

// TestMdRunStyleIotaValues verifies mdRunStyle values follow iota ordering (0-6).
// Spec: "mdRunStyle enum with 7 values: runNormal, runBold, runItalic,
//
//	runBoldItalic, runCode, runLink, runStrikethrough"
func TestMdRunStyleIotaValues(t *testing.T) {
	if runNormal != 0 {
		t.Errorf("runNormal = %d, want 0 (first iota)", runNormal)
	}
	if runBold != 1 {
		t.Errorf("runBold = %d, want 1", runBold)
	}
	if runItalic != 2 {
		t.Errorf("runItalic = %d, want 2", runItalic)
	}
	if runBoldItalic != 3 {
		t.Errorf("runBoldItalic = %d, want 3", runBoldItalic)
	}
	if runCode != 4 {
		t.Errorf("runCode = %d, want 4", runCode)
	}
	if runLink != 5 {
		t.Errorf("runLink = %d, want 5", runLink)
	}
	if runStrikethrough != 6 {
		t.Errorf("runStrikethrough = %d, want 6 (last iota)", runStrikethrough)
	}
}

// TestMdRunStyleDistinctValues verifies all mdRunStyle values are distinct.
// Spec: "mdRunStyle enum with 7 values" — each must be unique.
func TestMdRunStyleDistinctValues(t *testing.T) {
	seen := map[mdRunStyle]bool{}
	styles := []mdRunStyle{
		runNormal, runBold, runItalic, runBoldItalic,
		runCode, runLink, runStrikethrough,
	}
	for _, s := range styles {
		if seen[s] {
			t.Errorf("mdRunStyle value %d appears more than once", s)
		}
		seen[s] = true
	}
	if len(seen) != 7 {
		t.Errorf("expected 7 distinct mdRunStyle values, got %d", len(seen))
	}
}

// =============================================================================
// mdBlock struct field tests
// =============================================================================

// TestMdBlockFieldsExist verifies mdBlock has all required fields with correct types.
// Spec: "mdBlock struct with fields: kind, level, runs, language, code, items,
//
//	children, headers, rows"
func TestMdBlockFieldsExist(t *testing.T) {
	b := mdBlock{}

	// Each field should be assignable from a literal of the expected type.
	b.kind = blockParagraph
	b.level = 3
	b.runs = []mdRun{{text: "hello", style: runNormal}}
	b.language = "go"
	b.code = []string{"line1", "line2"}
	b.items = []mdItem{{runs: []mdRun{{text: "item", style: runNormal}}}}
	b.children = []mdBlock{{kind: blockParagraph}}
	b.headers = [][]mdRun{{{text: "col1", style: runBold}}}
	b.rows = [][][]mdRun{{{{text: "cell", style: runNormal}}}}

	// Verify assignments stuck (not silently ignored)
	if b.kind != blockParagraph {
		t.Error("kind field not retained")
	}
	if b.level != 3 {
		t.Error("level field not retained")
	}
	if len(b.runs) != 1 || b.runs[0].text != "hello" {
		t.Error("runs field not retained")
	}
	if b.language != "go" {
		t.Error("language field not retained")
	}
	if len(b.code) != 2 || b.code[1] != "line2" {
		t.Error("code field not retained")
	}
	if len(b.items) != 1 {
		t.Error("items field not retained")
	}
	if len(b.children) != 1 {
		t.Error("children field not retained")
	}
	if len(b.headers) != 1 || len(b.headers[0]) != 1 {
		t.Error("headers field not retained")
	}
	if len(b.rows) != 1 || len(b.rows[0]) != 1 || len(b.rows[0][0]) != 1 {
		t.Error("rows field not retained")
	}
}

// =============================================================================
// mdRun struct field tests
// =============================================================================

// TestMdRunFieldsExist verifies mdRun has all required fields with correct types.
// Spec: "mdRun struct with fields: text, style, url"
func TestMdRunFieldsExist(t *testing.T) {
	r := mdRun{}

	r.text = "hello world"
	r.style = runBold
	r.url = "https://example.com"

	if r.text != "hello world" {
		t.Error("text field not retained")
	}
	if r.style != runBold {
		t.Error("style field not retained")
	}
	if r.url != "https://example.com" {
		t.Error("url field not retained")
	}
}

// =============================================================================
// mdItem struct field tests
// =============================================================================

// TestMdItemFieldsExist verifies mdItem has all required fields with correct types.
// Spec: "mdItem struct with fields: runs, children, checked, term"
func TestMdItemFieldsExist(t *testing.T) {
	checked := true
	item := mdItem{}

	item.runs = []mdRun{{text: "item text", style: runNormal}}
	item.children = []mdBlock{{kind: blockParagraph}}
	item.checked = &checked
	item.term = []mdRun{{text: "term", style: runBold}}

	if len(item.runs) != 1 || item.runs[0].text != "item text" {
		t.Error("runs field not retained")
	}
	if len(item.children) != 1 {
		t.Error("children field not retained")
	}
	if item.checked == nil || *item.checked != true {
		t.Error("checked field not retained")
	}
	if len(item.term) != 1 || item.term[0].text != "term" {
		t.Error("term field not retained")
	}
}

// TestMdItemCheckedNilForRegularItems verifies checked is *bool, allowing nil.
// Spec: "mdItem struct with fields: ... checked"
// The checked field is *bool (pointer to bool) so it can be nil for regular items
// vs non-nil for checkbox items.
func TestMdItemCheckedNilForRegularItems(t *testing.T) {
	item := mdItem{}
	if item.checked != nil {
		t.Error("zero-value mdItem.checked should be nil (*bool zero value)")
	}

	// Verify we can set it to a pointer
	val := false
	item.checked = &val
	if item.checked == nil {
		t.Error("mdItem.checked should accept a *bool pointer")
	}
	if *item.checked != false {
		t.Errorf("*item.checked = %v, want false", *item.checked)
	}
}

// =============================================================================
// parseMarkdown — empty and whitespace input
// =============================================================================

// TestParseMarkdownEmptyString verifies empty input produces empty block slice.
// Spec: "parseMarkdown(src string) []mdBlock" — empty string should return nil
// or empty slice, not panic or return garbage.
func TestParseMarkdownEmptyString(t *testing.T) {
	blocks := parseMarkdown("")
	if len(blocks) != 0 {
		t.Errorf("parseMarkdown(\"\") returned %d blocks, want 0", len(blocks))
	}
}

// TestParseMarkdownWhitespaceOnly verifies whitespace-only input is handled.
// Spec: "Paragraphs: collects inline children into []mdRun" — whitespace-only
// text may produce an empty paragraph or no blocks; either is acceptable, but
// the function must not panic.
func TestParseMarkdownWhitespaceOnly(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("parseMarkdown whitespace-only input panicked: %v", r)
		}
	}()
	blocks := parseMarkdown("   \n\n   \n")
	// Whitespace-only should produce 0 or more blocks without panicking.
	// If it produces blocks, they should not contain garbage.
	for _, b := range blocks {
		if b.kind < blockParagraph || b.kind > blockCheckList {
			t.Errorf("unexpected mdBlockKind %d from whitespace input", b.kind)
		}
	}
	// Just exercising the function with whitespace — the main assertion is no panic.
	t.Logf("whitespace-only produced %d blocks", len(blocks))
}

// =============================================================================
// parseMarkdown — paragraph tests
// =============================================================================

// TestParseMarkdownSimpleParagraph verifies a plain text line produces a paragraph block.
// Spec: "Paragraphs: collects inline children into []mdRun"
func TestParseMarkdownSimpleParagraph(t *testing.T) {
	blocks := parseMarkdown("hello world")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockParagraph {
		t.Errorf("kind = %d, want blockParagraph (%d)", blocks[0].kind, blockParagraph)
	}
	if len(blocks[0].runs) != 1 {
		t.Fatalf("got %d runs, want 1", len(blocks[0].runs))
	}
	if blocks[0].runs[0].text != "hello world" {
		t.Errorf("text = %q, want %q", blocks[0].runs[0].text, "hello world")
	}
	if blocks[0].runs[0].style != runNormal {
		t.Errorf("style = %d, want runNormal (%d)", blocks[0].runs[0].style, runNormal)
	}
}

// TestParseMarkdownMultipleParagraphs verifies multiple paragraphs produce multiple blocks.
// Falsifying test: catches an implementation that only handles the first paragraph
// and ignores subsequent ones.
// Spec: "Paragraphs: collects inline children into []mdRun"
func TestParseMarkdownMultipleParagraphs(t *testing.T) {
	blocks := parseMarkdown("first paragraph\n\nsecond paragraph")
	if len(blocks) < 2 {
		t.Fatalf("got %d blocks, want at least 2 for two paragraphs", len(blocks))
	}
	if blocks[0].kind != blockParagraph {
		t.Errorf("block 0 kind = %d, want blockParagraph", blocks[0].kind)
	}
	if blocks[1].kind != blockParagraph {
		t.Errorf("block 1 kind = %d, want blockParagraph", blocks[1].kind)
	}
}

// =============================================================================
// parseMarkdown — bold text tests
// =============================================================================

// TestParseMarkdownBoldText verifies **text** produces runBold style.
// Spec: "Bold text → runBold"
func TestParseMarkdownBoldText(t *testing.T) {
	blocks := parseMarkdown("**bold**")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if len(blocks[0].runs) != 1 {
		t.Fatalf("got %d runs, want 1", len(blocks[0].runs))
	}
	if blocks[0].runs[0].text != "bold" {
		t.Errorf("text = %q, want %q", blocks[0].runs[0].text, "bold")
	}
	if blocks[0].runs[0].style != runBold {
		t.Errorf("style = %d, want runBold (%d)", blocks[0].runs[0].style, runBold)
	}
}

// TestParseMarkdownBoldWithNormalText verifies bold runs are interleaved with normal runs.
// Falsifying test: catches an implementation that ignores the normal text surrounding bold.
// Spec: "Bold text → runBold"
func TestParseMarkdownBoldWithNormalText(t *testing.T) {
	blocks := parseMarkdown("normal **bold** text")
	if len(blocks) != 1 || len(blocks[0].runs) < 2 {
		t.Fatalf("expected at least 2 runs (normal + bold), got %v", len(blocks[0].runs))
	}
	hasBold := false
	hasNormal := false
	for _, r := range blocks[0].runs {
		if r.style == runBold && r.text == "bold" {
			hasBold = true
		}
		if r.style == runNormal {
			hasNormal = true
		}
	}
	if !hasBold {
		t.Error("bold text not found in runs")
	}
	if !hasNormal {
		t.Error("normal text not found in runs")
	}
}

// =============================================================================
// parseMarkdown — italic text tests
// =============================================================================

// TestParseMarkdownItalicText verifies *text* produces runItalic style.
// Spec: "italic → runItalic"
func TestParseMarkdownItalicText(t *testing.T) {
	blocks := parseMarkdown("*italic*")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if len(blocks[0].runs) != 1 {
		t.Fatalf("got %d runs, want 1", len(blocks[0].runs))
	}
	if blocks[0].runs[0].text != "italic" {
		t.Errorf("text = %q, want %q", blocks[0].runs[0].text, "italic")
	}
	if blocks[0].runs[0].style != runItalic {
		t.Errorf("style = %d, want runItalic (%d)", blocks[0].runs[0].style, runItalic)
	}
}

// TestParseMarkdownItalicNotConfusedWithBold verifies single-asterisk italic is
// distinct from double-asterisk bold.
// Falsifying test: catches an implementation that treats all emphasis as bold.
// Spec: "italic → runItalic" (must be distinct from runBold)
func TestParseMarkdownItalicNotConfusedWithBold(t *testing.T) {
	blocks := parseMarkdown("*italic* **bold**")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	hasItalic := false
	hasBold := false
	for _, r := range blocks[0].runs {
		if r.style == runItalic && r.text == "italic" {
			hasItalic = true
		}
		if r.style == runBold && r.text == "bold" {
			hasBold = true
		}
	}
	if !hasItalic {
		t.Error("italic text not found with runItalic style")
	}
	if !hasBold {
		t.Error("bold text not found with runBold style")
	}
}

// =============================================================================
// parseMarkdown — bold italic tests
// =============================================================================

// TestParseMarkdownBoldItalicText verifies ***text*** produces runBoldItalic.
// Spec: "both → runBoldItalic"
func TestParseMarkdownBoldItalicText(t *testing.T) {
	blocks := parseMarkdown("***bold italic***")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	hasBoldItalic := false
	for _, r := range blocks[0].runs {
		if r.style == runBoldItalic {
			hasBoldItalic = true
			break
		}
	}
	if !hasBoldItalic {
		t.Error("***bold italic*** did not produce runBoldItalic; got styles:")
		for _, r := range blocks[0].runs {
			t.Logf("  text=%q style=%d", r.text, r.style)
		}
	}
}

// TestParseMarkdownBoldItalicAlternateSyntax verifies ___text___ also produces runBoldItalic.
// Falsifying test: catches an implementation that only handles asterisk delimiters.
// Spec: "both → runBoldItalic"
func TestParseMarkdownBoldItalicAlternateSyntax(t *testing.T) {
	blocks := parseMarkdown("___bold italic___")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	hasBoldItalic := false
	for _, r := range blocks[0].runs {
		if r.style == runBoldItalic {
			hasBoldItalic = true
			break
		}
	}
	if !hasBoldItalic {
		t.Error("___bold italic___ did not produce runBoldItalic")
	}
}

// =============================================================================
// parseMarkdown — inline code tests
// =============================================================================

// TestParseMarkdownInlineCode verifies `code` produces runCode style.
// Spec: "Inline code → runCode"
func TestParseMarkdownInlineCode(t *testing.T) {
	blocks := parseMarkdown("`code`")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if len(blocks[0].runs) != 1 {
		t.Fatalf("got %d runs, want 1", len(blocks[0].runs))
	}
	if blocks[0].runs[0].text != "code" {
		t.Errorf("text = %q, want %q", blocks[0].runs[0].text, "code")
	}
	if blocks[0].runs[0].style != runCode {
		t.Errorf("style = %d, want runCode (%d)", blocks[0].runs[0].style, runCode)
	}
}

// TestParseMarkdownInlineCodeInSentence verifies inline code within a sentence.
// Falsifying test: catches an implementation that drops surrounding normal text.
// Spec: "Inline code → runCode"
func TestParseMarkdownInlineCodeInSentence(t *testing.T) {
	blocks := parseMarkdown("use `fmt.Println()` to log")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	hasCode := false
	hasNormal := false
	for _, r := range blocks[0].runs {
		if r.style == runCode {
			hasCode = true
		}
		if r.style == runNormal && len(r.text) > 0 {
			hasNormal = true
		}
	}
	if !hasCode {
		t.Error("inline code not found")
	}
	if !hasNormal {
		t.Error("surrounding normal text not preserved")
	}
}

// =============================================================================
// parseMarkdown — link tests
// =============================================================================

// TestParseMarkdownLink verifies [text](url) produces runLink with URL.
// Spec: "Links → runLink with URL in url field, display text in text"
func TestParseMarkdownLink(t *testing.T) {
	blocks := parseMarkdown("[click here](https://example.com)")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	var found bool
	for _, r := range blocks[0].runs {
		if r.style == runLink {
			if r.text != "click here" {
				t.Errorf("link text = %q, want %q", r.text, "click here")
			}
			if r.url != "https://example.com" {
				t.Errorf("link url = %q, want %q", r.url, "https://example.com")
			}
			found = true
		}
	}
	if !found {
		t.Error("runLink not found in runs")
	}
}

// TestParseMarkdownMultipleLinks verifies multiple links are all captured.
// Falsifying test: catches an implementation that only stores the last link.
// Spec: "Links → runLink with URL in url field, display text in text"
func TestParseMarkdownMultipleLinks(t *testing.T) {
	blocks := parseMarkdown("[first](http://1.com) and [second](http://2.com)")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	linkCount := 0
	for _, r := range blocks[0].runs {
		if r.style == runLink {
			linkCount++
			if r.url == "" {
				t.Errorf("link %q has empty URL", r.text)
			}
		}
	}
	if linkCount != 2 {
		t.Errorf("found %d links, want 2", linkCount)
	}
}

// =============================================================================
// parseMarkdown — image tests
// =============================================================================

// TestParseMarkdownImageAltText verifies ![alt](url) produces runCode with [IMG: alt].
// Spec: "Images → runCode with text [IMG: alt]"
func TestParseMarkdownImageAltText(t *testing.T) {
	blocks := parseMarkdown("![photo](image.jpg)")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	var found bool
	for _, r := range blocks[0].runs {
		if r.style == runCode && len(r.text) > 0 && r.text[:5] == "[IMG:" {
			if r.text != "[IMG: photo]" {
				t.Errorf("image run text = %q, want %q", r.text, "[IMG: photo]")
			}
			found = true
		}
	}
	if !found {
		t.Error("[IMG: ...] code run not found for image")
	}
}

// TestParseMarkdownImageWithoutAltText verifies images without alt text get default label.
// Falsifying test: catches an implementation that produces empty image label.
// Spec: "Images → runCode with text [IMG: alt]" — when alt is empty, should still
// produce a meaningful label.
func TestParseMarkdownImageWithoutAltText(t *testing.T) {
	blocks := parseMarkdown("![](image.jpg)")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	var found bool
	for _, r := range blocks[0].runs {
		if r.style == runCode && len(r.text) >= 5 && r.text[:5] == "[IMG:" {
			if r.text == "[IMG: ]" {
				t.Error("image alt text is empty, expected a default label like 'image'")
			}
			found = true
		}
	}
	if !found {
		t.Error("[IMG: ...] code run not found for image without alt text")
	}
}

// =============================================================================
// parseMarkdown — strikethrough tests
// =============================================================================

// TestParseMarkdownStrikethrough verifies ~~text~~ produces runStrikethrough.
// Spec: "Strikethrough → runStrikethrough"
func TestParseMarkdownStrikethrough(t *testing.T) {
	blocks := parseMarkdown("~~struck~~")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	var found bool
	for _, r := range blocks[0].runs {
		if r.style == runStrikethrough && r.text == "struck" {
			found = true
		}
	}
	if !found {
		t.Error("runStrikethrough not found; got styles:")
		for _, r := range blocks[0].runs {
			t.Logf("  text=%q style=%d", r.text, r.style)
		}
	}
}

// TestParseMarkdownStrikethroughNotBold verifies ~~text~~ is not confused with bold.
// Falsifying test: catches an implementation that treats strikethrough as bold.
// Spec: "Strikethrough → runStrikethrough" (distinct from runBold)
func TestParseMarkdownStrikethroughNotBold(t *testing.T) {
	blocks := parseMarkdown("**bold** ~~strike~~")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	hasBold := false
	hasStrike := false
	for _, r := range blocks[0].runs {
		if r.style == runBold {
			hasBold = true
		}
		if r.style == runStrikethrough {
			hasStrike = true
		}
	}
	if !hasBold {
		t.Error("bold text not found")
	}
	if !hasStrike {
		t.Error("strikethrough text not found — may be confused with bold")
	}
}

// =============================================================================
// parseMarkdown — header tests
// =============================================================================

// TestParseMarkdownHeaderLevel1 verifies # text produces blockHeader with level 1.
// Spec: "Headers: sets level (1-6) and collects inline runs"
func TestParseMarkdownHeaderLevel1(t *testing.T) {
	blocks := parseMarkdown("# Title")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockHeader {
		t.Errorf("kind = %d, want blockHeader (%d)", blocks[0].kind, blockHeader)
	}
	if blocks[0].level != 1 {
		t.Errorf("level = %d, want 1", blocks[0].level)
	}
	if len(blocks[0].runs) == 0 || blocks[0].runs[0].text != "Title" {
		t.Errorf("header text = %q, want %q", blocks[0].runs[0].text, "Title")
	}
}

// TestParseMarkdownHeaderLevel3 verifies ### text produces level 3.
// Falsifying test: catches an implementation that hardcodes level to 1.
// Spec: "Headers: sets level (1-6) and collects inline runs"
func TestParseMarkdownHeaderLevel3(t *testing.T) {
	blocks := parseMarkdown("### Heading 3")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockHeader {
		t.Errorf("kind = %d, want blockHeader", blocks[0].kind)
	}
	if blocks[0].level != 3 {
		t.Errorf("level = %d, want 3", blocks[0].level)
	}
}

// TestParseMarkdownHeaderLevel6 verifies ###### text produces level 6.
// Falsifying test: catches an implementation that caps at an arbitrary number
// below the maximum.
// Spec: "Headers: sets level (1-6)"
func TestParseMarkdownHeaderLevel6(t *testing.T) {
	blocks := parseMarkdown("###### Deep")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].level != 6 {
		t.Errorf("level = %d, want 6", blocks[0].level)
	}
}

// TestParseMarkdownHeaderWithFormatting verifies bold/italic inside headers.
// Falsifying test: catches an implementation that drops inline formatting in headers.
// Spec: "Headers: sets level (1-6) and collects inline runs"
func TestParseMarkdownHeaderWithFormatting(t *testing.T) {
	blocks := parseMarkdown("## Hello **World**")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockHeader {
		t.Errorf("kind = %d, want blockHeader", blocks[0].kind)
	}
	hasBold := false
	for _, r := range blocks[0].runs {
		if r.style == runBold && r.text == "World" {
			hasBold = true
		}
	}
	if !hasBold {
		t.Error("bold text not found inside header")
	}
}

// =============================================================================
// parseMarkdown — code block tests
// =============================================================================

// TestParseMarkdownFencedCodeBlock verifies a fenced code block stores language and lines.
// Spec: "Code blocks: stores language tag and raw lines as []string (no trailing empty line)"
func TestParseMarkdownFencedCodeBlock(t *testing.T) {
	blocks := parseMarkdown("```go\nfmt.Println(\"hi\")\n```")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockCodeBlock {
		t.Errorf("kind = %d, want blockCodeBlock (%d)", blocks[0].kind, blockCodeBlock)
	}
	if blocks[0].language != "go" {
		t.Errorf("language = %q, want %q", blocks[0].language, "go")
	}
	if len(blocks[0].code) != 1 {
		t.Fatalf("got %d code lines, want 1", len(blocks[0].code))
	}
	if blocks[0].code[0] != "fmt.Println(\"hi\")" {
		t.Errorf("code[0] = %q, want %q", blocks[0].code[0], "fmt.Println(\"hi\")")
	}
}

// TestParseMarkdownFencedCodeBlockWithoutLanguage verifies code blocks without language tag.
// Falsifying test: catches an implementation that panics or errors when language is absent.
// Spec: "Code blocks: stores language tag and raw lines as []string"
func TestParseMarkdownFencedCodeBlockWithoutLanguage(t *testing.T) {
	blocks := parseMarkdown("```\ncode here\nmore code\n```")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockCodeBlock {
		t.Errorf("kind = %d, want blockCodeBlock", blocks[0].kind)
	}
	if blocks[0].language != "" {
		t.Errorf("language = %q, want empty", blocks[0].language)
	}
	if len(blocks[0].code) != 2 {
		t.Fatalf("got %d code lines, want 2", len(blocks[0].code))
	}
	if blocks[0].code[0] != "code here" {
		t.Errorf("code[0] = %q, want %q", blocks[0].code[0], "code here")
	}
	if blocks[0].code[1] != "more code" {
		t.Errorf("code[1] = %q, want %q", blocks[0].code[1], "more code")
	}
}

// TestParseMarkdownCodeBlockNoTrailingEmptyLine verifies code block lines exclude
// trailing newline artifacts.
// Spec: "raw lines as []string (no trailing empty line)"
func TestParseMarkdownCodeBlockNoTrailingEmptyLine(t *testing.T) {
	blocks := parseMarkdown("```\nline1\nline2\n```")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if len(blocks[0].code) != 2 {
		t.Fatalf("got %d code lines, want 2, lines: %q", len(blocks[0].code), blocks[0].code)
	}
	for i, line := range blocks[0].code {
		if len(line) > 0 && line[len(line)-1] == '\n' {
			t.Errorf("code line %d has trailing newline: %q", i, line)
		}
	}
}

// TestParseMarkdownIndentedCodeBlock verifies indented code blocks are recognized.
// Falsifying test: catches an implementation that only handles fenced code blocks.
// Spec: "Code blocks: stores language tag and raw lines as []string"
// Indented code blocks (4 spaces) are standard CommonMark.
func TestParseMarkdownIndentedCodeBlock(t *testing.T) {
	blocks := parseMarkdown("    indented code\n    line 2")
	if len(blocks) == 0 {
		t.Fatal("no blocks produced from indented code block")
	}
	b := blocks[0]
	if b.kind != blockCodeBlock {
		t.Errorf("kind = %d, want blockCodeBlock (%d)", b.kind, blockCodeBlock)
	}
	if len(b.code) != 2 {
		t.Errorf("got %d code lines, want 2", len(b.code))
	}
}

// =============================================================================
// parseMarkdown — bullet list tests
// =============================================================================

// TestParseMarkdownBulletList verifies - items produce blockBulletList.
// Spec: "Bullet lists: each item gets runs from inline content, children from nested blocks"
func TestParseMarkdownBulletList(t *testing.T) {
	blocks := parseMarkdown("- item one\n- item two")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockBulletList {
		t.Errorf("kind = %d, want blockBulletList (%d)", blocks[0].kind, blockBulletList)
	}
	if len(blocks[0].items) != 2 {
		t.Fatalf("got %d items, want 2", len(blocks[0].items))
	}
	if len(blocks[0].items[0].runs) == 0 {
		t.Error("item 0 has no runs")
	}
}

// TestParseMarkdownBulletListStarSyntax verifies * items also produce bullet lists.
// Falsifying test: catches an implementation that only recognizes - as list marker.
// Spec: "Bullet lists: each item gets runs from inline content"
func TestParseMarkdownBulletListStarSyntax(t *testing.T) {
	blocks := parseMarkdown("* star item")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockBulletList {
		t.Errorf("kind = %d, want blockBulletList (%d), got %d", blocks[0].kind, blockBulletList, blocks[0].kind)
	}
}

// TestParseMarkdownEmptyBulletList verifies a single-item list works.
// Falsifying test: catches an implementation that requires multiple items.
// Spec: "Bullet lists: each item gets runs from inline content"
func TestParseMarkdownEmptyBulletList(t *testing.T) {
	blocks := parseMarkdown("- single")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if len(blocks[0].items) != 1 {
		t.Errorf("got %d items, want 1", len(blocks[0].items))
	}
}

// =============================================================================
// parseMarkdown — numbered list tests
// =============================================================================

// TestParseMarkdownNumberedList verifies 1. items produce blockNumberList with level.
// Spec: "Ordered lists: same as bullet, plus level field stores start number"
func TestParseMarkdownNumberedList(t *testing.T) {
	blocks := parseMarkdown("1. first\n2. second")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockNumberList {
		t.Errorf("kind = %d, want blockNumberList (%d)", blocks[0].kind, blockNumberList)
	}
	if len(blocks[0].items) != 2 {
		t.Fatalf("got %d items, want 2", len(blocks[0].items))
	}
}

// TestParseMarkdownNumberedListStart verifies the level field stores the start number.
// Falsifying test: catches an implementation that hardcodes start to 1.
// Spec: "plus level field stores start number"
func TestParseMarkdownNumberedListStart(t *testing.T) {
	blocks := parseMarkdown("3. first\n4. second")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockNumberList {
		t.Errorf("kind = %d, want blockNumberList (%d)", blocks[0].kind, blockNumberList)
	}
	if blocks[0].level != 3 {
		t.Errorf("level = %d, want 3 (start number)", blocks[0].level)
	}
}

// =============================================================================
// parseMarkdown — nested list tests
// =============================================================================

// TestParseMarkdownNestedList verifies sub-items appear in item.children.
// Spec: "Bullet lists: each item gets ... children from nested blocks"
func TestParseMarkdownNestedList(t *testing.T) {
	blocks := parseMarkdown("- parent\n  - child")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if len(blocks[0].items) != 1 {
		t.Fatalf("got %d items, want 1", len(blocks[0].items))
	}
	parent := blocks[0].items[0]
	if len(parent.children) == 0 {
		t.Fatal("parent item has no children — nested list not captured")
	}
	subList := parent.children[0]
	if subList.kind != blockBulletList {
		t.Errorf("child kind = %d, want blockBulletList", subList.kind)
	}
	if len(subList.items) != 1 {
		t.Errorf("child list has %d items, want 1", len(subList.items))
	}
}

// TestParseMarkdownDeeplyNestedList verifies triple-nested lists work.
// Falsifying test: catches an implementation that only handles one level of nesting.
// Spec: "Bullet lists: each item gets ... children from nested blocks"
func TestParseMarkdownDeeplyNestedList(t *testing.T) {
	blocks := parseMarkdown("- level 1\n  - level 2\n    - level 3")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	// Walk to level 3
	if len(blocks[0].items) != 1 {
		t.Fatal("level 1 item missing")
	}
	item1 := blocks[0].items[0]
	if len(item1.children) == 0 {
		t.Fatal("level 2 list missing")
	}
	if len(item1.children[0].items) != 1 {
		t.Fatal("level 2 item missing")
	}
	item2 := item1.children[0].items[0]
	if len(item2.children) == 0 {
		t.Fatal("level 3 list missing — deep nesting not supported")
	}
}

// =============================================================================
// parseMarkdown — blockquote tests
// =============================================================================

// TestParseMarkdownBlockquote verifies > text produces blockBlockquote.
// Spec: "Blockquotes: children contains the recursive block contents"
func TestParseMarkdownBlockquote(t *testing.T) {
	blocks := parseMarkdown("> quoted text")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockBlockquote {
		t.Errorf("kind = %d, want blockBlockquote (%d)", blocks[0].kind, blockBlockquote)
	}
	if len(blocks[0].children) == 0 {
		t.Fatal("blockquote has no children")
	}
}

// TestParseMarkdownNestedBlockquote verifies nested > > text works.
// Falsifying test: catches an implementation that doesn't recurse into blockquotes.
// Spec: "Blockquotes: children contains the recursive block contents"
func TestParseMarkdownNestedBlockquote(t *testing.T) {
	blocks := parseMarkdown("> outer\n> > inner")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockBlockquote {
		t.Errorf("kind = %d, want blockBlockquote", blocks[0].kind)
	}
	// At least one child block in the outer blockquote
	if len(blocks[0].children) == 0 {
		t.Fatal("outer blockquote has no children")
	}
	// Check if any child is a nested blockquote
	hasNestedBlockquote := false
	for _, child := range blocks[0].children {
		if child.kind == blockBlockquote {
			hasNestedBlockquote = true
			break
		}
	}
	if !hasNestedBlockquote {
		t.Error("nested blockquote not found in children")
	}
}

// =============================================================================
// parseMarkdown — horizontal rule tests
// =============================================================================

// TestParseMarkdownHorizontalRule verifies --- produces blockHRule.
// Spec: "Horizontal rules: empty block with kind = blockHRule"
func TestParseMarkdownHorizontalRule(t *testing.T) {
	blocks := parseMarkdown("---")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockHRule {
		t.Errorf("kind = %d, want blockHRule (%d)", blocks[0].kind, blockHRule)
	}
}

// TestParseMarkdownHorizontalRuleVariants verifies *** and ___ also produce horizontal rules.
// Falsifying test: catches an implementation that only handles dashes.
// Spec: "Horizontal rules: empty block with kind = blockHRule"
func TestParseMarkdownHorizontalRuleVariants(t *testing.T) {
	for _, input := range []string{"***", "___"} {
		blocks := parseMarkdown(input)
		if len(blocks) != 1 {
			t.Errorf("%q: got %d blocks, want 1", input, len(blocks))
			continue
		}
		if blocks[0].kind != blockHRule {
			t.Errorf("%q: kind = %d, want blockHRule (%d)", input, blocks[0].kind, blockHRule)
		}
	}
}

// =============================================================================
// parseMarkdown — table tests
// =============================================================================

// TestParseMarkdownTable verifies GFM table produces blockTable with headers and rows.
// Spec: "Tables: headers is [][]mdRun, rows is [][][]mdRun"
func TestParseMarkdownTable(t *testing.T) {
	blocks := parseMarkdown("| A | B |\n|---|---|\n| 1 | 2 |")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockTable {
		t.Errorf("kind = %d, want blockTable (%d)", blocks[0].kind, blockTable)
	}
	if len(blocks[0].headers) == 0 {
		t.Error("table has no headers")
	}
	if len(blocks[0].rows) == 0 {
		t.Error("table has no rows")
	}
}

// TestParseMarkdownTableHeaderContent verifies header cells contain run content.
// Spec: "Tables: headers is [][]mdRun"
func TestParseMarkdownTableHeaderContent(t *testing.T) {
	blocks := parseMarkdown("| Name | Value |\n|------|-------|\n| foo  | 42    |")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if len(blocks[0].headers) != 2 {
		t.Fatalf("got %d header cells, want 2", len(blocks[0].headers))
	}
	// Each header cell should be []mdRun with actual text
	for i, cell := range blocks[0].headers {
		if len(cell) == 0 {
			t.Errorf("header cell %d is empty", i)
		}
	}
}

// TestParseMarkdownTableMultipleRows verifies multiple table rows are preserved.
// Falsifying test: catches an implementation that only captures the first row.
// Spec: "Tables: rows is [][][]mdRun"
func TestParseMarkdownTableMultipleRows(t *testing.T) {
	blocks := parseMarkdown("| Col |\n|-----|\n| a |\n| b |\n| c |")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if len(blocks[0].rows) != 3 {
		t.Errorf("got %d rows, want 3", len(blocks[0].rows))
	}
}

// =============================================================================
// parseMarkdown — definition list tests
// =============================================================================

// TestParseMarkdownDefinitionList verifies term/definition pairs produce blockDefList.
// Spec: "Definition lists: items with term for the term, runs for the definition"
func TestParseMarkdownDefinitionList(t *testing.T) {
	blocks := parseMarkdown("Term\n: Definition text")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockDefList {
		t.Errorf("kind = %d, want blockDefList (%d)", blocks[0].kind, blockDefList)
	}
	if len(blocks[0].items) != 1 {
		t.Fatalf("got %d items, want 1", len(blocks[0].items))
	}
	item := blocks[0].items[0]
	if len(item.term) == 0 {
		t.Error("definition term has no runs")
	}
	if len(item.runs) == 0 {
		t.Error("definition body has no runs")
	}
}

// TestParseMarkdownDefinitionListMultipleTerms verifies multiple term/definition pairs.
// Falsifying test: catches an implementation that only handles one definition.
// Spec: "Definition lists: items with term for the term, runs for the definition"
func TestParseMarkdownDefinitionListMultipleTerms(t *testing.T) {
	blocks := parseMarkdown("First\n: Definition one\n\nSecond\n: Definition two")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if len(blocks[0].items) < 2 {
		t.Errorf("got %d items, want at least 2", len(blocks[0].items))
	}
	for i, item := range blocks[0].items {
		if len(item.term) == 0 {
			t.Errorf("item %d has no term", i)
		}
	}
}

// =============================================================================
// parseMarkdown — task/checkbox list tests
// =============================================================================

// TestParseMarkdownUncheckedTaskItem verifies - [ ] produces checked=false.
// Spec: "Task list items: checked is *bool (non-nil)"
func TestParseMarkdownUncheckedTaskItem(t *testing.T) {
	blocks := parseMarkdown("- [ ] pending task")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if blocks[0].kind != blockCheckList {
		t.Errorf("kind = %d, want blockCheckList (%d)", blocks[0].kind, blockCheckList)
	}
	if len(blocks[0].items) != 1 {
		t.Fatalf("got %d items, want 1", len(blocks[0].items))
	}
	item := blocks[0].items[0]
	if item.checked == nil {
		t.Fatal("checked is nil for task item — must be non-nil *bool")
	}
	if *item.checked != false {
		t.Errorf("checked = %v, want false", *item.checked)
	}
}

// TestParseMarkdownCheckedTaskItem verifies - [x] produces checked=true.
// Spec: "Task list items: checked is *bool (non-nil)"
func TestParseMarkdownCheckedTaskItem(t *testing.T) {
	blocks := parseMarkdown("- [x] done task")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	if len(blocks[0].items) != 1 {
		t.Fatalf("got %d items, want 1", len(blocks[0].items))
	}
	item := blocks[0].items[0]
	if item.checked == nil {
		t.Fatal("checked is nil for checked task item")
	}
	if *item.checked != true {
		t.Errorf("checked = %v, want true", *item.checked)
	}
}

// TestParseMarkdownRegularListItemCheckedNil verifies regular (non-task) items have nil checked.
// Falsifying test: catches an implementation that sets checked on all items.
// Spec: "Task list items: checked is *bool (non-nil)"
// Regular list items should have checked == nil.
func TestParseMarkdownRegularListItemCheckedNil(t *testing.T) {
	blocks := parseMarkdown("- regular item")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	// Regular list items are not in a check list, so checked should be nil
	if blocks[0].kind == blockCheckList {
		// If it was incorrectly classified as a check list, the test should still
		// verify the item's checked field behavior.
		t.Log("note: regular item classified as blockCheckList")
	}
	for _, item := range blocks[0].items {
		if item.checked != nil {
			t.Errorf("regular list item has non-nil checked = %v", *item.checked)
		}
	}
}

// =============================================================================
// parseMarkdown — adjacent run merging tests
// =============================================================================

// TestParseMarkdownAdjacentRunsMerged verifies adjacent runs with same style are merged.
// Spec: "Adjacent runs with the same style are merged into a single run"
func TestParseMarkdownAdjacentRunsMerged(t *testing.T) {
	// This tests that when the AST walker produces multiple runs of the same style
	// (e.g., from adjacent text segments), they are merged.
	// Soft line breaks in paragraphs are a good test: goldmark produces separate
	// Text nodes for each line, with SoftLineBreak between them.
	blocks := parseMarkdown("line one\nline two\n")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	// Both lines should be merged into one runNormal (with a space replacing the soft break)
	// or at most one runNormal if the merge worked.
	normalCount := 0
	for _, r := range blocks[0].runs {
		if r.style == runNormal {
			normalCount++
		}
	}
	// After merging, there should be at most 1 runNormal (merged together)
	if normalCount > 1 {
		t.Errorf("got %d separate runNormal runs — adjacent same-style runs were not merged", normalCount)
	}
}

// TestParseMarkdownAdjacentRunsDifferentStylesNotMerged verifies runs with different
// styles are NOT merged.
// Spec: "Adjacent runs with the same style are merged into a single run"
// This means runs with DIFFERENT styles must remain separate.
func TestParseMarkdownAdjacentRunsDifferentStylesNotMerged(t *testing.T) {
	blocks := parseMarkdown("**bold** *italic*")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	styleCount := 0
	lastStyle := mdRunStyle(-1)
	for _, r := range blocks[0].runs {
		if r.style != lastStyle {
			styleCount++
			lastStyle = r.style
		}
	}
	if styleCount < 2 {
		t.Error("different styles were incorrectly merged")
	}
}

// =============================================================================
// parseMarkdown — nested emphasis tests
// =============================================================================

// TestParseMarkdownBoldInsideItalic verifies bold inside italic produces runBoldItalic.
// Spec: "Nested emphasis is resolved: bold inside italic = runBoldItalic"
func TestParseMarkdownBoldInsideItalic(t *testing.T) {
	blocks := parseMarkdown("*italic **bold inside** more*")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	hasBoldItalic := false
	for _, r := range blocks[0].runs {
		if r.style == runBoldItalic {
			hasBoldItalic = true
			break
		}
	}
	if !hasBoldItalic {
		t.Error("bold inside italic did not produce runBoldItalic; got styles:")
		for _, r := range blocks[0].runs {
			t.Logf("  text=%q style=%d", r.text, r.style)
		}
	}
}

// TestParseMarkdownItalicInsideBold verifies italic inside bold produces runBoldItalic.
// Spec: "Nested emphasis is resolved: bold inside italic = runBoldItalic"
// The spec explicitly mentions bold inside italic, but the resolver should handle
// both nesting orders.
func TestParseMarkdownItalicInsideBold(t *testing.T) {
	blocks := parseMarkdown("**bold *italic inside* more**")
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	hasBoldItalic := false
	for _, r := range blocks[0].runs {
		if r.style == runBoldItalic {
			hasBoldItalic = true
			break
		}
	}
	if !hasBoldItalic {
		t.Error("italic inside bold did not produce runBoldItalic; got styles:")
		for _, r := range blocks[0].runs {
			t.Logf("  text=%q style=%d", r.text, r.style)
		}
	}
}

// =============================================================================
// parseMarkdown — goldmark extension integration
// =============================================================================

// TestParseMarkdownGfmExtensionsEnabled verifies GFM features work: table, strikethrough, task list.
// Falsifying test: catches an implementation that doesn't enable GFM extensions.
// Spec: "Parses src via goldmark with GFM extensions (table, strikethrough, task list)"
func TestParseMarkdownGfmExtensionsEnabled(t *testing.T) {
	// All three GFM features in one document
	input := "# Test\n\n~~strike~~\n\n- [x] done\n\n| A |\n|---|\n| 1 |\n"
	blocks := parseMarkdown(input)

	hasHeader := false
	hasStrike := false
	hasTask := false
	hasTable := false

	for _, b := range blocks {
		switch b.kind {
		case blockHeader:
			hasHeader = true
		case blockCheckList:
			hasTask = true
		case blockTable:
			hasTable = true
		case blockParagraph:
			for _, r := range b.runs {
				if r.style == runStrikethrough {
					hasStrike = true
				}
			}
		}
	}

	if !hasStrike {
		t.Error("strikethrough not found — GFM extension may not be enabled")
	}
	if !hasTask {
		t.Error("task list not found — GFM extension may not be enabled")
	}
	if !hasTable {
		t.Error("table not found — GFM extension may not be enabled")
	}
	if !hasHeader {
		t.Log("header not found (may be expected with this input)")
	}
}

// TestParseMarkdownDefinitionListExtensionEnabled verifies definition list extension works.
// Falsifying test: catches an implementation that omits the definition list extension.
// Spec: "definition list extension"
func TestParseMarkdownDefinitionListExtensionEnabled(t *testing.T) {
	blocks := parseMarkdown("Term\n: Definition")
	if len(blocks) == 0 {
		t.Fatal("no blocks produced — definition list extension may not be enabled")
	}
	if blocks[0].kind != blockDefList {
		t.Errorf("kind = %d, want blockDefList (%d) — definition list extension may not be enabled",
			blocks[0].kind, blockDefList)
	}
}

// =============================================================================
// parseMarkdown — comprehensive document test
// =============================================================================

// TestParseMarkdownFullDocument verifies a document with multiple block types
// produces the expected count and kind sequence.
// Spec: "Walks the goldmark AST to produce []mdBlock"
func TestParseMarkdownFullDocument(t *testing.T) {
	markdown := `# Title

This is a paragraph with **bold** and *italic* text.

` + "```go" + `
func main() {}
` + "```" + `

- list item
- [x] task

> quoted

---

| X | Y |
|---|---|
| a | b |
`
	blocks := parseMarkdown(markdown)
	if len(blocks) == 0 {
		t.Fatal("no blocks produced from full document")
	}

	// Verify we have a variety of block kinds
	kindsFound := map[mdBlockKind]bool{}
	for _, b := range blocks {
		kindsFound[b.kind] = true
	}

	expectedKinds := []mdBlockKind{
		blockHeader, blockParagraph, blockCodeBlock,
		blockBulletList, blockCheckList, blockBlockquote,
		blockHRule, blockTable,
	}
	for _, kind := range expectedKinds {
		if !kindsFound[kind] {
			t.Errorf("expected block kind %d not found in output", kind)
		}
	}

	t.Logf("full document produced %d blocks with %d distinct kinds", len(blocks), len(kindsFound))
}
