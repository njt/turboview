package tv

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// --- Requirement: NewLabel sets OfPreProcess ---
//
// Spec: "NewLabel() sets both OfPreProcess and OfPostProcess on the Label"
// Spec: "label.HasOption(OfPreProcess) returns true after construction"

// TestLabelPreProcessNewLabelSetsOfPreProcess verifies that NewLabel sets
// the OfPreProcess option on the constructed Label (with a linked view).
//
// Spec: "label.HasOption(OfPreProcess) returns true after construction"
func TestLabelPreProcessNewLabelSetsOfPreProcess(t *testing.T) {
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)

	if !label.HasOption(OfPreProcess) {
		t.Error("NewLabel did not set OfPreProcess; spec requires label.HasOption(OfPreProcess) == true after construction")
	}
}

// TestLabelPreProcessNewLabelSetsOfPreProcessNilLink catches a lazy implementation
// that sets OfPreProcess only when a non-nil link is provided.
//
// Spec: "NewLabel() sets both OfPreProcess and OfPostProcess on the Label"
// (no exception for nil link)
func TestLabelPreProcessNewLabelSetsOfPreProcessNilLink(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)

	if !label.HasOption(OfPreProcess) {
		t.Error("NewLabel with nil link did not set OfPreProcess; OfPreProcess must be set regardless of link")
	}
}

// --- Requirement: NewLabel still sets OfPostProcess (must not regress) ---
//
// Spec: "label.HasOption(OfPostProcess) returns true after construction (existing behavior, must not regress)"

// TestLabelPreProcessNewLabelStillSetsOfPostProcess verifies that adding
// OfPreProcess does not inadvertently remove OfPostProcess.
//
// Spec: "label.HasOption(OfPostProcess) returns true after construction (existing behavior, must not regress)"
func TestLabelPreProcessNewLabelStillSetsOfPostProcess(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)

	if !label.HasOption(OfPostProcess) {
		t.Error("NewLabel did not set OfPostProcess; this is an existing behavior that must not regress")
	}
}

// TestLabelPreProcessBothOptionsSetSimultaneously falsifies an implementation
// that sets one option by toggling rather than ORing — i.e., catches a bug
// where setting OfPreProcess clears OfPostProcess.
//
// Spec: "NewLabel() sets both OfPreProcess and OfPostProcess on the Label"
func TestLabelPreProcessBothOptionsSetSimultaneously(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", nil)

	hasPreProcess := label.HasOption(OfPreProcess)
	hasPostProcess := label.HasOption(OfPostProcess)

	if !hasPreProcess || !hasPostProcess {
		t.Errorf(
			"NewLabel must set both OfPreProcess and OfPostProcess; got OfPreProcess=%v OfPostProcess=%v",
			hasPreProcess, hasPostProcess,
		)
	}
}

// --- Requirement: Label fires in preprocess BEFORE focused interceptor ---
//
// Spec: "When a Label with shortcut ~N~ and a focused altInterceptorView (that
// consumes Alt+N) are siblings in a Group, pressing Alt+N focuses the Label's
// link — the Label fires in preprocess BEFORE the focused child sees the event."
// Spec: "This matches original TV: TLabel intercepts hotkeys before the focused view."

// TestLabelPreProcessInterceptsBeforeFocusedAltConsumer is a unit test using
// manual state setup. It verifies that when a Label has OfPreProcess, its
// HandleEvent fires in the preprocess phase — before the focused child — so
// Alt+N focuses the link even when the focused child would have consumed the event.
//
// Spec: "the Label fires in preprocess BEFORE the focused child sees the event"
func TestLabelPreProcessInterceptsBeforeFocusedAltConsumer(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)

	interceptor := &altInterceptorView{intercept: 'n'}
	interceptor.SetBounds(NewRect(20, 0, 10, 1))
	interceptor.SetState(SfVisible, true)
	interceptor.SetOptions(OfSelectable, true)

	g.Insert(label)
	g.Insert(linked)
	g.Insert(interceptor) // inserted last — becomes focused child

	if g.FocusedChild() != interceptor {
		t.Fatalf("precondition: FocusedChild() = %v, want interceptor", g.FocusedChild())
	}

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	// Label must have fired in preprocess (before interceptor) and focused link.
	if g.FocusedChild() != linked {
		t.Errorf(
			"FocusedChild() = %v, want linked; Label must fire in preprocess before the focused interceptor",
			g.FocusedChild(),
		)
	}
}

// TestLabelPreProcessInterceptorDoesNotSeeEventAfterLabelHandlesIt is an
// integration test that runs the full Group event dispatch and verifies the
// interceptor never receives the Alt+N event when the Label handles it first
// in the preprocess phase.
//
// Spec: "the Label fires in preprocess BEFORE the focused child sees the event.
// If Label matches Alt+shortcut and clears the event, the focused child never sees it."
func TestLabelPreProcessInterceptorDoesNotSeeEventAfterLabelHandlesIt(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)

	interceptor := &altInterceptorView{intercept: 'n'}
	interceptor.SetBounds(NewRect(20, 0, 10, 1))
	interceptor.SetState(SfVisible, true)
	interceptor.SetOptions(OfSelectable, true)

	g.Insert(label)
	g.Insert(linked)
	g.Insert(interceptor)

	if g.FocusedChild() != interceptor {
		t.Fatalf("precondition: FocusedChild() = %v, want interceptor", g.FocusedChild())
	}

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	// Label cleared the event in preprocess — interceptor must not have seen it.
	if interceptor.gotEvent {
		t.Errorf(
			"interceptor received Alt+N; Label must have cleared the event in preprocess before the focused child's phase",
		)
	}
}

// --- Requirement: non-intercepting focused sibling does not break Alt+N (regression guard) ---
//
// Spec: "When a Label with shortcut ~N~ is in a Group with a non-intercepting
// focused sibling, Alt+N still focuses the Label's link (existing behavior, must not regress)"

// TestLabelPreProcessActivatesWithNonInterceptingFocusedSibling verifies that
// when the focused sibling does NOT consume Alt+N, the Label's shortcut still
// focuses the link (this is the existing post-process path, now also covered by
// the new pre-process path).
//
// Spec: "When a Label with shortcut ~N~ is in a Group with a non-intercepting
// focused sibling, Alt+N still focuses the Label's link (existing behavior, must not regress)"
func TestLabelPreProcessActivatesWithNonInterceptingFocusedSibling(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	other := newLabelLinkedView() // does not intercept any events

	g.Insert(label)
	g.Insert(linked)
	g.Insert(other) // other becomes focused child

	if g.FocusedChild() != other {
		t.Fatalf("precondition: FocusedChild() = %v, want other", g.FocusedChild())
	}

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() != linked {
		t.Errorf(
			"FocusedChild() = %v, want linked; Alt+N must focus the link when the focused sibling does not intercept",
			g.FocusedChild(),
		)
	}
}

// TestLabelPreProcessNonInterceptingFalsify catches a lazy implementation that
// only fires in preprocess (dropping the postprocess path) — if OfPreProcess is
// set but OfPostProcess is cleared, the shortcut must still work when no
// interceptor is present (since label is non-focused, preprocess still handles it).
// This test specifically verifies the event is cleared, confirming the label
// fully handled the shortcut and did not leave it for another consumer.
//
// Spec: "Alt+N still focuses the Label's link (existing behavior, must not regress)"
// Spec: "Clears the event." (from existing HandleEvent spec)
func TestLabelPreProcessNonInterceptingFalsify(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 20, 1), "~N~ame", linked)
	other := newLabelLinkedView()

	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("event was not cleared after Label handled Alt+N shortcut with non-intercepting sibling")
	}
}
