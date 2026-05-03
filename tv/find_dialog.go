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
