package health

import (
	"context"
	"sync"
	"time"
)

const (
	StatusOK       = "ok"
	StatusFailed   = "failed"
	StatusDisabled = "disabled"
	StatusReady    = "ready"
	StatusNotReady = "not_ready"
)

type CheckResult struct {
	Status   string `json:"status"`
	Required bool   `json:"required"`
	Error    string `json:"error,omitempty"`
}

type Checker interface {
	Name() string
	Check(ctx context.Context) CheckResult
}

type Report struct {
	Status string                 `json:"status"`
	Checks map[string]CheckResult `json:"checks"`
}

type Service struct {
	checkers []Checker
	timeout  time.Duration
	state    *State
}

type State struct {
	mu    sync.RWMutex
	ready bool
}

func NewState(ready bool) *State {
	return &State{ready: ready}
}

func (s *State) SetReady(ready bool) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.ready = ready
	s.mu.Unlock()
}

func (s *State) Ready() bool {
	if s == nil {
		return true
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready
}

func NewService(checkers []Checker, timeout time.Duration, state *State) *Service {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &Service{
		checkers: checkers,
		timeout:  timeout,
		state:    state,
	}
}

func (s *Service) Check(ctx context.Context) Report {
	if s.state != nil && !s.state.Ready() {
		return Report{
			Status: StatusNotReady,
			Checks: make(map[string]CheckResult, len(s.checkers)),
		}
	}

	checkCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	report := Report{
		Status: StatusReady,
		Checks: make(map[string]CheckResult, len(s.checkers)),
	}

	type result struct {
		name  string
		value CheckResult
	}
	results := make(chan result, len(s.checkers))
	var waitGroup sync.WaitGroup

	for _, checker := range s.checkers {
		if checker == nil {
			continue
		}

		waitGroup.Add(1)
		go func(checker Checker) {
			defer waitGroup.Done()
			results <- result{
				name:  checker.Name(),
				value: checker.Check(checkCtx),
			}
		}(checker)
	}

	waitGroup.Wait()
	close(results)

	for item := range results {
		report.Checks[item.name] = item.value
		if item.value.Required && item.value.Status != StatusOK {
			report.Status = StatusNotReady
		}
	}

	return report
}
