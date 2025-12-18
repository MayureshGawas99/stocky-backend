package controllers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"stock-reward-api/db"
	"stock-reward-api/repository"

	"stock-reward-api/logger"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
)

// CreateReward godoc
// @Summary Create stock reward
// @Description Assign stock reward to a user (idempotent via reward_id)
// @Tags Stocks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reward body RewardRequest true "Reward payload"
// @Success 200 {object} GenericSuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/stocks/reward [post]
func CreateReward(c *gin.Context) {
	var req RewardRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failure", "error": err.Error()})
		return
	}

	rewardedAt, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failure", "error": "invalid timestamp format"})
		return
	}
	pricePerShare, err := repository.GetStockPrice(c.Request.Context(), req.StockSymbol)
	if err != nil {
		logger.Log.Errorf("failed to get stock price for %s: %v", req.StockSymbol, err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failure", "error": "stock does not exist"})
		return
	} 
	fee := 10 + float64(time.Now().UnixNano()%90)              
	logger.Log.Infof("Calculated pricePerShare: %.2f, fee: %.2f", pricePerShare, fee)

	var existingID int64
	err = db.Pool.QueryRow(c.Request.Context(), "SELECT id FROM users WHERE id=$1", req.UserID).Scan(&existingID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"status": "failure", "error": "user_id does not exist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failure", "error": err.Error()})
		return
	}

	err = repository.CreateReward(
		c.Request.Context(),
		req.UserID,
		req.StockSymbol,
		req.Shares,
		req.RewardID,
		rewardedAt,
		pricePerShare,
		fee,
	)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Reward and ledger entries created successfully",
	})
}

// GetTodayStocks godoc
// @Summary Get today’s rewarded stocks
// @Description Returns stocks rewarded today for a user
// @Tags Stocks
// @Produce json
// @Security BearerAuth
// @Param userId path int true "User ID"
// @Success 200 {object} TodayStocksResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/stocks/today-stocks/{userId} [get]
func GetTodayStocks(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var existingID int64
	err = db.Pool.QueryRow(c.Request.Context(), "SELECT id FROM users WHERE id=$1", userId).Scan(&existingID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id does not exist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rewards, err := repository.GetTodayStocks(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date":    time.Now().Format("2006-01-02"),
		"rewards": rewards,
	})
}

// GetHistoricalINR godoc
// @Summary Get historical INR valuation
// @Description Returns historical INR valuation of user rewards
// @Tags Stocks
// @Produce json
// @Security BearerAuth
// @Param userId path int true "User ID"
// @Success 200 {object} HistoricalINRResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/stocks/historical-inr/{userId} [get]
func GetHistoricalINR(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}	
	
	var existingID int64
	err = db.Pool.QueryRow(c.Request.Context(), "SELECT id FROM users WHERE id=$1", userId).Scan(&existingID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id does not exist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Log.Infof("Fetching historical INR for user %d", userId)

	rewards, err := repository.GetHistoricalINR(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Log.Infof("Historical INR for user %d: %+v", userId, rewards)
	c.JSON(http.StatusOK, gin.H{
		"user_id": userId,
		"history": rewards,
	})
}

// GetUserStats godoc
// @Summary Get user stock stats
// @Description Returns aggregated stock statistics
// @Tags Stocks
// @Produce json
// @Security BearerAuth
// @Param userId path int true "User ID"
// @Success 200 {object} UserStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/stocks/stats/{userId} [get]
func GetUserStats (c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}	
	
	var existingID int64
	err = db.Pool.QueryRow(c.Request.Context(), "SELECT id FROM users WHERE id=$1", userId).Scan(&existingID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id does not exist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Log.Infof("Fetching user stats for user %d", userId)

	rewards, err := repository.GetUserStats(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Log.Infof("User stats for user %d: %+v", userId, rewards)
	c.JSON(http.StatusOK, gin.H{
		"user_id": userId,
		"history": rewards,
	})
}

// GetPortfolio godoc
// @Summary Get user portfolio
// @Description Returns current stock holdings
// @Tags Stocks
// @Produce json
// @Security BearerAuth
// @Param userId path int true "User ID"
// @Success 200 {object} PortfolioResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/stocks/portfolio/{userId} [get]
func GetPortfolio(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}	
	
	var existingID int64
	err = db.Pool.QueryRow(c.Request.Context(), "SELECT id FROM users WHERE id=$1", userId).Scan(&existingID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id does not exist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Log.Infof("Fetching portfolio for user %d", userId)

	rewards, err := repository.GetPortfolio(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Log.Infof("Portfolio for user %d: %+v", userId, rewards)
	c.JSON(http.StatusOK, gin.H{
		"user_id": userId,
		"history": rewards,
	})
}

// RegisterUser godoc
// @Summary Register new user
// @Description Creates a new user and returns JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration payload"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/user/register [post]
func RegisterUser(c *gin.Context) {
	type RegisterRequest struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if db.Pool == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	var existingID int64
	err := db.Pool.QueryRow(c.Request.Context(), "SELECT id FROM users WHERE email=$1", req.Email).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	if err != nil && err != pgx.ErrNoRows {
		logger.Log.Errorf("failed to query user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Errorf("failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var id int64
	var createdAt time.Time
	err = db.Pool.QueryRow(c.Request.Context(), "INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id, created_at", req.Name, req.Email, string(hash)).Scan(&id, &createdAt)
	if err != nil {
		logger.Log.Errorf("failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		logger.Log.Warn("JWT_SECRET not set; using empty secret")
	}
	claims := jwt.MapClaims{"sub": id, "exp": time.Now().Add(24 * time.Hour).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		logger.Log.Errorf("failed to sign token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         id,
		"name":       req.Name,
		"email":      req.Email,
		"created_at": createdAt,
		"token":      signed,
	})
}

// LoginUser godoc
// @Summary Login user
// @Description Authenticates user and returns JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body LoginRequest true "Login payload"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/user/login [post]
func LoginUser(c *gin.Context) {
	type LoginRequest struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if db.Pool == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	var id int64
	var name string
	var pwHash string
	err := db.Pool.QueryRow(c.Request.Context(), "SELECT id, name, password FROM users WHERE email=$1", req.Email).Scan(&id, &name, &pwHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		logger.Log.Errorf("failed to query user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(pwHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		logger.Log.Warn("JWT_SECRET not set; using empty secret")
	}
	claims := jwt.MapClaims{"sub": id, "exp": time.Now().Add(24 * time.Hour).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		logger.Log.Errorf("failed to sign token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": signed, "id": id, "name": name})
}