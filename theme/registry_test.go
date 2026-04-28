package theme

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestRegisterStoresScheme confirms Register stores a scheme by name.
func TestRegisterStoresScheme(t *testing.T) {
	// Clear any existing registration (if possible, or use a unique name)
	scheme := &ColorScheme{
		WindowBackground: tcell.StyleDefault.Foreground(tcell.ColorRed),
	}

	Register("test-scheme-1", scheme)

	// Verify we can get it back
	retrieved := Get("test-scheme-1")
	if retrieved == nil {
		t.Error("Register did not store scheme: Get returned nil")
	}
	if retrieved != scheme {
		t.Error("Register did not store the exact same scheme")
	}
}

// TestGetRetrievesScheme confirms Get retrieves a registered scheme by name.
func TestGetRetrievesScheme(t *testing.T) {
	scheme := &ColorScheme{
		ButtonNormal: tcell.StyleDefault.Foreground(tcell.ColorGreen),
	}

	Register("test-retrieve-1", scheme)
	retrieved := Get("test-retrieve-1")

	if retrieved == nil {
		t.Error("Get returned nil for a registered scheme")
	}
	if retrieved != scheme {
		t.Error("Get did not return the registered scheme")
	}
}

// TestGetUnknownNameReturnsNil confirms Get returns nil for unknown name.
func TestGetUnknownNameReturnsNil(t *testing.T) {
	// Use a name that is very unlikely to be registered
	result := Get("nonexistent-scheme-xyz-123456")

	if result != nil {
		t.Error("Get did not return nil for unknown scheme name")
	}
}

// TestRegisterDuplicateNameOverwrites confirms registering with duplicate name overwrites.
func TestRegisterDuplicateNameOverwrites(t *testing.T) {
	scheme1 := &ColorScheme{
		WindowBackground: tcell.StyleDefault.Foreground(tcell.ColorRed),
	}
	scheme2 := &ColorScheme{
		WindowBackground: tcell.StyleDefault.Foreground(tcell.ColorBlue),
	}

	Register("test-overwrite-1", scheme1)
	registered1 := Get("test-overwrite-1")

	if registered1 != scheme1 {
		t.Error("First registration did not store correctly")
	}

	// Register again with the same name
	Register("test-overwrite-1", scheme2)
	registered2 := Get("test-overwrite-1")

	if registered2 != scheme2 {
		t.Error("Duplicate registration did not overwrite the previous scheme")
	}

	// Verify it's the new one, not the old one
	if registered2 == scheme1 {
		t.Error("Get returned the old scheme instead of the new one after overwrite")
	}
}

// TestRegisterMultipleSchemesIndependent confirms multiple schemes can be registered independently.
func TestRegisterMultipleSchemesIndependent(t *testing.T) {
	scheme1 := &ColorScheme{
		WindowBackground: tcell.StyleDefault.Foreground(tcell.ColorRed),
	}
	scheme2 := &ColorScheme{
		WindowBackground: tcell.StyleDefault.Foreground(tcell.ColorGreen),
	}
	scheme3 := &ColorScheme{
		WindowBackground: tcell.StyleDefault.Foreground(tcell.ColorBlue),
	}

	Register("test-multi-1", scheme1)
	Register("test-multi-2", scheme2)
	Register("test-multi-3", scheme3)

	retrieved1 := Get("test-multi-1")
	retrieved2 := Get("test-multi-2")
	retrieved3 := Get("test-multi-3")

	if retrieved1 != scheme1 {
		t.Error("First scheme was not retrieved correctly")
	}
	if retrieved2 != scheme2 {
		t.Error("Second scheme was not retrieved correctly")
	}
	if retrieved3 != scheme3 {
		t.Error("Third scheme was not retrieved correctly")
	}
}

// TestRegisterWithDifferentNames confirms different names are keyed separately.
func TestRegisterWithDifferentNames(t *testing.T) {
	scheme := &ColorScheme{
		WindowBackground: tcell.StyleDefault.Foreground(tcell.ColorYellow),
	}

	Register("test-name-1", scheme)
	Register("test-name-2", scheme) // Same scheme, different name

	retrieved1 := Get("test-name-1")
	retrieved2 := Get("test-name-2")

	if retrieved1 != scheme {
		t.Error("First name did not retrieve the scheme")
	}
	if retrieved2 != scheme {
		t.Error("Second name did not retrieve the scheme")
	}

	// Both should return the same scheme object
	if retrieved1 != retrieved2 {
		t.Error("Different names should retrieve the same scheme")
	}
}

// TestGetIsCase-sensitive confirms Get is case-sensitive for names.
func TestGetIsCaseSensitive(t *testing.T) {
	scheme := &ColorScheme{
		WindowBackground: tcell.StyleDefault,
	}

	Register("TestCaseScheme", scheme)

	// Try retrieving with different cases
	if Get("TestCaseScheme") == nil {
		t.Error("Get failed with exact case match")
	}
	if Get("testcasescheme") != nil {
		t.Error("Get should be case-sensitive but found scheme with wrong case")
	}
	if Get("TESTCASESCHEME") != nil {
		t.Error("Get should be case-sensitive but found scheme with wrong case")
	}
}

// TestRegisterEmptyName confirms Register works with empty name.
func TestRegisterEmptyName(t *testing.T) {
	scheme := &ColorScheme{
		WindowBackground: tcell.StyleDefault,
	}

	Register("", scheme)
	retrieved := Get("")

	if retrieved != scheme {
		t.Error("Register/Get does not work with empty string name")
	}
}

// TestRegisterNilScheme confirms Register can store nil scheme.
func TestRegisterNilScheme(t *testing.T) {
	Register("test-nil-scheme", nil)
	retrieved := Get("test-nil-scheme")

	if retrieved != nil {
		t.Error("Register should be able to store nil scheme")
	}
}

// TestRegisterSameInstanceMultipleTimes confirms re-registering same instance works.
func TestRegisterSameInstanceMultipleTimes(t *testing.T) {
	scheme := &ColorScheme{
		WindowBackground: tcell.StyleDefault.Foreground(tcell.ColorPurple),
	}

	Register("test-same-1", scheme)
	Register("test-same-1", scheme) // Register the same instance again

	retrieved := Get("test-same-1")

	if retrieved != scheme {
		t.Error("Re-registering same instance should work")
	}
}

// TestGetAfterRegister confirms Get always retrieves after Register.
func TestGetAfterRegister(t *testing.T) {
	scheme := &ColorScheme{
		MenuNormal: tcell.StyleDefault.Foreground(tcell.ColorAqua),
	}

	Register("test-after-1", scheme)
	retrieved := Get("test-after-1")

	if retrieved == nil {
		t.Error("Get failed immediately after Register")
	}
	if retrieved != scheme {
		t.Error("Get did not return the registered scheme immediately after Register")
	}
}

// TestMultipleRegistrationsConcurrency confirms registry state is consistent.
func TestMultipleRegistrationsConcurrency(t *testing.T) {
	// Sequential registrations
	schemes := make([]*ColorScheme, 5)
	names := []string{"test-seq-0", "test-seq-1", "test-seq-2", "test-seq-3", "test-seq-4"}

	for i := range schemes {
		schemes[i] = &ColorScheme{
			StatusNormal: tcell.StyleDefault.Foreground(tcell.Color(i)),
		}
		Register(names[i], schemes[i])
	}

	// Verify all are retrievable
	for i := range schemes {
		retrieved := Get(names[i])
		if retrieved != schemes[i] {
			t.Errorf("Scheme at index %d was not retrieved correctly", i)
		}
	}
}

// TestRegisterNamePersistence confirms registered names persist.
func TestRegisterNamePersistence(t *testing.T) {
	scheme := &ColorScheme{
		DialogBackground: tcell.StyleDefault.Foreground(tcell.ColorWhite),
	}

	Register("test-persist-1", scheme)

	// Get the scheme multiple times
	for i := 0; i < 3; i++ {
		retrieved := Get("test-persist-1")
		if retrieved != scheme {
			t.Errorf("Registration did not persist on retrieval %d", i+1)
		}
	}
}
