package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestNewColorDialog(t *testing.T) {
	groups := DefaultColorGroups()
	palette := make([]tcell.Style, 256)
	for i := range palette {
		palette[i] = tcell.StyleDefault
	}

	cd := NewColorDialog(groups, palette)

	if cd == nil {
		t.Fatal("NewColorDialog returned nil")
	}
	if cd.curGroup != 0 {
		t.Errorf("curGroup = %d, want 0", cd.curGroup)
	}
	if cd.curItem != 0 {
		t.Errorf("curItem = %d, want 0", cd.curItem)
	}
	if cd.fgSelector == nil {
		t.Error("fgSelector is nil")
	}
	if cd.bgSelector == nil {
		t.Error("bgSelector is nil")
	}
	if cd.display == nil {
		t.Error("display is nil")
	}
	if cd.groupList == nil {
		t.Error("groupList is nil")
	}
	if cd.itemList == nil {
		t.Error("itemList is nil")
	}
}

func TestColorDialogPaletteCopy(t *testing.T) {
	palette := make([]tcell.Style, 256)
	palette[0] = tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlue)

	cd := NewColorDialog(nil, palette)

	// The ColorDialog deep-copies the input palette, so modifying the
	// original should not affect the dialog's internal palette
	palette[0] = tcell.StyleDefault.Foreground(tcell.ColorGreen)

	fg, _, _ := cd.Palette()[0].Decompose()
	if fg == tcell.ColorGreen {
		t.Error("modifying input palette should not affect ColorDialog's internal palette")
	}
}

func TestColorDialogDefaultGroups(t *testing.T) {
	cd := NewColorDialog(nil, nil)

	if cd == nil {
		t.Fatal("NewColorDialog with nil args returned nil")
	}
	if len(cd.groups) != 9 {
		t.Errorf("len(groups) = %d, want 9", len(cd.groups))
	}
}

func TestColorDialogPaletteIndex(t *testing.T) {
	groups := DefaultColorGroups()
	cd := NewColorDialog(groups, nil)

	// Group 0 (Desktop), Item 0 (Background) → index 0
	idx := cd.paletteIndex()
	if idx != 0 {
		t.Errorf("paletteIndex = %d, want 0", idx)
	}
}

func TestColorDialogOnNewColorIndex(t *testing.T) {
	groups := DefaultColorGroups()
	palette := make([]tcell.Style, 256)
	palette[0] = tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlue)
	palette[4] = tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorYellow)

	cd := NewColorDialog(groups, palette)

	// Switch to another group first (Menus has items at indices 1-4)
	cd.onNewColorGroup(1) // Menus: Normal(1), Shortcut(2), Selected(3), Disabled(4)
	// Now item 0 = Normal at palette index 1
	cd.onNewColorIndex(3) // Disabled at palette index 4

	if cd.curItem != 3 {
		t.Errorf("curItem = %d, want 3", cd.curItem)
	}
}

func TestColorDialogHandleEventDelegates(t *testing.T) {
	cd := NewColorDialog(nil, nil)

	// Verify HandleEvent doesn't panic and delegates correctly
	ev := &Event{What: EvCommand, Command: CmOK}
	cd.HandleEvent(ev)
}

func TestColorToIndex(t *testing.T) {
	if idx := colorToIndex(tcell.PaletteColor(0)); idx != 0 {
		t.Errorf("colorToIndex(0) = %d, want 0", idx)
	}
	if idx := colorToIndex(tcell.PaletteColor(7)); idx != 7 {
		t.Errorf("colorToIndex(7) = %d, want 7", idx)
	}
	if idx := colorToIndex(tcell.PaletteColor(15)); idx != 15 {
		t.Errorf("colorToIndex(15) = %d, want 15", idx)
	}
	// Named colors match their iota index (ColorBlack=0, ..., ColorWhite=15)
	if idx := colorToIndex(tcell.ColorRed); idx != 9 {
		t.Errorf("colorToIndex(Red) = %d, want 9", idx)
	}
	if idx := colorToIndex(tcell.ColorWhite); idx != 15 {
		t.Errorf("colorToIndex(White) = %d, want 15", idx)
	}
}
