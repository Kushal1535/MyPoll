package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"pulsepoll/backend/models"
)

type PollService struct {
	collection     *mongo.Collection
	voteCollection *mongo.Collection
	redisClient    *redis.Client
}

func NewPollService(
	collection *mongo.Collection,
	voteCollection *mongo.Collection,
	redisClient *redis.Client,
) *PollService {
	return &PollService{
		collection:     collection,
		voteCollection: voteCollection,
		redisClient:    redisClient,
	}
}

func (s *PollService) CreatePoll(
	ctx context.Context,
	userID bson.ObjectID,
	question string,
	description string,
	optionTexts []string,
) (*models.Poll, error) {

	question = strings.TrimSpace(question)
	description = strings.TrimSpace(description)

	if question == "" {
		return nil, errors.New("poll question is required")
	}

	if len(optionTexts) < 2 || len(optionTexts) > 10 {
		return nil, errors.New("a poll must have between 2 and 10 options")
	}

	options := make([]models.PollOption, 0, len(optionTexts))
	seen := make(map[string]bool)

	for _, text := range optionTexts {
		text = strings.TrimSpace(text)

		if text == "" {
			return nil, errors.New("poll options cannot be empty")
		}

		key := strings.ToLower(text)

		if seen[key] {
			return nil, errors.New("duplicate poll options are not allowed")
		}

		seen[key] = true

		options = append(options, models.PollOption{
			ID:    uuid.NewString(),
			Text:  text,
			Votes: 0,
		})
	}

	now := time.Now()

	poll := models.Poll{
		ID:          bson.NewObjectID(),
		CreatedBy:   userID,
		Question:    question,
		Description: description,
		Options:     options,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := s.collection.InsertOne(ctx, poll)
	if err != nil {
		return nil, err
	}

	return &poll, nil
}

// GetPollByID retrieves a public poll and its current vote counts.
func (s *PollService) GetPollByID(
	ctx context.Context,
	pollID string,
) (*models.Poll, int64, error) {

	objectID, err := bson.ObjectIDFromHex(pollID)
	if err != nil {
		return nil, 0, errors.New("invalid poll ID")
	}

	var poll models.Poll

	err = s.collection.FindOne(
		ctx,
		bson.M{"_id": objectID},
	).Decode(&poll)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, 0, errors.New("poll not found")
		}

		return nil, 0, err
	}

	totalVotes, err := s.loadVoteCounts(ctx, &poll)
	if err != nil {
		return nil, 0, err
	}

	return &poll, totalVotes, nil
}

// GetMyPolls retrieves only polls created by the authenticated user.
func (s *PollService) GetMyPolls(
	ctx context.Context,
	userID bson.ObjectID,
) ([]models.Poll, []int64, error) {

	findOptions := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := s.collection.Find(
		ctx,
		bson.M{"createdBy": userID},
		findOptions,
	)

	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	var polls []models.Poll

	if err := cursor.All(ctx, &polls); err != nil {
		return nil, nil, err
	}

	if polls == nil {
		polls = []models.Poll{}
	}

	totalVotesList := make([]int64, len(polls))

	for i := range polls {
		totalVotes, err := s.loadVoteCounts(ctx, &polls[i])
		if err != nil {
			return nil, nil, err
		}

		totalVotesList[i] = totalVotes
	}

	return polls, totalVotesList, nil
}

// loadVoteCounts reads Redis counters and rebuilds missing counters from MongoDB.
func (s *PollService) loadVoteCounts(
	ctx context.Context,
	poll *models.Poll,
) (int64, error) {

	var totalVotes int64

	pollID := poll.ID.Hex()

	for i := range poll.Options {
		counterKey := "poll:" + pollID +
			":option:" + poll.Options[i].ID + ":votes"

		votes, err := s.redisClient.Get(ctx, counterKey).Int64()

		if errors.Is(err, redis.Nil) {
			mongoCount, countErr := s.voteCollection.CountDocuments(
				ctx,
				bson.M{
					"pollId":   poll.ID,
					"optionId": poll.Options[i].ID,
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

			votes, err = s.redisClient.Get(ctx, counterKey).Int64()
		}

		if err != nil {
			return 0, err
		}

		poll.Options[i].Votes = votes
		totalVotes += votes
	}

	return totalVotes, nil
}
