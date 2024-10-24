package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
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

func (c *Rabbitmq) Notification(category, name, email, healthcareId string) error {
	notificationQueue, err := c.ch.QueueDeclare(
		"notification", // queue name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return err
	}

	var body interface{}
	switch category {
	case "account_created":
		body = map[string]interface{}{
			"name":     name,
			"category": category,
			"email":    email,
			"healthId": healthcareId,
		}
	case "account_login":
		body = map[string]interface{}{
			"name":     name,
			"category": category,
			"email":    email,
			"healthId": healthcareId,
		}
	case "patient_record_created":
		body = map[string]interface{}{
			"name":     name,
			"category": category,
			"email":    email,
			"healthId": healthcareId,
		}
	case "patient_record_viewed":
		body = map[string]interface{}{
			"name":     name,
			"category": category,
			"email":    email,
			"healthId": healthcareId,
		}
	case "appointment_confirm":
		body = map[string]interface{}{
			"name":     name,
			"category": category,
			"email":    email,
			"healthId": healthcareId,
		}
	case "patient_biodata_created":
		body = map[string]interface{}{
			"name":           name,
			"category":       category,
			"email":          email,
			"healthcare_id": healthcareId,
		}
	case "patient_biodata_viewed":
		body = map[string]interface{}{
			"name":     name,
			"category": category,
			"email":    email,
			"healthId": healthcareId,
		}
	case "delete_account":
		body = map[string]interface{}{
			"name":     name,
			"category": category,
			"email":    email,
			"healthId": healthcareId,
		}
	default:
		body = map[string]interface{}{
			"name":     "Vaibhav Yadav",
			"category": "missed",
			"email":    "tron21vaibhav@gmail",
			"healthId": "2021071042",
		}
	}

	// Convert body to JSON format
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Publish message to queue
	err = c.ch.Publish(
		"",                     // exchange
		notificationQueue.Name, // routing key
		true,                   // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bodyBytes,
		})
	if err != nil {
		return err
	}

	log.Printf("[x] Sent %s", string(bodyBytes))
	return nil
}

func (c *Rabbitmq) Appointment(category string) error {
	notification_queue, err := c.ch.QueueDeclare(
		category, // queue name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a queue")
	body := "This is notification"
	err = c.ch.Publish(
		"",                      // exchange
		notification_queue.Name, // routing key
		true,                    // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", body)
	return nil
}

func (c *Rabbitmq) Patient_records(category string) error {
	notification_queue, err := c.ch.QueueDeclare(
		category, // queue name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a queue")
	body := "This is notification"
	err = c.ch.Publish(
		"",                      // exchange
		notification_queue.Name, // routing key
		true,                    // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", body)
	return nil
}
