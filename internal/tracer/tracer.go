package tracer

import (
	"github.com/mojixcoder/caster/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTraceProvider initializes the trace provider
func InitTraceProvider(tc config.TracerConfig) error {
	exp, err := jaeger.New(
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(tc.JeagerAgent.Host),
			jaeger.WithAgentPort(tc.JeagerAgent.Port),
		),
	)
	if err != nil {
		return err
	}

	resource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(tc.Name),
		),
	)
	if err != nil {
		return err
	}

	sampler := tracesdk.ParentBased(tracesdk.TraceIDRatioBased(tc.Fraction))

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource),
		tracesdk.WithSampler(sampler),
	)
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagator)

	return nil
}
