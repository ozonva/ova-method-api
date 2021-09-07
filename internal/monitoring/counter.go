package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
)

type StatusCounter interface {
	Inc(status string, endpoint string) bool
}

type baseStatusCounter struct {
	status    string
	endpoints map[string]struct{}
	counter   prometheus.Counter
}

func NewStatusCounter(status string, endpoints []string, counter prometheus.Counter) StatusCounter {
	allows := make(map[string]struct{}, len(endpoints))
	for _, endpoint := range endpoints {
		allows[endpoint] = struct{}{}
	}

	return &baseStatusCounter{status: status, endpoints: allows, counter: counter}
}

func (base *baseStatusCounter) Inc(status string, endpoint string) bool {
	if base.status != status {
		return false
	}
	if _, ok := base.endpoints[endpoint]; !ok {
		return false
	}

	base.counter.Inc()
	return true
}
