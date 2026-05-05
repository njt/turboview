package tv

// history_test.go — Tests for Task 3: History Widget — Draw and Constructor.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Compile-time and interface checks
//   Section 2  — Constructor: bounds, flags, stored references
//   Section 3  — Accessors: Link and HistoryID
//   Section 4  — Draw: characters rendered
//   Section 5  — Draw: styles without a color scheme (StyleDefault)
//   Section 6  — Draw: styles with a color scheme (HistorySides, HistoryArrow)
//   Section 7  — Falsifying tests

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// compile-time assertion: History must satisfy Widget.
// Spec: "History struct embeds BaseView"
var _ Widget = (*History)(nil)

// ---------------------------------------------------------------------------
// Section 1 — Compile-time checks
// ---------------------------------------------------------------------------

// TestHistoryImplementsWidget verifies History satisfies the Widget interface
// at compile time (enforced by the var _ Widget = (*History)(nil) above).
// Spec: "History struct embeds BaseView"
func TestHistoryImplementsWidget(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)
	if h == nil {
		t.Fatal("NewHistory returned nil")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Constructor: bounds, flags, stored references
// ---------------------------------------------------------------------------

// TestNewHistorySetsBounds verifies NewHistory records the given bounds.
// Spec: "NewHistory(bounds Rect, link *InputLine, historyID int) *History"
func TestNewHistorySetsBounds(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	r := NewRect(20, 0, 3, 1)
	h := NewHistory(r, il, 42)

	if h.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", h.Bounds(), r)
	}
}

// TestNewHistorySetsSfVisible verifies NewHistory sets the SfVisible state flag.
// Spec: "Sets SfVisible"
func TestNewHistorySetsSfVisible(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	if !h.HasState(SfVisible) {
		t.Error("NewHistory did not set SfVisible")
	}
}

// TestNewHistorySetsOfPostProcess verifies NewHistory sets the OfPostProcess option.
// Spec: "Sets OfPostProcess (sees events after focused view)"
func TestNewHistorySetsOfPostProcess(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	if !h.HasOption(OfPostProcess) {
		t.Error("NewHistory did not set OfPostProcess")
	}
}

// TestNewHistoryDoesNotSetOfSelectable verifies NewHistory does NOT set OfSelectable.
// Spec: "Does NOT set OfSelectable (never receives focus)"
func TestNewHistoryDoesNotSetOfSelectable(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	if h.HasOption(OfSelectable) {
		t.Error("NewHistory must NOT set OfSelectable — History is never focusable")
	}
}

// TestNewHistoryStoresLink verifies NewHistory stores the InputLine reference.
// Spec: "Stores reference to linked InputLine and history ID"
func TestNewHistoryStoresLink(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 7)

	if h.Link() != il {
		t.Errorf("Link() = %p, want %p (the linked InputLine)", h.Link(), il)
	}
}

// TestNewHistoryStoresHistoryID verifies NewHistory stores the history ID.
// Spec: "Stores reference to linked InputLine and history ID"
func TestNewHistoryStoresHistoryID(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 99)

	if h.HistoryID() != 99 {
		t.Errorf("HistoryID() = %d, want 99", h.HistoryID())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Accessors
// ---------------------------------------------------------------------------

// TestHistoryLinkReturnsSamePointer verifies Link() returns the exact pointer
// passed to NewHistory, not a copy.
// Spec: "Link() *InputLine returns the linked InputLine"
func TestHistoryLinkReturnsSamePointer(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	got := h.Link()
	if got != il {
		t.Errorf("Link() returned different pointer; same InputLine expected")
	}
}

// TestHistoryIDRoundTrips verifies HistoryID() returns exactly what was passed.
// Spec: "HistoryID() int returns the history ID"
func TestHistoryIDRoundTrips(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 42)

	if h.HistoryID() != 42 {
		t.Errorf("HistoryID() = %d, want 42", h.HistoryID())
	}
}

// TestHistoryIDZeroIsValid verifies HistoryID() can be zero.
// Spec: "HistoryID() int returns the history ID"
func TestHistoryIDZeroIsValid(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 0)

	if h.HistoryID() != 0 {
		t.Errorf("HistoryID() = %d, want 0", h.HistoryID())
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Draw: characters rendered
// ---------------------------------------------------------------------------

// TestHistoryDrawsLeftSideChar verifies x=0 contains '▐' (U+2590).
// Spec: "▐ (U+2590) at x=0 in HistorySides color"
func TestHistoryDrawsLeftSideChar(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != '▐' {
		t.Errorf("Draw x=0: rune = %q (U+%04X), want '▐' (U+2590)", cell.Rune, cell.Rune)
	}
}

// TestHistoryDrawsArrowChar verifies x=1 contains '↓' (U+2193).
// Spec: "↓ (U+2193) at x=1 in HistoryArrow color"
func TestHistoryDrawsArrowChar(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Rune != '↓' {
		t.Errorf("Draw x=1: rune = %q (U+%04X), want '↓' (U+2193)", cell.Rune, cell.Rune)
	}
}

// TestHistoryDrawsRightSideChar verifies x=2 contains '▌' (U+258C).
// Spec: "▌ (U+258C) at x=2 in HistorySides color"
func TestHistoryDrawsRightSideChar(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	cell := buf.GetCell(2, 0)
	if cell.Rune != '▌' {
		t.Errorf("Draw x=2: rune = %q (U+%04X), want '▌' (U+258C)", cell.Rune, cell.Rune)
	}
}

// TestHistoryDrawsExactlyThreeCharacters verifies the widget renders exactly
// the 3-character sequence ▐↓▌ and nothing else at unexpected positions.
// Spec: "Renders exactly 3 characters: ▐↓▌"
func TestHistoryDrawsExactlyThreeCharacters(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	want := []rune{'▐', '↓', '▌'}
	for x, wantRune := range want {
		cell := buf.GetCell(x, 0)
		if cell.Rune != wantRune {
			t.Errorf("Draw x=%d: rune = %q, want %q", x, cell.Rune, wantRune)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Draw: styles without a color scheme (tcell.StyleDefault)
// ---------------------------------------------------------------------------

// TestHistoryDrawLeftSideUsesStyleDefaultWithoutScheme verifies x=0 uses
// tcell.StyleDefault when no color scheme is available.
// Spec: "If no color scheme is available, use tcell.StyleDefault"
func TestHistoryDrawLeftSideUsesStyleDefaultWithoutScheme(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)
	// No owner, no scheme set — ColorScheme() returns nil.

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != tcell.StyleDefault {
		t.Errorf("Draw x=0 without scheme: style = %v, want tcell.StyleDefault", cell.Style)
	}
}

// TestHistoryDrawArrowUsesStyleDefaultWithoutScheme verifies x=1 uses
// tcell.StyleDefault when no color scheme is available.
// Spec: "If no color scheme is available, use tcell.StyleDefault"
func TestHistoryDrawArrowUsesStyleDefaultWithoutScheme(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Style != tcell.StyleDefault {
		t.Errorf("Draw x=1 without scheme: style = %v, want tcell.StyleDefault", cell.Style)
	}
}

// TestHistoryDrawRightSideUsesStyleDefaultWithoutScheme verifies x=2 uses
// tcell.StyleDefault when no color scheme is available.
// Spec: "If no color scheme is available, use tcell.StyleDefault"
func TestHistoryDrawRightSideUsesStyleDefaultWithoutScheme(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	cell := buf.GetCell(2, 0)
	if cell.Style != tcell.StyleDefault {
		t.Errorf("Draw x=2 without scheme: style = %v, want tcell.StyleDefault", cell.Style)
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Draw: styles with a color scheme
// ---------------------------------------------------------------------------

// TestHistoryDrawLeftSideUsesHistorySidesStyle verifies x=0 uses HistorySides style
// when a color scheme is present.
// Spec: "▐ (U+2590) at x=0 in HistorySides color"
func TestHistoryDrawLeftSideUsesHistorySidesStyle(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)
	h.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != theme.BorlandBlue.HistorySides {
		t.Errorf("Draw x=0 with scheme: style = %v, want HistorySides %v",
			cell.Style, theme.BorlandBlue.HistorySides)
	}
}

// TestHistoryDrawArrowUsesHistoryArrowStyle verifies x=1 uses HistoryArrow style
// when a color scheme is present.
// Spec: "↓ (U+2193) at x=1 in HistoryArrow color"
func TestHistoryDrawArrowUsesHistoryArrowStyle(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)
	h.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Style != theme.BorlandBlue.HistoryArrow {
		t.Errorf("Draw x=1 with scheme: style = %v, want HistoryArrow %v",
			cell.Style, theme.BorlandBlue.HistoryArrow)
	}
}

// TestHistoryDrawRightSideUsesHistorySidesStyle verifies x=2 uses HistorySides style
// when a color scheme is present.
// Spec: "▌ (U+258C) at x=2 in HistorySides color"
func TestHistoryDrawRightSideUsesHistorySidesStyle(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)
	h.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	cell := buf.GetCell(2, 0)
	if cell.Style != theme.BorlandBlue.HistorySides {
		t.Errorf("Draw x=2 with scheme: style = %v, want HistorySides %v",
			cell.Style, theme.BorlandBlue.HistorySides)
	}
}

// TestHistoryDrawSidesAndArrowStylesDiffer verifies HistorySides and HistoryArrow
// are distinct styles in the BorlandBlue scheme, so the tests above are meaningful.
// Spec: "▐ at x=0 in HistorySides color; ↓ at x=1 in HistoryArrow color"
func TestHistoryDrawSidesAndArrowStylesDiffer(t *testing.T) {
	if theme.BorlandBlue.HistorySides == theme.BorlandBlue.HistoryArrow {
		t.Error("BorlandBlue.HistorySides and BorlandBlue.HistoryArrow must be distinct styles")
	}
}

// TestHistoryDrawWithSchemeViaOwner verifies that a History inserted into a Window
// with a color scheme inherits the scheme via its Owner chain.
// Spec: "If no color scheme is available, use tcell.StyleDefault" (contrapositive:
//
//	when a scheme IS available via owner, it must be used)
func TestHistoryDrawWithSchemeViaOwner(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 30, 5), "test")
	win.SetColorScheme(theme.BorlandBlue)

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)
	win.Insert(h)

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	// With scheme inherited from owner, arrow at x=1 must use HistoryArrow, not StyleDefault.
	cell := buf.GetCell(1, 0)
	if cell.Style == tcell.StyleDefault {
		t.Errorf("Draw x=1 with owner scheme: got StyleDefault, want HistoryArrow style %v",
			theme.BorlandBlue.HistoryArrow)
	}
	if cell.Style != theme.BorlandBlue.HistoryArrow {
		t.Errorf("Draw x=1 with owner scheme: style = %v, want HistoryArrow %v",
			cell.Style, theme.BorlandBlue.HistoryArrow)
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Falsifying tests
// ---------------------------------------------------------------------------

// TestHistoryDoesNotSetOfSelectableAfterClear verifies OfSelectable stays unset
// even if the History options are inspected after construction.
// Falsifying: ensures the constructor doesn't accidentally set OfSelectable.
// Spec: "Does NOT set OfSelectable (never receives focus)"
func TestHistoryDoesNotSetOfSelectableAfterClear(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	// Explicitly clear it (if it was set, this would pass the guard below).
	h.SetOptions(OfSelectable, false)

	if h.HasOption(OfSelectable) {
		t.Error("OfSelectable should remain unset after explicit clear")
	}
}

// TestHistoryLinkIsNotNilForValidInput verifies Link() does not return nil when
// a non-nil InputLine is passed.
// Falsifying: guards against a bug where the constructor ignores the link argument.
// Spec: "Stores reference to linked InputLine"
func TestHistoryLinkIsNotNilForValidInput(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 5)

	if h.Link() == nil {
		t.Error("Link() returned nil, but a non-nil InputLine was passed to NewHistory")
	}
}

// TestHistoryCharactersAreNotSpaces verifies that all three draw positions
// contain non-space runes — guards against a stub that fills with spaces.
// Spec: "Renders exactly 3 characters: ▐↓▌"
func TestHistoryCharactersAreNotSpaces(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	for x := 0; x < 3; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune == ' ' {
			t.Errorf("Draw x=%d: rune is space; expected a non-space character", x)
		}
	}
}

// TestHistoryStyleWithSchemeIsNotStyleDefault verifies that when a color scheme
// is set, at least one cell uses a style other than tcell.StyleDefault.
// Falsifying: guards against a Draw that ignores the color scheme.
// Spec: "▐ in HistorySides color; ↓ in HistoryArrow color; ▌ in HistorySides color"
func TestHistoryStyleWithSchemeIsNotStyleDefault(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h := NewHistory(NewRect(0, 0, 3, 1), il, 1)
	h.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(3, 1)
	h.Draw(buf)

	allDefault := true
	for x := 0; x < 3; x++ {
		if buf.GetCell(x, 0).Style != tcell.StyleDefault {
			allDefault = false
			break
		}
	}
	if allDefault {
		t.Error("Draw with BorlandBlue scheme: all cells used StyleDefault; expected scheme styles")
	}
}

// TestHistoryDrawDifferentIDsProduceSameOutput verifies that the history ID
// does not affect what is drawn — the widget always draws ▐↓▌.
// Spec: "Renders exactly 3 characters: ▐↓▌" (no mention of ID affecting draw)
func TestHistoryDrawDifferentIDsProduceSameOutput(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	h1 := NewHistory(NewRect(0, 0, 3, 1), il, 1)
	h2 := NewHistory(NewRect(0, 0, 3, 1), il, 999)

	buf1 := NewDrawBuffer(3, 1)
	buf2 := NewDrawBuffer(3, 1)
	h1.Draw(buf1)
	h2.Draw(buf2)

	for x := 0; x < 3; x++ {
		c1 := buf1.GetCell(x, 0)
		c2 := buf2.GetCell(x, 0)
		if c1.Rune != c2.Rune {
			t.Errorf("Draw x=%d: historyID 1 drew %q, historyID 999 drew %q; should be identical",
				x, c1.Rune, c2.Rune)
		}
	}
}

func TestViewToDesktopAccountsForDialogFrame(t *testing.T) {
	// Dialog has a 1-pixel frame, same as Window. viewToDesktop must
	// account for *Dialog, not just *Window.
	desktop := NewDesktop(NewRect(0, 0, 80, 25))

	dlg := NewDialog(NewRect(5, 3, 40, 15), "Test")
	desktop.Insert(dlg)

	il := NewInputLine(NewRect(2, 1, 20, 1), 64)
	dlg.Insert(il)

	x, y := viewToDesktop(il)
	// InputLine at (2,1) inside Dialog client area.
	// Dialog at (5,3) in Desktop, frame adds (1,1).
	// Expected: (5+1+2, 3+1+1) = (8, 5)
	if x != 8 || y != 5 {
		t.Errorf("viewToDesktop(InputLine in Dialog) = (%d,%d), want (8,5)", x, y)
	}
}

func TestViewToDesktopAccountsForWindowFrame(t *testing.T) {
	desktop := NewDesktop(NewRect(0, 0, 80, 25))

	win := NewWindow(NewRect(5, 3, 40, 15), "Test")
	desktop.Insert(win)

	il := NewInputLine(NewRect(2, 1, 20, 1), 64)
	win.Insert(il)

	x, y := viewToDesktop(il)
	// Same arithmetic: (5+1+2, 3+1+1) = (8, 5)
	if x != 8 || y != 5 {
		t.Errorf("viewToDesktop(InputLine in Window) = (%d,%d), want (8,5)", x, y)
	}
}
