package chassis

import (
	"errors"
	"fmt"
	"github.com/streadway/amqp"
)

type RabbitCommunication struct {
	address    string
	Connection *amqp.Connection
}

var DefaultRabbitCommunication = RabbitCommunication{
	address: "amqp://guest:guest@localhost:5672/",
}

func (c *RabbitCommunication) Connect(enableConnectionRecovery bool) error {
	conn, err := amqp.Dial(c.address)
	if err != nil {
		return err
	}
	c.Connection = conn
	if enableConnectionRecovery {
		err = c.EnableConnectionRecovery()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *RabbitCommunication) EnableConnectionRecovery() error {
	if c.Connection == nil {
		return errors.New("connection not initialised")
	}

	go func() {
		amqpErr, more := <-c.Connection.NotifyClose(make(chan *amqp.Error, 1))
		if !more { // if more is true then rabbit close was intentional
			fmt.Println("rabbit connection error: ", amqpErr)
			err := c.Connect(true)
			if err != nil {
				fmt.Println("failed to restart rabbit connection")
			}
		}
	}()
	return nil
}
