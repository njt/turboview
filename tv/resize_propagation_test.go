package tv

import "testing"

func TestListBoxSetBoundsResizesChildren(t *testing.T) {
	lb := NewStringListBox(NewRect(0, 0, 20, 10), []string{"A", "B", "C"})

	lb.SetBounds(NewRect(0, 0, 30, 15))

	vb := lb.viewer.Bounds()
	if vb.Width() != 29 || vb.Height() != 15 {
		t.Errorf("viewer should be 29x15 after resize, got %dx%d", vb.Width(), vb.Height())
	}

	sb := lb.scrollbar.Bounds()
	if sb.A.X != 29 || sb.Width() != 1 || sb.Height() != 15 {
		t.Errorf("scrollbar should be at x=29, 1x15 after resize, got x=%d, %dx%d",
			sb.A.X, sb.Width(), sb.Height())
	}
}

func TestCheckBoxesSetBoundsResizesItems(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	cbs.SetBounds(NewRect(0, 0, 30, 3))

	for i, item := range cbs.items {
		if item.Bounds().Width() != 30 {
			t.Errorf("checkbox %d width should be 30 after resize, got %d", i, item.Bounds().Width())
		}
	}
}

func TestRadioButtonsSetBoundsResizesItems(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})

	rbs.SetBounds(NewRect(0, 0, 30, 3))

	for i, item := range rbs.items {
		if item.Bounds().Width() != 30 {
			t.Errorf("radio button %d width should be 30 after resize, got %d", i, item.Bounds().Width())
		}
	}
}
