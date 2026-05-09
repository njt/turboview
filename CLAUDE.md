# TurboView — Project Conventions

## What this is

Go TUI framework reimplementing Borland's Turbo Vision. Behavioral fidelity to the original is the goal — not just API shape, but focus semantics, event dispatch, mouse interaction, keyboard handling, and visual conventions.

## Reference authority

When behavior is ambiguous, the original library is the authority:

- **Source code (primary):** [magiblot/tvision](https://github.com/magiblot/tvision) — modern C++ reimplementation. Clone to `/tmp/tvision` for research. This is where behavioral truth lives for most widgets.
- **Borland docs (secondary):** API reference, not behavioral specs.
  - [TV 2.0 Programming Guide (Pascal, 1992)](https://archive.org/details/bitsavers_borlandTurrogrammingGuide1992_25707423)
  - [TV for C++ User's Guide (1991/1993)](https://archive.org/details/kupdf.net_borland-turbo-vision-for-c-user39s-guide)
  - [Direct PDF mirror](http://bitsavers.informatik.uni-stuttgart.de/pdf/borland/Turbo_Vision_Version_2.0_Programming_Guide_1992.pdf)
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
go test ./e2e/ -timeout 600s
```

### Audit after fixing

When a bug is found and fixed in one widget, audit all similar widgets for the same class of bug before reporting done. The question is always: "which other components could have this same problem?"

## Development skills

This project uses a modified fork of [obra/superpowers](https://github.com/obra/superpowers) (v5.0.5 → v5.4.0), installed in `.claude/skills/`. The fork addresses failure modes discovered while building this project — specifically, multi-agent implementations that produce components passing all unit tests individually but failing when connected.

### What changed from upstream superpowers

**writing-plans:** Plans must sequence phases as capability slices (each phase produces a runnable app), not layer-by-layer (types first, wiring last). Every phase plan ends with an e2e test task that builds the binary and drives it through its real interface. Plans contain requirements and implementation code but NO test code — tests are written independently by a dedicated subagent to prevent confirmation bias.

**subagent-driven-development:** Rewired to: Test Writer → Test Reviewer → Implementer (with locked tests) → Spec Reviewer → Code Quality Reviewer. Test writers receive only the spec, never implementation code. Test reviewers independently derive expected behavior before reading tests. Implementers can challenge tests via a formal protocol (evidence required). Added mandatory integration testing between task batches. Added e2e suite as a phase gate before auto-chaining. Added spec coverage auditor after all phases complete.

**test-driven-development:** Added test quality gold standard (shared by test writers and reviewers). Three test levels: unit, integration, e2e. New anti-patterns: "all unit tests, no integration tests"; "tests that validate bugs as correct"; "e2e tests that route around broken behavior."

**executing-plans, finishing-a-development-branch, brainstorming, using-git-worktrees, verification-before-completion, using-superpowers:** Unchanged from upstream — included for the complete workflow pipeline.

## Issue and task tracking

This project uses two CLI tools for coordination:

- **`kata`** — issue tracker. File bugs, features, and audit findings. Project is bound via `.kata.toml` at repo root. Use `--as claude` for agent-authored issues/comments.
- **`job`** — hierarchical task tracker. Import workflow templates from `workflows/`, claim tasks, mark criteria, close with notes. Use `--as claude` for agent work.
- **`/work-on-issue`** — skill that picks up a kata issue, classifies it (feature / confirmed bug / triage), imports the right workflow template into job, and starts execution.

Workflow templates live in `workflows/`. The workflow requirements doc is `workflows/requirements.md`.

### CLI quick reference

```bash
# kata
kata list --status open                          # not --state
kata create "Title here" --as claude --body "…"  # title is positional, not --title
kata comment 3 --as claude --body "…"            # not -m
kata close 3 --as claude

# job
job import plan.md                               # file must have ```yaml tasks: block
job claim --next --as claude
job edit <id> --set-criterion "label=passed"     # mark criteria before closing
job done <id> --as claude -m "notes"
job done <id> --as claude -m "notes" --claim-next
```

## Demo app

`e2e/testapp/basic/main.go` — exercises all built widgets. Window names ("Controls", "List", "Untitled", "Outline", "Markdown") are referenced by e2e tests; don't rename without updating tests. Any new widget should be added to this demo app.
