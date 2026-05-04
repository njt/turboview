package theme

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// Helper: countNonZeroFields counts the number of non-zero tcell.Style fields in a ColorScheme.
func countNonZeroFields(cs *ColorScheme) int {
	count := 0
	v := reflect.ValueOf(cs).Elem()
	zeroStyle := tcell.StyleDefault

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Interface().(tcell.Style) != zeroStyle {
			count++
		}
	}
	return count
}

// TestBorlandCyanIsExported confirms BorlandCyan is an exported *ColorScheme.
func TestBorlandCyanIsExported(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}

	var scheme *ColorScheme = BorlandCyan
	if scheme == nil {
		t.Error("BorlandCyan cannot be assigned to *ColorScheme")
	}
}

// TestBorlandCyanAllFieldsSet confirms all 29 ColorScheme fields in BorlandCyan are non-zero.
func TestBorlandCyanAllFieldsSet(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}

	nonZeroCount := countNonZeroFields(BorlandCyan)
	if nonZeroCount != 41 {
		t.Errorf("BorlandCyan: expected 41 non-zero fields, got %d", nonZeroCount)
	}
}

// TestBorlandCyanRegistered confirms BorlandCyan is registered as "borland-cyan".
func TestBorlandCyanRegistered(t *testing.T) {
	retrieved := Get("borland-cyan")

	if retrieved == nil {
		t.Error("BorlandCyan is not registered")
	}

	if retrieved != BorlandCyan {
		t.Error("Get(\"borland-cyan\") did not return BorlandCyan")
	}
}

// TestBorlandGrayIsExported confirms BorlandGray is an exported *ColorScheme.
func TestBorlandGrayIsExported(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}

	var scheme *ColorScheme = BorlandGray
	if scheme == nil {
		t.Error("BorlandGray cannot be assigned to *ColorScheme")
	}
}

// TestBorlandGrayAllFieldsSet confirms all 29 ColorScheme fields in BorlandGray are non-zero.
func TestBorlandGrayAllFieldsSet(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}

	nonZeroCount := countNonZeroFields(BorlandGray)
	if nonZeroCount != 41 {
		t.Errorf("BorlandGray: expected 41 non-zero fields, got %d", nonZeroCount)
	}
}

// TestBorlandGrayRegistered confirms BorlandGray is registered as "borland-gray".
func TestBorlandGrayRegistered(t *testing.T) {
	retrieved := Get("borland-gray")

	if retrieved == nil {
		t.Error("BorlandGray is not registered")
	}

	if retrieved != BorlandGray {
		t.Error("Get(\"borland-gray\") did not return BorlandGray")
	}
}

// TestMatrixIsExported confirms Matrix is an exported *ColorScheme.
func TestMatrixIsExported(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}

	var scheme *ColorScheme = Matrix
	if scheme == nil {
		t.Error("Matrix cannot be assigned to *ColorScheme")
	}
}

// TestMatrixAllFieldsSet confirms all 29 ColorScheme fields in Matrix are non-zero.
func TestMatrixAllFieldsSet(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}

	nonZeroCount := countNonZeroFields(Matrix)
	if nonZeroCount != 41 {
		t.Errorf("Matrix: expected 41 non-zero fields, got %d", nonZeroCount)
	}
}

// TestMatrixRegistered confirms Matrix is registered as "matrix".
func TestMatrixRegistered(t *testing.T) {
	retrieved := Get("matrix")

	if retrieved == nil {
		t.Error("Matrix is not registered")
	}

	if retrieved != Matrix {
		t.Error("Get(\"matrix\") did not return Matrix")
	}
}

// TestC64IsExported confirms C64 is an exported *ColorScheme.
func TestC64IsExported(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}

	var scheme *ColorScheme = C64
	if scheme == nil {
		t.Error("C64 cannot be assigned to *ColorScheme")
	}
}

// TestC64AllFieldsSet confirms all 29 ColorScheme fields in C64 are non-zero.
func TestC64AllFieldsSet(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}

	nonZeroCount := countNonZeroFields(C64)
	if nonZeroCount != 41 {
		t.Errorf("C64: expected 41 non-zero fields, got %d", nonZeroCount)
	}
}

// TestC64Registered confirms C64 is registered as "c64".
func TestC64Registered(t *testing.T) {
	retrieved := Get("c64")

	if retrieved == nil {
		t.Error("C64 is not registered")
	}

	if retrieved != C64 {
		t.Error("Get(\"c64\") did not return C64")
	}
}

// TestRegistryContains5Themes confirms the registry has exactly 5 themes.
func TestRegistryContains5Themes(t *testing.T) {
	// Count themes in registry by trying to get each expected one
	expectedThemes := []string{
		"borland-blue",
		"borland-cyan",
		"borland-gray",
		"matrix",
		"c64",
	}

	for _, name := range expectedThemes {
		if Get(name) == nil {
			t.Errorf("Expected theme \"%s\" not found in registry", name)
		}
	}

	// Verify exact count by checking for unexpected entries
	// (This is a partial check; a full verification would require registry inspection)
	nonExistentThemes := []string{
		"unknown",
		"fake-theme",
		"borland-green",
		"matrix-reloaded",
		"c64-alt",
	}

	for _, name := range nonExistentThemes {
		if Get(name) != nil {
			t.Errorf("Unexpected theme \"%s\" found in registry", name)
		}
	}
}

// TestBorlandCyanDistinctFromBorlandBlue confirms BorlandCyan differs from BorlandBlue visually.
func TestBorlandCyanDistinctFromBorlandBlue(t *testing.T) {
	if BorlandCyan == nil || BorlandBlue == nil {
		t.Fatal("BorlandCyan or BorlandBlue is nil")
	}

	// Check at least one of the key visual fields differs
	hasDistinctField := false

	// Test WindowBackground
	if BorlandCyan.WindowBackground != BorlandBlue.WindowBackground {
		hasDistinctField = true
	}

	// Test DesktopBackground
	if !hasDistinctField && BorlandCyan.DesktopBackground != BorlandBlue.DesktopBackground {
		hasDistinctField = true
	}

	// Test MenuNormal
	if !hasDistinctField && BorlandCyan.MenuNormal != BorlandBlue.MenuNormal {
		hasDistinctField = true
	}

	if !hasDistinctField {
		t.Error("BorlandCyan should be visually distinct from BorlandBlue in at least WindowBackground, DesktopBackground, or MenuNormal")
	}
}

// TestBorlandGrayDistinctFromBorlandBlue confirms BorlandGray differs from BorlandBlue visually.
func TestBorlandGrayDistinctFromBorlandBlue(t *testing.T) {
	if BorlandGray == nil || BorlandBlue == nil {
		t.Fatal("BorlandGray or BorlandBlue is nil")
	}

	// Check at least one of the key visual fields differs
	hasDistinctField := false

	// Test WindowBackground
	if BorlandGray.WindowBackground != BorlandBlue.WindowBackground {
		hasDistinctField = true
	}

	// Test DesktopBackground
	if !hasDistinctField && BorlandGray.DesktopBackground != BorlandBlue.DesktopBackground {
		hasDistinctField = true
	}

	// Test MenuNormal
	if !hasDistinctField && BorlandGray.MenuNormal != BorlandBlue.MenuNormal {
		hasDistinctField = true
	}

	if !hasDistinctField {
		t.Error("BorlandGray should be visually distinct from BorlandBlue in at least WindowBackground, DesktopBackground, or MenuNormal")
	}
}

// TestMatrixDistinctFromBorlandBlue confirms Matrix differs from BorlandBlue visually.
func TestMatrixDistinctFromBorlandBlue(t *testing.T) {
	if Matrix == nil || BorlandBlue == nil {
		t.Fatal("Matrix or BorlandBlue is nil")
	}

	// Check at least one of the key visual fields differs
	hasDistinctField := false

	// Test WindowBackground
	if Matrix.WindowBackground != BorlandBlue.WindowBackground {
		hasDistinctField = true
	}

	// Test DesktopBackground
	if !hasDistinctField && Matrix.DesktopBackground != BorlandBlue.DesktopBackground {
		hasDistinctField = true
	}

	// Test MenuNormal
	if !hasDistinctField && Matrix.MenuNormal != BorlandBlue.MenuNormal {
		hasDistinctField = true
	}

	if !hasDistinctField {
		t.Error("Matrix should be visually distinct from BorlandBlue in at least WindowBackground, DesktopBackground, or MenuNormal")
	}
}

// TestC64DistinctFromBorlandBlue confirms C64 differs from BorlandBlue visually.
func TestC64DistinctFromBorlandBlue(t *testing.T) {
	if C64 == nil || BorlandBlue == nil {
		t.Fatal("C64 or BorlandBlue is nil")
	}

	// Check at least one of the key visual fields differs
	hasDistinctField := false

	// Test WindowBackground
	if C64.WindowBackground != BorlandBlue.WindowBackground {
		hasDistinctField = true
	}

	// Test DesktopBackground
	if !hasDistinctField && C64.DesktopBackground != BorlandBlue.DesktopBackground {
		hasDistinctField = true
	}

	// Test MenuNormal
	if !hasDistinctField && C64.MenuNormal != BorlandBlue.MenuNormal {
		hasDistinctField = true
	}

	if !hasDistinctField {
		t.Error("C64 should be visually distinct from BorlandBlue in at least WindowBackground, DesktopBackground, or MenuNormal")
	}
}

// TestBorlandCyanIsDistinctFromBorlandGray confirms BorlandCyan and BorlandGray are different.
func TestBorlandCyanIsDistinctFromBorlandGray(t *testing.T) {
	if BorlandCyan == nil || BorlandGray == nil {
		t.Fatal("BorlandCyan or BorlandGray is nil")
	}

	if BorlandCyan == BorlandGray {
		t.Error("BorlandCyan and BorlandGray should be different instances")
	}

	// They should also have different color schemes in at least some fields
	hasDistinctField := false
	if BorlandCyan.WindowBackground != BorlandGray.WindowBackground {
		hasDistinctField = true
	}
	if !hasDistinctField && BorlandCyan.MenuNormal != BorlandGray.MenuNormal {
		hasDistinctField = true
	}

	if !hasDistinctField {
		t.Error("BorlandCyan and BorlandGray should have different colors in at least one field")
	}
}

// TestMatrixIsDistinctFromC64 confirms Matrix and C64 are different.
func TestMatrixIsDistinctFromC64(t *testing.T) {
	if Matrix == nil || C64 == nil {
		t.Fatal("Matrix or C64 is nil")
	}

	if Matrix == C64 {
		t.Error("Matrix and C64 should be different instances")
	}

	// They should also have different color schemes in at least some fields
	hasDistinctField := false
	if Matrix.WindowBackground != C64.WindowBackground {
		hasDistinctField = true
	}
	if !hasDistinctField && Matrix.MenuNormal != C64.MenuNormal {
		hasDistinctField = true
	}

	if !hasDistinctField {
		t.Error("Matrix and C64 should have different colors in at least one field")
	}
}

// TestBorlandCyanGetMultipleCalls confirms Get("borland-cyan") always returns the same instance.
func TestBorlandCyanGetMultipleCalls(t *testing.T) {
	retrieved1 := Get("borland-cyan")
	retrieved2 := Get("borland-cyan")

	if retrieved1 != retrieved2 {
		t.Error("Multiple Get calls for \"borland-cyan\" returned different instances")
	}

	if retrieved1 != BorlandCyan {
		t.Error("Get(\"borland-cyan\") did not return BorlandCyan")
	}
}

// TestBorlandGrayGetMultipleCalls confirms Get("borland-gray") always returns the same instance.
func TestBorlandGrayGetMultipleCalls(t *testing.T) {
	retrieved1 := Get("borland-gray")
	retrieved2 := Get("borland-gray")

	if retrieved1 != retrieved2 {
		t.Error("Multiple Get calls for \"borland-gray\" returned different instances")
	}

	if retrieved1 != BorlandGray {
		t.Error("Get(\"borland-gray\") did not return BorlandGray")
	}
}

// TestMatrixGetMultipleCalls confirms Get("matrix") always returns the same instance.
func TestMatrixGetMultipleCalls(t *testing.T) {
	retrieved1 := Get("matrix")
	retrieved2 := Get("matrix")

	if retrieved1 != retrieved2 {
		t.Error("Multiple Get calls for \"matrix\" returned different instances")
	}

	if retrieved1 != Matrix {
		t.Error("Get(\"matrix\") did not return Matrix")
	}
}

// TestC64GetMultipleCalls confirms Get("c64") always returns the same instance.
func TestC64GetMultipleCalls(t *testing.T) {
	retrieved1 := Get("c64")
	retrieved2 := Get("c64")

	if retrieved1 != retrieved2 {
		t.Error("Multiple Get calls for \"c64\" returned different instances")
	}

	if retrieved1 != C64 {
		t.Error("Get(\"c64\") did not return C64")
	}
}

// TestBorlandCyanNameCaseSensitive confirms the name is exactly "borland-cyan" (case-sensitive).
func TestBorlandCyanNameCaseSensitive(t *testing.T) {
	// Correct name should exist
	if Get("borland-cyan") == nil {
		t.Error("BorlandCyan should be retrievable as \"borland-cyan\"")
	}

	// Case variations should not work
	if Get("Borland-Cyan") != nil {
		t.Error("Name lookup should be case-sensitive")
	}

	if Get("BORLAND-CYAN") != nil {
		t.Error("Name lookup should be case-sensitive")
	}
}

// TestBorlandGrayNameCaseSensitive confirms the name is exactly "borland-gray" (case-sensitive).
func TestBorlandGrayNameCaseSensitive(t *testing.T) {
	// Correct name should exist
	if Get("borland-gray") == nil {
		t.Error("BorlandGray should be retrievable as \"borland-gray\"")
	}

	// Case variations should not work
	if Get("Borland-Gray") != nil {
		t.Error("Name lookup should be case-sensitive")
	}

	if Get("BORLAND-GRAY") != nil {
		t.Error("Name lookup should be case-sensitive")
	}
}

// TestMatrixNameCaseSensitive confirms the name is exactly "matrix" (case-sensitive).
func TestMatrixNameCaseSensitive(t *testing.T) {
	// Correct name should exist
	if Get("matrix") == nil {
		t.Error("Matrix should be retrievable as \"matrix\"")
	}

	// Case variations should not work
	if Get("Matrix") != nil {
		t.Error("Name lookup should be case-sensitive")
	}

	if Get("MATRIX") != nil {
		t.Error("Name lookup should be case-sensitive")
	}
}

// TestC64NameCaseSensitive confirms the name is exactly "c64" (case-sensitive).
func TestC64NameCaseSensitive(t *testing.T) {
	// Correct name should exist
	if Get("c64") == nil {
		t.Error("C64 should be retrievable as \"c64\"")
	}

	// Case variations should not work
	if Get("C64") != nil {
		t.Error("Name lookup should be case-sensitive")
	}

	if Get("C-64") != nil {
		t.Error("Name lookup should use exact format \"c64\"")
	}
}

// TestBorlandCyanFieldsAreNonDefault confirms BorlandCyan fields are not all StyleDefault.
func TestBorlandCyanFieldsAreNonDefault(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}

	// Verify at least one field is not StyleDefault
	hasNonDefault := false
	if BorlandCyan.WindowBackground != tcell.StyleDefault {
		hasNonDefault = true
	}
	if !hasNonDefault && BorlandCyan.MenuNormal != tcell.StyleDefault {
		hasNonDefault = true
	}

	if !hasNonDefault {
		t.Error("BorlandCyan should have at least one non-default style")
	}
}

// TestBorlandGrayFieldsAreNonDefault confirms BorlandGray fields are not all StyleDefault.
func TestBorlandGrayFieldsAreNonDefault(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}

	// Verify at least one field is not StyleDefault
	hasNonDefault := false
	if BorlandGray.WindowBackground != tcell.StyleDefault {
		hasNonDefault = true
	}
	if !hasNonDefault && BorlandGray.MenuNormal != tcell.StyleDefault {
		hasNonDefault = true
	}

	if !hasNonDefault {
		t.Error("BorlandGray should have at least one non-default style")
	}
}

// TestMatrixFieldsAreNonDefault confirms Matrix fields are not all StyleDefault.
func TestMatrixFieldsAreNonDefault(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}

	// Verify at least one field is not StyleDefault
	hasNonDefault := false
	if Matrix.WindowBackground != tcell.StyleDefault {
		hasNonDefault = true
	}
	if !hasNonDefault && Matrix.MenuNormal != tcell.StyleDefault {
		hasNonDefault = true
	}

	if !hasNonDefault {
		t.Error("Matrix should have at least one non-default style")
	}
}

// TestC64FieldsAreNonDefault confirms C64 fields are not all StyleDefault.
func TestC64FieldsAreNonDefault(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}

	// Verify at least one field is not StyleDefault
	hasNonDefault := false
	if C64.WindowBackground != tcell.StyleDefault {
		hasNonDefault = true
	}
	if !hasNonDefault && C64.MenuNormal != tcell.StyleDefault {
		hasNonDefault = true
	}

	if !hasNonDefault {
		t.Error("C64 should have at least one non-default style")
	}
}

// TestBorlandCyanInitCalled confirms BorlandCyan is initialized (init() was called).
func TestBorlandCyanInitCalled(t *testing.T) {
	if BorlandCyan == nil {
		t.Error("init() did not initialize BorlandCyan")
	}
}

// TestBorlandGrayInitCalled confirms BorlandGray is initialized (init() was called).
func TestBorlandGrayInitCalled(t *testing.T) {
	if BorlandGray == nil {
		t.Error("init() did not initialize BorlandGray")
	}
}

// TestMatrixInitCalled confirms Matrix is initialized (init() was called).
func TestMatrixInitCalled(t *testing.T) {
	if Matrix == nil {
		t.Error("init() did not initialize Matrix")
	}
}

// TestC64InitCalled confirms C64 is initialized (init() was called).
func TestC64InitCalled(t *testing.T) {
	if C64 == nil {
		t.Error("init() did not initialize C64")
	}
}
