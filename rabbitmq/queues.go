package rabbitmq

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

// Important all COUNTERS, LOGS, EMAILS, ANALYTICS will be collected from here!!
func (c *Rabbitmq) Push_logs(category, name, email, healthId, healthcarename, healthcare_id interface{}) error {
	notificationQueue, err := c.ch.QueueDeclare(
		"logs", // queue name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return err
	}

	var body interface{}
	switch category {
	case "hip_accountCreated":
		body = map[string]interface{}{
			"hip_name":        name,
			"date":            time.Now().Format("2006-01-02 15:04:05"),
			"category":        category,
			"hip_email":       email,
			"hip_ipaddress":   healthId,
			"healthcare_id":   healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "hip_accountLogin":
		body = map[string]interface{}{
			"date":            time.Now().Format("2006-01-02 15:04:05"),
			"hip_name":        name,
			"category":        category,
			"hip_ipaddress":   healthId,
			"hip_email":       email,
			"healthcare_id":   healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "records_created":
		// since name and email is not present
		// comment them out
		body = map[string]interface{}{
			"date": time.Now().Format("2006-01-02 15:04:05"),
			// "name":         name,
			// "email":        email,
			"category":        category,
			"health_id":       healthId,
			"healthcare_id":   healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "records_viewed":
		// since name and email is not present
		// comment them out
		body = map[string]interface{}{
			// "patient_name":  name,
			// "patient_email": email,
			"date":            time.Now().Format("2006-01-02 15:04:05"),
			"category":        category,
			"health_id":       healthId,
			"healthcare_id":   healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "appointmentUpdate":
		body = map[string]interface{}{
			"name":            name,
			"date":            time.Now().Format("2006-01-02 15:04:05"),
			"category":        category,
			"email":           email,
			"health_id":       healthId,
			"healthcare_id":   healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "profile_created":
		body = map[string]interface{}{
			"patient_name":    name,
			"date":            time.Now().Format("2006-01-02 15:04:05"),
			"patient_email":   email,
			"category":        category,
			"health_id":       healthId,
			"healthcare_id":   healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "profile_viewed":
		body = map[string]interface{}{
			"patient_name":    name,
			"email":           email,
			"category":        category,
			"date":            time.Now().Format("2006-01-02 15:04:05"),
			"health_id":       healthId,
			"healthcare_id":   healthcare_id,
			"healthcare_name": healthcarename,
		}
	case "profile_updated":
		body = map[string]interface{}{
			"patient_name":    name,
			"category":        category,
			"patient_email":   email,
			"health_id":       healthId,
			"healthcare_id":   healthcare_id,
			"healthcare_name": healthcarename,
			"date":            time.Now().Format("2006-01-02 15:04:05"),
		}
	case "hip_deleteAccount":
		body = map[string]interface{}{
			"hip_name":        name,
			"category":        category,
			"hip_email":       email,
			"healthcare_id":   healthcare_id,
			"healthcare_name": healthcarename,
			"date":            time.Now().Format("2006-01-02 15:04:05"),
		}
	case "hip_request_blocked":
		body = map[string]interface{}{
			"hip_name":        name,
			"category":        category,
			"hip_email":       email,
			"healthcare_id":   healthcare_id,
			"healthcare_name": healthcarename,
			"date":            time.Now().Format("2006-01-02 15:04:05"),
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

	log.Printf(" [x] Sent patient_records_created %s", bodyjson)
	return nil
}

func (c *Rabbitmq) Push_update_appointment(appointment map[string]interface{}) error {
	notification_queue, err := c.ch.QueueDeclare(
		"appointment_update", // queue name
		false,                // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		return err
	}

	bodyjson, err := json.Marshal(appointment)
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

	// log.Printf(" [x] Sent uppate %s", bodyjson)
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
