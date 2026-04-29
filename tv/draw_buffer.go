package tv

import "github.com/gdamore/tcell/v2"

type Cell struct {
	Rune  rune
	Combc []rune
	Style tcell.Style
}

type DrawBuffer struct {
	cells  [][]Cell
	clip   Rect
	offset Point
}

func NewDrawBuffer(w, h int) *DrawBuffer {
	cells := make([][]Cell, h)
	for y := range cells {
		cells[y] = make([]Cell, w)
		for x := range cells[y] {
			cells[y][x] = Cell{Rune: ' ', Style: tcell.StyleDefault}
		}
	}
	return &DrawBuffer{
		cells: cells,
		clip:  NewRect(0, 0, w, h),
	}
}

func (db *DrawBuffer) WriteChar(x, y int, ch rune, style tcell.Style) {
	ax, ay := x+db.offset.X, y+db.offset.Y
	if !db.clip.Contains(NewPoint(ax, ay)) {
		return
	}
	if ay < 0 || ay >= len(db.cells) || ax < 0 || ax >= len(db.cells[0]) {
		return
	}
	db.cells[ay][ax] = Cell{Rune: ch, Style: style}
}

func (db *DrawBuffer) WriteStr(x, y int, s string, style tcell.Style) {
	col := 0
	for _, ch := range s {
		db.WriteChar(x+col, y, ch, style)
		col++
	}
}

func (db *DrawBuffer) Fill(r Rect, ch rune, style tcell.Style) {
	for y := r.A.Y; y < r.B.Y; y++ {
		for x := r.A.X; x < r.B.X; x++ {
			db.WriteChar(x, y, ch, style)
		}
	}
}

func (db *DrawBuffer) SubBuffer(r Rect) *DrawBuffer {
	absClip := Rect{
		A: NewPoint(r.A.X+db.offset.X, r.A.Y+db.offset.Y),
		B: NewPoint(r.B.X+db.offset.X, r.B.Y+db.offset.Y),
	}
	return &DrawBuffer{
		cells:  db.cells,
		clip:   db.clip.Intersect(absClip),
		offset: NewPoint(r.A.X+db.offset.X, r.A.Y+db.offset.Y),
	}
}

func (db *DrawBuffer) SetCellStyle(x, y int, style tcell.Style) {
	ax, ay := x+db.offset.X, y+db.offset.Y
	if !db.clip.Contains(NewPoint(ax, ay)) {
		return
	}
	if ay < 0 || ay >= len(db.cells) || ax < 0 || ax >= len(db.cells[0]) {
		return
	}
	db.cells[ay][ax].Style = style
}

func (db *DrawBuffer) GetCell(x, y int) Cell {
	ax, ay := x+db.offset.X, y+db.offset.Y
	if ay < 0 || ay >= len(db.cells) || ax < 0 || ax >= len(db.cells[0]) {
		return Cell{}
	}
	return db.cells[ay][ax]
}

func (db *DrawBuffer) Width() int  { return len(db.cells[0]) }
func (db *DrawBuffer) Height() int { return len(db.cells) }

func (db *DrawBuffer) FlushTo(screen tcell.Screen) {
	for y, row := range db.cells {
		for x, cell := range row {
			screen.SetContent(x, y, cell.Rune, cell.Combc, cell.Style)
		}
	}
}
