package tv

import "testing"

// disabledSpyView is a test double for disabled/EventMask filtering tests.
// It embeds BaseView, records all HandleEvent calls, and supports an optional
// shared call-order slice for ordering assertions.
type disabledSpyView struct {
	BaseView
	name        string
	handleCount int
	lastEvent   *Event
	callOrder   *[]string
}

func (s *disabledSpyView) Draw(buf *DrawBuffer) {}

func (s *disabledSpyView) HandleEvent(event *Event) {
	s.handleCount++
	s.lastEvent = event
	if s.callOrder != nil {
		*s.callOrder = append(*s.callOrder, s.name)
	}
}

// newDisabledSpy creates a visible disabledSpyView with the given name and bounds.
func newDisabledSpy(name string, bounds Rect) *disabledSpyView {
	v := &disabledSpyView{name: name}
	v.SetBounds(bounds)
	v.SetState(SfVisible, true)
	return v
}

// newSelectableDisabledSpy creates a visible, selectable disabledSpyView.
func newSelectableDisabledSpy(name string, bounds Rect) *disabledSpyView {
	v := newDisabledSpy(name, bounds)
	v.SetOptions(OfSelectable, true)
	return v
}

// ── Test 1: Disabled child skipped during positional mouse routing ────────────

// TestDisabledChildSkippedForMousePositionalRouting verifies that a child with
// SfDisabled is not delivered mouse events via positional routing even when the
// mouse point is inside its bounds.
// Spec: "During positional mouse routing: skip disabled children."
func TestDisabledChildSkippedForMousePositionalRouting(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	disabled := newDisabledSpy("disabled", NewRect(0, 0, 20, 10))
	disabled.SetState(SfDisabled, true)

	g.Insert(disabled)

	g.HandleEvent(mouseEventAt(5, 5))

	if disabled.handleCount != 0 {
		t.Errorf("disabled child handleCount = %d, want 0; disabled children must be skipped during positional mouse routing", disabled.handleCount)
	}
}

// TestDisabledChildSkippedMouseFallsThrough verifies that when a disabled child
// is the topmost child at a mouse point, routing falls through to a visible,
// enabled child beneath it.
// Spec: "During positional mouse routing: skip disabled children."
func TestDisabledChildSkippedMouseFallsThrough(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	enabled := newDisabledSpy("enabled", NewRect(0, 0, 20, 10))

	disabled := newDisabledSpy("disabled", NewRect(0, 0, 20, 10))
	disabled.SetState(SfDisabled, true)

	g.Insert(enabled)   // lower z-order, enabled
	g.Insert(disabled)  // higher z-order, disabled → must be skipped

	g.HandleEvent(mouseEventAt(5, 5))

	if disabled.handleCount != 0 {
		t.Errorf("disabled child (topmost) received mouse event; it should be skipped")
	}
	if enabled.handleCount != 1 {
		t.Errorf("enabled child beneath disabled one handleCount = %d, want 1; routing must fall through past disabled child", enabled.handleCount)
	}
}

// ── Test 2: Disabled focused child does not receive keyboard events ───────────

// TestDisabledFocusedChildSkippedForKeyboard verifies that a focused child with
// SfDisabled is not delivered keyboard events during the focused phase.
// Spec: "Skip children with SfDisabled state" during three-phase dispatch.
func TestDisabledFocusedChildSkippedForKeyboard(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newSelectableDisabledSpy("focused", defaultBounds())
	focused.SetState(SfDisabled, true)

	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	if focused.handleCount != 0 {
		t.Errorf("disabled focused child handleCount = %d, want 0; disabled focused child must not receive keyboard events", focused.handleCount)
	}
}

// ── Test 3: Disabled child skipped during preprocess phase ───────────────────

// TestDisabledChildSkippedDuringPreprocess verifies that a child with SfDisabled
// and OfPreProcess is not called during the preprocess phase.
// Spec: "Skip children with SfDisabled state" during three-phase dispatch.
func TestDisabledChildSkippedDuringPreprocess(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	pre := newDisabledSpy("pre", defaultBounds())
	pre.SetOptions(OfPreProcess, true)
	pre.SetState(SfDisabled, true)

	focused := newSelectableDisabledSpy("focused", defaultBounds())

	g.Insert(pre)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	if pre.handleCount != 0 {
		t.Errorf("disabled preprocess child handleCount = %d, want 0; disabled children must be skipped in preprocess phase", pre.handleCount)
	}
}

// ── Test 4: Disabled child skipped during postprocess phase ──────────────────

// TestDisabledChildSkippedDuringPostprocess verifies that a child with SfDisabled
// and OfPostProcess is not called during the postprocess phase.
// Spec: "Skip children with SfDisabled state" during three-phase dispatch.
func TestDisabledChildSkippedDuringPostprocess(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newSelectableDisabledSpy("focused", defaultBounds())

	post := newDisabledSpy("post", defaultBounds())
	post.SetOptions(OfPostProcess, true)
	post.SetState(SfDisabled, true)

	g.Insert(focused)
	g.Insert(post)

	g.HandleEvent(&Event{What: EvKeyboard})

	if post.handleCount != 0 {
		t.Errorf("disabled postprocess child handleCount = %d, want 0; disabled children must be skipped in postprocess phase", post.handleCount)
	}
}

// ── Test 5: Disabled child skipped for broadcast events ──────────────────────

// TestDisabledChildSkippedForBroadcast verifies that a child with SfDisabled
// does not receive EvBroadcast events.
// Spec: "Broadcast events (EvBroadcast) skip disabled children."
func TestDisabledChildSkippedForBroadcast(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	enabled := newDisabledSpy("enabled", defaultBounds())
	disabled := newDisabledSpy("disabled", defaultBounds())
	disabled.SetState(SfDisabled, true)

	g.Insert(enabled)
	g.Insert(disabled)

	g.HandleEvent(&Event{What: EvBroadcast})

	if disabled.handleCount != 0 {
		t.Errorf("disabled child handleCount = %d, want 0 for broadcast; disabled children must be skipped", disabled.handleCount)
	}
	if enabled.handleCount != 1 {
		t.Errorf("enabled child handleCount = %d, want 1 for broadcast", enabled.handleCount)
	}
}

// ── Test 6: EventMask filters out non-matching event types ───────────────────

// TestEventMaskFiltersKeyboardFromMouseOnlyChild verifies that a child whose
// EventMask is set to EvMouse does not receive EvKeyboard events.
// Spec: "Check each child's EventMask — only deliver events whose type is
// included in the mask."
func TestEventMaskFiltersKeyboardFromMouseOnlyChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	mouseOnly := newSelectableDisabledSpy("mouseOnly", defaultBounds())
	mouseOnly.SetEventMask(EvMouse) // only accepts mouse events

	g.Insert(mouseOnly)

	g.HandleEvent(&Event{What: EvKeyboard})

	if mouseOnly.handleCount != 0 {
		t.Errorf("mouse-only child handleCount = %d, want 0 for keyboard event; EventMask should filter it out", mouseOnly.handleCount)
	}
}

// TestEventMaskFiltersKeyboardFromPreprocessMouseOnlyChild verifies that the
// EventMask check also applies to preprocess-phase children.
// Spec: "only deliver events whose type is included in the mask" — preprocess.
func TestEventMaskFiltersKeyboardFromPreprocessMouseOnlyChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	pre := newDisabledSpy("pre", defaultBounds())
	pre.SetOptions(OfPreProcess, true)
	pre.SetEventMask(EvMouse) // only accepts mouse events

	focused := newSelectableDisabledSpy("focused", defaultBounds())

	g.Insert(pre)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	if pre.handleCount != 0 {
		t.Errorf("preprocess child with EvMouse mask handleCount = %d, want 0 for keyboard event", pre.handleCount)
	}
}

// TestEventMaskFiltersKeyboardFromPostprocessMouseOnlyChild verifies that the
// EventMask check also applies to postprocess-phase children.
// Spec: "only deliver events whose type is included in the mask" — postprocess.
func TestEventMaskFiltersKeyboardFromPostprocessMouseOnlyChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	focused := newSelectableDisabledSpy("focused", defaultBounds())

	post := newDisabledSpy("post", defaultBounds())
	post.SetOptions(OfPostProcess, true)
	post.SetEventMask(EvMouse) // only accepts mouse events

	g.Insert(focused)
	g.Insert(post)

	g.HandleEvent(&Event{What: EvKeyboard})

	if post.handleCount != 0 {
		t.Errorf("postprocess child with EvMouse mask handleCount = %d, want 0 for keyboard event", post.handleCount)
	}
}

// ── Test 7: EventMask 0 receives all events ───────────────────────────────────

// TestEventMaskZeroReceivesAllEvents verifies that a child with EventMask 0
// (the default) receives all event types — EventMask 0 means "accept everything".
// Spec: "if EventMask is 0, deliver all events — 0 means 'accept everything'."
func TestEventMaskZeroReceivesAllEvents(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	child := newSelectableDisabledSpy("child", defaultBounds())
	// EventMask defaults to 0 — must not filter anything.

	g.Insert(child)

	g.HandleEvent(&Event{What: EvKeyboard})

	if child.handleCount != 1 {
		t.Errorf("child with EventMask 0 handleCount = %d, want 1 for keyboard; mask 0 must accept all events", child.handleCount)
	}
}

// TestEventMaskZeroReceivesCommandEvents verifies EventMask 0 also lets
// command events through.
// Spec: "if EventMask is 0, deliver all events."
func TestEventMaskZeroReceivesCommandEvents(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	child := newSelectableDisabledSpy("child", defaultBounds())
	// EventMask defaults to 0.

	g.Insert(child)

	g.HandleEvent(&Event{What: EvCommand, Command: CmClose})

	if child.handleCount != 1 {
		t.Errorf("child with EventMask 0 handleCount = %d, want 1 for command; mask 0 must accept all events", child.handleCount)
	}
}

// ── Test 8: EventMask NOT checked for broadcast events ───────────────────────

// TestEventMaskNotCheckedForBroadcast verifies that broadcast events bypass
// the EventMask check — even a child with a restrictive mask receives EvBroadcast.
// Spec: "Broadcast events (EvBroadcast) skip disabled children but do NOT
// check EventMask."
func TestEventMaskNotCheckedForBroadcast(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// Child accepts only EvMouse — but broadcast must still reach it.
	mouseOnly := newDisabledSpy("mouseOnly", defaultBounds())
	mouseOnly.SetEventMask(EvMouse)

	g.Insert(mouseOnly)

	g.HandleEvent(&Event{What: EvBroadcast})

	if mouseOnly.handleCount != 1 {
		t.Errorf("child with EvMouse EventMask handleCount = %d for broadcast, want 1; EventMask must NOT be checked for broadcasts", mouseOnly.handleCount)
	}
}

// TestEventMaskNotCheckedForBroadcastKeyboardMaskChild verifies that a child
// with EventMask EvKeyboard still receives EvBroadcast (broadcast bypasses mask).
// Spec: "Broadcast events … do NOT check EventMask."
func TestEventMaskNotCheckedForBroadcastKeyboardMaskChild(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	// Child accepts only EvKeyboard — broadcast must still reach it.
	keyOnly := newDisabledSpy("keyOnly", defaultBounds())
	keyOnly.SetEventMask(EvKeyboard)

	g.Insert(keyOnly)

	g.HandleEvent(&Event{What: EvBroadcast})

	if keyOnly.handleCount != 1 {
		t.Errorf("child with EvKeyboard EventMask handleCount = %d for broadcast, want 1; EventMask must NOT be checked for broadcasts", keyOnly.handleCount)
	}
}
