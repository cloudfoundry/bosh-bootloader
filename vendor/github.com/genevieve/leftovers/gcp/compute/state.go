package compute

import (
	"errors"
	"fmt"
	"time"
)

type state struct {
	logger  logger
	refresh stateRefreshFunc
}

type stateRefreshFunc func() (result interface{}, state string, err error)

// Copied from terraform-provider-google implementation for compute operation polling.
func (s *state) Wait() error {
	notfoundTick := 0
	targetOccurence := 0
	notFoundChecks := 20
	delay := 10 * time.Second
	timeout := 10 * time.Minute
	minTimeout := 2 * time.Second
	refreshGracePeriod := 30 * time.Second

	type Result struct {
		Result interface{}
		State  string
		Error  error
		Done   bool
	}

	resultCh := make(chan Result, 1)
	cancellationCh := make(chan struct{})

	result := Result{}

	go func() {
		defer close(resultCh)

		time.Sleep(delay)

		var wait time.Duration

		for {
			resultCh <- result

			select {
			case <-cancellationCh:
				return
			case <-time.After(wait):
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
				resultCh <- result
				return
			}

			if res == nil {
				notfoundTick++
				if notfoundTick > notFoundChecks {
					result.Error = fmt.Errorf("Resource not found: %s", err)
					resultCh <- result
					return
				}
			} else {
				notfoundTick = 0
				found := false

				if currentState == "DONE" {
					found = true
					targetOccurence++
					if targetOccurence == 1 {
						result.Done = true
						resultCh <- result
						return
					}
					continue
				}

				if currentState == "PENDING" || currentState == "RUNNING" {
					found = true
					targetOccurence = 0
					break
				}

				if !found {
					result.Error = fmt.Errorf("Unexpected state: %s", err)
					resultCh <- result
					return
				}
			}

			if targetOccurence == 0 {
				wait *= 2
			}

			if wait < minTimeout {
				wait = minTimeout
			} else if wait > 10*time.Second {
				wait = 10 * time.Second
			}

			s.logger.Println("Waiting for operation to complete..")
		}
	}()

	lastResult := Result{}

	afterTimeout := time.After(timeout)
	for {
		select {
		case r, ok := <-resultCh:
			if !ok {
				return lastResult.Error
			}

			if r.Done {
				if r.Error != nil {
					return fmt.Errorf("Reached DONE state with error: %s", r.Error)
				}
				if r.Result == nil {
					return errors.New("Reached DONE state with no result.")
				}
				return nil
			}

			lastResult = r

		case <-afterTimeout:
			s.logger.Printf("Timeout after %s\n", timeout)
			s.logger.Printf("Starting %s refresh grace period\n", refreshGracePeriod)

			close(cancellationCh)
			afterTimeout := time.After(refreshGracePeriod)

		forSelect:
			for {
				select {
				case r, ok := <-resultCh:
					if r.Done {
						if r.Error != nil {
							return fmt.Errorf("Reached DONE state with error: %s", r.Error)
						}
						if r.Result == nil {
							return errors.New("Reached DONE state with no result.")
						}
						return nil
					}

					if !ok {
						break forSelect
					}

					lastResult = r
				case <-afterTimeout:
					s.logger.Printf("Exceeded refresh grace period\n")
					break forSelect
				}
			}

			return fmt.Errorf("Timeout error: %s", lastResult.Error)
		}
	}
}
