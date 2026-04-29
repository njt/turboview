package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newStatusLineWithScheme builds a StatusLine with a simple test scheme, sets
// bounds, and returns both the line and the scheme for style assertions.
func newStatusLineWithScheme(items ...*StatusItem) (*StatusLine, *theme.ColorScheme) {
	sl := NewStatusLine(items...)
	sl.SetBounds(NewRect(0, 0, 60, 1))
	scheme := &theme.ColorScheme{
		StatusNormal:   tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorTeal),
		StatusShortcut: tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorTeal),
		StatusSelected: tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite),
	}
	sl.scheme = scheme
	return sl, scheme
}

// mousePress returns an EvMouse event with Button1 pressed at (x, 0).
func mousePress(x int) *Event {
	return &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: x, Y: 0, Button: tcell.Button1},
	}
}

// mouseMove returns an EvMouse event with Button1 still held, moved to (x, 0).
func mouseMove(x int) *Event {
	return &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: x, Y: 0, Button: tcell.Button1},
	}
}

// mouseRelease returns an EvMouse event with no buttons (release) at (x, 0).
func mouseRelease(x int) *Event {
	return &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: x, Y: 0, Button: 0},
	}
}

// ---------------------------------------------------------------------------
// Test 1: Click and release on same item fires EvCommand
// ---------------------------------------------------------------------------

// TestStatusLineMouseClickOnItemFiresCommand verifies that pressing and releasing
// Button1 over the same visible item transforms the release event to EvCommand.
// Spec: "Mouse release: if release is over the same item as the press, transform
// event to EvCommand with item's command"
func TestStatusLineMouseClickOnItemFiresCommand(t *testing.T) {
	// "AB" is 2 chars wide; Draw starts at x=1. Item spans x=1..2.
	item := NewStatusItem("AB", KbNone(), CmUser)
	sl, _ := newStatusLineWithScheme(item)

	// Press at x=1 (inside item).
	pressEv := mousePress(1)
	sl.HandleEvent(pressEv)

	// Release at x=1 (same item).
	releaseEv := mouseRelease(1)
	sl.HandleEvent(releaseEv)

	if releaseEv.What != EvCommand {
		t.Errorf("release on same item: What = %v, want EvCommand", releaseEv.What)
	}
	if releaseEv.Command != CmUser {
		t.Errorf("release on same item: Command = %v, want CmUser (%v)", releaseEv.Command, CmUser)
	}
}

// ---------------------------------------------------------------------------
// Test 2: Release on different item fires no command
// ---------------------------------------------------------------------------

// TestStatusLineMouseReleaseOnDifferentItemNoCommand verifies that releasing
// over a different item than pressed does not fire a command.
// Spec: "If release is over a different item (or no item), no command fires"
func TestStatusLineMouseReleaseOnDifferentItemNoCommand(t *testing.T) {
	// item1 "AB" at x=1..2, item2 "CD" at x=5..6 (2-space gap).
	item1 := NewStatusItem("AB", KbNone(), CmUser)
	item2 := NewStatusItem("CD", KbNone(), CmUser+1)
	sl, _ := newStatusLineWithScheme(item1, item2)

	// Press on item1 at x=1.
	pressEv := mousePress(1)
	sl.HandleEvent(pressEv)

	// Release on item2 at x=5.
	releaseEv := mouseRelease(5)
	sl.HandleEvent(releaseEv)

	if releaseEv.What == EvCommand {
		t.Errorf("release on different item: got EvCommand (cmd=%v), want no command", releaseEv.Command)
	}
}

// ---------------------------------------------------------------------------
// Test 3: Release outside any item fires no command
// ---------------------------------------------------------------------------

// TestStatusLineMouseReleaseOutsideItemsNoCommand verifies that pressing on an
// item but releasing outside all items does not fire a command.
// Spec: "If release is over a different item (or no item), no command fires"
func TestStatusLineMouseReleaseOutsideItemsNoCommand(t *testing.T) {
	item := NewStatusItem("AB", KbNone(), CmUser)
	sl, _ := newStatusLineWithScheme(item)

	// Press on item at x=1.
	pressEv := mousePress(1)
	sl.HandleEvent(pressEv)

	// Release at x=0 (before item starts).
	releaseEv := mouseRelease(0)
	sl.HandleEvent(releaseEv)

	if releaseEv.What == EvCommand {
		t.Errorf("release outside item: got EvCommand (cmd=%v), want no command", releaseEv.Command)
	}
}

// ---------------------------------------------------------------------------
// Test 4: Mouse move while held updates pressedIdx
// ---------------------------------------------------------------------------

// TestStatusLineMouseMoveWhileHeldUpdatesPressedIdx verifies that moving the
// mouse while Button1 is held updates the tracked pressed item index.
// Spec: "Mouse move while held: update pressedIdx to track cursor position"
// We verify indirectly: after moving to item2 and releasing there, the command
// for item2 fires (the new pressedIdx = item2, release is on item2 → same).
func TestStatusLineMouseMoveWhileHeldUpdatesPressedIdx(t *testing.T) {
	item1 := NewStatusItem("AB", KbNone(), CmUser)
	item2 := NewStatusItem("CD", KbNone(), CmUser+1)
	sl, _ := newStatusLineWithScheme(item1, item2)

	// Press on item1.
	pressEv := mousePress(1)
	sl.HandleEvent(pressEv)

	// Drag to item2.
	moveEv := mouseMove(5)
	sl.HandleEvent(moveEv)

	// Release on item2.
	releaseEv := mouseRelease(5)
	sl.HandleEvent(releaseEv)

	// pressedIdx tracked item2 during the move, so release on item2 = same item → fires.
	if releaseEv.What != EvCommand {
		t.Errorf("drag-to-item2 then release: What = %v, want EvCommand", releaseEv.What)
	}
	if releaseEv.Command != CmUser+1 {
		t.Errorf("drag-to-item2 then release: Command = %v, want CmUser+1 (%v)", releaseEv.Command, CmUser+1)
	}
}

// ---------------------------------------------------------------------------
// Test 5: HelpCtx-filtered items cannot be clicked
// ---------------------------------------------------------------------------

// TestStatusLineMouseClickOnFilteredItemNoCommand verifies that items excluded
// by the active HelpContext cannot be triggered by mouse clicks.
// Spec: "Items filtered by HelpCtx cannot be clicked"
func TestStatusLineMouseClickOnFilteredItemNoCommand(t *testing.T) {
	// This item requires HelpCtx=5; we'll keep activeCtx=0 so it's filtered.
	item := NewStatusItem("AB", KbNone(), CmUser).ForHelpCtx(5)
	sl, _ := newStatusLineWithScheme(item)
	// activeCtx stays at 0 (HcNoContext) — item is hidden.

	// Press then release at x=1 (where item would be if visible).
	pressEv := mousePress(1)
	sl.HandleEvent(pressEv)

	releaseEv := mouseRelease(1)
	sl.HandleEvent(releaseEv)

	if releaseEv.What == EvCommand {
		t.Errorf("click on filtered item: got EvCommand (cmd=%v), want no command", releaseEv.Command)
	}
}

// ---------------------------------------------------------------------------
// Test 6: StatusSelected style used for pressed item
// ---------------------------------------------------------------------------

// TestStatusLineDrawUsesStatusSelectedForPressedItem verifies that the pressed
// item is rendered with the StatusSelected style.
// Spec: "StatusSelected is used for the pressed item in Draw"
func TestStatusLineDrawUsesStatusSelectedForPressedItem(t *testing.T) {
	item := NewStatusItem("AB", KbNone(), CmUser)
	sl, scheme := newStatusLineWithScheme(item)

	// Simulate a press at x=1.
	pressEv := mousePress(1)
	sl.HandleEvent(pressEv)

	// Draw and inspect.
	buf := NewDrawBuffer(60, 1)
	sl.Draw(buf)

	// 'A' is at x=1; it should use StatusSelected style.
	cell := buf.GetCell(1, 0)
	if cell.Style != scheme.StatusSelected {
		t.Errorf("Draw pressed item: cell(1,0) style = %v, want StatusSelected %v",
			cell.Style, scheme.StatusSelected)
	}
}

// ---------------------------------------------------------------------------
// Test 7: Non-pressed item uses normal style (regression)
// ---------------------------------------------------------------------------

// TestStatusLineDrawUnpressedItemUsesNormalStyle verifies that unpressed items
// still use StatusNormal style (regression guard for the selected highlight).
// Spec: "Draw: unpressed items use StatusNormal / StatusShortcut as before"
func TestStatusLineDrawUnpressedItemUsesNormalStyle(t *testing.T) {
	item1 := NewStatusItem("AB", KbNone(), CmUser)
	item2 := NewStatusItem("CD", KbNone(), CmUser+1)
	sl, scheme := newStatusLineWithScheme(item1, item2)

	// Press item1 (x=1).
	pressEv := mousePress(1)
	sl.HandleEvent(pressEv)

	buf := NewDrawBuffer(60, 1)
	sl.Draw(buf)

	// item2 starts at x=5; 'C' should be StatusNormal, not StatusSelected.
	cell := buf.GetCell(5, 0)
	if cell.Style != scheme.StatusNormal {
		t.Errorf("Draw unpressed item: cell(5,0) style = %v, want StatusNormal %v",
			cell.Style, scheme.StatusNormal)
	}
}
