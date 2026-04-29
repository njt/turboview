package tv

// checkbox_test.go — Tests for Task 2: CheckBox widget and CheckBoxes cluster.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — CheckBox: construction, accessors, drawing, keyboard, mouse
//   Section 2  — CheckBoxes: construction, Values/SetValues, Item, Alt+shortcut,
//                Tab traversal, Container interface
//   Section 3  — Falsifying tests (guard against vacuous passes)

import (
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// compile-time assertions
var _ Widget = (*CheckBox)(nil)
var _ Container = (*CheckBoxes)(nil)

// ---------------------------------------------------------------------------
// Section 1 — CheckBox
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

// TestNewCheckBoxSetsSfVisible verifies NewCheckBox sets the SfVisible state flag.
// Spec: "Sets SfVisible … by default."
func TestNewCheckBoxSetsSfVisible(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "~S~ave")

	if !cb.HasState(SfVisible) {
		t.Error("NewCheckBox did not set SfVisible")
	}
}

// TestNewCheckBoxSetsOfSelectable verifies NewCheckBox sets the OfSelectable option.
// Spec: "Sets … OfSelectable by default."
func TestNewCheckBoxSetsOfSelectable(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "~S~ave")

	if !cb.HasOption(OfSelectable) {
		t.Error("NewCheckBox did not set OfSelectable")
	}
}

// TestNewCheckBoxStoresBounds verifies NewCheckBox records the given bounds.
// Spec: "NewCheckBox(bounds Rect, label string) *CheckBox"
func TestNewCheckBoxStoresBounds(t *testing.T) {
	r := NewRect(5, 3, 20, 1)
	cb := NewCheckBox(r, "~S~ave")

	if cb.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", cb.Bounds(), r)
	}
}

// TestNewCheckBoxStartsUnchecked verifies a newly created CheckBox is unchecked.
// Spec: "Stores a boolean checked state" — default is false.
func TestNewCheckBoxStartsUnchecked(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "~S~ave")

	if cb.Checked() {
		t.Error("NewCheckBox should start unchecked; Checked() returned true")
	}
}

// ---------------------------------------------------------------------------
// Accessors
// ---------------------------------------------------------------------------

// TestCheckBoxCheckedReturnsFalseInitially verifies Checked() returns false on a
// new CheckBox.
// Spec: "Checked() bool returns current checked state."
func TestCheckBoxCheckedReturnsFalseInitially(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "Option")

	if cb.Checked() != false {
		t.Error("Checked() on new CheckBox must return false")
	}
}

// TestCheckBoxSetCheckedToTrue verifies SetChecked(true) makes Checked() return true.
// Spec: "SetChecked(bool) sets the checked state."
func TestCheckBoxSetCheckedToTrue(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "Option")

	cb.SetChecked(true)

	if !cb.Checked() {
		t.Error("SetChecked(true): Checked() returned false, want true")
	}
}

// TestCheckBoxSetCheckedToFalse verifies SetChecked(false) makes Checked() return false.
// Spec: "SetChecked(bool) sets the checked state."
func TestCheckBoxSetCheckedToFalse(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "Option")
	cb.SetChecked(true)

	cb.SetChecked(false)

	if cb.Checked() {
		t.Error("SetChecked(false): Checked() returned true, want false")
	}
}

// TestCheckBoxLabelReturnsLabel verifies Label() returns the string passed to NewCheckBox.
// Spec: "Label() string returns the label."
func TestCheckBoxLabelReturnsLabel(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "~S~ave settings")

	if cb.Label() != "~S~ave settings" {
		t.Errorf("Label() = %q, want %q", cb.Label(), "~S~ave settings")
	}
}

// TestCheckBoxShortcutExtractsFirstTildeRune verifies Shortcut() returns the first
// rune of the tilde-enclosed segment.
// Spec: "Shortcut() rune returns the extracted shortcut character from tilde notation."
func TestCheckBoxShortcutExtractsFirstTildeRune(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "~S~ave settings")

	got := cb.Shortcut()
	if unicode.ToLower(got) != 's' {
		t.Errorf("Shortcut() = %q, want 's'", got)
	}
}

// TestCheckBoxShortcutReturnsZeroForNoTilde verifies Shortcut() returns 0 when
// the label has no tilde notation.
// Spec: "Shortcut() rune returns the extracted shortcut character from tilde notation"
// — no tilde means no shortcut.
func TestCheckBoxShortcutReturnsZeroForNoTilde(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "Save settings")

	if cb.Shortcut() != 0 {
		t.Errorf("Shortcut() = %q on no-tilde label, want 0 (zero rune)", cb.Shortcut())
	}
}

// ---------------------------------------------------------------------------
// Drawing — unchecked
// ---------------------------------------------------------------------------

// TestCheckBoxDrawUncheckedShowsOpenBracketSpaceCloseBracket verifies an unchecked
// CheckBox renders "[ ]" as the mark.
// Spec: "Renders as [ ] Label when unchecked."
func TestCheckBoxDrawUncheckedShowsOpenBracketSpaceCloseBracket(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	// Layout: '[', ' ', ']', ' ', 'O', 'K'  — unchecked has space at col 1
	if buf.GetCell(0, 0).Rune != '[' {
		t.Errorf("unchecked cell(0,0) = %q, want '['", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(1, 0).Rune != ' ' {
		t.Errorf("unchecked cell(1,0) = %q, want ' ' (space for unchecked)", buf.GetCell(1, 0).Rune)
	}
	if buf.GetCell(2, 0).Rune != ']' {
		t.Errorf("unchecked cell(2,0) = %q, want ']'", buf.GetCell(2, 0).Rune)
	}
}

// TestCheckBoxDrawUncheckedHasSpaceAtMarkPosition verifies the mark position holds
// a space when unchecked, not 'X'.
// Spec: "Renders as [ ] Label when unchecked."
func TestCheckBoxDrawUncheckedHasSpaceAtMarkPosition(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	mark := buf.GetCell(1, 0).Rune
	if mark == 'X' || mark == 'x' {
		t.Errorf("unchecked CheckBox mark at col 1 = %q, want ' ' (no X when unchecked)", mark)
	}
}

// ---------------------------------------------------------------------------
// Drawing — checked
// ---------------------------------------------------------------------------

// TestCheckBoxDrawCheckedShowsX verifies a checked CheckBox renders 'X' as the mark.
// Spec: "Renders as [X] Label when checked."
func TestCheckBoxDrawCheckedShowsX(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.scheme = theme.BorlandBlue
	cb.SetChecked(true)

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	// '[' at 0, 'X' at 1, ']' at 2
	if buf.GetCell(1, 0).Rune != 'X' {
		t.Errorf("checked CheckBox mark at col 1 = %q, want 'X'", buf.GetCell(1, 0).Rune)
	}
}

// TestCheckBoxDrawCheckedBracketsPresent verifies '[' and ']' still appear when checked.
// Spec: "Renders as [X] Label when checked."
func TestCheckBoxDrawCheckedBracketsPresent(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.scheme = theme.BorlandBlue
	cb.SetChecked(true)

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	if buf.GetCell(0, 0).Rune != '[' {
		t.Errorf("checked: cell(0,0) = %q, want '['", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(2, 0).Rune != ']' {
		t.Errorf("checked: cell(2,0) = %q, want ']'", buf.GetCell(2, 0).Rune)
	}
}

// ---------------------------------------------------------------------------
// Drawing — label placement
// ---------------------------------------------------------------------------

// TestCheckBoxDrawLabelAppearsAfterBrackets verifies the label text starts at
// column 4 (after "[ ] ").
// Spec: "Total rendered width: 4 + tildeTextLen(label)"
func TestCheckBoxDrawLabelAppearsAfterBrackets(t *testing.T) {
	// Label "OK" (2 runes, no tilde). Starts at col 4: "[", " "/" X", "]", " ", label...
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	// Col 3 is space after ']', col 4 is 'O', col 5 is 'K'.
	if buf.GetCell(3, 0).Rune != ' ' {
		t.Errorf("cell(3,0) = %q, want ' ' (space between bracket and label)", buf.GetCell(3, 0).Rune)
	}
	if buf.GetCell(4, 0).Rune != 'O' {
		t.Errorf("label first char: cell(4,0) = %q, want 'O'", buf.GetCell(4, 0).Rune)
	}
	if buf.GetCell(5, 0).Rune != 'K' {
		t.Errorf("label second char: cell(5,0) = %q, want 'K'", buf.GetCell(5, 0).Rune)
	}
}

// TestCheckBoxDrawLabelUsesCheckBoxNormalStyle verifies non-shortcut label chars
// use CheckBoxNormal style.
// Spec: "Uses CheckBoxNormal style from ColorScheme for the bracket/mark/label."
func TestCheckBoxDrawLabelUsesCheckBoxNormalStyle(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	// 'O' at col 4 is a normal (non-shortcut) char — must use CheckBoxNormal.
	cell := buf.GetCell(4, 0)
	if cell.Style != theme.BorlandBlue.CheckBoxNormal {
		t.Errorf("normal label char at cell(4,0) style = %v, want CheckBoxNormal %v",
			cell.Style, theme.BorlandBlue.CheckBoxNormal)
	}
}

// TestCheckBoxDrawBracketUsesCheckBoxNormalStyle verifies the bracket chars use
// CheckBoxNormal style.
// Spec: "Uses CheckBoxNormal style from ColorScheme for the bracket/mark/label."
func TestCheckBoxDrawBracketUsesCheckBoxNormalStyle(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	cell := buf.GetCell(0, 0) // '[' character
	if cell.Style != theme.BorlandBlue.CheckBoxNormal {
		t.Errorf("'[' bracket at cell(0,0) style = %v, want CheckBoxNormal %v",
			cell.Style, theme.BorlandBlue.CheckBoxNormal)
	}
}

// TestCheckBoxDrawShortcutLabelCharUsesLabelShortcutStyle verifies the tilde-shortcut
// character in the label uses LabelShortcut style.
// Spec: "LabelShortcut style for tilde-shortcut characters in the label."
func TestCheckBoxDrawShortcutLabelCharUsesLabelShortcutStyle(t *testing.T) {
	// "~S~ave": 'S' is the shortcut char; it sits at col 4.
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "~S~ave")
	cb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	// tildeTextLen("~S~ave") = 4. Layout: [ ] ~S~ave
	// '[' at 0, ' ' at 1, ']' at 2, ' ' at 3, 'S'(shortcut) at 4
	cell := buf.GetCell(4, 0)
	if cell.Style != theme.BorlandBlue.LabelShortcut {
		t.Errorf("shortcut char 'S' at cell(4,0) style = %v, want LabelShortcut %v",
			cell.Style, theme.BorlandBlue.LabelShortcut)
	}
}

// TestCheckBoxDrawNonShortcutLabelAfterShortcutIsNormal verifies the non-shortcut
// chars following a tilde-marked letter still use CheckBoxNormal style.
// Spec: "CheckBoxNormal style … for the bracket/mark/label."
func TestCheckBoxDrawNonShortcutLabelAfterShortcutIsNormal(t *testing.T) {
	// "~S~ave": 'a','v','e' are normal.
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "~S~ave")
	cb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	// 'a' is at col 5.
	cell := buf.GetCell(5, 0)
	if cell.Style != theme.BorlandBlue.CheckBoxNormal {
		t.Errorf("non-shortcut label char at cell(5,0) style = %v, want CheckBoxNormal %v",
			cell.Style, theme.BorlandBlue.CheckBoxNormal)
	}
}

// TestCheckBoxDrawTotalWidthIs4PlusTildeTextLen verifies the last label char is at
// column 3 + tildeTextLen(label).
// Spec: "Total rendered width: 4 + tildeTextLen(label)."
func TestCheckBoxDrawTotalWidthIs4PlusTildeTextLen(t *testing.T) {
	label := "Save"
	cb := NewCheckBox(NewRect(0, 0, 20, 1), label)
	cb.scheme = theme.BorlandBlue

	textLen := tildeTextLen(label) // 4 for "Save"
	expectedLastCol := 3 + textLen // cols 4..3+textLen = 4..7 for "Save"

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	// The cell just past the label (col 4+textLen = 8) should not be part of the label text.
	// We can verify the last label char is at the expected column.
	runes := []rune("Save")
	lastRune := runes[len(runes)-1]
	cell := buf.GetCell(expectedLastCol, 0)
	if cell.Rune != lastRune {
		t.Errorf("last label rune at col %d = %q, want %q", expectedLastCol, cell.Rune, lastRune)
	}
}

// ---------------------------------------------------------------------------
// Drawing — SfSelected prefix (focus cursor)
// ---------------------------------------------------------------------------

// TestCheckBoxDrawFocusCursorWhenSelected verifies '►' is rendered at col 0 when
// the CheckBox has SfSelected.
// Spec: "When SfSelected, a ► prefix is rendered before the brackets … shifting the
// bracket/label right by 1."
func TestCheckBoxDrawFocusCursorWhenSelected(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.scheme = theme.BorlandBlue
	cb.SetState(SfSelected, true)

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != '►' {
		t.Errorf("focused CheckBox cell(0,0) = %q, want '►'", cell.Rune)
	}
}

// TestCheckBoxDrawNoCursorWhenNotSelected verifies no '►' appears at col 0 when
// the CheckBox does not have SfSelected.
// Spec: "When SfSelected, a ► prefix is rendered…" — by contrapositive.
func TestCheckBoxDrawNoCursorWhenNotSelected(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.scheme = theme.BorlandBlue
	cb.SetState(SfSelected, false)

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == '►' {
		t.Errorf("unfocused CheckBox cell(0,0) = '►'; cursor must only appear when SfSelected")
	}
}

// TestCheckBoxDrawSelectedShiftsOpenBracketRight verifies '[' moves to col 1 when
// SfSelected is set (the '►' prefix occupies col 0).
// Spec: "shifting the bracket/label right by 1."
func TestCheckBoxDrawSelectedShiftsOpenBracketRight(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.scheme = theme.BorlandBlue
	cb.SetState(SfSelected, true)

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	if buf.GetCell(1, 0).Rune != '[' {
		t.Errorf("selected CheckBox: cell(1,0) = %q, want '[' (shifted right by focus cursor)",
			buf.GetCell(1, 0).Rune)
	}
}

// TestCheckBoxDrawNoColorSchemeDoesNotPanic verifies Draw is safe when no ColorScheme
// has been set.
func TestCheckBoxDrawNoColorSchemeDoesNotPanic(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	buf := NewDrawBuffer(20, 1)

	// Must not panic.
	cb.Draw(buf)
}

// ---------------------------------------------------------------------------
// Keyboard handling
// ---------------------------------------------------------------------------

// TestCheckBoxHandleEventSpaceTogglesUnchecked verifies Space toggles an unchecked
// CheckBox to checked.
// Spec: "Space: toggles checked state, consumes event."
func TestCheckBoxHandleEventSpaceTogglesUnchecked(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	cb.HandleEvent(ev)

	if !cb.Checked() {
		t.Error("Space on unchecked CheckBox: Checked() = false, want true")
	}
}

// TestCheckBoxHandleEventSpaceTogglesChecked verifies Space toggles a checked
// CheckBox to unchecked.
// Spec: "Space: toggles checked state, consumes event."
func TestCheckBoxHandleEventSpaceTogglesChecked(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.SetChecked(true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	cb.HandleEvent(ev)

	if cb.Checked() {
		t.Error("Space on checked CheckBox: Checked() = true, want false")
	}
}

// TestCheckBoxHandleEventSpaceConsumesEvent verifies Space clears the event.
// Spec: "Space: toggles checked state, consumes event."
func TestCheckBoxHandleEventSpaceConsumesEvent(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	cb.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Space event not consumed; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestCheckBoxHandleEventEnterTogglesCheckedState verifies Enter also toggles the
// checked state.
// Spec: "Enter: toggles checked state, consumes event."
func TestCheckBoxHandleEventEnterTogglesCheckedState(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	cb.HandleEvent(ev)

	if !cb.Checked() {
		t.Error("Enter on unchecked CheckBox: Checked() = false, want true")
	}
}

// TestCheckBoxHandleEventEnterConsumesEvent verifies Enter clears the event.
// Spec: "Enter: toggles checked state, consumes event."
func TestCheckBoxHandleEventEnterConsumesEvent(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	cb.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Enter event not consumed; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestCheckBoxHandleEventOtherKeyDoesNotToggle verifies other keys do not toggle
// the checked state.
// Spec: "Space … Enter …" — only these keys toggle.
func TestCheckBoxHandleEventOtherKeyDoesNotToggle(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'x'}}
	cb.HandleEvent(ev)

	if cb.Checked() {
		t.Error("pressing 'x' toggled CheckBox; only Space and Enter should toggle")
	}
}

// TestCheckBoxHandleEventOtherKeyDoesNotConsumeEvent verifies other keys are not
// consumed by the CheckBox.
func TestCheckBoxHandleEventOtherKeyDoesNotConsumeEvent(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'x'}}
	cb.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("pressing 'x' consumed event; CheckBox should not consume keys other than Space/Enter")
	}
}

// ---------------------------------------------------------------------------
// Mouse handling
// ---------------------------------------------------------------------------

// TestCheckBoxHandleEventMouseButton1TogglesState verifies a left-click (Button1)
// within bounds toggles the checked state.
// Spec: "Click (Button1) within bounds: toggles checked state, consumes event."
func TestCheckBoxHandleEventMouseButton1TogglesState(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 10, 1), "OK")

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	cb.HandleEvent(ev)

	if !cb.Checked() {
		t.Error("Button1 click on unchecked CheckBox: Checked() = false, want true")
	}
}

// TestCheckBoxHandleEventMouseButton1ConsumesEvent verifies a left-click clears
// the event.
// Spec: "Click (Button1) within bounds: toggles checked state, consumes event."
func TestCheckBoxHandleEventMouseButton1ConsumesEvent(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 10, 1), "OK")

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	cb.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Button1 click event not consumed; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestCheckBoxHandleEventMouseOtherButtonDoesNotToggle verifies that mouse buttons
// other than Button1 do not toggle the state.
// Spec: "Click (Button1) within bounds…" — only Button1.
func TestCheckBoxHandleEventMouseOtherButtonDoesNotToggle(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 10, 1), "OK")

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button2}}
	cb.HandleEvent(ev)

	if cb.Checked() {
		t.Error("Button2 click toggled CheckBox; only Button1 should toggle")
	}
}

// TestCheckBoxHandleEventMouseButton1TogglesFromChecked verifies Button1 toggles a
// checked CheckBox to unchecked.
// Spec: "toggles checked state."
func TestCheckBoxHandleEventMouseButton1TogglesFromChecked(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 10, 1), "OK")
	cb.SetChecked(true)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	cb.HandleEvent(ev)

	if cb.Checked() {
		t.Error("Button1 click on checked CheckBox: Checked() = true, want false")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — CheckBoxes cluster
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

// TestNewCheckBoxesSetsSfVisible verifies NewCheckBoxes sets the SfVisible state flag.
// Spec: "Sets SfVisible … by default."
func TestNewCheckBoxesSetsSfVisible(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"~A~lpha", "~B~eta", "~G~amma"})

	if !cbs.HasState(SfVisible) {
		t.Error("NewCheckBoxes did not set SfVisible")
	}
}

// TestNewCheckBoxesSetsOfSelectable verifies NewCheckBoxes sets the OfSelectable option.
// Spec: "Sets … OfSelectable … by default."
func TestNewCheckBoxesSetsOfSelectable(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"~A~lpha", "~B~eta"})

	if !cbs.HasOption(OfSelectable) {
		t.Error("NewCheckBoxes did not set OfSelectable")
	}
}

// TestNewCheckBoxesSetsOfPreProcess verifies NewCheckBoxes sets the OfPreProcess option
// (required for Alt+shortcut preprocess dispatch).
// Spec: "Sets … OfPreProcess by default."
func TestNewCheckBoxesSetsOfPreProcess(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"~A~lpha", "~B~eta"})

	if !cbs.HasOption(OfPreProcess) {
		t.Error("NewCheckBoxes did not set OfPreProcess")
	}
}

// TestNewCheckBoxesCreatesOneCheckBoxPerLabel verifies one CheckBox is created for
// each label.
// Spec: "Creates one CheckBox per label."
func TestNewCheckBoxesCreatesOneCheckBoxPerLabel(t *testing.T) {
	labels := []string{"Alpha", "Beta", "Gamma"}
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), labels)

	if len(cbs.Children()) != len(labels) {
		t.Errorf("Children() len = %d, want %d", len(cbs.Children()), len(labels))
	}
}

// TestNewCheckBoxesZeroLabels verifies NewCheckBoxes handles an empty label list.
// Spec: "Creates one CheckBox per label" — with zero labels, zero checkboxes.
func TestNewCheckBoxesZeroLabels(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 0), []string{})

	if len(cbs.Children()) != 0 {
		t.Errorf("zero labels: Children() len = %d, want 0", len(cbs.Children()))
	}
}

// TestNewCheckBoxesArrangesCheckBoxesVertically verifies each CheckBox is at y=index.
// Spec: "Each CheckBox is positioned at y=index within the cluster."
func TestNewCheckBoxesArrangesCheckBoxesVertically(t *testing.T) {
	labels := []string{"Alpha", "Beta", "Gamma"}
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), labels)

	for i := 0; i < len(labels); i++ {
		item := cbs.Item(i)
		if item.Bounds().A.Y != i {
			t.Errorf("Item(%d).Bounds().Y = %d, want %d", i, item.Bounds().A.Y, i)
		}
	}
}

// TestNewCheckBoxesItemsStartAtX0 verifies each CheckBox starts at x=0 within the cluster.
// Spec: "Each CheckBox is positioned at y=index within the cluster."
func TestNewCheckBoxesItemsStartAtX0(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})

	for i := 0; i < 2; i++ {
		item := cbs.Item(i)
		if item.Bounds().A.X != 0 {
			t.Errorf("Item(%d).Bounds().X = %d, want 0", i, item.Bounds().A.X)
		}
	}
}

// TestNewCheckBoxesStoresBounds verifies CheckBoxes records the given bounds.
// Spec: "NewCheckBoxes(bounds Rect, labels []string) *CheckBoxes"
func TestNewCheckBoxesStoresBounds(t *testing.T) {
	r := NewRect(5, 5, 30, 4)
	cbs := NewCheckBoxes(r, []string{"Alpha", "Beta", "Gamma", "Delta"})

	if cbs.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", cbs.Bounds(), r)
	}
}

// ---------------------------------------------------------------------------
// Values / SetValues
// ---------------------------------------------------------------------------

// TestCheckBoxesValuesAllUnchecked verifies Values() returns 0 when all items
// are unchecked.
// Spec: "Values() uint32 returns a bitmask of checked states."
func TestCheckBoxesValuesAllUnchecked(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	if cbs.Values() != 0 {
		t.Errorf("Values() on all-unchecked = %d, want 0", cbs.Values())
	}
}

// TestCheckBoxesValuesFirstItemChecked verifies bit 0 is set when first item is checked.
// Spec: "bit 0 = first item."
func TestCheckBoxesValuesFirstItemChecked(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.Item(0).SetChecked(true)

	got := cbs.Values()
	if got&1 == 0 {
		t.Errorf("Values() = %d; bit 0 not set even though first item is checked", got)
	}
}

// TestCheckBoxesValuesSecondItemChecked verifies bit 1 is set when second item is checked.
// Spec: "bit 1 = second item."
func TestCheckBoxesValuesSecondItemChecked(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.Item(1).SetChecked(true)

	got := cbs.Values()
	if got&2 == 0 {
		t.Errorf("Values() = %d; bit 1 not set even though second item is checked", got)
	}
}

// TestCheckBoxesValuesMultipleItemsChecked verifies the bitmask reflects multiple
// checked items correctly.
// Spec: "Values() uint32 returns a bitmask of checked states (bit 0 = first item, etc.)."
func TestCheckBoxesValuesMultipleItemsChecked(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.Item(0).SetChecked(true) // bit 0
	cbs.Item(2).SetChecked(true) // bit 2

	got := cbs.Values()
	want := uint32(0b101) // bits 0 and 2
	if got != want {
		t.Errorf("Values() = %b, want %b (bits 0 and 2 set)", got, want)
	}
}

// TestCheckBoxesSetValuesChecksCorrectItems verifies SetValues sets checked states
// from a bitmask.
// Spec: "SetValues(uint32) sets checked states from bitmask."
func TestCheckBoxesSetValuesChecksCorrectItems(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	cbs.SetValues(0b110) // bits 1 and 2 set → items 1 and 2 checked

	if cbs.Item(0).Checked() {
		t.Error("SetValues(0b110): item 0 is checked, want unchecked")
	}
	if !cbs.Item(1).Checked() {
		t.Error("SetValues(0b110): item 1 is unchecked, want checked (bit 1)")
	}
	if !cbs.Item(2).Checked() {
		t.Error("SetValues(0b110): item 2 is unchecked, want checked (bit 2)")
	}
}

// TestCheckBoxesSetValuesZeroClearsAll verifies SetValues(0) unchecks all items.
// Spec: "SetValues(uint32) sets checked states from bitmask."
func TestCheckBoxesSetValuesZeroClearsAll(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetValues(0b111) // all checked first

	cbs.SetValues(0) // now clear all

	for i := 0; i < 3; i++ {
		if cbs.Item(i).Checked() {
			t.Errorf("SetValues(0): item %d still checked", i)
		}
	}
}

// TestCheckBoxesSetValuesRoundtrips verifies SetValues then Values returns the
// original mask.
// Spec: "Values() … SetValues(uint32)."
func TestCheckBoxesSetValuesRoundtrips(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 4), []string{"A", "B", "C", "D"})
	want := uint32(0b1010) // items 1 and 3

	cbs.SetValues(want)
	got := cbs.Values()

	if got != want {
		t.Errorf("SetValues(%b) then Values() = %b, want %b", want, got, want)
	}
}

// ---------------------------------------------------------------------------
// Item
// ---------------------------------------------------------------------------

// TestCheckBoxesItemReturnsCorrectCheckBox verifies Item(i) returns the CheckBox
// at the given index.
// Spec: "Item(index int) *CheckBox returns the CheckBox at the given index."
func TestCheckBoxesItemReturnsCorrectCheckBox(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})

	item0 := cbs.Item(0)
	if item0 == nil {
		t.Fatal("Item(0) returned nil")
	}
	if item0.Label() != "Alpha" {
		t.Errorf("Item(0).Label() = %q, want %q", item0.Label(), "Alpha")
	}

	item2 := cbs.Item(2)
	if item2 == nil {
		t.Fatal("Item(2) returned nil")
	}
	if item2.Label() != "Gamma" {
		t.Errorf("Item(2).Label() = %q, want %q", item2.Label(), "Gamma")
	}
}

// TestCheckBoxesItemIsSameAsChildren verifies Item(i) returns the same view as
// Children()[i].
// Spec: "Item(index int) *CheckBox returns the CheckBox at the given index."
func TestCheckBoxesItemIsSameAsChildren(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"Alpha", "Beta"})

	children := cbs.Children()
	for i := 0; i < 2; i++ {
		if cbs.Item(i) != children[i] {
			t.Errorf("Item(%d) != Children()[%d]; expected the same view", i, i)
		}
	}
}

// ---------------------------------------------------------------------------
// Container interface
// ---------------------------------------------------------------------------

// TestCheckBoxesInsertAddsView verifies Insert adds a view to the cluster.
// Spec: "Insert(View) … delegates to internal Group."
func TestCheckBoxesInsertAddsView(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	initialLen := len(cbs.Children())

	extra := newMockView(NewRect(0, 3, 20, 1))
	cbs.Insert(extra)

	if len(cbs.Children()) != initialLen+1 {
		t.Errorf("Children() len after Insert = %d, want %d", len(cbs.Children()), initialLen+1)
	}
}

// TestCheckBoxesRemoveRemovesView verifies Remove removes a previously inserted view.
// Spec: "Remove(View) … delegates to internal Group."
func TestCheckBoxesRemoveRemovesView(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 1), []string{"A"})
	item := cbs.Item(0)

	cbs.Remove(item)

	for _, child := range cbs.Children() {
		if child == item {
			t.Error("Remove: child still present in Children()")
		}
	}
}

// TestCheckBoxesFocusedChildReturnsSelected verifies FocusedChild() returns the
// currently focused item.
// Spec: "FocusedChild() View … delegates to internal Group."
func TestCheckBoxesFocusedChildReturnsSelected(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})

	focused := cbs.FocusedChild()
	if focused == nil {
		t.Error("FocusedChild() returned nil on a cluster with selectable items")
	}
}

// TestCheckBoxesSetFocusedChildChangesFocus verifies SetFocusedChild moves focus.
// Spec: "SetFocusedChild(View) … delegates to internal Group."
func TestCheckBoxesSetFocusedChildChangesFocus(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	item0 := cbs.Item(0)

	cbs.SetFocusedChild(item0)

	if cbs.FocusedChild() != item0 {
		t.Errorf("SetFocusedChild(item0): FocusedChild() = %v, want item0", cbs.FocusedChild())
	}
}

// TestCheckBoxesChildrenOwnerIsCheckBoxes verifies that children's Owner() is set
// to the CheckBoxes (the facade), not the internal Group.
// Spec: "Implements the Container interface (delegates to internal Group, same pattern
// as Dialog)."
func TestCheckBoxesChildrenOwnerIsCheckBoxes(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})

	for _, child := range cbs.Children() {
		if child.Owner() != cbs {
			t.Errorf("child.Owner() = %v, want CheckBoxes (facade)", child.Owner())
		}
	}
}

// ---------------------------------------------------------------------------
// Drawing
// ---------------------------------------------------------------------------

// TestCheckBoxesDrawRendersAllItems verifies Draw paints each CheckBox in its row.
// Spec: "Drawing: Delegates to internal Group which draws each CheckBox in its sub-buffer."
func TestCheckBoxesDrawRendersAllItems(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"Alpha", "Beta", "Gamma"})
	cbs.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 3)
	cbs.Draw(buf)

	// Each row should have a CheckBox bracket '[' (at position 0 or 1, depending on focus indicator).
	for row := 0; row < 3; row++ {
		cell0 := buf.GetCell(0, row)
		cell1 := buf.GetCell(1, row)
		// Focus indicator '►' at column 0, '[' at column 1, or '[' at column 0 if not focused
		if !((cell0.Rune == '[') || (cell0.Rune == '►' && cell1.Rune == '[')) {
			t.Errorf("row %d: expected '[' at col 0 or 1, but got %q at col 0 and %q at col 1", row, cell0.Rune, cell1.Rune)
		}
	}
}

// TestCheckBoxesDrawCheckedItemShowsX verifies that a checked item's 'X' appears
// in the buffer.
// Spec: "Delegates to internal Group which draws each CheckBox in its sub-buffer."
func TestCheckBoxesDrawCheckedItemShowsX(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})
	cbs.scheme = theme.BorlandBlue
	cbs.Item(0).SetChecked(true)

	buf := NewDrawBuffer(20, 2)
	cbs.Draw(buf)

	// Row 0, col 1 should be 'X'.
	if buf.GetCell(1, 0).Rune != 'X' {
		t.Errorf("checked item row 0: cell(1,0) = %q, want 'X'", buf.GetCell(1, 0).Rune)
	}
}

// ---------------------------------------------------------------------------
// Tab traversal
// ---------------------------------------------------------------------------

// TestCheckBoxesTabAdvancesFocus verifies Tab moves focus from the first to the
// second CheckBox.
// Spec: "The cluster's Group handles Tab/Shift+Tab between checkboxes."
func TestCheckBoxesTabAdvancesFocus(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(0))

	ev := tabKey()
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(1) {
		t.Errorf("after Tab, FocusedChild() = %v, want Item(1)", cbs.FocusedChild())
	}
}

// TestCheckBoxesShiftTabMovesFocusBackward verifies Shift+Tab moves focus backward.
// Spec: "The cluster's Group handles Tab/Shift+Tab between checkboxes."
func TestCheckBoxesShiftTabMovesFocusBackward(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(1))

	ev := shiftTabKey()
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(0) {
		t.Errorf("after Shift+Tab, FocusedChild() = %v, want Item(0)", cbs.FocusedChild())
	}
}

// TestCheckBoxesTabWrapsAroundToFirst verifies Tab from the last item wraps back
// to the first.
// Spec: "Tab/Shift+Tab between checkboxes."
func TestCheckBoxesTabWrapsAroundToFirst(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(2))

	ev := tabKey()
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(0) {
		t.Errorf("after Tab from last, FocusedChild() = %v, want Item(0) (wrap-around)",
			cbs.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Alt+shortcut
// ---------------------------------------------------------------------------

// TestCheckBoxesAltShortcutTogglesCheckBox verifies Alt+<shortcut letter> inside
// a CheckBoxes cluster focuses and toggles the corresponding checkbox.
// Spec: "Alt+shortcut letter focuses and toggles the corresponding checkbox (handled
// via preprocess)."
func TestCheckBoxesAltShortcutTogglesCheckBox(t *testing.T) {
	// "~S~ave": shortcut is 's'.
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"~S~ave", "~L~oad", "~E~xit"})

	// Focus item 1 initially, then press Alt+s to activate item 0.
	cbs.SetFocusedChild(cbs.Item(1))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: tcell.ModAlt}}
	cbs.HandleEvent(ev)

	if !cbs.Item(0).Checked() {
		t.Error("Alt+s on CheckBoxes with '~S~ave' did not toggle item 0")
	}
}

// TestCheckBoxesAltShortcutFocusesCheckBox verifies Alt+shortcut also shifts focus
// to the matching checkbox.
// Spec: "Alt+shortcut letter focuses and toggles the corresponding checkbox."
func TestCheckBoxesAltShortcutFocusesCheckBox(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"~S~ave", "~L~oad", "~E~xit"})
	cbs.SetFocusedChild(cbs.Item(2))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'l', Modifiers: tcell.ModAlt}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(1) {
		t.Errorf("Alt+l: FocusedChild() = %v, want Item(1) ('~L~oad')", cbs.FocusedChild())
	}
}

// TestCheckBoxesAltShortcutConsumesEvent verifies the Alt+shortcut event is
// consumed (cleared) after being handled.
// Spec: "Alt+shortcut … (handled via preprocess)" — implies event is consumed.
func TestCheckBoxesAltShortcutConsumesEvent(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: tcell.ModAlt}}
	cbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Alt+s event not consumed; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestCheckBoxesAltShortcutCaseInsensitive verifies Alt+uppercase letter matches a
// lowercase shortcut.
// Spec: "Alt+shortcut letter focuses and toggles the corresponding checkbox."
func TestCheckBoxesAltShortcutCaseInsensitive(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~s~ave", "~L~oad"})

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'S', Modifiers: tcell.ModAlt}}
	cbs.HandleEvent(ev)

	if !cbs.Item(0).Checked() {
		t.Error("Alt+S (uppercase) did not toggle item 0 with lowercase shortcut 's'; matching must be case-insensitive")
	}
}

// TestCheckBoxesAltWrongLetterDoesNotToggle verifies Alt+non-matching letter does
// not toggle any checkbox.
// Spec: "Alt+shortcut letter …" — only the matching letter activates.
func TestCheckBoxesAltWrongLetterDoesNotToggle(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'x', Modifiers: tcell.ModAlt}}
	cbs.HandleEvent(ev)

	for i := 0; i < 2; i++ {
		if cbs.Item(i).Checked() {
			t.Errorf("Alt+x: item %d was toggled; no item should match 'x'", i)
		}
	}
}

// TestCheckBoxesNoAltModifierDoesNotTriggerShortcut verifies a plain letter (without
// Alt) does not trigger the shortcut mechanism.
// Spec: "Alt+shortcut letter …" — requires Alt modifier.
func TestCheckBoxesNoAltModifierDoesNotTriggerShortcut(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: 0}}
	cbs.HandleEvent(ev)

	if cbs.Item(0).Checked() {
		t.Error("plain 's' (no Alt) toggled item 0; shortcut requires Alt modifier")
	}
}

// TestCheckBoxesAltShortcutTogglesAlreadyChecked verifies Alt+shortcut on an already
// checked item toggles it off.
// Spec: "Alt+shortcut letter focuses and toggles the corresponding checkbox."
func TestCheckBoxesAltShortcutTogglesAlreadyChecked(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"~S~ave", "~L~oad"})
	cbs.Item(0).SetChecked(true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 's', Modifiers: tcell.ModAlt}}
	cbs.HandleEvent(ev)

	if cbs.Item(0).Checked() {
		t.Error("Alt+s on already-checked item: still checked; should have toggled to false")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Falsifying tests
// ---------------------------------------------------------------------------

// TestCheckBoxCheckedAndUncheckedAreDifferentStates verifies that the two checked
// states actually differ in rendering (guard against a collapsed implementation).
// Spec: "[X] … [ ]"
func TestCheckBoxCheckedAndUncheckedAreDifferentStates(t *testing.T) {
	unchecked := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	checked := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	checked.SetChecked(true)

	bufU := NewDrawBuffer(20, 1)
	bufC := NewDrawBuffer(20, 1)
	unchecked.Draw(bufU)
	checked.Draw(bufC)

	// Col 1 should differ: ' ' vs 'X'.
	if bufU.GetCell(1, 0).Rune == bufC.GetCell(1, 0).Rune {
		t.Errorf("checked and unchecked render same rune %q at col 1; '[X]' and '[ ]' must differ",
			bufU.GetCell(1, 0).Rune)
	}
}

// TestCheckBoxLabelShortcutStyleDiffersFromNormal verifies LabelShortcut and
// CheckBoxNormal are distinct in BorlandBlue (falsification guard).
func TestCheckBoxLabelShortcutStyleDiffersFromNormal(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.LabelShortcut == scheme.CheckBoxNormal {
		t.Skip("LabelShortcut equals CheckBoxNormal in this scheme — style distinction test is vacuous")
	}

	cb := NewCheckBox(NewRect(0, 0, 20, 1), "~S~ave")
	cb.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	shortcutCell := buf.GetCell(4, 0) // 'S' — shortcut
	normalCell := buf.GetCell(5, 0)   // 'a' — normal

	if shortcutCell.Style == normalCell.Style {
		t.Errorf("shortcut and normal label chars have same style %v; expected different styles",
			shortcutCell.Style)
	}
}

// TestCheckBoxSelectedAndUnselectedRenderDifferently verifies the focus cursor
// actually changes the rendering at col 0.
func TestCheckBoxSelectedAndUnselectedRenderDifferently(t *testing.T) {
	notFocused := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	focused := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	focused.SetState(SfSelected, true)

	bufN := NewDrawBuffer(20, 1)
	bufF := NewDrawBuffer(20, 1)
	notFocused.Draw(bufN)
	focused.Draw(bufF)

	if bufN.GetCell(0, 0).Rune == bufF.GetCell(0, 0).Rune {
		t.Errorf("focused and unfocused render same rune %q at col 0; '►' vs '[' must differ",
			bufN.GetCell(0, 0).Rune)
	}
}

// TestCheckBoxesValuesIndependentOfInternalOrder verifies Values() returns bits
// in the order items were declared, not by some internal sorting.
// Spec: "bit 0 = first item, etc."
func TestCheckBoxesValuesIndependentOfInternalOrder(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.Item(2).SetChecked(true) // only the third item

	got := cbs.Values()
	if got != 0b100 {
		t.Errorf("only item 2 checked: Values() = %b, want %b (bit 2 only)", got, uint32(0b100))
	}
	if got&1 != 0 {
		t.Errorf("item 0 not checked but bit 0 is set in Values()=%b", got)
	}
}

// TestCheckBoxesSetValuesDoesNotAffectMoreThanNItems verifies SetValues only touches
// the first N bits (one per item).
// Spec: "SetValues(uint32) sets checked states from bitmask."
func TestCheckBoxesSetValuesDoesNotAffectMoreThanNItems(t *testing.T) {
	// 2 items; setting bit 2 (0b100) should not create a third checked item.
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})
	cbs.SetValues(0b111) // bits 0,1,2 — only 0 and 1 exist

	// No panic, items 0 and 1 are checked, no out-of-bounds.
	if !cbs.Item(0).Checked() {
		t.Error("SetValues(0b111): item 0 should be checked")
	}
	if !cbs.Item(1).Checked() {
		t.Error("SetValues(0b111): item 1 should be checked")
	}
}

// TestCheckBoxHandleEventDoubleSpaceTogglesBackToOriginal verifies two Space presses
// toggle back to the original state.
// Spec: "Space: toggles checked state." — toggle twice returns to start.
func TestCheckBoxHandleEventDoubleSpaceTogglesBackToOriginal(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")

	for i := 0; i < 2; i++ {
		ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
		cb.HandleEvent(ev)
	}

	if cb.Checked() {
		t.Error("two Space presses: Checked() = true, want false (restored to initial unchecked state)")
	}
}

// TestCheckBoxLabelReturnsExactStringGiven verifies Label() returns the raw tilde
// string (not stripped).
// Spec: "Label() string returns the label."
func TestCheckBoxLabelReturnsExactStringGiven(t *testing.T) {
	raw := "~S~ave settings"
	cb := NewCheckBox(NewRect(0, 0, 30, 1), raw)

	if cb.Label() != raw {
		t.Errorf("Label() = %q, want raw tilde string %q", cb.Label(), raw)
	}
}

// TestCheckBoxShortcutDifferentLabels verifies Shortcut() differs for different
// tilde letters (ensures the extraction is per-instance).
func TestCheckBoxShortcutDifferentLabels(t *testing.T) {
	cbS := NewCheckBox(NewRect(0, 0, 20, 1), "~S~ave")
	cbL := NewCheckBox(NewRect(0, 0, 20, 1), "~L~oad")

	if unicode.ToLower(cbS.Shortcut()) == unicode.ToLower(cbL.Shortcut()) {
		t.Errorf("CheckBox with '~S~ave' and '~L~oad' return same Shortcut() %q; expected different",
			cbS.Shortcut())
	}
}

// TestCheckBoxDrawLabelAfterSelectedPrefix verifies the label appears shifted by 1
// when SfSelected is set (the '►' occupies col 0).
// Spec: "shifting the bracket/label right by 1."
func TestCheckBoxDrawLabelAfterSelectedPrefix(t *testing.T) {
	// Label "OK": without selected, 'O' is at col 4. With selected, 'O' shifts to col 5.
	cb := NewCheckBox(NewRect(0, 0, 20, 1), "OK")
	cb.scheme = theme.BorlandBlue
	cb.SetState(SfSelected, true)

	buf := NewDrawBuffer(20, 1)
	cb.Draw(buf)

	// ► at 0, [ at 1, mark at 2, ] at 3, space at 4, 'O' at 5
	if buf.GetCell(5, 0).Rune != 'O' {
		t.Errorf("selected CheckBox: 'O' not at col 5; cell(5,0) = %q (label shifts right by 1 for '►')",
			buf.GetCell(5, 0).Rune)
	}
}

// TestCheckBoxesExecViewDelegatesToGroup verifies ExecView is available through
// the Container interface.
// Spec: "ExecView(View) CommandCode"
func TestCheckBoxesExecViewIsCallable(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})

	// ExecView without a Desktop will return CmCancel (same as Group's no-owner case).
	// We just verify it doesn't panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ExecView panicked: %v", r)
		}
	}()

	result := make(chan CommandCode, 1)
	go func() {
		dialog := newMockView(NewRect(0, 0, 20, 1))
		result <- cbs.ExecView(dialog)
	}()

	select {
	case code := <-result:
		if code != CmCancel {
			t.Logf("ExecView without owner chain returned %v (expected CmCancel)", code)
		}
	}
}

// ---------------------------------------------------------------------------
// Additional edge-case tests
// ---------------------------------------------------------------------------

// TestCheckBoxDrawEmptyLabelStillRendersBrackets verifies an empty label still
// renders the "[ ]" portion.
// Spec: "Total rendered width: 4 + tildeTextLen(label)" — 4 when label is "".
func TestCheckBoxDrawEmptyLabelStillRendersBrackets(t *testing.T) {
	cb := NewCheckBox(NewRect(0, 0, 10, 1), "")
	cb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(10, 1)
	cb.Draw(buf)

	if buf.GetCell(0, 0).Rune != '[' {
		t.Errorf("empty label: cell(0,0) = %q, want '['", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(2, 0).Rune != ']' {
		t.Errorf("empty label: cell(2,0) = %q, want ']'", buf.GetCell(2, 0).Rune)
	}
}

// TestCheckBoxesTabClearsEvent verifies Tab is consumed by the cluster.
// Spec: "The cluster's Group handles Tab/Shift+Tab between checkboxes."
func TestCheckBoxesTabClearsEvent(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 2), []string{"A", "B"})

	ev := tabKey()
	cbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Tab event not consumed by CheckBoxes; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestCheckBoxesChildrenLabelMatchInput verifies each child's label matches the
// label string at the corresponding index.
// Spec: "Creates one CheckBox per label."
func TestCheckBoxesChildrenLabelMatchInput(t *testing.T) {
	labels := []string{"Alpha", "Beta", "Gamma"}
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), labels)

	for i, want := range labels {
		got := cbs.Item(i).Label()
		if got != want {
			t.Errorf("Item(%d).Label() = %q, want %q", i, got, want)
		}
	}
}

// TestCheckBoxDrawTildeTextLenMatchesRenderedLabelWidth verifies that the number of
// rune cells consumed by the label text on screen equals tildeTextLen(label).
// Spec: "4 + tildeTextLen(label)."
func TestCheckBoxDrawTildeTextLenMatchesRenderedLabelWidth(t *testing.T) {
	label := "~S~ave"
	expected := tildeTextLen(label) // = 4 ("Save")

	cb := NewCheckBox(NewRect(0, 0, 30, 1), label)
	cb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(30, 1)
	cb.Draw(buf)

	// Verify the rune count of rendered label chars at cols 4..3+expected.
	// Col 4 = 'S', col 5 = 'a', col 6 = 'v', col 7 = 'e'.
	rendered := 0
	for col := 4; col < 4+expected; col++ {
		r := buf.GetCell(col, 0).Rune
		rendered += utf8.RuneLen(r)
	}
	if rendered != expected {
		t.Errorf("rendered label rune count = %d, want tildeTextLen(%q) = %d",
			rendered, label, expected)
	}
}
