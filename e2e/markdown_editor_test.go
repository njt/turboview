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

	// In Phase 1, cursor at (0,0) on "# Welcome" renders '#' at the screen
	// position where the heading text starts. This causes "Welcome" to appear
	// as "#elcome" (cursor overlays source '#' on the rendered 'W').
	// Verify this cursor-overlay behavior is present.
	if !containsAny(lines, "#elcome") {
		t.Error("cursor overlay pattern '#elcome' not visible -- cursor may not be rendering at heading position")
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
// horizontal scrollbars render in the MarkdownEditor window.
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

	lines := tmuxCapture(t, session)

	// Vertical scrollbar arrow characters should be visible
	if !containsAny(lines, "▲", "▼") {
		t.Error("vertical scrollbar arrow characters '▲' or '▼' not visible in MarkdownEditor")
	}

	// Horizontal scrollbar arrow characters should be visible
	if !containsAny(lines, "◄", "►") {
		t.Error("horizontal scrollbar arrow characters '◄' or '►' not visible in MarkdownEditor")
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
