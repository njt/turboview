package tv

import "testing"

// phaseTestView is a test double that supports three-phase dispatch testing.
// It tracks call counts, stores the last event received, can optionally clear
// the event on handling, and records dispatch order via a shared slice.
type phaseTestView struct {
	BaseView
	name          string
	handleCount   int
	lastEvent     *Event
	clearOnHandle bool
	callOrder     *[]string
}

func (p *phaseTestView) Draw(buf *DrawBuffer) {}

func (p *phaseTestView) HandleEvent(event *Event) {
	p.handleCount++
	p.lastEvent = event
	if p.callOrder != nil {
		*p.callOrder = append(*p.callOrder, p.name)
	}
	if p.clearOnHandle {
		event.What = EvNothing
	}
}

// newPhaseView creates a visible phaseTestView with the given name and bounds.
func newPhaseView(name string, bounds Rect) *phaseTestView {
	v := &phaseTestView{name: name}
	v.SetBounds(bounds)
	v.SetState(SfVisible, true)
	return v
}

// newSelectablePhaseView creates a visible, selectable phaseTestView.
func newSelectablePhaseView(name string, bounds Rect) *phaseTestView {
	v := newPhaseView(name, bounds)
	v.SetOptions(OfSelectable, true)
	return v
}

// sharedOrder returns a pointer to an order-tracking slice, set on all provided views.
func sharedOrder(views ...*phaseTestView) *[]string {
	order := &[]string{}
	for _, v := range views {
		v.callOrder = order
	}
	return order
}

// defaultBounds is a convenience rect for tests that don't care about position.
func defaultBounds() Rect { return NewRect(0, 0, 10, 5) }

// ── Preprocess phase ──────────────────────────────────────────────────────────

// TestPreprocessChildReceivesEventBeforeFocused verifies that a child with
// OfPreProcess receives the event before the focused child.
// Spec: "Preprocess: iterate all children — for each child with OfPreProcess
// that is NOT the focused child, call child.HandleEvent(event)."
func TestPreprocessChildReceivesEventBeforeFocused(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	pre := newPhaseView("pre", defaultBounds())
	pre.SetOptions(OfPreProcess, true)
	focused := newSelectablePhaseView("focused", defaultBounds())
	order := sharedOrder(pre, focused)

	g.Insert(pre)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	if len(*order) < 2 {
		t.Fatalf("expected at least 2 HandleEvent calls, got %d", len(*order))
	}
	if (*order)[0] != "pre" {
		t.Errorf("first call = %q, want %q (preprocess before focused)", (*order)[0], "pre")
	}
	if (*order)[1] != "focused" {
		t.Errorf("second call = %q, want %q (focused after preprocess)", (*order)[1], "focused")
	}
}

// TestPreprocessFocusedChildSkippedInPreprocessPhase verifies that the focused
// child is NOT called during the preprocess phase even when it has OfPreProcess.
// Spec: "for each child with OfPreProcess that is NOT the focused child".
func TestPreprocessFocusedChildSkippedInPreprocessPhase(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	focused := newSelectablePhaseView("focused", defaultBounds())
	focused.SetOptions(OfPreProcess, true)

	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	// focused has OfPreProcess but IS the focused child; it must receive the
	// event exactly once (via the focused phase), not twice.
	if focused.handleCount != 1 {
		t.Errorf("focused child with OfPreProcess handleCount = %d, want 1 (not double-dispatched)", focused.handleCount)
	}
}

// TestPreprocessChildWithoutFlagIsSkipped verifies that a child lacking
// OfPreProcess is not called during the preprocess phase.
// Spec: "for each child with OfPreProcess … call child.HandleEvent(event)"
// — children without the flag are not included in preprocess.
func TestPreprocessChildWithoutFlagIsSkipped(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	bystander := newPhaseView("bystander", defaultBounds()) // no OfPreProcess
	focused := newSelectablePhaseView("focused", defaultBounds())
	order := sharedOrder(bystander, focused)

	g.Insert(bystander)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	// Only the focused child should have been called; the bystander has no
	// OfPreProcess, so it must not appear in the order.
	for _, name := range *order {
		if name == "bystander" {
			t.Errorf("child without OfPreProcess was called during preprocess phase")
		}
	}
}

// TestPreprocessClearedEventStopsFocusedDispatch verifies that if a preprocess
// child clears the event, the focused child does NOT receive it.
// Spec: "Stop if event is cleared" (after preprocess) — focused child skipped.
func TestPreprocessClearedEventStopsFocusedDispatch(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	pre := newPhaseView("pre", defaultBounds())
	pre.SetOptions(OfPreProcess, true)
	pre.clearOnHandle = true
	focused := newSelectablePhaseView("focused", defaultBounds())

	g.Insert(pre)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	if focused.handleCount != 0 {
		t.Errorf("focused child handleCount = %d, want 0 when preprocess cleared event", focused.handleCount)
	}
}

// ── Focused phase ─────────────────────────────────────────────────────────────

// TestFocusedChildReceivesEvent verifies the focused child receives the event.
// Spec: "Focused: if not cleared, forward to the focused child."
func TestFocusedChildReceivesEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	focused := newSelectablePhaseView("focused", defaultBounds())
	g.Insert(focused)

	event := &Event{What: EvKeyboard}
	g.HandleEvent(event)

	if focused.lastEvent != event {
		t.Errorf("focused child did not receive the event")
	}
}

// TestFocusedPhaseSkippedWhenPreprocessCleared verifies that after a preprocess
// child clears the event, the focused phase is not executed.
// Spec: "Focused: if not cleared, forward to the focused child."
func TestFocusedPhaseSkippedWhenPreprocessCleared(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	pre := newPhaseView("pre", defaultBounds())
	pre.SetOptions(OfPreProcess, true)
	pre.clearOnHandle = true
	focused := newSelectablePhaseView("focused", defaultBounds())

	g.Insert(pre)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	if focused.handleCount > 0 {
		t.Errorf("focused phase executed despite preprocess clearing the event")
	}
}

// ── Postprocess phase ─────────────────────────────────────────────────────────

// TestPostprocessChildReceivesEventAfterFocused verifies that a child with
// OfPostProcess receives the event after the focused child.
// Spec: "Postprocess: iterate all children — for each child with OfPostProcess
// that is NOT the focused child, call child.HandleEvent(event)."
func TestPostprocessChildReceivesEventAfterFocused(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	focused := newSelectablePhaseView("focused", defaultBounds())
	post := newPhaseView("post", defaultBounds())
	post.SetOptions(OfPostProcess, true)
	order := sharedOrder(focused, post)

	g.Insert(focused)
	g.Insert(post)

	g.HandleEvent(&Event{What: EvKeyboard})

	if len(*order) < 2 {
		t.Fatalf("expected at least 2 HandleEvent calls, got %d", len(*order))
	}
	if (*order)[0] != "focused" {
		t.Errorf("first call = %q, want %q (focused before postprocess)", (*order)[0], "focused")
	}
	if (*order)[1] != "post" {
		t.Errorf("second call = %q, want %q (postprocess after focused)", (*order)[1], "post")
	}
}

// TestPostprocessFocusedChildSkippedInPostprocessPhase verifies that the focused
// child is NOT called again during the postprocess phase even when it has
// OfPostProcess.
// Spec: "for each child with OfPostProcess that is NOT the focused child".
func TestPostprocessFocusedChildSkippedInPostprocessPhase(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	focused := newSelectablePhaseView("focused", defaultBounds())
	focused.SetOptions(OfPostProcess, true)

	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	// focused has OfPostProcess but IS the focused child; it must receive the
	// event exactly once (via the focused phase), not twice.
	if focused.handleCount != 1 {
		t.Errorf("focused child with OfPostProcess handleCount = %d, want 1 (not double-dispatched)", focused.handleCount)
	}
}

// TestPostprocessChildWithoutFlagIsSkipped verifies that a child lacking
// OfPostProcess is not called during the postprocess phase.
// Spec: "for each child with OfPostProcess … call child.HandleEvent(event)"
// — children without the flag are not included in postprocess.
func TestPostprocessChildWithoutFlagIsSkipped(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	focused := newSelectablePhaseView("focused", defaultBounds())
	bystander := newPhaseView("bystander", defaultBounds()) // no OfPostProcess

	g.Insert(focused)
	g.Insert(bystander)

	g.HandleEvent(&Event{What: EvKeyboard})

	if bystander.handleCount != 0 {
		t.Errorf("child without OfPostProcess handleCount = %d, want 0", bystander.handleCount)
	}
}

// TestPostprocessSkippedWhenFocusedClearsEvent verifies that if the focused child
// clears the event, postprocess children do NOT receive it.
// Spec: "Stop if event is cleared" (after focused phase).
func TestPostprocessSkippedWhenFocusedClearsEvent(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	focused := newSelectablePhaseView("focused", defaultBounds())
	focused.clearOnHandle = true
	post := newPhaseView("post", defaultBounds())
	post.SetOptions(OfPostProcess, true)

	g.Insert(focused)
	g.Insert(post)

	g.HandleEvent(&Event{What: EvKeyboard})

	if post.handleCount != 0 {
		t.Errorf("postprocess child handleCount = %d, want 0 when focused cleared event", post.handleCount)
	}
}

// ── Full three-phase order ────────────────────────────────────────────────────

// TestThreePhaseDispatchOrder verifies the complete dispatch sequence:
// preprocess → focused → postprocess.
// Spec: Phase 1 (preprocess), Phase 2 (focused), Phase 3 (postprocess).
func TestThreePhaseDispatchOrder(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	pre := newPhaseView("pre", defaultBounds())
	pre.SetOptions(OfPreProcess, true)
	focused := newSelectablePhaseView("focused", defaultBounds())
	post := newPhaseView("post", defaultBounds())
	post.SetOptions(OfPostProcess, true)
	order := sharedOrder(pre, focused, post)

	g.Insert(pre)
	g.Insert(focused)
	g.Insert(post)

	g.HandleEvent(&Event{What: EvKeyboard})

	want := []string{"pre", "focused", "post"}
	if len(*order) != len(want) {
		t.Fatalf("call order len = %d, want %d; got %v", len(*order), len(want), *order)
	}
	for i, name := range want {
		if (*order)[i] != name {
			t.Errorf("call order[%d] = %q, want %q; full order: %v", i, (*order)[i], name, *order)
		}
	}
}

// TestThreePhaseDispatchOrderFalsified confirms the test above would fail if
// postprocess ran before focused (guards against a wrong implementation that
// accidentally passes by coincidence of insertion order).
// This is a falsifying test: it asserts that post → focused order is wrong.
func TestThreePhaseDispatchOrderFalsified(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	post := newPhaseView("post", defaultBounds())
	post.SetOptions(OfPostProcess, true)
	focused := newSelectablePhaseView("focused", defaultBounds())
	// Insert post before focused to ensure insertion order cannot mask a bug.
	order := sharedOrder(post, focused)

	g.Insert(post)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	if len(*order) < 2 {
		t.Fatalf("expected 2 calls, got %d", len(*order))
	}
	// Even though post was inserted first, it must still come AFTER focused.
	if (*order)[0] != "focused" {
		t.Errorf("post (inserted first) ran before focused: order = %v", *order)
	}
	if (*order)[1] != "post" {
		t.Errorf("post did not run after focused: order = %v", *order)
	}
}

// ── Mouse events ──────────────────────────────────────────────────────────────

// TestMouseEventForwardsDirectlyToFocused verifies that EvMouse skips the
// preprocess/postprocess phases and goes directly to the focused child.
// Spec: "Mouse events (EvMouse) skip three-phase dispatch and forward directly
// to the focused child."
func TestMouseEventForwardsDirectlyToFocused(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	pre := newPhaseView("pre", defaultBounds())
	pre.SetOptions(OfPreProcess, true)
	focused := newSelectablePhaseView("focused", defaultBounds())
	order := sharedOrder(pre, focused)

	g.Insert(pre)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{}})

	if focused.handleCount != 1 {
		t.Errorf("focused child handleCount = %d, want 1 for mouse event", focused.handleCount)
	}
	// pre must not have been called
	for _, name := range *order {
		if name == "pre" {
			t.Errorf("preprocess child was called for a mouse event")
		}
	}
}

// TestMouseEventPreprocessChildNotCalled verifies OfPreProcess children are
// completely bypassed for mouse events.
// Spec: "Mouse events (EvMouse) skip three-phase dispatch."
func TestMouseEventPreprocessChildNotCalled(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	pre := newPhaseView("pre", defaultBounds())
	pre.SetOptions(OfPreProcess, true)
	focused := newSelectablePhaseView("focused", defaultBounds())

	g.Insert(pre)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{}})

	if pre.handleCount != 0 {
		t.Errorf("preprocess child handleCount = %d, want 0 for mouse event", pre.handleCount)
	}
}

// TestMouseEventUsesPositionalRoutingNotThreePhase verifies mouse events use
// positional routing (topmost visible child at the point) rather than
// three-phase dispatch. A postprocess child at the same bounds is the topmost
// child and receives the event via positional routing.
// Spec: "Mouse events use positional routing, not three-phase dispatch."
func TestMouseEventUsesPositionalRoutingNotThreePhase(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	focused := newSelectablePhaseView("focused", defaultBounds())
	post := newPhaseView("post", defaultBounds())
	post.SetOptions(OfPostProcess, true)
	post.SetState(SfVisible, true)

	g.Insert(focused)
	g.Insert(post) // topmost — positional routing picks this one

	g.HandleEvent(&Event{What: EvMouse, Mouse: &MouseEvent{}})

	if post.handleCount != 1 {
		t.Errorf("topmost child handleCount = %d, want 1 via positional routing", post.handleCount)
	}
	if focused.handleCount != 0 {
		t.Errorf("focused child handleCount = %d, want 0 (not topmost at point)", focused.handleCount)
	}
}

// ── Clearing stops iteration within and across phases ────────────────────────

// TestPreprocessClearStopsSubsequentPreprocessChildren verifies that when a
// preprocess child clears the event, later preprocess children are skipped.
// Spec: "Stop if event is cleared" (preprocess phase).
func TestPreprocessClearStopsSubsequentPreprocessChildren(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	pre1 := newPhaseView("pre1", defaultBounds())
	pre1.SetOptions(OfPreProcess, true)
	pre1.clearOnHandle = true
	pre2 := newPhaseView("pre2", defaultBounds())
	pre2.SetOptions(OfPreProcess, true)
	focused := newSelectablePhaseView("focused", defaultBounds())

	g.Insert(pre1)
	g.Insert(pre2)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	if pre2.handleCount != 0 {
		t.Errorf("pre2 handleCount = %d, want 0 (first preprocess child cleared event)", pre2.handleCount)
	}
}

// TestPreprocessClearStopsPostprocessPhase verifies that when a preprocess child
// clears the event, the postprocess phase is also skipped.
// Spec: "Stop if event is cleared" applies across phases.
func TestPreprocessClearStopsPostprocessPhase(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	pre := newPhaseView("pre", defaultBounds())
	pre.SetOptions(OfPreProcess, true)
	pre.clearOnHandle = true
	focused := newSelectablePhaseView("focused", defaultBounds())
	post := newPhaseView("post", defaultBounds())
	post.SetOptions(OfPostProcess, true)

	g.Insert(pre)
	g.Insert(focused)
	g.Insert(post)

	g.HandleEvent(&Event{What: EvKeyboard})

	if post.handleCount != 0 {
		t.Errorf("post handleCount = %d, want 0 (preprocess cleared event)", post.handleCount)
	}
}

// TestPostprocessClearStopsSubsequentPostprocessChildren verifies that when a
// postprocess child clears the event, later postprocess children are skipped.
// Spec: "Stop if event is cleared" (postprocess phase).
func TestPostprocessClearStopsSubsequentPostprocessChildren(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	focused := newSelectablePhaseView("focused", defaultBounds())
	post1 := newPhaseView("post1", defaultBounds())
	post1.SetOptions(OfPostProcess, true)
	post1.clearOnHandle = true
	post2 := newPhaseView("post2", defaultBounds())
	post2.SetOptions(OfPostProcess, true)

	g.Insert(focused)
	g.Insert(post1)
	g.Insert(post2)

	g.HandleEvent(&Event{What: EvKeyboard})

	if post2.handleCount != 0 {
		t.Errorf("post2 handleCount = %d, want 0 (first postprocess child cleared event)", post2.handleCount)
	}
}

// TestDualFlagChildCalledInBothPhases verifies that a non-focused child with
// both OfPreProcess and OfPostProcess is called exactly twice — once in each
// phase.
func TestDualFlagChildCalledInBothPhases(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	dual := newPhaseView("dual", defaultBounds())
	dual.SetOptions(OfPreProcess|OfPostProcess, true)
	focused := newSelectablePhaseView("focused", defaultBounds())
	order := sharedOrder(dual, focused)

	g.Insert(dual)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	if dual.handleCount != 2 {
		t.Fatalf("dual-flag child handleCount = %d, want 2", dual.handleCount)
	}
	if len(*order) != 3 || (*order)[0] != "dual" || (*order)[1] != "focused" || (*order)[2] != "dual" {
		t.Errorf("dispatch order = %v, want [dual, focused, dual]", *order)
	}
}

// ── Cleared events ────────────────────────────────────────────────────────────

// TestClearedEventNotForwardedToAnyone verifies that an EvNothing event is not
// dispatched to any child (no phase runs).
// Spec: "Cleared events (EvNothing) are not forwarded at all."
func TestClearedEventNotForwardedToAnyone(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	pre := newPhaseView("pre", defaultBounds())
	pre.SetOptions(OfPreProcess, true)
	focused := newSelectablePhaseView("focused", defaultBounds())
	post := newPhaseView("post", defaultBounds())
	post.SetOptions(OfPostProcess, true)

	g.Insert(pre)
	g.Insert(focused)
	g.Insert(post)

	g.HandleEvent(&Event{What: EvNothing})

	if pre.handleCount != 0 {
		t.Errorf("preprocess child handleCount = %d, want 0 for cleared event", pre.handleCount)
	}
	if focused.handleCount != 0 {
		t.Errorf("focused child handleCount = %d, want 0 for cleared event", focused.handleCount)
	}
	if post.handleCount != 0 {
		t.Errorf("postprocess child handleCount = %d, want 0 for cleared event", post.handleCount)
	}
}

// ── Broadcast events ──────────────────────────────────────────────────────────

// TestBroadcastDeliveredToAllChildren verifies that EvBroadcast is delivered to
// ALL children, not just the focused child.
// Spec: "Broadcast events (EvBroadcast) are delivered to ALL children,
// not just focused."
func TestBroadcastDeliveredToAllChildren(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	child1 := newPhaseView("child1", defaultBounds())
	focused := newSelectablePhaseView("focused", defaultBounds())
	child3 := newPhaseView("child3", defaultBounds())

	g.Insert(child1)
	g.Insert(focused)
	g.Insert(child3)

	g.HandleEvent(&Event{What: EvBroadcast})

	if child1.handleCount != 1 {
		t.Errorf("child1 handleCount = %d, want 1 for broadcast", child1.handleCount)
	}
	if focused.handleCount != 1 {
		t.Errorf("focused handleCount = %d, want 1 for broadcast", focused.handleCount)
	}
	if child3.handleCount != 1 {
		t.Errorf("child3 handleCount = %d, want 1 for broadcast", child3.handleCount)
	}
}

// TestBroadcastIncludesNonSelectableChildren verifies that non-selectable
// (non-focused) children also receive broadcast events.
// Spec: "EvBroadcast includes non-focused, non-selectable children."
func TestBroadcastIncludesNonSelectableChildren(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	nonSelectable := newPhaseView("nonsel", defaultBounds()) // no OfSelectable
	focused := newSelectablePhaseView("focused", defaultBounds())

	g.Insert(nonSelectable)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvBroadcast})

	if nonSelectable.handleCount != 1 {
		t.Errorf("non-selectable child handleCount = %d, want 1 for broadcast", nonSelectable.handleCount)
	}
}

// TestBroadcastFalsified confirms that a non-broadcast keyboard event does NOT
// go to non-focused, non-preprocessor children (guards against an
// over-broad implementation that always sends to all).
func TestBroadcastFalsified(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	bystander := newPhaseView("bystander", defaultBounds())
	focused := newSelectablePhaseView("focused", defaultBounds())

	g.Insert(bystander)
	g.Insert(focused)

	g.HandleEvent(&Event{What: EvKeyboard})

	if bystander.handleCount != 0 {
		t.Errorf("bystander handleCount = %d, want 0 for keyboard event (only broadcast goes to all)", bystander.handleCount)
	}
}

// ── Backward compatibility ────────────────────────────────────────────────────

// TestBackwardCompatKeyboardGoesToFocusedWhenNoPhaseFlags verifies that when no
// children have OfPreProcess or OfPostProcess, keyboard events still reach the
// focused child exactly as before.
// Spec: "When no children have OfPreProcess or OfPostProcess, behavior is
// identical to the existing single-phase dispatch."
func TestBackwardCompatKeyboardGoesToFocusedWhenNoPhaseFlags(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	focused := newSelectablePhaseView("focused", defaultBounds())
	g.Insert(focused)

	event := &Event{What: EvKeyboard}
	g.HandleEvent(event)

	if focused.lastEvent != event {
		t.Errorf("backward compat: focused child did not receive keyboard event")
	}
	if focused.handleCount != 1 {
		t.Errorf("backward compat: focused child handleCount = %d, want 1", focused.handleCount)
	}
}

// TestBackwardCompatCommandGoesToFocusedWhenNoPhaseFlags verifies command events
// reach the focused child when no phase flags are present.
// Spec: "When no children have OfPreProcess or OfPostProcess, behavior is
// identical to the existing single-phase dispatch."
func TestBackwardCompatCommandGoesToFocusedWhenNoPhaseFlags(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	focused := newSelectablePhaseView("focused", defaultBounds())
	g.Insert(focused)

	event := &Event{What: EvCommand, Command: CmClose}
	g.HandleEvent(event)

	if focused.lastEvent != event {
		t.Errorf("backward compat: focused child did not receive command event")
	}
}

// TestBackwardCompatNoPanicWhenNoFocusedChildAndNoPhaseFlags verifies no panic
// when the group is empty and there are no phase-flagged children.
// Spec: "When no children have OfPreProcess or OfPostProcess, behavior is
// identical to the existing single-phase dispatch" — the nil-focus case must
// still be safe.
func TestBackwardCompatNoPanicWhenNoFocusedChildAndNoPhaseFlags(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	// Must not panic.
	g.HandleEvent(&Event{What: EvKeyboard})
}
