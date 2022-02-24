package main

import (
	"context"
	"fmt"
	"github.com/teken/GoMicroService/chassis"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func main() {
	c := chassis.NewEventSourceChassis("Ingress Service", "ingress_service")
	tracer := c.ConfigureOpenTelemetryWithStdOut()
	ctx := context.Background()
	var span trace.Span
	ctx, span = tracer.Start(ctx, "main")

	routeManager := NewRouteManager(&c.Id, c.Communication)

	err := routeManager.AddRoute(http.MethodGet, "/products", "products_service")
	if err != nil {
		panic(err)
	}

	handler := newRequestHandler(routeManager)

	http.Handle("/", handler)

	//signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	err = c.Communication.Connect(true)
	if err != nil {
		panic(err)
	}
	span.AddEvent("Communication Started")

	//err = c.Events.Serve()
	//if err != nil {
	//	panic(err)
	//}
	//span.AddEvent("Events Serving")

	err = routeManager.StartResponseListening()
	if err != nil {
		panic(err)
	}
	span.AddEvent("Responses Listening")

	fmt.Println("Listening...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}

	err = c.TraceProvider.Shutdown(ctx)
	if err != nil {
		panic(err)
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
