package main

import (
    "context"
    "log"
    "os"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func initTracer() func(context.Context) error {
    endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
    if endpoint == "" {
        endpoint = "otel-collector.monitoring.svc.cluster.local:4317"
    }

    conn, err := grpc.DialContext(context.Background(), endpoint,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithBlock(),
    )
    if err != nil {
        log.Printf("aviso: não conectou ao OTel Collector: %v", err)
        return func(context.Context) error { return nil }
    }

    exp, err := otlptracegrpc.New(context.Background(),
        otlptracegrpc.WithGRPCConn(conn),
    )
    if err != nil {
        log.Printf("aviso: falha ao criar exporter: %v", err)
        return func(context.Context) error { return nil }
    }

    res := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceName("auth-service"),
        semconv.ServiceVersion("1.0.0"),
    )

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )
    otel.SetTracerProvider(tp)

    return tp.Shutdown
}