package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"stock-reward-api/db"
	"stock-reward-api/logger"
)

func ExecuteSQLFile(path string) error {
	if db.Pool == nil {
		return fmt.Errorf("database not initialized")
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read sql file %s: %w", path, err)
	}

	sql := strings.TrimSpace(string(b))
	if sql == "" {
		return fmt.Errorf("sql file %s is empty", path)
	}

	ctx := context.Background()
	if _, err := db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("execute sql file %s: %w", path, err)
	}

	logger.Log.Infof("Executed SQL file %s", path)
	return nil
}

func LoadDummyUsers() error {
	rel := filepath.Join("resources", "dummy_users.sql")

	wd, err := os.Getwd()
	if err == nil {
		p := filepath.Join(wd, rel)
		if _, err := os.Stat(p); err == nil {
			return ExecuteSQLFile(p)
		}
	}

	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		p := filepath.Join(dir, rel)
		if _, err := os.Stat(p); err == nil {
			return ExecuteSQLFile(p)
		}
	}

	return ExecuteSQLFile(rel)
}


func LoadDummyStocks() error {

	rel := filepath.Join("resources", "dummy_stocks.sql")

	wd, err := os.Getwd()
	if err == nil {
		p := filepath.Join(wd, rel)
		if _, err := os.Stat(p); err == nil {
			return ExecuteSQLFile(p)
		}
	}

	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		p := filepath.Join(dir, rel)
		if _, err := os.Stat(p); err == nil {
			return ExecuteSQLFile(p)
		}
	}

	return ExecuteSQLFile(rel)
}

func StartStockPriceUpdater(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			updateStockPrices()
		}
	}()
}

func updateStockPrices() {
	query := `
		UPDATE stocks
		SET
			price = 1500 + random() * 500,
			updated_at = now();
	`

	_, err := db.Pool.Exec(context.Background(), query)
	if err != nil {
		logger.Log.Errorf("Failed to update stock prices: %v", err)
		return
	}

	logger.Log.Println("Stock prices updated successfully")
}