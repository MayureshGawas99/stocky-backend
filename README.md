# Stock Reward API

This repository contains a backend service developed as part of a case study assignment.  
The service records stock reward events for users, maintains ledger-style entries, and exposes APIs to view rewards and portfolio valuation in INR.

The implementation prioritizes clarity, correctness, and ease of reasoning over over-engineering.

---

## Problem Understanding

The requirements can be summarized as:

- Record stock reward events for users
- Maintain a ledger of stock and cash movements
- Provide APIs to:
  - Fetch today’s rewards for a user
  - Calculate historical INR value of rewards
  - View per-stock statistics and portfolio value

The system is modeled around immutable reward events and derived ledger entries, which aligns with common financial-system design patterns while keeping the scope appropriate for a take-home assignment.

---

## Tech Stack

- Go (Gin framework)
- PostgreSQL
- JWT-based authentication
- Swagger (OpenAPI) for API documentation

---

## Project Structure

- `main.go` – application entry point
- `controllers/` – HTTP handlers and Swagger models
- `services/` – business logic for rewards, ledger, and valuation
- `repositories/` – database access layer
- `models/` – domain models
- `resources/` – SQL files for seeding initial data
- `middlewares/` – authentication and request handling helpers

---

## Running the Project Locally

### Prerequisites

- Go >= 1.18
- PostgreSQL

### Setup Steps

1. Create a PostgreSQL database (example: `assignment`)
2. Create a `.env` file in the project root with the following contents:

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/assignment
DB_MAX_CONNS=10
JWT_SECRET=replace-with-a-secure-secret
```

3. Install dependencies and start the server:

```bash
go mod download
go run main.go
```

The application starts on `http://localhost:8080`.

Swagger UI is available at:
`http://localhost:8080/swagger/index.html`

---

## Data Initialization

On startup, the application automatically creates required tables if they do not exist.

For ease of testing and local development, the application seeds:

- Dummy users
- Sample stock price data

These are loaded from SQL files located in the `resources/` directory. This keeps the initialization logic simple and easy to inspect.

---

## API Overview

Base path: `/api`

Authentication is handled using JWT tokens.
All stock-related endpoints require the header:

```
Authorization: Bearer <token>
```

### Authentication Endpoints

- `POST /api/user/register`
- `POST /api/user/login`

Authentication is intentionally minimal to keep the focus on reward processing and ledger logic.

---

## Stock Reward APIs

- `POST /api/stocks/reward`
  Records a stock reward event and creates corresponding ledger entries.

- `GET /api/stocks/today-stocks/{userId}`
  Returns all rewards granted to a user for the current day.

- `GET /api/stocks/historical-inr/{userId}`
  Aggregates reward values per day and returns their INR valuation.

- `GET /api/stocks/stats/{userId}`
  Returns per-stock aggregated reward values.

- `GET /api/stocks/portfolio/{userId}`
  Returns the user’s portfolio with total shares and INR valuation per stock.

Request and response models are defined explicitly in `controllers/swagger_models.go`.

---

## Database Design

The database schema is intentionally kept simple and easy to reason about.

### Tables

**users**

- Stores basic user information and hashed passwords

**rewards**

- Represents immutable reward events
- Each reward maps to one or more ledger entries

**ledger_entries**

- Records stock and cash movements
- Serves as the source of truth for all calculations

**stocks**

- Stores latest stock prices used for valuation

---

## Design Notes

- Foreign key constraints are avoided to speed up iteration, but should be added in a production system.
- Floating point values are used for simplicity; monetary values should use numeric types or smallest currency units in real-world applications.
- Idempotency is handled at the application level; a unique constraint on `reward_id` is recommended for stronger guarantees.

---

## Postman Collection

A Postman collection is provided to make testing the APIs easier.

Link:
[https://www.postman.com/mayureshgawas99-6782218/stocky-backend/collection/50904887-19fa1adf-9070-4ac4-8328-bfbb48042e06](https://www.postman.com/mayureshgawas99-6782218/stocky-backend/collection/50904887-19fa1adf-9070-4ac4-8328-bfbb48042e06)

Required environment variables:

- `base_url`
- `jwt_token`

---

## Limitations and Assumptions

- Timezone handling assumes server time
- No pagination (expected data volume is small)
- No background jobs or asynchronous processing
- Single stock price source for valuation

These trade-offs were made to keep the solution focused within the scope of the assignment.

---

## Possible Improvements

- Add proper foreign key constraints and indexes
- Cache stock prices and reduce repeated calculations
- Introduce pagination and filtering for large datasets
- Move valuation logic to background jobs
- Improve error classification, logging, and monitoring

## Swagger UI 🕸️

Generate and view the API docs (Swagger UI):

1. Install the swag CLI (one-time) 🔧

```bash
# Install latest swag CLI
go install github.com/swaggo/swag/cmd/swag@latest
```

2. Generate / Regenerate docs 📄

```bash
# From the project root (where go.mod is)
swag init
# (Optional) explicitly point at entry file
# swag init -g main.go
```

- This creates/updates `docs/docs.go`, `swagger.json`, and `swagger.yaml`.

3. Run the server and open the UI 🔍

```bash
go run main.go
# Then open in your browser:
http://localhost:8080/swagger/index.html
```

4. Useful tips 💡

- Ensure `main.go` imports the generated package: `_ "stock-reward-api/docs"` (already present in this project).
- After adding or updating handler annotations, re-run `swag init` to refresh docs.
- To serve a specific JSON, use: `http://localhost:8080/swagger/doc.json`.
