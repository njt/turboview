package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// --- Task 13: Button Default Protocol ---

// TestPhase9ButtonWithDefaultIsDefaultTrue verifies that a button created with
// WithDefault() has IsDefault() == true (amDefault starts true).
// Spec: "amDefault starts true for WithDefault button"
func TestPhase9ButtonWithDefaultIsDefaultTrue(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK, WithDefault())

	if !b.IsDefault() {
		t.Error("WithDefault button: IsDefault() should be true initially")
	}
}

// TestPhase9ButtonWithoutDefaultIsDefaultFalse verifies that a button created
// without WithDefault() has IsDefault() == false (amDefault starts false).
// Spec: "amDefault starts false for non-default button"
func TestPhase9ButtonWithoutDefaultIsDefaultFalse(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)

	if b.IsDefault() {
		t.Error("non-default button: IsDefault() should be false initially")
	}
}

// TestPhase9NonDefaultGainsFocusBecomesDefault verifies that when a non-default button
// gains focus (SfSelected), it broadcasts CmGrabDefault and sets amDefault=true.
// The sibling WithDefault button should lose amDefault (receives CmGrabDefault).
// Spec: "Non-default button gains focus → broadcasts CmGrabDefault, becomes amDefault"
func TestPhase9NonDefaultGainsFocusBecomesDefault(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	defBtn := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK, WithDefault())
	plainBtn := NewButton(NewRect(14, 0, 14, 1), "Cancel", CmCancel)

	g.Insert(defBtn)
	g.Insert(plainBtn) // inserted last → gets focus

	// After insertion, plainBtn is focused — it broadcasts CmGrabDefault.
	// defBtn (bfDefault=true) should respond by setting amDefault=false.
	if defBtn.IsDefault() {
		t.Error("defBtn should have amDefault=false after non-default sibling gained focus")
	}
	if !plainBtn.IsDefault() {
		t.Error("plainBtn should have amDefault=true after gaining focus")
	}
}

// TestPhase9NonDefaultLosesFocusRestoresDefault verifies that when a non-default
// button loses focus, it broadcasts CmReleaseDefault and sets amDefault=false.
// The sibling WithDefault button should regain amDefault.
// Spec: "Non-default button loses focus → broadcasts CmReleaseDefault, amDefault false"
func TestPhase9NonDefaultLosesFocusRestoresDefault(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	defBtn := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK, WithDefault())
	plainBtn := NewButton(NewRect(14, 0, 14, 1), "Cancel", CmCancel)

	g.Insert(defBtn)
	g.Insert(plainBtn) // plainBtn focused → defBtn.amDefault = false

	// Now move focus back to defBtn.
	g.SetFocusedChild(defBtn)

	// plainBtn should have lost amDefault; defBtn should have regained amDefault.
	if plainBtn.IsDefault() {
		t.Error("plainBtn should have amDefault=false after losing focus")
	}
	if !defBtn.IsDefault() {
		t.Error("defBtn should have amDefault=true after non-default sibling lost focus")
	}
}

// TestPhase9DefaultButtonRespondsToCmDefault verifies that a WithDefault button
// with amDefault=true fires its command when it receives a CmDefault broadcast.
// Spec: "Default button responds to CmDefault broadcast"
func TestPhase9DefaultButtonRespondsToCmDefault(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK, WithDefault())

	ev := &Event{What: EvBroadcast, Command: CmDefault}
	b.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("CmDefault broadcast: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("CmDefault broadcast: ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
}

// TestPhase9FocusedNonDefaultRespondsToCmDefault verifies that a non-default button
// that has gained focus (amDefault=true) also responds to CmDefault broadcast.
// Spec: "Non-default focused button responds to CmDefault"
func TestPhase9FocusedNonDefaultRespondsToCmDefault(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	defBtn := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK, WithDefault())
	plainBtn := NewButton(NewRect(14, 0, 14, 1), "Cancel", CmCancel)

	g.Insert(defBtn)
	g.Insert(plainBtn) // plainBtn focused → plainBtn.amDefault = true, defBtn.amDefault = false

	// plainBtn is now the active default (amDefault=true).
	ev := &Event{What: EvBroadcast, Command: CmDefault}
	plainBtn.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("focused non-default: CmDefault broadcast should fire; ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmCancel {
		t.Errorf("focused non-default: ev.Command = %v, want CmCancel (%v)", ev.Command, CmCancel)
	}
}

// --- Task 14: Button Alt+Shortcut ---

// TestPhase9AltShortcutFiresButton verifies that pressing Alt+shortcut fires the button.
// Title "~Save" means 'S' is the shortcut.
// Spec: "Alt+shortcut fires button"
func TestPhase9AltShortcutFiresButton(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "~Save", CmOK)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'S', Modifiers: tcell.ModAlt},
	}
	b.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("Alt+S: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("Alt+S: ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
}

// TestPhase9AltShortcutCaseInsensitive verifies Alt+shortcut works regardless of case.
// Spec: "Alt+shortcut case-insensitive: Alt+s and Alt+S both work"
func TestPhase9AltShortcutLowercaseFires(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "~Save", CmOK)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: tcell.ModAlt},
	}
	b.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("Alt+s (lowercase): ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
}

// TestPhase9AltShortcutUppercaseFires verifies Alt+uppercase shortcut also fires.
// Spec: "Alt+shortcut case-insensitive"
func TestPhase9AltShortcutUppercaseFires(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "~save", CmOK) // shortcut is lowercase 's'

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'S', Modifiers: tcell.ModAlt},
	}
	b.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("Alt+S with lowercase shortcut: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
}

// --- Task 15: Button Space Guard and Enter Removal ---

// TestPhase9SpaceFiresWhenFocused verifies Space fires the button when it has focus (SfSelected).
// Spec: "Space only fires when SfSelected"
func TestPhase9SpaceFiresWhenFocused(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)
	b.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	b.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("Space when focused: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("Space when focused: ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
}

// TestPhase9SpaceDoesNotFireWhenUnfocused verifies Space does NOT fire the button when
// it does not have focus.
// Spec: "unfocused → Space does nothing"
func TestPhase9SpaceDoesNotFireWhenUnfocused(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)
	// SfSelected is not set

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	b.HandleEvent(ev)

	if ev.What == EvCommand {
		t.Error("Space when unfocused: should not fire command")
	}
}

// TestPhase9EnterDoesNotFireButton verifies that Enter key does NOT fire the button.
// Spec: "Enter handler completely removed"
func TestPhase9EnterDoesNotFireButton(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)
	b.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	b.HandleEvent(ev)

	if ev.What == EvCommand {
		t.Error("Enter key should NOT fire button in phase9 — Enter handler must be removed")
	}
}

// TestPhase9MouseClickFiresButton verifies a left-click fires the button.
// Spec: "Mouse click fires button"
func TestPhase9MouseClickFiresButton(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1},
	}
	b.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("mouse click: ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("mouse click: ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
}

// TestPhase9AllButtonsHaveOfPostProcess verifies that ALL buttons (not just WithDefault)
// have OfPostProcess set, for broadcasts and Alt+shortcuts.
// Spec: "All buttons set OfPostProcess (not just default buttons)"
func TestPhase9AllButtonsHaveOfPostProcess(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)

	if !b.HasOption(OfPostProcess) {
		t.Error("non-default button should have OfPostProcess set in phase9")
	}
}
