package tv

// integration_batch1_label_test.go — Integration tests for the Label validation
// batch (Tasks 1–3).
//
// These tests exercise all five requirements end-to-end using real components
// (InputLine, Button, Group) wired together the way a real dialog would be.
// No mocks.
//
// Requirements:
//   Req 1: Label+InputLine in a Group — Alt+shortcut focuses InputLine via full
//           Group → preprocess → Label.HandleEvent → SetFocusedChild chain.
//   Req 2: Label light state tracks InputLine focus — LabelHighlight when focused,
//           LabelNormal when not.
//   Req 3: Label linked to a disabled (non-selectable) view — Alt+shortcut clears
//           event but does NOT change focus.
//   Req 4: Column 0 of a drawn Label is always a space; text starts at column 1;
//           trailing columns are filled with the appropriate background style.
//   Req 5: Two Labels in the same Group with different shortcuts both work
//           independently.
//
// Test naming: TestIntegrationBatch1Label_<DescriptiveSuffix>

import (
	"testing"

	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Requirement 1: Alt+shortcut focuses InputLine through the full dispatch chain.
// ---------------------------------------------------------------------------

// TestIntegrationBatch1Label_AltShortcutFocusesLinkedInputLine verifies that
// pressing Alt+N in a Group containing a "~N~ame:" Label linked to an InputLine
// focuses the InputLine — exercising the full event dispatch path:
//
//	Group.HandleEvent → preprocess phase → Label.HandleEvent → SetFocusedChild
func TestIntegrationBatch1Label_AltShortcutFocusesLinkedInputLine(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	input := NewInputLine(NewRect(10, 0, 20, 1), 80)
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", input)

	// A second selectable widget so that the InputLine is not focused initially.
	other := newSelectableMockView(NewRect(0, 5, 20, 1))

	g.Insert(label)
	g.Insert(input)
	g.Insert(other) // last insert — other becomes the focused child

	if g.FocusedChild() != other {
		t.Fatalf("precondition: FocusedChild() = %v, want other", g.FocusedChild())
	}

	ev := altKeyEvent('n')
	g.HandleEvent(ev)

	if g.FocusedChild() != input {
		t.Errorf("after Alt+N: FocusedChild() = %v, want InputLine", g.FocusedChild())
	}
}

// TestIntegrationBatch1Label_AltShortcutEventClearedAfterFocusingInputLine
// verifies that the Alt+N keyboard event is cleared after Label.HandleEvent
// processes it, so it does not propagate further.
func TestIntegrationBatch1Label_AltShortcutEventClearedAfterFocusingInputLine(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	input := NewInputLine(NewRect(10, 0, 20, 1), 80)
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", input)
	other := newSelectableMockView(NewRect(0, 5, 20, 1))

	g.Insert(label)
	g.Insert(input)
	g.Insert(other)

	ev := altKeyEvent('n')
	g.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("after Alt+N: event not cleared (What = %v, want EvNothing)", ev.What)
	}
}

// TestIntegrationBatch1Label_AltShortcutFocusesLinkedButton verifies the same
// full dispatch path works when the linked view is a Button rather than an
// InputLine — confirming this is not InputLine-specific.
func TestIntegrationBatch1Label_AltShortcutFocusesLinkedButton(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	btn := NewButton(NewRect(10, 0, 12, 2), "Save", CmOK)
	label := NewLabel(NewRect(0, 0, 10, 1), "~S~ave", btn)

	// Insert an InputLine that becomes the focused widget initially.
	other := NewInputLine(NewRect(0, 5, 20, 1), 80)

	g.Insert(label)
	g.Insert(btn)
	g.Insert(other)

	if g.FocusedChild() != other {
		t.Fatalf("precondition: FocusedChild() = %v, want other (InputLine)", g.FocusedChild())
	}

	ev := altKeyEvent('s')
	g.HandleEvent(ev)

	if g.FocusedChild() != btn {
		t.Errorf("after Alt+S: FocusedChild() = %v, want Button", g.FocusedChild())
	}
}

// TestIntegrationBatch1Label_AltShortcutPreprocessFiresBeforeFocusedChild verifies
// that when there is a focused widget that would consume Alt+N (if it received it),
// the Label still intercepts in the preprocess phase and the focused child never
// sees the event.
func TestIntegrationBatch1Label_AltShortcutPreprocessFiresBeforeFocusedChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	input := NewInputLine(NewRect(10, 0, 20, 1), 80)
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", input)

	// An interceptor that would consume Alt+N if it saw it.
	interceptor := &altInterceptorView{intercept: 'n'}
	interceptor.SetBounds(NewRect(0, 5, 20, 1))
	interceptor.SetState(SfVisible, true)
	interceptor.SetOptions(OfSelectable, true)

	g.Insert(label)
	g.Insert(input)
	g.Insert(interceptor)

	if g.FocusedChild() != interceptor {
		t.Fatalf("precondition: FocusedChild() = %v, want interceptor", g.FocusedChild())
	}

	ev := altKeyEvent('n')
	g.HandleEvent(ev)

	// Label (preprocess) must win, focusing input.
	if g.FocusedChild() != input {
		t.Errorf("after Alt+N: FocusedChild() = %v, want InputLine; Label preprocess must fire first", g.FocusedChild())
	}
	// The interceptor must never have seen the event.
	if interceptor.gotEvent {
		t.Errorf("interceptor received Alt+N; Label must have consumed it in the preprocess phase")
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: Label light state tracks linked InputLine focus.
// ---------------------------------------------------------------------------

// TestIntegrationBatch1Label_LightSetWhenInputLineFocusedViaAltShortcut verifies
// that after Alt+N focuses the InputLine, the Label's light field becomes true and
// Draw uses LabelHighlight.
func TestIntegrationBatch1Label_LightSetWhenInputLineFocusedViaAltShortcut(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	scheme := theme.BorlandBlue

	input := NewInputLine(NewRect(10, 0, 20, 1), 80)
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", input)
	label.scheme = scheme

	other := newSelectableMockView(NewRect(0, 5, 20, 1))
	g.Insert(label)
	g.Insert(input)
	g.Insert(other)

	// Precondition: label is not lit before the shortcut fires.
	if label.light {
		t.Fatalf("precondition: label.light must be false before InputLine gains focus")
	}

	// Alt+N focuses InputLine, which triggers CmReceivedFocus broadcast.
	g.HandleEvent(altKeyEvent('n'))

	if !label.light {
		t.Errorf("after Alt+N focuses InputLine: label.light = false, want true")
	}

	// Draw should use LabelHighlight for the background/normal segments.
	buf := NewDrawBuffer(10, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.LabelHighlight {
		t.Errorf("after highlight: cell(0,0).Style = %v, want LabelHighlight %v", cell.Style, scheme.LabelHighlight)
	}
}

// TestIntegrationBatch1Label_LightClearedWhenInputLineLosesFocus verifies that
// when focus moves away from the InputLine (to another widget via Tab / direct
// SetFocusedChild), the Label's light field becomes false and Draw switches to
// LabelNormal.
func TestIntegrationBatch1Label_LightClearedWhenInputLineLosesFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	scheme := theme.BorlandBlue

	input := NewInputLine(NewRect(10, 0, 20, 1), 80)
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", input)
	label.scheme = scheme

	other := newSelectableMockView(NewRect(0, 5, 20, 1))
	g.Insert(label)
	g.Insert(input)
	g.Insert(other)

	// Step 1: Focus InputLine — label lights up.
	g.SetFocusedChild(input)
	if !label.light {
		t.Fatalf("precondition: label.light must be true after InputLine receives focus")
	}

	// Step 2: Move focus to other widget.
	g.SetFocusedChild(other)

	if label.light {
		t.Errorf("after InputLine loses focus: label.light = true, want false")
	}

	// Draw should now use LabelNormal.
	buf := NewDrawBuffer(10, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.LabelNormal {
		t.Errorf("after dim: cell(0,0).Style = %v, want LabelNormal %v", cell.Style, scheme.LabelNormal)
	}
}

// TestIntegrationBatch1Label_LightTracksTabNavigation verifies that when focus
// is moved away from the InputLine via Tab in an enclosing Window (which owns Tab
// handling per spec 13.3), the label's light field tracks the change correctly.
func TestIntegrationBatch1Label_LightTracksTabNavigation(t *testing.T) {
	// Wrap in a Window — Tab cycling is handled at Window level (spec 13.3).
	w := NewWindow(NewRect(0, 0, 80, 25), "Dialog")

	input := NewInputLine(NewRect(10, 0, 20, 1), 80)
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", input)
	other := newSelectableMockView(NewRect(0, 5, 20, 1))

	w.Insert(label)
	w.Insert(input)
	w.Insert(other)

	// Focus input explicitly so the label lights up.
	w.SetFocusedChild(input)
	if !label.light {
		t.Fatalf("precondition: label.light must be true after InputLine gains focus")
	}
	if w.FocusedChild() != input {
		t.Fatalf("precondition: FocusedChild() = %v, want input", w.FocusedChild())
	}

	// Tab via the Window cycles focus to the next selectable widget.
	w.HandleEvent(tabEvent())

	// Focus must have moved away from input.
	if w.FocusedChild() == input {
		t.Fatalf("after Tab in Window: FocusedChild() = input, focus should have advanced")
	}

	// Label light must be cleared because input lost focus.
	if label.light {
		t.Errorf("after Tab moves focus away from InputLine: label.light = true, want false")
	}
}

// TestIntegrationBatch1Label_LightNotSetForUnrelatedFocusChange verifies that
// a focus change to a widget OTHER than the Label's link does not set label.light.
func TestIntegrationBatch1Label_LightNotSetForUnrelatedFocusChange(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	input := NewInputLine(NewRect(10, 0, 20, 1), 80)
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", input)

	btn := NewButton(NewRect(0, 5, 12, 2), "OK", CmOK)

	g.Insert(label)
	g.Insert(input)
	g.Insert(btn)

	// Focus btn — CmReceivedFocus Info=btn (not input).
	g.SetFocusedChild(btn)

	if label.light {
		t.Errorf("after focusing Button (not linked InputLine): label.light = true, want false")
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: Label linked to a non-selectable view — event cleared, no focus change.
// ---------------------------------------------------------------------------

// TestIntegrationBatch1Label_NonSelectableLinkNoFocusChange verifies that when
// the linked view has OfSelectable cleared, Alt+shortcut does NOT change focus.
func TestIntegrationBatch1Label_NonSelectableLinkNoFocusChange(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// A non-selectable "linked" view — e.g. a display label or disabled input.
	disabledInput := NewInputLine(NewRect(10, 0, 20, 1), 80)
	disabledInput.SetOptions(OfSelectable, false)

	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", disabledInput)
	other := newSelectableMockView(NewRect(0, 5, 20, 1))

	g.Insert(label)
	g.Insert(disabledInput)
	g.Insert(other) // other becomes focused

	if g.FocusedChild() != other {
		t.Fatalf("precondition: FocusedChild() = %v, want other", g.FocusedChild())
	}

	ev := altKeyEvent('n')
	g.HandleEvent(ev)

	// Focus must stay on other.
	if g.FocusedChild() != other {
		t.Errorf("after Alt+N with non-selectable link: FocusedChild() = %v, want other (unchanged)", g.FocusedChild())
	}
	// Specifically, focus must not have moved to disabledInput.
	if g.FocusedChild() == disabledInput {
		t.Errorf("focus moved to non-selectable disabledInput; must not happen")
	}
}

// TestIntegrationBatch1Label_NonSelectableLinkEventCleared verifies that even
// when the linked view is non-selectable, the Alt+shortcut event IS still cleared
// (the event is consumed, focus is just not changed).
func TestIntegrationBatch1Label_NonSelectableLinkEventCleared(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	disabledInput := NewInputLine(NewRect(10, 0, 20, 1), 80)
	disabledInput.SetOptions(OfSelectable, false)

	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", disabledInput)
	other := newSelectableMockView(NewRect(0, 5, 20, 1))

	g.Insert(label)
	g.Insert(disabledInput)
	g.Insert(other)

	ev := altKeyEvent('n')
	g.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("after Alt+N with non-selectable link: event not cleared (What = %v); must be consumed", ev.What)
	}
}

// TestIntegrationBatch1Label_NonSelectableLinkEventClearedFalsification is a
// stricter falsification guard: the event must be cleared regardless of the
// OfSelectable guard outcome — an implementation that skips Clear() when the
// guard fails would be caught here.
func TestIntegrationBatch1Label_NonSelectableLinkEventClearedFalsification(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// Use a plain stub (labelLinkedView) with OfSelectable explicitly removed.
	stub := newLabelLinkedView()
	stub.SetOptions(OfSelectable, false)

	label := NewLabel(NewRect(0, 0, 10, 1), "~F~oo:", stub)
	anchor := newSelectableMockView(NewRect(0, 5, 20, 1))

	g.Insert(label)
	g.Insert(stub)
	g.Insert(anchor)

	beforeFocused := g.FocusedChild()

	ev := altKeyEvent('f')
	g.HandleEvent(ev)

	// Event must be cleared.
	if !ev.IsCleared() {
		t.Errorf("event not cleared when link is non-selectable; Label must consume the event even when focus does not move")
	}
	// Focus must be unchanged.
	if g.FocusedChild() != beforeFocused {
		t.Errorf("focus changed from %v to %v; non-selectable link must not receive focus", beforeFocused, g.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Draw layout — column 0 is margin space, text at column 1,
//                trailing columns filled with background style.
// ---------------------------------------------------------------------------

// TestIntegrationBatch1Label_DrawCol0IsSpaceInRealGroup verifies that column 0
// of a Label drawn inside a real Group is always a space (the margin column).
// This tests the full path: Label inserted in Group, scheme inherited via owner chain.
func TestIntegrationBatch1Label_DrawCol0IsSpaceInRealGroup(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	scheme := theme.BorlandBlue
	// Set the scheme on the group so the label inherits it via Owner chain.
	g.scheme = scheme

	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame:", nil)
	g.Insert(label)

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Errorf("cell(0,0).Rune = %q, want ' '; column 0 must always be the margin space", cell.Rune)
	}
}

// TestIntegrationBatch1Label_DrawCol0IsNotFirstTextChar falsifies an
// implementation that starts text at column 0, displacing the margin.
func TestIntegrationBatch1Label_DrawCol0IsNotFirstTextChar(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	g.scheme = theme.BorlandBlue

	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame:", nil)
	g.Insert(label)

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == 'N' {
		t.Errorf("cell(0,0).Rune = 'N'; text must not start at column 0 — margin space required")
	}
}

// TestIntegrationBatch1Label_DrawTextStartsAtColumn1 verifies the first
// text rune of a label drawn in a Group appears at column 1.
func TestIntegrationBatch1Label_DrawTextStartsAtColumn1(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	g.scheme = theme.BorlandBlue

	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame:", nil)
	g.Insert(label)

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(1, 0)
	if cell.Rune != 'N' {
		t.Errorf("cell(1,0).Rune = %q, want 'N'; first text rune must be at column 1", cell.Rune)
	}
}

// TestIntegrationBatch1Label_DrawTrailingColumnsFilledWithNormalStyle verifies
// that columns beyond the text are filled with spaces in LabelNormal style (not
// the zero-value/default style).
func TestIntegrationBatch1Label_DrawTrailingColumnsFilledWithNormalStyle(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	scheme := theme.BorlandBlue
	g.scheme = scheme

	// "~N~" → 1 text rune; bounds width = 20; columns 2..19 must be filled.
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~", nil)
	g.Insert(label)

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	for _, col := range []int{2, 10, 19} {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("cell(%d,0).Rune = %q, want ' '; trailing columns must be space-filled", col, cell.Rune)
		}
		if cell.Style != scheme.LabelNormal {
			t.Errorf("cell(%d,0).Style = %v, want LabelNormal %v; trailing fill must use current style", col, cell.Style, scheme.LabelNormal)
		}
	}
}

// TestIntegrationBatch1Label_DrawTrailingColumnsFilledWithHighlightStyle verifies
// that when the linked InputLine is focused (light=true), trailing columns use
// LabelHighlight rather than LabelNormal.
func TestIntegrationBatch1Label_DrawTrailingColumnsFilledWithHighlightStyle(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.LabelHighlight == scheme.LabelNormal {
		t.Skip("LabelHighlight == LabelNormal in BorlandBlue; distinction test is vacuous")
	}

	g := NewGroup(NewRect(0, 0, 80, 25))
	g.scheme = scheme

	input := NewInputLine(NewRect(10, 0, 20, 1), 80)
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~", input)
	other := newSelectableMockView(NewRect(0, 5, 20, 1))

	g.Insert(label)
	g.Insert(input)
	g.Insert(other)

	// Focus the InputLine so the label lights up.
	g.SetFocusedChild(input)
	if !label.light {
		t.Fatalf("precondition: label.light must be true after InputLine gains focus")
	}

	buf := NewDrawBuffer(10, 1)
	label.Draw(buf)

	// Column 2 onward is trailing fill — must be LabelHighlight.
	cell := buf.GetCell(5, 0)
	if cell.Style != scheme.LabelHighlight {
		t.Errorf("cell(5,0).Style = %v, want LabelHighlight %v when label is lit", cell.Style, scheme.LabelHighlight)
	}
	if cell.Style == scheme.LabelNormal {
		t.Errorf("cell(5,0).Style = LabelNormal; highlighted label must NOT use LabelNormal for trailing fill")
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: Two Labels with different shortcuts in the same Group both work.
// ---------------------------------------------------------------------------

// TestIntegrationBatch1Label_TwoLabelsAltNFocusesFirstLink verifies that in a
// Group with two Label+InputLine pairs (shortcuts N and S), pressing Alt+N
// focuses the InputLine linked to the N-label (not the S-label's link).
func TestIntegrationBatch1Label_TwoLabelsAltNFocusesFirstLink(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	inputN := NewInputLine(NewRect(10, 0, 20, 1), 80)
	labelN := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", inputN)

	inputS := NewInputLine(NewRect(10, 2, 20, 1), 80)
	labelS := NewLabel(NewRect(0, 2, 10, 1), "~S~urname:", inputS)

	// Insert a third widget so neither InputLine starts focused.
	anchor := newSelectableMockView(NewRect(0, 5, 20, 1))

	g.Insert(labelN)
	g.Insert(inputN)
	g.Insert(labelS)
	g.Insert(inputS)
	g.Insert(anchor)

	if g.FocusedChild() != anchor {
		t.Fatalf("precondition: FocusedChild() = %v, want anchor", g.FocusedChild())
	}

	// Alt+N must focus inputN.
	g.HandleEvent(altKeyEvent('n'))

	if g.FocusedChild() != inputN {
		t.Errorf("Alt+N: FocusedChild() = %v, want inputN (linked to ~N~ label)", g.FocusedChild())
	}
	// inputS must not have received focus.
	if g.FocusedChild() == inputS {
		t.Errorf("Alt+N focused inputS instead of inputN")
	}
}

// TestIntegrationBatch1Label_TwoLabelsAltSFocusesSecondLink verifies that in
// the same dual-label Group, pressing Alt+S focuses the InputLine linked to the
// S-label — confirming the two shortcuts operate independently.
func TestIntegrationBatch1Label_TwoLabelsAltSFocusesSecondLink(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	inputN := NewInputLine(NewRect(10, 0, 20, 1), 80)
	labelN := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", inputN)

	inputS := NewInputLine(NewRect(10, 2, 20, 1), 80)
	labelS := NewLabel(NewRect(0, 2, 10, 1), "~S~urname:", inputS)

	anchor := newSelectableMockView(NewRect(0, 5, 20, 1))

	g.Insert(labelN)
	g.Insert(inputN)
	g.Insert(labelS)
	g.Insert(inputS)
	g.Insert(anchor)

	if g.FocusedChild() != anchor {
		t.Fatalf("precondition: FocusedChild() = %v, want anchor", g.FocusedChild())
	}

	// Alt+S must focus inputS.
	g.HandleEvent(altKeyEvent('s'))

	if g.FocusedChild() != inputS {
		t.Errorf("Alt+S: FocusedChild() = %v, want inputS (linked to ~S~ label)", g.FocusedChild())
	}
	if g.FocusedChild() == inputN {
		t.Errorf("Alt+S focused inputN instead of inputS")
	}
}

// TestIntegrationBatch1Label_TwoLabelsAltNThenAltSTogglesFocus verifies the
// full round-trip: Alt+N → inputN focused, then Alt+S → inputS focused.
func TestIntegrationBatch1Label_TwoLabelsAltNThenAltSTogglesFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	inputN := NewInputLine(NewRect(10, 0, 20, 1), 80)
	labelN := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", inputN)

	inputS := NewInputLine(NewRect(10, 2, 20, 1), 80)
	labelS := NewLabel(NewRect(0, 2, 10, 1), "~S~urname:", inputS)

	anchor := newSelectableMockView(NewRect(0, 5, 20, 1))

	g.Insert(labelN)
	g.Insert(inputN)
	g.Insert(labelS)
	g.Insert(inputS)
	g.Insert(anchor)

	// Step 1: Alt+N → inputN.
	g.HandleEvent(altKeyEvent('n'))

	if g.FocusedChild() != inputN {
		t.Fatalf("after Alt+N: FocusedChild() = %v, want inputN", g.FocusedChild())
	}

	// Step 2: Alt+S → inputS.
	g.HandleEvent(altKeyEvent('s'))

	if g.FocusedChild() != inputS {
		t.Errorf("after Alt+N then Alt+S: FocusedChild() = %v, want inputS", g.FocusedChild())
	}
}

// TestIntegrationBatch1Label_TwoLabelsLightTracksCorrectLink verifies that when
// two Labels and two InputLines share a Group, each Label's light state tracks
// only its own link — not the other Label's link.
func TestIntegrationBatch1Label_TwoLabelsLightTracksCorrectLink(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	inputN := NewInputLine(NewRect(10, 0, 20, 1), 80)
	labelN := NewLabel(NewRect(0, 0, 10, 1), "~N~ame:", inputN)

	inputS := NewInputLine(NewRect(10, 2, 20, 1), 80)
	labelS := NewLabel(NewRect(0, 2, 10, 1), "~S~urname:", inputS)

	g.Insert(labelN)
	g.Insert(inputN)
	g.Insert(labelS)
	g.Insert(inputS)

	// Focus inputN — labelN must light up, labelS must stay dark.
	g.SetFocusedChild(inputN)

	if !labelN.light {
		t.Errorf("after inputN focused: labelN.light = false, want true")
	}
	if labelS.light {
		t.Errorf("after inputN focused: labelS.light = true, want false (unrelated link)")
	}

	// Now focus inputS — labelS must light up, labelN must dim.
	g.SetFocusedChild(inputS)

	if labelN.light {
		t.Errorf("after inputS focused: labelN.light = true, want false")
	}
	if !labelS.light {
		t.Errorf("after inputS focused: labelS.light = false, want true")
	}
}

// altKeyEvent is defined in desktop_window_mgmt_test.go (same package).
// It constructs an Alt+<rune> EvKeyboard event.
