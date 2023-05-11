package ticker

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTicker_ticksAsTimeTicker(t *testing.T) {
	// This test is based on tick_test.go licensed under a BSD-style
	// license.  It is essentiall testing time.Ticker, but also serves as a
	// starting point for testing our custom behaviour.

	// We want to test that a ticker takes as much time as expected.
	// Since we don't want the test to run for too long, we don't
	// want to use lengthy times. This makes the test inherently flaky.
	// Start with a short time, but try again with a long one if the
	// first test fails.

	baseCount := 10
	baseInterval := 20 * time.Millisecond

	var errs []string
	logErrs := func() {
		for _, e := range errs {
			t.Log(e)
		}
	}
	tests := []struct {
		count    int           // The number of ticks we're going to wait for.
		interval time.Duration // The interval between ticks.
	}{
		{count: baseCount, interval: baseInterval},
		{count: baseCount * 2, interval: baseInterval * 2},
		{count: baseCount, interval: baseInterval * 4},
		{count: baseCount * 4, interval: baseInterval},
	}

	for _, tt := range tests {
		count, interval := tt.count, tt.interval
		ticker := NewTicker(interval, time.Second)
		t0 := time.Now()

		// Wait until we've received the wanted number of ticks.
		for i := 0; i < count; i++ {
			<-ticker.C
		}

		ticker.Stop()
		t1 := time.Now()
		dt := t1.Sub(t0)
		target := time.Duration(count) * interval
		slop := target * 3 / 10
		if dt < target-slop || dt > target+slop {
			errs = append(
				errs,
				fmt.Sprintf("%d %s ticks took %s, expected [%s,%s]", count, interval, dt, target-slop, target+slop),
			)
			if dt > target+slop {
				// System may be overloaded; sleep a bit
				// in the hopes it will recover.
				time.Sleep(time.Second / 2)
			}
			continue
		}
		// Now test that the ticker stopped.
		time.Sleep(2 * interval)
		select {
		case <-ticker.C:
			errs = append(errs, "Ticker did not shut down")
			continue
		default:
			// ok
		}

		// Test passed, so all done.
		if len(errs) > 0 {
			t.Logf("saw %d errors, ignoring to avoid flakiness", len(errs))
			logErrs()
		}

		return
	}
	t.Errorf("saw %d errors", len(errs))
	logErrs()
}

func TestAStoppedTickerTicksNoMore(t *testing.T) {
	interval := 20 * time.Millisecond
	ticker := NewTicker(interval, time.Second)

	// Wait until we've received a tick.  That way we no that stop is causing it to stop.
	<-ticker.C
	ticker.Stop()

	// Now test that the ticker stopped.
	time.Sleep(2 * interval)
	select {
	case <-ticker.C:
		t.Error("Ticker did not shut down")
	default:
		// ok
	}
}

func TestAResumedTickerTicksSomeMore(t *testing.T) {
	interval := 20 * time.Millisecond
	ticker := NewTicker(interval, time.Second)

	// Wait until we've received a tick.  That way we no that stop is causing it to stop.
	<-ticker.C
	ticker.Stop()

	// Now test that the ticker stopped.
	time.Sleep(2 * interval)
	select {
	case <-ticker.C:
		t.Error("Ticker did not shut down")
	default:
		// ok
	}

	// Resume the ticker and check we get ticks again.
	ticker.Resume()
	time.Sleep(2 * interval)
	select {
	case <-ticker.C:
		// ok
	default:
		t.Error("Ticker did not resume")
	}
}

func TestTickNowTicksNow(t *testing.T) {
	interval := 20 * time.Millisecond
	throttle := time.Millisecond
	ticker := NewTicker(interval, throttle)

	<-ticker.C

	// We've just taken a tick, there shouldn't be another available yet.
	select {
	case <-ticker.C:
		t.Error("tick available to soon")
	default:
		// ok
	}

	<-ticker.C
	time.Sleep(throttle * 2)
	ticked := ticker.TickNow()
	if !ticked {
		t.Error("expected to have ticked")
	}

	// We've just manually ticked, there should be a tick waiting right now.
	select {
	case <-ticker.C:
		// ok
	default:
		t.Error("manual tick not available")
	}
}

func TestTickNowResetsTicker(t *testing.T) {
	interval := 20 * time.Millisecond
	throttle := time.Millisecond
	manualTickAfter := 5 * time.Millisecond
	ticker := NewTicker(interval, throttle)

	<-ticker.C

	// We've just taken a tick, there shouldn't be another available yet.
	select {
	case <-ticker.C:
		t.Error("tick available to soon")
	default:
		// ok
	}

	baseTick := <-ticker.C
	time.Sleep(manualTickAfter)
	ticked := ticker.TickNow()
	if !ticked {
		t.Error("expected to have ticked")
	}

	var manualTick time.Time

	// We've just manually ticked, there should be a tick waiting right now.
	select {
	case manualTick = <-ticker.C:
		// Check manual tick has arrived sooner than the ordinary interval tick would have done so.
		target := manualTickAfter
		slop := 2 * time.Millisecond
		dt := manualTick.Sub(baseTick)
		if dt < target-slop || dt > target+slop {
			t.Errorf("manual tick after %s, expected [%s,%s]", dt, target-slop, target+slop)
		}
		// ok
	default:
		t.Error("manual tick not available")
	}

	intervalTick := <-ticker.C
	// Check that the interval tick is interval ms after the manual tick.
	target := interval
	slop := 2 * time.Millisecond
	dt := intervalTick.Sub(manualTick)
	if dt < target-slop || dt > target+slop {
		t.Errorf("interval tick after %s, expected [%s,%s]", dt, target-slop, target+slop)
	}
}

func TestTickNowIsThrottled(t *testing.T) {
	interval := 20 * time.Millisecond
	throttle := 5 * time.Millisecond
	ticker := NewTicker(interval, throttle)

	<-ticker.C

	// We've just taken a tick, there shouldn't be another available yet.
	select {
	case <-ticker.C:
		t.Error("tick available to soon")
	default:
		// ok
	}

	baseTick := <-ticker.C
	time.Sleep(throttle - time.Millisecond)
	ticked := ticker.TickNow()
	if ticked {
		t.Error("expected tick to have been throttled")
	}

	// The tick should have been throttled.   There should be nothing waiting.
	select {
	case <-ticker.C:
		t.Error("manual tick available should have been throttled")
	default:
		// ok
	}

	// The next tick should arrive at the expected interval.
	intervalTick := <-ticker.C
	// Check that the interval tick is interval ms after the manual tick.
	target := interval
	slop := 2 * time.Millisecond
	dt := intervalTick.Sub(baseTick)
	if dt < target-slop || dt > target+slop {
		t.Errorf("interval tick after %s, expected [%s,%s]", dt, target-slop, target+slop)
	}
}

func TestStoppedTickersCannotBeTickedNow(t *testing.T) {
	interval := 20 * time.Millisecond
	throttle := time.Millisecond
	ticker := NewTicker(interval, throttle)

	// Wait until we've received a tick.  That way we no that stop is causing it to stop.
	<-ticker.C
	ticker.Stop()

	// Now test that the ticker stopped.
	time.Sleep(2 * throttle)
	ticked := ticker.TickNow()
	if ticked {
		t.Error("stopped ticker should not TickNow")
	}

	select {
	case <-ticker.C:
		t.Error("Ticker unexpectedly ticked")
	default:
		// ok
	}
}
