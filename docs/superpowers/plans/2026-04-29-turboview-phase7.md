# Phase 7: Themes, HelpContext StatusLine, Theme Config

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add four additional built-in themes (BorlandCyan, BorlandGray, Matrix, C64), context-sensitive StatusLine filtering via HelpContext, and JSON theme configuration with user overrides.

**Architecture:** Each theme is a self-registering file in the theme package using `init()`. StatusLine filters visible items by matching their HelpCtx against the active focused view's HelpCtx, resolved by walking the focus chain in Application. Theme config loads a JSON file, clones a registered base theme, and applies field-level style overrides via reflection.

**Tech Stack:** Go 1.22+, tcell/v2, encoding/json, reflect, os

---

## File Structure

### New Files
- `theme/borland_cyan.go` — BorlandCyan theme definition and registration
- `theme/borland_gray.go` — BorlandGray monochrome theme definition and registration
- `theme/matrix.go` — Matrix green-on-black theme definition and registration
- `theme/c64.go` — C64 retro theme definition and registration
- `theme/config.go` — JSON config loading, style string parsing, default config path resolution

### Modified Files
- `tv/status.go` — Add `ForHelpCtx` builder on StatusItem, `activeCtx` field on StatusLine, filtering in Draw and HandleEvent
- `tv/application.go` — Add `resolveHelpCtx()` method, `WithConfigFile` option, config file loading in NewApplication, set active context before drawing StatusLine
- `e2e/testapp/basic/main.go` — Set HelpCtx on windows, add context-specific status items
- `e2e/e2e_test.go` — New e2e tests for HelpContext filtering and theme registration

---

### Task 1: Four Additional Themes

**Files:**
- Create: `theme/borland_cyan.go`
- Create: `theme/borland_gray.go`
- Create: `theme/matrix.go`
- Create: `theme/c64.go`

**Requirements:**
- `BorlandCyan` is an exported `*ColorScheme` variable, initialized in `init()`, and registered as `"borland-cyan"` — all 29 style fields are set (non-zero)
- `BorlandGray` is an exported `*ColorScheme` variable, initialized in `init()`, and registered as `"borland-gray"` — all 29 style fields are set (non-zero)
- `Matrix` is an exported `*ColorScheme` variable, initialized in `init()`, and registered as `"matrix"` — all 29 style fields are set (non-zero)
- `C64` is an exported `*ColorScheme` variable, initialized in `init()`, and registered as `"c64"` — all 29 style fields are set (non-zero)
- `theme.Get("borland-cyan")` returns the same pointer as `theme.BorlandCyan`
- `theme.Get("borland-gray")` returns the same pointer as `theme.BorlandGray`
- `theme.Get("matrix")` returns the same pointer as `theme.Matrix`
- `theme.Get("c64")` returns the same pointer as `theme.C64`
- Each theme has visually distinct colors from BorlandBlue (different foreground or background on at least WindowBackground, DesktopBackground, and MenuNormal)
- The registry now contains exactly 5 themes: `"borland-blue"`, `"borland-cyan"`, `"borland-gray"`, `"matrix"`, `"c64"`

**Implementation:**

Each file follows the exact same pattern as `theme/borland.go`. Use the `s` helper for `tcell.StyleDefault.Foreground(fg).Background(bg)`.

`theme/borland_cyan.go`:
```go
package theme

import "github.com/gdamore/tcell/v2"

var BorlandCyan *ColorScheme

func init() {
	s := func(fg, bg tcell.Color) tcell.Style {
		return tcell.StyleDefault.Foreground(fg).Background(bg)
	}

	BorlandCyan = &ColorScheme{
		WindowBackground:    s(tcell.ColorWhite, tcell.ColorTeal),
		WindowFrameActive:   s(tcell.ColorYellow, tcell.ColorTeal),
		WindowFrameInactive: s(tcell.ColorSilver, tcell.ColorTeal),
		WindowTitle:         s(tcell.ColorWhite, tcell.ColorTeal),
		WindowShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		DesktopBackground:   s(tcell.ColorTeal, tcell.ColorDarkCyan),
		DialogBackground:    s(tcell.ColorBlack, tcell.ColorSilver),
		DialogFrame:         s(tcell.ColorWhite, tcell.ColorSilver),
		ButtonNormal:        s(tcell.ColorBlack, tcell.ColorGreen),
		ButtonDefault:       s(tcell.ColorWhite, tcell.ColorGreen),
		ButtonShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		ButtonShortcut:      s(tcell.ColorYellow, tcell.ColorGreen),
		InputNormal:         s(tcell.ColorBlack, tcell.ColorWhite),
		InputSelection:      s(tcell.ColorWhite, tcell.ColorBlue),
		LabelNormal:         s(tcell.ColorBlack, tcell.ColorSilver),
		LabelShortcut:       s(tcell.ColorYellow, tcell.ColorSilver),
		CheckBoxNormal:      s(tcell.ColorBlack, tcell.ColorSilver),
		RadioButtonNormal:   s(tcell.ColorBlack, tcell.ColorSilver),
		ListNormal:          s(tcell.ColorBlack, tcell.ColorSilver),
		ListSelected:        s(tcell.ColorWhite, tcell.ColorBlack),
		ListFocused:         s(tcell.ColorYellow, tcell.ColorTeal),
		ScrollBar:           s(tcell.ColorSilver, tcell.ColorTeal),
		ScrollThumb:         s(tcell.ColorWhite, tcell.ColorTeal),
		MenuNormal:          s(tcell.ColorBlack, tcell.ColorSilver),
		MenuShortcut:        s(tcell.ColorRed, tcell.ColorSilver),
		MenuSelected:        s(tcell.ColorWhite, tcell.ColorBlack),
		MenuDisabled:        s(tcell.ColorGray, tcell.ColorSilver),
		StatusNormal:        s(tcell.ColorBlack, tcell.ColorSilver),
		StatusShortcut:      s(tcell.ColorYellow, tcell.ColorSilver),
	}

	Register("borland-cyan", BorlandCyan)
}
```

`theme/borland_gray.go`:
```go
package theme

import "github.com/gdamore/tcell/v2"

var BorlandGray *ColorScheme

func init() {
	s := func(fg, bg tcell.Color) tcell.Style {
		return tcell.StyleDefault.Foreground(fg).Background(bg)
	}

	BorlandGray = &ColorScheme{
		WindowBackground:    s(tcell.ColorWhite, tcell.ColorDarkGray),
		WindowFrameActive:   s(tcell.ColorWhite, tcell.ColorDarkGray),
		WindowFrameInactive: s(tcell.ColorSilver, tcell.ColorDarkGray),
		WindowTitle:         s(tcell.ColorWhite, tcell.ColorDarkGray),
		WindowShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		DesktopBackground:   s(tcell.ColorSilver, tcell.ColorGray),
		DialogBackground:    s(tcell.ColorBlack, tcell.ColorSilver),
		DialogFrame:         s(tcell.ColorWhite, tcell.ColorSilver),
		ButtonNormal:        s(tcell.ColorBlack, tcell.ColorSilver),
		ButtonDefault:       s(tcell.ColorWhite, tcell.ColorDarkGray),
		ButtonShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		ButtonShortcut:      s(tcell.ColorYellow, tcell.ColorSilver),
		InputNormal:         s(tcell.ColorBlack, tcell.ColorWhite),
		InputSelection:      s(tcell.ColorWhite, tcell.ColorDarkGray),
		LabelNormal:         s(tcell.ColorBlack, tcell.ColorSilver),
		LabelShortcut:       s(tcell.ColorYellow, tcell.ColorSilver),
		CheckBoxNormal:      s(tcell.ColorBlack, tcell.ColorSilver),
		RadioButtonNormal:   s(tcell.ColorBlack, tcell.ColorSilver),
		ListNormal:          s(tcell.ColorBlack, tcell.ColorSilver),
		ListSelected:        s(tcell.ColorWhite, tcell.ColorBlack),
		ListFocused:         s(tcell.ColorYellow, tcell.ColorDarkGray),
		ScrollBar:           s(tcell.ColorSilver, tcell.ColorDarkGray),
		ScrollThumb:         s(tcell.ColorWhite, tcell.ColorDarkGray),
		MenuNormal:          s(tcell.ColorBlack, tcell.ColorSilver),
		MenuShortcut:        s(tcell.ColorRed, tcell.ColorSilver),
		MenuSelected:        s(tcell.ColorWhite, tcell.ColorBlack),
		MenuDisabled:        s(tcell.ColorGray, tcell.ColorSilver),
		StatusNormal:        s(tcell.ColorBlack, tcell.ColorSilver),
		StatusShortcut:      s(tcell.ColorYellow, tcell.ColorSilver),
	}

	Register("borland-gray", BorlandGray)
}
```

`theme/matrix.go`:
```go
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
		LabelShortcut:       s(tcell.ColorLime, tcell.ColorBlack),
		CheckBoxNormal:      s(tcell.ColorGreen, tcell.ColorBlack),
		RadioButtonNormal:   s(tcell.ColorGreen, tcell.ColorBlack),
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
	}

	Register("matrix", Matrix)
}
```

`theme/c64.go`:
```go
package theme

import "github.com/gdamore/tcell/v2"

var C64 *ColorScheme

func init() {
	s := func(fg, bg tcell.Color) tcell.Style {
		return tcell.StyleDefault.Foreground(fg).Background(bg)
	}

	C64 = &ColorScheme{
		WindowBackground:    s(tcell.ColorWhite, tcell.ColorNavy),
		WindowFrameActive:   s(tcell.ColorYellow, tcell.ColorNavy),
		WindowFrameInactive: s(tcell.ColorSilver, tcell.ColorNavy),
		WindowTitle:         s(tcell.ColorYellow, tcell.ColorNavy),
		WindowShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		DesktopBackground:   s(tcell.ColorBlue, tcell.ColorNavy),
		DialogBackground:    s(tcell.ColorWhite, tcell.ColorPurple),
		DialogFrame:         s(tcell.ColorYellow, tcell.ColorPurple),
		ButtonNormal:        s(tcell.ColorWhite, tcell.ColorRed),
		ButtonDefault:       s(tcell.ColorYellow, tcell.ColorRed),
		ButtonShadow:        s(tcell.ColorBlack, tcell.ColorBlack),
		ButtonShortcut:      s(tcell.ColorYellow, tcell.ColorRed),
		InputNormal:         s(tcell.ColorBlue, tcell.ColorWhite),
		InputSelection:      s(tcell.ColorWhite, tcell.ColorBlue),
		LabelNormal:         s(tcell.ColorWhite, tcell.ColorPurple),
		LabelShortcut:       s(tcell.ColorYellow, tcell.ColorPurple),
		CheckBoxNormal:      s(tcell.ColorWhite, tcell.ColorPurple),
		RadioButtonNormal:   s(tcell.ColorWhite, tcell.ColorPurple),
		ListNormal:          s(tcell.ColorTeal, tcell.ColorNavy),
		ListSelected:        s(tcell.ColorWhite, tcell.ColorBlue),
		ListFocused:         s(tcell.ColorYellow, tcell.ColorBlue),
		ScrollBar:           s(tcell.ColorBlue, tcell.ColorNavy),
		ScrollThumb:         s(tcell.ColorTeal, tcell.ColorNavy),
		MenuNormal:          s(tcell.ColorTeal, tcell.ColorNavy),
		MenuShortcut:        s(tcell.ColorYellow, tcell.ColorNavy),
		MenuSelected:        s(tcell.ColorNavy, tcell.ColorTeal),
		MenuDisabled:        s(tcell.ColorGray, tcell.ColorNavy),
		StatusNormal:        s(tcell.ColorWhite, tcell.ColorMaroon),
		StatusShortcut:      s(tcell.ColorYellow, tcell.ColorMaroon),
	}

	Register("c64", C64)
}
```

**Run tests:** `go test ./theme/... -v`

**Commit:** `git commit -m "feat: add BorlandCyan, BorlandGray, Matrix, and C64 themes"`

---

### Task 2: StatusLine HelpContext Filtering

**Files:**
- Modify: `tv/status.go`
- Modify: `tv/application.go`

**Requirements:**

StatusItem builder:
- `NewStatusItem("label", kb, cmd).ForHelpCtx(5)` sets the item's HelpCtx to 5 and returns the same `*StatusItem` for chaining
- `NewStatusItem("label", kb, cmd)` without `ForHelpCtx` has HelpCtx == HcNoContext (0)

StatusLine filtering in Draw:
- Items with `HelpCtx == HcNoContext` (0) are always drawn regardless of active context
- Items with a non-zero HelpCtx are drawn only when `StatusLine.activeCtx` matches their HelpCtx
- When some items are filtered out, the remaining items are drawn with correct spacing (no gaps where filtered items would have been)
- When `activeCtx` is HcNoContext (0), only items with HelpCtx == HcNoContext are shown (this is the default "no focused view" state)

StatusLine filtering in HandleEvent:
- Only items that would be drawn (passing the same HelpCtx filter) can match keybindings
- An item with `HelpCtx == 5` and `activeCtx == 3` does not match its keybinding
- An item with `HelpCtx == HcNoContext` always matches its keybinding

StatusLine.SetActiveContext:
- `SetActiveContext(hc HelpContext)` sets the `activeCtx` field
- Multiple calls overwrite the previous value

Application.resolveHelpCtx:
- With no desktop or no focused views, returns HcNoContext
- Walks the focus chain: Desktop → focused window → focused child → ... until leaf
- Returns the deepest (most specific) non-zero HelpCtx found in the chain
- If a Window has HelpCtx=1 and its focused Button has HelpCtx=5, returns 5
- If a Window has HelpCtx=1 and its focused Button has HelpCtx=0, returns 1
- If no view in the chain has a non-zero HelpCtx, returns HcNoContext

Application.Draw integration:
- Before drawing the StatusLine, Application calls `resolveHelpCtx()` and passes the result to `StatusLine.SetActiveContext()`
- This happens on every draw cycle so context-sensitive items update when focus changes

**Implementation:**

Add to `tv/status.go`:

```go
func (si *StatusItem) ForHelpCtx(hc HelpContext) *StatusItem {
	si.HelpCtx = hc
	return si
}
```

Add `activeCtx` field and setter to StatusLine:

```go
type StatusLine struct {
	BaseView
	items     []*StatusItem
	activeCtx HelpContext
}

func (sl *StatusLine) SetActiveContext(hc HelpContext) {
	sl.activeCtx = hc
}
```

Replace the existing `Draw` method with HelpCtx filtering:

```go
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
	first := true
	for _, item := range sl.items {
		if item.HelpCtx != HcNoContext && item.HelpCtx != sl.activeCtx {
			continue
		}
		if !first {
			x += 2
		}
		first = false
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
```

Replace the existing `HandleEvent` method with HelpCtx filtering:

```go
func (sl *StatusLine) HandleEvent(event *Event) {
	if event.What != EvKeyboard || event.Key == nil {
		return
	}
	for _, item := range sl.items {
		if item.HelpCtx != HcNoContext && item.HelpCtx != sl.activeCtx {
			continue
		}
		if item.KeyBinding.Matches(event.Key) {
			event.What = EvCommand
			event.Command = item.Command
			event.Key = nil
			return
		}
	}
}
```

Add to `tv/application.go`:

```go
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
```

Modify `Application.Draw` to set active context before drawing StatusLine — change the StatusLine drawing block:

```go
	if app.statusLine != nil && h > 0 {
		app.statusLine.SetActiveContext(app.resolveHelpCtx())
		statusBuf := buf.SubBuffer(NewRect(0, h-1, w, 1))
		app.statusLine.Draw(statusBuf)
	}
```

**Run tests:** `go test ./tv/... -v`

**Commit:** `git commit -m "feat: add HelpContext filtering to StatusLine"`

---

### Task 3: Theme JSON Config Loading

**Files:**
- Create: `theme/config.go`
- Modify: `tv/application.go`

**Requirements:**

ParseStyleString:
- `ParseStyleString("#ff0000")` returns a style with foreground set to RGB(255,0,0) and default background
- `ParseStyleString("#ffffff:#000000")` returns a style with foreground RGB(255,255,255) and background RGB(0,0,0)
- `ParseStyleString("#ffffff:#000000:bold")` returns a style with fg, bg, and bold attribute
- `ParseStyleString("#ffffff:#000000:bold,underline")` returns a style with fg, bg, bold, and underline attributes
- `ParseStyleString("#xyz")` returns an error (invalid hex)
- `ParseStyleString("")` returns a default style with no error (empty string = no override)
- Hex values are case-insensitive: `"#FF0000"` and `"#ff0000"` produce the same color

LoadConfig:
- Given a valid JSON file with `{"base": "borland-blue", "overrides": {"WindowBackground": "#00ff00"}}`, returns a `*ColorScheme` cloned from BorlandBlue with WindowBackground's foreground set to green
- Given `{"base": "borland-blue", "overrides": {"WindowBackground": "#00ff00:#000000"}}`, returns a scheme with WindowBackground foreground=green, background=black
- Given `{"base": "nonexistent"}`, returns an error mentioning "unknown base theme"
- Given `{"base": "borland-blue", "overrides": {"FakeField": "#000000"}}`, returns an error mentioning "unknown color scheme field"
- Given an invalid JSON file, returns an error
- Given a path to a non-existent file, returns `(nil, nil)` — file-not-found is silently handled, not an error
- All non-overridden fields retain the base theme's values

DefaultConfigPath:
- Returns `$HOME/.config/turboview/theme.json` where `$HOME` is the user's home directory
- Returns empty string if the home directory cannot be determined

WithConfigFile:
- `tv.WithConfigFile("/path/to/config.json")` sets a custom config path on the Application
- When set, overrides the default config path

Application config loading:
- If a config file exists (at custom path or default path), the scheme from that file becomes the Application's theme, overriding WithTheme
- If no config file exists at either path, the Application uses WithTheme or the BorlandBlue default — no error
- If a config file exists but contains invalid JSON, NewApplication returns an error
- Config loading happens after WithTheme is applied, so a valid config file takes precedence

**Implementation:**

Create `theme/config.go`:

```go
package theme

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type Config struct {
	Base      string            `json:"base"`
	Overrides map[string]string `json:"overrides"`
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "turboview", "theme.json")
}

func LoadConfig(path string) (*ColorScheme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid theme config: %w", err)
	}

	base := Get(cfg.Base)
	if base == nil {
		return nil, fmt.Errorf("unknown base theme: %q", cfg.Base)
	}

	scheme := *base

	rv := reflect.ValueOf(&scheme).Elem()
	rt := rv.Type()
	for fieldName, styleStr := range cfg.Overrides {
		found := false
		for i := 0; i < rt.NumField(); i++ {
			if rt.Field(i).Name == fieldName {
				style, err := ParseStyleString(styleStr)
				if err != nil {
					return nil, fmt.Errorf("invalid style for %s: %w", fieldName, err)
				}
				rv.Field(i).Set(reflect.ValueOf(style))
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown color scheme field: %q", fieldName)
		}
	}

	return &scheme, nil
}

func ParseStyleString(s string) (tcell.Style, error) {
	if s == "" {
		return tcell.StyleDefault, nil
	}

	parts := strings.SplitN(s, ":", 3)
	style := tcell.StyleDefault

	if len(parts) >= 1 && parts[0] != "" {
		fg, err := parseHexColor(parts[0])
		if err != nil {
			return style, fmt.Errorf("foreground: %w", err)
		}
		style = style.Foreground(fg)
	}

	if len(parts) >= 2 && parts[1] != "" {
		bg, err := parseHexColor(parts[1])
		if err != nil {
			return style, fmt.Errorf("background: %w", err)
		}
		style = style.Background(bg)
	}

	if len(parts) >= 3 {
		attrs := strings.ToLower(parts[2])
		for _, attr := range strings.Split(attrs, ",") {
			switch strings.TrimSpace(attr) {
			case "bold":
				style = style.Bold(true)
			case "underline":
				style = style.Underline(true)
			case "italic":
				style = style.Italic(true)
			case "reverse":
				style = style.Reverse(true)
			}
		}
	}

	return style, nil
}

func parseHexColor(s string) (tcell.Color, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return tcell.ColorDefault, fmt.Errorf("invalid hex color %q: must be 6 hex digits", "#"+s)
	}
	s = strings.ToLower(s)
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid hex color %q: %w", "#"+s, err)
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid hex color %q: %w", "#"+s, err)
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid hex color %q: %w", "#"+s, err)
	}
	return tcell.NewRGBColor(int32(r), int32(g), int32(b)), nil
}
```

Modify `tv/application.go` — add `configFile` field and `WithConfigFile` option:

```go
type Application struct {
	bounds     Rect
	screen     tcell.Screen
	screenOwn  bool
	desktop    *Desktop
	statusLine *StatusLine
	menuBar    *MenuBar
	scheme     *theme.ColorScheme
	quit       bool
	onCommand  func(CommandCode, any) bool
	configFile string
}

func WithConfigFile(path string) AppOption {
	return func(app *Application) {
		app.configFile = path
	}
}
```

In `NewApplication`, add config loading after the scheme default is set (after `if app.scheme == nil { app.scheme = theme.BorlandBlue }`):

```go
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
```

**Run tests:** `go test ./theme/... ./tv/... -v`

**Commit:** `git commit -m "feat: add JSON theme config loading with user overrides"`

---

### Task 4: Integration Checkpoint — HelpContext + Config + Themes

**Purpose:** Verify that HelpContext filtering works through the real focus chain, that config-loaded themes render correctly, and that all themes are accessible through the registry.

**Requirements (for test writer):**
- An Application with a StatusLine containing both HcNoContext items and context-specific items, plus two Windows with different HelpCtx values: when focus is on window1 (HelpCtx=1), only HcNoContext items and HelpCtx=1 items are drawn; when focus switches to window2 (HelpCtx=2), only HcNoContext items and HelpCtx=2 items are drawn
- The resolveHelpCtx walk finds the deepest non-zero HelpCtx: if a Window has HelpCtx=1 and its focused child (Button) has HelpCtx=5, resolveHelpCtx returns 5
- All 5 registered themes (borland-blue, borland-cyan, borland-gray, matrix, c64) can be passed to WithTheme and the Application draws with that scheme's styles
- A config file that specifies `"base": "matrix"` with an override on `"WindowBackground"` produces a scheme that has Matrix's values for all fields except WindowBackground
- StatusLine keybinding filtering: pressing a key bound to a context-specific item that is NOT in the active context does NOT fire its command

**Components to wire up:** Application (with SimulationScreen), Desktop, Window, Button, StatusLine (all real, no mocks)

**Run:** `go test ./tv/... -run TestIntegration -v`

---

### Task 5: E2E Test

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**

Demo app changes:
- win1 has `SetHelpCtx(1)` (file manager context)
- win2 has `SetHelpCtx(2)` (editor context)
- StatusLine includes at least one context-specific item: `"~F2~ Dialog"` with HelpCtx=1 (only visible when win1 is focused)
- StatusLine includes an editor-specific item: `"~F4~ Search"` with HelpCtx=2 (only visible when win2 is focused) — this item can map to a no-op command (CmUser+20)
- Items "~Alt+X~ Exit" and "~F10~ Menu" remain HcNoContext (always visible)

E2E tests:
- TestHelpContextFiltering: Boot app → capture status line → verify "Alt+X" is visible and "F2" or "Dialog" is visible (win1 focused by default) → Tab to switch to win2 → capture status line → verify "Dialog" is NOT visible and "Search" IS visible → Alt+X to exit cleanly
- TestThemeRegistration: A test that verifies the built binary can boot with the default theme (BorlandBlue) and render correctly (already covered by existing tests, but this test checks that importing theme doesn't break anything — effectively a smoke test that the new theme files compile and register without init() panics)

**Implementation:**

Update `e2e/testapp/basic/main.go` — modify StatusLine construction:

```go
	statusLine := tv.NewStatusLine(
		tv.NewStatusItem("~Alt+X~ Exit", tv.KbAlt('X'), tv.CmQuit),
		tv.NewStatusItem("~F2~ Dialog", tv.KbFunc(2), tv.CmUser).ForHelpCtx(1),
		tv.NewStatusItem("~F3~ Input", tv.KbFunc(3), tv.CmUser+10).ForHelpCtx(1),
		tv.NewStatusItem("~F4~ Search", tv.KbFunc(4), tv.CmUser+20).ForHelpCtx(2),
		tv.NewStatusItem("~F10~ Menu", tv.KbFunc(10), tv.CmMenu),
	)
```

Add HelpCtx to windows (after window construction):

```go
	win1.SetHelpCtx(1)
	win2.SetHelpCtx(2)
```

Add the `TestHelpContextFiltering` and `TestThemeRegistration` e2e tests to `e2e/e2e_test.go`.

**Run tests:** `cd /Users/gnat/Source/Personal/tv3 && go test ./e2e/... -v -timeout 120s`

**Commit:** `git commit -m "feat: add e2e tests for HelpContext filtering and themes"`
