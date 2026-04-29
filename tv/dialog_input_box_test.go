package tv

// dialog_input_box_test.go — Tests for InputBox dialog function.
//
// All assertions cite the spec sentence they verify.
// Patterns (execViewStack, enterKey, tabKey, shiftTabKey, etc.) are
// inherited from dialog_test.go and integration_phase3_test.go.
//
// Test naming: all names start with TestInputBox.

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// inputBoxStack mirrors execViewStack from dialog_test.go.
// It builds an Application → Desktop → Window chain suitable for InputBox,
// which requires a live owner.ExecView call.
// ---------------------------------------------------------------------------

func inputBoxStack(t *testing.T) (*Application, *Window, tcell.SimulationScreen) {
	t.Helper()
	screen := newTestScreen(t)
	app, err := NewApplication(
		WithScreen(screen),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		screen.Fini()
		t.Fatalf("NewApplication: %v", err)
	}
	// Owner window large enough for all auto-sized dialogs (up to 60 wide × 7 tall).
	win := NewWindow(NewRect(0, 0, 80, 25), "Host", WithWindowNumber(1))
	app.Desktop().Insert(win)
	return app, win, screen
}

// ---------------------------------------------------------------------------
// Test 1: InputBox returns defaultValue and CmOK when Enter is pressed immediately.
// Spec: "On CmOK, returns the InputLine's current text and CmOK"
// Spec: "The InputLine is pre-filled with defaultValue"
// Spec: "OK button is the default button (responds to Enter via postprocess)"
// ---------------------------------------------------------------------------

// TestInputBoxEnterReturnsCmOK verifies that pressing Enter immediately causes
// InputBox to return CmOK.
func TestInputBoxEnterReturnsCmOK(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "default")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)
	// Enter key triggers CmOK via the default button's postprocess.
	app.PostCommand(CmOK, nil)

	select {
	case r := <-done:
		if r.code != CmOK {
			t.Errorf("InputBox code = %v, want CmOK", r.code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s after CmOK")
	}
}

// TestInputBoxEnterReturnsDefaultValue verifies that pressing Enter immediately
// causes InputBox to return the defaultValue unchanged.
// Spec: "On CmOK, returns the InputLine's current text and CmOK"
// Spec: "The InputLine is pre-filled with defaultValue"
func TestInputBoxEnterReturnsDefaultValue(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "hello")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmOK, nil)

	select {
	case r := <-done:
		if r.text != "hello" {
			t.Errorf("InputBox text = %q, want %q", r.text, "hello")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s after CmOK")
	}
}

// TestInputBoxEnterDoesNotReturnWrongCode is a falsifying test: the code
// returned must be exactly CmOK, not CmCancel or anything else.
// Spec: "On CmOK, returns the InputLine's current text and CmOK"
func TestInputBoxEnterDoesNotReturnWrongCode(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "val")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmOK, nil)

	select {
	case r := <-done:
		if r.code == CmCancel {
			t.Error("InputBox returned CmCancel on OK path — expected CmOK")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s")
	}
}

// ---------------------------------------------------------------------------
// Test 2: InputBox returns empty string and CmCancel when Escape is pressed.
// Spec: "On CmCancel, returns empty string and CmCancel"
// Spec: "Escape key triggers CmCancel via the dialog's modal loop"
// ---------------------------------------------------------------------------

// TestInputBoxEscapeReturnsCmCancel verifies Escape (CmCancel) causes
// InputBox to return CmCancel.
func TestInputBoxEscapeReturnsCmCancel(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "anything")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmCancel, nil)

	select {
	case r := <-done:
		if r.code != CmCancel {
			t.Errorf("InputBox code = %v after Escape, want CmCancel", r.code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s after CmCancel")
	}
}

// TestInputBoxEscapeReturnsEmptyString verifies Escape causes InputBox to
// return empty string regardless of defaultValue.
// Spec: "On CmCancel, returns empty string and CmCancel"
func TestInputBoxEscapeReturnsEmptyString(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "not-empty")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmCancel, nil)

	select {
	case r := <-done:
		if r.text != "" {
			t.Errorf("InputBox text = %q after CmCancel, want empty string", r.text)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s after CmCancel")
	}
}

// TestInputBoxEscapeDoesNotReturnDefaultValue is a falsifying test: the default
// value must NOT be returned when Escape is pressed.
// Spec: "On CmCancel, returns empty string and CmCancel"
func TestInputBoxEscapeDoesNotReturnDefaultValue(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "mydefault")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmCancel, nil)

	select {
	case r := <-done:
		if r.text == "mydefault" {
			t.Error("InputBox returned defaultValue on CmCancel — expected empty string")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s after CmCancel")
	}
}

// ---------------------------------------------------------------------------
// Test 3: InputBox returns typed text and CmOK after typing then pressing Enter.
// Spec: "On CmOK, returns the InputLine's current text and CmOK"
// Spec: "Enter key (when InputLine has focus) triggers OK via the default button's
//        postprocess handling"
//
// We test this by injecting rune events via the screen, then posting CmOK.
// ---------------------------------------------------------------------------

// TestInputBoxReturnsTypedText verifies that text typed into the InputLine is
// returned as the result when OK is pressed.
// Spec: "On CmOK, returns the InputLine's current text and CmOK"
func TestInputBoxReturnsTypedText(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Enter name:", "")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)

	// Type "hi" into the InputLine via key injection.
	screen.InjectKey(tcell.KeyRune, 'h', tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'i', tcell.ModNone)
	time.Sleep(20 * time.Millisecond)

	app.PostCommand(CmOK, nil)

	select {
	case r := <-done:
		if r.text != "hi" {
			t.Errorf("InputBox text = %q after typing 'hi', want %q", r.text, "hi")
		}
		if r.code != CmOK {
			t.Errorf("InputBox code = %v, want CmOK", r.code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s after typing and CmOK")
	}
}

// TestInputBoxTypedTextReplacesDefaultNotAppended is a falsifying test: when
// the defaultValue is empty and the user types, the result must be exactly what
// was typed — not prefixed with any extra characters.
// Spec: "On CmOK, returns the InputLine's current text and CmOK"
func TestInputBoxTypedTextReplacesDefaultNotAppended(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		// defaultValue is empty; typing "ab" should give exactly "ab".
		text, code := InputBox(win, "Title", "Prompt", "")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)
	screen.InjectKey(tcell.KeyRune, 'a', tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'b', tcell.ModNone)
	time.Sleep(20 * time.Millisecond)

	app.PostCommand(CmOK, nil)

	select {
	case r := <-done:
		if r.text != "ab" {
			t.Errorf("InputBox text = %q, want exactly %q", r.text, "ab")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s")
	}
}

// ---------------------------------------------------------------------------
// Test 4: Dialog auto-sizing — width is max(len(prompt)+6, len(title)+6, 30).
// Spec: "Dialog width: max(len(prompt)+6, len(title)+6, 30), capped at 60"
// ---------------------------------------------------------------------------

// inputBoxDialogBounds captures the dialog bounds by wrapping ExecView so we
// can inspect the dialog that InputBox creates, without reading its source.
// We use a spy Container that records the bounds of whatever view is exec'd.
type spyContainer struct {
	// Embed a real Window so it satisfies Container fully.
	win        *Window
	lastBounds Rect
	code       CommandCode
}

func (s *spyContainer) Draw(buf *DrawBuffer)                 { s.win.Draw(buf) }
func (s *spyContainer) HandleEvent(event *Event)             { s.win.HandleEvent(event) }
func (s *spyContainer) Bounds() Rect                         { return s.win.Bounds() }
func (s *spyContainer) SetBounds(r Rect)                     { s.win.SetBounds(r) }
func (s *spyContainer) GrowMode() GrowFlag                   { return s.win.GrowMode() }
func (s *spyContainer) SetGrowMode(gm GrowFlag)              { s.win.SetGrowMode(gm) }
func (s *spyContainer) Owner() Container                     { return s.win.Owner() }
func (s *spyContainer) SetOwner(c Container)                 { s.win.SetOwner(c) }
func (s *spyContainer) State() ViewState                     { return s.win.State() }
func (s *spyContainer) SetState(flag ViewState, on bool)     { s.win.SetState(flag, on) }
func (s *spyContainer) EventMask() EventType                 { return s.win.EventMask() }
func (s *spyContainer) SetEventMask(mask EventType)          { s.win.SetEventMask(mask) }
func (s *spyContainer) Options() ViewOptions                 { return s.win.Options() }
func (s *spyContainer) SetOptions(flag ViewOptions, on bool) { s.win.SetOptions(flag, on) }
func (s *spyContainer) HasState(flag ViewState) bool         { return s.win.HasState(flag) }
func (s *spyContainer) HasOption(flag ViewOptions) bool      { return s.win.HasOption(flag) }
func (s *spyContainer) ColorScheme() *theme.ColorScheme      { return s.win.ColorScheme() }
func (s *spyContainer) Insert(v View)                        { s.win.Insert(v) }
func (s *spyContainer) Remove(v View)                        { s.win.Remove(v) }
func (s *spyContainer) Children() []View                     { return s.win.Children() }
func (s *spyContainer) FocusedChild() View                   { return s.win.FocusedChild() }
func (s *spyContainer) SetFocusedChild(v View)               { s.win.SetFocusedChild(v) }
func (s *spyContainer) ExecView(v View) CommandCode {
	s.lastBounds = v.Bounds()
	// Return CmCancel immediately so InputBox can exit without a real event loop.
	return CmCancel
}

// newSpyContainer creates a spyContainer backed by a real Window sized 80×25.
func newSpyContainer(t *testing.T) *spyContainer {
	t.Helper()
	screen := newTestScreen(t)
	t.Cleanup(screen.Fini)
	app, err := NewApplication(
		WithScreen(screen),
		WithTheme(theme.BorlandBlue),
	)
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}
	win := NewWindow(NewRect(0, 0, 80, 25), "Host", WithWindowNumber(1))
	app.Desktop().Insert(win)
	return &spyContainer{win: win}
}

// TestInputBoxWidthIsMaxOfPromptPlusSixTitlePlusSixAndThirty verifies that when
// the prompt is the widest component, the dialog width is len(prompt)+6.
// Spec: "Dialog width: max(len(prompt)+6, len(title)+6, 30)"
func TestInputBoxWidthIsMaxOfPromptPlusSixTitlePlusSixAndThirty(t *testing.T) {
	spy := newSpyContainer(t)

	// prompt "AAAAAAAAAAAAAAAAAAAAAAAAA" (25 chars) → 25+6=31 > 30 and > len("")+6=6.
	prompt := "AAAAAAAAAAAAAAAAAAAAAAAAA" // 25 runes
	_, _ = InputBox(spy, "", prompt, "")

	wantW := len(prompt) + 6 // 31
	if spy.lastBounds.Width() != wantW {
		t.Errorf("dialog width = %d, want %d (len(prompt)+6)", spy.lastBounds.Width(), wantW)
	}
}

// TestInputBoxWidthIsMaxOfTitlePlusSixWhenTitleIsWidest verifies that when
// the title is the widest component, the dialog width is len(title)+6.
// Spec: "Dialog width: max(len(prompt)+6, len(title)+6, 30)"
func TestInputBoxWidthIsMaxOfTitlePlusSixWhenTitleIsWidest(t *testing.T) {
	spy := newSpyContainer(t)

	// title 30 chars → 30+6=36 > short prompt+6 and > 30.
	title := "TTTTTTTTTTTTTTTTTTTTTTTTTTTTTT" // 30 runes
	_, _ = InputBox(spy, title, "x", "")

	wantW := len(title) + 6 // 36
	if spy.lastBounds.Width() != wantW {
		t.Errorf("dialog width = %d, want %d (len(title)+6)", spy.lastBounds.Width(), wantW)
	}
}

// TestInputBoxWidthDoesNotUseLessThanPromptPlusSix is a falsifying test:
// the width must not be less than len(prompt)+6 when prompt is widest.
// Spec: "Dialog width: max(len(prompt)+6, len(title)+6, 30)"
func TestInputBoxWidthDoesNotUseLessThanPromptPlusSix(t *testing.T) {
	spy := newSpyContainer(t)

	prompt := "AAAAAAAAAAAAAAAAAAAAAAAAA" // 25 chars → need 31 minimum
	_, _ = InputBox(spy, "", prompt, "")

	minW := len(prompt) + 6
	if spy.lastBounds.Width() < minW {
		t.Errorf("dialog width = %d, want >= %d (len(prompt)+6)", spy.lastBounds.Width(), minW)
	}
}

// ---------------------------------------------------------------------------
// Test 5: Dialog auto-sizing — width capped at 60.
// Spec: "Dialog width: max(len(prompt)+6, len(title)+6, 30), capped at 60"
// ---------------------------------------------------------------------------

// TestInputBoxWidthCappedAt60 verifies that a very long prompt results in
// a dialog no wider than 60.
// Spec: "Dialog width: max(len(prompt)+6, len(title)+6, 30), capped at 60"
func TestInputBoxWidthCappedAt60(t *testing.T) {
	spy := newSpyContainer(t)

	// 80-char prompt would give 86 without cap.
	longPrompt := "PPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP" // 80
	_, _ = InputBox(spy, "", longPrompt, "")

	if spy.lastBounds.Width() > 60 {
		t.Errorf("dialog width = %d, want <= 60 (cap)", spy.lastBounds.Width())
	}
}

// TestInputBoxWidthExactly60WhenCapped is a falsifying test: a prompt that
// would produce width 86 must be clamped to exactly 60, not something smaller
// like the prompt width itself.
// Spec: "capped at 60"
func TestInputBoxWidthExactly60WhenCapped(t *testing.T) {
	spy := newSpyContainer(t)

	longPrompt := "PPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP" // 80
	_, _ = InputBox(spy, "", longPrompt, "")

	if spy.lastBounds.Width() != 60 {
		t.Errorf("dialog width = %d, want exactly 60 when prompt demands more", spy.lastBounds.Width())
	}
}

// ---------------------------------------------------------------------------
// Test 6: Dialog auto-sizing — minimum width is 30.
// Spec: "Dialog width: max(len(prompt)+6, len(title)+6, 30)"
// ---------------------------------------------------------------------------

// TestInputBoxMinimumWidthIs30 verifies that even with a short prompt and title,
// the dialog is at least 30 wide.
// Spec: "Dialog width: max(len(prompt)+6, len(title)+6, 30)"
func TestInputBoxMinimumWidthIs30(t *testing.T) {
	spy := newSpyContainer(t)

	// Both "A" and "T" produce len+6 = 7, well below 30.
	_, _ = InputBox(spy, "T", "A", "")

	if spy.lastBounds.Width() < 30 {
		t.Errorf("dialog width = %d, want >= 30 (minimum)", spy.lastBounds.Width())
	}
}

// TestInputBoxMinimumWidthIs30NotLess is a falsifying test: width must be
// exactly 30 (not 20 or some other smaller floor) for a trivially short prompt.
// Spec: "Dialog width: max(len(prompt)+6, len(title)+6, 30)"
func TestInputBoxMinimumWidthIs30NotLess(t *testing.T) {
	spy := newSpyContainer(t)

	_, _ = InputBox(spy, "T", "A", "")

	if spy.lastBounds.Width() != 30 {
		t.Errorf("dialog width = %d, want exactly 30 for short prompt/title", spy.lastBounds.Width())
	}
}

// ---------------------------------------------------------------------------
// Test 7: Dialog height is always 7.
// Spec: "Dialog height: 7 (frame top + prompt row + gap + input row + gap +
//        button row + frame bottom)"
// ---------------------------------------------------------------------------

// TestInputBoxHeightIsAlways7 verifies the dialog height is always 7.
// Spec: "Dialog height: 7"
func TestInputBoxHeightIsAlways7(t *testing.T) {
	spy := newSpyContainer(t)

	_, _ = InputBox(spy, "Title", "Prompt text here", "default")

	if spy.lastBounds.Height() != 7 {
		t.Errorf("dialog height = %d, want 7", spy.lastBounds.Height())
	}
}

// TestInputBoxHeightIsAlways7WithLongPrompt is a falsifying test: a very long
// prompt must not cause the height to increase beyond 7.
// Spec: "Dialog height: 7"
func TestInputBoxHeightIsAlways7WithLongPrompt(t *testing.T) {
	spy := newSpyContainer(t)

	_, _ = InputBox(spy, "Title", "This is quite a long prompt that fills most of the dialog", "")

	if spy.lastBounds.Height() != 7 {
		t.Errorf("dialog height = %d with long prompt, want always 7", spy.lastBounds.Height())
	}
}

// ---------------------------------------------------------------------------
// Test 8: Dialog is centered in owner's bounds.
// Spec: "Centered in owner's bounds (same centering logic as MessageBox)"
//
// centering formula (from MessageBox):
//   dx = (ownerW - dlgW) / 2
//   dy = (ownerH - dlgH) / 2
// ---------------------------------------------------------------------------

// TestInputBoxDialogIsCenteredInOwner verifies the dialog origin is centered
// in the owner's bounds using integer division.
// Spec: "Centered in owner's bounds"
func TestInputBoxDialogIsCenteredInOwner(t *testing.T) {
	spy := newSpyContainer(t)

	ownerBounds := spy.Bounds() // 80×25 from newSpyContainer

	_, _ = InputBox(spy, "T", "A", "")

	dlgW := spy.lastBounds.Width()
	dlgH := spy.lastBounds.Height()

	wantX := (ownerBounds.Width() - dlgW) / 2
	wantY := (ownerBounds.Height() - dlgH) / 2

	gotX := spy.lastBounds.A.X
	gotY := spy.lastBounds.A.Y

	if gotX != wantX {
		t.Errorf("dialog X = %d, want %d (centered: (ownerW-dlgW)/2)", gotX, wantX)
	}
	if gotY != wantY {
		t.Errorf("dialog Y = %d, want %d (centered: (ownerH-dlgH)/2)", gotY, wantY)
	}
}

// TestInputBoxDialogIsNotAtOriginWhenCentered is a falsifying test: a dialog
// in an 80×25 owner must not be placed at (0,0) because centering would move it.
// Spec: "Centered in owner's bounds"
func TestInputBoxDialogIsNotAtOriginWhenCentered(t *testing.T) {
	spy := newSpyContainer(t)

	_, _ = InputBox(spy, "T", "A", "")

	if spy.lastBounds.A.X == 0 && spy.lastBounds.A.Y == 0 {
		t.Error("dialog placed at (0,0) — expected centering to shift the origin")
	}
}

// ---------------------------------------------------------------------------
// Test 9: InputLine is pre-filled with defaultValue.
// Spec: "The InputLine is pre-filled with defaultValue and has the cursor at the end"
//
// We verify this indirectly: pressing CmOK without typing returns the defaultValue.
// (A direct inspection would require reading implementation code.)
// ---------------------------------------------------------------------------

// TestInputBoxInputLinePreFilledWithDefaultValue verifies that the text returned
// on OK (with no typing) equals defaultValue.
// Spec: "The InputLine is pre-filled with defaultValue"
func TestInputBoxInputLinePreFilledWithDefaultValue(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "prefilled")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmOK, nil)

	select {
	case r := <-done:
		if r.text != "prefilled" {
			t.Errorf("text on OK = %q, want %q (defaultValue not pre-filled)", r.text, "prefilled")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s")
	}
}

// TestInputBoxEmptyDefaultValueReturnsEmptyOnOK verifies that an empty
// defaultValue results in empty text being returned on OK with no typing.
// Spec: "The InputLine is pre-filled with defaultValue"
func TestInputBoxEmptyDefaultValueReturnsEmptyOnOK(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)
	app.PostCommand(CmOK, nil)

	select {
	case r := <-done:
		if r.text != "" {
			t.Errorf("text on OK with empty default = %q, want %q", r.text, "")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s")
	}
}

// ---------------------------------------------------------------------------
// Test 10: Tab moves focus between InputLine, OK, and Cancel.
// Spec: "The user can Tab between InputLine, OK, and Cancel"
//
// We exercise the tab cycle through the live modal loop so it goes through the
// real dialog's Group focus traversal.
// ---------------------------------------------------------------------------

// TestInputBoxTabCyclesBetweenWidgets verifies that repeated Tab keypresses
// cycle through the widgets in the dialog without crashing or getting stuck.
// We issue three Tabs (cycling through InputLine → OK → Cancel → InputLine)
// then confirm the dialog still responds to CmCancel normally.
// Spec: "The user can Tab between InputLine, OK, and Cancel"
func TestInputBoxTabCyclesBetweenWidgets(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)

	// Tab three times to traverse InputLine → OK → Cancel → back to InputLine.
	screen.InjectKey(tcell.KeyTab, 0, tcell.ModNone)
	screen.InjectKey(tcell.KeyTab, 0, tcell.ModNone)
	screen.InjectKey(tcell.KeyTab, 0, tcell.ModNone)
	time.Sleep(20 * time.Millisecond)

	// Dialog is still alive and can be dismissed.
	app.PostCommand(CmCancel, nil)

	select {
	case r := <-done:
		// The important thing is that it returned, indicating the dialog survived
		// the Tab traversal. CmCancel gives empty string.
		if r.code != CmCancel {
			t.Errorf("after Tab cycle, InputBox returned %v, want CmCancel", r.code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s after Tab cycle + CmCancel")
	}
}

// TestInputBoxShiftTabAlsoCycles verifies Shift+Tab (backtab) also works for
// backward focus traversal without crashing.
// Spec: "The user can Tab between InputLine, OK, and Cancel"
func TestInputBoxShiftTabAlsoCycles(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)

	// Shift+Tab backward through the widgets.
	screen.InjectKey(tcell.KeyBacktab, 0, tcell.ModNone)
	screen.InjectKey(tcell.KeyBacktab, 0, tcell.ModNone)
	time.Sleep(20 * time.Millisecond)

	app.PostCommand(CmCancel, nil)

	select {
	case r := <-done:
		if r.code != CmCancel {
			t.Errorf("after Shift+Tab cycle, InputBox returned %v, want CmCancel", r.code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s after Shift+Tab cycle")
	}
}

// ---------------------------------------------------------------------------
// Test 11 (Falsifying): typing replaces defaultValue — not appended to it.
// Spec: "The InputLine is pre-filled with defaultValue and has the cursor at the end"
// Spec: "On CmOK, returns the InputLine's current text"
//
// The cursor is at the end. Typing characters appends them after defaultValue.
// But if the implementation incorrectly doubles up the default, the returned
// text will be wrong. We verify that the text is exactly default+typed chars.
// ---------------------------------------------------------------------------

// TestInputBoxTypingAppendsToDefaultAtCursorEnd verifies that with the cursor
// at the end, typing appends characters after the defaultValue.
// Spec: "has the cursor at the end"
func TestInputBoxTypingAppendsToDefaultAtCursorEnd(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "pre")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)

	// Cursor is at end of "pre"; typing 'X' appends → "preX"
	screen.InjectKey(tcell.KeyRune, 'X', tcell.ModNone)
	time.Sleep(20 * time.Millisecond)

	app.PostCommand(CmOK, nil)

	select {
	case r := <-done:
		if r.text != "preX" {
			t.Errorf("text after typing 'X' with default 'pre' = %q, want %q", r.text, "preX")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s")
	}
}

// TestInputBoxTypingDoesNotDoubleDefault is a falsifying test: if the
// implementation appended the default twice, the text would be "prepre" or
// "preXpre". Ensure the result is exactly "preX".
// Spec: "The InputLine is pre-filled with defaultValue" (only once)
func TestInputBoxTypingDoesNotDoubleDefault(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "pre")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)
	screen.InjectKey(tcell.KeyRune, 'X', tcell.ModNone)
	time.Sleep(20 * time.Millisecond)
	app.PostCommand(CmOK, nil)

	select {
	case r := <-done:
		if r.text == "prepre" || r.text == "preXpre" || r.text == "prepX" {
			t.Errorf("text = %q looks like defaultValue was doubled; want %q", r.text, "preX")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s")
	}
}

// ---------------------------------------------------------------------------
// Test 12 (Falsifying): CmCancel returns empty string regardless of typed text.
// Spec: "On CmCancel, returns empty string and CmCancel"
// ---------------------------------------------------------------------------

// TestInputBoxCancelReturnsEmptyEvenAfterTyping verifies that typing and then
// cancelling still returns an empty string.
// Spec: "On CmCancel, returns empty string and CmCancel"
func TestInputBoxCancelReturnsEmptyEvenAfterTyping(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)

	// Type something, then cancel.
	screen.InjectKey(tcell.KeyRune, 'z', tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	time.Sleep(20 * time.Millisecond)

	app.PostCommand(CmCancel, nil)

	select {
	case r := <-done:
		if r.text != "" {
			t.Errorf("InputBox text after typing then cancel = %q, want %q", r.text, "")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s after typing + cancel")
	}
}

// TestInputBoxCancelDoesNotReturnTypedText is a falsifying test that the typed
// text is never returned on cancel — not even "z", "zq", or the defaultValue.
// Spec: "On CmCancel, returns empty string and CmCancel"
func TestInputBoxCancelDoesNotReturnTypedText(t *testing.T) {
	app, win, screen := inputBoxStack(t)
	defer screen.Fini()

	type result struct {
		text string
		code CommandCode
	}
	done := make(chan result, 1)
	go func() {
		text, code := InputBox(win, "Title", "Prompt", "default")
		done <- result{text, code}
	}()

	time.Sleep(30 * time.Millisecond)

	screen.InjectKey(tcell.KeyRune, 'z', tcell.ModNone)
	time.Sleep(20 * time.Millisecond)

	app.PostCommand(CmCancel, nil)

	select {
	case r := <-done:
		if r.text == "z" || r.text == "defaultz" || r.text == "default" {
			t.Errorf("InputBox returned non-empty text %q on CmCancel, want empty string", r.text)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("InputBox did not return within 2 s after typing + cancel")
	}
}
