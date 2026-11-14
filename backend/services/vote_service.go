package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"pulsepoll/backend/models"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type VoteService struct {
	voteCollection *mongo.Collection
	pollCollection *mongo.Collection
	redisClient    *redis.Client
}

type VoteUpdate struct {
	PollID   string `json:"pollId"`
	OptionID string `json:"optionId"`
	Votes    int64  `json:"votes"`
}

func NewVoteService(
	voteCollection *mongo.Collection,
	pollCollection *mongo.Collection,
	redisClient *redis.Client,
) *VoteService {
	return &VoteService{
		voteCollection: voteCollection,
		pollCollection: pollCollection,
		redisClient:    redisClient,
	}
}

func (s *VoteService) CastVote(
	ctx context.Context,
	pollID string,
	optionID string,
	voterID string,
) (int64, error) {

	objectID, err := bson.ObjectIDFromHex(pollID)
	if err != nil {
		return 0, errors.New("invalid poll ID")
	}

	var poll models.Poll

	err = s.pollCollection.FindOne(
		ctx,
		bson.M{"_id": objectID},
	).Decode(&poll)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, errors.New("poll not found")
		}
		return 0, err
	}

	if !poll.IsActive {
		return 0, errors.New("this poll is closed")
	}

	validOption := false
	for _, option := range poll.Options {
		if option.ID == optionID {
			validOption = true
			break
		}
	}

	if !validOption {
		return 0, errors.New("invalid poll option")
	}

	if voterID == "" {
		return 0, errors.New("voter ID is required")
	}

	counterKey := "poll:" + pollID + ":option:" + optionID + ":votes"

	// Initialize Redis from existing MongoDB votes if necessary.
	_, err = s.redisClient.Get(ctx, counterKey).Result()

	if errors.Is(err, redis.Nil) {
		mongoCount, countErr := s.voteCollection.CountDocuments(
			ctx,
			bson.M{
				"pollId":   objectID,
				"optionId": optionID,
			},
		)
		if countErr != nil {
			return 0, countErr
		}

		_, err = s.redisClient.SetNX(
			ctx,
			counterKey,
			mongoCount,
			0,
		).Result()

		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	vote := models.Vote{
		ID:        bson.NewObjectID(),
		PollID:    objectID,
		OptionID:  optionID,
		VoterID:   voterID,
		CreatedAt: time.Now(),
	}

	// Persist the vote in MongoDB.
	_, err = s.voteCollection.InsertOne(ctx, vote)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return 0, errors.New("you have already voted in this poll")
		}
		return 0, err
	}

	// Increment the Redis counter after a successful MongoDB insert.
	updatedCount, err := s.redisClient.Incr(ctx, counterKey).Result()
	if err != nil {
		return 0, err
	}

	// Publish the updated count to this poll's Redis channel.
	update := VoteUpdate{
		PollID:   pollID,
		OptionID: optionID,
		Votes:    updatedCount,
	}

	message, err := json.Marshal(update)
	if err != nil {
		log.Printf("Failed to encode vote update: %v", err)
		return updatedCount, nil
	}

	channel := "poll:" + pollID + ":updates"

	if err := s.redisClient.Publish(ctx, channel, message).Err(); err != nil {
		log.Printf("Redis event publishing failed: %v", err)
	}

	return updatedCount, nil
}
