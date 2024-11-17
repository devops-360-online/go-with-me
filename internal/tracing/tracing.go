package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

func InitTracer() (*trace.TracerProvider, error) {
	// Create OTLP exporter with a 5-second timeout and insecure connection
	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint("otel-collector:4317"),
		otlptracegrpc.WithInsecure(),             // Use insecure connection (no TLS)
		otlptracegrpc.WithTimeout(5*time.Second), // Timeout for the connection
	)
	if err != nil {
		return nil, err
	}

	// Create a new TracerProvider with the OTLP exporter and resource attributes
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("event-service"),            // Name of your service
			semconv.ServiceVersionKey.String("v1.0.0"),            // Version of the service
			semconv.DeploymentEnvironmentKey.String("production"), // Environment
		)),
	)

	// Set the global tracer provider
	otel.SetTracerProvider(tp)
	return tp, nil
}

// ShutdownTracer gracefully shuts down the TracerProvider, flushing any remaining spans.
func ShutdownTracer(tp *trace.TracerProvider) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return tp.Shutdown(ctx)
}
