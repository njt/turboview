package tv

import (
	"testing"
)

// TestCommandCodeConstants verifies the spec requirement:
// "CommandCode constants: CmQuit through CmPrev sequential, CmUser = 1000"
// The spec shows:
// const (
//     CmQuit    CommandCode = iota + 1
//     CmClose
//     CmOK
//     CmCancel
//     CmYes
//     CmNo
//     CmMenu
//     CmResize
//     CmZoom
//     CmTile
//     CmCascade
//     CmNext
//     CmPrev
//     CmUser    CommandCode = 1000
// )
func TestCommandCodeConstantsSequential(t *testing.T) {
	// CmQuit should be 1 (iota + 1 starting from 0)
	if CmQuit != 1 {
		t.Errorf("CmQuit = %d, want 1", CmQuit)
	}
	// CmClose should be 2
	if CmClose != 2 {
		t.Errorf("CmClose = %d, want 2", CmClose)
	}
}

// TestCommandCodeAllSequential verifies all commands through CmPrev are sequential
func TestCommandCodeAllSequential(t *testing.T) {
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
	}

	for name, expectedVal := range expected {
		actualVal := actual[name]
		if actualVal != expectedVal {
			t.Errorf("%s = %d, want %d", name, actualVal, expectedVal)
		}
	}
}

// TestCommandCodeUserValue verifies: "CmUser = 1000"
func TestCommandCodeUserValue(t *testing.T) {
	if CmUser != 1000 {
		t.Errorf("CmUser = %d, want 1000", CmUser)
	}
}

// TestCommandCodeUserGreaterThanPrev verifies CmUser is separate from the sequential block
func TestCommandCodeUserGreaterThanPrev(t *testing.T) {
	if CmUser <= CmPrev {
		t.Errorf("CmUser (%d) should be greater than CmPrev (%d)", CmUser, CmPrev)
	}
}

// TestCommandCodeNoGaps verifies sequential values have no gaps
func TestCommandCodeNoGaps(t *testing.T) {
	commands := []CommandCode{
		CmQuit, CmClose, CmOK, CmCancel, CmYes, CmNo, CmMenu,
		CmResize, CmZoom, CmTile, CmCascade, CmNext, CmPrev,
	}

	for i := 0; i < len(commands)-1; i++ {
		if commands[i+1]-commands[i] != 1 {
			t.Errorf("Gap between command %d and %d: values are %d and %d",
				i, i+1, commands[i], commands[i+1])
		}
	}
}

// TestCommandCodeDistinctValues verifies all constants are unique
func TestCommandCodeDistinctValues(t *testing.T) {
	seen := make(map[CommandCode]string)
	constants := map[string]CommandCode{
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

	for name, val := range constants {
		if prev, ok := seen[val]; ok {
			t.Errorf("Duplicate value: %s and %s both have value %d", prev, name, val)
		}
		seen[val] = name
	}
}

// TestCommandCodeIntegerType verifies CommandCode is an integer type
func TestCommandCodeIntegerType(t *testing.T) {
	// This test verifies we can perform integer operations on CommandCode
	var c CommandCode = CmQuit
	_ = c + 1 // Should compile and work
	if c+1 != 2 {
		t.Errorf("CmQuit + 1 = %d, want 2", c+1)
	}
}

// TestCommandCodeComparable verifies CommandCode values can be compared
func TestCommandCodeComparable(t *testing.T) {
	if !(CmQuit < CmClose) {
		t.Error("CmQuit < CmClose failed")
	}
	if !(CmPrev < CmUser) {
		t.Error("CmPrev < CmUser failed")
	}
	if CmUser == CmPrev {
		t.Error("CmUser should not equal CmPrev")
	}
}
