package tv

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// =============================================================================
// Helpers
// =============================================================================

// mdMockContainer is a test double implementing Container for click-to-focus tests.
type mdMockContainer struct {
	BaseView
	focusedChild View
}

func (m *mdMockContainer) Insert(child View)            {}
func (m *mdMockContainer) Remove(child View)            {}
func (m *mdMockContainer) Children() []View             { return nil }
func (m *mdMockContainer) FocusedChild() View           { return m.focusedChild }
func (m *mdMockContainer) SetFocusedChild(child View)   { m.focusedChild = child }
func (m *mdMockContainer) ExecView(v View) CommandCode  { return CmCancel }

// renderText extracts all non-space runes from a DrawBuffer row as a string.
func renderText(buf *DrawBuffer, y int) string {
	var sb strings.Builder
	w := buf.Width()
	for x := 0; x < w; x++ {
		cell := buf.GetCell(x, y)
		if cell.Rune == 0 {
			continue
		}
		sb.WriteRune(cell.Rune)
	}
	return sb.String()
}

// cellStyle returns the style at (x, y) in the buffer.
func cellStyle(buf *DrawBuffer, x, y int) tcell.Style {
	cell := buf.GetCell(x, y)
	return cell.Style
}

// allCellsInRow returns all runes in row y (including spaces).
func allCellsInRow(buf *DrawBuffer, y int) []rune {
	w := buf.Width()
	runes := make([]rune, w)
	for x := 0; x < w; x++ {
		cell := buf.GetCell(x, y)
		runes[x] = cell.Rune
	}
	return runes
}

// =============================================================================
// Constructor tests (requirements 1-6)
// =============================================================================

// TestMarkdownViewer_NewSetsSfVisible verifies the constructor sets SfVisible.
// Req 1: "Sets SfVisible (visible by default)"
func TestMarkdownViewer_NewSetsSfVisible(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	if !mv.HasState(SfVisible) {
		t.Error("NewMarkdownViewer did not set SfVisible")
	}
}

// TestMarkdownViewer_NewSetsOfSelectable verifies the constructor sets OfSelectable.
// Req 2: "Sets OfSelectable (can receive focus)"
func TestMarkdownViewer_NewSetsOfSelectable(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	if !mv.HasOption(OfSelectable) {
		t.Error("NewMarkdownViewer did not set OfSelectable")
	}
}

// TestMarkdownViewer_NewSetsOfFirstClick verifies the constructor sets OfFirstClick.
// Req 3: "Sets OfFirstClick (focusing click processed)"
func TestMarkdownViewer_NewSetsOfFirstClick(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	if !mv.HasOption(OfFirstClick) {
		t.Error("NewMarkdownViewer did not set OfFirstClick")
	}
}

// TestMarkdownViewer_NewStoresBounds verifies the constructor records the given bounds.
func TestMarkdownViewer_NewStoresBounds(t *testing.T) {
	r := NewRect(5, 3, 40, 10)
	mv := NewMarkdownViewer(r)

	if mv.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", mv.Bounds(), r)
	}
}

// TestMarkdownViewer_NewCallsSetSelf verifies SetSelf was called so that
// click-to-focus via BaseView.HandleEvent can navigate the owner chain.
// Req 4: "Calls SetSelf(mv)"
func TestMarkdownViewer_NewCallsSetSelf(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	owner := &mdMockContainer{}
	mv.SetOwner(owner)

	// Simulate mouse click — BaseView.HandleEvent uses b.self to navigate focus
	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 1, Y: 1, Button: tcell.Button1},
	}
	mv.HandleEvent(ev)

	if owner.FocusedChild() != mv {
		t.Error("SetSelf was not called — click did not transfer focus to MarkdownViewer")
	}
}

// TestMarkdownViewer_WrapTextDefaultsToTrue verifies wrapText starts true.
// Req 5: "wrapText defaults to true"
func TestMarkdownViewer_WrapTextDefaultsToTrue(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	if !mv.WrapText() {
		t.Error("WrapText() should default to true")
	}
}

// TestMarkdownViewer_DeltaDefaultsToZero verifies deltaX and deltaY start at 0.
// Req 6: "deltaX and deltaY default to 0"
func TestMarkdownViewer_DeltaDefaultsToZero(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	if mv.deltaX != 0 {
		t.Errorf("deltaX = %d, want 0", mv.deltaX)
	}
	if mv.deltaY != 0 {
		t.Errorf("deltaY = %d, want 0", mv.deltaY)
	}
}

// =============================================================================
// SetMarkdown / Markdown tests (requirements 7-11)
// =============================================================================

// TestMarkdownViewer_SetMarkdownParsesBlocks verifies SetMarkdown parses and
// stores blocks via parseMarkdown.
// Req 7: "Parses via parseMarkdown and stores blocks"
func TestMarkdownViewer_SetMarkdownParsesBlocks(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetMarkdown("# Hello\n\nWorld")

	if mv.blocks == nil {
		t.Error("blocks is nil after SetMarkdown")
	}
	if len(mv.blocks) == 0 {
		t.Error("blocks is empty after SetMarkdown with content")
	}
}

// TestMarkdownViewer_SetMarkdownEmptyParsesEmptyBlocks verifies SetMarkdown with
// empty string produces empty blocks slice.
// Req 7: "Parses via parseMarkdown and stores blocks" (empty case)
func TestMarkdownViewer_SetMarkdownEmptyParsesEmptyBlocks(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetMarkdown("")

	if mv.blocks == nil {
		t.Error("blocks is nil after SetMarkdown with empty string")
	}
	if len(mv.blocks) != 0 {
		t.Errorf("blocks has %d entries for empty markdown, want 0", len(mv.blocks))
	}
}

// TestMarkdownViewer_SetMarkdownStoresSource verifies SetMarkdown stores the
// original source string.
// Req 8: "Stores the original source string"
func TestMarkdownViewer_SetMarkdownStoresSource(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	src := "## A Test\n\nSome **bold** text"
	mv.SetMarkdown(src)

	if mv.source != src {
		t.Errorf("source = %q, want %q", mv.source, src)
	}
}

// TestMarkdownViewer_SetMarkdownResetsDelta verifies SetMarkdown resets deltaX
// and deltaY to 0.
// Req 9: "Resets deltaX and deltaY to 0"
func TestMarkdownViewer_SetMarkdownResetsDelta(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.deltaX = 5
	mv.deltaY = 10
	mv.SetMarkdown("content")

	if mv.deltaX != 0 {
		t.Errorf("deltaX = %d after SetMarkdown, want 0", mv.deltaX)
	}
	if mv.deltaY != 0 {
		t.Errorf("deltaY = %d after SetMarkdown, want 0", mv.deltaY)
	}
}

// TestMarkdownViewer_SetMarkdownCallsSyncScrollBars verifies SetMarkdown syncs
// scrollbars after parsing.
// Req 10: "Calls syncScrollBars()"
func TestMarkdownViewer_SetMarkdownCallsSyncScrollBars(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	mv.SetVScrollBar(vsb)
	mv.SetMarkdown("# Line 1\n\nLine 2\n\nLine 3")

	// After SetMarkdown, the vertical scrollbar should have updated range
	if vsb.Max() == 0 {
		t.Error("syncScrollBars was not called by SetMarkdown — scrollbar range not updated")
	}
}

// TestMarkdownViewer_MarkdownReturnsSource verifies the Markdown() getter returns
// the original source string.
// Req 11: "Markdown() returns the original source string passed to SetMarkdown"
func TestMarkdownViewer_MarkdownReturnsSource(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	src := "plain text"
	mv.SetMarkdown(src)

	got := mv.Markdown()
	if got != src {
		t.Errorf("Markdown() = %q, want %q", got, src)
	}
}

// TestMarkdownViewer_MarkdownInitiallyEmpty verifies Markdown() returns empty
// string when SetMarkdown has not been called.
// Req 11: "Markdown() returns the original source string" (default empty)
func TestMarkdownViewer_MarkdownInitiallyEmpty(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	got := mv.Markdown()
	if got != "" {
		t.Errorf("Markdown() before SetMarkdown = %q, want \"\"", got)
	}
}

// =============================================================================
// WrapText tests (requirements 12-13)
// =============================================================================

// TestMarkdownViewer_SetWrapTextSetsWrap verifies SetWrapText sets the wrap flag.
// Req 12: "SetWrapText(wrap bool) sets wrapText"
func TestMarkdownViewer_SetWrapTextSetsWrap(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	mv.SetWrapText(false)
	if mv.WrapText() {
		t.Error("WrapText() should be false after SetWrapText(false)")
	}

	mv.SetWrapText(true)
	if !mv.WrapText() {
		t.Error("WrapText() should be true after SetWrapText(true)")
	}
}

// TestMarkdownViewer_SetWrapTextResetsDeltaX verifies SetWrapText resets deltaX.
// Req 12: "resets deltaX to 0"
func TestMarkdownViewer_SetWrapTextResetsDeltaX(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.deltaX = 7
	mv.SetWrapText(false)

	if mv.deltaX != 0 {
		t.Errorf("deltaX = %d after SetWrapText, want 0", mv.deltaX)
	}
}

// TestMarkdownViewer_SetWrapTextCallsSyncScrollBars verifies syncScrollBars is
// called when wrap changes.
// Req 12: "calls syncScrollBars()"
func TestMarkdownViewer_SetWrapTextCallsSyncScrollBars(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetMarkdown("Line 1\n\nLine 2\n\nLine 3\n\nLine 4\n\nLine 5")

	hsb := NewScrollBar(NewRect(0, 0, 40, 1), Horizontal)
	mv.SetHScrollBar(hsb)

	// Toggling wrap changes the rendered width, which affects horizontal range
	mv.SetWrapText(false)
	maxNoWrap := hsb.Max()

	mv.SetWrapText(true)
	maxWrap := hsb.Max()

	// The max range should differ because wrapping affects content width
	if maxNoWrap == 0 && maxWrap == 0 {
		t.Error("syncScrollBars not called — horizontal scrollbar range not updated")
	}
}

// TestMarkdownViewer_WrapTextGetterReturnsCurrent verifies the WrapText() getter.
// Req 13: "WrapText() returns current value"
func TestMarkdownViewer_WrapTextGetterReturnsCurrent(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	if !mv.WrapText() {
		t.Error("WrapText() initially = false, want true (default)")
	}

	mv.SetWrapText(false)
	if mv.WrapText() {
		t.Error("WrapText() = true after SetWrapText(false), want false")
	}
}

// =============================================================================
// Draw tests (requirements 14-24)
// =============================================================================

// TestMarkdownViewer_DrawFillsBackground verifies the entire bounds are filled
// with MarkdownNormal as background even when no content is rendered.
// Req 14: "Fills entire bounds with MarkdownNormal style as background"
func TestMarkdownViewer_DrawFillsBackground(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 20, 5))
	mv.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 5)
	mv.Draw(buf)

	normalStyle := theme.BorlandBlue.MarkdownNormal
	for y := 0; y < 5; y++ {
		for x := 0; x < 20; x++ {
			got := cellStyle(buf, x, y)
			if got != normalStyle {
				t.Errorf("cell (%d,%d) style = %v, want MarkdownNormal %v", x, y, got, normalStyle)
				return
			}
		}
	}
}

// TestMarkdownViewer_DrawEmptyShowsOnlyBackground verifies that when blocks is
// empty, nothing else is drawn beyond the background fill.
// Req 15: "When blocks is empty, draws nothing else (just background fill)"
func TestMarkdownViewer_DrawEmptyShowsOnlyBackground(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 20, 5))
	mv.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 5)
	mv.Draw(buf)

	// Every cell should have the background rune (space fill)
	// The fill uses ' ' rune with MarkdownNormal style
	for y := 0; y < 5; y++ {
		for x := 0; x < 20; x++ {
			cell := buf.GetCell(x, y)
			r := cell.Rune
			if r != ' ' {
				t.Errorf("cell (%d,%d) rune = %q, want ' ' (no content rendered)", x, y, string(r))
			}
		}
	}
}

// TestMarkdownViewer_DrawRendersText verifies that paragraphs render visible text.
// Req 16: "Renders lines from deltaY through deltaY + height - 1 via mdRenderer"
func TestMarkdownViewer_DrawRendersText(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 5))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("Hello World")

	buf := NewDrawBuffer(40, 5)
	mv.Draw(buf)

	text := renderText(buf, 0)
	if !strings.Contains(text, "Hello") {
		t.Errorf("rendered text %q does not contain 'Hello'", text)
	}
}

// TestMarkdownViewer_DrawRendersMultipleBlocks verifies that multiple blocks
// are rendered (separated by blank lines).
// Req 16: "Renders lines from deltaY through deltaY + height - 1 via mdRenderer"
func TestMarkdownViewer_DrawRendersMultipleBlocks(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("First paragraph.\n\nSecond paragraph.")

	buf := NewDrawBuffer(40, 10)
	mv.Draw(buf)

	text0 := renderText(buf, 0)
	text1 := renderText(buf, 1)
	text2 := renderText(buf, 2)

	if !strings.Contains(text0, "First") {
		t.Errorf("line 0 = %q, want 'First paragraph.'", text0)
	}
	// Line 1 should be blank between blocks
	if strings.TrimSpace(text1) != "" && text1 != "" {
		t.Errorf("line 1 = %q, expected blank line between blocks", text1)
	}
	if !strings.Contains(text2, "Second") {
		t.Errorf("line 2 = %q, want 'Second paragraph.'", text2)
	}
}

// TestMarkdownViewer_DrawRespectsDeltaY verifies scrolling via deltaY affects
// which lines are rendered.
// Req 16: "Renders lines from deltaY through deltaY + height - 1"
func TestMarkdownViewer_DrawRespectsDeltaY(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 3))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("Line 0\n\nLine 1\n\nLine 2")

	// Scroll down by 2 lines
	mv.deltaY = 2

	buf := NewDrawBuffer(40, 3)
	mv.Draw(buf)

	// The first rendered line should now be "Line 1"
	text := renderText(buf, 0)
	if !strings.Contains(text, "Line 1") {
		t.Errorf("first rendered line with deltaY=2 = %q, want 'Line 1'", text)
	}
}

// TestMarkdownViewer_DrawHeaderH1 verifies headers use the MarkdownH1 style.
// Req 19: "Headers use corresponding MarkdownH1-MarkdownH6 style"
func TestMarkdownViewer_DrawHeaderH1(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 3))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("# Heading 1")

	buf := NewDrawBuffer(40, 3)
	mv.Draw(buf)

	text := renderText(buf, 0)
	if !strings.Contains(text, "Heading 1") {
		t.Errorf("rendered text %q does not contain 'Heading 1'", text)
	}

	// Verify header characters use MarkdownH1 style
	normalStyle := mv.ColorScheme().MarkdownNormal
	h1Style := mv.ColorScheme().MarkdownH1

	// Find the first non-space header character and verify its style
	foundHeader := false
	for x := 0; x < 40; x++ {
		cell := buf.GetCell(x, 0)
		r, s := cell.Rune, cell.Style
		if r == 'H' {
			foundHeader = true
			if s != h1Style {
				t.Errorf("H1 'H' at (%d,0) style = %v, want MarkdownH1 %v", x, s, h1Style)
			}
			if s == tcell.StyleDefault {
				t.Error("H1 cell has default style, expected MarkdownH1")
			}
			if s == normalStyle {
				t.Error("H1 cell has MarkdownNormal style, expected MarkdownH1")
			}
			break
		}
	}
	if !foundHeader {
		t.Error("H1 header character 'H' not found")
	}
}

// TestMarkdownViewer_DrawHeaderH3 verifies H3 uses MarkdownH3 style.
// Req 19: "Headers use corresponding MarkdownH1-MarkdownH6 style"
func TestMarkdownViewer_DrawHeaderH3(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 3))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("### Heading 3")

	buf := NewDrawBuffer(40, 3)
	mv.Draw(buf)

	text := renderText(buf, 0)
	if !strings.Contains(text, "Heading") {
		t.Error("H3 text not rendered")
	}

	// Verify header characters use MarkdownH3 style
	normalStyle := mv.ColorScheme().MarkdownNormal
	h3Style := mv.ColorScheme().MarkdownH3

	foundHeader := false
	for x := 0; x < 40; x++ {
		cell := buf.GetCell(x, 0)
		r, s := cell.Rune, cell.Style
		if r == 'H' {
			foundHeader = true
			if s != h3Style {
				t.Errorf("H3 'H' at (%d,0) style = %v, want MarkdownH3 %v", x, s, h3Style)
			}
			if s == tcell.StyleDefault {
				t.Error("H3 cell has default style, expected MarkdownH3")
			}
			if s == normalStyle {
				t.Error("H3 cell has MarkdownNormal style, expected MarkdownH3")
			}
			break
		}
	}
	if !foundHeader {
		t.Error("H3 header character 'H' not found")
	}
}

// TestMarkdownViewer_DrawBoldText verifies bold runs use composed style.
// composeStyle with runBold should use MarkdownBold foreground.
// Req 17: "Each rendered line writes styled characters via buf.WriteChar" (inline)
func TestMarkdownViewer_DrawBoldText(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 3))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("normal **bold** normal")

	buf := NewDrawBuffer(40, 3)
	mv.Draw(buf)

	text := renderText(buf, 0)
	if !strings.Contains(text, "normal") || !strings.Contains(text, "bold") {
		t.Errorf("rendered text = %q, expected both 'normal' and 'bold'", text)
	}

	// Verify bold characters have a distinct style from normal characters.
	// Find a bold character and a normal character, compare their styles.
	normalStyle := mv.ColorScheme().MarkdownNormal
	var normalCellStyle, boldCellStyle tcell.Style
	for x := 0; x < 40; x++ {
		cell := buf.GetCell(x, 0)
		r := cell.Rune
		if r == 'n' && normalCellStyle == (tcell.Style{}) {
			normalCellStyle = cell.Style
		}
		if r == 'b' && boldCellStyle == (tcell.Style{}) {
			boldCellStyle = cell.Style
		}
		if normalCellStyle != (tcell.Style{}) && boldCellStyle != (tcell.Style{}) {
			break
		}
	}
	if boldCellStyle == (tcell.Style{}) {
		t.Error("bold character 'b' not found in rendered output")
	} else if boldCellStyle == normalCellStyle {
		t.Error("bold style equals normal style, expected different style")
	}
	_ = normalStyle
}

// TestMarkdownViewer_DrawItalicText verifies italic runs use composed style.
// composeStyle with runItalic should use MarkdownItalic foreground.
// Req 17: "Each rendered line writes styled characters via buf.WriteChar" (inline)
func TestMarkdownViewer_DrawItalicText(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 3))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("normal *italic* normal")

	buf := NewDrawBuffer(40, 3)
	mv.Draw(buf)

	text := renderText(buf, 0)
	if !strings.Contains(text, "normal") || !strings.Contains(text, "italic") {
		t.Errorf("rendered text = %q, expected both 'normal' and 'italic'", text)
	}

	// Verify italic characters have a distinct style from normal characters.
	var normalCellStyle, italicCellStyle tcell.Style
	for x := 0; x < 40; x++ {
		cell := buf.GetCell(x, 0)
		r := cell.Rune
		if r == 'n' && normalCellStyle == (tcell.Style{}) {
			normalCellStyle = cell.Style
		}
		if r == 'i' && italicCellStyle == (tcell.Style{}) {
			italicCellStyle = cell.Style
		}
		if normalCellStyle != (tcell.Style{}) && italicCellStyle != (tcell.Style{}) {
			break
		}
	}
	if italicCellStyle == (tcell.Style{}) {
		t.Error("italic character 'i' not found in rendered output")
	} else if italicCellStyle == normalCellStyle {
		t.Error("italic style equals normal style, expected different style")
	}
}

// TestMarkdownViewer_DrawCodeBlockBackground verifies code blocks fill full width
// with MarkdownCodeBlock style.
// Req 18: "Code blocks use MarkdownCodeBlock as background filling full widget width"
func TestMarkdownViewer_DrawCodeBlockBackground(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 30, 5))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("```\ncode line\n```")

	buf := NewDrawBuffer(30, 5)
	mv.Draw(buf)

	codeStyle := theme.BorlandBlue.MarkdownCodeBlock

	// The code line should be rendered and the code block row should have the
	// code block background. Check the entire width for code block background.
	codeFound := false
	for y := 0; y < 5; y++ {
		text := renderText(buf, y)
		if strings.Contains(text, "code line") {
			codeFound = true
			// Verify background fill at start of row
			cell := buf.GetCell(0, y)
			s := cell.Style
			if s != codeStyle {
				t.Errorf("code block row %d: cell (0,%d) style = %v, want MarkdownCodeBlock %v", y, y, s, codeStyle)
			}
			break
		}
	}
	if !codeFound {
		t.Error("code block text not found in rendered output")
	}
}

// TestMarkdownViewer_DrawHRuleCharacters verifies horizontal rules render as
// '─' (U+2500) characters in MarkdownHRule style.
// Req 20: "Horizontal rules render as '─' (U+2500) characters in MarkdownHRule
//         style across full width"
func TestMarkdownViewer_DrawHRuleCharacters(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 20, 3))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("above\n\n---\n\nbelow")

	buf := NewDrawBuffer(20, 3)
	mv.Draw(buf)

	hrStyle := theme.BorlandBlue.MarkdownHRule
	hrFound := false
	for y := 0; y < 3; y++ {
		row := allCellsInRow(buf, y)
		dashCount := 0
		for x, r := range row {
			if r == '─' {
				dashCount++
				cell := buf.GetCell(x, y)
				s := cell.Style
				if s != hrStyle {
					t.Errorf("HRule cell at (%d,%d) has style %v, want MarkdownHRule %v", x, y, s, hrStyle)
				}
			}
		}
		if dashCount >= 10 {
			hrFound = true
		}
	}
	if !hrFound {
		t.Error("horizontal rule '─' characters not found in rendered output")
	}
}

// TestMarkdownViewer_DrawBulletListMarker verifies bullet lists render with
// '•' marker in MarkdownListMarker style.
// Req 21: "Lists render with markers ('•', '1.', '☐', '☑') in MarkdownListMarker style"
func TestMarkdownViewer_DrawBulletListMarker(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 30, 5))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("- item one\n- item two")

	buf := NewDrawBuffer(30, 5)
	mv.Draw(buf)

	markerStyle := mv.ColorScheme().MarkdownListMarker

	// The bullet character '•' should appear in the rendered output with the correct style
	bulletFound := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 30; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune == '•' {
				bulletFound = true
				s := cell.Style
				if s != markerStyle {
					t.Errorf("bullet marker '•' at (%d,%d) style = %v, want MarkdownListMarker %v", x, y, s, markerStyle)
				}
			}
		}
	}
	if !bulletFound {
		t.Errorf("bullet marker '•' not found in rendered output")
	}
}

// TestMarkdownViewer_DrawNumberedListMarker verifies ordered lists render with
// '1.' style markers in MarkdownListMarker style.
// Req 21: "Lists render with markers ('•', '1.', '☐', '☑') in MarkdownListMarker style"
func TestMarkdownViewer_DrawNumberedListMarker(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 30, 5))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("1. first\n2. second")

	buf := NewDrawBuffer(30, 5)
	mv.Draw(buf)

	markerStyle := mv.ColorScheme().MarkdownListMarker

	// Should contain numbered items with markers in the correct style
	text0 := renderText(buf, 0)
	foundNumber := strings.Contains(text0, "1.") || strings.Contains(text0, "1")
	if !foundNumber {
		t.Errorf("numbered list marker not found in line 0: %q", text0)
	}

	// Verify the '1' digit has MarkdownListMarker style
	markerFound := false
	for x := 0; x < 30; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune == '1' {
			markerFound = true
			s := cell.Style
			if s != markerStyle {
				t.Errorf("numbered marker '1' at (%d,0) style = %v, want MarkdownListMarker %v", x, s, markerStyle)
			}
			break
		}
	}
	if !markerFound {
		t.Error("numbered marker digit '1' not found")
	}
}

// TestMarkdownViewer_DrawCheckListMarkerUnchecked verifies unchecked checklist
// items render with '☐' in MarkdownListMarker style.
// Req 21: "Lists render with markers ('•', '1.', '☐', '☑') in MarkdownListMarker style"
func TestMarkdownViewer_DrawCheckListMarkerUnchecked(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 30, 5))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("- [ ] unchecked item")

	buf := NewDrawBuffer(30, 5)
	mv.Draw(buf)

	markerStyle := mv.ColorScheme().MarkdownListMarker

	uncheckedFound := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 30; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune == '☐' {
				uncheckedFound = true
				s := cell.Style
				if s != markerStyle {
					t.Errorf("unchecked marker '☐' at (%d,%d) style = %v, want MarkdownListMarker %v", x, y, s, markerStyle)
				}
			}
		}
	}
	if !uncheckedFound {
		t.Errorf("unchecked marker '☐' not found in rendered output")
	}
}

// TestMarkdownViewer_DrawCheckListMarkerChecked verifies checked checklist items
// render with '☑' in MarkdownListMarker style.
// Req 21: "Lists render with markers ('•', '1.', '☐', '☑') in MarkdownListMarker style"
func TestMarkdownViewer_DrawCheckListMarkerChecked(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 30, 5))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("- [x] checked item")

	buf := NewDrawBuffer(30, 5)
	mv.Draw(buf)

	markerStyle := mv.ColorScheme().MarkdownListMarker

	checkedFound := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 30; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune == '☑' {
				checkedFound = true
				s := cell.Style
				if s != markerStyle {
					t.Errorf("checked marker '☑' at (%d,%d) style = %v, want MarkdownListMarker %v", x, y, s, markerStyle)
				}
			}
		}
	}
	if !checkedFound {
		t.Errorf("checked marker '☑' not found in rendered output")
	}
}

// TestMarkdownViewer_DrawBlockquoteBar verifies blockquotes render with '▌'
// (U+258C) left bar in MarkdownBlockquote style.
// Req 22: "Blockquotes render with '▌' (U+258C) left bar in MarkdownBlockquote style"
func TestMarkdownViewer_DrawBlockquoteBar(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 5))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("> quoted text")

	buf := NewDrawBuffer(40, 5)
	mv.Draw(buf)

	barFound := false
	for y := 0; y < 5; y++ {
		row := allCellsInRow(buf, y)
		for x, r := range row {
			if r == '▌' {
				barFound = true
				// Check style of the bar
				cell := buf.GetCell(x, y)
				s := cell.Style
				want := theme.BorlandBlue.MarkdownBlockquote
				if s != want {
					t.Errorf("blockquote bar at (%d,%d) has style %v, want MarkdownBlockquote %v", x, y, s, want)
				}
			}
		}
	}
	if !barFound {
		t.Error("blockquote '▌' bar not found in rendered output")
	}
}

// TestMarkdownViewer_DrawTableBorder verifies tables render with box-drawing
// borders in MarkdownTableBorder style.
// Req 23: "Tables render with box-drawing borders in MarkdownTableBorder style"
func TestMarkdownViewer_DrawTableBorder(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 50, 8))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("| A | B |\n|---|---|\n| 1 | 2 |")

	buf := NewDrawBuffer(50, 8)
	mv.Draw(buf)

	tableBorderStyle := theme.BorlandBlue.MarkdownTableBorder
	borderFound := false
	styleAsserted := false
	for y := 0; y < 8; y++ {
		row := allCellsInRow(buf, y)
		for x, r := range row {
			// Box-drawing characters used for table borders
			if r == '┌' || r == '┐' || r == '└' || r == '┘' || r == '─' || r == '│' || r == '├' || r == '┤' || r == '┬' || r == '┴' || r == '┼' {
				borderFound = true
				cell := buf.GetCell(x, y)
				s := cell.Style
				if s != tableBorderStyle {
					t.Errorf("table border cell '%c' at (%d,%d) style = %v, want MarkdownTableBorder %v", r, x, y, s, tableBorderStyle)
				} else {
					styleAsserted = true
				}
			}
		}
	}
	if !borderFound {
		t.Error("table box-drawing borders not found in rendered output")
	}
	if !styleAsserted {
		t.Error("no table border cell was checked for MarkdownTableBorder style")
	}
}

// TestMarkdownViewer_DrawDefTermStyle verifies definition terms render in
// MarkdownDefTerm style.
// Req 24: "Definition terms render in MarkdownDefTerm style"
func TestMarkdownViewer_DrawDefTermStyle(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 8))
	mv.scheme = theme.BorlandBlue
	mv.SetMarkdown("Term 1\n: Definition 1\n\nTerm 2\n: Definition 2")

	buf := NewDrawBuffer(40, 8)
	mv.Draw(buf)

	defTermStyle := mv.ColorScheme().MarkdownDefTerm

	// The term should be rendered with MarkdownDefTerm style
	termFound := false
	for y := 0; y < 8; y++ {
		text := renderText(buf, y)
		if strings.Contains(text, "Term") {
			termFound = true
			// Find the 'T' character and verify its style
			for x := 0; x < 40; x++ {
				cell := buf.GetCell(x, y)
				if cell.Rune == 'T' {
					s := cell.Style
					if s != defTermStyle {
						t.Errorf("definition term 'T' at (%d,%d) style = %v, want MarkdownDefTerm %v", x, y, s, defTermStyle)
					}
					if s == tcell.StyleDefault {
						t.Error("definition term cell has default style")
					}
					break
				}
			}
			break
		}
	}
	if !termFound {
		t.Error("definition term text not found in rendered output")
	}
}

// =============================================================================
// HandleEvent keyboard tests (requirements 25-35)
// =============================================================================

// TestMarkdownViewer_KeyUpDecrementsDeltaY verifies Up arrow decreases deltaY.
// Req 25: "Up: deltaY--, clamped to 0, consumes event"
func TestMarkdownViewer_KeyUpDecrementsDeltaY(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)
	mv.deltaY = 5

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	mv.HandleEvent(ev)

	if mv.deltaY != 4 {
		t.Errorf("deltaY after Up = %d, want 4", mv.deltaY)
	}
	if !ev.IsCleared() {
		t.Error("Up key did not consume event")
	}
}

// TestMarkdownViewer_KeyUpClampedToZero verifies Up clamped to 0.
// Req 25: "Up: deltaY--, clamped to 0"
func TestMarkdownViewer_KeyUpClampedToZero(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)
	mv.deltaY = 0

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	mv.HandleEvent(ev)

	if mv.deltaY != 0 {
		t.Errorf("deltaY after Up from 0 = %d, want 0 (clamped)", mv.deltaY)
	}
}

// TestMarkdownViewer_KeyDownIncrementsDeltaY verifies Down arrow increases deltaY.
// Req 26: "Down: deltaY++, clamped to max, consumes event"
func TestMarkdownViewer_KeyDownIncrementsDeltaY(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)
	mv.deltaY = 3

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	mv.HandleEvent(ev)

	if mv.deltaY != 4 {
		t.Errorf("deltaY after Down = %d, want 4", mv.deltaY)
	}
	if !ev.IsCleared() {
		t.Error("Down key did not consume event")
	}
}

// TestMarkdownViewer_KeyDownClampedToMax verifies Down clamped to max.
// Req 26: "Down: deltaY++, clamped to max"
func TestMarkdownViewer_KeyDownClampedToMax(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 5))
	mv.SetState(SfSelected, true)
	mv.SetMarkdown("a\n\nb") // small content, max is small
	mv.deltaY = 999         // way beyond max

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	mv.HandleEvent(ev)

	// deltaY should not exceed maximum scrollable range
	if mv.deltaY > 999 {
		t.Errorf("deltaY = %d, should be clamped to max", mv.deltaY)
	}
}

// TestMarkdownViewer_KeyPgUpDecrementsDeltaYByHeight verifies PgUp subtracts
// viewport height from deltaY.
// Req 27: "PgUp: deltaY -= viewport height, clamped to 0, consumes event"
func TestMarkdownViewer_KeyPgUpDecrementsDeltaYByHeight(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10)) // height = 10
	mv.SetState(SfSelected, true)
	mv.deltaY = 20

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp}}
	mv.HandleEvent(ev)

	if mv.deltaY != 10 {
		t.Errorf("deltaY after PgUp from 20 (height 10) = %d, want 10", mv.deltaY)
	}
	if !ev.IsCleared() {
		t.Error("PgUp key did not consume event")
	}
}

// TestMarkdownViewer_KeyPgUpClampedToZero verifies PgUp clamped to 0.
// Req 27: "PgUp: deltaY -= viewport height, clamped to 0"
func TestMarkdownViewer_KeyPgUpClampedToZero(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)
	mv.deltaY = 3

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgUp}}
	mv.HandleEvent(ev)

	if mv.deltaY != 0 {
		t.Errorf("deltaY after PgUp from 3 (height 10) = %d, want 0", mv.deltaY)
	}
}

// TestMarkdownViewer_KeyPgDnIncrementsDeltaYByHeight verifies PgDn adds
// viewport height to deltaY.
// Req 28: "PgDn: deltaY += viewport height, clamped to max, consumes event"
func TestMarkdownViewer_KeyPgDnIncrementsDeltaYByHeight(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10)) // height = 10
	mv.SetState(SfSelected, true)
	mv.deltaY = 0

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgDn}}
	mv.HandleEvent(ev)

	if mv.deltaY != 10 {
		t.Errorf("deltaY after PgDn from 0 (height 10) = %d, want 10", mv.deltaY)
	}
	if !ev.IsCleared() {
		t.Error("PgDn key did not consume event")
	}
}

// TestMarkdownViewer_KeyPgDnClampedToMax verifies PgDn is clamped to the
// maximum scroll position.
// Req 28: "PgDn: deltaY += viewport height, clamped to max"
func TestMarkdownViewer_KeyPgDnClampedToMax(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 5))
	mv.SetState(SfSelected, true)
	mv.SetMarkdown("line1\n\nline2\n\nline3\n\nline4\n\nline5\n\nline6\n\nline7\n\nline8")

	// Find the maximum scroll position by pressing End
	endEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd}}
	mv.HandleEvent(endEv)
	maxScroll := mv.deltaY

	if maxScroll <= 0 {
		t.Fatal("content does not overflow viewport; expected maxScroll > 0")
	}

	// Set deltaY to one less than max so PgDn would overshoot without clamping
	mv.deltaY = maxScroll - 1

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyPgDn}}
	mv.HandleEvent(ev)

	if mv.deltaY != maxScroll {
		t.Errorf("deltaY after PgDn near max = %d, want clamped to maxScroll %d", mv.deltaY, maxScroll)
	}
	if !ev.IsCleared() {
		t.Error("PgDn key did not consume event")
	}
}

// TestMarkdownViewer_KeyHomeResetsDeltaY verifies Home sets deltaY to 0.
// Req 29: "Home: deltaY = 0, consumes event"
func TestMarkdownViewer_KeyHomeResetsDeltaY(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)
	mv.deltaY = 15

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}}
	mv.HandleEvent(ev)

	if mv.deltaY != 0 {
		t.Errorf("deltaY after Home = %d, want 0", mv.deltaY)
	}
	if !ev.IsCleared() {
		t.Error("Home key did not consume event")
	}
}

// TestMarkdownViewer_KeyEndSetsDeltaYToMax verifies End sets deltaY to max and
// that max equals the actual computed maximum (rendered lines - viewport height).
// Req 30: "End: deltaY = max, consumes event"
func TestMarkdownViewer_KeyEndSetsDeltaYToMax(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 3))
	mv.SetState(SfSelected, true)
	// Content tall enough to scroll: 11 rendered lines with 3-row viewport,
	// so maxScroll should be 8.
	mv.SetMarkdown("a\n\nb\n\nc\n\nd\n\ne\n\nf")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd}}
	mv.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("End key did not consume event")
	}

	// After End, deltaY should be at the maximum scrollable position.
	maxScroll := mv.deltaY

	// Verify maxScroll is positive when content overflows the viewport.
	if maxScroll <= 0 {
		t.Fatalf("maxScroll = %d, expected > 0 for overflowing content", maxScroll)
	}

	// Verify we are actually at max: pressing Down should not increase deltaY.
	downEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	mv.HandleEvent(downEv)

	if mv.deltaY != maxScroll {
		t.Errorf("deltaY changed from %d to %d after Down at max, expected it to stay at %d", maxScroll, mv.deltaY, maxScroll)
	}
}

// TestMarkdownViewer_KeyLeftDecrementsDeltaX verifies Left arrow decreases deltaX.
// Req 31: "Left: deltaX--, clamped to 0, consumes event"
func TestMarkdownViewer_KeyLeftDecrementsDeltaX(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)
	mv.SetWrapText(false)
	mv.deltaX = 5

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	mv.HandleEvent(ev)

	if mv.deltaX != 4 {
		t.Errorf("deltaX after Left = %d, want 4", mv.deltaX)
	}
	if !ev.IsCleared() {
		t.Error("Left key did not consume event")
	}
}

// TestMarkdownViewer_KeyLeftClampedToZero verifies Left clamped to 0.
// Req 31: "Left: deltaX--, clamped to 0"
func TestMarkdownViewer_KeyLeftClampedToZero(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)
	mv.deltaX = 0

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	mv.HandleEvent(ev)

	if mv.deltaX != 0 {
		t.Errorf("deltaX after Left from 0 = %d, want 0 (clamped)", mv.deltaX)
	}
}

// TestMarkdownViewer_KeyRightIncrementsDeltaX verifies Right arrow increases deltaX.
// Req 32: "Right: deltaX++, clamped to max, consumes event"
func TestMarkdownViewer_KeyRightIncrementsDeltaX(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)
	mv.SetWrapText(false)
	mv.deltaX = 3

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	mv.HandleEvent(ev)

	if mv.deltaX != 4 {
		t.Errorf("deltaX after Right = %d, want 4", mv.deltaX)
	}
	if !ev.IsCleared() {
		t.Error("Right key did not consume event")
	}
}

// TestMarkdownViewer_KeyRightClampedToMax verifies Right is clamped to the
// maximum horizontal scroll position.
// Req 32: "Right: deltaX++, clamped to max"
func TestMarkdownViewer_KeyRightClampedToMax(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 20, 3))
	mv.SetState(SfSelected, true)
	mv.SetWrapText(false)
	mv.SetMarkdown("This is a very long line of text that exceeds the viewport width")

	// Find the maximum horizontal scroll position by scrolling right many times.
	for i := 0; i < 200; i++ {
		ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
		mv.HandleEvent(ev)
	}
	maxScroll := mv.deltaX

	if maxScroll <= 0 {
		t.Fatal("content does not overflow viewport horizontally; expected maxScroll > 0")
	}

	// Set deltaX to one less than max so Right would overshoot without clamping.
	mv.deltaX = maxScroll - 1

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	mv.HandleEvent(ev)

	if mv.deltaX != maxScroll {
		t.Errorf("deltaX after Right near max = %d, want clamped to maxScroll %d", mv.deltaX, maxScroll)
	}
	if !ev.IsCleared() {
		t.Error("Right key did not consume event")
	}
}

// TestMarkdownViewer_KeyWTogglesWrap verifies 'w' toggles wrapText and resets deltaX.
// Req 33: "W (rune 'w' or 'W'): toggle wrapText, reset deltaX to 0, consumes event"
func TestMarkdownViewer_KeyWTogglesWrap(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)
	mv.SetMarkdown("content")

	initial := mv.WrapText()

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'w'}}
	mv.HandleEvent(ev)

	if mv.WrapText() == initial {
		t.Errorf("WrapText() = %v after 'w', want %v (toggled)", mv.WrapText(), !initial)
	}
	if !ev.IsCleared() {
		t.Error("'w' key did not consume event")
	}
}

// TestMarkdownViewer_KeyWTogglesWrapUpperCase verifies uppercase 'W' also toggles.
// Req 33: "W (rune 'w' or 'W'): toggle wrapText"
func TestMarkdownViewer_KeyWTogglesWrapUpperCase(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)
	initial := mv.WrapText()

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'W'}}
	mv.HandleEvent(ev)

	if mv.WrapText() == initial {
		t.Errorf("WrapText() = %v after 'W', want %v (toggled)", mv.WrapText(), !initial)
	}
}

// TestMarkdownViewer_KeyWResetsDeltaX verifies 'w' toggling resets deltaX.
// Req 33: "W (rune 'w' or 'W'): toggle wrapText, reset deltaX to 0"
func TestMarkdownViewer_KeyWResetsDeltaX(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)
	mv.deltaX = 15

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'w'}}
	mv.HandleEvent(ev)

	if mv.deltaX != 0 {
		t.Errorf("deltaX = %d after 'w', want 0", mv.deltaX)
	}
}

// TestMarkdownViewer_UnrecognizedKeyPassesThrough verifies unknown keys are not
// consumed.
// Req 34: "Unrecognized keys pass through (event not cleared)"
func TestMarkdownViewer_UnrecognizedKeyPassesThrough(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'x'}}
	mv.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("unrecognized key 'x' was consumed, should pass through")
	}
}

// TestMarkdownViewer_KeyWhenNotFocusedPassesThrough verifies keyboard events
// pass through when SfSelected is not set.
// Req 35: "When NOT SfSelected, keyboard events pass through (not consumed)"
func TestMarkdownViewer_KeyWhenNotFocusedPassesThrough(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	// SfSelected is NOT set (default is not focused)
	mv.deltaY = 5

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	mv.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Up key was consumed when not focused, should pass through")
	}
	// deltaY should not have changed
	if mv.deltaY != 5 {
		t.Errorf("deltaY changed from 5 to %d when not focused, should remain 5", mv.deltaY)
	}
}

// =============================================================================
// HandleEvent mouse tests (requirements 36-41)
// =============================================================================

// TestMarkdownViewer_MouseCallsBaseViewHandleEvent verifies that mouse events
// first call BaseView.HandleEvent for click-to-focus.
// Req 36: "Calls BaseView.HandleEvent first (click-to-focus)"
// Req 37: "If BaseView cleared the event, returns immediately"
func TestMarkdownViewer_MouseCallsBaseViewHandleEvent(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	owner := &mdMockContainer{}
	mv.SetOwner(owner)
	// MarkdownViewer is NOT focused but IS selectable
	// A mouse click should transfer focus via BaseView.HandleEvent

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 1, Y: 1, Button: tcell.Button1},
	}
	mv.HandleEvent(ev)

	if owner.focusedChild != mv {
		t.Error("BaseView.HandleEvent was not called — click did not transfer focus")
	}
}

// TestMarkdownViewer_BaseViewClickDoesNotClearWithOfFirstClick verifies that
// when OfFirstClick is set, the clicking event is not cleared by BaseView.
// Req 37: "If BaseView cleared the event, returns immediately" — with OfFirstClick,
//         the event IS NOT cleared, so scroll processing continues
func TestMarkdownViewer_BaseViewClickDoesNotClearWithOfFirstClick(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	owner := &mdMockContainer{}
	mv.SetOwner(owner)
	mv.scheme = theme.BorlandBlue

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 1, Y: 1, Button: tcell.Button1},
	}
	mv.HandleEvent(ev)

	// With OfFirstClick set, BaseView does NOT clear the event
	if ev.IsCleared() {
		t.Error("event was cleared by BaseView; with OfFirstClick it should NOT be cleared")
	}
}

// TestMarkdownViewer_WheelUpDecrementsDeltaY verifies WheelUp scrolls content up
// (decreases deltaY).
// Req 38: "WheelUp: deltaY -= 3, clamped to 0, syncScrollBars, consumes event"
func TestMarkdownViewer_WheelUpDecrementsDeltaY(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetMarkdown("a\n\nb\n\nc\n\nd\n\ne\n\nf\n\ng\n\nh\n\ni")
	mv.deltaY = 10

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.WheelUp},
	}
	mv.HandleEvent(ev)

	if mv.deltaY != 7 {
		t.Errorf("deltaY after WheelUp = %d, want 7", mv.deltaY)
	}
	if !ev.IsCleared() {
		t.Error("WheelUp did not consume event")
	}
}

// TestMarkdownViewer_WheelUpClampedToZero verifies WheelUp clamped to 0.
// Req 38: "WheelUp: deltaY -= 3, clamped to 0"
func TestMarkdownViewer_WheelUpClampedToZero(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.deltaY = 1

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.WheelUp},
	}
	mv.HandleEvent(ev)

	if mv.deltaY != 0 {
		t.Errorf("deltaY after WheelUp from 1 = %d, want 0 (clamped)", mv.deltaY)
	}
}

// TestMarkdownViewer_WheelDownIncrementsDeltaY verifies WheelDown scrolls content
// down (increases deltaY).
// Req 39: "WheelDown: deltaY += 3, clamped to max, syncScrollBars, consumes event"
func TestMarkdownViewer_WheelDownIncrementsDeltaY(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetMarkdown("a\n\nb\n\nc\n\nd\n\ne\n\nf\n\ng\n\nh\n\ni")
	mv.deltaY = 0

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.WheelDown},
	}
	mv.HandleEvent(ev)

	if mv.deltaY != 3 {
		t.Errorf("deltaY after WheelDown = %d, want 3", mv.deltaY)
	}
	if !ev.IsCleared() {
		t.Error("WheelDown did not consume event")
	}
}

// TestMarkdownViewer_WheelDownClampedToMax verifies WheelDown clamped to max.
// Req 39: "WheelDown: deltaY += 3, clamped to max"
func TestMarkdownViewer_WheelDownClampedToMax(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 5))
	// Content tall enough to require scrolling
	mv.SetMarkdown("line1\n\nline2\n\nline3\n\nline4\n\nline5\n\nline6\n\nline7\n\nline8\n\nline9\n\nline10")

	// Find the maximum scroll position by pressing End
	mv.SetState(SfSelected, true)
	endEv := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd}}
	mv.HandleEvent(endEv)
	maxScroll := mv.deltaY

	if maxScroll <= 0 {
		t.Fatal("content does not overflow viewport; expected maxScroll > 0")
	}

	// Set deltaY to maxScroll and fire WheelDown — it must stay at max
	mv.deltaY = maxScroll

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.WheelDown},
	}
	mv.HandleEvent(ev)

	if mv.deltaY != maxScroll {
		t.Errorf("deltaY after WheelDown at max = %d, want clamped to maxScroll %d", mv.deltaY, maxScroll)
	}
	if !ev.IsCleared() {
		t.Error("WheelDown did not consume event")
	}
}

// TestMarkdownViewer_WheelScrollUpdatesScrollbar verifies that wheel scrolling
// updates the bound scrollbar via syncScrollBars.
// Req 39: "WheelDown: ... syncScrollBars"
func TestMarkdownViewer_WheelScrollUpdatesScrollbar(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetMarkdown("a\n\nb\n\nc\n\nd\n\ne\n\nf\n\ng\n\nh\n\ni")

	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	mv.SetVScrollBar(vsb)

	// Scroll to a known position
	mv.deltaY = 5

	// Fire WheelDown — syncScrollBars should update the scrollbar
	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.WheelDown},
	}
	mv.HandleEvent(ev)

	if vsb.Value() != mv.deltaY {
		t.Errorf("scrollbar Value = %d after WheelDown, want deltaY = %d", vsb.Value(), mv.deltaY)
	}
}

// TestMarkdownViewer_WheelLeftDecrementsDeltaX verifies WheelLeft decreases deltaX.
// Req 40: "WheelLeft: deltaX -= 3, clamped to 0, syncScrollBars, consumes event"
func TestMarkdownViewer_WheelLeftDecrementsDeltaX(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetWrapText(false)
	mv.deltaX = 10

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.WheelLeft},
	}
	mv.HandleEvent(ev)

	if mv.deltaX != 7 {
		t.Errorf("deltaX after WheelLeft = %d, want 7", mv.deltaX)
	}
	if !ev.IsCleared() {
		t.Error("WheelLeft did not consume event")
	}
}

// TestMarkdownViewer_WheelRightIncrementsDeltaX verifies WheelRight increases deltaX.
// Req 41: "WheelRight: deltaX += 3, clamped to max, syncScrollBars, consumes event"
func TestMarkdownViewer_WheelRightIncrementsDeltaX(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetWrapText(false)
	mv.deltaX = 0

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{Button: tcell.WheelRight},
	}
	mv.HandleEvent(ev)

	if mv.deltaX != 3 {
		t.Errorf("deltaX after WheelRight = %d, want 3", mv.deltaX)
	}
	if !ev.IsCleared() {
		t.Error("WheelRight did not consume event")
	}
}

// =============================================================================
// Scrollbar binding tests (requirements 42-47)
// =============================================================================

// TestMarkdownViewer_SetVScrollBarBindsOnChange verifies SetVScrollBar links
// the scrollbar's OnChange to update deltaY.
// Req 42: "SetVScrollBar(sb) binds sb.OnChange to update deltaY, calls syncScrollBars()"
func TestMarkdownViewer_SetVScrollBarBindsOnChange(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetMarkdown("a\n\nb\n\nc\n\nd\n\ne\n\nf\n\ng")

	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	mv.SetVScrollBar(vsb)

	// Verify OnChange was set (non-nil after SetVScrollBar)
	if vsb.OnChange == nil {
		t.Error("SetVScrollBar did not set OnChange callback")
	}

	// Simulate scrollbar change -> should update deltaY
	vsb.OnChange(3)
	if mv.deltaY != 3 {
		t.Errorf("deltaY after OnChange(3) = %d, want 3", mv.deltaY)
	}
}

// TestMarkdownViewer_SetVScrollBarCallsSyncScrollBars verifies SetVScrollBar
// calls syncScrollBars to update scrollbar parameters.
// Req 42: "SetVScrollBar(sb) ... calls syncScrollBars()"
func TestMarkdownViewer_SetVScrollBarCallsSyncScrollBars(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetMarkdown("Line 1\n\nLine 2\n\nLine 3")

	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	mv.SetVScrollBar(vsb)

	// After SetVScrollBar, the scrollbar should have updated range/page/value
	if vsb.Max() == 0 {
		t.Error("syncScrollBars not called — vertical scrollbar Max is 0")
	}
}

// TestMarkdownViewer_SetVScrollBarNilUnbinds verifies SetVScrollBar(nil) does
// not panic and properly clears the scrollbar binding.
// Req 43: "SetVScrollBar(nil) unbinds without panic"
func TestMarkdownViewer_SetVScrollBarNilUnbinds(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	// First bind a scrollbar
	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	mv.SetVScrollBar(vsb)

	// Then unbind by passing nil
	mv.SetVScrollBar(nil)

	// Should not panic (reaching here means no panic)
	// Verify the scrollbar's OnChange is no longer called (safety check)
	if mv.vScrollBar != nil {
		t.Error("vScrollBar should be nil after SetVScrollBar(nil)")
	}
}

// TestMarkdownViewer_SetVScrollBarNilBeforeBind verifies SetVScrollBar(nil)
// without prior binding doesn't panic.
// Req 43: "SetVScrollBar(nil) unbinds without panic"
func TestMarkdownViewer_SetVScrollBarNilBeforeBind(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	// Should not panic when no scrollbar was previously set
	mv.SetVScrollBar(nil)
}

// TestMarkdownViewer_SetHScrollBarBindsOnChange verifies SetHScrollBar links
// the scrollbar's OnChange to update deltaX.
// Req 44: "SetHScrollBar(sb) binds sb.OnChange to update deltaX, calls syncScrollBars()"
func TestMarkdownViewer_SetHScrollBarBindsOnChange(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetWrapText(false)
	mv.SetMarkdown("A somewhat longer line of text that might scroll horizontally")

	hsb := NewScrollBar(NewRect(0, 0, 40, 1), Horizontal)
	mv.SetHScrollBar(hsb)

	// Verify OnChange was set
	if hsb.OnChange == nil {
		t.Error("SetHScrollBar did not set OnChange callback")
	}

	// Simulate scrollbar change -> should update deltaX
	hsb.OnChange(5)
	if mv.deltaX != 5 {
		t.Errorf("deltaX after OnChange(5) = %d, want 5", mv.deltaX)
	}
}

// TestMarkdownViewer_SetHScrollBarCallsSyncScrollBars verifies SetHScrollBar
// calls syncScrollBars to update scrollbar parameters.
// Req 44: "SetHScrollBar(sb) ... calls syncScrollBars()"
func TestMarkdownViewer_SetHScrollBarCallsSyncScrollBars(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetWrapText(false)
	mv.SetMarkdown("A long line of text that extends beyond the viewport width horizontally")

	hsb := NewScrollBar(NewRect(0, 0, 40, 1), Horizontal)
	mv.SetHScrollBar(hsb)

	// After SetHScrollBar, the scrollbar should have its range set
	if hsb.Max() == 0 {
		t.Error("syncScrollBars not called — horizontal scrollbar Max is 0")
	}
}

// TestMarkdownViewer_SetHScrollBarNilUnbinds verifies SetHScrollBar(nil)
// unbinds without panic.
// Req 45: "SetHScrollBar(nil) unbinds without panic"
func TestMarkdownViewer_SetHScrollBarNilUnbinds(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	// First bind
	hsb := NewScrollBar(NewRect(0, 0, 40, 1), Horizontal)
	mv.SetHScrollBar(hsb)

	// Then unbind
	mv.SetHScrollBar(nil)

	if mv.hScrollBar != nil {
		t.Error("hScrollBar should be nil after SetHScrollBar(nil)")
	}
}

// TestMarkdownViewer_SyncScrollBarsUpdatesVerticalRange verifies syncScrollBars
// sets vertical scrollbar range, page size, and value.
// Req 46: "syncScrollBars updates vertical scrollbar range, page size, and value"
func TestMarkdownViewer_SyncScrollBarsUpdatesVerticalRange(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 8))
	mv.SetMarkdown("a\n\nb\n\nc\n\nd\n\ne\n\nf\n\ng\n\nh\n\ni\n\nj")

	vsb := NewScrollBar(NewRect(0, 0, 1, 8), Vertical)
	mv.SetVScrollBar(vsb)

	if vsb.Max() <= 0 {
		t.Error("vertical scrollbar Max not updated by syncScrollBars")
	}
	// PageSize should reflect viewport height
	if vsb.PageSize() == 0 {
		t.Error("vertical scrollbar PageSize not updated by syncScrollBars")
	}
	// Value should reflect current deltaY
	if vsb.Value() != mv.deltaY {
		t.Errorf("vertical scrollbar Value = %d, want deltaY = %d", vsb.Value(), mv.deltaY)
	}
}

// TestMarkdownViewer_SyncScrollBarsUpdatesHorizontalRange verifies syncScrollBars
// sets horizontal scrollbar range, page size, and value.
// Req 47: "syncScrollBars updates horizontal scrollbar range, page size, and value"
func TestMarkdownViewer_SyncScrollBarsUpdatesHorizontalRange(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 30, 10))
	mv.SetWrapText(false)
	mv.SetMarkdown("A very long line of text that exceeds the thirty character viewport width")

	hsb := NewScrollBar(NewRect(0, 0, 30, 1), Horizontal)
	mv.SetHScrollBar(hsb)

	if hsb.Max() <= 0 {
		t.Error("horizontal scrollbar Max not updated by syncScrollBars")
	}
	if hsb.PageSize() == 0 {
		t.Error("horizontal scrollbar PageSize not updated by syncScrollBars")
	}
	if hsb.Value() != mv.deltaX {
		t.Errorf("horizontal scrollbar Value = %d, want deltaX = %d", hsb.Value(), mv.deltaX)
	}
}

// TestMarkdownViewer_SyncScrollBarsSetsHorizontalZeroWhenWrapText verifies
// horizontal scrollbar range is 0 when wrapText is true (no horizontal scroll).
// Req 47: "syncScrollBars updates horizontal scrollbar range" (edge: wrapText=true)
func TestMarkdownViewer_SyncScrollBarsSetsHorizontalZeroWhenWrapText(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 30, 10))
	mv.SetWrapText(true) // wrapping on, no horizontal scroll needed
	mv.SetMarkdown("A very long line of text that would normally extend beyond the viewport")

	hsb := NewScrollBar(NewRect(0, 0, 30, 1), Horizontal)
	mv.SetHScrollBar(hsb)

	// With wrapText=true, horizontal scroll max should be 0 since content wraps
	// This depends on syncScrollBars handling the wrapText case
	if hsb.Value() != 0 {
		t.Errorf("horizontal scrollbar Value = %d, want 0 (wrapText=true, no h-scroll)", hsb.Value())
	}
}

// =============================================================================
// SetBounds test (requirement 48)
// =============================================================================

// TestMarkdownViewer_SetBoundsCallsSyncScrollBars verifies SetBounds triggers
// scrollbar sync after delegating to BaseView.
// Req 48: "SetBounds calls BaseView.SetBounds then syncScrollBars()"
func TestMarkdownViewer_SetBoundsCallsSyncScrollBars(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	mv.SetMarkdown("a\n\nb\n\nc\n\nd\n\ne\n\nf\n\ng")

	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	mv.SetVScrollBar(vsb)

	oldMax := vsb.Max()

	// Changing bounds changes viewport height, which changes scrollbar range
	mv.SetBounds(NewRect(0, 0, 40, 5))

	if vsb.Max() == oldMax {
		t.Error("SetBounds did not call syncScrollBars — vertical scrollbar range unchanged")
	}
}

// TestMarkdownViewer_SetBoundsDelegatesToBaseView verifies SetBounds correctly
// updates the bounds via BaseView.SetBounds.
// Req 48: "SetBounds calls BaseView.SetBounds"
func TestMarkdownViewer_SetBoundsDelegatesToBaseView(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	r := NewRect(5, 3, 50, 20)

	mv.SetBounds(r)

	if mv.Bounds() != r {
		t.Errorf("Bounds() after SetBounds(%v) = %v, want %v", r, mv.Bounds(), r)
	}
}

// =============================================================================
// SetState tests (requirements 49-50)
// =============================================================================

// TestMarkdownViewer_SetStateFocusLossHidesScrollBars verifies that losing focus
// (SfSelected cleared) hides bound scrollbars.
// Req 49: "When SfSelected changes to false, hides bound scrollbars"
func TestMarkdownViewer_SetStateFocusLossHidesScrollBars(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	hsb := NewScrollBar(NewRect(0, 0, 40, 1), Horizontal)
	mv.SetVScrollBar(vsb)
	mv.SetHScrollBar(hsb)

	// Initially focused: scrollbars should be visible
	mv.SetState(SfSelected, true)
	if !vsb.HasState(SfVisible) {
		t.Error("vertical scrollbar not visible when focused")
	}
	if !hsb.HasState(SfVisible) {
		t.Error("horizontal scrollbar not visible when focused")
	}

	// Lose focus: scrollbars should be hidden
	mv.SetState(SfSelected, false)
	if vsb.HasState(SfVisible) {
		t.Error("vertical scrollbar still visible after losing focus")
	}
	if hsb.HasState(SfVisible) {
		t.Error("horizontal scrollbar still visible after losing focus")
	}
}

// TestMarkdownViewer_SetStateFocusGainShowsScrollBars verifies that gaining focus
// (SfSelected set) shows bound scrollbars.
// Req 50: "When SfSelected changes to true, shows bound scrollbars"
func TestMarkdownViewer_SetStateFocusGainShowsScrollBars(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	vsb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	hsb := NewScrollBar(NewRect(0, 0, 40, 1), Horizontal)
	mv.SetVScrollBar(vsb)
	mv.SetHScrollBar(hsb)

	// Start with scrollbars hidden (not focused)
	mv.SetState(SfSelected, false)
	vsb.SetState(SfVisible, false)
	hsb.SetState(SfVisible, false)

	// Gain focus: scrollbars should be shown
	mv.SetState(SfSelected, true)
	if !vsb.HasState(SfVisible) {
		t.Error("vertical scrollbar not visible after gaining focus")
	}
	if !hsb.HasState(SfVisible) {
		t.Error("horizontal scrollbar not visible after gaining focus")
	}
}

// TestMarkdownViewer_SetStateOnlyAffectsWhenScrollbarsBound verifies SetState
// does not panic when scrollbars are nil.
// Req 49-50: Edge case — scrollbars may be nil
func TestMarkdownViewer_SetStateOnlyAffectsWhenScrollbarsBound(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))
	// No scrollbars bound

	// Should not panic
	mv.SetState(SfSelected, true)
	mv.SetState(SfSelected, false)
}

// TestMarkdownViewer_SetStateDelegatesToBaseView verifies SetState still
// delegates to BaseView.SetState for the flag itself.
// Req 49-50: SetState must call BaseView.SetState (implicit from embedding)
func TestMarkdownViewer_SetStateDelegatesToBaseView(t *testing.T) {
	mv := NewMarkdownViewer(NewRect(0, 0, 40, 10))

	mv.SetState(SfSelected, true)
	if !mv.HasState(SfSelected) {
		t.Error("SetState(SfSelected, true) did not set the flag via BaseView")
	}

	mv.SetState(SfSelected, false)
	if mv.HasState(SfSelected) {
		t.Error("SetState(SfSelected, false) did not clear the flag via BaseView")
	}
}
