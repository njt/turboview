package tv

import (
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type MenuBar struct {
	BaseView
	menus       []*SubMenu
	active      bool
	activeIndex int
	popup       *MenuPopup
	app         *Application
	menuXPos    []int
}

func NewMenuBar(menus ...*SubMenu) *MenuBar {
	mb := &MenuBar{
		menus: menus,
	}
	mb.SetState(SfVisible, true)
	return mb
}

func (mb *MenuBar) Popup() *MenuPopup { return mb.popup }
func (mb *MenuBar) IsActive() bool    { return mb.active }
func (mb *MenuBar) Menus() []*SubMenu { return mb.menus }

func (mb *MenuBar) Draw(buf *DrawBuffer) {
	w := mb.Bounds().Width()
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	selectedStyle := tcell.StyleDefault
	if cs := mb.ColorScheme(); cs != nil {
		normalStyle = cs.MenuNormal
		shortcutStyle = cs.MenuShortcut
		selectedStyle = cs.MenuSelected
	}

	buf.Fill(NewRect(0, 0, w, 1), ' ', normalStyle)

	mb.menuXPos = make([]int, len(mb.menus))
	x := 1
	for i, menu := range mb.menus {
		if i > 0 {
			x += 1
		}
		mb.menuXPos[i] = x

		isActive := mb.active && i == mb.activeIndex
		segments := ParseTildeLabel(menu.Label)
		for _, seg := range segments {
			style := normalStyle
			if isActive {
				style = selectedStyle
			} else if seg.Shortcut {
				style = shortcutStyle
			}
			buf.WriteStr(x, 0, seg.Text, style)
			x += utf8.RuneCountInString(seg.Text)
		}
	}
}

func (mb *MenuBar) menuEndX(i int) int {
	if i < 0 || i >= len(mb.menus) {
		return 0
	}
	return mb.menuXPos[i] + tildeTextLen(mb.menus[i].Label)
}

func (mb *MenuBar) menuIndexAtX(x int) int {
	mb.ensureMenuXPos()
	for i := range mb.menus {
		start := mb.menuXPos[i]
		end := mb.menuEndX(i)
		if x >= start && x < end {
			return i
		}
	}
	return -1
}

func (mb *MenuBar) ensureMenuXPos() {
	if len(mb.menuXPos) == len(mb.menus) {
		return
	}
	// Compute positions without a full Draw — used when Draw hasn't been called yet.
	mb.menuXPos = make([]int, len(mb.menus))
	x := 1
	for i, menu := range mb.menus {
		if i > 0 {
			x += 1
		}
		mb.menuXPos[i] = x
		x += tildeTextLen(menu.Label)
	}
}

func (mb *MenuBar) openPopup() {
	if mb.activeIndex < 0 || mb.activeIndex >= len(mb.menus) {
		return
	}
	mb.ensureMenuXPos()
	menu := mb.menus[mb.activeIndex]
	x := mb.menuXPos[mb.activeIndex]
	mb.popup = NewMenuPopup(menu.Items, x, 1)
}

func (mb *MenuBar) closePopup() {
	mb.popup = nil
}

func (mb *MenuBar) ActivateAt(app *Application, index int, openPopup bool) {
	mb.app = app
	mb.active = true
	mb.activeIndex = index
	mb.popup = nil
	if openPopup {
		mb.openPopup()
	}

	app.drawAndFlush()

	for mb.active {
		event := app.PollEvent()
		if event == nil {
			break
		}

		mb.handleModalEvent(event, app)
		app.drawAndFlush()
	}

	mb.active = false
	mb.popup = nil
	mb.app = nil
}

func (mb *MenuBar) Activate(app *Application) {
	mb.ActivateAt(app, 0, false)
}

func (mb *MenuBar) handleModalEvent(event *Event, app *Application) {
	if event.What == EvCommand {
		if event.Command == CmMenu {
			mb.active = false
		} else {
			app.handleCommand(event)
		}
		return
	}

	if event.What == EvKeyboard && event.Key != nil {
		switch event.Key.Key {
		case tcell.KeyEscape:
			if mb.popup != nil {
				mb.closePopup()
			} else {
				mb.active = false
			}
			return

		case tcell.KeyF10:
			mb.active = false
			return

		case tcell.KeyLeft:
			mb.activeIndex = (mb.activeIndex - 1 + len(mb.menus)) % len(mb.menus)
			if mb.popup != nil {
				mb.openPopup()
			}
			return

		case tcell.KeyRight:
			mb.activeIndex = (mb.activeIndex + 1) % len(mb.menus)
			if mb.popup != nil {
				mb.openPopup()
			}
			return

		case tcell.KeyEnter, tcell.KeyDown:
			if mb.popup == nil {
				mb.openPopup()
			} else {
				mb.popup.HandleEvent(event)
				mb.checkPopupResult(app)
			}
			return

		case tcell.KeyUp:
			if mb.popup != nil {
				mb.popup.HandleEvent(event)
				mb.checkPopupResult(app)
			}
			return

		default:
			if mb.popup != nil {
				mb.popup.HandleEvent(event)
				mb.checkPopupResult(app)
			} else if event.Key.Key == tcell.KeyRune {
				mb.matchMenuShortcut(event.Key.Rune)
			}
			return
		}
	}

	if event.What == EvMouse && event.Mouse != nil {
		mx, my := event.Mouse.X, event.Mouse.Y

		if my == 0 && event.Mouse.Button&tcell.Button1 != 0 {
			idx := mb.menuIndexAtX(mx)
			if idx >= 0 {
				mb.activeIndex = idx
				mb.openPopup()
			}
			return
		}

		if mb.popup != nil {
			pb := mb.popup.Bounds()
			if pb.Contains(NewPoint(mx, my)) {
				localEvent := *event
				localMouse := *event.Mouse
				localMouse.X = mx - pb.A.X
				localMouse.Y = my - pb.A.Y
				localEvent.Mouse = &localMouse
				mb.popup.HandleEvent(&localEvent)
				mb.checkPopupResult(app)
				return
			}
		}

		if event.Mouse.Button&tcell.Button1 != 0 {
			mb.active = false
		}
		return
	}
}

func (mb *MenuBar) checkPopupResult(app *Application) {
	if mb.popup == nil {
		return
	}
	result := mb.popup.Result()
	if result == 0 {
		return
	}
	if result == CmCancel {
		mb.closePopup()
		return
	}
	mb.closePopup()
	mb.active = false
	// Dispatch the command directly so it is handled even though the modal loop exits.
	cmdEvent := &Event{What: EvCommand, Command: result}
	app.handleCommand(cmdEvent)
}

func (mb *MenuBar) matchMenuShortcut(r rune) {
	r = unicode.ToLower(r)
	for i, menu := range mb.menus {
		segments := ParseTildeLabel(menu.Label)
		for _, seg := range segments {
			if seg.Shortcut && len(seg.Text) > 0 {
				sc, _ := utf8.DecodeRuneInString(seg.Text)
				if unicode.ToLower(sc) == r {
					mb.activeIndex = i
					mb.openPopup()
					return
				}
			}
		}
	}
}

func (mb *MenuBar) HandleEvent(event *Event) {
	if event.What == EvCommand && event.Command == CmMenu {
		if mb.app != nil {
			mb.Activate(mb.app)
		}
		event.Clear()
		return
	}
}
