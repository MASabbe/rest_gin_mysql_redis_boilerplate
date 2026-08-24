package domain

import (
	"context"
	"time"
)

// Status represents the operational status of the system or component.
type Status string

const (
	StatusUP       Status = "UP"
	StatusDOWN     Status = "DOWN"
	StatusDEGRADED Status = "DEGRADED"
)

// ComponentCheck holds the health check result of an individual component (e.g. MySQL, Redis).
type ComponentCheck struct {
	Name      string        `json:"name"`
	Status    Status        `json:"status"`
	Details   string        `json:"details,omitempty"`
	Latency   time.Duration `json:"latency_ns"`
	LatencyMs float64       `json:"latency_ms"`
	Error     string        `json:"error,omitempty"`
}

// HealthReport is the aggregate health result of the system.
type HealthReport struct {
	Status     Status                    `json:"status"`
	Timestamp  time.Time                 `json:"timestamp"`
	Components map[string]ComponentCheck `json:"components,omitempty"`
}

// Checker defines the interface required for probing an infrastructure dependency.
type Checker interface {
	Name() string
	Check(ctx context.Context) ComponentCheck
}
