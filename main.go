package main

import (
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log" 
	"os"
	db "vaibhavyadav-dev/healthcareServer/databases"
)
 
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	psqlInfo := os.Getenv("POSTGRES")
	mongoURI := os.Getenv("MONGOURL")

	store, err := db.NewCombinedStore(psqlInfo, mongoURI, "db", []string{"golang1", "golang2", "golang3", "golang4"})
	if err != nil {
		log.Fatal("Failed to initialize store:", err)
	}
	PORT := os.Getenv("PORT")
	server := NewAPIServer(PORT, store)

	server.Run()
}
