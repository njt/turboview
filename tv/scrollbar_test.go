package tv

// scrollbar_test.go — Tests for Task 1: ScrollBar widget.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Constructor, state flags, Widget interface
//   Section 2  — State accessors (SetRange, SetValue, SetPageSize, Value, Min, Max, PageSize)
//   Section 3  — Drawing: Vertical
//   Section 4  — Drawing: Horizontal
//   Section 5  — Mouse handling (arrows, track, wheel)
//   Section 6  — Falsifying tests (guard against vacuous passes)

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// compile-time assertion: ScrollBar must satisfy Widget.
// Spec: "Implements the Widget interface"
var _ Widget = (*ScrollBar)(nil)

// ---------------------------------------------------------------------------
// Section 1 — Constructor and state flags
// ---------------------------------------------------------------------------

// TestNewScrollBarSetsSfVisible verifies NewScrollBar sets the SfVisible state flag.
// Spec: "Sets SfVisible by default."
func TestNewScrollBarSetsSfVisible(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)

	if !sb.HasState(SfVisible) {
		t.Error("NewScrollBar did not set SfVisible")
	}
}

// TestNewScrollBarDoesNotSetOfSelectable verifies NewScrollBar sets OfSelectable.
// Spec: "SetOptions(OfSelectable, true) in NewScrollBar so scrollbar can receive focus"
func TestNewScrollBarDoesNotSetOfSelectable(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)

	if !sb.HasOption(OfSelectable) {
		t.Error("NewScrollBar must set OfSelectable so the scrollbar can receive focus")
	}
}

// TestNewScrollBarStoresBounds verifies NewScrollBar records the given bounds.
// Spec: "NewScrollBar(bounds Rect, orientation Orientation) *ScrollBar"
func TestNewScrollBarStoresBounds(t *testing.T) {
	r := NewRect(5, 3, 1, 12)
	sb := NewScrollBar(r, Vertical)

	if sb.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", sb.Bounds(), r)
	}
}

// TestNewScrollBarVerticalOrientation verifies the Vertical orientation constant exists
// and that a vertical scrollbar can be created.
// Spec: "Orientation is a new type: const (Horizontal Orientation = iota; Vertical)"
func TestNewScrollBarVerticalOrientation(t *testing.T) {
	// Vertical must equal 1 (iota starting at Horizontal=0).
	if Vertical != 1 {
		t.Errorf("Vertical = %d, want 1 (Horizontal=0, Vertical=1 via iota)", int(Vertical))
	}
}

// TestNewScrollBarHorizontalOrientation verifies Horizontal is the zero value of Orientation.
// Spec: "const (Horizontal Orientation = iota; Vertical)"
func TestNewScrollBarHorizontalOrientation(t *testing.T) {
	if Horizontal != 0 {
		t.Errorf("Horizontal = %d, want 0 (iota base)", int(Horizontal))
	}
}

// TestHorizontalAndVerticalOrientationsDiffer verifies the two orientation constants
// are distinct (falsification guard).
// Spec: "const (Horizontal Orientation = iota; Vertical)"
func TestHorizontalAndVerticalOrientationsDiffer(t *testing.T) {
	if Horizontal == Vertical {
		t.Error("Horizontal and Vertical must be different Orientation values")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — State accessors
// ---------------------------------------------------------------------------

// TestScrollBarDefaultMinIsZero verifies Min() returns 0 on a new ScrollBar.
// Spec: "min, max, value, pageSize int — the scrollable range"
func TestScrollBarDefaultMinIsZero(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)

	if sb.Min() != 0 {
		t.Errorf("Min() = %d, want 0 (default)", sb.Min())
	}
}

// TestScrollBarDefaultMaxIsZero verifies Max() returns 0 on a new ScrollBar.
// Spec: "min, max, value, pageSize int"
func TestScrollBarDefaultMaxIsZero(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)

	if sb.Max() != 0 {
		t.Errorf("Max() = %d, want 0 (default)", sb.Max())
	}
}

// TestScrollBarDefaultValueIsZero verifies Value() returns 0 on a new ScrollBar.
// Spec: "min, max, value, pageSize int"
func TestScrollBarDefaultValueIsZero(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)

	if sb.Value() != 0 {
		t.Errorf("Value() = %d, want 0 (default)", sb.Value())
	}
}

// TestScrollBarDefaultPageSizeIsZero verifies PageSize() returns 0 on a new ScrollBar.
// Spec: "min, max, value, pageSize int"
func TestScrollBarDefaultPageSizeIsZero(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)

	if sb.PageSize() != 0 {
		t.Errorf("PageSize() = %d, want 0 (default)", sb.PageSize())
	}
}

// TestScrollBarSetRangeStoresMin verifies SetRange sets Min().
// Spec: "SetRange(min, max int) sets the scrollable range"
func TestScrollBarSetRangeStoresMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(5, 100)

	if sb.Min() != 5 {
		t.Errorf("Min() = %d, want 5 after SetRange(5, 100)", sb.Min())
	}
}

// TestScrollBarSetRangeStoresMax verifies SetRange sets Max().
// Spec: "SetRange(min, max int) sets the scrollable range"
func TestScrollBarSetRangeStoresMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(5, 100)

	if sb.Max() != 100 {
		t.Errorf("Max() = %d, want 100 after SetRange(5, 100)", sb.Max())
	}
}

// TestScrollBarSetValueStoresValue verifies SetValue stores the given value.
// Spec: "SetValue(v int) sets current scroll position (clamped to [min, max-pageSize])"
func TestScrollBarSetValueStoresValue(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	if sb.Value() != 50 {
		t.Errorf("Value() = %d, want 50", sb.Value())
	}
}

// TestScrollBarSetValueClampsToMin verifies SetValue clamps values below min.
// Spec: "SetValue(v int) sets current scroll position (clamped to [min, max-pageSize])"
func TestScrollBarSetValueClampsToMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(10, 100)
	sb.SetPageSize(5)
	sb.SetValue(3) // below min=10

	if sb.Value() != 10 {
		t.Errorf("Value() = %d, want 10 (clamped to min)", sb.Value())
	}
}

// TestScrollBarSetValueClampsToMaxMinusPageSize verifies SetValue clamps values
// above max-pageSize.
// Spec: "clamped to [min, max-pageSize]"
func TestScrollBarSetValueClampsToMaxMinusPageSize(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(200) // above max-pageSize = 90

	if sb.Value() != 90 {
		t.Errorf("Value() = %d, want 90 (clamped to max-pageSize = 100-10)", sb.Value())
	}
}

// TestScrollBarSetValueAtMaxMinusPageSize verifies SetValue accepts the exact
// upper bound (max-pageSize).
// Spec: "clamped to [min, max-pageSize]"
func TestScrollBarSetValueAtMaxMinusPageSize(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90) // exactly max-pageSize

	if sb.Value() != 90 {
		t.Errorf("Value() = %d, want 90 (exact upper bound)", sb.Value())
	}
}

// TestScrollBarSetPageSizeStoresPageSize verifies SetPageSize stores the page size.
// Spec: "SetPageSize(n int) sets the visible page size"
func TestScrollBarSetPageSizeStoresPageSize(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetPageSize(15)

	if sb.PageSize() != 15 {
		t.Errorf("PageSize() = %d, want 15", sb.PageSize())
	}
}

// TestScrollBarValueReturnsCurrentValue verifies Value() returns what was set.
// Spec: "Value() int returns current value"
func TestScrollBarValueReturnsCurrentValue(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 200)
	sb.SetPageSize(20)
	sb.SetValue(75)

	if sb.Value() != 75 {
		t.Errorf("Value() = %d, want 75", sb.Value())
	}
}

// TestScrollBarMinReturnsMin verifies Min() returns the min set by SetRange.
// Spec: "Min() int … return current settings"
func TestScrollBarMinReturnsMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(20, 80)

	if sb.Min() != 20 {
		t.Errorf("Min() = %d, want 20", sb.Min())
	}
}

// TestScrollBarMaxReturnsMax verifies Max() returns the max set by SetRange.
// Spec: "Max() int … return current settings"
func TestScrollBarMaxReturnsMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(20, 80)

	if sb.Max() != 80 {
		t.Errorf("Max() = %d, want 80", sb.Max())
	}
}

// TestScrollBarPageSizeReturnsPageSize verifies PageSize() returns what was set.
// Spec: "PageSize() int return current settings"
func TestScrollBarPageSizeReturnsPageSize(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetPageSize(25)

	if sb.PageSize() != 25 {
		t.Errorf("PageSize() = %d, want 25", sb.PageSize())
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Drawing: Vertical scrollbar
// ---------------------------------------------------------------------------

// TestScrollBarVerticalDrawsUpArrowAtRow0 verifies the up arrow '▲' is at row 0.
// Spec: "Row 0: up arrow ▲ in ScrollBar style"
func TestScrollBarVerticalDrawsUpArrowAtRow0(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != '▲' {
		t.Errorf("vertical scrollbar row 0 = %q, want '▲' (up arrow)", cell.Rune)
	}
}

// TestScrollBarVerticalUpArrowUsesScrollBarStyle verifies the up arrow uses ScrollBar style.
// Spec: "Row 0: up arrow ▲ in ScrollBar style"
func TestScrollBarVerticalUpArrowUsesScrollBarStyle(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != theme.BorlandBlue.ScrollBar {
		t.Errorf("up arrow style = %v, want ScrollBar %v", cell.Style, theme.BorlandBlue.ScrollBar)
	}
}

// TestScrollBarVerticalDrawsDownArrowAtLastRow verifies the down arrow '▼' is at the last row.
// Spec: "Last row: down arrow ▼ in ScrollBar style"
func TestScrollBarVerticalDrawsDownArrowAtLastRow(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	cell := buf.GetCell(0, 9) // last row (height-1 = 9)
	if cell.Rune != '▼' {
		t.Errorf("vertical scrollbar last row = %q, want '▼' (down arrow)", cell.Rune)
	}
}

// TestScrollBarVerticalDownArrowUsesScrollBarStyle verifies the down arrow uses ScrollBar style.
// Spec: "Last row: down arrow ▼ in ScrollBar style"
func TestScrollBarVerticalDownArrowUsesScrollBarStyle(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	cell := buf.GetCell(0, 9)
	if cell.Style != theme.BorlandBlue.ScrollBar {
		t.Errorf("down arrow style = %v, want ScrollBar %v", cell.Style, theme.BorlandBlue.ScrollBar)
	}
}

// TestScrollBarVerticalTrackFilledWithTrackChar verifies track rows 1..height-2
// are filled with '░' in ScrollBar style.
// Spec: "Rows 1 to height-2: track area filled with ░ in ScrollBar style"
func TestScrollBarVerticalTrackFilledWithTrackChar(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue
	// pageSize >= (max-min): thumb fills entire track, so no thumb overlaps track
	// Actually thumb fills track but is drawn as '█'. Use max=min so thumb fills all.
	// We need to distinguish track char from thumb char; use a case where thumb is visible.
	// Set range such that no thumb obstructs checking the track chars.
	// Easier: set max==min so spec says thumb fills entire track = all '█'.
	// Instead use a range with thumb at one position and check non-thumb rows.
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0) // thumb at top of track
	// trackLen = 10-2 = 8. thumbLen = max(1, 8*10/100) = max(1, 0) = 1.
	// thumbPos = 0 * (8-1) / (100-10) = 0. thumb at row 1.
	// Track chars (non-thumb): rows 2..8 should be '░'

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	for row := 2; row <= 8; row++ {
		cell := buf.GetCell(0, row)
		if cell.Rune != '░' {
			t.Errorf("vertical track at row %d = %q, want '░'", row, cell.Rune)
		}
	}
}

// TestScrollBarVerticalTrackUsesScrollBarStyle verifies the track area uses ScrollBar style.
// Spec: "track area filled with ░ in ScrollBar style"
func TestScrollBarVerticalTrackUsesScrollBarStyle(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)
	// As above, rows 2..8 are track (not thumb)

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	for row := 2; row <= 8; row++ {
		cell := buf.GetCell(0, row)
		if cell.Style != theme.BorlandBlue.ScrollBar {
			t.Errorf("track at row %d style = %v, want ScrollBar %v", row, cell.Style, theme.BorlandBlue.ScrollBar)
			break
		}
	}
}

// TestScrollBarVerticalThumbUsesScrollThumbStyle verifies the thumb uses ScrollThumb style.
// Spec: "Thumb position within track area rendered as █ in ScrollThumb style"
func TestScrollBarVerticalThumbUsesScrollThumbStyle(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)
	// thumbPos=0, thumb at row 1

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	cell := buf.GetCell(0, 1)
	if cell.Style != theme.BorlandBlue.ScrollThumb {
		t.Errorf("thumb at row 1 style = %v, want ScrollThumb %v", cell.Style, theme.BorlandBlue.ScrollThumb)
	}
}

// TestScrollBarVerticalThumbCharIsBlock verifies the thumb character is '█'.
// Spec: "rendered as █ in ScrollThumb style"
func TestScrollBarVerticalThumbCharIsBlock(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)
	// thumb at row 1

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	cell := buf.GetCell(0, 1)
	if cell.Rune != '█' {
		t.Errorf("thumb char = %q, want '█'", cell.Rune)
	}
}

// TestScrollBarVerticalThumbFillsTrackWhenMaxEqualsMin verifies the thumb fills
// the entire track when max <= min.
// Spec: "if max <= min or pageSize >= (max-min), thumb fills entire track."
func TestScrollBarVerticalThumbFillsTrackWhenMaxEqualsMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(50, 50) // max == min
	sb.SetPageSize(0)
	sb.SetValue(50)

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	// Track = rows 1..8 (trackLen=8). All should be '█' (thumb fills entire track).
	for row := 1; row <= 8; row++ {
		cell := buf.GetCell(0, row)
		if cell.Rune != '█' {
			t.Errorf("max==min: track row %d = %q, want '█' (thumb fills entire track)", row, cell.Rune)
		}
	}
}

// TestScrollBarVerticalThumbFillsTrackWhenPageSizeGERange verifies the thumb fills
// the entire track when pageSize >= (max-min).
// Spec: "if max <= min or pageSize >= (max-min), thumb fills entire track."
func TestScrollBarVerticalThumbFillsTrackWhenPageSizeGERange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 20)
	sb.SetPageSize(20) // pageSize == max-min
	sb.SetValue(0)

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	// Track = rows 1..8. All should be '█'.
	for row := 1; row <= 8; row++ {
		cell := buf.GetCell(0, row)
		if cell.Rune != '█' {
			t.Errorf("pageSize==range: track row %d = %q, want '█' (thumb fills entire track)", row, cell.Rune)
		}
	}
}

// TestScrollBarVerticalThumbPositionAtBottom verifies the thumb is at the bottom of
// the track when value == max-pageSize (maximum scroll position).
// Spec: "thumbPos = (value - min) * (trackLen - thumbLen) / (max - min - pageSize)"
func TestScrollBarVerticalThumbPositionAtBottom(t *testing.T) {
	// height=12: trackLen=10, thumbLen=max(1,10*10/100)=1
	// At value=90 (max-pageSize): thumbPos = 90*(10-1)/(100-10) = 90*9/90 = 9
	// thumb at row 1+9 = 10.
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90) // max-pageSize = 90

	buf := NewDrawBuffer(1, 12)
	sb.Draw(buf)

	// thumb should be at row 10 (last track row = trackLen = row 1+9 = 10)
	cell := buf.GetCell(0, 10)
	if cell.Rune != '█' {
		t.Errorf("thumb at max scroll: row 10 = %q, want '█'", cell.Rune)
	}
}

// TestScrollBarVerticalThumbPositionAtTop verifies the thumb is at row 1 when value == min.
// Spec: "thumbPos = (value - min) * (trackLen - thumbLen) / (max - min - pageSize) (clamped to [0, trackLen-thumbLen])"
func TestScrollBarVerticalThumbPositionAtTop(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0) // min: thumbPos=0, thumb at row 1

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	cell := buf.GetCell(0, 1)
	if cell.Rune != '█' {
		t.Errorf("thumb at min scroll: row 1 = %q, want '█'", cell.Rune)
	}
}

// TestScrollBarVerticalSmallWidgetNoContentBetweenArrows verifies that when
// trackLen < 1, nothing is drawn between the arrows.
// Spec: "When trackLen < 1 (widget too small), draw nothing between arrows"
func TestScrollBarVerticalSmallWidgetNoContentBetweenArrows(t *testing.T) {
	// height=2: arrows at rows 0 and 1, no track (trackLen = 2-2 = 0 < 1)
	sb := NewScrollBar(NewRect(0, 0, 1, 2), Vertical)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)

	buf := NewDrawBuffer(1, 2)
	sb.Draw(buf)

	// Row 0 = ▲, row 1 = ▼, no row in between.
	if buf.GetCell(0, 0).Rune != '▲' {
		t.Errorf("height=2: row 0 = %q, want '▲'", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(0, 1).Rune != '▼' {
		t.Errorf("height=2: row 1 = %q, want '▼'", buf.GetCell(0, 1).Rune)
	}
}

// TestScrollBarVerticalThumbLenIsAtLeast1 verifies thumbLen is at least 1
// even when the calculation yields 0.
// Spec: "thumbLen = max(1, trackLen * pageSize / (max - min))"
func TestScrollBarVerticalThumbLenIsAtLeast1(t *testing.T) {
	// pageSize=1, max-min=1000, trackLen=8: trackLen*pageSize/(max-min) = 8/1000 = 0
	// thumbLen = max(1, 0) = 1. So exactly one row should be '█'.
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 1000)
	sb.SetPageSize(1)
	sb.SetValue(0)

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	blockCount := 0
	for row := 1; row <= 8; row++ {
		if buf.GetCell(0, row).Rune == '█' {
			blockCount++
		}
	}
	if blockCount < 1 {
		t.Error("thumb len must be at least 1; no '█' found in track")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Drawing: Horizontal scrollbar
// ---------------------------------------------------------------------------

// TestScrollBarHorizontalDrawsLeftArrowAtCol0 verifies the left arrow '◄' is at col 0.
// Spec: "Col 0: left arrow ◄ in ScrollBar style"
func TestScrollBarHorizontalDrawsLeftArrowAtCol0(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(12, 1)
	sb.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != '◄' {
		t.Errorf("horizontal scrollbar col 0 = %q, want '◄' (left arrow)", cell.Rune)
	}
}

// TestScrollBarHorizontalLeftArrowUsesScrollBarStyle verifies the left arrow uses ScrollBar style.
// Spec: "Col 0: left arrow ◄ in ScrollBar style"
func TestScrollBarHorizontalLeftArrowUsesScrollBarStyle(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(12, 1)
	sb.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != theme.BorlandBlue.ScrollBar {
		t.Errorf("left arrow style = %v, want ScrollBar %v", cell.Style, theme.BorlandBlue.ScrollBar)
	}
}

// TestScrollBarHorizontalDrawsRightArrowAtLastCol verifies the right arrow '►' is at last col.
// Spec: "Last col: right arrow ► in ScrollBar style"
func TestScrollBarHorizontalDrawsRightArrowAtLastCol(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(12, 1)
	sb.Draw(buf)

	cell := buf.GetCell(11, 0) // last col = width-1 = 11
	if cell.Rune != '►' {
		t.Errorf("horizontal scrollbar last col = %q, want '►' (right arrow)", cell.Rune)
	}
}

// TestScrollBarHorizontalRightArrowUsesScrollBarStyle verifies the right arrow uses ScrollBar style.
// Spec: "Last col: right arrow ► in ScrollBar style"
func TestScrollBarHorizontalRightArrowUsesScrollBarStyle(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(12, 1)
	sb.Draw(buf)

	cell := buf.GetCell(11, 0)
	if cell.Style != theme.BorlandBlue.ScrollBar {
		t.Errorf("right arrow style = %v, want ScrollBar %v", cell.Style, theme.BorlandBlue.ScrollBar)
	}
}

// TestScrollBarHorizontalTrackFilledWithTrackChar verifies cols 1..width-2 contain '░'.
// Spec: "Cols 1 to width-2: track area with ░ in ScrollBar style"
func TestScrollBarHorizontalTrackFilledWithTrackChar(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)
	// width=12: trackLen=10. thumbLen=max(1,10*10/100)=1. thumbPos=0, thumb at col 1.
	// Track cols 2..10 should be '░'.

	buf := NewDrawBuffer(12, 1)
	sb.Draw(buf)

	for col := 2; col <= 10; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != '░' {
			t.Errorf("horizontal track at col %d = %q, want '░'", col, cell.Rune)
		}
	}
}

// TestScrollBarHorizontalThumbUsesScrollThumbStyle verifies the thumb uses ScrollThumb style.
// Spec: "Thumb rendered as █ in ScrollThumb style"
func TestScrollBarHorizontalThumbUsesScrollThumbStyle(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)
	// thumb at col 1

	buf := NewDrawBuffer(12, 1)
	sb.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Style != theme.BorlandBlue.ScrollThumb {
		t.Errorf("horizontal thumb at col 1 style = %v, want ScrollThumb %v", cell.Style, theme.BorlandBlue.ScrollThumb)
	}
}

// TestScrollBarHorizontalThumbFillsTrackWhenMaxEqualsMin verifies the thumb fills
// the entire track when max <= min.
// Spec: "if max <= min or pageSize >= (max-min), thumb fills entire track."
func TestScrollBarHorizontalThumbFillsTrackWhenMaxEqualsMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(5, 5) // max == min

	buf := NewDrawBuffer(12, 1)
	sb.Draw(buf)

	// Cols 1..10 should all be '█'.
	for col := 1; col <= 10; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != '█' {
			t.Errorf("horizontal max==min: col %d = %q, want '█' (thumb fills track)", col, cell.Rune)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Mouse handling
// ---------------------------------------------------------------------------

// TestScrollBarVerticalClickUpArrowDecrementsValue verifies clicking the up arrow
// decrements value by 1 and calls OnChange.
// Spec: "Click on up arrow (vertical) … decrement value by 1, clamp, call OnChange"
func TestScrollBarVerticalClickUpArrowDecrementsValue(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if sb.Value() != 49 {
		t.Errorf("after click up arrow: Value() = %d, want 49 (decremented by 1)", sb.Value())
	}
}

// TestScrollBarVerticalClickUpArrowCallsOnChange verifies OnChange is called when
// clicking the up arrow.
// Spec: "call OnChange"
func TestScrollBarVerticalClickUpArrowCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	called := false
	sb.OnChange = func(v int) { called = true }

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called after clicking up arrow")
	}
}

// TestScrollBarVerticalClickUpArrowPassesNewValueToOnChange verifies OnChange receives
// the new (post-change) value.
// Spec: "OnChange func(value int) — called when user interaction changes the value"
func TestScrollBarVerticalClickUpArrowPassesNewValueToOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	var got int
	sb.OnChange = func(v int) { got = v }

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if got != 49 {
		t.Errorf("OnChange received value %d, want 49", got)
	}
}

// TestScrollBarVerticalClickUpArrowClampsAtMin verifies clicking up arrow at min
// does not go below min.
// Spec: "decrement value by 1, clamp"
func TestScrollBarVerticalClickUpArrowClampsAtMin(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0) // already at min

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if sb.Value() != 0 {
		t.Errorf("up arrow at min: Value() = %d, want 0 (clamped at min)", sb.Value())
	}
}

// TestScrollBarVerticalClickDownArrowIncrementsValue verifies clicking the down arrow
// increments value by 1.
// Spec: "Click on down arrow … increment value by 1, clamp, call OnChange"
func TestScrollBarVerticalClickDownArrowIncrementsValue(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	// Down arrow is at row height-1 = 9.
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if sb.Value() != 51 {
		t.Errorf("after click down arrow: Value() = %d, want 51 (incremented by 1)", sb.Value())
	}
}

// TestScrollBarVerticalClickDownArrowCallsOnChange verifies OnChange is called when
// clicking the down arrow.
// Spec: "call OnChange"
func TestScrollBarVerticalClickDownArrowCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	called := false
	sb.OnChange = func(v int) { called = true }

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called after clicking down arrow")
	}
}

// TestScrollBarVerticalClickDownArrowClampsAtMax verifies clicking down arrow at max
// does not exceed max-pageSize.
// Spec: "increment value by 1, clamp"
func TestScrollBarVerticalClickDownArrowClampsAtMax(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90) // already at max-pageSize

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if sb.Value() != 90 {
		t.Errorf("down arrow at max: Value() = %d, want 90 (clamped at max-pageSize)", sb.Value())
	}
}

// TestScrollBarVerticalClickTrackAboveThumbDecrementsPageSize verifies clicking in
// the track area above the thumb decrements by pageSize.
// Spec: "Click in track area above/left of thumb: decrement value by pageSize, clamp, call OnChange"
func TestScrollBarVerticalClickTrackAboveThumbDecrementsPageSize(t *testing.T) {
	// height=12, trackLen=10, pageSize=10, range=0..100
	// thumbLen=max(1,10*10/100)=1
	// At value=50: thumbPos=(50-0)*(10-1)/(100-10)=50*9/90=5. thumb at row 1+5=6.
	// Click at row 3 (above thumb at row 6) → decrement by pageSize=10.
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 3, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if sb.Value() != 40 {
		t.Errorf("click track above thumb: Value() = %d, want 40 (50 - pageSize 10)", sb.Value())
	}
}

// TestScrollBarVerticalClickTrackAboveThumbCallsOnChange verifies OnChange is called
// when clicking track above thumb.
// Spec: "call OnChange"
func TestScrollBarVerticalClickTrackAboveThumbCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	called := false
	sb.OnChange = func(v int) { called = true }

	// Click above thumb (thumb at row 6, click at row 3).
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 3, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called after clicking track above thumb")
	}
}

// TestScrollBarVerticalClickTrackBelowThumbIncrementsPageSize verifies clicking in
// the track area below the thumb increments by pageSize.
// Spec: "Click in track area below/right of thumb: increment value by pageSize, clamp, call OnChange"
func TestScrollBarVerticalClickTrackBelowThumbIncrementsPageSize(t *testing.T) {
	// height=12, trackLen=10, pageSize=10, range=0..100
	// At value=50: thumbPos=5, thumb at row 6.
	// Click at row 9 (below thumb) → increment by pageSize=10.
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if sb.Value() != 60 {
		t.Errorf("click track below thumb: Value() = %d, want 60 (50 + pageSize 10)", sb.Value())
	}
}

// TestScrollBarVerticalClickTrackBelowThumbCallsOnChange verifies OnChange is called
// when clicking track below thumb.
// Spec: "call OnChange"
func TestScrollBarVerticalClickTrackBelowThumbCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	called := false
	sb.OnChange = func(v int) { called = true }

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 9, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called after clicking track below thumb")
	}
}

// TestScrollBarWheelUpDecrementsValue verifies WheelUp decrements value by 3*arStep.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelUpDecrementsValue(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 47 {
		t.Errorf("WheelUp: Value() = %d, want 47 (decremented by 3*arStep=3)", sb.Value())
	}
}

// TestScrollBarWheelUpCallsOnChange verifies OnChange is called on WheelUp.
// Spec: "Mouse wheel up (WheelUp): decrement value by 1, clamp, call OnChange"
func TestScrollBarWheelUpCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	called := false
	sb.OnChange = func(v int) { called = true }

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called on WheelUp")
	}
}

// TestScrollBarWheelDownIncrementsValue verifies WheelDown increments value by 3*arStep.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarWheelDownIncrementsValue(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 53 {
		t.Errorf("WheelDown: Value() = %d, want 53 (incremented by 3*arStep=3)", sb.Value())
	}
}

// TestScrollBarWheelDownCallsOnChange verifies OnChange is called on WheelDown.
// Spec: "Mouse wheel down (WheelDown): increment value by 1, clamp, call OnChange"
func TestScrollBarWheelDownCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	called := false
	sb.OnChange = func(v int) { called = true }

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 5, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called on WheelDown")
	}
}

// TestScrollBarMouseEventsConsumed verifies all mouse interactions consume the event.
// Spec: "All mouse interactions consume the event"
func TestScrollBarMouseEventsConsumed(t *testing.T) {
	tests := []struct {
		name   string
		button tcell.ButtonMask
		y      int
	}{
		{"up arrow", tcell.Button1, 0},
		{"down arrow", tcell.Button1, 9},
		{"track", tcell.Button1, 4},
		{"wheel up", tcell.WheelUp, 5},
		{"wheel down", tcell.WheelDown, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
			sb.SetRange(0, 100)
			sb.SetPageSize(10)
			sb.SetValue(50)

			ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: tt.y, Button: tt.button}}
			sb.HandleEvent(ev)

			if !ev.IsCleared() {
				t.Errorf("%s: event not consumed; ev.What = %v, want EvNothing", tt.name, ev.What)
			}
		})
	}
}

// TestScrollBarHorizontalClickLeftArrowDecrementsValue verifies clicking the left arrow
// on a horizontal scrollbar decrements value by 1.
// Spec: "Click on … left arrow (horizontal): decrement value by 1, clamp, call OnChange"
func TestScrollBarHorizontalClickLeftArrowDecrementsValue(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if sb.Value() != 49 {
		t.Errorf("horizontal left arrow: Value() = %d, want 49", sb.Value())
	}
}

// TestScrollBarHorizontalClickRightArrowIncrementsValue verifies clicking the right arrow
// on a horizontal scrollbar increments value by 1.
// Spec: "Click on … right arrow: increment value by 1, clamp, call OnChange"
func TestScrollBarHorizontalClickRightArrowIncrementsValue(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	// Right arrow is at col width-1 = 11.
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 11, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if sb.Value() != 51 {
		t.Errorf("horizontal right arrow: Value() = %d, want 51", sb.Value())
	}
}

// TestScrollBarHorizontalClickTrackLeftOfThumbDecrementsPageSize verifies clicking
// in the track left of the thumb decrements by pageSize.
// Spec: "Click in track area above/left of thumb: decrement value by pageSize, clamp, call OnChange"
func TestScrollBarHorizontalClickTrackLeftOfThumbDecrementsPageSize(t *testing.T) {
	// width=22, trackLen=20, pageSize=10, range=0..100
	// thumbLen=max(1,20*10/100)=2
	// At value=50: thumbPos=50*(20-2)/(100-10)=50*18/90=10. thumb at cols 1+10=11..12.
	// Click at col 4 (left of thumb at col 11) → decrement by pageSize=10.
	sb := NewScrollBar(NewRect(0, 0, 22, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 4, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if sb.Value() != 40 {
		t.Errorf("horizontal track left of thumb: Value() = %d, want 40", sb.Value())
	}
}

// TestScrollBarHorizontalClickTrackRightOfThumbIncrementsPageSize verifies clicking
// in the track right of the thumb increments by pageSize.
// Spec: "Click in track area below/right of thumb: increment value by pageSize, clamp, call OnChange"
func TestScrollBarHorizontalClickTrackRightOfThumbIncrementsPageSize(t *testing.T) {
	// width=22, trackLen=20, pageSize=10, range=0..100
	// At value=50: thumb at cols 11..12.
	// Click at col 18 (right of thumb) → increment by pageSize=10.
	sb := NewScrollBar(NewRect(0, 0, 22, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 18, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if sb.Value() != 60 {
		t.Errorf("horizontal track right of thumb: Value() = %d, want 60", sb.Value())
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Falsifying tests
// ---------------------------------------------------------------------------

// TestScrollBarScrollBarStyleDiffersFromScrollThumbStyle verifies ScrollBar and
// ScrollThumb styles are distinct in BorlandBlue (falsification guard for style tests).
func TestScrollBarScrollBarStyleDiffersFromScrollThumbStyle(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.ScrollBar == scheme.ScrollThumb {
		t.Skip("ScrollBar equals ScrollThumb in this scheme — style distinction test is vacuous")
	}
	// If styles are distinct, proceed to verify track vs thumb render differently.
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = scheme
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)
	// thumb at row 1; track at rows 2..8

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	thumbCell := buf.GetCell(0, 1)  // thumb
	trackCell := buf.GetCell(0, 2)  // track

	if thumbCell.Style == trackCell.Style {
		t.Errorf("thumb and track share style %v; expected different styles (ScrollThumb vs ScrollBar)",
			thumbCell.Style)
	}
}

// TestScrollBarSetValueDoesNotExceedMaxMinusPageSize verifies the upper clamp boundary
// is max-pageSize, not max.
// Spec: "clamped to [min, max-pageSize]"
func TestScrollBarSetValueDoesNotExceedMaxMinusPageSize(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(100) // max, not max-pageSize

	if sb.Value() == 100 {
		t.Errorf("SetValue(100) with pageSize=10 and max=100: Value() = 100, want 90 (max-pageSize); upper clamp is max-pageSize, not max")
	}
	if sb.Value() != 90 {
		t.Errorf("SetValue(100): Value() = %d, want 90 (clamped to max-pageSize)", sb.Value())
	}
}

// TestScrollBarOnChangeNotCalledWhenValueUnchanged verifies OnChange is not called
// when an action doesn't change the value (e.g., arrow at boundary).
// Spec: "OnChange func(value int) — called when user interaction changes the value"
// Boundary: if already at min and up-arrow clicked, value does not change.
// The spec says OnChange is "called when user interaction changes the value" — at
// boundary, the value is already clamped so no change occurs.
func TestScrollBarOnChangeNotCalledWhenValueUnchanged(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0) // at min

	called := false
	sb.OnChange = func(v int) { called = true }

	// Click up arrow — value is already at min (0), should not change.
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if called {
		t.Error("OnChange was called even though value did not change (already at min); OnChange should only fire when value actually changes")
	}
}

// TestScrollBarUpArrowAndDownArrowRenderDifferentChars verifies up and down arrows
// are distinct characters.
// Spec: "up arrow ▲ … down arrow ▼"
func TestScrollBarUpArrowAndDownArrowRenderDifferentChars(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	up := buf.GetCell(0, 0).Rune
	down := buf.GetCell(0, 9).Rune

	if up == down {
		t.Errorf("up and down arrows render the same rune %q; expected '▲' and '▼'", up)
	}
}

// TestScrollBarThumbMovesWithValue verifies the thumb position changes as value changes.
// Spec: "thumbPos = (value - min) * (trackLen - thumbLen) / (max - min - pageSize)"
func TestScrollBarThumbMovesWithValue(t *testing.T) {
	// A scrollbar where we can verify thumb at two distinct positions.
	// height=12, trackLen=10, pageSize=10, range=0..100
	// thumbLen=max(1,10*10/100)=1
	// At value=0: thumbPos=0, thumb at row 1.
	// At value=90: thumbPos=9, thumb at row 10.
	sbA := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sbA.scheme = theme.BorlandBlue
	sbA.SetRange(0, 100)
	sbA.SetPageSize(10)
	sbA.SetValue(0)

	sbB := NewScrollBar(NewRect(0, 0, 1, 12), Vertical)
	sbB.scheme = theme.BorlandBlue
	sbB.SetRange(0, 100)
	sbB.SetPageSize(10)
	sbB.SetValue(90)

	bufA := NewDrawBuffer(1, 12)
	bufB := NewDrawBuffer(1, 12)
	sbA.Draw(bufA)
	sbB.Draw(bufB)

	// At value=0 thumb at row 1, at value=90 thumb at row 10.
	// Row 1 should be thumb in A but not in B; row 10 should be thumb in B but not in A.
	if bufA.GetCell(0, 1).Rune != '█' {
		t.Errorf("value=0: thumb not at row 1 (got %q); thumb should be at top", bufA.GetCell(0, 1).Rune)
	}
	if bufA.GetCell(0, 10).Rune == '█' {
		t.Error("value=0: thumb should NOT be at row 10 (bottom); it must move with value")
	}
	if bufB.GetCell(0, 10).Rune != '█' {
		t.Errorf("value=90: thumb not at row 10 (got %q); thumb should be at bottom", bufB.GetCell(0, 10).Rune)
	}
	if bufB.GetCell(0, 1).Rune == '█' {
		t.Error("value=90: thumb should NOT be at row 1 (top); it must move with value")
	}
}

// TestScrollBarDrawNoColorSchemeDoesNotPanic verifies Draw is safe when no ColorScheme
// has been set.
func TestScrollBarDrawNoColorSchemeDoesNotPanic(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	buf := NewDrawBuffer(1, 10)

	// Must not panic.
	sb.Draw(buf)
}

// TestScrollBarSetRangeIndependentOfSetPageSize verifies Min and Max are independent
// of PageSize (falsification guard against conflating them).
// Spec: "SetRange(min, max int) … SetPageSize(n int)"
func TestScrollBarSetRangeIndependentOfSetPageSize(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(10, 90)
	sb.SetPageSize(20)

	if sb.Min() != 10 {
		t.Errorf("Min() = %d, want 10 (SetPageSize must not affect Min)", sb.Min())
	}
	if sb.Max() != 90 {
		t.Errorf("Max() = %d, want 90 (SetPageSize must not affect Max)", sb.Max())
	}
	if sb.PageSize() != 20 {
		t.Errorf("PageSize() = %d, want 20", sb.PageSize())
	}
}

// ---------------------------------------------------------------------------
// Integration test — ScrollBar inside a Group (using real framework owner chain)
// ---------------------------------------------------------------------------

// TestScrollBarIntegrationColorSchemeFromOwner verifies that when a ScrollBar is
// inserted into a Group that has a scheme, the ScrollBar's ColorScheme() resolves
// via the owner chain.
// Spec: "Implements the Widget interface" — which delegates ColorScheme() to owner.
func TestScrollBarIntegrationColorSchemeFromOwner(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	g.scheme = theme.BorlandBlue

	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	g.Insert(sb)

	cs := sb.ColorScheme()
	if cs == nil {
		t.Fatal("ColorScheme() returned nil; should resolve to owner's scheme via owner chain")
	}
	if cs != theme.BorlandBlue {
		t.Errorf("ColorScheme() = %v, want theme.BorlandBlue", cs)
	}
}

// TestScrollBarIntegrationDrawWithOwnerScheme verifies a ScrollBar without a direct
// scheme draws using the owner's scheme (arrows appear with non-default style).
// Spec: "Implements the Widget interface" — ColorScheme delegates to owner.
func TestScrollBarIntegrationDrawWithOwnerScheme(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	g.scheme = theme.BorlandBlue

	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	g.Insert(sb)

	buf := NewDrawBuffer(1, 10)
	sb.Draw(buf)

	// Up arrow should use ScrollBar style from owner's scheme.
	cell := buf.GetCell(0, 0)
	if cell.Style != theme.BorlandBlue.ScrollBar {
		t.Errorf("integration: up arrow style = %v, want ScrollBar %v from owner scheme",
			cell.Style, theme.BorlandBlue.ScrollBar)
	}
}

// ---------------------------------------------------------------------------
// Section 4 (continued) — Drawing: Horizontal scrollbar — additional tests
// ---------------------------------------------------------------------------

// TestScrollBarHorizontalTrackUsesScrollBarStyle verifies track cols 1..width-2
// use the ScrollBar style.
// Spec: "Cols 1 to width-2: track area with ░ in ScrollBar style"
func TestScrollBarHorizontalTrackUsesScrollBarStyle(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)
	// width=12: trackLen=10. thumbLen=max(1,10*10/100)=1. thumbPos=0, thumb at col 1.
	// Track cols 2..10 are non-thumb and should carry ScrollBar style.

	buf := NewDrawBuffer(12, 1)
	sb.Draw(buf)

	for col := 2; col <= 10; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != theme.BorlandBlue.ScrollBar {
			t.Errorf("horizontal track at col %d style = %v, want ScrollBar %v",
				col, cell.Style, theme.BorlandBlue.ScrollBar)
			break
		}
	}
}

// TestScrollBarHorizontalThumbCharIsBlock verifies thumb cells render as '█'.
// Spec: "Thumb rendered as █ in ScrollThumb style"
func TestScrollBarHorizontalThumbCharIsBlock(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)
	// thumb at col 1

	buf := NewDrawBuffer(12, 1)
	sb.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Rune != '█' {
		t.Errorf("horizontal thumb char = %q, want '█'", cell.Rune)
	}
}

// TestScrollBarHorizontalThumbPositionAtMinIsLeftmost verifies the thumb is at col 1
// (leftmost track position) when value == min.
// Spec: "thumbPos = (value - min) * (trackLen - thumbLen) / (max - min - pageSize)"
func TestScrollBarHorizontalThumbPositionAtMinIsLeftmost(t *testing.T) {
	// width=12: trackLen=10, thumbLen=max(1,10*10/100)=1.
	// At value=0 (min): thumbPos=0*(10-1)/(100-10)=0. Thumb at col 1+0=1.
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(0)

	buf := NewDrawBuffer(12, 1)
	sb.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Rune != '█' {
		t.Errorf("horizontal thumb at min: col 1 = %q, want '█' (leftmost position)", cell.Rune)
	}
}

// TestScrollBarHorizontalThumbPositionAtMaxIsRightmost verifies the thumb is at the
// rightmost track position when value == max-pageSize.
// Spec: "thumbPos = (value - min) * (trackLen - thumbLen) / (max - min - pageSize)"
func TestScrollBarHorizontalThumbPositionAtMaxIsRightmost(t *testing.T) {
	// width=12: trackLen=10, thumbLen=max(1,10*10/100)=1.
	// At value=90 (max-pageSize): thumbPos=90*(10-1)/(100-10)=90*9/90=9.
	// Thumb at col 1+9=10.
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.scheme = theme.BorlandBlue
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(90)

	buf := NewDrawBuffer(12, 1)
	sb.Draw(buf)

	cell := buf.GetCell(10, 0)
	if cell.Rune != '█' {
		t.Errorf("horizontal thumb at max: col 10 = %q, want '█' (rightmost position)", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Section 5 (continued) — Mouse handling: Horizontal scrollbar
// ---------------------------------------------------------------------------

// TestScrollBarHorizontalClickLeftArrowCallsOnChange verifies OnChange fires when
// clicking the left arrow of a horizontal scrollbar.
// Spec: "Click on … left arrow (horizontal): decrement value by 1, clamp, call OnChange"
func TestScrollBarHorizontalClickLeftArrowCallsOnChange(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	called := false
	sb.OnChange = func(v int) { called = true }

	// Left arrow is at col 0.
	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1}}
	sb.HandleEvent(ev)

	if !called {
		t.Error("OnChange was not called after clicking left arrow of horizontal scrollbar")
	}
}

// TestScrollBarHorizontalWheelUpDecrementsValue verifies WheelUp decrements value by 3*arStep
// on a horizontal scrollbar.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarHorizontalWheelUpDecrementsValue(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.WheelUp}}
	sb.HandleEvent(ev)

	if sb.Value() != 47 {
		t.Errorf("horizontal WheelUp: Value() = %d, want 47 (decremented by 3*arStep=3)", sb.Value())
	}
}

// TestScrollBarHorizontalWheelDownIncrementsValue verifies WheelDown increments value by 3*arStep
// on a horizontal scrollbar.
// Spec: "Mouse wheel should scroll by 3 * arStep per wheel tick"
func TestScrollBarHorizontalWheelDownIncrementsValue(t *testing.T) {
	sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)

	ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.WheelDown}}
	sb.HandleEvent(ev)

	if sb.Value() != 53 {
		t.Errorf("horizontal WheelDown: Value() = %d, want 53 (incremented by 3*arStep=3)", sb.Value())
	}
}

// TestScrollBarHorizontalMouseEventsConsumed verifies all horizontal mouse interactions
// consume the event.
// Spec: "All mouse interactions consume the event"
func TestScrollBarHorizontalMouseEventsConsumed(t *testing.T) {
	tests := []struct {
		name   string
		button tcell.ButtonMask
		x      int
	}{
		{"left arrow", tcell.Button1, 0},
		{"right arrow", tcell.Button1, 11},
		{"track", tcell.Button1, 5},
		{"wheel up", tcell.WheelUp, 5},
		{"wheel down", tcell.WheelDown, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewScrollBar(NewRect(0, 0, 12, 1), Horizontal)
			sb.SetRange(0, 100)
			sb.SetPageSize(10)
			sb.SetValue(50)

			ev := &Event{What: EvMouse, Mouse: &MouseEvent{X: tt.x, Y: 0, Button: tt.button}}
			sb.HandleEvent(ev)

			if !ev.IsCleared() {
				t.Errorf("%s: event not consumed; ev.What = %v, want EvNothing", tt.name, ev.What)
			}
		})
	}
}
