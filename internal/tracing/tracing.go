package tracing

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer() (*trace.TracerProvider, error) {
    exporter, err := otlptracehttp.New(context.Background())
    if err != nil {
        return nil, err
    }
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
    )
    otel.SetTracerProvider(tp)
    return tp, nil
}
