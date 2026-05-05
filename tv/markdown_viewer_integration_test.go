package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// =============================================================================
// Integration tests — MarkdownViewer full pipeline (Task 6 checkpoint)
//
// Most requirements (1-3, 5-16) are already covered by unit tests in
// markdown_viewer_test.go and integration tests in markdown_render_integration_test.go.
// This file tests the two remaining gaps:
//   - Bold foreground+background composition through the full Draw pipeline
//   - Full end-to-end pipeline: SetMarkdown -> Draw -> HandleEvent -> scrollbar sync
// =============================================================================

// TestIntegration_BoldFgBgThroughDraw verifies that drawing "**bold**" through
// the full MarkdownViewer pipeline sets MarkdownBold foreground AND MarkdownNormal
// background on the bold characters.
//
// Req 4: "Drawing with '**bold**' shows 'bold' with MarkdownBold foreground and
// MarkdownNormal background"
func TestIntegration_BoldFgBgThroughDraw(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 5))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("normal **bold** normal")

	buf := NewDrawBuffer(40, 5)
	mv.Draw(buf)

	cs := theme.BorlandBlue
	wantFg, _, _ := cs.MarkdownBold.Decompose()
	_, wantBg, _ := cs.MarkdownNormal.Decompose()

	// Find a 'b' from "bold" and verify its foreground and background
	found := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 40; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune == 'b' &&
				x+1 < 40 &&
				buf.GetCell(x+1, y).Rune == 'o' &&
				x+2 < 40 &&
				buf.GetCell(x+2, y).Rune == 'l' {

				found = true
				fg, bg, attrs := cell.Style.Decompose()

				if fg != wantFg {
					t.Errorf("bold 'b' foreground = %v, want %v (from MarkdownBold)", fg, wantFg)
				}
				if bg != wantBg {
					t.Errorf("bold 'b' background = %v, want %v (from MarkdownNormal)", bg, wantBg)
				}
				if attrs == 0 {
					t.Error("bold 'b' has no attributes set, expected at minimum AttrBold")
				}
				if attrs&tcell.AttrBold == 0 {
					t.Error("bold 'b' should have AttrBold set")
				}
				return
			}
		}
	}
	if !found {
		t.Error("bold text 'bold' not found in rendered output")
	}
}

// TestIntegration_BoldStyleDistinctFromNormal verifies bold characters are
// visually distinct from adjacent normal text characters in the same line
// (foreground differs), and that normal text has MarkdownNormal fg and bg.
func TestIntegration_BoldStyleDistinctFromNormal(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 5))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("normal **bold** normal")

	buf := NewDrawBuffer(40, 5)
	mv.Draw(buf)

	cs := theme.BorlandBlue
	normalFg, normalBg, _ := cs.MarkdownNormal.Decompose()
	boldFg, _, _ := cs.MarkdownBold.Decompose()

	// Find a normal 'n' and a bold 'b' in the same row
	var normalStyle, boldStyle tcell.Style
	for x := 0; x < 40; x++ {
		cell := buf.GetCell(x, 0)
		r := cell.Rune
		if r == 'n' && normalStyle == (tcell.Style{}) {
			normalStyle = cell.Style
		}
		if r == 'b' && boldStyle == (tcell.Style{}) {
			// Make sure this is the 'b' from "bold", not some random 'b'
			if x+3 < 40 {
				nextRunes := []rune{
					buf.GetCell(x+1, 0).Rune,
					buf.GetCell(x+2, 0).Rune,
					buf.GetCell(x+3, 0).Rune,
				}
				if string(nextRunes[:3]) == "old" {
					boldStyle = cell.Style
				}
			}
		}
		if normalStyle != (tcell.Style{}) && boldStyle != (tcell.Style{}) {
			break
		}
	}

	if normalStyle == (tcell.Style{}) {
		t.Fatal("normal character 'n' not found in rendered output")
	}
	if boldStyle == (tcell.Style{}) {
		t.Fatal("bold character 'b' from 'bold' not found in rendered output")
	}

	// Normal text should have MarkdownNormal foreground and background.
	nFg, nBg, _ := normalStyle.Decompose()
	if nFg != normalFg {
		t.Errorf("normal text foreground = %v, want %v (MarkdownNormal)", nFg, normalFg)
	}
	if nBg != normalBg {
		t.Errorf("normal text background = %v, want %v (MarkdownNormal)", nBg, normalBg)
	}

	// Bold text should have different (MarkdownBold) foreground.
	bFg, bBg, _ := boldStyle.Decompose()
	if bFg != boldFg {
		t.Errorf("bold text foreground = %v, want %v (MarkdownBold)", bFg, boldFg)
	}
	if bBg != normalBg {
		t.Errorf("bold text background = %v, want %v (MarkdownNormal, preserved from block style)", bBg, normalBg)
	}

	// Foregrounds should differ between normal and bold.
	if bFg == nFg {
		t.Error("bold foreground equals normal foreground, expected different")
	}
}

// TestIntegration_FullPipeline verifies the complete end-to-end pipeline:
// MarkdownViewer -> SetMarkdown -> Draw -> HandleEvent -> scrollbar sync.
//
// This tests that all components (viewer, renderer, event handling, scrollbar
// binding) work together as a connected system.
func TestIntegration_FullPipeline(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.scheme = theme.BorlandBlue

	// 1. SetMarkdown populates blocks and stores source.
	// Use enough content to overflow a 40x10 viewport so scrolling tests work.
	src := "# Title\n\n" +
		"Body paragraph.\n\n" +
		"- One\n- Two\n- Three\n- Four\n- Five\n- Six\n- Seven\n- Eight\n\n" +
		"---\n\n" +
		"Another paragraph with more text.\n\n" +
		"```\ncode line 1\ncode line 2\ncode line 3\n```"
	mv.SetMarkdown(src)

	if mv.Markdown() != src {
		t.Fatalf("Markdown() = %q, want %q", mv.Markdown(), src)
	}
	if len(mv.blocks) == 0 {
		t.Fatal("SetMarkdown did not parse any blocks")
	}

	// 2. Bind a vertical scrollbar.
	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	mv.SetVScrollBar(vsb)

	if vsb.OnChange == nil {
		t.Fatal("SetVScrollBar did not set OnChange callback")
	}
	if vsb.Max() == 0 {
		t.Fatal("syncScrollBars did not set scrollbar range after SetVScrollBar")
	}

	// 3. Draw produces content.
	buf := NewDrawBuffer(40, 10)
	mv.Draw(buf)

	text := renderText(buf, 0)
	if text == "" {
		t.Error("Draw produced no content on row 0")
	}
	if !ContainsRune(buf, 'T') {
		t.Error("Draw did not render 'T' from the title")
	}

	// 4. Keyboard scroll updates deltaY and syncs scrollbar.
	mv.SetState(SfSelected, true) // Focus the viewer
	downEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	mv.HandleEvent(downEv)

	if mv.DeltaY() != 1 {
		t.Errorf("deltaY after Down = %d, want 1", mv.DeltaY())
	}
	if !downEv.IsCleared() {
		t.Error("Down key did not consume event")
	}
	if vsb.Value() != mv.DeltaY() {
		t.Errorf("scrollbar Value = %d, want deltaY = %d", vsb.Value(), mv.DeltaY())
	}

	// 5. Mouse wheel scroll updates deltaY and syncs scrollbar.
	wheelEv := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.WheelDown},
	}
	mv.HandleEvent(wheelEv)

	if mv.DeltaY() != 4 {
		t.Errorf("deltaY after WheelDown from 1 = %d, want 4", mv.DeltaY())
	}
	if !wheelEv.IsCleared() {
		t.Error("WheelDown did not consume event")
	}
	if vsb.Value() != mv.DeltaY() {
		t.Errorf("scrollbar Value = %d, want deltaY = %d", vsb.Value(), mv.DeltaY())
	}

	// 6. Home resets scroll position and scrollbar.
	homeEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}}
	mv.HandleEvent(homeEv)

	if mv.DeltaY() != 0 {
		t.Errorf("deltaY after Home = %d, want 0", mv.DeltaY())
	}
	if vsb.Value() != 0 {
		t.Errorf("scrollbar Value after Home = %d, want 0", vsb.Value())
	}

	// 7. W key toggles wrap text.
	initialWrap := mv.WrapText()
	wEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'w'}}
	mv.HandleEvent(wEv)

	if mv.WrapText() == initialWrap {
		t.Error("WrapText() did not toggle after 'w' key")
	}
	if !wEv.IsCleared() {
		t.Error("'w' key did not consume event")
	}

	// 8. SetState(SfSelected, false) hides bound scrollbars.
	mv.SetState(SfSelected, false)
	if vsb.HasState(SfVisible) {
		t.Error("vertical scrollbar still visible after losing focus")
	}

	// 9. SetState(SfSelected, true) shows bound scrollbars.
	mv.SetState(SfSelected, true)
	if !vsb.HasState(SfVisible) {
		t.Error("vertical scrollbar not visible after gaining focus")
	}

	// 10. Unfocus: keyboard events not consumed.
	mv.SetState(SfSelected, false)
	oldDeltaY := mv.DeltaY()
	passThroughEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	mv.HandleEvent(passThroughEv)

	if passThroughEv.IsCleared() {
		t.Error("Down key was consumed when unfocused, should pass through")
	}
	if mv.DeltaY() != oldDeltaY {
		t.Errorf("deltaY changed from %d to %d when unfocused, should not change", oldDeltaY, mv.DeltaY())
	}

	// 11. SetVScrollBar(nil) unbinds without panic.
	mv.SetVScrollBar(nil)
	// Reaching here without panic is the assertion.
}

// =============================================================================
// Integration: Mouse click with OfFirstClick does not clear event
// This ensures BaseView.HandleEvent is called but the event survives for
// scroll processing (already tested in unit tests, but included here as a
// pipeline check).
// =============================================================================

func TestIntegration_MouseClickWithOfFirstClickClearsEvent(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	owner := &mdMockContainer{}
	mv.SetOwner(owner)

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 1, Y: 1, Button: tcell.Button1},
	}
	mv.HandleEvent(ev)

	// With OfFirstClick, BaseView transfers focus but does NOT clear the event,
	// so the viewer can process mouse events for scrolling.
	if ev.IsCleared() {
		t.Error("mouse event was cleared by BaseView; with OfFirstClick it should NOT be cleared")
	}
	if owner.focusedChild != mv {
		t.Error("click did not transfer focus to MarkdownViewer")
	}
}

// =============================================================================
// Helpers
// =============================================================================

// ContainsRune returns true if any cell in the buffer has the given rune.
func ContainsRune(buf *DrawBuffer, r rune) bool {
	for y := 0; y < buf.Height(); y++ {
		for x := 0; x < buf.Width(); x++ {
			if buf.GetCell(x, y).Rune == r {
				return true
			}
		}
	}
	return false
}
