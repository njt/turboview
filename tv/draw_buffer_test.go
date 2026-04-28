package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// === NewDrawBuffer() ===

// TestNewDrawBufferCreates80x25Buffer verifies:
// "NewDrawBuffer(80, 25) creates an 80x25 buffer"
func TestNewDrawBufferCreates80x25Buffer(t *testing.T) {
	db := NewDrawBuffer(80, 25)
	if db == nil {
		t.Fatalf("NewDrawBuffer(80, 25) should not return nil")
	}
	// Verify dimensions by checking GetCell works at boundaries
	// Cell at (79, 24) should be valid
	cell := db.GetCell(79, 24)
	if cell.Rune != ' ' {
		t.Errorf("GetCell(79, 24) should return space in 80x25 buffer, got %c", cell.Rune)
	}
}

// TestNewDrawBufferFilledWithSpaces verifies:
// "buffer filled with spaces"
func TestNewDrawBufferFilledWithSpaces(t *testing.T) {
	db := NewDrawBuffer(80, 25)
	// Test several cells
	for _, x := range []int{0, 40, 79} {
		for _, y := range []int{0, 12, 24} {
			cell := db.GetCell(x, y)
			if cell.Rune != ' ' {
				t.Errorf("GetCell(%d, %d) should have space rune, got %c", x, y, cell.Rune)
			}
		}
	}
}

// TestNewDrawBufferFilledWithDefaultStyle verifies:
// "buffer filled with tcell.StyleDefault"
func TestNewDrawBufferFilledWithDefaultStyle(t *testing.T) {
	db := NewDrawBuffer(80, 25)
	cell := db.GetCell(0, 0)
	if cell.Style != tcell.StyleDefault {
		t.Errorf("GetCell(0, 0) should have StyleDefault, got %v", cell.Style)
	}
}

// TestNewDrawBufferSmallDimensions verifies:
// NewDrawBuffer works with small dimensions
func TestNewDrawBufferSmallDimensions(t *testing.T) {
	db := NewDrawBuffer(1, 1)
	if db == nil {
		t.Fatalf("NewDrawBuffer(1, 1) should not return nil")
	}
	cell := db.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Errorf("GetCell(0, 0) should return space in 1x1 buffer, got %c", cell.Rune)
	}
}

// TestNewDrawBufferLargeDimensions verifies:
// NewDrawBuffer works with large dimensions
func TestNewDrawBufferLargeDimensions(t *testing.T) {
	db := NewDrawBuffer(200, 100)
	if db == nil {
		t.Fatalf("NewDrawBuffer(200, 100) should not return nil")
	}
	cell := db.GetCell(199, 99)
	if cell.Rune != ' ' {
		t.Errorf("GetCell(199, 99) should return space in 200x100 buffer, got %c", cell.Rune)
	}
}

// === WriteChar() ===

// TestWriteCharWritesCellAtPosition verifies:
// "WriteChar(x, y, ch, style) writes a cell at the given position"
func TestWriteCharWritesCellAtPosition(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	style := tcell.StyleDefault.Foreground(tcell.ColorRed)
	db.WriteChar(5, 3, 'A', style)

	cell := db.GetCell(5, 3)
	if cell.Rune != 'A' {
		t.Errorf("After WriteChar(5, 3, 'A', style), GetCell(5, 3) should have 'A', got %c", cell.Rune)
	}
	if cell.Style != style {
		t.Errorf("After WriteChar, GetCell(5, 3) should have matching style")
	}
}

// TestWriteCharOutsideClipIsNoOp verifies:
// "outside clip is a no-op"
func TestWriteCharOutsideClipIsNoOp(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	originalCell := db.GetCell(0, 0)
	originalRune := originalCell.Rune

	// Try to write outside bounds
	db.WriteChar(-1, 0, 'X', tcell.StyleDefault)
	db.WriteChar(10, 5, 'Y', tcell.StyleDefault)
	db.WriteChar(5, 10, 'Z', tcell.StyleDefault)

	// Cell at (0, 0) should be unchanged
	cell := db.GetCell(0, 0)
	if cell.Rune != originalRune {
		t.Errorf("WriteChar outside clip should not affect buffer")
	}
}

// TestWriteCharNegativeY verifies:
// WriteChar at y=-1 is a no-op (outside buffer bounds)
func TestWriteCharNegativeY(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	originalCell := db.GetCell(5, 0)
	originalRune := originalCell.Rune

	// Try to write at y=-1
	db.WriteChar(5, -1, 'X', tcell.StyleDefault)

	// Cell at (5, 0) should be unchanged
	cell := db.GetCell(5, 0)
	if cell.Rune != originalRune {
		t.Errorf("WriteChar at y=-1 should be a no-op")
	}
}

// TestWriteCharMultipleCells verifies:
// WriteChar can write to multiple cells sequentially
func TestWriteCharMultipleCells(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	db.WriteChar(0, 0, 'A', tcell.StyleDefault)
	db.WriteChar(1, 0, 'B', tcell.StyleDefault)
	db.WriteChar(2, 0, 'C', tcell.StyleDefault)

	if db.GetCell(0, 0).Rune != 'A' {
		t.Errorf("Expected 'A' at (0, 0)")
	}
	if db.GetCell(1, 0).Rune != 'B' {
		t.Errorf("Expected 'B' at (1, 0)")
	}
	if db.GetCell(2, 0).Rune != 'C' {
		t.Errorf("Expected 'C' at (2, 0)")
	}
}

// TestWriteCharOverwritesExisting verifies:
// WriteChar overwrites previous cell content
func TestWriteCharOverwritesExisting(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	db.WriteChar(5, 5, 'A', tcell.StyleDefault)
	if db.GetCell(5, 5).Rune != 'A' {
		t.Errorf("First write failed")
	}
	db.WriteChar(5, 5, 'B', tcell.StyleDefault)
	if db.GetCell(5, 5).Rune != 'B' {
		t.Errorf("Second write should overwrite, got %c", db.GetCell(5, 5).Rune)
	}
}

// === WriteStr() ===

// TestWriteStrWritesStringHorizontally verifies:
// "WriteStr(x, y, s, style) writes a string horizontally starting at (x, y)"
func TestWriteStrWritesStringHorizontally(t *testing.T) {
	db := NewDrawBuffer(20, 10)
	style := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	db.WriteStr(0, 0, "Hello", style)

	expected := []rune{'H', 'e', 'l', 'l', 'o'}
	for i, ch := range expected {
		cell := db.GetCell(i, 0)
		if cell.Rune != ch {
			t.Errorf("WriteStr: cell at (%d, 0) should be %c, got %c", i, ch, cell.Rune)
		}
		if cell.Style != style {
			t.Errorf("WriteStr: cell at (%d, 0) should have matching style", i)
		}
	}
}

// TestWriteStrAtArbitraryPosition verifies:
// WriteStr writes at arbitrary (x, y) positions
func TestWriteStrAtArbitraryPosition(t *testing.T) {
	db := NewDrawBuffer(20, 10)
	db.WriteStr(5, 3, "Test", tcell.StyleDefault)

	expected := []rune{'T', 'e', 's', 't'}
	for i, ch := range expected {
		cell := db.GetCell(5+i, 3)
		if cell.Rune != ch {
			t.Errorf("WriteStr at (5, 3): cell at (%d, 3) should be %c, got %c", 5+i, ch, cell.Rune)
		}
	}
}

// TestWriteStrPartiallyOutsideClip verifies:
// WriteStr clips at boundaries (partially outside is a no-op for out-of-bounds chars)
func TestWriteStrPartiallyOutsideClip(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	// Write starting at position 8 with string of length 5
	// Characters at positions 8, 9 are valid; 10, 11, 12 are outside
	db.WriteStr(8, 0, "ABCDE", tcell.StyleDefault)

	// 8, 9 should be written
	if db.GetCell(8, 0).Rune != 'A' {
		t.Errorf("WriteStr partial clip: position 8 should have 'A'")
	}
	if db.GetCell(9, 0).Rune != 'B' {
		t.Errorf("WriteStr partial clip: position 9 should have 'B'")
	}
}

// TestWriteStrEmptyString verifies:
// WriteStr with empty string does nothing
func TestWriteStrEmptyString(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	original := db.GetCell(0, 0).Rune
	db.WriteStr(0, 0, "", tcell.StyleDefault)
	if db.GetCell(0, 0).Rune != original {
		t.Errorf("WriteStr with empty string should be no-op")
	}
}

// TestWriteStrMultibyteRunes verifies:
// WriteStr handles multi-byte runes
func TestWriteStrMultibyteRunes(t *testing.T) {
	db := NewDrawBuffer(20, 10)
	db.WriteStr(0, 0, "こんにちは", tcell.StyleDefault)

	expected := []rune{'こ', 'ん', 'に', 'ち', 'は'}
	for i, ch := range expected {
		cell := db.GetCell(i, 0)
		if cell.Rune != ch {
			t.Errorf("WriteStr multibyte: cell at (%d, 0) should be %c, got %c", i, ch, cell.Rune)
		}
	}
}

// === Fill() ===

// TestFillFilledRectangle verifies:
// "Fill(rect, ch, style) fills the given rectangle"
func TestFillFilledRectangle(t *testing.T) {
	db := NewDrawBuffer(20, 20)
	rect := NewRect(2, 2, 5, 5)
	db.Fill(rect, '#', tcell.StyleDefault)

	// Check all cells in the rectangle are filled
	for x := 2; x < 7; x++ {
		for y := 2; y < 7; y++ {
			cell := db.GetCell(x, y)
			if cell.Rune != '#' {
				t.Errorf("Fill: cell at (%d, %d) should be '#', got %c", x, y, cell.Rune)
			}
		}
	}
}

// TestFillPreservesOutsideRectangle verifies:
// Fill only affects cells inside the rectangle
func TestFillPreservesOutsideRectangle(t *testing.T) {
	db := NewDrawBuffer(20, 20)
	rect := NewRect(5, 5, 3, 3)
	db.Fill(rect, '#', tcell.StyleDefault)

	// Cell outside should still be space
	cell := db.GetCell(4, 5)
	if cell.Rune != ' ' {
		t.Errorf("Fill should not affect cells outside rectangle")
	}
}

// TestFillWithDifferentCharacters verifies:
// Fill works with various characters
func TestFillWithDifferentCharacters(t *testing.T) {
	tests := []rune{'@', '*', '░', 'X'}
	for _, ch := range tests {
		db := NewDrawBuffer(10, 10)
		rect := NewRect(0, 0, 5, 5)
		db.Fill(rect, ch, tcell.StyleDefault)
		cell := db.GetCell(2, 2)
		if cell.Rune != ch {
			t.Errorf("Fill with %c failed", ch)
		}
	}
}

// TestFillWithStyle verifies:
// Fill applies the given style to all filled cells
func TestFillWithStyle(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	style := tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorYellow)
	rect := NewRect(0, 0, 3, 3)
	db.Fill(rect, 'X', style)

	cell := db.GetCell(1, 1)
	if cell.Style != style {
		t.Errorf("Fill should apply style to all cells")
	}
}

// TestFillSingleCell verifies:
// Fill with a 1x1 rectangle fills just one cell
func TestFillSingleCell(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	rect := NewRect(5, 5, 1, 1)
	db.Fill(rect, 'X', tcell.StyleDefault)

	cell := db.GetCell(5, 5)
	if cell.Rune != 'X' {
		t.Errorf("Fill 1x1 should fill single cell")
	}
	cell = db.GetCell(5, 4)
	if cell.Rune != ' ' {
		t.Errorf("Fill should not affect adjacent cell")
	}
}

// TestFillExtendingBeyondBufferEdges verifies:
// Fill with rect extending beyond buffer edges fills only valid area (no panic)
func TestFillExtendingBeyondBufferEdges(t *testing.T) {
	db := NewDrawBuffer(20, 20)
	// Rect starting at (18, 18) with size (10, 10) extends beyond 20x20 bounds
	rect := NewRect(18, 18, 10, 10)
	db.Fill(rect, '#', tcell.StyleDefault)

	// Valid cells (18, 18) and (19, 19) should be filled
	if db.GetCell(18, 18).Rune != '#' {
		t.Errorf("Fill should fill valid cell at (18, 18)")
	}
	if db.GetCell(19, 19).Rune != '#' {
		t.Errorf("Fill should fill valid cell at (19, 19)")
	}

	// Just inside the boundary should be filled
	if db.GetCell(18, 19).Rune != '#' {
		t.Errorf("Fill should fill valid cell at (18, 19)")
	}

	// Far outside boundaries should return zero-value Cell
	oob := db.GetCell(25, 25)
	if oob.Rune != 0 {
		t.Errorf("GetCell at (25, 25) should return zero Cell (outside bounds), got rune %c", oob.Rune)
	}
}

// === SubBuffer() ===

// TestSubBufferSharesBackingStore verifies:
// "SubBuffer(rect) shares the parent's backing store; writes in the child appear in the parent"
func TestSubBufferSharesBackingStore(t *testing.T) {
	parent := NewDrawBuffer(20, 20)
	childRect := NewRect(5, 5, 5, 5)
	child := parent.SubBuffer(childRect)

	// Write in child at local (0, 0) which should be parent's (5, 5)
	child.WriteChar(0, 0, 'X', tcell.StyleDefault)

	// Read from parent at (5, 5)
	parentCell := parent.GetCell(5, 5)
	if parentCell.Rune != 'X' {
		t.Errorf("Child write should appear in parent: parent(5, 5) = %c, expected 'X'", parentCell.Rune)
	}
}

// TestSubBufferLocalsCoordinateSystem verifies:
// SubBuffer has its own local coordinate system where (0, 0) maps to the allocated region
func TestSubBufferLocalsCoordinateSystem(t *testing.T) {
	parent := NewDrawBuffer(20, 20)
	childRect := NewRect(10, 10, 5, 5)
	child := parent.SubBuffer(childRect)

	// Write at local (0, 0) in child
	child.WriteChar(0, 0, 'A', tcell.StyleDefault)
	// Write at local (4, 4) in child
	child.WriteChar(4, 4, 'B', tcell.StyleDefault)

	// Read from parent at absolute coordinates
	if parent.GetCell(10, 10).Rune != 'A' {
		t.Errorf("Child(0, 0) should map to Parent(10, 10)")
	}
	if parent.GetCell(14, 14).Rune != 'B' {
		t.Errorf("Child(4, 4) should map to Parent(14, 14)")
	}
}

// TestSubBufferClipIsIntersection verifies:
// "SubBuffer clip is the intersection of the parent's clip and the requested rect"
func TestSubBufferClipIsIntersection(t *testing.T) {
	parent := NewDrawBuffer(20, 20)
	childRect := NewRect(5, 5, 10, 10)
	child := parent.SubBuffer(childRect)

	// Verify child respects the clip by trying to write outside allocated region
	// Try to write at child's local (100, 100) which is well outside
	child.WriteChar(100, 100, 'X', tcell.StyleDefault)

	// Parent should not have 'X' written anywhere
	found := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if parent.GetCell(x, y).Rune == 'X' {
				found = true
			}
		}
	}
	if found {
		t.Errorf("Child write outside clip should not affect parent")
	}
}

// TestSubBufferCannotWriteOutsideAllocatedRegion verifies:
// "A child SubBuffer cannot write outside its allocated region even if it tries"
func TestSubBufferCannotWriteOutsideAllocatedRegion(t *testing.T) {
	parent := NewDrawBuffer(20, 20)
	childRect := NewRect(5, 5, 5, 5)
	child := parent.SubBuffer(childRect)

	// Try to write at local coordinates outside the allocated 5x5 region
	child.WriteChar(5, 0, 'X', tcell.StyleDefault)  // local (5, 0) is outside
	child.WriteChar(0, 5, 'Y', tcell.StyleDefault)  // local (0, 5) is outside
	child.WriteChar(-1, 0, 'Z', tcell.StyleDefault) // local (-1, 0) is outside

	// Verify none of these writes succeeded
	if parent.GetCell(10, 5).Rune == 'X' {
		t.Errorf("Child should not write at boundary or outside allocated region")
	}
	if parent.GetCell(5, 10).Rune == 'Y' {
		t.Errorf("Child should not write at boundary or outside allocated region")
	}
	if parent.GetCell(4, 5).Rune == 'Z' {
		t.Errorf("Child should not write outside allocated region")
	}
}

// TestSubBufferGetCell verifies:
// Child's GetCell reads from the correct position
func TestSubBufferGetCell(t *testing.T) {
	parent := NewDrawBuffer(20, 20)
	childRect := NewRect(5, 5, 5, 5)
	child := parent.SubBuffer(childRect)

	// Write to parent at (5, 5)
	parent.WriteChar(5, 5, 'X', tcell.StyleDefault)

	// Read from child at local (0, 0)
	cell := child.GetCell(0, 0)
	if cell.Rune != 'X' {
		t.Errorf("Child.GetCell(0, 0) should reflect parent.GetCell(5, 5)")
	}
}

// TestNestedSubBuffers verifies:
// SubBuffer can be called on a child SubBuffer
func TestNestedSubBuffers(t *testing.T) {
	parent := NewDrawBuffer(30, 30)
	child1Rect := NewRect(5, 5, 20, 20)
	child1 := parent.SubBuffer(child1Rect)

	child2Rect := NewRect(5, 5, 10, 10) // relative to child1
	child2 := child1.SubBuffer(child2Rect)

	// Write at child2 local (0, 0)
	child2.WriteChar(0, 0, 'A', tcell.StyleDefault)

	// Should appear at parent (10, 10): parent(5+5, 5+5)
	if parent.GetCell(10, 10).Rune != 'A' {
		t.Errorf("Nested SubBuffer write should propagate to grandparent")
	}
}

// TestSubBufferWithEmptyRect verifies:
// SubBuffer with an empty rect is safe
func TestSubBufferWithEmptyRect(t *testing.T) {
	parent := NewDrawBuffer(20, 20)
	emptyRect := NewRect(5, 5, 0, 0)
	child := parent.SubBuffer(emptyRect)

	// Child should exist but not allow writes
	child.WriteChar(0, 0, 'X', tcell.StyleDefault)

	// No 'X' should appear in parent
	found := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if parent.GetCell(x, y).Rune == 'X' {
				found = true
			}
		}
	}
	if found {
		t.Errorf("SubBuffer with empty rect should not write")
	}
}

// TestSubBufferRectEntirelyOutsideParent verifies:
// SubBuffer with rect entirely outside parent bounds results in all writes being no-ops
func TestSubBufferRectEntirelyOutsideParent(t *testing.T) {
	parent := NewDrawBuffer(20, 20)
	// Rect entirely outside parent's 20x20 bounds
	outsideRect := NewRect(100, 100, 5, 5)
	child := parent.SubBuffer(outsideRect)

	// Attempt to write at various positions in child
	child.WriteChar(0, 0, 'X', tcell.StyleDefault)
	child.WriteChar(2, 2, 'Y', tcell.StyleDefault)
	child.WriteStr(0, 0, "Test", tcell.StyleDefault)
	child.Fill(NewRect(0, 0, 5, 5), 'Z', tcell.StyleDefault)

	// Verify no writes appeared in parent
	found := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			cell := parent.GetCell(x, y)
			if cell.Rune != ' ' {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if found {
		t.Errorf("SubBuffer with rect outside parent bounds should not write to parent")
	}
}

// TestSubBufferClipIntersectionVerification verifies:
// SubBuffer with rect NewRect(15, 15, 10, 10) on 20x20 parent has 5x5 effective clip
// Writes at child-local (4, 4) succeed but (5, 5) do not
func TestSubBufferClipIntersectionVerification(t *testing.T) {
	parent := NewDrawBuffer(20, 20)
	childRect := NewRect(15, 15, 10, 10)
	child := parent.SubBuffer(childRect)

	// Child-local (4, 4) maps to parent (19, 19) which is valid
	child.WriteChar(4, 4, 'A', tcell.StyleDefault)
	if parent.GetCell(19, 19).Rune != 'A' {
		t.Errorf("Child write at local (4, 4) should succeed (maps to parent 19, 19)")
	}

	// Child-local (5, 5) would map to parent (20, 20) which is outside 20x20 bounds
	child.WriteChar(5, 5, 'B', tcell.StyleDefault)
	// Verify that no 'B' appears outside the parent bounds or at wrong position
	if parent.GetCell(19, 19).Rune != 'A' {
		t.Errorf("Child write at local (5, 5) should not affect valid cells")
	}
	// Also verify 'B' doesn't appear anywhere it shouldn't
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if parent.GetCell(x, y).Rune == 'B' {
				t.Errorf("Child write at local (5, 5) should be no-op (outside effective clip)")
			}
		}
	}
}

// === GetCell() ===

// TestGetCellReadsBackCell verifies:
// "GetCell(x, y) reads back the cell at local coordinates"
func TestGetCellReadsBackCell(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	style := tcell.StyleDefault.Foreground(tcell.ColorTeal)
	db.WriteChar(3, 4, 'Z', style)

	cell := db.GetCell(3, 4)
	if cell.Rune != 'Z' {
		t.Errorf("GetCell should return written rune")
	}
	if cell.Style != style {
		t.Errorf("GetCell should return written style")
	}
}

// TestGetCellReturnsNilOutsideBounds verifies:
// GetCell returns nil for out-of-bounds coordinates
func TestGetCellReturnsZeroOutsideBounds(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	if db.GetCell(-1, 0).Rune != 0 {
		t.Errorf("GetCell(-1, 0) should return zero Cell")
	}
	if db.GetCell(10, 5).Rune != 0 {
		t.Errorf("GetCell(10, 5) should return zero Cell")
	}
	if db.GetCell(5, 10).Rune != 0 {
		t.Errorf("GetCell(5, 10) should return zero Cell")
	}
}

// TestGetCellBoundaryValues verifies:
// GetCell works at buffer boundaries
func TestGetCellBoundaryValues(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	// Top-left corner
	cell := db.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Errorf("GetCell(0, 0) should return space, got %c", cell.Rune)
	}
	// Bottom-right corner (9, 9)
	cell = db.GetCell(9, 9)
	if cell.Rune != ' ' {
		t.Errorf("GetCell(9, 9) should return space, got %c", cell.Rune)
	}
}

// === FlushTo() ===

// TestFlushToCopiesAllCellsToScreen verifies:
// "FlushTo(screen) copies all cells to the tcell screen"
func TestFlushToCopiesAllCellsToScreen(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	db.WriteStr(0, 0, "Hello", tcell.StyleDefault)

	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	defer screen.Fini()

	db.FlushTo(screen)

	// Read back from screen to verify
	r, _, _, _ := screen.GetContent(0, 0)
	if r != 'H' {
		t.Errorf("FlushTo should copy 'H' to screen(0, 0), got %c", r)
	}
}

// TestFlushToWithDifferentBufferSizes verifies:
// FlushTo works with different buffer dimensions
func TestFlushToWithDifferentBufferSizes(t *testing.T) {
	tests := []struct{ w, h int }{
		{80, 24},
		{10, 10},
		{1, 1},
	}

	for _, tt := range tests {
		db := NewDrawBuffer(tt.w, tt.h)
		db.WriteChar(0, 0, 'X', tcell.StyleDefault)

		screen := tcell.NewSimulationScreen("UTF-8")
		screen.Init()
		defer screen.Fini()

		db.FlushTo(screen)

		r, _, _, _ := screen.GetContent(0, 0)
		if r != 'X' {
			t.Errorf("FlushTo(%dx%d) failed to copy cell", tt.w, tt.h)
		}
	}
}

// TestFlushToAppliesStyles verifies:
// FlushTo applies the stored styles
func TestFlushToAppliesStyles(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	style := tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlue)
	db.WriteChar(0, 0, 'A', style)

	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	defer screen.Fini()

	db.FlushTo(screen)

	_, _, screenStyle, _ := screen.GetContent(0, 0)
	// Styles should match (may require exact comparison)
	if screenStyle != style {
		t.Errorf("FlushTo should apply correct style")
	}
}

// TestFlushToAllPositions verifies:
// FlushTo copies cells from all positions in the buffer
func TestFlushToAllPositions(t *testing.T) {
	db := NewDrawBuffer(5, 5)
	// Fill buffer with a pattern
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			ch := rune('A' + (x+y*5)%26)
			db.WriteChar(x, y, ch, tcell.StyleDefault)
		}
	}

	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	defer screen.Fini()

	db.FlushTo(screen)

	// Spot check a few positions
	r, _, _, _ := screen.GetContent(0, 0)
	if r != 'A' {
		t.Errorf("FlushTo should copy position (0, 0)")
	}
	r, _, _, _ = screen.GetContent(4, 4)
	expected := rune('A' + (4+4*5)%26)
	if r != expected {
		t.Errorf("FlushTo should copy position (4, 4)")
	}
}

// === Integration tests ===

// TestComplexDrawingScenario verifies:
// Multiple operations work together correctly
func TestComplexDrawingScenario(t *testing.T) {
	parent := NewDrawBuffer(20, 20)

	// Create a child buffer for a "panel"
	panelRect := NewRect(2, 2, 10, 10)
	panel := parent.SubBuffer(panelRect)

	// Fill the panel with a border character
	panel.Fill(NewRect(0, 0, 10, 10), '░', tcell.StyleDefault)

	// Write some text in the panel
	panel.WriteStr(1, 1, "Title", tcell.StyleDefault)

	// Verify parent has both the border and text
	if parent.GetCell(2, 2).Rune != '░' {
		t.Errorf("Parent should have border cell from child fill")
	}
	if parent.GetCell(3, 3).Rune != 'T' {
		t.Errorf("Parent should have text from child write")
	}
}

// TestSubBufferWriteStrIntegration verifies:
// WriteStr in a SubBuffer works correctly
func TestSubBufferWriteStrIntegration(t *testing.T) {
	parent := NewDrawBuffer(30, 30)
	childRect := NewRect(10, 10, 15, 10)
	child := parent.SubBuffer(childRect)

	child.WriteStr(0, 0, "Test", tcell.StyleDefault)

	// Verify in parent at absolute coordinates
	if parent.GetCell(10, 10).Rune != 'T' {
		t.Errorf("Child WriteStr should appear in parent")
	}
	if parent.GetCell(13, 10).Rune != 't' {
		t.Errorf("Child WriteStr full string should appear in parent")
	}
}
