package main

import "log"

// THIS IS TESTING SERVER IMPLEMENTATION WILL TAKE TIME

func main() {

	postgres, err := NewPostgresStore("user=postgres dbname=postgres password=user sslmode=disable ")
	if err != nil {
		log.Fatal(err)
	}

	if err := postgres.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(":3000", postgres)
	server.Run()
}
