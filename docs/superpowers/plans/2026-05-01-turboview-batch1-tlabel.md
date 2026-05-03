# TLabel Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix four behavioral gaps in `tv/label.go` so TLabel matches original Turbo Vision behavior (reference: magiblot/tvision `tlabel.cpp`).

**Architecture:** The existing Label widget has the right structure but deviates from original TV in four observable ways: wrong event-dispatch phase (OfPostProcess only, should be OfPreProcess+OfPostProcess), wrong text column (0, should be 1), missing background fill, and missing OfSelectable guard on focus link activation. Each gap is a small, independent fix to `tv/label.go`. Plain-letter shortcut matching (spec 1.4) is intentionally skipped — only Alt+letter matching is supported.

**Tech Stack:** Go, tcell/v2, existing tv package infrastructure (BaseView, Group three-phase dispatch, DrawBuffer, theme.ColorScheme)

---

## File Map

- **Modify:** `tv/label.go` — constructor (add OfPreProcess), Draw() (column 1 + background fill), HandleEvent (OfSelectable guard)
- **Modify:** `tv/label_phase_test.go` — existing tests assert wrong behavior (no OfPreProcess); must be updated to match spec
- **Modify:** `e2e/testapp/basic/main.go` — add Label+InputLine pair to demonstrate shortcut behavior
- **Modify:** `e2e/e2e_test.go` — add e2e test for Label shortcut activation

---

- [ ] ### Task 1: Add OfPreProcess to Label Constructor

**Files:**
- Modify: `tv/label.go:27`
- Modify: `tv/label_phase_test.go` (entire file — tests assert wrong behavior)

**Requirements:**
- `NewLabel()` sets both `OfPreProcess` and `OfPostProcess` on the Label
- `label.HasOption(OfPreProcess)` returns true after construction
- `label.HasOption(OfPostProcess)` returns true after construction (existing behavior, must not regress)
- When a Label with shortcut `~N~` and a focused `altInterceptorView` (that consumes Alt+N) are siblings in a Group, pressing Alt+N focuses the Label's link — the Label fires in preprocess BEFORE the focused child sees the event. This matches original TV: `TLabel` intercepts hotkeys before the focused view.
- When a Label with shortcut `~N~` is in a Group with a non-intercepting focused sibling, Alt+N still focuses the Label's link (existing behavior, must not regress)

**Implementation:**

In `tv/label.go`, change the constructor to set both options:

```go
func NewLabel(bounds Rect, label string, link View) *Label {
	l := &Label{
		label: label,
		link:  link,
	}
	l.SetBounds(bounds)
	l.SetState(SfVisible, true)
	l.SetOptions(OfPreProcess, true)
	l.SetOptions(OfPostProcess, true)
	// ... rest unchanged
```

The HandleEvent code does not need to change — the same Alt+letter matching logic works in both preprocess and postprocess. With OfPreProcess set, the Group's three-phase dispatch calls Label's HandleEvent during preprocess (phase 1) for non-focused Labels. If the Label matches Alt+shortcut and clears the event, the focused child never sees it. This is the original TV behavior.

**Existing tests to update:** `tv/label_phase_test.go` contains two tests that assert the wrong behavior per the spec:
- `TestLabelPhaseNewLabelDoesNotSetOfPreProcess` — currently asserts OfPreProcess is NOT set. Must be changed to assert OfPreProcess IS set.
- `TestLabelPhasePostProcessDoesNotInterceptBeforeFocusedChild` — currently asserts that a focused interceptor beats the Label. Must be inverted: the Label fires in preprocess and beats the interceptor. The interceptor's `gotEvent` should be false.

**Run tests:** `go test ./tv/... -run TestLabel -v`

**Commit:** `git commit -m "feat(label): add OfPreProcess to match original TV TLabel dispatch phase"`

---

- [ ] ### Task 2: Text Column Offset and Background Fill in Draw

**Files:**
- Modify: `tv/label.go:41-63` (Draw method)

**Requirements:**
- `Draw()` fills the entire widget width (column 0 to bounds width) with spaces in the current style (LabelNormal or LabelHighlight depending on `l.light`) before rendering text. This ensures clean background when text is shorter than bounds.
- Column 0 is always a space (the monochrome marker column — we don't implement the marker but the margin must exist for visual alignment with other dialog controls).
- Text rendering starts at column 1, not column 0.
- For a Label with bounds width 20 and text `"~N~ame"`:
  - Column 0: space, in normal/highlight style
  - Column 1: `N`, in LabelShortcut style
  - Columns 2-4: `ame`, in normal/highlight style
  - Columns 5-19: space, in normal/highlight style (background fill)
- For a highlighted Label (`l.light == true`), the background fill and non-shortcut text use `LabelHighlight` style
- Shortcut segments always use `LabelShortcut` style regardless of highlight state (matching original TV where shortcut palette entries 3 and 4 map to the same index)

**Implementation:**

```go
func (l *Label) Draw(buf *DrawBuffer) {
	normalStyle := tcell.StyleDefault
	shortcutStyle := tcell.StyleDefault
	if cs := l.ColorScheme(); cs != nil {
		if l.light {
			normalStyle = cs.LabelHighlight
		} else {
			normalStyle = cs.LabelNormal
		}
		shortcutStyle = cs.LabelShortcut
	}

	// Background fill: entire bounds width with normal/highlight style
	b := l.Bounds()
	buf.Fill(NewRect(0, 0, b.Width(), 1), ' ', normalStyle)

	// Text starts at column 1 (column 0 is monochrome marker margin)
	x := 1
	segments := ParseTildeLabel(l.label)
	for _, seg := range segments {
		style := normalStyle
		if seg.Shortcut {
			style = shortcutStyle
		}
		buf.WriteStr(x, 0, seg.Text, style)
		x += utf8.RuneCountInString(seg.Text)
	}
}
```

**Existing tests affected:** Multiple tests in `label_test.go` check rune positions assuming text starts at column 0:
- `TestLabelDrawNormalSegmentUsesLabelNormalStyle` — checks cell(0,0), must check cell(1,0)
- `TestLabelDrawShortcutSegmentUsesLabelShortcutStyle` — checks cell(0,0), must check cell(1,0)
- `TestLabelDrawNormalSegmentAfterShortcut` — checks cell(1,0), must check cell(2,0)
- `TestLabelDrawRendersCorrectRunes` — checks cells 0-3, must check cells 1-4
- `TestLabelDrawNoTildeRendersEntireTextAsNormal` — iterates from index 0, must start from 1
- `TestLabelDrawShortcutStyleDiffersFromNormal` — checks cells(0,0) and (1,0), must check (1,0) and (2,0)

These tests verify correct behavior. They must be updated to use the new column positions (shift all expected positions by +1). The test writer will write fresh tests for the new behavior including background fill verification.

**Run tests:** `go test ./tv/... -run TestLabel -v`

**Commit:** `git commit -m "feat(label): start text at column 1 with background fill per original TV TLabel"`

---

- [ ] ### Task 3: OfSelectable Guard on Focus Link

**Files:**
- Modify: `tv/label.go:65-104` (HandleEvent method)

**Requirements:**
- Before calling `owner.SetFocusedChild(l.link)` in response to Alt+shortcut, Label must check `l.link.HasOption(OfSelectable)`. If the link does not have OfSelectable (e.g., disabled view), the event is still cleared but no focus change occurs.
- Before calling `owner.SetFocusedChild(l.link)` in response to mouse Button1 click, the same OfSelectable check applies.
- When the linked view IS selectable (has OfSelectable), the existing behavior is unchanged: focus moves to the link.
- When the linked view is NOT selectable (OfSelectable is false), Alt+shortcut clears the event but focus does NOT change.
- When the linked view is NOT selectable, mouse Button1 click clears the event but focus does NOT change.
- Broadcast handling (CmReceivedFocus / CmReleasedFocus) is NOT affected by this change — the label still tracks highlight state regardless of OfSelectable.

**Implementation:**

```go
func (l *Label) HandleEvent(event *Event) {
	if event.What == EvMouse && event.Mouse != nil {
		if event.Mouse.Button&tcell.Button1 != 0 && l.link != nil {
			if l.link.HasOption(OfSelectable) {
				if owner := l.Owner(); owner != nil {
					owner.SetFocusedChild(l.link)
				}
			}
			event.Clear()
		}
		return
	}

	if event.What == EvBroadcast && l.link != nil {
		switch event.Command {
		case CmReceivedFocus:
			if event.Info == l.link {
				l.light = true
			}
		case CmReleasedFocus:
			if event.Info == l.link {
				l.light = false
			}
		}
		return
	}

	if l.link == nil || l.shortcut == 0 {
		return
	}
	if event.What != EvKeyboard || event.Key == nil {
		return
	}
	if event.Key.Modifiers&tcell.ModAlt != 0 && event.Key.Key == tcell.KeyRune {
		if unicode.ToLower(event.Key.Rune) == unicode.ToLower(l.shortcut) {
			if l.link.HasOption(OfSelectable) {
				if owner := l.Owner(); owner != nil {
					owner.SetFocusedChild(l.link)
				}
			}
			event.Clear()
		}
	}
}
```

**Run tests:** `go test ./tv/... -run TestLabel -v`

**Commit:** `git commit -m "feat(label): guard focusLink with OfSelectable check per original TV"`

---

- [ ] ### Task 4: Integration Checkpoint — Label Behavioral Fixes

**Purpose:** Verify that Tasks 1-3 work together correctly in realistic dialog scenarios.

**Requirements (for test writer):**
- A Label with shortcut `~N~` in a Group (or Window), linked to an InputLine, focuses the InputLine when Alt+N is pressed — through the full event dispatch chain (Group → preprocess phase → Label.HandleEvent → SetFocusedChild)
- The Label's light state tracks its linked InputLine's focus: when InputLine gains focus (by any means — Tab, Alt+shortcut, mouse click), the Label draws with LabelHighlight. When InputLine loses focus, Label draws with LabelNormal.
- A Label linked to a disabled view (OfSelectable cleared) does NOT change focus on Alt+shortcut, but the event IS cleared.
- Column 0 of a drawn Label is always a space (margin), text starts at column 1, and trailing columns are filled with the appropriate background style.
- Two Labels in the same Group with different shortcuts both work independently — Alt+N focuses one link, Alt+S focuses the other.

**Components to wire up:** Label, InputLine, Button, Group or Window (all real, no mocks)

**Run:** `go test ./tv/... -run TestIntegration -v`

---

- [ ] ### Task 5: E2E Test — Label Shortcut Activation

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**
- The demo app includes a Label with shortcut linked to an InputLine, visible when the File Manager window (win1) is focused
- The e2e test verifies: press Alt+N (or whichever shortcut), the InputLine receives focus (cursor is visible in the InputLine area), type text, the text appears on screen
- The e2e test verifies: the Label text is visible on screen with correct spacing (text starting at column 1, space at column 0)

**Implementation:**

Add a Label+InputLine pair to `win1` in `e2e/testapp/basic/main.go`. Place them at a row that fits within win1's client area (win1 is 35x15, client area ~33x13 after frame):

```go
// After radioButtons insertion and before app.Desktop().Insert(win1)
inputLine := tv.NewInputLine(tv.NewRect(11, 12, 20, 1), 40)
win1.Insert(inputLine)
nameLabel := tv.NewLabel(tv.NewRect(1, 12, 10, 1), "~N~ame:", inputLine)
win1.Insert(nameLabel)
```

The e2e test infrastructure uses tmux: `buildBasicApp` builds the binary, `startTmux` launches it in a tmux session, `tmuxSendKeys` sends key events, `tmuxCapture` reads screen contents, and `containsAny` checks for substrings. Session names must be unique per test. Tests end with Alt+X exit and session cleanup.

**Run tests:** `go test ./e2e/... -run TestLabelShortcut -v -timeout 60s`

**Commit:** `git commit -m "test(e2e): add Label shortcut activation e2e test"`
