package main

import (
	"log"

	"github.com/njt/turboview/theme"
	"github.com/njt/turboview/tv"
)

func main() {
	statusLine := tv.NewStatusLine(
		tv.NewStatusItem("~Alt+X~ Exit", tv.KbAlt('X'), tv.CmQuit),
		tv.NewStatusItem("~F10~ Menu", tv.KbFunc(10), tv.CmMenu),
	)

	app, err := tv.NewApplication(
		tv.WithStatusLine(statusLine),
		tv.WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		log.Fatal(err)
	}

	win1 := tv.NewWindow(tv.NewRect(5, 2, 35, 15), "File Manager", tv.WithWindowNumber(1))
	win2 := tv.NewWindow(tv.NewRect(20, 5, 40, 12), "Editor", tv.WithWindowNumber(2))

	app.Desktop().Insert(win1)
	app.Desktop().Insert(win2)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
