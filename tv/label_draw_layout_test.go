package tv

import (
	"testing"

	"github.com/njt/turboview/theme"
)

// helper: build a highlighted label via broadcast, following the spec's
// prescribed pattern (insert other first so focus transition broadcasts
// CmReceivedFocus when linked is inserted last).
func highlightedLabel(bounds Rect, text string, linked *labelLinkedView) (*Label, *Group) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	other := newLabelLinkedView()
	label := NewLabel(bounds, text, linked)
	label.scheme = theme.BorlandBlue
	g.Insert(other)   // other gets initial focus (no broadcast — from nil)
	g.Insert(label)
	g.Insert(linked)  // linked steals focus from other → broadcasts CmReceivedFocus
	return label, g
}

// --- Column 0 is always a space ---

// TestDrawCol0IsSpace verifies that column 0 always contains a space character.
// Spec: "Column 0 is always a space (the monochrome marker column — we don't
// implement the marker but the margin must exist for visual alignment with other
// dialog controls)."
func TestDrawCol0IsSpace(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	label.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Errorf("cell(0,0).Rune = %q, want ' '; column 0 must always be a space (margin column)", cell.Rune)
	}
}

// TestDrawCol0IsNotFirstTextChar falsifies an implementation that renders text
// starting at column 0.
// Spec: "Column 0 is always a space … Text rendering starts at column 1, not column 0."
func TestDrawCol0IsNotFirstTextChar(t *testing.T) {
	// "~N~ame": the first text rune is 'N'. If it appears at col 0 the margin is missing.
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	label.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == 'N' {
		t.Errorf("cell(0,0).Rune = 'N'; text must not start at column 0 — column 0 is the margin space")
	}
}

// --- Text starts at column 1 ---

// TestDrawTextStartsAtColumn1Shortcut verifies that the shortcut rune of a
// tilde-encoded label appears at column 1.
// Spec: "Text rendering starts at column 1, not column 0."
// Spec example: "Column 1: N, in LabelShortcut style"
func TestDrawTextStartsAtColumn1Shortcut(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	label.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Rune != 'N' {
		t.Errorf("cell(1,0).Rune = %q, want 'N'; first text rune must be at column 1", cell.Rune)
	}
}

// TestDrawTextStartsAtColumn1Normal verifies text without a shortcut also
// starts at column 1.
// Spec: "Text rendering starts at column 1, not column 0."
func TestDrawTextStartsAtColumn1Normal(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "Open", nil)
	label.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Rune != 'O' {
		t.Errorf("cell(1,0).Rune = %q, want 'O'; first text rune must be at column 1", cell.Rune)
	}
}

// --- Background fill covers full width ---

// TestDrawBackgroundFillAfterText verifies that columns beyond the rendered text
// are filled with spaces.
// Spec: "fills the entire widget width (column 0 to bounds width) with spaces …
// before rendering text. This ensures clean background when text is shorter than bounds."
// Spec example: "Columns 5-19: space, in normal/highlight style (background fill)"
func TestDrawBackgroundFillAfterText(t *testing.T) {
	// "~N~ame" → 4 text runes ('N','a','m','e') at cols 1-4; cols 5-19 must be spaces.
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	label.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	// Verify a sampling of trailing columns.
	for _, col := range []int{5, 10, 19} {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("cell(%d,0).Rune = %q, want ' '; trailing columns must be spaces (background fill)", col, cell.Rune)
		}
	}
}

// TestDrawBackgroundFillStyleNotDefault falsifies an implementation that leaves
// trailing cells in the zero/default style.
// Spec: "fills the entire widget width … with spaces in the current style
// (LabelNormal or LabelHighlight depending on l.light)"
func TestDrawBackgroundFillStyleNotDefault(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	scheme := theme.BorlandBlue
	label.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	// Column 19 is trailing background — must use LabelNormal, not zero/default style.
	cell := buf.GetCell(19, 0)
	if cell.Style == (struct{ Style interface{} }{}).Style {
		// unreachable — just a structural guard; real check below
	}
	// The trailing cell must NOT be tcell.StyleDefault (the zero value of a DrawBuffer).
	// BorlandBlue.LabelNormal is a distinct style.
	if cell.Style != scheme.LabelNormal {
		t.Errorf("cell(19,0) trailing background style = %v, want LabelNormal %v; background fill must use the label's current style", cell.Style, scheme.LabelNormal)
	}
}

// --- Normal state uses LabelNormal ---

// TestDrawNormalStateNonShortcutTextUsesLabelNormal verifies that when
// label.light == false, normal (non-shortcut) text segments use LabelNormal.
// Spec: "For a highlighted Label (l.light == true), the background fill and
// non-shortcut text use LabelHighlight style" — by contrast, when light == false,
// the style is LabelNormal.
func TestDrawNormalStateNonShortcutTextUsesLabelNormal(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	scheme := theme.BorlandBlue
	label.scheme = scheme
	// light defaults to false after construction

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	// "ame" starts at col 2 — a normal segment in normal state.
	cell := buf.GetCell(2, 0)
	if cell.Style != scheme.LabelNormal {
		t.Errorf("cell(2,0) non-shortcut text style = %v, want LabelNormal %v when light=false", cell.Style, scheme.LabelNormal)
	}
}

// TestDrawNormalStateBackgroundUsesLabelNormal verifies that trailing background
// cells use LabelNormal when light == false.
// Spec: "fills … with spaces in the current style (LabelNormal or LabelHighlight
// depending on l.light)"
func TestDrawNormalStateBackgroundUsesLabelNormal(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	scheme := theme.BorlandBlue
	label.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(10, 0)
	if cell.Style != scheme.LabelNormal {
		t.Errorf("cell(10,0) trailing background style = %v, want LabelNormal %v when light=false", cell.Style, scheme.LabelNormal)
	}
}

// --- Highlighted state uses LabelHighlight ---

// TestDrawHighlightedStateNonShortcutTextUsesLabelHighlight verifies that when
// the label is highlighted (light == true), non-shortcut text uses LabelHighlight.
// Spec: "For a highlighted Label (l.light == true), the background fill and
// non-shortcut text use LabelHighlight style"
func TestDrawHighlightedStateNonShortcutTextUsesLabelHighlight(t *testing.T) {
	linked := newLabelLinkedView()
	label, _ := highlightedLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	scheme := theme.BorlandBlue

	if !label.light {
		t.Fatal("precondition: label.light must be true after focus broadcast setup")
	}

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	// "ame" starts at col 2 — a normal segment. In highlight state it must use LabelHighlight.
	cell := buf.GetCell(2, 0)
	if cell.Style != scheme.LabelHighlight {
		t.Errorf("cell(2,0) non-shortcut text style = %v, want LabelHighlight %v when light=true", cell.Style, scheme.LabelHighlight)
	}
}

// TestDrawHighlightedStateBackgroundUsesLabelHighlight verifies that trailing
// background fill uses LabelHighlight when light == true.
// Spec: "For a highlighted Label (l.light == true), the background fill and
// non-shortcut text use LabelHighlight style"
func TestDrawHighlightedStateBackgroundUsesLabelHighlight(t *testing.T) {
	linked := newLabelLinkedView()
	label, _ := highlightedLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	scheme := theme.BorlandBlue

	if !label.light {
		t.Fatal("precondition: label.light must be true after focus broadcast setup")
	}

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	// Column 10 is trailing background — must use LabelHighlight when light=true.
	cell := buf.GetCell(10, 0)
	if cell.Style != scheme.LabelHighlight {
		t.Errorf("cell(10,0) trailing background style = %v, want LabelHighlight %v when light=true", cell.Style, scheme.LabelHighlight)
	}
}

// TestDrawHighlightedStateUsesLabelHighlightNotNormal falsifies an implementation
// that ignores the light flag and always uses LabelNormal for background.
// Spec: "For a highlighted Label (l.light == true), the background fill and
// non-shortcut text use LabelHighlight style"
func TestDrawHighlightedStateUsesLabelHighlightNotNormal(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.LabelHighlight == scheme.LabelNormal {
		t.Skip("LabelHighlight equals LabelNormal in BorlandBlue — distinction test is vacuous")
	}

	linked := newLabelLinkedView()
	label, _ := highlightedLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)

	if !label.light {
		t.Fatal("precondition: label.light must be true after focus broadcast setup")
	}

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	// A trailing background cell must use LabelHighlight, not LabelNormal.
	cell := buf.GetCell(10, 0)
	if cell.Style == scheme.LabelNormal {
		t.Errorf("cell(10,0) background style = LabelNormal; when light=true it must be LabelHighlight (they differ in BorlandBlue)")
	}
}

// --- Shortcut style unchanged by highlight ---

// TestDrawShortcutUsesLabelShortcutWhenNormal verifies LabelShortcut is used for
// the shortcut rune when light == false.
// Spec: "Shortcut segments always use LabelShortcut style regardless of highlight state"
func TestDrawShortcutUsesLabelShortcutWhenNormal(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	scheme := theme.BorlandBlue
	label.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	// 'N' is the shortcut rune, at col 1.
	cell := buf.GetCell(1, 0)
	if cell.Style != scheme.LabelShortcut {
		t.Errorf("cell(1,0) shortcut rune style = %v, want LabelShortcut %v when light=false", cell.Style, scheme.LabelShortcut)
	}
}

// TestDrawShortcutUsesLabelShortcutWhenHighlighted verifies LabelShortcut is used
// for the shortcut rune even when light == true.
// Spec: "Shortcut segments always use LabelShortcut style regardless of highlight state
// (matching original TV where shortcut palette entries 3 and 4 map to the same index)"
func TestDrawShortcutUsesLabelShortcutWhenHighlighted(t *testing.T) {
	linked := newLabelLinkedView()
	label, _ := highlightedLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	scheme := theme.BorlandBlue

	if !label.light {
		t.Fatal("precondition: label.light must be true after focus broadcast setup")
	}

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	// 'N' is the shortcut rune, at col 1 — must still be LabelShortcut.
	cell := buf.GetCell(1, 0)
	if cell.Style != scheme.LabelShortcut {
		t.Errorf("cell(1,0) shortcut rune style = %v, want LabelShortcut %v when light=true; shortcut style is unaffected by highlight state", cell.Style, scheme.LabelShortcut)
	}
}

// TestDrawShortcutStyleUnchangedByHighlightNotLabelHighlight falsifies an
// implementation that switches the shortcut style to LabelHighlight when light==true.
// Spec: "Shortcut segments always use LabelShortcut style regardless of highlight state"
func TestDrawShortcutStyleUnchangedByHighlightNotLabelHighlight(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.LabelShortcut == scheme.LabelHighlight {
		t.Skip("LabelShortcut equals LabelHighlight in BorlandBlue — distinction test is vacuous")
	}

	linked := newLabelLinkedView()
	label, _ := highlightedLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)

	if !label.light {
		t.Fatal("precondition: label.light must be true after focus broadcast setup")
	}

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	// Shortcut rune at col 1 must NOT be LabelHighlight even in highlighted state.
	cell := buf.GetCell(1, 0)
	if cell.Style == scheme.LabelHighlight {
		t.Errorf("cell(1,0) shortcut style = LabelHighlight; shortcut segments must keep LabelShortcut regardless of highlight state")
	}
}

// --- Column 0 style matches state ---

// TestDrawCol0StyleIsLabelNormalWhenNotHighlighted verifies that the column 0
// margin space uses LabelNormal style when light == false.
// Spec: "fills the entire widget width (column 0 to bounds width) with spaces in
// the current style (LabelNormal or LabelHighlight depending on l.light)"
func TestDrawCol0StyleIsLabelNormalWhenNotHighlighted(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	scheme := theme.BorlandBlue
	label.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.LabelNormal {
		t.Errorf("cell(0,0) margin space style = %v, want LabelNormal %v when light=false", cell.Style, scheme.LabelNormal)
	}
}

// TestDrawCol0StyleIsLabelHighlightWhenHighlighted verifies that the column 0
// margin space uses LabelHighlight style when light == true.
// Spec: "fills the entire widget width (column 0 to bounds width) with spaces in
// the current style (LabelNormal or LabelHighlight depending on l.light)"
func TestDrawCol0StyleIsLabelHighlightWhenHighlighted(t *testing.T) {
	linked := newLabelLinkedView()
	label, _ := highlightedLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	scheme := theme.BorlandBlue

	if !label.light {
		t.Fatal("precondition: label.light must be true after focus broadcast setup")
	}

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.LabelHighlight {
		t.Errorf("cell(0,0) margin space style = %v, want LabelHighlight %v when light=true", cell.Style, scheme.LabelHighlight)
	}
}
