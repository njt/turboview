package tv

import (
	"testing"

	"github.com/njt/turboview/theme"
)

// TestLabelLightSetOnCmReceivedFocusForLink verifies that when the group
// broadcasts CmReceivedFocus with Info == link, the label's light field is set
// to true.
func TestLabelLightSetOnCmReceivedFocusForLink(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	// Move focus to linked — triggers CmReceivedFocus with Info=linked
	g.SetFocusedChild(linked)

	if !label.light {
		t.Errorf("label.light = false after CmReceivedFocus for link; want true")
	}
}

// TestLabelLightClearedOnCmReleasedFocusForLink verifies that when the group
// broadcasts CmReleasedFocus with Info == link, the label's light field is set
// to false.
func TestLabelLightClearedOnCmReleasedFocusForLink(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	// First light up the label by focusing the linked view.
	g.SetFocusedChild(linked)
	if !label.light {
		t.Fatalf("precondition: label.light should be true after focusing link")
	}

	// Now move focus away — triggers CmReleasedFocus with Info=linked
	g.SetFocusedChild(other)

	if label.light {
		t.Errorf("label.light = true after CmReleasedFocus for link; want false")
	}
}

// TestLabelLightNotSetForUnrelatedFocusChange verifies that focus changes for
// an unrelated view do not set label.light.
func TestLabelLightNotSetForUnrelatedFocusChange(t *testing.T) {
	g := NewGroup(NewRect(0, 0, 80, 25))
	linked := newLabelLinkedView()
	label := NewLabel(NewRect(0, 0, 10, 1), "~N~ame", linked)
	other := newLabelLinkedView()
	g.Insert(label)
	g.Insert(linked)
	g.Insert(other)

	// Focus other — CmReceivedFocus Info=other, not linked
	g.SetFocusedChild(other)

	if label.light {
		t.Errorf("label.light = true after focusing unrelated view; want false")
	}
}

// TestLabelDrawUsesLabelHighlightWhenLight verifies that when label.light is
// true, Draw uses cs.LabelHighlight for normal text segments.
func TestLabelDrawUsesLabelHighlightWhenLight(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "Name", nil)
	scheme := theme.BorlandBlue
	label.scheme = scheme
	label.light = true

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.LabelHighlight {
		t.Errorf("cell(0,0) style = %v, want LabelHighlight %v when light=true", cell.Style, scheme.LabelHighlight)
	}
}

// TestLabelDrawUsesLabelNormalWhenNotLight verifies that when label.light is
// false, Draw uses cs.LabelNormal for normal text segments.
func TestLabelDrawUsesLabelNormalWhenNotLight(t *testing.T) {
	label := NewLabel(NewRect(0, 0, 20, 1), "Name", nil)
	scheme := theme.BorlandBlue
	label.scheme = scheme
	label.light = false

	buf := NewDrawBuffer(20, 1)
	label.Draw(buf)

	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.LabelNormal {
		t.Errorf("cell(0,0) style = %v, want LabelNormal %v when light=false", cell.Style, scheme.LabelNormal)
	}
}
