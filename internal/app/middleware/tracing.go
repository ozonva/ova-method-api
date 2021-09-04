package middleware

import (
	"context"

	tracer "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

type tracingMiddleware struct {
	allowMethods map[string]string
}

func NewTracingMiddleware(allowMethods map[string]string) *tracingMiddleware {
	return &tracingMiddleware{allowMethods: allowMethods}
}

func (tm *tracingMiddleware) UnaryIntercept(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if !tm.allowTracing(info.FullMethod) {
		return handler(ctx, req)
	}

	parentSpan, ctx := tracer.StartSpanFromContext(ctx, tm.getAliasByMethod(info.FullMethod))
	defer parentSpan.Finish()

	return handler(ctx, req)
}

func (tm *tracingMiddleware) allowTracing(methodName string) bool {
	_, ok := tm.allowMethods[methodName]
	return ok
}

func (tm *tracingMiddleware) getAliasByMethod(methodName string) string {
	return tm.allowMethods[methodName]
}
