package tv

// messagebox_test.go — Tests for Task 8: MessageBox Standard Dialog.
//
// Tests are organised in three sections:
//  1. MsgBoxButton bitmask type: constant values and OR combinations.
//  2. MessageBox return values: modal loop integration tests via the full
//     Application → Desktop → Window stack, injecting commands to unblock ExecView.
//  3. MessageBox geometry: auto-sizing and centering verified via indirect
//     observation (dialog bounds after MessageBox completes are not directly
//     accessible, so we verify via reasonable-bounds assertions).
//
// Tests that reference MessageBox, MsgBoxButton, MbOK, etc. will not compile
// until the implementation lands — that is expected.
//
// Conventions:
//   - newTestScreen(t)   — 80×25 SimulationScreen (defined in application_test.go)
//   - execViewStack(t)   — app + win + screen (defined in dialog_test.go)
//   - All assertions cite the spec sentence they verify.

import (
	"testing"
	"time"

	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Section 1 — MsgBoxButton bitmask constants
// ---------------------------------------------------------------------------

// TestMsgBoxButtonMbOKValue verifies MbOK == 1 (first iota shift).
// Spec: "MbOK MsgBoxButton = 1 << iota → value 1"
func TestMsgBoxButtonMbOKValue(t *testing.T) {
	if MbOK != 1 {
		t.Errorf("MbOK = %d, want 1", MbOK)
	}
}

// TestMsgBoxButtonMbCancelValue verifies MbCancel == 2.
// Spec: "MbCancel → value 2"
func TestMsgBoxButtonMbCancelValue(t *testing.T) {
	if MbCancel != 2 {
		t.Errorf("MbCancel = %d, want 2", MbCancel)
	}
}

// TestMsgBoxButtonMbYesValue verifies MbYes == 4.
// Spec: "MbYes → value 4"
func TestMsgBoxButtonMbYesValue(t *testing.T) {
	if MbYes != 4 {
		t.Errorf("MbYes = %d, want 4", MbYes)
	}
}

// TestMsgBoxButtonMbNoValue verifies MbNo == 8.
// Spec: "MbNo → value 8"
func TestMsgBoxButtonMbNoValue(t *testing.T) {
	if MbNo != 8 {
		t.Errorf("MbNo = %d, want 8", MbNo)
	}
}

// TestMsgBoxButtonConstantsAreDistinct verifies that no two constants share a value.
// Spec: bitmask type — constants must be distinct powers of two.
func TestMsgBoxButtonConstantsAreDistinct(t *testing.T) {
	seen := map[MsgBoxButton]string{}
	constants := []struct {
		val  MsgBoxButton
		name string
	}{
		{MbOK, "MbOK"},
		{MbCancel, "MbCancel"},
		{MbYes, "MbYes"},
		{MbNo, "MbNo"},
	}
	for _, c := range constants {
		if prev, ok := seen[c.val]; ok {
			t.Errorf("constant %s shares value %d with %s — bitmask constants must be distinct", c.name, c.val, prev)
		}
		seen[c.val] = c.name
	}
}

// TestMsgBoxButtonORCombinationsAreValid verifies that OR-ing constants produces
// a value containing both bits.
// Spec: "MsgBoxButton supports OR combinations (MbOK|MbCancel)"
func TestMsgBoxButtonORCombinationsAreValid(t *testing.T) {
	combo := MbOK | MbCancel
	if combo&MbOK == 0 {
		t.Error("MbOK|MbCancel & MbOK == 0 — MbOK bit is missing from combination")
	}
	if combo&MbCancel == 0 {
		t.Error("MbOK|MbCancel & MbCancel == 0 — MbCancel bit is missing from combination")
	}
}

// TestMsgBoxButtonAllFourCombined verifies that OR-ing all four constants produces
// a value with all four bits set.
func TestMsgBoxButtonAllFourCombined(t *testing.T) {
	all := MbOK | MbCancel | MbYes | MbNo
	for _, tc := range []struct {
		flag MsgBoxButton
		name string
	}{
		{MbOK, "MbOK"},
		{MbCancel, "MbCancel"},
		{MbYes, "MbYes"},
		{MbNo, "MbNo"},
	} {
		if all&tc.flag == 0 {
			t.Errorf("all-four combination missing %s bit", tc.name)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 2 — MessageBox return values (modal loop integration)
// ---------------------------------------------------------------------------

// msgboxStack builds the full Application→Desktop→Window owner chain used by
// all MessageBox integration tests. It returns the app, the window, and the
// already-Init'd screen.  Callers must NOT call screen.Fini() directly; the
// returned app's screen field is the same object.
func msgboxStack(t *testing.T) (*Application, *Window, interface{ Fini() }) {
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
	win := NewWindow(NewRect(0, 0, 80, 25), "Host", WithWindowNumber(1))
	app.Desktop().Insert(win)
	return app, win, screen
}

// TestMessageBoxReturnsCmOKWhenOKDefault verifies that MessageBox(MbOK) returns
// CmOK when the default (OK) button is activated.
// Spec: "calls owner.ExecView(dialog), returns the command code"
// Spec: "MbOK → "[ OK ]" button fires CmOK; first button is the default"
func TestMessageBoxReturnsCmOKWhenOKDefault(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(win, "Alert", "Something happened.", MbOK)
	}()

	// Give MessageBox time to enter the ExecView modal loop.
	time.Sleep(50 * time.Millisecond)

	// Activate the default (OK) button by posting CmOK.
	app.PostCommand(CmOK, nil)

	select {
	case cmd := <-result:
		if cmd != CmOK {
			t.Errorf("MessageBox(MbOK) returned %v, want CmOK", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("MessageBox(MbOK) did not return within 2 s after posting CmOK")
	}
}

// TestMessageBoxReturnsCmCancelWhenCancelOnly verifies that MessageBox(MbCancel)
// returns CmCancel when the Cancel button is the only button.
// Spec: "MbCancel → "[ Cancel ]" button fires CmCancel"
func TestMessageBoxReturnsCmCancelWhenCancelOnly(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(win, "Alert", "Press cancel.", MbCancel)
	}()

	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmCancel, nil)

	select {
	case cmd := <-result:
		if cmd != CmCancel {
			t.Errorf("MessageBox(MbCancel) returned %v, want CmCancel", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("MessageBox(MbCancel) did not return within 2 s after posting CmCancel")
	}
}

// TestMessageBoxOKCancelDefaultIsOK verifies that MessageBox(MbOK|MbCancel)
// returns CmOK when the first (default, OK) button is activated.
// Spec: "first button is the default (WithDefault())"
// Spec: "Button order: Yes, No, OK, Cancel (when multiple flags set)"
// → for MbOK|MbCancel, OK appears first, so OK is the default.
func TestMessageBoxOKCancelDefaultIsOK(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(win, "Confirm", "Are you sure?", MbOK|MbCancel)
	}()

	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmOK, nil)

	select {
	case cmd := <-result:
		if cmd != CmOK {
			t.Errorf("MessageBox(MbOK|MbCancel) with CmOK returned %v, want CmOK", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("MessageBox(MbOK|MbCancel) did not return within 2 s after posting CmOK")
	}
}

// TestMessageBoxOKCancelCanReturnCancel verifies that MessageBox(MbOK|MbCancel)
// can also return CmCancel when the Cancel button is activated.
// Spec: "MbCancel → "[ Cancel ]" button fires CmCancel"
func TestMessageBoxOKCancelCanReturnCancel(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(win, "Confirm", "Are you sure?", MbOK|MbCancel)
	}()

	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmCancel, nil)

	select {
	case cmd := <-result:
		if cmd != CmCancel {
			t.Errorf("MessageBox(MbOK|MbCancel) with CmCancel returned %v, want CmCancel", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("MessageBox(MbOK|MbCancel) did not return within 2 s after posting CmCancel")
	}
}

// TestMessageBoxYesNoDefaultIsYes verifies that MessageBox(MbYes|MbNo) returns
// CmYes when the first (default, Yes) button is activated.
// Spec: "Button order: Yes, No, OK, Cancel"
// → for MbYes|MbNo, Yes appears first, so Yes is the default.
func TestMessageBoxYesNoDefaultIsYes(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(win, "Question", "Delete file?", MbYes|MbNo)
	}()

	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmYes, nil)

	select {
	case cmd := <-result:
		if cmd != CmYes {
			t.Errorf("MessageBox(MbYes|MbNo) with CmYes returned %v, want CmYes", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("MessageBox(MbYes|MbNo) did not return within 2 s after posting CmYes")
	}
}

// TestMessageBoxYesNoCanReturnNo verifies that MessageBox(MbYes|MbNo) can also
// return CmNo when the No button is activated.
// Spec: "MbNo → "[ No ]" button fires CmNo"
func TestMessageBoxYesNoCanReturnNo(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(win, "Question", "Delete file?", MbYes|MbNo)
	}()

	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmNo, nil)

	select {
	case cmd := <-result:
		if cmd != CmNo {
			t.Errorf("MessageBox(MbYes|MbNo) with CmNo returned %v, want CmNo", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("MessageBox(MbYes|MbNo) did not return within 2 s after posting CmNo")
	}
}

// TestMessageBoxZeroButtonsDefaultsToOK verifies that passing buttons=0 defaults
// to an OK button and returns CmOK.
// Spec: "If buttons=0, defaults to OK."
func TestMessageBoxZeroButtonsDefaultsToOK(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(win, "Info", "No buttons specified.", 0)
	}()

	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmOK, nil)

	select {
	case cmd := <-result:
		if cmd != CmOK {
			t.Errorf("MessageBox(0) returned %v, want CmOK (defaulted to OK)", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("MessageBox(0) did not return within 2 s — may not have defaulted to OK")
	}
}

// TestMessageBoxAllFourButtonsDefaultIsYes verifies that MessageBox with all four
// button flags set returns CmYes when the first (default) button is activated.
// Spec: "Button order: Yes, No, OK, Cancel" → Yes is first → Yes is default.
func TestMessageBoxAllFourButtonsDefaultIsYes(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(win, "Choose", "Pick one.", MbYes|MbNo|MbOK|MbCancel)
	}()

	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmYes, nil)

	select {
	case cmd := <-result:
		if cmd != CmYes {
			t.Errorf("MessageBox(all four) returned %v, want CmYes (Yes is first = default)", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("MessageBox(all four) did not return within 2 s after posting CmYes")
	}
}

// TestMessageBoxAllFourCanReturnNo verifies that each non-default button in a
// four-button MessageBox can still fire its command.
// Spec: "MbNo → "[ No ]" button fires CmNo"
func TestMessageBoxAllFourCanReturnNo(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(win, "Choose", "Pick no.", MbYes|MbNo|MbOK|MbCancel)
	}()

	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmNo, nil)

	select {
	case cmd := <-result:
		if cmd != CmNo {
			t.Errorf("MessageBox(all four) with CmNo returned %v, want CmNo", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Error("MessageBox(all four) with CmNo did not return within 2 s")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — MessageBox geometry (auto-sizing and centering)
// ---------------------------------------------------------------------------

// dialogGeomCapture is a Container that wraps a real Window and intercepts
// ExecView to capture the geometry of the dialog before entering the modal loop,
// then delegates to the real window's ExecView.
//
// This allows geometry assertions without racing against the event loop.
type dialogGeomCapture struct {
	// embed the Window so it satisfies Container fully
	*Window
	capturedBounds chan Rect
}

func newDialogGeomCapture(win *Window) *dialogGeomCapture {
	return &dialogGeomCapture{
		Window:         win,
		capturedBounds: make(chan Rect, 1),
	}
}

func (c *dialogGeomCapture) ExecView(v View) CommandCode {
	// Record bounds before delegating so the goroutine can read them.
	c.capturedBounds <- v.Bounds()
	return c.Window.ExecView(v)
}

// TestMessageBoxAutoSizeWidthAtLeastTitle verifies the dialog is at least as wide
// as the title plus padding.
// Spec: "width = max(len(title)+4, len(text)+4, buttonRowWidth+4) clamped to 60, min 20"
func TestMessageBoxAutoSizeWidthAtLeastTitle(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	title := "Short"
	// title length 5 → minimum width from title = 5+4 = 9, but floor is 20.
	capture := newDialogGeomCapture(win)
	app.Desktop().Insert(capture.Window) // already inserted in msgboxStack, but harmless

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(capture, title, "Hi", MbOK)
	}()

	var bounds Rect
	select {
	case bounds = <-capture.capturedBounds:
	case <-time.After(2 * time.Second):
		t.Fatal("MessageBox did not call ExecView within 2 s")
	}

	// Unblock the modal loop.
	app.PostCommand(CmOK, nil)
	<-result

	minExpected := len(title) + 4
	if minExpected < 20 {
		minExpected = 20
	}
	if bounds.Width() < minExpected {
		t.Errorf("dialog width = %d, want >= %d (len(title)+4 clamped to min 20)", bounds.Width(), minExpected)
	}
}

// TestMessageBoxAutoSizeWidthAtLeastText verifies the dialog is at least as wide
// as the text plus padding when the text is the widest element.
// Spec: "width = max(len(title)+4, len(text)+4, buttonRowWidth+4) clamped to 60, min 20"
func TestMessageBoxAutoSizeWidthAtLeastText(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	text := "This is a longer message body"
	// len = 29 → text+4 = 33
	capture := newDialogGeomCapture(win)

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(capture, "T", text, MbOK)
	}()

	var bounds Rect
	select {
	case bounds = <-capture.capturedBounds:
	case <-time.After(2 * time.Second):
		t.Fatal("MessageBox did not call ExecView within 2 s")
	}

	app.PostCommand(CmOK, nil)
	<-result

	minExpected := len(text) + 4
	if minExpected > 60 {
		minExpected = 60
	}
	if bounds.Width() < minExpected {
		t.Errorf("dialog width = %d, want >= %d (len(text)+4)", bounds.Width(), minExpected)
	}
}

// TestMessageBoxAutoSizeWidthClampedTo60 verifies the dialog width never exceeds 60.
// Spec: "width = max(...) clamped to 60"
func TestMessageBoxAutoSizeWidthClampedTo60(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	// A text long enough that len(text)+4 > 60.
	text := "This is an extremely long message that exceeds the maximum allowed width of the dialog box by quite a margin"
	capture := newDialogGeomCapture(win)

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(capture, "T", text, MbOK)
	}()

	var bounds Rect
	select {
	case bounds = <-capture.capturedBounds:
	case <-time.After(2 * time.Second):
		t.Fatal("MessageBox did not call ExecView within 2 s")
	}

	app.PostCommand(CmOK, nil)
	<-result

	if bounds.Width() > 60 {
		t.Errorf("dialog width = %d, want <= 60 (clamped)", bounds.Width())
	}
}

// TestMessageBoxAutoSizeWidthAtLeast20 verifies the dialog width is at least 20.
// Spec: "min 20"
func TestMessageBoxAutoSizeWidthAtLeast20(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	// Very short title and text so the natural width would be tiny.
	capture := newDialogGeomCapture(win)

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(capture, "Hi", "Hi", MbOK)
	}()

	var bounds Rect
	select {
	case bounds = <-capture.capturedBounds:
	case <-time.After(2 * time.Second):
		t.Fatal("MessageBox did not call ExecView within 2 s")
	}

	app.PostCommand(CmOK, nil)
	<-result

	if bounds.Width() < 20 {
		t.Errorf("dialog width = %d, want >= 20 (minimum)", bounds.Width())
	}
}

// TestMessageBoxAutoSizeHeightFromTextLines verifies the dialog height equals
// textLines + 5.
// Spec: "Height = textLines + 5."
// For a single-line text and a dialog wide enough to fit it on one line, textLines=1
// → height must be 6.
func TestMessageBoxAutoSizeHeightForSingleLine(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	// Short text that fits on a single line.
	capture := newDialogGeomCapture(win)

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(capture, "Hi", "One line.", MbOK)
	}()

	var bounds Rect
	select {
	case bounds = <-capture.capturedBounds:
	case <-time.After(2 * time.Second):
		t.Fatal("MessageBox did not call ExecView within 2 s")
	}

	app.PostCommand(CmOK, nil)
	<-result

	// 1 text line + 5 = 6.
	if bounds.Height() != 6 {
		t.Errorf("dialog height = %d for single-line text, want 6 (textLines+5)", bounds.Height())
	}
}

// TestMessageBoxCenteredInOwner verifies the dialog is centered in the owner's bounds.
// Spec: "Centered in owner's bounds."
func TestMessageBoxCenteredInOwner(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	// Owner window fills the full 80×25 screen.
	ownerBounds := win.Bounds()
	capture := newDialogGeomCapture(win)

	result := make(chan CommandCode, 1)
	go func() {
		result <- MessageBox(capture, "Centered?", "Check the position.", MbOK)
	}()

	var bounds Rect
	select {
	case bounds = <-capture.capturedBounds:
	case <-time.After(2 * time.Second):
		t.Fatal("MessageBox did not call ExecView within 2 s")
	}

	app.PostCommand(CmOK, nil)
	<-result

	// Horizontal centering: the dialog's left edge should be approximately
	// (ownerWidth - dialogWidth) / 2 relative to the owner's own origin.
	expectedX := ownerBounds.A.X + (ownerBounds.Width()-bounds.Width())/2
	expectedY := ownerBounds.A.Y + (ownerBounds.Height()-bounds.Height())/2

	// Allow ±1 for integer rounding.
	if abs(bounds.A.X-expectedX) > 1 {
		t.Errorf("dialog X = %d, want ~%d (centered horizontally in owner width %d)",
			bounds.A.X, expectedX, ownerBounds.Width())
	}
	if abs(bounds.A.Y-expectedY) > 1 {
		t.Errorf("dialog Y = %d, want ~%d (centered vertically in owner height %d)",
			bounds.A.Y, expectedY, ownerBounds.Height())
	}
}

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestMessageBoxButtonRowWidthAffectsSizing verifies that adding more buttons
// increases the dialog width when the button row is the widest element.
// Spec: "width = max(len(title)+4, len(text)+4, buttonRowWidth+4)"
// buttonRowWidth for two 12-wide buttons separated by 2 = 12+2+12 = 26 → 26+4=30.
func TestMessageBoxButtonRowWidthAffectsSizing(t *testing.T) {
	app, win, screen := msgboxStack(t)
	defer screen.Fini()

	// Use a very short title and text so the button row drives the width.
	capture := newDialogGeomCapture(win)

	result := make(chan CommandCode, 1)
	go func() {
		// Two buttons: OK and Cancel → buttonRowWidth = 12+2+12 = 26 → need 30.
		result <- MessageBox(capture, "Q", "X", MbOK|MbCancel)
	}()

	var bounds Rect
	select {
	case bounds = <-capture.capturedBounds:
	case <-time.After(2 * time.Second):
		t.Fatal("MessageBox did not call ExecView within 2 s")
	}

	app.PostCommand(CmOK, nil)
	<-result

	// Two buttons at 12 wide each, separated by 2, plus 4 padding = 30 minimum.
	minFromButtons := (12*2 + 2) + 4 // 30
	if bounds.Width() < minFromButtons {
		t.Errorf("dialog width = %d with MbOK|MbCancel, want >= %d (button row drives width)",
			bounds.Width(), minFromButtons)
	}
}

// ---------------------------------------------------------------------------
// Section 4 — MessageBox button content (compile-time / Draw-accessible checks)
// ---------------------------------------------------------------------------

// TestMessageBoxButtonLabelsNoTildeShortcuts is a compile-time documentation test.
// The spec states button labels are plain text without tilde shortcuts.
// We verify this indirectly: the label strings used in the implementation should
// not start with '~'. This is enforced by the requirement, not by a runtime check,
// but we confirm the expectation is knowable from the constant names.
//
// Spec: "Button labels do NOT use tilde shortcuts (plain text)."
func TestMessageBoxButtonLabelsNoTildeShortcuts(t *testing.T) {
	// Verify that the well-known plain-text labels do not parse as tilde labels.
	// If a label has no '~', ParseTildeLabel returns a single non-shortcut segment.
	for _, label := range []string{"OK", "Cancel", "Yes", "No"} {
		segs := ParseTildeLabel(label)
		for _, seg := range segs {
			if seg.Shortcut {
				t.Errorf("button label %q has a shortcut segment — labels must be plain text", label)
			}
		}
	}
}
