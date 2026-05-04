package tv

// dialog_test.go — Tests for Task 7: Dialog with Modal ExecView.
//
// Tests are organised in two sections:
//  1. Dialog widget tests (construction, Container delegation, Draw, HandleEvent).
//  2. Group.ExecView modal loop tests.
//
// Dialog and NewDialog do not exist yet; those tests will not compile until the
// implementation lands. ExecView tests similarly depend on the implementation
// removing the current panic. All other tests are written to compile and pass
// against the current codebase where possible.
//
// Conventions inherited from existing test files in this package:
//   - newTestScreen(t)              — 80×25 SimulationScreen
//   - newMockView / newSelectableMockView — test doubles from group_test.go
//   - drawOrderMockView             — sentinel view from group_test.go
//   - clickEvent / enterKey / tabKey — helpers from window_interaction_test.go
//
// All assertions cite the spec sentence they verify.

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Section 1 — Dialog
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Requirement 1: Dialog satisfies the Container interface at compile time.
// Spec: "Satisfies var _ Container = (*Dialog)(nil)."
// ---------------------------------------------------------------------------

var _ Container = (*Dialog)(nil)

// ---------------------------------------------------------------------------
// Requirement 2: NewDialog sets SfVisible and OfSelectable.
// Spec: "sets SfVisible, OfSelectable"
// ---------------------------------------------------------------------------

// TestNewDialogSetsSfVisible verifies that a freshly created Dialog has SfVisible.
func TestNewDialogSetsSfVisible(t *testing.T) {
	d := NewDialog(NewRect(5, 5, 40, 15), "Test")

	if !d.HasState(SfVisible) {
		t.Error("NewDialog did not set SfVisible")
	}
}

// TestNewDialogSetsOfSelectable verifies that a freshly created Dialog has OfSelectable.
func TestNewDialogSetsOfSelectable(t *testing.T) {
	d := NewDialog(NewRect(5, 5, 40, 15), "Test")

	if !d.HasOption(OfSelectable) {
		t.Error("NewDialog did not set OfSelectable")
	}
}

// TestNewDialogSfVisibleIsClearable is a falsifying test: clearing SfVisible must work.
func TestNewDialogSfVisibleIsClearable(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 30, 10), "Test")
	d.SetState(SfVisible, false)

	if d.HasState(SfVisible) {
		t.Error("SfVisible should be clearable")
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: NewDialog creates Group with correct client area bounds.
// Spec: "client area is (width-2, height-2)"
// ---------------------------------------------------------------------------

// TestNewDialogClientAreaRenderedInsideFrame verifies that a child inserted at
// (0,0) of the client area is drawn at absolute (1,1) in the dialog's DrawBuffer
// (i.e., just inside the double-line frame).
// Spec: "client area is (width-2, height-2)"
func TestNewDialogClientAreaRenderedInsideFrame(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 20, 10), "Test")
	sentinel := &drawOrderMockView{id: 'X', bounds: NewRect(0, 0, 18, 8)}
	sentinel.SetState(SfVisible, true)
	d.Insert(sentinel)

	buf := NewDrawBuffer(20, 10)
	d.Draw(buf)

	// Child's (0,0) should appear at window-buffer (1,1).
	cell := buf.GetCell(1, 1)
	if cell.Rune != 'X' {
		t.Errorf("client area origin: cell(1,1) = %q, want 'X' (child's (0,0))", cell.Rune)
	}
}

// TestNewDialogClientAreaSizeIsWidthMinus2HeightMinus2 verifies the group client
// area width and height are (width-2) and (height-2) by checking that a child
// clipped to the client rect cannot paint beyond the inner edge.
// Spec: "client area is (width-2, height-2)"
func TestNewDialogClientAreaSizeIsWidthMinus2HeightMinus2(t *testing.T) {
	// Dialog 10 wide × 6 tall → client 8×4, occupying cols 1-8 and rows 1-4.
	d := NewDialog(NewRect(0, 0, 10, 6), "T")
	// Child writes 'Z' at every cell; if the clip is wrong it bleeds onto the border.
	big := &drawOrderMockView{id: 'Z', bounds: NewRect(0, 0, 10, 6)}
	big.SetState(SfVisible, true)
	d.Insert(big)

	buf := NewDrawBuffer(10, 6)
	d.Draw(buf)

	// Border corners must NOT be 'Z'.
	if buf.GetCell(0, 0).Rune == 'Z' {
		t.Error("child painted over top-left border corner — client area clip is wrong")
	}
	if buf.GetCell(9, 0).Rune == 'Z' {
		t.Error("child painted over top-right border corner")
	}
	if buf.GetCell(0, 5).Rune == 'Z' {
		t.Error("child painted over bottom-left border corner")
	}
	if buf.GetCell(9, 5).Rune == 'Z' {
		t.Error("child painted over bottom-right border corner")
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Title() returns the dialog title.
// Spec: "Title() string returns dialog title"
// ---------------------------------------------------------------------------

// TestDialogTitleReturnsTitle verifies Title() returns exactly the string passed to NewDialog.
func TestDialogTitleReturnsTitle(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "Confirm")

	if d.Title() != "Confirm" {
		t.Errorf("Title() = %q, want %q", d.Title(), "Confirm")
	}
}

// TestDialogTitleEmptyString verifies Title() works with an empty title.
func TestDialogTitleEmptyString(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 15), "")

	if d.Title() != "" {
		t.Errorf("Title() = %q, want empty string", d.Title())
	}
}

// TestDialogTitleDoesNotRetainPreviousValue is a falsifying test: two dialogs
// with different titles must not share the same title.
func TestDialogTitleDoesNotRetainPreviousValue(t *testing.T) {
	d1 := NewDialog(NewRect(0, 0, 40, 15), "First")
	d2 := NewDialog(NewRect(0, 0, 40, 15), "Second")

	if d1.Title() == d2.Title() {
		t.Errorf("two dialogs share the same title %q — Title is not stored per instance", d1.Title())
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: SetBounds updates both BaseView and Group bounds.
// Spec: "SetBounds(r Rect) updates both BaseView and Group bounds (client area)"
// ---------------------------------------------------------------------------

// TestDialogSetBoundsUpdatesBounds verifies SetBounds changes the dialog's own bounds.
func TestDialogSetBoundsUpdatesBounds(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 20, 10), "Test")
	newR := NewRect(5, 5, 40, 20)
	d.SetBounds(newR)

	if d.Bounds() != newR {
		t.Errorf("SetBounds: Bounds() = %v, want %v", d.Bounds(), newR)
	}
}

// TestDialogSetBoundsUpdatesClientArea verifies that after SetBounds the internal
// group's client area reflects the new dimensions, evidenced by child rendering.
// Spec: "SetBounds(r Rect) updates both BaseView and Group bounds (client area)"
func TestDialogSetBoundsUpdatesClientArea(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 10, 6), "T")
	// Resize to 20×10.
	d.SetBounds(NewRect(0, 0, 20, 10))

	sentinel := &drawOrderMockView{id: 'Y', bounds: NewRect(0, 0, 18, 8)}
	sentinel.SetState(SfVisible, true)
	d.Insert(sentinel)

	buf := NewDrawBuffer(20, 10)
	d.Draw(buf)

	// After resize, client at (1,1) should still hold the child's sentinel rune.
	cell := buf.GetCell(1, 1)
	if cell.Rune != 'Y' {
		t.Errorf("SetBounds client area update: cell(1,1) = %q, want 'Y'", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: Insert / Remove / Children work via Group delegation.
// Spec: "Implements Container methods via internal Group: Insert, Remove, Children"
// ---------------------------------------------------------------------------

// TestDialogInsertAddsChild verifies Insert adds a view to Children().
func TestDialogInsertAddsChild(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newMockView(NewRect(0, 0, 5, 1))

	d.Insert(child)

	if len(d.Children()) != 1 || d.Children()[0] != child {
		t.Errorf("Insert: Children() = %v, want [child]", d.Children())
	}
}

// TestDialogRemoveRemovesChild verifies Remove removes a previously inserted view.
func TestDialogRemoveRemovesChild(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newMockView(NewRect(0, 0, 5, 1))
	d.Insert(child)

	d.Remove(child)

	for _, c := range d.Children() {
		if c == child {
			t.Error("Remove: child is still present in Children()")
		}
	}
}

// TestDialogChildrenReturnsAllInserted verifies Children() returns all inserted views.
func TestDialogChildrenReturnsAllInserted(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	c1 := newMockView(NewRect(0, 0, 5, 1))
	c2 := newMockView(NewRect(5, 0, 5, 1))
	d.Insert(c1)
	d.Insert(c2)

	got := d.Children()
	if len(got) != 2 {
		t.Errorf("Children() len = %d, want 2", len(got))
	}
}

// TestDialogInsertSetsChildOwnerToDialog verifies that a child's Owner() is the
// Dialog (the facade), not the internal Group.
// Spec: "internal Group with facade=Dialog"
func TestDialogInsertSetsChildOwnerToDialog(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newMockView(NewRect(0, 0, 5, 1))

	d.Insert(child)

	if child.Owner() != d {
		t.Errorf("child.Owner() = %v, want Dialog (facade)", child.Owner())
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: FocusedChild / SetFocusedChild work via Group delegation.
// Spec: "Implements Container methods via internal Group: FocusedChild, SetFocusedChild"
// ---------------------------------------------------------------------------

// TestDialogFocusedChildAfterInsert verifies FocusedChild() returns the last
// inserted selectable child.
func TestDialogFocusedChildAfterInsert(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 10, 1))

	d.Insert(child)

	if d.FocusedChild() != child {
		t.Errorf("FocusedChild() = %v, want inserted selectable child", d.FocusedChild())
	}
}

// TestDialogSetFocusedChildChangesFocus verifies SetFocusedChild moves focus.
func TestDialogSetFocusedChildChangesFocus(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	first := newSelectableMockView(NewRect(0, 0, 10, 1))
	second := newSelectableMockView(NewRect(10, 0, 10, 1))
	d.Insert(first)
	d.Insert(second)

	d.SetFocusedChild(first)

	if d.FocusedChild() != first {
		t.Errorf("SetFocusedChild: FocusedChild() = %v, want first", d.FocusedChild())
	}
}

// TestDialogSetFocusedChildSetsSelectedState verifies the newly focused child
// has SfSelected and the previously focused child does not.
func TestDialogSetFocusedChildSetsSelectedState(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	first := newSelectableMockView(NewRect(0, 0, 10, 1))
	second := newSelectableMockView(NewRect(10, 0, 10, 1))
	d.Insert(first)
	d.Insert(second)
	// second is focused after insertion

	d.SetFocusedChild(first)

	if !first.HasState(SfSelected) {
		t.Error("first should have SfSelected after SetFocusedChild")
	}
	if second.HasState(SfSelected) {
		t.Error("second should have lost SfSelected after SetFocusedChild(first)")
	}
}

// ---------------------------------------------------------------------------
// Requirement 8: Draw renders double-line frame at correct positions.
// Spec: "double-line frame (╔═╗║║╚═╝) using DialogFrame style"
// ---------------------------------------------------------------------------

// newDialogForDraw creates a 20×10 Dialog with a colour scheme attached so Draw
// tests can inspect frame characters and styles without needing a real owner.
func newDialogForDraw(title string) (*Dialog, *DrawBuffer, *theme.ColorScheme) {
	cs := &theme.ColorScheme{
		DialogBackground: tcell.StyleDefault.Background(tcell.ColorNavy),
		DialogFrame:      tcell.StyleDefault.Foreground(tcell.ColorWhite),
		WindowTitle:      tcell.StyleDefault.Foreground(tcell.ColorYellow),
	}
	d := NewDialog(NewRect(0, 0, 20, 10), title)
	d.scheme = cs
	buf := NewDrawBuffer(20, 10)
	return d, buf, cs
}

// TestDialogDrawFrameTopLeftCorner verifies '╔' is at (0,0).
// Spec: "double-line frame (╔═╗║║╚═╝)"
func TestDialogDrawFrameTopLeftCorner(t *testing.T) {
	d, buf, _ := newDialogForDraw("Test")
	d.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != '╔' {
		t.Errorf("frame top-left: cell(0,0) = %q, want '╔'", cell.Rune)
	}
}

// TestDialogDrawFrameTopRightCorner verifies '╗' is at (width-1, 0).
func TestDialogDrawFrameTopRightCorner(t *testing.T) {
	d, buf, _ := newDialogForDraw("Test")
	width := d.Bounds().Width()
	d.Draw(buf)

	cell := buf.GetCell(width-1, 0)
	if cell.Rune != '╗' {
		t.Errorf("frame top-right: cell(%d,0) = %q, want '╗'", width-1, cell.Rune)
	}
}

// TestDialogDrawFrameBottomLeftCorner verifies '╚' is at (0, height-1).
func TestDialogDrawFrameBottomLeftCorner(t *testing.T) {
	d, buf, _ := newDialogForDraw("Test")
	height := d.Bounds().Height()
	d.Draw(buf)

	cell := buf.GetCell(0, height-1)
	if cell.Rune != '╚' {
		t.Errorf("frame bottom-left: cell(0,%d) = %q, want '╚'", height-1, cell.Rune)
	}
}

// TestDialogDrawFrameBottomRightCorner verifies '╝' is at (width-1, height-1).
func TestDialogDrawFrameBottomRightCorner(t *testing.T) {
	d, buf, _ := newDialogForDraw("Test")
	b := d.Bounds()
	d.Draw(buf)

	cell := buf.GetCell(b.Width()-1, b.Height()-1)
	if cell.Rune != '╝' {
		t.Errorf("frame bottom-right: cell(%d,%d) = %q, want '╝'",
			b.Width()-1, b.Height()-1, cell.Rune)
	}
}

// TestDialogDrawFrameTopEdge verifies '═' fills the top horizontal edge.
func TestDialogDrawFrameTopEdge(t *testing.T) {
	// Use empty title so we can sample a mid-top cell unaffected by title text.
	d, buf, _ := newDialogForDraw("")
	width := d.Bounds().Width()
	d.Draw(buf)

	midX := width / 2
	cell := buf.GetCell(midX, 0)
	if cell.Rune != '═' {
		t.Errorf("frame top edge mid: cell(%d,0) = %q, want '═'", midX, cell.Rune)
	}
}

// TestDialogDrawFrameBottomEdge verifies '═' fills the bottom horizontal edge.
func TestDialogDrawFrameBottomEdge(t *testing.T) {
	d, buf, _ := newDialogForDraw("")
	b := d.Bounds()
	d.Draw(buf)

	midX := b.Width() / 2
	cell := buf.GetCell(midX, b.Height()-1)
	if cell.Rune != '═' {
		t.Errorf("frame bottom edge mid: cell(%d,%d) = %q, want '═'",
			midX, b.Height()-1, cell.Rune)
	}
}

// TestDialogDrawFrameLeftEdge verifies '║' on the left vertical edge.
func TestDialogDrawFrameLeftEdge(t *testing.T) {
	d, buf, _ := newDialogForDraw("Test")
	d.Draw(buf)

	cell := buf.GetCell(0, 5)
	if cell.Rune != '║' {
		t.Errorf("frame left edge: cell(0,5) = %q, want '║'", cell.Rune)
	}
}

// TestDialogDrawFrameRightEdge verifies '║' on the right vertical edge.
func TestDialogDrawFrameRightEdge(t *testing.T) {
	d, buf, _ := newDialogForDraw("Test")
	width := d.Bounds().Width()
	d.Draw(buf)

	cell := buf.GetCell(width-1, 5)
	if cell.Rune != '║' {
		t.Errorf("frame right edge: cell(%d,5) = %q, want '║'", width-1, cell.Rune)
	}
}

// TestDialogDrawFrameUsesDialogFrameStyle verifies corners are drawn in DialogFrame style.
// Spec: "double-line frame ... using DialogFrame style"
func TestDialogDrawFrameUsesDialogFrameStyle(t *testing.T) {
	d, buf, cs := newDialogForDraw("Test")
	d.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != cs.DialogFrame {
		t.Errorf("frame top-left style = %v, want DialogFrame %v", cell.Style, cs.DialogFrame)
	}
}

// TestDialogDrawNoCloseIcon verifies that no close icon '[×]' appears at (1,0).
// Spec: "No close/zoom icons."
func TestDialogDrawNoCloseIcon(t *testing.T) {
	d, buf, _ := newDialogForDraw("Test")
	d.Draw(buf)

	// The cell at (2,0) must not be '×' (Window's close icon placement).
	cell := buf.GetCell(2, 0)
	if cell.Rune == '×' {
		t.Error("Dialog should not render a close icon '×' at (2,0)")
	}
}

// TestDialogDrawNoZoomIcon verifies that no zoom icon appears at (width-3, 0).
// Spec: "No close/zoom icons."
func TestDialogDrawNoZoomIcon(t *testing.T) {
	d, buf, _ := newDialogForDraw("Test")
	width := d.Bounds().Width()
	d.Draw(buf)

	// Window places '↑' at (width-3, 0) — Dialog must not.
	cell := buf.GetCell(width-3, 0)
	if cell.Rune == '↑' || cell.Rune == '↕' {
		t.Errorf("Dialog should not render a zoom icon at (%d,0), got %q", width-3, cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Requirement 9: Draw renders title centered in top border using WindowTitle style.
// Spec: "title centered in top border using WindowTitle style"
// ---------------------------------------------------------------------------

// TestDialogDrawTitleAppearsInTopBorder verifies the title text is rendered on row 0.
func TestDialogDrawTitleAppearsInTopBorder(t *testing.T) {
	d, buf, _ := newDialogForDraw("Hi")
	width := d.Bounds().Width()
	d.Draw(buf)

	found := false
	for x := 1; x < width-1; x++ {
		if buf.GetCell(x, 0).Rune == 'H' {
			found = true
			break
		}
	}
	if !found {
		t.Error("title 'Hi' not found in top border row")
	}
}

// TestDialogDrawTitleUsesWindowTitleStyle verifies the title characters use WindowTitle style.
// Spec: "title centered in top border using WindowTitle style"
func TestDialogDrawTitleUsesWindowTitleStyle(t *testing.T) {
	d, buf, cs := newDialogForDraw("Hi")
	width := d.Bounds().Width()
	d.Draw(buf)

	for x := 1; x < width-1; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune == 'H' {
			if cell.Style != cs.WindowTitle {
				t.Errorf("title rune 'H' at (%d,0): style = %v, want WindowTitle %v",
					x, cell.Style, cs.WindowTitle)
			}
			return
		}
	}
	t.Error("title rune 'H' not found in top border — cannot check style")
}

// TestDialogDrawTitleCenteredInBorder verifies the title appears in the middle
// of the top border rather than being flush with a corner.
func TestDialogDrawTitleCenteredInBorder(t *testing.T) {
	// 20-wide dialog, title "AB" → centred somewhere in cols 1-18.
	// A centred title should not start at column 1 (flush left).
	d, buf, _ := newDialogForDraw("AB")
	d.Draw(buf)

	// Find 'A' in the top row.
	aPos := -1
	for x := 1; x < 19; x++ {
		if buf.GetCell(x, 0).Rune == 'A' {
			aPos = x
			break
		}
	}
	if aPos < 0 {
		t.Fatal("title 'AB' not found in top border")
	}
	// For a 20-wide dialog, 'A' must not be at the extreme edge (col 1) — it must be centred.
	if aPos == 1 {
		t.Errorf("title appears flush left at column 1 — expected centering")
	}
}

// TestDialogDrawTitleWrappedInSpaces verifies the character immediately before and
// after the title text are spaces (in WindowTitle style).
// Spec: "title centered in top border using WindowTitle style" — same as Window's treatment.
func TestDialogDrawTitleWrappedInSpaces(t *testing.T) {
	d, buf, cs := newDialogForDraw("XY")
	width := d.Bounds().Width()
	d.Draw(buf)

	xPos := -1
	for x := 1; x < width-1; x++ {
		if buf.GetCell(x, 0).Rune == 'X' && buf.GetCell(x+1, 0).Rune == 'Y' {
			xPos = x
			break
		}
	}
	if xPos < 0 {
		t.Fatal("title 'XY' not found in top border")
	}

	before := buf.GetCell(xPos-1, 0)
	after := buf.GetCell(xPos+2, 0)

	if before.Rune != ' ' || before.Style != cs.WindowTitle {
		t.Errorf("title leading space: cell(%d,0) = %q style=%v, want ' ' WindowTitle",
			xPos-1, before.Rune, before.Style)
	}
	if after.Rune != ' ' || after.Style != cs.WindowTitle {
		t.Errorf("title trailing space: cell(%d,0) = %q style=%v, want ' ' WindowTitle",
			xPos+2, after.Rune, after.Style)
	}
}

// ---------------------------------------------------------------------------
// Requirement 10: Draw renders children in client-area sub-buffer (offset by frame).
// Spec: "Draws children in client-area sub-buffer."
// ---------------------------------------------------------------------------

// TestDialogDrawChildrenInClientSubBuffer verifies a child's (0,0) maps to (1,1)
// in the dialog buffer.
func TestDialogDrawChildrenInClientSubBuffer(t *testing.T) {
	d, buf, _ := newDialogForDraw("Test")
	sentinel := &drawOrderMockView{id: 'Q', bounds: NewRect(0, 0, 18, 8)}
	sentinel.SetState(SfVisible, true)
	d.Insert(sentinel)

	d.Draw(buf)

	cell := buf.GetCell(1, 1)
	if cell.Rune != 'Q' {
		t.Errorf("children in client SubBuffer: cell(1,1) = %q, want 'Q'", cell.Rune)
	}
}

// TestDialogDrawChildrenCannotPaintOverFrame verifies children are clipped to the
// client area and cannot overwrite the border.
func TestDialogDrawChildrenCannotPaintOverFrame(t *testing.T) {
	d, buf, _ := newDialogForDraw("Test")
	// Child that fills the entire dialog buffer — must be clipped.
	big := &drawOrderMockView{id: 'F', bounds: NewRect(0, 0, 20, 10)}
	big.SetState(SfVisible, true)
	d.Insert(big)

	d.Draw(buf)

	// The top-left frame corner must still be a frame char, not 'F'.
	cell := buf.GetCell(0, 0)
	if cell.Rune == 'F' {
		t.Errorf("child overwrote border at (0,0): got %q, want frame char", cell.Rune)
	}
}

// TestDialogDrawClientAreaBackground verifies the client area is filled with
// DialogBackground style and a space rune.
// Spec: "client area filled with DialogBackground"
func TestDialogDrawClientAreaBackground(t *testing.T) {
	d, buf, cs := newDialogForDraw("Test")
	d.Draw(buf)

	cell := buf.GetCell(1, 1)
	if cell.Rune != ' ' {
		t.Errorf("client area bg: cell(1,1) rune = %q, want ' '", cell.Rune)
	}
	if cell.Style != cs.DialogBackground {
		t.Errorf("client area bg: cell(1,1) style = %v, want DialogBackground %v",
			cell.Style, cs.DialogBackground)
	}
}

// TestDialogDrawNoColorSchemeDoesNotPanic verifies Draw is safe with no ColorScheme.
func TestDialogDrawNoColorSchemeDoesNotPanic(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 20, 10), "Test")
	buf := NewDrawBuffer(20, 10)

	// Must not panic.
	d.Draw(buf)
}

// ---------------------------------------------------------------------------
// Requirement 11: HandleEvent routes keyboard to Group (Tab traversal works).
// Spec: "keyboard/command events delegate to Group (three-phase dispatch + Tab traversal)"
// ---------------------------------------------------------------------------

// TestDialogHandleEventTabAdvancesFocus verifies Tab moves focus forward between
// two selectable children.
func TestDialogHandleEventTabAdvancesFocus(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	btn1 := NewButton(NewRect(0, 0, 10, 1), "One", CmOK)
	btn2 := NewButton(NewRect(12, 0, 10, 1), "Two", CmCancel)
	d.Insert(btn1)
	d.Insert(btn2)
	d.SetFocusedChild(btn1)

	d.HandleEvent(tabKey())

	if d.FocusedChild() != btn2 {
		t.Errorf("after Tab, FocusedChild() = %v, want btn2", d.FocusedChild())
	}
}

// TestDialogHandleEventShiftTabMovesFocusBackward verifies Shift+Tab moves focus
// backward.
func TestDialogHandleEventShiftTabMovesFocusBackward(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	btn1 := NewButton(NewRect(0, 0, 10, 1), "One", CmOK)
	btn2 := NewButton(NewRect(12, 0, 10, 1), "Two", CmCancel)
	d.Insert(btn1)
	d.Insert(btn2)
	// btn2 is focused after insertion

	d.HandleEvent(shiftTabKey())

	if d.FocusedChild() != btn1 {
		t.Errorf("after Shift+Tab, FocusedChild() = %v, want btn1", d.FocusedChild())
	}
}

// TestDialogHandleEventKeyboardRoutesToFocusedChild verifies keyboard events
// (other than Tab or Enter, which have special handling) reach the focused child.
func TestDialogHandleEventKeyboardRoutesToFocusedChild(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 10, 1))
	d.Insert(child)

	// Use a plain rune key — not Enter (broadcasts CmDefault) or Tab (traversal).
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'a'}}
	d.HandleEvent(ev)

	if child.eventHandled != ev {
		t.Error("keyboard event not forwarded to focused child")
	}
}

// TestDialogHandleEventCommandRoutesToGroup verifies EvCommand events reach children.
func TestDialogHandleEventCommandRoutesToGroup(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 10, 1))
	d.Insert(child)

	ev := &Event{What: EvCommand, Command: CmClose}
	d.HandleEvent(ev)

	if child.eventHandled != ev {
		t.Error("command event not forwarded to focused child")
	}
}

// ---------------------------------------------------------------------------
// Requirement 12: HandleEvent routes mouse with (-1,-1) translation to client area.
// Spec: "Mouse events: translate coordinates by (-1,-1) for frame offset,
//        forward to Group only if in client area."
// ---------------------------------------------------------------------------

// TestDialogHandleEventMouseInClientAreaForwarded verifies a mouse click inside
// the client area (offset by -1,-1) is forwarded to the group.
func TestDialogHandleEventMouseInClientAreaForwarded(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 38, 18))
	d.Insert(child)

	// Click at dialog-local (5,5) — inside the client area (cols 1-38, rows 1-18).
	ev := clickEvent(5, 5, tcell.Button1)
	d.HandleEvent(ev)

	if child.eventHandled == nil {
		t.Error("mouse event inside client area was not forwarded to child")
	}
}

// TestDialogHandleEventMouseCoordinatesTranslated verifies the (-1,-1) translation
// is applied before the event reaches the Group.
// Spec: "translate coordinates by (-1,-1) for frame offset"
func TestDialogHandleEventMouseCoordinatesTranslated(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 38, 18))
	d.Insert(child)

	// Click at dialog-local (3, 4). After (-1,-1) the group receives (2, 3).
	ev := clickEvent(3, 4, tcell.Button1)
	d.HandleEvent(ev)

	if child.eventHandled == nil {
		t.Fatal("mouse event not forwarded to child")
	}
	got := child.eventHandled.Mouse
	if got == nil {
		t.Fatal("forwarded event has nil Mouse")
	}
	if got.X != 2 || got.Y != 3 {
		t.Errorf("translated mouse coords = (%d,%d), want (2,3)", got.X, got.Y)
	}
}

// ---------------------------------------------------------------------------
// Requirement 13: HandleEvent discards mouse events on frame (outside client area).
// Spec: "forward to Group only if in client area"
// ---------------------------------------------------------------------------

// TestDialogHandleEventMouseOnTopBorderDiscarded verifies a click on the top border
// (row 0) is not forwarded to children.
func TestDialogHandleEventMouseOnTopBorderDiscarded(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 38, 18))
	d.Insert(child)

	// Click on top border (row 0).
	ev := clickEvent(5, 0, tcell.Button1)
	d.HandleEvent(ev)

	if child.eventHandled != nil {
		t.Error("mouse click on top border was forwarded to child — should be discarded")
	}
}

// TestDialogHandleEventMouseOnLeftBorderDiscarded verifies a click on the left border
// (col 0) is not forwarded to children.
func TestDialogHandleEventMouseOnLeftBorderDiscarded(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 38, 18))
	d.Insert(child)

	// Click on left border (col 0).
	ev := clickEvent(0, 5, tcell.Button1)
	d.HandleEvent(ev)

	if child.eventHandled != nil {
		t.Error("mouse click on left border was forwarded to child — should be discarded")
	}
}

// TestDialogHandleEventMouseOnRightBorderDiscarded verifies a click on the right border
// is not forwarded to children.
func TestDialogHandleEventMouseOnRightBorderDiscarded(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 38, 18))
	d.Insert(child)

	width := d.Bounds().Width()
	ev := clickEvent(width-1, 5, tcell.Button1)
	d.HandleEvent(ev)

	if child.eventHandled != nil {
		t.Error("mouse click on right border was forwarded to child — should be discarded")
	}
}

// TestDialogHandleEventMouseOnBottomBorderDiscarded verifies a click on the bottom
// border is not forwarded to children.
func TestDialogHandleEventMouseOnBottomBorderDiscarded(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	child := newSelectableMockView(NewRect(0, 0, 38, 18))
	d.Insert(child)

	height := d.Bounds().Height()
	ev := clickEvent(5, height-1, tcell.Button1)
	d.HandleEvent(ev)

	if child.eventHandled != nil {
		t.Error("mouse click on bottom border was forwarded to child — should be discarded")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Group.ExecView (modal execution loop)
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// ExecView test setup helper
// ---------------------------------------------------------------------------

// execViewStack builds the full Application→Desktop→Window owner chain that
// ExecView needs to walk in order to find the *Desktop and its .app pointer.
// Returns the app, the window, and the screen (already Init'd).
//
// Callers must NOT call screen.Fini() — that is left for the ExecView modal
// loop to detect (nil PollEvent) or for the test-managed goroutine to handle.
func execViewStack(t *testing.T) (*Application, *Window, tcell.SimulationScreen) {
	t.Helper()
	screen := newTestScreen(t)
	app, err := NewApplication(
		WithScreen(screen),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		screen.Fini()
		t.Fatalf("NewApplication: %v", err)
	}
	win := NewWindow(NewRect(5, 5, 40, 15), "Host", WithWindowNumber(1))
	app.Desktop().Insert(win)
	return app, win, screen
}

// ---------------------------------------------------------------------------
// Requirement 16: ExecView sets SfModal on the inserted view.
// Spec: "Inserts v into Group, sets SfModal on v."
// ---------------------------------------------------------------------------

// TestExecViewSetsSfModalOnView verifies that SfModal is set on the view before
// the modal loop begins.  We use screen.Fini() to cause PollEvent to return nil
// immediately, which terminates the loop.
func TestExecViewSetsSfModalOnView(t *testing.T) {
	_, win, screen := execViewStack(t)

	dialog := NewDialog(NewRect(10, 5, 30, 12), "Modal")

	// Track whether SfModal was seen during the modal loop by sampling it just
	// after ExecView has started.  We give ExecView a tiny head-start then
	// finalize the screen.
	modalSeen := make(chan bool, 1)
	go func() {
		// Let ExecView run briefly.
		time.Sleep(20 * time.Millisecond)
		modalSeen <- dialog.HasState(SfModal)
		screen.Fini()
	}()

	// ExecView runs in a separate goroutine so the test doesn't block.
	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	select {
	case <-done:
		// loop ended; now check the recorded value
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not exit within 2 s after screen.Fini()")
	}

	select {
	case seen := <-modalSeen:
		if !seen {
			t.Error("SfModal was not set on the view during ExecView")
		}
	default:
		t.Error("modalSeen channel was empty — goroutine did not run before ExecView exited")
	}
}

// ---------------------------------------------------------------------------
// Requirement 14/18: ExecView returns CmOK when OK button is pressed.
// Spec: "Closing commands: CmOK, CmCancel, CmClose, CmYes, CmNo — returns the
//        command code."
// ---------------------------------------------------------------------------

// TestExecViewReturnsCmOKOnOKButton verifies the modal loop returns CmOK when an
// OK button inside the dialog is activated by an Enter key press.
func TestExecViewReturnsCmOKOnOKButton(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(10, 5, 30, 12), "Confirm")
	okBtn := NewButton(NewRect(0, 0, 10, 1), "OK", CmOK)
	dialog.Insert(okBtn)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	// Give ExecView time to enter the loop.
	time.Sleep(30 * time.Millisecond)

	// Inject Enter — the focused OK button transforms it to CmOK.
	app.PostCommand(CmOK, nil)

	select {
	case code := <-done:
		if code != CmOK {
			t.Errorf("ExecView returned %v, want CmOK", code)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmOK within 2 s after button press")
	}
}

// TestExecViewReturnsCmCancelOnCancelButton verifies the modal loop returns
// CmCancel when a Cancel button is activated.
func TestExecViewReturnsCmCancelOnCancelButton(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(10, 5, 30, 12), "Confirm")
	cancelBtn := NewButton(NewRect(0, 0, 12, 1), "Cancel", CmCancel)
	dialog.Insert(cancelBtn)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmCancel, nil)

	select {
	case code := <-done:
		if code != CmCancel {
			t.Errorf("ExecView returned %v, want CmCancel", code)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmCancel within 2 s")
	}
}

// TestExecViewReturnsCmYesOnYesButton verifies the modal loop returns CmYes.
func TestExecViewReturnsCmYesOnYesButton(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(10, 5, 30, 12), "Ask")
	yesBtn := NewButton(NewRect(0, 0, 8, 1), "Yes", CmYes)
	dialog.Insert(yesBtn)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmYes, nil)

	select {
	case code := <-done:
		if code != CmYes {
			t.Errorf("ExecView returned %v, want CmYes", code)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmYes within 2 s")
	}
}

// TestExecViewReturnsCmNoOnNoButton verifies the modal loop returns CmNo.
func TestExecViewReturnsCmNoOnNoButton(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(10, 5, 30, 12), "Ask")
	noBtn := NewButton(NewRect(0, 0, 8, 1), "No", CmNo)
	dialog.Insert(noBtn)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmNo, nil)

	select {
	case code := <-done:
		if code != CmNo {
			t.Errorf("ExecView returned %v, want CmNo", code)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmNo within 2 s")
	}
}

// TestExecViewReturnsCmCancelOnCmClose verifies the modal loop returns CmCancel
// when CmClose is posted to a modal dialog.
//
// Task 10: Dialog.HandleEvent transforms CmClose → CmCancel when SfModal is set,
// so ExecView's modal loop receives CmCancel and returns it.
func TestExecViewReturnsCmCancelOnCmClose(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(10, 5, 30, 12), "Ask")

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmClose, nil)

	select {
	case code := <-done:
		if code != CmCancel {
			t.Errorf("ExecView returned %v after CmClose posted to modal dialog, want CmCancel", code)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return within 2 s after CmClose posted to modal dialog")
	}
}

// ---------------------------------------------------------------------------
// Requirement 15: ExecView returns CmCancel when PollEvent returns nil.
// Spec: "If PollEvent returns nil, removes v, clears SfModal, returns CmCancel."
// ---------------------------------------------------------------------------

// TestExecViewReturnsCmCancelWhenScreenFinalized verifies that finalizing the
// screen (causing PollEvent to return nil) causes ExecView to return CmCancel.
func TestExecViewReturnsCmCancelWhenScreenFinalized(t *testing.T) {
	_, win, screen := execViewStack(t)

	dialog := NewDialog(NewRect(10, 5, 30, 12), "Modal")

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	// Finalize the screen to unblock PollEvent with nil.
	time.Sleep(20 * time.Millisecond)
	screen.Fini()

	select {
	case code := <-done:
		if code != CmCancel {
			t.Errorf("ExecView returned %v after screen.Fini(), want CmCancel", code)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView did not return CmCancel within 2 s after screen.Fini()")
	}
}

// ---------------------------------------------------------------------------
// Requirement 17: ExecView removes view and clears SfModal after completion.
// Spec: "If PollEvent returns nil, removes v, clears SfModal, returns CmCancel."
// (Also applies after any closing command.)
// ---------------------------------------------------------------------------

// TestExecViewClearsSfModalAfterCompletion verifies SfModal is cleared on the
// view after ExecView returns.
func TestExecViewClearsSfModalAfterCompletion(t *testing.T) {
	_, win, screen := execViewStack(t)

	dialog := NewDialog(NewRect(10, 5, 30, 12), "Modal")

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(20 * time.Millisecond)
	screen.Fini()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}

	if dialog.HasState(SfModal) {
		t.Error("SfModal is still set on the view after ExecView returned — it should be cleared")
	}
}

// TestExecViewRemovesViewFromGroupAfterCompletion verifies the view is removed
// from the group's children list after ExecView returns.
func TestExecViewRemovesViewFromGroupAfterCompletion(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(10, 5, 30, 12), "Modal")

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmOK, nil)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}

	// The dialog should no longer be in the host window's children.
	for _, child := range win.Children() {
		if child == dialog {
			t.Error("ExecView did not remove the view from the group after completion")
		}
	}
}

// ---------------------------------------------------------------------------
// Requirement 19: ExecView discards mouse events outside the modal view's bounds.
// Spec: "Mouse events inside modal view bounds are translated and forwarded;
//        outside clicks discarded."
// ---------------------------------------------------------------------------

// TestExecViewDiscardsMouseEventsOutsideModalBounds verifies that mouse clicks
// outside the modal dialog's bounds are not forwarded to the dialog's children.
func TestExecViewDiscardsMouseEventsOutsideModalBounds(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	// Dialog at absolute bounds (10, 5, 30, 12) — occupies cols 10-39, rows 5-16.
	dialog := NewDialog(NewRect(10, 5, 30, 12), "Modal")
	child := newSelectableMockView(NewRect(0, 0, 28, 10))
	dialog.Insert(child)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(30 * time.Millisecond)

	// Click well outside the dialog bounds (col 2, row 2 — in the desktop).
	app.screen.PostEvent(tcell.NewEventMouse(2, 2, tcell.Button1, tcell.ModNone))

	time.Sleep(20 * time.Millisecond)

	// The child should not have received the event.
	if child.eventHandled != nil {
		t.Error("ExecView forwarded a mouse event outside modal bounds to the child — it should be discarded")
	}

	// Terminate the loop cleanly.
	app.PostCommand(CmCancel, nil)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}
}

// ---------------------------------------------------------------------------
// Requirement 20: ExecView returns CmCancel when app is not found in owner chain.
// Spec: "Walks Owner() chain from g.facade to find *Desktop, accesses desktop.app."
// ---------------------------------------------------------------------------

// TestExecViewReturnsCmCancelWhenNoDesktopInOwnerChain verifies that calling
// ExecView on a Group whose facade has no Desktop in its Owner chain returns
// CmCancel immediately rather than hanging or panicking.
func TestExecViewReturnsCmCancelWhenNoDesktopInOwnerChain(t *testing.T) {
	// Create a standalone group with no owner chain leading to a Desktop.
	g := NewGroup(NewRect(0, 0, 40, 20))
	// Set facade to a bare container that has no owner.
	facade := &mockFacadeContainer{}
	g.SetFacade(facade)

	dialog := NewDialog(NewRect(0, 0, 20, 10), "Orphan")

	result := make(chan CommandCode, 1)
	go func() {
		result <- g.ExecView(dialog)
	}()

	select {
	case code := <-result:
		if code != CmCancel {
			t.Errorf("ExecView without Desktop in owner chain returned %v, want CmCancel", code)
		}
	case <-time.After(2 * time.Second):
		t.Error("ExecView blocked for 2 s without a Desktop in the owner chain — expected immediate CmCancel")
	}
}

// TestExecViewNoOwnerChainDoesNotPanic verifies ExecView on an ownerless group
// does not panic.
func TestExecViewNoOwnerChainDoesNotPanic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	dialog := NewDialog(NewRect(0, 0, 20, 10), "Orphan")

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ExecView panicked with no owner chain: %v", r)
		}
	}()

	result := make(chan CommandCode, 1)
	go func() {
		result <- g.ExecView(dialog)
	}()

	select {
	case <-result:
		// returned cleanly
	case <-time.After(2 * time.Second):
		t.Error("ExecView blocked for 2 s — expected CmCancel when no owner chain")
	}
}

// ---------------------------------------------------------------------------
// Extra: DialogOption functional option type is usable.
// Spec: "DialogOption is func(*Dialog)"
// ---------------------------------------------------------------------------

// TestDialogOptionFuncApplied verifies that a DialogOption (func(*Dialog)) is
// applied to the dialog during construction.
func TestDialogOptionFuncApplied(t *testing.T) {
	applied := false
	opt := func(d *Dialog) {
		applied = true
	}

	_ = NewDialog(NewRect(0, 0, 30, 10), "Opt", opt)

	if !applied {
		t.Error("DialogOption was not applied by NewDialog")
	}
}

// TestDialogOptionCanSetState verifies a DialogOption can mutate the dialog
// (e.g. set an additional state flag).
func TestDialogOptionCanSetState(t *testing.T) {
	opt := func(d *Dialog) {
		d.SetState(SfDisabled, true)
	}

	d := NewDialog(NewRect(0, 0, 30, 10), "Opt", opt)

	if !d.HasState(SfDisabled) {
		t.Error("DialogOption that sets SfDisabled was not applied")
	}
}

// ---------------------------------------------------------------------------
// Extra: BringToFront is available on Dialog.
// Spec: "Implements ... plus BringToFront."
// ---------------------------------------------------------------------------

// TestDialogBringToFrontMovesChildToEnd verifies BringToFront re-orders children
// so the specified view is last (and thus drawn on top).
// TestExecViewMouseCoordinatesTranslatedFromScreenToLocal verifies that when
// a dialog is opened via ExecView on a Window (not on Desktop), mouse events
// in screen-absolute coordinates are correctly translated to the dialog's
// local coordinate space. Bug: ExecView compared raw screen coords against
// the dialog's owner-local bounds, so clicks never matched for non-zero-origin windows.
func TestExecViewMouseCoordinatesTranslatedFromScreenToLocal(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	// Window is at (5, 5, 40, 15) in the Desktop which is at (0, 0).
	// Window client area starts at screen (6, 6) — frame offset (1, 1).
	// Place dialog at (2, 1) within client area → screen (8, 7).
	dialog := NewDialog(NewRect(2, 1, 20, 8), "Find")
	child := newSelectableMockView(NewRect(0, 0, 18, 6))
	dialog.Insert(child)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(30 * time.Millisecond)

	// Click inside the dialog at screen (12, 10).
	// Expected dialog-relative: (12-8, 10-7) = (4, 3)
	// Dialog.HandleEvent subtracts frame (1,1) → client (3, 2)
	// Child at (0,0) receives (3, 2).
	app.screen.PostEvent(tcell.NewEventMouse(12, 10, tcell.Button1, tcell.ModNone))
	time.Sleep(30 * time.Millisecond)

	if child.eventHandled == nil {
		t.Fatal("ExecView did not forward mouse event inside dialog — screen-to-local coordinate translation is broken")
	}
	if child.eventHandled.Mouse == nil {
		t.Fatal("child received event but Mouse is nil")
	}
	gotX, gotY := child.eventHandled.Mouse.X, child.eventHandled.Mouse.Y
	if gotX != 3 || gotY != 2 {
		t.Errorf("child got mouse at (%d, %d), want (3, 2)", gotX, gotY)
	}

	app.PostCommand(CmCancel, nil)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}
}

// TestExecViewKeyboardEventsReachDialogChild verifies keyboard events in ExecView
// are delivered to the dialog's focused child.
func TestExecViewKeyboardEventsReachDialogChild(t *testing.T) {
	app, win, screen := execViewStack(t)
	defer screen.Fini()

	dialog := NewDialog(NewRect(2, 1, 30, 10), "Find")
	inputLine := NewInputLine(NewRect(1, 1, 20, 1), 100)
	dialog.Insert(inputLine)
	okBtn := NewButton(NewRect(1, 5, 10, 2), "OK", CmOK, WithDefault())
	dialog.Insert(okBtn)
	dialog.SetFocusedChild(inputLine)

	done := make(chan CommandCode, 1)
	go func() {
		done <- win.ExecView(dialog)
	}()

	time.Sleep(30 * time.Millisecond)

	// Type 'a' — should reach the InputLine
	app.screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	time.Sleep(30 * time.Millisecond)

	if inputLine.Text() != "a" {
		t.Errorf("InputLine text = %q, want %q — keyboard events not reaching focused child in ExecView", inputLine.Text(), "a")
	}

	app.PostCommand(CmCancel, nil)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecView did not return within 2 s")
	}
}

func TestDialogBringToFrontMovesChildToEnd(t *testing.T) {
	d := NewDialog(NewRect(0, 0, 40, 20), "Test")
	first := newSelectableMockView(NewRect(0, 0, 10, 1))
	second := newSelectableMockView(NewRect(10, 0, 10, 1))
	d.Insert(first)
	d.Insert(second)

	d.BringToFront(first)

	children := d.Children()
	if len(children) == 0 {
		t.Fatal("Children() is empty after BringToFront")
	}
	if children[len(children)-1] != first {
		t.Errorf("BringToFront: last child = %v, want first (brought to front)", children[len(children)-1])
	}
}
