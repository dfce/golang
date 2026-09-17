package health

import (
	"context"
	"testing"
	"time"
)

type stubChecker struct {
	name   string
	result CheckResult
}

func (c stubChecker) Name() string {
	return c.name
}

func (c stubChecker) Check(context.Context) CheckResult {
	return c.result
}

func TestServiceMarksRequiredFailureAsNotReady(t *testing.T) {
	service := NewService([]Checker{
		stubChecker{
			name: "database",
			result: CheckResult{
				Status:   StatusFailed,
				Required: true,
				Error:    "connection refused",
			},
		},
		stubChecker{
			name: "redis",
			result: CheckResult{
				Status:   StatusDisabled,
				Required: false,
			},
		},
	}, time.Second, NewState(true))

	report := service.Check(context.Background())
	if report.Status != StatusNotReady {
		t.Fatalf("status = %q, want %q", report.Status, StatusNotReady)
	}
	if report.Checks["database"].Status != StatusFailed {
		t.Fatalf("database status = %q, want %q", report.Checks["database"].Status, StatusFailed)
	}
	if report.Checks["redis"].Status != StatusDisabled {
		t.Fatalf("redis status = %q, want %q", report.Checks["redis"].Status, StatusDisabled)
	}
}

func TestServiceSkipsChecksWhenStateIsNotReady(t *testing.T) {
	service := NewService([]Checker{
		stubChecker{
			name: "database",
			result: CheckResult{
				Status:   StatusOK,
				Required: true,
			},
		},
	}, time.Second, NewState(false))

	report := service.Check(context.Background())
	if report.Status != StatusNotReady {
		t.Fatalf("status = %q, want %q", report.Status, StatusNotReady)
	}
	if len(report.Checks) != 0 {
		t.Fatalf("checks = %#v, want empty checks while stopping", report.Checks)
	}
}
