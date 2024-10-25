package rabbitmq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type Rabbitmq struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func Connect2rabbitmq(URL string) (*Rabbitmq, error) {
	conn, err := amqp.Dial(URL)
	if err != nil {
		return nil, err
	}
	// defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Ping to check server connection
	if err := ch.ExchangeDeclarePassive("amq.direct", "direct", true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("failed to connect RabbitMQ server: %w", err)
	}

	log.Printf("Successfully Connected to RabbitMq server... :)")
	return &Rabbitmq{
		conn: conn,
		ch:   ch,
	}, nil
}

