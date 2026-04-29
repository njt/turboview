# turboview Design Spec

A Go TUI framework inspired by Borland's Turbo Vision. Provides a complete text-mode windowing system with overlapping windows, automatic event dispatch, and a rich widget set. Built on tcell.

**Module**: `github.com/njt/turboview`
**Primary package**: `tv` (`github.com/njt/turboview/tv`)
**Theme package**: `theme` (`github.com/njt/turboview/theme`)
**License**: TBD
**Go version**: 1.22+ (minimum)

## Design Principles

1. **Easy for agents** — single `tv` package for all widgets, windows, menus, and event handling. One import covers everything. Compile-time type safety prevents misuse (e.g., inserting children into leaf widgets).
2. **As flexible as the original** — the Turbo Vision interaction model, ported faithfully: overlapping windows, three-phase event dispatch, modal dialogs, automatic focus traversal, GrowMode-based resize. The Borland Turbo Vision Programming Guide should remain a useful conceptual reference.
3. **Testable from the start** — test-first development. Every widget tested against `tcell.SimulationScreen`. E2e smoke tests via tmux.

## Non-Goals

- Tier 3 widgets (TEditor, help system, color picker) — deferred to future work
- Faithful tvision API — we use tvision's naming conventions and conceptual model, but the Go API uses interfaces and composition, not embedding chains
- Charm/Bubble Tea compatibility — no dependency on or interop with the Charm ecosystem

---

## 1. Core Architecture

### Type Hierarchy

```
View (interface)       — anything that draws and handles events
├── Container (interface) — a View that holds child Views
│   ├── Group          — base container implementation
│   ├── Window         — Group + frame, title, drag, resize, zoom
│   ├── Dialog         — Window + modal execution
│   └── Desktop        — the workspace that manages windows
├── Widget (interface) — a leaf View (no children)
│   ├── Button
│   ├── InputLine
│   ├── CheckBox
│   ├── RadioButton
│   ├── Label
│   ├── StaticText
│   ├── ScrollBar
│   ├── ListViewer
│   └── ...
└── Application        — top-level: MenuBar + Desktop + StatusLine
```

### Key Interfaces

```go
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
    ColorScheme() *ColorScheme
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
    // Marker interface — the type system prevents Insert() calls on widgets
}
```

### BaseView

A `BaseView` struct provides default implementations for the `View` interface. Concrete types embed `BaseView` — this is the one permitted level of embedding. "No embedding chains" means no multi-level A-embeds-B-embeds-C hierarchies; every concrete type embeds `BaseView` directly.

```go
type BaseView struct {
    origin    Point       // position relative to owner's client area
    size      Point       // width and height
    growMode  GrowFlag
    state     ViewState
    eventMask EventType
    options   ViewOptions
    owner     Container
    scheme    *ColorScheme  // nil = inherit from owner
}
```

Only `Group` and its subtypes (`Window`, `Dialog`, `Desktop`) implement `Container`. Attempting to call `Insert()` on a `Button` is a compile error.

### Constructor Conventions

All constructors follow a consistent pattern:

- **Bounds-first**: constructors that create positioned views take `bounds Rect` as the first parameter.
- **Functional options**: constructors accept variadic `Option` functions for optional configuration. Each type defines its own option type to prevent cross-type misuse.
- **Fallible constructors**: `NewApplication` returns `(*Application, error)` because tcell.Screen initialization can fail. All other constructors are infallible — they produce valid zero-state views.

```go
// Infallible — widgets and containers
func NewButton(bounds Rect, title string, command CommandCode, opts ...ButtonOption) *Button
func NewInputLine(bounds Rect, maxLen int, opts ...InputLineOption) *InputLine
func NewWindow(bounds Rect, title string, opts ...WindowOption) *Window
func NewDialog(bounds Rect, title string, opts ...DialogOption) *Dialog
func NewGroup(bounds Rect) *Group
func NewDesktop(bounds Rect) *Desktop

// Fallible — application
func NewApplication(opts ...AppOption) (*Application, error)

// Geometry helpers
func NewRect(x, y, w, h int) Rect
func NewPoint(x, y int) Point
```

Application options:
```go
func WithMenuBar(mb *MenuBar) AppOption
func WithStatusLine(sl *StatusLine) AppOption
func WithTheme(scheme *ColorScheme) AppOption
func WithScreen(s tcell.Screen) AppOption   // inject screen for testing
func WithConfigFile(path string) AppOption
```

`WithScreen` allows injecting a `tcell.SimulationScreen` for unit-testing the Application event loop without a real terminal.

### Geometry Types

```go
type Point struct {
    X, Y int
}

type Rect struct {
    A, B Point  // A = top-left, B = bottom-right (exclusive)
}
```

Following tvision convention, `Rect` uses two-point form: `A` is inclusive top-left, `B` is exclusive bottom-right. Width = `B.X - A.X`, Height = `B.Y - A.Y`. Helper constructors: `NewRect(x, y, w, h int) Rect`.

### Coordinate System

All view positions are **owner-relative**. A view's `origin` is relative to its owner's client area (the interior of the owner, inside any frame). Screen-absolute coordinates are computed at draw time by walking the owner chain and accumulating offsets. Mouse events arrive in screen coordinates and are translated to local coordinates as they traverse the hierarchy.

### ViewState Flags

```go
const (
    SfVisible   ViewState = 1 << iota  // view is drawn
    SfFocused                           // view has input focus
    SfSelected                          // view is the active child of its owner
    SfModal                             // view is executing modally
    SfDisabled                          // view ignores events
    SfExposed                           // view is on screen (set by framework)
    SfDragging                          // view is being dragged
)
```

### ViewOptions Flags

```go
const (
    OfSelectable   ViewOptions = 1 << iota  // view can receive focus
    OfTopSelect                              // selecting this view brings its owner to top
    OfFirstClick                             // first click selects AND processes
    OfPreProcess                             // view sees focused events before the focused sibling
    OfPostProcess                            // view sees unhandled focused events after the focused sibling
    OfCentered                               // view centers itself in its owner
)
```

### Convenience Methods

`BaseView` provides convenience methods for common flag checks:

```go
func (b *BaseView) HasState(flag ViewState) bool    { return b.state&flag != 0 }
func (b *BaseView) HasOption(flag ViewOptions) bool  { return b.options&flag != 0 }
```

These avoid the `v.State()&SfFocused != 0` pattern that agents frequently get wrong (missing parentheses, wrong operator).

For tests, embed `BaseView` to satisfy the `View` interface with minimal boilerplate — all methods have sensible zero-value defaults.

---

## 2. Event System

### Event Types

```go
type Event struct {
    What    EventType      // EvMouse, EvKeyboard, EvCommand, EvBroadcast
    Mouse   *MouseEvent    // position, button, modifiers, click count
    Key     *KeyEvent      // key code, rune, modifiers
    Command CommandCode    // CmOK, CmCancel, CmClose, CmQuit, etc.
    Info    any            // arbitrary payload for broadcast/command
}
```

An event is consumed by calling `event.Clear()`, which stops propagation.

### Dispatch Rules

- **Positional events** (mouse): dispatched to the view under the cursor. Coordinates are translated to the target view's local space as they traverse the hierarchy.
- **Focused events** (keyboard, commands): dispatched through the focus chain using three phases.
- **Broadcast events**: delivered to every view in the tree.

### Three-Phase Focused Event Dispatch

1. **Preprocess** — siblings with `OfPreProcess` see the event first. Used for accelerator keys and menu hotkeys.
2. **Focused** — the focused child handles the event.
3. **Postprocess** — siblings with `OfPostProcess` get unhandled events. Used for default buttons responding to Enter.

This is what makes Tab-between-fields, Alt+letter shortcuts, and default-button-on-Enter work automatically without per-app wiring.

### Standard Commands

```go
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
    CmUser    CommandCode = 1000  // app-defined commands start here
)
```

### Modal Execution

`Container.ExecView(view)` enters a local event loop, dispatching events only to the modal view and its children until it closes. Returns a `CommandCode` indicating how the modal was dismissed (e.g., `CmOK`, `CmCancel`).

---

## 3. Window Management

### Window

Wraps a `Group` with:

- **Frame**: box-drawing border (single or double line), title centered in top border, close icon `[x]` on the left, zoom icon on the right.
- **Drag**: click and drag the title bar to reposition.
- **Resize**: drag bottom-right or bottom-left corner.
- **Zoom**: toggle between normal size and full desktop via zoom icon or double-click on title bar.
- **Shadow**: optional shadow effect (offset cells below and right of the window).

Active vs. inactive windows render with different color attributes from the current theme.

### Desktop

Manages the window collection:

- **Z-ordering**: maintains a front-to-back stack. Clicking a window brings it to front.
- **Tiling**: `Desktop.Tile()` arranges all windows in a non-overlapping grid.
- **Cascading**: `Desktop.Cascade()` arranges windows in a diagonal stack.
- **Background**: repeating pattern character fills empty desktop space (default: `░`).
- **Window switching**: Alt+1 through Alt+9 hotkeys, handled in the preprocess phase.

### Application

Top-level struct owning three children in fixed positions:

- **MenuBar** at row 0
- **Desktop** filling rows 1 through height-2
- **StatusLine** at the bottom row

On terminal resize, Application recalculates bounds for all three using their GrowModes, which cascades recursively through the view tree.

```go
app := tv.NewApplication(
    tv.WithMenuBar(menuBar),
    tv.WithStatusLine(statusLine),
    tv.WithTheme(theme.BorlandBlue),
)
app.Run()
```

---

## 4. Widgets

### Button

Renders with bracket notation: `[ OK ]`. Supports:
- Default button (responds to Enter via postprocess phase)
- Shortcut letters via `~O~K` tilde notation
- Shadow effect for 3D appearance
- Click or shortcut fires a command

### InputLine

Single-line text input with:
- Cursor and text selection
- Clipboard integration (copy/paste via OSC52 where supported; falls back to internal clipboard for terminals that block OSC52)
- Configurable max length
- Unicode-aware

### CheckBox

`[X]`/`[ ]` toggle. Grouped in a `CheckBoxes` cluster. Each item has a label with tilde shortcut.

### RadioButton

`(*)`/`( )` exclusive selection. Grouped in a `RadioButtons` cluster.

### Label

Static text with a tilde shortcut letter that transfers focus to a linked view:
```go
tv.NewLabel("~N~ame", nameInput)  // Alt+N focuses nameInput
```

### StaticText

Plain text display, word-wrapping within bounds.

### ScrollBar

Horizontal or vertical. Components: thumb, arrow buttons, page areas. Binds to a scrollable view and stays in sync. Supports:
- Mouse drag on thumb
- Click in page area to page up/down
- Click arrows to step
- Mouse wheel

### ListViewer

Abstract scrollable list with:
- Single selection with highlight
- Multiple columns
- Keyboard navigation (arrows, Home, End, PgUp, PgDn)
- Binds to a ScrollBar
- Concrete implementations supply the data via a data source interface

### Standard Dialogs

Built from the above widgets:

```go
type MsgBoxButton int

const (
    MbOK       MsgBoxButton = 1 << iota
    MbCancel
    MbYes
    MbNo
)

func MessageBox(title, text string, buttons MsgBoxButton) CommandCode
func InputBox(title, prompt, defaultValue string) (string, CommandCode)
```

`MessageBox` returns the `CommandCode` of the button pressed (`CmOK`, `CmCancel`, `CmYes`, `CmNo`). `InputBox` returns the entered string and a `CommandCode` — on cancel, the string is empty and the code is `CmCancel`.

Both functions create a `Dialog`, insert the appropriate widgets, and call `ExecView` on it. They are convenience functions, not separate types.

### Focus Traversal

Tab/Shift+Tab cycles through focusable views within a container. Tab order follows insertion order. This is automatic.

### Creating Custom Widgets

To create a custom widget, embed `BaseView` and implement the `Widget` interface:

```go
type MyWidget struct {
    tv.BaseView
    // custom fields
}

var _ tv.Widget = (*MyWidget)(nil)  // compile-time check

func NewMyWidget(bounds tv.Rect) *MyWidget {
    w := &MyWidget{}
    w.SetBounds(bounds)
    w.SetOptions(tv.OfSelectable, true)  // if it should receive focus
    return w
}

func (w *MyWidget) Draw(buf *tv.DrawBuffer) {
    scheme := w.ColorScheme()
    buf.Fill(tv.NewRect(0, 0, w.Bounds().Width(), w.Bounds().Height()), ' ', scheme.WindowBackground)
    // draw custom content
}

func (w *MyWidget) HandleEvent(event *tv.Event) {
    // handle events relevant to this widget
}
```

Since `Widget` is a marker interface (no additional methods beyond `View`), embedding `BaseView` provides all required method implementations. Override `Draw` and `HandleEvent` to add behavior.

---

## 5. Menu System

### MenuBar

Sits at row 0. Each item opens a pull-down menu. Activation: F10 or click. Arrow keys navigate between menus.

### Menu Items

Each item has:
- Label with tilde shortcut (`~F~ile`)
- Optional keyboard accelerator shown right-aligned (`Ctrl+S`)
- A command code fired on selection
- Optional submenu (nested)
- Separator lines between groups

### Context Menus

Any view can trigger a popup menu on right-click. Positioned at the mouse cursor, dismissed on click-outside or Escape.

### StatusLine

Bottom row. Displays context-sensitive key hints that change based on the focused view's help context. Each view can declare which status items are relevant when focused.

### HelpContext

Each view has an optional `HelpCtx` field (type `HelpContext`, which is `uint16`). The StatusLine queries the focused view's HelpCtx to decide which status items to display. Status items declare which HelpContext values they apply to. HelpContext 0 means "show always."

```go
type HelpContext uint16

const HcNoContext HelpContext = 0  // always visible
```

Views set their help context via `BaseView.SetHelpCtx(HelpContext)`. The StatusLine resolves the active context by reading `FocusedChild()` down the chain.

### Construction

Builder-style API. Keyboard accelerators use a `KeyBinding` type that pairs a `tcell.Key` (or rune) with modifier flags:

```go
type KeyBinding struct {
    Key  tcell.Key
    Rune rune        // for printable character shortcuts
    Mod  tcell.ModMask
}

func KbCtrl(ch rune) KeyBinding   // e.g., KbCtrl('N') = Ctrl+N
func KbAlt(ch rune) KeyBinding    // e.g., KbAlt('X') = Alt+X
func KbFunc(n int) KeyBinding     // e.g., KbFunc(10) = F10
func KbNone() KeyBinding          // no accelerator
```

Menu and status line construction:

```go
menuBar := tv.NewMenuBar(
    tv.NewSubMenu("~F~ile",
        tv.NewMenuItem("~N~ew", CmNew, tv.KbCtrl('N')),
        tv.NewMenuItem("~O~pen...", CmOpen, tv.KbCtrl('O')),
        tv.NewMenuSeparator(),
        tv.NewMenuItem("E~x~it", CmQuit, tv.KbAlt('X')),
    ),
    tv.NewSubMenu("~W~indow",
        tv.NewMenuItem("~T~ile", CmTile, tv.KbNone()),
        tv.NewMenuItem("~C~ascade", CmCascade, tv.KbNone()),
    ),
)

statusLine := tv.NewStatusLine(
    tv.NewStatusItem("~Alt+X~ Exit", tv.KbAlt('X'), CmQuit),
    tv.NewStatusItem("~F10~ Menu", tv.KbFunc(10), CmMenu),
)
```

---

## 6. Theme System

### ColorScheme

A flat struct with named fields for every semantic style. Each field is a `tcell.Style` (foreground + background + attributes), not a bare color, because every visual element needs both foreground and background to render:

```go
type ColorScheme struct {
    // Window
    WindowBackground    tcell.Style  // window client area
    WindowFrameActive   tcell.Style  // border when focused
    WindowFrameInactive tcell.Style  // border when not focused
    WindowTitle         tcell.Style  // title text in frame
    WindowShadow        tcell.Style  // shadow cells

    // Desktop
    DesktopBackground   tcell.Style  // desktop fill pattern

    // Dialog
    DialogBackground    tcell.Style  // dialog client area
    DialogFrame         tcell.Style  // dialog border

    // Widgets
    ButtonNormal        tcell.Style  // button text
    ButtonDefault       tcell.Style  // default button (responds to Enter)
    ButtonShadow        tcell.Style  // button shadow effect
    ButtonShortcut      tcell.Style  // highlighted shortcut letter
    InputNormal         tcell.Style  // input field text
    InputSelection      tcell.Style  // selected text in input
    LabelNormal         tcell.Style  // label text
    LabelShortcut       tcell.Style  // highlighted shortcut letter in label
    CheckBoxNormal      tcell.Style  // checkbox text
    RadioButtonNormal   tcell.Style  // radio button text
    ListNormal          tcell.Style  // list item text
    ListSelected        tcell.Style  // selected list item
    ListFocused         tcell.Style  // focused list item
    ScrollBar           tcell.Style  // scrollbar track
    ScrollThumb         tcell.Style  // scrollbar thumb

    // Menu
    MenuNormal          tcell.Style  // menu item text
    MenuShortcut        tcell.Style  // highlighted shortcut letter
    MenuSelected        tcell.Style  // highlighted menu item
    MenuDisabled        tcell.Style  // disabled menu item

    // Status
    StatusNormal        tcell.Style  // status bar background + text
    StatusShortcut      tcell.Style  // highlighted shortcut in status bar
}
```

Widgets look up the appropriate field directly from their resolved `ColorScheme`. No string-based lookup — the `View.ColorScheme()` method returns the nearest non-nil `*ColorScheme` in the owner chain, and widgets access fields by name: `scheme.ButtonNormal`, `scheme.InputSelection`, etc.

### Subtree Inheritance

Windows carry a `*ColorScheme` pointer, settable via `SetColorScheme(*ColorScheme)` on concrete types (Window, Dialog). `SetColorScheme` is not part of the `View` interface — scheme assignment is an explicit action on container types, not something arbitrary widgets do. If a view's scheme is nil, `ColorScheme()` walks up the owner chain to find the nearest non-nil scheme, ultimately falling back to the Application's default. This enables blue window next to gray window — same widget code, different schemes.

### Built-in Themes

Five themes ship with the framework:
- **BorlandBlue** — the classic blue-background Turbo Vision look
- **BorlandCyan** — cyan variant
- **BorlandGray** — gray/monochrome variant
- **Matrix** — green text on black background
- **C64** — dark blue background, red/yellow headings, cyan menu items (Fast Hack'em aesthetic)

### Theme Registry and User Overrides

Themes are registered by name. Apps register custom themes. Users can override via a JSON config file — JSON is used instead of TOML to avoid external dependencies (Go's standard library includes `encoding/json`). Partial overrides fall back to the base theme.

```go
theme.Register("midnight", myScheme)
app := tv.NewApplication(tv.WithTheme(theme.Get("c64")))
```

Config file location: `~/.config/turboview/theme.json` (follows XDG convention on Linux/macOS). Apps can override with `tv.WithConfigFile(path)`.

Config file example:
```json
{
  "base": "borland-blue",
  "overrides": {
    "WindowBackground": "#1a1a2e",
    "ButtonNormal": "#e94560"
  }
}
```

Override values are CSS-style hex color strings. For style fields that need both foreground and background, use `fg:bg` notation: `"#ffffff:#1a1a2e"`. Attributes can be appended: `"#ffffff:#1a1a2e:bold"`.

---

## 7. GrowMode and Resize

### GrowFlags

```go
const (
    GfGrowLoX  GrowFlag = 1 << iota  // left edge tracks owner's right edge
    GfGrowLoY                         // top edge tracks owner's bottom edge
    GfGrowHiX                         // right edge tracks owner's right edge
    GfGrowHiY                         // bottom edge tracks owner's bottom edge
    GfGrowAll  = GfGrowLoX | GfGrowLoY | GfGrowHiX | GfGrowHiY
    GfGrowRel                         // edges move proportionally to owner's size change
)
```

### Typical Usage

- OK button at bottom-right of dialog: `GfGrowLoX | GfGrowLoY`
- Input field that stretches: `GfGrowHiX`
- Desktop: `GfGrowAll`
- MenuBar: `GfGrowHiX` (stretches horizontally, stays at top)
- StatusLine: `GfGrowHiX | GfGrowLoY` (stretches horizontally, stays at bottom)
- Centered dialog that grows proportionally: `GfGrowRel` (if the owner doubles in size, the view's position and size scale proportionally — useful for views that should maintain their relative position and proportion within the owner)

### Resize Cascade

On terminal resize: Application receives the event → recalculates children's bounds based on GrowModes → each Container recursively does the same. One event, the whole tree adapts. No per-app resize handling needed.

---

## 8. Testing Strategy

### Unit Tests (primary weight)

Every widget and container is tested against `tcell.SimulationScreen`:

```go
sim := tcell.NewSimulationScreen("UTF-8")
sim.Init()
sim.SetSize(80, 25)
```

Test categories:
- **Drawing**: call `Draw()`, read cells back via `sim.GetContent(x, y)`, assert characters, colors, and styles
- **Event dispatch**: inject events via `sim.InjectKey()` / `sim.InjectMouse()`, assert state changes
- **Three-phase dispatch**: verify preprocess/focused/postprocess ordering
- **Focus traversal**: Tab cycles correctly, shortcuts land on the right widget
- **GrowMode**: resize the owner, assert children moved/stretched correctly
- **Modal execution**: `ExecView()` blocks, returns the right command code
- **Theme inheritance**: set scheme on a window, verify children resolve colors correctly

The `View` interface enables clean mocking. Test a Dialog without a real Application — give it a mock Container and a SimulationScreen.

### Shared Test Helpers

A `helpers_test.go` file in the `tv` package provides common test utilities:

```go
func newTestScreen(w, h int) tcell.SimulationScreen  // creates and initializes a SimulationScreen
func newTestBuffer(w, h int) *DrawBuffer              // creates a DrawBuffer for testing
func newTestApp(t *testing.T) *Application            // creates an Application with SimulationScreen
func injectKey(sim tcell.SimulationScreen, key tcell.Key, r rune, mod tcell.ModMask)
func assertCell(t *testing.T, buf *DrawBuffer, x, y int, ch rune, style tcell.Style)
```

All task implementations share these helpers. The first task (foundation-types-rect) creates this file; subsequent tasks add to it as needed.

### E2e Smoke Tests (secondary layer)

A tmux-based test harness for integration testing:

1. Start a tmux session
2. Launch a turboview test app inside it
3. Send keystrokes via `tmux send-keys`
4. Capture terminal content via `tmux capture-pane -p`
5. Assert on text content

```go
// Dialog appears
tmuxSendKeys(t, session, "F2")
output := tmuxCapture(t, session)
assert.Contains(t, output[8], "[ OK ]")

// Dialog disappears after Enter
tmuxSendKeys(t, session, "Enter")
output = tmuxCapture(t, session)
assert.NotContains(t, output[8], "[ OK ]")
assert.Contains(t, output[8], "░")  // desktop background restored
```

E2e tests cover: app boots, menus open/close, dialogs appear/disappear, window switching works, resize produces valid layout, app exits cleanly. Roughly 20 scenarios vs hundreds of unit tests.

Small purpose-built test binaries live in `e2e/testapp/`.

### What Is Tested Where

| Concern | Unit (SimulationScreen) | E2e (tmux) |
|---|---|---|
| Drawing correctness | Yes — cell-by-cell | No |
| Colors and styles | Yes — via GetContent() style | No |
| Event dispatch | Yes | Indirectly |
| Focus traversal | Yes | Yes (smoke) |
| Modal dialogs | Yes | Yes (smoke) |
| Terminal resize | Yes (simulated) | Yes (tmux resize-pane) |
| Real terminal rendering | No | Yes |
| Escape sequence handling | No | Yes |

---

## 9. Package Structure

```
turboview/
  tv/                  — core package: all views, widgets, events, menus
    base_view.go       — BaseView struct, View/Container/Widget interfaces
    draw_buffer.go     — DrawBuffer, Cell, clipping, sub-buffer creation
    focus.go           — focus chain management, traversal logic
    group.go           — Group (Container implementation)
    window.go          — Window, Frame
    dialog.go          — Dialog, MessageBox, InputBox
    desktop.go         — Desktop, z-ordering, tile, cascade
    application.go     — Application lifecycle
    event.go           — Event types, dispatch logic
    menu.go            — MenuBar, MenuItem, MenuPopup
    status.go          — StatusLine, StatusItem
    button.go          — Button widget
    input.go           — InputLine widget
    checkbox.go        — CheckBox, CheckBoxes cluster
    radio.go           — RadioButton, RadioButtons cluster
    label.go           — Label widget
    static_text.go     — StaticText widget
    scrollbar.go       — ScrollBar widget
    list_viewer.go     — ListViewer widget
    rect.go            — Rect, Point geometry types
    grow.go            — GrowMode flags and resize logic
    command.go         — CommandCode constants
    *_test.go          — unit tests alongside each file
  theme/
    scheme.go          — ColorScheme struct
    registry.go        — theme registration and lookup
    borland.go         — BorlandBlue, BorlandCyan, BorlandGray
    matrix.go          — Matrix theme
    c64.go             — C64 theme
    config.go          — TOML config loading and override logic
    *_test.go
  e2e/
    e2e_test.go        — tmux-based integration tests
    harness.go         — tmux session management helpers
    testapp/
      basic/main.go    — simple app: window + menu + status
      dialogs/main.go  — dialog and modal test app
      multi/main.go    — multiple windows, tiling, cascading
```

---

## 10. Rendering Pipeline

### DrawBuffer

Views never draw directly to `tcell.Screen`. Instead, each view receives a `*DrawBuffer` — a bounded cell buffer that enforces clipping:

```go
type DrawBuffer struct {
    cells  [][]Cell       // the underlying cell grid
    clip   Rect           // writable region (in buffer-local coordinates)
    offset Point          // translation from buffer-local to screen-absolute
}

type Cell struct {
    Rune  rune
    Combc []rune       // combining characters (e.g., accents, diacritics)
    Style tcell.Style
}
```

`DrawBuffer` provides drawing primitives:
- `WriteChar(x, y int, ch rune, style tcell.Style)` — write a single cell (no-op if outside clip)
- `WriteStr(x, y int, s string, style tcell.Style)` — write a string horizontally
- `Fill(r Rect, ch rune, style tcell.Style)` — fill a rectangle
- `SubBuffer(r Rect) *DrawBuffer` — create a child buffer with a narrower clip region and adjusted offset

SubBuffer shares the parent's backing `cells` grid — writes through a SubBuffer are immediately visible in the parent. No copying occurs. The SubBuffer's `clip` is the intersection of the parent's clip and the requested `Rect`, so a child cannot draw outside its allocated region even if it tries.

### Draw Sequence

1. `Application` creates a root `DrawBuffer` sized to the terminal. This is an in-memory cell grid; it is flushed to `tcell.Screen` after the draw pass completes.
2. `Application.Draw()` calls each of its three children (MenuBar, Desktop, StatusLine) with a `SubBuffer` clipped to their bounds.
3. `Desktop.Draw()` iterates windows back-to-front. For each window, it creates a `SubBuffer` clipped to the window's bounds.
4. `Window.Draw()` draws its frame into its buffer, then creates a `SubBuffer` for its client area (interior of the frame) and calls each child's `Draw()` with a further-clipped sub-buffer.
5. This recurses until leaf widgets draw their content.
6. After the full draw pass, `Application` flushes the buffer to `tcell.Screen` via `Show()`.

Overlapping windows work correctly because each window draws only within its clipped region, and windows are drawn back-to-front so the frontmost window's cells overwrite earlier ones.

### Shadow Rendering

Window shadows are drawn by the Desktop during its draw pass, after drawing each window. For each shadow cell, the Desktop reads the existing cell from the buffer, replaces its style with `ColorScheme.WindowShadow`, and writes it back — preserving the character but darkening the appearance. This happens at the Desktop level, not inside the Window, because the shadow extends outside the window's own bounds.

### Flushing to Screen

`DrawBuffer` provides a `FlushTo(screen tcell.Screen)` method that copies its cell grid to the tcell screen. Only `Application` calls this — after the full draw pass completes, it calls `rootBuffer.FlushTo(screen)` followed by `screen.Show()`. No other type should call `FlushTo`.

---

## 11. Focus Model

### Focusability

A view is focusable if it has `OfSelectable` in its options. Non-selectable views (labels, static text, frames) are skipped during focus traversal.

### Focus Chain

Within a Container, the focus chain is the ordered list of `OfSelectable` children, in insertion order. `FocusedChild()` returns the currently focused child (the one with `SfSelected` state).

### Focus Traversal

Tab advances to the next selectable sibling. Shift+Tab goes to the previous. When the end of the list is reached, it wraps around. The Container handles this in its `HandleEvent` — it intercepts Tab/Shift+Tab before forwarding to the focused child.

### Focus Delegation

When a Container receives focus (`SfFocused` set to true), it delegates focus to its first selectable child (or the previously selected child if one exists). This cascades: if that child is also a Container, it delegates to *its* first selectable child, until a leaf Widget receives focus.

### Focus and Selection

Two distinct states:
- `SfSelected` — this view is the active child of its owner (the one that gets events when the owner has focus). Exactly one child per Container is selected at a time.
- `SfFocused` — this view (and its entire owner chain up to the Application) currently has input focus. Only one leaf view in the entire application is focused at a time.

A view can be selected but not focused (its owner doesn't have focus). When its owner gains focus, the selected child becomes focused.

---

## 12. Application Event Loop

### `Application.Run()`

```
1. Initialize tcell.Screen (enter raw mode, alt screen, enable mouse)
2. Draw the full view tree (initial paint)
3. Loop:
   a. Poll tcell.Screen.PollEvent()
   b. Convert tcell event to tv.Event
   c. Route the event:
      - Resize → recalculate all bounds via GrowMode cascade, full redraw
      - Mouse → find the topmost view at the mouse coordinates (walk Desktop's
        window stack front-to-back, then recurse into the hit window's children).
        Translate coordinates to local space. Deliver to the target view.
      - Key → deliver to the focus chain via three-phase dispatch
   d. Redraw. For v1: full redraw every frame (simple, correct). Optimization
      via dirty flags and partial redraws can be added later without API changes.
   e. Flush DrawBuffer to tcell.Screen via Show()
   f. If CmQuit was issued, break
4. Finalize tcell.Screen (leave alt screen, restore terminal)
```

### Error Handling

The framework does not use sentinel errors or error wrapping beyond what the standard library provides. The conventions:

- `NewApplication` returns `(*Application, error)` — the only fallible constructor. The error comes from `tcell.Screen.Init()`.
- `Application.Run()` returns `error` if the event loop terminates abnormally (screen error). Normal termination via `CmQuit` returns nil.
- `ExecView(nil)` panics. `ExecView` on an already-modal view panics. These are programming errors, not runtime conditions.
- Widget methods do not return errors. Invalid operations (writing outside clip, removing a non-child) are silent no-ops.

### Async Command Injection

```go
func (app *Application) PostCommand(cmd CommandCode, info any)
```

`PostCommand` enqueues a command event that will be delivered on the next iteration of the event loop. It is safe to call from any goroutine. This is the only thread-safe method on Application — all other view operations must happen on the event loop goroutine.

Use case: a background goroutine completing I/O posts `CmUser+N` to notify the UI.

### Modal Event Loop

`ExecView(view)` works by:

1. Inserting the modal view into the Container
2. Setting `SfModal` on the view
3. Obtaining an event source: Group.ExecView() walks up the Owner() chain to find the Application, which implements the `EventSource` interface: `PollEvent() *Event`. This allows Group to poll for events without owning the screen directly.
4. Entering a nested event loop: poll from EventSource → dispatch to modal view only → check if closed → repeat
5. In the nested loop, all focused events are routed only to the modal view and its children. Positional events outside the modal view are discarded (or optionally beep).

```go
type EventSource interface {
    PollEvent() *Event
}
```

Application implements EventSource. In tests, a mock EventSource can feed canned events, making ExecView fully testable without a real terminal.
5. When the modal view issues a closing command (`CmOK`, `CmCancel`, `CmClose`), the nested loop exits
6. The modal view is removed, `SfModal` cleared, and the command code is returned

This nesting means modal dialogs can open other modal dialogs — the loops stack naturally.

### Menu Popup Lifecycle

Pull-down menus and context menus are implemented as modal views. When a MenuBar item is clicked:

1. A `MenuPopup` view is created at the appropriate screen position
2. `ExecView(menuPopup)` enters a modal loop
3. The popup handles keyboard navigation (arrows, Enter, Escape) and mouse clicks
4. On selection, the popup closes with the selected command, which is then broadcast
5. On Escape or click-outside, the popup closes with no command

This means menus participate in the normal view hierarchy and event dispatch — no special overlay system needed.

---

## 13. Dependencies

- **tcell** (`github.com/gdamore/tcell/v2`) — terminal abstraction, screen, events, SimulationScreen for tests
- **Go standard library** — everything else
- No other external dependencies for the core framework
- **tmux** — required only for e2e tests (not a Go dependency)
