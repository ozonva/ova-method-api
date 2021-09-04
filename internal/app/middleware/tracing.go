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

func (middleware *tracingMiddleware) UnaryIntercept(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if !middleware.allowTracing(info.FullMethod) {
		return handler(ctx, req)
	}

	parentSpan, ctx := tracer.StartSpanFromContext(ctx, middleware.getAliasByMethod(info.FullMethod))
	defer parentSpan.Finish()

	return handler(ctx, req)
}

func (middleware *tracingMiddleware) allowTracing(methodName string) bool {
	_, ok := middleware.allowMethods[methodName]
	return ok
}

func (middleware *tracingMiddleware) getAliasByMethod(methodName string) string {
	return middleware.allowMethods[methodName]
}
