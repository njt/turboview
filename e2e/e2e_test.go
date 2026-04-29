package e2e

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func containsAny(lines []string, substrs ...string) bool {
	for _, line := range lines {
		for _, s := range substrs {
			if strings.Contains(line, s) {
				return true
			}
		}
	}
	return false
}

func projectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..")
}

func buildBasicApp(t *testing.T) string {
	t.Helper()
	root := projectRoot()
	binPath := filepath.Join(root, "e2e", "testapp", "basic", "basic")
	out, err := exec.Command("go", "build", "-o", binPath, filepath.Join(root, "e2e", "testapp", "basic")).CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	t.Cleanup(func() { exec.Command("rm", binPath).Run() })
	return binPath
}

func TestBasicAppBoot(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-basic"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	lines := tmuxCapture(t, session)

	// Desktop pattern visible
	desktopHasPattern := false
	for _, line := range lines {
		if strings.Contains(line, "░") {
			desktopHasPattern = true
			break
		}
	}
	if !desktopHasPattern {
		t.Error("desktop background pattern '░' not found")
	}

	// Window frame characters visible
	frameFound := false
	for _, line := range lines {
		if strings.Contains(line, "╔") || strings.Contains(line, "═") {
			frameFound = true
			break
		}
	}
	if !frameFound {
		t.Error("window frame characters not found")
	}

	// Window title visible
	titleFound := false
	for _, line := range lines {
		if strings.Contains(line, "File Manager") || strings.Contains(line, "Editor") {
			titleFound = true
			break
		}
	}
	if !titleFound {
		t.Error("window title text not found")
	}

	// Button text visible inside window
	buttonFound := false
	for _, line := range lines {
		if strings.Contains(line, "OK") || strings.Contains(line, "Close") {
			buttonFound = true
			break
		}
	}
	if !buttonFound {
		t.Error("button text 'OK' or 'Close' not found in rendered output")
	}

	// Status line contains "Alt+X"
	statusFound := false
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			if strings.Contains(lines[i], "Alt+X") {
				statusFound = true
			}
			break
		}
	}
	if !statusFound {
		t.Error("status line should contain 'Alt+X'")
	}

	// Alt+X exits the app
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

func TestDialogFlow(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-dialog"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// win2 is focused on startup; switch to win1 (HelpCtx=1) so F2 is active
	tmuxSendKeys(t, session, "M-1")

	// Press F2 to open dialog
	tmuxSendKeys(t, session, "F2")
	lines := tmuxCapture(t, session)

	// Dialog title "Confirm" should appear
	dialogTitleFound := false
	for _, line := range lines {
		if strings.Contains(line, "Confirm") {
			dialogTitleFound = true
			break
		}
	}
	if !dialogTitleFound {
		t.Error("dialog title 'Confirm' not found after F2")
	}

	// Dialog double-line frame should appear
	dialogFrameFound := false
	for _, line := range lines {
		if strings.Contains(line, "╔") {
			dialogFrameFound = true
			break
		}
	}
	if !dialogFrameFound {
		t.Error("dialog frame character '╔' not found after F2")
	}

	// Press Enter to dismiss dialog (Yes is default → CmQuit)
	tmuxSendKeys(t, session, "Enter")

	// App should exit because Yes → CmQuit
	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after confirming dialog with Enter")
	}
}

func TestDialogDismissNoQuit(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-dialog-no"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// win2 is focused on startup; switch to win1 (HelpCtx=1) so F2 is active
	tmuxSendKeys(t, session, "M-1")

	// Press F2 to open dialog
	tmuxSendKeys(t, session, "F2")

	// Press Tab to move focus from Yes to No
	tmuxSendKeys(t, session, "Tab")

	// Press Enter on No button → dialog dismissed but app stays
	tmuxSendKeys(t, session, "Enter")
	lines := tmuxCapture(t, session)

	// Dialog should be gone (no "Confirm" title)
	dialogGone := true
	for _, line := range lines {
		if strings.Contains(line, "Confirm") {
			dialogGone = false
			break
		}
	}
	if !dialogGone {
		t.Error("dialog 'Confirm' still visible after pressing No")
	}

	// Desktop pattern should still be visible (app is running)
	desktopVisible := false
	for _, line := range lines {
		if strings.Contains(line, "░") {
			desktopVisible = true
			break
		}
	}
	if !desktopVisible {
		t.Error("desktop pattern not visible after dismissing dialog — app may have crashed")
	}

	// Clean exit with Alt+X
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

func TestMenuBarVisible(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-menubar"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	lines := tmuxCapture(t, session)

	// First row should contain "File" and "Window"
	firstRow := ""
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			firstRow = line
			break
		}
	}
	if !strings.Contains(firstRow, "File") {
		t.Errorf("menu bar first row should contain 'File', got: %q", firstRow)
	}
	if !strings.Contains(firstRow, "Window") {
		t.Errorf("menu bar first row should contain 'Window', got: %q", firstRow)
	}

	// Alt+X exits cleanly
	tmuxSendKeys(t, session, "M-x")
	exitedClean := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exitedClean = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exitedClean {
		t.Error("app did not exit after Alt+X")
	}
}

func TestMenuOpenAndSelect(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-menuopen"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// F10 activates menu bar, Enter opens the first (File) menu
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// File menu items should be visible
	if !containsAny(lines, "New") {
		t.Error("menu item 'New' not found after opening File menu")
	}
	if !containsAny(lines, "Open...") {
		t.Error("menu item 'Open...' not found after opening File menu")
	}

	// Escape twice to dismiss
	tmuxSendKeys(t, session, "Escape")
	time.Sleep(300 * time.Millisecond)
	tmuxSendKeys(t, session, "Escape")
	time.Sleep(500 * time.Millisecond)

	lines = tmuxCapture(t, session)

	// Desktop pattern should be visible after dismissal
	if !containsAny(lines, "░") {
		t.Error("desktop pattern not visible after dismissing menu — app may have crashed")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	exitedMenu := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exitedMenu = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exitedMenu {
		t.Error("app did not exit after Alt+X")
	}
}

func TestMenuSelectExit(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-menuexit"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// F10 activates menu bar, Enter opens File menu, x triggers E~x~it → CmQuit
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "x")
	time.Sleep(500 * time.Millisecond)

	exitedViaMenu := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exitedViaMenu = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exitedViaMenu {
		t.Error("app did not exit after selecting Exit from File menu")
	}
}

func TestInputBoxFlow(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-inputbox"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// win2 is focused on startup; switch to win1 (HelpCtx=1) so F3 is active
	tmuxSendKeys(t, session, "M-1")

	// Press F3 to open the InputBox dialog
	tmuxSendKeys(t, session, "F3")
	time.Sleep(500 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// "Name:" prompt should be visible in the dialog
	if !containsAny(lines, "Name:") {
		t.Error("InputBox prompt 'Name:' not found after F3")
	}

	// Select all existing text (Ctrl+A) then type new filename as literal text
	tmuxSendKeys(t, session, "C-a")
	tmuxType(t, session, "test.go")
	// Press Enter — passes through InputLine to default OK button
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)

	lines = tmuxCapture(t, session)

	// Dialog should be gone — "Name:" no longer visible
	if containsAny(lines, "Name:") {
		t.Error("InputBox dialog still visible after pressing Enter")
	}

	// Static text should now show "File: test.go"
	if !containsAny(lines, "File: test.go") {
		t.Error("static text 'File: test.go' not found after confirming InputBox")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestCheckBoxVisible(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-checkbox"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	lines := tmuxCapture(t, session)

	// CheckBox indicators "[ ]" should be visible in win1
	if !containsAny(lines, "[ ]") {
		t.Error("checkbox indicator '[ ]' not found in rendered output")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestRadioButtonVisible(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-radio"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	lines := tmuxCapture(t, session)

	// First radio button is selected by default — "(*)" should be visible
	if !containsAny(lines, "(*)") {
		t.Error("radio button indicator '(*)' not found in rendered output")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestListViewerVisible(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-listview"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	lines := tmuxCapture(t, session)

	// win2 should contain "Item 1" from the ListViewer
	if !containsAny(lines, "Item 1") {
		t.Error("ListViewer item 'Item 1' not visible in win2")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestScrollBarVisible(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-scrollbar"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	lines := tmuxCapture(t, session)

	// Scrollbar should render arrow characters
	if !containsAny(lines, "▲", "▼") {
		t.Error("scrollbar arrow characters '▲' or '▼' not visible in win2")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestListViewerNavigation(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-listnav"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Click inside win2 to focus it (win2 is at col 20, row 5, size 40x12)
	// Click at approximately col 30, row 8 (middle of win2's client area)
	exec.Command("tmux", "send-keys", "-t", session, "Tab").Run()
	time.Sleep(500 * time.Millisecond)

	// Press Down arrow multiple times to scroll through the list
	for i := 0; i < 8; i++ {
		tmuxSendKeys(t, session, "Down")
	}
	time.Sleep(500 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// After scrolling down 8 times, later items should be visible
	if !containsAny(lines, "Item 9", "Item 10", "Item 8") {
		t.Error("later list items not visible after navigating down — scrolling may not work")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestListViewerDifferentTheme(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-listtheme"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	lines := tmuxCapture(t, session)

	// Smoke test: win2 list items are visible (different theme applied)
	if !containsAny(lines, "Item 1", "Item 2") {
		t.Error("list items not visible in win2 — custom theme may have broken rendering")
	}

	// win2 "Editor" title should be visible
	if !containsAny(lines, "Editor") {
		t.Error("win2 'Editor' title not found — window may not have rendered")
	}

	// Clean exit
	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestInputBoxCancel(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-inputcancel"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// win2 is focused on startup; switch to win1 (HelpCtx=1) so F3 is active
	tmuxSendKeys(t, session, "M-1")

	// Press F3 to open the InputBox dialog
	tmuxSendKeys(t, session, "F3")
	time.Sleep(500 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// Dialog should be open — "Name:" prompt visible
	if !containsAny(lines, "Name:") {
		t.Error("InputBox prompt 'Name:' not found after F3")
	}

	// Tab twice to move focus past the InputLine and OK button to the Cancel button,
	// then press Enter to activate Cancel (there is no Escape handler in the dialog).
	tmuxSendKeys(t, session, "Tab")
	tmuxSendKeys(t, session, "Tab")
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)

	lines = tmuxCapture(t, session)

	// Dialog should be gone
	if containsAny(lines, "Name:") {
		t.Error("InputBox dialog still visible after pressing Cancel")
	}

	// App should still be running — desktop pattern visible
	if !containsAny(lines, "░") {
		t.Error("desktop pattern not visible after cancelling InputBox — app may have crashed")
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

func TestHelpContextFiltering(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-helpctx"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// win2 gets focus last (inserted after win1), so it is the focused window at startup.
	lines := tmuxCapture(t, session)

	// Alt+X is a global item (no HelpCtx) — must always be visible
	if !containsAny(lines, "Alt+X") {
		t.Error("status line should always show 'Alt+X' (global item)")
	}

	// win2 has HelpCtx=2 so F4 Search should be visible
	if !containsAny(lines, "Search") {
		t.Error("status line should show 'Search' when win2 (HelpCtx=2) is focused")
	}

	// F2 Dialog is HelpCtx=1 — should NOT appear when win2 is focused
	if containsAny(lines, "Dialog") {
		t.Error("status line should NOT show 'Dialog' when win2 (HelpCtx=2) is focused")
	}

	// Switch to win1 using Alt+1
	tmuxSendKeys(t, session, "M-1")

	lines = tmuxCapture(t, session)

	// win1 has HelpCtx=1 so F2 Dialog should now be visible
	if !containsAny(lines, "Dialog") {
		t.Error("status line should show 'Dialog' when win1 (HelpCtx=1) is focused")
	}

	// F4 Search is HelpCtx=2 — should NOT appear when win1 is focused
	if containsAny(lines, "Search") {
		t.Error("status line should NOT show 'Search' when win1 (HelpCtx=1) is focused")
	}

	// Alt+X exits cleanly
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

func TestThemeRegistration(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-themes"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// If any theme init() panicked the app would not have started; verify it booted.
	lines := tmuxCapture(t, session)

	if !containsAny(lines, "░") {
		t.Error("desktop background pattern '░' not found — app may have failed to start")
	}

	// Alt+X exits cleanly
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

func TestTerminalResize(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-resize"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Verify initial state
	lines := tmuxCapture(t, session)
	if !containsAny(lines, "░") {
		t.Fatal("desktop pattern not visible at startup")
	}

	// Resize the tmux pane to 100x30
	err := exec.Command("tmux", "resize-window", "-t", session, "-x", "100", "-y", "30").Run()
	if err != nil {
		t.Fatalf("tmux resize-window failed: %v", err)
	}
	time.Sleep(1 * time.Second) // give app time to handle resize

	// Capture after resize
	lines = tmuxCapture(t, session)

	// Desktop pattern should fill the wider area
	if !containsAny(lines, "░") {
		t.Error("desktop pattern not visible after resize")
	}

	// Window titles should still be visible
	if !containsAny(lines, "File Manager") {
		t.Error("'File Manager' window title not visible after resize")
	}
	if !containsAny(lines, "Editor") {
		t.Error("'Editor' window title not visible after resize")
	}

	// App should still be responsive — Alt+X exits cleanly
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
		t.Error("app did not exit after Alt+X following resize")
	}
}

func TestDialogEscapeDismiss(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-dlg-esc"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Switch to win1 (HelpCtx=1) so F2 is active
	tmuxSendKeys(t, session, "M-1")

	// Press F2 to open the "Confirm" dialog
	tmuxSendKeys(t, session, "F2")
	lines := tmuxCapture(t, session)

	if !containsAny(lines, "Confirm") {
		t.Fatal("dialog title 'Confirm' not found after F2")
	}

	// Press Escape — Dialog.HandleEvent converts KeyEscape → CmCancel → ExecView exits
	tmuxSendKeys(t, session, "Escape")
	lines = tmuxCapture(t, session)

	// Dialog should be gone
	if containsAny(lines, "Confirm") {
		t.Error("dialog 'Confirm' still visible after pressing Escape — Escape should dismiss the dialog")
	}

	// App should still be running (Escape = cancel, not quit)
	if !containsAny(lines, "░") {
		t.Error("desktop pattern not visible after Escape-dismissing dialog — app may have crashed")
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

func TestTabFocusNavigation(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-tabfocus"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Switch to win1 which has buttons, checkboxes, and radio buttons
	tmuxSendKeys(t, session, "M-1")

	// Tab cycles through win1's widgets. When a button gets focus it renders ►.
	tmuxSendKeys(t, session, "Tab")
	lines := tmuxCapture(t, session)

	focusIndicator := false
	for _, line := range lines {
		if strings.Contains(line, "►") {
			focusIndicator = true
			break
		}
	}
	if !focusIndicator {
		t.Error("focus indicator '►' not found after Tab in win1 — Tab focus navigation may not be working")
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

func TestContextMenuSmoke(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-ctxmenu"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Verify app boots and exits cleanly with ContextMenu code compiled in
	lines := tmuxCapture(t, session)
	if !containsAny(lines, "░") {
		t.Error("desktop pattern not visible at startup")
	}

	// Alt+X exits cleanly
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
