package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"pulsepoll/backend/controllers"
	"pulsepoll/backend/middleware"
	"pulsepoll/backend/routes"
	"pulsepoll/backend/services"
)

var mongoClient *mongo.Client
var redisClient *redis.Client

// Connect to MongoDB
func connectMongoDB() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}

	mongoClient = client
	log.Println("MongoDB connected successfully!")
}

// Connect to Redis
func connectRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	log.Println("Redis connected successfully!")
}

func main() {
	// Initialize database connections
	connectMongoDB()
	connectRedis()

	// Load JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// MongoDB database
	database := mongoClient.Database("PulsePoll")

	// Initialize MongoDB collections
	userCollection := database.Collection("users")
	pollCollection := database.Collection("polls")
	voteCollection := database.Collection("votes")

	// Initialize database indexes
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	// Authentication service
	authService := services.NewAuthService(
		userCollection,
		jwtSecret,
	)

	// Unique email index
	if err := authService.EnsureIndexes(ctx); err != nil {
		cancel()
		log.Fatal("Failed to create user indexes:", err)
	}

	// Unique vote index: one vote per voter per poll
	_, err := voteCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "pollId", Value: 1},
				{Key: "voterId", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)

	if err != nil {
		cancel()
		log.Fatal("Failed to create vote index:", err)
	}

	cancel()

	// Initialize authentication controller
	authController := controllers.NewAuthController(authService)

	// Initialize poll service with MongoDB and Redis
	pollService := services.NewPollService(
		pollCollection,
		voteCollection,
		redisClient,
	)

	pollController := controllers.NewPollController(pollService)

	// Initialize vote service and controller
	voteService := services.NewVoteService(
		voteCollection,
		pollCollection,
		redisClient,
	)

	voteController := controllers.NewVoteController(voteService)

	// Initialize WebSocket controller
	websocketController := controllers.NewWebSocketController(redisClient)

	// Initialize Gin router
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin == "http://localhost:5173" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal("Failed to configure trusted proxies:", err)
	}

	// Application health
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"application": "PulsePoll",
			"status":      "UP",
		})
	})

	// MongoDB health
	router.GET("/api/health/db", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := mongoClient.Ping(ctx, nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"database": "MongoDB",
				"status":   "DISCONNECTED",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"database": "MongoDB",
			"status":   "CONNECTED",
		})
	})

	// Redis health
	router.GET("/api/health/redis", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"database": "Redis",
				"status":   "DISCONNECTED",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"database": "Redis",
			"status":   "CONNECTED",
		})
	})

	// API routes
	api := router.Group("/api")

	// Public authentication routes
	routes.RegisterAuthRoutes(api, authController)

	// Poll creation and public retrieval routes
	routes.RegisterPollRoutes(api, pollController)

	// Public voting route
	routes.RegisterVoteRoutes(api, voteController)

	// WebSocket route for live poll updates
	routes.RegisterWebSocketRoutes(api, websocketController)

	// Protected authentication test endpoint
	api.GET(
		"/auth/me",
		middleware.AuthMiddleware(),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")

			c.JSON(http.StatusOK, gin.H{
				"message": "Authentication successful",
				"userID":  userID,
			})
		},
	)

	// Start server
	log.Println("PulsePoll server starting on port 8081")

	if err := router.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
