package tv

// Tests for menu_popup.go — Task 2: MenuPopup Rendering and Navigation
// Each test verifies a specific spec requirement.
// Falsifying tests are included to ensure implementations can't trivially pass.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testCS returns a distinguishable color scheme for Draw tests.
func testCS() *theme.ColorScheme {
	return &theme.ColorScheme{
		MenuNormal:   tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorGreen),
		MenuShortcut: tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorGreen),
		MenuSelected: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue),
		MenuDisabled: tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorGreen),
	}
}

// cellAt reads a cell from a draw buffer using the spec-described GetCell method.
func cellAt(buf *DrawBuffer, x, y int) Cell {
	return buf.GetCell(x, y)
}

// ---------------------------------------------------------------------------
// 1. Constructor — NewMenuPopup
// ---------------------------------------------------------------------------

// TestNewMenuPopupReturnsNonNil verifies that NewMenuPopup returns a non-nil pointer.
// Spec: "NewMenuPopup(items []any, x, y int) *MenuPopup creates a popup."
func TestNewMenuPopupReturnsNonNil(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	if p == nil {
		t.Fatal("NewMenuPopup returned nil, want *MenuPopup")
	}
}

// TestNewMenuPopupBoundsXY verifies that the popup bounds are anchored at the given (x, y).
// Spec: "positioned at screen coordinates (x, y)."
func TestNewMenuPopupBoundsXY(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 10, 5)
	b := p.Bounds()
	if b.A.X != 10 || b.A.Y != 5 {
		t.Errorf("Bounds top-left = (%d, %d), want (10, 5)", b.A.X, b.A.Y)
	}
}

// TestNewMenuPopupBoundsXYVaries verifies that different (x, y) produce different bounds.
// Falsifying: two popups at different positions must have different bounds.
func TestNewMenuPopupBoundsXYVaries(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p1 := NewMenuPopup(items, 0, 0)
	p2 := NewMenuPopup(items, 20, 10)
	if p1.Bounds() == p2.Bounds() {
		t.Error("Popups at different positions must have different bounds")
	}
}

// TestNewMenuPopupBoundsWidthFromPopupWidth verifies popup width = popupWidth(items).
// Spec: "Width and height are auto-computed from items using popupWidth and popupHeight."
func TestNewMenuPopupBoundsWidthFromPopupWidth(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())} // "New" = 3, popupWidth = 7
	p := NewMenuPopup(items, 0, 0)
	wantW := popupWidth(items)
	gotW := p.Bounds().Width()
	if gotW != wantW {
		t.Errorf("Bounds().Width() = %d, want %d (from popupWidth)", gotW, wantW)
	}
}

// TestNewMenuPopupBoundsHeightFromPopupHeight verifies popup height = popupHeight(items).
// Spec: "Width and height are auto-computed from items using popupWidth and popupHeight."
func TestNewMenuPopupBoundsHeightFromPopupHeight(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	} // 2 items → popupHeight = 4
	p := NewMenuPopup(items, 0, 0)
	wantH := popupHeight(items)
	gotH := p.Bounds().Height()
	if gotH != wantH {
		t.Errorf("Bounds().Height() = %d, want %d (from popupHeight)", gotH, wantH)
	}
}

// TestNewMenuPopupInitialResultZero verifies Result() == 0 before any selection.
// Spec: "result CommandCode — command of selected item, 0 if not yet selected."
func TestNewMenuPopupInitialResultZero(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	if p.Result() != 0 {
		t.Errorf("Result() = %v, want 0 before any selection", p.Result())
	}
}

// TestNewMenuPopupInitialSelectedIsZero verifies Selected() == 0 when first item is not a separator.
// Spec: "The initial selected is 0 (first item). If the first item is a separator, advance."
func TestNewMenuPopupInitialSelectedIsZero(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	if p.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0 (first item is not a separator)", p.Selected())
	}
}

// TestNewMenuPopupInitialSelectedSkipsSeparator verifies that if the first item is a separator,
// Selected() advances to the first non-separator.
// Spec: "If the first item is a separator, advance to the next non-separator."
func TestNewMenuPopupInitialSelectedSkipsSeparator(t *testing.T) {
	items := []any{
		NewMenuSeparator(),
		NewMenuItem("~N~ew", CmUser, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	if p.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1 (first item is separator, should advance)", p.Selected())
	}
}

// TestNewMenuPopupInitialSelectedNotNegativeOne verifies Selected() is never -1 initially
// if items are present.
// Falsifying: initial selected must not be -1 when there is a selectable item.
func TestNewMenuPopupInitialSelectedNotNegativeOne(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	if p.Selected() == -1 {
		t.Error("Selected() must not be -1 initially when items exist")
	}
}

// ---------------------------------------------------------------------------
// 2. Bounds() and Result() and Selected() accessors
// ---------------------------------------------------------------------------

// TestBoundsReturnsScreenCoordinates verifies Bounds() returns screen-coordinate Rect.
// Spec: "MenuPopup.Bounds() Rect returns the popup's bounds in screen coordinates."
func TestBoundsReturnsScreenCoordinates(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 15, 8)
	b := p.Bounds()
	if b.A.X != 15 || b.A.Y != 8 {
		t.Errorf("Bounds().A = (%d, %d), want (15, 8)", b.A.X, b.A.Y)
	}
}

// TestResultReturnsZeroInitially verifies Result() returns 0 before any command is fired.
// Spec: "MenuPopup.Result() CommandCode returns the selected command (0 if none)."
func TestResultReturnsZeroInitially(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	if p.Result() != 0 {
		t.Errorf("Result() = %v, want 0 initially", p.Result())
	}
}

// TestSelectedReturnsCurrentHighlightedIndex verifies Selected() reflects current highlight.
// Spec: "MenuPopup.Selected() int returns the selected index."
func TestSelectedReturnsCurrentHighlightedIndex(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	// Press Down to move to second item
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if p.Selected() != 1 {
		t.Errorf("Selected() = %d after Down, want 1", p.Selected())
	}
}

// ---------------------------------------------------------------------------
// 3. Navigation — Down arrow
// ---------------------------------------------------------------------------

// TestDownArrowMovesToNextItem verifies Down moves Selected to the next item.
// Spec: "Down arrow: moves selected to next non-separator item, wrapping around."
func TestDownArrowMovesToNextItem(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if p.Selected() != 1 {
		t.Errorf("Selected() = %d after Down from 0, want 1", p.Selected())
	}
}

// TestDownArrowWrapsAround verifies Down wraps from last item back to first.
// Spec: "Down arrow: moves selected to next non-separator item, wrapping around."
func TestDownArrowWrapsAround(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	// Start at 0, move to 1
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	// Wrap: move from 1 back to 0
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if p.Selected() != 0 {
		t.Errorf("Selected() = %d after Down wrap, want 0", p.Selected())
	}
}

// TestDownArrowSkipsSeparators verifies Down skips separator items.
// Spec: "Down arrow: moves selected to next non-separator item, wrapping around."
func TestDownArrowSkipsSeparators(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),   // index 0
		NewMenuSeparator(),                        // index 1 — skip
		NewMenuItem("~O~pen", CmUser+1, KbNone()), // index 2
	}
	p := NewMenuPopup(items, 0, 0)
	// Down from 0 should skip sep at 1 and land on 2
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if p.Selected() != 2 {
		t.Errorf("Selected() = %d after Down (skip separator), want 2", p.Selected())
	}
}

// TestDownArrowDoesNotStopOnSeparator falsifies an impl that halts on separators.
func TestDownArrowDoesNotStopOnSeparator(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuSeparator(),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	if p.Selected() == 1 {
		t.Error("Selected() must not stop at a separator (index 1)")
	}
}

// ---------------------------------------------------------------------------
// 4. Navigation — Up arrow
// ---------------------------------------------------------------------------

// TestUpArrowMovesToPreviousItem verifies Up moves Selected to the previous item.
// Spec: "Up arrow: moves selected to previous non-separator item, wrapping around."
func TestUpArrowMovesToPreviousItem(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	// Move to item 1 first
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}})
	// Then Up should bring us back to 0
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}})
	if p.Selected() != 0 {
		t.Errorf("Selected() = %d after Up from 1, want 0", p.Selected())
	}
}

// TestUpArrowWrapsAround verifies Up wraps from first item to last.
// Spec: "Up arrow: moves selected to previous non-separator item, wrapping around."
func TestUpArrowWrapsAround(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	// Currently at 0; Up should wrap to 1
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}})
	if p.Selected() != 1 {
		t.Errorf("Selected() = %d after Up wrap from 0, want 1", p.Selected())
	}
}

// TestUpArrowSkipsSeparators verifies Up skips separator items.
// Spec: "Up arrow: moves selected to previous non-separator item, wrapping around."
func TestUpArrowSkipsSeparators(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),   // index 0
		NewMenuSeparator(),                        // index 1 — skip
		NewMenuItem("~O~pen", CmUser+1, KbNone()), // index 2
	}
	p := NewMenuPopup(items, 0, 0)
	// Start at 0; Up wraps, should skip sep at 1 and land on 2
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}})
	if p.Selected() != 2 {
		t.Errorf("Selected() = %d after Up wrap (skip separator), want 2", p.Selected())
	}
}

// ---------------------------------------------------------------------------
// 5. Selection — Enter
// ---------------------------------------------------------------------------

// TestEnterFiresCommandOfSelectedItem verifies Enter sets Result to selected item's Command.
// Spec: "Enter: if selected points to a non-disabled MenuItem, sets result to that item's Command."
func TestEnterFiresCommandOfSelectedItem(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}})
	if p.Result() != CmUser {
		t.Errorf("Result() = %v after Enter, want %v (CmUser)", p.Result(), CmUser)
	}
}

// TestEnterFiresCorrectCommandWhenMultipleItems verifies Enter fires the highlighted item's command.
// Falsifying: result must reflect the selected index, not always item 0.
func TestEnterFiresCorrectCommandWhenMultipleItems(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}) // select index 1
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}})
	if p.Result() != CmUser+1 {
		t.Errorf("Result() = %v after Enter on item 1, want %v (CmUser+1)", p.Result(), CmUser+1)
	}
}

// TestEnterDoesNotFireOnDisabledItem verifies Enter does nothing on a disabled MenuItem.
// Spec: "Enter: if selected points to a non-disabled MenuItem, sets result."
func TestEnterDoesNotFireOnDisabledItem(t *testing.T) {
	disabled := NewMenuItem("~N~ew", CmUser, KbNone())
	disabled.Disabled = true
	items := []any{disabled}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}})
	if p.Result() != 0 {
		t.Errorf("Result() = %v after Enter on disabled item, want 0", p.Result())
	}
}

// TestEnterOnDisabledDoesNotSetNonZeroResult falsifies implementations that fire disabled items.
func TestEnterOnDisabledDoesNotSetNonZeroResult(t *testing.T) {
	disabled := NewMenuItem("~Q~uit", CmQuit, KbNone())
	disabled.Disabled = true
	items := []any{disabled}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}})
	if p.Result() == CmQuit {
		t.Error("Enter on disabled item must not set Result to item's command")
	}
}

// ---------------------------------------------------------------------------
// 6. Selection — Escape
// ---------------------------------------------------------------------------

// TestEscapeSetsCmCancel verifies Escape sets Result to CmCancel.
// Spec: "Escape: sets result to CmCancel."
func TestEscapeSetsCmCancel(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEscape}})
	if p.Result() != CmCancel {
		t.Errorf("Result() = %v after Escape, want CmCancel (%v)", p.Result(), CmCancel)
	}
}

// TestEscapeResultIsNotZero verifies Escape result is distinguishable from "no selection".
// Falsifying: CmCancel must not be 0.
func TestEscapeResultIsNotZero(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEscape}})
	if p.Result() == 0 {
		t.Error("Result() after Escape must not be 0 (CmCancel is a distinct command)")
	}
}

// ---------------------------------------------------------------------------
// 7. Shortcut matching — rune keys
// ---------------------------------------------------------------------------

// TestRuneKeyMatchesTildeShortcutCaseInsensitiveLower verifies lowercase rune triggers shortcut.
// Spec: "Rune key: if a menu item's tilde shortcut matches the rune (case-insensitive), fires."
func TestRuneKeyMatchesTildeShortcutCaseInsensitiveLower(t *testing.T) {
	// "~N~ew" → shortcut 'N'; pressing 'n' (lowercase) should fire
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'n'}})
	if p.Result() != CmUser {
		t.Errorf("Result() = %v after rune 'n' (shortcut 'N'), want %v", p.Result(), CmUser)
	}
}

// TestRuneKeyMatchesTildeShortcutCaseInsensitiveUpper verifies uppercase rune triggers shortcut.
func TestRuneKeyMatchesTildeShortcutCaseInsensitiveUpper(t *testing.T) {
	// "~N~ew" → shortcut 'N'; pressing 'N' (uppercase) should fire
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'N'}})
	if p.Result() != CmUser {
		t.Errorf("Result() = %v after rune 'N' (shortcut 'N'), want %v", p.Result(), CmUser)
	}
}

// TestRuneKeySelectsCorrectItemAmongMultiple verifies rune fires the matching item's command.
// Falsifying: must fire item with matching shortcut, not first item.
func TestRuneKeySelectsCorrectItemAmongMultiple(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'o'}})
	if p.Result() != CmUser+1 {
		t.Errorf("Result() = %v after 'o', want %v (CmUser+1 for ~O~pen)", p.Result(), CmUser+1)
	}
}

// TestRuneKeyNonMatchingDoesNotFire verifies unmatched rune does not set Result.
// Falsifying: an unmatched rune must not fire any command.
func TestRuneKeyNonMatchingDoesNotFire(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'x'}})
	if p.Result() != 0 {
		t.Errorf("Result() = %v after unmatched rune 'x', want 0", p.Result())
	}
}

// TestRuneKeyAlsoSelectsItem verifies the rune press also updates Selected() to the matching index.
// Spec: "selects and fires that item."
func TestRuneKeyAlsoSelectsItem(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'o'}})
	if p.Selected() != 1 {
		t.Errorf("Selected() = %d after rune 'o', want 1", p.Selected())
	}
}

// ---------------------------------------------------------------------------
// 8. Mouse — click
// ---------------------------------------------------------------------------

// TestMouseClickOnItemRowSelectsAndFires verifies Button1 click on item row fires that item.
// Spec: "Mouse click (Button1): if click is on an item row (y=1 to y=height-2 in popup-local
//
//	coords), selects that item. If the item is a non-disabled MenuItem, fires it."
func TestMouseClickOnItemRowSelectsAndFires(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	// y=1 in popup-local coords corresponds to items[0]
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 1, Button: tcell.Button1}})
	if p.Result() != CmUser {
		t.Errorf("Result() = %v after click at y=1, want %v (item 0)", p.Result(), CmUser)
	}
}

// TestMouseClickOnSecondItemFires verifies clicking second item row fires second item.
func TestMouseClickOnSecondItemFires(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	// y=2 in popup-local coords corresponds to items[1]
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 2, Button: tcell.Button1}})
	if p.Result() != CmUser+1 {
		t.Errorf("Result() = %v after click at y=2, want %v (item 1)", p.Result(), CmUser+1)
	}
}

// TestMouseClickOnItemUpdatesSelected verifies the click also updates Selected().
// Spec: "selects that item."
func TestMouseClickOnItemUpdatesSelected(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 2, Button: tcell.Button1}})
	if p.Selected() != 1 {
		t.Errorf("Selected() = %d after click at y=2, want 1", p.Selected())
	}
}

// TestMouseClickOnDisabledItemDoesNotFire verifies click on disabled item does not set Result.
// Spec: "If the item is a non-disabled MenuItem, fires it."
func TestMouseClickOnDisabledItemDoesNotFire(t *testing.T) {
	disabled := NewMenuItem("~N~ew", CmUser, KbNone())
	disabled.Disabled = true
	items := []any{disabled}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 1, Button: tcell.Button1}})
	if p.Result() != 0 {
		t.Errorf("Result() = %v after click on disabled item, want 0", p.Result())
	}
}

// TestMouseClickOnTopBorderIgnored verifies click on top border (y=0) is ignored.
// Spec: "Clicks on borders or separators are ignored."
func TestMouseClickOnTopBorderIgnored(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 0, Button: tcell.Button1}})
	if p.Result() != 0 {
		t.Errorf("Result() = %v after click on top border (y=0), want 0", p.Result())
	}
}

// TestMouseClickOnBottomBorderIgnored verifies click on bottom border is ignored.
// Spec: "Clicks on borders or separators are ignored."
func TestMouseClickOnBottomBorderIgnored(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	h := p.Bounds().Height()
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: h - 1, Button: tcell.Button1}})
	if p.Result() != 0 {
		t.Errorf("Result() = %v after click on bottom border (y=%d), want 0", p.Result(), h-1)
	}
}

// TestMouseClickOnSeparatorIgnored verifies click on separator row does not fire.
// Spec: "Clicks on borders or separators are ignored."
func TestMouseClickOnSeparatorIgnored(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()), // y=1
		NewMenuSeparator(),                      // y=2
		NewMenuItem("~O~pen", CmUser+1, KbNone()), // y=3
	}
	p := NewMenuPopup(items, 0, 0)
	// Click on separator row (y=2 in popup-local)
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 2, Button: tcell.Button1}})
	if p.Result() != 0 {
		t.Errorf("Result() = %v after click on separator row, want 0", p.Result())
	}
}

// ---------------------------------------------------------------------------
// 9. Mouse — hover (move)
// ---------------------------------------------------------------------------

// TestMouseMoveHighlightsItemRow verifies ButtonNone mouse move updates Selected().
// Spec: "Mouse move (ButtonNone, no buttons pressed): if mouse is on an item row, highlight it."
func TestMouseMoveHighlightsItemRow(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	// Hover over item 1 (y=2)
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 2, Button: tcell.ButtonNone}})
	if p.Selected() != 1 {
		t.Errorf("Selected() = %d after hover at y=2, want 1", p.Selected())
	}
}

// TestMouseMoveDoesNotFireResult verifies hover does not set Result.
// Falsifying: move/hover must NOT fire the command.
func TestMouseMoveDoesNotFireResult(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 1, Button: tcell.ButtonNone}})
	if p.Result() != 0 {
		t.Errorf("Result() = %v after hover, want 0 (hover must not fire)", p.Result())
	}
}

// TestMouseMoveOverDifferentRowsUpdatesHighlight verifies hover tracks as mouse moves.
func TestMouseMoveOverDifferentRowsUpdatesHighlight(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	// Hover item 0 (y=1)
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 1, Button: tcell.ButtonNone}})
	if p.Selected() != 0 {
		t.Errorf("Selected() = %d after hover at y=1, want 0", p.Selected())
	}
	// Move to item 1 (y=2)
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 2, Button: tcell.ButtonNone}})
	if p.Selected() != 1 {
		t.Errorf("Selected() = %d after hover at y=2, want 1", p.Selected())
	}
}

// ---------------------------------------------------------------------------
// 10. SubMenu treated as disabled
// ---------------------------------------------------------------------------

// TestSubMenuTreatedAsDisabledEnterDoesNotFire verifies Enter on a SubMenu item does not fire.
// Spec: "Items that are *SubMenu are treated as disabled menu items."
func TestSubMenuTreatedAsDisabledEnterDoesNotFire(t *testing.T) {
	items := []any{NewSubMenu("~F~ile")}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}})
	if p.Result() != 0 {
		t.Errorf("Result() = %v after Enter on SubMenu item, want 0 (treated as disabled)", p.Result())
	}
}

// TestSubMenuTreatedAsDisabledClickDoesNotFire verifies click on a SubMenu row does not fire.
// Spec: "Items that are *SubMenu are treated as disabled menu items."
func TestSubMenuTreatedAsDisabledClickDoesNotFire(t *testing.T) {
	items := []any{NewSubMenu("~F~ile")}
	p := NewMenuPopup(items, 0, 0)
	p.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{X: 2, Y: 1, Button: tcell.Button1}})
	if p.Result() != 0 {
		t.Errorf("Result() = %v after click on SubMenu item, want 0 (treated as disabled)", p.Result())
	}
}

// ---------------------------------------------------------------------------
// 11. Rendering — border characters
// ---------------------------------------------------------------------------

// TestDrawBorderTopLeft verifies top-left corner is '┌' with MenuNormal style.
// Spec: "Single-line border using MenuNormal style: ┌─┐│└─┘."
func TestDrawBorderTopLeft(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	cell := cellAt(buf, 0, 0)
	if cell.Rune != '┌' {
		t.Errorf("Top-left corner: got %q, want '┌'", cell.Rune)
	}
	if cell.Style != cs.MenuNormal {
		t.Error("Top-left corner must use MenuNormal style")
	}
}

// TestDrawBorderTopRight verifies top-right corner is '┐' with MenuNormal style.
func TestDrawBorderTopRight(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	cell := cellAt(buf, w-1, 0)
	if cell.Rune != '┐' {
		t.Errorf("Top-right corner: got %q, want '┐'", cell.Rune)
	}
	if cell.Style != cs.MenuNormal {
		t.Error("Top-right corner must use MenuNormal style")
	}
}

// TestDrawBorderBottomLeft verifies bottom-left corner is '└' with MenuNormal style.
func TestDrawBorderBottomLeft(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	cell := cellAt(buf, 0, h-1)
	if cell.Rune != '└' {
		t.Errorf("Bottom-left corner: got %q, want '└'", cell.Rune)
	}
	if cell.Style != cs.MenuNormal {
		t.Error("Bottom-left corner must use MenuNormal style")
	}
}

// TestDrawBorderBottomRight verifies bottom-right corner is '┘' with MenuNormal style.
func TestDrawBorderBottomRight(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	cell := cellAt(buf, w-1, h-1)
	if cell.Rune != '┘' {
		t.Errorf("Bottom-right corner: got %q, want '┘'", cell.Rune)
	}
	if cell.Style != cs.MenuNormal {
		t.Error("Bottom-right corner must use MenuNormal style")
	}
}

// TestDrawBorderTopHorizontal verifies top border cells between corners are '─'.
func TestDrawBorderTopHorizontal(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// Cells at y=0 between corners (x=1 .. w-2) must be '─'
	for x := 1; x < w-1; x++ {
		cell := cellAt(buf, x, 0)
		if cell.Rune != '─' {
			t.Errorf("Top border at x=%d: got %q, want '─'", x, cell.Rune)
		}
	}
}

// TestDrawBorderBottomHorizontal verifies bottom border cells between corners are '─'.
func TestDrawBorderBottomHorizontal(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	for x := 1; x < w-1; x++ {
		cell := cellAt(buf, x, h-1)
		if cell.Rune != '─' {
			t.Errorf("Bottom border at x=%d: got %q, want '─'", x, cell.Rune)
		}
	}
}

// TestDrawBorderLeftVertical verifies left border cells between corners are '│'.
func TestDrawBorderLeftVertical(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	for y := 1; y < h-1; y++ {
		cell := cellAt(buf, 0, y)
		if cell.Rune != '│' {
			t.Errorf("Left border at y=%d: got %q, want '│'", y, cell.Rune)
		}
	}
}

// TestDrawBorderRightVertical verifies right border cells between corners are '│'.
func TestDrawBorderRightVertical(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	for y := 1; y < h-1; y++ {
		cell := cellAt(buf, w-1, y)
		if cell.Rune != '│' {
			t.Errorf("Right border at y=%d: got %q, want '│'", y, cell.Rune)
		}
	}
}

// ---------------------------------------------------------------------------
// 12. Rendering — separator characters
// ---------------------------------------------------------------------------

// TestDrawSeparatorUsesLTee verifies separator left join is '├'.
// Spec: "Each *MenuSeparator: horizontal line ├─┤ across the full popup width."
func TestDrawSeparatorUsesLTee(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuSeparator(),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// Separator is at item index 1 → row y=2 (border at y=0, item0 at y=1, sep at y=2)
	cell := cellAt(buf, 0, 2)
	if cell.Rune != '├' {
		t.Errorf("Separator left: got %q, want '├'", cell.Rune)
	}
}

// TestDrawSeparatorUsesRTee verifies separator right join is '┤'.
func TestDrawSeparatorUsesRTee(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuSeparator(),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	cell := cellAt(buf, w-1, 2)
	if cell.Rune != '┤' {
		t.Errorf("Separator right: got %q, want '┤'", cell.Rune)
	}
}

// TestDrawSeparatorFillsHorizontalLine verifies separator interior is '─' across full width.
func TestDrawSeparatorFillsHorizontalLine(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuSeparator(),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	for x := 1; x < w-1; x++ {
		cell := cellAt(buf, x, 2)
		if cell.Rune != '─' {
			t.Errorf("Separator interior at x=%d: got %q, want '─'", x, cell.Rune)
		}
	}
}

// TestDrawSeparatorUsesMenuNormalStyle verifies separator is drawn with MenuNormal style.
func TestDrawSeparatorUsesMenuNormalStyle(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuSeparator(),
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// Separator row y=2; check the left tee
	cell := cellAt(buf, 0, 2)
	if cell.Style != cs.MenuNormal {
		t.Error("Separator must be drawn with MenuNormal style")
	}
}

// ---------------------------------------------------------------------------
// 13. Rendering — item label positioning and shortcut style
// ---------------------------------------------------------------------------

// TestDrawItemLabelStartsAfterLeftBorderPad verifies label starts at x=1 (border) + 1 (pad) = x=1.
// Spec: "left-aligned label."
// Note: The spec says left-aligned within the border+padding area. The border is at x=0,
// so the label starts at x=1 (immediately after left border).
func TestDrawItemLabelAppearsInsideBorder(t *testing.T) {
	// "New" (no tilde shortcut at x=1): "~N~ew" → N is shortcut, "ew" is normal
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// Item row y=1; label starts after left border at x=1
	// First char is shortcut 'N'
	cell := cellAt(buf, 1, 1)
	if cell.Rune != 'N' {
		t.Errorf("Item label first char at (1, 1): got %q, want 'N'", cell.Rune)
	}
}

// TestDrawItemShortcutCharUsesMenuShortcutStyle verifies tilde shortcut char uses MenuShortcut style.
// Spec: "shortcut in MenuShortcut style."
func TestDrawItemShortcutCharUsesMenuShortcutStyle(t *testing.T) {
	// Use two items; selection starts at index 0, so index 1 is unselected.
	items := []any{
		NewMenuItem("~O~pen", CmUser, KbNone()),
		NewMenuItem("~N~ew", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// 'N' is the shortcut char at x=1, y=2 (unselected item at index 1)
	cell := cellAt(buf, 1, 2)
	if cell.Style != cs.MenuShortcut {
		t.Errorf("Shortcut char 'N' style = %v, want MenuShortcut (%v)", cell.Style, cs.MenuShortcut)
	}
}

// TestDrawItemNonShortcutCharUsesMenuNormalStyle verifies non-shortcut label chars use MenuNormal.
// Spec: "rest in MenuNormal style."
func TestDrawItemNonShortcutCharUsesMenuNormalStyle(t *testing.T) {
	// Use two items; selection starts at index 0, so index 1 is unselected.
	items := []any{
		NewMenuItem("~O~pen", CmUser, KbNone()),
		NewMenuItem("~N~ew", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// 'e' is at x=2, y=2 — non-shortcut on unselected item at index 1
	cell := cellAt(buf, 2, 2)
	if cell.Style != cs.MenuNormal {
		t.Errorf("Non-shortcut char 'e' style = %v, want MenuNormal (%v)", cell.Style, cs.MenuNormal)
	}
}

// ---------------------------------------------------------------------------
// 14. Rendering — selected item style
// ---------------------------------------------------------------------------

// TestDrawSelectedItemUsesMenuSelectedStyle verifies the selected row uses MenuSelected.
// Spec: "If the item is the selected item (selected == index), use MenuSelected style for the entire row."
func TestDrawSelectedItemUsesMenuSelectedStyle(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	p := NewMenuPopup(items, 0, 0) // selected = 0 initially
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// Item 0 is selected; row y=1; check a non-border cell, e.g. x=2 (inside row content)
	cell := cellAt(buf, 2, 1)
	if cell.Style != cs.MenuSelected {
		t.Errorf("Selected item row style = %v, want MenuSelected (%v)", cell.Style, cs.MenuSelected)
	}
}

// TestDrawUnselectedItemUsesMenuNormalNotSelected verifies non-selected row does NOT use MenuSelected.
// Falsifying: items other than the selected must not show MenuSelected style.
func TestDrawUnselectedItemUsesMenuNormalNotSelected(t *testing.T) {
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),   // selected (index 0)
		NewMenuItem("~O~pen", CmUser+1, KbNone()), // not selected
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// Item 1 is NOT selected; row y=2; interior cells must not have MenuSelected style
	cell := cellAt(buf, 2, 2)
	if cell.Style == cs.MenuSelected {
		t.Error("Non-selected item row must not use MenuSelected style")
	}
}

// TestDrawSelectedRowEntirelyMenuSelected verifies the entire selected row uses MenuSelected.
// Spec: "use MenuSelected style for the entire row."
func TestDrawSelectedRowEntirelyMenuSelected(t *testing.T) {
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// Row y=1 is the selected item; all interior cells x=1..w-2 should be MenuSelected
	for x := 1; x < w-1; x++ {
		cell := cellAt(buf, x, 1)
		if cell.Style != cs.MenuSelected {
			t.Errorf("Selected row at x=%d: style = %v, want MenuSelected", x, cell.Style)
		}
	}
}

// ---------------------------------------------------------------------------
// 15. Rendering — disabled item style
// ---------------------------------------------------------------------------

// TestDrawDisabledItemUsesMenuDisabledStyle verifies disabled item uses MenuDisabled style.
// Spec: "If item is disabled, use MenuDisabled style."
func TestDrawDisabledItemUsesMenuDisabledStyle(t *testing.T) {
	disabled := NewMenuItem("~N~ew", CmUser, KbNone())
	disabled.Disabled = true
	// Need a second selectable item so 'disabled' is not selected
	items := []any{
		NewMenuItem("~O~pen", CmUser+1, KbNone()), // selected (index 0)
		disabled,                                   // disabled, not selected (index 1)
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// Disabled item at y=2; check an interior cell
	cell := cellAt(buf, 2, 2)
	if cell.Style != cs.MenuDisabled {
		t.Errorf("Disabled item style = %v, want MenuDisabled (%v)", cell.Style, cs.MenuDisabled)
	}
}

// TestDrawDisabledItemDoesNotUseMenuNormal falsifies an impl that ignores disabled state.
func TestDrawDisabledItemDoesNotUseMenuNormal(t *testing.T) {
	disabled := NewMenuItem("~N~ew", CmUser, KbNone())
	disabled.Disabled = true
	items := []any{
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
		disabled,
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	cell := cellAt(buf, 2, 2)
	// The disabled item must NOT have MenuNormal (it must be MenuDisabled)
	if cell.Style == cs.MenuNormal {
		t.Error("Disabled item must not use MenuNormal style")
	}
}

// ---------------------------------------------------------------------------
// 16. Rendering — accelerator text
// ---------------------------------------------------------------------------

// TestDrawAcceleratorAppearsRightAligned verifies the accelerator text appears right-aligned.
// Spec: "right-aligned accelerator text in MenuNormal style."
func TestDrawAcceleratorAppearsRightAligned(t *testing.T) {
	// "~N~ew" Ctrl+N → accel "Ctrl+N" (6 chars); right-aligned means last char at x=w-2
	items := []any{NewMenuItem("~N~ew", CmUser, KbCtrl('N'))}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// Right border at x=w-1; rightmost accel char should be at x=w-2
	cell := cellAt(buf, w-2, 1)
	if cell.Rune != 'N' {
		t.Errorf("Rightmost accel char at (w-2, 1): got %q, want 'N' (last char of 'Ctrl+N')", cell.Rune)
	}
}

// TestDrawAcceleratorUsesMenuNormalStyle verifies unselected accel text uses MenuNormal style.
// Spec: "right-aligned accelerator text in MenuNormal style."
func TestDrawAcceleratorUsesMenuNormalStyle(t *testing.T) {
	// Use second (non-selected) item so the row uses MenuNormal/Shortcut, not MenuSelected
	items := []any{
		NewMenuItem("~O~pen", CmUser+1, KbNone()),        // selected
		NewMenuItem("~N~ew", CmUser, KbCtrl('N')),        // not selected; has accel
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// Non-selected item 1 at y=2; accel 'N' at x=w-2
	cell := cellAt(buf, w-2, 2)
	if cell.Style != cs.MenuNormal {
		t.Errorf("Accel text style = %v, want MenuNormal (%v)", cell.Style, cs.MenuNormal)
	}
}

// ---------------------------------------------------------------------------
// 17. Rendering — SubMenu shows label with ► suffix (treated as disabled)
// ---------------------------------------------------------------------------

// TestDrawSubMenuShowsArrowSuffix verifies SubMenu items render with ► suffix.
// Spec: "treated as disabled menu items showing the label with a ► suffix."
func TestDrawSubMenuShowsArrowSuffix(t *testing.T) {
	items := []any{
		NewMenuItem("~O~pen", CmUser+1, KbNone()), // selected (index 0)
		NewSubMenu("~F~ile"),                       // index 1, rendered with ► suffix
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// SubMenu at row y=2; the ► character should appear somewhere in that row
	found := false
	for x := 1; x < w-1; x++ {
		if cellAt(buf, x, 2).Rune == '►' {
			found = true
			break
		}
	}
	if !found {
		t.Error("SubMenu item must render with '►' suffix")
	}
}

// TestDrawSubMenuUsesMenuDisabledStyle verifies SubMenu items use MenuDisabled style.
// Spec: "treated as disabled menu items."
func TestDrawSubMenuUsesMenuDisabledStyle(t *testing.T) {
	items := []any{
		NewMenuItem("~O~pen", CmUser+1, KbNone()), // selected (index 0)
		NewSubMenu("~F~ile"),                       // disabled (index 1)
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// SubMenu row y=2; interior cell should be MenuDisabled
	cell := cellAt(buf, 2, 2)
	if cell.Style != cs.MenuDisabled {
		t.Errorf("SubMenu row style = %v, want MenuDisabled (%v)", cell.Style, cs.MenuDisabled)
	}
}

// ---------------------------------------------------------------------------
// 18. Rendering — background fill
// ---------------------------------------------------------------------------

// TestDrawItemRowFilledWithMenuNormal verifies item rows have MenuNormal background.
// Spec: "Background fill with MenuNormal style for all item rows."
func TestDrawItemRowFilledWithMenuNormal(t *testing.T) {
	items := []any{
		NewMenuItem("~O~pen", CmUser+1, KbNone()), // selected
		NewMenuItem("~Q~uit", CmQuit, KbNone()),   // not selected; row y=2 filled with MenuNormal
	}
	p := NewMenuPopup(items, 0, 0)
	cs := testCS()
	w, h := p.Bounds().Width(), p.Bounds().Height()
	buf := NewDrawBuffer(w, h)
	p.Draw(buf, cs)
	// Non-selected item 1 at y=2; a trailing space cell at end of label area should be MenuNormal
	// Find a cell that is past the label ('Quit' = 4 chars, starts at x=1, so x=5 onward is fill)
	// w = popupWidth = max(4,4)+4 = 8; x=5 should be fill
	if w > 6 {
		cell := cellAt(buf, w-3, 2)
		if cell.Style != cs.MenuNormal {
			t.Errorf("Background fill of non-selected item row at x=%d, y=2: style = %v, want MenuNormal", w-3, cell.Style)
		}
	}
}

func TestMenuPopupRuneShortcutDoesNotFireDisabledItem(t *testing.T) {
	// Spec: disabled items should not fire. Rune shortcut matching should
	// skip disabled items just as Enter and Click do.
	disabled := NewMenuItem("~N~ew", CmUser, KbNone())
	disabled.Disabled = true
	items := []any{disabled}
	p := NewMenuPopup(items, 0, 0)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'n'}}
	p.HandleEvent(ev)

	if p.Result() != 0 {
		t.Errorf("Rune shortcut on disabled item: got result %d, want 0 (should not fire)", p.Result())
	}
}
