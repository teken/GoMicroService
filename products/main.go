package main

import (
	"context"
	"fmt"
	"github.com/teken/GoMicroService/chassis"
)

func main() {
	c := chassis.NewEventSourceChassis("Product Service", "product_service")
	tracer := c.ConfigureOpenTelemetryWithStdOut()
	ctx := context.Background()
	tracer.Start(ctx, "main")

	r := requests{}
	c.Requests.Get("/products/{id:uuid}", r.Get)
	c.Requests.Post("/products", r.Create)
	c.Requests.Put("/products/{id:uuid}", r.Update)
	c.Requests.Delete("/products/{id:uuid}", r.Delete)
	c.Requests.Unhandled(r.Unhandled)

	if ready, err := c.ReadyAndServe(ctx); err != nil {
		fmt.Println("Failed to Start: " + err.Error())
	} else {
		<-ready
	}
}
