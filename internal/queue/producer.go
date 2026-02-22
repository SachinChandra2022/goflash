package queue

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type OrderMessage struct {
	UserID         int    `json:"user_id"`
	ProductID      int    `json:"product_id"`
	IdempotencyKey string `json:"idempotency_key"`
}

var Writer *kafka.Writer

func ConnectProducer() {
	Writer = &kafka.Writer{
		Addr:         kafka.TCP("localhost:9092"),
		Topic:        "orders",
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,                   
		BatchTimeout: 10 * time.Millisecond,
		Async:        true,
	}
	log.Println("Kafka Producer connected!")
}

func PushOrder(productID, userID int) error {
	orderID := uuid.New().String()

	msg := OrderMessage{
		UserID:         userID,
		ProductID:      productID,
		IdempotencyKey: orderID,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	kafkaMsg := kafka.Message{
		Key:   []byte(string(rune(userID))),
		Value: payload,
	}

	err = Writer.WriteMessages(context.Background(), kafkaMsg)
	if err != nil {
		log.Printf("Failed to write to kafka: %v", err)
		return err
	}
	return nil
}