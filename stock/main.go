package main

import (
	"context"
	"fmt"
	"github.com/asdine/storm/v3"
	"github.com/teken/GoMicroService/chassis"
	orderEvents "github.com/teken/GoMicroService/orders/events"
	productEvents "github.com/teken/GoMicroService/products/events"
)

func main() {
	c := chassis.NewEventSourceChassis("Stock Service", "stock_service")
	tracer := c.ConfigureOpenTelemetryWithStdOut()
	ctx := context.Background()
	tracer.Start(ctx, "main")

	db, err := storm.Open("stock.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	r := requestHandlers{db}
	c.Requests.Get("/products/{id:uuid}/stock", r.Get)
	c.Requests.Put("/products/{id:uuid}/stock", r.Update)
	c.Requests.Unhandled(r.Unhandled)

	e := eventHandlers{db}
	c.Events.Subscribe(orderEvents.OrderCreated, e.orderCreated)
	c.Events.Subscribe(orderEvents.OrderCancelled, e.orderCancelled)
	c.Events.Subscribe(orderEvents.OrderCompleted, e.orderCompleted)

	c.Events.Subscribe(productEvents.ProductCreated, e.productCreated)
	c.Events.Subscribe(productEvents.ProductDeleted, e.productDeleted)

	if ready, err := c.ReadyAndServe(ctx); err != nil {
		fmt.Println("Failed to Start: " + err.Error())
	} else {
		<-ready
	}
}
