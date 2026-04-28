package tv

import (
	"testing"

	"github.com/njt/turboview/theme"
)

// mockContainer is a test double that implements Container interface
// using BaseView for the View part, with stub implementations for Container methods.
type mockContainer struct {
	BaseView
}

func (m *mockContainer) Insert(child View) {
}

func (m *mockContainer) Remove(child View) {
}

func (m *mockContainer) Children() []View {
	return nil
}

func (m *mockContainer) FocusedChild() View {
	return nil
}

func (m *mockContainer) SetFocusedChild(child View) {
}

func (m *mockContainer) ExecView(v View) CommandCode {
	return CmCancel
}

// TestBaseViewBounds tests SetBounds and Bounds.
func TestBaseViewBounds(t *testing.T) {
	bv := &BaseView{}
	r := NewRect(10, 20, 30, 40)

	bv.SetBounds(r)
	got := bv.Bounds()

	if got != r {
		t.Errorf("Bounds() after SetBounds(%v) = %v, want %v", r, got, r)
	}
}

// TestBaseViewBoundsZero tests Bounds with zero rect.
func TestBaseViewBoundsZero(t *testing.T) {
	bv := &BaseView{}
	zero := NewRect(0, 0, 0, 0)

	bv.SetBounds(zero)
	got := bv.Bounds()

	if got != zero {
		t.Errorf("Bounds() after SetBounds(zero) = %v, want %v", got, zero)
	}
}

// TestBaseViewBoundsNegative tests Bounds with negative coordinates.
func TestBaseViewBoundsNegative(t *testing.T) {
	bv := &BaseView{}
	r := NewRect(-5, -10, 20, 30)

	bv.SetBounds(r)
	got := bv.Bounds()

	if got != r {
		t.Errorf("Bounds() after SetBounds(%v) = %v, want %v", r, got, r)
	}
}

// TestBaseViewGrowMode tests SetGrowMode and GrowMode.
func TestBaseViewGrowMode(t *testing.T) {
	bv := &BaseView{}
	mode := GfGrowAll

	bv.SetGrowMode(mode)
	got := bv.GrowMode()

	if got != mode {
		t.Errorf("GrowMode() after SetGrowMode(%v) = %v, want %v", mode, got, mode)
	}
}

// TestBaseViewOwner tests SetOwner and Owner.
func TestBaseViewOwner(t *testing.T) {
	bv := &BaseView{}
	container := &mockContainer{}

	bv.SetOwner(container)
	got := bv.Owner()

	if got != container {
		t.Errorf("Owner() after SetOwner() = %v, want %v", got, container)
	}
}

// TestBaseViewOwnerNil tests Owner when not set.
func TestBaseViewOwnerNil(t *testing.T) {
	bv := &BaseView{}

	got := bv.Owner()

	if got != nil {
		t.Errorf("Owner() without SetOwner() = %v, want nil", got)
	}
}

// TestBaseViewState tests SetState and State.
func TestBaseViewState(t *testing.T) {
	bv := &BaseView{}
	state := SfVisible | SfFocused

	bv.state = state
	got := bv.State()

	if got != state {
		t.Errorf("State() = %v, want %v", got, state)
	}
}

// TestBaseViewSetStateFlag tests SetState with true sets the flag.
func TestBaseViewSetStateFlag(t *testing.T) {
	bv := &BaseView{}
	flag := SfVisible

	bv.SetState(flag, true)

	if !bv.HasState(flag) {
		t.Errorf("SetState(%v, true) did not set flag", flag)
	}
}

// TestBaseViewClearStateFlag tests SetState with false clears the flag.
func TestBaseViewClearStateFlag(t *testing.T) {
	bv := &BaseView{}
	flag := SfVisible
	bv.state = flag

	bv.SetState(flag, false)

	if bv.HasState(flag) {
		t.Errorf("SetState(%v, false) did not clear flag", flag)
	}
}

// TestBaseViewSetStateFlagIndependent tests SetState doesn't affect other flags.
func TestBaseViewSetStateFlagIndependent(t *testing.T) {
	bv := &BaseView{}
	bv.state = SfSelected

	bv.SetState(SfVisible, true)

	if !bv.HasState(SfVisible) || !bv.HasState(SfSelected) {
		t.Errorf("SetState(SfVisible, true) affected other flags")
	}
}

// TestBaseViewHasStateTrue tests HasState returns true when flag is set.
func TestBaseViewHasStateTrue(t *testing.T) {
	bv := &BaseView{}
	flag := SfFocused
	bv.state = flag

	if !bv.HasState(flag) {
		t.Errorf("HasState(%v) = false, want true", flag)
	}
}

// TestBaseViewHasStateFalse tests HasState returns false when flag is not set.
func TestBaseViewHasStateFalse(t *testing.T) {
	bv := &BaseView{}

	if bv.HasState(SfFocused) {
		t.Errorf("HasState(SfFocused) = true, want false")
	}
}

// TestBaseViewHasStateMultiple tests HasState with multiple flags set.
func TestBaseViewHasStateMultiple(t *testing.T) {
	bv := &BaseView{}
	bv.state = SfVisible | SfFocused | SfSelected

	if !bv.HasState(SfVisible) || !bv.HasState(SfFocused) || !bv.HasState(SfSelected) {
		t.Errorf("HasState() failed with multiple flags set")
	}

	if bv.HasState(SfModal) {
		t.Errorf("HasState(SfModal) = true, want false when not set")
	}
}

// TestBaseViewEventMask tests SetEventMask and EventMask.
func TestBaseViewEventMask(t *testing.T) {
	bv := &BaseView{}
	mask := EvKeyboard | EvMouse

	bv.SetEventMask(mask)
	got := bv.EventMask()

	if got != mask {
		t.Errorf("EventMask() after SetEventMask(%v) = %v, want %v", mask, got, mask)
	}
}

// TestBaseViewOptions tests Options returns the options.
func TestBaseViewOptions(t *testing.T) {
	bv := &BaseView{}
	options := OfSelectable | OfCentered
	bv.options = options

	got := bv.Options()

	if got != options {
		t.Errorf("Options() = %v, want %v", got, options)
	}
}

// TestBaseViewSetOptionsFlag tests SetOptions with true sets the flag.
func TestBaseViewSetOptionsFlag(t *testing.T) {
	bv := &BaseView{}
	flag := OfSelectable

	bv.SetOptions(flag, true)

	if !bv.HasOption(flag) {
		t.Errorf("SetOptions(%v, true) did not set flag", flag)
	}
}

// TestBaseViewClearOptionsFlag tests SetOptions with false clears the flag.
func TestBaseViewClearOptionsFlag(t *testing.T) {
	bv := &BaseView{}
	flag := OfSelectable
	bv.options = flag

	bv.SetOptions(flag, false)

	if bv.HasOption(flag) {
		t.Errorf("SetOptions(%v, false) did not clear flag", flag)
	}
}

// TestBaseViewSetOptionsFlagIndependent tests SetOptions doesn't affect other flags.
func TestBaseViewSetOptionsFlagIndependent(t *testing.T) {
	bv := &BaseView{}
	bv.options = OfCentered

	bv.SetOptions(OfSelectable, true)

	if !bv.HasOption(OfSelectable) || !bv.HasOption(OfCentered) {
		t.Errorf("SetOptions(OfSelectable, true) affected other options")
	}
}

// TestBaseViewHasOptionTrue tests HasOption returns true when flag is set.
func TestBaseViewHasOptionTrue(t *testing.T) {
	bv := &BaseView{}
	flag := OfTopSelect
	bv.options = flag

	if !bv.HasOption(flag) {
		t.Errorf("HasOption(%v) = false, want true", flag)
	}
}

// TestBaseViewHasOptionFalse tests HasOption returns false when flag is not set.
func TestBaseViewHasOptionFalse(t *testing.T) {
	bv := &BaseView{}

	if bv.HasOption(OfTopSelect) {
		t.Errorf("HasOption(OfTopSelect) = true, want false")
	}
}

// TestBaseViewHasOptionMultiple tests HasOption with multiple flags set.
func TestBaseViewHasOptionMultiple(t *testing.T) {
	bv := &BaseView{}
	bv.options = OfSelectable | OfTopSelect | OfCentered

	if !bv.HasOption(OfSelectable) || !bv.HasOption(OfTopSelect) || !bv.HasOption(OfCentered) {
		t.Errorf("HasOption() failed with multiple flags set")
	}

	if bv.HasOption(OfFirstClick) {
		t.Errorf("HasOption(OfFirstClick) = true, want false when not set")
	}
}

// TestBaseViewColorSchemeOwnScheme tests ColorScheme returns own scheme when set.
func TestBaseViewColorSchemeOwnScheme(t *testing.T) {
	bv := &BaseView{}
	scheme := &theme.ColorScheme{}
	bv.scheme = scheme

	got := bv.ColorScheme()

	if got != scheme {
		t.Errorf("ColorScheme() = %v, want %v", got, scheme)
	}
}

// TestBaseViewColorSchemeNil tests ColorScheme returns nil when not set and no owner.
func TestBaseViewColorSchemeNil(t *testing.T) {
	bv := &BaseView{}

	got := bv.ColorScheme()

	if got != nil {
		t.Errorf("ColorScheme() = %v, want nil", got)
	}
}

// TestBaseViewColorSchemeFromOwner tests ColorScheme walks up owner chain.
func TestBaseViewColorSchemeFromOwner(t *testing.T) {
	bv := &BaseView{}
	owner := &mockContainer{}
	ownerScheme := &theme.ColorScheme{}
	owner.scheme = ownerScheme

	bv.SetOwner(owner)
	got := bv.ColorScheme()

	if got != ownerScheme {
		t.Errorf("ColorScheme() from owner = %v, want %v", got, ownerScheme)
	}
}

// TestBaseViewColorSchemeFromOwnerChain tests ColorScheme walks up multiple levels.
func TestBaseViewColorSchemeFromOwnerChain(t *testing.T) {
	bv := &BaseView{}
	parent := &mockContainer{}
	grandparent := &mockContainer{}
	scheme := &theme.ColorScheme{}

	grandparent.scheme = scheme
	parent.SetOwner(grandparent)
	bv.SetOwner(parent)

	got := bv.ColorScheme()

	if got != scheme {
		t.Errorf("ColorScheme() from owner chain = %v, want %v", got, scheme)
	}
}

// TestBaseViewColorSchemeSelfBeforeOwner tests own scheme takes precedence.
func TestBaseViewColorSchemeSelfBeforeOwner(t *testing.T) {
	bv := &BaseView{}
	owner := &mockContainer{}
	ownScheme := &theme.ColorScheme{}
	ownerScheme := &theme.ColorScheme{}

	bv.scheme = ownScheme
	owner.scheme = ownerScheme
	bv.SetOwner(owner)

	got := bv.ColorScheme()

	if got != ownScheme {
		t.Errorf("ColorScheme() = %v, want own scheme %v (not owner %v)", got, ownScheme, ownerScheme)
	}
}

// TestBaseViewDrawNoOp tests Draw is a no-op by default.
func TestBaseViewDrawNoOp(t *testing.T) {
	bv := &BaseView{}
	buf := NewDrawBuffer(10, 10)

	// Should not panic or raise error
	bv.Draw(buf)
}

// TestBaseViewHandleEventNoOp tests HandleEvent is a no-op by default.
func TestBaseViewHandleEventNoOp(t *testing.T) {
	bv := &BaseView{}
	event := &Event{What: EvKeyboard}

	// Should not panic or raise error
	bv.HandleEvent(event)
}

// TestBaseViewInitialStateZero tests initial state is zero.
func TestBaseViewInitialStateZero(t *testing.T) {
	bv := &BaseView{}

	if bv.State() != 0 {
		t.Errorf("initial State() = %v, want 0", bv.State())
	}
}

// TestBaseViewInitialOptionsZero tests initial options are zero.
func TestBaseViewInitialOptionsZero(t *testing.T) {
	bv := &BaseView{}

	if bv.Options() != 0 {
		t.Errorf("initial Options() = %v, want 0", bv.Options())
	}
}

// TestBaseViewInitialEventMaskZero tests initial event mask is zero.
func TestBaseViewInitialEventMaskZero(t *testing.T) {
	bv := &BaseView{}

	if bv.EventMask() != 0 {
		t.Errorf("initial EventMask() = %v, want 0", bv.EventMask())
	}
}

// TestBaseViewMultipleStateChanges tests multiple state changes are independent.
func TestBaseViewMultipleStateChanges(t *testing.T) {
	bv := &BaseView{}

	bv.SetState(SfVisible, true)
	bv.SetState(SfFocused, true)
	bv.SetState(SfSelected, true)

	if !bv.HasState(SfVisible) || !bv.HasState(SfFocused) || !bv.HasState(SfSelected) {
		t.Errorf("multiple SetState calls failed")
	}

	bv.SetState(SfFocused, false)

	if !bv.HasState(SfVisible) || bv.HasState(SfFocused) || !bv.HasState(SfSelected) {
		t.Errorf("clearing one flag affected others")
	}
}

// TestBaseViewBoundsStoresOriginAndSize tests that SetBounds stores both origin and size.
func TestBaseViewBoundsStoresOriginAndSize(t *testing.T) {
	bv := &BaseView{}
	r := NewRect(5, 10, 20, 30)

	bv.SetBounds(r)
	got := bv.Bounds()

	if got.A.X != 5 || got.A.Y != 10 {
		t.Errorf("Bounds() origin = (%d, %d), want (5, 10)", got.A.X, got.A.Y)
	}
	if got.B.X != 25 || got.B.Y != 40 {
		t.Errorf("Bounds() end = (%d, %d), want (25, 40)", got.B.X, got.B.Y)
	}
}

// TestBaseViewAllViewStateFlagsIndependent tests all ViewState flags can be set independently.
func TestBaseViewAllViewStateFlagsIndependent(t *testing.T) {
	flags := []ViewState{SfVisible, SfFocused, SfSelected, SfModal, SfDisabled, SfExposed, SfDragging}

	for i, flag := range flags {
		bv := &BaseView{}
		bv.SetState(flag, true)

		for j, otherFlag := range flags {
			expected := i == j
			got := bv.HasState(otherFlag)
			if got != expected {
				t.Errorf("flag %d: HasState(%d) = %v, want %v", i, j, got, expected)
			}
		}
	}
}

// TestBaseViewAllViewOptionsIndependent tests all ViewOptions flags can be set independently.
func TestBaseViewAllViewOptionsIndependent(t *testing.T) {
	flags := []ViewOptions{OfSelectable, OfTopSelect, OfFirstClick, OfPreProcess, OfPostProcess, OfCentered}

	for i, flag := range flags {
		bv := &BaseView{}
		bv.SetOptions(flag, true)

		for j, otherFlag := range flags {
			expected := i == j
			got := bv.HasOption(otherFlag)
			if got != expected {
				t.Errorf("flag %d: HasOption(%d) = %v, want %v", i, j, got, expected)
			}
		}
	}
}
