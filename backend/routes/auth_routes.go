package routes

import (
	"github.com/gin-gonic/gin"

	"pulsepoll/backend/controllers"
)

func RegisterAuthRoutes(router *gin.RouterGroup, authController *controllers.AuthController) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
	}
}
