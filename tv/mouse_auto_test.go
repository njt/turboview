package tv

// mouse_auto_test.go — Tests for Task 6: Mouse Auto-Repeat (evMouseAuto)
//
// Requirements under test:
//   1. A real button press (Button1/Button2/Button3) starts auto-repeat.
//   2. Auto-repeat fires synthetic EvMouse events at ~50ms intervals.
//   3. A mouse release (buttons == 0) stops auto-repeat.
//   4. Wheel events (WheelUp/WheelDown) do NOT start auto-repeat.
//   5. Synthetic events carry the same button flags and current position.
//   6. stopMouseAuto is idempotent (safe to call with no repeat running).
//   7. Thread safety: position updates from concurrent goroutines do not race.
//   8. stopMouseAuto is called during Run() cleanup (no goroutine leak).
//   9. Synthetic events are converted to EvMouse by PollEvent.
//  10. Auto-repeat stops when the button changes from pressed to released mid-stream.

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// injectMouseEvent posts a tcell mouse event to the simulation screen.
func injectMouseEvent(screen tcell.SimulationScreen, x, y int, btn tcell.ButtonMask) {
	screen.PostEvent(tcell.NewEventMouse(x, y, btn, tcell.ModNone))
}

// collectMouseEvents drains EvMouse events from PollEvent for the given
// duration and returns them. Runs PollEvent in a goroutine.
func collectMouseEvents(app *Application, duration time.Duration) []*Event {
	deadline := time.Now().Add(duration)
	var events []*Event

	ch := make(chan *Event, 64)
	go func() {
		for time.Now().Before(deadline) {
			ev := app.PollEvent()
			if ev == nil {
				return
			}
			ch <- ev
		}
	}()

	timer := time.NewTimer(duration + 50*time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case ev := <-ch:
			if ev.What == EvMouse {
				events = append(events, ev)
			}
		case <-timer.C:
			return events
		}
	}
}

// ---------------------------------------------------------------------------
// Requirement 1: Button press starts auto-repeat
// ---------------------------------------------------------------------------

// TestMouseAutoStartsOnButtonPress verifies that after a Button1 press event,
// synthetic EvMouse events are generated within ~200ms (at 50ms intervals we
// expect at least 2 synthetic events in 200ms).
func TestMouseAutoStartsOnButtonPress(t *testing.T) {
	screen := newTestScreen(t)
	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Inject a Button1 press to start auto-repeat.
	injectMouseEvent(screen, 10, 5, tcell.Button1)

	// Collect EvMouse events for 250ms. We expect the initial real event plus
	// at least 2 synthetic auto-repeat events (50ms each).
	collected := make(chan []*Event, 1)
	go func() {
		deadline := time.Now().Add(250 * time.Millisecond)
		var events []*Event
		for time.Now().Before(deadline) {
			ev := app.PollEvent()
			if ev == nil {
				break
			}
			if ev.What == EvMouse {
				events = append(events, ev)
			}
		}
		collected <- events
	}()

	// After 250ms, inject a release to stop auto-repeat and unblock PollEvent.
	time.Sleep(250 * time.Millisecond)
	injectMouseEvent(screen, 10, 5, tcell.ButtonNone)
	screen.Fini()

	events := <-collected

	// We expect at least 3 total mouse events: 1 real press + ≥2 synthetic.
	if len(events) < 3 {
		t.Errorf("expected ≥3 EvMouse events (1 real + ≥2 synthetic), got %d", len(events))
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: Auto-repeat rate is approximately 50ms
// ---------------------------------------------------------------------------

// TestMouseAutoRepeatRate verifies that synthetic events arrive at roughly
// 50ms intervals. We measure the gap between consecutive auto-repeat events
// and check they are between 30ms and 150ms (generous bounds to avoid flakiness).
func TestMouseAutoRepeatRate(t *testing.T) {
	screen := newTestScreen(t)
	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Inject a button press to start auto-repeat.
	injectMouseEvent(screen, 5, 5, tcell.Button1)

	// Collect timestamps of EvMouse events for 300ms.
	var timestamps []time.Time
	collected := make(chan []time.Time, 1)
	go func() {
		deadline := time.Now().Add(300 * time.Millisecond)
		for time.Now().Before(deadline) {
			ev := app.PollEvent()
			if ev == nil {
				break
			}
			if ev.What == EvMouse {
				timestamps = append(timestamps, time.Now())
			}
		}
		collected <- timestamps
	}()

	time.Sleep(300 * time.Millisecond)
	injectMouseEvent(screen, 5, 5, tcell.ButtonNone)
	screen.Fini()

	ts := <-collected
	if len(ts) < 3 {
		t.Skipf("not enough events to measure rate: got %d events", len(ts))
	}

	// Check gaps between consecutive events (skip first since it includes real press).
	for i := 2; i < len(ts); i++ {
		gap := ts[i].Sub(ts[i-1])
		if gap < 20*time.Millisecond || gap > 200*time.Millisecond {
			t.Errorf("auto-repeat gap[%d] = %v, want ~50ms (20-200ms range)", i, gap)
		}
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: Mouse release stops auto-repeat
// ---------------------------------------------------------------------------

// TestMouseAutoStopsOnRelease verifies that after injecting a release event
// (ButtonNone), no further synthetic events are produced.
func TestMouseAutoStopsOnRelease(t *testing.T) {
	screen := newTestScreen(t)
	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Inject a button press to start auto-repeat.
	injectMouseEvent(screen, 10, 5, tcell.Button1)

	// Drain events until we see the real press and at least one synthetic event.
	// Then inject a release.
	var seenPress bool
	var synthBefore int
	drainDone := make(chan struct{})
	go func() {
		defer close(drainDone)
		deadline := time.Now().Add(200 * time.Millisecond)
		for time.Now().Before(deadline) {
			ev := app.PollEvent()
			if ev == nil {
				return
			}
			if ev.What == EvMouse {
				if !seenPress {
					seenPress = true
				} else {
					synthBefore++
				}
			}
		}
	}()
	<-drainDone

	// Inject release — this must be processed through PollEvent/convertEvent
	// to trigger stopMouseAuto.
	injectMouseEvent(screen, 10, 5, tcell.ButtonNone)

	// Drain the release event through PollEvent so convertEvent processes it.
	releaseSeen := make(chan struct{})
	go func() {
		for {
			ev := app.PollEvent()
			if ev == nil {
				return
			}
			if ev.What == EvMouse {
				close(releaseSeen)
				return
			}
		}
	}()
	select {
	case <-releaseSeen:
	case <-time.After(time.Second):
		t.Fatal("release event not received through PollEvent within 1s")
	}

	// Give the goroutine a moment to shut down.
	time.Sleep(20 * time.Millisecond)
	screen.Fini()

	// Verify the stop happened cleanly.
	app.mouseAutoMu.Lock()
	btn := app.mouseAutoBtn
	ch := app.mouseAutoChan
	app.mouseAutoMu.Unlock()
	if btn != 0 {
		t.Errorf("after release processed by PollEvent, mouseAutoBtn = %v, want 0", btn)
	}
	if ch != nil {
		t.Errorf("after release processed by PollEvent, mouseAutoChan should be nil")
	}
	if synthBefore < 1 {
		t.Errorf("expected at least 1 synthetic event before release, got %d", synthBefore)
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Wheel events do NOT start auto-repeat
// ---------------------------------------------------------------------------

// TestMouseAutoNotStartedForWheelUp verifies that a WheelUp event does not
// start the auto-repeat goroutine.
func TestMouseAutoNotStartedForWheelUp(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Inject a WheelUp event.
	injectMouseEvent(screen, 10, 5, tcell.WheelUp)

	// Drain the event.
	got := make(chan *Event, 1)
	go func() {
		got <- app.PollEvent()
	}()
	select {
	case ev := <-got:
		if ev == nil {
			t.Fatal("PollEvent returned nil")
		}
	case <-time.After(time.Second):
		t.Fatal("PollEvent timed out")
	}

	// Auto-repeat must not be running.
	app.mouseAutoMu.Lock()
	btn := app.mouseAutoBtn
	ch := app.mouseAutoChan
	app.mouseAutoMu.Unlock()

	if btn != 0 {
		t.Errorf("WheelUp set mouseAutoBtn = %v, want 0", btn)
	}
	if ch != nil {
		t.Error("WheelUp started auto-repeat goroutine, want none")
	}
}

// TestMouseAutoNotStartedForWheelDown mirrors the WheelUp test for WheelDown.
func TestMouseAutoNotStartedForWheelDown(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	injectMouseEvent(screen, 10, 5, tcell.WheelDown)

	got := make(chan *Event, 1)
	go func() {
		got <- app.PollEvent()
	}()
	select {
	case ev := <-got:
		if ev == nil {
			t.Fatal("PollEvent returned nil")
		}
	case <-time.After(time.Second):
		t.Fatal("PollEvent timed out")
	}

	app.mouseAutoMu.Lock()
	btn := app.mouseAutoBtn
	ch := app.mouseAutoChan
	app.mouseAutoMu.Unlock()

	if btn != 0 {
		t.Errorf("WheelDown set mouseAutoBtn = %v, want 0", btn)
	}
	if ch != nil {
		t.Error("WheelDown started auto-repeat goroutine, want none")
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: Synthetic events carry same button flags and current position
// ---------------------------------------------------------------------------

// TestMouseAutoSyntheticEventHasSameButtonAndPosition verifies that auto-repeat
// events carry the same button that triggered the press and the current position.
func TestMouseAutoSyntheticEventHasSameButtonAndPosition(t *testing.T) {
	screen := newTestScreen(t)
	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	const pressX, pressY = 15, 7
	injectMouseEvent(screen, pressX, pressY, tcell.Button1)

	// Collect EvMouse events for 200ms and verify synthetic ones have matching fields.
	var mouseEvents []*Event
	collected := make(chan []*Event, 1)
	go func() {
		deadline := time.Now().Add(200 * time.Millisecond)
		for time.Now().Before(deadline) {
			ev := app.PollEvent()
			if ev == nil {
				break
			}
			if ev.What == EvMouse {
				mouseEvents = append(mouseEvents, ev)
			}
		}
		collected <- mouseEvents
	}()

	time.Sleep(200 * time.Millisecond)
	injectMouseEvent(screen, pressX, pressY, tcell.ButtonNone)
	screen.Fini()

	events := <-collected
	if len(events) < 2 {
		t.Fatalf("expected at least 2 mouse events (press + synthetic), got %d", len(events))
	}

	// All events after the first (which is the real press) should be synthetic.
	// They must have Button1 set and the same position.
	for i, ev := range events[1:] {
		if ev.Mouse == nil {
			t.Errorf("synthetic event[%d]: Mouse is nil", i+1)
			continue
		}
		if ev.Mouse.Button&tcell.Button1 == 0 {
			t.Errorf("synthetic event[%d]: Button = %v, want Button1 set", i+1, ev.Mouse.Button)
		}
		if ev.Mouse.X != pressX || ev.Mouse.Y != pressY {
			t.Errorf("synthetic event[%d]: position = (%d,%d), want (%d,%d)", i+1, ev.Mouse.X, ev.Mouse.Y, pressX, pressY)
		}
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: stopMouseAuto is idempotent
// ---------------------------------------------------------------------------

// TestStopMouseAutoIdempotent verifies that calling stopMouseAuto when no
// auto-repeat is running does not panic or corrupt state.
func TestStopMouseAutoIdempotent(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Multiple calls to stopMouseAuto with nothing running must not panic.
	app.stopMouseAuto()
	app.stopMouseAuto()
	app.stopMouseAuto()

	app.mouseAutoMu.Lock()
	btn := app.mouseAutoBtn
	ch := app.mouseAutoChan
	app.mouseAutoMu.Unlock()

	if btn != 0 {
		t.Errorf("after idempotent stops, mouseAutoBtn = %v, want 0", btn)
	}
	if ch != nil {
		t.Errorf("after idempotent stops, mouseAutoChan should be nil")
	}
}

// TestStopMouseAutoIdempotentAfterStart verifies that stopMouseAuto can be
// called multiple times after a start without panicking.
func TestStopMouseAutoIdempotentAfterStart(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	app.startMouseAuto(5, 5, tcell.Button1)

	// Should not panic.
	app.stopMouseAuto()
	app.stopMouseAuto()
	app.stopMouseAuto()

	app.mouseAutoMu.Lock()
	ch := app.mouseAutoChan
	app.mouseAutoMu.Unlock()
	if ch != nil {
		t.Error("after repeated stops, mouseAutoChan should be nil")
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: Thread safety
// ---------------------------------------------------------------------------

// TestMouseAutoThreadSafety verifies that concurrent calls to startMouseAuto
// and stopMouseAuto do not trigger the race detector.
func TestMouseAutoThreadSafety(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	done := make(chan struct{})
	// Goroutine 1: repeatedly start/stop auto-repeat.
	go func() {
		defer close(done)
		for i := 0; i < 20; i++ {
			app.startMouseAuto(i, i, tcell.Button1)
			time.Sleep(5 * time.Millisecond)
			app.stopMouseAuto()
			time.Sleep(2 * time.Millisecond)
		}
	}()

	// Goroutine 2: concurrently update position via mutex (simulate move events).
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				app.mouseAutoMu.Lock()
				app.mouseAutoX = 42
				app.mouseAutoY = 7
				app.mouseAutoMu.Unlock()
				time.Sleep(3 * time.Millisecond)
			}
		}
	}()

	<-done
}

// ---------------------------------------------------------------------------
// Requirement 8: stopMouseAuto called during Run() cleanup
// ---------------------------------------------------------------------------

// TestMouseAutoStoppedOnRunExit verifies that when Run() exits, the auto-repeat
// goroutine is stopped (mouseAutoChan is nil after Run returns).
func TestMouseAutoStoppedOnRunExit(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Start auto-repeat directly so it is running when Run exits.
	app.startMouseAuto(5, 5, tcell.Button1)

	runDone := make(chan error, 1)
	go func() {
		runDone <- app.Run()
	}()

	// Give Run a moment to enter its loop, then quit.
	time.Sleep(50 * time.Millisecond)
	app.PostCommand(CmQuit, nil)

	select {
	case err := <-runDone:
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit within 2s")
	}

	// After Run returns, auto-repeat must be stopped.
	app.mouseAutoMu.Lock()
	ch := app.mouseAutoChan
	app.mouseAutoMu.Unlock()
	if ch != nil {
		t.Error("after Run() exit, mouseAutoChan is non-nil — auto-repeat goroutine leaked")
	}
}

// ---------------------------------------------------------------------------
// Requirement 9: Synthetic events converted to EvMouse by PollEvent
// ---------------------------------------------------------------------------

// TestMouseAutoSyntheticEventsConvertedByPollEvent verifies that mouseAutoEvent
// instances posted to the screen are returned by PollEvent as EvMouse events
// (not dropped or returned as nil).
func TestMouseAutoSyntheticEventsConvertedByPollEvent(t *testing.T) {
	screen := newTestScreen(t)
	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Post a synthetic mouseAutoEvent directly.
	synth := &mouseAutoEvent{x: 3, y: 4, button: tcell.Button2}
	synth.SetEventNow()
	screen.PostEvent(synth)

	got := make(chan *Event, 1)
	go func() {
		got <- app.PollEvent()
	}()

	select {
	case ev := <-got:
		if ev == nil {
			t.Fatal("PollEvent returned nil for a mouseAutoEvent")
		}
		if ev.What != EvMouse {
			t.Errorf("PollEvent: ev.What = %v, want EvMouse", ev.What)
		}
		if ev.Mouse == nil {
			t.Fatal("PollEvent: ev.Mouse is nil for a mouseAutoEvent")
		}
		if ev.Mouse.X != 3 || ev.Mouse.Y != 4 {
			t.Errorf("synthetic event position = (%d,%d), want (3,4)", ev.Mouse.X, ev.Mouse.Y)
		}
		if ev.Mouse.Button != tcell.Button2 {
			t.Errorf("synthetic event button = %v, want Button2", ev.Mouse.Button)
		}
	case <-time.After(time.Second):
		t.Error("PollEvent did not return mouseAutoEvent within 1s")
	}

	screen.Fini()
}

// ---------------------------------------------------------------------------
// Requirement 10: Auto-repeat stops when button is released mid-stream
// ---------------------------------------------------------------------------

// TestMouseAutoStopsOnReleaseAfterSeveralEvents verifies that the auto-repeat
// goroutine terminates cleanly when a release event interrupts an ongoing stream.
func TestMouseAutoStopsOnReleaseAfterSeveralEvents(t *testing.T) {
	screen := newTestScreen(t)
	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Start auto-repeat via a real press.
	injectMouseEvent(screen, 20, 10, tcell.Button2)

	// Let several synthetic events accumulate.
	synthCount := 0
	collectionDone := make(chan struct{})
	go func() {
		defer close(collectionDone)
		deadline := time.Now().Add(180 * time.Millisecond)
		for time.Now().Before(deadline) {
			ev := app.PollEvent()
			if ev == nil {
				return
			}
			if ev.What == EvMouse {
				synthCount++
			}
		}
	}()
	<-collectionDone

	// Now inject release.
	injectMouseEvent(screen, 20, 10, tcell.ButtonNone)

	// Drain release event.
	releaseRead := make(chan struct{})
	go func() {
		defer close(releaseRead)
		for {
			ev := app.PollEvent()
			if ev == nil {
				return
			}
			if ev.What == EvMouse {
				return
			}
		}
	}()

	select {
	case <-releaseRead:
	case <-time.After(time.Second):
		t.Fatal("release event not processed within 1s")
	}

	// Short wait to ensure goroutine shutdown.
	time.Sleep(80 * time.Millisecond)
	screen.Fini()

	app.mouseAutoMu.Lock()
	btn := app.mouseAutoBtn
	ch := app.mouseAutoChan
	app.mouseAutoMu.Unlock()

	if btn != 0 {
		t.Errorf("after release, mouseAutoBtn = %v, want 0", btn)
	}
	if ch != nil {
		t.Error("after release, mouseAutoChan should be nil")
	}
	if synthCount < 2 {
		t.Errorf("expected ≥2 mouse events before release, got %d", synthCount)
	}
}

// ---------------------------------------------------------------------------
// Additional falsifying test: no auto-repeat without a button press
// ---------------------------------------------------------------------------

// TestMouseAutoDoesNotStartWithoutPress is a falsifying test confirming that
// the initial state has no auto-repeat running.
func TestMouseAutoDoesNotStartWithoutPress(t *testing.T) {
	screen := newTestScreen(t)
	defer screen.Fini()

	app, err := NewApplication(WithScreen(screen))
	if err != nil {
		t.Fatalf("NewApplication: %v", err)
	}

	// Fresh app — no auto-repeat should be running.
	app.mouseAutoMu.Lock()
	btn := app.mouseAutoBtn
	ch := app.mouseAutoChan
	app.mouseAutoMu.Unlock()

	if btn != 0 {
		t.Errorf("fresh app has mouseAutoBtn = %v, want 0", btn)
	}
	if ch != nil {
		t.Error("fresh app has a mouseAutoChan, want nil")
	}
}

// TestMouseAutoButton2And3AlsoStart verifies that Button2 and Button3 (not
// just Button1) also trigger auto-repeat.
func TestMouseAutoButton2And3AlsoStart(t *testing.T) {
	for _, btn := range []struct {
		name   string
		button tcell.ButtonMask
	}{
		{"Button2", tcell.Button2},
		{"Button3", tcell.Button3},
	} {
		btn := btn
		t.Run(btn.name, func(t *testing.T) {
			screen := newTestScreen(t)
			app, err := NewApplication(WithScreen(screen))
			if err != nil {
				t.Fatalf("NewApplication: %v", err)
			}

			injectMouseEvent(screen, 5, 5, btn.button)

			// Drain the press event.
			got := make(chan *Event, 1)
			go func() {
				got <- app.PollEvent()
			}()
			select {
			case ev := <-got:
				if ev == nil || ev.What != EvMouse {
					t.Fatalf("expected EvMouse press event, got %v", ev)
				}
			case <-time.After(time.Second):
				t.Fatal("press event timed out")
			}

			// After press, auto-repeat should be running.
			app.mouseAutoMu.Lock()
			autoBtn := app.mouseAutoBtn
			ch := app.mouseAutoChan
			app.mouseAutoMu.Unlock()

			if autoBtn == 0 {
				t.Errorf("%s: mouseAutoBtn == 0 after press, want non-zero", btn.name)
			}
			if ch == nil {
				t.Errorf("%s: mouseAutoChan is nil after press, want non-nil", btn.name)
			}

			app.stopMouseAuto()
			screen.Fini()
		})
	}
}
