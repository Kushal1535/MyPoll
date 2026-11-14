package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PollOption struct {
	ID    string `bson:"id" json:"id"`
	Text  string `bson:"text" json:"text"`
	Votes int64  `bson:"votes" json:"votes"`
}

type Poll struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedBy   bson.ObjectID `bson:"createdBy" json:"createdBy"`
	Question    string        `bson:"question" json:"question"`
	Description string        `bson:"description,omitempty" json:"description,omitempty"`
	Options     []PollOption  `bson:"options" json:"options"`
	IsActive    bool          `bson:"isActive" json:"isActive"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
}
