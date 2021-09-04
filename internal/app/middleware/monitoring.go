package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"ova-method-api/internal/monitoring"
)

type statusMonitoringMiddleware struct {
	counters []monitoring.StatusCounter
}

func NewStatusMonitoringMiddleware(counters []monitoring.StatusCounter) *statusMonitoringMiddleware {
	return &statusMonitoringMiddleware{counters: counters}
}

func (middleware *statusMonitoringMiddleware) UnaryIntercept(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	resp, err := handler(ctx, req)

	if st, ok := status.FromError(err); ok {
		middleware.updateCounters(st.Code().String(), info.FullMethod)
	}

	return resp, err
}

func (middleware *statusMonitoringMiddleware) updateCounters(status string, endpoint string) {
	for _, counter := range middleware.counters {
		counter.Inc(status, endpoint)
	}
}
