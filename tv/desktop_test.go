package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// TestNewDesktopSetsBounds verifies that NewDesktop stores the given bounds.
// Spec: "NewDesktop(bounds) creates a Desktop with ..."
func TestNewDesktopSetsBounds(t *testing.T) {
	r := NewRect(0, 0, 80, 25)
	d := NewDesktop(r)

	if d.Bounds() != r {
		t.Errorf("NewDesktop bounds = %v, want %v", d.Bounds(), r)
	}
}

// TestNewDesktopSetsSfVisible verifies that NewDesktop sets the SfVisible state flag.
// Spec: "NewDesktop(bounds) creates a Desktop with SfVisible state"
func TestNewDesktopSetsSfVisible(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))

	if !d.HasState(SfVisible) {
		t.Errorf("NewDesktop did not set SfVisible")
	}
}

// TestNewDesktopSetsGfGrowAll verifies that NewDesktop sets GfGrowAll grow mode.
// Spec: "NewDesktop(bounds) creates a Desktop with ... GfGrowAll grow mode"
func TestNewDesktopSetsGfGrowAll(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 80, 25))

	if d.GrowMode() != GfGrowAll {
		t.Errorf("NewDesktop grow mode = %v, want GfGrowAll (%v)", d.GrowMode(), GfGrowAll)
	}
}

// TestDesktopDrawFillsWithDesktopBackgroundStyle verifies that Draw uses the
// ColorScheme's DesktopBackground style when a scheme is set.
// Spec: "Desktop.Draw(buf) fills its area with the '░' character using ColorScheme().DesktopBackground"
func TestDesktopDrawFillsWithDesktopBackgroundStyle(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 5, 3))
	scheme := &theme.ColorScheme{
		DesktopBackground: tcell.StyleDefault.Foreground(tcell.ColorTeal).Background(tcell.ColorBlue),
	}
	d.scheme = scheme

	buf := NewDrawBuffer(5, 3)
	d.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.DesktopBackground {
		t.Errorf("Draw cell style = %v, want DesktopBackground %v", cell.Style, scheme.DesktopBackground)
	}
}

// TestDesktopDrawFillsWithStyleDefaultWhenNoScheme verifies that Draw uses
// tcell.StyleDefault when no ColorScheme is set.
// Spec: "If Desktop has no ColorScheme, Draw fills with '░' and tcell.StyleDefault"
func TestDesktopDrawFillsWithStyleDefaultWhenNoScheme(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 5, 3))
	// no scheme set

	buf := NewDrawBuffer(5, 3)
	d.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != tcell.StyleDefault {
		t.Errorf("Draw cell style without scheme = %v, want tcell.StyleDefault", cell.Style)
	}
}

// TestDesktopDrawFillsWithMediumShadeRune verifies that Draw uses the '░' rune.
// Spec: "Desktop.Draw(buf) fills its area with the '░' character"
func TestDesktopDrawFillsWithMediumShadeRune(t *testing.T) {
	d := NewDesktop(NewRect(0, 0, 5, 3))

	buf := NewDrawBuffer(5, 3)
	d.Draw(buf)

	cell := buf.GetCell(2, 1)
	if cell.Rune != '░' {
		t.Errorf("Draw cell rune = %q, want '░'", cell.Rune)
	}
}

// TestDesktopDrawFillsAllCellsInBounds verifies that Draw fills every cell
// within its bounds, not just a subset.
// Spec: "Desktop.Draw(buf) fills its area with the '░' character"
func TestDesktopDrawFillsAllCellsInBounds(t *testing.T) {
	w, h := 6, 4
	d := NewDesktop(NewRect(0, 0, w, h))

	buf := NewDrawBuffer(w, h)
	d.Draw(buf)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != '░' {
				t.Errorf("Draw cell (%d,%d) rune = %q, want '░'", x, y, cell.Rune)
			}
		}
	}
}

// TestDesktopDrawAllCellsUseSchemeStyle verifies that every cell in bounds
// has the DesktopBackground style when a scheme is set.
// Spec: "Desktop.Draw(buf) fills its area with the '░' character using ColorScheme().DesktopBackground"
func TestDesktopDrawAllCellsUseSchemeStyle(t *testing.T) {
	w, h := 4, 3
	d := NewDesktop(NewRect(0, 0, w, h))
	scheme := &theme.ColorScheme{
		DesktopBackground: tcell.StyleDefault.Foreground(tcell.ColorMaroon).Background(tcell.ColorNavy),
	}
	d.scheme = scheme

	buf := NewDrawBuffer(w, h)
	d.Draw(buf)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := buf.GetCell(x, y)
			if cell.Style != scheme.DesktopBackground {
				t.Errorf("Draw cell (%d,%d) style = %v, want DesktopBackground", x, y, cell.Style)
			}
		}
	}
}
