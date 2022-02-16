package chassis

import (
	"context"
	"fmt"
	"github.com/pborman/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type EventSourceChassis struct {
	Id            uuid.UUID
	Communication *RabbitCommunication
	Requests      *Requests
	Events        *Events
	TraceProvider *sdktrace.TracerProvider

	*ServiceInfo
}

type ServiceInfo struct {
	displayName string
	serviceName string
}

func NewEventSourceChassis(displayName string, serviceName string) *EventSourceChassis {
	info := &ServiceInfo{
		displayName,
		serviceName,
	}
	com := &DefaultRabbitCommunication
	serviceId := uuid.NewRandom()
	return &EventSourceChassis{
		Id:            serviceId,
		Communication: com,
		Requests: &Requests{
			NewRequestManager(com, info, nil),
		},
		Events: &Events{
			NewEventManager(com, info, nil),
		},
		ServiceInfo: info,
	}
}

func (c *EventSourceChassis) ConfigureOpenTelemetryWithStdOut(attrs ...attribute.KeyValue) trace.Tracer {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("creating stdout exporter: %v", err)
	}
	return c.ConfigureOpenTelemetry(exporter, attrs...)
}

func (c *EventSourceChassis) ConfigureOpenTelemetry(exp sdktrace.SpanExporter, attrs ...attribute.KeyValue) trace.Tracer {
	attrs = append(attrs, semconv.ServiceNameKey.String(c.serviceName))
	currentResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	)

	c.TraceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(currentResource))

	otel.SetTracerProvider(c.TraceProvider)

	return c.TraceProvider.Tracer(c.serviceName)
}

func (c *EventSourceChassis) ReadyAndServe(ctx context.Context) (<-chan bool, error) {
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer c.TraceProvider.Shutdown(ctx)

		sig := <-sigs
		fmt.Println("Shutting Down: " + sig.String())
		done <- true
	}()

	err := c.Communication.Connect(true)
	if err != nil {
		return nil, err
	}

	err = c.Requests.Serve()
	if err != nil {
		return nil, err
	}

	err = c.Events.Serve()
	if err != nil {
		return nil, err
	}

	return done, nil
}
