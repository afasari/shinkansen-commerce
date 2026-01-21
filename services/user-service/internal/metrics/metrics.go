package metrics

import (
	"context"
	"time"
)

type (
	Counter interface {
		Inc()
	}

	Histogram interface {
		Observe(float64)
	}
)

var (
	UserRequests Counter   = &noopCounter{}
	UserLatency  Histogram = &noopHistogram{}
)

type noopCounter struct{}

func (n *noopCounter) Inc() {}

type noopHistogram struct{}

func (n *noopHistogram) Observe(float64) {}

func RecordRequest(ctx context.Context, method, status string) {
	UserRequests.Inc()
}

func RecordLatency(ctx context.Context, method string, duration time.Duration) {
	UserLatency.Observe(duration.Seconds())
}
