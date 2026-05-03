package tv

import "strconv"

// Validator defines the two-phase validation interface for InputLine.
type Validator interface {
	IsValid(s string) bool
	IsValidInput(s string, noAutoFill bool) bool
	Error()
}

// FilterValidator accepts only characters from a whitelist.
type FilterValidator struct {
	validChars map[rune]bool
}

func NewFilterValidator(validChars string) *FilterValidator {
	m := make(map[rune]bool)
	for _, r := range validChars {
		m[r] = true
	}
	return &FilterValidator{validChars: m}
}

func (f *FilterValidator) IsValidInput(s string, noAutoFill bool) bool {
	for _, r := range s {
		if !f.validChars[r] {
			return false
		}
	}
	return true
}

func (f *FilterValidator) IsValid(s string) bool {
	return f.IsValidInput(s, true)
}

func (f *FilterValidator) Error() {
	// In production: show message box "Invalid character in input"
	// For unit testing: no-op to avoid panic when no Application is running
}

// RangeValidator validates integer input within a min/max range.
// It embeds FilterValidator for character-level filtering and adds range checking.
type RangeValidator struct {
	FilterValidator
	min, max int
}

// NewRangeValidator creates a RangeValidator that accepts integers in [min, max].
// When min >= 0, only "+0123456789" characters are accepted.
// When min < 0, "+-0123456789" characters are accepted.
func NewRangeValidator(min, max int) *RangeValidator {
	chars := "+-0123456789"
	if min >= 0 {
		chars = "+0123456789"
	}
	rv := &RangeValidator{min: min, max: max}
	rv.FilterValidator = *NewFilterValidator(chars)
	return rv
}

// IsValidInput validates partial numeric input during typing.
// Returns true for: empty string, lone sign (+/-), or valid partial numbers.
// Structural validation ensures: optional sign at position 0, then digits only.
func (r *RangeValidator) IsValidInput(s string, noAutoFill bool) bool {
	if s == "" {
		return true
	}

	// Check all characters are in the valid set first
	if !r.FilterValidator.IsValidInput(s, noAutoFill) {
		return false
	}

	// Structural validation: optional sign at start, then digits
	start := 0
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		start = 1
	}
	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// IsValid validates a complete integer input against the range [min, max].
// Returns false for: empty string, signs without digits, or values outside the range.
func (r *RangeValidator) IsValid(s string) bool {
	if s == "" {
		return false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return n >= r.min && n <= r.max
}

// Error is called when validation fails.
// In production: show message box; for unit testing: no-op.
func (r *RangeValidator) Error() {
	// In production: show message box "Value out of range"
	// For unit testing: no-op to avoid panic when no Application is running
}
