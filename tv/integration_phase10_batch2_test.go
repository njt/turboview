package tv

// integration_phase10_batch2_test.go — Integration tests for Phase 10 Tasks 6–8:
// Label Behavior Checkpoint.
//
// Verifies that Label widgets work correctly inside real Groups and with real
// Buttons:
//
//   Task 6: Label click focuses the linked view
//   Task 7: Label light field tracks linked view focus via broadcasts
//   Task 8: Label Alt+shortcut runs in post-process phase (after focused view)
//
// Test naming: TestIntegrationPhase10Batch2<DescriptiveSuffix>

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Test 1: In a Group with a Label linked to a Button, clicking the Label
// focuses the Button
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch2LabelClickFocusesLinkedButton verifies that a
// Button1 mouse click on a Label routes through the Group's positional mouse
// dispatch to the Label, which then calls SetFocusedChild on the Group to
// focus the linked Button.
//
// Chain: Group.HandleEvent(mouse at Label) → Label.HandleEvent →
//        owner.SetFocusedChild(button) → Group focuses button.
func TestIntegrationPhase10Batch2LabelClickFocusesLinkedButton(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	btn := NewButton(NewRect(20, 0, 15, 1), "~O~K", CmOK)
	// Label at (0,0,10,1); mouse click at (0,0) lands inside label bounds.
	lbl := NewLabel(NewRect(0, 0, 10, 1), "~O~K label", btn)

	g.Insert(lbl)
	g.Insert(btn) // btn gets focus (selectable, inserted last)

	// btn is currently focused; focus something else so we can observe a change.
	// Insert a third selectable view to displace initial focus target.
	other := newLabelLinkedView()
	other.SetBounds(NewRect(0, 5, 10, 1))
	g.Insert(other) // other now has focus

	if g.FocusedChild() != other {
		t.Fatalf("pre-condition: FocusedChild() = %v, want other", g.FocusedChild())
	}

	// Deliver a mouse click at (0,0) — inside the label's bounds at (0,0,10,1).
	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() != btn {
		t.Errorf("after Label click: FocusedChild() = %v, want btn", g.FocusedChild())
	}
}

// TestIntegrationPhase10Batch2LabelClickEventIsConsumed verifies that after a
// Button1 click on the Label, the event is cleared (consumed) by the Label.
func TestIntegrationPhase10Batch2LabelClickEventIsConsumed(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	btn := NewButton(NewRect(20, 0, 15, 1), "~O~K", CmOK)
	lbl := NewLabel(NewRect(0, 0, 10, 1), "~O~K label", btn)

	g.Insert(lbl)
	g.Insert(btn)

	// Click at label coordinates (0,0).
	ev := &Event{
		What:  EvMouse,
		Mouse: &MouseEvent{X: 0, Y: 0, Button: tcell.Button1},
	}
	g.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("mouse event was not cleared after Label click with linked view")
	}
}

// ---------------------------------------------------------------------------
// Test 2: When the linked view gains focus via SetFocusedChild, the Label's
// light field becomes true
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch2LabelLightTrueWhenLinkedButtonFocused verifies
// that when SetFocusedChild is called with the Label's linked Button, the Group
// broadcasts CmReceivedFocus, which causes label.light to become true.
//
// Chain: group.SetFocusedChild(btn) → broadcasts CmReceivedFocus{Info: btn} →
//        label.HandleEvent(broadcast) → label.light = true.
func TestIntegrationPhase10Batch2LabelLightTrueWhenLinkedButtonFocused(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	btn := NewButton(NewRect(20, 0, 15, 1), "OK", CmOK)
	lbl := NewLabel(NewRect(0, 0, 10, 1), "~O~K label", btn)
	other := newLabelLinkedView()
	other.SetBounds(NewRect(0, 5, 10, 1))

	g.Insert(lbl)
	g.Insert(btn)
	g.Insert(other) // other starts with focus

	// Pre-condition: light must be false while linked button is not focused.
	if lbl.light {
		t.Fatalf("pre-condition: label.light should be false before linked button is focused")
	}

	// Focus the linked button — should trigger CmReceivedFocus{Info: btn}.
	g.SetFocusedChild(btn)

	if !lbl.light {
		t.Errorf("label.light = false after SetFocusedChild(btn); want true")
	}
}

// ---------------------------------------------------------------------------
// Test 3: When focus moves away from the linked view, light becomes false
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch2LabelLightFalseWhenLinkedButtonLosesFocus verifies
// that when focus moves away from the Label's linked Button, the Group broadcasts
// CmReleasedFocus, which causes label.light to become false.
//
// Chain: group.SetFocusedChild(other) → broadcasts CmReleasedFocus{Info: btn} →
//        label.HandleEvent(broadcast) → label.light = false.
func TestIntegrationPhase10Batch2LabelLightFalseWhenLinkedButtonLosesFocus(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	btn := NewButton(NewRect(20, 0, 15, 1), "OK", CmOK)
	lbl := NewLabel(NewRect(0, 0, 10, 1), "~O~K label", btn)
	other := newLabelLinkedView()
	other.SetBounds(NewRect(0, 5, 10, 1))

	g.Insert(lbl)
	g.Insert(btn)
	g.Insert(other) // other starts with focus

	// First light up the label by focusing the linked button.
	g.SetFocusedChild(btn)
	if !lbl.light {
		t.Fatalf("pre-condition: label.light should be true after focusing btn")
	}

	// Now move focus to other — should broadcast CmReleasedFocus{Info: btn}.
	g.SetFocusedChild(other)

	if lbl.light {
		t.Errorf("label.light = true after focus moved away from btn; want false")
	}
}

// TestIntegrationPhase10Batch2LabelLightNotAffectedByUnrelatedFocusChange verifies
// that label.light remains false when an unrelated view (not the linked button)
// receives focus.
func TestIntegrationPhase10Batch2LabelLightNotAffectedByUnrelatedFocusChange(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	btn := NewButton(NewRect(20, 0, 15, 1), "OK", CmOK)
	lbl := NewLabel(NewRect(0, 0, 10, 1), "~O~K label", btn)
	other := newLabelLinkedView()
	other.SetBounds(NewRect(0, 5, 10, 1))
	third := newLabelLinkedView()
	third.SetBounds(NewRect(0, 10, 10, 1))

	g.Insert(lbl)
	g.Insert(btn)
	g.Insert(other)
	g.Insert(third) // third starts with focus

	// Focus other — neither btn nor lbl is involved.
	g.SetFocusedChild(other)

	if lbl.light {
		t.Errorf("label.light = true after focusing unrelated view (other); want false")
	}
}

// ---------------------------------------------------------------------------
// Test 4: Label's Alt+shortcut works in postprocess phase (after focused view)
// ---------------------------------------------------------------------------

// TestIntegrationPhase10Batch2LabelAltShortcutActivatesWhenFocusedViewIgnores
// verifies that when the focused view does NOT consume an Alt+shortcut event,
// the Label (which runs in Phase 3 / post-process) handles it by focusing the
// linked Button.
//
// This tests that OfPostProcess allows the Label to act when the focused view
// passes the event through.
func TestIntegrationPhase10Batch2LabelAltShortcutActivatesWhenFocusedViewIgnores(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))

	btn := NewButton(NewRect(20, 0, 15, 1), "OK", CmOK)
	lbl := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", btn)

	// A plain selectable view that does NOT consume Alt+N.
	passive := newLabelLinkedView()
	passive.SetBounds(NewRect(0, 5, 10, 1))

	g.Insert(lbl)
	g.Insert(btn)
	g.Insert(passive) // passive gets initial focus

	if g.FocusedChild() != passive {
		t.Fatalf("pre-condition: FocusedChild() = %v, want passive", g.FocusedChild())
	}

	ev := &Event{
		What: EvKeyboard,
		Key:  &KeyEvent{Key: tcell.KeyRune, Rune: 'n', Modifiers: tcell.ModAlt},
	}
	g.HandleEvent(ev)

	if g.FocusedChild() != btn {
		t.Errorf("after Alt+N with passive focused view: FocusedChild() = %v, want btn", g.FocusedChild())
	}
	if !ev.IsCleared() {
		t.Errorf("event was not cleared after Label handled Alt+N in post-process phase")
	}
}

