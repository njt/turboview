package tv

import (
	"fmt"
	"testing"
)

// TestFilterValidatorIsValidInputValidChars verifies:
// "FilterValidator.IsValidInput("abc", false) returns true when validChars contains 'a', 'b', 'c'"
func TestFilterValidatorIsValidInputValidChars(t *testing.T) {
	v := NewFilterValidator("abc")
	if !v.IsValidInput("abc", false) {
		t.Error("IsValidInput(\"abc\", false) should return true when all characters are in validChars")
	}
}

// TestFilterValidatorIsValidInputValidCharsSubset verifies:
// "FilterValidator.IsValidInput("abc", false) returns true when validChars contains 'a', 'b', 'c'" (subset case)
func TestFilterValidatorIsValidInputValidCharsSubset(t *testing.T) {
	v := NewFilterValidator("abcdef")
	if !v.IsValidInput("abc", false) {
		t.Error("IsValidInput(\"abc\", false) should return true when validChars is superset of input")
	}
}

// TestFilterValidatorIsValidInputMissingChar verifies:
// "FilterValidator.IsValidInput("abc", false) returns false when validChars is "ab" (missing 'c')"
func TestFilterValidatorIsValidInputMissingChar(t *testing.T) {
	v := NewFilterValidator("ab")
	if v.IsValidInput("abc", false) {
		t.Error("IsValidInput(\"abc\", false) should return false when validChars is \"ab\" and input contains 'c'")
	}
}

// TestFilterValidatorIsValidInputEmptyStringValid verifies:
// "FilterValidator.IsValidInput("", false) returns true (empty string is valid partial input)"
func TestFilterValidatorIsValidInputEmptyStringValid(t *testing.T) {
	v := NewFilterValidator("abc")
	if !v.IsValidInput("", false) {
		t.Error("IsValidInput(\"\", false) should return true")
	}
}

// TestFilterValidatorIsValidInputEmptyStringValidEmptyChars verifies:
// empty input is valid even with empty validChars
func TestFilterValidatorIsValidInputEmptyStringValidEmptyChars(t *testing.T) {
	v := NewFilterValidator("")
	if !v.IsValidInput("", false) {
		t.Error("IsValidInput(\"\", false) should return true even when validChars is empty")
	}
}

// TestFilterValidatorIsValidAllCharsValid verifies:
// "FilterValidator.IsValid("abc") returns true when all chars are in validChars"
func TestFilterValidatorIsValidAllCharsValid(t *testing.T) {
	v := NewFilterValidator("abc")
	if !v.IsValid("abc") {
		t.Error("IsValid(\"abc\") should return true when all characters are in validChars")
	}
}

// TestFilterValidatorIsValidAllCharsValidSubset verifies:
// "FilterValidator.IsValid("abc") returns true when all chars are in validChars" (subset case)
func TestFilterValidatorIsValidAllCharsValidSubset(t *testing.T) {
	v := NewFilterValidator("abcdef")
	if !v.IsValid("abc") {
		t.Error("IsValid(\"abc\") should return true when validChars is superset of input")
	}
}

// TestFilterValidatorIsValidMissingChar verifies:
// IsValid returns false when input contains a character not in validChars
func TestFilterValidatorIsValidMissingChar(t *testing.T) {
	v := NewFilterValidator("ab")
	if v.IsValid("abc") {
		t.Error("IsValid(\"abc\") should return false when validChars is \"ab\" and input contains 'c'")
	}
}

// TestFilterValidatorIsValidEmptyString verifies:
// "FilterValidator.IsValid("") returns true (empty is valid — FilterValidator only checks character membership)"
func TestFilterValidatorIsValidEmptyString(t *testing.T) {
	v := NewFilterValidator("abc")
	if !v.IsValid("") {
		t.Error("IsValid(\"\") should return true")
	}
}

// TestFilterValidatorIsValidEmptyStringEmptyChars verifies:
// empty string is valid even when validChars is empty
func TestFilterValidatorIsValidEmptyStringEmptyChars(t *testing.T) {
	v := NewFilterValidator("")
	if !v.IsValid("") {
		t.Error("IsValid(\"\") should return true even when validChars is empty")
	}
}

// TestFilterValidatorErrorCallable verifies:
// "FilterValidator.Error() is callable (in production it shows a message box; for unit testing, it should not panic when no Application is running)"
func TestFilterValidatorErrorCallable(t *testing.T) {
	v := NewFilterValidator("abc")
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Error() should not panic, but recovered: %v", r)
		}
	}()
	v.Error()
}

// TestFilterValidatorRuneLevelAscii verifies:
// "Rune-level matching: FilterValidator with validChars "äöü" accepts "ä" input"
func TestFilterValidatorRuneLevelAscii(t *testing.T) {
	v := NewFilterValidator("äöü")
	if !v.IsValidInput("ä", false) {
		t.Error("IsValidInput(\"ä\", false) should return true when validChars is \"äöü\"")
	}
}

// TestFilterValidatorRuneLevelUnicode verifies rune-level matching with Unicode
func TestFilterValidatorRuneLevelUnicode(t *testing.T) {
	v := NewFilterValidator("äöü")
	if !v.IsValid("ä") {
		t.Error("IsValid(\"ä\") should return true when validChars is \"äöü\"")
	}
}

// TestFilterValidatorRuneLevelMissingUnicode verifies rune-level matching rejects missing runes
func TestFilterValidatorRuneLevelMissingUnicode(t *testing.T) {
	v := NewFilterValidator("äö")
	if v.IsValidInput("ü", false) {
		t.Error("IsValidInput(\"ü\", false) should return false when validChars is \"äö\"")
	}
}

// TestFilterValidatorSingleCharValid verifies:
// single character matching works
func TestFilterValidatorSingleCharValid(t *testing.T) {
	v := NewFilterValidator("a")
	if !v.IsValid("a") {
		t.Error("IsValid(\"a\") should return true when validChars is \"a\"")
	}
}

// TestFilterValidatorSingleCharInvalid verifies:
// single character rejection works
func TestFilterValidatorSingleCharInvalid(t *testing.T) {
	v := NewFilterValidator("a")
	if v.IsValid("b") {
		t.Error("IsValid(\"b\") should return false when validChars is \"a\"")
	}
}

// TestFilterValidatorMultipleValidChars verifies:
// FilterValidator works with larger character sets
func TestFilterValidatorMultipleValidChars(t *testing.T) {
	validChars := "abcdefghijklmnopqrstuvwxyz0123456789"
	v := NewFilterValidator(validChars)
	if !v.IsValid("abc123xyz") {
		t.Error("IsValid should return true for inputs within the valid character set")
	}
}

// TestFilterValidatorMultipleInvalidChar verifies:
// FilterValidator rejects inputs with invalid characters in large sets
func TestFilterValidatorMultipleInvalidChar(t *testing.T) {
	validChars := "abcdefghijklmnopqrstuvwxyz"
	v := NewFilterValidator(validChars)
	if v.IsValid("abc123") {
		t.Error("IsValid should return false when input contains '1' not in validChars")
	}
}

// TestFilterValidatorIsValidInputPartialValidTrue verifies:
// IsValidInput returns true for partial input that is valid so far
func TestFilterValidatorIsValidInputPartialValidTrue(t *testing.T) {
	v := NewFilterValidator("abc")
	if !v.IsValidInput("a", false) {
		t.Error("IsValidInput(\"a\", false) should return true when 'a' is in validChars")
	}
}

// TestFilterValidatorIsValidInputPartialInvalidFalse verifies:
// IsValidInput returns false for partial input with invalid characters
func TestFilterValidatorIsValidInputPartialInvalidFalse(t *testing.T) {
	v := NewFilterValidator("abc")
	if v.IsValidInput("d", false) {
		t.Error("IsValidInput(\"d\", false) should return false when 'd' is not in validChars")
	}
}

// TestFilterValidatorIsValidInputNoAutoFillFalse verifies:
// IsValidInput with noAutoFill=false parameter works
func TestFilterValidatorIsValidInputNoAutoFillFalse(t *testing.T) {
	v := NewFilterValidator("abc")
	if !v.IsValidInput("ab", false) {
		t.Error("IsValidInput(\"ab\", false) should return true when both 'a' and 'b' are in validChars")
	}
}

// TestFilterValidatorNewFilterValidatorReturnsValidator verifies:
// NewFilterValidator returns a value that implements Validator
func TestFilterValidatorNewFilterValidatorReturnsValidator(t *testing.T) {
	v := NewFilterValidator("abc")
	var _ Validator = v
}

// TestFilterValidatorNewFilterValidatorMultipleCalls verifies:
// multiple FilterValidator instances are independent
func TestFilterValidatorNewFilterValidatorMultipleCalls(t *testing.T) {
	v1 := NewFilterValidator("a")
	v2 := NewFilterValidator("b")

	if !v1.IsValid("a") {
		t.Error("v1.IsValid(\"a\") should return true")
	}
	if v1.IsValid("b") {
		t.Error("v1.IsValid(\"b\") should return false")
	}
	if !v2.IsValid("b") {
		t.Error("v2.IsValid(\"b\") should return true")
	}
	if v2.IsValid("a") {
		t.Error("v2.IsValid(\"a\") should return false")
	}
}

// TestFilterValidatorEmptyValidCharsRejectsNonEmpty verifies:
// FilterValidator with empty validChars rejects non-empty input
func TestFilterValidatorEmptyValidCharsRejectsNonEmpty(t *testing.T) {
	v := NewFilterValidator("")
	if v.IsValid("a") {
		t.Error("IsValid(\"a\") should return false when validChars is empty")
	}
}

// TestFilterValidatorEmptyValidCharsRejectsNonEmptyInput verifies:
// FilterValidator with empty validChars rejects non-empty partial input
func TestFilterValidatorEmptyValidCharsRejectsNonEmptyInput(t *testing.T) {
	v := NewFilterValidator("")
	if v.IsValidInput("a", false) {
		t.Error("IsValidInput(\"a\", false) should return false when validChars is empty")
	}
}

// TestFilterValidatorIsValidInputNoAutoFillTrue verifies:
// IsValidInput accepts noAutoFill=true without different behavior for FilterValidator
func TestFilterValidatorIsValidInputNoAutoFillTrue(t *testing.T) {
	v := NewFilterValidator("abc")
	if !v.IsValidInput("abc", true) {
		t.Error("IsValidInput(\"abc\", true) should return true — FilterValidator has no auto-fill behavior")
	}
}

// RangeValidator tests

// TestRangeValidatorNewRangeValidatorReturnsValidator verifies:
// "NewRangeValidator(min, max int) returns a *RangeValidator that implements Validator"
func TestRangeValidatorNewRangeValidatorReturnsValidator(t *testing.T) {
	v := NewRangeValidator(0, 100)
	var _ Validator = v
	if v == nil {
		t.Error("NewRangeValidator should return non-nil value")
	}
}

// TestRangeValidatorIsValidInputEmptyStringPartial verifies:
// "RangeValidator.IsValidInput("", false) returns true (empty is valid partial input during typing)"
func TestRangeValidatorIsValidInputEmptyStringPartial(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if !v.IsValidInput("", false) {
		t.Error("IsValidInput(\"\", false) should return true")
	}
}

// TestRangeValidatorIsValidInputEmptyStringPartialNegRange verifies:
// empty string is valid partial input even with negative range
func TestRangeValidatorIsValidInputEmptyStringPartialNegRange(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	if !v.IsValidInput("", false) {
		t.Error("IsValidInput(\"\", false) should return true even with negative range")
	}
}

// TestRangeValidatorIsValidInputPlusSignPartial verifies:
// "RangeValidator.IsValidInput("+", false) returns true (lone sign is valid partial input)"
func TestRangeValidatorIsValidInputPlusSignPartial(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if !v.IsValidInput("+", false) {
		t.Error("IsValidInput(\"+\", false) should return true")
	}
}

// TestRangeValidatorIsValidInputMinusSignPartialPositiveRange verifies:
// "RangeValidator.IsValidInput("-", false) returns true only when min < 0"
func TestRangeValidatorIsValidInputMinusSignPartialPositiveRange(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if v.IsValidInput("-", false) {
		t.Error("IsValidInput(\"-\", false) should return false when min >= 0")
	}
}

// TestRangeValidatorIsValidInputMinusSignPartialNegativeRange verifies:
// minus sign is accepted in partial input when min < 0
func TestRangeValidatorIsValidInputMinusSignPartialNegativeRange(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	if !v.IsValidInput("-", false) {
		t.Error("IsValidInput(\"-\", false) should return true when min < 0")
	}
}

// TestRangeValidatorIsValidInputDigitPartialPositive verifies:
// digits are accepted as partial input with positive range
func TestRangeValidatorIsValidInputDigitPartialPositive(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if !v.IsValidInput("5", false) {
		t.Error("IsValidInput(\"5\", false) should return true")
	}
}

// TestRangeValidatorIsValidInputDigitPartialNegative verifies:
// digits are accepted as partial input with negative range
func TestRangeValidatorIsValidInputDigitPartialNegative(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	if !v.IsValidInput("5", false) {
		t.Error("IsValidInput(\"5\", false) should return true")
	}
}

// TestRangeValidatorIsValidInputZeroPartial verifies:
// zero digit is accepted as partial input
func TestRangeValidatorIsValidInputZeroPartial(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if !v.IsValidInput("0", false) {
		t.Error("IsValidInput(\"0\", false) should return true")
	}
}

// TestRangeValidatorIsValidInputMultipleDigitsPartial verifies:
// multiple digits are accepted as partial input
func TestRangeValidatorIsValidInputMultipleDigitsPartial(t *testing.T) {
	v := NewRangeValidator(0, 1000)
	if !v.IsValidInput("123", false) {
		t.Error("IsValidInput(\"123\", false) should return true")
	}
}

// TestRangeValidatorIsValidInputPlusSignWithDigitsPartial verifies:
// plus sign followed by digits is accepted as partial input
func TestRangeValidatorIsValidInputPlusSignWithDigitsPartial(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if !v.IsValidInput("+42", false) {
		t.Error("IsValidInput(\"+42\", false) should return true")
	}
}

// TestRangeValidatorIsValidInputMinusSignWithDigitsPartialNegRange verifies:
// minus sign followed by digits is accepted as partial input when min < 0
func TestRangeValidatorIsValidInputMinusSignWithDigitsPartialNegRange(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	if !v.IsValidInput("-42", false) {
		t.Error("IsValidInput(\"-42\", false) should return true when min < 0")
	}
}

// TestRangeValidatorIsValidInputNonDigitCharPartial verifies:
// "RangeValidator.IsValidInput("abc", false) returns false"
func TestRangeValidatorIsValidInputNonDigitCharPartial(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if v.IsValidInput("abc", false) {
		t.Error("IsValidInput(\"abc\", false) should return false")
	}
}

// TestRangeValidatorIsValidInputInvalidCharInMiddlePartial verifies:
// non-digit character in middle of string is rejected
func TestRangeValidatorIsValidInputInvalidCharInMiddlePartial(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if v.IsValidInput("1a3", false) {
		t.Error("IsValidInput(\"1a3\", false) should return false")
	}
}

// TestRangeValidatorIsValidInputInvalidCharAtEndPartial verifies:
// non-digit character at end is rejected
func TestRangeValidatorIsValidInputInvalidCharAtEndPartial(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if v.IsValidInput("42x", false) {
		t.Error("IsValidInput(\"42x\", false) should return false")
	}
}

// TestRangeValidatorIsValidInputMultipleSignsPartial verifies:
// multiple signs are not accepted
func TestRangeValidatorIsValidInputMultipleSignsPartial(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	if v.IsValidInput("++5", false) {
		t.Error("IsValidInput(\"++5\", false) should return false")
	}
}

// TestRangeValidatorIsValidInputSignInMiddlePartial verifies:
// sign in middle of number is not accepted
func TestRangeValidatorIsValidInputSignInMiddlePartial(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	if v.IsValidInput("1-3", false) {
		t.Error("IsValidInput(\"1-3\", false) should return false")
	}
}

// TestRangeValidatorIsValidInputOnlyPlusSignPartial verifies:
// only a plus sign with no digits is valid as partial input
func TestRangeValidatorIsValidInputOnlyPlusSignPartial(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if !v.IsValidInput("+", false) {
		t.Error("IsValidInput(\"+\", false) should return true as partial input")
	}
}

// TestRangeValidatorIsValidInputOnlyMinusSignPartialNegRange verifies:
// only a minus sign with no digits is valid as partial input when min < 0
func TestRangeValidatorIsValidInputOnlyMinusSignPartialNegRange(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	if !v.IsValidInput("-", false) {
		t.Error("IsValidInput(\"-\", false) should return true as partial input when min < 0")
	}
}

// TestRangeValidatorIsValidInRangeCommitted verifies:
// "RangeValidator.IsValid("42") returns true when 42 is within [min, max]"
func TestRangeValidatorIsValidInRangeCommitted(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if !v.IsValid("42") {
		t.Error("IsValid(\"42\") should return true when 42 is within [0, 100]")
	}
}

// TestRangeValidatorIsValidOutOfRangeCommitted verifies:
// "RangeValidator.IsValid("42") returns false when 42 is outside [min, max]"
func TestRangeValidatorIsValidOutOfRangeCommitted(t *testing.T) {
	v := NewRangeValidator(50, 100)
	if v.IsValid("42") {
		t.Error("IsValid(\"42\") should return false when 42 is outside [50, 100]")
	}
}

// TestRangeValidatorIsValidOutOfRangeHighCommitted verifies:
// value above max is rejected
func TestRangeValidatorIsValidOutOfRangeHighCommitted(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if v.IsValid("150") {
		t.Error("IsValid(\"150\") should return false when 150 > 100")
	}
}

// TestRangeValidatorIsValidBoundaryLowCommitted verifies:
// value at lower boundary is accepted
func TestRangeValidatorIsValidBoundaryLowCommitted(t *testing.T) {
	v := NewRangeValidator(10, 100)
	if !v.IsValid("10") {
		t.Error("IsValid(\"10\") should return true when at lower boundary")
	}
}

// TestRangeValidatorIsValidBoundaryHighCommitted verifies:
// value at upper boundary is accepted
func TestRangeValidatorIsValidBoundaryHighCommitted(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if !v.IsValid("100") {
		t.Error("IsValid(\"100\") should return true when at upper boundary")
	}
}

// TestRangeValidatorIsValidNegativeInRange verifies:
// negative values are accepted when in range
func TestRangeValidatorIsValidNegativeInRange(t *testing.T) {
	v := NewRangeValidator(-100, 0)
	if !v.IsValid("-50") {
		t.Error("IsValid(\"-50\") should return true when -50 is within [-100, 0]")
	}
}

// TestRangeValidatorIsValidNegativeOutOfRange verifies:
// negative values are rejected when out of range
func TestRangeValidatorIsValidNegativeOutOfRange(t *testing.T) {
	v := NewRangeValidator(-50, 0)
	if v.IsValid("-100") {
		t.Error("IsValid(\"-100\") should return false when -100 < -50")
	}
}

// TestRangeValidatorIsValidZeroInRange verifies:
// zero is accepted when in range
func TestRangeValidatorIsValidZeroInRange(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	if !v.IsValid("0") {
		t.Error("IsValid(\"0\") should return true when 0 is within range")
	}
}

// TestRangeValidatorIsValidEmptyStringCommitted verifies:
// "RangeValidator.IsValid("") returns false (empty string is not a valid committed value)"
func TestRangeValidatorIsValidEmptyStringCommitted(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if v.IsValid("") {
		t.Error("IsValid(\"\") should return false")
	}
}

// TestRangeValidatorIsValidSignAloneCommitted verifies:
// "RangeValidator.IsValid("+") returns false (sign alone is not a valid number)"
func TestRangeValidatorIsValidSignAloneCommitted(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if v.IsValid("+") {
		t.Error("IsValid(\"+\") should return false")
	}
}

// TestRangeValidatorIsValidMinusSignAloneCommitted verifies:
// minus sign alone is not a valid committed number
func TestRangeValidatorIsValidMinusSignAloneCommitted(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	if v.IsValid("-") {
		t.Error("IsValid(\"-\") should return false")
	}
}

// TestRangeValidatorIsValidOverflowCommitted verifies:
// "RangeValidator.IsValid("99999999999999") returns false (overflow beyond int range)"
func TestRangeValidatorIsValidOverflowCommitted(t *testing.T) {
	v := NewRangeValidator(0, 1000000000000)
	if v.IsValid("99999999999999") {
		t.Error("IsValid(\"99999999999999\") should return false due to int overflow")
	}
}

// TestRangeValidatorIsValidLargeNegativeOverflow verifies:
// large negative numbers beyond int range are rejected
func TestRangeValidatorIsValidLargeNegativeOverflow(t *testing.T) {
	v := NewRangeValidator(-1000000000000, 0)
	if v.IsValid("-99999999999999") {
		t.Error("IsValid(\"-99999999999999\") should return false due to int overflow")
	}
}

// TestRangeValidatorIsValidWithPlusSignCommitted verifies:
// value with explicit plus sign is accepted if in range
func TestRangeValidatorIsValidWithPlusSignCommitted(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if !v.IsValid("+42") {
		t.Error("IsValid(\"+42\") should return true when 42 is within range")
	}
}

// TestRangeValidatorIsValidWithMinusSignCommitted verifies:
// value with minus sign is accepted if in range
func TestRangeValidatorIsValidWithMinusSignCommitted(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	if !v.IsValid("-42") {
		t.Error("IsValid(\"-42\") should return true when -42 is within range")
	}
}

// TestRangeValidatorIsValidWithMinusSignPositiveRangeCommitted verifies:
// value with minus sign is rejected when min >= 0
func TestRangeValidatorIsValidWithMinusSignPositiveRangeCommitted(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if v.IsValid("-42") {
		t.Error("IsValid(\"-42\") should return false when min >= 0")
	}
}

// TestRangeValidatorIsValidNonDigitCharCommitted verifies:
// non-digit characters are not valid in committed input
func TestRangeValidatorIsValidNonDigitCharCommitted(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if v.IsValid("42a") {
		t.Error("IsValid(\"42a\") should return false")
	}
}

// TestRangeValidatorIsValidSpaceInNumberCommitted verifies:
// spaces in number are not valid
func TestRangeValidatorIsValidSpaceInNumberCommitted(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if v.IsValid("4 2") {
		t.Error("IsValid(\"4 2\") should return false")
	}
}

// TestRangeValidatorIsValidLeadingZeroCommitted verifies:
// leading zeros are accepted as valid numbers
func TestRangeValidatorIsValidLeadingZeroCommitted(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if !v.IsValid("042") {
		t.Error("IsValid(\"042\") should return true (leading zeros accepted)")
	}
}

// TestRangeValidatorIsValidOnlyZeroCommitted verifies:
// just "0" is a valid committed value
func TestRangeValidatorIsValidOnlyZeroCommitted(t *testing.T) {
	v := NewRangeValidator(0, 100)
	if !v.IsValid("0") {
		t.Error("IsValid(\"0\") should return true")
	}
}

// TestRangeValidatorErrorCallable verifies:
// "RangeValidator.Error() is callable"
func TestRangeValidatorErrorCallable(t *testing.T) {
	v := NewRangeValidator(0, 100)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Error() should not panic, but recovered: %v", r)
		}
	}()
	v.Error()
}

// TestRangeValidatorErrorCallableNegRange verifies:
// Error() is callable even with negative range
func TestRangeValidatorErrorCallableNegRange(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Error() should not panic, but recovered: %v", r)
		}
	}()
	v.Error()
}

// TestRangeValidatorMinEqualsMaxInRange verifies:
// when min == max, only that value is in range
func TestRangeValidatorMinEqualsMaxInRange(t *testing.T) {
	v := NewRangeValidator(42, 42)
	if !v.IsValid("42") {
		t.Error("IsValid(\"42\") should return true when min == max == 42")
	}
}

// TestRangeValidatorMinEqualsMaxOutOfRange verifies:
// when min == max, other values are out of range
func TestRangeValidatorMinEqualsMaxOutOfRange(t *testing.T) {
	v := NewRangeValidator(42, 42)
	if v.IsValid("41") {
		t.Error("IsValid(\"41\") should return false when min == max == 42")
	}
}

// TestRangeValidatorNegativeMinMaxInRange verifies:
// when both min and max are negative, negative values can be in range
func TestRangeValidatorNegativeMinMaxInRange(t *testing.T) {
	v := NewRangeValidator(-100, -10)
	if !v.IsValid("-50") {
		t.Error("IsValid(\"-50\") should return true when -50 is within [-100, -10]")
	}
}

// TestRangeValidatorNegativeMinMaxOutOfRange verifies:
// positive values are out of range when both bounds are negative
func TestRangeValidatorNegativeMinMaxOutOfRange(t *testing.T) {
	v := NewRangeValidator(-100, -10)
	if v.IsValid("50") {
		t.Error("IsValid(\"50\") should return false when 50 > -10")
	}
}

// TestRangeValidatorIsValidInputAcceptsOnlyZeroNineWithPositiveMin verifies:
// "When min >= 0, only "+0123456789" are accepted per-keystroke (no minus sign)"
func TestRangeValidatorIsValidInputAcceptsOnlyZeroNineWithPositiveMin(t *testing.T) {
	v := NewRangeValidator(0, 100)
	// Test each allowed character individually
	for _, c := range "+0123456789" {
		input := string(c)
		if !v.IsValidInput(input, false) {
			t.Errorf("IsValidInput(\"%s\", false) should return true when min >= 0", input)
		}
	}
}

// TestRangeValidatorIsValidInputAcceptsSignsWithNegativeMin verifies:
// "When min < 0, "+-0123456789" are accepted per-keystroke"
func TestRangeValidatorIsValidInputAcceptsSignsWithNegativeMin(t *testing.T) {
	v := NewRangeValidator(-100, 100)
	// Test each allowed character individually
	for _, c := range "+-0123456789" {
		input := string(c)
		if !v.IsValidInput(input, false) {
			t.Errorf("IsValidInput(\"%s\", false) should return true when min < 0", input)
		}
	}
}

// TestRangeValidatorMultipleInstancesIndependent verifies:
// multiple RangeValidator instances are independent
func TestRangeValidatorMultipleInstancesIndependent(t *testing.T) {
	v1 := NewRangeValidator(0, 50)
	v2 := NewRangeValidator(50, 100)

	if !v1.IsValid("25") {
		t.Error("v1.IsValid(\"25\") should return true")
	}
	if v1.IsValid("75") {
		t.Error("v1.IsValid(\"75\") should return false")
	}
	if !v2.IsValid("75") {
		t.Error("v2.IsValid(\"75\") should return true")
	}
	if v2.IsValid("25") {
		t.Error("v2.IsValid(\"25\") should return false")
	}
}

// StringLookupValidator tests

// TestStringLookupValidatorNewReturnsValidator verifies:
// "NewStringLookupValidator(items []string) returns a *StringLookupValidator that implements Validator"
func TestStringLookupValidatorNewReturnsValidator(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo", "bar"})
	var _ Validator = v
	if v == nil {
		t.Error("NewStringLookupValidator should return non-nil value")
	}
}

// TestStringLookupValidatorIsValidInputAlwaysTrue verifies:
// "StringLookupValidator.IsValidInput always returns true (any partial input is acceptable)"
func TestStringLookupValidatorIsValidInputAlwaysTrue(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo", "bar"})
	if !v.IsValidInput("f", false) {
		t.Error("IsValidInput(\"f\", false) should return true")
	}
	if !v.IsValidInput("fo", false) {
		t.Error("IsValidInput(\"fo\", false) should return true")
	}
	if !v.IsValidInput("baz", false) {
		t.Error("IsValidInput(\"baz\", false) should return true")
	}
}

// TestStringLookupValidatorIsValidInputEmptyString verifies:
// empty string is valid as partial input
func TestStringLookupValidatorIsValidInputEmptyString(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo"})
	if !v.IsValidInput("", false) {
		t.Error("IsValidInput(\"\", false) should return true")
	}
}

// TestStringLookupValidatorIsValidInputNonexistentString verifies:
// IsValidInput returns true even for strings not in the list
func TestStringLookupValidatorIsValidInputNonexistentString(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo"})
	if !v.IsValidInput("xyz", false) {
		t.Error("IsValidInput(\"xyz\", false) should return true")
	}
}

// TestStringLookupValidatorIsValidInputNoAutoFillTrue verifies:
// IsValidInput with noAutoFill=true also returns true
func TestStringLookupValidatorIsValidInputNoAutoFillTrue(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo"})
	if !v.IsValidInput("f", true) {
		t.Error("IsValidInput(\"f\", true) should return true")
	}
}

// TestStringLookupValidatorIsValidInList verifies:
// "StringLookupValidator.IsValid(\"foo\") returns true when \"foo\" is in the items list"
func TestStringLookupValidatorIsValidInList(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo", "bar"})
	if !v.IsValid("foo") {
		t.Error("IsValid(\"foo\") should return true when \"foo\" is in the list")
	}
}

// TestStringLookupValidatorIsValidMultipleItems verifies:
// IsValid returns true for multiple items in the list
func TestStringLookupValidatorIsValidMultipleItems(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo", "bar", "baz"})
	if !v.IsValid("foo") {
		t.Error("IsValid(\"foo\") should return true")
	}
	if !v.IsValid("bar") {
		t.Error("IsValid(\"bar\") should return true")
	}
	if !v.IsValid("baz") {
		t.Error("IsValid(\"baz\") should return true")
	}
}

// TestStringLookupValidatorIsValidNotInList verifies:
// "StringLookupValidator.IsValid(\"foo\") returns false when \"foo\" is not in the items list"
func TestStringLookupValidatorIsValidNotInList(t *testing.T) {
	v := NewStringLookupValidator([]string{"bar"})
	if v.IsValid("foo") {
		t.Error("IsValid(\"foo\") should return false when \"foo\" is not in the list")
	}
}

// TestStringLookupValidatorIsValidCaseSensitive verifies:
// "Matching is case-sensitive: IsValid(\"Foo\") returns false when only \"foo\" is in the list"
func TestStringLookupValidatorIsValidCaseSensitive(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo"})
	if v.IsValid("Foo") {
		t.Error("IsValid(\"Foo\") should return false when only \"foo\" is in the list (case-sensitive)")
	}
	if v.IsValid("FOO") {
		t.Error("IsValid(\"FOO\") should return false when only \"foo\" is in the list (case-sensitive)")
	}
}

// TestStringLookupValidatorIsValidEmptyStringInList verifies:
// "StringLookupValidator.IsValid(\"\") returns false when \"\" is not in the list"
func TestStringLookupValidatorIsValidEmptyStringInList(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo", "bar"})
	if v.IsValid("") {
		t.Error("IsValid(\"\") should return false when empty string is not in the list")
	}
}

// TestStringLookupValidatorIsValidEmptyStringExplicit verifies:
// IsValid returns true when empty string is explicitly in the list
func TestStringLookupValidatorIsValidEmptyStringExplicit(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo", ""})
	if !v.IsValid("") {
		t.Error("IsValid(\"\") should return true when empty string is in the list")
	}
}

// TestStringLookupValidatorEmptyItemsList verifies:
// "StringLookupValidator.IsValid(anything) returns false when items list is empty"
func TestStringLookupValidatorEmptyItemsList(t *testing.T) {
	v := NewStringLookupValidator([]string{})
	if v.IsValid("foo") {
		t.Error("IsValid(\"foo\") should return false when items list is empty")
	}
	if v.IsValid("") {
		t.Error("IsValid(\"\") should return false when items list is empty")
	}
}

// TestStringLookupValidatorEmptyItemsListPartialInput verifies:
// IsValidInput still returns true with empty items list
func TestStringLookupValidatorEmptyItemsListPartialInput(t *testing.T) {
	v := NewStringLookupValidator([]string{})
	if !v.IsValidInput("foo", false) {
		t.Error("IsValidInput(\"foo\", false) should return true even when items list is empty")
	}
}

// TestStringLookupValidatorErrorCallable verifies:
// "StringLookupValidator.Error() is callable"
func TestStringLookupValidatorErrorCallable(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo"})
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Error() should not panic, but recovered: %v", r)
		}
	}()
	v.Error()
}

// TestStringLookupValidatorMultipleInstances verifies:
// "multiple independent StringLookupValidator instances"
func TestStringLookupValidatorMultipleInstances(t *testing.T) {
	v1 := NewStringLookupValidator([]string{"foo", "bar"})
	v2 := NewStringLookupValidator([]string{"baz", "qux"})

	if !v1.IsValid("foo") {
		t.Error("v1.IsValid(\"foo\") should return true")
	}
	if v1.IsValid("baz") {
		t.Error("v1.IsValid(\"baz\") should return false")
	}
	if !v2.IsValid("baz") {
		t.Error("v2.IsValid(\"baz\") should return true")
	}
	if v2.IsValid("foo") {
		t.Error("v2.IsValid(\"foo\") should return false")
	}
}

// TestStringLookupValidatorSingleItem verifies:
// StringLookupValidator with a single item works correctly
func TestStringLookupValidatorSingleItem(t *testing.T) {
	v := NewStringLookupValidator([]string{"only"})
	if !v.IsValid("only") {
		t.Error("IsValid(\"only\") should return true")
	}
	if v.IsValid("other") {
		t.Error("IsValid(\"other\") should return false")
	}
}

// TestStringLookupValidatorDuplicateItems verifies:
// duplicates in items list are handled (map deduplicates them)
func TestStringLookupValidatorDuplicateItems(t *testing.T) {
	v := NewStringLookupValidator([]string{"foo", "foo", "bar"})
	if !v.IsValid("foo") {
		t.Error("IsValid(\"foo\") should return true")
	}
	if !v.IsValid("bar") {
		t.Error("IsValid(\"bar\") should return true")
	}
}

// TestStringLookupValidatorSpecialCharacters verifies:
// StringLookupValidator handles special characters in strings
func TestStringLookupValidatorSpecialCharacters(t *testing.T) {
	v := NewStringLookupValidator([]string{"@user", "#tag", "hello-world"})
	if !v.IsValid("@user") {
		t.Error("IsValid(\"@user\") should return true")
	}
	if !v.IsValid("#tag") {
		t.Error("IsValid(\"#tag\") should return true")
	}
	if !v.IsValid("hello-world") {
		t.Error("IsValid(\"hello-world\") should return true")
	}
}

// TestStringLookupValidatorWhitespace verifies:
// IsValid distinguishes between strings with and without whitespace
func TestStringLookupValidatorWhitespace(t *testing.T) {
	v := NewStringLookupValidator([]string{"hello world"})
	if !v.IsValid("hello world") {
		t.Error("IsValid(\"hello world\") should return true")
	}
	if v.IsValid("helloworld") {
		t.Error("IsValid(\"helloworld\") should return false")
	}
}

// TestStringLookupValidatorUnicode verifies:
// StringLookupValidator handles Unicode characters
func TestStringLookupValidatorUnicode(t *testing.T) {
	v := NewStringLookupValidator([]string{"café", "naïve", "résumé"})
	if !v.IsValid("café") {
		t.Error("IsValid(\"café\") should return true")
	}
	if v.IsValid("cafe") {
		t.Error("IsValid(\"cafe\") should return false (not \"café\")")
	}
}

// TestStringLookupValidatorLargeList verifies:
// StringLookupValidator works with large item lists
func TestStringLookupValidatorLargeList(t *testing.T) {
	items := []string{}
	for i := 0; i < 1000; i++ {
		items = append(items, fmt.Sprintf("item_%d", i))
	}
	v := NewStringLookupValidator(items)
	if !v.IsValid("item_0") {
		t.Error("IsValid(\"item_0\") should return true")
	}
	if !v.IsValid("item_999") {
		t.Error("IsValid(\"item_999\") should return true")
	}
	if v.IsValid("item_1000") {
		t.Error("IsValid(\"item_1000\") should return false")
	}
}

// PictureValidator tests

// TestPictureValidatorImplementsValidator verifies:
// NewPictureValidator returns a value that implements Validator
func TestPictureValidatorImplementsValidator(t *testing.T) {
	v := NewPictureValidator("###", false)
	var _ Validator = v
}

// TestPictureValidatorImplementsValidatorWithAutoFill verifies:
// NewPictureValidator returns a value that implements ValidatorWithAutoFill
func TestPictureValidatorImplementsValidatorWithAutoFill(t *testing.T) {
	v := NewPictureValidator("###", false)
	var _ ValidatorWithAutoFill = v
}

// TestPictureValidatorDigitPatternPartialOnedigit verifies:
// IsValidInput("1", false) returns true for picture "###" (one digit, partial match)
func TestPictureValidatorDigitPatternPartialOnedigit(t *testing.T) {
	v := NewPictureValidator("###", false)
	if !v.IsValidInput("1", false) {
		t.Error("IsValidInput(\"1\", false) should return true for picture \"###\" — partial match is accepted")
	}
}

// TestPictureValidatorDigitPatternPartialTwodigits verifies:
// IsValidInput("12", false) returns true for picture "###" (two digits, partial match)
func TestPictureValidatorDigitPatternPartialTwodigits(t *testing.T) {
	v := NewPictureValidator("###", false)
	if !v.IsValidInput("12", false) {
		t.Error("IsValidInput(\"12\", false) should return true for picture \"###\" — partial match is accepted")
	}
}

// TestPictureValidatorDigitPatternFullMatch verifies:
// IsValidInput("123", false) returns true for picture "###" (full match)
func TestPictureValidatorDigitPatternFullMatch(t *testing.T) {
	v := NewPictureValidator("###", false)
	if !v.IsValidInput("123", false) {
		t.Error("IsValidInput(\"123\", false) should return true for picture \"###\" — full match")
	}
}

// TestPictureValidatorDigitPatternExceedsPattern verifies:
// IsValidInput("1234", false) returns false for picture "###" (too many characters)
func TestPictureValidatorDigitPatternExceedsPattern(t *testing.T) {
	v := NewPictureValidator("###", false)
	if v.IsValidInput("1234", false) {
		t.Error("IsValidInput(\"1234\", false) should return false for picture \"###\" — exceeds pattern length")
	}
}

// TestPictureValidatorDigitPatternRejectsLetter verifies:
// IsValidInput("a", false) returns false for picture "###" (letter does not match digit slot)
func TestPictureValidatorDigitPatternRejectsLetter(t *testing.T) {
	v := NewPictureValidator("###", false)
	if v.IsValidInput("a", false) {
		t.Error("IsValidInput(\"a\", false) should return false for picture \"###\" — letter does not satisfy '#'")
	}
}

// TestPictureValidatorDigitPatternIsValidComplete verifies:
// IsValid("123") returns true for picture "###" (fully matches)
func TestPictureValidatorDigitPatternIsValidComplete(t *testing.T) {
	v := NewPictureValidator("###", false)
	if !v.IsValid("123") {
		t.Error("IsValid(\"123\") should return true for picture \"###\" — complete match")
	}
}

// TestPictureValidatorDigitPatternIsValidIncomplete verifies:
// IsValid("12") returns false for picture "###" (incomplete — not all positions filled)
func TestPictureValidatorDigitPatternIsValidIncomplete(t *testing.T) {
	v := NewPictureValidator("###", false)
	if v.IsValid("12") {
		t.Error("IsValid(\"12\") should return false for picture \"###\" — incomplete input is not fully valid")
	}
}

// TestPictureValidatorDigitPatternIsValidEmptyString verifies:
// IsValid("") returns false for picture "###" (empty is incomplete)
func TestPictureValidatorDigitPatternIsValidEmptyString(t *testing.T) {
	v := NewPictureValidator("###", false)
	if v.IsValid("") {
		t.Error("IsValid(\"\") should return false for picture \"###\" — pattern requires digits")
	}
}

// TestPictureValidatorLetterPatternAcceptsLetter verifies:
// IsValidInput("a", false) returns true for picture "???" (letter matches '?')
func TestPictureValidatorLetterPatternAcceptsLetter(t *testing.T) {
	v := NewPictureValidator("???", false)
	if !v.IsValidInput("a", false) {
		t.Error("IsValidInput(\"a\", false) should return true for picture \"???\" — letter satisfies '?'")
	}
}

// TestPictureValidatorLetterPatternRejectsDigit verifies:
// IsValidInput("1", false) returns false for picture "???" (digit does not match '?')
func TestPictureValidatorLetterPatternRejectsDigit(t *testing.T) {
	v := NewPictureValidator("???", false)
	if v.IsValidInput("1", false) {
		t.Error("IsValidInput(\"1\", false) should return false for picture \"???\" — digit does not satisfy '?'")
	}
}

// TestPictureValidatorLetterPatternFullMatch verifies:
// IsValidInput("abc", false) returns true for picture "???" (full letter match)
func TestPictureValidatorLetterPatternFullMatch(t *testing.T) {
	v := NewPictureValidator("???", false)
	if !v.IsValidInput("abc", false) {
		t.Error("IsValidInput(\"abc\", false) should return true for picture \"???\" — full letter match")
	}
}

// TestPictureValidatorLetterPatternIsValidComplete verifies:
// IsValid("abc") returns true for picture "???" (fully matches)
func TestPictureValidatorLetterPatternIsValidComplete(t *testing.T) {
	v := NewPictureValidator("???", false)
	if !v.IsValid("abc") {
		t.Error("IsValid(\"abc\") should return true for picture \"???\" — complete match")
	}
}

// TestPictureValidatorAutoUppercaseAcceptsLowercase verifies:
// IsValidInput("abc", false) returns true for picture "&&&"
func TestPictureValidatorAutoUppercaseAcceptsLowercase(t *testing.T) {
	v := NewPictureValidator("&&&", false)
	if !v.IsValidInput("abc", false) {
		t.Error("IsValidInput(\"abc\", false) should return true for picture \"&&&\" — letters satisfy '&'")
	}
}

// TestPictureValidatorAutoUppercaseRejectsDigit verifies:
// IsValidInput("1", false) returns false for picture "&&&" (digit does not match '&')
func TestPictureValidatorAutoUppercaseRejectsDigit(t *testing.T) {
	v := NewPictureValidator("&&&", false)
	if v.IsValidInput("1", false) {
		t.Error("IsValidInput(\"1\", false) should return false for picture \"&&&\" — digit does not satisfy '&'")
	}
}

// TestPictureValidatorAutoUppercaseModifiesInput verifies:
// '&' auto-uppercases letters — IsValidInputAutoFill("abc") mutates to "ABC"
func TestPictureValidatorAutoUppercaseModifiesInput(t *testing.T) {
	v := NewPictureValidator("&&&", false)
	vaf, ok := v.(ValidatorWithAutoFill)
	if !ok {
		t.Fatal("PictureValidator must implement ValidatorWithAutoFill")
	}
	s := "abc"
	if !vaf.IsValidInputAutoFill(&s, false) {
		t.Error("IsValidInputAutoFill should return true for \"abc\" with picture \"&&&\"")
	}
	if s != "ABC" {
		t.Errorf("IsValidInputAutoFill should uppercase the input: got %q, want \"ABC\"", s)
	}
}

// TestPictureValidatorAutoUppercaseNoAutoFillSkipsMutation verifies:
// when noAutoFill=true, auto-uppercase does not modify the string
func TestPictureValidatorAutoUppercaseNoAutoFillSkipsMutation(t *testing.T) {
	v := NewPictureValidator("&&&", false)
	vaf, ok := v.(ValidatorWithAutoFill)
	if !ok {
		t.Fatal("PictureValidator must implement ValidatorWithAutoFill")
	}
	s := "abc"
	result := vaf.IsValidInputAutoFill(&s, true)
	// Must still validate (letters satisfy '&'), but must NOT uppercase
	if !result {
		t.Error("IsValidInputAutoFill should still return true with noAutoFill=true")
	}
	if s != "abc" {
		t.Errorf("IsValidInputAutoFill must not modify string when noAutoFill=true: got %q", s)
	}
}

// TestPictureValidatorAnyCharAcceptsAnything verifies:
// '@' matches any character as-is
func TestPictureValidatorAnyCharAcceptsAnything(t *testing.T) {
	v := NewPictureValidator("@@@", false)
	inputs := []string{"abc", "123", "!@#"}
	for _, input := range inputs {
		s := input
		if !v.IsValidInput(s, false) {
			t.Errorf("IsValidInput(%q, false) should return true for picture \"@@@\" — '@' accepts any char", s)
		}
	}
}

// TestPictureValidatorAnyCharDoesNotUppercase verifies:
// '@' does not uppercase — lowercase stays lowercase
func TestPictureValidatorAnyCharDoesNotUppercase(t *testing.T) {
	v := NewPictureValidator("@@@", false)
	vaf, ok := v.(ValidatorWithAutoFill)
	if !ok {
		t.Fatal("PictureValidator must implement ValidatorWithAutoFill")
	}
	s := "abc"
	vaf.IsValidInputAutoFill(&s, false)
	if s != "abc" {
		t.Errorf("'@' should not uppercase: got %q, want \"abc\"", s)
	}
}

// TestPictureValidatorBangAcceptsAnything verifies:
// '!' matches any character
func TestPictureValidatorBangAcceptsAnything(t *testing.T) {
	v := NewPictureValidator("!!!", false)
	inputs := []string{"abc", "123", "A!z"}
	for _, input := range inputs {
		if !v.IsValidInput(input, false) {
			t.Errorf("IsValidInput(%q, false) should return true for picture \"!!!\" — '!' accepts any char", input)
		}
	}
}

// TestPictureValidatorBangAutoUppercase verifies:
// '!' auto-uppercases letters (any char, but uppercased)
func TestPictureValidatorBangAutoUppercase(t *testing.T) {
	v := NewPictureValidator("!!!", false)
	vaf, ok := v.(ValidatorWithAutoFill)
	if !ok {
		t.Fatal("PictureValidator must implement ValidatorWithAutoFill")
	}
	s := "abc"
	if !vaf.IsValidInputAutoFill(&s, false) {
		t.Error("IsValidInputAutoFill should return true for \"abc\" with picture \"!!!\"")
	}
	if s != "ABC" {
		t.Errorf("'!' should uppercase letters: got %q, want \"ABC\"", s)
	}
}

// TestPictureValidatorEscapeRequiresLiteralHash verifies:
// ";#" requires a literal '#' character, not a digit
func TestPictureValidatorEscapeRequiresLiteralHash(t *testing.T) {
	v := NewPictureValidator(";#", false)
	if !v.IsValidInput("#", false) {
		t.Error("IsValidInput(\"#\", false) should return true for picture \";#\" — semicolon escapes '#' as literal")
	}
}

// TestPictureValidatorEscapeRejectsDigitWhenLiteralExpected verifies:
// ";#" rejects a digit — only a literal '#' is accepted
func TestPictureValidatorEscapeRejectsDigitWhenLiteralExpected(t *testing.T) {
	v := NewPictureValidator(";#", false)
	if v.IsValidInput("5", false) {
		t.Error("IsValidInput(\"5\", false) should return false for picture \";#\" — '5' does not match literal '#'")
	}
}

// TestPictureValidatorEscapeIsValidLiteralHash verifies:
// IsValid("#") returns true for picture ";#"
func TestPictureValidatorEscapeIsValidLiteralHash(t *testing.T) {
	v := NewPictureValidator(";#", false)
	if !v.IsValid("#") {
		t.Error("IsValid(\"#\") should return true for picture \";#\" — literal '#' fully matches")
	}
}

// TestPictureValidatorOptionalGroupEmptyIsValid verifies:
// IsValid("") returns true for picture "[###]" (all content is optional)
func TestPictureValidatorOptionalGroupEmptyIsValid(t *testing.T) {
	v := NewPictureValidator("[###]", false)
	if !v.IsValid("") {
		t.Error("IsValid(\"\") should return true for picture \"[###]\" — optional group allows empty input")
	}
}

// TestPictureValidatorOptionalGroupPartialIsValid verifies:
// IsValid("123") returns true for picture "###[#]" (3 required + 1 optional filled)
func TestPictureValidatorOptionalGroupPartialIsValid(t *testing.T) {
	v := NewPictureValidator("###[#]", false)
	if !v.IsValid("123") {
		t.Error("IsValid(\"123\") should return true for picture \"###[#]\" — required digits filled, optional absent")
	}
}

// TestPictureValidatorOptionalGroupFullIsValid verifies:
// IsValid("1234") returns true for picture "###[#]" (all positions filled)
func TestPictureValidatorOptionalGroupFullIsValid(t *testing.T) {
	v := NewPictureValidator("###[#]", false)
	if !v.IsValid("1234") {
		t.Error("IsValid(\"1234\") should return true for picture \"###[#]\" — all positions filled")
	}
}

// TestPictureValidatorOptionalGroupExceedsIsInvalid verifies:
// IsValidInput("12345", false) returns false for picture "###[#]" (too many chars)
func TestPictureValidatorOptionalGroupExceedsIsInvalid(t *testing.T) {
	v := NewPictureValidator("###[#]", false)
	if v.IsValidInput("12345", false) {
		t.Error("IsValidInput(\"12345\", false) should return false for picture \"###[#]\" — exceeds pattern")
	}
}

// TestPictureValidatorRequiredGroupBehavesLikeNoGrouping verifies:
// "{###}" behaves the same as "###" for validation
func TestPictureValidatorRequiredGroupBehavesLikeNoGrouping(t *testing.T) {
	vGroup := NewPictureValidator("{###}", false)
	vPlain := NewPictureValidator("###", false)

	tests := []string{"", "1", "12", "123", "1234", "abc"}
	for _, s := range tests {
		if vGroup.IsValidInput(s, false) != vPlain.IsValidInput(s, false) {
			t.Errorf("IsValidInput(%q) differs between \"{###}\" and \"###\": got %v vs %v",
				s, vGroup.IsValidInput(s, false), vPlain.IsValidInput(s, false))
		}
		if vGroup.IsValid(s) != vPlain.IsValid(s) {
			t.Errorf("IsValid(%q) differs between \"{###}\" and \"###\": got %v vs %v",
				s, vGroup.IsValid(s), vPlain.IsValid(s))
		}
	}
}

// TestPictureValidatorAutoFillLiterals verifies:
// literal characters in the picture are auto-filled into the string
// Picture "(###) ###-####" with input "123" — parenthesis and space are auto-inserted
func TestPictureValidatorAutoFillLiterals(t *testing.T) {
	v := NewPictureValidator("(###) ###-####", true)
	vaf, ok := v.(ValidatorWithAutoFill)
	if !ok {
		t.Fatal("PictureValidator must implement ValidatorWithAutoFill")
	}
	s := "123"
	result := vaf.IsValidInputAutoFill(&s, false)
	if !result {
		t.Error("IsValidInputAutoFill should return true for \"123\" with phone picture")
	}
	// After auto-fill, leading literal '(' should be present in the string
	if len(s) == 0 || s[0] != '(' {
		t.Errorf("Auto-fill should insert leading '(' literal: got %q", s)
	}
}

// TestPictureValidatorAutoFillDisabledWhenNoAutoFill verifies:
// when noAutoFill=true, literals are not auto-inserted
func TestPictureValidatorAutoFillDisabledWhenNoAutoFill(t *testing.T) {
	v := NewPictureValidator("(###) ###-####", true)
	vaf, ok := v.(ValidatorWithAutoFill)
	if !ok {
		t.Fatal("PictureValidator must implement ValidatorWithAutoFill")
	}
	s := "123"
	vaf.IsValidInputAutoFill(&s, true)
	// With noAutoFill=true, the string must not have literals auto-inserted
	if len(s) > 0 && s[0] == '(' {
		t.Errorf("Auto-fill should NOT insert literals when noAutoFill=true: got %q", s)
	}
}

// TestPictureValidatorErrorCallable verifies:
// Error() is callable without panic
func TestPictureValidatorErrorCallable(t *testing.T) {
	v := NewPictureValidator("###", false)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Error() should not panic, but recovered: %v", r)
		}
	}()
	v.Error()
}

// TestPictureValidatorMultipleInstancesIndependent verifies:
// multiple PictureValidator instances are independent
func TestPictureValidatorMultipleInstancesIndependent(t *testing.T) {
	vDigits := NewPictureValidator("###", false)
	vLetters := NewPictureValidator("???", false)

	if !vDigits.IsValidInput("123", false) {
		t.Error("vDigits.IsValidInput(\"123\") should return true")
	}
	if vDigits.IsValidInput("abc", false) {
		t.Error("vDigits.IsValidInput(\"abc\") should return false")
	}
	if !vLetters.IsValidInput("abc", false) {
		t.Error("vLetters.IsValidInput(\"abc\") should return true")
	}
	if vLetters.IsValidInput("123", false) {
		t.Error("vLetters.IsValidInput(\"123\") should return false")
	}
}

// TestPictureValidatorEmptyPatternEmptyInput verifies:
// empty picture with empty input — empty input is fully valid
func TestPictureValidatorEmptyPatternEmptyInput(t *testing.T) {
	v := NewPictureValidator("", false)
	if !v.IsValid("") {
		t.Error("IsValid(\"\") should return true for empty picture — no pattern requirements")
	}
}

// TestPictureValidatorEmptyPatternNonEmptyInput verifies:
// empty picture rejects non-empty input (no slots to consume characters)
func TestPictureValidatorEmptyPatternNonEmptyInput(t *testing.T) {
	v := NewPictureValidator("", false)
	if v.IsValidInput("a", false) {
		t.Error("IsValidInput(\"a\", false) should return false for empty picture — no slots available")
	}
}
