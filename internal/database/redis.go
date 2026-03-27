package database

import (
	"context"
	"crypto/tls"
	"github.com/sachinchandra/goflash/internal/config"
	"log"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
	addr := config.GetEnv("REDIS_ADDR", "localhost:6379")
	password := config.GetEnv("REDIS_PASSWORD", "")

	opts := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	}

	// IMPORTANT: Upstash requires TLS (Encryption). 
	// If the address is NOT localhost, we must enable TLS.
	if addr != "localhost:6379" {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	Rdb = redis.NewClient(opts)

	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}
	log.Println("Successfully connected to Redis!")
}