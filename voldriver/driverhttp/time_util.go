package driverhttp

import (
	"time"

	"code.cloudfoundry.org/clock"
)

const (
	backoffInitialInterval = 500 * time.Millisecond
	backoffIncrement       = 1.5
)

type exponentialBackOff struct {
	maxElapsedTime time.Duration
	clock          clock.Clock
}

// newExponentialBackOff takes a maximum elapsed time, after which the
// exponentialBackOff stops retrying the operation.
func newExponentialBackOff(maxElapsedTime time.Duration, clock clock.Clock) *exponentialBackOff {
	return &exponentialBackOff{
		maxElapsedTime: maxElapsedTime,
		clock:          clock,
	}
}

// Retry takes a retriable operation, and calls it until either the operation
// succeeds, or the retry timeout occurs.
func (b *exponentialBackOff) Retry(operation func() error) error {
	var (
		startTime       time.Time = b.clock.Now()
		backoffInterval time.Duration
		backoffExpired  bool
	)

	for {
		err := operation()
		if err == nil {
			return nil
		}

		backoffInterval, backoffExpired = b.incrementInterval(startTime, backoffInterval)
		if backoffExpired {
			return err
		}

		b.clock.Sleep(backoffInterval)
	}
}

func (b *exponentialBackOff) incrementInterval(startTime time.Time, currentInterval time.Duration) (nextInterval time.Duration, expired bool) {
	elapsedTime := b.clock.Now().Sub(startTime)

	if elapsedTime > b.maxElapsedTime {
		return 0, true
	}

	switch {
	case currentInterval == 0:
		nextInterval = backoffInitialInterval
	case elapsedTime+backoff(currentInterval) > b.maxElapsedTime:
		nextInterval = time.Millisecond + b.maxElapsedTime - elapsedTime
	default:
		nextInterval = backoff(currentInterval)
	}

	return nextInterval, false
}

func backoff(interval time.Duration) time.Duration {
	return time.Duration(float64(interval) * backoffIncrement)
}
