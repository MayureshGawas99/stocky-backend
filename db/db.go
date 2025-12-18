package db

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"

	"stock-reward-api/logger"
)

var Pool *pgxpool.Pool


func Connect() {
    if err := godotenv.Load(); err != nil {
        logger.Log.Warn("Error loading .env file")
    }

    databaseUrl := os.Getenv("DATABASE_URL")
    if databaseUrl == "" {
        logger.Log.Error("DATABASE_URL is not set")
    }

    cfg, err := pgxpool.ParseConfig(databaseUrl)
    if err != nil {
        logger.Log.Errorf("unable to parse DATABASE_URL: %v", err)
        os.Exit(1)
    }

    if v := os.Getenv("DB_MAX_CONNS"); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            cfg.MaxConns = int32(n)
        }
    }

    Pool, err = pgxpool.ConnectConfig(context.Background(), cfg)
    if err != nil {
        logger.Log.Errorf("Unable to connect to database pool: %v", err)
        os.Exit(1)
    }

    logger.Log.Info("Connected to database (pool)")

    if err := ensureTables(); err != nil {
        logger.Log.Errorf("Failed to ensure tables: %v", err)
        Pool.Close()
        os.Exit(1)
    }
    logger.Log.Info("Database tables ensured")
} 

func ensureTables() error {
    ctx := context.Background()

    if _, err := Pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "pgcrypto";`); err != nil {
        return fmt.Errorf("create extension: %w", err)
    }

    rewards := `CREATE TABLE IF NOT EXISTS rewards (
        id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
        user_id bigint NOT NULL,
        stock_symbol text NOT NULL,
        shares double precision NOT NULL,
        timestamp timestamptz NOT NULL,
        reward_id text,
        created_at timestamptz NOT NULL DEFAULT now()
    );`

    if _, err := Pool.Exec(ctx, rewards); err != nil {
        return fmt.Errorf("create rewards table: %w", err)
    }
    logger.Log.Info("rewards table created")

    ledger := `CREATE TABLE IF NOT EXISTS ledger_entries (
        id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
        user_id bigint NOT NULL,
        entry_type text NOT NULL,
        stock_symbol text,
        quantity double precision,
        amount_inr double precision,
        direction text NOT NULL,
        reference_id uuid,
        created_at timestamptz NOT NULL DEFAULT now()
    );`

    if _, err := Pool.Exec(ctx, ledger); err != nil {
        return fmt.Errorf("create ledger_entries table: %w", err)
    }
    logger.Log.Info("ledger_entries table created")

    users := `CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        created_at timestamptz NOT NULL DEFAULT now(),
        name text NOT NULL,
        email text NOT NULL,
        password text NOT NULL
    );`

    if _, err := Pool.Exec(ctx, users); err != nil {
        return fmt.Errorf("create users table: %w", err)
    }
    logger.Log.Info("users table created")

    stocks := `CREATE TABLE IF NOT EXISTS stocks (
        stock_symbol text PRIMARY KEY,
        price double precision NOT NULL,
        updated_at timestamptz NOT NULL DEFAULT now()
    );`

    if _, err := Pool.Exec(ctx, stocks); err != nil {
        return fmt.Errorf("create stocks table: %w", err)
    }
    logger.Log.Info("stocks table created")

    return nil
}

func Close() {
    if Pool != nil {
        Pool.Close()
        logger.Log.Info("Database pool closed")
    }
}

