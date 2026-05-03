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
