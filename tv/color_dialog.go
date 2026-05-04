package tv

import "github.com/gdamore/tcell/v2"

// colorDialogBridge intercepts broadcasts from ColorDialog children and routes
// them to the ColorDialog. Broadcasts from children (ColorSelector, list
// viewers) flow through the Group to siblings, bypassing ColorDialog.HandleEvent.
// The bridge is inserted as a child of the Group so it receives these broadcasts
// and routes them to the ColorDialog's handler methods.
type colorDialogBridge struct {
	BaseView
	dialog *ColorDialog
}

func newColorDialogBridge(cd *ColorDialog) *colorDialogBridge {
	b := &colorDialogBridge{dialog: cd}
	b.SetBounds(NewRect(0, 0, 1, 1))
	b.SetState(SfVisible, true)
	b.SetSelf(b)
	return b
}

func (b *colorDialogBridge) HandleEvent(event *Event) {
	if event.What == EvBroadcast {
		switch event.Command {
		case CmNewColorGroup:
			if idx, ok := event.Info.(int); ok {
				b.dialog.onNewColorGroup(idx)
			}
			event.Clear()
			return
		case CmNewColorIndex:
			if idx, ok := event.Info.(int); ok {
				b.dialog.onNewColorIndex(idx)
			}
			event.Clear()
			return
		case CmColorForegroundChanged:
			if idx, ok := event.Info.(int); ok {
				b.dialog.onFgChanged(idx)
			}
			event.Clear()
			return
		case CmColorBackgroundChanged:
			if idx, ok := event.Info.(int); ok {
				b.dialog.onBgChanged(idx)
			}
			event.Clear()
			return
		}
	}
	b.BaseView.HandleEvent(event)
}

func (b *colorDialogBridge) Draw(buf *DrawBuffer) {}

// ColorDialog is a modal dialog for browsing and editing color attributes.
type ColorDialog struct {
	*Dialog
	groups     []ColorGroup
	palette    []tcell.Style
	curGroup   int
	curItem    int
	groupList  *ColorGroupListViewer
	itemList   *ColorItemListViewer
	fgSelector *ColorSelector
	bgSelector *ColorSelector
	display    *ColorDisplay
}

// NewColorDialog creates a ColorDialog with the given color groups and palette.
// The palette is deep-copied so the original is not mutated until OK is pressed.
func NewColorDialog(groups []ColorGroup, palette []tcell.Style) *ColorDialog {
	if len(groups) == 0 {
		groups = DefaultColorGroups()
	}
	if len(palette) == 0 {
		palette = make([]tcell.Style, 256)
		for i := range palette {
			palette[i] = tcell.StyleDefault
		}
	}

	cd := &ColorDialog{
		Dialog:  NewDialog(NewRect(0, 0, 64, 21), "Colors"),
		groups:  groups,
		palette: PaletteCopy(palette),
	}
	cd.SetSelf(cd)

	// Row 0: Labels
	cd.Insert(NewLabel(NewRect(1, 0, 12, 1), "~G~roup", nil))
	cd.Insert(NewLabel(NewRect(22, 0, 12, 1), "~I~tem", nil))

	// Rows 1-8: ColorGroupList + ScrollBar
	groupDS := NewColorGroupDataSource(groups)
	groupLV := NewColorGroupListViewer(NewRect(1, 1, 20, 8), groupDS)
	groupSB := NewScrollBar(NewRect(21, 1, 1, 8), Vertical)
	groupLV.SetScrollBar(groupSB)
	cd.groupList = groupLV
	cd.Insert(groupLV)
	cd.Insert(groupSB)

	// Rows 1-8: ColorItemList + ScrollBar
	itemDS := NewColorItemDataSource(groups[0].Items)
	itemLV := NewColorItemListViewer(NewRect(22, 1, 20, 8), itemDS)
	itemSB := NewScrollBar(NewRect(42, 1, 1, 8), Vertical)
	itemLV.SetScrollBar(itemSB)
	cd.itemList = itemLV
	cd.Insert(itemLV)
	cd.Insert(itemSB)

	// Row 10: Selector labels
	cd.Insert(NewLabel(NewRect(1, 10, 12, 1), "Foreground", nil))
	cd.Insert(NewLabel(NewRect(30, 10, 12, 1), "Background", nil))

	// Rows 11-14: Foreground ColorSelector (4x4 = 16)
	fgSel := NewColorSelector(NewRect(1, 11, 12, 4), 16)
	cd.fgSelector = fgSel
	cd.Insert(fgSel)

	// Rows 11-12: Background ColorSelector (4x2 = 8)
	bgSel := NewColorSelector(NewRect(30, 11, 12, 2), 8)
	cd.bgSelector = bgSel
	cd.Insert(bgSel)

	// Bridge: intercepts internal broadcasts from siblings. Must be inserted
	// before ColorDisplay so it receives CmColorForegroundChanged /
	// CmColorBackgroundChanged before ColorDisplay would clear them.
	bridge := newColorDialogBridge(cd)
	cd.Insert(bridge)

	// Rows 16-18: ColorDisplay (height 3)
	display := NewColorDisplay(NewRect(1, 16, 48, 3))
	cd.display = display
	cd.Insert(display)

	// Right-side buttons
	cd.Insert(NewButton(NewRect(50, 1, 12, 2), "~O~K", CmOK, WithDefault()))
	cd.Insert(NewButton(NewRect(50, 4, 12, 2), "~C~ancel", CmCancel))

	// Initialize: select group 0, item 0
	cd.curGroup = 0
	cd.curItem = 0
	cd.loadPaletteEntry()

	return cd
}

// Palette returns the (possibly modified) palette.
func (cd *ColorDialog) Palette() []tcell.Style {
	return cd.palette
}

// HandleEvent processes top-level events. Broadcasts from children are handled
// by colorDialogBridge, not here.
func (cd *ColorDialog) HandleEvent(event *Event) {
	cd.Dialog.HandleEvent(event)
}

func (cd *ColorDialog) paletteIndex() int {
	if cd.curGroup < 0 || cd.curGroup >= len(cd.groups) {
		return -1
	}
	items := cd.groups[cd.curGroup].Items
	if cd.curItem < 0 || cd.curItem >= len(items) {
		return -1
	}
	return items[cd.curItem].Index
}

func (cd *ColorDialog) loadPaletteEntry() {
	idx := cd.paletteIndex()
	if idx < 0 || idx >= len(cd.palette) {
		return
	}
	style := cd.palette[idx]
	fg, bg, _ := style.Decompose()
	fgIdx := colorToIndex(fg)
	bgIdx := colorToIndex(bg)
	cd.fgSelector.SetSelected(fgIdx)
	cd.bgSelector.SetSelected(bgIdx)
	cd.display.SetForeground(fgIdx)
	cd.display.SetBackground(bgIdx)
}

func (cd *ColorDialog) onNewColorGroup(idx int) {
	if idx < 0 || idx >= len(cd.groups) {
		return
	}
	// Save focus on old group
	if cd.curGroup >= 0 && cd.curGroup < len(cd.groups) {
		cd.groups[cd.curGroup].LastFocused = cd.itemList.Selected()
	}
	cd.curGroup = idx
	group := cd.groups[idx]
	itemDS := NewColorItemDataSource(group.Items)
	cd.itemList.SetDataSource(itemDS)
	cd.itemList.SetSelected(group.LastFocused)
	cd.curItem = group.LastFocused
	cd.loadPaletteEntry()
}

func (cd *ColorDialog) onNewColorIndex(idx int) {
	if idx < 0 {
		return
	}
	cd.curItem = idx
	cd.itemList.SetSelected(idx)
	cd.loadPaletteEntry()
}

func (cd *ColorDialog) onFgChanged(fgIdx int) {
	itemIdx := cd.paletteIndex()
	if itemIdx >= 0 && itemIdx < len(cd.palette) {
		_, bg, _ := cd.palette[itemIdx].Decompose()
		cd.palette[itemIdx] = tcell.StyleDefault.
			Foreground(tcell.PaletteColor(fgIdx)).
			Background(bg)
	}
	cd.display.SetForeground(fgIdx)
}

func (cd *ColorDialog) onBgChanged(bgIdx int) {
	itemIdx := cd.paletteIndex()
	if itemIdx >= 0 && itemIdx < len(cd.palette) {
		fg, _, _ := cd.palette[itemIdx].Decompose()
		cd.palette[itemIdx] = tcell.StyleDefault.
			Foreground(fg).
			Background(tcell.PaletteColor(bgIdx))
	}
	cd.display.SetBackground(bgIdx)
}

// colorToIndex converts a tcell.Color to a palette index (0-15).
func colorToIndex(c tcell.Color) int {
	for i := 0; i < 16; i++ {
		if c == tcell.PaletteColor(i) {
			return i
		}
	}
	return 0
}
