package routes

import (
	"pulsepoll/backend/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterWebSocketRoutes(
	router *gin.RouterGroup,
	websocketController *controllers.WebSocketController,
) {
	router.GET(
		"/polls/:id/ws",
		websocketController.HandlePollUpdates,
	)
}
