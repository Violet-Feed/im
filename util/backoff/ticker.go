package backoff

import (
	"sync"
	"time"
)

type Ticker struct {
	C        <-chan time.Time
	c        chan time.Time
	b        BackOffContext
	stop     chan struct{}
	stopOnce sync.Once
}

func NewTicker(b BackOff) *Ticker {
	c := make(chan time.Time)
	t := &Ticker{
		C:    c,
		c:    c,
		b:    ensureContext(b),
		stop: make(chan struct{}),
	}
	t.b.Reset()
	go t.run()
	return t
}

func (t *Ticker) Stop() {
	t.stopOnce.Do(func() { close(t.stop) })
}

func (t *Ticker) run() {
	c := t.c
	defer close(c)

	afterC := t.send(time.Now())

	for {
		if afterC == nil {
			return
		}

		select {
		case tick := <-afterC:
			afterC = t.send(tick)
		case <-t.stop:
			t.c = nil
			return
		case <-t.b.Context().Done():
			return
		}
	}
}

func (t *Ticker) send(tick time.Time) <-chan time.Time {
	select {
	case t.c <- tick:
	case <-t.stop:
		return nil
	}

	next := t.b.NextBackOff()
	if next == Stop {
		t.Stop()
		return nil
	}

	return time.After(next)
}
