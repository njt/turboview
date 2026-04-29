package tv

import (
	"testing"
)

// TestPhase9CommandNewConstantsExist verifies that all five new command constants exist
// per the spec requirement: "CmDefault, CmGrabDefault, CmReleaseDefault, CmReceivedFocus,
// CmReleasedFocus constants exist"
func TestPhase9CommandNewConstantsExist(t *testing.T) {
	// If any constant doesn't exist, this won't compile, but we also verify they're non-zero
	if CmDefault == 0 {
		t.Error("CmDefault should have a non-zero value")
	}
	if CmGrabDefault == 0 {
		t.Error("CmGrabDefault should have a non-zero value")
	}
	if CmReleaseDefault == 0 {
		t.Error("CmReleaseDefault should have a non-zero value")
	}
	if CmReceivedFocus == 0 {
		t.Error("CmReceivedFocus should have a non-zero value")
	}
	if CmReleasedFocus == 0 {
		t.Error("CmReleasedFocus should have a non-zero value")
	}
}

// TestPhase9CommandNewConstantsDistinct verifies all new constants have distinct values
// per the spec requirement: "All new constants have distinct values"
func TestPhase9CommandNewConstantsDistinct(t *testing.T) {
	newConstants := map[string]CommandCode{
		"CmDefault":        CmDefault,
		"CmGrabDefault":    CmGrabDefault,
		"CmReleaseDefault": CmReleaseDefault,
		"CmReceivedFocus":  CmReceivedFocus,
		"CmReleasedFocus":  CmReleasedFocus,
	}

	seen := make(map[CommandCode]string)
	for name, val := range newConstants {
		if prev, ok := seen[val]; ok {
			t.Errorf("Duplicate value: %s and %s both have value %d", prev, name, val)
		}
		seen[val] = name
	}
}

// TestPhase9CommandNewConstantsDistinctFromExisting verifies new constants don't overlap
// with existing constants per the spec requirement: "All new constants are distinct
// from existing constants"
func TestPhase9CommandNewConstantsDistinctFromExisting(t *testing.T) {
	existingConstants := map[string]CommandCode{
		"CmQuit":    CmQuit,
		"CmClose":   CmClose,
		"CmOK":      CmOK,
		"CmCancel":  CmCancel,
		"CmYes":     CmYes,
		"CmNo":      CmNo,
		"CmMenu":    CmMenu,
		"CmResize":  CmResize,
		"CmZoom":    CmZoom,
		"CmTile":    CmTile,
		"CmCascade": CmCascade,
		"CmNext":    CmNext,
		"CmPrev":    CmPrev,
		"CmUser":    CmUser,
	}

	newConstants := []CommandCode{
		CmDefault, CmGrabDefault, CmReleaseDefault, CmReceivedFocus, CmReleasedFocus,
	}

	for _, newVal := range newConstants {
		for existingName, existingVal := range existingConstants {
			if newVal == existingVal {
				t.Errorf("New constant has same value as existing %s: %d", existingName, existingVal)
			}
		}
	}
}

// TestPhase9CommandExistingConstantsRetainValues verifies the spec requirement:
// "All existing constants (CmQuit through CmUser) retain their current values"
func TestPhase9CommandExistingConstantsRetainValues(t *testing.T) {
	expected := map[string]CommandCode{
		"CmQuit":    1,
		"CmClose":   2,
		"CmOK":      3,
		"CmCancel":  4,
		"CmYes":     5,
		"CmNo":      6,
		"CmMenu":    7,
		"CmResize":  8,
		"CmZoom":    9,
		"CmTile":    10,
		"CmCascade": 11,
		"CmNext":    12,
		"CmPrev":    13,
		"CmUser":    1000,
	}

	actual := map[string]CommandCode{
		"CmQuit":    CmQuit,
		"CmClose":   CmClose,
		"CmOK":      CmOK,
		"CmCancel":  CmCancel,
		"CmYes":     CmYes,
		"CmNo":      CmNo,
		"CmMenu":    CmMenu,
		"CmResize":  CmResize,
		"CmZoom":    CmZoom,
		"CmTile":    CmTile,
		"CmCascade": CmCascade,
		"CmNext":    CmNext,
		"CmPrev":    CmPrev,
		"CmUser":    CmUser,
	}

	for name, expectedVal := range expected {
		actualVal := actual[name]
		if actualVal != expectedVal {
			t.Errorf("%s = %d, want %d", name, actualVal, expectedVal)
		}
	}
}

// TestPhase9CommandDefaultAfterExisting verifies the spec requirement:
// "CmDefault constant exists with a unique value after the existing constants"
// The last existing sequential constant is CmPrev (13), and CmUser (1000) is special.
// CmDefault should be > 13 and != 1000.
func TestPhase9CommandDefaultAfterExisting(t *testing.T) {
	if CmDefault <= CmPrev {
		t.Errorf("CmDefault (%d) should be after CmPrev (%d)", CmDefault, CmPrev)
	}
	if CmDefault == CmUser {
		t.Errorf("CmDefault (%d) should not equal CmUser (%d)", CmDefault, CmUser)
	}
}

// TestPhase9CommandNewConstantsAreCommandCodeType verifies the new constants
// are properly typed as CommandCode
func TestPhase9CommandNewConstantsAreCommandCodeType(t *testing.T) {
	// These assignments verify type compatibility at compile time
	var _ CommandCode = CmDefault
	var _ CommandCode = CmGrabDefault
	var _ CommandCode = CmReleaseDefault
	var _ CommandCode = CmReceivedFocus
	var _ CommandCode = CmReleasedFocus

	// Verify they can participate in integer operations as CommandCode
	if CmDefault+1 <= CmDefault {
		t.Error("CmDefault arithmetic failed")
	}
}

// TestPhase9CommandAllConstantsUnique verifies no collisions across all constants
// (existing + new) per the spec requirement about distinct values
func TestPhase9CommandAllConstantsUnique(t *testing.T) {
	allConstants := map[string]CommandCode{
		// Existing constants
		"CmQuit":    CmQuit,
		"CmClose":   CmClose,
		"CmOK":      CmOK,
		"CmCancel":  CmCancel,
		"CmYes":     CmYes,
		"CmNo":      CmNo,
		"CmMenu":    CmMenu,
		"CmResize":  CmResize,
		"CmZoom":    CmZoom,
		"CmTile":    CmTile,
		"CmCascade": CmCascade,
		"CmNext":    CmNext,
		"CmPrev":    CmPrev,
		"CmUser":    CmUser,
		// New constants
		"CmDefault":        CmDefault,
		"CmGrabDefault":    CmGrabDefault,
		"CmReleaseDefault": CmReleaseDefault,
		"CmReceivedFocus":  CmReceivedFocus,
		"CmReleasedFocus":  CmReleasedFocus,
	}

	seen := make(map[CommandCode]string)
	for name, val := range allConstants {
		if prev, ok := seen[val]; ok {
			t.Errorf("Duplicate value: %s and %s both have value %d", prev, name, val)
		}
		seen[val] = name
	}
}

// TestPhase9CommandDefaultIsPositive verifies CmDefault is positive
// (as required by the spec: "non-zero value")
func TestPhase9CommandDefaultIsPositive(t *testing.T) {
	if CmDefault <= 0 {
		t.Errorf("CmDefault = %d, should be positive", CmDefault)
	}
}

// TestPhase9CommandGrabDefaultIsPositive verifies CmGrabDefault is positive
func TestPhase9CommandGrabDefaultIsPositive(t *testing.T) {
	if CmGrabDefault <= 0 {
		t.Errorf("CmGrabDefault = %d, should be positive", CmGrabDefault)
	}
}

// TestPhase9CommandReleaseDefaultIsPositive verifies CmReleaseDefault is positive
func TestPhase9CommandReleaseDefaultIsPositive(t *testing.T) {
	if CmReleaseDefault <= 0 {
		t.Errorf("CmReleaseDefault = %d, should be positive", CmReleaseDefault)
	}
}

// TestPhase9CommandReceivedFocusIsPositive verifies CmReceivedFocus is positive
func TestPhase9CommandReceivedFocusIsPositive(t *testing.T) {
	if CmReceivedFocus <= 0 {
		t.Errorf("CmReceivedFocus = %d, should be positive", CmReceivedFocus)
	}
}

// TestPhase9CommandReleasedFocusIsPositive verifies CmReleasedFocus is positive
func TestPhase9CommandReleasedFocusIsPositive(t *testing.T) {
	if CmReleasedFocus <= 0 {
		t.Errorf("CmReleasedFocus = %d, should be positive", CmReleasedFocus)
	}
}
