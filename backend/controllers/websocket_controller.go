
package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 25 * time.Second
	maxMessageSize = 512
)

type WebSocketController struct {
	redisClient *redis.Client
}

func NewWebSocketController(redisClient *redis.Client) *WebSocketController {
	return &WebSocketController{
		redisClient: redisClient,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:5173"
	},
}

func (wc *WebSocketController) HandlePollUpdates(c *gin.Context) {
	pollID := c.Param("id")

	if pollID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Poll ID is required",
		})
		return
	}

	// Establish WebSocket connection.
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket connected for poll: %s", pollID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	channel := "poll:" + pollID + ":updates"

	// Subscribe to Redis Pub/Sub.
	pubsub := wc.redisClient.Subscribe(ctx, channel)
	defer pubsub.Close()

	// Wait until Redis confirms the subscription.
	if _, err := pubsub.Receive(ctx); err != nil {
		log.Println("Redis subscription failed:", err)
		return
	}

	log.Printf("Redis subscribed successfully: %s", channel)

	messages := pubsub.Channel()

	// Configure WebSocket read handling.
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	readErrors := make(chan error, 1)

	// Read incoming WebSocket control messages and detect disconnects.
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				readErrors <- err
				return
			}
		}
	}()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	// Forward Redis events to the browser.
	for {
		select {
		case <-ctx.Done():
			return

		case err := <-readErrors:
			log.Printf("WebSocket client disconnected for poll %s: %v", pollID, err)
			return

		case message, ok := <-messages:
			if !ok {
				log.Printf("Redis message channel closed for poll: %s", pollID)
				return
			}

			conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := conn.WriteMessage(
				websocket.TextMessage,
				[]byte(message.Payload),
			); err != nil {
				log.Printf("WebSocket write failed for poll %s: %v", pollID, err)
				return
			}

			log.Printf("WebSocket event delivered for poll %s: %s", pollID, message.Payload)

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := conn.WriteMessage(
				websocket.PingMessage,
				nil,
			); err != nil {
				log.Printf("WebSocket ping failed for poll %s: %v", pollID, err)
				return
			}
		}
	}
}
