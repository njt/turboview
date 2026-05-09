package main

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
	"github.com/njt/turboview/tv"
)

func main() {
	statusLine := tv.NewStatusLine(
		tv.NewStatusItem("~Alt+X~ Exit", tv.KbAlt('X'), tv.CmQuit),
		tv.NewStatusItem("~F2~ Dialog", tv.KbFunc(2), tv.CmUser).ForHelpCtx(1),
		tv.NewStatusItem("~F3~ Input", tv.KbFunc(3), tv.CmUser+10).ForHelpCtx(1),
		tv.NewStatusItem("~F4~ Search", tv.KbFunc(4), tv.CmUser+20).ForHelpCtx(2),
		tv.NewStatusItem("~F10~ Menu", tv.KbFunc(10), tv.CmMenu),
	)

	menuBar := tv.NewMenuBar(
		tv.NewSubMenu("~F~ile",
			tv.NewMenuItem("~N~ew", tv.CmUser+1, tv.KbCtrl('N')),
			tv.NewMenuItem("~O~pen...", tv.CmFileOpen, tv.KbFunc(3)),
			tv.NewMenuSeparator(),
			tv.NewMenuItem("E~x~it", tv.CmQuit, tv.KbAlt('X')),
		),
		tv.NewSubMenu("~E~dit",
			tv.NewMenuItem("~F~ind...", tv.CmFind, tv.KbCtrl('F')),
			tv.NewMenuItem("~R~eplace...", tv.CmReplace, tv.KbCtrl('H')),
			tv.NewMenuItem("~S~earch Again", tv.CmSearchAgain, tv.KbFunc(3)),
		),
		tv.NewSubMenu("~O~ptions",
			tv.NewMenuItem("~C~olors...", tv.CmUser+30, tv.KbNone()),
			tv.NewMenuItem("~M~arkdown Editor", tv.CmUser+40, tv.KbAlt('M')),
		),
		tv.NewSubMenu("~W~indow",
			tv.NewMenuItem("~T~ile", tv.CmTile, tv.KbNone()),
			tv.NewMenuItem("~C~ascade", tv.CmCascade, tv.KbNone()),
		),
	)

	st := tv.NewStaticText(tv.NewRect(1, 1, 30, 1), "Press F2 for dialog")

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
			if cmd == tv.CmUser+10 {
				text, result := tv.InputBox(app.Desktop(), "Open File", "~N~ame:", "untitled.txt")
				if result == tv.CmOK {
					st.SetText("File: " + text)
				}
				return true
			}
			if cmd == tv.CmFileOpen {
				fd := tv.NewFileDialogInDir(".", "*.go", "Open File", tv.FdOpenButton)
				result := app.Desktop().ExecView(fd)
				if result == tv.CmOK {
					fn := fd.FileName()
					if fn != "" {
						ew := tv.NewEditWindow(tv.NewRect(2, 1, 50, 18), fn)
						app.Desktop().Insert(ew)
					}
				}
				return true
			}
			if cmd == tv.CmUser+30 {
				cd := tv.NewColorDialog(nil, nil)
				app.Desktop().ExecView(cd)
				return true
			}
			if cmd == tv.CmUser+40 {
				w := tv.NewWindow(tv.NewRect(2, 1, 60, 20), "Markdown Editor", tv.WithWindowNumber(7))
				iw, ih := w.Bounds().Width()-2, w.Bounds().Height()-2
				editor := tv.NewMarkdownEditor(tv.NewRect(1, 1, iw-1, ih-1))
				// 50+ lines of scroll-test content so the MarkdownEditor
				// viewport overflows and scrollbar behavior is testable.
				mdText := "# Welcome\n\nType **markdown** here.\n\n- item one\n- item two"
				for i := 1; i <= 50; i++ {
					mdText += fmt.Sprintf("\nLine %02d\n", i)
				}
				editor.SetText(mdText)
				vScroll := tv.NewScrollBar(tv.NewRect(iw, 1, 1, ih-1), tv.Vertical)
				hScroll := tv.NewScrollBar(tv.NewRect(1, ih, iw-1, 1), tv.Horizontal)
				editor.SetVScrollBar(vScroll)
				editor.SetHScrollBar(hScroll)
				w.Insert(editor)
				w.Insert(vScroll)
				w.Insert(hScroll)
				app.Desktop().Insert(w)
				return true
			}
			return false
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Window 1 — buttons, checkboxes, radio buttons, input, history, label, validated port
	win1 := tv.NewWindow(tv.NewRect(5, 2, 35, 14), "Controls", tv.WithWindowNumber(1))
	win1.Insert(st)
	btnOK := tv.NewButton(tv.NewRect(1, 3, 12, 2), "OK", tv.CmOK)
	win1.Insert(btnOK)
	btnClose := tv.NewButton(tv.NewRect(15, 3, 12, 2), "Close", tv.CmClose)
	win1.Insert(btnClose)
	checkBoxes := tv.NewCheckBoxes(tv.NewRect(1, 5, 30, 2), []string{"~R~ead only", "~H~idden", "~S~ystem"})
	win1.Insert(checkBoxes)
	radioButtons := tv.NewRadioButtons(tv.NewRect(1, 8, 30, 2), []string{"~T~ext", "~B~inary", "~H~ex"})
	win1.Insert(radioButtons)
	inputLine := tv.NewInputLine(tv.NewRect(11, 10, 20, 1), 40)
	win1.Insert(inputLine)
	history := tv.NewHistory(tv.NewRect(31, 10, 3, 1), inputLine, 1)
	win1.Insert(history)
	nameLabel := tv.NewLabel(tv.NewRect(1, 10, 10, 1), "~N~ame:", inputLine)
	win1.Insert(nameLabel)
	portInput := tv.NewInputLine(tv.NewRect(11, 11, 10, 1), 5)
	portInput.SetValidator(tv.NewRangeValidator(1, 65535))
	portInput.SetText("8080")
	win1.Insert(portInput)
	portLabel := tv.NewLabel(tv.NewRect(1, 11, 10, 1), "~P~ort:", portInput)
	win1.Insert(portLabel)

	// Window 2 — ListBox (ListViewer + ScrollBar)
	win2 := tv.NewWindow(tv.NewRect(20, 5, 40, 12), "List", tv.WithWindowNumber(2))

	listScheme := &theme.ColorScheme{}
	*listScheme = *theme.BorlandBlue
	listScheme.ListNormal = tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)
	listScheme.ListSelected = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorGreen)
	listScheme.ListFocused = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorGreen)
	listScheme.WindowBackground = tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)
	win2.SetColorScheme(listScheme)

	clientW := 40 - 2
	clientH := 12 - 2

	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d", i+1)
	}

	listBox := tv.NewStringListBox(tv.NewRect(0, 0, clientW, clientH), items)
	listBox.ListViewer().SetNumCols(2)
	win2.Insert(listBox)

	// Window 3 — EditWindow (full editor with undo, find/replace)
	win3 := tv.NewEditWindow(tv.NewRect(45, 1, 35, 16), "", tv.WithWindowNumber(3))
	win3.Editor().SetText(`Hello, Editor!
This is the Editor widget.
It supports undo (Ctrl+Z), find (Ctrl+F),
replace (Ctrl+H), and search-again (F3).
Arrow keys navigate. Shift+arrow selects.
Ctrl+A selects all. Ctrl+C/X/V for clipboard.
Home/End for line start/end. Ctrl+Home/End for doc.
PgUp/PgDn scroll. Mouse wheel scrolls too.
Click positions cursor. Double-click selects word.

Line 11: Tab between windows to test focus.
Line 12: Try scrolling past this point.
Line 13: More content below visible area.
Line 14: Horizontal scrolling test — this line extends past the visible width.
Line 15: Almost at the bottom.
Line 16: Last line of demo content.`)
	win3.SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)

	win1.SetHelpCtx(1)
	win2.SetHelpCtx(2)
	win2.SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)
	listBox.SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)

	// Window 4 — Outline (tree viewer with scrollbar)
	goMod := tv.NewNode("go.mod", nil, nil)

	e2eTest := tv.NewNode("e2e_test.go", nil, nil)
	tests := tv.NewNode("tests", e2eTest, goMod) // tests.Next = goMod

	design := tv.NewNode("DESIGN.md", nil, nil)
	readme := tv.NewNode("README.md", nil, design) // README.Next = DESIGN
	docs := tv.NewNode("docs", readme, tests)       // docs.Next = tests

	utilGo := tv.NewNode("util.go", nil, nil)
	mainGo := tv.NewNode("main.go", nil, utilGo) // main.Next = util
	src := tv.NewNode("src", mainGo, docs)         // src.Next = docs

	project := tv.NewNode("Project", src, nil)

	outline := tv.NewOutline(tv.NewRect(1, 1, 28, 13), project)
	vsb := tv.NewScrollBar(tv.NewRect(29, 1, 1, 13), tv.Vertical)
	outline.SetVScrollBar(vsb)

	win4 := tv.NewWindow(tv.NewRect(2, 4, 31, 15), "Outline", tv.WithWindowNumber(4))
	win4.Insert(outline)
	win4.Insert(vsb)

	// Window 5 — Markdown viewer
	mdViewer := tv.NewMarkdownViewer(tv.NewRect(0, 0, 43, 16))
	mdViewer.SetMarkdown(`# MarkdownViewer Demo

This is **bold**, *italic*, ***bold italic***, and ~~strikethrough~~.
Here is ` + "`inline code`" + ` in a sentence.
A [link](https://example.com) and an ![image](photo.jpg).

## Code Block

` + "```go" + `
func main() {
    fmt.Println("Hello, TurboView!")
}
` + "```" + `

## Lists

- First **bullet** item
- Second item with ` + "`code`" + `
- Third item

1. Numbered one
2. Numbered two

- [x] Task complete
- [ ] Task pending

## Blockquote

> This is a blockquote with **bold** text.
> It can span multiple lines.

---

## Table

| Name    | Type   | Description         |
|---------|--------|---------------------|
| width   | int    | Widget width        |
| height  | int    | Widget height       |
| visible | bool   | Visibility flag     |

## Definition List

Markdown
: A lightweight markup language

TurboView
: A TUI framework reimplementing Borland Turbo Vision

### H3 Heading
#### H4 Heading
##### H5 Heading
###### H6 Heading
`)

	mdVSB := tv.NewScrollBar(tv.NewRect(43, 0, 1, 16), tv.Vertical)
	mdViewer.SetVScrollBar(mdVSB)
	mdViewer.SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)
	mdVSB.SetGrowMode(tv.GfGrowLoX | tv.GfGrowHiX | tv.GfGrowHiY)

	win5 := tv.NewWindow(tv.NewRect(3, 3, 45, 18), "Markdown", tv.WithWindowNumber(5))
	win5.SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)
	win5.Insert(mdViewer)
	win5.Insert(mdVSB)

	app.Desktop().Insert(win5)
	app.Desktop().Insert(win4)
	app.Desktop().Insert(win1)
	app.Desktop().Insert(win3)
	app.Desktop().Insert(win2)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
