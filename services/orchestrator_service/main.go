package main

import (
	"fmt"
	"log"
	"orchestrator_service/kafka/kafka_consumer"
)

func main() {
	fmt.Println("Starting Orchestrator Service...")

	err := kafka_consumer.StartListening()
	if err != nil {
		log.Fatal("Failed to start Kafka consumer:", err)
	}

	select {} // Keep the service running
}
