package routes

import (
	"github.com/gin-gonic/gin"

	"stock-reward-api/controllers"
	"stock-reward-api/middleware"
)

func RegisterRoutes(router *gin.Engine) {

	api := router.Group("/api/stocks")
	{
		api.Use(middleware.AuthMiddleware())

		api.POST("/reward", controllers.CreateReward)

		api.GET("/today-stocks/:userId", controllers.GetTodayStocks)

		api.GET("/historical-inr/:userId", controllers.GetHistoricalINR)

		api.GET("/stats/:userId", controllers.GetUserStats)

		api.GET("/portfolio/:userId", controllers.GetPortfolio) 
	}
}

func RegisterUserRoutes(router *gin.Engine) {
	userRoutes := router.Group("/api/user")
	{
		userRoutes.POST("/register", controllers.RegisterUser)
		userRoutes.POST("/login", controllers.LoginUser)
	}
}