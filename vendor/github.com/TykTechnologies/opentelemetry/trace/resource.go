package trace

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

type resourceConfig struct {
	id      string
	version string

	withHost      bool
	withContainer bool
	withProcess   bool

	customAttrs []Attribute
}

func resourceFactory(ctx context.Context, resourceName string, cfg resourceConfig) (*resource.Resource, error) {
	opts := []resource.Option{}

	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(resourceName),
	}

	if cfg.id != "" {
		attrs = append(attrs, semconv.ServiceInstanceID(cfg.id))
	}

	if cfg.version != "" {
		attrs = append(attrs, semconv.ServiceVersion(cfg.version))
	}

	// add custom attributes
	attrs = append(attrs, cfg.customAttrs...)

	opts = append(opts, resource.WithAttributes(attrs...))

	if cfg.withContainer {
		opts = append(opts, resource.WithContainer())
	}

	if cfg.withHost {
		opts = append(opts, resource.WithHost())
	}

	if cfg.withProcess {
		// adding all the resource.WithProcess() options, except WithProcessOwner() since it's failing in k8s environments
		opts = append(opts, resource.WithProcessPID(),
			resource.WithProcessExecutableName(),
			resource.WithProcessCommandArgs(),
			resource.WithProcessRuntimeName(),
			resource.WithProcessRuntimeVersion(),
			resource.WithProcessRuntimeDescription())
	}

	return resource.New(ctx, opts...)
}
