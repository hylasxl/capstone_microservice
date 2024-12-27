package main

import (
	"gateway/routes"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Configure router
	router := mux.NewRouter()

	// Set timezone
	if err := os.Setenv("TZ", "Asia/Bangkok"); err != nil {
		log.Fatalf("Failed to set timezone: %v", err)
	}
	time.Local = time.FixedZone("UTC+7", 7*3600)

	// Initialize gRPC service clients
	clients, err := routes.InitializeServiceClients()
	if err != nil {
		log.Fatalf("Failed to initialize service clients: %v", err)
	}
	defer clients.CloseConnections()

	// Initialize routes
	routes.InitializeRoutes(router, clients)

	// Determine port from environment or use default
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8081"
	}

	// Start the server
	log.Printf("Starting Gateway on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
