package main

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/teken/GoMicroService/chassis"
)

func main() {
	c := chassis.NewEventSourceChassis("Sump Service", "sump_service")
	tracer := c.ConfigureOpenTelemetryWithStdOut()
	ctx := context.Background()
	tracer.Start(ctx, "main")

	ch, err := c.Communication.Connection.Channel()
	if err != nil {
		panic(err)
	}

	err = ch.ExchangeDeclare("events", "fanout", false, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	q, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		panic(err)
	}

	err = ch.QueueBind(q.Name, "", "events", false, nil)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(q.Name, "events.sump", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	go consumeRabbit(msgs)

	if ready, err := c.ReadyAndServe(ctx); err != nil {
		fmt.Println("Failed to Start: " + err.Error())
	} else {
		<-ready
	}
}

func consumeRabbit(msgs <-chan amqp.Delivery) {
	for msg := range msgs {

	}
}
