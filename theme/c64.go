package theme

import "github.com/gdamore/tcell/v2"

var C64 *ColorScheme

func init() {
	s := func(fg, bg tcell.Color) tcell.Style {
		return tcell.StyleDefault.Foreground(fg).Background(bg)
	}

	C64 = &ColorScheme{
		WindowBackground:    s(tcell.ColorWhite, tcell.ColorNavy),
		WindowFrameActive:   s(tcell.ColorYellow, tcell.ColorNavy),
		WindowFrameInactive: s(tcell.ColorSilver, tcell.ColorNavy),
		WindowTitle:         s(tcell.ColorYellow, tcell.ColorNavy),
		WindowShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		DesktopBackground:   s(tcell.ColorBlue, tcell.ColorNavy),
		DialogBackground:    s(tcell.ColorWhite, tcell.ColorPurple),
		DialogFrame:         s(tcell.ColorYellow, tcell.ColorPurple),
		ButtonNormal:        s(tcell.ColorWhite, tcell.ColorRed),
		ButtonDefault:       s(tcell.ColorYellow, tcell.ColorRed),
		ButtonShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		ButtonShortcut:      s(tcell.ColorYellow, tcell.ColorRed),
		InputNormal:         s(tcell.ColorBlue, tcell.ColorWhite),
		InputSelection:      s(tcell.ColorWhite, tcell.ColorBlue),
		LabelNormal:         s(tcell.ColorWhite, tcell.ColorPurple),
		LabelShortcut:       s(tcell.ColorYellow, tcell.ColorPurple),
		CheckBoxNormal:      s(tcell.ColorWhite, tcell.ColorPurple),
		CheckBoxSelected:    s(tcell.ColorYellow, tcell.ColorPurple),
		RadioButtonNormal:   s(tcell.ColorWhite, tcell.ColorPurple),
		RadioButtonSelected: s(tcell.ColorYellow, tcell.ColorPurple),
		ListNormal:          s(tcell.ColorTeal, tcell.ColorNavy),
		ListSelected:        s(tcell.ColorWhite, tcell.ColorBlue),
		ListFocused:         s(tcell.ColorYellow, tcell.ColorBlue),
		ScrollBar:           s(tcell.ColorBlue, tcell.ColorNavy),
		ScrollThumb:         s(tcell.ColorTeal, tcell.ColorNavy),
		MenuNormal:          s(tcell.ColorTeal, tcell.ColorNavy),
		MenuShortcut:        s(tcell.ColorYellow, tcell.ColorNavy),
		MenuSelected:        s(tcell.ColorNavy, tcell.ColorTeal),
		MenuDisabled:        s(tcell.ColorGray, tcell.ColorNavy),
		StatusNormal:        s(tcell.ColorWhite, tcell.ColorMaroon),
		StatusShortcut:      s(tcell.ColorYellow, tcell.ColorMaroon),
		MemoNormal:          s(tcell.ColorWhite, tcell.ColorNavy),
		MemoSelected:        s(tcell.ColorWhite, tcell.ColorBlue),
		HistoryArrow:        s(tcell.ColorLightBlue, tcell.ColorBlue),
		HistorySides:        s(tcell.ColorBlue, tcell.ColorLightBlue),
	}

	Register("c64", C64)
}
