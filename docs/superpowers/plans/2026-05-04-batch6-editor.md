# Batch 6: TEditor + TEditWindow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a full text editor widget (Editor) that extends Memo with single-level undo, find/replace, file I/O, and an indicator; wrap it in EditWindow with save-on-close behavior.

**Architecture:** Editor embeds `*Memo` (same package, full field access) and overrides `HandleEvent` to intercept Ctrl+Z/F/H/L/F3, snapshot before edits for undo, and broadcast `CmIndicatorUpdate` after cursor/text changes. EditWindow is a Window subclass that constructs and wires Editor + Indicator + ScrollBars. Find/Replace are modal dialogs launched via `ExecView`.

**Tech Stack:** Go, tcell/v2, existing tv package (Memo, Window, Dialog, ScrollBar, MessageBox, History, CheckBoxes)

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `tv/command.go` | Modify | Add CmFind, CmReplace, CmSearchAgain, CmIndicatorUpdate |
| `tv/indicator.go` | Create | Indicator view (line:col + modified flag display) |
| `tv/editor.go` | Create | Editor struct (embeds *Memo), undo, search, file I/O, modified flag |
| `tv/find_dialog.go` | Create | Find and Replace dialog constructors, prompt-on-replace loop |
| `tv/edit_window.go` | Create | EditWindow (Window subclass), wiring, save-prompt close behavior |
| `e2e/testapp/basic/main.go` | Modify | Replace Window 3 Memo with EditWindow, add Edit menu |

---

### Task 1: New Commands + Indicator

**Files:**
- Modify: `tv/command.go`
- Create: `tv/indicator.go`
- Test: `tv/indicator_test.go`

- [ ] **Step 1: Write failing test for Indicator**

```go
// tv/indicator_test.go
package tv

import "testing"

func TestIndicatorSetValue(t *testing.T) {
	ind := NewIndicator(NewRect(0, 0, 20, 1))
	ind.SetValue(5, 10, false)
	if ind.line != 5 || ind.col != 10 || ind.modified != false {
		t.Fatalf("got line=%d col=%d modified=%v", ind.line, ind.col, ind.modified)
	}
	ind.SetValue(1, 1, true)
	if ind.line != 1 || ind.col != 1 || ind.modified != true {
		t.Fatalf("got line=%d col=%d modified=%v", ind.line, ind.col, ind.modified)
	}
}

func TestIndicatorDraw(t *testing.T) {
	ind := NewIndicator(NewRect(0, 0, 20, 1))
	ind.SetValue(2, 15, false)

	buf := NewDrawBuffer(20, 1)
	ind.Draw(buf)

	got := extractIndicatorText(buf, 0, 0, 10)
	if got != " 2:15     " {
		t.Fatalf("expected ' 2:15     ' got %q", got)
	}
}

func TestIndicatorDrawModified(t *testing.T) {
	ind := NewIndicator(NewRect(0, 0, 20, 1))
	ind.SetValue(10, 3, true)

	buf := NewDrawBuffer(20, 1)
	ind.Draw(buf)

	got := extractIndicatorText(buf, 0, 0, 10)
	if got != " 10:3  *  " {
		t.Fatalf("expected ' 10:3  *  ' got %q", got)
	}
}

func TestIndicatorNotSelectable(t *testing.T) {
	ind := NewIndicator(NewRect(0, 0, 20, 1))
	if ind.HasOption(OfSelectable) {
		t.Fatal("Indicator should not be selectable")
	}
	if !ind.HasOption(OfPostProcess) {
		t.Fatal("Indicator should have OfPostProcess")
	}
}

func TestIndicatorHandlesBroadcast(t *testing.T) {
	ind := NewIndicator(NewRect(0, 0, 20, 1))

	ed := &Editor{}
	ed.Memo = NewMemo(NewRect(0, 0, 40, 10))
	ed.Memo.SetText("line1\nline2\nline3")
	ed.Memo.cursorRow = 2
	ed.Memo.cursorCol = 3
	ed.modified = true

	ev := &Event{
		What:    EvBroadcast,
		Command: CmIndicatorUpdate,
		Info:    ed,
	}
	ind.HandleEvent(ev)

	if ind.line != 3 || ind.col != 4 || ind.modified != true {
		t.Fatalf("got line=%d col=%d modified=%v", ind.line, ind.col, ind.modified)
	}
}

func extractIndicatorText(buf *DrawBuffer, row, startCol, count int) string {
	result := make([]rune, count)
	for i := 0; i < count; i++ {
		result[i] = buf.cells[row][startCol+i].Rune
	}
	return string(result)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tv/ -run TestIndicator -v -count=1`
Expected: FAIL — `NewIndicator` undefined, `CmIndicatorUpdate` undefined, `Editor` undefined

- [ ] **Step 3: Add new commands to command.go**

Add after `CmRecordHistory` (before `CmUser`):

```go
CmFind
CmReplace
CmSearchAgain
CmIndicatorUpdate
```

- [ ] **Step 4: Create indicator.go**

```go
// tv/indicator.go
package tv

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type Indicator struct {
	BaseView
	line     int
	col      int
	modified bool
}

func NewIndicator(bounds Rect) *Indicator {
	ind := &Indicator{
		line: 1,
		col:  1,
	}
	ind.SetBounds(bounds)
	ind.SetState(SfVisible, true)
	ind.SetOptions(OfPostProcess, true)
	return ind
}

func (ind *Indicator) SetValue(line, col int, modified bool) {
	ind.line = line
	ind.col = col
	ind.modified = modified
}

func (ind *Indicator) Draw(buf *DrawBuffer) {
	cs := ind.ColorScheme()
	style := tcell.StyleDefault
	if cs != nil {
		style = cs.WindowTitle
	}

	w := ind.Bounds().Width()
	buf.Fill(NewRect(0, 0, w, 1), ' ', style)

	text := fmt.Sprintf(" %d:%d", ind.line, ind.col)
	if ind.modified {
		text += "  *"
	}

	for i, ch := range []rune(text) {
		if i < w {
			buf.WriteChar(i, 0, ch, style)
		}
	}
}

func (ind *Indicator) HandleEvent(event *Event) {
	if event.What == EvBroadcast && event.Command == CmIndicatorUpdate {
		if ed, ok := event.Info.(*Editor); ok {
			row, col := ed.Memo.CursorPos()
			ind.SetValue(row+1, col+1, ed.Modified())
		}
	}
}
```

- [ ] **Step 5: Create minimal Editor stub for test compilation**

```go
// tv/editor.go
package tv

type Editor struct {
	*Memo
	modified bool
}

func (e *Editor) Modified() bool {
	return e.modified
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./tv/ -run TestIndicator -v -count=1`
Expected: PASS (all 5 indicator tests)

- [ ] **Step 7: Commit**

```bash
git add tv/command.go tv/indicator.go tv/indicator_test.go tv/editor.go
git commit -m "feat: add Indicator view and new editor commands (CmFind, CmReplace, CmSearchAgain, CmIndicatorUpdate)"
```

---

### Task 2: Editor Core with Undo

**Files:**
- Modify: `tv/editor.go`
- Test: `tv/editor_test.go`

- [ ] **Step 1: Write failing tests for Editor undo and modified flag**

```go
// tv/editor_test.go
package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func newTestEditor(text string) *Editor {
	ed := NewEditor(NewRect(0, 0, 40, 10))
	ed.SetText(text)
	return ed
}

func sendEditorKey(v View, key tcell.Key, r rune, mod tcell.ModMask) *Event {
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: key, Rune: r, Modifiers: mod},
	}
	v.HandleEvent(ev)
	return ev
}

func TestEditorEmbedsMemo(t *testing.T) {
	ed := newTestEditor("hello")
	if ed.Text() != "hello" {
		t.Fatalf("expected 'hello' got %q", ed.Text())
	}
	row, col := ed.CursorPos()
	if row != 0 || col != 0 {
		t.Fatalf("cursor at %d:%d", row, col)
	}
}

func TestEditorModifiedFlagAfterEdit(t *testing.T) {
	ed := newTestEditor("hello")
	if ed.Modified() {
		t.Fatal("should not be modified initially")
	}
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	if !ed.Modified() {
		t.Fatal("should be modified after typing")
	}
}

func TestEditorUndoRestoresText(t *testing.T) {
	ed := newTestEditor("hello")
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	if ed.Text() != "xhello" {
		t.Fatalf("expected 'xhello' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello" {
		t.Fatalf("after undo expected 'hello' got %q", ed.Text())
	}
}

func TestEditorUndoRestoresCursor(t *testing.T) {
	ed := newTestEditor("hello")
	sendEditorKey(ed, tcell.KeyEnd, 0, 0)
	_, col := ed.CursorPos()
	if col != 5 {
		t.Fatalf("expected col 5, got %d", col)
	}
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	_, col = ed.CursorPos()
	if col != 6 {
		t.Fatalf("expected col 6, got %d", col)
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	row, col := ed.CursorPos()
	if row != 0 || col != 5 {
		t.Fatalf("after undo expected 0:5, got %d:%d", row, col)
	}
}

func TestEditorUndoSingleLevel(t *testing.T) {
	ed := newTestEditor("abc")
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	sendEditorKey(ed, tcell.KeyRune, 'y', 0)
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "xabc" {
		t.Fatalf("expected 'xabc' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "xabc" {
		t.Fatalf("second undo should be no-op, got %q", ed.Text())
	}
}

func TestEditorUndoAfterDelete(t *testing.T) {
	ed := newTestEditor("hello")
	sendEditorKey(ed, tcell.KeyEnd, 0, 0)
	sendEditorKey(ed, tcell.KeyBackspace2, 0, 0)
	if ed.Text() != "hell" {
		t.Fatalf("expected 'hell' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello" {
		t.Fatalf("after undo expected 'hello' got %q", ed.Text())
	}
}

func TestEditorUndoAfterEnter(t *testing.T) {
	ed := newTestEditor("hello")
	ed.Memo.cursorCol = 3
	sendEditorKey(ed, tcell.KeyEnter, 0, 0)
	if ed.Text() != "hel\nlo" {
		t.Fatalf("expected 'hel\\nlo' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello" {
		t.Fatalf("after undo expected 'hello' got %q", ed.Text())
	}
}

func TestEditorUndoAfterPaste(t *testing.T) {
	ed := newTestEditor("hello")
	clipboard = "PASTED"
	sendEditorKey(ed, tcell.KeyCtrlV, 0, tcell.ModCtrl)
	if ed.Text() != "PASTEDhello" {
		t.Fatalf("expected 'PASTEDhello' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello" {
		t.Fatalf("after undo expected 'hello' got %q", ed.Text())
	}
}

func TestEditorUndoAfterCut(t *testing.T) {
	ed := newTestEditor("hello")
	sendEditorKey(ed, tcell.KeyCtrlA, 0, tcell.ModCtrl)
	sendEditorKey(ed, tcell.KeyCtrlX, 0, tcell.ModCtrl)
	if ed.Text() != "" {
		t.Fatalf("expected empty after cut, got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello" {
		t.Fatalf("after undo expected 'hello' got %q", ed.Text())
	}
}

func TestEditorUndoAfterDeleteLine(t *testing.T) {
	ed := newTestEditor("line1\nline2\nline3")
	ed.Memo.cursorRow = 1
	sendEditorKey(ed, tcell.KeyCtrlY, 0, tcell.ModCtrl)
	if ed.Text() != "line1\nline3" {
		t.Fatalf("expected 'line1\\nline3' got %q", ed.Text())
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "line1\nline2\nline3" {
		t.Fatalf("after undo expected original, got %q", ed.Text())
	}
}

func TestEditorCanUndo(t *testing.T) {
	ed := newTestEditor("hello")
	if ed.CanUndo() {
		t.Fatal("should not be able to undo initially")
	}
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	if !ed.CanUndo() {
		t.Fatal("should be able to undo after edit")
	}
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.CanUndo() {
		t.Fatal("should not be able to undo after undoing")
	}
}

func TestEditorNavigationDoesNotSnapshot(t *testing.T) {
	ed := newTestEditor("hello\nworld")
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	sendEditorKey(ed, tcell.KeyDown, 0, 0)
	sendEditorKey(ed, tcell.KeyRight, 0, 0)
	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "hello\nworld" {
		t.Fatalf("expected original after undo, got %q", ed.Text())
	}
}

func TestEditorSetTextClearsUndo(t *testing.T) {
	ed := newTestEditor("hello")
	sendEditorKey(ed, tcell.KeyRune, 'x', 0)
	ed.SetText("fresh")
	if ed.CanUndo() {
		t.Fatal("SetText should clear undo")
	}
	if ed.Modified() {
		t.Fatal("SetText should clear modified")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tv/ -run TestEditor -v -count=1`
Expected: FAIL — `NewEditor` undefined, `CanUndo` undefined

- [ ] **Step 3: Implement Editor with undo**

Replace the stub in `tv/editor.go` with:

```go
// tv/editor.go
package tv

import "github.com/gdamore/tcell/v2"

type undoState struct {
	lines                                          [][]rune
	cursorRow, cursorCol                           int
	selStartRow, selStartCol, selEndRow, selEndCol int
	deltaX, deltaY                                 int
}

type searchState struct {
	findText        string
	replaceText     string
	caseSensitive   bool
	wholeWords      bool
	promptOnReplace bool
}

type Editor struct {
	*Memo
	undo_      *undoState
	modified   bool
	search     searchState
	filename   string
	lineEnding string
}

func NewEditor(bounds Rect) *Editor {
	ed := &Editor{
		lineEnding: "\n",
	}
	ed.Memo = NewMemo(bounds)
	return ed
}

func (e *Editor) Modified() bool { return e.modified }
func (e *Editor) CanUndo() bool  { return e.undo_ != nil }

func (e *Editor) SetText(s string) {
	e.Memo.SetText(s)
	e.undo_ = nil
	e.modified = false
}

func (e *Editor) saveSnapshot() {
	linesCopy := make([][]rune, len(e.Memo.lines))
	for i, line := range e.Memo.lines {
		linesCopy[i] = make([]rune, len(line))
		copy(linesCopy[i], line)
	}
	e.undo_ = &undoState{
		lines:       linesCopy,
		cursorRow:   e.Memo.cursorRow,
		cursorCol:   e.Memo.cursorCol,
		selStartRow: e.Memo.selStartRow,
		selStartCol: e.Memo.selStartCol,
		selEndRow:   e.Memo.selEndRow,
		selEndCol:   e.Memo.selEndCol,
		deltaX:      e.Memo.deltaX,
		deltaY:      e.Memo.deltaY,
	}
}

func (e *Editor) restoreSnapshot() {
	if e.undo_ == nil {
		return
	}
	u := e.undo_
	e.Memo.lines = u.lines
	e.Memo.cursorRow = u.cursorRow
	e.Memo.cursorCol = u.cursorCol
	e.Memo.selStartRow = u.selStartRow
	e.Memo.selStartCol = u.selStartCol
	e.Memo.selEndRow = u.selEndRow
	e.Memo.selEndCol = u.selEndCol
	e.Memo.deltaX = u.deltaX
	e.Memo.deltaY = u.deltaY
	e.undo_ = nil
	e.Memo.syncScrollBars()
}

func (e *Editor) isEditKey(event *Event) bool {
	if event.What != EvKeyboard || event.Key == nil {
		return false
	}
	k := event.Key
	switch k.Key {
	case tcell.KeyRune, tcell.KeyEnter,
		tcell.KeyBackspace, tcell.KeyBackspace2,
		tcell.KeyDelete, tcell.KeyCtrlX, tcell.KeyCtrlV,
		tcell.KeyCtrlY:
		return true
	}
	return false
}

func (e *Editor) HandleEvent(event *Event) {
	if event.IsCleared() {
		return
	}

	// Handle command events from menus
	if event.What == EvCommand {
		switch event.Command {
		case CmFind:
			e.openFindDialog()
			event.Clear()
			return
		case CmReplace:
			e.openReplaceDialog()
			event.Clear()
			return
		case CmSearchAgain:
			if e.search.findText != "" {
				e.findNext()
			}
			e.broadcastIndicator()
			event.Clear()
			return
		}
	}

	if event.What == EvKeyboard && event.Key != nil {
		k := event.Key
		if k.Key == tcell.KeyCtrlZ {
			e.restoreSnapshot()
			e.broadcastIndicator()
			event.Clear()
			return
		}
		if k.Key == tcell.KeyCtrlF {
			e.openFindDialog()
			event.Clear()
			return
		}
		// Ctrl+H: note that tcell.KeyCtrlH == tcell.KeyBackspace (0x08).
		// On terminals that send 0x7F for Backspace (most modern terminals),
		// this works correctly. On terminals that send 0x08 for Backspace,
		// this key opens Replace instead — matching original Turbo Vision
		// behavior. Backspace2 (0x7F) is unaffected.
		if k.Key == tcell.KeyCtrlH {
			e.openReplaceDialog()
			event.Clear()
			return
		}
		if k.Key == tcell.KeyF3 || k.Key == tcell.KeyCtrlL {
			if e.search.findText != "" {
				e.findNext()
			}
			e.broadcastIndicator()
			event.Clear()
			return
		}
	}

	isEdit := e.isEditKey(event)
	if isEdit {
		e.saveSnapshot()
	}

	e.Memo.HandleEvent(event)

	if event.IsCleared() {
		if isEdit {
			e.modified = true
		}
		e.broadcastIndicator()
	}
}

func (e *Editor) broadcastIndicator() {
	owner := e.Owner()
	if owner == nil {
		return
	}
	ev := &Event{
		What:    EvBroadcast,
		Command: CmIndicatorUpdate,
		Info:    e,
	}
	for _, child := range owner.Children() {
		child.HandleEvent(ev)
	}
}
```

Note: `search`, `filename`, `lineEnding` fields are declared here but their methods are implemented in Tasks 3 and 4. `openFindDialog` and `openReplaceDialog` are defined in Task 3 (`find_dialog.go`). `findNext` is also in Task 3. For Task 2 compilation, add stubs:

```go
func (e *Editor) openFindDialog()    {}
func (e *Editor) openReplaceDialog() {}
func (e *Editor) findNext() bool     { return false }
```

These stubs go in `editor.go` and will be replaced by real implementations in `find_dialog.go` (Task 3).

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./tv/ -run TestEditor -v -count=1`
Expected: PASS

- [ ] **Step 5: Run all existing tests to check for regressions**

Run: `go test ./tv/ -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add tv/editor.go tv/editor_test.go
git commit -m "feat: implement Editor with single-level undo and modified flag"
```

---

### Task 3: Find/Replace Dialogs and Search

**Files:**
- Create: `tv/find_dialog.go`
- Modify: `tv/editor.go` (remove stubs, search methods live in find_dialog.go)
- Test: `tv/editor_search_test.go`

- [ ] **Step 1: Write failing tests for search**

```go
// tv/editor_search_test.go
package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestEditorFindForwardBasic(t *testing.T) {
	ed := newTestEditor("hello world hello")
	ed.search.findText = "hello"
	found := ed.findNext()
	if !found {
		t.Fatal("expected to find 'hello'")
	}
	sr, sc, er, ec := ed.Selection()
	if sr != 0 || sc != 0 || er != 0 || ec != 5 {
		t.Fatalf("expected selection 0:0-0:5, got %d:%d-%d:%d", sr, sc, er, ec)
	}
}

func TestEditorFindForwardFromCursor(t *testing.T) {
	ed := newTestEditor("hello world hello")
	ed.search.findText = "hello"
	ed.Memo.cursorCol = 1
	found := ed.findNext()
	if !found {
		t.Fatal("expected to find second 'hello'")
	}
	sr, sc, er, ec := ed.Selection()
	if sr != 0 || sc != 12 || er != 0 || ec != 17 {
		t.Fatalf("expected selection 0:12-0:17, got %d:%d-%d:%d", sr, sc, er, ec)
	}
}

func TestEditorFindWraps(t *testing.T) {
	ed := newTestEditor("hello world")
	ed.search.findText = "hello"
	ed.Memo.cursorCol = 6
	found := ed.findNext()
	if !found {
		t.Fatal("expected to find 'hello' by wrapping")
	}
	sr, sc, _, _ := ed.Selection()
	if sr != 0 || sc != 0 {
		t.Fatalf("expected wrap to 0:0, got %d:%d", sr, sc)
	}
}

func TestEditorFindNotFound(t *testing.T) {
	ed := newTestEditor("hello world")
	ed.search.findText = "xyz"
	found := ed.findNext()
	if found {
		t.Fatal("should not find 'xyz'")
	}
}

func TestEditorFindCaseInsensitive(t *testing.T) {
	ed := newTestEditor("Hello World")
	ed.search.findText = "hello"
	ed.search.caseSensitive = false
	found := ed.findNext()
	if !found {
		t.Fatal("case-insensitive search should find 'Hello'")
	}
}

func TestEditorFindCaseSensitive(t *testing.T) {
	ed := newTestEditor("Hello World")
	ed.search.findText = "hello"
	ed.search.caseSensitive = true
	found := ed.findNext()
	if found {
		t.Fatal("case-sensitive search should not find 'hello' in 'Hello'")
	}
}

func TestEditorFindWholeWords(t *testing.T) {
	ed := newTestEditor("helloworld hello world")
	ed.search.findText = "hello"
	ed.search.wholeWords = true
	found := ed.findNext()
	if !found {
		t.Fatal("expected to find whole word 'hello'")
	}
	_, sc, _, ec := ed.Selection()
	if sc != 11 || ec != 16 {
		t.Fatalf("expected col 11-16, got %d-%d", sc, ec)
	}
}

func TestEditorFindMultiline(t *testing.T) {
	ed := newTestEditor("line1\nfoo bar\nline3")
	ed.search.findText = "foo"
	found := ed.findNext()
	if !found {
		t.Fatal("expected to find 'foo'")
	}
	sr, sc, er, ec := ed.Selection()
	if sr != 1 || sc != 0 || er != 1 || ec != 3 {
		t.Fatalf("expected 1:0-1:3, got %d:%d-%d:%d", sr, sc, er, ec)
	}
}

func TestEditorFindScrollsToMatch(t *testing.T) {
	lines := "line0\n"
	for i := 1; i <= 30; i++ {
		lines += "filler\n"
	}
	lines += "target here"
	ed := newTestEditor(lines)
	ed.search.findText = "target"
	ed.findNext()
	row, _ := ed.CursorPos()
	if row < 30 {
		t.Fatalf("cursor should be at target row, got %d", row)
	}
	if ed.Memo.deltaY == 0 {
		t.Fatal("viewport should have scrolled to make match visible")
	}
}

func TestEditorReplaceBasic(t *testing.T) {
	ed := newTestEditor("hello world")
	ed.search.findText = "hello"
	ed.search.replaceText = "hi"
	ed.findNext()
	ed.replaceCurrent()
	if ed.Text() != "hi world" {
		t.Fatalf("expected 'hi world' got %q", ed.Text())
	}
}

func TestEditorReplaceAll(t *testing.T) {
	ed := newTestEditor("hello world hello")
	ed.search.findText = "hello"
	ed.search.replaceText = "hi"
	count := ed.replaceAll()
	if count != 2 {
		t.Fatalf("expected 2 replacements, got %d", count)
	}
	if ed.Text() != "hi world hi" {
		t.Fatalf("expected 'hi world hi' got %q", ed.Text())
	}
}

func TestEditorReplaceAllCaseInsensitive(t *testing.T) {
	ed := newTestEditor("Hello hello HELLO")
	ed.search.findText = "hello"
	ed.search.replaceText = "hi"
	ed.search.caseSensitive = false
	count := ed.replaceAll()
	if count != 3 {
		t.Fatalf("expected 3 replacements, got %d", count)
	}
	if ed.Text() != "hi hi hi" {
		t.Fatalf("expected 'hi hi hi' got %q", ed.Text())
	}
}

func TestEditorCtrlFClearsEvent(t *testing.T) {
	ed := newTestEditor("hello")
	ev := sendEditorKey(ed, tcell.KeyCtrlF, 0, tcell.ModCtrl)
	if !ev.IsCleared() {
		t.Fatal("Ctrl+F should be consumed by Editor")
	}
}

func TestEditorCtrlHClearsEvent(t *testing.T) {
	ed := newTestEditor("hello")
	ev := sendEditorKey(ed, tcell.KeyCtrlH, 0, tcell.ModCtrl)
	if !ev.IsCleared() {
		t.Fatal("Ctrl+H should be consumed by Editor")
	}
}

func TestEditorSearchAgainNoOp(t *testing.T) {
	ed := newTestEditor("hello")
	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyF3},
	}
	ed.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Fatal("F3 should be consumed")
	}
}

func TestEditorSearchAgainRepeats(t *testing.T) {
	ed := newTestEditor("aaa bbb aaa bbb")
	ed.search.findText = "bbb"
	ed.findNext()
	_, sc1, _, _ := ed.Selection()
	ed.findNext()
	_, sc2, _, _ := ed.Selection()
	if sc1 == sc2 {
		t.Fatal("search again should advance to next match")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tv/ -run "TestEditorFind|TestEditorReplace|TestEditorCtrl|TestEditorSearch" -v -count=1`
Expected: FAIL — `findNext` always returns false (stub), `replaceCurrent`, `replaceAll` undefined

- [ ] **Step 3: Implement search and replace on Editor**

Create `tv/find_dialog.go` with search/replace methods and dialog constructors. Remove the stubs from `editor.go`:

```go
// tv/find_dialog.go
package tv

import (
	"fmt"
	"strings"
	"unicode"
)

func (e *Editor) findNext() bool {
	if e.search.findText == "" {
		return false
	}
	needle := []rune(e.search.findText)
	if !e.search.caseSensitive {
		needle = toLowerRunes(needle)
	}
	needleLen := len(needle)

	startRow := e.Memo.cursorRow
	startCol := e.Memo.cursorCol

	// Search forward from cursor position
	if row, col, ok := e.findFrom(startRow, startCol, needle, needleLen); ok {
		e.selectMatch(row, col, needleLen)
		return true
	}

	// Wrap: search from start to cursor position
	if row, col, ok := e.findFromTo(0, 0, startRow, startCol, needle, needleLen); ok {
		e.selectMatch(row, col, needleLen)
		return true
	}

	return false
}

func (e *Editor) findFrom(startRow, startCol int, needle []rune, needleLen int) (int, int, bool) {
	for row := startRow; row < len(e.Memo.lines); row++ {
		line := e.Memo.lines[row]
		fromCol := 0
		if row == startRow {
			fromCol = startCol
		}
		if col, ok := e.findInLine(row, line, fromCol, needle, needleLen); ok {
			return row, col, true
		}
	}
	return 0, 0, false
}

func (e *Editor) findFromTo(startRow, startCol, endRow, endCol int, needle []rune, needleLen int) (int, int, bool) {
	for row := startRow; row <= endRow && row < len(e.Memo.lines); row++ {
		line := e.Memo.lines[row]
		fromCol := 0
		if row == startRow {
			fromCol = startCol
		}
		maxCol := len(line)
		if row == endRow && endCol < maxCol {
			maxCol = endCol
		}
		for col := fromCol; col <= maxCol-needleLen; col++ {
			if e.matchAtPos(line, col, needle, needleLen) && e.checkWholeWord(row, col, needleLen) {
				return row, col, true
			}
		}
	}
	return 0, 0, false
}

func (e *Editor) findInLine(row int, line []rune, fromCol int, needle []rune, needleLen int) (int, bool) {
	for col := fromCol; col <= len(line)-needleLen; col++ {
		if e.matchAtPos(line, col, needle, needleLen) && e.checkWholeWord(row, col, needleLen) {
			return col, true
		}
	}
	return 0, false
}

func (e *Editor) matchAtPos(line []rune, col int, needle []rune, needleLen int) bool {
	if col+needleLen > len(line) {
		return false
	}
	for i := 0; i < needleLen; i++ {
		ch := line[col+i]
		if !e.search.caseSensitive {
			ch = unicode.ToLower(ch)
		}
		if ch != needle[i] {
			return false
		}
	}
	return true
}

func (e *Editor) checkWholeWord(row, col, needleLen int) bool {
	if !e.search.wholeWords {
		return true
	}
	line := e.Memo.lines[row]
	if col > 0 && charClass(line[col-1]) == charClass(line[col]) {
		return false
	}
	endCol := col + needleLen
	if endCol < len(line) && charClass(line[endCol-1]) == charClass(line[endCol]) {
		return false
	}
	return true
}

func (e *Editor) selectMatch(row, col, length int) {
	e.Memo.cursorRow = row
	e.Memo.cursorCol = col + length
	e.Memo.selStartRow = row
	e.Memo.selStartCol = col
	e.Memo.selEndRow = row
	e.Memo.selEndCol = col + length
	e.Memo.ensureCursorVisible()
	e.Memo.syncScrollBars()
}

func (e *Editor) replaceCurrent() {
	if !e.Memo.HasSelection() {
		return
	}
	sr, sc, er, ec := e.Memo.normalizedSelection()
	selText := e.selectedTextRange(sr, sc, er, ec)
	needle := e.search.findText
	if !e.search.caseSensitive {
		selText = strings.ToLower(selText)
		needle = strings.ToLower(needle)
	}
	if selText != needle {
		return
	}
	e.saveSnapshot()
	e.Memo.deleteSelection()
	e.Memo.insertText(e.search.replaceText)
	e.modified = true
	e.broadcastIndicator()
}

func (e *Editor) selectedTextRange(sr, sc, er, ec int) string {
	if sr == er {
		line := e.Memo.lines[sr]
		if sc > len(line) {
			sc = len(line)
		}
		if ec > len(line) {
			ec = len(line)
		}
		return string(line[sc:ec])
	}
	var sb strings.Builder
	sb.WriteString(string(e.Memo.lines[sr][sc:]))
	for r := sr + 1; r < er; r++ {
		sb.WriteRune('\n')
		sb.WriteString(string(e.Memo.lines[r]))
	}
	sb.WriteRune('\n')
	sb.WriteString(string(e.Memo.lines[er][:ec]))
	return sb.String()
}

func (e *Editor) replaceAll() int {
	needle := []rune(e.search.findText)
	if !e.search.caseSensitive {
		needle = toLowerRunes(needle)
	}
	needleLen := len(needle)
	replaceRunes := []rune(e.search.replaceText)
	count := 0

	for row := 0; row < len(e.Memo.lines); row++ {
		line := e.Memo.lines[row]
		col := 0
		for col <= len(line)-needleLen {
			if e.matchAtPos(line, col, needle, needleLen) && e.checkWholeWord(row, col, needleLen) {
				newLine := make([]rune, len(line)-needleLen+len(replaceRunes))
				copy(newLine, line[:col])
				copy(newLine[col:], replaceRunes)
				copy(newLine[col+len(replaceRunes):], line[col+needleLen:])
				e.Memo.lines[row] = newLine
				line = newLine
				col += len(replaceRunes)
				count++
			} else {
				col++
			}
		}
	}

	if count > 0 {
		e.Memo.cursorRow = 0
		e.Memo.cursorCol = 0
		e.Memo.clearSelection()
		e.Memo.syncScrollBars()
		e.modified = true
		e.broadcastIndicator()
	}
	return count
}

func toLowerRunes(rs []rune) []rune {
	out := make([]rune, len(rs))
	for i, r := range rs {
		out[i] = unicode.ToLower(r)
	}
	return out
}

func (e *Editor) openFindDialog() {
	owner := e.Owner()
	if owner == nil {
		return
	}

	dlg := NewDialog(NewRect(0, 0, 40, 10), "Find")
	ob := owner.Bounds()
	dx := (ob.Width() - 40) / 2
	dy := (ob.Height() - 10) / 2
	if dx < 0 {
		dx = 0
	}
	if dy < 0 {
		dy = 0
	}
	dlg.SetBounds(NewRect(dx, dy, 40, 10))

	searchInput := NewInputLine(NewRect(1, 1, 25, 1), 256)
	searchInput.SetText(e.search.findText)
	searchLabel := NewLabel(NewRect(1, 0, 25, 1), "~S~earch for:", searchInput)
	searchHist := NewHistory(NewRect(26, 1, 1, 1), searchInput, 10)
	dlg.Insert(searchLabel)
	dlg.Insert(searchInput)
	dlg.Insert(searchHist)

	optCB := NewCheckBoxes(NewRect(1, 3, 36, 2),
		[]string{"~C~ase sensitive", "~W~hole words only"},
	)
	if e.search.caseSensitive {
		optCB.Item(0).SetChecked(true)
	}
	if e.search.wholeWords {
		optCB.Item(1).SetChecked(true)
	}
	dlg.Insert(optCB)

	okBtn := NewButton(NewRect(1, 6, 12, 2), "OK", CmOK, WithDefault())
	cancelBtn := NewButton(NewRect(15, 6, 12, 2), "Cancel", CmCancel)
	dlg.Insert(okBtn)
	dlg.Insert(cancelBtn)

	result := owner.ExecView(dlg)
	if result == CmOK {
		e.search.findText = searchInput.Text()
		e.search.caseSensitive = optCB.Item(0).Checked()
		e.search.wholeWords = optCB.Item(1).Checked()
		e.findNext()
		e.broadcastIndicator()
	}
}

func (e *Editor) openReplaceDialog() {
	owner := e.Owner()
	if owner == nil {
		return
	}

	dlg := NewDialog(NewRect(0, 0, 40, 14), "Replace")
	ob := owner.Bounds()
	dx := (ob.Width() - 40) / 2
	dy := (ob.Height() - 14) / 2
	if dx < 0 {
		dx = 0
	}
	if dy < 0 {
		dy = 0
	}
	dlg.SetBounds(NewRect(dx, dy, 40, 14))

	searchInput := NewInputLine(NewRect(1, 1, 25, 1), 256)
	searchInput.SetText(e.search.findText)
	searchLabel := NewLabel(NewRect(1, 0, 25, 1), "~S~earch for:", searchInput)
	searchHist := NewHistory(NewRect(26, 1, 1, 1), searchInput, 10)
	dlg.Insert(searchLabel)
	dlg.Insert(searchInput)
	dlg.Insert(searchHist)

	replaceInput := NewInputLine(NewRect(1, 3, 25, 1), 256)
	replaceInput.SetText(e.search.replaceText)
	replaceLabel := NewLabel(NewRect(1, 2, 25, 1), "~R~eplace with:", replaceInput)
	replaceHist := NewHistory(NewRect(26, 3, 1, 1), replaceInput, 11)
	dlg.Insert(replaceLabel)
	dlg.Insert(replaceInput)
	dlg.Insert(replaceHist)

	optCB := NewCheckBoxes(NewRect(1, 5, 36, 3),
		[]string{"~C~ase sensitive", "~W~hole words only", "~P~rompt on replace"},
	)
	if e.search.caseSensitive {
		optCB.Item(0).SetChecked(true)
	}
	if e.search.wholeWords {
		optCB.Item(1).SetChecked(true)
	}
	if e.search.promptOnReplace {
		optCB.Item(2).SetChecked(true)
	}
	dlg.Insert(optCB)

	okBtn := NewButton(NewRect(1, 9, 12, 2), "OK", CmOK, WithDefault())
	replAllBtn := NewButton(NewRect(14, 9, 14, 2), "Replace All", CmYes)
	cancelBtn := NewButton(NewRect(29, 9, 10, 2), "Cancel", CmCancel)
	dlg.Insert(okBtn)
	dlg.Insert(replAllBtn)
	dlg.Insert(cancelBtn)

	result := owner.ExecView(dlg)
	if result == CmOK || result == CmYes {
		e.search.findText = searchInput.Text()
		e.search.replaceText = replaceInput.Text()
		e.search.caseSensitive = optCB.Item(0).Checked()
		e.search.wholeWords = optCB.Item(1).Checked()
		e.search.promptOnReplace = optCB.Item(2).Checked()

		if result == CmYes {
			e.saveSnapshot()
			count := e.replaceAll()
			if count > 0 {
				MessageBox(owner, "Replace",
					fmt.Sprintf("%d occurrence(s) replaced.", count),
					MbOK)
			} else {
				MessageBox(owner, "Replace", "Search string not found.", MbOK)
			}
		} else if e.search.promptOnReplace {
			e.replaceWithPrompt()
		} else {
			if e.findNext() {
				e.replaceCurrent()
				e.findNext()
			}
			e.broadcastIndicator()
		}
	}
}

func (e *Editor) replaceWithPrompt() {
	owner := e.Owner()
	if owner == nil {
		return
	}
	for e.findNext() {
		e.broadcastIndicator()
		result := MessageBox(owner, "Replace",
			"Replace this occurrence?",
			MbYes|MbNo|MbCancel)
		switch result {
		case CmYes:
			e.replaceCurrent()
		case CmNo:
			continue
		default:
			return
		}
	}
}
```

- [ ] **Step 4: Remove stubs from editor.go**

Remove these three stub functions from `tv/editor.go`:

```go
func (e *Editor) openFindDialog()    {}
func (e *Editor) openReplaceDialog() {}
func (e *Editor) findNext() bool     { return false }
```

They are now implemented in `tv/find_dialog.go`.

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./tv/ -run "TestEditorFind|TestEditorReplace|TestEditorCtrl|TestEditorSearch" -v -count=1`
Expected: PASS

- [ ] **Step 6: Run all tests**

Run: `go test ./tv/ -count=1`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add tv/editor.go tv/find_dialog.go tv/editor_search_test.go
git commit -m "feat: add find/replace with search dialogs, history, and prompt-on-replace"
```

---

### Task 4: File I/O

**Files:**
- Modify: `tv/editor.go`
- Test: `tv/editor_fileio_test.go`

- [ ] **Step 1: Write failing tests for file I/O**

```go
// tv/editor_fileio_test.go
package tv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEditorLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello\nworld"), 0644)

	ed := newTestEditor("")
	err := ed.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if ed.Text() != "hello\nworld" {
		t.Fatalf("expected 'hello\\nworld' got %q", ed.Text())
	}
	if ed.FileName() != path {
		t.Fatalf("expected filename %q got %q", path, ed.FileName())
	}
	if ed.Modified() {
		t.Fatal("should not be modified after load")
	}
	if ed.CanUndo() {
		t.Fatal("should not have undo after load")
	}
}

func TestEditorSaveFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	ed := newTestEditor("save me")
	ed.modified = true
	err := ed.SaveFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "save me" {
		t.Fatalf("expected 'save me' got %q", string(data))
	}
	if ed.Modified() {
		t.Fatal("should not be modified after save")
	}
	if ed.FileName() != path {
		t.Fatalf("expected filename %q got %q", path, ed.FileName())
	}
}

func TestEditorLineEndingDetectionCRLF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "crlf.txt")
	os.WriteFile(path, []byte("hello\r\nworld\r\n"), 0644)

	ed := newTestEditor("")
	ed.LoadFile(path)
	if ed.Text() != "hello\nworld\n" {
		t.Fatalf("CRLF not normalized on load: %q", ed.Text())
	}

	outPath := filepath.Join(dir, "out.txt")
	ed.SaveFile(outPath)
	data, _ := os.ReadFile(outPath)
	if string(data) != "hello\r\nworld\r\n" {
		t.Fatalf("expected CRLF preserved on save, got %q", string(data))
	}
}

func TestEditorLineEndingDetectionLF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lf.txt")
	os.WriteFile(path, []byte("hello\nworld\n"), 0644)

	ed := newTestEditor("")
	ed.LoadFile(path)

	outPath := filepath.Join(dir, "out.txt")
	ed.SaveFile(outPath)
	data, _ := os.ReadFile(outPath)
	if string(data) != "hello\nworld\n" {
		t.Fatalf("expected LF preserved on save, got %q", string(data))
	}
}

func TestEditorLoadFileNotFound(t *testing.T) {
	ed := newTestEditor("")
	err := ed.LoadFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestEditorSaveFileError(t *testing.T) {
	ed := newTestEditor("data")
	err := ed.SaveFile("/nonexistent/dir/file.txt")
	if err == nil {
		t.Fatal("expected error for bad path")
	}
}

func TestEditorNewFileHasNoFilename(t *testing.T) {
	ed := newTestEditor("hello")
	if ed.FileName() != "" {
		t.Fatalf("new editor should have empty filename, got %q", ed.FileName())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tv/ -run "TestEditorLoad|TestEditorSave|TestEditorLineEnding|TestEditorNewFile" -v -count=1`
Expected: FAIL — `LoadFile`, `SaveFile`, `FileName` undefined

- [ ] **Step 3: Implement file I/O on Editor**

Add to `tv/editor.go` (the `filename` and `lineEnding` fields are already declared in the struct from Task 2):

```go
import (
	"os"
	"strings"
)

func (e *Editor) FileName() string { return e.filename }

func (e *Editor) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	if strings.Contains(content, "\r\n") {
		e.lineEnding = "\r\n"
	} else {
		e.lineEnding = "\n"
	}
	e.SetText(content)
	e.filename = path
	return nil
}

func (e *Editor) SaveFile(path string) error {
	text := e.Text()
	if e.lineEnding == "\r\n" {
		text = strings.ReplaceAll(text, "\n", "\r\n")
	}
	err := os.WriteFile(path, []byte(text), 0644)
	if err != nil {
		return err
	}
	e.filename = path
	e.modified = false
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./tv/ -run "TestEditorLoad|TestEditorSave|TestEditorLineEnding|TestEditorNewFile" -v -count=1`
Expected: PASS

- [ ] **Step 5: Run all tests**

Run: `go test ./tv/ -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add tv/editor.go tv/editor_fileio_test.go
git commit -m "feat: add file I/O with line ending detection to Editor"
```

---

### Task 5: EditWindow + Save Prompt

**Files:**
- Create: `tv/edit_window.go`
- Test: `tv/edit_window_test.go`

- [ ] **Step 1: Write failing tests for EditWindow**

```go
// tv/edit_window_test.go
package tv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEditWindowCreation(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	if ew.Title() != "Untitled" {
		t.Fatalf("expected title 'Untitled' got %q", ew.Title())
	}
	if ew.Editor() == nil {
		t.Fatal("Editor should not be nil")
	}
}

func TestEditWindowWithFilename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0644)

	ew := NewEditWindow(NewRect(0, 0, 60, 20), path)
	if ew.Editor().Text() != "hello" {
		t.Fatalf("expected 'hello' got %q", ew.Editor().Text())
	}
	if ew.Title() != "test.txt" {
		t.Fatalf("expected title 'test.txt' got %q", ew.Title())
	}
}

func TestEditWindowHasAllChildren(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	children := ew.Children()
	hasVScroll := false
	hasHScroll := false
	hasEditor := false
	hasIndicator := false
	for _, child := range children {
		switch v := child.(type) {
		case *ScrollBar:
			if v.Bounds().Width() == 1 {
				hasVScroll = true
			} else {
				hasHScroll = true
			}
		case *Editor:
			hasEditor = true
		case *Indicator:
			hasIndicator = true
		}
	}
	if !hasVScroll {
		t.Fatal("missing vertical scrollbar")
	}
	if !hasHScroll {
		t.Fatal("missing horizontal scrollbar")
	}
	if !hasEditor {
		t.Fatal("missing editor")
	}
	if !hasIndicator {
		t.Fatal("missing indicator")
	}
}

func TestEditWindowIndicatorVisible(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	var ind *Indicator
	for _, child := range ew.Children() {
		if i, ok := child.(*Indicator); ok {
			ind = i
			break
		}
	}
	if ind == nil {
		t.Fatal("indicator not found")
	}
	clientH := 20 - 2
	if ind.Bounds().A.Y >= clientH {
		t.Fatalf("indicator at y=%d is outside client area (height %d)", ind.Bounds().A.Y, clientH)
	}
}

func TestEditWindowCloseIntercepted(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	ew.Editor().modified = true
	ev := &Event{What: EvCommand, Command: CmClose}
	ew.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Fatal("CmClose on modified EditWindow should be intercepted (cancelled)")
	}
}

func TestEditWindowEditorIsFocused(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	focused := ew.FocusedChild()
	if _, ok := focused.(*Editor); !ok {
		t.Fatal("Editor should be the focused child")
	}
}

func TestEditWindowIndicatorUpdates(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	ed := ew.Editor()
	ed.SetText("line1\nline2\nline3")

	ed.Memo.cursorRow = 1
	ed.Memo.cursorCol = 3
	ed.broadcastIndicator()

	var ind *Indicator
	for _, child := range ew.Children() {
		if i, ok := child.(*Indicator); ok {
			ind = i
			break
		}
	}
	if ind == nil {
		t.Fatal("indicator not found")
	}
	if ind.line != 2 || ind.col != 4 {
		t.Fatalf("expected indicator 2:4, got %d:%d", ind.line, ind.col)
	}
}

func TestEditWindowValidUnmodified(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	if !ew.Valid(CmClose) {
		t.Fatal("unmodified EditWindow should be valid for close")
	}
}

func TestEditWindowValidUnmodifiedQuit(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	if !ew.Valid(CmQuit) {
		t.Fatal("unmodified EditWindow should be valid for quit")
	}
}

func TestEditWindowMinSize(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 24, 6), "")
	b := ew.Bounds()
	if b.Width() < 24 || b.Height() < 6 {
		t.Fatalf("minimum size not enforced: %dx%d", b.Width(), b.Height())
	}
}

func TestEditWindowMinSizeEnforced(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 10, 3), "")
	b := ew.Bounds()
	if b.Width() < 24 || b.Height() < 6 {
		t.Fatalf("small bounds should be clamped to minimum: got %dx%d", b.Width(), b.Height())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tv/ -run TestEditWindow -v -count=1`
Expected: FAIL — `NewEditWindow` undefined

- [ ] **Step 3: Implement EditWindow**

```go
// tv/edit_window.go
package tv

import "path/filepath"

type EditWindow struct {
	*Window
	editor    *Editor
	indicator *Indicator
}

const editWindowIndicatorWidth = 14

func NewEditWindow(bounds Rect, filename string, opts ...WindowOption) *EditWindow {
	title := "Untitled"
	if filename != "" {
		title = filepath.Base(filename)
	}

	w := bounds.Width()
	h := bounds.Height()
	if w < 24 {
		w = 24
	}
	if h < 6 {
		h = 6
	}
	bounds = NewRect(bounds.A.X, bounds.A.Y, w, h)

	win := NewWindow(bounds, title, opts...)
	ew := &EditWindow{Window: win}

	clientW := max(w-2, 0)
	clientH := max(h-2, 0)

	// Layout: Editor fills top rows. Bottom row shared by Indicator (left) and HScroll (right).
	// VScroll on right edge, above bottom row.
	bottomY := clientH - 1

	vScroll := NewScrollBar(NewRect(clientW-1, 0, 1, bottomY), Vertical)
	vScroll.SetGrowMode(GfGrowLoX | GfGrowHiX | GfGrowHiY)

	indW := editWindowIndicatorWidth
	if indW > clientW-2 {
		indW = clientW - 2
	}
	hScrollX := indW
	hScrollW := max(clientW-1-hScrollX, 0)
	hScroll := NewScrollBar(NewRect(hScrollX, bottomY, hScrollW, 1), Horizontal)
	hScroll.SetGrowMode(GfGrowLoY | GfGrowHiY | GfGrowHiX)

	editorW := max(clientW-1, 0)
	editorH := max(bottomY, 0)
	editor := NewEditor(NewRect(0, 0, editorW, editorH))
	editor.SetVScrollBar(vScroll)
	editor.SetHScrollBar(hScroll)
	editor.SetGrowMode(GfGrowHiX | GfGrowHiY)

	ind := NewIndicator(NewRect(0, bottomY, indW, 1))
	ind.SetGrowMode(GfGrowLoY | GfGrowHiY)

	ew.editor = editor
	ew.indicator = ind

	win.Insert(editor)
	win.Insert(vScroll)
	win.Insert(hScroll)
	win.Insert(ind)

	if filename != "" {
		editor.LoadFile(filename)
	}

	return ew
}

func (ew *EditWindow) Editor() *Editor { return ew.editor }

func (ew *EditWindow) HandleEvent(event *Event) {
	if event.What == EvCommand && event.Command == CmClose {
		if !ew.Valid(CmClose) {
			event.Clear()
			return
		}
	}
	ew.Window.HandleEvent(event)
}

func (ew *EditWindow) Valid(cmd CommandCode) bool {
	if cmd != CmClose && cmd != CmQuit {
		return true
	}
	if !ew.editor.Modified() {
		return true
	}
	result := MessageBox(ew, "Confirm",
		"Save changes to "+ew.Title()+"?",
		MbYes|MbNo|MbCancel)
	switch result {
	case CmYes:
		if ew.editor.FileName() == "" {
			return false
		}
		err := ew.editor.SaveFile(ew.editor.FileName())
		return err == nil
	case CmNo:
		return true
	default:
		return false
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./tv/ -run TestEditWindow -v -count=1`
Expected: PASS

- [ ] **Step 5: Run all tests**

Run: `go test ./tv/ -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add tv/edit_window.go tv/edit_window_test.go
git commit -m "feat: add EditWindow with Indicator, ScrollBars, min size, and save prompt"
```

---

### Task 6: Integration Checkpoint — Editor Pipeline

**Files:**
- Test: `tv/integration_editor_test.go`

- [ ] **Step 1: Write integration tests**

```go
// tv/integration_editor_test.go
package tv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

func TestIntegrationEditorUndoFlow(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	ew.SetColorScheme(&theme.BorlandBlue)
	ed := ew.Editor()
	ed.SetText("original")

	sendEditorKey(ed, tcell.KeyRune, 'X', 0)
	if ed.Text() != "Xoriginal" {
		t.Fatalf("expected 'Xoriginal' got %q", ed.Text())
	}
	if !ed.Modified() {
		t.Fatal("should be modified")
	}

	sendEditorKey(ed, tcell.KeyCtrlZ, 0, tcell.ModCtrl)
	if ed.Text() != "original" {
		t.Fatalf("expected 'original' after undo, got %q", ed.Text())
	}
}

func TestIntegrationEditorSearchReplace(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	ew.SetColorScheme(&theme.BorlandBlue)
	ed := ew.Editor()
	ed.SetText("foo bar foo baz foo")

	ed.search.findText = "foo"
	ed.search.replaceText = "qux"

	found := ed.findNext()
	if !found {
		t.Fatal("should find 'foo'")
	}
	_, sc, _, _ := ed.Selection()
	if sc != 0 {
		t.Fatalf("first match at col 0, got %d", sc)
	}

	ed.replaceCurrent()
	if ed.Text() != "qux bar foo baz foo" {
		t.Fatalf("expected 'qux bar foo baz foo' got %q", ed.Text())
	}

	count := ed.replaceAll()
	if count != 2 {
		t.Fatalf("expected 2 more replacements, got %d", count)
	}
	if ed.Text() != "qux bar qux baz qux" {
		t.Fatalf("expected all replaced, got %q", ed.Text())
	}
}

func TestIntegrationEditorFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello\nworld"), 0644)

	ew := NewEditWindow(NewRect(0, 0, 60, 20), path)
	ed := ew.Editor()

	if ed.Text() != "hello\nworld" {
		t.Fatalf("load failed: %q", ed.Text())
	}
	if ew.Title() != "test.txt" {
		t.Fatalf("title: %q", ew.Title())
	}

	sendEditorKey(ed, tcell.KeyRune, '!', 0)
	if !ed.Modified() {
		t.Fatal("should be modified")
	}

	outPath := filepath.Join(dir, "out.txt")
	err := ed.SaveFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if ed.Modified() {
		t.Fatal("should not be modified after save")
	}

	data, _ := os.ReadFile(outPath)
	if string(data) != "!hello\nworld" {
		t.Fatalf("saved content: %q", string(data))
	}
}

func TestIntegrationIndicatorTracksEditor(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	ew.SetColorScheme(&theme.BorlandBlue)
	ed := ew.Editor()
	ed.SetText("line1\nline2\nline3")

	var ind *Indicator
	for _, child := range ew.Children() {
		if i, ok := child.(*Indicator); ok {
			ind = i
			break
		}
	}
	if ind == nil {
		t.Fatal("no indicator")
	}

	sendEditorKey(ed, tcell.KeyDown, 0, 0)
	if ind.line != 2 || ind.col != 1 {
		t.Fatalf("expected 2:1, got %d:%d", ind.line, ind.col)
	}

	sendEditorKey(ed, tcell.KeyRune, 'X', 0)
	if !ind.modified {
		t.Fatal("indicator should show modified")
	}
}

func TestIntegrationEditWindowDraw(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 40, 12), "")
	ew.SetColorScheme(&theme.BorlandBlue)
	ew.Editor().SetText("Hello Editor")

	buf := NewDrawBuffer(40, 12)
	ew.Draw(buf)

	titleFound := false
	for x := 0; x < 40; x++ {
		if buf.cells[0][x].Rune == 'U' {
			titleFound = true
			break
		}
	}
	if !titleFound {
		t.Fatal("title 'Untitled' not found in frame")
	}
}

func TestIntegrationEditWindowCloseUnmodified(t *testing.T) {
	ew := NewEditWindow(NewRect(0, 0, 60, 20), "")
	if !ew.Valid(CmClose) {
		t.Fatal("unmodified window should allow close")
	}
}
```

- [ ] **Step 2: Run integration tests**

Run: `go test ./tv/ -run "TestIntegration.*Editor|TestIntegration.*Indicator|TestIntegration.*EditWindow" -v -count=1`
Expected: PASS

- [ ] **Step 3: Run full test suite**

Run: `go test ./tv/ ./theme/ -count=1`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add tv/integration_editor_test.go
git commit -m "test: add integration tests for Editor, Indicator, EditWindow pipeline"
```

---

### Task 7: Demo App + E2E Tests

**Files:**
- Modify: `e2e/testapp/basic/main.go`
- Modify: `e2e/e2e_test.go`

- [ ] **Step 1: Update demo app to use EditWindow**

In `e2e/testapp/basic/main.go`, replace the Window 3 section (lines 112-137 approximately):

**Remove** these lines:
```go
// Window 3 — Memo (multi-line editor with scrollbars)
win3 := tv.NewWindow(tv.NewRect(45, 1, 35, 16), "Notes", tv.WithWindowNumber(3))
vScroll := tv.NewScrollBar(tv.NewRect(32, 0, 1, 13), tv.Vertical)
hScroll := tv.NewScrollBar(tv.NewRect(0, 13, 32, 1), tv.Horizontal)
memo := tv.NewMemo(tv.NewRect(0, 0, 32, 13), tv.WithScrollBars(hScroll, vScroll))
memo.SetText(`Hello, World!
...content...`)
win3.Insert(memo)
win3.Insert(vScroll)
win3.Insert(hScroll)
vScroll.SetGrowMode(tv.GfGrowLoX | tv.GfGrowHiX | tv.GfGrowHiY)
hScroll.SetGrowMode(tv.GfGrowLoY | tv.GfGrowHiY | tv.GfGrowHiX)
```

**Replace** with:
```go
// Window 3 — EditWindow (full editor with undo, find/replace)
win3 := tv.NewEditWindow(tv.NewRect(45, 1, 35, 16), "", tv.WithWindowNumber(3))
win3.Editor().SetText(`Hello, Editor!
This is the Editor widget.
It supports undo (Ctrl+Z), find (Ctrl+F),
replace (Ctrl+H), and search-again (F3).
Arrow keys navigate. Shift+arrow selects.
Ctrl+A selects all. Ctrl+C/X/V for clipboard.
Home/End for line start/end. Ctrl+Home/End for doc.
PgUp/PgDn scroll. Mouse wheel scrolls too.
Click positions cursor. Double-click selects word.

Line 11: Tab between windows to test focus.
Line 12: Try scrolling past this point.
Line 13: More content below visible area.
Line 14: Horizontal scrolling test — this line extends past the visible width.
Line 15: Almost at the bottom.
Line 16: Last line of demo content.`)
win3.SetGrowMode(tv.GfGrowHiX | tv.GfGrowHiY)
```

Add an Edit menu to the menu bar (after File, before Window):
```go
tv.NewSubMenu("~E~dit",
    tv.NewMenuItem("~F~ind...", tv.CmFind, tv.KbCtrl('F')),
    tv.NewMenuItem("~R~eplace...", tv.CmReplace, tv.KbCtrl('H')),
    tv.NewMenuItem("~S~earch Again", tv.CmSearchAgain, tv.KbFunc(3)),
),
```

- [ ] **Step 2: Build and verify demo compiles**

Run: `go build ./e2e/testapp/basic/`
Expected: SUCCESS

- [ ] **Step 3: Add E2E tests**

Add to `e2e/e2e_test.go`:

```go
func TestEditWindowVisible(t *testing.T) {
	sess := startTestApp(t)
	defer sess.cleanup()

	snap := sess.snapshot()
	if !strings.Contains(snap, "Editor") {
		t.Fatal("EditWindow content not visible")
	}
}

func TestEditWindowUndoE2E(t *testing.T) {
	sess := startTestApp(t)
	defer sess.cleanup()

	sess.sendKeys("M-3")
	time.Sleep(200 * time.Millisecond)

	sess.sendKeys("Z")
	time.Sleep(100 * time.Millisecond)

	snap := sess.snapshot()
	if !strings.Contains(snap, "Z") {
		t.Fatal("typed character not visible")
	}

	sess.sendKeys("C-z")
	time.Sleep(100 * time.Millisecond)
}

func TestEditWindowIndicatorVisible(t *testing.T) {
	sess := startTestApp(t)
	defer sess.cleanup()

	sess.sendKeys("M-3")
	time.Sleep(200 * time.Millisecond)

	snap := sess.snapshot()
	if !strings.Contains(snap, "1:1") {
		t.Fatal("indicator not showing initial position")
	}
}
```

- [ ] **Step 4: Run E2E tests**

Run: `go test ./e2e/ -run TestEditWindow -v -count=1 -timeout 60s`
Expected: PASS

- [ ] **Step 5: Run full E2E suite**

Run: `go test ./e2e/ -v -count=1 -timeout 120s`
Expected: PASS (new + pre-existing)

- [ ] **Step 6: Run full unit test suite**

Run: `go test ./tv/ ./theme/ -count=1`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add e2e/testapp/basic/main.go e2e/e2e_test.go
git commit -m "feat: add EditWindow to demo app with E2E tests"
```

---

## Notes for Implementers

1. **Memo field access:** Editor is in the same package (`tv`) as Memo, so it can directly access unexported fields like `lines`, `cursorRow`, `cursorCol`, etc. through the embedded `*Memo`.

2. **Clipboard:** The existing `var clipboard string` in `tv/input_line.go` is package-level. Both Memo and Editor share it. No changes needed.

3. **Ctrl+H collision:** `tcell.KeyCtrlH` and `tcell.KeyBackspace` are the same key code (0x08). Editor intercepts `KeyCtrlH` before Memo's `HandleEvent` runs, so 0x08 opens Replace instead of backspacing. Most modern terminals send `KeyBackspace2` (0x7F) for the Backspace key, which Memo handles normally. On rare terminals that send 0x08 for Backspace, the Backspace key opens Replace instead — this matches original Turbo Vision behavior. Editor also handles `CmReplace` as a command event so the menu item always works.

4. **charClass reuse:** The `charClass` function in `memo.go` is package-level (unexported but accessible within `tv`). Editor's whole-word matching reuses it for word boundary detection.

5. **ExecView for dialogs:** Find/Replace dialogs use `e.Owner().ExecView(dlg)` which runs a modal event loop. The Editor must have an owner for dialogs to work. Unit tests that test `findNext()` directly don't need an owner.

6. **Indicator position:** The Indicator and HScroll share the bottom row of the client area. The Indicator occupies the left 14 columns, the HScroll occupies the rest. This keeps the Indicator within the drawable client area (the Window's `SubBuffer` clips children to client bounds).

7. **History widget IDs:** The find dialog uses historyID=10, replace dialog uses historyID=10 for search and historyID=11 for replace text. These are arbitrary but must be distinct.

8. **DrawBuffer API:** Use `buf.WriteChar(x, y, ch, style)` not `buf.SetCell(...)`. Cell struct uses `.Rune` not `.ch`.

9. **CheckBoxes API:** Use `optCB.Item(i).Checked()` and `optCB.Item(i).SetChecked(true)` — not `SetValue(i, bool)` or `Value(i)` which don't exist.

10. **EditWindow.HandleEvent:** EditWindow overrides `HandleEvent` to intercept `CmClose` and call `Valid()` — the framework does NOT call `Valid()` automatically. Without this override, closing an EditWindow with unsaved changes would silently discard them.

11. **Editor command handling:** Editor handles both keyboard shortcuts (Ctrl+F/H/Z, F3) AND command events (CmFind, CmReplace, CmSearchAgain). This ensures menu items work even if the keyboard shortcut has a collision (like Ctrl+H/Backspace).
