package trace

import (
	"context"
	"fmt"
	"time"

	"github.com/TykTechnologies/opentelemetry/config"
	"go.opentelemetry.io/otel"
	noopMetricProvider "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Provider is the interface that wraps the basic methods of a tracer provider.
// If missconfigured or disabled, the provider will return a noop tracer
type Provider interface {
	// Shutdown execute the underlying exporter shutdown function
	Shutdown(context.Context) error
	// Tracer returns a tracer with pre-configured name. It's used to create spans.
	Tracer() Tracer
	// Type returns the type of the provider, it can be either "noop" or "otel"
	Type() string
}

type Tracer = oteltrace.Tracer

const (
	NOOP_PROVIDER = "noop"
	OTEL_PROVIDER = "otel"
)

type traceProvider struct {
	traceProvider      oteltrace.TracerProvider
	providerShutdownFn func(context.Context) error

	cfg    *config.OpenTelemetry
	logger Logger

	ctx          context.Context
	providerType string

	resources resourceConfig
}

/*
	 NewProvider creates a new tracer provider with the given options
	 The tracer provider is responsible for creating spans and sending them to the exporter

	 Example
		provider, err := trace.NewProvider(
			trace.WithContext(context.Background()),
			trace.WithConfig(&config.OpenTelemetry{
				Enabled:  true,
				Exporter: "grpc",
				Endpoint: "localhost:4317",
			}),
			trace.WithLogger(logrus.New().WithField("component", "tyk")),
		)
		if err != nil {
			panic(err)
		}
*/
func NewProvider(opts ...Option) (Provider, error) {
	provider := &traceProvider{
		traceProvider:      oteltrace.NewNoopTracerProvider(),
		providerShutdownFn: nil,
		logger:             &noopLogger{},
		cfg:                &config.OpenTelemetry{},
		ctx:                context.Background(),
		providerType:       NOOP_PROVIDER,
	}

	// apply the given options
	for _, opt := range opts {
		opt.apply(provider)
	}

	// set the config defaults - this does not override the config values
	provider.cfg.SetDefaults()

	// if the provider is not enabled, return a noop provider
	if !provider.cfg.Enabled {
		return provider, nil
	}

	// create the resource
	resource, err := resourceFactory(provider.ctx, provider.cfg.ResourceName, provider.resources)
	if err != nil {
		provider.logger.Error("failed to create exporter", err)
		return provider, fmt.Errorf("failed to create resource: %w", err)
	}

	// create the exporter - here's where connecting to the collector happens
	exporter, err := exporterFactory(provider.ctx, provider.cfg)
	if err != nil {
		provider.logger.Error("failed to create exporter", err)
		return provider, fmt.Errorf("failed to create exporter: %w", err)
	}

	// create the span processor - this is what will send the spans to the exporter.
	spanProcesor := spanProcessorFactory(provider.cfg.SpanProcessorType, exporter)

	// create the sampler based on the configs
	samplerType := provider.cfg.Sampling.Type
	samplingRate := provider.cfg.Sampling.Rate
	parentBasedSampling := provider.cfg.Sampling.ParentBased
	sampler := getSampler(samplerType, samplingRate, parentBasedSampling)

	// Create the tracer provider
	// The tracer provider will use the resource and exporter created previously
	// to generate spans and send them to the exporter
	// The tracer provider must be registered as a global tracer provider
	// so that any other package can use it

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(resource),
		sdktrace.WithSpanProcessor(spanProcesor),
	)

	propagator, err := propagatorFactory(provider.cfg)
	if err != nil {
		provider.logger.Error("failed to create context propagator", err)
		return provider, fmt.Errorf("failed to create context propagator: %w", err)
	}

	// set the local tracer provider
	provider.traceProvider = tracerProvider
	provider.providerShutdownFn = tracerProvider.Shutdown
	provider.providerType = OTEL_PROVIDER

	// set global otel tracer provider
	otel.SetTracerProvider(tracerProvider)

	otel.SetMeterProvider(noopMetricProvider.NewMeterProvider())

	// set the global otel context propagator
	otel.SetTextMapPropagator(propagator)

	// set the global otel error handler
	otel.SetErrorHandler(&errHandler{
		logger: provider.logger,
	})

	provider.logger.Info("Tracer provider initialized successfully")

	return provider, nil
}

func (tp *traceProvider) Shutdown(ctx context.Context) error {
	if tp.providerShutdownFn == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(tp.cfg.ConnectionTimeout)*time.Second)
	defer cancel()

	return tp.providerShutdownFn(ctx)
}

func (tp *traceProvider) Tracer() Tracer {
	return tp.traceProvider.Tracer(tp.cfg.ResourceName)
}

func (tp *traceProvider) Type() string {
	return tp.providerType
}
