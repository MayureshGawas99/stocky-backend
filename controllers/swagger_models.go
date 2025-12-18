package controllers

type ErrorResponse struct {
	Error  string `json:"error" example:"invalid user id"`
	Status string `json:"status,omitempty" example:"failure"`
}

type GenericSuccessResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Reward and ledger entries created successfully"`
}

type RewardRequest struct {
	UserID      int64   `json:"user_id" example:"1"`
	StockSymbol string  `json:"stock_symbol" example:"AAPL"`
	Shares      float64 `json:"shares" example:"10"`
	RewardID    string  `json:"reward_id" example:"reward-uuid-123"`
	Timestamp   string  `json:"timestamp" example:"2024-12-18T10:00:00Z"`
}

type TodayStocksResponse struct {
	Date    string      `json:"date" example:"2024-12-18"`
	Rewards interface{} `json:"rewards"`
}

type HistoricalINRResponse struct {
	UserID  int64       `json:"user_id" example:"1"`
	History interface{} `json:"history"`
}

type UserStatsResponse struct {
	UserID  int64       `json:"user_id" example:"1"`
	History interface{} `json:"history"`
}

type PortfolioResponse struct {
	UserID  int64       `json:"user_id" example:"1"`
	History interface{} `json:"history"`
}

type RegisterRequest struct {
	Name     string `json:"name" example:"Mayuresh"`
	Email    string `json:"email" example:"mayuresh@gmail.com"`
	Password string `json:"password" example:"password123"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"mayuresh@gmail.com"`
	Password string `json:"password" example:"password123"`
}

type AuthResponse struct {
	Token string `json:"token"`
	ID    int64  `json:"id"`
	Name  string `json:"name,omitempty"`
}
