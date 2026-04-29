package tv

import (
	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

type cmdTcellEvent struct {
	tcell.EventTime
	cmd  CommandCode
	info any
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

type Application struct {
	bounds     Rect
	screen     tcell.Screen
	screenOwn  bool
	desktop    *Desktop
	statusLine *StatusLine
	scheme     *theme.ColorScheme
	quit       bool
	onCommand  func(CommandCode, any) bool
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

	app.desktop = NewDesktop(NewRect(0, 0, 0, 0))
	app.desktop.scheme = app.scheme
	app.desktop.app = app

	if app.statusLine != nil {
		app.statusLine.scheme = app.scheme
	}

	// Set initial bounds from screen size so Draw works without Run.
	w, h := app.screen.Size()
	app.bounds = NewRect(0, 0, w, h)
	app.layoutChildren()

	return app, nil
}

func (app *Application) Desktop() *Desktop      { return app.desktop }
func (app *Application) StatusLine() *StatusLine { return app.statusLine }
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

func (app *Application) Draw(buf *DrawBuffer) {
	h := app.bounds.Height()
	w := app.bounds.Width()

	desktopBottom := h
	if app.statusLine != nil {
		desktopBottom = h - 1
	}

	if app.desktop != nil && desktopBottom > 0 {
		desktopBuf := buf.SubBuffer(NewRect(0, 0, w, desktopBottom))
		app.desktop.Draw(desktopBuf)
	}

	if app.statusLine != nil && h > 0 {
		statusBuf := buf.SubBuffer(NewRect(0, h-1, w, 1))
		app.statusLine.Draw(statusBuf)
	}
}

func (app *Application) layoutChildren() {
	w, h := app.bounds.Width(), app.bounds.Height()
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
		if desktopBottom < 0 {
			desktopBottom = 0
		}
		app.desktop.SetBounds(NewRect(0, 0, w, desktopBottom))
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
		return &Event{
			What: EvMouse,
			Mouse: &MouseEvent{
				X:         x,
				Y:         y,
				Button:    ev.Buttons(),
				Modifiers: ev.Modifiers(),
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
