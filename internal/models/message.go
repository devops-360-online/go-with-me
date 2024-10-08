package models

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

type Message struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    EventID   uint               `bson:"event_id"`
    SenderID  uint               `bson:"sender_id"`
    Content   string             `bson:"content"`
    Timestamp time.Time          `bson:"timestamp"`
}
