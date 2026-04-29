package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// compile-time assertion: Button must satisfy Widget.
// Spec: "Button embeds BaseView and satisfies the Widget interface: var _ Widget = (*Button)(nil)."
var _ Widget = (*Button)(nil)

// --- Construction ---

// TestNewButtonSetsSfVisible verifies NewButton sets the SfVisible state flag.
// Spec: "NewButton … Sets SfVisible."
func TestNewButtonSetsSfVisible(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "~O~K", CmOK)

	if !b.HasState(SfVisible) {
		t.Error("NewButton did not set SfVisible")
	}
}

// TestNewButtonSetsOfSelectable verifies NewButton sets the OfSelectable option.
// Spec: "NewButton … Sets … OfSelectable."
func TestNewButtonSetsOfSelectable(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "~O~K", CmOK)

	if !b.HasOption(OfSelectable) {
		t.Error("NewButton did not set OfSelectable")
	}
}

// TestNewButtonStoresBounds verifies NewButton records the given bounds.
// Spec: "NewButton(bounds Rect, title string, command CommandCode, opts ...ButtonOption) *Button"
func TestNewButtonStoresBounds(t *testing.T) {
	r := NewRect(5, 3, 14, 2)
	b := NewButton(r, "~O~K", CmOK)

	if b.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", b.Bounds(), r)
	}
}

// TestNewButtonStoresCommand verifies the button stores its CommandCode for later firing.
// Spec: "NewButton(bounds Rect, title string, command CommandCode, opts ...ButtonOption) *Button"
// Verified indirectly: pressing Space (when focused) fires event.Command == b.command.
func TestNewButtonStoresCommand(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "~O~K", CmOK)
	b.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	b.HandleEvent(ev)

	if ev.Command != CmOK {
		t.Errorf("after Space, ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
}

// TestNewButtonIsNotDefaultByDefault verifies NewButton without WithDefault does not
// act as the default button (IsDefault() == false, amDefault starts false).
// All buttons now have OfPostProcess for broadcasts and Alt+shortcuts,
// but only WithDefault() buttons start as the active default.
// Spec: "amDefault starts false for non-default button"
func TestNewButtonIsNotDefaultByDefault(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "~O~K", CmOK)

	if b.IsDefault() {
		t.Error("NewButton without WithDefault must not start as active default (IsDefault() must be false)")
	}
}

// --- ButtonOption: WithDefault ---

// TestWithDefaultSetsOfPostProcess verifies WithDefault sets OfPostProcess.
// Spec: "WithDefault() … sets OfPostProcess."
func TestWithDefaultSetsOfPostProcess(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "~O~K", CmOK, WithDefault())

	if !b.HasOption(OfPostProcess) {
		t.Error("WithDefault did not set OfPostProcess")
	}
}

// TestWithDefaultMakesButtonUseDefaultStyle verifies a button created with WithDefault
// renders with ButtonDefault background style instead of ButtonNormal.
// Spec: "Fill background with ButtonNormal style (or ButtonDefault style if isDefault)."
func TestWithDefaultMakesButtonUseDefaultStyle(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK, WithDefault())
	b.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(12, 1)
	b.Draw(buf)

	// Any cell in the button area should use ButtonDefault style, not ButtonNormal.
	cell := buf.GetCell(0, 0)
	if cell.Style == theme.BorlandBlue.ButtonNormal {
		t.Errorf("default button rendered with ButtonNormal style; expected ButtonDefault")
	}
	if cell.Style != theme.BorlandBlue.ButtonDefault {
		t.Errorf("default button cell(0,0) style = %v, want ButtonDefault %v", cell.Style, theme.BorlandBlue.ButtonDefault)
	}
}

// TestWithDefaultStyleDiffersFromNormal verifies ButtonDefault and ButtonNormal are
// distinct in BorlandBlue (falsification guard for the two style tests above).
func TestWithDefaultStyleDiffersFromNormal(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.ButtonDefault == scheme.ButtonNormal {
		t.Skip("ButtonDefault equals ButtonNormal in this scheme — style distinction test is vacuous")
	}

	normal := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)
	normal.scheme = scheme
	deflt := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK, WithDefault())
	deflt.scheme = scheme

	bufN := NewDrawBuffer(12, 1)
	bufD := NewDrawBuffer(12, 1)
	normal.Draw(bufN)
	deflt.Draw(bufD)

	cellN := bufN.GetCell(0, 0)
	cellD := bufD.GetCell(0, 0)

	if cellN.Style == cellD.Style {
		t.Errorf("normal and default buttons share background style %v; expected different styles", cellN.Style)
	}
}

// --- Draw: background fill ---

// TestButtonDrawFillsBackgroundWithButtonNormal verifies a non-default button fills
// its background with ButtonNormal style.
// Spec: "Fill background with ButtonNormal style (or ButtonDefault style if isDefault)."
func TestButtonDrawFillsBackgroundWithButtonNormal(t *testing.T) {
	b := NewButton(NewRect(0, 0, 10, 1), "OK", CmOK)
	b.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(10, 1)
	b.Draw(buf)

	// Check multiple cells across the button to confirm full fill.
	for x := 0; x < 10; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Style != theme.BorlandBlue.ButtonNormal {
			t.Errorf("non-default button cell(%d,0) style = %v, want ButtonNormal %v",
				x, cell.Style, theme.BorlandBlue.ButtonNormal)
			break
		}
	}
}

// --- Draw: bracket text ---

// TestButtonDrawRendersBrackets verifies the "[ ]" frame is drawn around the title.
// Spec: "Draw the bracket text '[ ]' around the title."
func TestButtonDrawRendersBrackets(t *testing.T) {
	// Bounds wide enough to center "[ OK ]". Title "OK" (2 chars), frame = "[ OK ]" (6 chars).
	// Use a 10-wide, 1-tall buffer. Center of "[ OK ]" in 10 is at x=2.
	b := NewButton(NewRect(0, 0, 10, 1), "OK", CmOK)
	b.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(10, 1)
	b.Draw(buf)

	// Find the '[' bracket somewhere in the row.
	found := false
	for x := 0; x < 10; x++ {
		if buf.GetCell(x, 0).Rune == '[' {
			found = true
			break
		}
	}
	if !found {
		t.Error("Draw did not render '[' bracket")
	}
}

// TestButtonDrawRendersClosingBracket verifies ']' appears in the output.
// Spec: "Draw the bracket text '[ ]' around the title."
func TestButtonDrawRendersClosingBracket(t *testing.T) {
	b := NewButton(NewRect(0, 0, 10, 1), "OK", CmOK)
	b.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(10, 1)
	b.Draw(buf)

	found := false
	for x := 0; x < 10; x++ {
		if buf.GetCell(x, 0).Rune == ']' {
			found = true
			break
		}
	}
	if !found {
		t.Error("Draw did not render ']' bracket")
	}
}

// TestButtonDrawBracketsEncloseTitleWithSpaces verifies the layout is "[ Title ]"
// (space-bracket-space pattern): '[' then space then title then space then ']'.
// Spec: "renders as '[ Title ]' (space-bracket-space pattern)."
func TestButtonDrawBracketsEncloseTitleWithSpaces(t *testing.T) {
	// Title "AB" (2 chars). Full bracketed text is "[ AB ]" (6 chars).
	// Use an exact-fit 6-wide buffer so the bracket string starts at column 0.
	b := NewButton(NewRect(0, 0, 6, 1), "AB", CmOK)
	b.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(6, 1)
	b.Draw(buf)

	// Expected layout: '[', ' ', 'A', 'B', ' ', ']'
	expected := []rune{'[', ' ', 'A', 'B', ' ', ']'}
	for i, want := range expected {
		got := buf.GetCell(i, 0).Rune
		if got != want {
			t.Errorf("cell(%d,0) = %q, want %q in layout \"[ AB ]\"", i, got, want)
		}
	}
}

// --- Draw: tilde shortcut rendering ---

// TestButtonDrawNormalTitleSegmentUsesButtonNormalStyle verifies non-shortcut title
// characters are drawn with ButtonNormal style (or ButtonDefault for default buttons).
// Spec: "Title segments use tilde parsing — shortcut letters render in ButtonShortcut style."
func TestButtonDrawNormalTitleSegmentUsesButtonNormalStyle(t *testing.T) {
	// "OK" — no tilde, entire title is normal text.
	b := NewButton(NewRect(0, 0, 6, 1), "OK", CmOK)
	b.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(6, 1)
	b.Draw(buf)

	// 'O' is at column 2 in "[ OK ]".
	cell := buf.GetCell(2, 0)
	if cell.Style == theme.BorlandBlue.ButtonShortcut {
		t.Errorf("non-shortcut title char uses ButtonShortcut style; expected ButtonNormal")
	}
}

// TestButtonDrawShortcutLetterUsesButtonShortcutStyle verifies the tilde-enclosed
// letter is drawn with ButtonShortcut style.
// Spec: "shortcut letters render in ButtonShortcut style."
func TestButtonDrawShortcutLetterUsesButtonShortcutStyle(t *testing.T) {
	// "~O~K": 'O' is the shortcut letter. Layout: "[ OK ]", shortcut 'O' at col 2.
	b := NewButton(NewRect(0, 0, 6, 1), "~O~K", CmOK)
	b.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(6, 1)
	b.Draw(buf)

	// 'O' should be at column 2 ("[ ~O~K ]" → "[ OK ]").
	cell := buf.GetCell(2, 0)
	if cell.Style != theme.BorlandBlue.ButtonShortcut {
		t.Errorf("shortcut letter 'O' cell(2,0) style = %v, want ButtonShortcut %v",
			cell.Style, theme.BorlandBlue.ButtonShortcut)
	}
}

// TestButtonDrawShortcutStyleDiffersFromNormal verifies ButtonShortcut and
// ButtonNormal are distinct in BorlandBlue (falsification guard).
func TestButtonDrawShortcutStyleDiffersFromNormal(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.ButtonShortcut == scheme.ButtonNormal {
		t.Skip("ButtonShortcut equals ButtonNormal in this scheme — style distinction test is vacuous")
	}

	b := NewButton(NewRect(0, 0, 6, 1), "~O~K", CmOK)
	b.scheme = scheme

	buf := NewDrawBuffer(6, 1)
	b.Draw(buf)

	shortcutCell := buf.GetCell(2, 0) // 'O' — shortcut
	normalCell := buf.GetCell(3, 0)   // 'K' — normal

	if shortcutCell.Style == normalCell.Style {
		t.Errorf("shortcut and normal title chars have same style %v; expected different styles", shortcutCell.Style)
	}
}

// TestButtonDrawNormalCharAfterShortcutUsesNormalStyle verifies the non-shortcut
// characters following a tilde-marked letter still use ButtonNormal style.
// Spec: "Title segments use tilde parsing — shortcut letters render in ButtonShortcut style."
func TestButtonDrawNormalCharAfterShortcutUsesNormalStyle(t *testing.T) {
	// "~O~K": 'K' at column 3 is a normal segment.
	b := NewButton(NewRect(0, 0, 6, 1), "~O~K", CmOK)
	b.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(6, 1)
	b.Draw(buf)

	cell := buf.GetCell(3, 0) // 'K'
	if cell.Style == theme.BorlandBlue.ButtonShortcut {
		t.Errorf("non-shortcut char 'K' at cell(3,0) uses ButtonShortcut style; expected ButtonNormal")
	}
	if cell.Style != theme.BorlandBlue.ButtonNormal {
		t.Errorf("non-shortcut char 'K' at cell(3,0) style = %v, want ButtonNormal %v",
			cell.Style, theme.BorlandBlue.ButtonNormal)
	}
}

// --- Draw: focus cursor ---

// TestButtonDrawFocusCursorWhenSelected verifies a '►' cursor is drawn at (0,0) when
// the button has focus (SfSelected state).
// Spec: "If the button has focus (SfSelected state), draw a '►' cursor at position (0, 0)."
func TestButtonDrawFocusCursorWhenSelected(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)
	b.scheme = theme.BorlandBlue
	b.SetState(SfSelected, true)

	buf := NewDrawBuffer(12, 1)
	b.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != '►' {
		t.Errorf("focused button cell(0,0) = %q, want '►' cursor", cell.Rune)
	}
}

// TestButtonDrawNoCursorWhenNotSelected verifies no '►' cursor is drawn at (0,0) when
// the button does not have focus.
// Spec: "If the button has focus (SfSelected state), draw a '►' cursor at position (0, 0)."
func TestButtonDrawNoCursorWhenNotSelected(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)
	b.scheme = theme.BorlandBlue
	b.SetState(SfSelected, false)

	buf := NewDrawBuffer(12, 1)
	b.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == '►' {
		t.Errorf("unfocused button cell(0,0) = '►'; cursor must only appear when button has focus")
	}
}

// --- Draw: shadow ---

// TestButtonDrawShadowRightColumnWhenHeightAtLeastTwo verifies that when bounds
// height >= 2, a ButtonShadow-styled cell appears at the right column, row 1.
// Spec: "Draw a 1-cell shadow … right column at (width-1, 1..height) … only if bounds height >= 2."
func TestButtonDrawShadowRightColumnWhenHeightAtLeastTwo(t *testing.T) {
	// 10 wide, 2 tall — height >= 2.
	b := NewButton(NewRect(0, 0, 10, 2), "OK", CmOK)
	b.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(10, 2)
	b.Draw(buf)

	// Right column is x=9 (width-1=9). Shadow row is y=1.
	cell := buf.GetCell(9, 1)
	if cell.Style != theme.BorlandBlue.ButtonShadow {
		t.Errorf("shadow at cell(9,1) style = %v, want ButtonShadow %v",
			cell.Style, theme.BorlandBlue.ButtonShadow)
	}
}

// TestButtonDrawShadowBottomRowWhenHeightAtLeastTwo verifies that when bounds
// height >= 2, a ButtonShadow-styled cell appears at the bottom row, column 1.
// Spec: "bottom row at (1, height-1..height) … only if bounds height >= 2."
func TestButtonDrawShadowBottomRowWhenHeightAtLeastTwo(t *testing.T) {
	b := NewButton(NewRect(0, 0, 10, 2), "OK", CmOK)
	b.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(10, 2)
	b.Draw(buf)

	// Bottom row is y=1 (height-1=1). Shadow column starts at x=1.
	cell := buf.GetCell(1, 1)
	if cell.Style != theme.BorlandBlue.ButtonShadow {
		t.Errorf("shadow at cell(1,1) style = %v, want ButtonShadow %v",
			cell.Style, theme.BorlandBlue.ButtonShadow)
	}
}

// TestButtonDrawNoShadowWhenHeightOne verifies that when bounds height == 1,
// no shadow is drawn (there is no second row to draw into).
// Spec: "only if bounds height >= 2."
func TestButtonDrawNoShadowWhenHeightOne(t *testing.T) {
	// Use a 2-row buffer to give room, but button bounds are height=1.
	b := NewButton(NewRect(0, 0, 10, 1), "OK", CmOK)
	b.scheme = theme.BorlandBlue

	buf := NewDrawBuffer(10, 2)
	b.Draw(buf)

	// Row 1 is outside the button's 1-row bounds; it must not be written by the button.
	// All cells in row 1 should still have default style (unmodified by Draw).
	for x := 0; x < 10; x++ {
		cell := buf.GetCell(x, 1)
		if cell.Style == theme.BorlandBlue.ButtonShadow {
			t.Errorf("shadow drawn at cell(%d,1) even though bounds height=1; must not draw shadow", x)
		}
	}
}

// TestButtonDrawShadowStyleDiffersFromNormal verifies ButtonShadow and ButtonNormal
// are distinct in BorlandBlue (falsification guard for shadow tests).
func TestButtonDrawShadowStyleDiffersFromNormal(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.ButtonShadow == scheme.ButtonNormal {
		t.Skip("ButtonShadow equals ButtonNormal in this scheme — shadow style distinction test is vacuous")
	}
}

// --- HandleEvent: Enter key (removed in phase9) ---

// TestButtonHandleEventEnterDoesNotFireCommand verifies Enter key does NOT fire the
// button. Enter handling has been removed; Dialog converts Enter to a CmDefault
// broadcast instead, which reaches the active default button via EvBroadcast.
// Spec: "Enter handler completely removed."
func TestButtonHandleEventEnterDoesNotFireCommand(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)
	b.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnter}}
	b.HandleEvent(ev)

	if ev.What == EvCommand {
		t.Errorf("Enter should NOT fire button (handler removed); ev.What = %v", ev.What)
	}
}

// TestButtonHandleEventSpaceFiresCommandForCancel verifies the fired Space command
// matches the command passed to NewButton, tested with a non-standard command code.
// Spec: "event.Command = b.command"
func TestButtonHandleEventSpaceFiresCommandForCancel(t *testing.T) {
	b := NewButton(NewRect(0, 0, 14, 1), "~C~ancel", CmCancel)
	b.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	b.HandleEvent(ev)

	if ev.Command != CmCancel {
		t.Errorf("after Space on Cancel button, ev.Command = %v, want CmCancel (%v)", ev.Command, CmCancel)
	}
}

// --- HandleEvent: Space key ---

// TestButtonHandleEventSpaceFiresCommand verifies Space key fires the command,
// same transformation as Enter.
// Spec: "Space key (tcell.KeyRune, rune ' '): same as Enter — fires the command."
func TestButtonHandleEventSpaceFiresCommand(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)
	b.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	b.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("after Space, ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("after Space, ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
	if ev.Key != nil {
		t.Errorf("after Space, ev.Key = %v, want nil", ev.Key)
	}
}

// TestButtonHandleEventSpaceNilsKey verifies Space sets ev.Key to nil.
// Spec: "event.Key = nil."
func TestButtonHandleEventSpaceNilsKey(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)
	b.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: ' '}}
	b.HandleEvent(ev)

	if ev.Key != nil {
		t.Errorf("after Space, ev.Key = %v, want nil", ev.Key)
	}
}

// TestButtonHandleEventOtherKeyDoesNotFireCommand verifies that a non-Enter/Space key
// does not fire a command.
// Spec: only Enter and Space fire the command.
func TestButtonHandleEventOtherKeyDoesNotFireCommand(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)
	b.SetState(SfSelected, true)

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'x'}}
	b.HandleEvent(ev)

	if ev.What == EvCommand {
		t.Errorf("pressing 'x' fired a command; only Enter and Space should fire the command")
	}
}

// --- HandleEvent: Mouse click ---

// TestButtonHandleEventMouseButton1FiresCommand verifies a left-click (Button1)
// fires the command.
// Spec: "Mouse click (EvMouse, Button1): fires the command (same transformation as Enter)."
func TestButtonHandleEventMouseButton1FiresCommand(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button1},
	}
	b.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("after Button1 click, ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("after Button1 click, ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
	if ev.Key != nil {
		t.Errorf("after Button1 click, ev.Key = %v, want nil", ev.Key)
	}
}

// TestButtonHandleEventMouseOtherButtonDoesNotFireCommand verifies that mouse buttons
// other than Button1 do not fire the command.
// Spec: "Mouse click (EvMouse, Button1)" — only Button1 fires.
func TestButtonHandleEventMouseOtherButtonDoesNotFireCommand(t *testing.T) {
	b := NewButton(NewRect(0, 0, 12, 1), "OK", CmOK)

	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 5, Y: 0, Button: tcell.Button2},
	}
	b.HandleEvent(ev)

	if ev.What == EvCommand {
		t.Errorf("Button2 click fired a command; only Button1 should fire the command")
	}
}

// --- HandleEvent: postprocess default button ---

// TestDefaultButtonFiresViaCmDefaultBroadcast verifies that a default button
// (WithDefault) fires its command when it receives a CmDefault broadcast via the
// postprocess phase.
// Spec: "As a default button with OfPostProcess: responds to CmDefault broadcast
// (the focused child gets first crack via normal event dispatch; the dialog converts
// Enter to a CmDefault broadcast which reaches the default button via postprocess)."
func TestDefaultButtonFiresViaCmDefaultBroadcast(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// A non-consuming focused child (plain BaseView doesn't consume any event).
	nonConsumer := &BaseView{}
	nonConsumer.SetBounds(NewRect(0, 0, 20, 1))
	nonConsumer.SetState(SfVisible, true)
	nonConsumer.SetOptions(OfSelectable, true)

	okBtn := NewButton(NewRect(22, 0, 12, 1), "~O~K", CmOK, WithDefault())

	g.Insert(okBtn)
	g.Insert(nonConsumer) // inserted last → steals focus from okBtn

	// Confirm nonConsumer is focused, not okBtn.
	if g.FocusedChild() != nonConsumer {
		t.Fatalf("precondition: FocusedChild() = %v, want nonConsumer", g.FocusedChild())
	}

	// Dialog sends CmDefault broadcast when Enter is pressed; replicate that here.
	ev := &Event{What: EvBroadcast, Command: CmDefault}
	g.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("default button did not fire via CmDefault broadcast; ev.What = %v, want EvCommand (%v)", ev.What, EvCommand)
	}
	if ev.Command != CmOK {
		t.Errorf("default button CmDefault broadcast fired wrong command; ev.Command = %v, want CmOK (%v)", ev.Command, CmOK)
	}
}

// TestDefaultButtonDoesNotFireWhenAmDefaultFalse verifies that a WithDefault button
// with amDefault=false (because a non-default sibling has focus) does NOT fire on
// CmDefault broadcast.
// Spec: "When button receives CmDefault broadcast: if amDefault, call press(event)"
func TestDefaultButtonDoesNotFireWhenAmDefaultFalse(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// A selectable child that will steal the default status.
	sibling := NewButton(NewRect(0, 0, 12, 1), "Other", CmCancel)
	okBtn := NewButton(NewRect(14, 0, 12, 1), "~O~K", CmOK, WithDefault())

	g.Insert(okBtn)
	g.Insert(sibling) // steals focus → sibling.amDefault=true, okBtn.amDefault=false

	// okBtn.amDefault is now false; CmDefault should fire sibling (Cancel).
	ev := &Event{What: EvBroadcast, Command: CmDefault}
	g.HandleEvent(ev)

	if ev.What == EvCommand && ev.Command == CmOK {
		t.Errorf("default button (amDefault=false) fired on CmDefault; should not fire when not active default")
	}
}

// TestNonDefaultButtonWithFocusFiresViaCmDefault verifies that a non-default button
// that has gained focus (amDefault=true) fires when it receives CmDefault broadcast,
// even though it was not created with WithDefault.
// Spec: "Non-default focused button responds to CmDefault"
func TestNonDefaultButtonWithFocusFiresViaCmDefault(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	defBtn := NewButton(NewRect(0, 0, 12, 1), "~O~K", CmOK, WithDefault())
	plainBtn := NewButton(NewRect(14, 0, 14, 1), "~C~ancel", CmCancel)

	g.Insert(defBtn)
	g.Insert(plainBtn) // plainBtn focused → plainBtn.amDefault = true, defBtn.amDefault = false

	// plainBtn is the active default; CmDefault should fire CmCancel.
	ev := &Event{What: EvBroadcast, Command: CmDefault}
	g.HandleEvent(ev)

	if ev.What != EvCommand {
		t.Errorf("focused non-default button should fire on CmDefault; ev.What = %v", ev.What)
	}
	if ev.Command != CmCancel {
		t.Errorf("focused non-default button CmDefault fired %v, want CmCancel (%v)", ev.Command, CmCancel)
	}
}

// TestDefaultButtonFiresCommandForCustomCode verifies the CmDefault broadcast mechanism
// works for command codes other than CmOK.
// Spec: "event.Command = b.command"
func TestDefaultButtonFiresCommandForCustomCode(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	nonConsumer := &BaseView{}
	nonConsumer.SetBounds(NewRect(0, 0, 20, 1))
	nonConsumer.SetState(SfVisible, true)
	nonConsumer.SetOptions(OfSelectable, true)

	customCmd := CmUser + 42
	btn := NewButton(NewRect(22, 0, 14, 1), "~A~pply", customCmd, WithDefault())

	g.Insert(btn)
	g.Insert(nonConsumer)

	// btn loses amDefault because nonConsumer (non-button) stole focus but didn't
	// broadcast CmGrabDefault (BaseView doesn't), so btn still has amDefault=true.
	// Wait — nonConsumer is BaseView, not Button. Only buttons broadcast.
	// So btn.amDefault stays true. Confirm and send CmDefault.
	ev := &Event{What: EvBroadcast, Command: CmDefault}
	g.HandleEvent(ev)

	if ev.Command != customCmd {
		t.Errorf("default button CmDefault broadcast fired command %v, want %v", ev.Command, customCmd)
	}
}
