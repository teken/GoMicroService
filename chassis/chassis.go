package chassis

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"os"
	"os/signal"
	"syscall"
)

type EventSourceChassis struct {
	Requests      *Requests
	Events        *Events
	traceProvider *sdktrace.TracerProvider
}

func NewEventSourceChassis() *EventSourceChassis {
	return &EventSourceChassis{
		Requests: &Requests{
			NewRequestManager(nil),
		},
		Events: &Events{
			NewEventManager(nil),
		},
	}
}

func (c *EventSourceChassis) ConfigureOpenTelemetry(serviceName string, exp sdktrace.SpanExporter, attrs ...attribute.KeyValue) {
	attrs = append(attrs, semconv.ServiceNameKey.String(serviceName))
	currentResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	)
	c.traceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(currentResource))

	otel.SetTracerProvider(c.traceProvider)

	c.traceProvider.Tracer(serviceName)
}

func (c *EventSourceChassis) ReadyAndServe(ctx context.Context) <-chan bool {
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer c.traceProvider.Shutdown(ctx)

		sig := <-sigs
		fmt.Println("Shutting Down: " + sig.String())
		done <- true
	}()

	return done
}
