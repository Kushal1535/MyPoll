package controllers

import (
	"net/http"

	"pulsepoll/backend/services"

	"github.com/gin-gonic/gin"
)

type VoteController struct {
	voteService *services.VoteService
}

func NewVoteController(voteService *services.VoteService) *VoteController {
	return &VoteController{
		voteService: voteService,
	}
}

type CastVoteRequest struct {
	OptionID string `json:"optionId" binding:"required"`
	VoterID  string `json:"voterId" binding:"required"`
}

func (vc *VoteController) CastVote(c *gin.Context) {
	pollID := c.Param("id")

	var request CastVoteRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "optionId and voterId are required",
		})
		return
	}

	votes, err := vc.voteService.CastVote(
		c.Request.Context(),
		pollID,
		request.OptionID,
		request.VoterID,
	)

	if err != nil {
		status := http.StatusBadRequest

		if err.Error() == "poll not found" {
			status = http.StatusNotFound
		}

		if err.Error() == "you have already voted in this poll" {
			status = http.StatusConflict
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Vote submitted successfully",
		"votes":   votes,
	})
}
