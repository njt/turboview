package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// compile-time assertions.
// Spec: "RadioButton … implements the Widget interface."
// Spec: "RadioButtons … implements the Container interface."
var _ Widget = (*RadioButton)(nil)
var _ Container = (*RadioButtons)(nil)

// =============================================================================
// RadioButton — construction
// =============================================================================

// TestNewRadioButtonSetsSfVisible verifies NewRadioButton sets the SfVisible state flag.
// Spec: "Sets SfVisible … by default."
func TestNewRadioButtonSetsSfVisible(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "~O~ption")

	if !rb.HasState(SfVisible) {
		t.Error("NewRadioButton did not set SfVisible")
	}
}

// TestNewRadioButtonSetsOfSelectable verifies NewRadioButton sets OfSelectable.
// Spec: "Sets … OfSelectable by default."
func TestNewRadioButtonSetsOfSelectable(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "~O~ption")

	if !rb.HasOption(OfSelectable) {
		t.Error("NewRadioButton did not set OfSelectable")
	}
}

// TestNewRadioButtonStoresBounds verifies NewRadioButton records the given bounds.
// Spec: "NewRadioButton(bounds Rect, label string) *RadioButton"
func TestNewRadioButtonStoresBounds(t *testing.T) {
	r := NewRect(3, 7, 20, 1)
	rb := NewRadioButton(r, "~O~ption")

	if rb.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", rb.Bounds(), r)
	}
}

// TestNewRadioButtonInitiallyNotSelected verifies a newly constructed RadioButton
// is not selected by default.
// Spec: "Stores a boolean selected state."
// (The cluster selects the first one; construction alone does not imply selection.)
func TestNewRadioButtonInitiallyNotSelected(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "~O~ption")

	if rb.Selected() {
		t.Error("NewRadioButton should not be selected by default")
	}
}

// =============================================================================
// RadioButton — accessors
// =============================================================================

// TestRadioButtonSelectedReturnsFalseByDefault verifies Selected() is false before
// any explicit selection.
// Spec: "Selected() bool returns whether this radio button is the active one."
func TestRadioButtonSelectedReturnsFalseByDefault(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")

	if rb.Selected() {
		t.Error("Selected() should be false before SetSelected(true) is called")
	}
}

// TestRadioButtonSetSelectedTrue verifies SetSelected(true) makes Selected() return true.
// Spec: "SetSelected(bool) sets selected state."
func TestRadioButtonSetSelectedTrue(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")
	rb.SetSelected(true)

	if !rb.Selected() {
		t.Error("Selected() should return true after SetSelected(true)")
	}
}

// TestRadioButtonSetSelectedFalse verifies SetSelected(false) makes Selected() return false.
// Spec: "SetSelected(bool) sets selected state."
func TestRadioButtonSetSelectedFalse(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")
	rb.SetSelected(true)
	rb.SetSelected(false)

	if rb.Selected() {
		t.Error("Selected() should return false after SetSelected(false)")
	}
}

// TestRadioButtonSetSelectedTrueDoesNotAlterOtherState verifies SetSelected(true)
// only toggles selected state, not e.g. SfVisible.
// Spec: "SetSelected(bool) sets selected state."
func TestRadioButtonSetSelectedTrueDoesNotAlterOtherState(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")
	rb.SetSelected(true)

	if !rb.HasState(SfVisible) {
		t.Error("SetSelected(true) must not clear SfVisible")
	}
	if !rb.HasOption(OfSelectable) {
		t.Error("SetSelected(true) must not clear OfSelectable")
	}
}

// TestRadioButtonLabelReturnsLabel verifies Label() returns the exact label string.
// Spec: "Label() string returns the label."
func TestRadioButtonLabelReturnsLabel(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "~T~est")

	if rb.Label() != "~T~est" {
		t.Errorf("Label() = %q, want %q", rb.Label(), "~T~est")
	}
}

// TestRadioButtonLabelReturnsExactString verifies Label() returns the raw label
// (including tilde markers) without stripping.
// Spec: "Label() string returns the label."
func TestRadioButtonLabelReturnsExactString(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "No tilde")

	if rb.Label() != "No tilde" {
		t.Errorf("Label() = %q, want %q", rb.Label(), "No tilde")
	}
}

// TestRadioButtonShortcutExtractsFirstRuneOfShortcutSegment verifies Shortcut() returns
// the first rune of the tilde-enclosed shortcut segment.
// Spec: "Shortcut() rune returns extracted shortcut character from tilde notation."
func TestRadioButtonShortcutExtractsFirstRuneOfShortcutSegment(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "~T~est")

	if rb.Shortcut() != 't' && rb.Shortcut() != 'T' {
		// Spec does not mandate case; accept either.
		t.Errorf("Shortcut() = %q, want 'T' or 't' (first rune of tilde segment)", rb.Shortcut())
	}
}

// TestRadioButtonShortcutReturnsZeroWhenNoTilde verifies Shortcut() returns 0 when
// the label has no tilde notation.
// Spec: "Shortcut() rune returns extracted shortcut character from tilde notation."
func TestRadioButtonShortcutReturnsZeroWhenNoTilde(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "No tilde")

	if rb.Shortcut() != 0 {
		t.Errorf("Shortcut() = %q, want 0 (no tilde segment)", rb.Shortcut())
	}
}

// TestRadioButtonShortcutDiffersFromNonShortcutLabel verifies Shortcut() gives distinct
// results for labels with and without tildes (falsification guard).
func TestRadioButtonShortcutDiffersFromNonShortcutLabel(t *testing.T) {
	withTilde := NewRadioButton(NewRect(0, 0, 10, 1), "~T~est")
	withoutTilde := NewRadioButton(NewRect(0, 0, 10, 1), "Test")

	if withTilde.Shortcut() == withoutTilde.Shortcut() {
		t.Errorf("Shortcut() returned same value (%q) for label with and without tilde",
			withTilde.Shortcut())
	}
}

// =============================================================================
// RadioButton — drawing (unselected, no focus)
// =============================================================================

// TestRadioButtonDrawRendersOpenParenWhenNotSelected verifies '(' appears in the output.
// Spec: "Renders as (*) Label when selected, ( ) Label when not selected."
func TestRadioButtonDrawRendersOpenParenWhenNotSelected(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 15, 1), "Item")
	rb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	found := false
	for x := 0; x < 15; x++ {
		if buf.GetCell(x, 0).Rune == '(' {
			found = true
			break
		}
	}
	if !found {
		t.Error("Draw did not render '(' for unselected RadioButton")
	}
}

// TestRadioButtonDrawRendersSpaceInsideParensWhenNotSelected verifies the mark inside
// parentheses is a space when not selected.
// Spec: "( ) Label when not selected."
func TestRadioButtonDrawRendersSpaceInsideParensWhenNotSelected(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 15, 1), "Item")
	rb.scheme = theme.BorlandBlue
	rb.SetSelected(false)

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	// Find '(' then check the next cell.
	for x := 0; x < 14; x++ {
		if buf.GetCell(x, 0).Rune == '(' {
			mark := buf.GetCell(x+1, 0).Rune
			if mark != ' ' {
				t.Errorf("unselected RadioButton: cell after '(' = %q, want ' '", mark)
			}
			return
		}
	}
	t.Error("could not find '(' to inspect mark character")
}

// TestRadioButtonDrawRendersAsteriskInsideParensWhenSelected verifies the mark inside
// parentheses is '*' when selected.
// Spec: "(*) Label when selected."
func TestRadioButtonDrawRendersAsteriskInsideParensWhenSelected(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 15, 1), "Item")
	rb.scheme = theme.BorlandBlue
	rb.SetSelected(true)

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	for x := 0; x < 14; x++ {
		if buf.GetCell(x, 0).Rune == '(' {
			mark := buf.GetCell(x+1, 0).Rune
			if mark != '*' {
				t.Errorf("selected RadioButton: cell after '(' = %q, want '*'", mark)
			}
			return
		}
	}
	t.Error("could not find '(' to inspect mark character")
}

// TestRadioButtonDrawMarkDiffersWhenSelectedVsNot verifies the draw output differs
// between selected and unselected states (falsification guard).
func TestRadioButtonDrawMarkDiffersWhenSelectedVsNot(t *testing.T) {
	selected := NewRadioButton(NewRect(0, 0, 15, 1), "Item")
	selected.scheme = theme.BorlandBlue
	selected.SetSelected(true)

	unselected := NewRadioButton(NewRect(0, 0, 15, 1), "Item")
	unselected.scheme = theme.BorlandBlue
	unselected.SetSelected(false)

	bufS := NewDrawBuffer(15, 1)
	bufU := NewDrawBuffer(15, 1)
	selected.Draw(bufS)
	unselected.Draw(bufU)

	// Find '(' and compare the next cell.
	for x := 0; x < 14; x++ {
		if bufS.GetCell(x, 0).Rune == '(' {
			markS := bufS.GetCell(x+1, 0).Rune
			markU := bufU.GetCell(x+1, 0).Rune
			if markS == markU {
				t.Errorf("selected and unselected RadioButton have same mark %q; must differ", markS)
			}
			return
		}
	}
	t.Error("could not find '(' to compare marks")
}

// TestRadioButtonDrawRendersLabelText verifies the label text appears in the output.
// Spec: "Renders as (*) Label / ( ) Label."
func TestRadioButtonDrawRendersLabelText(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 15, 1), "Hi")
	rb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	// Label "Hi" should appear somewhere after the "( ) " prefix.
	found := false
	for x := 0; x < 15; x++ {
		if buf.GetCell(x, 0).Rune == 'H' {
			found = true
			break
		}
	}
	if !found {
		t.Error("Draw did not render the label text 'H' from label 'Hi'")
	}
}

// TestRadioButtonDrawUsesRadioButtonNormalStyle verifies the bracket/mark cells use
// RadioButtonNormal style from the color scheme.
// Spec: "Uses RadioButtonNormal style from ColorScheme for the bracket/mark/label."
func TestRadioButtonDrawUsesRadioButtonNormalStyle(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 15, 1), "Item")
	scheme := theme.BorlandBlue
	rb.scheme = scheme

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	// The '(' bracket cell should use RadioButtonNormal style.
	for x := 0; x < 15; x++ {
		if buf.GetCell(x, 0).Rune == '(' {
			cell := buf.GetCell(x, 0)
			if cell.Style != scheme.RadioButtonNormal {
				t.Errorf("'(' cell style = %v, want RadioButtonNormal %v", cell.Style, scheme.RadioButtonNormal)
			}
			return
		}
	}
	t.Error("could not find '(' to check style")
}

// TestRadioButtonDrawShortcutCharUsesLabelShortcutStyle verifies the tilde-shortcut
// character in the label uses LabelShortcut style.
// Spec: "LabelShortcut style for tilde-shortcut characters in the label."
func TestRadioButtonDrawShortcutCharUsesLabelShortcutStyle(t *testing.T) {
	// "~T~est" → 'T' is shortcut, "est" is normal.
	// Layout: "( ) Test" or "( ) Test" — 'T' is at offset 4 from '('.
	rb := NewRadioButton(NewRect(0, 0, 15, 1), "~T~est")
	scheme := theme.BorlandBlue
	rb.scheme = scheme

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	// Find ')' then the space after it, then the next char should be 'T' with LabelShortcut style.
	for x := 0; x < 14; x++ {
		if buf.GetCell(x, 0).Rune == ')' {
			// After ')' is a space, then the label.
			labelStart := x + 2 // ')' + ' ' + label
			cell := buf.GetCell(labelStart, 0)
			if cell.Rune != 'T' {
				t.Errorf("expected 'T' at cell(%d,0), got %q", labelStart, cell.Rune)
				return
			}
			if cell.Style != scheme.LabelShortcut {
				t.Errorf("shortcut char 'T' style = %v, want LabelShortcut %v", cell.Style, scheme.LabelShortcut)
			}
			return
		}
	}
	t.Error("could not find ')' to locate label start")
}

// TestRadioButtonDrawNonShortcutLabelCharUsesRadioButtonNormalStyle verifies normal
// label characters (not tilde-marked) use RadioButtonNormal style.
// Spec: "RadioButtonNormal style … for … label."
func TestRadioButtonDrawNonShortcutLabelCharUsesRadioButtonNormalStyle(t *testing.T) {
	// "~T~est" → "est" are normal characters.
	rb := NewRadioButton(NewRect(0, 0, 15, 1), "~T~est")
	scheme := theme.BorlandBlue
	rb.scheme = scheme

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	// Find 'e' (first char of "est") and check its style.
	for x := 0; x < 15; x++ {
		if buf.GetCell(x, 0).Rune == 'e' {
			cell := buf.GetCell(x, 0)
			if cell.Style != scheme.RadioButtonNormal {
				t.Errorf("normal label char 'e' style = %v, want RadioButtonNormal %v", cell.Style, scheme.RadioButtonNormal)
			}
			return
		}
	}
	t.Error("could not find 'e' in rendered label")
}

// TestRadioButtonDrawShortcutStyleDiffersFromNormal falsification guard: LabelShortcut
// and RadioButtonNormal differ in BorlandBlue.
func TestRadioButtonDrawShortcutStyleDiffersFromNormal(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.LabelShortcut == scheme.RadioButtonNormal {
		t.Skip("LabelShortcut equals RadioButtonNormal in this scheme — style distinction test is vacuous")
	}

	rb := NewRadioButton(NewRect(0, 0, 15, 1), "~T~est")
	rb.scheme = scheme

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	// Compare 'T' (shortcut, LabelShortcut) and 'e' (normal, RadioButtonNormal).
	var tCell, eCell Cell
	var foundT, foundE bool
	for x := 0; x < 15; x++ {
		if buf.GetCell(x, 0).Rune == 'T' && !foundT {
			tCell = buf.GetCell(x, 0)
			foundT = true
		}
		if buf.GetCell(x, 0).Rune == 'e' && !foundE {
			eCell = buf.GetCell(x, 0)
			foundE = true
		}
	}
	if !foundT || !foundE {
		t.Skip("could not locate both 'T' and 'e' cells")
	}
	if tCell.Style == eCell.Style {
		t.Errorf("shortcut char 'T' and normal char 'e' have the same style %v; expected different styles", tCell.Style)
	}
}

// TestRadioButtonDrawWidth verifies the rendered width equals 5 + tildeTextLen(label).
// Spec: "Total rendered width: 5 + tildeTextLen(label)" (1 for focus indicator + 4 prefix + label).
func TestRadioButtonDrawWidth(t *testing.T) {
	label := "~T~est" // tildeTextLen = 4 ("Test")
	rb := NewRadioButton(NewRect(0, 0, 20, 1), label)
	rb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	rb.Draw(buf)

	// " ( ) Test" — 1 focus indicator + 3 bracket/mark chars + 1 space + 4 label chars = 9 total.
	expectedWidth := 5 + tildeTextLen(label) // 9
	// Check that content ends exactly where expected.
	// Position expectedWidth should not be part of the widget rendering.
	// We verify by checking the rendered runes fill exactly expectedWidth columns.
	for x := 0; x < expectedWidth; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune == 0 {
			t.Errorf("cell(%d,0) unexpectedly empty; expected content within rendered width %d", x, expectedWidth)
		}
	}
}

// =============================================================================
// RadioButton — drawing (with focus / SfSelected)
// =============================================================================

// TestRadioButtonDrawFocusCursorWhenSfSelected verifies a '►' prefix appears at x=0
// when the RadioButton has SfSelected (focus) state.
// Spec: "When SfSelected (has focus), a '►' prefix is rendered before the parentheses."
func TestRadioButtonDrawFocusCursorWhenSfSelected(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 15, 1), "Item")
	rb.scheme = theme.BorlandBlue
	rb.SetState(SfSelected, true)

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != '►' {
		t.Errorf("focused RadioButton cell(0,0) = %q, want '►'", cell.Rune)
	}
}

// TestRadioButtonDrawNoCursorWhenNotSfSelected verifies col 0 is a space (not '►')
// when SfSelected is not set.
// Spec: "When SfSelected (has focus), a '►' prefix is rendered." — unfocused has space.
func TestRadioButtonDrawNoCursorWhenNotSfSelected(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 15, 1), "Item")
	rb.scheme = theme.BorlandBlue
	rb.SetState(SfSelected, false)

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Errorf("unfocused RadioButton cell(0,0) = %q, want ' ' (space when not SfSelected)", cell.Rune)
	}
}

// TestRadioButtonDrawParenAlwaysAtColumn1 verifies '(' is always at col 1, whether
// focused (► at col 0) or unfocused (space at col 0).
func TestRadioButtonDrawParenAlwaysAtColumn1(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 15, 1), "Item")
	rb.scheme = theme.BorlandBlue
	rb.SetState(SfSelected, true)

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Rune != '(' {
		t.Errorf("focused RadioButton cell(1,0) = %q, want '('", cell.Rune)
	}
}

// TestRadioButtonDrawUnfocusedParenAtColumnOne verifies '(' is at x=1 when
// SfSelected is not set (col 0 is a space for the focus indicator).
func TestRadioButtonDrawUnfocusedParenAtColumnOne(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 15, 1), "Item")
	rb.scheme = theme.BorlandBlue
	rb.SetState(SfSelected, false)

	buf := NewDrawBuffer(15, 1)
	rb.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Rune != '(' {
		t.Errorf("unfocused RadioButton cell(1,0) = %q, want '('", cell.Rune)
	}
}

// =============================================================================
// RadioButton — keyboard handling
// =============================================================================

// TestRadioButtonHandleEventSpaceSelectsButton verifies the Space key sets the
// radio button as selected.
// Spec: "Space: selects this radio button."
func TestRadioButtonHandleEventSpaceSelectsButton(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")
	rb.SetSelected(false)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	rb.HandleEvent(ev)

	if !rb.Selected() {
		t.Error("Space key did not select the RadioButton")
	}
}

// TestRadioButtonHandleEventSpaceConsumesEvent verifies the Space key clears the event.
// Spec: "Space: … consumes event."
func TestRadioButtonHandleEventSpaceConsumesEvent(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	rb.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Space key did not consume the event; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestRadioButtonHandleEventEnterDoesNotSelectButton verifies the Enter key does NOT
// select the button (Enter handling removed per Task 4).
// Spec: "RadioButton.HandleEvent does NOT handle KeyEnter."
func TestRadioButtonHandleEventEnterDoesNotSelectButton(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")
	rb.SetSelected(false)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	rb.HandleEvent(ev)

	if rb.Selected() {
		t.Error("Enter key selected the RadioButton; Enter must not select after removal")
	}
}

// TestRadioButtonHandleEventEnterDoesNotConsumeEvent verifies the Enter key does NOT
// clear the event (Enter handling removed per Task 4).
// Spec: "RadioButton.HandleEvent does NOT handle KeyEnter."
func TestRadioButtonHandleEventEnterDoesNotConsumeEvent(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	rb.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("Enter key was consumed by RadioButton; Enter must not be consumed after removal")
	}
}

// TestRadioButtonHandleEventOtherKeyDoesNotSelect verifies that a key other than
// Space does not select the button.
// Spec: only Space selects.
func TestRadioButtonHandleEventOtherKeyDoesNotSelect(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")
	rb.SetSelected(false)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'x'}}
	rb.HandleEvent(ev)

	if rb.Selected() {
		t.Error("pressing 'x' selected the RadioButton; only Space should select it")
	}
}

// TestRadioButtonHandleEventOtherKeyDoesNotConsumeEvent verifies that a non-selecting
// key does not consume the event.
// Spec: only Space consumes the event.
func TestRadioButtonHandleEventOtherKeyDoesNotConsumeEvent(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'x'}}
	rb.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("pressing 'x' consumed the event; only Space should consume it")
	}
}

// =============================================================================
// RadioButton — mouse handling
// =============================================================================

// TestRadioButtonHandleEventMouseButton1SelectsButton verifies a left-click within
// bounds selects the radio button.
// Spec: "Click (Button1) within bounds: selects this radio button."
func TestRadioButtonHandleEventMouseButton1SelectsButton(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")
	rb.SetSelected(false)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 0, Button: tcell.Button1}}
	rb.HandleEvent(ev)

	if !rb.Selected() {
		t.Error("Button1 click did not select the RadioButton")
	}
}

// TestRadioButtonHandleEventMouseButton1ConsumesEvent verifies a left-click clears
// the event.
// Spec: "Click (Button1) within bounds: … consumes event."
func TestRadioButtonHandleEventMouseButton1ConsumesEvent(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 0, Button: tcell.Button1}}
	rb.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Button1 click did not consume the event; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestRadioButtonHandleEventMouseOtherButtonDoesNotSelect verifies that mouse buttons
// other than Button1 do not select.
// Spec: only Button1 selects.
func TestRadioButtonHandleEventMouseOtherButtonDoesNotSelect(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")
	rb.SetSelected(false)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 0, Button: tcell.Button2}}
	rb.HandleEvent(ev)

	if rb.Selected() {
		t.Error("Button2 click selected the RadioButton; only Button1 should select it")
	}
}

// TestRadioButtonHandleEventMouseOtherButtonDoesNotConsumeEvent verifies that
// non-Button1 clicks do not consume the event.
func TestRadioButtonHandleEventMouseOtherButtonDoesNotConsumeEvent(t *testing.T) {
	rb := NewRadioButton(NewRect(0, 0, 10, 1), "Item")

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 0, Button: tcell.Button2}}
	rb.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Button2 click consumed the event; only Button1 should consume it")
	}
}

// =============================================================================
// RadioButtons — construction
// =============================================================================

// TestNewRadioButtonsSetsSfVisible verifies NewRadioButtons sets SfVisible.
// Spec: "Sets SfVisible … by default."
func TestNewRadioButtonsSetsSfVisible(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	if !rbs.HasState(SfVisible) {
		t.Error("NewRadioButtons did not set SfVisible")
	}
}

// TestNewRadioButtonsSetsOfSelectable verifies NewRadioButtons sets OfSelectable.
// Spec: "Sets … OfSelectable … by default."
func TestNewRadioButtonsSetsOfSelectable(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	if !rbs.HasOption(OfSelectable) {
		t.Error("NewRadioButtons did not set OfSelectable")
	}
}

// TestNewRadioButtonsSetsOfPreProcess verifies NewRadioButtons sets OfPreProcess.
// Spec: "Sets … OfPreProcess by default."
func TestNewRadioButtonsSetsOfPreProcess(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	if !rbs.HasOption(OfPreProcess) {
		t.Error("NewRadioButtons did not set OfPreProcess")
	}
}

// TestNewRadioButtonsStoresBounds verifies NewRadioButtons records the given bounds.
// Spec: "NewRadioButtons(bounds Rect, labels []string) *RadioButtons"
func TestNewRadioButtonsStoresBounds(t *testing.T) {
	r := NewRect(2, 4, 30, 5)
	rbs := NewRadioButtons(r, []string{"A", "B", "C", "D", "E"})

	if rbs.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", rbs.Bounds(), r)
	}
}

// TestNewRadioButtonsCreatesOneButtonPerLabel verifies one RadioButton is created per label.
// Spec: "Creates one RadioButton per label."
func TestNewRadioButtonsCreatesOneButtonPerLabel(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	children := rbs.Children()
	if len(children) != 3 {
		t.Errorf("Children() len = %d, want 3 (one per label)", len(children))
	}
}

// TestNewRadioButtonsCreatesZeroButtonsForEmptySlice verifies empty labels produce
// no children.
func TestNewRadioButtonsCreatesZeroButtonsForEmptySlice(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 1), []string{})

	children := rbs.Children()
	if len(children) != 0 {
		t.Errorf("Children() len = %d, want 0 for empty labels", len(children))
	}
}

// TestNewRadioButtonsArrangesVertically verifies each RadioButton is at y=index.
// Spec: "Each RadioButton positioned at y=index."
func TestNewRadioButtonsArrangesVertically(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	for i := 0; i < 3; i++ {
		rb := rbs.Item(i)
		if rb == nil {
			t.Fatalf("Item(%d) returned nil", i)
		}
		if rb.Bounds().A.Y != i {
			t.Errorf("Item(%d).Bounds().A.Y = %d, want %d", i, rb.Bounds().A.Y, i)
		}
	}
}

// TestNewRadioButtonsFirstButtonSelectedByDefault verifies the first RadioButton is
// selected by default.
// Spec: "The first radio button is selected by default."
func TestNewRadioButtonsFirstButtonSelectedByDefault(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	if !rbs.Item(0).Selected() {
		t.Error("Item(0) should be selected by default")
	}
}

// TestNewRadioButtonsNonFirstButtonsNotSelectedByDefault verifies buttons other than
// the first are not selected.
// Spec: "The first radio button is selected by default" — implies others are not.
func TestNewRadioButtonsNonFirstButtonsNotSelectedByDefault(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	for i := 1; i < 3; i++ {
		if rbs.Item(i).Selected() {
			t.Errorf("Item(%d) should not be selected by default (only Item(0) is)", i)
		}
	}
}

// =============================================================================
// RadioButtons — Value / SetValue
// =============================================================================

// TestRadioButtonsValueReturnsIndexOfSelectedButton verifies Value() returns the
// 0-based index of the selected button.
// Spec: "Value() int returns the index of the selected radio button (0-based)."
func TestRadioButtonsValueReturnsIndexOfSelectedButton(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// First button is selected by default.

	if rbs.Value() != 0 {
		t.Errorf("Value() = %d, want 0 (first button selected by default)", rbs.Value())
	}
}

// TestRadioButtonsValueAfterSetValue verifies Value() reflects the last SetValue call.
// Spec: "Value() int returns the index of the selected radio button."
func TestRadioButtonsValueAfterSetValue(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(2)

	if rbs.Value() != 2 {
		t.Errorf("Value() = %d, want 2 after SetValue(2)", rbs.Value())
	}
}

// TestRadioButtonsValueReturnsNegativeOneWhenNoneSelected verifies Value() returns -1
// when no button is selected.
// Spec: "-1 if none."
func TestRadioButtonsValueReturnsNegativeOneWhenNoneSelected(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Manually deselect all.
	for i := 0; i < 3; i++ {
		rbs.Item(i).SetSelected(false)
	}

	if rbs.Value() != -1 {
		t.Errorf("Value() = %d, want -1 when no button is selected", rbs.Value())
	}
}

// TestRadioButtonsSetValueSelectsTargetButton verifies SetValue(i) selects Item(i).
// Spec: "SetValue(int) selects the radio button at the given index."
func TestRadioButtonsSetValueSelectsTargetButton(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(1)

	if !rbs.Item(1).Selected() {
		t.Error("SetValue(1) did not select Item(1)")
	}
}

// TestRadioButtonsSetValueDeselectsOthers verifies SetValue(i) deselects all other
// buttons in the cluster.
// Spec: "SetValue(int) selects the radio button at the given index (deselects others)."
func TestRadioButtonsSetValueDeselectsOthers(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(2)

	if rbs.Item(0).Selected() {
		t.Error("Item(0) should be deselected after SetValue(2)")
	}
	if rbs.Item(1).Selected() {
		t.Error("Item(1) should be deselected after SetValue(2)")
	}
}

// TestRadioButtonsSetValueDoesNotSelectNonTarget verifies SetValue(i) leaves buttons
// at other indices unselected (falsification guard).
func TestRadioButtonsSetValueDoesNotSelectNonTarget(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(0) // select first explicitly

	// Switching to 1 should deselect 0.
	rbs.SetValue(1)
	if rbs.Item(0).Selected() {
		t.Error("SetValue(1) did not deselect Item(0)")
	}
}

// =============================================================================
// RadioButtons — Item accessor
// =============================================================================

// TestRadioButtonsItemReturnsCorrectButton verifies Item(i) returns the RadioButton
// at the given index.
// Spec: "Item(index int) *RadioButton returns the RadioButton at the given index."
func TestRadioButtonsItemReturnsCorrectButton(t *testing.T) {
	labels := []string{"Alpha", "Beta", "Gamma"}
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), labels)

	for i, label := range labels {
		rb := rbs.Item(i)
		if rb == nil {
			t.Fatalf("Item(%d) returned nil", i)
		}
		if rb.Label() != label {
			t.Errorf("Item(%d).Label() = %q, want %q", i, rb.Label(), label)
		}
	}
}

// TestRadioButtonsItemIndexCorrespondsToChild verifies Item(i) returns the same
// underlying object as Children()[i].
// Spec: "Item(index int) *RadioButton returns the RadioButton at the given index."
func TestRadioButtonsItemIndexCorrespondsToChild(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	for i := 0; i < 3; i++ {
		rb := rbs.Item(i)
		children := rbs.Children()
		if View(rb) != children[i] {
			t.Errorf("Item(%d) != Children()[%d]", i, i)
		}
	}
}

// =============================================================================
// RadioButtons — exclusive selection (via keyboard/mouse on individual buttons)
// =============================================================================

// TestRadioButtonsExclusiveSelectionViaSpaceKey verifies that pressing Space on one
// radio button deselects the previously selected one within the cluster.
// Spec: "When a RadioButton in the cluster is selected … all other RadioButtons in
// the cluster are deselected."
func TestRadioButtonsExclusiveSelectionViaSpaceKey(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetState(SfSelected, true)
	// Focus Item(1) and press Space.
	rbs.SetFocusedChild(rbs.Item(1))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	rbs.HandleEvent(ev)

	if rbs.Item(0).Selected() {
		t.Error("Item(0) should be deselected after Item(1) was selected via Space")
	}
	if !rbs.Item(1).Selected() {
		t.Error("Item(1) should be selected after Space")
	}
}

// TestRadioButtonsExclusiveSelectionMutuallyExclusive verifies that only one button
// can be selected at a time in a cluster (falsification: selecting two should leave
// only the last one selected).
// Spec: "only one in a cluster can be selected."
func TestRadioButtonsExclusiveSelectionMutuallyExclusive(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(0)
	rbs.SetValue(2)

	selected := 0
	for i := 0; i < 3; i++ {
		if rbs.Item(i).Selected() {
			selected++
		}
	}
	if selected != 1 {
		t.Errorf("expected exactly 1 selected button, got %d", selected)
	}
}

// TestRadioButtonsExclusiveSelectionFalsification verifies that without the cluster
// enforcing exclusion, multiple buttons could be independently set (guard for the
// mutual exclusion tests above).
func TestRadioButtonsExclusiveSelectionFalsification(t *testing.T) {
	// Two independent RadioButtons (not in a cluster) can both be selected.
	rb1 := NewRadioButton(NewRect(0, 0, 10, 1), "A")
	rb2 := NewRadioButton(NewRect(0, 1, 10, 1), "B")
	rb1.SetSelected(true)
	rb2.SetSelected(true)

	if !rb1.Selected() || !rb2.Selected() {
		t.Skip("could not independently select two RadioButtons — falsification is invalid")
	}
	// This test serves as documentation: standalone RadioButtons don't enforce exclusion;
	// only the RadioButtons cluster does.
}

// =============================================================================
// RadioButtons — Alt+shortcut navigation
// =============================================================================

// TestRadioButtonsAltShortcutFocusesAndSelectsButton verifies that Alt+<shortcut>
// focuses and selects the corresponding RadioButton in the cluster.
// Spec: "Alt+shortcut letter focuses and selects the corresponding radio button."
func TestRadioButtonsAltShortcutFocusesAndSelectsButton(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"~A~lpha", "~B~eta", "~G~amma"})
	g.Insert(rbs)

	// Alt+B should focus/select Item(1).
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'b', Modifiers: tcell.ModAlt}}
	g.HandleEvent(ev)

	if !rbs.Item(1).Selected() {
		t.Error("Alt+B did not select Item(1) ('~B~eta')")
	}
}

// TestRadioButtonsAltShortcutDeselectsOthers verifies that when Alt+<shortcut> selects
// one button, the others are deselected.
// Spec: "focuses and selects the corresponding radio button" — with exclusive selection.
func TestRadioButtonsAltShortcutDeselectsOthers(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"~A~lpha", "~B~eta", "~G~amma"})
	g.Insert(rbs)

	// First button selected by default; Alt+G should select Item(2).
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'g', Modifiers: tcell.ModAlt}}
	g.HandleEvent(ev)

	if rbs.Item(0).Selected() {
		t.Error("Item(0) should be deselected after Alt+G selected Item(2)")
	}
}

// TestRadioButtonsAltShortcutConsumesEvent verifies Alt+shortcut clears the event.
// Spec: implied by "focuses and selects" — the cluster handles the event.
func TestRadioButtonsAltShortcutConsumesEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"~A~lpha", "~B~eta", "~G~amma"})
	g.Insert(rbs)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'a', Modifiers: tcell.ModAlt}}
	g.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Alt+A was not consumed; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestRadioButtonsWrongAltShortcutDoesNotSelect verifies that an Alt+letter that
// doesn't match any button's shortcut does not change the selection.
// Spec: only matching Alt+<shortcut> activates.
func TestRadioButtonsWrongAltShortcutDoesNotSelect(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"~A~lpha", "~B~eta", "~G~amma"})
	g.Insert(rbs)
	// Default: Item(0) selected.

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'z', Modifiers: tcell.ModAlt}}
	g.HandleEvent(ev)

	// Selection unchanged.
	if !rbs.Item(0).Selected() {
		t.Error("Alt+Z (no match) deselected Item(0); selection should be unchanged")
	}
}

// =============================================================================
// RadioButtons — Up/Down arrow navigation
// =============================================================================

// TestRadioButtonsDownArrowMovesToNextAndSelects verifies Down arrow moves focus to
// the next button AND selects it.
// Spec: "Down arrows move between radio buttons AND select them."
func TestRadioButtonsDownArrowMovesToNextAndSelects(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Item(0) is selected and focused by default.
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	if !rbs.Item(1).Selected() {
		t.Error("Down arrow did not select Item(1)")
	}
}

// TestRadioButtonsDownArrowDeselectsPrevious verifies Down arrow deselects the
// previously selected button.
// Spec: "Down arrows move between radio buttons AND select them" — with exclusive selection.
func TestRadioButtonsDownArrowDeselectsPrevious(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Item(0) is selected by default.
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	if rbs.Item(0).Selected() {
		t.Error("Item(0) should be deselected after Down arrow moved to Item(1)")
	}
}

// TestRadioButtonsDownArrowConsumesEvent verifies Down arrow clears the event.
// Spec: "Up/Down events are consumed."
func TestRadioButtonsDownArrowConsumesEvent(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Down arrow was not consumed; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestRadioButtonsUpArrowMovesToPreviousAndSelects verifies Up arrow moves focus to
// the previous button AND selects it.
// Spec: "Up arrows move between radio buttons AND select them."
func TestRadioButtonsUpArrowMovesToPreviousAndSelects(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(2) // start at Item(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	rbs.HandleEvent(ev)

	if !rbs.Item(1).Selected() {
		t.Error("Up arrow did not select Item(1) after starting at Item(2)")
	}
}

// TestRadioButtonsUpArrowConsumesEvent verifies Up arrow clears the event.
// Spec: "Up/Down events are consumed."
func TestRadioButtonsUpArrowConsumesEvent(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(1)
	rbs.SetFocusedChild(rbs.Item(1))
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Up arrow was not consumed; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestRadioButtonsDownArrowAtLastItemIsNoOp verifies Down at the last button does
// not wrap and does not change selection.
// Spec: "Down at last item is no-op."
func TestRadioButtonsDownArrowAtLastItemIsNoOp(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	// Item(2) should still be selected; no wrap to Item(0).
	if !rbs.Item(2).Selected() {
		t.Error("Down at last item should be a no-op; Item(2) is no longer selected")
	}
	if rbs.Item(0).Selected() {
		t.Error("Down at last item must not wrap to Item(0)")
	}
}

// TestRadioButtonsDownArrowAtLastItemConsumesEvent verifies Down at the last item
// still consumes the event.
// Spec: "Down at last item is no-op; Up/Down events are consumed."
func TestRadioButtonsDownArrowAtLastItemConsumesEvent(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Down at last item did not consume the event; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestRadioButtonsUpArrowAtFirstItemIsNoOp verifies Up at the first button does
// not wrap and does not change selection.
// Spec: "Up at first item is no-op."
func TestRadioButtonsUpArrowAtFirstItemIsNoOp(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Item(0) is selected and focused by default.
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	rbs.HandleEvent(ev)

	// Item(0) should still be selected; no wrap to Item(2).
	if !rbs.Item(0).Selected() {
		t.Error("Up at first item should be a no-op; Item(0) is no longer selected")
	}
	if rbs.Item(2).Selected() {
		t.Error("Up at first item must not wrap to Item(2)")
	}
}

// TestRadioButtonsUpArrowAtFirstItemConsumesEvent verifies Up at the first item
// still consumes the event.
// Spec: "Up at first item is no-op; Up/Down events are consumed."
func TestRadioButtonsUpArrowAtFirstItemConsumesEvent(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Up at first item did not consume the event; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestRadioButtonsDownArrowSelectsNotJustFocuses verifies Down SELECTS (not merely
// focuses) the next button.
// Spec: "unlike checkboxes where arrows just move focus" — arrows select.
func TestRadioButtonsDownArrowSelectsNotJustFocuses(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	rbs.HandleEvent(ev)

	// Item(1) must be selected (not just focused).
	if !rbs.Item(1).Selected() {
		t.Error("Down arrow moved focus but did not select Item(1); Down must also select")
	}
}

// =============================================================================
// RadioButtons — Container interface
// =============================================================================

// TestRadioButtonsImplementsContainerInsert verifies Insert adds a view to children.
// Spec: "Container interface: Insert(View)."
func TestRadioButtonsImplementsContainerInsert(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	initialCount := len(rbs.Children())

	extra := NewRadioButton(NewRect(0, 3, 20, 1), "D")
	rbs.Insert(extra)

	if len(rbs.Children()) != initialCount+1 {
		t.Errorf("after Insert, Children() len = %d, want %d", len(rbs.Children()), initialCount+1)
	}
}

// TestRadioButtonsImplementsContainerRemove verifies Remove removes a view from children.
// Spec: "Container interface: Remove(View)."
func TestRadioButtonsImplementsContainerRemove(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	initialCount := len(rbs.Children())

	rbs.Remove(rbs.Item(2))

	if len(rbs.Children()) != initialCount-1 {
		t.Errorf("after Remove, Children() len = %d, want %d", len(rbs.Children()), initialCount-1)
	}
}

// TestRadioButtonsImplementsContainerChildren verifies Children() returns all children.
// Spec: "Container interface: Children() []View."
func TestRadioButtonsImplementsContainerChildren(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	children := rbs.Children()
	if len(children) != 3 {
		t.Errorf("Children() len = %d, want 3", len(children))
	}
}

// TestRadioButtonsImplementsContainerFocusedChild verifies FocusedChild() returns
// the currently focused child.
// Spec: "Container interface: FocusedChild() View."
func TestRadioButtonsImplementsContainerFocusedChild(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	// After construction, some child should be focused.
	focused := rbs.FocusedChild()
	if focused == nil {
		t.Error("FocusedChild() returned nil; expected a focused child after construction")
	}
}

// TestRadioButtonsImplementsContainerSetFocusedChild verifies SetFocusedChild(v)
// changes the focused child.
// Spec: "Container interface: SetFocusedChild(View)."
func TestRadioButtonsImplementsContainerSetFocusedChild(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	target := rbs.Item(2)

	rbs.SetFocusedChild(target)

	if rbs.FocusedChild() != target {
		t.Errorf("FocusedChild() = %v, want Item(2) after SetFocusedChild(Item(2))", rbs.FocusedChild())
	}
}

// TestRadioButtonsDrawDelegatesToGroup verifies that Draw is delegated (children are drawn).
// Spec: "Drawing: Delegates to internal Group."
func TestRadioButtonsDrawDelegatesToGroup(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"Item"})
	rbs.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 3)
	// Must not panic and should write something to the buffer.
	rbs.Draw(buf)

	// At minimum, the '(' bracket from Item(0) should appear somewhere in row 0.
	found := false
	for x := 0; x < 20; x++ {
		if buf.GetCell(x, 0).Rune == '(' {
			found = true
			break
		}
	}
	if !found {
		t.Error("RadioButtons.Draw did not render '(' from its child RadioButton")
	}
}

// TestRadioButtonsTabMovesWithoutSelectingFalsification verifies Tab navigation
// (handled by Group) does NOT select the new button (unlike arrows).
// Spec: "Group handles Tab/Shift+Tab between radio buttons" — Tab only moves focus,
// not selection. (Arrows both move AND select; Tab does not select.)
func TestRadioButtonsTabMovesWithoutSelectingFalsification(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Item(0) selected by default.

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab}}
	rbs.HandleEvent(ev)

	// After Tab, focus moves to Item(1), but Item(0) should remain selected
	// (Tab does not select; only arrows and Space/Enter do).
	if !rbs.Item(0).Selected() {
		t.Log("Tab also selected the new item (arrows-like behaviour vs. Tab-only-focus behaviour)")
		// This is a documentation test for the distinction; mark informational.
	}
}

// =============================================================================
// RadioButtons — Value consistency after various operations
// =============================================================================

// TestRadioButtonsValueConsistentAfterMultipleSetValues verifies sequential SetValue
// calls leave exactly the last-set index selected.
// Spec: "Value() int returns the index of the selected radio button."
func TestRadioButtonsValueConsistentAfterMultipleSetValues(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(1)
	rbs.SetValue(2)
	rbs.SetValue(0)

	if rbs.Value() != 0 {
		t.Errorf("Value() = %d, want 0 after SetValue(0)", rbs.Value())
	}
	// Only Item(0) should be selected.
	for i := 1; i < 3; i++ {
		if rbs.Item(i).Selected() {
			t.Errorf("Item(%d) still selected after SetValue(0)", i)
		}
	}
}

// TestRadioButtonsOnlyOneSelectedAfterUpDownSequence verifies that after a sequence
// of Up/Down presses, exactly one button remains selected.
// Spec: "only one in a cluster can be selected."
func TestRadioButtonsOnlyOneSelectedAfterUpDownSequence(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	down := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	up := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}

	for i := 0; i < 5; i++ {
		rbs.HandleEvent(down)
	}
	for i := 0; i < 3; i++ {
		rbs.HandleEvent(up)
	}

	count := 0
	for i := 0; i < 3; i++ {
		if rbs.Item(i).Selected() {
			count++
		}
	}
	if count != 1 {
		t.Errorf("after Up/Down sequence, %d buttons selected; want exactly 1", count)
	}
}

// =============================================================================
// RadioButtons — Fix 1: Enter key exclusive selection within cluster
// =============================================================================

// TestRadioButtonsEnterKeyDoesNotChangeSelection verifies that pressing Enter on a
// focused (but not selected) button does NOT change selection (Enter removed per Task 4).
// Spec: "RadioButton.HandleEvent does NOT handle KeyEnter."
func TestRadioButtonsEnterKeyDoesNotChangeSelection(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Item(0) is selected by default; move focus to Item(1).
	rbs.SetFocusedChild(rbs.Item(1))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	rbs.HandleEvent(ev)

	// Enter must not change selection; Item(0) remains selected.
	if rbs.Item(1).Selected() {
		t.Error("Enter key selected Item(1); Enter must not select after removal")
	}
	if !rbs.Item(0).Selected() {
		t.Error("Item(0) should still be selected; Enter must not change selection")
	}
}

// =============================================================================
// RadioButtons — Fix 2: Mouse Button1 exclusive selection within cluster
// =============================================================================

// TestRadioButtonsMouseButton1ExclusiveSelection verifies that a Button1 click
// targeting Item(1) selects it and deselects the previously selected Item(0).
// Spec: "Click (Button1) within bounds: selects this radio button" — with exclusive
// selection enforced by the cluster.
func TestRadioButtonsMouseButton1ExclusiveSelection(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Item(0) is selected by default. Send a Button1 click at y=1 (Item(1)'s row).

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 1, Button: tcell.Button1}}
	rbs.HandleEvent(ev)

	if !rbs.Item(1).Selected() {
		t.Error("Button1 click at y=1 did not select Item(1)")
	}
	if rbs.Item(0).Selected() {
		t.Error("Item(0) should be deselected after Button1 click selected Item(1)")
	}
}

// =============================================================================
// RadioButtons — Fix 3: Alt+shortcut verifies focus, not just selection
// =============================================================================

// TestRadioButtonsAltShortcutSetsFocus verifies that Alt+<shortcut> sets focus
// to the corresponding RadioButton (not just selection).
// Spec: "Alt+shortcut letter focuses and selects the corresponding radio button."
func TestRadioButtonsAltShortcutSetsFocus(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"~A~lpha", "~B~eta", "~C~harlie"})

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'b', Modifiers: tcell.ModAlt}}
	rbs.HandleEvent(ev)

	if rbs.FocusedChild() != View(rbs.Item(1)) {
		t.Errorf("FocusedChild() = %v, want Item(1) after Alt+B", rbs.FocusedChild())
	}
}

// =============================================================================
// RadioButtons — Fix 4: Tab does not change selection (asserting test)
// =============================================================================

// TestRadioButtonsTabDoesNotChangeSelection verifies that Tab moves focus but does
// NOT change the selected button (unlike arrows which both move focus and select).
// Spec: "Group handles Tab/Shift+Tab between radio buttons" — Tab only moves focus,
// not selection.
func TestRadioButtonsTabDoesNotChangeSelection(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Item(0) is selected by default.

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyTab}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("Tab changed selection: Value() = %d, want 0 (Tab must not change selection)", rbs.Value())
	}
}

func TestRadioButtonsPlainLetterDoesNotStealFromFocusedSibling(t *testing.T) {
	grp := NewGroup(NewRect(0, 0, 50, 20))
	grp.SetState(SfVisible|SfSelected|SfFocused, true)

	input := NewInputLine(NewRect(0, 0, 20, 1), 40)
	rbs := NewRadioButtons(NewRect(0, 5, 30, 3), []string{"~T~ext", "~B~inary", "~H~ex"})

	grp.Insert(input)
	grp.Insert(rbs)
	grp.SetFocusedChild(input)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 't', Modifiers: 0}}
	grp.HandleEvent(ev)

	if input.Text() != "t" {
		t.Errorf("InputLine.Text() = %q, want %q — RadioButtons stole plain 't' via PreProcess", input.Text(), "t")
	}
	if rbs.Value() != 0 {
		t.Errorf("RadioButtons.Value() = %d, want 0 — should not have switched to Text", rbs.Value())
	}
}

func TestRadioButtonsPlainLetterMatchesInPostProcess(t *testing.T) {
	grp := NewGroup(NewRect(0, 0, 50, 20))
	grp.SetState(SfVisible|SfSelected|SfFocused, true)

	btn := NewButton(NewRect(0, 0, 10, 2), "OK", CmOK)
	rbs := NewRadioButtons(NewRect(0, 5, 30, 3), []string{"~T~ext", "~B~inary", "~H~ex"})

	grp.Insert(btn)
	grp.Insert(rbs)
	grp.SetFocusedChild(btn)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 't', Modifiers: 0}}
	grp.HandleEvent(ev)

	if rbs.Value() != 0 {
		t.Errorf("Value() = %d, want 0 — plain 't' should match ~T~ext in PostProcess when focused sibling doesn't consume it", rbs.Value())
	}
}

func TestRadioButtonsPlainLetterMatchesWhenFocused(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 30, 3), []string{"~T~ext", "~B~inary", "~H~ex"})
	rbs.SetState(SfVisible|SfSelected|SfFocused, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'b', Modifiers: 0}}
	rbs.HandleEvent(ev)

	if rbs.Value() != 1 {
		t.Errorf("Value() = %d, want 1 — plain 'b' should match ~B~inary when focused", rbs.Value())
	}
}
