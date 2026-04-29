package e2e

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

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
