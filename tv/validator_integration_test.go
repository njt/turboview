package tv

// validator_integration_test.go — Integration tests for the complete validator
// pipeline wired through real InputLine and Dialog components.
//
// These tests use no mocks. Every component is the real production type.
// Each test targets observable pipeline behavior: keystrokes reaching the
// InputLine, the validator deciding accept/reject, and Dialog.Valid() gating
// submission.

import (
	"testing"
)

// ---------------------------------------------------------------------------
// 1. RangeValidator: accepted keystrokes land in the field; rejected are dropped
// ---------------------------------------------------------------------------

// TestValidatorIntegrationRangeAcceptsValidTyping verifies that typing "50" into
// an InputLine with RangeValidator(1,100) produces text "50".
func TestValidatorIntegrationRangeAcceptsValidTyping(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewRangeValidator(1, 100))

	il.HandleEvent(runeEv('5'))
	il.HandleEvent(runeEv('0'))

	if got := il.Text(); got != "50" {
		t.Errorf("Text() after typing \"50\" with RangeValidator(1,100) = %q, want %q", got, "50")
	}
}

// TestValidatorIntegrationRangeRejectsNonDigitInput verifies that typing "abc"
// after "50" leaves the text unchanged at "50" — the letters are rejected
// character by character by RangeValidator's character whitelist.
func TestValidatorIntegrationRangeRejectsNonDigitInput(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewRangeValidator(1, 100))

	il.HandleEvent(runeEv('5'))
	il.HandleEvent(runeEv('0'))
	// Each of these should be rejected (letters not in "0-9+" whitelist).
	il.HandleEvent(runeEv('a'))
	il.HandleEvent(runeEv('b'))
	il.HandleEvent(runeEv('c'))

	if got := il.Text(); got != "50" {
		t.Errorf("Text() after typing \"50abc\" with RangeValidator(1,100) = %q, want %q (letters rejected)", got, "50")
	}
}

// ---------------------------------------------------------------------------
// 2. RangeValidator: IsValid on programmatically-set out-of-range text
// ---------------------------------------------------------------------------

// TestValidatorIntegrationRangeIsValidOutOfRange verifies that RangeValidator.IsValid
// returns false for "200" when the range is [1, 100].
func TestValidatorIntegrationRangeIsValidOutOfRange(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	v := NewRangeValidator(1, 100)
	il.SetValidator(v)
	il.SetText("200")

	if v.IsValid(il.Text()) {
		t.Errorf("RangeValidator(1,100).IsValid(%q) = true, want false (out of range)", il.Text())
	}
}

// TestValidatorIntegrationRangeIsValidInRange verifies that RangeValidator.IsValid
// returns true for "50" when the range is [1, 100].
func TestValidatorIntegrationRangeIsValidInRange(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	v := NewRangeValidator(1, 100)
	il.SetValidator(v)
	il.SetText("50")

	if !v.IsValid(il.Text()) {
		t.Errorf("RangeValidator(1,100).IsValid(%q) = false, want true (in range)", il.Text())
	}
}

// ---------------------------------------------------------------------------
// 3. Dialog.Valid() with RangeValidator gating CmOK
// ---------------------------------------------------------------------------

// TestValidatorIntegrationDialogValidFalseForOutOfRange verifies that Dialog.Valid(CmOK)
// returns false when the InputLine text "200" fails RangeValidator(1,100).
func TestValidatorIntegrationDialogValidFalseForOutOfRange(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Test")

	il := NewInputLine(NewRect(0, 0, 30, 1), 0)
	il.SetValidator(NewRangeValidator(1, 100))
	il.SetText("200")
	dlg.Insert(il)

	if dlg.Valid(CmOK) {
		t.Error("Dialog.Valid(CmOK) returned true when InputLine text \"200\" is out of range [1,100], want false")
	}
}

// TestValidatorIntegrationDialogValidTrueForInRange verifies that Dialog.Valid(CmOK)
// returns true when the InputLine text "50" passes RangeValidator(1,100).
func TestValidatorIntegrationDialogValidTrueForInRange(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Test")

	il := NewInputLine(NewRect(0, 0, 30, 1), 0)
	il.SetValidator(NewRangeValidator(1, 100))
	il.SetText("50")
	dlg.Insert(il)

	if !dlg.Valid(CmOK) {
		t.Error("Dialog.Valid(CmOK) returned false when InputLine text \"50\" is in range [1,100], want true")
	}
}

// TestValidatorIntegrationDialogCmOKBlockedWhenInvalid verifies that HandleEvent
// clears the CmOK event when the validator rejects the current text.
func TestValidatorIntegrationDialogCmOKBlockedWhenInvalid(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Test")

	il := NewInputLine(NewRect(0, 0, 30, 1), 0)
	il.SetValidator(NewRangeValidator(1, 100))
	il.SetText("200")
	dlg.Insert(il)

	ev := &Event{What: EvCommand, Command: CmOK}
	dlg.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("HandleEvent did not clear CmOK event when InputLine has out-of-range text \"200\"")
	}
}

// TestValidatorIntegrationDialogCmOKAllowedWhenValid verifies that HandleEvent does
// NOT clear the CmOK event when the validator accepts the current text.
func TestValidatorIntegrationDialogCmOKAllowedWhenValid(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Test")

	il := NewInputLine(NewRect(0, 0, 30, 1), 0)
	il.SetValidator(NewRangeValidator(1, 100))
	il.SetText("50")
	dlg.Insert(il)

	ev := &Event{What: EvCommand, Command: CmOK}
	dlg.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("HandleEvent cleared CmOK event when InputLine text \"50\" is valid, want event to pass through")
	}
}

// ---------------------------------------------------------------------------
// 4. FilterValidator: keystroke filtering wired through InputLine
// ---------------------------------------------------------------------------

// TestValidatorIntegrationFilterAcceptsValidChars verifies that typing characters
// from the allowed set is accepted and appended to the text.
func TestValidatorIntegrationFilterAcceptsValidChars(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewFilterValidator("abcdef"))

	il.HandleEvent(runeEv('a'))
	il.HandleEvent(runeEv('b'))
	il.HandleEvent(runeEv('c'))

	if got := il.Text(); got != "abc" {
		t.Errorf("Text() after typing \"abc\" with FilterValidator(\"abcdef\") = %q, want %q", got, "abc")
	}
}

// TestValidatorIntegrationFilterRejectsInvalidChar verifies that a character outside
// the allowed set is rejected, leaving the text unchanged.
func TestValidatorIntegrationFilterRejectsInvalidChar(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewFilterValidator("abcdef"))

	il.HandleEvent(runeEv('a'))
	il.HandleEvent(runeEv('b'))
	il.HandleEvent(runeEv('c'))
	// 'z' is not in "abcdef" — should be rejected.
	il.HandleEvent(runeEv('z'))

	if got := il.Text(); got != "abc" {
		t.Errorf("Text() after typing \"abcz\" with FilterValidator(\"abcdef\") = %q, want %q (z rejected)", got, "abc")
	}
}

// TestValidatorIntegrationFilterRejectDoesNotMoveCursor verifies that a rejected
// keystroke leaves the cursor position unchanged.
func TestValidatorIntegrationFilterRejectDoesNotMoveCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewFilterValidator("abcdef"))

	il.HandleEvent(runeEv('a'))
	il.HandleEvent(runeEv('b'))
	il.HandleEvent(runeEv('c'))
	pos := il.CursorPos() // should be 3

	il.HandleEvent(runeEv('z')) // rejected

	if after := il.CursorPos(); after != pos {
		t.Errorf("CursorPos() after rejected 'z' = %d, want %d (unchanged)", after, pos)
	}
}

// ---------------------------------------------------------------------------
// 5. StringLookupValidator: IsValidInput always true; IsValid checks membership
// ---------------------------------------------------------------------------

// TestValidatorIntegrationStringLookupTypingAccepted verifies that any character
// can be typed because StringLookupValidator.IsValidInput always returns true.
func TestValidatorIntegrationStringLookupTypingAccepted(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewStringLookupValidator([]string{"apple", "banana"}))

	// Type a string that is not in the list — partial input is still accepted.
	il.HandleEvent(runeEv('a'))
	il.HandleEvent(runeEv('p'))
	il.HandleEvent(runeEv('p'))
	il.HandleEvent(runeEv('l'))
	il.HandleEvent(runeEv('e'))

	if got := il.Text(); got != "apple" {
		t.Errorf("Text() after typing \"apple\" with StringLookupValidator = %q, want %q", got, "apple")
	}
}

// TestValidatorIntegrationStringLookupIsValidTrue verifies that IsValid returns
// true for a string that is in the lookup list.
func TestValidatorIntegrationStringLookupIsValidTrue(t *testing.T) {
	v := NewStringLookupValidator([]string{"apple", "banana"})

	if !v.IsValid("apple") {
		t.Error("StringLookupValidator.IsValid(\"apple\") = false, want true (\"apple\" is in list)")
	}
}

// TestValidatorIntegrationStringLookupIsValidFalse verifies that IsValid returns
// false for a string that is NOT in the lookup list.
func TestValidatorIntegrationStringLookupIsValidFalse(t *testing.T) {
	v := NewStringLookupValidator([]string{"apple", "banana"})

	if v.IsValid("pear") {
		t.Error("StringLookupValidator.IsValid(\"pear\") = true, want false (\"pear\" is not in list)")
	}
}

// TestValidatorIntegrationStringLookupIsValidInputAlwaysTrue verifies that
// IsValidInput returns true for arbitrary strings (no character-level filtering).
func TestValidatorIntegrationStringLookupIsValidInputAlwaysTrue(t *testing.T) {
	v := NewStringLookupValidator([]string{"apple", "banana"})

	for _, s := range []string{"", "z", "pear", "not_in_list_at_all_12345"} {
		if !v.IsValidInput(s, false) {
			t.Errorf("StringLookupValidator.IsValidInput(%q, false) = false, want true (always accepts partial input)", s)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. PictureValidator "###": accepts 1–3 digits, rejects letters and 4+ digits
// ---------------------------------------------------------------------------

// TestValidatorIntegrationPictureAcceptsSingleDigit verifies that "1" is accepted
// as valid partial input for picture "###".
func TestValidatorIntegrationPictureAcceptsSingleDigit(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewPictureValidator("###", false))

	il.HandleEvent(runeEv('1'))

	if got := il.Text(); got != "1" {
		t.Errorf("Text() after typing '1' with PictureValidator(\"###\") = %q, want %q", got, "1")
	}
}

// TestValidatorIntegrationPictureAcceptsTwoDigits verifies that "12" is accepted.
func TestValidatorIntegrationPictureAcceptsTwoDigits(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewPictureValidator("###", false))

	il.HandleEvent(runeEv('1'))
	il.HandleEvent(runeEv('2'))

	if got := il.Text(); got != "12" {
		t.Errorf("Text() after typing '12' with PictureValidator(\"###\") = %q, want %q", got, "12")
	}
}

// TestValidatorIntegrationPictureAcceptsThreeDigits verifies that "123" is accepted
// as a complete match for picture "###".
func TestValidatorIntegrationPictureAcceptsThreeDigits(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewPictureValidator("###", false))

	il.HandleEvent(runeEv('1'))
	il.HandleEvent(runeEv('2'))
	il.HandleEvent(runeEv('3'))

	if got := il.Text(); got != "123" {
		t.Errorf("Text() after typing '123' with PictureValidator(\"###\") = %q, want %q", got, "123")
	}
}

// TestValidatorIntegrationPictureRejectsLetter verifies that a letter is rejected
// by PictureValidator("###") — '#' only accepts digits.
func TestValidatorIntegrationPictureRejectsLetter(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewPictureValidator("###", false))

	// 'a' is not a digit — should be rejected for picture '#'.
	il.HandleEvent(runeEv('a'))

	if got := il.Text(); got != "" {
		t.Errorf("Text() after typing 'a' with PictureValidator(\"###\") = %q, want empty (letter rejected)", got)
	}
}

// TestValidatorIntegrationPictureRejectsFourthDigit verifies that a fourth digit
// is rejected once three digits have already been typed (picture "###" is full).
func TestValidatorIntegrationPictureRejectsFourthDigit(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetValidator(NewPictureValidator("###", false))

	il.HandleEvent(runeEv('1'))
	il.HandleEvent(runeEv('2'))
	il.HandleEvent(runeEv('3'))
	// Fourth digit: "1234" would exceed the 3-char picture — should be rejected.
	il.HandleEvent(runeEv('4'))

	if got := il.Text(); got != "123" {
		t.Errorf("Text() after typing '1234' with PictureValidator(\"###\") = %q, want %q (4th digit rejected)", got, "123")
	}
}
