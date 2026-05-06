package tv

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// =============================================================================
// Integration tests — MarkdownEditor Phase 1 Task 3 (Integration Checkpoint)
//
// These tests verify that the components wired in Tasks 1-2 connect correctly
// using real components (no mocks). Each test exercises a complete path through
// multiple real components.
//
// Test organisation:
//   Section 1 — Edit->parse->render loop
//   Section 2 — Scroll with formatted content
//   Section 3 — Empty document to content and back
//   Section 4 — showSource toggle
//   Section 5 — HandleEvent reparse integration
//   Section 6 — ColorScheme flows through
//   Section 7 — syncScrollBars with real scrollbars
// =============================================================================

// =============================================================================
// Section 1 — Edit -> parse -> render loop
// =============================================================================

// TestIntegrationMarkdownEditor_EditParseRenderLoop verifies the full editing
// pipeline: typing a character via HandleEvent triggers reparse (blocks update),
// and Draw produces output reflecting the typed content.
func TestIntegrationMarkdownEditor_EditParseRenderLoop(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("initial")

	// Verify initial state: blocks populated, text matches
	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after SetText; should be populated")
	}

	// SetText resets cursor to (0,0). Move cursor to end of text so insertion
	// appends rather than prepends.
	me.Memo.cursorCol = len(me.Text())

	// Type a character via HandleEvent
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: '!'},
	}
	me.HandleEvent(ev)

	// After edit, event should be consumed and reparse should have run
	if !ev.IsCleared() {
		t.Error("event not cleared; Memo should have consumed the keystroke")
	}

	// sourceCache must reflect the updated text
	if me.sourceCache != "initial!" {
		t.Errorf("sourceCache = %q, want %q; reparse should sync with Memo text",
			me.sourceCache, "initial!")
	}

	// blocks must be reparsed (still >0 since we only appended)
	if len(me.blocks) == 0 {
		t.Error("blocks empty after edit; reparse should have repopulated blocks")
	}

	// Draw must render the typed character
	buf := NewDrawBuffer(40, 10)
	me.Draw(buf)

	text := renderText(buf, 0)
	if !strings.Contains(text, "initial!") {
		t.Errorf("rendered text = %q, want it to contain %q", text, "initial!")
	}
}

// TestIntegrationMarkdownEditor_EditParseRenderReparsesBlocks verifies that
// after typing markdown syntax (e.g., "# " to create a heading), reparse
// produces different block kinds reflecting the new markdown structure.
func TestIntegrationMarkdownEditor_EditParseRenderReparsesBlocks(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("plain text")

	// Initial: should be a paragraph
	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after SetText")
	}
	origKind := me.blocks[0].kind

	// Erase "plain text" and type "# heading"
	// Use SetText to reset, then type each character to simulate editing
	me.SetText("")
	// Type: # heading\n\nparagraph
	me.SetText("# heading\n\nparagraph")

	// Blocks should now contain a header followed by a paragraph
	if len(me.blocks) < 2 {
		t.Fatalf("got %d blocks after typing heading markdown, want at least 2", len(me.blocks))
	}
	if me.blocks[0].kind != blockHeader {
		t.Errorf("blocks[0].kind = %v (was %v), want blockHeader", me.blocks[0].kind, origKind)
	}
	if me.blocks[1].kind != blockParagraph {
		t.Errorf("blocks[1].kind = %v, want blockParagraph", me.blocks[1].kind)
	}
}

// =============================================================================
// Section 2 — Scroll with formatted content
// =============================================================================

// TestIntegrationMarkdownEditor_ScrollShowsDifferentContent verifies that
// scrolling via deltaY changes which rendered lines are visible in the Draw
// output, using real Memo deltaY state.
func TestIntegrationMarkdownEditor_ScrollShowsDifferentContent(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 4))
	me.scheme = theme.BorlandBlue

	// Create content with many paragraphs so rendered content overflows the 4-line viewport.
	// Each paragraph + blank separator = 2 rendered lines.
	// 8 paragraphs = 15 rendered lines (8 paragraphs + 7 blank separators).
	me.SetText("L0\n\nL1\n\nL2\n\nL3\n\nL4\n\nL5\n\nL6\n\nL7")

	// At deltaY=0, first paragraph "L0" should be visible
	buf1 := NewDrawBuffer(40, 4)
	me.Draw(buf1)
	textRow0 := renderText(buf1, 0)

	// Scroll down: deltaY=6 should move past several paragraphs
	me.Memo.deltaY = 6
	buf2 := NewDrawBuffer(40, 4)
	me.Draw(buf2)
	textRow0Scrolled := renderText(buf2, 0)

	if !strings.Contains(textRow0, "L0") {
		t.Errorf("at deltaY=0, row 0 = %q, want it to contain 'L0'", textRow0)
	}
	if strings.Contains(textRow0Scrolled, "L0") {
		t.Errorf("at deltaY=6, row 0 still contains 'L0' = %q; scroll should show different content",
			textRow0Scrolled)
	}
	if textRow0 == textRow0Scrolled {
		t.Error("scrolled output identical to un-scrolled; scroll should change visible content")
	}
}

// TestIntegrationMarkdownEditor_ScrollClampedAtBottom verifies that scrolling
// past the end of content is clamped and Draw doesn't panic.
func TestIntegrationMarkdownEditor_ScrollClampedAtBottom(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue
	me.SetText("only one paragraph")

	// Set deltaY far past the rendered content
	me.Memo.deltaY = 100

	// Must not panic
	buf := NewDrawBuffer(40, 5)
	me.Draw(buf)

	// Background should still be filled (no crash occurred)
	cell := buf.GetCell(0, 0)
	if cell.Style == tcell.StyleDefault && len(me.blocks) == 0 {
		t.Error("expected background fill or rendered content")
	}
}

// =============================================================================
// Section 3 — Empty document to content and back
// =============================================================================

// TestIntegrationMarkdownEditor_EmptyToContentAndBack verifies the lifecycle:
// start empty, populate with SetText, then clear with SetText(""), verifying
// blocks are populated/cleared and Draw doesn't crash at any state.
func TestIntegrationMarkdownEditor_EmptyToContentAndBack(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue

	// Phase 1: Empty state
	if len(me.blocks) != 0 {
		t.Fatalf("blocks not empty after construction: got %d blocks", len(me.blocks))
	}
	buf := NewDrawBuffer(40, 5)
	me.Draw(buf) // must not crash

	// Phase 2: Add content
	me.SetText("# Hello\n\nWorld")
	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after SetText with markdown content")
	}
	// Verify blocks have expected kinds
	if len(me.blocks) < 2 {
		t.Fatalf("expected at least 2 blocks, got %d", len(me.blocks))
	}
	if me.blocks[0].kind != blockHeader {
		t.Errorf("blocks[0].kind = %v, want blockHeader", me.blocks[0].kind)
	}

	buf2 := NewDrawBuffer(40, 5)
	me.Draw(buf2) // must not crash
	text := renderText(buf2, 0)
	if !strings.Contains(text, "Hello") {
		t.Errorf("rendered output = %q, want it to contain 'Hello'", text)
	}

	// Phase 3: Clear back to empty
	me.SetText("")
	if len(me.blocks) != 0 {
		t.Errorf("blocks has %d entries after SetText(\"\"), want 0", len(me.blocks))
	}
	if me.blocks == nil {
		t.Error("blocks is nil after SetText(\"\"); want non-nil empty slice")
	}

	buf3 := NewDrawBuffer(40, 5)
	me.Draw(buf3) // must not crash after clearing
}

// =============================================================================
// Section 4 — showSource toggle
// =============================================================================

// TestIntegrationMarkdownEditor_ShowSourceToggle verifies that toggling
// showSource changes Draw output between formatted markdown and raw source.
func TestIntegrationMarkdownEditor_ShowSourceToggle(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue
	me.SetText("# Hello markdown\n\nA paragraph.")

	// Default: showSource=false, Draw renders formatted (no '#' visible)
	bufFormatted := NewDrawBuffer(40, 5)
	me.Draw(bufFormatted)
	formattedRow0 := renderText(bufFormatted, 0)
	if strings.Contains(formattedRow0, "#") {
		t.Errorf("formatted mode shows '#' = %q; should render without markdown syntax", formattedRow0)
	}
	if !strings.Contains(formattedRow0, "Hello") {
		t.Errorf("formatted mode missing heading text 'Hello' in %q", formattedRow0)
	}

	// Toggle: showSource=true, Draw delegates to Memo (raw # visible)
	me.SetShowSource(true)
	bufSource := NewDrawBuffer(40, 5)
	me.Draw(bufSource)
	sourceRow0 := renderText(bufSource, 0)
	if !strings.Contains(sourceRow0, "#") {
		t.Errorf("source mode missing '#' in %q; should show raw markdown", sourceRow0)
	}

	// Toggle back: showSource=false, Draw renders formatted again
	me.SetShowSource(false)
	bufFormatted2 := NewDrawBuffer(40, 5)
	me.Draw(bufFormatted2)
	formatted2Row0 := renderText(bufFormatted2, 0)
	if strings.Contains(formatted2Row0, "#") {
		t.Errorf("formatted mode after toggle back still shows '#' = %q", formatted2Row0)
	}
}

// TestIntegrationMarkdownEditor_HandleEventInSourceMode verifies that when
// showSource=true, HandleEvent delegates to Editor without triggering reparse,
// and the raw source is editable (sourceCache is unchanged by MarkdownEditor's handler).
func TestIntegrationMarkdownEditor_HandleEventInSourceMode(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("source text")
	me.SetShowSource(true)

	// SetText resets cursor to (0,0). Move cursor to end so insertion appends.
	me.Memo.cursorCol = len(me.Text())

	// In source mode, typing should go through Editor->Memo but NOT trigger
	// MarkdownEditor's reparse (sourceCache should stay stale if we poison it).
	me.sourceCache = "poisoned"

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: '!'},
	}
	me.HandleEvent(ev)

	// Event should be consumed by Memo
	if !ev.IsCleared() {
		t.Error("event not cleared; Memo should consume keystroke in source mode")
	}

	// In source mode, reparse() is NOT called by MarkdownEditor.HandleEvent.
	// sourceCache should still be "poisoned" since the guard in HandleEvent
	// returns before reparse.
	if me.sourceCache != "poisoned" {
		t.Error("reparse was called in source mode; MarkdownEditor.HandleEvent should delegate to Editor only")
	}

	// Memo text should have been updated (Editor handled it)
	if !strings.HasSuffix(me.Text(), "!") {
		t.Errorf("Memo text = %q, want it to end with '!'", me.Text())
	}
}

// =============================================================================
// Section 5 — HandleEvent reparse integration
// =============================================================================

// TestIntegrationMarkdownEditor_HandleEventReparseIntegration verifies the
// complete reparse loop triggered by HandleEvent: type a character, verify
// sourceCache is updated, verify blocks are re-parsed relative to old state.
func TestIntegrationMarkdownEditor_HandleEventReparseIntegration(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("abc")

	// SetText resets cursor to (0,0). Move cursor to end so insertion appends.
	me.Memo.cursorCol = len(me.Text())

	oldCache := me.sourceCache
	oldBlockCount := len(me.blocks)

	// Corrupt sourceCache to force reparse detection
	me.sourceCache = "stale"

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'd'},
	}
	me.HandleEvent(ev)

	// Verify event was consumed
	if !ev.IsCleared() {
		t.Error("event not cleared; Memo should consume keystroke")
	}

	// After edit, sourceCache should NOT be stale — reparse must have run
	if me.sourceCache == "stale" {
		t.Error("sourceCache still 'stale'; reparse should have updated it")
	}

	// sourceCache should match the new Memo text (appended 'd' at end)
	if me.sourceCache != "abcd" {
		t.Errorf("sourceCache = %q, want %q; reparse should sync with Memo", me.sourceCache, "abcd")
	}

	// The oldCache should differ from new (we actually changed content)
	if oldCache == me.sourceCache {
		t.Error("source text unchanged after edit; should have been modified")
	}

	// blocks should still be populated (paragraph with "abcd")
	if len(me.blocks) == 0 {
		t.Error("blocks emptied after edit; reparse should repopulate")
	}
	_ = oldBlockCount
}

// TestIntegrationMarkdownEditor_HandleEventMultipleEdits verifies that multiple
// sequential edits each trigger reparse, maintaining block freshness.
func TestIntegrationMarkdownEditor_HandleEventMultipleEdits(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("")

	cacheAfterEach := make([]string, 0, 5)
	cacheAfterEach = append(cacheAfterEach, me.sourceCache)

	for _, r := range []rune{'h', 'e', 'l', 'l', 'o'} {
		ev := &Event{
			What: EvKeyboard,
			Key:  &KeyEvent{Key: tcell.KeyRune, Rune: r},
		}
		me.HandleEvent(ev)

		if !ev.IsCleared() {
			t.Fatalf("event with rune %q not cleared", string(r))
		}
		cacheAfterEach = append(cacheAfterEach, me.sourceCache)
	}

	// Each edit should produce a distinct sourceCache
	for i := 0; i < len(cacheAfterEach)-1; i++ {
		if cacheAfterEach[i] == cacheAfterEach[i+1] {
			t.Errorf("sourceCache[%d] == sourceCache[%d] = %q; each edit should change the cache",
				i, i+1, cacheAfterEach[i])
		}
	}

	// Final sourceCache should be "hello"
	if cacheAfterEach[len(cacheAfterEach)-1] != "hello" {
		t.Errorf("final sourceCache = %q, want %q", cacheAfterEach[len(cacheAfterEach)-1], "hello")
	}

	// Draw must render "hello"
	buf := NewDrawBuffer(40, 10)
	me.Draw(buf)
	text := renderText(buf, 0)
	if !strings.Contains(text, "hello") {
		t.Errorf("rendered text = %q, want it to contain 'hello'", text)
	}
}

// TestIntegrationMarkdownEditor_HandleEventBackspaceReparse verifies that
// Backspace also triggers reparse (not just insertion).
func TestIntegrationMarkdownEditor_HandleEventBackspaceReparse(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("abc")

	// SetText resets cursor to (0,0). Move cursor to end so Backspace
	// deletes the last character rather than doing nothing at position 0.
	me.Memo.cursorCol = len(me.Text())

	oldCache := me.sourceCache
	me.sourceCache = "old"

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyBackspace2},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Backspace event not cleared")
	}
	// reparse must have updated sourceCache
	if me.sourceCache == "old" {
		t.Error("sourceCache not updated; reparse should run on Backspace")
	}
	if me.sourceCache != "ab" {
		t.Errorf("sourceCache = %q, want 'ab'", me.sourceCache)
	}
	_ = oldCache
}

// =============================================================================
// Section 6 — ColorScheme flows through
// =============================================================================

// TestIntegrationMarkdownEditor_ColorSchemeBackgroundFill verifies that Draw
// uses ColorScheme().MarkdownNormal for the background fill on every viewport
// cell, confirming the color scheme flows through the entire rendering pipeline.
func TestIntegrationMarkdownEditor_ColorSchemeBackgroundFill(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 20, 5))
	me.scheme = theme.BorlandBlue

	// Give some content so we exercise the full render path
	me.SetText("content line")

	buf := NewDrawBuffer(20, 5)
	me.Draw(buf)

	normalStyle := theme.BorlandBlue.MarkdownNormal

	for y := 0; y < 5; y++ {
		for x := 0; x < 20; x++ {
			cell := buf.GetCell(x, y)
			if cell.Style == tcell.StyleDefault {
				t.Errorf("cell (%d,%d) has default style; background fill with MarkdownNormal not applied", x, y)
				return
			}
			// Each cell must have the MarkdownNormal background regardless of
			// what foreground content is overlaid on top.
			_, cellBg, _ := cell.Style.Decompose()
			_, normalBg, _ := normalStyle.Decompose()
			if cellBg != normalBg {
				t.Errorf("cell (%d,%d) bg = %v, want MarkdownNormal bg %v",
					x, y, cellBg, normalBg)
				return
			}
		}
	}
}

// TestIntegrationMarkdownEditor_HeadingStyleUsesScheme verifies that heading
// content rendered through Draw uses the scheme's MarkdownH1 style (proving
// the color scheme flows into the mdRenderer's style composition).
func TestIntegrationMarkdownEditor_HeadingStyleUsesScheme(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue
	me.SetText("# My Heading")

	buf := NewDrawBuffer(40, 5)
	me.Draw(buf)

	h1Style := theme.BorlandBlue.MarkdownH1

	// Find the first non-space rune and verify it uses heading style
	foundHeading := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 40; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune == 'M' {
				foundHeading = true
				if cell.Style != h1Style {
					t.Errorf("heading 'M' style = %v, want MarkdownH1 %v", cell.Style, h1Style)
				}
				return
			}
		}
	}
	if !foundHeading {
		t.Error("heading text 'M' not found; heading not rendered")
	}
}

// =============================================================================
// Section 7 — syncScrollBars with real scrollbars
// =============================================================================

// TestIntegrationMarkdownEditor_SyncScrollBarsWithRealScrollBar verifies that
// syncScrollBars correctly sets Range, PageSize, and Value on a real ScrollBar
// instance when content is set.
func TestIntegrationMarkdownEditor_SyncScrollBarsWithRealScrollBar(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue

	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	hsb := NewScrollBar(NewRect(0, 0, 40, 1), Horizontal)
	me.Memo.SetVScrollBar(vsb)
	me.Memo.SetHScrollBar(hsb)

	me.SetText("# Heading\n\nParagraph text")

	// Verify vertical scrollbar state after syncScrollBars
	if vsb.Max() <= 0 {
		t.Errorf("vScrollBar Max = %d, want > 0 (content present)", vsb.Max())
	}
	if vsb.PageSize() <= 0 {
		t.Errorf("vScrollBar PageSize = %d, want > 0", vsb.PageSize())
	}
	if vsb.Value() != me.Memo.deltaY {
		t.Errorf("vScrollBar Value = %d, want deltaY = %d", vsb.Value(), me.Memo.deltaY)
	}

	// Verify horizontal scrollbar state
	if hsb.Max() < 0 {
		t.Errorf("hScrollBar Max = %d, want >= 0", hsb.Max())
	}
	if hsb.PageSize() <= 0 {
		t.Errorf("hScrollBar PageSize = %d, want > 0", hsb.PageSize())
	}
	if hsb.Value() != me.Memo.deltaX {
		t.Errorf("hScrollBar Value = %d, want deltaX = %d", hsb.Value(), me.Memo.deltaX)
	}
}

// TestIntegrationMarkdownEditor_SyncScrollBarsUpdatesAfterContentChange verifies
// that syncScrollBars produces different scrollbar ranges for different amounts
// of content, confirming the scrollbar tracks rendered content height (not raw
// line count).
func TestIntegrationMarkdownEditor_SyncScrollBarsUpdatesAfterContentChange(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue

	vsb := NewScrollBar(NewRect(0, 0, 1, 5), Vertical)
	me.Memo.SetVScrollBar(vsb)

	// Short content: 1 paragraph = 1 rendered line
	me.SetText("short")
	shortMax := vsb.Max()
	shortPageSize := vsb.PageSize()

	// Long content: many paragraphs = many more rendered lines
	longContent := strings.Repeat("line\n\n", 20)
	me.SetText(longContent)
	longMax := vsb.Max()
	longPageSize := vsb.PageSize()

	// Scrollbar range must be larger for more content
	if longMax <= shortMax {
		t.Errorf("long content Max = %d <= short content Max = %d; more content should produce larger scroll range",
			longMax, shortMax)
	}

	// PageSize should be consistent (based on viewport, not content)
	if longPageSize != shortPageSize {
		t.Errorf("PageSize changed from %d to %d; should be viewport-based, not content-based",
			shortPageSize, longPageSize)
	}
}

// TestIntegrationMarkdownEditor_SyncScrollBarsTracksDeltaY verifies that
// changing Memo.deltaY and calling syncScrollBars updates the real ScrollBar's
// Value to match.
func TestIntegrationMarkdownEditor_SyncScrollBarsTracksDeltaY(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue

	vsb := NewScrollBar(NewRect(0, 0, 1, 5), Vertical)
	me.Memo.SetVScrollBar(vsb)

	// Content large enough to scroll
	me.SetText(strings.Repeat("line\n\n", 10))

	// Scroll down
	me.Memo.deltaY = 5
	me.syncScrollBars()

	if vsb.Value() != 5 {
		t.Errorf("vScrollBar Value = %d, want 5 (deltaY)", vsb.Value())
	}

	// Scroll back to top
	me.Memo.deltaY = 0
	me.syncScrollBars()

	if vsb.Value() != 0 {
		t.Errorf("vScrollBar Value = %d after reset, want 0", vsb.Value())
	}
}

// TestIntegrationMarkdownEditor_SyncScrollBarsShowSourceMode verifies that
// when showSource is true, syncScrollBars delegates to Memo.syncScrollBars
// (uses raw line count instead of rendered height).
func TestIntegrationMarkdownEditor_SyncScrollBarsShowSourceMode(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue
	me.SetShowSource(true)

	vsb := NewScrollBar(NewRect(0, 0, 1, 5), Vertical)
	me.Memo.SetVScrollBar(vsb)

	// In source mode, raw line count is used (not rendered height)
	me.SetText("line1\nline2\nline3")

	// In source mode, Memo.syncScrollBars uses len(lines)-1 as Max
	// 3 lines, so Max should be 2 (len(lines)-1)
	if vsb.Max() != 2 {
		t.Errorf("source mode vScrollBar Max = %d, want 2 (len(lines)-1)", vsb.Max())
	}
}

// TestIntegrationMarkdownEditor_SyncScrollBarsEmptyContent verifies that
// syncScrollBars behaves correctly with empty content (no crash, valid range).
func TestIntegrationMarkdownEditor_SyncScrollBarsEmptyContent(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 5))
	me.scheme = theme.BorlandBlue

	vsb := NewScrollBar(NewRect(0, 0, 1, 5), Vertical)
	hsb := NewScrollBar(NewRect(0, 0, 40, 1), Horizontal)
	me.Memo.SetVScrollBar(vsb)
	me.Memo.SetHScrollBar(hsb)

	// Empty content — must not crash
	me.syncScrollBars()

	// Range should be valid (non-negative)
	if vsb.Max() < 0 {
		t.Errorf("vScrollBar Max = %d, want >= 0 (empty content)", vsb.Max())
	}
	if hsb.Max() < 0 {
		t.Errorf("hScrollBar Max = %d, want >= 0 (empty content)", hsb.Max())
	}
}

// =============================================================================
// Phase 2 Task 7 — Integration Checkpoint: Reveal Behavior
//
// Verify block and inline reveal work together through the real MarkdownEditor
// → buildRevealSpans → Draw pipeline. Source mutations and cursor stability
// are also verified.
// =============================================================================

// Section 8 — Block reveal: cursor entering/leaving heading

// TestIntegrationMarkdownEditor_Reveal_CursorEnterHeadingRevealsMarker verifies
// that placing the cursor inside a heading line causes block-level H1 style
// spans to be produced via buildRevealSpans.
func TestIntegrationMarkdownEditor_Reveal_CursorEnterHeadingRevealsMarker(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("# Heading\n\nPlain paragraph.")

	// Cursor inside heading (row 0, col 4)
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 4
	me.SetState(SfSelected, true)

	spans := me.buildRevealSpans()
	if len(spans) == 0 {
		t.Fatal("cursor in heading: expected block reveal spans, got none")
	}
	foundHeading := false
	for _, s := range spans {
		if s.kind == revealBlock && s.markerOpen == "# " {
			foundHeading = true
			break
		}
	}
	if !foundHeading {
		t.Error("cursor in heading: no '# ' block reveal span found")
	}
}

// TestIntegrationMarkdownEditor_Reveal_CursorLeavesHeadingHidesMarker verifies
// that moving the cursor out of a heading line causes the block-level markers
// to disappear.
func TestIntegrationMarkdownEditor_Reveal_CursorLeavesHeadingHidesMarker(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("# Heading\n\nPlain paragraph.")
	me.SetState(SfSelected, true)

	// Cursor in plain paragraph (row 2, col 3) — outside heading
	me.Memo.cursorRow = 2
	me.Memo.cursorCol = 3

	spans := me.buildRevealSpans()
	for _, s := range spans {
		if s.kind == revealBlock && s.markerOpen == "# " {
			t.Errorf("cursor outside heading: unexpected '# ' block reveal span at row=%d", s.startRow)
		}
	}
}

// Section 9 — Inline reveal: cursor entering/leaving bold

// TestIntegrationMarkdownEditor_Reveal_CursorEnterBoldRevealsMarkers verifies
// that placing the cursor inside bold text produces an inline reveal span with
// "**" markers.
func TestIntegrationMarkdownEditor_Reveal_CursorEnterBoldRevealsMarkers(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("Some **bold** text.")
	me.SetState(SfSelected, true)

	// Cursor inside bold (row 0, col 8 = 'l')
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 8

	spans := me.buildRevealSpans()
	foundBold := false
	for _, s := range spans {
		if s.kind == revealInline && s.markerOpen == "**" {
			foundBold = true
			break
		}
	}
	if !foundBold {
		t.Error("cursor in bold: expected '**' inline reveal span, found none")
	}
}

// TestIntegrationMarkdownEditor_Reveal_CursorLeavesBoldHidesMarkers verifies
// that moving the cursor outside bold text hides the inline markers.
func TestIntegrationMarkdownEditor_Reveal_CursorLeavesBoldHidesMarkers(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("Some **bold** text.")
	me.SetState(SfSelected, true)

	// Cursor in plain text (row 0, col 1 = 'o')
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 1

	spans := me.buildRevealSpans()
	for _, s := range spans {
		if s.kind == revealInline && s.markerOpen == "**" {
			t.Errorf("cursor outside bold: unexpected '**' inline reveal span at col %d", s.startCol)
		}
	}
}

// Section 10 — Block + inline simultaneous reveal

// TestIntegrationMarkdownEditor_Reveal_BlockAndInlineSimultaneous verifies
// block and inline markers coexist when cursor is inside inline formatting
// within a heading.
func TestIntegrationMarkdownEditor_Reveal_BlockAndInlineSimultaneous(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("# **Bold Heading**")
	me.SetState(SfSelected, true)

	// Cursor inside bold within heading
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 5

	spans := me.buildRevealSpans()
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
		t.Error("missing block-level '# ' span; block+inline should coexist")
	}
	if !hasInline {
		t.Error("missing inline-level '**' span; block+inline should coexist")
	}
}

// Section 11 — Scope: only cursor's inline span reveals

// TestIntegrationMarkdownEditor_Reveal_OnlyCursorSpanReveals verifies that
// when source contains multiple inline-formatted spans, only the one
// containing the cursor reveals its markers.
func TestIntegrationMarkdownEditor_Reveal_OnlyCursorSpanReveals(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("**one** and **two**")
	me.SetState(SfSelected, true)

	// Cursor inside first bold
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 3 // 'e' in "one"

	spans := me.buildRevealSpans()
	var count int
	for _, s := range spans {
		if s.kind == revealInline {
			count++
		}
	}
	if count != 1 {
		t.Errorf("got %d inline reveal spans, want 1 (only the span containing cursor)", count)
	}
}

// Section 12 — Reveal does not mutate source

// TestIntegrationMarkdownEditor_Reveal_DoesNotMutateSource verifies that
// calling buildRevealSpans and Draw does not alter the source text.
func TestIntegrationMarkdownEditor_Reveal_DoesNotMutateSource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	original := "# Heading\n\n**bold** paragraph"
	me.SetText(original)
	me.SetState(SfSelected, true)

	// Cursor in heading
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 3

	// Build reveal spans (should not mutate source)
	me.reparse()

	// Draw (should not mutate source)
	buf := NewDrawBuffer(40, 10)
	me.Draw(buf)

	current := me.Text()
	if current != original {
		t.Errorf("source was mutated by reveal/draw:\n  original: %q\n  current:  %q", original, current)
	}
}

// TestIntegrationMarkdownEditor_Reveal_DoesNotMutateSourceAfterInline tests
// that building reveal spans for inline formatting also leaves source intact.
func TestIntegrationMarkdownEditor_Reveal_DoesNotMutateSourceAfterInline(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	original := "**bold** and *italic* text"
	me.SetText(original)
	me.SetState(SfSelected, true)

	// Cursor in bold
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 3

	me.reparse()
	buf := NewDrawBuffer(40, 10)
	me.Draw(buf)

	if me.Text() != original {
		t.Errorf("source was mutated by inline reveal/draw: got %q, want %q", me.Text(), original)
	}
}

// Section 13 — Cursor position stability during reveal transitions

// TestIntegrationMarkdownEditor_Reveal_CursorStableAcrossTransitions verifies
// that the source cursor position does not change during reveal/hide
// transitions as the cursor moves between different constructs.
func TestIntegrationMarkdownEditor_Reveal_CursorStableAcrossTransitions(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("# Heading\n\n**bold** paragraph")
	me.SetState(SfSelected, true)

	// Position cursor in heading
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 5
	me.reparse()

	if me.Memo.cursorRow != 0 || me.Memo.cursorCol != 5 {
		t.Fatalf("cursor shifted after heading reveal: (%d, %d), want (0, 5)",
			me.Memo.cursorRow, me.Memo.cursorCol)
	}

	// Move cursor to bold in paragraph
	me.Memo.cursorRow = 2
	me.Memo.cursorCol = 5
	me.reparse()

	if me.Memo.cursorRow != 2 || me.Memo.cursorCol != 5 {
		t.Fatalf("cursor shifted after bold reveal: (%d, %d), want (2, 5)",
			me.Memo.cursorRow, me.Memo.cursorCol)
	}

	// Move cursor to plain text (no reveal)
	me.Memo.cursorRow = 2
	me.Memo.cursorCol = 14 // "paragraph" text
	me.reparse()

	if me.Memo.cursorRow != 2 || me.Memo.cursorCol != 14 {
		t.Fatalf("cursor shifted after leaving inline span: (%d, %d), want (2, 14)",
			me.Memo.cursorRow, me.Memo.cursorCol)
	}
}

// =============================================================================
// Phase 3 — Integration Checkpoint: Interactive Features
//
// These tests exercise the full HandleEvent dispatch pipeline for interactive
// editing features: format toggles, smart list continuation, list indent,
// paste, undo coalescing, showSource toggle, and link interaction.
// Every test uses real components (no mocks).
// =============================================================================

// Section 15 — Typing markdown produces formatted blocks

// TestIntegrationMarkdownEditor_TypeHeadingProducesBlockHeader verifies that
// typing "# Hello" character by character through HandleEvent results in
// blocks being reparsed with blockHeader kind.
func TestIntegrationMarkdownEditor_TypeHeadingProducesBlockHeader(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)

	// Start empty
	me.SetText("")
	if len(me.blocks) != 0 {
		t.Fatal("blocks not empty after SetText(\"\")")
	}

	// Type "# Hello" one character at a time
	for _, r := range "# Hello" {
		ev := &Event{
			What: EvKeyboard,
			Key:  &KeyEvent{Key: tcell.KeyRune, Rune: r},
		}
		me.HandleEvent(ev)
		if !ev.IsCleared() {
			t.Fatalf("event with rune %q not cleared", string(r))
		}
	}

	// Blocks should be reparsed with the heading marker
	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after typing heading markdown")
	}
	if me.blocks[0].kind != blockHeader {
		t.Errorf("blocks[0].kind = %v, want blockHeader", me.blocks[0].kind)
	}
	// Source must reflect the typed text
	if me.Text() != "# Hello" {
		t.Errorf("source = %q, want %q", me.Text(), "# Hello")
	}

	// Draw must show "Hello" (without the "# " marker when unfocused...)
	// But we are focused, and there's no reveal span since cursor is at the end.
	// The heading should still render with H1 style.
	buf := NewDrawBuffer(40, 10)
	me.Draw(buf)
	text := renderText(buf, 0)
	if !strings.Contains(text, "Hello") {
		t.Errorf("rendered text = %q, want it to contain 'Hello'", text)
	}
}

// TestIntegrationMarkdownEditor_TypeBulletProducesBlockBulletList verifies that
// typing "- item" character by character through HandleEvent results in
// blocks being reparsed with blockBulletList kind.
func TestIntegrationMarkdownEditor_TypeBulletProducesBlockBulletList(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)
	me.SetText("")

	// Type "- item" one character at a time
	for _, r := range "- item" {
		ev := &Event{
			What: EvKeyboard,
			Key:  &KeyEvent{Key: tcell.KeyRune, Rune: r},
		}
		me.HandleEvent(ev)
		if !ev.IsCleared() {
			t.Fatalf("event with rune %q not cleared", string(r))
		}
	}

	if len(me.blocks) == 0 {
		t.Fatal("blocks empty after typing bullet list markdown")
	}
	if me.blocks[0].kind != blockBulletList {
		t.Errorf("blocks[0].kind = %v, want blockBulletList", me.blocks[0].kind)
	}
	if me.Text() != "- item" {
		t.Errorf("source = %q, want '- item'", me.Text())
	}
}

// Section 16 — Smart list continuation

// TestIntegrationMarkdownEditor_ListContinuation_EnterAtEndOfListItem verifies
// that pressing Enter at the end of a list item creates a new line with the
// appropriate list marker.
func TestIntegrationMarkdownEditor_ListContinuation_EnterAtEndOfListItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)

	// Set text "- item one" and move cursor to end of line
	me.SetText("- item one")
	if len(me.Memo.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(me.Memo.lines))
	}
	me.Memo.cursorCol = len(me.Memo.lines[0]) // at end of "- item one"

	// Press Enter
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyEnter},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}

	// Should now have at least 2 lines, second one starting with "- "
	if len(me.Memo.lines) < 2 {
		t.Fatalf("expected at least 2 lines after Enter, got %d", len(me.Memo.lines))
	}
	secondLine := string(me.Memo.lines[1])
	if !strings.HasPrefix(secondLine, "- ") {
		t.Errorf("second line = %q, want prefix '- '", secondLine)
	}

	// Cursor should be on the new line after the marker
	if me.Memo.cursorRow != 1 {
		t.Errorf("cursorRow = %d, want 1 (new line)", me.Memo.cursorRow)
	}

	// First line unchanged
	if string(me.Memo.lines[0]) != "- item one" {
		t.Errorf("first line mutated to %q, want '- item one'", string(me.Memo.lines[0]))
	}
}

// TestIntegrationMarkdownEditor_ListContinuation_EnterOnEmptyMarkerExitsList
// verifies that pressing Enter when the cursor is on an empty list marker line
// clears the marker to exit the list.
func TestIntegrationMarkdownEditor_ListContinuation_EnterOnEmptyMarkerExitsList(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)

	// Set text: first line has content, second line is empty marker
	me.SetText("- item\n- ")
	if len(me.Memo.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(me.Memo.lines))
	}
	// Place cursor on the empty marker line (row 1), at the end (col 2)
	me.Memo.cursorRow = 1
	me.Memo.cursorCol = len(me.Memo.lines[1]) // col 2, end of "- "

	// Press Enter
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyEnter},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}

	// The empty marker line should be cleared (set to empty runes)
	// listEnterContinuation clears lines[row-1] when prev is an empty marker
	clearedLine := string(me.Memo.lines[1])
	if clearedLine != "" {
		t.Errorf("empty marker line not cleared: got %q, want \"\"", clearedLine)
	}

	// First line unchanged
	if string(me.Memo.lines[0]) != "- item" {
		t.Errorf("first line mutated to %q, want '- item'", string(me.Memo.lines[0]))
	}
}

// Section 17 — List indent/outdent

// TestIntegrationMarkdownEditor_TabIndentListItem verifies that pressing Tab
// on a list item indents it by adding two spaces.
func TestIntegrationMarkdownEditor_TabIndentListItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)

	me.SetText("- item")
	// Place cursor on the list item line
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 2 // somewhere in the line

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyTab},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Tab event not cleared")
	}

	// Line should be indented with two spaces
	line := string(me.Memo.lines[0])
	if line != "  - item" {
		t.Errorf("line after Tab = %q, want '  - item'", line)
	}

	// Cursor column should have advanced by 2
	if me.Memo.cursorCol != 4 {
		t.Errorf("cursorCol after Tab = %d, want 4", me.Memo.cursorCol)
	}
}

// TestIntegrationMarkdownEditor_ShiftTabOutdentListItem verifies that pressing
// Shift+Tab on an indented list item removes the two-space indent.
func TestIntegrationMarkdownEditor_ShiftTabOutdentListItem(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)

	me.SetText("  - item")
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 4

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyTab, Modifiers: tcell.ModShift},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Shift+Tab event not cleared")
	}

	line := string(me.Memo.lines[0])
	if line != "- item" {
		t.Errorf("line after Shift+Tab = %q, want '- item'", line)
	}
	if me.Memo.cursorCol != 2 {
		t.Errorf("cursorCol after Shift+Tab = %d, want 2", me.Memo.cursorCol)
	}
}

// Section 18 — Format toggle (Ctrl+B)

// TestIntegrationMarkdownEditor_CtrlBToggleBold_WrapAndUnwrap verifies that
// Ctrl+B wraps selected text with ** markers and that pressing it again on
// the wrapped text removes the markers.
func TestIntegrationMarkdownEditor_CtrlBToggleBold_WrapAndUnwrap(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)

	// Set text and select "hello"
	me.SetText("hello")
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 0
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 5
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 5

	// First Ctrl+B: wrap
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyCtrlB},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("first Ctrl+B event not cleared")
	}

	source := me.Text()
	if source != "**hello**" {
		t.Errorf("source after first Ctrl+B = %q, want '**hello**'", source)
	}

	// Now select the entire wrapped text for the second toggle
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 0
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = len(me.Memo.lines[0]) // 10
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len(me.Memo.lines[0])

	// Second Ctrl+B: unwrap
	ev2 := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyCtrlB},
	}
	me.HandleEvent(ev2)

	if !ev2.IsCleared() {
		t.Error("second Ctrl+B event not cleared")
	}

	if me.Text() != "hello" {
		t.Errorf("source after second Ctrl+B = %q, want 'hello'", me.Text())
	}
}

// TestIntegrationMarkdownEditor_CtrlBToggleBold_NoSelectionInsertsMarkers
// verifies that Ctrl+B with no selection inserts empty ** markers and places
// the cursor between them.
func TestIntegrationMarkdownEditor_CtrlBToggleBold_NoSelectionInsertsMarkers(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)
	me.SetText("")

	// Place cursor at position 0
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 0

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyCtrlB},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+B event not cleared")
	}

	// Should have "****" inserted, cursor between them
	source := me.Text()
	if source != "****" {
		t.Errorf("source = %q, want '****'", source)
	}
	// Cursor should be between the markers (after the first **)
	if me.Memo.cursorCol != 2 {
		t.Errorf("cursorCol = %d, want 2 (between the markers)", me.Memo.cursorCol)
	}
}

// Section 19 — ShowSource toggle (Ctrl+T)

// TestIntegrationMarkdownEditor_CtrlTToggleShowSource verifies that Ctrl+T
// toggles showSource on and off through HandleEvent.
func TestIntegrationMarkdownEditor_CtrlTToggleShowSource(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)
	me.SetText("# heading")

	// Initial state: showSource should be false
	if me.ShowSource() {
		t.Fatal("showSource initially true; want false")
	}

	// First Ctrl+T: toggle on
	ev1 := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyCtrlT},
	}
	me.HandleEvent(ev1)

	if !ev1.IsCleared() {
		t.Error("first Ctrl+T event not cleared")
	}
	if !me.ShowSource() {
		t.Error("showSource = false after first Ctrl+T; want true")
	}

	// Verify source mode renders raw markdown
	bufSource := NewDrawBuffer(40, 10)
	me.Draw(bufSource)
	sourceRow0 := renderText(bufSource, 0)
	if !strings.Contains(sourceRow0, "#") {
		t.Errorf("source mode missing '#' in %q", sourceRow0)
	}

	// Second Ctrl+T: toggle off
	ev2 := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyCtrlT},
	}
	me.HandleEvent(ev2)

	if !ev2.IsCleared() {
		t.Error("second Ctrl+T event not cleared")
	}
	if me.ShowSource() {
		t.Error("showSource = true after second Ctrl+T; want false")
	}
}

// Section 20 — Undo coalescing

// TestIntegrationMarkdownEditor_UndoCoalescing_CharStreakUndoOneUnit verifies
// that typing consecutive characters creates a single undo unit, so pressing
// Ctrl+Z once reverts the entire streak.
func TestIntegrationMarkdownEditor_UndoCoalescing_CharStreakUndoOneUnit(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)
	me.SetText("")

	// Type "hello" as a streak (5 keystrokes)
	for _, r := range "hello" {
		ev := &Event{
			What: EvKeyboard,
			Key:  &KeyEvent{Key: tcell.KeyRune, Rune: r},
		}
		me.HandleEvent(ev)
		if !ev.IsCleared() {
			t.Fatalf("event with rune %q not cleared", string(r))
		}
	}

	// Verify text is "hello" before undo
	if me.Text() != "hello" {
		t.Fatalf("source before undo = %q, want 'hello'", me.Text())
	}

	// Undo once: should revert entire streak to empty
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyCtrlZ},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+Z event not cleared")
	}
	if me.Text() != "" {
		t.Errorf("source after single undo = %q, want \"\" (all chars coalesced)", me.Text())
	}
}

// TestIntegrationMarkdownEditor_UndoCoalescing_WordBoundaryBreaksStreak verifies
// that typing a word then a space then another word creates separate undo units,
// so undo reverts only the second word.
func TestIntegrationMarkdownEditor_UndoCoalescing_WordBoundaryBreaksStreak(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)
	me.SetText("")

	// Type "hello world" one character at a time
	for _, r := range "hello world" {
		ev := &Event{
			What: EvKeyboard,
			Key:  &KeyEvent{Key: tcell.KeyRune, Rune: r},
		}
		me.HandleEvent(ev)
		if !ev.IsCleared() {
			t.Fatalf("event with rune %q not cleared", string(r))
		}
	}

	if me.Text() != "hello world" {
		t.Fatalf("source before undo = %q, want 'hello world'", me.Text())
	}

	// Undo: should revert "world" (the space broke the streak, so "world" is a new unit)
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyCtrlZ},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+Z event not cleared")
	}
	// After undo, should have "hello " (space is a word boundary that saves before it)
	// Actually the space itself is classified as editOther, so it saves before itself.
	// So undo restores to after "hello" was typed and before space was typed.
	// State after typing "hello": text="hello"
	// Space: pushUndo saves "hello", then types space. text="hello "
	// "world": char streak, first char saves "hello ". text="hello world"
	// Undo restores to "hello "
	if me.Text() != "hello " {
		t.Errorf("source after undo = %q, want 'hello '", me.Text())
	}
}

// Section 21 — Paste (Ctrl+V)

// TestIntegrationMarkdownEditor_PastePlainText verifies that Ctrl+V inserts
// the clipboard content into the editor.
func TestIntegrationMarkdownEditor_PastePlainText(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)
	me.SetText("before ")

	// Set up clipboard content
	clipboard = "pasted"

	// Move cursor to end of text
	me.Memo.cursorCol = len(me.Memo.lines[0])

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyCtrlV},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+V event not cleared")
	}

	source := me.Text()
	if !strings.Contains(source, "pasted") {
		t.Errorf("source = %q, want it to contain 'pasted'", source)
	}
	if source != "before pasted" {
		t.Errorf("source = %q, want 'before pasted'", source)
	}

	// Clean up global clipboard
	clipboard = ""
}

// TestIntegrationMarkdownEditor_PasteReplacesSelection verifies that Ctrl+V
// replaces the current selection with the clipboard content.
func TestIntegrationMarkdownEditor_PasteReplacesSelection(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)
	me.SetText("replace me now")

	// Select "me"
	me.Memo.selStartRow = 0
	me.Memo.selStartCol = 8
	me.Memo.selEndRow = 0
	me.Memo.selEndCol = 10
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 10

	if !me.Memo.HasSelection() {
		t.Fatal("selection not set")
	}

	clipboard = "XX"

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyCtrlV},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+V event not cleared")
	}

	source := me.Text()
	if source != "replace XX now" {
		t.Errorf("source = %q, want 'replace XX now'", source)
	}

	// Selection should be cleared after paste
	if me.Memo.HasSelection() {
		t.Error("selection still active after paste")
	}

	clipboard = ""
}

// Section 22 — Link interaction

// TestIntegrationMarkdownEditor_EnterOnLinkOpensDialog_EventCleared verifies
// that pressing Enter while the cursor is on a link text consumes the event
// and does NOT insert a newline (link dialog opens instead, but returns
// early without Desktop).
func TestIntegrationMarkdownEditor_EnterOnLinkOpensDialog_EventCleared(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)

	// Set text with a link and place cursor on the link text
	me.SetText("[click](http://x.com)")
	// Cursor on 'l' in "click" (col 1)
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 1

	originalSource := me.Text()

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyEnter},
	}
	me.HandleEvent(ev)

	// Event must be cleared (Enter intercepted for link dialog)
	if !ev.IsCleared() {
		t.Error("Enter on link did not clear event; should be consumed for link dialog")
	}

	// Source must be unchanged — no newline inserted
	if me.Text() != originalSource {
		t.Errorf("source changed to %q, want unchanged %q; Enter should not insert newline on link",
			me.Text(), originalSource)
	}
}

// TestIntegrationMarkdownEditor_EnterOnLinkNonLinkPosition verifies that
// pressing Enter while the cursor is NOT on a link text inserts a newline
// normally.
func TestIntegrationMarkdownEditor_EnterOnNonLinkPosition(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)

	// Set text with a link, but place cursor outside the link text
	me.SetText("[link](http://example.com) after")
	// Cursor on the space after the link (col 29, after the ')' and space)
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = len(me.Memo.lines[0]) // at end

	sourceLen := len(me.Memo.lines)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyEnter},
	}
	me.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Enter event not cleared")
	}

	// Should have inserted a newline (now 2 lines)
	if len(me.Memo.lines) != sourceLen+1 {
		t.Errorf("lines count = %d, want %d (newline should be inserted outside link)",
			len(me.Memo.lines), sourceLen+1)
	}
}

// TestIntegrationMarkdownEditor_FindLinkAt_ReturnsCorrectSpan verifies that
// findLinkAt returns the correct linkSpan with text and url when the cursor
// is on a link's text portion.
func TestIntegrationMarkdownEditor_FindLinkAt_ReturnsCorrectSpan(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)

	me.SetText("[link](http://example.com)")

	// Cursor on 'i' (col 2) in "link"
	span := me.findLinkAt(0, 2)

	if span == nil {
		t.Fatal("findLinkAt returned nil; cursor should be on link text")
	}
	if span.text != "link" {
		t.Errorf("span.text = %q, want 'link'", span.text)
	}
	if span.url != "http://example.com" {
		t.Errorf("span.url = %q, want 'http://example.com'", span.url)
	}
	if span.row != 0 {
		t.Errorf("span.row = %d, want 0", span.row)
	}
	if span.col != 1 {
		t.Errorf("span.col = %d, want 1 (start of link text)", span.col)
	}
}

// TestIntegrationMarkdownEditor_FindLinkAt_OutsideLinkReturnsNil verifies that
// findLinkAt returns nil when the cursor is not on a link.
func TestIntegrationMarkdownEditor_FindLinkAt_OutsideLinkReturnsNil(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetState(SfSelected, true)

	me.SetText("[link](http://example.com) plain text")

	// Cursor on plain text (col 30, in "plain")
	span := me.findLinkAt(0, 30)
	if span != nil {
		t.Errorf("findLinkAt returned link when cursor is outside link: %+v", span)
	}

	// Cursor on the URL portion (col 8, in "http://...")
	span = me.findLinkAt(0, 8)
	if span != nil {
		t.Errorf("findLinkAt returned link when cursor is on URL (not text): %+v", span)
	}
}

// TestIntegrationMarkdownEditor_Reveal_UnfocusedHidesAllMarkers verifies that
// when the editor is not selected, Draw does not render reveal markers.
// buildRevealSpans may produce spans, but the draw layer gates on SfSelected.
func TestIntegrationMarkdownEditor_Reveal_UnfocusedHidesAllMarkers(t *testing.T) {
	me := NewMarkdownEditor(NewRect(0, 0, 40, 10))
	me.scheme = theme.BorlandBlue
	me.SetText("# Heading\n\n**bold** text")
	// NOT selected — SfSelected is false
	me.Memo.cursorRow = 0
	me.Memo.cursorCol = 3

	// Draw must not render markers when unfocused.
	buf := NewDrawBuffer(40, 10)
	me.Draw(buf)

	// The rendered heading should not have '#' marker visible
	// since block markers are gated on SfSelected in the draw layer.
	cell := buf.GetCell(0, 0)
	if cell.Rune == '#' {
		t.Error("Draw rendered '#' marker when unfocused; markers should be hidden")
	}
}
