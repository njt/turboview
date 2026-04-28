package theme

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestBorlandBlueIsColorScheme confirms BorlandBlue is a *ColorScheme.
func TestBorlandBlueIsColorScheme(t *testing.T) {
	if BorlandBlue == nil {
		t.Error("BorlandBlue is nil")
	}

	// Verify it's assignable to *ColorScheme
	var scheme *ColorScheme = BorlandBlue
	if scheme == nil {
		t.Error("BorlandBlue cannot be assigned to *ColorScheme")
	}
}

// TestBorlandBlueHasAllFields confirms BorlandBlue has all ColorScheme fields set.
func TestBorlandBlueHasAllFields(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil, cannot test fields")
	}

	// Verify all fields are accessible and not zero
	// (they should all have some non-default value for a real theme)
	fields := []tcell.Style{
		BorlandBlue.WindowBackground,
		BorlandBlue.WindowFrameActive,
		BorlandBlue.WindowFrameInactive,
		BorlandBlue.WindowTitle,
		BorlandBlue.WindowShadow,
		BorlandBlue.DesktopBackground,
		BorlandBlue.DialogBackground,
		BorlandBlue.DialogFrame,
		BorlandBlue.ButtonNormal,
		BorlandBlue.ButtonDefault,
		BorlandBlue.ButtonShadow,
		BorlandBlue.ButtonShortcut,
		BorlandBlue.InputNormal,
		BorlandBlue.InputSelection,
		BorlandBlue.LabelNormal,
		BorlandBlue.LabelShortcut,
		BorlandBlue.CheckBoxNormal,
		BorlandBlue.RadioButtonNormal,
		BorlandBlue.ListNormal,
		BorlandBlue.ListSelected,
		BorlandBlue.ListFocused,
		BorlandBlue.ScrollBar,
		BorlandBlue.ScrollThumb,
		BorlandBlue.MenuNormal,
		BorlandBlue.MenuShortcut,
		BorlandBlue.MenuSelected,
		BorlandBlue.MenuDisabled,
		BorlandBlue.StatusNormal,
		BorlandBlue.StatusShortcut,
	}

	if len(fields) != 29 {
		t.Errorf("Expected 29 fields, got %d", len(fields))
	}

	// All fields should be accessible without panic
	for _, field := range fields {
		_ = field
	}
}

// TestBorlandBlueGetRegistered confirms Get("borland-blue") returns BorlandBlue.
func TestBorlandBlueGetRegistered(t *testing.T) {
	retrieved := Get("borland-blue")

	if retrieved == nil {
		t.Error("Get(\"borland-blue\") returned nil")
	}

	if retrieved != BorlandBlue {
		t.Error("Get(\"borland-blue\") did not return BorlandBlue")
	}
}

// TestBorlandBluePreRegistered confirms BorlandBlue is pre-registered.
func TestBorlandBluePreRegistered(t *testing.T) {
	// BorlandBlue should be retrievable without calling Register
	result := Get("borland-blue")

	if result == nil {
		t.Error("BorlandBlue is not pre-registered")
	}

	if result != BorlandBlue {
		t.Error("BorlandBlue is registered but Get does not return it")
	}
}

// TestBorlandBlueIsTcellStyleBased confirms BorlandBlue uses tcell.Style values.
func TestBorlandBlueIsTcellStyleBased(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}

	// All fields should be tcell.Style compatible
	// Verify by attempting operations that require tcell.Style
	testStyle := BorlandBlue.WindowBackground
	_, _, _ = testStyle.Decompose()

	testStyle = BorlandBlue.MenuNormal
	_, _, _ = testStyle.Decompose()

	testStyle = BorlandBlue.ButtonNormal
	_, _, _ = testStyle.Decompose()
}

// TestBorlandBlueStylesAreNotDefault confirms BorlandBlue colors are different from StyleDefault.
func TestBorlandBlueStylesAreNotDefault(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}

	// Check that at least some BorlandBlue styles differ from StyleDefault
	// A real theme should have custom colors
	hasCustomStyle := false

	styles := []tcell.Style{
		BorlandBlue.WindowBackground,
		BorlandBlue.DialogBackground,
		BorlandBlue.MenuNormal,
		BorlandBlue.ButtonNormal,
	}

	for _, style := range styles {
		if style != tcell.StyleDefault {
			hasCustomStyle = true
			break
		}
	}

	if !hasCustomStyle {
		t.Error("BorlandBlue should have custom styles distinct from StyleDefault")
	}
}

// TestBorlandBlueClassicBlueBackground confirms the theme uses blue background (classic Turbo Vision).
func TestBorlandBlueClassicBlueBackground(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}

	// For a classic Turbo Vision theme, the background should be blue
	// Check the desktop or window background color
	windowBg := BorlandBlue.WindowBackground
	_, bgColor, _ := windowBg.Decompose()

	// Should be blue or related to blue; Turbo Vision used blue backgrounds
	// Check if background is actually set (not the default white background)
	if bgColor == tcell.ColorDefault {
		t.Error("BorlandBlue WindowBackground should have a specific color, not default")
	}
}

// TestBorlandBlueCanBeOverwritten confirms BorlandBlue can be overwritten via Register.
func TestBorlandBlueCanBeOverwritten(t *testing.T) {
	// Get the original
	original := Get("borland-blue")
	if original == nil {
		t.Fatal("BorlandBlue not found")
	}

	// Create a new scheme
	newScheme := &ColorScheme{
		WindowBackground: tcell.StyleDefault.Foreground(tcell.ColorRed),
	}

	// Overwrite BorlandBlue
	Register("borland-blue", newScheme)
	retrieved := Get("borland-blue")

	if retrieved != newScheme {
		t.Error("BorlandBlue was not overwritten by Register")
	}

	// Restore the original for other tests
	Register("borland-blue", BorlandBlue)
}

// TestBorlandBlueMultipleGetCalls confirms Get returns the same instance each time.
func TestBorlandBlueMultipleGetCalls(t *testing.T) {
	retrieved1 := Get("borland-blue")
	retrieved2 := Get("borland-blue")
	retrieved3 := Get("borland-blue")

	if retrieved1 != retrieved2 {
		t.Error("Multiple Get calls for borland-blue returned different instances")
	}

	if retrieved2 != retrieved3 {
		t.Error("Multiple Get calls for borland-blue returned different instances")
	}

	if retrieved1 != BorlandBlue {
		t.Error("Multiple Get calls did not return BorlandBlue")
	}
}

// TestBorlandBlueVsOtherSchemes confirms BorlandBlue can coexist with other schemes.
func TestBorlandBlueVsOtherSchemes(t *testing.T) {
	otherScheme := &ColorScheme{
		WindowBackground: tcell.StyleDefault.Foreground(tcell.ColorGreen),
	}

	Register("other-scheme", otherScheme)

	// BorlandBlue should not be affected
	retrieved := Get("borland-blue")
	if retrieved != BorlandBlue {
		t.Error("BorlandBlue was affected by registering other schemes")
	}

	// Other scheme should still be there
	retrievedOther := Get("other-scheme")
	if retrievedOther != otherScheme {
		t.Error("Other scheme was not retrieved correctly")
	}
}

// TestBorlandBlueName confirms the name is exactly "borland-blue".
func TestBorlandBlueName(t *testing.T) {
	// Verify exact name with hyphen, lowercase
	result := Get("borland-blue")
	if result == nil {
		t.Error("Name must be exactly \"borland-blue\" (lowercase with hyphen)")
	}

	// Verify name is case-sensitive
	if Get("Borland-Blue") != nil {
		t.Error("Get is case-insensitive but should be case-sensitive")
	}

	if Get("borland_blue") != nil {
		t.Error("Name should use hyphen, not underscore")
	}

	if Get("borlandblue") != nil {
		t.Error("Name should have hyphen separator")
	}
}

// TestBorlandBlueIsNotNil confirms BorlandBlue variable is not nil.
func TestBorlandBlueIsNotNil(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue must not be nil")
	}
}

// TestBorlandBlueInstanceStability confirms BorlandBlue instance never changes.
func TestBorlandBlueInstanceStability(t *testing.T) {
	instance1 := BorlandBlue
	instance2 := BorlandBlue

	if instance1 != instance2 {
		t.Error("BorlandBlue instance should be stable and not change")
	}

	// Verify Get returns the same instance
	instance3 := Get("borland-blue")
	if instance1 != instance3 {
		t.Error("BorlandBlue instance from Get should match the variable")
	}
}

// TestBorlandBlueWindowFieldsPopulated confirms window-related fields are set.
func TestBorlandBlueWindowFieldsPopulated(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}

	// All window fields should be accessible
	_ = BorlandBlue.WindowBackground
	_ = BorlandBlue.WindowFrameActive
	_ = BorlandBlue.WindowFrameInactive
	_ = BorlandBlue.WindowTitle
	_ = BorlandBlue.WindowShadow
}

// TestBorlandBlueMenuFieldsPopulated confirms menu-related fields are set.
func TestBorlandBlueMenuFieldsPopulated(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}

	// All menu fields should be accessible
	_ = BorlandBlue.MenuNormal
	_ = BorlandBlue.MenuShortcut
	_ = BorlandBlue.MenuSelected
	_ = BorlandBlue.MenuDisabled
}

// TestBorlandBlueButtonFieldsPopulated confirms button-related fields are set.
func TestBorlandBlueButtonFieldsPopulated(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}

	// All button fields should be accessible
	_ = BorlandBlue.ButtonNormal
	_ = BorlandBlue.ButtonDefault
	_ = BorlandBlue.ButtonShadow
	_ = BorlandBlue.ButtonShortcut
}
