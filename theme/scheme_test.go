package theme

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestColorSchemeStructExists confirms ColorScheme is a struct with all required fields.
func TestColorSchemeStructExists(t *testing.T) {
	scheme := &ColorScheme{}

	// Verify all required fields exist by checking they can be assigned
	scheme.WindowBackground = tcell.StyleDefault
	scheme.WindowFrameActive = tcell.StyleDefault
	scheme.WindowFrameInactive = tcell.StyleDefault
	scheme.WindowTitle = tcell.StyleDefault
	scheme.WindowShadow = tcell.StyleDefault
	scheme.WindowBody = tcell.StyleDefault
	scheme.DesktopBackground = tcell.StyleDefault
	scheme.DialogBackground = tcell.StyleDefault
	scheme.DialogFrame = tcell.StyleDefault
	scheme.ButtonNormal = tcell.StyleDefault
	scheme.ButtonDefault = tcell.StyleDefault
	scheme.ButtonShadow = tcell.StyleDefault
	scheme.ButtonShortcut = tcell.StyleDefault
	scheme.InputNormal = tcell.StyleDefault
	scheme.InputSelection = tcell.StyleDefault
	scheme.LabelNormal = tcell.StyleDefault
	scheme.LabelShortcut = tcell.StyleDefault
	scheme.CheckBoxNormal = tcell.StyleDefault
	scheme.RadioButtonNormal = tcell.StyleDefault
	scheme.ListNormal = tcell.StyleDefault
	scheme.ListSelected = tcell.StyleDefault
	scheme.ListFocused = tcell.StyleDefault
	scheme.ScrollBar = tcell.StyleDefault
	scheme.ScrollThumb = tcell.StyleDefault
	scheme.MenuNormal = tcell.StyleDefault
	scheme.MenuShortcut = tcell.StyleDefault
	scheme.MenuSelected = tcell.StyleDefault
	scheme.MenuDisabled = tcell.StyleDefault
	scheme.StatusNormal = tcell.StyleDefault
	scheme.StatusShortcut = tcell.StyleDefault

	// If all assignments succeed, the struct has all required fields
	if scheme == nil {
		t.Error("ColorScheme is nil")
	}
}

// TestColorSchemeFieldTypes confirms all fields are tcell.Style.
func TestColorSchemeFieldTypes(t *testing.T) {
	scheme := &ColorScheme{}

	// Create a style with specific values to verify type compatibility
	white := tcell.ColorWhite
	blue := tcell.ColorBlue
	style := tcell.StyleDefault.Foreground(white).Background(blue)

	scheme.WindowBackground = style
	scheme.WindowFrameActive = style
	scheme.WindowFrameInactive = style
	scheme.WindowTitle = style
	scheme.WindowShadow = style
	scheme.WindowBody = style
	scheme.DesktopBackground = style
	scheme.DialogBackground = style
	scheme.DialogFrame = style
	scheme.ButtonNormal = style
	scheme.ButtonDefault = style
	scheme.ButtonShadow = style
	scheme.ButtonShortcut = style
	scheme.InputNormal = style
	scheme.InputSelection = style
	scheme.LabelNormal = style
	scheme.LabelShortcut = style
	scheme.CheckBoxNormal = style
	scheme.RadioButtonNormal = style
	scheme.ListNormal = style
	scheme.ListSelected = style
	scheme.ListFocused = style
	scheme.ScrollBar = style
	scheme.ScrollThumb = style
	scheme.MenuNormal = style
	scheme.MenuShortcut = style
	scheme.MenuSelected = style
	scheme.MenuDisabled = style
	scheme.StatusNormal = style
	scheme.StatusShortcut = style

	// Verify fields can be read back
	if scheme.WindowBackground != style {
		t.Error("WindowBackground not correctly assigned")
	}
	if scheme.ButtonNormal != style {
		t.Error("ButtonNormal not correctly assigned")
	}
	if scheme.MenuSelected != style {
		t.Error("MenuSelected not correctly assigned")
	}
	if scheme.StatusShortcut != style {
		t.Error("StatusShortcut not correctly assigned")
	}
}

// TestColorSchemeCanBeCreatedWithZeroValue confirms ColorScheme works with zero initialization.
func TestColorSchemeCanBeCreatedWithZeroValue(t *testing.T) {
	scheme := &ColorScheme{}

	if scheme == nil {
		t.Error("ColorScheme cannot be created with zero value")
	}

	// All fields should be assignable after zero init
	scheme.WindowBackground = tcell.StyleDefault
	if scheme.WindowBackground != tcell.StyleDefault {
		t.Error("ColorScheme field not initialized correctly")
	}
}

// TestColorSchemeDifferentValuesPerField confirms each field holds its own value independently.
func TestColorSchemeDifferentValuesPerField(t *testing.T) {
	scheme := &ColorScheme{}

	style1 := tcell.StyleDefault.Foreground(tcell.ColorRed)
	style2 := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	style3 := tcell.StyleDefault.Foreground(tcell.ColorBlue)

	scheme.WindowBackground = style1
	scheme.ButtonNormal = style2
	scheme.MenuNormal = style3

	if scheme.WindowBackground != style1 {
		t.Error("WindowBackground changed unexpectedly")
	}
	if scheme.ButtonNormal != style2 {
		t.Error("ButtonNormal changed unexpectedly")
	}
	if scheme.MenuNormal != style3 {
		t.Error("MenuNormal changed unexpectedly")
	}
}

// TestColorSchemeWindowFields confirms all window-related fields exist.
func TestColorSchemeWindowFields(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.WindowBackground = style
	scheme.WindowFrameActive = style
	scheme.WindowFrameInactive = style
	scheme.WindowTitle = style
	scheme.WindowShadow = style

	if scheme.WindowBackground != style || scheme.WindowFrameActive != style ||
		scheme.WindowFrameInactive != style || scheme.WindowTitle != style ||
		scheme.WindowShadow != style {
		t.Error("One or more window fields are missing or not assignable")
	}
}

// TestColorSchemeDialogFields confirms all dialog-related fields exist.
func TestColorSchemeDialogFields(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.DialogBackground = style
	scheme.DialogFrame = style

	if scheme.DialogBackground != style || scheme.DialogFrame != style {
		t.Error("One or more dialog fields are missing or not assignable")
	}
}

// TestColorSchemeButtonFields confirms all button-related fields exist.
func TestColorSchemeButtonFields(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.ButtonNormal = style
	scheme.ButtonDefault = style
	scheme.ButtonShadow = style
	scheme.ButtonShortcut = style

	if scheme.ButtonNormal != style || scheme.ButtonDefault != style ||
		scheme.ButtonShadow != style || scheme.ButtonShortcut != style {
		t.Error("One or more button fields are missing or not assignable")
	}
}

// TestColorSchemeInputFields confirms all input-related fields exist.
func TestColorSchemeInputFields(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.InputNormal = style
	scheme.InputSelection = style

	if scheme.InputNormal != style || scheme.InputSelection != style {
		t.Error("One or more input fields are missing or not assignable")
	}
}

// TestColorSchemeLabelFields confirms all label-related fields exist.
func TestColorSchemeLabelFields(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.LabelNormal = style
	scheme.LabelShortcut = style

	if scheme.LabelNormal != style || scheme.LabelShortcut != style {
		t.Error("One or more label fields are missing or not assignable")
	}
}

// TestColorSchemeCheckBoxFields confirms all checkbox-related fields exist.
func TestColorSchemeCheckBoxFields(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.CheckBoxNormal = style

	if scheme.CheckBoxNormal != style {
		t.Error("CheckBoxNormal field is missing or not assignable")
	}
}

// TestColorSchemeRadioButtonFields confirms all radio button-related fields exist.
func TestColorSchemeRadioButtonFields(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.RadioButtonNormal = style

	if scheme.RadioButtonNormal != style {
		t.Error("RadioButtonNormal field is missing or not assignable")
	}
}

// TestColorSchemeListFields confirms all list-related fields exist.
func TestColorSchemeListFields(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.ListNormal = style
	scheme.ListSelected = style
	scheme.ListFocused = style

	if scheme.ListNormal != style || scheme.ListSelected != style || scheme.ListFocused != style {
		t.Error("One or more list fields are missing or not assignable")
	}
}

// TestColorSchemeScrollBarFields confirms all scrollbar-related fields exist.
func TestColorSchemeScrollBarFields(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.ScrollBar = style
	scheme.ScrollThumb = style

	if scheme.ScrollBar != style || scheme.ScrollThumb != style {
		t.Error("One or more scrollbar fields are missing or not assignable")
	}
}

// TestColorSchemeMenuFields confirms all menu-related fields exist.
func TestColorSchemeMenuFields(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.MenuNormal = style
	scheme.MenuShortcut = style
	scheme.MenuSelected = style
	scheme.MenuDisabled = style

	if scheme.MenuNormal != style || scheme.MenuShortcut != style ||
		scheme.MenuSelected != style || scheme.MenuDisabled != style {
		t.Error("One or more menu fields are missing or not assignable")
	}
}

// TestColorSchemeStatusFields confirms all status-related fields exist.
func TestColorSchemeStatusFields(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.StatusNormal = style
	scheme.StatusShortcut = style

	if scheme.StatusNormal != style || scheme.StatusShortcut != style {
		t.Error("One or more status fields are missing or not assignable")
	}
}

// TestColorSchemeDesktopField confirms desktop field exists.
func TestColorSchemeDesktopField(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault

	scheme.DesktopBackground = style

	if scheme.DesktopBackground != style {
		t.Error("DesktopBackground field is missing or not assignable")
	}
}
