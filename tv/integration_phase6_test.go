package tv

// integration_phase6_test.go — Integration tests for Phase 6: ScrollBar and SetColorScheme.
//
// Each test verifies one requirement from the Phase 6 plan using REAL components
// wired end-to-end: Window → ScrollBar with ColorScheme inheritance.
//
// Test naming: TestIntegrationPhase6<DescriptiveSuffix>.
//
// Requirements covered:
//   1. ScrollBar inside Window inherits Window's default ColorScheme styles.
//   2. ScrollBar inside Window with custom ColorScheme uses that scheme's styles.
//   3. Two windows with different color schemes render ScrollBars with different styles.
//   4. Mouse click events on ScrollBar (arrow clicks, page clicks) work through Window.
//   5. ScrollBar OnChange callback fires when clicked via the real event dispatch chain.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Test 1: ScrollBar inside Window inherits the Window's ColorScheme styles.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6ScrollBarInheritsWindowColorScheme verifies that a
// ScrollBar inserted into a Window inherits the Window's ColorScheme via the
// owner chain (BaseView.ColorScheme() walks up to the owner).
func TestIntegrationPhase6ScrollBarInheritsWindowColorScheme(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 20, 15), "Test")
	win.SetColorScheme(theme.BorlandBlue)

	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	win.Insert(sb)

	cs := sb.ColorScheme()
	if cs == nil {
		t.Fatal("ScrollBar.ColorScheme() returned nil; expected inheritance from Window")
	}
	if cs != theme.BorlandBlue {
		t.Errorf("ScrollBar.ColorScheme() = %v, want Window's BorlandBlue scheme", cs)
	}
}

// TestIntegrationPhase6ScrollBarRendersWithWindowColorScheme verifies that
// Draw() uses the inherited Window ColorScheme by checking the rendered cell
// styles against the scheme's ScrollBar and ScrollThumb styles.
func TestIntegrationPhase6ScrollBarRendersWithWindowColorScheme(t *testing.T) {
	scheme := theme.BorlandBlue
	if scheme.ScrollBar == tcell.StyleDefault && scheme.ScrollThumb == tcell.StyleDefault {
		t.Skip("BorlandBlue ScrollBar/ScrollThumb equal StyleDefault — rendering test would be vacuous")
	}

	win := NewWindow(NewRect(0, 0, 20, 15), "Render Test")
	win.SetColorScheme(scheme)

	// Scrollbar: height 10, range 0–100, pageSize 10, value 0.
	// Track length = 10-2 = 8. With range/pageSize that large, thumb won't fill all.
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	win.Insert(sb)

	// Draw the window (which draws children into the client area sub-buffer).
	// Window frame is 1 cell, so client area starts at (1,1) in window coords.
	buf := NewDrawBuffer(20, 15)
	win.Draw(buf)

	// Arrow button at top of scrollbar: window(1+0, 1+0) = buf(1, 1).
	arrowCell := buf.GetCell(1, 1)
	if arrowCell.Style != scheme.ScrollBar {
		t.Errorf("top arrow cell style = %v, want scheme.ScrollBar %v", arrowCell.Style, scheme.ScrollBar)
	}
	if arrowCell.Rune != '▲' {
		t.Errorf("top arrow cell rune = %q, want '▲'", arrowCell.Rune)
	}

	// Track area (row 2 in buf, which is y=1 in scrollbar = track[0]).
	trackCell := buf.GetCell(1, 2)
	// Track cells are filled with '░' in barStyle or '█' in thumbStyle.
	// At value=0, thumb is at pos 0 so the first track cell is '█' with thumbStyle.
	if trackCell.Style != scheme.ScrollThumb && trackCell.Style != scheme.ScrollBar {
		t.Errorf("track cell[0] style = %v, want ScrollThumb or ScrollBar style", trackCell.Style)
	}
}

// ---------------------------------------------------------------------------
// Test 2: Window with custom ColorScheme — ScrollBar uses that scheme's styles.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6ScrollBarUsesCustomWindowColorScheme verifies that when
// a Window is given a custom ColorScheme via SetColorScheme, a ScrollBar inside
// that Window renders with the custom scheme's ScrollBar and ScrollThumb styles.
func TestIntegrationPhase6ScrollBarUsesCustomWindowColorScheme(t *testing.T) {
	customScheme := &theme.ColorScheme{}
	*customScheme = *theme.BorlandBlue
	customScheme.ScrollBar = tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorMaroon)
	customScheme.ScrollThumb = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed)

	win := NewWindow(NewRect(0, 0, 20, 15), "Custom Scheme")
	win.SetColorScheme(customScheme)

	sb := NewScrollBar(NewRect(0, 0, 1, 8), Vertical)
	win.Insert(sb)

	// Verify ColorScheme inheritance returns the custom scheme.
	cs := sb.ColorScheme()
	if cs == nil {
		t.Fatal("ScrollBar.ColorScheme() returned nil after Window.SetColorScheme")
	}
	if cs != customScheme {
		t.Errorf("ScrollBar.ColorScheme() != customScheme; expected custom scheme propagation via Window")
	}

	// Verify rendering uses the custom scheme's ScrollBar style.
	buf := NewDrawBuffer(20, 15)
	win.Draw(buf)

	// ScrollBar at (0,0) in client, client at (1,1) in window → buf(1,1).
	arrowCell := buf.GetCell(1, 1)
	if arrowCell.Style != customScheme.ScrollBar {
		t.Errorf("arrow cell style = %v, want customScheme.ScrollBar %v", arrowCell.Style, customScheme.ScrollBar)
	}
}

// ---------------------------------------------------------------------------
// Test 3: Two windows with different color schemes render ScrollBars differently.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6TwoWindowsDifferentSchemesProduceDifferentScrollBarStyles
// verifies that two Windows with different color schemes produce visually distinct
// ScrollBar rendering.
func TestIntegrationPhase6TwoWindowsDifferentSchemesProduceDifferentScrollBarStyles(t *testing.T) {
	schemeA := &theme.ColorScheme{}
	*schemeA = *theme.BorlandBlue
	schemeA.ScrollBar = tcell.StyleDefault.Foreground(tcell.ColorRed)

	schemeB := &theme.ColorScheme{}
	*schemeB = *theme.BorlandBlue
	schemeB.ScrollBar = tcell.StyleDefault.Foreground(tcell.ColorGreen)

	// Build window A with scheme A.
	winA := NewWindow(NewRect(0, 0, 20, 15), "Window A")
	winA.SetColorScheme(schemeA)
	sbA := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	winA.Insert(sbA)

	// Build window B with scheme B.
	winB := NewWindow(NewRect(0, 0, 20, 15), "Window B")
	winB.SetColorScheme(schemeB)
	sbB := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	winB.Insert(sbB)

	bufA := NewDrawBuffer(20, 15)
	winA.Draw(bufA)
	bufB := NewDrawBuffer(20, 15)
	winB.Draw(bufB)

	// Arrow cell: client starts at (1,1) in each window's buf.
	cellA := bufA.GetCell(1, 1)
	cellB := bufB.GetCell(1, 1)

	if cellA.Style == cellB.Style {
		t.Errorf("both windows produced the same ScrollBar arrow style %v; expected them to differ", cellA.Style)
	}
	if cellA.Style != schemeA.ScrollBar {
		t.Errorf("Window A arrow style = %v, want schemeA.ScrollBar %v", cellA.Style, schemeA.ScrollBar)
	}
	if cellB.Style != schemeB.ScrollBar {
		t.Errorf("Window B arrow style = %v, want schemeB.ScrollBar %v", cellB.Style, schemeB.ScrollBar)
	}
}

// ---------------------------------------------------------------------------
// Test 4: Mouse click events on ScrollBar (arrow and page) work inside Window.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6ScrollBarArrowClickThroughWindow verifies that a mouse
// click on the top arrow button of a ScrollBar inside a Window (using window-local
// coordinates) correctly routes through Window.HandleEvent → ScrollBar.HandleEvent
// and decrements the ScrollBar's value.
func TestIntegrationPhase6ScrollBarArrowClickThroughWindow(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 20, 15), "Arrow Click")

	// ScrollBar at (0,0) in client area. Set initial value to 5 so we can step down.
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 20)
	sb.SetPageSize(5)
	sb.SetValue(5)
	win.Insert(sb)
	// Make the ScrollBar the focused child so Group routes mouse events to it.
	win.SetFocusedChild(sb)

	if sb.Value() != 5 {
		t.Fatalf("pre-condition: ScrollBar.Value() = %d, want 5", sb.Value())
	}

	// In window-local coordinates:
	//   Frame takes row 0 and col 0.
	//   Client area starts at (1,1).
	//   ScrollBar at (0,0) in client → top arrow is at window-local (1, 1).
	ev := clickEvent(1, 1, tcell.Button1)
	win.HandleEvent(ev)

	if sb.Value() != 4 {
		t.Errorf("after top arrow click through Window, Value() = %d, want 4", sb.Value())
	}
}

// TestIntegrationPhase6ScrollBarBottomArrowClickThroughWindow verifies that a
// click on the bottom arrow of a vertical ScrollBar inside a Window increments
// the value through the full event routing chain.
func TestIntegrationPhase6ScrollBarBottomArrowClickThroughWindow(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 20, 15), "Bottom Arrow Click")

	// ScrollBar height 5 → bottom arrow is at y=4 in scrollbar-local space.
	// In window-local: client starts at y=1, so bottom arrow is at y=1+4=5.
	sb := NewScrollBar(NewRect(0, 0, 1, 5), Vertical)
	sb.SetRange(0, 20)
	sb.SetPageSize(5)
	sb.SetValue(3)
	win.Insert(sb)
	// Make the ScrollBar the focused child so Group routes mouse events to it.
	win.SetFocusedChild(sb)

	ev := clickEvent(1, 5, tcell.Button1) // window-local: x=1(client col 0), y=5(sb row 4)
	win.HandleEvent(ev)

	if sb.Value() != 4 {
		t.Errorf("after bottom arrow click through Window, Value() = %d, want 4", sb.Value())
	}
}

// TestIntegrationPhase6ScrollBarPageClickThroughWindow verifies that a click
// in the track area above the thumb triggers a page-up through the Window
// event routing chain.
func TestIntegrationPhase6ScrollBarPageClickThroughWindow(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 20, 15), "Page Click")

	// ScrollBar height 10, range 0-100, pageSize 10, value 50.
	// Track length = 8. Thumb will be somewhere in the middle.
	// A click at track position near the top (below top arrow) is y=2 in window-local.
	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 100)
	sb.SetPageSize(10)
	sb.SetValue(50)
	win.Insert(sb)
	// Make the ScrollBar the focused child so Group routes mouse events to it.
	win.SetFocusedChild(sb)

	// Get thumb position to determine where to click for page-up.
	thumbPos, _ := sb.thumbInfo()

	// Click at track position 0 (y=1 in scrollbar-local, y=2 in window-local).
	// If thumb is not at position 0, this click is above the thumb → page-up.
	if thumbPos == 0 {
		t.Skip("thumb is at position 0, cannot click above it for page-up")
	}

	startValue := sb.Value()
	// Window-local y=2 = scrollbar y=1 (track[0]).
	ev := clickEvent(1, 2, tcell.Button1)
	win.HandleEvent(ev)

	if sb.Value() >= startValue {
		t.Errorf("page-up click through Window: Value() = %d, want < %d", sb.Value(), startValue)
	}
}

// ---------------------------------------------------------------------------
// Test 5: ScrollBar OnChange callback fires through the Window dispatch chain.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6ScrollBarOnChangeFiresThroughWindow verifies that
// ScrollBar.OnChange is invoked when a mouse click routes through Window.HandleEvent
// and triggers a value change in the ScrollBar.
func TestIntegrationPhase6ScrollBarOnChangeFiresThroughWindow(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 20, 15), "OnChange Test")

	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 20)
	sb.SetPageSize(5)
	sb.SetValue(5)
	win.Insert(sb)
	// Make the ScrollBar the focused child so Group routes mouse events to it.
	win.SetFocusedChild(sb)

	fired := false
	firedWith := -1
	sb.OnChange = func(v int) {
		fired = true
		firedWith = v
	}

	// Click the top arrow (window-local 1,1) to step the value down by 1.
	ev := clickEvent(1, 1, tcell.Button1)
	win.HandleEvent(ev)

	if !fired {
		t.Error("ScrollBar.OnChange was not called after click through Window dispatch chain")
	}
	if firedWith != 4 {
		t.Errorf("OnChange called with %d, want 4", firedWith)
	}
}

// TestIntegrationPhase6ScrollBarOnChangeNotFiredWhenValueUnchanged verifies
// that OnChange is NOT called when the click does not produce a value change
// (e.g., stepping down when already at minimum).
func TestIntegrationPhase6ScrollBarOnChangeNotFiredWhenValueUnchanged(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 20, 15), "OnChange No-Op Test")

	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 20)
	sb.SetPageSize(5)
	sb.SetValue(0) // already at minimum
	win.Insert(sb)
	// Make the ScrollBar the focused child so Group routes mouse events to it.
	win.SetFocusedChild(sb)

	fired := false
	sb.OnChange = func(v int) {
		fired = true
	}

	// Click the top arrow — value is already 0, so no change should occur.
	ev := clickEvent(1, 1, tcell.Button1)
	win.HandleEvent(ev)

	if fired {
		t.Error("ScrollBar.OnChange fired unexpectedly when clicking at minimum value")
	}
}

// TestIntegrationPhase6ScrollBarOnChangeMultipleClicksAccumulate verifies that
// multiple arrow clicks through the Window dispatch chain accumulate correctly
// and OnChange fires once per value change.
func TestIntegrationPhase6ScrollBarOnChangeMultipleClicksAccumulate(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 20, 15), "Multi-Click Test")

	sb := NewScrollBar(NewRect(0, 0, 1, 10), Vertical)
	sb.SetRange(0, 20)
	sb.SetPageSize(5)
	sb.SetValue(10)
	win.Insert(sb)
	// Make the ScrollBar the focused child so Group routes mouse events to it.
	win.SetFocusedChild(sb)

	callCount := 0
	lastValue := -1
	sb.OnChange = func(v int) {
		callCount++
		lastValue = v
	}

	// Click top arrow 3 times — should step from 10 down to 7.
	for i := 0; i < 3; i++ {
		ev := clickEvent(1, 1, tcell.Button1)
		win.HandleEvent(ev)
	}

	if callCount != 3 {
		t.Errorf("OnChange call count = %d, want 3 (once per arrow click)", callCount)
	}
	if sb.Value() != 7 {
		t.Errorf("final Value() = %d, want 7 after 3 top-arrow clicks from 10", sb.Value())
	}
	if lastValue != 7 {
		t.Errorf("last OnChange value = %d, want 7", lastValue)
	}
}

// ---------------------------------------------------------------------------
// Additional: Horizontal ScrollBar inside Window.
// ---------------------------------------------------------------------------

// TestIntegrationPhase6HorizontalScrollBarClickThroughWindow verifies that
// a horizontal ScrollBar inside a Window also routes click events correctly.
func TestIntegrationPhase6HorizontalScrollBarClickThroughWindow(t *testing.T) {
	win := NewWindow(NewRect(0, 0, 20, 15), "Horizontal SB")

	// Horizontal scrollbar at (0,0) in client, width 10.
	// In window-local: left arrow at x=1, y=1; right arrow at x=10, y=1.
	sb := NewScrollBar(NewRect(0, 0, 10, 1), Horizontal)
	sb.SetRange(0, 20)
	sb.SetPageSize(5)
	sb.SetValue(5)
	win.Insert(sb)
	// Make the ScrollBar the focused child so Group routes mouse events to it.
	win.SetFocusedChild(sb)

	fired := false
	firedWith := -1
	sb.OnChange = func(v int) {
		fired = true
		firedWith = v
	}

	// Click right arrow at window-local (10, 1) — scrollbar x=9 (width-1) = right arrow.
	ev := clickEvent(10, 1, tcell.Button1)
	win.HandleEvent(ev)

	if !fired {
		t.Error("Horizontal ScrollBar OnChange not fired after right-arrow click through Window")
	}
	if firedWith != 6 {
		t.Errorf("Horizontal ScrollBar OnChange value = %d, want 6", firedWith)
	}
}
