package repository

import (
	"context"
	"errors"
	"time"

	"stock-reward-api/db"
	"stock-reward-api/logger"
	"stock-reward-api/models"

	"github.com/google/uuid"
)


func CreateReward(
	ctx context.Context,
	userID int64,
	stockSymbol string,
	shares float64,
	rewardID string,
	rewardedAt time.Time,
	pricePerShare float64,
	fee float64,
) error {

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) 

	//check if rewqardID already exists
	var existingID uuid.UUID
	err = tx.QueryRow(ctx, "SELECT id FROM rewards WHERE reward_id=$1", rewardID).Scan(&existingID)
	if err == nil {
		logger.Log.Errorf("duplicate reward_id: %s", rewardID)
		return errors.New("duplicate reward")
	}

	var rewardUUID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO rewards 
		(user_id, stock_symbol, shares, reward_id, timestamp)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, stockSymbol, shares, rewardID, rewardedAt).Scan(&rewardUUID)

	if err != nil {
		logger.Log.Errorf("failed to insert reward_event: %v", err)
		return errors.New("duplicate reward or failed to insert reward_event")
	}

	totalStockCost := shares * pricePerShare

	_, err = tx.Exec(ctx, `
		INSERT INTO ledger_entries
		(user_id, entry_type, stock_symbol, quantity, direction, reference_id, amount_inr,created_at)
		VALUES ($1,'STOCK',$2,$3,'DEBIT',$4,0,$5)
	`, userID, stockSymbol, shares, rewardUUID, rewardedAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO ledger_entries
		(user_id, entry_type, amount_inr, direction, reference_id,created_at)
		VALUES ($1,'CASH',$2,'CREDIT',$3,$4)
	`, userID, totalStockCost, rewardUUID, rewardedAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO ledger_entries
		(user_id, entry_type, amount_inr, direction, reference_id,created_at)
		VALUES ($1,'FEE',$2,'CREDIT',$3,$4)
	`, userID, fee, rewardUUID, rewardedAt)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func GetTodayStocks(ctx context.Context, userID int64) ([]models.RewardEvent, error) {
	rows, err := db.Pool.Query(ctx, "SELECT * FROM rewards WHERE user_id=$1 AND DATE(timestamp) = CURRENT_DATE ", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.RewardEvent
	for rows.Next() {
		var r models.RewardEvent
		if err := rows.Scan(&r.ID, &r.UserID, &r.StockSymbol, &r.Shares, &r.RewardedAt, &r.ReferenceID, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func GetStockPrice(ctx context.Context, stockSymbol string) (float64, error) {
	var price float64
	err := db.Pool.QueryRow(ctx, "SELECT price FROM stocks WHERE stock_symbol=$1", stockSymbol).Scan(&price)
	return price, err
}

func GetHistoricalINR(ctx context.Context, userID int64) (map[time.Time]float64, error) {
	query := `
		SELECT
			DATE(l.created_at) AS reward_date,
			SUM(l.quantity * sp.price) AS total_value_inr
		FROM ledger_entries l
		JOIN stocks sp
		ON l.stock_symbol = sp.stock_symbol
		WHERE l.user_id = $1
			AND l.entry_type = 'STOCK'
			AND l.direction = 'DEBIT'
		GROUP BY reward_date
		ORDER BY reward_date;
	`

	rows, err := db.Pool.Query(ctx, query, userID)
	// return an arr which ha s date and total value in inr
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[time.Time]float64)
	for rows.Next() {
		var rewardDate time.Time
		var totalValueInr float64
		if err := rows.Scan(&rewardDate, &totalValueInr); err != nil {
			return nil, err
		}
		result[rewardDate] = totalValueInr
	}

	return result, nil
}

func GetUserStats(ctx context.Context, userID int64) (map[string]float64, error) {
	query := `
		SELECT
			l.stock_symbol,
			SUM(l.quantity * sp.price) AS total_value_inr
		FROM ledger_entries l
		JOIN stocks sp
		ON l.stock_symbol = sp.stock_symbol
		WHERE l.user_id = $1
			AND l.entry_type = 'STOCK'
			AND l.direction = 'DEBIT'
			AND DATE(l.created_at) = CURRENT_DATE
		GROUP BY l.stock_symbol
		ORDER BY l.stock_symbol;
	`

	rows, err := db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var symbol string
		var totalValueInr float64
		if err := rows.Scan(&symbol, &totalValueInr); err != nil {
			return nil, err
		}
		result[symbol] = totalValueInr
	}

	return result, nil
}

func GetPortfolio(ctx context.Context, userID int64) (map[string]map[string]float64, error) {
	query := `
		SELECT
			l.stock_symbol,
			SUM(l.quantity) AS total_shares,
			MAX(sp.price) ,
			SUM(l.quantity * sp.price) AS total_value_inr
		FROM ledger_entries l
		JOIN stocks sp
		ON l.stock_symbol = sp.stock_symbol
		WHERE l.user_id = $1
			AND l.entry_type = 'STOCK'
			AND l.direction = 'DEBIT'
		GROUP BY l.stock_symbol
		ORDER BY l.stock_symbol;
	`

	rows, err := db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// return map of stock symbol as follows "symbol": {"shares": quantity, "stock_price": stock_price, "total_value_inr": totalValueInr}
	result := make(map[string]map[string]float64)
	for rows.Next() {
		var symbol string
		var totalShares float64
		var stockPrice float64
		var totalValueInr float64
		if err := rows.Scan(&symbol, &totalShares, &stockPrice, &totalValueInr); err != nil {
			return nil, err
		}
		result[symbol] = map[string]float64{
			"shares": totalShares,
			"stock_price": stockPrice,
			"total_value_inr": totalValueInr,
		}	
	}

	return result, nil
}