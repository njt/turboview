package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// compile-time assertions: data sources must satisfy ListDataSource.
var _ ListDataSource = (*ColorGroupDataSource)(nil)
var _ ListDataSource = (*ColorItemDataSource)(nil)

// TestNewColorGroupDataSourceCount verifies NewColorGroupDataSource returns
// a data source with the correct count.
func TestNewColorGroupDataSourceCount(t *testing.T) {
	groups := []ColorGroup{
		{Name: "Desktop"},
		{Name: "Menus"},
		{Name: "Dialogs"},
	}
	ds := NewColorGroupDataSource(groups)
	if ds.Count() != 3 {
		t.Errorf("Count = %d, want 3", ds.Count())
	}
}

// TestNewColorGroupDataSourceItem0 verifies Item(0) returns the first group name.
func TestNewColorGroupDataSourceItem0(t *testing.T) {
	groups := []ColorGroup{
		{Name: "Desktop"},
		{Name: "Menus"},
	}
	ds := NewColorGroupDataSource(groups)
	if ds.Item(0) != "Desktop" {
		t.Errorf("Item(0) = %q, want %q", ds.Item(0), "Desktop")
	}
}

// TestNewColorGroupDataSourceItemNeg1 verifies Item(-1) returns "".
func TestNewColorGroupDataSourceItemNeg1(t *testing.T) {
	groups := []ColorGroup{{Name: "Desktop"}}
	ds := NewColorGroupDataSource(groups)
	if ds.Item(-1) != "" {
		t.Errorf("Item(-1) = %q, want empty string", ds.Item(-1))
	}
}

// TestNewColorGroupDataSourceItemOOB verifies out-of-bounds Item returns "".
func TestNewColorGroupDataSourceItemOOB(t *testing.T) {
	groups := []ColorGroup{{Name: "Desktop"}}
	ds := NewColorGroupDataSource(groups)
	if ds.Item(1) != "" {
		t.Errorf("Item(1) = %q, want empty string", ds.Item(1))
	}
}

// TestNewColorItemDataSourceCount verifies NewColorItemDataSource returns
// a data source with the correct count.
func TestNewColorItemDataSourceCount(t *testing.T) {
	items := []ColorItem{
		{Name: "Normal", Index: 1},
		{Name: "Shortcut", Index: 2},
		{Name: "Selected", Index: 3},
	}
	ds := NewColorItemDataSource(items)
	if ds.Count() != 3 {
		t.Errorf("Count = %d, want 3", ds.Count())
	}
}

// TestNewColorItemDataSourceItem0 verifies Item(0) returns the first item name.
func TestNewColorItemDataSourceItem0(t *testing.T) {
	items := []ColorItem{
		{Name: "Normal", Index: 1},
		{Name: "Shortcut", Index: 2},
	}
	ds := NewColorItemDataSource(items)
	if ds.Item(0) != "Normal" {
		t.Errorf("Item(0) = %q, want %q", ds.Item(0), "Normal")
	}
}

// TestNewColorItemDataSourceItemNeg1 verifies Item(-1) returns "".
func TestNewColorItemDataSourceItemNeg1(t *testing.T) {
	items := []ColorItem{{Name: "Normal", Index: 1}}
	ds := NewColorItemDataSource(items)
	if ds.Item(-1) != "" {
		t.Errorf("Item(-1) = %q, want empty string", ds.Item(-1))
	}
}

// TestColorGroupFields verifies ColorGroup has correct fields and
// lastFocused starts at 0 (zero value of int).
func TestColorGroupFields(t *testing.T) {
	cg := ColorGroup{Name: "Test", Items: []ColorItem{{"A", 1}}}
	if cg.Name != "Test" {
		t.Errorf("Name = %q, want %q", cg.Name, "Test")
	}
	if len(cg.Items) != 1 {
		t.Errorf("Items len = %d, want 1", len(cg.Items))
	}
	if cg.lastFocused != 0 {
		t.Errorf("lastFocused = %d, want 0", cg.lastFocused)
	}
}

// TestColorItemFields verifies ColorItem has correct fields.
func TestColorItemFields(t *testing.T) {
	ci := ColorItem{Name: "Background", Index: 5}
	if ci.Name != "Background" {
		t.Errorf("Name = %q, want %q", ci.Name, "Background")
	}
	if ci.Index != 5 {
		t.Errorf("Index = %d, want 5", ci.Index)
	}
}

// TestDefaultColorGroupsCount verifies DefaultColorGroups returns 9 groups.
func TestDefaultColorGroupsCount(t *testing.T) {
	groups := DefaultColorGroups()
	if len(groups) != 9 {
		t.Errorf("DefaultColorGroups len = %d, want 9", len(groups))
	}
}

// TestDefaultColorGroupsFirstGroup verifies first group is "Desktop" with
// 1 item (index 0).
func TestDefaultColorGroupsFirstGroup(t *testing.T) {
	groups := DefaultColorGroups()
	if groups[0].Name != "Desktop" {
		t.Errorf("groups[0].Name = %q, want %q", groups[0].Name, "Desktop")
	}
	if len(groups[0].Items) != 1 {
		t.Errorf("groups[0].Items len = %d, want 1", len(groups[0].Items))
	}
	if groups[0].Items[0].Name != "Background" {
		t.Errorf("groups[0].Items[0].Name = %q, want %q",
			groups[0].Items[0].Name, "Background")
	}
	if groups[0].Items[0].Index != 0 {
		t.Errorf("groups[0].Items[0].Index = %d, want 0",
			groups[0].Items[0].Index)
	}
}

// TestDefaultColorGroupsMenusHas4Items verifies the Menus group has 4 items.
func TestDefaultColorGroupsMenusHas4Items(t *testing.T) {
	groups := DefaultColorGroups()
	if len(groups[1].Items) != 4 {
		t.Errorf("Menus group items len = %d, want 4", len(groups[1].Items))
	}
}

// TestPaletteCopyIndependent verifies PaletteCopy creates an independent copy;
// modifying the copy does not affect the original.
func TestPaletteCopyIndependent(t *testing.T) {
	original := []tcell.Style{
		tcell.StyleDefault.Foreground(tcell.ColorRed),
		tcell.StyleDefault.Foreground(tcell.ColorGreen),
	}
	copied := PaletteCopy(original)
	// Modify copy
	copied[0] = tcell.StyleDefault.Foreground(tcell.ColorBlue)

	// Original should be unchanged
	origFg, _, _ := original[0].Decompose()
	if origFg != tcell.ColorRed {
		t.Errorf("Original style was modified: fg = %v, want %v", origFg, tcell.ColorRed)
	}
}

// TestPaletteCopyEmptySlice verifies PaletteCopy on an empty slice returns
// an empty slice.
func TestPaletteCopyEmptySlice(t *testing.T) {
	original := []tcell.Style{}
	copied := PaletteCopy(original)
	if len(copied) != 0 {
		t.Errorf("len(copied) = %d, want 0", len(copied))
	}
}
