package handlers

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "time"

    jwt "github.com/appleboy/gin-jwt/v2"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/go-redis/redis/v8"
    "github.com/devops-360-online/go-with-me/config"
    "github.com/devops-360-online/go-with-me/internal/middlewares"
    "github.com/devops-360-online/go-with-me/internal/models"
    "github.com/devops-360-online/go-with-me/internal/websockets"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Adjust in production
    },
}

func EventChatHandler(c *gin.Context) {
    eventID := c.Param("id")
    // Get user ID from JWT
    claims := jwt.ExtractClaims(c)
    userID := uint(claims[middlewares.IdentityKey].(float64))

    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    defer conn.Close()

    // Get configuration from context
    cfg := c.MustGet("config").(*config.Config)

    // Get MongoDB client from context
    mongoClient := c.MustGet("mongoClient").(*mongo.Client)
    messageCollection := mongoClient.Database(cfg.MongoDatabase).Collection("messages")

    // Get Redis client from context
    redisClient := c.MustGet("redisClient").(*redis.Client)

    // Send previous messages to the client
    go sendPreviousMessages(eventID, conn, messageCollection)

    // Register client in websockets hub
    client := &websockets.Client{
        Conn:    conn,
        EventID: eventID,
    }
    websockets.RegisterClient(client)
    defer websockets.UnregisterClient(client)

    // Handle incoming messages
    for {
        var msg models.Message
        err := conn.ReadJSON(&msg)
        if err != nil {
            log.Println("Read error:", err)
            break
        }
        msg.EventID = parseEventID(eventID)
        msg.SenderID = userID
        msg.Timestamp = time.Now()

        // Save message to MongoDB
        _, err = messageCollection.InsertOne(context.TODO(), msg)
        if err != nil {
            log.Println("MongoDB insert error:", err)
            continue
        }

        // Publish message to Redis
        msgJSON, _ := json.Marshal(msg)
        redisClient.Publish(context.Background(), "chat_messages", msgJSON)
    }
}

func sendPreviousMessages(eventID string, conn *websocket.Conn, messageCollection *mongo.Collection) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    cursor, err := messageCollection.Find(ctx, bson.M{"event_id": parseEventID(eventID)})
    if err != nil {
        log.Println("MongoDB find error:", err)
        return
    }
    defer cursor.Close(ctx)
    for cursor.Next(ctx) {
        var msg models.Message
        err := cursor.Decode(&msg)
        if err != nil {
            log.Println("Cursor decode error:", err)
            continue
        }
        conn.WriteJSON(msg)
    }
}

func parseEventID(eventID string) uint {
    id, err := strconv.ParseUint(eventID, 10, 64)
    if err != nil {
        log.Println("Invalid event ID:", err)
        return 0 // Handle the error appropriately
    }
    return uint(id)
}
