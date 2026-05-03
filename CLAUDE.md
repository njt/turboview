# TurboView — Project Conventions

## What this is

Go TUI framework reimplementing Borland's Turbo Vision. Behavioral fidelity to the original is the goal — not just API shape, but focus semantics, event dispatch, mouse interaction, keyboard handling, and visual conventions.

## Reference authority

When behavior is ambiguous, the original library is the authority:

- **Source code (primary):** [magiblot/tvision](https://github.com/magiblot/tvision) — modern C++ reimplementation. Clone to `/tmp/tvision` for research. This is where behavioral truth lives for most widgets.
- **Borland docs (secondary):** Archive.org has the Programming Guide and User's Guide. These give API reference but rarely specify behavioral details (focus transitions, draw state changes, event consumption).
- **Key finding:** For most widgets, source code is the ONLY reliable behavioral reference. The manuals describe what methods exist, not what they do in edge cases.

When a bug report involves *behavior*, *consistency*, or *defaults* — check the original before deciding what correct behavior is.

## Architecture

- Three-phase event dispatch: PreProcess (unfocused + OfPreProcess), focused child, PostProcess (unfocused + OfPostProcess)
- View ownership tree: focus, coordinates, color schemes, and events flow through the owner chain
- Composite widgets (ListBox, CheckBoxes, RadioButtons) have internal Groups and must propagate SetBounds and set GrowMode on children
- Click-to-focus via BaseView.HandleEvent — every widget that handles mouse events must call it. OfFirstClick controls whether the focusing click is also processed.

## Testing

### Test-first, always

Never fix bugs directly. Write a failing test first, then fix the code. Tests must be strong validators — weak tests get strengthened, never sidelined.

### End-to-end tests are mandatory for UI behavior

Unit tests prove components work in isolation. E2e tests prove the app works when a user runs it. These are different things — this project has had passing unit tests with broken interactive behavior multiple times.

E2e tests: build the binary, launch in tmux, send keystrokes via `tmux send-keys`, capture screen via `tmux capture-pane`, assert on visible output.

```bash
# Run the demo app interactively
go run ./e2e/testapp/basic

# Run unit tests
go test ./tv/ ./theme/

# Run e2e tests (requires tmux)
go test ./e2e/ -timeout 180s
```

### Audit after fixing

When a bug is found and fixed in one widget, audit all similar widgets for the same class of bug before reporting done. The question is always: "which other components could have this same problem?"

## Demo app

`e2e/testapp/basic/main.go` — exercises all built widgets. Window names ("File Manager", "Editor", "Notes") are referenced by e2e tests; don't rename without updating tests. Any new widget should be added to this demo app.
