package tv

// Tests for menu.go — Task 1: Menu Data Types
// Each test verifies a specific spec requirement.
// Falsifying tests are included for every happy-path test.

import (
	"testing"
)

// ---------------------------------------------------------------------------
// MenuItem struct fields
// Spec: "MenuItem struct has fields: Label string, Command CommandCode,
//        Accel KeyBinding, Disabled bool."
// ---------------------------------------------------------------------------

func TestMenuItemHasLabelField(t *testing.T) {
	item := &MenuItem{Label: "~N~ew"}
	if item.Label != "~N~ew" {
		t.Errorf("MenuItem.Label: got %q, want %q", item.Label, "~N~ew")
	}
}

func TestMenuItemLabelFieldIsNotIgnored(t *testing.T) {
	// Falsifying: two different labels must not compare equal.
	a := &MenuItem{Label: "~O~pen"}
	b := &MenuItem{Label: "~C~lose"}
	if a.Label == b.Label {
		t.Errorf("MenuItem.Label: distinct labels must not be equal")
	}
}

func TestMenuItemHasCommandField(t *testing.T) {
	item := &MenuItem{Command: CmQuit}
	if item.Command != CmQuit {
		t.Errorf("MenuItem.Command: got %v, want %v", item.Command, CmQuit)
	}
}

func TestMenuItemCommandFieldIsNotIgnored(t *testing.T) {
	a := &MenuItem{Command: CmQuit}
	b := &MenuItem{Command: CmClose}
	if a.Command == b.Command {
		t.Errorf("MenuItem.Command: distinct commands must not be equal")
	}
}

func TestMenuItemHasAccelField(t *testing.T) {
	kb := KbCtrl('N')
	item := &MenuItem{Accel: kb}
	if item.Accel != kb {
		t.Errorf("MenuItem.Accel: got %v, want %v", item.Accel, kb)
	}
}

func TestMenuItemAccelFieldIsNotIgnored(t *testing.T) {
	a := &MenuItem{Accel: KbCtrl('N')}
	b := &MenuItem{Accel: KbCtrl('O')}
	if a.Accel == b.Accel {
		t.Errorf("MenuItem.Accel: distinct accels must not be equal")
	}
}

func TestMenuItemHasDisabledField(t *testing.T) {
	item := &MenuItem{Disabled: true}
	if !item.Disabled {
		t.Errorf("MenuItem.Disabled: got false, want true")
	}
}

func TestMenuItemDisabledDefaultsToFalse(t *testing.T) {
	// Spec says it's a bool field; zero value must be false (not disabled by default).
	item := &MenuItem{}
	if item.Disabled {
		t.Errorf("MenuItem.Disabled: zero value should be false")
	}
}

// ---------------------------------------------------------------------------
// NewMenuItem constructor
// Spec: "NewMenuItem(label string, cmd CommandCode, accel KeyBinding) *MenuItem
//        creates a MenuItem."
// ---------------------------------------------------------------------------

func TestNewMenuItemReturnsNonNilPointer(t *testing.T) {
	item := NewMenuItem("~N~ew", CmUser, KbCtrl('N'))
	if item == nil {
		t.Fatal("NewMenuItem() returned nil, want *MenuItem")
	}
}

func TestNewMenuItemSetsLabel(t *testing.T) {
	// Spec: label uses tilde notation (stored as-is, tilde chars preserved).
	item := NewMenuItem("~N~ew", CmUser, KbCtrl('N'))
	if item.Label != "~N~ew" {
		t.Errorf("NewMenuItem label: got %q, want %q", item.Label, "~N~ew")
	}
}

func TestNewMenuItemLabelStoredVerbatim(t *testing.T) {
	// Falsifying: different labels must produce different Label fields.
	a := NewMenuItem("~O~pen", CmUser, KbNone())
	b := NewMenuItem("~S~ave", CmUser, KbNone())
	if a.Label == b.Label {
		t.Errorf("NewMenuItem: different label args must produce different Label fields")
	}
}

func TestNewMenuItemSetsCommand(t *testing.T) {
	item := NewMenuItem("~Q~uit", CmQuit, KbNone())
	if item.Command != CmQuit {
		t.Errorf("NewMenuItem command: got %v, want %v", item.Command, CmQuit)
	}
}

func TestNewMenuItemCommandIsNotIgnored(t *testing.T) {
	a := NewMenuItem("~Q~uit", CmQuit, KbNone())
	b := NewMenuItem("~Q~uit", CmClose, KbNone())
	if a.Command == b.Command {
		t.Errorf("NewMenuItem: different cmd args must produce different Command fields")
	}
}

func TestNewMenuItemSetsAccel(t *testing.T) {
	kb := KbCtrl('N')
	item := NewMenuItem("~N~ew", CmUser, kb)
	if item.Accel != kb {
		t.Errorf("NewMenuItem accel: got %v, want %v", item.Accel, kb)
	}
}

func TestNewMenuItemAccelIsNotIgnored(t *testing.T) {
	a := NewMenuItem("~N~ew", CmUser, KbCtrl('N'))
	b := NewMenuItem("~N~ew", CmUser, KbCtrl('O'))
	if a.Accel == b.Accel {
		t.Errorf("NewMenuItem: different accel args must produce different Accel fields")
	}
}

// ---------------------------------------------------------------------------
// MenuSeparator type
// Spec: "MenuSeparator is a struct marker type (no fields)."
// Spec: "NewMenuSeparator() *MenuSeparator creates a separator."
// ---------------------------------------------------------------------------

func TestNewMenuSeparatorReturnsNonNilPointer(t *testing.T) {
	sep := NewMenuSeparator()
	if sep == nil {
		t.Fatal("NewMenuSeparator() returned nil, want *MenuSeparator")
	}
}

func TestMenuSeparatorIsDistinctType(t *testing.T) {
	// Spec: MenuSeparator is a struct type — verify it can be used as *MenuSeparator.
	var _ *MenuSeparator = NewMenuSeparator()
}

// ---------------------------------------------------------------------------
// SubMenu struct fields
// Spec: "SubMenu struct has fields: Label string, Items []any.
//        Items can be *MenuItem, *MenuSeparator, or *SubMenu (nested)."
// ---------------------------------------------------------------------------

func TestSubMenuHasLabelField(t *testing.T) {
	sm := &SubMenu{Label: "~F~ile"}
	if sm.Label != "~F~ile" {
		t.Errorf("SubMenu.Label: got %q, want %q", sm.Label, "~F~ile")
	}
}

func TestSubMenuHasItemsField(t *testing.T) {
	item := NewMenuItem("~N~ew", CmUser, KbCtrl('N'))
	sm := &SubMenu{Items: []any{item}}
	if len(sm.Items) != 1 {
		t.Errorf("SubMenu.Items: len %d, want 1", len(sm.Items))
	}
}

func TestSubMenuItemsCanContainMenuItems(t *testing.T) {
	item := NewMenuItem("~N~ew", CmUser, KbCtrl('N'))
	sm := &SubMenu{Items: []any{item}}
	got, ok := sm.Items[0].(*MenuItem)
	if !ok || got != item {
		t.Errorf("SubMenu.Items[0]: expected *MenuItem, got %T", sm.Items[0])
	}
}

func TestSubMenuItemsCanContainSeparators(t *testing.T) {
	sep := NewMenuSeparator()
	sm := &SubMenu{Items: []any{sep}}
	_, ok := sm.Items[0].(*MenuSeparator)
	if !ok {
		t.Errorf("SubMenu.Items[0]: expected *MenuSeparator, got %T", sm.Items[0])
	}
}

func TestSubMenuItemsCanContainNestedSubMenus(t *testing.T) {
	// Spec: Items can be *SubMenu (nested).
	inner := &SubMenu{Label: "~R~ecent"}
	outer := &SubMenu{Items: []any{inner}}
	_, ok := outer.Items[0].(*SubMenu)
	if !ok {
		t.Errorf("SubMenu.Items[0]: expected *SubMenu, got %T", outer.Items[0])
	}
}

// ---------------------------------------------------------------------------
// NewSubMenu constructor
// Spec: "NewSubMenu(label string, items ...any) *SubMenu creates a SubMenu.
//        Label uses tilde notation."
// ---------------------------------------------------------------------------

func TestNewSubMenuReturnsNonNilPointer(t *testing.T) {
	sm := NewSubMenu("~F~ile")
	if sm == nil {
		t.Fatal("NewSubMenu() returned nil, want *SubMenu")
	}
}

func TestNewSubMenuSetsLabel(t *testing.T) {
	sm := NewSubMenu("~F~ile")
	if sm.Label != "~F~ile" {
		t.Errorf("NewSubMenu label: got %q, want %q", sm.Label, "~F~ile")
	}
}

func TestNewSubMenuLabelIsNotIgnored(t *testing.T) {
	a := NewSubMenu("~F~ile")
	b := NewSubMenu("~E~dit")
	if a.Label == b.Label {
		t.Errorf("NewSubMenu: different label args must produce different Label fields")
	}
}

func TestNewSubMenuSetsItems(t *testing.T) {
	item1 := NewMenuItem("~N~ew", CmUser, KbCtrl('N'))
	item2 := NewMenuItem("~O~pen", CmUser+1, KbCtrl('O'))
	sm := NewSubMenu("~F~ile", item1, item2)
	if len(sm.Items) != 2 {
		t.Fatalf("NewSubMenu items: got len %d, want 2", len(sm.Items))
	}
}

func TestNewSubMenuItemsAreStoredInOrder(t *testing.T) {
	item1 := NewMenuItem("~N~ew", CmUser, KbCtrl('N'))
	item2 := NewMenuItem("~O~pen", CmUser+1, KbCtrl('O'))
	sm := NewSubMenu("~F~ile", item1, item2)
	if sm.Items[0] != item1 {
		t.Errorf("NewSubMenu: Items[0] = %v, want item1", sm.Items[0])
	}
	if sm.Items[1] != item2 {
		t.Errorf("NewSubMenu: Items[1] = %v, want item2", sm.Items[1])
	}
}

func TestNewSubMenuWithNoItemsHasEmptyOrNilItems(t *testing.T) {
	// Spec allows variadic with zero items.
	sm := NewSubMenu("~F~ile")
	if len(sm.Items) != 0 {
		t.Errorf("NewSubMenu with no items: got len(Items) = %d, want 0", len(sm.Items))
	}
}

func TestNewSubMenuWithMixedItemTypes(t *testing.T) {
	// Spec: Items can be *MenuItem, *MenuSeparator, or *SubMenu.
	mi := NewMenuItem("~N~ew", CmUser, KbNone())
	sep := NewMenuSeparator()
	sub := NewSubMenu("~R~ecent")
	sm := NewSubMenu("~F~ile", mi, sep, sub)
	if len(sm.Items) != 3 {
		t.Fatalf("NewSubMenu: got %d items, want 3", len(sm.Items))
	}
}

// ---------------------------------------------------------------------------
// FormatAccel
// Spec: KbCtrl('N') → "Ctrl+N"
//       KbAlt('X')  → "Alt+X"
//       KbFunc(10)  → "F10"
//       KbNone()    → "" (empty string)
//       For KbCtrl, the letter is always uppercase in the display string.
// ---------------------------------------------------------------------------

func TestFormatAccelCtrlNReturnsCtrlPlusN(t *testing.T) {
	got := FormatAccel(KbCtrl('N'))
	if got != "Ctrl+N" {
		t.Errorf("FormatAccel(KbCtrl('N')): got %q, want %q", got, "Ctrl+N")
	}
}

func TestFormatAccelCtrlLetterIsAlwaysUppercase(t *testing.T) {
	// Spec: "For KbCtrl, the letter is always uppercase in the display string."
	// KbCtrl stores the key as tcell.Key(ch-'A'+1), so passing lowercase 'n'
	// should still display as 'N'.
	got := FormatAccel(KbCtrl('n'))
	if got != "Ctrl+N" {
		t.Errorf("FormatAccel(KbCtrl('n')): got %q, want %q (letter must be uppercase)", got, "Ctrl+N")
	}
}

func TestFormatAccelCtrlLetterVariesByInput(t *testing.T) {
	// Falsifying: two different Ctrl bindings must produce different strings.
	a := FormatAccel(KbCtrl('N'))
	b := FormatAccel(KbCtrl('O'))
	if a == b {
		t.Errorf("FormatAccel: KbCtrl('N') and KbCtrl('O') must not produce the same string")
	}
}

func TestFormatAccelAltXReturnsAltPlusX(t *testing.T) {
	got := FormatAccel(KbAlt('X'))
	if got != "Alt+X" {
		t.Errorf("FormatAccel(KbAlt('X')): got %q, want %q", got, "Alt+X")
	}
}

func TestFormatAccelAltLetterVariesByInput(t *testing.T) {
	a := FormatAccel(KbAlt('X'))
	b := FormatAccel(KbAlt('Y'))
	if a == b {
		t.Errorf("FormatAccel: KbAlt('X') and KbAlt('Y') must not produce the same string")
	}
}

func TestFormatAccelFunc10ReturnsF10(t *testing.T) {
	got := FormatAccel(KbFunc(10))
	if got != "F10" {
		t.Errorf("FormatAccel(KbFunc(10)): got %q, want %q", got, "F10")
	}
}

func TestFormatAccelFuncNumberVariesByInput(t *testing.T) {
	a := FormatAccel(KbFunc(1))
	b := FormatAccel(KbFunc(10))
	if a == b {
		t.Errorf("FormatAccel: KbFunc(1) and KbFunc(10) must not produce the same string")
	}
}

func TestFormatAccelNoneReturnsEmptyString(t *testing.T) {
	got := FormatAccel(KbNone())
	if got != "" {
		t.Errorf("FormatAccel(KbNone()): got %q, want \"\"", got)
	}
}

func TestFormatAccelNoneIsDistinctFromCtrl(t *testing.T) {
	// Falsifying: KbNone must not accidentally return same as a real accel.
	ctrl := FormatAccel(KbCtrl('N'))
	none := FormatAccel(KbNone())
	if ctrl == none {
		t.Errorf("FormatAccel: KbNone() must not return same as KbCtrl('N')")
	}
}

// ---------------------------------------------------------------------------
// menuItemWidth (unexported)
// Spec: "returns runeCount(label without tildes) + 2 (gap) + runeCount(accel string).
//        If accel is KbNone, no gap or accel is counted."
// ---------------------------------------------------------------------------

func TestMenuItemWidthWithAccel(t *testing.T) {
	// Label "~N~ew" without tildes = "New" → 3 runes
	// KbCtrl('N') → "Ctrl+N" → 6 runes
	// Width = 3 + 2 + 6 = 11
	item := NewMenuItem("~N~ew", CmUser, KbCtrl('N'))
	got := menuItemWidth(item)
	if got != 11 {
		t.Errorf("menuItemWidth(\"~N~ew\", KbCtrl('N')): got %d, want 11", got)
	}
}

func TestMenuItemWidthWithNoAccel(t *testing.T) {
	// Spec: "If accel is KbNone, no gap or accel is counted."
	// Label "~N~ew" without tildes = "New" → 3 runes
	// Width = 3 (no gap, no accel)
	item := NewMenuItem("~N~ew", CmUser, KbNone())
	got := menuItemWidth(item)
	if got != 3 {
		t.Errorf("menuItemWidth(\"~N~ew\", KbNone()): got %d, want 3", got)
	}
}

func TestMenuItemWidthAccelAddsGapOf2(t *testing.T) {
	// Falsifying: width with accel must be strictly greater than label-only width.
	// The gap is exactly 2; adding any accel must add at least 2+1=3.
	withAccel := menuItemWidth(NewMenuItem("~N~ew", CmUser, KbCtrl('N')))
	withNone := menuItemWidth(NewMenuItem("~N~ew", CmUser, KbNone()))
	diff := withAccel - withNone
	// diff = 2 (gap) + len("Ctrl+N")=6 → 8
	if diff != 8 {
		t.Errorf("menuItemWidth gap+accel difference: got %d, want 8", diff)
	}
}

func TestMenuItemWidthLabelWithoutTildes(t *testing.T) {
	// Spec says runeCount(label without tildes).
	// "~S~ave As" without tildes = "Save As" → 7 runes; KbNone → width = 7.
	item := NewMenuItem("~S~ave As", CmUser, KbNone())
	got := menuItemWidth(item)
	if got != 7 {
		t.Errorf("menuItemWidth(\"~S~ave As\", KbNone()): got %d, want 7", got)
	}
}

func TestMenuItemWidthTildesNotCountedInLabel(t *testing.T) {
	// Falsifying: a label with tildes must not count tilde characters.
	// "~N~ew" → "New" (3), not "~N~ew" (6).
	item := NewMenuItem("~N~ew", CmUser, KbNone())
	got := menuItemWidth(item)
	if got == 6 {
		t.Errorf("menuItemWidth: tilde characters are being counted (got %d), tildes must be excluded", got)
	}
}

func TestMenuItemWidthWithFuncAccel(t *testing.T) {
	// "~O~pen" without tildes = "Open" → 4 runes
	// KbFunc(10) → "F10" → 3 runes
	// Width = 4 + 2 + 3 = 9
	item := NewMenuItem("~O~pen", CmUser, KbFunc(10))
	got := menuItemWidth(item)
	if got != 9 {
		t.Errorf("menuItemWidth(\"~O~pen\", KbFunc(10)): got %d, want 9", got)
	}
}

// ---------------------------------------------------------------------------
// popupWidth (unexported)
// Spec: "returns max(menuItemWidth for all MenuItems) + 4
//        (2 for borders + 2 for left/right padding)"
// ---------------------------------------------------------------------------

func TestPopupWidthAddsFourForBordersAndPadding(t *testing.T) {
	// Single item: "~N~ew" (3) + KbNone → menuItemWidth = 3
	// popupWidth = 3 + 4 = 7
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	got := popupWidth(items)
	if got != 7 {
		t.Errorf("popupWidth with one 3-rune item: got %d, want 7", got)
	}
}

func TestPopupWidthUsesMaxMenuItemWidth(t *testing.T) {
	// "~N~ew" → 3, "~O~pen" → 4
	// max = 4, popupWidth = 4 + 4 = 8
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	}
	got := popupWidth(items)
	if got != 8 {
		t.Errorf("popupWidth with items of width 3 and 4: got %d, want 8", got)
	}
}

func TestPopupWidthIgnoresSeparators(t *testing.T) {
	// Spec: max(menuItemWidth for all MenuItems) — separators are not MenuItems.
	// Only the MenuItem contributes to width.
	// "~N~ew" → 3; sep has no width contribution
	// popupWidth = 3 + 4 = 7
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuSeparator(),
	}
	got := popupWidth(items)
	if got != 7 {
		t.Errorf("popupWidth with one item + separator: got %d, want 7", got)
	}
}

func TestPopupWidthIgnoresSubMenusInMaxCalc(t *testing.T) {
	// Spec: max(menuItemWidth for all MenuItems) — only *MenuItem contributes.
	// SubMenu items are not *MenuItem; they should not affect max computation.
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()), // width 3
		NewSubMenu("~V~ery Long Sub Menu Label"), // should be ignored
	}
	got := popupWidth(items)
	if got != 7 {
		t.Errorf("popupWidth ignoring SubMenu: got %d, want 7", got)
	}
}

func TestPopupWidthFalsifyingNotJustFirstItem(t *testing.T) {
	// Falsifying: if implementation only looks at first item, this will fail.
	// First item is narrow (3), second is wide (7 = "~S~ave As" without tildes).
	// max = 7, popupWidth = 7 + 4 = 11
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),       // "New" = 3
		NewMenuItem("~S~ave As", CmUser+1, KbNone()), // "Save As" = 7
	}
	got := popupWidth(items)
	if got != 11 {
		t.Errorf("popupWidth must use max, not first item: got %d, want 11", got)
	}
}

// ---------------------------------------------------------------------------
// popupHeight (unexported)
// Spec: "returns len(items) + 2 (2 for top/bottom borders)"
// ---------------------------------------------------------------------------

func TestPopupHeightAddsTwoForBorders(t *testing.T) {
	// 1 item → height = 1 + 2 = 3
	items := []any{NewMenuItem("~N~ew", CmUser, KbNone())}
	got := popupHeight(items)
	if got != 3 {
		t.Errorf("popupHeight with 1 item: got %d, want 3", got)
	}
}

func TestPopupHeightCountsAllItemTypes(t *testing.T) {
	// Spec: "len(items) + 2" — all items count (MenuItems, separators, SubMenus).
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuSeparator(),
		NewMenuItem("~Q~uit", CmQuit, KbNone()),
	}
	got := popupHeight(items)
	if got != 5 {
		t.Errorf("popupHeight with 3 items: got %d, want 5", got)
	}
}

func TestPopupHeightWithZeroItems(t *testing.T) {
	// Spec: len(items) + 2; len([]) = 0, so height = 2.
	got := popupHeight([]any{})
	if got != 2 {
		t.Errorf("popupHeight with 0 items: got %d, want 2", got)
	}
}

func TestPopupHeightFalsifyingNotConstant(t *testing.T) {
	// Falsifying: adding an item must increase height.
	h1 := popupHeight([]any{NewMenuItem("~N~ew", CmUser, KbNone())})
	h2 := popupHeight([]any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewMenuItem("~O~pen", CmUser+1, KbNone()),
	})
	if h2 != h1+1 {
		t.Errorf("popupHeight must increase by 1 per item: h1=%d, h2=%d", h1, h2)
	}
}

func TestPopupHeightSeparatorCountsAsItem(t *testing.T) {
	// Falsifying: separator must count towards height.
	hWithout := popupHeight([]any{NewMenuItem("~N~ew", CmUser, KbNone())})
	hWith := popupHeight([]any{NewMenuItem("~N~ew", CmUser, KbNone()), NewMenuSeparator()})
	if hWith != hWithout+1 {
		t.Errorf("popupHeight: separator must count as 1 item; hWithout=%d, hWith=%d", hWithout, hWith)
	}
}

// ---------------------------------------------------------------------------
// Additional targeted tests
// ---------------------------------------------------------------------------

func TestNewMenuItemDisabledDefaultsToFalseViaConstructor(t *testing.T) {
	// Constructor must not set Disabled; it must default to false.
	item := NewMenuItem("~N~ew", CmUser, KbCtrl('N'))
	if item.Disabled {
		t.Errorf("NewMenuItem: Disabled should default to false, got true")
	}
}

func TestFormatAccelAltLowercaseInputProducesUppercase(t *testing.T) {
	// KbAlt('x') and KbAlt('X') both store Rune: 'x' (lowercased by constructor).
	// Spec says KbAlt('X') → "Alt+X", so FormatAccel must uppercase the rune.
	got := FormatAccel(KbAlt('x'))
	want := "Alt+X"
	if got != want {
		t.Errorf("FormatAccel(KbAlt('x')): got %q, want %q", got, want)
	}
}

func TestFormatAccelFunc1ReturnsF1(t *testing.T) {
	// Pin the exact formatted value for KbFunc(1).
	got := FormatAccel(KbFunc(1))
	if got != "F1" {
		t.Errorf("FormatAccel(KbFunc(1)): got %q, want %q", got, "F1")
	}
}

func TestPopupWidthWithAccelBearingItems(t *testing.T) {
	// Items with a real accelerator must have full menuItemWidth counted.
	// "~N~ew" without tildes = "New" → 3 runes
	// KbCtrl('N') → "Ctrl+N" → 6 runes
	// menuItemWidth = 3 + 2 + 6 = 11
	// popupWidth = 11 + 4 = 15
	items := []any{NewMenuItem("~N~ew", CmUser, KbCtrl('N'))}
	got := popupWidth(items)
	if got != 15 {
		t.Errorf("popupWidth with accel item (width 11): got %d, want 15", got)
	}
}

func TestPopupHeightWithSubMenuItemCounted(t *testing.T) {
	// Spec: "len(items) + 2" — *SubMenu entries count towards len(items).
	items := []any{
		NewMenuItem("~N~ew", CmUser, KbNone()),
		NewSubMenu("~R~ecent"),
	}
	got := popupHeight(items)
	if got != 4 {
		t.Errorf("popupHeight with 1 MenuItem + 1 SubMenu: got %d, want 4", got)
	}
}
