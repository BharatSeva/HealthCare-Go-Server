package main

import (
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
	db "vaibhavyadav-dev/healthcareServer/databases"
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
	rabbitMqURL := os.Getenv("RABBITMQ")
	psqlInfo := os.Getenv("POSTGRES")
	mongoURI := os.Getenv("MONGOURL")

	store, err := db.NewCombinedStore(rabbitMqURL, psqlInfo, mongoURI, "db", []string{"golang1", "golang2", "golang3", "golang4"})
	if err != nil {
		log.Fatal("Failed to initialize store:", err)
	}
	PORT := os.Getenv("PORT")
	server := NewAPIServer(PORT, store)

	server.Run()
}
