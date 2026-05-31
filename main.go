package main

import (
	"log"
	"os"

	"go-final-project/internal/db"
	"go-final-project/internal/server"
)

func main() {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = ":7540"
	}
	
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == ""{
		dbFile = "scheduler.db"
	}

	password := os.Getenv("TODO_PASSWORD")
	if password == ""{
		password = "12345"
	}

	logger := log.New(os.Stdout, "SERVER: ", log.Ldate|log.Ltime)

	sr, err := db.NewSchedulerRepo(dbFile)
	if err != nil {
		logger.Fatal(err)
	}

	server := server.NewServer(logger, port, sr, password)
	err = server.Server.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}


}
