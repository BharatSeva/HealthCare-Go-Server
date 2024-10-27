package rabbitmq

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

// Important all COUNTERS, LOGS, EMAILS, ANALYTICS will be collected from here!!
func (c *Rabbitmq) Push_logs(category, name, email, healthId, healthcarename, healthcare_id interface{}) error {
	notificationQueue, err := c.ch.QueueDeclare(
		"hip:logs", // queue name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return err
	}

	var body interface{}
	switch category {
	case "hip:account_created":
		body = map[string]interface{}{
			"hip_name":        name,
			"category":        category,
			"hip_email":       email,
			"hip_ipaddress":   healthId,
			"healthcareId":    healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "hip:account_login":
		body = map[string]interface{}{
			"hip_name":      name,
			"category":      category,
			"hip_ipaddress": healthId,
			"hip_email":     email,
			"healthcareId":    healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "hip:patient_record_created":
		body = map[string]interface{}{
			// "name":         name,
			// "email":        email,
			"category":     category,
			"healthId":     healthId,
			"healthcareId":    healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "hip:patient_record_viewed":
		body = map[string]interface{}{
			// "patient_name":  name,
			// "patient_email": email,
			"category":        category,
			"healthId":        healthId,
			"healthcareId":    healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "hip:appointment_confirm":
		body = map[string]interface{}{
			"name":         name,
			"category":     category,
			"email":        email,
			"healthId":     healthId,
			"healthcareId":    healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "hip:patient_biodata_created":
		body = map[string]interface{}{
			"patient_name":  name,
			"patient_email": email,
			"category":      category,
			"healthId":      healthId,
			"healthcareId":    healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "hip:patient_biodata_viewed":
		body = map[string]interface{}{
			"patient_name":    name,
			"email":           email,
			"category":        category,
			"healthId":        healthId,
			"healthcareId":    healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "hip:patient_biodata_updated":
		body = map[string]interface{}{
			"patient_name":    name,
			"category":        category,
			"patient_email":   email,
			"healthId":        healthId,
			"healthcareId":    healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "hip:delete_account":
		body = map[string]interface{}{
			"hip_name":     name,
			"category":     category,
			"hip_email":    email,
			"healthcareId":    healthcare_id,
			"healthcare_name": healthcarename,
		}
	default:
		body = map[string]interface{}{
			"name":         "Vaibhav Yadav",
			"category":     "hip:missed",
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

// patient records goes here...
func (c *Rabbitmq) Push_patient_records(record map[string]interface{}) error {
	notification_queue, err := c.ch.QueueDeclare(
		"hip:patient_records", // queue name
		false,                 // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
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

	log.Printf(" [x] Sent patient_records_created %s", bodyjson)
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

// Depreciated will be removed soon
// With this consumer will also collect logs and push it into separate collection
func (c *Rabbitmq) Push_counters(category, healthcareId string) error {
	notificationQueue, err := c.ch.QueueDeclare(
		"hip:counters", // queue name
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
	case "hip:requestcounter":
		body = map[string]interface{}{
			"healthcareId": healthcareId,
		}
	case "hip:recordsviewed_counter":
		body = map[string]interface{}{
			"healthcareId": healthcareId,
		}
	case "hip:recordscreated_counter":
		body = map[string]interface{}{
			"healthcareId": healthcareId,
		}
	case "hip:patientbiodata_created_counter":
		body = map[string]interface{}{
			"healthcareId": healthcareId,
		}
	case "hip:patientbiodata_viewed_counter":
		body = map[string]interface{}{
			"healthcareId": healthcareId,
		}
	default:
		body = map[string]interface{}{
			"healthcareId": "2021071042",
			"to":           "missed",
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
