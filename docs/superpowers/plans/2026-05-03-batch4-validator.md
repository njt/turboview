# Batch 4: Input Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a validator framework to InputLine that rejects invalid input on keystroke and on blur, with four concrete validators: FilterValidator, RangeValidator, PictureValidator, StringLookupValidator.

**Architecture:** Validator is a Go interface. InputLine gains an optional validator field. On every text mutation, InputLine saves state, mutates, validates, and rolls back if rejected. On blur (SfSelected cleared), InputLine calls IsValid and blocks focus change if invalid. Dialog.Valid() checks all child InputLines.

**Tech Stack:** Go, tcell/v2, existing tv package

**Spec:** `docs/superpowers/specs/2026-05-03-remaining-widgets-design.md` sections 4.1-4.9

---

## File Structure

| File | Responsibility |
|------|---------------|
| `tv/validator.go` (create) | Validator interface, FilterValidator, RangeValidator, StringLookupValidator |
| `tv/picture_validator.go` (create) | PictureValidator (complex enough for its own file) |
| `tv/input_line.go` (modify) | Add validator field, save/restore state, blur validation hook |
| `tv/dialog.go` (modify) | Add Valid() method that checks all InputLine validators |
| `e2e/testapp/basic/main.go` (modify) | Add validated InputLine to demo |
| `e2e/e2e_test.go` (modify) | E2e tests for validation behavior |

---

- [ ] ### Task 1: Validator Interface + FilterValidator

**Files:**
- Create: `tv/validator.go`

**Requirements:**
- A `Validator` interface exists with three methods: `IsValid(s string) bool`, `IsValidInput(s string, noAutoFill bool) bool`, `Error()`
- `NewFilterValidator(validChars string)` returns a `*FilterValidator` that implements `Validator`
- `FilterValidator.IsValidInput("abc", false)` returns true when validChars contains 'a', 'b', 'c'
- `FilterValidator.IsValidInput("abc", false)` returns false when validChars is "ab" (missing 'c')
- `FilterValidator.IsValidInput("", false)` returns true (empty string is valid partial input)
- `FilterValidator.IsValid("abc")` returns true when all chars are in validChars
- `FilterValidator.IsValid("")` returns true (empty is valid — FilterValidator only checks character membership)
- `FilterValidator.Error()` is callable (in production it shows a message box; for unit testing, it should not panic when no Application is running)
- Rune-level matching: `FilterValidator` with validChars "äöü" accepts "ä" input

**Implementation:**

```go
package tv

import "fmt"

// Validator defines the two-phase validation interface for InputLine.
type Validator interface {
	IsValid(s string) bool
	IsValidInput(s string, noAutoFill bool) bool
	Error()
}

// FilterValidator accepts only characters from a whitelist.
type FilterValidator struct {
	validChars map[rune]bool
	chars      string
}

func NewFilterValidator(validChars string) *FilterValidator {
	m := make(map[rune]bool)
	for _, r := range validChars {
		m[r] = true
	}
	return &FilterValidator{validChars: m, chars: validChars}
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
	// TODO: show message box "Invalid character in input" when Application exists
	_ = fmt.Sprintf("Invalid character in input")
}
```

**Run tests:** `go test ./tv/ -run TestFilter -v`

**Commit:** `git commit -m "feat: add Validator interface and FilterValidator"`

---

- [ ] ### Task 2: RangeValidator

**Files:**
- Modify: `tv/validator.go`

**Requirements:**
- `NewRangeValidator(min, max int)` returns a `*RangeValidator` that implements `Validator`
- RangeValidator embeds FilterValidator behavior: keystroke-level character filtering
- When `min >= 0`, only `"+0123456789"` are accepted per-keystroke (no minus sign)
- When `min < 0`, `"+-0123456789"` are accepted per-keystroke
- `RangeValidator.IsValidInput("", false)` returns true (empty is valid partial input during typing)
- `RangeValidator.IsValidInput("+", false)` returns true (lone sign is valid partial input)
- `RangeValidator.IsValidInput("-", false)` returns true only when min < 0
- `RangeValidator.IsValidInput("abc", false)` returns false
- `RangeValidator.IsValid("42")` returns true when 42 is within [min, max]
- `RangeValidator.IsValid("42")` returns false when 42 is outside [min, max]
- `RangeValidator.IsValid("")` returns false (empty string is not a valid committed value)
- `RangeValidator.IsValid("+")` returns false (sign alone is not a valid number)
- `RangeValidator.IsValid("99999999999999")` returns false (overflow beyond int range)
- `RangeValidator.Error()` is callable

**Implementation:**

```go
type RangeValidator struct {
	FilterValidator
	min, max int
}

func NewRangeValidator(min, max int) *RangeValidator {
	chars := "+-0123456789"
	if min >= 0 {
		chars = "+0123456789"
	}
	rv := &RangeValidator{min: min, max: max}
	rv.FilterValidator = *NewFilterValidator(chars)
	return rv
}

func (r *RangeValidator) IsValidInput(s string, noAutoFill bool) bool {
	if s == "" || s == "+" || s == "-" {
		return true
	}
	return r.FilterValidator.IsValidInput(s, noAutoFill)
}

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

func (r *RangeValidator) Error() {
	_ = fmt.Sprintf("Value not in the range %d to %d", r.min, r.max)
}
```

**Run tests:** `go test ./tv/ -run TestRange -v`

**Commit:** `git commit -m "feat: add RangeValidator with min/max checking"`

---

- [ ] ### Task 3: StringLookupValidator

**Files:**
- Modify: `tv/validator.go`

**Requirements:**
- `NewStringLookupValidator(items []string)` returns a `*StringLookupValidator` implementing `Validator`
- `StringLookupValidator.IsValidInput` always returns true (any partial input is acceptable)
- `StringLookupValidator.IsValid("foo")` returns true when "foo" is in the items list
- `StringLookupValidator.IsValid("foo")` returns false when "foo" is not in the items list
- Matching is case-sensitive: `IsValid("Foo")` returns false when only "foo" is in the list
- `StringLookupValidator.IsValid("")` returns false when "" is not in the list
- `StringLookupValidator.Error()` is callable

**Implementation:**

```go
type StringLookupValidator struct {
	items map[string]bool
}

func NewStringLookupValidator(items []string) *StringLookupValidator {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item] = true
	}
	return &StringLookupValidator{items: m}
}

func (v *StringLookupValidator) IsValidInput(s string, noAutoFill bool) bool {
	return true
}

func (v *StringLookupValidator) IsValid(s string) bool {
	return v.items[s]
}

func (v *StringLookupValidator) Error() {
	_ = fmt.Sprintf("Input not in valid list")
}
```

**Run tests:** `go test ./tv/ -run TestStringLookup -v`

**Commit:** `git commit -m "feat: add StringLookupValidator"`

---

- [ ] ### Task 4: PictureValidator

**Files:**
- Create: `tv/picture_validator.go`

**Requirements:**
- `NewPictureValidator(pic string, autoFill bool)` returns a `*PictureValidator` implementing `Validator`
- Picture codes: `#` (digit), `?` (letter), `&` (letter, auto-uppercase), `!` (any, auto-uppercase), `@` (any)
- `PictureValidator.IsValidInput("1", false)` returns true for picture `"###"`
- `PictureValidator.IsValidInput("12", false)` returns true for picture `"###"` (partial match)
- `PictureValidator.IsValidInput("123", false)` returns true for picture `"###"` (full match)
- `PictureValidator.IsValidInput("1234", false)` returns false for picture `"###"` (exceeds pattern)
- `PictureValidator.IsValidInput("a", false)` returns false for picture `"###"` (letter not digit)
- `PictureValidator.IsValidInput("a", false)` returns true for picture `"???"` (letter matches `?`)
- `PictureValidator.IsValid("123")` returns true for picture `"###"` (fully matches)
- `PictureValidator.IsValid("12")` returns false for picture `"###"` (incomplete)
- Auto-uppercase: for picture `"&&&"`, IsValidInput returns true for "abc" and the input string is uppercased to "ABC" (when noAutoFill is false). The validator may modify the input string passed by pointer or return the modified string.
- Escape: `;` escapes the next character as literal. Picture `";#"` requires a literal `#` character, not a digit.
- Optional groups: `[...]` means the contents are optional. Picture `"###[#]"` accepts both "123" and "1234" as valid.
- Required groups: `{...}` means the contents are required. Picture `"#{#}#"` requires the middle digit.
- Alternation: `,` tries different sub-patterns. Not implemented in this phase — return `prError` for now and document as TODO.
- Repetition: `*` followed by optional count. Not implemented in this phase — return `prError` for now and document as TODO.
- Auto-fill: when enabled and noAutoFill is false, literal characters in the picture are automatically inserted. Picture `"(###) ###-####"` with input "123" auto-fills to "(123) " (inserting parentheses, space, and awaiting more digits).
- `PictureValidator.Error()` is callable

**Implementation notes:**

The PictureValidator walks the picture string and input string in parallel using a recursive descent parser. State machine has positions in both strings. The key function is `scan(pic, picIdx, input, inIdx) (result, picEnd, inEnd)` where result indicates match quality.

This is the most complex validator. The core algorithm: advance through picture positions. For each position, check if the current input character matches the pattern at this picture position. Literal characters in the picture must match exactly (or be auto-filled). Pattern characters (`#`, `?`, `&`, `!`, `@`) consume one input character if it matches.

For auto-uppercase (`&`, `!`): when a match succeeds, replace the input character with its uppercase form.

For auto-fill: when the picture position is a literal and the input position doesn't have a matching character, insert the literal into the input string.

```go
package tv

type PicResult int

const (
	PrComplete PicResult = iota
	PrIncomplete
	PrEmpty
	PrError
)

type PictureValidator struct {
	pic      string
	autoFill bool
}

func NewPictureValidator(pic string, autoFill bool) *PictureValidator {
	return &PictureValidator{pic: pic, autoFill: autoFill}
}
```

The `scan` function is recursive: `[...]` groups call scan recursively for the bracketed content and treat failure as acceptable (optional). `{...}` groups call scan recursively and require success. `;` consumes the next picture character as literal.

**Run tests:** `go test ./tv/ -run TestPicture -v`

**Commit:** `git commit -m "feat: add PictureValidator with pattern matching and auto-fill"`

---

- [ ] ### Task 5: InputLine Validation Integration

**Files:**
- Modify: `tv/input_line.go`

**Requirements:**
- `InputLine.SetValidator(v Validator)` stores a validator on the InputLine
- `InputLine.Validator()` returns the stored validator (nil if none)
- **Keystroke rejection:** When a validator is set and the user types a character, the InputLine saves its state (text, cursorPos, selStart, selEnd, scrollOffset) before mutation. After mutation, it calls `validator.IsValidInput(newText, false)`. If false, the entire state is rolled back to the snapshot. The rejected keystroke is still cleared (consumed by InputLine).
- **Keystroke rejection applies to all mutations:** character insert, backspace, delete, Ctrl+Backspace (word delete), Ctrl+Delete (word delete), Ctrl+V (paste), Ctrl+X (cut), Ctrl+Y (clear all)
- **Blur validation:** InputLine overrides `SetState`. When `SfSelected` transitions from true to false, and a validator is set, it calls `validator.IsValid(text)`. If false, it calls `validator.Error()` and re-focuses itself by calling `owner.SetFocusedChild(self)`.
- **PictureValidator auto-fill integration:** When `IsValidInput` modifies the input string (auto-fill), InputLine accepts the modified text. This requires IsValidInput to take a `*string` parameter OR the InputLine re-reads the text after validation. **Design choice:** Use a `ValidatorWithAutoFill` interface that has `IsValidInputAutoFill(s *string, noAutoFill bool) bool` — if the validator implements this, InputLine calls it and accepts the modified string. Otherwise, calls `IsValidInput` normally.
- State snapshot/restore must be internal — not exposed publicly
- When no validator is set, InputLine behaves exactly as before (no regression)

**Implementation:**

Add to InputLine struct:
```go
type InputLine struct {
	// ... existing fields ...
	validator Validator
}

func (il *InputLine) SetValidator(v Validator) { il.validator = v }
func (il *InputLine) Validator() Validator     { return il.validator }
```

Add state save/restore:
```go
type inputLineState struct {
	text         []rune
	cursorPos    int
	scrollOffset int
	selStart     int
	selEnd       int
}

func (il *InputLine) saveState() inputLineState {
	textCopy := make([]rune, len(il.text))
	copy(textCopy, il.text)
	return inputLineState{
		text: textCopy, cursorPos: il.cursorPos,
		scrollOffset: il.scrollOffset, selStart: il.selStart, selEnd: il.selEnd,
	}
}

func (il *InputLine) restoreState(s inputLineState) {
	il.text = s.text
	il.cursorPos = s.cursorPos
	il.scrollOffset = s.scrollOffset
	il.selStart = s.selStart
	il.selEnd = s.selEnd
}
```

Wrap every text-mutating operation in HandleEvent:
```go
// Before mutation:
var saved inputLineState
if il.validator != nil {
	saved = il.saveState()
}

// ... existing mutation code ...

// After mutation, before event.Clear():
if il.validator != nil {
	newText := string(il.text)
	if !il.validator.IsValidInput(newText, false) {
		il.restoreState(saved)
	}
}
```

Add SetState override for blur validation:
```go
func (il *InputLine) SetState(flag ViewState, on bool) {
	wasSelected := il.HasState(SfSelected)
	il.BaseView.SetState(flag, on)
	if flag == SfSelected && wasSelected && !on {
		if il.validator != nil && !il.validator.IsValid(string(il.text)) {
			il.validator.Error()
			if owner := il.Owner(); owner != nil {
				owner.SetFocusedChild(il.Self())
			}
		}
	}
}
```

**Run tests:** `go test ./tv/ -run TestInputLineValid -v`

**Commit:** `git commit -m "feat: integrate validator into InputLine with save/restore and blur validation"`

---

- [ ] ### Task 6: Dialog.Valid() for Commit Validation

**Files:**
- Modify: `tv/dialog.go`

**Requirements:**
- `Dialog` has a `Valid(cmd CommandCode) bool` method
- `Valid(CmCancel)` always returns true (skip validation on cancel)
- `Valid(CmOK)` iterates all children recursively. For each InputLine with a validator, calls `validator.IsValid(text)`. If any fails, focuses that InputLine, calls `validator.Error()`, and returns false.
- When `Valid()` returns false, the dialog does NOT close (modal loop continues)
- The dialog's existing event handling for CmOK/CmCancel must call `Valid()` before closing
- If no children have validators, `Valid()` returns true (no regression)

**Implementation:**

Add to dialog.go:
```go
func (d *Dialog) Valid(cmd CommandCode) bool {
	if cmd == CmCancel {
		return true
	}
	for _, child := range d.group.Children() {
		if il, ok := child.(*InputLine); ok {
			if il.Validator() != nil && !il.Validator().IsValid(il.Text()) {
				d.group.SetFocusedChild(il)
				il.Validator().Error()
				return false
			}
		}
	}
	return true
}
```

**Critical: Intercept CmOK in Dialog.HandleEvent.** The modal termination happens in `Group.ExecView` (group.go:194-199) which checks `event.Command` AFTER `HandleEvent` returns. The Dialog must intercept CmOK before it reaches ExecView. Add this at the END of `Dialog.HandleEvent`, after the existing `CmClose → CmCancel` rewrite (around line 287):

```go
// Validate before allowing modal termination
if !event.IsCleared() && event.What == EvCommand && event.Command == CmOK {
	if !d.Valid(CmOK) {
		event.Clear() // prevent ExecView from seeing CmOK
	}
}
```

This must go after the group has processed the event (so button presses have already transformed the event into CmOK) but before HandleEvent returns (so ExecView doesn't see it if validation fails).

**Run tests:** `go test ./tv/ -run TestDialogValid -v`

**Commit:** `git commit -m "feat: add Dialog.Valid() for commit-time validator checking"`

---

- [ ] ### Task 7: Integration Checkpoint — Validator Pipeline

**Purpose:** Verify that validators work end-to-end through InputLine and Dialog.

**Requirements:**
- Create an InputLine with a RangeValidator(1, 100). Type "50" — text is "50". Type "abc" — text stays "50" (rejected).
- Create an InputLine with a RangeValidator(1, 100). Set text to "200". Call IsValid — returns false. 
- Create a Dialog with an InputLine that has a RangeValidator. Simulate CmOK — the dialog's Valid() method returns false when text is "200". Returns true when text is "50".
- Create an InputLine with FilterValidator("abcdef"). Type "abc" — accepted. Type "z" after — rejected, text stays "abc".
- Create an InputLine with StringLookupValidator(["apple","banana"]). Type "apple" freely (IsValidInput always true). IsValid("apple") returns true. IsValid("pear") returns false.
- PictureValidator with "###": InputLine accepts "1", "12", "123" but rejects "a" and "1234".

**Components to wire up:** InputLine, Dialog, all four validator types (real, no mocks)

**Run:** `go test ./tv/ -run TestValidatorIntegration -v`

---

- [ ] ### Task 8: Demo App + E2e Tests

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

**Requirements:**

**Demo app changes:**
- Add a "Port:" Label + InputLine to Window 1 "File Manager" dialog, positioned below the existing controls
- The Port InputLine has `NewRangeValidator(1, 65535)` set
- The Port InputLine is 6 characters wide (enough for "65535")
- The Port Label uses tilde shortcut: `~P~ort:`
- The existing filename InputLine (linked to the History widget) gets a `NewFilterValidator` that accepts all printable ASCII except null bytes. Use a validChars string containing all printable runes (space through tilde, 0x20-0x7E, plus common extended chars). This matches spec section 4.9.

**E2e test requirements:**
- Build the binary with `go build -o basic ./e2e/testapp/basic`
- Launch in tmux, navigate to the validated InputLine
- Type valid digits (e.g., "8080") — verify they appear in the InputLine
- Type invalid characters (e.g., letters) — verify they do NOT appear (text stays as the valid digits)
- All existing e2e tests continue to pass (regression check)

**Implementation guidance for demo app:**

In the File Manager window setup, after the existing History widget, add:
```go
portLabel := tv.NewLabel(tv.NewRect(1, 9, 8, 1), "~P~ort:", portInput)
portInput := tv.NewInputLine(tv.NewRect(9, 9, 8, 1), 5)
portInput.SetValidator(tv.NewRangeValidator(1, 65535))
// Insert portLabel and portInput into the dialog group
```

Adjust positions as needed to fit the existing layout.

**Run tests:** `go test ./e2e/ -timeout 180s -v`

**Commit:** `git commit -m "feat: add validated Port input to demo app with e2e tests"`
