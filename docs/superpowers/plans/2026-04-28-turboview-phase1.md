# TurboView Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a runnable TUI application that boots, shows a desktop background with the classic ░ pattern, displays a status line with keyboard shortcuts, and exits cleanly on Alt+X.

**Architecture:** The `tv` package contains all views, widgets, events, and the application lifecycle. A separate `theme` package holds the ColorScheme struct, a registry, and built-in themes. Views draw into a clipped DrawBuffer (never directly to tcell.Screen), and the Application flushes the buffer to the screen after each draw pass. Event dispatch is handled by the Application, which forwards events to the StatusLine (for hotkey matching) and Desktop.

**Tech Stack:** Go 1.22+, tcell/v2 for terminal abstraction, tmux for e2e tests.

---

## Phase Roadmap

| Phase | Capability | Demo |
|-------|-----------|------|
| **1 (this plan)** | App boots, desktop, status line, clean exit | Blue desktop with ░, status bar, Alt+X exits |
| 2 | Windows | 2-3 windows, click to front, drag, resize, zoom, close, Alt+N switching |
| 3 | Widgets + focus traversal | Form with Button, InputLine, Label, CheckBox, RadioButton; Tab/Shift+Tab |
| 4 | Menus + dialogs | MenuBar, pull-downs, Dialog, ExecView, MessageBox, InputBox, context menus |
| 5 | Theme system + remaining widgets | 5 built-in themes, JSON config, ScrollBar, ListViewer |

---

## File Structure

```
turboview/
  go.mod
  tv/
    rect.go            — Point, Rect geometry types
    command.go         — CommandCode constants
    grow.go            — GrowFlag constants
    event.go           — Event, KeyEvent, MouseEvent, KeyBinding, EventType
    draw_buffer.go     — DrawBuffer, Cell, clipping, SubBuffer, FlushTo
    base_view.go       — ViewState, ViewOptions, View/Container/Widget/EventSource interfaces, BaseView
    tilde.go           — tilde label parsing (~shortcut~ notation)
    group.go           — Group (Container implementation)
    desktop.go         — Desktop (background pattern drawing)
    status.go          — StatusLine, StatusItem
    application.go     — Application lifecycle, event loop
  theme/
    scheme.go          — ColorScheme struct
    registry.go        — theme registration and lookup
    borland.go         — BorlandBlue theme
  e2e/
    harness.go         — tmux session management
    e2e_test.go        — tmux-based smoke tests
    testapp/
      basic/main.go    — Phase 1 demo app
```

---

### Task 1: Module Init + Geometry + Constants

- [ ] Complete

**Files:**
- Create: `go.mod` (via `go mod init`)
- Create: `tv/rect.go`
- Create: `tv/command.go`
- Create: `tv/grow.go`

**Requirements:**
- `NewPoint(3, 7)` returns `Point{X: 3, Y: 7}`
- `NewRect(5, 3, 20, 10)` returns `Rect` with `A=(5,3)`, `B=(25,13)`
- `Rect.Width()` returns `B.X - A.X`
- `Rect.Height()` returns `B.Y - A.Y`
- `Rect.Contains(point)` returns true for points within the rect (A inclusive, B exclusive)
- `Rect.Contains(B)` returns false (B is exclusive)
- `Rect.Intersect(other)` returns the overlapping region; two non-overlapping rects yield an empty rect
- `Rect.IsEmpty()` returns true when width or height is zero or negative
- `Rect.Moved(dx, dy)` shifts both A and B by the given offsets
- `CommandCode` constants: `CmQuit` through `CmPrev` sequential, `CmUser = 1000`
- `GrowFlag` constants: `GfGrowLoX` through `GfGrowHiY` as bit flags, `GfGrowAll` combines all four, `GfGrowRel` is a separate bit

**Implementation:**

```go
// tv/rect.go
package tv

type Point struct {
	X, Y int
}

func NewPoint(x, y int) Point {
	return Point{X: x, Y: y}
}

type Rect struct {
	A, B Point
}

func NewRect(x, y, w, h int) Rect {
	return Rect{
		A: Point{X: x, Y: y},
		B: Point{X: x + w, Y: y + h},
	}
}

func (r Rect) Width() int  { return r.B.X - r.A.X }
func (r Rect) Height() int { return r.B.Y - r.A.Y }

func (r Rect) Contains(p Point) bool {
	return p.X >= r.A.X && p.X < r.B.X && p.Y >= r.A.Y && p.Y < r.B.Y
}

func (r Rect) Intersect(s Rect) Rect {
	a := Point{X: max(r.A.X, s.A.X), Y: max(r.A.Y, s.A.Y)}
	b := Point{X: min(r.B.X, s.B.X), Y: min(r.B.Y, s.B.Y)}
	if a.X >= b.X || a.Y >= b.Y {
		return Rect{}
	}
	return Rect{A: a, B: b}
}

func (r Rect) IsEmpty() bool {
	return r.Width() <= 0 || r.Height() <= 0
}

func (r Rect) Moved(dx, dy int) Rect {
	return Rect{
		A: Point{X: r.A.X + dx, Y: r.A.Y + dy},
		B: Point{X: r.B.X + dx, Y: r.B.Y + dy},
	}
}
```

```go
// tv/command.go
package tv

type CommandCode int

const (
	CmQuit    CommandCode = iota + 1
	CmClose
	CmOK
	CmCancel
	CmYes
	CmNo
	CmMenu
	CmResize
	CmZoom
	CmTile
	CmCascade
	CmNext
	CmPrev
	CmUser CommandCode = 1000
)
```

```go
// tv/grow.go
package tv

type GrowFlag uint8

const (
	GfGrowLoX GrowFlag = 1 << iota
	GfGrowLoY
	GfGrowHiX
	GfGrowHiY
	GfGrowAll GrowFlag = GfGrowLoX | GfGrowLoY | GfGrowHiX | GfGrowHiY
	GfGrowRel GrowFlag = 1 << 4
)
```

**Setup commands:**
```bash
cd /Users/gnat/Source/Personal/tv3
git init
go mod init github.com/njt/turboview
go get github.com/gdamore/tcell/v2
```

**Run tests:** `go test ./tv/... -v -run TestRect -count=1`

**Commit:** `git commit -m "feat: add geometry types, command codes, and grow flags"`

---

### Task 2: Event Types + DrawBuffer

- [ ] Complete

**Files:**
- Create: `tv/event.go`
- Create: `tv/draw_buffer.go`

**Requirements:**
- `Event.Clear()` sets `What` to `EvNothing`; `Event.IsCleared()` returns true afterward
- `KbAlt('X')` matches a `KeyEvent` with `Key=KeyRune, Rune='x', Modifiers=ModAlt` (case-insensitive)
- `KbFunc(10)` matches a `KeyEvent` with `Key=KeyF10`
- `KbCtrl('N')` matches a `KeyEvent` with the Ctrl+N key code
- `KbNone()` never matches any event
- `NewDrawBuffer(80, 25)` creates an 80x25 buffer filled with spaces and `tcell.StyleDefault`
- `WriteChar(x, y, ch, style)` writes a cell at the given position; outside clip is a no-op
- `WriteStr(x, y, s, style)` writes a string horizontally starting at (x, y)
- `Fill(rect, ch, style)` fills the given rectangle
- `SubBuffer(rect)` shares the parent's backing store; writes in the child appear in the parent
- `SubBuffer` clip is the intersection of the parent's clip and the requested rect
- A child SubBuffer cannot write outside its allocated region
- `GetCell(x, y)` reads back the cell at local coordinates
- `FlushTo(screen)` copies all cells to the tcell screen

**Implementation:**

```go
// tv/event.go
package tv

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
)

type EventType uint16

const (
	EvNothing   EventType = 0
	EvMouse     EventType = 1 << 0
	EvKeyboard  EventType = 1 << 1
	EvCommand   EventType = 1 << 2
	EvBroadcast EventType = 1 << 3
)

type MouseEvent struct {
	X, Y       int
	Button     tcell.ButtonMask
	Modifiers  tcell.ModMask
	ClickCount int
}

type KeyEvent struct {
	Key       tcell.Key
	Rune      rune
	Modifiers tcell.ModMask
}

type Event struct {
	What    EventType
	Mouse   *MouseEvent
	Key     *KeyEvent
	Command CommandCode
	Info    any
}

func (e *Event) Clear() {
	e.What = EvNothing
}

func (e *Event) IsCleared() bool {
	return e.What == EvNothing
}

type KeyBinding struct {
	Key  tcell.Key
	Rune rune
	Mod  tcell.ModMask
}

func KbCtrl(ch rune) KeyBinding {
	ch = unicode.ToUpper(ch)
	return KeyBinding{Key: tcell.Key(ch - 'A' + 1), Mod: tcell.ModCtrl}
}

func KbAlt(ch rune) KeyBinding {
	return KeyBinding{Key: tcell.KeyRune, Rune: unicode.ToLower(ch), Mod: tcell.ModAlt}
}

func KbFunc(n int) KeyBinding {
	return KeyBinding{Key: tcell.Key(int(tcell.KeyF1) + n - 1)}
}

func KbNone() KeyBinding {
	return KeyBinding{}
}

func (kb KeyBinding) Matches(ke *KeyEvent) bool {
	if kb == (KeyBinding{}) {
		return false
	}
	if kb.Key == tcell.KeyRune {
		return unicode.ToLower(ke.Rune) == unicode.ToLower(kb.Rune) &&
			ke.Modifiers&kb.Mod == kb.Mod
	}
	return ke.Key == kb.Key && ke.Modifiers&kb.Mod == kb.Mod
}
```

```go
// tv/draw_buffer.go
package tv

import "github.com/gdamore/tcell/v2"

type Cell struct {
	Rune  rune
	Combc []rune
	Style tcell.Style
}

type DrawBuffer struct {
	cells  [][]Cell
	clip   Rect
	offset Point
}

func NewDrawBuffer(w, h int) *DrawBuffer {
	cells := make([][]Cell, h)
	for y := range cells {
		cells[y] = make([]Cell, w)
		for x := range cells[y] {
			cells[y][x] = Cell{Rune: ' ', Style: tcell.StyleDefault}
		}
	}
	return &DrawBuffer{
		cells: cells,
		clip:  NewRect(0, 0, w, h),
	}
}

func (db *DrawBuffer) WriteChar(x, y int, ch rune, style tcell.Style) {
	ax, ay := x+db.offset.X, y+db.offset.Y
	if !db.clip.Contains(NewPoint(ax, ay)) {
		return
	}
	if ay < 0 || ay >= len(db.cells) || ax < 0 || ax >= len(db.cells[0]) {
		return
	}
	db.cells[ay][ax] = Cell{Rune: ch, Style: style}
}

func (db *DrawBuffer) WriteStr(x, y int, s string, style tcell.Style) {
	for i, ch := range s {
		db.WriteChar(x+i, y, ch, style)
	}
}

func (db *DrawBuffer) Fill(r Rect, ch rune, style tcell.Style) {
	for y := r.A.Y; y < r.B.Y; y++ {
		for x := r.A.X; x < r.B.X; x++ {
			db.WriteChar(x, y, ch, style)
		}
	}
}

func (db *DrawBuffer) SubBuffer(r Rect) *DrawBuffer {
	absClip := Rect{
		A: NewPoint(r.A.X+db.offset.X, r.A.Y+db.offset.Y),
		B: NewPoint(r.B.X+db.offset.X, r.B.Y+db.offset.Y),
	}
	return &DrawBuffer{
		cells:  db.cells,
		clip:   db.clip.Intersect(absClip),
		offset: NewPoint(r.A.X+db.offset.X, r.A.Y+db.offset.Y),
	}
}

func (db *DrawBuffer) GetCell(x, y int) Cell {
	ax, ay := x+db.offset.X, y+db.offset.Y
	if ay < 0 || ay >= len(db.cells) || ax < 0 || ax >= len(db.cells[0]) {
		return Cell{}
	}
	return db.cells[ay][ax]
}

func (db *DrawBuffer) Width() int  { return len(db.cells[0]) }
func (db *DrawBuffer) Height() int { return len(db.cells) }

func (db *DrawBuffer) FlushTo(screen tcell.Screen) {
	for y, row := range db.cells {
		for x, cell := range row {
			screen.SetContent(x, y, cell.Rune, cell.Combc, cell.Style)
		}
	}
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add event system, key bindings, and draw buffer"`

---

### Task 3: ColorScheme + BorlandBlue Theme

- [ ] Complete

**Files:**
- Create: `theme/scheme.go`
- Create: `theme/registry.go`
- Create: `theme/borland.go`

**Requirements:**
- `ColorScheme` is a struct with named `tcell.Style` fields for every semantic UI element (window, desktop, dialog, widgets, menu, status)
- `Register(name, scheme)` stores a scheme by name; `Get(name)` retrieves it
- `Get` with unknown name returns nil
- Register with a duplicate name overwrites the previous scheme
- `BorlandBlue` is a pre-registered `*ColorScheme` with classic blue-background Turbo Vision colors
- `Get("borland-blue")` returns the BorlandBlue scheme

**Implementation:**

```go
// theme/scheme.go
package theme

import "github.com/gdamore/tcell/v2"

type ColorScheme struct {
	WindowBackground    tcell.Style
	WindowFrameActive   tcell.Style
	WindowFrameInactive tcell.Style
	WindowTitle         tcell.Style
	WindowShadow        tcell.Style

	DesktopBackground tcell.Style

	DialogBackground tcell.Style
	DialogFrame      tcell.Style

	ButtonNormal   tcell.Style
	ButtonDefault  tcell.Style
	ButtonShadow   tcell.Style
	ButtonShortcut tcell.Style

	InputNormal    tcell.Style
	InputSelection tcell.Style

	LabelNormal   tcell.Style
	LabelShortcut tcell.Style

	CheckBoxNormal    tcell.Style
	RadioButtonNormal tcell.Style

	ListNormal   tcell.Style
	ListSelected tcell.Style
	ListFocused  tcell.Style

	ScrollBar   tcell.Style
	ScrollThumb tcell.Style

	MenuNormal   tcell.Style
	MenuShortcut tcell.Style
	MenuSelected tcell.Style
	MenuDisabled tcell.Style

	StatusNormal   tcell.Style
	StatusShortcut tcell.Style
}
```

```go
// theme/registry.go
package theme

var registry = map[string]*ColorScheme{}

func Register(name string, scheme *ColorScheme) {
	registry[name] = scheme
}

func Get(name string) *ColorScheme {
	return registry[name]
}
```

```go
// theme/borland.go
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

		DesktopBackground: s(tcell.ColorTeal, tcell.ColorBlue),

		DialogBackground: s(tcell.ColorBlack, tcell.ColorTeal),
		DialogFrame:      s(tcell.ColorWhite, tcell.ColorTeal),

		ButtonNormal:   s(tcell.ColorBlack, tcell.ColorGreen),
		ButtonDefault:  s(tcell.ColorWhite, tcell.ColorGreen),
		ButtonShadow:   s(tcell.ColorBlack, tcell.ColorBlack),
		ButtonShortcut: s(tcell.ColorYellow, tcell.ColorGreen),

		InputNormal:    s(tcell.ColorWhite, tcell.ColorBlue),
		InputSelection: s(tcell.ColorBlue, tcell.ColorTeal),

		LabelNormal:   s(tcell.ColorBlack, tcell.ColorTeal),
		LabelShortcut: s(tcell.ColorYellow, tcell.ColorTeal),

		CheckBoxNormal:    s(tcell.ColorBlack, tcell.ColorTeal),
		RadioButtonNormal: s(tcell.ColorBlack, tcell.ColorTeal),

		ListNormal:   s(tcell.ColorBlack, tcell.ColorTeal),
		ListSelected: s(tcell.ColorWhite, tcell.ColorBlack),
		ListFocused:  s(tcell.ColorYellow, tcell.ColorBlue),

		ScrollBar:   s(tcell.ColorTeal, tcell.ColorBlue),
		ScrollThumb: s(tcell.ColorWhite, tcell.ColorBlue),

		MenuNormal:   s(tcell.ColorBlack, tcell.ColorTeal),
		MenuShortcut: s(tcell.ColorYellow, tcell.ColorTeal),
		MenuSelected: s(tcell.ColorWhite, tcell.ColorBlack),
		MenuDisabled: s(tcell.ColorSilver, tcell.ColorTeal),

		StatusNormal:   s(tcell.ColorBlack, tcell.ColorTeal),
		StatusShortcut: s(tcell.ColorYellow, tcell.ColorTeal),
	}

	Register("borland-blue", BorlandBlue)
}
```

**Run tests:** `go test ./theme/... -v -count=1`

**Commit:** `git commit -m "feat: add color scheme, theme registry, and BorlandBlue theme"`

---

### Task 4: View/Container/Widget Interfaces + BaseView + Tilde Parsing

- [ ] Complete

**Files:**
- Create: `tv/base_view.go`
- Create: `tv/tilde.go`

**Requirements:**
- `ViewState` flags: `SfVisible` through `SfDragging` as bit flags
- `ViewOptions` flags: `OfSelectable` through `OfCentered` as bit flags
- `View` interface declares: `Draw`, `HandleEvent`, `Bounds/SetBounds`, `GrowMode/SetGrowMode`, `Owner/SetOwner`, `State/SetState`, `HasState`, `EventMask/SetEventMask`, `Options/SetOptions`, `HasOption`, `ColorScheme`
- `Container` interface extends `View` with: `Insert`, `Remove`, `Children`, `FocusedChild`, `SetFocusedChild`, `ExecView`
- `Widget` interface extends `View` (marker interface, no additional methods)
- `EventSource` interface declares `PollEvent() *Event`
- `BaseView` provides default implementations for all `View` methods
- `BaseView.SetBounds(rect)` stores origin and size; `Bounds()` reconstructs the Rect
- `BaseView.SetState(flag, true)` sets the flag; `SetState(flag, false)` clears it
- `BaseView.HasState(flag)` returns true if the flag is set
- `BaseView.HasOption(flag)` returns true if the flag is set
- `BaseView.SetOptions(flag, true)` sets; `SetOptions(flag, false)` clears
- `BaseView.ColorScheme()` returns its own scheme if non-nil, else walks up the owner chain, else returns nil
- `BaseView.Draw()` and `HandleEvent()` are no-ops by default
- `ParseTildeLabel("~Alt+X~ Exit")` returns `[{Text:"Alt+X", Shortcut:true}, {Text:" Exit", Shortcut:false}]`
- `ParseTildeLabel("No tilde")` returns `[{Text:"No tilde", Shortcut:false}]`
- `ParseTildeLabel("~O~K")` returns `[{Text:"O", Shortcut:true}, {Text:"K", Shortcut:false}]`

**Implementation:**

```go
// tv/base_view.go
package tv

import "github.com/njt/turboview/theme"

type ViewState uint16

const (
	SfVisible  ViewState = 1 << iota
	SfFocused
	SfSelected
	SfModal
	SfDisabled
	SfExposed
	SfDragging
)

type ViewOptions uint16

const (
	OfSelectable  ViewOptions = 1 << iota
	OfTopSelect
	OfFirstClick
	OfPreProcess
	OfPostProcess
	OfCentered
)

type HelpContext uint16

const HcNoContext HelpContext = 0

type View interface {
	Draw(buf *DrawBuffer)
	HandleEvent(event *Event)
	Bounds() Rect
	SetBounds(Rect)
	GrowMode() GrowFlag
	SetGrowMode(GrowFlag)
	Owner() Container
	SetOwner(Container)
	State() ViewState
	SetState(ViewState, bool)
	EventMask() EventType
	SetEventMask(EventType)
	Options() ViewOptions
	SetOptions(ViewOptions, bool)
	HasState(ViewState) bool
	HasOption(ViewOptions) bool
	ColorScheme() *theme.ColorScheme
}

type Container interface {
	View
	Insert(View)
	Remove(View)
	Children() []View
	FocusedChild() View
	SetFocusedChild(View)
	ExecView(View) CommandCode
}

type Widget interface {
	View
}

type EventSource interface {
	PollEvent() *Event
}

type BaseView struct {
	origin    Point
	size      Point
	growMode  GrowFlag
	state     ViewState
	eventMask EventType
	options   ViewOptions
	owner     Container
	scheme    *theme.ColorScheme
	helpCtx   HelpContext
}

func (b *BaseView) Draw(buf *DrawBuffer)        {}
func (b *BaseView) HandleEvent(event *Event)     {}

func (b *BaseView) Bounds() Rect {
	return Rect{
		A: b.origin,
		B: Point{X: b.origin.X + b.size.X, Y: b.origin.Y + b.size.Y},
	}
}

func (b *BaseView) SetBounds(r Rect) {
	b.origin = r.A
	b.size = Point{X: r.Width(), Y: r.Height()}
}

func (b *BaseView) GrowMode() GrowFlag      { return b.growMode }
func (b *BaseView) SetGrowMode(gm GrowFlag) { b.growMode = gm }

func (b *BaseView) Owner() Container      { return b.owner }
func (b *BaseView) SetOwner(c Container)  { b.owner = c }

func (b *BaseView) State() ViewState { return b.state }
func (b *BaseView) SetState(flag ViewState, on bool) {
	if on {
		b.state |= flag
	} else {
		b.state &^= flag
	}
}

func (b *BaseView) EventMask() EventType       { return b.eventMask }
func (b *BaseView) SetEventMask(mask EventType) { b.eventMask = mask }

func (b *BaseView) Options() ViewOptions { return b.options }
func (b *BaseView) SetOptions(flag ViewOptions, on bool) {
	if on {
		b.options |= flag
	} else {
		b.options &^= flag
	}
}

func (b *BaseView) HasState(flag ViewState) bool    { return b.state&flag != 0 }
func (b *BaseView) HasOption(flag ViewOptions) bool  { return b.options&flag != 0 }

func (b *BaseView) HelpCtx() HelpContext        { return b.helpCtx }
func (b *BaseView) SetHelpCtx(hc HelpContext)    { b.helpCtx = hc }

func (b *BaseView) ColorScheme() *theme.ColorScheme {
	if b.scheme != nil {
		return b.scheme
	}
	if b.owner != nil {
		return b.owner.ColorScheme()
	}
	return nil
}
```

```go
// tv/tilde.go
package tv

type LabelSegment struct {
	Text     string
	Shortcut bool
}

func ParseTildeLabel(label string) []LabelSegment {
	var segments []LabelSegment
	inTilde := false
	current := ""
	for _, r := range label {
		if r == '~' {
			if current != "" {
				segments = append(segments, LabelSegment{Text: current, Shortcut: inTilde})
				current = ""
			}
			inTilde = !inTilde
			continue
		}
		current += string(r)
	}
	if current != "" {
		segments = append(segments, LabelSegment{Text: current, Shortcut: inTilde})
	}
	return segments
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add View/Container/Widget interfaces, BaseView, and tilde parsing"`

---

### Task 5: Group (Container Implementation)

- [ ] Complete

**Files:**
- Create: `tv/group.go`

**Requirements:**
- `NewGroup(bounds)` creates a Group with the given bounds, `SfVisible` state set
- `SetFacade(container)` sets an optional facade: when set, inserted children see the facade as their owner instead of the Group. This enables composition — Desktop/Window hold an internal Group and set the facade to themselves so the owner chain is correct.
- `Insert(view)` adds a child to the end of the child list and sets the child's owner to this Group (or the facade if set)
- `Insert(view)` on an `OfSelectable` view selects it (sets `SfSelected`) and deselects the previous selection
- `Remove(view)` removes the child and clears its owner; if it was focused, selects the previous child
- `Children()` returns the child list in insertion order
- `FocusedChild()` returns the child with `SfSelected` state, or nil
- `SetFocusedChild(view)` deselects the current selection and selects the given view
- `Draw(buf)` iterates children back-to-front; for each visible child, creates a SubBuffer clipped to the child's bounds and calls `child.Draw(sub)`
- `HandleEvent(event)` forwards keyboard/command events to the focused child
- `ExecView` is a stub that panics with "not implemented" (implemented in Phase 4)

**Implementation:**

```go
// tv/group.go
package tv

type Group struct {
	BaseView
	children []View
	focused  View
	facade   Container
}

func NewGroup(bounds Rect) *Group {
	g := &Group{}
	g.SetBounds(bounds)
	g.SetState(SfVisible, true)
	return g
}

func (g *Group) SetFacade(c Container) {
	g.facade = c
}

func (g *Group) Insert(v View) {
	owner := Container(g)
	if g.facade != nil {
		owner = g.facade
	}
	v.SetOwner(owner)
	g.children = append(g.children, v)
	if v.HasOption(OfSelectable) {
		g.selectChild(v)
	}
}

func (g *Group) Remove(v View) {
	for i, child := range g.children {
		if child == v {
			g.children = append(g.children[:i], g.children[i+1:]...)
			v.SetOwner(nil)
			if g.focused == v {
				g.focused = nil
				g.selectPrevious()
			}
			return
		}
	}
}

func (g *Group) Children() []View {
	return g.children
}

func (g *Group) FocusedChild() View {
	return g.focused
}

func (g *Group) SetFocusedChild(v View) {
	g.selectChild(v)
}

func (g *Group) ExecView(v View) CommandCode {
	panic("ExecView not implemented")
}

func (g *Group) Draw(buf *DrawBuffer) {
	for _, child := range g.children {
		if !child.HasState(SfVisible) {
			continue
		}
		childBounds := child.Bounds()
		sub := buf.SubBuffer(childBounds)
		child.Draw(sub)
	}
}

func (g *Group) HandleEvent(event *Event) {
	if event.IsCleared() {
		return
	}
	if g.focused != nil {
		g.focused.HandleEvent(event)
	}
}

func (g *Group) selectChild(v View) {
	if g.focused != nil && g.focused != v {
		g.focused.SetState(SfSelected, false)
	}
	g.focused = v
	if v != nil {
		v.SetState(SfSelected, true)
	}
}

func (g *Group) selectPrevious() {
	for i := len(g.children) - 1; i >= 0; i-- {
		if g.children[i].HasOption(OfSelectable) {
			g.selectChild(g.children[i])
			return
		}
	}
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add Group container with child management and event forwarding"`

---

### Task 6: Desktop + StatusLine

- [ ] Complete

**Files:**
- Create: `tv/desktop.go`
- Create: `tv/status.go`

**Requirements:**
- `NewDesktop(bounds)` creates a Desktop with `SfVisible` state and `GfGrowAll` grow mode; Desktop embeds `BaseView` directly (not Group) to avoid embedding chains per the spec's "one level of embedding" rule. Phase 2 will add Container functionality via composition with an internal Group.
- `Desktop.Draw(buf)` fills its area with the `'░'` character using `ColorScheme().DesktopBackground`
- If Desktop has no ColorScheme, Draw fills with `'░'` and `tcell.StyleDefault`
- `NewStatusItem(label, keyBinding, command)` creates a StatusItem with the given fields
- `NewStatusLine(items...)` creates a StatusLine with `SfVisible` state
- `StatusLine.Draw(buf)` fills its row with `StatusNormal` style, then renders each item's label with tilde shortcut segments in `StatusShortcut` style and normal text in `StatusNormal` style, separated by 2-space gaps
- `StatusLine.HandleEvent(event)` checks keyboard events against item key bindings; on match, transforms the event from `EvKeyboard` to `EvCommand` with the item's command code

**Implementation:**

```go
// tv/desktop.go
package tv

import "github.com/gdamore/tcell/v2"

type Desktop struct {
	BaseView
	pattern rune
}

func NewDesktop(bounds Rect) *Desktop {
	d := &Desktop{
		pattern: '░',
	}
	d.SetBounds(bounds)
	d.SetState(SfVisible, true)
	d.SetGrowMode(GfGrowAll)
	return d
}

func (d *Desktop) Draw(buf *DrawBuffer) {
	w, h := d.Bounds().Width(), d.Bounds().Height()
	style := tcell.StyleDefault
	if cs := d.ColorScheme(); cs != nil {
		style = cs.DesktopBackground
	}
	buf.Fill(NewRect(0, 0, w, h), d.pattern, style)
}
```

```go
// tv/status.go
package tv

import (
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type StatusItem struct {
	Label      string
	KeyBinding KeyBinding
	Command    CommandCode
	HelpCtx    HelpContext
}

func NewStatusItem(label string, kb KeyBinding, cmd CommandCode) *StatusItem {
	return &StatusItem{
		Label:      label,
		KeyBinding: kb,
		Command:    cmd,
	}
}

type StatusLine struct {
	BaseView
	items []*StatusItem
}

func NewStatusLine(items ...*StatusItem) *StatusLine {
	sl := &StatusLine{
		items: items,
	}
	sl.SetState(SfVisible, true)
	return sl
}

func (sl *StatusLine) Draw(buf *DrawBuffer) {
	w := sl.Bounds().Width()
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	if cs := sl.ColorScheme(); cs != nil {
		normalStyle = cs.StatusNormal
		shortcutStyle = cs.StatusShortcut
	}

	buf.Fill(NewRect(0, 0, w, 1), ' ', normalStyle)

	x := 1
	for i, item := range sl.items {
		if i > 0 {
			x += 2
		}
		segments := ParseTildeLabel(item.Label)
		for _, seg := range segments {
			style := normalStyle
			if seg.Shortcut {
				style = shortcutStyle
			}
			buf.WriteStr(x, 0, seg.Text, style)
			x += utf8.RuneCountInString(seg.Text)
		}
	}
}

func (sl *StatusLine) HandleEvent(event *Event) {
	if event.What != EvKeyboard || event.Key == nil {
		return
	}
	for _, item := range sl.items {
		if item.KeyBinding.Matches(event.Key) {
			event.What = EvCommand
			event.Command = item.Command
			event.Key = nil
			return
		}
	}
}
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add Desktop with background pattern and StatusLine with hotkey dispatch"`

---

### Task 7: Application

- [ ] Complete

**Files:**
- Create: `tv/application.go`

**Requirements:**
- `NewApplication()` with no `WithScreen` option creates a real terminal screen; returns error if screen init fails
- `NewApplication(WithScreen(sim))` uses the injected screen; does not call `Init()` or `Fini()` on it
- `NewApplication(WithStatusLine(sl))` attaches the status line
- `NewApplication(WithTheme(scheme))` sets the application's color scheme
- Default theme is `theme.BorlandBlue` when no `WithTheme` is given
- Desktop is always created automatically; its ColorScheme is set to the application's theme
- StatusLine's ColorScheme is set to the application's theme if a StatusLine is provided
- `Run()` enters the event loop: poll → convert → dispatch → redraw → repeat
- `Run()` exits cleanly when a `CmQuit` command event is processed, returning nil
- `Run()` handles terminal resize events by recalculating Desktop and StatusLine bounds
- `Draw(buf)` renders Desktop into a SubBuffer for rows 0..h-2, StatusLine into a SubBuffer for the last row (Phase 1 has no MenuBar, so Desktop starts at row 0; Phase 4 will adjust to row 1 when MenuBar is added)
- `PostCommand(cmd, info)` is safe to call from any goroutine and enqueues a command processed on the next loop iteration
- `PostCommand(CmQuit, nil)` causes `Run()` to exit
- Event conversion: `tcell.EventKey` becomes `EvKeyboard`, `tcell.EventMouse` becomes `EvMouse`, `tcell.EventResize` becomes `EvCommand/CmResize`

**Implementation:**

```go
// tv/application.go
package tv

import (
	"time"

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

type Application struct {
	bounds     Rect
	screen     tcell.Screen
	screenOwn  bool
	desktop    *Desktop
	statusLine *StatusLine
	scheme     *theme.ColorScheme
	quit       bool
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

	if app.statusLine != nil {
		app.statusLine.scheme = app.scheme
	}

	return app, nil
}

func (app *Application) Desktop() *Desktop       { return app.desktop }
func (app *Application) StatusLine() *StatusLine  { return app.statusLine }
func (app *Application) Screen() tcell.Screen     { return app.screen }

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
		tcellEv := app.screen.PollEvent()
		if tcellEv == nil {
			break
		}

		if _, ok := tcellEv.(*tcell.EventResize); ok {
			w, h = app.screen.Size()
			app.bounds = NewRect(0, 0, w, h)
			app.layoutChildren()
		}

		event := app.convertEvent(tcellEv)
		if event != nil {
			app.handleEvent(event)
		}

		app.drawAndFlush()
	}

	return nil
}

func (app *Application) PostCommand(cmd CommandCode, info any) {
	ev := &cmdTcellEvent{cmd: cmd, info: info}
	ev.SetEventTime(time.Now())
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
	if app.statusLine != nil {
		app.statusLine.HandleEvent(event)
	}

	if !event.IsCleared() && app.desktop != nil {
		app.desktop.HandleEvent(event)
	}

	if !event.IsCleared() && event.What == EvCommand {
		switch event.Command {
		case CmQuit:
			app.quit = true
			event.Clear()
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
```

**Run tests:** `go test ./tv/... -v -count=1`

**Commit:** `git commit -m "feat: add Application with event loop, screen management, and PostCommand"`

---

### Task 8: Integration Checkpoint — Full Draw and Event Chain

- [ ] Complete

**Purpose:** Verify that Application, Desktop, StatusLine, and theme system work together.

**Requirements (for test writer):**
- Creating an Application with a SimulationScreen, StatusLine, and BorlandBlue theme succeeds without error
- Calling `app.Draw(buf)` on an 80x25 buffer renders: Desktop fills rows 0-23 with `'░'` in `DesktopBackground` style, StatusLine fills row 24
- StatusLine row contains the text "Alt+X" rendered partly in `StatusShortcut` style and " Exit" in `StatusNormal` style
- Injecting an Alt+X key event through `app.handleEvent()` causes the event to be transformed from `EvKeyboard` to `EvCommand/CmQuit` and sets `app.quit` to true
- Injecting a regular key (e.g., 'a') does NOT trigger quit
- After terminal resize, `layoutChildren()` recalculates bounds: desktop height shrinks/grows, status line stays on the last row
- `PostCommand(CmQuit, nil)` causes `Run()` to exit (test runs `Run()` in a goroutine)
- ColorScheme propagation: Desktop's `ColorScheme()` returns the scheme set by Application; StatusLine's `ColorScheme()` returns the same

**Components to wire up:** Application, Desktop, StatusLine, BorlandBlue (all real, no mocks)

**Run tests:** `go test ./tv/... -v -run TestIntegration -count=1`

**Commit:** `git commit -m "test: add integration tests for Application draw and event chain"`

---

### Task 9: Demo App + E2E Tests

- [ ] Complete

**Files:**
- Create: `e2e/testapp/basic/main.go`
- Create: `e2e/harness.go`
- Create: `e2e/e2e_test.go`

**Requirements:**
- Demo app launches, displays blue desktop background with `░` pattern, shows status line with "Alt+X Exit" and "F10 Menu" items, and exits on Alt+X
- E2E test builds the demo app binary
- E2E test starts the binary in a tmux session sized 80x25
- E2E test captures the pane and verifies the `░` pattern appears in the desktop area
- E2E test captures the pane and verifies status line text contains "Alt+X"
- E2E test sends Alt+X via `tmux send-keys M-x` and verifies the app exits (tmux session terminates)
- E2E harness provides `startTmux`, `tmuxSendKeys`, `tmuxCapture`, `tmuxHasSession` helpers

**Implementation:**

```go
// e2e/testapp/basic/main.go
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

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
```

```go
// e2e/harness.go
package e2e

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

func startTmux(t *testing.T, session string, cmd string) {
	t.Helper()
	err := exec.Command("tmux", "new-session", "-d", "-s", session, "-x", "80", "-y", "25", cmd).Run()
	if err != nil {
		t.Fatalf("failed to start tmux session: %v", err)
	}
	t.Cleanup(func() {
		exec.Command("tmux", "kill-session", "-t", session).Run()
	})
	time.Sleep(1 * time.Second)
}

func tmuxSendKeys(t *testing.T, session string, keys ...string) {
	t.Helper()
	args := append([]string{"send-keys", "-t", session}, keys...)
	if err := exec.Command("tmux", args...).Run(); err != nil {
		t.Fatalf("failed to send keys: %v", err)
	}
	time.Sleep(300 * time.Millisecond)
}

func tmuxCapture(t *testing.T, session string) []string {
	t.Helper()
	out, err := exec.Command("tmux", "capture-pane", "-t", session, "-p").Output()
	if err != nil {
		t.Fatalf("failed to capture pane: %v", err)
	}
	return strings.Split(string(out), "\n")
}

func tmuxHasSession(session string) bool {
	return exec.Command("tmux", "has-session", "-t", session).Run() == nil
}
```

```go
// e2e/e2e_test.go
package e2e

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func projectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..")
}

func TestBasicAppBoot(t *testing.T) {
	root := projectRoot()
	binPath := filepath.Join(root, "e2e", "testapp", "basic", "basic")

	out, err := exec.Command("go", "build", "-o", binPath, filepath.Join(root, "e2e", "testapp", "basic")).CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	t.Cleanup(func() { exec.Command("rm", binPath).Run() })

	session := "tv3-e2e-basic"
	exec.Command("tmux", "kill-session", "-t", session).Run()

	startTmux(t, session, binPath)

	lines := tmuxCapture(t, session)
	desktopHasPattern := false
	for _, line := range lines[:len(lines)-2] {
		if strings.Contains(line, "░") {
			desktopHasPattern = true
			break
		}
	}
	if !desktopHasPattern {
		t.Error("desktop background pattern '░' not found")
	}

	statusFound := false
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			if strings.Contains(lines[i], "Alt+X") {
				statusFound = true
			}
			break
		}
	}
	if !statusFound {
		t.Error("status line should contain 'Alt+X'")
	}

	tmuxSendKeys(t, session, "M-x")

	exited := false
	for i := 0; i < 15; i++ {
		if !tmuxHasSession(session) {
			exited = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !exited {
		t.Error("app did not exit after Alt+X")
	}
}
```

**Run tests:** `go test ./e2e/... -v -count=1 -timeout 30s`

**Commit:** `git commit -m "feat: add demo app and tmux-based e2e smoke tests"`
