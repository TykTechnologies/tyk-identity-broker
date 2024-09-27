package trace

import (
	"fmt"

	"github.com/TykTechnologies/opentelemetry/config"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"
)

func propagatorFactory(cfg *config.OpenTelemetry) (propagation.TextMapPropagator, error) {
	switch cfg.ContextPropagation {
	case config.PROPAGATOR_B3:
		propagator := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
		return propagator, nil
	case config.PROPAGATOR_TRACECONTEXT:
		return propagation.TraceContext{}, nil
	default:
		return nil, fmt.Errorf("invalid context propagation type: %s", cfg.ContextPropagation)
	}
}
