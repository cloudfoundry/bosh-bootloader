package common

import (
	"fmt"
	"time"
)

type logger interface {
	Printf(m string, a ...interface{})
}

type State struct {
	logger  logger
	refresh StateRefreshFunc
	pending []string
	target  []string
}

func NewState(logger logger, refresh StateRefreshFunc, pending, target []string) State {
	return State{
		logger:  logger,
		refresh: refresh,
		pending: pending,
		target:  target,
	}
}

type StateRefreshFunc func() (result interface{}, state string, err error)

var refreshGracePeriod = 30 * time.Second

// Copied from terraform-provider-google implementation for compute operation polling.
func (s *State) Wait() (interface{}, error) {
	notfoundTick := 0
	targetOccurence := 0
	notFoundChecks := 20
	continuousTargetOccurence := 1
	minTimeout := 2 * time.Second
	delay := 2 * time.Second

	type Result struct {
		Result interface{}
		State  string
		Error  error
		Done   bool
	}

	// Read every result from the refresh loop, waiting for a positive result.Done.
	resCh := make(chan Result, 1)
	// cancellation channel for the refresh loop
	cancelCh := make(chan struct{})

	result := Result{}

	go func() {
		defer close(resCh)

		time.Sleep(delay)

		// start with 0 delay for the first loop
		var wait time.Duration

		for {
			// store the last result
			resCh <- result

			// wait and watch for cancellation
			select {
			case <-cancelCh:
				return
			case <-time.After(wait):
				// first round had no wait
				if wait == 0 {
					wait = 100 * time.Millisecond
				}
			}

			res, currentState, err := s.refresh()
			result = Result{
				Result: res,
				State:  currentState,
				Error:  err,
			}

			if err != nil {
				resCh <- result
				return
			}

			if res == nil {
				// If we didn't find the resource, check if we have been
				// not finding it for awhile, and if so, report an error.
				notfoundTick++
				if notfoundTick > notFoundChecks {
					result.Error = fmt.Errorf("Resource not found: %s", err)
					resCh <- result
					return
				}
			} else {
				// Reset the counter for when a resource isn't found
				notfoundTick = 0
				found := false

				for _, allowed := range s.target {
					if currentState == allowed {
						found = true
						targetOccurence++
						if continuousTargetOccurence == targetOccurence {
							result.Done = true
							resCh <- result
							return
						}
						continue
					}
				}

				for _, allowed := range s.pending {
					if currentState == allowed {
						found = true
						targetOccurence = 0
						break
					}
				}

				if !found {
					result.Error = fmt.Errorf("Unexpected state %s: %s", result.State, err)
					resCh <- result
					return
				}
			}

			// Wait between refreshes using exponential backoff, except when
			// waiting for the target state to reoccur.
			if targetOccurence == 0 {
				wait *= 2
			}

			if wait < minTimeout {
				wait = minTimeout
			} else if wait > 10*time.Second {
				wait = 10 * time.Second
			}

			s.logger.Printf("Waiting %s before next try.\n", wait)
		}
	}()

	// store the last value result from the refresh loop
	lastResult := Result{}

	timeout := time.After(10 * time.Minute)
	for {
		select {
		case r, ok := <-resCh:
			// channel closed, so return the last result
			if !ok {
				return lastResult.Result, lastResult.Error
			}

			// we reached the intended state
			if r.Done {
				return r.Result, r.Error
			}

			// still waiting, store the last result
			lastResult = r

		case <-timeout:
			// cancel the goroutine and start our grace period timer
			close(cancelCh)
			timeout := time.After(refreshGracePeriod)

			// we need a for loop and a label to break on, because we may have
			// an extra response value to read, but still want to wait for the
			// channel to close.
		forSelect:
			for {
				select {
				case r, ok := <-resCh:
					if r.Done {
						// the last refresh loop reached the desired state
						return r.Result, r.Error
					}

					if !ok {
						// the goroutine returned
						break forSelect
					}

					// target state not reached, save the result for the
					// TimeoutError and wait for the channel to close
					lastResult = r
				case <-timeout:
					s.logger.Printf("Waiting for state %s exceeded refresh grace period.\n", s.target[0])
					break forSelect
				}
			}

			return nil, fmt.Errorf("Timeout waiting for state to be %s: %s", s.target[0], lastResult.Error)
		}
	}
}
