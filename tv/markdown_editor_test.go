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
	"testing"
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
