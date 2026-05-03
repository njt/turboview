package theme

import "github.com/gdamore/tcell/v2"

var Matrix *ColorScheme

func init() {
	s := func(fg, bg tcell.Color) tcell.Style {
		return tcell.StyleDefault.Foreground(fg).Background(bg)
	}

	Matrix = &ColorScheme{
		WindowBackground:    s(tcell.ColorGreen, tcell.ColorBlack),
		WindowFrameActive:   s(tcell.ColorLime, tcell.ColorBlack),
		WindowFrameInactive: s(tcell.ColorDarkGreen, tcell.ColorBlack),
		WindowTitle:         s(tcell.ColorLime, tcell.ColorBlack),
		WindowShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		DesktopBackground:   s(tcell.ColorDarkGreen, tcell.ColorBlack),
		DialogBackground:    s(tcell.ColorGreen, tcell.ColorBlack),
		DialogFrame:         s(tcell.ColorLime, tcell.ColorBlack),
		ButtonNormal:        s(tcell.ColorBlack, tcell.ColorGreen),
		ButtonDefault:       s(tcell.ColorBlack, tcell.ColorLime),
		ButtonShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		ButtonShortcut:      s(tcell.ColorLime, tcell.ColorGreen),
		InputNormal:         s(tcell.ColorGreen, tcell.ColorBlack),
		InputSelection:      s(tcell.ColorBlack, tcell.ColorGreen),
		LabelNormal:         s(tcell.ColorGreen, tcell.ColorBlack),
		LabelHighlight:      s(tcell.ColorLime, tcell.ColorBlack),
		LabelShortcut:       s(tcell.ColorLime, tcell.ColorBlack),
		CheckBoxNormal:      s(tcell.ColorGreen, tcell.ColorBlack),
		CheckBoxSelected:    s(tcell.ColorWhite, tcell.ColorBlack),
		RadioButtonNormal:   s(tcell.ColorGreen, tcell.ColorBlack),
		RadioButtonSelected: s(tcell.ColorWhite, tcell.ColorBlack),
		ClusterDisabled:     s(tcell.ColorDarkGreen, tcell.ColorBlack),
		ListNormal:          s(tcell.ColorGreen, tcell.ColorBlack),
		ListSelected:        s(tcell.ColorBlack, tcell.ColorGreen),
		ListFocused:         s(tcell.ColorBlack, tcell.ColorLime),
		ScrollBar:           s(tcell.ColorDarkGreen, tcell.ColorBlack),
		ScrollThumb:         s(tcell.ColorGreen, tcell.ColorBlack),
		MenuNormal:          s(tcell.ColorGreen, tcell.ColorBlack),
		MenuShortcut:        s(tcell.ColorLime, tcell.ColorBlack),
		MenuSelected:        s(tcell.ColorBlack, tcell.ColorGreen),
		MenuDisabled:        s(tcell.ColorDarkGreen, tcell.ColorBlack),
		StatusNormal:        s(tcell.ColorBlack, tcell.ColorGreen),
		StatusShortcut:      s(tcell.ColorLime, tcell.ColorGreen),
		StatusSelected:      s(tcell.ColorBlack, tcell.ColorGreen),
		MemoNormal:          s(tcell.ColorGreen, tcell.ColorBlack),
		MemoSelected:        s(tcell.ColorBlack, tcell.ColorGreen),
		HistoryArrow:        s(tcell.ColorWhite, tcell.ColorDarkGreen),
		HistorySides:        s(tcell.ColorGreen, tcell.ColorBlack),
	}

	Register("matrix", Matrix)
}
