package tv

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
