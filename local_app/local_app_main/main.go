package main

import (
	localapp "local_app"
	"log"
)

func main() {
	db, err := localapp.Connect()

	if err != nil {
		db.Close()
		log.Fatal("Failed to connect to the database:", err)
	}
	defer db.Close()

	localapp.SetupRouter(db)

	localapp.StartServer()
}
