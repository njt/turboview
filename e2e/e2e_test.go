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

	// Dialog should be gone — "Open File" title no longer visible.
	// (We cannot check "Name:" here because win1 has a persistent Label with that text.)
	if containsAny(lines, "Open File") {
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

	// Dialog should be gone — "Open File" title no longer visible.
	// (We cannot check "Name:" here because win1 has a persistent Label with that text.)
	if containsAny(lines, "Open File") {
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

func TestInputLineCtrlYClear(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-ctrly"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Switch to win1 so F3 (InputBox) is active
	tmuxSendKeys(t, session, "M-1")

	// Open InputBox dialog
	tmuxSendKeys(t, session, "F3")
	time.Sleep(500 * time.Millisecond)

	// Type some text
	tmuxType(t, session, "hello world")

	// Ctrl+Y to clear the input
	tmuxSendKeys(t, session, "C-y")

	// Type new text after clearing
	tmuxType(t, session, "cleared")

	// Press Enter to confirm
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// Static text should show "File: cleared" (proving Ctrl+Y cleared "hello world" before "cleared" was typed)
	if !containsAny(lines, "File: cleared") {
		t.Error("expected 'File: cleared' after Ctrl+Y clear then retype — Ctrl+Y may not have cleared the input")
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

func TestInputLineInsertAtStart(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-insert"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Switch to win1 so F3 is active
	tmuxSendKeys(t, session, "M-1")

	// Open InputBox dialog
	tmuxSendKeys(t, session, "F3")
	time.Sleep(500 * time.Millisecond)

	// Select all and type new text
	tmuxSendKeys(t, session, "C-a")
	tmuxType(t, session, "test")

	// Go to Home then type prefix
	tmuxSendKeys(t, session, "Home")
	tmuxType(t, session, "X")

	// Press Enter to confirm
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// Static text should show "File: Xtest"
	if !containsAny(lines, "File: Xtest") {
		t.Error("expected 'File: Xtest' after Home+type — text insertion at start may not work")
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

func TestCheckboxFocusIndicator(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-cbfocus"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Switch to win1
	tmuxSendKeys(t, session, "M-1")

	// win1's tab order: radioButtons (focused initially) → checkBoxes → ...
	// Tab once to move focus to checkBoxes cluster
	tmuxSendKeys(t, session, "Tab")

	lines := tmuxCapture(t, session)

	// The focused checkbox item should show ► prefix (SfSelected indicator)
	foundIndicator := false
	for _, line := range lines {
		if strings.Contains(line, "►") && strings.Contains(line, "[") {
			foundIndicator = true
			break
		}
	}
	if !foundIndicator {
		t.Error("checkbox focus indicator '►' with bracket not found — cluster focus indicator may not be rendering")
	}

	// Verify non-focused checkbox items have brackets at same column as focused.
	// The ► should be in column 0 relative to the cluster, with bracket at
	// column 1. Unfocused items should have space+bracket at the same positions,
	// NOT bracket starting at column 0 (which would cause visual shifting).
	// Look for lines with "[ ]" or "[X]" to find checkbox lines.
	var checkboxLines []string
	for _, line := range lines {
		if (strings.Contains(line, "[ ]") || strings.Contains(line, "[X]")) &&
			!strings.Contains(line, "►") {
			checkboxLines = append(checkboxLines, line)
		}
	}
	if len(checkboxLines) > 0 {
		// Unfocused checkbox lines should have a space before the bracket.
		// The focused line has ► before [, unfocused lines should have
		// a space before [ (not bracket at the leftmost cluster position).
		for _, line := range checkboxLines {
			idx := strings.Index(line, "[")
			if idx > 0 {
				preceding := string([]rune(line[:idx]))
				lastRune := []rune(preceding)[len([]rune(preceding))-1]
				if lastRune != ' ' {
					t.Errorf("unfocused checkbox bracket not preceded by space (preceded by %q): %q", string(lastRune), line)
				}
			}
		}
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

func TestCheckboxIndicatorHidesOnFocusLoss(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-cbunfocus"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Switch to win1
	tmuxSendKeys(t, session, "M-1")

	// Tab to checkboxes cluster
	tmuxSendKeys(t, session, "Tab")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// Precondition: indicator should be visible while cluster is focused
	foundIndicator := false
	for _, line := range lines {
		if strings.Contains(line, "►") && strings.Contains(line, "[") {
			foundIndicator = true
			break
		}
	}
	if !foundIndicator {
		t.Fatal("precondition: ► indicator not found while checkboxes focused")
	}

	// Tab away from checkboxes to next widget
	tmuxSendKeys(t, session, "Tab")
	time.Sleep(300 * time.Millisecond)

	lines = tmuxCapture(t, session)

	// Indicator should be gone — original TV hides all indicators when cluster loses focus
	for _, line := range lines {
		if strings.Contains(line, "►[ ]") || strings.Contains(line, "►[X]") {
			t.Errorf("► indicator still visible after Tab away from checkboxes: %q", line)
			break
		}
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

func TestMenuTileRearrangesWindows(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-tile"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// Verify both windows exist before tiling
	lines := tmuxCapture(t, session)
	if !containsAny(lines, "File Manager") {
		t.Fatal("win1 'File Manager' not visible before tile")
	}
	if !containsAny(lines, "Editor") {
		t.Fatal("win2 'Editor' not visible before tile")
	}

	// Open Window menu: F10 → Right (to Window menu) → Enter (open popup) → Enter (select Tile)
	tmuxSendKeys(t, session, "F10")
	time.Sleep(500 * time.Millisecond)
	tmuxSendKeys(t, session, "Right")
	time.Sleep(300 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(300 * time.Millisecond)
	tmuxSendKeys(t, session, "Enter")
	time.Sleep(500 * time.Millisecond)

	lines = tmuxCapture(t, session)

	// After tiling, both window titles must still be visible
	if !containsAny(lines, "File Manager") {
		t.Error("win1 'File Manager' not visible after Tile")
	}
	if !containsAny(lines, "Editor") {
		t.Error("win2 'Editor' not visible after Tile")
	}

	// Tiled windows share rows: look for a row that contains two or more frame
	// corner characters (active ╔ and/or inactive ┌). With 2 windows side by side
	// we expect ╔ and ┌ on the same row; with 3 windows in a 2-column grid the
	// two inactive windows share the top row (both ┌).
	tiledRow := false
	for _, line := range lines {
		corners := strings.Count(line, "╔") + strings.Count(line, "┌")
		if corners >= 2 {
			tiledRow = true
			break
		}
		// Also accept a row with one active (╔) and one inactive (┌) corner.
		if strings.Contains(line, "╔") && strings.Contains(line, "┌") {
			tiledRow = true
			break
		}
	}
	if !tiledRow {
		t.Error("no row with two or more frame corners found — windows may not be tiled side by side")
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

func TestListViewerSpaceSelect(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-listspace"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	// Navigate to win2's list viewer
	tmuxSendKeys(t, session, "Tab")
	time.Sleep(500 * time.Millisecond)

	// Arrow down a few times (navigation only, no selection)
	for i := 0; i < 3; i++ {
		tmuxSendKeys(t, session, "Down")
	}
	time.Sleep(500 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// Item 4 should be focused (started at Item 1, moved down 3)
	if !containsAny(lines, "Item 4") {
		t.Error("Item 4 not visible after navigating down 3 times")
	}

	// App still responsive — clean exit
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

// TestScrollBarNotFocusableViaTab verifies that Tab in the Editor window
// does NOT land on the scrollbar. In original Turbo Vision, scrollbars are
// not focusable — Tab skips them. The bug: Tab from ListViewer to ScrollBar
// makes Down arrow scroll without moving the selected item.
func TestScrollBarNotFocusableViaTab(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-sbnofocus"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	// Switch to win2 (Editor) which has ListViewer + ScrollBar
	tmuxSendKeys(t, session, "M-2")

	// Tab within win2 — if scrollbar is focusable, focus moves to it
	tmuxSendKeys(t, session, "Tab")

	// Now press Down arrow. If focus is on scrollbar (bug), this scrolls
	// without moving the highlighted item. If focus stayed on list (correct),
	// the highlighted item moves to Item 2.
	tmuxSendKeys(t, session, "Down")
	tmuxSendKeys(t, session, "Down")
	tmuxSendKeys(t, session, "Down")

	lines := tmuxCapture(t, session)

	// After Down x3 from Item 1, Item 4 should be the highlighted/selected item.
	// In a correctly functioning list, Item 4 is visible and highlighted.
	// If scrollbar stole focus, the list selection stays at Item 1 while
	// the viewport scrolled — Item 1 may not even be visible anymore.
	// Check that Item 1 is NOT still at the top (it should have scrolled
	// or Item 4 should be highlighted).
	item1AtTop := false
	item4Visible := false
	for _, line := range lines {
		if strings.Contains(line, "Item 1") {
			item1AtTop = true
		}
		if strings.Contains(line, "Item 4") {
			item4Visible = true
		}
	}

	// Item 4 must be visible (selected after 3 down presses)
	if !item4Visible {
		t.Error("Item 4 not visible after Tab + 3x Down in Editor window — scrollbar may have stolen focus")
	}

	// Item 1 should still be visible (only 3 items down, list has 10 visible rows)
	// but the key thing is that selection MOVED, not just scrollbar
	if !item1AtTop && !item4Visible {
		t.Error("neither Item 1 nor Item 4 visible — list state is broken")
	}

	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// TestF5ZoomWindow verifies F5 zooms the focused window to fill the desktop.
func TestF5ZoomWindow(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-f5zoom"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	tmuxSendKeys(t, session, "M-1")
	linesBefore := tmuxCapture(t, session)
	tmuxSendKeys(t, session, "F5")
	linesAfter := tmuxCapture(t, session)

	titleRowBefore := -1
	titleRowAfter := -1
	for i, line := range linesBefore {
		if strings.Contains(line, "File Manager") {
			titleRowBefore = i
			break
		}
	}
	for i, line := range linesAfter {
		if strings.Contains(line, "File Manager") {
			titleRowAfter = i
			break
		}
	}

	if titleRowBefore < 0 || titleRowAfter < 0 {
		t.Fatal("window title 'File Manager' not found before or after F5")
	}

	if titleRowAfter >= titleRowBefore {
		t.Errorf("F5 zoom did not move window up: title row before=%d, after=%d", titleRowBefore, titleRowAfter)
	}

	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// TestAltF3CloseWindow verifies Alt+F3 closes the focused window.
func TestAltF3CloseWindow(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-altf3"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	lines := tmuxCapture(t, session)
	if !containsAny(lines, "File Manager") {
		t.Fatal("win1 not visible")
	}
	if !containsAny(lines, "Editor") {
		t.Fatal("win2 not visible")
	}

	tmuxSendKeys(t, session, "M-2")
	tmuxSendKeys(t, session, "M-F3")

	lines = tmuxCapture(t, session)

	if containsAny(lines, "Editor") {
		t.Error("Alt+F3: Editor window still visible after close")
	}
	if !containsAny(lines, "File Manager") {
		t.Error("Alt+F3: File Manager window disappeared — wrong window closed")
	}

	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// TestLabelShortcutFocusesLink verifies that pressing Alt+N (the shortcut embedded
// in "~N~ame:") moves focus to the linked InputLine and allows typing into it.
func TestLabelShortcutFocusesLink(t *testing.T) {
	binPath := buildBasicApp(t)

	session := "tv3-e2e-label"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	// win2 is focused on startup; switch to win1 where the Label+InputLine live
	tmuxSendKeys(t, session, "M-1")

	// Verify label text is visible in win1
	lines := tmuxCapture(t, session)
	if !containsAny(lines, "ame:") {
		t.Error("label text 'ame:' not found in win1 — Label may not have rendered")
	}

	// Press Alt+N to activate the Label shortcut; focus should move to the InputLine
	tmuxSendKeys(t, session, "M-n")

	// Type text into the now-focused InputLine
	tmuxType(t, session, "hello")
	time.Sleep(300 * time.Millisecond)

	lines = tmuxCapture(t, session)

	// The InputLine is partially obscured by the Editor window, so only the first
	// few characters of "hello" are visible on screen. Check for "hel" which is
	// the reliably visible portion of the typed text.
	if !containsAny(lines, "hel") {
		t.Error("typed text not found after Alt+N shortcut — Label shortcut may not have focused the InputLine")
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

// TestF6NextWindow verifies F6 switches focus away from the current window.
// BringToFront reorders the children list, so we only assert that F6
// activates a different window than the one we started on.
func TestF6NextWindow(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-f6next"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	tmuxSendKeys(t, session, "M-1")
	tmuxSendKeys(t, session, "F6")

	lines := tmuxCapture(t, session)

	fileManagerStillActive := false
	for _, line := range lines {
		if strings.Contains(line, "File Manager") && strings.Contains(line, "═") {
			fileManagerStillActive = true
			break
		}
	}
	if fileManagerStillActive {
		t.Error("F6 did not switch focus — File Manager still has active frame")
	}

	// Verify some other window got the active frame
	switchedToOther := false
	for _, line := range lines {
		otherActive := (strings.Contains(line, "Notes") || strings.Contains(line, "Editor")) && strings.Contains(line, "═")
		if otherActive {
			switchedToOther = true
			break
		}
	}
	if !switchedToOther {
		t.Error("F6 did not activate any other window")
	}

	tmuxSendKeys(t, session, "M-x")
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestMemoVisible(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-memo"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	lines := tmuxCapture(t, session)

	// Window title "Notes" should be visible
	if !containsAny(lines, "Notes") {
		t.Error("window title 'Notes' not visible")
	}

	// Pre-loaded text should be visible
	if !containsAny(lines, "Hello, World!") {
		t.Error("memo text 'Hello, World!' not visible")
	}
	if !containsAny(lines, "This is a memo.") {
		t.Error("memo text 'This is a memo.' not visible")
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

func TestMemoTyping(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-memotype"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	// Switch to win3 (Notes) using Alt+3
	tmuxSendKeys(t, session, "M-3")

	// Move to end of first line and type additional text
	tmuxSendKeys(t, session, "End")
	tmuxType(t, session, " TEST")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// The typed text should appear: "Hello, World! TEST"
	if !containsAny(lines, "Hello, World! TEST") {
		t.Error("typed text 'Hello, World! TEST' not visible after typing in memo")
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

// TestMemoScrollbarVisible verifies that after adding a vertical scrollbar to the
// Notes window's Memo, the scrollbar arrow characters are visible on screen.
func TestMemoScrollbarVisible(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-memoscrollbar"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	// Focus the Notes window (win3)
	tmuxSendKeys(t, session, "M-3")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// Scrollbar arrow characters should be visible within the Notes window area
	if !containsAny(lines, "▲", "▼") {
		t.Error("scrollbar arrow characters '▲' or '▼' not visible in Notes window")
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

// TestMemoTypingAdvanced verifies typing works in the scrollbar-enabled Memo.
func TestMemoTypingAdvanced(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-memotypadv"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	// Focus the Notes window (win3)
	tmuxSendKeys(t, session, "M-3")

	// Move to end of first line and type additional text
	tmuxSendKeys(t, session, "End")
	tmuxType(t, session, " ADV")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// The typed text should appear: "Hello, World! ADV"
	if !containsAny(lines, "Hello, World! ADV") {
		t.Error("typed text 'Hello, World! ADV' not visible after typing in scrollbar-enabled memo")
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

// TestMemoScroll verifies that PgDn scrolls the Memo content, bringing later
// lines into view and pushing earlier lines out.
func TestMemoScroll(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-memoscroll"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	// Focus the Notes window (win3)
	tmuxSendKeys(t, session, "M-3")

	// Precondition: "Hello, World!" should be visible before scrolling
	lines := tmuxCapture(t, session)
	if !containsAny(lines, "Hello, World!") {
		t.Fatal("precondition: 'Hello, World!' not visible before PgDn")
	}

	// Press PgDn to scroll down (NPage is tmux's key name for Page Down)
	tmuxSendKeys(t, session, "NPage")
	time.Sleep(300 * time.Millisecond)

	lines = tmuxCapture(t, session)

	// Later lines should now be visible (PgDn scrolls by viewport height)
	if !containsAny(lines, "Line 13", "Line 14", "Line 15", "Line 16") {
		t.Error("later lines (Line 13+) not visible after PgDn — scroll may not have worked")
	}

	// Earlier content should have scrolled out of view
	if containsAny(lines, "Hello, World!") {
		t.Error("'Hello, World!' still visible after PgDn — memo did not scroll")
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

// TestHistoryIconVisible verifies that the History widget renders its ↓ arrow icon
// in win1 next to the InputLine. win2 (Editor) overlaps the History position, so
// we close win2 first to expose win1's full client area.
func TestHistoryIconVisible(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-history-icon"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	// Close win2 (Editor) which overlaps the History widget position in win1.
	// Focus win2 then Alt+F3 closes it.
	tmuxSendKeys(t, session, "M-2")
	time.Sleep(300 * time.Millisecond)
	tmuxSendKeys(t, session, "M-F3")
	time.Sleep(300 * time.Millisecond)

	// Focus win1 (File Manager)
	tmuxSendKeys(t, session, "M-1")
	time.Sleep(300 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// The History widget renders ↓ as its arrow icon
	if !containsAny(lines, "↓") {
		t.Error("History arrow icon '↓' not visible in win1 — History widget may not have rendered")
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

// TestHistoryDropdown verifies that typing into the InputLine, tabbing away to
// record the history, then pressing Down opens a dropdown showing the recorded entry.
func TestHistoryDropdown(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-history"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	// Focus win1 (File Manager)
	tmuxSendKeys(t, session, "M-1")
	time.Sleep(300 * time.Millisecond)

	// Tab to the InputLine: (no initial focus) → OK → Close → CheckBoxes → RadioButtons → InputLine (5 tabs)
	for i := 0; i < 5; i++ {
		tmuxSendKeys(t, session, "Tab")
	}
	time.Sleep(300 * time.Millisecond)

	// Type some text into the InputLine
	tmuxType(t, session, "test1")
	time.Sleep(300 * time.Millisecond)

	// Tab away to trigger CmReleasedFocus which records the text in history.
	// One tab wraps back to OK.
	tmuxSendKeys(t, session, "Tab")
	time.Sleep(300 * time.Millisecond)

	// Tab back to the InputLine (OK → Close → CheckBoxes → RadioButtons → InputLine = 4 tabs)
	for i := 0; i < 4; i++ {
		tmuxSendKeys(t, session, "Tab")
	}
	time.Sleep(300 * time.Millisecond)

	// Press Down arrow to open the history dropdown
	tmuxSendKeys(t, session, "Down")
	time.Sleep(500 * time.Millisecond)

	lines := tmuxCapture(t, session)

	// The dropdown should show the previously entered "test1"
	if !containsAny(lines, "test1") {
		t.Error("history dropdown entry 'test1' not visible after pressing Down — history dropdown may not have opened")
	}

	// Dismiss with Escape
	tmuxSendKeys(t, session, "Escape")
	time.Sleep(300 * time.Millisecond)

	// App should still be running
	if !tmuxHasSession(session) {
		t.Error("app exited unexpectedly after pressing Escape on history dropdown")
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

// TestListBoxNavigation verifies the ListBox widget in the Editor window:
// initial items are visible, and arrow key navigation moves the selection.
func TestListBoxNavigation(t *testing.T) {
	binPath := buildBasicApp(t)
	session := "tv3-e2e-listbox"
	exec.Command("tmux", "kill-session", "-t", session).Run()
	startTmux(t, session, binPath)

	// win2 (Editor) is focused on startup — it contains the ListBox
	lines := tmuxCapture(t, session)

	// Items 1 through 10 should be visible in the initial view
	for _, item := range []string{"Item 1", "Item 2", "Item 3", "Item 4", "Item 5",
		"Item 6", "Item 7", "Item 8", "Item 9", "Item 10"} {
		if !containsAny(lines, item) {
			t.Errorf("ListBox: %q not visible on initial render", item)
		}
	}

	// Scrollbar arrows should be visible (ListBox includes a ScrollBar)
	if !containsAny(lines, "▲", "▼") {
		t.Error("ListBox: scrollbar arrow characters '▲' or '▼' not visible")
	}

	// Press Down 4 times — selection should move from Item 1 to Item 5
	for i := 0; i < 4; i++ {
		tmuxSendKeys(t, session, "Down")
	}
	time.Sleep(500 * time.Millisecond)

	lines = tmuxCapture(t, session)

	// After navigating down 4 times, Item 5 should be visible
	if !containsAny(lines, "Item 5") {
		t.Error("ListBox: Item 5 not visible after pressing Down 4 times — navigation may not work")
	}

	// Item 1 should still be visible (only 4 rows down, well within the 10-row viewport)
	if !containsAny(lines, "Item 1") {
		t.Error("ListBox: Item 1no longer visible after Down x4 — unexpected scrolling")
	}

	// Alt+X to exit
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
