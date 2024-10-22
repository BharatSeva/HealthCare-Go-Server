package main

import (
	"fmt"
	"log"
	"sync"
	db "vaibhavyadav-dev/healthcareServer/databases"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

// THIS IS TESTING SERVER IMPLEMENTATION WILL TAKE TIME
func main() {
	var wg sync.WaitGroup // Create a WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done() // Decrement the counter when the goroutine completes
		// This will help to connect to MongoDB Databases
		mongostore, err := db.ConnectToMongoDB(
			"",
			"test",
			"appointments", // Collection name
		)
		if err != nil {
			log.Fatal(err)
		}
		m := NewMONOGODB_SERVER(":3001", mongostore)
		m.Run()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done() // Decrement the counter when the goroutine completes
		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
		postgres, err := db.NewPostgresStore(psqlInfo)
		if err != nil {
			log.Fatal(err)
		}

		if err := postgres.Init(); err != nil {
			log.Fatal(err)
		}
		server := NewAPIServer(":3000", postgres)
		server.Run()
	}()

	// Wait for all goroutines to finish
	wg.Wait()
}
