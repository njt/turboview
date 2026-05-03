package tv

// dialog_validator_test.go — Tests for Dialog.Valid() validation before closing.
//
// Dialog.Valid(cmd CommandCode) checks all child InputLine validators before
// allowing the dialog to close. If any InputLine has an invalid value, Valid()
// focuses that field, calls Error(), and returns false. HandleEvent intercepts
// CmOK after all other processing and calls Valid(); if false, event is cleared.
//
// Test organization:
//  1. CmCancel behavior (always returns true)
//  2. CmOK with all valid InputLines
//  3. CmOK with one invalid InputLine
//  4. Focus shifting to invalid field
//  5. No validators present
//  6. No InputLines in dialog
//  7. HandleEvent CmOK blocking when Valid() is false
//  8. HandleEvent CmOK passing through when Valid() is true

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Test 1: Dialog.Valid(CmCancel) always returns true
// Spec: "Dialog.Valid(CmCancel) always returns true"
// ---------------------------------------------------------------------------

// TestDialogValidCmCancelAlwaysTrue verifies that Valid(CmCancel) returns true
// even when child InputLines have invalid text.
func TestDialogValidCmCancelAlwaysTrue(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Validate")

	// Insert an InputLine with a validator and invalid text.
	il := NewInputLine(NewRect(0, 0, 30, 1), 100)
	il.SetValidator(NewRangeValidator(1, 100))
	il.SetText("") // Empty string is invalid for RangeValidator
	dlg.Insert(il)

	// Valid(CmCancel) must return true regardless of invalid child.
	if !dlg.Valid(CmCancel) {
		t.Error("Valid(CmCancel) returned false, want true")
	}
}

// ---------------------------------------------------------------------------
// Test 2: Dialog.Valid(CmOK) returns true when all InputLines are valid
// Spec: "Valid(CmOK) returns true when all InputLines have valid text"
// ---------------------------------------------------------------------------

// TestDialogValidCmOKAllValid verifies that Valid(CmOK) returns true when all
// InputLines with validators contain valid text.
func TestDialogValidCmOKAllValid(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Validate")

	// Add two InputLines with validators, both with valid text.
	il1 := NewInputLine(NewRect(0, 0, 30, 1), 100)
	il1.SetValidator(NewRangeValidator(1, 100))
	il1.SetText("50")
	dlg.Insert(il1)

	il2 := NewInputLine(NewRect(0, 2, 30, 1), 100)
	il2.SetValidator(NewRangeValidator(1, 100))
	il2.SetText("75")
	dlg.Insert(il2)

	// Valid(CmOK) must return true when all validators pass.
	if !dlg.Valid(CmOK) {
		t.Error("Valid(CmOK) returned false when all InputLines are valid, want true")
	}
}

// ---------------------------------------------------------------------------
// Test 3: Dialog.Valid(CmOK) returns false when one InputLine is invalid
// Spec: "Valid(CmOK) returns false when any InputLine has invalid text"
// ---------------------------------------------------------------------------

// TestDialogValidCmOKOneInvalid verifies that Valid(CmOK) returns false when
// one of multiple InputLines has invalid text.
func TestDialogValidCmOKOneInvalid(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Validate")

	// Add first InputLine with valid text.
	il1 := NewInputLine(NewRect(0, 0, 30, 1), 100)
	il1.SetValidator(NewRangeValidator(1, 100))
	il1.SetText("50")
	dlg.Insert(il1)

	// Add second InputLine with invalid text.
	il2 := NewInputLine(NewRect(0, 2, 30, 1), 100)
	il2.SetValidator(NewRangeValidator(1, 100))
	il2.SetText("") // Empty is invalid
	dlg.Insert(il2)

	// Valid(CmOK) must return false when any InputLine is invalid.
	if dlg.Valid(CmOK) {
		t.Error("Valid(CmOK) returned true when one InputLine is invalid, want false")
	}
}

// ---------------------------------------------------------------------------
// Test 4: Dialog.Valid(CmOK) focuses the invalid InputLine
// Spec: "When valid fails, the invalid InputLine gets focus"
// ---------------------------------------------------------------------------

// TestDialogValidCmOKFocusesInvalidField verifies that when Valid(CmOK) fails,
// the invalid InputLine becomes the focused child.
func TestDialogValidCmOKFocusesInvalidField(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Validate")

	// Add first InputLine with valid text.
	il1 := NewInputLine(NewRect(0, 0, 30, 1), 100)
	il1.SetValidator(NewRangeValidator(1, 100))
	il1.SetText("50")
	dlg.Insert(il1)

	// Add second InputLine with invalid text.
	il2 := NewInputLine(NewRect(0, 2, 30, 1), 100)
	il2.SetValidator(NewRangeValidator(1, 100))
	il2.SetText("") // Empty is invalid
	dlg.Insert(il2)

	// Call Valid(CmOK) — should return false and focus il2.
	dlg.Valid(CmOK)

	if dlg.FocusedChild() != il2 {
		t.Errorf("Valid(CmOK) did not focus the invalid InputLine: FocusedChild() = %v, want il2", dlg.FocusedChild())
	}
}

// ---------------------------------------------------------------------------
// Test 5: Dialog.Valid(CmOK) calls validator.Error() on invalid field
// Spec: "calls validator.Error() when validation fails"
// ---------------------------------------------------------------------------

// TestDialogValidCmOKCallsError verifies that when an InputLine is invalid,
// Valid(CmOK) calls its validator's Error() method.
func TestDialogValidCmOKCallsError(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Validate")

	// Create a mock validator that tracks Error() calls.
	mockVal := &mockValidator{
		isValidFn: func(s string) bool { return false }, // Always invalid
		errorCalled: false,
	}

	il := NewInputLine(NewRect(0, 0, 30, 1), 100)
	il.SetValidator(mockVal)
	il.SetText("anything")
	dlg.Insert(il)

	dlg.Valid(CmOK)

	if !mockVal.errorCalled {
		t.Error("Valid(CmOK) did not call validator.Error() when validation failed")
	}
}

// ---------------------------------------------------------------------------
// Test 6: Dialog.Valid(CmOK) returns true when no InputLines have validators
// Spec: "If no children have validators, Valid() returns true"
// ---------------------------------------------------------------------------

// TestDialogValidNoValidators verifies that Valid(CmOK) returns true when
// InputLines are present but have no validators set.
func TestDialogValidNoValidators(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Validate")

	// Add InputLines without validators.
	il1 := NewInputLine(NewRect(0, 0, 30, 1), 100)
	il1.SetText("anything") // No validator
	dlg.Insert(il1)

	il2 := NewInputLine(NewRect(0, 2, 30, 1), 100)
	il2.SetText("goes") // No validator
	dlg.Insert(il2)

	// Valid(CmOK) must return true when no validators are present.
	if !dlg.Valid(CmOK) {
		t.Error("Valid(CmOK) returned false when no InputLines have validators, want true")
	}
}

// ---------------------------------------------------------------------------
// Test 7: Dialog.Valid(CmOK) returns true when no InputLines exist
// Spec: "If no children are InputLines, Valid() returns true"
// ---------------------------------------------------------------------------

// TestDialogValidNoInputLines verifies that Valid(CmOK) returns true when
// the dialog contains no InputLines.
func TestDialogValidNoInputLines(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Validate")

	// Add non-InputLine children.
	btn := NewButton(NewRect(0, 0, 10, 1), "OK", CmOK)
	dlg.Insert(btn)

	st := NewStaticText(NewRect(0, 2, 20, 1), "Some text")
	dlg.Insert(st)

	// Valid(CmOK) must return true when no InputLines exist.
	if !dlg.Valid(CmOK) {
		t.Error("Valid(CmOK) returned false when no InputLines exist, want true")
	}
}

// ---------------------------------------------------------------------------
// Test 8: Dialog.HandleEvent intercepts CmOK and blocks it if Valid() fails
// Spec: "HandleEvent intercepts CmOK after all other processing and calls Valid().
//        If false, event is cleared."
// ---------------------------------------------------------------------------

// TestDialogHandleEventCmOKBlocked verifies that when HandleEvent receives a CmOK
// command and Valid() returns false, the event is cleared (not forwarded).
func TestDialogHandleEventCmOKBlocked(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Validate")

	// Add an InputLine with invalid text.
	il := NewInputLine(NewRect(0, 0, 30, 1), 100)
	il.SetValidator(NewRangeValidator(1, 100))
	il.SetText("") // Empty is invalid
	dlg.Insert(il)

	// Create and send a CmOK event.
	ev := &Event{What: EvCommand, Command: CmOK}
	dlg.HandleEvent(ev)

	// Event must be cleared (not passed through).
	if !ev.IsCleared() {
		t.Error("HandleEvent did not clear CmOK event when Valid() returned false")
	}
}

// ---------------------------------------------------------------------------
// Test 9: Dialog.HandleEvent allows CmOK through if Valid() succeeds
// Spec: "HandleEvent allows CmOK to pass through when Valid() returns true"
// ---------------------------------------------------------------------------

// TestDialogHandleEventCmOKAllowed verifies that when HandleEvent receives a CmOK
// command and Valid() returns true, the event is NOT cleared.
func TestDialogHandleEventCmOKAllowed(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Validate")

	// Add an InputLine with valid text.
	il := NewInputLine(NewRect(0, 0, 30, 1), 100)
	il.SetValidator(NewRangeValidator(1, 100))
	il.SetText("50")
	dlg.Insert(il)

	// Create and send a CmOK event.
	ev := &Event{What: EvCommand, Command: CmOK}
	dlg.HandleEvent(ev)

	// Event must NOT be cleared (should pass through to Group/ExecView).
	if ev.IsCleared() {
		t.Error("HandleEvent cleared CmOK event when Valid() returned true, want event to pass through")
	}
}

// ---------------------------------------------------------------------------
// Test 10: Dialog.Valid checks only the first invalid InputLine
// Spec: "For each InputLine with a validator, calls IsValid(text).
//        If any fails, focuses that InputLine, calls Error(), and returns false."
// ---------------------------------------------------------------------------

// TestDialogValidStopsAtFirstInvalid verifies that when multiple InputLines have
// invalid text, Valid() focuses and reports error for the FIRST invalid one.
func TestDialogValidStopsAtFirstInvalid(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Validate")

	// Create mock validators to track which Error() was called.
	val1 := &mockValidator{
		isValidFn: func(s string) bool { return false }, // Invalid
		errorCalled: false,
	}
	val2 := &mockValidator{
		isValidFn: func(s string) bool { return false }, // Also invalid
		errorCalled: false,
	}

	il1 := NewInputLine(NewRect(0, 0, 30, 1), 100)
	il1.SetValidator(val1)
	il1.SetText("bad1")
	dlg.Insert(il1)

	il2 := NewInputLine(NewRect(0, 2, 30, 1), 100)
	il2.SetValidator(val2)
	il2.SetText("bad2")
	dlg.Insert(il2)

	dlg.Valid(CmOK)

	// First InputLine's Error should have been called.
	if !val1.errorCalled {
		t.Error("Valid(CmOK) did not call Error() on the first invalid InputLine's validator")
	}

	// Second InputLine's Error should NOT have been called (stops at first).
	if val2.errorCalled {
		t.Error("Valid(CmOK) called Error() on the second InputLine, want to stop at the first invalid")
	}
}

// ---------------------------------------------------------------------------
// Test 11: Mix of InputLines with and without validators
// Spec: "For each InputLine with a validator, calls IsValid(text).
//        (Skip InputLines without validators.)"
// ---------------------------------------------------------------------------

// TestDialogValidSkipsInputLinesWithoutValidators verifies that Valid() only
// validates InputLines that have validators set.
func TestDialogValidSkipsInputLinesWithoutValidators(t *testing.T) {
	dlg := NewDialog(NewRect(0, 0, 40, 15), "Validate")

	// Add first InputLine WITHOUT validator, with "invalid-looking" text.
	il1 := NewInputLine(NewRect(0, 0, 30, 1), 100)
	il1.SetText("") // No validator, so no check
	dlg.Insert(il1)

	// Add second InputLine WITH validator and valid text.
	il2 := NewInputLine(NewRect(0, 2, 30, 1), 100)
	il2.SetValidator(NewRangeValidator(1, 100))
	il2.SetText("50")
	dlg.Insert(il2)

	// Valid(CmOK) must return true (first InputLine has no validator to check).
	if !dlg.Valid(CmOK) {
		t.Error("Valid(CmOK) returned false when skipped InputLine has empty text but no validator")
	}
}

// ---------------------------------------------------------------------------
// Helper: mockValidator tracks calls to Error()
// ---------------------------------------------------------------------------

type mockValidator struct {
	isValidFn   func(string) bool
	errorCalled bool
}

func (m *mockValidator) IsValid(s string) bool {
	return m.isValidFn(s)
}

func (m *mockValidator) IsValidInput(s string, noAutoFill bool) bool {
	return m.isValidFn(s)
}

func (m *mockValidator) Error() {
	m.errorCalled = true
}
