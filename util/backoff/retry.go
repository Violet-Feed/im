package backoff

import (
	"math"
	"time"
)

const Infinite = math.MaxUint64

type Operation func() error

type Notify func(error, time.Duration)

func Retry(o Operation, b BackOff) error { return RetryNotify(o, b, nil) }

func RetryNotify(operation Operation, b BackOff, notify Notify) error {
	var err error
	var next time.Duration
	var t *time.Timer

	cb := ensureContext(b)

	b.Reset()
	for {
		if err = operation(); err == nil {
			return nil
		}

		if permanent, ok := err.(*PermanentError); ok {
			return permanent.Err
		}

		if next = cb.NextBackOff(); next == Stop {
			return err
		}

		if notify != nil {
			notify(err, next)
		}

		if t == nil {
			t = time.NewTimer(next)
			defer t.Stop()
		} else {
			t.Reset(next)
		}

		select {
		case <-cb.Context().Done():
			return err
		case <-t.C:
		}
	}
}

type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	return e.Err.Error()
}

func Permanent(err error) *PermanentError {
	return &PermanentError{
		Err: err,
	}
}
