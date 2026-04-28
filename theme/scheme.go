package theme

import "github.com/gdamore/tcell/v2"

type ColorScheme struct {
	WindowBackground    tcell.Style
	WindowFrameActive   tcell.Style
	WindowFrameInactive tcell.Style
	WindowTitle         tcell.Style
	WindowShadow        tcell.Style
	DesktopBackground   tcell.Style
	DialogBackground    tcell.Style
	DialogFrame         tcell.Style
	ButtonNormal        tcell.Style
	ButtonDefault       tcell.Style
	ButtonShadow        tcell.Style
	ButtonShortcut      tcell.Style
	InputNormal         tcell.Style
	InputSelection      tcell.Style
	LabelNormal         tcell.Style
	LabelShortcut       tcell.Style
	CheckBoxNormal      tcell.Style
	RadioButtonNormal   tcell.Style
	ListNormal          tcell.Style
	ListSelected        tcell.Style
	ListFocused         tcell.Style
	ScrollBar           tcell.Style
	ScrollThumb         tcell.Style
	MenuNormal          tcell.Style
	MenuShortcut        tcell.Style
	MenuSelected        tcell.Style
	MenuDisabled        tcell.Style
	StatusNormal        tcell.Style
	StatusShortcut      tcell.Style
}
