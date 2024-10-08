package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type ChatGroup struct {
    ID      primitive.ObjectID `bson:"_id,omitempty"`
    EventID uint               `bson:"event_id"`
    Members []uint             `bson:"members"`
}
