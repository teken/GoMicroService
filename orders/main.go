package main

import (
	"context"
	"fmt"
	"github.com/teken/GoMicroService/chassis"
)

func main() {
	c := chassis.NewEventSourceChassis("Order Service", "order_service")
	tracer := c.ConfigureOpenTelemetryWithStdOut()
	ctx := context.Background()
	tracer.Start(ctx, "main")

	r := requests{}
	c.Requests.Unhandled(r.Unhandled)

	if ready, err := c.ReadyAndServe(ctx); err != nil {
		fmt.Println("Failed to Start: " + err.Error())
	} else {
		<-ready
	}
}
