package e2e

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

func startTmux(t *testing.T, session string, cmd string) {
	t.Helper()
	err := exec.Command("tmux", "new-session", "-d", "-s", session, "-x", "80", "-y", "25", cmd).Run()
	if err != nil {
		t.Fatalf("failed to start tmux session: %v", err)
	}
	t.Cleanup(func() {
		exec.Command("tmux", "kill-session", "-t", session).Run()
	})
	time.Sleep(1 * time.Second)
}

func tmuxSendKeys(t *testing.T, session string, keys ...string) {
	t.Helper()
	args := append([]string{"send-keys", "-t", session}, keys...)
	if err := exec.Command("tmux", args...).Run(); err != nil {
		t.Fatalf("failed to send keys: %v", err)
	}
	time.Sleep(300 * time.Millisecond)
}

// tmuxType sends a literal string to the tmux session without key-name interpretation.
func tmuxType(t *testing.T, session string, text string) {
	t.Helper()
	if err := exec.Command("tmux", "send-keys", "-t", session, "-l", text).Run(); err != nil {
		t.Fatalf("failed to type text: %v", err)
	}
	time.Sleep(300 * time.Millisecond)
}

func tmuxCapture(t *testing.T, session string) []string {
	t.Helper()
	out, err := exec.Command("tmux", "capture-pane", "-t", session, "-p").Output()
	if err != nil {
		t.Fatalf("failed to capture pane: %v", err)
	}
	return strings.Split(string(out), "\n")
}

func tmuxHasSession(session string) bool {
	return exec.Command("tmux", "has-session", "-t", session).Run() == nil
}
