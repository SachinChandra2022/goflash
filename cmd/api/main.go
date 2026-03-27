package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sachinchandra/goflash/internal/config"
	"github.com/sachinchandra/goflash/internal/database"
	"github.com/sachinchandra/goflash/internal/queue"
	"github.com/sachinchandra/goflash/internal/repository"
	
)

// --- PROMETHEUS METRICS ---
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// --- MIDDLEWARES ---

func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := http.StatusText(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}

// CORSMiddleware allows the React Frontend to talk to this API
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	// 1. INITIALIZE CONFIG (Must be first)
	config.LoadEnv()

	// 2. CONNECT SERVICES
	database.ConnectDB()
	database.ConnectRedis()
	queue.ConnectProducer()

	// 3. SETUP DATABASE SCHEMA
	repository.IntializeSchema()

	// 4. SETUP GIN
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 5. ATTACH MIDDLEWARES
	r.Use(CORSMiddleware())
	r.Use(prometheusMiddleware())

	// --- ROUTES ---

	// Metrics for Prometheus
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Phase 1: Naive (Broken)
	r.POST("/purchase", func(c *gin.Context) {
		userID := 1
		productID := 1
		err := repository.PurchaseProductNaive(productID, userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Sold Out or Error", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully"})
	})

	// Phase 2: Pessimistic Locking
	r.POST("/purchase-lock", func(c *gin.Context) {
		userID := 1
		productID := 1
		err := repository.PurchaseProductPessimistic(productID, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Sold out!"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "order placed successfully"})
	})

	// Phase 3: Redis Atomic
	r.POST("/purchase-redis", func(c *gin.Context) {
		productID := 1
		err := repository.PurchaseProductRedis(productID)
		if err != nil {
			if err.Error() == "sold out" {
				c.JSON(http.StatusConflict, gin.H{"message": "Sold out!"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully (Redis)"})
	})

	// Phase 4: Async Architecture (Redis + Kafka)
	r.POST("/purchase-async", func(c *gin.Context) {
		productID := 1
		userID := 1
		err := repository.PurchaseProductRedis(productID)
		if err != nil {
			if err.Error() == "sold out" {
				c.JSON(http.StatusConflict, gin.H{"message": "Sold out!"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		err = queue.PushOrder(productID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue order"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Order processing started"})
	})

	// Utility: Reset Inventory
	r.POST("/reset", func(c *gin.Context) {
		database.Rdb.Set(database.Ctx, "product:1:quantity", 100, 0)
		database.DB.Exec("UPDATE products SET quantity = 100 WHERE id = 1")
		database.DB.Exec("DELETE FROM orders")
		c.JSON(http.StatusOK, gin.H{"message": "Inventory reset to 100"})
	})

	// --- NEW: THE REAL-TIME STREAM (SSE) ---
	r.GET("/stream", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		for {
			select {
			case <-c.Request.Context().Done():
				return
			default:
				val, err := database.Rdb.Get(database.Ctx, "product:1:quantity").Result()
				stock := "0"
				if err == nil {
					stock = val
				}
				// Format data for EventSource
				fmt.Fprintf(c.Writer, "data: {\"stock\": \"%s\"}\n\n", stock)
				c.Writer.Flush()
				time.Sleep(500 * time.Millisecond)
			}
		}
	})

	// 6. START SERVER
	port := config.GetEnv("PORT", "8081")
	log.Printf("Starting server on port %s...", port)
	err := r.Run(":" + port)
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}