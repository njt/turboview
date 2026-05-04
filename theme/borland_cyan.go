package theme

import "github.com/gdamore/tcell/v2"

var BorlandCyan *ColorScheme

func init() {
	s := func(fg, bg tcell.Color) tcell.Style {
		return tcell.StyleDefault.Foreground(fg).Background(bg)
	}

	BorlandCyan = &ColorScheme{
		WindowBackground:    s(tcell.ColorWhite, tcell.ColorTeal),
		WindowFrameActive:   s(tcell.ColorYellow, tcell.ColorTeal),
		WindowFrameInactive: s(tcell.ColorSilver, tcell.ColorTeal),
		WindowTitle:         s(tcell.ColorWhite, tcell.ColorTeal),
		WindowShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		DesktopBackground:   s(tcell.ColorTeal, tcell.ColorDarkCyan),
		DialogBackground:    s(tcell.ColorBlack, tcell.ColorSilver),
		DialogFrame:         s(tcell.ColorWhite, tcell.ColorSilver),
		ButtonNormal:        s(tcell.ColorBlack, tcell.ColorGreen),
		ButtonDefault:       s(tcell.ColorWhite, tcell.ColorGreen),
		ButtonShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		ButtonShortcut:      s(tcell.ColorYellow, tcell.ColorGreen),
		InputNormal:         s(tcell.ColorBlack, tcell.ColorWhite),
		InputSelection:      s(tcell.ColorWhite, tcell.ColorBlue),
		LabelNormal:         s(tcell.ColorBlack, tcell.ColorSilver),
		LabelHighlight:      s(tcell.ColorWhite, tcell.ColorBlue),
		LabelShortcut:       s(tcell.ColorYellow, tcell.ColorSilver),
		CheckBoxNormal:      s(tcell.ColorBlack, tcell.ColorSilver),
		CheckBoxSelected:    s(tcell.ColorYellow, tcell.ColorSilver),
		RadioButtonNormal:   s(tcell.ColorBlack, tcell.ColorSilver),
		RadioButtonSelected: s(tcell.ColorYellow, tcell.ColorSilver),
		ClusterDisabled:     s(tcell.ColorDarkGray, tcell.ColorSilver),
		ListNormal:          s(tcell.ColorBlack, tcell.ColorSilver),
		ListSelected:        s(tcell.ColorWhite, tcell.ColorBlack),
		ListFocused:         s(tcell.ColorYellow, tcell.ColorTeal),
		ScrollBar:           s(tcell.ColorSilver, tcell.ColorTeal),
		ScrollThumb:         s(tcell.ColorWhite, tcell.ColorTeal),
		MenuNormal:          s(tcell.ColorBlack, tcell.ColorSilver),
		MenuShortcut:        s(tcell.ColorRed, tcell.ColorSilver),
		MenuSelected:        s(tcell.ColorWhite, tcell.ColorBlack),
		MenuDisabled:        s(tcell.ColorGray, tcell.ColorSilver),
		StatusNormal:        s(tcell.ColorBlack, tcell.ColorSilver),
		StatusShortcut:      s(tcell.ColorYellow, tcell.ColorSilver),
		StatusSelected:      s(tcell.ColorBlack, tcell.ColorWhite),
		MemoNormal:          s(tcell.ColorBlack, tcell.ColorWhite),
		MemoSelected:        s(tcell.ColorWhite, tcell.ColorBlue),
		HistoryArrow:        s(tcell.ColorDarkCyan, tcell.ColorWhite),
		HistorySides:        s(tcell.ColorWhite, tcell.ColorTeal),
		ColorSelectorNormal: s(tcell.ColorLightGray, tcell.ColorBlack),
		ColorSelectorCursor: s(tcell.ColorYellow, tcell.ColorBlack),
		ColorDisplayText:    s(tcell.ColorWhite, tcell.ColorBlack),
	}

	Register("borland-cyan", BorlandCyan)
}
