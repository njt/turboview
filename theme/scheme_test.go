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

// TestMarkdownColorSchemeFieldsExist confirms ColorScheme has all 18 markdown style fields.
// Spec requirement: "ColorScheme has 18 new fields: MarkdownNormal, MarkdownH1 through
// MarkdownH6, MarkdownBold, MarkdownItalic, MarkdownBoldItalic, MarkdownCode,
// MarkdownCodeBlock, MarkdownBlockquote, MarkdownLink, MarkdownHRule, MarkdownListMarker,
// MarkdownTableBorder, MarkdownDefTerm"
func TestMarkdownColorSchemeFieldsExist(t *testing.T) {
	scheme := &ColorScheme{}

	scheme.MarkdownNormal = tcell.StyleDefault
	scheme.MarkdownH1 = tcell.StyleDefault
	scheme.MarkdownH2 = tcell.StyleDefault
	scheme.MarkdownH3 = tcell.StyleDefault
	scheme.MarkdownH4 = tcell.StyleDefault
	scheme.MarkdownH5 = tcell.StyleDefault
	scheme.MarkdownH6 = tcell.StyleDefault
	scheme.MarkdownBold = tcell.StyleDefault
	scheme.MarkdownItalic = tcell.StyleDefault
	scheme.MarkdownBoldItalic = tcell.StyleDefault
	scheme.MarkdownCode = tcell.StyleDefault
	scheme.MarkdownCodeBlock = tcell.StyleDefault
	scheme.MarkdownBlockquote = tcell.StyleDefault
	scheme.MarkdownLink = tcell.StyleDefault
	scheme.MarkdownHRule = tcell.StyleDefault
	scheme.MarkdownListMarker = tcell.StyleDefault
	scheme.MarkdownTableBorder = tcell.StyleDefault
	scheme.MarkdownDefTerm = tcell.StyleDefault

	// Read back a sample to verify assignments were accepted
	if scheme.MarkdownNormal != tcell.StyleDefault {
		t.Error("MarkdownNormal not correctly assigned")
	}
	if scheme.MarkdownH1 != tcell.StyleDefault {
		t.Error("MarkdownH1 not correctly assigned")
	}
	if scheme.MarkdownH6 != tcell.StyleDefault {
		t.Error("MarkdownH6 not correctly assigned")
	}
	if scheme.MarkdownCode != tcell.StyleDefault {
		t.Error("MarkdownCode not correctly assigned")
	}
	if scheme.MarkdownDefTerm != tcell.StyleDefault {
		t.Error("MarkdownDefTerm not correctly assigned")
	}
}

// TestMarkdownColorSchemeFieldsAreIndependent verifies each markdown field stores its
// own value independently.
// Falsifying for: "ColorScheme has 18 new fields" — catches the shortcut where fields
// share a backing store or alias each other.
func TestMarkdownColorSchemeFieldsAreIndependent(t *testing.T) {
	scheme := &ColorScheme{}

	style1 := tcell.StyleDefault.Foreground(tcell.ColorRed)
	style2 := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	style3 := tcell.StyleDefault.Foreground(tcell.ColorBlue)

	scheme.MarkdownNormal = style1
	scheme.MarkdownH1 = style2
	scheme.MarkdownCode = style3

	if scheme.MarkdownNormal != style1 {
		t.Error("MarkdownNormal changed unexpectedly — fields may share backing store")
	}
	if scheme.MarkdownH1 != style2 {
		t.Error("MarkdownH1 changed unexpectedly — fields may share backing store")
	}
	if scheme.MarkdownCode != style3 {
		t.Error("MarkdownCode changed unexpectedly — fields may share backing store")
	}
}

// TestMarkdownColorSchemeFieldsAreStyleType confirms all markdown fields accept and
// return tcell.Style values.
// Spec requirement: "All fields are tcell.Style type"
func TestMarkdownColorSchemeFieldsAreStyleType(t *testing.T) {
	scheme := &ColorScheme{}

	white := tcell.ColorWhite
	blue := tcell.ColorBlue
	style := tcell.StyleDefault.Foreground(white).Background(blue)

	scheme.MarkdownNormal = style
	scheme.MarkdownH1 = style
	scheme.MarkdownH2 = style
	scheme.MarkdownH3 = style
	scheme.MarkdownH4 = style
	scheme.MarkdownH5 = style
	scheme.MarkdownH6 = style
	scheme.MarkdownBold = style
	scheme.MarkdownItalic = style
	scheme.MarkdownBoldItalic = style
	scheme.MarkdownCode = style
	scheme.MarkdownCodeBlock = style
	scheme.MarkdownBlockquote = style
	scheme.MarkdownLink = style
	scheme.MarkdownHRule = style
	scheme.MarkdownListMarker = style
	scheme.MarkdownTableBorder = style
	scheme.MarkdownDefTerm = style

	// Verify roundtrip: assigned style is returned unchanged
	if scheme.MarkdownNormal != style {
		t.Error("MarkdownNormal did not roundtrip tcell.Style correctly")
	}
	if scheme.MarkdownBold != style {
		t.Error("MarkdownBold did not roundtrip tcell.Style correctly")
	}
	if scheme.MarkdownCode != style {
		t.Error("MarkdownCode did not roundtrip tcell.Style correctly")
	}
	if scheme.MarkdownLink != style {
		t.Error("MarkdownLink did not roundtrip tcell.Style correctly")
	}
}

// TestMarkdownColorSchemeFieldsRetainAttributes verifies markdown fields preserve
// text attributes like Bold, not just color values.
// Falsifying for: "All fields are tcell.Style type" — catches the shortcut where
// fields accept styles but strip attributes (e.g., using a color-pair wrapper
// instead of true tcell.Style).
func TestMarkdownColorSchemeFieldsRetainAttributes(t *testing.T) {
	scheme := &ColorScheme{}

	boldStyle := tcell.StyleDefault.Bold(true)

	scheme.MarkdownBold = boldStyle

	// Verify Bold attribute survived the roundtrip
	_, _, attrs := scheme.MarkdownBold.Decompose()
	if attrs&tcell.AttrBold == 0 {
		t.Error("MarkdownBold lost Bold attribute — fields may not be true tcell.Style")
	}
}

// TestBorlandBlueAssignsAllMarkdownFields confirms BorlandBlue initializes every
// markdown style field to a non-zero value.
// Spec requirement: "BorlandBlue assigns all 18 fields with the values from the
// spec's BorlandBlue Defaults table"
func TestBorlandBlueAssignsAllMarkdownFields(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}

	// Verify each markdown field is not the zero-value StyleDefault
	if BorlandBlue.MarkdownNormal == tcell.StyleDefault {
		t.Error("MarkdownNormal is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownH1 == tcell.StyleDefault {
		t.Error("MarkdownH1 is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownH2 == tcell.StyleDefault {
		t.Error("MarkdownH2 is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownH3 == tcell.StyleDefault {
		t.Error("MarkdownH3 is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownH4 == tcell.StyleDefault {
		t.Error("MarkdownH4 is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownH5 == tcell.StyleDefault {
		t.Error("MarkdownH5 is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownH6 == tcell.StyleDefault {
		t.Error("MarkdownH6 is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownBold == tcell.StyleDefault {
		t.Error("MarkdownBold is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownItalic == tcell.StyleDefault {
		t.Error("MarkdownItalic is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownBoldItalic == tcell.StyleDefault {
		t.Error("MarkdownBoldItalic is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownCode == tcell.StyleDefault {
		t.Error("MarkdownCode is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownCodeBlock == tcell.StyleDefault {
		t.Error("MarkdownCodeBlock is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownBlockquote == tcell.StyleDefault {
		t.Error("MarkdownBlockquote is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownLink == tcell.StyleDefault {
		t.Error("MarkdownLink is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownHRule == tcell.StyleDefault {
		t.Error("MarkdownHRule is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownListMarker == tcell.StyleDefault {
		t.Error("MarkdownListMarker is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownTableBorder == tcell.StyleDefault {
		t.Error("MarkdownTableBorder is not assigned (zero value)")
	}
	if BorlandBlue.MarkdownDefTerm == tcell.StyleDefault {
		t.Error("MarkdownDefTerm is not assigned (zero value)")
	}
}

// TestBorlandBlueMarkdownValuesMatchSpec verifies BorlandBlue markdown style values
// match the specification's BorlandBlue Defaults table for foreground, background,
// and text attributes.
// Falsifying for: "BorlandBlue assigns all 18 fields with the values from the spec's
// BorlandBlue Defaults table" — catches the shortcut where fields are initialized
// but with wrong colors or missing attributes.
func TestBorlandBlueMarkdownValuesMatchSpec(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}

	// MarkdownNormal: LightGray foreground, Blue background, no attributes
	t.Run("MarkdownNormal", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownNormal.Decompose()
		if fg != tcell.ColorLightGray {
			t.Errorf("foreground: got %v, want LightGray", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs != tcell.AttrNone {
			t.Errorf("attributes: got %v, want none", attrs)
		}
	})

	// MarkdownH1: White foreground, Blue background, Bold + Underline
	t.Run("MarkdownH1", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownH1.Decompose()
		if fg != tcell.ColorWhite {
			t.Errorf("foreground: got %v, want White", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs&tcell.AttrBold == 0 {
			t.Error("missing Bold attribute")
		}
		if attrs&tcell.AttrUnderline == 0 {
			t.Error("missing Underline attribute")
		}
		expectedAttrs := tcell.AttrBold | tcell.AttrUnderline
		if attrs & ^expectedAttrs != tcell.AttrNone {
			t.Errorf("unexpected attributes set: got %v, want only Bold|Underline (%v)", attrs, expectedAttrs)
		}
	})

	// MarkdownH2: Yellow foreground, Blue background, Bold
	t.Run("MarkdownH2", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownH2.Decompose()
		if fg != tcell.ColorYellow {
			t.Errorf("foreground: got %v, want Yellow", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs&tcell.AttrBold == 0 {
			t.Error("missing Bold attribute")
		}
		expectedAttrs := tcell.AttrBold
		if attrs & ^expectedAttrs != tcell.AttrNone {
			t.Errorf("unexpected attributes set: got %v, want only Bold (%v)", attrs, expectedAttrs)
		}
	})

	// MarkdownH3: DarkCyan foreground, Blue background, Bold
	t.Run("MarkdownH3", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownH3.Decompose()
		if fg != tcell.ColorDarkCyan {
			t.Errorf("foreground: got %v, want DarkCyan", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs&tcell.AttrBold == 0 {
			t.Error("missing Bold attribute")
		}
		expectedAttrs := tcell.AttrBold
		if attrs & ^expectedAttrs != tcell.AttrNone {
			t.Errorf("unexpected attributes set: got %v, want only Bold (%v)", attrs, expectedAttrs)
		}
	})

	// MarkdownH4: DarkCyan foreground, Blue background, no attributes
	t.Run("MarkdownH4", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownH4.Decompose()
		if fg != tcell.ColorDarkCyan {
			t.Errorf("foreground: got %v, want DarkCyan", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs != tcell.AttrNone {
			t.Errorf("attributes: got %v, want none", attrs)
		}
	})

	// MarkdownH5: LightGray foreground, Blue background, Bold
	t.Run("MarkdownH5", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownH5.Decompose()
		if fg != tcell.ColorLightGray {
			t.Errorf("foreground: got %v, want LightGray", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs&tcell.AttrBold == 0 {
			t.Error("missing Bold attribute")
		}
		expectedAttrs := tcell.AttrBold
		if attrs & ^expectedAttrs != tcell.AttrNone {
			t.Errorf("unexpected attributes set: got %v, want only Bold (%v)", attrs, expectedAttrs)
		}
	})

	// MarkdownH6: LightGray foreground, Blue background, no attributes
	t.Run("MarkdownH6", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownH6.Decompose()
		if fg != tcell.ColorLightGray {
			t.Errorf("foreground: got %v, want LightGray", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs != tcell.AttrNone {
			t.Errorf("attributes: got %v, want none", attrs)
		}
	})

	// MarkdownBold: White foreground, Blue background, Bold
	t.Run("MarkdownBold", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownBold.Decompose()
		if fg != tcell.ColorWhite {
			t.Errorf("foreground: got %v, want White", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs&tcell.AttrBold == 0 {
			t.Error("missing Bold attribute")
		}
		expectedAttrs := tcell.AttrBold
		if attrs & ^expectedAttrs != tcell.AttrNone {
			t.Errorf("unexpected attributes set: got %v, want only Bold (%v)", attrs, expectedAttrs)
		}
	})

	// MarkdownItalic: LightGray foreground, Blue background, Italic
	t.Run("MarkdownItalic", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownItalic.Decompose()
		if fg != tcell.ColorLightGray {
			t.Errorf("foreground: got %v, want LightGray", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs&tcell.AttrItalic == 0 {
			t.Error("missing Italic attribute")
		}
		expectedAttrs := tcell.AttrItalic
		if attrs & ^expectedAttrs != tcell.AttrNone {
			t.Errorf("unexpected attributes set: got %v, want only Italic (%v)", attrs, expectedAttrs)
		}
	})

	// MarkdownBoldItalic: White foreground, Blue background, Bold + Italic
	t.Run("MarkdownBoldItalic", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownBoldItalic.Decompose()
		if fg != tcell.ColorWhite {
			t.Errorf("foreground: got %v, want White", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs&tcell.AttrBold == 0 {
			t.Error("missing Bold attribute")
		}
		if attrs&tcell.AttrItalic == 0 {
			t.Error("missing Italic attribute")
		}
		expectedAttrs := tcell.AttrBold | tcell.AttrItalic
		if attrs & ^expectedAttrs != tcell.AttrNone {
			t.Errorf("unexpected attributes set: got %v, want only Bold|Italic (%v)", attrs, expectedAttrs)
		}
	})

	// MarkdownCode: DarkCyan foreground, Navy background, no attributes
	t.Run("MarkdownCode", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownCode.Decompose()
		if fg != tcell.ColorDarkCyan {
			t.Errorf("foreground: got %v, want DarkCyan", fg)
		}
		if bg != tcell.ColorNavy {
			t.Errorf("background: got %v, want Navy", bg)
		}
		if attrs != tcell.AttrNone {
			t.Errorf("attributes: got %v, want none", attrs)
		}
	})

	// MarkdownCodeBlock: Green foreground, Navy background, no attributes
	t.Run("MarkdownCodeBlock", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownCodeBlock.Decompose()
		if fg != tcell.ColorGreen {
			t.Errorf("foreground: got %v, want Green", fg)
		}
		if bg != tcell.ColorNavy {
			t.Errorf("background: got %v, want Navy", bg)
		}
		if attrs != tcell.AttrNone {
			t.Errorf("attributes: got %v, want none", attrs)
		}
	})

	// MarkdownBlockquote: DarkGray foreground, Blue background, no attributes
	t.Run("MarkdownBlockquote", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownBlockquote.Decompose()
		if fg != tcell.ColorDarkGray {
			t.Errorf("foreground: got %v, want DarkGray", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs != tcell.AttrNone {
			t.Errorf("attributes: got %v, want none", attrs)
		}
	})

	// MarkdownLink: Green foreground, Blue background, Underline
	t.Run("MarkdownLink", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownLink.Decompose()
		if fg != tcell.ColorGreen {
			t.Errorf("foreground: got %v, want Green", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs&tcell.AttrUnderline == 0 {
			t.Error("missing Underline attribute")
		}
		expectedAttrs := tcell.AttrUnderline
		if attrs & ^expectedAttrs != tcell.AttrNone {
			t.Errorf("unexpected attributes set: got %v, want only Underline (%v)", attrs, expectedAttrs)
		}
	})

	// MarkdownHRule: DarkGray foreground, Blue background, no attributes
	t.Run("MarkdownHRule", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownHRule.Decompose()
		if fg != tcell.ColorDarkGray {
			t.Errorf("foreground: got %v, want DarkGray", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs != tcell.AttrNone {
			t.Errorf("attributes: got %v, want none", attrs)
		}
	})

	// MarkdownListMarker: Yellow foreground, Blue background, no attributes
	t.Run("MarkdownListMarker", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownListMarker.Decompose()
		if fg != tcell.ColorYellow {
			t.Errorf("foreground: got %v, want Yellow", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs != tcell.AttrNone {
			t.Errorf("attributes: got %v, want none", attrs)
		}
	})

	// MarkdownTableBorder: DarkGray foreground, Blue background, no attributes
	t.Run("MarkdownTableBorder", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownTableBorder.Decompose()
		if fg != tcell.ColorDarkGray {
			t.Errorf("foreground: got %v, want DarkGray", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs != tcell.AttrNone {
			t.Errorf("attributes: got %v, want none", attrs)
		}
	})

	// MarkdownDefTerm: White foreground, Blue background, Bold
	t.Run("MarkdownDefTerm", func(t *testing.T) {
		fg, bg, attrs := BorlandBlue.MarkdownDefTerm.Decompose()
		if fg != tcell.ColorWhite {
			t.Errorf("foreground: got %v, want White", fg)
		}
		if bg != tcell.ColorBlue {
			t.Errorf("background: got %v, want Blue", bg)
		}
		if attrs&tcell.AttrBold == 0 {
			t.Error("missing Bold attribute")
		}
		expectedAttrs := tcell.AttrBold
		if attrs & ^expectedAttrs != tcell.AttrNone {
			t.Errorf("unexpected attributes set: got %v, want only Bold (%v)", attrs, expectedAttrs)
		}
	})
}
