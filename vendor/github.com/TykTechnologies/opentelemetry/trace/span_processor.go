package trace

import (
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func spanProcessorFactory(spanProcessorType string, exporter sdktrace.SpanExporter) sdktrace.SpanProcessor {
	switch spanProcessorType {
	case "simple":
		return newSimpleSpanProcessor(exporter)
	default:
		// Default to BatchSpanProcessor
		return newBatchSpanProcessor(exporter)
	}
}

func newSimpleSpanProcessor(exporter sdktrace.SpanExporter) sdktrace.SpanProcessor {
	return sdktrace.NewSimpleSpanProcessor(exporter)
}

func newBatchSpanProcessor(exporter sdktrace.SpanExporter) sdktrace.SpanProcessor {
	return sdktrace.NewBatchSpanProcessor(exporter)
}
