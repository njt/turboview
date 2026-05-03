package tv

// history_colors_test.go — tests for Task 2: CmRecordHistory and InputLine.SelectAll.
//
// Every assertion cites the spec requirement it verifies.
// Each test covers exactly one behaviour.
//
// Spec requirements tested:
//   CmRecordHistory:
//     1. CmRecordHistory exists as a CommandCode constant.
//     2. CmRecordHistory is numerically greater than CmSelectWindowNum.
//     3. CmRecordHistory is less than CmUser (< 1000).
//     4. CmRecordHistory is distinct from every other named command constant.
//
//   InputLine.SelectAll:
//     5. On an empty InputLine: selStart=0, selEnd=0, cursorPos=0.
//     6. On a single-char InputLine: selStart=0, selEnd=1, cursorPos=1.
//     7. On a multi-char InputLine: selStart=0, selEnd=len(text), cursorPos=len(text).
//     8. SelectAll is idempotent — calling it twice gives the same result.
//     9. (falsifying) Before SelectAll is called, selection is not necessarily full.

import "testing"

// ---------------------------------------------------------------------------
// Requirement 1 — CmRecordHistory exists as a CommandCode
// Spec: "Add CmRecordHistory to the iota block in tv/command.go, after CmSelectWindowNum"
// Spec: "It must be a valid CommandCode constant"
// ---------------------------------------------------------------------------

// TestCmRecordHistoryIsCommandCode verifies CmRecordHistory can be assigned to
// a CommandCode variable, confirming it has the correct type.
func TestCmRecordHistoryIsCommandCode(t *testing.T) {
	var code CommandCode = CmRecordHistory
	if code != CmRecordHistory {
		t.Errorf("CmRecordHistory assigned to CommandCode = %d, does not round-trip", code)
	}
}

// ---------------------------------------------------------------------------
// Requirement 2 — CmRecordHistory comes after CmSelectWindowNum
// Spec: "Add CmRecordHistory to the iota block in tv/command.go, after CmSelectWindowNum"
// ---------------------------------------------------------------------------

// TestCmRecordHistoryAfterCmSelectWindowNum verifies CmRecordHistory is
// numerically greater than CmSelectWindowNum, as required by the iota ordering.
func TestCmRecordHistoryAfterCmSelectWindowNum(t *testing.T) {
	if CmRecordHistory <= CmSelectWindowNum {
		t.Errorf("CmRecordHistory (%d) must be greater than CmSelectWindowNum (%d)",
			CmRecordHistory, CmSelectWindowNum)
	}
}

// ---------------------------------------------------------------------------
// Requirement 3 — CmRecordHistory is before CmUser (< 1000)
// Spec: "It must be a valid CommandCode constant" (within the pre-CmUser block)
// ---------------------------------------------------------------------------

// TestCmRecordHistoryBeforeCmUser verifies CmRecordHistory is less than
// CmUser = 1000, placing it in the reserved system command range.
func TestCmRecordHistoryBeforeCmUser(t *testing.T) {
	if CmRecordHistory >= CmUser {
		t.Errorf("CmRecordHistory (%d) must be less than CmUser (%d)",
			CmRecordHistory, CmUser)
	}
}

// ---------------------------------------------------------------------------
// Requirement 4 — CmRecordHistory is distinct from all other command constants
// Spec: "It must be a valid CommandCode constant"
// A valid iota constant must have a unique numeric value.
// ---------------------------------------------------------------------------

// TestCmRecordHistoryIsDistinctFromOtherCommands verifies CmRecordHistory does
// not collide with any other named command constant.
func TestCmRecordHistoryIsDistinctFromOtherCommands(t *testing.T) {
	others := map[string]CommandCode{
		"CmQuit":            CmQuit,
		"CmClose":           CmClose,
		"CmOK":              CmOK,
		"CmCancel":          CmCancel,
		"CmYes":             CmYes,
		"CmNo":              CmNo,
		"CmMenu":            CmMenu,
		"CmResize":          CmResize,
		"CmZoom":            CmZoom,
		"CmTile":            CmTile,
		"CmCascade":         CmCascade,
		"CmNext":            CmNext,
		"CmPrev":            CmPrev,
		"CmDefault":         CmDefault,
		"CmGrabDefault":     CmGrabDefault,
		"CmReleaseDefault":  CmReleaseDefault,
		"CmReceivedFocus":   CmReceivedFocus,
		"CmReleasedFocus":   CmReleasedFocus,
		"CmScrollBarClicked": CmScrollBarClicked,
		"CmScrollBarChanged": CmScrollBarChanged,
		"CmSelectWindowNum": CmSelectWindowNum,
		"CmUser":            CmUser,
	}
	for name, val := range others {
		if CmRecordHistory == val {
			t.Errorf("CmRecordHistory (%d) collides with %s (%d)", CmRecordHistory, name, val)
		}
	}
}

// ---------------------------------------------------------------------------
// Requirement 5 — SelectAll on empty InputLine
// Spec: "Add SelectAll() method to InputLine that selects all text:
//         sets selStart = 0, selEnd = len(il.text), cursorPos = len(il.text)"
// ---------------------------------------------------------------------------

// TestSelectAllOnEmptyInputLine verifies that calling SelectAll on an InputLine
// with no text sets selStart=0, selEnd=0, cursorPos=0.
func TestSelectAllOnEmptyInputLine(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	// text is empty: len == 0

	il.SelectAll()

	selStart, selEnd := il.Selection()
	cursorPos := il.CursorPos()

	if selStart != 0 {
		t.Errorf("SelectAll on empty: selStart = %d, want 0", selStart)
	}
	if selEnd != 0 {
		t.Errorf("SelectAll on empty: selEnd = %d, want 0 (len of empty text)", selEnd)
	}
	if cursorPos != 0 {
		t.Errorf("SelectAll on empty: cursorPos = %d, want 0", cursorPos)
	}
}

// ---------------------------------------------------------------------------
// Requirement 6 — SelectAll on a single-character InputLine
// Spec: "sets selStart = 0, selEnd = len(il.text), cursorPos = len(il.text)"
// ---------------------------------------------------------------------------

// TestSelectAllOnSingleCharInputLine verifies SelectAll on a one-character
// InputLine sets selStart=0, selEnd=1, cursorPos=1.
func TestSelectAllOnSingleCharInputLine(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("x")

	il.SelectAll()

	selStart, selEnd := il.Selection()
	cursorPos := il.CursorPos()

	if selStart != 0 {
		t.Errorf("SelectAll single-char: selStart = %d, want 0", selStart)
	}
	if selEnd != 1 {
		t.Errorf("SelectAll single-char: selEnd = %d, want 1", selEnd)
	}
	if cursorPos != 1 {
		t.Errorf("SelectAll single-char: cursorPos = %d, want 1", cursorPos)
	}
}

// ---------------------------------------------------------------------------
// Requirement 7 — SelectAll on a multi-character InputLine
// Spec: "sets selStart = 0, selEnd = len(il.text), cursorPos = len(il.text)"
// ---------------------------------------------------------------------------

// TestSelectAllOnMultiCharInputLine verifies SelectAll on "hello" sets
// selStart=0, selEnd=5, cursorPos=5.
func TestSelectAllOnMultiCharInputLine(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")

	il.SelectAll()

	selStart, selEnd := il.Selection()
	cursorPos := il.CursorPos()
	wantLen := len("hello") // 5

	if selStart != 0 {
		t.Errorf("SelectAll multi-char: selStart = %d, want 0", selStart)
	}
	if selEnd != wantLen {
		t.Errorf("SelectAll multi-char: selEnd = %d, want %d", selEnd, wantLen)
	}
	if cursorPos != wantLen {
		t.Errorf("SelectAll multi-char: cursorPos = %d, want %d", cursorPos, wantLen)
	}
}

// ---------------------------------------------------------------------------
// Requirement 7 (extended) — selEnd equals exactly len(text)
// Spec: "sets selEnd = len(il.text)"
// This falsifying variant uses a longer string to ensure selEnd tracks text length.
// ---------------------------------------------------------------------------

// TestSelectAllSelEndEqualsTextLength verifies selEnd equals len(text) for
// a longer string, not a hardcoded value.
func TestSelectAllSelEndEqualsTextLength(t *testing.T) {
	text := "select all of this text"
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText(text)

	il.SelectAll()

	_, selEnd := il.Selection()
	want := len([]rune(text))

	if selEnd != want {
		t.Errorf("SelectAll: selEnd = %d, want len(text) = %d", selEnd, want)
	}
}

// ---------------------------------------------------------------------------
// Requirement 8 — SelectAll is idempotent
// Spec: "sets selStart = 0, selEnd = len(il.text), cursorPos = len(il.text)"
// Calling it twice must produce the same result as calling it once.
// ---------------------------------------------------------------------------

// TestSelectAllIsIdempotent verifies calling SelectAll twice leaves the
// selection in the same state as calling it once.
func TestSelectAllIsIdempotent(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")

	il.SelectAll()
	il.SelectAll()

	selStart, selEnd := il.Selection()
	cursorPos := il.CursorPos()

	if selStart != 0 {
		t.Errorf("SelectAll twice: selStart = %d, want 0", selStart)
	}
	if selEnd != 5 {
		t.Errorf("SelectAll twice: selEnd = %d, want 5", selEnd)
	}
	if cursorPos != 5 {
		t.Errorf("SelectAll twice: cursorPos = %d, want 5", cursorPos)
	}
}

// ---------------------------------------------------------------------------
// Requirement 9 (falsifying) — before SelectAll, selection is not full
// Spec: "This is needed by History's dropdown to select all text after pasting"
// (meaning the state before SelectAll is called must differ from after)
// ---------------------------------------------------------------------------

// TestSelectAllChangesStateFromDefault verifies that immediately after
// SetText (before SelectAll), the selection is not full — confirming SelectAll
// does meaningful work.
func TestSelectAllChangesStateFromDefault(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 20, 1), 0)
	il.SetText("hello")

	// After SetText, per InputLine.SetText: selStart=0, selEnd=0.
	// SelectAll must change selEnd from 0 to 5.
	selStart, selEndBefore := il.Selection()
	_ = selStart

	il.SelectAll()

	_, selEndAfter := il.Selection()

	if selEndBefore == selEndAfter {
		t.Errorf("SelectAll had no effect: selEnd was %d before and %d after — expected change from 0 to 5",
			selEndBefore, selEndAfter)
	}
}
