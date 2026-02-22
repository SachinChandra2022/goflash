# GoFlash: High-Concurrency Flash Sale Engine ⚡

A backend system designed to handle **100,000+ concurrent requests** for a limited inventory flash sale (e.g., selling 100 iPhones to 10k users) without overselling or crashing.

## 🚀 The Challenge
In a typical e-commerce system, a naive "Check-Then-Act" logic leads to **Race Conditions** under load.
- **Problem:** Two users read `stock=1` at the same time. Both buy. Stock becomes `-1`.
- **Goal:** Ensure exactly 100 items are sold, no more, no less, while maintaining low latency.

## 🏗️ Architecture
The system evolved through 5 phases of engineering:

1.  **Phase 1 (Naive):** REST API + Postgres. (Failed: Oversold by 47%).
2.  **Phase 2 (Pessimistic Locking):** Used `SELECT ... FOR UPDATE`. (Safe but Slow).
3.  **Phase 3 (Redis + Lua):** moved locks to memory. (Fast, Atomic, but risky data loss).
4.  **Phase 4 (Async/Event-Driven):** Decoupled Write path using **Kafka**.
    - API -> Redis (Lua) -> Kafka -> Worker -> Postgres.
    - **Result:** Latency < 10ms, Eventual Consistency.
5.  **Phase 5 (Reliability):** Added Idempotency keys to handle duplicate Kafka messages and Graceful Shutdowns.

## 🛠️ Tech Stack
- **Language:** Golang (Gin Framework)
- **Database:** PostgreSQL (Persistence)
- **Caching/Locking:** Redis + Lua Scripts (Atomicity)
- **Message Broker:** Apache Kafka (Asynchronous Processing)
- **Load Testing:** k6 (Simulating 200-500 VUs)
- **Infrastructure:** Docker & Docker Compose

## ⚡ How to Run

### Prerequisites
- Docker & Docker Compose
- Go 1.21+

### 1. Start Infrastructure
```bash
docker-compose up -d
```

### 2. Run the Worker (Consumer)
```bash
go run cmd/worker/main.go
```

### Run the API (Producer)
```bash
go run cmd/api/main.go
```

### 4. Reset Inventory
```bash
curl -X POST http://localhost:8080/reset
```

### 5. Simulate Load (k6)
```bash
k6 run scripts/attack.js
```

## 📈 Results
- Concurrency: 200 Virtual Users
- Requests: 10,000+
- Overselling: 0
- Avg Latency: ~8ms

Built for learning High-Concurrency Systems.