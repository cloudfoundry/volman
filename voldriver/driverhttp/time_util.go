package driverhttp

import (
	"time"

	"github.com/pivotal-golang/clock"
)

type ExponentialBackOff struct {
	// After MaxElapsedTime the ExponentialBackOff stops.
	// It never stops if MaxElapsedTime == 0.
	MaxElapsedTime time.Duration
	clock          clock.Clock

	currentInterval time.Duration
	startTime       time.Time
}

const Stop time.Duration = -1

// Default values for ExponentialBackOff.
const (
	InitialInterval       = 500 * time.Millisecond
	Multiplier            = 1.5
	DefaultMaxElapsedTime = 30 * time.Second
)

func NewExponentialBackOff(clock clock.Clock) *ExponentialBackOff {
	b := &ExponentialBackOff{
		MaxElapsedTime: DefaultMaxElapsedTime,
		clock:          clock,
	}

	b.Reset()
	return b
}

func (b *ExponentialBackOff) Reset() {
	b.currentInterval = InitialInterval
	b.startTime = b.clock.Now()
}

func (b *ExponentialBackOff) NextBackOff() time.Duration {
	if b.MaxElapsedTime != 0 && b.GetElapsedTime() > b.MaxElapsedTime {
		return Stop
	}
	defer b.incrementCurrentInterval()
	if b.MaxElapsedTime != 0 && (b.GetElapsedTime()+b.currentInterval) > b.MaxElapsedTime {
		// only wait until we hit max elapsed time before quitting
		return time.Millisecond + b.MaxElapsedTime - b.GetElapsedTime()
	}
	return b.currentInterval
}

func (b *ExponentialBackOff) GetElapsedTime() time.Duration {
	return b.clock.Now().Sub(b.startTime)
}

func (b *ExponentialBackOff) incrementCurrentInterval() {
	b.currentInterval = time.Duration(float64(b.currentInterval) * Multiplier)
}

type Operation func() error

type Notify func(error, time.Duration)

func Retry(o Operation, b *ExponentialBackOff) error { return RetryNotify(o, b, nil) }

func RetryNotify(operation Operation, b *ExponentialBackOff, notify Notify) error {
	var err error
	var next time.Duration

	b.Reset()
	for {
		if err = operation(); err == nil {
			return nil
		}

		if next = b.NextBackOff(); next == Stop {
			return err
		}

		if notify != nil {
			notify(err, next)
		}

		b.clock.Sleep(next)
	}
}
