// @title Stocky Reward Backend API
// @version 1.0
// @description Backend service for stock reward calculation and INR valuation. All stocks related endpoints are protected and require JWT authentication.

// @contact.name Narayan Gawas
// @contact.email narayan.gawas@spit.ac.in
// @host localhost:8080
// @BasePath /

// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter JWT as: Bearer <token>

package main

import (
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"stock-reward-api/db"
	_ "stock-reward-api/docs"
	"stock-reward-api/logger"
	"stock-reward-api/routes"
	"stock-reward-api/utils"
)

func main() {
    logger.Init()

    db.Connect()
	defer db.Close()
	utils.LoadDummyUsers()
	utils.LoadDummyStocks()

	// Seed RNG for stock price updates
	utils.StartStockPriceUpdater(10 * time.Second)
	
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	routes.RegisterRoutes(r)
	routes.RegisterUserRoutes(r)

	r.Run(":8080")
}