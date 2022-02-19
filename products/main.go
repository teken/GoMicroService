package main

import (
	"context"
	"fmt"
	"github.com/asdine/storm/v3"
	"github.com/teken/GoMicroService/chassis"
)

func main() {
	c := chassis.NewEventSourceChassis("Product Service", "products_service")
	tracer := c.ConfigureOpenTelemetryWithStdOut()
	ctx := context.Background()
	tracer.Start(ctx, "main")

	db, err := storm.Open("products.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	r := requestHandlers{db}
	c.Requests.Get("/products", r.GetAll)
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
