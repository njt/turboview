package theme

import "github.com/gdamore/tcell/v2"

var BorlandGray *ColorScheme

func init() {
	s := func(fg, bg tcell.Color) tcell.Style {
		return tcell.StyleDefault.Foreground(fg).Background(bg)
	}

	BorlandGray = &ColorScheme{
		WindowBackground:    s(tcell.ColorWhite, tcell.ColorDarkGray),
		WindowFrameActive:   s(tcell.ColorWhite, tcell.ColorDarkGray),
		WindowFrameInactive: s(tcell.ColorSilver, tcell.ColorDarkGray),
		WindowTitle:         s(tcell.ColorWhite, tcell.ColorDarkGray),
		WindowShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		DesktopBackground:   s(tcell.ColorSilver, tcell.ColorGray),
		DialogBackground:    s(tcell.ColorBlack, tcell.ColorSilver),
		DialogFrame:         s(tcell.ColorWhite, tcell.ColorSilver),
		ButtonNormal:        s(tcell.ColorBlack, tcell.ColorSilver),
		ButtonDefault:       s(tcell.ColorWhite, tcell.ColorDarkGray),
		ButtonShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		ButtonShortcut:      s(tcell.ColorYellow, tcell.ColorSilver),
		InputNormal:         s(tcell.ColorBlack, tcell.ColorWhite),
		InputSelection:      s(tcell.ColorWhite, tcell.ColorDarkGray),
		LabelNormal:         s(tcell.ColorBlack, tcell.ColorSilver),
		LabelShortcut:       s(tcell.ColorYellow, tcell.ColorSilver),
		CheckBoxNormal:      s(tcell.ColorBlack, tcell.ColorSilver),
		RadioButtonNormal:   s(tcell.ColorBlack, tcell.ColorSilver),
		ListNormal:          s(tcell.ColorBlack, tcell.ColorSilver),
		ListSelected:        s(tcell.ColorWhite, tcell.ColorBlack),
		ListFocused:         s(tcell.ColorYellow, tcell.ColorDarkGray),
		ScrollBar:           s(tcell.ColorSilver, tcell.ColorDarkGray),
		ScrollThumb:         s(tcell.ColorWhite, tcell.ColorDarkGray),
		MenuNormal:          s(tcell.ColorBlack, tcell.ColorSilver),
		MenuShortcut:        s(tcell.ColorRed, tcell.ColorSilver),
		MenuSelected:        s(tcell.ColorWhite, tcell.ColorBlack),
		MenuDisabled:        s(tcell.ColorGray, tcell.ColorSilver),
		StatusNormal:        s(tcell.ColorBlack, tcell.ColorSilver),
		StatusShortcut:      s(tcell.ColorYellow, tcell.ColorSilver),
	}

	Register("borland-gray", BorlandGray)
}
