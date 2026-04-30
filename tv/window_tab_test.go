package tv

// window_tab_test.go — tests for Task 6 of Phase 14:
// Moving Tab/Shift+Tab handling from Group to Window.
//
// After the move:
//   - Window.HandleEvent intercepts Tab/Shift+Tab before group dispatch.
//   - Group.HandleEvent no longer handles Tab or Shift+Tab.
//   - Desktop.HandleEvent no longer forwards Tab to the focused window.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// windowTabBacktabEvent returns a Shift+Tab keyboard event.
// tabEvent() is declared in group_focus_test.go (same package); we reuse it.
func windowTabBacktabEvent() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyBacktab}}
}

// ---------------------------------------------------------------------------
// 1. Tab in Window moves focus to next child widget
// ---------------------------------------------------------------------------

// TestWindowTabMovesFocusToNext verifies that sending a Tab event to a Window
// advances focus from the first child to the second child.
// Spec: "Window.HandleEvent handles Tab (KeyTab, no modifiers): calls w.group.FocusNext()"
func TestWindowTabMovesFocusToNext(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	first := newSelectableMockView(NewRect(0, 0, 5, 3))
	second := newSelectableMockView(NewRect(5, 0, 5, 3))
	w.Insert(first)
	w.Insert(second)
	// After two inserts, second is focused. Move focus back to first so Tab can advance.
	w.SetFocusedChild(first)

	w.HandleEvent(tabEvent())

	if w.FocusedChild() != second {
		t.Errorf("Tab: FocusedChild() = %v, want second child", w.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// 2. Tab event is consumed (cleared)
// ---------------------------------------------------------------------------

// TestWindowTabConsumesEvent verifies that Tab is cleared after Window handles it.
// Spec: "Window.HandleEvent handles Tab ... consumes event"
func TestWindowTabConsumesEvent(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	first := newSelectableMockView(NewRect(0, 0, 5, 3))
	second := newSelectableMockView(NewRect(5, 0, 5, 3))
	w.Insert(first)
	w.Insert(second)

	ev := tabEvent()
	w.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Tab: event not consumed (IsCleared() = false)")
	}
}

// ---------------------------------------------------------------------------
// 3. Shift+Tab in Window moves focus to previous child widget
// ---------------------------------------------------------------------------

// TestWindowShiftTabMovesFocusToPrev verifies that sending a Shift+Tab event
// to a Window moves focus to the previous child.
// Spec: "Window.HandleEvent handles Shift+Tab (KeyBacktab): calls w.group.FocusPrev()"
func TestWindowShiftTabMovesFocusToPrev(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	first := newSelectableMockView(NewRect(0, 0, 5, 3))
	second := newSelectableMockView(NewRect(5, 0, 5, 3))
	w.Insert(first)
	w.Insert(second)
	// second is focused after two inserts; Shift+Tab should wrap back to first.

	w.HandleEvent(windowTabBacktabEvent())

	if w.FocusedChild() != first {
		t.Errorf("Shift+Tab: FocusedChild() = %v, want first child", w.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// 4. Shift+Tab event is consumed (cleared)
// ---------------------------------------------------------------------------

// TestWindowShiftTabConsumesEvent verifies that Shift+Tab is cleared after
// Window handles it.
// Spec: "Window.HandleEvent handles Shift+Tab (KeyBacktab) ... consumes event"
func TestWindowShiftTabConsumesEvent(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	first := newSelectableMockView(NewRect(0, 0, 5, 3))
	second := newSelectableMockView(NewRect(5, 0, 5, 3))
	w.Insert(first)
	w.Insert(second)

	ev := windowTabBacktabEvent()
	w.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Shift+Tab: event not consumed (IsCleared() = false)")
	}
}

// ---------------------------------------------------------------------------
// 5. Tab wraps around (from last child back to first)
// ---------------------------------------------------------------------------

// TestWindowTabWrapsAroundToFirst verifies that Tab cycles focus from the last
// selectable child back to the first.
// Spec: "focusNext() wraps around to the beginning of the child list"
func TestWindowTabWrapsAroundToFirst(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	first := newSelectableMockView(NewRect(0, 0, 5, 3))
	second := newSelectableMockView(NewRect(5, 0, 5, 3))
	w.Insert(first)
	w.Insert(second)
	// second is focused (last inserted). Tab should wrap to first.

	w.HandleEvent(tabEvent())

	if w.FocusedChild() != first {
		t.Errorf("Tab wrap: FocusedChild() = %v, want first child (wrap-around)", w.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// 6. Group does NOT handle Tab anymore
// ---------------------------------------------------------------------------

// TestGroupDoesNotHandleTab verifies that a standalone Group (not inside a Window)
// does NOT consume a Tab event.
// Spec: "Group.HandleEvent removes Tab/Shift+Tab handling entirely"
func TestGroupDoesNotHandleTab(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	first := newSelectableMockView(NewRect(0, 0, 5, 3))
	second := newSelectableMockView(NewRect(5, 0, 5, 3))
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	ev := tabEvent()
	g.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("Group.HandleEvent consumed Tab — it should NOT handle Tab after the move")
	}
}

// ---------------------------------------------------------------------------
// 7. Group does NOT handle Shift+Tab anymore
// ---------------------------------------------------------------------------

// TestGroupDoesNotHandleShiftTab verifies that a standalone Group does NOT
// consume a Shift+Tab event.
// Spec: "Group.HandleEvent removes Tab/Shift+Tab handling entirely"
func TestGroupDoesNotHandleShiftTab(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	first := newSelectableMockView(NewRect(0, 0, 5, 3))
	second := newSelectableMockView(NewRect(5, 0, 5, 3))
	g.Insert(first)
	g.Insert(second)

	ev := windowTabBacktabEvent()
	g.HandleEvent(ev)

	if ev.IsCleared() {
		t.Errorf("Group.HandleEvent consumed Shift+Tab — it should NOT handle Shift+Tab after the move")
	}
}

// ---------------------------------------------------------------------------
// 8. Desktop does NOT forward Tab to focused window
// ---------------------------------------------------------------------------

// TestDesktopTabReachesWindowThroughGroupDispatch verifies that after removing
// Desktop's explicit Tab forwarding, Tab still reaches the focused Window
// through normal group dispatch and is handled there.
// Spec: "Desktop.HandleEvent removes the Tab/Shift+Tab forwarding to focused window"
func TestDesktopTabReachesWindowThroughGroupDispatch(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))
	win := NewWindow(NewRect(0, 0, 40, 20), "Win")
	child1 := newSelectableMockView(NewRect(0, 0, 5, 3))
	child2 := newSelectableMockView(NewRect(5, 0, 5, 3))
	win.Insert(child1)
	win.Insert(child2)
	win.SetFocusedChild(child1)
	d.Insert(win)

	ev := tabEvent()
	d.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Tab sent to Desktop was not handled — Tab should reach Window through group dispatch")
	}
	if win.FocusedChild() != child2 {
		t.Errorf("Tab via Desktop: FocusedChild() = %v, want child2 (Tab should advance focus in Window)", win.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// 9. Falsification: Tab actually changes FocusedChild, not just consuming event
// ---------------------------------------------------------------------------

// TestWindowTabActuallyChangesFocus is a falsifying test verifying that Tab
// does more than consume the event — it genuinely changes which child is focused.
// Without this test a stub that only clears the event would pass tests 1-5.
// Spec: "calls w.group.FocusNext()" — must advance focused child
func TestWindowTabActuallyChangesFocus(t *testing.T) {
	w := NewWindow(NewRect(0, 0, 40, 20), "Test")
	first := newSelectableMockView(NewRect(0, 0, 5, 3))
	second := newSelectableMockView(NewRect(5, 0, 5, 3))
	w.Insert(first)
	w.Insert(second)
	w.SetFocusedChild(first)

	before := w.FocusedChild()
	w.HandleEvent(tabEvent())
	after := w.FocusedChild()

	if before == after {
		t.Errorf("Tab did not change FocusedChild: before=%v, after=%v (same)", before, after)
	}
	if after != second {
		t.Errorf("Tab: FocusedChild after Tab = %v, want second child", after)
	}
}

// ---------------------------------------------------------------------------
// 10. Falsification: Tab in a plain Group does NOT cycle focus
// ---------------------------------------------------------------------------

// TestGroupTabDoesNotCycleFocus verifies that sending Tab directly to a plain
// Group does not move focus — proving the Tab handling was moved out of Group.
// This is the key behavioral proof of the refactor: clusters' internal Groups
// no longer intercept Tab and cycle within themselves.
// Spec: "prevents Tab from cycling within clusters' internal groups"
func TestGroupTabDoesNotCycleFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	first := newSelectableMockView(NewRect(0, 0, 5, 3))
	second := newSelectableMockView(NewRect(5, 0, 5, 3))
	g.Insert(first)
	g.Insert(second)
	g.SetFocusedChild(first)

	before := g.FocusedChild()
	g.HandleEvent(tabEvent())
	after := g.FocusedChild()

	if before != after {
		t.Errorf("Group Tab cycled focus: before=%v, after=%v — Group should NOT handle Tab", before, after)
	}
	if after != first {
		t.Errorf("Group Tab: FocusedChild = %v, want first (unchanged)", after)
	}
}
