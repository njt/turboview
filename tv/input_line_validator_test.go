package tv

// input_line_validator_test.go — Tests for InputLine validator integration.
//
// Tests are organized by spec section. Each test cites the behaviour it verifies.
//
// Validators used: FilterValidator (character whitelist), RangeValidator (numeric range).
// Both are real, already-tested types; they are used here only as test inputs.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// 1. SetValidator / Validator getter-setter
// ---------------------------------------------------------------------------

// TestInputLineValidatorGetSetNil verifies Validator() returns nil for a new InputLine.
// Spec: "Validator() returns the stored validator (nil if none)."
func TestInputLineValidatorGetSetNil(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	if v := il.Validator(); v != nil {
		t.Errorf("Validator() on new InputLine = %v, want nil", v)
	}
}

// TestInputLineValidatorSetAndGet verifies SetValidator stores a validator and Validator() returns it.
// Spec: "SetValidator(v Validator) stores a validator."
func TestInputLineValidatorSetAndGet(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	v := NewFilterValidator("abc")
	il.SetValidator(v)
	if got := il.Validator(); got != v {
		t.Errorf("Validator() after SetValidator = %v, want %v", got, v)
	}
}

// TestInputLineValidatorReplaceable verifies SetValidator can replace an existing validator.
// Spec: "SetValidator(v Validator) stores a validator."
func TestInputLineValidatorReplaceable(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	v1 := NewFilterValidator("abc")
	v2 := NewRangeValidator(0, 100)
	il.SetValidator(v1)
	il.SetValidator(v2)
	if got := il.Validator(); got != v2 {
		t.Errorf("Validator() after replacing validator = %v, want %v", got, v2)
	}
}

// TestInputLineValidatorClearable verifies SetValidator(nil) clears the validator.
// Spec: "Validator() returns the stored validator (nil if none)."
func TestInputLineValidatorClearable(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewFilterValidator("abc"))
	il.SetValidator(nil)
	if v := il.Validator(); v != nil {
		t.Errorf("Validator() after SetValidator(nil) = %v, want nil", v)
	}
}

// ---------------------------------------------------------------------------
// 2. Character insert accepted by validator
// ---------------------------------------------------------------------------

// TestInputLineValidatorInsertAccepted verifies that a character accepted by the
// validator is inserted normally.
// Spec: "calls validator.IsValidInput(newText, false) after mutation. If true, mutation stands."
func TestInputLineValidatorInsertAccepted(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewFilterValidator("abc"))
	il.HandleEvent(runeEv('a'))
	if got := il.Text(); got != "a" {
		t.Errorf("Text() after accepted insert = %q, want %q", got, "a")
	}
	if pos := il.CursorPos(); pos != 1 {
		t.Errorf("CursorPos() after accepted insert = %d, want 1", pos)
	}
}

// TestInputLineValidatorInsertAcceptedCursorAdvances verifies cursor advances when
// insert is accepted.
// Spec: mutation stands — full state (including cursor) reflects the new character.
func TestInputLineValidatorInsertAcceptedCursorAdvances(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewFilterValidator("abc"))
	il.HandleEvent(runeEv('b'))
	il.HandleEvent(runeEv('c'))
	if pos := il.CursorPos(); pos != 2 {
		t.Errorf("CursorPos() after two accepted inserts = %d, want 2", pos)
	}
}

// ---------------------------------------------------------------------------
// 3. Character insert rejected by validator — text and cursor unchanged
// ---------------------------------------------------------------------------

// TestInputLineValidatorInsertRejected verifies that a character rejected by the
// validator leaves the text unchanged.
// Spec: "If false, the entire state (text, cursorPos, selStart, selEnd, scrollOffset)
// is rolled back."
func TestInputLineValidatorInsertRejected(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("ab")
	il.SetValidator(NewFilterValidator("abc")) // 'd' is not in the whitelist
	il.HandleEvent(runeEv('d'))
	if got := il.Text(); got != "ab" {
		t.Errorf("Text() after rejected insert = %q, want %q (unchanged)", got, "ab")
	}
}

// TestInputLineValidatorInsertRejectedCursorUnchanged verifies the cursor is not
// advanced when an insert is rejected.
// Spec: "the entire state (text, cursorPos, …) is rolled back."
func TestInputLineValidatorInsertRejectedCursorUnchanged(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("ab") // cursor at 2
	il.SetValidator(NewFilterValidator("abc"))
	il.HandleEvent(runeEv('d'))
	if pos := il.CursorPos(); pos != 2 {
		t.Errorf("CursorPos() after rejected insert = %d, want 2 (unchanged)", pos)
	}
}

// TestInputLineValidatorInsertRejectedEventConsumed verifies the rejected keystroke
// is still consumed even though nothing changed.
// Spec: "The rejected keystroke is still consumed (event cleared)."
func TestInputLineValidatorInsertRejectedEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewFilterValidator("abc"))
	ev := runeEv('z')
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("rejected insert event was not consumed (IsCleared() = false)")
	}
}

// TestInputLineValidatorInsertRejectedSelectionRestored verifies that selection
// state is also rolled back when an insert is rejected.
// Spec: "the entire state (text, cursorPos, selStart, selEnd, scrollOffset) is rolled back."
func TestInputLineValidatorInsertRejectedSelectionRestored(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	// Create a selection covering "ab" by using Shift+Home from pos 3.
	il.HandleEvent(shiftKeyEv(tcell.KeyHome))
	selStart, selEnd := il.Selection()
	// Now reject an insert — selection state must remain as it was.
	il.SetValidator(NewFilterValidator("abc")) // 'd' is not accepted
	il.HandleEvent(runeEv('d'))
	gotStart, gotEnd := il.Selection()
	if gotStart != selStart || gotEnd != selEnd {
		t.Errorf("Selection() after rejected insert = (%d,%d), want (%d,%d) (rolled back)",
			gotStart, gotEnd, selStart, selEnd)
	}
}

// ---------------------------------------------------------------------------
// 4. Backspace rejected by validator — text and cursor restored
// ---------------------------------------------------------------------------

// TestInputLineValidatorBackspaceRejected verifies that backspace is rolled back
// when the resulting text fails validator.IsValidInput.
// Spec: "Keystroke rejection applies to ALL text-mutating operations: … backspace"
func TestInputLineValidatorBackspaceRejected(t *testing.T) {
	// RangeValidator rejects empty string as invalid partial input? No — it accepts "".
	// Use FilterValidator with a 2-char minimum by using a mock scenario:
	// Start with "12", use RangeValidator(10, 99) which accepts "1" (partial) but
	// IsValidInput("1") returns true. We need a case where backspace would produce
	// a string IsValidInput rejects.
	//
	// FilterValidator("") rejects everything non-empty. Use the opposite:
	// FilterValidator("") accepts only "" — so any text fails IsValidInput.
	// But that means we can never type in the first place...
	//
	// Better: FilterValidator with chars "abc" — start with "ab", backspace to "a".
	// "a" is accepted. Let's use a validator that rejects single chars but not two:
	// We need a custom setup. Use RangeValidator(10, 99): "1" IS valid partial
	// (digit only). "10" IS valid. Backspace from "10" → "1" → still valid.
	//
	// The simplest scenario: FilterValidator("") only accepts the empty string
	// via IsValidInput. Start text "a". Backspace → "" → IsValidInput("") → true
	// (accepted). Not a rejection case.
	//
	// Reliable rejection: Use a FilterValidator that rejects the post-backspace
	// result. Start with "ba", backspace would give "b", but FilterValidator("a")
	// rejects "b" because 'b' is not in valid chars.
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("ba") // cursor at 2; backspace → "b"
	il.SetValidator(NewFilterValidator("a")) // only 'a' is valid
	il.HandleEvent(keyEv(tcell.KeyBackspace2))
	if got := il.Text(); got != "ba" {
		t.Errorf("Text() after rejected backspace = %q, want %q (rolled back)", got, "ba")
	}
}

// TestInputLineValidatorBackspaceRejectedCursorRestored verifies cursor is restored
// after a rejected backspace.
// Spec: "the entire state (text, cursorPos, …) is rolled back."
func TestInputLineValidatorBackspaceRejectedCursorRestored(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("ba") // cursor at 2
	il.SetValidator(NewFilterValidator("a"))
	il.HandleEvent(keyEv(tcell.KeyBackspace2))
	if pos := il.CursorPos(); pos != 2 {
		t.Errorf("CursorPos() after rejected backspace = %d, want 2 (rolled back)", pos)
	}
}

// ---------------------------------------------------------------------------
// 5. Delete rejected by validator
// ---------------------------------------------------------------------------

// TestInputLineValidatorDeleteRejected verifies that delete is rolled back when
// the resulting text fails IsValidInput.
// Spec: "Keystroke rejection applies to ALL text-mutating operations: … delete"
func TestInputLineValidatorDeleteRejected(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("ab")
	il.HandleEvent(keyEv(tcell.KeyHome)) // cursor to 0
	// Delete would remove 'a', leaving "b". FilterValidator("a") rejects "b".
	il.SetValidator(NewFilterValidator("a"))
	il.HandleEvent(keyEv(tcell.KeyDelete))
	if got := il.Text(); got != "ab" {
		t.Errorf("Text() after rejected delete = %q, want %q (rolled back)", got, "ab")
	}
}

// TestInputLineValidatorDeleteRejectedCursorRestored verifies cursor is restored
// after a rejected delete.
// Spec: "the entire state (text, cursorPos, …) is rolled back."
func TestInputLineValidatorDeleteRejectedCursorRestored(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("ab")
	il.HandleEvent(keyEv(tcell.KeyHome)) // cursor to 0
	il.SetValidator(NewFilterValidator("a"))
	il.HandleEvent(keyEv(tcell.KeyDelete))
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("CursorPos() after rejected delete = %d, want 0 (rolled back)", pos)
	}
}

// ---------------------------------------------------------------------------
// 6. Ctrl+V paste rejected by validator
// ---------------------------------------------------------------------------

// TestInputLineValidatorCtrlVRejected verifies that paste is rolled back when the
// resulting text fails IsValidInput.
// Spec: "Keystroke rejection applies to ALL text-mutating operations: … Ctrl+V (paste)"
func TestInputLineValidatorCtrlVRejected(t *testing.T) {
	// Put "xyz" (invalid chars) in clipboard.
	seed := NewInputLine(NewRect(0, 0, 20, 1), 0)
	seed.SetText("xyz")
	seed.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	seed.HandleEvent(ctrlEv(tcell.KeyCtrlC))

	// Widget only accepts digits (via RangeValidator).
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("5")
	il.SetValidator(NewRangeValidator(0, 999))
	il.HandleEvent(ctrlEv(tcell.KeyCtrlV)) // pastes "xyz" → "5xyz" — rejected
	if got := il.Text(); got != "5" {
		t.Errorf("Text() after rejected Ctrl+V paste = %q, want %q (rolled back)", got, "5")
	}
}

// TestInputLineValidatorCtrlVRejectedCursorRestored verifies cursor position is
// restored after a rejected paste.
// Spec: "the entire state (text, cursorPos, …) is rolled back."
func TestInputLineValidatorCtrlVRejectedCursorRestored(t *testing.T) {
	seed := NewInputLine(NewRect(0, 0, 20, 1), 0)
	seed.SetText("xyz")
	seed.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	seed.HandleEvent(ctrlEv(tcell.KeyCtrlC))

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("5") // cursor at 1
	il.SetValidator(NewRangeValidator(0, 999))
	il.HandleEvent(ctrlEv(tcell.KeyCtrlV))
	if pos := il.CursorPos(); pos != 1 {
		t.Errorf("CursorPos() after rejected Ctrl+V = %d, want 1 (rolled back)", pos)
	}
}

// ---------------------------------------------------------------------------
// 7. Ctrl+X cut rejected by validator
// ---------------------------------------------------------------------------

// TestInputLineValidatorCtrlXRejected verifies that cut is rolled back when the
// resulting text fails IsValidInput.
// Spec: "Keystroke rejection applies to ALL text-mutating operations: … Ctrl+X (cut)"
func TestInputLineValidatorCtrlXRejected(t *testing.T) {
	// RangeValidator: "5" is valid partial; "" (after cut) is also valid partial.
	// Use FilterValidator("") so that the empty result of cut is rejected:
	// FilterValidator with chars "5" — cutting everything leaves "", which
	// IsValidInput rejects because 'nothing' has no invalid chars... actually
	// FilterValidator("5").IsValidInput("") = true (empty iterates nothing).
	//
	// Use a validator whose IsValidInput rejects empty string.
	// RangeValidator rejects "" via IsValidInput? Let's check: RangeValidator.IsValidInput("") returns true.
	//
	// Use PictureValidator("###") — IsValidInput("") returns false because scan
	// returns prEmpty (empty picture with no input)... actually for "###", prEmpty
	// is not prError so it passes IsValidInput (result != prError).
	// For PictureValidator("###"): scan(pic,0,[]rune{},0,false) → pic[0]='#', inIdx>=len(input)
	// → returns prIncomplete. prIncomplete != prError → IsValidInput returns true.
	//
	// We need IsValidInput to return false for empty string. The only built-in
	// that rejects empty partial input is FilterValidator with validChars that
	// doesn't cover ' '... but empty string has no characters so it always passes.
	//
	// Actually: The spec says IsValidInput checks the POST-mutation string. For
	// Ctrl+X we need to find a validator that rejects the result of cutting.
	// If we have "ab" selected and cut it, result is "". Most validators accept
	// "". So let's test Ctrl+X partial cut (not all selected).
	//
	// Better scenario: text "ab", select only "b" (positions 1..2).
	// After cut: "a". FilterValidator("b") → "a" is rejected.
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("ab") // cursor at 2
	// Select just "b": Shift+Left from pos 2 selects positions 1..2.
	il.HandleEvent(shiftKeyEv(tcell.KeyLeft))
	// Now "b" is selected. After cut → "a". FilterValidator("b") rejects "a".
	il.SetValidator(NewFilterValidator("b"))
	il.HandleEvent(ctrlEv(tcell.KeyCtrlX))
	if got := il.Text(); got != "ab" {
		t.Errorf("Text() after rejected Ctrl+X cut = %q, want %q (rolled back)", got, "ab")
	}
}

// TestInputLineValidatorCtrlXRejectedSelectionRestored verifies selection state is
// restored after a rejected cut.
// Spec: "the entire state (text, cursorPos, selStart, selEnd, scrollOffset) is rolled back."
func TestInputLineValidatorCtrlXRejectedSelectionRestored(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("ab")
	il.HandleEvent(shiftKeyEv(tcell.KeyLeft)) // select "b" at positions 1..2
	selStart, selEnd := il.Selection()
	il.SetValidator(NewFilterValidator("b"))
	il.HandleEvent(ctrlEv(tcell.KeyCtrlX))
	gotStart, gotEnd := il.Selection()
	if gotStart != selStart || gotEnd != selEnd {
		t.Errorf("Selection() after rejected Ctrl+X = (%d,%d), want (%d,%d) (rolled back)",
			gotStart, gotEnd, selStart, selEnd)
	}
}

// ---------------------------------------------------------------------------
// 8. Ctrl+Y clear rejected by validator
// ---------------------------------------------------------------------------

// TestInputLineValidatorCtrlYRejected verifies that Ctrl+Y (clear all) is rolled
// back when the resulting empty text fails IsValidInput.
// Spec: "Keystroke rejection applies to ALL text-mutating operations: … Ctrl+Y (clear all)"
//
// Note: Most validators accept "" as valid partial input. To test rejection,
// we use a FilterValidator whose empty-string check would pass. Instead we
// construct a validator where clearing text results in an invalid state.
// We use a custom approach: put "42" in the field with RangeValidator but ensure
// the validator rejects empty. The only way to do that cleanly is to verify
// the rollback path is exercised by using a FilterValidator that is set to
// reject the empty string.
//
// Since all built-in validators accept "" as partial input, we test the Ctrl+Y
// path by using a PictureValidator that rejects empty via IsValidInput being false.
// For PictureValidator("###") with empty input: scan returns prIncomplete (not prError),
// so IsValidInput("") = true. All validators accept "".
//
// Given this, we instead verify the accepted case — the state is wiped when
// Ctrl+Y passes validation. For the rejection test we use a validator
// that rejects the non-empty text BEFORE the clear (not the empty text after).
// That is not how Ctrl+Y rejection works (it checks the result "").
//
// Honest approach: use a validator that has IsValidInput("") = false.
// We can create a StringLookupValidator whose IsValidInput always returns true, so
// that can't help. We need to synthesize a validator.
//
// We use an inline test-only validator that rejects empty strings:
func TestInputLineValidatorCtrlYRejected(t *testing.T) {
	// rejectEmptyValidator is an inline Validator that rejects empty partial input.
	type rejectEmptyValidator struct{}
	v := &struct{ rejectEmptyValidator }{}
	_ = v // unused; define inline Validator below

	// We implement Validator inline via a struct with embedded method set.
	// Go doesn't support anonymous struct method sets, so we use a named
	// local type alias approach via a closure-based struct or just implement
	// the interface with a named test helper type.
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	il.SetValidator(&ctrlYRejectValidator{})
	il.HandleEvent(ctrlEv(tcell.KeyCtrlY))
	if got := il.Text(); got != "hello" {
		t.Errorf("Text() after rejected Ctrl+Y = %q, want %q (rolled back)", got, "hello")
	}
}

// ctrlYRejectValidator is a test-only Validator that rejects empty strings in
// IsValidInput, simulating a validator that requires at least one character.
type ctrlYRejectValidator struct{}

func (v *ctrlYRejectValidator) IsValidInput(s string, noAutoFill bool) bool {
	return s != "" // rejects empty — the result of Ctrl+Y clearing all text
}

func (v *ctrlYRejectValidator) IsValid(s string) bool {
	return s != ""
}

func (v *ctrlYRejectValidator) Error() {
	// no-op in tests
}

// TestInputLineValidatorCtrlYRejectedCursorRestored verifies cursor is restored after
// a rejected Ctrl+Y.
// Spec: "the entire state (text, cursorPos, …) is rolled back."
func TestInputLineValidatorCtrlYRejectedCursorRestored(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello") // cursor at 5
	il.SetValidator(&ctrlYRejectValidator{})
	il.HandleEvent(ctrlEv(tcell.KeyCtrlY))
	if pos := il.CursorPos(); pos != 5 {
		t.Errorf("CursorPos() after rejected Ctrl+Y = %d, want 5 (rolled back)", pos)
	}
}

// TestInputLineValidatorCtrlYAccepted verifies that Ctrl+Y clears all text when
// the validator accepts the empty result.
// Spec: "Ctrl+Y: clear all" — accepted case; regression guard.
func TestInputLineValidatorCtrlYAccepted(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	il.SetValidator(NewFilterValidator("helo ")) // empty "" passes IsValidInput
	il.HandleEvent(ctrlEv(tcell.KeyCtrlY))
	if got := il.Text(); got != "" {
		t.Errorf("Text() after accepted Ctrl+Y = %q, want empty string", got)
	}
}

// ---------------------------------------------------------------------------
// 9. No validator — regression check (behaves as before)
// ---------------------------------------------------------------------------

// TestInputLineNoValidatorInsertWorks verifies that without a validator, character
// insert works exactly as before.
// Spec: "When no validator is set, InputLine behaves exactly as before (no regression)."
func TestInputLineNoValidatorInsertWorks(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	// No validator set.
	il.HandleEvent(runeEv('x'))
	il.HandleEvent(runeEv('y'))
	if got := il.Text(); got != "xy" {
		t.Errorf("Text() without validator after typing = %q, want %q", got, "xy")
	}
}

// TestInputLineNoValidatorBackspaceWorks verifies backspace without validator behaves normally.
// Spec: "When no validator is set, InputLine behaves exactly as before (no regression)."
func TestInputLineNoValidatorBackspaceWorks(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyBackspace2))
	if got := il.Text(); got != "ab" {
		t.Errorf("Text() without validator after backspace = %q, want %q", got, "ab")
	}
}

// TestInputLineNoValidatorDeleteWorks verifies delete without validator behaves normally.
// Spec: "When no validator is set, InputLine behaves exactly as before (no regression)."
func TestInputLineNoValidatorDeleteWorks(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyHome))
	il.HandleEvent(keyEv(tcell.KeyDelete))
	if got := il.Text(); got != "bc" {
		t.Errorf("Text() without validator after delete = %q, want %q", got, "bc")
	}
}

// TestInputLineNoValidatorCtrlYWorks verifies Ctrl+Y without validator clears all.
// Spec: "When no validator is set, InputLine behaves exactly as before (no regression)."
func TestInputLineNoValidatorCtrlYWorks(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlY))
	if got := il.Text(); got != "" {
		t.Errorf("Text() without validator after Ctrl+Y = %q, want empty string", got)
	}
}

// ---------------------------------------------------------------------------
// 10. Blur validation — text valid on blur, no re-focus
// ---------------------------------------------------------------------------

// TestInputLineBlurValidationValid verifies that when SfSelected transitions from
// true to false and the text is valid, the validator's Error() is not called and
// the widget is not re-focused.
// Spec: "When SfSelected transitions from true to false AND a validator is set,
// it calls validator.IsValid(text). If false, calls validator.Error() and re-focuses."
func TestInputLineBlurValidationValid(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 5))

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("50") // valid for RangeValidator(0, 100)
	il.SetValidator(NewRangeValidator(0, 100))

	// Insert a second selectable widget so the group can have an alternative focus.
	il2 := NewInputLine(NewRect(0, 1, 20, 1), 0)

	g.Insert(il)
	g.Insert(il2)

	// Focus il.
	g.SetFocusedChild(il)

	// Move focus to il2 — this triggers il's blur (SfSelected false).
	g.SetFocusedChild(il2)

	// After blur with valid text, il must NOT be re-focused.
	if g.FocusedChild() == il {
		t.Error("after blur with valid text, InputLine was re-focused; expected il2 to remain focused")
	}
	if !il2.HasState(SfSelected) {
		t.Error("il2 does not have SfSelected after SetFocusedChild(il2)")
	}
}

// ---------------------------------------------------------------------------
// 11. Blur validation — text invalid on blur, validator.IsValid called
// ---------------------------------------------------------------------------

// TestInputLineBlurValidationInvalidCallsError verifies that when blur occurs with
// invalid text, the validator's IsValid returns false and the widget re-focuses.
// Spec: "If false, calls validator.Error() and re-focuses by calling
// owner.SetFocusedChild(self)."
func TestInputLineBlurValidationInvalidRefocuses(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 5))

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	// RangeValidator(0, 100): "+" is valid partial input but IsValid("+") = false
	// (not a complete integer). So on blur, IsValid returns false.
	il.SetText("+")
	il.SetValidator(NewRangeValidator(0, 100))

	il2 := NewInputLine(NewRect(0, 1, 20, 1), 0)

	g.Insert(il)
	g.Insert(il2)

	g.SetFocusedChild(il)

	// Move focus to il2. il has invalid text ("+") so blur validation should re-focus il.
	g.SetFocusedChild(il2)

	// il should be re-focused because its text is invalid.
	if g.FocusedChild() != il {
		t.Errorf("after blur with invalid text, FocusedChild() = %v, want il (re-focused)", g.FocusedChild())
	}
}

// TestInputLineBlurValidationIsValidCalled verifies that IsValid is called with the
// current text on blur (not an empty string or some other value).
// Spec: "calls validator.IsValid(text)"
func TestInputLineBlurValidationIsValidCalledWithText(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 5))

	// trackingValidator records the text passed to IsValid.
	called := false
	calledWith := ""
	tv := &blurTrackingValidator{
		onIsValid: func(s string) bool {
			called = true
			calledWith = s
			return true // valid — we just want to track the call
		},
	}

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")
	il.SetValidator(tv)

	il2 := NewInputLine(NewRect(0, 1, 20, 1), 0)

	g.Insert(il)
	g.Insert(il2)
	g.SetFocusedChild(il)
	g.SetFocusedChild(il2) // trigger blur

	if !called {
		t.Error("validator.IsValid was not called on blur")
	}
	if calledWith != "hello" {
		t.Errorf("validator.IsValid called with %q, want %q", calledWith, "hello")
	}
}

// TestInputLineBlurValidationNoValidatorNoReocus verifies that without a validator,
// blur does not cause any refocus even if text is "invalid" by some external criteria.
// Spec: "When SfSelected transitions from true to false AND a validator is set, it calls…"
func TestInputLineBlurValidationNoValidatorNoRefocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 5))

	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("whatever") // no validator set
	il2 := NewInputLine(NewRect(0, 1, 20, 1), 0)

	g.Insert(il)
	g.Insert(il2)
	g.SetFocusedChild(il)
	g.SetFocusedChild(il2)

	// Without a validator, il should not be re-focused.
	if g.FocusedChild() == il {
		t.Error("without validator, InputLine was unexpectedly re-focused on blur")
	}
}

// blurTrackingValidator is a test-only Validator that delegates IsValid to a
// closure, allowing tests to verify when and with what value IsValid is called.
type blurTrackingValidator struct {
	onIsValid func(s string) bool
}

func (v *blurTrackingValidator) IsValidInput(s string, noAutoFill bool) bool {
	return true
}

func (v *blurTrackingValidator) IsValid(s string) bool {
	if v.onIsValid != nil {
		return v.onIsValid(s)
	}
	return true
}

func (v *blurTrackingValidator) Error() {
	// no-op in tests
}

// ---------------------------------------------------------------------------
// 12. Auto-fill integration (ValidatorWithAutoFill)
// ---------------------------------------------------------------------------

// TestInputLineValidatorAutoFillModifiesText verifies that when the validator
// implements ValidatorWithAutoFill, accepted keystrokes can have their text
// modified by IsValidInputAutoFill.
// Spec: "If the validator implements ValidatorWithAutoFill, InputLine calls
// IsValidInputAutoFill(&text, false) and accepts any modifications to the text string."
func TestInputLineValidatorAutoFillModifiesText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	// PictureValidator("&&&") auto-uppercases typed letters.
	il.SetValidator(NewPictureValidator("&&&", false))
	il.HandleEvent(runeEv('a'))
	il.HandleEvent(runeEv('b'))
	il.HandleEvent(runeEv('c'))
	// Each letter is a valid insert, and auto-fill should uppercase it.
	if got := il.Text(); got != "ABC" {
		t.Errorf("Text() after typing with auto-fill validator = %q, want %q", got, "ABC")
	}
}

// TestInputLineValidatorAutoFillDoesNotRejectValid verifies that auto-fill accepts
// valid input (the mutation stands).
// Spec: "InputLine calls IsValidInputAutoFill(&text, false) and accepts any modifications."
func TestInputLineValidatorAutoFillDoesNotRejectValid(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewPictureValidator("###", false)) // 3-digit picture
	il.HandleEvent(runeEv('1'))
	il.HandleEvent(runeEv('2'))
	// "12" is valid partial — insert should be accepted.
	if got := il.Text(); got != "12" {
		t.Errorf("Text() after valid inserts with PictureValidator = %q, want %q", got, "12")
	}
}

// TestInputLineValidatorAutoFillRejectsInvalid verifies that auto-fill also rejects
// invalid inserts (returning false from IsValidInputAutoFill means rollback).
// Spec: "accepts any modifications to the text string" — but if AutoFill returns false,
// the state is rolled back just as with a regular validator rejection.
func TestInputLineValidatorAutoFillRejectsInvalid(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewPictureValidator("###", false)) // digit-only picture
	il.HandleEvent(runeEv('a'))                        // letter rejected by "###"
	if got := il.Text(); got != "" {
		t.Errorf("Text() after rejected insert with PictureValidator = %q, want empty (rolled back)", got)
	}
}
