# Project Goal

The goal of this project is to build a scalable and modular backend system that can track brand mentions from Hacker News stories in near real-time. The system fetches top stories, detects mentions of predefined brands, stores the results efficiently in a database, and exposes APIs to retrieve aggregated statistics and related stories.

# 1. Project Scope

The "Hacker News Brand Mention Tracker" project aims to build a scalable backend system that:

- Fetches top stories from Hacker News (HN).
- Detects mentions of predefined brands/keywords in story titles.
- Stores the detected mentions into a PostgreSQL database.
- Provides APIs to retrieve aggregated brand mention statistics and story lists by brand.
- Communicates asynchronously between services using NATS messaging.
- Foo now will  consider brand mention in story title only
- This project is split into two services: **hn-fetcher** and **hn-processor**.

---

# 2. Functional Requirements

- **Service A (hn-fetcher)**:
    - Fetch top stories  from Hacker News Async.
    - Store stories into PostgreSQL if they don't already exist., skip if exist
    - Publish fetched stories to a NATS subject.
    - User should be able to retrieves mention counts and Return all stories where a brand was mentioned.
    - fetcher should have a retry pattern for fetching the story with limit and retryDelay
- **Service B (hn-processor)**:
    - Subscribe to NATS subject and receive stories.
    - Detect predefined brand mentions in the story titles in Async pattern.
    - Store brand mentions in the database linked to story IDs.
- **APIs (exposed by hn-fetcher)**:
    - `GET /brands/stats`: Return brand mention counts.
    - `GET /brands/:brand/stories`: Return all stories where a brand was mentioned.

---

# 3. Non-Functional Requirements

- **Performance**: Must handle high throughput of stories without overwhelming the database or memory.
- **Scalability**: Both services should scale independently based on load.
- **Resiliency**: Graceful shutdowns on Ctrl+C (SIGTERM), retries, and error handling for external failures.
- **Modularity**: Code structured cleanly with dependency injection and clear boundaries between modules.
- **Logging**: Informative logs for key operations, errors, and warnings (using Zerolog).

Note : I will assume that **scalability** is prioritized over **and Consistency** . Accordingly, all design decisions will be made with the primary goal of building a **highly scalable system**.

---

# 4. Architecture Design

### Components:

- **hn-fetcher Service**:
    - Fetches HN top stories.
    - Stores unique stories into PostgreSQL.
    - Publishes story events to NATS.
- **hn-processor Service**:
    - Subscribes to NATS subject.
    - Detects brand mentions.
    - Stores mentions into PostgreSQL.

### Key Design Patterns Used:

- **Dependency Injection**:
    - Each client (DB, NATS, Rest) is injected into services to improve testability and modularity.
- **Worker Pool Pattern**:
    - Implemented via `pond` library.
    - hn-fetcher: worker pool for fetching multiple stories concurrently.
    - hn-processor: worker pool for processing multiple messages concurrently.
- **NATS Communication**:
    - Async messaging via NATS Pub/Sub.
    - Decouples the fetcher and processor services.

### Sequence Diagram:


![sequence Diagram](https://github.com/user-attachments/assets/fc31a325-abbf-416d-8f25-49860f59b124)

### UML Class Diagram

![Hn_fetcher](https://github.com/user-attachments/assets/80591d07-5c4f-41ef-8cbd-907f5eb0da50)

![HN processor](https://github.com/user-attachments/assets/4a1dd946-12c5-4bfb-9d5a-fa6d693bff5d)

Technologies:

- Go (Golang)
- PostgreSQL
- NATS (nats.go library)
- Fiber (Web Framework)
- Pond (Worker Pool)
- Zerolog (Logging)
- GORM (ORM for PostgreSQL)

---

# 5. Areas to Improve

- **Test Coverage** : Add unit test coverage
- **Better Error Handling**: fix story not found error
- **Pagination**: Add pagination to `/brands/:brand/stories` endpoint
- **Brand List Management**: Load brands dynamically from DB or configuration service instead of statically configured
- **Add mutex in Hn-processor** : since we have workers writing to the Db concurrently we need to have shared mutex to prevent racing condition
- Some modules could be exported as packges and reuse it in both service to avoid code duplication
- **Caching**: Cache popular brand stats to reduce DB load
- Detect Brand mentioned in the story text also

---

# 6. Environment Variables

## hn-fetcher `.envrc`

```bash

export PORT="8080"
export DB_HOST="localhost"
export NATS_URL="nats://localhost:4222"
export DB_URL="postgres://hnuser:hnpass@localhost:5432/hackernews?sslmode=disable"
export MAX_WORKERS=10
export MAX_CAPACITY=100
export REST_TIMEOUT_IN_SEC=5
export NATS_SUBJECT="hnfetcher.topstories"
export HN_BASE_URL="https://hacker-news.firebaseio.com/v0/"

```

---

## hn-processor `.envrc`

```bash

export DB_URL="postgres://hnuser:hnpass@localhost:5432/hackernews?sslmode=disable"
export NATS_URL="nats://localhost:4222"
export PORT=8080
export MaxWorkers=10
export MaxCapacity=100

```

---

# 7. Running the Services

## Step 1: Load environment variables

```bash

source .envrc

```

## Step 2: Run the service

```bash

go run cmd/server/main.go

```

✅ Repeat this for both **hn-fetcher** and **hn-processor**.

---

# 8. Infrastructure Setup

## Running NATS Server (Locally)

Use Docker:

```bash

docker run -d --name nats-server -p 4222:4222 nats

```

✅ NATS server will be accessible at `nats://localhost:4222`.

## Running PostgreSQL (Locally)

Both services depend on PostgreSQL.

If you are using the included `docker-compose.yaml`:

```bash

docker-compose up -d

```

✅ PostgreSQL will run with:

- Database: `hackernews`
- Username: `hnuser`
- Password: `hnpass`
# Example API Requests (cURL)

### 1. Get Brand Mention Statistics

```bash

curl -X GET http://localhost:8080/brands/stats

```

✅ **Returns**: A list of brands with their total number of mentions.

Example response:

```json

[
    {
        "brand": "linux",
        "mentions": 2
    },
    {
        "brand": "slack",
        "mentions": 1
    },
    {
        "brand": "microsoft",
        "mentions": 1
    },
    {
        "brand": "openai",
        "mentions": 1
    }
]

```

---

### 2. Get Stories Where a Specific Brand Was Mentioned (e.g., `linux`)

```bash

curl -X GET http://localhost:8080/brands/linux/stories

```

✅ **Returns**: All Hacker News stories where the brand **Linux** was mentioned.

Example response:

```json

[
    {
        "ID": 103,
        "Title": "Writing a Linux Debugger (2017)",
        "Author": "",
        "URL": "https://blog.tartanllama.xyz/writing-a-linux-debugger-setup/",
        "Time": "2025-04-25T15:39:32+03:00",
        "HnID": 43792941
    },
    {
        "ID": 43,
        "Title": "Daily driving a Linux phone, but why?",
        "Author": "",
        "URL": "https://thefoggiest.dev/2025/04/24/daily-driving-a-linux-phone-but-why",
        "Time": "2025-04-24T08:13:35+02:00",
        "HnID": 43779766
    }
]

```
