# Stock-Reward API ✅

Backend service for assigning stock rewards, maintaining ledger entries, and providing INR valuation and portfolio endpoints.

---

## Table of Contents

- [Quick Start](#quick-start) ⚡
- [Environment (.env)](#environment-env) 🔧
- [Run & Development](#run--development) 🧪
- [Postman Collection](#postman-collection) 📪
- [API Specification](#api-specification) 📋
- [Database Schema](#database-schema) 🗄️
- [Edge Cases & Scaling](#edge-cases--scaling) 🛡️
- [Swagger UI](#swagger-ui) 🕸️

---

## Quick Start

1. Install Go (>= 1.18) and PostgreSQL.
2. Create a PostgreSQL database (example: `assignment`).
3. Create a `.env` file (example shown below).
4. Run the server: `go run main.go` or build with `go build -o stock-reward-api .` and run the binary.

## Environment (.env) 🔧

Create a `.env` file in the project root. Example:

```env
# Database connection (required)
DATABASE_URL=postgres://postgres:password@localhost:5432/assignment

# Optional tuning
DB_MAX_CONNS=10

# JWT secret for auth (required for production)
JWT_SECRET=replace-with-a-secure-random-secret

```

Notes:

- The app will create required tables on startup if they do not exist.
- `LoadDummyUsers()` and `LoadDummyStocks()` are run by `main.go` to seed example data from `resources/` SQL files.

---

## Run & Development 🧪

- Install deps and run:
  - `go run main.go`
  - Open `http://localhost:8080/swagger/index.html` for API docs.
- To run in production build mode:
  - `go build -o stock-reward-api .`
  - `./stock-reward-api` (or `stock-reward-api.exe` on Windows)
- To import dummy SQL manually (optional):
  - `psql "${DATABASE_URL}" -f resources/dummy_users.sql`
  - `psql "${DATABASE_URL}" -f resources/dummy_stocks.sql`

---

## Postman Collection 📪

A shared Postman collection is available for testing all APIs.

🔗 Postman Collection (Cloud Link)

https://www.postman.com/mayureshgawas99-6782218/stocky-backend/collection/50904887-19fa1adf-9070-4ac4-8328-bfbb48042e06?action=share&creator=50904887&active-environment=50904887-58eea175-5870-466f-b475-038f96396a97

⚠️ Secrets are included in the shared collection.

### How to Use

1. Open the link above.
2. Click Fork or Run in Postman.
3. Create or select an environment.

Add the following variables:

| Variable    | Description  | Example                 |
| ----------- | ------------ | ----------------------- |
| `base_url`  | API base URL | `http://localhost:8080` |
| `jwt_token` | JWT token    | `Bearer <token>`        |

Select the environment before sending requests.

Authorization Header

All secured endpoints use:

`Authorization: {{jwt_token}}`

---

## API Specification 📋

Base path: `/api`

Authentication: JWT Bearer token required for `/api/stocks/*` endpoints. Obtain token by registering or logging in at `/api/user`.

### Auth

POST /api/user/register

- Request body:

```json
{
  "name": "John",
  "email": "john@example.com",
  "password": "secret123"
}
```

- Response (201):

```json
{
  "id": 1,
  "name": "John",
  "email": "john@example.com",
  "created_at": "2024-12-18T..Z",
  "token": "<jwt>"
}
```

POST /api/user/login

- Request body:

```json
{
  "email": "john@example.com",
  "password": "secret123"
}
```

- Response (200):

```json
{ "token": "<jwt>", "id": 1, "name": "John" }
```

### Stocks (require Authorization: `Bearer <token>`)

POST /api/stocks/reward

- Request body (`RewardRequest`):

```json
{
  "user_id": 1,
  "stock_symbol": "AAPL",
  "shares": 10.5,
  "reward_id": "reward-uuid-123",
  "timestamp": "2024-12-18T10:00:00Z"
}
```

- Success (200):

```json
{
  "status": "success",
  "message": "Reward and ledger entries created successfully"
}
```

- Failure example (409): duplicate reward or constraint failure.

GET /api/stocks/today-stocks/:userId

- Response (200):

```json
{
  "date": "2024-12-18",
  "rewards": [
    /* Reward events */
  ]
}
```

GET /api/stocks/historical-inr/:userId

- Response (200):

```json
{ "user_id": 1, "history": { "2024-12-18": 12345.67 } }
```

GET /api/stocks/stats/:userId

- Response (200):

```json
{ "user_id": 1, "history": { "AAPL": 3400.5 } }
```

GET /api/stocks/portfolio/:userId

- Response (200):

```json
{
  "user_id": 1,
  "history": {
    "AAPL": {
      "shares": 10.5,
      "stock_price": 340.05,
      "total_value_inr": 3570.525
    }
  }
}
```

(Models are found in `controllers/swagger_models.go`.)

---

## Database Schema 🗄️

Current tables (created at startup):

- users

  - id SERIAL PRIMARY KEY
  - created_at timestamptz
  - name text
  - email text
  - password text (hashed)

- rewards

  - id uuid PRIMARY KEY
  - user_id bigint
  - stock_symbol text
  - shares double precision
  - timestamp timestamptz
  - reward_id text
  - created_at timestamptz

- ledger_entries

  - id uuid PRIMARY KEY
  - user_id bigint
  - entry_type text (e.g., STOCK, CASH, FEE)
  - stock_symbol text (nullable)
  - quantity double precision (shares)
  - amount_inr double precision
  - direction text (DEBIT/CREDIT)
  - reference_id uuid (references rewards.id semantically)
  - created_at timestamptz

- stocks
  - stock_symbol text PRIMARY KEY
  - price double precision
  - updated_at timestamptz

Relationships (logical):

- `rewards.user_id` -> `users.id` (1:N)
- `ledger_entries.user_id` -> `users.id` (1:N)
- `ledger_entries.reference_id` -> `rewards.id` (reversals/associations)
- `ledger_entries.stock_symbol` -> `stocks.stock_symbol` (denormalized price lookup)

Recommendation: add FK/unique constraints in production:

- UNIQUE(reward_id) on `rewards` to guarantee idempotency
- Foreign key constraints for `user_id` and `reference_id` if you want DB-enforced integrity
- Use `numeric(20,6)` or integer smallest-unit (paise) for money to avoid floating point rounding

---

## Edge Cases & Scaling 🛡️

This section explains how the system currently behaves and recommended improvements.

1. Duplicate reward events / replay attacks

- Current: `CreateReward` uses a DB insert inside a transaction and will return an error if the insert fails. The repo assumes duplicates will surface as insert errors.
- Recommendation: add a `UNIQUE(reward_id)` constraint on `rewards.reward_id` and return HTTP 409 for duplicate reward IDs. Implement idempotent behaviour by returning the existing reward if the same `reward_id` is resubmitted.
- Additional: validate and sign reward source (HMAC) or use a nonce + short TTL for public APIs to prevent replay.

2. Stock splits, mergers, or delisting

- Approach:
  - Track corporate actions (table `corporate_actions`): { symbol, action_type, factor, effective_date, note }
  - On split/merge: apply an adjustment to historical `ledger_entries` or maintain an `adjusted_shares` view that retroactively applies factors. Prefer storing original shares and a factor to compute normalized holdings.
  - On delisting: mark stock as `delisted=true`, stop price updates, and keep historical valuations. Provide admin workflow for tender/compensation.

3. Rounding errors in INR valuation

- Current: uses float64 for amounts which can cause rounding.
- Recommendation: store monetary values as integers in paise (int64) or use PostgreSQL `numeric(20,6)` and use fixed-scale decimal arithmetic in application (github.com/shopspring/decimal or built-in big.Rat) and consistent rounding (banker’s rounding or financial rounding rules).

4. Price API downtime or stale data

- Current: app simulates prices via `StartStockPriceUpdater()` and stores last known price in `stocks` table.
- Recommendations for external price APIs:
  - Cache last successful price and include `updated_at` and a TTL; if price is older than TTL, mark it as stale and return a `425`/warning or fallback to last price with a `stale` flag.
  - Rate-limit and back-off on the price API; implement retry + circuit breaker pattern, and send alerts if stale for > X minutes.
  - Ensure the service has a read-only degraded mode: allow read of portfolios/historical valuations with warnings when price data is stale.

5. Adjustments/refunds of previously given rewards

- Design: use negative ledger entries referencing the original `reference_id`. A refund is recorded as a `REFUND` or `REVERSAL` entry with the same `reference_id` and an opposite `direction`/amount to fully or partially reverse a reward.
- Implement an audit trail and admin endpoints that create reversal entries and optionally a `reconciliation` table to track manual interventions.

Scaling notes

- DB: add read replicas for heavy read workloads (reports, historical queries); ensure connection pool tuning (`DB_MAX_CONNS`).
- Queries: paginate and add time-based partitioning if `rewards`/`ledger_entries` grow large.
- Caching: cache commonly-read aggregated results (user stats, portfolio snapshots) in Redis with short TTLs and invalidate on reward writes.
- Async processing: move heavy or non-critical work (e.g., sending emails, notifications, recomputing aggregates) to background workers (RabbitMQ/Redis queue or serverless jobs).

---

## Observability & Testing 💡

- Logs: logger is initialized in `logger/` — ensure `LOG_LEVEL` is set in env for production.
- Health checks: add `/health` endpoint to detect DB connectivity and price feed health.
- Tests: add unit tests for `repository` and controller logic, and integration tests that run against a disposable PostgreSQL (Docker).

---

## Swagger UI 🕸️

Open `http://localhost:8080/swagger/index.html` after starting the server to explore endpoints and try example payloads.
