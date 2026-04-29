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

func TestBasicAppBoot(t *testing.T) {
	root := projectRoot()
	binPath := filepath.Join(root, "e2e", "testapp", "basic", "basic")

	out, err := exec.Command("go", "build", "-o", binPath, filepath.Join(root, "e2e", "testapp", "basic")).CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	t.Cleanup(func() { exec.Command("rm", binPath).Run() })

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

	// Window frame characters visible (double-line border for active window)
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
