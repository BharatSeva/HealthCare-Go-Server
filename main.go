package main
import (
	"fmt"
	"log"
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
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	postgres, err := NewPostgresStore(psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	if err := postgres.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(":3000", postgres)
	server.Run()
}
