package queue

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/sachinchandra/goflash/internal/config"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type OrderMessage struct {
	UserID         int    `json:"user_id"`
	ProductID      int    `json:"product_id"`
	IdempotencyKey string `json:"idempotency_key"`
}

var Writer *kafka.Writer

func ConnectProducer() {
	broker := config.GetEnv("KAFKA_BROKER", "localhost:9092")
	broker = "localhost:9092"
	username := config.GetEnv("KAFKA_USERNAME", "")
	password := config.GetEnv("KAFKA_PASSWORD", "")

	// 1. Setup default dialer for Localhost
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	// 2. If a password exists, we are in the Cloud. Enable TLS and SASL.
	if password != "" {
		mechanism, err := scram.Mechanism(scram.SHA256, username, password)
		if err != nil {
			log.Fatal("Failed to configure SASL: ", err)
		}
		dialer.SASLMechanism = mechanism
		dialer.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	// 3. Configure the Writer
	Writer = &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        "orders",
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 5 * time.Millisecond,
		Async:        true,
		Transport: &kafka.Transport{
			Dial: dialer.DialFunc,
			TLS:  dialer.TLS,
			SASL: dialer.SASLMechanism,
		},
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