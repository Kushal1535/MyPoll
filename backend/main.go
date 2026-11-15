package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"pulsepoll/backend/config"
	"pulsepoll/backend/controllers"
	"pulsepoll/backend/middleware"
	"pulsepoll/backend/routes"
	"pulsepoll/backend/services"
)

var mongoClient *mongo.Client
var redisClient *redis.Client

// Connect to MongoDB using the existing database configuration.
func connectMongoDB() {
	client, err := config.ConnectMongoDB()
	if err != nil {
		log.Fatal("MongoDB connection failed:", err)
	}

	mongoClient = client
	log.Println("MongoDB connected successfully!")
}

// Connect to Redis.
func connectRedis() {
	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))

	var redisOptions *redis.Options

	if redisURL != "" {
		parsedOptions, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatal("Invalid REDIS_URL:", err)
		}

		// redis.ParseURL automatically configures TLS for rediss:// URLs.
		redisOptions = parsedOptions
	} else {
		redisOptions = &redis.Options{
			Addr: "127.0.0.1:6379",
		}
	}

	redisOptions.DialTimeout = 10 * time.Second
	redisOptions.ReadTimeout = 5 * time.Second
	redisOptions.WriteTimeout = 5 * time.Second

	redisClient = redis.NewClient(redisOptions)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		20*time.Second,
	)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		_ = redisClient.Close()
		log.Fatal("Redis connection failed:", err)
	}

	log.Println("Redis connected successfully!")
}

// Configure allowed frontend origins.
func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]bool{
		"http://localhost:5173": true,
	}

	frontendURLs := strings.TrimSpace(os.Getenv("FRONTEND_URL"))

	if frontendURLs != "" {
		for _, origin := range strings.Split(frontendURLs, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins[origin] = true
			}
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header(
				"Access-Control-Allow-Methods",
				"GET, POST, PUT, PATCH, DELETE, OPTIONS",
			)
			c.Header(
				"Access-Control-Allow-Headers",
				"Origin, Content-Type, Authorization",
			)
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {
	// Load required JWT secret.
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Initialize MongoDB.
	connectMongoDB()

	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Println("MongoDB disconnect error:", err)
		}
	}()

	// Initialize Redis.
	connectRedis()

	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Println("Redis disconnect error:", err)
		}
	}()

	// MongoDB database.
	database := mongoClient.Database("PulsePoll")

	// MongoDB collections.
	userCollection := database.Collection("users")
	pollCollection := database.Collection("polls")
	voteCollection := database.Collection("votes")

	// Initialize database indexes.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)

	authService := services.NewAuthService(
		userCollection,
		jwtSecret,
	)

	// Unique email index.
	if err := authService.EnsureIndexes(ctx); err != nil {
		cancel()
		log.Fatal("Failed to create user indexes:", err)
	}

	// Unique vote index.
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

	cancel()

	if err != nil {
		log.Fatal("Failed to create vote index:", err)
	}

	// Initialize authentication controller.
	authController := controllers.NewAuthController(authService)

	// Initialize poll service.
	pollService := services.NewPollService(
		pollCollection,
		voteCollection,
		redisClient,
	)

	pollController := controllers.NewPollController(pollService)

	// Initialize vote service and controller.
	voteService := services.NewVoteService(
		voteCollection,
		pollCollection,
		redisClient,
	)

	voteController := controllers.NewVoteController(voteService)

	// Initialize WebSocket controller.
	websocketController := controllers.NewWebSocketController(redisClient)

	// Initialize Gin router.
	router := gin.Default()

	router.Use(corsMiddleware())

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal("Failed to configure trusted proxies:", err)
	}

	// Application health.
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"application": "PulsePoll",
			"status":      "UP",
		})
	})

	// MongoDB health.
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

	// Redis health.
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

	// API routes.
	api := router.Group("/api")

	routes.RegisterAuthRoutes(api, authController)
	routes.RegisterPollRoutes(api, pollController)
	routes.RegisterVoteRoutes(api, voteController)
	routes.RegisterWebSocketRoutes(api, websocketController)

	// Protected authentication test endpoint.
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

	// Start server using Render's assigned port.
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8081"
	}

	log.Printf("PulsePoll server starting on port %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Server startup failed:", err)
	}
}
