package kafka_producer

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

func PublishEvent(topic, message string) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Println("Kafka producer error:", err)
		return err
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	_, _, err = producer.SendMessage(msg)
	if err != nil {
		log.Println("Kafka publish error:", err)
		return err
	}

	fmt.Printf("Published event: %s -> %s\n", topic, message)
	return nil
}
