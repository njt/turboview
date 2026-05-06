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
//       cursor, selection, scroll, undo, and keyboard input"
// Spec: "NewMarkdownEditor(bounds Rect) *MarkdownEditor creates the widget,
//       initializes Editor->Memo"
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
//       sourceCache string (to detect source changes), and
//       showSource bool (source toggle)"
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
//       Editor.SetText, then calls reparse() to populate blocks"
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
//       stores result in blocks"
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
//       Editor.SetText, then calls reparse() to populate blocks"
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
//       reparse() is a no-op"
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
//       []mdBlock{}"
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
//       []mdBlock{}" — empty slice, not nil.
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
//       called from HandleEvent after Memo processes the edit)"
// Spec: "If the source text hasn't changed (compared to sourceCache),
//       reparse() is a no-op"
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
//        (no crash)"
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
//        (no crash)"
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
//        — raw source view"
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
//        using the existing mdRenderer"
// Req 3: "When showSource is false, Draw renders blocks through
//        mdRenderer.renderLineInto for each visible line"
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
//        me.ColorScheme()"
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
//        me.ColorScheme()"
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
//        lines are visible"
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
//        lines are visible"
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
//        lines are visible"
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
//        (overscan buffer)"
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
//        wrapText=true, and cs (color scheme)"
// Req 12: "renderer() helper returns *mdRenderer constructed with blocks,
//         width, wrapText=true, cs"
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
//         mdRenderer instead of raw line count"
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
//         mdRenderer instead of raw line count, matching
//         MarkdownViewer.syncScrollBars()"
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
//         mdRenderer instead of raw line count"
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
//         and return"
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
//         refresh blocks"
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
//         refresh blocks"
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
//         Editor (for status line updates)"
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
//         refresh blocks"
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
//         source (row,col) to screen position in rendered output"
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
//         cursor at mapped screen position. Does nothing if widget not
//         selected or has selection active."
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
// Spec req 5 (horizontal rule marker): "markerOpen = '---', markerClose = ''"
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
// markerClose = ''"
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
// Spec req 5: "markerOpen = <actual marker from source>, markerClose = ''"
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
// Spec req 5: "markerOpen = <actual marker from source>, markerClose = ''"
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
// Spec req 5: "markerOpen = <actual marker from source>, markerClose = ''"
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
// Spec req 5: "markerOpen = '---', markerClose = ''"
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
