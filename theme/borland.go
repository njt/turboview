package theme

import "github.com/gdamore/tcell/v2"

var BorlandBlue *ColorScheme

func init() {
	s := func(fg, bg tcell.Color) tcell.Style {
		return tcell.StyleDefault.Foreground(fg).Background(bg)
	}

	BorlandBlue = &ColorScheme{
		WindowBackground:    s(tcell.ColorYellow, tcell.ColorBlue),
		WindowFrameActive:   s(tcell.ColorWhite, tcell.ColorBlue),
		WindowFrameInactive: s(tcell.ColorSilver, tcell.ColorBlue),
		WindowTitle:         s(tcell.ColorWhite, tcell.ColorBlue),
		WindowShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		DesktopBackground:   s(tcell.ColorTeal, tcell.ColorBlue),
		DialogBackground:    s(tcell.ColorBlack, tcell.ColorTeal),
		DialogFrame:         s(tcell.ColorWhite, tcell.ColorTeal),
		ButtonNormal:        s(tcell.ColorBlack, tcell.ColorGreen),
		ButtonDefault:       s(tcell.ColorWhite, tcell.ColorGreen),
		ButtonShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		ButtonShortcut:      s(tcell.ColorYellow, tcell.ColorGreen),
		InputNormal:         s(tcell.ColorWhite, tcell.ColorBlue),
		InputSelection:      s(tcell.ColorBlue, tcell.ColorTeal),
		LabelNormal:         s(tcell.ColorBlack, tcell.ColorTeal),
		LabelHighlight:      s(tcell.ColorWhite, tcell.ColorBlue),
		LabelShortcut:       s(tcell.ColorYellow, tcell.ColorTeal),
		CheckBoxNormal:      s(tcell.ColorBlack, tcell.ColorTeal),
		CheckBoxSelected:    s(tcell.ColorYellow, tcell.ColorTeal),
		RadioButtonNormal:   s(tcell.ColorBlack, tcell.ColorTeal),
		RadioButtonSelected: s(tcell.ColorYellow, tcell.ColorTeal),
		ListNormal:          s(tcell.ColorBlack, tcell.ColorTeal),
		ListSelected:        s(tcell.ColorWhite, tcell.ColorBlack),
		ListFocused:         s(tcell.ColorYellow, tcell.ColorBlue),
		ScrollBar:           s(tcell.ColorTeal, tcell.ColorBlue),
		ScrollThumb:         s(tcell.ColorWhite, tcell.ColorBlue),
		MenuNormal:          s(tcell.ColorBlack, tcell.ColorTeal),
		MenuShortcut:        s(tcell.ColorYellow, tcell.ColorTeal),
		MenuSelected:        s(tcell.ColorWhite, tcell.ColorBlack),
		MenuDisabled:        s(tcell.ColorSilver, tcell.ColorTeal),
		StatusNormal:        s(tcell.ColorBlack, tcell.ColorTeal),
		StatusShortcut:      s(tcell.ColorYellow, tcell.ColorTeal),
		StatusSelected:      s(tcell.ColorBlack, tcell.ColorWhite),
		MemoNormal:          s(tcell.ColorYellow, tcell.ColorBlue),
		MemoSelected:        s(tcell.ColorWhite, tcell.ColorGreen),
	}

	Register("borland-blue", BorlandBlue)
}
