package main

import (
	"log"

	"github.com/njt/turboview/theme"
	"github.com/njt/turboview/tv"
)

func main() {
	statusLine := tv.NewStatusLine(
		tv.NewStatusItem("~Alt+X~ Exit", tv.KbAlt('X'), tv.CmQuit),
		tv.NewStatusItem("~F2~ Dialog", tv.KbFunc(2), tv.CmUser),
		tv.NewStatusItem("~F10~ Menu", tv.KbFunc(10), tv.CmMenu),
	)

	menuBar := tv.NewMenuBar(
		tv.NewSubMenu("~F~ile",
			tv.NewMenuItem("~N~ew", tv.CmUser+1, tv.KbCtrl('N')),
			tv.NewMenuItem("~O~pen...", tv.CmUser+2, tv.KbCtrl('O')),
			tv.NewMenuSeparator(),
			tv.NewMenuItem("E~x~it", tv.CmQuit, tv.KbAlt('X')),
		),
		tv.NewSubMenu("~W~indow",
			tv.NewMenuItem("~T~ile", tv.CmTile, tv.KbNone()),
			tv.NewMenuItem("~C~ascade", tv.CmCascade, tv.KbNone()),
		),
	)

	var app *tv.Application
	var err error

	app, err = tv.NewApplication(
		tv.WithStatusLine(statusLine),
		tv.WithMenuBar(menuBar),
		tv.WithTheme(theme.BorlandBlue),
		tv.WithOnCommand(func(cmd tv.CommandCode, info any) bool {
			if cmd == tv.CmUser {
				result := tv.MessageBox(app.Desktop(), "Confirm", "Exit the application?", tv.MbYes|tv.MbNo)
				if result == tv.CmYes {
					app.PostCommand(tv.CmQuit, nil)
				}
				return true
			}
			return false
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	win1 := tv.NewWindow(tv.NewRect(5, 2, 35, 15), "File Manager", tv.WithWindowNumber(1))
	st := tv.NewStaticText(tv.NewRect(1, 1, 30, 1), "Press F2 for dialog")
	win1.Insert(st)
	btnOK := tv.NewButton(tv.NewRect(1, 3, 12, 2), "OK", tv.CmOK)
	win1.Insert(btnOK)
	btnClose := tv.NewButton(tv.NewRect(15, 3, 12, 2), "Close", tv.CmClose)
	win1.Insert(btnClose)

	win2 := tv.NewWindow(tv.NewRect(20, 5, 40, 12), "Editor", tv.WithWindowNumber(2))

	app.Desktop().Insert(win1)
	app.Desktop().Insert(win2)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
