package routes

import (
	"pulsepoll/backend/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterVoteRoutes(
	router *gin.RouterGroup,
	voteController *controllers.VoteController,
) {
	router.POST("/polls/:id/vote", voteController.CastVote)
}
