package tv

import "github.com/gdamore/tcell/v2"

type ColorItem struct {
	Name  string
	Index int
}

type ColorGroup struct {
	Name        string
	Items       []ColorItem
	LastFocused int
}

// ColorGroupDataSource implements ListDataSource for a []ColorGroup.
type ColorGroupDataSource struct {
	groups []ColorGroup
}

func NewColorGroupDataSource(groups []ColorGroup) *ColorGroupDataSource {
	return &ColorGroupDataSource{groups: groups}
}

func (ds *ColorGroupDataSource) Count() int { return len(ds.groups) }

func (ds *ColorGroupDataSource) Item(index int) string {
	if index < 0 || index >= len(ds.groups) {
		return ""
	}
	return ds.groups[index].Name
}

// ColorItemDataSource implements ListDataSource for a []ColorItem.
type ColorItemDataSource struct {
	items []ColorItem
}

func NewColorItemDataSource(items []ColorItem) *ColorItemDataSource {
	return &ColorItemDataSource{items: items}
}

func (ds *ColorItemDataSource) Count() int { return len(ds.items) }

func (ds *ColorItemDataSource) Item(index int) string {
	if index < 0 || index >= len(ds.items) {
		return ""
	}
	return ds.items[index].Name
}

func DefaultColorGroups() []ColorGroup {
	return []ColorGroup{
		{Name: "Desktop", Items: []ColorItem{{"Background", 0}}},
		{Name: "Menus", Items: []ColorItem{
			{"Normal", 1}, {"Shortcut", 2}, {"Selected", 3}, {"Disabled", 4},
		}},
		{Name: "Dialogs", Items: []ColorItem{
			{"Frame", 5}, {"Background", 6},
		}},
		{Name: "Buttons", Items: []ColorItem{
			{"Normal", 7}, {"Default", 8}, {"Shadow", 9}, {"Shortcut", 10},
		}},
		{Name: "Input", Items: []ColorItem{
			{"Normal", 11}, {"Selection", 12},
		}},
		{Name: "Labels", Items: []ColorItem{
			{"Normal", 13}, {"Highlight", 14}, {"Shortcut", 15},
		}},
		{Name: "Lists", Items: []ColorItem{
			{"Normal", 16}, {"Selected", 17}, {"Focused", 18},
		}},
		{Name: "Status", Items: []ColorItem{
			{"Normal", 19}, {"Shortcut", 20}, {"Selected", 21},
		}},
		{Name: "Memo", Items: []ColorItem{
			{"Normal", 22}, {"Selected", 23},
		}},
	}
}

func PaletteCopy(palette []tcell.Style) []tcell.Style {
	c := make([]tcell.Style, len(palette))
	copy(c, palette)
	return c
}
