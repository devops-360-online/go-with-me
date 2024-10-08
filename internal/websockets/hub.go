package websockets

import (
    "context"
    "github.com/gorilla/websocket"
    "github.com/go-redis/redis/v8"
    "log"
	"sync"
)

type Client struct {
    Conn    *websocket.Conn
    EventID string
}

var clients = make(map[*Client]bool)
var register = make(chan *Client)
var unregister = make(chan *Client)
var redisClient *redis.Client
var mu sync.Mutex

func RunHub(rdb *redis.Client) {
    redisClient = rdb
    go subscribeToRedis()
    for {
        select {
        case client := <-register:
            clients[client] = true
        case client := <-unregister:
            if _, ok := clients[client]; ok {
                delete(clients, client)
                client.Conn.Close()
            }
        }
    }
}

func subscribeToRedis() {
    pubsub := redisClient.Subscribe(context.Background(), "chat_messages")
    defer pubsub.Close()
    for {
        msg, err := pubsub.ReceiveMessage(context.Background())
        if err != nil {
            log.Println("Redis subscribe error:", err)
            continue
        }
        // Broadcast the message to all clients
        for client := range clients {
            // Filter clients by EventID
            if client.EventID == msg.Channel {
                client.Conn.WriteJSON(msg.Payload)
            }
        }
    }
}

func RegisterClient(client *Client) {
    mu.Lock()
    defer mu.Unlock()
    clients[client] = true
}

func UnregisterClient(client *Client) {
    mu.Lock()
    defer mu.Unlock()
    if _, ok := clients[client]; ok {
        delete(clients, client)
        client.Conn.Close()
    }
}
