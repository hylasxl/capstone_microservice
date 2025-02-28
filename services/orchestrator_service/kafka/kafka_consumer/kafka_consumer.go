package kafka_consumer

import (
	"fmt"
	"orchestrator_service/kafka/kafka_producer"

	"github.com/IBM/sarama"
)

const (
	orderCreatedTopic     = "order-events"
	paymentProcessedTopic = "payment-events"
	shipmentTopic         = "shipment-events"
)

func StartListening() error {
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)
	if err != nil {
		return err
	}

	partitionConsumer, err := consumer.ConsumePartition(orderCreatedTopic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}

	go func() {
		for message := range partitionConsumer.Messages() {
			event := string(message.Value)
			fmt.Println("Received event:", event)
			handleSagaEvent(event)
		}
	}()

	fmt.Println("Listening for OrderCreated events...")
	return nil
}

func handleSagaEvent(event string) {
	if event == "Order Created" {
		kafka_producer.PublishEvent(paymentProcessedTopic, "Payment Processed")
	} else if event == "Payment Processed" {
		kafka_producer.PublishEvent(shipmentTopic, "Shipment Started")
	}
}
