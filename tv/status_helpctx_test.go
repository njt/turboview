package tv

// Tests for the StatusLine HelpContext filtering feature.
//
// Each test maps to a specific requirement in the spec.  The feature covers
// four areas:
//
//  1. StatusItem builder — ForHelpCtx chaining method
//  2. StatusLine.Draw — HelpCtx-aware item filtering
//  3. StatusLine.HandleEvent — HelpCtx-aware keybinding filtering
//  4. StatusLine.SetActiveContext — setter behaviour
//  5. Application.resolveHelpCtx — focus-chain traversal
//  6. Application.Draw integration — resolveHelpCtx wired into Draw

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ─── helpers ────────────────────────────────────────────────────────────────

// bufContainsRune returns true when any cell in row 0 of buf holds ch.
func bufContainsRune(buf *DrawBuffer, ch rune) bool {
	for x := 0; x < buf.Width(); x++ {
		if buf.GetCell(x, 0).Rune == ch {
			return true
		}
	}
	return false
}

// firstRuneX returns the x-coordinate of the first occurrence of ch in row 0,
// or -1 if not found.
func firstRuneX(buf *DrawBuffer, ch rune) int {
	for x := 0; x < buf.Width(); x++ {
		if buf.GetCell(x, 0).Rune == ch {
			return x
		}
	}
	return -1
}

// drawStatusLine creates a 60-wide, 1-high DrawBuffer, sets that as sl's
// bounds, and calls sl.Draw(buf).  The buffer is returned for inspection.
func drawStatusLine(sl *StatusLine) *DrawBuffer {
	sl.SetBounds(NewRect(0, 0, 60, 1))
	buf := NewDrawBuffer(60, 1)
	sl.Draw(buf)
	return buf
}

// ─── 1. StatusItem builder ───────────────────────────────────────────────────

// TestNewStatusItemDefaultHelpCtxIsHcNoContext verifies that a StatusItem
// created without ForHelpCtx has HelpCtx == HcNoContext (0).
// Spec: "NewStatusItem(...) without ForHelpCtx has HelpCtx == HcNoContext (0)"
func TestNewStatusItemDefaultHelpCtxIsHcNoContext(t *testing.T) {
	item := NewStatusItem("~F1~ Help", KbFunc(1), CmUser)

	if item.HelpCtx != HcNoContext {
		t.Errorf("NewStatusItem HelpCtx = %d, want HcNoContext (%d)", item.HelpCtx, HcNoContext)
	}
}

// TestForHelpCtxSetsHelpCtxField verifies that ForHelpCtx stores the supplied
// value in the item's HelpCtx field.
// Spec: "NewStatusItem(...).ForHelpCtx(5) sets the item's HelpCtx to 5"
func TestForHelpCtxSetsHelpCtxField(t *testing.T) {
	item := NewStatusItem("~F2~ Save", KbFunc(2), CmUser+1).ForHelpCtx(5)

	if item.HelpCtx != 5 {
		t.Errorf("ForHelpCtx(5): HelpCtx = %d, want 5", item.HelpCtx)
	}
}

// TestForHelpCtxReturnsSamePointer verifies that ForHelpCtx returns the same
// *StatusItem so that builder chaining works.
// Spec: "ForHelpCtx returns the same *StatusItem for chaining"
func TestForHelpCtxReturnsSamePointer(t *testing.T) {
	original := NewStatusItem("~F3~ Open", KbFunc(3), CmUser+2)
	returned := original.ForHelpCtx(7)

	if returned != original {
		t.Errorf("ForHelpCtx did not return the same *StatusItem (got %p, want %p)", returned, original)
	}
}

// TestForHelpCtxZeroSetsHcNoContext verifies that ForHelpCtx(0) is equivalent
// to not calling it — the item becomes unconditionally visible.
// Spec: "Items with HelpCtx == HcNoContext (0) are always drawn"
func TestForHelpCtxZeroSetsHcNoContext(t *testing.T) {
	item := NewStatusItem("~F4~ Close", KbFunc(4), CmClose).ForHelpCtx(0)

	if item.HelpCtx != HcNoContext {
		t.Errorf("ForHelpCtx(0): HelpCtx = %d, want HcNoContext (%d)", item.HelpCtx, HcNoContext)
	}
}

// TestForHelpCtxCanBeCalledMultipleTimesLastValueWins verifies that repeated
// calls to ForHelpCtx on the same item use the last value (simple assignment).
// Spec: "ForHelpCtx sets the item's HelpCtx"
func TestForHelpCtxCanBeCalledMultipleTimesLastValueWins(t *testing.T) {
	item := NewStatusItem("~F5~ Run", KbFunc(5), CmUser+3).ForHelpCtx(3).ForHelpCtx(9)

	if item.HelpCtx != 9 {
		t.Errorf("After two ForHelpCtx calls: HelpCtx = %d, want 9", item.HelpCtx)
	}
}

// ─── 2. StatusLine.SetActiveContext ─────────────────────────────────────────

// TestSetActiveContextStoresValue verifies that calling SetActiveContext stores
// the supplied value so that subsequent Draw / HandleEvent calls use it.
// Spec: "SetActiveContext(hc HelpContext) sets the activeCtx field"
func TestSetActiveContextStoresValue(t *testing.T) {
	sl := NewStatusLine()
	sl.SetActiveContext(HelpContext(5))

	// We can only observe activeCtx indirectly via draw / event behaviour;
	// the simplest confirmation is that an item whose HelpCtx matches is drawn.
	item := NewStatusItem("A", KbNone(), CmUser).ForHelpCtx(5)
	sl2 := NewStatusLine(item)
	sl2.SetActiveContext(HelpContext(5))

	buf := drawStatusLine(sl2)
	if !bufContainsRune(buf, 'A') {
		t.Error("SetActiveContext(5): item with HelpCtx=5 should be drawn, but 'A' not found")
	}
}

// TestSetActiveContextOverwritesPreviousValue verifies that multiple calls
// overwrite the previous value and only the last value is in effect.
// Spec: "Multiple calls overwrite the previous value"
func TestSetActiveContextOverwritesPreviousValue(t *testing.T) {
	itemCtx3 := NewStatusItem("P", KbNone(), CmUser).ForHelpCtx(3)
	itemCtx7 := NewStatusItem("Q", KbNone(), CmUser+1).ForHelpCtx(7)
	sl := NewStatusLine(itemCtx3, itemCtx7)

	sl.SetActiveContext(HelpContext(3))
	sl.SetActiveContext(HelpContext(7)) // second call must overwrite

	buf := drawStatusLine(sl)

	if bufContainsRune(buf, 'P') {
		t.Error("SetActiveContext last=7: item with HelpCtx=3 should NOT be drawn")
	}
	if !bufContainsRune(buf, 'Q') {
		t.Error("SetActiveContext last=7: item with HelpCtx=7 should be drawn")
	}
}

// ─── 3. StatusLine.Draw — HelpCtx filtering ──────────────────────────────────

// TestDrawAlwaysShowsHcNoContextItems verifies that items with
// HelpCtx == HcNoContext are drawn regardless of the active context.
// Spec: "Items with HelpCtx == HcNoContext (0) are always drawn regardless of
//
//	active context"
func TestDrawAlwaysShowsHcNoContextItems(t *testing.T) {
	unconditional := NewStatusItem("U", KbNone(), CmUser) // HelpCtx == HcNoContext by default
	sl := NewStatusLine(unconditional)
	sl.SetActiveContext(HelpContext(99)) // arbitrary non-zero context

	buf := drawStatusLine(sl)

	if !bufContainsRune(buf, 'U') {
		t.Error("Draw: item with HcNoContext should always appear, but 'U' not found")
	}
}

// TestDrawAlwaysShowsHcNoContextItemsWhenActiveCtxIsAlsoHcNoContext confirms
// the trivial case: with no active context, unconditional items are shown.
// Spec: "Items with HelpCtx == HcNoContext are always drawn"
func TestDrawAlwaysShowsHcNoContextItemsWhenActiveCtxIsAlsoHcNoContext(t *testing.T) {
	unconditional := NewStatusItem("V", KbNone(), CmUser)
	sl := NewStatusLine(unconditional)
	// activeCtx is HcNoContext by default (zero value)

	buf := drawStatusLine(sl)

	if !bufContainsRune(buf, 'V') {
		t.Error("Draw: item with HcNoContext should appear at default activeCtx=0, but 'V' not found")
	}
}

// TestDrawHidesContextItemWhenActiveCtxDoesNotMatch verifies that an item with
// a non-zero HelpCtx is hidden when activeCtx != that HelpCtx.
// Spec: "Items with a non-zero HelpCtx are drawn only when StatusLine.activeCtx
//
//	matches their HelpCtx"
func TestDrawHidesContextItemWhenActiveCtxDoesNotMatch(t *testing.T) {
	contextItem := NewStatusItem("X", KbNone(), CmUser).ForHelpCtx(5)
	sl := NewStatusLine(contextItem)
	sl.SetActiveContext(HelpContext(3)) // does not match 5

	buf := drawStatusLine(sl)

	if bufContainsRune(buf, 'X') {
		t.Error("Draw: item with HelpCtx=5 should be hidden when activeCtx=3")
	}
}

// TestDrawShowsContextItemWhenActiveCtxMatches verifies that a context-specific
// item is drawn when activeCtx equals its HelpCtx.
// Spec: "Items with a non-zero HelpCtx are drawn only when StatusLine.activeCtx
//
//	matches their HelpCtx"
func TestDrawShowsContextItemWhenActiveCtxMatches(t *testing.T) {
	contextItem := NewStatusItem("Y", KbNone(), CmUser).ForHelpCtx(5)
	sl := NewStatusLine(contextItem)
	sl.SetActiveContext(HelpContext(5))

	buf := drawStatusLine(sl)

	if !bufContainsRune(buf, 'Y') {
		t.Error("Draw: item with HelpCtx=5 should be shown when activeCtx=5")
	}
}

// TestDrawHidesAllContextItemsWhenActiveCtxIsHcNoContext verifies that when
// activeCtx is HcNoContext only items with HelpCtx == HcNoContext are drawn.
// Spec: "When activeCtx is HcNoContext (0), only items with HelpCtx == HcNoContext
//
//	are shown"
func TestDrawHidesAllContextItemsWhenActiveCtxIsHcNoContext(t *testing.T) {
	contextItem := NewStatusItem("Z", KbNone(), CmUser).ForHelpCtx(4)
	sl := NewStatusLine(contextItem)
	// activeCtx stays at zero (HcNoContext)

	buf := drawStatusLine(sl)

	if bufContainsRune(buf, 'Z') {
		t.Error("Draw: item with HelpCtx=4 should be hidden when activeCtx=0")
	}
}

// TestDrawMixedItemsOnlyMatchingAndUnconditionalAreShown verifies that a mix of
// unconditional, matching, and non-matching items produces the correct visible
// subset.
// Spec: combined behaviour of HcNoContext always-shown + context matching
func TestDrawMixedItemsOnlyMatchingAndUnconditionalAreShown(t *testing.T) {
	always := NewStatusItem("A", KbNone(), CmUser)           // HcNoContext — always shown
	matchCtx := NewStatusItem("B", KbNone(), CmUser+1).ForHelpCtx(5) // shown when ctx=5
	wrongCtx := NewStatusItem("C", KbNone(), CmUser+2).ForHelpCtx(9) // hidden when ctx=5
	sl := NewStatusLine(always, matchCtx, wrongCtx)
	sl.SetActiveContext(HelpContext(5))

	buf := drawStatusLine(sl)

	if !bufContainsRune(buf, 'A') {
		t.Error("Draw: unconditional item 'A' should always be shown")
	}
	if !bufContainsRune(buf, 'B') {
		t.Error("Draw: item 'B' with HelpCtx=5 should be shown when activeCtx=5")
	}
	if bufContainsRune(buf, 'C') {
		t.Error("Draw: item 'C' with HelpCtx=9 should be hidden when activeCtx=5")
	}
}

// TestDrawFilteredItemsHaveNoGapInSpacing verifies that when items are filtered
// out, the remaining items are packed without gaps where the hidden item would
// have been.
// Spec: "When some items are filtered out, the remaining items are drawn with
//
//	correct spacing (no gaps where filtered items would have been)"
func TestDrawFilteredItemsHaveNoGapInSpacing(t *testing.T) {
	// Three items: first and third visible, middle hidden.
	// If filtering left a gap the third item would appear at the position it
	// would occupy if all three were drawn.  With correct packing it will be
	// adjacent to the first item (with the standard 2-space gap).
	first := NewStatusItem("AB", KbNone(), CmUser)                      // always shown
	middle := NewStatusItem("CD", KbNone(), CmUser+1).ForHelpCtx(9)    // hidden
	third := NewStatusItem("EF", KbNone(), CmUser+2)                    // always shown
	sl := NewStatusLine(first, middle, third)
	sl.SetActiveContext(HelpContext(3)) // middle not shown

	buf := drawStatusLine(sl)

	// "AB" starts at x=1 (Draw always starts at x=1).
	if buf.GetCell(1, 0).Rune != 'A' || buf.GetCell(2, 0).Rune != 'B' {
		t.Fatalf("Draw: 'AB' not at expected position (x=1,2); got %q %q",
			buf.GetCell(1, 0).Rune, buf.GetCell(2, 0).Rune)
	}

	// "EF" should start immediately after the 2-space gap: x=5.
	// If there were a gap for "CD", "EF" would start later.
	if buf.GetCell(5, 0).Rune != 'E' {
		t.Errorf("Draw: 'E' expected at x=5 (tight packing), got %q (gap may have been left for filtered item)",
			buf.GetCell(5, 0).Rune)
	}
	if buf.GetCell(6, 0).Rune != 'F' {
		t.Errorf("Draw: 'F' expected at x=6, got %q", buf.GetCell(6, 0).Rune)
	}
}

// TestDrawFilteredItemsSpacingMatchesAllVisibleItems confirms the same packing
// property by comparing against a StatusLine that never had the hidden item.
// Spec: "remaining items are drawn with correct spacing"
func TestDrawFilteredItemsSpacingMatchesAllVisibleItems(t *testing.T) {
	// Build a reference line with only the items that will survive filtering.
	ref1 := NewStatusItem("GH", KbNone(), CmUser)
	ref2 := NewStatusItem("IJ", KbNone(), CmUser+1)
	reference := NewStatusLine(ref1, ref2)
	refBuf := drawStatusLine(reference)

	// Build the filtered line: same visible items plus a hidden context item in between.
	vis1 := NewStatusItem("GH", KbNone(), CmUser)
	hidden := NewStatusItem("KL", KbNone(), CmUser+2).ForHelpCtx(7)
	vis2 := NewStatusItem("IJ", KbNone(), CmUser+1)
	filtered := NewStatusLine(vis1, hidden, vis2)
	filtered.SetActiveContext(HelpContext(2)) // 7 != 2, so "KL" is hidden
	filteredBuf := drawStatusLine(filtered)

	// The first 20 columns should match exactly.
	for x := 0; x < 20; x++ {
		rr := refBuf.GetCell(x, 0).Rune
		fr := filteredBuf.GetCell(x, 0).Rune
		if rr != fr {
			t.Errorf("col %d: reference has %q, filtered has %q — spacing diverged", x, rr, fr)
		}
	}
}

// ─── 4. StatusLine.HandleEvent — HelpCtx filtering ──────────────────────────

// TestHandleEventAlwaysMatchesHcNoContextItem verifies that a key bound to an
// unconditional item fires regardless of the active context.
// Spec: "An item with HelpCtx == HcNoContext always matches its keybinding"
func TestHandleEventAlwaysMatchesHcNoContextItem(t *testing.T) {
	item := NewStatusItem("~F1~ Help", KbFunc(1), CmUser) // HcNoContext
	sl := NewStatusLine(item)
	sl.SetActiveContext(HelpContext(42)) // non-zero, should not suppress unconditional

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF1}}
	sl.HandleEvent(event)

	if event.What != EvCommand || event.Command != CmUser {
		t.Errorf("HandleEvent: unconditional item should always match; What=%v Command=%v",
			event.What, event.Command)
	}
}

// TestHandleEventMatchesContextItemWhenActiveCtxMatches verifies that a key
// bound to a context item fires when activeCtx == item.HelpCtx.
// Spec: "Only items that would be drawn (passing the same HelpCtx filter) can
//
//	match keybindings"
func TestHandleEventMatchesContextItemWhenActiveCtxMatches(t *testing.T) {
	item := NewStatusItem("~F2~ Save", KbFunc(2), CmUser+1).ForHelpCtx(5)
	sl := NewStatusLine(item)
	sl.SetActiveContext(HelpContext(5))

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF2}}
	sl.HandleEvent(event)

	if event.What != EvCommand || event.Command != CmUser+1 {
		t.Errorf("HandleEvent: context item should match when activeCtx=5; What=%v Command=%v",
			event.What, event.Command)
	}
}

// TestHandleEventDoesNotMatchContextItemWhenActiveCtxDiffers verifies that a
// key bound to a context item does not fire when activeCtx != item.HelpCtx.
// Spec: "An item with HelpCtx == 5 and activeCtx == 3 does not match its keybinding"
func TestHandleEventDoesNotMatchContextItemWhenActiveCtxDiffers(t *testing.T) {
	item := NewStatusItem("~F3~ Load", KbFunc(3), CmUser+2).ForHelpCtx(5)
	sl := NewStatusLine(item)
	sl.SetActiveContext(HelpContext(3)) // 3 != 5

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF3}}
	sl.HandleEvent(event)

	if event.What == EvCommand {
		t.Errorf("HandleEvent: item with HelpCtx=5 should not match when activeCtx=3; got EvCommand")
	}
	if event.What != EvKeyboard {
		t.Errorf("HandleEvent: event.What should remain EvKeyboard when no item matches; got %v", event.What)
	}
}

// TestHandleEventDoesNotMatchContextItemWhenActiveCtxIsHcNoContext verifies that
// a context item is also suppressed when activeCtx is HcNoContext.
// Spec: "When activeCtx is HcNoContext, only items with HelpCtx == HcNoContext are shown"
//
// The same filter governs HandleEvent.
func TestHandleEventDoesNotMatchContextItemWhenActiveCtxIsHcNoContext(t *testing.T) {
	item := NewStatusItem("~F4~ Close", KbFunc(4), CmClose).ForHelpCtx(6)
	sl := NewStatusLine(item)
	// activeCtx is HcNoContext by default

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF4}}
	sl.HandleEvent(event)

	if event.What == EvCommand {
		t.Errorf("HandleEvent: item with HelpCtx=6 should not match when activeCtx=0")
	}
}

// TestHandleEventOnlyMatchesVisibleItemAmongMixed verifies the correct item is
// matched when a mix of unconditional and context items share keybindings that
// could otherwise collide.
// Spec: "Only items that would be drawn ... can match keybindings"
func TestHandleEventOnlyMatchesVisibleItemAmongMixed(t *testing.T) {
	// Both items bind F5 but belong to different contexts.
	itemAlways := NewStatusItem("~F5~ Global", KbFunc(5), CmUser)         // HcNoContext
	itemCtx8 := NewStatusItem("~F5~ Local", KbFunc(5), CmUser+1).ForHelpCtx(8) // ctx=8
	sl := NewStatusLine(itemAlways, itemCtx8)
	sl.SetActiveContext(HelpContext(3)) // ctx=8 item hidden; HcNoContext item visible

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF5}}
	sl.HandleEvent(event)

	// The unconditional item should win because the ctx=8 item is filtered.
	if event.What != EvCommand {
		t.Fatalf("HandleEvent: expected EvCommand, got %v", event.What)
	}
	if event.Command != CmUser {
		t.Errorf("HandleEvent: expected CmUser (unconditional item), got command %v", event.Command)
	}
}

// ─── 5. Application.resolveHelpCtx ──────────────────────────────────────────

// TestResolveHelpCtxNoDesktopReturnsHcNoContext verifies the base case: with no
// desktop no context can be resolved.
// Spec: "With no desktop or no focused views, returns HcNoContext"
func TestResolveHelpCtxNoDesktopReturnsHcNoContext(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Remove the desktop to test the no-desktop path.
	app.desktop = nil

	got := app.resolveHelpCtx()
	if got != HcNoContext {
		t.Errorf("resolveHelpCtx with no desktop = %d, want HcNoContext (%d)", got, HcNoContext)
	}
}

// TestResolveHelpCtxEmptyDesktopReturnsHcNoContext verifies that when the
// desktop exists but has no focused window, HcNoContext is returned.
// Spec: "With no desktop or no focused views, returns HcNoContext"
func TestResolveHelpCtxEmptyDesktopReturnsHcNoContext(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}
	// Desktop was created but nothing was inserted, so FocusedChild() == nil.

	got := app.resolveHelpCtx()
	if got != HcNoContext {
		t.Errorf("resolveHelpCtx with empty desktop = %d, want HcNoContext (%d)", got, HcNoContext)
	}
}

// TestResolveHelpCtxWindowOnlyReturnsWindowCtx verifies that when a window with
// a non-zero HelpCtx is focused and has no focused child with its own context,
// the window's HelpCtx is returned.
// Spec: "Returns the deepest (most specific) non-zero HelpCtx found in the chain"
// Spec: "If a Window has HelpCtx=1 and its focused Button has HelpCtx=0, returns 1"
func TestResolveHelpCtxWindowOnlyReturnsWindowCtx(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	win := NewWindow(NewRect(2, 2, 40, 20), "Test")
	win.SetHelpCtx(HelpContext(1))
	app.Desktop().Insert(win)

	got := app.resolveHelpCtx()
	if got != HelpContext(1) {
		t.Errorf("resolveHelpCtx = %d, want 1 (window context)", got)
	}
}

// TestResolveHelpCtxFocusedButtonOverridesWindowCtx verifies that when both
// window and button have non-zero HelpCtx, the button's value (deepest) is
// returned.
// Spec: "If a Window has HelpCtx=1 and its focused Button has HelpCtx=5,
//
//	returns 5"
func TestResolveHelpCtxFocusedButtonOverridesWindowCtx(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	win := NewWindow(NewRect(2, 2, 40, 20), "Test")
	win.SetHelpCtx(HelpContext(1))

	btn := NewButton(NewRect(1, 1, 10, 1), "~O~K", CmOK)
	btn.SetHelpCtx(HelpContext(5))
	win.Insert(btn)

	app.Desktop().Insert(win)

	got := app.resolveHelpCtx()
	if got != HelpContext(5) {
		t.Errorf("resolveHelpCtx = %d, want 5 (focused button context)", got)
	}
}

// TestResolveHelpCtxFallsBackToWindowWhenButtonCtxIsZero verifies the fallback
// to the parent's context when the leaf has HelpCtx == 0.
// Spec: "If a Window has HelpCtx=1 and its focused Button has HelpCtx=0,
//
//	returns 1"
func TestResolveHelpCtxFallsBackToWindowWhenButtonCtxIsZero(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	win := NewWindow(NewRect(2, 2, 40, 20), "Test")
	win.SetHelpCtx(HelpContext(1))

	btn := NewButton(NewRect(1, 1, 10, 1), "~O~K", CmOK)
	// btn.HelpCtx is HcNoContext by default (zero)
	win.Insert(btn)

	app.Desktop().Insert(win)

	got := app.resolveHelpCtx()
	if got != HelpContext(1) {
		t.Errorf("resolveHelpCtx = %d, want 1 (fallback to window when button ctx=0)", got)
	}
}

// TestResolveHelpCtxNoContextInChainReturnsHcNoContext verifies that when no
// view in the focus chain has a non-zero HelpCtx, HcNoContext is returned.
// Spec: "If no view in the chain has a non-zero HelpCtx, returns HcNoContext"
func TestResolveHelpCtxNoContextInChainReturnsHcNoContext(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	win := NewWindow(NewRect(2, 2, 40, 20), "Test")
	// win.HelpCtx left at 0

	btn := NewButton(NewRect(1, 1, 10, 1), "~O~K", CmOK)
	// btn.HelpCtx left at 0
	win.Insert(btn)

	app.Desktop().Insert(win)

	got := app.resolveHelpCtx()
	if got != HcNoContext {
		t.Errorf("resolveHelpCtx = %d, want HcNoContext when no view has a context", got)
	}
}

// TestResolveHelpCtxWalksFullChainToLeaf verifies the walk order: the most
// specific (deepest) non-zero context wins even when an ancestor also has a
// non-zero context.
// Spec: "Returns the deepest (most specific) non-zero HelpCtx found in the chain"
func TestResolveHelpCtxWalksFullChainToLeaf(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	win := NewWindow(NewRect(2, 2, 40, 20), "Outer")
	win.SetHelpCtx(HelpContext(10))

	btn := NewButton(NewRect(1, 1, 10, 1), "~D~one", CmOK)
	btn.SetHelpCtx(HelpContext(20))
	win.Insert(btn)

	app.Desktop().Insert(win)

	got := app.resolveHelpCtx()
	if got != HelpContext(20) {
		t.Errorf("resolveHelpCtx = %d, want 20 (deepest leaf context)", got)
	}
}

// ─── 6. Application.Draw integration ─────────────────────────────────────────

// TestDrawIntegrationSetsActiveContextBeforeDrawingStatusLine verifies that
// Application.Draw calls resolveHelpCtx() and passes the result to
// StatusLine.SetActiveContext() before drawing, so that context-sensitive items
// are shown or hidden correctly in the same draw cycle.
// Spec: "Before drawing the StatusLine, Application calls resolveHelpCtx() and
//
//	passes the result to StatusLine.SetActiveContext()"
func TestDrawIntegrationSetsActiveContextBeforeDrawingStatusLine(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	const w, h = 80, 25
	screen.SetSize(w, h)

	// Context item visible only when HelpCtx=5 is active.
	contextItem := NewStatusItem("Z", KbNone(), CmUser).ForHelpCtx(5)
	sl := NewStatusLine(contextItem)

	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Insert a window with HelpCtx=5; it will be the focused view, so
	// resolveHelpCtx() should return 5.
	win := NewWindow(NewRect(2, 2, 40, 20), "Test")
	win.SetHelpCtx(HelpContext(5))
	app.Desktop().Insert(win)

	buf := NewDrawBuffer(w, h)
	app.Draw(buf)

	// The status line occupies the last row (row h-1).
	// 'Z' should appear there because activeCtx was set to 5 by Draw.
	found := false
	for x := 0; x < w; x++ {
		if buf.GetCell(x, h-1).Rune == 'Z' {
			found = true
			break
		}
	}
	if !found {
		t.Error("Draw integration: context item 'Z' should appear in last row when focused window has HelpCtx=5")
	}
}

// TestDrawIntegrationHidesContextItemWhenFocusChanges verifies that context
// items disappear when focus shifts to a view with a different (or no) HelpCtx.
// Spec: "This happens on every draw cycle so context-sensitive items update
//
//	when focus changes"
func TestDrawIntegrationHidesContextItemWhenFocusChanges(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	const w, h = 80, 25
	screen.SetSize(w, h)

	contextItem := NewStatusItem("M", KbNone(), CmUser).ForHelpCtx(5)
	sl := NewStatusLine(contextItem)

	app, err := NewApplication(WithScreen(screen), WithStatusLine(sl))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// First window has HelpCtx=5 — item should appear.
	win1 := NewWindow(NewRect(0, 0, 30, 15), "Win1")
	win1.SetHelpCtx(HelpContext(5))
	app.Desktop().Insert(win1)

	buf1 := NewDrawBuffer(w, h)
	app.Draw(buf1)

	foundBefore := false
	for x := 0; x < w; x++ {
		if buf1.GetCell(x, h-1).Rune == 'M' {
			foundBefore = true
			break
		}
	}
	if !foundBefore {
		t.Error("Draw: 'M' should appear when focused window has HelpCtx=5")
	}

	// Second window has no HelpCtx — item should disappear.
	win2 := NewWindow(NewRect(30, 0, 30, 15), "Win2")
	// win2.HelpCtx stays at 0
	app.Desktop().Insert(win2) // Insert also focuses win2 (it is selectable)

	buf2 := NewDrawBuffer(w, h)
	app.Draw(buf2)

	foundAfter := false
	for x := 0; x < w; x++ {
		if buf2.GetCell(x, h-1).Rune == 'M' {
			foundAfter = true
			break
		}
	}
	if foundAfter {
		t.Error("Draw: 'M' should be hidden after focus shifts to window with no HelpCtx")
	}
}
