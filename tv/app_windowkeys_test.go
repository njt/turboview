package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// app_windowkeys_test.go — Tests for standard Turbo Vision window keyboard
// shortcuts handled at the Application level: F5→Zoom, Ctrl+F5→Resize,
// Alt+F3→Close, F6→Next, Shift+F6→Prev.

func newWindowKeysApp(t *testing.T) (*Application, *Window, *Window) {
	t.Helper()
	sl := NewStatusLine(NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit))
	app, err := NewApplication(
		WithStatusLine(sl),
		WithTheme(theme.BorlandBlue),
		WithScreen(newTestScreen(t)),
	)
	if err != nil {
		t.Fatal(err)
	}

	w1 := NewWindow(NewRect(5, 2, 30, 10), "Win1", WithWindowNumber(1))
	w2 := NewWindow(NewRect(10, 4, 30, 10), "Win2", WithWindowNumber(2))
	app.Desktop().Insert(w1)
	app.Desktop().Insert(w2)
	return app, w1, w2
}

// --- F5 → CmZoom ---

// TestAppF5ZoomsWindow verifies F5 generates CmZoom, which toggles the
// focused window's zoom state.
func TestAppF5ZoomsWindow(t *testing.T) {
	app, _, w2 := newWindowKeysApp(t)


	origBounds := w2.Bounds()

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF5}}
	app.handleEvent(ev)

	if w2.Bounds() == origBounds {
		t.Error("F5 did not change window bounds — CmZoom not triggered")
	}
}

// TestAppF5EventCleared verifies F5 clears the event.
func TestAppF5EventCleared(t *testing.T) {
	app, _, _ := newWindowKeysApp(t)


	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF5}}
	app.handleEvent(ev)

	if !ev.IsCleared() {
		t.Error("F5 event not cleared after handling")
	}
}

// --- Alt+F3 → CmClose ---

// TestAppAltF3ClosesWindow verifies Alt+F3 closes the focused window.
func TestAppAltF3ClosesWindow(t *testing.T) {
	app, _, w2 := newWindowKeysApp(t)


	childCount := len(app.Desktop().Children())

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF3, Modifiers: tcell.ModAlt}}
	app.handleEvent(ev)

	if len(app.Desktop().Children()) != childCount-1 {
		t.Error("Alt+F3 did not close the focused window")
	}
	// w2 should no longer be in the children
	for _, child := range app.Desktop().Children() {
		if child == w2 {
			t.Error("Alt+F3: closed window w2 still in Desktop children")
		}
	}
}

// TestAppAltF3EventCleared verifies Alt+F3 clears the event.
func TestAppAltF3EventCleared(t *testing.T) {
	app, _, _ := newWindowKeysApp(t)


	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF3, Modifiers: tcell.ModAlt}}
	app.handleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Alt+F3 event not cleared")
	}
}

// --- F6 → CmNext ---

// TestAppF6SelectsNextWindow verifies F6 cycles to the next window.
func TestAppF6SelectsNextWindow(t *testing.T) {
	app, w1, w2 := newWindowKeysApp(t)


	// w2 is focused (last inserted)
	if app.Desktop().FocusedChild() != w2 {
		t.Fatal("expected w2 focused initially")
	}

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF6}}
	app.handleEvent(ev)

	if app.Desktop().FocusedChild() != w1 {
		t.Error("F6 did not switch focus to next window")
	}
}

// TestAppF6EventCleared verifies F6 clears the event.
func TestAppF6EventCleared(t *testing.T) {
	app, _, _ := newWindowKeysApp(t)


	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF6}}
	app.handleEvent(ev)

	if !ev.IsCleared() {
		t.Error("F6 event not cleared")
	}
}

// --- Shift+F6 → CmPrev ---

// TestAppShiftF6SelectsPrevWindow verifies Shift+F6 cycles to the previous window.
func TestAppShiftF6SelectsPrevWindow(t *testing.T) {
	app, w1, w2 := newWindowKeysApp(t)


	// Move to w1 first
	app.Desktop().SetFocusedChild(w1)
	if app.Desktop().FocusedChild() != w1 {
		t.Fatal("expected w1 focused")
	}

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF6, Modifiers: tcell.ModShift}}
	app.handleEvent(ev)

	if app.Desktop().FocusedChild() != w2 {
		t.Error("Shift+F6 did not switch focus to previous window")
	}
}

// --- Ctrl+F5 → CmResize ---

// TestAppCtrlF5SendsResize verifies Ctrl+F5 generates CmResize command.
func TestAppCtrlF5SendsResize(t *testing.T) {
	app, _, _ := newWindowKeysApp(t)


	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF5, Modifiers: tcell.ModCtrl}}
	app.handleEvent(ev)

	// CmResize is currently a no-op in Window.HandleEvent (just clears the event).
	// The test verifies the key binding exists and the event is consumed.
	if !ev.IsCleared() {
		t.Error("Ctrl+F5 event not cleared — CmResize binding missing")
	}
}

// --- Edge cases ---

// TestAppF5NoopWithoutWindow verifies F5 doesn't panic when no window is focused.
func TestAppF5NoopWithoutWindow(t *testing.T) {
	sl := NewStatusLine(NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit))
	app, err := NewApplication(
		WithStatusLine(sl),
		WithTheme(theme.BorlandBlue),
		WithScreen(newTestScreen(t)),
	)
	if err != nil {
		t.Fatal(err)
	}


	// No windows inserted
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF5}}
	// Should not panic
	app.handleEvent(ev)
}
