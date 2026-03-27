package main

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go/sasl/scram"
    "crypto/tls"
	"log"
	"os/signal"
	"fmt"
	"syscall"
	"time"
	"github.com/sachinchandra/goflash/internal/database"
	"github.com/sachinchandra/goflash/internal/config"
	"github.com/sachinchandra/goflash/internal/queue"
	"github.com/segmentio/kafka-go"
)

func main() {
	database.ConnectDB()
	config.LoadEnv()

	broker := config.GetEnv("KAFKA_BROKER", "localhost:9092")
	fmt.Println("DEBUG: The broker string is currently ->", broker)

	// 3. NUCLEAR OVERRIDE: Force it to localhost for now
	broker = "localhost:9092"
	username := config.GetEnv("KAFKA_USERNAME", "")
	password := config.GetEnv("KAFKA_PASSWORD", "")

	// 1. Setup Dialer
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	if password != "" {
		mechanism, err := scram.Mechanism(scram.SHA256, username, password)
		if err != nil {
			log.Fatal("Failed to configure SASL: ", err)
		}
		dialer.SASLMechanism = mechanism
		dialer.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}


	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker}, // Use the forced variable
		Topic:    "orders",
		GroupID:  "order-group",
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  10 * time.Millisecond,
		// Dialer: dialer, // If you have SASL dialer setup, keep it. Otherwise comment out.
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	log.Println("Worker started. Processing orders...")
	for {
		m, err := r.ReadMessage(ctx)
		
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Shutdown signal received. Exiting loop...")
				break
			}
			log.Printf("could not read message: %v", err)
			continue
		}
		var order queue.OrderMessage
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("Invalid JSON: %v", err)
			continue
		}

		query := `INSERT INTO orders (product_id, user_id, idempotency_key) VALUES ($1, $2, $3) ON CONFLICT (idempotency_key) DO NOTHING`
		
		res, err := database.DB.Exec(query, order.ProductID, order.UserID, order.IdempotencyKey)
		if err != nil {
			log.Printf("Error processing order %s: %v", order.IdempotencyKey, err)
			continue
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			log.Printf("Duplicate Order detected (Idempotent): %s", order.IdempotencyKey)
		} else {
			log.Printf("Processed Order: %s", order.IdempotencyKey)
		}
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
	log.Println("Worker exited gracefully.")
}