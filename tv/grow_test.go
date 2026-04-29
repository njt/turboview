package tv

import (
	"testing"
)

// TestGrowFlagConstants verifies the spec requirement:
// "GrowFlag constants: GfGrowLoX through GfGrowHiY as bit flags, GfGrowAll combines all four, GfGrowRel is a separate bit"
// The spec shows:
// const (
//     GfGrowLoX  GrowFlag = 1 << iota  // left edge tracks owner's right edge
//     GfGrowLoY                         // top edge tracks owner's bottom edge
//     GfGrowHiX                         // right edge tracks owner's right edge
//     GfGrowHiY                         // bottom edge tracks owner's bottom edge
//     GfGrowAll  = GfGrowLoX | GfGrowLoY | GfGrowHiX | GfGrowHiY
//     GfGrowRel                         // edges move proportionally to owner's size change
// )
func TestGrowFlagIsBitFlag(t *testing.T) {
	// GfGrowLoX should be 1 (1 << 0)
	if GfGrowLoX != 1 {
		t.Errorf("GfGrowLoX = %d, want 1 (1 << 0)", GfGrowLoX)
	}
	// GfGrowLoY should be 2 (1 << 1)
	if GfGrowLoY != 2 {
		t.Errorf("GfGrowLoY = %d, want 2 (1 << 1)", GfGrowLoY)
	}
	// GfGrowHiX should be 4 (1 << 2)
	if GfGrowHiX != 4 {
		t.Errorf("GfGrowHiX = %d, want 4 (1 << 2)", GfGrowHiX)
	}
	// GfGrowHiY should be 8 (1 << 3)
	if GfGrowHiY != 8 {
		t.Errorf("GfGrowHiY = %d, want 8 (1 << 3)", GfGrowHiY)
	}
}

// TestGrowFlagAllCombinesFour verifies: "GfGrowAll combines all four"
func TestGrowFlagAllCombinesFour(t *testing.T) {
	expected := GfGrowLoX | GfGrowLoY | GfGrowHiX | GfGrowHiY
	if GfGrowAll != expected {
		t.Errorf("GfGrowAll = %d, want %d (GfGrowLoX | GfGrowLoY | GfGrowHiX | GfGrowHiY)", GfGrowAll, expected)
	}
}

// TestGrowFlagAllContainsEachFlag verifies GfGrowAll includes each individual flag
func TestGrowFlagAllContainsEachFlag(t *testing.T) {
	flags := []struct {
		name string
		flag GrowFlag
	}{
		{"GfGrowLoX", GfGrowLoX},
		{"GfGrowLoY", GfGrowLoY},
		{"GfGrowHiX", GfGrowHiX},
		{"GfGrowHiY", GfGrowHiY},
	}

	for _, f := range flags {
		if (GfGrowAll & f.flag) == 0 {
			t.Errorf("GfGrowAll does not contain %s", f.name)
		}
	}
}

// TestGrowFlagRelIsSeparateBit verifies: "GfGrowRel is a separate bit"
func TestGrowFlagRelIsSeparateBit(t *testing.T) {
	// GfGrowRel should be 16 (1 << 4), a separate bit after GfGrowHiY
	if GfGrowRel != 16 {
		t.Errorf("GfGrowRel = %d, want 16 (1 << 4)", GfGrowRel)
	}
}

// TestGrowFlagRelIsIndependent verifies GfGrowRel is independent from GfGrowAll
func TestGrowFlagRelIsIndependent(t *testing.T) {
	if (GfGrowRel & GfGrowAll) != 0 {
		t.Error("GfGrowRel should be independent from GfGrowAll (no shared bits)")
	}
}

// TestGrowFlagBitwise verifies flags support bitwise operations
func TestGrowFlagBitwise(t *testing.T) {
	// Test OR operation
	combined := GfGrowLoX | GfGrowHiX
	if (combined & GfGrowLoX) == 0 {
		t.Error("Bitwise OR didn't set GfGrowLoX")
	}
	if (combined & GfGrowHiX) == 0 {
		t.Error("Bitwise OR didn't set GfGrowHiX")
	}

	// Test AND operation
	if (combined & GfGrowLoY) != 0 {
		t.Error("AND should not set GfGrowLoY which wasn't ORed")
	}

	// Test NOT operation (clear flag)
	flags := GfGrowAll
	flags = flags &^ GfGrowLoX
	if (flags & GfGrowLoX) != 0 {
		t.Error("Clearing GfGrowLoX with &^ didn't work")
	}
	if (flags & GfGrowLoY) == 0 {
		t.Error("Clearing GfGrowLoX should not clear GfGrowLoY")
	}
}

// TestGrowFlagDistinct verifies each flag has a unique bit value
func TestGrowFlagDistinct(t *testing.T) {
	flags := []struct {
		name string
		flag GrowFlag
	}{
		{"GfGrowLoX", GfGrowLoX},
		{"GfGrowLoY", GfGrowLoY},
		{"GfGrowHiX", GfGrowHiX},
		{"GfGrowHiY", GfGrowHiY},
		{"GfGrowRel", GfGrowRel},
	}

	seen := make(map[GrowFlag]string)
	for _, f := range flags {
		if prev, ok := seen[f.flag]; ok {
			t.Errorf("Duplicate flag value: %s and %s both have value %d", prev, f.name, f.flag)
		}
		seen[f.flag] = f.name
	}
}

// TestGrowFlagPowerOfTwo verifies each flag is a power of two (single bit set)
func TestGrowFlagPowerOfTwo(t *testing.T) {
	flags := []struct {
		name string
		flag GrowFlag
	}{
		{"GfGrowLoX", GfGrowLoX},
		{"GfGrowLoY", GfGrowLoY},
		{"GfGrowHiX", GfGrowHiX},
		{"GfGrowHiY", GfGrowHiY},
		{"GfGrowRel", GfGrowRel},
	}

	for _, f := range flags {
		// A power of two has exactly one bit set.
		// Check using: (n & (n-1)) == 0 and n != 0
		if f.flag == 0 || (f.flag&(f.flag-1)) != 0 {
			t.Errorf("%s = %d is not a power of two", f.name, f.flag)
		}
	}
}

// TestGrowFlagAllValue verifies GfGrowAll equals 15 (1|2|4|8)
func TestGrowFlagAllValue(t *testing.T) {
	expectedAll := GrowFlag(15)
	if GfGrowAll != expectedAll {
		t.Errorf("GfGrowAll = %d, want 15", GfGrowAll)
	}
}

// TestGrowFlagAllDoesNotIncludeRel verifies GfGrowAll is 15 and GfGrowRel is 16
func TestGrowFlagAllDoesNotIncludeRel(t *testing.T) {
	// GfGrowAll should be 15 (binary 01111)
	// GfGrowRel should be 16 (binary 10000)
	// Together they should be 31 (binary 11111)
	both := GfGrowAll | GfGrowRel
	if both != 31 {
		t.Errorf("GfGrowAll | GfGrowRel = %d, want 31", both)
	}
}

// TestGrowFlagCanCombineFlags verifies flags can be combined with OR
func TestGrowFlagCanCombineFlags(t *testing.T) {
	combo := GfGrowLoX | GfGrowHiY | GfGrowRel
	// Verify each component is present
	if (combo & GfGrowLoX) == 0 {
		t.Error("Combined flag missing GfGrowLoX")
	}
	if (combo & GfGrowHiY) == 0 {
		t.Error("Combined flag missing GfGrowHiY")
	}
	if (combo & GfGrowRel) == 0 {
		t.Error("Combined flag missing GfGrowRel")
	}
	if (combo & GfGrowLoY) != 0 {
		t.Error("Combined flag should not have GfGrowLoY")
	}
}

// TestGrowFlagIntegerType verifies GrowFlag is an integer type
func TestGrowFlagIntegerType(t *testing.T) {
	// This test verifies we can perform integer operations on GrowFlag
	var g GrowFlag = GfGrowLoX
	_ = g | GfGrowLoY // Should compile and work
	result := g | GfGrowLoY
	if result != 3 { // 1 | 2 = 3
		t.Errorf("GfGrowLoX | GfGrowLoY = %d, want 3", result)
	}
}

// TestGrowFlagZeroValue verifies the zero value is not a valid flag
func TestGrowFlagZeroValue(t *testing.T) {
	var g GrowFlag = 0
	// Zero should not match any individual flag
	if g == GfGrowLoX || g == GfGrowLoY || g == GfGrowHiX || g == GfGrowHiY || g == GfGrowRel {
		t.Error("Zero value should not be a valid flag")
	}
}

// TestGrowFlagComparable verifies GrowFlag values can be compared
func TestGrowFlagComparable(t *testing.T) {
	if !(GfGrowLoX < GfGrowLoY) {
		t.Error("GfGrowLoX < GfGrowLoY failed")
	}
	if !(GfGrowHiY < GfGrowRel) {
		t.Error("GfGrowHiY < GfGrowRel failed")
	}
	if GfGrowLoX == GfGrowLoY {
		t.Error("GfGrowLoX should not equal GfGrowLoY")
	}
}
