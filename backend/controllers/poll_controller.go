package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"pulsepoll/backend/services"
)

type PollController struct {
	pollService *services.PollService
}

func NewPollController(pollService *services.PollService) *PollController {
	return &PollController{pollService: pollService}
}

type CreatePollRequest struct {
	Question    string   `json:"question" binding:"required"`
	Description string   `json:"description"`
	Options     []string `json:"options" binding:"required,min=2,max=10"`
}

func (pc *PollController) CreatePoll(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	objectID, ok := userID.(bson.ObjectID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user identity"})
		return
	}

	var request CreatePollRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provide a question and between 2 and 10 options",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	poll, err := pc.pollService.CreatePoll(
		ctx,
		objectID,
		request.Question,
		request.Description,
		request.Options,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Poll created successfully",
		"poll":    poll,
	})
}

// GetPoll retrieves a public poll and its current vote counts.
func (pc *PollController) GetPoll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	poll, totalVotes, err := pc.pollService.GetPollByID(
		ctx,
		c.Param("id"),
	)

	if err != nil {
		status := http.StatusInternalServerError

		switch err.Error() {
		case "invalid poll ID":
			status = http.StatusBadRequest
		case "poll not found":
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"poll":       poll,
		"totalVotes": totalVotes,
	})
}

// GetMyPolls retrieves polls belonging to the logged-in user.
func (pc *PollController) GetMyPolls(c *gin.Context) {
	userID, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	objectID, ok := userID.(bson.ObjectID)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user identity"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	polls, totalVotes, err := pc.pollService.GetMyPolls(ctx, objectID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to retrieve your polls",
		})
		return
	}

	type MyPollResponse struct {
		Poll       interface{} `json:"poll"`
		TotalVotes int64       `json:"totalVotes"`
	}

	results := make([]MyPollResponse, 0, len(polls))

	for i, poll := range polls {
		results = append(results, MyPollResponse{
			Poll:       poll,
			TotalVotes: totalVotes[i],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"polls": results,
		"count": len(results),
	})
}
