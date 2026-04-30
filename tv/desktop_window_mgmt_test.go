package tv

// desktop_window_mgmt_test.go — tests for Task 7: Desktop Window Management.
// Written BEFORE any implementation exists; all tests are driven by the spec.
// Each test cites the exact spec sentence it verifies.
//
// Spec summary (Task 7):
//   - Desktop handles Alt+1..Alt+9: selects the window whose Number() matches
//     the digit. Clears the event on match.
//   - Desktop.SelectNextWindow(): moves focus to the next window in children
//     list (wraps), brings it to front. BringToFront moves to END of list, so
//     subsequent calls cycle in z-order (most-recently-used stack).
//   - Desktop.SelectPrevWindow(): moves focus to the previous window (wraps),
//     brings it to front. Same z-order cycling behaviour.
//   - Desktop handles CmNext → SelectNextWindow, clears event.
//   - Desktop handles CmPrev → SelectPrevWindow, clears event.
//   - Desktop.Tile(): arranges all VISIBLE windows in a non-overlapping grid.
//     cols = ceil(sqrt(n)), rows = ceil(n/cols).
//     Each window gets one cell; last column/row absorbs remaining space.
//     Minimum cell size: 10 wide, 5 tall.
//   - Desktop.Cascade(): arranges VISIBLE windows diagonally.
//     Each window: 3/4 of Desktop width × 3/4 of Desktop height (min 10×5).
//     Each window offset by (2, 1) from previous, wrapping if it would exceed
//     Desktop bounds.
//   - Desktop handles CmTile → Tile; CmCascade → Cascade.
//   - Window management keyboard events are handled BEFORE delegating to Group.

import (
	"math"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// altKeyEvent creates an EvKeyboard event for Alt+<digit rune>.
func altKeyEvent(digit rune) *Event {
	return &Event{
		What: EvKeyboard,
		Key: &KeyEvent{
			Key:       tcell.KeyRune,
			Rune:      digit,
			Modifiers: tcell.ModAlt,
		},
	}
}

// cmdEvent creates an EvCommand event.
func cmdEvent(cmd CommandCode) *Event {
	return &Event{What: EvCommand, Command: cmd}
}

// newNumberedWindow creates a Window with the given number. The bounds given
// must fit inside the Desktop; typical use is small non-overlapping rects.
func newNumberedWindow(bounds Rect, n int) *Window {
	return NewWindow(bounds, "W", WithWindowNumber(n))
}

// ---------------------------------------------------------------------------
// Alt+1..Alt+9 — window selection by number
// ---------------------------------------------------------------------------

// Spec: "Desktop handles Alt+1 through Alt+9 keyboard events: selects the
// window whose Number() matches the digit."
func TestAltNumberSelectsMatchingWindow(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := newNumberedWindow(NewRect(0, 0, 20, 10), 1)
	w2 := newNumberedWindow(NewRect(20, 0, 20, 10), 2)
	d.Insert(w1)
	d.Insert(w2)
	// w2 is focused (last inserted selectable)

	d.HandleEvent(altKeyEvent('1'))

	if d.FocusedChild() != w1 {
		t.Errorf("Alt+1: FocusedChild() = %v, want w1 (number 1)", d.FocusedChild())
	}
}

// Spec: same — falsifying: ensure it selects the window with the MATCHING
// number, not just any window.
func TestAltNumberSelectsWindowWithExactNumber(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := newNumberedWindow(NewRect(0, 0, 20, 10), 1)
	w3 := newNumberedWindow(NewRect(20, 0, 20, 10), 3)
	d.Insert(w1)
	d.Insert(w3)

	d.HandleEvent(altKeyEvent('3'))

	if d.FocusedChild() != w3 {
		t.Errorf("Alt+3: FocusedChild() = %v, want w3 (number 3)", d.FocusedChild())
	}
}

// Spec: "Clears the event on match."
func TestAltNumberClearsEventOnMatch(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w2 := newNumberedWindow(NewRect(0, 0, 20, 10), 2)
	d.Insert(w2)

	ev := altKeyEvent('2')
	d.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Alt+2 match: event was not cleared (IsCleared=%v)", ev.IsCleared())
	}
}

// Spec: contrapositive of "Clears the event on match" — when there is NO
// matching window, the event must NOT be cleared (so Group can still process it).
func TestAltNumberDoesNotClearEventOnNoMatch(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := newNumberedWindow(NewRect(0, 0, 20, 10), 1)
	d.Insert(w1)

	ev := altKeyEvent('9') // no window with number 9
	d.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("Alt+9 with no match: event was cleared, want it left unconsumed")
	}
}

// Spec: Alt+N range is 1..9. Alt+0 should NOT trigger window selection.
func TestAltZeroDoesNotSelectWindow(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	// We use a window with number 0; if Alt+0 matched it that would be wrong.
	w0 := newNumberedWindow(NewRect(0, 0, 20, 10), 0)
	d.Insert(w0)

	ev := altKeyEvent('0')
	d.HandleEvent(ev)

	// Event must not be cleared (no window-number selection).
	if ev.IsCleared() {
		t.Errorf("Alt+0 should not trigger window number selection (event was cleared)")
	}
}

// Spec: "Desktop handles Alt+1 through Alt+9" — the window must actually
// receive SfSelected after the Alt+N dispatch.
func TestAltNumberSetsSelectedOnMatchedWindow(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := newNumberedWindow(NewRect(0, 0, 20, 10), 1)
	w2 := newNumberedWindow(NewRect(20, 0, 20, 10), 2)
	d.Insert(w1)
	d.Insert(w2)

	d.HandleEvent(altKeyEvent('1'))

	if !w1.HasState(SfSelected) {
		t.Errorf("Alt+1: w1 should have SfSelected after selection")
	}
}

// Spec: "selects the window whose Number() matches the digit" — Alt+N with
// empty Desktop must not panic.
func TestAltNumberEmptyDesktopDoesNotPanic(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	// Must not panic.
	d.HandleEvent(altKeyEvent('5'))
}

// Spec: "window management keyboard events are handled BEFORE delegating to Group."
// Verify: a focused child does NOT receive the Alt+N event when it was handled
// (i.e. a matching window exists). We use a phaseTestView as the focused child
// so we can count HandleEvent calls.
func TestAltNumberHandledBeforeDelegatingToGroup(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	// Window with number 1.
	w1 := newNumberedWindow(NewRect(0, 0, 20, 10), 1)
	d.Insert(w1)
	// A selectable, post-process child to observe whether the event leaks.
	observer := newPhaseView("observer", NewRect(30, 0, 10, 5))
	observer.SetOptions(OfPostProcess, true)
	d.Insert(observer)
	// Make observer focused so it would receive the event in normal dispatch.
	d.SetFocusedChild(observer)

	ev := altKeyEvent('1')
	d.HandleEvent(ev)

	// Event must be cleared (handled before group dispatch).
	if !ev.IsCleared() {
		t.Errorf("Alt+1 match: event not cleared, Group may have processed it")
	}
	// Observer must not have received the Alt+1 event.
	if observer.lastEvent == ev {
		t.Errorf("Alt+1: event reached Group (postprocess observer received it)")
	}
}

// ---------------------------------------------------------------------------
// SelectNextWindow
// ---------------------------------------------------------------------------

// Spec: "Desktop.SelectNextWindow() moves focus to the next window in the
// children list (wraps around)."
func TestSelectNextWindowMovesFocusToNextChild(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	d.Insert(w1)
	d.Insert(w2)
	// After inserts: children=[w1,w2], w2 focused (last inserted).
	// BringToFront(w2) already happened on insert, so BringToFront effect is already there.
	// SelectNextWindow from w2 should wrap to w1 (the only other child, which is earlier in list).
	// After insert: children=[w1,w2]. Focused=w2. Next after w2 wraps to w1.

	d.SelectNextWindow()

	if d.FocusedChild() != w1 {
		t.Errorf("SelectNextWindow: FocusedChild() = %v, want w1", d.FocusedChild())
	}
}

// Spec: "Brings the newly focused window to front."
func TestSelectNextWindowBringsWindowToFront(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	d.Insert(w1)
	d.Insert(w2)
	// children=[w1,w2]; w2 focused.

	d.SelectNextWindow()
	// SelectNextWindow moves to w1 and brings it to front.
	// After BringToFront(w1): children=[w2,w1].

	children := d.Children()
	if children[len(children)-1] != w1 {
		t.Errorf("SelectNextWindow: newly focused window should be last in children (frontmost), got %v", children[len(children)-1])
	}
}

// Spec: "wraps around" — when the focused window is the last in the list
// (frontmost), SelectNextWindow wraps to the first.
func TestSelectNextWindowWrapsAroundFromFront(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	w3 := NewWindow(NewRect(40, 0, 20, 10), "W3")
	d.Insert(w1)
	d.Insert(w2)
	d.Insert(w3)
	// children=[w1,w2,w3]; w3 focused (last in list).
	// SelectNextWindow from w3 (last) should wrap to first (w1).

	d.SelectNextWindow()

	if d.FocusedChild() != w1 {
		t.Errorf("SelectNextWindow wrap: FocusedChild() = %v, want w1 (first child)", d.FocusedChild())
	}
}

// Spec: "Note: BringToFront moves the window to the end of the children list,
// so subsequent SelectNextWindow calls cycle in z-order, not insertion order."
// After SelectNextWindow the next call should cycle by z-order.
func TestSelectNextWindowCyclesInZOrder(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	w3 := NewWindow(NewRect(40, 0, 20, 10), "W3")
	d.Insert(w1)
	d.Insert(w2)
	d.Insert(w3)
	// children=[w1,w2,w3], w3 focused.

	// First call: wraps to w1, then BringToFront(w1) → children=[w2,w3,w1].
	d.SelectNextWindow()
	if d.FocusedChild() != w1 {
		t.Fatalf("first SelectNextWindow: want w1, got %v", d.FocusedChild())
	}

	// Second call from w1 (last in list=[w2,w3,w1]): next from w1 wraps to w2.
	// BringToFront(w2) → children=[w3,w1,w2].
	d.SelectNextWindow()
	if d.FocusedChild() != w2 {
		t.Errorf("second SelectNextWindow (z-order cycle): FocusedChild() = %v, want w2", d.FocusedChild())
	}
}

// Spec: SelectNextWindow with no children must not panic.
func TestSelectNextWindowNoChildrenDoesNotPanic(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	// Must not panic.
	d.SelectNextWindow()
}

// Spec: SelectNextWindow with one window keeps that window focused.
func TestSelectNextWindowOneWindowStaysFocused(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	d.Insert(w1)

	d.SelectNextWindow()

	if d.FocusedChild() != w1 {
		t.Errorf("SelectNextWindow with 1 window: FocusedChild() = %v, want w1", d.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// SelectPrevWindow
// ---------------------------------------------------------------------------

// Spec: "Desktop.SelectPrevWindow() moves focus to the previous window
// (wraps around)."
func TestSelectPrevWindowMovesFocusToPrevChild(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	w3 := NewWindow(NewRect(40, 0, 20, 10), "W3")
	d.Insert(w1)
	d.Insert(w2)
	d.Insert(w3)
	// children=[w1,w2,w3], w3 focused.
	// SelectPrevWindow from w3 should move to w2 (the one before w3 in the list).

	d.SelectPrevWindow()

	if d.FocusedChild() != w2 {
		t.Errorf("SelectPrevWindow: FocusedChild() = %v, want w2", d.FocusedChild())
	}
}

// Spec: "Brings it to front."
func TestSelectPrevWindowBringsWindowToFront(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	w3 := NewWindow(NewRect(40, 0, 20, 10), "W3")
	d.Insert(w1)
	d.Insert(w2)
	d.Insert(w3)

	d.SelectPrevWindow()
	// Should select w2 and bring it to front.

	children := d.Children()
	if children[len(children)-1] != w2 {
		t.Errorf("SelectPrevWindow: newly focused window should be last (frontmost) in children, got %v", children[len(children)-1])
	}
}

// Spec: "wraps around" — when the focused window is the first in the list,
// SelectPrevWindow wraps to the last.
func TestSelectPrevWindowWrapsAroundFromFirst(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	w3 := NewWindow(NewRect(40, 0, 20, 10), "W3")
	d.Insert(w1)
	d.Insert(w2)
	d.Insert(w3)
	// children=[w1,w2,w3], w3 focused.
	// Manually bring w1 to focus via BringToFront to make it the focused window,
	// then verify that SelectPrevWindow wraps to the last position in the new order.
	// After BringToFront(w1): children=[w2,w3,w1], w1 is focused and LAST.
	d.BringToFront(w1)
	// children=[w2,w3,w1]. w1 is focused and at index 2 (last).
	// Previous of w1 (index 2) is w3 (index 1).
	// Alternatively if "previous" means wrapping backward from first: we need w1
	// to be at the front. Let's check: BringToFront puts w1 at the END.
	// So children=[w2,w3,w1] and w1 is LAST. SelectPrevWindow from the last item
	// goes backward to w3 — not a wrap.
	//
	// To test the actual wrap (from first/index-0 to last), we need the focused
	// window to be at index 0 (the beginning of the list). Let's use a simpler
	// setup: 2 windows, manually focus the one that is NOT last.

	// Reset with a cleaner setup.
	d2 := NewDesktop(NewRect(0, 0, 80, 25))
	a := NewWindow(NewRect(0, 0, 20, 10), "A")
	b := NewWindow(NewRect(20, 0, 20, 10), "B")
	d2.Insert(a)
	d2.Insert(b)
	// children=[a,b], b focused (last).
	// BringToFront(a) → children=[b,a], a focused.
	d2.BringToFront(a)
	// children=[b,a], a at index 1 (last). a is focused.
	// SelectPrevWindow from a (index 1) goes to b (index 0).
	// That is not a wrap scenario yet.

	// For a true wrap, the focused window must be at index 0.
	// Use SetFocusedChild to put b in focus without changing order.
	// children still [b,a]; now focus b which is index 0.
	d2.SetFocusedChild(b)
	// Now SelectPrevWindow from b (index 0) should wrap to a (last, index 1).
	d2.SelectPrevWindow()

	if d2.FocusedChild() != a {
		t.Errorf("SelectPrevWindow wrap from first: FocusedChild() = %v, want a (last child)", d2.FocusedChild())
	}
}

// Spec: "Same z-order cycling behavior as SelectNextWindow."
func TestSelectPrevWindowCyclesInZOrder(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	w3 := NewWindow(NewRect(40, 0, 20, 10), "W3")
	d.Insert(w1)
	d.Insert(w2)
	d.Insert(w3)
	// children=[w1,w2,w3], w3 focused.

	// First call: SelectPrevWindow from w3 → w2. BringToFront(w2) → children=[w1,w3,w2].
	d.SelectPrevWindow()
	if d.FocusedChild() != w2 {
		t.Fatalf("first SelectPrevWindow: want w2, got %v", d.FocusedChild())
	}

	// After BringToFront(w2): children=[w1,w3,w2]. w2 at index 2.
	// Second SelectPrevWindow from w2 (index 2) → w3 (index 1).
	// BringToFront(w3) → children=[w1,w2,w3].
	d.SelectPrevWindow()
	if d.FocusedChild() != w3 {
		t.Errorf("second SelectPrevWindow (z-order cycle): FocusedChild() = %v, want w3", d.FocusedChild())
	}
}

// Spec: SelectPrevWindow with no children must not panic.
func TestSelectPrevWindowNoChildrenDoesNotPanic(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	// Must not panic.
	d.SelectPrevWindow()
}

// Spec: SelectPrevWindow with one window keeps that window focused.
func TestSelectPrevWindowOneWindowStaysFocused(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	d.Insert(w1)

	d.SelectPrevWindow()

	if d.FocusedChild() != w1 {
		t.Errorf("SelectPrevWindow with 1 window: FocusedChild() = %v, want w1", d.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// CmNext command
// ---------------------------------------------------------------------------

// Spec: "Desktop handles CmNext command by calling SelectNextWindow() and
// clearing the event."
func TestCmNextCallsSelectNextWindow(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	d.Insert(w1)
	d.Insert(w2)
	// w2 focused. CmNext should move to w1.

	d.HandleEvent(cmdEvent(CmNext))

	if d.FocusedChild() != w1 {
		t.Errorf("CmNext: FocusedChild() = %v, want w1 (SelectNextWindow result)", d.FocusedChild())
	}
}

// Spec: "clearing the event" after CmNext.
func TestCmNextClearsEvent(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))

	ev := cmdEvent(CmNext)
	d.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmNext: event was not cleared")
	}
}

// Spec: CmNext clears event even when there are no windows (no-op, but still clears).
func TestCmNextClearsEventWithNoWindows(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))

	ev := cmdEvent(CmNext)
	d.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmNext with no windows: event was not cleared")
	}
}

// ---------------------------------------------------------------------------
// CmPrev command
// ---------------------------------------------------------------------------

// Spec: "Desktop handles CmPrev command by calling SelectPrevWindow() and
// clearing the event."
func TestCmPrevCallsSelectPrevWindow(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	w1 := NewWindow(NewRect(0, 0, 20, 10), "W1")
	w2 := NewWindow(NewRect(20, 0, 20, 10), "W2")
	w3 := NewWindow(NewRect(40, 0, 20, 10), "W3")
	d.Insert(w1)
	d.Insert(w2)
	d.Insert(w3)
	// w3 focused. CmPrev should move to w2.

	d.HandleEvent(cmdEvent(CmPrev))

	if d.FocusedChild() != w2 {
		t.Errorf("CmPrev: FocusedChild() = %v, want w2 (SelectPrevWindow result)", d.FocusedChild())
	}
}

// Spec: "clearing the event" after CmPrev.
func TestCmPrevClearsEvent(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))

	ev := cmdEvent(CmPrev)
	d.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmPrev: event was not cleared")
	}
}

// Spec: CmPrev clears event even when there are no windows.
func TestCmPrevClearsEventWithNoWindows(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))

	ev := cmdEvent(CmPrev)
	d.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmPrev with no windows: event was not cleared")
	}
}

// ---------------------------------------------------------------------------
// Tile
// ---------------------------------------------------------------------------

// gridCols returns ceil(sqrt(n)) as the spec defines.
func gridCols(n int) int {
	return int(math.Ceil(math.Sqrt(float64(n))))
}

// gridRows returns ceil(n/cols).
func gridRows(n, cols int) int {
	return (n + cols - 1) / cols
}

// Spec: "Desktop.Tile() arranges all visible windows in a non-overlapping grid."
// With 1 window it should fill the Desktop.
func TestTileOneWindowFillsDesktop(t *testing.T) {
	dw, dh := 80, 25
	d := NewDesktop(NewRect(0, 0, dw, dh))
	w := NewWindow(NewRect(5, 5, 10, 5), "W")
	d.Insert(w)

	d.Tile()

	b := w.Bounds()
	if b.A.X != 0 || b.A.Y != 0 || b.Width() != dw || b.Height() != dh {
		t.Errorf("Tile 1 window: bounds = %v, want (0,0,%d,%d)", b, dw, dh)
	}
}

// Spec: "Computes a grid of cols×rows cells where cols = ceil(sqrt(n)) and
// rows = ceil(n/cols)."
// With 4 windows: cols=2, rows=2. Each cell is 40×12 (or similar).
func TestTileFourWindowsGridDimensions(t *testing.T) {
	dw, dh := 80, 24
	d := NewDesktop(NewRect(0, 0, dw, dh))
	windows := make([]*Window, 4)
	for i := range windows {
		windows[i] = NewWindow(NewRect(0, 0, 10, 5), "W")
		d.Insert(windows[i])
	}

	d.Tile()

	// cols=ceil(sqrt(4))=2, rows=ceil(4/2)=2.
	cols := gridCols(4)
	rows := gridRows(4, cols)
	cellW := dw / cols
	cellH := dh / rows

	// cols=2, rows=2, cellW=40, cellH=12.
	// Expected grid positions:
	//   w[0]: (0,  0,  40, 12)
	//   w[1]: (40, 0,  40, 12)
	//   w[2]: (0,  12, 40, 12)
	//   w[3]: (40, 12, 40, 12)
	type wantPos struct{ x, y, w, h int }
	want := []wantPos{
		{0, 0, cellW, cellH},
		{cellW, 0, cellW, cellH},
		{0, cellH, cellW, cellH},
		{cellW, cellH, cellW, cellH},
	}
	for i, w := range windows {
		b := w.Bounds()
		wp := want[i]
		if b.A.X != wp.x || b.A.Y != wp.y {
			t.Errorf("Tile 4 windows: w[%d] origin = (%d,%d), want (%d,%d)", i, b.A.X, b.A.Y, wp.x, wp.y)
		}
		if b.Width() != wp.w || b.Height() != wp.h {
			t.Errorf("Tile 4 windows: w[%d] size = (%d,%d), want (%d,%d)", i, b.Width(), b.Height(), wp.w, wp.h)
		}
	}
	_ = rows
}

// Spec: "Each window gets one cell; the last column/row absorbs remaining space."
// With 2 windows: cols=ceil(sqrt(2))=2, rows=ceil(2/2)=1.
// Cell width = dw/2; last column may get remaining space.
func TestTileTwoWindowsLastColumnAbsorbsRemainder(t *testing.T) {
	dw, dh := 81, 25 // odd width so last column gets remainder
	d := NewDesktop(NewRect(0, 0, dw, dh))
	w1 := NewWindow(NewRect(0, 0, 10, 5), "W1")
	w2 := NewWindow(NewRect(0, 0, 10, 5), "W2")
	d.Insert(w1)
	d.Insert(w2)

	d.Tile()

	// cols=2, rows=1. cellW = 81/2 = 40.
	// w1 at col=0: x=0, width=40.
	// w2 at col=1 (last): x=40, width=81-40=41 (last column absorbs).
	cols := gridCols(2)
	cellW := dw / cols

	// Verify the two windows together span the full desktop width.
	totalWidth := 0
	for _, w := range []*Window{w1, w2} {
		b := w.Bounds()
		// Each window's right edge must not exceed dw.
		if b.B.X > dw {
			t.Errorf("Tile: window %v right edge %d exceeds desktop width %d", b, b.B.X, dw)
		}
		// All windows must start from a valid column offset.
		if b.A.X < 0 {
			t.Errorf("Tile: window left edge %d < 0", b.A.X)
		}
		totalWidth += b.Width()
	}
	if totalWidth != dw {
		t.Errorf("Tile 2 windows: total width = %d, want %d (last column absorbs remainder)", totalWidth, dw)
	}
	_ = cellW
}

// Spec: "Each window gets one cell; the last column/row absorbs remaining space."
// With 3 windows: cols=ceil(sqrt(3))=2, rows=ceil(3/2)=2.
// Row 0 has 2 windows; row 1 has 1 window. The single window in row 1 spans
// the full height remainder (last row absorbs remaining height).
func TestTileThreeWindowsLastRowAbsorbsRemainder(t *testing.T) {
	dw, dh := 80, 25
	d := NewDesktop(NewRect(0, 0, dw, dh))
	w1 := NewWindow(NewRect(0, 0, 10, 5), "W1")
	w2 := NewWindow(NewRect(0, 0, 10, 5), "W2")
	w3 := NewWindow(NewRect(0, 0, 10, 5), "W3")
	d.Insert(w1)
	d.Insert(w2)
	d.Insert(w3)

	d.Tile()

	// cols=2, rows=2. cellH = 25/2 = 12.
	// row 0 windows: height=12. row 1 window (w3): height=25-12=13.
	cols := gridCols(3)
	rows := gridRows(3, cols)
	cellH := dh / rows

	// Verify all windows together span the full desktop height in their column.
	// The last row window must reach dh.
	allBottoms := map[int]int{}
	for _, w := range []*Window{w1, w2, w3} {
		b := w.Bounds()
		col := b.A.X / (dw / cols)
		if b.B.Y > allBottoms[col] {
			allBottoms[col] = b.B.Y
		}
	}
	for col, bottom := range allBottoms {
		if bottom != dh {
			t.Errorf("Tile 3 windows: column %d bottom = %d, want %d (last row absorbs remainder)", col, bottom, dh)
		}
	}
	_ = cellH
	_ = cols
}

// Spec: "Tile() arranges all VISIBLE windows."
// Invisible windows must not be repositioned.
func TestTileSkipsInvisibleWindows(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	visible := NewWindow(NewRect(0, 0, 20, 10), "Visible")
	d.Insert(visible)
	invisible := NewWindow(NewRect(30, 5, 10, 5), "Invisible")
	invisible.SetState(SfVisible, false)
	d.Insert(invisible)

	originalBounds := invisible.Bounds()

	d.Tile()

	if invisible.Bounds() != originalBounds {
		t.Errorf("Tile: invisible window bounds changed from %v to %v", originalBounds, invisible.Bounds())
	}
}

// Spec: "Tile() with zero windows — no-op."
func TestTileZeroWindowsIsNoOp(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	// Must not panic, no children to arrange.
	d.Tile()
}

// Spec: "Minimum window size per cell: 10 wide, 5 tall."
// With a small desktop and many windows, each window must be at least 10×5.
func TestTileMinimumCellSize(t *testing.T) {
	// 5 windows, desktop 50×25. cols=ceil(sqrt(5))=3, rows=ceil(5/3)=2.
	// cellW=50/3=16, cellH=25/2=12. Both >= 10 wide, 5 tall — these pass.
	// To force the minimum, use a desktop that would produce cells < 10 wide.
	// 9 windows on a 27×9 desktop: cols=3, rows=3, cellW=9 < 10.
	dw, dh := 27, 9
	d := NewDesktop(NewRect(0, 0, dw, dh))
	for i := 0; i < 9; i++ {
		d.Insert(NewWindow(NewRect(0, 0, 10, 5), "W"))
	}

	d.Tile()

	for _, child := range d.Children() {
		b := child.Bounds()
		if b.Width() < 10 {
			t.Errorf("Tile: window width = %d, want at least 10", b.Width())
		}
		if b.Height() < 5 {
			t.Errorf("Tile: window height = %d, want at least 5", b.Height())
		}
	}
}


// ---------------------------------------------------------------------------
// Cascade
// ---------------------------------------------------------------------------

// Spec: "Desktop.Cascade() arranges visible windows in a diagonal stack.
// Each window is 3/4 of Desktop width and 3/4 of Desktop height (minimum 10×5)."
func TestCascadeWindowSize(t *testing.T) {
	dw, dh := 80, 24
	d := NewDesktop(NewRect(0, 0, dw, dh))
	w := NewWindow(NewRect(0, 0, 10, 5), "W")
	d.Insert(w)

	d.Cascade()

	wantW := max(dw*3/4, 10)
	wantH := max(dh*3/4, 10) // spec says min 10×5
	// The spec minimum is 10 wide, 5 tall.
	wantH = max(dh*3/4, 5)
	wantW = max(dw*3/4, 10)

	b := w.Bounds()
	if b.Width() != wantW {
		t.Errorf("Cascade: window width = %d, want %d (3/4 of %d)", b.Width(), wantW, dw)
	}
	if b.Height() != wantH {
		t.Errorf("Cascade: window height = %d, want %d (3/4 of %d)", b.Height(), wantH, dh)
	}
}

// Spec: "Windows offset by (2, 1) from the previous."
func TestCascadeWindowsOffsetBy2And1(t *testing.T) {
	dw, dh := 80, 24
	d := NewDesktop(NewRect(0, 0, dw, dh))
	w1 := NewWindow(NewRect(0, 0, 10, 5), "W1")
	w2 := NewWindow(NewRect(0, 0, 10, 5), "W2")
	d.Insert(w1)
	d.Insert(w2)

	d.Cascade()

	b1, b2 := w1.Bounds(), w2.Bounds()
	dx := b2.A.X - b1.A.X
	dy := b2.A.Y - b1.A.Y
	if dx != 2 || dy != 1 {
		t.Errorf("Cascade: offset between w1 and w2 = (%d,%d), want (2,1)", dx, dy)
	}
}

// Spec: "Each window offset by (2, 1) from the previous" — the first window has
// no previous, so it starts at (0, 0) with no offset applied.
func TestCascadeFirstWindowAtOrigin(t *testing.T) {
	dw, dh := 80, 24
	d := NewDesktop(NewRect(0, 0, dw, dh))
	w1 := NewWindow(NewRect(5, 5, 10, 5), "W1")
	w2 := NewWindow(NewRect(5, 5, 10, 5), "W2")
	d.Insert(w1)
	d.Insert(w2)

	d.Cascade()

	b := w1.Bounds()
	if b.A.X != 0 || b.A.Y != 0 {
		t.Errorf("Cascade: first window origin = (%d,%d), want (0,0)", b.A.X, b.A.Y)
	}
}

// Spec: "wrapping if they'd exceed Desktop bounds."
// When the offset would push a window beyond the desktop, it wraps back.
func TestCascadeWrapsWhenExceedingDesktopBounds(t *testing.T) {
	// Use a desktop width of 40, height of 20.
	// Window size: 3/4 * 40 = 30, 3/4 * 20 = 15.
	// Offset (2,1) per window.
	// First window at (0,0). Right edge = 30.
	// Second at (2,1). Right edge = 32.
	// At some point (2*n, n) + 30 > 40 → wrap.
	// After (40-30)/2 = 5 offsets right, x=10, right=40 → next x=12+30>40 → wrap.
	dw, dh := 40, 20
	d := NewDesktop(NewRect(0, 0, dw, dh))
	// Insert enough windows to force wrapping: 8 should be more than enough.
	for i := 0; i < 8; i++ {
		d.Insert(NewWindow(NewRect(0, 0, 10, 5), "W"))
	}

	d.Cascade()

	winW := max(dw*3/4, 10)
	winH := max(dh*3/4, 5)

	visible := make([]View, 0, 8)
	for _, child := range d.Children() {
		if child.HasState(SfVisible) {
			visible = append(visible, child)
		}
	}

	for _, child := range visible {
		b := child.Bounds()
		// No window should exceed the desktop dimensions.
		if b.B.X > dw {
			t.Errorf("Cascade: window right edge %d exceeds desktop width %d", b.B.X, dw)
		}
		if b.B.Y > dh {
			t.Errorf("Cascade: window bottom edge %d exceeds desktop height %d", b.B.Y, dh)
		}
	}

	// Verify that wrapping actually occurred: at least one window must have an
	// x-origin LESS THAN the preceding window's x-origin (i.e. it wrapped back
	// toward the left rather than continuing the rightward diagonal).
	wrappedX := false
	for i := 1; i < len(visible); i++ {
		prev := visible[i-1].Bounds().A.X
		cur := visible[i].Bounds().A.X
		if cur < prev {
			wrappedX = true
			break
		}
	}
	if !wrappedX {
		t.Errorf("Cascade: no wrapping detected — expected at least one window to have x < preceding window's x")
	}

	_ = winW
	_ = winH
}

// Spec: "Cascade() arranges all VISIBLE windows." Invisible windows unchanged.
func TestCascadeSkipsInvisibleWindows(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	visible := NewWindow(NewRect(0, 0, 20, 10), "Visible")
	d.Insert(visible)
	invisible := NewWindow(NewRect(30, 5, 10, 5), "Invisible")
	invisible.SetState(SfVisible, false)
	d.Insert(invisible)

	originalBounds := invisible.Bounds()

	d.Cascade()

	if invisible.Bounds() != originalBounds {
		t.Errorf("Cascade: invisible window bounds changed from %v to %v", originalBounds, invisible.Bounds())
	}
}

// Spec: "Cascade() with zero windows — no-op."
func TestCascadeZeroWindowsIsNoOp(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	// Must not panic.
	d.Cascade()
}

// Spec: "Each window is 3/4 of Desktop width and 3/4 of Desktop height
// (minimum 10×5)." Verify minimum is enforced.
func TestCascadeMinimumWindowSize(t *testing.T) {
	// Desktop so small that 3/4 of width < 10 and 3/4 of height < 5.
	dw, dh := 8, 4
	d := NewDesktop(NewRect(0, 0, dw, dh))
	w := NewWindow(NewRect(0, 0, 10, 5), "W")
	d.Insert(w)

	d.Cascade()

	b := w.Bounds()
	if b.Width() < 10 {
		t.Errorf("Cascade: window width = %d, want at least 10 (minimum)", b.Width())
	}
	if b.Height() < 5 {
		t.Errorf("Cascade: window height = %d, want at least 5 (minimum)", b.Height())
	}
}


// ---------------------------------------------------------------------------
// "Handled BEFORE delegating to Group" — window-mgmt events take priority
// ---------------------------------------------------------------------------

// Spec: "All window management keyboard events are handled BEFORE delegating
// to Group." Verify CmNext is handled by Desktop before reaching children.
func TestCmNextHandledBeforeGroup(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	// A postprocess observer that would receive the event if it leaked.
	observer := newPhaseView("observer", NewRect(0, 0, 10, 5))
	observer.SetOptions(OfPostProcess, true)
	d.Insert(observer)

	ev := cmdEvent(CmNext)
	d.HandleEvent(ev)

	// Event must be cleared before Group dispatch.
	if !ev.IsCleared() {
		t.Errorf("CmNext: event not cleared, Group may have processed it")
	}
	// Observer must not have received the CmNext event.
	if observer.lastEvent == ev {
		t.Errorf("CmNext: event reached Group (postprocess observer received it)")
	}
}

// Spec: same — CmPrev handled before Group.
func TestCmPrevHandledBeforeGroup(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	observer := newPhaseView("observer", NewRect(0, 0, 10, 5))
	observer.SetOptions(OfPostProcess, true)
	d.Insert(observer)

	ev := cmdEvent(CmPrev)
	d.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("CmPrev: event not cleared, Group may have processed it")
	}
	if observer.lastEvent == ev {
		t.Errorf("CmPrev: event reached Group (postprocess observer received it)")
	}
}


// Spec: Non-window-management events must still be delegated to Group after
// Desktop's own handling (falsifying the "only window-mgmt events" constraint).
func TestNonWindowMgmtEventsDelegateToGroup(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	observer := newSelectablePhaseView("observer", NewRect(0, 0, 10, 5))
	d.Insert(observer)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'z'}}
	d.HandleEvent(ev)

	// The observer (focused child) must have received the non-window-mgmt key.
	if observer.lastEvent != ev {
		t.Errorf("non-window-mgmt key: focused child did not receive event (event = %v)", observer.lastEvent)
	}
}
