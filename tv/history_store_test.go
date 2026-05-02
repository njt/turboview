package tv

// history_store_test.go — Tests for HistoryStore.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test has a doc comment citing the relevant spec sentence it verifies.
//
// Test organisation:
//   Section 1 — Constructor and package-level default
//   Section 2 — Add: empty string no-op
//   Section 3 — Add: dedup at tail
//   Section 4 — Add: dedup at other positions (move-to-end)
//   Section 5 — Add: eviction when exceeding maxPerID
//   Section 6 — Add: unknown ID creates new list
//   Section 7 — Entries: chronological order
//   Section 8 — Entries: unknown ID
//   Section 9 — Clear
//   Section 10 — Falsifying tests

import (
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// Section 1 — Constructor and package-level default
// ---------------------------------------------------------------------------

// TestNewHistoryStoreCreatesStore verifies the constructor returns a usable store.
// Spec: "NewHistoryStore(maxPerID int) creates a store with the given max entries per ID"
func TestNewHistoryStoreCreatesStore(t *testing.T) {
	hs := NewHistoryStore(5)
	if hs == nil {
		t.Fatal("NewHistoryStore returned nil")
	}
}

// TestDefaultHistoryExists verifies the package-level default is initialised.
// Spec: "var DefaultHistory = NewHistoryStore(20) is the package-level default"
func TestDefaultHistoryExists(t *testing.T) {
	if DefaultHistory == nil {
		t.Fatal("DefaultHistory is nil; expected a non-nil *HistoryStore")
	}
}

// TestDefaultHistoryMaxIs20 verifies the package-level default allows up to 20
// entries per ID before eviction.
// Spec: "var DefaultHistory = NewHistoryStore(20)"
func TestDefaultHistoryMaxIs20(t *testing.T) {
	DefaultHistory.Clear()
	for i := 1; i <= 20; i++ {
		DefaultHistory.Add(9999, fmt.Sprintf("entry%d", i))
	}
	entries := DefaultHistory.Entries(9999)
	if len(entries) != 20 {
		t.Fatalf("DefaultHistory should hold 20 entries, got %d", len(entries))
	}
	DefaultHistory.Add(9999, "entry21")
	entries = DefaultHistory.Entries(9999)
	if len(entries) != 20 {
		t.Fatalf("DefaultHistory should evict at 20, got %d entries", len(entries))
	}
	if entries[0] != "entry2" {
		t.Errorf("expected oldest to be evicted, first entry = %q, want %q", entries[0], "entry2")
	}
	DefaultHistory.Clear()
}

// ---------------------------------------------------------------------------
// Section 2 — Add: empty string is a no-op
// ---------------------------------------------------------------------------

// TestAddEmptyStringIsNoop verifies that adding an empty string does nothing.
// Spec: "Empty strings are never stored (no-op)"
func TestAddEmptyStringIsNoop(t *testing.T) {
	hs := NewHistoryStore(5)
	hs.Add(1, "")
	entries := hs.Entries(1)
	if len(entries) != 0 {
		t.Errorf("Add(\"\") stored an entry; Entries = %v", entries)
	}
}

// TestAddEmptyStringDoesNotCountAgainstMax verifies empty adds don't consume
// capacity so that max real entries can still be stored.
// Spec: "Empty strings are never stored (no-op)" — they must not affect count.
func TestAddEmptyStringDoesNotCountAgainstMax(t *testing.T) {
	hs := NewHistoryStore(2)
	hs.Add(1, "")
	hs.Add(1, "")
	hs.Add(1, "hello")
	hs.Add(1, "world")
	entries := hs.Entries(1)
	if len(entries) != 2 {
		t.Errorf("expected 2 entries after 2 empty + 2 real adds, got %d: %v", len(entries), entries)
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Add: dedup at tail (most recent)
// ---------------------------------------------------------------------------

// TestAddDedupAtTailNoStore verifies that adding the same string as the most
// recent entry for an ID is a no-op.
// Spec: "If the string equals the most recent entry for this ID, it is not added (dedup at tail)"
func TestAddDedupAtTailNoStore(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "alpha")
	hs.Add(1, "alpha") // duplicate of most recent
	entries := hs.Entries(1)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after duplicate tail add, got %d: %v", len(entries), entries)
	}
}

// TestAddDedupAtTailValueUnchanged verifies the stored value is the original,
// not replaced.
// Spec: "If the string equals the most recent entry for this ID, it is not added (dedup at tail)"
func TestAddDedupAtTailValueUnchanged(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "alpha")
	hs.Add(1, "alpha")
	entries := hs.Entries(1)
	if len(entries) != 1 || entries[0] != "alpha" {
		t.Errorf("expected [alpha], got %v", entries)
	}
}

// TestAddDifferentStringAfterDupIsStored verifies that a non-duplicate IS added
// even after a duplicate was suppressed.
// Spec: "If the string equals the most recent entry … it is not added" — converse.
func TestAddDifferentStringAfterDupIsStored(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "alpha")
	hs.Add(1, "alpha") // no-op
	hs.Add(1, "beta")  // must be stored
	entries := hs.Entries(1)
	if len(entries) != 2 {
		t.Errorf("expected 2 entries [alpha beta], got %v", entries)
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Add: dedup at other positions (move-to-end)
// ---------------------------------------------------------------------------

// TestAddDuplicateAtOtherPositionRemovesOld verifies that when a duplicate
// exists at a non-tail position, the older copy is removed before adding the
// new one at the end.
// Spec: "If a duplicate exists at any other position, that older entry is removed
//         before adding the new one at the end"
func TestAddDuplicateAtOtherPositionRemovesOld(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "alpha")
	hs.Add(1, "beta")
	hs.Add(1, "alpha") // duplicate of a non-tail entry
	entries := hs.Entries(1)
	// "alpha" must appear only once
	count := 0
	for _, e := range entries {
		if e == "alpha" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly one 'alpha' after move-to-end, got %d in %v", count, entries)
	}
}

// TestAddDuplicateAtOtherPositionAppearsAtEnd verifies the re-added duplicate
// is placed at the end (newest position).
// Spec: "that older entry is removed before adding the new one at the end"
func TestAddDuplicateAtOtherPositionAppearsAtEnd(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "alpha")
	hs.Add(1, "beta")
	hs.Add(1, "alpha") // should move to end
	entries := hs.Entries(1)
	if len(entries) == 0 || entries[len(entries)-1] != "alpha" {
		t.Errorf("expected 'alpha' at end after move-to-end, got %v", entries)
	}
}

// TestAddDuplicateAtOtherPositionCountUnchanged verifies total count stays the
// same (old copy removed, new copy added — net zero).
// Spec: "that older entry is removed before adding the new one at the end"
func TestAddDuplicateAtOtherPositionCountUnchanged(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "alpha")
	hs.Add(1, "beta")
	hs.Add(1, "gamma")
	hs.Add(1, "beta") // duplicate at position 1; net count should stay 3
	entries := hs.Entries(1)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries after move-to-end, got %d: %v", len(entries), entries)
	}
}

// TestAddDuplicatePreservesOtherEntries verifies non-duplicate entries are
// unchanged after a move-to-end.
// Spec: "that older entry is removed before adding the new one at the end"
func TestAddDuplicatePreservesOtherEntries(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "alpha")
	hs.Add(1, "beta")
	hs.Add(1, "gamma")
	hs.Add(1, "beta") // move "beta" to end; "alpha" and "gamma" must remain
	entries := hs.Entries(1)
	// Expected order: alpha, gamma, beta
	expected := []string{"alpha", "gamma", "beta"}
	if len(entries) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, entries)
	}
	for i, want := range expected {
		if entries[i] != want {
			t.Errorf("entries[%d] = %q, want %q (full: %v)", i, entries[i], want, entries)
		}
	}
}

// TestAddDuplicateAtOtherPositionDoesNotEvictAtCapacity verifies that when the
// list is at maxPerID and a non-tail duplicate is added, the old copy is removed
// and the new one added at the end without triggering eviction (net count stays
// the same).
// Spec: "that older entry is removed before adding the new one at the end" —
//
//	the remove+add is a move, not an add, so eviction must not also occur.
func TestAddDuplicateAtOtherPositionDoesNotEvictAtCapacity(t *testing.T) {
	hs := NewHistoryStore(3)
	hs.Add(1, "alpha")
	hs.Add(1, "beta")
	hs.Add(1, "gamma") // at capacity (3/3)
	hs.Add(1, "alpha") // non-tail dup; remove old + add at end = still 3
	entries := hs.Entries(1)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after move-to-end at capacity, got %d: %v", len(entries), entries)
	}
	expected := []string{"beta", "gamma", "alpha"}
	for i, want := range expected {
		if entries[i] != want {
			t.Errorf("entries[%d] = %q, want %q (full: %v)", i, entries[i], want, entries)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Add: eviction when exceeding maxPerID
// ---------------------------------------------------------------------------

// TestAddEvictsOldestWhenFull verifies that when the count would exceed maxPerID,
// the oldest entry is evicted.
// Spec: "If the entry count exceeds maxPerID, the oldest entry is evicted"
func TestAddEvictsOldestWhenFull(t *testing.T) {
	hs := NewHistoryStore(3)
	hs.Add(1, "one")
	hs.Add(1, "two")
	hs.Add(1, "three")
	hs.Add(1, "four") // "one" should be evicted
	entries := hs.Entries(1)
	for _, e := range entries {
		if e == "one" {
			t.Errorf("oldest entry 'one' should have been evicted; got %v", entries)
		}
	}
}

// TestAddEvictionKeepsNewestEntries verifies the retained entries are the most
// recent ones after eviction.
// Spec: "If the entry count exceeds maxPerID, the oldest entry is evicted"
func TestAddEvictionKeepsNewestEntries(t *testing.T) {
	hs := NewHistoryStore(3)
	hs.Add(1, "one")
	hs.Add(1, "two")
	hs.Add(1, "three")
	hs.Add(1, "four")
	entries := hs.Entries(1)
	expected := []string{"two", "three", "four"}
	if len(entries) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, entries)
	}
	for i, want := range expected {
		if entries[i] != want {
			t.Errorf("entries[%d] = %q, want %q", i, entries[i], want)
		}
	}
}

// TestAddDoesNotExceedMax verifies the list never grows past maxPerID.
// Spec: "If the entry count exceeds maxPerID, the oldest entry is evicted"
func TestAddDoesNotExceedMax(t *testing.T) {
	const max = 4
	hs := NewHistoryStore(max)
	for i := 0; i < max+10; i++ {
		hs.Add(1, string(rune('a'+i)))
	}
	entries := hs.Entries(1)
	if len(entries) > max {
		t.Errorf("Entries count %d exceeds maxPerID %d", len(entries), max)
	}
}

// TestAddMaxOneAllowsExactlyOne verifies that maxPerID=1 keeps exactly one entry.
// Spec: "If the entry count exceeds maxPerID, the oldest entry is evicted"
func TestAddMaxOneAllowsExactlyOne(t *testing.T) {
	hs := NewHistoryStore(1)
	hs.Add(1, "first")
	hs.Add(1, "second")
	entries := hs.Entries(1)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry for maxPerID=1, got %d: %v", len(entries), entries)
	}
	if entries[0] != "second" {
		t.Errorf("expected newest entry 'second', got %q", entries[0])
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Add: unknown ID creates new list
// ---------------------------------------------------------------------------

// TestAddToNewIDCreatesEntry verifies that adding to an ID that has no prior
// entries creates the entry.
// Spec: "Adding to an ID that doesn't exist yet creates the list"
func TestAddToNewIDCreatesEntry(t *testing.T) {
	hs := NewHistoryStore(5)
	hs.Add(42, "hello")
	entries := hs.Entries(42)
	if len(entries) != 1 || entries[0] != "hello" {
		t.Errorf("expected [hello] for new ID 42, got %v", entries)
	}
}

// TestAddToMultipleIDsAreIndependent verifies entries for different IDs are
// stored independently.
// Spec: "Adding to an ID that doesn't exist yet creates the list" — each ID is
//        its own list.
func TestAddToMultipleIDsAreIndependent(t *testing.T) {
	hs := NewHistoryStore(5)
	hs.Add(1, "alpha")
	hs.Add(2, "beta")
	if got := hs.Entries(1); len(got) != 1 || got[0] != "alpha" {
		t.Errorf("ID 1: expected [alpha], got %v", got)
	}
	if got := hs.Entries(2); len(got) != 1 || got[0] != "beta" {
		t.Errorf("ID 2: expected [beta], got %v", got)
	}
}

// TestAddToIDDoesNotAffectOtherIDs verifies that eviction in one ID's list
// does not affect another ID's list.
// Spec: "Adding to an ID that doesn't exist yet creates the list"
func TestAddToIDDoesNotAffectOtherIDs(t *testing.T) {
	hs := NewHistoryStore(2)
	hs.Add(1, "x")
	hs.Add(1, "y")
	hs.Add(1, "z") // evicts "x" from ID 1
	hs.Add(2, "x") // ID 2 is unaffected
	if got := hs.Entries(2); len(got) != 1 || got[0] != "x" {
		t.Errorf("ID 2 should be unaffected; got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Entries: chronological order
// ---------------------------------------------------------------------------

// TestEntriesChronologicalOrder verifies Entries returns oldest first, newest last.
// Spec: "Entries(id int) returns entries in chronological order (oldest first, newest last)"
func TestEntriesChronologicalOrder(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "first")
	hs.Add(1, "second")
	hs.Add(1, "third")
	entries := hs.Entries(1)
	expected := []string{"first", "second", "third"}
	if len(entries) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, entries)
	}
	for i, want := range expected {
		if entries[i] != want {
			t.Errorf("entries[%d] = %q, want %q", i, entries[i], want)
		}
	}
}

// TestEntriesSingleEntryOrder verifies that a single-entry list is returned
// correctly (boundary: one entry).
// Spec: "Entries(id int) returns entries in chronological order"
func TestEntriesSingleEntryOrder(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "only")
	entries := hs.Entries(1)
	if len(entries) != 1 || entries[0] != "only" {
		t.Errorf("expected [only], got %v", entries)
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Entries: unknown ID returns nil or empty
// ---------------------------------------------------------------------------

// TestEntriesUnknownIDReturnsNilOrEmpty verifies Entries returns nil or an
// empty slice for an ID that has never had any entries added.
// Spec: "For an unknown ID, returns nil (or empty slice)"
func TestEntriesUnknownIDReturnsNilOrEmpty(t *testing.T) {
	hs := NewHistoryStore(5)
	entries := hs.Entries(99)
	if len(entries) != 0 {
		t.Errorf("expected nil/empty for unknown ID 99, got %v", entries)
	}
}

// TestEntriesUnknownIDZeroReturnsNilOrEmpty verifies Entries for ID 0 when
// nothing has been added (boundary: zero value of int).
// Spec: "For an unknown ID, returns nil (or empty slice)"
func TestEntriesUnknownIDZeroReturnsNilOrEmpty(t *testing.T) {
	hs := NewHistoryStore(5)
	entries := hs.Entries(0)
	if len(entries) != 0 {
		t.Errorf("expected nil/empty for ID 0, got %v", entries)
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Clear
// ---------------------------------------------------------------------------

// TestClearRemovesAllEntries verifies Clear removes all entries for all IDs.
// Spec: "Clear() removes all entries for all IDs"
func TestClearRemovesAllEntries(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "alpha")
	hs.Add(2, "beta")
	hs.Clear()
	if got := hs.Entries(1); len(got) != 0 {
		t.Errorf("after Clear, ID 1 should be empty, got %v", got)
	}
	if got := hs.Entries(2); len(got) != 0 {
		t.Errorf("after Clear, ID 2 should be empty, got %v", got)
	}
}

// TestClearAllowsSubsequentAdds verifies the store is usable after Clear.
// Spec: "Clear() removes all entries for all IDs" — the store must remain
//        functional; "Adding to an ID that doesn't exist yet creates the list"
func TestClearAllowsSubsequentAdds(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "before")
	hs.Clear()
	hs.Add(1, "after")
	entries := hs.Entries(1)
	if len(entries) != 1 || entries[0] != "after" {
		t.Errorf("after Clear and re-add, expected [after], got %v", entries)
	}
}

// TestClearOnEmptyStoreIsNoop verifies Clear on a brand-new store does not panic.
// Spec: "Clear() removes all entries for all IDs"
func TestClearOnEmptyStoreIsNoop(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Clear() // must not panic
}

// ---------------------------------------------------------------------------
// Section 10 — Falsifying tests
// ---------------------------------------------------------------------------

// TestAddAppendsDontPrepend guards against an implementation that adds new
// entries to the front rather than the end.
// Falsifies: a lazy "prepend" implementation for "adds the new one at the end".
func TestAddAppendsDontPrepend(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "first")
	hs.Add(1, "second")
	entries := hs.Entries(1)
	if len(entries) < 2 {
		t.Fatalf("expected 2 entries, got %v", entries)
	}
	if entries[0] != "first" {
		t.Errorf("oldest entry should be 'first' (chronological order), got %q", entries[0])
	}
	if entries[len(entries)-1] != "second" {
		t.Errorf("newest entry should be 'second', got %q", entries[len(entries)-1])
	}
}

// TestAddDedupOnlyAtTail guards against an implementation that treats any
// duplicate as a tail dedup (suppressing move-to-end).
// Falsifies: a lazy "just ignore all duplicates" implementation.
func TestAddDedupOnlyAtTail(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "alpha")
	hs.Add(1, "beta")
	hs.Add(1, "alpha") // not at tail; must move to end, not be silently dropped
	entries := hs.Entries(1)
	if len(entries) == 0 || entries[len(entries)-1] != "alpha" {
		t.Errorf("'alpha' should appear at the end after move-to-end, got %v", entries)
	}
}

// TestEvictionRemovesOldestNotNewest guards against an off-by-one that removes
// the most recent entry instead of the oldest.
// Falsifies: a lazy "remove last" eviction strategy.
func TestEvictionRemovesOldestNotNewest(t *testing.T) {
	hs := NewHistoryStore(2)
	hs.Add(1, "old")
	hs.Add(1, "new")
	hs.Add(1, "newest") // triggers eviction; "old" must go, "new" and "newest" remain
	entries := hs.Entries(1)
	for _, e := range entries {
		if e == "old" {
			t.Errorf("'old' should have been evicted, but it is still in %v", entries)
		}
	}
	found := false
	for _, e := range entries {
		if e == "newest" {
			found = true
		}
	}
	if !found {
		t.Errorf("'newest' should be retained after eviction, but got %v", entries)
	}
}

// TestClearActuallyRemovesVsEmpty guards against a Clear that zeros a counter
// but leaves data, so Entries still sees phantom results.
// Falsifies: an implementation that "hides" data without deleting it.
func TestClearActuallyRemovesVsEmpty(t *testing.T) {
	hs := NewHistoryStore(10)
	hs.Add(1, "alpha")
	hs.Clear()
	hs.Add(1, "beta")
	entries := hs.Entries(1)
	for _, e := range entries {
		if e == "alpha" {
			t.Errorf("'alpha' survived a Clear(); entries = %v", entries)
		}
	}
}

// TestDefaultHistoryIsDistinctFromNewStore guards against DefaultHistory being
// nil or being the same object returned by a fresh NewHistoryStore call in a
// way that could expose shared state. This test simply verifies DefaultHistory
// is non-nil and independent from a newly created store.
// Falsifies: a lazy "return nil" default.
func TestDefaultHistoryIsDistinctFromNewStore(t *testing.T) {
	fresh := NewHistoryStore(20)
	fresh.Add(1, "sentinel")
	// DefaultHistory must not have been contaminated by the add above.
	if got := DefaultHistory.Entries(1); len(got) != 0 {
		// Only a problem if it actually returns our sentinel; the store may
		// have pre-existing state from other tests if they use DefaultHistory,
		// so we specifically check for our canary value.
		for _, e := range got {
			if e == "sentinel" {
				t.Errorf("DefaultHistory shares state with a fresh store; found 'sentinel' in Entries(1)")
			}
		}
	}
}
