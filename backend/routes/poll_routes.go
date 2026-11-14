package routes

import (
	"github.com/gin-gonic/gin"

	"pulsepoll/backend/controllers"
	"pulsepoll/backend/middleware"
)

func RegisterPollRoutes(
	router *gin.RouterGroup,
	pollController *controllers.PollController,
) {
	// Public route: view a poll using its shared link.
	router.GET("/polls/:id", pollController.GetPoll)

	// Protected routes: login required.
	polls := router.Group("/polls")
	polls.Use(middleware.AuthMiddleware())

	polls.POST("", pollController.CreatePoll)
	polls.GET("/mine", pollController.GetMyPolls)
}
