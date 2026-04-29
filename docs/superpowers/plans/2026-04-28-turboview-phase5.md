# TurboView Phase 5: Form Input Widgets Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add form input widgets — InputLine (single-line text editor), CheckBox/CheckBoxes cluster, RadioButton/RadioButtons cluster, and the InputBox dialog convenience function.

**Architecture:** Each widget embeds `BaseView`, implements `Widget`, and follows the same constructor/Draw/HandleEvent pattern established by Button and Label. CheckBoxes and RadioButtons are Container-like clusters (they embed a Group internally, same pattern as Dialog). InputBox is a convenience function that creates a Dialog with a Label + InputLine + OK/Cancel buttons and calls ExecView, paralleling MessageBox.

**Tech Stack:** Go, tcell/v2, existing tv package patterns (BaseView embedding, ParseTildeLabel, ColorScheme field access)

---

## File Structure

| File | Responsibility |
|------|---------------|
| `tv/input_line.go` | InputLine widget — single-line text input with cursor movement, selection, clipboard, max length |
| `tv/checkbox.go` | CheckBox widget + CheckBoxes cluster container |
| `tv/radio.go` | RadioButton widget + RadioButtons cluster container |
| `tv/dialog.go` | Modified — add InputBox convenience function |

## Batch Structure

- **Batch 1 (Tasks 1-3):** Individual widgets — InputLine, CheckBox, RadioButton
- **Task 4:** Integration Checkpoint — widgets work inside Dialog containers
- **Batch 2 (Tasks 5-6):** InputBox dialog function + Demo App & E2E

---

### Task 1: InputLine Widget

**Files:**
- Create: `tv/input_line.go`
- Test: `tv/input_line_test.go`

**Requirements:**

Constructor:
- `NewInputLine(bounds Rect, maxLen int, opts ...InputLineOption)` creates an InputLine
- Sets `SfVisible` and `OfSelectable` by default
- `maxLen` limits the number of runes that can be entered (0 means unlimited)
- Stores an internal text buffer as `[]rune`

Accessors:
- `Text() string` returns current text content
- `SetText(s string)` sets the text, clamping to maxLen if set, and resets cursor to end
- `CursorPos() int` returns the cursor position (rune index)

Drawing:
- Fills its bounds width×1 with `InputNormal` style from ColorScheme
- Renders text starting from a scroll offset so the cursor is always visible
- When the widget has `SfSelected` state, displays a cursor indicator: the character at cursor position uses `InputSelection` style (or a reverse-video style if no selection exists)
- If text is longer than visible width, scrolls so cursor position is visible within the bounds
- Left-arrow indicators or text clipping at edges when text extends beyond visible area

Keyboard handling (when focused/selected):
- Printable runes: insert at cursor position, advance cursor; no-op if at maxLen
- `Left arrow`: move cursor left one rune (no-op at position 0)
- `Right arrow`: move cursor right one rune (no-op at end of text)
- `Home`: move cursor to position 0
- `End`: move cursor to end of text
- `Backspace`: delete rune before cursor, move cursor left (no-op at position 0)
- `Delete (tcell.KeyDelete)`: delete rune at cursor position (no-op at end)
- `Ctrl+A`: select all text (set selection to cover entire text)
- `Ctrl+C`: copy selected text to internal clipboard
- `Ctrl+V`: paste from internal clipboard at cursor position (respecting maxLen)
- `Ctrl+X`: cut selected text to internal clipboard
- All keyboard events that InputLine handles must be consumed via `event.Clear()`
- Events the widget doesn't handle (Tab, Escape, etc.) pass through unconsumed

Selection:
- Selection is tracked as `selStart, selEnd int` (rune indices); when equal, no selection
- Shift+Left/Shift+Right extends or contracts the selection
- Shift+Home selects from cursor to start; Shift+End selects from cursor to end
- Typing with active selection replaces the selected text
- Backspace/Delete with active selection deletes the selected text
- Selected text renders with `InputSelection` style

Mouse handling:
- Click (Button1) within the widget bounds positions the cursor at the clicked column (accounting for scroll offset and the widget's origin X, since Group delivers mouse events in owner-local coordinates without translating to the child's origin)
- The click event is consumed

Internal clipboard:
- A package-level `var clipboard string` shared across all InputLine instances
- Ctrl+C copies selection to clipboard, Ctrl+X cuts, Ctrl+V pastes

Scroll behavior:
- `scrollOffset int` tracks how many runes are scrolled off the left edge
- When cursor moves left of scrollOffset, scrollOffset decreases to show cursor
- When cursor moves past scrollOffset + visible width, scrollOffset increases
- Drawing starts from text[scrollOffset:]

**Implementation:**

```go
package tv

import "github.com/gdamore/tcell/v2"

var _ Widget = (*InputLine)(nil)

var clipboard string

type InputLine struct {
    BaseView
    text         []rune
    maxLen       int
    cursorPos    int
    scrollOffset int
    selStart     int
    selEnd       int
}

type InputLineOption func(*InputLine)

func NewInputLine(bounds Rect, maxLen int, opts ...InputLineOption) *InputLine {
    il := &InputLine{maxLen: maxLen}
    il.SetBounds(bounds)
    il.SetState(SfVisible, true)
    il.SetOptions(OfSelectable, true)
    for _, opt := range opts {
        opt(il)
    }
    return il
}

func (il *InputLine) Text() string        { return string(il.text) }
func (il *InputLine) CursorPos() int      { return il.cursorPos }
func (il *InputLine) Selection() (int, int) { return il.selStart, il.selEnd }

func (il *InputLine) SetText(s string) {
    runes := []rune(s)
    if il.maxLen > 0 && len(runes) > il.maxLen {
        runes = runes[:il.maxLen]
    }
    il.text = runes
    il.cursorPos = len(runes)
    il.selStart = 0
    il.selEnd = 0
    il.adjustScroll()
}

func (il *InputLine) Draw(buf *DrawBuffer) {
    w := il.Bounds().Width()
    if w <= 0 {
        return
    }

    normalStyle := tcell.StyleDefault
    selStyle := tcell.StyleDefault
    if cs := il.ColorScheme(); cs != nil {
        normalStyle = cs.InputNormal
        selStyle = cs.InputSelection
    }

    buf.Fill(NewRect(0, 0, w, 1), ' ', normalStyle)

    il.adjustScroll()

    // Determine normalized selection range
    sStart, sEnd := il.normalizedSel()

    for i := 0; i < w; i++ {
        ri := il.scrollOffset + i
        if ri >= len(il.text) {
            break
        }

        style := normalStyle
        if ri >= sStart && ri < sEnd {
            style = selStyle
        }
        buf.WriteChar(i, 0, il.text[ri], style)
    }

    // Cursor indicator when focused and no selection at cursor
    if il.HasState(SfSelected) && sStart == sEnd {
        cx := il.cursorPos - il.scrollOffset
        if cx >= 0 && cx < w {
            cell := buf.GetCell(cx, 0)
            buf.WriteChar(cx, 0, cell.Rune, selStyle)
        }
    }
}

func (il *InputLine) normalizedSel() (int, int) {
    if il.selStart <= il.selEnd {
        return il.selStart, il.selEnd
    }
    return il.selEnd, il.selStart
}

func (il *InputLine) adjustScroll() {
    w := il.Bounds().Width()
    if w <= 0 {
        return
    }
    if il.cursorPos < il.scrollOffset {
        il.scrollOffset = il.cursorPos
    }
    if il.cursorPos >= il.scrollOffset+w {
        il.scrollOffset = il.cursorPos - w + 1
    }
    if il.scrollOffset < 0 {
        il.scrollOffset = 0
    }
}

func (il *InputLine) hasSelection() bool {
    return il.selStart != il.selEnd
}

func (il *InputLine) selectedText() string {
    s, e := il.normalizedSel()
    if s == e || s >= len(il.text) {
        return ""
    }
    if e > len(il.text) {
        e = len(il.text)
    }
    return string(il.text[s:e])
}

func (il *InputLine) deleteSelection() {
    s, e := il.normalizedSel()
    if s == e {
        return
    }
    if e > len(il.text) {
        e = len(il.text)
    }
    il.text = append(il.text[:s], il.text[e:]...)
    il.cursorPos = s
    il.selStart = 0
    il.selEnd = 0
}

func (il *InputLine) insertAtCursor(runes []rune) {
    if il.maxLen > 0 {
        avail := il.maxLen - len(il.text)
        if avail <= 0 {
            return
        }
        if len(runes) > avail {
            runes = runes[:avail]
        }
    }
    newText := make([]rune, 0, len(il.text)+len(runes))
    newText = append(newText, il.text[:il.cursorPos]...)
    newText = append(newText, runes...)
    newText = append(newText, il.text[il.cursorPos:]...)
    il.text = newText
    il.cursorPos += len(runes)
}

func (il *InputLine) HandleEvent(event *Event) {
    if event.What == EvMouse && event.Mouse != nil {
        if event.Mouse.Button&tcell.Button1 != 0 {
            // Mouse X is in owner-local (Group) space; subtract our origin
            // to get widget-local X, then add scrollOffset to get rune index.
            col := event.Mouse.X - il.Bounds().A.X + il.scrollOffset
            if col > len(il.text) {
                col = len(il.text)
            }
            if col < 0 {
                col = 0
            }
            il.cursorPos = col
            il.selStart = 0
            il.selEnd = 0
            il.adjustScroll()
            event.Clear()
        }
        return
    }

    if event.What != EvKeyboard || event.Key == nil {
        return
    }

    ke := event.Key

    switch ke.Key {
    case tcell.KeyLeft:
        if ke.Modifiers&tcell.ModShift != 0 {
            if il.selStart == il.selEnd {
                il.selStart = il.cursorPos
                il.selEnd = il.cursorPos
            }
            if il.cursorPos > 0 {
                il.cursorPos--
                il.selEnd = il.cursorPos
            }
        } else {
            if il.cursorPos > 0 {
                il.cursorPos--
            }
            il.selStart = 0
            il.selEnd = 0
        }
        il.adjustScroll()
        event.Clear()

    case tcell.KeyRight:
        if ke.Modifiers&tcell.ModShift != 0 {
            if il.selStart == il.selEnd {
                il.selStart = il.cursorPos
                il.selEnd = il.cursorPos
            }
            if il.cursorPos < len(il.text) {
                il.cursorPos++
                il.selEnd = il.cursorPos
            }
        } else {
            if il.cursorPos < len(il.text) {
                il.cursorPos++
            }
            il.selStart = 0
            il.selEnd = 0
        }
        il.adjustScroll()
        event.Clear()

    case tcell.KeyHome:
        if ke.Modifiers&tcell.ModShift != 0 {
            if il.selStart == il.selEnd {
                il.selStart = il.cursorPos
                il.selEnd = il.cursorPos
            }
            il.cursorPos = 0
            il.selEnd = 0
        } else {
            il.cursorPos = 0
            il.selStart = 0
            il.selEnd = 0
        }
        il.scrollOffset = 0
        event.Clear()

    case tcell.KeyEnd:
        if ke.Modifiers&tcell.ModShift != 0 {
            if il.selStart == il.selEnd {
                il.selStart = il.cursorPos
                il.selEnd = il.cursorPos
            }
            il.cursorPos = len(il.text)
            il.selEnd = len(il.text)
        } else {
            il.cursorPos = len(il.text)
            il.selStart = 0
            il.selEnd = 0
        }
        il.adjustScroll()
        event.Clear()

    case tcell.KeyBackspace, tcell.KeyBackspace2:
        if il.hasSelection() {
            il.deleteSelection()
        } else if il.cursorPos > 0 {
            il.text = append(il.text[:il.cursorPos-1], il.text[il.cursorPos:]...)
            il.cursorPos--
        }
        il.adjustScroll()
        event.Clear()

    case tcell.KeyDelete:
        if il.hasSelection() {
            il.deleteSelection()
        } else if il.cursorPos < len(il.text) {
            il.text = append(il.text[:il.cursorPos], il.text[il.cursorPos+1:]...)
        }
        il.adjustScroll()
        event.Clear()

    case tcell.KeyRune:
        if ke.Modifiers&tcell.ModCtrl != 0 || ke.Modifiers&tcell.ModAlt != 0 {
            // Let Ctrl/Alt combos pass through (handled below for Ctrl+A/C/V/X)
            break
        }
        if il.hasSelection() {
            il.deleteSelection()
        }
        il.insertAtCursor([]rune{ke.Rune})
        il.adjustScroll()
        event.Clear()

    default:
        if ke.Key == tcell.KeyCtrlA {
            il.selStart = 0
            il.selEnd = len(il.text)
            il.cursorPos = len(il.text)
            event.Clear()
            return
        }
        if ke.Key == tcell.KeyCtrlC {
            if il.hasSelection() {
                clipboard = il.selectedText()
            }
            event.Clear()
            return
        }
        if ke.Key == tcell.KeyCtrlV {
            if il.hasSelection() {
                il.deleteSelection()
            }
            if clipboard != "" {
                il.insertAtCursor([]rune(clipboard))
            }
            il.adjustScroll()
            event.Clear()
            return
        }
        if ke.Key == tcell.KeyCtrlX {
            if il.hasSelection() {
                clipboard = il.selectedText()
                il.deleteSelection()
            }
            il.adjustScroll()
            event.Clear()
            return
        }
    }
}
```

**Run tests:** `go test ./tv/... -run TestInputLine -v`

**Commit:** `git commit -m "feat(tv): add InputLine single-line text input widget"`

---

### Task 2: CheckBox Widget and CheckBoxes Cluster

**Files:**
- Create: `tv/checkbox.go`
- Test: `tv/checkbox_test.go`

**Requirements:**

**CheckBox (individual widget):**

Constructor:
- `NewCheckBox(bounds Rect, label string) *CheckBox`
- Sets `SfVisible` and `OfSelectable` by default
- Implements the `Widget` interface
- Stores a boolean `checked` state

Accessors:
- `Checked() bool` returns current checked state
- `SetChecked(bool)` sets the checked state
- `Label() string` returns the label

Drawing:
- Renders as `[X] Label` when checked, `[ ] Label` when unchecked
- Uses `CheckBoxNormal` style from ColorScheme for the bracket/mark/label, and `LabelShortcut` style for tilde-shortcut characters in the label
- The label supports tilde shortcut notation (e.g., `~S~ave settings`) — shortcut character uses `LabelShortcut` style; shortcut only acts as an accelerator when part of a CheckBoxes cluster
- Total rendered width: 4 + tildeTextLen(label) (bracket + X/space + bracket + space + label)
- When `SfSelected`, a `►` prefix is rendered before the brackets (same pattern as Button), shifting the bracket/label right by 1

Keyboard handling:
- `Space`: toggles checked state, consumes event
- `Enter`: toggles checked state, consumes event

Mouse handling:
- Click (Button1) within bounds: toggles checked state, consumes event

**CheckBoxes (cluster container):**

Constructor:
- `NewCheckBoxes(bounds Rect, labels []string) *CheckBoxes`
- Creates one CheckBox per label, arranged vertically (one per row)
- Each CheckBox is positioned at y=index within the cluster
- Implements the `Container` interface (delegates to internal Group, same pattern as Dialog)
- Sets `SfVisible` and `OfSelectable` by default

Accessors:
- `Values() uint32` returns a bitmask of checked states (bit 0 = first item, etc.)
- `SetValues(uint32)` sets checked states from bitmask
- `Item(index int) *CheckBox` returns the CheckBox at the given index

Drawing:
- Delegates to internal Group which draws each CheckBox in its sub-buffer

Keyboard handling:
- The cluster's Group handles Tab/Shift+Tab between checkboxes
- Alt+shortcut letter focuses and toggles the corresponding checkbox (handled via preprocess, same pattern as Label)

Event handling:
- Delegates to internal Group for three-phase dispatch

**Implementation:**

```go
package tv

import (
    "unicode"
    "unicode/utf8"

    "github.com/gdamore/tcell/v2"
)

var _ Widget = (*CheckBox)(nil)

type CheckBox struct {
    BaseView
    label    string
    checked  bool
    shortcut rune
}

func NewCheckBox(bounds Rect, label string) *CheckBox {
    cb := &CheckBox{label: label}
    cb.SetBounds(bounds)
    cb.SetState(SfVisible, true)
    cb.SetOptions(OfSelectable, true)

    segments := ParseTildeLabel(label)
    for _, seg := range segments {
        if seg.Shortcut && len(seg.Text) > 0 {
            cb.shortcut, _ = utf8.DecodeRuneInString(seg.Text)
            break
        }
    }
    return cb
}

func (cb *CheckBox) Checked() bool       { return cb.checked }
func (cb *CheckBox) SetChecked(v bool)   { cb.checked = v }
func (cb *CheckBox) Label() string       { return cb.label }
func (cb *CheckBox) Shortcut() rune      { return cb.shortcut }

func (cb *CheckBox) Draw(buf *DrawBuffer) {
    style := tcell.StyleDefault
    shortcutStyle := tcell.StyleDefault
    if cs := cb.ColorScheme(); cs != nil {
        style = cs.CheckBoxNormal
        shortcutStyle = cs.LabelShortcut
    }

    if cb.HasState(SfSelected) {
        buf.WriteChar(0, 0, '►', style)
    }

    mark := ' '
    if cb.checked {
        mark = 'X'
    }
    startX := 0
    if cb.HasState(SfSelected) {
        startX = 1
    }
    buf.WriteChar(startX, 0, '[', style)
    buf.WriteChar(startX+1, 0, mark, style)
    buf.WriteChar(startX+2, 0, ']', style)
    buf.WriteChar(startX+3, 0, ' ', style)

    x := startX + 4
    segments := ParseTildeLabel(cb.label)
    for _, seg := range segments {
        s := style
        if seg.Shortcut {
            s = shortcutStyle
        }
        buf.WriteStr(x, 0, seg.Text, s)
        x += utf8.RuneCountInString(seg.Text)
    }
}

func (cb *CheckBox) toggle() {
    cb.checked = !cb.checked
}

func (cb *CheckBox) HandleEvent(event *Event) {
    if event.What == EvMouse && event.Mouse != nil {
        if event.Mouse.Button&tcell.Button1 != 0 {
            cb.toggle()
            event.Clear()
        }
        return
    }
    if event.What == EvKeyboard && event.Key != nil {
        switch event.Key.Key {
        case tcell.KeyRune:
            if event.Key.Rune == ' ' {
                cb.toggle()
                event.Clear()
            }
        case tcell.KeyEnter:
            cb.toggle()
            event.Clear()
        }
    }
}

// --- CheckBoxes cluster ---

var _ Container = (*CheckBoxes)(nil)

type CheckBoxes struct {
    BaseView
    group *Group
    items []*CheckBox
}

func NewCheckBoxes(bounds Rect, labels []string) *CheckBoxes {
    cbs := &CheckBoxes{}
    cbs.SetBounds(bounds)
    cbs.SetState(SfVisible, true)
    cbs.SetOptions(OfSelectable|OfPreProcess, true)

    cbs.group = NewGroup(NewRect(0, 0, bounds.Width(), bounds.Height()))
    cbs.group.SetFacade(cbs)

    for i, label := range labels {
        cb := NewCheckBox(NewRect(0, i, bounds.Width(), 1), label)
        cbs.group.Insert(cb)
        cbs.items = append(cbs.items, cb)
    }

    return cbs
}

func (cbs *CheckBoxes) Item(index int) *CheckBox {
    if index < 0 || index >= len(cbs.items) {
        return nil
    }
    return cbs.items[index]
}

func (cbs *CheckBoxes) Values() uint32 {
    var v uint32
    for i, cb := range cbs.items {
        if cb.Checked() && i < 32 {
            v |= 1 << uint(i)
        }
    }
    return v
}

func (cbs *CheckBoxes) SetValues(v uint32) {
    for i, cb := range cbs.items {
        if i < 32 {
            cb.SetChecked(v&(1<<uint(i)) != 0)
        }
    }
}

func (cbs *CheckBoxes) Insert(v View)               { cbs.group.Insert(v) }
func (cbs *CheckBoxes) Remove(v View)               { cbs.group.Remove(v) }
func (cbs *CheckBoxes) Children() []View            { return cbs.group.Children() }
func (cbs *CheckBoxes) FocusedChild() View          { return cbs.group.FocusedChild() }
func (cbs *CheckBoxes) SetFocusedChild(v View)      { cbs.group.SetFocusedChild(v) }
func (cbs *CheckBoxes) ExecView(v View) CommandCode { return cbs.group.ExecView(v) }
func (cbs *CheckBoxes) BringToFront(v View)         { cbs.group.BringToFront(v) }

func (cbs *CheckBoxes) Draw(buf *DrawBuffer) {
    cbs.group.Draw(buf)
}

func (cbs *CheckBoxes) HandleEvent(event *Event) {
    // Alt+shortcut: find matching checkbox, focus it, toggle it
    if event.What == EvKeyboard && event.Key != nil {
        if event.Key.Modifiers&tcell.ModAlt != 0 && event.Key.Key == tcell.KeyRune {
            r := unicode.ToLower(event.Key.Rune)
            for _, cb := range cbs.items {
                if cb.shortcut != 0 && unicode.ToLower(cb.shortcut) == r {
                    cbs.group.SetFocusedChild(cb)
                    cb.toggle()
                    event.Clear()
                    return
                }
            }
        }
    }
    cbs.group.HandleEvent(event)
}
```

**Run tests:** `go test ./tv/... -run TestCheckBox -v`

**Commit:** `git commit -m "feat(tv): add CheckBox widget and CheckBoxes cluster"`

---

### Task 3: RadioButton Widget and RadioButtons Cluster

**Files:**
- Create: `tv/radio.go`
- Test: `tv/radio_test.go`

**Requirements:**

**RadioButton (individual widget):**

Constructor:
- `NewRadioButton(bounds Rect, label string) *RadioButton`
- Sets `SfVisible` and `OfSelectable` by default
- Implements the `Widget` interface
- Stores a boolean `selected` state (only one in a cluster can be selected)

Accessors:
- `Selected() bool` returns whether this radio button is the active one
- `SetSelected(bool)` sets selected state
- `Label() string` returns the label

Drawing:
- Renders as `(*) Label` when selected, `( ) Label` when not selected
- Uses `RadioButtonNormal` style from ColorScheme for the bracket/mark/label, and `LabelShortcut` style for tilde-shortcut characters in the label
- The label supports tilde shortcut notation — shortcut character uses `LabelShortcut` style
- Total rendered width: 4 + tildeTextLen(label)
- When `SfSelected` (has focus), a `►` prefix is rendered before the parentheses, shifting them right by 1

Keyboard handling:
- `Space`: selects this radio button (deselects siblings via cluster), consumes event
- `Enter`: same as Space

Mouse handling:
- Click (Button1) within bounds: selects this radio button, consumes event

**RadioButtons (cluster container):**

Constructor:
- `NewRadioButtons(bounds Rect, labels []string) *RadioButtons`
- Creates one RadioButton per label, arranged vertically
- Each RadioButton positioned at y=index
- Implements the `Container` interface (delegates to internal Group)
- The first radio button is selected by default
- Sets `SfVisible` and `OfSelectable` by default

Accessors:
- `Value() int` returns the index of the selected radio button (0-based); -1 if none
- `SetValue(int)` selects the radio button at the given index (deselects others)
- `Item(index int) *RadioButton` returns the RadioButton at the given index

Exclusive selection:
- When a RadioButton in the cluster is selected (via keyboard, mouse, or SetValue), all other RadioButtons in the cluster are deselected
- The cluster intercepts the selection event and enforces mutual exclusion

Drawing:
- Delegates to internal Group

Keyboard/Event handling:
- Group handles Tab/Shift+Tab between radio buttons
- Alt+shortcut letter focuses and selects the corresponding radio button
- Up/Down arrows move between radio buttons AND select them (unlike checkboxes where arrows just move focus)

**Implementation:**

```go
package tv

import (
    "unicode"
    "unicode/utf8"

    "github.com/gdamore/tcell/v2"
)

var _ Widget = (*RadioButton)(nil)

type RadioButton struct {
    BaseView
    label    string
    selected bool
    shortcut rune
    cluster  *RadioButtons
}

func NewRadioButton(bounds Rect, label string) *RadioButton {
    rb := &RadioButton{label: label}
    rb.SetBounds(bounds)
    rb.SetState(SfVisible, true)
    rb.SetOptions(OfSelectable, true)

    segments := ParseTildeLabel(label)
    for _, seg := range segments {
        if seg.Shortcut && len(seg.Text) > 0 {
            rb.shortcut, _ = utf8.DecodeRuneInString(seg.Text)
            break
        }
    }
    return rb
}

func (rb *RadioButton) Selected() bool      { return rb.selected }
func (rb *RadioButton) Label() string       { return rb.label }
func (rb *RadioButton) Shortcut() rune      { return rb.shortcut }

func (rb *RadioButton) SetSelected(v bool) {
    rb.selected = v
}

func (rb *RadioButton) selectInCluster() {
    if rb.cluster != nil {
        rb.cluster.selectItem(rb)
    } else {
        rb.selected = true
    }
}

func (rb *RadioButton) Draw(buf *DrawBuffer) {
    style := tcell.StyleDefault
    shortcutStyle := tcell.StyleDefault
    if cs := rb.ColorScheme(); cs != nil {
        style = cs.RadioButtonNormal
        shortcutStyle = cs.LabelShortcut
    }

    startX := 0
    if rb.HasState(SfSelected) {
        buf.WriteChar(0, 0, '►', style)
        startX = 1
    }

    mark := ' '
    if rb.selected {
        mark = '*'
    }
    buf.WriteChar(startX, 0, '(', style)
    buf.WriteChar(startX+1, 0, mark, style)
    buf.WriteChar(startX+2, 0, ')', style)
    buf.WriteChar(startX+3, 0, ' ', style)

    x := startX + 4
    segments := ParseTildeLabel(rb.label)
    for _, seg := range segments {
        s := style
        if seg.Shortcut {
            s = shortcutStyle
        }
        buf.WriteStr(x, 0, seg.Text, s)
        x += utf8.RuneCountInString(seg.Text)
    }
}

func (rb *RadioButton) HandleEvent(event *Event) {
    if event.What == EvMouse && event.Mouse != nil {
        if event.Mouse.Button&tcell.Button1 != 0 {
            rb.selectInCluster()
            event.Clear()
        }
        return
    }
    if event.What == EvKeyboard && event.Key != nil {
        switch event.Key.Key {
        case tcell.KeyRune:
            if event.Key.Rune == ' ' {
                rb.selectInCluster()
                event.Clear()
            }
        case tcell.KeyEnter:
            rb.selectInCluster()
            event.Clear()
        }
    }
}

// --- RadioButtons cluster ---

var _ Container = (*RadioButtons)(nil)

type RadioButtons struct {
    BaseView
    group *Group
    items []*RadioButton
}

func NewRadioButtons(bounds Rect, labels []string) *RadioButtons {
    rbs := &RadioButtons{}
    rbs.SetBounds(bounds)
    rbs.SetState(SfVisible, true)
    rbs.SetOptions(OfSelectable|OfPreProcess, true)

    rbs.group = NewGroup(NewRect(0, 0, bounds.Width(), bounds.Height()))
    rbs.group.SetFacade(rbs)

    for i, label := range labels {
        rb := NewRadioButton(NewRect(0, i, bounds.Width(), 1), label)
        rb.cluster = rbs
        rbs.group.Insert(rb)
        rbs.items = append(rbs.items, rb)
    }

    // Select first item by default
    if len(rbs.items) > 0 {
        rbs.items[0].selected = true
    }

    return rbs
}

func (rbs *RadioButtons) Item(index int) *RadioButton {
    if index < 0 || index >= len(rbs.items) {
        return nil
    }
    return rbs.items[index]
}

func (rbs *RadioButtons) Value() int {
    for i, rb := range rbs.items {
        if rb.selected {
            return i
        }
    }
    return -1
}

func (rbs *RadioButtons) SetValue(index int) {
    for i, rb := range rbs.items {
        rb.selected = (i == index)
    }
}

func (rbs *RadioButtons) selectItem(target *RadioButton) {
    for _, rb := range rbs.items {
        rb.selected = (rb == target)
    }
}

func (rbs *RadioButtons) Insert(v View)               { rbs.group.Insert(v) }
func (rbs *RadioButtons) Remove(v View)               { rbs.group.Remove(v) }
func (rbs *RadioButtons) Children() []View            { return rbs.group.Children() }
func (rbs *RadioButtons) FocusedChild() View          { return rbs.group.FocusedChild() }
func (rbs *RadioButtons) SetFocusedChild(v View)      { rbs.group.SetFocusedChild(v) }
func (rbs *RadioButtons) ExecView(v View) CommandCode { return rbs.group.ExecView(v) }
func (rbs *RadioButtons) BringToFront(v View)         { rbs.group.BringToFront(v) }

func (rbs *RadioButtons) Draw(buf *DrawBuffer) {
    rbs.group.Draw(buf)
}

func (rbs *RadioButtons) HandleEvent(event *Event) {
    if event.What == EvKeyboard && event.Key != nil {
        // Alt+shortcut
        if event.Key.Modifiers&tcell.ModAlt != 0 && event.Key.Key == tcell.KeyRune {
            r := unicode.ToLower(event.Key.Rune)
            for _, rb := range rbs.items {
                if rb.shortcut != 0 && unicode.ToLower(rb.shortcut) == r {
                    rbs.group.SetFocusedChild(rb)
                    rbs.selectItem(rb)
                    event.Clear()
                    return
                }
            }
        }
        // Up arrow: move to previous radio button and select
        if event.Key.Key == tcell.KeyUp {
            idx := rbs.Value()
            if idx > 0 {
                rbs.SetValue(idx - 1)
                rbs.group.SetFocusedChild(rbs.items[idx-1])
            }
            event.Clear()
            return
        }
        // Down arrow: move to next radio button and select
        if event.Key.Key == tcell.KeyDown {
            idx := rbs.Value()
            if idx < len(rbs.items)-1 {
                rbs.SetValue(idx + 1)
                rbs.group.SetFocusedChild(rbs.items[idx+1])
            }
            event.Clear()
            return
        }
    }
    rbs.group.HandleEvent(event)
}
```

**Run tests:** `go test ./tv/... -run TestRadio -v`

**Commit:** `git commit -m "feat(tv): add RadioButton widget and RadioButtons cluster"`

---

### Task 4: Integration Checkpoint — Form Widgets in Dialogs

**Purpose:** Verify that InputLine, CheckBoxes, and RadioButtons work correctly when placed inside Dialog containers, with proper focus traversal, event dispatch, and color scheme inheritance.

**Requirements (for test writer):**
- An InputLine inside a Dialog receives keystrokes through the full dispatch chain (Application → Desktop → Dialog → Group → InputLine)
- Tab cycles focus between InputLine, CheckBoxes, and Button widgets inside a Dialog
- A CheckBox inside a Dialog toggles when Space is pressed while it has focus
- A RadioButtons cluster inside a Dialog allows Up/Down selection between radio buttons
- Alt+shortcut on a Label that links to an InputLine focuses the InputLine inside a Dialog
- ColorScheme inheritance works: InputLine inside a Dialog with a custom ColorScheme uses that scheme's InputNormal/InputSelection styles
- ExecView on a Dialog containing form widgets returns the correct command code when OK/Cancel is pressed
- InputLine text is preserved during modal dialog execution (type text, press OK, check Text() value before dialog closes)

**Components to wire up:** Application (with SimulationScreen), Desktop, Dialog, InputLine, CheckBoxes, RadioButtons, Button, Label (all real, no mocks)

**Run:** `go test ./tv/... -run TestIntegrationPhase5 -v`

**Commit:** `git commit -m "test(tv): add Phase 5 integration tests for form widgets in dialogs"`

---

### Task 5: InputBox Dialog Function

**Files:**
- Modify: `tv/dialog.go`
- Test: `tv/dialog_input_box_test.go`

**Requirements:**

Function signature:
- `InputBox(owner Container, title, prompt, defaultValue string) (string, CommandCode)`
- Creates a Dialog containing: a Label (for prompt), an InputLine (for text entry), OK and Cancel buttons
- The Label's tilde shortcut (if any) focuses the InputLine
- The InputLine is pre-filled with `defaultValue` and has the cursor at the end
- OK button is the default button (responds to Enter via postprocess)
- Calls `owner.ExecView(dlg)` and returns the InputLine's text + the closing command code
- On CmCancel, returns empty string and CmCancel
- On CmOK, returns the InputLine's current text and CmOK

Auto-sizing:
- Dialog width: max(len(prompt)+6, len(title)+6, 30), capped at 60
- Dialog height: 7 (frame top + prompt row + gap + input row + gap + button row + frame bottom)
- Centered in owner's bounds (same centering logic as MessageBox)

Layout inside dialog (client area coordinates):
- Label at (1, 0, innerW, 1)
- InputLine at (1, 1, innerW, 1) with maxLen=255
- OK button at (centered, 3, 12, 2)
- Cancel button next to OK

Event flow:
- Enter key (when InputLine has focus) triggers OK via the default button's postprocess handling
- Escape key triggers CmCancel via the dialog's modal loop
- The user can Tab between InputLine, OK, and Cancel

**Implementation:**

```go
func InputBox(owner Container, title, prompt, defaultValue string) (string, CommandCode) {
    promptRunes := []rune(prompt)
    titleRunes := []rune(title)

    contentW := len(promptRunes) + 2
    if tw := len(titleRunes) + 4; tw > contentW {
        contentW = tw
    }
    dialogW := contentW + 4
    if dialogW < 30 {
        dialogW = 30
    }
    if dialogW > 60 {
        dialogW = 60
    }
    dialogH := 7
    innerW := dialogW - 4

    ob := owner.Bounds()
    dx := (ob.Width() - dialogW) / 2
    dy := (ob.Height() - dialogH) / 2
    if dx < 0 { dx = 0 }
    if dy < 0 { dy = 0 }

    dlg := NewDialog(NewRect(dx, dy, dialogW, dialogH), title)

    input := NewInputLine(NewRect(1, 1, innerW, 1), 255)
    input.SetText(defaultValue)

    lbl := NewLabel(NewRect(1, 0, innerW, 1), prompt, input)
    dlg.Insert(lbl)
    dlg.Insert(input)

    btnW := 12
    btnGap := 2
    totalBtnW := 2*btnW + btnGap
    startX := (innerW - totalBtnW) / 2
    if startX < 0 { startX = 0 }

    okBtn := NewButton(NewRect(startX, 3, btnW, 2), "OK", CmOK, WithDefault())
    cancelBtn := NewButton(NewRect(startX+btnW+btnGap, 3, btnW, 2), "Cancel", CmCancel)
    dlg.Insert(okBtn)
    dlg.Insert(cancelBtn)
    dlg.SetFocusedChild(input)

    result := owner.ExecView(dlg)
    if result == CmCancel {
        return "", CmCancel
    }
    return input.Text(), result
}
```

**Run tests:** `go test ./tv/... -run TestInputBox -v`

**Commit:** `git commit -m "feat(tv): add InputBox dialog convenience function"`

---

### Task 6: Demo App Update and E2E Tests

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`
- Test: `e2e/e2e_test.go`

**Requirements:**

**Demo app changes:**
- Add a new keyboard shortcut F3 (via StatusLine item and onCommand handler) that opens an InputBox dialog
- The InputBox asks for a "File name" with prompt "~N~ame:" and default value "untitled.txt"
- After the InputBox returns with CmOK, the result text is displayed in win1's StaticText (update the StaticText content)
- Add a CheckBoxes cluster to win1 with labels: `["~R~ead only", "~H~idden", "~S~ystem"]`
- Add a RadioButtons cluster to win1 with labels: `["~T~ext", "~B~inary", "~H~ex"]`

**New E2E tests:**

1. `TestInputBoxFlow` — F3 opens InputBox dialog, "Name:" prompt visible, type text, Enter confirms, dialog closes
2. `TestCheckBoxVisible` — CheckBoxes cluster visible in win1 with checkbox indicators `[ ]`
3. `TestRadioButtonVisible` — RadioButtons cluster visible in win1 with radio indicators `(*)`
4. `TestInputBoxCancel` — F3 opens InputBox, Escape cancels, dialog closes, app still running

**E2E test approach:**
- Build binary via `buildBasicApp(t)`
- Launch in tmux session
- Send keystrokes via `tmuxSendKeys`
- Capture pane content via `tmuxCapture`
- Assert on visible text content

**Implementation guidance for demo app:**

```go
// In main.go, add to imports and modify onCommand handler:

// Add to statusLine:
tv.NewStatusItem("~F3~ Input", tv.KbFunc(3), tv.CmUser+10),

// Add widgets to win1:
checkboxes := tv.NewCheckBoxes(tv.NewRect(1, 5, 25, 3), []string{"~R~ead only", "~H~idden", "~S~ystem"})
win1.Insert(checkboxes)

radios := tv.NewRadioButtons(tv.NewRect(1, 9, 25, 3), []string{"~T~ext", "~B~inary", "~H~ex"})
win1.Insert(radios)

// In onCommand handler, add case:
if cmd == tv.CmUser+10 {
    text, result := tv.InputBox(app.Desktop(), "Open File", "~N~ame:", "untitled.txt")
    if result == tv.CmOK {
        st.SetText("File: " + text)
    }
    return true
}
```

**Run tests:** `cd e2e && go test -v -timeout 120s`

**Commit:** `git commit -m "feat(e2e): add form widgets to demo app and e2e tests for Phase 5"`
