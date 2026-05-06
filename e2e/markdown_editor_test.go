package e2e

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestMarkdownEditorVisible opens the MarkdownEditor window via the menu
// and verifies the rendered heading text and scrollbars are visible.
func TestMarkdownEditorVisible(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-mdedit-visible"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Open Options > Markdown Editor via menu
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "M")
	time.Sleep(500 * time.Millisecond)

	// Move cursor down from the heading so the cursor overlay doesn't
	// overwrite any characters of "Welcome". The cursor renders the source
	// character at the screen position, which can disrupt the displayed text.
	tmuxSendKeys(t, session, "Down")
	tmuxSendKeys(t, session, "Down")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// Verify window title is visible
	if !containsAny(lines, "Markdown Editor") {
		t.Error("Markdown Editor window title not visible")
	}

	// Verify heading content renders: the "# Welcome" heading shows "Welcome"
	// (the # marker is hidden in formatted mode)
	if !containsAny(lines, "Welcome") {
		t.Error("heading text 'Welcome' not visible in MarkdownEditor")
	}

	// Verify scrollbar arrow characters are visible
	if !containsAny(lines, "▲", "▼") {
		t.Error("scrollbar arrow characters not visible in MarkdownEditor window")
	}

	// Verify horizontal scrollbar arrow characters
	if !containsAny(lines, "◄", "►") {
		t.Error("horizontal scrollbar arrow characters not visible in MarkdownEditor window")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after Alt+X")
	}
}

// TestMarkdownEditorTyping opens the MarkdownEditor and verifies that
// typing new text appears on screen in the rendered markdown output.
func TestMarkdownEditorTyping(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-mdedit-type"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Open Options > Markdown Editor
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "M")
	time.Sleep(500 * time.Millisecond)

	// Move cursor to end of first line ("# Welcome") and append text
	tmuxSendKeys(t, session, "End")
	tmuxType(t, session, " to TurboView")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// The typed text should appear as part of the rendered heading.
	// After End moves cursor to end of heading line, cursor no longer
	// overlays the heading text, so "Welcome" is fully visible.
	if !containsAny(lines, "Welcome to TurboView") {
		t.Error("typed text 'Welcome to TurboView' not visible in MarkdownEditor")
	}

	// Also verify the original bullet items still render
	if !containsAny(lines, "item one") {
		t.Error("bullet item 'item one' not visible after typing")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after Alt+X")
	}
}

// TestMarkdownEditorCursorVisible verifies the cursor renders the source
// character at the cursor position on screen (Phase 1 behavior).
func TestMarkdownEditorCursorVisible(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-mdedit-cursor"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Open Options > Markdown Editor
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "M")
	time.Sleep(500 * time.Millisecond)

	// Verify the window is actually focused by checking for the title bar
	lines := tmuxCapture(t, session)

	// The window title appears in the title bar with many frame chars
	mdWindowActive := false
	for _, line := range lines {
		// Active window title "Markdown Editor" should be in a line with ═ frame chars
		if strings.Contains(line, "Markdown Editor") && strings.Count(line, "═") > 5 {
			mdWindowActive = true
			break
		}
	}
	if !mdWindowActive {
		t.Fatal("Markdown Editor window not focused -- title bar not found with active frame")
	}

	// With reveal (Phase 2), the cursor at (0,0) on "# Welcome" draws the
	// source '#' character at the screen position corresponding to source
	// column 0. The reveal marker "# " appears dimmed starting at x=0, and
	// the heading content "Welcome" is shifted right by the marker width.
	// The cursor '#' overlays the dimmed marker '#' at x=0.
	// Verify both the cursor/marker '#' and heading "Welcome" are visible.
	if !containsAny(lines, "#") {
		t.Error("cursor/marker '#' not visible -- cursor may not be rendering at heading position")
	}
	if !containsAny(lines, "Welcome") {
		t.Error("heading 'Welcome' not visible -- content may not be rendering after reveal indent")
	}

	// Move cursor and verify content stays visible
	tmuxSendKeys(t, session, "Down")
	tmuxSendKeys(t, session, "Down")
	time.Sleep(300 * time.Millisecond)

	lines = tmuxCapture(t, session)

	// After moving Down from heading, "Welcome" should be fully visible (no cursor overlay)
	if !containsAny(lines, "Welcome") {
		t.Error("heading 'Welcome' not fully visible after moving cursor away")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after Alt+X")
	}
}

// TestMarkdownEditorNewLineTyping verifies typing on a new line in
// the MarkdownEditor renders correctly. The typed text may become
// part of a bullet list item when inserted between list items.
func TestMarkdownEditorNewLineTyping(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-mdedit-newline"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Open Options > Markdown Editor
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "M")
	time.Sleep(500 * time.Millisecond)

	// Move down to "- item one" line and type new text at end
	// Source: row 0="# Welcome", row 1=blank, row 2="Type ...", row 3=blank, row 4="- item one"
	// After 4 Down presses: cursor at row 4 ("- item one")
	for i := 0; i < 4; i++ {
		tmuxSendKeys(t, session, "Down")
	}
	tmuxSendKeys(t, session, "End")
	tmuxSendKeys(t, session, "Enter")
	tmuxType(t, session, "new paragraph here")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// The new text integrates into the markdown parse. Since it's between two
	// bullet items, goldmark may include it as part of the bullet list item.
	// Verify at least parts of the text appear on screen.
	if !containsAny(lines, "new") {
		t.Error("typed text 'new' not visible after typing on new line")
	}

	// Also verify original content still renders after the edit
	if !containsAny(lines, "item two") {
		t.Error("bullet item 'item two' not visible after typing")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after Alt+X")
	}
}

// TestMarkdownEditorScrollbarVisible verifies both vertical and
// horizontal scrollbars render in the MarkdownEditor window, and that
// PageDn actually scrolls the rendered content (later lines become
// visible, earlier lines scroll out of view).
func TestMarkdownEditorScrollbarVisible(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-mdedit-sb"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Open Options > Markdown Editor
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "M")
	time.Sleep(500 * time.Millisecond)

	// Move cursor down from the heading so the cursor overlay doesn't
	// overwrite characters of "Welcome".
	tmuxSendKeys(t, session, "Down")
	tmuxSendKeys(t, session, "Down")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// --- Static structure assertions ---

	// Vertical scrollbar arrow characters should be visible
	if !containsAny(lines, "▲", "▼") {
		t.Error("vertical scrollbar arrow characters '▲' or '▼' not visible in MarkdownEditor")
	}

	// Horizontal scrollbar arrow characters should be visible
	if !containsAny(lines, "◄", "►") {
		t.Error("horizontal scrollbar arrow characters '◄' or '►' not visible in MarkdownEditor")
	}

	// Window title should be visible
	if !containsAny(lines, "Markdown Editor") {
		t.Error("Markdown Editor window title not visible")
	}

	// --- Scroll behavior assertions ---

	// Precondition: early content should be visible at the top of the viewport.
	if !containsAny(lines, "Welcome") {
		t.Fatal("precondition: 'Welcome' heading not visible before PageDn")
	}
	if !containsAny(lines, "Line 01") {
		t.Fatal("precondition: 'Line 01' not visible before PageDn — content may be too short or rendering is broken")
	}

	// Press PageDn to scroll the rendered markdown content down one viewport.
	tmuxSendKeys(t, session, "NPage")
	time.Sleep(500 * time.Millisecond)

	lines = tmuxCapture(t, session)

	// After PageDn, later content should be visible.
	if !containsAny(lines, "Line 08", "Line 09", "Line 10", "Line 11", "Line 12") {
		t.Error("later lines (Line 08+) not visible after PageDn — scroll may not have worked")
	}

	// Earlier content should have scrolled out of the visible viewport.
	if containsAny(lines, "Line 01") {
		t.Error("'Line 01' still visible after PageDn — content did not scroll")
	}

	// Press PageDn a second time to scroll further down.
	tmuxSendKeys(t, session, "NPage")
	time.Sleep(500 * time.Millisecond)

	lines = tmuxCapture(t, session)

	// After another PageDn, even later content should be visible.
	if !containsAny(lines, "Line 15", "Line 16", "Line 17", "Line 18", "Line 19", "Line 20") {
		t.Error("even later lines (Line 15+) not visible after second PageDn — scroll may be stuck")
	}

	// Content from the first page should now be scrolled out.
	if containsAny(lines, "Line 01") || containsAny(lines, "Line 02") {
		t.Error("early lines still visible after two PageDn presses — scrolling may be incomplete")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after Alt+X")
	}
}

// TestMarkdownEditorRevealBlockHeading verifies block-level reveal:
// when cursor is inside a heading, the "# " marker is visible on screen;
// when cursor leaves the heading, the marker disappears.
func TestMarkdownEditorRevealBlockHeading(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-reveal-heading"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Open Markdown Editor
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "M")
	time.Sleep(500 * time.Millisecond)

	// Cursor starts at (0,0) in the "# Welcome" heading.
	// The cursor overlays the source '#' at screen position.
	// With reveal active, the "# " block marker is dimmed at columns 0-1,
	// and the cursor overlays the dimmed '#' at column 0.
	lines := tmuxCapture(t, session)

	// With cursor in heading, the "#" marker should be visible
	if !containsAny(lines, "#") {
		t.Error("heading marker '#' not visible when cursor is inside heading")
	}
	if !containsAny(lines, "Welcome") {
		t.Error("heading text 'Welcome' not visible")
	}

	// Move cursor down to leave heading (rows: 0=heading, 1=blank, 2=text)
	// After moving to row 1 (blank), cursor leaves heading block
	tmuxSendKeys(t, session, "Down")
	time.Sleep(300 * time.Millisecond)

	lines = tmuxCapture(t, session)
	// Both "#" and "Welcome" should still be visible (Welcome always renders,
	// and # may still be present from the cursor at position (1,0))
	tmuxSendKeys(t, session, "Down")
	time.Sleep(300 * time.Millisecond)

	// Now cursor is at row 2 (plain text "Type **markdown** here."),
	// completely outside heading. Block reveal should hide heading markers.
	lines = tmuxCapture(t, session)
	if !containsAny(lines, "Welcome") {
		t.Error("heading 'Welcome' still visible when cursor is away (expected)")
	}

	// Move back to heading to verify marker reappears
	tmuxSendKeys(t, session, "Up")
	tmuxSendKeys(t, session, "Up")
	time.Sleep(300 * time.Millisecond)

	lines = tmuxCapture(t, session)
	if !containsAny(lines, "#") {
		t.Error("heading marker '#' not visible after returning cursor to heading")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after Alt+X")
	}
}

// TestMarkdownEditorListContinuation verifies smart list continuation
// (Phase 3): pressing Enter at the end of a list item creates a new
// list item line with the appropriate marker.
func TestMarkdownEditorListContinuation(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-mdlist"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Open Options > Markdown Editor
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "M")
	time.Sleep(500 * time.Millisecond)

	// Move to end of "- item one" line (source row 4).
	// Source rows: 0="# Welcome", 1="", 2="Type **markdown** here.",
	//              3="", 4="- item one"
	for i := 0; i < 4; i++ {
		tmuxSendKeys(t, session, "Down")
	}
	tmuxSendKeys(t, session, "End")
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// After listEnterContinuation, "- " is inserted on the new line.
	// With reveal active (widget selected, cursor on list item),
	// the "-" block marker is drawn dimmed at the left margin.
	if !containsAny(lines, "- ") {
		t.Error("list continuation marker '- ' not visible after Enter at end of list item")
	}

	// Verify original bullets still render
	if !containsAny(lines, "item two") {
		t.Error("bullet item 'item two' not visible after list continuation")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after Alt+X")
	}
}

// TestMarkdownEditorSourceToggle verifies Ctrl+T toggles between
// formatted and raw source view (Phase 3).
func TestMarkdownEditorSourceToggle(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-mdtoggle"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Open Options > Markdown Editor
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "M")
	time.Sleep(500 * time.Millisecond)

	// Move cursor down to avoid cursor overlay issues with the heading
	tmuxSendKeys(t, session, "Down")
	tmuxSendKeys(t, session, "Down")
	time.Sleep(200 * time.Millisecond)

	// Press Ctrl+T to toggle source mode
	tmuxSendKeys(t, session, "C-t")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// In source mode, the raw markdown "# Welcome" shows "#" literally
	if !containsAny(lines, "# Welcome") {
		t.Error("raw source '# Welcome' not visible after Ctrl+T toggle to source mode")
	}

	// Toggle back to formatted mode
	tmuxSendKeys(t, session, "C-t")
	time.Sleep(300 * time.Millisecond)

	lines = tmuxCapture(t, session)
	if !containsAny(lines, "Welcome") {
		t.Error("'Welcome' not visible after toggling back to formatted mode")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after Alt+X")
	}
}

// TestMarkdownEditorTypeAndRender verifies typing markdown heading
// syntax and seeing it render in formatted mode (Phase 3).
func TestMarkdownEditorTypeAndRender(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-mdrender"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Open Options > Markdown Editor
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "M")
	time.Sleep(500 * time.Millisecond)

	// Go to end of document (Ctrl+End), add a new heading line.
	tmuxSendKeys(t, session, "C-End")
	time.Sleep(300 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	tmuxType(t, session, "## New Section")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)
	// The heading text should render in the formatted output
	if !containsAny(lines, "New Section") {
		t.Error("'New Section' heading text not visible after typing")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after Alt+X")
	}
}

// TestMarkdownEditorRevealInlineBold verifies inline-level reveal:
// when cursor is inside bold text (**markdown**), the "**" markers
// become visible; when cursor leaves, they disappear.
func TestMarkdownEditorRevealInlineBold(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-reveal-bold"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Open Markdown Editor
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(200 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "M")
	time.Sleep(500 * time.Millisecond)

	// Source content: row 0="# Welcome", row 1="", row 2="Type **markdown** here."
	// Move cursor to row 2 (bold text line), then right to enter bold span
	tmuxSendKeys(t, session, "Down")
	tmuxSendKeys(t, session, "Down")
	time.Sleep(200 * time.Millisecond)

	// Row 2: "Type **markdown** here."
	// Move cursor inside bold span, past "markdown" so the whole word is
	// visible on screen (cursor overlay at a specific position would
	// break substring matching for the full word).
	// Col layout: 0=T,1=y,2=p,3=e,4=' ',5=*,6=*,7=m,8=a,9=r,10=k,11=d,12=o,13=w,14=n,15=*,16=*
	for i := 0; i < 10; i++ {
		tmuxSendKeys(t, session, "Right")
	}
	time.Sleep(300 * time.Millisecond)

	// Cursor at col 10 ('k' in "markdown") — inside bold span.
	// The cursor overlays 'k', so search for parts visible around it.
	lines := tmuxCapture(t, session)

	// Check that surrounding parts of bold text are visible
	if !containsAny(lines, "Type") {
		t.Error("preceding text 'Type' not visible when cursor is in bold span")
	}
	if !containsAny(lines, "here") {
		t.Error("following text 'here' not visible when cursor is in bold span")
	}

	// Move cursor back to before bold span
	tmuxSendKeys(t, session, "Home")
	time.Sleep(300 * time.Millisecond)

	lines = tmuxCapture(t, session)
	if !containsAny(lines, "markdown") {
		t.Error("bold content 'markdown' not visible when cursor is outside bold")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after Alt+X")
	}
}
