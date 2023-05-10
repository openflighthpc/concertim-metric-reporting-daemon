package ticker

import (
	"errors"
	"time"
)

// Ticker is a wrapper around a time.Ticker.  In addition to the periodic ticks
// provided by time.Ticker, it also has a TickNow() method which causes an
// immediate attempt to tick.  If that tick is successful, the wrapped
// time.Ticker is reset with its existing duration.  The reasons that a call to
// TickNow might not cause a successful tick are documented with that method.
type Ticker struct {
	C        <-chan time.Time
	c        chan<- time.Time
	throttle time.Duration
	interval time.Duration
	lastTick time.Time
	stopped  bool
	ticker   *time.Ticker
}

// NewTicker returns a Ticker which ticks with the given interval and throttles
// any calls to TickNow with the given throttle.
func NewTicker(interval, throttle time.Duration) *Ticker {
	if interval <= 0 {
		panic(errors.New("non-positive interval for NewTicker"))
	}
	// Buffer a single tick if the client is not listening.  Any additional
	// ticks are dropped on the floor.
	c := make(chan time.Time, 1)
	t := &Ticker{
		C:        c,
		c:        c,
		throttle: throttle,
		interval: interval,
		stopped:  false,
		ticker:   time.NewTicker(interval),
	}
	go func() {
		for {
			now := <-t.ticker.C
			t.tick(now)
		}
	}()
	return t
}

// TickNow makes an attempt to immediately tick. The attempt to tick may be
// unsuccessful for three reasons:
//
// 1. The ticker has been stopped and not resumed.
// 2. The ticker has ticked within throttle of the call to TickNow.
// 3. The client is not currently waiting on the chan and there is already a
//    queued tick.
func (t *Ticker) TickNow() bool {
	if !t.stopped && time.Since(t.lastTick) > t.throttle {
		if t.tick(time.Now()) {
			t.Resume()
			return true
		}
	}
	return false
}

// Stop stops the ticker.  No further ticks will take place until Resume() is
// called.
func (t *Ticker) Stop() {
	t.stopped = true
	t.ticker.Stop()
}

// Resume resumes the ticker.  It will tick the previously given interval.
func (t *Ticker) Resume() {
	t.stopped = false
	t.ticker.Reset(t.interval)
}

// tick does a non-blocking send of the given time to t.c. Returns true if the
// send succeeded and false if it was dropped.
func (t *Ticker) tick(tt time.Time) bool {
	select {
	case t.c <- tt:
		t.lastTick = tt
		return true
	default:
	}
	return false
}
