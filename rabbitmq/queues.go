package rabbitmq

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

// With this consumer will also collect logs and push it into separate collection
func (c *Rabbitmq) Push_SendNotification(category, name, email, healthcareId interface{}) error {
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
			"name":         name,
			"category":     category,
			"email":        email,
			"healthcareId": healthcareId,
		}
	case "account_login":
		body = map[string]interface{}{
			"name":         name,
			"category":     category,
			"email":        email,
			"healthcareId": healthcareId,
		}
	case "patient_record_created":
		body = map[string]interface{}{
			"name":         name,
			"category":     category,
			"email":        email,
			"healthcareId": healthcareId,
		}
	case "patient_record_viewed":
		body = map[string]interface{}{
			"name":         name,
			"category":     category,
			"email":        email,
			"healthcareId": healthcareId,
		}
	case "appointment_confirm":
		body = map[string]interface{}{
			"name":         name,
			"category":     category,
			"email":        email,
			"healthcareId": healthcareId,
		}
	case "patient_biodata_created":
		body = map[string]interface{}{
			"name":          name,
			"category":      category,
			"email":         email,
			"healthcare_id": healthcareId,
		}
	case "patient_biodata_viewed":
		body = map[string]interface{}{
			"name":         name,
			"category":     category,
			"email":        email,
			"healthcareId": healthcareId,
		}
	case "patient_biodata_updated":
		body = map[string]interface{}{
			"name":         name,
			"category":     category,
			"email":        email,
			"healthcareId": healthcareId,
		}
	case "delete_account":
		body = map[string]interface{}{
			"name":         name,
			"category":     category,
			"email":        email,
			"healthcareId": healthcareId,
		}
	default:
		body = map[string]interface{}{
			"name":         "Vaibhav Yadav",
			"category":     "missed",
			"email":        "tron21vaibhav@gmail",
			"healthcareId": "2021071042",
		}
	}

	bodyjson, err := json.Marshal(body)
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
			Body:        bodyjson,
		})
	if err != nil {
		return err
	}

	log.Printf("[x] Sent %s", bodyjson)
	return nil
}

// Depreciated as of now (will be removed soon)
func (c *Rabbitmq) Push_patientbiodata(biodata map[string]interface{}) error {
	notification_queue, err := c.ch.QueueDeclare(
		"patientbiodata", // queue name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return err
	}
	bodyjson, err := json.Marshal(biodata)
	if err != nil {
		return err
	}

	err = c.ch.Publish(
		"",                      // exchange
		notification_queue.Name, // routing key
		true,                    // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bodyjson,
		})
	if err != nil {
		return err
	}

	log.Printf(" [x] Sent %s", bodyjson)
	return nil
}

// patient records goes here...
func (c *Rabbitmq) Push_patient_records(record map[string]interface{}) error {
	notification_queue, err := c.ch.QueueDeclare(
		"patient_records", // queue name
		false,             // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return err
	}

	bodyjson, err := json.Marshal(record)
	if err != nil {
		return err
	}

	err = c.ch.Publish(
		"",                      // exchange
		notification_queue.Name, // routing key
		true,                    // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bodyjson,
		})
	if err != nil {
		return err
	}

	log.Printf(" [x] Sent %s", bodyjson)
	return nil
}

func (c *Rabbitmq) Push_appointment(category string) error {
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
