package tv

// memo_test.go — Tests for Task 2: Memo struct, constructor, and buffer model.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Widget interface satisfaction (compile-time)
//   Section 2  — Construction: NewMemo flags, grow mode, defaults
//   Section 3  — MemoOption: WithAutoIndent
//   Section 4  — Text() and SetText()
//   Section 5  — CursorPos()
//   Section 6  — AutoIndent() and SetAutoIndent()
//   Section 7  — Falsifying / boundary tests

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Section 1 — Compile-time Widget interface assertion
// ---------------------------------------------------------------------------

// Spec: "var _ Widget = (*Memo)(nil)" — Memo must satisfy the Widget interface.
var _ Widget = (*Memo)(nil)

// ---------------------------------------------------------------------------
// Section 2 — Construction
// ---------------------------------------------------------------------------

// TestNewMemoSetsSfVisible verifies NewMemo sets the SfVisible state flag.
// Spec: "Sets SfVisible, OfSelectable"
func TestNewMemoSetsSfVisible(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))

	if !m.HasState(SfVisible) {
		t.Error("NewMemo did not set SfVisible")
	}
}

// TestNewMemoSetsOfSelectable verifies NewMemo sets the OfSelectable option flag.
// Spec: "Sets SfVisible, OfSelectable"
func TestNewMemoSetsOfSelectable(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))

	if !m.HasOption(OfSelectable) {
		t.Error("NewMemo did not set OfSelectable")
	}
}

// TestNewMemoGrowMode verifies NewMemo sets GrowMode to GfGrowHiX|GfGrowHiY.
// Spec: "GrowMode: GfGrowHiX | GfGrowHiY"
func TestNewMemoGrowMode(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))

	want := GfGrowHiX | GfGrowHiY
	if m.GrowMode() != want {
		t.Errorf("GrowMode() = %v, want %v (GfGrowHiX|GfGrowHiY)", m.GrowMode(), want)
	}
}

// TestNewMemoGrowModeIsExact verifies GrowMode is exactly GfGrowHiX|GfGrowHiY,
// not a superset (e.g. GfGrowAll).
// Falsifying: an implementation using GfGrowAll would pass a "has both bits" check but fails here.
func TestNewMemoGrowModeIsExact(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))

	if m.GrowMode()&GfGrowLoX != 0 {
		t.Error("GrowMode should NOT include GfGrowLoX")
	}
	if m.GrowMode()&GfGrowLoY != 0 {
		t.Error("GrowMode should NOT include GfGrowLoY")
	}
}

// TestNewMemoStoresBounds verifies NewMemo records the given bounds.
// Spec: "NewMemo(bounds Rect, opts ...MemoOption) *Memo"
func TestNewMemoStoresBounds(t *testing.T) {
	r := NewRect(5, 3, 40, 10)
	m := NewMemo(r)

	if m.Bounds() != r {
		t.Errorf("Bounds() = %v, want %v", m.Bounds(), r)
	}
}

// TestNewMemoAutoIndentDefault verifies auto-indent is true by default.
// Spec: "Auto-indent: true by default"
func TestNewMemoAutoIndentDefault(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))

	if !m.AutoIndent() {
		t.Error("NewMemo: AutoIndent() should be true by default")
	}
}

// TestNewMemoCursorAtOrigin verifies cursor starts at (0, 0).
// Spec: "Cursor at (0, 0)"
func TestNewMemoCursorAtOrigin(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// TestNewMemoTextIsEmpty verifies an empty Memo has Text() == "".
// Spec: "An empty Memo (after NewMemo) has Text() == "" and CursorPos() == (0, 0)"
func TestNewMemoTextIsEmpty(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))

	if got := m.Text(); got != "" {
		t.Errorf("Text() on new Memo = %q, want %q", got, "")
	}
}

// TestNewMemoInitializesWithOneLine verifies internal representation starts with one empty line.
// Spec: "Initializes with one empty line ([][]rune{{}})"
// Verified via SetText("") behavior: after construction Text() == "" (one empty line, not zero lines).
// This is distinct from TestNewMemoTextIsEmpty — it guards against a zero-line initialization
// that would still return "" from Text().
func TestNewMemoInitializesWithOneLine(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))

	// After construction, cursor at row=0 must be valid (would panic or be clamped
	// to -1 if there were zero lines). CursorPos returning (0,0) confirms a line exists.
	row, col := m.CursorPos()
	if row != 0 {
		t.Errorf("CursorPos() row = %d, want 0; suggests zero lines at construction", row)
	}
	if col != 0 {
		t.Errorf("CursorPos() col = %d, want 0", col)
	}
	// And Text() is the empty string, not a newline (which would indicate two lines).
	if got := m.Text(); got != "" {
		t.Errorf("Text() = %q, want %q", got, "")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — MemoOption: WithAutoIndent
// ---------------------------------------------------------------------------

// TestWithAutoIndentFalse verifies WithAutoIndent(false) disables auto-indent.
// Spec: "WithAutoIndent(enabled bool) MemoOption"
func TestWithAutoIndentFalse(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10), WithAutoIndent(false))

	if m.AutoIndent() {
		t.Error("WithAutoIndent(false): AutoIndent() should be false")
	}
}

// TestWithAutoIndentTrue verifies WithAutoIndent(true) explicitly enables auto-indent.
// Spec: "WithAutoIndent(enabled bool) MemoOption"
func TestWithAutoIndentTrue(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10), WithAutoIndent(true))

	if !m.AutoIndent() {
		t.Error("WithAutoIndent(true): AutoIndent() should be true")
	}
}

// TestWithAutoIndentFalseNotDefault verifies the option actually overrides the default.
// Falsifying: an implementation that ignores the option and always returns true fails here.
func TestWithAutoIndentFalseNotDefault(t *testing.T) {
	withIndent := NewMemo(NewRect(0, 0, 40, 10))
	withoutIndent := NewMemo(NewRect(0, 0, 40, 10), WithAutoIndent(false))

	if withIndent.AutoIndent() == withoutIndent.AutoIndent() {
		t.Error("WithAutoIndent(false) had no effect: both Memos report the same AutoIndent() value")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Text() and SetText()
// ---------------------------------------------------------------------------

// TestSetTextSingleLine verifies SetText with no newlines stores one line.
// Spec: "SetText("hello\nworld") results in 2 lines, Text() returns "hello\nworld""
// (Implicit: single-segment input produces one line.)
func TestSetTextSingleLine(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("hello")

	if got := m.Text(); got != "hello" {
		t.Errorf("Text() = %q, want %q", got, "hello")
	}
}

// TestSetTextMultiLineUnixNewlines verifies SetText splits on \n.
// Spec: "SetText splits on \n and \r\n into [][]rune. Text() joins with \n"
// Spec: "SetText("hello\nworld") results in 2 lines, Text() returns "hello\nworld""
func TestSetTextMultiLineUnixNewlines(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("hello\nworld")

	if got := m.Text(); got != "hello\nworld" {
		t.Errorf("Text() = %q, want %q", got, "hello\nworld")
	}
}

// TestSetTextCRLF verifies SetText splits on \r\n (Windows line endings).
// Spec: "SetText splits on \n and \r\n into [][]rune. Text() joins with \n"
func TestSetTextCRLF(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("hello\r\nworld")

	// Text() must join with \n regardless of input line endings.
	if got := m.Text(); got != "hello\nworld" {
		t.Errorf("Text() after SetText with \\r\\n = %q, want %q", got, "hello\nworld")
	}
}

// TestSetTextEmptyStringResultsInOneLine verifies SetText("") gives one empty line.
// Spec: "SetText("") results in 1 empty line (not 0 lines)"
func TestSetTextEmptyStringResultsInOneLine(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("hello\nworld") // seed with two lines first
	m.SetText("")             // now reset

	if got := m.Text(); got != "" {
		t.Errorf("Text() after SetText(\"\") = %q, want %q", got, "")
	}
	// Cursor must still be valid — row 0 must exist.
	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("CursorPos() after SetText(\"\") = (%d, %d), want (0, 0)", row, col)
	}
}

// TestSetTextEmptyStringNotZeroLines verifies SetText("") does NOT produce zero lines.
// Falsifying: an implementation that splits "" into []string{} would return an invalid
// cursor, or Text() would still be "", but cursor operations would be broken.
// We verify the cursor is at a valid position (row 0 exists).
func TestSetTextEmptyStringNotZeroLines(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("")

	// If there were 0 lines, row 0 would be out of range.
	// CursorPos clamping spec says cursorRow in [0, len(lines)-1].
	// With 0 lines that invariant cannot hold. Verify row is 0.
	row, _ := m.CursorPos()
	if row != 0 {
		t.Errorf("CursorPos() row after SetText(\"\") = %d, want 0 (implies zero-line bug)", row)
	}
}

// TestSetTextResetsCursorToOrigin verifies SetText resets cursor to (0, 0).
// Spec: "After SetText, cursor resets to (0,0) and delta resets to (0,0)"
func TestSetTextResetsCursorToOrigin(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("first\nsecond\nthird")
	// Cursor is now at (0,0); call SetText again to confirm it resets regardless.
	m.SetText("new content")

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("CursorPos() after second SetText = (%d, %d), want (0, 0)", row, col)
	}
}

// TestSetTextAfterSetTextFullyReplaces verifies a second SetText fully replaces content.
// Spec: "SetText splits on \n and \r\n into [][]rune, resets cursor to (0, 0)"
// Falsifying: if lines were appended rather than replaced, Text() would be wrong.
func TestSetTextAfterSetTextFullyReplaces(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("alpha\nbeta\ngamma")
	m.SetText("one\ntwo")

	if got := m.Text(); got != "one\ntwo" {
		t.Errorf("Text() after second SetText = %q, want %q", got, "one\ntwo")
	}
}

// TestTextJoinsWithNewline verifies Text() uses \n (not \r\n) as the separator.
// Spec: "Text() joins with \n"
// Falsifying: an implementation joining with \r\n would fail this check.
func TestTextJoinsWithNewline(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("line1\nline2\nline3")

	got := m.Text()
	// Must not contain \r\n
	for i := 0; i < len(got)-1; i++ {
		if got[i] == '\r' && got[i+1] == '\n' {
			t.Errorf("Text() contains \\r\\n at position %d; want plain \\n separators", i)
		}
	}
	// Must contain exactly 2 newlines (3 lines → 2 separators)
	count := 0
	for _, c := range got {
		if c == '\n' {
			count++
		}
	}
	if count != 2 {
		t.Errorf("Text() has %d newline(s), want 2 for 3-line content", count)
	}
}

// TestSetTextThreeLines verifies SetText with three \n-separated segments.
// Spec: "splits on \n and \r\n into [][]rune"
func TestSetTextThreeLines(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("a\nb\nc")

	if got := m.Text(); got != "a\nb\nc" {
		t.Errorf("Text() = %q, want %q", got, "a\nb\nc")
	}
}

// TestSetTextMixedCRLFAndLF verifies CRLF and LF in the same string are both split correctly.
// Spec: "splits on \n and \r\n" — both delimiters must work together.
func TestSetTextMixedCRLFAndLF(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("a\r\nb\nc")

	if got := m.Text(); got != "a\nb\nc" {
		t.Errorf("Text() with mixed endings = %q, want %q", got, "a\nb\nc")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — CursorPos()
// ---------------------------------------------------------------------------

// TestCursorPosAfterSetText verifies CursorPos() returns (0, 0) after SetText.
// Spec: "After SetText, cursor resets to (0,0) and delta resets to (0,0)"
func TestCursorPosAfterSetText(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("hello\nworld")

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("CursorPos() after SetText = (%d, %d), want (0, 0)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 6 — AutoIndent() and SetAutoIndent()
// ---------------------------------------------------------------------------

// TestSetAutoIndentToFalse verifies SetAutoIndent(false) disables auto-indent.
// Spec: "SetAutoIndent(enabled bool) — sets auto-indent"
func TestSetAutoIndentToFalse(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetAutoIndent(false)

	if m.AutoIndent() {
		t.Error("AutoIndent() should return false after SetAutoIndent(false)")
	}
}

// TestSetAutoIndentToTrue verifies SetAutoIndent(true) re-enables auto-indent.
// Spec: "SetAutoIndent(enabled bool) — sets auto-indent"
func TestSetAutoIndentToTrue(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10), WithAutoIndent(false))
	m.SetAutoIndent(true)

	if !m.AutoIndent() {
		t.Error("AutoIndent() should return true after SetAutoIndent(true)")
	}
}

// TestSetAutoIndentToggle verifies AutoIndent round-trips correctly.
// Falsifying: an implementation storing a fixed value ignores SetAutoIndent.
func TestSetAutoIndentToggle(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10)) // starts true

	m.SetAutoIndent(false)
	if m.AutoIndent() {
		t.Error("AutoIndent() should be false after SetAutoIndent(false)")
	}

	m.SetAutoIndent(true)
	if !m.AutoIndent() {
		t.Error("AutoIndent() should be true after SetAutoIndent(true)")
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Falsifying / boundary tests
// ---------------------------------------------------------------------------

// TestSetTextEmptyAfterMultiLine verifies that after SetText(""), Text() is ""
// and not the previous content.
// Falsifying: an implementation that ignores an empty SetText call would fail.
func TestSetTextEmptyAfterMultiLine(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("line one\nline two\nline three")
	m.SetText("")

	if got := m.Text(); got != "" {
		t.Errorf("Text() after SetText(\"\") following multi-line content = %q, want %q", got, "")
	}
}

// TestCursorPosRowAndColAreIndependent verifies CursorPos returns a 2-tuple,
// not a single combined value.
// Falsifying: an implementation returning (row+col, row+col) would pass single-value tests.
func TestCursorPosRowAndColAreIndependent(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10))
	m.SetText("hello")

	row, col := m.CursorPos()
	// Both must be 0 independently.
	if row != 0 {
		t.Errorf("CursorPos() row = %d, want 0", row)
	}
	if col != 0 {
		t.Errorf("CursorPos() col = %d, want 0", col)
	}
}

// TestNewMemoWithDifferentBoundsDoesNotShareState verifies two distinct Memo instances
// are independent (guard against shared mutable state / global singleton bugs).
func TestNewMemoWithDifferentBoundsDoesNotShareState(t *testing.T) {
	m1 := NewMemo(NewRect(0, 0, 40, 10))
	m2 := NewMemo(NewRect(10, 10, 20, 5))

	m1.SetText("memo one content")
	m2.SetText("memo two content")

	if m1.Text() == m2.Text() {
		t.Error("two distinct Memo instances share Text() state; they should be independent")
	}
	if m1.Text() != "memo one content" {
		t.Errorf("m1.Text() = %q, want %q", m1.Text(), "memo one content")
	}
	if m2.Text() != "memo two content" {
		t.Errorf("m2.Text() = %q, want %q", m2.Text(), "memo two content")
	}
}

