package tv

import (
	"testing"
)

// TestNewPoint verifies: "NewPoint(3, 7) returns Point{X: 3, Y: 7}"
func TestNewPoint(t *testing.T) {
	p := NewPoint(3, 7)
	if p.X != 3 || p.Y != 7 {
		t.Errorf("NewPoint(3, 7) = {X: %d, Y: %d}, want {X: 3, Y: 7}", p.X, p.Y)
	}
}

// TestNewPointDifferentValues verifies NewPoint works with different values
func TestNewPointDifferentValues(t *testing.T) {
	tests := []struct {
		x, y int
	}{
		{0, 0},
		{-5, 10},
		{100, 200},
		{-100, -200},
	}
	for _, tt := range tests {
		p := NewPoint(tt.x, tt.y)
		if p.X != tt.x || p.Y != tt.y {
			t.Errorf("NewPoint(%d, %d) = {X: %d, Y: %d}, want {X: %d, Y: %d}", tt.x, tt.y, p.X, p.Y, tt.x, tt.y)
		}
	}
}

// TestNewRect verifies: "NewRect(5, 3, 20, 10) returns Rect with A=(5,3), B=(25,13)"
// The spec shows NewRect takes (x, y, w, h) and computes B = A + (w, h)
func TestNewRect(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	if r.A.X != 5 || r.A.Y != 3 {
		t.Errorf("NewRect(5, 3, 20, 10): A = (%d, %d), want (5, 3)", r.A.X, r.A.Y)
	}
	if r.B.X != 25 || r.B.Y != 13 {
		t.Errorf("NewRect(5, 3, 20, 10): B = (%d, %d), want (25, 13)", r.B.X, r.B.Y)
	}
}

// TestNewRectZeroSize verifies NewRect works with zero dimensions
func TestNewRectZeroSize(t *testing.T) {
	r := NewRect(10, 20, 0, 0)
	if r.A.X != 10 || r.A.Y != 20 {
		t.Errorf("NewRect(10, 20, 0, 0): A = (%d, %d), want (10, 20)", r.A.X, r.A.Y)
	}
	if r.B.X != 10 || r.B.Y != 20 {
		t.Errorf("NewRect(10, 20, 0, 0): B = (%d, %d), want (10, 20)", r.B.X, r.B.Y)
	}
}

// TestNewRectNegativeCoordinates verifies NewRect works with negative coordinates
func TestNewRectNegativeCoordinates(t *testing.T) {
	r := NewRect(-5, -3, 10, 6)
	if r.A.X != -5 || r.A.Y != -3 {
		t.Errorf("NewRect(-5, -3, 10, 6): A = (%d, %d), want (-5, -3)", r.A.X, r.A.Y)
	}
	if r.B.X != 5 || r.B.Y != 3 {
		t.Errorf("NewRect(-5, -3, 10, 6): B = (%d, %d), want (5, 3)", r.B.X, r.B.Y)
	}
}

// TestRectWidth verifies: "Rect.Width() returns B.X - A.X"
func TestRectWidth(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	width := r.Width()
	if width != 20 {
		t.Errorf("NewRect(5, 3, 20, 10).Width() = %d, want 20", width)
	}
}

// TestRectWidthZero verifies Width with zero width
func TestRectWidthZero(t *testing.T) {
	r := NewRect(10, 10, 0, 5)
	if r.Width() != 0 {
		t.Errorf("NewRect(10, 10, 0, 5).Width() = %d, want 0", r.Width())
	}
}

// TestRectWidthNegative verifies Width with negative width
func TestRectWidthNegative(t *testing.T) {
	r := Rect{A: NewPoint(10, 5), B: NewPoint(5, 10)}
	if r.Width() != -5 {
		t.Errorf("Rect with A.X=10, B.X=5: Width() = %d, want -5", r.Width())
	}
}

// TestRectHeight verifies: "Rect.Height() returns B.Y - A.Y"
func TestRectHeight(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	height := r.Height()
	if height != 10 {
		t.Errorf("NewRect(5, 3, 20, 10).Height() = %d, want 10", height)
	}
}

// TestRectHeightZero verifies Height with zero height
func TestRectHeightZero(t *testing.T) {
	r := NewRect(10, 10, 5, 0)
	if r.Height() != 0 {
		t.Errorf("NewRect(10, 10, 5, 0).Height() = %d, want 0", r.Height())
	}
}

// TestRectHeightNegative verifies Height with negative height
func TestRectHeightNegative(t *testing.T) {
	r := Rect{A: NewPoint(5, 10), B: NewPoint(10, 5)}
	if r.Height() != -5 {
		t.Errorf("Rect with A.Y=10, B.Y=5: Height() = %d, want -5", r.Height())
	}
}

// TestRectContainsPointInside verifies:
// "Rect.Contains(point) returns true for points within the rect (A inclusive, B exclusive)"
func TestRectContainsPointInside(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	// Test A (inclusive)
	if !r.Contains(NewPoint(5, 3)) {
		t.Error("NewRect(5, 3, 20, 10).Contains(Point{5, 3}) = false, want true (A is inclusive)")
	}
	// Test point well inside
	if !r.Contains(NewPoint(10, 8)) {
		t.Error("NewRect(5, 3, 20, 10).Contains(Point{10, 8}) = false, want true")
	}
	// Test point near B but before it
	if !r.Contains(NewPoint(24, 12)) {
		t.Error("NewRect(5, 3, 20, 10).Contains(Point{24, 12}) = false, want true (just before B)")
	}
}

// TestRectContainsBExclusive verifies:
// "Rect.Contains(B) returns false (B is exclusive)"
func TestRectContainsBExclusive(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	if r.Contains(NewPoint(25, 13)) {
		t.Error("NewRect(5, 3, 20, 10).Contains(Point{25, 13}) = true, want false (B is exclusive)")
	}
}

// TestRectContainsBoundaryExclusive verifies B boundaries are exclusive in both dimensions
func TestRectContainsBoundaryExclusive(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	// B.X is exclusive
	if r.Contains(NewPoint(25, 8)) {
		t.Error("Contains(B.X, inside Y) = true, want false (B.X is exclusive)")
	}
	// B.Y is exclusive
	if r.Contains(NewPoint(10, 13)) {
		t.Error("Contains(inside X, B.Y) = true, want false (B.Y is exclusive)")
	}
}

// TestRectContainsPointOutside verifies Contains returns false for points outside
func TestRectContainsPointOutside(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	testCases := []struct {
		name string
		p    Point
	}{
		{"before A.X", NewPoint(4, 8)},
		{"before A.Y", NewPoint(10, 2)},
		{"after B.X", NewPoint(26, 8)},
		{"after B.Y", NewPoint(10, 14)},
		{"completely outside", NewPoint(100, 100)},
	}
	for _, tc := range testCases {
		if r.Contains(tc.p) {
			t.Errorf("Contains(%s): got true, want false", tc.name)
		}
	}
}

// TestRectContainsEdgesInclusive verifies edges of A are inclusive
func TestRectContainsEdgesInclusive(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	testCases := []struct {
		name string
		p    Point
	}{
		{"A.X edge, inside Y", NewPoint(5, 8)},
		{"A.Y edge, inside X", NewPoint(10, 3)},
		{"A corner", NewPoint(5, 3)},
	}
	for _, tc := range testCases {
		if !r.Contains(tc.p) {
			t.Errorf("Contains(%s): got false, want true (A edges are inclusive)", tc.name)
		}
	}
}

// TestRectIntersectOverlapping verifies:
// "Rect.Intersect(other) returns the overlapping region"
func TestRectIntersectOverlapping(t *testing.T) {
	r1 := NewRect(5, 3, 20, 10)  // A=(5,3), B=(25,13)
	r2 := NewRect(10, 5, 20, 10) // A=(10,5), B=(30,15)
	// Overlap: A=(10,5), B=(25,13)
	inter := r1.Intersect(r2)
	if inter.A.X != 10 || inter.A.Y != 5 {
		t.Errorf("Intersect: A = (%d, %d), want (10, 5)", inter.A.X, inter.A.Y)
	}
	if inter.B.X != 25 || inter.B.Y != 13 {
		t.Errorf("Intersect: B = (%d, %d), want (25, 13)", inter.B.X, inter.B.Y)
	}
}

// TestRectIntersectNoOverlap verifies:
// "two non-overlapping rects yield an empty rect"
func TestRectIntersectNoOverlap(t *testing.T) {
	r1 := NewRect(0, 0, 10, 10)   // A=(0,0), B=(10,10)
	r2 := NewRect(20, 20, 10, 10) // A=(20,20), B=(30,30)
	inter := r1.Intersect(r2)
	if !inter.IsEmpty() {
		t.Errorf("Intersect of non-overlapping rects: got width=%d height=%d, want empty", inter.Width(), inter.Height())
	}
}

// TestRectIntersectPartialOverlap verifies Intersect with edge cases
func TestRectIntersectPartialOverlap(t *testing.T) {
	r1 := NewRect(0, 0, 10, 10)   // A=(0,0), B=(10,10)
	r2 := NewRect(5, 5, 10, 10)   // A=(5,5), B=(15,15)
	inter := r1.Intersect(r2)
	// Overlap: A=(5,5), B=(10,10)
	if inter.A.X != 5 || inter.A.Y != 5 || inter.B.X != 10 || inter.B.Y != 10 {
		t.Errorf("Intersect partial overlap: A=(%d,%d) B=(%d,%d), want A=(5,5) B=(10,10)",
			inter.A.X, inter.A.Y, inter.B.X, inter.B.Y)
	}
}

// TestRectIntersectTouchingEdges verifies Intersect when rects touch at edges
func TestRectIntersectTouchingEdges(t *testing.T) {
	r1 := NewRect(0, 0, 10, 10)  // A=(0,0), B=(10,10)
	r2 := NewRect(10, 0, 10, 10) // A=(10,0), B=(20,10)
	inter := r1.Intersect(r2)
	// B.X=10 is exclusive in r1, A.X=10 is inclusive in r2, so no overlap
	if !inter.IsEmpty() {
		t.Errorf("Intersect of touching edges: got width=%d height=%d, want empty", inter.Width(), inter.Height())
	}
}

// TestRectIntersectCommutativity verifies Intersect is commutative
func TestRectIntersectCommutativity(t *testing.T) {
	r1 := NewRect(0, 0, 10, 10)
	r2 := NewRect(5, 5, 10, 10)
	inter1 := r1.Intersect(r2)
	inter2 := r2.Intersect(r1)
	if inter1.A.X != inter2.A.X || inter1.A.Y != inter2.A.Y ||
		inter1.B.X != inter2.B.X || inter1.B.Y != inter2.B.Y {
		t.Errorf("Intersect not commutative: r1.Intersect(r2) != r2.Intersect(r1)")
	}
}

// TestRectIsEmptyZeroWidth verifies:
// "Rect.IsEmpty() returns true when width or height is zero or negative"
func TestRectIsEmptyZeroWidth(t *testing.T) {
	r := NewRect(5, 5, 0, 10)
	if !r.IsEmpty() {
		t.Error("Rect with width=0: IsEmpty() = false, want true")
	}
}

// TestRectIsEmptyZeroHeight verifies IsEmpty with zero height
func TestRectIsEmptyZeroHeight(t *testing.T) {
	r := NewRect(5, 5, 10, 0)
	if !r.IsEmpty() {
		t.Error("Rect with height=0: IsEmpty() = false, want true")
	}
}

// TestRectIsEmptyNegativeWidth verifies IsEmpty with negative width
func TestRectIsEmptyNegativeWidth(t *testing.T) {
	r := Rect{A: NewPoint(10, 5), B: NewPoint(5, 15)}
	if !r.IsEmpty() {
		t.Error("Rect with negative width: IsEmpty() = false, want true")
	}
}

// TestRectIsEmptyNegativeHeight verifies IsEmpty with negative height
func TestRectIsEmptyNegativeHeight(t *testing.T) {
	r := Rect{A: NewPoint(5, 10), B: NewPoint(15, 5)}
	if !r.IsEmpty() {
		t.Error("Rect with negative height: IsEmpty() = false, want true")
	}
}

// TestRectIsEmptyPositive verifies IsEmpty returns false for non-empty rects
func TestRectIsEmptyPositive(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	if r.IsEmpty() {
		t.Error("NewRect(5, 3, 20, 10): IsEmpty() = true, want false")
	}
}

// TestRectMovedBasic verifies:
// "Rect.Moved(dx, dy) shifts both A and B by the given offsets"
func TestRectMovedBasic(t *testing.T) {
	r := NewRect(5, 3, 20, 10)  // A=(5,3), B=(25,13)
	moved := r.Moved(3, 7)
	// A moves to (8, 10), B moves to (28, 20)
	if moved.A.X != 8 || moved.A.Y != 10 {
		t.Errorf("Moved(3, 7): A = (%d, %d), want (8, 10)", moved.A.X, moved.A.Y)
	}
	if moved.B.X != 28 || moved.B.Y != 20 {
		t.Errorf("Moved(3, 7): B = (%d, %d), want (28, 20)", moved.B.X, moved.B.Y)
	}
}

// TestRectMovedNegativeOffset verifies Moved with negative offsets
func TestRectMovedNegativeOffset(t *testing.T) {
	r := NewRect(10, 10, 20, 20)  // A=(10,10), B=(30,30)
	moved := r.Moved(-5, -3)
	if moved.A.X != 5 || moved.A.Y != 7 {
		t.Errorf("Moved(-5, -3): A = (%d, %d), want (5, 7)", moved.A.X, moved.A.Y)
	}
	if moved.B.X != 25 || moved.B.Y != 27 {
		t.Errorf("Moved(-5, -3): B = (%d, %d), want (25, 27)", moved.B.X, moved.B.Y)
	}
}

// TestRectMovedZeroOffset verifies Moved with zero offset
func TestRectMovedZeroOffset(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	moved := r.Moved(0, 0)
	if moved.A.X != r.A.X || moved.A.Y != r.A.Y ||
		moved.B.X != r.B.X || moved.B.Y != r.B.Y {
		t.Error("Moved(0, 0) should return equivalent rect")
	}
}

// TestRectMovedPreservesSize verifies Moved preserves width and height
func TestRectMovedPreservesSize(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	moved := r.Moved(100, 200)
	if moved.Width() != r.Width() || moved.Height() != r.Height() {
		t.Errorf("Moved changed dimensions: width %d->%d, height %d->%d",
			r.Width(), moved.Width(), r.Height(), moved.Height())
	}
}

// TestRectMovedDoesNotModifyOriginal verifies Moved returns new rect without modifying original
func TestRectMovedDoesNotModifyOriginal(t *testing.T) {
	r := NewRect(5, 3, 20, 10)
	originalA := r.A
	originalB := r.B
	_ = r.Moved(10, 20)
	if r.A != originalA || r.B != originalB {
		t.Error("Moved modified the original rect")
	}
}
