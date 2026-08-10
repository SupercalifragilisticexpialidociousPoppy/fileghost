package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"fileghost_server/pkg/api"
	"fileghost_server/pkg/db"
)

func main() {
	fmt.Println("[    SERVER     ] Initializing server...")

	// Initialize Database from pkg/db
	database, err := db.InitializeDB("server_data.db")
	if err != nil {
		log.Fatalf("[  DB -> SERVER  ] Failed to initialize database: %v", err)
	}
	fmt.Println("[  DB -> SERVER  ] Database confirmation received.")
	defer database.Close()

	// Initialize the API server from pkg/api
	srv := &api.Server{DB: database}
	mux := api.NewRouter(srv)

	port := ":2050"
	fmt.Printf("[    SERVER     ] Server listening locally on http://localhost%s\n", port)

	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("[    SERVER     ] Server crashed: %v", err)
	}
}
