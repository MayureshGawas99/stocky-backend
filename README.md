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
- [Swagger UI](#swagger-ui) 🕸️

---

## Quick Start

1. Install Go (>= 1.18) and PostgreSQL.
2. Create a PostgreSQL database (example: `assignment`).
3. Create a `.env` file (example shown below).
4. Run the server: `go run main.go` or build with `go build -o stock-reward-api .` and run the binary.

## Environment (.env)

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

## Run & Development

- Install deps and run:
  - `go mod download` # download module dependencies
  - `go run main.go`
  - Open `http://localhost:8080/swagger/index.html` for API docs.
- To run in production build mode:
  - `go build -o stock-reward-api .`
  - `./stock-reward-api` (or `stock-reward-api.exe` on Windows)

---

## Postman Collection

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

## API Specification

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

## Database Schema

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

