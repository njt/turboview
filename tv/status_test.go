package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// TestNewStatusItemStoresLabel verifies NewStatusItem stores the label field.
// Spec: "NewStatusItem(label, keyBinding, command) creates a StatusItem with the given fields"
func TestNewStatusItemStoresLabel(t *testing.T) {
	item := NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit)

	if item.Label != "~F10~ Quit" {
		t.Errorf("NewStatusItem label = %q, want %q", item.Label, "~F10~ Quit")
	}
}

// TestNewStatusItemStoresKeyBinding verifies NewStatusItem stores the key binding.
// Spec: "NewStatusItem(label, keyBinding, command) creates a StatusItem with the given fields"
func TestNewStatusItemStoresKeyBinding(t *testing.T) {
	kb := KbFunc(10)
	item := NewStatusItem("~F10~ Quit", kb, CmQuit)

	if item.KeyBinding != kb {
		t.Errorf("NewStatusItem key binding = %v, want %v", item.KeyBinding, kb)
	}
}

// TestNewStatusItemStoresCommand verifies NewStatusItem stores the command code.
// Spec: "NewStatusItem(label, keyBinding, command) creates a StatusItem with the given fields"
func TestNewStatusItemStoresCommand(t *testing.T) {
	item := NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit)

	if item.Command != CmQuit {
		t.Errorf("NewStatusItem command = %v, want %v", item.Command, CmQuit)
	}
}

// TestNewStatusLineSetsSfVisible verifies that NewStatusLine sets the SfVisible state flag.
// Spec: "NewStatusLine(items...) creates a StatusLine with SfVisible state"
func TestNewStatusLineSetsSfVisible(t *testing.T) {
	sl := NewStatusLine()

	if !sl.HasState(SfVisible) {
		t.Errorf("NewStatusLine did not set SfVisible")
	}
}

// TestStatusLineDrawFillsRowWithStatusNormalStyle verifies that Draw fills its row
// with the StatusNormal style as the background.
// Spec: "StatusLine.Draw(buf) fills its row with StatusNormal style"
func TestStatusLineDrawFillsRowWithStatusNormalStyle(t *testing.T) {
	sl := NewStatusLine()
	sl.SetBounds(NewRect(0, 0, 20, 1))
	scheme := &theme.ColorScheme{
		StatusNormal:   tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorTeal),
		StatusShortcut: tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorTeal),
	}
	sl.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	sl.Draw(buf)

	// A cell that is not part of any item label should use StatusNormal style.
	// With no items, every cell should be StatusNormal.
	for x := 0; x < 20; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Style != scheme.StatusNormal {
			t.Errorf("Draw cell (%d,0) style = %v, want StatusNormal", x, cell.Style)
			break
		}
	}
}

// TestStatusLineDrawRendersShortcutTextInStatusShortcutStyle verifies that tilde-delimited
// shortcut segments are drawn with StatusShortcut style.
// Spec: "renders each item's label with tilde shortcut segments in StatusShortcut style"
func TestStatusLineDrawRendersShortcutTextInStatusShortcutStyle(t *testing.T) {
	// Label "~F10~" — the shortcut segment is "F10"
	item := NewStatusItem("~F10~Quit", KbFunc(10), CmQuit)
	sl := NewStatusLine(item)
	sl.SetBounds(NewRect(0, 0, 40, 1))

	normalStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorTeal)
	shortcutStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorTeal)
	scheme := &theme.ColorScheme{
		StatusNormal:   normalStyle,
		StatusShortcut: shortcutStyle,
	}
	sl.scheme = scheme

	buf := NewDrawBuffer(40, 1)
	sl.Draw(buf)

	// The shortcut segment "F10" starts at the first character of the row.
	// Scan across the row for a cell bearing the shortcut style.
	found := false
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 0).Style == shortcutStyle {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Draw did not render any cell with StatusShortcut style")
	}
}

// TestStatusLineDrawRendersNormalTextInStatusNormalStyle verifies that non-shortcut
// label segments are drawn with StatusNormal style.
// Spec: "normal text in StatusNormal style"
func TestStatusLineDrawRendersNormalTextInStatusNormalStyle(t *testing.T) {
	// Label "~F10~ Quit" — shortcut is "F10", normal text is " Quit"
	item := NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit)
	sl := NewStatusLine(item)
	sl.SetBounds(NewRect(0, 0, 40, 1))

	normalStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorTeal)
	shortcutStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorTeal)
	scheme := &theme.ColorScheme{
		StatusNormal:   normalStyle,
		StatusShortcut: shortcutStyle,
	}
	sl.scheme = scheme

	buf := NewDrawBuffer(40, 1)
	sl.Draw(buf)

	// Find the 'Q' of "Quit" — it must have normal style, not shortcut.
	found := false
	for x := 0; x < 40; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune == 'Q' {
			found = true
			if cell.Style != normalStyle {
				t.Errorf("Draw 'Q' cell style = %v, want StatusNormal", cell.Style)
			}
			break
		}
	}
	if !found {
		t.Errorf("Draw did not render the 'Q' from item label")
	}
}

// TestStatusLineDrawSeparatesItemsWithTwoSpaceGap verifies that consecutive items
// are separated by exactly two space characters.
// Spec: "separated by 2-space gaps"
func TestStatusLineDrawSeparatesItemsWithTwoSpaceGap(t *testing.T) {
	// Use simple labels with no tildes so we can track exact column positions.
	item1 := NewStatusItem("AB", KbFunc(1), CmUser)
	item2 := NewStatusItem("CD", KbFunc(2), CmUser+1)
	sl := NewStatusLine(item1, item2)
	sl.SetBounds(NewRect(0, 0, 20, 1))

	normalStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorTeal)
	scheme := &theme.ColorScheme{
		StatusNormal:   normalStyle,
		StatusShortcut: normalStyle,
	}
	sl.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	sl.Draw(buf)

	// Locate "AB" and "CD". Implementation starts at x=1 with 2-space gap between items.
	// "AB" occupies columns 1-2. Gap at columns 3-4. "CD" starts at column 5.
	if buf.GetCell(1, 0).Rune != 'A' {
		t.Fatalf("expected 'A' at col 1, got %q", buf.GetCell(1, 0).Rune)
	}
	if buf.GetCell(2, 0).Rune != 'B' {
		t.Fatalf("expected 'B' at col 2, got %q", buf.GetCell(2, 0).Rune)
	}

	// Columns 3 and 4 must be spaces (the gap).
	for _, col := range []int{3, 4} {
		if buf.GetCell(col, 0).Rune != ' ' {
			t.Errorf("gap at col %d: got %q, want ' '", col, buf.GetCell(col, 0).Rune)
		}
	}

	// "CD" must start at column 5.
	if buf.GetCell(5, 0).Rune != 'C' {
		t.Errorf("second item at col 5: got %q, want 'C'", buf.GetCell(5, 0).Rune)
	}
	if buf.GetCell(6, 0).Rune != 'D' {
		t.Errorf("second item at col 6: got %q, want 'D'", buf.GetCell(6, 0).Rune)
	}
}

// TestStatusLineHandleEventTransformsMatchingKeyboardEventToCommand verifies that
// HandleEvent changes the event type from EvKeyboard to EvCommand on a matching key.
// Spec: "on match, transforms the event from EvKeyboard to EvCommand"
func TestStatusLineHandleEventTransformsMatchingKeyboardEventToCommand(t *testing.T) {
	item := NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit)
	sl := NewStatusLine(item)

	event := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyF10},
	}
	sl.HandleEvent(event)

	if event.What != EvCommand {
		t.Errorf("HandleEvent: What = %v, want EvCommand", event.What)
	}
}

// TestStatusLineHandleEventSetsCorrectCommandCode verifies that HandleEvent sets
// the Command field to the matched item's command code.
// Spec: "transforms the event from EvKeyboard to EvCommand with the item's command code"
func TestStatusLineHandleEventSetsCorrectCommandCode(t *testing.T) {
	item := NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit)
	sl := NewStatusLine(item)

	event := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyF10},
	}
	sl.HandleEvent(event)

	if event.Command != CmQuit {
		t.Errorf("HandleEvent command = %v, want CmQuit (%v)", event.Command, CmQuit)
	}
}

// TestStatusLineHandleEventClearsKeyFieldOnMatch verifies that HandleEvent sets
// the Key field to nil when transforming to a command event.
// Spec: "transforms the event from EvKeyboard to EvCommand"
func TestStatusLineHandleEventClearsKeyFieldOnMatch(t *testing.T) {
	item := NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit)
	sl := NewStatusLine(item)

	ke := &KeyEvent{Key: tcell.KeyF10}
	event := &Event{
		What: EvKeyboard,
		Key:  ke,
	}
	sl.HandleEvent(event)

	if event.Key != nil {
		t.Errorf("HandleEvent: Key field = %v after transform, want nil", event.Key)
	}
}

// TestStatusLineHandleEventDoesNothingForNonKeyboardEvent verifies that HandleEvent
// leaves non-keyboard events untouched.
// Spec: "HandleEvent does nothing for non-keyboard events"
func TestStatusLineHandleEventDoesNothingForNonKeyboardEvent(t *testing.T) {
	item := NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit)
	sl := NewStatusLine(item)

	event := &Event{What: EvMouse, Mouse: &MouseEvent{X: 1, Y: 0}}
	sl.HandleEvent(event)

	if event.What != EvMouse {
		t.Errorf("HandleEvent changed What for mouse event: got %v, want EvMouse", event.What)
	}
}

// TestStatusLineHandleEventDoesNothingForNonMatchingKey verifies that HandleEvent
// leaves the event untouched when no item binding matches.
// Spec: "HandleEvent checks keyboard events against item key bindings; on match, transforms"
// — by contrapositive, no match means no transformation.
func TestStatusLineHandleEventDoesNothingForNonMatchingKey(t *testing.T) {
	item := NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit)
	sl := NewStatusLine(item)

	// F9 is not bound.
	event := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyF9},
	}
	sl.HandleEvent(event)

	if event.What != EvKeyboard {
		t.Errorf("HandleEvent changed What for non-matching key: got %v, want EvKeyboard", event.What)
	}
	if event.Command != 0 {
		t.Errorf("HandleEvent set Command for non-matching key: got %v, want 0", event.Command)
	}
}

// TestStatusLineHandleEventMatchesCorrectItemAmongMultiple verifies that
// HandleEvent selects the correct item's command when multiple items are present.
// Spec: "checks keyboard events against item key bindings; on match, transforms the event
// ... with the item's command code"
func TestStatusLineHandleEventMatchesCorrectItemAmongMultiple(t *testing.T) {
	item1 := NewStatusItem("~F1~ Help", KbFunc(1), CmUser)
	item2 := NewStatusItem("~F10~ Quit", KbFunc(10), CmQuit)
	sl := NewStatusLine(item1, item2)

	event := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyF1},
	}
	sl.HandleEvent(event)

	if event.What != EvCommand {
		t.Errorf("HandleEvent: What = %v, want EvCommand", event.What)
	}
	if event.Command != CmUser {
		t.Errorf("HandleEvent command = %v, want CmUser (%v)", event.Command, CmUser)
	}
}
