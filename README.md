# DelayedNotifier

**DelayedNotifier** is a backend service for scheduling and sending delayed notifications via queues (RabbitMQ).  
It allows you to create notifications that should be delivered at a specific time via multiple channels (Email, Telegram).

---

## Features

- **HTTP API** for creating, cancelling, and checking notifications
- **Background workers** consume messages from RabbitMQ and send notifications at the right time
- **Retry mechanism** with exponential backoff in case of delivery failures
- **Channels supported:** Email, Telegram
- **Simple frontend** (port **3000**) to test the service via a UI

---

## Project Structure

```bash
.
│   ├── cmd/                 # Application entry points (HTTP server, worker, etc.)
│   ├── config/              # Configuration files
│   ├── internal/            # Internal application packages
│   │   ├── api/             # HTTP handlers, routers, server
│   │   ├── config/          # Config parsing logic
│   │   ├── middlewares/     # HTTP middlewares
│   │   ├── mocks/           # Generated mocks for testing
│   │   ├── models/          # Data models
│   │   ├── rabbitmq/        # RabbitMQ connection and consumers
│   │   ├── repository/      # Database repositories (PostgreSQL, Redis)
│   │   ├── service/         # Business logic
│   │   ├── sender/          # Sender logic
│   │   └── worker/          # Background workers for scheduled delivery
│   ├── migrations/          # Database migrations
│   ├── go.mod
│   └── go.sum
│   └── frontend
├── .env.example             # Example environment variables
├── docker-compose.yml       # Multi-service Docker setup
├── Makefile                 # Development commands
└── README.md
````

---

## Makefile Commands

```make
# Run all backend tests with verbose output
make test

# Run linters (vet + golangci-lint)
make lint

# Build and start all Docker services
make up

# Stop and remove all Docker services and volumes
make down
```

---

## Configuration (`.env`)

Before running the project, copy `.env.example` to `.env` and set your own values:

```bash
cp .env.example .env
```

#### 🔑 Notes:

* **SMTP credentials**: Create an account, for example, on [Mailtrap](https://mailtrap.io/) and copy the SMTP login + password into `.env`.
* **Telegram Chat ID**: Open Telegram, start your bot, then go to `https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates` and find your `chat.id`.

---

## Running the Project

1. Copy and update `.env`:

   ```bash
   cp .env.example .env
   ```

2. Build and run services via Docker:

   ```bash
   make up
   ```

3. The backend will be available at:

    * **Backend API** → `http://localhost:8080/api/notify`
    * **Frontend UI** → `http://localhost:3000`

4. To stop services:

   ```bash
   make down
   ```

---

## API Endpoints

All endpoints are available under `/api/notify`:

| Method | Endpoint | Description                  |
| ------ | -------- | ---------------------------- |
| POST   | `/`      | Create a new notification    |
| GET    | `/:id`   | Get status of a notification |
| DELETE | `/:id`   | Cancel a notification        |

---

## Example Requests

### 1. Create a Notification

**POST** `http://localhost:8080/api/notify/`

Request body:

```json
{
    "message": "finish this notifier",
    "date_time": "2025-10-20T16:47:00Z",
    "mail": "example@mail.ru",
    "tg_id": "6176317974"
}
```

Response:

```json
{
  "result": "1"
}
```

---

### 2. Get Notification Status

**GET** `http://localhost:8080/api/notify/1`

Response:

```json
{
  "status": "in process"
}
```
---

### 3. Cancel a Notification

**DELETE** `http://localhost:8080/api/notify/1`

Response:

```json
{
  "result": "notification deleted"
}
```

---

## Frontend

A simple UI is available at **[http://localhost:3000](http://localhost:3000)**.
It provides:

* A form to create a notification
* A table with all notifications and their statuses
* Buttons to cancel a notification

---

## Summary

* **Backend** (Go + RabbitMQ + PostgreSQL) → runs on **port 8080**
* **Frontend** → runs on **port 3000**
* Notifications can be created via **API or UI**
* Notifications are delivered via **Email (SMTP)** and **Telegram Bot**
* Failed deliveries are retried automatically

```