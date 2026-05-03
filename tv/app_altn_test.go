package tv

// app_altn_test.go — Tests for Task 5 (Phase 14): Moving Alt+N window switching
// from Desktop.HandleEvent to Application.handleEvent.
//
// Spec summary (Task 5, section 13.2):
//   - Application.handleEvent handles Alt+1-9 keyboard events.
//   - It creates an EvBroadcast/CmSelectWindowNum event with the number as Info.
//   - It sends the broadcast through app.desktop.HandleEvent (Group dispatch).
//   - If the broadcast is consumed (a window was found and focused), the
//     original keyboard event is cleared.
//   - Desktop.HandleEvent NO LONGER handles Alt+N keyboard events.
//   - Desktop.selectWindowByNumber may still exist but is no longer called for
//     this code path.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// appAltNKeyEvent creates an EvKeyboard event for Alt+<digit rune> suitable for
// dispatching through Application.handleEvent.
func appAltNKeyEvent(digit rune) *Event {
	return &Event{
		What: EvKeyboard,
		Key: &KeyEvent{
			Key:       tcell.KeyRune,
			Rune:      digit,
			Modifiers: tcell.ModAlt,
		},
	}
}

// newAppWithWindows creates an Application (with a simulation screen) and
// inserts windows into its desktop. Each element of nums is assigned as the
// window number; windows are given 20-wide, 10-tall non-overlapping bounds
// starting from x=0. Returns the app and the slice of created windows.
func newAppWithWindows(t *testing.T, nums ...int) (*Application, []*Window) {
	t.Helper()
	screen := newTestScreen(t)
	t.Cleanup(func() { screen.Fini() })

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	windows := make([]*Window, len(nums))
	for i, n := range nums {
		bounds := NewRect(i*20, 0, 20, 10)
		w := NewWindow(bounds, "W", WithWindowNumber(n))
		app.Desktop().Insert(w)
		windows[i] = w
	}
	return app, windows
}

// ---------------------------------------------------------------------------
// Requirement 1: Alt+1 through Application.handleEvent focuses the window
// with number 1.
// ---------------------------------------------------------------------------

// TestAppAltN_Alt1FocusesWindowNumber1 verifies that dispatching an Alt+1
// keyboard event through Application.handleEvent causes the window numbered 1
// to become the focused child of the desktop.
//
// Spec: "Application.handleEvent handles Alt+1-9 keyboard events; broadcasts
// CmSelectWindowNum; windows respond by focusing if their number matches."
func TestAppAltN_Alt1FocusesWindowNumber1(t *testing.T) {
	app, windows := newAppWithWindows(t, 1, 2)
	w1, w2 := windows[0], windows[1]

	// Make w2 the active window so we can verify a real focus change.
	app.Desktop().BringToFront(w2)
	if app.Desktop().FocusedChild() != w2 {
		t.Fatalf("precondition: FocusedChild() = %v, want w2", app.Desktop().FocusedChild())
	}

	ev := appAltNKeyEvent('1')
	app.handleEvent(ev)

	if app.Desktop().FocusedChild() != w1 {
		t.Errorf("Alt+1 via Application: FocusedChild() = %v, want w1 (number 1)", app.Desktop().FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: Alt+N event is cleared when a window is found.
// ---------------------------------------------------------------------------

// TestAppAltN_EventClearedWhenWindowFound verifies that the original keyboard
// event is cleared (marked consumed) after a matching window is focused.
//
// Spec: "If the broadcast is consumed (window found), the original keyboard
// event is cleared."
func TestAppAltN_EventClearedWhenWindowFound(t *testing.T) {
	app, _ := newAppWithWindows(t, 3)

	ev := appAltNKeyEvent('3')
	app.handleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Alt+3 with matching window: event not cleared (IsCleared = false), want true")
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: Alt+N with no matching window — event is NOT cleared.
// ---------------------------------------------------------------------------

// TestAppAltN_EventNotClearedWhenNoMatchingWindow verifies that when no
// window exists with the pressed number, the event is left unconsumed.
//
// Spec: "If the broadcast is not consumed (no window found), the event is not
// cleared."
func TestAppAltN_EventNotClearedWhenNoMatchingWindow(t *testing.T) {
	app, _ := newAppWithWindows(t, 1)

	// Alt+9 — no window has number 9.
	ev := appAltNKeyEvent('9')
	app.handleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("Alt+9 with no matching window: event was cleared, want it left unconsumed")
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Alt+0 is not handled (only 1-9).
// ---------------------------------------------------------------------------

// TestAppAltN_Alt0IsNotHandled verifies that Alt+0 does not trigger window
// number selection. Only digits 1-9 are in range.
//
// Spec: "Alt+1-9 only; Alt+0 is not handled."
func TestAppAltN_Alt0IsNotHandled(t *testing.T) {
	// A window with number 0 exists; Alt+0 must not focus it.
	screen := newTestScreen(t)
	defer screen.Fini()
	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	w0 := NewWindow(NewRect(0, 0, 20, 10), "W0", WithWindowNumber(0))
	anchor := NewWindow(NewRect(20, 0, 20, 10), "Anchor", WithWindowNumber(1))
	app.Desktop().Insert(w0)
	app.Desktop().Insert(anchor)
	app.Desktop().BringToFront(anchor)

	ev := appAltNKeyEvent('0')
	app.handleEvent(ev)

	// Focus must remain on anchor — w0 must not be selected.
	if app.Desktop().FocusedChild() == w0 {
		t.Errorf("Alt+0: FocusedChild() switched to w0 (number 0), want no change (0 is out of range)")
	}
	// Event must not be cleared by the Application's Alt+N handler.
	if ev.IsCleared() {
		t.Errorf("Alt+0: event was cleared — Application must not handle Alt+0 as a window selector")
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: Desktop.HandleEvent NO LONGER handles Alt+N directly.
// ---------------------------------------------------------------------------

// TestAppAltN_DesktopDoesNotHandleAltNDirectly verifies that sending an Alt+N
// keyboard event directly to Desktop.HandleEvent does NOT consume the event
// (i.e., does not clear it). The window-number dispatch has been moved to
// Application.
//
// Spec: "Desktop.HandleEvent NO LONGER handles Alt+N keyboard events."
func TestAppAltN_DesktopDoesNotHandleAltNDirectly(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1", WithWindowNumber(1))
	d.Insert(w1)

	// Send Alt+1 directly to the Desktop (bypassing Application).
	ev := &Event{
		What: EvKeyboard,
		Key: &KeyEvent{
			Key:       tcell.KeyRune,
			Rune:      '1',
			Modifiers: tcell.ModAlt,
		},
	}
	d.HandleEvent(ev)

	// After Task 5 is implemented, the Desktop must NOT consume this event.
	if ev.IsCleared() {
		t.Errorf("Desktop.HandleEvent consumed Alt+1 — it must no longer handle Alt+N keyboard events; that responsibility has moved to Application")
	}
}

// TestAppAltN_DesktopDoesNotChangeFocusOnAltN verifies that delivering Alt+N
// directly to the Desktop does not change window focus. The Desktop no longer
// drives window-number selection.
func TestAppAltN_DesktopDoesNotChangeFocusOnAltN(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1", WithWindowNumber(1))
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2", WithWindowNumber(2))
	d.Insert(w1)
	d.Insert(w2)
	d.BringToFront(w2) // w2 is focused

	ev := &Event{
		What: EvKeyboard,
		Key: &KeyEvent{
			Key:       tcell.KeyRune,
			Rune:      '1',
			Modifiers: tcell.ModAlt,
		},
	}
	d.HandleEvent(ev)

	// Focus must remain on w2 — the Desktop no longer handles Alt+N.
	if d.FocusedChild() != w2 {
		t.Errorf("Desktop.HandleEvent(Alt+1): FocusedChild() = %v, want w2 (Desktop must not change focus on Alt+N)", d.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: Falsification — Alt+2 on a 3-window desktop actually
// switches focus to window 2, not just clearing the event.
// ---------------------------------------------------------------------------

// TestAppAltN_FalsificationAlt2SwitchesToWindow2 verifies that Alt+2 on a
// 3-window desktop results in window 2 being the focused child, not merely
// clearing the event without a real focus change.
//
// This test guards against a naive implementation that clears the event but
// does not actually route the broadcast to the windows.
func TestAppAltN_FalsificationAlt2SwitchesToWindow2(t *testing.T) {
	app, windows := newAppWithWindows(t, 1, 2, 3)
	w1, w2, w3 := windows[0], windows[1], windows[2]

	// Start with w1 focused.
	app.Desktop().BringToFront(w1)
	if app.Desktop().FocusedChild() != w1 {
		t.Fatalf("precondition: FocusedChild() = %v, want w1", app.Desktop().FocusedChild())
	}

	app.handleEvent(appAltNKeyEvent('2'))

	focused := app.Desktop().FocusedChild()
	if focused != w2 {
		t.Errorf("Alt+2: FocusedChild() = %v, want w2 (number 2)", focused)
	}
	if focused == w1 {
		t.Errorf("Alt+2: FocusedChild() is still w1 — focus did not change to w2")
	}
	if focused == w3 {
		t.Errorf("Alt+2: FocusedChild() = w3 — wrong window focused (want w2)")
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: Application uses CmSelectWindowNum broadcast mechanism.
//
// The spec says to verify indirectly: a non-selectable window must NOT be
// focused by Alt+N (because CmSelectWindowNum checks OfSelectable), whereas
// the old Desktop.selectWindowByNumber used BringToFront directly and did NOT
// check OfSelectable. If Alt+N skips a non-selectable window, the broadcast
// mechanism must be in use.
// ---------------------------------------------------------------------------

// TestAppAltN_BroadcastMechanismVerified_NonSelectableNotFocused verifies that
// pressing Alt+N for a non-selectable window's number does NOT focus that window.
// This proves the Application uses the CmSelectWindowNum broadcast (which checks
// OfSelectable in Window.HandleEvent) rather than the old
// Desktop.selectWindowByNumber (which called BringToFront directly without
// checking OfSelectable).
//
// Spec: "Application broadcasts CmSelectWindowNum to the desktop, and windows
// respond via the CmSelectWindowNum handler (already implemented in Task 2)."
// Task 2 spec: "If HasOption(OfSelectable) is false, ignore."
func TestAppAltN_BroadcastMechanismVerified_NonSelectableNotFocused(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()
	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	anchor := NewWindow(NewRect(0, 0, 20, 10), "Anchor", WithWindowNumber(1))
	target := NewWindow(NewRect(20, 0, 20, 10), "Target", WithWindowNumber(2))
	app.Desktop().Insert(anchor)
	app.Desktop().Insert(target)
	app.Desktop().BringToFront(anchor)

	if app.Desktop().FocusedChild() != anchor {
		t.Fatalf("precondition: FocusedChild() = %v, want anchor", app.Desktop().FocusedChild())
	}

	// Disable selectability on target. The old selectWindowByNumber ignored
	// this flag; the broadcast mechanism honours it.
	target.SetOptions(OfSelectable, false)

	ev := appAltNKeyEvent('2')
	app.handleEvent(ev)

	// The event must NOT be cleared (no window accepted the broadcast).
	if ev.IsCleared() {
		t.Errorf("Alt+2 for non-selectable window: event was cleared — either the old BringToFront path is still in use, or Window.HandleEvent does not check OfSelectable")
	}
	// Focus must remain on anchor.
	if app.Desktop().FocusedChild() != anchor {
		t.Errorf("Alt+2 for non-selectable target: FocusedChild() = %v, want anchor (non-selectable must not be focused)", app.Desktop().FocusedChild())
	}
}

// TestAppAltN_BroadcastMechanismVerified_SelectableWindowIsFocused is the
// affirmative counterpart: a selectable window with the matching number IS
// focused, proving the test above is not vacuous.
func TestAppAltN_BroadcastMechanismVerified_SelectableWindowIsFocused(t *testing.T) {
	app, windows := newAppWithWindows(t, 1, 2)
	w1, w2 := windows[0], windows[1]
	app.Desktop().BringToFront(w1)

	// w2 has OfSelectable enabled (NewWindow default).
	app.handleEvent(appAltNKeyEvent('2'))

	if app.Desktop().FocusedChild() != w2 {
		t.Errorf("Alt+2 for selectable window: FocusedChild() = %v, want w2 (number 2)", app.Desktop().FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Additional edge cases
// ---------------------------------------------------------------------------

// TestAppAltN_AllDigits1Through9AreHandled verifies that each digit 1-9
// causes Application to attempt window-number selection (clears event when a
// matching window exists).
func TestAppAltN_AllDigits1Through9AreHandled(t *testing.T) {
	for digit := rune('1'); digit <= '9'; digit++ {
		digit := digit
		n := int(digit - '0')
		t.Run(string(digit), func(t *testing.T) {
			app, _ := newAppWithWindows(t, n)

			ev := appAltNKeyEvent(digit)
			app.handleEvent(ev)

			if !ev.IsCleared() {
				t.Errorf("Alt+%c with matching window number %d: event not cleared", digit, n)
			}
			if app.Desktop().FocusedChild() == nil {
				t.Errorf("Alt+%c: FocusedChild() is nil after selection", digit)
			}
		})
	}
}

// TestAppAltN_EmptyDesktopDoesNotPanic verifies that pressing Alt+N when the
// desktop has no windows does not panic and does not clear the event.
func TestAppAltN_EmptyDesktopDoesNotPanic(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()
	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	ev := appAltNKeyEvent('5')
	// Must not panic.
	app.handleEvent(ev)

	// No matching window — event must not be cleared.
	if ev.IsCleared() {
		t.Errorf("Alt+5 on empty desktop: event was cleared, want it left unconsumed")
	}
}
