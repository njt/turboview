package tv

import (
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

type cmdTcellEvent struct {
	tcell.EventTime
	cmd  CommandCode
	info any
}

type mouseAutoEvent struct {
	tcell.EventTime
	x, y   int
	button tcell.ButtonMask
}

type AppOption func(*Application)

func WithScreen(s tcell.Screen) AppOption {
	return func(app *Application) {
		app.screen = s
		app.screenOwn = false
	}
}

func WithStatusLine(sl *StatusLine) AppOption {
	return func(app *Application) {
		app.statusLine = sl
	}
}

func WithMenuBar(mb *MenuBar) AppOption {
	return func(app *Application) {
		app.menuBar = mb
	}
}

func WithTheme(scheme *theme.ColorScheme) AppOption {
	return func(app *Application) {
		app.scheme = scheme
	}
}

func WithOnCommand(fn func(CommandCode, any) bool) AppOption {
	return func(app *Application) {
		app.onCommand = fn
	}
}

func WithConfigFile(path string) AppOption {
	return func(app *Application) {
		app.configFile = path
	}
}

type Application struct {
	bounds       Rect
	screen       tcell.Screen
	screenOwn    bool
	desktop      *Desktop
	statusLine   *StatusLine
	menuBar      *MenuBar
	scheme       *theme.ColorScheme
	configFile   string
	quit         bool
	onCommand    func(CommandCode, any) bool
	contextPopup *MenuPopup

	mouseAutoMu   sync.Mutex
	mouseAutoBtn  tcell.ButtonMask
	mouseAutoX    int
	mouseAutoY    int
	mouseAutoChan chan struct{}
}

func NewApplication(opts ...AppOption) (*Application, error) {
	app := &Application{
		screenOwn: true,
	}
	for _, opt := range opts {
		opt(app)
	}

	if app.screen == nil {
		s, err := tcell.NewScreen()
		if err != nil {
			return nil, err
		}
		app.screen = s
	}

	if app.scheme == nil {
		app.scheme = theme.BorlandBlue
	}

	configPath := app.configFile
	if configPath == "" {
		configPath = theme.DefaultConfigPath()
	}
	if configPath != "" {
		cs, err := theme.LoadConfig(configPath)
		if err != nil {
			return nil, err
		}
		if cs != nil {
			app.scheme = cs
		}
	}

	app.desktop = NewDesktop(NewRect(0, 0, 0, 0))
	app.desktop.scheme = app.scheme
	app.desktop.app = app

	if app.statusLine != nil {
		app.statusLine.scheme = app.scheme
	}

	if app.menuBar != nil {
		app.menuBar.scheme = app.scheme
		app.menuBar.app = app
	}

	// Set initial bounds from screen size so Draw works without Run.
	w, h := app.screen.Size()
	app.bounds = NewRect(0, 0, w, h)
	app.layoutChildren()

	return app, nil
}

func (app *Application) Desktop() *Desktop      { return app.desktop }
func (app *Application) StatusLine() *StatusLine { return app.statusLine }
func (app *Application) MenuBar() *MenuBar       { return app.menuBar }
func (app *Application) Screen() tcell.Screen    { return app.screen }

func (app *Application) PollEvent() *Event {
	for {
		tcellEv := app.screen.PollEvent()
		if tcellEv == nil {
			return nil
		}
		if resizeEv, ok := tcellEv.(*tcell.EventResize); ok {
			w, h := resizeEv.Size()
			app.bounds = NewRect(0, 0, w, h)
			app.layoutChildren()
		}
		if event := app.convertEvent(tcellEv); event != nil {
			return event
		}
	}
}

func (app *Application) Run() error {
	defer app.stopMouseAuto()
	if app.screenOwn {
		if err := app.screen.Init(); err != nil {
			return err
		}
		defer app.screen.Fini()
	}

	app.screen.EnableMouse()
	app.screen.Clear()

	w, h := app.screen.Size()
	app.bounds = NewRect(0, 0, w, h)
	app.layoutChildren()
	app.drawAndFlush()

	for !app.quit {
		event := app.PollEvent()
		if event == nil {
			break
		}
		app.handleEvent(event)
		app.drawAndFlush()
	}

	return nil
}

func (app *Application) PostCommand(cmd CommandCode, info any) {
	ev := &cmdTcellEvent{cmd: cmd, info: info}
	ev.SetEventNow()
	app.screen.PostEvent(ev)
}

func (app *Application) startMouseAuto(x, y int, button tcell.ButtonMask) {
	app.stopMouseAuto()
	app.mouseAutoMu.Lock()
	app.mouseAutoBtn = button
	app.mouseAutoX = x
	app.mouseAutoY = y
	done := make(chan struct{})
	app.mouseAutoChan = done
	app.mouseAutoMu.Unlock()
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				app.mouseAutoMu.Lock()
				ev := &mouseAutoEvent{x: app.mouseAutoX, y: app.mouseAutoY, button: app.mouseAutoBtn}
				app.mouseAutoMu.Unlock()
				ev.SetEventNow()
				_ = app.screen.PostEvent(ev)
			}
		}
	}()
}

func (app *Application) stopMouseAuto() {
	app.mouseAutoMu.Lock()
	defer app.mouseAutoMu.Unlock()
	if app.mouseAutoChan != nil {
		close(app.mouseAutoChan)
		app.mouseAutoChan = nil
	}
	app.mouseAutoBtn = 0
}

func (app *Application) Draw(buf *DrawBuffer) {
	h := app.bounds.Height()
	w := app.bounds.Width()

	menuH := 0
	if app.menuBar != nil {
		menuH = 1
	}

	desktopBottom := h
	if app.statusLine != nil {
		desktopBottom = h - 1
	}
	desktopH := desktopBottom - menuH

	if app.desktop != nil && desktopH > 0 {
		desktopBuf := buf.SubBuffer(NewRect(0, menuH, w, desktopH))
		app.desktop.Draw(desktopBuf)
	}

	if app.menuBar != nil {
		menuBuf := buf.SubBuffer(NewRect(0, 0, w, 1))
		app.menuBar.Draw(menuBuf)

		if popup := app.menuBar.Popup(); popup != nil {
			popup.Draw(buf.SubBuffer(popup.Bounds()), app.scheme)
		}
	}

	if app.statusLine != nil && h > 0 {
		app.statusLine.SetActiveContext(app.resolveHelpCtx())
		statusBuf := buf.SubBuffer(NewRect(0, h-1, w, 1))
		app.statusLine.Draw(statusBuf)
	}

	if app.contextPopup != nil {
		pb := app.contextPopup.Bounds()
		app.contextPopup.Draw(buf.SubBuffer(pb), app.scheme)
	}
}

func (app *Application) ContextMenu(x, y int, items ...any) CommandCode {
	popup := NewMenuPopup(items, x, y)
	app.contextPopup = popup
	app.drawAndFlush()

	var result CommandCode
	for result == 0 {
		event := app.PollEvent()
		if event == nil {
			result = CmCancel
			break
		}

		if event.What == EvKeyboard && event.Key != nil {
			popup.HandleEvent(event)
			if r := popup.Result(); r != 0 {
				result = r
				break
			}
		} else if event.What == EvMouse && event.Mouse != nil {
			pb := popup.Bounds()
			mx, my := event.Mouse.X, event.Mouse.Y
			if pb.Contains(NewPoint(mx, my)) {
				localEvent := *event
				localMouse := *event.Mouse
				localMouse.X = mx - pb.A.X
				localMouse.Y = my - pb.A.Y
				localEvent.Mouse = &localMouse
				popup.HandleEvent(&localEvent)
				if r := popup.Result(); r != 0 {
					result = r
					break
				}
			} else if event.Mouse.Button&tcell.Button1 != 0 {
				result = CmCancel
				break
			}
		}

		app.drawAndFlush()
	}

	app.contextPopup = nil
	app.drawAndFlush()

	if result == CmCancel {
		return CmCancel
	}
	return result
}

func (app *Application) resolveHelpCtx() HelpContext {
	if app.desktop == nil {
		return HcNoContext
	}
	type helpCtxer interface {
		HelpCtx() HelpContext
	}
	ctx := HcNoContext
	var current View = app.desktop
	for {
		if h, ok := current.(helpCtxer); ok {
			if hc := h.HelpCtx(); hc != HcNoContext {
				ctx = hc
			}
		}
		c, ok := current.(Container)
		if !ok {
			break
		}
		focused := c.FocusedChild()
		if focused == nil {
			break
		}
		current = focused
	}
	return ctx
}

func (app *Application) layoutChildren() {
	w, h := app.bounds.Width(), app.bounds.Height()

	menuH := 0
	if app.menuBar != nil {
		menuH = 1
		app.menuBar.SetBounds(NewRect(0, 0, w, 1))
	}

	desktopBottom := h
	if app.statusLine != nil {
		statusRow := h - 1
		if statusRow < 0 {
			statusRow = 0
		}
		app.statusLine.SetBounds(NewRect(0, statusRow, w, 1))
		desktopBottom = statusRow
	}

	if app.desktop != nil {
		desktopH := desktopBottom - menuH
		if desktopH < 0 {
			desktopH = 0
		}
		app.desktop.SetBounds(NewRect(0, menuH, w, desktopH))
	}
}

func (app *Application) handleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		app.routeMouseEvent(event)
		if !event.IsCleared() && event.What == EvCommand {
			app.handleCommand(event)
		}
		return
	}

	if app.statusLine != nil {
		app.statusLine.HandleEvent(event)
	}

	if !event.IsCleared() && event.What == EvCommand && event.Command == CmMenu && app.menuBar != nil {
		app.menuBar.Activate(app)
		event.Clear()
		return
	}

	if !event.IsCleared() && app.desktop != nil {
		app.desktop.HandleEvent(event)
	}

	if !event.IsCleared() {
		app.handleCommand(event)
	}
}

func (app *Application) handleCommand(event *Event) {
	if event.What == EvCommand {
		switch event.Command {
		case CmTile:
			if app.desktop != nil {
				app.desktop.Tile()
			}
			event.Clear()
			return
		case CmCascade:
			if app.desktop != nil {
				app.desktop.Cascade()
			}
			event.Clear()
			return
		case CmQuit:
			app.quit = true
			event.Clear()
			return
		}
		if app.onCommand != nil {
			if app.onCommand(event.Command, event.Info) {
				event.Clear()
			}
		}
	}
}

func (app *Application) routeMouseEvent(event *Event) {
	mx, my := event.Mouse.X, event.Mouse.Y

	if app.menuBar != nil && my == 0 {
		idx := app.menuBar.menuIndexAtX(mx)
		if idx >= 0 {
			app.stopMouseAuto()
			app.menuBar.ActivateAt(app, idx, true)
		}
		return
	}

	if app.statusLine != nil {
		slBounds := app.statusLine.Bounds()
		if slBounds.Contains(NewPoint(mx, my)) {
			event.Mouse.X -= slBounds.A.X
			event.Mouse.Y -= slBounds.A.Y
			app.statusLine.HandleEvent(event)
			return
		}
	}

	if app.desktop != nil {
		dBounds := app.desktop.Bounds()
		if dBounds.Contains(NewPoint(mx, my)) {
			event.Mouse.X -= dBounds.A.X
			event.Mouse.Y -= dBounds.A.Y
			app.desktop.HandleEvent(event)
		}
	}
}

func (app *Application) convertEvent(tcellEv tcell.Event) *Event {
	switch ev := tcellEv.(type) {
	case *tcell.EventKey:
		return &Event{
			What: EvKeyboard,
			Key: &KeyEvent{
				Key:       ev.Key(),
				Rune:      ev.Rune(),
				Modifiers: ev.Modifiers(),
			},
		}
	case *tcell.EventMouse:
		x, y := ev.Position()
		buttons := ev.Buttons()
		realButtons := buttons & (tcell.Button1 | tcell.Button2 | tcell.Button3)
		if realButtons != 0 && app.mouseAutoBtn == 0 {
			app.startMouseAuto(x, y, realButtons)
		} else if realButtons != 0 {
			app.mouseAutoMu.Lock()
			app.mouseAutoX = x
			app.mouseAutoY = y
			app.mouseAutoMu.Unlock()
		} else if realButtons == 0 && app.mouseAutoBtn != 0 {
			app.stopMouseAuto()
		}
		return &Event{
			What: EvMouse,
			Mouse: &MouseEvent{
				X:         x,
				Y:         y,
				Button:    buttons,
				Modifiers: ev.Modifiers(),
			},
		}
	case *mouseAutoEvent:
		return &Event{
			What: EvMouse,
			Mouse: &MouseEvent{
				X:      ev.x,
				Y:      ev.y,
				Button: ev.button,
			},
		}
	case *tcell.EventResize:
		return &Event{
			What:    EvCommand,
			Command: CmResize,
		}
	case *cmdTcellEvent:
		return &Event{
			What:    EvCommand,
			Command: ev.cmd,
			Info:    ev.info,
		}
	}
	return nil
}

func (app *Application) drawAndFlush() {
	w, h := app.screen.Size()
	if w <= 0 || h <= 0 {
		return
	}
	buf := NewDrawBuffer(w, h)
	app.Draw(buf)
	buf.FlushTo(app.screen)
	app.screen.Show()
}
