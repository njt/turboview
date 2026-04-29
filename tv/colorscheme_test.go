package tv

// colorscheme_test.go — tests for Task 2: SetColorScheme on Window and Dialog.
//
// Every assertion cites the spec sentence it verifies.
// Each test covers exactly one behaviour.
//
// Spec requirements tested:
//   1. Window.SetColorScheme sets scheme; ColorScheme() returns it.
//   2. Window.SetColorScheme(nil) clears scheme; ColorScheme() falls back to owner chain.
//   3. Dialog.SetColorScheme sets scheme; ColorScheme() returns it.
//   4. Dialog.SetColorScheme(nil) clears scheme; ColorScheme() falls back to owner chain.
//   5. Children of a window with a custom scheme inherit that scheme via owner-chain walk.
//   6. (falsifying) Setting scheme on one window does not affect a sibling window's scheme.
//   7. (falsifying) A child of a scheme-bearing window gets the window's scheme, not nil.

import (
	"testing"

	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Requirement 1 — Window.SetColorScheme stores the scheme
// Spec: "SetColorScheme(cs *theme.ColorScheme) sets the BaseView.scheme field on the Window"
// Spec: "After calling SetColorScheme, window.ColorScheme() returns the set scheme"
// ---------------------------------------------------------------------------

// TestWindowSetColorSchemeColorSchemeReturnsIt verifies that after calling
// SetColorScheme with a non-nil scheme, ColorScheme() returns that exact scheme.
func TestWindowSetColorSchemeColorSchemeReturnsIt(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	cs := &theme.ColorScheme{}
	*cs = *theme.BorlandBlue

	w.SetColorScheme(cs)
	got := w.ColorScheme()

	if got != cs {
		t.Errorf("after SetColorScheme, ColorScheme() = %v, want %v", got, cs)
	}
}

// TestWindowSetColorSchemeDifferentSchemesAreDistinct is a falsifying test: setting
// a different scheme must replace the previous one.
// Spec: "After calling SetColorScheme, window.ColorScheme() returns the set scheme (not the owner-chain scheme)"
func TestWindowSetColorSchemeDifferentSchemesAreDistinct(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	first := &theme.ColorScheme{}
	second := &theme.ColorScheme{}
	*first = *theme.BorlandBlue
	*second = *theme.BorlandBlue

	w.SetColorScheme(first)
	w.SetColorScheme(second)
	got := w.ColorScheme()

	if got == first {
		t.Errorf("after two SetColorScheme calls, ColorScheme() still returns first scheme — should return second")
	}
	if got != second {
		t.Errorf("after SetColorScheme(second), ColorScheme() = %v, want second %v", got, second)
	}
}

// ---------------------------------------------------------------------------
// Requirement 2 — Window.SetColorScheme(nil) falls back to owner chain
// Spec: "After calling SetColorScheme(nil), window.ColorScheme() falls back to the owner-chain scheme"
// ---------------------------------------------------------------------------

// TestWindowSetColorSchemeNilFallsBackToOwnerChain verifies that after calling
// SetColorScheme(nil), ColorScheme() walks up to the owner's scheme.
func TestWindowSetColorSchemeNilFallsBackToOwnerChain(t *testing.T) {
	// Build owner chain: parent group with a scheme, window as child.
	parent := NewGroup(NewRect(0, 0, 80, 25))
	parent.scheme = theme.BorlandBlue

	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	parent.Insert(w) // sets w's owner to parent

	// First set a scheme, then clear it.
	own := &theme.ColorScheme{}
	*own = *theme.BorlandBlue
	w.SetColorScheme(own)
	w.SetColorScheme(nil)

	got := w.ColorScheme()

	if got != theme.BorlandBlue {
		t.Errorf("after SetColorScheme(nil), ColorScheme() = %v, want owner's scheme %v", got, theme.BorlandBlue)
	}
}

// TestWindowSetColorSchemeNilWithNoOwnerReturnsNil is a falsifying test: with no
// owner chain and nil scheme, ColorScheme() must return nil, not a stale scheme.
// Spec: "After calling SetColorScheme(nil), window.ColorScheme() falls back to the owner-chain scheme"
func TestWindowSetColorSchemeNilWithNoOwnerReturnsNil(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	cs := &theme.ColorScheme{}
	*cs = *theme.BorlandBlue

	w.SetColorScheme(cs)
	w.SetColorScheme(nil)

	got := w.ColorScheme()

	if got != nil {
		t.Errorf("after SetColorScheme(nil) with no owner, ColorScheme() = %v, want nil", got)
	}
}

// ---------------------------------------------------------------------------
// Requirement 3 — Dialog.SetColorScheme stores the scheme
// Spec: "SetColorScheme(cs *theme.ColorScheme) sets the BaseView.scheme field on the Dialog"
// Spec: "After calling SetColorScheme, dialog.ColorScheme() returns the set scheme"
// ---------------------------------------------------------------------------

// TestDialogSetColorSchemeColorSchemeReturnsIt verifies that after calling
// SetColorScheme with a non-nil scheme, a Dialog's ColorScheme() returns that scheme.
func TestDialogSetColorSchemeColorSchemeReturnsIt(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 30, 15), "Test")
	cs := &theme.ColorScheme{}
	*cs = *theme.BorlandBlue

	d.SetColorScheme(cs)
	got := d.ColorScheme()

	if got != cs {
		t.Errorf("after SetColorScheme on Dialog, ColorScheme() = %v, want %v", got, cs)
	}
}

// TestDialogSetColorSchemeDifferentSchemesAreDistinct is a falsifying test:
// setting a second scheme on a Dialog must displace the first.
// Spec: "After calling SetColorScheme, dialog.ColorScheme() returns the set scheme"
func TestDialogSetColorSchemeDifferentSchemesAreDistinct(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 30, 15), "Test")
	first := &theme.ColorScheme{}
	second := &theme.ColorScheme{}
	*first = *theme.BorlandBlue
	*second = *theme.BorlandBlue

	d.SetColorScheme(first)
	d.SetColorScheme(second)
	got := d.ColorScheme()

	if got == first {
		t.Errorf("after two SetColorScheme calls on Dialog, ColorScheme() still returns first — should return second")
	}
	if got != second {
		t.Errorf("after SetColorScheme(second) on Dialog, ColorScheme() = %v, want second %v", got, second)
	}
}

// ---------------------------------------------------------------------------
// Requirement 4 — Dialog.SetColorScheme(nil) falls back to owner chain
// Spec: "After calling SetColorScheme(nil), dialog.ColorScheme() falls back to the owner-chain scheme"
// ---------------------------------------------------------------------------

// TestDialogSetColorSchemeNilFallsBackToOwnerChain verifies that after calling
// SetColorScheme(nil) on a Dialog, ColorScheme() returns the owner's scheme.
func TestDialogSetColorSchemeNilFallsBackToOwnerChain(t *testing.T) {
	parent := NewGroup(NewRect(0, 0, 80, 25))
	parent.scheme = theme.BorlandBlue

	d := NewDialog(NewRect(0, 0, 30, 15), "Test")
	parent.Insert(d) // sets d's owner to parent

	own := &theme.ColorScheme{}
	*own = *theme.BorlandBlue
	d.SetColorScheme(own)
	d.SetColorScheme(nil)

	got := d.ColorScheme()

	if got != theme.BorlandBlue {
		t.Errorf("after SetColorScheme(nil) on Dialog, ColorScheme() = %v, want owner's scheme %v", got, theme.BorlandBlue)
	}
}

// TestDialogSetColorSchemeNilWithNoOwnerReturnsNil is a falsifying test:
// a Dialog with nil scheme and no owner must return nil, not a stale scheme.
// Spec: "After calling SetColorScheme(nil), dialog.ColorScheme() falls back to the owner-chain scheme"
func TestDialogSetColorSchemeNilWithNoOwnerReturnsNil(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 30, 15), "Test")
	cs := &theme.ColorScheme{}
	*cs = *theme.BorlandBlue

	d.SetColorScheme(cs)
	d.SetColorScheme(nil)

	got := d.ColorScheme()

	if got != nil {
		t.Errorf("after SetColorScheme(nil) on Dialog with no owner, ColorScheme() = %v, want nil", got)
	}
}

// ---------------------------------------------------------------------------
// Requirement 5 — Children of a window inherit the window's scheme
// Spec: "Children of the window inherit the window's scheme via ColorScheme() owner-chain walk"
// ---------------------------------------------------------------------------

// TestWindowChildInheritsWindowScheme verifies that a child inserted into a window
// whose scheme was set via SetColorScheme receives that scheme from its own ColorScheme().
func TestWindowChildInheritsWindowScheme(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	cs := &theme.ColorScheme{}
	*cs = *theme.BorlandBlue
	w.SetColorScheme(cs)

	btn := NewButton(NewRect(0, 0, 10, 1), "OK", CmOK)
	w.Insert(btn)

	got := btn.ColorScheme()

	if got != cs {
		t.Errorf("child's ColorScheme() = %v, want window's scheme %v", got, cs)
	}
}

// ---------------------------------------------------------------------------
// Requirement 6 (falsifying) — scheme isolation: setting scheme on one window
// does not change a sibling window's scheme.
// Spec: "SetColorScheme(cs) sets the BaseView.scheme field on the Window" — only on that window.
// ---------------------------------------------------------------------------

// TestWindowSetColorSchemeDoesNotAffectSiblingWindow verifies that setting a custom
// scheme on one window does not alter a sibling window's ColorScheme() result.
func TestWindowSetColorSchemeDoesNotAffectSiblingWindow(t *testing.T) {
	parent := NewGroup(NewRect(0, 0, 80, 25))
	parent.scheme = theme.BorlandBlue

	w1 := NewWindow(NewRect(0, 0, 40, 20), "W1")
	w2 := NewWindow(NewRect(0, 0, 40, 20), "W2")
	parent.Insert(w1)
	parent.Insert(w2)

	custom := &theme.ColorScheme{}
	*custom = *theme.BorlandBlue
	w1.SetColorScheme(custom)

	// w2 has no scheme set directly — it should still get parent's scheme.
	got := w2.ColorScheme()

	if got == custom {
		t.Errorf("setting scheme on w1 affected w2: w2.ColorScheme() returned w1's custom scheme")
	}
	if got != theme.BorlandBlue {
		t.Errorf("w2.ColorScheme() = %v, want parent's BorlandBlue (not affected by sibling)", got)
	}
}

// ---------------------------------------------------------------------------
// Requirement 7 (falsifying) — child gets window's scheme, not nil
// Spec: "Children of the window inherit the window's scheme via ColorScheme() owner-chain walk"
// A child inserted into a scheme-bearing window must NOT return nil.
// ---------------------------------------------------------------------------

// TestWindowChildDoesNotGetNilWhenWindowHasScheme is a falsifying test: a child
// inserted into a window with a custom scheme must not get nil from ColorScheme().
func TestWindowChildDoesNotGetNilWhenWindowHasScheme(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	cs := &theme.ColorScheme{}
	*cs = *theme.BorlandBlue
	w.SetColorScheme(cs)

	btn := NewButton(NewRect(0, 0, 10, 1), "OK", CmOK)
	w.Insert(btn)

	got := btn.ColorScheme()

	if got == nil {
		t.Error("child.ColorScheme() returned nil — expected to inherit window's scheme via owner-chain walk")
	}
}
