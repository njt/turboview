package tv

import (
	"testing"
)

// Test: GfGrowLoX shifts left edge by width delta
// Spec: "GfGrowLoX: child's left edge (A.X) shifts by the width delta (owner grew wider → child shifts right)"
func TestGfGrowLoXShiftsLeftEdgeByWidthDelta(t *testing.T) {
	// Confirming test: child with GfGrowLoX shifts right when owner grows wider
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 10))
	child.SetGrowMode(GfGrowLoX)
	g.Insert(child)

	// Owner grows from 40 to 60 (deltaW = 20)
	g.SetBounds(NewRect(0, 0, 60, 20))

	// Child's left edge should shift right by 20
	if child.Bounds().A.X != 25 {
		t.Errorf("GfGrowLoX: child A.X = %d, want 25 (5 + 20)", child.Bounds().A.X)
	}
	// Width should remain unchanged
	if child.Bounds().Width() != 10 {
		t.Errorf("GfGrowLoX: child width = %d, want 10 (unchanged)", child.Bounds().Width())
	}
}

// Falsifying test: ensure child with GfGrowLoX does NOT shift when owner width doesn't change
func TestGfGrowLoXNoShiftWhenWidthUnchanged(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 10))
	child.SetGrowMode(GfGrowLoX)
	g.Insert(child)

	// Owner changes height only, width stays 40
	g.SetBounds(NewRect(0, 0, 40, 30))

	// Child should not move horizontally
	if child.Bounds().A.X != 5 {
		t.Errorf("GfGrowLoX with no width change: child A.X = %d, want 5 (unchanged)", child.Bounds().A.X)
	}
}

// Test: GfGrowHiX shifts right edge by width delta
// Spec: "GfGrowHiX: child's right edge (B.X) shifts by the width delta (owner grew wider → child's right edge moves right, making child wider)"
func TestGfGrowHiXShiftsRightEdgeByWidthDelta(t *testing.T) {
	// Confirming test: child with GfGrowHiX stretches right when owner grows wider
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 10))
	child.SetGrowMode(GfGrowHiX)
	g.Insert(child)

	// Owner grows from 40 to 60 (deltaW = 20)
	g.SetBounds(NewRect(0, 0, 60, 20))

	// Child's right edge should shift right by 20, making child wider
	// Original: A.X=5, B.X=15, Width=10
	// New: A.X=5, B.X=35, Width=30
	if child.Bounds().B.X != 35 {
		t.Errorf("GfGrowHiX: child B.X = %d, want 35 (15 + 20)", child.Bounds().B.X)
	}
	if child.Bounds().Width() != 30 {
		t.Errorf("GfGrowHiX: child width = %d, want 30", child.Bounds().Width())
	}
}

// Falsifying test: ensure child with GfGrowHiX does NOT shift when owner width doesn't change
func TestGfGrowHiXNoShiftWhenWidthUnchanged(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 10))
	child.SetGrowMode(GfGrowHiX)
	g.Insert(child)

	// Owner changes height only, width stays 40
	g.SetBounds(NewRect(0, 0, 40, 30))

	// Child's width should not change
	if child.Bounds().Width() != 10 {
		t.Errorf("GfGrowHiX with no width change: child width = %d, want 10 (unchanged)", child.Bounds().Width())
	}
}

// Test: GfGrowLoY shifts top edge by height delta
// Spec: "GfGrowLoY: child's top edge (A.Y) shifts by the height delta (owner grew taller → child shifts down)"
func TestGfGrowLoYShiftsTopEdgeByHeightDelta(t *testing.T) {
	// Confirming test: child with GfGrowLoY shifts down when owner grows taller
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 10))
	child.SetGrowMode(GfGrowLoY)
	g.Insert(child)

	// Owner grows from 20 to 30 (deltaH = 10)
	g.SetBounds(NewRect(0, 0, 40, 30))

	// Child's top edge should shift down by 10
	if child.Bounds().A.Y != 15 {
		t.Errorf("GfGrowLoY: child A.Y = %d, want 15 (5 + 10)", child.Bounds().A.Y)
	}
	// Height should remain unchanged
	if child.Bounds().Height() != 10 {
		t.Errorf("GfGrowLoY: child height = %d, want 10 (unchanged)", child.Bounds().Height())
	}
}

// Falsifying test: ensure child with GfGrowLoY does NOT shift when owner height doesn't change
func TestGfGrowLoYNoShiftWhenHeightUnchanged(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 10))
	child.SetGrowMode(GfGrowLoY)
	g.Insert(child)

	// Owner changes width only, height stays 20
	g.SetBounds(NewRect(0, 0, 60, 20))

	// Child should not move vertically
	if child.Bounds().A.Y != 5 {
		t.Errorf("GfGrowLoY with no height change: child A.Y = %d, want 5 (unchanged)", child.Bounds().A.Y)
	}
}

// Test: GfGrowHiY shifts bottom edge by height delta
// Spec: "GfGrowHiY: child's bottom edge (B.Y) shifts by the height delta (owner grew taller → child's bottom edge moves down, making child taller)"
func TestGfGrowHiYShiftsBottomEdgeByHeightDelta(t *testing.T) {
	// Confirming test: child with GfGrowHiY stretches down when owner grows taller
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 10))
	child.SetGrowMode(GfGrowHiY)
	g.Insert(child)

	// Owner grows from 20 to 30 (deltaH = 10)
	g.SetBounds(NewRect(0, 0, 40, 30))

	// Child's bottom edge should shift down by 10, making child taller
	// Original: A.Y=5, B.Y=15, Height=10
	// New: A.Y=5, B.Y=25, Height=20
	if child.Bounds().B.Y != 25 {
		t.Errorf("GfGrowHiY: child B.Y = %d, want 25 (15 + 10)", child.Bounds().B.Y)
	}
	if child.Bounds().Height() != 20 {
		t.Errorf("GfGrowHiY: child height = %d, want 20", child.Bounds().Height())
	}
}

// Falsifying test: ensure child with GfGrowHiY does NOT shift when owner height doesn't change
func TestGfGrowHiYNoShiftWhenHeightUnchanged(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 10))
	child.SetGrowMode(GfGrowHiY)
	g.Insert(child)

	// Owner changes width only, height stays 20
	g.SetBounds(NewRect(0, 0, 60, 20))

	// Child's height should not change
	if child.Bounds().Height() != 10 {
		t.Errorf("GfGrowHiY with no height change: child height = %d, want 10 (unchanged)", child.Bounds().Height())
	}
}

// Test: GfGrowAll shifts both edges, maintaining size
// Spec: "GfGrowAll (all four flags): both edges on both axes shift by delta — the child shifts position but maintains its size"
func TestGfGrowAllShiftsPositionMaintainsSize(t *testing.T) {
	// Confirming test: child with GfGrowAll shifts but keeps its size
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 3))
	child.SetGrowMode(GfGrowAll)
	g.Insert(child)

	// Owner grows by deltaW=20, deltaH=5
	g.SetBounds(NewRect(0, 0, 60, 25))

	// Child should shift by the deltas but maintain size
	if child.Bounds().A.X != 25 {
		t.Errorf("GfGrowAll: child A.X = %d, want 25 (5 + 20)", child.Bounds().A.X)
	}
	if child.Bounds().A.Y != 10 {
		t.Errorf("GfGrowAll: child A.Y = %d, want 10 (5 + 5)", child.Bounds().A.Y)
	}
	if child.Bounds().Width() != 10 {
		t.Errorf("GfGrowAll: child width = %d, want 10 (maintained)", child.Bounds().Width())
	}
	if child.Bounds().Height() != 3 {
		t.Errorf("GfGrowAll: child height = %d, want 3 (maintained)", child.Bounds().Height())
	}
}

// Falsifying test: ensure GfGrowAll doesn't shrink child when owner shrinks
func TestGfGrowAllShrinksPositionWithOwner(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 60, 25))
	child := newMockView(NewRect(25, 10, 10, 3))
	child.SetGrowMode(GfGrowAll)
	g.Insert(child)

	// Owner shrinks by deltaW=-20, deltaH=-5
	g.SetBounds(NewRect(0, 0, 40, 20))

	// Child should shift back by the deltas but maintain size
	if child.Bounds().A.X != 5 {
		t.Errorf("GfGrowAll on shrink: child A.X = %d, want 5 (25 - 20)", child.Bounds().A.X)
	}
	if child.Bounds().A.Y != 5 {
		t.Errorf("GfGrowAll on shrink: child A.Y = %d, want 5 (10 - 5)", child.Bounds().A.Y)
	}
	if child.Bounds().Width() != 10 {
		t.Errorf("GfGrowAll on shrink: child width = %d, want 10 (maintained)", child.Bounds().Width())
	}
	if child.Bounds().Height() != 3 {
		t.Errorf("GfGrowAll on shrink: child height = %d, want 3 (maintained)", child.Bounds().Height())
	}
}

// Test: GfGrowHiX|GfGrowHiY stretches (concrete example from spec)
// Spec: "A child with GfGrowHiX | GfGrowHiY at position (5,5) with size (30,15): when owner grows by deltaW=20, deltaH=5, the child becomes position (5,5) size (50,20)"
func TestGfGrowHiXHiYStretches(t *testing.T) {
	// Confirming test: child with GfGrowHiX|GfGrowHiY stretches
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 30, 15))
	child.SetGrowMode(GfGrowHiX | GfGrowHiY)
	g.Insert(child)

	// Owner grows by deltaW=20, deltaH=5 (40→60, 20→25)
	g.SetBounds(NewRect(0, 0, 60, 25))

	// Child position stays (5,5), size becomes (50,20)
	if child.Bounds().A.X != 5 || child.Bounds().A.Y != 5 {
		t.Errorf("GfGrowHiX|GfGrowHiY: child position = (%d,%d), want (5,5)", child.Bounds().A.X, child.Bounds().A.Y)
	}
	if child.Bounds().Width() != 50 {
		t.Errorf("GfGrowHiX|GfGrowHiY: child width = %d, want 50 (30 + 20)", child.Bounds().Width())
	}
	if child.Bounds().Height() != 20 {
		t.Errorf("GfGrowHiX|GfGrowHiY: child height = %d, want 20 (15 + 5)", child.Bounds().Height())
	}
}

// Falsifying test: ensure GfGrowHiX|GfGrowHiY doesn't shift position
func TestGfGrowHiXHiYDoesNotShiftPosition(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 30, 15))
	child.SetGrowMode(GfGrowHiX | GfGrowHiY)
	g.Insert(child)

	// Owner grows
	g.SetBounds(NewRect(0, 0, 60, 25))

	// Position must not change
	if child.Bounds().A.X != 5 {
		t.Errorf("GfGrowHiX|GfGrowHiY should not shift X: A.X = %d, want 5", child.Bounds().A.X)
	}
	if child.Bounds().A.Y != 5 {
		t.Errorf("GfGrowHiX|GfGrowHiY should not shift Y: A.Y = %d, want 5", child.Bounds().A.Y)
	}
}

// Test: GfGrowLoX|GfGrowLoY shifts position (concrete example from spec)
// Spec: "A child with GfGrowLoX | GfGrowLoY at position (5,5) with size (10,3): when owner grows by deltaW=20, deltaH=5, the child becomes position (25,10) size (10,3) — it shifts to stay at the bottom-right region"
func TestGfGrowLoXLoYShiftsPosition(t *testing.T) {
	// Confirming test: child with GfGrowLoX|GfGrowLoY shifts to bottom-right region
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 3))
	child.SetGrowMode(GfGrowLoX | GfGrowLoY)
	g.Insert(child)

	// Owner grows by deltaW=20, deltaH=5 (40→60, 20→25)
	g.SetBounds(NewRect(0, 0, 60, 25))

	// Child position becomes (25,10), size stays (10,3)
	if child.Bounds().A.X != 25 {
		t.Errorf("GfGrowLoX|GfGrowLoY: child A.X = %d, want 25 (5 + 20)", child.Bounds().A.X)
	}
	if child.Bounds().A.Y != 10 {
		t.Errorf("GfGrowLoX|GfGrowLoY: child A.Y = %d, want 10 (5 + 5)", child.Bounds().A.Y)
	}
	if child.Bounds().Width() != 10 {
		t.Errorf("GfGrowLoX|GfGrowLoY: child width = %d, want 10 (unchanged)", child.Bounds().Width())
	}
	if child.Bounds().Height() != 3 {
		t.Errorf("GfGrowLoX|GfGrowLoY: child height = %d, want 3 (unchanged)", child.Bounds().Height())
	}
}

// Falsifying test: ensure GfGrowLoX|GfGrowLoY doesn't stretch
func TestGfGrowLoXLoYDoesNotStretch(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 3))
	child.SetGrowMode(GfGrowLoX | GfGrowLoY)
	g.Insert(child)

	g.SetBounds(NewRect(0, 0, 60, 25))

	// Size must not change
	if child.Bounds().Width() != 10 {
		t.Errorf("GfGrowLoX|GfGrowLoY should not stretch width: width = %d, want 10", child.Bounds().Width())
	}
	if child.Bounds().Height() != 3 {
		t.Errorf("GfGrowLoX|GfGrowLoY should not stretch height: height = %d, want 3", child.Bounds().Height())
	}
}

// Test: GfGrowRel proportional scaling
// Spec: "GfGrowRel: all four edges scale proportionally. If owner was 40x20 and becomes 80x40 (2x), a child at (10,5,20,10) becomes (20,10,40,20)"
func TestGfGrowRelProportionalScaling(t *testing.T) {
	// Confirming test: child with GfGrowRel scales proportionally
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(10, 5, 20, 10))
	child.SetGrowMode(GfGrowRel)
	g.Insert(child)

	// Owner becomes 2x: 40→80, 20→40
	g.SetBounds(NewRect(0, 0, 80, 40))

	// All coordinates should double
	if child.Bounds().A.X != 20 {
		t.Errorf("GfGrowRel: child A.X = %d, want 20 (10 * 2)", child.Bounds().A.X)
	}
	if child.Bounds().A.Y != 10 {
		t.Errorf("GfGrowRel: child A.Y = %d, want 10 (5 * 2)", child.Bounds().A.Y)
	}
	if child.Bounds().Width() != 40 {
		t.Errorf("GfGrowRel: child width = %d, want 40 (20 * 2)", child.Bounds().Width())
	}
	if child.Bounds().Height() != 20 {
		t.Errorf("GfGrowRel: child height = %d, want 20 (10 * 2)", child.Bounds().Height())
	}
}

// Falsifying test: ensure GfGrowRel doesn't apply when scale factor is 1:1
func TestGfGrowRelNoChangeWhenScaleIs1(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(10, 5, 20, 10))
	child.SetGrowMode(GfGrowRel)
	g.Insert(child)

	// Owner stays the same size
	g.SetBounds(NewRect(0, 0, 40, 20))

	// Child should not move or resize
	if child.Bounds().A.X != 10 || child.Bounds().A.Y != 5 {
		t.Errorf("GfGrowRel 1:1 scale: child position = (%d,%d), want (10,5)", child.Bounds().A.X, child.Bounds().A.Y)
	}
	if child.Bounds().Width() != 20 || child.Bounds().Height() != 10 {
		t.Errorf("GfGrowRel 1:1 scale: child size = (%d,%d), want (20,10)", child.Bounds().Width(), child.Bounds().Height())
	}
}

// Test: GfGrowRel division by zero safety
// Spec: "Division by zero is safe — if old width or height is 0, skip proportional for that axis"
func TestGfGrowRelDivisionByZeroSafetyZeroOwnerWidth(t *testing.T) {
	// Owner starts with 0 width — proportional X-axis scaling would divide by zero
	g := NewGroup(NewRect(0, 0, 0, 20))
	child := newMockView(NewRect(0, 5, 0, 10))
	child.SetGrowMode(GfGrowRel)
	g.Insert(child)

	// Owner grows to non-zero width
	g.SetBounds(NewRect(0, 0, 80, 40))

	// Must not panic. X-axis skipped (old width 0). Y-axis scales: 40/20 = 2x.
	if child.Bounds().A.Y != 10 {
		t.Errorf("GfGrowRel zero-owner-width: child A.Y = %d, want 10 (5 * 2)", child.Bounds().A.Y)
	}
	if child.Bounds().Height() != 20 {
		t.Errorf("GfGrowRel zero-owner-width: child height = %d, want 20 (10 * 2)", child.Bounds().Height())
	}
}

// Test: GfGrowRel division by zero safety for owner height
func TestGfGrowRelDivisionByZeroSafetyZeroOwnerHeight(t *testing.T) {
	// Owner starts with 0 height — proportional Y-axis scaling would divide by zero
	g := NewGroup(NewRect(0, 0, 40, 0))
	child := newMockView(NewRect(10, 0, 20, 0))
	child.SetGrowMode(GfGrowRel)
	g.Insert(child)

	// Owner grows to non-zero height
	g.SetBounds(NewRect(0, 0, 80, 40))

	// Must not panic. Y-axis skipped (old height 0). X-axis scales: 80/40 = 2x.
	if child.Bounds().A.X != 20 {
		t.Errorf("GfGrowRel zero-owner-height: child A.X = %d, want 20 (10 * 2)", child.Bounds().A.X)
	}
	if child.Bounds().Width() != 40 {
		t.Errorf("GfGrowRel zero-owner-height: child width = %d, want 40 (20 * 2)", child.Bounds().Width())
	}
}

// Test: GfGrowRel division by zero safety for both owner dimensions
func TestGfGrowRelDivisionByZeroSafetyBothOwnerZero(t *testing.T) {
	// Owner starts with 0x0 — proportional scaling would divide by zero on both axes
	g := NewGroup(NewRect(0, 0, 0, 0))
	child := newMockView(NewRect(0, 0, 0, 0))
	child.SetGrowMode(GfGrowRel)
	g.Insert(child)

	// Owner grows to non-zero
	g.SetBounds(NewRect(0, 0, 80, 40))

	// Must not panic. Both axes skipped (old dimensions 0).
	// Child should remain at origin with zero size.
	if child.Bounds().A.X != 0 || child.Bounds().A.Y != 0 {
		t.Errorf("GfGrowRel zero-owner-both: child position = (%d,%d), want (0,0)", child.Bounds().A.X, child.Bounds().A.Y)
	}
}

// Test: GfGrowRel fractional scaling
func TestGfGrowRelFractionalScaling(t *testing.T) {
	// Confirming test: child with GfGrowRel scales fractionally (e.g., 1.5x)
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(10, 10, 20, 10))
	child.SetGrowMode(GfGrowRel)
	g.Insert(child)

	// Owner becomes 1.5x: 40→60, 20→30
	g.SetBounds(NewRect(0, 0, 60, 30))

	// All coordinates should scale by 1.5
	// 10*1.5=15, 20*1.5=30, 10*1.5=15
	if child.Bounds().A.X != 15 {
		t.Errorf("GfGrowRel 1.5x: child A.X = %d, want 15 (10 * 1.5)", child.Bounds().A.X)
	}
	if child.Bounds().Width() != 30 {
		t.Errorf("GfGrowRel 1.5x: child width = %d, want 30 (20 * 1.5)", child.Bounds().Width())
	}
}

// Test: GrowMode=0 child unchanged
// Spec: "A child with GrowMode=0 is not moved or resized"
func TestGrowModeZeroUnchanged(t *testing.T) {
	// Confirming test: child with GrowMode=0 is not affected by owner resize
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 10))
	child.SetGrowMode(0)
	g.Insert(child)

	// Owner grows significantly
	g.SetBounds(NewRect(0, 0, 100, 100))

	// Child should remain exactly as it was
	if child.Bounds().A.X != 5 || child.Bounds().A.Y != 5 {
		t.Errorf("GrowMode=0: child position = (%d,%d), want (5,5)", child.Bounds().A.X, child.Bounds().A.Y)
	}
	if child.Bounds().Width() != 10 || child.Bounds().Height() != 10 {
		t.Errorf("GrowMode=0: child size = (%d,%d), want (10,10)", child.Bounds().Width(), child.Bounds().Height())
	}
}

// Falsifying test: ensure GrowMode=0 is really not affected
func TestGrowModeZeroUnchangedOnShrink(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 100, 100))
	child := newMockView(NewRect(50, 50, 10, 10))
	child.SetGrowMode(0)
	g.Insert(child)

	// Owner shrinks significantly
	g.SetBounds(NewRect(0, 0, 40, 20))

	// Child should remain exactly as it was
	if child.Bounds().A.X != 50 || child.Bounds().A.Y != 50 {
		t.Errorf("GrowMode=0 on shrink: child position = (%d,%d), want (50,50)", child.Bounds().A.X, child.Bounds().A.Y)
	}
	if child.Bounds().Width() != 10 || child.Bounds().Height() != 10 {
		t.Errorf("GrowMode=0 on shrink: child size = (%d,%d), want (10,10)", child.Bounds().Width(), child.Bounds().Height())
	}
}

// Test: Zero delta means no child bounds modified
// Spec: "When the delta is zero (same width and height), no child bounds are modified"
func TestZeroDeltaNoChildBoundsModified(t *testing.T) {
	// Confirming test: when owner bounds don't change, children are not modified
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 10))
	child.SetGrowMode(GfGrowAll)
	g.Insert(child)

	originalBounds := child.Bounds()

	// Call SetBounds with same dimensions (different position, same size)
	g.SetBounds(NewRect(10, 10, 40, 20))

	// Child bounds should not change
	if child.Bounds() != originalBounds {
		t.Errorf("Zero delta: child bounds changed from %v to %v", originalBounds, child.Bounds())
	}
}

// Falsifying test: ensure zero delta truly means zero
func TestZeroDeltaNoChildBoundsModifiedMultipleChildren(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 40, 20))
	child1 := newMockView(NewRect(0, 0, 10, 5))
	child1.SetGrowMode(GfGrowLoX)
	child2 := newMockView(NewRect(10, 0, 10, 5))
	child2.SetGrowMode(GfGrowHiX)
	g.Insert(child1)
	g.Insert(child2)

	bounds1Before := child1.Bounds()
	bounds2Before := child2.Bounds()

	// Call SetBounds with same size, different position
	g.SetBounds(NewRect(100, 100, 40, 20))

	// Neither child should move
	if child1.Bounds() != bounds1Before {
		t.Errorf("Zero delta child1: bounds changed from %v to %v", bounds1Before, child1.Bounds())
	}
	if child2.Bounds() != bounds2Before {
		t.Errorf("Zero delta child2: bounds changed from %v to %v", bounds2Before, child2.Bounds())
	}
}

// Test: Cascade is recursive through Window
// Spec: "The cascade is recursive: if a child is a Container (Window, Dialog), and its SetBounds triggers its internal group's SetBounds, that group cascades to its own children"
// Spec: "When Window.SetBounds is called (it calls group.SetBounds), the GrowMode cascade propagates to widgets inside the window"
func TestCascadeRecursiveThroughWindow(t *testing.T) {
	// Desktop → Window (GfGrowHiX|GfGrowHiY) → Widget (GfGrowHiX)
	// When Desktop grows, Window stretches, Window's internal group grows, Widget stretches.
	desktop := NewDesktop(NewRect(0, 0, 100, 50))
	win := NewWindow(NewRect(10, 10, 40, 20), "Test")
	win.SetGrowMode(GfGrowHiX | GfGrowHiY) // Window stretches with desktop
	desktop.Insert(win)

	widget := newMockView(NewRect(2, 2, 10, 5))
	widget.SetGrowMode(GfGrowHiX)
	win.Insert(widget)

	// Desktop grows from 100x50 to 120x60 (deltaW=20, deltaH=10)
	desktop.SetBounds(NewRect(0, 0, 120, 60))

	// Window: GfGrowHiX|GfGrowHiY → right edge shifts +20, bottom edge shifts +10
	// Window.A stays (10,10), Window.B goes from (50,30) to (70,40)
	// Window size: 60x30 (was 40x20)
	winBounds := win.Bounds()
	if winBounds.A.X != 10 || winBounds.A.Y != 10 {
		t.Errorf("Window cascade: position = (%d,%d), want (10,10)", winBounds.A.X, winBounds.A.Y)
	}
	if winBounds.Width() != 60 || winBounds.Height() != 30 {
		t.Errorf("Window cascade: size = (%dx%d), want (60x30)", winBounds.Width(), winBounds.Height())
	}

	// Window's internal group: Window.SetBounds recalculates client area
	// Old client area: NewRect(0,0,38,18) (40-2, 20-2)
	// New client area: NewRect(0,0,58,28) (60-2, 30-2)
	// Client area deltaW=20, deltaH=10
	// Widget with GfGrowHiX: right edge shifts +20
	// Widget was (2,2,10,5) → B.X shifts from 12 to 32 → width becomes 30
	if widget.Bounds().Width() != 30 {
		t.Errorf("Widget cascade through window: width = %d, want 30 (10 + 20)", widget.Bounds().Width())
	}
	if widget.Bounds().A.X != 2 || widget.Bounds().A.Y != 2 {
		t.Errorf("Widget cascade through window: position = (%d,%d), want (2,2)", widget.Bounds().A.X, widget.Bounds().A.Y)
	}
}

// Test: Cascade is recursive through Desktop
// Spec: "When Desktop.SetBounds is called (it calls group.SetBounds), the GrowMode cascade propagates to windows inside the desktop"
func TestCascadeRecursiveThroughDesktop(t *testing.T) {
	desktop := NewDesktop(NewRect(0, 0, 80, 40))
	win := NewWindow(NewRect(10, 5, 30, 15), "Test")
	win.SetGrowMode(GfGrowLoX | GfGrowLoY) // Window shifts with desktop growth
	desktop.Insert(win)

	// Desktop grows from 80x40 to 120x50 (deltaW=40, deltaH=10)
	desktop.SetBounds(NewRect(0, 0, 120, 50))

	// Window with GfGrowLoX|GfGrowLoY: left edge shifts +40, top edge shifts +10
	// Position: (10+40, 5+10) = (50, 15), size unchanged (30x15)
	winBounds := win.Bounds()
	if winBounds.A.X != 50 || winBounds.A.Y != 15 {
		t.Errorf("Window in desktop: position = (%d,%d), want (50,15)", winBounds.A.X, winBounds.A.Y)
	}
	if winBounds.Width() != 30 || winBounds.Height() != 15 {
		t.Errorf("Window in desktop: size = (%d,%d), want (30,15)", winBounds.Width(), winBounds.Height())
	}
}

// Test: Multiple children with different GrowModes cascade independently
func TestMultipleChildrenDifferentGrowModesCascadeIndependently(t *testing.T) {
	// Confirming test: multiple children with different GrowModes all cascade correctly
	g := NewGroup(NewRect(0, 0, 40, 20))

	child1 := newMockView(NewRect(0, 0, 10, 10))
	child1.SetGrowMode(GfGrowLoX)

	child2 := newMockView(NewRect(10, 0, 10, 10))
	child2.SetGrowMode(GfGrowHiX)

	child3 := newMockView(NewRect(0, 10, 10, 10))
	child3.SetGrowMode(GfGrowLoY)

	child4 := newMockView(NewRect(10, 10, 10, 10))
	child4.SetGrowMode(GfGrowHiY)

	g.Insert(child1)
	g.Insert(child2)
	g.Insert(child3)
	g.Insert(child4)

	// Owner grows by deltaW=20, deltaH=10
	g.SetBounds(NewRect(0, 0, 60, 30))

	// child1: GfGrowLoX, A.X should shift by 20
	if child1.Bounds().A.X != 20 {
		t.Errorf("child1 (GfGrowLoX): A.X = %d, want 20", child1.Bounds().A.X)
	}

	// child2: GfGrowHiX, B.X should shift by 20. Original B.X=20 (10+10), new B.X=40 (20+20)
	if child2.Bounds().B.X != 40 {
		t.Errorf("child2 (GfGrowHiX): B.X = %d, want 40 (20 + 20)", child2.Bounds().B.X)
	}

	// child3: GfGrowLoY, A.Y should shift by 10
	if child3.Bounds().A.Y != 20 {
		t.Errorf("child3 (GfGrowLoY): A.Y = %d, want 20", child3.Bounds().A.Y)
	}

	// child4: GfGrowHiY, B.Y should shift by 10, making it taller
	if child4.Bounds().B.Y != 30 {
		t.Errorf("child4 (GfGrowHiY): B.Y = %d, want 30 (20 + 10)", child4.Bounds().B.Y)
	}
}

// Test: Child with negative growth (owner shrinks)
func TestNegativeDeltaShrinkWithGrowMode(t *testing.T) {
	// Confirming test: GrowMode flags work with negative delta (shrinking owner)
	g := NewGroup(NewRect(0, 0, 80, 40))
	child := newMockView(NewRect(10, 10, 20, 15))
	child.SetGrowMode(GfGrowAll)
	g.Insert(child)

	// Owner shrinks by deltaW=-20, deltaH=-10 (80→60, 40→30)
	g.SetBounds(NewRect(0, 0, 60, 30))

	// Child should shift back
	if child.Bounds().A.X != -10 {
		t.Errorf("GfGrowAll negative delta: child A.X = %d, want -10 (10 - 20)", child.Bounds().A.X)
	}
	if child.Bounds().A.Y != 0 {
		t.Errorf("GfGrowAll negative delta: child A.Y = %d, want 0 (10 - 10)", child.Bounds().A.Y)
	}
	// Size should remain the same
	if child.Bounds().Width() != 20 || child.Bounds().Height() != 15 {
		t.Errorf("GfGrowAll negative delta: size = (%d,%d), want (20,15)", child.Bounds().Width(), child.Bounds().Height())
	}
}

// Test: Combined flags: GfGrowLoX | GfGrowHiY
func TestCombinedFlagsGfGrowLoXHiY(t *testing.T) {
	// Confirming test: GfGrowLoX | GfGrowHiY shifts left and stretches down
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(5, 5, 10, 10))
	child.SetGrowMode(GfGrowLoX | GfGrowHiY)
	g.Insert(child)

	// Owner grows by deltaW=20, deltaH=10
	g.SetBounds(NewRect(0, 0, 60, 30))

	// Left edge shifts by deltaW
	if child.Bounds().A.X != 25 {
		t.Errorf("GfGrowLoX|GfGrowHiY: A.X = %d, want 25", child.Bounds().A.X)
	}
	// Bottom edge shifts by deltaH
	if child.Bounds().B.Y != 25 {
		t.Errorf("GfGrowLoX|GfGrowHiY: B.Y = %d, want 25", child.Bounds().B.Y)
	}
	// Top edge should not move
	if child.Bounds().A.Y != 5 {
		t.Errorf("GfGrowLoX|GfGrowHiY: A.Y = %d, want 5 (unchanged)", child.Bounds().A.Y)
	}
	// Width should remain unchanged
	if child.Bounds().Width() != 10 {
		t.Errorf("GfGrowLoX|GfGrowHiY: width = %d, want 10 (unchanged)", child.Bounds().Width())
	}
}

// Test: Group with no children
func TestGroupWithNoChildrenSetBounds(t *testing.T) {
	// Confirming test: Group.SetBounds works safely with no children
	g := NewGroup(NewRect(0, 0, 40, 20))

	// Should not panic
	g.SetBounds(NewRect(0, 0, 80, 40))

	// Group bounds should be updated
	if g.Bounds().Width() != 80 || g.Bounds().Height() != 40 {
		t.Errorf("Empty group: bounds = %v, want width=80, height=40", g.Bounds())
	}
}

// Test: Child at boundary of owner
func TestChildAtBoundaryGrowMode(t *testing.T) {
	// Confirming test: child at the exact boundary of owner
	g := NewGroup(NewRect(0, 0, 40, 20))
	child := newMockView(NewRect(30, 15, 10, 5))
	child.SetGrowMode(GfGrowHiX | GfGrowHiY)
	g.Insert(child)

	// Owner grows from 40x20 to 60x30 (deltaW=20, deltaH=10)
	g.SetBounds(NewRect(0, 0, 60, 30))

	// Child spans from 30 to 40 (width 10), should span from 30 to 60 after growth
	if child.Bounds().B.X != 60 {
		t.Errorf("Child at boundary: B.X = %d, want 60 (40 + 20)", child.Bounds().B.X)
	}
	if child.Bounds().B.Y != 30 {
		t.Errorf("Child at boundary: B.Y = %d, want 30 (20 + 10)", child.Bounds().B.Y)
	}
}

// Test: Very small child with GfGrowRel
func TestTinyChildGfGrowRel(t *testing.T) {
	// Confirming test: 1x1 child scales proportionally
	g := NewGroup(NewRect(0, 0, 10, 10))
	child := newMockView(NewRect(5, 5, 1, 1))
	child.SetGrowMode(GfGrowRel)
	g.Insert(child)

	// Owner becomes 2x
	g.SetBounds(NewRect(0, 0, 20, 20))

	// 5*2=10, 5*2=10, 1*2=2, 1*2=2
	if child.Bounds().A.X != 10 || child.Bounds().A.Y != 10 {
		t.Errorf("Tiny GfGrowRel: position = (%d,%d), want (10,10)", child.Bounds().A.X, child.Bounds().A.Y)
	}
	if child.Bounds().Width() != 2 || child.Bounds().Height() != 2 {
		t.Errorf("Tiny GfGrowRel: size = (%d,%d), want (2,2)", child.Bounds().Width(), child.Bounds().Height())
	}
}

// Test: Child larger than owner
func TestChildLargerThanOwner(t *testing.T) {
	// Confirming test: child can be larger than its owner (no clipping in cascade logic)
	g := NewGroup(NewRect(0, 0, 20, 20))
	child := newMockView(NewRect(0, 0, 100, 100))
	child.SetGrowMode(GfGrowHiX | GfGrowHiY)
	g.Insert(child)

	// Owner grows
	g.SetBounds(NewRect(0, 0, 40, 40))

	// Child should stretch further
	if child.Bounds().Width() != 120 {
		t.Errorf("Child larger than owner: width = %d, want 120 (100 + 20)", child.Bounds().Width())
	}
	if child.Bounds().Height() != 120 {
		t.Errorf("Child larger than owner: height = %d, want 120 (100 + 20)", child.Bounds().Height())
	}
}
