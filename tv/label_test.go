package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// compile-time assertion: Label must satisfy Widget.
// Spec: "Label embeds BaseView and satisfies the Widget interface: var _ Widget = (*Label)(nil)."
var _ Widget = (*Label)(nil)

// labelLinkedView is a stub view used as the focus target in Label tests.
type labelLinkedView struct {
	BaseView
}

func newLabelLinkedView() *labelLinkedView {
	v := &labelLinkedView{}
	v.SetBounds(NewRect(0, 0, 10, 1))
	v.SetState(SfVisible, true)
	v.SetOptions(OfSelectable, true)
	return v
}

// --- Construction ---

// TestNewLabelSetsSfVisible verifies NewLabel sets the SfVisible state flag.
// Spec: "NewLabel(bounds, label, link) … Sets SfVisible."
func TestNewLabelSetsSfVisible(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)

	if !label.HasState(SfVisible) {
		t.Error("NewLabel did not set SfVisible")
	}
}

// TestNewLabelIsNotSelectable verifies NewLabel does NOT set OfSelectable.
// Spec: "NOT selectable."
func TestNewLabelIsNotSelectable(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)

	if label.HasOption(OfSelectable) {
		t.Error("NewLabel must not set OfSelectable")
	}
}

// TestNewLabelStoresBounds verifies NewLabel records the given bounds.
// Spec: "NewLabel(bounds Rect, label string, link View) *Label"
func TestNewLabelStoresBounds(t *testing.T) {
	r := NewRect(5, 10, 30, 1)
	label := NewLabel(r, "~N~ame", nil)

	if label.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", label.Bounds(), r)
	}
}

// TestNewLabelSetsOfPostProcess verifies NewLabel sets the OfPostProcess option
// so the focused child has priority before the label handles Alt+letter.
// Spec: "Label sets OfPostProcess so focused view gets priority for Alt+shortcut."
func TestNewLabelSetsOfPostProcess(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)

	if !label.HasOption(OfPostProcess) {
		t.Error("NewLabel did not set OfPostProcess")
	}
}

// TestNewLabelWithNilLink verifies NewLabel accepts a nil link without panicking.
// Spec: "link is the view to focus when the shortcut is activated."
func TestNewLabelWithNilLink(t *testing.T) {
	// Must not panic.
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	if label == nil {
		t.Error("NewLabel returned nil")
	}
}

// --- Draw: normal segments use LabelNormal style ---

// TestLabelDrawNormalSegmentUsesLabelNormalStyle verifies non-shortcut segments
// are drawn with LabelNormal style.
// Spec: "renders the label text with LabelNormal style for normal segments."
func TestLabelDrawNormalSegmentUsesLabelNormalStyle(t *testing.T) {
	// "Name" has no tilde — entire text is a normal segment.
	label := NewLabel(NewRect(0, 0, 20, 1), "Name", nil)
	scheme := theme.BorlandBlue
	label.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.LabelNormal {
		t.Errorf("normal segment cell(0,0) style = %v, want LabelNormal %v", cell.Style, scheme.LabelNormal)
	}
}

// TestLabelDrawShortcutSegmentUsesLabelShortcutStyle verifies shortcut segments
// (tilde-enclosed text) are drawn with LabelShortcut style.
// Spec: "LabelShortcut style for tilde-enclosed segments."
func TestLabelDrawShortcutSegmentUsesLabelShortcutStyle(t *testing.T) {
	// "~N~ame": segment 0 is "N" (shortcut), segment 1 is "ame" (normal).
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	scheme := theme.BorlandBlue
	label.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.LabelShortcut {
		t.Errorf("shortcut segment cell(0,0) style = %v, want LabelShortcut %v", cell.Style, scheme.LabelShortcut)
	}
}

// TestLabelDrawNormalSegmentAfterShortcut verifies normal text following a
// shortcut segment still uses LabelNormal style.
// Spec: "LabelNormal style for normal segments."
func TestLabelDrawNormalSegmentAfterShortcut(t *testing.T) {
	// "~N~ame": "N" at col 0 is shortcut; "ame" at cols 1-3 is normal.
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	scheme := theme.BorlandBlue
	label.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	// Column 1 is the start of "ame" — a normal segment.
	cell := buf.GetCell(1, 0)
	if cell.Style != scheme.LabelNormal {
		t.Errorf("normal segment after shortcut cell(1,0) style = %v, want LabelNormal %v", cell.Style, scheme.LabelNormal)
	}
}

// TestLabelDrawRendersCorrectRunes verifies the runes appear at the correct positions.
// Spec: "renders the label text … Uses ParseTildeLabel."
func TestLabelDrawRendersCorrectRunes(t *testing.T) {
	// "~N~ame" → 'N' at col 0, 'a' at col 1, 'm' at col 2, 'e' at col 3.
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	label.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	expected := []rune{'N', 'a', 'm', 'e'}
	for i, want := range expected {
		got := buf.GetCell(i, 0).Rune
		if got != want {
			t.Errorf("cell(%d,0) rune = %q, want %q", i, got, want)
		}
	}
}

// TestLabelDrawNoTildeRendersEntireTextAsNormal verifies a label with no tilde
// renders everything with LabelNormal.
// Spec: "LabelNormal style for normal segments."
func TestLabelDrawNoTildeRendersEntireTextAsNormal(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "Open", nil)
	scheme := theme.BorlandBlue
	label.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	// All four characters must use LabelNormal style.
	for i := range "Open" {
		cell := buf.GetCell(i, 0)
		if cell.Style != scheme.LabelNormal {
			t.Errorf("no-tilde label cell(%d,0) style = %v, want LabelNormal", i, cell.Style)
		}
	}
}

// TestLabelDrawShortcutStyleDiffersFromNormal verifies LabelShortcut and
// LabelNormal are actually distinct in BorlandBlue (falsification guard).
func TestLabelDrawShortcutStyleDiffersFromNormal(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.LabelShortcut == scheme.LabelNormal {
		t.Skip("LabelShortcut equals LabelNormal in this scheme — style distinction test is vacuous")
	}

	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	label.scheme = scheme

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	shortcutCell := buf.GetCell(0, 0) // 'N' — shortcut segment
	normalCell := buf.GetCell(1, 0)   // 'a' — normal segment

	if shortcutCell.Style == normalCell.Style {
		t.Errorf("shortcut and normal segments have the same style %v; expected different styles", shortcutCell.Style)
	}
}

// --- HandleEvent: shortcut activation ---

// TestLabelHandleEventAltShortcutSetsFocusedChild verifies that Alt+<shortcut>
// causes the Label to call owner.SetFocusedChild(link).
// Spec: "if the event is Alt+<shortcut letter> … Label sets its owner's focused child
// to link by calling owner.SetFocusedChild(link)."
func TestLabelHandleEventAltShortcutSetsFocusedChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other) // steals focus from linked

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() != linked {
		t.Errorf("FocusedChild() = %v, want linked view after Alt+N", g.FocusedChild())
	}
}

// TestLabelHandleEventAltShortcutClearsEvent verifies the event is cleared after
// the Label handles its shortcut.
// Spec: "Clears the event."
func TestLabelHandleEventAltShortcutClearsEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("event.What = %v after Alt+N shortcut, want EvNothing (cleared)", ev.What)
	}
}

// TestLabelHandleEventShortcutCaseInsensitiveUppercase verifies the shortcut
// matches when the event rune is uppercase.
// Spec: "The shortcut letter matching is case-insensitive."
func TestLabelHandleEventShortcutCaseInsensitiveUppercase(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'N', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() != linked {
		t.Errorf("FocusedChild() = %v, want linked after Alt+N (uppercase), shortcut is case-insensitive", g.FocusedChild())
	}
}

// TestLabelHandleEventShortcutCaseInsensitiveLowercase verifies a lowercase
// tilde-shortcut is triggered by Alt+uppercase.
// Spec: "The shortcut letter matching is case-insensitive."
func TestLabelHandleEventShortcutCaseInsensitiveLowercase(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~s~ave", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'S', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() != linked {
		t.Errorf("FocusedChild() = %v, want linked after Alt+S (uppercase), shortcut is case-insensitive", g.FocusedChild())
	}
}

// TestLabelHandleEventWrongLetterDoesNotActivate verifies a different Alt+letter
// does not trigger the shortcut.
// Spec: "if the event is Alt+<shortcut letter>"
func TestLabelHandleEventWrongLetterDoesNotActivate(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other) // steals focus from linked

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'x', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() == linked {
		t.Errorf("FocusedChild() = linked after Alt+X (wrong letter), must not activate")
	}
}

// TestLabelHandleEventNoAltModifierDoesNotActivate verifies a plain letter (no
// Alt modifier) does not trigger the shortcut.
// Spec: "if the event is Alt+<shortcut letter>"
func TestLabelHandleEventNoAltModifierDoesNotActivate(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: 0},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() == linked {
		t.Errorf("FocusedChild() = linked after plain 'n' (no Alt), must not activate")
	}
}

// TestLabelHandleEventNilLinkDoesNotPanic verifies that when link is nil,
// an Alt+shortcut event does not panic.
// Spec: "If link is nil, the Label renders normally but does not respond to keyboard events."
func TestLabelHandleEventNilLinkDoesNotPanic(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	g.Insert(label)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	// Must not panic.
	g.HandleEvent(ev)
}

// TestLabelHandleEventNilLinkDoesNotClearEvent verifies that when link is nil,
// an Alt+shortcut event is NOT consumed (event is not cleared).
// Spec: "If link is nil, the Label … does not respond to keyboard events."
func TestLabelHandleEventNilLinkDoesNotClearEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)
	g.Insert(label)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	// The event must not have been cleared (link is nil, label does not handle it).
	if ev.IsCleared() {
		t.Errorf("event was cleared by Label with nil link; it should not respond to keyboard events")
	}
}

// TestLabelPostProcessActivatesWhenFocusedChildDoesNotConsume verifies that
// because Label has OfPostProcess, the shortcut fires when another child has
// focus but does not consume the Alt+letter event.
// Spec: "Label sets OfPostProcess so focused view gets priority for Alt+shortcut;
// if the focused child ignores the event, the label activates."
func TestLabelPostProcessActivatesWhenFocusedChildDoesNotConsume(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)

	// A second selectable view that gets focus first.
	other := &labelLinkedView{}
	other.SetBounds(NewRect(20, 0, 10, 1))
	other.SetState(SfVisible, true)
	other.SetOptions(OfSelectable, true)

	g.Insert(label)
	g.Insert(linked)
	g.Insert(other) // other gets focus because it's inserted last and is selectable

	// Confirm other is focused.
	if g.FocusedChild() != other {
		t.Fatalf("precondition: FocusedChild() = %v, want other", g.FocusedChild())
	}

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() != linked {
		t.Errorf("FocusedChild() = %v, want linked; label post-process must activate when focused child ignores Alt+n", g.FocusedChild())
	}
}

// TestLabelHandleEventUsesFirstRuneOfShortcutSegment verifies that only the first
// rune of a multi-character tilde segment is used as the shortcut letter.
// Spec: "shortcut letter is the first rune of the tilde-enclosed segment."
func TestLabelHandleEventUsesFirstRuneOfShortcutSegment(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~Na~me", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	evN := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(evN)

	if g.FocusedChild() != linked {
		t.Errorf("FocusedChild() = %v, want linked after Alt+N (first rune of shortcut segment)", g.FocusedChild())
	}
}

// TestLabelHandleEventSecondRuneOfShortcutSegmentDoesNotActivate verifies that
// the second (and beyond) rune of a multi-rune tilde segment does NOT act as a shortcut.
// Spec: "shortcut letter is the first rune of the tilde-enclosed segment."
func TestLabelHandleEventSecondRuneOfShortcutSegmentDoesNotActivate(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~Na~me", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	evA := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'a', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(evA)

	if g.FocusedChild() == linked {
		t.Errorf("FocusedChild() = linked after Alt+A (second rune of shortcut segment); only first rune is the shortcut")
	}
}

// TestLabelNoTildeHasNoShortcut verifies that a label with no tilde notation
// does not activate on any Alt+letter event.
// Spec: shortcut letter is the first rune of the tilde-enclosed segment — if
// there is no tilde-enclosed segment, there is no shortcut.
func TestLabelNoTildeHasNoShortcut(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "Name", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() == linked {
		t.Errorf("FocusedChild() = linked after Alt+N on label with no tilde; must not activate")
	}
}
