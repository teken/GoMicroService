package main

import (
	"context"
	"github.com/teken/GoMicroService/chassis"
	"net/http"
)

func main() {
	c := chassis.NewEventSourceChassis("Ingress Service", "ingress_service")
	tracer := c.ConfigureOpenTelemetryWithStdOut()
	ctx := context.Background()
	tracer.Start(ctx, "main")

	routeManager := NewRouteManager(&c.Id, c.Communication)

	http.HandleFunc("/", handleRequest(routeManager))

	//signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	err := c.Communication.Connect(true)
	if err != nil {
		panic(err)
	}

	err = c.Requests.Serve()
	if err != nil {
		panic(err)
	}

	err = c.Events.Serve()
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}

	err = c.TraceProvider.Shutdown(ctx)
	if err != nil {
		panic(err)
	}
}

func handleRequest(routeManager *RouteManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		route := routeManager.GetRoute(method, path)

		awaitResponse, err := route.SendRequest(r)
		if err != nil {
			w.WriteHeader(500) //todo: fill out
			return
		}

		resp := <-awaitResponse

		w.Header().Add("Content-Type", resp.ContentType)
		w.WriteHeader(resp.StatusCode)
		w.Write(resp.Body)
	}
}
