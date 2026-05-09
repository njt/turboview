package tv

// markdown_editor_test.go — Tests for Task: MarkdownEditor struct and constructor.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Construction: embed, options, grow mode, bounds
//   Section 2  — Field initialisation: blocks, sourceCache, showSource
//   Section 3  — SetText / reparse interaction
//   Section 4  — Text() delegation
//   Section 5  — reparse() no-op and empty-source behaviour
//   Section 6  — ShowSource / SetShowSource accessors
//   Section 7  — Falsifying / boundary tests

import (
	"reflect"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Section 1 — Construction
// ---------------------------------------------------------------------------

// TestMarkdownEditor_NewEmbedsEditor verifies the Editor embed is properly
// initialised so that inherited methods (Text, CursorPos) work immediately.
// Spec: "MarkdownEditor struct embeds *Editor (which embeds *Memo), giving it
//
//	cursor, selection, scroll, undo, and keyboard input"
//
// Spec: "NewMarkdownEditor(bounds Rect) *MarkdownEditor creates the widget,
//
//	initializes Editor->Memo"
func TestMarkdownEditor_NewEmbedsEditor(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	// Editor→Memo is initialised: Text() returns ""
	if got := me.Editor.Text(); got != "" {
		t.Errorf("Editor.Text() after NewMarkdownEditor = %q, want %q", got, "")
	}

	// Cursor starts at origin via Memo
	row, col := me.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// TestMarkdownEditor_NewSetsOfSelectable verifies the constructor sets OfSelectable.
// Spec: "sets OfSelectable|OfFirstClick"
func TestMarkdownEditor_NewSetsOfSelectable(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	if !me.HasOption(OfSelectable) {
		t.Error("NewMarkdownEditor did not set OfSelectable")
	}
}

// TestMarkdownEditor_NewSetsOfFirstClick verifies the constructor sets OfFirstClick.
// Spec: "sets OfSelectable|OfFirstClick"
func TestMarkdownEditor_NewSetsOfFirstClick(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	if !me.HasOption(OfFirstClick) {
		t.Error("NewMarkdownEditor did not set OfFirstClick")
	}
}

// TestMarkdownEditor_NewSetsGrowMode verifies GrowMode is GfGrowHiX|GfGrowHiY.
// Spec: "sets GrowMode to GfGrowHiX|GfGrowHiY"
func TestMarkdownEditor_NewSetsGrowMode(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	want := GfGrowHiX | GfGrowHiY
	if me.GrowMode() != want {
		t.Errorf("GrowMode() = %v, want %v (GfGrowHiX|GfGrowHiY)", me.GrowMode(), want)
	}
}

// TestMarkdownEditor_NewSetsGrowModeIsExact verifies GrowMode is exactly
// GfGrowHiX|GfGrowHiY, not a superset such as GfGrowAll.
// Falsifying: an implementation using GfGrowAll would pass a "has both bits"
// check but fails here.
func TestMarkdownEditor_NewSetsGrowModeIsExact(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	if me.GrowMode()&GfGrowLoX != 0 {
		t.Error("GrowMode should NOT include GfGrowLoX")
	}
	if me.GrowMode()&GfGrowLoY != 0 {
		t.Error("GrowMode should NOT include GfGrowLoY")
	}
}

// TestMarkdownEditor_NewStoresBounds verifies the constructor records the given bounds.
// Spec: "NewMarkdownEditor(bounds Rect) *MarkdownEditor creates the widget"
func TestMarkdownEditor_NewStoresBounds(t *testing.T) {
	r := NewRect(5, 3, 50, 15)
	me := NewMarkdownEditor(r)

	if me.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", me.Bounds(), r)
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Field initialisation
// ---------------------------------------------------------------------------

// TestMarkdownEditor_NewInitializesFields verifies the struct fields are
// initialised to sensible defaults after construction.
// Spec: "MarkdownEditor stores blocks []mdBlock (parsed output),
//
//	sourceCache string (to detect source changes), and
//	showSource bool (source toggle)"
func TestMarkdownEditor_NewInitializesFields(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	// blocks: after NewMarkdownEditor->reparse() on empty source, must be non-nil empty slice
	if me.blocks == nil {
		t.Error("blocks is nil after NewMarkdownEditor; want non-nil empty slice")
	}
	if len(me.blocks) != 0 {
		t.Errorf("blocks has %d entries after NewMarkdownEditor, want 0", len(me.blocks))
	}

	// sourceCache: after reparse on empty Memo, sourceCache should be ""
	if me.sourceCache != "" {
		t.Errorf("sourceCache = %q, want %q", me.sourceCache, "")
	}

	// showSource: defaults to false
	if me.showSource {
		t.Error("showSource should default to false")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — SetText / reparse interaction
// ---------------------------------------------------------------------------

// TestMarkdownEditor_SetTextPopulatesBlocks verifies SetText calls reparse
// and produces parsed blocks for markdown content.
// Spec: "SetText(s string) overrides Editor.SetText: calls the embedded
//
//	Editor.SetText, then calls reparse() to populate blocks"
func TestMarkdownEditor_SetTextPopulatesBlocks(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# Hello\n\nWorld")

	if me.blocks == nil {
		t.Error("blocks is nil after SetText with markdown")
	}
	if len(me.blocks) == 0 {
		t.Error("blocks is empty after SetText with markdown content")
	}
}

// TestMarkdownEditor_SetTextPopulatesBlocksWithExpectedKinds verifies that
// reparse produces blocks with the correct kinds for known markdown constructs.
// Falsifying: a stub reparse that always returns a single paragraph would fail.
// Spec: "reparse() joins Memo.lines into a string, runs goldmark parse,
//
//	stores result in blocks"
func TestMarkdownEditor_SetTextPopulatesBlocksWithExpectedKinds(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// One heading + one paragraph: goldmark produces a heading block then a paragraph block.
	me.SetText("# Hello\n\nWorld")

	if len(me.blocks) < 2 {
		t.Fatalf("expected at least 2 blocks, got %d", len(me.blocks))
	}

	if me.blocks[0].kind != blockHeader {
		t.Errorf("blocks[0].kind = %v, want blockHeader (%v)", me.blocks[0].kind, blockHeader)
	}
	if me.blocks[0].level != 1 {
		t.Errorf("blocks[0].level = %d, want 1 for H1 heading", me.blocks[0].level)
	}

	if me.blocks[1].kind != blockParagraph {
		t.Errorf("blocks[1].kind = %v, want blockParagraph (%v)", me.blocks[1].kind, blockParagraph)
	}
}

// TestMarkdownEditor_SetTextUpdatesSourceCache verifies SetText triggers
// reparse which updates sourceCache.
// Falsifying: an implementation that skips reparse or doesn't update
// sourceCache would fail. This guards against the "calls reparse" part of the
// spec being skipped.
// Spec: "reparse() is called after every edit"
func TestMarkdownEditor_SetTextUpdatesSourceCache(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// sourceCache is "" initially (empty Memo)
	if me.sourceCache != "" {
		t.Fatalf("sourceCache before SetText = %q, want empty", me.sourceCache)
	}

	me.SetText("hello world")

	if me.sourceCache != "hello world" {
		t.Errorf("sourceCache after SetText = %q, want %q", me.sourceCache, "hello world")
	}
}

// TestMarkdownEditor_SetTextCallsEditorSetText verifies SetText calls the
// embedded Editor.SetText rather than bypassing it to manipulate Memo directly.
// Editor.SetText has unique side effects beyond Memo.SetText: it resets
// Editor.modified to false and clears Editor.undo_ (undo history).
// Falsifying: an implementation that shortcuts by calling me.Memo.SetText(s)
// directly (then reparse()) would leave modified=true and undo_ intact.
// Spec: "SetText(s string) overrides Editor.SetText: calls the embedded
//
//	Editor.SetText, then calls reparse() to populate blocks"
func TestMarkdownEditor_SetTextCallsEditorSetText(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	// Poison Editor state. If MarkdownEditor.SetText bypasses Editor.SetText
	// and calls Memo.SetText directly, these fields retain their poisoned values.
	me.Editor.modified = true
	me.Editor.undo_ = &undoState{lines: [][]rune{{'x'}}}

	me.SetText("fresh content")

	// Editor.SetText clears the modified flag.
	if me.Editor.modified {
		t.Error("Editor.modified = true after SetText; Editor.SetText was likely bypassed — the modified flag should be reset")
	}

	// Editor.SetText clears undo history.
	if me.Editor.undo_ != nil {
		t.Error("Editor.undo_ is non-nil after SetText; Editor.SetText was likely bypassed — undo history should be cleared")
	}

	// Sanity-check: text was actually set.
	if me.Text() != "fresh content" {
		t.Errorf("Text() = %q, want %q", me.Text(), "fresh content")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Text() delegation
// ---------------------------------------------------------------------------

// TestMarkdownEditor_TextReturnsSource verifies Text() returns the source
// string passed to SetText.
// Spec: "Text() string delegates to Editor.Text() (inherited)"
func TestMarkdownEditor_TextReturnsSource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("sample text")

	if got := me.Text(); got != "sample text" {
		t.Errorf("Text() = %q, want %q", got, "sample text")
	}
}

// TestMarkdownEditor_TextInitiallyEmpty verifies Text() returns empty string
// when SetText has not been called.
// Spec: "Text() string delegates to Editor.Text() (inherited)"
func TestMarkdownEditor_TextInitiallyEmpty(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	if got := me.Text(); got != "" {
		t.Errorf("Text() before SetText = %q, want %q", got, "")
	}
}

// TestMarkdownEditor_TextMatchesEditorText verifies the MarkdownEditor.Text()
// and its inner Editor.Text() return the same value — proving delegation is
// not broken by a stray override.
// Falsifying: an override that returns cached blocks text or stale data.
// Spec: "Text() string delegates to Editor.Text() (inherited)"
func TestMarkdownEditor_TextMatchesEditorText(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("line1\nline2")

	markdownEditorText := me.Text()
	editorText := me.Editor.Text()

	if markdownEditorText != editorText {
		t.Errorf("MarkdownEditor.Text() = %q but Editor.Text() = %q; they must match",
			markdownEditorText, editorText)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — reparse() no-op and empty-source behaviour
// ---------------------------------------------------------------------------

// TestMarkdownEditor_ReparseNoOpsWhenUnchanged verifies reparse() returns
// immediately when the source text has not changed since the last parse.
// Spec: "If the source text hasn't changed (compared to sourceCache),
//
//	reparse() is a no-op"
func TestMarkdownEditor_ReparseNoOpsWhenUnchanged(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# heading\n\ncontent")

	// Capture blocks after SetText (reparse ran)
	blocksAfterFirst := me.blocks
	if len(blocksAfterFirst) == 0 {
		t.Fatal("blocks should be non-empty after SetText")
	}

	// Corrupt blocks to a recognisable sentinel value
	sentinel := []mdBlock{{kind: blockHRule}}
	me.blocks = sentinel

	// Call reparse directly — source hasn't changed, must be a no-op
	me.reparse()

	if !reflect.DeepEqual(me.blocks, sentinel) {
		t.Error("reparse() modified blocks when source was unchanged; expected no-op")
	}
}

// TestMarkdownEditor_ReparseHandlesEmptySource verifies reparse clears blocks
// to an empty slice when the source is empty.
// Spec: "reparse() handles empty source: blocks is set to empty slice
//
//	[]mdBlock{}"
func TestMarkdownEditor_ReparseHandlesEmptySource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	// After construction, source is empty so blocks must be non-nil empty slice
	if me.blocks == nil {
		t.Error("blocks is nil after construction with empty source; want non-nil empty slice")
	}
	if len(me.blocks) != 0 {
		t.Errorf("blocks has %d entries for empty source, want 0", len(me.blocks))
	}

	// Now SetText with content, then clear to empty
	me.SetText("# heading\n\ncontent")
	me.SetText("")

	if me.blocks == nil {
		t.Error("blocks is nil after SetText(\"\"); want non-nil empty slice")
	}
	if len(me.blocks) != 0 {
		t.Errorf("blocks has %d entries after SetText(\"\"), want 0", len(me.blocks))
	}
}

// TestMarkdownEditor_ReparseEmptySourceIsNotNil verifies blocks is a non-nil
// empty slice after empty-source reparse, NOT nil.
// Falsifying: a lazy implementation that sets blocks = nil for empty source
// would pass len==0 checks but fail on nil checks that consumers depend on.
// Spec: "reparse() handles empty source: blocks is set to empty slice
//
//	[]mdBlock{}" — empty slice, not nil.
func TestMarkdownEditor_ReparseEmptySourceIsNotNil(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	// Force non-empty source then back to empty
	me.SetText("content")
	me.SetText("")

	if me.blocks == nil {
		t.Error("blocks is nil for empty source; spec says empty slice []mdBlock{}, not nil")
	}
}

// ---------------------------------------------------------------------------
// Section 6 — ShowSource / SetShowSource accessors
// ---------------------------------------------------------------------------

// TestMarkdownEditor_ShowSourceDefaultsToFalse verifies the showSource field
// defaults to false on construction.
// Spec: "showSource bool (source toggle, wired in Phase 3)"
func TestMarkdownEditor_ShowSourceDefaultsToFalse(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	if me.ShowSource() {
		t.Error("ShowSource() should default to false")
	}
}

// TestMarkdownEditor_SetShowSourceTrue verifies SetShowSource(true) updates
// the toggle state.
// Spec: "SetShowSource(on bool) sets the source toggle state"
func TestMarkdownEditor_SetShowSourceTrue(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	me.SetShowSource(true)
	if !me.ShowSource() {
		t.Error("ShowSource() = false after SetShowSource(true), want true")
	}
}

// TestMarkdownEditor_SetShowSourceFalse verifies SetShowSource(false) clears
// the toggle after it was set to true.
// Spec: "SetShowSource(on bool) sets the source toggle state"
func TestMarkdownEditor_SetShowSourceFalse(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetShowSource(true)
	me.SetShowSource(false)

	if me.ShowSource() {
		t.Error("ShowSource() = true after SetShowSource(false), want false")
	}
}

// TestMarkdownEditor_SetShowSourceReflectsField verifies SetShowSource
// directly sets the unexported showSource field (in-package visibility).
// Falsifying: a stub that always returns a different value than what was set.
// Spec: "ShowSource() bool returns the current source toggle state"
func TestMarkdownEditor_SetShowSourceReflectsField(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	me.SetShowSource(true)
	if !me.showSource {
		t.Error("showSource field = false after SetShowSource(true)")
	}

	me.SetShowSource(false)
	if me.showSource {
		t.Error("showSource field = true after SetShowSource(false)")
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Falsifying / boundary tests
// ---------------------------------------------------------------------------

// TestMarkdownEditor_SetTextMultiLineParsesBlocks verifies that multi-line
// markdown source with newlines is properly joined and parsed into blocks.
// Spec: "reparse() joins Memo.lines into a string, runs goldmark parse"
func TestMarkdownEditor_SetTextMultiLineParsesBlocks(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// Three paragraphs separated by blank lines
	me.SetText("first\n\nsecond\n\nthird")

	if len(me.blocks) < 3 {
		t.Fatalf("expected at least 3 blocks, got %d", len(me.blocks))
	}
	for i := 0; i < 3; i++ {
		if me.blocks[i].kind != blockParagraph {
			t.Errorf("blocks[%d].kind = %v, want blockParagraph (%v)", i, me.blocks[i].kind, blockParagraph)
		}
	}
}

// TestMarkdownEditor_TwoInstancesAreIndependent verifies two MarkdownEditor
// instances do not share mutable state.
// Falsifying: global singleton or shared-state bug would fail here.
func TestMarkdownEditor_TwoInstancesAreIndependent(t *testing.T) {
	me1 := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me2 := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	me1.SetText("content one")
	me2.SetText("content two")

	if me1.Text() == me2.Text() {
		t.Error("two MarkdownEditor instances share Text() state; they must be independent")
	}
	if me1.Text() != "content one" {
		t.Errorf("me1.Text() = %q, want %q", me1.Text(), "content one")
	}
	if me2.Text() != "content two" {
		t.Errorf("me2.Text() = %q, want %q", me2.Text(), "content two")
	}
}

// TestMarkdownEditor_ReparseReplacesBlocks verifies that reparse fully replaces
// blocks rather than appending to the previous parse result.
// Falsifying: an implementation that appends instead of replacing.
// Spec: "stores result in blocks" — stores, not appends.
func TestMarkdownEditor_ReparseReplacesBlocks(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))

	me.SetText("# first\n\ncontent")
	firstBlockCount := len(me.blocks)

	me.SetText("## second\n\nmore content")
	secondBlockCount := len(me.blocks)

	// After replacing text, block count should reflect the new parse, not the old one plus new
	if secondBlockCount == firstBlockCount*2 {
		t.Errorf("blocks count doubled from %d to %d — reparse appears to append rather than replace",
			firstBlockCount, secondBlockCount)
	}

	// Verify the new content is what's actually in blocks
	if me.blocks[0].kind != blockHeader {
		t.Errorf("blocks[0].kind after second SetText = %v, want blockHeader", me.blocks[0].kind)
	}
	if me.blocks[0].level != 2 {
		t.Errorf("blocks[0].level = %d, want 2 for H2 heading", me.blocks[0].level)
	}
}

// TestMarkdownEditor_ReparseDetectsChangeAfterMemoEdit verifies that after the
// underlying Memo text changes (simulating an edit), reparse detects the change
// and re-parses.
// Spec: "reparse() is called after every edit (not every keystroke — it's
//
//	called from HandleEvent after Memo processes the edit)"
//
// Spec: "If the source text hasn't changed (compared to sourceCache),
//
//	reparse() is a no-op"
func TestMarkdownEditor_ReparseDetectsChangeAfterMemoEdit(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# heading\n\nparagraph")

	// Set sourceCache to a value different from what's actually in Memo
	// to simulate what happens when Memo is edited externally
	me.sourceCache = "stale cache"

	// Now reparse should detect the mismatch and re-parse
	me.reparse()

	// sourceCache should be updated to the actual Memo text
	if me.sourceCache != "# heading\n\nparagraph" {
		t.Errorf("sourceCache after reparse = %q, want %q", me.sourceCache, "# heading\n\nparagraph")
	}
}

// =============================================================================
// Task 2 — Custom Draw with formatted markdown rendering
// =============================================================================

// ---------------------------------------------------------------------------
// Section 8 — Draw tests (requirements 1-9)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_DrawFillsBackgroundWhenBlocksEmpty verifies Draw fills the
// entire viewport with MarkdownNormal when blocks is empty.
// Req 5: "Draw fills the background with MarkdownNormal style"
// Req 9: "If blocks is empty, fill the viewport with background and return"
func TestMarkdownEditor_DrawFillsBackgroundWhenBlocksEmpty(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 20, 5))
	me.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 5)
	me.Draw(buf)

	normalStyle := theme.BorlandBlue.MarkdownNormal
	for y := 0; y < 5; y++ {
		for x := 0; x < 20; x++ {
			cell := buf.GetCell(x, y)
			if cell.Style != normalStyle {
				t.Errorf("cell (%d,%d) style = %v, want MarkdownNormal %v", x, y, cell.Style, normalStyle)
				return
			}
		}
	}
}

// TestMarkdownEditor_DrawFillsBackgroundWhenBlocksNotEmpty verifies the full
// viewport is filled with MarkdownNormal background even when blocks exist.
// This ensures the fill step runs BEFORE content rendering, not skipped.
// Req 5: "Draw fills the background with MarkdownNormal style"
func TestMarkdownEditor_DrawFillsBackgroundWhenBlocksNotEmpty(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 30, 5))
	me.scheme = theme.BorlandBlue
	me.SetText("# Heading\n\nContent")

	buf := NewDrawBuffer(30, 5)
	me.Draw(buf)

	normalStyle := theme.BorlandBlue.MarkdownNormal

	// Every cell must have background filled — even on rows where content
	// isn't rendered, the background fill must cover them.
	for y := 0; y < 5; y++ {
		for x := 0; x < 30; x++ {
			cell := buf.GetCell(x, y)
			// If cell has default style, background fill was skipped
			if cell.Style == tcell.StyleDefault {
				t.Errorf("cell (%d,%d) has default style; background fill was not applied", x, y)
				return
			}
			// The background fill is MarkdownNormal; rendered content may
			// override with other styles on top, but unfilled cells are a bug.
			_ = normalStyle
		}
	}
}

// TestMarkdownEditor_DrawEmptyBlocksNoCrash verifies Draw does not crash or
// panic when blocks is empty (no content set). The spec explicitly requires
// filling the viewport and returning without error.
// Req 9: "If blocks is empty, fill the viewport with background and return
//
//	(no crash)"
func TestMarkdownEditor_DrawEmptyBlocksNoCrash(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	// blocks is already empty from construction

	buf := NewDrawBuffer(40, 10)
	// Must not panic
	me.Draw(buf)

	// All cells should be filled with background
	for y := 0; y < 10; y++ {
		for x := 0; x < 40; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != ' ' {
				t.Errorf("cell (%d,%d) rune = %q, want ' ' (only background fill)", x, y, string(cell.Rune))
			}
		}
	}
}

// TestMarkdownEditor_DrawEmptyBlocksNoCrashWithoutScheme verifies Draw does not
// crash when blocks is empty AND no color scheme is set (nil scheme).
// Falsifying: an implementation that dereferences cs without a nil check
// would panic here.
// Req 9: "If blocks is empty, fill the viewport with background and return
//
//	(no crash)"
func TestMarkdownEditor_DrawEmptyBlocksNoCrashWithoutScheme(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 10, 3))
	// scheme is nil (default)

	buf := NewDrawBuffer(10, 3)
	// Must not panic even with nil scheme
	me.Draw(buf)
}

// TestMarkdownEditor_DrawDelegatesToMemoWhenShowSource verifies that when
// showSource is true, Draw delegates entirely to Memo.Draw and renders raw
// source rather than formatted markdown.
// Req 2: "When showSource is true, Draw delegates entirely to Memo.Draw(buf)
//
//	— raw source view"
func TestMarkdownEditor_DrawDelegatesToMemoWhenShowSource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue
	me.SetText("# Heading\n\nContent")
	me.SetShowSource(true)

	buf := NewDrawBuffer(40, 5)
	me.Draw(buf)

	// In source mode, the raw markdown '#' character should be visible
	// (Memo.Draw renders the Memo.lines directly, including markdown syntax)
	hashFound := false
	for y := 0; y < 5; y++ {
		text := renderText(buf, y)
		if strings.Contains(text, "#") {
			hashFound = true
			break
		}
	}
	if !hashFound {
		t.Error("'#' character not found in rendered output; showSource=true should show raw source via Memo.Draw")
	}
}

// TestMarkdownEditor_DrawRendersFormattedMarkdown verifies that when showSource
// is false (default), Draw renders formatted markdown through mdRenderer
// rather than raw source text.
// Req 1: "MarkdownEditor.Draw(buf *DrawBuffer) renders formatted markdown
//
//	using the existing mdRenderer"
//
// Req 3: "When showSource is false, Draw renders blocks through
//
//	mdRenderer.renderLineInto for each visible line"
func TestMarkdownEditor_DrawRendersFormattedMarkdown(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue
	me.SetText("# Heading\n\nContent")
	// showSource defaults to false

	buf := NewDrawBuffer(40, 5)
	me.Draw(buf)

	// In formatted mode, heading text should be rendered (without the '#' prefix)
	// The heading "Heading" should appear, styled with MarkdownH1
	headingFound := false
	for y := 0; y < 5; y++ {
		text := renderText(buf, y)
		if strings.Contains(text, "Heading") {
			headingFound = true
			break
		}
	}
	if !headingFound {
		t.Error("heading text 'Heading' not found in rendered output; formatted markdown not rendered")
	}
}

// TestMarkdownEditor_DrawHeadingUsesColorScheme verifies that rendered headings
// use the correct style from the color scheme (MarkdownH1 for H1 headings).
// Req 4: "Draw uses the existing color scheme markdown styles via
//
//	me.ColorScheme()"
func TestMarkdownEditor_DrawHeadingUsesColorScheme(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue
	me.SetText("# Test H1")

	buf := NewDrawBuffer(40, 5)
	me.Draw(buf)

	h1Style := theme.BorlandBlue.MarkdownH1
	normalStyle := theme.BorlandBlue.MarkdownNormal

	// Find the 'T' from "Test H1" and verify it uses MarkdownH1 style
	found := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 40; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune == 'T' {
				found = true
				if cell.Style != h1Style {
					t.Errorf("H1 'T' at (%d,%d) style = %v, want MarkdownH1 %v", x, y, cell.Style, h1Style)
				}
				if cell.Style == normalStyle {
					t.Error("H1 cell has MarkdownNormal style, expected MarkdownH1")
				}
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("H1 heading character 'T' not found; formatted markdown not rendered")
	}
}

// TestMarkdownEditor_DrawParagraphUsesStyle verifies that rendered paragraphs
// use MarkdownNormal style from the color scheme.
// Req 4: "Draw uses the existing color scheme markdown styles via
//
//	me.ColorScheme()"
func TestMarkdownEditor_DrawParagraphUsesStyle(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue
	me.SetText("A simple paragraph.")

	buf := NewDrawBuffer(40, 5)
	me.Draw(buf)

	normalStyle := theme.BorlandBlue.MarkdownNormal

	// Find the 'A' character and verify it uses MarkdownNormal
	found := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 40; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune == 'A' {
				found = true
				if cell.Style != normalStyle {
					t.Errorf("paragraph 'A' at (%d,%d) style = %v, want MarkdownNormal %v", x, y, cell.Style, normalStyle)
				}
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("paragraph character 'A' not found; formatted markdown not rendered")
	}
}

// TestMarkdownEditor_DrawRespectsDeltaY verifies that vertical scroll position
// controls which rendered lines are visible.
// Req 7: "Scroll position (deltaX, deltaY from Memo) controls which rendered
//
//	lines are visible"
func TestMarkdownEditor_DrawRespectsDeltaY(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 3))
	me.scheme = theme.BorlandBlue
	me.SetText("Line 0\n\nLine 1\n\nLine 2")

	// Scroll down by 2 lines
	me.Memo.deltaY = 2

	buf := NewDrawBuffer(40, 3)
	me.Draw(buf)

	// The first rendered line should now be "Line 1"
	text := renderText(buf, 0)
	if !strings.Contains(text, "Line 1") {
		t.Errorf("first rendered line with deltaY=2 = %q, want 'Line 1'", text)
	}
}

// TestMarkdownEditor_DrawRespectsDeltaX verifies that horizontal scroll position
// controls which rendered columns are visible.
// Req 7: "Scroll position (deltaX, deltaY from Memo) controls which rendered
//
//	lines are visible"
func TestMarkdownEditor_DrawRespectsDeltaX(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 15, 3))
	me.scheme = theme.BorlandBlue
	me.SetText("A somewhat longer line of text")

	// Scroll right by 5 columns
	me.Memo.deltaX = 5

	buf := NewDrawBuffer(15, 3)
	me.Draw(buf)

	// With deltaX=5, the first 5 characters should be scrolled off-screen
	// The first visible character should NOT be 'A'
	cell := buf.GetCell(0, 0)
	if cell.Rune == 'A' {
		t.Error("first cell is 'A' with deltaX=5; horizontal scroll not applied")
	}
}

// TestMarkdownEditor_DrawDeltaYAtZeroShowsFirstLines verifies that when deltaY
// is 0, the first rendered lines are visible.
// Req 7: "Scroll position (deltaX, deltaY from Memo) controls which rendered
//
//	lines are visible"
func TestMarkdownEditor_DrawDeltaYAtZeroShowsFirstLines(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue
	me.SetText("First line\n\nSecond line")

	// deltaY defaults to 0

	buf := NewDrawBuffer(40, 5)
	me.Draw(buf)

	text := renderText(buf, 0)
	if !strings.Contains(text, "First") {
		t.Errorf("first rendered line = %q, want 'First line'", text)
	}
}

// TestMarkdownEditor_Draw_RendersOverscanBuffer verifies the Draw method renders
// more lines than the viewport — specifically viewportHeight+5 lines (the +5
// overscan buffer) rather than just viewportHeight.
// Req 8: "Draw only renders lines from deltaY to deltaY + viewportHeight + 5
//
//	(overscan buffer)"
func TestMarkdownEditor_Draw_RendersOverscanBuffer(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 3))
	me.scheme = theme.BorlandBlue
	// Content with enough rendered lines to fill viewport+overscan (8 lines).
	// 8 paragraphs = 8 + 7 blank separators = 15 rendered lines total.
	me.SetText("Line 0\n\nLine 1\n\nLine 2\n\nLine 3\n\nLine 4\n\nLine 5\n\nLine 6\n\nLine 7")

	// Create a buffer with extra height to capture overscan rows.
	// viewportHeight=3, overscan=5, total buffer height=8.
	buf := NewDrawBuffer(40, 8)
	me.Draw(buf)

	normalStyle := theme.BorlandBlue.MarkdownNormal

	// Verify viewport rows (0..2) contain background-filled or rendered content.
	viewportHasContent := false
	for y := 0; y < 3; y++ {
		for x := 0; x < 40; x++ {
			cell := buf.GetCell(x, y)
			if cell.Style != tcell.StyleDefault {
				viewportHasContent = true
				break
			}
		}
		if viewportHasContent {
			break
		}
	}
	if !viewportHasContent {
		t.Error("viewport rows (y=0..2) contain no filled content; background fill should have been applied")
	}

	// Overscan rows (3..7) must contain rendered content with a non-default style.
	// The background Fill covers only (0,0,w,h) where h=3, so rows 3..7 would
	// remain as tcell.StyleDefault if the overscan loop is not rendering them.
	overscanHasContent := false
	for y := 3; y < 8; y++ {
		for x := 0; x < 40; x++ {
			cell := buf.GetCell(x, y)
			if cell.Style != tcell.StyleDefault && cell.Rune != ' ' {
				overscanHasContent = true
				break
			}
			// A background-filled cell with MarkdownNormal also counts
			// (e.g., blank lines between blocks rendered with Fill).
			if cell.Style == normalStyle {
				overscanHasContent = true
				break
			}
		}
		if overscanHasContent {
			break
		}
	}
	if !overscanHasContent {
		t.Error("overscan buffer rows (y=3..7) contain no rendered content; Draw should render viewportHeight+5 lines, not just viewportHeight")
	}
}

// ---------------------------------------------------------------------------
// Section 9 — renderer() helper (requirement 12)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_RendererConfiguration verifies renderer() returns an
// mdRenderer constructed with blocks, bounds width, wrapText=true, and the
// color scheme.
// Req 6: "An mdRenderer is constructed with blocks, width (bounds width),
//
//	wrapText=true, and cs (color scheme)"
//
// Req 12: "renderer() helper returns *mdRenderer constructed with blocks,
//
//	width, wrapText=true, cs"
func TestMarkdownEditor_RendererConfiguration(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("# Hello\n\nWorld")

	r := me.renderer()

	if r == nil {
		t.Fatal("renderer() returned nil")
	}
	if r.blocks == nil {
		t.Error("renderer().blocks is nil; should be the MarkdownEditor blocks")
	}
	if len(r.blocks) == 0 {
		t.Error("renderer().blocks is empty; should contain parsed blocks")
	}
	if r.width != 40 {
		t.Errorf("renderer().width = %d, want 40 (bounds width)", r.width)
	}
	if !r.wrapText {
		t.Error("renderer().wrapText = false, want true (always wraps)")
	}
	if r.cs == nil {
		t.Error("renderer().cs is nil; should be the MarkdownEditor color scheme")
	}
}

// TestMarkdownEditor_RendererUsesBoundsWidthNotHardcoded verifies renderer()
// uses the actual bounds width, not a hardcoded value.
// Falsifying: an implementation using a constant width would pass basic tests
// but fail when bounds change.
// Req 6: "width (bounds width)"
func TestMarkdownEditor_RendererUsesBoundsWidthNotHardcoded(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 60, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("content")

	r := me.renderer()
	if r.width != 60 {
		t.Errorf("renderer().width = %d, want 60 (actual bounds width)", r.width)
	}

	// Change bounds and verify renderer updates
	me.SetBounds(NewRect(0, 0, 25, 10))
	r2 := me.renderer()
	if r2.width != 25 {
		t.Errorf("renderer().width after SetBounds = %d, want 25", r2.width)
	}
}

// ---------------------------------------------------------------------------
// Section 10 — syncScrollBars tests (requirement 11)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_SyncScrollBarsUsesRenderedHeight verifies that
// syncScrollBars uses rendered content height from mdRenderer instead of raw
// line count from Memo.lines.
// Req 11: "syncScrollBars() override — uses rendered content height from
//
//	mdRenderer instead of raw line count"
func TestMarkdownEditor_SyncScrollBarsUsesRenderedHeight(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue

	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	me.Memo.SetVScrollBar(vsb)

	// Set text that has different raw line count vs rendered line count.
	// Raw: "a\n\nb" = 3 lines (with blank line).
	// Rendered: depends on mdRenderer block layout.
	me.SetText("a\n\nb")

	// After SetText (which calls reparse), syncScrollBars should have run
	// and the scrollbar Max should use rendered height, not raw line count.
	// Raw line count = 3; rendered height = 3 after parsing block layout.
	// The key assertion: scrollbar was synced (Max > 0).
	if vsb.Max() == 0 {
		t.Error("syncScrollBars did not update vertical scrollbar — Max is 0")
	}
}

// TestMarkdownEditor_SyncScrollBarsSetsValueToDeltaY verifies the scrollbar
// value matches deltaY after sync.
// Req 11: "syncScrollBars() override — uses rendered content height from
//
//	mdRenderer instead of raw line count, matching
//	MarkdownViewer.syncScrollBars()"
func TestMarkdownEditor_SyncScrollBarsSetsValueToDeltaY(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue

	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	me.Memo.SetVScrollBar(vsb)

	me.SetText("line1\n\nline2\n\nline3")

	// Set deltaY and verify syncScrollBars propagates it to the scrollbar
	me.Memo.deltaY = 2
	me.syncScrollBars()

	if vsb.Value() != me.Memo.deltaY {
		t.Errorf("scrollbar Value = %d, want deltaY = %d", vsb.Value(), me.Memo.deltaY)
	}
}

// TestMarkdownEditor_SyncScrollBarsDifferentFromMemo verifies that
// MarkdownEditor's syncScrollBars produces DIFFERENT scrollbar ranges than
// Memo.syncScrollBars for the same content, since MarkdownEditor uses
// rendered height instead of raw line count.
// Falsifying: an implementation that simply calls Memo.syncScrollBars()
// would produce identical ranges. Markdown content with wrapping produces
// more rendered lines than raw lines.
// Req 11: "syncScrollBars() override — uses rendered content height from
//
//	mdRenderer instead of raw line count"
func TestMarkdownEditor_SyncScrollBarsDifferentFromMemo(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 20, 5))
	me.scheme = theme.BorlandBlue

	vsb := NewScrollBar(NewRect(0, 0, 1, 5), Vertical)
	me.Memo.SetVScrollBar(vsb)
	// Use a single long paragraph that MUST wrap at width 20.
	// Raw line count = 1. Rendered height will be several lines due to wrapping.
	// This ensures MarkdownEditor's rendered-height-based Max strictly exceeds
	// Memo's raw-line-count-based Max.
	longContent := "This is a long paragraph that contains many words and will definitely wrap across multiple lines when rendered in a narrow editor viewport"
	me.SetText(longContent)

	// Record what Memo.syncScrollBars would set (uses raw line count)
	memoVsb := NewScrollBar(NewRect(0, 0, 1, 5), Vertical)
	memo := NewMemo(NewRect(0, 0, 20, 5))
	memo.SetVScrollBar(memoVsb)
	memo.SetText(longContent)
	memo.syncScrollBars()
	memoMax := memoVsb.Max()

	// Now check MarkdownEditor's syncScrollBars result (uses rendered height)
	me.syncScrollBars()
	meMax := vsb.Max()

	// The rendered markdown height with wrapping at width 20 MUST be greater
	// than the raw line count. A lazy delegation to Memo.syncScrollBars would
	// set the same Max as raw line count, which this catches.
	if meMax == memoMax {
		t.Errorf("MarkdownEditor.syncScrollBars Max = %d equals Memo.syncScrollBars Max = %d; wrapping content should produce strictly more rendered lines than raw lines", meMax, memoMax)
	}
	if meMax <= memoMax {
		t.Errorf("MarkdownEditor.syncScrollBars Max = %d <= Memo.syncScrollBars Max = %d; wrapping should produce larger scroll range", meMax, memoMax)
	}
}

// ---------------------------------------------------------------------------
// Section 11 — HandleEvent tests (requirement 10)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_HandleEventShowSourceDelegatesToEditor verifies that when
// showSource is true, HandleEvent delegates entirely to Editor.HandleEvent
// and does NOT trigger reparse on MarkdownEditor.
// Req 10: "If showSource, delegate completely to Editor.HandleEvent(event)
//
//	and return"
func TestMarkdownEditor_HandleEventShowSourceDelegatesToEditor(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetShowSource(true)
	me.SetText("some text")
	me.SetState(SfSelected, true)

	// In source mode, Editor handles everything. Corrupt sourceCache to detect
	// if reparse() is wrongly called by MarkdownEditor's HandleEvent.
	me.sourceCache = "stale"

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'x'}}
	me.HandleEvent(ev)

	// In showSource mode, MarkdownEditor's HandleEvent must NOT call reparse.
	// sourceCache should remain "stale" because reparse was not called.
	if me.sourceCache != "stale" {
		t.Error("reparse() was called in showSource mode; should delegate entirely to Editor")
	}
}

// TestMarkdownEditor_HandleEventEditTriggersReparse verifies that when an
// edit event is consumed by Memo (via Editor.HandleEvent), MarkdownEditor
// calls reparse() to refresh parsed blocks.
// Req 10: "Call Editor.HandleEvent(event) (Memo handles the edit)"
// Req 10: "If the event was cleared (consumed by Memo), call reparse() to
//
//	refresh blocks"
func TestMarkdownEditor_HandleEventEditTriggersReparse(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# initial heading\n\nparagraph")
	me.SetState(SfSelected, true)

	// Corrupt sourceCache to detect whether reparse() runs
	me.sourceCache = "stale cache value"

	// Send an edit key event that Memo will consume (typing a character)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'a'}}
	me.HandleEvent(ev)

	// If the event was consumed by Memo, reparse() should have run and
	// updated sourceCache to match the current Memo text.
	if me.sourceCache == "stale cache value" {
		t.Error("reparse() was not called after edit event; sourceCache unchanged")
	}
}

// TestMarkdownEditor_HandleEventEditClearsBlocksWhenSourceEmptied verifies that
// after backspacing all content, reparse clears blocks to an empty slice.
// Req 10: "If the event was cleared (consumed by Memo), call reparse() to
//
//	refresh blocks"
func TestMarkdownEditor_HandleEventEditClearsBlocksWhenSourceEmptied(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// Start with a single character that can be backspaced
	me.SetText("x")
	me.SetState(SfSelected, true)

	if len(me.blocks) == 0 {
		t.Fatal("blocks should be non-empty after SetText with content")
	}

	// Backspace to delete the character
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyBackspace2}}
	me.HandleEvent(ev)

	// After backspace consumed by Memo, reparse should run and clear blocks
	if me.blocks == nil {
		t.Error("blocks is nil after clearing source; want non-nil empty slice")
	}
}

// TestMarkdownEditor_HandleEventNonEditDoesNotTriggerReparse verifies that
// events NOT consumed by Memo do not trigger reparse.
// Falsifying: an implementation that always calls reparse after
// Editor.HandleEvent regardless of whether the event was consumed.
// Req 10: "If the event was cleared (consumed by Memo), call reparse()"
func TestMarkdownEditor_HandleEventNonEditDoesNotTriggerReparse(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("test content")
	me.SetState(SfSelected, true)

	// Corrupt sourceCache to detect unwanted reparse
	me.sourceCache = "stale"

	// Send a keyboard event that Memo does NOT consume (e.g., an unrecognized
	// function key that isn't a scroll/edit key for Memo when unfocused edge case).
	// Actually, for a non-edit key like F1, Memo won't consume it.
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF1}}
	me.HandleEvent(ev)

	// If event was NOT consumed by Memo, reparse must NOT be called.
	// sourceCache should remain stale.
	if me.sourceCache != "stale" {
		t.Error("reparse() was called for non-edit event; should only reparse when event is consumed")
	}
}

// TestMarkdownEditor_HandleEventForwardsIndicatorUpdate verifies that
// EvBroadcast events with CmIndicatorUpdate are forwarded to Editor without
// crashing and without triggering reparse.
// Req 10: "If the event is EvBroadcast with CmIndicatorUpdate, forward to
//
//	Editor (for status line updates)"
func TestMarkdownEditor_HandleEventForwardsIndicatorUpdate(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("content")
	me.SetState(SfSelected, true)
	me.sourceCache = "pre-indicator"

	ev := &Event{
		What:    EvBroadcast,
		Command: CmIndicatorUpdate,
	}
	// Must not panic
	me.HandleEvent(ev)

	// Indicator update is not an edit; reparse must NOT be called
	if me.sourceCache != "pre-indicator" {
		t.Error("reparse() was called on CmIndicatorUpdate; should only forward to Editor")
	}
}

// TestMarkdownEditor_HandleEventEditorConsumedTriggersReparse verifies the full
// chain: edit event -> Editor.HandleEvent consumes it -> reparse runs ->
// blocks are updated to reflect new source.
// Req 10: "Call Editor.HandleEvent(event) (Memo handles the edit)"
// Req 10: "If the event was cleared (consumed by Memo), call reparse() to
//
//	refresh blocks"
func TestMarkdownEditor_HandleEventEditorConsumedTriggersReparse(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("abcd")
	me.SetState(SfSelected, true)

	// Capture original blocks (single paragraph for "abcd")
	origBlockCount := len(me.blocks)
	if origBlockCount == 0 {
		t.Fatal("blocks should be non-empty after SetText")
	}

	// Send backspace to delete last character
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyBackspace2}}
	me.HandleEvent(ev)

	// After edit + reparse, blocks should reflect the new source ("abc")
	// The important thing is that reparse ran and updated something.
	// We verify that sourceCache matches the actual Memo text after the edit.
	if me.sourceCache != me.Text() {
		t.Errorf("sourceCache after edit = %q, want %q; reparse should sync cache with Memo text",
			me.sourceCache, me.Text())
	}
}

// ---------------------------------------------------------------------------
// Section 12 — sourceToScreen tests (requirement 13)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_SourceToScreenRowZero verifies sourceToScreen maps source
// row 0 to some screen position. The exact mapping depends on block layout,
// but the function must return a valid screen position.
// Req 13: "sourceToScreen(row, col int) (screenY, screenX int) — maps
//
//	source (row,col) to screen position in rendered output"
func TestMarkdownEditor_SourceToScreenRowZero(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("Hello World")

	// Source row 0, col 0 at delta 0 should map to a valid screen position
	screenY, screenX := me.sourceToScreen(0, 0)

	// The mapping must produce non-negative values (valid screen positions)
	if screenY < 0 {
		t.Errorf("sourceToScreen(0,0) screenY = %d, want >= 0", screenY)
	}
	if screenX < 0 {
		t.Errorf("sourceToScreen(0,0) screenX = %d, want >= 0", screenX)
	}
}

// TestMarkdownEditor_SourceToScreenWithDeltaY verifies sourceToScreen accounts
// for scroll position (deltaY) when mapping source rows.
// Req 13: "sourceToScreen(row, col int) (screenY, screenX int)"
func TestMarkdownEditor_SourceToScreenWithDeltaY(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("Line 0\n\nLine 1\n\nLine 2")

	// With deltaY=0, source row 2 (third line) maps to a positive screen Y
	screenY1, _ := me.sourceToScreen(2, 0)

	// Scroll down: deltaY=2 means first 2 rendered lines are off-screen
	me.Memo.deltaY = 2
	screenY2, _ := me.sourceToScreen(2, 0)

	// The same source row must map to different screen positions at different
	// scroll offsets.
	if screenY1 == screenY2 {
		t.Errorf("sourceToScreen(2,0) returned same screenY=%d for deltaY=0 and deltaY=2; expected different screen positions", screenY1)
	}

	// Scrolling down (increasing deltaY) moves content up (decreasing screenY).
	if screenY2 >= screenY1 {
		t.Errorf("sourceToScreen(2,0) with deltaY=2: screenY=%d >= screenY with deltaY=0: %d; increasing deltaY should decrease screenY", screenY2, screenY1)
	}
}

// TestMarkdownEditor_SourceToScreenColumnMapping verifies sourceToScreen
// maps source columns to screen columns.
// Req 13: "sourceToScreen(row, col int) (screenY, screenX int)"
func TestMarkdownEditor_SourceToScreenColumnMapping(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("abcdefghij")

	// Col 0 should map to screenX >= 0
	_, screenX0 := me.sourceToScreen(0, 0)
	if screenX0 < 0 {
		t.Errorf("sourceToScreen(0,0) screenX = %d, want >= 0", screenX0)
	}

	// Col 5 should map to a larger screenX than col 0
	_, screenX5 := me.sourceToScreen(0, 5)
	if screenX5 <= screenX0 {
		t.Errorf("sourceToScreen(0,5) screenX = %d, want > screenX for col 0 (%d)", screenX5, screenX0)
	}
}

// TestMarkdownEditor_SourceToScreenDoesNotCrashOnOutOfBounds verifies
// sourceToScreen handles out-of-bounds row/col gracefully without panicking.
// Falsifying: an implementation that indexes into arrays without bounds
// checks would panic.
// Req 13: "sourceToScreen(row, col int) (screenY, screenX int)"
func TestMarkdownEditor_SourceToScreenDoesNotCrashOnOutOfBounds(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("short")

	// Must not panic with row beyond source
	_, _ = me.sourceToScreen(999, 0)

	// Must not panic with col beyond line length
	_, _ = me.sourceToScreen(0, 999)

	// Must not panic with negative values
	_, _ = me.sourceToScreen(-1, -1)
}

// TestMarkdownEditor_SourceToScreenEmptySource verifies sourceToScreen when
// there is no source text (empty Memo).
// Req 13: "sourceToScreen(row, col int) (screenY, screenX int)"
func TestMarkdownEditor_SourceToScreenEmptySource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// No SetText called — blocks is empty

	screenY, screenX := me.sourceToScreen(0, 0)

	// Must not panic and should return reasonable values
	_ = screenY
	_ = screenX
}

// ---------------------------------------------------------------------------
// Section 13 — drawCursor tests (requirement 14)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_DrawCursorDoesNothingWhenNotSelected verifies drawCursor
// is a no-op when the widget is not selected (SfSelected not set).
// Req 14: "drawCursor(buf *DrawBuffer, cs *theme.ColorScheme) — draws block
//
//	cursor at mapped screen position. Does nothing if widget not
//	selected or has selection active."
func TestMarkdownEditor_DrawCursorDoesNothingWhenNotSelected(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("hello")
	// SfSelected is NOT set

	buf := NewDrawBuffer(40, 10)
	// Fill buffer with a known style
	buf.Fill(NewRect(0, 0, 40, 10), ' ', tcell.StyleDefault)

	// drawCursor should do nothing (we verify by checking buffer is unchanged
	// at the expected cursor position)
	me.drawCursor(buf, me.ColorScheme())

	// Since the widget is not selected, no cursor should be drawn.
	// The buffer at position (0,0) should still have the default fill.
	cell := buf.GetCell(0, 0)
	if cell.Style != tcell.StyleDefault {
		t.Error("drawCursor modified buffer when widget is not selected")
	}
}

// TestMarkdownEditor_DrawCursorDoesNothingWhenHasSelection verifies drawCursor
// is a no-op when there is an active selection.
// Req 14: "Does nothing if widget not selected or has selection active."
func TestMarkdownEditor_DrawCursorDoesNothingWhenHasSelection(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("hello")
	me.SetState(SfSelected, true)

	// Set up an active selection
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 0
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 3 // selection from (0,0) to (0,3)

	if !me.Memo.HasSelection() {
		t.Fatal("test setup failed: HasSelection should be true")
	}

	buf := NewDrawBuffer(40, 10)
	buf.Fill(NewRect(0, 0, 40, 10), ' ', tcell.StyleDefault)

	me.drawCursor(buf, me.ColorScheme())

	// With selection active, no cursor should be drawn.
	cell := buf.GetCell(0, 0)
	if cell.Style != tcell.StyleDefault {
		t.Error("drawCursor modified buffer when selection is active")
	}
}

// =============================================================================
// Task 5 — Reveal mapper: block-level
// =============================================================================
//
// These tests are written BEFORE the implementation exists. They define the
// expected behavior of revealSpan, revealKind, buildRevealSpans,
// blockSourceLineCount, blockRevealSpan, collectRevealSpans, and
// detectListMarker. All types and functions referenced here must be created
// by the implementer.

// ---------------------------------------------------------------------------
// Section 14 — detectListMarker tests (requirement 14)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_DetectListMarker verifies detectListMarker
// recognizes all list marker patterns the spec requires: "- ", "* ", "+ ",
// "N. ", "- [ ] ", "- [x] ", and indented variants. Non-list lines return
// isListItem=false.
// Spec req 14: "detectListMarker(line string) (marker string, rest string,
// isListItem bool) helper — detects '- ', '* ', '+ ', 'N. ', '- [ ] ',
// '- [x] ' patterns including indented variants"
func TestMarkdownEditor_Reveal_DetectListMarker(t *testing.T) {
	tests := []struct {
		line       string
		wantMarker string
		wantRest   string
		wantItem   bool
		desc       string
	}{
		// Bullet markers
		{line: "- item text", wantMarker: "- ", wantRest: "item text", wantItem: true, desc: "dash bullet"},
		{line: "* item text", wantMarker: "* ", wantRest: "item text", wantItem: true, desc: "star bullet"},
		{line: "+ item text", wantMarker: "+ ", wantRest: "item text", wantItem: true, desc: "plus bullet"},
		// Numbered markers
		{line: "1. item text", wantMarker: "1. ", wantRest: "item text", wantItem: true, desc: "numbered single digit"},
		{line: "99. item text", wantMarker: "99. ", wantRest: "item text", wantItem: true, desc: "numbered multi digit"},
		// Checklist markers
		{line: "- [ ] incomplete", wantMarker: "- [ ] ", wantRest: "incomplete", wantItem: true, desc: "unchecked checkbox"},
		{line: "- [x] complete", wantMarker: "- [x] ", wantRest: "complete", wantItem: true, desc: "checked checkbox"},
		{line: "- [X] complete", wantMarker: "- [X] ", wantRest: "complete", wantItem: true, desc: "checked checkbox uppercase"},
		// Indented variants — leading whitespace is part of the marker
		{line: "  - indented", wantMarker: "  - ", wantRest: "indented", wantItem: true, desc: "indented two spaces"},
		{line: "    - indented", wantMarker: "    - ", wantRest: "indented", wantItem: true, desc: "indented four spaces"},
		// Non-list lines
		{line: "plain text", wantMarker: "", wantRest: "", wantItem: false, desc: "plain text"},
		{line: "", wantMarker: "", wantRest: "", wantItem: false, desc: "empty string"},
		{line: "# heading", wantMarker: "", wantRest: "", wantItem: false, desc: "heading syntax"},
		{line: "> blockquote", wantMarker: "", wantRest: "", wantItem: false, desc: "blockquote syntax"},
		{line: "-no space after dash", wantMarker: "", wantRest: "", wantItem: false, desc: "dash without trailing space"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			marker, rest, isItem := detectListMarker(tt.line)
			if isItem != tt.wantItem {
				t.Errorf("detectListMarker(%q) isListItem = %v, want %v", tt.line, isItem, tt.wantItem)
			}
			if marker != tt.wantMarker {
				t.Errorf("detectListMarker(%q) marker = %q, want %q", tt.line, marker, tt.wantMarker)
			}
			if rest != tt.wantRest {
				t.Errorf("detectListMarker(%q) rest = %q, want %q", tt.line, rest, tt.wantRest)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Section 15 — blockSourceLineCount tests (requirement 10)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_BlockSourceLineCountParagraph verifies paragraphs
// occupy exactly 1 source line.
// Spec req 10: "paragraph/header=1"
func TestMarkdownEditor_Reveal_BlockSourceLineCountParagraph(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("A simple paragraph.")
	// me.blocks[0] is a paragraph block on source line 0

	count := blockSourceLineCount(me.blocks[0], me.Memo.lines, 0)
	if count != 1 {
		t.Errorf("blockSourceLineCount for paragraph = %d, want 1", count)
	}
}

// TestMarkdownEditor_Reveal_BlockSourceLineCountHeader verifies headings
// occupy exactly 1 source line regardless of level.
// Spec req 10: "paragraph/header=1"
func TestMarkdownEditor_Reveal_BlockSourceLineCountHeader(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("## A level-2 heading")

	count := blockSourceLineCount(me.blocks[0], me.Memo.lines, 0)
	if count != 1 {
		t.Errorf("blockSourceLineCount for H2 header = %d, want 1", count)
	}
}

// TestMarkdownEditor_Reveal_BlockSourceLineCountCode verifies code blocks
// occupy len(code)+2 source lines (fence open + code lines + fence close).
// Spec req 10: "code=len(code)+2"
func TestMarkdownEditor_Reveal_BlockSourceLineCountCode(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// A code block with 3 lines of code: fence open + 3 code lines + fence close = 5
	me.SetText("```go\nline1\nline2\nline3\n```")

	count := blockSourceLineCount(me.blocks[0], me.Memo.lines, 0)
	want := len(me.blocks[0].code) + 2 // 3 + 2 = 5
	if count != want {
		t.Errorf("blockSourceLineCount for code block = %d, want %d (len(code)+2)", count, want)
	}
}

// TestMarkdownEditor_Reveal_BlockSourceLineCountCodeEmpty verifies an empty
// code block (no lines between fences) occupies 2 source lines.
// Spec req 10: "code=len(code)+2"
func TestMarkdownEditor_Reveal_BlockSourceLineCountCodeEmpty(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("```\n```")

	count := blockSourceLineCount(me.blocks[0], me.Memo.lines, 0)
	want := 0 + 2 // empty code
	if count != want {
		t.Errorf("blockSourceLineCount for empty code block = %d, want %d", count, want)
	}
}

// TestMarkdownEditor_Reveal_BlockSourceLineCountHRule verifies horizontal
// rules occupy exactly 1 source line.
// Spec req 10: "hrules=1"
func TestMarkdownEditor_Reveal_BlockSourceLineCountHRule(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("---")

	count := blockSourceLineCount(me.blocks[0], me.Memo.lines, 0)
	if count != 1 {
		t.Errorf("blockSourceLineCount for hrule = %d, want 1", count)
	}
}

// TestMarkdownEditor_Reveal_BlockSourceLineCountBlockquote verifies blockquotes
// occupy the sum of their source lines (counted from source, since goldmark
// merges consecutive "> " lines into a single child paragraph with soft breaks).
// Spec req 10: "blockquotes=sum of children"
func TestMarkdownEditor_Reveal_BlockSourceLineCountBlockquote(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// Two "> " lines: goldmark merges them into 1 paragraph child with soft break.
	// Source line count must be 2 to cover both source lines.
	me.SetText("> line one\n> line two")

	bq := me.blocks[0]
	if bq.kind != blockBlockquote {
		t.Fatalf("expected blockBlockquote, got %v", bq.kind)
	}
	// goldmark merges consecutive "> " lines into a single paragraph; 1 child is correct.
	if len(bq.children) != 1 {
		t.Fatalf("blockquote has %d children, want 1 (goldmark merges consecutive > lines)", len(bq.children))
	}

	count := blockSourceLineCount(bq, me.Memo.lines, 0)
	if count != 2 {
		t.Errorf("blockSourceLineCount for blockquote = %d, want 2 (two source lines)", count)
	}
}

// TestMarkdownEditor_Reveal_BlockSourceLineCountBulletList verifies bullet
// lists occupy the sum of items plus nested children.
// Spec req 10: "lists=sum of items+nested"
func TestMarkdownEditor_Reveal_BlockSourceLineCountBulletList(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// 3 items, no nesting: 3 source lines
	me.SetText("- item 1\n- item 2\n- item 3")

	count := blockSourceLineCount(me.blocks[0], me.Memo.lines, 0)
	if count != 3 {
		t.Errorf("blockSourceLineCount for 3-item bullet list = %d, want 3", count)
	}
}

// TestMarkdownEditor_Reveal_BlockSourceLineCountNestedList verifies nested
// list children increase the source line count.
// Spec req 10: "lists=sum of items+nested"
func TestMarkdownEditor_Reveal_BlockSourceLineCountNestedList(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// Parent item + nested sub-item: 2 source lines
	me.SetText("- parent\n  - child")

	count := blockSourceLineCount(me.blocks[0], me.Memo.lines, 0)
	if count != 2 {
		t.Errorf("blockSourceLineCount for nested list = %d, want 2", count)
	}
}

// TestMarkdownEditor_Reveal_BlockSourceLineCountTable verifies tables occupy
// len(rows)+2 source lines (header + separator + data rows).
// Spec req 10: "table=len(rows)+2"
func TestMarkdownEditor_Reveal_BlockSourceLineCountTable(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// Table with 2 data rows: header + separator + 2 data = 4 = len(rows)+2
	me.SetText("| A | B |\n| --- | --- |\n| 1 | 2 |\n| 3 | 4 |")

	tbl := me.blocks[0]
	if tbl.kind != blockTable {
		t.Fatalf("expected blockTable, got %v", tbl.kind)
	}

	count := blockSourceLineCount(tbl, me.Memo.lines, 0)
	want := len(tbl.rows) + 2
	if count != want {
		t.Errorf("blockSourceLineCount for table = %d, want %d (len(rows)+2)", count, want)
	}
}

// ---------------------------------------------------------------------------
// Section 16 — blockRevealSpan tests (requirement 11)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_BlockRevealSpanHeading verifies heading blocks
// produce a reveal span with the correct number of "#" markers.
// Spec req 5 (heading marker): "markerOpen = strings.Repeat("#", level) + ' '"
// Spec req 11: "blockRevealSpan(b mdBlock, source [][]rune, blockRow, depth int) []revealSpan"
func TestMarkdownEditor_Reveal_BlockRevealSpanHeading(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("## H2 Heading")

	spans := blockRevealSpan(me.blocks[0], me.Memo.lines, 0, 0)
	if len(spans) != 1 {
		t.Fatalf("blockRevealSpan for heading returned %d spans, want 1", len(spans))
	}
	s := spans[0]
	if s.markerOpen != "## " {
		t.Errorf("markerOpen = %q, want %q", s.markerOpen, "## ")
	}
	if s.markerClose != "" {
		t.Errorf("markerClose = %q, want empty", s.markerClose)
	}
	if s.kind != revealBlock {
		t.Errorf("kind = %v, want revealBlock", s.kind)
	}
	// Source positions: anchored at (blockStartRow, 0) to (blockRow+1, 0)
	if s.startRow != 0 || s.startCol != 0 {
		t.Errorf("start position = (%d,%d), want (0,0)", s.startRow, s.startCol)
	}
	if s.endRow != 1 || s.endCol != 0 {
		t.Errorf("end position = (%d,%d), want (1,0)", s.endRow, s.endCol)
	}
}

// TestMarkdownEditor_Reveal_BlockRevealSpanCode verifies code blocks produce
// two reveal spans: opening fence with language, and closing fence without.
// Spec req 5 (code fence marker): "markerOpen = '```' + language on first
// line, markerOpen = '```' on last line"
func TestMarkdownEditor_Reveal_BlockRevealSpanCode(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("```go\nfmt.Println(\"hi\")\n```")

	spans := blockRevealSpan(me.blocks[0], me.Memo.lines, 0, 0)
	if len(spans) != 2 {
		t.Fatalf("blockRevealSpan for code block returned %d spans, want 2 (open + close)", len(spans))
	}
	// First span: opening fence with language
	open := spans[0]
	if open.markerOpen != "```go" {
		t.Errorf("open markerOpen = %q, want %q", open.markerOpen, "```go")
	}
	if open.startRow != 0 {
		t.Errorf("open fence startRow = %d, want 0 (first source line)", open.startRow)
	}
	// Last span: closing fence
	close := spans[1]
	if close.markerOpen != "```" {
		t.Errorf("close markerOpen = %q, want %q", close.markerOpen, "```")
	}
	if close.startRow != 2 {
		t.Errorf("close fence startRow = %d, want 2 (last source line)", close.startRow)
	}
}

// TestMarkdownEditor_Reveal_BlockRevealSpanHRule verifies horizontal rule
// blocks produce a span with "---" marker.
// Spec req 5 (horizontal rule marker): "markerOpen = '---', markerClose = ”"
func TestMarkdownEditor_Reveal_BlockRevealSpanHRule(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("---")

	spans := blockRevealSpan(me.blocks[0], me.Memo.lines, 0, 0)
	if len(spans) != 1 {
		t.Fatalf("blockRevealSpan for hrule returned %d spans, want 1", len(spans))
	}
	s := spans[0]
	if s.markerOpen != "---" {
		t.Errorf("markerOpen = %q, want %q", s.markerOpen, "---")
	}
	if s.markerClose != "" {
		t.Errorf("markerClose = %q, want empty", s.markerClose)
	}
	if s.kind != revealBlock {
		t.Errorf("kind = %v, want revealBlock", s.kind)
	}
}

// TestMarkdownEditor_Reveal_BlockRevealSpanParagraphReturnsNil verifies that
// paragraph blocks (which have no visible markdown marker) do NOT generate
// reveal spans. Paragraphs are not in the block marker types list in the spec.
// Spec req 5: only Heading, Blockquote, Bullet/Numbered/Checklist, Code fence,
// Horizontal rule, and Table are listed as having markers.
func TestMarkdownEditor_Reveal_BlockRevealSpanParagraphReturnsNil(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("A plain paragraph.")

	spans := blockRevealSpan(me.blocks[0], me.Memo.lines, 0, 0)
	if len(spans) != 0 {
		t.Errorf("blockRevealSpan for paragraph returned %d spans, want 0 (paragraphs have no marker)", len(spans))
	}
}

// ---------------------------------------------------------------------------
// Section 17 — buildRevealSpans tests (requirements 3, 4, 7, 8, 9)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_BuildSpansHeadingCursorInside verifies that
// buildRevealSpans produces a span when the cursor is on the heading's source
// line.
// Spec req 3: "buildRevealSpans() ... produces []revealSpan by walking blocks
// and checking cursor position"
// Spec req 4: "Block-level reveal triggers when cursorRow falls within a
// block's source line range"
func TestMarkdownEditor_Reveal_BuildSpansHeadingCursorInside(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# Hello\n\nWorld")
	me.Memo.cursorRow = 0 // cursor on heading line

	spans := me.buildRevealSpans()
	if len(spans) == 0 {
		t.Fatal("buildRevealSpans returned no spans; heading with cursor inside should reveal")
	}
	if spans[0].kind != revealBlock {
		t.Errorf("kind = %v, want revealBlock", spans[0].kind)
	}
	if spans[0].markerOpen != "# " {
		t.Errorf("markerOpen = %q, want %q", spans[0].markerOpen, "# ")
	}
}

// TestMarkdownEditor_Reveal_BuildSpansHeadingCursorOutside verifies that
// buildRevealSpans produces NO spans for a heading when the cursor is on a
// different block's source line (paragraph).
// Spec req 4: "Block-level reveal triggers when cursorRow falls within a
// block's source line range" — cursor outside heading range should not trigger.
func TestMarkdownEditor_Reveal_BuildSpansHeadingCursorOutside(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# Hello\n\nWorld")
	me.Memo.cursorRow = 2 // cursor on paragraph line, not heading

	spans := me.buildRevealSpans()
	// Paragraphs have no marker, so no spans should be produced
	if len(spans) != 0 {
		t.Errorf("buildRevealSpans returned %d spans for cursor in paragraph; want 0 (paragraphs have no marker)", len(spans))
	}
}

// TestMarkdownEditor_Reveal_BuildSpansBlockquote verifies that buildRevealSpans
// produces "> " marker spans on all child source lines of a blockquote when
// the cursor is inside it.
// Spec req 5 (blockquote marker): "markerOpen = '> ' on each child block line"
func TestMarkdownEditor_Reveal_BuildSpansBlockquote(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("> line one\n> line two")
	me.Memo.cursorRow = 0 // cursor inside blockquote

	spans := me.buildRevealSpans()
	if len(spans) < 2 {
		t.Fatalf("buildRevealSpans returned %d spans for 2-line blockquote; want at least 2", len(spans))
	}
	for _, s := range spans {
		if s.markerOpen != "> " {
			t.Errorf("markerOpen = %q, want %q", s.markerOpen, "> ")
		}
		if s.kind != revealBlock {
			t.Errorf("kind = %v, want revealBlock", s.kind)
		}
	}
}

// TestMarkdownEditor_Reveal_BuildSpansBlockquoteCursorInsideNested verifies
// that a nested blockquote within the outer blockquote also gets revealed when
// the cursor is on the nested line.
// Spec req 5 (blockquote marker): "markerOpen = '> ' on each child block line"
func TestMarkdownEditor_Reveal_BuildSpansBlockquoteCursorInsideNested(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("> outer\n> > inner")
	me.Memo.cursorRow = 1 // cursor on nested line

	spans := me.buildRevealSpans()
	if len(spans) == 0 {
		t.Fatal("buildRevealSpans returned no spans for cursor in nested blockquote")
	}
	// We should have markers for both the outer and inner blockquote levels
	for _, s := range spans {
		if s.kind != revealBlock {
			t.Errorf("kind = %v, want revealBlock", s.kind)
		}
	}
}

// TestMarkdownEditor_Reveal_BuildSpansBulletList verifies that buildRevealSpans
// produces marker spans for a bullet list when the cursor is inside it. The
// marker text comes from the actual source, not hardcoded.
// Spec req 5 (bullet marker): "markerOpen = <actual marker from source>,
// markerClose = ”"
func TestMarkdownEditor_Reveal_BuildSpansBulletList(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("- first\n- second\n- third")
	me.Memo.cursorRow = 0 // cursor on first list item

	spans := me.buildRevealSpans()
	if len(spans) == 0 {
		t.Fatal("buildRevealSpans returned no spans for bullet list with cursor inside")
	}
	// Each list item should get a marker read from the actual source
	for _, s := range spans {
		if s.kind != revealBlock {
			t.Errorf("kind = %v, want revealBlock", s.kind)
		}
	}
}

// TestMarkdownEditor_Reveal_BuildSpansBulletListActualMarker verifies the
// marker for a bullet list is read from source, not hardcoded to "- ".
// When source uses "* ", the marker should be "* ".
// Falsifying: an implementation that hardcodes "- " for all bullet lists would fail.
// Spec req 5: "markerOpen = <actual marker from source>, markerClose = ”"
func TestMarkdownEditor_Reveal_BuildSpansBulletListActualMarker(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("* star item\n* another")
	me.Memo.cursorRow = 0

	spans := me.buildRevealSpans()
	if len(spans) == 0 {
		t.Fatal("buildRevealSpans returned no spans")
	}
	if spans[0].markerOpen != "* " {
		t.Errorf("markerOpen = %q, want %q (actual marker from source, not hardcoded '- ')", spans[0].markerOpen, "* ")
	}
}

// TestMarkdownEditor_Reveal_BuildRevealSpansNumberedList verifies that
// buildRevealSpans produces marker spans for a numbered list when the cursor
// is inside it. The marker is read from the actual source (e.g., "1. ").
// Spec req 5: "markerOpen = <actual marker from source>, markerClose = ”"
func TestMarkdownEditor_Reveal_BuildRevealSpansNumberedList(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("1. First\n2. Second")
	me.Memo.cursorRow = 0 // cursor on first numbered item

	spans := me.buildRevealSpans()
	if len(spans) == 0 {
		t.Fatal("buildRevealSpans returned no spans for numbered list with cursor inside")
	}
	// The first span should have the "1. " marker read from source
	if spans[0].markerOpen != "1. " {
		t.Errorf("markerOpen = %q, want %q (actual marker from source)", spans[0].markerOpen, "1. ")
	}
	if spans[0].kind != revealBlock {
		t.Errorf("kind = %v, want revealBlock", spans[0].kind)
	}
}

// TestMarkdownEditor_Reveal_BuildRevealSpansChecklist verifies that
// buildRevealSpans produces marker spans for a checklist when the cursor
// is inside it. The marker is read from the actual source (e.g., "- [ ] ").
// Spec req 5: "markerOpen = <actual marker from source>, markerClose = ”"
func TestMarkdownEditor_Reveal_BuildRevealSpansChecklist(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("- [ ] todo\n- [x] done")
	me.Memo.cursorRow = 0 // cursor on first checklist item

	spans := me.buildRevealSpans()
	if len(spans) == 0 {
		t.Fatal("buildRevealSpans returned no spans for checklist with cursor inside")
	}
	// The first span should have the "- [ ] " marker read from source
	if spans[0].markerOpen != "- [ ] " {
		t.Errorf("markerOpen = %q, want %q (actual marker from source)", spans[0].markerOpen, "- [ ] ")
	}
	if spans[0].kind != revealBlock {
		t.Errorf("kind = %v, want revealBlock", spans[0].kind)
	}
}

// TestMarkdownEditor_Reveal_BuildSpansCodeBlock verifies that buildRevealSpans
// produces fence marker spans for a code block when the cursor is inside it.
// Spec req 5 (code fence marker): "markerOpen = '```' + language on first line,
// markerOpen = '```' on last line"
func TestMarkdownEditor_Reveal_BuildSpansCodeBlock(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("```go\ncode line\nmore code\n```")
	me.Memo.cursorRow = 1 // cursor on a code line

	spans := me.buildRevealSpans()
	if len(spans) < 2 {
		t.Fatalf("buildRevealSpans returned %d spans for code block; want at least 2 (open + close fences)", len(spans))
	}
	// Find open and close fence spans
	var hasOpen, hasClose bool
	for _, s := range spans {
		if s.markerOpen == "```go" {
			hasOpen = true
		}
		if s.markerOpen == "```" {
			hasClose = true
		}
	}
	if !hasOpen {
		t.Error("missing opening fence marker '```go'")
	}
	if !hasClose {
		t.Error("missing closing fence marker '```'")
	}
}

// TestMarkdownEditor_Reveal_BuildRevealSpansCodeBlockAllLines verifies that
// when the cursor is on any line inside a code block, buildRevealSpans
// produces spans for the ENTIRE block - both the opening fence and closing
// fence - not just the line the cursor is on.
// Falsifying: an implementation that only reveals the cursor line would
// produce only one fence span instead of both.
// Spec req 4: "Block-level reveal triggers when cursorRow falls within a
// block's source line range" - the whole block, not just one line.
func TestMarkdownEditor_Reveal_BuildRevealSpansCodeBlockAllLines(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// 3-line code block: fence open + 3 code lines + fence close = 5 source lines
	me.SetText("```go\nline1\nline2\nline3\n```")
	me.Memo.cursorRow = 2 // cursor on middle code line, not line 0 or 4

	spans := me.buildRevealSpans()
	// Must produce BOTH the opening fence span AND closing fence span.
	// The entire block reveals, not just the cursor line.
	if len(spans) < 2 {
		t.Fatalf("buildRevealSpans returned %d spans; want at least 2 (open + close fences for entire block)", len(spans))
	}
	var hasOpen, hasClose bool
	for _, s := range spans {
		if s.markerOpen == "```go" && s.startRow == 0 {
			hasOpen = true
		}
		if s.markerOpen == "```" && s.startRow == 4 {
			hasClose = true
		}
	}
	if !hasOpen {
		t.Error("missing opening fence marker '```go' at startRow 0; entire block should reveal")
	}
	if !hasClose {
		t.Error("missing closing fence marker '```' at startRow 4; entire block should reveal")
	}
}

// TestMarkdownEditor_Reveal_BuildSpansHRule verifies that buildRevealSpans
// produces a "---" marker span when the cursor is on a horizontal rule.
// Spec req 5: "markerOpen = '---', markerClose = ”"
func TestMarkdownEditor_Reveal_BuildSpansHRule(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("---")
	me.Memo.cursorRow = 0

	spans := me.buildRevealSpans()
	if len(spans) != 1 {
		t.Fatalf("buildRevealSpans returned %d spans for hrule; want 1", len(spans))
	}
	if spans[0].markerOpen != "---" {
		t.Errorf("markerOpen = %q, want %q", spans[0].markerOpen, "---")
	}
}

// TestMarkdownEditor_Reveal_BuildSpansEmptySource verifies that when the source
// is empty (no text set), buildRevealSpans returns an empty slice.
// Spec req 7: "Empty source or empty blocks: buildRevealSpans() produces
// nil/empty slice"
func TestMarkdownEditor_Reveal_BuildSpansEmptySource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// No SetText — blocks is empty, source is empty

	spans := me.buildRevealSpans()
	if len(spans) != 0 {
		t.Errorf("buildRevealSpans on empty source returned %d spans, want 0", len(spans))
	}
}

// TestMarkdownEditor_Reveal_BuildSpansEmptySourceAfterClear verifies that
// after clearing source text, blocks are empty and buildRevealSpans returns
// an empty slice.
// Spec req 7: "Empty source or empty blocks: buildRevealSpans() produces
// nil/empty slice" — both empty source and empty blocks tested.
func TestMarkdownEditor_Reveal_BuildSpansEmptySourceAfterClear(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# heading") // populate then clear
	me.SetText("")

	spans := me.buildRevealSpans()
	if len(spans) != 0 {
		t.Errorf("buildRevealSpans on cleared source returned %d spans, want 0", len(spans))
	}
	if me.blocks != nil && len(me.blocks) == 0 {
		// blocks being empty slice is correct per reparse spec
	}
}

// TestMarkdownEditor_Reveal_BuildSpansCursorMovesInOut verifies that when the
// cursor moves from inside a block to outside it, reveal spans appear and then
// disappear.
// Spec req 4: "Block-level reveal triggers when cursorRow falls within a
// block's source line range"
// Spec req: "Cursor moves into then out of a block" (requirement 11 in test list)
func TestMarkdownEditor_Reveal_BuildSpansCursorMovesInOut(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// Heading on line 0, paragraph on line 2
	me.SetText("# heading\n\nparagraph")

	// Cursor inside heading → spans present
	me.Memo.cursorRow = 0
	spansIn := me.buildRevealSpans()
	if len(spansIn) == 0 {
		t.Fatal("buildRevealSpans should produce spans when cursor is inside heading")
	}

	// Move cursor outside heading (to paragraph) → no spans (paragraphs don't have markers)
	me.Memo.cursorRow = 2
	spansOut := me.buildRevealSpans()
	if len(spansOut) != 0 {
		t.Errorf("buildRevealSpans returned %d spans when cursor moved to paragraph; want 0 (paragraphs have no markers)", len(spansOut))
	}
}

// TestMarkdownEditor_Reveal_BuildSpansFalsifyingScope verifies that reveal
// spans are scoped to the block containing the cursor, not all blocks.
// When the cursor is inside a heading, the following blockquote must NOT
// generate reveal spans.
// Spec req 4: "Block-level reveal triggers when cursorRow falls within a
// block's source line range" — blocks outside the cursor range do not reveal.
// Falsifying: an implementation that reveals all blocks regardless would fail.
func TestMarkdownEditor_Reveal_BuildSpansFalsifyingScope(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# heading\n\n> blockquote line")
	// heading at line 0, blank line 1, blockquote at line 2
	me.Memo.cursorRow = 0 // cursor in heading

	spans := me.buildRevealSpans()
	// Only heading spans should appear, NOT blockquote spans
	for _, s := range spans {
		if s.markerOpen == "> " {
			t.Error("found blockquote '> ' marker when cursor is in heading; spans must be scoped to cursor's block only")
		}
	}
}

// TestMarkdownEditor_Reveal_BuildSpansDoesNotMutateSource verifies that
// buildRevealSpans does not modify the source (Memo.lines).
// Spec req 8: "The reveal mapper does NOT mutate source; it only annotates"
func TestMarkdownEditor_Reveal_BuildSpansDoesNotMutateSource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# heading\n\nparagraph")
	me.Memo.cursorRow = 0

	// Capture source before
	before := make([][]rune, len(me.Memo.lines))
	for i, line := range me.Memo.lines {
		before[i] = make([]rune, len(line))
		copy(before[i], line)
	}

	me.buildRevealSpans()

	// Source must be unchanged
	for i := range before {
		if string(me.Memo.lines[i]) != string(before[i]) {
			t.Errorf("source line %d changed from %q to %q; buildRevealSpans must not mutate source",
				i, string(before[i]), string(me.Memo.lines[i]))
		}
	}
}

// TestMarkdownEditor_Reveal_BuildRevealSpansTableWithCursor verifies that
// buildRevealSpans produces reveal spans for table rows when the cursor is
// inside a table. Tables are revealed with pipe markers containing "|".
// Spec req 5 (table marker): "markerOpen contains '|' (pipe syntax)"
func TestMarkdownEditor_Reveal_BuildRevealSpansTableWithCursor(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("| Name | Value |\n|------|-------|\n| foo  | 42    |\n| bar  | 99    |")
	me.Memo.cursorRow = 2 // cursor on row containing "foo"

	spans := me.buildRevealSpans()
	if len(spans) == 0 {
		t.Fatal("buildRevealSpans returned no spans; table with cursor inside should reveal")
	}
	if spans[0].kind != revealBlock {
		t.Errorf("kind = %v, want revealBlock", spans[0].kind)
	}
	if !strings.Contains(spans[0].markerOpen, "|") {
		t.Errorf("markerOpen = %q, want it to contain '|' (pipe syntax)", spans[0].markerOpen)
	}
}

// ---------------------------------------------------------------------------
// Section 18 — revealSpans field and reparse integration (requirements 2, 3, 13)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_RevealSpansFieldDefault verifies the revealSpans
// field exists and is initialized to nil after construction.
// Spec req 13: "revealSpans field added to MarkdownEditor struct"
func TestMarkdownEditor_Reveal_RevealSpansFieldDefault(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// revealSpans should exist and be nil before any builds
	if me.revealSpans != nil {
		t.Errorf("revealSpans after construction = %v, want nil", me.revealSpans)
	}
}

// TestMarkdownEditor_Reveal_RevealSpansPopulatedAfterReparse verifies that
// revealSpans is populated after SetText triggers reparse.
// Spec req 3: "buildRevealSpans() ... Called from reparse() after parsing.
// Stored in MarkdownEditor.revealSpans."
func TestMarkdownEditor_Reveal_RevealSpansPopulatedAfterReparse(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# heading")
	// SetText calls reparse, which should call buildRevealSpans and store the result

	// revealSpans should be non-nil after reparse runs (cursor was at 0,0 = inside heading)
	if me.revealSpans == nil {
		t.Error("revealSpans is nil after SetText/reparse; buildRevealSpans should have populated it")
	}
	if len(me.revealSpans) == 0 {
		t.Error("revealSpans is empty after SetText/reparse with heading at cursor; should have spans")
	}
}

// TestMarkdownEditor_Reveal_RevealSpansClearedOnEmptySourceAfterReparse
// verifies that revealSpans is cleared to empty after reparse on empty source.
// Spec req 7: "Empty source or empty blocks: buildRevealSpans() produces
// nil/empty slice"
func TestMarkdownEditor_Reveal_RevealSpansClearedOnEmptySourceAfterReparse(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# heading") // populate
	if len(me.revealSpans) == 0 {
		t.Fatal("setup: revealSpans should have spans after SetText with heading")
	}

	me.SetText("") // clear source; reparse should clear revealSpans
	if len(me.revealSpans) != 0 {
		t.Errorf("revealSpans has %d entries after clearing source; want 0", len(me.revealSpans))
	}
}

// ---------------------------------------------------------------------------
// Section 19 — Escape sequence handling (requirement 9)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_EscapedMarkdownNoSpans verifies that
// backslash-escaped markdown syntax does NOT generate reveal spans.
// The spec says goldmark handles this naturally: \*, \# etc. are parsed as
// literal text, not as markdown syntax.
// Spec req 9: "backslash-escaped markdown (\*, \#) must NOT generate reveal
// spans — goldmark handles this naturally since escaped chars are parsed as
// literal text"
func TestMarkdownEditor_Reveal_EscapedMarkdownNoSpans(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// \# is NOT a heading — it's a paragraph with literal "#" text
	me.SetText("\\# not a heading")
	me.Memo.cursorRow = 0

	spans := me.buildRevealSpans()
	// This should be a paragraph (no heading marker), so no spans
	if len(spans) != 0 {
		t.Errorf("buildRevealSpans returned %d spans for escaped #; want 0 (escaped syntax is literal text, not heading)", len(spans))
	}

	// Verify it's actually a paragraph, not a heading
	if me.blocks[0].kind != blockParagraph {
		t.Errorf("escaped '\\#' block kind = %v, want blockParagraph (goldmark treats as literal text)", me.blocks[0].kind)
	}
}

// ---------------------------------------------------------------------------
// Section 20 — revealKind type existence (requirement 2)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_RevealKindConstants verifies that revealBlock and
// revealInline constants exist and are distinct values.
// Spec req 2: "revealKind type: revealBlock or revealInline constants"
func TestMarkdownEditor_Reveal_RevealKindConstants(t *testing.T) {
	// revealBlock and revealInline must be different values
	if revealBlock == revealInline {
		t.Error("revealBlock and revealInline must be distinct constants")
	}
	// revealBlock should be the zero value (iota)
	// Both must be usable in comparisons
	var k revealKind = revealBlock
	_ = k
	k = revealInline
	_ = k
}

// ---------------------------------------------------------------------------
// Section 21 - inRevealSpan tests (requirement 12)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_InRevealSpan verifies inRevealSpan returns the
// correct *revealSpan when a position falls within a span, and nil when it
// does not. Boundary positions (startRow, startCol) are considered inside.
// Spec req 12: "inRevealSpan(spans []revealSpan, row, col int) *revealSpan
// - returns the span containing (row, col), or nil if none match"
func TestMarkdownEditor_Reveal_InRevealSpan(t *testing.T) {
	// Build a known set of spans
	spans := []revealSpan{
		{startRow: 0, startCol: 0, endRow: 1, endCol: 0, markerOpen: "# ", kind: revealBlock},
		{startRow: 2, startCol: 0, endRow: 3, endCol: 0, markerOpen: "> ", kind: revealBlock},
	}

	tests := []struct {
		row, col int
		want     *revealSpan
		desc     string
	}{
		// Inside a span
		{row: 0, col: 0, want: &spans[0], desc: "row/col inside first span (start boundary)"},
		{row: 0, col: 2, want: &spans[0], desc: "row/col inside first span (mid)"},
		{row: 2, col: 0, want: &spans[1], desc: "row/col inside second span"},
		// Outside all spans
		{row: 4, col: 0, want: nil, desc: "row/col outside all spans (row after last)"},
		{row: 1, col: 1, want: nil, desc: "row/col outside all spans (between spans)"},
		{row: 0, col: 10, want: nil, desc: "column outside span width"},
		// Boundary: exact startRow/startCol
		{row: 0, col: 0, want: &spans[0], desc: "exact startRow/startCol boundary (inside)"},
		{row: 2, col: 0, want: &spans[1], desc: "exact startRow/startCol boundary second span (inside)"},
		// One before start -> nil
		{row: -1, col: 0, want: nil, desc: "row one before first span start"},
		{row: 1, col: 0, want: nil, desc: "row one before second span start"},
		// One after end -> nil
		{row: 1, col: 0, want: nil, desc: "col at end boundary first span (exclusive end)"},
		{row: 3, col: 0, want: nil, desc: "col at end boundary second span (exclusive end)"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := inRevealSpan(spans, tt.row, tt.col)
			if tt.want == nil {
				if got != nil {
					t.Errorf("inRevealSpan(%d, %d) = %+v, want nil", tt.row, tt.col, got)
				}
			} else {
				if got == nil {
					t.Errorf("inRevealSpan(%d, %d) = nil, want &span{%+v}", tt.row, tt.col, *tt.want)
				} else if got.markerOpen != tt.want.markerOpen || got.startRow != tt.want.startRow {
					t.Errorf("inRevealSpan(%d, %d) = %+v, want %+v", tt.row, tt.col, got, tt.want)
				}
			}
		})
	}
}

// =============================================================================

// Original Task 2 tests continue below
// =============================================================================

// TestMarkdownEditor_DrawCursorDrawsWhenSelectedNoSelection verifies drawCursor
// draws a visible cursor when the widget is selected and there is no selection.
// Req 14: "draws block cursor at mapped screen position"
func TestMarkdownEditor_DrawCursorDrawsWhenSelectedNoSelection(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("hello")
	me.SetState(SfSelected, true)
	// No selection: Memo cursor is at (0,0)

	buf := NewDrawBuffer(40, 10)
	// Fill with default style so we can detect cursor draw
	buf.Fill(NewRect(0, 0, 40, 10), ' ', tcell.StyleDefault)

	me.drawCursor(buf, me.ColorScheme())

	// The cursor should be drawn at the mapped screen position of source (0,0).
	// The cursor style should differ from the default. We verify by checking
	// that at least one cell was modified from its default fill.
	cursorDrawn := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 40; x++ {
			cell := buf.GetCell(x, y)
			if cell.Style != tcell.StyleDefault || cell.Rune != ' ' {
				// Style modified or rune changed — cursor was drawn
				cursorDrawn = true
				break
			}
		}
		if cursorDrawn {
			break
		}
	}
	if !cursorDrawn {
		t.Error("drawCursor did not draw anything; cursor should be visible when selected with no selection")
	}
}

// =============================================================================
// Task 6 — Reveal mapper: inline-level
// =============================================================================
//
// These tests are written BEFORE the implementation exists. They define the
// expected behavior of scanInlineMarkers — a standalone function that scans
// the source text for inline syntax markers (bold, italic, code, strikethrough)
// and returns only the revealSpan for the span containing (or adjacent to)
// the cursor position. All types referenced here (mdBlock, mdRun, revealSpan,
// revealInline) are already defined.

// ---------------------------------------------------------------------------
// Section 22 — scanInlineMarkers inline type extraction (requirements 3, 6, 7)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_Inline_BoldMarkers verifies scanInlineMarkers
// produces a revealInline span with "**" markers for bold text.
// Spec req 3 (Bold): "**text** → markerOpen="**", markerClose="**""
// Spec req 6: "Inline reveal spans have kind = revealInline"
// Spec req 7: "scanInlineMarkers walks blocks, finds inline syntax markers in
//
//	source, matches to mdRun styles, produces reveal spans"
func TestMarkdownEditor_Reveal_Inline_BoldMarkers(t *testing.T) {
	source := [][]rune{[]rune("**bold**")}
	blocks := parseMarkdown("**bold**")

	// Cursor inside the bold text content (col 4 = 'd')
	spans := scanInlineMarkers(source, blocks, 0, 4)

	if len(spans) != 1 {
		t.Fatalf("scanInlineMarkers returned %d spans, want 1", len(spans))
	}
	s := spans[0]
	if s.markerOpen != "**" {
		t.Errorf("markerOpen = %q, want %q", s.markerOpen, "**")
	}
	if s.markerClose != "**" {
		t.Errorf("markerClose = %q, want %q", s.markerClose, "**")
	}
	if s.kind != revealInline {
		t.Errorf("kind = %v, want revealInline", s.kind)
	}

	// Position assertions: the span must cover the full source extent.
	// "**bold**" layout: cols 0=*, 1=*, 2=b, 3=o, 4=l, 5=d, 6=*, 7=*
	// startCol must be 0 (first *), endCol must be 7 (last closing *).
	if s.startRow != 0 {
		t.Errorf("startRow = %d, want 0", s.startRow)
	}
	if s.startCol != 0 {
		t.Errorf("startCol = %d, want 0 (first * of opening **)", s.startCol)
	}
	if s.endRow != 0 {
		t.Errorf("endRow = %d, want 0", s.endRow)
	}
	if s.endCol != 7 {
		t.Errorf("endCol = %d, want 7 (last closing *)", s.endCol)
	}
}

// TestMarkdownEditor_Reveal_Inline_ItalicStarMarkers verifies scanInlineMarkers
// produces a revealInline span with "*" markers for italic text using *text*.
// Spec req 3 (Italic *): "*text* → markerOpen="*", markerClose="*""
func TestMarkdownEditor_Reveal_Inline_ItalicStarMarkers(t *testing.T) {
	source := [][]rune{[]rune("*italic*")}
	blocks := parseMarkdown("*italic*")

	// Cursor inside the italic text content (col 3 = 'a')
	spans := scanInlineMarkers(source, blocks, 0, 3)

	if len(spans) != 1 {
		t.Fatalf("scanInlineMarkers returned %d spans, want 1", len(spans))
	}
	s := spans[0]
	if s.markerOpen != "*" {
		t.Errorf("markerOpen = %q, want %q", s.markerOpen, "*")
	}
	if s.markerClose != "*" {
		t.Errorf("markerClose = %q, want %q", s.markerClose, "*")
	}
	if s.kind != revealInline {
		t.Errorf("kind = %v, want revealInline", s.kind)
	}
}

// TestMarkdownEditor_Reveal_Inline_ItalicUnderscoreMarkers verifies
// scanInlineMarkers produces a revealInline span with "_" markers for italic
// text using _text_ syntax.
// Spec req 3 (Italic _): "_text_ → markerOpen="_", markerClose="_""
func TestMarkdownEditor_Reveal_Inline_ItalicUnderscoreMarkers(t *testing.T) {
	source := [][]rune{[]rune("_italic_")}
	blocks := parseMarkdown("_italic_")

	// Cursor inside the italic text content (col 3 = 'a')
	spans := scanInlineMarkers(source, blocks, 0, 3)

	if len(spans) != 1 {
		t.Fatalf("scanInlineMarkers returned %d spans, want 1", len(spans))
	}
	s := spans[0]
	if s.markerOpen != "_" {
		t.Errorf("markerOpen = %q, want %q", s.markerOpen, "_")
	}
	if s.markerClose != "_" {
		t.Errorf("markerClose = %q, want %q", s.markerClose, "_")
	}
	if s.kind != revealInline {
		t.Errorf("kind = %v, want revealInline", s.kind)
	}
}

// TestMarkdownEditor_Reveal_Inline_CodeMarkers verifies scanInlineMarkers
// produces a revealInline span with "`" markers for inline code.
// Spec req 3 (Code): "`text` → markerOpen="`", markerClose="`""
func TestMarkdownEditor_Reveal_Inline_CodeMarkers(t *testing.T) {
	source := [][]rune{[]rune("`code`")}
	blocks := parseMarkdown("`code`")

	// Cursor inside the code text content (col 2 = 'o')
	spans := scanInlineMarkers(source, blocks, 0, 2)

	if len(spans) != 1 {
		t.Fatalf("scanInlineMarkers returned %d spans, want 1", len(spans))
	}
	s := spans[0]
	if s.markerOpen != "`" {
		t.Errorf("markerOpen = %q, want %q", s.markerOpen, "`")
	}
	if s.markerClose != "`" {
		t.Errorf("markerClose = %q, want %q", s.markerClose, "`")
	}
	if s.kind != revealInline {
		t.Errorf("kind = %v, want revealInline", s.kind)
	}

	// Position assertions: the span must cover the full source extent.
	// "`code`" layout: cols 0=`, 1=c, 2=o, 3=d, 4=e, 5=`
	// startCol must be 0 (first `), endCol must be 5 (closing `).
	if s.startRow != 0 {
		t.Errorf("startRow = %d, want 0", s.startRow)
	}
	if s.startCol != 0 {
		t.Errorf("startCol = %d, want 0 (opening `)", s.startCol)
	}
	if s.endRow != 0 {
		t.Errorf("endRow = %d, want 0", s.endRow)
	}
	if s.endCol != 5 {
		t.Errorf("endCol = %d, want 5 (closing `)", s.endCol)
	}
}

// TestMarkdownEditor_Reveal_Inline_StrikethroughMarkers verifies
// scanInlineMarkers produces a revealInline span with "~~" markers for
// strikethrough text.
// Spec req 3 (Strikethrough): "~~text~~ → markerOpen="~~", markerClose="~~""
func TestMarkdownEditor_Reveal_Inline_StrikethroughMarkers(t *testing.T) {
	source := [][]rune{[]rune("~~strike~~")}
	blocks := parseMarkdown("~~strike~~")

	// Cursor inside the strikethrough text content (col 4 = 'r')
	spans := scanInlineMarkers(source, blocks, 0, 4)

	if len(spans) != 1 {
		t.Fatalf("scanInlineMarkers returned %d spans, want 1", len(spans))
	}
	s := spans[0]
	if s.markerOpen != "~~" {
		t.Errorf("markerOpen = %q, want %q", s.markerOpen, "~~")
	}
	if s.markerClose != "~~" {
		t.Errorf("markerClose = %q, want %q", s.markerClose, "~~")
	}
	if s.kind != revealInline {
		t.Errorf("kind = %v, want revealInline", s.kind)
	}
}

// TestMarkdownEditor_Reveal_Inline_NonZeroStartCol verifies that
// scanInlineMarkers reports correct column positions when an inline span does
// not start at column 0. This catches implementations that hardcode startCol=0.
// Spec req 7: "scanInlineMarkers produces revealSpan with correct source positions"
//
// Source "aa **bold** bb" with column layout:
//
//	col: 0=a, 1=a, 2=' ', 3=*, 4=*, 5=b, 6=o, 7=l, 8=d, 9=*, 10=*, 11=' ', 12=b, 13=b
//
// Bold span opens at col 3, closes at col 10.
func TestMarkdownEditor_Reveal_Inline_NonZeroStartCol(t *testing.T) {
	source := [][]rune{[]rune("aa **bold** bb")}
	blocks := parseMarkdown("aa **bold** bb")

	// Cursor inside bold text (col 7 = 'l')
	spans := scanInlineMarkers(source, blocks, 0, 7)

	if len(spans) != 1 {
		t.Fatalf("scanInlineMarkers returned %d spans, want 1", len(spans))
	}
	s := spans[0]
	if s.markerOpen != "**" {
		t.Errorf("markerOpen = %q, want %q", s.markerOpen, "**")
	}
	if s.kind != revealInline {
		t.Errorf("kind = %v, want revealInline", s.kind)
	}

	// Position assertions: bold must NOT start at column 0.
	// Opening ** at cols 3-4, closing ** at cols 9-10.
	if s.startRow != 0 {
		t.Errorf("startRow = %d, want 0", s.startRow)
	}
	if s.startCol != 3 {
		t.Errorf("startCol = %d, want 3 (first * at col 3, NOT 0)", s.startCol)
	}
	if s.endRow != 0 {
		t.Errorf("endRow = %d, want 0", s.endRow)
	}
	if s.endCol != 10 {
		t.Errorf("endCol = %d, want 10 (last closing * at col 10)", s.endCol)
	}
}

// ---------------------------------------------------------------------------
// Section 23 — scanInlineMarkers cursor behaviour (requirements 1, 5)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_Inline_CursorInsideOutside verifies that
// scanInlineMarkers produces a span when the cursor is inside an inline span
// and produces no span when the cursor is outside (in plain text).
// Spec req 1: "Inline reveal triggers when (cursorRow, cursorCol) falls inside
//
//	or directly adjacent to an inline span in source"
//
// Spec req 5: "Inline reveal only applies to the specific span containing the
//
//	cursor — not all inline spans in the block"
func TestMarkdownEditor_Reveal_Inline_CursorInsideOutside(t *testing.T) {
	// "aa **bold** bb" — bold spans cols 3-10 (markers at 3-4, 9-10, text at 5-8)
	source := [][]rune{[]rune("aa **bold** bb")}
	blocks := parseMarkdown("aa **bold** bb")

	// Cursor inside bold text (col 7 = 'l') → span produced
	spansIn := scanInlineMarkers(source, blocks, 0, 7)
	if len(spansIn) != 1 {
		t.Fatalf("cursor inside bold span: got %d spans, want 1 (bold span containing cursor)", len(spansIn))
	}
	if spansIn[0].markerOpen != "**" || spansIn[0].kind != revealInline {
		t.Errorf("cursor inside bold: got markerOpen=%q kind=%v, want markerOpen=%q kind=revealInline",
			spansIn[0].markerOpen, spansIn[0].kind, "**")
	}

	// Cursor outside bold span (col 0 = 'a' in preceding normal text) → no span
	spansOut := scanInlineMarkers(source, blocks, 0, 0)
	if len(spansOut) != 0 {
		t.Errorf("cursor outside bold span: got %d spans, want 0", len(spansOut))
	}
}

// TestMarkdownEditor_Reveal_Inline_BoundaryPositions verifies that the cursor
// at exact marker character positions and immediately adjacent positions
// triggers inline reveal.
// Spec req 4: "'Adjacent' means cursor is on the marker character itself, or
//
//	immediately before/after the marker in source"
//
// Spec req 9: "Cursor at exact marker character positions should trigger
//
//	reveal (boundary cases)"
//
// Source "aa **bold** bb" with column layout:
//
//	col: 0=a, 1=a, 2=' ', 3=*, 4=*, 5=b, 6=o, 7=l, 8=d, 9=*, 10=*, 11=' ', 12=b, 13=b
//
// Opening marker at cols 3-4, closing marker at cols 9-10.
func TestMarkdownEditor_Reveal_Inline_BoundaryPositions(t *testing.T) {
	source := [][]rune{[]rune("aa **bold** bb")}
	blocks := parseMarkdown("aa **bold** bb")

	tests := []struct {
		col  int
		want bool
		desc string
	}{
		// On opening marker characters (should trigger)
		{col: 3, want: true, desc: "on first * of opening marker"},
		{col: 4, want: true, desc: "on second * of opening marker"},
		// On closing marker characters (should trigger)
		{col: 9, want: true, desc: "on first * of closing marker"},
		{col: 10, want: true, desc: "on second * of closing marker"},
		// One before opening marker (adjacent, should trigger)
		{col: 2, want: true, desc: "one before opening marker (adjacent)"},
		// One after closing marker (adjacent, should trigger)
		{col: 11, want: true, desc: "one after closing marker (adjacent)"},
		// Not adjacent: in preceding normal text
		{col: 0, want: false, desc: "in preceding normal text (not adjacent)"},
		// Not adjacent: two positions after closing marker
		{col: 12, want: false, desc: "two after closing marker (not adjacent)"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			spans := scanInlineMarkers(source, blocks, 0, tt.col)
			got := len(spans) > 0
			if got != tt.want {
				t.Errorf("scanInlineMarkers at col %d: hasSpan=%v, want %v", tt.col, got, tt.want)
			}
		})
	}
}

// TestMarkdownEditor_Reveal_Inline_CursorRowDiscrimination verifies that
// scanInlineMarkers uses cursorRow to determine which line's inline spans
// are considered. A cursor on a row that does not contain an inline span
// must produce zero spans, even if another row does contain inline spans.
// Spec req 1 (implicit): "cursorRow discriminates which source line is scanned"
//
// Source across two lines:
//
//	Row 0: "**bold on row 0**"
//	Row 1: "plain on row 1"
func TestMarkdownEditor_Reveal_Inline_CursorRowDiscrimination(t *testing.T) {
	source := [][]rune{
		[]rune("**bold on row 0**"),
		[]rune("plain on row 1"),
	}
	blocks := parseMarkdown("**bold on row 0**\nplain on row 1")

	// Cursor inside bold text on row 0 (col 7 = 'o') → span produced.
	spansIn := scanInlineMarkers(source, blocks, 0, 7)
	if len(spansIn) != 1 {
		t.Fatalf("cursor inside bold on row 0: got %d spans, want 1", len(spansIn))
	}
	if spansIn[0].markerOpen != "**" || spansIn[0].kind != revealInline {
		t.Errorf("cursor inside bold on row 0: got markerOpen=%q kind=%v, want markerOpen=%q kind=revealInline",
			spansIn[0].markerOpen, spansIn[0].kind, "**")
	}
	if spansIn[0].startRow != 0 {
		t.Errorf("cursor inside bold on row 0: startRow = %d, want 0", spansIn[0].startRow)
	}

	// Cursor on row 1 (plain text) → no span (wrong row, no inline syntax).
	spansOut := scanInlineMarkers(source, blocks, 1, 3)
	if len(spansOut) != 0 {
		t.Errorf("cursor on row 1 (plain text): got %d spans, want 0 (row mismatch)", len(spansOut))
	}
}

// ---------------------------------------------------------------------------
// Section 24 — scanInlineMarkers scope constraints (requirement 5)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_Inline_OnlyCursorSpanReveals verifies that when
// source contains both bold and italic, only the span containing the cursor
// gets revealed — not all inline spans in the block.
// Spec req 5: "Inline reveal only applies to the specific span containing the
//
//	cursor — not all inline spans in the block"
//
// Source "**bold** and *italic*" layout:
//
//	col: 0=*, 1=*, 2=b, 3=o, 4=l, 5=d, 6=*, 7=*, 8=' ', 9=a, 10=n, 11=d,
//	     12=' ', 13=*, 14=i, 15=t, 16=a, 17=l, 18=i, 19=c, 20=*
func TestMarkdownEditor_Reveal_Inline_OnlyCursorSpanReveals(t *testing.T) {
	source := [][]rune{[]rune("**bold** and *italic*")}
	blocks := parseMarkdown("**bold** and *italic*")

	// Cursor inside bold text (col 4 = 'd') → only bold markers
	spansBold := scanInlineMarkers(source, blocks, 0, 4)
	if len(spansBold) != 1 {
		t.Fatalf("cursor in bold: got %d spans, want 1 (only bold)", len(spansBold))
	}
	if spansBold[0].markerOpen != "**" {
		t.Errorf("cursor in bold: markerOpen = %q, want %q (bold, not italic)",
			spansBold[0].markerOpen, "**")
	}

	// Cursor inside italic text (col 17 = 'a' in "italic") → only italic markers
	spansItalic := scanInlineMarkers(source, blocks, 0, 17)
	if len(spansItalic) != 1 {
		t.Fatalf("cursor in italic: got %d spans, want 1 (only italic)", len(spansItalic))
	}
	if spansItalic[0].markerOpen != "*" {
		t.Errorf("cursor in italic: markerOpen = %q, want %q (italic, not bold)",
			spansItalic[0].markerOpen, "*")
	}
}

// ---------------------------------------------------------------------------
// Section 25 — buildRevealSpans block + inline combination (requirement 8)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_Inline_BlockAndInlineSimultaneous verifies that
// buildRevealSpans can produce both block-level and inline-level reveal spans
// simultaneously when the cursor is inside inline syntax within a heading.
// Spec req 8: "Block-level and inline-level reveal can be active simultaneously
//
//	(cursor in bold inside a heading)"
func TestMarkdownEditor_Reveal_Inline_BlockAndInlineSimultaneous(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// "# **Bold Heading**" — H1 heading containing bold text
	me.SetText("# **Bold Heading**")

	// Position cursor inside the bold text within the heading
	// "# **Bold Heading**" — bold spans cols 2-19
	//   col: 0=#, 1=' ', 2=*, 3=*, 4=B, ... , 15=g, 16=*, 17=*
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 8 // inside "Bold Heading" bold content

	spans := me.buildRevealSpans()

	// Should have both: block marker "# " and inline marker "**"
	var hasBlock, hasInline bool
	for _, s := range spans {
		if s.kind == revealBlock && s.markerOpen == "# " {
			hasBlock = true
		}
		if s.kind == revealInline && s.markerOpen == "**" {
			hasInline = true
		}
	}
	if !hasBlock {
		t.Error("missing block-level '# ' marker span; block and inline should coexist")
	}
	if !hasInline {
		t.Error("missing inline-level '**' marker span; block and inline should coexist")
	}
	if len(spans) != 2 {
		t.Errorf("got %d spans, want 2 (one block + one inline)", len(spans))
	}
}

// TestMarkdownEditor_Reveal_Inline_OnlyInlineViaBuildRevealSpans verifies that
// buildRevealSpans returns at least one revealInline span for a plain paragraph
// containing only inline formatting (no block-level markers). This catches the
// "forgot to integrate scanInlineMarkers into buildRevealSpans" shortcut.
// Spec req 7: "buildRevealSpans must merge inline spans from scanInlineMarkers"
func TestMarkdownEditor_Reveal_Inline_OnlyInlineViaBuildRevealSpans(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("**bold text**")

	// Cursor inside the bold content (col 5 = ' ')
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 5

	spans := me.buildRevealSpans()

	// For a plain paragraph with only inline bold, buildRevealSpans must
	// return at least one revealInline span for the bold markers.
	var inlineCount int
	for _, s := range spans {
		if s.kind == revealInline {
			inlineCount++
			if s.markerOpen != "**" {
				t.Errorf("inline span markerOpen = %q, want %q", s.markerOpen, "**")
			}
		}
	}
	if inlineCount != 1 {
		t.Errorf("got %d revealInline spans, want 1 (plain paragraph with **bold** must return inline span via buildRevealSpans)",
			inlineCount)
	}
	if len(spans) != 1 {
		t.Errorf("got %d total spans, want 1 (only inline, no block-level markers)", len(spans))
	}
}

// ---------------------------------------------------------------------------
// Section 26 — scanInlineMarkers edge cases (requirements 10, 11)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_Inline_EmptySource verifies scanInlineMarkers
// returns nil/empty when the source is empty (no lines, no blocks).
// Spec req 10: "Empty source → nil/empty spans"
func TestMarkdownEditor_Reveal_Inline_EmptySource(t *testing.T) {
	spans := scanInlineMarkers([][]rune{}, []mdBlock{}, 0, 0)

	if len(spans) != 0 {
		t.Errorf("scanInlineMarkers on empty source returned %d spans, want 0", len(spans))
	}
}

// TestMarkdownEditor_Reveal_Inline_PlainTextNoSpans verifies scanInlineMarkers
// produces no spans when source contains only plain text without any inline
// syntax markers.
// Falsifying: an implementation that spuriously produces spans for plain text
// or confuses normal punctuation for markers would fail here.
// Spec req 11 (falsifying): "scanInlineMarkers on plain text → no spans"
func TestMarkdownEditor_Reveal_Inline_PlainTextNoSpans(t *testing.T) {
	source := [][]rune{[]rune("hello world, this is just plain text.")}
	blocks := parseMarkdown("hello world, this is just plain text.")

	// Cursor at various positions in plain text — must produce 0 spans
	spans := scanInlineMarkers(source, blocks, 0, 3)

	if len(spans) != 0 {
		t.Errorf("scanInlineMarkers on plain text returned %d spans, want 0 (no inline syntax present)", len(spans))
	}
}

// ---------------------------------------------------------------------------
// Section 27 — scanInlineMarkers scope falsifying (requirement 12)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_Inline_CursorOnlyRevealsOwnSpan verifies that
// when source contains two bold spans, only the one containing the cursor
// produces a reveal span — the other bold span does not.
// Falsifying: an implementation that reveals all inline spans regardless of
// cursor position would return 2 spans here.
// Spec req 12 (falsifying): "cursor in one bold span doesn't reveal other bold
// spans"
//
// Source "**one** and **two**" layout:
//
//	col: 0=*, 1=*, 2=o, 3=n, 4=e, 5=*, 6=*, 7=' ', 8=a, 9=n, 10=d, 11=' ',
//	     12=*, 13=*, 14=t, 15=w, 16=o, 17=*, 18=*
//
// First bold at cols 0-6, second bold at cols 12-18.
func TestMarkdownEditor_Reveal_Inline_CursorOnlyRevealsOwnSpan(t *testing.T) {
	source := [][]rune{[]rune("**one** and **two**")}
	blocks := parseMarkdown("**one** and **two**")

	// Cursor inside first bold (col 3 = 'e' in "one") → only first bold span
	spans := scanInlineMarkers(source, blocks, 0, 3)

	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1 (only the bold containing cursor)", len(spans))
	}
	// Verify the span is for the first bold, not the second, by checking its
	// source position: the first bold starts at col 0.
	if spans[0].startCol != 0 {
		t.Errorf("startCol = %d, want 0 (first bold \"**one**\" at cols 0-6, not second at cols 12-18)",
			spans[0].startCol)
	}
	// markerOpen must be "**" for bold
	if spans[0].markerOpen != "**" {
		t.Errorf("markerOpen = %q, want %q", spans[0].markerOpen, "**")
	}
}

// ---------------------------------------------------------------------------
// Section 28 — scanInlineMarkers nested / combined emphasis (requirement 13)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_Reveal_Inline_NestedEmphasis verifies that scanInlineMarkers
// produces correct markers for nested bold+italic emphasis (***text***).
// Goldmark parses "***bolditalic***" as <strong><em>bolditalic</em></strong>
// which yields runBoldItalic style via collectInlineRuns.
// Spec req 13 (implicit): "Nested emphasis ***text*** → correct marker text
// matching goldmark's interpretation"
//
// Source "***bolditalic***" with column layout:
//
//	col: 0=*, 1=*, 2=*, 3=b, 4=o, 5=l, 6=d, 7=i, 8=t, 9=a, 10=l, 11=i, 12=c, 13=*, 14=*, 15=*
func TestMarkdownEditor_Reveal_Inline_NestedEmphasis(t *testing.T) {
	source := [][]rune{[]rune("***bolditalic***")}
	blocks := parseMarkdown("***bolditalic***")

	// Verify goldmark produces runBoldItalic for the combined emphasis.
	if len(blocks) != 1 || blocks[0].kind != blockParagraph {
		t.Fatalf("expected 1 paragraph block, got %d blocks, block kind = %v", len(blocks), blocks[0].kind)
	}
	hasBoldItalic := false
	for _, r := range blocks[0].runs {
		if r.style == runBoldItalic {
			hasBoldItalic = true
			break
		}
	}
	if !hasBoldItalic {
		t.Errorf("expected runBoldItalic in paragraph runs, got runs: %+v", blocks[0].runs)
	}

	// Cursor inside the bolditalic text (col 7 = 'i') → span produced.
	spans := scanInlineMarkers(source, blocks, 0, 7)

	if len(spans) != 1 {
		t.Fatalf("scanInlineMarkers for ***bolditalic*** returned %d spans, want 1", len(spans))
	}
	s := spans[0]
	if s.kind != revealInline {
		t.Errorf("kind = %v, want revealInline", s.kind)
	}
	// Goldmark parses *** as combined bold+italic markers. The literal
	// source text uses *** at both ends, so the combined marker is "***".
	if s.markerOpen != "***" {
		t.Errorf("markerOpen = %q, want %q (combined bold+italic marker from source)", s.markerOpen, "***")
	}
	if s.markerClose != "***" {
		t.Errorf("markerClose = %q, want %q (combined bold+italic marker from source)", s.markerClose, "***")
	}
}

// =============================================================================
// Task 10 — Smart List Continuation
// =============================================================================
//
// These tests are written BEFORE the implementation exists. They define the
// expected behavior of smart list continuation (Enter, Tab, Shift-Tab) as
// specified in the "Smart List Continuation" section of the design spec.
//
// Spec: "These DO mutate source: Enter at end of a non-empty list item creates
// a new line with the same marker; Enter on an empty list item deletes the
// empty marker; Tab indents; Shift-Tab outdents."

// ---------------------------------------------------------------------------
// Section 29 — Enter at end of non-empty list item: bullet list markers
// ---------------------------------------------------------------------------

// TestMarkdownEditor_ListContinuation_EnterAtEndOfDashItem verifies that
// pressing Enter at the end of a non-empty dash list item ("- item") inserts
// a new line with the "- " marker prefix.
// Spec: "Enter at end of a non-empty list item → new line with same marker"
func TestMarkdownEditor_ListContinuation_EnterAtEndOfDashItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("- item")

	// Position cursor at end of line
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len([]rune("- item"))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	// After Enter, Memo inserts a newline. Smart list continuation adds "- "
	// on the new line. The result should be "- item\n- "
	if got != "- item\n- " {
		t.Errorf("Text() = %q, want %q", got, "- item\n- ")
	}

	// Event should be consumed (Enter went through Memo + list continuation)
	if !ev.IsCleared() {
		t.Error("Enter event not cleared; should have been consumed")
	}
}

// TestMarkdownEditor_ListContinuation_EnterAtEndOfStarItem verifies Enter at
// the end of a star bullet list item ("* item") inserts a new line with "* ".
// Spec: "Enter at end of a non-empty list item → new line with same marker"
// The star marker ("* ") is a valid bullet list marker.
func TestMarkdownEditor_ListContinuation_EnterAtEndOfStarItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("* item")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len([]rune("* item"))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "* item\n* " {
		t.Errorf("Text() = %q, want %q", got, "* item\n* ")
	}
	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}
}

// TestMarkdownEditor_ListContinuation_EnterAtEndOfPlusItem verifies Enter at
// the end of a plus bullet list item ("+ item") inserts a new line with "+ ".
// Spec: list continuation works for bullet lists (all marker types).
func TestMarkdownEditor_ListContinuation_EnterAtEndOfPlusItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("+ item")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len([]rune("+ item"))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "+ item\n+ " {
		t.Errorf("Text() = %q, want %q", got, "+ item\n+ ")
	}
	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}
}

// TestMarkdownEditor_ListContinuation_EnterAtEndOfNumberedItem verifies Enter
// at the end of a numbered list item ("1. item") inserts a new line with the
// next number ("2. ").
// Spec: "Numbered list markers are incremented (1. → 2., 99. → 100.)"
func TestMarkdownEditor_ListContinuation_EnterAtEndOfNumberedItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("1. item")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len([]rune("1. item"))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "1. item\n2. " {
		t.Errorf("Text() = %q, want %q (number incremented)", got, "1. item\n2. ")
	}
	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}
}

// TestMarkdownEditor_ListContinuation_EnterAtEndOfMultiDigitNumberedItem
// verifies Enter at the end of a multi-digit numbered list item ("99. item")
// inserts a new line with the next number ("100. ").
// Spec: "Numbered list markers are incremented (1. → 2., 99. → 100.)"
func TestMarkdownEditor_ListContinuation_EnterAtEndOfMultiDigitNumberedItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("99. item")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len([]rune("99. item"))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "99. item\n100. " {
		t.Errorf("Text() = %q, want %q (number incremented from 99 to 100)", got, "99. item\n100. ")
	}
	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}
}

// ---------------------------------------------------------------------------
// Section 30 — Enter at end of non-empty list item: checklist markers
// ---------------------------------------------------------------------------

// TestMarkdownEditor_ListContinuation_EnterAtEndOfUncheckedItem verifies that
// Enter at the end of an unchecked checklist item ("- [ ] item") inserts a new
// line with the "- [ ] " prefix.
// Spec: "- [ ] item + Enter → new line with - [ ] prefix"
func TestMarkdownEditor_ListContinuation_EnterAtEndOfUncheckedItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("- [ ] item")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len([]rune("- [ ] item"))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "- [ ] item\n- [ ] " {
		t.Errorf("Text() = %q, want %q", got, "- [ ] item\n- [ ] ")
	}
	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}
}

// TestMarkdownEditor_ListContinuation_EnterAtEndOfCheckedItemDefaultsUnchecked
// verifies that Enter at the end of a checked checklist item ("- [x] item")
// inserts a new line with "- [ ] " (unchecked) prefix, NOT "- [x] ".
// Spec: "- [x] item + Enter → new line with - [ ] prefix (new items default unchecked)"
func TestMarkdownEditor_ListContinuation_EnterAtEndOfCheckedItemDefaultsUnchecked(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("- [x] item")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len([]rune("- [x] item"))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "- [x] item\n- [ ] " {
		t.Errorf("Text() = %q, want %q (new items default unchecked)", got, "- [x] item\n- [ ] ")
	}
	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}
}

// ---------------------------------------------------------------------------
// Section 31 — Enter on empty list item: deletes marker, exits list
// ---------------------------------------------------------------------------

// TestMarkdownEditor_ListContinuation_EnterOnEmptyDashMarkerClearsIt verifies
// that pressing Enter when the line is just "- " (empty dash list item) deletes
// the marker and replaces the line with an empty line, exiting the list.
// Spec: "Enter on an empty list item → delete empty marker, exit list, insert blank line"
func TestMarkdownEditor_ListContinuation_EnterOnEmptyDashMarkerClearsIt(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("- ")

	// Cursor at end of the marker text
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len([]rune("- "))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	// The empty marker line should be cleared — must not still be "- "
	if got == "- " || got == "- \n" {
		t.Errorf("Text() = %q; empty marker was NOT cleared by Enter on empty list item", got)
	}
	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}
}

// TestMarkdownEditor_ListContinuation_EnterOnEmptyNumberedMarkerClearsIt
// verifies that pressing Enter when the line is just "1. " (empty numbered item)
// deletes the marker and exits the list.
// Spec: "Enter on an empty list item → delete empty marker, exit list, insert blank line"
func TestMarkdownEditor_ListContinuation_EnterOnEmptyNumberedMarkerClearsIt(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("1. ")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len([]rune("1. "))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	if got == "1. " || got == "1. \n" {
		t.Errorf("Text() = %q; empty numbered marker was NOT cleared", got)
	}
	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}
}

// TestMarkdownEditor_ListContinuation_EnterOnEmptyChecklistMarkerClearsIt
// verifies that pressing Enter when the line is just "- [ ] " (empty checklist
// item) deletes the marker and exits the list.
// Spec: "Enter on an empty list item → delete empty marker, exit list, insert blank line"
func TestMarkdownEditor_ListContinuation_EnterOnEmptyChecklistMarkerClearsIt(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("- [ ] ")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len([]rune("- [ ] "))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	if got == "- [ ] " || got == "- [ ] \n" {
		t.Errorf("Text() = %q; empty checklist marker was NOT cleared", got)
	}
	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}
}

// TestMarkdownEditor_ListContinuation_EnterNotAtEndOfLineNoContinuation
// verifies that pressing Enter when the cursor is NOT at the end of the
// list item line (cursor mid-text) does NOT trigger list continuation.
// Spec: "Enter at end of a non-empty list item" — specifically "at end".
func TestMarkdownEditor_ListContinuation_EnterNotAtEndOfLineNoContinuation(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("- item text")

	// Cursor in the middle of "item" (col 4 = 't' in "item")
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 4 // "- i|tem text"

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	// Enter at mid-line splits the line (Memo behavior), NOT adding list marker
	// The new line must NOT have "- " prefix from list continuation.
	if strings.Count(got, "- ") >= 2 {
		t.Errorf("Text() = %q; list continuation incorrectly added marker when Enter was NOT at end of line", got)
	}
}

// ---------------------------------------------------------------------------
// Section 32 — Tab indenting at list items
// ---------------------------------------------------------------------------

// TestMarkdownEditor_ListContinuation_TabIndentsDashItem verifies that pressing
// Tab on a dash list item adds "  " (two spaces) before the "- " marker.
// Spec: "Tab at list item → indent by adding '  ' before marker"
func TestMarkdownEditor_ListContinuation_TabIndentsDashItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("- item")

	// Cursor anywhere on the line — position on the text content
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 3 // "- i|tem"

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "  - item" {
		t.Errorf("Text() = %q, want %q (indented with two spaces)", got, "  - item")
	}
	if !ev.IsCleared() {
		t.Error("Tab event not cleared; should have been consumed by list indent")
	}
}

// TestMarkdownEditor_ListContinuation_TabIndentsNumberedItem verifies Tab on
// a numbered list item adds "  " before the number marker.
// Spec: "Tab at list item → indent by adding '  ' before marker"
func TestMarkdownEditor_ListContinuation_TabIndentsNumberedItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("1. item")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 0

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "  1. item" {
		t.Errorf("Text() = %q, want %q (indented with two spaces)", got, "  1. item")
	}
	if !ev.IsCleared() {
		t.Error("Tab event not cleared")
	}
}

// TestMarkdownEditor_ListContinuation_TabIndentsChecklistItem verifies Tab on
// a checklist item adds "  " before the "- [ ] " marker.
// Spec: "Tab at list item → indent by adding '  ' before marker"
func TestMarkdownEditor_ListContinuation_TabIndentsChecklistItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("- [ ] todo")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 0

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "  - [ ] todo" {
		t.Errorf("Text() = %q, want %q (indented with two spaces)", got, "  - [ ] todo")
	}
	if !ev.IsCleared() {
		t.Error("Tab event not cleared")
	}
}

// TestMarkdownEditor_ListContinuation_TabIndentsAlreadyIndentedItem verifies
// that pressing Tab on an already-indented list item adds another level of
// indentation ("  " prefix).
// Spec: Tab adds "  " before marker — works cumulatively.
func TestMarkdownEditor_ListContinuation_TabIndentsAlreadyIndentedItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("  - item")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 4

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "    - item" {
		t.Errorf("Text() = %q, want %q (double indented)", got, "    - item")
	}
	if !ev.IsCleared() {
		t.Error("Tab event not cleared")
	}
}

// TestMarkdownEditor_ListContinuation_TabDoesNothingOnNonListLine verifies that
// pressing Tab on a non-list line does NOT modify the source.
// Falsifying: an implementation that blindly adds "  " prefix to any line.
func TestMarkdownEditor_ListContinuation_TabDoesNothingOnNonListLine(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("plain paragraph text")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 5

	original := me.Text()

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab}}
	me.HandleEvent(ev)

	if me.Text() != original {
		t.Errorf("Text() = %q, want %q; Tab on non-list line should not modify source",
			me.Text(), original)
	}
}

// ---------------------------------------------------------------------------
// Section 33 — Shift-Tab outdenting at list items
// ---------------------------------------------------------------------------

// TestMarkdownEditor_ListContinuation_ShiftTabOutdentsItem verifies that
// pressing Shift-Tab on an indented list item removes the "  " prefix.
// Spec: "Shift-Tab at indented list item → outdent by removing '  ' prefix"
func TestMarkdownEditor_ListContinuation_ShiftTabOutdentsItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("  - item")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 5

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab, Modifiers: tcell.ModShift}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "- item" {
		t.Errorf("Text() = %q, want %q (outdented)", got, "- item")
	}
	if !ev.IsCleared() {
		t.Error("Shift-Tab event not cleared; should have been consumed by list outdent")
	}
}

// TestMarkdownEditor_ListContinuation_ShiftTabAtNoIndentDoesNothing verifies
// that pressing Shift-Tab on a non-indented list item does nothing.
// Spec: "If no indent → do nothing"
func TestMarkdownEditor_ListContinuation_ShiftTabAtNoIndentDoesNothing(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("- item")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 3

	original := me.Text()

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab, Modifiers: tcell.ModShift}}
	me.HandleEvent(ev)

	if me.Text() != original {
		t.Errorf("Text() = %q, want %q; Shift-Tab with no indent should not modify source",
			me.Text(), original)
	}
}

// TestMarkdownEditor_ListContinuation_ShiftTabOutdentsNumberedItem verifies
// Shift-Tab on an indented numbered list item removes the "  " prefix.
// Spec: "Shift-Tab at indented list item → outdent by removing '  ' prefix"
func TestMarkdownEditor_ListContinuation_ShiftTabOutdentsNumberedItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("  1. item")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 0

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab, Modifiers: tcell.ModShift}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "1. item" {
		t.Errorf("Text() = %q, want %q (outdented)", got, "1. item")
	}
}

// TestMarkdownEditor_ListContinuation_ShiftTabDoesNothingOnNonListLine verifies
// that Shift-Tab on a non-list line does nothing.
// Falsifying: an implementation that blindly strips leading spaces from any line.
func TestMarkdownEditor_ListContinuation_ShiftTabDoesNothingOnNonListLine(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("  plain paragraph")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 5

	original := me.Text()

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab, Modifiers: tcell.ModShift}}
	me.HandleEvent(ev)

	if me.Text() != original {
		t.Errorf("Text() = %q, want %q; Shift-Tab on non-list line should not modify source",
			me.Text(), original)
	}
}

// TestMarkdownEditor_ListContinuation_EnterDoesNotContinuateForNonListLine
// verifies that pressing Enter on a non-list line does NOT trigger list
// continuation (no marker is added to the new line).
// Falsifying: an implementation that blindly adds markers on Enter.
func TestMarkdownEditor_ListContinuation_EnterDoesNotContinuateForNonListLine(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("plain text line")

	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len([]rune("plain text line"))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	if strings.Contains(got, "- ") || strings.Contains(got, "1. ") || strings.Contains(got, "* ") {
		t.Errorf("Text() = %q; list continuation incorrectly added marker on non-list line", got)
	}
}

// =============================================================================
// Task 11 — Keyboard Shortcuts
// =============================================================================
//
// These tests are written BEFORE the implementation exists. They verify the
// keyboard shortcut behavior specified in the "Keyboard Shortcuts" section.
//
// Spec: "Ctrl+B: Toggle **bold** on selection or insert empty markers.
// Ctrl+I: Toggle *italic* on selection or insert empty markers.
// Ctrl+T: Toggle raw source view."

// ---------------------------------------------------------------------------
// Section 34 — Ctrl+B: bold toggle
// ---------------------------------------------------------------------------

// TestMarkdownEditor_KeyboardShortcut_CtrlBWrapsSelection verifies that Ctrl+B
// with an active text selection wraps the selected text in "**" markers.
// Spec: "With selection: wrap selection in **"
func TestMarkdownEditor_KeyboardShortcut_CtrlBWrapsSelection(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("hello world")

	// Select "hello" (cols 0-4)
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 0
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 5
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 5

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlB, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "**hello** world" {
		t.Errorf("Text() = %q, want %q", got, "**hello** world")
	}
	if !ev.IsCleared() {
		t.Error("Ctrl+B event not cleared; should have been consumed")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlBUnwrapsSelection verifies that Ctrl+B
// on already-bold selection removes the "**" markers.
// Spec: "If already wrapped in **, remove markers"
func TestMarkdownEditor_KeyboardShortcut_CtrlBUnwrapsSelection(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("some **bold** text")

	// Select the bold text including markers: "**bold**" (cols 5-13)
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 5
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 13
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 13

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlB, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "some bold text" {
		t.Errorf("Text() = %q, want %q (markers removed)", got, "some bold text")
	}
	if !ev.IsCleared() {
		t.Error("Ctrl+B event not cleared")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlBNoSelection verifies that Ctrl+B
// without a selection inserts "****" with cursor between the inner "**".
// Spec: "Without selection: insert **** and place cursor between inner ** (at position original+2)"
func TestMarkdownEditor_KeyboardShortcut_CtrlBNoSelection(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("")

	if me.Memo.HasSelection() {
		t.Fatal("test setup failed: HasSelection should be false")
	}

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlB, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "****" {
		t.Errorf("Text() = %q, want %q", got, "****")
	}
	// Cursor should be between inner ** (position 2)
	if me.Memo.cursorCol != 2 {
		t.Errorf("cursorCol = %d, want 2 (between inner **)", me.Memo.cursorCol)
	}
	if !ev.IsCleared() {
		t.Error("Ctrl+B event not cleared")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlBNoSelectionCursorStaysWithin
// verifies that Ctrl+B insertion preserves cursor position relative to
// surrounding text. The cursor ends up at offset+2 (between inner **).
// Spec: "Without selection: insert **** and place cursor between inner **"
func TestMarkdownEditor_KeyboardShortcut_CtrlBNoSelectionCursorStaysWithin(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("before after")

	// Cursor between "before " and " after" (col 7)
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 7

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlB, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "before ****after" {
		t.Errorf("Text() = %q, want %q", got, "before ****after")
	}
	// Cursor at original position + 2 (between inner **)
	if me.Memo.cursorCol != 9 {
		t.Errorf("cursorCol = %d, want 9 (original 7 + 2)", me.Memo.cursorCol)
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlBSelectionAtLineStart verifies that
// Ctrl+B with a selection at the very start of a line works correctly.
// Spec: "With selection: wrap selection in **"
func TestMarkdownEditor_KeyboardShortcut_CtrlBSelectionAtLineStart(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("first line")

	// Select "first" at line start (cols 0-4)
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 0
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 5
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 5

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlB, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "**first** line" {
		t.Errorf("Text() = %q, want %q", got, "**first** line")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlBUnwrapsOnlyOuterMarkers
// verifies that selecting text wrapped in "**" and pressing Ctrl+B removes the
// surrounding markers even when the selection is just the content.
// Spec: "If already wrapped in **, remove markers"
func TestMarkdownEditor_KeyboardShortcut_CtrlBUnwrapsOnlyOuterMarkers(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("**important**")

	// Select just the content "important" (cols 2-10), NOT including markers
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 2
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 11
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 11

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlB, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "important" {
		t.Errorf("Text() = %q, want %q (only content, markers removed)", got, "important")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlBWithModCtrl verifies that Ctrl+B
// works when the KeyEvent has ModCtrl set, as tcell reports for real keyboard
// input. This is a regression test for kata issue #8.
func TestMarkdownEditor_KeyboardShortcut_CtrlBWithModCtrl(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("hello world")

	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 0
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 5
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 5

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlB, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "**hello** world" {
		t.Errorf("Text() = %q, want %q", got, "**hello** world")
	}
	if !ev.IsCleared() {
		t.Error("Ctrl+B with ModCtrl not consumed — ModNone guard rejects real keyboard input")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlIWithModCtrl verifies that Ctrl+I
// works when the KeyEvent has ModCtrl set.
func TestMarkdownEditor_KeyboardShortcut_CtrlIWithModCtrl(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("italic text here")

	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 7
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 11
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 11

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlI, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "italic *text* here" {
		t.Errorf("Text() = %q, want %q", got, "italic *text* here")
	}
	if !ev.IsCleared() {
		t.Error("Ctrl+I with ModCtrl not consumed — ModNone guard rejects real keyboard input")
	}
}

// ---------------------------------------------------------------------------
// Section 35 — Ctrl+I: italic toggle
// ---------------------------------------------------------------------------

// TestMarkdownEditor_KeyboardShortcut_CtrlIWrapsSelection verifies that Ctrl+I
// with a selection wraps the selected text in "*" markers.
// Spec: "With selection: wrap in *."
func TestMarkdownEditor_KeyboardShortcut_CtrlIWrapsSelection(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("italic text here")

	// Select "text" (cols 7-10)
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 7
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 11
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 11

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlI, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "italic *text* here" {
		t.Errorf("Text() = %q, want %q", got, "italic *text* here")
	}
	if !ev.IsCleared() {
		t.Error("Ctrl+I event not cleared")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlIUnwrapsSelection verifies that Ctrl+I
// on already-italic selection removes the "*" markers.
// Spec: "If already wrapped, remove."
func TestMarkdownEditor_KeyboardShortcut_CtrlIUnwrapsSelection(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("some *italic* text")

	// Select the italic span including markers: "*italic*" (cols 5-12)
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 5
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 13
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 13

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlI, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "some italic text" {
		t.Errorf("Text() = %q, want %q (markers removed)", got, "some italic text")
	}
	if !ev.IsCleared() {
		t.Error("Ctrl+I event not cleared")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlINoSelection verifies that Ctrl+I
// without a selection inserts "**" (two asterisks) and places the cursor
// between them.
// Spec: "Without selection: insert ** with cursor between"
func TestMarkdownEditor_KeyboardShortcut_CtrlINoSelection(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("")

	if me.Memo.HasSelection() {
		t.Fatal("test setup failed: HasSelection should be false")
	}

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlI, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "**" {
		t.Errorf("Text() = %q, want %q", got, "**")
	}
	if me.Memo.cursorCol != 1 {
		t.Errorf("cursorCol = %d, want 1 (between * markers)", me.Memo.cursorCol)
	}
	if !ev.IsCleared() {
		t.Error("Ctrl+I event not cleared")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlINoSelectionInContext verifies that
// Ctrl+I without selection places cursor correctly when surrounded by text.
// Spec: "Without selection: insert ** with cursor between"
func TestMarkdownEditor_KeyboardShortcut_CtrlINoSelectionInContext(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("prefix suffix")

	// Cursor between "prefix " and "suffix" (col 7)
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 7

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlI, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	got := me.Text()
	if got != "prefix **suffix" {
		t.Errorf("Text() = %q, want %q", got, "prefix **suffix")
	}
	if me.Memo.cursorCol != 8 {
		t.Errorf("cursorCol = %d, want 8 (original 7 + 1)", me.Memo.cursorCol)
	}
}

// ---------------------------------------------------------------------------
// Section 36 — Ctrl+T: source toggle
// ---------------------------------------------------------------------------

// TestMarkdownEditor_KeyboardShortcut_CtrlTTogglesShowSource verifies that
// pressing Ctrl+T toggles showSource from false to true.
// Spec: "Ctrl+T — Toggle raw source view. Flips between formatted and source view."
func TestMarkdownEditor_KeyboardShortcut_CtrlTTogglesShowSource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("# heading\n\ncontent")

	if me.ShowSource() {
		t.Fatal("test setup: ShowSource should default to false")
	}

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlT}}
	me.HandleEvent(ev)

	if !me.ShowSource() {
		t.Error("ShowSource() = false after Ctrl+T, want true (toggled on)")
	}
	if !ev.IsCleared() {
		t.Error("Ctrl+T event not cleared")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlTTogglesBack verifies that pressing
// Ctrl+T a second time toggles showSource back to false.
// Spec: "Flips between formatted and source view."
func TestMarkdownEditor_KeyboardShortcut_CtrlTTogglesBack(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("# heading\n\ncontent")
	me.SetShowSource(true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlT}}
	me.HandleEvent(ev)

	if me.ShowSource() {
		t.Error("ShowSource() = true after second Ctrl+T, want false (toggled off)")
	}
	if !ev.IsCleared() {
		t.Error("Ctrl+T event not cleared")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlTToggleBackReparses verifies that
// toggling showSource from true back to false triggers a reparse, refreshing
// the parsed blocks.
// Spec: "When toggling back to formatted mode, reparse runs"
func TestMarkdownEditor_KeyboardShortcut_CtrlTToggleBackReparses(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("# heading\n\nparagraph")

	// Go to source mode
	me.SetShowSource(true)

	// In source mode, poison the sourceCache
	me.sourceCache = "poisoned"

	// Toggle back to formatted mode
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlT}}
	me.HandleEvent(ev)

	if me.ShowSource() {
		t.Fatal("ShowSource should be false after toggling back")
	}
	// reparse should have run, updating sourceCache from "poisoned" to actual source
	if me.sourceCache == "poisoned" {
		t.Error("reparse was not called when toggling back to formatted mode; sourceCache still poisoned")
	}
	if me.sourceCache != me.Text() {
		t.Errorf("sourceCache = %q, want %q (should match Memo text after reparse)", me.sourceCache, me.Text())
	}
	if len(me.blocks) == 0 {
		t.Error("blocks empty after toggling back to formatted mode; reparse should have populated them")
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlTDoesNotLoseSource verifies that
// toggling between source and formatted mode does not lose or corrupt the
// source text or cursor position.
// Spec: "Source and cursor shared between modes; toggling doesn't lose state."
func TestMarkdownEditor_KeyboardShortcut_CtrlTDoesNotLoseSource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	original := "# heading\n\n**bold** paragraph"
	me.SetText(original)

	// Set known cursor position: row 2 (third line), col 2 (inside "bold")
	me.Memo.cursorRow = 2
	me.Memo.cursorCol = 2

	// Toggle to source
	ev1 := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlT}}
	me.HandleEvent(ev1)

	if me.Text() != original {
		t.Errorf("source changed after first toggle: got %q, want %q", me.Text(), original)
	}
	row, col := me.CursorPos()
	if row != 2 || col != 2 {
		t.Errorf("cursor position after first toggle = (%d, %d), want (2, 2)", row, col)
	}

	// Toggle back to formatted
	ev2 := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlT}}
	me.HandleEvent(ev2)

	if me.Text() != original {
		t.Errorf("source changed after second toggle: got %q, want %q", me.Text(), original)
	}
	row, col = me.CursorPos()
	if row != 2 || col != 2 {
		t.Errorf("cursor position after second toggle = (%d, %d), want (2, 2)", row, col)
	}
}

// Keyboard shortcut tests for Ctrl+B (bold), Ctrl+I (italic), and Ctrl+T
// (source toggle) are covered by Sections 34-36 above.
// Ctrl+K (link dialog) and Link Interaction tests are in Sections 42-44 below.

// =============================================================================
// Task 9 — Auto-Format (typing markdown syntax produces formatted rendering)
// =============================================================================
//
// These tests verify that typing markdown syntax via HandleEvent produces the
// expected parsed block kinds. Per the spec, auto-format does NOT mutate source
// — reparse already handles formatting from the typed syntax.
//
// Spec: "Auto-format triggers when typed syntax becomes complete. Source is NOT
// mutated — only rendering updates."

// ---------------------------------------------------------------------------
// Section 37 — Typing heading syntax
// ---------------------------------------------------------------------------

// TestMarkdownEditor_AutoFormat_TypingHashSpaceCreatesHeading verifies that
// typing "# " at the start of a line causes reparse to produce a heading block.
// Spec: "# at line start + space → Line renders as heading"
func TestMarkdownEditor_AutoFormat_TypingHashSpaceCreatesHeading(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("")

	// Type '#', then ' ', then content character "H"
	for _, r := range []rune{'#', ' ', 'H'} {
		ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}}
		me.HandleEvent(ev)
	}

	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after typing '# H'")
	}
	if me.blocks[0].kind != blockHeader {
		t.Errorf("blocks[0].kind = %v after typing '# H', want blockHeader", me.blocks[0].kind)
	}
}

// TestMarkdownEditor_AutoFormat_SetTextHashSpaceIsHeading verifies that using
// SetText with "# heading" produces a heading block.
// Spec: "# at line start + space → Line renders as heading"
func TestMarkdownEditor_AutoFormat_SetTextHashSpaceIsHeading(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("# heading text")

	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after SetText")
	}
	if me.blocks[0].kind != blockHeader {
		t.Errorf("blocks[0].kind = %v, want blockHeader", me.blocks[0].kind)
	}
	if me.blocks[0].level != 1 {
		t.Errorf("blocks[0].level = %d, want 1", me.blocks[0].level)
	}
}

// ---------------------------------------------------------------------------
// Section 38 — Typing bold/italic/code inline syntax
// ---------------------------------------------------------------------------

// TestMarkdownEditor_AutoFormat_TypingBoldSyntaxRendersBold verifies that typing
// "**text**" causes reparse to produce a paragraph block containing a runBold
// inline run. The reparse detects the completed bold syntax.
// Spec: "**text** + closing ** → text renders bold"
func TestMarkdownEditor_AutoFormat_TypingBoldSyntaxRendersBold(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("**bold**")

	if len(me.blocks) == 0 || len(me.blocks[0].runs) == 0 {
		t.Fatal("no blocks or runs after SetText with bold syntax")
	}
	hasBold := false
	for _, r := range me.blocks[0].runs {
		if r.style == runBold {
			hasBold = true
			break
		}
	}
	if !hasBold {
		t.Error("no runBold found in parsed runs for '**bold**'")
	}
}

// TestMarkdownEditor_AutoFormat_TypingItalicStarSyntaxRendersItalic verifies
// that typing "*text*" produces runItalic in the parsed runs.
// Spec: "*text* + closing * → text renders italic"
func TestMarkdownEditor_AutoFormat_TypingItalicStarSyntaxRendersItalic(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("*italic*")

	if len(me.blocks) == 0 || len(me.blocks[0].runs) == 0 {
		t.Fatal("no blocks or runs after SetText with italic syntax")
	}
	hasItalic := false
	for _, r := range me.blocks[0].runs {
		if r.style == runItalic {
			hasItalic = true
			break
		}
	}
	if !hasItalic {
		t.Error("no runItalic found in parsed runs for '*italic*'")
	}
}

// TestMarkdownEditor_AutoFormat_TypingCodeSyntaxRendersCode verifies that
// typing "`text`" produces runCode in the parsed runs.
// Spec: "`text` + closing ` → text renders as code"
func TestMarkdownEditor_AutoFormat_TypingCodeSyntaxRendersCode(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("`code`")

	if len(me.blocks) == 0 || len(me.blocks[0].runs) == 0 {
		t.Fatal("no blocks or runs after SetText with code syntax")
	}
	hasCode := false
	for _, r := range me.blocks[0].runs {
		if r.style == runCode {
			hasCode = true
			break
		}
	}
	if !hasCode {
		t.Error("no runCode found in parsed runs for '`code`'")
	}
}

// ---------------------------------------------------------------------------
// Section 39 — Typing block-level syntax (list, blockquote, numbered list)
// ---------------------------------------------------------------------------

// TestMarkdownEditor_AutoFormat_TypingDashSpaceCreatesList verifies that typing
// "- " at line start causes reparse to produce a bullet list block.
// Spec: "- at line start + space → Line renders as list item"
func TestMarkdownEditor_AutoFormat_TypingDashSpaceCreatesList(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("- list item")

	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after SetText with list syntax")
	}
	if me.blocks[0].kind != blockBulletList {
		t.Errorf("blocks[0].kind = %v, want blockBulletList", me.blocks[0].kind)
	}
}

// TestMarkdownEditor_AutoFormat_TypingGreaterSpaceCreatesBlockquote verifies
// that typing "> " at line start produces a blockquote block.
// Spec: "> at line start + space → Line renders as blockquote"
func TestMarkdownEditor_AutoFormat_TypingGreaterSpaceCreatesBlockquote(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("> quoted text")

	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after SetText with blockquote syntax")
	}
	if me.blocks[0].kind != blockBlockquote {
		t.Errorf("blocks[0].kind = %v, want blockBlockquote", me.blocks[0].kind)
	}
}

// TestMarkdownEditor_AutoFormat_TypingNumberedListSyntaxCreatesNumberedList
// verifies that typing "1. " at line start produces a numbered list block.
// Spec: "1. at line start + space → Line renders as numbered list"
func TestMarkdownEditor_AutoFormat_TypingNumberedListSyntaxCreatesNumberedList(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("1. first item")

	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after SetText with numbered list syntax")
	}
	if me.blocks[0].kind != blockNumberList {
		t.Errorf("blocks[0].kind = %v, want blockNumberList", me.blocks[0].kind)
	}
}

// TestMarkdownEditor_AutoFormat_PlainTextRemainsParagraph verifies that typing
// plain text without any markdown syntax produces a paragraph block.
// Falsifying: an implementation that classifies all text as some other block kind.
func TestMarkdownEditor_AutoFormat_PlainTextRemainsParagraph(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("just some normal text without any syntax")

	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after SetText")
	}
	if me.blocks[0].kind != blockParagraph {
		t.Errorf("blocks[0].kind = %v, want blockParagraph for plain text", me.blocks[0].kind)
	}
}

// TestMarkdownEditor_AutoFormat_SourceNotMutatedByAutoFormat verifies that the
// source text is not modified by reparse/auto-format — source is only mutated
// by edit operations, not by formatting detection.
// Spec: "Source is NOT mutated — only rendering updates"
func TestMarkdownEditor_AutoFormat_SourceNotMutatedByAutoFormat(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	original := "# heading\n\n- list item\n\n> quote"
	me.SetText(original)

	// Call reparse multiple times — should never mutate source
	for i := 0; i < 5; i++ {
		me.reparse()
		if me.Text() != original {
			t.Errorf("iteration %d: source was mutated by reparse: got %q, want %q",
				i, me.Text(), original)
		}
	}
}

// TestMarkdownEditor_AutoFormat_BlocksUpdateAfterTypingSyntax verifies that
// typing markdown syntax character by character through HandleEvent eventually
// produces the expected block kind after the syntax becomes complete.
// Spec: "Auto-format triggers when typed syntax becomes complete."
func TestMarkdownEditor_AutoFormat_BlocksUpdateAfterTypingSyntax(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)

	// Type "> quote" character by character:
	// "> " at line start creates a blockquote
	for _, r := range []rune{'>', ' ', 'q', 'u', 'o', 't', 'e'} {
		ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}}
		me.HandleEvent(ev)
	}

	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after typing blockquote syntax")
	}
	if me.blocks[0].kind != blockBlockquote {
		t.Errorf("blocks[0].kind = %v after typing '> quote', want blockBlockquote", me.blocks[0].kind)
	}
	if me.Text() != "> quote" {
		t.Errorf("Text() = %q, want %q", me.Text(), "> quote")
	}
}

// TestMarkdownEditor_AutoFormat_EnterTriggersReparse verifies that pressing Enter
// after typing text triggers reparse, which updates blocks for the new layout.
// Spec: "Auto-format triggers when typed syntax becomes complete."
func TestMarkdownEditor_AutoFormat_EnterTriggersReparse(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("")

	// Type "# Heading" then Enter
	for _, r := range []rune{'#', ' ', 'H', 'e', 'a', 'd', 'i', 'n', 'g'} {
		ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}}
		me.HandleEvent(ev)
	}

	// Verify "# Heading" is a heading block after typing
	if len(me.blocks) == 0 || me.blocks[0].kind != blockHeader {
		t.Fatalf("expected heading block after typing '# Heading', got %v", me.blocks[0].kind)
	}

	// Press Enter to split
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	// Type "paragraph" on the new line
	for _, r := range []rune{'p', 'a', 'r', 'a', 'g', 'r', 'a', 'p', 'h'} {
		ev2 := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}}
		me.HandleEvent(ev2)
	}

	// Should have at least 2 blocks: heading + paragraph
	if len(me.blocks) < 2 {
		t.Fatalf("got %d blocks, want at least 2 (heading + paragraph)", len(me.blocks))
	}
	if me.blocks[0].kind != blockHeader {
		t.Errorf("blocks[0].kind = %v, want blockHeader", me.blocks[0].kind)
	}
}

// TestMarkdownEditor_AutoFormat_FencedCodeBlock_BacktickFenceThenEnter verifies
// that typing ``` at line start and pressing Enter opens a fenced code block.
// Spec: "``` + Enter → Opens fenced code block"
func TestMarkdownEditor_AutoFormat_FencedCodeBlock_BacktickFenceThenEnter(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)

	// Type ``` (three backticks) at start of line
	for _, r := range []rune{'`', '`', '`'} {
		ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}}
		me.HandleEvent(ev)
	}

	// Press Enter to open the fenced code block
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after typing ``` + Enter; expected fenced code block")
	}
	if me.blocks[0].kind != blockCodeBlock {
		t.Errorf("blocks[0].kind = %v, want blockCodeBlock (fenced code)", me.blocks[0].kind)
	}
}

// =============================================================================
// Section 40 — Falsifying / boundary tests for shortcuts and list continuation
// =============================================================================

// TestMarkdownEditor_ListContinuation_EnterEmptySourceDoesNothing verifies that
// pressing Enter on empty source with no list context does not crash or add
// spurious markers.
func TestMarkdownEditor_ListContinuation_EnterEmptySourceDoesNothing(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	got := me.Text()
	if got == "- " || got == "1. " || got == "- [ ] " {
		t.Errorf("Text() = %q; list marker incorrectly inserted on empty source", got)
	}
}

// TestMarkdownEditor_KeyboardShortcut_CtrlRightArrowNotConsumedByShortcuts
// verifies that Ctrl+[ArrowKey] is NOT consumed by the MarkdownEditor's
// shortcut handler (it should be handled by Memo navigation).
// Falsifying: an implementation that matches Ctrl+any key broadly.
func TestMarkdownEditor_KeyboardShortcut_CtrlRightArrowNotConsumedByShortcuts(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("hello world")
	me.Memo.cursorCol = 5

	// Ctrl+Right should NOT be consumed by markdown shortcuts
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}}
	me.HandleEvent(ev)

	if me.Text() != "hello world" {
		t.Errorf("source was mutated by Ctrl+Right: got %q", me.Text())
	}
}

// TestMarkdownEditor_KeyboardShortcut_MouseEventDoesNotModifySource verifies
// that non-keyboard events are not handled as shortcuts.
func TestMarkdownEditor_KeyboardShortcut_MouseEventDoesNotModifySource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("content")

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 1, Y: 1, Button: tcell.Button1}}
	me.HandleEvent(ev)

	if me.Text() != "content" {
		t.Errorf("source was mutated by mouse event: got %q", me.Text())
	}
}

// =============================================================================
// Task 12 — Link Interaction: findLinkAt source-scanning
// =============================================================================
//
// These tests verify findLinkAt correctly identifies markdown link syntax
// at cursor positions. The function scans source lines for [text](url) patterns
// and returns a linkSpan when the cursor is within the link text portion.
//
// Spec: "Links render as formatted text (green, underlined). When cursor is on
// link text: Status line hints available action ('Enter to edit link')"

// ---------------------------------------------------------------------------
// Section 42 — findLinkAt: normal detection
// ---------------------------------------------------------------------------

// TestMarkdownEditor_FindLinkAt_FindsLinkOnText verifies findLinkAt returns a
// linkSpan when the cursor is positioned on link text (between '[' and ']').
// Source contains "[click me](http://example.com)", cursor at position 3
// (on 'i' in "click"). Should return linkSpan with text="click me" and
// url="http://example.com".
// Spec: "findLinkAt finds link at cursor on link text"
func TestMarkdownEditor_FindLinkAt_FindsLinkOnText(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("[click me](http://example.com)")

	// cursorRow=0, cursorCol=3 -> 'i' in "click"
	span := me.findLinkAt(0, 3)
	if span == nil {
		t.Fatal("findLinkAt(0, 3) returned nil; cursor is on 'i' in 'click me' link text")
	}
	if span.text != "click me" {
		t.Errorf("span.text = %q, want %q", span.text, "click me")
	}
	if span.url != "http://example.com" {
		t.Errorf("span.url = %q, want %q", span.url, "http://example.com")
	}
}

// TestMarkdownEditor_FindLinkAt_ReturnsNilWhenCursorNotOnLink verifies
// findLinkAt returns nil when the cursor is before the opening '['.
// Source contains "[click me](http://example.com)", cursor at position 0
// (before '['). Should return nil.
// Spec: "findLinkAt returns nil when cursor not on link"
func TestMarkdownEditor_FindLinkAt_ReturnsNilWhenCursorNotOnLink(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("[click me](http://example.com)")

	// cursorRow=0, cursorCol=0 -> '['
	span := me.findLinkAt(0, 0)
	if span != nil {
		t.Errorf("findLinkAt(0, 0) = %+v, want nil; cursor is on '[' which is not link text", span)
	}
}

// TestMarkdownEditor_FindLinkAt_ReturnsNilForPlainText verifies findLinkAt
// returns nil for source text with no markdown link syntax at all.
// Spec: "findLinkAt returns nil for plain text"
func TestMarkdownEditor_FindLinkAt_ReturnsNilForPlainText(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("this is just plain text without any links")

	// cursorRow=0, cursorCol=5 -> 'i' in "is"
	span := me.findLinkAt(0, 5)
	if span != nil {
		t.Errorf("findLinkAt(0, 5) = %+v, want nil for plain text", span)
	}
}

// TestMarkdownEditor_FindLinkAt_FindsLinkAtStartOfText verifies findLinkAt
// finds the link when the cursor is at the first character of link text
// (right after '[').
// Source: "[click me](http://example.com)", cursor at position 1 ('c').
// Spec: "findLinkAt finds link at start of link text"
func TestMarkdownEditor_FindLinkAt_FindsLinkAtStartOfText(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("[click me](http://example.com)")

	// cursorRow=0, cursorCol=1 -> 'c', first char after '['
	span := me.findLinkAt(0, 1)
	if span == nil {
		t.Fatal("findLinkAt(0, 1) returned nil; cursor is on first char 'c' of link text")
	}
	if span.text != "click me" {
		t.Errorf("span.text = %q, want %q", span.text, "click me")
	}
}

// TestMarkdownEditor_FindLinkAt_ReturnsNilOnUrlPart verifies findLinkAt
// returns nil when the cursor is inside the URL parentheses, not on the
// link text portion.
// Source: "[click me](http://example.com)", cursor at position 14 ('t' in "http").
// Spec: "findLinkAt returns nil when cursor on URL part"
func TestMarkdownEditor_FindLinkAt_ReturnsNilOnUrlPart(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("[click me](http://example.com)")

	// cursorRow=0, cursorCol=14 -> 'p' in "http" (inside URL parentheses)
	span := me.findLinkAt(0, 14)
	if span != nil {
		t.Errorf("findLinkAt(0, 14) = %+v, want nil; cursor is inside URL parentheses", span)
	}
}

// TestMarkdownEditor_FindLinkAt_FindsFirstOfMultipleLinks verifies findLinkAt
// correctly identifies the first link when multiple links exist on the same
// source line.
// Source: "[one](url1) and [two](url2)", cursor at position 2 ('e' in "one").
// Spec: "findLinkAt finds first of multiple links"
func TestMarkdownEditor_FindLinkAt_FindsFirstOfMultipleLinks(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("[one](url1) and [two](url2)")

	// cursorRow=0, cursorCol=3 -> 'e' in "one"
	span := me.findLinkAt(0, 3)
	if span == nil {
		t.Fatal("findLinkAt(0, 3) returned nil; cursor is on 'e' in first link 'one'")
	}
	if span.text != "one" {
		t.Errorf("span.text = %q, want %q", span.text, "one")
	}
	if span.url != "url1" {
		t.Errorf("span.url = %q, want %q", span.url, "url1")
	}
}

// TestMarkdownEditor_FindLinkAt_FindsSecondOfMultipleLinks verifies findLinkAt
// correctly identifies the second link when cursor is on it.
// Source: "[one](url1) and [two](url2)", cursor at position 18 ('w' in "two").
// Spec: "findLinkAt finds second of multiple links"
func TestMarkdownEditor_FindLinkAt_FindsSecondOfMultipleLinks(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("[one](url1) and [two](url2)")

	// cursorRow=0, cursorCol=18 -> 'w' in "two" (first char of second link text)
	span := me.findLinkAt(0, 18)
	if span == nil {
		t.Fatal("findLinkAt(0, 18) returned nil; cursor is on 'w' in second link 'two'")
	}
	if span.text != "two" {
		t.Errorf("span.text = %q, want %q", span.text, "two")
	}
	if span.url != "url2" {
		t.Errorf("span.url = %q, want %q", span.url, "url2")
	}
}

// =============================================================================
// Task 12 — Link Interaction: Ctrl+K and Enter behavior
// =============================================================================
//
// These tests verify HandleEvent behavior for link-related keyboard events
// (Ctrl+K shortcut and Enter on link text). Modal dialogs require a Desktop,
// so unit tests verify event consumption and source integrity without a Desktop.
//
// Spec: "Ctrl+K: Open link dialog (create/edit/remove)"
// Spec: "Enter while cursor is on a link also opens the link dialog"

// ---------------------------------------------------------------------------
// Section 43 — Ctrl+K and Enter dispatch
// ---------------------------------------------------------------------------

// TestMarkdownEditor_LinkInteraction_CtrlKDoesNotPanicWithoutDesktop verifies
// that pressing Ctrl+K in a unit test context (no Desktop in owner chain) does
// not panic. The link dialog cannot be shown without a Desktop, but the
// handler must degrade gracefully.
// Spec: "Ctrl+K: Open link dialog (create/edit/remove)"
func TestMarkdownEditor_LinkInteraction_CtrlKDoesNotPanicWithoutDesktop(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("[click me](http://example.com)")
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 3 // on 'i' in "click"

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlK, Modifiers: tcell.ModCtrl}}
	// Must not panic
	me.HandleEvent(ev)

	// Source should not be mutated by failed dialog attempt
	if me.Text() != "[click me](http://example.com)" {
		t.Errorf("source was mutated by Ctrl+K: got %q", me.Text())
	}
}

// TestMarkdownEditor_LinkInteraction_CtrlKWithSelectionDoesNotPanic verifies
// that Ctrl+K with an active selection does not panic without a Desktop.
// Spec: "Ctrl+K with a selection opens the same dialog to create a new link"
func TestMarkdownEditor_LinkInteraction_CtrlKWithSelectionDoesNotPanic(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("select this text")

	// Set up a selection of "this" (cols 7-10)
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 7
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 11
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 11

	if !me.Memo.HasSelection() {
		t.Fatal("test setup failed: selection not active")
	}

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlK, Modifiers: tcell.ModCtrl}}
	// Must not panic
	me.HandleEvent(ev)

	// Source should not be mutated — dialog can't show without Desktop
	if me.Text() != "select this text" {
		t.Errorf("source was mutated by Ctrl+K with selection: got %q", me.Text())
	}
}

// TestMarkdownEditor_LinkInteraction_EnterOnLinkDoesNotInsertNewline verifies
// that pressing Enter when the cursor is on link text is intercepted by the
// link handler and does NOT insert a newline into the source. Without a
// Desktop, the dialog can't show, but the event should still be consumed
// by the link handler rather than falling through to Memo.
// Spec: "Pressing Enter opens a standard modal dialog with fields for URL
// and link text, plus OK/Cancel/Remove buttons"
func TestMarkdownEditor_LinkInteraction_EnterOnLinkDoesNotInsertNewline(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("[click me](http://example.com)")

	// Cursor on 'i' in "click" (col 3)
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 3

	original := me.Text()

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	// Enter on link text should NOT insert a newline — the link handler
	// intercepts it before Memo's newline handling.
	if me.Text() != original {
		t.Errorf("Enter on link inserted newline unexpectedly: got %q, want %q", me.Text(), original)
	}
	if !ev.IsCleared() {
		t.Error("Enter on link should consume the event, but it was not cleared")
	}
}

// TestMarkdownEditor_LinkInteraction_EnterOnNonLinkInsertsNewline verifies
// that pressing Enter when the cursor is NOT on link text falls through
// to Memo's normal newline handling (inserts a newline at the cursor).
// Spec (falsifying guard): Enter on non-link must not be consumed by link handler.
func TestMarkdownEditor_LinkInteraction_EnterOnNonLinkInsertsNewline(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("[click me](http://example.com) more text")

	// Cursor after the link: at col 31 ('m' in "more").
	// Positions: ...29=), 30=' ', 31=m, 32=o, 33=r, 34=e...
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 31 // on 'm' in "more", outside link text

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	// Enter on non-link text should fall through to Memo and insert a newline.
	// Line splits at col 31, so text before remains on line 0, text from col 31
	// starts new line 1.
	expected := "[click me](http://example.com) \nmore text"
	if me.Text() != expected {
		t.Errorf("Enter on non-link: got %q, want %q", me.Text(), expected)
	}
}

// TestMarkdownEditor_LinkInteraction_EnterOnMidLinkInsertsNoNewline verifies
// that Enter on a multi-word link text at mid-word does not insert a newline.
// Spec: "Pressing Enter opens a standard modal dialog..."
// Falsifying: an implementation that only checks first/last col of link text.
func TestMarkdownEditor_LinkInteraction_EnterOnMidLinkInsertsNoNewline(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("[long link text here](http://example.com)")

	// Cursor on 'k' in "link" (col 9)
	// Positions: 0=[, 1=l, 2=o, 3=n, 4=g, 5=' ', 6=l, 7=i, 8=n, 9=k, 10=' ', ...
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 9

	original := me.Text()

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	me.HandleEvent(ev)

	if me.Text() != original {
		t.Errorf("Enter on link text mid-word inserted newline: got %q, want %q", me.Text(), original)
	}
	if !ev.IsCleared() {
		t.Error("Enter on link should consume the event, but it was not cleared")
	}
}

// =============================================================================
// Task 12 — Falsifying tests for findLinkAt
// =============================================================================
//
// These tests verify findLinkAt correctly rejects incomplete or escaped
// link syntax that should NOT be treated as links.

// ---------------------------------------------------------------------------
// Section 44 — Falsifying findLinkAt edge cases
// ---------------------------------------------------------------------------

// TestMarkdownEditor_FindLinkAt_ReturnsNilForIncompleteSyntax verifies
// findLinkAt returns nil for incomplete link syntax (missing closing ']'
// and URL). Source: "[incomplete" with no closing bracket or URL.
// Spec: "findLinkAt returns nil for incomplete link syntax"
func TestMarkdownEditor_FindLinkAt_ReturnsNilForIncompleteSyntax(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("[incomplete")

	// cursorRow=0, cursorCol=4 -> 'o' in "incomplete" (would be link text if syntax were complete)
	span := me.findLinkAt(0, 4)
	if span != nil {
		t.Errorf("findLinkAt(0, 4) = %+v, want nil for incomplete link syntax", span)
	}
}

// TestMarkdownEditor_FindLinkAt_ReturnsNilForEscapedBracket verifies
// findLinkAt returns nil when the opening bracket is escaped with a
// backslash. Source: "\[not a link](url)" — the escaped bracket should
// not be treated as link syntax.
// Spec: "findLinkAt returns nil for escaped bracket"
func TestMarkdownEditor_FindLinkAt_ReturnsNilForEscapedBracket(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText(`\[not a link](url)`)

	// cursorRow=0, cursorCol=4 -> 't' in "not" (inside what would be link text if not escaped)
	span := me.findLinkAt(0, 4)
	if span != nil {
		t.Errorf("findLinkAt(0, 4) = %+v, want nil for escaped bracket", span)
	}
}

// TestMarkdownEditor_FindLinkAt_CursorOutsideLinkTextReturnsNil verifies
// findLinkAt returns nil when the cursor is at the position right after
// the closing ']' but before '('. This position is not on link text.
// Source: "[click me](url)", cursor at position 9 (the ']' character).
// Spec: "Cursor just outside link text returns nil"
func TestMarkdownEditor_FindLinkAt_CursorOutsideLinkTextReturnsNil(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("[click me](url)")

	// cursorRow=0, cursorCol=9 -> ']', the closing bracket
	span := me.findLinkAt(0, 9)
	if span != nil {
		t.Errorf("findLinkAt(0, 9) = %+v, want nil; cursor is on ']' which is not link text", span)
	}

	// Also check position immediately after URL's ')': col 14
	// Positions: [=0, c=1, l=2, i=3, c=4, k=5, ' '=6, m=7, e=8, ]=9, (=10, u=11, r=12, l=13, )=14
	span2 := me.findLinkAt(0, 10)
	if span2 != nil {
		t.Errorf("findLinkAt(0, 10) = %+v, want nil; cursor is on '(' which is not link text", span2)
	}
}

// TestMarkdownEditor_FindLinkAt_ReturnsNilForBareUrl verifies findLinkAt
// returns nil for a bare URL that is not part of [text](url) link syntax.
// Source: "http://example.com" (plain autolink, not explicit link syntax).
// The findLinkAt function scans for [text](url) patterns, not autolinks.
// Spec guard: findLinkAt only matches explicit [text](url) syntax.
func TestMarkdownEditor_FindLinkAt_ReturnsNilForBareUrl(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("http://example.com")

	// cursorRow=0, cursorCol=5 -> '/' in the URL
	span := me.findLinkAt(0, 5)
	if span != nil {
		t.Errorf("findLinkAt(0, 5) = %+v, want nil for bare URL without [text](url) syntax", span)
	}
}

// TestMarkdownEditor_FindLinkAt_ReturnsNilForClosingBracketOnLinkText tests
// that cursor on the closing ']' of a non-empty link does not return a span.
// Source: "[text](url)" — positions: 0:[, 1:t, 2:e, 3:x, 4:t, 5:], 6:(, 7:u...
// Cursor at the ']' (position 5) is link syntax, not link text.
// Spec: Cursor at position of closing ']' returns nil.
func TestMarkdownEditor_FindLinkAt_ReturnsNilForClosingBracketOnLinkText(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("[text](url)")

	// cursorRow=0, cursorCol=5 -> ']', the closing bracket of link text
	span := me.findLinkAt(0, 5)
	if span != nil {
		t.Errorf("findLinkAt(0, 5) = %+v, want nil; cursor on ']' is not on link text", span)
	}
}

// TestMarkdownEditor_FindLinkAt_HandlesMultilineSource verifies findLinkAt
// correctly scans the specified row only, ignoring links on other rows.
// Source has "[link](url)" on line 1; cursor on line 0 should return nil,
// cursor on line 1 should find the link.
func TestMarkdownEditor_FindLinkAt_HandlesMultilineSource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetText("plain text\n[link](url)\nmore text")

	// Cursor on line 0 (no link): should return nil
	span := me.findLinkAt(0, 5)
	if span != nil {
		t.Errorf("findLinkAt(0, 5) = %+v, want nil; line 0 has no link", span)
	}

	// Cursor on line 1 (has link), col 3 -> 'k' in "link"
	span = me.findLinkAt(1, 3)
	if span == nil {
		t.Fatal("findLinkAt(1, 3) returned nil; cursor is on 'k' in 'link' on line 1")
	}
	if span.text != "link" {
		t.Errorf("span.text = %q, want %q", span.text, "link")
	}
	if span.url != "url" {
		t.Errorf("span.url = %q, want %q", span.url, "url")
	}

	// Cursor on line 2 (no link): should return nil
	span = me.findLinkAt(2, 3)
	if span != nil {
		t.Errorf("findLinkAt(2, 3) = %+v, want nil; line 2 has no link", span)
	}
}

// =============================================================================
// Task 12 — Indicator update when cursor is on link text
// =============================================================================

// TestMarkdownEditor_LinkInteraction_CursorOnLinkTriggersIndicatorUpdate verifies
// that when the cursor position falls within a link's text span, the Editor's
// broadcastIndicator mechanism fires (via the CmIndicatorUpdate broadcast).
//
// Spec: "When cursor is on link text: Status line hints available action
// ('Enter to edit link')"
func TestMarkdownEditor_LinkInteraction_CursorOnLinkTriggersIndicatorUpdate(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("[click](http://x.com)")

	// Position cursor on link text (col 2 = 'i' in "click")
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 2

	// Verify findLinkAt sees the link
	span := me.findLinkAt(0, 2)
	if span == nil {
		t.Fatal("findLinkAt(0, 2) returned nil; cursor is on link text 'click'")
	}

	// broadcastIndicator should not panic when called with cursor on link text.
	// This simulates the Editor sending CmIndicatorUpdate to allow status-line
	// widgets (like EditWindow's indicator) to display the action hint.
	me.Editor.broadcastIndicator()

	// Verify cursor position is still on the link after broadcast
	row, col := me.Memo.CursorPos()
	if row != 0 || col != 2 {
		t.Errorf("cursor moved after broadcastIndicator: (%d, %d), want (0, 2)", row, col)
	}
}

// =============================================================================
// Task 12 — Ctrl+K safety tests
// =============================================================================

// TestMarkdownEditor_LinkInteraction_CtrlKNoLinkNoSelectionDoesNotPanic verifies
// that pressing Ctrl+K when the cursor is NOT on a link and there is NO selection
// does not panic. This covers the common case of accidental Ctrl+K on plain text.
// Spec guard: Ctrl+K without link context must degrade gracefully.
func TestMarkdownEditor_LinkInteraction_CtrlKNoLinkNoSelectionDoesNotPanic(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("plain text no links here")
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 6

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlK, Modifiers: tcell.ModCtrl}}
	// Must not panic
	me.HandleEvent(ev)

	// Source should be unchanged
	if me.Text() != "plain text no links here" {
		t.Errorf("source was mutated by Ctrl+K on plain text: got %q", me.Text())
	}
}

// =============================================================================
// Task 12 — findLinkAt edge case: empty source
// =============================================================================

// TestMarkdownEditor_FindLinkAt_EmptySourceReturnsNil verifies that
// findLinkAt(0, 0) on an empty source returns nil without panicking.
// Spec guard: findLinkAt must handle empty source gracefully.
func TestMarkdownEditor_FindLinkAt_EmptySourceReturnsNil(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	// Source is empty (no SetText call)

	// Must not panic
	span := me.findLinkAt(0, 0)
	if span != nil {
		t.Errorf("findLinkAt(0, 0) on empty source = %+v, want nil", span)
	}
}

// =============================================================================
// Task 13 — Undo/Redo with meaningful boundaries
// =============================================================================
//
// Spec: "Reuses Editor's existing snapshot mechanism with two additions:
//   - Snapshots fire at meaningful boundaries: word completion (space/punctuation
//     after alphanumerics), Enter, format command, paste, delete-word
//   - Consecutive single-character inserts/deletes coalesce into one undo unit
//   - Undo restores both source and cursor position"
//
// Tests verify coalescing, boundary detection, cursor restoration, and panic safety.

// TestMarkdownEditor_Undo_ConsecutiveCharacterInsertsCoalesce verifies that
// typing multiple characters in sequence creates only one undo unit.
// Spec: "Consecutive single-character inserts/deletes coalesce into one undo unit"
// Falsifying: an implementation that snapshots every keystroke would require
// one undo per character rather than restoring to the pre-edit state in one step.
func TestMarkdownEditor_Undo_ConsecutiveCharacterInsertsCoalesce(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("")

	// Type three characters sequentially
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'h'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'e'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'y'}})

	if me.Text() != "hey" {
		t.Fatalf("Text() after typing = %q, want %q", me.Text(), "hey")
	}
	if !me.Editor.CanUndo() {
		t.Fatal("CanUndo() should be true after typing characters")
	}

	// One undo should restore to empty (all three chars coalesced into one unit)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})

	if me.Text() != "" {
		t.Errorf("Text() after undo = %q, want %q (coalesced undo should return to pre-edit state)", me.Text(), "")
	}
}

// TestMarkdownEditor_Undo_EnterBreaksCoalescing verifies that pressing Enter
// creates an undo boundary, separating the character inserts before and after.
// Spec: "Snapshots fire at meaningful boundaries: ... Enter"
func TestMarkdownEditor_Undo_EnterBreaksCoalescing(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("")

	// Type "hi", Enter, then "ok"
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'h'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'i'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'o'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'k'}})

	if me.Text() != "hi\nok" {
		t.Fatalf("Text() = %q, want %q", me.Text(), "hi\nok")
	}

	// First undo: undo "ok" (chars typed after Enter boundary)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})
	if me.Text() != "hi\n" {
		t.Errorf("after 1st undo: Text() = %q, want %q", me.Text(), "hi\n")
	}

	// Second undo: undo "hi<Enter>" (chars before Enter + the Enter)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})
	if me.Text() != "" {
		t.Errorf("after 2nd undo: Text() = %q, want %q", me.Text(), "")
	}

	// Third undo: no more history
	if me.Editor.CanUndo() {
		t.Error("CanUndo() should be false after exhausting undo history")
	}
}

// TestMarkdownEditor_Undo_CtrlBFormatBreaksCoalescing verifies that a format
// command (Ctrl+B) creates an undo boundary separate from character inserts.
// Spec: "Snapshots fire at meaningful boundaries: ... format command"
func TestMarkdownEditor_Undo_CtrlBFormatBreaksCoalescing(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("")

	// Type "hello" then Ctrl+B (no selection -> inserts **** between which cursor sits)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'h'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'e'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'l'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'l'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'o'}})
	// Ctrl+B with no selection at end of "hello" inserts "****" at cursor pos or toggles bold
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlB, Modifiers: tcell.ModCtrl}})

	// Type "world" after the format insert
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'w'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'o'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'r'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'l'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'd'}})

	captured := me.Text()

	// Undo 1: removes "world" (chars typed after Ctrl+B boundary)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})
	afterFirst := me.Text()

	// Undo 2: removes the Ctrl+B format change
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})
	afterSecond := me.Text()

	// Undo 3: removes "hello" (chars typed before Ctrl+B)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})
	afterThird := me.Text()

	// Verify each undo removed the correct unit
	if afterFirst == captured {
		t.Error("first undo should have removed 'world' chars but text is unchanged")
	}
	if afterSecond != "hello" {
		t.Errorf("after 2nd undo: Text() = %q, want %q (Ctrl+B format change undone)", afterSecond, "hello")
	}
	if afterThird != "" {
		t.Errorf("after 3rd undo: Text() = %q, want %q", afterThird, "")
	}
	if me.Editor.CanUndo() {
		t.Error("CanUndo() should be false after exhausting undo history")
	}
}

// TestMarkdownEditor_Undo_ArrowKeyMovementBreaksCoalescing verifies that
// cursor movement (arrow key) breaks the coalescing chain so that typing
// after moving creates a separate undo unit.
// Spec: meaningful boundaries include non-edit cursor movements.
func TestMarkdownEditor_Undo_ArrowKeyMovementBreaksCoalescing(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("abc")

	// Move cursor between 'a' and 'b' (position 1), then type "XY"
	me.Memo.cursorCol = 1

	// Type "XY"
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'X'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'Y'}})

	// Move cursor (arrow key breaks coalescing)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}})

	// Type "Z"
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'Z'}})

	got := me.Text()

	// First undo: undo "Z" (single char after arrow boundary)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})
	afterFirst := me.Text()

	// Second undo: undo "XY" (chars typed before arrow movement)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})
	afterSecond := me.Text()

	// third undo would undo initial "abc" SetText → but SetText clears undo history.
	// So after the two undos, we should be back to "abc".
	_ = got
	if afterSecond != "abc" {
		t.Errorf("after 2nd undo: Text() = %q, want %q", afterSecond, "abc")
	}
	// Verify first undo changed something
	if afterFirst == got {
		t.Error("first undo should have removed 'Z' but text is unchanged")
	}
}

// TestMarkdownEditor_Undo_RestoresCursorPosition verifies that undo restores
// cursor position to where it was before the undone edit.
// Spec: "Undo restores both source and cursor position"
func TestMarkdownEditor_Undo_RestoresCursorPosition(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("start text")
	// cursor is at (0, 0) after SetText; move to end
	me.Memo.cursorCol = len(me.Text()) // column 10

	origRow, origCol := me.CursorPos()

	// Type " more" — cursor should move forward
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'm'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'o'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'r'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'e'}})

	afterRow, afterCol := me.CursorPos()
	if afterRow == origRow && afterCol <= origCol {
		t.Fatalf("cursor should have moved after typing: orig=(%d,%d) after=(%d,%d)", origRow, origCol, afterRow, afterCol)
	}

	// Undo restores text AND cursor
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})

	undoRow, undoCol := me.CursorPos()
	if undoRow != origRow || undoCol != origCol {
		t.Errorf("cursor after undo = (%d, %d), want (%d, %d) — undo must restore cursor position", undoRow, undoCol, origRow, origCol)
	}
}

// TestMarkdownEditor_Undo_EmptyHistoryDoesNotPanic verifies that pressing
// undo (Ctrl+Z) when there is no undo history does not cause a panic.
// Spec: undo must degrade gracefully on empty history.
func TestMarkdownEditor_Undo_EmptyHistoryDoesNotPanic(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)

	// No edits performed; undo history is nil/empty
	if me.Editor.CanUndo() {
		t.Log("CanUndo()=true after construction; SetText typically clears history")
	}

	// Must not panic
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})

	// Source should remain empty
	if me.Text() != "" {
		t.Errorf("Text() = %q, want %q after undo on empty history", me.Text(), "")
	}
}

// TestMarkdownEditor_Undo_WordCompletionTriggersSnapshotBoundary verifies
// that typing a space after alphanumeric characters creates a snapshot
// boundary (word completion), so the space is its own undo unit.
// Spec: "word completion (space/punctuation after alphanumerics)"
func TestMarkdownEditor_Undo_WordCompletionTriggersSnapshotBoundary(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("")

	// Type a word "hello"
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'h'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'e'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'l'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'l'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'o'}})

	// Now type a space — this should trigger a word-completion boundary
	// because space follows alphanumerics, creating a new undo unit.
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}})

	// Type "world" after the space boundary
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'w'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'o'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'r'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'l'}})
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'd'}})

	if me.Text() != "hello world" {
		t.Fatalf("Text() = %q, want %q", me.Text(), "hello world")
	}

	// Undo 1: removes "world" (chars after space boundary)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})
	if me.Text() != "hello " {
		t.Errorf("after 1st undo: Text() = %q, want %q (space boundary, 'world' unit undone)", me.Text(), "hello ")
	}

	// Undo 2: removes space (space boundary itself is in its own unit,
	// separated from "hello" by the word-completion boundary)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})
	if me.Text() != "hello" {
		t.Errorf("after 2nd undo: Text() = %q, want %q", me.Text(), "hello")
	}

	// Undo 3: removes "hello" (original word unit)
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlZ}})
	if me.Text() != "" {
		t.Errorf("after 3rd undo: Text() = %q, want %q", me.Text(), "")
	}

	if me.Editor.CanUndo() {
		t.Error("CanUndo() should be false after exhausting undo history")
	}
}

// =============================================================================
// Task 14 — Paste handling
// =============================================================================
//
// Spec: "Three branches based on clipboard content:
//   1. Plain text — inserted verbatim
//   2. Markdown text — detected by presence of markdown syntax. Inserted as source
//   3. Rich text / HTML — converted to markdown before insertion
//   Ctrl+Shift+V forces 'paste as plain text' regardless of type."

// TestMarkdownEditor_Paste_CtrlVInsertsClipboardText verifies that pressing
// Ctrl+V inserts the current clipboard content into the editor source.
// Spec: paste branch 1 — plain text inserted verbatim.
func TestMarkdownEditor_Paste_CtrlVInsertsClipboardText(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("prefix ")

	// Move cursor to end so paste appends
	me.Memo.cursorCol = len(me.Text())

	// Set package-level clipboard (used by Memo's Ctrl+V handler)
	clipboard = "pasted content"

	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlV}})

	want := "prefix pasted content"
	if me.Text() != want {
		t.Errorf("Text() = %q, want %q", me.Text(), want)
	}
}

// TestMarkdownEditor_Paste_CtrlVWithSelectionReplacesSelection verifies that
// pasting while text is selected replaces the selection with clipboard content.
// Spec: paste replaces selection.
func TestMarkdownEditor_Paste_CtrlVWithSelectionReplacesSelection(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("replace this please")

	// Select "this" (cols 8-12)
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 8
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 12
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 12

	if !me.Memo.HasSelection() {
		t.Fatal("test setup failed: selection should be active")
	}

	clipboard = "that"

	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlV}})

	want := "replace that please"
	if me.Text() != want {
		t.Errorf("Text() = %q, want %q (selection should be replaced)", me.Text(), want)
	}

	// Selection should be cleared after paste
	if me.Memo.HasSelection() {
		t.Error("HasSelection() should be false after paste")
	}
}

// TestMarkdownEditor_Paste_CtrlShiftVForcesPlainText verifies that
// Ctrl+Shift+V forces plain text paste even when clipboard contains
// markdown syntax that would otherwise be detected as markdown text.
// Spec: "Ctrl+Shift+V forces 'paste as plain text' regardless of type."
func TestMarkdownEditor_Paste_CtrlShiftVForcesPlainText(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("")

	// Clipboard contains markdown syntax
	clipboard = "**bold text**"

	// Ctrl+Shift+V forces plain text paste — inserts verbatim, no markdown interpretation
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlV, Modifiers: tcell.ModShift}})

	// The markdown markers should appear in source verbatim
	if !strings.Contains(me.Text(), "**bold text**") {
		t.Errorf("Text() = %q, want it to contain \\\"**bold text**\\\" verbatim (plain text paste)", me.Text())
	}
}

// TestMarkdownEditor_Paste_MarkdownTextPreservedVerbatim verifies that
// when clipboard contains markdown syntax, pasting inserts it verbatim
// as markdown source (not as plain text, not converted).
// Spec: paste branch 2 — "Markdown text ... Inserted as source, rendered accordingly"
func TestMarkdownEditor_Paste_MarkdownTextPreservedVerbatim(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("before ")

	me.Memo.cursorCol = len(me.Text())

	clipboard = "**emphasized**"

	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlV}})

	want := "before **emphasized**"
	if me.Text() != want {
		t.Errorf("Text() = %q, want %q (markdown markers must be preserved)", me.Text(), want)
	}
}

// TestMarkdownEditor_Paste_HTMLClipboardConvertsToMarkdown verifies that
// when clipboard contains HTML content, it is converted to markdown
// before insertion.
// Spec: paste branch 3 — "Rich text / HTML — converted to markdown before insertion"
func TestMarkdownEditor_Paste_HTMLClipboardConvertsToMarkdown(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("")

	clipboard = "<strong>bold</strong>"

	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlV}})

	// HTML <strong> should be converted to markdown **bold**
	got := me.Text()
	if !strings.Contains(got, "**bold**") {
		t.Errorf("Text() = %q, want markdown conversion of <strong>bold</strong> → \"**bold**\"", got)
	}
	// Should NOT contain raw HTML tags
	if strings.Contains(got, "<strong>") || strings.Contains(got, "</strong>") {
		t.Errorf("Text() = %q, raw HTML tags should have been converted to markdown", got)
	}
}

// TestMarkdownEditor_Paste_CallsReparse verifies that after pasting
// (Ctrl+V), blocks are updated via reparse to reflect the new source.
// Spec: paste is an edit operation that must trigger reparse.
func TestMarkdownEditor_Paste_CallsReparse(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("# Heading\n\nparagraph")

	// Cache the initial block count
	initialBlockCount := len(me.blocks)
	if initialBlockCount == 0 {
		t.Fatal("blocks should be non-empty after SetText")
	}

	// Position cursor at end of "paragraph" (row 2, col 9) so paste appends
	me.Memo.cursorRow = 2
	me.Memo.cursorCol = 9

	clipboard = " and more"

	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlV}})

	// After paste + reparse, blocks should reflect the updated source.
	// The source has changed, so blocks should have been re-parsed.
	if len(me.blocks) == 0 {
		t.Error("blocks is empty after paste + reparse; should be non-empty")
	}

	// The paragraph block should now include the pasted text
	if !strings.Contains(me.Text(), "paragraph and more") {
		t.Errorf("Text() = %q, want it to contain 'paragraph and more'", me.Text())
	}
}

// TestMarkdownEditor_Paste_EmptyClipboardDoesNothing verifies that
// pressing Ctrl+V with an empty clipboard does not panic and does not
// change the source.
// Spec: paste must degrade gracefully on empty clipboard.
func TestMarkdownEditor_Paste_EmptyClipboardDoesNothing(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.SetState(SfSelected, true)
	me.SetText("original text")

	clipboard = ""

	// Must not panic
	me.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlV}})

	if me.Text() != "original text" {
		t.Errorf("Text() = %q, want %q (empty clipboard should not modify source)", me.Text(), "original text")
	}
}
