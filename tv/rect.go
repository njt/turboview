package tv

type Point struct {
	X, Y int
}

func NewPoint(x, y int) Point {
	return Point{X: x, Y: y}
}

type Rect struct {
	A, B Point
}

func NewRect(x, y, w, h int) Rect {
	return Rect{
		A: Point{X: x, Y: y},
		B: Point{X: x + w, Y: y + h},
	}
}

func (r Rect) Width() int  { return r.B.X - r.A.X }
func (r Rect) Height() int { return r.B.Y - r.A.Y }

func (r Rect) Contains(p Point) bool {
	return p.X >= r.A.X && p.X < r.B.X && p.Y >= r.A.Y && p.Y < r.B.Y
}

func (r Rect) Intersect(s Rect) Rect {
	a := Point{X: max(r.A.X, s.A.X), Y: max(r.A.Y, s.A.Y)}
	b := Point{X: min(r.B.X, s.B.X), Y: min(r.B.Y, s.B.Y)}
	if a.X >= b.X || a.Y >= b.Y {
		return Rect{}
	}
	return Rect{A: a, B: b}
}

func (r Rect) IsEmpty() bool {
	return r.Width() <= 0 || r.Height() <= 0
}

func (r Rect) Moved(dx, dy int) Rect {
	return Rect{
		A: Point{X: r.A.X + dx, Y: r.A.Y + dy},
		B: Point{X: r.B.X + dx, Y: r.B.Y + dy},
	}
}
