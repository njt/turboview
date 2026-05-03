# TurboView

A Go TUI framework that faithfully reimplements Borland's [Turbo Vision](https://en.wikipedia.org/wiki/Turbo_Vision) — the text-mode application framework from the early 1990s that shipped with Turbo Pascal and Turbo C++.

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## What is this?

TurboView is a from-scratch Go implementation of Turbo Vision's architecture: overlapping resizable windows, a desktop with window management, modal dialogs, menus, status lines, and the classic three-phase event dispatch model (PreProcess, focused, PostProcess). It targets behavioral fidelity to the original C++ library — not just API shape, but focus semantics, keyboard handling, mouse interaction, and the visual language of the Borland-era TUI.

Built on [tcell/v2](https://github.com/gdamore/tcell) for terminal I/O.

## Widgets

| Widget | Description |
|--------|-------------|
| Application | Top-level event loop, screen management, command dispatch |
| Desktop | Window container with tiling, cascading, and z-order management |
| Window | Resizable, draggable, zoomable overlapping window with frame |
| Dialog | Modal window for user interaction |
| MenuBar / MenuPopup | Pull-down menus with keyboard shortcuts and tilde-shortcut labels |
| StatusLine | Context-sensitive status bar with keyboard-bound commands |
| Button | Push button with tilde-shortcut labels |
| InputLine | Single-line text editor with selection, clipboard, and scroll |
| Memo | Multi-line text editor with selection, clipboard, word movement, scrollbars |
| CheckBoxes / CheckBox | Grouped checkboxes with keyboard navigation and focus indicators |
| RadioButtons / RadioButton | Exclusive-selection radio group |
| Label | Linked label with Alt+shortcut that focuses its target widget |
| ListViewer | Virtual list with keyboard/mouse selection and scrollbar linkage |
| ListBox | ListViewer + ScrollBar composite |
| ScrollBar | Vertical/horizontal scrollbar with thumb dragging and arrow clicks |
| History | Input line history dropdown (Turbo Vision's THistory) |
| StaticText | Non-interactive text display |

## Architectural choices

**Event dispatch.** The original three-phase model: Phase 1 broadcasts to unfocused children with `OfPreProcess`, Phase 2 delivers to the focused child, Phase 3 broadcasts to unfocused children with `OfPostProcess`. This is how Label shortcuts, History's Down-arrow interception, and keyboard accelerators all work without special-casing.

**View ownership.** Every view has an owner (its parent container). Focus, coordinate mapping, color scheme inheritance, and event routing all flow through the ownership tree — same as the original.

**GrowMode.** Views declare how their edges move when their parent resizes, using the same flag system (`GfGrowLoX`, `GfGrowHiX`, `GfGrowLoY`, `GfGrowHiY`, `GfGrowRel`) as the original. Composite widgets propagate `SetBounds` to their internal groups so children actually resize.

**Click-to-focus with OfFirstClick.** Clicking an unfocused widget focuses it. Without `OfFirstClick`, the focusing click is consumed; with it, the click is also processed by the widget. This matches the original's `ofFirstClick` option.

**Themes.** Color schemes are plain structs — no global state, no init functions. The demo ships with Borland Blue (the classic), plus Borland Cyan, Borland Gray, C64, and Matrix.

**No hardware cursor.** The framework renders block cursors by drawing the character at the cursor position with a highlight style, rather than using the terminal's hardware cursor. This avoids cursor-positioning edge cases across terminal emulators.

## Testing approach

This project was built with an emphasis on behavioral testing that goes beyond what's typical for AI-assisted development. The experience of building a UI framework taught us that unit tests alone miss interaction bugs — a widget can pass all its unit tests while being completely broken in the running application.

**3,074 unit tests** cover individual widget behavior: focus semantics, keyboard handling, mouse interaction, event consumption, coordinate mapping, selection, clipboard, and drawing output.

**39 end-to-end tests** build the actual binary, launch it in a tmux session, send real keystrokes and mouse events via `tmux send-keys`, capture the screen with `tmux capture-pane`, and assert on visible output. These tests caught bugs that no amount of unit testing found — like Space being eaten by unfocused checkbox clusters during PreProcess, or Alt+N window selection not bringing windows to front in z-order.

The e2e tests require `tmux` to be installed.

## Running

```bash
# Run the demo application
go run ./e2e/testapp/basic

# Run unit tests
go test ./tv/ ./theme/

# Run end-to-end tests (requires tmux)
go test ./e2e/ -timeout 180s
```

## Status

Work in progress. The core framework and the widgets listed above are functional. Not yet implemented from the original: TFileDialog, TColorDialog, TEditWindow, and various other specialized dialogs.

## Acknowledgments

Behavioral reference: [magiblot/tvision](https://github.com/magiblot/tvision) (modern C++ port of Turbo Vision) and the [original Borland documentation](https://archive.org/details/bitsavers_aboraborla_49702702).

## License

MIT
