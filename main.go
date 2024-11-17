package main

import (
	"log"
	"os"
	"time"
	db "vaibhavyadav-dev/healthcareServer/databases"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Start Rabbit mq server for message queueing :)
	// This will handle all the asynchronous task like notification, patient_records, and
	// appointments

	// One thing that I'm deeply interested and passionate about --> MACHINE LEARNING
	// ONE STEP CLOSER TO IT

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	redisURL := os.Getenv("REDIS")
	rabbitMqURL := os.Getenv("RABBITMQ")
	psqlInfo := os.Getenv("POSTGRES")
	mongoURI := os.Getenv("MONGOURL") 

	// first one is redis url, second one is limit, and third one is time.Second
	// limit -> 10
	// window -> per 5 second
	// 							  redisurl limit timesecond
	store, err := db.Combinedstore(redisURL, 30, 20*time.Second, rabbitMqURL, psqlInfo, mongoURI, "db", []string{"golang1", "golang2", "golang3", "golang4"})
	if err != nil {
		log.Fatal("Failed to initialize store:", err)
	}
	PORT := os.Getenv("PORT")
	server := NewAPIServer(PORT, store)
	server.Run()
}
