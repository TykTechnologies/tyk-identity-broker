package trace

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type Span trace.Span

// SpanFromContext returns the span attached to the given context.
// If the context does not have a span attached to it, a no-op span will be returned.
// Example:
//
// span := trace.SpanFromContext(ctx)
// defer span.End()
func SpanFromContext(ctx context.Context) Span {
	return trace.SpanFromContext(ctx)
}

// ContextWithSpan returns a child context from ctx with span embedded
func ContextWithSpan(ctx context.Context, span Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}

// NewSpanFromContext creates a new span from the given context.
// If the context already has a span attached to it, the new span will be a child of the existing span.
// If the context does not have a span attached to it, the new span will be a root span.
// The tracer name is used to identify the tracer to be used to create the span.
// If the tracer name is not provided, the default 'tyk' tracer name will be used.
// Example:
//
//	spanCtx, span := trace.NewSpanFromContext(ctx, "my-tracer", "my-span")
//	defer span.End()
func NewSpanFromContext(ctx context.Context, tracerName, spanName string) (context.Context, Span) {
	if tracerName == "" {
		tracerName = "tyk"
	}

	return SpanFromContext(ctx).TracerProvider().Tracer(tracerName).Start(ctx, spanName)
}
