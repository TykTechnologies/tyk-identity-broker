package trace

import (
	"context"

	"github.com/TykTechnologies/opentelemetry/config"
)

type Option interface {
	apply(*traceProvider)
}

type opts struct {
	fn func(*traceProvider)
}

func (o *opts) apply(tp *traceProvider) {
	o.fn(tp)
}

/*
	WithConfig sets the configuration options for the tracer provider

Example

	config := &config.OpenTelemetry{
		Enabled:  true,
		Exporter: "grpc",
		Endpoint: "localhost:4317",
	}
	provider, err := trace.NewProvider(trace.WithConfig(config))
	if err != nil {
		panic(err)
	}
*/
func WithConfig(cfg *config.OpenTelemetry) Option {
	return &opts{
		fn: func(tp *traceProvider) {
			tp.cfg = cfg
		},
	}
}

/*
	WithLogger sets the logger for the tracer provider
	This is used to log errors and info messages for underlying operations

Example

	logger := logrus.New().WithField("component", "trace")
	provider, err := trace.NewProvider(trace.WithLogger(logger))
	if err != nil {
		panic(err)
	}
*/
func WithLogger(logger Logger) Option {
	return &opts{
		fn: func(tp *traceProvider) {
			tp.logger = logger
		},
	}
}

/*
	WithContext sets the context for the tracer provider

Example

	ctx := context.Background()
	provider, err := trace.NewProvider(trace.WithContext(ctx))
	if err != nil {
		panic(err)
	}
*/
func WithContext(ctx context.Context) Option {
	return &opts{
		fn: func(tp *traceProvider) {
			tp.ctx = ctx
		},
	}
}

/*
	WithServiceID sets the resource service.id for the tracer provider
	This is useful to identify service instance on the trace resource.

Example

	provider, err := trace.NewProvider(trace.WithServiceID("instance-id"))
	if err != nil {
		panic(err)
	}
*/
func WithServiceID(id string) Option {
	return &opts{
		fn: func(tp *traceProvider) {
			tp.resources.id = id
		},
	}
}

/*
	WithServiceVersion sets the resource service.version for the tracer provider
	This is useful to identify service version on the trace resource.

Example

	provider, err := trace.NewProvider(trace.WithServiceVersion("v4.0.5"))
	if err != nil {
		panic(err)
	}
*/
func WithServiceVersion(version string) Option {
	return &opts{
		fn: func(tp *traceProvider) {
			tp.resources.version = version
		},
	}
}

/*
	WithHostDetector adds attributes from the host to the configured resource.

Example

	provider, err := trace.NewProvider(trace.WithHostDetector())
	if err != nil {
		panic(err)
	}
*/
func WithHostDetector() Option {
	return &opts{
		fn: func(tp *traceProvider) {
			tp.resources.withHost = true
		},
	}
}

/*
	WithContainerDetector adds attributes from the container to the configured resource.

Example

	provider, err := trace.NewProvider(trace.WithContainerDetector())
	if err != nil {
		panic(err)
	}
*/
func WithContainerDetector() Option {
	return &opts{
		fn: func(tp *traceProvider) {
			tp.resources.withContainer = true
		},
	}
}

/*
	WithProcessDetector adds attributes from the process to the configured resource.

Example

	provider, err := trace.NewProvider(trace.WithProcessDetector())
	if err != nil {
		panic(err)
	}
*/

func WithProcessDetector() Option {
	return &opts{
		fn: func(tp *traceProvider) {
			tp.resources.withProcess = true
		},
	}
}

/*
	WithCustomResourceAttributes adds custom attributes to the configured resource.

Example

	attrs := []trace.Attribute{trace.NewAttribute("key", "value")}
	provider, err := trace.NewProvider(trace.WithCustomResourceAttributes(attrs...))
	if err != nil {
		panic(err)
	}
*/
func WithCustomResourceAttributes(attrs ...Attribute) Option {
	return &opts{
		fn: func(tp *traceProvider) {
			tp.resources.customAttrs = attrs
		},
	}
}
