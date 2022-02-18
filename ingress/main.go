package main

import (
	"context"
	"fmt"
	"github.com/teken/GoMicroService/chassis"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"time"
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

	http.HandleFunc("/", handleRequest(routeManager))

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

func handleRequest(routeManager *RouteManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		route := routeManager.GetRoute(method, path)
		if route == nil {
			w.WriteHeader(404) //todo: fill out
			return
		}

		awaitResponse, err := route.SendRequest(r)
		if err != nil {
			w.WriteHeader(500) //todo: fill out
			return
		}

		select {
		case resp := <-awaitResponse:
			w.Header().Add("Content-Type", resp.ContentType)
			w.WriteHeader(resp.StatusCode)
			w.Write(resp.Body)
		case <-time.After(5 * time.Second):
			w.WriteHeader(408)
		}

	}
}
