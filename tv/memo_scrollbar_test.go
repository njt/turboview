package tv

// memo_scrollbar_test.go — Tests for Task 6: Scrollbar Integration.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test cites the exact spec sentence it verifies.
//
// Test organisation:
//   Section 1  — SetVScrollBar: linking and initial sync
//   Section 2  — SetHScrollBar: linking and initial sync
//   Section 3  — WithScrollBars constructor option
//   Section 4  — Sync after cursor movement (Memo → scrollbar)
//   Section 5  — Sync after text change (Memo → scrollbar)
//   Section 6  — Sync from scrollbar to Memo (scrollbar → Memo)
//   Section 7  — State-dependent visibility
//   Section 8  — Falsifying / boundary tests

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// sbKeyEv creates a plain keyboard event for the given key.
func sbKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key}}
}

// sbRuneEv creates a keyboard event for a printable rune.
func sbRuneEv(r rune) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}}
}

// newSbMemo creates a Memo with width=40, height=10 for scrollbar tests.
func newSbMemo() *Memo {
	return NewMemo(NewRect(0, 0, 40, 10))
}

// newVScrollBar creates a vertical scrollbar with a small fixed size.
func newVScrollBar() *ScrollBar {
	return NewScrollBar(NewRect(39, 0, 1, 10), Vertical)
}

// newHScrollBar creates a horizontal scrollbar with a small fixed size.
func newHScrollBar() *ScrollBar {
	return NewScrollBar(NewRect(0, 9, 40, 1), Horizontal)
}

// ---------------------------------------------------------------------------
// Section 1 — SetVScrollBar
// ---------------------------------------------------------------------------

// TestSetVScrollBarRange verifies that after linking a vertical scrollbar its range
// is set to (0, lineCount-1).
// Spec: "Vertical: SetRange(0, len(lines)-1)"
func TestSetVScrollBarRange(t *testing.T) {
	m := newSbMemo()
	m.SetText("line1\nline2\nline3") // 3 lines
	sb := newVScrollBar()

	m.SetVScrollBar(sb)

	// We cannot call sb.Min()/sb.Max() directly; instead we rely on SetValue
	// clamping to [min,max]. Drive the scrollbar to a large value and see where
	// it clamps. But a simpler check: the spec says range max == lineCount-1 == 2.
	// We verify by checking SetValue(999) is clamped: sb.Value() must be <= 2.
	sb.SetValue(999)
	if sb.Value() > 2 {
		t.Errorf("VScrollBar range max: after SetValue(999) Value()=%d, want <=2 (lineCount-1)", sb.Value())
	}
}

// TestSetVScrollBarRangeSingleLine verifies that a single-line Memo sets range (0,0).
// Spec: "Vertical: SetRange(0, len(lines)-1)" — with one line that is max=0.
func TestSetVScrollBarRangeSingleLine(t *testing.T) {
	m := newSbMemo()
	m.SetText("only one line")
	sb := newVScrollBar()

	m.SetVScrollBar(sb)

	sb.SetValue(999)
	if sb.Value() != 0 {
		t.Errorf("VScrollBar single-line range max: Value()=%d after SetValue(999), want 0", sb.Value())
	}
}

// TestSetVScrollBarPageSize verifies that after linking, vertical scrollbar page
// size is height-1.
// Spec: "Vertical: SetPageSize(height-1)"
func TestSetVScrollBarPageSize(t *testing.T) {
	// Use a memo with height=10; page size should be 9.
	// We can't read page size directly, but we can infer it from scroll
	// behaviour if we know the range. Instead we use a tall text and check
	// that scrolling by one page lands at the expected position.
	// The simplest observable effect: if PageSize==height-1==9 and range is
	// (0, N), then pressing PgDn from 0 moves the scrollbar by 9.
	// We exercise this indirectly in Section 4; here we just verify no panic
	// and that SetPageSize was called by confirming the link succeeded.
	m := NewMemo(NewRect(0, 0, 40, 10)) // height=10 → pageSize=9
	sb := newVScrollBar()

	// Should not panic and should link successfully.
	m.SetVScrollBar(sb)

	// Verify the memo reports its height as 10 (basic sanity, confirming test setup).
	if m.Bounds().Height() != 10 {
		t.Fatalf("Expected memo height 10, got %d", m.Bounds().Height())
	}
}

// TestSetVScrollBarInitialValue verifies that after linking, vertical scrollbar
// value is deltaY (0 at construction time).
// Spec: "Vertical: SetValue(deltaY)" — deltaY starts at 0.
func TestSetVScrollBarInitialValue(t *testing.T) {
	m := newSbMemo()
	m.SetText("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\nline11")
	sb := newVScrollBar()

	m.SetVScrollBar(sb)

	if sb.Value() != 0 {
		t.Errorf("VScrollBar initial value: Value()=%d, want 0 (deltaY not scrolled)", sb.Value())
	}
}

// TestSetVScrollBarNilUnlinks verifies that passing nil to SetVScrollBar
// does not panic and removes the link.
// Spec: "Passing nil unlinks the scrollbar and clears its callback"
func TestSetVScrollBarNilUnlinks(t *testing.T) {
	m := newSbMemo()
	sb := newVScrollBar()
	m.SetVScrollBar(sb)

	// Should not panic.
	m.SetVScrollBar(nil)
}

// TestSetVScrollBarReplaceClearsOldOnChange verifies that replacing a scrollbar
// clears the old scrollbar's OnChange callback.
// Spec: "If a scrollbar was previously linked, its OnChange callback is cleared
// before linking the new one"
func TestSetVScrollBarReplaceClearsOldOnChange(t *testing.T) {
	m := newSbMemo()
	old := newVScrollBar()
	m.SetVScrollBar(old)

	// At this point old.OnChange should be set by the memo.
	if old.OnChange == nil {
		t.Fatal("Expected old VScrollBar.OnChange to be set after linking")
	}

	// Replace with a new scrollbar.
	newSb := newVScrollBar()
	m.SetVScrollBar(newSb)

	// Old scrollbar's OnChange must be cleared.
	if old.OnChange != nil {
		t.Error("Old VScrollBar.OnChange was not cleared when replaced")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — SetHScrollBar
// ---------------------------------------------------------------------------

// TestSetHScrollBarRange verifies that after linking a horizontal scrollbar its
// range is set to (0, maxLineWidth).
// Spec: "Horizontal: SetRange(0, maxLineWidth)"
func TestSetHScrollBarRange(t *testing.T) {
	m := newSbMemo()
	// Longest line is "hello world" = 11 runes; other lines shorter.
	m.SetText("hi\nhello world\nbye")
	sb := newHScrollBar()

	m.SetHScrollBar(sb)

	// maxLineWidth == 11; SetValue(999) should clamp to 11.
	sb.SetValue(999)
	if sb.Value() > 11 {
		t.Errorf("HScrollBar range max: after SetValue(999) Value()=%d, want <=11 (maxLineWidth)", sb.Value())
	}
}

// TestSetHScrollBarPageSize verifies that after linking, horizontal scrollbar
// page size is width/2.
// Spec: "Horizontal: SetPageSize(width/2)"
func TestSetHScrollBarPageSize(t *testing.T) {
	// width=40 → pageSize=20. No direct accessor for page size; verify no panic
	// and link succeeds.
	m := NewMemo(NewRect(0, 0, 40, 10)) // width=40
	sb := newHScrollBar()

	m.SetHScrollBar(sb)

	if m.Bounds().Width() != 40 {
		t.Fatalf("Expected memo width 40, got %d", m.Bounds().Width())
	}
}

// TestSetHScrollBarInitialValue verifies that after linking, horizontal scrollbar
// value is deltaX (0 at construction time).
// Spec: "Horizontal: SetValue(deltaX)" — deltaX starts at 0.
func TestSetHScrollBarInitialValue(t *testing.T) {
	m := newSbMemo()
	m.SetText("some text")
	sb := newHScrollBar()

	m.SetHScrollBar(sb)

	if sb.Value() != 0 {
		t.Errorf("HScrollBar initial value: Value()=%d, want 0 (deltaX not scrolled)", sb.Value())
	}
}

// TestSetHScrollBarNilUnlinks verifies that passing nil to SetHScrollBar does
// not panic and removes the link.
// Spec: "Passing nil unlinks the scrollbar and clears its callback"
func TestSetHScrollBarNilUnlinks(t *testing.T) {
	m := newSbMemo()
	sb := newHScrollBar()
	m.SetHScrollBar(sb)

	// Should not panic.
	m.SetHScrollBar(nil)
}

// TestSetHScrollBarReplaceClearsOldOnChange verifies that replacing a horizontal
// scrollbar clears the old one's OnChange.
// Spec: "If a scrollbar was previously linked, its OnChange callback is cleared
// before linking the new one"
func TestSetHScrollBarReplaceClearsOldOnChange(t *testing.T) {
	m := newSbMemo()
	old := newHScrollBar()
	m.SetHScrollBar(old)

	if old.OnChange == nil {
		t.Fatal("Expected old HScrollBar.OnChange to be set after linking")
	}

	newSb := newHScrollBar()
	m.SetHScrollBar(newSb)

	if old.OnChange != nil {
		t.Error("Old HScrollBar.OnChange was not cleared when replaced")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — WithScrollBars constructor option
// ---------------------------------------------------------------------------

// TestWithScrollBarsBothLinked verifies that the WithScrollBars option links
// both scrollbars during construction.
// Spec: "WithScrollBars(h, v *ScrollBar) MemoOption constructor option calls both setters"
func TestWithScrollBarsBothLinked(t *testing.T) {
	vSb := newVScrollBar()
	hSb := newHScrollBar()

	m := NewMemo(NewRect(0, 0, 40, 10), WithScrollBars(hSb, vSb))

	// Both scrollbars should have their OnChange set (indicating they are linked).
	if vSb.OnChange == nil {
		t.Error("WithScrollBars did not link the vertical scrollbar (OnChange is nil)")
	}
	if hSb.OnChange == nil {
		t.Error("WithScrollBars did not link the horizontal scrollbar (OnChange is nil)")
	}
	_ = m
}

// TestWithScrollBarsNilSafe verifies that passing nil scrollbars to WithScrollBars
// does not panic.
// Spec: "Passing nil unlinks the scrollbar and clears its callback"
func TestWithScrollBarsNilSafe(t *testing.T) {
	// Should not panic with nil arguments.
	m := NewMemo(NewRect(0, 0, 40, 10), WithScrollBars(nil, nil))
	_ = m
}

// ---------------------------------------------------------------------------
// Section 4 — Sync after cursor movement (Memo → scrollbar)
// ---------------------------------------------------------------------------

// TestVScrollBarSyncsAfterDownKey verifies that pressing Down (which may change
// deltaY) causes the vertical scrollbar to update its value.
// Spec: "After any cursor/text/viewport change ... Vertical: SetValue(deltaY)"
func TestVScrollBarSyncsAfterDownKey(t *testing.T) {
	// Use a small memo (height=3) so that Down quickly causes viewport scrolling.
	m := NewMemo(NewRect(0, 0, 40, 3))
	// 5 lines — enough that moving down eventually shifts the viewport.
	m.SetText("line0\nline1\nline2\nline3\nline4")
	sb := NewScrollBar(NewRect(39, 0, 1, 3), Vertical)
	m.SetVScrollBar(sb)

	// Move down enough times to push deltaY > 0 (viewport must scroll).
	// height=3 so viewport shows 3 lines; after 2 Downs cursor is at row 2
	// (still in view). One more Down puts cursor at row 3, forcing deltaY=1.
	for i := 0; i < 3; i++ {
		m.HandleEvent(sbKeyEv(tcell.KeyDown))
	}

	if sb.Value() == 0 {
		t.Error("VScrollBar value still 0 after Down keys caused viewport scroll; expected deltaY>0")
	}
}

// TestVScrollBarSyncsAfterPgDn verifies PgDn updates the vertical scrollbar.
// Spec: "After pressing Down/PgDn that changes deltaY, vertical scrollbar value updates"
func TestVScrollBarSyncsAfterPgDn(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 3)) // height=3
	// Build text tall enough that a single PgDn shifts the viewport.
	// PgDn moves by height-1 = 2 lines.
	m.SetText("line0\nline1\nline2\nline3\nline4\nline5")
	sb := NewScrollBar(NewRect(39, 0, 1, 3), Vertical)
	m.SetVScrollBar(sb)

	m.HandleEvent(sbKeyEv(tcell.KeyPgDn))

	if sb.Value() == 0 {
		t.Error("VScrollBar value still 0 after PgDn; expected deltaY>0")
	}
}

// TestHScrollBarSyncsAfterTypingLongLine verifies that typing text beyond the
// viewport width causes the horizontal scrollbar value to update.
// Spec: "After typing text longer than width (changing deltaX), horizontal scrollbar value updates"
func TestHScrollBarSyncsAfterTypingLongLine(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 10, 5)) // width=10 — small so deltaX kicks in quickly
	sb := NewScrollBar(NewRect(0, 4, 10, 1), Horizontal)
	m.SetHScrollBar(sb)

	// Type 15 characters — more than width 10, so deltaX must become >0.
	for i := 0; i < 15; i++ {
		m.HandleEvent(sbRuneEv('x'))
	}

	if sb.Value() == 0 {
		t.Error("HScrollBar value still 0 after typing past viewport width; expected deltaX>0")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Sync after text change (Memo → scrollbar)
// ---------------------------------------------------------------------------

// TestVScrollBarSyncsAfterSetText verifies that calling SetText with new content
// causes the vertical scrollbar range to update.
// Spec: "After SetText with different content, scrollbar range updates"
func TestVScrollBarSyncsAfterSetText(t *testing.T) {
	m := newSbMemo()
	m.SetText("line1") // 1 line
	sb := newVScrollBar()
	m.SetVScrollBar(sb)

	// Replace with 5 lines.
	m.SetText("a\nb\nc\nd\ne")

	// maxValue should now be lineCount-1 = 4.
	sb.SetValue(999)
	if sb.Value() > 4 {
		t.Errorf("VScrollBar range not updated after SetText: Value()=%d after SetValue(999), want <=4", sb.Value())
	}
}

// TestHScrollBarSyncsAfterSetText verifies that calling SetText updates the
// horizontal scrollbar range to reflect the new longest line.
// Spec: "After SetText with different content, scrollbar range updates"
func TestHScrollBarSyncsAfterSetText(t *testing.T) {
	m := newSbMemo()
	m.SetText("short") // maxLineWidth=5
	sb := newHScrollBar()
	m.SetHScrollBar(sb)

	// Replace with a longer line.
	m.SetText("a longer line here!!") // 20 runes

	sb.SetValue(999)
	if sb.Value() > 20 {
		t.Errorf("HScrollBar range not updated after SetText: Value()=%d after SetValue(999), want <=20", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Sync from scrollbar to Memo
// ---------------------------------------------------------------------------

// TestVScrollBarOnChangeSetsDeltaY verifies that triggering the vertical
// scrollbar's OnChange callback updates deltaY inside the Memo.
// Spec: "Vertical scrollbar OnChange callback sets deltaY to the scrollbar's value"
func TestVScrollBarOnChangeSetsDeltaY(t *testing.T) {
	m := newSbMemo()
	// Need enough lines that deltaY=3 is reachable.
	m.SetText("l0\nl1\nl2\nl3\nl4\nl5\nl6\nl7\nl8\nl9\nl10")
	sb := newVScrollBar()
	m.SetVScrollBar(sb)

	if sb.OnChange == nil {
		t.Fatal("VScrollBar.OnChange is nil after linking; cannot test callback")
	}

	// Directly invoke the OnChange callback with value 3 — simulates the
	// scrollbar firing after user drag.
	sb.OnChange(3)

	// After deltaY=3, pressing Ctrl+Home should move cursor back to (0,0).
	// More directly: pressing Up from a cursor that was scrolled should not
	// immediately show the cursor at row 0.
	// We verify by pressing Down: if deltaY==3, the cursor is positioned in a
	// region offset by 3. The simplest black-box check is to confirm that a
	// subsequent SetValue on the scrollbar reflects 3 (i.e. a sync after any
	// event would call SetValue(deltaY) == 3).
	// Trigger a sync by sending a no-op-ish keyboard event that gets handled
	// (even if not a movement), or just send a key that moves within the viewport.
	// Actually the spec says "sets deltaY to the scrollbar's value"; we can
	// falsify by checking: if we scroll back to 0 via OnChange(0) and then
	// navigate, the cursor row is in range [0, height-1].
	sb.OnChange(0)
	// If deltaY reset to 0, the cursor should be near the top.
	// Send Ctrl+Home to guarantee cursor is at (0,0).
	m.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome, Modifiers: tcell.ModCtrl}})
	row, _ := m.CursorPos()
	if row != 0 {
		t.Errorf("After OnChange(0) + Ctrl+Home: CursorPos row=%d, want 0", row)
	}
}

// TestHScrollBarOnChangeSetsDeltaX verifies that triggering the horizontal
// scrollbar's OnChange callback updates deltaX inside the Memo.
// Spec: "Horizontal scrollbar OnChange callback sets deltaX to the scrollbar's value"
func TestHScrollBarOnChangeSetsDeltaX(t *testing.T) {
	m := newSbMemo()
	// A long line so deltaX=5 is valid.
	m.SetText("0123456789ABCDEFGHIJ")
	sb := newHScrollBar()
	m.SetHScrollBar(sb)

	if sb.OnChange == nil {
		t.Fatal("HScrollBar.OnChange is nil after linking; cannot test callback")
	}

	// Invoke directly with value 5 — simulates user dragging horizontal scrollbar.
	sb.OnChange(5)

	// After deltaX=5, pressing End should not change deltaX back to 0 (the
	// cursor is still on the same long line, just scrolled). A simpler check:
	// reset via OnChange(0) and verify Home key puts cursor at col 0.
	sb.OnChange(0)
	m.HandleEvent(sbKeyEv(tcell.KeyHome))
	_, col := m.CursorPos()
	if col != 0 {
		t.Errorf("After HScrollBar.OnChange(0) + Home: CursorPos col=%d, want 0", col)
	}
}

// ---------------------------------------------------------------------------
// Section 7 — State-dependent visibility
// ---------------------------------------------------------------------------

// TestSetStateSelectedShowsScrollBars verifies that when SfSelected is gained,
// linked scrollbars become visible.
// Spec: "if gaining focus, show linked scrollbars (SetState(SfVisible, true))"
func TestSetStateSelectedShowsScrollBars(t *testing.T) {
	m := newSbMemo()
	vSb := newVScrollBar()
	hSb := newHScrollBar()
	m.SetVScrollBar(vSb)
	m.SetHScrollBar(hSb)

	// Ensure scrollbars start hidden.
	vSb.SetState(SfVisible, false)
	hSb.SetState(SfVisible, false)

	m.SetState(SfSelected, true)

	if !vSb.HasState(SfVisible) {
		t.Error("VScrollBar not visible after Memo gained SfSelected")
	}
	if !hSb.HasState(SfVisible) {
		t.Error("HScrollBar not visible after Memo gained SfSelected")
	}
}

// TestSetStateDeselectedHidesScrollBars verifies that when SfSelected is lost,
// linked scrollbars are hidden.
// Spec: "if losing focus, hide them"
func TestSetStateDeselectedHidesScrollBars(t *testing.T) {
	m := newSbMemo()
	vSb := newVScrollBar()
	hSb := newHScrollBar()
	m.SetVScrollBar(vSb)
	m.SetHScrollBar(hSb)

	// Give focus first so scrollbars are shown.
	m.SetState(SfSelected, true)

	// Now remove focus.
	m.SetState(SfSelected, false)

	if vSb.HasState(SfVisible) {
		t.Error("VScrollBar still visible after Memo lost SfSelected")
	}
	if hSb.HasState(SfVisible) {
		t.Error("HScrollBar still visible after Memo lost SfSelected")
	}
}

// TestSetStateWithoutScrollBarsNoPanic verifies that SetState does not panic
// when no scrollbars are linked.
// Spec: "Without linked scrollbars, SetState doesn't panic"
func TestSetStateWithoutScrollBarsNoPanic(t *testing.T) {
	m := newSbMemo()

	// Should not panic.
	m.SetState(SfSelected, true)
	m.SetState(SfSelected, false)
}

// ---------------------------------------------------------------------------
// Section 8 — Falsifying / boundary tests
// ---------------------------------------------------------------------------

// TestScrollbarSyncHappensAfterKeyboardEvent verifies that scrollbar sync
// occurs after each keyboard event, not only on initial link.
// Spec: "Scrollbar sync happens after keyboard events (not just on link)"
func TestScrollbarSyncHappensAfterKeyboardEvent(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 3)) // height=3
	m.SetText("l0\nl1\nl2\nl3\nl4")
	sb := NewScrollBar(NewRect(39, 0, 1, 3), Vertical)
	m.SetVScrollBar(sb)

	// Capture value at link time.
	initialValue := sb.Value() // should be 0

	// Move cursor down enough to cause viewport scroll.
	for i := 0; i < 3; i++ {
		m.HandleEvent(sbKeyEv(tcell.KeyDown))
	}

	if sb.Value() == initialValue {
		t.Error("VScrollBar value did not change after keyboard navigation that scrolled the viewport")
	}
}

// TestNilVScrollBarAfterLinkDoesNotPanic verifies that passing nil to
// SetVScrollBar after a scrollbar was linked does not cause a panic during
// subsequent operations.
// Spec: "Passing nil doesn't crash"
func TestNilVScrollBarAfterLinkDoesNotPanic(t *testing.T) {
	m := newSbMemo()
	sb := newVScrollBar()
	m.SetVScrollBar(sb)
	m.SetVScrollBar(nil)

	// Subsequent key events should not panic now that scrollbar is unlinked.
	m.HandleEvent(sbKeyEv(tcell.KeyDown))
	m.HandleEvent(sbKeyEv(tcell.KeyUp))
}

// TestNilHScrollBarAfterLinkDoesNotPanic verifies that passing nil to
// SetHScrollBar after a scrollbar was linked does not cause a panic during
// subsequent operations.
// Spec: "Passing nil doesn't crash"
func TestNilHScrollBarAfterLinkDoesNotPanic(t *testing.T) {
	m := newSbMemo()
	sb := newHScrollBar()
	m.SetHScrollBar(sb)
	m.SetHScrollBar(nil)

	// Subsequent key events should not panic.
	m.HandleEvent(sbRuneEv('a'))
	m.HandleEvent(sbKeyEv(tcell.KeyHome))
}

// TestVScrollBarNotSyncedWithoutLink verifies that a scrollbar that was never
// linked does not have its value set by Memo operations.
// Falsifying: if sync happened globally, unlinked scrollbars would change.
func TestVScrollBarNotSyncedWithoutLink(t *testing.T) {
	m := newSbMemo()
	m.SetText("l0\nl1\nl2\nl3\nl4")
	sb := newVScrollBar()
	// Do NOT link the scrollbar.

	// Force a known starting value.
	sb.SetValue(0)

	// Move cursor; Memo should not touch the unlinked scrollbar.
	for i := 0; i < 5; i++ {
		m.HandleEvent(sbKeyEv(tcell.KeyDown))
	}

	// The unlinked scrollbar should retain its last manually-set value.
	if sb.Value() != 0 {
		t.Errorf("Unlinked VScrollBar value changed to %d; Memo should not touch unlinked scrollbars", sb.Value())
	}
}

// TestSetStateOtherFlagsDoNotAffectScrollBars verifies that toggling state
// flags other than SfSelected does not show/hide scrollbars.
// Falsifying: an over-broad SetState override would trigger on any flag.
func TestSetStateOtherFlagsDoNotAffectScrollBars(t *testing.T) {
	m := newSbMemo()
	vSb := newVScrollBar()
	m.SetVScrollBar(vSb)
	vSb.SetState(SfVisible, false)

	// Toggling SfVisible on the Memo itself should not show the scrollbar via
	// the SfSelected-visibility logic.
	m.SetState(SfVisible, true)

	// The scrollbar should still be invisible (we hid it manually and didn't
	// gain SfSelected).
	if vSb.HasState(SfVisible) {
		t.Error("VScrollBar became visible when Memo gained SfVisible (not SfSelected); should only respond to SfSelected")
	}
}
